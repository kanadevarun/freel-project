package carrier_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/freel/backend/internal/carrier/adapters"
	"github.com/freel/backend/internal/carrier/adapters/token"
	"github.com/freel/backend/internal/carrier/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Task 3.4: Maersk Booking Payload, Idempotency & Safe Retries (12 Tests)
// ─────────────────────────────────────────────────────────────────────────────

// TEST 1 — Correct payload
// Create a valid internal booking. Mock Maersk.
// Verify the outgoing JSON matches the intended Maersk/DCSA booking structure.
func TestMaerskBooking_Test1_CorrectPayload(t *testing.T) {
	var capturedPayload map[string]interface{}
	var capturedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "valid-token-payload", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}

		if r.URL.Path == "/booking/v1/bookings" && r.Method == "POST" {
			capturedHeaders = r.Header.Clone()
			_ = json.NewDecoder(r.Body).Decode(&capturedPayload)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{
				"carrierBookingReference": "MSK-BKG-9901",
				"bookingStatus": "CONFIRMED",
				"vesselName": "MAERSK MC-KINNEY MOLLER",
				"voyageNumber": "2401W"
			}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()
	creds := domain.DecryptedCredentials{
		APIKey:    "test-consumer-key",
		APISecret: "test-api-secret",
		BaseURL:   server.URL,
	}

	readyDate := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)
	req := domain.BookingRequest{
		CarrierSCAC:       "MAEU",
		ContractNumber:    "CTR-2026-MSK",
		OriginPort:        "INNSA",
		DestinationPort:   "NLRTM",
		EquipmentType:     "40HC",
		Quantity:          2,
		CargoReadyDate:    readyDate,
		Commodity:         "Automotive Parts",
		ShipperName:       "Auto Exports India Pvt Ltd",
		ConsigneeName:     "Euro Car Parts BV",
		CustomerReference: "LHQ-BKG-2026-101",
		IdempotencyKey:    "lhq-org1-bkg101",
	}

	res, err := adapter.CreateBooking(ctx, creds, domain.EnvSandbox, req)
	require.NoError(t, err)
	require.NotNil(t, res)

	// Verify DCSA BKG Structure
	require.NotNil(t, capturedPayload, "Outgoing payload must not be nil")
	assert.Equal(t, "LHQ-BKG-2026-101", capturedPayload["carrierBookingRequestReference"])
	assert.Equal(t, "CTR-2026-MSK", capturedPayload["contractQuotationReference"])
	assert.Equal(t, "CY", capturedPayload["receiptTypeAtOrigin"])
	assert.Equal(t, "CY", capturedPayload["deliveryTypeAtDestination"])
	assert.Equal(t, "FCL", capturedPayload["cargoMovementTypeAtOrigin"])
	assert.Equal(t, "FCL", capturedPayload["cargoMovementTypeAtDestination"])

	// Verify Commodities
	commodities, ok := capturedPayload["commodities"].([]interface{})
	require.True(t, ok, "commodities must be an array")
	require.Len(t, commodities, 1)
	commMap := commodities[0].(map[string]interface{})
	assert.Equal(t, "Automotive Parts", commMap["commodityType"])

	// Verify Requested Equipments
	equipments, ok := capturedPayload["requestedEquipments"].([]interface{})
	require.True(t, ok, "requestedEquipments must be an array")
	require.Len(t, equipments, 1)
	equipMap := equipments[0].(map[string]interface{})
	assert.Equal(t, "40HC", equipMap["requestedEquipmentTypeCode"])
	assert.Equal(t, float64(2), equipMap["units"])

	// Verify Document Parties (Shipper "OS" and Consignee "CN")
	parties, ok := capturedPayload["documentParties"].([]interface{})
	require.True(t, ok, "documentParties must be an array")
	require.Len(t, parties, 2)

	var hasShipper, hasConsignee bool
	for _, p := range parties {
		partyMap := p.(map[string]interface{})
		fn := partyMap["partyFunction"].(string)
		partyObj := partyMap["party"].(map[string]interface{})
		name := partyObj["partyName"].(string)

		if fn == "OS" && name == "Auto Exports India Pvt Ltd" {
			hasShipper = true
		}
		if fn == "CN" && name == "Euro Car Parts BV" {
			hasConsignee = true
		}
	}
	assert.True(t, hasShipper, "Shipper (OS) must be mapped correctly")
	assert.True(t, hasConsignee, "Consignee (CN) must be mapped correctly")

	// Verify Transports & Locations
	transports, ok := capturedPayload["transports"].(map[string]interface{})
	require.True(t, ok, "transports must be an object")
	assert.Equal(t, "2026-11-15", transports["plannedDepartureDate"])

	pol := transports["portOfLoading"].(map[string]interface{})
	assert.Equal(t, "INNSA", pol["unLocationCode"])
	pod := transports["portOfDischarge"].(map[string]interface{})
	assert.Equal(t, "NLRTM", pod["unLocationCode"])

	// Verify Headers
	assert.Equal(t, "Bearer valid-token-payload", capturedHeaders.Get("Authorization"))
	assert.Equal(t, "test-consumer-key", capturedHeaders.Get("Consumer-Key"))
	assert.Equal(t, "lhq-org1-bkg101", capturedHeaders.Get("Idempotency-Key"))
}

