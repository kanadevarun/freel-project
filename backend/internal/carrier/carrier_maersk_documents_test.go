package carrier_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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
// Task 3.5: Maersk Documents Workflow Tests (9 Required Scenarios)
// ─────────────────────────────────────────────────────────────────────────────

// TEST 1 — Maersk GetDocuments
// Mock /oauth2/access_token
// Then mock /transportation-documents/v1/{ref}
// Verify: OAuth token acquired, Consumer-Key sent, Bearer token sent, correct endpoint, response parsed.
func TestMaerskDocuments_Test1_GetDocumentsSuccess(t *testing.T) {
	var capturedAuth, capturedConsumerKey, capturedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "valid-token-docs-1",
				"token_type": "Bearer",
				"expires_in": 3600
			}`))
			return
		}

		if strings.HasPrefix(r.URL.Path, "/transportation-documents/v1/") {
			capturedAuth = r.Header.Get("Authorization")
			capturedConsumerKey = r.Header.Get("Consumer-Key")
			capturedPath = r.URL.Path

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"documentID": "DOC-MSK-BOL-101",
					"documentType": "Bill of Lading",
					"documentNumber": "MSKU987654321",
					"downloadUrl": "https://api.maersk.com/documents/v1/download/DOC-MSK-BOL-101",
					"issueDate": "2026-10-01T10:00:00Z"
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
		APIKey:    "test-consumer-key-docs",
		APISecret: "test-api-secret-docs",
		BaseURL:   server.URL,
	}

	req := domain.DocumentRequest{
		CarrierSCAC:   "MAEU",
		BookingNumber: "MSKU987654321",
	}

	docs, err := adapter.GetDocuments(ctx, creds, domain.EnvSandbox, req)
	require.NoError(t, err)
	require.Len(t, docs, 1)

	assert.Equal(t, "Bearer valid-token-docs-1", capturedAuth)
	assert.Equal(t, "test-consumer-key-docs", capturedConsumerKey)
	assert.Equal(t, "/transportation-documents/v1/MSKU987654321", capturedPath)

	assert.Equal(t, "DOC-MSK-BOL-101", docs[0].DocumentID)
	assert.Equal(t, "MAEU", docs[0].CarrierSCAC)
	assert.Equal(t, "BILL_OF_LADING", docs[0].DocumentType)
	assert.Equal(t, "MSKU987654321", docs[0].Reference)
	assert.Equal(t, "https://api.maersk.com/documents/v1/download/DOC-MSK-BOL-101", docs[0].DocumentURL)
	assert.Equal(t, "application/pdf", docs[0].MimeType)
	assert.False(t, docs[0].CreatedAt.IsZero())
}

// TEST 2 — Document normalization
// Return representative Maersk response (multiple documents, various types).
// Verify normalized fields: document type, document number, download URL, issue date.
func TestMaerskDocuments_Test2_DocumentNormalization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "token-norm", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/transportation-documents/v1/") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"documentID": "BOL-001",
					"documentType": "Bill of Lading",
					"documentNumber": "MSK-BOL-7711",
					"downloadUrl": "https://api.maersk.com/download/bol-7711",
					"mimeType": "application/pdf",
					"issueDate": "2026-09-15T08:30:00Z"
				},
				{
					"document_id": "SWB-002",
					"document_type": "Sea Waybill",
					"document_number": "MSK-SWB-7712",
					"download_url": "https://api.maersk.com/download/swb-7712",
					"mime_type": "application/pdf",
					"issue_date": "2026-09-16T14:00:00Z"
				},
				{
					"id": "BC-003",
					"type": "Booking Confirmation",
					"reference": "MSK-BC-7713",
					"url": "https://api.maersk.com/download/bc-7713",
					"contentType": "application/pdf",
					"createdAt": "2026-09-10T12:00:00Z"
				}
			]`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()
	creds := domain.DecryptedCredentials{APIKey: "key", APISecret: "secret", BaseURL: server.URL}

	docs, err := adapter.GetDocuments(ctx, creds, domain.EnvSandbox, domain.DocumentRequest{
		BookingNumber: "BKG-7711",
	})
	require.NoError(t, err)
	require.Len(t, docs, 3)

	// Doc 1: Bill of Lading
	assert.Equal(t, "BOL-001", docs[0].DocumentID)
	assert.Equal(t, "BILL_OF_LADING", docs[0].DocumentType)
	assert.Equal(t, "MSK-BOL-7711", docs[0].Reference)
	assert.Equal(t, "https://api.maersk.com/download/bol-7711", docs[0].DocumentURL)
	assert.Equal(t, "application/pdf", docs[0].MimeType)
	assert.Equal(t, 2026, docs[0].CreatedAt.Year())
	assert.Equal(t, time.September, docs[0].CreatedAt.Month())

	// Doc 2: Sea Waybill
	assert.Equal(t, "SWB-002", docs[1].DocumentID)
	assert.Equal(t, "SEA_WAYBILL", docs[1].DocumentType)
	assert.Equal(t, "MSK-SWB-7712", docs[1].Reference)
	assert.Equal(t, "https://api.maersk.com/download/swb-7712", docs[1].DocumentURL)

	// Doc 3: Booking Confirmation
	assert.Equal(t, "BC-003", docs[2].DocumentID)
	assert.Equal(t, "BOOKING_CONFIRMATION", docs[2].DocumentType)
	assert.Equal(t, "MSK-BC-7713", docs[2].Reference)
	assert.Equal(t, "https://api.maersk.com/download/bc-7713", docs[2].DocumentURL)
}

