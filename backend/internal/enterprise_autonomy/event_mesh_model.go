package enterprise_autonomy

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ---------------------------------------------------------------------
// Phase 7.8: Enterprise Event Mesh & Autonomous Workflow Engine Model
// ---------------------------------------------------------------------

// EventProcessingStatus defines the deterministic state of an ingested event
type EventProcessingStatus string

const (
	EventStatusReceived       EventProcessingStatus = "RECEIVED"
	EventStatusValidating     EventProcessingStatus = "VALIDATING"
	EventStatusDeduplicated   EventProcessingStatus = "DEDUPLICATED"
	EventStatusRouted         EventProcessingStatus = "ROUTED"
	EventStatusWorkflowActive EventProcessingStatus = "WORKFLOW_ACTIVE"
	EventStatusCompleted      EventProcessingStatus = "COMPLETED"
	EventStatusDeadLetter     EventProcessingStatus = "DEAD_LETTER"
	EventStatusStaleIgnored   EventProcessingStatus = "STALE_IGNORED"
	EventStatusLoopSuppressed EventProcessingStatus = "LOOP_SUPPRESSED"
)

// EventPriority defines the processing queue priority
type EventPriority string

const (
	PriorityCritical EventPriority = "CRITICAL"
	PriorityHigh     EventPriority = "HIGH"
	PriorityNormal   EventPriority = "NORMAL"
	PriorityLow      EventPriority = "LOW"
)

// Common error definitions for the Event Mesh
var (
	ErrMalformedEventSchema    = errors.New("event schema validation failed: missing required fields")
	ErrStaleEventIgnored       = errors.New("event timestamp is stale or superseded by newer state")
	ErrEventLoopDetected       = errors.New("event loop detected: event caused by own completed action without material change")
	ErrWorkflowLimitExceeded   = errors.New("event processing backpressure limit reached: maximum concurrent workflows active")
	ErrDeadLetterPersisted     = errors.New("unrecoverable event routed to dead-letter queue")
	ErrUnsupportedEventVersion = errors.New("unsupported event schema version")
)

// NormalizedBusinessEvent represents the normalized enterprise business event
type NormalizedBusinessEvent struct {
	EventID          string                 `json:"event_id" db:"event_id"`
	OrgID            int64                  `json:"org_id" db:"org_id"`
	EventType        string                 `json:"event_type" db:"event_type"`
	SourceModule     string                 `json:"source_module" db:"source_module"`
	EntityType       string                 `json:"entity_type" db:"entity_type"`
	EntityID         string                 `json:"entity_id" db:"entity_id"`
	ParentEntityID   *string                `json:"parent_entity_id,omitempty" db:"parent_entity_id"`
	EventVersion     string                 `json:"event_version" db:"event_version"`
	Priority         EventPriority          `json:"priority" db:"priority"`
	Payload          map[string]interface{} `json:"payload" db:"payload"`
	CorrelationID    string                 `json:"correlation_id" db:"correlation_id"`
	CausationID      string                 `json:"causation_id" db:"causation_id"`
	ActionID         *string                `json:"action_id,omitempty" db:"action_id"`
	OccurredAt       time.Time              `json:"occurred_at" db:"occurred_at"`
	ReceivedAt       time.Time              `json:"received_at" db:"received_at"`
	ProcessingStatus EventProcessingStatus  `json:"processing_status" db:"processing_status"`
	WorkflowID       *string                `json:"workflow_id,omitempty" db:"workflow_id"`
	WorkflowType     *EnterpriseWorkflowType `json:"workflow_type,omitempty" db:"workflow_type"`
	RetryCount       int                    `json:"retry_count" db:"retry_count"`
	MaxRetries       int                    `json:"max_retries" db:"max_retries"`
	FailureReason    *string                `json:"failure_reason,omitempty" db:"failure_reason"`
	DedupKey         string                 `json:"dedup_key" db:"dedup_key"`
}

// EventToWorkflowRule defines the declarative mapping between an event and its target workflow
type EventToWorkflowRule struct {
	EventType          string                 `json:"event_type"`
	SourceModule       string                 `json:"source_module"`
	TargetWorkflowType EnterpriseWorkflowType `json:"target_workflow_type"`
	DefaultPriority    EventPriority          `json:"default_priority"`
	RequiredCapability string                 `json:"required_capability"`
	MinAutonomyLevel   string                 `json:"min_autonomy_level"`
	RequiresApproval   bool                   `json:"requires_approval"`
	CooldownSeconds    int                    `json:"cooldown_seconds"`
	MaxRetries         int                    `json:"max_retries"`
	IsEnabled          bool                   `json:"is_enabled"`
	Description        string                 `json:"description"`
}

