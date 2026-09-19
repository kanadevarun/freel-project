package workforce

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	// Agent Registry
	r.Get("/agents", h.ListAgents)
	r.Get("/agents/{agent_id}", h.GetAgent)
	r.Post("/agents", h.RegisterAgent)
	r.Put("/agents/{agent_id}", h.UpdateAgent)
	r.Get("/capabilities", h.ListCapabilities)

	// Workforce Tasks & Delegation
	r.Post("/tasks", h.CreateTask)
	r.Get("/tasks", h.ListTasks)
	r.Get("/tasks/{task_id}", h.GetTask)
	r.Get("/tasks/{task_id}/hierarchy", h.GetTaskHierarchy)
	r.Post("/tasks/{task_id}/execute", h.ExecuteTask)
	r.Post("/tasks/{task_id}/workflow-chain", h.ExecuteWorkflowChain)
	r.Post("/tasks/{task_id}/delegate", h.DelegateTask)
	r.Post("/tasks/{task_id}/handoff", h.InitiateHandoff)
	r.Get("/tasks/{task_id}/handoffs", h.ListHandoffs)

	// Context Items
	r.Post("/tasks/{task_id}/contexts", h.AddContextItem)
	r.Get("/tasks/{task_id}/contexts", h.ListContextItems)

	// Inter-Agent Messages
	r.Post("/tasks/{task_id}/messages", h.PostMessage)
	r.Get("/tasks/{task_id}/messages", h.ListMessages)

	// Phase 6.4: Collaborative Planning & Decision Making
	r.Post("/plans", h.CreateCollaborativePlan)
	r.Get("/plans/{plan_id}", h.GetCollaborativePlan)
	r.Post("/plans/{plan_id}/execute", h.ExecuteCollaborativePlan)

	// Phase 6.5: Multi-Agent Shipment & Exception Operations
	r.Post("/shipments/{shipment_id}/health", h.AssessShipmentHealth)
	r.Post("/shipments/{shipment_id}/exceptions/{exception_id}/investigate", h.InvestigateShipmentException)
	r.Post("/plans/{plan_id}/replan", h.ReplanShipmentOperation)
	r.Post("/events/shipment", h.HandleShipmentEvent)

	// Phase 6.6: Multi-Agent Customer, Sales, Pricing & Finance Workflows
	r.Post("/customers/{customer_id}/assess", h.AssessCustomerRelationship)
	r.Post("/leads/{lead_id}/evaluate", h.EvaluateLeadIntelligence)
	r.Post("/rfqs/{rfq_id}/evaluate", h.EvaluateRFQCommercialWorkflow)
	r.Post("/collections/assess", h.AssessInvoiceCollections)
	r.Post("/events/commercial", h.HandleCommercialEvent)

	// Phase 6.7: Multi-Agent Contract, Compliance & Risk
	r.Post("/risks/contract", h.AssessContractRisk)
	r.Post("/risks/compliance", h.AssessComplianceRisk)
	r.Post("/risks/cross-module", h.AssessCrossModuleRisk)
	r.Post("/risks/reassess", h.ReassessCrossModuleRisk)

	// Phase 6.8: Conflict Resolution, Memory & Learning
	r.Post("/conflicts/resolve", h.ResolveConflict)
	r.Post("/outcomes", h.RecordOutcome)
	r.Get("/outcomes", h.ListOutcomes)
	r.Post("/memory/query", h.QueryMemory)

	// Phase 6.9: Governed Multi-Agent Autonomy & Workforce Command Center
	r.Get("/command-center/overview", h.GetCommandCenterOverview)
	r.Get("/command-center/health", h.GetWorkforceHealth)
	r.Get("/command-center/workload", h.GetAgentWorkload)
	r.Post("/command-center/emergency-stop", h.EmergencyStop)
	r.Get("/command-center/emergency-stop", h.GetEmergencyStopStatus)
	r.Get("/command-center/approvals", h.ListWorkforceApprovals)
	r.Get("/command-center/escalations", h.ListWorkforceEscalations)
	r.Get("/command-center/activity", h.GetRecentActivity)
	r.Post("/command-center/evaluate-action", h.EvaluateActionPolicy)
	r.Post("/agents/{agent_id}/control", h.ControlAgent)
	r.Get("/plans/{plan_id}/inspect", h.InspectWorkflow)
	r.Post("/plans/{plan_id}/control", h.ControlWorkflow)
}

