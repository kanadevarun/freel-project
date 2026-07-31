package middleware

import (
	"net/http"

	"github.com/freel/backend/internal/rbac"
)

// RBACMiddleware handles permission checks for routes.
type RBACMiddleware struct {
	rbacSvc rbac.Service
}

// NewRBACMiddleware creates a new permission checking middleware.
// Simple meaning: It prepares a security guard that checks if you have the right clearance level.
// Example: rbacGuard := NewRBACMiddleware(rbacSvc)
func NewRBACMiddleware(rbacSvc rbac.Service) *RBACMiddleware {
	return &RBACMiddleware{rbacSvc: rbacSvc}
}

// RequirePermission wraps a route to ensure the user has specific permissions.
// Simple meaning: It stops anyone from accessing the page unless they have the required permission (like "Delete Lead").
// Example: router.With(rbacGuard.RequirePermission("LEADS", "DELETE")).Delete("/{id}", handler)
func (m *RBACMiddleware) RequirePermission(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract UserContext from the request
			userCtxVal := r.Context().Value(UserContextKey)
			if userCtxVal == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userCtx, ok := userCtxVal.(UserContext)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// If the user is a super admin, let them through immediately
			if userCtx.Role == "ADMIN" {
				next.ServeHTTP(w, r)
				return
			}

			// Check DB for the permission
			hasPerm, err := m.rbacSvc.HasPermission(r.Context(), userCtx.Role, userCtx.OrgID, resource, action)
			if err != nil {
				http.Error(w, "Error verifying permissions", http.StatusInternalServerError)
				return
			}

			if !hasPerm {
				http.Error(w, "Forbidden: You don't have permission to perform this action", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
