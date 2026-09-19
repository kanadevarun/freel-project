package enterprise_autonomy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/freel/backend/internal/actions"
	"github.com/freel/backend/internal/approvals"
	"github.com/freel/backend/internal/audit/domain"
	auditSvc "github.com/freel/backend/internal/audit/service"
	"github.com/freel/backend/internal/workforce"
)

type Service interface {
	StartWorkflow(ctx context.Context, orgID int64, req StartWorkflowRequest) (*EnterpriseWorkflow, error)
	GetWorkflow(ctx context.Context, orgID int64, workflowID string) (*EnterpriseWorkflow, error)
	ListWorkflows(ctx context.Context, filter WorkflowFilter) ([]*EnterpriseWorkflow, int, error)
	PauseWorkflow(ctx context.Context, orgID int64, workflowID string, req PauseWorkflowRequest) error
	ResumeWorkflow(ctx context.Context, orgID int64, workflowID string) error
	CancelWorkflow(ctx context.Context, orgID int64, workflowID string, req CancelWorkflowRequest) error
	ApproveWorkflowStep(ctx context.Context, orgID int64, workflowID, stepID string, approverID int64, req ApproveStepRequest) error
	RejectWorkflowStep(ctx context.Context, orgID int64, workflowID, stepID string, approverID int64, req RejectStepRequest) error
	TriggerBusinessEvent(ctx context.Context, orgID int64, req TriggerBusinessEventRequest) (*EnterpriseWorkflow, error)
	RecoverInterruptedWorkflows(ctx context.Context) (*RecoveryReport, error)
	EmergencyControl(ctx context.Context, orgID int64, req EmergencyControlRequest, actor string) error
	GetPlatformOverview(ctx context.Context, orgID int64) (*EnterprisePlatformOverview, error)
	SetResilienceService(svc EnterpriseResilienceService)
}

type defaultService struct {
	repo          Repository
	workforceSvc  workforce.Service
	actionsSvc    actions.Service
	approvalsSvc  approvals.Service
	auditSvc      auditSvc.Service
	resilienceSvc EnterpriseResilienceService
	mu            sync.RWMutex
}

func NewService(
	repo Repository,
	workforceSvc workforce.Service,
	actionsSvc actions.Service,
	approvalsSvc approvals.Service,
	auditService auditSvc.Service,
) Service {
	return &defaultService{
		repo:         repo,
		workforceSvc: workforceSvc,
		actionsSvc:   actionsSvc,
		approvalsSvc: approvalsSvc,
		auditSvc:     auditService,
	}
}

func (s *defaultService) SetResilienceService(svc EnterpriseResilienceService) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resilienceSvc = svc
}

