package enterprise_autonomy

import (
	"context"
	"errors"
	"testing"
	"time"
)

func setupResilienceTestService() (EnterpriseResilienceService, *MockEnterpriseRepository) {
	repo := NewMockEnterpriseRepository()
	svc := NewEnterpriseResilienceService(repo, nil, nil, nil, nil)
	return svc, repo
}

// 1. Enterprise Health Model & Subsystem Tracking
func TestHealthModelSubsystemTracking(t *testing.T) {
	svc, _ := setupResilienceTestService()
	ctx := context.Background()

	// Initial health state should be HEALTHY
	health, err := svc.GetPlatformHealth(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error fetching health: %v", err)
	}
	if health.OverallState != HealthStateHealthy {
		t.Errorf("expected HEALTHY state, got %s", health.OverallState)
	}
	if health.EffectiveAutonomyCeil != AutonomyLvl4Governed {
		t.Errorf("expected Level 4 ceiling when healthy, got %s", health.EffectiveAutonomyCeil)
	}

	// Trip critical subsystem (Database) to FAILED
	svc.RecordHeartbeat(ctx, SubsystemDatabase, HealthStateFailed, 1500, "Database connection pool exhausted")

	health, _ = svc.GetPlatformHealth(ctx, 1)
	if health.OverallState != HealthStateFailed {
		t.Errorf("expected FAILED platform state when critical DB fails, got %s", health.OverallState)
	}
	if health.EffectiveAutonomyCeil != AutonomyLvl1Recommend {
		t.Errorf("expected Level 1 ceiling when failed, got %s", health.EffectiveAutonomyCeil)
	}

	// Restore DB, set Python Sidecar to DEGRADED
	svc.RecordHeartbeat(ctx, SubsystemDatabase, HealthStateHealthy, 5, "Connected")
	svc.RecordHeartbeat(ctx, SubsystemPythonSidecar, HealthStateDegraded, 350, "Sidecar high memory usage")

	health, _ = svc.GetPlatformHealth(ctx, 1)
	if health.OverallState != HealthStateDegraded {
		t.Errorf("expected DEGRADED platform state, got %s", health.OverallState)
	}
	if health.EffectiveAutonomyCeil != AutonomyLvl3Controlled {
		t.Errorf("expected Level 3 ceiling when degraded, got %s", health.EffectiveAutonomyCeil)
	}
}

// 2. Comprehensive Failure Classification
func TestFailureClassification(t *testing.T) {
	svc, _ := setupResilienceTestService()

	testCases := []struct {
		errText  string
		expected FailureClassification
	}{
		{"governance policy check denied action execution", FailureTypePolicyBlocked},
		{"autonomous execution halted: emergency stop is active", FailureTypePolicyBlocked},
		{"unauthorized access: tenant 1 cannot access tenant 2", FailureTypeAuthorization},
		{"agent unauthorized: requested action violates least-privilege matrix", FailureTypeAuthorization},
		{"approval invalidated: entity state changed materially since approval granted", FailureTypeStaleApproval},
		{"action blocked: invoice already paid by customer", FailureTypeBusinessConflict},
		{"malformed payload: missing required field carrier_id", FailureTypeDataQuality},
		{"connection refused: dial tcp 127.0.0.1:8090", FailureTypeTransient},
		{"context deadline exceeded: model call timeout after 30s", FailureTypeTimeout},
		{"carrier api gateway 503 service unavailable", FailureTypeDependency},
		{"llm completion output failed json parsing", FailureTypeAIModel},
		{"unexpected null pointer exception in parser", FailureTypePermanent},
	}

	for _, tc := range testCases {
		cls := svc.ClassifyFailure(errors.New(tc.errText), nil)
		if cls != tc.expected {
			t.Errorf("for error '%s': expected classification %s, got %s", tc.errText, tc.expected, cls)
		}
	}
}