// TEST 2 — Required validation
// Missing mandatory fields -> validation error -> ZERO HTTP requests to Maersk.
func TestMaerskBooking_Test2_RequiredValidation(t *testing.T) {
	var httpCalls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&httpCalls, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()
	creds := domain.DecryptedCredentials{
		APIKey:    "test-key",
		APISecret: "test-secret",
		BaseURL:   server.URL,
	}

	testCases := []struct {
		name        string
		req         domain.BookingRequest
		expectedErr string
	}{
		{
			name: "Missing Origin Port",
			req: domain.BookingRequest{
				DestinationPort: "NLRTM",
				EquipmentType:   "40HC",
				Quantity:        1,
				ShipperName:     "Shipper A",
				ConsigneeName:   "Consignee B",
			},
			expectedErr: "origin port is required",
		},
		{
			name: "Missing Destination Port",
			req: domain.BookingRequest{
				OriginPort:    "INNSA",
				EquipmentType: "40HC",
				Quantity:      1,
				ShipperName:   "Shipper A",
				ConsigneeName: "Consignee B",
			},
			expectedErr: "destination port is required",
		},
		{
			name: "Missing Equipment Type",
			req: domain.BookingRequest{
				OriginPort:      "INNSA",
				DestinationPort: "NLRTM",
				Quantity:        1,
				ShipperName:     "Shipper A",
				ConsigneeName:   "Consignee B",
			},
			expectedErr: "equipment type is required",
		},
		{
			name: "Zero Quantity",
			req: domain.BookingRequest{
				OriginPort:      "INNSA",
				DestinationPort: "NLRTM",
				EquipmentType:   "40HC",
				Quantity:        0,
				ShipperName:     "Shipper A",
				ConsigneeName:   "Consignee B",
			},
			expectedErr: "equipment quantity must be at least 1",
		},
		{
			name: "Missing Shipper Name",
			req: domain.BookingRequest{
				OriginPort:      "INNSA",
				DestinationPort: "NLRTM",
				EquipmentType:   "40HC",
				Quantity:        1,
				ConsigneeName:   "Consignee B",
			},
			expectedErr: "shipper name is required",
		},
		{
			name: "Missing Consignee Name",
			req: domain.BookingRequest{
				OriginPort:      "INNSA",
				DestinationPort: "NLRTM",
				EquipmentType:   "40HC",
				Quantity:        1,
				ShipperName:     "Shipper A",
			},
			expectedErr: "consignee name is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			beforeCalls := atomic.LoadInt32(&httpCalls)
			res, err := adapter.CreateBooking(ctx, creds, domain.EnvSandbox, tc.req)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), tc.expectedErr)
			afterCalls := atomic.LoadInt32(&httpCalls)
			assert.Equal(t, beforeCalls, afterCalls, "Zero HTTP requests must be sent on validation failure")
		})
	}
}

// TEST 3 — Idempotency header
// Verify Idempotency-Key header is present, non-empty, and stable.
func TestMaerskBooking_Test3_IdempotencyHeader(t *testing.T) {
	var capturedIdempotencyKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "token-test3", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}
		if r.URL.Path == "/booking/v1/bookings" {
			capturedIdempotencyKey = r.Header.Get("Idempotency-Key")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"carrierBookingReference": "MSK-BKG-333"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()
	creds := domain.DecryptedCredentials{
		APIKey:    "test-key",
		APISecret: "test-secret",
		BaseURL:   server.URL,
	}

	req := domain.BookingRequest{
		OriginPort:        "INNSA",
		DestinationPort:   "NLRTM",
		EquipmentType:     "40HC",
		Quantity:          1,
		ShipperName:       "Shipper 3",
		ConsigneeName:     "Consignee 3",
		CustomerReference: "BKG-REF-3001",
	}

	res, err := adapter.CreateBooking(ctx, creds, domain.EnvSandbox, req)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.NotEmpty(t, capturedIdempotencyKey)
	assert.Equal(t, "idemp-maersk-BKG-REF-3001", capturedIdempotencyKey)
}

