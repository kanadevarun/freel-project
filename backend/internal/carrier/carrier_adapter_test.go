package carrier_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/freel/backend/internal/carrier/adapters"
	"github.com/freel/backend/internal/carrier/adapters/httpclient"
	"github.com/freel/backend/internal/carrier/adapters/token"
	"github.com/freel/backend/internal/carrier/domain"
	"github.com/freel/backend/internal/carrier/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// 1. DCSA Track & Trace Normalization Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestDCSATrackingNormalization(t *testing.T) {
	rawDCSAEvents := []domain.DCSAEvent{
		{
			EventID:                "evt-001",
			EventType:              domain.DCSAEventEquipment,
			EventDateTime:          "2026-08-10T08:30:00Z",
			EventClassifierCode:    domain.DCSAClassifierActual,
			EquipmentEventTypeCode: domain.DCSAMilestoneGateIn,
			EquipmentReference:     "MSKU9041285",
			CarrierBookingRef:      "BKG-MAEU-88219",
			EventLocation: &domain.DCSALocation{
				LocationName:   "Nhava Sheva Container Terminal",
				UNLocationCode: "INNSA",
			},
		},
		{
			EventID:                "evt-002",
			EventType:              domain.DCSAEventEquipment,
			EventDateTime:          "2026-08-11T14:15:00Z",
			EventClassifierCode:    domain.DCSAClassifierActual,
			EquipmentEventTypeCode: domain.DCSAMilestoneLoaded,
			EquipmentReference:     "MSKU9041285",
			TransportCall: &domain.DCSATransportCall{
				CarrierVoyageNumber: "2608W",
				Vessel: &domain.DCSAVessel{
					VesselName:      "MAERSK MC-KINNEY MOLLER",
					VesselIMONumber: "9619907",
				},
				Location: &domain.DCSALocation{
					LocationName: "Jawaharlal Nehru Port",
				},
			},
		},
		{
			EventID:                "evt-003",
			EventType:              domain.DCSAEventTransport,
			EventDateTime:          "2026-08-12T02:00:00Z",
			EventClassifierCode:    domain.DCSAClassifierActual,
			TransportEventTypeCode: domain.DCSAMilestoneDeparture,
			TransportCall: &domain.DCSATransportCall{
				CarrierVoyageNumber: "2608W",
				Vessel: &domain.DCSAVessel{
					VesselName: "MAERSK MC-KINNEY MOLLER",
				},
				Location: &domain.DCSALocation{
					LocationName: "Nhava Sheva",
				},
			},
		},
		{
			EventID:                "evt-004",
			EventType:              domain.DCSAEventTransport,
			EventDateTime:          "2026-08-28T18:00:00Z",
			EventClassifierCode:    domain.DCSAClassifierEstimated,
			TransportEventTypeCode: domain.DCSAMilestoneArrival,
			TransportCall: &domain.DCSATransportCall{
				CarrierVoyageNumber: "2608W",
				Location: &domain.DCSALocation{
					LocationName:   "Port of Rotterdam",
					UNLocationCode: "NLRTM",
				},
			},
		},
	}

	norm := adapters.NormalizeDCSATrackingEvents("MAEU", "MSKU9041285", rawDCSAEvents)
	require.NotNil(t, norm)

	assert.Equal(t, "MAEU", norm.CarrierSCAC)
	assert.Equal(t, "MSKU9041285", norm.ContainerNumber)
	assert.Equal(t, "BKG-MAEU-88219", norm.BookingNumber)
	assert.Equal(t, "DEPARTED", norm.CurrentStatus)
	assert.Equal(t, "Nhava Sheva", norm.LatestLocation)
	assert.False(t, norm.IsDelivered)

	require.NotNil(t, norm.ActualDeparture)
	assert.Equal(t, 12, norm.ActualDeparture.Day())

	require.NotNil(t, norm.EstimatedArrival)
	assert.Equal(t, 28, norm.EstimatedArrival.Day())

	// Verify events are sorted chronologically
	require.Len(t, norm.Events, 4)
	assert.Equal(t, "GATE_IN", norm.Events[0].MilestoneCode)
	assert.Equal(t, "LOADED", norm.Events[1].MilestoneCode)
	assert.Equal(t, "MAERSK MC-KINNEY MOLLER", norm.Events[1].VesselName)
	assert.Equal(t, "2608W", norm.Events[1].VoyageNumber)
	assert.Equal(t, "DEPARTED", norm.Events[2].MilestoneCode)
	assert.Equal(t, "ARRIVED", norm.Events[3].MilestoneCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// 2. HTTP Client Resilience & Error Mapping Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestCarrierHTTPClientResilienceAndErrorMapping(t *testing.T) {
	var attempts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if r.URL.Path == "/retry-test" {
			if attempts < 3 {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error": "rate limit reached"}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
			return
		}

		if r.URL.Path == "/auth-fail" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "invalid consumer key"}`))
			return
		}

		if r.URL.Path == "/forbidden" {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "forbidden"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := httpclient.NewCarrierHTTPClient(httpclient.HTTPClientConfig{
		BaseURL:        server.URL,
		Timeout:        5 * time.Second,
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
		ProviderCode:   "MAERSK",
	})

	ctx := context.Background()

	// 1. Verify successful retry on 429
	attempts = 0
	body, status, err := client.Execute(ctx, "GET", "/retry-test", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, string(body), "ok")
	assert.Equal(t, 3, attempts, "Client should have retried until success")

	// 2. Verify immediate translation on 401 Unauthorized (Non-retryable)
	attempts = 0
	_, status, err = client.Execute(ctx, "GET", "/auth-fail", nil, nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, status)
	assert.Equal(t, 1, attempts, "Non-retryable 401 must not retry")

	var intErr *domain.IntegrationError
	require.ErrorAs(t, err, &intErr)
	assert.Equal(t, domain.ErrCodeAuthFailed, intErr.ErrorCode)
	assert.Contains(t, intErr.UserMessage, "authentication failed")

	// 3. Verify Forbidden on 403
	_, status, err = client.Execute(ctx, "GET", "/forbidden", nil, nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, status)
	require.ErrorAs(t, err, &intErr)
	assert.Equal(t, domain.ErrCodeForbidden, intErr.ErrorCode)
}

// ─────────────────────────────────────────────────────────────────────────────
// 3. OAuth2 Token Manager Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestOAuth2TokenManagerCaching(t *testing.T) {
	var tokenCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCalls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"access_token": "token-xyz-12345",
			"token_type": "Bearer",
			"expires_in": 3600
		}`))
	}))
	defer server.Close()

	mgr := token.GetDefaultTokenManager()
	ctx := context.Background()

	req := token.TokenRequest{
		TokenURL:     server.URL,
		ClientID:     "client-id-test",
		ClientSecret: "client-secret-test",
		ProviderCode: "MAERSK",
		Environment:  domain.EnvSandbox,
	}

	tokenCalls = 0
	tok1, err := mgr.GetToken(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "token-xyz-12345", tok1)
	assert.Equal(t, 1, tokenCalls)

	// Second request within TTL should hit memory cache
	tok2, err := mgr.GetToken(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "token-xyz-12345", tok2)
	assert.Equal(t, 1, tokenCalls, "Subsequent request should use cached token without HTTP call")
}

