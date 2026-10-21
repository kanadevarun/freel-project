package carrier_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/freel/backend/internal/carrier/adapters"
	"github.com/freel/backend/internal/carrier/adapters/token"
	"github.com/freel/backend/internal/carrier/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Task 3.1: Maersk Authentication & Bearer Token Verification Tests (10 Required Tests)
// ─────────────────────────────────────────────────────────────────────────────

// Test 1 — OAuth token request
// Verify the Maersk adapter/TokenManager sends:
// - correct OAuth endpoint (/oauth2/access_token)
// - POST
// - correct Content-Type (application/x-www-form-urlencoded)
// - grant_type=client_credentials
// - correct client credentials
func TestMaerskAuth_Test1_OAuthTokenRequest(t *testing.T) {
	var receivedMethod, receivedPath, receivedContentType, receivedGrantType, receivedClientID, receivedSecret string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedPath = r.URL.Path
		receivedContentType = r.Header.Get("Content-Type")

		bodyBytes, _ := io.ReadAll(r.Body)
		bodyStr := string(bodyBytes)

		// Parse form data
		_ = r.ParseForm()
		receivedGrantType = r.FormValue("grant_type")
		receivedClientID = r.FormValue("client_id")
		receivedSecret = r.FormValue("client_secret")

		// Fallback parse if ParseForm already consumed or urlencoded string check
		if receivedGrantType == "" && strings.Contains(bodyStr, "grant_type=client_credentials") {
			receivedGrantType = "client_credentials"
		}
		if receivedClientID == "" && strings.Contains(bodyStr, "client_id=test-client-id-1") {
			receivedClientID = "test-client-id-1"
		}
		if receivedSecret == "" && strings.Contains(bodyStr, "client_secret=test-client-secret-1") {
			receivedSecret = "test-client-secret-1"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"access_token": "token-test-1",
			"token_type": "Bearer",
			"expires_in": 3600
		}`))
	}))
	defer server.Close()

	mgr := token.GetDefaultTokenManager()
	ctx := context.Background()

	tokenReq := token.TokenRequest{
		TokenURL:     server.URL + "/oauth2/access_token",
		ClientID:     "test-client-id-1",
		ClientSecret: "test-client-secret-1",
		ProviderCode: "MAERSK",
		Environment:  domain.EnvSandbox,
	}

	tok, err := mgr.GetToken(ctx, tokenReq)
	require.NoError(t, err)
	assert.Equal(t, "token-test-1", tok)

	assert.Equal(t, "POST", receivedMethod, "OAuth request must be HTTP POST")
	assert.Equal(t, "/oauth2/access_token", receivedPath, "OAuth request must hit /oauth2/access_token")
	assert.Equal(t, "application/x-www-form-urlencoded", receivedContentType, "Content-Type must be application/x-www-form-urlencoded")
	assert.Equal(t, "client_credentials", receivedGrantType, "grant_type must be client_credentials")
	assert.Equal(t, "test-client-id-1", receivedClientID, "client_id must match credentials")
	assert.Equal(t, "test-client-secret-1", receivedSecret, "client_secret must match credentials")
}

// Test 2 — OAuth token response
// Mock standard OAuth token response and verify token parsing
func TestMaerskAuth_Test2_OAuthTokenResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"access_token": "test-access-token-parse-check",
			"token_type": "Bearer",
			"expires_in": 3600
		}`))
	}))
	defer server.Close()

	mgr := token.GetDefaultTokenManager()
	ctx := context.Background()

	tokenReq := token.TokenRequest{
		TokenURL:     server.URL + "/oauth2/access_token",
		ClientID:     "client-id-test-2",
		ClientSecret: "client-secret-test-2",
		ProviderCode: "MAERSK",
		Environment:  domain.EnvSandbox,
	}

	tok, err := mgr.GetToken(ctx, tokenReq)
	require.NoError(t, err)
	assert.Equal(t, "test-access-token-parse-check", tok, "access_token should be correctly decoded from JSON response")
}

