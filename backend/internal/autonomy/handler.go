package autonomy

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	// Planning Goals
	r.Post("/goals", h.HandleCreateGoal)
	r.Get("/goals", h.HandleListGoals)
	r.Get("/goals/{id}", h.HandleGetGoal)

	// Context Assembly
	r.Get("/context/{module}/{entityId}", h.HandleGetContext)

	// Autonomous Operational Plans & Candidates
	r.Post("/plans/generate", h.HandleGeneratePlan)
	r.Get("/plans", h.HandleListPlans)
	r.Get("/plans/{id}", h.HandleGetPlan)
	r.Get("/plans/{id}/candidates", h.HandleGetCandidates)
	r.Post("/plans/{id}/select-candidate", h.HandleSelectCandidate)
	r.Post("/plans/{id}/revalidate", h.HandleRevalidatePlan)
	r.Get("/plans/{id}/versions", h.HandleGetPlanVersions)
	r.Post("/plans/{id}/approve", h.HandleApprovePlan)
	r.Post("/plans/{id}/reject", h.HandleRejectPlan)
	r.Post("/plans/{id}/steps/{stepId}/execute", h.HandleExecuteStep)
	r.Post("/plans/{id}/pause", h.HandlePausePlan)
	r.Post("/plans/{id}/resume", h.HandleResumePlan)
	r.Post("/plans/{id}/cancel", h.HandleCancelPlan)
	r.Post("/plans/{id}/replan", h.HandleReplan)
	r.Get("/plans/{id}/audit", h.HandleGetAuditHistory)

	// Phase 5 Task 5.9: Multi-Step AI Planning and Execution
	r.Post("/plans/cross-module", h.HandleGenerateCrossModulePlan)
	r.Get("/plans/metrics", h.HandleGetPlanningMetrics)
	r.Get("/plans/conflicts", h.HandleListEntityConflicts)
	r.Post("/plans/{id}/execute-next", h.HandleExecuteNextStep)
	r.Post("/plans/{id}/steps/{stepId}/approve", h.HandleApproveStep)
	r.Post("/plans/{id}/steps/{stepId}/retry", h.HandleRetryStep)
	r.Post("/plans/{id}/steps/{stepId}/compensate", h.HandleCompensateStep)
	r.Post("/plans/{id}/validate", h.HandleValidatePlan)
	r.Get("/plans/{id}/conflicts", h.HandleCheckPlanConflicts)

	// Policies
	r.Get("/policies", h.HandleListPolicies)
	r.Get("/policies/{module}", h.HandleGetPolicy)
	r.Post("/policies", h.HandleSetPolicy)

	// Phase 5 Task 5.3: Adaptive Shipment Management
	r.Post("/shipments/{shipmentId}/events", h.HandleIngestShipmentEvent)
	r.Get("/shipments/{shipmentId}/adaptive-state", h.HandleGetShipmentAdaptiveState)
	r.Get("/shipments/{shipmentId}/events", h.HandleListShipmentEvents)
	r.Post("/plans/{id}/waiting-state", h.HandleTransitionWaitingState)

	// Phase 5 Task 5.4: Autonomous Customer Follow-Up
	r.Post("/customers/{customerId}/events", h.HandleIngestCustomerFollowupEvent)
	r.Get("/customers/{customerId}/followup-state", h.HandleGetCustomerFollowupState)
	r.Put("/customers/{customerId}/preferences", h.HandleUpdateCustomerPreferences)
	r.Post("/customers/{customerId}/preferences", h.HandleUpdateCustomerPreferences)
	r.Post("/customers/{customerId}/records/{recId}/send", h.HandleSendCustomerFollowup)
	r.Post("/customers/{customerId}/records/{recId}/response", h.HandleIngestCustomerResponse)

	// Phase 5 Task 5.5: Intelligent RFQ and Pricing Optimization
	r.Post("/pricing/rfqs/{rfqId}/evaluate", h.HandleEvaluateRfqPricing)
	r.Get("/pricing/rfqs/{rfqId}/state", h.HandleGetRfqPricingState)
	r.Post("/pricing/rfqs/{rfqId}/select-strategy", h.HandleSelectPricingStrategy)
	r.Post("/pricing/rfqs/{rfqId}/execute-quote", h.HandleExecutePricingQuotation)
	r.Post("/pricing/rfqs/{rfqId}/replan", h.HandleReplanPricing)

	// Phase 5 Task 5.6: Adaptive Finance and Collections
	r.Post("/finance/invoices/{invoiceId}/evaluate", h.HandleEvaluateFinanceCollection)
	r.Get("/finance/invoices/{invoiceId}/state", h.HandleGetFinanceCollectionState)
	r.Post("/finance/invoices/{invoiceId}/select-strategy", h.HandleSelectFinanceCollectionStrategy)
	r.Post("/finance/invoices/{invoiceId}/execute-action", h.HandleExecuteFinanceCollectionAction)
	r.Post("/finance/invoices/{invoiceId}/replan", h.HandleReplanFinanceCollection)

	// Phase 5 Task 5.7: Contract and Compliance Monitoring
	r.Post("/compliance/contracts/{contractId}/evaluate", h.HandleEvaluateContractCompliance)
	r.Get("/compliance/contracts/{contractId}/state", h.HandleGetContractComplianceState)
	r.Post("/compliance/contracts/{contractId}/select-strategy", h.HandleSelectContractComplianceStrategy)
	r.Post("/compliance/contracts/{contractId}/execute-action", h.HandleExecuteContractComplianceAction)
	r.Post("/compliance/contracts/{contractId}/replan", h.HandleReplanContractCompliance)

	// Phase 5 Task 5.8: Autonomous Exception Resolution
	r.Post("/exceptions/{exceptionId}/evaluate", h.HandleEvaluateExceptionResolution)
	r.Get("/exceptions/{exceptionId}/state", h.HandleGetExceptionResolutionState)
	r.Post("/exceptions/{exceptionId}/select-strategy", h.HandleSelectExceptionResolutionStrategy)
	r.Post("/exceptions/{exceptionId}/execute-action", h.HandleExecuteExceptionResolutionAction)
	r.Post("/exceptions/{exceptionId}/replan", h.HandleReplanExceptionResolution)

	// Phase 5 Task 5.10: Continuous Monitoring and Replanning
	r.Post("/monitoring/events", h.HandleIngestMonitoringEvent)
	r.Get("/monitoring/events", h.HandleListMonitoringEvents)
	r.Get("/monitoring/plans/{id}/health", h.HandleGetPlanHealth)
	r.Post("/monitoring/plans/{id}/replan", h.HandleTriggerAdaptiveReplan)
	r.Get("/monitoring/metrics", h.HandleGetContinuousMonitoringMetrics)

	// Phase 5 Task 5.11: Human + AI Operating Model
	r.Post("/human-ai/decisions", h.HandleCreateDecisionPoint)
	r.Get("/human-ai/decisions", h.HandleListDecisionPoints)
	r.Get("/human-ai/decisions/{decisionId}", h.HandleGetDecisionPoint)
	r.Post("/human-ai/decisions/{decisionId}/decide", h.HandleSubmitHumanDecision)
	r.Post("/human-ai/plans/{id}/stop", h.HandleStopWorkflow)
	r.Post("/human-ai/plans/{id}/steps/{stepId}/edit", h.HandleUpdateStepHumanEdit)
	r.Post("/human-ai/invalidate-approvals", h.HandleInvalidateApprovals)
	r.Get("/human-ai/decision-center/summary", h.HandleGetDecisionCenterSummary)

	// Phase 5 Task 5.12: Autonomous Operations Command Center
	r.Get("/command-center/overview", h.HandleGetCommandCenterOverview)
	r.Get("/command-center/critical-attention", h.HandleGetCommandCenterCriticalAttention)
	r.Get("/command-center/workflows", h.HandleGetCommandCenterWorkflows)
	r.Get("/command-center/decisions", h.HandleGetCommandCenterDecisions)
	r.Get("/command-center/risks", h.HandleGetCommandCenterRisks)
	r.Get("/command-center/activity", h.HandleGetCommandCenterActivity)
	r.Get("/command-center/system-health", h.HandleGetCommandCenterSystemHealth)

	// Phase 5 Task 5.13: Agent Memory and Learning from Outcomes
	r.Post("/memory/outcomes", h.HandleRecordOutcome)
	r.Get("/memory/outcomes", h.HandleListOutcomes)
	r.Get("/memory/outcomes/{outcomeId}", h.HandleGetOutcome)
	r.Post("/memory/outcomes/{outcomeId}/verify", h.HandleVerifyOutcome)

	r.Post("/memory/retrieve", h.HandleRetrieveMemory)
	r.Get("/memory/items", h.HandleListMemories)
	r.Get("/memory/items/{memoryId}", h.HandleGetMemory)
	r.Put("/memory/items/{memoryId}/correct", h.HandleCorrectMemory)
	r.Post("/memory/items/{memoryId}/invalidate", h.HandleInvalidateMemory)
	r.Post("/memory/items/{memoryId}/flag-unreliable", h.HandleFlagMemoryUnreliable)

	r.Post("/memory/patterns/detect", h.HandleDetectPatterns)
	r.Get("/memory/patterns", h.HandleListPatterns)
	r.Get("/memory/summary", h.HandleGetMemoryLearningSummary)

	// Phase 5 Task 5.14: Governance for Controlled Autonomy
	r.Post("/governance/evaluate", h.HandleEvaluateGovernance)
	r.Get("/governance/limits", h.HandleGetTenantLimits)
	r.Put("/governance/limits", h.HandleUpdateTenantLimits)
	r.Post("/governance/kill-switch", h.HandleToggleKillSwitch)
	r.Get("/governance/allowlist", h.HandleGetActionAllowlist)
	r.Post("/governance/allowlist", h.HandleUpsertActionAllowlistItem)
	r.Get("/governance/flags", h.HandleGetFeatureFlags)
	r.Put("/governance/flags/{flagKey}", h.HandleUpdateFeatureFlag)
	r.Get("/governance/evaluations", h.HandleListPolicyEvaluations)
	r.Get("/governance/audit-logs", h.HandleListPolicyAuditLogs)
	r.Get("/governance/telemetry", h.HandleGetGovernanceTelemetry)
	r.Post("/governance/preview", h.HandlePreviewPlan)
}



