package enterprise_autonomy

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------
// Phase 7.9: Enterprise Autonomous Control Tower Test Suite
// ---------------------------------------------------------------------

func setupControlTowerTestService() (EnterpriseControlTowerService, *MockEnterpriseRepository) {
	repo := NewMockEnterpriseRepository()
	ctSvc := NewEnterpriseControlTowerService(
		repo,
		nil, // workforceSvc
		nil, // actionsSvc
		nil, // approvalsSvc
		nil, // auditSvc
		nil, // predictionsSvc
		nil, // shipmentLifecycleSvc
		nil, // commercialLifecycleSvc
		nil, // exceptionManagementSvc
		nil, // customerRelationshipSvc
		nil, // revenueOptimizationSvc
		nil, // contractComplianceRiskSvc
		nil, // eventMeshSvc
		nil, // govSvc
		nil, // resilienceSvc
		nil, // db
	)
	return ctSvc, repo
}

// 1. Empty State & Clean View Generation
func TestControlTowerEmptyState(t *testing.T) {
	svc, _ := setupControlTowerTestService()
	ctx := context.Background()
	orgID := int64(1)

	view, err := svc.GetControlTowerView(ctx, orgID)
	require.NoError(t, err)
	assert.NotNil(t, view)
	assert.Equal(t, orgID, view.OrgID)
	assert.Empty(t, view.HumanAttentionItems, "Should have zero attention items when no issues exist")
	assert.Empty(t, view.ActiveWorkflows, "Should have zero active workflows")
	assert.Len(t, view.DomainSummaries, 5, "Must return 5 core domain summaries (Shipments, Commercial, Finance, Contracts/Compliance, Customers)")
	assert.False(t, view.WorkforceHealth.EmergencyStopActive)
	assert.Equal(t, "HEALTHY", view.PlatformHealth)

	// Verify all 5 domains are present in map
	assert.Contains(t, view.DomainSummaries, "SHIPMENTS")
	assert.Contains(t, view.DomainSummaries, "COMMERCIAL")
	assert.Contains(t, view.DomainSummaries, "FINANCE")
	assert.Contains(t, view.DomainSummaries, "CONTRACTS_COMPLIANCE")
	assert.Contains(t, view.DomainSummaries, "CUSTOMERS")
}

// 2. Prioritized Human Attention Items Generation
func TestControlTowerHumanAttentionPrioritization(t *testing.T) {
	svc, repo := setupControlTowerTestService()
	ctx := context.Background()
	orgID := int64(1)
	now := time.Now().UTC()

	// Seed multiple workflows of various severities and states
	wfCriticalEscalated := &EnterpriseWorkflow{
		WorkflowID:        "wf-esc-001",
		OrgID:             orgID,
		WorkflowType:      WorkflowShipmentRecovery,
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "SH-CRIT-99",
		CurrentState:      StateEscalated,
		AutonomyLevel:     "LEVEL_3_CONTROLLED_EXECUTION",
		Objective:         "Critical delay re-routing escalated due to carrier SLA breach",
		Confidence:        0.88,
		CreatedAt:         now.Add(-1 * time.Hour),
		UpdatedAt:         now,
	}
	require.NoError(t, repo.CreateWorkflow(ctx, wfCriticalEscalated))

	wfWaitingApproval := &EnterpriseWorkflow{
		WorkflowID:        "wf-wait-002",
		OrgID:             orgID,
		WorkflowType:      WorkflowRevenueOptimization,
		RelatedEntityType: "QUOTE",
		RelatedEntityID:   "Q-HIGH-42",
		CurrentState:      StateWaitingForApproval,
		AutonomyLevel:     "LEVEL_2_PREPARE",
		Objective:         "Margin defense concession requires executive discount approval",
		Confidence:        0.91,
		CreatedAt:         now.Add(-30 * time.Minute),
		UpdatedAt:         now,
	}
	require.NoError(t, repo.CreateWorkflow(ctx, wfWaitingApproval))

	wfHealthyRunning := &EnterpriseWorkflow{
		WorkflowID:        "wf-run-003",
		OrgID:             orgID,
		WorkflowType:      WorkflowContractComplianceRisk,
		RelatedEntityType: "CONTRACT",
		RelatedEntityID:   "CTR-101",
		CurrentState:      StateRunning,
		AutonomyLevel:     "LEVEL_4_FULL_AUTONOMY",
		Objective:         "Routine commercial rate card verification",
		Confidence:        0.98,
		CreatedAt:         now.Add(-5 * time.Minute),
		UpdatedAt:         now,
	}
	require.NoError(t, repo.CreateWorkflow(ctx, wfHealthyRunning))

	view, err := svc.GetControlTowerView(ctx, orgID)
	require.NoError(t, err)

	// We expect 2 attention items: WAITING_FOR_APPROVAL (HIGH) and ESCALATED (CRITICAL)
	assert.Len(t, view.HumanAttentionItems, 2)

	// Active workflows should contain all 3
	assert.Len(t, view.ActiveWorkflows, 3)

	// Check recent escalations
	assert.Len(t, view.RecentEscalations, 1)
	assert.Equal(t, "wf-esc-001", *view.RecentEscalations[0].WorkflowID)
	assert.Equal(t, AttentionSeverityCritical, view.RecentEscalations[0].Severity)
}

