package carrier_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/freel/backend/internal/carrier/adapters"
	"github.com/freel/backend/internal/carrier/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Task 3.3: Maersk Tracking & Rates API Verification Tests
// ─────────────────────────────────────────────────────────────────────────────

// 1. TRACKING VERIFICATION
// Endpoint: /track-and-trace/v2/events
// Verifies Bearer token, Consumer-Key, query parameters, and DCSA normalization
func TestTask33_TrackingVerification(t *testing.T) {
	var capturedAuth, capturedConsumerKey, capturedEquipmentRef string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "tracking-test-token",
				"token_type": "Bearer",
				"expires_in": 3600
			}`))
			return
		}

		if strings.HasPrefix(r.URL.Path, "/track-and-trace/v2/events") {
			capturedAuth = r.Header.Get("Authorization")
			capturedConsumerKey = r.Header.Get("Consumer-Key")
			capturedEquipmentRef = r.URL.Query().Get("equipmentReference")

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"eventID": "maersk-evt-101",
					"eventType": "EQUIPMENT",
					"eventDateTime": "2026-08-10T08:00:00Z",
					"eventClassifierCode": "ACT",
					"equipmentEventTypeCode": "GTIN",
					"equipmentReference": "MSKU9041285",
					"carrierBookingReference": "BKG-MAEU-88219",
					"eventLocation": {
						"locationName": "Nhava Sheva Terminal",
						"UNLocationCode": "INNSA"
					}
				},
				{
					"eventID": "maersk-evt-102",
					"eventType": "EQUIPMENT",
					"eventDateTime": "2026-08-12T14:00:00Z",
					"eventClassifierCode": "ACT",
					"equipmentEventTypeCode": "LOAD",
					"equipmentReference": "MSKU9041285",
					"carrierBookingReference": "BKG-MAEU-88219",
					"transportCall": {
						"carrierVoyageNumber": "2608W",
						"vessel": {
							"vesselName": "MAERSK MC-KINNEY MOLLER",
							"vesselIMONumber": "9619907"
						},
						"location": {
							"locationName": "Jawaharlal Nehru Port"
						}
					}
				},
				{
					"eventID": "maersk-evt-103",
					"eventType": "TRANSPORT",
					"eventDateTime": "2026-08-13T04:00:00Z",
					"eventClassifierCode": "ACT",
					"transportEventTypeCode": "DEPA",
					"transportCall": {
						"carrierVoyageNumber": "2608W",
						"vessel": {
							"vesselName": "MAERSK MC-KINNEY MOLLER"
						},
						"location": {
							"locationName": "Nhava Sheva"
						}
					}
				}
			]`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()

	creds := domain.DecryptedCredentials{
		APIKey:    "maersk-test-consumer-key",
		APISecret: "maersk-test-secret",
		BaseURL:   server.URL,
	}

	res, err := adapter.GetTracking(ctx, creds, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU9041285",
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	// Verify headers & params
	assert.Equal(t, "Bearer tracking-test-token", capturedAuth)
	assert.Equal(t, "maersk-test-consumer-key", capturedConsumerKey)
	assert.Equal(t, "MSKU9041285", capturedEquipmentRef)

	// Verify DCSA normalization
	assert.Equal(t, "MAEU", res.CarrierSCAC)
	assert.Equal(t, "MSKU9041285", res.ContainerNumber)
	assert.Equal(t, "BKG-MAEU-88219", res.BookingNumber)
	assert.Equal(t, "DEPARTED", res.CurrentStatus)
	assert.Len(t, res.Events, 3)
	assert.Equal(t, "GATE_IN", res.Events[0].MilestoneCode)
	assert.Equal(t, "LOADED", res.Events[1].MilestoneCode)
	assert.Equal(t, "MAERSK MC-KINNEY MOLLER", res.Events[1].VesselName)
	assert.Equal(t, "2608W", res.Events[1].VoyageNumber)
	assert.Equal(t, "DEPARTED", res.Events[2].MilestoneCode)
}