// TEST 4 — Same key across retry
// Force the first booking request to fail transiently (503).
// Then retry. Capture both requests.
// Verify Request 1 and Request 2 carry the EXACT same Idempotency-Key.
func TestMaerskBooking_Test4_SameKeyAcrossRetry(t *testing.T) {
	var attemptCount int32
	var capturedKeys []string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "token-test4", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}
		if r.URL.Path == "/booking/v1/bookings" {
			attempt := atomic.AddInt32(&attemptCount, 1)
			mu.Lock()
			capturedKeys = append(capturedKeys, r.Header.Get("Idempotency-Key"))
			mu.Unlock()

			if attempt == 1 {
				w.WriteHeader(http.StatusServiceUnavailable) // 503 transient failure
				w.Write([]byte(`{"error": "temporarily unavailable"}`))
				return
			}
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"carrierBookingReference": "MSK-RETRY-SUCCESS"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()
	creds := domain.DecryptedCredentials{
		APIKey:    "test-key",
		APISecret: "test-secret",
		BaseURL:   server.URL,
	}

	req := domain.BookingRequest{
		OriginPort:        "INNSA",
		DestinationPort:   "NLRTM",
		EquipmentType:     "20GP",
		Quantity:          1,
		ShipperName:       "Shipper 4",
		ConsigneeName:     "Consignee 4",
		CustomerReference: "BKG-REF-4001",
		IdempotencyKey:    "lhq-org1-bkg4001",
	}

	res, err := adapter.CreateBooking(ctx, creds, domain.EnvSandbox, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "MSK-RETRY-SUCCESS", res.BookingNumber)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, capturedKeys, 2, "Expected exactly 2 attempts due to 503 retry")
	assert.Equal(t, "lhq-org1-bkg4001", capturedKeys[0])
	assert.Equal(t, "lhq-org1-bkg4001", capturedKeys[1])
	assert.Equal(t, capturedKeys[0], capturedKeys[1], "Idempotency-Key MUST be identical across retries")
}

// TEST 5 — No duplicate random keys
// Call CreateBooking multiple times for the same logical booking request.
// Verify a new random key is NOT generated.
func TestMaerskBooking_Test5_NoDuplicateRandomKeys(t *testing.T) {
	var capturedKeys []string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "token-test5", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}
		if r.URL.Path == "/booking/v1/bookings" {
			mu.Lock()
			capturedKeys = append(capturedKeys, r.Header.Get("Idempotency-Key"))
			mu.Unlock()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"carrierBookingReference": "MSK-BKG-500"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()
	creds := domain.DecryptedCredentials{
		APIKey:    "test-key",
		APISecret: "test-secret",
		BaseURL:   server.URL,
	}

	// Request without explicit IdempotencyKey but with CustomerReference
	req := domain.BookingRequest{
		OriginPort:        "INNSA",
		DestinationPort:   "NLRTM",
		EquipmentType:     "40HC",
		Quantity:          1,
		ShipperName:       "Shipper 5",
		ConsigneeName:     "Consignee 5",
		CustomerReference: "LHQ-STABLE-REF-5",
	}

	// First execution
	res1, err := adapter.CreateBooking(ctx, creds, domain.EnvSandbox, req)
	require.NoError(t, err)
	require.NotNil(t, res1)

	// Second execution for same logical booking
	res2, err := adapter.CreateBooking(ctx, creds, domain.EnvSandbox, req)
	require.NoError(t, err)
	require.NotNil(t, res2)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, capturedKeys, 2)
	assert.Equal(t, "idemp-maersk-LHQ-STABLE-REF-5", capturedKeys[0])
	assert.Equal(t, "idemp-maersk-LHQ-STABLE-REF-5", capturedKeys[1])
	assert.Equal(t, capturedKeys[0], capturedKeys[1], "Same logical booking must reuse identical stable idempotency key")
}

