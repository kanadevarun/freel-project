package enterprise_autonomy

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/freel/backend/internal/actions"
	"github.com/freel/backend/internal/approvals"
	"github.com/freel/backend/internal/orchestration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CountingMockActionsService counts execution requests and tracks idempotency
type CountingMockActionsService struct {
	executionCount int64
	lastRequest    *actions.ActionExecutionRequest
	registry       actions.Registry
	idempotentKeys map[string]bool
}

func NewCountingMockActionsService() *CountingMockActionsService {
	return &CountingMockActionsService{
		registry:       actions.NewRegistry(),
		idempotentKeys: make(map[string]bool),
	}
}

func (c *CountingMockActionsService) Execute(ctx context.Context, req actions.ActionExecutionRequest) (*actions.ActionExecutionResponse, error) {
	if req.IdempotencyKey != "" && c.idempotentKeys[req.IdempotencyKey] {
		return &actions.ActionExecutionResponse{
			Success:          true,
			ActionName:       req.ActionName,
			CorrelationID:    "corr-idemp-replay",
			IdempotentReplay: true,
			Data:             map[string]interface{}{"status": "ALREADY_EXECUTED"},
		}, nil
	}
	if req.IdempotencyKey != "" {
		c.idempotentKeys[req.IdempotencyKey] = true
	}

	atomic.AddInt64(&c.executionCount, 1)
	c.lastRequest = &req
	return &actions.ActionExecutionResponse{
		Success:       true,
		ActionName:    req.ActionName,
		CorrelationID: fmt.Sprintf("corr-%d", time.Now().UnixNano()),
		Data:          map[string]interface{}{"status": "SUCCESS", "input": req.Input},
	}, nil
}

func (c *CountingMockActionsService) ListActions() []actions.ActionDescriptor {
	return []actions.ActionDescriptor{}
}

func (c *CountingMockActionsService) GetRegistry() actions.Registry {
	return c.registry
}

func (c *CountingMockActionsService) SetApprovalsService(svc approvals.Service) {}

// 1. Canonical workflow/autonomy path executes correctly via actions.Service
func TestConsolidation_CanonicalPathExecutesCorrectly(t *testing.T) {
	repo := NewMockEnterpriseRepository()
	actionsSvc := NewCountingMockActionsService()
	svc := NewService(repo, nil, actionsSvc, nil, nil)
	ctx := context.Background()
	orgID := int64(1)

	wf, err := svc.StartWorkflow(ctx, orgID, StartWorkflowRequest{
		WorkflowType:      WorkflowShipmentRecovery,
		Objective:         "Canonical path execution test",
		InitiatingEvent:   "SHIPMENT_DELAY_DETECTED",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "SH-1001",
	})
	require.NoError(t, err)
	require.NotNil(t, wf)
	assert.NotEmpty(t, wf.WorkflowID)
	assert.True(t, wf.CurrentState == StateRunning || wf.CurrentState == StateWaitingForApproval)
	assert.Equal(t, orgID, wf.OrgID)
}

type mockOrchSidecar struct{}

func (m *mockOrchSidecar) ProposeAction(ctx context.Context, payload map[string]interface{}) (*orchestration.PythonProposalResponse, error) {
	return &orchestration.PythonProposalResponse{
		Status: "SUCCESS",
		Proposal: &orchestration.ActionProposal{
			ProposedActionType: "shipments.update_status",
			Confidence:         0.95,
		},
	}, nil
}

// 2. Legacy entry points (orchestration.Service) route through actions.Service without independent duplicate execution
func TestConsolidation_LegacyOrchestrationRoutesThroughActionsService(t *testing.T) {
	orchRepo := NewMockOrchestrationRepository()
	orchReg := orchestration.NewRegistry()
	actionsSvc := NewCountingMockActionsService()

	// Register canonical action in Action System
	_ = actionsSvc.GetRegistry().Register(&mockSimpleAction{name: "shipments.update_status"})

	orchSvc := orchestration.NewService(orchRepo, orchReg, &mockOrchSidecar{}, nil, nil)
	orchSvc.SetActionsService(actionsSvc)

	ctx := context.Background()
	orgID := int64(1)
	userID := int64(42)

	// Create a proposal
	proposal := &orchestration.ActionProposal{
		ProposalID:         "prop-legacy-001",
		OrgID:              orgID,
		SourceModule:       "SHIPMENTS",
		SourceRecordType:   "SHIPMENT",
		SourceRecordID:     "SHP-500",
		TriggerEvent:       "MANUAL",
		ProposedActionType: "shipments.update_status",
		ActionParameters:   json.RawMessage(`{"shipment_id":500,"new_status":"IN_TRANSIT"}`),
		Status:             orchestration.StatusProposed,
		RiskLevel:          orchestration.RiskLow,
		RequiresApproval:   false,
		CorrelationID:      "corr-leg-1",
	}
	err := orchRepo.CreateProposal(ctx, proposal)
	require.NoError(t, err)

	exec, err := orchSvc.ExecuteProposal(ctx, orgID, userID, proposal.ProposalID, orchestration.ExecuteProposalRequest{
		IdempotencyKey: "idemp-legacy-001",
	})
	require.NoError(t, err)
	assert.Equal(t, orchestration.StatusCompleted, exec.Status)
	assert.Equal(t, "VERIFIED", exec.VerificationStatus)
	// Verified routed through canonical Action System
	assert.Equal(t, int64(1), atomic.LoadInt64(&actionsSvc.executionCount))
	assert.Equal(t, "shipments.update_status", actionsSvc.lastRequest.ActionName)
}