func (h *Handler) getOrgID(r *http.Request) (int64, error) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		return 0, errors.New("unauthorized: missing user context")
	}
	return userCtx.OrgID, nil
}

func (h *Handler) ListAgents(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	agentType := r.URL.Query().Get("agent_type")
	var isEnabled *bool
	if enabledStr := r.URL.Query().Get("is_enabled"); enabledStr != "" {
		b := enabledStr == "true" || enabledStr == "1"
		isEnabled = &b
	}

	agents, err := h.svc.ListAgents(r.Context(), orgID, agentType, isEnabled)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Workforce agents retrieved successfully", agents)
}

func (h *Handler) GetAgent(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	agentID := chi.URLParam(r, "agent_id")

	agent, err := h.svc.GetAgent(r.Context(), orgID, agentID)
	if errors.Is(err, ErrAgentNotFound) {
		utils.Error(w, http.StatusNotFound, "Agent not found", "NOT_FOUND")
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Workforce agent retrieved successfully", agent)
}

func (h *Handler) RegisterAgent(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req RegisterAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "BAD_REQUEST")
		return
	}

	agent, err := h.svc.RegisterAgent(r.Context(), orgID, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}
	utils.Success(w, http.StatusCreated, "Agent registered successfully", agent)
}

func (h *Handler) UpdateAgent(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	agentID := chi.URLParam(r, "agent_id")

	var req UpdateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "BAD_REQUEST")
		return
	}

	agent, err := h.svc.UpdateAgent(r.Context(), orgID, agentID, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}
	utils.Success(w, http.StatusOK, "Agent updated successfully", agent)
}

func (h *Handler) ListCapabilities(w http.ResponseWriter, r *http.Request) {
	caps := []map[string]string{
		{"key": CapShipmentRead, "name": "Shipment Read", "description": "Inspect tracking and operational milestones"},
		{"key": CapShipmentAnalyze, "name": "Shipment Analysis", "description": "Analyze delays, routings, and events"},
		{"key": CapShipmentPredict, "name": "Shipment Prediction", "description": "Predict ETA movements and transit risk"},
		{"key": CapExceptionRead, "name": "Exception Read", "description": "Inspect exception records and disruptions"},
		{"key": CapExceptionAnalyze, "name": "Exception Analysis", "description": "Diagnose root causes and operational impact"},
		{"key": CapExceptionRecommend, "name": "Exception Recommendation", "description": "Propose disruption mitigations and recovery options"},
		{"key": CapCustomerRead, "name": "Customer Read", "description": "Query customer profile, contacts, and preferences"},
		{"key": CapCustomerAnalyze, "name": "Customer Analysis", "description": "Assess customer sentiment, health, and relationship"},
		{"key": CapCustomerFollowupRecommend, "name": "Customer Follow-up Recommendation", "description": "Prepare communication recommendations and outreach drafts"},
		{"key": CapRFQRead, "name": "RFQ Read", "description": "Inspect RFQ parameters and cargo specifications"},
		{"key": CapPricingAnalyze, "name": "Pricing Analysis", "description": "Evaluate freight market benchmarks and margin constraints"},
		{"key": CapPricingRecommend, "name": "Pricing Recommendation", "description": "Recommend spot quotes and margin optimization strategies"},
		{"key": CapInvoiceRead, "name": "Invoice Read", "description": "Query billing records and payment statuses"},
		{"key": CapFinanceAnalyze, "name": "Finance Analysis", "description": "Audit billing, aging receivables, and discrepancies"},
		{"key": CapFinanceRecommend, "name": "Finance Recommendation", "description": "Recommend collection priorities and cash-flow mitigation"},
		{"key": CapContractRead, "name": "Contract Read", "description": "Inspect contract terms, rate agreements, and MSAs"},
		{"key": CapContractAnalyze, "name": "Contract Analysis", "description": "Extract clauses, free days, and validate operational commitments"},
		{"key": CapComplianceRead, "name": "Compliance Read", "description": "Inspect shipping documentation and regulatory filings"},
		{"key": CapComplianceAnalyze, "name": "Compliance Analysis", "description": "Audit documentation gaps, hazmat, and customs compliance"},
		{"key": CapPlanningCreate, "name": "Planning Creation", "description": "Formulate multi-step operational plans"},
		{"key": CapPlanningEvaluate, "name": "Planning Evaluation", "description": "Assess risk and feasibility of execution plans"},
		{"key": CapTaskDelegate, "name": "Task Delegation", "description": "Delegate specialized subtasks across the workforce"},
		{"key": CapWorkforceObserve, "name": "Workforce Observation", "description": "Observe agent execution health and task pipeline"},
		{"key": CapTaskMonitor, "name": "Task Monitoring", "description": "Detect stalled tasks and SLA deviations"},
		{"key": CapMonitoringObserve, "name": "Monitoring Observation", "description": "Track system events and execution health"},
		{"key": CapMemoryRetrieve, "name": "Memory Retrieval", "description": "Retrieve historical agent memory and learned outcomes"},
		{"key": CapMemoryAnalyze, "name": "Memory Analysis", "description": "Analyze recurring patterns and lessons learned"},
		{"key": CapOutcomeRecord, "name": "Outcome Recording", "description": "Record verified outcomes for continuous improvement"},
	}
	utils.Success(w, http.StatusOK, "Standard capabilities retrieved", caps)
}