// TEST 3 — Empty documents
// Mock empty response ([]).
// Verify: successful response, empty list, no error.
func TestMaerskDocuments_Test3_EmptyDocuments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "token-empty", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/transportation-documents/v1/") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()
	creds := domain.DecryptedCredentials{APIKey: "key", APISecret: "secret", BaseURL: server.URL}

	docs, err := adapter.GetDocuments(ctx, creds, domain.EnvSandbox, domain.DocumentRequest{
		BookingNumber: "MSKU-EMPTY-REF",
	})
	require.NoError(t, err)
	assert.NotNil(t, docs)
	assert.Empty(t, docs, "Empty document list must return empty slice and no error")
}

// TEST 4 — 401 Unauthorized & Token Refresh
// First request -> 401. Token invalidated.
// OAuth refresh -> fresh token.
// Second request -> 200 OK.
// Verify existing Task 3.2 refresh mechanism is reused without duplication.
func TestMaerskDocuments_Test4_401AutomaticTokenRefresh(t *testing.T) {
	var oauthCalls int32
	var docsCalls int32
	var capturedTokens []string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			n := atomic.AddInt32(&oauthCalls, 1)
			w.WriteHeader(http.StatusOK)
			if n == 1 {
				w.Write([]byte(`{"access_token": "token-stale-A", "token_type": "Bearer", "expires_in": 3600}`))
			} else {
				w.Write([]byte(`{"access_token": "token-fresh-B", "token_type": "Bearer", "expires_in": 3600}`))
			}
			return
		}

		if strings.HasPrefix(r.URL.Path, "/transportation-documents/v1/") {
			n := atomic.AddInt32(&docsCalls, 1)
			mu.Lock()
			capturedTokens = append(capturedTokens, r.Header.Get("Authorization"))
			mu.Unlock()

			if n == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "invalid_token", "error_description": "Token expired"}`))
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"documentID": "DOC-REFRESH-OK",
					"documentType": "BILL_OF_LADING",
					"documentNumber": "MSK-401-DOC",
					"downloadUrl": "https://api.maersk.com/download/refreshed"
				}
			]`))
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	token.GetDefaultTokenManager().Invalidate("MAERSK", "test-consumer-401-docs", domain.EnvSandbox)
	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()
	creds := domain.DecryptedCredentials{
		APIKey:    "test-consumer-401-docs",
		APISecret: "test-secret-401-docs",
		BaseURL:   server.URL,
	}

	docs, err := adapter.GetDocuments(ctx, creds, domain.EnvSandbox, domain.DocumentRequest{
		BookingNumber: "MSK-401-DOC",
	})
	require.NoError(t, err)
	require.Len(t, docs, 1)
	assert.Equal(t, "DOC-REFRESH-OK", docs[0].DocumentID)

	assert.Equal(t, int32(2), atomic.LoadInt32(&oauthCalls), "Expected 2 OAuth token requests")
	assert.Equal(t, int32(2), atomic.LoadInt32(&docsCalls), "Expected 2 document requests (initial + 401 retry)")

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, capturedTokens, 2)
	assert.Equal(t, "Bearer token-stale-A", capturedTokens[0])
	assert.Equal(t, "Bearer token-fresh-B", capturedTokens[1])
}

// TEST 5 — 403 Forbidden
// Verify permission failure is returned safely.
func TestMaerskDocuments_Test5_403Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "token-test5", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/transportation-documents/v1/") {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "forbidden", "message": "Access to documents restricted for this account"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()
	creds := domain.DecryptedCredentials{APIKey: "key", APISecret: "secret", BaseURL: server.URL}

	docs, err := adapter.GetDocuments(ctx, creds, domain.EnvSandbox, domain.DocumentRequest{
		BookingNumber: "MSKU-FORBIDDEN-REF",
	})
	require.Error(t, err)
	assert.Nil(t, docs)

	var intErr *domain.IntegrationError
	require.ErrorAs(t, err, &intErr)
	assert.Equal(t, domain.ErrCodeForbidden, intErr.ErrorCode)
	assert.Equal(t, http.StatusForbidden, intErr.HTTPStatus)
}

