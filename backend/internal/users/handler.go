package users

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
)

// Handler processes HTTP requests for the users module.
// It acts as the transport layer, converting HTTP requests into Service method calls
// and formatting the responses back to the client using standardized JSON structures.
type Handler struct {
	service Service
}

// NewHandler initializes and returns a new HTTP handler for the users module.
// It requires an instantiated users Service to perform the core business logic.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// InviteUser handles the POST /users/invite endpoint.
// It expects a JSON payload containing the target email and the role ID to assign.
// It extracts the orgID from the request context (populated by authentication middleware),
// invokes the Service to create the invite and send the email, and returns a success response.
func (h *Handler) InviteUser(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON request body into the InviteUserRequest struct.
	var req InviteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	// Validate the request fields explicitly (could also use a validator library).
	if req.Email == "" || req.RoleID == 0 {
		utils.Error(w, http.StatusBadRequest, "Email and Role ID are required", "VALIDATION_ERROR")
		return
	}

	// Extract Organization ID from context.
	orgID, err := getOrgIDFromContext(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	// Delegate the business logic to the Service layer.
	err = h.service.InviteUser(r.Context(), orgID, req)
	if err != nil {
		if strings.Contains(err.Error(), "INVITATION_ALREADY_EXISTS") {
			utils.Error(w, http.StatusConflict, "An invitation has already been sent to this email address.", "DUPLICATE_INVITATION")
			return
		}
		// Return a generic error to the client to avoid leaking internal state,
		// but in production this should be logged centrally.
		utils.Error(w, http.StatusInternalServerError, "Failed to send invitation", "INVITE_FAILED")
		return
	}

	// Respond with a standardized success envelope.
	utils.Success(w, http.StatusOK, "Invitation sent successfully", nil)
}

// ListUsers handles the GET /users endpoint.
// It retrieves the organization ID from the context and fetches all active members
// via the Service layer.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	orgID, err := getOrgIDFromContext(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	members, err := h.service.ListUsers(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve users", "FETCH_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Users retrieved successfully", members)
}

// ListInvitations handles the GET /users/invites endpoint.
// It retrieves the organization ID from the context and fetches all pending invitations.
func (h *Handler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	orgID, err := getOrgIDFromContext(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	invites, err := h.service.ListInvitations(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve invitations", "FETCH_FAILED")
		return
	}

	// If no invites, return an empty array instead of null
	if invites == nil {
		invites = []InvitationResponse{}
	}

	utils.Success(w, http.StatusOK, "Invitations retrieved successfully", invites)
}

// CancelInvitation handles the DELETE /users/invites/:id endpoint.
// It extracts the invitation ID from the URL path and cancels it.
func (h *Handler) CancelInvitation(w http.ResponseWriter, r *http.Request) {
	// Extract the target inviteID from the URL path
	inviteID, err := extractIDFromPath(r.URL.Path, "invites", "")
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid invitation ID in path", "INVALID_ID")
		return
	}

	orgID, err := getOrgIDFromContext(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	err = h.service.CancelInvitation(r.Context(), orgID, inviteID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to cancel invitation", "CANCEL_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Invitation canceled successfully", nil)
}

// UpdateRole handles the PATCH /users/:id/role endpoint.
// It parses the user ID from the URL path and the new role ID from the JSON body.
// It ensures the user updating the role has the authority (handled by middleware)
// and updates the database record.
func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	// Extract the target userID from the URL path (e.g. /users/123/role)
	userID, err := extractIDFromPath(r.URL.Path, "users", "role")
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid user ID in path", "INVALID_ID")
		return
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	if req.RoleID == 0 {
		utils.Error(w, http.StatusBadRequest, "Role ID is required", "VALIDATION_ERROR")
		return
	}

	orgID, err := getOrgIDFromContext(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	err = h.service.UpdateRole(r.Context(), orgID, userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "SUPER_ADMIN") {
			utils.Error(w, http.StatusBadRequest, err.Error(), "SUPER_ADMIN_RESTRICTION")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to update user role", "UPDATE_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Role updated successfully", nil)
}

// RemoveUser handles the DELETE /users/:id endpoint.
// It extracts the user ID from the URL path and revokes their access to the current organization.
func (h *Handler) RemoveUser(w http.ResponseWriter, r *http.Request) {
	// Extract the target userID from the URL path (e.g. /users/123)
	userID, err := extractIDFromPath(r.URL.Path, "users", "")
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid user ID in path", "INVALID_ID")
		return
	}

	orgID, err := getOrgIDFromContext(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	err = h.service.RemoveUser(r.Context(), orgID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "SUPER_ADMIN") {
			utils.Error(w, http.StatusBadRequest, err.Error(), "SUPER_ADMIN_RESTRICTION")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to remove user", "REMOVE_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "User removed successfully", nil)
}

// --- Helper Functions ---

// getOrgIDFromContext extracts the authenticated Organization ID from the request context.
func getOrgIDFromContext(r *http.Request) (int64, error) {
	if userCtx, ok := middleware.GetUserContext(r.Context()); ok && userCtx.OrgID > 0 {
		return userCtx.OrgID, nil
	}
	return 0, fmt.Errorf("unauthorized or missing organization context")
}

// extractIDFromPath is a simplistic URL parser to extract an integer ID from a path pattern.
// Examples:
// path: "/api/users/456", entity: "users", suffix: "" -> returns 456
// path: "/api/users/456/role", entity: "users", suffix: "role" -> returns 456
func extractIDFromPath(path, entity, suffix string) (int64, error) {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == entity && i+1 < len(parts) {
			idStr := parts[i+1]
			// Check if the next part matches the expected suffix (if provided)
			if suffix != "" && (i+2 >= len(parts) || parts[i+2] != suffix) {
				continue
			}
			return strconv.ParseInt(idStr, 10, 64)
		}
	}
	return 0, fmt.Errorf("id not found in path")
}