func (h *Handler) getUser(r *http.Request) (middleware.UserContext, bool) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		return middleware.UserContext{}, false
	}
	return userCtx, true
}

// -----------------------------------------------------------------------------
// Goal Handlers
// -----------------------------------------------------------------------------

func (h *Handler) HandleCreateGoal(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	var req CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_ARGUMENT")
		return
	}

	if req.Objective == "" {
		utils.Error(w, http.StatusBadRequest, "Missing required field: objective", "VALIDATION_ERROR")
		return
	}

	var uid *int64
	if user.UserID > 0 {
		val := int64(user.UserID)
		uid = &val
	}

	goal, plan, steps, err := h.svc.CreatePlanningGoal(r.Context(), int64(user.OrgID), uid, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "GOAL_CREATION_FAILED")
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"goal":    goal,
		"plan":    plan,
		"steps":   steps,
	})
}

func (h *Handler) HandleGetGoal(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	goalID := chi.URLParam(r, "id")
	goal, err := h.svc.GetGoal(r.Context(), int64(user.OrgID), goalID)
	if err != nil {
		if err == ErrGoalNotFound {
			utils.Error(w, http.StatusNotFound, "Goal not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"goal":    goal,
	})
}

func (h *Handler) HandleListGoals(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	module := r.URL.Query().Get("module")
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	goals, total, err := h.svc.ListGoals(r.Context(), int64(user.OrgID), module, status, limit, offset)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"goals":   goals,
		"total":   total,
	})
}

func (h *Handler) HandleGetContext(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	module := chi.URLParam(r, "module")
	entityID := chi.URLParam(r, "entityId")
	entityType := r.URL.Query().Get("type")
	if entityType == "" {
		entityType = "SHIPMENT"
	}

	assemb, err := h.svc.GetContext(r.Context(), int64(user.OrgID), module, entityType, entityID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"context": assemb,
	})
}

// -----------------------------------------------------------------------------
// Plan Handlers
// -----------------------------------------------------------------------------

func (h *Handler) HandleGeneratePlan(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	var req GeneratePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_ARGUMENT")
		return
	}

	if req.Goal == "" || req.Module == "" || req.RelatedEntityType == "" || req.RelatedEntityID == "" {
		utils.Error(w, http.StatusBadRequest, "Missing mandatory plan fields: goal, module, related_entity_type, related_entity_id", "VALIDATION_ERROR")
		return
	}

	var uid *int64
	if user.UserID > 0 {
		val := int64(user.UserID)
		uid = &val
	}

	plan, steps, err := h.svc.GeneratePlan(r.Context(), int64(user.OrgID), uid, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "PLAN_GENERATION_FAILED")
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"plan":    plan,
		"steps":   steps,
	})
}

func (h *Handler) HandleGetPlan(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	plan, steps, err := h.svc.GetPlan(r.Context(), int64(user.OrgID), planID)
	if err != nil {
		if err == ErrPlanNotFound {
			utils.Error(w, http.StatusNotFound, "Plan not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"plan":    plan,
		"steps":   steps,
	})
}

func (h *Handler) HandleGetCandidates(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	plan, _, err := h.svc.GetPlan(r.Context(), int64(user.OrgID), planID)
	if err != nil {
		if err == ErrPlanNotFound {
			utils.Error(w, http.StatusNotFound, "Plan not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	var candidates []interface{}
	if len(plan.CandidatePlans) > 0 {
		_ = json.Unmarshal(plan.CandidatePlans, &candidates)
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":               true,
		"plan_id":               plan.PlanID,
		"selected_candidate_id": plan.SelectedCandidateID.String,
		"candidates":            candidates,
		"evaluation_summary":    plan.EvaluationSummary,
	})
}

func (h *Handler) HandleSelectCandidate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	var req SelectCandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_ARGUMENT")
		return
	}

	if req.CandidateID == "" {
		utils.Error(w, http.StatusBadRequest, "candidate_id is required", "VALIDATION_ERROR")
		return
	}

	plan, steps, err := h.svc.SelectPlanCandidate(r.Context(), int64(user.OrgID), int64(user.UserID), planID, req.CandidateID)
	if err != nil {
		if errors.Is(err, ErrCandidateInfeasible) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "HARD_CONSTRAINT_VIOLATION")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "CANDIDATE_SELECTION_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"plan":    plan,
		"steps":   steps,
	})
}

func (h *Handler) HandleRevalidatePlan(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	plan, err := h.svc.RevalidatePlan(r.Context(), int64(user.OrgID), planID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "REVALIDATION_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":          true,
		"plan_id":          plan.PlanID,
		"staleness_status": plan.StalenessStatus,
		"plan":             plan,
	})
}

func (h *Handler) HandleListPlans(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	module := r.URL.Query().Get("module")
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	plans, total, err := h.svc.ListPlans(r.Context(), int64(user.OrgID), module, status, limit, offset)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"plans":   plans,
		"total":   total,
	})
}

func (h *Handler) HandleGetPlanVersions(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	versions, err := h.svc.GetPlanVersions(r.Context(), int64(user.OrgID), planID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"versions": versions,
	})
}

