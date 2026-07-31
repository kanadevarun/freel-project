package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type contextKey string

const UserContextKey = contextKey("user_context")

// UserContext holds the authenticated user's details.
type UserContext struct {
	UserID    int64
	OrgID     int64
	Role      string
	CognitoID string
}

// AuthMiddleware manages the JWT authentication.
type AuthMiddleware struct {
	jwksCache *jwk.Cache
	jwksURL   string
}

// NewAuthMiddleware sets up the authentication checker.
// Simple meaning: It prepares the security guard that will check ID badges (tokens) at the door.
// Example: authGuard := NewAuthMiddleware("us-east-1", "pool-123")
func NewAuthMiddleware(region, userPoolID string) *AuthMiddleware {
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", region, userPoolID)
	
	cache := jwk.NewCache(context.Background())
	cache.Register(jwksURL, jwk.WithMinRefreshInterval(15*time.Minute))
	
	// Pre-fetch the keys
	_, _ = cache.Refresh(context.Background(), jwksURL)

	return &AuthMiddleware{
		jwksCache: cache,
		jwksURL:   jwksURL,
	}
}

// RequireAuth is the actual security guard function applied to routes.
// Simple meaning: It stops anyone without a valid login token from accessing protected pages.
// Example: router.Use(authGuard.RequireAuth)
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing or invalid authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Fetch the public keys from AWS Cognito
		keyset, err := m.jwksCache.Get(r.Context(), m.jwksURL)
		if err != nil {
			http.Error(w, "Failed to fetch validation keys", http.StatusInternalServerError)
			return
		}

		// Verify the token
		token, err := jwt.Parse([]byte(tokenString), jwt.WithKeySet(keyset), jwt.WithValidate(true))
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Extract cognito subject ID
		cognitoSub := token.Subject()

		// TODO: Here we would look up the actual User in the PostgreSQL `users` table using cognitoSub
		// and find their active `org_members` record to populate UserID, OrgID, and Role.
		// For the sake of the scaffold, using mock data:
		userCtx := UserContext{
			CognitoID: cognitoSub,
			UserID:    1, 
			OrgID:     1, 
			Role:      "ADMIN", 
		}

		// Attach the user context to the request so handlers can use it
		ctx := context.WithValue(r.Context(), UserContextKey, userCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