func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	task, err := h.svc.CreateTask(r.Context(), orgID, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}
	utils.Success(w, http.StatusCreated, "Workforce task created successfully", task)
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	filter := TaskFilter{
		OrgID:           orgID,
		Status:          r.URL.Query().Get("status"),
		AssignedAgentID: r.URL.Query().Get("assigned_agent_id"),
		CorrelationID:   r.URL.Query().Get("correlation_id"),
	}
	if pID := r.URL.Query().Get("parent_task_id"); pID != "" {
		filter.ParentTaskID = &pID
	}
	if rID := r.URL.Query().Get("root_task_id"); rID != "" {
		filter.RootTaskID = &rID
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		filter.Limit, _ = strconv.Atoi(limitStr)
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		filter.Offset, _ = strconv.Atoi(offsetStr)
	}

	tasks, total, err := h.svc.ListTasks(r.Context(), filter)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Workforce tasks retrieved successfully", map[string]interface{}{
		"tasks":  tasks,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	taskID := chi.URLParam(r, "task_id")

	task, err := h.svc.GetTask(r.Context(), orgID, taskID)
	if errors.Is(err, ErrTaskNotFound) {
		utils.Error(w, http.StatusNotFound, "Task not found", "NOT_FOUND")
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Workforce task retrieved successfully", task)
}

func (h *Handler) GetTaskHierarchy(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	taskID := chi.URLParam(r, "task_id")

	hierarchy, err := h.svc.GetTaskHierarchy(r.Context(), orgID, taskID)
	if errors.Is(err, ErrTaskNotFound) {
		utils.Error(w, http.StatusNotFound, "Task not found", "NOT_FOUND")
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Task hierarchy retrieved successfully", hierarchy)
}

func (h *Handler) ExecuteTask(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	taskID := chi.URLParam(r, "task_id")

	task, sidecarResp, err := h.svc.ExecuteTask(r.Context(), orgID, taskID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Execution failed: "+err.Error(), "EXECUTION_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Task executed successfully", map[string]interface{}{
		"task":    task,
		"sidecar": sidecarResp,
	})
}

func (h *Handler) ExecuteWorkflowChain(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	taskID := chi.URLParam(r, "task_id")

	finalRoot, children, err := h.svc.ExecuteWorkflowChain(r.Context(), orgID, taskID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Workflow chain execution failed: "+err.Error(), "WORKFLOW_CHAIN_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Workflow chain executed successfully", map[string]interface{}{
		"root_task": finalRoot,
		"children":  children,
	})
}

func (h *Handler) DelegateTask(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	taskID := chi.URLParam(r, "task_id")

	var req DelegateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	childTask, err := h.svc.DelegateTask(r.Context(), orgID, taskID, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}
	utils.Success(w, http.StatusCreated, "Task delegated successfully", childTask)
}

func (h *Handler) InitiateHandoff(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	taskID := chi.URLParam(r, "task_id")

	var req InitiateHandoffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	handoff, err := h.svc.InitiateHandoff(r.Context(), orgID, taskID, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}
	utils.Success(w, http.StatusOK, "Task handoff recorded successfully", handoff)
}

