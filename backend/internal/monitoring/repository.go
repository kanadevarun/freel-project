package monitoring

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// Repository defines data access methods for AI monitoring and observability.
type Repository interface {
	RecordExecutionTrace(ctx context.Context, trace *ExecutionTraceRecord) (int64, error)
	ListExecutionTraces(ctx context.Context, orgID int64, filter ListExecutionsFilter) ([]ExecutionTraceRecord, int64, error)
	GetExecutionTrace(ctx context.Context, orgID int64, id int64) (*ExecutionTraceRecord, error)

	GetRuntimePerformance(ctx context.Context, orgID int64, days int) (*PerformanceMetrics, error)
	GetCostMetrics(ctx context.Context, orgID int64, days int) (*CostSummary, error)
	GetQueueWorkerMetrics(ctx context.Context, orgID int64) (*QueueWorkerMetrics, error)
	GetRecommendationMetrics(ctx context.Context, orgID int64) (*RecommendationMetrics, error)
	GetApprovalMetrics(ctx context.Context, orgID int64) (*ApprovalMetrics, error)
	GetMemoryMetrics(ctx context.Context, orgID int64) (*MemorySafetyMetrics, error)

	RecordQualityEvaluation(ctx context.Context, eval *QualityEvaluationRecord) (int64, error)
	ListQualityEvaluations(ctx context.Context, orgID int64, limit int) ([]QualityEvaluationRecord, error)

	RecordSecurityEvent(ctx context.Context, event *SecurityEventRecord) (int64, error)
	ListSecurityEvents(ctx context.Context, orgID int64, limit int) ([]SecurityEventRecord, error)

	GetModelPricing(ctx context.Context) (map[string]ModelPricingRecord, []ModelPricingRecord, error)
	UpsertModelPricing(ctx context.Context, pricing *ModelPricingRecord) error

	GetHealthThresholds(ctx context.Context, orgID int64) (map[string]HealthThresholdRecord, []HealthThresholdRecord, error)
	UpsertHealthThreshold(ctx context.Context, threshold *HealthThresholdRecord) error
}

type repository struct {
	db *sqlx.DB
}

// NewRepository creates a new persistent monitoring repository.
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) RecordExecutionTrace(ctx context.Context, trace *ExecutionTraceRecord) (int64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	query := `
		INSERT INTO ai_execution_traces (
			org_id, user_id, task_id, thread_id, request_id, correlation_id,
			workflow_name, feature, assistant, module, request_type, runtime_route,
			prompt_key, prompt_version, primary_provider, primary_model,
			final_provider, final_model, failover_occurred, failover_reason,
			is_mock, status, duration_ms, input_tokens, output_tokens, total_tokens,
			estimated_cost, cost_currency, retry_count, safety_status, grounding_status,
			is_memory_assisted, error_category, error_message, execution_metadata,
			created_at, completed_at
		) VALUES (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?
		)
	`
	now := time.Now()
	if trace.CreatedAt.IsZero() {
		trace.CreatedAt = now
	}

	var completedAt *time.Time
	if trace.CompletedAt != nil && !trace.CompletedAt.IsZero() {
		completedAt = trace.CompletedAt
	}

	res, err := r.db.ExecContext(ctx, query,
		trace.OrgID, trace.UserID, trace.TaskID, trace.ThreadID, trace.RequestID, trace.CorrelationID,
		trace.WorkflowName, trace.Feature, trace.Assistant, trace.Module, trace.RequestType, trace.RuntimeRoute,
		trace.PromptKey, trace.PromptVersion, trace.PrimaryProvider, trace.PrimaryModel,
		trace.FinalProvider, trace.FinalModel, trace.FailoverOccurred, trace.FailoverReason,
		trace.IsMock, trace.Status, trace.DurationMs, trace.InputTokens, trace.OutputTokens, trace.TotalTokens,
		trace.EstimatedCost, trace.CostCurrency, trace.RetryCount, trace.SafetyStatus, trace.GroundingStatus,
		trace.IsMemoryAssisted, trace.ErrorCategory, trace.ErrorMessage, trace.ExecutionMetadata,
		trace.CreatedAt, completedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert execution trace: %w", err)
	}
	return res.LastInsertId()
}

