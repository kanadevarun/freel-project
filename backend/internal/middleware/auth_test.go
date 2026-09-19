package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/freel/backend/internal/middleware"
)

// helperHandler returns 200 OK and inspects user context if present.
func mockProtectedHandler(capturedCtx *middleware.UserContext, executed *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*executed = true
		if uCtx, ok := middleware.GetUserContext(r.Context()); ok {
			*capturedCtx = uCtx
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
}

// ============================================================================
// SECTION 4: TEST-TOKEN TESTS ACROSS ENVIRONMENTS
// ============================================================================

func TestAuthMiddleware_DevelopmentEnvironment(t *testing.T) {
	guard := middleware.NewAuthMiddleware("", "", nil, middleware.WithEnvironment("development"))

	tests := []struct {
		name          string
		token         string
		testOrgHeader string
		expectedOrgID int64
		expectedRole  string
	}{
		{
			name:          "Default test-token yields Org 1 SUPER_ADMIN",
			token:         "test-token",
			expectedOrgID: 1,
			expectedRole:  "SUPER_ADMIN",
		},
		{
			name:          "test-token-org2 yields Org 2 SUPER_ADMIN",
			token:         "test-token-org2",
			expectedOrgID: 2,
			expectedRole:  "SUPER_ADMIN",
		},
		{
			name:          "test-token-org42 yields Org 42 SUPER_ADMIN",
			token:         "test-token-org42",
			expectedOrgID: 42,
			expectedRole:  "SUPER_ADMIN",
		},
		{
			name:          "test-token with X-Test-Org-ID header yields custom Org SUPER_ADMIN",
			token:         "test-token",
			testOrgHeader: "7",
			expectedOrgID: 7,
			expectedRole:  "SUPER_ADMIN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedCtx middleware.UserContext
			executed := false
			handler := guard.RequireAuth(mockProtectedHandler(&capturedCtx, &executed))

			req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			if tt.testOrgHeader != "" {
				req.Header.Set("X-Test-Org-ID", tt.testOrgHeader)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("Expected 200 OK in development, got %d: %s", rr.Code, rr.Body.String())
			}
			if !executed {
				t.Fatal("Expected downstream handler to be executed in development")
			}
			if capturedCtx.Role != tt.expectedRole {
				t.Errorf("Expected role %s, got %s", tt.expectedRole, capturedCtx.Role)
			}
			if capturedCtx.OrgID != tt.expectedOrgID {
				t.Errorf("Expected OrgID %d, got %d", tt.expectedOrgID, capturedCtx.OrgID)
			}
			if capturedCtx.UserID != 1 {
				t.Errorf("Expected UserID 1, got %d", capturedCtx.UserID)
			}
		})
	}
}

func TestAuthMiddleware_TestEnvironment(t *testing.T) {
	guard := middleware.NewAuthMiddleware("", "", nil, middleware.WithEnvironment("test"))

	var capturedCtx middleware.UserContext
	executed := false
	handler := guard.RequireAuth(mockProtectedHandler(&capturedCtx, &executed))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK in test environment, got %d: %s", rr.Code, rr.Body.String())
	}
	if !executed {
		t.Fatal("Expected downstream handler to be executed in test environment")
	}
	if capturedCtx.Role != "SUPER_ADMIN" || capturedCtx.OrgID != 1 {
		t.Errorf("Unexpected user context in test environment: %+v", capturedCtx)
	}
}

func TestAuthMiddleware_StagingEnvironment_RejectsTestTokens(t *testing.T) {
	guard := middleware.NewAuthMiddleware("", "", nil, middleware.WithEnvironment("staging"))

	attackTokens := []string{
		"test-token",
		"test-token-org1",
		"test-token-org2",
		"test-token-org999",
	}

	for _, token := range attackTokens {
		t.Run("Token_"+token, func(t *testing.T) {
			var capturedCtx middleware.UserContext
			executed := false
			handler := guard.RequireAuth(mockProtectedHandler(&capturedCtx, &executed))

			req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("X-Test-Org-ID", "2")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("Staging environment: expected 401 Unauthorized for %s, got %d", token, rr.Code)
			}
			if executed {
				t.Errorf("Staging environment: downstream handler executed for %s!", token)
			}
			if capturedCtx.Role != "" || capturedCtx.UserID != 0 {
				t.Errorf("Staging environment: user context fabricated for %s: %+v", token, capturedCtx)
			}
		})
	}
}

