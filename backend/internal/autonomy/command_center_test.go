package autonomy_test

import (
	"context"
	"testing"

	"github.com/freel/backend/internal/autonomy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandCenter_Overview(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	orgID := int64(1)

	overview, err := svc.GetCommandCenterOverview(ctx, orgID)
	require.NoError(t, err)
	require.NotNil(t, overview)

	assert.Equal(t, 5, overview.ActiveShipments)
	assert.Equal(t, 2, overview.ShipmentsAtRisk)
	assert.Equal(t, 3, overview.ActiveExceptions)
	assert.Equal(t, 1, overview.CriticalExceptions)
	assert.Equal(t, 4, overview.ActiveWorkflows)
	assert.Equal(t, 1, overview.WorkflowsWaitingHuman)
	assert.Equal(t, 2, overview.PendingApprovals)
	assert.Equal(t, 1, overview.EscalationsCount)
	assert.Equal(t, 0, overview.FailedActionsCount)
	assert.Equal(t, 1, overview.StalledPlansCount)

	// Factual vs Predicted vs AI Analysis separation
	assert.NotEmpty(t, overview.ActualSummary)
	assert.NotEmpty(t, overview.PredictedSummary)
	assert.NotEmpty(t, overview.AIAnalysisSummary)
	assert.Contains(t, overview.ActualSummary, "5 active shipments")
	assert.Contains(t, overview.PredictedSummary, "2 shipments with predicted delay")
	assert.Equal(t, "HEALTHY", overview.SystemHealthStatus)
	assert.False(t, overview.IsStale)
}

func TestCommandCenter_CriticalAttention(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	orgID := int64(1)

	items, err := svc.GetCommandCenterCriticalAttention(ctx, orgID, 10)
	require.NoError(t, err)
	require.NotEmpty(t, items)

	first := items[0]
	assert.Equal(t, "item-crit-1", first.ID)
	assert.Equal(t, "CRITICAL", first.Severity)
	assert.Equal(t, "CRITICAL_SAFETY_COMPLIANCE", first.PriorityTier)
	assert.Equal(t, 1, first.PriorityRank)
	assert.GreaterOrEqual(t, first.PriorityScore, 90.0)
	assert.True(t, first.RequiresHuman)
	assert.NotEmpty(t, first.ActualFacts)
	assert.NotEmpty(t, first.PredictedImpact)
	assert.NotEmpty(t, first.RecommendedAction)
	assert.NotEmpty(t, first.WhyFlagged)
}

func TestCommandCenter_DomainRisks(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	orgID := int64(1)

	domains, err := svc.GetCommandCenterDomainRisks(ctx, orgID)
	require.NoError(t, err)
	require.NotEmpty(t, domains)

	shipmentDomain := domains[0]
	assert.Equal(t, "SHIPMENT", shipmentDomain.Domain)
	assert.Equal(t, 1, shipmentDomain.TotalAtRisk)
	assert.NotEmpty(t, shipmentDomain.AuthoritativeState)
	assert.NotEmpty(t, shipmentDomain.PredictedRisk)
	require.NotEmpty(t, shipmentDomain.Items)

	item := shipmentDomain.Items[0]
	assert.Equal(t, "101", item.EntityID)
	assert.Equal(t, "SHP-101", item.EntityReference)
	assert.NotEmpty(t, item.ActualFact)
	assert.NotEmpty(t, item.PredictedRisk)
	assert.NotEmpty(t, item.RecommendedAction)
}

func TestCommandCenter_SystemHealth(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	health, err := svc.GetCommandCenterSystemHealth(ctx)
	require.NoError(t, err)
	require.NotNil(t, health)

	assert.Equal(t, "HEALTHY", health.OverallStatus)
	assert.Contains(t, health.Subsystems, "go_backend")
	assert.Contains(t, health.Subsystems, "database")
	assert.Equal(t, "HEALTHY", health.Subsystems["go_backend"].Status)
}

func TestCommandCenter_Activity(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	orgID := int64(1)

	activity, err := svc.GetCommandCenterActivity(ctx, orgID, 10)
	require.NoError(t, err)
	require.NotNil(t, activity)
	assert.NotNil(t, activity.RecentActions)
	assert.NotNil(t, activity.ReplanningEvents)
	assert.NotNil(t, activity.Escalations)
}