func (r *repository) ListExecutionTraces(ctx context.Context, orgID int64, filter ListExecutionsFilter) ([]ExecutionTraceRecord, int64, error) {
	if r.db == nil {
		return nil, 0, fmt.Errorf("database connection is nil")
	}

	days := filter.Days
	if days <= 0 || days > 90 {
		days = 7
	}
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	var conditions []string
	var args []interface{}

	conditions = append(conditions, "org_id = ?")
	args = append(args, orgID)

	conditions = append(conditions, "created_at >= NOW() - INTERVAL ? DAY")
	args = append(args, days)

	if filter.Feature != "" && filter.Feature != "ALL" {
		conditions = append(conditions, "feature = ?")
		args = append(args, filter.Feature)
	}
	if filter.Assistant != "" && filter.Assistant != "ALL" {
		conditions = append(conditions, "assistant = ?")
		args = append(args, filter.Assistant)
	}
	if filter.ModelProvider != "" && filter.ModelProvider != "ALL" {
		conditions = append(conditions, "final_provider = ?")
		args = append(args, filter.ModelProvider)
	}
	if filter.Status != "" && filter.Status != "ALL" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.CorrelationID != "" {
		conditions = append(conditions, "correlation_id = ?")
		args = append(args, filter.CorrelationID)
	}

	whereClause := strings.Join(conditions, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ai_execution_traces WHERE %s", whereClause)
	var totalCount int64
	if err := r.db.GetContext(ctx, &totalCount, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count execution traces: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, org_id, user_id, task_id, thread_id, request_id, correlation_id,
		       workflow_name, feature, assistant, module, request_type, runtime_route,
		       prompt_key, prompt_version, primary_provider, primary_model,
		       final_provider, final_model, failover_occurred, failover_reason,
		       is_mock, status, duration_ms, input_tokens, output_tokens, total_tokens,
		       estimated_cost, cost_currency, retry_count, safety_status, grounding_status,
		       is_memory_assisted, error_category, error_message, execution_metadata,
		       created_at, completed_at
		FROM ai_execution_traces
		WHERE %s
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	queryArgs := append(args, limit, offset)
	var traces []ExecutionTraceRecord
	if err := r.db.SelectContext(ctx, &traces, query, queryArgs...); err != nil {
		return nil, 0, fmt.Errorf("failed to select execution traces: %w", err)
	}
	if traces == nil {
		traces = make([]ExecutionTraceRecord, 0)
	}

	return traces, totalCount, nil
}

func (r *repository) GetExecutionTrace(ctx context.Context, orgID int64, id int64) (*ExecutionTraceRecord, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT id, org_id, user_id, task_id, thread_id, request_id, correlation_id,
		       workflow_name, feature, assistant, module, request_type, runtime_route,
		       prompt_key, prompt_version, primary_provider, primary_model,
		       final_provider, final_model, failover_occurred, failover_reason,
		       is_mock, status, duration_ms, input_tokens, output_tokens, total_tokens,
		       estimated_cost, cost_currency, retry_count, safety_status, grounding_status,
		       is_memory_assisted, error_category, error_message, execution_metadata,
		       created_at, completed_at
		FROM ai_execution_traces
		WHERE org_id = ? AND id = ?
	`
	var trace ExecutionTraceRecord
	if err := r.db.GetContext(ctx, &trace, query, orgID, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get execution trace: %w", err)
	}
	return &trace, nil
}

func (r *repository) GetRuntimePerformance(ctx context.Context, orgID int64, days int) (*PerformanceMetrics, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	if days <= 0 || days > 90 {
		days = 7
	}

	// 1. Overall counts
	type countRow struct {
		TotalCount    int64 `db:"total_count"`
		SuccessCount  int64 `db:"success_count"`
		FailureCount  int64 `db:"failure_count"`
		TimeoutCount  int64 `db:"timeout_count"`
		RetryCount    int64 `db:"retry_count"`
		FailoverCount int64 `db:"failover_count"`
		MockCount     int64 `db:"mock_count"`
	}
	var counts countRow
	countQ := `
		SELECT
			COUNT(*) AS total_count,
			COALESCE(SUM(CASE WHEN status = 'COMPLETED' THEN 1 ELSE 0 END), 0) AS success_count,
			COALESCE(SUM(CASE WHEN status = 'FAILED' THEN 1 ELSE 0 END), 0) AS failure_count,
			COALESCE(SUM(CASE WHEN status = 'TIMEOUT' THEN 1 ELSE 0 END), 0) AS timeout_count,
			COALESCE(SUM(retry_count), 0) AS retry_count,
			COALESCE(SUM(CASE WHEN failover_occurred = 1 THEN 1 ELSE 0 END), 0) AS failover_count,
			COALESCE(SUM(CASE WHEN is_mock = 1 THEN 1 ELSE 0 END), 0) AS mock_count
		FROM ai_execution_traces
		WHERE org_id = ? AND created_at >= NOW() - INTERVAL ? DAY
	`
	if err := r.db.GetContext(ctx, &counts, countQ, orgID, days); err != nil {
		return nil, fmt.Errorf("failed to query performance counts: %w", err)
	}

	metrics := &PerformanceMetrics{
		TotalRequests:    counts.TotalCount,
		SuccessCount:     counts.SuccessCount,
		FailureCount:     counts.FailureCount,
		TimeoutCount:     counts.TimeoutCount,
		RetryCount:       counts.RetryCount,
		FailoverCount:    counts.FailoverCount,
		MockCount:        counts.MockCount,
		TimeWindowDays:   days,
		ByFeatureLatency: make(map[string]int64),
		ByStatusCount:    make(map[string]int64),
	}
	if metrics.TotalRequests > 0 {
		metrics.SuccessRate = (float64(metrics.SuccessCount) / float64(metrics.TotalRequests)) * 100.0
	}

	// 2. Fetch completed durations for exact percentile computation
	durQ := `
		SELECT duration_ms
		FROM ai_execution_traces
		WHERE org_id = ? AND status = 'COMPLETED' AND duration_ms IS NOT NULL AND created_at >= NOW() - INTERVAL ? DAY
		ORDER BY duration_ms ASC
	`
	var durations []int64
	_ = r.db.SelectContext(ctx, &durations, durQ, orgID, days)

	if len(durations) > 0 {
		var sum int64
		for _, d := range durations {
			sum += d
		}
		metrics.Latency.AverageMs = sum / int64(len(durations))
		metrics.Latency.MinMs = durations[0]
		metrics.Latency.MaxMs = durations[len(durations)-1]

		p50Idx := int(float64(len(durations)) * 0.50)
		if p50Idx >= len(durations) {
			p50Idx = len(durations) - 1
		}
		metrics.Latency.P50Ms = durations[p50Idx]

		p95Idx := int(float64(len(durations)) * 0.95)
		if p95Idx >= len(durations) {
			p95Idx = len(durations) - 1
		}
		metrics.Latency.P95Ms = durations[p95Idx]

		p99Idx := int(float64(len(durations)) * 0.99)
		if p99Idx >= len(durations) {
			p99Idx = len(durations) - 1
		}
		metrics.Latency.P99Ms = durations[p99Idx]
	}

	// 3. Status breakdown
	type statusRow struct {
		Status string `db:"status"`
		Cnt    int64  `db:"cnt"`
	}
	var sRows []statusRow
	statusQ := `
		SELECT status, COUNT(*) AS cnt
		FROM ai_execution_traces
		WHERE org_id = ? AND created_at >= NOW() - INTERVAL ? DAY
		GROUP BY status
	`
	_ = r.db.SelectContext(ctx, &sRows, statusQ, orgID, days)
	for _, sr := range sRows {
		metrics.ByStatusCount[sr.Status] = sr.Cnt
	}

	// 4. By feature latency
	type featLatRow struct {
		Feature string  `db:"feature"`
		AvgDur  float64 `db:"avg_dur"`
	}
	var fRows []featLatRow
	featQ := `
		SELECT feature, COALESCE(AVG(duration_ms), 0) AS avg_dur
		FROM ai_execution_traces
		WHERE org_id = ? AND status = 'COMPLETED' AND duration_ms IS NOT NULL AND created_at >= NOW() - INTERVAL ? DAY
		GROUP BY feature
	`
	_ = r.db.SelectContext(ctx, &fRows, featQ, orgID, days)
	for _, fr := range fRows {
		metrics.ByFeatureLatency[fr.Feature] = int64(fr.AvgDur)
	}

	return metrics, nil
}

func (r *repository) GetCostMetrics(ctx context.Context, orgID int64, days int) (*CostSummary, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	if days <= 0 || days > 90 {
		days = 30
	}

	// 1. Overall cost and tokens
	type costTotals struct {
		TotalRequests int64   `db:"total_requests"`
		InputTokens   int64   `db:"input_tokens"`
		OutputTokens  int64   `db:"output_tokens"`
		TotalTokens   int64   `db:"total_tokens"`
		EstimatedCost float64 `db:"estimated_cost"`
	}
	var totals costTotals
	totalsQ := `
		SELECT
			COUNT(*) AS total_requests,
			COALESCE(SUM(input_tokens), 0) AS input_tokens,
			COALESCE(SUM(output_tokens), 0) AS output_tokens,
			COALESCE(SUM(total_tokens), 0) AS total_tokens,
			COALESCE(SUM(estimated_cost), 0.0) AS estimated_cost
		FROM ai_execution_traces
		WHERE org_id = ? AND created_at >= NOW() - INTERVAL ? DAY
	`
	if err := r.db.GetContext(ctx, &totals, totalsQ, orgID, days); err != nil {
		return nil, fmt.Errorf("failed to query cost totals: %w", err)
	}

	summary := &CostSummary{
		TotalEstimatedCost: totals.EstimatedCost,
		Currency:           "USD",
		IsEstimated:        true,
		CostDisclaimer:     "Estimated using published provider token pricing. Final billing is subject to direct provider rate agreements and discounts.",
		TotalRequests:      totals.TotalRequests,
		TotalInputTokens:   totals.InputTokens,
		TotalOutputTokens:  totals.OutputTokens,
		TotalTokens:        totals.TotalTokens,
		TimeWindowDays:     days,
		ByModel:            make([]CostBreakdownItem, 0),
		ByProvider:         make([]CostBreakdownItem, 0),
		ByFeature:          make([]CostBreakdownItem, 0),
		ByAssistant:        make([]CostBreakdownItem, 0),
		DailyTrend:         make([]CostDailyTrendItem, 0),
	}

	// 2. Cost by Model
	type groupRow struct {
		DimName   string  `db:"dim_name"`
		Provider  string  `db:"provider"`
		Requests  int64   `db:"requests"`
		InTokens  int64   `db:"in_tokens"`
		OutTokens int64   `db:"out_tokens"`
		TotTokens int64   `db:"tot_tokens"`
		Cost      float64 `db:"cost"`
	}

	var modelRows []groupRow
	modelQ := `
		SELECT
			final_model AS dim_name,
			final_provider AS provider,
			COUNT(*) AS requests,
			COALESCE(SUM(input_tokens), 0) AS in_tokens,
			COALESCE(SUM(output_tokens), 0) AS out_tokens,
			COALESCE(SUM(total_tokens), 0) AS tot_tokens,
			COALESCE(SUM(estimated_cost), 0.0) AS cost
		FROM ai_execution_traces
		WHERE org_id = ? AND created_at >= NOW() - INTERVAL ? DAY
		GROUP BY final_model, final_provider
		ORDER BY cost DESC
	`
	_ = r.db.SelectContext(ctx, &modelRows, modelQ, orgID, days)
	for _, m := range modelRows {
		pct := 0.0
		if summary.TotalEstimatedCost > 0 {
			pct = (m.Cost / summary.TotalEstimatedCost) * 100.0
		}
		summary.ByModel = append(summary.ByModel, CostBreakdownItem{
			DimensionName:  m.DimName,
			Provider:       m.Provider,
			ModelName:      m.DimName,
			TotalRequests:  m.Requests,
			InputTokens:    m.InTokens,
			OutputTokens:   m.OutTokens,
			TotalTokens:    m.TotTokens,
			EstimatedCost:  m.Cost,
			Currency:       "USD",
			PercentageCost: pct,
		})
	}

	// 3. Cost by Provider
	var provRows []groupRow
	provQ := `
		SELECT
			final_provider AS dim_name,
			final_provider AS provider,
			COUNT(*) AS requests,
			COALESCE(SUM(input_tokens), 0) AS in_tokens,
			COALESCE(SUM(output_tokens), 0) AS out_tokens,
			COALESCE(SUM(total_tokens), 0) AS tot_tokens,
			COALESCE(SUM(estimated_cost), 0.0) AS cost
		FROM ai_execution_traces
		WHERE org_id = ? AND created_at >= NOW() - INTERVAL ? DAY
		GROUP BY final_provider
		ORDER BY cost DESC
	`
	_ = r.db.SelectContext(ctx, &provRows, provQ, orgID, days)
	for _, p := range provRows {
		pct := 0.0
		if summary.TotalEstimatedCost > 0 {
			pct = (p.Cost / summary.TotalEstimatedCost) * 100.0
		}
		summary.ByProvider = append(summary.ByProvider, CostBreakdownItem{
			DimensionName:  p.DimName,
			Provider:       p.Provider,
			TotalRequests:  p.Requests,
			InputTokens:    p.InTokens,
			OutputTokens:   p.OutTokens,
			TotalTokens:    p.TotTokens,
			EstimatedCost:  p.Cost,
			Currency:       "USD",
			PercentageCost: pct,
		})
	}

	// 4. Cost by Feature
	var featRows []groupRow
	featQ := `
		SELECT
			feature AS dim_name,
			'' AS provider,
			COUNT(*) AS requests,
			COALESCE(SUM(input_tokens), 0) AS in_tokens,
			COALESCE(SUM(output_tokens), 0) AS out_tokens,
			COALESCE(SUM(total_tokens), 0) AS tot_tokens,
			COALESCE(SUM(estimated_cost), 0.0) AS cost
		FROM ai_execution_traces
		WHERE org_id = ? AND created_at >= NOW() - INTERVAL ? DAY
		GROUP BY feature
		ORDER BY cost DESC
	`
	_ = r.db.SelectContext(ctx, &featRows, featQ, orgID, days)
	for _, f := range featRows {
		pct := 0.0
		if summary.TotalEstimatedCost > 0 {
			pct = (f.Cost / summary.TotalEstimatedCost) * 100.0
		}
		summary.ByFeature = append(summary.ByFeature, CostBreakdownItem{
			DimensionName:  f.DimName,
			TotalRequests:  f.Requests,
			InputTokens:    f.InTokens,
			OutputTokens:   f.OutTokens,
			TotalTokens:    f.TotTokens,
			EstimatedCost:  f.Cost,
			Currency:       "USD",
			PercentageCost: pct,
		})
	}

	// 5. Cost by Assistant
	var asstRows []groupRow
	asstQ := `
		SELECT
			assistant AS dim_name,
			'' AS provider,
			COUNT(*) AS requests,
			COALESCE(SUM(input_tokens), 0) AS in_tokens,
			COALESCE(SUM(output_tokens), 0) AS out_tokens,
			COALESCE(SUM(total_tokens), 0) AS tot_tokens,
			COALESCE(SUM(estimated_cost), 0.0) AS cost
		FROM ai_execution_traces
		WHERE org_id = ? AND created_at >= NOW() - INTERVAL ? DAY
		GROUP BY assistant
		ORDER BY cost DESC
	`
	_ = r.db.SelectContext(ctx, &asstRows, asstQ, orgID, days)
	for _, a := range asstRows {
		pct := 0.0
		if summary.TotalEstimatedCost > 0 {
			pct = (a.Cost / summary.TotalEstimatedCost) * 100.0
		}
		summary.ByAssistant = append(summary.ByAssistant, CostBreakdownItem{
			DimensionName:  a.DimName,
			TotalRequests:  a.Requests,
			InputTokens:    a.InTokens,
			OutputTokens:   a.OutTokens,
			TotalTokens:    a.TotTokens,
			EstimatedCost:  a.Cost,
			Currency:       "USD",
			PercentageCost: pct,
		})
	}

	// 6. Daily trend
	type dayRow struct {
		D         string  `db:"d"`
		Requests  int64   `db:"requests"`
		TotTokens int64   `db:"tot_tokens"`
		Cost      float64 `db:"cost"`
	}
	var dayRows []dayRow
	dayQ := `
		SELECT
			DATE_FORMAT(created_at, '%Y-%m-%d') AS d,
			COUNT(*) AS requests,
			COALESCE(SUM(total_tokens), 0) AS tot_tokens,
			COALESCE(SUM(estimated_cost), 0.0) AS cost
		FROM ai_execution_traces
		WHERE org_id = ? AND created_at >= NOW() - INTERVAL ? DAY
		GROUP BY d
		ORDER BY d ASC
	`
	_ = r.db.SelectContext(ctx, &dayRows, dayQ, orgID, days)
	for _, dr := range dayRows {
		summary.DailyTrend = append(summary.DailyTrend, CostDailyTrendItem{
			Date:          dr.D,
			TotalRequests: dr.Requests,
			TotalTokens:   dr.TotTokens,
			EstimatedCost: dr.Cost,
			Currency:      "USD",
		})
	}

	return summary, nil
}

func (r *repository) GetQueueWorkerMetrics(ctx context.Context, orgID int64) (*QueueWorkerMetrics, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	type queueRow struct {
		QueuedJobs     int64 `db:"queued_jobs"`
		ProcessingJobs int64 `db:"processing_jobs"`
		RetryingJobs   int64 `db:"retrying_jobs"`
		DeadLetterJobs int64 `db:"dead_letter_jobs"`
		Completed24h   int64 `db:"completed_24h"`
		Failed24h      int64 `db:"failed_24h"`
	}
	var qr queueRow
	qQuery := `
		SELECT
			COALESCE(SUM(CASE WHEN status = 'QUEUED' THEN 1 ELSE 0 END), 0) AS queued_jobs,
			COALESCE(SUM(CASE WHEN status = 'PROCESSING' THEN 1 ELSE 0 END), 0) AS processing_jobs,
			COALESCE(SUM(CASE WHEN status = 'RETRYING' THEN 1 ELSE 0 END), 0) AS retrying_jobs,
			COALESCE(SUM(CASE WHEN status = 'DEAD_LETTER' OR (status = 'FAILED' AND retry_count >= max_retries) THEN 1 ELSE 0 END), 0) AS dead_letter_jobs,
			COALESCE(SUM(CASE WHEN status = 'COMPLETED' AND completed_at >= NOW() - INTERVAL 24 HOUR THEN 1 ELSE 0 END), 0) AS completed_24h,
			COALESCE(SUM(CASE WHEN status = 'FAILED' AND updated_at >= NOW() - INTERVAL 24 HOUR THEN 1 ELSE 0 END), 0) AS failed_24h
		FROM ai_processing_tasks
		WHERE org_id = ?
	`
	if err := r.db.GetContext(ctx, &qr, qQuery, orgID); err != nil {
		return nil, fmt.Errorf("failed to query queue metrics: %w", err)
	}

	// Oldest queued job age
	var oldestAgeSec sql.NullInt64
	ageQuery := `
		SELECT TIMESTAMPDIFF(SECOND, created_at, NOW())
		FROM ai_processing_tasks
		WHERE org_id = ? AND status = 'QUEUED'
		ORDER BY created_at ASC
		LIMIT 1
	`
	_ = r.db.GetContext(ctx, &oldestAgeSec, ageQuery, orgID)

	// Average duration of completed tasks
	var avgDur sql.NullFloat64
	durQuery := `
		SELECT AVG(TIMESTAMPDIFF(MILLISECOND, started_at, completed_at))
		FROM ai_processing_tasks
		WHERE org_id = ? AND status = 'COMPLETED' AND started_at IS NOT NULL AND completed_at IS NOT NULL AND completed_at >= NOW() - INTERVAL 24 HOUR
	`
	_ = r.db.GetContext(ctx, &avgDur, durQuery, orgID)

	// Failure categories
	type catRow struct {
		Category string `db:"cat"`
		Cnt      int64  `db:"cnt"`
	}
	var cats []catRow
	catQuery := `
		SELECT COALESCE(last_error_code, 'UNCLASSIFIED') AS cat, COUNT(*) AS cnt
		FROM ai_processing_tasks
		WHERE org_id = ? AND status = 'FAILED' AND updated_at >= NOW() - INTERVAL 24 HOUR
		GROUP BY cat
	`
	_ = r.db.SelectContext(ctx, &cats, catQuery, orgID)
	failMap := make(map[string]int64)
	for _, c := range cats {
		failMap[c.Category] = c.Cnt
	}

	health := HealthStateHealthy
	if qr.QueuedJobs > 30 || qr.DeadLetterJobs > 10 {
		health = HealthStateDegraded
	}

	return &QueueWorkerMetrics{
		QueuedJobs:         qr.QueuedJobs,
		ProcessingJobs:     qr.ProcessingJobs,
		Completed24h:       qr.Completed24h,
		Failed24h:          qr.Failed24h,
		RetryingJobs:       qr.RetryingJobs,
		DeadLetterJobs:     qr.DeadLetterJobs,
		OldestQueuedAgeSec: oldestAgeSec.Int64,
		AverageDurationMs:  int64(avgDur.Float64),
		WorkerHealth:       health,
		FailureCategories:  failMap,
	}, nil
}

func (r *repository) GetRecommendationMetrics(ctx context.Context, orgID int64) (*RecommendationMetrics, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	type recRow struct {
		TotalGenerated    int64 `db:"total_generated"`
		ActiveCount       int64 `db:"active_count"`
		AcceptedCount     int64 `db:"accepted_count"`
		DismissedCount    int64 `db:"dismissed_count"`
		CompletedCount    int64 `db:"completed_count"`
		ExpiredCount      int64 `db:"expired_count"`
		RequiringApproval int64 `db:"requiring_approval"`
	}
	var rr recRow
	query := `
		SELECT
			COUNT(*) AS total_generated,
			COALESCE(SUM(CASE WHEN status IN ('new', 'in_progress') THEN 1 ELSE 0 END), 0) AS active_count,
			COALESCE(SUM(CASE WHEN status IN ('accepted', 'completed') THEN 1 ELSE 0 END), 0) AS accepted_count,
			COALESCE(SUM(CASE WHEN status = 'dismissed' THEN 1 ELSE 0 END), 0) AS dismissed_count,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END), 0) AS completed_count,
			COALESCE(SUM(CASE WHEN status = 'expired' OR (expires_at IS NOT NULL AND expires_at < NOW()) THEN 1 ELSE 0 END), 0) AS expired_count,
			COALESCE(SUM(CASE WHEN requires_approval = 1 THEN 1 ELSE 0 END), 0) AS requiring_approval
		FROM ai_recommendations
		WHERE org_id = ?
	`
	if err := r.db.GetContext(ctx, &rr, query, orgID); err != nil {
		return nil, fmt.Errorf("failed to query recommendation metrics: %w", err)
	}

	// Avg resolution time in minutes
	var avgResMin sql.NullFloat64
	resQ := `
		SELECT AVG(TIMESTAMPDIFF(MINUTE, created_at, COALESCE(completed_at, dismissed_at)))
		FROM ai_recommendations
		WHERE org_id = ? AND (completed_at IS NOT NULL OR dismissed_at IS NOT NULL)
	`
	_ = r.db.GetContext(ctx, &avgResMin, resQ, orgID)

	// By priority
	type prioRow struct {
		Priority string `db:"prio"`
		Cnt      int64  `db:"cnt"`
	}
	var prios []prioRow
	pQ := `SELECT priority AS prio, COUNT(*) AS cnt FROM ai_recommendations WHERE org_id = ? GROUP BY priority`
	_ = r.db.SelectContext(ctx, &prios, pQ, orgID)
	byPriority := make(map[string]int64)
	for _, p := range prios {
		byPriority[p.Priority] = p.Cnt
	}

	// By category
	var cats []prioRow
	cQ := `SELECT category AS prio, COUNT(*) AS cnt FROM ai_recommendations WHERE org_id = ? GROUP BY category`
	_ = r.db.SelectContext(ctx, &cats, cQ, orgID)
	byCategory := make(map[string]int64)
	for _, c := range cats {
		byCategory[c.Priority] = c.Cnt
	}

	accRate := 0.0
	disRate := 0.0
	if rr.TotalGenerated > 0 {
		accRate = (float64(rr.AcceptedCount) / float64(rr.TotalGenerated)) * 100.0
		disRate = (float64(rr.DismissedCount) / float64(rr.TotalGenerated)) * 100.0
	}

	return &RecommendationMetrics{
		TotalGenerated:       rr.TotalGenerated,
		ActiveCount:          rr.ActiveCount,
		AcceptedCount:        rr.AcceptedCount,
		DismissedCount:       rr.DismissedCount,
		CompletedCount:       rr.CompletedCount,
		ExpiredCount:         rr.ExpiredCount,
		RequiringApproval:    rr.RequiringApproval,
		AcceptanceRate:       accRate,
		DismissalRate:        disRate,
		AvgResolutionTimeMin: avgResMin.Float64,
		ByPriority:           byPriority,
		ByCategory:           byCategory,
	}, nil
}

func (r *repository) GetApprovalMetrics(ctx context.Context, orgID int64) (*ApprovalMetrics, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	type appRow struct {
		TotalProposed  int64 `db:"total_proposed"`
		PendingCount   int64 `db:"pending_count"`
		ApprovedCount  int64 `db:"approved_count"`
		RejectedCount  int64 `db:"rejected_count"`
		ReturnedCount  int64 `db:"returned_count"`
		ExecutingCount int64 `db:"executing_count"`
		CompletedCount int64 `db:"completed_count"`
		FailedCount    int64 `db:"failed_count"`
	}
	var ar appRow
	query := `
		SELECT
			COUNT(*) AS total_proposed,
			COALESCE(SUM(CASE WHEN status = 'Pending' THEN 1 ELSE 0 END), 0) AS pending_count,
			COALESCE(SUM(CASE WHEN status = 'Approved' THEN 1 ELSE 0 END), 0) AS approved_count,
			COALESCE(SUM(CASE WHEN status = 'Rejected' THEN 1 ELSE 0 END), 0) AS rejected_count,
			COALESCE(SUM(CASE WHEN status = 'Returned' THEN 1 ELSE 0 END), 0) AS returned_count,
			COALESCE(SUM(CASE WHEN execution_status = 'EXECUTING' THEN 1 ELSE 0 END), 0) AS executing_count,
			COALESCE(SUM(CASE WHEN execution_status = 'COMPLETED' THEN 1 ELSE 0 END), 0) AS completed_count,
			COALESCE(SUM(CASE WHEN execution_status = 'FAILED' THEN 1 ELSE 0 END), 0) AS failed_count
		FROM approval_requests
		WHERE org_id = ?
	`
	if err := r.db.GetContext(ctx, &ar, query, orgID); err != nil {
		return nil, fmt.Errorf("failed to query approval metrics: %w", err)
	}

	var avgAppMin sql.NullFloat64
	durQ := `
		SELECT AVG(TIMESTAMPDIFF(MINUTE, created_at, approved_at))
		FROM approval_requests
		WHERE org_id = ? AND approved_at IS NOT NULL
	`
	_ = r.db.GetContext(ctx, &avgAppMin, durQ, orgID)

	appFailRate := 0.0
	execFailRate := 0.0
	if ar.TotalProposed > 0 {
		appFailRate = (float64(ar.RejectedCount) / float64(ar.TotalProposed)) * 100.0
	}
	totalExecuted := ar.CompletedCount + ar.FailedCount
	if totalExecuted > 0 {
		execFailRate = (float64(ar.FailedCount) / float64(totalExecuted)) * 100.0
	}

	return &ApprovalMetrics{
		TotalProposed:        ar.TotalProposed,
		PendingCount:         ar.PendingCount,
		ApprovedCount:        ar.ApprovedCount,
		RejectedCount:        ar.RejectedCount,
		ReturnedCount:        ar.ReturnedCount,
		ExecutingCount:       ar.ExecutingCount,
		CompletedCount:       ar.CompletedCount,
		FailedCount:          ar.FailedCount,
		AvgApprovalTimeMin:   avgAppMin.Float64,
		ApprovalFailureRate:  appFailRate,
		ExecutionFailureRate: execFailRate,
	}, nil
}

func (r *repository) GetMemoryMetrics(ctx context.Context, orgID int64) (*MemorySafetyMetrics, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// 1. Active memory counts
	type memRow struct {
		ActivePersonal int64 `db:"active_personal"`
		ActiveOrg      int64 `db:"active_org"`
	}
	var mr memRow
	memQ := `
		SELECT
			COALESCE(SUM(CASE WHEN scope = 'USER' AND status = 'ACTIVE' THEN 1 ELSE 0 END), 0) AS active_personal,
			COALESCE(SUM(CASE WHEN scope = 'ORGANIZATION' AND status = 'ACTIVE' THEN 1 ELSE 0 END), 0) AS active_org
		FROM ai_memory_items
		WHERE org_id = ?
	`
	_ = r.db.GetContext(ctx, &mr, memQ, orgID)

	// 2. Audit event counts
	type auditRow struct {
		Proposals    int64 `db:"proposals"`
		Confirmed    int64 `db:"confirmed"`
		Deleted      int64 `db:"deleted"`
		SensitiveRej int64 `db:"sensitive_rej"`
		Unauth       int64 `db:"unauth"`
	}
	var ar auditRow
	auditQ := `
		SELECT
			COALESCE(SUM(CASE WHEN event_type = 'PROPOSED' THEN 1 ELSE 0 END), 0) AS proposals,
			COALESCE(SUM(CASE WHEN event_type IN ('CONFIRMED', 'CREATED') THEN 1 ELSE 0 END), 0) AS confirmed,
			COALESCE(SUM(CASE WHEN event_type IN ('DELETED', 'CLEARED') THEN 1 ELSE 0 END), 0) AS deleted,
			COALESCE(SUM(CASE WHEN event_type = 'REJECTED_SENSITIVE' THEN 1 ELSE 0 END), 0) AS sensitive_rej,
			COALESCE(SUM(CASE WHEN event_type = 'BLOCKED_AUTH' THEN 1 ELSE 0 END), 0) AS unauth
		FROM ai_memory_audit_events
		WHERE org_id = ?
	`
	_ = r.db.GetContext(ctx, &ar, auditQ, orgID)

	return &MemorySafetyMetrics{
		ActivePersonalMemories:    mr.ActivePersonal,
		ActiveOrgMemories:         mr.ActiveOrg,
		ProposalsCount:            ar.Proposals,
		ConfirmationsCount:        ar.Confirmed,
		DeletionsCount:            ar.Deleted,
		SensitiveRejectionsCount:  ar.SensitiveRej,
		UnauthorizedAttemptsCount: ar.Unauth,
		PersonalizationEnabled:    true,
	}, nil
}

func (r *repository) RecordQualityEvaluation(ctx context.Context, eval *QualityEvaluationRecord) (int64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	query := `
		INSERT INTO ai_quality_evaluations (
			org_id, execution_id, feature, evaluation_type, score,
			pass_status, evidence_status, safety_status, grounding_status,
			reviewer_type, reviewer_id, evaluation_notes, correlation_id, created_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?
		)
	`
	if eval.CreatedAt.IsZero() {
		eval.CreatedAt = time.Now()
	}
	notes := ""
	if eval.EvaluationNotes != nil {
		notes = SanitizeDetails(*eval.EvaluationNotes)
	}

	res, err := r.db.ExecContext(ctx, query,
		eval.OrgID, eval.ExecutionID, eval.Feature, eval.EvaluationType, eval.Score,
		eval.PassStatus, eval.EvidenceStatus, eval.SafetyStatus, eval.GroundingStatus,
		eval.ReviewerType, eval.ReviewerID, notes, eval.CorrelationID, eval.CreatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert quality evaluation: %w", err)
	}
	return res.LastInsertId()
}

func (r *repository) ListQualityEvaluations(ctx context.Context, orgID int64, limit int) ([]QualityEvaluationRecord, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT id, org_id, execution_id, feature, evaluation_type, score,
		       pass_status, evidence_status, safety_status, grounding_status,
		       reviewer_type, reviewer_id, evaluation_notes, correlation_id, created_at
		FROM ai_quality_evaluations
		WHERE org_id = ?
		ORDER BY id DESC
		LIMIT ?
	`
	var evals []QualityEvaluationRecord
	if err := r.db.SelectContext(ctx, &evals, query, orgID, limit); err != nil {
		return nil, fmt.Errorf("failed to select quality evaluations: %w", err)
	}
	if evals == nil {
		evals = make([]QualityEvaluationRecord, 0)
	}
	return evals, nil
}

