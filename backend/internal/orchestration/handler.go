package orchestration

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

// GenerateProposal handles POST /api/v1/orchestration/proposals
func (h *Handler) GenerateProposal(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	var req GenerateProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_INPUT")
		return
	}

	proposal, err := h.svc.GenerateProposal(r.Context(), userCtx.OrgID, userCtx.UserID, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "PROPOSAL_GENERATION_FAILED")
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"success":  true,
		"proposal": proposal,
	})
}

// ListProposals handles GET /api/v1/orchestration/proposals
func (h *Handler) ListProposals(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := ProposalFilter{
		Limit:  limit,
		Offset: offset,
	}
	if s := q.Get("status"); s != "" {
		filter.Status = &s
	}
	if m := q.Get("source_module"); m != "" {
		filter.SourceModule = &m
	}
	if rk := q.Get("risk_level"); rk != "" {
		filter.RiskLevel = &rk
	}
	if sc := q.Get("search"); sc != "" {
		filter.Search = &sc
	}

	proposals, total, err := h.svc.ListProposals(r.Context(), userCtx.OrgID, filter)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "DB_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"proposals": proposals,
		"total":     total,
		"limit":     filter.Limit,
		"offset":    filter.Offset,
	})
}

// GetProposal handles GET /api/v1/orchestration/proposals/{id}
func (h *Handler) GetProposal(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idParam := chi.URLParam(r, "id")
	if id, err := strconv.ParseInt(idParam, 10, 64); err == nil && id > 0 {
		proposal, err := h.svc.GetProposal(r.Context(), userCtx.OrgID, id)
		if err != nil {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.JSON(w, http.StatusOK, map[string]interface{}{
			"success":  true,
			"proposal": proposal,
		})
		return
	}

	// Lookup by string proposal_id
	proposal, err := h.svc.GetProposalByProposalID(r.Context(), userCtx.OrgID, idParam)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"proposal": proposal,
	})
}

// ExecuteProposal handles POST /api/v1/orchestration/proposals/{id}/execute
func (h *Handler) ExecuteProposal(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idParam := chi.URLParam(r, "id")
	var req ExecuteProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_INPUT")
		return
	}

	if strings.TrimSpace(req.IdempotencyKey) == "" {
		utils.Error(w, http.StatusBadRequest, "idempotency_key is required", "MISSING_IDEMPOTENCY_KEY")
		return
	}

	exec, err := h.svc.ExecuteProposal(r.Context(), userCtx.OrgID, userCtx.UserID, idParam, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "EXECUTION_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"execution": exec,
	})
}

// ListExecutions handles GET /api/v1/orchestration/executions
func (h *Handler) ListExecutions(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := ExecutionFilter{
		Limit:  limit,
		Offset: offset,
	}
	if s := q.Get("status"); s != "" {
		filter.Status = &s
	}
	if a := q.Get("action_name"); a != "" {
		filter.ActionName = &a
	}
	if p := q.Get("proposal_id"); p != "" {
		filter.ProposalID = &p
	}

	executions, total, err := h.svc.ListExecutions(r.Context(), userCtx.OrgID, filter)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "DB_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"executions": executions,
		"total":      total,
		"limit":      filter.Limit,
		"offset":     filter.Offset,
	})
}

// GetExecution handles GET /api/v1/orchestration/executions/{id}
func (h *Handler) GetExecution(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid execution ID", "INVALID_INPUT")
		return
	}

	exec, err := h.svc.GetExecution(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"execution": exec,
	})
}

// CancelExecution handles POST /api/v1/orchestration/executions/{id}/cancel
func (h *Handler) CancelExecution(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid execution ID", "INVALID_INPUT")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Reason == "" {
		req.Reason = "Cancelled by operator"
	}

	exec, err := h.svc.CancelExecution(r.Context(), userCtx.OrgID, userCtx.UserID, id, req.Reason)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "CANCEL_FAILED")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"execution": exec,
	})
}

// RetryExecution handles POST /api/v1/orchestration/executions/{id}/retry
func (h *Handler) RetryExecution(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid execution ID", "INVALID_INPUT")
		return
	}

	exec, err := h.svc.RetryExecution(r.Context(), userCtx.OrgID, userCtx.UserID, id)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "RETRY_FAILED")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"execution": exec,
	})
}

// ListRegisteredActions handles GET /api/v1/orchestration/actions
func (h *Handler) ListRegisteredActions(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	actions := h.svc.ListRegisteredActions()
	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"actions": actions,
		"count":   len(actions),
	})
}