func TestAuthMiddleware_ProductionEnvironment_RejectsTestTokens(t *testing.T) {
	guard := middleware.NewAuthMiddleware("", "", nil, middleware.WithEnvironment("production"))

	attackTokens := []string{
		"test-token",
		"test-token-org1",
		"test-token-org2",
		"test-token-org999",
	}

	for _, token := range attackTokens {
		t.Run("Token_"+token, func(t *testing.T) {
			var capturedCtx middleware.UserContext
			executed := false
			handler := guard.RequireAuth(mockProtectedHandler(&capturedCtx, &executed))

			req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("X-Test-Org-ID", "1")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("Production environment: expected 401 Unauthorized for %s, got %d", token, rr.Code)
			}
			if executed {
				t.Errorf("Production environment: downstream handler executed for %s!", token)
			}
			if capturedCtx.Role != "" || capturedCtx.UserID != 0 {
				t.Errorf("Production environment: user context fabricated for %s: %+v", token, capturedCtx)
			}
		})
	}
}

func TestAuthMiddleware_MissingEnvironment_SafeDefault_RejectsTestTokens(t *testing.T) {
	// Calling NewAuthMiddleware without WithEnvironment defaults to "" (non-dev, non-test)
	guard := middleware.NewAuthMiddleware("", "", nil)

	if guard.IsDevOrTest() {
		t.Fatal("Safe default: AuthMiddleware without explicit env must NOT be dev/test")
	}

	var capturedCtx middleware.UserContext
	executed := false
	handler := guard.RequireAuth(mockProtectedHandler(&capturedCtx, &executed))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Missing environment: expected 401 Unauthorized, got %d", rr.Code)
	}
	if executed {
		t.Error("Missing environment: downstream handler should not have executed")
	}
}

func TestAuthMiddleware_UnknownEnvironment_RejectsTestTokens(t *testing.T) {
	unknownEnvs := []string{"sandbox", "preview", "qa", "local_docker", "ci_staging"}

	for _, env := range unknownEnvs {
		t.Run("Env_"+env, func(t *testing.T) {
			guard := middleware.NewAuthMiddleware("", "", nil, middleware.WithEnvironment(env))
			if guard.IsDevOrTest() {
				t.Fatalf("Unknown environment %s must not evaluate to dev/test", env)
			}

			var capturedCtx middleware.UserContext
			executed := false
			handler := guard.RequireAuth(mockProtectedHandler(&capturedCtx, &executed))

			req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
			req.Header.Set("Authorization", "Bearer test-token")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("Unknown env %s: expected 401 Unauthorized, got %d", env, rr.Code)
			}
			if executed {
				t.Errorf("Unknown env %s: downstream handler executed!", env)
			}
		})
	}
}

// ============================================================================
// SECTION 5: TENANT ISOLATION TESTS (Impersonation Prevention)
// ============================================================================

func TestAuthMiddleware_TenantIsolation_ImpersonationBlockedOutsideDev(t *testing.T) {
	nonDevEnvs := []string{"staging", "production", ""}

	impersonationAttempts := []struct {
		desc          string
		token         string
		testOrgHeader string
	}{
		{
			desc:          "Impersonate Org 1 via test-token-org1",
			token:         "test-token-org1",
			testOrgHeader: "",
		},
		{
			desc:          "Impersonate Org 2 via test-token-org2",
			token:         "test-token-org2",
			testOrgHeader: "",
		},
		{
			desc:          "Impersonate Org 999 via test-token-org999",
			token:         "test-token-org999",
			testOrgHeader: "",
		},
		{
			desc:          "Cross-tenant attempt via test-token + X-Test-Org-ID 888",
			token:         "test-token",
			testOrgHeader: "888",
		},
		{
			desc:          "Cross-tenant attempt via test-token-org1 + X-Test-Org-ID 2",
			token:         "test-token-org1",
			testOrgHeader: "2",
		},
	}

	for _, env := range nonDevEnvs {
		guard := middleware.NewAuthMiddleware("", "", nil, middleware.WithEnvironment(env))

		for _, att := range impersonationAttempts {
			t.Run("Env_"+env+"_"+att.desc, func(t *testing.T) {
				var capturedCtx middleware.UserContext
				executed := false
				handler := guard.RequireAuth(mockProtectedHandler(&capturedCtx, &executed))

				req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants/sensitive-data", nil)
				req.Header.Set("Authorization", "Bearer "+att.token)
				if att.testOrgHeader != "" {
					req.Header.Set("X-Test-Org-ID", att.testOrgHeader)
				}
				rr := httptest.NewRecorder()

				handler.ServeHTTP(rr, req)

				if rr.Code != http.StatusUnauthorized {
					t.Errorf("[%s] Expected 401 for tenant impersonation (%s), got %d", env, att.desc, rr.Code)
				}
				if executed {
					t.Errorf("[%s] Impersonation breached! Downstream handler executed for %s", env, att.desc)
				}
				if capturedCtx.OrgID != 0 {
					t.Errorf("[%s] Impersonation breached! Captured OrgID %d for %s", env, capturedCtx.OrgID, att.desc)
				}
			})
		}
	}
}

