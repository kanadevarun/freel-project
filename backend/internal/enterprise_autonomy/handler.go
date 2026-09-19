package enterprise_autonomy

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
)

type Handler struct {
	svc                      Service
	shipmentLifecycleSvc     ShipmentLifecycleService
	commercialLifecycleSvc   CommercialLifecycleService
	exceptionManagementSvc   ExceptionManagementService
	customerRelationshipSvc  CustomerRelationshipService
	revenueOptimizationSvc   RevenueOptimizationService
	contractComplianceRiskSvc ContractComplianceRiskService
	eventMeshSvc             EnterpriseEventMeshService
	controlTowerSvc          EnterpriseControlTowerService
	governanceSvc            EnterpriseGovernanceService
	resilienceSvc            EnterpriseResilienceService
}

func NewHandler(
	svc Service,
	shipLifeSvc ShipmentLifecycleService,
	commLifeSvc CommercialLifecycleService,
	excMgmtSvc ExceptionManagementService,
	crmSvc CustomerRelationshipService,
	revOptSvc RevenueOptimizationService,
	riskSvc ContractComplianceRiskService,
	eventMeshSvc EnterpriseEventMeshService,
	controlTowerSvc EnterpriseControlTowerService,
	govSvc EnterpriseGovernanceService,
	resilienceSvc EnterpriseResilienceService,
) *Handler {
	return &Handler{
		svc:                      svc,
		shipmentLifecycleSvc:     shipLifeSvc,
		commercialLifecycleSvc:   commLifeSvc,
		exceptionManagementSvc:   excMgmtSvc,
		customerRelationshipSvc:  crmSvc,
		revenueOptimizationSvc:   revOptSvc,
		contractComplianceRiskSvc: riskSvc,
		eventMeshSvc:             eventMeshSvc,
		controlTowerSvc:          controlTowerSvc,
		governanceSvc:            govSvc,
		resilienceSvc:            resilienceSvc,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/workflows", h.StartWorkflow)
	r.Get("/workflows", h.ListWorkflows)
	r.Get("/workflows/{workflow_id}", h.GetWorkflow)
	r.Post("/workflows/{workflow_id}/pause", h.PauseWorkflow)
	r.Post("/workflows/{workflow_id}/resume", h.ResumeWorkflow)
	r.Post("/workflows/{workflow_id}/cancel", h.CancelWorkflow)
	r.Post("/workflows/{workflow_id}/steps/{step_id}/approve", h.ApproveWorkflowStep)
	r.Post("/workflows/{workflow_id}/steps/{step_id}/reject", h.RejectWorkflowStep)
	r.Post("/workflows/recover", h.RecoverInterruptedWorkflows)
	r.Post("/events/trigger", h.TriggerBusinessEvent)
	r.Post("/emergency-control", h.EmergencyControl)
	r.Get("/overview", h.GetPlatformOverview)

	// Phase 7.2 Autonomous Shipment Lifecycle routes
	r.Post("/shipments/{shipment_id}/lifecycle/initiate", h.InitiateShipmentLifecycle)
	r.Get("/shipments/{shipment_id}/lifecycle", h.GetShipmentLifecycle)
	r.Post("/shipments/{shipment_id}/lifecycle/event", h.ProcessShipmentLifecycleEvent)
	r.Post("/shipments/{shipment_id}/lifecycle/eta-evaluate", h.EvaluateShipmentETA)
	r.Post("/shipments/{shipment_id}/lifecycle/replan", h.ReplanShipmentLifecycle)
	r.Post("/shipments/{shipment_id}/lifecycle/deliver", h.DeliverShipmentLifecycle)
	r.Post("/shipments/{shipment_id}/lifecycle/post-delivery-audit", h.AuditPostDelivery)
	r.Post("/shipments/{shipment_id}/lifecycle/outcome", h.RecordShipmentOutcome)

	// Phase 7.3 Autonomous Quote-to-Cash Commercial Lifecycle routes
	r.Post("/commercial/rfqs/{rfq_id}/initiate", h.InitiateCommercialLifecycle)
	r.Get("/commercial/{workflow_id}", h.GetCommercialLifecycle)
	r.Post("/commercial/event", h.ProcessCommercialLifecycleEvent)
	r.Post("/commercial/{workflow_id}/extract", h.ExtractCommercialRFQ)
	r.Post("/commercial/{workflow_id}/qualify", h.QualifyCommercialRFQ)
	r.Post("/commercial/{workflow_id}/optimize-pricing", h.OptimizeCommercialPricing)
	r.Post("/commercial/{workflow_id}/check-compliance", h.CheckCommercialCompliance)
	r.Post("/commercial/{workflow_id}/prepare-quote", h.PrepareCommercialQuote)
	r.Post("/commercial/{workflow_id}/negotiate", h.NegotiateCommercialQuote)
	r.Post("/commercial/{workflow_id}/apply-negotiation", h.ApplyCommercialNegotiation)
	r.Post("/commercial/{workflow_id}/accept", h.AcceptCommercialQuote)
	r.Post("/commercial/{workflow_id}/handoff-booking", h.HandoffCommercialBooking)
	r.Post("/commercial/{workflow_id}/handoff-shipment", h.HandoffCommercialShipment)
	r.Post("/commercial/{workflow_id}/invoice", h.EvaluateCommercialInvoice)
	r.Post("/commercial/{workflow_id}/collection", h.AssessCommercialCollection)
	r.Post("/commercial/{workflow_id}/outcome", h.RecordCommercialOutcome)

	// Phase 7.4 Autonomous Enterprise Exception Management routes
	r.Post("/exceptions/detect", h.DetectException)
	r.Get("/exceptions/{workflow_id}", h.GetExceptionWorkflow)
	r.Post("/exceptions/{workflow_id}/investigate", h.InvestigateException)
	r.Post("/exceptions/{workflow_id}/impact", h.AssessExceptionImpact)
	r.Post("/exceptions/{workflow_id}/plan", h.PlanExceptionRecovery)
	r.Post("/exceptions/{workflow_id}/options/{option_id}/select", h.SelectExceptionRecoveryOption)
	r.Post("/exceptions/{workflow_id}/execute", h.ExecuteExceptionRecoveryStep)
	r.Post("/exceptions/{workflow_id}/verify", h.VerifyExceptionRecovery)
	r.Post("/exceptions/{workflow_id}/replan", h.TriggerAdaptiveRecovery)
	r.Post("/exceptions/{workflow_id}/monitor", h.TransitionExceptionToMonitoring)
	r.Post("/exceptions/{workflow_id}/resolve", h.ResolveException)
	r.Post("/exceptions/{workflow_id}/escalate", h.EscalateException)
	r.Post("/exceptions/{workflow_id}/outcome", h.RecordExceptionOutcome)

	// Phase 7.5 Autonomous Customer Relationship Management routes
	r.Post("/crm/customers/{customer_id}/initiate", h.InitiateCustomerRelationshipWorkflow)
	r.Get("/crm/{workflow_id}", h.GetCustomerRelationshipWorkflow)
	r.Post("/crm/{workflow_id}/health", h.EvaluateCustomerRelationshipHealth)
	r.Post("/crm/{workflow_id}/detect", h.DetectCustomerRisksAndOpportunities)
	r.Post("/crm/{workflow_id}/investigate", h.InvestigateCustomerRootCauses)
	r.Post("/crm/{workflow_id}/plan", h.PlanCustomerInterventions)
	r.Post("/crm/{workflow_id}/options/{option_id}/select", h.SelectCustomerInterventionOption)
	r.Post("/crm/{workflow_id}/execute", h.ExecuteCustomerIntervention)
	r.Post("/crm/{workflow_id}/verify", h.VerifyCustomerIntervention)
	r.Post("/crm/{workflow_id}/response", h.ProcessCustomerResponse)
	r.Post("/crm/event", h.ProcessCustomerLifecycleEvent)
	r.Post("/crm/{workflow_id}/outcome", h.RecordCustomerOutcome)

	// Phase 7.6 Autonomous Revenue and Margin Optimization routes
	r.Post("/revenue/initiate", h.InitiateRevenueWorkflow)
	r.Get("/revenue/{workflow_id}", h.GetRevenueWorkflow)
	r.Post("/revenue/{workflow_id}/cost-margin", h.EvaluateCostAndMargin)
	r.Post("/revenue/{workflow_id}/carrier-economics", h.EvaluateCarrierEconomics)
	r.Post("/revenue/{workflow_id}/customer-value", h.AssessRevenueCustomerValue)
	r.Post("/revenue/{workflow_id}/optimize", h.ConductMultiAgentOptimization)
	r.Post("/revenue/{workflow_id}/recommend", h.FormulatePricingRecommendation)
	r.Post("/revenue/{workflow_id}/options/{option_id}/select", h.SelectRevenueOptimizationOption)
	r.Post("/revenue/{workflow_id}/execute", h.ExecuteRevenueCommercialAction)
	r.Post("/revenue/{workflow_id}/verify", h.VerifyRevenueCommercialAction)
	r.Post("/revenue/{workflow_id}/negotiate", h.OptimizeNegotiationRequest)
	r.Post("/revenue/event", h.ProcessRevenueLifecycleEvent)
	r.Post("/revenue/{workflow_id}/outcome", h.RecordRevenueOutcome)

	// Phase 7.7 Autonomous Contract, Compliance and Risk Governance routes
	r.Post("/risk/initiate", h.InitiateRiskWorkflow)
	r.Get("/risk/{workflow_id}", h.GetRiskWorkflow)
	r.Post("/risk/{workflow_id}/evidence", h.CollectRiskEvidence)
	r.Post("/risk/{workflow_id}/assess", h.ConductMultiAgentRiskAssessment)
	r.Post("/risk/{workflow_id}/impact", h.AssessCrossDomainImpact)
	r.Post("/risk/{workflow_id}/plan", h.PlanRiskMitigationOptions)
	r.Post("/risk/{workflow_id}/options/{option_id}/select", h.SelectRiskMitigationOption)
	r.Post("/risk/{workflow_id}/execute", h.ExecuteRiskMitigationAction)
	r.Post("/risk/{workflow_id}/verify", h.VerifyRiskMitigationAction)
	r.Post("/risk/{workflow_id}/monitor", h.TransitionRiskToMonitoring)
	r.Post("/risk/{workflow_id}/resolve", h.ResolveRiskWorkflow)
	r.Post("/risk/{workflow_id}/reassess", h.ReassessRiskCondition)
	r.Post("/risk/event", h.ProcessRiskLifecycleEvent)
	r.Post("/risk/{workflow_id}/outcome", h.RecordRiskOutcome)

	// Phase 7.8 Enterprise Event Mesh and Autonomous Workflow Engine routes
	r.Post("/mesh/events", h.IngestEventMeshEvent)
	r.Get("/mesh/events", h.ListEventMeshEvents)
	r.Get("/mesh/events/{event_id}", h.GetEventMeshEvent)
	r.Get("/mesh/dead-letters", h.GetEventMeshDeadLetters)
	r.Post("/mesh/dead-letters/{event_id}/replay", h.ReplayEventMeshDeadLetter)
	r.Get("/mesh/overview", h.GetEventMeshOverview)
	r.Get("/mesh/rules", h.ListEventMeshRoutingRules)

	// Phase 7.9 Enterprise Autonomous Control Tower routes
	r.Get("/control-tower/view", h.GetControlTowerView)
	r.Get("/control-tower/workflows/{workflow_id}/trace", h.GetControlTowerWorkflowTrace)
	r.Post("/control-tower/workflows/{workflow_id}/control", h.PerformControlTowerAction)

	// Phase 7.10 Enterprise Autonomy, Governance & Safety routes
	r.Post("/governance/evaluate", h.EvaluateGovernanceAction)
	r.Post("/governance/emergency", h.ApplyGovernanceEmergencyControl)
	r.Get("/governance/emergency", h.GetGovernanceEmergencyControls)
	r.Get("/governance/status", h.GetGovernanceStatus)
	r.Post("/governance/rejections", h.RecordHumanRejection)

	// Phase 7.11 Enterprise Autonomous Operations Optimization & Resilience routes
	r.Get("/resilience/health", h.GetResilienceHealth)
	r.Get("/resilience/stuck-workflows", h.GetStuckWorkflows)
	r.Post("/resilience/recover-stuck", h.RecoverStuckWorkflow)
	r.Get("/resilience/failed-work", h.GetFailedWorkItems)
	r.Post("/resilience/replay-work", h.ReplayFailedWork)
	r.Get("/resilience/backpressure", h.GetBackpressureMetrics)
}

func (h *Handler) StartWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req StartWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	wf, err := h.svc.StartWorkflow(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		if errors.Is(err, ErrEmergencyStopActive) {
			utils.Error(w, http.StatusServiceUnavailable, err.Error(), "EMERGENCY_STOP_ACTIVE")
			return
		}
		if errors.Is(err, ErrAutonomyRestricted) {
			utils.Error(w, http.StatusForbidden, err.Error(), "AUTONOMY_RESTRICTED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusCreated, "Enterprise autonomous workflow started successfully", wf)
}

func (h *Handler) GetWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	workflowID := chi.URLParam(r, "workflow_id")
	if workflowID == "" {
		utils.Error(w, http.StatusBadRequest, "workflow_id path parameter is required", "BAD_REQUEST")
		return
	}

	wf, err := h.svc.GetWorkflow(r.Context(), orgID, workflowID)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		if errors.Is(err, ErrWorkflowNotFound) {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Enterprise workflow retrieved successfully", wf)
}

func (h *Handler) ListWorkflows(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	filter := WorkflowFilter{
		OrgID:        orgID,
		WorkflowType: r.URL.Query().Get("type"),
		State:        r.URL.Query().Get("state"),
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			filter.Limit = val
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil {
			filter.Offset = val
		}
	}

	workflows, total, err := h.svc.ListWorkflows(r.Context(), filter)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Workflows listed successfully", map[string]interface{}{
		"workflows": workflows,
		"total":     total,
		"limit":     filter.Limit,
		"offset":    filter.Offset,
	})
}

func (h *Handler) PauseWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	workflowID := chi.URLParam(r, "workflow_id")
	var req PauseWorkflowRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.svc.PauseWorkflow(r.Context(), orgID, workflowID, req); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusOK, "Workflow paused successfully", nil)
}

