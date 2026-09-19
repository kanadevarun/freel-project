package predictions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShipmentExceptionForecast_WorkflowAndGovernance(t *testing.T) {
	repo := newMockRepo()

	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-disrupt-103-test",
			OrgID:               2,
			Module:              "shipments",
			PredictionType:      "SHIPMENT_EXCEPTION_RISK",
			RelatedRecordType:   "SHIPMENT",
			RelatedRecordID:     "103",
			PredictionStatement: "Regulatory Disruption Forecast: Shipment #103 is under critical customs detention.",
			PredictedValue:      func(s string) *string { return &s }("CUSTOMS_CLEARANCE_RISK"),
			DisruptionCategory:  func(s string) *string { return &s }("CUSTOMS_CLEARANCE_RISK"),
			LinkedExceptionID:   func(i int64) *int64 { return &i }(101),
			TimeHorizon:         func(s string) *string { return &s }("7_DAYS"),
			Severity:            "CRITICAL",
			ConfidenceScore:     0.94,
			ConfidenceBand:      "HIGH",
			Explanation:         "Active customs hold detected. HS code discrepancy requires commercial broker resolution.",
			SupportingSignals: []SupportingSignal{
				{SignalName: "customs_detention_status", ObservedValue: "CUSTOMS_HOLD", BaselineValue: "CLEARED", ImportanceWeight: 0.98},
				{SignalName: "confirmed_exception_severity", ObservedValue: "CRITICAL", BaselineValue: "NONE", ImportanceWeight: 0.95},
			},
			SourceReferences: []SourceReference{
				{SourceModule: "shipment_exceptions", SourceRecordID: "101", SourceField: "shipment_exceptions.exception_type", SourceTimestamp: time.Now().Format(time.RFC3339)},
				{SourceModule: "shipments", SourceRecordID: "103", SourceField: "shipments.status", SourceTimestamp: time.Now().Format(time.RFC3339)},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Submit amended commercial invoice to customs authority for expedited clearance."),
			ActionType:        func(s string) *string { return &s }("shipments.expedite_customs_clearance"),
			IsActionRequired:  true,
			RequiresApproval:  true,
		},
	}

	svc := NewService(repo, sidecar)
	svcImpl := svc.(*service)
	svcImpl.dataProvider = &ShipmentDataProvider{}

	ctx := context.Background()

	// 1. Tenant access control test
	_, err := svc.GetOrForecastShipmentExceptions(ctx, 0, nil, 103, false)
	assert.ErrorIs(t, err, ErrInvalidOrgID)

	// 2. Invalid shipment ID test
	_, err = svc.GetOrForecastShipmentExceptions(ctx, 2, nil, 0, false)
	assert.Error(t, err)

	// 3. Pre-seed active forecast
	prePred := &Prediction{
		OrgID:               2,
		PredictionID:        "pred-disrupt-103-initial",
		IdempotencyKey:      "org:2:shipment_exception_forecast:103:cycle_initial",
		Module:              "shipments",
		PredictionType:      "SHIPMENT_EXCEPTION_RISK",
		Status:              StatusPublished,
		Severity:            SeverityCritical,
		ConfidenceScore:     0.90,
		ConfidenceBand:      ConfidenceHigh,
		RelatedRecordType:   "SHIPMENT",
		RelatedRecordID:     "103",
		PredictionStatement: "Initial customs hold forecast",
		Explanation:         "Active customs hold",
		SourceReferences: []SourceReference{
			{SourceModule: "shipments", SourceRecordID: "103", SourceField: "status", SourceTimestamp: time.Now().Format(time.RFC3339)},
		},
		SourceTimestamp: time.Now(),
		ReviewStatus:    "PENDING",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err = repo.Create(ctx, prePred)
	require.NoError(t, err)

	// Verify existing active forecast is retrieved when forceRefresh is false
	cached, err := svc.GetOrForecastShipmentExceptions(ctx, 2, nil, 103, false)
	require.NoError(t, err)
	assert.Equal(t, "pred-disrupt-103-initial", cached.PredictionID)
	assert.Equal(t, SeverityCritical, cached.Severity)
}
