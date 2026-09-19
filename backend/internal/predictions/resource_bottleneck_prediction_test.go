package predictions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceBottleneck_ApprovalBottleneck(t *testing.T) {
	repo := newMockRepo()

	predCategory := "APPROVAL_BOTTLENECK"
	compPeriod := "LAST_30_DAYS"
	sampleSize := 20
	bottleneckType := "APPROVALS"
	affectedStage := "PRICING_APPROVAL"
	assignedOwner := "kanadevarun123@gmail.com"
	dwellHours := 22.4
	pendingCount := 20
	capLimit := 15

	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-btln-appr-test1",
			OrgID:               2,
			Module:              "bottleneck",
			PredictionType:      "APPROVAL_BOTTLENECK",
			PredictionCategory:  &predCategory,
			RelatedRecordType:   "BOTTLENECK",
			RelatedRecordID:     "approvals",
			PredictionStatement: "Likely Approval Bottleneck: 20 pending review requests exceed throughput limit with 22.4h average queue dwell time.",
			PredictedValue:      func(s string) *string { return &s }("APPROVAL_QUEUE_SATURATED"),
			TimeHorizon:         func(s string) *string { return &s }("7_DAYS"),
			Severity:            "HIGH",
			ConfidenceScore:     0.93,
			ConfidenceBand:      "HIGH",
			Explanation:         "Operational approval queue contains 20 pending requests exceeding operational throughput limit.",
			BottleneckType:      &bottleneckType,
			AffectedStage:       &affectedStage,
			AssignedOwner:       &assignedOwner,
			QueueDwellHours:     &dwellHours,
			PendingCount:        &pendingCount,
			CapacityLimit:       &capLimit,
			ComparisonPeriod:    &compPeriod,
			SampleSize:          &sampleSize,
			SupportingSignals: []SupportingSignal{
				{SignalName: "pending_approvals_count", ObservedValue: "20", BaselineValue: "<= 15", ImportanceWeight: 0.95},
			},
			SourceReferences: []SourceReference{
				{
					SourceModule:     "approval_requests",
					SourceRecordID:   "105",
					SourceField:      "approval_requests.status",
					SourceTimestamp:  time.Now().Format(time.RFC3339),
					BottleneckType:   bottleneckType,
					AffectedStage:    affectedStage,
					AssignedOwner:    assignedOwner,
					QueueDwellHours:  dwellHours,
					PendingCount:     pendingCount,
					CapacityLimit:    capLimit,
					ComparisonPeriod: compPeriod,
					SampleSize:       sampleSize,
				},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Delegate commercial quote reviews to authorized team leads and batch-approve pricing exceptions."),
			ActionType:        func(s string) *string { return &s }("bottlenecks.rebalance_approval_queue"),
			IsActionRequired:  true,
			RequiresApproval:  true,
			ModelVersion:      "gemini-1.5-pro",
		},
	}

	svc := NewService(repo, sidecar)
	pred, err := svc.GetOrPredictOperationalBottleneck(context.Background(), 2, nil, "approvals", false)
	require.NoError(t, err)
	require.NotNil(t, pred)

	assert.Equal(t, "pred-btln-appr-test1", pred.PredictionID)
	assert.Equal(t, int64(2), pred.OrgID)
	assert.Equal(t, "bottleneck", pred.Module)
	assert.Equal(t, "APPROVAL_BOTTLENECK", pred.PredictionType)
	assert.Equal(t, SeverityHigh, pred.Severity)
	assert.Equal(t, 0.93, pred.ConfidenceScore)
	assert.NotNil(t, pred.BottleneckType)
	assert.Equal(t, "APPROVALS", *pred.BottleneckType)
	assert.NotNil(t, pred.PendingCount)
	assert.Equal(t, 20, *pred.PendingCount)
	assert.NotNil(t, pred.QueueDwellHours)
	assert.Equal(t, 22.4, *pred.QueueDwellHours)
	assert.True(t, pred.IsActionRequired)
	assert.True(t, pred.RequiresApproval)
}

