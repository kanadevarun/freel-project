package predictions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContractComplianceRisk_WorkflowAndGovernance(t *testing.T) {
	repo := newMockRepo()

	predCategory := "CONTRACT_CLAUSE_COMMERCIAL_RISK"
	docRef := "contract_documents:101"
	clauseRef := "CLAUSE-LIA-02"
	pageNo := 6
	secHeading := "Section 8 - Carrier Liability & Indemnity"

	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-comp-ctr-101-test",
			OrgID:               1,
			Module:              "contracts",
			PredictionType:      "CONTRACT_CLAUSE_COMMERCIAL_RISK",
			PredictionCategory:  &predCategory,
			RelatedRecordType:   "CONTRACT",
			RelatedRecordID:     "101",
			PredictionStatement: "Commercial Clause Risk: Liability clause mismatch with required forwarder indemnification standards.",
			PredictedValue:      func(s string) *string { return &s }("CLAUSE_DISCREPANCY"),
			TimeHorizon:         func(s string) *string { return &s }("PRIOR_TO_ACTIVATION"),
			Severity:            "HIGH",
			ConfidenceScore:     0.94,
			ConfidenceBand:      "HIGH",
			Explanation:         "Clause CLAUSE-LIA-02 restricts carrier cargo liability to $2/kg, conflicting with customer SLA baseline.",
			DocumentReference:   &docRef,
			ClauseReference:     &clauseRef,
			PageNumber:          &pageNo,
			SectionHeading:      &secHeading,
			SupportingSignals: []SupportingSignal{
				{SignalName: "discrepancy_count", ObservedValue: "1", BaselineValue: "0", ImportanceWeight: 0.95},
				{SignalName: "missing_document_count", ObservedValue: "3", BaselineValue: "0", ImportanceWeight: 0.88},
			},
			SourceReferences: []SourceReference{
				{SourceModule: "contracts", SourceRecordID: "101", SourceField: "contracts.id", SourceTimestamp: time.Now().Format(time.RFC3339)},
				{SourceModule: "contract_documents", SourceRecordID: "101", SourceField: "contract_documents.status", SourceTimestamp: time.Now().Format(time.RFC3339)},
				{SourceModule: "ai_contract_compliance_reviews", SourceRecordID: "1", SourceField: "ai_contract_compliance_reviews.clause_analysis", SourceTimestamp: time.Now().Format(time.RFC3339)},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Initiate legal review of Clause CLAUSE-LIA-02 before executing master agreement."),
			ActionType:        func(s string) *string { return &s }("contracts.request_compliance_review"),
			IsActionRequired:  true,
			RequiresApproval:  true,
		},
	}

	svc := NewService(repo, sidecar)
	svcImpl := svc.(*service)
	svcImpl.contractComplianceData = &ContractComplianceDataProvider{}

	ctx := context.Background()

	// 1. Tenant access control validation
	_, err := svc.GetOrPredictContractComplianceRisk(ctx, 0, nil, 101, false)
	assert.ErrorIs(t, err, ErrInvalidOrgID)

	// 2. Invalid Contract ID validation
	_, err = svc.GetOrPredictContractComplianceRisk(ctx, 1, nil, 0, false)
	assert.Error(t, err)

	// 3. Pre-seed active prediction in mock repo
	prePred := &Prediction{
		OrgID:               1,
		PredictionID:        "pred-comp-ctr-101-initial",
		IdempotencyKey:      "pred:contracts:comp:1:101:initial",
		Module:              "contracts",
		PredictionType:      "CONTRACT_CLAUSE_COMMERCIAL_RISK",
		Status:              StatusPublished,
		Severity:            SeverityHigh,
		ConfidenceScore:     0.94,
		ConfidenceBand:      ConfidenceHigh,
		RelatedRecordType:   "CONTRACT",
		RelatedRecordID:     "101",
		PredictionStatement: "Initial clause compliance risk",
		Explanation:         "Initial liability gap",
		DocumentReference:   &docRef,
		ClauseReference:     &clauseRef,
		PageNumber:          &pageNo,
		SectionHeading:      &secHeading,
		SourceReferences: []SourceReference{
			{SourceModule: "contracts", SourceRecordID: "101", SourceField: "contracts.id", SourceTimestamp: time.Now().Format(time.RFC3339)},
		},
		SourceTimestamp:  time.Now(),
		ReviewStatus:     "PENDING",
		IsActionRequired: true,
		RequiresApproval: true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	err = repo.Create(ctx, prePred)
	require.NoError(t, err)

	// Verify cached prediction returned when forceRefresh is false
	cached, err := svc.GetOrPredictContractComplianceRisk(ctx, 1, nil, 101, false)
	require.NoError(t, err)
	assert.Equal(t, "pred-comp-ctr-101-initial", cached.PredictionID)
	assert.Equal(t, "CONTRACT_CLAUSE_COMMERCIAL_RISK", string(cached.PredictionType))
	assert.Equal(t, &clauseRef, cached.ClauseReference)
	assert.Equal(t, &pageNo, cached.PageNumber)
	assert.True(t, cached.RequiresApproval)
}