// ============================================================================
// SECTION 6: RBAC TESTS (Privilege Fabrication Prevention)
// ============================================================================

func TestAuthMiddleware_RBAC_PrivilegeFabricationBlockedOutsideDev(t *testing.T) {
	guard := middleware.NewAuthMiddleware("", "", nil, middleware.WithEnvironment("production"))

	var capturedCtx middleware.UserContext
	executed := false
	handler := guard.RequireAuth(mockProtectedHandler(&capturedCtx, &executed))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/grant-superadmin", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401 Unauthorized, got %d", rr.Code)
	}
	if executed {
		t.Fatal("Security critical failure: privileged endpoint executed under test-token in production!")
	}
	if capturedCtx.Role == "SUPER_ADMIN" {
		t.Fatal("Security critical failure: SUPER_ADMIN context was fabricated in production!")
	}
}

// ============================================================================
// SECTION 7: SECURITY REGRESSION TESTS (Explicit 7 Attack Vectors)
// ============================================================================

func TestAuthMiddleware_SecurityRegression_AllSevenAttackPatterns(t *testing.T) {
	guard := middleware.NewAuthMiddleware("", "", nil, middleware.WithEnvironment("production"))

	patterns := []struct {
		vectorID string
		name     string
		token    string
		header   string
		path     string
	}{
		{
			vectorID: "Vector 1",
			name:     "Bearer test-token",
			token:    "test-token",
			path:     "/api/v1/shipments",
		},
		{
			vectorID: "Vector 2",
			name:     "Bearer test-token-org2",
			token:    "test-token-org2",
			path:     "/api/v1/shipments",
		},
		{
			vectorID: "Vector 3",
			name:     "Bearer test-token-org999",
			token:    "test-token-org999",
			path:     "/api/v1/invoices",
		},
		{
			vectorID: "Vector 4",
			name:     "X-Test-Org-ID manipulation",
			token:    "test-token",
			header:   "99",
			path:     "/api/v1/rfqs",
		},
		{
			vectorID: "Vector 5",
			name:     "test token + arbitrary organization",
			token:    "test-token-org555",
			header:   "555",
			path:     "/api/v1/governance",
		},
		{
			vectorID: "Vector 6",
			name:     "test token + privileged endpoint",
			token:    "test-token",
			path:     "/api/v1/enterprise/system-override",
		},
		{
			vectorID: "Vector 7",
			name:     "test token + cross-tenant endpoint",
			token:    "test-token-org2",
			path:     "/api/v1/tenants/1/audit-logs",
		},
	}

	for _, p := range patterns {
		t.Run(p.vectorID+": "+p.name, func(t *testing.T) {
			var capturedCtx middleware.UserContext
			executed := false
			handler := guard.RequireAuth(mockProtectedHandler(&capturedCtx, &executed))

			req := httptest.NewRequest(http.MethodPost, p.path, nil)
			req.Header.Set("Authorization", "Bearer "+p.token)
			if p.header != "" {
				req.Header.Set("X-Test-Org-ID", p.header)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("[%s] Expected 401 Unauthorized, got %d", p.vectorID, rr.Code)
			}
			if executed {
				t.Errorf("[%s] Attack vector succeeded in reaching handler!", p.vectorID)
			}
			if capturedCtx.Role != "" || capturedCtx.UserID != 0 || capturedCtx.OrgID != 0 {
				t.Errorf("[%s] User context fabricated! %+v", p.vectorID, capturedCtx)
			}
		})
	}
}

// ============================================================================
// SECTION 9: ERROR SANITIZATION TESTS
// ============================================================================

func TestAuthMiddleware_ErrorSanitization_NoInfoLeakage(t *testing.T) {
	guard := middleware.NewAuthMiddleware("", "", nil, middleware.WithEnvironment("production"))

	var capturedCtx middleware.UserContext
	executed := false
	handler := guard.RequireAuth(mockProtectedHandler(&capturedCtx, &executed))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-Test-Org-ID", "99")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401 Unauthorized, got %d", rr.Code)
	}

	body := strings.TrimSpace(rr.Body.String())
	// Expected standard generic message
	if body != "Invalid token" {
		t.Errorf("Expected generic error 'Invalid token', got %q", body)
	}

	// Verify no internal variables or system details leaked
	forbiddenTerms := []string{
		"production", "development", "staging", "test",
		"SUPER_ADMIN", "UserID", "OrgID", "bypass", "mock-cognito-id",
	}
	for _, term := range forbiddenTerms {
		if strings.Contains(body, term) {
			t.Errorf("Security violation: error body leaks internal term %q: %s", term, body)
		}
	}
}
