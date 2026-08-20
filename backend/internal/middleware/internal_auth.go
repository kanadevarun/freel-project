package middleware

import (
	"net/http"
	"os"

	"github.com/freel/backend/internal/utils"
)

// InternalServiceAuthMiddleware authenticates internal machine-to-machine requests using service keys
func InternalServiceAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-LogisticsHQ-Service-Key")
		expectedToken := os.Getenv("INTERNAL_SERVICE_TOKEN")
		// 18. Consistently enforce environment secrets in production (remove hardcoded fallback secrets)
		isProd := os.Getenv("APP_ENV") == "production"
		if expectedToken == "" {
			if isProd {
				utils.Error(w, http.StatusInternalServerError, "Configuration error: INTERNAL_SERVICE_TOKEN must be specified in production environments", "CONFIG_ERROR")
				return
			}
			expectedToken = "internal-service-key-logisticshq"
		}

		if token != expectedToken || token == "" {
			utils.Error(w, http.StatusUnauthorized, "Unauthorized access: Invalid service key token", "UNAUTHORIZED")
			return
		}
		next.ServeHTTP(w, r)
	})
}