func TestResourceBottleneck_DocumentationBottleneck(t *testing.T) {
	repo := newMockRepo()

	predCategory := "DOCUMENTATION_BOTTLENECK"
	compPeriod := "LAST_30_DAYS"
	sampleSize := 5
	bottleneckType := "DOCUMENTATION"
	affectedStage := "MANIFEST_AND_CUSTOMS_FILING"
	dwellHours := 36.5
	pendingCount := 5

	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-btln-docs-test1",
			OrgID:               2,
			Module:              "bottleneck",
			PredictionType:      "DOCUMENTATION_BOTTLENECK",
			PredictionCategory:  &predCategory,
			RelatedRecordType:   "BOTTLENECK",
			RelatedRecordID:     "documentation",
			PredictionStatement: "Likely Documentation Bottleneck: 5 document compliance discrepancies pending across 3 active ocean shipments.",
			PredictedValue:      func(s string) *string { return &s }("DOCUMENTATION_BACKLOG_HIGH"),
			TimeHorizon:         func(s string) *string { return &s }("7_DAYS"),
			Severity:            "HIGH",
			ConfidenceScore:     0.91,
			ConfidenceBand:      "HIGH",
			Explanation:         "Operational documentation backlog exhibits 5 unresolved discrepancies across 3 shipments.",
			BottleneckType:      &bottleneckType,
			AffectedStage:       &affectedStage,
			QueueDwellHours:     &dwellHours,
			PendingCount:        &pendingCount,
			ComparisonPeriod:    &compPeriod,
			SampleSize:          &sampleSize,
			SupportingSignals: []SupportingSignal{
				{SignalName: "active_discrepancies_count", ObservedValue: "5", BaselineValue: "0", ImportanceWeight: 0.95},
			},
			SourceReferences: []SourceReference{
				{
					SourceModule:     "shipment_document_discrepancies",
					SourceRecordID:   "101",
					SourceField:      "shipment_document_discrepancies.status",
					SourceTimestamp:  time.Now().Format(time.RFC3339),
					BottleneckType:   bottleneckType,
					AffectedStage:    affectedStage,
					QueueDwellHours:  dwellHours,
					PendingCount:     pendingCount,
					ComparisonPeriod: compPeriod,
					SampleSize:       sampleSize,
				},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Prioritize weight amendment filing for Shipment #101."),
			ActionType:        func(s string) *string { return &s }("bottlenecks.resolve_document_discrepancies"),
			IsActionRequired:  true,
			RequiresApproval:  true,
			ModelVersion:      "gemini-1.5-pro",
		},
	}

	svc := NewService(repo, sidecar)
	pred, err := svc.GetOrPredictOperationalBottleneck(context.Background(), 2, nil, "documentation", false)
	require.NoError(t, err)
	require.NotNil(t, pred)

	assert.Equal(t, "pred-btln-docs-test1", pred.PredictionID)
	assert.Equal(t, "DOCUMENTATION_BOTTLENECK", pred.PredictionType)
	assert.Equal(t, 5, *pred.PendingCount)
	assert.Equal(t, 36.5, *pred.QueueDwellHours)
}

func TestResourceBottleneck_CrossModuleBottleneck(t *testing.T) {
	repo := newMockRepo()

	predCategory := "CROSS_MODULE_BOTTLENECK"
	compPeriod := "LAST_30_DAYS"
	sampleSize := 1
	bottleneckType := "EXCEPTIONS"
	carrierSCAC := "CMDU"
	dwellHours := 48.0
	linkedExc := int64(101)

	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-btln-cross-test1",
			OrgID:               2,
			Module:              "bottleneck",
			PredictionType:      "CROSS_MODULE_BOTTLENECK",
			PredictionCategory:  &predCategory,
			RelatedRecordType:   "BOTTLENECK",
			RelatedRecordID:     "exceptions",
			PredictionStatement: "Likely Cross-Module Bottleneck: Terminal customs hold on Shipment #103 blocks downstream delivery.",
			Severity:            "HIGH",
			ConfidenceScore:     0.94,
			ConfidenceBand:      "HIGH",
			Explanation:         "Shipment #103 is detained under CUSTOMS_HOLD at Port of NY/NJ Terminal.",
			BottleneckType:      &bottleneckType,
			CarrierReference:    &carrierSCAC,
			QueueDwellHours:     &dwellHours,
			LinkedExceptionID:   &linkedExc,
			ComparisonPeriod:    &compPeriod,
			SampleSize:          &sampleSize,
			SourceReferences: []SourceReference{
				{
					SourceModule:     "shipment_exceptions",
					SourceRecordID:   "103",
					SourceField:      "shipments.status",
					SourceTimestamp:  time.Now().Format(time.RFC3339),
					CarrierReference: carrierSCAC,
					BottleneckType:   bottleneckType,
					QueueDwellHours:  dwellHours,
					ComparisonPeriod: compPeriod,
					SampleSize:       sampleSize,
				},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Engage customs broker exam liaison to clear hold."),
			IsActionRequired:  true,
			RequiresApproval:  true,
			ModelVersion:      "gemini-1.5-pro",
		},
	}

	svc := NewService(repo, sidecar)
	pred, err := svc.GetOrPredictOperationalBottleneck(context.Background(), 2, nil, "exceptions", false)
	require.NoError(t, err)
	require.NotNil(t, pred)

	assert.Equal(t, "pred-btln-cross-test1", pred.PredictionID)
	assert.Equal(t, "CROSS_MODULE_BOTTLENECK", pred.PredictionType)
	assert.Equal(t, "CMDU", *pred.CarrierReference)
	assert.Equal(t, 48.0, *pred.QueueDwellHours)
}