func TestContractExpiryRisk_WorkflowAndGovernance(t *testing.T) {
	repo := newMockRepo()

	predCategory := "CONTRACT_EXPIRY_RENEWAL_RISK"
	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-comp-ctr-104-expired",
			OrgID:               1,
			Module:              "contracts",
			PredictionType:      "CONTRACT_EXPIRY_RENEWAL_RISK",
			PredictionCategory:  &predCategory,
			RelatedRecordType:   "CONTRACT",
			RelatedRecordID:     "104",
			PredictionStatement: "Expired Contract Risk: Agreement AIR-CH-2024-EUR expired on 2025-05-31.",
			PredictedValue:      func(s string) *string { return &s }("CONTRACT_EXPIRED"),
			TimeHorizon:         func(s string) *string { return &s }("IMMEDIATE"),
			Severity:            "CRITICAL",
			ConfidenceScore:     0.99,
			ConfidenceBand:      "HIGH",
			Explanation:         "Contract expired 467 days ago. Quotations and shipments cannot bind expired tariff schedules.",
			SourceReferences: []SourceReference{
				{SourceModule: "contracts", SourceRecordID: "104", SourceField: "contracts.expiry_date", SourceTimestamp: time.Now().Format(time.RFC3339)},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Draft renewal agreement or decommission carrier tariff schedule."),
			ActionType:        func(s string) *string { return &s }("contracts.initiate_renewal"),
			IsActionRequired:  true,
			RequiresApproval:  true,
		},
	}

	svc := NewService(repo, sidecar)
	ctx := context.Background()

	// Seed expired contract prediction
	pred := &Prediction{
		OrgID:               1,
		PredictionID:        "pred-comp-ctr-104-expired",
		IdempotencyKey:      "pred:contracts:comp:1:104:initial",
		Module:              "contracts",
		PredictionType:      "CONTRACT_EXPIRY_RENEWAL_RISK",
		Status:              StatusPublished,
		Severity:            SeverityCritical,
		ConfidenceScore:     0.99,
		ConfidenceBand:      ConfidenceHigh,
		RelatedRecordType:   "CONTRACT",
		RelatedRecordID:     "104",
		PredictionStatement: "Expired Contract Risk: Agreement AIR-CH-2024-EUR expired on 2025-05-31.",
		SourceReferences: []SourceReference{
			{SourceModule: "contracts", SourceRecordID: "104", SourceField: "contracts.expiry_date", SourceTimestamp: time.Now().Format(time.RFC3339)},
		},
		SourceTimestamp:  time.Now(),
		ReviewStatus:     "PENDING",
		IsActionRequired: true,
		RequiresApproval: true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	err := repo.Create(ctx, pred)
	require.NoError(t, err)

	res, err := svc.GetOrPredictContractComplianceRisk(ctx, 1, nil, 104, false)
	require.NoError(t, err)
	assert.Equal(t, SeverityCritical, res.Severity)
	assert.Equal(t, "CONTRACT_EXPIRY_RENEWAL_RISK", string(res.PredictionType))
	assert.True(t, res.RequiresApproval)
}