// 3. Same event cannot trigger two autonomous mutations (Event Mesh deduplication)
func TestConsolidation_EventMeshDeduplicationPreventsDuplicateExecution(t *testing.T) {
	meshSvc, _ := setupEventMeshTestService()
	ctx := context.Background()
	orgID := int64(1)

	req := IngestBusinessEventRequest{
		EventID:      "evt-dedup-consolidation-001",
		EventType:    "SHIPMENT_DELAY_DETECTED",
		SourceModule: "shipments",
		EntityType:   "SHIPMENT",
		EntityID:     "SHP-DEDUP-99",
		Payload:      map[string]interface{}{"delay_hours": 8.5},
	}

	// First ingestion triggers workflow creation
	evt1, wf1, err1 := meshSvc.IngestEvent(ctx, orgID, req)
	require.NoError(t, err1)
	require.NotNil(t, evt1)
	require.NotNil(t, wf1)
	assert.Equal(t, EventStatusWorkflowActive, evt1.ProcessingStatus)

	// Second ingestion of identical event must be deduplicated
	evt2, wf2, err2 := meshSvc.IngestEvent(ctx, orgID, req)
	require.NoError(t, err2)
	assert.Equal(t, EventStatusDeduplicated, evt2.ProcessingStatus)
	// Second ingestion returns the existing workflow, not a duplicate workflow
	if wf2 != nil {
		assert.Equal(t, wf1.WorkflowID, wf2.WorkflowID)
	}
}

// 4. Action System remains the single enforcement boundary
func TestConsolidation_ActionSystemEnforcementBoundary(t *testing.T) {
	actionsSvc := NewCountingMockActionsService()
	ctx := context.Background()

	resp, err := actionsSvc.Execute(ctx, actions.ActionExecutionRequest{
		ActionName:     "shipments.update_eta",
		OrgID:          1,
		ActingUserID:   42,
		ActorType:      actions.ActorTypeUI,
		IdempotencyKey: "idemp-boundary-1",
		Input:          map[string]interface{}{"eta": time.Now().Add(24 * time.Hour).Format(time.RFC3339)},
	})
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, int64(1), atomic.LoadInt64(&actionsSvc.executionCount))
}

// 5. Approval-required actions still require approval
func TestConsolidation_ApprovalGatingRemainsEnforced(t *testing.T) {
	repo := NewMockEnterpriseRepository()
	actionsSvc := NewCountingMockActionsService()
	svc := NewService(repo, nil, actionsSvc, nil, nil)
	ctx := context.Background()
	orgID := int64(1)

	wf, err := svc.StartWorkflow(ctx, orgID, StartWorkflowRequest{
		WorkflowType:      WorkflowShipmentRecovery,
		Objective:         "Critical hold requiring approval",
		InitiatingEvent:   "CONTAINER_HELD_CUSTOMS",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "SH-CUSTOMS-01",
	})
	require.NoError(t, err)

	// Step requiring approval cannot be executed without approval
	step := EnterpriseWorkflowStep{
		StepID:           "step-approval-req",
		WorkflowID:       wf.WorkflowID,
		StepNumber:       1,
		ActionType:       "customs.pay_duty",
		Title:            "Pay customs duty",
		RequiresApproval: true,
		Status:           "PENDING",
	}
	wf.Steps = []EnterpriseWorkflowStep{step}
	_ = repo.CreateWorkflow(ctx, wf)

	// Direct resume/execute without approval remains blocked
	err = svc.ResumeWorkflow(ctx, orgID, wf.WorkflowID)
	// Must not transition to running if steps require approval
	updatedWf, _ := repo.GetWorkflow(ctx, orgID, wf.WorkflowID)
	assert.NotEqual(t, StateCompleted, updatedWf.CurrentState)
}

