package autonomy_test

import (
	"context"
	"testing"

	"github.com/freel/backend/internal/autonomy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGovernanceEngine_KillSwitchBlock(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	govEngine := autonomy.NewGovernanceEngine(repo, sidecar)

	// Activate kill switch for Org 1
	err := repo.ToggleKillSwitch(context.Background(), 1, true, "Emergency drill test", 999)
	require.NoError(t, err)

	req := autonomy.GovernanceEvaluationRequest{
		OrgID:             1,
		UserID:            10,
		UserRole:          "operations",
		Module:            "shipments",
		ActionType:        "shipments.update_milestone",
		EntityType:        "SHIPMENT",
		EntityID:          "101",
		RequestedAutonomy: 3,
	}

	resp, err := govEngine.Evaluate(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, autonomy.GovernanceDecisionBlock, resp.Decision)
	assert.Equal(t, 0, resp.EffectiveAutonomy)
	assert.Contains(t, resp.Reasons[0], "emergency stop is currently ACTIVE")
}

func TestGovernanceEngine_FeatureFlagBlock(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	govEngine := autonomy.NewGovernanceEngine(repo, sidecar)

	// Disable shipments flag
	err := repo.UpdateFeatureFlag(context.Background(), 1, "shipment_automation", false, 0, false, 999)
	require.NoError(t, err)

	req := autonomy.GovernanceEvaluationRequest{
		OrgID:             1,
		UserID:            10,
		UserRole:          "operations",
		Module:            "shipments",
		ActionType:        "shipments.update_milestone",
		EntityType:        "SHIPMENT",
		EntityID:          "101",
		RequestedAutonomy: 3,
	}

	resp, err := govEngine.Evaluate(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, autonomy.GovernanceDecisionBlock, resp.Decision)
	assert.Contains(t, resp.Reasons[0], "disabled by tenant feature flag")
}

func TestGovernanceEngine_AllowlistGating(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	govEngine := autonomy.NewGovernanceEngine(repo, sidecar)

	// Action not on allowlist
	req := autonomy.GovernanceEvaluationRequest{
		OrgID:             1,
		UserID:            10,
		UserRole:          "operations",
		Module:            "shipments",
		ActionType:        "arbitrary.unauthorized_mutation",
		EntityType:        "SHIPMENT",
		EntityID:          "101",
		RequestedAutonomy: 3,
	}

	resp, err := govEngine.Evaluate(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, autonomy.GovernanceDecisionBlock, resp.Decision)
	assert.Contains(t, resp.Reasons[0], "not registered or enabled in tenant allowlist")
}

func TestGovernanceEngine_FourEyesControl(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	govEngine := autonomy.NewGovernanceEngine(repo, sidecar)

	// High impact action where preparer == approver
	req := autonomy.GovernanceEvaluationRequest{
		OrgID:              1,
		UserID:             10,
		UserRole:           "finance_manager",
		PreparerUserID:     10,
		ApprovalApproverID: 10,
		IsApproved:         true,
		Module:             "finance",
		ActionType:         "finance.apply_invoice_discount",
		EntityType:         "INVOICE",
		EntityID:           "INV-2026-001",
		RequestedAutonomy:  3,
	}

	resp, err := govEngine.Evaluate(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, autonomy.GovernanceDecisionBlock, resp.Decision)
	assert.True(t, resp.FourEyesEnforced)
	assert.Contains(t, resp.Reasons[0], "Four-Eyes Control Violation")
}

func TestGovernanceEngine_ComplianceStatusBlock(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	govEngine := autonomy.NewGovernanceEngine(repo, sidecar)

	req := autonomy.GovernanceEvaluationRequest{
		OrgID:            1,
		UserID:           10,
		UserRole:         "compliance_manager",
		IsApproved:       true,
		Module:           "compliance",
		ActionType:       "compliance.record_sanctions_check",
		EntityType:       "CONTRACT",
		EntityID:         "CON-001",
		RequestedAutonomy: 3,
		ComplianceStatus: "BLOCKED",
	}

	resp, err := govEngine.Evaluate(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, autonomy.GovernanceDecisionBlock, resp.Decision)
	assert.Contains(t, resp.Reasons[0], "BLOCKED compliance status")
}

func TestGovernanceEngine_EffectiveAutonomyCalculation(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	govEngine := autonomy.NewGovernanceEngine(repo, sidecar)

	// Seed limit with max autonomy = 2 (Prepare only)
	err := repo.UpdateTenantLimits(context.Background(), &autonomy.TenantGovernanceLimits{
		OrgID:             1,
		MaxTenantAutonomy: 2,
		KillSwitchActive:  false,
		MaxActionsPerHour: 100,
		EnforceFourEyes:   true,
	})
	require.NoError(t, err)

	req := autonomy.GovernanceEvaluationRequest{
		OrgID:             1,
		UserID:            10,
		UserRole:          "operations",
		Module:            "shipments",
		ActionType:        "shipments.update_milestone",
		EntityType:        "SHIPMENT",
		EntityID:          "101",
		RequestedAutonomy: 4, // Requested Level 4
	}

	resp, err := govEngine.Evaluate(context.Background(), req)
	require.NoError(t, err)
	// Effective autonomy capped at tenant limit 2
	assert.LessOrEqual(t, resp.EffectiveAutonomy, 2)
}

func TestGovernanceEngine_PlanPreview(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	govEngine := autonomy.NewGovernanceEngine(repo, sidecar)

	steps := []map[string]interface{}{
		{
			"step_id":     "step-1",
			"step_number": 1,
			"action_type": "shipments.update_milestone",
			"title":       "Update ETA Milestone",
		},
	}

	preview, err := govEngine.PreviewPlan(context.Background(), autonomy.PlanPreviewRequest{
		OrgID:  1,
		PlanID: "plan-preview-1",
		Goal:   "Update shipment milestones",
		Module: "shipments",
		Steps:  steps,
	})

	require.NoError(t, err)
	assert.NotNil(t, preview)
	assert.Equal(t, "plan-preview-1", preview.PlanID)
	assert.Equal(t, 1, preview.TotalSteps)
}

func TestGovernanceEngine_FinancialExposureReview(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	govEngine := autonomy.NewGovernanceEngine(repo, sidecar)

	// Financial amount exceeds workflow limit of 5000.0
	req := autonomy.GovernanceEvaluationRequest{
		OrgID:             1,
		UserID:            10,
		UserRole:          "operations",
		Module:            "shipments",
		ActionType:        "shipments.update_milestone",
		EntityType:        "SHIPMENT",
		EntityID:          "101",
		RequestedAutonomy: 3,
		Parameters: map[string]interface{}{
			"cost_impact": 15000.0,
		},
	}

	resp, err := govEngine.Evaluate(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, autonomy.GovernanceDecisionRequireReview, resp.Decision)
	assert.True(t, resp.RequiresApproval)
	assert.Contains(t, resp.Reasons[0], "Financial exposure")
}

func TestGovernanceEngine_TerminalEntityStateBlock(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	govEngine := autonomy.NewGovernanceEngine(repo, sidecar)

	req := autonomy.GovernanceEvaluationRequest{
		OrgID:             1,
		UserID:            10,
		UserRole:          "operations",
		Module:            "shipments",
		ActionType:        "shipments.update_milestone",
		EntityType:        "SHIPMENT",
		EntityID:          "101",
		EntityState:       "CANCELLED",
		RequestedAutonomy: 3,
	}

	resp, err := govEngine.Evaluate(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, autonomy.GovernanceDecisionBlock, resp.Decision)
	assert.Contains(t, resp.Reasons[0], "terminal state 'CANCELLED'")
}
