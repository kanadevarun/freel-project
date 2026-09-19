package enterprise_autonomy

import (
	"errors"
	"time"
)

// ---------------------------------------------------------------------
// Phase 7.10: Enterprise Autonomy, Governance & Safety Models & DTOs
// ---------------------------------------------------------------------

// AutonomyLevel constants
type AutonomyLevel string

const (
	AutonomyLvl0Observe    = "LEVEL_0_OBSERVE"
	AutonomyLvl1Recommend  = "LEVEL_1_RECOMMEND"
	AutonomyLvl2Prepare    = "LEVEL_2_PREPARE"
	AutonomyLvl3Controlled = "LEVEL_3_CONTROLLED"
	AutonomyLvl4Governed   = "LEVEL_4_GOVERNED"
)

// ActionRiskLevel defines standardized risk tiers enforced authoritatively by Go
type ActionRiskLevel string

const (
	RiskLevelLow      ActionRiskLevel = "LOW"
	RiskLevelMedium   ActionRiskLevel = "MEDIUM"
	RiskLevelHigh     ActionRiskLevel = "HIGH"
	RiskLevelCritical ActionRiskLevel = "CRITICAL"
)

// Standardized Governance Errors (Fail-Closed)
var (
	ErrGovernanceDenied         = errors.New("governance policy check denied action execution")
	ErrEmergencyHaltActive      = errors.New("autonomous execution halted: emergency stop is active")
	ErrAgentPrivilegeViolation  = errors.New("agent unauthorized: requested action violates least-privilege matrix")
	ErrAutonomousLoopDetected   = errors.New("autonomous loop detected: repetitive cycle threshold exceeded")
	ErrStaleApprovalInvalidated = errors.New("approval invalidated: entity state changed materially since approval granted")
	ErrHumanRejectionConflict   = errors.New("action blocked: human operator previously rejected this action on entity")
	ErrInsufficientConfidence   = errors.New("action blocked: confidence score is below required safety policy threshold")
	ErrPolicyVersionMismatch    = errors.New("policy version mismatch or obsolete governance context")
)

// EmergencyScope defines the granularity of emergency halt controls
type EmergencyScope string

const (
	EmergencyScopeAll          EmergencyScope = "ALL"
	EmergencyScopeAgent        EmergencyScope = "AGENT"
	EmergencyScopeWorkflowType EmergencyScope = "WORKFLOW_TYPE"
	EmergencyScopeActionType   EmergencyScope = "ACTION_TYPE"
	EmergencyScopeWorkflowID   EmergencyScope = "WORKFLOW_ID"
)

// EmergencyControlEntry stores persistent or in-memory emergency halts
type EmergencyControlEntry struct {
	OrgID            int64          `json:"org_id"`
	Scope            EmergencyScope `json:"scope"`
	TargetIdentifier string         `json:"target_identifier"` // AgentID, WorkflowType, ActionType, or "*"
	Active           bool           `json:"active"`
	Reason           string         `json:"reason"`
	HaltedBy         string         `json:"halted_by"`
	HaltedAt         time.Time      `json:"halted_at"`
}

// EnterpriseGovernedAction encapsulates all metadata required for governed execution
type EnterpriseGovernedAction struct {
	ActionID           string                 `json:"action_id"`
	OrgID              int64                  `json:"org_id"`
	WorkflowID         string                 `json:"workflow_id"`
	AgentID            string                 `json:"agent_id"`
	ActionType         string                 `json:"action_type"`
	TargetEntityType   string                 `json:"target_entity_type"`
	TargetEntityID     string                 `json:"target_entity_id"`
	ProposedParameters map[string]interface{} `json:"proposed_parameters"`
	RiskClassification ActionRiskLevel        `json:"risk_classification"`
	AutonomyLevel      string                 `json:"autonomy_level"`
	PolicyVersion      string                 `json:"policy_version"`
	RequiresApproval   bool                   `json:"requires_approval"`
	ApprovalID         *int64                 `json:"approval_id,omitempty"`
	StateFingerprint   string                 `json:"state_fingerprint"`
	ExecutionStatus    string                 `json:"execution_status"` // PERMITTED, BLOCKED, WAITING_APPROVAL, EXECUTED, FAILED
	VerificationResult string                 `json:"verification_result,omitempty"`
	CorrelationID      string                 `json:"correlation_id"`
	CausationID        string                 `json:"causation_id"`
	ExecutionAttempt   int                    `json:"execution_attempt"`
	CreatedAt          time.Time              `json:"created_at"`
	ExecutedAt         *time.Time             `json:"executed_at,omitempty"`
}