func (h *Handler) HandleApprovePlan(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	var body struct {
		Notes string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	if err := h.svc.ApprovePlan(r.Context(), int64(user.OrgID), int64(user.UserID), planID, body.Notes); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "APPROVAL_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Plan successfully approved",
	})
}

func (h *Handler) HandleRejectPlan(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Reason == "" {
		body.Reason = "Rejected by operations supervisor"
	}

	if err := h.svc.RejectPlan(r.Context(), int64(user.OrgID), int64(user.UserID), planID, body.Reason); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "REJECTION_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Plan rejected",
	})
}

func (h *Handler) HandleExecuteStep(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	stepID := chi.URLParam(r, "stepId")

	res, err := h.svc.ExecuteStep(r.Context(), int64(user.OrgID), int64(user.UserID), planID, stepID)
	if err != nil {
		if errors.Is(err, ErrPlanNotFound) || errors.Is(err, ErrStepNotFound) {
			utils.Error(w, http.StatusNotFound, err.Error(), "STEP_NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "STEP_EXECUTION_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"result":  res,
	})
}

func (h *Handler) HandlePausePlan(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	if err := h.svc.PausePlan(r.Context(), int64(user.OrgID), int64(user.UserID), planID, body.Reason); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "PAUSE_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Plan paused",
	})
}

func (h *Handler) HandleResumePlan(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	if err := h.svc.ResumePlan(r.Context(), int64(user.OrgID), int64(user.UserID), planID); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "RESUME_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Plan resumed",
	})
}

func (h *Handler) HandleCancelPlan(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	if err := h.svc.CancelPlan(r.Context(), int64(user.OrgID), int64(user.UserID), planID, body.Reason); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "CANCEL_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Plan cancelled",
	})
}

func (h *Handler) HandleReplan(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	var req ReplanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid replan payload", "INVALID_ARGUMENT")
		return
	}

	if req.ReplanReason == "" {
		req.ReplanReason = "Adaptive operational revision triggered"
	}

	var uid *int64
	if user.UserID > 0 {
		val := int64(user.UserID)
		uid = &val
	}

	revisedPlan, revisedSteps, err := h.svc.Replan(r.Context(), int64(user.OrgID), uid, planID, req.ReplanReason, req.ObservedState)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "REPLAN_FAILED")
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"success":       true,
		"plan":          revisedPlan,
		"revised_plan":  revisedPlan,
		"steps":         revisedSteps,
		"revised_steps": revisedSteps,
	})
}

func (h *Handler) HandleGetAuditHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	history, err := h.svc.GetAuditHistory(r.Context(), int64(user.OrgID), planID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"history": history,
	})
}

// -----------------------------------------------------------------------------
// Policies
// -----------------------------------------------------------------------------

func (h *Handler) HandleGetPolicy(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	module := chi.URLParam(r, "module")
	p, err := h.svc.GetPolicy(r.Context(), int64(user.OrgID), module)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"policy":  p,
	})
}

func (h *Handler) HandleListPolicies(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	policies, err := h.svc.ListPolicies(r.Context(), int64(user.OrgID))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"policies": policies,
	})
}

func (h *Handler) HandleSetPolicy(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	var req SetPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_ARGUMENT")
		return
	}

	if req.Module == "" {
		utils.Error(w, http.StatusBadRequest, "module is required", "VALIDATION_ERROR")
		return
	}

	if err := h.svc.SetPolicy(r.Context(), int64(user.OrgID), int64(user.UserID), req); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Policy updated successfully",
	})
}

// -----------------------------------------------------------------------------
// Phase 5 Task 5.3: Adaptive Shipment Management Handlers
// -----------------------------------------------------------------------------

func (h *Handler) HandleIngestShipmentEvent(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	shipmentIDStr := chi.URLParam(r, "shipmentId")
	shipmentID, err := strconv.ParseInt(shipmentIDStr, 10, 64)
	if err != nil || shipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment ID", "INVALID_ARGUMENT")
		return
	}

	var req IngestShipmentEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_ARGUMENT")
		return
	}
	req.ShipmentID = shipmentID

	if req.EventType == "" {
		utils.Error(w, http.StatusBadRequest, "event_type is required", "VALIDATION_ERROR")
		return
	}

	uID := int64(user.UserID)
	result, err := h.svc.ProcessShipmentEvent(r.Context(), int64(user.OrgID), &uID, req)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"result":  result,
	})
}

func (h *Handler) HandleGetShipmentAdaptiveState(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	shipmentIDStr := chi.URLParam(r, "shipmentId")
	shipmentID, err := strconv.ParseInt(shipmentIDStr, 10, 64)
	if err != nil || shipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment ID", "INVALID_ARGUMENT")
		return
	}

	state, err := h.svc.GetShipmentAdaptiveState(r.Context(), int64(user.OrgID), shipmentID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"state":   state,
	})
}

func (h *Handler) HandleListShipmentEvents(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	shipmentIDStr := chi.URLParam(r, "shipmentId")
	shipmentID, err := strconv.ParseInt(shipmentIDStr, 10, 64)
	if err != nil || shipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment ID", "INVALID_ARGUMENT")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	events, err := h.svc.ListShipmentEvents(r.Context(), int64(user.OrgID), shipmentID, limit)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"events":  events,
	})
}

type transitionWaitingStateReq struct {
	WaitingState string  `json:"waiting_state"`
	WaitingUntil *string `json:"waiting_until,omitempty"`
}

func (h *Handler) HandleTransitionWaitingState(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	if planID == "" {
		utils.Error(w, http.StatusBadRequest, "plan ID is required", "INVALID_ARGUMENT")
		return
	}

	var req transitionWaitingStateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_ARGUMENT")
		return
	}

	var waitingUntilTime *time.Time
	if req.WaitingUntil != nil && *req.WaitingUntil != "" {
		if t, err := time.Parse(time.RFC3339, *req.WaitingUntil); err == nil {
			waitingUntilTime = &t
		}
	}

	if err := h.svc.TransitionPlanWaitingState(r.Context(), int64(user.OrgID), planID, req.WaitingState, waitingUntilTime); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Waiting state transitioned successfully",
	})
}

// -----------------------------------------------------------------------------
// Phase 5 Task 5.4: Autonomous Customer Follow-Up Handlers
// -----------------------------------------------------------------------------

func (h *Handler) HandleIngestCustomerFollowupEvent(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	custIDStr := chi.URLParam(r, "customerId")
	custID, err := strconv.ParseInt(custIDStr, 10, 64)
	if err != nil || custID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid customer ID", "INVALID_ARGUMENT")
		return
	}

	var req IngestCustomerFollowupEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_ARGUMENT")
		return
	}
	req.CustomerID = custID

	var userIDPtr *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		userIDPtr = &uid
	}

	res, err := h.svc.ProcessCustomerFollowupEvent(r.Context(), int64(user.OrgID), userIDPtr, req)
	if err != nil {
		if errors.Is(err, ErrEmergencyStopActive) {
			utils.Error(w, http.StatusForbidden, err.Error(), "EMERGENCY_STOP_ACTIVE")
			return
		}
		if errors.Is(err, ErrPlanNotFound) {
			utils.Error(w, http.StatusNotFound, "Customer not found in tenant", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"result":  res,
	})
}

func (h *Handler) HandleGetCustomerFollowupState(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	custIDStr := chi.URLParam(r, "customerId")
	custID, err := strconv.ParseInt(custIDStr, 10, 64)
	if err != nil || custID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid customer ID", "INVALID_ARGUMENT")
		return
	}

	state, err := h.svc.GetCustomerFollowupState(r.Context(), int64(user.OrgID), custID)
	if err != nil {
		if errors.Is(err, ErrPlanNotFound) {
			utils.Error(w, http.StatusNotFound, "Customer not found in tenant", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"state":   state,
	})
}

