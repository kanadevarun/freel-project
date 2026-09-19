package autonomy_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/freel/backend/internal/autonomy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExceptionResolution_CustomsHoldEvaluation(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	ctx := context.Background()
	orgID := int64(2)
	excID := int64(101)

	resp, err := svc.EvaluateExceptionResolution(ctx, orgID, nil, excID)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, excID, resp.ExceptionID)
	assert.Equal(t, "CUSTOMS_HOLD", resp.ExceptionType)
	assert.Equal(t, "CRITICAL", resp.Severity)
	assert.Equal(t, "strat-customs-document-remedy", resp.Plan.SelectedStrategyID)
	assert.True(t, resp.Plan.RequiresApproval, "CRITICAL customs hold must require approval")
	assert.Equal(t, "WAITING_FOR_APPROVAL", *resp.Plan.WaitingState)

	// 5 candidate recovery strategies
	assert.Len(t, resp.Candidates, 5)
	// 7 sequential recovery plan steps
	assert.Len(t, resp.RecoveryPlanSteps, 7)
	// Root cause vs symptom
	assert.NotEmpty(t, resp.Symptom)
	assert.NotEmpty(t, resp.LikelyRootCause)
}

func TestExceptionResolution_EtaDelayEvaluation(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	ctx := context.Background()
	orgID := int64(2)
	excID := int64(102)

	resp, err := svc.EvaluateExceptionResolution(ctx, orgID, nil, excID)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, excID, resp.ExceptionID)
	assert.Equal(t, "ETA_DELAY", resp.ExceptionType)
	assert.Equal(t, "strat-carrier-escalation", resp.Plan.SelectedStrategyID)
	assert.Equal(t, "WAITING_FOR_CARRIER", *resp.Plan.WaitingState)
}

func TestExceptionResolution_SelectStrategy_FeasibleVsInfeasible(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	ctx := context.Background()
	orgID := int64(2)
	excID := int64(101)
	uid := int64(10)

	// Initial evaluation
	_, err := svc.EvaluateExceptionResolution(ctx, orgID, &uid, excID)
	require.NoError(t, err)

	// Selecting feasible candidate 'strat-customer-proactive-notice'
	updated, err := svc.SelectExceptionResolutionStrategy(ctx, orgID, &uid, excID, "strat-customer-proactive-notice")
	require.NoError(t, err)
	assert.Equal(t, "strat-customer-proactive-notice", updated.Plan.SelectedStrategyID)

	// Selecting infeasible candidate 'strat-re-route-alternate-corridor' for Customs Hold must fail
	_, err = svc.SelectExceptionResolutionStrategy(ctx, orgID, &uid, excID, "strat-re-route-alternate-corridor")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "infeasible")
}

func TestExceptionResolution_ApprovalGatingAndExecution(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	ctx := context.Background()
	orgID := int64(2)
	excID := int64(101) // CRITICAL customs hold -> requires approval

	// Evaluate
	_, err := svc.EvaluateExceptionResolution(ctx, orgID, nil, excID)
	require.NoError(t, err)

	// Attempt executing without user context (unauthorized automated attempt)
	_, err = svc.ExecuteExceptionResolutionAction(ctx, orgID, nil, excID, nil)
	assert.Error(t, err, "Action requiring approval must fail without authorized user context")
	assert.Contains(t, err.Error(), "approval")

	// Execute with authorized supervisor user ID
	uid := int64(42)
	res, err := svc.ExecuteExceptionResolutionAction(ctx, orgID, &uid, excID, nil)
	require.NoError(t, err)
	assert.True(t, res.Success)

	// Verify plan transitioned to RESOLVING and waiting state updated
	state, err := svc.GetExceptionResolutionState(ctx, orgID, excID)
	require.NoError(t, err)
	assert.Equal(t, "RESOLVING", state.Plan.LifecycleStatus)
	assert.Equal(t, "WAITING_FOR_DOCUMENT", *state.Plan.WaitingState)
}

func TestExceptionResolution_EmergencyStopPolicy(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}

	// Set Emergency Stop on module 'exceptions'
	_ = repo.SetPolicy(context.Background(), &autonomy.AutonomyPolicy{
		OrgID:                  2,
		Module:                 "exceptions",
		AutonomyLevel:          autonomy.Level2Prepare,
		AllowedActionTypes:     json.RawMessage(`["carrier_inquiry"]`),
		MinConfidenceThreshold: 0.80,
		EmergencyStop:          true, // STOPPED
		IsActive:               true,
	})

	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()
	orgID := int64(2)
	excID := int64(102)
	uid := int64(10)

	_, err := svc.EvaluateExceptionResolution(ctx, orgID, &uid, excID)
	require.NoError(t, err)

	// Execution must be halted by emergency stop policy
	_, err = svc.ExecuteExceptionResolutionAction(ctx, orgID, &uid, excID, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "emergency stop")
}

func TestExceptionResolution_Replanning(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	ctx := context.Background()
	orgID := int64(2)
	excID := int64(102)
	uid := int64(10)

	// Evaluate initial plan (version 1)
	initial, err := svc.EvaluateExceptionResolution(ctx, orgID, &uid, excID)
	require.NoError(t, err)
	assert.Equal(t, 1, initial.Plan.Version)

	// Replan on carrier update
	replanned, err := svc.ReplanExceptionResolution(ctx, orgID, &uid, excID, "CARRIER_UPDATE", map[string]interface{}{
		"revised_eta": "2026-03-24",
	})
	require.NoError(t, err)
	assert.Equal(t, 2, replanned.Plan.Version)
	assert.Equal(t, "RESOLVING", replanned.Plan.LifecycleStatus)
	assert.Equal(t, "WAITING_FOR_VERIFICATION", *replanned.Plan.WaitingState)
}
