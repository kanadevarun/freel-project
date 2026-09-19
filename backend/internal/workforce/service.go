package workforce

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/freel/backend/internal/actions"
	"github.com/freel/backend/internal/approvals"
	auditDomain "github.com/freel/backend/internal/audit/domain"
	auditSvc "github.com/freel/backend/internal/audit/service"
)

// In-memory emergency stop manager for application-level safety controls
type emergencyStopManager struct {
	mu    sync.RWMutex
	stops map[int64]*EmergencyStopStatus
}

func newEmergencyStopManager() *emergencyStopManager {
	return &emergencyStopManager{
		stops: make(map[int64]*EmergencyStopStatus),
	}
}

func (m *emergencyStopManager) getOrCreate(orgID int64) *EmergencyStopStatus {
	st, exists := m.stops[orgID]
	if !exists {
		st = &EmergencyStopStatus{
			OrgID:            orgID,
			WorkforceStopped: false,
			ActiveStops:      []SingleStop{},
		}
		m.stops[orgID] = st
	}
	return st
}

func (m *emergencyStopManager) IsWorkforceStopped(orgID int64) (bool, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st, exists := m.stops[orgID]
	if !exists {
		return false, ""
	}
	if st.WorkforceStopped {
		return true, st.Reason
	}
	return false, ""
}

func (m *emergencyStopManager) IsAgentStopped(orgID int64, agentID string) (bool, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st, exists := m.stops[orgID]
	if !exists {
		return false, ""
	}
	if st.WorkforceStopped {
		return true, st.Reason
	}
	for _, s := range st.ActiveStops {
		if s.Scope == EmergencyStopScopeAgent && s.Target == agentID {
			return true, s.Reason
		}
	}
	return false, ""
}

func (m *emergencyStopManager) IsWorkflowStopped(orgID int64, workflowID string) (bool, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st, exists := m.stops[orgID]
	if !exists {
		return false, ""
	}
	if st.WorkforceStopped {
		return true, st.Reason
	}
	for _, s := range st.ActiveStops {
		if s.Scope == EmergencyStopScopeWorkflow && s.Target == workflowID {
			return true, s.Reason
		}
	}
	return false, ""
}

func (m *emergencyStopManager) IsActionClassStopped(orgID int64, actionClass string) (bool, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st, exists := m.stops[orgID]
	if !exists {
		return false, ""
	}
	if st.WorkforceStopped {
		return true, st.Reason
	}
	for _, s := range st.ActiveStops {
		if s.Scope == EmergencyStopScopeActionClass && strings.EqualFold(s.Target, actionClass) {
			return true, s.Reason
		}
	}
	return false, ""
}

func (m *emergencyStopManager) SetStop(orgID int64, stop SingleStop) *EmergencyStopStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := m.getOrCreate(orgID)

	if stop.Scope == EmergencyStopScopeWorkforce {
		st.WorkforceStopped = true
		st.Reason = stop.Reason
		st.TriggeredBy = stop.TriggeredBy
		st.TriggeredAt = stop.TriggeredAt
	}

	filtered := make([]SingleStop, 0, len(st.ActiveStops)+1)
	for _, s := range st.ActiveStops {
		if !(s.Scope == stop.Scope && s.Target == stop.Target) {
			filtered = append(filtered, s)
		}
	}
	filtered = append(filtered, stop)
	st.ActiveStops = filtered
	return st
}

func (m *emergencyStopManager) RemoveStop(orgID int64, scope, target string) *EmergencyStopStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := m.getOrCreate(orgID)

	if scope == EmergencyStopScopeWorkforce {
		st.WorkforceStopped = false
		st.Reason = ""
	}

	filtered := make([]SingleStop, 0, len(st.ActiveStops))
	for _, s := range st.ActiveStops {
		if !(s.Scope == scope && (target == "" || s.Target == target)) {
			filtered = append(filtered, s)
		}
	}
	st.ActiveStops = filtered
	return st
}

func (m *emergencyStopManager) GetStatus(orgID int64) *EmergencyStopStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st, exists := m.stops[orgID]
	if !exists {
		return &EmergencyStopStatus{
			OrgID:            orgID,
			WorkforceStopped: false,
			ActiveStops:      []SingleStop{},
		}
	}
	copied := *st
	copiedStops := make([]SingleStop, len(st.ActiveStops))
	copy(copiedStops, st.ActiveStops)
	copied.ActiveStops = copiedStops
	return &copied
}

const (
	MaxDelegationDepth = 5
	MaxWorkflowTasks   = 20
)

var (
	ErrUnauthorizedTenant         = errors.New("unauthorized tenant access")
	ErrAgentDisabled              = errors.New("assigned agent is disabled")
	ErrMissingCapability          = errors.New("agent lacks required capability for task")
	ErrInvalidStatusTransition    = errors.New("invalid task status transition")
	ErrDelegationLoopDetected     = errors.New("delegation loop or cycle detected in agent chain")
	ErrMaxDelegationDepthExceeded = errors.New("maximum delegation depth exceeded")
	ErrMaxWorkflowTasksExceeded   = errors.New("maximum tasks per workflow exceeded")
	ErrDependencyPending          = errors.New("prerequisite task dependency is pending or incomplete")
	ErrDependencyFailed           = errors.New("prerequisite task dependency failed")
	ErrDuplicateDelegation        = errors.New("duplicate delegation request detected")
	ErrEmergencyStopActive        = errors.New("emergency stop is active for requested scope")
	ErrAgentPaused                = errors.New("agent is temporarily paused")
	ErrAutonomyRestricted         = errors.New("action restricted by agent autonomy level")
	ErrWorkflowStopped            = errors.New("workflow execution was stopped")
)

type Service interface {
	// Agent Registry
	ListAgents(ctx context.Context, orgID int64, agentType string, isEnabled *bool) ([]*WorkforceAgent, error)
	GetAgent(ctx context.Context, orgID int64, agentID string) (*WorkforceAgent, error)
	RegisterAgent(ctx context.Context, orgID int64, req RegisterAgentRequest) (*WorkforceAgent, error)
	UpdateAgent(ctx context.Context, orgID int64, agentID string, req UpdateAgentRequest) (*WorkforceAgent, error)

	// Tasks & Orchestration
	CreateTask(ctx context.Context, orgID int64, req CreateTaskRequest) (*WorkforceTask, error)
	GetTask(ctx context.Context, orgID int64, taskID string) (*WorkforceTask, error)
	ListTasks(ctx context.Context, filter TaskFilter) ([]*WorkforceTask, int, error)
	GetTaskHierarchy(ctx context.Context, orgID int64, rootOrTaskID string) (*TaskHierarchyNode, error)
	ExecuteTask(ctx context.Context, orgID int64, taskID string) (*WorkforceTask, *SidecarTaskResponse, error)
	DelegateTask(ctx context.Context, orgID int64, parentTaskID string, req DelegateTaskRequest) (*WorkforceTask, error)
	InitiateHandoff(ctx context.Context, orgID int64, taskID string, req InitiateHandoffRequest) (*WorkforceHandoff, error)
	ExecuteWorkflowChain(ctx context.Context, orgID int64, rootTaskID string) (*WorkforceTask, []*WorkforceTask, error)

	// Collaborative Planning & Decision Making (Phase 6.4)
	CreateCollaborativePlan(ctx context.Context, orgID int64, req CreateCollaborativePlanRequest) (*CollaborativePlan, error)
	ExecuteCollaborativePlan(ctx context.Context, orgID int64, planID string) (*CollaborativePlan, error)
	GetCollaborativePlan(ctx context.Context, orgID int64, planID string) (*CollaborativePlan, error)

	// Multi-Agent Shipment & Exception Operations (Phase 6.5)
	AssessShipmentHealth(ctx context.Context, orgID int64, req AssessShipmentHealthRequest) (*CollaborativePlan, error)
	InvestigateShipmentException(ctx context.Context, orgID int64, req InvestigateExceptionRequest) (*CollaborativePlan, error)
	ReplanShipmentOperation(ctx context.Context, orgID int64, req ReplanOperationRequest) (*CollaborativePlan, error)
	HandleShipmentEvent(ctx context.Context, orgID int64, req ShipmentEventWorkflowRequest) (*CollaborativePlan, error)

	// Multi-Agent Customer, Sales, Pricing & Finance (Phase 6.6)
	AssessCustomerRelationship(ctx context.Context, orgID int64, req AssessCustomerRequest) (*CollaborativePlan, error)
	EvaluateLeadIntelligence(ctx context.Context, orgID int64, req EvaluateLeadRequest) (*CollaborativePlan, error)
	EvaluateRFQCommercialWorkflow(ctx context.Context, orgID int64, req EvaluateRFQRequest) (*CollaborativePlan, error)
	AssessInvoiceCollections(ctx context.Context, orgID int64, req AssessCollectionsRequest) (*CollaborativePlan, error)
	HandleCommercialEvent(ctx context.Context, orgID int64, req CommercialEventWorkflowRequest) (*CollaborativePlan, error)

	// Multi-Agent Contract, Compliance & Risk (Phase 6.7)
	AssessContractRisk(ctx context.Context, orgID int64, req AssessContractRiskRequest) (*CollaborativePlan, error)
	AssessComplianceRisk(ctx context.Context, orgID int64, req AssessComplianceRiskRequest) (*CollaborativePlan, error)
	AssessCrossModuleRisk(ctx context.Context, orgID int64, req AssessCrossModuleRiskRequest) (*CollaborativePlan, error)
	ReassessCrossModuleRisk(ctx context.Context, orgID int64, req ReassessRiskRequest) (*CollaborativePlan, error)

	// Conflict Resolution, Memory & Learning (Phase 6.8)
	ResolveConflict(ctx context.Context, orgID int64, req ResolveConflictRequest) (*ConflictRecord, error)
	RecordOutcome(ctx context.Context, orgID int64, req RecordOutcomeRequest) (*AgentOutcome, error)
	ListOutcomes(ctx context.Context, orgID int64, entityType, entityID string, limit int) ([]AgentOutcome, error)
	QueryMemory(ctx context.Context, orgID int64, req QueryMemoryRequest) (*MemoryQueryResponse, error)

	// Governed Autonomy & Command Center (Phase 6.9)
	EvaluateActionPolicy(ctx context.Context, orgID int64, agentID, actionType string, payload map[string]interface{}) ActionPolicyDecision
	GetCommandCenterOverview(ctx context.Context, orgID int64) (*CommandCenterOverview, error)
	GetWorkforceHealth(ctx context.Context, orgID int64) (*WorkforceHealthSummary, error)
	GetAgentWorkload(ctx context.Context, orgID int64) ([]AgentWorkloadMetrics, error)
	ControlAgent(ctx context.Context, orgID int64, agentID string, req AgentControlRequest, actor string) (*WorkforceAgent, error)
	EmergencyStop(ctx context.Context, orgID int64, req EmergencyStopRequest, actor string) (*EmergencyStopStatus, error)
	GetEmergencyStopStatus(ctx context.Context, orgID int64) (*EmergencyStopStatus, error)
	ListWorkforceApprovals(ctx context.Context, orgID int64) ([]WorkforceApprovalItem, error)
	ListWorkforceEscalations(ctx context.Context, orgID int64) ([]WorkforceEscalationItem, error)
	GetRecentActivity(ctx context.Context, orgID int64, limit int) ([]WorkforceActivityItem, error)
	InspectWorkflow(ctx context.Context, orgID int64, planID string) (*WorkflowInspectionDetail, error)
	ControlWorkflow(ctx context.Context, orgID int64, planID string, req WorkflowControlRequest, actor string) (*CollaborativePlan, error)

	// Context & Messages
	AddContextItem(ctx context.Context, orgID int64, taskID string, req AddContextItemRequest) (*WorkforceContextItem, error)
	ListContextItems(ctx context.Context, orgID int64, taskID string) ([]*WorkforceContextItem, error)
	PostMessage(ctx context.Context, orgID int64, taskID string, req PostMessageRequest) (*WorkforceMessage, error)
	ListMessages(ctx context.Context, orgID int64, taskID string) ([]*WorkforceMessage, error)
	ListHandoffs(ctx context.Context, orgID int64, taskID string) ([]*WorkforceHandoff, error)
}

type workforceService struct {
	repo         Repository
	sidecar      SidecarClient
	actionsSvc   actions.Service
	approvalsSvc approvals.Service
	auditLogger  auditSvc.Service
	emergStops   *emergencyStopManager
}

func NewService(
	repo Repository,
	sidecar SidecarClient,
	actionsSvc actions.Service,
	approvalsSvc approvals.Service,
	auditLogger auditSvc.Service,
) Service {
	return &workforceService{
		repo:         repo,
		sidecar:      sidecar,
		actionsSvc:   actionsSvc,
		approvalsSvc: approvalsSvc,
		auditLogger:  auditLogger,
		emergStops:   newEmergencyStopManager(),
	}
}


// Helpers
func generateID(prefix string) string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b))
}

func (s *workforceService) ListAgents(ctx context.Context, orgID int64, agentType string, isEnabled *bool) ([]*WorkforceAgent, error) {
	return s.repo.ListAgents(ctx, orgID, agentType, isEnabled)
}

func (s *workforceService) GetAgent(ctx context.Context, orgID int64, agentID string) (*WorkforceAgent, error) {
	return s.repo.GetAgent(ctx, orgID, agentID)
}

func (s *workforceService) RegisterAgent(ctx context.Context, orgID int64, req RegisterAgentRequest) (*WorkforceAgent, error) {
	if req.AgentID == "" || req.Name == "" {
		return nil, errors.New("agent_id and name are required")
	}

	capsJSON, _ := json.Marshal(req.Capabilities)
	tasksJSON, _ := json.Marshal(req.AllowedTasks)
	entitiesJSON, _ := json.Marshal(req.AllowedEntities)

	autonomyLevel := req.AutonomyLevel
	if autonomyLevel == "" {
		autonomyLevel = "LEVEL_1_RECOMMEND"
	}
	ver := req.Version
	if ver == "" {
		ver = "1.0.0"
	}
	pVer := req.PromptVersion
	if pVer == "" {
		pVer = "1.0.0"
	}

	agent := &WorkforceAgent{
		OrgID:           orgID,
		AgentID:         req.AgentID,
		AgentType:       req.AgentType,
		Name:            req.Name,
		Description:     req.Description,
		Capabilities:    capsJSON,
		AllowedTasks:    tasksJSON,
		AllowedEntities: entitiesJSON,
		AutonomyLevel:   autonomyLevel,
		IsEnabled:       true,
		Version:         ver,
		PromptVersion:   pVer,
		HealthStatus:    "HEALTHY",
	}

	if err := s.repo.UpsertAgent(ctx, agent); err != nil {
		return nil, err
	}

	s.logAuditEvent(ctx, orgID, "WORKFORCE_AGENT_REGISTERED", "AGENT", agent.AgentID, map[string]interface{}{
		"agent_id":     agent.AgentID,
		"capabilities": req.Capabilities,
	})

	return s.GetAgent(ctx, orgID, req.AgentID)
}

func (s *workforceService) UpdateAgent(ctx context.Context, orgID int64, agentID string, req UpdateAgentRequest) (*WorkforceAgent, error) {
	agent, err := s.GetAgent(ctx, orgID, agentID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		agent.Name = *req.Name
	}
	if req.Description != nil {
		agent.Description = *req.Description
	}
	if req.Capabilities != nil {
		agent.Capabilities, _ = json.Marshal(req.Capabilities)
	}
	if req.AllowedTasks != nil {
		agent.AllowedTasks, _ = json.Marshal(req.AllowedTasks)
	}
	if req.AllowedEntities != nil {
		agent.AllowedEntities, _ = json.Marshal(req.AllowedEntities)
	}
	if req.AutonomyLevel != nil {
		agent.AutonomyLevel = *req.AutonomyLevel
	}
	if req.IsEnabled != nil {
		agent.IsEnabled = *req.IsEnabled
	}
	if req.HealthStatus != nil {
		agent.HealthStatus = *req.HealthStatus
	}

	// Always write under tenant orgID
	agent.OrgID = orgID
	if err := s.repo.UpsertAgent(ctx, agent); err != nil {
		return nil, err
	}

	return s.GetAgent(ctx, orgID, agentID)
}

