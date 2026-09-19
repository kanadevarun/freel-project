package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/freel/backend/internal/middleware"
)

func TestInternalServiceAuthMiddleware_FailClosedInProductionAndStaging(t *testing.T) {
	origToken := os.Getenv("INTERNAL_SERVICE_TOKEN")
	origKey := os.Getenv("INTERNAL_SERVICE_KEY")
	origSidecarKey := os.Getenv("AI_SIDECAR_SERVICE_KEY")
	origAppEnv := os.Getenv("APP_ENV")
	defer func() {
		os.Setenv("INTERNAL_SERVICE_TOKEN", origToken)
		os.Setenv("INTERNAL_SERVICE_KEY", origKey)
		os.Setenv("AI_SIDECAR_SERVICE_KEY", origSidecarKey)
		os.Setenv("APP_ENV", origAppEnv)
	}()

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	t.Run("Production with missing token fails closed with 500 config error", func(t *testing.T) {
		os.Setenv("APP_ENV", "production")
		os.Unsetenv("INTERNAL_SERVICE_TOKEN")
		os.Unsetenv("INTERNAL_SERVICE_KEY")
		os.Unsetenv("AI_SIDECAR_SERVICE_KEY")

		req := httptest.NewRequest(http.MethodGet, "/internal/status", nil)
		req.Header.Set(middleware.InternalServiceKeyHeader, "any-token")
		rr := httptest.NewRecorder()

		middleware.InternalServiceAuthMiddleware(dummyHandler).ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("Expected 500 InternalServerError for unconfigured production service key, got %d", rr.Code)
		}

		body := rr.Body.String()
		if strings.Contains(body, "dev-local-only") {
			t.Fatal("Security violation: response body leaks dev-local-only fallback key")
		}
	})

	t.Run("Staging with missing token fails closed with 500 config error", func(t *testing.T) {
		os.Setenv("APP_ENV", "staging")
		os.Unsetenv("INTERNAL_SERVICE_TOKEN")
		os.Unsetenv("INTERNAL_SERVICE_KEY")
		os.Unsetenv("AI_SIDECAR_SERVICE_KEY")

		req := httptest.NewRequest(http.MethodGet, "/internal/status", nil)
		req.Header.Set(middleware.InternalServiceKeyHeader, "any-token")
		rr := httptest.NewRecorder()

		middleware.InternalServiceAuthMiddleware(dummyHandler).ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("Expected 500 for unconfigured staging service key, got %d", rr.Code)
		}
	})

	t.Run("Unknown environment with missing token fails closed", func(t *testing.T) {
		os.Setenv("APP_ENV", "qa-preview")
		os.Unsetenv("INTERNAL_SERVICE_TOKEN")

		req := httptest.NewRequest(http.MethodGet, "/internal/status", nil)
		rr := httptest.NewRecorder()

		middleware.InternalServiceAuthMiddleware(dummyHandler).ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("Expected 500 for unknown environment without token, got %d", rr.Code)
		}
	})

	t.Run("Missing environment with missing token fails closed", func(t *testing.T) {
		os.Unsetenv("APP_ENV")
		os.Unsetenv("ENV")
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("INTERNAL_SERVICE_TOKEN")

		req := httptest.NewRequest(http.MethodGet, "/internal/status", nil)
		rr := httptest.NewRecorder()

		middleware.InternalServiceAuthMiddleware(dummyHandler).ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("Expected 500 for missing environment without token, got %d", rr.Code)
		}
	})

	t.Run("Production with valid configured token succeeds", func(t *testing.T) {
		const validKey = "prod-secret-key-1234567890abcdef"
		os.Setenv("APP_ENV", "production")
		os.Setenv("INTERNAL_SERVICE_TOKEN", validKey)

		req := httptest.NewRequest(http.MethodGet, "/internal/status", nil)
		req.Header.Set(middleware.InternalServiceKeyHeader, validKey)
		rr := httptest.NewRecorder()

		middleware.InternalServiceAuthMiddleware(dummyHandler).ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK for valid configured service token, got %d", rr.Code)
		}
	})

	t.Run("Production with invalid token rejected with 401 Unauthorized", func(t *testing.T) {
		const validKey = "prod-secret-key-1234567890abcdef"
		os.Setenv("APP_ENV", "production")
		os.Setenv("INTERNAL_SERVICE_TOKEN", validKey)

		req := httptest.NewRequest(http.MethodGet, "/internal/status", nil)
		req.Header.Set(middleware.InternalServiceKeyHeader, "attacker-fake-token")
		rr := httptest.NewRecorder()

		middleware.InternalServiceAuthMiddleware(dummyHandler).ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("Expected 401 Unauthorized for invalid token, got %d", rr.Code)
		}
	})
}