func (s *defaultService) StartWorkflow(ctx context.Context, orgID int64, req StartWorkflowRequest) (*EnterpriseWorkflow, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if req.Objective == "" {
		return nil, errors.New("workflow objective is required")
	}

	// 1. Fetch Organization Autonomy Policy Context
	moduleStr := string(req.WorkflowType)
	if moduleStr == "" {
		moduleStr = "CROSS_MODULE"
	}
	policy, err := s.repo.GetPolicyContext(ctx, orgID, moduleStr)
	if err != nil {
		return nil, fmt.Errorf("failed fetching enterprise policy context: %w", err)
	}

	// Check emergency stop
	if policy.EmergencyStopActive {
		return nil, ErrEmergencyStopActive
	}

	// Level 0 Observe: Cannot initiate active autonomous workflows
	effectiveAutonomy := policy.AutonomyLevel
	if req.AutonomyLevel != "" && req.AutonomyLevel != policy.AutonomyLevel {
		// AI or client request cannot elevate autonomy above organization policy
		if isElevationAttempt(req.AutonomyLevel, policy.AutonomyLevel) {
			s.logSecurityEvent(ctx, orgID, "AUTONOMY_ELEVATION_REJECTED", map[string]interface{}{
				"requested": req.AutonomyLevel,
				"effective": policy.AutonomyLevel,
				"objective": req.Objective,
			})
			return nil, ErrAutonomyRestricted
		}
		effectiveAutonomy = req.AutonomyLevel
	}

	if effectiveAutonomy == "LEVEL_0_OBSERVE" {
		return nil, fmt.Errorf("%w: LEVEL_0_OBSERVE is read-only and cannot start autonomous workflows", ErrAutonomyRestricted)
	}

	// Validate allowed workflow type
	if len(policy.AllowedWorkflowTypes) > 0 {
		allowed := false
		for _, wt := range policy.AllowedWorkflowTypes {
			if string(req.WorkflowType) == wt || wt == "*" {
				allowed = true
				break
			}
		}
		if !allowed {
			s.logSecurityEvent(ctx, orgID, "DISALLOWED_WORKFLOW_TYPE", map[string]interface{}{
				"workflow_type": req.WorkflowType,
				"allowed":       policy.AllowedWorkflowTypes,
			})
			return nil, ErrAutonomyRestricted
		}
	}

	// Generate durable workflow identity
	workflowID := generateID("ent-wf")
	corrID := req.CorrelationID
	if corrID == "" {
		corrID = generateID("corr")
	}

	wf := &EnterpriseWorkflow{
		WorkflowID:        workflowID,
		OrgID:             orgID,
		WorkflowType:      req.WorkflowType,
		Objective:         req.Objective,
		InitiatingEvent:   req.InitiatingEvent,
		CurrentState:      StateRunning,
		AutonomyLevel:     effectiveAutonomy,
		PolicyDecision:    "PERMITTED",
		PolicyContext:     policy,
		CorrelationID:     corrID,
		IdempotencyKey:    req.IdempotencyKey,
		RelatedEntityType: req.RelatedEntityType,
		RelatedEntityID:   req.RelatedEntityID,
		Confidence:        0.90,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	// 2. Cross-Module Orchestration via Phase 6 Workforce Specialists
	var planSteps []EnterpriseWorkflowStep
	var assignedAgents []string

	switch req.WorkflowType {
	case WorkflowShipmentRecovery:
		assignedAgents = []string{"planning_agent", "shipment_agent", "exception_agent", "customer_agent", "finance_agent"}
		planSteps = s.buildShipmentRecoverySteps(workflowID, orgID, req.Objective)
	case WorkflowCommercialCycle:
		assignedAgents = []string{"planning_agent", "customer_agent", "pricing_agent", "finance_agent", "contract_agent"}
		planSteps = s.buildCommercialCycleSteps(workflowID, orgID, req.Objective)
	case WorkflowFinancialCollection:
		assignedAgents = []string{"planning_agent", "finance_agent", "customer_agent", "memory_agent"}
		planSteps = s.buildFinancialCollectionSteps(workflowID, orgID, req.Objective)
	case WorkflowCrossModuleRisk:
		assignedAgents = []string{"planning_agent", "shipment_agent", "contract_agent", "compliance_agent", "finance_agent"}
		planSteps = s.buildCrossModuleRiskSteps(workflowID, orgID, req.Objective)
	default:
		assignedAgents = []string{"planning_agent", "shipment_agent", "exception_agent"}
		planSteps = s.buildGenericSteps(workflowID, orgID, req.Objective)
	}

	wf.AssignedAgents = assignedAgents
	wf.Steps = planSteps
	if len(planSteps) > 0 {
		wf.CurrentStep = planSteps[0].StepID
	}

	// 3. Persist Durable Workflow Record and Steps
	if err := s.repo.CreateWorkflow(ctx, wf); err != nil {
		return nil, fmt.Errorf("failed persisting enterprise workflow: %w", err)
	}
	if err := s.repo.SaveWorkflowSteps(ctx, orgID, workflowID, planSteps); err != nil {
		return nil, fmt.Errorf("failed persisting enterprise workflow steps: %w", err)
	}

	// 4. Audit Log Workflow Initiation
	s.logAuditEvent(ctx, orgID, "ENTERPRISE_WORKFLOW_STARTED", workflowID, map[string]interface{}{
		"workflow_type":  string(req.WorkflowType),
		"objective":      req.Objective,
		"autonomy_level": effectiveAutonomy,
		"step_count":     len(planSteps),
		"correlation_id": corrID,
	})

	// 5. Evaluate and Execute First Step
	_ = s.executeWorkflowSteps(ctx, wf)

	// Reload fresh persisted state
	fresh, err := s.repo.GetWorkflow(ctx, orgID, workflowID)
	if err == nil {
		return fresh, nil
	}
	return wf, nil
}

func (s *defaultService) GetWorkflow(ctx context.Context, orgID int64, workflowID string) (*EnterpriseWorkflow, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	return s.repo.GetWorkflow(ctx, orgID, workflowID)
}

func (s *defaultService) ListWorkflows(ctx context.Context, filter WorkflowFilter) ([]*EnterpriseWorkflow, int, error) {
	if filter.OrgID <= 0 {
		return nil, 0, ErrUnauthorizedTenant
	}
	return s.repo.ListWorkflows(ctx, filter)
}

func (s *defaultService) PauseWorkflow(ctx context.Context, orgID int64, workflowID string, req PauseWorkflowRequest) error {
	if orgID <= 0 {
		return ErrUnauthorizedTenant
	}

	wf, err := s.repo.GetWorkflow(ctx, orgID, workflowID)
	if err != nil {
		return err
	}

	if err := ValidateStateTransition(wf.CurrentState, StatePaused); err != nil {
		return fmt.Errorf("cannot pause workflow in state %s: %w", wf.CurrentState, err)
	}

	msg := req.Reason
	if msg == "" {
		msg = "Workflow paused by operator"
	}

	if err := s.repo.UpdateWorkflowState(ctx, orgID, workflowID, StatePaused, nil, &msg); err != nil {
		return err
	}

	s.logAuditEvent(ctx, orgID, "ENTERPRISE_WORKFLOW_PAUSED", workflowID, map[string]interface{}{
		"reason": msg,
	})
	return nil
}

func (s *defaultService) ResumeWorkflow(ctx context.Context, orgID int64, workflowID string) error {
	if orgID <= 0 {
		return ErrUnauthorizedTenant
	}

	wf, err := s.repo.GetWorkflow(ctx, orgID, workflowID)
	if err != nil {
		return err
	}

	if err := ValidateStateTransition(wf.CurrentState, StateRunning); err != nil {
		return fmt.Errorf("cannot resume workflow from state %s: %w", wf.CurrentState, err)
	}

	if err := s.repo.UpdateWorkflowState(ctx, orgID, workflowID, StateRunning, nil, nil); err != nil {
		return err
	}

	s.logAuditEvent(ctx, orgID, "ENTERPRISE_WORKFLOW_RESUMED", workflowID, nil)

	// Continue step execution
	wf.CurrentState = StateRunning
	go func() {
		_ = s.executeWorkflowSteps(context.Background(), wf)
	}()

	return nil
}

func (s *defaultService) CancelWorkflow(ctx context.Context, orgID int64, workflowID string, req CancelWorkflowRequest) error {
	if orgID <= 0 {
		return ErrUnauthorizedTenant
	}

	wf, err := s.repo.GetWorkflow(ctx, orgID, workflowID)
	if err != nil {
		return err
	}

	if err := ValidateStateTransition(wf.CurrentState, StateCancelled); err != nil {
		return fmt.Errorf("cannot cancel workflow from state %s: %w", wf.CurrentState, err)
	}

	msg := req.Reason
	if msg == "" {
		msg = "Workflow cancelled by operator"
	}

	if err := s.repo.UpdateWorkflowState(ctx, orgID, workflowID, StateCancelled, nil, &msg); err != nil {
		return err
	}

	s.logAuditEvent(ctx, orgID, "ENTERPRISE_WORKFLOW_CANCELLED", workflowID, map[string]interface{}{
		"reason": msg,
	})
	return nil
}

func (s *defaultService) ApproveWorkflowStep(ctx context.Context, orgID int64, workflowID, stepID string, approverID int64, req ApproveStepRequest) error {
	if orgID <= 0 {
		return ErrUnauthorizedTenant
	}

	wf, err := s.repo.GetWorkflow(ctx, orgID, workflowID)
	if err != nil {
		return err
	}

	steps, err := s.repo.GetWorkflowSteps(ctx, orgID, workflowID)
	if err != nil {
		return err
	}

	var targetStep *EnterpriseWorkflowStep
	for i := range steps {
		if steps[i].StepID == stepID {
			targetStep = &steps[i]
			break
		}
	}
	if targetStep == nil {
		return errors.New("workflow step not found")
	}

	// Mark step as completed / approved
	targetStep.Status = "COMPLETED"
	now := time.Now().UTC()
	targetStep.ExecutedAt = &now
	targetStep.ExecutionResult = map[string]interface{}{
		"approved_by": approverID,
		"notes":       req.Notes,
		"approved_at": now.Format(time.RFC3339),
		"status":      "HUMAN_APPROVED",
	}

	if err := s.repo.UpdateWorkflowStep(ctx, orgID, targetStep); err != nil {
		return err
	}

	s.logAuditEvent(ctx, orgID, "ENTERPRISE_WORKFLOW_STEP_APPROVED", workflowID, map[string]interface{}{
		"step_id":     stepID,
		"approver_id": approverID,
		"notes":       req.Notes,
	})

	// Resume workflow execution
	if wf.CurrentState == StateWaitingForApproval {
		_ = s.repo.UpdateWorkflowState(ctx, orgID, workflowID, StateRunning, nil, nil)
		wf.CurrentState = StateRunning
		_ = s.executeWorkflowSteps(ctx, wf)
	}

	return nil
}

func (s *defaultService) RejectWorkflowStep(ctx context.Context, orgID int64, workflowID, stepID string, approverID int64, req RejectStepRequest) error {
	if orgID <= 0 {
		return ErrUnauthorizedTenant
	}

	_, err := s.repo.GetWorkflow(ctx, orgID, workflowID)
	if err != nil {
		return err
	}

	steps, err := s.repo.GetWorkflowSteps(ctx, orgID, workflowID)
	if err != nil {
		return err
	}

	var targetStep *EnterpriseWorkflowStep
	for i := range steps {
		if steps[i].StepID == stepID {
			targetStep = &steps[i]
			break
		}
	}
	if targetStep == nil {
		return errors.New("workflow step not found")
	}

	// Mark step as failed / rejected
	reason := req.Reason
	if reason == "" {
		reason = "Step proposal rejected by human approver"
	}
	targetStep.Status = "FAILED"
	targetStep.ErrorMessage = &reason
	now := time.Now().UTC()
	targetStep.ExecutedAt = &now
	targetStep.ExecutionResult = map[string]interface{}{
		"rejected_by": approverID,
		"reason":      reason,
		"rejected_at": now.Format(time.RFC3339),
		"status":      "HUMAN_REJECTED",
	}

	_ = s.repo.UpdateWorkflowStep(ctx, orgID, targetStep)

	// Transition workflow to BLOCKED or ESCALATED
	_ = s.repo.UpdateWorkflowState(ctx, orgID, workflowID, StateBlocked, nil, &reason)

	s.logAuditEvent(ctx, orgID, "ENTERPRISE_WORKFLOW_STEP_REJECTED", workflowID, map[string]interface{}{
		"step_id":     stepID,
		"approver_id": approverID,
		"reason":      reason,
	})

	return nil
}

func (s *defaultService) TriggerBusinessEvent(ctx context.Context, orgID int64, req TriggerBusinessEventRequest) (*EnterpriseWorkflow, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if req.EventType == "" {
		return nil, errors.New("event_type is required")
	}

	// 1. Event Deduplication Check
	dedupKey := req.IdempotencyKey
	if dedupKey == "" && req.EventID != "" {
		dedupKey = fmt.Sprintf("evt-dedup-%s-%s", req.EventType, req.EventID)
	}
	if dedupKey != "" {
		isNew, err := s.repo.CheckAndRecordEventDedup(ctx, orgID, dedupKey, req.EventType)
		if err != nil {
			return nil, err
		}
		if !isNew {
			return nil, ErrDuplicateEventTrigger
		}
	}

	// 2. Map event to enterprise workflow type
	var wfType EnterpriseWorkflowType
	var objective string

	switch req.EventType {
	case "SHIPMENT_DELAY_DETECTED", "TEMPERATURE_EXCURSION", "PORT_CONGESTION":
		wfType = WorkflowShipmentRecovery
		objective = fmt.Sprintf("Autonomous operational recovery for shipment %s following event %s", req.EntityID, req.EventType)
	case "RFQ_RECEIVED", "QUOTE_REQUESTED":
		wfType = WorkflowCommercialCycle
		objective = fmt.Sprintf("Autonomous commercial pricing and margin evaluation for RFQ %s", req.EntityID)
	case "INVOICE_OVERDUE", "COLLECTION_TRIGGER":
		wfType = WorkflowFinancialCollection
		objective = fmt.Sprintf("Autonomous collections assessment and receivable triage for %s", req.EntityID)
	case "CONTRACT_EXPIRATION_APPROACHING", "DEMURRAGE_EXPOSURE_DETECTED":
		wfType = WorkflowCrossModuleRisk
		objective = fmt.Sprintf("Cross-module risk assessment for contract %s and shipment %s", req.EntityID, req.Payload["shipment_id"])
	default:
		wfType = WorkflowOperationalRecovery
		objective = fmt.Sprintf("Event-driven autonomous workflow for %s on %s", req.EventType, req.EntityID)
	}

	startReq := StartWorkflowRequest{
		WorkflowType:      wfType,
		Objective:         objective,
		InitiatingEvent:   req.EventType,
		RelatedEntityType: req.EntityType,
		RelatedEntityID:   req.EntityID,
		InitialContext:    req.Payload,
		CorrelationID:     req.CorrelationID,
		IdempotencyKey:    dedupKey,
	}

	return s.StartWorkflow(ctx, orgID, startReq)
}

func (s *defaultService) RecoverInterruptedWorkflows(ctx context.Context) (*RecoveryReport, error) {
	report := &RecoveryReport{
		ResumedIDs: []string{},
		BlockedIDs: []string{},
		Errors:     []string{},
	}

	interrupted, err := s.repo.GetActiveOrInterruptedWorkflows(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed scanning interrupted workflows: %w", err)
	}

	report.ScannedCount = len(interrupted)

	for _, wf := range interrupted {
		// Verify policy context
		policy, err := s.repo.GetPolicyContext(ctx, wf.OrgID, string(wf.WorkflowType))
		if err != nil || (policy != nil && policy.EmergencyStopActive) {
			// Cannot resume under active emergency stop
			msg := "Recovery blocked: emergency stop is active for organization"
			_ = s.repo.UpdateWorkflowState(ctx, wf.OrgID, wf.WorkflowID, StateBlocked, nil, &msg)
			report.BlockedCount++
			report.BlockedIDs = append(report.BlockedIDs, wf.WorkflowID)
			continue
		}

		if wf.CurrentState == StateWaitingForApproval {
			// Leave in approval wait state safely
			continue
		}

		// Resume workflow execution safely and idempotently
		err = s.executeWorkflowSteps(ctx, wf)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", wf.WorkflowID, err))
		} else {
			report.RecoveredCount++
			report.ResumedIDs = append(report.ResumedIDs, wf.WorkflowID)
		}
	}

	return report, nil
}

