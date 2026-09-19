package actions

import (
	"encoding/json"
	"net/http"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
)

// Handler exposes the Centralized Action System over internal HTTP routes.
type Handler struct {
	svc Service
}

// NewHandler creates a new actions HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// ExecuteAction handles POST /internal/actions/execute.
func (h *Handler) ExecuteAction(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized access: Invalid service key token", "UNAUTHORIZED")
		return
	}

	var req ActionExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}

	resp, err := h.svc.Execute(r.Context(), req)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if resp.Error != nil {
		switch resp.Error.Type {
		case "Validation":
			w.WriteHeader(http.StatusBadRequest)
		case "Unauthorized":
			w.WriteHeader(http.StatusForbidden)
		case "NotFound":
			w.WriteHeader(http.StatusNotFound)
		case "Conflict":
			w.WriteHeader(http.StatusConflict)
		case "ConfirmationRequired":
			w.WriteHeader(http.StatusOK) // Return 200 with structured confirmation requirement
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// ListActions handles GET /internal/actions.
func (h *Handler) ListActions(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized access: Invalid service key token", "UNAUTHORIZED")
		return
	}

	descriptors := h.svc.ListActions()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"actions": descriptors,
		"count":   len(descriptors),
	})
}
