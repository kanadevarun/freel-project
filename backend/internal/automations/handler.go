package automations

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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

// ListAutomations handles GET /api/v1/automations
func (h *Handler) ListAutomations(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := AutomationFilter{
		Page:           page,
		Limit:          limit,
		AutomationType: q.Get("automation_type"),
		TriggerType:    q.Get("trigger_type"),
		Scope:          q.Get("scope"),
		OwnerTeam:      q.Get("owner_team"),
		Search:         q.Get("search"),
		SortBy:         q.Get("sort_by"),
		SortDir:        q.Get("sort_dir"),
	}

	if en := q.Get("is_enabled"); en != "" {
		val := en == "true" || en == "1"
		filter.IsEnabled = &val
	}

	automations, total, err := h.svc.ListAutomations(r.Context(), userCtx.OrgID, filter)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "AUTOMATION_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"automations": automations,
		"total":       total,
		"page":        filter.Page,
		"limit":       filter.Limit,
	})
}

// GetSupportedTypes handles GET /api/v1/automations/supported-types
func (h *Handler) GetSupportedTypes(w http.ResponseWriter, r *http.Request) {
	types := h.svc.GetSupportedTypes()
	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"supported_types": types,
	})
}

// GetStats handles GET /api/v1/automations/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	stats, err := h.svc.GetStats(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "STATS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"stats":   stats,
	})
}

// CreateAutomation handles POST /api/v1/automations
func (h *Handler) CreateAutomation(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	var input CreateAutomationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON payload", "INVALID_INPUT")
		return
	}

	created, err := h.svc.CreateAutomation(r.Context(), userCtx.OrgID, userCtx.UserID, input)
	if err != nil {
		if errors.Is(err, ErrInvalidAutomationType) || errors.Is(err, ErrInvalidSchedule) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "VALIDATION_FAILED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "CREATE_FAILED")
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"success":    true,
		"automation": created,
	})
}

// GetAutomation handles GET /api/v1/automations/{id}
func (h *Handler) GetAutomation(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid automation ID", "INVALID_ID")
		return
	}

	auto, err := h.svc.GetAutomation(r.Context(), userCtx.OrgID, id)
	if err != nil {
		if errors.Is(err, ErrAutomationNotFound) {
			utils.Error(w, http.StatusNotFound, "Automation not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "AUTOMATION_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"automation": auto,
	})
}

// UpdateAutomation handles PUT /api/v1/automations/{id}
func (h *Handler) UpdateAutomation(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid automation ID", "INVALID_ID")
		return
	}

	var input UpdateAutomationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON payload", "INVALID_INPUT")
		return
	}

	updated, err := h.svc.UpdateAutomation(r.Context(), userCtx.OrgID, id, userCtx.UserID, input)
	if err != nil {
		if errors.Is(err, ErrAutomationNotFound) {
			utils.Error(w, http.StatusNotFound, "Automation not found", "NOT_FOUND")
			return
		}
		if errors.Is(err, ErrInvalidSchedule) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "VALIDATION_FAILED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "UPDATE_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"automation": updated,
	})
}

// DeleteAutomation handles DELETE /api/v1/automations/{id}
func (h *Handler) DeleteAutomation(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid automation ID", "INVALID_ID")
		return
	}

	if err := h.svc.DeleteAutomation(r.Context(), userCtx.OrgID, id, userCtx.UserID); err != nil {
		if errors.Is(err, ErrAutomationNotFound) {
			utils.Error(w, http.StatusNotFound, "Automation not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "DELETE_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Automation deleted successfully",
	})
}

// EnableAutomation handles POST /api/v1/automations/{id}/enable
func (h *Handler) EnableAutomation(w http.ResponseWriter, r *http.Request) {
	h.toggleEnabled(w, r, true)
}

// DisableAutomation handles POST /api/v1/automations/{id}/disable
func (h *Handler) DisableAutomation(w http.ResponseWriter, r *http.Request) {
	h.toggleEnabled(w, r, false)
}

func (h *Handler) toggleEnabled(w http.ResponseWriter, r *http.Request, isEnabled bool) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid automation ID", "INVALID_ID")
		return
	}

	updated, err := h.svc.SetEnabled(r.Context(), userCtx.OrgID, id, isEnabled, userCtx.UserID)
	if err != nil {
		if errors.Is(err, ErrAutomationNotFound) {
			utils.Error(w, http.StatusNotFound, "Automation not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "STATUS_UPDATE_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"automation": updated,
	})
}