func (s *defaultService) EmergencyControl(ctx context.Context, orgID int64, req EmergencyControlRequest, actor string) error {
	if orgID <= 0 {
		return ErrUnauthorizedTenant
	}

	policy, err := s.repo.GetPolicyContext(ctx, orgID, "CROSS_MODULE")
	if err != nil {
		return err
	}

	switch req.Action {
	case "PAUSE", "HALT":
		policy.EmergencyStopActive = true
		_ = s.repo.UpdatePolicyContext(ctx, orgID, policy)
		s.logAuditEvent(ctx, orgID, "ENTERPRISE_EMERGENCY_HALT_ACTIVATED", "SYSTEM", map[string]interface{}{
			"scope":  req.Scope,
			"reason": req.Reason,
			"actor":  actor,
		})
	case "RESUME":
		policy.EmergencyStopActive = false
		_ = s.repo.UpdatePolicyContext(ctx, orgID, policy)
		s.logAuditEvent(ctx, orgID, "ENTERPRISE_EMERGENCY_HALT_DEACTIVATED", "SYSTEM", map[string]interface{}{
			"scope": req.Scope,
			"actor": actor,
		})
	default:
		return errors.New("unsupported emergency action (must be PAUSE or RESUME)")
	}

	return nil
}

func (s *defaultService) GetPlatformOverview(ctx context.Context, orgID int64) (*EnterprisePlatformOverview, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	return s.repo.GetPlatformOverview(ctx, orgID)
}