func (h *Handler) HandleUpdateCustomerPreferences(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	custIDStr := chi.URLParam(r, "customerId")
	custID, err := strconv.ParseInt(custIDStr, 10, 64)
	if err != nil || custID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid customer ID", "INVALID_ARGUMENT")
		return
	}

	var req UpdateCustomerPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid preferences payload", "INVALID_ARGUMENT")
		return
	}

	if err := h.svc.UpdateCustomerPreferences(r.Context(), int64(user.OrgID), custID, req); err != nil {
		if errors.Is(err, ErrPlanNotFound) {
			utils.Error(w, http.StatusNotFound, "Customer not found in tenant", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Preferences updated successfully",
	})
}

func (h *Handler) HandleSendCustomerFollowup(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	recIDStr := chi.URLParam(r, "recId")
	recID, err := strconv.ParseInt(recIDStr, 10, 64)
	if err != nil || recID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid record ID", "INVALID_ARGUMENT")
		return
	}

	rec, err := h.svc.SendCustomerFollowup(r.Context(), int64(user.OrgID), int64(user.UserID), recID)
	if err != nil {
		if errors.Is(err, ErrApprovalRequired) {
			utils.Error(w, http.StatusForbidden, "Communication requires supervisor approval before sending", "APPROVAL_REQUIRED")
			return
		}
		if errors.Is(err, ErrPlanNotFound) {
			utils.Error(w, http.StatusNotFound, "Followup record not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"record":  rec,
	})
}

func (h *Handler) HandleIngestCustomerResponse(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	recIDStr := chi.URLParam(r, "recId")
	recID, err := strconv.ParseInt(recIDStr, 10, 64)
	if err != nil || recID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid record ID", "INVALID_ARGUMENT")
		return
	}

	var req IngestCustomerResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ResponseText == "" {
		utils.Error(w, http.StatusBadRequest, "Response text is required", "INVALID_ARGUMENT")
		return
	}

	var userIDPtr *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		userIDPtr = &uid
	}

	result, err := h.svc.IngestCustomerResponse(r.Context(), int64(user.OrgID), userIDPtr, recID, req.ResponseText)
	if err != nil {
		if errors.Is(err, ErrPlanNotFound) {
			utils.Error(w, http.StatusNotFound, "Followup record not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"result":  result,
	})
}

// -----------------------------------------------------------------------------
// Phase 5 Task 5.5: Intelligent RFQ and Pricing Optimization Handlers
// -----------------------------------------------------------------------------

func (h *Handler) HandleEvaluateRfqPricing(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	rfqIDStr := chi.URLParam(r, "rfqId")
	rfqID, err := strconv.ParseInt(rfqIDStr, 10, 64)
	if err != nil || rfqID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid RFQ ID", "INVALID_ARGUMENT")
		return
	}

	var userIDPtr *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		userIDPtr = &uid
	}

	res, err := h.svc.EvaluateRfqPricing(r.Context(), int64(user.OrgID), userIDPtr, rfqID)
	if err != nil {
		if errors.Is(err, ErrEmergencyStopActive) {
			utils.Error(w, http.StatusConflict, err.Error(), "EMERGENCY_STOP")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"result":  res,
	})
}

func (h *Handler) HandleGetRfqPricingState(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	rfqIDStr := chi.URLParam(r, "rfqId")
	rfqID, err := strconv.ParseInt(rfqIDStr, 10, 64)
	if err != nil || rfqID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid RFQ ID", "INVALID_ARGUMENT")
		return
	}

	state, err := h.svc.GetRfqPricingState(r.Context(), int64(user.OrgID), rfqID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"state":   state,
	})
}

func (h *Handler) HandleSelectPricingStrategy(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	rfqIDStr := chi.URLParam(r, "rfqId")
	rfqID, err := strconv.ParseInt(rfqIDStr, 10, 64)
	if err != nil || rfqID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid RFQ ID", "INVALID_ARGUMENT")
		return
	}

	var req SelectPricingStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.StrategyID == "" {
		utils.Error(w, http.StatusBadRequest, "strategy_id is required", "INVALID_ARGUMENT")
		return
	}

	opt, err := h.svc.SelectPricingStrategy(r.Context(), int64(user.OrgID), int64(user.UserID), rfqID, req.StrategyID)
	if err != nil {
		if errors.Is(err, ErrCandidateNotFound) {
			utils.Error(w, http.StatusNotFound, "Candidate strategy not found", "NOT_FOUND")
			return
		}
		if errors.Is(err, ErrCandidateInfeasible) {
			utils.Error(w, http.StatusBadRequest, "Candidate strategy violates hard policy constraints and cannot be selected", "INFEASIBLE_STRATEGY")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":      true,
		"optimization": opt,
	})
}

func (h *Handler) HandleExecutePricingQuotation(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	rfqIDStr := chi.URLParam(r, "rfqId")
	rfqID, err := strconv.ParseInt(rfqIDStr, 10, 64)
	if err != nil || rfqID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid RFQ ID", "INVALID_ARGUMENT")
		return
	}

	execRes, err := h.svc.ExecutePricingQuotation(r.Context(), int64(user.OrgID), int64(user.UserID), rfqID)
	if err != nil {
		if errors.Is(err, ErrApprovalRequired) {
			utils.Error(w, http.StatusForbidden, "Quotation requires manager approval before execution", "APPROVAL_REQUIRED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"quotation": execRes,
	})
}

func (h *Handler) HandleReplanPricing(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	rfqIDStr := chi.URLParam(r, "rfqId")
	rfqID, err := strconv.ParseInt(rfqIDStr, 10, 64)
	if err != nil || rfqID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid RFQ ID", "INVALID_ARGUMENT")
		return
	}

	var req ReplanPricingRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Reason == "" {
		req.Reason = "Carrier rate adjustment or operational constraint update"
	}

	var userIDPtr *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		userIDPtr = &uid
	}

	result, err := h.svc.ReplanRfqPricing(r.Context(), int64(user.OrgID), userIDPtr, rfqID, req.Reason, req.RateDelta)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"result":  result,
	})
}

// -----------------------------------------------------------------------------
// Phase 5 Task 5.6: Adaptive Finance and Collections Handlers
// -----------------------------------------------------------------------------

func (h *Handler) HandleEvaluateFinanceCollection(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	invIDStr := chi.URLParam(r, "invoiceId")
	invID, err := strconv.ParseInt(invIDStr, 10, 64)
	if err != nil || invID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid Invoice ID", "INVALID_ARGUMENT")
		return
	}

	var userIDPtr *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		userIDPtr = &uid
	}

	state, err := h.svc.EvaluateFinanceCollection(r.Context(), int64(user.OrgID), userIDPtr, invID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(strings.ToLower(err.Error()), "no rows") {
			utils.Error(w, http.StatusNotFound, "Invoice not found in current organization", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"state":   state,
	})
}

func (h *Handler) HandleGetFinanceCollectionState(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	invIDStr := chi.URLParam(r, "invoiceId")
	invID, err := strconv.ParseInt(invIDStr, 10, 64)
	if err != nil || invID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid Invoice ID", "INVALID_ARGUMENT")
		return
	}

	state, err := h.svc.GetFinanceCollectionState(r.Context(), int64(user.OrgID), invID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"state":   state,
	})
}