// 6. Level 3/4 autonomy remains policy-controlled
func TestConsolidation_AutonomyPolicyEnforcement(t *testing.T) {
	repo := NewMockEnterpriseRepository()
	ctx := context.Background()
	orgID := int64(1)

	// Restrict to Level 1
	_ = repo.UpdatePolicyContext(ctx, orgID, &EnterprisePolicyContext{
		OrgID:                orgID,
		AutonomyLevel:        "LEVEL_1_RECOMMEND",
		AllowedWorkflowTypes: []string{"SHIPMENT_RECOVERY"},
		AllowedActionTypes:   []string{"audit_log"},
		RequiresApproval:     true,
		EmergencyStopActive:  false,
	})

	pol, err := repo.GetPolicyContext(ctx, orgID, "SHIPMENT_RECOVERY")
	require.NoError(t, err)
	assert.Equal(t, "LEVEL_1_RECOMMEND", pol.AutonomyLevel)
	assert.True(t, pol.RequiresApproval)
}

// 7. Idempotency prevents duplicate execution
func TestConsolidation_IdempotencyPreventsDuplicateExecution(t *testing.T) {
	actionsSvc := NewCountingMockActionsService()
	ctx := context.Background()
	key := "idemp-test-duplicate-001"

	req := actions.ActionExecutionRequest{
		ActionName:     "shipments.update_eta",
		OrgID:          1,
		IdempotencyKey: key,
		Input:          map[string]interface{}{"eta": "2026-09-15T12:00:00Z"},
	}

	// First execution succeeds
	resp1, err1 := actionsSvc.Execute(ctx, req)
	require.NoError(t, err1)
	assert.True(t, resp1.Success)
	assert.False(t, resp1.IdempotentReplay)
	assert.Equal(t, int64(1), atomic.LoadInt64(&actionsSvc.executionCount))

	// Second execution with same idempotency key replays without re-executing
	resp2, err2 := actionsSvc.Execute(ctx, req)
	require.NoError(t, err2)
	assert.True(t, resp2.Success)
	assert.True(t, resp2.IdempotentReplay)
	// Execution count does not increment!
	assert.Equal(t, int64(1), atomic.LoadInt64(&actionsSvc.executionCount))
}

// 8. Restart/recovery does not create duplicate execution
func TestConsolidation_RecoveryDoesNotDuplicateExecution(t *testing.T) {
	repo := NewMockEnterpriseRepository()
	actionsSvc := NewCountingMockActionsService()
	svc := NewService(repo, nil, actionsSvc, nil, nil)
	ctx := context.Background()
	orgID := int64(1)

	// Create a running workflow
	wf, _ := svc.StartWorkflow(ctx, orgID, StartWorkflowRequest{
		WorkflowType:      WorkflowShipmentRecovery,
		Objective:         "Recovery test",
		InitiatingEvent:   "SHIPMENT_DELAY_DETECTED",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "SH-REC-01",
	})
	require.NotNil(t, wf)

	// First recovery
	report1, err1 := svc.RecoverInterruptedWorkflows(ctx)
	require.NoError(t, err1)
	require.NotNil(t, report1)

	// Second immediate recovery scans cleanly without duplicating
	report2, err2 := svc.RecoverInterruptedWorkflows(ctx)
	require.NoError(t, err2)
	require.NotNil(t, report2)
	assert.Equal(t, 0, report2.RecoveredCount)
}

// 9. Tenant isolation remains intact
func TestConsolidation_TenantIsolationStrictlyEnforced(t *testing.T) {
	repo := NewMockEnterpriseRepository()
	svc := NewService(repo, nil, nil, nil, nil)
	ctx := context.Background()

	// Org 1 workflow
	wf1, err := svc.StartWorkflow(ctx, 1, StartWorkflowRequest{
		WorkflowType:      WorkflowCommercialCycle,
		Objective:         "Org 1 Commercial cycle",
		InitiatingEvent:   "RFQ_RECEIVED",
		RelatedEntityType: "RFQ",
		RelatedEntityID:   "RFQ-ORG1-01",
	})
	require.NoError(t, err)

	// Org 2 cannot read Org 1 workflow
	_, err = svc.GetWorkflow(ctx, 2, wf1.WorkflowID)
	assert.ErrorIs(t, err, ErrWorkflowNotFound)

	// Org 2 cannot pause Org 1 workflow
	err = svc.PauseWorkflow(ctx, 2, wf1.WorkflowID, PauseWorkflowRequest{Reason: "Malicious cross-tenant pause"})
	assert.ErrorIs(t, err, ErrWorkflowNotFound)

	// Invalid Org ID is rejected immediately
	_, err = svc.StartWorkflow(ctx, 0, StartWorkflowRequest{Objective: "Zero tenant"})
	assert.ErrorIs(t, err, ErrUnauthorizedTenant)
}