// ---------------------------------------------------------------------
// Internal Execution Engine & Step Handlers
// ---------------------------------------------------------------------

func (s *defaultService) executeWorkflowSteps(ctx context.Context, wf *EnterpriseWorkflow) error {
	steps, err := s.repo.GetWorkflowSteps(ctx, wf.OrgID, wf.WorkflowID)
	if err != nil {
		return err
	}

	completedCount := 0
	for i := range steps {
		step := &steps[i]

		if step.Status == "COMPLETED" || step.Status == "SKIPPED" {
			completedCount++
			continue
		}

		if step.Status == "WAITING_FOR_APPROVAL" {
			_ = s.repo.UpdateWorkflowState(ctx, wf.OrgID, wf.WorkflowID, StateWaitingForApproval, nil, nil)
			return nil
		}

		// Check step dependencies
		if !areDependenciesMet(step.Dependencies, steps) {
			return nil
		}

		// Check autonomy policy for this step action
		if step.RequiresApproval || wf.AutonomyLevel == "LEVEL_2_PREPARE" || isHighRiskAction(step.ActionType) {
			step.Status = "WAITING_FOR_APPROVAL"
			_ = s.repo.UpdateWorkflowStep(ctx, wf.OrgID, step)
			_ = s.repo.UpdateWorkflowState(ctx, wf.OrgID, wf.WorkflowID, StateWaitingForApproval, nil, nil)
			s.logAuditEvent(ctx, wf.OrgID, "ENTERPRISE_APPROVAL_REQUIRED", wf.WorkflowID, map[string]interface{}{
				"step_id":     step.StepID,
				"action_type": step.ActionType,
				"risk_level":  step.RiskLevel,
			})
			return nil
		}

		// Execute permitted step via Action System boundary (Level 3 or Level 4)
		step.Status = "RUNNING"
		now := time.Now().UTC()
		step.ExecutedAt = &now
		step.ExecutionAttempt++

		if s.actionsSvc != nil && s.isExecutableAction(step.ActionType) {
			actResp, err := s.actionsSvc.Execute(ctx, actions.ActionExecutionRequest{
				ActionName:     step.ActionType,
				OrgID:          wf.OrgID,
				ActorType:      actions.ActorTypeAIAgent,
				Source:         "enterprise_autonomous_platform",
				TaskID:         wf.WorkflowID,
				IdempotencyKey: step.IdempotencyKey,
				Input:          step.Parameters,
			})
			if err != nil {
				errMsg := err.Error()
				cls := FailureTypePermanent
				if s.resilienceSvc != nil {
					cls = s.resilienceSvc.ClassifyFailure(err, nil)
				}
				step.Status = "FAILED"
				step.ErrorMessage = &errMsg
				_ = s.repo.UpdateWorkflowStep(ctx, wf.OrgID, step)

				shouldRetry, _ := false, time.Duration(0)
				if s.resilienceSvc != nil {
					shouldRetry, _ = s.resilienceSvc.ShouldRetry(cls, step.ExecutionAttempt, step.MaxAttempts)
				}

				if shouldRetry {
					retryMsg := fmt.Sprintf("Step failed (%s: %s). Scheduled for governed retry.", cls, errMsg)
					_ = s.repo.UpdateWorkflowState(ctx, wf.OrgID, wf.WorkflowID, StateWaiting, nil, &retryMsg)
				} else {
					_ = s.repo.UpdateWorkflowState(ctx, wf.OrgID, wf.WorkflowID, StateFailed, nil, &errMsg)
				}
				return err
			}

			if actResp.ConfirmationRequired {
				step.Status = "WAITING_FOR_APPROVAL"
				step.RequiresApproval = true
				step.ExecutionResult = map[string]interface{}{
					"confirmation_required": true,
					"approval_reference":   actResp.ApprovalReference,
					"correlation_id":       actResp.CorrelationID,
				}
				_ = s.repo.UpdateWorkflowStep(ctx, wf.OrgID, step)
				_ = s.repo.UpdateWorkflowState(ctx, wf.OrgID, wf.WorkflowID, StateWaitingForApproval, nil, nil)
				if s.resilienceSvc != nil {
					_ = s.resilienceSvc.SaveCheckpoint(ctx, wf.OrgID, WorkflowCheckpoint{
						WorkflowID: wf.WorkflowID,
						OrgID:      wf.OrgID,
						Milestone:  CheckpointApprovalRequested,
						StepIndex:  i,
						CreatedAt:  time.Now().UTC(),
					})
				}
				return nil
			}

			if !actResp.Success {
				errMsg := "Action execution failed"
				if actResp.Error != nil {
					errMsg = actResp.Error.Message
				}
				step.Status = "FAILED"
				step.ErrorMessage = &errMsg
				_ = s.repo.UpdateWorkflowStep(ctx, wf.OrgID, step)
				_ = s.repo.UpdateWorkflowState(ctx, wf.OrgID, wf.WorkflowID, StateFailed, nil, &errMsg)
				return errors.New(errMsg)
			}

			step.Status = "COMPLETED"
			step.ActionSystemActionID = &actResp.CorrelationID
			step.ExecutionResult = map[string]interface{}{
				"action_name":       actResp.ActionName,
				"correlation_id":   actResp.CorrelationID,
				"idempotent_replay": actResp.IdempotentReplay,
				"data":             actResp.Data,
			}
		} else {
			// Simulated/read-only autonomous step completion
			step.Status = "COMPLETED"
			step.ExecutionResult = map[string]interface{}{
				"status":    "COMPLETED",
				"simulated": true,
			}
		}

		_ = s.repo.UpdateWorkflowStep(ctx, wf.OrgID, step)
		completedCount++

		if s.resilienceSvc != nil {
			_ = s.resilienceSvc.SaveCheckpoint(ctx, wf.OrgID, WorkflowCheckpoint{
				WorkflowID:     wf.WorkflowID,
				OrgID:          wf.OrgID,
				Milestone:      CheckpointActionExecuted,
				StepIndex:      completedCount,
				CompletedSteps: []string{step.StepID},
				CreatedAt:      time.Now().UTC(),
			})
		}
	}

	if completedCount == len(steps) {
		_ = s.repo.UpdateWorkflowState(ctx, wf.OrgID, wf.WorkflowID, StateCompleted, nil, nil)
		if s.resilienceSvc != nil {
			_ = s.resilienceSvc.SaveCheckpoint(ctx, wf.OrgID, WorkflowCheckpoint{
				WorkflowID: wf.WorkflowID,
				OrgID:      wf.OrgID,
				Milestone:  CheckpointVerificationComplete,
				StepIndex:  completedCount,
				CreatedAt:  time.Now().UTC(),
			})
		}
		s.logAuditEvent(ctx, wf.OrgID, "ENTERPRISE_WORKFLOW_COMPLETED", wf.WorkflowID, map[string]interface{}{
			"step_count": len(steps),
		})
	}

	return nil
}