func (h *Handler) HandleSelectFinanceCollectionStrategy(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	invIDStr := chi.URLParam(r, "invoiceId")
	invID, err := strconv.ParseInt(invIDStr, 10, 64)
	if err != nil || invID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid Invoice ID", "INVALID_ARGUMENT")
		return
	}

	var req SelectCollectionStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.StrategyID == "" {
		utils.Error(w, http.StatusBadRequest, "strategy_id is required", "INVALID_ARGUMENT")
		return
	}

	var userIDPtr *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		userIDPtr = &uid
	}

	state, err := h.svc.SelectFinanceCollectionStrategy(r.Context(), int64(user.OrgID), userIDPtr, invID, req.StrategyID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "INVALID_STRATEGY")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"state":   state,
	})
}

func (h *Handler) HandleExecuteFinanceCollectionAction(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	invIDStr := chi.URLParam(r, "invoiceId")
	invID, err := strconv.ParseInt(invIDStr, 10, 64)
	if err != nil || invID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid Invoice ID", "INVALID_ARGUMENT")
		return
	}

	var userIDPtr *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		userIDPtr = &uid
	}

	res, err := h.svc.ExecuteFinanceCollectionAction(r.Context(), int64(user.OrgID), userIDPtr, invID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "EXECUTION_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"result":  res,
	})
}

func (h *Handler) HandleReplanFinanceCollection(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	invIDStr := chi.URLParam(r, "invoiceId")
	invID, err := strconv.ParseInt(invIDStr, 10, 64)
	if err != nil || invID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid Invoice ID", "INVALID_ARGUMENT")
		return
	}

	var req ReplanCollectionRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.TriggerEvent == "" {
		req.TriggerEvent = "MANUAL_REPLAN"
	}

	var userIDPtr *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		userIDPtr = &uid
	}

	state, err := h.svc.ReplanFinanceCollection(r.Context(), int64(user.OrgID), userIDPtr, invID, req.TriggerEvent, req.Payload)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"state":   state,
	})
}

// -----------------------------------------------------------------------------
// Phase 5 Task 5.7: Contract and Compliance Monitoring Handlers
// -----------------------------------------------------------------------------

func (h *Handler) HandleEvaluateContractCompliance(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	cIDStr := chi.URLParam(r, "contractId")
	cID, err := strconv.ParseInt(cIDStr, 10, 64)
	if err != nil || cID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid Contract ID", "INVALID_ARGUMENT")
		return
	}

	var userIDPtr *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		userIDPtr = &uid
	}

	state, err := h.svc.EvaluateContractCompliance(r.Context(), int64(user.OrgID), userIDPtr, cID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(strings.ToLower(err.Error()), "no rows") {
			utils.Error(w, http.StatusNotFound, "Contract not found in current organization", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"state":   state,
	})
}

func (h *Handler) HandleGetContractComplianceState(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	cIDStr := chi.URLParam(r, "contractId")
	cID, err := strconv.ParseInt(cIDStr, 10, 64)
	if err != nil || cID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid Contract ID", "INVALID_ARGUMENT")
		return
	}

	state, err := h.svc.GetContractComplianceState(r.Context(), int64(user.OrgID), cID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"state":   state,
	})
}

func (h *Handler) HandleSelectContractComplianceStrategy(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	cIDStr := chi.URLParam(r, "contractId")
	cID, err := strconv.ParseInt(cIDStr, 10, 64)
	if err != nil || cID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid Contract ID", "INVALID_ARGUMENT")
		return
	}

	var req SelectComplianceStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.StrategyID == "" {
		utils.Error(w, http.StatusBadRequest, "strategy_id is required", "INVALID_ARGUMENT")
		return
	}

	var userIDPtr *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		userIDPtr = &uid
	}

	state, err := h.svc.SelectContractComplianceStrategy(r.Context(), int64(user.OrgID), userIDPtr, cID, req.StrategyID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "FAILED_SELECTING_STRATEGY")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"state":   state,
	})
}

func (h *Handler) HandleExecuteContractComplianceAction(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	cIDStr := chi.URLParam(r, "contractId")
	cID, err := strconv.ParseInt(cIDStr, 10, 64)
	if err != nil || cID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid Contract ID", "INVALID_ARGUMENT")
		return
	}

	var userIDPtr *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		userIDPtr = &uid
	}

	res, err := h.svc.ExecuteContractComplianceAction(r.Context(), int64(user.OrgID), userIDPtr, cID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "EXECUTION_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"result":  res,
	})
}

func (h *Handler) HandleReplanContractCompliance(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	cIDStr := chi.URLParam(r, "contractId")
	cID, err := strconv.ParseInt(cIDStr, 10, 64)
	if err != nil || cID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid Contract ID", "INVALID_ARGUMENT")
		return
	}

	var req ReplanComplianceRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.TriggerEvent == "" {
		req.TriggerEvent = "MANUAL_REPLAN"
	}

	var userIDPtr *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		userIDPtr = &uid
	}

	state, err := h.svc.ReplanContractCompliance(r.Context(), int64(user.OrgID), userIDPtr, cID, req.TriggerEvent, req.Payload)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"state":   state,
	})
}

// -------------------------------------------------------------------------
// Phase 5 Task 5.8: Autonomous Exception Resolution HTTP Handlers
// -------------------------------------------------------------------------

func (h *Handler) HandleEvaluateExceptionResolution(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	excIDStr := chi.URLParam(r, "exceptionId")
	excID, err := strconv.ParseInt(excIDStr, 10, 64)
	if err != nil || excID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid Exception ID", "INVALID_ARGUMENT")
		return
	}

	var userIDPtr *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		userIDPtr = &uid
	}

	state, err := h.svc.EvaluateExceptionResolution(r.Context(), int64(user.OrgID), userIDPtr, excID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(strings.ToLower(err.Error()), "no rows") || strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, "Exception not found in current organization", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"state":   state,
	})
}

func (h *Handler) HandleGetExceptionResolutionState(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	excIDStr := chi.URLParam(r, "exceptionId")
	excID, err := strconv.ParseInt(excIDStr, 10, 64)
	if err != nil || excID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid Exception ID", "INVALID_ARGUMENT")
		return
	}

	state, err := h.svc.GetExceptionResolutionState(r.Context(), int64(user.OrgID), excID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"state":   state,
	})
}

func (h *Handler) HandleSelectExceptionResolutionStrategy(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	excIDStr := chi.URLParam(r, "exceptionId")
	excID, err := strconv.ParseInt(excIDStr, 10, 64)
	if err != nil || excID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid Exception ID", "INVALID_ARGUMENT")
		return
	}

	var req SelectExceptionStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.StrategyID == "" {
		utils.Error(w, http.StatusBadRequest, "Missing or invalid strategy_id in request body", "INVALID_ARGUMENT")
		return
	}

	var userIDPtr *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		userIDPtr = &uid
	}

	state, err := h.svc.SelectExceptionResolutionStrategy(r.Context(), int64(user.OrgID), userIDPtr, excID, req.StrategyID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "FAILED_SELECTING_STRATEGY")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"state":   state,
	})
}

func (h *Handler) HandleExecuteExceptionResolutionAction(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	excIDStr := chi.URLParam(r, "exceptionId")
	excID, err := strconv.ParseInt(excIDStr, 10, 64)
	if err != nil || excID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid Exception ID", "INVALID_ARGUMENT")
		return
	}

	var req ExecuteExceptionActionRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	var userIDPtr *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		userIDPtr = &uid
	}

	res, err := h.svc.ExecuteExceptionResolutionAction(r.Context(), int64(user.OrgID), userIDPtr, excID, &req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "EXECUTION_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"result":  res,
	})
}

func (h *Handler) HandleReplanExceptionResolution(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	excIDStr := chi.URLParam(r, "exceptionId")
	excID, err := strconv.ParseInt(excIDStr, 10, 64)
	if err != nil || excID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid Exception ID", "INVALID_ARGUMENT")
		return
	}

	var req ReplanExceptionRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.TriggerEvent == "" {
		req.TriggerEvent = "MANUAL_REPLAN"
	}

	var userIDPtr *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		userIDPtr = &uid
	}

	state, err := h.svc.ReplanExceptionResolution(r.Context(), int64(user.OrgID), userIDPtr, excID, req.TriggerEvent, req.Payload)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"state":   state,
	})
}