func (h *Handler) ListHandoffs(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	taskID := chi.URLParam(r, "task_id")

	handoffs, err := h.svc.ListHandoffs(r.Context(), orgID, taskID)
	if errors.Is(err, ErrTaskNotFound) {
		utils.Error(w, http.StatusNotFound, "Task not found", "NOT_FOUND")
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Task handoffs retrieved successfully", handoffs)
}

func (h *Handler) AddContextItem(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	taskID := chi.URLParam(r, "task_id")

	var req AddContextItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	item, err := h.svc.AddContextItem(r.Context(), orgID, taskID, req)
	if errors.Is(err, ErrTaskNotFound) {
		utils.Error(w, http.StatusNotFound, "Task not found", "NOT_FOUND")
		return
	}
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}
	utils.Success(w, http.StatusCreated, "Context item added successfully", item)
}

func (h *Handler) ListContextItems(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	taskID := chi.URLParam(r, "task_id")

	items, err := h.svc.ListContextItems(r.Context(), orgID, taskID)
	if errors.Is(err, ErrTaskNotFound) {
		utils.Error(w, http.StatusNotFound, "Task not found", "NOT_FOUND")
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Context items retrieved successfully", items)
}

func (h *Handler) PostMessage(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	taskID := chi.URLParam(r, "task_id")

	var req PostMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	msg, err := h.svc.PostMessage(r.Context(), orgID, taskID, req)
	if errors.Is(err, ErrTaskNotFound) {
		utils.Error(w, http.StatusNotFound, "Task not found", "NOT_FOUND")
		return
	}
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}
	utils.Success(w, http.StatusCreated, "Message posted successfully", msg)
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	taskID := chi.URLParam(r, "task_id")

	messages, err := h.svc.ListMessages(r.Context(), orgID, taskID)
	if errors.Is(err, ErrTaskNotFound) {
		utils.Error(w, http.StatusNotFound, "Task not found", "NOT_FOUND")
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Messages retrieved successfully", messages)
}

// Phase 6.4: Collaborative Planning & Decision Making Handlers

func (h *Handler) CreateCollaborativePlan(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req CreateCollaborativePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	plan, err := h.svc.CreateCollaborativePlan(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		if errors.Is(err, ErrAgentDisabled) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "AGENT_DISABLED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusCreated, "Collaborative plan created successfully", plan)
}

func (h *Handler) GetCollaborativePlan(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "plan_id")
	if planID == "" {
		utils.Error(w, http.StatusBadRequest, "Plan ID is required", "BAD_REQUEST")
		return
	}

	plan, err := h.svc.GetCollaborativePlan(r.Context(), orgID, planID)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}
	utils.Success(w, http.StatusOK, "Collaborative plan retrieved successfully", plan)
}

func (h *Handler) ExecuteCollaborativePlan(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "plan_id")
	if planID == "" {
		utils.Error(w, http.StatusBadRequest, "Plan ID is required", "BAD_REQUEST")
		return
	}

	plan, err := h.svc.ExecuteCollaborativePlan(r.Context(), orgID, planID)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXECUTION_FAILED")
		return
	}
	utils.Success(w, http.StatusOK, "Collaborative plan executed successfully", plan)
}

func (h *Handler) AssessShipmentHealth(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	shipmentID := chi.URLParam(r, "shipment_id")
	if shipmentID == "" {
		utils.Error(w, http.StatusBadRequest, "Shipment ID is required", "BAD_REQUEST")
		return
	}

	var req AssessShipmentHealthRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
			return
		}
	}
	req.ShipmentID = shipmentID

	plan, err := h.svc.AssessShipmentHealth(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXECUTION_FAILED")
		return
	}
	utils.Success(w, http.StatusOK, "Shipment health assessed successfully", plan)
}

func (h *Handler) InvestigateShipmentException(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	shipmentID := chi.URLParam(r, "shipment_id")
	exceptionID := chi.URLParam(r, "exception_id")
	if shipmentID == "" || exceptionID == "" {
		utils.Error(w, http.StatusBadRequest, "Shipment ID and Exception ID are required", "BAD_REQUEST")
		return
	}

	var req InvestigateExceptionRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
			return
		}
	}
	req.ShipmentID = shipmentID
	req.ExceptionID = exceptionID

	plan, err := h.svc.InvestigateShipmentException(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXECUTION_FAILED")
		return
	}
	utils.Success(w, http.StatusOK, "Shipment exception investigated successfully", plan)
}

