package enterprise_autonomy

import (
	"errors"
	"time"
)

// ---------------------------------------------------------------------
// Phase 7.11: Enterprise Autonomous Operations Optimization & Resilience
// ---------------------------------------------------------------------

// SubsystemIdentifier defines each operational layer tracked in the health model
type SubsystemIdentifier string

const (
	SubsystemGoBackend      SubsystemIdentifier = "GO_BACKEND"
	SubsystemPythonSidecar  SubsystemIdentifier = "PYTHON_SIDECAR"
	SubsystemAIWorker       SubsystemIdentifier = "AI_WORKER"
	SubsystemEventMesh      SubsystemIdentifier = "EVENT_MESH"
	SubsystemWorkflowEngine SubsystemIdentifier = "WORKFLOW_ENGINE"
	SubsystemAIAgents       SubsystemIdentifier = "AI_AGENTS"
	SubsystemActionSystem   SubsystemIdentifier = "ACTION_SYSTEM"
	SubsystemApprovals      SubsystemIdentifier = "APPROVALS"
	SubsystemDatabase       SubsystemIdentifier = "DATABASE"
)

// EnterpriseHealthState represents the discrete, understandable operational health
type EnterpriseHealthState string

const (
	HealthStateHealthy  EnterpriseHealthState = "HEALTHY"
	HealthStateDegraded EnterpriseHealthState = "DEGRADED"
	HealthStateBlocked  EnterpriseHealthState = "BLOCKED"
	HealthStateFailed   EnterpriseHealthState = "FAILED"
	HealthStatePaused   EnterpriseHealthState = "PAUSED"
)

// SubsystemHealth encapsulates real-time status and operational vitals of a subsystem
type SubsystemHealth struct {
	Subsystem       SubsystemIdentifier   `json:"subsystem"`
	State           EnterpriseHealthState `json:"state"`
	LatencyMs       int64                 `json:"latency_ms"`
	ActiveLoad      int                   `json:"active_load"`
	CapacityLimit   int                   `json:"capacity_limit"`
	ErrorCount      int                   `json:"error_count"`
	Message         string                `json:"message"`
	LastHeartbeatAt time.Time             `json:"last_heartbeat_at"`
	IsCritical      bool                  `json:"is_critical"`
}

// EnterprisePlatformHealthSummary provides the unified enterprise autonomous-health model
type EnterprisePlatformHealthSummary struct {
	OrgID                 int64                             `json:"org_id"`
	OverallState          EnterpriseHealthState             `json:"overall_state"`
	EffectiveAutonomyCeil AutonomyLevel                     `json:"effective_autonomy_ceil"`
	Subsystems            map[SubsystemIdentifier]SubsystemHealth `json:"subsystems"`
	StuckWorkflowCount    int                               `json:"stuck_workflow_count"`
	DeadLetterCount       int                               `json:"dead_letter_count"`
	EventBacklogDepth     int                               `json:"event_backlog_depth"`
	BackpressureActive    bool                              `json:"backpressure_active"`
	DegradationReason     string                            `json:"degradation_reason,omitempty"`
	CheckedAt             time.Time                         `json:"checked_at"`
}

// FailureClassification rigorously categorizes failures to govern retry behavior
type FailureClassification string

const (
	FailureTypeTransient        FailureClassification = "TRANSIENT"
	FailureTypePermanent        FailureClassification = "PERMANENT"
	FailureTypePolicyBlocked    FailureClassification = "POLICY_BLOCKED"
	FailureTypeAuthorization    FailureClassification = "AUTHORIZATION_FAILURE"
	FailureTypeDataQuality      FailureClassification = "DATA_QUALITY_FAILURE"
	FailureTypeDependency       FailureClassification = "DEPENDENCY_FAILURE"
	FailureTypeAIModel          FailureClassification = "AI_MODEL_FAILURE"
	FailureTypeBusinessConflict FailureClassification = "BUSINESS_STATE_CONFLICT"
	FailureTypeStaleApproval    FailureClassification = "STALE_APPROVAL"
	FailureTypeTimeout          FailureClassification = "TIMEOUT"
)

// Standard Resilience Errors
var (
	ErrNonRetryableFailure       = errors.New("failure is classified as permanent or policy-blocked and will not be retried")
	ErrMaxRetriesExceeded        = errors.New("maximum retry attempts exceeded for step")
	ErrWorkflowStuck             = errors.New("workflow heartbeat timeout: workflow marked as stuck")
	ErrBackpressureThrottled     = errors.New("event ingestion throttled: backpressure threshold exceeded")
	ErrAgentOverloaded           = errors.New("target specialist agent has reached maximum concurrency limit")
	ErrAdaptiveAutonomyRestricted = errors.New("action blocked: system health requires degraded autonomy level")
)

// CheckpointMilestone represents durable progress boundaries
type CheckpointMilestone string

const (
	CheckpointPlanningComplete      CheckpointMilestone = "PLANNING_COMPLETE"
	CheckpointInvestigationComplete CheckpointMilestone = "INVESTIGATION_COMPLETE"
	CheckpointApprovalRequested     CheckpointMilestone = "APPROVAL_REQUESTED"
	CheckpointActionExecuted        CheckpointMilestone = "ACTION_EXECUTED"
	CheckpointVerificationComplete  CheckpointMilestone = "VERIFICATION_COMPLETE"
)

