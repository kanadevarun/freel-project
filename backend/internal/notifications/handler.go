package notifications

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

// GetNotifications lists notifications for the authenticated user and organization with filtering and pagination.
func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query()
	filter := NotificationFilter{
		OrgID:        userCtx.OrgID,
		UserID:       userCtx.UserID,
		UserRole:     userCtx.Role,
		Severity:     q.Get("severity"),
		SourceModule: q.Get("module"),
		Search:       q.Get("search"),
	}

	if readStr := q.Get("is_read"); readStr != "" {
		if b, err := strconv.ParseBool(readStr); err == nil {
			filter.IsRead = &b
		}
	}

	if actionStr := q.Get("action_required"); actionStr != "" {
		if b, err := strconv.ParseBool(actionStr); err == nil {
			filter.ActionRequired = &b
		}
	}

	if escStr := q.Get("is_escalated"); escStr != "" {
		if b, err := strconv.ParseBool(escStr); err == nil {
			filter.IsEscalated = &b
		}
	}

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	filter.Limit = pageSize
	filter.Offset = (page - 1) * pageSize

	items, total, err := h.service.ListNotifications(ctx, filter)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to fetch notifications: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":      items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetNotificationByID fetches a single notification ensuring tenant isolation.
func (h *Handler) GetNotificationByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	notif, err := h.service.GetNotification(ctx, userCtx.OrgID, id)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get notification")
		return
	}
	if notif == nil {
		h.respondError(w, http.StatusNotFound, "notification not found")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": notif,
	})
}

// GetUnread is retained for legacy route /api/v1/notifications/unread.
func (h *Handler) GetUnread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	notifs, err := h.service.GetUnreadNotifications(ctx, int32(userCtx.OrgID))
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get notifications")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": notifs,
	})
}

// GetUnreadCount returns the count of unread notifications for badge rendering.
func (h *Handler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	count, err := h.service.GetUnreadCount(ctx, userCtx.OrgID, userCtx.UserID, userCtx.Role)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get unread count")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"count": count,
	})
}

// GetStats returns counts broken down by unread, action-required, escalated, and critical.
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	stats, err := h.service.GetStats(ctx, userCtx.OrgID, userCtx.UserID, userCtx.Role)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get notification stats")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": stats,
	})
}

// MarkAsRead marks a notification as read.
func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	err = h.service.MarkRead(ctx, userCtx.OrgID, id, true)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondError(w, http.StatusNotFound, "notification not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to mark notification as read")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": "success",
	})
}

// MarkAsUnread marks a notification as unread.
func (h *Handler) MarkAsUnread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	err = h.service.MarkRead(ctx, userCtx.OrgID, id, false)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondError(w, http.StatusNotFound, "notification not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to mark notification as unread")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": "success",
	})
}

// Dismiss soft-dismisses a notification.
func (h *Handler) Dismiss(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	err = h.service.Dismiss(ctx, userCtx.OrgID, id)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to dismiss notification")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": "success",
	})
}

// MarkAllAsRead marks all accessible unread notifications as read.
func (h *Handler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	err := h.service.MarkAllAsRead(ctx, userCtx.OrgID, userCtx.UserID, userCtx.Role)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to mark all notifications as read")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": "success",
	})
}

// GetEscalations lists auditable escalation transitions for the organization.
func (h *Handler) GetEscalations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	events, err := h.service.ListEscalations(ctx, userCtx.OrgID, limit)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to list escalation events")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": events,
	})
}

// GetPreferences returns user notification preferences.
func (h *Handler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	prefs, err := h.service.GetPreferences(ctx, userCtx.OrgID, userCtx.UserID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get preferences")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": prefs,
	})
}

// UpdatePreferences updates user notification preferences.
func (h *Handler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var payload UserNotificationPreferences
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	payload.OrgID = userCtx.OrgID
	payload.UserID = userCtx.UserID
	if strings.TrimSpace(payload.MinSeverity) == "" {
		payload.MinSeverity = SeverityInformational
	}

	err := h.service.UpdatePreferences(ctx, &payload)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to update preferences")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": payload,
	})
}

// Evaluate triggers an evaluation sweep of all real business records.
func (h *Handler) Evaluate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	created, err := h.service.EvaluateNotifications(ctx, userCtx.OrgID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "evaluation failed: "+err.Error())
		return
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"status": "success", "created_count": created})
}

// Acknowledge marks a notification as explicitly acknowledged by an operator.
func (h *Handler) Acknowledge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	err = h.service.Acknowledge(ctx, userCtx.OrgID, id, userCtx.UserID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to acknowledge notification: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "notification acknowledged",
	})
}

// Snooze temporarily silences a notification for the specified duration.
func (h *Handler) Snooze(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	var reqBody struct {
		DurationMinutes int `json:"duration_minutes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&reqBody)
	if reqBody.DurationMinutes <= 0 {
		reqBody.DurationMinutes = 60
	}

	err = h.service.Snooze(ctx, userCtx.OrgID, id, reqBody.DurationMinutes)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to snooze notification: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "notification snoozed successfully",
	})
}

// Escalate triggers an AI-assisted escalation event with approval gating.
func (h *Handler) Escalate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	var reqBody struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&reqBody)
	if strings.TrimSpace(reqBody.Reason) == "" {
		reqBody.Reason = "Unresolved operational signal requiring managerial intervention"
	}

	escEvent, err := h.service.EscalateWithAI(ctx, userCtx.OrgID, id, userCtx.UserID, reqBody.Reason)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to escalate notification: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": escEvent,
	})
}

// AnalyzeAI calls the Python AI sidecar to evaluate priority, escalation reasoning, and overload reduction.
func (h *Handler) AnalyzeAI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	analysis, err := h.service.AnalyzeWithAI(ctx, userCtx.OrgID, id)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "AI notification analysis failed: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": analysis,
	})
}

// GenerateDraft calls the Python AI sidecar to draft an internal memo or external advisory.
func (h *Handler) GenerateDraft(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	var reqBody struct {
		DraftType string `json:"draft_type"`
	}
	_ = json.NewDecoder(r.Body).Decode(&reqBody)
	if strings.TrimSpace(reqBody.DraftType) == "" {
		reqBody.DraftType = "INTERNAL_ESCALATION"
	}

	draft, err := h.service.GenerateDraftWithAI(ctx, userCtx.OrgID, id, reqBody.DraftType)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "AI draft generation failed: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": draft,
	})
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
