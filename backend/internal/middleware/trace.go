package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const CorrelationIDKey contextKey = "correlation_id"


// TraceMiddleware generates or propagates a Correlation ID.
//
// Simple meaning:
//   Every request that comes into our server is assigned a unique Correlation ID.
//   If the frontend already passed one in the 'X-Correlation-ID' header, we use it.
//   Otherwise, we generate a new UUID.
//   We then store it in the Go request context so all logging and service layers can read it.
//
// Example:
//   r.Use(middleware.TraceMiddleware)
func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corrID := r.Header.Get("X-Correlation-ID")
		if corrID == "" {
			corrID = uuid.NewString()
		}

		// Also set it in the response header so frontend can see it.
		w.Header().Set("X-Correlation-ID", corrID)

		ctx := context.WithValue(r.Context(), CorrelationIDKey, corrID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetCorrelationID extracts the correlation ID from the context.
//
// Simple meaning:
//   Reads the stored correlation ID from the context. Returns empty string if not found.
func GetCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return v
	}
	return ""
}