// TEST 6 — 404 Not Found
// Verify missing reference is handled correctly.
func TestMaerskDocuments_Test6_404NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "token-test6", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/transportation-documents/v1/") {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "not_found", "message": "Transport document reference not found"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()
	creds := domain.DecryptedCredentials{APIKey: "key", APISecret: "secret", BaseURL: server.URL}

	docs, err := adapter.GetDocuments(ctx, creds, domain.EnvSandbox, domain.DocumentRequest{
		BookingNumber: "NON_EXISTENT_REF_999",
	})
	require.Error(t, err)
	assert.Nil(t, docs)

	var intErr *domain.IntegrationError
	require.ErrorAs(t, err, &intErr)
	assert.Equal(t, domain.ErrCodeNotFound, intErr.ErrorCode)
	assert.Equal(t, http.StatusNotFound, intErr.HTTPStatus)
}

// TEST 7 — Unsupported capability
// Use a carrier without Documents support (e.g. HapagLloydAdapter).
// Verify: zero HTTP requests, standardized unsupported-operation response.
func TestMaerskDocuments_Test7_UnsupportedCapability(t *testing.T) {
	var httpCalls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&httpCalls, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	hapagAdapter := adapters.NewHapagLloydAdapter()
	ctx := context.Background()
	creds := domain.DecryptedCredentials{APIKey: "key", APISecret: "secret", BaseURL: server.URL}

	docs, err := hapagAdapter.GetDocuments(ctx, creds, domain.EnvSandbox, domain.DocumentRequest{
		BookingNumber: "HLCU-123456",
	})
	require.Error(t, err)
	assert.Nil(t, docs)
	assert.Equal(t, int32(0), atomic.LoadInt32(&httpCalls), "Zero HTTP calls must be made for unsupported capability")
	assert.Contains(t, err.Error(), "DOCUMENTS")
}

// TEST 8 — Tenant isolation
// Verify the document request uses the correct organization's integration and credentials.
func TestMaerskDocuments_Test8_TenantIsolation(t *testing.T) {
	var mu sync.Mutex
	capturedKeys := make(map[string]string)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/oauth2/access_token" {
			_ = r.ParseForm()
			clientID := r.FormValue("client_id")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"access_token": "token-for-%s", "token_type": "Bearer", "expires_in": 3600}`, clientID)))
			return
		}

		if strings.HasPrefix(r.URL.Path, "/transportation-documents/v1/") {
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

	// Org A (Tenant 1)
	credsOrgA := domain.DecryptedCredentials{
		APIKey:    "org-A-consumer-key",
		APISecret: "org-A-secret",
		BaseURL:   server.URL,
	}
	_, err := adapter.GetDocuments(ctx, credsOrgA, domain.EnvSandbox, domain.DocumentRequest{
		BookingNumber: "BKG-ORG-A",
	})
	require.NoError(t, err)

	// Org B (Tenant 2)
	credsOrgB := domain.DecryptedCredentials{
		APIKey:    "org-B-consumer-key",
		APISecret: "org-B-secret",
		BaseURL:   server.URL,
	}
	_, err = adapter.GetDocuments(ctx, credsOrgB, domain.EnvSandbox, domain.DocumentRequest{
		BookingNumber: "BKG-ORG-B",
	})
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, "Bearer token-for-org-A-consumer-key", capturedKeys["org-A-consumer-key"])
	assert.Equal(t, "Bearer token-for-org-B-consumer-key", capturedKeys["org-B-consumer-key"])
	assert.NotEqual(t, capturedKeys["org-A-consumer-key"], capturedKeys["org-B-consumer-key"], "Tenants must have isolated credentials and bearer tokens")
}

// TEST 9 — Security
// Verify: no access token in response, no API secret in response, no credentials in logs.
func TestMaerskDocuments_Test9_Security(t *testing.T) {
	secretKey := "SUPER_SECRET_CLIENT_KEY_8888"
	secretToken := "SENSITIVE_BEARER_TOKEN_9999"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/oauth2/access_token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"access_token": "%s", "token_type": "Bearer", "expires_in": 3600}`, secretToken)))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/transportation-documents/v1/") {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": "Internal error evaluating secret %s and token %s"}`, secretKey, secretToken)))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	adapter := adapters.NewMaerskAdapter()
	ctx := context.Background()
	creds := domain.DecryptedCredentials{
		APIKey:    "test-consumer-security",
		APISecret: secretKey,
		BaseURL:   server.URL,
	}

	docs, err := adapter.GetDocuments(ctx, creds, domain.EnvSandbox, domain.DocumentRequest{
		BookingNumber: "BKG-SEC-101",
	})
	require.Error(t, err)
	assert.Nil(t, docs)

	errString := err.Error()
	assert.NotContains(t, errString, secretKey, "Client secret must never leak in error messages")
	assert.NotContains(t, errString, secretToken, "Access token must never leak in error messages")
	assert.NotContains(t, errString, "Bearer ", "Authorization header must never leak in error messages")
}
