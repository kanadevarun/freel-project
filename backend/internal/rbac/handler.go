package rbac

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/utils"
)

// Handler manages HTTP endpoints for RBAC roles and permissions.
type Handler struct {
	service Service
}

// NewHandler initializes a new RBAC HTTP handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// getOrgIDFromContext is a helper to extract the org_id from the request context.
// In a real app, this is typically set by the auth middleware.
func getOrgIDFromContext(r *http.Request) int64 {
	orgID, ok := r.Context().Value("org_id").(int64)
	if !ok {
		// For MVP fallback if context not set, assume org 1
		return 1
	}
	return orgID
}

// extractIDFromPath is a simple helper to get an ID from a URL pattern like /roles/:id/permissions
func extractIDFromPath(path, prefix, suffix string) (int64, error) {
	// e.g. path = "/api/v1/roles/123/permissions"
	// prefix = "roles", suffix = "permissions"
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == prefix && i+1 < len(parts) {
			idStr := parts[i+1]
			if suffix != "" && i+2 < len(parts) && parts[i+2] != suffix {
				continue // wrong route format
			}
			return strconv.ParseInt(idStr, 10, 64)
		}
	}
	return 0, strconv.ErrSyntax
}

// GetRoles handles the GET /api/v1/roles endpoint.
func (h *Handler) GetRoles(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgIDFromContext(r)
	
	roles, err := h.service.GetRoles(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch roles", "FETCH_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Roles retrieved successfully", roles)
}

// GetRolePermissions handles the GET /api/v1/roles/:id/permissions endpoint.
func (h *Handler) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgIDFromContext(r)
	roleID, err := extractIDFromPath(r.URL.Path, "roles", "permissions")
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid role ID", "INVALID_ID")
		return
	}

	perms, err := h.service.GetRolePermissions(r.Context(), orgID, roleID)
	if err != nil {
		if err.Error() == "role not found in organization" {
			utils.Error(w, http.StatusNotFound, "Role not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch permissions", "FETCH_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Permissions retrieved successfully", perms)
}

// UpdateRolePermissions handles the PUT /api/v1/roles/:id/permissions endpoint.
func (h *Handler) UpdateRolePermissions(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgIDFromContext(r)
	roleID, err := extractIDFromPath(r.URL.Path, "roles", "permissions")
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid role ID", "INVALID_ID")
		return
	}

	var req UpdatePermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON payload", "INVALID_PAYLOAD")
		return
	}

	err = h.service.UpdateRolePermissions(r.Context(), orgID, roleID, req)
	if err != nil {
		if err.Error() == "role not found in organization" {
			utils.Error(w, http.StatusNotFound, "Role not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to update permissions", "UPDATE_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Permissions updated successfully", nil)
}
