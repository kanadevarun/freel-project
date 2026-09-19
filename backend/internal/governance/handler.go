package governance

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
	r.Get("/policy", h.HandleGetPolicy)
	r.Put("/policy", h.HandleUpdatePolicy)
	r.Get("/kill-switches", h.HandleGetKillSwitches)
	r.Post("/kill-switches", h.HandleSetKillSwitch)
	r.Get("/violations", h.HandleListViolations)
	r.Post("/inspect", h.HandleInspect)
	r.Post("/evaluate", h.HandleEvaluate)
}

func (h *Handler) getUser(r *http.Request) (middleware.UserContext, bool) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		return middleware.UserContext{}, false
	}
	return userCtx, true
}

func (h *Handler) HandleGetOverview(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	overview, err := h.svc.GetOverview(r.Context(), int64(user.OrgID))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.JSON(w, http.StatusOK, overview)
}

func (h *Handler) HandleGetPolicy(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	policy, err := h.svc.GetPolicy(r.Context(), int64(user.OrgID))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.JSON(w, http.StatusOK, policy)
}

func (h *Handler) HandleUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	var policy GovernancePolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_ARGUMENT")
		return
	}
	policy.OrgID = int64(user.OrgID)

	if err := h.svc.UpdatePolicy(r.Context(), int64(user.OrgID), &policy); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Governance policy updated"})
}

func (h *Handler) HandleGetKillSwitches(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	switches, err := h.svc.GetKillSwitches(r.Context(), int64(user.OrgID))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"kill_switches": switches,
		"total":         len(switches),
	})
}

func (h *Handler) HandleSetKillSwitch(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	var input SetKillSwitchInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_ARGUMENT")
		return
	}

	uid := int64(user.UserID)
	if err := h.svc.SetKillSwitch(r.Context(), int64(user.OrgID), input, &uid); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}

	actionDesc := "deactivated"
	if input.IsKilled {
		actionDesc = "ACTIVATED (Blocked)"
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Kill switch " + actionDesc + " for " + input.TargetIdentifier,
	})
}

func (h *Handler) HandleListViolations(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	violations, total, err := h.svc.ListViolations(r.Context(), int64(user.OrgID), limit, offset)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"violations": violations,
		"total":      total,
	})
}

func (h *Handler) HandleInspect(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	var req PreExecutionCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_ARGUMENT")
		return
	}

	uid := int64(user.UserID)
	result, err := h.svc.EnforcePreExecution(r.Context(), int64(user.OrgID), &uid, req)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.JSON(w, http.StatusOK, result)
}

func (h *Handler) HandleEvaluate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	var req struct {
		WorkflowName string                   `json:"workflow_name"`
		TestCases    []map[string]interface{} `json:"test_cases"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_ARGUMENT")
		return
	}

	evalResult, err := h.svc.EvaluateQuality(r.Context(), int64(user.OrgID), req.WorkflowName, req.TestCases)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.JSON(w, http.StatusOK, evalResult)
}