func (h *Handler) ResumeWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	workflowID := chi.URLParam(r, "workflow_id")
	if err := h.svc.ResumeWorkflow(r.Context(), orgID, workflowID); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusOK, "Workflow resumed successfully", nil)
}

func (h *Handler) CancelWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	workflowID := chi.URLParam(r, "workflow_id")
	var req CancelWorkflowRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.svc.CancelWorkflow(r.Context(), orgID, workflowID, req); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusOK, "Workflow cancelled successfully", nil)
}

func (h *Handler) ApproveWorkflowStep(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	workflowID := chi.URLParam(r, "workflow_id")
	stepID := chi.URLParam(r, "step_id")
	userID := h.getUserID(r)

	var req ApproveStepRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.svc.ApproveWorkflowStep(r.Context(), orgID, workflowID, stepID, userID, req); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusOK, "Workflow step approved successfully", nil)
}

func (h *Handler) RejectWorkflowStep(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	workflowID := chi.URLParam(r, "workflow_id")
	stepID := chi.URLParam(r, "step_id")
	userID := h.getUserID(r)

	var req RejectStepRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.svc.RejectWorkflowStep(r.Context(), orgID, workflowID, stepID, userID, req); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusOK, "Workflow step rejected successfully", nil)
}

func (h *Handler) RecoverInterruptedWorkflows(w http.ResponseWriter, r *http.Request) {
	report, err := h.svc.RecoverInterruptedWorkflows(r.Context())
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Workflow recovery executed successfully", report)
}

