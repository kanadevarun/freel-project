package autonomy_test

import (
	"context"
	"testing"

	"github.com/freel/backend/internal/autonomy"
	"github.com/stretchr/testify/assert"
)

func TestFinanceCollections_EvaluationAndStrategyGeneration(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	// Evaluate overdue invoice 103 ($4,500 overdue 27 days)
	res, err := svc.EvaluateFinanceCollection(context.Background(), 2, nil, 103)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Plan)
	assert.Equal(t, int64(103), res.InvoiceID)
	assert.Equal(t, "USD", res.Currency)
	assert.Equal(t, 4500.0, res.BalanceDue)
	assert.Equal(t, 27, res.DaysOverdue)
	assert.Equal(t, "16-30_DAYS", res.AgingBucket)

	// Strategy verification
	assert.GreaterOrEqual(t, len(res.Candidates), 3)
	assert.NotEmpty(t, res.Plan.SelectedStrategyID)

	// Segregation of facts, predictions, and assumptions
	assert.NotEmpty(t, res.ActualFacts)
	assert.NotEmpty(t, res.Predictions)
	assert.NotEmpty(t, res.Assumptions)

	// 7-step autonomous execution plan
	assert.Equal(t, 7, len(res.PlanSteps))
}

func TestFinanceCollections_DisputedInvoiceRequiresApproval(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	// Evaluate disputed invoice 104
	res, err := svc.EvaluateFinanceCollection(context.Background(), 2, nil, 104)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.True(t, res.Plan.RequiresApproval, "Disputed invoice must require human-in-the-loop approval")
	assert.NotNil(t, res.Plan.ApprovalReason)
	assert.Contains(t, *res.Plan.ApprovalReason, "dispute")
}

func TestFinanceCollections_SettledInvoiceStopCondition(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	// Evaluate settled invoice 101 (balance = 0.0, status = Paid)
	res, err := svc.EvaluateFinanceCollection(context.Background(), 2, nil, 101)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 0.0, res.BalanceDue)
	assert.Equal(t, "STOPPED", res.Plan.Status)
	assert.NotNil(t, res.Plan.StopReason)

	// Executing action on settled invoice must be prohibited
	_, err = svc.ExecuteFinanceCollectionAction(context.Background(), 2, nil, 101)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "settled or collection is stopped")
}

func TestFinanceCollections_StrategySelectionAndActionExecution(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	// 1. Evaluate invoice
	evalRes, err := svc.EvaluateFinanceCollection(context.Background(), 2, nil, 103)
	assert.NoError(t, err)
	assert.NotNil(t, evalRes)

	// 2. Select different candidate strategy
	stateRes, err := svc.SelectFinanceCollectionStrategy(context.Background(), 2, nil, 103, "strat-friendly-reminder")
	assert.NoError(t, err)
	assert.Equal(t, "strat-friendly-reminder", stateRes.Plan.SelectedStrategyID)

	// 3. Execute collection action via Action System boundary
	execRes, err := svc.ExecuteFinanceCollectionAction(context.Background(), 2, nil, 103)
	assert.NoError(t, err)
	assert.NotNil(t, execRes)
	assert.True(t, execRes.Success)

	// 4. Verify updated state
	updatedState, err := svc.GetFinanceCollectionState(context.Background(), 2, 103)
	assert.NoError(t, err)
	assert.Equal(t, "EXECUTED", updatedState.Plan.Status)
}

func TestFinanceCollections_ReplanningAndVersioning(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	// Initial evaluation
	_, err := svc.EvaluateFinanceCollection(context.Background(), 2, nil, 103)
	assert.NoError(t, err)

	// Replan upon payment received event
	replanPayload := map[string]interface{}{
		"amount_paid": 2000.0,
		"payment_ref": "WIRE-2026-9912",
	}
	replanRes, err := svc.ReplanFinanceCollection(context.Background(), 2, nil, 103, "PARTIAL_PAYMENT", replanPayload)
	assert.NoError(t, err)
	assert.NotNil(t, replanRes)
	assert.Equal(t, 2, replanRes.Plan.Version, "Plan version must increment upon replanning event")
	assert.Equal(t, 2500.0, replanRes.BalanceDue, "Outstanding balance must reflect partial payment")
}