func (h *Handler) ReplanShipmentOperation(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "plan_id")
	if planID == "" {
		utils.Error(w, http.StatusBadRequest, "Plan ID is required", "BAD_REQUEST")
		return
	}

	var req ReplanOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}
	req.OriginalPlanID = planID

	plan, err := h.svc.ReplanShipmentOperation(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXECUTION_FAILED")
		return
	}
	utils.Success(w, http.StatusOK, "Shipment operation replanned successfully", plan)
}

func (h *Handler) HandleShipmentEvent(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req ShipmentEventWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	plan, err := h.svc.HandleShipmentEvent(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXECUTION_FAILED")
		return
	}
	utils.Success(w, http.StatusOK, "Shipment event workflow processed successfully", plan)
}

// Phase 6.6: Commercial HTTP Handlers

func (h *Handler) AssessCustomerRelationship(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	customerID := chi.URLParam(r, "customer_id")
	if customerID == "" {
		utils.Error(w, http.StatusBadRequest, "customer_id is required", "BAD_REQUEST")
		return
	}

	var req AssessCustomerRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	req.CustomerID = customerID

	plan, err := h.svc.AssessCustomerRelationship(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXECUTION_FAILED")
		return
	}
	utils.Success(w, http.StatusOK, "Customer relationship assessed successfully", plan)
}

func (h *Handler) EvaluateLeadIntelligence(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	leadID := chi.URLParam(r, "lead_id")
	if leadID == "" {
		utils.Error(w, http.StatusBadRequest, "lead_id is required", "BAD_REQUEST")
		return
	}

	var req EvaluateLeadRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	req.LeadID = leadID

	plan, err := h.svc.EvaluateLeadIntelligence(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXECUTION_FAILED")
		return
	}
	utils.Success(w, http.StatusOK, "Lead intelligence evaluated successfully", plan)
}

func (h *Handler) EvaluateRFQCommercialWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}
	rfqID := chi.URLParam(r, "rfq_id")
	if rfqID == "" {
		utils.Error(w, http.StatusBadRequest, "rfq_id is required", "BAD_REQUEST")
		return
	}

	var req EvaluateRFQRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	req.RFQID = rfqID

	plan, err := h.svc.EvaluateRFQCommercialWorkflow(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXECUTION_FAILED")
		return
	}
	utils.Success(w, http.StatusOK, "RFQ commercial workflow evaluated successfully", plan)
}

func (h *Handler) AssessInvoiceCollections(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req AssessCollectionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	plan, err := h.svc.AssessInvoiceCollections(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXECUTION_FAILED")
		return
	}
	utils.Success(w, http.StatusOK, "Collections assessment processed successfully", plan)
}

func (h *Handler) HandleCommercialEvent(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req CommercialEventWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	plan, err := h.svc.HandleCommercialEvent(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXECUTION_FAILED")
		return
	}
	utils.Success(w, http.StatusOK, "Commercial event workflow processed successfully", plan)
}

func (h *Handler) AssessContractRisk(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req AssessContractRiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	plan, err := h.svc.AssessContractRisk(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXECUTION_FAILED")
		return
	}
	utils.Success(w, http.StatusOK, "Contract risk assessed successfully", plan)
}

func (h *Handler) AssessComplianceRisk(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req AssessComplianceRiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	plan, err := h.svc.AssessComplianceRisk(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXECUTION_FAILED")
		return
	}
	utils.Success(w, http.StatusOK, "Compliance risk assessed successfully", plan)
}

func (h *Handler) AssessCrossModuleRisk(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req AssessCrossModuleRiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	plan, err := h.svc.AssessCrossModuleRisk(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXECUTION_FAILED")
		return
	}
	utils.Success(w, http.StatusOK, "Cross-module risk assessed successfully", plan)
}

func (h *Handler) ReassessCrossModuleRisk(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req ReassessRiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	plan, err := h.svc.ReassessCrossModuleRisk(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXECUTION_FAILED")
		return
	}
	utils.Success(w, http.StatusOK, "Cross-module risk reassessed successfully", plan)
}

// ----------------------------------------------------------------------
// Phase 6.8: Conflict Resolution, Memory & Learning Handlers
// ----------------------------------------------------------------------

