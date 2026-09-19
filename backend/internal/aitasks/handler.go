package aitasks

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// ListTasks handles GET /api/v1/ai/tasks
func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	filter := TaskFilter{
		Status:     r.URL.Query().Get("status"),
		TaskType:   r.URL.Query().Get("task_type"),
		EntityType: r.URL.Query().Get("entity_type"),
		EntityID:   r.URL.Query().Get("entity_id"),
		ThreadID:   r.URL.Query().Get("thread_id"),
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		filter.Limit, _ = strconv.Atoi(limitStr)
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		filter.Offset, _ = strconv.Atoi(offsetStr)
	}

	tasks, total, err := h.svc.ListTasks(r.Context(), userCtx.OrgID, filter)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to list tasks: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Tasks retrieved successfully", map[string]interface{}{
		"tasks":  tasks,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetTaskStats handles GET /api/v1/ai/tasks/stats
func (h *Handler) GetTaskStats(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	stats, err := h.svc.GetTaskStats(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve task stats: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Task stats retrieved successfully", stats)
}

// GetTaskByID handles GET /api/v1/ai/tasks/{id}
func (h *Handler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid task ID", "INVALID_PARAM")
		return
	}

	task, err := h.svc.GetTaskByID(r.Context(), userCtx.OrgID, id)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			utils.Error(w, http.StatusNotFound, "Task not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve task: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Task retrieved successfully", task)
}

// CancelTask handles POST /api/v1/ai/tasks/{id}/cancel
func (h *Handler) CancelTask(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid task ID", "INVALID_PARAM")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	actorName := "User #" + strconv.FormatInt(userCtx.UserID, 10)
	updated, err := h.svc.CancelTask(r.Context(), userCtx.OrgID, id, actorName, req.Reason)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			utils.Error(w, http.StatusNotFound, "Task not found", "NOT_FOUND")
			return
		}
		if errors.Is(err, ErrTerminalState) {
			utils.Error(w, http.StatusConflict, err.Error(), "CONFLICT")
			return
		}
		utils.Error(w, http.StatusBadRequest, "Failed to cancel task: "+err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusOK, "Task cancelled successfully", updated)
}

// RetryTask handles POST /api/v1/ai/tasks/{id}/retry
func (h *Handler) RetryTask(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid task ID", "INVALID_PARAM")
		return
	}

	actorName := "User #" + strconv.FormatInt(userCtx.UserID, 10)
	updated, err := h.svc.RetryTask(r.Context(), userCtx.OrgID, id, actorName)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			utils.Error(w, http.StatusNotFound, "Task not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusBadRequest, "Failed to retry task: "+err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusOK, "Task retried successfully", updated)
}

// ── INTERNAL SERVICE HANDLERS ────────────────────────────────────────────────

// InternalClaimTask handles POST /internal/ai/tasks/claim
func (h *Handler) InternalClaimTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WorkerID      string   `json:"worker_id"`
		LeaseSeconds  int      `json:"lease_seconds"`
		TaskTypes     []string `json:"task_types"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}

	if req.WorkerID == "" {
		utils.Error(w, http.StatusBadRequest, "worker_id is required", "MISSING_PARAM")
		return
	}

	leaseDuration := 5 * time.Minute
	if req.LeaseSeconds > 0 {
		leaseDuration = time.Duration(req.LeaseSeconds) * time.Second
	}

	task, err := h.svc.ClaimNextTask(r.Context(), req.WorkerID, leaseDuration, req.TaskTypes)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to claim task: "+err.Error(), "CLAIM_FAILED")
		return
	}

	if task == nil {
		utils.Success(w, http.StatusOK, "No eligible tasks found in queue", nil)
		return
	}

	utils.Success(w, http.StatusOK, "Task claimed successfully", task)
}

// InternalHeartbeatTask handles POST /internal/ai/tasks/{id}/heartbeat
func (h *Handler) InternalHeartbeatTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid task ID", "INVALID_PARAM")
		return
	}

	var req struct {
		WorkerID     string `json:"worker_id"`
		LeaseSeconds int    `json:"lease_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}

	leaseDuration := 5 * time.Minute
	if req.LeaseSeconds > 0 {
		leaseDuration = time.Duration(req.LeaseSeconds) * time.Second
	}

	if err := h.svc.HeartbeatTask(r.Context(), id, req.WorkerID, leaseDuration); err != nil {
		utils.Error(w, http.StatusConflict, err.Error(), "HEARTBEAT_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Heartbeat acknowledged", nil)
}

