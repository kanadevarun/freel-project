package predictions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkloadCapacity_ApprovalWorkload(t *testing.T) {
	repo := newMockRepo()

	predCategory := "APPROVAL_WORKLOAD_SPIKE"
	compPeriod := "LAST_30_DAYS"
	sampleSize := 20
	workloadType := "APPROVALS"
	pendingCount := 20
	capLimit := 15
	utilRate := 92.5

	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-workload-appr-test1",
			OrgID:               2,
			Module:              "workload",
			PredictionType:      "APPROVAL_WORKLOAD_SPIKE",
			PredictionCategory:  &predCategory,
			RelatedRecordType:   "WORKLOAD",
			RelatedRecordID:     "approvals",
			PredictionStatement: "Approval Workload Spike: 20 pending high/critical review requests are concentrated in PRICING workflows.",
			PredictedValue:      func(s string) *string { return &s }("APPROVAL_SPIKE_HIGH"),
			TimeHorizon:         func(s string) *string { return &s }("7_DAYS"),
			Severity:            "HIGH",
			ConfidenceScore:     0.92,
			ConfidenceBand:      "HIGH",
			Explanation:         "Approval queue has 20 open requests with average dwell of 22.4 hours. Concentration in PRICING risks quote turnaround delays.",
			WorkloadType:        &workloadType,
			PendingCount:        &pendingCount,
			CapacityLimit:       &capLimit,
			UtilizationRate:     &utilRate,
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
					WorkloadType:     workloadType,
					PendingCount:     pendingCount,
					CapacityLimit:    capLimit,
					UtilizationRate:  utilRate,
					ComparisonPeriod: compPeriod,
					SampleSize:       sampleSize,
				},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Triage pending pricing and commercial draft approvals to clear operator queue backlog."),
			ActionType:        func(s string) *string { return &s }("workload.triage_approvals"),
			IsActionRequired:  true,
			RequiresApproval:  true,
			ModelVersion:      "gemini-1.5-pro",
		},
	}

	svc := NewService(repo, sidecar)
	svcImpl := svc.(*service)
	svcImpl.workloadCapacityData = &WorkloadCapacityDataProvider{}

	ctx := context.Background()

	// 1. Tenant validation
	_, err := svc.GetOrPredictWorkloadPlanning(ctx, 0, nil, "approvals", false)
	assert.ErrorIs(t, err, ErrInvalidOrgID)

	// 2. Fetch fresh prediction
	pred, err := svc.GetOrPredictWorkloadPlanning(ctx, 2, nil, "approvals", false)
	require.NoError(t, err)
	assert.Equal(t, "pred-workload-appr-test1", pred.PredictionID)
	assert.Equal(t, "APPROVAL_WORKLOAD_SPIKE", pred.PredictionType)
	assert.Equal(t, SeverityHigh, pred.Severity)
	assert.Equal(t, "APPROVALS", *pred.WorkloadType)
	assert.Equal(t, 20, *pred.PendingCount)
	assert.Equal(t, 92.5, *pred.UtilizationRate)
	assert.True(t, pred.IsActionRequired)
	assert.True(t, pred.RequiresApproval)

	// 3. Cache verification
	cached, err := svc.GetOrPredictWorkloadPlanning(ctx, 2, nil, "approvals", false)
	require.NoError(t, err)
	assert.Equal(t, "pred-workload-appr-test1", cached.PredictionID)
}

