package monitoring

import (
	"context"
	"strings"
	"testing"
)

// mockRepo is an in-memory repository for unit testing the service layer.
type mockRepo struct {
	traces     []ExecutionTraceRecord
	evals      []QualityEvaluationRecord
	security   []SecurityEventRecord
	pricing    map[string]ModelPricingRecord
	thresholds map[string]HealthThresholdRecord
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		pricing:    make(map[string]ModelPricingRecord),
		thresholds: make(map[string]HealthThresholdRecord),
	}
}

func (m *mockRepo) RecordExecutionTrace(ctx context.Context, trace *ExecutionTraceRecord) (int64, error) {
	trace.ID = int64(len(m.traces) + 1)
	m.traces = append(m.traces, *trace)
	return trace.ID, nil
}

func (m *mockRepo) ListExecutionTraces(ctx context.Context, orgID int64, filter ListExecutionsFilter) ([]ExecutionTraceRecord, int64, error) {
	var filtered []ExecutionTraceRecord
	for _, t := range m.traces {
		if t.OrgID == orgID {
			filtered = append(filtered, t)
		}
	}
	return filtered, int64(len(filtered)), nil
}

func (m *mockRepo) GetExecutionTrace(ctx context.Context, orgID int64, id int64) (*ExecutionTraceRecord, error) {
	for _, t := range m.traces {
		if t.OrgID == orgID && t.ID == id {
			return &t, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) GetRuntimePerformance(ctx context.Context, orgID int64, days int) (*PerformanceMetrics, error) {
	var total, succ, fail int64
	for _, t := range m.traces {
		if t.OrgID == orgID {
			total++
			if t.Status == "COMPLETED" {
				succ++
			} else {
				fail++
			}
		}
	}
	rate := 100.0
	if total > 0 {
		rate = (float64(succ) / float64(total)) * 100.0
	}
	return &PerformanceMetrics{
		TotalRequests: total,
		SuccessCount:  succ,
		FailureCount:  fail,
		SuccessRate:   rate,
		Latency: LatencyPercentiles{
			P50Ms: 250,
			P95Ms: 450,
		},
	}, nil
}

func (m *mockRepo) GetCostMetrics(ctx context.Context, orgID int64, days int) (*CostSummary, error) {
	var cost float64
	var reqs, inT, outT int64
	for _, t := range m.traces {
		if t.OrgID == orgID {
			cost += t.EstimatedCost
			reqs++
			inT += int64(t.InputTokens)
			outT += int64(t.OutputTokens)
		}
	}
	return &CostSummary{
		TotalEstimatedCost: cost,
		Currency:           "USD",
		IsEstimated:        true,
		TotalRequests:      reqs,
		TotalInputTokens:   inT,
		TotalOutputTokens:  outT,
		TotalTokens:        inT + outT,
	}, nil
}

func (m *mockRepo) GetQueueWorkerMetrics(ctx context.Context, orgID int64) (*QueueWorkerMetrics, error) {
	return &QueueWorkerMetrics{
		QueuedJobs:         2,
		ProcessingJobs:     1,
		Completed24h:       25,
		Failed24h:          0,
		WorkerHealth:       HealthStateHealthy,
		OldestQueuedAgeSec: 15,
	}, nil
}

func (m *mockRepo) GetRecommendationMetrics(ctx context.Context, orgID int64) (*RecommendationMetrics, error) {
	return &RecommendationMetrics{
		TotalGenerated: 10,
		ActiveCount:    4,
		AcceptedCount:  5,
		DismissedCount: 1,
		AcceptanceRate: 50.0,
	}, nil
}

func (m *mockRepo) GetApprovalMetrics(ctx context.Context, orgID int64) (*ApprovalMetrics, error) {
	return &ApprovalMetrics{
		TotalProposed: 8,
		PendingCount:  2,
		ApprovedCount: 6,
	}, nil
}

func (m *mockRepo) GetMemoryMetrics(ctx context.Context, orgID int64) (*MemorySafetyMetrics, error) {
	return &MemorySafetyMetrics{
		ActivePersonalMemories: 3,
		ActiveOrgMemories:      2,
		PersonalizationEnabled: true,
	}, nil
}

func (m *mockRepo) RecordQualityEvaluation(ctx context.Context, eval *QualityEvaluationRecord) (int64, error) {
	eval.ID = int64(len(m.evals) + 1)
	m.evals = append(m.evals, *eval)
	return eval.ID, nil
}

func (m *mockRepo) ListQualityEvaluations(ctx context.Context, orgID int64, limit int) ([]QualityEvaluationRecord, error) {
	var res []QualityEvaluationRecord
	for _, e := range m.evals {
		if e.OrgID == orgID {
			res = append(res, e)
		}
	}
	return res, nil
}

func (m *mockRepo) RecordSecurityEvent(ctx context.Context, event *SecurityEventRecord) (int64, error) {
	event.ID = int64(len(m.security) + 1)
	m.security = append(m.security, *event)
	return event.ID, nil
}

func (m *mockRepo) ListSecurityEvents(ctx context.Context, orgID int64, limit int) ([]SecurityEventRecord, error) {
	var res []SecurityEventRecord
	for _, s := range m.security {
		if s.OrgID == orgID {
			res = append(res, s)
		}
	}
	return res, nil
}

func (m *mockRepo) GetModelPricing(ctx context.Context) (map[string]ModelPricingRecord, []ModelPricingRecord, error) {
	var list []ModelPricingRecord
	for _, v := range m.pricing {
		list = append(list, v)
	}
	return m.pricing, list, nil
}

func (m *mockRepo) UpsertModelPricing(ctx context.Context, p *ModelPricingRecord) error {
	key := strings.ToLower(p.Provider) + ":" + strings.ToLower(p.ModelName)
	m.pricing[key] = *p
	return nil
}

func (m *mockRepo) GetHealthThresholds(ctx context.Context, orgID int64) (map[string]HealthThresholdRecord, []HealthThresholdRecord, error) {
	var list []HealthThresholdRecord
	for _, v := range m.thresholds {
		list = append(list, v)
	}
	return m.thresholds, list, nil
}

func (m *mockRepo) UpsertHealthThreshold(ctx context.Context, t *HealthThresholdRecord) error {
	m.thresholds[t.ThresholdKey] = *t
	return nil
}

// ── TESTS ───────────────────────────────────────────────────────────────────

func TestPricingCalculation_StandardModels(t *testing.T) {
	// Gemini Flash: $0.000075 / 1k in, $0.000300 / 1k out
	cost, curr := CalculateCost("gemini", "gemini-1.5-flash", 1000, 1000, nil)
	expected := 0.000075 + 0.000300
	if curr != "USD" {
		t.Errorf("expected currency USD, got %s", curr)
	}
	diff := cost - expected
	if diff < 0 {
		diff = -diff
	}
	if diff > 1e-7 {
		t.Errorf("expected cost %f, got %f", expected, cost)
	}

	// Mock provider should be $0.00
	mockCost, _ := CalculateCost("mock", "mock-evaluator", 5000, 5000, nil)
	if mockCost != 0.0 {
		t.Errorf("expected mock cost 0, got %f", mockCost)
	}
}

func TestPricingCalculation_CustomDbRates(t *testing.T) {
	custom := map[string]ModelPricingRecord{
		"openai:gpt-4o-mini": {
			Provider:             "openai",
			ModelName:            "gpt-4o-mini",
			InputCostPer1kTokens: 0.000200,
			OutputCostPer1kTokens: 0.000800,
			Currency:             "USD",
			IsActive:             true,
		},
	}
	cost, _ := CalculateCost("openai", "gpt-4o-mini", 1000, 1000, custom)
	expected := 0.000200 + 0.000800
	diff := cost - expected
	if diff < 0 {
		diff = -diff
	}
	if diff > 1e-7 {
		t.Errorf("expected custom rate %f, got %f", expected, cost)
	}
}

func TestTokenEstimation(t *testing.T) {
	text := "This is an automated operational shipment summary."
	tokens := EstimateTokens(text)
	if tokens <= 0 {
		t.Errorf("expected positive token estimate, got %d", tokens)
	}
	if EstimateTokens("") != 0 {
		t.Errorf("expected 0 tokens for empty string")
	}
}

func TestHealthStateEvaluation_Normal(t *testing.T) {
	perf := PerformanceMetrics{
		TotalRequests: 100,
		SuccessCount:  98,
		FailureCount:  2,
		Latency:       LatencyPercentiles{P95Ms: 1200},
	}
	queue := QueueWorkerMetrics{QueuedJobs: 5, OldestQueuedAgeSec: 10}
	cost := CostSummary{TotalEstimatedCost: 12.50}
	sec := SecurityMetrics{TotalSecurityEvents24h: 0}

	status, alerts := EvaluateHealthState(perf, queue, cost, sec, nil)
	if status != HealthStateHealthy {
		t.Errorf("expected HEALTHY state, got %s", status)
	}
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts for healthy metrics, got %d", len(alerts))
	}
}

func TestHealthStateEvaluation_DegradedOnFailureSpike(t *testing.T) {
	perf := PerformanceMetrics{
		TotalRequests: 20,
		SuccessCount:  15,
		FailureCount:  5, // 25% failure rate, exceeds 10% critical threshold
		Latency:       LatencyPercentiles{P95Ms: 1000},
	}
	queue := QueueWorkerMetrics{QueuedJobs: 2}
	cost := CostSummary{TotalEstimatedCost: 5.0}
	sec := SecurityMetrics{}

	status, alerts := EvaluateHealthState(perf, queue, cost, sec, nil)
	if status != HealthStateDegraded {
		t.Errorf("expected DEGRADED state for high failure rate, got %s", status)
	}
	if len(alerts) == 0 {
		t.Errorf("expected critical failure rate alert, got none")
	}
}

func TestHealthStateEvaluation_LatencySpike(t *testing.T) {
	perf := PerformanceMetrics{
		TotalRequests: 50,
		SuccessCount:  50,
		FailureCount:  0,
		Latency:       LatencyPercentiles{P95Ms: 6500}, // Exceeds 5000ms max_p95_latency_ms
	}
	queue := QueueWorkerMetrics{QueuedJobs: 2}
	cost := CostSummary{}
	sec := SecurityMetrics{}

	status, alerts := EvaluateHealthState(perf, queue, cost, sec, nil)
	if status != HealthStateDegraded {
		t.Errorf("expected DEGRADED state for latency spike, got %s", status)
	}
	foundLatencyAlert := false
	for _, a := range alerts {
		if a.MetricKey == "p95_latency_ms" && a.Severity == "CRITICAL" {
			foundLatencyAlert = true
		}
	}
	if !foundLatencyAlert {
		t.Errorf("expected critical latency alert")
	}
}

func TestHealthStateEvaluation_QueueBacklogAlert(t *testing.T) {
	perf := PerformanceMetrics{TotalRequests: 10, SuccessCount: 10, Latency: LatencyPercentiles{P95Ms: 1000}}
	queue := QueueWorkerMetrics{QueuedJobs: 75} // Exceeds 50 backlog threshold
	cost := CostSummary{}
	sec := SecurityMetrics{}

	status, alerts := EvaluateHealthState(perf, queue, cost, sec, nil)
	if status != HealthStateDegraded {
		t.Errorf("expected DEGRADED state for queue backlog, got %s", status)
	}
	if len(alerts) == 0 {
		t.Errorf("expected queue backlog alert")
	}
}

func TestRedaction_SanitizesSensitiveCredentials(t *testing.T) {
	dirtyText := "Failed authorization with Bearer sk-ant-api03-abcdef123456789 and password=SuperSecretPassword123"
	clean := SanitizeDetails(dirtyText)

	if strings.Contains(clean, "sk-ant-api03-abcdef123456789") {
		t.Errorf("sensitive token leaked in sanitized details: %s", clean)
	}
	if strings.Contains(clean, "SuperSecretPassword123") {
		t.Errorf("sensitive password leaked in sanitized details: %s", clean)
	}
	if !strings.Contains(clean, "[REDACTED_") {
		t.Errorf("expected redaction tokens in output: %s", clean)
	}
}

func TestQualityEvaluation_ScoringAndGrounding(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	eval := &QualityEvaluationRecord{
		Feature:         "contract_compliance",
		EvaluationType:  "GROUNDING",
		Score:           0.95,
		GroundingStatus: "VALID",
		EvidenceStatus:  "SUFFICIENT",
		SafetyStatus:    "SAFE",
	}

	res, err := svc.EvaluateTraceQuality(context.Background(), 1, eval)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID <= 0 {
		t.Errorf("expected positive record ID, got %d", res.ID)
	}
	if res.PassStatus != "PASSED" {
		t.Errorf("expected PASSED for score 0.95, got %s", res.PassStatus)
	}

	summary, err := svc.GetQualitySummary(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error fetching quality summary: %v", err)
	}
	if summary.TotalEvaluations != 1 {
		t.Errorf("expected 1 evaluation in summary, got %d", summary.TotalEvaluations)
	}
	if summary.OverallPassRate != 100.0 {
		t.Errorf("expected 100%% pass rate, got %f", summary.OverallPassRate)
	}
}

func TestAdminPermissions_PricingAndThresholds(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	// Non-admin attempt should be rejected
	err := svc.UpdateModelPricing(context.Background(), "OPERATOR", &ModelPricingRecord{
		Provider:  "gemini",
		ModelName: "gemini-1.5-flash",
	})
	if err == nil {
		t.Errorf("expected error for non-admin role, got nil")
	}

	// Admin attempt should succeed
	err = svc.UpdateModelPricing(context.Background(), "ADMIN", &ModelPricingRecord{
		Provider:  "gemini",
		ModelName: "gemini-1.5-flash",
	})
	if err != nil {
		t.Errorf("expected admin update to succeed, got %v", err)
	}

	// Threshold non-admin rejection
	err = svc.UpdateHealthThreshold(context.Background(), 1, "VIEWER", &HealthThresholdRecord{
		ThresholdKey:   "max_p95_latency_ms",
		ThresholdValue: 4000,
	})
	if err == nil {
		t.Errorf("expected error for viewer role, got nil")
	}
}
