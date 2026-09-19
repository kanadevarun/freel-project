package autonomy_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/freel/backend/internal/autonomy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiStepPlanning_SequentialExecution(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()
	orgID := int64(1)
	userID := int64(42)

	_ = repo.SetPolicy(ctx, &autonomy.AutonomyPolicy{
		OrgID:         orgID,
		Module:        "shipments",
		AutonomyLevel: autonomy.Level3ControlledExecution,
		IsActive:      true,
	})

	plan := &autonomy.AutonomousPlan{
		OrgID:             orgID,
		PlanID:            "plan-seq-001",
		Version:           1,
		Goal:              "Recover disrupted shipment",
		GoalType:          "MULTI_STEP_RECOVERY",
		Priority:          "HIGH",
		Module:            "shipments",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "101",
		Status:            autonomy.PlanStatusApproved,
		StalenessStatus:   "FRESH",
	}

	steps := []autonomy.AutonomousPlanStep{
		{
			PlanID:         plan.PlanID,
			OrgID:          orgID,
			StepNumber:     1,
			StepID:         "step-1",
			ActionType:     "shipments.update_milestone",
			Title:          "Analyze Disruption Impact",
			Status:         autonomy.StepStatusReady,
			IdempotencyKey: "idemp-seq-1",
			MaxAttempts:    3,
		},
		{
			PlanID:         plan.PlanID,
			OrgID:          orgID,
			StepNumber:     2,
			StepID:         "step-2",
			ActionType:     "shipments.update_milestone",
			Title:          "Notify Carrier for Recovery",
			Dependencies:   json.RawMessage(`["step-1"]`),
			Status:         autonomy.StepStatusPending,
			IdempotencyKey: "idemp-seq-2",
			MaxAttempts:    3,
		},
	}

	err := repo.CreatePlan(ctx, plan, steps)
	require.NoError(t, err)

	// Step 1 Execution
	res1, err := svc.ExecuteNextStep(ctx, orgID, userID, plan.PlanID)
	require.NoError(t, err)
	assert.Equal(t, "step-1", res1.StepID)
	assert.Equal(t, autonomy.StepStatusCompleted, res1.StepStatus)

	// Step 2 Execution (Dependency satisfied)
	res2, err := svc.ExecuteNextStep(ctx, orgID, userID, plan.PlanID)
	require.NoError(t, err)
	assert.Equal(t, "step-2", res2.StepID)
	assert.Equal(t, autonomy.StepStatusCompleted, res2.StepStatus)

	// Next call: all steps complete
	res3, err := svc.ExecuteNextStep(ctx, orgID, userID, plan.PlanID)
	require.NoError(t, err)
	assert.Equal(t, "ALL_STEPS_COMPLETED", res3.VerificationStatus)
}

func TestMultiStepPlanning_StepApprovalGating(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()
	orgID := int64(1)
	userID := int64(42)

	_ = repo.SetPolicy(ctx, &autonomy.AutonomyPolicy{
		OrgID:         orgID,
		Module:        "shipments",
		AutonomyLevel: autonomy.Level3ControlledExecution,
		IsActive:      true,
	})

	plan := &autonomy.AutonomousPlan{
		OrgID:             orgID,
		PlanID:            "plan-approval-001",
		Version:           1,
		Goal:              "High risk shipment reroute",
		GoalType:          "STANDARD",
		Priority:          "HIGH",
		Module:            "shipments",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "202",
		Status:            autonomy.PlanStatusApproved,
		StalenessStatus:   "FRESH",
	}

	steps := []autonomy.AutonomousPlanStep{
		{
			PlanID:           plan.PlanID,
			OrgID:            orgID,
			StepNumber:       1,
			StepID:           "step-approval-required",
			ActionType:       "shipments.reroute",
			Title:            "Execute Expensive Reroute",
			RequiresApproval: true,
			Status:           autonomy.StepStatusPending,
			IdempotencyKey:   "idemp-appr-1",
			MaxAttempts:      3,
		},
	}

	err := repo.CreatePlan(ctx, plan, steps)
	require.NoError(t, err)

	// Should be blocked by approval requirement
	_, err = svc.ExecuteNextStep(ctx, orgID, userID, plan.PlanID)
	require.Error(t, err)
	assert.ErrorIs(t, err, autonomy.ErrApprovalRequired)

	// Approve the step
	appResp, err := svc.ApproveStep(ctx, orgID, userID, plan.PlanID, "step-approval-required")
	require.NoError(t, err)
	assert.Equal(t, autonomy.StepStatusApproved, appResp.StepStatus)

	// Now execution should succeed
	res, err := svc.ExecuteNextStep(ctx, orgID, userID, plan.PlanID)
	require.NoError(t, err)
	assert.Equal(t, "step-approval-required", res.StepID)
	assert.Equal(t, autonomy.StepStatusCompleted, res.StepStatus)
}

