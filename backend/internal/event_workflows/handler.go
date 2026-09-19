package event_workflows

import (
	"encoding/json"
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

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/overview", h.HandleGetOverview)
	r.Get("/events", h.HandleListEvents)
	r.Get("/events/{id:[0-9]+}", h.HandleGetEvent)
	r.Get("/instances", h.HandleListWorkflows)
	r.Get("/instances/{id:[0-9]+}", h.HandleGetWorkflow)
	r.Post("/instances/{id:[0-9]+}/retry", h.HandleRetryWorkflow)
	r.Post("/instances/{id:[0-9]+}/cancel", h.HandleCancelWorkflow)
	r.Post("/simulate-event", h.HandleSimulateEvent)
}

func (h *Handler) getUserCtx(r *http.Request) (int64, int64, error) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		return 0, 0, http.ErrNoCookie
	}
	return userCtx.OrgID, userCtx.UserID, nil
}

func (h *Handler) HandleGetOverview(w http.ResponseWriter, r *http.Request) {
	orgID, _, err := h.getUserCtx(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	overview, err := h.svc.GetOverview(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.JSON(w, http.StatusOK, overview)
}

func (h *Handler) HandleListEvents(w http.ResponseWriter, r *http.Request) {
	orgID, _, err := h.getUserCtx(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	events, total, err := h.svc.ListEvents(r.Context(), orgID, limit, offset)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) HandleGetEvent(w http.ResponseWriter, r *http.Request) {
	orgID, _, err := h.getUserCtx(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid event id", "BAD_REQUEST")
		return
	}

	evt, err := h.svc.GetEventByID(r.Context(), orgID, id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	if evt == nil {
		utils.Error(w, http.StatusNotFound, "Event not found", "NOT_FOUND")
		return
	}
	utils.JSON(w, http.StatusOK, evt)
}

func (h *Handler) HandleListWorkflows(w http.ResponseWriter, r *http.Request) {
	orgID, _, err := h.getUserCtx(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	workflows, total, err := h.svc.ListWorkflowInstances(r.Context(), orgID, limit, offset)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"workflows": workflows,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

func (h *Handler) HandleGetWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, _, err := h.getUserCtx(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid workflow id", "BAD_REQUEST")
		return
	}

	wf, err := h.svc.GetWorkflowInstanceByID(r.Context(), orgID, id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	if wf == nil {
		utils.Error(w, http.StatusNotFound, "Workflow not found", "NOT_FOUND")
		return
	}
	utils.JSON(w, http.StatusOK, wf)
}

func (h *Handler) HandleRetryWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, userID, err := h.getUserCtx(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid workflow id", "BAD_REQUEST")
		return
	}

	wf, err := h.svc.RetryWorkflow(r.Context(), orgID, id, userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "RETRY_FAILED")
		return
	}
	utils.JSON(w, http.StatusOK, wf)
}

func (h *Handler) HandleCancelWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, _, err := h.getUserCtx(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid workflow id", "BAD_REQUEST")
		return
	}

	var p struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&p)
	if p.Reason == "" {
		p.Reason = "Operator requested cancellation"
	}

	wf, err := h.svc.CancelWorkflow(r.Context(), orgID, id, p.Reason)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "CANCEL_FAILED")
		return
	}
	utils.JSON(w, http.StatusOK, wf)
}

func (h *Handler) HandleSimulateEvent(w http.ResponseWriter, r *http.Request) {
	orgID, userID, err := h.getUserCtx(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	var input IngestEventInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	wf, err := h.svc.IngestAndProcessEvent(r.Context(), orgID, userID, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EVENT_PROCESSING_FAILED")
		return
	}
	utils.JSON(w, http.StatusCreated, wf)
}