func (h *Handler) TriggerBusinessEvent(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req TriggerBusinessEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	wf, err := h.svc.TriggerBusinessEvent(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrDuplicateEventTrigger) {
			utils.Error(w, http.StatusConflict, err.Error(), "DUPLICATE_EVENT")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusCreated, "Event-triggered enterprise workflow initiated", wf)
}

func (h *Handler) EmergencyControl(w http.ResponseWriter, r *http.Request) {
	orgID, actor, err := h.getUser(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req EmergencyControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	if err := h.svc.EmergencyControl(r.Context(), orgID, req, actor); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusOK, "Enterprise emergency control updated successfully", nil)
}

func (h *Handler) GetPlatformOverview(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	overview, err := h.svc.GetPlatformOverview(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Platform overview retrieved successfully", overview)
}

func (h *Handler) getOrgID(r *http.Request) (int64, error) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		return 0, ErrUnauthorizedTenant
	}
	return userCtx.OrgID, nil
}

func (h *Handler) getUser(r *http.Request) (int64, string, error) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		return 0, "", ErrUnauthorizedTenant
	}
	actor := userCtx.Role
	if actor == "" {
		actor = "EnterpriseOperator"
	}
	return userCtx.OrgID, actor, nil
}

func (h *Handler) getUserID(r *http.Request) int64 {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		return 0
	}
	return userCtx.UserID
}

func (h *Handler) parseShipmentID(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "shipment_id")
	return strconv.ParseInt(idStr, 10, 64)
}

func (h *Handler) InitiateShipmentLifecycle(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	shipmentID, err := h.parseShipmentID(r)
	if err != nil || shipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment_id", "BAD_REQUEST")
		return
	}

	var req InitiateShipmentLifecycleRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	wf, err := h.shipmentLifecycleSvc.InitiateShipmentLifecycle(r.Context(), orgID, shipmentID, req.CorrelationID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusCreated, "Autonomous shipment lifecycle initiated", wf)
}

func (h *Handler) GetShipmentLifecycle(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	shipmentID, err := h.parseShipmentID(r)
	if err != nil || shipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment_id", "BAD_REQUEST")
		return
	}

	wf, err := h.shipmentLifecycleSvc.GetShipmentLifecycle(r.Context(), orgID, shipmentID)
	if err != nil {
		if errors.Is(err, ErrShipmentWorkflowNotFound) {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Autonomous shipment lifecycle retrieved", wf)
}

func (h *Handler) ProcessShipmentLifecycleEvent(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	shipmentID, err := h.parseShipmentID(r)
	if err != nil || shipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment_id", "BAD_REQUEST")
		return
	}

	var event ShipmentLifecycleEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "BAD_REQUEST")
		return
	}
	event.ShipmentID = shipmentID

	wf, err := h.shipmentLifecycleSvc.ProcessShipmentEvent(r.Context(), orgID, event)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Shipment event processed by autonomous lifecycle", wf)
}

func (h *Handler) EvaluateShipmentETA(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	shipmentID, err := h.parseShipmentID(r)
	if err != nil || shipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment_id", "BAD_REQUEST")
		return
	}

	wf, err := h.shipmentLifecycleSvc.EvaluateETAPrediction(r.Context(), orgID, shipmentID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Predictive ETA evaluated against authoritative milestone", wf)
}

func (h *Handler) ReplanShipmentLifecycle(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	shipmentID, err := h.parseShipmentID(r)
	if err != nil || shipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment_id", "BAD_REQUEST")
		return
	}

	var req ReplanShipmentLifecycleRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Reason == "" {
		req.Reason = "Operator requested adaptive replanning"
	}

	wf, err := h.shipmentLifecycleSvc.TriggerAdaptiveReplanning(r.Context(), orgID, shipmentID, req.Reason)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Adaptive replanning completed", wf)
}

func (h *Handler) DeliverShipmentLifecycle(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	shipmentID, err := h.parseShipmentID(r)
	if err != nil || shipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment_id", "BAD_REQUEST")
		return
	}

	wf, err := h.shipmentLifecycleSvc.TransitionToDelivered(r.Context(), orgID, shipmentID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Shipment transitioned to delivered and post-delivery audit initiated", wf)
}

func (h *Handler) AuditPostDelivery(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	shipmentID, err := h.parseShipmentID(r)
	if err != nil || shipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment_id", "BAD_REQUEST")
		return
	}

	wf, err := h.shipmentLifecycleSvc.ExecutePostDeliveryAudit(r.Context(), orgID, shipmentID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Post-delivery audit completed", wf)
}

func (h *Handler) RecordShipmentOutcome(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	shipmentID, err := h.parseShipmentID(r)
	if err != nil || shipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment_id", "BAD_REQUEST")
		return
	}

	var req RecordShipmentOutcomeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "BAD_REQUEST")
		return
	}

	res, err := h.shipmentLifecycleSvc.RecordShipmentOutcome(r.Context(), orgID, shipmentID, req)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Outcome and learning recorded", res)
}

// =====================================================================
// Phase 7.3 Autonomous Quote-to-Cash Commercial Lifecycle Handlers
// =====================================================================

func (h *Handler) parseWorkflowIDParam(r *http.Request) string {
	return chi.URLParam(r, "workflow_id")
}

func (h *Handler) parseRFQIDParam(r *http.Request) string {
	return chi.URLParam(r, "rfq_id")
}

