package autonomy_test

import (
	"context"
	"testing"

	"github.com/freel/backend/internal/autonomy"
	"github.com/stretchr/testify/assert"
)

func TestContractCompliance_EvaluationAndStrategyGeneration(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	// Evaluate compliant contract 201
	res, err := svc.EvaluateContractCompliance(context.Background(), 2, nil, 201)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Plan)
	assert.Equal(t, int64(201), res.ContractID)
	assert.Equal(t, "CTR-2026-DEV-201", res.ContractReference)
	assert.Equal(t, "COMPLIANT", res.ComplianceStatus)
	assert.Equal(t, "CURRENT", res.ExpirationStatus)

	// Strategy verification
	assert.NotEmpty(t, res.Candidates)
	assert.Equal(t, "strat-compliance-ok-monitor", res.Plan.SelectedRemediationStrategyID)

	// Segregation of facts, predictions, and assumptions
	assert.NotEmpty(t, res.AuthoritativeFacts)
	assert.NotEmpty(t, res.Predictions)
	assert.NotEmpty(t, res.Assumptions)

	// 7-step autonomous execution plan
	assert.Equal(t, 7, len(res.RemediationPlanSteps))
}

func TestContractCompliance_ExpiringContractRequiresApproval(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	// Evaluate expiring contract 202 (16 days remaining)
	res, err := svc.EvaluateContractCompliance(context.Background(), 2, nil, 202)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "EXPIRING_SOON", res.ExpirationStatus)
	assert.True(t, res.Plan.RequiresApproval, "Expiring contract renewal prep must require human approval")
	assert.NotNil(t, res.Plan.ApprovalReason)
	assert.Contains(t, *res.Plan.ApprovalReason, "expiration threshold")
	assert.Equal(t, "strat-expiring-renewal-prep", res.Plan.SelectedRemediationStrategyID)
}

func TestContractCompliance_MissingStatutoryDocument(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	// Evaluate contract 203 (missing mandatory GDP Pharma Certificate)
	res, err := svc.EvaluateContractCompliance(context.Background(), 2, nil, 203)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "NON_COMPLIANT", res.ComplianceStatus)
	assert.Equal(t, "strat-missing-document-remediation", res.Plan.SelectedRemediationStrategyID)
	var chosenCandidate *autonomy.ComplianceRemediationStrategyCandidateDTO
	for _, c := range res.Candidates {
		if c.StrategyID == res.Plan.SelectedRemediationStrategyID {
			chosenCandidate = &c
			break
		}
	}
	assert.NotNil(t, chosenCandidate)
	assert.Equal(t, "REQUEST_STATUTORY_DOCUMENT", chosenCandidate.RecommendedAction)
}

func TestContractCompliance_ExpiredContractStopCondition(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	// Evaluate expired contract 204
	res, err := svc.EvaluateContractCompliance(context.Background(), 2, nil, 204)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "EXPIRED", res.ExpirationStatus)
	assert.Equal(t, "STOPPED", res.Plan.ExecutionStatus)
	assert.NotNil(t, res.Plan.StopReason)
	assert.Contains(t, *res.Plan.StopReason, "expired")

	// Executing action on expired contract must be prohibited by Go authoritative guardrails
	_, err = svc.ExecuteContractComplianceAction(context.Background(), 2, nil, 204)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestContractCompliance_ReplanningAndVersioning(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	// 1. Initial evaluation of non-compliant contract 203
	initRes, err := svc.EvaluateContractCompliance(context.Background(), 2, nil, 203)
	assert.NoError(t, err)
	assert.NotNil(t, initRes)
	assert.Equal(t, 1, initRes.Plan.Version)

	// 2. Select alternative strategy
	stateRes, err := svc.SelectContractComplianceStrategy(context.Background(), 2, nil, 203, "strat-compliance-ok-monitor")
	assert.NoError(t, err)
	assert.Equal(t, "strat-compliance-ok-monitor", stateRes.Plan.SelectedRemediationStrategyID)

	// 3. Replan upon DOCUMENT_VERIFIED event
	replanPayload := map[string]interface{}{
		"document_id":   901,
		"document_type": "GDP_PHARMA_CERT",
		"verified_by":   "compliance-officer-42",
	}
	replanRes, err := svc.ReplanContractCompliance(context.Background(), 2, nil, 203, "DOCUMENT_VERIFIED", replanPayload)
	assert.NoError(t, err)
	assert.NotNil(t, replanRes)
	assert.Equal(t, 2, replanRes.Plan.Version, "Plan version must increment upon replanning event")
	assert.Equal(t, "COMPLIANT", replanRes.ComplianceStatus, "Authoritative status transitioned to COMPLIANT")
}
