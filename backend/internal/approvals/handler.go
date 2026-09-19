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

	actorName := h.svc.ResolveUserName(r.Context(), userCtx.UserID)

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

	actorName := h.svc.ResolveUserName(r.Context(), userCtx.UserID)

	updated, err := h.svc.ApproveRequest(r.Context(), userCtx.OrgID, id, actorName, userCtx.UserID, payload.Notes)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "forbidden") || strings.Contains(strings.ToLower(err.Error()), "not an active member") {
			status = http.StatusForbidden
		} else if strings.Contains(strings.ToLower(err.Error()), "not found") {
			status = http.StatusNotFound
		} else if strings.Contains(strings.ToLower(err.Error()), "conflict") {
			status = http.StatusConflict
		}
		utils.Error(w, status, "Failed to approve request: "+err.Error(), "ACTION_FAILED")
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

	actorName := h.svc.ResolveUserName(r.Context(), userCtx.UserID)

	updated, err := h.svc.RejectRequest(r.Context(), userCtx.OrgID, id, actorName, userCtx.UserID, payload.Reason, payload.Notes)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "forbidden") || strings.Contains(strings.ToLower(err.Error()), "not an active member") {
			status = http.StatusForbidden
		} else if strings.Contains(strings.ToLower(err.Error()), "not found") {
			status = http.StatusNotFound
		}
		utils.Error(w, status, "Failed to reject request: "+err.Error(), "ACTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Approval request rejected successfully", updated)
}

// CancelRequest handles POST /api/v1/approvals/{id}/cancel
func (h *Handler) CancelRequest(w http.ResponseWriter, r *http.Request) {
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

	actorName := h.svc.ResolveUserName(r.Context(), userCtx.UserID)

	updated, err := h.svc.CancelRequest(r.Context(), userCtx.OrgID, id, actorName, userCtx.UserID, payload.Notes)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			status = http.StatusNotFound
		}
		utils.Error(w, status, "Failed to cancel request: "+err.Error(), "CANCEL_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Approval request cancelled successfully", updated)
}

// ReturnRequest handles POST /api/v1/approvals/{id}/return
func (h *Handler) ReturnRequest(w http.ResponseWriter, r *http.Request) {
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

	actorName := h.svc.ResolveUserName(r.Context(), userCtx.UserID)

	updated, err := h.svc.ReturnRequest(r.Context(), userCtx.OrgID, id, actorName, userCtx.UserID, payload.Reason, payload.Notes)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "forbidden") || strings.Contains(strings.ToLower(err.Error()), "not an active member") {
			status = http.StatusForbidden
		} else if strings.Contains(strings.ToLower(err.Error()), "not found") {
			status = http.StatusNotFound
		}
		utils.Error(w, status, "Failed to return request for changes: "+err.Error(), "ACTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Approval request returned for changes", updated)
}

// GetActionPreview handles GET /api/v1/approvals/{id}/preview
func (h *Handler) GetActionPreview(w http.ResponseWriter, r *http.Request) {
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

	preview, err := h.svc.GetActionPreview(r.Context(), userCtx.OrgID, id)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			status = http.StatusNotFound
		}
		utils.Error(w, status, "Failed to generate action preview: "+err.Error(), "PREVIEW_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Action preview retrieved successfully", preview)
}

// GetDecisionHistory handles GET /api/v1/approvals/{id}/history
func (h *Handler) GetDecisionHistory(w http.ResponseWriter, r *http.Request) {
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

	history, err := h.svc.GetDecisionHistory(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch decision history: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Retrieved decision history", history)
}

// GetExecutionStatus handles GET /api/v1/approvals/{id}/execution-status
func (h *Handler) GetExecutionStatus(w http.ResponseWriter, r *http.Request) {
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

	execStatus, err := h.svc.GetExecutionStatus(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "Failed to fetch execution status: "+err.Error(), "NOT_FOUND")
		return
	}

	utils.Success(w, http.StatusOK, "Retrieved execution status", execStatus)
}

// RetryExecution handles POST /api/v1/approvals/{id}/retry-execution
func (h *Handler) RetryExecution(w http.ResponseWriter, r *http.Request) {
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

	actorName := h.svc.ResolveUserName(r.Context(), userCtx.UserID)

	updated, err := h.svc.RetryExecution(r.Context(), userCtx.OrgID, id, actorName, userCtx.UserID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Failed to retry execution: "+err.Error(), "RETRY_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Execution retry triggered successfully", updated)
}

// GetRelatedRecommendation handles GET /api/v1/approvals/{id}/recommendation
func (h *Handler) GetRelatedRecommendation(w http.ResponseWriter, r *http.Request) {
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

	rec, err := h.svc.GetRelatedRecommendation(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch related recommendation: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Retrieved related recommendation", rec)
}

// GetRelatedSourceRecord handles GET /api/v1/approvals/{id}/source-record
func (h *Handler) GetRelatedSourceRecord(w http.ResponseWriter, r *http.Request) {
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

	srcRecord, err := h.svc.GetRelatedSourceRecord(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch related source record: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Retrieved related source record", srcRecord)
}

// GetAuditHistory handles GET /api/v1/approvals/{id}/audit
func (h *Handler) GetAuditHistory(w http.ResponseWriter, r *http.Request) {
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

	audits, err := h.svc.GetAuditHistory(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch audit history: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Retrieved approval audit history", audits)
}

// GetApprovalRequirements handles GET /api/v1/approvals/requirements?action=...
func (h *Handler) GetApprovalRequirements(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	actionName := r.URL.Query().Get("action")
	if actionName == "" {
		utils.Error(w, http.StatusBadRequest, "action query parameter is required", "INVALID_PARAM")
		return
	}

	reqs, err := h.svc.GetApprovalRequirements(r.Context(), userCtx.OrgID, actionName)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "REQUIREMENTS_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Retrieved approval requirements", reqs)
}

// ProposeAIApproval handles POST /internal/approvals/propose (Internal AI Bridge Endpoint)
func (h *Handler) ProposeAIApproval(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized internal request", "UNAUTHORIZED")
		return
	}

	var input struct {
		OrgID int64 `json:"org_id"`
		ProposeAIApprovalInput
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error(), "INVALID_JSON")
		return
	}

	if input.OrgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "org_id is required", "MISSING_ORG_ID")
		return
	}

	approval, err := h.svc.ProposeAIApproval(r.Context(), input.OrgID, &input.ProposeAIApprovalInput)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to propose AI approval: "+err.Error(), "PROPOSE_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "AI approval proposed successfully", approval)
}