// Helper builders for standard cross-module workflows

func (s *defaultService) buildShipmentRecoverySteps(workflowID string, orgID int64, objective string) []EnterpriseWorkflowStep {
	return []EnterpriseWorkflowStep{
		{
			StepID:           "step-1-telemetry",
			WorkflowID:       workflowID,
			OrgID:            orgID,
			StepNumber:       1,
			AgentID:          "shipment_agent",
			ActionType:       "CARRIER_TELEMETRY_REFRESH",
			Title:            "Refresh Carrier AIS & Telemetry",
			Description:      "Verify vessel AIS position, speed, and container sensor temperature readings",
			ExpectedOutcome:  "Current operational coordinates and temperature telemetry",
			RiskLevel:        "LOW",
			RequiresApproval: false,
			Status:           "PENDING",
			Dependencies:     []string{},
		},
		{
			StepID:           "step-2-exception",
			WorkflowID:       workflowID,
			OrgID:            orgID,
			StepNumber:       2,
			AgentID:          "exception_agent",
			ActionType:       "TRIAGE_EXCEPTION_IMPACT",
			Title:            "Triage Port Congestion & Disruption",
			Description:      "Calculate ETA delay hours, demurrage free time impact, and root cause",
			ExpectedOutcome:  "Disruption classification and risk severity",
			RiskLevel:        "LOW",
			RequiresApproval: false,
			Status:           "PENDING",
			Dependencies:     []string{"step-1-telemetry"},
		},
		{
			StepID:           "step-3-notification",
			WorkflowID:       workflowID,
			OrgID:            orgID,
			StepNumber:       3,
			AgentID:          "customer_agent",
			ActionType:       "CUSTOMER_NOTIFICATION_DISPATCH",
			Title:            "Prepare Customer Revised ETA Notification",
			Description:      "Draft proactive notification explaining schedule adjustment and preservation safeguards",
			ExpectedOutcome:  "Customer advisory dispatched upon review",
			RiskLevel:        "MEDIUM",
			RequiresApproval: true,
			Status:           "PENDING",
			Dependencies:     []string{"step-2-exception"},
		},
		{
			StepID:           "step-4-demurrage",
			WorkflowID:       workflowID,
			OrgID:            orgID,
			StepNumber:       4,
			AgentID:          "finance_agent",
			ActionType:       "DEMURRAGE_EXPOSURE_AUDIT",
			Title:            "Audit Port Storage & Demurrage Liability",
			Description:      "Reconcile contracted terminal free days against projected container discharge date",
			ExpectedOutcome:  "Estimated financial liability calculation",
			RiskLevel:        "LOW",
			RequiresApproval: false,
			Status:           "PENDING",
			Dependencies:     []string{"step-2-exception"},
		},
		{
			StepID:           "step-5-compliance",
			WorkflowID:       workflowID,
			OrgID:            orgID,
			StepNumber:       5,
			AgentID:          "compliance_agent",
			ActionType:       "COMPLIANCE_DISCREPANCY_CHECK",
			Title:            "Audit Customs & Regulatory Status",
			Description:      "Verify customs filings, seal numbers, and regulatory clearance paperwork",
			ExpectedOutcome:  "Regulatory risk clearance certification",
			RiskLevel:        "LOW",
			RequiresApproval: false,
			Status:           "PENDING",
			Dependencies:     []string{"step-1-telemetry"},
		},
	}
}

