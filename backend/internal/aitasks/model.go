package aitasks

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Task Status Constants
const (
	StatusQueued             = "QUEUED"
	StatusProcessing         = "PROCESSING"
	StatusWaitingForApproval = "WAITING_FOR_APPROVAL"
	StatusRetrying           = "RETRYING"
	StatusCompleted          = "COMPLETED"
	StatusFailed             = "FAILED"
	StatusRejected           = "REJECTED"
	StatusCancelled          = "CANCELLED"
)

// Common Error Codes
const (
	ErrCodeInvalidTransition   = "INVALID_TRANSITION"
	ErrCodeTaskNotFound        = "TASK_NOT_FOUND"
	ErrCodeTaskCancelled       = "TASK_CANCELLED"
	ErrCodeTaskExpired         = "TASK_EXPIRED"
	ErrCodeTaskRejected        = "TASK_REJECTED"
	ErrCodeMaxRetriesExceeded  = "MAX_RETRIES_EXCEEDED"
	ErrCodeLeaseExpired        = "LEASE_EXPIRED"
	ErrCodeWorkerMismatch      = "WORKER_MISMATCH"
	ErrCodeTenantMismatch      = "TENANT_MISMATCH"
	ErrCodeTerminalState       = "TERMINAL_STATE"
	ErrCodeApprovalRequired    = "APPROVAL_REQUIRED"
)

// Sentinel Errors
var (
	ErrInvalidTransition = errors.New("invalid task status transition")
	ErrTaskNotFound      = errors.New("task not found")
	ErrTaskCancelled     = errors.New("task has been cancelled")
	ErrTaskRejected      = errors.New("task action was rejected by human reviewer")
	ErrTerminalState     = errors.New("task is in a terminal state and cannot be modified")
	ErrWorkerMismatch    = errors.New("worker lease mismatch")
	ErrTenantMismatch    = errors.New("organization tenant access denied")
)

// Task represents an autonomous AI execution record in ai_processing_tasks.
type Task struct {
	ID             int64      `db:"id" json:"id"`
	OrgID          int64      `db:"org_id" json:"org_id"`
	DocumentID     *string    `db:"document_id" json:"document_id,omitempty"`
	EntityType     *string    `db:"entity_type" json:"entity_type,omitempty"`
	EntityID       *string    `db:"entity_id" json:"entity_id,omitempty"`
	TaskType       string     `db:"task_type" json:"task_type"`
	ResourceID     *int64     `db:"resource_id" json:"resource_id,omitempty"`
	Payload        *string    `db:"payload" json:"payload,omitempty"`
	Status         string     `db:"status" json:"status"`
	ErrorMessage   *string    `db:"error_message" json:"error_message,omitempty"`
	LastErrorCode  *string    `db:"last_error_code" json:"last_error_code,omitempty"`
	RetryCount     int        `db:"retry_count" json:"retry_count"`
	MaxRetries     int        `db:"max_retries" json:"max_retries"`
	WorkerID       *string    `db:"worker_id" json:"worker_id,omitempty"`
	LeaseExpiresAt *time.Time `db:"lease_expires_at" json:"lease_expires_at,omitempty"`
	HeartbeatAt    *time.Time `db:"heartbeat_at" json:"heartbeat_at,omitempty"`
	StartedAt      *time.Time `db:"started_at" json:"started_at,omitempty"`
	CompletedAt    *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	AvailableAt    *time.Time `db:"available_at" json:"available_at,omitempty"`
	ThreadID       *string    `db:"thread_id" json:"thread_id,omitempty"`
	CorrelationID  *string    `db:"correlation_id" json:"correlation_id,omitempty"`
	ApprovalID     *int64     `db:"approval_id" json:"approval_id,omitempty"`
	ActingUserID   *int64     `db:"acting_user_id" json:"acting_user_id,omitempty"`
	ActorType      string     `db:"actor_type" json:"actor_type"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}

// CreateTaskInput holds parameters for enqueuing a new AI processing task.
type CreateTaskInput struct {
	OrgID         int64                  `json:"org_id"`
	TaskType      string                 `json:"task_type"`
	EntityType    string                 `json:"entity_type,omitempty"`
	EntityID      string                 `json:"entity_id,omitempty"`
	DocumentID    string                 `json:"document_id,omitempty"`
	ResourceID    *int64                 `json:"resource_id,omitempty"`
	Payload       map[string]interface{} `json:"payload"`
	ThreadID      string                 `json:"thread_id,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	ActingUserID  *int64                 `json:"acting_user_id,omitempty"`
	ActorType     string                 `json:"actor_type,omitempty"`
	MaxRetries    int                    `json:"max_retries,omitempty"`
}