// -----------------------------------------------------------------------------
// Phase 5 Task 5.9: Multi-Step Planning and Execution Handlers
// -----------------------------------------------------------------------------

func (h *Handler) HandleGenerateCrossModulePlan(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req CrossModulePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_ARGUMENT")
		return
	}

	var userIDPtr *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		userIDPtr = &uid
	}

	plan, steps, err := h.svc.GenerateCrossModulePlan(r.Context(), int64(user.OrgID), userIDPtr, req)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PLAN_GENERATION_FAILED")
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"plan":    plan,
		"steps":   steps,
	})
}

func (h *Handler) HandleGetPlanningMetrics(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	metrics, err := h.svc.GetPlanningMetrics(r.Context(), int64(user.OrgID))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"metrics": metrics,
	})
}

func (h *Handler) HandleListEntityConflicts(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")
	if entityType == "" || entityID == "" {
		utils.Error(w, http.StatusBadRequest, "entity_type and entity_id are required", "INVALID_ARGUMENT")
		return
	}

	conflicts, err := h.svc.ListEntityConflicts(r.Context(), int64(user.OrgID), entityType, entityID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"conflicts": conflicts,
	})
}

func (h *Handler) HandleExecuteNextStep(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	if planID == "" {
		utils.Error(w, http.StatusBadRequest, "Plan ID is required", "INVALID_ARGUMENT")
		return
	}

	result, err := h.svc.ExecuteNextStep(r.Context(), int64(user.OrgID), int64(user.UserID), planID)
	if err != nil {
		if errors.Is(err, ErrApprovalRequired) {
			utils.Error(w, http.StatusForbidden, err.Error(), "APPROVAL_REQUIRED")
			return
		}
		if errors.Is(err, ErrEmergencyStopActive) {
			utils.Error(w, http.StatusForbidden, err.Error(), "EMERGENCY_STOP")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXECUTION_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"result":  result,
	})
}

func (h *Handler) HandleApproveStep(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	stepID := chi.URLParam(r, "stepId")
	if planID == "" || stepID == "" {
		utils.Error(w, http.StatusBadRequest, "Plan ID and Step ID are required", "INVALID_ARGUMENT")
		return
	}

	resp, err := h.svc.ApproveStep(r.Context(), int64(user.OrgID), int64(user.UserID), planID, stepID)
	if err != nil {
		if errors.Is(err, ErrStepNotFound) {
			utils.Error(w, http.StatusNotFound, "Step not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"approval": resp,
	})
}

func (h *Handler) HandleRetryStep(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	stepID := chi.URLParam(r, "stepId")
	if planID == "" || stepID == "" {
		utils.Error(w, http.StatusBadRequest, "Plan ID and Step ID are required", "INVALID_ARGUMENT")
		return
	}

	step, err := h.svc.RetryStep(r.Context(), int64(user.OrgID), int64(user.UserID), planID, stepID)
	if err != nil {
		if errors.Is(err, ErrMaxAttemptsExceeded) {
			utils.Error(w, http.StatusConflict, err.Error(), "MAX_ATTEMPTS_EXCEEDED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"step":    step,
	})
}

func (h *Handler) HandleCompensateStep(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	stepID := chi.URLParam(r, "stepId")
	if planID == "" || stepID == "" {
		utils.Error(w, http.StatusBadRequest, "Plan ID and Step ID are required", "INVALID_ARGUMENT")
		return
	}

	step, err := h.svc.CompensateStep(r.Context(), int64(user.OrgID), int64(user.UserID), planID, stepID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"step":    step,
	})
}

func (h *Handler) HandleValidatePlan(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	if planID == "" {
		utils.Error(w, http.StatusBadRequest, "Plan ID is required", "INVALID_ARGUMENT")
		return
	}

	valResp, err := h.svc.ValidatePlan(r.Context(), int64(user.OrgID), planID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "VALIDATION_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"validation": valResp,
	})
}

func (h *Handler) HandleCheckPlanConflicts(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	if planID == "" {
		utils.Error(w, http.StatusBadRequest, "Plan ID is required", "INVALID_ARGUMENT")
		return
	}

	confResp, err := h.svc.CheckPlanConflicts(r.Context(), int64(user.OrgID), planID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"conflicts": confResp,
	})
}

// ==============================================================================
// Phase 5 Task 5.10: Continuous Monitoring and Replanning Handlers
// ==============================================================================

func (h *Handler) HandleIngestMonitoringEvent(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req IngestMonitoringEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	if req.EventID == "" || req.EventType == "" {
		utils.Error(w, http.StatusBadRequest, "event_id and event_type are required", "INVALID_ARGUMENT")
		return
	}

	resp, err := h.svc.IngestAndEvaluateEvent(r.Context(), int64(user.OrgID), req)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INGESTION_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"response": resp,
	})
}

func (h *Handler) HandleListMonitoringEvents(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	events, err := h.svc.ListMonitoringEvents(r.Context(), int64(user.OrgID), limit)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"events":  events,
		"count":   len(events),
	})
}

func (h *Handler) HandleGetPlanHealth(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	if planID == "" {
		utils.Error(w, http.StatusBadRequest, "Plan ID is required", "INVALID_ARGUMENT")
		return
	}

	healthData, err := h.svc.GetPlanHealth(r.Context(), int64(user.OrgID), planID)
	if err != nil {
		if errors.Is(err, ErrPlanNotFound) {
			utils.Error(w, http.StatusNotFound, "Plan not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"health":  healthData,
	})
}

func (h *Handler) HandleTriggerAdaptiveReplan(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	if planID == "" {
		utils.Error(w, http.StatusBadRequest, "Plan ID is required", "INVALID_ARGUMENT")
		return
	}

	var reqBody struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&reqBody)
	reason := reqBody.Reason
	if reason == "" {
		reason = "Operator manually requested adaptive replanning"
	}

	plan, steps, err := h.svc.TriggerAdaptiveReplan(r.Context(), int64(user.OrgID), int64(user.UserID), planID, reason)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "REPLAN_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"plan":    plan,
		"steps":   steps,
	})
}

func (h *Handler) HandleGetContinuousMonitoringMetrics(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	metrics, err := h.svc.GetContinuousMonitoringMetrics(r.Context(), int64(user.OrgID))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "METRICS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"metrics": metrics,
	})
}

// -----------------------------------------------------------------------------
// Phase 5 Task 5.11: Human + AI Operating Model Handlers
// -----------------------------------------------------------------------------

func (h *Handler) HandleCreateDecisionPoint(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req CreateDecisionPointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_ARGUMENT")
		return
	}

	dec, err := h.svc.CreateDecisionPoint(r.Context(), int64(user.OrgID), &req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "DECISION_CREATE_ERROR")
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"success":  true,
		"decision": dec,
	})
}

func (h *Handler) HandleListDecisionPoints(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	status := r.URL.Query().Get("status")
	module := r.URL.Query().Get("module")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	decisions, total, err := h.svc.ListDecisionPoints(r.Context(), int64(user.OrgID), status, module, limit, offset)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "DECISION_LIST_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"decisions": decisions,
		"total":     total,
	})
}

func (h *Handler) HandleGetDecisionPoint(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	decisionID := chi.URLParam(r, "decisionId")
	if decisionID == "" {
		utils.Error(w, http.StatusBadRequest, "Decision ID is required", "INVALID_ARGUMENT")
		return
	}

	dec, err := h.svc.GetDecisionPoint(r.Context(), int64(user.OrgID), decisionID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "Decision not found", "NOT_FOUND")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"decision": dec,
	})
}

