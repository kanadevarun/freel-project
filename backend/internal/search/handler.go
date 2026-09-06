package search

import (
	"net/http"
	"strconv"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) HandleGlobalSearch(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	query := r.URL.Query().Get("q")
	entityType := r.URL.Query().Get("type")
	limitStr := r.URL.Query().Get("limit")
	limit := 30
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	res, err := h.svc.Search(r.Context(), userCtx.OrgID, query, entityType, limit)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Search failed: "+err.Error(), "SEARCH_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Search results retrieved", res)
}