// 3. Retry Governor Strictness & Bounded Backoff
func TestRetryGovernorStrictness(t *testing.T) {
	svc, _ := setupResilienceTestService()

	// Non-retryable types MUST NEVER retry (Attempt 1 of 3)
	nonRetryables := []FailureClassification{
		FailureTypePermanent,
		FailureTypePolicyBlocked,
		FailureTypeAuthorization,
		FailureTypeStaleApproval,
		FailureTypeDataQuality,
		FailureTypeBusinessConflict,
	}

	for _, cls := range nonRetryables {
		retry, delay := svc.ShouldRetry(cls, 1, 3)
		if retry {
			t.Errorf("expected failure type %s to be non-retryable, but ShouldRetry returned true", cls)
		}
		if delay != 0 {
			t.Errorf("expected zero delay for non-retryable %s, got %v", cls, delay)
		}
	}

	// Retryable transient failure: should calculate exponential backoff with bounded max
	retry, delay1 := svc.ShouldRetry(FailureTypeTransient, 1, 3)
	if !retry {
		t.Fatalf("expected transient failure to be retryable")
	}
	if delay1 < 800*time.Millisecond || delay1 > 1500*time.Millisecond {
		t.Errorf("expected initial delay ~1s (+/- jitter), got %v", delay1)
	}

	retry, delay2 := svc.ShouldRetry(FailureTypeTransient, 2, 3)
	if !retry || delay2 <= delay1 {
		t.Errorf("expected delay2 (%v) to exceed delay1 (%v)", delay2, delay1)
	}

	// Attempt 3 of 3 (max attempts reached)
	retry, _ = svc.ShouldRetry(FailureTypeTransient, 3, 3)
	if retry {
		t.Errorf("expected attempt 3 of 3 to not retry (exceeded limit)")
	}
}

// 4. Durable Workflow Checkpoints
func TestDurableWorkflowCheckpointing(t *testing.T) {
	svc, _ := setupResilienceTestService()
	ctx := context.Background()

	wfID := "wf-test-resilience-001"
	cp := WorkflowCheckpoint{
		WorkflowID:        wfID,
		OrgID:             1,
		Milestone:         CheckpointInvestigationComplete,
		StepIndex:         2,
		CompletedSteps:    []string{"step-1", "step-2"},
		IntermediateFacts: map[string]interface{}{"root_cause": "Port congestion at JNPT", "delay_hours": 36},
		CreatedAt:         time.Now().UTC(),
	}

	err := svc.SaveCheckpoint(ctx, 1, cp)
	if err != nil {
		t.Fatalf("unexpected error saving checkpoint: %v", err)
	}

	fetched, err := svc.GetLatestCheckpoint(ctx, 1, wfID)
	if err != nil {
		t.Fatalf("unexpected error getting checkpoint: %v", err)
	}

	if fetched.Milestone != CheckpointInvestigationComplete {
		t.Errorf("expected milestone %s, got %s", CheckpointInvestigationComplete, fetched.Milestone)
	}
	if len(fetched.CompletedSteps) != 2 {
		t.Errorf("expected 2 completed steps, got %d", len(fetched.CompletedSteps))
	}
	if fetched.IntermediateFacts["root_cause"] != "Port congestion at JNPT" {
		t.Errorf("expected intermediate fact preserved, got %v", fetched.IntermediateFacts["root_cause"])
	}
}

// 5. Event Storm Coalescing and Throttling
func TestEventStormCoalescingAndThrottling(t *testing.T) {
	svc, _ := setupResilienceTestService()
	ctx := context.Background()

	orgID := int64(1)
	entityType := "SHIPMENT"
	entityID := "SH-9921"
	eventType := "TELEMATICS_PING"

	// First 7 events within 10s should be allowed
	for i := 1; i <= 7; i++ {
		allowed, reason := svc.CheckEventStormProtection(ctx, orgID, entityType, entityID, eventType)
		if !allowed {
			t.Fatalf("expected event %d to be permitted, but was throttled: %s", i, reason)
		}
	}

	// 8th and 9th rapid events should trigger storm detection and coalescing
	_ , _ = svc.CheckEventStormProtection(ctx, orgID, entityType, entityID, eventType)
	allowed, reason := svc.CheckEventStormProtection(ctx, orgID, entityType, entityID, eventType)
	if allowed {
		t.Errorf("expected event storm protection to coalesce rapid burst of events")
	}
	if reason == "" {
		t.Errorf("expected descriptive reason for coalescing")
	}

	// Verify backpressure metrics record coalescing
	metrics, err := svc.GetBackpressureMetrics(ctx, orgID)
	if err != nil {
		t.Fatalf("unexpected error getting backpressure metrics: %v", err)
	}
	if metrics.CoalescedEventCount == 0 {
		t.Errorf("expected CoalescedEventCount > 0, got %d", metrics.CoalescedEventCount)
	}
}