// Test 3 — Tracking Authorization
// Mock the Maersk tracking endpoint and verify outgoing request contains:
// Authorization: Bearer <access-token> and Consumer-Key
func TestMaerskAuth_Test3_TrackingAuthorization(t *testing.T) {
	var trackingAuthHeader, trackingConsumerKey, trackingXAPIKey string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "bearer-token-tracking-99",
				"token_type": "Bearer",
				"expires_in": 3600
			}`))
			return
		}

		if strings.HasPrefix(r.URL.Path, "/track-and-trace/v2/events") {
			trackingAuthHeader = r.Header.Get("Authorization")
			trackingConsumerKey = r.Header.Get("Consumer-Key")
			trackingXAPIKey = r.Header.Get("X-API-Key")

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
		APIKey:    "maersk-consumer-key-trk",
		APISecret: "maersk-secret-trk",
		BaseURL:   server.URL,
	}

	_, err := adapter.GetTracking(ctx, creds, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU1234567",
	})
	require.NoError(t, err)

	assert.Equal(t, "Bearer bearer-token-tracking-99", trackingAuthHeader, "Tracking request must contain OAuth2 Bearer token")
	assert.Equal(t, "maersk-consumer-key-trk", trackingConsumerKey, "Tracking request must contain Consumer-Key header")
	assert.Equal(t, "maersk-consumer-key-trk", trackingXAPIKey, "Tracking request must contain X-API-Key header")
}

// Test 4 — Rates Authorization
// Mock the rates endpoint and verify identical authentication headers are present
func TestMaerskAuth_Test4_RatesAuthorization(t *testing.T) {
	var ratesAuthHeader, ratesConsumerKey string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "bearer-token-rates-88",
				"token_type": "Bearer",
				"expires_in": 3600
			}`))
			return
		}

		if r.URL.Path == "/rates/v1/spot-rates" {
			ratesAuthHeader = r.Header.Get("Authorization")
			ratesConsumerKey = r.Header.Get("Consumer-Key")

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
		APIKey:    "maersk-consumer-key-rates",
		APISecret: "maersk-secret-rates",
		BaseURL:   server.URL,
	}

	_, err := adapter.GetRates(ctx, creds, domain.EnvSandbox, domain.RateRequest{
		OriginPort:      "INNSA",
		DestinationPort: "NLRTM",
		EquipmentType:   "40HC",
	})
	require.NoError(t, err)

	assert.Equal(t, "Bearer bearer-token-rates-88", ratesAuthHeader, "Rates request must contain OAuth2 Bearer token")
	assert.Equal(t, "maersk-consumer-key-rates", ratesConsumerKey, "Rates request must contain Consumer-Key header")
}

// Test 5 — Booking Authorization
// Verify CreateBooking receives correct Authorization and Consumer-Key headers
func TestMaerskAuth_Test5_BookingAuthorization(t *testing.T) {
	var bookingAuthHeader, bookingConsumerKey string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "bearer-token-booking-77",
				"token_type": "Bearer",
				"expires_in": 3600
			}`))
			return
		}

		if r.URL.Path == "/booking/v1/bookings" {
			bookingAuthHeader = r.Header.Get("Authorization")
			bookingConsumerKey = r.Header.Get("Consumer-Key")

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"booking_reference": "BKG-MAEU-7712",
				"booking_status": "CONFIRMED"
			}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()

	creds := domain.DecryptedCredentials{
		APIKey:    "maersk-consumer-key-bkg",
		APISecret: "maersk-secret-bkg",
		BaseURL:   server.URL,
	}

	res, err := adapter.CreateBooking(ctx, creds, domain.EnvSandbox, domain.BookingRequest{
		CarrierSCAC:     "MAEU",
		OriginPort:      "INNSA",
		DestinationPort: "NLRTM",
		EquipmentType:   "40HC",
		Quantity:        1,
		ShipperName:     "Global Freight LLC",
		ConsigneeName:   "Euro Importers BV",
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, "Bearer bearer-token-booking-77", bookingAuthHeader, "Booking request must contain OAuth2 Bearer token")
	assert.Equal(t, "maersk-consumer-key-bkg", bookingConsumerKey, "Booking request must contain Consumer-Key header")
}

// Test 6 — Documents Authorization
// Verify GetDocuments receives correct Authorization and Consumer-Key headers
func TestMaerskAuth_Test6_DocumentsAuthorization(t *testing.T) {
	var docsAuthHeader, docsConsumerKey string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "bearer-token-docs-66",
				"token_type": "Bearer",
				"expires_in": 3600
			}`))
			return
		}

		if strings.HasPrefix(r.URL.Path, "/transportation-documents/v1/") || strings.HasPrefix(r.URL.Path, "/documents/v1/bookings/") {
			docsAuthHeader = r.Header.Get("Authorization")
			docsConsumerKey = r.Header.Get("Consumer-Key")

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
		APIKey:    "maersk-consumer-key-docs",
		APISecret: "maersk-secret-docs",
		BaseURL:   server.URL,
	}

	_, err := adapter.GetDocuments(ctx, creds, domain.EnvSandbox, domain.DocumentRequest{
		BookingNumber: "BKG-MAEU-7712",
	})
	require.NoError(t, err)

	assert.Equal(t, "Bearer bearer-token-docs-66", docsAuthHeader, "Documents request must contain OAuth2 Bearer token")
	assert.Equal(t, "maersk-consumer-key-docs", docsConsumerKey, "Documents request must contain Consumer-Key header")
}