func TestWorkloadCapacity_DocumentationWorkload(t *testing.T) {
	repo := newMockRepo()

	predCategory := "DOCUMENTATION_WORKLOAD_SPIKE"
	compPeriod := "LAST_30_DAYS"
	sampleSize := 6
	workloadType := "DOCUMENTATION"
	pendingCount := 5
	capLimit := 2
	utilRate := 88.5

	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-workload-docs-test1",
			OrgID:               2,
			Module:              "workload",
			PredictionType:      "DOCUMENTATION_WORKLOAD_SPIKE",
			PredictionCategory:  &predCategory,
			RelatedRecordType:   "WORKLOAD",
			RelatedRecordID:     "documentation",
			PredictionStatement: "Operational Documentation Workload: 5 compliance issues pending across 3 active ocean shipments.",
			PredictedValue:      func(s string) *string { return &s }("DOCUMENTATION_BACKLOG_RISK"),
			TimeHorizon:         func(s string) *string { return &s }("7_DAYS"),
			Severity:            "HIGH",
			ConfidenceScore:     0.89,
			ConfidenceBand:      "HIGH",
			Explanation:         "Active ocean shipments exhibit documentation friction including customs hold and gross weight discrepancy.",
			WorkloadType:        &workloadType,
			PendingCount:        &pendingCount,
			CapacityLimit:       &capLimit,
			UtilizationRate:     &utilRate,
			ComparisonPeriod:    &compPeriod,
			SampleSize:          &sampleSize,
			SupportingSignals: []SupportingSignal{
				{SignalName: "open_discrepancies_count", ObservedValue: "1", BaselineValue: "0", ImportanceWeight: 0.94},
			},
			SourceReferences: []SourceReference{
				{
					SourceModule:    "shipments",
					SourceRecordID:  "101",
					SourceField:     "shipments.status",
					SourceTimestamp: time.Now().Format(time.RFC3339),
				},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Resolve MBL/HBL gross weight discrepancy and dispatch customs release documentation before upcoming arrival cutoffs."),
			ActionType:        func(s string) *string { return &s }("workload.expedite_documentation"),
			IsActionRequired:  true,
			RequiresApproval:  true,
			ModelVersion:      "gemini-1.5-pro",
		},
	}

	svc := NewService(repo, sidecar)
	svcImpl := svc.(*service)
	svcImpl.workloadCapacityData = &WorkloadCapacityDataProvider{}

	ctx := context.Background()

	pred, err := svc.GetOrPredictWorkloadPlanning(ctx, 2, nil, "documentation", false)
	require.NoError(t, err)
	assert.Equal(t, "pred-workload-docs-test1", pred.PredictionID)
	assert.Equal(t, "DOCUMENTATION_WORKLOAD_SPIKE", pred.PredictionType)
	assert.Equal(t, "DOCUMENTATION", *pred.WorkloadType)
	assert.Equal(t, 5, *pred.PendingCount)
	assert.True(t, pred.IsActionRequired)
}

func TestWorkloadCapacity_CorridorCapacity(t *testing.T) {
	repo := newMockRepo()

	predCategory := "DEMAND_CAPACITY_MISMATCH"
	compPeriod := "LAST_90_DAYS"
	sampleSize := 12
	laneRef := "INNSA-USNYC / INNSA-NLRTM"
	workloadType := "CAPACITY"
	utilRate := 87.5

	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-capacity-lane-test1",
			OrgID:               2,
			Module:              "capacity",
			PredictionType:      "DEMAND_CAPACITY_MISMATCH",
			PredictionCategory:  &predCategory,
			RelatedRecordType:   "CAPACITY",
			RelatedRecordID:     "corridors",
			PredictionStatement: "Trade Corridor Capacity Pressure: Corridors INNSA-USNYC / INNSA-NLRTM display demand concentration against terminal customs hold dwell.",
			PredictedValue:      func(s string) *string { return &s }("CAPACITY_PRESSURE_HIGH"),
			TimeHorizon:         func(s string) *string { return &s }("14_DAYS"),
			Severity:            "HIGH",
			ConfidenceScore:     0.88,
			ConfidenceBand:      "HIGH",
			Explanation:         "Active booking commitments and RFQ pipeline are concentrated on West Coast India export lanes.",
			LaneReference:       &laneRef,
			WorkloadType:        &workloadType,
			UtilizationRate:     &utilRate,
			ComparisonPeriod:    &compPeriod,
			SampleSize:          &sampleSize,
			SupportingSignals: []SupportingSignal{
				{SignalName: "carrier_utilization_pct", ObservedValue: "87.5%", BaselineValue: "<= 80.0%", ImportanceWeight: 0.92},
			},
			SourceReferences: []SourceReference{
				{
					SourceModule:    "bookings",
					SourceRecordID:  "101",
					SourceField:     "bookings.carrier_booking_status",
					SourceTimestamp: time.Now().Format(time.RFC3339),
				},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Review secondary carrier allocations on Trans-Atlantic and European corridors to mitigate space constraints."),
			ActionType:        func(s string) *string { return &s }("capacity.reallocate_carrier_space"),
			IsActionRequired:  true,
			RequiresApproval:  true,
			ModelVersion:      "gemini-1.5-pro",
		},
	}

	svc := NewService(repo, sidecar)
	svcImpl := svc.(*service)
	svcImpl.workloadCapacityData = &WorkloadCapacityDataProvider{}

	ctx := context.Background()

	pred, err := svc.GetOrPredictCapacityPlanning(ctx, 2, nil, "corridors", false)
	require.NoError(t, err)
	assert.Equal(t, "pred-capacity-lane-test1", pred.PredictionID)
	assert.Equal(t, "DEMAND_CAPACITY_MISMATCH", pred.PredictionType)
	assert.Equal(t, "INNSA-USNYC / INNSA-NLRTM", *pred.LaneReference)
	assert.Equal(t, 87.5, *pred.UtilizationRate)
}