// GovernanceEvaluationRequest is sent by workflow engines or agents to request permission
type GovernanceEvaluationRequest struct {
	OrgID               int64                  `json:"org_id"`
	WorkflowID          string                 `json:"workflow_id"`
	WorkflowType        string                 `json:"workflow_type"`
	AgentID             string                 `json:"agent_id"`
	ActionType          string                 `json:"action_type"`
	TargetEntityType    string                 `json:"target_entity_type"`
	TargetEntityID      string                 `json:"target_entity_id"`
	MonetaryExposure    float64                `json:"monetary_exposure"`
	CurrentEntityState  map[string]interface{} `json:"current_entity_state"`
	Parameters          map[string]interface{} `json:"parameters"`
	ProposedRiskLevel     ActionRiskLevel        `json:"proposed_risk_level"`
	ProposedAutonomyLevel AutonomyLevel          `json:"proposed_autonomy_level,omitempty"`
	Confidence            float64                `json:"confidence"`
	UntrustedContent    []string               `json:"untrusted_content,omitempty"`
	CorrelationID       string                 `json:"correlation_id"`
	CausationID         string                 `json:"causation_id"`
}

// GovernanceDecision is the authoritative verdict generated by Go
type GovernanceDecision struct {
	Decision               string          `json:"decision"` // PERMITTED, BLOCKED, REQUIRE_APPROVAL
	Allowed                bool            `json:"allowed"`
	RequiresApproval       bool            `json:"requires_approval"`
	ApprovalID             *int64          `json:"approval_id,omitempty"`
	AuthoritativeRiskLevel ActionRiskLevel `json:"authoritative_risk_level"`
	PolicyVersion          string          `json:"policy_version"`
	Reason                 string          `json:"reason"`
	EvaluatedAutonomyLevel string          `json:"evaluated_autonomy_level"`
	StateFingerprint       string          `json:"state_fingerprint"`
	CircuitBreakerTripped  bool            `json:"circuit_breaker_tripped"`
	Timestamp              time.Time       `json:"timestamp"`
}

// HumanRejectionRecord records human negative decisions to prevent agent override
type HumanRejectionRecord struct {
	OrgID      int64     `json:"org_id"`
	WorkflowID string    `json:"workflow_id"`
	ActionType string    `json:"action_type"`
	EntityID   string    `json:"entity_id"`
	Reason     string    `json:"reason"`
	RejectedBy string    `json:"rejected_by"`
	RejectedAt time.Time `json:"rejected_at"`
}

// GovernanceStatusSummary exposes high-level health and safety metrics for Control Tower
type GovernanceStatusSummary struct {
	OrgID                   int64                   `json:"org_id"`
	PolicyVersion           string                  `json:"policy_version"`
	EmergencyHaltActive     bool                    `json:"emergency_halt_active"`
	ActiveHaltCount         int                     `json:"active_halt_count"`
	TotalActionsEvaluated   int64                   `json:"total_actions_evaluated"`
	BlockedActionsCount     int64                   `json:"blocked_actions_count"`
	ApprovalsRequiredCount  int64                   `json:"approvals_required_count"`
	StaleApprovalsPrevented int64                   `json:"stale_approvals_prevented"`
	LoopsDetectedCount      int64                   `json:"loops_detected_count"`
	InjectionAttemptsCount  int64                   `json:"injection_attempts_count"`
	ActiveEmergencyControls []EmergencyControlEntry `json:"active_emergency_controls"`
	LastEvaluatedAt         time.Time               `json:"last_evaluated_at"`
}