// Test 7 — Connection Test (Success)
// Verify valid mocked credentials cause the connection test to succeed through the actual OAuth flow
func TestMaerskAuth_Test7_ConnectionTest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "bearer-token-conn-ok",
				"token_type": "Bearer",
				"expires_in": 3600
			}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()

	// Use standard UI credential naming convention: api_key + api_secret
	creds := domain.DecryptedCredentials{
		APIKey:    "maersk-test-consumer-key-ui",
		APISecret: "maersk-test-secret-ui",
		BaseURL:   server.URL,
	}

	res, err := adapter.TestConnection(ctx, creds, domain.EnvSandbox)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.True(t, res.Success, "Connection test must report Success=true with valid OAuth credentials")
	assert.Equal(t, http.StatusOK, res.HTTPStatus)
	assert.Contains(t, res.Message, "Successfully authenticated with Maersk Developer Portal OAuth2 Gateway")
}

// Test 8 — Wrong Credentials
// Mock OAuth failure and verify connection test fails safely without exposing secrets
func TestMaerskAuth_Test8_WrongCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{
				"error": "invalid_client",
				"error_description": "Bad client credentials"
			}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()

	testSecret := "super-secret-password-xyz-991"
	creds := domain.DecryptedCredentials{
		APIKey:    "invalid-consumer-key",
		APISecret: testSecret,
		BaseURL:   server.URL,
	}

	res, err := adapter.TestConnection(ctx, creds, domain.EnvSandbox)
	require.NoError(t, err, "TestConnection returns error details inside TestConnectionResult struct")
	require.NotNil(t, res)

	assert.False(t, res.Success, "Connection test must report Success=false on rejected credentials")
	assert.Equal(t, http.StatusUnauthorized, res.HTTPStatus)
	assert.Equal(t, domain.ErrCodeAuthFailed, res.ErrorCode)

	// Verify no secret leak in message or error
	assert.NotContains(t, res.Message, testSecret, "Connection test failure message must NEVER expose secret")
	assert.NotContains(t, res.Message, "invalid-consumer-key", "Message should not contain raw keys")
}

