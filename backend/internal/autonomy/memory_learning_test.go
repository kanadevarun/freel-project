package autonomy_test

import (
	"context"
	"testing"

	"github.com/freel/backend/internal/autonomy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryLearning_CaptureOutcome_Success(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	orgID := int64(1)
	userID := int64(42)

	planID := "plan-recov-001"
	stepID := "step-escalate-1"
	actType := "carrier.escalate"
	failCat := "NONE"
	req := autonomy.RecordOutcomeRequest{
		SourceEntityType: "shipment",
		SourceEntityID:   "101",
		PlanID:           &planID,
		StepID:           &stepID,
		ActionType:       &actType,
		OutcomeType:      "EXCEPTION_RECOVERY",
		ExpectedResult:   "Carrier acknowledges delay within 6 hours and expedites berth assignment",
		ActualResult:     "Carrier responded in 4.5 hours with expedited berth confirmed for tomorrow",
		Status:           "SUCCESS",
		IsVerified:       true,
		FailureCategory:  &failCat,
	}

	outcome, mem, err := svc.CaptureOutcome(ctx, orgID, &userID, &req)
	require.NoError(t, err)
	require.NotNil(t, outcome)

	assert.Equal(t, orgID, outcome.OrgID)
	assert.Equal(t, "SUCCESS", outcome.Status)
	assert.True(t, outcome.IsVerified)
	assert.Equal(t, "shipment", outcome.SourceEntityType)
	assert.Equal(t, "101", outcome.SourceEntityID)

	// Verify structured memory item was derived
	require.NotNil(t, mem)
	assert.Equal(t, "OPERATIONAL", mem.Category)
	assert.Equal(t, "AI_DERIVED", mem.ProvenanceType)
	assert.Equal(t, "ACTIVE", mem.Status)
	assert.Equal(t, 1.0, mem.RecencyWeight)
}

func TestMemoryLearning_CaptureOutcome_Failure(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	orgID := int64(1)
	failCat := "CARRIER_NON_RESPONSIVE"
	reason := "Standard reminder timed out after 24h with no reply from dispatch team"
	req := autonomy.RecordOutcomeRequest{
		SourceEntityType: "shipment",
		SourceEntityID:   "102",
		OutcomeType:      "CARRIER_COMMUNICATION",
		ExpectedResult:   "Carrier confirms updated container tracking numbers",
		ActualResult:     "No response received from carrier within window",
		Status:           "FAILED",
		FailureCategory:  &failCat,
		Reason:           &reason,
		IsVerified:       false,
	}

	outcome, mem, err := svc.CaptureOutcome(ctx, orgID, nil, &req)
	require.NoError(t, err)
	require.NotNil(t, outcome)
	assert.Equal(t, "FAILED", outcome.Status)
	assert.Equal(t, "CARRIER_NON_RESPONSIVE", outcome.FailureCategory)
	assert.False(t, outcome.IsVerified)
	// Sidecar mock returns memory candidate, but failure is safely classified
	assert.NotNil(t, mem)
}

func TestMemoryLearning_VerifyOutcome_CreatesAuthoritativeMemory(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	orgID := int64(1)
	userID := int64(42)

	actRes := "Terminal gate confirmed container departed on schedule"
	stat := "SUCCESS"
	req := autonomy.VerifyOutcomeRequest{
		ActualResult:       &actRes,
		VerificationMethod: "PORT_API_EDI_CONFIRMATION",
		Status:             &stat,
	}

	outcome, createdMem, err := svc.VerifyOutcome(ctx, orgID, "out-mock-1", &userID, &req)
	require.NoError(t, err)
	require.NotNil(t, outcome)

	// Verification promotes outcome to verified memory
	require.NotNil(t, createdMem)
	assert.Equal(t, "SYSTEM_DERIVED", createdMem.ProvenanceType)
	assert.Equal(t, 0.95, createdMem.Confidence)
	assert.Equal(t, "ACTIVE", createdMem.Status)
	assert.Equal(t, "OPERATIONAL", createdMem.Category)
	assert.Contains(t, createdMem.Title, "Verified Outcome")
}

func TestMemoryLearning_RetrieveContextualMemory(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	orgID := int64(1)

	resp, err := svc.RetrieveContextualMemory(ctx, orgID, "Vessel delay approaching Port of Oakland", "shipments", "OPERATIONAL", "shipment", "101", 5)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.GreaterOrEqual(t, resp.TotalFound, 1)
	assert.NotEmpty(t, resp.RetrievedMemories)
	assert.False(t, resp.HasConflicts)
}

func TestMemoryLearning_HumanCorrection(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	orgID := int64(1)
	userID := int64(42)

	req := autonomy.CorrectMemoryRequest{
		CorrectedContent: "Customer prefers SMS notifications only between 9 AM and 4 PM PST.",
		Reason:           "Customer requested contact window change directly via account manager",
	}

	corrected, err := svc.CorrectMemory(ctx, orgID, 9001, userID, &req)
	require.NoError(t, err)
	require.NotNil(t, corrected)
	assert.Equal(t, int64(9001), corrected.ID)
}

func TestMemoryLearning_Invalidation(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	orgID := int64(1)
	userID := int64(42)

	req := autonomy.InvalidateMemoryRequest{
		Reason: "Carrier route discontinued by ocean operator",
	}

	err := svc.InvalidateMemory(ctx, orgID, 9001, userID, &req)
	assert.NoError(t, err)
}

func TestMemoryLearning_FlagUnreliable(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	orgID := int64(1)
	userID := int64(42)

	err := svc.FlagMemoryUnreliable(ctx, orgID, 9001, userID, "Observed contradictory outcome during recent sailing")
	assert.NoError(t, err)
}

func TestMemoryLearning_DetectAndSyncPatterns(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	orgID := int64(1)

	resp, err := svc.DetectAndSyncPatterns(ctx, orgID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 1, resp.TotalPatterns)
	assert.NotEmpty(t, resp.DetectedPatterns)
}

func TestMemoryLearning_SummaryTelemetry(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()

	orgID := int64(1)

	summary, err := svc.GetMemoryLearningSummary(ctx, orgID)
	require.NoError(t, err)
	require.NotNil(t, summary)

	assert.Equal(t, 15, summary.TotalMemories)
	assert.Equal(t, 12, summary.ActiveMemories)
	assert.Equal(t, 2, summary.StaleMemories)
	assert.Equal(t, 1, summary.InvalidatedMemories)
	assert.Equal(t, 20, summary.TotalOutcomes)
	assert.Equal(t, 18, summary.VerifiedOutcomes)
	assert.Equal(t, 16, summary.SuccessfulOutcomes)
	assert.Equal(t, 2, summary.FailedOutcomes)
	assert.Equal(t, 4, summary.DetectedPatterns)
	assert.Greater(t, summary.OverallSuccessRate, 0.8)
	assert.Greater(t, summary.RecommendationAcceptPct, 0.8)
}
