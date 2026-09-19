package predictions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCarrierPerformance_WorkflowAndGovernance(t *testing.T) {
	repo := newMockRepo()

	predCategory := "CARRIER_PERFORMANCE_RISK"
	compPeriod := "LAST_90_DAYS"
	sampleSize := 14
	carrierRef := "CMDU"
	laneRef := "INNSA-USNYC"

	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-carrier-cmdu-test",
			OrgID:               2,
			Module:              "carriers",
			PredictionType:      "CARRIER_PERFORMANCE_RISK",
			PredictionCategory:  &predCategory,
			RelatedRecordType:   "CARRIER",
			RelatedRecordID:     "CMDU",
			PredictionStatement: "Carrier Regulatory Hold Risk: CMA CGM (CMDU) exhibits elevated operational risk on corridor INNSA-USNYC due to active customs hold.",
			PredictedValue:      func(s string) *string { return &s }("CUSTOMS_HOLD_RISK"),
			TimeHorizon:         func(s string) *string { return &s }("14_DAYS"),
			Severity:            "HIGH",
			ConfidenceScore:     0.93,
			ConfidenceBand:      "HIGH",
			Explanation:         "Active customs hold on shipment #103 at INNSA. Historical transshipment dwell poses acute schedule risk.",
			SupportingSignals: []SupportingSignal{
				{SignalName: "customs_holds_count", ObservedValue: "1", BaselineValue: "0", ImportanceWeight: 0.96},
				{SignalName: "on_time_reliability", ObservedValue: "78.5%", BaselineValue: ">= 90.0%", ImportanceWeight: 0.90},
			},
			SourceReferences: []SourceReference{
				{
					SourceModule:     "shipments",
					SourceRecordID:   "103",
					SourceField:      "shipments.status",
					SourceTimestamp:  time.Now().Format(time.RFC3339),
					CarrierReference: carrierRef,
					LaneReference:    laneRef,
					ComparisonPeriod: compPeriod,
					SampleSize:       sampleSize,
				},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Review carrier operational hold with compliance team."),
			ActionType:        func(s string) *string { return &s }("carriers.review_performance"),
			IsActionRequired:  true,
			RequiresApproval:  true,
			ComparisonPeriod:  &compPeriod,
			SampleSize:        &sampleSize,
			CarrierReference:  &carrierRef,
			LaneReference:     &laneRef,
		},
	}

	svc := NewService(repo, sidecar)
	svcImpl := svc.(*service)
	svcImpl.networkPerfData = &NetworkPerformanceDataProvider{}

	ctx := context.Background()

	// 1. Tenant access control validation
	_, err := svc.GetOrPredictCarrierPerformance(ctx, 0, nil, "CMDU", false)
	assert.ErrorIs(t, err, ErrInvalidOrgID)

	// 2. Missing SCAC validation
	_, err = svc.GetOrPredictCarrierPerformance(ctx, 2, nil, "", false)
	assert.ErrorIs(t, err, ErrMissingRecordID)

	// 3. Pre-seed active prediction in mock repo
	prePred := &Prediction{
		OrgID:               2,
		PredictionID:        "pred-carrier-cmdu-initial",
		IdempotencyKey:      "pred-carrier-2-CMDU-CARRIER_PERFORMANCE_RISK",
		Module:              "carriers",
		PredictionType:      "CARRIER_PERFORMANCE_RISK",
		Status:              StatusPublished,
		Severity:            SeverityHigh,
		ConfidenceScore:     0.93,
		ConfidenceBand:      ConfidenceHigh,
		RelatedRecordType:   "CARRIER",
		RelatedRecordID:     "CMDU",
		PredictionStatement: "Carrier Regulatory Hold Risk: CMA CGM (CMDU)",
		Explanation:         "Active customs hold",
		SourceReferences: []SourceReference{
			{SourceModule: "shipments", SourceRecordID: "103", SourceField: "status", SourceTimestamp: time.Now().Format(time.RFC3339)},
		},
		SourceTimestamp:  time.Now(),
		ReviewStatus:     "PENDING",
		ComparisonPeriod: &compPeriod,
		SampleSize:       &sampleSize,
		CarrierReference: &carrierRef,
		LaneReference:    &laneRef,
		IsActionRequired: true,
		RequiresApproval: true,
	}
	err = repo.Create(ctx, prePred)
	require.NoError(t, err)

	// Fetch cached prediction
	cached, err := svc.GetOrPredictCarrierPerformance(ctx, 2, nil, "CMDU", false)
	require.NoError(t, err)
	assert.Equal(t, "pred-carrier-cmdu-initial", cached.PredictionID)
	assert.Equal(t, "CMDU", *cached.CarrierReference)
	assert.Equal(t, 14, *cached.SampleSize)

	// Force refresh should supersede
	refreshed, err := svc.GetOrPredictCarrierPerformance(ctx, 2, nil, "CMDU", true)
	require.NoError(t, err)
	assert.Equal(t, "pred-carrier-cmdu-test", refreshed.PredictionID)
	assert.True(t, refreshed.RequiresApproval)
	assert.True(t, refreshed.IsActionRequired)
}