// PreviewNextRun handles POST /api/v1/automations/{id}/preview-next
func (h *Handler) PreviewNextRun(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid automation ID", "INVALID_ID")
		return
	}

	preview, err := h.svc.PreviewNextRun(r.Context(), userCtx.OrgID, id)
	if err != nil {
		if errors.Is(err, ErrAutomationNotFound) {
			utils.Error(w, http.StatusNotFound, "Automation not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "PREVIEW_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"preview": preview,
	})
}

// TriggerManualRun handles POST /api/v1/automations/{id}/run
func (h *Handler) TriggerManualRun(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid automation ID", "INVALID_ID")
		return
	}

	exec, err := h.svc.TriggerManualRun(r.Context(), userCtx.OrgID, id, userCtx.UserID)
	if err != nil {
		if errors.Is(err, ErrAutomationNotFound) {
			utils.Error(w, http.StatusNotFound, "Automation not found", "NOT_FOUND")
			return
		}
		if errors.Is(err, ErrAutomationDisabled) {
			utils.Error(w, http.StatusBadRequest, "Cannot trigger a disabled automation. Enable it first.", "AUTOMATION_DISABLED")
			return
		}
		if errors.Is(err, ErrExecutionAlreadyRunning) {
			utils.Error(w, http.StatusConflict, "An execution is already queued or running for this automation.", "RUN_IN_PROGRESS")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "RUN_FAILED")
		return
	}

	utils.JSON(w, http.StatusAccepted, map[string]interface{}{
		"success":   true,
		"message":   "Automation job triggered successfully",
		"execution": exec,
	})
}

// ListExecutions handles GET /api/v1/automations/executions and /api/v1/automations/{id}/executions
func (h *Handler) ListExecutions(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := ExecutionFilter{
		Page:        page,
		Limit:       limit,
		Status:      q.Get("status"),
		TriggerType: q.Get("trigger_type"),
		CurrentStep: q.Get("current_step"),
		SortBy:      q.Get("sort_by"),
		SortDir:     q.Get("sort_dir"),
	}

	if autoIDStr := chi.URLParam(r, "id"); autoIDStr != "" {
		if autoID, err := strconv.ParseInt(autoIDStr, 10, 64); err == nil && autoID > 0 {
			filter.AutomationID = &autoID
		}
	} else if autoIDStr := q.Get("automation_id"); autoIDStr != "" {
		if autoID, err := strconv.ParseInt(autoIDStr, 10, 64); err == nil && autoID > 0 {
			filter.AutomationID = &autoID
		}
	}

	executions, total, err := h.svc.ListExecutions(r.Context(), userCtx.OrgID, filter)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXECUTIONS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"executions": executions,
		"total":      total,
		"page":       filter.Page,
		"limit":      filter.Limit,
	})
}

// GetExecution handles GET /api/v1/automations/executions/{executionId}
func (h *Handler) GetExecution(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	execID, err := strconv.ParseInt(chi.URLParam(r, "executionId"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid execution ID", "INVALID_ID")
		return
	}

	exec, err := h.svc.GetExecution(r.Context(), userCtx.OrgID, execID)
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			utils.Error(w, http.StatusNotFound, "Execution not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXECUTION_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"execution": exec,
	})
}

// CancelExecution handles POST /api/v1/automations/executions/{executionId}/cancel
func (h *Handler) CancelExecution(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	execID, err := strconv.ParseInt(chi.URLParam(r, "executionId"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid execution ID", "INVALID_ID")
		return
	}

	var reqBody struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&reqBody)

	exec, err := h.svc.CancelExecution(r.Context(), userCtx.OrgID, execID, userCtx.UserID, reqBody.Reason)
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			utils.Error(w, http.StatusNotFound, "Execution not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "CANCEL_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"message":   "Execution cancelled successfully",
		"execution": exec,
	})
}