func (r *repository) RecordSecurityEvent(ctx context.Context, event *SecurityEventRecord) (int64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	query := `
		INSERT INTO ai_security_events (
			org_id, event_type, severity, actor_type, actor_id,
			resource_type, resource_id, sanitized_details, correlation_id,
			ip_address, created_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?
		)
	`
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	var details *string
	if event.SanitizedDetails != nil {
		san := SanitizeDetails(*event.SanitizedDetails)
		details = &san
	}

	res, err := r.db.ExecContext(ctx, query,
		event.OrgID, event.EventType, event.Severity, event.ActorType, event.ActorID,
		event.ResourceType, event.ResourceID, details, event.CorrelationID,
		event.IPAddress, event.CreatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert security event: %w", err)
	}
	return res.LastInsertId()
}

func (r *repository) ListSecurityEvents(ctx context.Context, orgID int64, limit int) ([]SecurityEventRecord, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT id, org_id, event_type, severity, actor_type, actor_id,
		       resource_type, resource_id, sanitized_details, correlation_id,
		       ip_address, created_at
		FROM ai_security_events
		WHERE org_id = ?
		ORDER BY id DESC
		LIMIT ?
	`
	var events []SecurityEventRecord
	if err := r.db.SelectContext(ctx, &events, query, orgID, limit); err != nil {
		return nil, fmt.Errorf("failed to select security events: %w", err)
	}
	if events == nil {
		events = make([]SecurityEventRecord, 0)
	}
	return events, nil
}

func (r *repository) GetModelPricing(ctx context.Context) (map[string]ModelPricingRecord, []ModelPricingRecord, error) {
	if r.db == nil {
		return nil, nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT id, provider, model_name, input_cost_per_1k_tokens, output_cost_per_1k_tokens,
		       currency, is_active, effective_from, created_at, updated_at
		FROM ai_model_pricing
		ORDER BY provider, model_name
	`
	var list []ModelPricingRecord
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, nil, fmt.Errorf("failed to select model pricing: %w", err)
	}

	rateMap := make(map[string]ModelPricingRecord)
	for _, rec := range list {
		key := strings.ToLower(rec.Provider) + ":" + strings.ToLower(rec.ModelName)
		rateMap[key] = rec
	}
	return rateMap, list, nil
}