func (s *workforceService) CreateTask(ctx context.Context, orgID int64, req CreateTaskRequest) (*WorkforceTask, error) {
	if req.Objective == "" {
		return nil, errors.New("objective is required")
	}
	if req.AssignedAgentID == "" {
		return nil, errors.New("assigned_agent_id is required")
	}

	// 1. Validate agent exists and is active
	agent, err := s.GetAgent(ctx, orgID, req.AssignedAgentID)
	if err != nil {
		return nil, fmt.Errorf("invalid assigned agent: %w", err)
	}
	if !agent.IsEnabled {
		return nil, ErrAgentDisabled
	}
	if stopped, reason := s.emergStops.IsWorkforceStopped(orgID); stopped {
		return nil, fmt.Errorf("%w: %s", ErrEmergencyStopActive, reason)
	}
	if stopped, reason := s.emergStops.IsAgentStopped(orgID, req.AssignedAgentID); stopped {
		return nil, fmt.Errorf("%w: agent %s stopped: %s", ErrEmergencyStopActive, req.AssignedAgentID, reason)
	}
	if agent.OperationalStatus == AgentStatusPaused || agent.HealthStatus == AgentStatusPaused {
		return nil, fmt.Errorf("%w: agent %s is currently paused", ErrAgentPaused, req.AssignedAgentID)
	}

	// 2. Capability verification: Agent must possess required capabilities
	if err := s.verifyCapabilities(agent, req.RequiredCapabilities); err != nil {
		return nil, err
	}

	taskID := generateID("wft")
	corrID := req.CorrelationID
	if corrID == "" {
		corrID = generateID("corr")
	}

	priority := req.Priority
	if priority == "" {
		priority = PriorityMedium
	}

	var ctxRefJSON json.RawMessage
	if req.ContextReference != nil {
		ctxRefJSON, _ = json.Marshal(req.ContextReference)
	}
	reqCapsJSON, _ := json.Marshal(req.RequiredCapabilities)
	depsJSON, _ := json.Marshal(req.Dependencies)

	task := &WorkforceTask{
		TaskID:               taskID,
		OrgID:                orgID,
		Objective:            req.Objective,
		InitiatingEvent:      &req.InitiatingEvent,
		AssignedAgentID:      req.AssignedAgentID,
		Status:               TaskStatusAssigned,
		Priority:             priority,
		ContextReference:     ctxRefJSON,
		RequiredCapabilities: reqCapsJSON,
		Dependencies:         depsJSON,
		CorrelationID:        corrID,
	}

	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, err
	}

	// Persist any initial context items
	for _, item := range req.InitialContextItems {
		_, _ = s.AddContextItem(ctx, orgID, taskID, item)
	}

	s.logAuditEvent(ctx, orgID, "WORKFORCE_TASK_CREATED", "WORKFORCE_TASK", taskID, map[string]interface{}{
		"assigned_agent": req.AssignedAgentID,
		"objective":      req.Objective,
		"correlation_id": corrID,
	})

	return s.GetTask(ctx, orgID, taskID)
}

func (s *workforceService) GetTask(ctx context.Context, orgID int64, taskID string) (*WorkforceTask, error) {
	return s.repo.GetTask(ctx, orgID, taskID)
}

func (s *workforceService) ListTasks(ctx context.Context, filter TaskFilter) ([]*WorkforceTask, int, error) {
	return s.repo.ListTasks(ctx, filter)
}

func (s *workforceService) GetTaskHierarchy(ctx context.Context, orgID int64, rootOrTaskID string) (*TaskHierarchyNode, error) {
	return s.repo.GetTaskHierarchy(ctx, orgID, rootOrTaskID)
}

func (s *workforceService) DelegateTask(ctx context.Context, orgID int64, parentTaskID string, req DelegateTaskRequest) (*WorkforceTask, error) {
	parent, err := s.GetTask(ctx, orgID, parentTaskID)
	if err != nil {
		return nil, fmt.Errorf("parent task not found: %w", err)
	}

	// 1. Idempotency check: if child task with same TargetAgentID and Objective exists under this parent, return it
	existingChildren, _, _ := s.repo.ListTasks(ctx, TaskFilter{
		OrgID:        orgID,
		ParentTaskID: &parent.TaskID,
	})
	for _, ch := range existingChildren {
		if ch.AssignedAgentID == req.TargetAgentID && ch.Objective == req.Objective {
			return ch, nil
		}
	}

	// 2. Loop & cycle safeguards enforcement
	if err := s.checkDelegationSafeguards(ctx, orgID, parent, req.TargetAgentID); err != nil {
		errPayload, _ := json.Marshal(map[string]interface{}{
			"error":        err.Error(),
			"target_agent": req.TargetAgentID,
			"objective":    req.Objective,
		})
		_ = s.repo.CreateMessage(ctx, &WorkforceMessage{
			MessageID:        generateID("msg"),
			OrgID:            orgID,
			TaskID:           parent.TaskID,
			ParentTaskID:     parent.ParentTaskID,
			SenderAgentID:    "system_guard",
			RecipientAgentID: parent.AssignedAgentID,
			MessageType:      MessageTypeError,
			Objective:        req.Objective,
			Payload:          errPayload,
			Priority:         PriorityHigh,
			CorrelationID:    parent.CorrelationID,
		})
		s.logAuditEvent(ctx, orgID, "WORKFORCE_DELEGATION_REJECTED", "WORKFORCE_TASK", parent.TaskID, map[string]interface{}{
			"reason":       err.Error(),
			"target_agent": req.TargetAgentID,
		})
		return nil, err
	}

	targetAgent, err := s.GetAgent(ctx, orgID, req.TargetAgentID)
	if err != nil {
		return nil, fmt.Errorf("target agent not found: %w", err)
	}
	if !targetAgent.IsEnabled {
		return nil, ErrAgentDisabled
	}
	if stopped, reason := s.emergStops.IsWorkforceStopped(orgID); stopped {
		return nil, fmt.Errorf("%w: %s", ErrEmergencyStopActive, reason)
	}
	if stopped, reason := s.emergStops.IsAgentStopped(orgID, req.TargetAgentID); stopped {
		return nil, fmt.Errorf("%w: target agent %s stopped: %s", ErrEmergencyStopActive, req.TargetAgentID, reason)
	}
	if targetAgent.OperationalStatus == AgentStatusPaused || targetAgent.HealthStatus == AgentStatusPaused {
		return nil, fmt.Errorf("%w: target agent %s is currently paused", ErrAgentPaused, req.TargetAgentID)
	}

	// 3. Capability verification: target agent must have requested capabilities
	if err := s.verifyCapabilities(targetAgent, req.RequiredCapabilities); err != nil {
		return nil, err
	}

	childTaskID := generateID("wft")
	rootTaskID := parent.RootTaskID
	if rootTaskID == nil || *rootTaskID == "" {
		rootTaskID = &parent.TaskID
	}

	priority := req.Priority
	if priority == "" {
		priority = parent.Priority
	}

	var ctxRefJSON json.RawMessage
	if req.ContextReference != nil {
		ctxRefJSON, _ = json.Marshal(req.ContextReference)
	}
	reqCapsJSON, _ := json.Marshal(req.RequiredCapabilities)
	depsJSON, _ := json.Marshal(req.Dependencies)

	childTask := &WorkforceTask{
		TaskID:               childTaskID,
		OrgID:                orgID,
		Objective:            req.Objective,
		ParentTaskID:         &parent.TaskID,
		RootTaskID:           rootTaskID,
		AssignedAgentID:      req.TargetAgentID,
		Status:               TaskStatusAssigned,
		Priority:             priority,
		ContextReference:     ctxRefJSON,
		RequiredCapabilities: reqCapsJSON,
		Dependencies:         depsJSON,
		CorrelationID:        parent.CorrelationID,
	}

	if err := s.repo.CreateTask(ctx, childTask); err != nil {
		return nil, fmt.Errorf("failed creating delegated child task: %w", err)
	}

	// 4. Context Minimization: inherit only specified context references or relevant facts
	if len(req.ContextReferences) > 0 {
		parentContexts, _ := s.repo.ListContextItems(ctx, orgID, parent.TaskID)
		refMap := make(map[string]bool)
		for _, ref := range req.ContextReferences {
			refMap[ref] = true
		}
		for _, pc := range parentContexts {
			if refMap[pc.ContextID] {
				var contentMap map[string]interface{}
				_ = json.Unmarshal(pc.Content, &contentMap)
				eType := ""
				if pc.EntityType != nil {
					eType = *pc.EntityType
				}
				eID := ""
				if pc.EntityID != nil {
					eID = *pc.EntityID
				}
				provID := ""
				if pc.ProvenanceID != nil {
					provID = *pc.ProvenanceID
				}
				sAgent := ""
				if pc.SourceAgentID != nil {
					sAgent = *pc.SourceAgentID
				}
				_, _ = s.AddContextItem(ctx, orgID, childTaskID, AddContextItemRequest{
					ItemType:      pc.ItemType,
					EntityType:    eType,
					EntityID:      eID,
					SourceAgentID: sAgent,
					ProvenanceID:  provID,
					Content:       contentMap,
					Confidence:    pc.Confidence,
				})
			}
		}
	} else {
		// Minimization fallback: propagate only verified FACTS from parent
		parentContexts, _ := s.repo.ListContextItems(ctx, orgID, parent.TaskID)
		for _, pc := range parentContexts {
			if pc.ItemType == ContextTypeFact {
				var contentMap map[string]interface{}
				_ = json.Unmarshal(pc.Content, &contentMap)
				eType := ""
				if pc.EntityType != nil {
					eType = *pc.EntityType
				}
				eID := ""
				if pc.EntityID != nil {
					eID = *pc.EntityID
				}
				_, _ = s.AddContextItem(ctx, orgID, childTaskID, AddContextItemRequest{
					ItemType:   pc.ItemType,
					EntityType: eType,
					EntityID:   eID,
					Content:    contentMap,
					Confidence: pc.Confidence,
				})
			}
		}
	}

	// 5. Update parent status to WAITING
	_ = s.repo.UpdateTaskStatus(ctx, orgID, parent.TaskID, TaskStatusWaiting, nil, 0, nil, nil)

	// 6. Emit structured DELEGATION message
	msgPayload, _ := json.Marshal(map[string]interface{}{
		"delegated_to":    req.TargetAgentID,
		"expected_output": req.ExpectedOutput,
		"child_task_id":   childTaskID,
	})
	delegationMsg := &WorkforceMessage{
		MessageID:        generateID("msg"),
		OrgID:            orgID,
		TaskID:           parent.TaskID,
		ParentTaskID:     &parent.TaskID,
		SenderAgentID:    parent.AssignedAgentID,
		RecipientAgentID: req.TargetAgentID,
		MessageType:      MessageTypeDelegation,
		Objective:        req.Objective,
		Payload:          msgPayload,
		Priority:         priority,
		CorrelationID:    parent.CorrelationID,
	}
	_ = s.repo.CreateMessage(ctx, delegationMsg)

	s.logAuditEvent(ctx, orgID, "WORKFORCE_TASK_DELEGATED", "WORKFORCE_TASK", childTaskID, map[string]interface{}{
		"parent_task_id": parent.TaskID,
		"source_agent":   parent.AssignedAgentID,
		"target_agent":   req.TargetAgentID,
		"objective":      req.Objective,
	})

	return s.GetTask(ctx, orgID, childTaskID)
}

func (s *workforceService) ExecuteTask(ctx context.Context, orgID int64, taskID string) (*WorkforceTask, *SidecarTaskResponse, error) {
	task, err := s.GetTask(ctx, orgID, taskID)
	if err != nil {
		return nil, nil, err
	}

	// 1. Check prerequisite task dependencies
	if err := s.checkDependencies(ctx, orgID, task); err != nil {
		updatedTask, _ := s.GetTask(ctx, orgID, taskID)
		return updatedTask, nil, err
	}

	// 2. Set task to RUNNING
	now := time.Now().UTC()
	task.StartedAt = &now
	_ = s.repo.UpdateTaskStatus(ctx, orgID, taskID, TaskStatusRunning, nil, 0, nil, nil)

	// 3. Fetch shared contexts for task
	contexts, err := s.repo.ListContextItems(ctx, orgID, taskID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed reading contexts for task: %w", err)
	}

	sidecarContexts := make([]SidecarContextReference, 0, len(contexts))
	for _, c := range contexts {
		var contentMap map[string]interface{}
		_ = json.Unmarshal(c.Content, &contentMap)

		eType := ""
		if c.EntityType != nil {
			eType = *c.EntityType
		}
		eID := ""
		if c.EntityID != nil {
			eID = *c.EntityID
		}
		sAgent := ""
		if c.SourceAgentID != nil {
			sAgent = *c.SourceAgentID
		}
		provID := ""
		if c.ProvenanceID != nil {
			provID = *c.ProvenanceID
		}

		sidecarContexts = append(sidecarContexts, SidecarContextReference{
			ContextID:       c.ContextID,
			ItemType:        string(c.ItemType),
			EntityType:      eType,
			EntityID:        eID,
			SourceAgentID:   sAgent,
			ProvenanceID:    provID,
			Content:         contentMap,
			Confidence:      c.Confidence,
			IsAuthoritative: c.IsAuthoritative,
		})
	}

	reqCaps := []string{}
	if len(task.RequiredCapabilities) > 0 {
		_ = json.Unmarshal(task.RequiredCapabilities, &reqCaps)
	}
	if reqCaps == nil {
		reqCaps = []string{}
	}

	deps := []string{}
	if len(task.Dependencies) > 0 {
		_ = json.Unmarshal(task.Dependencies, &deps)
	}
	if deps == nil {
		deps = []string{}
	}

	pTaskID := ""
	if task.ParentTaskID != nil {
		pTaskID = *task.ParentTaskID
	}
	rTaskID := ""
	if task.RootTaskID != nil {
		rTaskID = *task.RootTaskID
	}
	initEvent := ""
	if task.InitiatingEvent != nil {
		initEvent = *task.InitiatingEvent
	}

	sidecarReq := &SidecarTaskRequest{
		TaskID:               task.TaskID,
		OrgID:                orgID,
		Objective:            task.Objective,
		InitiatingEvent:      initEvent,
		ParentTaskID:         pTaskID,
		RootTaskID:           rTaskID,
		AssignedAgentID:      task.AssignedAgentID,
		Status:               string(task.Status),
		Priority:             string(task.Priority),
		RequiredCapabilities: reqCaps,
		ContextReferences:    sidecarContexts,
		Dependencies:         deps,
		CorrelationID:        task.CorrelationID,
	}

	// 4. Call AI sidecar
	sidecarResp, err := s.sidecar.ExecuteTask(ctx, sidecarReq)
	if err != nil {
		errStr := err.Error()
		errCode := "SIDECAR_EXECUTION_FAILED"
		_ = s.repo.UpdateTaskStatus(ctx, orgID, taskID, TaskStatusFailed, nil, 0, &errCode, &errStr)
		return nil, nil, err
	}

	// 5. Save new context items returned by Python AI
	for _, rawItem := range sidecarResp.NewContextItems {
		itemTypeStr, _ := rawItem["item_type"].(string)
		if itemTypeStr == "" {
			itemTypeStr = string(ContextTypeAgentResult)
		}
		content, _ := rawItem["content"].(map[string]interface{})
		if content == nil {
			content = rawItem
		}
		conf, _ := rawItem["confidence"].(float64)
		if conf <= 0 {
			conf = sidecarResp.Confidence
		}

		_, _ = s.AddContextItem(ctx, orgID, taskID, AddContextItemRequest{
			ItemType:      ContextItemType(itemTypeStr),
			SourceAgentID: task.AssignedAgentID,
			Content:       content,
			Confidence:    conf,
		})
	}

	// 6. Save structured inter-agent messages
	for _, msg := range sidecarResp.Messages {
		payloadJSON, _ := json.Marshal(msg.Payload)
		_ = s.repo.CreateMessage(ctx, &WorkforceMessage{
			MessageID:           msg.MessageID,
			OrgID:               orgID,
			TaskID:              task.TaskID,
			ParentTaskID:        task.ParentTaskID,
			SenderAgentID:       msg.SenderAgentID,
			RecipientAgentID:    msg.RecipientAgentID,
			MessageType:         MessageType(msg.MessageType),
			Objective:           msg.Objective,
			RequestedCapability: &msg.RequestedCapability,
			Payload:             payloadJSON,
			Priority:            task.Priority,
			CorrelationID:       task.CorrelationID,
		})
	}

	// 7. Handle proposed delegations automatically if any
	hasDelegations := len(sidecarResp.ProposedDelegations) > 0
	for _, del := range sidecarResp.ProposedDelegations {
		_, _ = s.DelegateTask(ctx, orgID, taskID, DelegateTaskRequest{
			TargetAgentID:        del.TargetAgentID,
			Objective:            del.Objective,
			RequiredCapabilities: del.RequiredCapabilities,
			ExpectedOutput:       del.ExpectedOutput,
			Priority:             TaskPriority(del.Priority),
		})
	}

	// 8. Handle proposed handoff if any
	if sidecarResp.ProposedHandoff != nil {
		h := sidecarResp.ProposedHandoff
		_, _ = s.InitiateHandoff(ctx, orgID, taskID, InitiateHandoffRequest{
			DestinationAgentID: h.DestinationAgentID,
			Reason:             h.Reason,
			Objective:          h.Objective,
			RequiredCapability: h.RequiredCapability,
			CurrentFindings:    h.CurrentFindings,
			ExpectedOutput:     h.ExpectedOutput,
			Confidence:         h.Confidence,
		})
	}

	// 9. Determine final task status
	finalStatus := TaskStatusCompleted
	if hasDelegations {
		finalStatus = TaskStatusWaiting // Parent waits for child subtasks
	}

	resJSON, _ := json.Marshal(sidecarResp.Findings)
	_ = s.repo.UpdateTaskStatus(ctx, orgID, taskID, finalStatus, resJSON, sidecarResp.Confidence, nil, nil)

	// 10. Result Return Flow: If this task is a child task, return structured RESULT to parent
	if task.ParentTaskID != nil && *task.ParentTaskID != "" {
		parentTask, errParent := s.GetTask(ctx, orgID, *task.ParentTaskID)
		if errParent == nil && parentTask != nil {
			resultPayload, _ := json.Marshal(map[string]interface{}{
				"child_task_id":   task.TaskID,
				"assigned_agent":  task.AssignedAgentID,
				"objective":       task.Objective,
				"confidence":      sidecarResp.Confidence,
				"summary":         sidecarResp.Summary,
				"findings":        sidecarResp.Findings,
				"facts":           sidecarResp.Facts,
				"predictions":     sidecarResp.Predictions,
				"recommendations": sidecarResp.Recommendations,
			})
			_ = s.repo.CreateMessage(ctx, &WorkforceMessage{
				MessageID:        generateID("msg"),
				OrgID:            orgID,
				TaskID:           parentTask.TaskID,
				ParentTaskID:     parentTask.ParentTaskID,
				SenderAgentID:    task.AssignedAgentID,
				RecipientAgentID: parentTask.AssignedAgentID,
				MessageType:      MessageTypeResult,
				Objective:        task.Objective,
				Payload:          resultPayload,
				Priority:         task.Priority,
				CorrelationID:    task.CorrelationID,
			})

			// Add structured AGENT_RESULT context item to parent task
			contentMap := map[string]interface{}{
				"child_task_id":   task.TaskID,
				"agent_id":        task.AssignedAgentID,
				"objective":       task.Objective,
				"summary":         sidecarResp.Summary,
				"findings":        sidecarResp.Findings,
				"recommendations": sidecarResp.Recommendations,
				"confidence":      sidecarResp.Confidence,
			}
			if len(sidecarResp.RecoveryOptions) > 0 {
				contentMap["recovery_options"] = sidecarResp.RecoveryOptions
			}
			_, _ = s.AddContextItem(ctx, orgID, parentTask.TaskID, AddContextItemRequest{
				ItemType:      ContextTypeAgentResult,
				SourceAgentID: task.AssignedAgentID,
				Content:       contentMap,
				Confidence:    sidecarResp.Confidence,
			})
		}
		if finalStatus == TaskStatusCompleted {
			s.checkParentResume(ctx, orgID, *task.ParentTaskID)
		}
	}

	s.logAuditEvent(ctx, orgID, "WORKFORCE_TASK_EXECUTED", "WORKFORCE_TASK", taskID, map[string]interface{}{
		"status":          finalStatus,
		"confidence":      sidecarResp.Confidence,
		"summary":         sidecarResp.Summary,
		"delegated_count": len(sidecarResp.ProposedDelegations),
	})

	updatedTask, _ := s.GetTask(ctx, orgID, taskID)
	return updatedTask, sidecarResp, nil
}

