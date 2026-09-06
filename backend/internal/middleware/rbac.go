package middleware

import (
	"context"
	"net/http"
)

// PermissionChecker defines the interface required by the RBAC middleware.
type PermissionChecker interface {
	HasPermission(ctx context.Context, roleName string, orgID int64, resource string, action string) (bool, error)
}

// RBACMiddleware handles permission checks for routes.
type RBACMiddleware struct {
	rbacSvc PermissionChecker
}

// NewRBACMiddleware creates a new permission checking middleware.
func NewRBACMiddleware(rbacSvc PermissionChecker) *RBACMiddleware {
	return &RBACMiddleware{rbacSvc: rbacSvc}
}

// RequirePermission wraps a route to ensure the user has specific permissions.
func (m *RBACMiddleware) RequirePermission(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract UserContext from the request
			userCtx, ok := GetUserContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// If the user is SUPER_ADMIN, they have all permissions — skip DB lookup
			if userCtx.Role == "SUPER_ADMIN" {
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