func (s *defaultService) buildCommercialCycleSteps(workflowID string, orgID int64, objective string) []EnterpriseWorkflowStep {
	return []EnterpriseWorkflowStep{
		{
			StepID:           "step-1-rfq-parse",
			WorkflowID:       workflowID,
			OrgID:            orgID,
			StepNumber:       1,
			AgentID:          "customer_agent",
			ActionType:       "PARSE_RFQ_REQUIREMENTS",
			Title:            "Extract Customer RFQ Terms & Lane",
			Description:      "Extract origins, destinations, container types, and delivery deadlines",
			ExpectedOutcome:  "Structured RFQ specification record",
			RiskLevel:        "LOW",
			RequiresApproval: false,
			Status:           "PENDING",
			Dependencies:     []string{},
		},
		{
			StepID:           "step-2-pricing",
			WorkflowID:       workflowID,
			OrgID:            orgID,
			StepNumber:       2,
			AgentID:          "pricing_agent",
			ActionType:       "OPTIMIZE_QUOTE_MARGIN",
			Title:            "Calculate Dynamic Corridor Rate & Margin",
			Description:      "Evaluate carrier buy rate, historical win probabilities, and benchmark tariffs",
			ExpectedOutcome:  "Optimized sell quote recommendation",
			RiskLevel:        "LOW",
			RequiresApproval: false,
			Status:           "PENDING",
			Dependencies:     []string{"step-1-rfq-parse"},
		},
		{
			StepID:           "step-3-credit",
			WorkflowID:       workflowID,
			OrgID:            orgID,
			StepNumber:       3,
			AgentID:          "finance_agent",
			ActionType:       "CUSTOMER_CREDIT_CHECK",
			Title:            "Assess Customer Credit & Outstanding Receivables",
			Description:      "Check credit limit headroom and overdue invoice aging prior to formal quotation",
			ExpectedOutcome:  "Credit approval verification",
			RiskLevel:        "MEDIUM",
			RequiresApproval: true,
			Status:           "PENDING",
			Dependencies:     []string{"step-2-pricing"},
		},
		{
			StepID:           "step-4-contract",
			WorkflowID:       workflowID,
			OrgID:            orgID,
			StepNumber:       4,
			AgentID:          "contract_agent",
			ActionType:       "GENERATE_QUOTE_CONTRACT_PROPOSAL",
			Title:            "Generate Formal Quotation Agreement",
			Description:      "Compile rate schedule, service level agreement terms, and validity expiry",
			ExpectedOutcome:  "Formal quote agreement document draft",
			RiskLevel:        "LOW",
			RequiresApproval: false,
			Status:           "PENDING",
			Dependencies:     []string{"step-3-credit"},
		},
	}
}