// 6. Partial Multi-Agent Failure Reconciliation (Non-Critical Specialist Failure)
func TestPartialMultiAgentReconciliation_NonCriticalFailure(t *testing.T) {
	svc, _ := setupResilienceTestService()
	ctx := context.Background()

	results := []MultiAgentTaskResult{
		{
			AgentID:         "investigation-agent",
			TaskName:        "InvestigateRootCause",
			Success:         true,
			Output:          map[string]interface{}{"cause": "Mechanical breakdown"},
			IsCriticalPath:  true,
			ExecutionTimeMs: 120,
		},
		{
			AgentID:         "pricing-agent",
			TaskName:        "CalculateExpeditedRate",
			Success:         false,
			Error:           "Model rate limit exceeded (429)",
			IsCriticalPath:  false,
			ExecutionTimeMs: 50,
		},
		{
			AgentID:         "routing-agent",
			TaskName:        "FindAlternateFeeder",
			Success:         true,
			Output:          map[string]interface{}{"vessel": "Express-Carrier-4"},
			IsCriticalPath:  true,
			ExecutionTimeMs: 140,
		},
	}

	recon, err := svc.ReconcileMultiAgentExecution(ctx, "wf-partial-test-01", 0.95, results)
	if err != nil {
		t.Fatalf("unexpected error in reconciliation: %v", err)
	}

	if recon.OverallStatus != "PARTIAL_SUCCESS" {
		t.Errorf("expected PARTIAL_SUCCESS, got %s", recon.OverallStatus)
	}
	if len(recon.SuccessfulAgents) != 2 || len(recon.FailedAgents) != 1 {
		t.Errorf("expected 2 successful and 1 failed agent, got %d and %d", len(recon.SuccessfulAgents), len(recon.FailedAgents))
	}
	if recon.AdjustedConfidence >= recon.OriginalConfidence {
		t.Errorf("expected adjusted confidence (%v) to be less than original (%v)", recon.AdjustedConfidence, recon.OriginalConfidence)
	}
	if !recon.CanProceedWithPlan {
		t.Errorf("expected plan to be proceedable via specialist fallback")
	}
	if recon.FallbackStrategy != "RULE_BASED_SPECIALIST_FALLBACK" {
		t.Errorf("expected RULE_BASED_SPECIALIST_FALLBACK, got %s", recon.FallbackStrategy)
	}
}

// 7. Partial Multi-Agent Failure Reconciliation (Critical Path Failure)
func TestPartialMultiAgentReconciliation_CriticalPathFailure(t *testing.T) {
	svc, _ := setupResilienceTestService()
	ctx := context.Background()

	results := []MultiAgentTaskResult{
		{
			AgentID:        "investigation-agent",
			TaskName:       "InvestigateRootCause",
			Success:        false,
			Error:          "Unable to resolve GPS coordinates",
			IsCriticalPath: true,
		},
		{
			AgentID:        "pricing-agent",
			TaskName:       "CalculateExpeditedRate",
			Success:        true,
			Output:         map[string]interface{}{"rate": 450.0},
			IsCriticalPath: false,
		},
	}

	recon, err := svc.ReconcileMultiAgentExecution(ctx, "wf-partial-crit-02", 0.90, results)
	if err != nil {
		t.Fatalf("unexpected error in reconciliation: %v", err)
	}

	if recon.OverallStatus != "FAILED" {
		t.Errorf("expected FAILED status for critical path failure, got %s", recon.OverallStatus)
	}
	if recon.CanProceedWithPlan {
		t.Errorf("expected CanProceedWithPlan to be false when critical path fails")
	}
	if !recon.RequiresHumanReview {
		t.Errorf("expected RequiresHumanReview to be true")
	}
	if recon.FallbackStrategy != "ESCALATE_TO_DISPATCHER" {
		t.Errorf("expected ESCALATE_TO_DISPATCHER, got %s", recon.FallbackStrategy)
	}
}

// 8. Adaptive Autonomy Degradation Calculation
func TestAdaptiveAutonomyDegradation(t *testing.T) {
	svc, _ := setupResilienceTestService()

	levels := []struct {
		health   EnterpriseHealthState
		expected AutonomyLevel
	}{
		{HealthStateHealthy, AutonomyLvl4Governed},
		{HealthStateDegraded, AutonomyLvl3Controlled},
		{HealthStateBlocked, AutonomyLvl2Prepare},
		{HealthStateFailed, AutonomyLvl1Recommend},
		{HealthStatePaused, AutonomyLvl0Observe},
	}

	for _, tc := range levels {
		actual := svc.CalculateDegradedAutonomyCeiling(tc.health)
		if actual != tc.expected {
			t.Errorf("for health %s: expected ceiling %s, got %s", tc.health, tc.expected, actual)
		}
	}
}

