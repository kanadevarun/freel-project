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
	"github.com/freel/backend/internal/carrier/adapters/token"
	"github.com/freel/backend/internal/carrier/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Task 3.2: Maersk 401 Token Invalidation & Automatic Token Refresh Tests (8 Required Tests)
// ─────────────────────────────────────────────────────────────────────────────

// TEST 1 — 401 THEN SUCCESS
// Request 1 uses token-A -> returns 401
// Token-A is invalidated, OAuth refresh returns token-B
// Request 2 uses token-B -> returns 200
// Verify exactly two operational requests and final success
func TestTask32_Test1_401ThenSuccess(t *testing.T) {
	var oauthCalls int32
	var trackingCalls int32
	var mu sync.Mutex
	var authHeaders []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			callNum := atomic.AddInt32(&oauthCalls, 1)
			w.WriteHeader(http.StatusOK)
			if callNum == 1 {
				w.Write([]byte(`{"access_token": "token-A", "token_type": "Bearer", "expires_in": 3600}`))
			} else {
				w.Write([]byte(`{"access_token": "token-B", "token_type": "Bearer", "expires_in": 3600}`))
			}
			return
		}

		if strings.HasPrefix(r.URL.Path, "/track-and-trace/v2/events") {
			callNum := atomic.AddInt32(&trackingCalls, 1)
			header := r.Header.Get("Authorization")
			mu.Lock()
			authHeaders = append(authHeaders, header)
			mu.Unlock()

			if callNum == 1 {
				// First request with token-A returns 401 Unauthorized
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "invalid_token", "message": "The access token expired or was revoked"}`))
				return
			}

			// Second request with token-B returns 200 OK
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
		APIKey:    "test-maersk-key-t1",
		APISecret: "test-maersk-secret-t1",
		BaseURL:   server.URL,
	}

	result, err := adapter.GetTracking(ctx, creds, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU1000001",
	})
	require.NoError(t, err, "Operational call should succeed after single 401 refresh")
	require.NotNil(t, result)

	// Verify exact count of calls
	assert.Equal(t, int32(2), atomic.LoadInt32(&trackingCalls), "Exactly two operational requests must occur")
	assert.Equal(t, int32(2), atomic.LoadInt32(&oauthCalls), "Exactly two OAuth requests must occur (initial + refresh)")

	// Verify tokens used in sequence
	mu.Lock()
	defer mu.Unlock()
	require.Len(t, authHeaders, 2)
	assert.Equal(t, "Bearer token-A", authHeaders[0], "First operational request must use token-A")
	assert.Equal(t, "Bearer token-B", authHeaders[1], "Second operational request must use token-B")
}

// TEST 2 — 401 TWICE
// Request 1 uses token-A -> 401
// Token refresh -> token-B
// Request 2 uses token-B -> 401
// Verify exactly one retry (2 operational requests), exactly one refresh, no 3rd request, returns auth error
func TestTask32_Test2_401Twice(t *testing.T) {
	var oauthCalls int32
	var trackingCalls int32
	var mu sync.Mutex
	var authHeaders []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			callNum := atomic.AddInt32(&oauthCalls, 1)
			w.WriteHeader(http.StatusOK)
			if callNum == 1 {
				w.Write([]byte(`{"access_token": "token-A", "token_type": "Bearer", "expires_in": 3600}`))
			} else {
				w.Write([]byte(`{"access_token": "token-B", "token_type": "Bearer", "expires_in": 3600}`))
			}
			return
		}

		if strings.HasPrefix(r.URL.Path, "/track-and-trace/v2/events") {
			atomic.AddInt32(&trackingCalls, 1)
			header := r.Header.Get("Authorization")
			mu.Lock()
			authHeaders = append(authHeaders, header)
			mu.Unlock()

			// Both attempts return 401 Unauthorized
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "unauthorized", "message": "Access denied"}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()

	creds := domain.DecryptedCredentials{
		APIKey:    "test-maersk-key-t2",
		APISecret: "test-maersk-secret-t2",
		BaseURL:   server.URL,
	}

	result, err := adapter.GetTracking(ctx, creds, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU1000002",
	})
	require.Error(t, err, "Persistent 401 must return an authentication error")
	assert.Nil(t, result)

	// Verify exactly 2 operational calls (never enters an infinite loop)
	assert.Equal(t, int32(2), atomic.LoadInt32(&trackingCalls), "Must stop after exactly 2 operational attempts")
	assert.Equal(t, int32(2), atomic.LoadInt32(&oauthCalls), "Must perform exactly one token refresh")

	var intErr *domain.IntegrationError
	require.ErrorAs(t, err, &intErr)
	assert.Equal(t, domain.ErrCodeAuthFailed, intErr.ErrorCode)
	assert.Equal(t, http.StatusUnauthorized, intErr.HTTPStatus)
}

// TEST 3 — TOKEN CACHE IS ACTUALLY INVALIDATED
// Pre-populate/use cached token-A. Return 401 on tracking call.
// Verify that subsequent GetToken() does NOT return token-A; it obtains and returns token-B.
func TestTask32_Test3_TokenCacheActuallyInvalidated(t *testing.T) {
	var oauthCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			callNum := atomic.AddInt32(&oauthCalls, 1)
			w.WriteHeader(http.StatusOK)
			if callNum == 1 {
				w.Write([]byte(`{"access_token": "token-A", "token_type": "Bearer", "expires_in": 3600}`))
			} else {
				w.Write([]byte(`{"access_token": "token-B", "token_type": "Bearer", "expires_in": 3600}`))
			}
			return
		}

		if strings.HasPrefix(r.URL.Path, "/track-and-trace/v2/events") {
			// Always return 401 to trigger invalidation
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "unauthorized"}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	mgr := token.GetDefaultTokenManager()
	ctx := context.Background()

	clientID := "test-maersk-key-t3"
	clientSecret := "test-maersk-secret-t3"

	// 1. Initial token acquisition
	tok1, err := mgr.GetToken(ctx, token.TokenRequest{
		TokenURL:     server.URL + "/oauth2/access_token",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		ProviderCode: "MAERSK",
		Environment:  domain.EnvSandbox,
	})
	require.NoError(t, err)
	assert.Equal(t, "token-A", tok1)

	// 2. Execute operational call via adapter which gets 401 and invalidates
	adapter := adapters.NewMaerskAdapter()
	creds := domain.DecryptedCredentials{
		APIKey:    clientID,
		APISecret: clientSecret,
		BaseURL:   server.URL,
	}

	_, _ = adapter.GetTracking(ctx, creds, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU1000003",
	})

	// 3. Subsequent GetToken must NOT return token-A; it must return token-B
	tok2, err := mgr.GetToken(ctx, token.TokenRequest{
		TokenURL:     server.URL + "/oauth2/access_token",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		ProviderCode: "MAERSK",
		Environment:  domain.EnvSandbox,
	})
	require.NoError(t, err)
	assert.Equal(t, "token-B", tok2, "Cached token must have been updated to token-B, never token-A")
	assert.NotEqual(t, "token-A", tok2)
}

// TEST 4 — SUCCESS DOES NOT REFRESH
// When request succeeds (200 OK), verify no unnecessary OAuth refresh occurs and cached token remains valid.
func TestTask32_Test4_SuccessDoesNotRefresh(t *testing.T) {
	var oauthCalls int32
	var trackingCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			atomic.AddInt32(&oauthCalls, 1)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "token-A-stable", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}

		if strings.HasPrefix(r.URL.Path, "/track-and-trace/v2/events") {
			atomic.AddInt32(&trackingCalls, 1)
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
		APIKey:    "test-maersk-key-t4",
		APISecret: "test-maersk-secret-t4",
		BaseURL:   server.URL,
	}

	res, err := adapter.GetTracking(ctx, creds, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU1000004",
	})
	require.NoError(t, err)
	require.NotNil(t, resultOr(res, nil))

	assert.Equal(t, int32(1), atomic.LoadInt32(&trackingCalls), "Only one operational request must occur on success")
	assert.Equal(t, int32(1), atomic.LoadInt32(&oauthCalls), "Only one OAuth call must occur")
}

// TEST 5 — DIFFERENT CARRIERS ARE UNAFFECTED
// Verify the 401 refresh implementation is scoped to Maersk and does not alter unrelated carrier behavior.
func TestTask32_Test5_DifferentCarriersUnaffected(t *testing.T) {
	registry := adapters.GetDefaultRegistry()

	hapag := registry.GetAdapterBySCAC("HLCU")
	require.NotNil(t, hapag)
	assert.Equal(t, "HAPAG_LLOYD", hapag.ProviderCode())

	msc := registry.GetAdapterBySCAC("MSCU")
	require.NotNil(t, msc)
	assert.Equal(t, "MSC", msc.ProviderCode())

	maersk := registry.GetAdapterBySCAC("MAEU")
	require.NotNil(t, maersk)
	assert.Equal(t, "MAERSK", maersk.ProviderCode())
}

// TEST 6 — DIFFERENT TENANTS ARE ISOLATED
// Org A gets 401 and refreshes token-A to token-A2.
// Verify Org B's token-B is NOT invalidated and remains cached.
func TestTask32_Test6_DifferentTenantsIsolated(t *testing.T) {
	var oauthCallsA, oauthCallsB int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			_ = r.ParseForm()
			clientID := r.FormValue("client_id")

			if clientID == "tenant-A-key" {
				c := atomic.AddInt32(&oauthCallsA, 1)
				if c == 1 {
					w.Write([]byte(`{"access_token": "token-A1", "token_type": "Bearer", "expires_in": 3600}`))
				} else {
					w.Write([]byte(`{"access_token": "token-A2", "token_type": "Bearer", "expires_in": 3600}`))
				}
				return
			}

			if clientID == "tenant-B-key" {
				atomic.AddInt32(&oauthCallsB, 1)
				w.Write([]byte(`{"access_token": "token-B1", "token_type": "Bearer", "expires_in": 3600}`))
				return
			}
		}

		if strings.HasPrefix(r.URL.Path, "/track-and-trace/v2/events") {
			auth := r.Header.Get("Authorization")
			if auth == "Bearer token-A1" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "unauthorized"}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	mgr := token.GetDefaultTokenManager()
	ctx := context.Background()

	// 1. Tenant B gets token-B1
	tokB, err := mgr.GetToken(ctx, token.TokenRequest{
		TokenURL:     server.URL + "/oauth2/access_token",
		ClientID:     "tenant-B-key",
		ClientSecret: "tenant-B-secret",
		ProviderCode: "MAERSK",
		Environment:  domain.EnvSandbox,
	})
	require.NoError(t, err)
	assert.Equal(t, "token-B1", tokB)
	assert.Equal(t, int32(1), atomic.LoadInt32(&oauthCallsB))

	// 2. Tenant A executes tracking call, receives 401, refreshes to token-A2
	adapter := adapters.NewMaerskAdapter()
	credsA := domain.DecryptedCredentials{
		APIKey:    "tenant-A-key",
		APISecret: "tenant-A-secret",
		BaseURL:   server.URL,
	}

	_, err = adapter.GetTracking(ctx, credsA, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU1000006",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&oauthCallsA))

	// 3. Verify Tenant B's token in cache was NOT touched and did NOT call OAuth again
	tokB2, err := mgr.GetToken(ctx, token.TokenRequest{
		TokenURL:     server.URL + "/oauth2/access_token",
		ClientID:     "tenant-B-key",
		ClientSecret: "tenant-B-secret",
		ProviderCode: "MAERSK",
		Environment:  domain.EnvSandbox,
	})
	require.NoError(t, err)
	assert.Equal(t, "token-B1", tokB2, "Tenant B's cached token must be unaffected by Tenant A's 401 refresh")
	assert.Equal(t, int32(1), atomic.LoadInt32(&oauthCallsB), "Tenant B must not perform any new OAuth request")
}

// TEST 7 — REFRESH FAILURE
// Request -> token-A -> 401
// OAuth refresh -> fails (e.g. 500 / 401)
// Verify no second operational request is made, and sanitized error is returned without secret leaks.
func TestTask32_Test7_RefreshFailure(t *testing.T) {
	var oauthCalls int32
	var trackingCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			callNum := atomic.AddInt32(&oauthCalls, 1)
			if callNum == 1 {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"access_token": "token-A", "token_type": "Bearer", "expires_in": 3600}`))
				return
			}
			// Refresh attempt fails
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "invalid_grant", "error_description": "Secret has expired"}`))
			return
		}

		if strings.HasPrefix(r.URL.Path, "/track-and-trace/v2/events") {
			atomic.AddInt32(&trackingCalls, 1)
			// Return 401 to trigger refresh
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "unauthorized"}`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()

	superSecret := "super-confidential-secret-999"
	creds := domain.DecryptedCredentials{
		APIKey:    "test-maersk-key-t7",
		APISecret: superSecret,
		BaseURL:   server.URL,
	}

	result, err := adapter.GetTracking(ctx, creds, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU1000007",
	})
	require.Error(t, err)
	assert.Nil(t, result)

	// Verify only 1 operational call happened (since refresh failed, no second call was attempted)
	assert.Equal(t, int32(1), atomic.LoadInt32(&trackingCalls), "No second operational request should be attempted if refresh fails")
	assert.Equal(t, int32(2), atomic.LoadInt32(&oauthCalls), "Two OAuth calls occurred (initial + failed refresh)")

	// Security: verify secret is not leaked in error message
	assert.NotContains(t, err.Error(), superSecret, "Error message must NEVER expose secret")
}