func TestLanePerformance_WorkflowAndGovernance(t *testing.T) {
	repo := newMockRepo()

	predCategory := "LANE_PERFORMANCE_RISK"
	compPeriod := "LAST_90_DAYS"
	sampleSize := 18
	laneRef := "INNSA-USNYC"
	carrierRef := "CMDU"

	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-lane-usnyc-test",
			OrgID:               2,
			Module:              "network",
			PredictionType:      "LANE_PERFORMANCE_RISK",
			PredictionCategory:  &predCategory,
			RelatedRecordType:   "LANE",
			RelatedRecordID:     "INNSA-USNYC",
			PredictionStatement: "Trade Corridor Risk: Corridor INNSA-USNYC shows elevated regulatory inspection exposure.",
			PredictedValue:      func(s string) *string { return &s }("CUSTOMS_DWELL_RISK"),
			TimeHorizon:         func(s string) *string { return &s }("14_DAYS"),
			Severity:            "HIGH",
			ConfidenceScore:     0.92,
			ConfidenceBand:      "HIGH",
			Explanation:         "Shipment #103 on lane INNSA-USNYC is actively held under customs inspection.",
			SupportingSignals: []SupportingSignal{
				{SignalName: "active_customs_hold", ObservedValue: "1 ACTIVE HOLD", BaselineValue: "0", ImportanceWeight: 0.95},
			},
			SourceReferences: []SourceReference{
				{
					SourceModule:     "shipments",
					SourceRecordID:   "103",
					SourceField:      "shipments.destination_port",
					SourceTimestamp:  time.Now().Format(time.RFC3339),
					LaneReference:    laneRef,
					CarrierReference: carrierRef,
					ComparisonPeriod: compPeriod,
					SampleSize:       sampleSize,
				},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Request pre-clearance documentation audit."),
			ActionType:        func(s string) *string { return &s }("lanes.audit_preclearance"),
			IsActionRequired:  true,
			RequiresApproval:  true,
			ComparisonPeriod:  &compPeriod,
			SampleSize:        &sampleSize,
			LaneReference:     &laneRef,
			CarrierReference:  &carrierRef,
		},
	}

	svc := NewService(repo, sidecar)
	svcImpl := svc.(*service)
	svcImpl.networkPerfData = &NetworkPerformanceDataProvider{}

	ctx := context.Background()

	// 1. Tenant access control validation
	_, err := svc.GetOrPredictLanePerformance(ctx, 0, nil, "INNSA-USNYC", false)
	assert.ErrorIs(t, err, ErrInvalidOrgID)

	// 2. Missing lane code validation
	_, err = svc.GetOrPredictLanePerformance(ctx, 2, nil, "", false)
	assert.ErrorIs(t, err, ErrMissingRecordID)

	// 3. Generate prediction
	pred, err := svc.GetOrPredictLanePerformance(ctx, 2, nil, "INNSA-USNYC", true)
	require.NoError(t, err)
	assert.Equal(t, "pred-lane-usnyc-test", pred.PredictionID)
	assert.Equal(t, "INNSA-USNYC", *pred.LaneReference)
	assert.Equal(t, 18, *pred.SampleSize)
	assert.Equal(t, "LAST_90_DAYS", *pred.ComparisonPeriod)
}

func TestCustomerServicePerformance_WorkflowAndGovernance(t *testing.T) {
	repo := newMockRepo()

	predCategory := "CUSTOMER_SERVICE_RISK"
	compPeriod := "LAST_90_DAYS"
	sampleSize := 8
	custRef := "DEV-CUST-003"

	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-cust-103-test",
			OrgID:               2,
			Module:              "customers",
			PredictionType:      "CUSTOMER_SERVICE_RISK",
			PredictionCategory:  &predCategory,
			RelatedRecordType:   "CUSTOMER",
			RelatedRecordID:     "103",
			PredictionStatement: "Customer Service Risk: Bharat Tech Exports Pvt Ltd exhibits elevated service friction.",
			PredictedValue:      func(s string) *string { return &s }("ELEVATED_SERVICE_RISK"),
			TimeHorizon:         func(s string) *string { return &s }("14_DAYS"),
			Severity:            "HIGH",
			ConfidenceScore:     0.94,
			ConfidenceBand:      "HIGH",
			Explanation:         "Active customs hold on Shipment #103 at INNSA arising from Exception #101.",
			SupportingSignals: []SupportingSignal{
				{SignalName: "active_customs_exceptions", ObservedValue: "1", BaselineValue: "0", ImportanceWeight: 0.96},
				{SignalName: "account_health_score", ObservedValue: "65", BaselineValue: ">= 80", ImportanceWeight: 0.92},
			},
			SourceReferences: []SourceReference{
				{
					SourceModule:      "customers",
					SourceRecordID:    "103",
					SourceField:       "customers.health_score",
					SourceTimestamp:   time.Now().Format(time.RFC3339),
					CustomerReference: custRef,
					ComparisonPeriod:  compPeriod,
					SampleSize:        sampleSize,
				},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Schedule customer operations sync with account executive."),
			ActionType:        func(s string) *string { return &s }("customers.schedule_account_review"),
			IsActionRequired:  true,
			RequiresApproval:  true,
			ComparisonPeriod:  &compPeriod,
			SampleSize:        &sampleSize,
			CustomerReference: &custRef,
		},
	}

	svc := NewService(repo, sidecar)
	svcImpl := svc.(*service)
	svcImpl.networkPerfData = &NetworkPerformanceDataProvider{}

	ctx := context.Background()

	// 1. Tenant access control validation
	_, err := svc.GetOrPredictCustomerServicePerformance(ctx, 0, nil, 103, false)
	assert.ErrorIs(t, err, ErrInvalidOrgID)

	// 2. Missing customer ID validation
	_, err = svc.GetOrPredictCustomerServicePerformance(ctx, 2, nil, 0, false)
	assert.ErrorIs(t, err, ErrMissingRecordID)
}