// 2. SPOT RATES VERIFICATION
// Endpoint: /ocean-spot/v1/instant-quotes
// Verifies authentication, JSON payload construction, and rate normalization
func TestTask33_SpotRatesVerification(t *testing.T) {
	var capturedAuth, capturedConsumerKey, capturedContentType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "spot-rates-test-token",
				"token_type": "Bearer",
				"expires_in": 3600
			}`))
			return
		}

		if r.URL.Path == "/ocean-spot/v1/instant-quotes" {
			capturedAuth = r.Header.Get("Authorization")
			capturedConsumerKey = r.Header.Get("Consumer-Key")
			capturedContentType = r.Header.Get("Content-Type")

			assert.Equal(t, "POST", r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"rates": [
					{
						"priceId": "MSK-SPOT-9921",
						"origin": "INNSA",
						"destination": "NLRTM",
						"commodity": "General Cargo",
						"equipmentType": "40HC",
						"currency": "USD",
						"baseOceanPrice": 2400.00,
						"originSurcharge": 200.00,
						"destSurcharge": 300.00,
						"totalPrice": 2900.00,
						"transitTimeDays": 24,
						"freeDays": 14,
						"validityStart": "2026-08-01T00:00:00Z",
						"validityEnd": "2026-09-30T23:59:59Z",
						"isContract": false
					}
				]
			}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()

	creds := domain.DecryptedCredentials{
		APIKey:    "maersk-test-consumer-key-spot",
		APISecret: "maersk-test-secret-spot",
		BaseURL:   server.URL,
	}

	rates, err := adapter.GetSpotRates(ctx, creds, domain.EnvSandbox, domain.SpotRateRequest{
		RateRequest: domain.RateRequest{
			OriginPort:      "INNSA",
			DestinationPort: "NLRTM",
			EquipmentType:   "40HC",
		},
	})
	require.NoError(t, err)
	require.Len(t, rates, 1)

	assert.Equal(t, "Bearer spot-rates-test-token", capturedAuth)
	assert.Equal(t, "maersk-test-consumer-key-spot", capturedConsumerKey)
	assert.Contains(t, capturedContentType, "application/json")

	rate := rates[0]
	assert.Equal(t, "MSK-SPOT-9921", rate.RateID)
	assert.Equal(t, "MAEU", rate.CarrierSCAC)
	assert.Equal(t, "INNSA", rate.OriginPort)
	assert.Equal(t, "NLRTM", rate.DestinationPort)
	assert.Equal(t, "40HC", rate.EquipmentType)
	assert.Equal(t, "USD", rate.Currency)
	assert.Equal(t, 2400.00, rate.OceanFreight)
	assert.Equal(t, 200.00, rate.OriginCharges)
	assert.Equal(t, 300.00, rate.DestinationCharges)
	assert.Equal(t, 2900.00, rate.TotalBuyPrice)
	assert.Equal(t, 24, rate.TransitDays)
	assert.Equal(t, 14, rate.FreeDays)
	assert.False(t, rate.IsContractRate)
}

// 3. CONTRACT RATES VERIFICATION
// Endpoint: /rates/v1/contract-rates
// Verifies X-Contract-Number header, authentication, and contract rate normalization
func TestTask33_ContractRatesVerification(t *testing.T) {
	var capturedAuth, capturedConsumerKey, capturedContractNum string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "contract-rates-token",
				"token_type": "Bearer",
				"expires_in": 3600
			}`))
			return
		}

		if r.URL.Path == "/rates/v1/contract-rates" {
			capturedAuth = r.Header.Get("Authorization")
			capturedConsumerKey = r.Header.Get("Consumer-Key")
			capturedContractNum = r.Header.Get("X-Contract-Number")

			assert.Equal(t, "POST", r.Method)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"rates": [
					{
						"priceId": "MSK-CT-5001",
						"origin": "INNSA",
						"destination": "NLRTM",
						"commodity": "Electronics",
						"equipmentType": "40HC",
						"currency": "USD",
						"baseOceanPrice": 1800.00,
						"originSurcharge": 150.00,
						"destSurcharge": 250.00,
						"totalPrice": 2200.00,
						"transitTimeDays": 22,
						"freeDays": 21,
						"validityStart": "2026-01-01T00:00:00Z",
						"validityEnd": "2026-12-31T23:59:59Z",
						"isContract": true
					}
				]
			}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()

	creds := domain.DecryptedCredentials{
		APIKey:    "maersk-test-consumer-key-ct",
		APISecret: "maersk-test-secret-ct",
		BaseURL:   server.URL,
	}

	rates, err := adapter.GetContractRates(ctx, creds, domain.EnvSandbox, domain.ContractRateRequest{
		RateRequest: domain.RateRequest{
			OriginPort:      "INNSA",
			DestinationPort: "NLRTM",
			EquipmentType:   "40HC",
		},
		ContractNumber: "MAEU-CT-8891",
	})
	require.NoError(t, err)
	require.Len(t, rates, 1)

	assert.Equal(t, "Bearer contract-rates-token", capturedAuth)
	assert.Equal(t, "maersk-test-consumer-key-ct", capturedConsumerKey)
	assert.Equal(t, "MAEU-CT-8891", capturedContractNum)

	rate := rates[0]
	assert.Equal(t, "MSK-CT-5001", rate.RateID)
	assert.Equal(t, 2200.00, rate.TotalBuyPrice)
	assert.True(t, rate.IsContractRate, "Must be tagged as contract rate")
	assert.Equal(t, 21, rate.FreeDays)
}

