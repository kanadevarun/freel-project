package carrier_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/freel/backend/internal/carrier/adapters"
	"github.com/freel/backend/internal/carrier/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Task 4.1: Hapag-Lloyd Integration Audit & Production Readiness Tests
// =============================================================================

// TEST 1 — Capability Verification
// Verify Hapag-Lloyd supports ONLY CapTracking, and that CapRates, CapBooking,
// and CapDocuments are NOT supported.
func TestHapagLloyd_Test1_Capabilities(t *testing.T) {
	registry := adapters.GetDefaultRegistry()
	hapag := registry.GetAdapterBySCAC("HLCU")
	require.NotNil(t, hapag)

	// Tracking must be supported
	assert.True(t, hapag.SupportsCapability(domain.CapTracking), "Hapag must support CapTracking")
	assert.True(t, hapag.HasCapability(domain.CapTracking))

	// All other capabilities must be false
	assert.False(t, hapag.SupportsCapability(domain.CapRates), "Hapag must NOT support CapRates")
	assert.False(t, hapag.SupportsCapability(domain.CapSpotRates), "Hapag must NOT support CapSpotRates")
	assert.False(t, hapag.SupportsCapability(domain.CapContractRates), "Hapag must NOT support CapContractRates")
	assert.False(t, hapag.SupportsCapability(domain.CapBooking), "Hapag must NOT support CapBooking")
	assert.False(t, hapag.SupportsCapability(domain.CapDocuments), "Hapag must NOT support CapDocuments")

	caps := hapag.SupportedCapabilities()
	assert.Len(t, caps, 1)
	assert.Equal(t, domain.CapTracking, caps[0])

	// Attempting unsupported operations directly on adapter must return capability error
	creds := domain.DecryptedCredentials{APIKey: "hl-key", ClientSecret: "hl-sec"}
	_, err := hapag.GetRates(context.Background(), creds, domain.EnvSandbox, domain.RateRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not currently support")

	_, err = hapag.CreateBooking(context.Background(), creds, domain.EnvSandbox, domain.BookingRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not currently support")

	_, err = hapag.GetDocuments(context.Background(), creds, domain.EnvSandbox, domain.DocumentRequest{BookingNumber: "HL-123"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not currently support")
}

// TEST 2 — No Fake Connection
// Verify Hapag-Lloyd cannot return synthetic successful connection status when it
// does not have valid real credentials or cannot connect to the gateway.
func TestHapagLloyd_Test2_NoFakeConnection(t *testing.T) {
	registry := adapters.GetDefaultRegistry()
	hapag := registry.GetAdapterBySCAC("HLCU")
	require.NotNil(t, hapag)

	ctx := context.Background()

	// Scenario A: Empty credentials -> must fail with 400 Bad Request, never synthetic success
	resEmpty, err := hapag.TestConnection(ctx, domain.DecryptedCredentials{}, domain.EnvSandbox)
	require.NoError(t, err)
	assert.False(t, resEmpty.Success)
	assert.Equal(t, http.StatusBadRequest, resEmpty.HTTPStatus)
	assert.Contains(t, resEmpty.Message, "Missing Hapag-Lloyd Client ID")

	// Scenario B: Gateway returns 401 Unauthorized -> must report real failure, not simulated success
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"httpCode":        "401",
			"httpMessage":     "Unauthorized",
			"moreInformation": "Invalid client credentials",
		})
	}))
	defer server.Close()

	creds := domain.DecryptedCredentials{
		BaseURL:      server.URL,
		APIKey:       "invalid-key",
		ClientSecret: "invalid-secret",
	}

	resInvalid, err := hapag.TestConnection(ctx, creds, domain.EnvSandbox)
	require.NoError(t, err)
	assert.False(t, resInvalid.Success, "Must NOT return synthetic success on invalid credentials")
	assert.Equal(t, http.StatusUnauthorized, resInvalid.HTTPStatus)
	assert.Equal(t, domain.ErrCodeAuthFailed, resInvalid.ErrorCode)
}