func (h *Handler) InitiateCommercialLifecycle(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	rfqID := h.parseRFQIDParam(r)
	if rfqID == "" {
		utils.Error(w, http.StatusBadRequest, "rfq_id path parameter is required", "BAD_REQUEST")
		return
	}

	var body struct {
		CorrelationID string `json:"correlation_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	wf, err := h.commercialLifecycleSvc.InitiateCommercialLifecycle(r.Context(), orgID, rfqID, body.CorrelationID)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		if errors.Is(err, ErrPromptInjectionDetected) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "PROMPT_INJECTION_DETECTED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusCreated, "Commercial quote-to-cash workflow initiated", wf)
}

func (h *Handler) GetCommercialLifecycle(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)
	if wfID == "" {
		utils.Error(w, http.StatusBadRequest, "workflow_id is required", "BAD_REQUEST")
		return
	}

	wf, err := h.commercialLifecycleSvc.GetCommercialLifecycle(r.Context(), orgID, wfID)
	if err != nil {
		if errors.Is(err, ErrCommercialWorkflowNotFound) {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Commercial lifecycle retrieved", wf)
}

func (h *Handler) ProcessCommercialLifecycleEvent(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var event CommercialLifecycleEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid event payload", "BAD_REQUEST")
		return
	}

	wf, err := h.commercialLifecycleSvc.ProcessCommercialEvent(r.Context(), orgID, event)
	if err != nil {
		if errors.Is(err, ErrDuplicateEventTrigger) {
			utils.Error(w, http.StatusConflict, err.Error(), "DUPLICATE_EVENT")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Commercial lifecycle event processed", wf)
}

func (h *Handler) ExtractCommercialRFQ(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.commercialLifecycleSvc.ExtractRFQRequirements(r.Context(), orgID, wfID)
	if err != nil {
		if errors.Is(err, ErrMissingRFQRequirements) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "MISSING_RFQ_REQUIREMENTS")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "RFQ requirements extracted", wf)
}

func (h *Handler) QualifyCommercialRFQ(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.commercialLifecycleSvc.QualifyRFQ(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Commercial RFQ qualified", wf)
}

func (h *Handler) OptimizeCommercialPricing(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.commercialLifecycleSvc.OptimizePricingAndMargin(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Pricing and margin optimization complete", wf)
}

func (h *Handler) CheckCommercialCompliance(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.commercialLifecycleSvc.ValidateContractAndCompliance(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Contract and compliance validation complete", wf)
}

func (h *Handler) PrepareCommercialQuote(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.commercialLifecycleSvc.PrepareQuotation(r.Context(), orgID, wfID)
	if err != nil {
		if errors.Is(err, ErrCommercialApprovalRequired) {
			utils.Success(w, http.StatusAccepted, "Quotation prepared but awaiting policy approval", wf)
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Quotation prepared and ready for delivery", wf)
}

func (h *Handler) NegotiateCommercialQuote(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		CounterPrice float64 `json:"counter_price"`
		Notes        string  `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	wf, err := h.commercialLifecycleSvc.HandleCustomerNegotiation(r.Context(), orgID, wfID, req.CounterPrice, req.Notes)
	if err != nil {
		if errors.Is(err, ErrPromptInjectionDetected) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "PROMPT_INJECTION_DETECTED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Negotiation options generated", wf)
}

func (h *Handler) ApplyCommercialNegotiation(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		OptionID string `json:"option_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.OptionID == "" {
		utils.Error(w, http.StatusBadRequest, "option_id is required", "BAD_REQUEST")
		return
	}

	wf, err := h.commercialLifecycleSvc.ApplyNegotiationOption(r.Context(), orgID, wfID, req.OptionID)
	if err != nil {
		if errors.Is(err, ErrCommercialApprovalRequired) {
			utils.Success(w, http.StatusAccepted, "Negotiation option requires human approval", wf)
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Negotiation option applied", wf)
}

func (h *Handler) AcceptCommercialQuote(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		QuotationID int64 `json:"quotation_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.QuotationID <= 0 {
		req.QuotationID = 3001
	}

	wf, err := h.commercialLifecycleSvc.ProcessQuoteAcceptance(r.Context(), orgID, wfID, req.QuotationID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Quote acceptance verified", wf)
}

func (h *Handler) HandoffCommercialBooking(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.commercialLifecycleSvc.HandoffToBooking(r.Context(), orgID, wfID)
	if err != nil {
		if errors.Is(err, ErrDuplicateCommercialAction) {
			utils.Error(w, http.StatusConflict, err.Error(), "DUPLICATE_ACTION")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Commercial booking converted successfully", wf)
}

func (h *Handler) HandoffCommercialShipment(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		ShipmentID int64 `json:"shipment_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.ShipmentID <= 0 {
		req.ShipmentID = 101
	}

	wf, err := h.commercialLifecycleSvc.HandoffToShipmentLifecycle(r.Context(), orgID, wfID, req.ShipmentID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Handoff to Phase 7.2 shipment lifecycle completed", wf)
}

func (h *Handler) EvaluateCommercialInvoice(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.commercialLifecycleSvc.EvaluateInvoiceReadiness(r.Context(), orgID, wfID)
	if err != nil {
		if errors.Is(err, ErrDuplicateCommercialAction) {
			utils.Error(w, http.StatusConflict, err.Error(), "DUPLICATE_ACTION")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Invoice evaluated and generated", wf)
}

func (h *Handler) AssessCommercialCollection(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.commercialLifecycleSvc.AssessCollectionsStrategy(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Receivables collection strategy assessed", wf)
}

func (h *Handler) RecordCommercialOutcome(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req CommercialOutcomeFeedback
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid feedback payload", "BAD_REQUEST")
		return
	}
	if req.WorkflowID == "" {
		req.WorkflowID = h.parseWorkflowIDParam(r)
	}

	wf, err := h.commercialLifecycleSvc.RecordCommercialOutcome(r.Context(), orgID, req)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Commercial outcome and learning recorded", wf)
}

// ---------------------------------------------------------------------
// Phase 7.4 Exception Management Handlers
// ---------------------------------------------------------------------

func (h *Handler) DetectException(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var event EnterpriseExceptionEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid exception event payload", "BAD_REQUEST")
		return
	}

	wf, err := h.exceptionManagementSvc.DetectAndInitiateException(r.Context(), orgID, event)
	if err != nil {
		if errors.Is(err, ErrDuplicateCommercialAction) {
			utils.Error(w, http.StatusConflict, err.Error(), "DUPLICATE_EVENT")
			return
		}
		if errors.Is(err, ErrPromptInjectionDetected) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "PROMPT_INJECTION_DETECTED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusCreated, "Exception detected and workflow initiated", wf)
}

func (h *Handler) GetExceptionWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.exceptionManagementSvc.GetExceptionWorkflow(r.Context(), orgID, wfID)
	if err != nil {
		if errors.Is(err, ErrWorkflowNotFound) {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Exception workflow retrieved", wf)
}

func (h *Handler) InvestigateException(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.exceptionManagementSvc.InvestigateRootCause(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Root-cause investigation completed", wf)
}

func (h *Handler) AssessExceptionImpact(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.exceptionManagementSvc.AssessCrossModuleImpact(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Cross-module impact analysis completed", wf)
}

func (h *Handler) PlanExceptionRecovery(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.exceptionManagementSvc.PlanRecovery(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Recovery options planned", wf)
}

func (h *Handler) SelectExceptionRecoveryOption(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)
	optionID := chi.URLParam(r, "option_id")

	wf, err := h.exceptionManagementSvc.SelectAndGovernRecoveryOption(r.Context(), orgID, wfID, optionID)
	if err != nil {
		if errors.Is(err, ErrApprovalRequired) {
			utils.Success(w, http.StatusAccepted, "Recovery option requires approval; approval request registered", wf)
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Recovery option selected and governed", wf)
}

func (h *Handler) ExecuteExceptionRecoveryStep(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		StepID string `json:"step_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	wf, err := h.exceptionManagementSvc.ExecuteRecoveryStep(r.Context(), orgID, wfID, req.StepID)
	if err != nil {
		if errors.Is(err, ErrApprovalRequired) {
			utils.Error(w, http.StatusForbidden, err.Error(), "APPROVAL_REQUIRED")
			return
		}
		if errors.Is(err, ErrDuplicateCommercialAction) {
			utils.Error(w, http.StatusConflict, err.Error(), "DUPLICATE_EXECUTION")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Recovery action executed through Action System", wf)
}

func (h *Handler) VerifyExceptionRecovery(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		StepID string `json:"step_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	result, err := h.exceptionManagementSvc.VerifyRecoveryAction(r.Context(), orgID, wfID, req.StepID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Recovery action verified against authoritative data", result)
}

func (h *Handler) TriggerAdaptiveRecovery(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		FailureReason string `json:"failure_reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.FailureReason == "" {
		req.FailureReason = "Verification or action failure"
	}

	wf, err := h.exceptionManagementSvc.TriggerAdaptiveRecovery(r.Context(), orgID, wfID, req.FailureReason)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Adaptive recovery replanning completed", wf)
}

func (h *Handler) TransitionExceptionToMonitoring(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		MonitoringMilestone string `json:"monitoring_milestone"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.MonitoringMilestone == "" {
		req.MonitoringMilestone = "next_checkpoint_cleared"
	}

	wf, err := h.exceptionManagementSvc.TransitionToMonitoring(r.Context(), orgID, wfID, req.MonitoringMilestone)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Exception moved to post-resolution monitoring", wf)
}

func (h *Handler) ResolveException(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		ResolutionEvidence string `json:"resolution_evidence"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	wf, err := h.exceptionManagementSvc.ResolveException(r.Context(), orgID, wfID, req.ResolutionEvidence)
	if err != nil {
		if errors.Is(err, ErrResolutionEvidenceMissing) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "EVIDENCE_REQUIRED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Exception resolved with authoritative proof", wf)
}

func (h *Handler) EscalateException(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Reason == "" {
		req.Reason = "Escalated by operator/policy"
	}

	wf, err := h.exceptionManagementSvc.EscalateException(r.Context(), orgID, wfID, req.Reason)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Exception escalated to human intervention", wf)
}

func (h *Handler) RecordExceptionOutcome(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var feedback ExceptionOutcomeFeedback
	if err := json.NewDecoder(r.Body).Decode(&feedback); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid feedback payload", "BAD_REQUEST")
		return
	}
	if feedback.WorkflowID == "" {
		feedback.WorkflowID = h.parseWorkflowIDParam(r)
	}

	wf, err := h.exceptionManagementSvc.RecordExceptionOutcome(r.Context(), orgID, feedback)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Exception outcome recorded and memory updated", wf)
}

// ---------------------------------------------------------------------
// Phase 7.5 Customer Relationship Management Handlers
// ---------------------------------------------------------------------

func (h *Handler) InitiateCustomerRelationshipWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	custIDStr := chi.URLParam(r, "customer_id")
	custID, err := strconv.ParseInt(custIDStr, 10, 64)
	if err != nil || custID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid customer_id", "BAD_REQUEST")
		return
	}

	wf, err := h.customerRelationshipSvc.InitiateCustomerWorkflow(r.Context(), orgID, custID, "")
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusCreated, "Autonomous customer relationship workflow initiated", wf)
}

func (h *Handler) GetCustomerRelationshipWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.customerRelationshipSvc.GetCustomerWorkflow(r.Context(), orgID, wfID)
	if err != nil {
		if errors.Is(err, ErrCustomerWorkflowNotFound) {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Customer relationship workflow retrieved", wf)
}

func (h *Handler) EvaluateCustomerRelationshipHealth(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.customerRelationshipSvc.EvaluateCustomerHealth(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Customer health evaluated", wf)
}

func (h *Handler) DetectCustomerRisksAndOpportunities(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.customerRelationshipSvc.DetectRisksAndOpportunities(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Customer risks and opportunities detected", wf)
}

func (h *Handler) InvestigateCustomerRootCauses(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.customerRelationshipSvc.InvestigateRootCauses(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Multi-agent root-cause investigation completed", wf)
}

func (h *Handler) PlanCustomerInterventions(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.customerRelationshipSvc.PlanInterventionOptions(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Customer intervention options planned", wf)
}

func (h *Handler) SelectCustomerInterventionOption(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)
	optionID := chi.URLParam(r, "option_id")

	wf, err := h.customerRelationshipSvc.SelectAndGovernIntervention(r.Context(), orgID, wfID, optionID)
	if err != nil {
		if errors.Is(err, ErrCustomerApprovalRequired) {
			utils.Success(w, http.StatusAccepted, "Customer intervention requires human approval; approval request registered", wf)
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Customer intervention selected and governed", wf)
}

func (h *Handler) ExecuteCustomerIntervention(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		StepID string `json:"step_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	wf, err := h.customerRelationshipSvc.ExecuteIntervention(r.Context(), orgID, wfID, req.StepID)
	if err != nil {
		if errors.Is(err, ErrDuplicateCustomerCommunication) {
			utils.Error(w, http.StatusConflict, err.Error(), "DUPLICATE_COMMUNICATION")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Customer intervention executed through Action System", wf)
}

func (h *Handler) VerifyCustomerIntervention(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		StepID string `json:"step_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	wf, err := h.customerRelationshipSvc.VerifyIntervention(r.Context(), orgID, wfID, req.StepID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Customer intervention delivery verified against authoritative gateway", wf)
}

func (h *Handler) ProcessCustomerResponse(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid message payload", "BAD_REQUEST")
		return
	}

	wf, err := h.customerRelationshipSvc.ProcessCustomerResponse(r.Context(), orgID, wfID, req.Message)
	if err != nil {
		if errors.Is(err, ErrPromptInjectionDetected) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "PROMPT_INJECTION_DETECTED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Customer message classified and processed securely", wf)
}

func (h *Handler) ProcessCustomerLifecycleEvent(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var event CustomerLifecycleEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid event payload", "BAD_REQUEST")
		return
	}

	wf, err := h.customerRelationshipSvc.ProcessLifecycleEvent(r.Context(), orgID, event)
	if err != nil {
		if errors.Is(err, ErrDuplicateEventTrigger) {
			utils.Error(w, http.StatusConflict, err.Error(), "DUPLICATE_EVENT")
			return
		}
		if errors.Is(err, ErrPromptInjectionDetected) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "PROMPT_INJECTION_DETECTED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusCreated, "Customer lifecycle event ingested", wf)
}

func (h *Handler) RecordCustomerOutcome(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var feedback CustomerOutcomeFeedback
	if err := json.NewDecoder(r.Body).Decode(&feedback); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid feedback payload", "BAD_REQUEST")
		return
	}
	if feedback.WorkflowID == "" {
		feedback.WorkflowID = h.parseWorkflowIDParam(r)
	}

	wf, err := h.customerRelationshipSvc.RecordCustomerOutcome(r.Context(), orgID, feedback)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Customer relationship outcome recorded and memory updated", wf)
}

// ---------------------------------------------------------------------
// Phase 7.6 Autonomous Revenue and Margin Optimization HTTP Handlers
// ---------------------------------------------------------------------

func (h *Handler) InitiateRevenueWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req struct {
		EntityType string `json:"entity_type"`
		EntityID   string `json:"entity_id"`
		CorrID     string `json:"correlation_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "BAD_REQUEST")
		return
	}

	wf, err := h.revenueOptimizationSvc.InitiateRevenueWorkflow(r.Context(), orgID, req.EntityType, req.EntityID, req.CorrID)
	if err != nil {
		if errors.Is(err, ErrPromptInjectionDetected) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "PROMPT_INJECTION_DETECTED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusCreated, "Autonomous revenue optimization workflow initiated", wf)
}

func (h *Handler) GetRevenueWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.revenueOptimizationSvc.GetRevenueWorkflow(r.Context(), orgID, wfID)
	if err != nil {
		if errors.Is(err, ErrRevenueWorkflowNotFound) {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Revenue optimization workflow retrieved", wf)
}

func (h *Handler) EvaluateCostAndMargin(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.revenueOptimizationSvc.EvaluateCostAndMargin(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Cost and margin evaluated against authoritative data", wf)
}

func (h *Handler) EvaluateCarrierEconomics(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.revenueOptimizationSvc.EvaluateCarrierEconomics(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Carrier economics and historical reliability evaluated", wf)
}

func (h *Handler) AssessRevenueCustomerValue(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.revenueOptimizationSvc.AssessCustomerValue(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Customer multi-dimensional value scorecard assessed", wf)
}

func (h *Handler) ConductMultiAgentOptimization(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.revenueOptimizationSvc.ConductMultiAgentOptimization(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Multi-agent commercial reasoning synthesized", wf)
}

func (h *Handler) FormulatePricingRecommendation(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.revenueOptimizationSvc.FormulatePricingRecommendation(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Pricing recommendations and options formulated", wf)
}

func (h *Handler) SelectRevenueOptimizationOption(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)
	optionID := chi.URLParam(r, "option_id")

	wf, err := h.revenueOptimizationSvc.SelectAndGovernOptimizationOption(r.Context(), orgID, wfID, optionID)
	if err != nil {
		if errors.Is(err, ErrPricingPolicyViolation) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "PRICING_POLICY_VIOLATION")
			return
		}
		if errors.Is(err, ErrRevenueApprovalRequired) {
			utils.Success(w, http.StatusAccepted, "Option requires executive approval; routed to waiting queue", wf)
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Commercial option governed and approved for execution", wf)
}

func (h *Handler) ExecuteRevenueCommercialAction(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		StepID string `json:"step_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	wf, err := h.revenueOptimizationSvc.ExecuteCommercialAction(r.Context(), orgID, wfID, req.StepID)
	if err != nil {
		if errors.Is(err, ErrActionExecutionFailed) {
			utils.Error(w, http.StatusBadGateway, err.Error(), "ACTION_EXECUTION_FAILED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Commercial action executed via authoritative Action System", wf)
}

func (h *Handler) VerifyRevenueCommercialAction(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		StepID string `json:"step_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	wf, err := h.revenueOptimizationSvc.VerifyCommercialAction(r.Context(), orgID, wfID, req.StepID)
	if err != nil {
		if errors.Is(err, ErrCommercialVerificationFailed) {
			utils.Error(w, http.StatusExpectationFailed, err.Error(), "VERIFICATION_FAILED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Commercial action verified against authoritative business database", wf)
}

func (h *Handler) OptimizeNegotiationRequest(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		TargetDiscountPct float64 `json:"target_discount_pct"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid negotiation payload", "BAD_REQUEST")
		return
	}

	wf, err := h.revenueOptimizationSvc.OptimizeNegotiationRequest(r.Context(), orgID, wfID, req.TargetDiscountPct)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Negotiation alternatives evaluated and prepared", wf)
}

func (h *Handler) ProcessRevenueLifecycleEvent(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var event RevenueLifecycleEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid event payload", "BAD_REQUEST")
		return
	}

	wf, err := h.revenueOptimizationSvc.ProcessLifecycleEvent(r.Context(), orgID, event)
	if err != nil {
		if errors.Is(err, ErrDuplicateEventTrigger) {
			utils.Error(w, http.StatusConflict, err.Error(), "DUPLICATE_EVENT")
			return
		}
		if errors.Is(err, ErrPromptInjectionDetected) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "PROMPT_INJECTION_DETECTED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusCreated, "Revenue lifecycle event ingested", wf)
}

func (h *Handler) RecordRevenueOutcome(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var feedback RevenueOutcomeFeedback
	if err := json.NewDecoder(r.Body).Decode(&feedback); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid feedback payload", "BAD_REQUEST")
		return
	}
	if feedback.WorkflowID == "" {
		feedback.WorkflowID = h.parseWorkflowIDParam(r)
	}

	wf, err := h.revenueOptimizationSvc.RecordRevenueOutcome(r.Context(), orgID, feedback)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Revenue outcome recorded and workforce memory updated", wf)
}

// ---------------------------------------------------------------------
// Phase 7.7 Autonomous Contract, Compliance and Risk Governance HTTP Handlers
// ---------------------------------------------------------------------

func (h *Handler) InitiateRiskWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req struct {
		EntityType string `json:"entity_type"`
		EntityID   string `json:"entity_id"`
		CorrID     string `json:"correlation_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "BAD_REQUEST")
		return
	}

	wf, err := h.contractComplianceRiskSvc.InitiateRiskWorkflow(r.Context(), orgID, req.EntityType, req.EntityID, req.CorrID)
	if err != nil {
		if errors.Is(err, ErrPromptInjectionDetected) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "PROMPT_INJECTION_DETECTED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusCreated, "Autonomous risk governance workflow initiated", wf)
}

func (h *Handler) GetRiskWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.contractComplianceRiskSvc.GetRiskWorkflow(r.Context(), orgID, wfID)
	if err != nil {
		if errors.Is(err, ErrRiskWorkflowNotFound) {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Risk governance workflow retrieved", wf)
}

func (h *Handler) CollectRiskEvidence(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var evidence RiskEvidenceItem
	if err := json.NewDecoder(r.Body).Decode(&evidence); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid evidence payload", "BAD_REQUEST")
		return
	}

	wf, err := h.contractComplianceRiskSvc.CollectRiskEvidence(r.Context(), orgID, wfID, evidence)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Risk evidence collected with verified provenance", wf)
}

func (h *Handler) ConductMultiAgentRiskAssessment(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.contractComplianceRiskSvc.ConductMultiAgentRiskAssessment(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Multi-agent contract and compliance assessment synthesized", wf)
}

func (h *Handler) AssessCrossDomainImpact(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.contractComplianceRiskSvc.AssessCrossDomainImpact(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Cross-domain cascading impact assessed across operational and commercial boundaries", wf)
}

func (h *Handler) PlanRiskMitigationOptions(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.contractComplianceRiskSvc.PlanMitigationOptions(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Governed risk mitigation proposals formulated", wf)
}

func (h *Handler) SelectRiskMitigationOption(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)
	optionID := chi.URLParam(r, "option_id")

	wf, err := h.contractComplianceRiskSvc.SelectAndGovernMitigationOption(r.Context(), orgID, wfID, optionID)
	if err != nil {
		if errors.Is(err, ErrCriticalRiskApprovalRequired) {
			utils.Success(w, http.StatusAccepted, "Mitigation option requires executive approval; routed to waiting queue", wf)
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Mitigation strategy selected and governed for execution", wf)
}

func (h *Handler) ExecuteRiskMitigationAction(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		StepID string `json:"step_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	wf, err := h.contractComplianceRiskSvc.ExecuteMitigationAction(r.Context(), orgID, wfID, req.StepID)
	if err != nil {
		if errors.Is(err, ErrDuplicateMitigationAction) {
			utils.Error(w, http.StatusConflict, err.Error(), "DUPLICATE_MITIGATION")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Mitigation action executed via authoritative Action System", wf)
}

func (h *Handler) VerifyRiskMitigationAction(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		StepID string `json:"step_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	wf, err := h.contractComplianceRiskSvc.VerifyMitigationAction(r.Context(), orgID, wfID, req.StepID)
	if err != nil {
		if errors.Is(err, ErrMitigationVerificationFailed) {
			utils.Error(w, http.StatusExpectationFailed, err.Error(), "VERIFICATION_FAILED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Mitigation action verified against authoritative external gateway", wf)
}

func (h *Handler) TransitionRiskToMonitoring(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	wf, err := h.contractComplianceRiskSvc.TransitionToMonitoring(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Workflow transitioned to active post-mitigation monitoring", wf)
}

func (h *Handler) ResolveRiskWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		ResolutionSummary string `json:"resolution_summary"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	wf, err := h.contractComplianceRiskSvc.ResolveRiskWorkflow(r.Context(), orgID, wfID, req.ResolutionSummary)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Risk formally resolved and closed in ledger", wf)
}

func (h *Handler) ReassessRiskCondition(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	wfID := h.parseWorkflowIDParam(r)

	var req struct {
		NewEvidence RiskEvidenceItem `json:"new_evidence"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid evidence payload", "BAD_REQUEST")
		return
	}

	wf, err := h.contractComplianceRiskSvc.ReassessRiskCondition(r.Context(), orgID, wfID, req.NewEvidence)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Risk condition adaptively reassessed with new evidence", wf)
}

func (h *Handler) ProcessRiskLifecycleEvent(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var event RiskLifecycleEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid event payload", "BAD_REQUEST")
		return
	}

	wf, err := h.contractComplianceRiskSvc.ProcessRiskLifecycleEvent(r.Context(), orgID, event)
	if err != nil {
		if errors.Is(err, ErrDuplicateEventTrigger) {
			utils.Error(w, http.StatusConflict, err.Error(), "DUPLICATE_EVENT")
			return
		}
		if errors.Is(err, ErrPromptInjectionDetected) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "PROMPT_INJECTION_DETECTED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusCreated, "Risk lifecycle event ingested", wf)
}

func (h *Handler) RecordRiskOutcome(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var feedback RiskOutcomeFeedback
	if err := json.NewDecoder(r.Body).Decode(&feedback); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid feedback payload", "BAD_REQUEST")
		return
	}
	if feedback.WorkflowID == "" {
		feedback.WorkflowID = h.parseWorkflowIDParam(r)
	}

	wf, err := h.contractComplianceRiskSvc.RecordRiskOutcome(r.Context(), orgID, feedback)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Risk outcome recorded and workforce memory trained", wf)
}

// ---------------------------------------------------------------------
// Phase 7.8 Event Mesh & Autonomous Workflow Engine Handlers
// ---------------------------------------------------------------------

func (h *Handler) IngestEventMeshEvent(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req IngestBusinessEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid event request payload", "BAD_REQUEST")
		return
	}

	evt, wf, err := h.eventMeshSvc.IngestEvent(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrPromptInjectionDetected) {
			utils.Error(w, http.StatusBadRequest, "Malicious or suspicious prompt injection detected in event payload", "SECURITY_VIOLATION")
			return
		}
		if errors.Is(err, ErrStaleEventIgnored) {
			utils.Error(w, http.StatusUnprocessableEntity, "Stale event superseded by newer state", "STALE_EVENT")
			return
		}
		if strings.Contains(err.Error(), "validation failed") || strings.Contains(err.Error(), "missing required") {
			utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusCreated, "Business event ingested and processed by Event Mesh", map[string]interface{}{
		"event":    evt,
		"workflow": wf,
	})
}

func (h *Handler) GetEventMeshEvent(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	eventID := chi.URLParam(r, "event_id")
	if eventID == "" {
		utils.Error(w, http.StatusBadRequest, "event_id is required", "BAD_REQUEST")
		return
	}

	evt, err := h.eventMeshSvc.GetEvent(r.Context(), orgID, eventID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}
	utils.Success(w, http.StatusOK, "Event retrieved", evt)
}

func (h *Handler) ListEventMeshEvents(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	limit := 50
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

	events, total, err := h.eventMeshSvc.ListEvents(r.Context(), orgID, limit, offset)
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

func (h *Handler) GetEventMeshDeadLetters(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	dlList, err := h.eventMeshSvc.GetDeadLetters(r.Context(), orgID, limit)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Dead-letter events retrieved", dlList)
}

func (h *Handler) ReplayEventMeshDeadLetter(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	eventID := chi.URLParam(r, "event_id")
	if eventID == "" {
		utils.Error(w, http.StatusBadRequest, "event_id is required", "BAD_REQUEST")
		return
	}

	evt, wf, err := h.eventMeshSvc.ReplayDeadLetter(r.Context(), orgID, eventID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Dead-letter event replayed successfully", map[string]interface{}{
		"event":    evt,
		"workflow": wf,
	})
}

func (h *Handler) GetEventMeshOverview(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	ov, err := h.eventMeshSvc.GetEventMeshOverview(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Event mesh overview retrieved", ov)
}

func (h *Handler) ListEventMeshRoutingRules(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	rules, err := h.eventMeshSvc.ListRoutingRules(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Event routing rules retrieved", rules)
}

// ---------------------------------------------------------------------
// Phase 7.9 Enterprise Autonomous Control Tower Handlers
// ---------------------------------------------------------------------

func (h *Handler) GetControlTowerView(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	view, err := h.controlTowerSvc.GetControlTowerView(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Control Tower comprehensive view retrieved", view)
}

func (h *Handler) GetControlTowerWorkflowTrace(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	wfID := chi.URLParam(r, "workflow_id")
	if wfID == "" {
		utils.Error(w, http.StatusBadRequest, "workflow_id is required", "BAD_REQUEST")
		return
	}

	trace, err := h.controlTowerSvc.GetWorkflowTrace(r.Context(), orgID, wfID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}
	utils.Success(w, http.StatusOK, "Workflow trace retrieved", trace)
}

func (h *Handler) PerformControlTowerAction(w http.ResponseWriter, r *http.Request) {
	orgID, actor, err := h.getUser(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	wfID := chi.URLParam(r, "workflow_id")
	if wfID == "" {
		utils.Error(w, http.StatusBadRequest, "workflow_id is required", "BAD_REQUEST")
		return
	}

	var req struct {
		Action string `json:"action"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "BAD_REQUEST")
		return
	}

	err = h.controlTowerSvc.PerformGovernedControlAction(r.Context(), orgID, wfID, req.Action, actor, req.Reason)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, fmt.Sprintf("Workflow %s action '%s' executed", wfID, req.Action), nil)
}

// ---------------------------------------------------------------------
// Phase 7.10: Enterprise Autonomy, Governance & Safety Handlers
// ---------------------------------------------------------------------

func (h *Handler) EvaluateGovernanceAction(w http.ResponseWriter, r *http.Request) {
	orgID, _, err := h.getUser(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req GovernanceEvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid evaluation request body", "BAD_REQUEST")
		return
	}

	// Server-side tenant isolation enforcement: overwrite/verify OrgID
	if req.OrgID == 0 {
		req.OrgID = orgID
	} else if req.OrgID != orgID {
		utils.Error(w, http.StatusForbidden, "Cross-tenant action evaluation prohibited", "FORBIDDEN")
		return
	}

	decision, err := h.governanceSvc.EvaluateAction(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrEmergencyHaltActive) {
			utils.Error(w, http.StatusForbidden, err.Error(), "EMERGENCY_HALT")
			return
		}
		if errors.Is(err, ErrPromptInjectionDetected) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "PROMPT_INJECTION_DETECTED")
			return
		}
		if errors.Is(err, ErrAutonomousLoopDetected) {
			utils.Error(w, http.StatusTooManyRequests, err.Error(), "AUTONOMOUS_LOOP_DETECTED")
			return
		}
		if errors.Is(err, ErrAgentPrivilegeViolation) {
			utils.Error(w, http.StatusForbidden, err.Error(), "AGENT_PRIVILEGE_VIOLATION")
			return
		}
		if errors.Is(err, ErrHumanRejectionConflict) {
			utils.Error(w, http.StatusConflict, err.Error(), "HUMAN_REJECTION_CONFLICT")
			return
		}
		if errors.Is(err, ErrGovernanceDenied) {
			utils.Error(w, http.StatusForbidden, err.Error(), "GOVERNANCE_DENIED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "GOVERNANCE_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Governance evaluation complete", decision)
}

func (h *Handler) ApplyGovernanceEmergencyControl(w http.ResponseWriter, r *http.Request) {
	orgID, actor, err := h.getUser(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req struct {
		Scope    EmergencyScope `json:"scope"`
		Target   string         `json:"target"`
		IsHalted bool           `json:"is_halted"`
		Reason   string         `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid emergency control payload", "BAD_REQUEST")
		return
	}

	if req.Scope == "" {
		req.Scope = EmergencyScopeAll
	}
	if req.Reason == "" {
		req.Reason = "Emergency control executed by authorized operator"
	}

	err = h.governanceSvc.ApplyEmergencyControl(r.Context(), orgID, req.Scope, req.Target, req.IsHalted, req.Reason, actor)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	actionDesc := "resumed"
	if req.IsHalted {
		actionDesc = "halted"
	}
	utils.Success(w, http.StatusOK, fmt.Sprintf("Emergency control %s for scope=%s target=%s", actionDesc, req.Scope, req.Target), nil)
}

func (h *Handler) GetGovernanceEmergencyControls(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	controls, err := h.governanceSvc.GetActiveEmergencyControls(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Active emergency controls retrieved", controls)
}

func (h *Handler) GetGovernanceStatus(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	status, err := h.governanceSvc.GetGovernanceStatus(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Governance status retrieved", status)
}

func (h *Handler) RecordHumanRejection(w http.ResponseWriter, r *http.Request) {
	orgID, actor, err := h.getUser(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req struct {
		WorkflowID string `json:"workflow_id"`
		EntityType string `json:"entity_type"`
		EntityID   string `json:"entity_id"`
		ActionType string `json:"action_type"`
		Reason     string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid rejection request body", "BAD_REQUEST")
		return
	}

	err = h.governanceSvc.RecordHumanRejection(r.Context(), orgID, req.WorkflowID, req.ActionType, req.EntityID, req.Reason, actor)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Human rejection recorded and protected", nil)
}

// ---------------------------------------------------------------------
// Phase 7.11: Enterprise Autonomous Operations Optimization & Resilience Handlers
// ---------------------------------------------------------------------

func (h *Handler) GetResilienceHealth(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	health, err := h.resilienceSvc.GetPlatformHealth(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Enterprise platform health retrieved", health)
}

func (h *Handler) GetStuckWorkflows(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	threshStr := r.URL.Query().Get("threshold_seconds")
	thresh := 10 * time.Minute
	if threshStr != "" {
		if sec, err := strconv.Atoi(threshStr); err == nil && sec > 0 {
			thresh = time.Duration(sec) * time.Second
		}
	}

	stuck, err := h.resilienceSvc.DetectStuckWorkflows(r.Context(), orgID, thresh)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Stuck workflows scanned successfully", map[string]interface{}{
		"stuck_count": len(stuck),
		"workflows":   stuck,
	})
}

func (h *Handler) RecoverStuckWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, actor, err := h.getUser(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req struct {
		WorkflowID string `json:"workflow_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.WorkflowID == "" {
		utils.Error(w, http.StatusBadRequest, "workflow_id is required", "BAD_REQUEST")
		return
	}

	err = h.resilienceSvc.RecoverStuckWorkflow(r.Context(), orgID, req.WorkflowID, actor)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Stuck workflow recovered from checkpoint", map[string]interface{}{
		"workflow_id": req.WorkflowID,
		"status":      "RECOVERED",
	})
}

func (h *Handler) GetFailedWorkItems(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	items, err := h.resilienceSvc.GetFailedWorkItems(r.Context(), orgID, limit)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Failed work items retrieved", map[string]interface{}{
		"count": len(items),
		"items": items,
	})
}

func (h *Handler) ReplayFailedWork(w http.ResponseWriter, r *http.Request) {
	orgID, actor, err := h.getUser(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	var req struct {
		ItemID string `json:"item_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ItemID == "" {
		utils.Error(w, http.StatusBadRequest, "item_id is required", "BAD_REQUEST")
		return
	}

	err = h.resilienceSvc.ReplayFailedWorkItem(r.Context(), orgID, req.ItemID, actor)
	if err != nil {
		if errors.Is(err, ErrNonRetryableFailure) {
			utils.Error(w, http.StatusConflict, err.Error(), "NON_RETRYABLE_FAILURE")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Failed work item replayed successfully", map[string]interface{}{
		"item_id": req.ItemID,
		"status":  "REPLAYED",
	})
}

func (h *Handler) GetBackpressureMetrics(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	metrics, err := h.resilienceSvc.GetBackpressureMetrics(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Backpressure metrics retrieved", metrics)
}

