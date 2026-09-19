package predictions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShipmentReadiness_WorkflowAndGovernance(t *testing.T) {
	repo := newMockRepo()

	predCategory := "DOCUMENTATION_DISCREPANCY"
	docRef := "shipment_documents:111"
	milestoneRef := "ARRIVAL"
	cutoffRef := "IMPORT_MANIFEST_DEADLINE"

	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-readiness-sh-101-test",
			OrgID:               2,
			Module:              "shipments",
			PredictionType:      "DOCUMENTATION_DELAY_RISK",
			PredictionCategory:  &predCategory,
			RelatedRecordType:   "SHIPMENT",
			RelatedRecordID:     "101",
			PredictionStatement: "Documentation Discrepancy & Cutoff Risk: Unresolved MBL vs HBL gross weight mismatch may breach import manifest deadline.",
			PredictedValue:      func(s string) *string { return &s }("DISCREPANCY_MANIFEST_BREACH"),
			TimeHorizon:         func(s string) *string { return &s }("48_HOURS"),
			Severity:            "HIGH",
			ConfidenceScore:     0.92,
			ConfidenceBand:      "HIGH",
			Explanation:         "Gross weight variance between MBL (24,500 kg) and HBL (21,200 kg) threatens customs hold.",
			DocumentReference:   &docRef,
			MilestoneReference:  &milestoneRef,
			CutoffReference:     &cutoffRef,
			SupportingSignals: []SupportingSignal{
				{SignalName: "open_discrepancy_count", ObservedValue: "1", BaselineValue: "0", ImportanceWeight: 0.95},
				{SignalName: "missing_document_count", ObservedValue: "2", BaselineValue: "0", ImportanceWeight: 0.88},
			},
			SourceReferences: []SourceReference{
				{SourceModule: "shipments", SourceRecordID: "101", SourceField: "shipments.id", SourceTimestamp: time.Now().Format(time.RFC3339)},
				{SourceModule: "shipment_documents", SourceRecordID: "111", SourceField: "shipment_documents.status", SourceTimestamp: time.Now().Format(time.RFC3339)},
				{SourceModule: "shipment_document_discrepancies", SourceRecordID: "5", SourceField: "shipment_document_discrepancies.gross_weight", SourceTimestamp: time.Now().Format(time.RFC3339)},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Request shipper confirmation on gross weight variance before vessel arrival."),
			ActionType:        func(s string) *string { return &s }("shipments.request_document_review"),
			IsActionRequired:  true,
			RequiresApproval:  true,
		},
	}

	svc := NewService(repo, sidecar)
	svcImpl := svc.(*service)
	svcImpl.shipmentReadinessData = &ShipmentReadinessDataProvider{}

	ctx := context.Background()

	// 1. Tenant access control validation
	_, err := svc.GetOrPredictShipmentReadiness(ctx, 0, nil, 101, false)
	assert.ErrorIs(t, err, ErrInvalidOrgID)

	// 2. Invalid Shipment ID validation
	_, err = svc.GetOrPredictShipmentReadiness(ctx, 2, nil, 0, false)
	assert.Error(t, err)

	// 3. Pre-seed active prediction in mock repo
	prePred := &Prediction{
		OrgID:               2,
		PredictionID:        "pred-readiness-sh-101-initial",
		IdempotencyKey:      "pred:shipments:readiness:2:101:initial",
		Module:              "shipments",
		PredictionType:      "DOCUMENTATION_DELAY_RISK",
		Status:              StatusPublished,
		Severity:            SeverityHigh,
		ConfidenceScore:     0.92,
		ConfidenceBand:      ConfidenceHigh,
		RelatedRecordType:   "SHIPMENT",
		RelatedRecordID:     "101",
		PredictionStatement: "Initial readiness prediction",
		Explanation:         "Initial documentation variance",
		DocumentReference:   &docRef,
		MilestoneReference:  &milestoneRef,
		CutoffReference:     &cutoffRef,
		SourceReferences: []SourceReference{
			{SourceModule: "shipments", SourceRecordID: "101", SourceField: "shipments.id", SourceTimestamp: time.Now().Format(time.RFC3339)},
		},
		SourceTimestamp:   time.Now(),
		RecommendedAction: func(s string) *string { return &s }("Request shipper confirmation"),
		ActionType:        func(s string) *string { return &s }("shipments.request_document_review"),
		ReviewStatus:      "PENDING",
		IsActionRequired:  true,
		RequiresApproval:  true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	err = repo.Create(ctx, prePred)
	require.NoError(t, err)

	// Fetch without force refresh: should return the cached prediction
	cached, err := svc.GetOrPredictShipmentReadiness(ctx, 2, nil, 101, false)
	require.NoError(t, err)
	assert.Equal(t, "pred-readiness-sh-101-initial", cached.PredictionID)
	assert.Equal(t, "ARRIVAL", *cached.MilestoneReference)
	assert.Equal(t, "IMPORT_MANIFEST_DEADLINE", *cached.CutoffReference)

	// 4. Governance lifecycle actions
	userID := int64(10)
	err = svc.AcknowledgePrediction(ctx, 2, cached.PredictionID, &userID)
	require.NoError(t, err)
	updated, err := svc.GetPrediction(ctx, 2, cached.PredictionID)
	require.NoError(t, err)
	assert.Equal(t, StatusAcknowledged, updated.Status)

	err = svc.RequestAction(ctx, 2, cached.PredictionID, &userID, "Requested urgent manifest amendment")
	require.NoError(t, err)
	actioned, err := svc.GetPrediction(ctx, 2, cached.PredictionID)
	require.NoError(t, err)
	assert.Equal(t, StatusAwaitingApproval, actioned.Status)

	// 5. Audit history tracking
	history, err := svc.GetAuditHistory(ctx, 2, cached.PredictionID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(history), 2)
}

