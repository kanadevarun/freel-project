package enterprise_autonomy

import (
	"errors"
	"time"
)

// EnterpriseWorkflowType defines the recognized cross-module autonomous workflows
type EnterpriseWorkflowType string

const (
	WorkflowShipmentRecovery    EnterpriseWorkflowType = "SHIPMENT_RECOVERY"
	WorkflowCommercialCycle     EnterpriseWorkflowType = "COMMERCIAL_CYCLE"
	WorkflowFinancialCollection EnterpriseWorkflowType = "FINANCIAL_COLLECTION"
	WorkflowCrossModuleRisk     EnterpriseWorkflowType = "CROSS_MODULE_RISK"
	WorkflowAdaptiveReplanning  EnterpriseWorkflowType = "ADAPTIVE_REPLANNING"
	WorkflowComplianceAudit     EnterpriseWorkflowType = "COMPLIANCE_AUDIT"
	WorkflowOperationalRecovery EnterpriseWorkflowType = "OPERATIONAL_RECOVERY"
	WorkflowCustomerRelationship EnterpriseWorkflowType = "CUSTOMER_RELATIONSHIP"
	WorkflowRevenueOptimization  EnterpriseWorkflowType = "REVENUE_OPTIMIZATION"
	WorkflowContractComplianceRisk EnterpriseWorkflowType = "CONTRACT_COMPLIANCE_RISK"
)

// EnterpriseWorkflowState represents durable lifecycle states of an autonomous workflow
type EnterpriseWorkflowState string

const (
	StatePending            EnterpriseWorkflowState = "PENDING"
	StateRunning            EnterpriseWorkflowState = "RUNNING"
	StateWaiting            EnterpriseWorkflowState = "WAITING"
	StateWaitingForApproval EnterpriseWorkflowState = "WAITING_FOR_APPROVAL"
	StateBlocked            EnterpriseWorkflowState = "BLOCKED"
	StatePaused             EnterpriseWorkflowState = "PAUSED"
	StateEscalated          EnterpriseWorkflowState = "ESCALATED"
	StateCompleted          EnterpriseWorkflowState = "COMPLETED"
	StateFailed             EnterpriseWorkflowState = "FAILED"
	StateCancelled          EnterpriseWorkflowState = "CANCELLED"
)

var (
	ErrInvalidStateTransition   = errors.New("invalid enterprise workflow state transition")
	ErrUnauthorizedTenant       = errors.New("unauthorized tenant access")
	ErrWorkflowNotFound         = errors.New("enterprise workflow not found")
	ErrWorkflowBlocked          = errors.New("workflow is currently blocked")
	ErrWorkflowPaused           = errors.New("workflow is currently paused")
	ErrEmergencyStopActive      = errors.New("emergency stop is active for requested scope")
	ErrAutonomyRestricted       = errors.New("action restricted by enterprise autonomy policy")
	ErrApprovalRequired         = errors.New("action requires human-in-the-loop approval before execution")
	ErrDuplicateEventTrigger    = errors.New("duplicate event trigger rejected by deduplication filter")
	ErrDependencyIncomplete     = errors.New("prerequisite step dependency is not completed")
	ErrMalformedAIResponse      = errors.New("malformed or unauthorized response structure from AI sidecar")
)

// ValidateStateTransition enforces explicit, deterministic workflow lifecycle rules
func ValidateStateTransition(current, next EnterpriseWorkflowState) error {
	if current == next {
		return nil
	}

	valid := false
	switch current {
	case StatePending:
		valid = (next == StateRunning || next == StateWaiting || next == StatePaused || next == StateCancelled)
	case StateRunning:
		valid = (next == StateWaiting || next == StateWaitingForApproval || next == StateBlocked ||
			next == StatePaused || next == StateEscalated || next == StateCompleted ||
			next == StateFailed || next == StateCancelled)
	case StateWaiting:
		valid = (next == StateRunning || next == StatePaused || next == StateEscalated ||
			next == StateFailed || next == StateCancelled)
	case StateWaitingForApproval:
		valid = (next == StateRunning || next == StatePaused || next == StateBlocked || next == StateEscalated ||
			next == StateFailed || next == StateCancelled)
	case StateBlocked:
		valid = (next == StateRunning || next == StateEscalated || next == StateCancelled || next == StateFailed)
	case StatePaused:
		valid = (next == StateRunning || next == StateWaitingForApproval || next == StateCancelled || next == StateFailed)
	case StateEscalated:
		valid = (next == StateRunning || next == StateCompleted || next == StateCancelled || next == StateFailed)
	case StateCompleted, StateFailed, StateCancelled:
		// Terminal states cannot transition further
		valid = false
	default:
		valid = false
	}

	if !valid {
		return ErrInvalidStateTransition
	}
	return nil
}