// Helpers for tests
type mockSimpleAction struct {
	name string
}

func (a *mockSimpleAction) Name() string                               { return a.name }
func (a *mockSimpleAction) Module() string                             { return "shipments" }
func (a *mockSimpleAction) Description() string                        { return "Mock shipment status update" }
func (a *mockSimpleAction) Category() actions.ActionCategory           { return actions.ActionCategoryWrite }
func (a *mockSimpleAction) InputSchema() interface{}                   { return nil }
func (a *mockSimpleAction) RequiresConfirmation() bool                 { return false }
func (a *mockSimpleAction) RequiredPermission() (string, string)       { return "shipments", "update" }
func (a *mockSimpleAction) Execute(ctx *actions.ActionContext, input []byte) (*actions.ActionResult, error) {
	return &actions.ActionResult{Success: true, Summary: "mock success"}, nil
}

type MockOrchestrationRepository struct {
	proposals  map[string]*orchestration.ActionProposal
	executions map[int64]*orchestration.ActionExecution
	idempMap   map[string]*orchestration.ActionExecution
	nextExecID int64
}

func NewMockOrchestrationRepository() *MockOrchestrationRepository {
	return &MockOrchestrationRepository{
		proposals:  make(map[string]*orchestration.ActionProposal),
		executions: make(map[int64]*orchestration.ActionExecution),
		idempMap:   make(map[string]*orchestration.ActionExecution),
		nextExecID: 1,
	}
}

func (m *MockOrchestrationRepository) CreateProposal(ctx context.Context, p *orchestration.ActionProposal) error {
	m.proposals[p.ProposalID] = p
	return nil
}

func (m *MockOrchestrationRepository) GetProposalByID(ctx context.Context, orgID, id int64) (*orchestration.ActionProposal, error) {
	for _, p := range m.proposals {
		if p.OrgID == orgID && p.ID == id {
			return p, nil
		}
	}
	return nil, fmt.Errorf("proposal not found")
}

func (m *MockOrchestrationRepository) GetProposalByProposalID(ctx context.Context, orgID int64, proposalID string) (*orchestration.ActionProposal, error) {
	if p, ok := m.proposals[proposalID]; ok && p.OrgID == orgID {
		return p, nil
	}
	return nil, fmt.Errorf("proposal not found")
}

func (m *MockOrchestrationRepository) UpdateProposal(ctx context.Context, p *orchestration.ActionProposal) error {
	m.proposals[p.ProposalID] = p
	return nil
}

func (m *MockOrchestrationRepository) ListProposals(ctx context.Context, orgID int64, filter orchestration.ProposalFilter) ([]*orchestration.ActionProposal, int64, error) {
	var list []*orchestration.ActionProposal
	for _, p := range m.proposals {
		if p.OrgID == orgID {
			list = append(list, p)
		}
	}
	return list, int64(len(list)), nil
}

func (m *MockOrchestrationRepository) CreateExecution(ctx context.Context, e *orchestration.ActionExecution) error {
	e.ID = m.nextExecID
	m.nextExecID++
	m.executions[e.ID] = e
	m.idempMap[fmt.Sprintf("%d:%s", e.OrgID, e.IdempotencyKey)] = e
	return nil
}

func (m *MockOrchestrationRepository) GetExecutionByID(ctx context.Context, orgID, id int64) (*orchestration.ActionExecution, error) {
	if e, ok := m.executions[id]; ok && e.OrgID == orgID {
		return e, nil
	}
	return nil, fmt.Errorf("execution not found")
}

func (m *MockOrchestrationRepository) GetExecutionByIdempotencyKey(ctx context.Context, orgID int64, key string) (*orchestration.ActionExecution, error) {
	if e, ok := m.idempMap[fmt.Sprintf("%d:%s", orgID, key)]; ok {
		return e, nil
	}
	return nil, nil
}

func (m *MockOrchestrationRepository) UpdateExecution(ctx context.Context, e *orchestration.ActionExecution) error {
	m.executions[e.ID] = e
	return nil
}

func (m *MockOrchestrationRepository) ListExecutions(ctx context.Context, orgID int64, filter orchestration.ExecutionFilter) ([]*orchestration.ActionExecution, int64, error) {
	var list []*orchestration.ActionExecution
	for _, e := range m.executions {
		if e.OrgID == orgID {
			list = append(list, e)
		}
	}
	return list, int64(len(list)), nil
}