func TestResourceBottleneck_ResourceAllocation(t *testing.T) {
	repo := newMockRepo()

	predCategory := "RESOURCE_ALLOCATION_IMBALANCE"
	compPeriod := "LAST_30_DAYS"
	sampleSize := 20
	bottleneckType := "RESOURCE"
	assignedOwner := "kanadevarun123@gmail.com"
	pendingCount := 20
	utilRate := 95.0

	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-rsrc-alloc-test1",
			OrgID:               2,
			Module:              "resource",
			PredictionType:      "RESOURCE_ALLOCATION_IMBALANCE",
			PredictionCategory:  &predCategory,
			RelatedRecordType:   "RESOURCE",
			RelatedRecordID:     "allocation",
			PredictionStatement: "Likely Resource Allocation Imbalance: 100% of operational approvals and escalation reviews are assigned to owner kanadevarun123@gmail.com.",
			Severity:            "HIGH",
			ConfidenceScore:     0.89,
			ConfidenceBand:      "HIGH",
			Explanation:         "Ownership telemetry indicates 20 pending approvals assigned exclusively to kanadevarun123@gmail.com.",
			BottleneckType:      &bottleneckType,
			AssignedOwner:       &assignedOwner,
			PendingCount:        &pendingCount,
			UtilizationRate:     &utilRate,
			ComparisonPeriod:    &compPeriod,
			SampleSize:          &sampleSize,
			SourceReferences: []SourceReference{
				{
					SourceModule:     "users",
					SourceRecordID:   assignedOwner,
					SourceField:      "users.id",
					SourceTimestamp:  time.Now().Format(time.RFC3339),
					BottleneckType:   bottleneckType,
					AssignedOwner:    assignedOwner,
					PendingCount:     pendingCount,
					ComparisonPeriod: compPeriod,
					SampleSize:       sampleSize,
				},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Rebalance approval authority across secondary operations managers."),
			IsActionRequired:  true,
			RequiresApproval:  true,
			ModelVersion:      "gemini-1.5-pro",
		},
	}

	svc := NewService(repo, sidecar)
	pred, err := svc.GetOrPredictResourceAllocation(context.Background(), 2, nil, "allocation", false)
	require.NoError(t, err)
	require.NotNil(t, pred)

	assert.Equal(t, "pred-rsrc-alloc-test1", pred.PredictionID)
	assert.Equal(t, "RESOURCE_ALLOCATION_IMBALANCE", pred.PredictionType)
	assert.Equal(t, "kanadevarun123@gmail.com", *pred.AssignedOwner)
	assert.Equal(t, 20, *pred.PendingCount)
	assert.Equal(t, 95.0, *pred.UtilizationRate)
}

func TestResourceBottleneck_ServiceCaching(t *testing.T) {
	repo := newMockRepo()

	bottleneckType := "APPROVALS"
	pendingCount := 20
	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-btln-appr-cache1",
			OrgID:               2,
			Module:              "bottleneck",
			PredictionType:      "APPROVAL_BOTTLENECK",
			RelatedRecordType:   "BOTTLENECK",
			RelatedRecordID:     "approvals",
			PredictionStatement: "Approval queue saturated.",
			Severity:            "HIGH",
			ConfidenceScore:     0.90,
			ConfidenceBand:      "HIGH",
			Explanation:         "Dwell time exceeds SLA limit.",
			BottleneckType:      &bottleneckType,
			PendingCount:        &pendingCount,
			SourceReferences: []SourceReference{
				{
					SourceModule:    "approval_requests",
					SourceRecordID:  "105",
					SourceField:     "approval_requests.status",
					SourceTimestamp: time.Now().Format(time.RFC3339),
				},
			},
			SourceTimestamp:  time.Now().Format(time.RFC3339),
			IsActionRequired: true,
			ModelVersion:     "gemini-1.5-pro",
		},
	}

	svc := NewService(repo, sidecar)

	// First invocation -> calls sidecar
	pred1, err := svc.GetOrPredictOperationalBottleneck(context.Background(), 2, nil, "approvals", false)
	require.NoError(t, err)
	assert.Equal(t, "pred-btln-appr-cache1", pred1.PredictionID)

	// Second invocation with forceRefresh=false -> returns cached without sidecar call
	sidecar.response = nil
	pred2, err := svc.GetOrPredictOperationalBottleneck(context.Background(), 2, nil, "approvals", false)
	require.NoError(t, err)
	assert.Equal(t, pred1.PredictionID, pred2.PredictionID)
}