func (h *Handler) HandleSubmitHumanDecision(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	decisionID := chi.URLParam(r, "decisionId")
	if decisionID == "" {
		utils.Error(w, http.StatusBadRequest, "Decision ID is required", "INVALID_ARGUMENT")
		return
	}

	var req SubmitHumanDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_ARGUMENT")
		return
	}

	userName := fmt.Sprintf("User-%d (%s)", user.UserID, user.Role)
	dec, err := h.svc.SubmitHumanDecision(r.Context(), int64(user.OrgID), decisionID, int64(user.UserID), userName, &req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "DECISION_SUBMIT_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"decision": dec,
	})
}

func (h *Handler) HandleStopWorkflow(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	if planID == "" {
		utils.Error(w, http.StatusBadRequest, "Plan ID is required", "INVALID_ARGUMENT")
		return
	}

	var reqBody struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&reqBody)
	reason := reqBody.Reason
	if reason == "" {
		reason = "Human operator triggered emergency stop control"
	}

	if err := h.svc.StopWorkflow(r.Context(), int64(user.OrgID), planID, int64(user.UserID), reason); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "STOP_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Workflow successfully stopped and pending steps cancelled",
	})
}

func (h *Handler) HandleUpdateStepHumanEdit(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "id")
	stepID := chi.URLParam(r, "stepId")
	if planID == "" || stepID == "" {
		utils.Error(w, http.StatusBadRequest, "Plan ID and Step ID are required", "INVALID_ARGUMENT")
		return
	}

	var reqBody struct {
		HumanContent string `json:"human_content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_ARGUMENT")
		return
	}

	if err := h.svc.UpdateStepHumanEdit(r.Context(), int64(user.OrgID), planID, stepID, reqBody.HumanContent); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "STEP_EDIT_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Step human content updated and plan marked as human modified",
	})
}

func (h *Handler) HandleInvalidateApprovals(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var reqBody struct {
		Module     string `json:"module"`
		EntityType string `json:"entity_type"`
		EntityID   string `json:"entity_id"`
		Reason     string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_ARGUMENT")
		return
	}

	count, err := h.svc.InvalidatePendingApprovalsOnMaterialChange(r.Context(), int64(user.OrgID), reqBody.Module, reqBody.EntityType, reqBody.EntityID, reqBody.Reason)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "INVALIDATE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":           true,
		"invalidated_count": count,
	})
}

func (h *Handler) HandleGetDecisionCenterSummary(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	summary, err := h.svc.GetHumanDecisionCenterSummary(r.Context(), int64(user.OrgID), int64(user.UserID))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "SUMMARY_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"summary": summary,
	})
}

// Phase 5 Task 5.12: Autonomous Operations Command Center HTTP Handlers

func (h *Handler) HandleGetCommandCenterOverview(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	overview, err := h.svc.GetCommandCenterOverview(r.Context(), int64(user.OrgID))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "OVERVIEW_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"overview": overview,
	})
}

func (h *Handler) HandleGetCommandCenterCriticalAttention(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	items, err := h.svc.GetCommandCenterCriticalAttention(r.Context(), int64(user.OrgID), limit)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "CRITICAL_ATTENTION_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"items":   items,
		"count":   len(items),
	})
}

func (h *Handler) HandleGetCommandCenterWorkflows(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	module := r.URL.Query().Get("module")
	status := r.URL.Query().Get("status")
	autonomyLevel := r.URL.Query().Get("autonomy_level")
	search := r.URL.Query().Get("search")

	limit := 20
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	plans, total, err := h.svc.GetCommandCenterWorkflows(r.Context(), int64(user.OrgID), module, status, autonomyLevel, search, limit, offset)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "WORKFLOWS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"plans":   plans,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

func (h *Handler) HandleGetCommandCenterDecisions(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	limit := 20
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	decisions, total, err := h.svc.GetCommandCenterDecisions(r.Context(), int64(user.OrgID), limit, offset)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "DECISIONS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"decisions": decisions,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

func (h *Handler) HandleGetCommandCenterRisks(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	domains, err := h.svc.GetCommandCenterDomainRisks(r.Context(), int64(user.OrgID))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "RISKS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"domains": domains,
	})
}

func (h *Handler) HandleGetCommandCenterActivity(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	limit := 20
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	activity, err := h.svc.GetCommandCenterActivity(r.Context(), int64(user.OrgID), limit)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "ACTIVITY_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"activity": activity,
	})
}

func (h *Handler) HandleGetCommandCenterSystemHealth(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}
	_ = user

	health, err := h.svc.GetCommandCenterSystemHealth(r.Context())
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "HEALTH_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"health":  health,
	})
}

// ============================================================================
// Phase 5 Task 5.13: Agent Memory and Learning from Outcomes Handlers
// ============================================================================

func (h *Handler) HandleRecordOutcome(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req RecordOutcomeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_REQUEST")
		return
	}

	var uID *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		uID = &uid
	}

	outcome, createdMem, err := h.svc.CaptureOutcome(r.Context(), int64(user.OrgID), uID, &req)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "RECORD_OUTCOME_FAILED")
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"success":        true,
		"outcome":        outcome,
		"created_memory": createdMem,
	})
}

func (h *Handler) HandleListOutcomes(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	entityType := q.Get("entity_type")
	entityID := q.Get("entity_id")
	outcomeType := q.Get("outcome_type")
	status := q.Get("status")

	limit := 50
	if lStr := q.Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	offset := 0
	if oStr := q.Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	outcomes, total, err := h.svc.ListOutcomes(r.Context(), int64(user.OrgID), entityType, entityID, outcomeType, status, limit, offset)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "LIST_OUTCOMES_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"outcomes": outcomes,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h *Handler) HandleGetOutcome(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	outcomeID := chi.URLParam(r, "outcomeId")
	if outcomeID == "" {
		utils.Error(w, http.StatusBadRequest, "Missing outcomeId parameter", "INVALID_PARAM")
		return
	}

	outcome, err := h.svc.GetOutcome(r.Context(), int64(user.OrgID), outcomeID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error(), "OUTCOME_NOT_FOUND")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"outcome": outcome,
	})
}

func (h *Handler) HandleVerifyOutcome(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	outcomeID := chi.URLParam(r, "outcomeId")
	if outcomeID == "" {
		utils.Error(w, http.StatusBadRequest, "Missing outcomeId parameter", "INVALID_PARAM")
		return
	}

	var req VerifyOutcomeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_REQUEST")
		return
	}

	var uID *int64
	if user.UserID > 0 {
		uid := int64(user.UserID)
		uID = &uid
	}

	outcome, createdMem, err := h.svc.VerifyOutcome(r.Context(), int64(user.OrgID), outcomeID, uID, &req)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "VERIFY_OUTCOME_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"outcome":        outcome,
		"created_memory": createdMem,
	})
}

func (h *Handler) HandleRetrieveMemory(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req struct {
		QueryContext string `json:"query_context"`
		Module       string `json:"module"`
		Category     string `json:"category"`
		EntityType   string `json:"entity_type"`
		EntityID     string `json:"entity_id"`
		Limit        int    `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_REQUEST")
		return
	}

	resp, err := h.svc.RetrieveContextualMemory(r.Context(), int64(user.OrgID), req.QueryContext, req.Module, req.Category, req.EntityType, req.EntityID, req.Limit)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "RETRIEVE_MEMORY_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"retrieval": resp,
	})
}