func (s *workforceService) checkParentResume(ctx context.Context, orgID int64, parentTaskID string) {
	children, _, err := s.repo.ListTasks(ctx, TaskFilter{
		OrgID:        orgID,
		ParentTaskID: &parentTaskID,
	})
	if err != nil {
		return
	}

	allDone := true
	hasFailure := false
	var failedTaskID string
	for _, ch := range children {
		if ch.Status == TaskStatusFailed {
			hasFailure = true
			failedTaskID = ch.TaskID
		}
		if ch.Status != TaskStatusCompleted && ch.Status != TaskStatusCancelled && ch.Status != TaskStatusFailed {
			allDone = false
			break
		}
	}

	if allDone {
		if hasFailure {
			errMsg := fmt.Sprintf("Child task %s failed during execution", failedTaskID)
			errCode := "CHILD_TASK_FAILED"
			_ = s.repo.UpdateTaskStatus(ctx, orgID, parentTaskID, TaskStatusEscalated, nil, 0, &errCode, &errMsg)
			s.logAuditEvent(ctx, orgID, "WORKFORCE_PARENT_TASK_ESCALATED", "WORKFORCE_TASK", parentTaskID, map[string]interface{}{
				"reason":         errMsg,
				"failed_task_id": failedTaskID,
			})
		} else {
			// All completed successfully
			_ = s.repo.UpdateTaskStatus(ctx, orgID, parentTaskID, TaskStatusCompleted, nil, 0, nil, nil)
			s.logAuditEvent(ctx, orgID, "WORKFORCE_PARENT_TASK_COMPLETED", "WORKFORCE_TASK", parentTaskID, map[string]interface{}{
				"reason": "All delegated child tasks have completed.",
			})
		}
	}
}

func (s *workforceService) InitiateHandoff(ctx context.Context, orgID int64, taskID string, req InitiateHandoffRequest) (*WorkforceHandoff, error) {
	task, err := s.GetTask(ctx, orgID, taskID)
	if err != nil {
		return nil, err
	}

	// Ping-pong handoff loop detection: check if there is an existing handoff on this task from dest to source
	existingHandoffs, _ := s.repo.ListHandoffs(ctx, orgID, taskID)
	for _, h := range existingHandoffs {
		if h.SourceAgentID == req.DestinationAgentID && h.DestinationAgentID == task.AssignedAgentID {
			return nil, fmt.Errorf("%w: repeated handoff ping-pong between %s and %s is prohibited", ErrDelegationLoopDetected, task.AssignedAgentID, req.DestinationAgentID)
		}
	}

	destAgent, err := s.GetAgent(ctx, orgID, req.DestinationAgentID)
	if err != nil {
		return nil, fmt.Errorf("destination agent not found: %w", err)
	}
	if !destAgent.IsEnabled {
		return nil, ErrAgentDisabled
	}

	// Verify capability
	if req.RequiredCapability != "" {
		if err := s.verifyCapabilities(destAgent, []string{req.RequiredCapability}); err != nil {
			return nil, err
		}
	}

	handoffID := generateID("hnd")
	findingsJSON, _ := json.Marshal(req.CurrentFindings)

	handoff := &WorkforceHandoff{
		HandoffID:          handoffID,
		OrgID:              orgID,
		TaskID:             taskID,
		OriginatingTaskID:  taskID,
		SourceAgentID:      task.AssignedAgentID,
		DestinationAgentID: req.DestinationAgentID,
		Reason:             req.Reason,
		Objective:          req.Objective,
		RequiredCapability: req.RequiredCapability,
		CurrentFindings:    findingsJSON,
		ExpectedOutput:     req.ExpectedOutput,
		Confidence:         req.Confidence,
		Status:             HandoffStatusInitiated,
	}

	if err := s.repo.CreateHandoff(ctx, handoff); err != nil {
		return nil, fmt.Errorf("failed creating workforce handoff: %w", err)
	}

	// Reassign task to destination agent
	task.AssignedAgentID = req.DestinationAgentID
	task.Status = TaskStatusAssigned
	_ = s.repo.UpdateTaskStatus(ctx, orgID, taskID, TaskStatusAssigned, nil, 0, nil, nil)

	// Emit structured HANDOFF message
	msgPayload, _ := json.Marshal(map[string]interface{}{
		"handoff_id":       handoffID,
		"reason":           req.Reason,
		"current_findings": req.CurrentFindings,
		"expected_output":  req.ExpectedOutput,
	})
	handoffMsg := &WorkforceMessage{
		MessageID:           generateID("msg"),
		OrgID:               orgID,
		TaskID:              taskID,
		ParentTaskID:        task.ParentTaskID,
		SenderAgentID:       handoff.SourceAgentID,
		RecipientAgentID:    handoff.DestinationAgentID,
		MessageType:         MessageTypeHandoff,
		Objective:           req.Objective,
		RequestedCapability: &req.RequiredCapability,
		Payload:             msgPayload,
		Priority:            task.Priority,
		CorrelationID:       task.CorrelationID,
	}
	_ = s.repo.CreateMessage(ctx, handoffMsg)

	s.logAuditEvent(ctx, orgID, "WORKFORCE_HANDOFF_INITIATED", "WORKFORCE_HANDOFF", handoffID, map[string]interface{}{
		"task_id":           taskID,
		"source_agent":      handoff.SourceAgentID,
		"destination_agent": handoff.DestinationAgentID,
		"reason":            req.Reason,
	})

	return handoff, nil
}

func (s *workforceService) AddContextItem(ctx context.Context, orgID int64, taskID string, req AddContextItemRequest) (*WorkforceContextItem, error) {
	// Verify task belongs to tenant
	_, err := s.GetTask(ctx, orgID, taskID)
	if err != nil {
		return nil, err
	}

	contentJSON, err := json.Marshal(req.Content)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling context item content: %w", err)
	}

	// CRITICAL SECURITY ENFORCEMENT:
	// Only FACT, HUMAN_DECISION, or validated SYSTEM_EVENT can be authoritative!
	// Predictions and recommendations can NEVER silently become authoritative business facts.
	isAuth := false
	if req.ItemType == ContextTypeFact || req.ItemType == ContextTypeHumanDecision {
		isAuth = true
	}

	conf := req.Confidence
	if conf <= 0 {
		if isAuth {
			conf = 1.0
		} else {
			conf = 0.85
		}
	}

	ctxID := generateID("ctx")
	item := &WorkforceContextItem{
		ContextID:       ctxID,
		OrgID:           orgID,
		TaskID:          taskID,
		ItemType:        req.ItemType,
		EntityType:      &req.EntityType,
		EntityID:        &req.EntityID,
		SourceAgentID:   &req.SourceAgentID,
		ProvenanceID:    &req.ProvenanceID,
		Content:         contentJSON,
		Confidence:      conf,
		IsAuthoritative: isAuth,
	}

	if err := s.repo.CreateContextItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed storing context item: %w", err)
	}
	return item, nil
}

func (s *workforceService) ListContextItems(ctx context.Context, orgID int64, taskID string) ([]*WorkforceContextItem, error) {
	// Verify task belongs to tenant
	if _, err := s.GetTask(ctx, orgID, taskID); err != nil {
		return nil, err
	}
	return s.repo.ListContextItems(ctx, orgID, taskID)
}

func (s *workforceService) PostMessage(ctx context.Context, orgID int64, taskID string, req PostMessageRequest) (*WorkforceMessage, error) {
	task, err := s.GetTask(ctx, orgID, taskID)
	if err != nil {
		return nil, err
	}

	payloadJSON, _ := json.Marshal(req.Payload)
	priority := req.Priority
	if priority == "" {
		priority = task.Priority
	}

	msg := &WorkforceMessage{
		MessageID:           generateID("msg"),
		OrgID:               orgID,
		TaskID:              taskID,
		ParentTaskID:        task.ParentTaskID,
		SenderAgentID:       task.AssignedAgentID,
		RecipientAgentID:    req.RecipientAgentID,
		MessageType:         req.MessageType,
		Objective:           req.Objective,
		RequestedCapability: &req.RequestedCapability,
		Payload:             payloadJSON,
		Priority:            priority,
		CorrelationID:       task.CorrelationID,
	}

	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *workforceService) ListMessages(ctx context.Context, orgID int64, taskID string) ([]*WorkforceMessage, error) {
	if _, err := s.GetTask(ctx, orgID, taskID); err != nil {
		return nil, err
	}
	return s.repo.ListMessages(ctx, orgID, taskID)
}

func (s *workforceService) ListHandoffs(ctx context.Context, orgID int64, taskID string) ([]*WorkforceHandoff, error) {
	if _, err := s.GetTask(ctx, orgID, taskID); err != nil {
		return nil, err
	}
	return s.repo.ListHandoffs(ctx, orgID, taskID)
}

func (s *workforceService) verifyCapabilities(agent *WorkforceAgent, requiredCaps []string) error {
	if len(requiredCaps) == 0 {
		return nil
	}

	var agentCaps []string
	if len(agent.Capabilities) > 0 {
		_ = json.Unmarshal(agent.Capabilities, &agentCaps)
	}

	capMap := make(map[string]bool)
	for _, c := range agentCaps {
		capMap[c] = true
	}

	for _, req := range requiredCaps {
		if !capMap[req] {
			return fmt.Errorf("%w: agent %s lacks '%s'", ErrMissingCapability, agent.AgentID, req)
		}
	}
	return nil
}

func (s *workforceService) logAuditEvent(ctx context.Context, orgID int64, action, entityType, entityID string, details map[string]interface{}) {
	if s.auditLogger == nil {
		return
	}
	s.auditLogger.RecordAsync(ctx, auditDomain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorType:    auditDomain.ActorTypeAIAgent,
		ActorName:    "Multi-Agent Workforce Foundation",
		Action:       action,
		Module:       "WORKFORCE",
		ResourceType: entityType,
		ResourceID:   entityID,
		Metadata:     details,
		Result:       auditDomain.ResultSuccess,
	})
}

func (s *workforceService) checkDelegationSafeguards(ctx context.Context, orgID int64, parent *WorkforceTask, targetAgentID string) error {
	// 1. Direct self-delegation detection
	if parent.AssignedAgentID == targetAgentID {
		return fmt.Errorf("%w: agent %s cannot delegate directly to itself", ErrDelegationLoopDetected, targetAgentID)
	}

	// 2. Walk ancestor chain to calculate depth and detect cycles
	depth := 1
	currParentID := parent.ParentTaskID
	for currParentID != nil && *currParentID != "" {
		depth++
		if depth > MaxDelegationDepth {
			return fmt.Errorf("%w: current depth %d exceeds maximum allowed depth %d", ErrMaxDelegationDepthExceeded, depth, MaxDelegationDepth)
		}
		ancestor, err := s.GetTask(ctx, orgID, *currParentID)
		if err != nil {
			break
		}
		if ancestor.AssignedAgentID == targetAgentID {
			return fmt.Errorf("%w: agent %s is already in ancestor delegation chain", ErrDelegationLoopDetected, targetAgentID)
		}
		currParentID = ancestor.ParentTaskID
	}

	// 3. Check total tasks for this workflow / correlation_id
	if parent.CorrelationID != "" {
		_, count, err := s.repo.ListTasks(ctx, TaskFilter{
			OrgID:         orgID,
			CorrelationID: parent.CorrelationID,
			Limit:         100,
		})
		if err == nil && count >= MaxWorkflowTasks {
			return fmt.Errorf("%w: total tasks %d reaches workflow limit %d", ErrMaxWorkflowTasksExceeded, count, MaxWorkflowTasks)
		}
	}

	return nil
}

func (s *workforceService) checkDependencies(ctx context.Context, orgID int64, task *WorkforceTask) error {
	if len(task.Dependencies) == 0 {
		return nil
	}

	var depIDs []string
	if err := json.Unmarshal(task.Dependencies, &depIDs); err != nil || len(depIDs) == 0 {
		return nil
	}

	for _, depID := range depIDs {
		depTask, err := s.GetTask(ctx, orgID, depID)
		if err != nil {
			return fmt.Errorf("%w: unable to find dependency task %s", ErrDependencyFailed, depID)
		}
		if depTask.Status == TaskStatusFailed || depTask.Status == TaskStatusCancelled {
			errCode := "DEPENDENCY_FAILED"
			errMsg := fmt.Sprintf("Dependency task %s is in state %s", depID, depTask.Status)
			_ = s.repo.UpdateTaskStatus(ctx, orgID, task.TaskID, TaskStatusBlocked, nil, 0, &errCode, &errMsg)
			return fmt.Errorf("%w: dependency task %s %s", ErrDependencyFailed, depID, depTask.Status)
		}
		if depTask.Status != TaskStatusCompleted {
			errCode := "DEPENDENCY_PENDING"
			errMsg := fmt.Sprintf("Waiting on dependency task %s (status: %s)", depID, depTask.Status)
			_ = s.repo.UpdateTaskStatus(ctx, orgID, task.TaskID, TaskStatusBlocked, nil, 0, &errCode, &errMsg)
			return fmt.Errorf("%w: dependency task %s is %s", ErrDependencyPending, depID, depTask.Status)
		}
	}
	return nil
}