// 3. Autonomy Visibility & Five Autonomy Levels
func TestControlTowerAutonomyVisibility(t *testing.T) {
	svc, repo := setupControlTowerTestService()
	ctx := context.Background()
	orgID := int64(1)

	levels := []string{
		"LEVEL_0_OBSERVE",
		"LEVEL_1_RECOMMEND",
		"LEVEL_2_PREPARE",
		"LEVEL_3_CONTROLLED_EXECUTION",
		"LEVEL_4_FULL_AUTONOMY",
	}

	for i, lvl := range levels {
		wf := &EnterpriseWorkflow{
			WorkflowID:        "wf-lvl-" + string(rune('A'+i)),
			OrgID:             orgID,
			WorkflowType:      WorkflowShipmentRecovery,
			RelatedEntityType: "SHIPMENT",
			RelatedEntityID:   "SH-LVL",
			CurrentState:      StateRunning,
			AutonomyLevel:     lvl,
			Objective:         "Autonomy verification level " + lvl,
			CreatedAt:         time.Now().UTC().Add(time.Duration(-i) * time.Minute),
			UpdatedAt:         time.Now().UTC(),
		}
		require.NoError(t, repo.CreateWorkflow(ctx, wf))
	}

	view, err := svc.GetControlTowerView(ctx, orgID)
	require.NoError(t, err)
	assert.Len(t, view.ActiveWorkflows, 5)

	// Verify autonomy counts in overview
	assert.Equal(t, 1, view.AutonomyOverview["LEVEL_0_OBSERVE"])
	assert.Equal(t, 1, view.AutonomyOverview["LEVEL_1_RECOMMEND"])
	assert.Equal(t, 1, view.AutonomyOverview["LEVEL_2_PREPARE"])
	assert.Equal(t, 1, view.AutonomyOverview["LEVEL_3_CONTROLLED_EXECUTION"])
	assert.Equal(t, 1, view.AutonomyOverview["LEVEL_4_FULL_AUTONOMY"])
}

