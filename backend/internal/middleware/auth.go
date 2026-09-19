package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
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

// GetUserContext safely extracts UserContext from context whether passed as value or pointer.
func GetUserContext(ctx context.Context) (UserContext, bool) {
	if val, ok := ctx.Value(UserContextKey).(UserContext); ok {
		return val, true
	}
	if ptr, ok := ctx.Value(UserContextKey).(*UserContext); ok && ptr != nil {
		return *ptr, true
	}
	return UserContext{}, false
}

// AuthMiddleware manages the JWT authentication.
type AuthMiddleware struct {
	jwksCache   *jwk.Cache
	jwksURL     string
	db          *sqlx.DB
	environment string
}

// Option defines a functional configuration option for AuthMiddleware.
type Option func(*AuthMiddleware)

// WithEnvironment configures the operating environment (e.g. "development", "test", "staging", "production").
func WithEnvironment(env string) Option {
	return func(m *AuthMiddleware) {
		m.environment = strings.ToLower(strings.TrimSpace(env))
	}
}

// WithJWKSURL overrides the default Cognito JWKS URL (e.g. for mock testing).
func WithJWKSURL(url string) Option {
	return func(m *AuthMiddleware) {
		m.jwksURL = url
		if m.jwksCache != nil && url != "" {
			m.jwksCache.Register(url, jwk.WithMinRefreshInterval(15*time.Minute))
			_, _ = m.jwksCache.Refresh(context.Background(), url)
		}
	}
}

// NewAuthMiddleware sets up the authentication checker.
// Simple meaning: It prepares the security guard that will check ID badges (tokens) at the door.
// Example: authGuard := NewAuthMiddleware("us-east-1", "pool-123", db, WithEnvironment("development"))
func NewAuthMiddleware(region, userPoolID string, db *sqlx.DB, opts ...Option) *AuthMiddleware {
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", region, userPoolID)
	
	cache := jwk.NewCache(context.Background())
	if region != "" && userPoolID != "" {
		cache.Register(jwksURL, jwk.WithMinRefreshInterval(15*time.Minute))
		// Pre-fetch the keys
		_, _ = cache.Refresh(context.Background(), jwksURL)
	}

	m := &AuthMiddleware{
		jwksCache:   cache,
		jwksURL:     jwksURL,
		db:          db,
		environment: "", // Safe default: non-development / non-test (fail-closed)
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// SetEnvironment updates or overrides the environment setting on AuthMiddleware.
func (m *AuthMiddleware) SetEnvironment(env string) {
	m.environment = strings.ToLower(strings.TrimSpace(env))
}

// GetEnvironment returns the current environment setting.
func (m *AuthMiddleware) GetEnvironment() string {
	return m.environment
}

// IsDevOrTest returns true only if the environment is explicitly "development" or "test".
// Any other value (e.g. "staging", "production", "sandbox", or empty "") returns false.
func (m *AuthMiddleware) IsDevOrTest() bool {
	return m.environment == "development" || m.environment == "test"
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

		// Support test-token bypass strictly during explicitly enabled development/test environments
		if tokenString == "test-token" || tokenString == "test-token-org2" || strings.HasPrefix(tokenString, "test-token-org") {
			if !m.IsDevOrTest() {
				// Outside development/test, reject test tokens as invalid credentials immediately.
				// Do not reveal internal security implementation or environment details.
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			orgID := int64(1)
			if tokenString == "test-token-org2" || r.Header.Get("X-Test-Org-ID") == "2" {
				orgID = 2
			} else if strings.HasPrefix(tokenString, "test-token-org") {
				if parsed, err := strconv.ParseInt(strings.TrimPrefix(tokenString, "test-token-org"), 10, 64); err == nil && parsed > 0 {
					orgID = parsed
				}
			} else if testOrgHeader := r.Header.Get("X-Test-Org-ID"); testOrgHeader != "" {
				if parsed, err := strconv.ParseInt(testOrgHeader, 10, 64); err == nil && parsed > 0 {
					orgID = parsed
				}
			}
			role := "SUPER_ADMIN"
			if testRole := r.Header.Get("X-Test-Role"); testRole != "" {
				role = testRole
			}
			userCtx := UserContext{
				CognitoID: "mock-cognito-id",
				UserID:    1,
				OrgID:     orgID,
				Role:      role,
			}
			ctx := context.WithValue(r.Context(), UserContextKey, userCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Fetch the public keys from AWS Cognito
		keyset, err := m.jwksCache.Get(r.Context(), m.jwksURL)
		if err != nil {
			http.Error(w, "Failed to fetch validation keys", http.StatusInternalServerError)
			return
		}

		// Verify the token (allowing up to 5 minutes of clock skew for local/AWS NTP drift)
		token, err := jwt.Parse([]byte(tokenString), jwt.WithKeySet(keyset), jwt.WithValidate(true), jwt.WithAcceptableSkew(5*time.Minute))
		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Extract cognito subject ID
		cognitoSub := token.Subject()

		// Look up the actual User and their Organization details in Postgres
		var dbUser struct {
			UserID int64  `db:"user_id"`
			OrgID  int64  `db:"org_id"`
			Role   string `db:"role"`
		}
		query := `
			SELECT u.id AS user_id, om.org_id, r.name AS role
			FROM users u
			JOIN org_members om ON u.id = om.user_id
			JOIN roles r ON om.role_id = r.id
			WHERE u.cognito_sub = ? AND om.status = 'ACTIVE'
			LIMIT 1
		`
		err = m.db.GetContext(r.Context(), &dbUser, query, cognitoSub)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "User context not found or inactive in database", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Failed to resolve organization context: "+err.Error(), http.StatusInternalServerError)
			return
		}

		userCtx := UserContext{
			CognitoID: cognitoSub,
			UserID:    dbUser.UserID,
			OrgID:     dbUser.OrgID,
			Role:      dbUser.Role,
		}

		// Attach the user context to the request so handlers can use it
		ctx := context.WithValue(r.Context(), UserContextKey, userCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
