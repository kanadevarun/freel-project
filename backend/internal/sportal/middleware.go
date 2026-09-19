package sportal

import (
	"fmt"
	"net/http"

	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
)

// RequireInternalStaff enforces that only authenticated internal LogisticsHQ users can access SPortal endpoints.
// Customer organization users (e.g. OrgID != 1, or customer roles) are rejected with HTTP 403 Forbidden.
func RequireInternalStaff(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userCtx, ok := middleware.GetUserContext(r.Context())
		if !ok || userCtx.UserID <= 0 {
			utils.Error(w, http.StatusUnauthorized, "Authentication required to access SPortal", "UNAUTHORIZED")
			return
		}

		// Enforce internal organization and internal role boundary
		if !IsInternalOrganization(userCtx.OrgID) || !IsInternalStaffRole(userCtx.Role) {
			actorID := userCtx.UserID
			_, _ = audit.Record(r.Context(), domain.CreateAuditLogParams{
				OrgID:        userCtx.OrgID,
				ActorID:      &actorID,
				ActorType:    domain.ActorTypeUser,
				ActorRole:    userCtx.Role,
				Action:       "sportal.forbidden_access_attempt",
				Module:       domain.ModuleAuthentication,
				ResourceType: "SPORTAL_API",
				ResourceID:   r.URL.Path,
				Description:  fmt.Sprintf("Customer tenant user (User %d, Org %d, Role %s) attempted to access SPortal endpoint %s", userCtx.UserID, userCtx.OrgID, userCtx.Role, r.URL.Path),
				Result:       domain.ResultFailed,
				ErrorMessage: "Customer tenant attempted access to internal SPortal API",
				IPAddress:    r.RemoteAddr,
				UserAgent:    r.UserAgent(),
			})

			utils.Error(w, http.StatusForbidden, "Forbidden: Only authorized LogisticsHQ internal personnel may access SPortal administration", "FORBIDDEN")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequirePermission wraps a handler to check that an authenticated internal user has a specific granular SPortal permission.
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx, ok := middleware.GetUserContext(r.Context())
			if !ok || userCtx.UserID <= 0 {
				utils.Error(w, http.StatusUnauthorized, "Authentication required to access SPortal", "UNAUTHORIZED")
				return
			}

			if !IsInternalOrganization(userCtx.OrgID) || !IsInternalStaffRole(userCtx.Role) {
				utils.Error(w, http.StatusForbidden, "Forbidden: Only authorized LogisticsHQ internal personnel may access SPortal", "FORBIDDEN")
				return
			}

			if !RoleHasPermission(userCtx.Role, permission) {
				actorID := userCtx.UserID
				_, _ = audit.Record(r.Context(), domain.CreateAuditLogParams{
					OrgID:        userCtx.OrgID,
					ActorID:      &actorID,
					ActorType:    domain.ActorTypeUser,
					ActorRole:    userCtx.Role,
					Action:       "sportal.permission_denied",
					Module:       domain.ModuleRolesPermissions,
					ResourceType: "SPORTAL_PERMISSION",
					ResourceID:   permission,
					Description:  fmt.Sprintf("User %d (Role %s) denied access to %s; missing permission %s", userCtx.UserID, userCtx.Role, r.URL.Path, permission),
					Result:       domain.ResultFailed,
					ErrorMessage: fmt.Sprintf("Missing required permission: %s", permission),
				})

				utils.Error(w, http.StatusForbidden, fmt.Sprintf("Forbidden: Role '%s' lacks required permission '%s'", userCtx.Role, permission), "INSUFFICIENT_PERMISSIONS")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