// EnterprisePolicyContext encapsulates tenant, organization, and workflow governance rules
type EnterprisePolicyContext struct {
	OrgID                  int64           `json:"org_id"`
	AutonomyLevel          string          `json:"autonomy_level"`
	AllowedWorkflowTypes   []string        `json:"allowed_workflow_types"`
	AllowedActionTypes     []string        `json:"allowed_action_types"`
	ProhibitedActionTypes  []string        `json:"prohibited_action_types"`
	RequiresApproval       bool            `json:"requires_approval"`
	MaxMonetaryThreshold   float64         `json:"max_monetary_threshold"`
	MinConfidenceThreshold float64         `json:"min_confidence_threshold"`
	EmergencyStopActive    bool            `json:"emergency_stop_active"`
	FeatureFlags           map[string]bool `json:"feature_flags,omitempty"`
}

// EnterpriseWorkflow represents a durable enterprise-grade autonomous workflow
type EnterpriseWorkflow struct {
	ID                  int64                      `json:"id" db:"id"`
	WorkflowID          string                     `json:"workflow_id" db:"plan_id"`
	OrgID               int64                      `json:"org_id" db:"org_id"`
	WorkflowType        EnterpriseWorkflowType     `json:"workflow_type" db:"module"`
	Objective           string                     `json:"objective" db:"goal"`
	InitiatingEvent     string                     `json:"initiating_event" db:"triggering_event"`
	ParentWorkflowID    *string                    `json:"parent_workflow_id,omitempty" db:"parent_plan_id"`
	CurrentState        EnterpriseWorkflowState    `json:"current_state" db:"status"`
	AutonomyLevel       string                     `json:"autonomy_level" db:"autonomy_level"`
	PolicyDecision      string                     `json:"policy_decision" db:"policy_decision"`
	PolicyReason        *string                    `json:"policy_reason,omitempty" db:"policy_reason"`
	PolicyContext       *EnterprisePolicyContext   `json:"policy_context,omitempty"`
	AssignedAgents      []string                   `json:"assigned_agents"`
	CurrentStep         string                     `json:"current_step" db:"current_step_id"`
	Steps               []EnterpriseWorkflowStep   `json:"steps"`
	PendingApprovals    []EnterprisePendingApproval `json:"pending_approvals,omitempty"`
	Result              map[string]interface{}     `json:"result,omitempty"`
	Confidence          float64                    `json:"confidence" db:"confidence_score"`
	ErrorCode           *string                    `json:"error_code,omitempty"`
	ErrorMessage        *string                    `json:"error_message,omitempty" db:"escalation_reason"`
	CorrelationID       string                     `json:"correlation_id" db:"correlation_id"`
	IdempotencyKey      string                     `json:"idempotency_key"`
	RelatedEntityType   string                     `json:"related_entity_type" db:"related_entity_type"`
	RelatedEntityID     string                     `json:"related_entity_id" db:"related_entity_id"`
	StoppedByUserID     *int64                     `json:"stopped_by_user_id,omitempty" db:"stopped_by_user_id"`
	StoppedAt           *time.Time                 `json:"stopped_at,omitempty" db:"stopped_at"`
	StopReason          *string                    `json:"stop_reason,omitempty" db:"stop_reason"`
	CreatedAt           time.Time                  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time                  `json:"updated_at" db:"updated_at"`
	CompletedAt         *time.Time                 `json:"completed_at,omitempty"`
}

// EnterpriseWorkflowStep defines a discrete step within an enterprise autonomous workflow
type EnterpriseWorkflowStep struct {
	ID                   int64                  `json:"id" db:"id"`
	StepID               string                 `json:"step_id" db:"step_id"`
	WorkflowID           string                 `json:"workflow_id" db:"plan_id"`
	OrgID                int64                  `json:"org_id" db:"org_id"`
	StepNumber           int                    `json:"step_number" db:"step_number"`
	AgentID              string                 `json:"agent_id"`
	ActionType           string                 `json:"action_type" db:"action_type"`
	Title                string                 `json:"title" db:"title"`
	Description          string                 `json:"description" db:"description"`
	Parameters           map[string]interface{} `json:"parameters"`
	Dependencies         []string               `json:"dependencies"`
	ExpectedOutcome      string                 `json:"expected_outcome" db:"expected_outcome"`
	RiskLevel            string                 `json:"risk_level" db:"risk_level"`
	RequiresApproval     bool                   `json:"requires_approval" db:"requires_approval"`
	ActionSystemActionID *string                `json:"action_system_action_id,omitempty" db:"action_system_action_id"`
	IdempotencyKey       string                 `json:"idempotency_key" db:"idempotency_key"`
	Status               string                 `json:"status" db:"status"` // PENDING, RUNNING, WAITING_FOR_APPROVAL, COMPLETED, FAILED, SKIPPED
	ExecutionAttempt     int                    `json:"execution_attempt" db:"execution_attempt"`
	MaxAttempts          int                    `json:"max_attempts" db:"max_attempts"`
	ExecutedAt           *time.Time             `json:"executed_at,omitempty" db:"executed_at"`
	ExecutionResult      map[string]interface{} `json:"execution_result,omitempty"`
	ErrorMessage         *string                `json:"error_message,omitempty" db:"error_message"`
	CreatedAt            time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at" db:"updated_at"`
}

