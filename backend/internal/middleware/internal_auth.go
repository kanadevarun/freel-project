package middleware

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/freel/backend/internal/utils"
)

// InternalServiceKeyHeader is the standard HTTP header used for machine-to-machine internal service authentication.
const InternalServiceKeyHeader = "X-LogisticsHQ-Service-Key"

// devLocalOnlyFallbackKey is a clearly documented local-only fallback constant when INTERNAL_SERVICE_TOKEN is not configured in development.
// This is strictly forbidden in staging, production, and non-dev/non-test environments.
const devLocalOnlyFallbackKey = "dev-local-only-insecure-service-token-not-for-prod"

// isDevelopmentOrTestEnv checks if the current runtime environment is explicitly "development" or "test".
func isDevelopmentOrTestEnv() bool {
	rawEnv := os.Getenv("APP_ENV")
	if rawEnv == "" {
		rawEnv = os.Getenv("ENV")
	}
	if rawEnv == "" {
		rawEnv = os.Getenv("ENVIRONMENT")
	}
	normEnv := strings.ToLower(strings.TrimSpace(rawEnv))
	return normEnv == "development" || normEnv == "test"
}

// getExpectedInternalServiceToken resolves the expected service token.
// Outside explicit development or test environments, it strictly requires configuration (fail closed).
func getExpectedInternalServiceToken() (string, error) {
	expectedToken := os.Getenv("INTERNAL_SERVICE_TOKEN")
	if expectedToken == "" {
		expectedToken = os.Getenv("INTERNAL_SERVICE_KEY")
	}
	if expectedToken == "" {
		expectedToken = os.Getenv("AI_SIDECAR_SERVICE_KEY")
	}

	if expectedToken == "" {
		if !isDevelopmentOrTestEnv() {
			return "", errors.New("Configuration error: INTERNAL_SERVICE_TOKEN must be specified in non-development environments")
		}
		expectedToken = devLocalOnlyFallbackKey
	}
	return expectedToken, nil
}

// ValidateInternalServiceToken performs constant-time comparison of the internal service key header against configuration.
func ValidateInternalServiceToken(r *http.Request) error {
	token := r.Header.Get(InternalServiceKeyHeader)
	expectedToken, err := getExpectedInternalServiceToken()
	if err != nil {
		return err
	}

	if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
		return errors.New("Unauthorized access: Invalid service key token")
	}
	return nil
}

// InternalServiceAuthMiddleware authenticates internal machine-to-machine requests using service keys
func InternalServiceAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := getExpectedInternalServiceToken(); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Internal service configuration error", "CONFIG_ERROR")
			return
		}

		if err := ValidateInternalServiceToken(r); err != nil {
			utils.Error(w, http.StatusUnauthorized, "Unauthorized access: Invalid service key token", "UNAUTHORIZED")
			return
		}
		next.ServeHTTP(w, r)
	})
}