// 4. Complete Workflow Trace (Lineage: Event -> Workflow -> Steps -> Actions -> Outcomes)
func TestControlTowerWorkflowTrace(t *testing.T) {
	svc, repo := setupControlTowerTestService()
	ctx := context.Background()
	orgID := int64(1)
	now := time.Now().UTC()

	wf := &EnterpriseWorkflow{
		WorkflowID:        "wf-trace-123",
		OrgID:             orgID,
		WorkflowType:      WorkflowShipmentRecovery,
		RelatedEntityType: "SHIPMENT",
		RelatedEntityID:   "SH-9901",
		CurrentState:      StateRunning,
		AutonomyLevel:     "LEVEL_3_CONTROLLED_EXECUTION",
		Objective:         "Automated re-routing via alternative feeder vessel",
		Confidence:        0.94,
		InitiatingEvent:   "evt-disrupt-777",
		CorrelationID:     "corr-xyz-888",
		PolicyDecision:    "Governed Autonomous Execution (Autonomy Level 3)",
		CreatedAt:         now.Add(-10 * time.Minute),
		UpdatedAt:         now,
		Steps: []EnterpriseWorkflowStep{
			{
				StepID:     "step-1",
				WorkflowID: "wf-trace-123",
				StepNumber: 1,
				Title:      "Disruption Assessment",
				AgentID:    "ShipmentAgent",
				Status:     "COMPLETED",
				CreatedAt:  now.Add(-9 * time.Minute),
			},
			{
				StepID:     "step-2",
				WorkflowID: "wf-trace-123",
				StepNumber: 2,
				Title:      "Alternative Route Formulation",
				AgentID:    "RoutingAgent",
				Status:     "RUNNING",
				CreatedAt:  now.Add(-5 * time.Minute),
			},
		},
	}
	require.NoError(t, repo.CreateWorkflow(ctx, wf))

	trace, err := svc.GetWorkflowTrace(ctx, orgID, "wf-trace-123")
	require.NoError(t, err)
	assert.NotNil(t, trace)
	assert.Equal(t, "wf-trace-123", trace.WorkflowID)
	assert.Equal(t, "evt-disrupt-777", trace.InitiatingEvent)
	assert.Equal(t, "corr-xyz-888", trace.CorrelationID)
	assert.Len(t, trace.Steps, 2)
	assert.Equal(t, "Governed Autonomous Execution (Autonomy Level 3)", trace.PolicyDecision)
	assert.Equal(t, "AUTHORITATIVE_VERIFIED", trace.VerificationStatus)
	assert.Equal(t, "SHIPMENT", trace.AuthoritativeFacts["source_entity"])
}

// 5. Governed Control Actions (PAUSE, RESUME, CANCEL) and Tenant Isolation
func TestControlTowerGovernedControlActions(t *testing.T) {
	svc, repo := setupControlTowerTestService()
	ctx := context.Background()
	orgID := int64(1)
	otherOrgID := int64(999)

	wf := &EnterpriseWorkflow{
		WorkflowID:        "wf-ctrl-555",
		OrgID:             orgID,
		WorkflowType:      WorkflowRevenueOptimization,
		RelatedEntityType: "INVOICE",
		RelatedEntityID:   "INV-888",
		CurrentState:      StateRunning,
		AutonomyLevel:     "LEVEL_3_CONTROLLED_EXECUTION",
		Objective:         "Overdue collections autonomous contact sequence",
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	require.NoError(t, repo.CreateWorkflow(ctx, wf))

	// Tenant Isolation: Attempting action from other tenant must fail
	err := svc.PerformGovernedControlAction(ctx, otherOrgID, "wf-ctrl-555", "PAUSE", "admin@logisticshq.io", "Cross tenant test")
	assert.Error(t, err)

	// Action 1: PAUSE
	err = svc.PerformGovernedControlAction(ctx, orgID, "wf-ctrl-555", "PAUSE", "admin@logisticshq.io", "Operator manual review requested")
	require.NoError(t, err)

	savedWf, err := repo.GetWorkflow(ctx, orgID, "wf-ctrl-555")
	require.NoError(t, err)
	assert.Equal(t, StatePaused, savedWf.CurrentState)

	// Action 2: RESUME
	err = svc.PerformGovernedControlAction(ctx, orgID, "wf-ctrl-555", "RESUME", "admin@logisticshq.io", "Approved to continue")
	require.NoError(t, err)

	savedWf, err = repo.GetWorkflow(ctx, orgID, "wf-ctrl-555")
	require.NoError(t, err)
	assert.Equal(t, StateRunning, savedWf.CurrentState)

	// Action 3: CANCEL
	err = svc.PerformGovernedControlAction(ctx, orgID, "wf-ctrl-555", "CANCEL", "admin@logisticshq.io", "Superseded by manual billing adjustment")
	require.NoError(t, err)

	savedWf, err = repo.GetWorkflow(ctx, orgID, "wf-ctrl-555")
	require.NoError(t, err)
	assert.Equal(t, StateCancelled, savedWf.CurrentState)

	// Action 4: Invalid Action
	err = svc.PerformGovernedControlAction(ctx, orgID, "wf-ctrl-555", "DESTROY_ALL", "admin@logisticshq.io", "invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported control action")
}