func (h *Handler) ResolveConflict(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req ResolveConflictRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	conflict, err := h.svc.ResolveConflict(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Conflict resolved successfully", conflict)
}

func (h *Handler) RecordOutcome(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req RecordOutcomeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	outcome, err := h.svc.RecordOutcome(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusCreated, "Agent outcome recorded successfully", outcome)
}

func (h *Handler) ListOutcomes(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	outcomes, err := h.svc.ListOutcomes(r.Context(), orgID, entityType, entityID, limit)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Agent outcomes retrieved successfully", outcomes)
}

func (h *Handler) QueryMemory(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req QueryMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	resp, err := h.svc.QueryMemory(r.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Memory queried successfully", resp)
}

func (h *Handler) getUser(r *http.Request) (int64, string, error) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		return 0, "", errors.New("unauthorized: missing user context")
	}
	actor := userCtx.CognitoID
	if actor == "" {
		actor = strconv.FormatInt(userCtx.UserID, 10)
	}
	return userCtx.OrgID, actor, nil
}

func (h *Handler) GetCommandCenterOverview(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	overview, err := h.svc.GetCommandCenterOverview(r.Context(), orgID)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Workforce command center overview retrieved successfully", overview)
}

func (h *Handler) GetWorkforceHealth(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	health, err := h.svc.GetWorkforceHealth(r.Context(), orgID)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Workforce health retrieved successfully", health)
}

func (h *Handler) GetAgentWorkload(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	workload, err := h.svc.GetAgentWorkload(r.Context(), orgID)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Agent workload retrieved successfully", workload)
}

func (h *Handler) EmergencyStop(w http.ResponseWriter, r *http.Request) {
	orgID, actor, err := h.getUser(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req EmergencyStopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	status, err := h.svc.EmergencyStop(r.Context(), orgID, req, actor)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}
	utils.Success(w, http.StatusOK, "Emergency stop status updated successfully", status)
}

func (h *Handler) GetEmergencyStopStatus(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	status, err := h.svc.GetEmergencyStopStatus(r.Context(), orgID)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Emergency stop status retrieved successfully", status)
}

func (h *Handler) ListWorkforceApprovals(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	items, err := h.svc.ListWorkforceApprovals(r.Context(), orgID)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Workforce approvals retrieved successfully", items)
}

func (h *Handler) ListWorkforceEscalations(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	items, err := h.svc.ListWorkforceEscalations(r.Context(), orgID)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Workforce escalations retrieved successfully", items)
}

func (h *Handler) GetRecentActivity(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	activity, err := h.svc.GetRecentActivity(r.Context(), orgID, limit)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Recent activity retrieved successfully", activity)
}

func (h *Handler) EvaluateActionPolicy(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req struct {
		AgentID    string                 `json:"agent_id"`
		ActionType string                 `json:"action_type"`
		Payload    map[string]interface{} `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	decision := h.svc.EvaluateActionPolicy(r.Context(), orgID, req.AgentID, req.ActionType, req.Payload)
	utils.Success(w, http.StatusOK, "Action policy evaluated successfully", decision)
}

func (h *Handler) ControlAgent(w http.ResponseWriter, r *http.Request) {
	orgID, actor, err := h.getUser(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	agentID := chi.URLParam(r, "agent_id")
	if agentID == "" {
		utils.Error(w, http.StatusBadRequest, "agent_id path parameter is required", "BAD_REQUEST")
		return
	}

	var req AgentControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	agent, err := h.svc.ControlAgent(r.Context(), orgID, agentID, req, actor)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}
	utils.Success(w, http.StatusOK, "Agent control updated successfully", agent)
}

func (h *Handler) InspectWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "plan_id")
	if planID == "" {
		utils.Error(w, http.StatusBadRequest, "plan_id path parameter is required", "BAD_REQUEST")
		return
	}

	detail, err := h.svc.InspectWorkflow(r.Context(), orgID, planID)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}
	utils.Success(w, http.StatusOK, "Workflow inspected successfully", detail)
}

func (h *Handler) ControlWorkflow(w http.ResponseWriter, r *http.Request) {
	orgID, actor, err := h.getUser(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	planID := chi.URLParam(r, "plan_id")
	if planID == "" {
		utils.Error(w, http.StatusBadRequest, "plan_id path parameter is required", "BAD_REQUEST")
		return
	}

	var req WorkflowControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	plan, err := h.svc.ControlWorkflow(r.Context(), orgID, planID, req, actor)
	if err != nil {
		if errors.Is(err, ErrUnauthorizedTenant) {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}
	utils.Success(w, http.StatusOK, "Workflow control updated successfully", plan)
}