func (s *defaultService) buildFinancialCollectionSteps(workflowID string, orgID int64, objective string) []EnterpriseWorkflowStep {
	return []EnterpriseWorkflowStep{
		{
			StepID:           "step-1-invoice-audit",
			WorkflowID:       workflowID,
			OrgID:            orgID,
			StepNumber:       1,
			AgentID:          "finance_agent",
			ActionType:       "AUDIT_AGING_RECEIVABLES",
			Title:            "Audit Overdue Freight Invoices",
			Description:      "Identify invoices past payment terms and calculate interest charges",
			ExpectedOutcome:  "Aged receivables schedule",
			RiskLevel:        "LOW",
			RequiresApproval: false,
			Status:           "PENDING",
			Dependencies:     []string{},
		},
		{
			StepID:           "step-2-collections",
			WorkflowID:       workflowID,
			OrgID:            orgID,
			StepNumber:       2,
			AgentID:          "customer_agent",
			ActionType:       "COLLECTION_REMINDER_DISPATCH",
			Title:            "Draft Collection Reminder Notice",
			Description:      "Generate structured payment reminder calibrated to customer relationship profile",
			ExpectedOutcome:  "Reminder notice ready for approval",
			RiskLevel:        "MEDIUM",
			RequiresApproval: true,
			Status:           "PENDING",
			Dependencies:     []string{"step-1-invoice-audit"},
		},
		{
			StepID:           "step-3-contract-clauses",
			WorkflowID:       workflowID,
			OrgID:            orgID,
			StepNumber:       3,
			AgentID:          "contract_agent",
			ActionType:       "AUDIT_CREDIT_DEFAULT_CLAUSES",
			Title:            "Verify Contractual Credit Suspension Rights",
			Description:      "Check contract terms for cargo lien and booking hold authorization on persistent default",
			ExpectedOutcome:  "Legal enforcement entitlement memorandum",
			RiskLevel:        "LOW",
			RequiresApproval: false,
			Status:           "PENDING",
			Dependencies:     []string{"step-1-invoice-audit"},
		},
		{
			StepID:           "step-4-reconcile",
			WorkflowID:       workflowID,
			OrgID:            orgID,
			StepNumber:       4,
			AgentID:          "finance_agent",
			ActionType:       "POST_DISPUTE_RECONCILIATION",
			Title:            "Audit Discrepant Charges & Offsets",
			Description:      "Verify disputed accessorials and calculate net payable adjustment",
			ExpectedOutcome:  "Reconciled balance summary",
			RiskLevel:        "LOW",
			RequiresApproval: false,
			Status:           "PENDING",
			Dependencies:     []string{"step-3-contract-clauses"},
		},
	}
}

func (s *defaultService) buildCrossModuleRiskSteps(workflowID string, orgID int64, objective string) []EnterpriseWorkflowStep {
	return []EnterpriseWorkflowStep{
		{
			StepID:           "step-1-shipment-risk",
			WorkflowID:       workflowID,
			OrgID:            orgID,
			StepNumber:       1,
			AgentID:          "shipment_agent",
			ActionType:       "INSPECT_TRANSIT_RISK",
			Title:            "Assess Cargo Transit Vulnerability",
			Description:      "Analyze weather routing, geopolitical chokepoints, and transshipment dwell times",
			ExpectedOutcome:  "Operational transit risk report",
			RiskLevel:        "LOW",
			RequiresApproval: false,
			Status:           "PENDING",
			Dependencies:     []string{},
		},
		{
			StepID:           "step-2-compliance",
			WorkflowID:       workflowID,
			OrgID:            orgID,
			StepNumber:       2,
			AgentID:          "compliance_agent",
			ActionType:       "AUDIT_CUSTOMS_COMPLIANCE",
			Title:            "Audit Customs Filing & Sanctions",
			Description:      "Inspect export declarations, sanctioned entity lists, and DG permits",
			ExpectedOutcome:  "Regulatory compliance risk audit",
			RiskLevel:        "LOW",
			RequiresApproval: false,
			Status:           "PENDING",
			Dependencies:     []string{"step-1-shipment-risk"},
		},
		{
			StepID:           "step-3-contract",
			WorkflowID:       workflowID,
			OrgID:            orgID,
			StepNumber:       3,
			AgentID:          "contract_agent",
			ActionType:       "AUDIT_CONTRACT_TERMS",
			Title:            "Inspect Free Days & Liability Clauses",
			Description:      "Audit contract terms for demurrage limits and force majeure provisions",
			ExpectedOutcome:  "Contractual rights and liability boundaries",
			RiskLevel:        "LOW",
			RequiresApproval: false,
			Status:           "PENDING",
			Dependencies:     []string{"step-2-compliance"},
		},
		{
			StepID:           "step-4-finance-impact",
			WorkflowID:       workflowID,
			OrgID:            orgID,
			StepNumber:       4,
			AgentID:          "finance_agent",
			ActionType:       "ASSESS_EXPOSURE_IMPACT",
			Title:            "Quantify Financial Exposure & Insurance Cover",
			Description:      "Calculate total risk exposure against cargo insurance thresholds and reserves",
			ExpectedOutcome:  "Financial exposure calculation",
			RiskLevel:        "MEDIUM",
			RequiresApproval: true,
			Status:           "PENDING",
			Dependencies:     []string{"step-3-contract"},
		},
	}
}