// RetryExecution handles POST /api/v1/automations/executions/{executionId}/retry
func (h *Handler) RetryExecution(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	execID, err := strconv.ParseInt(chi.URLParam(r, "executionId"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid execution ID", "INVALID_ID")
		return
	}

	exec, err := h.svc.RetryExecution(r.Context(), userCtx.OrgID, execID, userCtx.UserID)
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			utils.Error(w, http.StatusNotFound, "Execution not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "RETRY_FAILED")
		return
	}

	utils.JSON(w, http.StatusAccepted, map[string]interface{}{
		"success":   true,
		"message":   "Execution retry triggered successfully",
		"execution": exec,
	})
}

// GetExecutionRecommendations handles GET /api/v1/automations/executions/{executionId}/recommendations
func (h *Handler) GetExecutionRecommendations(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	execID, err := strconv.ParseInt(chi.URLParam(r, "executionId"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid execution ID", "INVALID_ID")
		return
	}

	recs, err := h.svc.GetExecutionRecommendations(r.Context(), userCtx.OrgID, execID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "RECOMMENDATIONS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"recommendations": recs,
		"total":           len(recs),
	})
}

// GetExecutionInsights handles GET /api/v1/automations/executions/{executionId}/insights
func (h *Handler) GetExecutionInsights(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	execID, err := strconv.ParseInt(chi.URLParam(r, "executionId"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid execution ID", "INVALID_ID")
		return
	}

	insights, err := h.svc.GetExecutionInsights(r.Context(), userCtx.OrgID, execID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INSIGHTS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"insights": insights,
		"total":    len(insights),
	})
}

// ListInsights handles GET /api/v1/automations/insights
func (h *Handler) ListInsights(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := OperationalInsightFilter{
		Page:         page,
		Limit:        limit,
		SourceModule: q.Get("source_module"),
		Status:       q.Get("status"),
		Severity:     q.Get("severity"),
		Priority:     q.Get("priority"),
		Search:       q.Get("search"),
		SortBy:       q.Get("sort_by"),
		SortDir:      q.Get("sort_dir"),
	}

	if autoIDStr := q.Get("automation_id"); autoIDStr != "" {
		if id, err := strconv.ParseInt(autoIDStr, 10, 64); err == nil {
			filter.AutomationID = &id
		}
	}
	if execIDStr := q.Get("execution_id"); execIDStr != "" {
		if id, err := strconv.ParseInt(execIDStr, 10, 64); err == nil {
			filter.ExecutionID = &id
		}
	}
	if recIDStr := q.Get("source_record_id"); recIDStr != "" {
		if id, err := strconv.ParseInt(recIDStr, 10, 64); err == nil {
			filter.SourceRecordID = &id
		}
	}

	insights, total, err := h.svc.ListInsights(r.Context(), userCtx.OrgID, filter)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INSIGHTS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"insights": insights,
		"total":    total,
		"page":     filter.Page,
		"limit":    filter.Limit,
	})
}

// GetInsight handles GET /api/v1/automations/insights/{id}
func (h *Handler) GetInsight(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid insight ID", "INVALID_ID")
		return
	}

	insight, err := h.svc.GetInsight(r.Context(), userCtx.OrgID, id)
	if err != nil {
		if errors.Is(err, ErrInsightNotFound) {
			utils.Error(w, http.StatusNotFound, "Operational insight not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INSIGHT_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"insight": insight,
	})
}

// AcknowledgeInsight handles POST /api/v1/automations/insights/{id}/acknowledge
func (h *Handler) AcknowledgeInsight(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid insight ID", "INVALID_ID")
		return
	}

	insight, err := h.svc.AcknowledgeInsight(r.Context(), userCtx.OrgID, id, userCtx.UserID)
	if err != nil {
		if errors.Is(err, ErrInsightNotFound) {
			utils.Error(w, http.StatusNotFound, "Operational insight not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "ACKNOWLEDGE_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Operational insight acknowledged",
		"insight": insight,
	})
}

// DismissInsight handles POST /api/v1/automations/insights/{id}/dismiss
func (h *Handler) DismissInsight(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid insight ID", "INVALID_ID")
		return
	}

	var reqBody struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&reqBody)

	insight, err := h.svc.DismissInsight(r.Context(), userCtx.OrgID, id, userCtx.UserID, reqBody.Reason)
	if err != nil {
		if errors.Is(err, ErrInsightNotFound) {
			utils.Error(w, http.StatusNotFound, "Operational insight not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "DISMISS_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Operational insight dismissed",
		"insight": insight,
	})
}

// EvaluateEvent handles POST /api/v1/automations/evaluate
func (h *Handler) EvaluateEvent(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	var input EvaluateEventInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON payload", "INVALID_INPUT")
		return
	}

	result, err := h.svc.EvaluateBusinessEvent(r.Context(), userCtx.OrgID, userCtx.UserID, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EVALUATION_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"result":  result,
	})
}