// Test 9 — Token Caching
// Verify repeated Maersk requests reuse the cached token instead of requesting a new token every time
func TestMaerskAuth_Test9_TokenCaching(t *testing.T) {
	var oauthCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			atomic.AddInt32(&oauthCalls, 1)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "cached-token-maersk-reuse",
				"token_type": "Bearer",
				"expires_in": 3600
			}`))
			return
		}

		if strings.HasPrefix(r.URL.Path, "/track-and-trace/v2/events") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
			return
		}

		if r.URL.Path == "/rates/v1/spot-rates" {
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
		APIKey:    "maersk-cache-consumer-key",
		APISecret: "maersk-cache-secret",
		BaseURL:   server.URL,
	}

	// First operational call: Tracking -> should fetch fresh token
	_, err := adapter.GetTracking(ctx, creds, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU8888888",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&oauthCalls), "Initial call should negotiate OAuth token once")

	// Second operational call: Rates -> should reuse cached token from memory
	_, err = adapter.GetRates(ctx, creds, domain.EnvSandbox, domain.RateRequest{
		OriginPort:      "INNSA",
		DestinationPort: "NLRTM",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&oauthCalls), "Subsequent call within TTL must reuse cached token")

	// Third operational call: Booking -> should reuse cached token
	_, _ = adapter.GetBooking(ctx, creds, domain.EnvSandbox, "BKG-12345")
	assert.Equal(t, int32(1), atomic.LoadInt32(&oauthCalls), "Third call within TTL must reuse cached token without network roundtrip")
}

// Test 10 — Credential Naming Compatibility
// Verify whichever supported credential naming conventions currently exist are correctly normalized
func TestMaerskAuth_Test10_CredentialNamingCompatibility(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			_ = r.ParseForm()
			clientID := r.FormValue("client_id")
			clientSecret := r.FormValue("client_secret")

			if clientID == "my-canonical-key" && clientSecret == "my-canonical-secret" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"access_token": "token-compat-canonical",
					"token_type": "Bearer",
					"expires_in": 3600
				}`))
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "unauthorized"}`))
			return
		}

		if strings.HasPrefix(r.URL.Path, "/track-and-trace/v2/events") {
			assert.Equal(t, "Bearer token-compat-canonical", r.Header.Get("Authorization"))
			assert.Equal(t, "my-canonical-key", r.Header.Get("Consumer-Key"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	ctx := context.Background()
	adapter := adapters.NewMaerskAdapter()

	// Case A: UI naming convention (api_key + api_secret)
	jsonUI := `{"api_key": "my-canonical-key", "api_secret": "my-canonical-secret", "base_url": "` + server.URL + `"}`
	var credsUI domain.DecryptedCredentials
	err := json.Unmarshal([]byte(jsonUI), &credsUI)
	require.NoError(t, err)

	assert.Equal(t, "my-canonical-key", credsUI.GetConsumerKey())
	assert.Equal(t, "my-canonical-secret", credsUI.GetAPISecret())

	_, err = adapter.GetTracking(ctx, credsUI, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU1111111",
	})
	require.NoError(t, err, "Case A (UI naming) must succeed and inject token")

	// Case B: Standard OAuth naming convention (client_id + client_secret)
	jsonOAuth := `{"client_id": "my-canonical-key", "client_secret": "my-canonical-secret", "base_url": "` + server.URL + `"}`
	var credsOAuth domain.DecryptedCredentials
	err = json.Unmarshal([]byte(jsonOAuth), &credsOAuth)
	require.NoError(t, err)

	assert.Equal(t, "my-canonical-key", credsOAuth.GetConsumerKey())
	assert.Equal(t, "my-canonical-secret", credsOAuth.GetAPISecret())

	_, err = adapter.GetTracking(ctx, credsOAuth, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU1111111",
	})
	require.NoError(t, err, "Case B (OAuth naming) must succeed and inject token")

	// Case C: Mixed convention (api_key + client_secret)
	jsonMixed := `{"api_key": "my-canonical-key", "client_secret": "my-canonical-secret", "base_url": "` + server.URL + `"}`
	var credsMixed domain.DecryptedCredentials
	err = json.Unmarshal([]byte(jsonMixed), &credsMixed)
	require.NoError(t, err)

	assert.Equal(t, "my-canonical-key", credsMixed.GetConsumerKey())
	assert.Equal(t, "my-canonical-secret", credsMixed.GetAPISecret())

	_, err = adapter.GetTracking(ctx, credsMixed, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU1111111",
	})
	require.NoError(t, err, "Case C (Mixed naming) must succeed and inject token")
}
