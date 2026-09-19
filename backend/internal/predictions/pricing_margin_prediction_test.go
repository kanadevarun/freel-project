package predictions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRFQMarginIntelligence_WorkflowAndGovernance(t *testing.T) {
	repo := newMockRepo()

	predCategory := "RFQ_MARGIN_RISK"
	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-price-rfq-101-test",
			OrgID:               2,
			Module:              "pricing",
			PredictionType:      "RFQ_MARGIN_RISK",
			PredictionCategory:  &predCategory,
			RelatedRecordType:   "RFQ",
			RelatedRecordID:     "101",
			PredictionStatement: "Margin Compression Risk: Projected gross margin of 7.5% is below target threshold.",
			PredictedValue:      func(s string) *string { return &s }("MARGIN_COMPRESSION"),
			TimeHorizon:         func(s string) *string { return &s }("14_DAYS"),
			Severity:            "HIGH",
			ConfidenceScore:     0.91,
			ConfidenceBand:      "HIGH",
			Explanation:         "Lane Nhava Sheva to Rotterdam exhibits volatile carrier surcharges.",
			SupportingSignals: []SupportingSignal{
				{SignalName: "margin_percentage", ObservedValue: "7.5%", BaselineValue: "15.0%", ImportanceWeight: 0.95},
			},
			SourceReferences: []SourceReference{
				{SourceModule: "rfqs", SourceRecordID: "101", SourceField: "rfq_quotes.sell_price", SourceTimestamp: time.Now().Format(time.RFC3339)},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Apply 4-6% surcharge contingency markup to safeguard profitability."),
			ActionType:        func(s string) *string { return &s }("rfqs.request_margin_adjustment"),
			IsActionRequired:  true,
			RequiresApproval:  true,
		},
	}

	svc := NewService(repo, sidecar)
	svcImpl := svc.(*service)
	svcImpl.pricingDataProvider = &PricingMarginDataProvider{}

	ctx := context.Background()

	// 1. Tenant access control validation
	_, err := svc.GetOrPredictRFQMarginIntelligence(ctx, 0, nil, 101, false)
	assert.ErrorIs(t, err, ErrInvalidOrgID)

	// 2. Invalid RFQ ID validation
	_, err = svc.GetOrPredictRFQMarginIntelligence(ctx, 2, nil, 0, false)
	assert.Error(t, err)

	// 3. Pre-seed active prediction in mock repo
	prePred := &Prediction{
		OrgID:               2,
		PredictionID:        "pred-price-rfq-101-initial",
		IdempotencyKey:      "org:2:rfq_margin:101:cycle_initial",
		Module:              "pricing",
		PredictionType:      "RFQ_MARGIN_RISK",
		Status:              StatusPublished,
		Severity:            SeverityHigh,
		ConfidenceScore:     0.90,
		ConfidenceBand:      ConfidenceHigh,
		RelatedRecordType:   "RFQ",
		RelatedRecordID:     "101",
		PredictionStatement: "Initial RFQ margin prediction",
		Explanation:         "Historical compression detected",
		SourceReferences: []SourceReference{
			{SourceModule: "rfqs", SourceRecordID: "101", SourceField: "rfq_quotes.sell_price", SourceTimestamp: time.Now().Format(time.RFC3339)},
		},
		SourceTimestamp: time.Now(),
		ReviewStatus:    "PENDING",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err = repo.Create(ctx, prePred)
	require.NoError(t, err)

	// Verify cached prediction returned when forceRefresh is false
	cached, err := svc.GetOrPredictRFQMarginIntelligence(ctx, 2, nil, 101, false)
	require.NoError(t, err)
	assert.Equal(t, "pred-price-rfq-101-initial", cached.PredictionID)
	assert.Equal(t, "RFQ_MARGIN_RISK", string(cached.PredictionType))
}

func TestContractRatePressure_WorkflowAndGovernance(t *testing.T) {
	repo := newMockRepo()

	predCategory := "CONTRACT_RATE_PRESSURE"
	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-ctr-103-test",
			OrgID:               1,
			Module:              "contracts",
			PredictionType:      "CONTRACT_RATE_PRESSURE",
			PredictionCategory:  &predCategory,
			RelatedRecordType:   "CONTRACT",
			RelatedRecordID:     "103",
			PredictionStatement: "Imminent Tariff Expiry: Agreement with Apex Drayage expires in 15 days.",
			PredictedValue:      func(s string) *string { return &s }("UPCOMING_RATE_PRESSURE"),
			TimeHorizon:         func(s string) *string { return &s }("30_DAYS"),
			Severity:            "HIGH",
			ConfidenceScore:     0.92,
			ConfidenceBand:      "HIGH",
			Explanation:         "Expiring contracted buying rates will elevate drayage costs.",
			SupportingSignals: []SupportingSignal{
				{SignalName: "days_until_expiry", ObservedValue: "15", BaselineValue: "90+", ImportanceWeight: 0.92},
			},
			SourceReferences: []SourceReference{
				{SourceModule: "contracts", SourceRecordID: "103", SourceField: "contracts.expiry_date", SourceTimestamp: time.Now().Format(time.RFC3339)},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Initiate proactive contract extension negotiations."),
			ActionType:        func(s string) *string { return &s }("contracts.review_renewal_terms"),
			IsActionRequired:  true,
			RequiresApproval:  true,
		},
	}

	svc := NewService(repo, sidecar)
	svcImpl := svc.(*service)
	svcImpl.pricingDataProvider = &PricingMarginDataProvider{}

	ctx := context.Background()

	// 1. Tenant access control validation
	_, err := svc.GetOrPredictContractRatePressure(ctx, 0, nil, 103, false)
	assert.ErrorIs(t, err, ErrInvalidOrgID)

	// 2. Invalid Contract ID validation
	_, err = svc.GetOrPredictContractRatePressure(ctx, 1, nil, 0, false)
	assert.Error(t, err)

	// 3. Pre-seed active prediction
	prePred := &Prediction{
		OrgID:               1,
		PredictionID:        "pred-ctr-103-initial",
		IdempotencyKey:      "org:1:contract_rate:103:cycle_initial",
		Module:              "contracts",
		PredictionType:      "CONTRACT_RATE_PRESSURE",
		Status:              StatusPublished,
		Severity:            SeverityHigh,
		ConfidenceScore:     0.91,
		ConfidenceBand:      ConfidenceHigh,
		RelatedRecordType:   "CONTRACT",
		RelatedRecordID:     "103",
		PredictionStatement: "Initial contract rate pressure",
		Explanation:         "Impending expiry in 15 days",
		SourceReferences: []SourceReference{
			{SourceModule: "contracts", SourceRecordID: "103", SourceField: "contracts.expiry_date", SourceTimestamp: time.Now().Format(time.RFC3339)},
		},
		SourceTimestamp: time.Now(),
		ReviewStatus:    "PENDING",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err = repo.Create(ctx, prePred)
	require.NoError(t, err)

	cached, err := svc.GetOrPredictContractRatePressure(ctx, 1, nil, 103, false)
	require.NoError(t, err)
	assert.Equal(t, "pred-ctr-103-initial", cached.PredictionID)
	assert.Equal(t, "CONTRACT_RATE_PRESSURE", string(cached.PredictionType))
}
