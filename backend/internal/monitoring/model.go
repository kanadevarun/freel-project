package monitoring

import (
	"time"
)

// SystemHealthState represents deterministic operational health states.
type SystemHealthState string

const (
	HealthStateHealthy       SystemHealthState = "HEALTHY"
	HealthStateDegraded      SystemHealthState = "DEGRADED"
	HealthStateUnavailable   SystemHealthState = "UNAVAILABLE"
	HealthStateMisconfigured SystemHealthState = "MISCONFIGURED"
)

// ExecutionTraceRecord maps to MariaDB ai_execution_traces table.
type ExecutionTraceRecord struct {
	ID                int64      `db:"id" json:"id"`
	OrgID             int64      `db:"org_id" json:"org_id"`
	UserID            *int64     `db:"user_id" json:"user_id,omitempty"`
	TaskID            *int64     `db:"task_id" json:"task_id,omitempty"`
	ThreadID          *string    `db:"thread_id" json:"thread_id,omitempty"`
	RequestID         string     `db:"request_id" json:"request_id"`
	CorrelationID     *string    `db:"correlation_id" json:"correlation_id,omitempty"`
	WorkflowName      string     `db:"workflow_name" json:"workflow_name"`
	Feature           string     `db:"feature" json:"feature"`
	Assistant         string     `db:"assistant" json:"assistant"`
	Module            string     `db:"module" json:"module"`
	RequestType       string     `db:"request_type" json:"request_type"`
	RuntimeRoute      *string    `db:"runtime_route" json:"runtime_route,omitempty"`
	PromptKey         *string    `db:"prompt_key" json:"prompt_key,omitempty"`
	PromptVersion     *string    `db:"prompt_version" json:"prompt_version,omitempty"`
	PrimaryProvider   string     `db:"primary_provider" json:"primary_provider"`
	PrimaryModel      string     `db:"primary_model" json:"primary_model"`
	FinalProvider     string     `db:"final_provider" json:"final_provider"`
	FinalModel        string     `db:"final_model" json:"final_model"`
	FailoverOccurred  bool       `db:"failover_occurred" json:"failover_occurred"`
	FailoverReason    *string    `db:"failover_reason" json:"failover_reason,omitempty"`
	IsMock            bool       `db:"is_mock" json:"is_mock"`
	Status            string     `db:"status" json:"status"`
	DurationMs        *int64     `db:"duration_ms" json:"duration_ms,omitempty"`
	InputTokens       int        `db:"input_tokens" json:"input_tokens"`
	OutputTokens      int        `db:"output_tokens" json:"output_tokens"`
	TotalTokens       int        `db:"total_tokens" json:"total_tokens"`
	EstimatedCost     float64    `db:"estimated_cost" json:"estimated_cost"`
	CostCurrency      string     `db:"cost_currency" json:"cost_currency"`
	RetryCount        int        `db:"retry_count" json:"retry_count"`
	SafetyStatus      string     `db:"safety_status" json:"safety_status"`
	GroundingStatus   string     `db:"grounding_status" json:"grounding_status"`
	IsMemoryAssisted  bool       `db:"is_memory_assisted" json:"is_memory_assisted"`
	ErrorCategory     *string    `db:"error_category" json:"error_category,omitempty"`
	ErrorMessage      *string    `db:"error_message" json:"error_message,omitempty"`
	ExecutionMetadata *string    `db:"execution_metadata" json:"execution_metadata,omitempty"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	CompletedAt       *time.Time `db:"completed_at" json:"completed_at,omitempty"`
}

// ModelPricingRecord maps to MariaDB ai_model_pricing table.
type ModelPricingRecord struct {
	ID                   int64     `db:"id" json:"id"`
	Provider             string    `db:"provider" json:"provider"`
	ModelName            string    `db:"model_name" json:"model_name"`
	InputCostPer1kTokens float64   `db:"input_cost_per_1k_tokens" json:"input_cost_per_1k_tokens"`
	OutputCostPer1kTokens float64  `db:"output_cost_per_1k_tokens" json:"output_cost_per_1k_tokens"`
	Currency             string    `db:"currency" json:"currency"`
	IsActive             bool      `db:"is_active" json:"is_active"`
	EffectiveFrom        time.Time `db:"effective_from" json:"effective_from"`
	CreatedAt            time.Time `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time `db:"updated_at" json:"updated_at"`
}

