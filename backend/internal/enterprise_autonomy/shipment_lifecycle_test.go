package enterprise_autonomy

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/freel/backend/internal/actions"
	"github.com/freel/backend/internal/approvals"
	"github.com/freel/backend/internal/predictions"
	"github.com/freel/backend/internal/shipments"
	"github.com/freel/backend/internal/shipments/spec"
	"github.com/freel/backend/internal/workforce"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockShipmentsService implements minimal shipments.Service for lifecycle tests
type MockShipmentsService struct {
	shipments.Service
	Shipment   *spec.Shipment
	Milestones []*spec.ShipmentMilestone
	Exceptions []*spec.ShipmentException
}

func (m *MockShipmentsService) GetShipmentByID(ctx context.Context, orgID int64, id int64) (*spec.Shipment, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if m.Shipment != nil && m.Shipment.ID == id && m.Shipment.OrgID == orgID {
		return m.Shipment, nil
	}
	return nil, fmt.Errorf("shipment %d not found", id)
}

func (m *MockShipmentsService) GetMilestones(ctx context.Context, shipmentID int64) ([]*spec.ShipmentMilestone, error) {
	return m.Milestones, nil
}

func (m *MockShipmentsService) GetShipmentExceptions(ctx context.Context, orgID int64, shipmentID int64) ([]*spec.ShipmentException, error) {
	return m.Exceptions, nil
}

func (m *MockShipmentsService) UpdateShipment(ctx context.Context, s *spec.Shipment) error {
	m.Shipment = s
	return nil
}

func (m *MockShipmentsService) UpdateMilestone(ctx context.Context, orgID int64, shipmentID int64, milestoneCode string, actualDate *time.Time, location *string, notes *string) error {
	for _, ms := range m.Milestones {
		if ms.MilestoneCode == milestoneCode {
			ms.ActualDate = actualDate
			ms.Location = location
			return nil
		}
	}
	m.Milestones = append(m.Milestones, &spec.ShipmentMilestone{
		ShipmentID:    shipmentID,
		MilestoneCode: milestoneCode,
		ActualDate:    actualDate,
		Location:      location,
	})
	return nil
}

// MockPredictionsService implements minimal predictions.Service for lifecycle tests
type MockPredictionsService struct {
	predictions.Service
	PredictedETA *string
	Confidence   float64
}

func (m *MockPredictionsService) GetOrPredictShipmentETA(ctx context.Context, orgID int64, userID *int64, shipmentID int64, forceRefresh bool) (*predictions.Prediction, error) {
	return &predictions.Prediction{
		ID:                1001,
		PredictionID:      fmt.Sprintf("pred-eta-%d", shipmentID),
		OrgID:             orgID,
		Module:            "SHIPMENTS",
		PredictionType:    "SHIPMENT_ETA",
		RelatedRecordType: "SHIPMENT",
		RelatedRecordID:   fmt.Sprintf("%d", shipmentID),
		PredictedValue:    m.PredictedETA,
		ConfidenceScore:   m.Confidence,
		CreatedAt:         time.Now().UTC(),
	}, nil
}

func (m *MockPredictionsService) GetOrPredictCustomerIntelligence(ctx context.Context, orgID int64, userID *int64, customerID int64, forceRefresh bool) (*predictions.Prediction, error) {
	val := "TIER_1_ENTERPRISE"
	return &predictions.Prediction{
		ID:                1002,
		PredictionID:      fmt.Sprintf("pred-cust-%d", customerID),
		OrgID:             orgID,
		Module:            "CUSTOMERS",
		PredictionType:    "CUSTOMER_INTELLIGENCE",
		RelatedRecordType: "CUSTOMER",
		RelatedRecordID:   fmt.Sprintf("%d", customerID),
		PredictedValue:    &val,
		ConfidenceScore:   m.Confidence,
		CreatedAt:         time.Now().UTC(),
	}, nil
}