func (s *defaultService) buildGenericSteps(workflowID string, orgID int64, objective string) []EnterpriseWorkflowStep {
	return []EnterpriseWorkflowStep{
		{
			StepID:           "step-1",
			WorkflowID:       workflowID,
			OrgID:            orgID,
			StepNumber:       1,
			AgentID:          "planning_agent",
			ActionType:       "TAG_INTERNAL_STATE",
			Title:            "Analyze & Formulate Plan",
			Description:      objective,
			ExpectedOutcome:  "Plan formulation",
			RiskLevel:        "LOW",
			RequiresApproval: false,
			Status:           "PENDING",
			Dependencies:     []string{},
		},
	}
}

func areDependenciesMet(deps []string, allSteps []EnterpriseWorkflowStep) bool {
	if len(deps) == 0 {
		return true
	}
	stepMap := make(map[string]string)
	for _, s := range allSteps {
		stepMap[s.StepID] = s.Status
	}
	for _, d := range deps {
		if stepMap[d] != "COMPLETED" && stepMap[d] != "SKIPPED" {
			return false
		}
	}
	return true
}

func isHighRiskAction(actionType string) bool {
	switch actionType {
	case "CUSTOMER_NOTIFICATION_DISPATCH", "CUSTOMER_CREDIT_CHECK", "COLLECTION_REMINDER_DISPATCH",
		"EXECUTE_INVOICE_HOLD", "CANCEL_BOOKING", "SYNTHESIZE_CROSS_MODULE_RISK":
		return true
	default:
		return false
	}
}

func isExecutableAction(actionType string) bool {
	switch actionType {
	case "TAG_INTERNAL_STATE", "LOG_OBSERVATION", "CARRIER_TELEMETRY_REFRESH":
		return true
	default:
		return false
	}
}

func isElevationAttempt(requested, allowed string) bool {
	order := map[string]int{
		"LEVEL_0_OBSERVE":              0,
		"LEVEL_1_RECOMMEND":            1,
		"LEVEL_2_PREPARE":              2,
		"LEVEL_3_CONTROLLED":           3,
		"LEVEL_3_CONTROLLED_EXECUTION": 3,
		"LEVEL_4_GOVERNED_MULTI_STEP":  4,
		"LEVEL_4_FULL_AUTONOMY":        4,
	}
	return order[requested] > order[allowed]
}

func (s *defaultService) logAuditEvent(ctx context.Context, orgID int64, eventType, entityID string, metadata map[string]interface{}) {
	if s.auditSvc != nil {
		s.auditSvc.RecordAsync(ctx, domain.CreateAuditLogParams{
			OrgID:        orgID,
			ActorType:    "SYSTEM",
			ActorName:    "EnterpriseAutonomousPlatform",
			Action:       eventType,
			Module:       "ENTERPRISE_AUTONOMY",
			ResourceType: "WORKFLOW",
			ResourceID:   entityID,
			Metadata:     metadata,
			Result:       "SUCCESS",
		})
	}
}

func (s *defaultService) logSecurityEvent(ctx context.Context, orgID int64, eventType string, metadata map[string]interface{}) {
	if s.auditSvc != nil {
		s.auditSvc.RecordAsync(ctx, domain.CreateAuditLogParams{
			OrgID:        orgID,
			ActorType:    "SYSTEM",
			ActorName:    "EnterpriseAutonomousPlatform",
			Action:       eventType,
			Module:       "SECURITY_POLICY",
			ResourceType: "ENTERPRISE_POLICY",
			ResourceID:   "SECURITY_VIOLATION",
			Metadata:     metadata,
			Result:       "SECURITY_VIOLATION",
		})
	}
}

var idCounter uint64

func generateID(prefix string) string {
	cnt := atomic.AddUint64(&idCounter, 1)
	ts := time.Now().UnixNano()
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%d-%d", ts, cnt)))
	hash := hex.EncodeToString(hasher.Sum(nil))[:12]
	return fmt.Sprintf("%s-%s", prefix, hash)
}

func (s *defaultService) isExecutableAction(actionName string) bool {
	if s.actionsSvc == nil {
		return false
	}
	reg := s.actionsSvc.GetRegistry()
	if reg == nil {
		return false
	}
	_, err := reg.GetAction(actionName)
	return err == nil
}
