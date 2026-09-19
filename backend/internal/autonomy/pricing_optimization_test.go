package autonomy_test

import (
	"context"
	"testing"

	"github.com/freel/backend/internal/autonomy"
	"github.com/stretchr/testify/assert"
)

func TestPricingOptimization_EvaluationAndStrategyGeneration(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	res, err := svc.EvaluateRfqPricing(context.Background(), 2, nil, 101)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Optimization)
	assert.Equal(t, int64(101), res.Optimization.RfqID)
	assert.Equal(t, "USD", res.Optimization.Currency)
	assert.Equal(t, "strat-competitive-std", res.Optimization.RecommendedStrategyID)
	assert.Equal(t, 16.0, res.Optimization.RecommendedMarginPct)
	assert.Equal(t, 3214.29, res.Optimization.RecommendedPrice)

	// Verify candidate strategies
	assert.GreaterOrEqual(t, len(res.Candidates), 2)

	// Verify facts and predictions separated
	assert.NotEmpty(t, res.ActualFacts)
	assert.NotEmpty(t, res.Predictions)
	assert.NotEmpty(t, res.Assumptions)

	// Verify 7-step plan generated
	assert.NotNil(t, res.Plan)
	assert.Equal(t, 7, len(res.PlanSteps))
}

func TestPricingOptimization_HardConstraintViolation(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	// Evaluate RFQ first
	evalRes, err := svc.EvaluateRfqPricing(context.Background(), 2, nil, 101)
	assert.NoError(t, err)
	assert.NotNil(t, evalRes)

	// Mock repository stores optimization
	mockRepoOpt := evalRes.Optimization

	// Override GetRfqPricingOptimization to return this optimization
	// Attempt to select the infeasible candidate (margin 3.57% < floor 8.0%)
	// Note: in mock, we can test SelectPricingStrategy with the infeasible strategy
	// We verify that ErrCandidateInfeasible is returned
	repoPricing := &MockPricingAutonomyRepo{
		MockAutonomyRepo: repo,
		opt:              mockRepoOpt,
	}
	pricingSvc := autonomy.NewService(repoPricing, sidecar, nil, nil, nil)

	_, err = pricingSvc.SelectPricingStrategy(context.Background(), 2, 1000, 101, "strat-below-floor-infeasible")
	assert.Error(t, err)
	assert.Equal(t, autonomy.ErrCandidateInfeasible, err)
}

func TestPricingOptimization_ApprovalRequiredBeforeExecution(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	evalRes, err := svc.EvaluateRfqPricing(context.Background(), 2, nil, 101)
	assert.NoError(t, err)

	// Set optimization to require approval and approval status = PENDING
	evalRes.Optimization.RequiresApproval = true
	evalRes.Optimization.ApprovalStatus = "PENDING"

	repoPricing := &MockPricingAutonomyRepo{
		MockAutonomyRepo: repo,
		opt:              evalRes.Optimization,
	}
	pricingSvc := autonomy.NewService(repoPricing, sidecar, nil, nil, nil)

	// Attempt to execute quotation without approval -> must return ErrApprovalRequired
	_, err = pricingSvc.ExecutePricingQuotation(context.Background(), 2, 1000, 101)
	assert.Error(t, err)
	assert.Equal(t, autonomy.ErrApprovalRequired, err)
}

func TestPricingOptimization_Replanning(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	evalRes, err := svc.EvaluateRfqPricing(context.Background(), 2, nil, 101)
	assert.NoError(t, err)

	repoPricing := &MockPricingAutonomyRepo{
		MockAutonomyRepo: repo,
		opt:              evalRes.Optimization,
	}
	pricingSvc := autonomy.NewService(repoPricing, sidecar, nil, nil, nil)

	replanRes, err := pricingSvc.ReplanRfqPricing(
		context.Background(), 2, nil, 101,
		"Carrier increased bunker fuel surcharge by $200", 200.0,
	)
	assert.NoError(t, err)
	assert.NotNil(t, replanRes)
	assert.Equal(t, 2, replanRes.Optimization.CurrentVersion)
}

type MockPricingAutonomyRepo struct {
	*MockAutonomyRepo
	opt *autonomy.RfqPricingOptimization
}

func (m *MockPricingAutonomyRepo) GetRfqPricingOptimization(ctx context.Context, orgID, rfqID int64) (*autonomy.RfqPricingOptimization, error) {
	if m.opt != nil && m.opt.RfqID == rfqID && m.opt.OrgID == orgID {
		return m.opt, nil
	}
	return nil, nil
}

func (m *MockPricingAutonomyRepo) UpdateRfqPricingOptimizationStrategy(ctx context.Context, orgID, rfqID int64, strategyID string, price, marginPct float64, requiresApproval bool, approvalReason *string) error {
	if m.opt != nil {
		m.opt.RecommendedStrategyID = strategyID
		m.opt.RecommendedPrice = price
		m.opt.RecommendedMarginPct = marginPct
		m.opt.RequiresApproval = requiresApproval
		m.opt.ApprovalReason = approvalReason
	}
	return nil
}