func (m *MockPredictionsService) GetOrPredictRFQMarginIntelligence(ctx context.Context, orgID int64, userID *int64, rfqID int64, forceRefresh bool) (*predictions.Prediction, error) {
	val := "20.0"
	return &predictions.Prediction{
		ID:                1003,
		PredictionID:      fmt.Sprintf("pred-rfq-%d", rfqID),
		OrgID:             orgID,
		Module:            "RFQ",
		PredictionType:    "MARGIN_OPTIMIZATION",
		RelatedRecordType: "RFQ",
		RelatedRecordID:   fmt.Sprintf("%d", rfqID),
		PredictedValue:    &val,
		ConfidenceScore:   m.Confidence,
		CreatedAt:         time.Now().UTC(),
	}, nil
}

// TestMockAction implements actions.Action for lifecycle tests
type TestMockAction struct {
	name    string
	desc    string
	reqConf bool
}

func (a *TestMockAction) Name() string                         { return a.name }
func (a *TestMockAction) Module() string                       { return "shipments" }
func (a *TestMockAction) Description() string                  { return a.desc }
func (a *TestMockAction) Category() actions.ActionCategory     { return actions.ActionCategoryWrite }
func (a *TestMockAction) InputSchema() interface{}             { return nil }
func (a *TestMockAction) RequiresConfirmation() bool           { return a.reqConf }
func (a *TestMockAction) RequiredPermission() (string, string) { return "shipments", "update" }
func (a *TestMockAction) Execute(ctx *actions.ActionContext, input []byte) (*actions.ActionResult, error) {
	return &actions.ActionResult{
		Success: true,
		Data:    map[string]interface{}{"status": "EXECUTED"},
	}, nil
}

type MockApprovalsService struct {
	approvals.Service
	createdRequests []*approvals.ApprovalRequest
}

func (m *MockApprovalsService) CreateApproval(ctx context.Context, orgID int64, input *approvals.CreateApprovalInput, actorName string) (*approvals.ApprovalRequest, error) {
	req := &approvals.ApprovalRequest{
		ID:        int64(len(m.createdRequests) + 1),
		OrgID:     orgID,
		Status:    approvals.StatusPendingApproval,
		CreatedAt: time.Now().UTC(),
	}
	m.createdRequests = append(m.createdRequests, req)
	return req, nil
}

func (m *MockApprovalsService) ListApprovals(ctx context.Context, orgID int64) ([]*approvals.ApprovalRequest, error) {
	return m.createdRequests, nil
}

type MockWorkforceService struct {
	workforce.Service
	recordedOutcomes []workforce.RecordOutcomeRequest
}

func (m *MockWorkforceService) RecordOutcome(ctx context.Context, orgID int64, req workforce.RecordOutcomeRequest) (*workforce.AgentOutcome, error) {
	m.recordedOutcomes = append(m.recordedOutcomes, req)
	return &workforce.AgentOutcome{
		OutcomeID: fmt.Sprintf("outc-%d", len(m.recordedOutcomes)),
		TenantID:  orgID,
		Status:    "RECORDED",
	}, nil
}

type MockTestIdempotencyStore struct {
	store map[string]*actions.ActionResult
}

func (m *MockTestIdempotencyStore) Start(ctx context.Context, orgID int64, key string) (bool, *actions.ActionResult, error) {
	if res, ok := m.store[key]; ok {
		return false, res, nil
	}
	return true, nil, nil
}

func (m *MockTestIdempotencyStore) Complete(ctx context.Context, orgID int64, key string, result *actions.ActionResult, ttl time.Duration) error {
	m.store[key] = result
	return nil
}

func (m *MockTestIdempotencyStore) CheckConflict(ctx context.Context, orgID int64, key string, actionName string) (bool, error) {
	return false, nil
}

