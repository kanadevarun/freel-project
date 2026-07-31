package organization

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// Handler manages HTTP requests for the organization module.
type Handler struct {
	svc Service
}

// NewHandler creates a new handler.
// Simple meaning: It sets up the API endpoints so they can talk to the business logic (service).
// Example: handler := NewHandler(myOrgService)
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Create handles POST /organizations
// Simple meaning: This is the web endpoint that frontend calls when someone clicks "Create Workspace".
// Example URL: POST /organizations (Body: {"name": "Acme Corp"})
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	org, err := h.svc.CreateOrganization(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(org)
}

// Update handles PUT /organizations/{id}
// Simple meaning: This is the web endpoint that frontend calls to change the name of the workspace.
// Example URL: PUT /organizations/123 (Body: {"name": "Acme Global"})
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL query or path params (assuming router will inject it)
	// Fallback to query param `id` for now
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid organization ID", http.StatusBadRequest)
		return
	}
	
	var req UpdateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	org, err := h.svc.UpdateOrganization(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(org)
}

// Invite handles POST /organizations/{id}/invite
// Simple meaning: This is the web endpoint that frontend calls when you fill out the invite team member form.
// Example URL: POST /organizations/123/invite (Body: {"email": "bob@acme.com", "role": "SALES"})
func (h *Handler) Invite(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid organization ID", http.StatusBadRequest)
		return
	}

	var req InviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.svc.InviteUser(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "invite sent successfully"})
}

// Settings handles GET /organizations/{id}/settings
// Simple meaning: This endpoint gives the frontend all the configuration settings (like GST number) for a workspace.
// Example URL: GET /organizations/123/settings
func (h *Handler) Settings(w http.ResponseWriter, r *http.Request) {
	// Placeholder for returning specific settings (GST, logos, etc.)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "settings ok"})
}
