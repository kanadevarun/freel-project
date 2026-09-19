package event_workflows

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// RawJSON is a raw encoded JSON value that supports database/sql scanning of NULL and TEXT/JSON values
type RawJSON json.RawMessage

func (r *RawJSON) Scan(value interface{}) error {
	if value == nil {
		*r = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*r = append((*r)[0:0], v...)
		return nil
	case string:
		*r = append((*r)[0:0], []byte(v)...)
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into RawJSON", value)
	}
}

func (r RawJSON) Value() (driver.Value, error) {
	if len(r) == 0 {
		return nil, nil
	}
	return []byte(r), nil
}

func (r RawJSON) MarshalJSON() ([]byte, error) {
	if len(r) == 0 {
		return []byte("null"), nil
	}
	return r, nil
}

func (r *RawJSON) UnmarshalJSON(data []byte) error {
	if r == nil {
		return fmt.Errorf("RawJSON: UnmarshalJSON on nil pointer")
	}
	*r = append((*r)[0:0], data...)
	return nil
}

// Ingested Domain Event Record
type EventRecord struct {
	ID               int64      `db:"id" json:"id"`
	OrgID            int64      `db:"org_id" json:"org_id"`
	EventType        string     `db:"event_type" json:"event_type"`
	SourceModule     string     `db:"source_module" json:"source_module"`
	SourceRecordType string     `db:"source_record_type" json:"source_record_type"`
	SourceRecordID   string     `db:"source_record_id" json:"source_record_id"`
	EventVersion     string     `db:"event_version" json:"event_version"`
	ActorType        string     `db:"actor_type" json:"actor_type"`
	ActorID          *int64     `db:"actor_id" json:"actor_id,omitempty"`
	Payload          RawJSON    `db:"payload" json:"payload"`
	DedupKey         string     `db:"dedup_key" json:"dedup_key"`
	Status           string     `db:"status" json:"status"` // RECEIVED, DEDUPLICATED, PROCESSED, IGNORED, FAILED
	FailureReason    *string    `db:"failure_reason" json:"failure_reason,omitempty"`
	RetryCount       int        `db:"retry_count" json:"retry_count"`
	CorrelationID    string     `db:"correlation_id" json:"correlation_id"`
	CausationID      *string    `db:"causation_id" json:"causation_id,omitempty"`
	EventTimestamp   time.Time  `db:"event_timestamp" json:"event_timestamp"`
	ProcessedAt      *time.Time `db:"processed_at" json:"processed_at,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
}

// Cross-Module Workflow Instance Record
type WorkflowInstance struct {
	ID               int64      `db:"id" json:"id"`
	OrgID            int64      `db:"org_id" json:"org_id"`
	EventID          int64      `db:"event_id" json:"event_id"`
	WorkflowType     string     `db:"workflow_type" json:"workflow_type"` // LEAD_FOLLOWUP, SHIPMENT_EXCEPTION_RESPONSE, INVOICE_COLLECTION_ESCALATION, CONTRACT_COMPLIANCE_RENEWAL, RFQ_QUOTATION_DISPATCH
	Status           string     `db:"status" json:"status"`               // RECEIVED, RUNNING, AWAITING_APPROVAL, APPROVED, EXECUTING, COMPLETED, FAILED, CANCELLED, NEEDS_REVIEW, IGNORED
	TriggerEventType string     `db:"trigger_event_type" json:"trigger_event_type"`
	SourceModule     string     `db:"source_module" json:"source_module"`
	SourceRecordType string     `db:"source_record_type" json:"source_record_type"`
	SourceRecordID   string     `db:"source_record_id" json:"source_record_id"`
	Urgency          string     `db:"urgency" json:"urgency"` // LOW, MEDIUM, HIGH, CRITICAL
	AISummary        string     `db:"ai_summary" json:"ai_summary"`
	AIAnalysis       RawJSON    `db:"ai_analysis" json:"ai_analysis"`
	Recommendations  RawJSON    `db:"recommendations" json:"recommendations"`
	ActionIntent     RawJSON    `db:"action_intent" json:"action_intent,omitempty"`
	ApprovalID       *int64     `db:"approval_id" json:"approval_id,omitempty"`
	ActionName       *string    `db:"action_name" json:"action_name,omitempty"`
	ActionStatus     string     `db:"action_status" json:"action_status"` // NOT_STARTED, AWAITING_APPROVAL, EXECUTING, EXECUTED, FAILED, SKIPPED
	DraftID          *int64     `db:"draft_id" json:"draft_id,omitempty"`
	DedupKey         string     `db:"dedup_key" json:"dedup_key"`
	CorrelationID    string     `db:"correlation_id" json:"correlation_id"`
	RetryCount       int        `db:"retry_count" json:"retry_count"`
	ErrorMessage     *string    `db:"error_message" json:"error_message,omitempty"`
	StartedAt        time.Time  `db:"started_at" json:"started_at"`
	CompletedAt      *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

// IngestEventInput for ingesting domain events
type IngestEventInput struct {
	EventType        string                 `json:"event_type"`
	SourceModule     string                 `json:"source_module"`
	SourceRecordType string                 `json:"source_record_type"`
	SourceRecordID   string                 `json:"source_record_id"`
	EventVersion     string                 `json:"event_version,omitempty"`
	ActorType        string                 `json:"actor_type,omitempty"`
	ActorID          *int64                 `json:"actor_id,omitempty"`
	Payload          map[string]interface{} `json:"payload"`
	CorrelationID    string                 `json:"correlation_id,omitempty"`
	CausationID      *string                `json:"causation_id,omitempty"`
	Timestamp        *time.Time             `json:"timestamp,omitempty"`
}

// WorkflowRecommendation from AI sidecar
type WorkflowRecommendation struct {
	RecommendationType string                 `json:"recommendation_type"`
	Title              string                 `json:"title"`
	Description        string                 `json:"description"`
	ActionName         string                 `json:"action_name"`
	ActionPayload      map[string]interface{} `json:"action_payload"`
	RequiresApproval   bool                   `json:"requires_approval"`
	RiskLevel          string                 `json:"risk_level"`
	TargetModule       string                 `json:"target_module"`
}

// Sidecar Event Analysis Response
type EventAnalysisResponse struct {
	WorkflowType        string                   `json:"workflow_type"`
	Urgency             string                   `json:"urgency"`
	SignificanceScore   float64                  `json:"significance_score"`
	Summary             string                   `json:"summary"`
	CrossModuleInsights []string                 `json:"cross_module_insights"`
	RecommendedActions  []WorkflowRecommendation `json:"recommended_actions"`
	MissingInformation  []string                 `json:"missing_information"`
	ConfidenceScore     float64                  `json:"confidence_score"`
	CorrelationID       string                   `json:"correlation_id"`
}

// Overview Summary Response
type EventWorkflowsOverview struct {
	TotalEventsIngested    int                 `json:"total_events_ingested"`
	ActiveWorkflowsCount   int                 `json:"active_workflows_count"`
	PendingApprovalsCount  int                 `json:"pending_approvals_count"`
	RecentEvents           []*EventRecord      `json:"recent_events"`
	RecentWorkflows        []*WorkflowInstance `json:"recent_workflows"`
	SupportedWorkflowsList []string            `json:"supported_workflows"`
}