// Helper to assemble test services
func setupShipmentLifecycleTest(t *testing.T) (
	ShipmentLifecycleService,
	*MockEnterpriseRepository,
	*MockShipmentsService,
	*MockPredictionsService,
	*MockApprovalsService,
	*MockWorkforceService,
) {
	mockRepo := NewMockEnterpriseRepository()
	mockWorkforce := &MockWorkforceService{recordedOutcomes: make([]workforce.RecordOutcomeRequest, 0)}
	mockApprovals := &MockApprovalsService{createdRequests: make([]*approvals.ApprovalRequest, 0)}

	reg := actions.NewRegistry()
	_ = reg.Register(&TestMockAction{name: "shipments.update_milestone", desc: "Update shipment milestone", reqConf: false})
	_ = reg.Register(&TestMockAction{name: "customer_advisory", desc: "Send customer advisory", reqConf: true})
	_ = reg.Register(&TestMockAction{name: "carrier_inquiry", desc: "Initiate carrier inquiry", reqConf: false})

	store := &MockTestIdempotencyStore{store: make(map[string]*actions.ActionResult)}
	actSvc := actions.NewService(reg, store, nil, nil)
	actSvc.SetApprovalsService(mockApprovals)

	eta := time.Now().UTC().Add(48 * time.Hour)
	mockShip := &MockShipmentsService{
		Shipment: &spec.Shipment{
			ID:              101,
			OrgID:           1,
			CarrierSCAC:     "MAEU",
			Status:          "BOOKED",
			OriginPort:      "USLAX",
			DestinationPort: "SGSIN",
			ETD:             &eta,
			ETA:             &eta,
		},
		Milestones: []*spec.ShipmentMilestone{
			{
				ShipmentID:    101,
				MilestoneCode: "BOOKED",
			},
		},
		Exceptions: []*spec.ShipmentException{},
	}

	predETAStr := eta.Add(6 * time.Hour).Format(time.RFC3339)
	mockPred := &MockPredictionsService{
		PredictedETA: &predETAStr,
		Confidence:   0.92,
	}

	svc := NewShipmentLifecycleService(
		mockRepo,
		mockWorkforce,
		actSvc,
		mockApprovals,
		nil,
		mockShip,
		mockPred,
	)

	return svc, mockRepo, mockShip, mockPred, mockApprovals, mockWorkforce
}

// ---------------------------------------------------------------------
// 1. Stage Transition Matrix Tests
// ---------------------------------------------------------------------

func TestValidateShipmentStageTransition(t *testing.T) {
	// Valid sequential transitions
	assert.NoError(t, ValidateShipmentStageTransition(StageShipmentCreated, StagePlanning))
	assert.NoError(t, ValidateShipmentStageTransition(StagePlanning, StageBooked))
	assert.NoError(t, ValidateShipmentStageTransition(StageBooked, StageInTransit))
	assert.NoError(t, ValidateShipmentStageTransition(StageInTransit, StageMonitoring))
	assert.NoError(t, ValidateShipmentStageTransition(StageMonitoring, StageDelivered))
	assert.NoError(t, ValidateShipmentStageTransition(StageDelivered, StagePostDelivery))
	assert.NoError(t, ValidateShipmentStageTransition(StagePostDelivery, StageCompleted))

	// Monitoring can transition to exception investigation / replanning, then resume monitoring
	assert.NoError(t, ValidateShipmentStageTransition(StageMonitoring, StagePlanning))
	assert.NoError(t, ValidateShipmentStageTransition(StageInTransit, StageDelivered))

	// Invalid backward or invalid jumps
	assert.Error(t, ValidateShipmentStageTransition(StageShipmentCreated, StageCompleted))
	assert.Error(t, ValidateShipmentStageTransition(StagePlanning, StageDelivered))
	assert.Error(t, ValidateShipmentStageTransition(StageCompleted, StagePlanning))
	assert.Error(t, ValidateShipmentStageTransition(StagePostDelivery, StageBooked))
}

// ---------------------------------------------------------------------
// 2. Lifecycle Initiation & Planning Tests
// ---------------------------------------------------------------------