// QualityEvaluationRecord maps to MariaDB ai_quality_evaluations table.
type QualityEvaluationRecord struct {
	ID              int64     `db:"id" json:"id"`
	OrgID           int64     `db:"org_id" json:"org_id"`
	ExecutionID     *int64    `db:"execution_id" json:"execution_id,omitempty"`
	Feature         string    `db:"feature" json:"feature"`
	EvaluationType  string    `db:"evaluation_type" json:"evaluation_type"` // GROUNDING, FACTUALITY, COMPLETENESS, SAFETY
	Score           float64   `db:"score" json:"score"`
	PassStatus      string    `db:"pass_status" json:"pass_status"`         // PASSED, FAILED, WARNING
	EvidenceStatus  string    `db:"evidence_status" json:"evidence_status"` // SUFFICIENT, INSUFFICIENT, MISSING
	SafetyStatus    string    `db:"safety_status" json:"safety_status"`     // SAFE, SENSITIVE_REDACTED, INJECTION_BLOCKED
	GroundingStatus string    `db:"grounding_status" json:"grounding_status"` // VALID, UNVERIFIED, FABRICATED_CLAIM
	ReviewerType    string    `db:"reviewer_type" json:"reviewer_type"`     // AUTOMATED_CHECK, HUMAN_SUPERVISOR
	ReviewerID      *int64    `db:"reviewer_id" json:"reviewer_id,omitempty"`
	EvaluationNotes *string   `db:"evaluation_notes" json:"evaluation_notes,omitempty"`
	CorrelationID   *string   `db:"correlation_id" json:"correlation_id,omitempty"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

// HealthThresholdRecord maps to MariaDB ai_health_thresholds table.
type HealthThresholdRecord struct {
	ID             int64     `db:"id" json:"id"`
	OrgID          int64     `db:"org_id" json:"org_id"`
	ThresholdKey   string    `db:"threshold_key" json:"threshold_key"`
	ThresholdValue float64   `db:"threshold_value" json:"threshold_value"`
	Unit           string    `db:"unit" json:"unit"`
	Severity       string    `db:"severity" json:"severity"`
	Description    *string   `db:"description" json:"description,omitempty"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// SecurityEventRecord maps to MariaDB ai_security_events table.
type SecurityEventRecord struct {
	ID               int64     `db:"id" json:"id"`
	OrgID            int64     `db:"org_id" json:"org_id"`
	EventType        string    `db:"event_type" json:"event_type"`
	Severity         string    `db:"severity" json:"severity"`
	ActorType        string    `db:"actor_type" json:"actor_type"`
	ActorID          *int64    `db:"actor_id" json:"actor_id,omitempty"`
	ResourceType     *string   `db:"resource_type" json:"resource_type,omitempty"`
	ResourceID       *string   `db:"resource_id" json:"resource_id,omitempty"`
	SanitizedDetails *string   `db:"sanitized_details" json:"sanitized_details,omitempty"`
	CorrelationID    *string   `db:"correlation_id" json:"correlation_id,omitempty"`
	IPAddress        *string   `db:"ip_address" json:"ip_address,omitempty"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

// --- DTOs for Monitoring Endpoints ---

// SubsystemHealth represents health for a specific component.
type SubsystemHealth struct {
	Name        string            `json:"name"`
	Status      SystemHealthState `json:"status"`
	Message     string            `json:"message"`
	LastChecked time.Time         `json:"last_checked"`
	Metrics     map[string]any    `json:"metrics,omitempty"`
}

// OperationalAlert represents an active health or safety breach.
type OperationalAlert struct {
	ID             string    `json:"id"`
	Severity       string    `json:"severity"` // WARNING, CRITICAL
	Subsystem      string    `json:"subsystem"`
	Title          string    `json:"title"`
	Message        string    `json:"message"`
	MetricKey      string    `json:"metric_key"`
	CurrentValue   float64   `json:"current_value"`
	ThresholdValue float64   `json:"threshold_value"`
	Unit           string    `json:"unit"`
	TriggeredAt    time.Time `json:"triggered_at"`
}

// AIHealthSummary provides an overall health status and operational snapshot.
type AIHealthSummary struct {
	OverallStatus       SystemHealthState          `json:"overall_status"`
	Subsystems          map[string]SubsystemHealth `json:"subsystems"`
	ActiveAlerts        []OperationalAlert         `json:"active_alerts"`
	TotalRequests24h    int64                      `json:"total_requests_24h"`
	SuccessRate24h      float64                    `json:"success_rate_24h"`
	P95LatencyMs24h     int64                      `json:"p95_latency_ms_24h"`
	QueueBacklogCount   int64                      `json:"queue_backlog_count"`
	EstimatedCost24hUSD float64                    `json:"estimated_cost_24h_usd"`
	SecurityEvents24h   int64                      `json:"security_events_24h"`
	Failures24h         int64                      `json:"failures_24h"`
	EvaluatedAt         time.Time                  `json:"evaluated_at"`
}

// CostBreakdownItem represents cost grouped by a specific dimension.
type CostBreakdownItem struct {
	DimensionName  string  `json:"dimension_name"`
	Provider       string  `json:"provider,omitempty"`
	ModelName      string  `json:"model_name,omitempty"`
	TotalRequests  int64   `json:"total_requests"`
	InputTokens    int64   `json:"input_tokens"`
	OutputTokens   int64   `json:"output_tokens"`
	TotalTokens    int64   `json:"total_tokens"`
	EstimatedCost  float64 `json:"estimated_cost"`
	Currency       string  `json:"currency"`
	PercentageCost float64 `json:"percentage_cost"`
}

// CostDailyTrendItem represents daily cost tracking.
type CostDailyTrendItem struct {
	Date          string  `json:"date"`
	TotalRequests int64   `json:"total_requests"`
	TotalTokens   int64   `json:"total_tokens"`
	EstimatedCost float64 `json:"estimated_cost"`
	Currency      string  `json:"currency"`
}

// CostSummary provides aggregated cost reporting across models, providers, and features.
type CostSummary struct {
	TotalEstimatedCost float64              `json:"total_estimated_cost"`
	Currency           string               `json:"currency"`
	IsEstimated        bool                 `json:"is_estimated"`
	CostDisclaimer     string               `json:"cost_disclaimer"`
	TotalRequests      int64                `json:"total_requests"`
	TotalInputTokens   int64                `json:"total_input_tokens"`
	TotalOutputTokens  int64                `json:"total_output_tokens"`
	TotalTokens        int64                `json:"total_tokens"`
	ByModel            []CostBreakdownItem  `json:"by_model"`
	ByProvider         []CostBreakdownItem  `json:"by_provider"`
	ByFeature          []CostBreakdownItem  `json:"by_feature"`
	ByAssistant        []CostBreakdownItem  `json:"by_assistant"`
	DailyTrend         []CostDailyTrendItem `json:"daily_trend"`
	TimeWindowDays     int                  `json:"time_window_days"`
}

// LatencyPercentiles aggregates latency metrics.
type LatencyPercentiles struct {
	AverageMs int64 `json:"average_ms"`
	P50Ms     int64 `json:"p50_ms"`
	P95Ms     int64 `json:"p95_ms"`
	P99Ms     int64 `json:"p99_ms"`
	MinMs     int64 `json:"min_ms"`
	MaxMs     int64 `json:"max_ms"`
}

// PerformanceMetrics represents runtime request success, failures, and latency distributions.
type PerformanceMetrics struct {
	TotalRequests    int64               `json:"total_requests"`
	SuccessCount     int64               `json:"success_count"`
	FailureCount     int64               `json:"failure_count"`
	TimeoutCount     int64               `json:"timeout_count"`
	RetryCount       int64               `json:"retry_count"`
	SuccessRate      float64             `json:"success_rate"`
	FailoverCount    int64               `json:"failover_count"`
	MockCount        int64               `json:"mock_count"`
	Latency          LatencyPercentiles  `json:"latency"`
	ByFeatureLatency map[string]int64    `json:"by_feature_latency"`
	ByStatusCount    map[string]int64    `json:"by_status_count"`
	TimeWindowDays   int                 `json:"time_window_days"`
}

// QualitySummary aggregates quality evaluation results.
type QualitySummary struct {
	TotalEvaluations        int64                     `json:"total_evaluations"`
	PassCount               int64                     `json:"pass_count"`
	FailCount               int64                     `json:"fail_count"`
	WarningCount            int64                     `json:"warning_count"`
	OverallPassRate         float64                   `json:"overall_pass_rate"`
	AverageScore            float64                   `json:"average_score"`
	GroundingValidRate      float64                   `json:"grounding_valid_rate"`
	EvidenceSufficientRate  float64                   `json:"evidence_sufficient_rate"`
	SafetyPassRate          float64                   `json:"safety_pass_rate"`
	RecentEvaluations       []QualityEvaluationRecord `json:"recent_evaluations"`
	AutomatedCheckLabel     string                    `json:"automated_check_label"`
}

// QueueWorkerMetrics tracks AI background task queue status.
type QueueWorkerMetrics struct {
	QueuedJobs         int64            `json:"queued_jobs"`
	ProcessingJobs     int64            `json:"processing_jobs"`
	Completed24h       int64            `json:"completed_24h"`
	Failed24h          int64            `json:"failed_24h"`
	RetryingJobs       int64            `json:"retrying_jobs"`
	DeadLetterJobs     int64            `json:"dead_letter_jobs"`
	OldestQueuedAgeSec int64            `json:"oldest_queued_age_sec"`
	AverageDurationMs  int64            `json:"average_duration_ms"`
	WorkerHealth       SystemHealthState `json:"worker_health"`
	FailureCategories  map[string]int64 `json:"failure_categories"`
}

// RecommendationMetrics tracks Recommendation Center health and conversion rates.
type RecommendationMetrics struct {
	TotalGenerated       int64            `json:"total_generated"`
	ActiveCount          int64            `json:"active_count"`
	AcceptedCount        int64            `json:"accepted_count"`
	DismissedCount       int64            `json:"dismissed_count"`
	CompletedCount       int64            `json:"completed_count"`
	ExpiredCount         int64            `json:"expired_count"`
	RequiringApproval    int64            `json:"requiring_approval"`
	AcceptanceRate       float64          `json:"acceptance_rate"`
	DismissalRate        float64          `json:"dismissal_rate"`
	AvgResolutionTimeMin float64          `json:"avg_resolution_time_min"`
	ByPriority           map[string]int64 `json:"by_priority"`
	ByCategory           map[string]int64 `json:"by_category"`
}

// ApprovalMetrics tracks Action and Human-in-the-Loop throughput.
type ApprovalMetrics struct {
	TotalProposed        int64   `json:"total_proposed"`
	PendingCount         int64   `json:"pending_count"`
	ApprovedCount        int64   `json:"approved_count"`
	RejectedCount        int64   `json:"rejected_count"`
	ReturnedCount        int64   `json:"returned_count"`
	ExecutingCount       int64   `json:"executing_count"`
	CompletedCount       int64   `json:"completed_count"`
	FailedCount          int64   `json:"failed_count"`
	AvgApprovalTimeMin   float64 `json:"avg_approval_time_min"`
	ApprovalFailureRate  float64 `json:"approval_failure_rate"`
	ExecutionFailureRate float64 `json:"execution_failure_rate"`
}

// MemorySafetyMetrics tracks safe telemetry for AI Memory.
type MemorySafetyMetrics struct {
	ActivePersonalMemories    int64 `json:"active_personal_memories"`
	ActiveOrgMemories         int64 `json:"active_org_memories"`
	ProposalsCount            int64 `json:"proposals_count"`
	ConfirmationsCount        int64 `json:"confirmations_count"`
	DeletionsCount            int64 `json:"deletions_count"`
	SensitiveRejectionsCount  int64 `json:"sensitive_rejections_count"`
	UnauthorizedAttemptsCount int64 `json:"unauthorized_attempts_count"`
	PersonalizationEnabled    bool  `json:"personalization_enabled"`
}

// SecurityMetrics tracks safety and security indicators.
type SecurityMetrics struct {
	TotalSecurityEvents24h int64                 `json:"total_security_events_24h"`
	ByEventType            map[string]int64      `json:"by_event_type"`
	BySeverity             map[string]int64      `json:"by_severity"`
	RecentEvents           []SecurityEventRecord `json:"recent_events"`
}

// ListExecutionsFilter defines query filters for execution trace listings.
type ListExecutionsFilter struct {
	Feature       string `json:"feature"`
	Assistant     string `json:"assistant"`
	ModelProvider string `json:"model_provider"`
	Status        string `json:"status"`
	CorrelationID string `json:"correlation_id"`
	Days          int    `json:"days"`
	Limit         int    `json:"limit"`
	Offset        int    `json:"offset"`
}