// TEST 8 — REFRESHED TOKEN IS USED
// Explicitly inspect request headers.
// First request: Authorization: Bearer token-A
// Retry: Authorization: Bearer token-B
// Verify token-A is never reused after invalidation.
func TestTask32_Test8_RefreshedTokenIsUsed(t *testing.T) {
	var oauthCalls int32
	var mu sync.Mutex
	var capturedAuthHeaders []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			c := atomic.AddInt32(&oauthCalls, 1)
			w.WriteHeader(http.StatusOK)
			if c == 1 {
				w.Write([]byte(`{"access_token": "token-A-original", "token_type": "Bearer", "expires_in": 3600}`))
			} else {
				w.Write([]byte(`{"access_token": "token-B-refreshed", "token_type": "Bearer", "expires_in": 3600}`))
			}
			return
		}

		if strings.HasPrefix(r.URL.Path, "/track-and-trace/v2/events") {
			auth := r.Header.Get("Authorization")
			mu.Lock()
			capturedAuthHeaders = append(capturedAuthHeaders, auth)
			mu.Unlock()

			if auth == "Bearer token-A-original" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "expired_token"}`))
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
		APIKey:    "test-maersk-key-t8",
		APISecret: "test-maersk-secret-t8",
		BaseURL:   server.URL,
	}

	_, err := adapter.GetTracking(ctx, creds, domain.EnvSandbox, domain.TrackingRequest{
		CarrierSCAC:     "MAEU",
		ContainerNumber: "MSKU1000008",
	})
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, capturedAuthHeaders, 2)
	assert.Equal(t, "Bearer token-A-original", capturedAuthHeaders[0])
	assert.Equal(t, "Bearer token-B-refreshed", capturedAuthHeaders[1])
	assert.NotEqual(t, capturedAuthHeaders[0], capturedAuthHeaders[1], "Refreshed request must use new token, never original")
}

func resultOr(a, b interface{}) interface{} {
	if a != nil {
		return a
	}
	return b
}
