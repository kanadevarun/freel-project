package copilot

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

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/chat", h.HandleChat)
	r.Get("/sessions", h.HandleListSessions)
	r.Get("/sessions/{sessionId}", h.HandleGetSession)
	r.Get("/sessions/{sessionId}/messages", h.HandleListMessages)
	r.Post("/sessions/{sessionId}/archive", h.HandleArchiveSession)
	r.Post("/actions/execute", h.HandleExecuteAction)
	r.Get("/actions", h.HandleListActions)
}

func (h *Handler) getUser(r *http.Request) (middleware.UserContext, bool) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		return middleware.UserContext{}, false
	}
	return userCtx, true
}

func (h *Handler) HandleChat(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	var input ChatInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_ARGUMENT")
		return
	}

	if strings.TrimSpace(input.Query) == "" {
		utils.Error(w, http.StatusBadRequest, "Query cannot be empty", "INVALID_ARGUMENT")
		return
	}

	userName := user.CognitoID
	if userName == "" {
		userName = "LogisticsHQ User"
	}

	resp, err := h.svc.Chat(r.Context(), user.OrgID, user.UserID, userName, user.Role, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleListSessions(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	sessions, err := h.svc.ListSessions(r.Context(), user.OrgID, user.UserID, limit)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	if sessions == nil {
		sessions = []CopilotSession{}
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

func (h *Handler) HandleGetSession(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		utils.Error(w, http.StatusBadRequest, "sessionId is required", "INVALID_ARGUMENT")
		return
	}

	session, err := h.svc.GetSession(r.Context(), user.OrgID, sessionID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	if session == nil {
		utils.Error(w, http.StatusNotFound, "Session not found", "NOT_FOUND")
		return
	}

	utils.JSON(w, http.StatusOK, session)
}

func (h *Handler) HandleListMessages(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		utils.Error(w, http.StatusBadRequest, "sessionId is required", "INVALID_ARGUMENT")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	messages, err := h.svc.ListMessages(r.Context(), user.OrgID, sessionID, limit)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	if messages == nil {
		messages = []CopilotMessage{}
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"messages": messages,
		"total":    len(messages),
	})
}

func (h *Handler) HandleArchiveSession(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		utils.Error(w, http.StatusBadRequest, "sessionId is required", "INVALID_ARGUMENT")
		return
	}

	if err := h.svc.ArchiveSession(r.Context(), user.OrgID, sessionID); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"session_id": sessionID,
		"archived":   true,
	})
}

func (h *Handler) HandleExecuteAction(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	var input ExecuteActionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_ARGUMENT")
		return
	}

	userName := user.CognitoID
	if userName == "" {
		userName = "LogisticsHQ User"
	}

	actionRecord, err := h.svc.ExecuteAction(r.Context(), user.OrgID, user.UserID, userName, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, actionRecord)
}

func (h *Handler) HandleListActions(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	actions, err := h.svc.ListActions(r.Context(), user.OrgID, sessionID, limit)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	if actions == nil {
		actions = []CopilotActionHistory{}
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"actions": actions,
		"total":   len(actions),
	})
}