// InternalUpdateStatus handles POST /internal/ai/tasks/{id}/status
func (h *Handler) InternalUpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid task ID", "INVALID_PARAM")
		return
	}

	var req struct {
		OrgID         int64  `json:"org_id"`
		Status        string `json:"status"`
		WorkerID      string `json:"worker_id"`
		ErrorMessage  string `json:"error_message"`
		LastErrorCode string `json:"last_error_code"`
		ThreadID      string `json:"thread_id"`
		CorrelationID string `json:"correlation_id"`
		ApprovalID    *int64 `json:"approval_id"`
		DelaySeconds  int    `json:"delay_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}

	input := UpdateTaskStatusInput{
		Status:        req.Status,
		WorkerID:      req.WorkerID,
		ErrorMessage:  req.ErrorMessage,
		LastErrorCode: req.LastErrorCode,
		ThreadID:      req.ThreadID,
		CorrelationID: req.CorrelationID,
		ApprovalID:    req.ApprovalID,
		DelaySeconds:  req.DelaySeconds,
	}

	updated, err := h.svc.UpdateTaskStatus(r.Context(), req.OrgID, id, input)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			utils.Error(w, http.StatusNotFound, "Task not found", "NOT_FOUND")
			return
		}
		if errors.Is(err, ErrInvalidTransition) || errors.Is(err, ErrTerminalState) {
			utils.Error(w, http.StatusConflict, err.Error(), "CONFLICT")
			return
		}
		if errors.Is(err, ErrTenantMismatch) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusBadRequest, "Failed to update task: "+err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusOK, "Task status updated successfully", updated)
}

// InternalRecoverStale handles POST /internal/ai/tasks/recover-stale
func (h *Handler) InternalRecoverStale(w http.ResponseWriter, r *http.Request) {
	recovered, failed, err := h.svc.RecoverStaleTasks(r.Context(), 5*time.Minute)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Recovery failed: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Stale task recovery complete", map[string]interface{}{
		"recovered_to_queued": recovered,
		"marked_failed":       failed,
	})
}

// InternalGetTask handles GET /internal/ai/tasks/{id}
func (h *Handler) InternalGetTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid task ID", "INVALID_PARAM")
		return
	}

	task, err := h.svc.GetTaskByIDUnscoped(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			utils.Error(w, http.StatusNotFound, "Task not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve task: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Task retrieved successfully", task)
}

// GetWorkforceSummary handles GET /api/v1/ai/workforce/summary
func (h *Handler) GetWorkforceSummary(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Authenticated organization context required", "UNAUTHORIZED")
		return
	}

	summary, err := h.svc.GetWorkforceSummary(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve AI workforce summary: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "AI workforce summary retrieved successfully", summary)
}

// ListWorkforceTasks handles GET /api/v1/ai/workforce/tasks
func (h *Handler) ListWorkforceTasks(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Authenticated organization context required", "UNAUTHORIZED")
		return
	}

	filter := WorkforceTaskFilter{
		Status:   r.URL.Query().Get("status"),
		AgentKey: r.URL.Query().Get("agent_key"),
		Module:   r.URL.Query().Get("module"),
		Search:   r.URL.Query().Get("search"),
	}

	if reqApp := r.URL.Query().Get("requires_approval"); reqApp != "" {
		val := (reqApp == "true" || reqApp == "1")
		filter.RequiresApproval = &val
	}
	if failStale := r.URL.Query().Get("failed_or_stale"); failStale != "" {
		val := (failStale == "true" || failStale == "1")
		filter.FailedOrStale = &val
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		filter.Limit, _ = strconv.Atoi(limitStr)
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		filter.Offset, _ = strconv.Atoi(offsetStr)
	}

	tasks, total, err := h.svc.ListWorkforceTasks(r.Context(), userCtx.OrgID, filter)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to list workforce tasks: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Workforce tasks retrieved successfully", map[string]interface{}{
		"tasks":  tasks,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetWorkforceHealth handles GET /api/v1/ai/workforce/health
func (h *Handler) GetWorkforceHealth(w http.ResponseWriter, r *http.Request) {
	health, err := h.svc.GetWorkforceHealth(r.Context())
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to check workforce health: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Workforce health retrieved successfully", health)
}