func TestWorkloadCapacity_QuoteDemand(t *testing.T) {
	repo := newMockRepo()

	predCategory := "QUOTE_PROCESSING_BOTTLENECK"
	compPeriod := "LAST_30_DAYS"
	sampleSize := 5
	custRef := "Apex Global Logistics"
	workloadType := "DEMAND"
	pendingCount := 4

	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-demand-quote-test1",
			OrgID:               2,
			Module:              "demand",
			PredictionType:      "QUOTE_PROCESSING_BOTTLENECK",
			PredictionCategory:  &predCategory,
			RelatedRecordType:   "DEMAND",
			RelatedRecordID:     "commercial",
			PredictionStatement: "Commercial Quote Processing Demand: 4 active RFQs in pipeline with 60% concentration from account Apex Global Logistics.",
			PredictedValue:      func(s string) *string { return &s }("QUOTE_BACKLOG_MEDIUM"),
			TimeHorizon:         func(s string) *string { return &s }("14_DAYS"),
			Severity:            "MEDIUM",
			ConfidenceScore:     0.86,
			ConfidenceBand:      "HIGH",
			Explanation:         "RFQ pipeline contains 4 active inquiries awaiting sales pricing and tariff confirmation.",
			CustomerReference:   &custRef,
			WorkloadType:        &workloadType,
			PendingCount:        &pendingCount,
			ComparisonPeriod:    &compPeriod,
			SampleSize:          &sampleSize,
			SupportingSignals: []SupportingSignal{
				{SignalName: "pipeline_rfq_count", ObservedValue: "4", BaselineValue: "<= 2", ImportanceWeight: 0.90},
			},
			SourceReferences: []SourceReference{
				{
					SourceModule:    "rfqs",
					SourceRecordID:  "102",
					SourceField:     "rfqs.status",
					SourceTimestamp: time.Now().Format(time.RFC3339),
				},
			},
			SourceTimestamp:   time.Now().Format(time.RFC3339),
			RecommendedAction: func(s string) *string { return &s }("Prioritize draft quote formulation for high-value accounts and dispatch pricing reviews."),
			ActionType:        func(s string) *string { return &s }("demand.prioritize_rfqs"),
			IsActionRequired:  true,
			RequiresApproval:  true,
			ModelVersion:      "gemini-1.5-pro",
		},
	}

	svc := NewService(repo, sidecar)
	svcImpl := svc.(*service)
	svcImpl.workloadCapacityData = &WorkloadCapacityDataProvider{}

	ctx := context.Background()

	pred, err := svc.GetOrPredictDemandPlanning(ctx, 2, nil, "commercial", false)
	require.NoError(t, err)
	assert.Equal(t, "pred-demand-quote-test1", pred.PredictionID)
	assert.Equal(t, "QUOTE_PROCESSING_BOTTLENECK", pred.PredictionType)
	assert.Equal(t, "Apex Global Logistics", *pred.CustomerReference)
	assert.Equal(t, 4, *pred.PendingCount)
}

func TestWorkloadCapacity_InsufficientData(t *testing.T) {
	repo := newMockRepo()

	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-workload-nodata-test1",
			OrgID:               2,
			Module:              "workload",
			PredictionType:      "APPROVAL_WORKLOAD_SPIKE",
			RelatedRecordType:   "WORKLOAD",
			RelatedRecordID:     "approvals",
			PredictionStatement: "Insufficient historical planning telemetry for WORKLOAD #approvals.",
			PredictedValue:      func(s string) *string { return &s }("INSUFFICIENT_DATA"),
			Severity:            "LOW",
			ConfidenceScore:     0.0,
			ConfidenceBand:      "LOW",
			Explanation:         "Operational records lack sufficient active workload, booking, or pipeline history.",
			SupportingSignals:   []SupportingSignal{},
			SourceReferences: []SourceReference{
				{
					SourceModule:    "workload",
					SourceRecordID:  "approvals",
					SourceField:     "operational.sample_size",
					SourceTimestamp: time.Now().Format(time.RFC3339),
					SampleSize:      0,
				},
			},
			SourceTimestamp:    time.Now().Format(time.RFC3339),
			RecommendedAction:  func(s string) *string { return &s }("Wait for operational transactions to accumulate before re-evaluating workload capacity."),
			IsActionRequired:   false,
			RequiresApproval:   false,
			InsufficientData:   true,
			InsufficientReason: func(s string) *string { return &s }("Sample size is zero or telemetry context is below minimum statistical thresholds."),
			ModelVersion:       "gemini-1.5-pro",
		},
	}

	svc := NewService(repo, sidecar)
	svcImpl := svc.(*service)
	svcImpl.workloadCapacityData = &WorkloadCapacityDataProvider{}

	ctx := context.Background()

	pred, err := svc.GetOrPredictWorkloadPlanning(ctx, 2, nil, "approvals", false)
	require.NoError(t, err)
	assert.True(t, pred.InsufficientData)
	assert.Equal(t, "INSUFFICIENT_DATA", *pred.PredictedValue)
	assert.Equal(t, SeverityLow, pred.Severity)
	assert.Equal(t, 0.0, pred.ConfidenceScore)
	assert.False(t, pred.IsActionRequired)
}