// TEST 6 — Successful booking
// Mock POST /booking/v1/bookings -> 201 Created.
// Verify response parsed, carrier booking reference extracted, status and vessel info mapped.
func TestMaerskBooking_Test6_SuccessfulBooking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "token-test6", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}
		if r.URL.Path == "/booking/v1/bookings" {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{
				"carrierBookingReference": "MSK-CONFIRMED-8899",
				"bookingStatus": "CONFIRMED",
				"confirmationRef": "CONF-MSK-8899",
				"vesselName": "MAERSK MADRID",
				"voyageNumber": "2605E",
				"expectedDepartureDate": "2026-11-20T10:00:00Z",
				"expectedArrivalDate": "2026-12-05T14:30:00Z"
			}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()
	creds := domain.DecryptedCredentials{
		APIKey:    "test-key",
		APISecret: "test-secret",
		BaseURL:   server.URL,
	}

	req := domain.BookingRequest{
		OriginPort:        "SGSIN",
		DestinationPort:   "NLRTM",
		EquipmentType:     "40HC",
		Quantity:          3,
		ShipperName:       "Singapore Logistics Pte",
		ConsigneeName:     "Rotterdam Hub BV",
		CustomerReference: "BKG-2026-601",
	}

	res, err := adapter.CreateBooking(ctx, creds, domain.EnvSandbox, req)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, "MSK-CONFIRMED-8899", res.BookingNumber)
	assert.Equal(t, "MAEU", res.CarrierSCAC)
	assert.Equal(t, "CONFIRMED", res.Status)
	assert.Equal(t, "CONF-MSK-8899", res.ConfirmationRef)
	assert.Equal(t, "MAERSK MADRID", res.VesselName)
	assert.Equal(t, "2605E", res.VoyageNumber)
	require.NotNil(t, res.ETD)
	assert.Equal(t, 2026, res.ETD.Year())
	assert.Equal(t, time.November, res.ETD.Month())
	require.NotNil(t, res.ETA)
	assert.Equal(t, 2026, res.ETA.Year())
	assert.Equal(t, time.December, res.ETA.Month())
}

// TEST 7 — 400 Bad Request
// Maersk returns 400 Bad Request.
// Verify: NO retry, NO second booking request, sanitized error.
func TestMaerskBooking_Test7_400BadRequest(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "token-test7", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}
		if r.URL.Path == "/booking/v1/bookings" {
			atomic.AddInt32(&callCount, 1)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Invalid commodity code or port restriction"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()
	creds := domain.DecryptedCredentials{
		APIKey:    "test-key",
		APISecret: "test-secret",
		BaseURL:   server.URL,
	}

	req := domain.BookingRequest{
		OriginPort:      "INNSA",
		DestinationPort: "NLRTM",
		EquipmentType:   "40HC",
		Quantity:        1,
		ShipperName:     "Shipper 7",
		ConsigneeName:   "Consignee 7",
	}

	res, err := adapter.CreateBooking(ctx, creds, domain.EnvSandbox, req)
	require.Error(t, err)
	assert.Nil(t, res)

	// Verify no retry took place
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount), "400 Bad Request must NOT be retried")
}