// UpdateTaskStatusInput specifies updates when transitioning a task's state.
type UpdateTaskStatusInput struct {
	Status        string  `json:"status"`
	WorkerID      string  `json:"worker_id,omitempty"`
	ErrorMessage  string  `json:"error_message,omitempty"`
	LastErrorCode string  `json:"last_error_code,omitempty"`
	ThreadID      string  `json:"thread_id,omitempty"`
	CorrelationID string  `json:"correlation_id,omitempty"`
	ApprovalID    *int64  `json:"approval_id,omitempty"`
	DelaySeconds  int     `json:"delay_seconds,omitempty"` // For RETRYING -> QUEUED backoff
}

// TaskFilter defines criteria for querying tasks.
type TaskFilter struct {
	Status     string `json:"status,omitempty"`
	TaskType   string `json:"task_type,omitempty"`
	EntityType string `json:"entity_type,omitempty"`
	EntityID   string `json:"entity_id,omitempty"`
	ThreadID   string `json:"thread_id,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
}

// TaskStats aggregates task metrics for telemetry.
type TaskStats struct {
	Queued             int64 `json:"queued"`
	Processing         int64 `json:"processing"`
	WaitingForApproval int64 `json:"waiting_for_approval"`
	Retrying           int64 `json:"retrying"`
	Completed          int64 `json:"completed"`
	Failed             int64 `json:"failed"`
	Rejected           int64 `json:"rejected"`
	Cancelled          int64 `json:"cancelled"`
	Total              int64 `json:"total"`
}

// validTransitions defines authoritative permitted state transitions in the state machine.
var validTransitions = map[string]map[string]bool{
	StatusQueued: {
		StatusProcessing: true,
		StatusCancelled:  true,
		StatusFailed:     true, // E.g. unrecoverable pre-execution validation failure
	},
	StatusProcessing: {
		StatusCompleted:          true,
		StatusWaitingForApproval: true,
		StatusRetrying:           true,
		StatusFailed:             true,
		StatusCancelled:          true,
	},
	StatusWaitingForApproval: {
		StatusProcessing: true, // Resume upon approval
		StatusCompleted:  true, // Direct resolution upon approval
		StatusRejected:    true, // Human reviewer rejected
		StatusCancelled:   true, // Human reviewer or user cancelled
	},
	StatusRetrying: {
		StatusQueued:    true, // Re-staged after backoff expiry
		StatusCancelled: true,
		StatusFailed:    true, // Max retries exceeded
	},
	// Terminal states have no normal transitions
	StatusCompleted: {},
	StatusRejected:  {},
	StatusCancelled: {},
	StatusFailed: {
		StatusQueued: true, // Permitted ONLY via explicit manual retry operation!
	},
}

// IsTerminalStatus returns true if the status represents a final, unalterable state.
func IsTerminalStatus(status string) bool {
	upper := strings.ToUpper(status)
	return upper == StatusCompleted || upper == StatusRejected || upper == StatusCancelled || upper == StatusFailed
}

// ValidateTransition validates whether moving from currentStatus to targetStatus is permitted.
// If isExplicitRetry is true, transitioning from FAILED to QUEUED is permitted.
func ValidateTransition(currentStatus, targetStatus string, isExplicitRetry bool) error {
	from := strings.ToUpper(strings.TrimSpace(currentStatus))
	to := strings.ToUpper(strings.TrimSpace(targetStatus))

	if from == to {
		return nil // No-op transition is harmless
	}

	if from == StatusFailed && to == StatusQueued {
		if isExplicitRetry {
			return nil
		}
		return fmt.Errorf("%w: cannot transition from FAILED to QUEUED without explicit retry authorization", ErrInvalidTransition)
	}

	allowedTargets, exists := validTransitions[from]
	if !exists {
		return fmt.Errorf("%w: unknown current status '%s'", ErrInvalidTransition, from)
	}

	if !allowedTargets[to] {
		return fmt.Errorf("%w: illegal transition from '%s' to '%s'", ErrInvalidTransition, from, to)
	}

	return nil
}