func TestShipmentReadiness_InsufficientDataAndCustomsHold(t *testing.T) {
	repo := newMockRepo()

	customsCategory := "REGULATORY_COMPLIANCE"
	docRef := "shipment_documents:103"
	milestoneRef := "CUSTOMS_CLEARANCE"
	cutoffRef := "CUSTOMS_SUBMISSION_DEADLINE"

	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-readiness-sh-103-customs",
			OrgID:               2,
			Module:              "shipments",
			PredictionType:      "CUSTOMS_PROCESSING_RISK",
			PredictionCategory:  &customsCategory,
			RelatedRecordType:   "SHIPMENT",
			RelatedRecordID:     "103",
			PredictionStatement: "Customs Hold & Clearance Risk: Active customs hold on HS code declaration requires tariff review.",
			PredictedValue:      func(s string) *string { return &s }("REGULATORY_DETENTION"),
			TimeHorizon:         func(s string) *string { return &s }("IMMEDIATE"),
			Severity:            "CRITICAL",
			ConfidenceScore:     0.95,
			ConfidenceBand:      "HIGH",
			Explanation:         "Active customs hold with critical exception 101.",
			DocumentReference:   &docRef,
			MilestoneReference:  &milestoneRef,
			CutoffReference:     &cutoffRef,
			SourceReferences: []SourceReference{
				{SourceModule: "shipments", SourceRecordID: "103", SourceField: "shipments.status", SourceTimestamp: time.Now().Format(time.RFC3339)},
				{SourceModule: "shipment_exceptions", SourceRecordID: "101", SourceField: "shipment_exceptions.exception_type", SourceTimestamp: time.Now().Format(time.RFC3339)},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Submit corrected tariff declaration to customs authority immediately."),
			ActionType:        func(s string) *string { return &s }("shipments.escalate_customs_hold"),
			IsActionRequired:  true,
			RequiresApproval:  true,
		},
	}

	svc := NewService(repo, sidecar)
	ctx := context.Background()

	// Direct persistence & validation
	genReq := &GenerateRequest{
		OrgID:             2,
		Module:            "shipments",
		PredictionType:    "CUSTOMS_PROCESSING_RISK",
		RelatedRecordType: "SHIPMENT",
		RelatedRecordID:   "103",
		RecordContext: map[string]interface{}{
			"shipment_id": 103,
			"status":      "CUSTOMS_HOLD",
		},
		TimeHorizon: "IMMEDIATE",
	}

	pred, err := svc.GenerateAndPersist(ctx, 2, nil, genReq)
	require.NoError(t, err)
	assert.Equal(t, SeverityCritical, pred.Severity)
	assert.Equal(t, "CUSTOMS_PROCESSING_RISK", pred.PredictionType)
	assert.Equal(t, "CUSTOMS_CLEARANCE", *pred.MilestoneReference)
	assert.Equal(t, "CUSTOMS_SUBMISSION_DEADLINE", *pred.CutoffReference)
}
