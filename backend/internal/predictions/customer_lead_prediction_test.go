package predictions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLeadIntelligence_WorkflowAndGovernance(t *testing.T) {
	repo := newMockRepo()

	predCategory := "CONVERSION_LIKELIHOOD"
	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-lead-101-test",
			OrgID:               2,
			Module:              "leads",
			PredictionType:      "LEAD_CONVERSION_LIKELIHOOD",
			PredictionCategory:  &predCategory,
			RelatedRecordType:   "LEAD",
			RelatedRecordID:     "101",
			PredictionStatement: "High Conversion Likelihood: Lead #101 demonstrated high engagement and completed conversion.",
			PredictedValue:      func(s string) *string { return &s }("HIGH"),
			TimeHorizon:         func(s string) *string { return &s }("14_DAYS"),
			Severity:            "LOW",
			ConfidenceScore:     0.94,
			ConfidenceBand:      "HIGH",
			Explanation:         "Lead has strong interaction history, won quotation records, and converted account linkage.",
			SupportingSignals: []SupportingSignal{
				{SignalName: "lead_status", ObservedValue: "CONVERTED", BaselineValue: "NEW", ImportanceWeight: 0.95},
				{SignalName: "won_quotes_count", ObservedValue: "1", BaselineValue: "0", ImportanceWeight: 0.90},
			},
			SourceReferences: []SourceReference{
				{SourceModule: "leads", SourceRecordID: "101", SourceField: "leads.status", SourceTimestamp: time.Now().Format(time.RFC3339)},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Maintain account onboarding and review repeat RFQ opportunities."),
			ActionType:        func(s string) *string { return &s }("leads.schedule_followup"),
			IsActionRequired:  true,
			RequiresApproval:  true,
		},
	}

	svc := NewService(repo, sidecar)
	svcImpl := svc.(*service)
	svcImpl.clDataProvider = &CustomerLeadDataProvider{}

	ctx := context.Background()

	// 1. Tenant access control validation
	_, err := svc.GetOrPredictLeadIntelligence(ctx, 0, nil, 101, false)
	assert.ErrorIs(t, err, ErrInvalidOrgID)

	// 2. Invalid lead ID validation
	_, err = svc.GetOrPredictLeadIntelligence(ctx, 2, nil, 0, false)
	assert.Error(t, err)

	// 3. Pre-seed active prediction in mock repo
	prePred := &Prediction{
		OrgID:               2,
		PredictionID:        "pred-lead-101-initial",
		IdempotencyKey:      "org:2:lead_intelligence:101:cycle_initial",
		Module:              "leads",
		PredictionType:      "LEAD_CONVERSION_LIKELIHOOD",
		Status:              StatusPublished,
		Severity:            SeverityLow,
		ConfidenceScore:     0.92,
		ConfidenceBand:      ConfidenceHigh,
		RelatedRecordType:   "LEAD",
		RelatedRecordID:     "101",
		PredictionStatement: "Initial conversion prediction",
		Explanation:         "High responsiveness detected",
		SourceReferences: []SourceReference{
			{SourceModule: "leads", SourceRecordID: "101", SourceField: "status", SourceTimestamp: time.Now().Format(time.RFC3339)},
		},
		SourceTimestamp: time.Now(),
		ReviewStatus:    "PENDING",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err = repo.Create(ctx, prePred)
	require.NoError(t, err)

	// Verify cached prediction returned when forceRefresh is false
	cached, err := svc.GetOrPredictLeadIntelligence(ctx, 2, nil, 101, false)
	require.NoError(t, err)
	assert.Equal(t, "pred-lead-101-initial", cached.PredictionID)
	assert.Equal(t, "LEAD_CONVERSION_LIKELIHOOD", string(cached.PredictionType))
}

func TestCustomerIntelligence_WorkflowAndGovernance(t *testing.T) {
	repo := newMockRepo()

	predCategory := "REPEAT_BUSINESS_LIKELIHOOD"
	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-cust-101-test",
			OrgID:               2,
			Module:              "leads",
			PredictionType:      "CUSTOMER_REPEAT_BUSINESS",
			PredictionCategory:  &predCategory,
			RelatedRecordType:   "CUSTOMER",
			RelatedRecordID:     "101",
			PredictionStatement: "High Repeat Business Momentum: Customer #101 shows strong cadence of bookings.",
			PredictedValue:      func(s string) *string { return &s }("HIGH"),
			TimeHorizon:         func(s string) *string { return &s }("30_DAYS"),
			Severity:            "LOW",
			ConfidenceScore:     0.88,
			ConfidenceBand:      "HIGH",
			Explanation:         "Consistent RFQs and high shipment throughput indicate robust repeat business.",
			SupportingSignals: []SupportingSignal{
				{SignalName: "health_score", ObservedValue: "92", BaselineValue: "50", ImportanceWeight: 0.85},
			},
			SourceReferences: []SourceReference{
				{SourceModule: "customers", SourceRecordID: "101", SourceField: "customers.health_score", SourceTimestamp: time.Now().Format(time.RFC3339)},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Proactively share preferred lane capacity and quarterly schedule."),
			ActionType:        func(s string) *string { return &s }("customers.send_lane_rates"),
			IsActionRequired:  false,
			RequiresApproval:  true,
		},
	}

	svc := NewService(repo, sidecar)
	svcImpl := svc.(*service)
	svcImpl.clDataProvider = &CustomerLeadDataProvider{}

	ctx := context.Background()

	// 1. Tenant access control validation
	_, err := svc.GetOrPredictCustomerIntelligence(ctx, 0, nil, 101, false)
	assert.ErrorIs(t, err, ErrInvalidOrgID)

	// 2. Invalid customer ID validation
	_, err = svc.GetOrPredictCustomerIntelligence(ctx, 2, nil, 0, false)
	assert.Error(t, err)

	// 3. Pre-seed active prediction
	prePred := &Prediction{
		OrgID:               2,
		PredictionID:        "pred-cust-101-initial",
		IdempotencyKey:      "org:2:customer_intelligence:101:cycle_initial",
		Module:              "customers",
		PredictionType:      "CUSTOMER_REPEAT_BUSINESS",
		Status:              StatusPublished,
		Severity:            SeverityLow,
		ConfidenceScore:     0.85,
		ConfidenceBand:      ConfidenceHigh,
		RelatedRecordType:   "CUSTOMER",
		RelatedRecordID:     "101",
		PredictionStatement: "Initial repeat business momentum",
		Explanation:         "Active bookings observed",
		SourceReferences: []SourceReference{
			{SourceModule: "customers", SourceRecordID: "101", SourceField: "health_score", SourceTimestamp: time.Now().Format(time.RFC3339)},
		},
		SourceTimestamp: time.Now(),
		ReviewStatus:    "PENDING",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err = repo.Create(ctx, prePred)
	require.NoError(t, err)

	cached, err := svc.GetOrPredictCustomerIntelligence(ctx, 2, nil, 101, false)
	require.NoError(t, err)
	assert.Equal(t, "pred-cust-101-initial", cached.PredictionID)
	assert.Equal(t, "CUSTOMER_REPEAT_BUSINESS", string(cached.PredictionType))
}