// 4. ERROR HANDLING: 400 Bad Request
func TestTask33_ErrorHandling_400_BadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "valid-token", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/track-and-trace/v2/events") {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid_parameter", "message": "Equipment reference format is invalid"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()

	superSecret := "secret-400-check-991"
	creds := domain.DecryptedCredentials{
		APIKey:    "test-key-400",
		APISecret: superSecret,
		BaseURL:   server.URL,
	}

	_, err := adapter.GetTracking(ctx, creds, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "INVALID9999",
	})
	require.Error(t, err)

	var intErr *domain.IntegrationError
	require.ErrorAs(t, err, &intErr)
	assert.Equal(t, domain.ErrCodeInvalidRequest, intErr.ErrorCode)
	assert.Equal(t, http.StatusBadRequest, intErr.HTTPStatus)
	assert.NotContains(t, err.Error(), superSecret, "Never expose secret")
}

// 5. ERROR HANDLING: 403 Forbidden
func TestTask33_ErrorHandling_403_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "valid-token", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}
		if r.URL.Path == "/ocean-spot/v1/instant-quotes" {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "forbidden", "message": "Subscription plan does not permit live quotes"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()

	creds := domain.DecryptedCredentials{
		APIKey:    "test-key-403",
		APISecret: "test-secret-403",
		BaseURL:   server.URL,
	}

	_, err := adapter.GetSpotRates(ctx, creds, domain.EnvSandbox, domain.SpotRateRequest{
		RateRequest: domain.RateRequest{OriginPort: "INNSA", DestinationPort: "NLRTM"},
	})
	require.Error(t, err)

	var intErr *domain.IntegrationError
	require.ErrorAs(t, err, &intErr)
	assert.Equal(t, domain.ErrCodeForbidden, intErr.ErrorCode)
	assert.Equal(t, http.StatusForbidden, intErr.HTTPStatus)
}

// 6. ERROR HANDLING: 404 Not Found
func TestTask33_ErrorHandling_404_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "valid-token", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/track-and-trace/v2/events") {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "not_found", "message": "No records found for container"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()

	creds := domain.DecryptedCredentials{
		APIKey:    "test-key-404",
		APISecret: "test-secret-404",
		BaseURL:   server.URL,
	}

	_, err := adapter.GetTracking(ctx, creds, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU0000000",
	})
	require.Error(t, err)

	var intErr *domain.IntegrationError
	require.ErrorAs(t, err, &intErr)
	assert.Equal(t, domain.ErrCodeNotFound, intErr.ErrorCode)
	assert.Equal(t, http.StatusNotFound, intErr.HTTPStatus)
}

