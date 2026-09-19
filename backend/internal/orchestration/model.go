package orchestration

import (
	"encoding/json"
	"time"
)

// Lifecycle Statuses for Action Proposals & Executions
const (
	StatusProposed            = "PROPOSED"
	StatusValidating          = "VALIDATING"
	StatusWaitingForApproval  = "WAITING_FOR_APPROVAL"
	StatusApproved            = "APPROVED"
	StatusRejected            = "REJECTED"
	StatusExpired             = "EXPIRED"
	StatusQueued              = "QUEUED"
	StatusExecuting           = "EXECUTING"
	StatusVerificationPending = "VERIFICATION_PENDING"
	StatusCompleted           = "COMPLETED"
	StatusPartiallyCompleted  = "PARTIALLY_COMPLETED"
	StatusFailed              = "FAILED"
	StatusCancelled           = "CANCELLED"
	StatusRetrying            = "RETRYING"
)

// Execution Steps
const (
	StepQueued            = "QUEUED"
	StepValidating        = "VALIDATING"
	StepExecutingAction   = "EXECUTING_ACTION"
	StepVerifyingOutcome  = "VERIFYING_OUTCOME"
	StepCompleted         = "COMPLETED"
	StepFailed            = "FAILED"
)

// Risk Levels
const (
	RiskLow      = "LOW"
	RiskMedium   = "MEDIUM"
	RiskHigh     = "HIGH"
	RiskCritical = "CRITICAL"
)

// Approval Policies
const (
	ApprovalPolicyAlways        = "ALWAYS_REQUIRE_APPROVAL"
	ApprovalPolicyThreshold     = "THRESHOLD_BASED"
	ApprovalPolicyAutomaticSafe = "AUTOMATIC_FOR_SAFE_ACTIONS"
)

// Reversibility Ratings
const (
	ReversibilityReversible          = "REVERSIBLE"
	ReversibilityPartiallyReversible = "PARTIALLY_REVERSIBLE"
	ReversibilityIrreversible        = "IRREVERSIBLE"
)

// FactualEvidence represents a verified piece of evidence used by the AI
type FactualEvidence struct {
	FieldName     string      `json:"field_name"`
	ObservedValue interface{} `json:"observed_value"`
	BaselineValue interface{} `json:"baseline_value,omitempty"`
	SourceRecord  string      `json:"source_record,omitempty"`
	Timestamp     string      `json:"timestamp,omitempty"`
	FactType      string      `json:"fact_type"` // CONFIRMED_FACT, CALCULATED_VALUE, INFERENCE, THRESHOLD_BREACH
}

// ActionProposal represents a durable proposal produced by the Python AI Orchestrator
type ActionProposal struct {
	ID                 int64             `db:"id" json:"id"`
	OrgID              int64             `db:"org_id" json:"org_id"`
	ProposalID         string            `db:"proposal_id" json:"proposal_id"`
	SourceModule       string            `db:"source_module" json:"source_module"`
	SourceRecordType   string            `db:"source_record_type" json:"source_record_type"`
	SourceRecordID     string            `db:"source_record_id" json:"source_record_id"`
	TriggerEvent       string            `db:"trigger_event" json:"trigger_event"`
	ProposedActionType string            `db:"proposed_action_type" json:"proposed_action_type"`
	ActionParameters   json.RawMessage   `db:"action_parameters" json:"action_parameters"`
	Explanation        string            `db:"explanation" json:"explanation"`
	Evidence           json.RawMessage   `db:"evidence" json:"evidence"`
	Confidence         float64           `db:"confidence" json:"confidence"`
	RiskLevel          string            `db:"risk_level" json:"risk_level"`
	Priority           string            `db:"priority" json:"priority"`
	RequiresApproval   bool              `db:"requires_approval" json:"requires_approval"`
	ApprovalPolicy     string            `db:"approval_policy" json:"approval_policy"`
	ExpectedImpact     string            `db:"expected_impact" json:"expected_impact"`
	Reversibility      string            `db:"reversibility" json:"reversibility"`
	MissingInformation *json.RawMessage  `db:"missing_information" json:"missing_information,omitempty"`
	DataFreshness      string            `db:"data_freshness" json:"data_freshness"`
	CorrelationID      string            `db:"correlation_id" json:"correlation_id"`
	ApprovalID         *int64            `db:"approval_id" json:"approval_id,omitempty"`
	ExecutionID        *int64            `db:"execution_id" json:"execution_id,omitempty"`
	Status             string            `db:"status" json:"status"`
	SchemaVersion      string            `db:"schema_version" json:"schema_version"`
	CreatedBy          *int64            `db:"created_by" json:"created_by,omitempty"`
	ExpiresAt          *time.Time        `db:"expires_at" json:"expires_at,omitempty"`
	CreatedAt          time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time         `db:"updated_at" json:"updated_at"`

	// Parsed fields for convenient API presentation
	ParsedEvidence []FactualEvidence `json:"parsed_evidence,omitempty"`
}

