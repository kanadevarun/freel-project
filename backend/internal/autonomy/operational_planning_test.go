package autonomy_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/freel/backend/internal/autonomy"
	"github.com/stretchr/testify/assert"
)

func TestOperationalPlanning_GoalCreationAndCandidateGeneration(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{
		generateResp: &autonomy.SidecarPlanGenResponse{
			PlanID:              "plan-ship-goal-001",
			Version:             1,
			Goal:                "Mitigate 24h delay on route",
			Module:              "shipments",
			RelatedEntityType:   "SHIPMENT",
			RelatedEntityID:     "101",
			CurrentStateSumm:    "Analyzed 3 candidate operational alternatives",
			ConfidenceScore:     0.90,
			DataSufficiency:     true,
			EstimatedImpact:     "18.5h delay recovery",
			RiskLevel:           "LOW",
			AutonomyLevel:       autonomy.Level2Prepare,
			SelectedCandidateID: "candidate-A",
			Candidates: []map[string]interface{}{
				{
					"candidate_id":  "candidate-A",
					"strategy_name": "Direct Carrier Expediting",
					"is_feasible":   true,
					"rank":          1,
					"evaluation": map[string]interface{}{
						"feasibility":                true,
						"hard_constraints_satisfied": true,
						"overall_utility_score":      0.92,
						"estimated_cost":             0.0,
					},
					"steps": []interface{}{
						map[string]interface{}{
							"step_id":          "step-1",
							"step_number":      1,
							"action_type":      "shipments.update_milestone",
							"title":            "Expedite Handling",
							"expected_outcome": "Carrier confirms priority transit",
						},
					},
				},
				{
					"candidate_id":  "candidate-B",
					"strategy_name": "Express Feeder Reroute",
					"is_feasible":   false,
					"rank":          2,
					"evaluation": map[string]interface{}{
						"feasibility":                false,
						"hard_constraints_satisfied": false,
						"overall_utility_score":      0.0,
						"estimated_cost":             450.0,
					},
				},
			},
			OrderedSteps: []autonomy.SidecarPlanStep{
				{
					StepID:           "step-1",
					StepNumber:       1,
					ActionType:       "shipments.update_milestone",
					Title:            "Expedite Handling",
					ExpectedOutcome:  "Carrier confirms priority transit",
					RiskLevel:        "LOW",
					RequiresApproval: true,
					IdempotencyKey:   "step-goal-001-1",
				},
			},
		},
	}

	_ = repo.SetPolicy(context.Background(), &autonomy.AutonomyPolicy{
		OrgID:                  1,
		Module:                 "shipments",
		AutonomyLevel:          autonomy.Level2Prepare,
		AllowedActionTypes:     json.RawMessage(`["shipments.update_milestone"]`),
		MinConfidenceThreshold: 0.80,
		RequiresApproval:       true,
		IsActive:               true,
	})

	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	userID := int64(10)

	// Create Planning Goal
	goalReq := autonomy.CreateGoalRequest{
		Module:            "shipments",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "101",
		Objective:         "Recover from 24h delay on route without exceeding budget",
		Priority:          "HIGH",
		RiskTolerance:     "BALANCED",
		HardConstraints: []autonomy.ConstraintDTO{
			{
				ConstraintType: "COST",
				Description:    "Max additional cost $200",
				IsHard:         true,
				ThresholdValue: 200.0,
			},
		},
		AutonomyLevel: autonomy.Level2Prepare,
	}

	goal, plan, steps, err := svc.CreatePlanningGoal(context.Background(), 1, &userID, goalReq)
	assert.NoError(t, err)
	assert.NotNil(t, goal)
	assert.Equal(t, "ACTIVE", goal.Status)
	assert.NotNil(t, plan)
	assert.Equal(t, "plan-ship-goal-001", plan.PlanID)
	assert.Equal(t, "candidate-A", plan.SelectedCandidateID.String)
	assert.Len(t, steps, 1)
}

func TestOperationalPlanning_HardConstraintSelectionGating(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{
		generateResp: &autonomy.SidecarPlanGenResponse{
			PlanID:              "plan-ship-candidates-002",
			Version:             1,
			Goal:                "Budget constrained operational recovery",
			Module:              "shipments",
			RelatedEntityType:   "SHIPMENT",
			RelatedEntityID:     "102",
			CurrentStateSumm:    "Analyzed alternatives",
			ConfidenceScore:     0.90,
			DataSufficiency:     true,
			SelectedCandidateID: "candidate-A",
			Candidates: []map[string]interface{}{
				{
					"candidate_id":  "candidate-A",
					"strategy_name": "Direct Expediting",
					"is_feasible":   true,
					"rank":          1,
					"evaluation": map[string]interface{}{
						"feasibility":                true,
						"hard_constraints_satisfied": true,
						"overall_utility_score":      0.90,
					},
					"steps": []interface{}{
						map[string]interface{}{
							"step_id":     "step-1",
							"step_number": 1,
							"action_type": "shipments.update_milestone",
						},
					},
				},
				{
					"candidate_id":  "candidate-B",
					"strategy_name": "Express Reroute",
					"is_feasible":   false, // Violated hard budget constraint
					"rank":          2,
					"evaluation": map[string]interface{}{
						"feasibility":                false,
						"hard_constraints_satisfied": false,
						"overall_utility_score":      0.0,
					},
				},
			},
			OrderedSteps: []autonomy.SidecarPlanStep{
				{
					StepID:     "step-1",
					StepNumber: 1,
					ActionType: "shipments.update_milestone",
				},
			},
		},
	}

	_ = repo.SetPolicy(context.Background(), &autonomy.AutonomyPolicy{
		OrgID:                  1,
		Module:                 "shipments",
		AutonomyLevel:          autonomy.Level2Prepare,
		AllowedActionTypes:     json.RawMessage(`["shipments.update_milestone"]`),
		MinConfidenceThreshold: 0.80,
		RequiresApproval:       true,
		IsActive:               true,
	})

	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	userID := int64(10)


	plan, _, err := svc.GeneratePlan(context.Background(), 1, &userID, autonomy.GeneratePlanRequest{
		Goal:              "Budget constrained operational recovery",
		Module:            "shipments",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "102",
		AutonomyLevel:     autonomy.Level2Prepare,
	})
	assert.NoError(t, err)

	// Attempting to select Candidate B (violates hard constraint) must be rejected
	_, _, err = svc.SelectPlanCandidate(context.Background(), 1, 10, plan.PlanID, "candidate-B")
	assert.Error(t, err)
	assert.ErrorIs(t, err, autonomy.ErrCandidateInfeasible)

	// Candidate A (feasible) is accepted
	updatedPlan, _, err := svc.SelectPlanCandidate(context.Background(), 1, 10, plan.PlanID, "candidate-A")
	assert.NoError(t, err)
	assert.Equal(t, "candidate-A", updatedPlan.SelectedCandidateID.String)
}

