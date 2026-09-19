package enterprise_autonomy

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/freel/backend/internal/actions"
	"github.com/freel/backend/internal/approvals"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockEnterpriseRepository implements in-memory enterprise_autonomy.Repository
type MockEnterpriseRepository struct {
	mu            sync.RWMutex
	workflows     map[string]*EnterpriseWorkflow
	workflowSteps map[string][]EnterpriseWorkflowStep
	policies      map[string]*EnterprisePolicyContext
	dedupKeys     map[string]time.Time
}

func NewMockEnterpriseRepository() *MockEnterpriseRepository {
	return &MockEnterpriseRepository{
		workflows:     make(map[string]*EnterpriseWorkflow),
		workflowSteps: make(map[string][]EnterpriseWorkflowStep),
		policies:      make(map[string]*EnterprisePolicyContext),
		dedupKeys:     make(map[string]time.Time),
	}
}

func (m *MockEnterpriseRepository) CreateWorkflow(ctx context.Context, wf *EnterpriseWorkflow) error {
	if wf.OrgID <= 0 {
		return ErrUnauthorizedTenant
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%d:%s", wf.OrgID, wf.WorkflowID)
	m.workflows[key] = wf
	if len(wf.Steps) > 0 {
		m.workflowSteps[key] = append([]EnterpriseWorkflowStep(nil), wf.Steps...)
	}
	return nil
}

func (m *MockEnterpriseRepository) GetWorkflow(ctx context.Context, orgID int64, workflowID string) (*EnterpriseWorkflow, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	wf, exists := m.workflows[fmt.Sprintf("%d:%s", orgID, workflowID)]
	if !exists {
		return nil, ErrWorkflowNotFound
	}
	// Copy to prevent accidental shared pointer mutation
	copied := *wf
	return &copied, nil
}

func (m *MockEnterpriseRepository) UpdateWorkflowState(ctx context.Context, orgID int64, workflowID string, state EnterpriseWorkflowState, errCode, errMsg *string) error {
	if orgID <= 0 {
		return ErrUnauthorizedTenant
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%d:%s", orgID, workflowID)
	wf, exists := m.workflows[key]
	if !exists {
		return ErrWorkflowNotFound
	}
	wf.CurrentState = state
	wf.ErrorCode = errCode
	wf.ErrorMessage = errMsg
	wf.UpdatedAt = time.Now().UTC()
	if state == StateCompleted || state == StateFailed || state == StateCancelled {
		now := time.Now().UTC()
		wf.CompletedAt = &now
	}
	return nil
}

func (m *MockEnterpriseRepository) UpdateWorkflowStop(ctx context.Context, orgID int64, workflowID string, stoppedByUserID *int64, stopReason string) error {
	if orgID <= 0 {
		return ErrUnauthorizedTenant
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%d:%s", orgID, workflowID)
	wf, exists := m.workflows[key]
	if !exists {
		return ErrWorkflowNotFound
	}
	now := time.Now().UTC()
	wf.CurrentState = StateCancelled
	wf.StoppedByUserID = stoppedByUserID
	wf.StoppedAt = &now
	wf.StopReason = &stopReason
	wf.UpdatedAt = now
	return nil
}

func (m *MockEnterpriseRepository) ListWorkflows(ctx context.Context, filter WorkflowFilter) ([]*EnterpriseWorkflow, int, error) {
	if filter.OrgID <= 0 {
		return nil, 0, ErrUnauthorizedTenant
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var res []*EnterpriseWorkflow
	for _, wf := range m.workflows {
		if wf.OrgID != filter.OrgID {
			continue
		}
		if filter.State != "" && string(wf.CurrentState) != filter.State {
			continue
		}
		if filter.WorkflowType != "" && string(wf.WorkflowType) != filter.WorkflowType {
			continue
		}
		res = append(res, wf)
	}
	return res, len(res), nil
}

func (m *MockEnterpriseRepository) SaveWorkflowSteps(ctx context.Context, orgID int64, workflowID string, steps []EnterpriseWorkflowStep) error {
	if orgID <= 0 {
		return ErrUnauthorizedTenant
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%d:%s", orgID, workflowID)
	existing := m.workflowSteps[key]
	stepMap := make(map[string]int)
	for i, s := range existing {
		stepMap[s.StepID] = i
	}
	for _, s := range steps {
		if idx, found := stepMap[s.StepID]; found {
			existing[idx] = s
		} else {
			existing = append(existing, s)
			stepMap[s.StepID] = len(existing) - 1
		}
	}
	m.workflowSteps[key] = existing
	return nil
}

func (m *MockEnterpriseRepository) GetWorkflowSteps(ctx context.Context, orgID int64, workflowID string) ([]EnterpriseWorkflowStep, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := fmt.Sprintf("%d:%s", orgID, workflowID)
	steps, exists := m.workflowSteps[key]
	if !exists {
		return []EnterpriseWorkflowStep{}, nil
	}
	return steps, nil
}

func (m *MockEnterpriseRepository) UpdateWorkflowStep(ctx context.Context, orgID int64, step *EnterpriseWorkflowStep) error {
	if orgID <= 0 {
		return ErrUnauthorizedTenant
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%d:%s", orgID, step.WorkflowID)
	steps := m.workflowSteps[key]
	for i := range steps {
		if steps[i].StepID == step.StepID {
			steps[i] = *step
			m.workflowSteps[key] = steps
			return nil
		}
	}
	return nil
}

func (m *MockEnterpriseRepository) GetActiveOrInterruptedWorkflows(ctx context.Context) ([]*EnterpriseWorkflow, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*EnterpriseWorkflow
	for _, wf := range m.workflows {
		if wf.CurrentState == StateRunning || wf.CurrentState == StatePending || wf.CurrentState == StateWaiting {
			list = append(list, wf)
		}
	}
	return list, nil
}

func (m *MockEnterpriseRepository) GetPolicyContext(ctx context.Context, orgID int64, module string) (*EnterprisePolicyContext, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := fmt.Sprintf("%d:%s", orgID, module)
	if pol, exists := m.policies[key]; exists {
		return pol, nil
	}
	// Return default governed policy
	return &EnterprisePolicyContext{
		OrgID:                  orgID,
		AutonomyLevel:          "LEVEL_3_CONTROLLED_EXECUTION",
		AllowedWorkflowTypes:   []string{"SHIPMENT_RECOVERY", "COMMERCIAL_CYCLE", "FINANCIAL_COLLECTION", "CROSS_MODULE_RISK"},
		AllowedActionTypes:     []string{"exceptions.execute_recovery", "customer_advisory", "carrier_inquiry", "customs_broker_notification", "audit_log"},
		RequiresApproval:       false,
		MaxMonetaryThreshold:   10000.0,
		MinConfidenceThreshold: 0.70,
		EmergencyStopActive:    false,
	}, nil
}

func (m *MockEnterpriseRepository) UpdatePolicyContext(ctx context.Context, orgID int64, policy *EnterprisePolicyContext) error {
	if orgID <= 0 {
		return ErrUnauthorizedTenant
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.policies[fmt.Sprintf("%d:%s", orgID, "GLOBAL")] = policy
	m.policies[fmt.Sprintf("%d:%s", orgID, "SHIPMENT_RECOVERY")] = policy
	m.policies[fmt.Sprintf("%d:%s", orgID, "COMMERCIAL_CYCLE")] = policy
	m.policies[fmt.Sprintf("%d:%s", orgID, "FINANCIAL_COLLECTION")] = policy
	m.policies[fmt.Sprintf("%d:%s", orgID, "CROSS_MODULE_RISK")] = policy
	return nil
}

func (m *MockEnterpriseRepository) CheckAndRecordEventDedup(ctx context.Context, orgID int64, dedupKey, eventType string) (bool, error) {
	if orgID <= 0 {
		return false, ErrUnauthorizedTenant
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%d:%s", orgID, dedupKey)
	if _, exists := m.dedupKeys[key]; exists {
		return false, nil // Already processed
	}
	m.dedupKeys[key] = time.Now().UTC()
	return true, nil
}

func (m *MockEnterpriseRepository) GetPlatformOverview(ctx context.Context, orgID int64) (*EnterprisePlatformOverview, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	overview := &EnterprisePlatformOverview{
		OrgID:               orgID,
		AutonomyLevel:       "LEVEL_3_CONTROLLED_EXECUTION",
		PlatformHealth:      "HEALTHY",
		EmergencyStopActive: false,
		RecentWorkflows:     make([]*EnterpriseWorkflow, 0),
	}
	for _, wf := range m.workflows {
		if wf.OrgID != orgID {
			continue
		}
		overview.RecentWorkflows = append(overview.RecentWorkflows, wf)
		switch wf.CurrentState {
		case StateRunning:
			overview.ActiveWorkflows++
		case StatePaused:
			overview.PausedWorkflows++
		case StateWaitingForApproval:
			overview.WaitingApprovals++
		case StateEscalated:
			overview.EscalatedWorkflows++
		case StateFailed:
			overview.FailedWorkflows++
		case StateCompleted:
			overview.CompletedWorkflows++
		}
	}
	return overview, nil
}

// MockActionsService implements actions.Service for tests
type MockActionsService struct {
	registry actions.Registry
	executed []actions.ActionExecutionRequest
}

func NewMockActionsService() *MockActionsService {
	return &MockActionsService{
		registry: actions.NewRegistry(),
		executed: make([]actions.ActionExecutionRequest, 0),
	}
}

func (m *MockActionsService) Execute(ctx context.Context, req actions.ActionExecutionRequest) (*actions.ActionExecutionResponse, error) {
	m.executed = append(m.executed, req)
	return &actions.ActionExecutionResponse{
		Success:       true,
		ActionName:    req.ActionName,
		CorrelationID: fmt.Sprintf("act-corr-%d", time.Now().UnixNano()),
		Data:          map[string]interface{}{"status": "EXECUTED", "input": req.Input},
	}, nil
}

func (m *MockActionsService) ListActions() []actions.ActionDescriptor {
	return []actions.ActionDescriptor{}
}

func (m *MockActionsService) GetRegistry() actions.Registry {
	return m.registry
}

func (m *MockActionsService) SetApprovalsService(svc approvals.Service) {}

// Test 1: Enterprise Workflow State Transitions
func TestValidateStateTransition(t *testing.T) {
	// Valid transitions
	assert.NoError(t, ValidateStateTransition(StatePending, StateRunning))
	assert.NoError(t, ValidateStateTransition(StatePending, StateCancelled))
	assert.NoError(t, ValidateStateTransition(StateRunning, StateWaitingForApproval))
	assert.NoError(t, ValidateStateTransition(StateRunning, StateCompleted))
	assert.NoError(t, ValidateStateTransition(StateRunning, StatePaused))
	assert.NoError(t, ValidateStateTransition(StateRunning, StateFailed))
	assert.NoError(t, ValidateStateTransition(StateWaitingForApproval, StateRunning))
	assert.NoError(t, ValidateStateTransition(StateWaitingForApproval, StateBlocked))
	assert.NoError(t, ValidateStateTransition(StatePaused, StateRunning))
	assert.NoError(t, ValidateStateTransition(StatePaused, StateCancelled))

	// Invalid transitions
	assert.ErrorIs(t, ValidateStateTransition(StateCompleted, StateRunning), ErrInvalidStateTransition)
	assert.ErrorIs(t, ValidateStateTransition(StateFailed, StateRunning), ErrInvalidStateTransition)
	assert.ErrorIs(t, ValidateStateTransition(StateCancelled, StateRunning), ErrInvalidStateTransition)
	assert.ErrorIs(t, ValidateStateTransition(StatePending, StateCompleted), ErrInvalidStateTransition)
}

// Test 2: Workflow Creation and Tenant Isolation
func TestEnterpriseTenantIsolation(t *testing.T) {
	repo := NewMockEnterpriseRepository()
	actionsSvc := NewMockActionsService()
	svc := NewService(repo, nil, actionsSvc, nil, nil)
	ctx := context.Background()

	// Tenant 1 creates a workflow
	wf1, err := svc.StartWorkflow(ctx, 1, StartWorkflowRequest{
		WorkflowType:      WorkflowShipmentRecovery,
		Objective:         "Resolve clearance hold for container",
		InitiatingEvent:   "SHIPMENT_CUSTOMS_HOLD",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "SH-1001",
	})
	require.NoError(t, err)
	require.NotNil(t, wf1)
	assert.Equal(t, int64(1), wf1.OrgID)
	assert.NotEmpty(t, wf1.WorkflowID)

	// Tenant 2 attempts to access Tenant 1's workflow
	_, err = svc.GetWorkflow(ctx, 2, wf1.WorkflowID)
	assert.ErrorIs(t, err, ErrWorkflowNotFound, "Cross-tenant access must be rejected")

	// Invalid tenant (orgID <= 0) must be rejected
	_, err = svc.GetWorkflow(ctx, 0, wf1.WorkflowID)
	assert.ErrorIs(t, err, ErrUnauthorizedTenant)

	// Tenant 2 attempts to pause Tenant 1's workflow
	err = svc.PauseWorkflow(ctx, 2, wf1.WorkflowID, PauseWorkflowRequest{Reason: "Malicious pause"})
	assert.ErrorIs(t, err, ErrWorkflowNotFound)

	// Tenant 1 can successfully view and manage it
	fetched, err := svc.GetWorkflow(ctx, 1, wf1.WorkflowID)
	require.NoError(t, err)
	assert.Equal(t, wf1.WorkflowID, fetched.WorkflowID)
}

// Test 3: Cross-Module Orchestration Steps
func TestCrossModuleOrchestration(t *testing.T) {
	repo := NewMockEnterpriseRepository()
	actionsSvc := NewMockActionsService()
	svc := NewService(repo, nil, actionsSvc, nil, nil)
	ctx := context.Background()

	// 1. Shipment Recovery Workflow: Shipment -> Exception -> Customer -> Finance -> Compliance
	wfShip, err := svc.StartWorkflow(ctx, 1, StartWorkflowRequest{
		WorkflowType:      WorkflowShipmentRecovery,
		Objective:         "Recover stalled multimodal transit",
		InitiatingEvent:   "DELAY_PREDICTED",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "SH-992",
	})
	require.NoError(t, err)
	require.Len(t, wfShip.Steps, 5)
	assert.Equal(t, "shipment_agent", wfShip.Steps[0].AgentID)
	assert.Equal(t, "exception_agent", wfShip.Steps[1].AgentID)
	assert.Equal(t, "customer_agent", wfShip.Steps[2].AgentID)
	assert.Equal(t, "finance_agent", wfShip.Steps[3].AgentID)
	assert.Equal(t, "compliance_agent", wfShip.Steps[4].AgentID)

	// Verify step dependencies
	assert.Empty(t, wfShip.Steps[0].Dependencies)
	assert.Contains(t, wfShip.Steps[1].Dependencies, wfShip.Steps[0].StepID)
	assert.Contains(t, wfShip.Steps[2].Dependencies, wfShip.Steps[1].StepID)

	// 2. Commercial Cycle Workflow: Lead -> RFQ -> Pricing -> Customer -> Quotation
	wfComm, err := svc.StartWorkflow(ctx, 1, StartWorkflowRequest{
		WorkflowType:      WorkflowCommercialCycle,
		Objective:         "Commercial quote pipeline",
		InitiatingEvent:   "RFQ_RECEIVED",
		RelatedEntityType: "RFQ",
		RelatedEntityID:   "RFQ-505",
	})
	require.NoError(t, err)
	require.Len(t, wfComm.Steps, 4)
	assert.Equal(t, "customer_agent", wfComm.Steps[0].AgentID)
	assert.Equal(t, "pricing_agent", wfComm.Steps[1].AgentID)
	assert.Equal(t, "finance_agent", wfComm.Steps[2].AgentID)
	assert.Equal(t, "contract_agent", wfComm.Steps[3].AgentID)

	// 3. Financial Collection Workflow: Invoice -> Customer -> Finance -> Contract
	wfFin, err := svc.StartWorkflow(ctx, 1, StartWorkflowRequest{
		WorkflowType:      WorkflowFinancialCollection,
		Objective:         "Reconcile overdue freight balance",
		InitiatingEvent:   "INVOICE_OVERDUE",
		RelatedEntityType: "INVOICE",
		RelatedEntityID:   "INV-888",
	})
	require.NoError(t, err)
	require.Len(t, wfFin.Steps, 4)
	assert.Equal(t, "finance_agent", wfFin.Steps[0].AgentID)
	assert.Equal(t, "customer_agent", wfFin.Steps[1].AgentID)

	// 4. Cross-Module Risk Workflow: Shipment -> Compliance -> Contract -> Finance
	wfRisk, err := svc.StartWorkflow(ctx, 1, StartWorkflowRequest{
		WorkflowType:      WorkflowCrossModuleRisk,
		Objective:         "Sanctions and tariff audit",
		InitiatingEvent:   "RISK_ALERT",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "SH-777",
	})
	require.NoError(t, err)
	require.Len(t, wfRisk.Steps, 4)
	assert.Equal(t, "shipment_agent", wfRisk.Steps[0].AgentID)
	assert.Equal(t, "compliance_agent", wfRisk.Steps[1].AgentID)
	assert.Equal(t, "contract_agent", wfRisk.Steps[2].AgentID)
	assert.Equal(t, "finance_agent", wfRisk.Steps[3].AgentID)
}

// Test 4: Autonomy Level Enforcement and Self-Elevation Rejection
func TestAutonomyEnforcementAndElevationDefense(t *testing.T) {
	repo := NewMockEnterpriseRepository()
	actionsSvc := NewMockActionsService()
	svc := NewService(repo, nil, actionsSvc, nil, nil)
	ctx := context.Background()

	// 1. Level 0 Observe: Cannot execute any modifying workflow
	err := repo.UpdatePolicyContext(ctx, 10, &EnterprisePolicyContext{
		OrgID:                10,
		AutonomyLevel:        "LEVEL_0_OBSERVE",
		AllowedWorkflowTypes: []string{"SHIPMENT_RECOVERY"},
	})
	require.NoError(t, err)

	_, err = svc.StartWorkflow(ctx, 10, StartWorkflowRequest{
		WorkflowType:    WorkflowShipmentRecovery,
		Objective:       "Test Level 0",
		InitiatingEvent: "TEST",
	})
	assert.ErrorIs(t, err, ErrAutonomyRestricted, "Level 0 Observe must refuse autonomous workflow execution")

	// 2. Self-elevation rejection: User/Agent requests LEVEL_4 when org allows only LEVEL_2
	err = repo.UpdatePolicyContext(ctx, 20, &EnterprisePolicyContext{
		OrgID:                20,
		AutonomyLevel:        "LEVEL_2_PREPARE",
		AllowedWorkflowTypes: []string{"SHIPMENT_RECOVERY"},
	})
	require.NoError(t, err)

	_, err = svc.StartWorkflow(ctx, 20, StartWorkflowRequest{
		WorkflowType:         WorkflowShipmentRecovery,
		Objective:            "Attempt privilege escalation",
		InitiatingEvent:      "TEST",
		AutonomyLevel:        "LEVEL_4_GOVERNED_MULTI_STEP",
	})
	assert.ErrorIs(t, err, ErrAutonomyRestricted, "Privilege escalation attempts must be rejected server-side")

	// 3. Prohibited workflow type rejection
	err = repo.UpdatePolicyContext(ctx, 30, &EnterprisePolicyContext{
		OrgID:                30,
		AutonomyLevel:        "LEVEL_3_CONTROLLED_EXECUTION",
		AllowedWorkflowTypes: []string{"COMMERCIAL_CYCLE"}, // Only commercial allowed
	})
	require.NoError(t, err)

	_, err = svc.StartWorkflow(ctx, 30, StartWorkflowRequest{
		WorkflowType:    WorkflowShipmentRecovery,
		Objective:       "Disallowed type",
		InitiatingEvent: "TEST",
	})
	assert.ErrorIs(t, err, ErrAutonomyRestricted, "Disallowed workflow type must be rejected")
}

// Test 5: Human-in-the-loop Approval and Rejection
func TestApprovalWorkflowGating(t *testing.T) {
	repo := NewMockEnterpriseRepository()
	actionsSvc := NewMockActionsService()
	svc := NewService(repo, nil, actionsSvc, nil, nil)
	ctx := context.Background()

	// Create workflow with a step requiring approval
	wf, err := svc.StartWorkflow(ctx, 1, StartWorkflowRequest{
		WorkflowType:      WorkflowShipmentRecovery,
		Objective:         "Test Approval Step",
		InitiatingEvent:   "EXCEPTION_LOGGED",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "SH-200",
	})
	require.NoError(t, err)

	steps, err := repo.GetWorkflowSteps(ctx, 1, wf.WorkflowID)
	require.NoError(t, err)
	require.NotEmpty(t, steps)

	// Step 2 (Exception recovery) requires approval because of HIGH risk
	var approvalStep *EnterpriseWorkflowStep
	for i := range steps {
		if steps[i].RequiresApproval {
			approvalStep = &steps[i]
			break
		}
	}
	require.NotNil(t, approvalStep, "Must have an approval required step")

	// Approve step
	err = svc.ApproveWorkflowStep(ctx, 1, wf.WorkflowID, approvalStep.StepID, 42, ApproveStepRequest{
		Notes: "Approved by Senior Supervisor",
	})
	require.NoError(t, err)

	updatedSteps, err := repo.GetWorkflowSteps(ctx, 1, wf.WorkflowID)
	require.NoError(t, err)
	for _, s := range updatedSteps {
		if s.StepID == approvalStep.StepID {
			assert.Equal(t, "COMPLETED", s.Status)
		}
	}

	// Test Rejection
	wf2, err := svc.StartWorkflow(ctx, 1, StartWorkflowRequest{
		WorkflowType:      WorkflowShipmentRecovery,
		Objective:         "Test Rejection Step",
		InitiatingEvent:   "EXCEPTION_LOGGED",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "SH-201",
	})
	require.NoError(t, err)

	steps2, _ := repo.GetWorkflowSteps(ctx, 1, wf2.WorkflowID)
	var stepToReject *EnterpriseWorkflowStep
	for i := range steps2 {
		if steps2[i].RequiresApproval {
			stepToReject = &steps2[i]
			break
		}
	}
	require.NotNil(t, stepToReject)

	err = svc.RejectWorkflowStep(ctx, 1, wf2.WorkflowID, stepToReject.StepID, 42, RejectStepRequest{
		Reason: "Carrier tariff too expensive",
	})
	require.NoError(t, err)

	wf2After, err := svc.GetWorkflow(ctx, 1, wf2.WorkflowID)
	require.NoError(t, err)
	assert.Equal(t, StateBlocked, wf2After.CurrentState)
}

// Test 6: Durable Workflow Recovery Across Restarts (Idempotent)
func TestInterruptedWorkflowRecovery(t *testing.T) {
	repo := NewMockEnterpriseRepository()
	actionsSvc := NewMockActionsService()
	svc := NewService(repo, nil, actionsSvc, nil, nil)
	ctx := context.Background()

	// Simulate an interrupted workflow: was in RUNNING state when Go restarted
	now := time.Now().UTC()
	interruptedWF := &EnterpriseWorkflow{
		ID:              999,
		WorkflowID:      "plan-interrupted-01",
		OrgID:           1,
		WorkflowType:    WorkflowShipmentRecovery,
		Objective:       "Interrupted recovery",
		InitiatingEvent: "RESTART_EVENT",
		CurrentState:    StateRunning,
		AutonomyLevel:   "LEVEL_3_CONTROLLED_EXECUTION",
		CorrelationID:   "corr-restart-999",
		CreatedAt:       now.Add(-10 * time.Minute),
		UpdatedAt:       now.Add(-5 * time.Minute),
	}
	err := repo.CreateWorkflow(ctx, interruptedWF)
	require.NoError(t, err)

	// Step 1 was completed before crash; Step 2 was pending
	steps := []EnterpriseWorkflowStep{
		{
			StepID:           "step-1",
			WorkflowID:       "plan-interrupted-01",
			OrgID:            1,
			StepNumber:       1,
			AgentID:          "shipment_agent",
			ActionType:       "shipment_telemetry_scan",
			Title:            "Scan shipment",
			Status:           "COMPLETED",
			RequiresApproval: false,
			IdempotencyKey:   "idem-step-1",
		},
		{
			StepID:           "step-2",
			WorkflowID:       "plan-interrupted-01",
			OrgID:            1,
			StepNumber:       2,
			AgentID:          "exception_agent",
			ActionType:       "carrier_inquiry",
			Title:            "Inquire carrier",
			Status:           "PENDING",
			RequiresApproval: false,
			IdempotencyKey:   "idem-step-2",
		},
	}
	err = repo.SaveWorkflowSteps(ctx, 1, "plan-interrupted-01", steps)
	require.NoError(t, err)

	// Execute recovery
	report, err := svc.RecoverInterruptedWorkflows(ctx)
	require.NoError(t, err)
	require.NotNil(t, report)
	assert.Equal(t, 1, report.ScannedCount)
	assert.Equal(t, 1, report.RecoveredCount)
	assert.Equal(t, 0, report.BlockedCount)

	// Verify workflow resumed and finished pending steps
	updatedWF, err := repo.GetWorkflow(ctx, 1, "plan-interrupted-01")
	require.NoError(t, err)
	assert.Equal(t, StateCompleted, updatedWF.CurrentState)

	// Verify Step 1 remained completed without duplicate execution
	updatedSteps, err := repo.GetWorkflowSteps(ctx, 1, "plan-interrupted-01")
	require.NoError(t, err)
	assert.Equal(t, "COMPLETED", updatedSteps[0].Status)
	assert.Equal(t, "COMPLETED", updatedSteps[1].Status)

	// Rerunning recovery again should be completely idempotent (0 interrupted)
	report2, err := svc.RecoverInterruptedWorkflows(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, report2.ScannedCount)
}

// Test 7: Event Deduplication
func TestEventDeduplication(t *testing.T) {
	repo := NewMockEnterpriseRepository()
	actionsSvc := NewMockActionsService()
	svc := NewService(repo, nil, actionsSvc, nil, nil)
	ctx := context.Background()

	req := TriggerBusinessEventRequest{
		EventID:        "evt-unique-12345",
		EventType:      "SHIPMENT_DELAY_DETECTED",
		EntityType:     "SHIPMENT",
		EntityID:       "SH-555",
		Payload:        map[string]interface{}{"delay_hours": 12},
		IdempotencyKey: "evt-unique-12345",
	}

	// First trigger succeeds
	wf1, err := svc.TriggerBusinessEvent(ctx, 1, req)
	require.NoError(t, err)
	require.NotNil(t, wf1)

	// Duplicate trigger with identical event ID is rejected as duplicate
	_, err = svc.TriggerBusinessEvent(ctx, 1, req)
	assert.ErrorIs(t, err, ErrDuplicateEventTrigger, "Duplicate business events must be prevented from triggering multiple workflows")
}

// Test 8: Emergency Control (Pause and Kill-Switch)
func TestEnterpriseEmergencyControl(t *testing.T) {
	repo := NewMockEnterpriseRepository()
	actionsSvc := NewMockActionsService()
	svc := NewService(repo, nil, actionsSvc, nil, nil)
	ctx := context.Background()

	// 1. Activate tenant emergency stop
	err := svc.EmergencyControl(ctx, 1, EmergencyControlRequest{
		Action: "PAUSE",
		Reason: "Critical network anomaly detected",
		Scope:  "TENANT",
	}, "SuperAdmin")
	require.NoError(t, err)

	// Verifying that new workflows are immediately blocked
	_, err = svc.StartWorkflow(ctx, 1, StartWorkflowRequest{
		WorkflowType:    WorkflowShipmentRecovery,
		Objective:       "Emergency blocked workflow",
		InitiatingEvent: "ALERT",
	})
	assert.ErrorIs(t, err, ErrEmergencyStopActive)

	// 2. Resume emergency stop
	err = svc.EmergencyControl(ctx, 1, EmergencyControlRequest{
		Action: "RESUME",
		Reason: "All systems verified normal",
		Scope:  "TENANT",
	}, "SuperAdmin")
	require.NoError(t, err)

	// Workflows now succeed
	wf, err := svc.StartWorkflow(ctx, 1, StartWorkflowRequest{
		WorkflowType:    WorkflowShipmentRecovery,
		Objective:       "Post-emergency workflow",
		InitiatingEvent: "NORMAL",
	})
	require.NoError(t, err)
	require.NotNil(t, wf)
}

// Test 9: Platform Overview Aggregation
func TestPlatformOverview(t *testing.T) {
	repo := NewMockEnterpriseRepository()
	actionsSvc := NewMockActionsService()
	svc := NewService(repo, nil, actionsSvc, nil, nil)
	ctx := context.Background()

	_, _ = svc.StartWorkflow(ctx, 1, StartWorkflowRequest{
		WorkflowType:    WorkflowShipmentRecovery,
		Objective:       "Shipment 1",
		InitiatingEvent: "EVT1",
	})
	_, _ = svc.StartWorkflow(ctx, 1, StartWorkflowRequest{
		WorkflowType:    WorkflowCommercialCycle,
		Objective:       "Commercial 1",
		InitiatingEvent: "EVT2",
	})

	overview, err := repo.GetPlatformOverview(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), overview.OrgID)
	assert.Equal(t, 2, len(overview.RecentWorkflows))
}

// Test 10: Python/Go Enterprise Contract Validation & Malformed Output Defense
func TestPythonContractValidation(t *testing.T) {
	repo := NewMockEnterpriseRepository()
	actionsSvc := NewMockActionsService()
	svc := NewService(repo, nil, actionsSvc, nil, nil)
	ctx := context.Background()

	// 1. Rejection of empty objective
	_, err := svc.StartWorkflow(ctx, 1, StartWorkflowRequest{
		WorkflowType: WorkflowShipmentRecovery,
		Objective:    "", // Empty
	})
	assert.Error(t, err, "Empty objective must be rejected")

	// 2. Rejection of unauthorized tenant access
	_, err = svc.StartWorkflow(ctx, -1, StartWorkflowRequest{
		WorkflowType: WorkflowShipmentRecovery,
		Objective:    "Valid objective",
	})
	assert.ErrorIs(t, err, ErrUnauthorizedTenant)

	// 3. Rejection of unknown workflow type
	_, err = svc.StartWorkflow(ctx, 1, StartWorkflowRequest{
		WorkflowType: "UNKNOWN_MALICIOUS_TYPE",
		Objective:    "Inject unauthorized workflow",
	})
	assert.ErrorIs(t, err, ErrAutonomyRestricted, "Unknown workflow type must be restricted")
}

// Test 11: Live MariaDB Repository Integration Test
func TestLiveMySQLRepository(t *testing.T) {
	dsn := "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC"
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		t.Skip("MySQL not accessible, skipping live repository integration test")
		return
	}
	if err := db.Ping(); err != nil {
		t.Skipf("MySQL ping failed: %v, skipping live repository test", err)
		return
	}
	defer db.Close()

	liveRepo := NewMySQLRepository(db)
	ctx := context.Background()
	orgID := int64(1)
	testWfID := fmt.Sprintf("test-live-wf-%d", time.Now().UnixNano())
	now := time.Now().UTC()

	// 1. Create Workflow Record in autonomous_plans
	wf := &EnterpriseWorkflow{
		WorkflowID:        testWfID,
		OrgID:             orgID,
		WorkflowType:      WorkflowShipmentRecovery,
		Objective:         "Live DB integration test workflow",
		InitiatingEvent:   "TEST_SUITE_INIT",
		CurrentState:      StateRunning,
		AutonomyLevel:     "LEVEL_3_CONTROLLED_EXECUTION",
		PolicyDecision:    "PERMITTED",
		CorrelationID:     fmt.Sprintf("corr-live-%d", time.Now().UnixNano()),
		IdempotencyKey:    fmt.Sprintf("idem-live-%d", time.Now().UnixNano()),
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "SH-TEST-001",
		Confidence:        0.95,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	err = liveRepo.CreateWorkflow(ctx, wf)
	require.NoError(t, err, "CreateWorkflow must succeed in live MySQL")

	// 2. Fetch Workflow
	fetched, err := liveRepo.GetWorkflow(ctx, orgID, testWfID)
	require.NoError(t, err)
	assert.Equal(t, testWfID, fetched.WorkflowID)
	assert.Equal(t, orgID, fetched.OrgID)
	assert.Equal(t, StateRunning, fetched.CurrentState)

	// Cross-tenant access must fail in live DB
	_, err = liveRepo.GetWorkflow(ctx, 99999, testWfID)
	assert.ErrorIs(t, err, ErrWorkflowNotFound)

	// 3. Save Workflow Steps in autonomous_plan_steps
	steps := []EnterpriseWorkflowStep{
		{
			StepID:           fmt.Sprintf("step-live-1-%d", time.Now().UnixNano()),
			WorkflowID:       testWfID,
			OrgID:            orgID,
			StepNumber:       1,
			AgentID:          "shipment_agent",
			ActionType:       "shipment_telemetry_scan",
			Title:            "Scan shipment telemetry",
			Description:      "Live DB step test",
			ExpectedOutcome:  "Sensor values",
			RiskLevel:        "LOW",
			RequiresApproval: false,
			Status:           "PENDING",
			IdempotencyKey:   fmt.Sprintf("idem-s1-%d", time.Now().UnixNano()),
			Dependencies:     []string{},
		},
	}
	err = liveRepo.SaveWorkflowSteps(ctx, orgID, testWfID, steps)
	require.NoError(t, err, "SaveWorkflowSteps must succeed in live MySQL")

	// 4. Fetch Workflow Steps
	dbSteps, err := liveRepo.GetWorkflowSteps(ctx, orgID, testWfID)
	require.NoError(t, err)
	require.NotEmpty(t, dbSteps)
	assert.Equal(t, steps[0].StepID, dbSteps[0].StepID)

	// 5. Update Step Status
	dbSteps[0].Status = "COMPLETED"
	err = liveRepo.UpdateWorkflowStep(ctx, orgID, &dbSteps[0])
	require.NoError(t, err)

	// 6. Update Workflow State
	err = liveRepo.UpdateWorkflowState(ctx, orgID, testWfID, StateCompleted, nil, nil)
	require.NoError(t, err)

	finalWF, err := liveRepo.GetWorkflow(ctx, orgID, testWfID)
	require.NoError(t, err)
	assert.Equal(t, StateCompleted, finalWF.CurrentState)

	// 7. Event Deduplication check
	dedupKey := fmt.Sprintf("evt-dedup-test-%d", time.Now().UnixNano())
	isNew, err := liveRepo.CheckAndRecordEventDedup(ctx, orgID, dedupKey, "TEST_EVENT")
	require.NoError(t, err)
	assert.True(t, isNew, "First dedup check must be true")

	isNew2, err := liveRepo.CheckAndRecordEventDedup(ctx, orgID, dedupKey, "TEST_EVENT")
	require.NoError(t, err)
	assert.False(t, isNew2, "Duplicate dedup check must be false")

	// 8. Platform Overview from live DB
	overview, err := liveRepo.GetPlatformOverview(ctx, orgID)
	require.NoError(t, err)
	assert.NotNil(t, overview)
	assert.Equal(t, orgID, overview.OrgID)
}