// EnterprisePendingApproval links a workflow step to an action requiring human approval
type EnterprisePendingApproval struct {
	ApprovalID     int64                  `json:"approval_id"`
	WorkflowID     string                 `json:"workflow_id"`
	StepID         string                 `json:"step_id"`
	ActionType     string                 `json:"action_type"`
	ProposedAction map[string]interface{} `json:"proposed_action"`
	Reason         string                 `json:"reason"`
	RiskLevel      string                 `json:"risk_level"`
	Confidence     float64                `json:"confidence"`
	CreatedAt      time.Time              `json:"created_at"`
}

// Request & Response DTOs

type StartWorkflowRequest struct {
	WorkflowType      EnterpriseWorkflowType `json:"workflow_type"`
	Objective         string                 `json:"objective"`
	InitiatingEvent   string                 `json:"initiating_event,omitempty"`
	RelatedEntityType string                 `json:"related_entity_type,omitempty"`
	RelatedEntityID   string                 `json:"related_entity_id,omitempty"`
	AutonomyLevel     string                 `json:"autonomy_level,omitempty"`
	InitialContext    map[string]interface{} `json:"initial_context,omitempty"`
	CorrelationID     string                 `json:"correlation_id,omitempty"`
	IdempotencyKey    string                 `json:"idempotency_key,omitempty"`
}

type PauseWorkflowRequest struct {
	Reason string `json:"reason"`
}

type CancelWorkflowRequest struct {
	Reason string `json:"reason"`
}

type ApproveStepRequest struct {
	Notes string `json:"notes,omitempty"`
}

type RejectStepRequest struct {
	Reason string `json:"reason"`
}

type TriggerBusinessEventRequest struct {
	EventType      string                 `json:"event_type"`
	EventID        string                 `json:"event_id"`
	EntityType     string                 `json:"entity_type"`
	EntityID       string                 `json:"entity_id"`
	Payload        map[string]interface{} `json:"payload"`
	CorrelationID  string                 `json:"correlation_id,omitempty"`
	IdempotencyKey string                 `json:"idempotency_key,omitempty"`
}

type WorkflowFilter struct {
	OrgID        int64
	WorkflowType string
	State        string
	Limit        int
	Offset       int
}

type EmergencyControlRequest struct {
	Scope  string `json:"scope"`  // GLOBAL, TENANT, WORKFLOW, AGENT
	Action string `json:"action"` // PAUSE, RESUME
	Reason string `json:"reason"`
}

type RecoveryReport struct {
	ScannedCount   int      `json:"scanned_count"`
	RecoveredCount int      `json:"recovered_count"`
	BlockedCount   int      `json:"blocked_count"`
	ResumedIDs     []string `json:"resumed_ids"`
	BlockedIDs     []string `json:"blocked_ids"`
	Errors         []string `json:"errors,omitempty"`
}

type EnterprisePlatformOverview struct {
	OrgID               int64                  `json:"org_id"`
	ActiveWorkflows     int                    `json:"active_workflows"`
	PausedWorkflows     int                    `json:"paused_workflows"`
	WaitingApprovals    int                    `json:"waiting_approvals"`
	EscalatedWorkflows  int                    `json:"escalated_workflows"`
	FailedWorkflows     int                    `json:"failed_workflows"`
	CompletedWorkflows  int                    `json:"completed_workflows"`
	AutonomyLevel       string                 `json:"autonomy_level"`
	EmergencyStopActive bool                   `json:"emergency_stop_active"`
	PlatformHealth      string                 `json:"platform_health"`
	RecentWorkflows     []*EnterpriseWorkflow  `json:"recent_workflows"`
}