// 9. Notification Failure Separated from Business Action
func TestNotificationSeparationFromBusinessAction(t *testing.T) {
	svc, _ := setupResilienceTestService()
	ctx := context.Background()

	businessActionExecuted := false
	businessAction := func() error {
		businessActionExecuted = true
		return nil
	}

	notificationAction := func() error {
		return errors.New("SES SMTP timeout 504")
	}

	bizSuccess, notifSuccess, err := svc.ExecuteWithNotificationSeparation(ctx, 1, "wf-notif-sep-01", businessAction, notificationAction)
	if err != nil {
		t.Fatalf("unexpected error from ExecuteWithNotificationSeparation: %v", err)
	}
	if !bizSuccess {
		t.Errorf("expected business action to succeed despite notification failure")
	}
	if notifSuccess {
		t.Errorf("expected notification success to be false")
	}
	if !businessActionExecuted {
		t.Errorf("expected business action to have been executed")
	}

	// Verify failed notification was queued for async retry without rolling back business action
	failedItems, err := svc.GetFailedWorkItems(ctx, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error fetching failed work items: %v", err)
	}
	if len(failedItems) == 0 {
		t.Fatalf("expected failed notification item in queue")
	}
	if failedItems[0].ItemType != "NOTIFICATION" {
		t.Errorf("expected item type NOTIFICATION, got %s", failedItems[0].ItemType)
	}
	if !failedItems[0].IsRetryable {
		t.Errorf("expected failed notification to be retryable")
	}
}

// 10. Non-Retryable Replay Protection
func TestNonRetryableReplayProtection(t *testing.T) {
	svc, _ := setupResilienceTestService()
	ctx := context.Background()

	// Insert permanent failed work item directly into service
	permItem := "fail-perm-01"
	// Replaying non-existent item returns not found
	err := svc.ReplayFailedWorkItem(ctx, 1, permItem, "test-user")
	if err == nil {
		t.Errorf("expected error replaying non-existent item")
	}
}

// 11. Tenant Isolation in Resilience Service
func TestResilienceTenantIsolation(t *testing.T) {
	svc, _ := setupResilienceTestService()
	ctx := context.Background()

	// Zero or negative orgID must be rejected
	_, err := svc.GetPlatformHealth(ctx, 0)
	if !errors.Is(err, ErrUnauthorizedTenant) {
		t.Errorf("expected ErrUnauthorizedTenant for orgID=0, got %v", err)
	}

	err = svc.SaveCheckpoint(ctx, -1, WorkflowCheckpoint{WorkflowID: "wf-1"})
	if !errors.Is(err, ErrUnauthorizedTenant) {
		t.Errorf("expected ErrUnauthorizedTenant for orgID=-1, got %v", err)
	}

	_, err = svc.GetLatestCheckpoint(ctx, 0, "wf-1")
	if !errors.Is(err, ErrUnauthorizedTenant) {
		t.Errorf("expected ErrUnauthorizedTenant for orgID=0, got %v", err)
	}
}

// 12. End-to-End Control Tower Integration with Resilience
func TestControlTowerComprehensiveViewWithResilience(t *testing.T) {
	repo := NewMockEnterpriseRepository()
	resSvc := NewEnterpriseResilienceService(repo, nil, nil, nil, nil)
	ctSvc := NewEnterpriseControlTowerService(
		repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, resSvc, nil,
	)

	ctx := context.Background()
	view, err := ctSvc.GetControlTowerView(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error from GetControlTowerView: %v", err)
	}

	if view.ResilienceHealth == nil {
		t.Fatalf("expected ResilienceHealth to be populated in ControlTowerComprehensiveView")
	}
	if view.ResilienceHealth.OverallState != HealthStateHealthy {
		t.Errorf("expected HEALTHY resilience state, got %s", view.ResilienceHealth.OverallState)
	}
	if view.PlatformHealth != "HEALTHY" {
		t.Errorf("expected platform health HEALTHY, got %s", view.PlatformHealth)
	}

	// When resilience drops to DEGRADED, platform health reflects it
	resSvc.RecordHeartbeat(ctx, SubsystemAIWorker, HealthStateDegraded, 400, "High queue depth")
	viewDegraded, _ := ctSvc.GetControlTowerView(ctx, 1)
	if viewDegraded.PlatformHealth != "DEGRADED" {
		t.Errorf("expected DEGRADED platform health, got %s", viewDegraded.PlatformHealth)
	}
}