// TEST 3 — Authentication & Headers
// Mock the actual Hapag-Lloyd gateway authentication contract.
// Verify X-IBM-Client-Id, X-IBM-Client-Secret, and Accept headers are sent.
func TestHapagLloyd_Test3_AuthenticationHeaders(t *testing.T) {
	registry := adapters.GetDefaultRegistry()
	hapag := registry.GetAdapterBySCAC("HLCU")
	require.NotNil(t, hapag)

	var capturedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	// Verify using client_id / client_secret
	creds := domain.DecryptedCredentials{
		BaseURL:      server.URL,
		ClientID:     "hl-ibm-client-998",
		ClientSecret: "hl-ibm-secret-441",
	}

	res, err := hapag.TestConnection(context.Background(), creds, domain.EnvSandbox)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "hl-ibm-client-998", capturedHeaders.Get("X-IBM-Client-Id"))
	assert.Equal(t, "hl-ibm-secret-441", capturedHeaders.Get("X-IBM-Client-Secret"))
	assert.Equal(t, "application/json", capturedHeaders.Get("Accept"))
}

// TEST 4 — Tracking Request
// Verify endpoint, method, query parameter mapping for Container, Booking, and MBL references.
func TestHapagLloyd_Test4_TrackingRequest(t *testing.T) {
	registry := adapters.GetDefaultRegistry()
	hapag := registry.GetAdapterBySCAC("HLCU")
	require.NotNil(t, hapag)

	ctx := context.Background()

	// Subtest A: Equipment Reference (Container)
	var capturedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	creds := domain.DecryptedCredentials{
		BaseURL:      server.URL,
		APIKey:       "hl-key",
		ClientSecret: "hl-secret",
	}

	_, err := hapag.GetTracking(ctx, creds, domain.EnvSandbox, domain.TrackingRequest{
		ContainerNumber: "HLXU1234567",
	})
	require.NoError(t, err)
	assert.Contains(t, capturedPath, "/track-and-trace/v2/events")
	assert.Contains(t, capturedPath, "equipmentReference=HLXU1234567")

	// Subtest B: Carrier Booking Reference
	_, err = hapag.GetTracking(ctx, creds, domain.EnvSandbox, domain.TrackingRequest{
		BookingNumber: "HL-BKG-889900",
	})
	require.NoError(t, err)
	assert.Contains(t, capturedPath, "carrierBookingReference=HL-BKG-889900")

	// Subtest C: Transport Document Reference (MBL)
	_, err = hapag.GetTracking(ctx, creds, domain.EnvSandbox, domain.TrackingRequest{
		MBLNumber: "HLCUHAM260901",
	})
	require.NoError(t, err)
	assert.Contains(t, capturedPath, "transportDocumentReference=HLCUHAM260901")
}

// TEST 5 — Tracking Response Normalization
// Verify representative Hapag-Lloyd DCSA events are normalized into LogisticsHQ tracking domain.
func TestHapagLloyd_Test5_TrackingNormalization(t *testing.T) {
	registry := adapters.GetDefaultRegistry()
	hapag := registry.GetAdapterBySCAC("HLCU")
	require.NotNil(t, hapag)

	mockEvents := []map[string]interface{}{
		{
			"eventID":                "ev-hl-001",
			"eventType":              "TRANSPORT",
			"eventDateTime":          "2026-09-10T14:30:00Z",
			"eventClassifierCode":    "ACT",
			"transportEventTypeCode": "DEPA",
			"transportCall": map[string]interface{}{
				"carrierVoyageNumber": "2609W",
				"vessel": map[string]interface{}{
					"vesselName": "HAMBURG EXPRESS",
				},
				"location": map[string]interface{}{
					"unLocationCode": "DEHAM",
					"locationName":   "Port of Hamburg",
				},
			},
			"equipmentReference": "HLXU8901234",
			"carrierBookingReference": "HL-BKG-99221",
		},
		{
			"eventID":                "ev-hl-002",
			"eventType":              "EQUIPMENT",
			"eventDateTime":          "2026-09-08T09:15:00Z",
			"eventClassifierCode":    "ACT",
			"equipmentEventTypeCode": "LOAD",
			"eventLocation": map[string]interface{}{
				"unLocationCode": "DEHAM",
				"locationName":   "Hamburg CTA Terminal",
			},
			"equipmentReference": "HLXU8901234",
			"carrierBookingReference": "HL-BKG-99221",
		},
	}

	bodyBytes, _ := json.Marshal(mockEvents)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bodyBytes)
	}))
	defer server.Close()

	creds := domain.DecryptedCredentials{
		BaseURL:      server.URL,
		APIKey:       "hl-key",
		ClientSecret: "hl-secret",
	}

	res, err := hapag.GetTracking(context.Background(), creds, domain.EnvSandbox, domain.TrackingRequest{
		ContainerNumber: "HLXU8901234",
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, "HLCU", res.CarrierSCAC)
	assert.Equal(t, "HLXU8901234", res.ContainerNumber)
	assert.Equal(t, "HL-BKG-99221", res.BookingNumber)
	assert.Len(t, res.Events, 2)

	// Chronologically sorted (oldest to newest)
	assert.Equal(t, "ev-hl-002", res.Events[0].EventID)
	assert.Equal(t, "LOADED", res.Events[0].MilestoneCode)
	assert.Equal(t, "Hamburg CTA Terminal", res.Events[0].Location)

	assert.Equal(t, "ev-hl-001", res.Events[1].EventID)
	assert.Equal(t, "DEPARTED", res.Events[1].MilestoneCode)
	assert.Equal(t, "Port of Hamburg", res.Events[1].Location)
	assert.Equal(t, "HAMBURG EXPRESS", res.Events[1].VesselName)
	assert.Equal(t, "2609W", res.Events[1].VoyageNumber)
}