func TestMultiStepPlanning_StalePlanProtection(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()
	orgID := int64(1)
	userID := int64(42)

	_ = repo.SetPolicy(ctx, &autonomy.AutonomyPolicy{
		OrgID:         orgID,
		Module:        "shipments",
		AutonomyLevel: autonomy.Level3ControlledExecution,
		IsActive:      true,
	})

	plan := &autonomy.AutonomousPlan{
		OrgID:             orgID,
		PlanID:            "plan-stale-001",
		Version:           1,
		Goal:              "Outdated route recovery",
		Module:            "shipments",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "303",
		Status:            autonomy.PlanStatusApproved,
		StalenessStatus:   "STALE", // STALE PLAN
	}

	steps := []autonomy.AutonomousPlanStep{
		{
			PlanID:         plan.PlanID,
			OrgID:          orgID,
			StepNumber:     1,
			StepID:         "step-stale-1",
			ActionType:     "shipments.update_milestone",
			Status:         autonomy.StepStatusReady,
			IdempotencyKey: "idemp-stale-1",
		},
	}

	err := repo.CreatePlan(ctx, plan, steps)
	require.NoError(t, err)

	_, err = svc.ExecuteNextStep(ctx, orgID, userID, plan.PlanID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is stale")
}

func TestMultiStepPlanning_ConcurrentPlanConflicts(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()
	orgID := int64(1)

	// Plan A: Medium priority
	planA := &autonomy.AutonomousPlan{
		OrgID:             orgID,
		PlanID:            "plan-A",
		Version:           1,
		Goal:              "Recover shipment A",
		Priority:          "MEDIUM",
		Module:            "shipments",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "SHIP-999",
		Status:            autonomy.PlanStatusExecuting,
		CreatedAt:         time.Now().Add(-10 * time.Minute),
	}
	_ = repo.CreatePlan(ctx, planA, nil)

	// Plan B: Critical priority on same entity
	planB := &autonomy.AutonomousPlan{
		OrgID:             orgID,
		PlanID:            "plan-B",
		Version:           1,
		Goal:              "Emergency expedite shipment A",
		Priority:          "CRITICAL",
		Module:            "shipments",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "SHIP-999",
		Status:            autonomy.PlanStatusApproved,
		CreatedAt:         time.Now(),
	}
	_ = repo.CreatePlan(ctx, planB, nil)

	// Check conflicts for Plan A
	confA, err := svc.CheckPlanConflicts(ctx, orgID, "plan-A")
	require.NoError(t, err)
	assert.True(t, confA.HasConflict)
	assert.Equal(t, "BLOCKED_BY_HIGHER_PRIORITY", confA.PriorityAction)

	// Check entity level conflicts
	confEntity, err := svc.ListEntityConflicts(ctx, orgID, "SHIPMENT", "SHIP-999")
	require.NoError(t, err)
	assert.True(t, confEntity.HasConflict)
	assert.Len(t, confEntity.ActivePlans, 2)
}

func TestMultiStepPlanning_PlanValidationAndMetrics(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()
	orgID := int64(1)

	_ = repo.SetPolicy(ctx, &autonomy.AutonomyPolicy{
		OrgID:                 orgID,
		Module:                "shipments",
		AutonomyLevel:         autonomy.Level3ControlledExecution,
		ProhibitedActionTypes: json.RawMessage(`["shipments.force_cancel"]`),
		IsActive:              true,
	})

	plan := &autonomy.AutonomousPlan{
		OrgID:             orgID,
		PlanID:            "plan-val-001",
		Version:           1,
		Goal:              "Validation test plan",
		Module:            "shipments",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "404",
		Status:            autonomy.PlanStatusGenerated,
	}

	steps := []autonomy.AutonomousPlanStep{
		{
			PlanID:         plan.PlanID,
			OrgID:          orgID,
			StepNumber:     1,
			StepID:         "step-val-1",
			ActionType:     "shipments.force_cancel", // Prohibited!
			IdempotencyKey: "idemp-val-1",
		},
	}

	_ = repo.CreatePlan(ctx, plan, steps)

	valResp, err := svc.ValidatePlan(ctx, orgID, plan.PlanID)
	require.NoError(t, err)
	assert.False(t, valResp.IsValid)
	assert.NotEmpty(t, valResp.Issues)
	assert.Contains(t, valResp.Issues[0], "is prohibited by tenant policy")

	// Verify metrics
	metrics, err := svc.GetPlanningMetrics(ctx, orgID)
	require.NoError(t, err)
	assert.Equal(t, orgID, metrics.OrgID)
}

func TestMultiStepPlanning_RetryAndCompensate(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()
	orgID := int64(1)
	userID := int64(42)

	plan := &autonomy.AutonomousPlan{
		OrgID:             orgID,
		PlanID:            "plan-retry-001",
		Version:           1,
		Goal:              "Retry test plan",
		Module:            "shipments",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "505",
		Status:            autonomy.PlanStatusExecuting,
	}

	steps := []autonomy.AutonomousPlanStep{
		{
			PlanID:           plan.PlanID,
			OrgID:            orgID,
			StepNumber:       1,
			StepID:           "step-fail-1",
			ActionType:       "shipments.update_milestone",
			Status:           autonomy.StepStatusFailed,
			ExecutionAttempt: 1,
			MaxAttempts:      3,
			IdempotencyKey:   "idemp-fail-1",
		},
	}
	_ = repo.CreatePlan(ctx, plan, steps)

	// Retry step
	retried, err := svc.RetryStep(ctx, orgID, userID, plan.PlanID, "step-fail-1")
	require.NoError(t, err)
	assert.Equal(t, autonomy.StepStatusReady, retried.Status)

	// Compensate step
	compensated, err := svc.CompensateStep(ctx, orgID, userID, plan.PlanID, "step-fail-1")
	require.NoError(t, err)
	assert.Equal(t, autonomy.StepStatusCancelled, compensated.Status)
}

func TestMultiStepPlanning_CrossModulePlan(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()
	orgID := int64(1)

	_ = repo.SetPolicy(ctx, &autonomy.AutonomyPolicy{
		OrgID:         orgID,
		Module:        "shipments",
		AutonomyLevel: autonomy.Level3ControlledExecution,
		IsActive:      true,
	})

	req := autonomy.CrossModulePlanRequest{
		Goal:            "Coordinate shipment disruption with customer and finance",
		PrimaryModule:   "shipments",
		PrimaryEntityID: "SH-8888",
		InvolvedModules: []string{"shipments", "customers", "finance"},
		Priority:        "HIGH",
	}

	plan, steps, err := svc.GenerateCrossModulePlan(ctx, orgID, nil, req)
	require.NoError(t, err)
	assert.Equal(t, "CROSS_MODULE", plan.GoalType)
	assert.NotEmpty(t, steps)
	assert.Equal(t, "cm-step-1", steps[0].StepID)
}