// WorkflowCheckpoint encapsulates the durable recovery boundary of an autonomous workflow
type WorkflowCheckpoint struct {
	WorkflowID        string                 `json:"workflow_id"`
	OrgID             int64                  `json:"org_id"`
	Milestone         CheckpointMilestone    `json:"milestone"`
	StepIndex         int                    `json:"step_index"`
	CompletedSteps    []string               `json:"completed_steps"`
	IntermediateFacts map[string]interface{} `json:"intermediate_facts"`
	CreatedAt         time.Time              `json:"created_at"`
}

// StuckWorkflowDescriptor contains diagnostic detail for stuck workflow detection
type StuckWorkflowDescriptor struct {
	WorkflowID         string                  `json:"workflow_id"`
	OrgID              int64                   `json:"org_id"`
	WorkflowType       EnterpriseWorkflowType  `json:"workflow_type"`
	CurrentState       EnterpriseWorkflowState `json:"current_state"`
	CurrentStepID      string                  `json:"current_step_id"`
	StuckSince         time.Time               `json:"stuck_since"`
	InactivityDuration time.Duration           `json:"inactivity_duration"`
	SuspectedCause     string                  `json:"suspected_cause"`
	LastMilestone      CheckpointMilestone     `json:"last_milestone"`
	CanAutoRecover     bool                    `json:"can_auto_recover"`
	RecommendedAction  string                  `json:"recommended_action"`
}

// FailedWorkItem represents visible failed tasks, workflows, and dead letters
type FailedWorkItem struct {
	ID                    string                `json:"id"`
	OrgID                 int64                 `json:"org_id"`
	ItemType              string                `json:"item_type"` // WORKFLOW, STEP, EVENT, NOTIFICATION
	WorkflowID            string                `json:"workflow_id,omitempty"`
	EntityID              string                `json:"entity_id"`
	EntityType            string                `json:"entity_type"`
	AgentID               string                `json:"agent_id,omitempty"`
	FailureClassification FailureClassification `json:"failure_classification"`
	FailureReason         string                `json:"failure_reason"`
	RetryCount            int                   `json:"retry_count"`
	MaxRetries            int                   `json:"max_retries"`
	IsRetryable           bool                  `json:"is_retryable"`
	NextScheduledRetryAt  *time.Time            `json:"next_scheduled_retry_at,omitempty"`
	LastAttemptAt         time.Time             `json:"last_attempt_at"`
	NextAction            string                `json:"next_action"`
	EscalationState       string                `json:"escalation_state"` // NONE, ESCALATED_OPERATOR, BLOCKED
}

// MultiAgentTaskResult represents individual specialist execution within a multi-agent plan
type MultiAgentTaskResult struct {
	AgentID         string                 `json:"agent_id"`
	TaskName        string                 `json:"task_name"`
	Success         bool                   `json:"success"`
	Output          map[string]interface{} `json:"output,omitempty"`
	Error           string                 `json:"error,omitempty"`
	ExecutionTimeMs int64                  `json:"execution_time_ms"`
	IsCriticalPath  bool                   `json:"is_critical_path"`
}

// MultiAgentReconciliationResult handles partial multi-agent execution safely
type MultiAgentReconciliationResult struct {
	WorkflowID          string                 `json:"workflow_id"`
	OverallStatus       string                 `json:"overall_status"` // SUCCEEDED, PARTIAL_SUCCESS, FAILED
	SuccessfulAgents    []string               `json:"successful_agents"`
	FailedAgents        []string               `json:"failed_agents"`
	OriginalConfidence  float64                `json:"original_confidence"`
	AdjustedConfidence  float64                `json:"adjusted_confidence"`
	ConfidenceImpact    float64                `json:"confidence_impact"`
	CanProceedWithPlan  bool                   `json:"can_proceed_with_plan"`
	RequiresHumanReview bool                   `json:"requires_human_review"`
	FallbackStrategy    string                 `json:"fallback_strategy"`
	ConsolidatedOutput  map[string]interface{} `json:"consolidated_output"`
}

// EventBackpressureMetrics exposes real-time mesh capacity and storm safeguards
type EventBackpressureMetrics struct {
	OrgID                int64     `json:"org_id"`
	CurrentBacklog       int       `json:"current_backlog"`
	BacklogThreshold     int       `json:"backlog_threshold"`
	BackpressureActive   bool      `json:"backpressure_active"`
	CoalescedEventCount  int64     `json:"coalesced_event_count"`
	ThrottledEventCount  int64     `json:"throttled_event_count"`
	DeadLetterCount      int       `json:"dead_letter_count"`
	ActiveWorkers        int       `json:"active_workers"`
	WorkerConcurrencyCap int       `json:"worker_concurrency_cap"`
	LastEvaluatedAt      time.Time `json:"last_evaluated_at"`
}

// RetryPolicy defines bounded retry behavior with exponential backoff and jitter
type RetryPolicy struct {
	MaxAttempts    int           `json:"max_attempts"`
	InitialBackoff time.Duration `json:"initial_backoff"`
	MaxBackoff     time.Duration `json:"max_backoff"`
	BackoffFactor  float64       `json:"backoff_factor"`
	JitterFraction float64       `json:"jitter_fraction"`
}

// DefaultBoundedRetryPolicy returns the hardened enterprise retry policy
func DefaultBoundedRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		BackoffFactor:  2.0,
		JitterFraction: 0.2,
	}
}
