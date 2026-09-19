package memory

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

func getActorName(userCtx middleware.UserContext) string {
	return fmt.Sprintf("User #%d", userCtx.UserID)
}

// Handler handles HTTP requests for AI memory, preferences, and personalization
type Handler struct {
	svc Service
}

// NewHandler instantiates a memory HTTP handler
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// ListMemories handles GET /api/v1/memory
func (h *Handler) ListMemories(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := MemoryFilter{
		Scope:      q.Get("scope"),
		MemoryType: q.Get("memory_type"),
		Status:     q.Get("status"),
		Search:     q.Get("search"),
		Limit:      limit,
		Offset:     offset,
	}

	items, total, err := h.svc.ListMemories(r.Context(), userCtx.OrgID, userCtx.UserID, filter)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "MEMORY_LIST_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    items,
		"total":   total,
		"limit":   filter.Limit,
		"offset":  filter.Offset,
	})
}

// GetMemory handles GET /api/v1/memory/{id}
func (h *Handler) GetMemory(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid memory item ID", "INVALID_ID")
		return
	}

	item, err := h.svc.GetMemory(r.Context(), userCtx.OrgID, userCtx.UserID, id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error(), "MEMORY_NOT_FOUND")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    item,
	})
}

// ProposeMemory handles POST /api/v1/memory/propose
func (h *Handler) ProposeMemory(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	var input ProposeMemoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	proposal, err := h.svc.ProposeMemory(r.Context(), userCtx.OrgID, userCtx.UserID, userCtx.Role, input)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "MEMORY_PROPOSAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    proposal,
	})
}

// CreateMemory handles POST /api/v1/memory
func (h *Handler) CreateMemory(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	var input CreateMemoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	item, err := h.svc.CreateMemory(r.Context(), userCtx.OrgID, userCtx.UserID, getActorName(userCtx), userCtx.Role, input)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "MEMORY_CREATE_ERROR")
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    item,
	})
}

// UpdateMemory handles PUT /api/v1/memory/{id}
func (h *Handler) UpdateMemory(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid memory item ID", "INVALID_ID")
		return
	}

	var input UpdateMemoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	updated, err := h.svc.UpdateMemory(r.Context(), userCtx.OrgID, userCtx.UserID, getActorName(userCtx), userCtx.Role, id, input)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "MEMORY_UPDATE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    updated,
	})
}

// DeleteMemory handles DELETE /api/v1/memory/{id}
func (h *Handler) DeleteMemory(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid memory item ID", "INVALID_ID")
		return
	}

	if err := h.svc.DeleteMemory(r.Context(), userCtx.OrgID, userCtx.UserID, getActorName(userCtx), userCtx.Role, id); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "MEMORY_DELETE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Memory item deleted successfully",
	})
}

// DisableMemory handles POST /api/v1/memory/{id}/disable
func (h *Handler) DisableMemory(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid memory item ID", "INVALID_ID")
		return
	}

	disabled, err := h.svc.DisableMemory(r.Context(), userCtx.OrgID, userCtx.UserID, getActorName(userCtx), userCtx.Role, id)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "MEMORY_DISABLE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    disabled,
	})
}

// EnableMemory handles POST /api/v1/memory/{id}/enable
func (h *Handler) EnableMemory(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid memory item ID", "INVALID_ID")
		return
	}

	enabled, err := h.svc.EnableMemory(r.Context(), userCtx.OrgID, userCtx.UserID, getActorName(userCtx), userCtx.Role, id)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "MEMORY_ENABLE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    enabled,
	})
}

// ClearPersonalMemories handles POST /api/v1/memory/clear-personal
func (h *Handler) ClearPersonalMemories(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	cleared, err := h.svc.ClearPersonalMemories(r.Context(), userCtx.OrgID, userCtx.UserID, getActorName(userCtx))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "MEMORY_CLEAR_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"cleared": cleared,
		"message": "Personal memories cleared successfully",
	})
}

// GetUserSettings handles GET /api/v1/memory/settings
func (h *Handler) GetUserSettings(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	settings, err := h.svc.GetUserSettings(r.Context(), userCtx.OrgID, userCtx.UserID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "SETTINGS_GET_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    settings,
	})
}

// UpdateUserSettings handles PUT /api/v1/memory/settings
func (h *Handler) UpdateUserSettings(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	var settings UserPersonalizationSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	updated, err := h.svc.UpdateUserSettings(r.Context(), userCtx.OrgID, userCtx.UserID, getActorName(userCtx), &settings)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "SETTINGS_UPDATE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    updated,
	})
}

// TogglePersonalization handles POST /api/v1/memory/toggle
func (h *Handler) TogglePersonalization(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	var payload struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	updated, err := h.svc.TogglePersonalization(r.Context(), userCtx.OrgID, userCtx.UserID, getActorName(userCtx), payload.Enabled)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "TOGGLE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    updated,
	})
}

// ListPreferences handles GET /api/v1/memory/preferences
func (h *Handler) ListPreferences(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	scope := r.URL.Query().Get("scope")
	prefs, err := h.svc.ListPreferences(r.Context(), userCtx.OrgID, userCtx.UserID, scope)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREF_LIST_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    prefs,
	})
}

// SetPreference handles PUT /api/v1/memory/preferences
func (h *Handler) SetPreference(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	var pref Preference
	if err := json.NewDecoder(r.Body).Decode(&pref); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	saved, err := h.svc.SetPreference(r.Context(), userCtx.OrgID, userCtx.UserID, getActorName(userCtx), userCtx.Role, &pref)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "PREF_SET_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    saved,
	})
}

// DeletePreference handles DELETE /api/v1/memory/preferences/{key}
func (h *Handler) DeletePreference(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	key := chi.URLParam(r, "key")
	scope := r.URL.Query().Get("scope")

	if err := h.svc.DeletePreference(r.Context(), userCtx.OrgID, userCtx.UserID, getActorName(userCtx), userCtx.Role, scope, key); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "PREF_DELETE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Preference deleted successfully",
	})
}

// GetRuntimeContext handles POST /api/v1/memory/runtime-context
func (h *Handler) GetRuntimeContext(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	var req RuntimeContextRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	contextResp, err := h.svc.GetRuntimeContext(r.Context(), userCtx.OrgID, userCtx.UserID, req)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "RUNTIME_CONTEXT_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    contextResp,
	})
}

// GetStats handles GET /api/v1/memory/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	stats, err := h.svc.GetStats(r.Context(), userCtx.OrgID, userCtx.UserID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "STATS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// ListAuditEvents handles GET /api/v1/memory/audit
func (h *Handler) ListAuditEvents(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	events, err := h.svc.ListAuditEvents(r.Context(), userCtx.OrgID, userCtx.UserID, limit, offset)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "AUDIT_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    events,
	})
}