// 7. ERROR HANDLING: 429 Rate Limit (Retry Preserved)
// Verify existing CarrierHTTPClient retry recovers and succeeds on second attempt
func TestTask33_ErrorHandling_429_RateLimit_RetryPreserved(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "valid-token-429", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/track-and-trace/v2/events") {
			a := atomic.AddInt32(&attempts, 1)
			if a == 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error": "rate_limit_exceeded"}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()

	creds := domain.DecryptedCredentials{
		APIKey:    "test-key-429",
		APISecret: "test-secret-429",
		BaseURL:   server.URL,
	}

	res, err := adapter.GetTracking(ctx, creds, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU4290001",
	})
	require.NoError(t, err, "CarrierHTTPClient retry must succeed on 2nd attempt")
	require.NotNil(t, res)
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts), "Should have retried after 429")
}

// 8. ERROR HANDLING: 5xx Server Error (Retry Preserved)
// Verify existing CarrierHTTPClient retry recovers from 502/503 and succeeds on second attempt
func TestTask33_ErrorHandling_5xx_ServerError_RetryPreserved(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "valid-token-5xx", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/track-and-trace/v2/events") {
			a := atomic.AddInt32(&attempts, 1)
			if a == 1 {
				w.WriteHeader(http.StatusBadGateway)
				w.Write([]byte(`{"error": "bad_gateway"}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()

	creds := domain.DecryptedCredentials{
		APIKey:    "test-key-5xx",
		APISecret: "test-secret-5xx",
		BaseURL:   server.URL,
	}

	res, err := adapter.GetTracking(ctx, creds, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU5020001",
	})
	require.NoError(t, err, "CarrierHTTPClient retry must succeed after 502")
	require.NotNil(t, res)
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts), "Should have retried after 5xx")
}

// 9. ERROR HANDLING: OAuth Failure
func TestTask33_ErrorHandling_OAuthFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "invalid_client", "error_description": "Credentials revoked"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()

	testSecret := "super-secret-oauth-fail"
	creds := domain.DecryptedCredentials{
		APIKey:    "test-key-oauth-fail",
		APISecret: testSecret,
		BaseURL:   server.URL,
	}

	_, err := adapter.GetSpotRates(ctx, creds, domain.EnvSandbox, domain.SpotRateRequest{
		RateRequest: domain.RateRequest{OriginPort: "INNSA", DestinationPort: "NLRTM"},
	})
	require.Error(t, err)

	var intErr *domain.IntegrationError
	require.ErrorAs(t, err, &intErr)
	assert.Equal(t, domain.ErrCodeAuthFailed, intErr.ErrorCode)
	assert.NotContains(t, err.Error(), testSecret, "No secret in error")
}

// 10. 401 REGRESSION: Tracking 401 Token Invalidation & Single Refresh
func TestTask33_401Regression_TrackingRefresh(t *testing.T) {
	var oauthCalls, trackingCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			c := atomic.AddInt32(&oauthCalls, 1)
			w.WriteHeader(http.StatusOK)
			if c == 1 {
				w.Write([]byte(`{"access_token": "token-trk-A", "token_type": "Bearer", "expires_in": 3600}`))
			} else {
				w.Write([]byte(`{"access_token": "token-trk-B", "token_type": "Bearer", "expires_in": 3600}`))
			}
			return
		}

		if strings.HasPrefix(r.URL.Path, "/track-and-trace/v2/events") {
			c := atomic.AddInt32(&trackingCalls, 1)
			auth := r.Header.Get("Authorization")
			if auth == "Bearer token-trk-A" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "token_expired"}`))
				return
			}
			assert.Equal(t, "Bearer token-trk-B", auth, "Retry must use fresh token-B")
			assert.Equal(t, int32(2), c)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()

	creds := domain.DecryptedCredentials{
		APIKey:    "test-trk-refresh-key",
		APISecret: "test-trk-refresh-secret",
		BaseURL:   server.URL,
	}

	res, err := adapter.GetTracking(ctx, creds, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU4010001",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, int32(2), atomic.LoadInt32(&trackingCalls))
	assert.Equal(t, int32(2), atomic.LoadInt32(&oauthCalls))
}

// 11. 401 REGRESSION: Rates 401 Token Invalidation & Single Refresh
func TestTask33_401Regression_RatesRefresh(t *testing.T) {
	var oauthCalls, ratesCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			c := atomic.AddInt32(&oauthCalls, 1)
			w.WriteHeader(http.StatusOK)
			if c == 1 {
				w.Write([]byte(`{"access_token": "token-rates-A", "token_type": "Bearer", "expires_in": 3600}`))
			} else {
				w.Write([]byte(`{"access_token": "token-rates-B", "token_type": "Bearer", "expires_in": 3600}`))
			}
			return
		}

		if r.URL.Path == "/ocean-spot/v1/instant-quotes" {
			c := atomic.AddInt32(&ratesCalls, 1)
			auth := r.Header.Get("Authorization")
			if auth == "Bearer token-rates-A" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "token_expired"}`))
				return
			}
			assert.Equal(t, "Bearer token-rates-B", auth, "Retry must use fresh token-B")
			assert.Equal(t, int32(2), c)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"rates": []}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()

	creds := domain.DecryptedCredentials{
		APIKey:    "test-rates-refresh-key",
		APISecret: "test-rates-refresh-secret",
		BaseURL:   server.URL,
	}

	rates, err := adapter.GetSpotRates(ctx, creds, domain.EnvSandbox, domain.SpotRateRequest{
		RateRequest: domain.RateRequest{OriginPort: "INNSA", DestinationPort: "NLRTM"},
	})
	require.NoError(t, err)
	assert.NotNil(t, rates)
	assert.Equal(t, int32(2), atomic.LoadInt32(&ratesCalls))
	assert.Equal(t, int32(2), atomic.LoadInt32(&oauthCalls))
}

// 12. MULTI-TENANCY ISOLATION: Organization Credentials Verification
func TestTask33_MultiTenantIsolation(t *testing.T) {
	var mu sync.Mutex
	capturedKeys := make(map[string]string)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			_ = r.ParseForm()
			clientID := r.FormValue("client_id")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "tok-for-` + clientID + `", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}

		if strings.HasPrefix(r.URL.Path, "/track-and-trace/v2/events") {
			consumerKey := r.Header.Get("Consumer-Key")
			auth := r.Header.Get("Authorization")
			mu.Lock()
			capturedKeys[consumerKey] = auth
			mu.Unlock()

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()

	// Tenant 1 (Org 101)
	credsOrg1 := domain.DecryptedCredentials{
		APIKey:    "org-101-consumer-key",
		APISecret: "org-101-secret",
		BaseURL:   server.URL,
	}
	_, err := adapter.GetTracking(ctx, credsOrg1, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU1010001",
	})
	require.NoError(t, err)

	// Tenant 2 (Org 202)
	credsOrg2 := domain.DecryptedCredentials{
		APIKey:    "org-202-consumer-key",
		APISecret: "org-202-secret",
		BaseURL:   server.URL,
	}
	_, err = adapter.GetTracking(ctx, credsOrg2, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU2020002",
	})
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, "Bearer tok-for-org-101-consumer-key", capturedKeys["org-101-consumer-key"])
	assert.Equal(t, "Bearer tok-for-org-202-consumer-key", capturedKeys["org-202-consumer-key"])
	assert.NotEqual(t, capturedKeys["org-101-consumer-key"], capturedKeys["org-202-consumer-key"])
}