// TEST 8 — 401 Unauthorized
// Request 1 uses Token-A -> returns 401.
// Invalidate Token-A, fetch Token-B.
// Request 2 uses Token-B + SAME Idempotency-Key.
// Verify exactly one authentication retry.
func TestMaerskBooking_Test8_401TokenRefreshWithSameIdempotencyKey(t *testing.T) {
	var oauthCalls int32
	var bookingCalls int32
	var capturedTokens []string
	var capturedIdempKeys []string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			n := atomic.AddInt32(&oauthCalls, 1)
			w.WriteHeader(http.StatusOK)
			if n == 1 {
				w.Write([]byte(`{"access_token": "token-expired-A", "token_type": "Bearer", "expires_in": 3600}`))
			} else {
				w.Write([]byte(`{"access_token": "token-fresh-B", "token_type": "Bearer", "expires_in": 3600}`))
			}
			return
		}
		if r.URL.Path == "/booking/v1/bookings" {
			n := atomic.AddInt32(&bookingCalls, 1)
			mu.Lock()
			capturedTokens = append(capturedTokens, r.Header.Get("Authorization"))
			capturedIdempKeys = append(capturedIdempKeys, r.Header.Get("Idempotency-Key"))
			mu.Unlock()

			if n == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "invalid_token", "error_description": "Token expired"}`))
				return
			}
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"carrierBookingReference": "MSK-401-RESOLVED"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	token.GetDefaultTokenManager().Invalidate("MAERSK", "test-consumer-key-8", domain.EnvSandbox)
	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()
	creds := domain.DecryptedCredentials{
		APIKey:    "test-consumer-key-8",
		APISecret: "test-secret-8",
		BaseURL:   server.URL,
	}

	req := domain.BookingRequest{
		OriginPort:        "INNSA",
		DestinationPort:   "NLRTM",
		EquipmentType:     "40HC",
		Quantity:          1,
		ShipperName:       "Shipper 8",
		ConsigneeName:     "Consignee 8",
		IdempotencyKey:    "lhq-stable-idemp-key-888",
		CustomerReference: "BKG-888",
	}

	res, err := adapter.CreateBooking(ctx, creds, domain.EnvSandbox, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "MSK-401-RESOLVED", res.BookingNumber)

	assert.Equal(t, int32(2), atomic.LoadInt32(&oauthCalls), "Expected 2 OAuth token calls (initial + refresh)")
	assert.Equal(t, int32(2), atomic.LoadInt32(&bookingCalls), "Expected 2 booking calls (initial + 401 retry)")

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, capturedTokens, 2)
	assert.Equal(t, "Bearer token-expired-A", capturedTokens[0])
	assert.Equal(t, "Bearer token-fresh-B", capturedTokens[1])

	require.Len(t, capturedIdempKeys, 2)
	assert.Equal(t, "lhq-stable-idemp-key-888", capturedIdempKeys[0])
	assert.Equal(t, "lhq-stable-idemp-key-888", capturedIdempKeys[1])
	assert.Equal(t, capturedIdempKeys[0], capturedIdempKeys[1], "Idempotency-Key MUST remain identical across 401 refresh retry")
}

// TEST 9 — 5xx / transient failure
// Simulate retryable server failure (502 Bad Gateway).
// Verify retry follows policy, same Idempotency-Key, and succeeds.
func TestMaerskBooking_Test9_5xxTransientFailure(t *testing.T) {
	var attemptCount int32
	var capturedKeys []string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "token-test9", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}
		if r.URL.Path == "/booking/v1/bookings" {
			n := atomic.AddInt32(&attemptCount, 1)
			mu.Lock()
			capturedKeys = append(capturedKeys, r.Header.Get("Idempotency-Key"))
			mu.Unlock()

			if n == 1 {
				w.WriteHeader(http.StatusBadGateway) // 502
				w.Write([]byte(`{"error": "bad gateway"}`))
				return
			}
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"carrierBookingReference": "MSK-502-RECOVERED"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()
	creds := domain.DecryptedCredentials{
		APIKey:    "test-key",
		APISecret: "test-secret",
		BaseURL:   server.URL,
	}

	req := domain.BookingRequest{
		OriginPort:      "INNSA",
		DestinationPort: "NLRTM",
		EquipmentType:   "20GP",
		Quantity:        2,
		ShipperName:     "Shipper 9",
		ConsigneeName:   "Consignee 9",
		IdempotencyKey:  "lhq-idemp-test-999",
	}

	res, err := adapter.CreateBooking(ctx, creds, domain.EnvSandbox, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "MSK-502-RECOVERED", res.BookingNumber)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, capturedKeys, 2)
	assert.Equal(t, "lhq-idemp-test-999", capturedKeys[0])
	assert.Equal(t, "lhq-idemp-test-999", capturedKeys[1])
}

// TEST 10 — Timeout / unknown outcome
// Simulate: first request drops connection (network timeout), second succeeds.
// Verify the retry uses the SAME idempotency key.
func TestMaerskBooking_Test10_TimeoutUnknownOutcome(t *testing.T) {
	var callCount int32
	var capturedKeys []string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "token-test10", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}
		if r.URL.Path == "/booking/v1/bookings" {
			n := atomic.AddInt32(&callCount, 1)
			mu.Lock()
			capturedKeys = append(capturedKeys, r.Header.Get("Idempotency-Key"))
			mu.Unlock()

			if n == 1 {
				// Hijack and abruptly close connection to simulate network drop / timeout
				hj, ok := w.(http.Hijacker)
				if ok {
					conn, _, _ := hj.Hijack()
					conn.Close()
					return
				}
				w.WriteHeader(http.StatusGatewayTimeout)
				return
			}
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"carrierBookingReference": "MSK-CONN-DROP-RECOVERED"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()
	creds := domain.DecryptedCredentials{
		APIKey:    "test-key",
		APISecret: "test-secret",
		BaseURL:   server.URL,
	}

	req := domain.BookingRequest{
		OriginPort:      "INNSA",
		DestinationPort: "NLRTM",
		EquipmentType:   "40HC",
		Quantity:        1,
		ShipperName:     "Shipper 10",
		ConsigneeName:   "Consignee 10",
		IdempotencyKey:  "lhq-network-timeout-key",
	}

	res, err := adapter.CreateBooking(ctx, creds, domain.EnvSandbox, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "MSK-CONN-DROP-RECOVERED", res.BookingNumber)

	mu.Lock()
	defer mu.Unlock()
	require.GreaterOrEqual(t, len(capturedKeys), 2, "Expected at least 2 attempts after connection drop")
	for i := 1; i < len(capturedKeys); i++ {
		assert.Equal(t, capturedKeys[0], capturedKeys[i], "All retry attempts following timeout must reuse the exact same Idempotency-Key")
	}
}

// TEST 11 — Tenant isolation
// Verify two organizations cannot accidentally share the same booking idempotency identity.
func TestMaerskBooking_Test11_TenantIsolation(t *testing.T) {
	org1ID := 101
	org2ID := 202
	bookingID := 5555

	keyOrg1 := fmt.Sprintf("lhq-org%d-bkg%d", org1ID, bookingID)
	keyOrg2 := fmt.Sprintf("lhq-org%d-bkg%d", org2ID, bookingID)

	assert.NotEqual(t, keyOrg1, keyOrg2, "Idempotency keys between different orgs must never conflict")
	assert.Contains(t, keyOrg1, "org101")
	assert.Contains(t, keyOrg2, "org202")

	// Also verify deterministic hashing produces different keys if parameters differ
	reqOrg1 := domain.BookingRequest{
		OriginPort:      "INNSA",
		DestinationPort: "NLRTM",
		EquipmentType:   "40HC",
		Quantity:        1,
		ShipperName:     "Tenant 1 Client",
		ConsigneeName:   "Receiver 1",
		IdempotencyKey:  keyOrg1,
	}
	reqOrg2 := domain.BookingRequest{
		OriginPort:      "INNSA",
		DestinationPort: "NLRTM",
		EquipmentType:   "40HC",
		Quantity:        1,
		ShipperName:     "Tenant 2 Client",
		ConsigneeName:   "Receiver 2",
		IdempotencyKey:  keyOrg2,
	}

	adapter := adapters.NewMaerskAdapter()
	var capturedKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "token-test11", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}
		if r.URL.Path == "/booking/v1/bookings" {
			capturedKey = r.Header.Get("Idempotency-Key")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"carrierBookingReference": "MSK-TENANT-OK"}`))
			return
		}
	}))
	defer server.Close()

	ctx := context.Background()
	creds := domain.DecryptedCredentials{APIKey: "key", APISecret: "secret", BaseURL: server.URL}

	_, err := adapter.CreateBooking(ctx, creds, domain.EnvSandbox, reqOrg1)
	require.NoError(t, err)
	assert.Equal(t, keyOrg1, capturedKey)

	_, err = adapter.CreateBooking(ctx, creds, domain.EnvSandbox, reqOrg2)
	require.NoError(t, err)
	assert.Equal(t, keyOrg2, capturedKey)
}