func TestShipmentLifecycleInitiateAndPlanning(t *testing.T) {
	ctx := context.Background()
	svc, mockRepo, mockShip, _, _, _ := setupShipmentLifecycleTest(t)

	// Initiate lifecycle
	wf, err := svc.InitiateShipmentLifecycle(ctx, 1, 101, "corr-init-1")
	require.NoError(t, err)
	require.NotNil(t, wf)

	assert.Equal(t, int64(1), wf.OrgID)
	assert.Equal(t, int64(101), wf.ShipmentID)
	assert.Equal(t, StagePlanning, wf.CurrentStage)
	assert.Equal(t, "BOOKED", wf.AuthoritativeStatus)
	assert.NotEmpty(t, wf.AssignedSpecialists)

	// Multi-agent workforce assigned for planning
	assert.Contains(t, wf.AssignedSpecialists, "planning_agent")
	assert.Contains(t, wf.AssignedSpecialists, "shipment_agent")
	assert.Contains(t, wf.AssignedSpecialists, "pricing_agent")
	assert.Contains(t, wf.AssignedSpecialists, "compliance_agent")

	// Verify plan was persisted in repository
	savedWF, err := mockRepo.GetWorkflow(ctx, 1, wf.WorkflowID)
	require.NoError(t, err)
	assert.Equal(t, StateRunning, savedWF.CurrentState)

	// Cross-tenant access is blocked
	_, err = svc.InitiateShipmentLifecycle(ctx, 2, 101, "cross-tenant")
	assert.Error(t, err)

	// Non-existent shipment returns error
	mockShip.Shipment = nil
	_, err = svc.InitiateShipmentLifecycle(ctx, 1, 999, "not-found")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------
// 3. Inbound Event Processing & Loop Prevention Tests
// ---------------------------------------------------------------------

func TestShipmentEventProcessingAndLoopPrevention(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _, _, _ := setupShipmentLifecycleTest(t)

	// Initiate workflow
	_, err := svc.InitiateShipmentLifecycle(ctx, 1, 101, "corr-event-test")
	require.NoError(t, err)

	// 1. Process normal CARRIER_ASSIGNED event
	evt1 := ShipmentLifecycleEvent{
		EventID:       "evt-carrier-1",
		ShipmentID:    101,
		EventType:     "CARRIER_ASSIGNED",
		MilestoneCode: "BOOKED",
		Source:        "carrier_edi",
		CorrelationID: "corr-edi-1",
	}
	updatedWF, err := svc.ProcessShipmentEvent(ctx, 1, evt1)
	require.NoError(t, err)
	assert.Equal(t, StageBooked, updatedWF.CurrentStage)

	// 2. Process DEPARTURE event -> stage advances to IN_TRANSIT
	evt2 := ShipmentLifecycleEvent{
		EventID:       "evt-dep-1",
		ShipmentID:    101,
		EventType:     "DEPARTURE",
		MilestoneCode: "DEPARTED",
		Source:        "port_terminal_api",
		CorrelationID: "corr-dep-1",
	}
	updatedWF2, err := svc.ProcessShipmentEvent(ctx, 1, evt2)
	require.NoError(t, err)
	assert.Equal(t, StageInTransit, updatedWF2.CurrentStage)

	// 3. Loop Prevention: Self-originating event from autonomous platform must be suppressed
	selfEvt := ShipmentLifecycleEvent{
		EventID:       "evt-self-1",
		ShipmentID:    101,
		EventType:     "ETA_UPDATED",
		Source:        "enterprise_autonomous_platform", // Self-originating
		CorrelationID: "corr-self-1",
	}
	loopWF, err := svc.ProcessShipmentEvent(ctx, 1, selfEvt)
	require.NoError(t, err)
	// Workflow stage is preserved without recursive retriggering
	assert.Equal(t, StageInTransit, loopWF.CurrentStage)
}

// ---------------------------------------------------------------------
// 4. Event Deduplication Tests
// ---------------------------------------------------------------------

func TestShipmentEventDeduplication(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _, _, _ := setupShipmentLifecycleTest(t)

	_, err := svc.InitiateShipmentLifecycle(ctx, 1, 101, "corr-dedup-test")
	require.NoError(t, err)

	evt := ShipmentLifecycleEvent{
		EventID:       "evt-duplicate-unique-key",
		ShipmentID:    101,
		EventType:     "CARRIER_ASSIGNED",
		MilestoneCode: "BOOKED",
		Source:        "carrier_api",
		CorrelationID: "corr-dedup-1",
	}

	// First time: processed
	wf1, err := svc.ProcessShipmentEvent(ctx, 1, evt)
	require.NoError(t, err)
	assert.Equal(t, StageBooked, wf1.CurrentStage)

	// Second time: duplicate is gracefully suppressed without error or state corruption
	wf2, err := svc.ProcessShipmentEvent(ctx, 1, evt)
	require.NoError(t, err)
	assert.Equal(t, wf1.WorkflowID, wf2.WorkflowID)
}

// ---------------------------------------------------------------------
// 5. ETA Prediction Intelligence & Threshold Tests
// ---------------------------------------------------------------------

func TestPredictiveETAIntelligenceAndThreshold(t *testing.T) {
	ctx := context.Background()
	svc, _, mockShip, mockPred, _, _ := setupShipmentLifecycleTest(t)

	// 1. Initiate workflow
	_, err := svc.InitiateShipmentLifecycle(ctx, 1, 101, "corr-eta-test")
	require.NoError(t, err)

	// 2. Case A: Minor delay (1.5 hours) < operational threshold (4.0h)
	// Should update prediction facts without triggering an exception
	minorPredETA := mockShip.Shipment.ETA.Add(90 * time.Minute).Format(time.RFC3339)
	mockPred.PredictedETA = &minorPredETA
	mockPred.Confidence = 0.94

	wfMinor, err := svc.EvaluateETAPrediction(ctx, 1, 101)
	require.NoError(t, err)
	assert.Equal(t, 0.94, wfMinor.ETAConfidence)
	assert.InDelta(t, 1.5, wfMinor.PredictedDelayHours, 0.1)
	assert.False(t, wfMinor.IsDelayOperationallySignificant)
	assert.Equal(t, 0, wfMinor.ActiveExceptionsCount) // Below threshold

	// 3. Case B: Significant operational delay (8.0 hours) >= threshold (4.0h)
	// Must automatically flag exception and trigger multi-agent recovery
	sigPredETA := mockShip.Shipment.ETA.Add(8 * time.Hour).Format(time.RFC3339)
	mockPred.PredictedETA = &sigPredETA
	mockPred.Confidence = 0.89

	wfSig, err := svc.EvaluateETAPrediction(ctx, 1, 101)
	require.NoError(t, err)
	assert.InDelta(t, 8.0, wfSig.PredictedDelayHours, 0.1)
	assert.True(t, wfSig.IsDelayOperationallySignificant)
	assert.Greater(t, wfSig.ActiveExceptionsCount, 0)
	assert.Equal(t, "CRITICAL", wfSig.ExceptionSeverity)
}

// ---------------------------------------------------------------------
// 6. Multi-Agent Exception Investigation & Least-Privilege Selection
// ---------------------------------------------------------------------

func TestMultiAgentExceptionInvestigation(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _, _, _ := setupShipmentLifecycleTest(t)

	_, err := svc.InitiateShipmentLifecycle(ctx, 1, 101, "corr-inv-test")
	require.NoError(t, err)

	// Trigger autonomous investigation for port congestion exception
	wf, err := svc.InvestigateAndPlanRecovery(ctx, 1, 101, "PORT_CONGESTION", "HIGH", "Severe port congestion at destination transshipment hub")
	require.NoError(t, err)
	assert.Greater(t, wf.ActiveExceptionsCount, 0)
	assert.Equal(t, "HIGH", wf.ExceptionSeverity)

	// Verify least-privilege specialists dynamically assigned
	assert.Contains(t, wf.AssignedSpecialists, "shipment_agent")
	assert.Contains(t, wf.AssignedSpecialists, "exception_agent")
	assert.Contains(t, wf.AssignedSpecialists, "planning_agent")
	assert.Contains(t, wf.AssignedSpecialists, "customer_agent")
	assert.Contains(t, wf.AssignedSpecialists, "finance_agent")
	assert.Contains(t, wf.AssignedSpecialists, "contract_agent")

	// Verify structured recovery steps were generated
	require.NotEmpty(t, wf.Steps)
	assert.GreaterOrEqual(t, len(wf.Steps), 2)
}

// ---------------------------------------------------------------------
// 7. Recovery Planning & HITL Approval Gating
// ---------------------------------------------------------------------

func TestGovernedRecoveryPlanningAndApprovalGating(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _, mockApprovals, _ := setupShipmentLifecycleTest(t)

	// 1. Initiate workflow & investigate
	wf, err := svc.InitiateShipmentLifecycle(ctx, 1, 101, "corr-appr-test")
	require.NoError(t, err)

	wf, err = svc.InvestigateAndPlanRecovery(ctx, 1, 101, "CARRIER_ROLLOVER", "CRITICAL", "Container rolled to next vessel")
	require.NoError(t, err)

	// Look for the customer_advisory step which requires approval
	var advisoryStep *EnterpriseWorkflowStep
	for i := range wf.Steps {
		if wf.Steps[i].ActionType == "customer_advisory" || wf.Steps[i].RequiresApproval {
			advisoryStep = &wf.Steps[i]
			break
		}
	}
	require.NotNil(t, advisoryStep, "Expected customer_advisory step in recovery plan")
	assert.True(t, advisoryStep.RequiresApproval)

	// Execute the step: because it requires approval, it must transition to WAITING_FOR_APPROVAL
	updatedWF, err := svc.ExecuteGovernedRecoveryStep(ctx, 1, wf.WorkflowID, advisoryStep.StepID)
	require.NoError(t, err)
	assert.Equal(t, StateWaitingForApproval, updatedWF.WorkflowState)
	assert.Greater(t, updatedWF.PendingApprovalsCount, 0)

	// Verify approval request was recorded in ApprovalsService
	reqs, err := mockApprovals.ListApprovals(ctx, 1)
	require.NoError(t, err)
	assert.NotEmpty(t, reqs)
}

// ---------------------------------------------------------------------
// 8. Governed Action Execution & Verification
// ---------------------------------------------------------------------

func TestGovernedActionExecutionAndVerification(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _, _, _ := setupShipmentLifecycleTest(t)

	wf, err := svc.InitiateShipmentLifecycle(ctx, 1, 101, "corr-exec-test")
	require.NoError(t, err)

	wf, err = svc.InvestigateAndPlanRecovery(ctx, 1, 101, "DOCUMENT_DISCREPANCY", "MEDIUM", "Commercial invoice discrepancy")
	require.NoError(t, err)

	// Find carrier inquiry or milestone update step that does not require approval
	var autoStep *EnterpriseWorkflowStep
	for i := range wf.Steps {
		if !wf.Steps[i].RequiresApproval {
			autoStep = &wf.Steps[i]
			break
		}
	}
	require.NotNil(t, autoStep)
	assert.False(t, autoStep.RequiresApproval)

	// Execute the permitted step
	execWF, err := svc.ExecuteGovernedRecoveryStep(ctx, 1, wf.WorkflowID, autoStep.StepID)
	require.NoError(t, err)
	assert.Equal(t, StateRunning, execWF.WorkflowState)

	// Verify the executed action against authoritative data
	verResult, err := svc.VerifyActionExecution(ctx, 1, wf.WorkflowID, autoStep.StepID)
	require.NoError(t, err)
	assert.True(t, verResult.VerifiedSuccess)
	assert.Equal(t, 101, int(verResult.ShipmentID))
}

// ---------------------------------------------------------------------
// 9. Adaptive Replanning Tests
// ---------------------------------------------------------------------

func TestAdaptiveReplanning(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _, _, _ := setupShipmentLifecycleTest(t)

	wf, err := svc.InitiateShipmentLifecycle(ctx, 1, 101, "corr-replan-test")
	require.NoError(t, err)
	assert.Equal(t, 1, wf.ReplanVersion)

	// Trigger adaptive replanning due to major carrier delay
	replannedWF, err := svc.TriggerAdaptiveReplanning(ctx, 1, 101, "Major 48-hour transshipment blank sailing")
	require.NoError(t, err)
	assert.Equal(t, 2, replannedWF.ReplanVersion)
	assert.Equal(t, StagePlanning, replannedWF.CurrentStage)
	assert.Contains(t, replannedWF.AssignedSpecialists, "planning_agent")

	// Trigger second replan
	replannedWF2, err := svc.TriggerAdaptiveReplanning(ctx, 1, 101, "Secondary feeder schedule shift")
	require.NoError(t, err)
	assert.Equal(t, 3, replannedWF2.ReplanVersion)
}

// ---------------------------------------------------------------------
// 10. Delivery Transition & Post-Delivery Audit Tests
// ---------------------------------------------------------------------

func TestDeliveryTransitionAndPostDeliveryAudit(t *testing.T) {
	ctx := context.Background()
	svc, _, mockShip, _, _, _ := setupShipmentLifecycleTest(t)

	_, err := svc.InitiateShipmentLifecycle(ctx, 1, 101, "corr-del-test")
	require.NoError(t, err)

	// 1. Transition to DELIVERED
	deliveredWF, err := svc.TransitionToDelivered(ctx, 1, 101)
	require.NoError(t, err)
	assert.Equal(t, StageDelivered, deliveredWF.CurrentStage)
	assert.Equal(t, "DELIVERED", deliveredWF.AuthoritativeStatus)

	// 2. Case A: Active unresolved exception blocks final completion
	deliveredWF.ActiveExceptionsCount = 1
	mockShip.Exceptions = []*spec.ShipmentException{
		{
			ID:            555,
			ShipmentID:    101,
			ExceptionType: "CARRIER_DAMAGE_CLAIM",
			Severity:      "HIGH",
			Status:        "OPEN",
			Resolved:      false,
		},
	}
	auditBlockedWF, err := svc.ExecutePostDeliveryAudit(ctx, 1, 101)
	require.NoError(t, err)
	assert.Equal(t, StateBlocked, auditBlockedWF.WorkflowState)
	assert.Equal(t, StagePostDelivery, auditBlockedWF.CurrentStage)

	// 3. Case B: Exception cleared, post-delivery audit succeeds and transitions to COMPLETED
	deliveredWF.ActiveExceptionsCount = 0
	mockShip.Exceptions = []*spec.ShipmentException{} // None active
	auditCompleteWF, err := svc.ExecutePostDeliveryAudit(ctx, 1, 101)
	require.NoError(t, err)
	assert.Equal(t, StageCompleted, auditCompleteWF.CurrentStage)
	assert.Equal(t, StateCompleted, auditCompleteWF.WorkflowState)
}

// ---------------------------------------------------------------------
// 11. Outcome Learning & Memory Tests
// ---------------------------------------------------------------------

func TestOutcomeLearningAndMemory(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _, _, mockWorkforce := setupShipmentLifecycleTest(t)

	_, err := svc.InitiateShipmentLifecycle(ctx, 1, 101, "corr-out-test")
	require.NoError(t, err)

	outReq := RecordShipmentOutcomeRequest{
		MetricName:     "TRANSIT_TIME_VARIANCE",
		PredictedValue: "48.0",
		ActualValue:    "52.5",
		Feedback:       "Arrival delayed due to port congestion, recovery reroute reduced total delay by 12 hours.",
	}

	result, err := svc.RecordShipmentOutcome(ctx, 1, 101, outReq)
	require.NoError(t, err)
	assert.NotEmpty(t, result.OutcomeID)
	assert.True(t, result.LearningApplied)
	assert.Equal(t, 101, int(result.ShipmentID))

	// Verify workforce memory received recorded outcome
	assert.NotEmpty(t, mockWorkforce.recordedOutcomes)
	assert.Equal(t, "SHIPMENT", mockWorkforce.recordedOutcomes[0].SourceEntityType)
	assert.Equal(t, "101", mockWorkforce.recordedOutcomes[0].SourceEntityID)
	assert.True(t, mockWorkforce.recordedOutcomes[0].IsVerified)
}

// ---------------------------------------------------------------------
// 12. Tenant Isolation & Security Tests
// ---------------------------------------------------------------------

func TestShipmentLifecycleTenantIsolation(t *testing.T) {
	ctx := context.Background()
	svc, _, _, _, _, _ := setupShipmentLifecycleTest(t)

	// orgID <= 0 rejected
	_, err := svc.InitiateShipmentLifecycle(ctx, 0, 101, "unauth")
	assert.Equal(t, ErrUnauthorizedTenant, err)

	_, err = svc.GetShipmentLifecycle(ctx, -1, 101)
	assert.Equal(t, ErrUnauthorizedTenant, err)

	_, err = svc.ProcessShipmentEvent(ctx, 0, ShipmentLifecycleEvent{})
	assert.Equal(t, ErrUnauthorizedTenant, err)

	_, err = svc.EvaluateETAPrediction(ctx, 0, 101)
	assert.Equal(t, ErrUnauthorizedTenant, err)

	_, err = svc.TriggerAdaptiveReplanning(ctx, 0, 101, "unauth")
	assert.Equal(t, ErrUnauthorizedTenant, err)

	_, err = svc.TransitionToDelivered(ctx, 0, 101)
	assert.Equal(t, ErrUnauthorizedTenant, err)

	_, err = svc.RecordShipmentOutcome(ctx, 0, 101, RecordShipmentOutcomeRequest{})
	assert.Equal(t, ErrUnauthorizedTenant, err)
}

// ---------------------------------------------------------------------
// 13. Live MySQL Integration Test
// ---------------------------------------------------------------------

func TestLiveShipmentLifecycleIntegration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "freel:freel_secret@tcp(127.0.0.1:3306)/freel_db?parseTime=true&charset=utf8mb4"
	}
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		t.Skip("Live MySQL not available, skipping live test")
		return
	}
	if err := db.Ping(); err != nil {
		t.Skip("Live MySQL connection failed, skipping live test")
		return
	}
	defer db.Close()

	ctx := context.Background()
	repo := NewMySQLRepository(db)

	// Find an existing shipment in real DB if any
	var shipmentID int64
	var orgID int64
	err = db.QueryRowContext(ctx, "SELECT id, org_id FROM shipments ORDER BY id DESC LIMIT 1").Scan(&shipmentID, &orgID)
	if err != nil {
		t.Skip("No real shipments found in database, skipping live test")
		return
	}

	wfID := fmt.Sprintf("live-wf-ship-%d", shipmentID)
	now := time.Now().UTC()
	err = repo.CreateWorkflow(ctx, &EnterpriseWorkflow{
		WorkflowID:        wfID,
		OrgID:             orgID,
		WorkflowType:      WorkflowShipmentRecovery,
		Objective:         fmt.Sprintf("Live Shipment %d Lifecycle", shipmentID),
		InitiatingEvent:   "BOOKING_CONFIRMED",
		CurrentState:      StateRunning,
		AutonomyLevel:     "LEVEL_3_CONTROLLED_EXECUTION",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   fmt.Sprintf("%d", shipmentID),
		CorrelationID:     fmt.Sprintf("corr-live-%d", shipmentID),
		CreatedAt:         now,
		UpdatedAt:         now,
	})
	require.NoError(t, err)

	fetched, err := repo.GetWorkflow(ctx, orgID, wfID)
	require.NoError(t, err)
	assert.Equal(t, wfID, fetched.WorkflowID)
	assert.Equal(t, StateRunning, fetched.CurrentState)

	// Clean up
	_ = repo.UpdateWorkflowState(ctx, orgID, wfID, StateCompleted, nil, nil)
}
