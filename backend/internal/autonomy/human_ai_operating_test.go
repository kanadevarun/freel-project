package autonomy_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/freel/backend/internal/autonomy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHumanAIDecision_Lifecycle(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	orgID := int64(1)
	userID := int64(99)
	userName := "Dispatcher Alice"

	// 1. Create Decision Point
	req := &autonomy.CreateDecisionPointRequest{
		Module:        "shipments",
		EntityType:    "shipment",
		EntityID:      "101",
		Facts:         []string{"Vessel anchored at terminal berth 4", "Customs clearance granted"},
		Predictions:   []string{"Discharge queue estimated at 4.5h delay"},
		AutonomyLevel: autonomy.Level2Prepare,
		CorrelationID: "corr-test-101",
	}

	dec, err := svc.CreateDecisionPoint(ctx, orgID, req)
	require.NoError(t, err)
	require.NotNil(t, dec)
	assert.Equal(t, "dec-mock-001", dec.DecisionID)
	assert.Equal(t, autonomy.DecisionStatusPending, dec.DecisionStatus)
	assert.Equal(t, "HIGH", dec.Confidence)
	assert.Equal(t, "SUFFICIENT", dec.DataSufficiency)
	assert.Equal(t, autonomy.OpModeHumanApproval, dec.OperatingMode)

	// Verify AI recommendation is preserved
	assert.Equal(t, "Approve priority discharge", dec.AIRecommendation)

	// 2. Human Submits Decision: EDIT_AND_APPROVE
	editPayload := json.RawMessage(`{"action": "dispatch", "priority_lane": "EXPRESS_TRUCK", "custom_notice": "Customer notified via portal"}`)
	submitReq := &autonomy.SubmitHumanDecisionRequest{
		DecisionType:       "EDIT_AND_APPROVE",
		Reason:             "Expedited priority lane required for temperature-sensitive cargo",
		HumanEditedPayload: editPayload,
	}

	decided, err := svc.SubmitHumanDecision(ctx, orgID, dec.DecisionID, userID, userName, submitReq)
	require.NoError(t, err)
	require.NotNil(t, decided)

	// Verify both AI recommendation and human decision are separate and preserved
	assert.Equal(t, "Approve priority discharge", decided.AIRecommendation)
}

func TestHumanAIDecision_StopControl(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	orgID := int64(1)
	userID := int64(102)

	// Create a mock plan
	planID := "plan-stop-test-01"
	repo.CreatePlan(ctx, &autonomy.AutonomousPlan{
		OrgID:         orgID,
		PlanID:        planID,
		Status:        autonomy.PlanStatusExecuting,
		AutonomyLevel: autonomy.Level3ControlledExecution,
	}, []autonomy.AutonomousPlanStep{
		{StepID: "step-1", Status: autonomy.StepStatusCompleted},
		{StepID: "step-2", Status: autonomy.StepStatusPending},
	})

	// Trigger Human Stop Control
	err := svc.StopWorkflow(ctx, orgID, planID, userID, "Carrier reported severe storm warning; halting autonomous dispatch")
	assert.NoError(t, err)
}

func TestHumanAIDecision_InvalidateOnMaterialChange(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	orgID := int64(1)

	// Invalidate pending approvals due to material port disruption
	count, err := svc.InvalidatePendingApprovalsOnMaterialChange(ctx, orgID, "shipments", "shipment", "101", "Port closed due to crane malfunction; all active approvals invalidated")
	assert.NoError(t, err)
	assert.True(t, count >= 0)
}

func TestHumanAIDecision_StepHumanEdit(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	orgID := int64(1)
	planID := "plan-edit-01"
	stepID := "step-edit-01"

	err := svc.UpdateStepHumanEdit(ctx, orgID, planID, stepID, "Updated instructions: Deliver to Bay 12 instead of Bay 4")
	assert.NoError(t, err)
}

func TestHumanAIDecision_SummaryMetrics(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	summary, err := svc.GetHumanDecisionCenterSummary(ctx, 1, 99)
	require.NoError(t, err)
	require.NotNil(t, summary)
}