// TEST 12 — Security
// Verify tests and error responses do not leak:
// API secret, client secret, access token, or Authorization header.
func TestMaerskBooking_Test12_Security(t *testing.T) {
	secretAPI := "SUPER_SECRET_CLIENT_SECRET_99999"
	secretToken := "SENSITIVE_BEARER_TOKEN_ABCD1234EFGH"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"access_token": "%s", "token_type": "Bearer", "expires_in": 3600}`, secretToken)))
			return
		}
		if r.URL.Path == "/booking/v1/bookings" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": "Internal failure while evaluating token %s with secret %s"}`, secretToken, secretAPI)))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()
	creds := domain.DecryptedCredentials{
		APIKey:    "test-consumer-key-12",
		APISecret: secretAPI,
		BaseURL:   server.URL,
	}

	req := domain.BookingRequest{
		OriginPort:      "INNSA",
		DestinationPort: "NLRTM",
		EquipmentType:   "40HC",
		Quantity:        1,
		ShipperName:     "Shipper 12",
		ConsigneeName:   "Consignee 12",
	}

	res, err := adapter.CreateBooking(ctx, creds, domain.EnvSandbox, req)
	require.Error(t, err)
	assert.Nil(t, res)

	errString := err.Error()
	assert.NotContains(t, errString, secretAPI, "API secret MUST NEVER leak in error messages")
	assert.NotContains(t, errString, secretToken, "Access token MUST NEVER leak in error messages")
	assert.NotContains(t, errString, "Bearer ", "Authorization header MUST NEVER leak in error messages")
}