func (s *workforceService) ExecuteWorkflowChain(ctx context.Context, orgID int64, rootTaskID string) (*WorkforceTask, []*WorkforceTask, error) {
	// 1. Execute root task (e.g. Planning Agent) to generate plan and delegations
	rootTask, _, err := s.ExecuteTask(ctx, orgID, rootTaskID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed executing root workflow task: %w", err)
	}

	// 2. Fetch all child tasks delegated by root task
	childTasks, _, err := s.repo.ListTasks(ctx, TaskFilter{
		OrgID:        orgID,
		ParentTaskID: &rootTask.TaskID,
	})
	if err != nil {
		return rootTask, nil, fmt.Errorf("failed listing delegated child tasks: %w", err)
	}

	executedChildren := make([]*WorkforceTask, 0, len(childTasks))

	// 3. Execute child tasks sequentially respecting dependencies
	for _, ch := range childTasks {
		executedChild, _, err := s.ExecuteTask(ctx, orgID, ch.TaskID)
		if err != nil {
			s.logAuditEvent(ctx, orgID, "WORKFORCE_CHAIN_CHILD_FAILED", "WORKFORCE_TASK", ch.TaskID, map[string]interface{}{
				"error": err.Error(),
			})
		}
		if executedChild != nil {
			executedChildren = append(executedChildren, executedChild)
		}
	}

	// 4. Re-execute root task (Planning Agent) for final consolidation
	// Specialist findings have been returned to rootTask's context items via Result Return Flow!
	finalRootTask, _, err := s.ExecuteTask(ctx, orgID, rootTaskID)
	if err != nil {
		return rootTask, executedChildren, nil
	}

	return finalRootTask, executedChildren, nil
}

// Phase 6.4: Collaborative Planning & Decision Making