func (h *Handler) HandleListMemories(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	category := q.Get("category")
	if category == "undefined" || category == "null" || category == "ALL" {
		category = ""
	}
	entityType := q.Get("entity_type")
	if entityType == "undefined" || entityType == "null" {
		entityType = ""
	}
	entityID := q.Get("entity_id")
	if entityID == "undefined" || entityID == "null" {
		entityID = ""
	}
	scope := q.Get("scope")
	if scope == "undefined" || scope == "null" || scope == "ALL" {
		scope = ""
	}
	includeStale := q.Get("include_stale") == "true"

	limit := 50
	if lStr := q.Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	offset := 0
	if oStr := q.Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	items, total, err := h.svc.ListMemories(r.Context(), int64(user.OrgID), category, entityType, entityID, scope, includeStale, limit, offset)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "LIST_MEMORIES_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"memories": items,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h *Handler) HandleGetMemory(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	memIDStr := chi.URLParam(r, "memoryId")
	memID, err := strconv.ParseInt(memIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid memoryId", "INVALID_PARAM")
		return
	}

	mem, err := h.svc.GetMemory(r.Context(), int64(user.OrgID), memID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error(), "MEMORY_NOT_FOUND")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"memory":  mem,
	})
}

func (h *Handler) HandleCorrectMemory(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	memIDStr := chi.URLParam(r, "memoryId")
	memID, err := strconv.ParseInt(memIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid memoryId", "INVALID_PARAM")
		return
	}

	var req CorrectMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_REQUEST")
		return
	}

	updated, err := h.svc.CorrectMemory(r.Context(), int64(user.OrgID), memID, int64(user.UserID), &req)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "CORRECT_MEMORY_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"memory":  updated,
	})
}

func (h *Handler) HandleInvalidateMemory(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	memIDStr := chi.URLParam(r, "memoryId")
	memID, err := strconv.ParseInt(memIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid memoryId", "INVALID_PARAM")
		return
	}

	var req InvalidateMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_REQUEST")
		return
	}

	if err := h.svc.InvalidateMemory(r.Context(), int64(user.OrgID), memID, int64(user.UserID), &req); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INVALIDATE_MEMORY_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Memory item invalidated successfully",
	})
}

func (h *Handler) HandleFlagMemoryUnreliable(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	memIDStr := chi.URLParam(r, "memoryId")
	memID, err := strconv.ParseInt(memIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid memoryId", "INVALID_PARAM")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.svc.FlagMemoryUnreliable(r.Context(), int64(user.OrgID), memID, int64(user.UserID), req.Reason); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "FLAG_UNRELIABLE_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Memory item flagged as unreliable",
	})
}

func (h *Handler) HandleDetectPatterns(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	resp, err := h.svc.DetectAndSyncPatterns(r.Context(), int64(user.OrgID))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "DETECT_PATTERNS_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"patterns": resp,
	})
}

func (h *Handler) HandleListPatterns(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	patternType := q.Get("pattern_type")
	entityType := q.Get("entity_type")

	limit := 50
	if lStr := q.Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	offset := 0
	if oStr := q.Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	patterns, total, err := h.svc.ListLearnedPatterns(r.Context(), int64(user.OrgID), patternType, entityType, limit, offset)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "LIST_PATTERNS_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"patterns": patterns,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h *Handler) HandleGetMemoryLearningSummary(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	summary, err := h.svc.GetMemoryLearningSummary(r.Context(), int64(user.OrgID))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "SUMMARY_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"summary": summary,
	})
}

// -----------------------------------------------------------------------------
// Phase 5 Task 5.14: Governance for Controlled Autonomy Handlers
// -----------------------------------------------------------------------------

func (h *Handler) HandleEvaluateGovernance(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req GovernanceEvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "BAD_REQUEST")
		return
	}

	req.OrgID = int64(user.OrgID)
	if req.UserID == 0 {
		req.UserID = int64(user.UserID)
	}

	res, err := h.svc.EvaluateGovernance(r.Context(), req)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "GOVERNANCE_EVAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"result":  res,
	})
}

func (h *Handler) HandleGetTenantLimits(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	limits, err := h.svc.GetTenantLimits(r.Context(), int64(user.OrgID))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "GET_LIMITS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"limits":  limits,
	})
}

func (h *Handler) HandleUpdateTenantLimits(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var limits TenantGovernanceLimits
	if err := json.NewDecoder(r.Body).Decode(&limits); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "BAD_REQUEST")
		return
	}

	limits.OrgID = int64(user.OrgID)
	if err := h.svc.UpdateTenantLimits(r.Context(), int64(user.OrgID), &limits); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "UPDATE_LIMITS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"limits":  limits,
	})
}

func (h *Handler) HandleToggleKillSwitch(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var body struct {
		Enabled bool   `json:"enabled"`
		Reason  string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "BAD_REQUEST")
		return
	}

	err := h.svc.ToggleKillSwitch(r.Context(), int64(user.OrgID), body.Enabled, body.Reason, int64(user.UserID))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "KILL_SWITCH_ERROR")
		return
	}

	limits, _ := h.svc.GetTenantLimits(r.Context(), int64(user.OrgID))

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"limits":  limits,
	})
}

func (h *Handler) HandleGetActionAllowlist(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	module := q.Get("module")

	items, err := h.svc.GetActionAllowlist(r.Context(), int64(user.OrgID), module)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "ALLOWLIST_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"items":   items,
		"count":   len(items),
	})
}

func (h *Handler) HandleUpsertActionAllowlistItem(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var item ActionAllowlistItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "BAD_REQUEST")
		return
	}
	item.OrgID = int64(user.OrgID)

	if err := h.svc.SaveActionAllowlistItem(r.Context(), &item); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "UPSERT_ALLOWLIST_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Allowlist item saved successfully",
	})
}

func (h *Handler) HandleGetFeatureFlags(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	flags, err := h.svc.GetFeatureFlags(r.Context(), int64(user.OrgID))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "FLAGS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"flags":   flags,
		"count":   len(flags),
	})
}

func (h *Handler) HandleUpdateFeatureFlag(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	flagKey := chi.URLParam(r, "flagKey")
	if flagKey == "" {
		utils.Error(w, http.StatusBadRequest, "Flag key is required", "BAD_REQUEST")
		return
	}

	var body struct {
		Enabled          bool `json:"enabled"`
		MaxAutonomyLevel int  `json:"max_autonomy_level"`
		RequiresApproval bool `json:"requires_approval"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "BAD_REQUEST")
		return
	}

	err := h.svc.UpdateFeatureFlag(r.Context(), int64(user.OrgID), flagKey, body.Enabled, body.MaxAutonomyLevel, body.RequiresApproval, int64(user.UserID))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "UPDATE_FLAG_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Feature flag updated successfully",
	})
}

func (h *Handler) HandleListPolicyEvaluations(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	limit := 50
	if lStr := q.Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	offset := 0
	if oStr := q.Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	records, total, err := h.svc.ListPolicyEvaluations(r.Context(), int64(user.OrgID), limit, offset)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "LIST_EVALS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"evaluations": records,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

func (h *Handler) HandleListPolicyAuditLogs(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	limit := 50
	if lStr := q.Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	offset := 0
	if oStr := q.Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	logs, total, err := h.svc.ListPolicyAuditLogs(r.Context(), int64(user.OrgID), limit, offset)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "LIST_AUDIT_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"logs":    logs,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

func (h *Handler) HandleGetGovernanceTelemetry(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	telemetry, err := h.svc.GetGovernanceTelemetry(r.Context(), int64(user.OrgID))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "TELEMETRY_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"telemetry": telemetry,
	})
}

func (h *Handler) HandlePreviewPlan(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req PlanPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "BAD_REQUEST")
		return
	}
	req.OrgID = int64(user.OrgID)

	preview, err := h.svc.PreviewPlan(r.Context(), req)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREVIEW_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"preview": preview,
	})
}