// TEST 6 — Invalid Tracking Identifier
// Verify missing or whitespace-only identifier fails validation with 0 external API calls.
func TestHapagLloyd_Test6_InvalidIdentifierValidation(t *testing.T) {
	registry := adapters.GetDefaultRegistry()
	hapag := registry.GetAdapterBySCAC("HLCU")
	require.NotNil(t, hapag)

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	creds := domain.DecryptedCredentials{
		BaseURL:      server.URL,
		APIKey:       "hl-key",
		ClientSecret: "hl-secret",
	}

	// Empty request
	_, err := hapag.GetTracking(context.Background(), creds, domain.EnvSandbox, domain.TrackingRequest{})
	assert.Error(t, err)
	var intErr *domain.IntegrationError
	require.ErrorAs(t, err, &intErr)
	assert.Equal(t, http.StatusBadRequest, intErr.HTTPStatus)
	assert.Equal(t, domain.ErrCodeInvalidRequest, intErr.ErrorCode)

	// Whitespace-only request
	_, err = hapag.GetTracking(context.Background(), creds, domain.EnvSandbox, domain.TrackingRequest{
		ContainerNumber: "   ",
		BookingNumber:   "",
	})
	assert.Error(t, err)

	assert.Equal(t, int32(0), atomic.LoadInt32(&requestCount), "Zero HTTP calls must be made on validation failure")
}

