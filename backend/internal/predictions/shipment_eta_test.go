package predictions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShipmentETAPrediction_ValidationAndWorkflow(t *testing.T) {
	repo := newMockRepo()

	targetDateStr := "2026-09-18 18:00:00"
	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-ship-101-test",
			OrgID:               2,
			Module:              "shipments",
			PredictionType:      "SHIPMENT_ETA_DELAY",
			RelatedRecordType:   "SHIPMENT",
			RelatedRecordID:     "101",
			PredictionStatement: "Vessel schedule modeling projects +48h delay arriving at Rotterdam.",
			PredictedValue:      func(s string) *string { return &s }("+48h Variance (17 Sep – 19 Sep 2026)"),
			TimeHorizon:         func(s string) *string { return &s }("7_DAYS"),
			TargetDate:          &targetDateStr,
			Severity:            "HIGH",
			ConfidenceScore:     0.88,
			ConfidenceBand:      "HIGH",
			Explanation:         "Delay drivers: Origin departure delay of +24h; Destination berth dwell queuing.",
			SupportingSignals: []SupportingSignal{
				{SignalName: "departure_schedule_variance", ObservedValue: "+24.0h", BaselineValue: "0.0h", ImportanceWeight: 0.85},
			},
			SourceReferences: []SourceReference{
				{SourceModule: "shipments", SourceRecordID: "101", SourceField: "shipment_milestones.DEPARTED", SourceTimestamp: time.Now().Format(time.RFC3339)},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Issue proactive consignee delivery update."),
			ActionType:        func(s string) *string { return &s }("shipments.send_consignee_delay_notice"),
			IsActionRequired:  true,
			RequiresApproval:  true,
		},
	}

	svc := NewService(repo, sidecar)
	// Inject a test mock data provider directly
	now := time.Now()
	eta := now.Add(7 * 24 * time.Hour)
	svcImpl := svc.(*service)
	svcImpl.dataProvider = &ShipmentDataProvider{} // will use custom runner or override

	// Test Tenant context check
	_, err := svc.GetOrPredictShipmentETA(context.Background(), 0, nil, 101, false)
	assert.ErrorIs(t, err, ErrInvalidOrgID)

	// Test Invalid shipment ID
	_, err = svc.GetOrPredictShipmentETA(context.Background(), 2, nil, 0, false)
	assert.Error(t, err)

	// Pre-seed an existing active prediction in repo
	prePred := &Prediction{
		OrgID:               2,
		PredictionID:        "pred-ship-101-initial",
		IdempotencyKey:      "2:shipment:101:eta:init",
		Module:              "shipments",
		PredictionType:      "SHIPMENT_ETA_DELAY",
		Status:              StatusPublished,
		Severity:            SeverityHigh,
		ConfidenceScore:     0.85,
		ConfidenceBand:      ConfidenceHigh,
		RelatedRecordType:   "SHIPMENT",
		RelatedRecordID:     "101",
		PredictionStatement: "Initial forecast",
		ExpiresAt:           &eta,
	}
	require.NoError(t, repo.Create(context.Background(), prePred))

	// When forceRefresh=false, it should return the active cached prediction without calling data provider or sidecar
	cached, err := svc.GetOrPredictShipmentETA(context.Background(), 2, nil, 101, false)
	require.NoError(t, err)
	assert.Equal(t, "pred-ship-101-initial", cached.PredictionID)
	assert.Equal(t, StatusPublished, cached.Status)

	// Test Superseding functionality in repo
	err = repo.SupersedeExisting(context.Background(), 2, "shipments", "SHIPMENT", "101", "SHIPMENT_ETA_DELAY", "pred-new-test")
	require.NoError(t, err)

	oldPred, err := repo.GetByID(context.Background(), 2, "pred-ship-101-initial")
	require.NoError(t, err)
	assert.Equal(t, StatusSuperseded, oldPred.Status)
}
