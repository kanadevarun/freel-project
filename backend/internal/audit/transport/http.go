package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/audit/service"
	"github.com/freel/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for universal audit logs.
type Handler struct {
	svc service.Service
}

// NewHandler initializes a new AuditLog HTTP handler.
func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts the secure AuditLog endpoints onto the Chi router.
func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware, rbacMiddleware *middleware.RBACMiddleware) {
	r.Route("/api/v1/settings/audit-logs", func(sub chi.Router) {
		sub.Use(authMiddleware.RequireAuth)
		sub.Use(rbacMiddleware.RequirePermission("SETTINGS", "READ"))

		sub.Get("/", h.ListAuditLogs)
		sub.Get("/{id}", h.GetAuditLogByID)
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error":   message,
		"status":  status,
		"success": false,
	})
}

// ListAuditLogs handles GET /api/v1/settings/audit-logs with filtering and pagination.
func (h *Handler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 20
	}

	filter := domain.AuditLogFilter{
		OrgID:        userCtx.OrgID,
		ActorType:    strings.TrimSpace(q.Get("actor_type")),
		Module:       strings.TrimSpace(q.Get("module")),
		Action:       strings.TrimSpace(q.Get("action")),
		ResourceType: strings.TrimSpace(q.Get("resource_type")),
		ResourceID:   strings.TrimSpace(q.Get("resource_id")),
		Result:       strings.TrimSpace(q.Get("result")),
		Search:       strings.TrimSpace(q.Get("search")),
		Page:         page,
		Limit:        limit,
	}

	if actorIDStr := strings.TrimSpace(q.Get("actor_id")); actorIDStr != "" {
		if uid, err := strconv.ParseInt(actorIDStr, 10, 64); err == nil && uid > 0 {
			filter.ActorID = &uid
		}
	}

	if startDateStr := strings.TrimSpace(q.Get("start_date")); startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = &t
		} else if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &t
		}
	}

	if endDateStr := strings.TrimSpace(q.Get("end_date")); endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = &t
		} else if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endOfDay := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			filter.EndDate = &endOfDay
		}
	}

	res, err := h.svc.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch audit logs: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// GetAuditLogByID handles GET /api/v1/settings/audit-logs/{id}.
func (h *Handler) GetAuditLogByID(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid audit log ID")
		return
	}

	entry, err := h.svc.GetByID(r.Context(), userCtx.OrgID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Audit log not found")
		return
	}

	writeJSON(w, http.StatusOK, entry)
}