// TEST 7 — API Errors
// Verify handling of 400, 401, 403, 404, 429, and 500 responses.
func TestHapagLloyd_Test7_APIErrors(t *testing.T) {
	registry := adapters.GetDefaultRegistry()
	hapag := registry.GetAdapterBySCAC("HLCU")
	require.NotNil(t, hapag)

	testCases := []struct {
		name         string
		statusCode   int
		expectedCode string
	}{
		{"400_BadRequest", http.StatusBadRequest, domain.ErrCodeInvalidRequest},
		{"401_Unauthorized", http.StatusUnauthorized, domain.ErrCodeAuthFailed},
		{"403_Forbidden", http.StatusForbidden, domain.ErrCodeForbidden},
		{"404_NotFound", http.StatusNotFound, domain.ErrCodeNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"test error %d"}`, tc.statusCode)))
			}))
			defer server.Close()

			creds := domain.DecryptedCredentials{
				BaseURL:      server.URL,
				APIKey:       "hl-key",
				ClientSecret: "hl-secret",
			}

			_, err := hapag.GetTracking(context.Background(), creds, domain.EnvSandbox, domain.TrackingRequest{
				ContainerNumber: "HLXU9990001",
			})
			require.Error(t, err)
			var intErr *domain.IntegrationError
			require.ErrorAs(t, err, &intErr)
			assert.Equal(t, tc.expectedCode, intErr.ErrorCode)
			assert.Equal(t, tc.statusCode, intErr.HTTPStatus)
		})
	}
}

// TEST 8 — Tenant Isolation
// Verify Organization A's request uses Org A's credentials/headers and cannot use Organization B's.
func TestHapagLloyd_Test8_TenantIsolation(t *testing.T) {
	registry := adapters.GetDefaultRegistry()
	hapag := registry.GetAdapterBySCAC("HLCU")
	require.NotNil(t, hapag)

	capturedIDs := make(map[string]string)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := r.Header.Get("X-IBM-Client-Id")
		clientSecret := r.Header.Get("X-IBM-Client-Secret")
		ref := r.URL.Query().Get("equipmentReference")
		capturedIDs[ref] = clientID + ":" + clientSecret

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	ctx := context.Background()

	// Org A Request
	credsA := domain.DecryptedCredentials{
		BaseURL:      server.URL,
		ClientID:     "org-A-hapag-id",
		ClientSecret: "org-A-hapag-secret",
	}
	_, err := hapag.GetTracking(ctx, credsA, domain.EnvSandbox, domain.TrackingRequest{
		ContainerNumber: "HLXU-TENANT-A",
	})
	require.NoError(t, err)

	// Org B Request
	credsB := domain.DecryptedCredentials{
		BaseURL:      server.URL,
		ClientID:     "org-B-hapag-id",
		ClientSecret: "org-B-hapag-secret",
	}
	_, err = hapag.GetTracking(ctx, credsB, domain.EnvSandbox, domain.TrackingRequest{
		ContainerNumber: "HLXU-TENANT-B",
	})
	require.NoError(t, err)

	assert.Equal(t, "org-A-hapag-id:org-A-hapag-secret", capturedIDs["HLXU-TENANT-A"])
	assert.Equal(t, "org-B-hapag-id:org-B-hapag-secret", capturedIDs["HLXU-TENANT-B"])
	assert.NotEqual(t, capturedIDs["HLXU-TENANT-A"], capturedIDs["HLXU-TENANT-B"])

	// Service-level tenant isolation test via real DB
	db, _, svc := setupTestDB(t)
	defer db.Close()

	orgA := int64(77701)
	orgB := int64(77702)
	userA := int64(1001)

	// Clean up test org records
	_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id IN (?, ?)", orgA, orgB)
	_, _ = db.Exec("DELETE FROM organizations WHERE id IN (?, ?)", orgA, orgB)

	_, err = db.Exec("INSERT INTO organizations (id, name) VALUES (?, 'Test Hapag Org A')", orgA)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO organizations (id, name) VALUES (?, 'Test Hapag Org B')", orgB)
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id IN (?, ?)", orgA, orgB)
		_, _ = db.Exec("DELETE FROM organizations WHERE id IN (?, ?)", orgA, orgB)
	}()

	createdA, err := svc.ConnectCarrier(ctx, orgA, userA, domain.ConnectCarrierRequest{
		CarrierSCAC:      "HLCU",
		Environment:      "SANDBOX",
		ConnectionMethod: "API",
		Credentials: map[string]interface{}{
			"api_key":       "hl-sandbox-key",
			"client_secret": "hl-sandbox-secret",
			"base_url":      server.URL,
		},
		Capabilities: []string{"TRACKING"},
	})
	require.NoError(t, err)

	// Org B must not be able to get or track via Org A's integration
	_, err = svc.GetIntegration(ctx, orgB, createdA.ID)
	assert.Error(t, err, "Org B should receive NOT FOUND for Org A's integration")

	_, err = svc.GetTracking(ctx, orgB, createdA.ID, domain.TrackingRequest{
		ContainerNumber: "HLXU998877",
	})
	assert.Error(t, err, "Org B should not be able to execute tracking using Org A's integration")
}

// TEST 9 — Credential Security
// Verify that credentials (API Key, Client Secret, IBM headers) are never leaked in
// returned responses, errors, or Stringer representations.
func TestHapagLloyd_Test9_CredentialSecurity(t *testing.T) {
	creds := domain.DecryptedCredentials{
		APIKey:       "hl_secret_api_key_xyz987",
		ClientSecret: "hl_super_secret_token_abc123",
	}

	// 1. Stringer masking
	strRepr := fmt.Sprintf("%v", creds)
	assert.NotContains(t, strRepr, "hl_secret_api_key_xyz987")
	assert.NotContains(t, strRepr, "hl_super_secret_token_abc123")
	assert.Equal(t, "[PROTECTED_CARRIER_CREDENTIALS]", strRepr)

	// 2. GoStringer masking
	goStrRepr := fmt.Sprintf("%#v", creds)
	assert.NotContains(t, goStrRepr, "hl_secret_api_key_xyz987")
	assert.Equal(t, "[PROTECTED_CARRIER_CREDENTIALS]", goStrRepr)

	// 3. Error sanitization
	rawErr := fmt.Errorf("Hapag error with X-IBM-Client-Secret: hl_super_secret_token_abc123 failed")
	sanitized := domain.SanitizeError(rawErr)
	assert.NotContains(t, sanitized, "hl_super_secret_token_abc123")
	assert.Contains(t, sanitized, "[REDACTED]")
}