// ─────────────────────────────────────────────────────────────────────────────
// 4. Maersk Adapter Operations & Capability Enforcement
// ─────────────────────────────────────────────────────────────────────────────

func TestMaerskAdapterOperationsAndCapabilityEnforcement(t *testing.T) {
	// Mock Maersk API Gateway server
	maerskMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/track-and-trace/v2/events" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"eventID": "maersk-evt-991",
					"eventType": "EQUIPMENT",
					"eventDateTime": "2026-08-20T10:00:00Z",
					"eventClassifierCode": "ACT",
					"equipmentEventTypeCode": "LOAD",
					"equipmentReference": "MSKU9041285",
					"carrierBookingReference": "BKG-992120",
					"eventLocation": {
						"locationName": "Jawaharlal Nehru Port",
						"UNLocationCode": "INNSA"
					}
				}
			]`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer maerskMockServer.Close()

	db, _, svc := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	testOrgID := int64(999881)
	testUserID := int64(777661)

	// Clean up test org
	_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id = ?", testOrgID)
	_, _ = db.Exec("DELETE FROM organizations WHERE id = ?", testOrgID)
	_, _ = db.Exec("INSERT INTO organizations (id, name) VALUES (?, 'Test Adapter Org')", testOrgID)
	defer func() {
		_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id = ?", testOrgID)
		_, _ = db.Exec("DELETE FROM organizations WHERE id = ?", testOrgID)
	}()

	// 1. Connect Maersk with TRACKING capability ONLY
	connectReq := domain.ConnectCarrierRequest{
		CarrierSCAC:      "MAEU",
		Environment:      "SANDBOX",
		ConnectionMethod: "API",
		Credentials: map[string]interface{}{
			"api_key":   "maersk-test-consumer-key-8812",
			"base_url":  maerskMockServer.URL,
			"client_id": "LOGISTIQ-TEST",
		},
		Capabilities: []string{"TRACKING"}, // Only Tracking enabled
	}

	integration, err := svc.ConnectCarrier(ctx, testOrgID, testUserID, connectReq)
	require.NoError(t, err)
	require.NotNil(t, integration)

	// 2. Execute Test Connection (should succeed against gateway)
	testRes, err := svc.TestConnection(ctx, testOrgID, integration.ID)
	require.NoError(t, err)
	assert.True(t, testRes.Success, "Test connection should succeed with valid gateway")

	// 3. Execute Real GetTracking and verify normalization
	trackingRes, err := svc.GetTracking(ctx, testOrgID, integration.ID, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU9041285",
	})
	require.NoError(t, err)
	require.NotNil(t, trackingRes)
	assert.Equal(t, "MAEU", trackingRes.CarrierSCAC)
	assert.Equal(t, "MSKU9041285", trackingRes.ContainerNumber)
	assert.Equal(t, "LOADED", trackingRes.CurrentStatus)
	assert.Len(t, trackingRes.Events, 1)

	// 4. Verify Capability Enforcement: Rates must be rejected since only TRACKING is enabled
	_, err = svc.GetRates(ctx, testOrgID, integration.ID, domain.RateRequest{
		OriginPort:      "INNSA",
		DestinationPort: "NLRTM",
		EquipmentType:   "40HC",
	})
	assert.ErrorIs(t, err, service.ErrCapabilityDisabled, "Rates operation must be blocked when capability is not enabled")

	// 5. Update Integration to enable RATES
	newCaps := []string{"TRACKING", "RATES", "CONTRACT_RATES", "BOOKING", "DOCUMENTS"}
	_, err = svc.UpdateCarrier(ctx, testOrgID, testUserID, integration.ID, domain.UpdateCarrierRequest{
		Capabilities: newCaps,
	})
	require.NoError(t, err)

	// 6. Verify Disabled Integration Guard
	_, err = svc.ToggleCarrier(ctx, testOrgID, testUserID, integration.ID, false)
	require.NoError(t, err)

	_, err = svc.GetTracking(ctx, testOrgID, integration.ID, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU9041285",
	})
	assert.ErrorIs(t, err, service.ErrIntegrationDisabled, "Disabled integration must reject operations")
}