// ActionExecution records the execution and verification lifecycle of an action
type ActionExecution struct {
	ID                  int64            `db:"id" json:"id"`
	OrgID               int64            `db:"org_id" json:"org_id"`
	ProposalID          string           `db:"proposal_id" json:"proposal_id"`
	ActionName          string           `db:"action_name" json:"action_name"`
	IdempotencyKey      string           `db:"idempotency_key" json:"idempotency_key"`
	Status              string           `db:"status" json:"status"`
	CurrentStep         string           `db:"current_step" json:"current_step"`
	StepResults         *json.RawMessage `db:"step_results" json:"step_results,omitempty"`
	InputPayload        json.RawMessage  `db:"input_payload" json:"input_payload"`
	OutputPayload       *json.RawMessage `db:"output_payload" json:"output_payload,omitempty"`
	VerificationStatus  string           `db:"verification_status" json:"verification_status"`
	VerificationDetails *json.RawMessage `db:"verification_details" json:"verification_details,omitempty"`
	RetryCount          int              `db:"retry_count" json:"retry_count"`
	MaxRetries          int              `db:"max_retries" json:"max_retries"`
	ErrorMessage        *string          `db:"error_message" json:"error_message,omitempty"`
	ExecutedBy          *int64           `db:"executed_by" json:"executed_by,omitempty"`
	CorrelationID       string           `db:"correlation_id" json:"correlation_id"`
	StartedAt           *time.Time       `db:"started_at" json:"started_at,omitempty"`
	CompletedAt         *time.Time       `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt           time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time        `db:"updated_at" json:"updated_at"`
}

// Request and Filter DTOs
type GenerateProposalRequest struct {
	SourceModule         string                 `json:"source_module" binding:"required"`
	SourceRecordType     string                 `json:"source_record_type" binding:"required"`
	SourceRecordID       string                 `json:"source_record_id" binding:"required"`
	TriggerEvent         string                 `json:"trigger_event" binding:"required"`
	OperationalSignals   []OperationalSignalDTO `json:"operational_signals,omitempty"`
	PreferredActionTypes []string               `json:"preferred_action_types,omitempty"`
	CorrelationID        string                 `json:"correlation_id,omitempty"`
}

type OperationalSignalDTO struct {
	SignalType  string      `json:"signal_type"`
	SignalValue interface{} `json:"signal_value"`
	Severity    string      `json:"severity"`
	DetectedAt  string      `json:"detected_at,omitempty"`
}

type ExecuteProposalRequest struct {
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

type ProposalFilter struct {
	Status       *string
	SourceModule *string
	RiskLevel    *string
	Search       *string
	Limit        int
	Offset       int
}

type ExecutionFilter struct {
	Status     *string
	ActionName *string
	ProposalID *string
	Limit      int
	Offset     int
}
