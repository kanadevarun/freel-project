package ai

import (
	"time"
)

// ExecutionContext standardizes context parameters across all AI execution paths.
type ExecutionContext struct {
	OrgID         int64                  `json:"org_id"`
	ActingUserID  *int64                 `json:"acting_user_id,omitempty"`
	ActorType     string                 `json:"actor_type"` // AI_AGENT, USER, SYSTEM
	Source        string                 `json:"source"`     // API, WORKER, SCHEDULER, etc.
	TaskID        *int64                 `json:"task_id,omitempty"`
	ThreadID      string                 `json:"thread_id,omitempty"`
	RequestID     string                 `json:"request_id"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	WorkflowName  string                 `json:"workflow_name"` // pricing, sales, operations, contracts, etc.
	PromptKey     string                 `json:"prompt_key,omitempty"`
	PromptVersion string                 `json:"prompt_version,omitempty"`
	ExecutionMode string                 `json:"execution_mode,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

// ExecutionResult records the telemetry and outcome of an AI run.
type ExecutionResult struct {
	RawOutput        string                 `json:"raw_output"`
	PrimaryProvider  string                 `json:"primary_provider"`
	PrimaryModel     string                 `json:"primary_model"`
	FinalProvider    string                 `json:"final_provider"`
	FinalModel       string                 `json:"final_model"`
	FailoverOccurred bool                   `json:"failover_occurred"`
	FailoverReason   string                 `json:"failover_reason,omitempty"`
	IsMock           bool                   `json:"is_mock"`
	DurationMs       int64                  `json:"duration_ms"`
	TotalTokens      int                    `json:"total_tokens,omitempty"`
	ErrorCategory    string                 `json:"error_category,omitempty"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	CompletedAt      time.Time              `json:"completed_at"`
}