func TestOperationalPlanning_StalenessRevalidation(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{
		generateResp: &autonomy.SidecarPlanGenResponse{
			PlanID:            "plan-ship-staleness-003",
			Version:           1,
			Goal:              "Test staleness revalidation",
			Module:            "shipments",
			RelatedEntityType: "SHIPMENT",
			RelatedEntityID:   "103",
			CurrentStateSumm:  "Initial state",
			ConfidenceScore:   0.88,
			DataSufficiency:   true,
			StalenessStatus:   "FRESH",
			OrderedSteps: []autonomy.SidecarPlanStep{
				{
					StepID:     "step-1",
					StepNumber: 1,
					ActionType: "shipments.update_milestone",
				},
			},
		},
	}

	_ = repo.SetPolicy(context.Background(), &autonomy.AutonomyPolicy{
		OrgID:                  1,
		Module:                 "shipments",
		AutonomyLevel:          autonomy.Level2Prepare,
		AllowedActionTypes:     json.RawMessage(`["shipments.update_milestone"]`),
		MinConfidenceThreshold: 0.80,
		RequiresApproval:       true,
		IsActive:               true,
	})

	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	userID := int64(10)

	plan, _, err := svc.GeneratePlan(context.Background(), 1, &userID, autonomy.GeneratePlanRequest{
		Goal:              "Test staleness revalidation",
		Module:            "shipments",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "103",
		AutonomyLevel:     autonomy.Level2Prepare,
	})
	assert.NoError(t, err)
	assert.Equal(t, "FRESH", plan.StalenessStatus)

	// Revalidate plan
	revalPlan, err := svc.RevalidatePlan(context.Background(), 1, plan.PlanID)
	assert.NoError(t, err)
	assert.NotNil(t, revalPlan)
	assert.NotEmpty(t, revalPlan.StalenessStatus)
}

func TestOperationalPlanning_MultiTenantIsolation(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{
		generateResp: &autonomy.SidecarPlanGenResponse{
			PlanID:            "plan-tenant-org1-999",
			Version:           1,
			Goal:              "Org 1 Confidential Planning",
			Module:            "shipments",
			RelatedEntityType: "SHIPMENT",
			RelatedEntityID:   "999",
			ConfidenceScore:   0.88,
			DataSufficiency:   true,
			OrderedSteps:      []autonomy.SidecarPlanStep{},
		},
	}

	_ = repo.SetPolicy(context.Background(), &autonomy.AutonomyPolicy{
		OrgID:                  1,
		Module:                 "shipments",
		AutonomyLevel:          autonomy.Level2Prepare,
		AllowedActionTypes:     json.RawMessage(`["shipments.update_milestone"]`),
		MinConfidenceThreshold: 0.80,
		RequiresApproval:       true,
		IsActive:               true,
	})

	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	userID := int64(10)


	// Org 1 creates plan
	planOrg1, _, err := svc.GeneratePlan(context.Background(), 1, &userID, autonomy.GeneratePlanRequest{
		Goal:              "Org 1 Confidential Planning",
		Module:            "shipments",
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "999",
		AutonomyLevel:     autonomy.Level2Prepare,
	})
	assert.NoError(t, err)

	// Org 2 attempts to query Org 1's plan -> Must return ErrPlanNotFound
	_, _, err = svc.GetPlan(context.Background(), 2, planOrg1.PlanID)
	assert.ErrorIs(t, err, autonomy.ErrPlanNotFound)

	// Org 2 attempts to select candidate on Org 1's plan -> Must return ErrPlanNotFound
	_, _, err = svc.SelectPlanCandidate(context.Background(), 2, 20, planOrg1.PlanID, "candidate-A")
	assert.ErrorIs(t, err, autonomy.ErrPlanNotFound)

	// Org 2 attempts to revalidate Org 1's plan -> Must return ErrPlanNotFound
	_, err = svc.RevalidatePlan(context.Background(), 2, planOrg1.PlanID)
	assert.ErrorIs(t, err, autonomy.ErrPlanNotFound)
}