func (s *workforceService) CreateCollaborativePlan(ctx context.Context, orgID int64, req CreateCollaborativePlanRequest) (*CollaborativePlan, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if req.Objective == "" {
		return nil, errors.New("objective is required for collaborative plan")
	}

	coordAgentID := req.CoordinatorAgentID
	if coordAgentID == "" {
		coordAgentID = "planning_agent"
	}

	agent, err := s.GetAgent(ctx, orgID, coordAgentID)
	if err != nil {
		return nil, fmt.Errorf("coordinator agent not found: %w", err)
	}
	if !agent.IsEnabled {
		return nil, ErrAgentDisabled
	}

	planID := generateID("plan")
	corrID := req.CorrelationID
	if corrID == "" {
		corrID = generateID("corr")
	}

	var ctxRef map[string]interface{}
	if len(req.ContextReferences) > 0 {
		ctxRef = map[string]interface{}{"references": req.ContextReferences}
	}

	// 1. Create root coordinator task
	coordTask, err := s.CreateTask(ctx, orgID, CreateTaskRequest{
		Objective:        req.Objective,
		AssignedAgentID:  coordAgentID,
		InitiatingEvent:  "COLLABORATIVE_PLANNING_INITIATED",
		Priority:         PriorityHigh,
		CorrelationID:    corrID,
		ContextReference: ctxRef,
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating coordinator task: %w", err)
	}

	// Persist initial context facts into coordinator task before decomposition
	for _, ci := range req.ContextItems {
		_, _ = s.AddContextItem(ctx, orgID, coordTask.TaskID, ci)
	}

	var plan CollaborativePlan
	plan.PlanID = planID
	plan.PlanningTaskID = coordTask.TaskID
	plan.Objective = req.Objective
	plan.CoordinatorAgentID = coordAgentID
	plan.Status = "CREATED"
	plan.CompletedSteps = []string{}
	plan.FailedSteps = []string{}
	plan.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	if len(req.Steps) > 0 {
		plan.Steps = req.Steps
		participants := []string{coordAgentID}
		for _, st := range req.Steps {
			found := false
			for _, p := range participants {
				if p == st.AgentID {
					found = true
					break
				}
			}
			if !found {
				participants = append(participants, st.AgentID)
			}
		}
		plan.ParticipatingAgents = participants
		plan.OverallConfidence = 0.90
	} else {
		// Invoke Planning Agent to decompose objective into structured plan steps
		_, sidecarResp, err := s.ExecuteTask(ctx, orgID, coordTask.TaskID)
		if err != nil {
			return nil, fmt.Errorf("failed initial planning decomposition: %w", err)
		}
		if sidecarResp != nil && sidecarResp.CollaborativePlan != nil {
			plan.Steps = sidecarResp.CollaborativePlan.Steps
			plan.ParticipatingAgents = sidecarResp.CollaborativePlan.ParticipatingAgents
			plan.OverallConfidence = sidecarResp.CollaborativePlan.OverallConfidence
		} else {
			// Structured fallback
			plan.ParticipatingAgents = []string{coordAgentID, "shipment_agent", "exception_agent", "customer_agent", "finance_agent"}
			plan.Steps = []CollaborativePlanStep{
				{
					StepID:               "step-1",
					AgentID:              "shipment_agent",
					Objective:            fmt.Sprintf("Inspect tracking and milestones for %s", req.Objective),
					Dependencies:         []string{},
					RequiredCapabilities: []string{"shipment.read", "shipment.analyze"},
					ExpectedOutput:       "Shipment tracking status and ETA forecast",
					IsRequired:           true,
					Status:               "PENDING",
				},
				{
					StepID:               "step-2",
					AgentID:              "exception_agent",
					Objective:            fmt.Sprintf("Analyze operational disruption for %s", req.Objective),
					Dependencies:         []string{"step-1"},
					RequiredCapabilities: []string{"exception.read", "exception.analyze"},
					ExpectedOutput:       "Exception impact and mitigation options",
					IsRequired:           true,
					Status:               "PENDING",
				},
				{
					StepID:               "step-3",
					AgentID:              "customer_agent",
					Objective:            fmt.Sprintf("Assess customer impact for %s", req.Objective),
					Dependencies:         []string{"step-1"},
					RequiredCapabilities: []string{"customer.read", "customer.analyze"},
					ExpectedOutput:       "Customer relationship impact and communication draft",
					IsRequired:           true,
					Status:               "PENDING",
				},
				{
					StepID:               "step-4",
					AgentID:              "finance_agent",
					Objective:            fmt.Sprintf("Assess financial exposure for %s", req.Objective),
					Dependencies:         []string{"step-2"},
					RequiredCapabilities: []string{"finance.analyze"},
					ExpectedOutput:       "Demurrage and financial cost exposure",
					IsRequired:           false,
					Status:               "PENDING",
				},
			}
			plan.OverallConfidence = 0.92
		}
	}

	// Persist plan in coordinator task
	s.savePlanToContext(ctx, orgID, coordTask.TaskID, &plan)

	s.logAuditEvent(ctx, orgID, "WORKFORCE_COLLABORATIVE_PLAN_CREATED", "WORKFORCE_TASK", coordTask.TaskID, map[string]interface{}{
		"plan_id":        planID,
		"step_count":     len(plan.Steps),
		"objective":      req.Objective,
		"coordinator_id": coordAgentID,
	})

	return &plan, nil
}

func (s *workforceService) GetCollaborativePlan(ctx context.Context, orgID int64, planID string) (*CollaborativePlan, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}

	tasks, _, err := s.repo.ListTasks(ctx, TaskFilter{
		OrgID: orgID,
	})
	if err != nil {
		return nil, err
	}

	for _, t := range tasks {
		contexts, _ := s.repo.ListContextItems(ctx, orgID, t.TaskID)
		for _, c := range contexts {
			if c.EntityType != nil && *c.EntityType == "COLLABORATIVE_PLAN" && c.EntityID != nil && *c.EntityID == planID {
				var plan CollaborativePlan
				if err := json.Unmarshal(c.Content, &plan); err == nil {
					return &plan, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("collaborative plan '%s' not found", planID)
}

func (s *workforceService) ExecuteCollaborativePlan(ctx context.Context, orgID int64, planID string) (*CollaborativePlan, error) {
	plan, err := s.GetCollaborativePlan(ctx, orgID, planID)
	if err != nil {
		return nil, err
	}

	coordTask, err := s.GetTask(ctx, orgID, plan.PlanningTaskID)
	if err != nil {
		return nil, fmt.Errorf("coordinator task not found: %w", err)
	}

	plan.Status = "IN_PROGRESS"
	_ = s.repo.UpdateTaskStatus(ctx, orgID, coordTask.TaskID, TaskStatusRunning, nil, 0, nil, nil)

	completedStepMap := make(map[string]bool)
	failedStepMap := make(map[string]bool)

	// Step Execution Loop: Runs ready steps, supporting parallel execution for independent steps
	for {
		var readySteps []*CollaborativePlanStep
		for i := range plan.Steps {
			step := &plan.Steps[i]
			if step.Status != "PENDING" {
				if step.Status == "COMPLETED" {
					completedStepMap[step.StepID] = true
				} else if step.Status == "FAILED" {
					failedStepMap[step.StepID] = true
				}
				continue
			}

			// Check dependencies
			depsSatisfied := true
			hasFailedDep := false
			for _, depID := range step.Dependencies {
				if failedStepMap[depID] {
					hasFailedDep = true
					break
				}
				if !completedStepMap[depID] {
					depsSatisfied = false
					break
				}
			}

			if hasFailedDep {
				step.Status = "SKIPPED"
				step.ErrorMessage = "Prerequisite dependency failed"
				continue
			}

			if depsSatisfied {
				readySteps = append(readySteps, step)
			}
		}

		if len(readySteps) == 0 {
			break
		}

		// Execute ready steps concurrently (Section 5: Parallel vs Sequential Work)
		var wg sync.WaitGroup
		type stepExecutionResult struct {
			step *CollaborativePlanStep
			resp *SidecarTaskResponse
			err  error
		}
		resultsChan := make(chan stepExecutionResult, len(readySteps))

		for _, st := range readySteps {
			st.Status = "RUNNING"
			wg.Add(1)
			go func(stepPtr *CollaborativePlanStep) {
				defer wg.Done()

				mappedDeps := make([]string, 0, len(stepPtr.Dependencies))
				for _, depID := range stepPtr.Dependencies {
					for _, prev := range plan.Steps {
						if prev.StepID == depID && prev.TaskID != "" {
							mappedDeps = append(mappedDeps, prev.TaskID)
							break
						}
					}
				}

				childTask, err := s.DelegateTask(ctx, orgID, coordTask.TaskID, DelegateTaskRequest{
					TargetAgentID:        stepPtr.AgentID,
					Objective:            stepPtr.Objective,
					RequiredCapabilities: stepPtr.RequiredCapabilities,
					ExpectedOutput:       stepPtr.ExpectedOutput,
					Priority:             coordTask.Priority,
					Dependencies:         mappedDeps,
					IsOptional:           !stepPtr.IsRequired,
				})
				if err != nil {
					resultsChan <- stepExecutionResult{step: stepPtr, err: err}
					return
				}
				stepPtr.TaskID = childTask.TaskID

				// Propagate context facts to child
				parentContexts, _ := s.repo.ListContextItems(ctx, orgID, coordTask.TaskID)
				for _, pc := range parentContexts {
					if pc.ItemType == ContextTypeFact {
						var cnt map[string]interface{}
						_ = json.Unmarshal(pc.Content, &cnt)
						eType := ""
						if pc.EntityType != nil {
							eType = *pc.EntityType
						}
						eID := ""
						if pc.EntityID != nil {
							eID = *pc.EntityID
						}
						_, _ = s.AddContextItem(ctx, orgID, childTask.TaskID, AddContextItemRequest{
							ItemType:   pc.ItemType,
							EntityType: eType,
							EntityID:   eID,
							Content:    cnt,
							Confidence: pc.Confidence,
						})
					}
				}

				// Execute child specialist task
				_, sidecarResp, err := s.ExecuteTask(ctx, orgID, childTask.TaskID)
				resultsChan <- stepExecutionResult{step: stepPtr, resp: sidecarResp, err: err}
			}(st)
		}

		wg.Wait()
		close(resultsChan)

		// Process step execution outcomes
		for res := range resultsChan {
			if res.err != nil || (res.resp != nil && res.resp.Status == "FAILED") {
				errMsg := "execution failed"
				if res.err != nil {
					errMsg = res.err.Error()
				} else if res.resp != nil && res.resp.ErrorMessage != "" {
					errMsg = res.resp.ErrorMessage
				}
				res.step.Status = "FAILED"
				res.step.ErrorMessage = errMsg
				failedStepMap[res.step.StepID] = true
				plan.FailedSteps = append(plan.FailedSteps, res.step.StepID)

				if res.step.IsRequired {
					// Section 11: Required specialist failure -> Halt and Escalate
					plan.Status = "FAILED"
					plan.EscalationState = true
					errCode := "REQUIRED_SPECIALIST_FAILED"
					_ = s.repo.UpdateTaskStatus(ctx, orgID, coordTask.TaskID, TaskStatusEscalated, nil, 0, &errCode, &errMsg)
					s.logAuditEvent(ctx, orgID, "WORKFORCE_COLLABORATIVE_PLAN_HALTED", "WORKFORCE_TASK", coordTask.TaskID, map[string]interface{}{
						"failed_step_id": res.step.StepID,
						"agent_id":       res.step.AgentID,
						"reason":         errMsg,
					})
					s.savePlanToContext(ctx, orgID, coordTask.TaskID, plan)
					return plan, fmt.Errorf("required specialist '%s' failed: %s", res.step.AgentID, errMsg)
				} else {
					// Section 11: Optional specialist failure -> Continue with limitation
					s.logAuditEvent(ctx, orgID, "WORKFORCE_OPTIONAL_SPECIALIST_FAILED", "WORKFORCE_TASK", coordTask.TaskID, map[string]interface{}{
						"step_id":  res.step.StepID,
						"agent_id": res.step.AgentID,
						"reason":   errMsg,
					})
				}
			} else {
				res.step.Status = "COMPLETED"
				res.step.Confidence = res.resp.Confidence
				res.step.Result = res.resp.Findings
				completedStepMap[res.step.StepID] = true
				plan.CompletedSteps = append(plan.CompletedSteps, res.step.StepID)
				if len(res.resp.RecoveryOptions) > 0 {
					plan.RecoveryOptions = append(plan.RecoveryOptions, res.resp.RecoveryOptions...)
				}
			}
		}
	}

	// PHASE 2: Result Aggregation & Decision Synthesis by Coordinator (Planning Agent)
	_, sidecarResp, err := s.ExecuteTask(ctx, orgID, coordTask.TaskID)
	if err != nil {
		return plan, fmt.Errorf("failed executing coordinator decision synthesis: %w", err)
	}

	if sidecarResp != nil {
		if sidecarResp.DecisionRecord != nil {
			plan.FinalDecision = sidecarResp.DecisionRecord
			plan.OverallConfidence = sidecarResp.DecisionRecord.OverallConfidence

			// Section 13: Human Decision Integration
			if sidecarResp.DecisionRecord.RequiresHumanApproval {
				plan.RequiresApproval = true
				plan.Status = "WAITING_APPROVAL"
				_ = s.repo.UpdateTaskStatus(ctx, orgID, coordTask.TaskID, TaskStatusWaiting, nil, plan.OverallConfidence, nil, nil)
				s.logAuditEvent(ctx, orgID, "WORKFORCE_COLLABORATIVE_DECISION_APPROVAL_REQUIRED", "WORKFORCE_TASK", coordTask.TaskID, map[string]interface{}{
					"plan_id":         plan.PlanID,
					"approval_reason": sidecarResp.DecisionRecord.ApprovalReason,
					"conflicts":       len(sidecarResp.DecisionRecord.Conflicts),
				})
			} else {
				plan.Status = "COMPLETED"
				_ = s.repo.UpdateTaskStatus(ctx, orgID, coordTask.TaskID, TaskStatusCompleted, nil, plan.OverallConfidence, nil, nil)
				s.logAuditEvent(ctx, orgID, "WORKFORCE_COLLABORATIVE_DECISION_COMPLETED", "WORKFORCE_TASK", coordTask.TaskID, map[string]interface{}{
					"plan_id":    plan.PlanID,
					"confidence": plan.OverallConfidence,
				})
			}
		} else {
			plan.Status = "COMPLETED"
			_ = s.repo.UpdateTaskStatus(ctx, orgID, coordTask.TaskID, TaskStatusCompleted, nil, plan.OverallConfidence, nil, nil)
		}

		// Extract Recovery Options (Phase 6.5)
		if len(sidecarResp.RecoveryOptions) > 0 {
			plan.RecoveryOptions = sidecarResp.RecoveryOptions
		} else if sidecarResp.DecisionRecord != nil && len(sidecarResp.DecisionRecord.RecoveryOptions) > 0 {
			plan.RecoveryOptions = sidecarResp.DecisionRecord.RecoveryOptions
		} else if sidecarResp.CollaborativePlan != nil && len(sidecarResp.CollaborativePlan.RecoveryOptions) > 0 {
			plan.RecoveryOptions = sidecarResp.CollaborativePlan.RecoveryOptions
		}

		// Extract Cross-Module Risk (Phase 6.7)
		if sidecarResp.CrossModuleRisk != nil {
			plan.CrossModuleRisk = sidecarResp.CrossModuleRisk
		} else if sidecarResp.DecisionRecord != nil && sidecarResp.DecisionRecord.CrossModuleRisk != nil {
			plan.CrossModuleRisk = sidecarResp.DecisionRecord.CrossModuleRisk
		} else if sidecarResp.CollaborativePlan != nil && sidecarResp.CollaborativePlan.CrossModuleRisk != nil {
			plan.CrossModuleRisk = sidecarResp.CollaborativePlan.CrossModuleRisk
		}
		if len(sidecarResp.RiskAssessments) > 0 {
			plan.RiskAssessments = sidecarResp.RiskAssessments
		} else if sidecarResp.DecisionRecord != nil && len(sidecarResp.DecisionRecord.RiskAssessments) > 0 {
			plan.RiskAssessments = sidecarResp.DecisionRecord.RiskAssessments
		}

		// Populate plan versioning and supersession metadata
		if sidecarResp.CollaborativePlan != nil {
			if sidecarResp.CollaborativePlan.Version > 0 {
				plan.Version = sidecarResp.CollaborativePlan.Version
			}
			if sidecarResp.CollaborativePlan.SupersededPlanID != "" {
				plan.SupersededPlanID = sidecarResp.CollaborativePlan.SupersededPlanID
			}
			if sidecarResp.CollaborativePlan.TriggerEvent != "" {
				plan.TriggerEvent = sidecarResp.CollaborativePlan.TriggerEvent
			}
		}
	} else {
		plan.Status = "COMPLETED"
		_ = s.repo.UpdateTaskStatus(ctx, orgID, coordTask.TaskID, TaskStatusCompleted, nil, plan.OverallConfidence, nil, nil)
	}

	// Persist final plan state
	s.savePlanToContext(ctx, orgID, coordTask.TaskID, plan)

	return plan, nil
}

func (s *workforceService) savePlanToContext(ctx context.Context, orgID int64, taskID string, plan *CollaborativePlan) {
	planJSON, _ := json.Marshal(plan)
	var planMap map[string]interface{}
	_ = json.Unmarshal(planJSON, &planMap)
	_, _ = s.AddContextItem(ctx, orgID, taskID, AddContextItemRequest{
		ItemType:   ContextTypeAgentResult,
		EntityType: "COLLABORATIVE_PLAN",
		EntityID:   plan.PlanID,
		Content:    planMap,
		Confidence: plan.OverallConfidence,
	})
}

// =====================================================================
// Phase 6.5: Multi-Agent Shipment & Exception Operations
// =====================================================================

func (s *workforceService) AssessShipmentHealth(ctx context.Context, orgID int64, req AssessShipmentHealthRequest) (*CollaborativePlan, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if req.ShipmentID == "" {
		return nil, errors.New("shipment_id is required")
	}

	corrID := req.CorrelationID
	if corrID == "" {
		corrID = generateID("corr-health")
	}

	// 1. Fetch real shipment facts from database with tenant isolation
	shContext, err := s.repo.GetShipmentContext(ctx, orgID, req.ShipmentID)
	if err != nil {
		return nil, err
	}
	if shContext == nil {
		shContext = make(map[string]interface{})
	}
	// Merge extra facts provided in request
	for k, v := range req.ExtraFacts {
		shContext[k] = v
	}
	if _, ok := shContext["shipment_id"]; !ok {
		shContext["shipment_id"] = req.ShipmentID
	}
	if _, ok := shContext["status"]; !ok {
		shContext["status"] = "IN_TRANSIT"
	}

	objective := fmt.Sprintf("Assess the current operational health of shipment %s", req.ShipmentID)

	// 2. Formulate collaborative plan with context facts
	plan, err := s.CreateCollaborativePlan(ctx, orgID, CreateCollaborativePlanRequest{
		Objective:     objective,
		CorrelationID: corrID,
		ContextItems: []AddContextItemRequest{
			{
				ItemType:   ContextTypeFact,
				EntityType: "SHIPMENT",
				EntityID:   req.ShipmentID,
				Content:    shContext,
				Confidence: 1.0,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating shipment health collaborative plan: %w", err)
	}

	// 3. Execute plan
	executedPlan, err := s.ExecuteCollaborativePlan(ctx, orgID, plan.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed executing shipment health workflow: %w", err)
	}

	s.logAuditEvent(ctx, orgID, "WORKFORCE_SHIPMENT_HEALTH_ASSESSED", "COLLABORATIVE_PLAN", executedPlan.PlanID, map[string]interface{}{
		"shipment_id": req.ShipmentID,
		"confidence":  executedPlan.OverallConfidence,
		"status":      executedPlan.Status,
	})

	return executedPlan, nil
}

func (s *workforceService) InvestigateShipmentException(ctx context.Context, orgID int64, req InvestigateExceptionRequest) (*CollaborativePlan, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if req.ShipmentID == "" {
		return nil, errors.New("shipment_id is required")
	}

	corrID := req.CorrelationID
	if corrID == "" {
		corrID = generateID("corr-exc")
	}

	// 1. Fetch real shipment & exception facts from database with tenant isolation
	shContext, err := s.repo.GetShipmentContext(ctx, orgID, req.ShipmentID)
	if err != nil {
		return nil, err
	}
	if shContext == nil {
		shContext = make(map[string]interface{})
	}

	exContext, err := s.repo.GetExceptionContext(ctx, orgID, req.ShipmentID, req.ExceptionID)
	if err != nil {
		return nil, err
	}
	if exContext == nil {
		exContext = make(map[string]interface{})
	}

	// Merge all operational facts
	mergedFacts := make(map[string]interface{})
	for k, v := range shContext {
		mergedFacts[k] = v
	}
	for k, v := range exContext {
		mergedFacts[k] = v
	}
	for k, v := range req.ExtraFacts {
		mergedFacts[k] = v
	}
	if _, ok := mergedFacts["shipment_id"]; !ok {
		mergedFacts["shipment_id"] = req.ShipmentID
	}
	if _, ok := mergedFacts["exception_id"]; !ok {
		mergedFacts["exception_id"] = req.ExceptionID
	}

	objective := fmt.Sprintf("Investigate an active exception %s affecting shipment %s", req.ExceptionID, req.ShipmentID)

	// 2. Formulate collaborative plan with context facts
	plan, err := s.CreateCollaborativePlan(ctx, orgID, CreateCollaborativePlanRequest{
		Objective:     objective,
		CorrelationID: corrID,
		ContextItems: []AddContextItemRequest{
			{
				ItemType:   ContextTypeFact,
				EntityType: "EXCEPTION",
				EntityID:   req.ExceptionID,
				Content:    mergedFacts,
				Confidence: 1.0,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating exception investigation collaborative plan: %w", err)
	}

	// 3. Execute collaborative plan
	executedPlan, err := s.ExecuteCollaborativePlan(ctx, orgID, plan.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed executing exception investigation workflow: %w", err)
	}

	// 4. Governance & Action System Boundary Check (Section 9 & 18 & 20)
	// If proposed actions were generated, log audit and verify approval requirements
	if executedPlan.FinalDecision != nil && len(executedPlan.FinalDecision.ProposedActions) > 0 {
		for _, pa := range executedPlan.FinalDecision.ProposedActions {
			s.logAuditEvent(ctx, orgID, "WORKFORCE_PROPOSED_ACTION_EVALUATED", "PROPOSED_ACTION", pa.ActionType, map[string]interface{}{
				"entity_type":       pa.EntityType,
				"entity_id":         pa.EntityID,
				"risk_level":        pa.RiskLevel,
				"requires_approval": pa.RequiresApproval,
				"reasoning":         pa.Reasoning,
			})
		}
	}

	s.logAuditEvent(ctx, orgID, "WORKFORCE_SHIPMENT_EXCEPTION_INVESTIGATED", "COLLABORATIVE_PLAN", executedPlan.PlanID, map[string]interface{}{
		"shipment_id":       req.ShipmentID,
		"exception_id":      req.ExceptionID,
		"confidence":        executedPlan.OverallConfidence,
		"recovery_options": len(executedPlan.RecoveryOptions),
		"requires_approval": executedPlan.RequiresApproval,
	})

	return executedPlan, nil
}

func (s *workforceService) ReplanShipmentOperation(ctx context.Context, orgID int64, req ReplanOperationRequest) (*CollaborativePlan, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if req.OriginalPlanID == "" {
		return nil, errors.New("original_plan_id is required")
	}

	corrID := req.CorrelationID
	if corrID == "" {
		corrID = generateID("corr-replan")
	}

	// 1. Retrieve original plan
	origPlan, err := s.GetCollaborativePlan(ctx, orgID, req.OriginalPlanID)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving original plan '%s': %w", req.OriginalPlanID, err)
	}

	newVersion := origPlan.Version + 1
	if newVersion <= 1 {
		newVersion = 2
	}

	objective := fmt.Sprintf("Replan shipment operation following new event '%s' (superseding plan %s)", req.TriggerEvent, req.OriginalPlanID)

	// 2. Prepare replanning facts
	replanFacts := map[string]interface{}{
		"superseded_plan_id": origPlan.PlanID,
		"prior_plan_id":      origPlan.PlanID,
		"version":            origPlan.Version,
		"new_version":        newVersion,
		"trigger_event":      req.TriggerEvent,
		"is_replanning":      true,
	}
	for k, v := range req.NewFacts {
		replanFacts[k] = v
	}

	// 3. Create replanned collaborative plan
	newPlan, err := s.CreateCollaborativePlan(ctx, orgID, CreateCollaborativePlanRequest{
		Objective:     objective,
		CorrelationID: corrID,
		ContextItems: []AddContextItemRequest{
			{
				ItemType:   ContextTypeFact,
				EntityType: "REPLANNING_EVENT",
				EntityID:   req.TriggerEvent,
				Content:    replanFacts,
				Confidence: 1.0,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating replanning collaborative plan: %w", err)
	}

	// 4. Execute replanning plan
	executedPlan, err := s.ExecuteCollaborativePlan(ctx, orgID, newPlan.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed executing replanning workflow: %w", err)
	}

	executedPlan.Version = newVersion
	executedPlan.SupersededPlanID = origPlan.PlanID
	executedPlan.TriggerEvent = req.TriggerEvent
	s.savePlanToContext(ctx, orgID, executedPlan.PlanningTaskID, executedPlan)

	// 5. Traceability: Audit log the supersession of the prior plan
	s.logAuditEvent(ctx, orgID, "WORKFORCE_PLAN_SUPERSEDED", "COLLABORATIVE_PLAN", origPlan.PlanID, map[string]interface{}{
		"superseded_by": executedPlan.PlanID,
		"prior_version": origPlan.Version,
		"new_version":   newVersion,
		"trigger_event": req.TriggerEvent,
	})

	s.logAuditEvent(ctx, orgID, "WORKFORCE_SHIPMENT_OPERATION_REPLANNED", "COLLABORATIVE_PLAN", executedPlan.PlanID, map[string]interface{}{
		"original_plan_id": origPlan.PlanID,
		"version":          newVersion,
		"trigger_event":    req.TriggerEvent,
	})

	return executedPlan, nil
}

func (s *workforceService) HandleShipmentEvent(ctx context.Context, orgID int64, req ShipmentEventWorkflowRequest) (*CollaborativePlan, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if req.ShipmentID == "" {
		return nil, errors.New("shipment_id is required")
	}

	s.logAuditEvent(ctx, orgID, "WORKFORCE_SHIPMENT_EVENT_RECEIVED", "SHIPMENT_EVENT", req.EventType, map[string]interface{}{
		"shipment_id": req.ShipmentID,
		"event_data":  req.EventData,
	})

	switch req.EventType {
	case "EXCEPTION_CREATED", "EXCEPTION_SEVERITY_CHANGED", "CUSTOMS_HOLD":
		exID := "EXC-EVENT"
		if req.EventData != nil {
			if idVal, ok := req.EventData["exception_id"]; ok {
				exID = fmt.Sprintf("%v", idVal)
			}
		}
		return s.InvestigateShipmentException(ctx, orgID, InvestigateExceptionRequest{
			ShipmentID:    req.ShipmentID,
			ExceptionID:   exID,
			CorrelationID: req.CorrelationID,
			ExtraFacts:    req.EventData,
		})

	default: // "MILESTONE_DELAY", "ETA_RISK_INCREASE", "CARRIER_UPDATE"
		return s.AssessShipmentHealth(ctx, orgID, AssessShipmentHealthRequest{
			ShipmentID:    req.ShipmentID,
			CorrelationID: req.CorrelationID,
			ExtraFacts:    req.EventData,
		})
	}
}

// Phase 6.6: Multi-Agent Customer, Sales, Pricing & Finance Workflows

func (s *workforceService) AssessCustomerRelationship(ctx context.Context, orgID int64, req AssessCustomerRequest) (*CollaborativePlan, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if req.CustomerID == "" {
		return nil, errors.New("customer_id is required")
	}

	corrID := req.CorrelationID
	if corrID == "" {
		corrID = generateID("corr-cust")
	}

	// 1. Fetch real customer & billing facts from database with tenant isolation
	custContext, err := s.repo.GetCustomerContext(ctx, orgID, req.CustomerID)
	if err != nil {
		return nil, err
	}
	if custContext == nil {
		custContext = make(map[string]interface{})
	}
	for k, v := range req.ExtraFacts {
		custContext[k] = v
	}
	if _, ok := custContext["customer_id"]; !ok {
		custContext["customer_id"] = req.CustomerID
	}

	objective := fmt.Sprintf("Assess the current commercial and operational relationship with customer %s", req.CustomerID)

	// 2. Formulate collaborative plan with context facts
	plan, err := s.CreateCollaborativePlan(ctx, orgID, CreateCollaborativePlanRequest{
		Objective:     objective,
		CorrelationID: corrID,
		ContextItems: []AddContextItemRequest{
			{
				ItemType:   ContextTypeFact,
				EntityType: "CUSTOMER",
				EntityID:   req.CustomerID,
				Content:    custContext,
				Confidence: 1.0,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating customer relationship collaborative plan: %w", err)
	}

	// 3. Execute plan
	executedPlan, err := s.ExecuteCollaborativePlan(ctx, orgID, plan.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed executing customer relationship workflow: %w", err)
	}

	s.logAuditEvent(ctx, orgID, "WORKFORCE_CUSTOMER_RELATIONSHIP_ASSESSED", "COLLABORATIVE_PLAN", executedPlan.PlanID, map[string]interface{}{
		"customer_id": req.CustomerID,
		"confidence":  executedPlan.OverallConfidence,
		"status":      executedPlan.Status,
	})

	return executedPlan, nil
}

func (s *workforceService) EvaluateLeadIntelligence(ctx context.Context, orgID int64, req EvaluateLeadRequest) (*CollaborativePlan, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if req.LeadID == "" {
		return nil, errors.New("lead_id is required")
	}

	corrID := req.CorrelationID
	if corrID == "" {
		corrID = generateID("corr-lead")
	}

	// 1. Fetch real lead facts from database with tenant isolation
	leadContext, err := s.repo.GetLeadContext(ctx, orgID, req.LeadID)
	if err != nil {
		return nil, err
	}
	if leadContext == nil {
		leadContext = make(map[string]interface{})
	}
	for k, v := range req.ExtraFacts {
		leadContext[k] = v
	}
	if _, ok := leadContext["lead_id"]; !ok {
		leadContext["lead_id"] = req.LeadID
	}

	objective := fmt.Sprintf("Assess lead quality, evaluate conversion likelihood, and determine sales follow-up priority for lead %s", req.LeadID)

	// 2. Formulate collaborative plan with context facts
	plan, err := s.CreateCollaborativePlan(ctx, orgID, CreateCollaborativePlanRequest{
		Objective:     objective,
		CorrelationID: corrID,
		ContextItems: []AddContextItemRequest{
			{
				ItemType:   ContextTypeFact,
				EntityType: "LEAD",
				EntityID:   req.LeadID,
				Content:    leadContext,
				Confidence: 1.0,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating lead intelligence collaborative plan: %w", err)
	}

	// 3. Execute plan
	executedPlan, err := s.ExecuteCollaborativePlan(ctx, orgID, plan.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed executing lead intelligence workflow: %w", err)
	}

	s.logAuditEvent(ctx, orgID, "WORKFORCE_LEAD_INTELLIGENCE_EVALUATED", "COLLABORATIVE_PLAN", executedPlan.PlanID, map[string]interface{}{
		"lead_id":    req.LeadID,
		"confidence": executedPlan.OverallConfidence,
		"status":     executedPlan.Status,
	})

	return executedPlan, nil
}

func (s *workforceService) EvaluateRFQCommercialWorkflow(ctx context.Context, orgID int64, req EvaluateRFQRequest) (*CollaborativePlan, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if req.RFQID == "" {
		return nil, errors.New("rfq_id is required")
	}

	corrID := req.CorrelationID
	if corrID == "" {
		corrID = generateID("corr-rfq")
	}

	// 1. Fetch real RFQ facts from database with tenant isolation
	rfqContext, err := s.repo.GetRFQContext(ctx, orgID, req.RFQID)
	if err != nil {
		return nil, err
	}
	if rfqContext == nil {
		rfqContext = make(map[string]interface{})
	}
	for k, v := range req.ExtraFacts {
		rfqContext[k] = v
	}
	if _, ok := rfqContext["rfq_id"]; !ok {
		rfqContext["rfq_id"] = req.RFQID
	}

	// If RFQ has an associated customer, augment with customer context
	var custContext map[string]interface{}
	if custIDVal, ok := rfqContext["customer_id"]; ok {
		custIDStr := fmt.Sprintf("%v", custIDVal)
		if custIDStr != "" && custIDStr != "0" {
			custContext, _ = s.repo.GetCustomerContext(ctx, orgID, custIDStr)
		}
	}

	objective := fmt.Sprintf("Collaboratively evaluate RFQ %s: analyze pricing, margin, commercial risk, contract terms, and formulate quotation recommendation", req.RFQID)

	contextItems := []AddContextItemRequest{
		{
			ItemType:   ContextTypeFact,
			EntityType: "RFQ",
			EntityID:   req.RFQID,
			Content:    rfqContext,
			Confidence: 1.0,
		},
	}
	if len(custContext) > 0 {
		contextItems = append(contextItems, AddContextItemRequest{
			ItemType:   ContextTypeFact,
			EntityType: "CUSTOMER",
			EntityID:   fmt.Sprintf("%v", custContext["customer_id"]),
			Content:    custContext,
			Confidence: 1.0,
		})
	}

	// 2. Formulate collaborative plan with context facts
	plan, err := s.CreateCollaborativePlan(ctx, orgID, CreateCollaborativePlanRequest{
		Objective:     objective,
		CorrelationID: corrID,
		ContextItems:  contextItems,
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating RFQ commercial collaborative plan: %w", err)
	}

	// 3. Execute plan
	executedPlan, err := s.ExecuteCollaborativePlan(ctx, orgID, plan.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed executing RFQ commercial workflow: %w", err)
	}

	s.logAuditEvent(ctx, orgID, "WORKFORCE_RFQ_COMMERCIAL_EVALUATED", "COLLABORATIVE_PLAN", executedPlan.PlanID, map[string]interface{}{
		"rfq_id":            req.RFQID,
		"confidence":        executedPlan.OverallConfidence,
		"status":            executedPlan.Status,
		"requires_approval": executedPlan.RequiresApproval,
	})

	return executedPlan, nil
}

func (s *workforceService) AssessInvoiceCollections(ctx context.Context, orgID int64, req AssessCollectionsRequest) (*CollaborativePlan, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if req.CustomerID == "" && req.InvoiceID == "" {
		return nil, errors.New("either customer_id or invoice_id is required")
	}

	corrID := req.CorrelationID
	if corrID == "" {
		corrID = generateID("corr-coll")
	}

	// 1. Fetch real invoice context
	invContext, err := s.repo.GetInvoiceContext(ctx, orgID, req.CustomerID, req.InvoiceID)
	if err != nil {
		return nil, err
	}
	if invContext == nil {
		invContext = make(map[string]interface{})
	}
	for k, v := range req.ExtraFacts {
		invContext[k] = v
	}
	if req.InvoiceID != "" {
		if _, ok := invContext["invoice_id"]; !ok {
			invContext["invoice_id"] = req.InvoiceID
		}
	}
	if req.CustomerID != "" {
		if _, ok := invContext["customer_id"]; !ok {
			invContext["customer_id"] = req.CustomerID
		}
	}

	// Also retrieve customer context if customer ID is available
	var custContext map[string]interface{}
	cID := req.CustomerID
	if cID == "" && invContext["customer_id"] != nil {
		cID = fmt.Sprintf("%v", invContext["customer_id"])
	}
	if cID != "" {
		custContext, _ = s.repo.GetCustomerContext(ctx, orgID, cID)
	}

	targetDesc := req.InvoiceID
	if targetDesc == "" {
		targetDesc = fmt.Sprintf("customer %s", req.CustomerID)
	}
	objective := fmt.Sprintf("Determine the appropriate next step for overdue customer invoices and collections triage for %s", targetDesc)

	contextItems := []AddContextItemRequest{
		{
			ItemType:   ContextTypeFact,
			EntityType: "INVOICE",
			EntityID:   req.InvoiceID,
			Content:    invContext,
			Confidence: 1.0,
		},
	}
	if len(custContext) > 0 {
		contextItems = append(contextItems, AddContextItemRequest{
			ItemType:   ContextTypeFact,
			EntityType: "CUSTOMER",
			EntityID:   cID,
			Content:    custContext,
			Confidence: 1.0,
		})
	}

	// 2. Formulate collaborative plan with context facts
	plan, err := s.CreateCollaborativePlan(ctx, orgID, CreateCollaborativePlanRequest{
		Objective:     objective,
		CorrelationID: corrID,
		ContextItems:  contextItems,
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating collections collaborative plan: %w", err)
	}

	// 3. Execute plan
	executedPlan, err := s.ExecuteCollaborativePlan(ctx, orgID, plan.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed executing collections workflow: %w", err)
	}

	s.logAuditEvent(ctx, orgID, "WORKFORCE_COLLECTIONS_ASSESSED", "COLLABORATIVE_PLAN", executedPlan.PlanID, map[string]interface{}{
		"target":            targetDesc,
		"confidence":        executedPlan.OverallConfidence,
		"status":            executedPlan.Status,
		"requires_approval": executedPlan.RequiresApproval,
	})

	return executedPlan, nil
}

func (s *workforceService) HandleCommercialEvent(ctx context.Context, orgID int64, req CommercialEventWorkflowRequest) (*CollaborativePlan, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if req.EntityID == "" {
		return nil, errors.New("entity_id is required")
	}

	s.logAuditEvent(ctx, orgID, "WORKFORCE_COMMERCIAL_EVENT_RECEIVED", "COMMERCIAL_EVENT", req.EventType, map[string]interface{}{
		"entity_id":  req.EntityID,
		"event_data": req.EventData,
	})

	switch req.EventType {
	case "NEW_RFQ", "RFQ_UPDATED":
		return s.EvaluateRFQCommercialWorkflow(ctx, orgID, EvaluateRFQRequest{
			RFQID:         req.EntityID,
			CorrelationID: req.CorrelationID,
			ExtraFacts:    req.EventData,
		})

	case "INVOICE_OVERDUE", "PAYMENT_OVERDUE", "DISPUTE_OPENED":
		return s.AssessInvoiceCollections(ctx, orgID, AssessCollectionsRequest{
			InvoiceID:     req.EntityID,
			CorrelationID: req.CorrelationID,
			ExtraFacts:    req.EventData,
		})

	case "LEAD_CREATED", "LEAD_STATUS_CHANGED":
		return s.EvaluateLeadIntelligence(ctx, orgID, EvaluateLeadRequest{
			LeadID:        req.EntityID,
			CorrelationID: req.CorrelationID,
			ExtraFacts:    req.EventData,
		})

	default: // "CUSTOMER_FOLLOWUP_DUE", "ACCOUNT_HEALTH_ALERT"
		return s.AssessCustomerRelationship(ctx, orgID, AssessCustomerRequest{
			CustomerID:    req.EntityID,
			CorrelationID: req.CorrelationID,
			ExtraFacts:    req.EventData,
		})
	}
}

// Phase 6.7: Multi-Agent Contract, Compliance & Risk Service Implementations

func (s *workforceService) AssessContractRisk(ctx context.Context, orgID int64, req AssessContractRiskRequest) (*CollaborativePlan, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if req.ContractID == "" {
		return nil, errors.New("contract_id is required")
	}

	corrID := req.CorrelationID
	if corrID == "" {
		corrID = generateID("corr-ctr-risk")
	}

	// 1. Fetch Contract Context
	ctrContext, err := s.repo.GetContractContext(ctx, orgID, req.ContractID)
	if err != nil {
		return nil, fmt.Errorf("failed fetching contract context: %w", err)
	}
	for k, v := range req.ExtraFacts {
		ctrContext[k] = v
	}

	contextItems := []AddContextItemRequest{
		{
			ItemType:   ContextTypeFact,
			EntityType: "CONTRACT",
			EntityID:   req.ContractID,
			Content:    ctrContext,
			Confidence: 1.0,
		},
	}

	// If shipment ID provided, also fetch shipment context
	if req.ShipmentID != "" {
		shpContext, _ := s.repo.GetShipmentContext(ctx, orgID, req.ShipmentID)
		if len(shpContext) > 0 {
			contextItems = append(contextItems, AddContextItemRequest{
				ItemType:   ContextTypeFact,
				EntityType: "SHIPMENT",
				EntityID:   req.ShipmentID,
				Content:    shpContext,
				Confidence: 1.0,
			})
		}
	}

	objective := req.Objective
	if objective == "" {
		objective = fmt.Sprintf("Assess contractual implications and risk of operational events for contract %s", req.ContractID)
	}

	// 2. Create and execute plan
	plan, err := s.CreateCollaborativePlan(ctx, orgID, CreateCollaborativePlanRequest{
		Objective:     objective,
		CorrelationID: corrID,
		ContextItems:  contextItems,
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating contract risk collaborative plan: %w", err)
	}

	executedPlan, err := s.ExecuteCollaborativePlan(ctx, orgID, plan.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed executing contract risk workflow: %w", err)
	}

	s.logAuditEvent(ctx, orgID, "WORKFORCE_CONTRACT_RISK_ASSESSED", "COLLABORATIVE_PLAN", executedPlan.PlanID, map[string]interface{}{
		"contract_id":       req.ContractID,
		"confidence":        executedPlan.OverallConfidence,
		"status":            executedPlan.Status,
		"requires_approval": executedPlan.RequiresApproval,
	})

	return executedPlan, nil
}

func (s *workforceService) AssessComplianceRisk(ctx context.Context, orgID int64, req AssessComplianceRiskRequest) (*CollaborativePlan, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if req.ShipmentID == "" {
		return nil, errors.New("shipment_id is required")
	}

	corrID := req.CorrelationID
	if corrID == "" {
		corrID = generateID("corr-cmp-risk")
	}

	// 1. Fetch Compliance & Shipment Context
	cmpContext, err := s.repo.GetComplianceContext(ctx, orgID, req.ShipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed fetching compliance context: %w", err)
	}
	for k, v := range req.ExtraFacts {
		cmpContext[k] = v
	}

	shpContext, _ := s.repo.GetShipmentContext(ctx, orgID, req.ShipmentID)

	contextItems := []AddContextItemRequest{
		{
			ItemType:   ContextTypeFact,
			EntityType: "COMPLIANCE",
			EntityID:   req.ShipmentID,
			Content:    cmpContext,
			Confidence: 1.0,
		},
	}
	if len(shpContext) > 0 {
		contextItems = append(contextItems, AddContextItemRequest{
			ItemType:   ContextTypeFact,
			EntityType: "SHIPMENT",
			EntityID:   req.ShipmentID,
			Content:    shpContext,
			Confidence: 1.0,
		})
	}

	objective := req.Objective
	if objective == "" {
		objective = fmt.Sprintf("Assess compliance risk, regulatory requirements, and documentation gaps for shipment %s", req.ShipmentID)
	}

	// 2. Create and execute plan
	plan, err := s.CreateCollaborativePlan(ctx, orgID, CreateCollaborativePlanRequest{
		Objective:     objective,
		CorrelationID: corrID,
		ContextItems:  contextItems,
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating compliance risk collaborative plan: %w", err)
	}

	executedPlan, err := s.ExecuteCollaborativePlan(ctx, orgID, plan.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed executing compliance risk workflow: %w", err)
	}

	s.logAuditEvent(ctx, orgID, "WORKFORCE_COMPLIANCE_RISK_ASSESSED", "COLLABORATIVE_PLAN", executedPlan.PlanID, map[string]interface{}{
		"shipment_id":       req.ShipmentID,
		"confidence":        executedPlan.OverallConfidence,
		"status":            executedPlan.Status,
		"requires_approval": executedPlan.RequiresApproval,
	})

	return executedPlan, nil
}

func (s *workforceService) AssessCrossModuleRisk(ctx context.Context, orgID int64, req AssessCrossModuleRiskRequest) (*CollaborativePlan, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if req.ShipmentID == "" {
		return nil, errors.New("shipment_id is required")
	}

	corrID := req.CorrelationID
	if corrID == "" {
		corrID = generateID("corr-xmod-risk")
	}

	// 1. Fetch Multi-Domain Contexts: Shipment, Exception, Contract, Compliance
	shpContext, _ := s.repo.GetShipmentContext(ctx, orgID, req.ShipmentID)
	for k, v := range req.ExtraFacts {
		if shpContext == nil {
			shpContext = make(map[string]interface{})
		}
		shpContext[k] = v
	}

	contextItems := []AddContextItemRequest{
		{
			ItemType:   ContextTypeFact,
			EntityType: "SHIPMENT",
			EntityID:   req.ShipmentID,
			Content:    shpContext,
			Confidence: 1.0,
		},
	}

	if req.ExceptionID != "" {
		excContext, _ := s.repo.GetExceptionContext(ctx, orgID, req.ShipmentID, req.ExceptionID)
		if len(excContext) > 0 {
			contextItems = append(contextItems, AddContextItemRequest{
				ItemType:   ContextTypeFact,
				EntityType: "EXCEPTION",
				EntityID:   req.ExceptionID,
				Content:    excContext,
				Confidence: 1.0,
			})
		}
	}

	if req.ContractID != "" {
		ctrContext, _ := s.repo.GetContractContext(ctx, orgID, req.ContractID)
		if len(ctrContext) > 0 {
			contextItems = append(contextItems, AddContextItemRequest{
				ItemType:   ContextTypeFact,
				EntityType: "CONTRACT",
				EntityID:   req.ContractID,
				Content:    ctrContext,
				Confidence: 1.0,
			})
		}
	}

	cmpContext, _ := s.repo.GetComplianceContext(ctx, orgID, req.ShipmentID)
	if len(cmpContext) > 0 {
		contextItems = append(contextItems, AddContextItemRequest{
			ItemType:   ContextTypeFact,
			EntityType: "COMPLIANCE",
			EntityID:   req.ShipmentID,
			Content:    cmpContext,
			Confidence: 1.0,
		})
	}

	objective := req.Objective
	if objective == "" {
		objective = fmt.Sprintf("Assess the cross-module risk of shipment %s across operational, contractual, compliance, and financial dimensions", req.ShipmentID)
	}

	// 2. Create and execute plan
	plan, err := s.CreateCollaborativePlan(ctx, orgID, CreateCollaborativePlanRequest{
		Objective:     objective,
		CorrelationID: corrID,
		ContextItems:  contextItems,
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating cross-module risk collaborative plan: %w", err)
	}

	executedPlan, err := s.ExecuteCollaborativePlan(ctx, orgID, plan.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed executing cross-module risk workflow: %w", err)
	}

	s.logAuditEvent(ctx, orgID, "WORKFORCE_CROSS_MODULE_RISK_ASSESSED", "COLLABORATIVE_PLAN", executedPlan.PlanID, map[string]interface{}{
		"shipment_id":       req.ShipmentID,
		"confidence":        executedPlan.OverallConfidence,
		"status":            executedPlan.Status,
		"requires_approval": executedPlan.RequiresApproval,
	})

	return executedPlan, nil
}

func (s *workforceService) ReassessCrossModuleRisk(ctx context.Context, orgID int64, req ReassessRiskRequest) (*CollaborativePlan, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	origID := req.OriginalPlanID
	if origID == "" {
		origID = req.PriorRiskID
	}
	if origID == "" {
		return nil, errors.New("original_plan_id is required")
	}

	originalPlan, err := s.GetCollaborativePlan(ctx, orgID, origID)
	if err != nil {
		return nil, fmt.Errorf("original plan '%s' not found: %w", origID, err)
	}

	corrID := req.CorrelationID
	if corrID == "" {
		corrID = generateID("corr-risk-replan")
	}

	replanObjective := fmt.Sprintf("Reassess cross-module risk following event '%s' (superseding plan %s): %s",
		req.TriggerEvent, originalPlan.PlanID, originalPlan.Objective)

	contextItems := []AddContextItemRequest{
		{
			ItemType: ContextTypeFact,
			Content: map[string]interface{}{
				"version":            originalPlan.Version,
				"superseded_plan_id": originalPlan.PlanID,
				"prior_plan_id":      originalPlan.PlanID,
				"trigger_event":      req.TriggerEvent,
				"event_data":         req.NewFacts,
			},
			Confidence: 1.0,
		},
	}
	if req.NewFacts != nil {
		contextItems = append(contextItems, AddContextItemRequest{
			ItemType:   ContextTypeFact,
			Content:    req.NewFacts,
			Confidence: 1.0,
		})
	}

	plan, err := s.CreateCollaborativePlan(ctx, orgID, CreateCollaborativePlanRequest{
		Objective:     replanObjective,
		CorrelationID: corrID,
		ContextItems:  contextItems,
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating risk reassessment plan: %w", err)
	}

	executedPlan, err := s.ExecuteCollaborativePlan(ctx, orgID, plan.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed executing risk reassessment plan: %w", err)
	}

	s.logAuditEvent(ctx, orgID, "WORKFORCE_RISK_REASSESSED", "COLLABORATIVE_PLAN", executedPlan.PlanID, map[string]interface{}{
		"superseded_plan_id": originalPlan.PlanID,
		"trigger_event":      req.TriggerEvent,
		"new_version":        executedPlan.Version,
		"confidence":         executedPlan.OverallConfidence,
	})

	return executedPlan, nil
}

// ----------------------------------------------------------------------
// Phase 6.8: Conflict Resolution, Memory & Learning
// ----------------------------------------------------------------------

func (s *workforceService) ResolveConflict(ctx context.Context, orgID int64, req ResolveConflictRequest) (*ConflictRecord, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}

	conflictID := req.ConflictID
	if conflictID == "" {
		conflictID = fmt.Sprintf("conf_%d_%d", orgID, time.Now().UnixNano())
	}

	record := &ConflictRecord{
		ConflictID:          conflictID,
		TenantID:            orgID,
		ConflictType:        req.ConflictType,
		ParticipatingAgents: req.Participating,
		ConflictingResults:  req.Findings,
		ResolutionStatus:    "RESOLVED",
		ResolutionStrategy:  "CONSENSUS",
		IsResolved:          true,
		CreatedAt:           time.Now().UTC().Format(time.RFC3339),
		ResolvedAt:          time.Now().UTC().Format(time.RFC3339),
	}

	if req.HumanDecision != "" {
		// Human decision boundary - distinct from AI recommendation
		record.ResolutionStatus = "HUMAN_DECISION"
		record.ResolutionStrategy = "ESCALATE_TO_HUMAN"
		record.SelectedOutcome = map[string]interface{}{
			"decision":       req.HumanDecision,
			"source":         "HUMAN_DECISION",
			"human_feedback": req.Reason,
		}
		record.Reasoning = fmt.Sprintf("Resolved by human operator: %s", req.HumanDecision)
	} else if req.AuthoritativeKey != "" {
		record.ResolutionStatus = "RESOLVED"
		record.ResolutionStrategy = "AUTHORITATIVE_DATA_OVERRIDE"
		record.SelectedOutcome = map[string]interface{}{
			"authoritative_key": req.AuthoritativeKey,
			"source":            "BUSINESS_RECORD",
		}
		record.Reasoning = "Resolved by authoritative business record overriding AI heuristics."
	} else {
		record.Reasoning = "Resolved via multi-agent consensus trade-off."
	}

	s.logAuditEvent(ctx, orgID, "WORKFORCE_CONFLICT_RESOLVED", "CONFLICT_RECORD", conflictID, map[string]interface{}{
		"conflict_type":       req.ConflictType,
		"resolution_strategy": record.ResolutionStrategy,
		"resolution_status":   record.ResolutionStatus,
		"has_human_decision":  req.HumanDecision != "",
	})

	return record, nil
}

func (s *workforceService) RecordOutcome(ctx context.Context, orgID int64, req RecordOutcomeRequest) (*AgentOutcome, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}

	outcome := &AgentOutcome{
		OutcomeID:        fmt.Sprintf("out_%d_%d", orgID, time.Now().UnixNano()),
		TenantID:         orgID,
		SourceTaskID:     req.SourceTaskID,
		SourceAgentID:    req.SourceAgentID,
		SourceEntityType: req.SourceEntityType,
		SourceEntityID:   req.SourceEntityID,
		WorkflowID:       req.WorkflowID,
		PlanID:           req.PlanID,
		Objective:        req.Objective,
		Recommendation:   req.Recommendation,
		ActionType:       req.ActionType,
		HumanDecision:    req.HumanDecision,
		HumanFeedback:    req.HumanFeedback,
		ActualOutcome:    req.ActualOutcome,
		Status:           req.Status,
		IsVerified:       req.IsVerified,
		SuccessIndicator: req.SuccessIndicator,
		Evidence:         req.Evidence,
		Lesson:           req.Lesson,
		Confidence:       req.Confidence,
		CorrelationID:    fmt.Sprintf("corr_%d_%d", orgID, time.Now().UnixNano()),
		Metadata:         req.Metadata,
		CreatedAt:        time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.repo.RecordOutcome(ctx, outcome); err != nil {
		return nil, fmt.Errorf("failed to record agent outcome: %w", err)
	}

	s.logAuditEvent(ctx, orgID, "WORKFORCE_OUTCOME_RECORDED", "AGENT_OUTCOME", outcome.OutcomeID, map[string]interface{}{
		"entity_type":         req.SourceEntityType,
		"entity_id":           req.SourceEntityID,
		"status":              req.Status,
		"success_indicator":   req.SuccessIndicator,
		"has_human_feedback": req.HumanFeedback != "",
	})

	return outcome, nil
}

func (s *workforceService) ListOutcomes(ctx context.Context, orgID int64, entityType, entityID string, limit int) ([]AgentOutcome, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	return s.repo.ListOutcomes(ctx, orgID, entityType, entityID, limit)
}

func (s *workforceService) QueryMemory(ctx context.Context, orgID int64, req QueryMemoryRequest) (*MemoryQueryResponse, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}

	items, err := s.repo.GetMemoryContext(ctx, orgID, req.Domain, req.EntityType, req.EntityID, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query memory items: %w", err)
	}

	return &MemoryQueryResponse{
		Domain:        req.Domain,
		TotalFound:    len(items),
		Items:         items,
		PrecedenceMsg: "Current authoritative business telemetry and ground truth records supersede historical memory patterns.",
	}, nil
}

// =====================================================================
// Phase 6.9: Governed Multi-Agent Autonomy & Workforce Command Center
// =====================================================================

func (s *workforceService) EvaluateActionPolicy(ctx context.Context, orgID int64, agentID, actionType string, payload map[string]interface{}) ActionPolicyDecision {
	nowStr := time.Now().UTC().Format(time.RFC3339)
	if orgID <= 0 {
		return ActionPolicyDecision{
			AgentID:            agentID,
			ActionType:         actionType,
			Allowed:            false,
			Decision:           "BLOCKED",
			Reason:             "Unauthorized tenant access",
			RiskLevel:          "CRITICAL",
			RequiresEscalation: true,
			Timestamp:          nowStr,
		}
	}

	// 1. Emergency Stop Check
	if stopped, reason := s.emergStops.IsWorkforceStopped(orgID); stopped {
		return ActionPolicyDecision{
			AgentID:            agentID,
			ActionType:         actionType,
			Allowed:            false,
			Decision:           "BLOCKED",
			Reason:             fmt.Sprintf("Workforce emergency stop active: %s", reason),
			RiskLevel:          "CRITICAL",
			RequiresEscalation: true,
			Timestamp:          nowStr,
		}
	}

	if stopped, reason := s.emergStops.IsAgentStopped(orgID, agentID); stopped {
		return ActionPolicyDecision{
			AgentID:            agentID,
			ActionType:         actionType,
			Allowed:            false,
			Decision:           "BLOCKED",
			Reason:             fmt.Sprintf("Emergency stop active for agent %s: %s", agentID, reason),
			RiskLevel:          "CRITICAL",
			RequiresEscalation: true,
			Timestamp:          nowStr,
		}
	}

	// Determine Action Category and Risk Level
	actionUpper := strings.ToUpper(actionType)
	var actionCategory string
	var riskLevel string

	switch {
	case strings.Contains(actionUpper, "INVOICE") || strings.Contains(actionUpper, "PAYMENT") || strings.Contains(actionUpper, "FINANCE") || strings.Contains(actionUpper, "CREDIT") || strings.Contains(actionUpper, "DISCOUNT"):
		actionCategory = "FINANCIAL"
		riskLevel = "HIGH"
	case strings.Contains(actionUpper, "QUOTE") || strings.Contains(actionUpper, "COMMERCIAL") || strings.Contains(actionUpper, "RATE") || strings.Contains(actionUpper, "BOOKING") || strings.Contains(actionUpper, "RFQ"):
		actionCategory = "COMMERCIAL"
		riskLevel = "HIGH"
	case strings.Contains(actionUpper, "COMMUNICATION") || strings.Contains(actionUpper, "EMAIL") || strings.Contains(actionUpper, "NOTIFY") || strings.Contains(actionUpper, "MESSAGE"):
		actionCategory = "COMMUNICATION"
		riskLevel = "MEDIUM"
	case strings.Contains(actionUpper, "SECURITY") || strings.Contains(actionUpper, "AUTH") || strings.Contains(actionUpper, "CUSTOMS") || strings.Contains(actionUpper, "HAZMAT") || strings.Contains(actionUpper, "BYPASS"):
		actionCategory = "COMPLIANCE"
		riskLevel = "CRITICAL"
	case strings.Contains(actionUpper, "TAG") || strings.Contains(actionUpper, "OBSERVE") || strings.Contains(actionUpper, "LOG") || strings.Contains(actionUpper, "CACHE") || strings.Contains(actionUpper, "METRIC"):
		actionCategory = "INTERNAL"
		riskLevel = "LOW"
	default:
		actionCategory = "OPERATIONAL"
		riskLevel = "MEDIUM"
	}

	if stopped, reason := s.emergStops.IsActionClassStopped(orgID, actionCategory); stopped {
		return ActionPolicyDecision{
			AgentID:            agentID,
			ActionType:         actionType,
			Allowed:            false,
			Decision:           "BLOCKED",
			Reason:             fmt.Sprintf("Emergency stop active for action category %s: %s", actionCategory, reason),
			RiskLevel:          riskLevel,
			RequiresEscalation: true,
			Timestamp:          nowStr,
		}
	}

	// 2. Fetch Agent Profile
	agent, err := s.repo.GetAgent(ctx, orgID, agentID)
	if err != nil || agent == nil {
		return ActionPolicyDecision{
			AgentID:            agentID,
			ActionType:         actionType,
			Allowed:            false,
			Decision:           "BLOCKED",
			Reason:             "Assigned agent not registered or invalid for this tenant",
			RiskLevel:          "HIGH",
			RequiresEscalation: true,
			Timestamp:          nowStr,
		}
	}

	if !agent.IsEnabled {
		return ActionPolicyDecision{
			AgentID:            agentID,
			AutonomyLevel:      agent.AutonomyLevel,
			ActionType:         actionType,
			Allowed:            false,
			Decision:           "BLOCKED",
			Reason:             fmt.Sprintf("Agent '%s' is disabled", agentID),
			RiskLevel:          riskLevel,
			Timestamp:          nowStr,
		}
	}

	if agent.OperationalStatus == AgentStatusPaused || agent.HealthStatus == AgentStatusPaused {
		return ActionPolicyDecision{
			AgentID:            agentID,
			AutonomyLevel:      agent.AutonomyLevel,
			ActionType:         actionType,
			Allowed:            false,
			Decision:           "BLOCKED",
			Reason:             fmt.Sprintf("Agent '%s' is temporarily paused", agentID),
			RiskLevel:          riskLevel,
			Timestamp:          nowStr,
		}
	}

	// 3. Prevent Self-Elevation / Malicious Bypass Attempts
	if riskLevel == "CRITICAL" {
		return ActionPolicyDecision{
			AgentID:            agentID,
			AutonomyLevel:      agent.AutonomyLevel,
			ActionType:         actionType,
			Allowed:            false,
			CanRecommend:       false,
			CanPrepare:         false,
			CanAutoExecute:     false,
			RequiresApproval:   true,
			RequiresEscalation: true,
			Decision:           "ESCALATE_CRITICAL",
			Reason:             "Critical risk action cannot be executed autonomously or prepared without administrative authorization",
			RiskLevel:          "CRITICAL",
			Timestamp:          nowStr,
		}
	}

	autonomy := agent.AutonomyLevel
	if autonomy == "" {
		autonomy = AutonomyLevel1Recommend
	}

	switch autonomy {
	case AutonomyLevel0Observe:
		return ActionPolicyDecision{
			AgentID:            agentID,
			AutonomyLevel:      autonomy,
			ActionType:         actionType,
			Allowed:            false,
			CanRecommend:       false,
			CanPrepare:         false,
			CanAutoExecute:     false,
			RequiresApproval:   false,
			RequiresEscalation: false,
			Decision:           "BLOCKED",
			Reason:             "Level 0 (Observe): Agent may only monitor telemetry. Action proposals and executions are prohibited.",
			RiskLevel:          riskLevel,
			Timestamp:          nowStr,
		}

	case AutonomyLevel1Recommend:
		return ActionPolicyDecision{
			AgentID:            agentID,
			AutonomyLevel:      autonomy,
			ActionType:         actionType,
			Allowed:            true,
			CanRecommend:       true,
			CanPrepare:         false,
			CanAutoExecute:     false,
			RequiresApproval:   false,
			RequiresEscalation: false,
			Decision:           "RECOMMEND_ONLY",
			Reason:             "Level 1 (Recommend): Agent may produce advisory recommendations only.",
			RiskLevel:          riskLevel,
			Timestamp:          nowStr,
		}

	case AutonomyLevel2Prepare:
		return ActionPolicyDecision{
			AgentID:            agentID,
			AutonomyLevel:      autonomy,
			ActionType:         actionType,
			Allowed:            true,
			CanRecommend:       true,
			CanPrepare:         true,
			CanAutoExecute:     false,
			RequiresApproval:   true,
			RequiresEscalation: false,
			Decision:           "PREPARE_FOR_APPROVAL",
			Reason:             "Level 2 (Prepare): Action prepared for human-in-the-loop review and approval gate.",
			RiskLevel:          riskLevel,
			Timestamp:          nowStr,
		}

	case AutonomyLevel3ControlledExecution:
		if riskLevel == "LOW" {
			return ActionPolicyDecision{
				AgentID:            agentID,
				AutonomyLevel:      autonomy,
				ActionType:         actionType,
				Allowed:            true,
				CanRecommend:       true,
				CanPrepare:         true,
				CanAutoExecute:     true,
				RequiresApproval:   false,
				RequiresEscalation: false,
				Decision:           "EXECUTE_PERMITTED",
				Reason:             "Level 3: Low-risk internal telemetry action permitted for automatic execution under Go policy.",
				RiskLevel:          riskLevel,
				Timestamp:          nowStr,
			}
		}
		return ActionPolicyDecision{
			AgentID:            agentID,
			AutonomyLevel:      autonomy,
			ActionType:         actionType,
			Allowed:            true,
			CanRecommend:       true,
			CanPrepare:         true,
			CanAutoExecute:     false,
			RequiresApproval:   true,
			RequiresEscalation: riskLevel == "HIGH",
			Decision:           "APPROVAL_REQUIRED",
			Reason:             fmt.Sprintf("Level 3: %s risk action requires explicit human approval.", riskLevel),
			RiskLevel:          riskLevel,
			Timestamp:          nowStr,
		}

	case AutonomyLevel4GovernedMultiStep:
		if riskLevel == "LOW" {
			return ActionPolicyDecision{
				AgentID:            agentID,
				AutonomyLevel:      autonomy,
				ActionType:         actionType,
				Allowed:            true,
				CanRecommend:       true,
				CanPrepare:         true,
				CanAutoExecute:     true,
				RequiresApproval:   false,
				RequiresEscalation: false,
				Decision:           "EXECUTE_PERMITTED",
				Reason:             "Level 4: Low-risk multi-step action permitted for autonomous coordination.",
				RiskLevel:          riskLevel,
				Timestamp:          nowStr,
			}
		}
		if riskLevel == "MEDIUM" {
			return ActionPolicyDecision{
				AgentID:            agentID,
				AutonomyLevel:      autonomy,
				ActionType:         actionType,
				Allowed:            true,
				CanRecommend:       true,
				CanPrepare:         true,
				CanAutoExecute:     false,
				RequiresApproval:   true,
				RequiresEscalation: false,
				Decision:           "APPROVAL_REQUIRED",
				Reason:             "Level 4: Medium-risk workflow step prepared for human review gate.",
				RiskLevel:          riskLevel,
				Timestamp:          nowStr,
			}
		}
		return ActionPolicyDecision{
			AgentID:            agentID,
			AutonomyLevel:      autonomy,
			ActionType:         actionType,
			Allowed:            true,
			CanRecommend:       true,
			CanPrepare:         true,
			CanAutoExecute:     false,
			RequiresApproval:   true,
			RequiresEscalation: true,
			Decision:           "ESCALATE_REQUIRED",
			Reason:             "Level 4: High-risk operational/commercial step requires formal escalation and approval.",
			RiskLevel:          riskLevel,
			Timestamp:          nowStr,
		}

	default:
		return ActionPolicyDecision{
			AgentID:            agentID,
			AutonomyLevel:      autonomy,
			ActionType:         actionType,
			Allowed:            false,
			Decision:           "BLOCKED",
			Reason:             fmt.Sprintf("Unknown autonomy level '%s'", autonomy),
			RiskLevel:          riskLevel,
			Timestamp:          nowStr,
		}
	}
}

func (s *workforceService) ControlAgent(ctx context.Context, orgID int64, agentID string, req AgentControlRequest, actor string) (*WorkforceAgent, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if agentID == "" {
		return nil, errors.New("agent_id is required")
	}

	actionUpper := strings.ToUpper(req.Action)
	switch actionUpper {
	case "ENABLE", "DISABLE", "PAUSE", "RESUME", "SET_AUTONOMY":
	default:
		return nil, fmt.Errorf("invalid control action '%s': supported: ENABLE, DISABLE, PAUSE, RESUME, SET_AUTONOMY", req.Action)
	}

	existingAgent, err := s.repo.GetAgent(ctx, orgID, agentID)
	if err != nil || existingAgent == nil {
		return nil, fmt.Errorf("agent '%s' not found for tenant: %w", agentID, err)
	}

	newEnabled := existingAgent.IsEnabled
	newHealth := existingAgent.HealthStatus
	newAutonomy := existingAgent.AutonomyLevel

	if actionUpper == "SET_AUTONOMY" {
		targetAutonomy := req.AutonomyLevel
		if targetAutonomy == "" && req.NewAutonomyLevel != nil {
			targetAutonomy = *req.NewAutonomyLevel
		}
		switch targetAutonomy {
		case AutonomyLevel0Observe, AutonomyLevel1Recommend, AutonomyLevel2Prepare, AutonomyLevel3ControlledExecution, AutonomyLevel4GovernedMultiStep:
			newAutonomy = targetAutonomy
		default:
			return nil, fmt.Errorf("invalid autonomy level '%s'", targetAutonomy)
		}
	} else if actionUpper == "ENABLE" {
		newEnabled = true
		newHealth = HealthStatusHealthy
	} else if actionUpper == "DISABLE" {
		newEnabled = false
		newHealth = HealthStatusDisabled
	} else if actionUpper == "PAUSE" {
		newHealth = AgentStatusPaused
	} else if actionUpper == "RESUME" {
		newHealth = HealthStatusHealthy
	}

	if err := s.repo.UpdateAgentControl(ctx, orgID, agentID, newEnabled, newHealth, newAutonomy); err != nil {
		return nil, fmt.Errorf("failed updating agent control: %w", err)
	}

	s.logAuditEvent(ctx, orgID, "WORKFORCE_AGENT_CONTROL_UPDATED", "WORKFORCE_AGENT", agentID, map[string]interface{}{
		"action":            actionUpper,
		"actor":             actor,
		"previous_status":   existingAgent.OperationalStatus,
		"previous_autonomy": existingAgent.AutonomyLevel,
		"new_autonomy":      req.AutonomyLevel,
		"reason":            req.Reason,
	})

	return s.repo.GetAgent(ctx, orgID, agentID)
}

func (s *workforceService) EmergencyStop(ctx context.Context, orgID int64, req EmergencyStopRequest, actor string) (*EmergencyStopStatus, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}

	actionUpper := strings.ToUpper(req.Action)
	if actionUpper == "" {
		actionUpper = "STOP"
	}
	scopeUpper := strings.ToUpper(req.Scope)
	if scopeUpper == "" {
		scopeUpper = EmergencyStopScopeWorkforce
	}

	switch scopeUpper {
	case EmergencyStopScopeWorkforce, EmergencyStopScopeAgent, EmergencyStopScopeWorkflow, EmergencyStopScopeActionClass:
	default:
		return nil, fmt.Errorf("invalid emergency stop scope '%s'", req.Scope)
	}

	nowStr := time.Now().UTC().Format(time.RFC3339)
	var status *EmergencyStopStatus

	if actionUpper == "STOP" {
		reason := req.Reason
		if reason == "" {
			reason = fmt.Sprintf("Emergency stop initiated by %s", actor)
		}
		status = s.emergStops.SetStop(orgID, SingleStop{
			Scope:       scopeUpper,
			Target:      req.Target,
			Reason:      reason,
			TriggeredBy: actor,
			TriggeredAt: nowStr,
		})
		s.logAuditEvent(ctx, orgID, "WORKFORCE_EMERGENCY_STOP_TRIGGERED", "WORKFORCE_SAFETY", req.Target, map[string]interface{}{
			"scope":        scopeUpper,
			"target":       req.Target,
			"reason":       reason,
			"triggered_by": actor,
		})
	} else if actionUpper == "RESUME" {
		status = s.emergStops.RemoveStop(orgID, scopeUpper, req.Target)
		s.logAuditEvent(ctx, orgID, "WORKFORCE_EMERGENCY_STOP_CLEARED", "WORKFORCE_SAFETY", req.Target, map[string]interface{}{
			"scope":      scopeUpper,
			"target":     req.Target,
			"cleared_by": actor,
			"reason":     req.Reason,
		})
	} else {
		return nil, fmt.Errorf("invalid action '%s': must be STOP or RESUME", req.Action)
	}

	return status, nil
}

func (s *workforceService) GetEmergencyStopStatus(ctx context.Context, orgID int64) (*EmergencyStopStatus, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	return s.emergStops.GetStatus(orgID), nil
}

func (s *workforceService) ControlWorkflow(ctx context.Context, orgID int64, planID string, req WorkflowControlRequest, actor string) (*CollaborativePlan, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if planID == "" {
		return nil, errors.New("plan_id is required")
	}

	plan, err := s.GetCollaborativePlan(ctx, orgID, planID)
	if err != nil {
		return nil, err
	}

	cmdUpper := strings.ToUpper(req.Command)
	nowStr := time.Now().UTC().Format(time.RFC3339)

	switch cmdUpper {
	case "STOP":
		plan.Status = "STOPPED"
		s.emergStops.SetStop(orgID, SingleStop{
			Scope:       EmergencyStopScopeWorkflow,
			Target:      planID,
			Reason:      req.Reason,
			TriggeredBy: actor,
			TriggeredAt: nowStr,
		})
		_ = s.repo.UpdateTaskStatus(ctx, orgID, plan.PlanningTaskID, TaskStatusCancelled, nil, 0, nil, nil)
	case "PAUSE":
		plan.Status = "PAUSED"
		_ = s.repo.UpdateTaskStatus(ctx, orgID, plan.PlanningTaskID, TaskStatusWaiting, nil, 0, nil, nil)
	case "RESUME":
		plan.Status = "IN_PROGRESS"
		s.emergStops.RemoveStop(orgID, EmergencyStopScopeWorkflow, planID)
		_ = s.repo.UpdateTaskStatus(ctx, orgID, plan.PlanningTaskID, TaskStatusRunning, nil, 0, nil, nil)
	default:
		return nil, fmt.Errorf("invalid workflow command '%s': supported: STOP, PAUSE, RESUME", req.Command)
	}

	s.savePlanToContext(ctx, orgID, plan.PlanningTaskID, plan)
	s.logAuditEvent(ctx, orgID, "WORKFORCE_WORKFLOW_CONTROL_UPDATED", "COLLABORATIVE_PLAN", planID, map[string]interface{}{
		"command": cmdUpper,
		"actor":   actor,
		"reason":  req.Reason,
		"status":  plan.Status,
	})

	return plan, nil
}

func (s *workforceService) GetWorkforceHealth(ctx context.Context, orgID int64) (*WorkforceHealthSummary, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	return s.repo.GetWorkforceHealth(ctx, orgID)
}

func (s *workforceService) GetAgentWorkload(ctx context.Context, orgID int64) ([]AgentWorkloadMetrics, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	return s.repo.GetAgentWorkload(ctx, orgID)
}

func (s *workforceService) ListWorkforceApprovals(ctx context.Context, orgID int64) ([]WorkforceApprovalItem, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	items := make([]WorkforceApprovalItem, 0)
	if s.approvalsSvc != nil {
		reqs, err := s.approvalsSvc.ListApprovals(ctx, orgID)
		if err == nil {
			for _, ap := range reqs {
				if ap.Status == approvals.StatusPendingApproval || ap.Status == approvals.StatusInReview || ap.Status == approvals.StatusReturnedForChanges {
					actName := ""
					if ap.ActionName != nil {
						actName = *ap.ActionName
					}
					reason := ap.Title
					if ap.Description != nil && *ap.Description != "" {
						reason = *ap.Description
					}
					items = append(items, WorkforceApprovalItem{
						ApprovalID:          ap.ID,
						WorkflowID:          actName,
						AgentID:             ap.RequestedByName,
						InitiatingAgent:     ap.RequestedByName,
						ProposedAction:      actName,
						Reason:              reason,
						Confidence:          0.88,
						Risk:                ap.RiskLevel,
						AffectedEntity:      fmt.Sprintf("%s:%d", actName, ap.ID),
						ApprovalRequirement: "Human approval required prior to irreversible action execution",
						Status:              ap.Status,
						CreatedAt:           ap.CreatedAt.UTC().Format(time.RFC3339),
					})
				}
			}
		}
	}
	return items, nil
}

func (s *workforceService) ListWorkforceEscalations(ctx context.Context, orgID int64) ([]WorkforceEscalationItem, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}

	escalations := make([]WorkforceEscalationItem, 0)

	// 1. Failed tasks that require human attention
	failedTasks, _, err := s.repo.ListTasks(ctx, TaskFilter{
		OrgID:  orgID,
		Status: string(TaskStatusFailed),
		Limit:  15,
	})
	if err == nil {
		for _, ft := range failedTasks {
			errStr := "Task execution failure"
			if ft.ErrorMessage != nil && *ft.ErrorMessage != "" {
				errStr = *ft.ErrorMessage
			}
			rTask := ""
			if ft.RootTaskID != nil {
				rTask = *ft.RootTaskID
			}
			escalations = append(escalations, WorkforceEscalationItem{
				EscalationID: fmt.Sprintf("esc-%s", ft.TaskID),
				WorkflowID:   rTask,
				AgentID:      ft.AssignedAgentID,
				TaskID:       ft.TaskID,
				Reason:       errStr,
				Severity:     "HIGH",
				Status:       "OPEN",
				Evidence: map[string]interface{}{
					"objective":  ft.Objective,
					"error_code": ft.ErrorCode,
				},
				CreatedAt: ft.CreatedAt.UTC().Format(time.RFC3339),
			})
		}
	}

	// 2. Active emergency stops
	st := s.emergStops.GetStatus(orgID)
	for _, sStop := range st.ActiveStops {
		escalations = append(escalations, WorkforceEscalationItem{
			EscalationID: fmt.Sprintf("stop-%s-%s", sStop.Scope, sStop.Target),
			WorkflowID:   sStop.Target,
			AgentID:      sStop.Target,
			TaskID:       "",
			Reason:       fmt.Sprintf("Emergency Stop Active [%s]: %s", sStop.Scope, sStop.Reason),
			Severity:     "CRITICAL",
			Status:       "BLOCKED",
			Evidence: map[string]interface{}{
				"scope":        sStop.Scope,
				"target":       sStop.Target,
				"triggered_by": sStop.TriggeredBy,
			},
			CreatedAt: sStop.TriggeredAt,
		})
	}

	return escalations, nil
}

func (s *workforceService) GetRecentActivity(ctx context.Context, orgID int64, limit int) ([]WorkforceActivityItem, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	activity := make([]WorkforceActivityItem, 0, limit)

	// Fetch recent tasks
	tasks, _, err := s.repo.ListTasks(ctx, TaskFilter{
		OrgID: orgID,
		Limit: limit,
	})
	if err == nil {
		for _, t := range tasks {
			summary := fmt.Sprintf("Task '%s' is %s", t.Objective, t.Status)
			activity = append(activity, WorkforceActivityItem{
				ActivityID: fmt.Sprintf("act-task-%s", t.TaskID),
				Timestamp:  t.CreatedAt.UTC().Format(time.RFC3339),
				EventType:  "TASK_" + string(t.Status),
				AgentID:    t.AssignedAgentID,
				TaskID:     t.TaskID,
				WorkflowID: "",
				Summary:    summary,
				Status:     string(t.Status),
			})
		}
	}

	// Fetch recent outcomes
	outcomes, err := s.repo.ListOutcomes(ctx, orgID, "", "", limit/2)
	if err == nil {
		for _, o := range outcomes {
			activity = append(activity, WorkforceActivityItem{
				ActivityID: fmt.Sprintf("act-out-%s", o.OutcomeID),
				Timestamp:  o.CreatedAt,
				EventType:  "OUTCOME_" + o.Status,
				AgentID:    o.SourceAgentID,
				TaskID:     o.SourceTaskID,
				WorkflowID: o.WorkflowID,
				Summary:    fmt.Sprintf("Outcome for %s: %s", o.SourceEntityType, o.Recommendation),
				Status:     o.Status,
			})
		}
	}

	return activity, nil
}

func (s *workforceService) InspectWorkflow(ctx context.Context, orgID int64, planID string) (*WorkflowInspectionDetail, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	plan, err := s.GetCollaborativePlan(ctx, orgID, planID)
	if err != nil {
		return nil, err
	}

	detail := &WorkflowInspectionDetail{
		PlanID:              plan.PlanID,
		Objective:           plan.Objective,
		CoordinatorAgentID:  plan.CoordinatorAgentID,
		ParticipatingAgents: plan.ParticipatingAgents,
		Status:              plan.Status,
		OverallConfidence:   plan.OverallConfidence,
		Steps:               plan.Steps,
		FinalDecision:       plan.FinalDecision,
		RequiresApproval:    plan.RequiresApproval,
		CreatedAt:           plan.CreatedAt,
	}

	if s.approvalsSvc != nil {
		approvalsList, _ := s.approvalsSvc.ListApprovals(ctx, orgID)
		for _, ap := range approvalsList {
			actName := ""
			if ap.ActionName != nil {
				actName = *ap.ActionName
			}
			if strings.Contains(actName, planID) || ap.Title == plan.Objective {
				statusStr := ap.Status
				detail.ApprovalStatus = &statusStr
				break
			}
		}
	}

	return detail, nil
}

func (s *workforceService) GetCommandCenterOverview(ctx context.Context, orgID int64) (*CommandCenterOverview, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}

	health, err := s.GetWorkforceHealth(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed fetching workforce health: %w", err)
	}

	workload, err := s.GetAgentWorkload(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed fetching agent workload: %w", err)
	}

	stopStatus, _ := s.GetEmergencyStopStatus(ctx, orgID)
	approvalsList, _ := s.ListWorkforceApprovals(ctx, orgID)
	escalationsList, _ := s.ListWorkforceEscalations(ctx, orgID)
	recentActivity, _ := s.GetRecentActivity(ctx, orgID, 15)

	autonomyCounts := make(map[string]int)
	for _, a := range workload {
		level := a.AutonomyLevel
		if level == "" {
			level = AutonomyLevel1Recommend
		}
		autonomyCounts[level]++
	}

	activeWorkflows := 0
	for _, a := range workload {
		activeWorkflows += a.RunningTasks
	}

	return &CommandCenterOverview{
		Health:           health,
		AgentWorkload:    workload,
		EmergencyStop:    stopStatus,
		ActiveWorkflows:  activeWorkflows,
		WaitingApprovals: approvalsList,
		Escalations:      escalationsList,
		RecentActivity:   recentActivity,
		AutonomySummary:  autonomyCounts,
	}, nil
}
