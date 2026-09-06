package rbac

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/middleware"
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

// getOrgIDFromContext extracts the authenticated Organization ID from the request context.
func getOrgIDFromContext(r *http.Request) (int64, error) {
	if userCtx, ok := middleware.GetUserContext(r.Context()); ok && userCtx.OrgID > 0 {
		return userCtx.OrgID, nil
	}
	return 0, fmt.Errorf("unauthorized or missing organization context")
}

// extractIDFromPath extracts a numeric ID from a URL path segment after a named prefix.
// e.g. extractIDFromPath("/api/v1/roles/123/permissions", "roles", "permissions") → 123
func extractIDFromPath(path, prefix, suffix string) (int64, error) {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == prefix && i+1 < len(parts) {
			idStr := parts[i+1]
			if suffix != "" && i+2 < len(parts) && parts[i+2] != suffix {
				continue
			}
			return strconv.ParseInt(idStr, 10, 64)
		}
	}
	return 0, strconv.ErrSyntax
}

// GetStats handles GET /api/v1/roles/stats — returns dynamic RBAC dashboard statistics.
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	orgID, err := getOrgIDFromContext(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	stats, err := h.service.GetStats(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch stats", "FETCH_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Stats retrieved successfully", stats)
}

// GetRoles handles GET /api/v1/roles — returns all roles with permission_count.
func (h *Handler) GetRoles(w http.ResponseWriter, r *http.Request) {
	orgID, err := getOrgIDFromContext(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	roles, err := h.service.GetRoles(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch roles", "FETCH_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Roles retrieved successfully", roles)
}

// CreateRole handles POST /api/v1/roles — creates a new custom role.
func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	orgID, err := getOrgIDFromContext(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON payload", "INVALID_PAYLOAD")
		return
	}

	role, err := h.service.CreateRole(r.Context(), orgID, req)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create role", "CREATE_FAILED")
		return
	}

	utils.Success(w, http.StatusCreated, "Role created successfully", role)
}

// GetRolePermissions handles GET /api/v1/roles/{id}/permissions.
func (h *Handler) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	orgID, err := getOrgIDFromContext(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

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

// UpdateRolePermissions handles PUT /api/v1/roles/{id}/permissions.
func (h *Handler) UpdateRolePermissions(w http.ResponseWriter, r *http.Request) {
	orgID, err := getOrgIDFromContext(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

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

// UpdateRole handles PUT /api/v1/roles/{id}
func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	orgID, err := getOrgIDFromContext(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	roleID, err := extractIDFromPath(r.URL.Path, "roles", "")
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid role ID", "INVALID_ID")
		return
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON payload", "INVALID_PAYLOAD")
		return
	}

	err = h.service.UpdateRole(r.Context(), orgID, roleID, req)
	if err != nil {
		if err.Error() == "role not found in organization" {
			utils.Error(w, http.StatusNotFound, "Role not found", "NOT_FOUND")
			return
		}
		if err.Error() == "role with this name already exists in the organization" || err.Error() == "cannot modify metadata of a system role" {
			utils.Error(w, http.StatusConflict, err.Error(), "CONFLICT")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to update role", "UPDATE_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Role updated successfully", nil)
}

// DeleteRole handles DELETE /api/v1/roles/{id}
func (h *Handler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	orgID, err := getOrgIDFromContext(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	roleID, err := extractIDFromPath(r.URL.Path, "roles", "")
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid role ID", "INVALID_ID")
		return
	}

	err = h.service.DeleteRole(r.Context(), orgID, roleID)
	if err != nil {
		if err.Error() == "role not found in organization" {
			utils.Error(w, http.StatusNotFound, "Role not found", "NOT_FOUND")
			return
		}
		if err.Error() == "role is assigned to users" || err.Error() == "cannot delete a system role" {
			utils.Error(w, http.StatusConflict, err.Error(), "CONFLICT")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to delete role", "DELETE_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Role deleted successfully", nil)
}
