package approvals

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// ListApprovals handles GET /api/v1/approvals
func (h *Handler) ListApprovals(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	list, err := h.svc.ListApprovals(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve approvals: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	// Query params filtering
	categoryFilter := strings.ToUpper(r.URL.Query().Get("category"))
	statusFilter := r.URL.Query().Get("status")
	typeFilter := r.URL.Query().Get("type")
	searchQuery := strings.ToLower(r.URL.Query().Get("search"))

	if categoryFilter != "" && categoryFilter != "ALL" || statusFilter != "" && statusFilter != "ALL" || typeFilter != "" && typeFilter != "ALL" || searchQuery != "" {
		filtered := make([]*ApprovalRequest, 0)
		for _, item := range list {
			if categoryFilter != "" && categoryFilter != "ALL" && strings.ToUpper(item.Category) != categoryFilter {
				continue
			}
			if statusFilter != "" && statusFilter != "ALL" && !strings.EqualFold(item.Status, statusFilter) {
				continue
			}
			if typeFilter != "" && typeFilter != "ALL" && !strings.EqualFold(item.Type, typeFilter) {
				continue
			}
			if searchQuery != "" {
				title := strings.ToLower(item.Title)
				code := strings.ToLower(item.RequestCode)
				ref := ""
				if item.RelatedRef != nil {
					ref = strings.ToLower(*item.RelatedRef)
				}
				cust := ""
				if item.CustomerName != nil {
					cust = strings.ToLower(*item.CustomerName)
				}

				if !strings.Contains(title, searchQuery) && !strings.Contains(code, searchQuery) && !strings.Contains(ref, searchQuery) && !strings.Contains(cust, searchQuery) {
					continue
				}
			}
			filtered = append(filtered, item)
		}
		list = filtered
	}

	utils.Success(w, http.StatusOK, "Retrieved approval requests successfully", list)
}

// GetApprovalStats handles GET /api/v1/approvals/stats
func (h *Handler) GetApprovalStats(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	stats, err := h.svc.GetStats(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve approval stats: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Retrieved approval stats successfully", stats)
}

// GetApprovalByID handles GET /api/v1/approvals/{id}
func (h *Handler) GetApprovalByID(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid approval id parameter", "INVALID_PARAM")
		return
	}

	req, err := h.svc.GetApprovalByID(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}

	utils.Success(w, http.StatusOK, "Retrieved approval request details", req)
}

// CreateApproval handles POST /api/v1/approvals
func (h *Handler) CreateApproval(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	var input CreateApprovalInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error(), "INVALID_JSON")
		return
	}

	actorName := "Varun Kanade"
	if userCtx.UserID > 0 {
		actorName = "User #" + strconv.FormatInt(userCtx.UserID, 10)
	}

	newApproval, err := h.svc.CreateApproval(r.Context(), userCtx.OrgID, &input, actorName)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "CREATE_FAILED")
		return
	}

	utils.Success(w, http.StatusCreated, "Approval request created successfully", newApproval)
}

// ApproveRequest handles POST /api/v1/approvals/{id}/approve
func (h *Handler) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid approval id parameter", "INVALID_PARAM")
		return
	}

	var payload struct {
		Notes string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	actorName := "Varun Kanade"

	updated, err := h.svc.ApproveRequest(r.Context(), userCtx.OrgID, id, actorName, payload.Notes)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Failed to approve request: "+err.Error(), "ACTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Approval request approved successfully", updated)
}

// RejectRequest handles POST /api/v1/approvals/{id}/reject
func (h *Handler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid approval id parameter", "INVALID_PARAM")
		return
	}

	var payload struct {
		Reason string `json:"reason"`
		Notes  string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	actorName := "Varun Kanade"

	updated, err := h.svc.RejectRequest(r.Context(), userCtx.OrgID, id, actorName, payload.Reason, payload.Notes)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Failed to reject request: "+err.Error(), "ACTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Approval request rejected successfully", updated)
}