// DeadLetterEvent represents an unprocessable or permanently failed event for audit & manual triage
type DeadLetterEvent struct {
	EventID        string                 `json:"event_id"`
	OrgID          int64                  `json:"org_id"`
	EventType      string                 `json:"event_type"`
	SourceModule   string                 `json:"source_module"`
	EntityType     string                 `json:"entity_type"`
	EntityID       string                 `json:"entity_id"`
	CorrelationID  string                 `json:"correlation_id"`
	CausationID    string                 `json:"causation_id"`
	Payload        map[string]interface{} `json:"payload"`
	FailureReason  string                 `json:"failure_reason"`
	AttemptCount   int                    `json:"attempt_count"`
	DeadLetteredAt time.Time              `json:"dead_lettered_at"`
	Resolved       bool                   `json:"resolved"`
	ResolutionNote *string                `json:"resolution_note,omitempty"`
}

// EventMeshOverview provides high-level observability for operational control
type EventMeshOverview struct {
	OrgID                 int64                      `json:"org_id"`
	TotalEventsReceived   int                        `json:"total_events_received"`
	EventsRoutedCount     int                        `json:"events_routed_count"`
	WorkflowsTriggered    int                        `json:"workflows_triggered"`
	DeduplicatedCount     int                        `json:"deduplicated_count"`
	LoopsSuppressedCount  int                        `json:"loops_suppressed_count"`
	DeadLetterCount       int                        `json:"dead_letter_count"`
	ActiveWorkflowsCount  int                        `json:"active_workflows_count"`
	RecentEvents          []*NormalizedBusinessEvent `json:"recent_events"`
	RecentDeadLetters     []*DeadLetterEvent         `json:"recent_dead_letters"`
	ActiveRoutingRules    []EventToWorkflowRule      `json:"active_routing_rules"`
}

// IngestBusinessEventRequest is the incoming payload schema from any domain or external gateway
type IngestBusinessEventRequest struct {
	EventID        string                 `json:"event_id"`
	EventType      string                 `json:"event_type"`
	SourceModule   string                 `json:"source_module"`
	EntityType     string                 `json:"entity_type"`
	EntityID       string                 `json:"entity_id"`
	ParentEntityID *string                `json:"parent_entity_id,omitempty"`
	EventVersion   string                 `json:"event_version,omitempty"`
	Priority       string                 `json:"priority,omitempty"`
	Payload        map[string]interface{} `json:"payload"`
	CorrelationID  string                 `json:"correlation_id,omitempty"`
	CausationID    string                 `json:"causation_id,omitempty"`
	ActionID       *string                `json:"action_id,omitempty"`
	OccurredAt     *time.Time             `json:"occurred_at,omitempty"`
}

// ValidateNormalizedEvent performs strict server-side structural and semantic validation
func ValidateNormalizedEvent(evt *NormalizedBusinessEvent) error {
	if evt.OrgID <= 0 {
		return ErrUnauthorizedTenant
	}
	if strings.TrimSpace(evt.EventID) == "" {
		return fmt.Errorf("%w: event_id is required", ErrMalformedEventSchema)
	}
	if strings.TrimSpace(evt.EventType) == "" {
		return fmt.Errorf("%w: event_type is required", ErrMalformedEventSchema)
	}
	if strings.TrimSpace(evt.EntityType) == "" || strings.TrimSpace(evt.EntityID) == "" {
		return fmt.Errorf("%w: entity_type and entity_id are required", ErrMalformedEventSchema)
	}
	if evt.EventVersion == "" {
		evt.EventVersion = "1.0"
	}
	if evt.EventVersion != "1.0" && evt.EventVersion != "2.0" {
		return fmt.Errorf("%w: version %s is not supported", ErrUnsupportedEventVersion, evt.EventVersion)
	}
	if evt.Priority == "" {
		evt.Priority = PriorityNormal
	}
	if evt.MaxRetries <= 0 {
		evt.MaxRetries = 3
	}
	return nil
}
