package notifications

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/freel/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

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

func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	err = h.service.MarkAsRead(ctx, int32(userCtx.OrgID), int32(id))
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to mark notification as read")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": "success",
	})
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}