func (r *repository) UpsertModelPricing(ctx context.Context, p *ModelPricingRecord) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	query := `
		INSERT INTO ai_model_pricing (
			provider, model_name, input_cost_per_1k_tokens, output_cost_per_1k_tokens,
			currency, is_active, effective_from
		) VALUES (?, ?, ?, ?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE
			input_cost_per_1k_tokens = VALUES(input_cost_per_1k_tokens),
			output_cost_per_1k_tokens = VALUES(output_cost_per_1k_tokens),
			currency = VALUES(currency),
			is_active = VALUES(is_active),
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query,
		p.Provider, p.ModelName, p.InputCostPer1kTokens, p.OutputCostPer1kTokens,
		p.Currency, p.IsActive,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert model pricing: %w", err)
	}
	return nil
}

func (r *repository) GetHealthThresholds(ctx context.Context, orgID int64) (map[string]HealthThresholdRecord, []HealthThresholdRecord, error) {
	if r.db == nil {
		return nil, nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT id, org_id, threshold_key, threshold_value, unit, severity, description, created_at, updated_at
		FROM ai_health_thresholds
		WHERE org_id = ? OR org_id = 0
		ORDER BY org_id DESC, threshold_key ASC
	`
	var list []HealthThresholdRecord
	if err := r.db.SelectContext(ctx, &list, query, orgID); err != nil {
		return nil, nil, fmt.Errorf("failed to select health thresholds: %w", err)
	}

	threshMap := make(map[string]HealthThresholdRecord)
	for _, t := range list {
		// Org-specific override takes precedence over global (org_id = 0)
		if _, exists := threshMap[t.ThresholdKey]; !exists {
			threshMap[t.ThresholdKey] = t
		}
	}

	// Sort list deterministically
	sort.Slice(list, func(i, j int) bool {
		return list[i].ThresholdKey < list[j].ThresholdKey
	})

	return threshMap, list, nil
}

func (r *repository) UpsertHealthThreshold(ctx context.Context, t *HealthThresholdRecord) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	query := `
		INSERT INTO ai_health_thresholds (
			org_id, threshold_key, threshold_value, unit, severity, description
		) VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			threshold_value = VALUES(threshold_value),
			unit = VALUES(unit),
			severity = VALUES(severity),
			description = VALUES(description),
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query,
		t.OrgID, t.ThresholdKey, t.ThresholdValue, t.Unit, t.Severity, t.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert health threshold: %w", err)
	}
	return nil
}
