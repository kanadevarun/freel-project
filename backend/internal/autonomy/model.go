package autonomy

import (
	"database/sql"
	"encoding/json"
	"time"
)

// AutonomyLevel defines the 5 tiered operating levels of controlled autonomy
type AutonomyLevel string

const (
	Level0Observe             AutonomyLevel = "LEVEL_0_OBSERVE"
	Level1Recommend           AutonomyLevel = "LEVEL_1_RECOMMEND"
	Level2Prepare             AutonomyLevel = "LEVEL_2_PREPARE"
	Level3ControlledExecution AutonomyLevel = "LEVEL_3_CONTROLLED_EXECUTION"
	Level4ControlledMultiStep AutonomyLevel = "LEVEL_4_CONTROLLED_MULTI_STEP"
)

// PlanStatus defines the durable lifecycle statuses of an operational plan
type PlanStatus string

const (
	PlanStatusDraft              PlanStatus = "DRAFT"
	PlanStatusGenerated          PlanStatus = "GENERATED"
	PlanStatusValidating         PlanStatus = "VALIDATING"
	PlanStatusRequiresApproval   PlanStatus = "REQUIRES_APPROVAL"
	PlanStatusApproved           PlanStatus = "APPROVED"
	PlanStatusRejected           PlanStatus = "REJECTED"
	PlanStatusExecuting          PlanStatus = "EXECUTING"
	PlanStatusPaused             PlanStatus = "PAUSED"
	PlanStatusWaiting            PlanStatus = "WAITING"
	PlanStatusCompleted          PlanStatus = "COMPLETED"
	PlanStatusPartiallyCompleted PlanStatus = "PARTIALLY_COMPLETED"
	PlanStatusFailed             PlanStatus = "FAILED"
	PlanStatusCancelled          PlanStatus = "CANCELLED"
	PlanStatusExpired            PlanStatus = "EXPIRED"
	PlanStatusReplanning         PlanStatus = "REPLANNING"
	PlanStatusSuperseded         PlanStatus = "SUPERSEDED"
)

// StepStatus defines the lifecycle status of an individual execution step
type StepStatus string

const (
	StepStatusPending          StepStatus = "PENDING"
	StepStatusReady            StepStatus = "READY"
	StepStatusBlocked          StepStatus = "BLOCKED"
	StepStatusAwaitingApproval StepStatus = "AWAITING_APPROVAL"
	StepStatusApproved         StepStatus = "APPROVED"
	StepStatusRejected         StepStatus = "REJECTED"
	StepStatusExecuting        StepStatus = "EXECUTING"
	StepStatusWaiting          StepStatus = "WAITING"
	StepStatusCompleted        StepStatus = "COMPLETED"
	StepStatusSucceeded        StepStatus = "SUCCEEDED"
	StepStatusFailed           StepStatus = "FAILED"
	StepStatusSkipped          StepStatus = "SKIPPED"
	StepStatusCancelled        StepStatus = "CANCELLED"
	StepStatusReplanning       StepStatus = "REPLANNING"
)

// PolicyDecision describes the outcome of policy evaluation
type PolicyDecision string

const (
	DecisionPermitted            PolicyDecision = "PERMITTED"
	DecisionRequiresApproval     PolicyDecision = "REQUIRES_APPROVAL"
	DecisionBlockedPolicy        PolicyDecision = "BLOCKED_POLICY"
	DecisionBlockedEmergencyStop PolicyDecision = "BLOCKED_EMERGENCY_STOP"
)

type EventDecision string

const (
	EventDecisionNoAction            EventDecision = "NO_ACTION"
	EventDecisionContinueMonitoring  EventDecision = "CONTINUE_MONITORING"
	EventDecisionRecommendation      EventDecision = "RECOMMENDATION"
	EventDecisionNewPlan             EventDecision = "NEW_PLAN"
	EventDecisionApproval            EventDecision = "APPROVAL"
	EventDecisionControlledExecution EventDecision = "CONTROLLED_EXECUTION"
	EventDecisionEscalation          EventDecision = "ESCALATION"
	EventDecisionReplanning          EventDecision = "REPLANNING"
)

type WaitingState string

const (
	WaitingStateNone          WaitingState = "NONE"
	WaitingStateCarrier       WaitingState = "WAITING_FOR_CARRIER"
	WaitingStateApproval      WaitingState = "WAITING_FOR_APPROVAL"
	WaitingStateCustomer      WaitingState = "WAITING_FOR_CUSTOMER"
	WaitingStateMilestone     WaitingState = "WAITING_FOR_MILESTONE"
	WaitingStateExternalEvent WaitingState = "WAITING_FOR_EXTERNAL_EVENT"
	WaitingStateVerification  WaitingState = "WAITING_FOR_VERIFICATION"
)

// AutonomyPolicy represents tenant-specific autonomy rules and boundaries
type AutonomyPolicy struct {
	ID                     int64           `db:"id" json:"id"`
	OrgID                  int64           `db:"org_id" json:"org_id"`
	Module                 string          `db:"module" json:"module"`
	AutonomyLevel          AutonomyLevel   `db:"autonomy_level" json:"autonomy_level"`
	AllowedActionTypes     json.RawMessage `db:"allowed_action_types" json:"allowed_action_types"`
	ProhibitedActionTypes  json.RawMessage `db:"prohibited_action_types" json:"prohibited_action_types"`
	RequiresApproval       bool            `db:"requires_approval" json:"requires_approval"`
	MaxMonetaryThreshold   float64         `db:"max_monetary_threshold" json:"max_monetary_threshold"`
	CustomerImpactLimit    string          `db:"customer_impact_threshold" json:"customer_impact_threshold"`
	ShipmentImpactLimit    string          `db:"shipment_impact_threshold" json:"shipment_impact_threshold"`
	ComplianceSensitivity  string          `db:"compliance_sensitivity" json:"compliance_sensitivity"`
	MinConfidenceThreshold float64         `db:"min_confidence_threshold" json:"min_confidence_threshold"`
	RequireDataSufficiency bool            `db:"require_data_sufficiency" json:"require_data_sufficiency"`
	MaxPlanSteps           int             `db:"max_plan_steps" json:"max_plan_steps"`
	MaxExecutionAttempts   int             `db:"max_execution_attempts" json:"max_execution_attempts"`
	CooldownSeconds        int             `db:"cooldown_seconds" json:"cooldown_seconds"`
	EmergencyStop          bool            `db:"emergency_stop" json:"emergency_stop"`
	IsActive               bool            `db:"is_active" json:"is_active"`
	PolicyVersion          int             `db:"policy_version" json:"policy_version"`
	CreatedAt              time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time       `db:"updated_at" json:"updated_at"`
}

// PlanningGoal represents a structured operational objective
type PlanningGoal struct {
	ID                  int64           `db:"id" json:"id"`
	OrgID               int64           `db:"org_id" json:"org_id"`
	UserID              sql.NullInt64   `db:"user_id" json:"user_id"`
	GoalID              string          `db:"goal_id" json:"goal_id"`
	CorrelationID       string          `db:"correlation_id" json:"correlation_id"`
	Source              string          `db:"source" json:"source"`
	Module              string          `db:"module" json:"module"`
	RelatedEntityType   string          `db:"related_entity_type" json:"related_entity_type"`
	RelatedEntityID     string          `db:"related_entity_id" json:"related_entity_id"`
	Objective           string          `db:"objective" json:"objective"`
	Priority            string          `db:"priority" json:"priority"`
	Deadline            sql.NullTime    `db:"deadline" json:"deadline"`
	HardConstraints     json.RawMessage `db:"hard_constraints" json:"hard_constraints"`
	SoftConstraints     json.RawMessage `db:"soft_constraints" json:"soft_constraints"`
	SuccessCriteria     sql.NullString  `db:"success_criteria" json:"success_criteria"`
	RiskTolerance       string          `db:"risk_tolerance" json:"risk_tolerance"`
	AutonomyLevel       AutonomyLevel   `db:"autonomy_level" json:"autonomy_level"`
	RequiredPermissions json.RawMessage `db:"required_permissions" json:"required_permissions"`
	Status              string          `db:"status" json:"status"`
	CreatedAt           time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time       `db:"updated_at" json:"updated_at"`
}

// AutonomousPlan represents a durable operational plan with candidate alternatives
type AutonomousPlan struct {
	ID                  int64           `db:"id" json:"id"`
	OrgID               int64           `db:"org_id" json:"org_id"`
	UserID              sql.NullInt64   `db:"user_id" json:"user_id"`
	PlanID              string          `db:"plan_id" json:"plan_id"`
	Version             int             `db:"version" json:"version"`
	ParentPlanID        sql.NullString  `db:"parent_plan_id" json:"parent_plan_id"`
	CorrelationID       string          `db:"correlation_id" json:"correlation_id"`
	GoalID              sql.NullString  `db:"goal_id" json:"goal_id"`
	Goal                string          `db:"goal" json:"goal"`
	GoalType            string          `db:"goal_type" json:"goal_type"`
	Module              string          `db:"module" json:"module"`
	RelatedEntityType   string          `db:"related_entity_type" json:"related_entity_type"`
	RelatedEntityID     string          `db:"related_entity_id" json:"related_entity_id"`
	CurrentStateSumm    string          `db:"current_state_summary" json:"current_state_summary"`
	Constraints         json.RawMessage `db:"constraints" json:"constraints"`
	HardConstraints     json.RawMessage `db:"hard_constraints" json:"hard_constraints"`
	SoftConstraints     json.RawMessage `db:"soft_constraints" json:"soft_constraints"`
	Assumptions         json.RawMessage `db:"assumptions" json:"assumptions"`
	Risks               json.RawMessage `db:"risks" json:"risks"`
	CandidatePlans      json.RawMessage `db:"candidate_plans" json:"candidate_plans"`
	SelectedCandidateID sql.NullString  `db:"selected_candidate_id" json:"selected_candidate_id"`
	EvaluationSummary   json.RawMessage `db:"evaluation_summary" json:"evaluation_summary"`
	ConfidenceScore     float64         `db:"confidence_score" json:"confidence_score"`
	DataSufficiency     bool            `db:"data_sufficiency" json:"data_sufficiency"`
	EstimatedImpact     sql.NullString  `db:"estimated_impact" json:"estimated_impact"`
	RiskLevel           string          `db:"risk_level" json:"risk_level"`
	Priority            string          `db:"priority" json:"priority"`
	AutonomyLevel       AutonomyLevel   `db:"autonomy_level" json:"autonomy_level"`
	PolicyDecision      PolicyDecision  `db:"policy_decision" json:"policy_decision"`
	PolicyReason        sql.NullString  `db:"policy_reason" json:"policy_reason"`
	Status              PlanStatus      `db:"status" json:"status"`
	CurrentStepID       sql.NullString  `db:"current_step_id" json:"current_step_id"`
	ReplanStatus        string          `db:"replan_status" json:"replan_status"`
	ReplanReason        sql.NullString  `db:"replan_reason" json:"replan_reason"`
	StopConditions      json.RawMessage `db:"stop_conditions" json:"stop_conditions"`
	FallbackStrategy    sql.NullString  `db:"fallback_strategy" json:"fallback_strategy"`
	TriggeringEvent     sql.NullString  `db:"triggering_event" json:"triggering_event"`
	ExecutionStatus        string         `db:"execution_status" json:"execution_status"`
	WaitingState           sql.NullString `db:"waiting_state" json:"waiting_state"`
	WaitingUntil           sql.NullTime   `db:"waiting_until" json:"waiting_until"`
	EscalationReason       sql.NullString `db:"escalation_reason" json:"escalation_reason"`
	CustomerCommitmentDate sql.NullTime   `db:"customer_commitment_date" json:"customer_commitment_date"`
	PredictedETA           sql.NullTime   `db:"predicted_eta" json:"predicted_eta"`
	ETADeviationHours      float64        `db:"eta_deviation_hours" json:"eta_deviation_hours"`
	CommitmentRiskSeverity string         `db:"commitment_risk_severity" json:"commitment_risk_severity"`
	VerificationStatus     string          `db:"verification_status" json:"verification_status"`
	StalenessStatus        string          `db:"staleness_status" json:"staleness_status"`
	ExpiresAt              sql.NullTime    `db:"expires_at" json:"expires_at"`
	PlanHealth             PlanHealthState `db:"plan_health" json:"plan_health"`
	HealthReason           sql.NullString  `db:"health_reason" json:"health_reason"`
	ChangedAssumptions     json.RawMessage `db:"changed_assumptions" json:"changed_assumptions"`
	ReplanCount            int             `db:"replan_count" json:"replan_count"`
	LastMonitoredAt        sql.NullTime    `db:"last_monitored_at" json:"last_monitored_at"`
	OperatingMode          sql.NullString  `db:"operating_mode" json:"operating_mode"`
	StoppedByUserID        sql.NullInt64   `db:"stopped_by_user_id" json:"stopped_by_user_id"`
	StoppedAt              sql.NullTime    `db:"stopped_at" json:"stopped_at"`
	StopReason             sql.NullString  `db:"stop_reason" json:"stop_reason"`
	HumanModified          bool            `db:"human_modified" json:"human_modified"`
	CreatedAt              time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time       `db:"updated_at" json:"updated_at"`
}

// AutonomousPlanStep represents an individual ordered step of a plan
type AutonomousPlanStep struct {
	ID                   int64           `db:"id" json:"id"`
	PlanID               string          `db:"plan_id" json:"plan_id"`
	OrgID                int64           `db:"org_id" json:"org_id"`
	StepNumber           int             `db:"step_number" json:"step_number"`
	StepID               string          `db:"step_id" json:"step_id"`
	ActionType           string          `db:"action_type" json:"action_type"`
	Title                string          `db:"title" json:"title"`
	Description          string          `db:"description" json:"description"`
	Parameters           json.RawMessage `db:"parameters" json:"parameters"`
	Dependencies         json.RawMessage `db:"dependencies" json:"dependencies"`
	ConditionPredicate   json.RawMessage `db:"condition_predicate" json:"condition_predicate"`
	Preconditions        json.RawMessage `db:"preconditions" json:"preconditions"`
	ExpectedOutcome      string          `db:"expected_outcome" json:"expected_outcome"`
	VerificationCriteria json.RawMessage `db:"verification_criteria" json:"verification_criteria"`
	RiskLevel            string          `db:"risk_level" json:"risk_level"`
	Reversibility        string          `db:"reversibility" json:"reversibility"`
	FallbackAction       json.RawMessage `db:"fallback_action" json:"fallback_action"`
	CompensationAction   json.RawMessage `db:"compensation_action" json:"compensation_action"`
	TimeoutSeconds       int             `db:"timeout_seconds" json:"timeout_seconds"`
	RequiresApproval     bool            `db:"requires_approval" json:"requires_approval"`
	ActionSystemActionID sql.NullString  `db:"action_system_action_id" json:"action_system_action_id"`
	IdempotencyKey       string          `db:"idempotency_key" json:"idempotency_key"`
	Status               StepStatus      `db:"status" json:"status"`
	ExecutionAttempt     int             `db:"execution_attempt" json:"execution_attempt"`
	MaxAttempts          int             `db:"max_attempts" json:"max_attempts"`
	RetryPolicy          json.RawMessage `db:"retry_policy" json:"retry_policy"`
	ExecutedAt           sql.NullTime    `db:"executed_at" json:"executed_at"`
	ExecutionResult      json.RawMessage `db:"execution_result" json:"execution_result"`
	ErrorMessage         sql.NullString  `db:"error_message" json:"error_message"`
	VerificationStatus   string          `db:"verification_status" json:"verification_status"`
	VerificationDetails  json.RawMessage `db:"verification_details" json:"verification_details"`
	HumanEditedContent   sql.NullString  `db:"human_edited_content" json:"human_edited_content"`
	CreatedAt            time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time       `db:"updated_at" json:"updated_at"`
}

// AutonomousPlanAuditHistory records every state transition and policy decision
type AutonomousPlanAuditHistory struct {
	ID             int64           `db:"id" json:"id"`
	PlanID         string          `db:"plan_id" json:"plan_id"`
	OrgID          int64           `db:"org_id" json:"org_id"`
	UserID         sql.NullInt64   `db:"user_id" json:"user_id"`
	EventType      string          `db:"event_type" json:"event_type"`
	PreviousStatus sql.NullString  `db:"previous_status" json:"previous_status"`
	NewStatus      string          `db:"new_status" json:"new_status"`
	StepID         sql.NullString  `db:"step_id" json:"step_id"`
	Details        json.RawMessage `db:"details" json:"details"`
	Notes          sql.NullString  `db:"notes" json:"notes"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
}

// OperationalMemory stores structured operational learnings across executions
type OperationalMemory struct {
	ID                int64           `db:"id" json:"id"`
	OrgID             int64           `db:"org_id" json:"org_id"`
	MemoryType        string          `db:"memory_type" json:"memory_type"`
	EntityType        string          `db:"entity_type" json:"entity_type"`
	EntityID          string          `db:"entity_id" json:"entity_id"`
	Summary           string          `db:"summary" json:"summary"`
	StructuredPayload json.RawMessage `db:"structured_payload" json:"structured_payload"`
	SuccessRating     float64         `db:"success_rating" json:"success_rating"`
	UsageCount        int             `db:"usage_count" json:"usage_count"`
	CreatedAt         time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time       `db:"updated_at" json:"updated_at"`
}

// DTO Requests & Responses
type ConstraintDTO struct {
	ConstraintType string      `json:"constraint_type"`
	Description    string      `json:"description"`
	IsHard         bool        `json:"is_hard"`
	FieldTarget    string      `json:"field_target,omitempty"`
	Operator       string      `json:"operator,omitempty"`
	ThresholdValue interface{} `json:"threshold_value,omitempty"`
}

type CreateGoalRequest struct {
	Source            string          `json:"source"`
	Module            string          `json:"module"`
	RelatedEntityType string          `json:"related_entity_type"`
	RelatedEntityID   string          `json:"related_entity_id"`
	Objective         string          `json:"objective"`
	Priority          string          `json:"priority"`
	Deadline          *time.Time      `json:"deadline,omitempty"`
	HardConstraints   []ConstraintDTO `json:"hard_constraints"`
	SoftConstraints   []ConstraintDTO `json:"soft_constraints"`
	SuccessCriteria   string          `json:"success_criteria"`
	RiskTolerance     string          `json:"risk_tolerance"`
	AutonomyLevel     AutonomyLevel   `json:"autonomy_level"`
	CorrelationID     string          `json:"correlation_id,omitempty"`
}

type GeneratePlanRequest struct {
	GoalID            string                 `json:"goal_id,omitempty"`
	Goal              string                 `json:"goal"`
	Module            string                 `json:"module"`
	RelatedEntityType string                 `json:"related_entity_type"`
	RelatedEntityID   string                 `json:"related_entity_id"`
	CurrentState      map[string]interface{} `json:"current_state"`
	Constraints       []string               `json:"constraints"`
	HardConstraints   []ConstraintDTO        `json:"hard_constraints"`
	SoftConstraints   []ConstraintDTO        `json:"soft_constraints"`
	AutonomyLevel     AutonomyLevel          `json:"autonomy_level"`
	RiskTolerance     string                 `json:"risk_tolerance,omitempty"`
	CorrelationID     string                 `json:"correlation_id"`
}

type SetPolicyRequest struct {
	Module                 string        `json:"module"`
	AutonomyLevel          AutonomyLevel `json:"autonomy_level"`
	AllowedActionTypes     []string      `json:"allowed_action_types"`
	ProhibitedActionTypes  []string      `json:"prohibited_action_types"`
	RequiresApproval       bool          `json:"requires_approval"`
	MaxMonetaryThreshold   float64       `json:"max_monetary_threshold"`
	CustomerImpactLimit    string        `json:"customer_impact_threshold"`
	ShipmentImpactLimit    string        `json:"shipment_impact_threshold"`
	ComplianceSensitivity  string        `json:"compliance_sensitivity"`
	MinConfidenceThreshold float64       `json:"min_confidence_threshold"`
	RequireDataSufficiency bool          `json:"require_data_sufficiency"`
	MaxPlanSteps           int           `json:"max_plan_steps"`
	MaxExecutionAttempts   int           `json:"max_execution_attempts"`
	CooldownSeconds        int           `json:"cooldown_seconds"`
	EmergencyStop          bool          `json:"emergency_stop"`
	IsActive               bool          `json:"is_active"`
}

type StepExecutionResult struct {
	PlanID             string      `json:"plan_id"`
	StepID             string      `json:"step_id"`
	StepStatus         StepStatus  `json:"step_status"`
	ActionID           string      `json:"action_id"`
	ExecutionResult    interface{} `json:"execution_result"`
	VerificationStatus string      `json:"verification_status"`
	ErrorMessage       string      `json:"error_message,omitempty"`
}

type ReplanRequest struct {
	ReplanReason    string                 `json:"replan_reason"`
	TriggeringEvent string                 `json:"triggering_event"`
	ObservedState   map[string]interface{} `json:"observed_new_state"`
}

type SelectCandidateRequest struct {
	CandidateID string `json:"candidate_id"`
	Reason      string `json:"reason,omitempty"`
}

type RevalidatePlanRequest struct {
	ForceRefresh bool `json:"force_refresh"`
}

// ── Phase 5 Task 5.3: Adaptive Shipment Management Structures ────────────────

type ShipmentAdaptiveEvent struct {
	ID               int64          `db:"id" json:"id"`
	OrgID            int64          `db:"org_id" json:"org_id"`
	ShipmentID       int64          `db:"shipment_id" json:"shipment_id"`
	EventID          string         `db:"event_id" json:"event_id"`
	EventType        string         `db:"event_type" json:"event_type"`
	CorrelationID    string         `db:"correlation_id" json:"correlation_id"`
	DeduplicationKey string         `db:"deduplication_key" json:"deduplication_key"`
	Severity         string         `db:"severity" json:"severity"`
	Payload          json.RawMessage `db:"payload" json:"payload"`
	Decision         string         `db:"decision" json:"decision"`
	DecisionReason   string         `db:"decision_reason" json:"decision_reason"`
	PlanID           string         `db:"plan_id" json:"plan_id"`
	CreatedAt        time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at" json:"updated_at"`
}

type IngestShipmentEventRequest struct {
	EventID                string                 `json:"event_id"`
	ShipmentID             int64                  `json:"shipment_id"`
	EventType              string                 `json:"event_type"`
	Severity               string                 `json:"severity"`
	Payload                map[string]interface{} `json:"payload"`
	CorrelationID          string                 `json:"correlation_id"`
	DeduplicationKey       string                 `json:"deduplication_key"`
	PredictedETA           *string                `json:"predicted_eta,omitempty"`
	CustomerCommitmentDate *string                `json:"customer_commitment_date,omitempty"`
}

type ShipmentEventEvaluationResult struct {
	EventID                string          `json:"event_id"`
	ShipmentID             int64           `json:"shipment_id"`
	EventType              string          `json:"event_type"`
	Decision               string          `json:"decision"`
	DecisionReason         string          `json:"decision_reason"`
	IsMeaningfulChange     bool            `json:"is_meaningful_change"`
	ETADeviationHours      float64         `json:"eta_deviation_hours"`
	CommitmentRiskSeverity string          `json:"commitment_risk_severity"`
	RecommendedActionType  string          `json:"recommended_action_type,omitempty"`
	EscalationReason       string          `json:"escalation_reason,omitempty"`
	PlanID                 string          `json:"plan_id,omitempty"`
	ActivePlanID           string          `json:"active_plan_id,omitempty"`
	WaitingState           string          `json:"waiting_state,omitempty"`
	Plan                   *AutonomousPlan `json:"plan,omitempty"`
	Steps                  []AutonomousPlanStep `json:"steps,omitempty"`
	CreatedAt              time.Time       `json:"created_at"`
}

type ShipmentAdaptiveStateResponse struct {
	ShipmentID             int64                   `json:"shipment_id"`
	OrgID                  int64                   `json:"org_id"`
	Status                 string                  `json:"status"`
	CurrentRiskLevel       string                  `json:"current_risk_level"`
	AdaptiveStatus         string                  `json:"adaptive_status"`
	ETA                    *time.Time              `json:"eta,omitempty"`
	PredictedETA           *time.Time              `json:"predicted_eta,omitempty"`
	CustomerCommitmentDate *time.Time              `json:"customer_commitment_date,omitempty"`
	ETADeviationHours      float64                 `json:"eta_deviation_hours"`
	CommitmentRiskSeverity string                  `json:"commitment_risk_severity"`
	ActivePlan             *AutonomousPlan         `json:"active_plan,omitempty"`
	ActivePlanSteps        []AutonomousPlanStep    `json:"active_plan_steps,omitempty"`
	RecentEvents           []ShipmentAdaptiveEvent `json:"recent_events,omitempty"`
	RecentMilestones       []map[string]interface{}`json:"recent_milestones,omitempty"`
	ActiveExceptions       []map[string]interface{}`json:"active_exceptions,omitempty"`
}

// -----------------------------------------------------------------------------
// Phase 5 Task 5.4: Autonomous Customer Follow-Up Models
// -----------------------------------------------------------------------------

type CustomerCommunicationPreferences struct {
	ID                       int64     `db:"id" json:"id"`
	OrgID                    int64     `db:"org_id" json:"org_id"`
	CustomerID               int64     `db:"customer_id" json:"customer_id"`
	PreferredChannel         string    `db:"preferred_channel" json:"preferred_channel"`
	OptOut                   bool      `db:"opt_out" json:"opt_out"`
	OptOutReason             *string   `db:"opt_out_reason" json:"opt_out_reason,omitempty"`
	ContactRestrictions      string    `db:"contact_restrictions" json:"contact_restrictions"`
	BusinessHoursOnly        bool      `db:"business_hours_only" json:"business_hours_only"`
	DesignatedContactID      *int64    `db:"designated_contact_id" json:"designated_contact_id,omitempty"`
	MaxFollowupsPerIncident  int       `db:"max_followups_per_incident" json:"max_followups_per_incident"`
	MinFollowupIntervalHours int       `db:"min_followup_interval_hours" json:"min_followup_interval_hours"`
	CreatedAt                time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                time.Time `db:"updated_at" json:"updated_at"`
}

type CustomerFollowupRecord struct {
	ID                     int64      `db:"id" json:"id"`
	OrgID                  int64      `db:"org_id" json:"org_id"`
	CustomerID             int64      `db:"customer_id" json:"customer_id"`
	ContactID              *int64     `db:"contact_id" json:"contact_id,omitempty"`
	PlanID                 *string    `db:"plan_id" json:"plan_id,omitempty"`
	StepID                 *string    `db:"step_id" json:"step_id,omitempty"`
	EventType              string     `db:"event_type" json:"event_type"`
	Channel                string     `db:"channel" json:"channel"`
	RecipientEmail         string     `db:"recipient_email" json:"recipient_email"`
	RecipientName          string     `db:"recipient_name" json:"recipient_name"`
	Subject                string     `db:"subject" json:"subject"`
	ActualFacts            *string    `db:"actual_facts" json:"actual_facts,omitempty"`
	Predictions            *string    `db:"predictions" json:"predictions,omitempty"`
	Recommendations        *string    `db:"recommendations" json:"recommendations,omitempty"`
	FullBody               string     `db:"full_body" json:"full_body"`
	Version                int        `db:"version" json:"version"`
	Status                 string     `db:"status" json:"status"`
	ApprovalID             *string    `db:"approval_id" json:"approval_id,omitempty"`
	ApprovalStatus         string     `db:"approval_status" json:"approval_status"`
	IdempotencyKey         string     `db:"idempotency_key" json:"idempotency_key"`
	SentAt                 *time.Time `db:"sent_at" json:"sent_at,omitempty"`
	CustomerResponse       *string    `db:"customer_response" json:"customer_response,omitempty"`
	ResponseReceivedAt     *time.Time `db:"response_received_at" json:"response_received_at,omitempty"`
	ResponseClassification *string    `db:"response_classification" json:"response_classification,omitempty"`
	StopReason             *string    `db:"stop_reason" json:"stop_reason,omitempty"`
	CreatedAt              time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at" json:"updated_at"`
}

type VerifiedContact struct {
	ContactID int64  `json:"contact_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
	JobTitle  string `json:"job_title,omitempty"`
	IsPrimary bool   `json:"is_primary"`
}

type CustomerFollowupDraft struct {
	Subject         string   `json:"subject"`
	ActualFacts     []string `json:"actual_facts"`
	Predictions     []string `json:"predictions"`
	Recommendations []string `json:"recommendations"`
	FullBody        string   `json:"full_body"`
	Channel         string   `json:"channel"`
}

type IngestCustomerFollowupEventRequest struct {
	CustomerID       int64                  `json:"customer_id"`
	ContactID        *int64                 `json:"contact_id,omitempty"`
	EventType        string                 `json:"event_type"`
	Payload          map[string]interface{} `json:"payload"`
	CorrelationID    string                 `json:"correlation_id"`
	DeduplicationKey string                 `json:"deduplication_key"`
}

type CustomerFollowupEventResult struct {
	RecordID               int64                  `json:"record_id,omitempty"`
	CustomerID             int64                  `json:"customer_id"`
	EventType              string                 `json:"event_type"`
	Decision               string                 `json:"decision"`
	DecisionReason         string                 `json:"decision_reason"`
	Urgency                string                 `json:"urgency"`
	Draft                  *CustomerFollowupDraft `json:"draft,omitempty"`
	Channel                string                 `json:"channel"`
	RequiresApproval       bool                   `json:"requires_approval"`
	ApprovalReason         string                 `json:"approval_reason,omitempty"`
	PlanID                 string                 `json:"plan_id,omitempty"`
	Plan                   *AutonomousPlan        `json:"plan,omitempty"`
	Steps                  []AutonomousPlanStep   `json:"steps,omitempty"`
	StopConditions         []string               `json:"stop_conditions,omitempty"`
	CreatedAt              time.Time              `json:"created_at"`
}

type CustomerFollowupStateResponse struct {
	CustomerID      int64                             `json:"customer_id"`
	OrgID           int64                             `json:"org_id"`
	CustomerName    string                            `json:"customer_name"`
	AccountTier     string                            `json:"account_tier"`
	FollowupStatus  string                            `json:"followup_status"`
	Preferences     *CustomerCommunicationPreferences `json:"preferences,omitempty"`
	PrimaryContact  *VerifiedContact                  `json:"primary_contact,omitempty"`
	Contacts        []VerifiedContact                 `json:"contacts,omitempty"`
	ActivePlan      *AutonomousPlan                   `json:"active_plan,omitempty"`
	ActivePlanSteps []AutonomousPlanStep              `json:"active_plan_steps,omitempty"`
	RecentRecords   []CustomerFollowupRecord          `json:"recent_records,omitempty"`
}

type UpdateCustomerPreferencesRequest struct {
	PreferredChannel         string  `json:"preferred_channel"`
	OptOut                   bool    `json:"opt_out"`
	OptOutReason             *string `json:"opt_out_reason,omitempty"`
	ContactRestrictions      string  `json:"contact_restrictions"`
	BusinessHoursOnly        bool    `json:"business_hours_only"`
	DesignatedContactID      *int64  `json:"designated_contact_id,omitempty"`
	MaxFollowupsPerIncident  int     `json:"max_followups_per_incident"`
	MinFollowupIntervalHours int     `json:"min_followup_interval_hours"`
}

type IngestCustomerResponseRequest struct {
	ResponseText string `json:"response_text"`
}

type ClassifyCustomerResponseResult struct {
	RecordID             int64  `json:"record_id"`
	CustomerID           int64  `json:"customer_id"`
	Classification       string `json:"classification"`
	Sentiment            string `json:"sentiment"`
	ActionRequested      string `json:"action_requested,omitempty"`
	RecommendedNextStep  string `json:"recommended_next_step"`
	Confidence           float64 `json:"confidence"`
	UpdatedPlanStatus    string `json:"updated_plan_status,omitempty"`
}

// -----------------------------------------------------------------------------
// Phase 5 Task 5.5: Intelligent RFQ and Pricing Optimization Models
// -----------------------------------------------------------------------------

type RfqPricingOptimization struct {
	ID                    int64     `db:"id" json:"id"`
	OrgID                 int64     `db:"org_id" json:"org_id"`
	RfqID                 int64     `db:"rfq_id" json:"rfq_id"`
	QuotationID           *int64    `db:"quotation_id" json:"quotation_id,omitempty"`
	PlanID                *string   `db:"plan_id" json:"plan_id,omitempty"`
	CurrentVersion        int       `db:"current_version" json:"current_version"`
	Status                string    `db:"status" json:"status"`
	Currency              string    `db:"currency" json:"currency"`
	BaseCost              float64   `db:"base_cost" json:"base_cost"`
	PredictedCost         float64   `db:"predicted_cost" json:"predicted_cost"`
	ActualFacts           *string   `db:"actual_facts" json:"actual_facts,omitempty"`
	Predictions           *string   `db:"predictions" json:"predictions,omitempty"`
	Assumptions           *string   `db:"assumptions" json:"assumptions,omitempty"`
	CandidateStrategies   string    `db:"candidate_strategies" json:"candidate_strategies"`
	RecommendedStrategyID string    `db:"recommended_strategy_id" json:"recommended_strategy_id"`
	RecommendedPrice      float64   `db:"recommended_price" json:"recommended_price"`
	RecommendedMarginPct  float64   `db:"recommended_margin_pct" json:"recommended_margin_pct"`
	TargetMarginPct       float64   `db:"target_margin_pct" json:"target_margin_pct"`
	MinMarginPct          float64   `db:"min_margin_pct" json:"min_margin_pct"`
	MarginRiskLevel       string    `db:"margin_risk_level" json:"margin_risk_level"`
	OperationalRiskLevel  string    `db:"operational_risk_level" json:"operational_risk_level"`
	ConfidenceScore       float64   `db:"confidence_score" json:"confidence_score"`
	DataSufficiency       string    `db:"data_sufficiency" json:"data_sufficiency"`
	RateFreshnessStatus   string    `db:"rate_freshness_status" json:"rate_freshness_status"`
	RateSource            string    `db:"rate_source" json:"rate_source"`
	RequiresApproval      bool      `db:"requires_approval" json:"requires_approval"`
	ApprovalReason        *string   `db:"approval_reason" json:"approval_reason,omitempty"`
	ApprovalStatus        string    `db:"approval_status" json:"approval_status"`
	ApprovalID            *string   `db:"approval_id" json:"approval_id,omitempty"`
	IdempotencyKey        string    `db:"idempotency_key" json:"idempotency_key"`
	ReasoningSummary      string    `db:"reasoning_summary" json:"reasoning_summary"`
	CreatedAt             time.Time `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time `db:"updated_at" json:"updated_at"`
}

type RfqPricingVersion struct {
	ID             int64     `db:"id" json:"id"`
	OrgID          int64     `db:"org_id" json:"org_id"`
	OptimizationID int64     `db:"optimization_id" json:"optimization_id"`
	RfqID          int64     `db:"rfq_id" json:"rfq_id"`
	QuotationID    *int64    `db:"quotation_id" json:"quotation_id,omitempty"`
	Version        int       `db:"version" json:"version"`
	StrategyName   string    `db:"strategy_name" json:"strategy_name"`
	Price          float64   `db:"price" json:"price"`
	Cost           float64   `db:"cost" json:"cost"`
	MarginPct      float64   `db:"margin_pct" json:"margin_pct"`
	ChangeReason   string    `db:"change_reason" json:"change_reason"`
	ApprovalStatus string    `db:"approval_status" json:"approval_status"`
	CreatedBy      string    `db:"created_by" json:"created_by"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

type PricingStrategyCandidateDTO struct {
	StrategyID            string  `json:"strategy_id"`
	StrategyType          string  `json:"strategy_type"`
	Title                 string  `json:"title"`
	Description           string  `json:"description"`
	Price                 float64 `json:"price"`
	BaseCost              float64 `json:"base_cost"`
	PredictedCost         float64 `json:"predicted_cost"`
	MarginPct             float64 `json:"margin_pct"`
	MarginAmount          float64 `json:"margin_amount"`
	OperationalRiskLevel  string  `json:"operational_risk_level"`
	MarginRiskLevel       string  `json:"margin_risk_level"`
	AcceptanceProbability float64 `json:"acceptance_probability"`
	ConfidenceScore       float64 `json:"confidence_score"`
	IsFeasible            bool    `json:"is_feasible"`
	InfeasibilityReason   *string `json:"infeasibility_reason,omitempty"`
	RequiresApproval      bool    `json:"requires_approval"`
	ApprovalReason        *string `json:"approval_reason,omitempty"`
	Score                 float64 `json:"score"`
}

type RfqPricingContextDTO struct {
	OrgID               int64                  `json:"org_id"`
	RfqID               int64                  `json:"rfq_id"`
	RfqNumber           string                 `json:"rfq_number"`
	CustomerID          int64                  `json:"customer_id"`
	CustomerName        string                 `json:"customer_name"`
	AccountTier         string                 `json:"account_tier"`
	Origin              string                 `json:"origin"`
	OriginCode          *string                `json:"origin_code,omitempty"`
	Destination         string                 `json:"destination"`
	DestinationCode     *string                `json:"destination_code,omitempty"`
	TransportMode       string                 `json:"transport_mode"`
	Incoterms           string                 `json:"incoterms"`
	CargoWeightKg       float64                `json:"cargo_weight_kg"`
	CargoVolumeCbm      float64                `json:"cargo_volume_cbm"`
	ContainerType       string                 `json:"container_type"`
	TargetDate          *string                `json:"target_date,omitempty"`
	RateBasis           map[string]interface{} `json:"rate_basis"`
	PricingPolicy       map[string]interface{} `json:"pricing_policy"`
	ExistingQuotation   map[string]interface{} `json:"existing_quotation,omitempty"`
	CustomerHistory     map[string]interface{} `json:"customer_history"`
	OperationalRisks    []string               `json:"operational_risks"`
	SpecialInstructions *string                `json:"special_instructions,omitempty"`
}

type SidecarEvaluateRfqPricingRequest struct {
	Context       RfqPricingContextDTO `json:"context"`
	CorrelationID string               `json:"correlation_id"`
}

type SidecarEvaluateRfqPricingResponse struct {
	RfqID                 int64                         `json:"rfq_id"`
	Currency              string                        `json:"currency"`
	BaseCost              float64                       `json:"base_cost"`
	PredictedCost         float64                       `json:"predicted_cost"`
	ActualFacts           []string                      `json:"actual_facts"`
	Predictions           []string                      `json:"predictions"`
	Assumptions           []string                      `json:"assumptions"`
	CandidateStrategies   []PricingStrategyCandidateDTO `json:"candidate_strategies"`
	RecommendedStrategyID string                        `json:"recommended_strategy_id"`
	RecommendedPrice      float64                       `json:"recommended_price"`
	RecommendedMarginPct  float64                       `json:"recommended_margin_pct"`
	MarginRiskLevel       string                        `json:"margin_risk_level"`
	OperationalRiskLevel  string                        `json:"operational_risk_level"`
	ConfidenceScore       float64                       `json:"confidence_score"`
	DataSufficiency       string                        `json:"data_sufficiency"`
	RateFreshnessStatus   string                        `json:"rate_freshness_status"`
	RequiresApproval      bool                          `json:"requires_approval"`
	ApprovalReason        *string                       `json:"approval_reason,omitempty"`
	ReasoningSummary      string                        `json:"reasoning_summary"`
	PlanSteps             []map[string]interface{}      `json:"plan_steps"`
	CorrelationID         string                        `json:"correlation_id"`
}

type SidecarReplanPricingRequest struct {
	Context       RfqPricingContextDTO `json:"context"`
	ReplanReason  string               `json:"replan_reason"`
	RateDelta     float64              `json:"rate_delta"`
	CorrelationID string               `json:"correlation_id"`
}

type RfqPricingEvaluationResult struct {
	Optimization     *RfqPricingOptimization       `json:"optimization"`
	Candidates       []PricingStrategyCandidateDTO `json:"candidates"`
	ActualFacts      []string                      `json:"actual_facts"`
	Predictions      []string                      `json:"predictions"`
	Assumptions      []string                      `json:"assumptions"`
	Plan             *AutonomousPlan               `json:"plan,omitempty"`
	PlanSteps        []AutonomousPlanStep          `json:"plan_steps,omitempty"`
	RequiresApproval bool                          `json:"requires_approval"`
	ApprovalReason   string                        `json:"approval_reason,omitempty"`
}

type RfqPricingStateResponse struct {
	RfqID           int64                         `json:"rfq_id"`
	OrgID           int64                         `json:"org_id"`
	RfqNumber       string                        `json:"rfq_number"`
	CustomerName    string                        `json:"customer_name"`
	Origin          string                        `json:"origin"`
	Destination     string                        `json:"destination"`
	TransportMode   string                        `json:"transport_mode"`
	Status          string                        `json:"status"`
	Optimization    *RfqPricingOptimization       `json:"optimization,omitempty"`
	Candidates      []PricingStrategyCandidateDTO `json:"candidates,omitempty"`
	ActualFacts     []string                      `json:"actual_facts,omitempty"`
	Predictions     []string                      `json:"predictions,omitempty"`
	Assumptions     []string                      `json:"assumptions,omitempty"`
	ActivePlan      *AutonomousPlan               `json:"active_plan,omitempty"`
	ActivePlanSteps []AutonomousPlanStep          `json:"active_plan_steps,omitempty"`
	Versions        []RfqPricingVersion           `json:"versions,omitempty"`
}

type SelectPricingStrategyRequest struct {
	StrategyID string `json:"strategy_id"`
}

type ReplanPricingRequest struct {
	Reason    string  `json:"reason"`
	RateDelta float64 `json:"rate_delta"`
}

type QuotationExecutionResult struct {
	QuotationID        int64   `json:"quotation_id"`
	QuotationNumber    string  `json:"quotation_number"`
	RfqID              int64   `json:"rfq_id"`
	TotalAmount        float64 `json:"total_amount"`
	TotalCost          float64 `json:"total_cost"`
	GrossMarginPct     float64 `json:"gross_margin_pct"`
	Status             string  `json:"status"`
	ActionExecutionID  string  `json:"action_execution_id"`
}

// -----------------------------------------------------------------------------
// Phase 5 Task 5.6: Adaptive Finance and Collections Models
// -----------------------------------------------------------------------------

type FinanceCollectionPlan struct {
	ID                    int64     `db:"id" json:"id"`
	OrgID                 int64     `db:"org_id" json:"org_id"`
	InvoiceID             int64     `db:"invoice_id" json:"invoice_id"`
	CustomerID            int64     `db:"customer_id" json:"customer_id"`
	InvoiceNumber         string    `db:"invoice_number" json:"invoice_number"`
	CustomerName          string    `db:"customer_name" json:"customer_name"`
	Currency              string    `db:"currency" json:"currency"`
	TotalAmount           float64   `db:"total_amount" json:"total_amount"`
	BalanceDue            float64   `db:"balance_due" json:"balance_due"`
	DueDate               *string   `db:"due_date" json:"due_date,omitempty"`
	DaysOverdue           int       `db:"days_overdue" json:"days_overdue"`
	AgingBucket           string    `db:"aging_bucket" json:"aging_bucket"`
	PriorityLevel         string    `db:"priority_level" json:"priority_level"`
	PriorityScore         float64   `db:"priority_score" json:"priority_score"`
	RiskLevel             string    `db:"risk_level" json:"risk_level"`
	RiskScore             float64   `db:"risk_score" json:"risk_score"`
	RecommendedStrategyID string    `db:"recommended_strategy_id" json:"recommended_strategy_id"`
	SelectedStrategyID    string    `db:"selected_strategy_id" json:"selected_strategy_id"`
	CandidateStrategies   string    `db:"candidate_strategies" json:"candidate_strategies"`
	ActualFacts           *string   `db:"actual_facts" json:"actual_facts,omitempty"`
	Predictions           *string   `db:"predictions" json:"predictions,omitempty"`
	Assumptions           *string   `db:"assumptions" json:"assumptions,omitempty"`
	DraftMessage          *string   `db:"draft_message" json:"draft_message,omitempty"`
	PlanSteps             *string   `db:"plan_steps" json:"plan_steps,omitempty"`
	RequiresApproval      bool      `db:"requires_approval" json:"requires_approval"`
	ApprovalReason        *string   `db:"approval_reason" json:"approval_reason,omitempty"`
	AutonomyLevel         string    `db:"autonomy_level" json:"autonomy_level"`
	StopReason            *string   `db:"stop_reason" json:"stop_reason,omitempty"`
	Status                string    `db:"status" json:"status"`
	Version               int       `db:"version" json:"version"`
	IdempotencyKey        string    `db:"idempotency_key" json:"idempotency_key"`
	CorrelationID         string    `db:"correlation_id" json:"correlation_id"`
	CreatedAt             time.Time `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time `db:"updated_at" json:"updated_at"`
}

type FinanceCollectionVersion struct {
	ID            int64     `db:"id" json:"id"`
	PlanID        int64     `db:"plan_id" json:"plan_id"`
	OrgID         int64     `db:"org_id" json:"org_id"`
	VersionNumber int       `db:"version_number" json:"version_number"`
	TriggerEvent  string    `db:"trigger_event" json:"trigger_event"`
	BalanceDue    float64   `db:"balance_due" json:"balance_due"`
	Status        string    `db:"status" json:"status"`
	StrategyID    string    `db:"strategy_id" json:"strategy_id"`
	ChangeReason  *string   `db:"change_reason" json:"change_reason,omitempty"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type CollectionStrategyCandidateDTO struct {
	StrategyID          string  `json:"strategy_id"`
	StrategyName        string  `json:"strategy_name"`
	Description         string  `json:"description"`
	RecommendedAction   string  `json:"recommended_action"`
	PriorityLevel       string  `json:"priority_level"`
	PriorityScore       float64 `json:"priority_score"`
	Urgency             string  `json:"urgency"`
	CooldownDays        int     `json:"cooldown_days"`
	RequiresApproval    bool    `json:"requires_approval"`
	ApprovalReason      *string `json:"approval_reason,omitempty"`
	ExpectedOutcome     string  `json:"expected_outcome"`
	Score               float64 `json:"score"`
	DraftSubject        *string `json:"draft_subject,omitempty"`
	DraftMessage        *string `json:"draft_message,omitempty"`
	IsFeasible          bool    `json:"is_feasible"`
	InfeasibilityReason *string `json:"infeasibility_reason,omitempty"`
}

type FinanceInvoiceContextDTO struct {
	OrgID                   int64                    `json:"org_id"`
	InvoiceID               int64                    `json:"invoice_id"`
	InvoiceNumber           string                   `json:"invoice_number"`
	CustomerID              int64                    `json:"customer_id"`
	CustomerName            string                   `json:"customer_name"`
	AccountTier             string                   `json:"account_tier"`
	Currency                string                   `json:"currency"`
	TotalAmount             float64                  `json:"total_amount"`
	PaidAmount              float64                  `json:"paid_amount"`
	BalanceDue              float64                  `json:"balance_due"`
	DueDate                 *string                  `json:"due_date,omitempty"`
	DaysOverdue             int                      `json:"days_overdue"`
	AgingBucket             string                   `json:"aging_bucket"`
	InvoiceStatus           string                   `json:"invoice_status"`
	IsDisputed              bool                     `json:"is_disputed"`
	DisputeReason           *string                  `json:"dispute_reason,omitempty"`
	CustomerPaymentBehavior map[string]interface{}   `json:"customer_payment_behavior,omitempty"`
	OtherCustomerInvoices   []map[string]interface{} `json:"other_customer_invoices,omitempty"`
	CommunicationHistory    []map[string]interface{} `json:"communication_history,omitempty"`
	SpecialNotes            *string                  `json:"special_notes,omitempty"`
}

type FinanceCollectionEvaluationRequestDTO struct {
	Context       FinanceInvoiceContextDTO `json:"context"`
	CorrelationID string                   `json:"correlation_id"`
}

type FinanceCollectionEvaluationResponseDTO struct {
	InvoiceID             int64                            `json:"invoice_id"`
	InvoiceNumber         string                           `json:"invoice_number"`
	CustomerName          string                           `json:"customer_name"`
	Currency              string                           `json:"currency"`
	TotalAmount           float64                          `json:"total_amount"`
	BalanceDue            float64                          `json:"balance_due"`
	DaysOverdue           int                              `json:"days_overdue"`
	AgingBucket           string                           `json:"aging_bucket"`
	PriorityLevel         string                           `json:"priority_level"`
	PriorityScore         float64                          `json:"priority_score"`
	RiskLevel             string                           `json:"risk_level"`
	RiskScore             float64                          `json:"risk_score"`
	ActualFacts           []string                         `json:"actual_facts"`
	Predictions           []string                         `json:"predictions"`
	Assumptions           []string                         `json:"assumptions"`
	CandidateStrategies   []CollectionStrategyCandidateDTO `json:"candidate_strategies"`
	RecommendedStrategyID string                           `json:"recommended_strategy_id"`
	RecommendedAction     string                           `json:"recommended_action"`
	DraftSubject          string                           `json:"draft_subject"`
	DraftMessage          string                           `json:"draft_message"`
	RequiresApproval      bool                             `json:"requires_approval"`
	ApprovalReason        *string                          `json:"approval_reason,omitempty"`
	StopReason            *string                          `json:"stop_reason,omitempty"`
	ConfidenceScore       float64                          `json:"confidence_score"`
	DataSufficiency       string                           `json:"data_sufficiency"`
	PlanSteps             []map[string]interface{}         `json:"plan_steps"`
	MultiInvoiceSummary   *string                          `json:"multi_invoice_summary,omitempty"`
	CorrelationID         string                           `json:"correlation_id"`
}

type FinanceCollectionReplanningRequestDTO struct {
	Context       FinanceInvoiceContextDTO `json:"context"`
	TriggerEvent  string                   `json:"trigger_event"`
	EventPayload  map[string]interface{}   `json:"event_payload,omitempty"`
	CorrelationID string                   `json:"correlation_id"`
}

type FinanceCollectionStateResponse struct {
	InvoiceID           int64                            `json:"invoice_id"`
	OrgID               int64                            `json:"org_id"`
	InvoiceNumber       string                           `json:"invoice_number"`
	CustomerName        string                           `json:"customer_name"`
	Currency            string                           `json:"currency"`
	TotalAmount         float64                          `json:"total_amount"`
	BalanceDue          float64                          `json:"balance_due"`
	DueDate             *string                          `json:"due_date,omitempty"`
	DaysOverdue         int                              `json:"days_overdue"`
	AgingBucket         string                           `json:"aging_bucket"`
	InvoiceStatus       string                           `json:"invoice_status"`
	Plan                *FinanceCollectionPlan           `json:"plan,omitempty"`
	Candidates          []CollectionStrategyCandidateDTO `json:"candidates,omitempty"`
	ActualFacts         []string                         `json:"actual_facts,omitempty"`
	Predictions         []string                         `json:"predictions,omitempty"`
	Assumptions         []string                         `json:"assumptions,omitempty"`
	PlanSteps           []map[string]interface{}         `json:"plan_steps,omitempty"`
	Versions            []FinanceCollectionVersion       `json:"versions,omitempty"`
	MultiInvoiceSummary *string                          `json:"multi_invoice_summary,omitempty"`
}

type SelectCollectionStrategyRequest struct {
	StrategyID string `json:"strategy_id"`
}

type ReplanCollectionRequest struct {
	TriggerEvent string                 `json:"trigger_event"`
	Payload      map[string]interface{} `json:"payload,omitempty"`
}

// -----------------------------------------------------------------------------
// Phase 5 Task 5.7: Contract and Compliance Monitoring Models
// -----------------------------------------------------------------------------

type ComplianceRemediationStrategyCandidateDTO struct {
	StrategyID          string  `json:"strategy_id"`
	StrategyName        string  `json:"strategy_name"`
	StrategyType        string  `json:"strategy_type"`
	Description         string  `json:"description"`
	RecommendedAction   string  `json:"recommended_action"`
	SeverityLevel       string  `json:"severity_level"`
	Urgency             string  `json:"urgency"`
	RequiresApproval    bool    `json:"requires_approval"`
	ApprovalReason      *string `json:"approval_reason,omitempty"`
	ExpectedOutcome     string  `json:"expected_outcome"`
	Score               float64 `json:"score"`
	DraftSubject        *string `json:"draft_subject,omitempty"`
	DraftMessage        *string `json:"draft_message,omitempty"`
	IsFeasible          bool    `json:"is_feasible"`
	InfeasibilityReason *string `json:"infeasibility_reason,omitempty"`
}

type ContractComplianceContextDTO struct {
	OrgID                  int64                    `json:"org_id"`
	ContractID             int64                    `json:"contract_id"`
	ContractReference      string                   `json:"contract_reference"`
	ContractName           string                   `json:"contract_name"`
	ContractType           string                   `json:"contract_type"`
	PartyID                *int64                   `json:"party_id,omitempty"`
	PartyName              string                   `json:"party_name"`
	TransportMode          string                   `json:"transport_mode"`
	Status                 string                   `json:"status"`
	Currency               string                   `json:"currency"`
	ContractValue          float64                  `json:"contract_value"`
	EffectiveDate          *string                  `json:"effective_date,omitempty"`
	ExpiryDate             *string                  `json:"expiry_date,omitempty"`
	DaysUntilExpiration    int                      `json:"days_until_expiration"`
	Terms                  []map[string]interface{} `json:"terms,omitempty"`
	ComplianceRequirements []map[string]interface{} `json:"compliance_requirements,omitempty"`
	Documents              []map[string]interface{} `json:"documents,omitempty"`
	ActiveShipmentsCount   int                      `json:"active_shipments_count"`
	RateDeviations         []map[string]interface{} `json:"rate_deviations,omitempty"`
	OperationalNotes       *string                  `json:"operational_notes,omitempty"`
}

type ContractComplianceEvaluationRequestDTO struct {
	Context       ContractComplianceContextDTO `json:"context"`
	CorrelationID string                       `json:"correlation_id"`
}

type ContractComplianceEvaluationResponseDTO struct {
	ContractID            int64                                       `json:"contract_id"`
	ContractReference     string                                      `json:"contract_reference"`
	ContractName          string                                      `json:"contract_name"`
	PartyName             string                                      `json:"party_name"`
	ContractType          string                                      `json:"contract_type"`
	Status                string                                      `json:"status"`
	EffectiveDate         *string                                     `json:"effective_date,omitempty"`
	ExpiryDate            *string                                     `json:"expiry_date,omitempty"`
	DaysUntilExpiration   int                                         `json:"days_until_expiration"`
	ExpirationStatus      string                                      `json:"expiration_status"`
	ComplianceStatus      string                                      `json:"compliance_status"`
	HardRequirementCount  int                                         `json:"hard_requirement_count"`
	SoftRequirementCount  int                                         `json:"soft_requirement_count"`
	HardViolationsCount   int                                         `json:"hard_violations_count"`
	SoftDeviationsCount   int                                         `json:"soft_deviations_count"`
	MissingDocumentsCount int                                         `json:"missing_documents_count"`
	ExpiredDocumentsCount int                                         `json:"expired_documents_count"`
	RiskLevel             string                                      `json:"risk_level"`
	RiskScore             float64                                     `json:"risk_score"`
	AuthoritativeFacts    []string                                    `json:"authoritative_facts"`
	ExtractedTerms        []string                                    `json:"extracted_terms"`
	Predictions           []string                                    `json:"predictions"`
	Assumptions           []string                                    `json:"assumptions"`
	Deviations            []map[string]interface{}                    `json:"deviations"`
	CandidateStrategies   []ComplianceRemediationStrategyCandidateDTO `json:"candidate_strategies"`
	RecommendedStrategyID string                                      `json:"recommended_strategy_id"`
	RecommendedAction     string                                      `json:"recommended_action"`
	DraftSubject          string                                      `json:"draft_subject"`
	DraftMessage          string                                      `json:"draft_message"`
	RequiresApproval      bool                                        `json:"requires_approval"`
	ApprovalReason        *string                                     `json:"approval_reason,omitempty"`
	StopReason            *string                                     `json:"stop_reason,omitempty"`
	ConfidenceScore       float64                                     `json:"confidence_score"`
	DataSufficiency       string                                      `json:"data_sufficiency"`
	RemediationPlanSteps  []map[string]interface{}                    `json:"remediation_plan_steps"`
	CorrelationID         string                                      `json:"correlation_id"`
}

type ContractComplianceReplanningRequestDTO struct {
	Context       ContractComplianceContextDTO `json:"context"`
	TriggerEvent  string                       `json:"trigger_event"`
	EventPayload  map[string]interface{}       `json:"event_payload,omitempty"`
	CorrelationID string                       `json:"correlation_id"`
}

type ContractComplianceMonitoringPlan struct {
	ID                                int64     `json:"id"`
	OrgID                             int64     `json:"org_id"`
	ContractID                        int64     `json:"contract_id"`
	ContractReference                 string    `json:"contract_reference"`
	ContractName                      string    `json:"contract_name"`
	PartyName                         string    `json:"party_name"`
	ContractType                      string    `json:"contract_type"`
	Status                            string    `json:"status"`
	EffectiveDate                     *string   `json:"effective_date,omitempty"`
	ExpiryDate                        *string   `json:"expiry_date,omitempty"`
	DaysUntilExpiration               int       `json:"days_until_expiration"`
	ExpirationStatus                  string    `json:"expiration_status"`
	ComplianceStatus                  string    `json:"compliance_status"`
	HardRequirementCount             int       `json:"hard_requirement_count"`
	SoftRequirementCount             int       `json:"soft_requirement_count"`
	HardViolationsCount              int       `json:"hard_violations_count"`
	SoftDeviationsCount              int       `json:"soft_deviations_count"`
	MissingDocumentsCount            int       `json:"missing_documents_count"`
	ExpiredDocumentsCount            int       `json:"expired_documents_count"`
	RiskLevel                         string    `json:"risk_level"`
	RiskScore                         float64   `json:"risk_score"`
	RecommendedRemediationStrategyID string    `json:"recommended_remediation_strategy_id"`
	SelectedRemediationStrategyID    string    `json:"selected_remediation_strategy_id"`
	CandidateStrategies               *string   `json:"candidate_strategies,omitempty"`
	AuthoritativeFacts                *string   `json:"authoritative_facts,omitempty"`
	ExtractedTerms                    *string   `json:"extracted_terms,omitempty"`
	Predictions                       *string   `json:"predictions,omitempty"`
	Assumptions                       *string   `json:"assumptions,omitempty"`
	RemediationPlanSteps              *string   `json:"remediation_plan_steps,omitempty"`
	Deviations                        *string   `json:"deviations,omitempty"`
	RequiresApproval                  bool      `json:"requires_approval"`
	ApprovalReason                    *string   `json:"approval_reason,omitempty"`
	AutonomyLevel                     string    `json:"autonomy_level"`
	StopReason                        *string   `json:"stop_reason,omitempty"`
	ExecutionStatus                   string    `json:"execution_status"`
	Version                           int       `json:"version"`
	IdempotencyKey                    string    `json:"idempotency_key"`
	CorrelationID                     string    `json:"correlation_id"`
	CreatedAt                         time.Time `json:"created_at"`
	UpdatedAt                         time.Time `json:"updated_at"`
}

type ContractComplianceMonitoringVersion struct {
	ID               int64     `json:"id"`
	PlanID           int64     `json:"plan_id"`
	OrgID            int64     `json:"org_id"`
	VersionNumber    int       `json:"version_number"`
	TriggerEvent     string    `json:"trigger_event"`
	ComplianceStatus string    `json:"compliance_status"`
	StrategyID       string    `json:"strategy_id"`
	ChangeReason     *string   `json:"change_reason,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type ContractComplianceStateResponse struct {
	ContractID           int64                                       `json:"contract_id"`
	OrgID                int64                                       `json:"org_id"`
	ContractReference    string                                      `json:"contract_reference"`
	ContractName         string                                      `json:"contract_name"`
	PartyName            string                                      `json:"party_name"`
	ContractType         string                                      `json:"contract_type"`
	Status               string                                      `json:"status"`
	EffectiveDate        *string                                     `json:"effective_date,omitempty"`
	ExpiryDate           *string                                     `json:"expiry_date,omitempty"`
	DaysUntilExpiration  int                                         `json:"days_until_expiration"`
	ExpirationStatus     string                                      `json:"expiration_status"`
	ComplianceStatus     string                                      `json:"compliance_status"`
	Plan                 *ContractComplianceMonitoringPlan           `json:"plan,omitempty"`
	Candidates           []ComplianceRemediationStrategyCandidateDTO `json:"candidates,omitempty"`
	AuthoritativeFacts   []string                                    `json:"authoritative_facts,omitempty"`
	ExtractedTerms       []string                                    `json:"extracted_terms,omitempty"`
	Predictions          []string                                    `json:"predictions,omitempty"`
	Assumptions          []string                                    `json:"assumptions,omitempty"`
	Deviations           []map[string]interface{}                    `json:"deviations,omitempty"`
	RemediationPlanSteps []map[string]interface{}                    `json:"remediation_plan_steps,omitempty"`
	Versions             []ContractComplianceMonitoringVersion       `json:"versions,omitempty"`
}

type SelectComplianceStrategyRequest struct {
	StrategyID string `json:"strategy_id"`
}

type ReplanComplianceRequest struct {
	TriggerEvent string                 `json:"trigger_event"`
	Payload      map[string]interface{} `json:"payload,omitempty"`
}

// -------------------------------------------------------------------------
// Phase 5 Task 5.8: Autonomous Exception Resolution Types
// -------------------------------------------------------------------------

type ExceptionCandidateRecoveryStrategyDTO struct {
	StrategyID             string  `json:"strategy_id"`
	StrategyName           string  `json:"strategy_name"`
	StrategyType           string  `json:"strategy_type"`
	Description            string  `json:"description"`
	RecommendedAction      string  `json:"recommended_action"`
	ExpectedResolutionProb float64 `json:"expected_resolution_prob"`
	TimeToResolution       string  `json:"time_to_resolution"`
	CostImpact             float64 `json:"cost_impact"`
	MarginImpact           string  `json:"margin_impact"`
	Reversibility          string  `json:"reversibility"`
	ExecutionComplexity    string  `json:"execution_complexity"`
	RequiresApproval       bool    `json:"requires_approval"`
	ApprovalReason         *string `json:"approval_reason,omitempty"`
	ExpectedOutcome        string  `json:"expected_outcome"`
	Score                  float64 `json:"score"`
	IsFeasible             bool    `json:"is_feasible"`
	InfeasibilityReason    *string `json:"infeasibility_reason,omitempty"`
}

type ExceptionResolutionPlan struct {
	ID                   int64     `db:"id" json:"id"`
	OrgID                int64     `db:"org_id" json:"org_id"`
	ExceptionID          int64     `db:"exception_id" json:"exception_id"`
	ShipmentID           int64     `db:"shipment_id" json:"shipment_id"`
	ExceptionType        string    `db:"exception_type" json:"exception_type"`
	Severity             string    `db:"severity" json:"severity"`
	LifecycleStatus      string    `db:"lifecycle_status" json:"lifecycle_status"`
	WaitingState         *string   `db:"waiting_state" json:"waiting_state,omitempty"`
	LikelyRootCause      *string   `db:"likely_root_cause" json:"likely_root_cause,omitempty"`
	Symptom              *string   `db:"symptom" json:"symptom,omitempty"`
	ContributingFactors  *string   `db:"contributing_factors" json:"contributing_factors,omitempty"`
	Evidence             *string   `db:"evidence" json:"evidence,omitempty"`
	Confidence           float64   `db:"confidence" json:"confidence"`
	ImpactAssessment     *string   `db:"impact_assessment" json:"impact_assessment,omitempty"`
	Constraints          *string   `db:"constraints" json:"constraints,omitempty"`
	CandidateStrategies  *string   `db:"candidate_strategies" json:"candidate_strategies,omitempty"`
	SelectedStrategyID   string    `db:"selected_strategy_id" json:"selected_strategy_id"`
	RecoveryPlanSteps    *string   `db:"recovery_plan_steps" json:"recovery_plan_steps,omitempty"`
	VerificationCriteria *string   `db:"verification_criteria" json:"verification_criteria,omitempty"`
	RequiresApproval     bool      `db:"requires_approval" json:"requires_approval"`
	ApprovalReason       *string   `db:"approval_reason" json:"approval_reason,omitempty"`
	AutonomyLevel        string    `db:"autonomy_level" json:"autonomy_level"`
	StopReason           *string   `db:"stop_reason" json:"stop_reason,omitempty"`
	EscalationReason     *string   `db:"escalation_reason" json:"escalation_reason,omitempty"`
	ResolutionNotes      *string   `db:"resolution_notes" json:"resolution_notes,omitempty"`
	Version              int       `db:"version" json:"version"`
	IdempotencyKey       string    `db:"idempotency_key" json:"idempotency_key"`
	CorrelationID        string    `db:"correlation_id" json:"correlation_id"`
	CreatedAt            time.Time `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time `db:"updated_at" json:"updated_at"`
}

type ExceptionResolutionVersion struct {
	ID                 int64     `db:"id" json:"id"`
	PlanID             int64     `db:"plan_id" json:"plan_id"`
	OrgID              int64     `db:"org_id" json:"org_id"`
	ExceptionID        int64     `db:"exception_id" json:"exception_id"`
	VersionNumber      int       `db:"version_number" json:"version_number"`
	TriggerEvent       string    `db:"trigger_event" json:"trigger_event"`
	LifecycleStatus    string    `db:"lifecycle_status" json:"lifecycle_status"`
	SelectedStrategyID string    `db:"selected_strategy_id" json:"selected_strategy_id"`
	ChangeReason       *string   `db:"change_reason" json:"change_reason,omitempty"`
	Snapshot           *string   `db:"snapshot" json:"snapshot,omitempty"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
}

type ExceptionResolutionStateResponse struct {
	ExceptionID          int64                                   `json:"exception_id"`
	ShipmentID           int64                                   `json:"shipment_id"`
	OrgID                int64                                   `json:"org_id"`
	ExceptionType        string                                  `json:"exception_type"`
	Severity             string                                  `json:"severity"`
	Title                string                                  `json:"title"`
	Description          *string                                 `json:"description,omitempty"`
	Status               string                                  `json:"status"`
	Plan                 *ExceptionResolutionPlan                `json:"plan,omitempty"`
	Candidates           []ExceptionCandidateRecoveryStrategyDTO `json:"candidates,omitempty"`
	Symptom              string                                  `json:"symptom"`
	LikelyRootCause      string                                  `json:"likely_root_cause"`
	ContributingFactors  []string                                `json:"contributing_factors,omitempty"`
	Evidence             []string                                `json:"evidence,omitempty"`
	UnknownFactors       []string                                `json:"unknown_factors,omitempty"`
	ImpactAssessment     map[string]interface{}                  `json:"impact_assessment,omitempty"`
	HardConstraints      []string                                `json:"hard_constraints,omitempty"`
	RecoveryPlanSteps    []map[string]interface{}                `json:"recovery_plan_steps,omitempty"`
	VerificationCriteria map[string]interface{}                  `json:"verification_criteria,omitempty"`
	Versions             []ExceptionResolutionVersion            `json:"versions,omitempty"`
}

type SelectExceptionStrategyRequest struct {
	StrategyID string `json:"strategy_id"`
}

type ReplanExceptionRequest struct {
	TriggerEvent string                 `json:"trigger_event"`
	Payload      map[string]interface{} `json:"payload,omitempty"`
}

type ExecuteExceptionActionRequest struct {
	ActionType     string                 `json:"action_type"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
	IdempotencyKey string                 `json:"idempotency_key"`
}

// ==============================================================================
// Phase 5 Task 5.9: Multi-Step AI Planning & Execution DTOs
// ==============================================================================

type ValidatePlanRequest struct {
	PlanID string `json:"plan_id"`
}

type ValidatePlanResponse struct {
	PlanID          string     `json:"plan_id"`
	IsValid         bool       `json:"is_valid"`
	Issues          []string   `json:"issues"`
	Warnings        []string   `json:"warnings"`
	ExecutionOrder  []string   `json:"execution_order"`
	ParallelGroups  [][]string `json:"parallel_groups"`
	ContainsCycles  bool       `json:"contains_cycles"`
	HasApprovalGate bool       `json:"has_approval_gate"`
	StalenessCheck  string     `json:"staleness_check"`
}

type ApproveStepRequest struct {
	Notes string `json:"notes,omitempty"`
}

type StepApprovalResponse struct {
	PlanID     string     `json:"plan_id"`
	StepID     string     `json:"step_id"`
	StepStatus StepStatus `json:"step_status"`
	ApproverID int64      `json:"approver_id"`
	ApprovedAt time.Time  `json:"approved_at"`
	Notes      string     `json:"notes,omitempty"`
}

type RetryStepRequest struct {
	MaxAttempts *int   `json:"max_attempts,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

type CompensateStepRequest struct {
	Reason string `json:"reason,omitempty"`
}

type ConcurrentPlanConflict struct {
	PlanID        string     `json:"plan_id"`
	Goal          string     `json:"goal"`
	Module        string     `json:"module"`
	Priority      string     `json:"priority"`
	Status        PlanStatus `json:"status"`
	CurrentStepID *string    `json:"current_step_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type ConcurrentPlanConflictResponse struct {
	HasConflict    bool                     `json:"has_conflict"`
	EntityType     string                   `json:"entity_type"`
	EntityID       string                   `json:"entity_id"`
	ActivePlans    []ConcurrentPlanConflict `json:"active_plans"`
	PriorityAction string                   `json:"priority_action"` // E.g., "EXECUTION_ALLOWED", "BLOCKED_BY_HIGHER_PRIORITY", "REVIEW_MANDATED"
	Reason         string                   `json:"reason"`
}

type PlanningMetricsResponse struct {
	OrgID                   int64   `json:"org_id"`
	TotalPlans              int64   `json:"total_plans"`
	CompletedPlans          int64   `json:"completed_plans"`
	ActivePlans             int64   `json:"active_plans"`
	FailedPlans             int64   `json:"failed_plans"`
	ReplannedPlans          int64   `json:"replanned_plans"`
	PlanSuccessRate         float64 `json:"plan_success_rate"`
	TotalSteps              int64   `json:"total_steps"`
	CompletedSteps          int64   `json:"completed_steps"`
	FailedSteps             int64   `json:"failed_steps"`
	StepFailureRate         float64 `json:"step_failure_rate"`
	AverageStepsPerPlan     float64 `json:"average_steps_per_plan"`
	ApprovalRequiredCount   int64   `json:"approval_required_count"`
	AutonomousExecutionRate float64 `json:"autonomous_execution_rate"`
}

type CrossModulePlanRequest struct {
	Goal             string                 `json:"goal"`
	PrimaryModule    string                 `json:"primary_module"`
	PrimaryEntityID  string                 `json:"primary_entity_id"`
	InvolvedModules  []string               `json:"involved_modules"`
	Context          map[string]interface{} `json:"context,omitempty"`
	HardConstraints []ConstraintDTO        `json:"hard_constraints,omitempty"`
	SoftConstraints []ConstraintDTO        `json:"soft_constraints,omitempty"`
	AutonomyLevel    AutonomyLevel          `json:"autonomy_level,omitempty"`
	Priority         string                 `json:"priority,omitempty"`
	CorrelationID    string                 `json:"correlation_id,omitempty"`
}

type SidecarPlanValidationRequest struct {
	PlanID        string                 `json:"plan_id"`
	Steps         []AutonomousPlanStep   `json:"steps"`
	Module        string                 `json:"module"`
	AutonomyLevel AutonomyLevel          `json:"autonomy_level"`
}

type SidecarPlanValidationResponse struct {
	IsValid         bool       `json:"is_valid"`
	Issues          []string   `json:"issues"`
	Warnings        []string   `json:"warnings"`
	ExecutionOrder  []string   `json:"execution_order"`
	ParallelGroups  [][]string `json:"parallel_groups"`
	ContainsCycles  bool       `json:"contains_cycles"`
	HasApprovalGate bool       `json:"has_approval_gate"`
}

type SidecarCrossModulePlanRequest struct {
	OrgID            int64                  `json:"org_id"`
	Goal             string                 `json:"goal"`
	PrimaryModule    string                 `json:"primary_module"`
	PrimaryEntityID  string                 `json:"primary_entity_id"`
	InvolvedModules  []string               `json:"involved_modules"`
	Context          map[string]interface{} `json:"context"`
	HardConstraints []ConstraintDTO        `json:"hard_constraints"`
	SoftConstraints []ConstraintDTO        `json:"soft_constraints"`
	AutonomyLevel    AutonomyLevel          `json:"autonomy_level"`
	CorrelationID    string                 `json:"correlation_id,omitempty"`
}

type SidecarCrossModulePlanResponse struct {
	PlanID                string                   `json:"plan_id"`
	Goal                  string                   `json:"goal"`
	GoalType              string                   `json:"goal_type"`
	PrimaryModule         string                   `json:"primary_module"`
	PrimaryEntityID       string                   `json:"primary_entity_id"`
	InvolvedModules       []string                 `json:"involved_modules"`
	OrderedSteps          []map[string]interface{} `json:"ordered_steps"`
	ParallelGroups        [][]string               `json:"parallel_groups"`
	StopConditions        []map[string]interface{} `json:"stop_conditions"`
	ConfidenceScore       float64                  `json:"confidence_score"`
	DataSufficiency       bool                     `json:"data_sufficiency"`
	RequiresHumanApproval bool                     `json:"requires_human_approval"`
	CorrelationID         string                   `json:"correlation_id"`
}

// ==============================================================================
// Phase 5 Task 5.10: Continuous Monitoring and Replanning Models
// ==============================================================================

type PlanHealthState string

const (
	PlanHealthHealthy    PlanHealthState = "HEALTHY"
	PlanHealthAtRisk     PlanHealthState = "AT_RISK"
	PlanHealthStale      PlanHealthState = "STALE"
	PlanHealthBlocked    PlanHealthState = "BLOCKED"
	PlanHealthFailed     PlanHealthState = "FAILED"
	PlanHealthCompleted  PlanHealthState = "COMPLETED"
	PlanHealthReplanning PlanHealthState = "REPLANNING"
	PlanHealthEscalated  PlanHealthState = "ESCALATED"
	PlanHealthWaiting    PlanHealthState = "WAITING"
)

type RecommendedAction string

const (
	ActionContinue RecommendedAction = "CONTINUE"
	ActionPause    RecommendedAction = "PAUSE"
	ActionReplan   RecommendedAction = "REPLAN"
	ActionEscalate RecommendedAction = "ESCALATE"
	ActionStop     RecommendedAction = "STOP"
)

type AIMonitoringEvent struct {
	ID              int64           `db:"id" json:"id"`
	OrgID           int64           `db:"org_id" json:"org_id"`
	EventID         string          `db:"event_id" json:"event_id"`
	CorrelationID   string          `db:"correlation_id" json:"correlation_id"`
	EventType       string          `db:"event_type" json:"event_type"`
	EntityType      string          `db:"entity_type" json:"entity_type"`
	EntityID        string          `db:"entity_id" json:"entity_id"`
	Source          string          `db:"source" json:"source"`
	EventPayload    json.RawMessage `db:"event_payload" json:"event_payload"`
	IsMaterial      bool            `db:"is_material" json:"is_material"`
	FilterReason    string          `db:"filter_reason" json:"filter_reason"`
	AIEvaluated     bool            `db:"ai_evaluated" json:"ai_evaluated"`
	ReplanTriggered bool            `db:"replan_triggered" json:"replan_triggered"`
	PlanID          sql.NullString  `db:"plan_id" json:"plan_id,omitempty"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
}

type IngestMonitoringEventRequest struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"`
	EntityType    string                 `json:"entity_type"`
	EntityID      string                 `json:"entity_id"`
	Source        string                 `json:"source"`
	Payload       map[string]interface{} `json:"payload"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	CausationID   string                 `json:"causation_id,omitempty"`
	Timestamp     string                 `json:"timestamp,omitempty"`
}

type IngestMonitoringEventResponse struct {
	EventID           string            `json:"event_id"`
	IsDuplicate       bool              `json:"is_duplicate"`
	IsMaterial        bool              `json:"is_material"`
	FilterReason      string            `json:"filter_reason"`
	PlanID            string            `json:"plan_id,omitempty"`
	PlanHealth        PlanHealthState   `json:"plan_health,omitempty"`
	RecommendedAction RecommendedAction `json:"recommended_action,omitempty"`
	ReplanTriggered   bool              `json:"replan_triggered"`
	NewPlanID         string            `json:"new_plan_id,omitempty"`
	NewVersion        int               `json:"new_version,omitempty"`
	CorrelationID     string            `json:"correlation_id"`
}

type SidecarStateChangeEvaluationRequest struct {
	OrgID               int64                  `json:"org_id"`
	PlanID              string                 `json:"plan_id"`
	CurrentPlan         map[string]interface{} `json:"current_plan"`
	Steps               []AutonomousPlanStep   `json:"steps"`
	Event               map[string]interface{} `json:"event"`
	AuthoritativeState  map[string]interface{} `json:"authoritative_state"`
	PreviousPredictions map[string]interface{} `json:"previous_predictions,omitempty"`
	NewPredictions      map[string]interface{} `json:"new_predictions,omitempty"`
	AutonomyLevel       AutonomyLevel          `json:"autonomy_level"`
	CorrelationID       string                 `json:"correlation_id,omitempty"`
}

type SidecarStateChangeEvaluationResponse struct {
	PlanID               string            `json:"plan_id"`
	PlanHealth           PlanHealthState   `json:"plan_health"`
	IsPlanValid          bool              `json:"is_plan_valid"`
	MaterialityAnalysis  string            `json:"materiality_analysis"`
	InvalidatedStepIDs   []string          `json:"invalidated_step_ids"`
	ChangedAssumptions   []string          `json:"changed_assumptions"`
	RecommendedAction    RecommendedAction `json:"recommended_action"`
	ReplanRationale      string            `json:"replan_rationale,omitempty"`
	EscalationDetails    string            `json:"escalation_details,omitempty"`
	ConfidenceScore      float64           `json:"confidence_score"`
	CorrelationID        string            `json:"correlation_id"`
}

type SidecarContinuousReplanningRequest struct {
	OrgID              int64                  `json:"org_id"`
	PlanID             string                 `json:"plan_id"`
	CurrentPlan        map[string]interface{} `json:"current_plan"`
	CompletedSteps     []AutonomousPlanStep   `json:"completed_steps"`
	InvalidatedStepIDs []string               `json:"invalidated_step_ids"`
	Event              map[string]interface{} `json:"event"`
	AuthoritativeState map[string]interface{} `json:"authoritative_state"`
	AutonomyLevel      AutonomyLevel          `json:"autonomy_level"`
	CorrelationID      string                 `json:"correlation_id,omitempty"`
}

type SidecarContinuousReplanningResponse struct {
	NewPlanID                string                   `json:"new_plan_id"`
	Version                  int                      `json:"version"`
	ParentPlanID             string                   `json:"parent_plan_id"`
	OrderedSteps             []map[string]interface{} `json:"ordered_steps"`
	ParallelGroups           [][]string               `json:"parallel_groups"`
	ProtectedCompletedSteps  []string                 `json:"protected_completed_steps"`
	ChangedAssumptions       []string                 `json:"changed_assumptions"`
	ReplanRationale          string                   `json:"replan_rationale"`
	ConfidenceScore          float64                  `json:"confidence_score"`
	RequiresApproval         bool                     `json:"requires_approval"`
	CorrelationID            string                   `json:"correlation_id"`
}

type ContinuousMonitoringMetrics struct {
	EventsProcessed             int64   `json:"events_processed"`
	EventsFiltered              int64   `json:"events_filtered"`
	AIEvaluatedCount            int64   `json:"ai_evaluated_count"`
	MaterialChangeRate          float64 `json:"material_change_rate"`
	ReplanningCount             int64   `json:"replanning_count"`
	StalePlanCount              int64   `json:"stale_plan_count"`
	EscalatedCount              int64   `json:"escalated_count"`
	ProtectedCompletedStepsCount int64   `json:"protected_completed_steps_count"`
	AverageReplansPerPlan       float64 `json:"average_replans_per_plan"`
}

// -----------------------------------------------------------------------------
// Phase 5 Task 5.11: Human + AI Operating Model
// -----------------------------------------------------------------------------

// OperatingMode defines the 9 operational modes of human-AI collaboration
type OperatingMode string

const (
	OpModeAIObserve     OperatingMode = "AI_OBSERVE"
	OpModeAIRecommend   OperatingMode = "AI_RECOMMEND"
	OpModeAIPrepare     OperatingMode = "AI_PREPARE"
	OpModeAIExecute     OperatingMode = "AI_EXECUTE"
	OpModeHumanReview   OperatingMode = "HUMAN_REVIEW"
	OpModeHumanApproval OperatingMode = "HUMAN_APPROVAL"
	OpModeHumanOverride OperatingMode = "HUMAN_OVERRIDE"
	OpModeAIVerify      OperatingMode = "AI_VERIFY"
	OpModeAIEscalate    OperatingMode = "AI_ESCALATE"
)

type HumanAIDecisionStatus string

const (
	DecisionStatusPending            HumanAIDecisionStatus = "PENDING"
	DecisionStatusApproved           HumanAIDecisionStatus = "APPROVED"
	DecisionStatusRejected           HumanAIDecisionStatus = "REJECTED"
	DecisionStatusExpired            HumanAIDecisionStatus = "EXPIRED"
	DecisionStatusCancelled          HumanAIDecisionStatus = "CANCELLED"
	DecisionStatusSuperseded         HumanAIDecisionStatus = "SUPERSEDED"
	DecisionStatusRequiresReapproval HumanAIDecisionStatus = "REQUIRES_REAPPROVAL"
	DecisionStatusStopped            HumanAIDecisionStatus = "STOPPED"
	DecisionStatusEscalated          HumanAIDecisionStatus = "ESCALATED"
)

type HumanAIDecision struct {
	ID                 int64                 `db:"id" json:"id"`
	OrgID              int64                 `db:"org_id" json:"org_id"`
	DecisionID         string                `db:"decision_id" json:"decision_id"`
	CorrelationID      string                `db:"correlation_id" json:"correlation_id"`
	PlanID             sql.NullString        `db:"plan_id" json:"plan_id"`
	StepID             sql.NullString        `db:"step_id" json:"step_id"`
	ApprovalID         sql.NullInt64         `db:"approval_id" json:"approval_id"`
	Module             string                `db:"module" json:"module"`
	EntityType         string                `db:"entity_type" json:"entity_type"`
	EntityID           string                `db:"entity_id" json:"entity_id"`
	OperatingMode      OperatingMode         `db:"operating_mode" json:"operating_mode"`
	AutonomyLevel      AutonomyLevel         `db:"autonomy_level" json:"autonomy_level"`
	Title              string                `db:"title" json:"title"`
	ContextSummary     string                `db:"context_summary" json:"context_summary"`
	Facts              json.RawMessage       `db:"facts" json:"facts"`
	Predictions        json.RawMessage       `db:"predictions" json:"predictions"`
	AIRecommendation   string                `db:"ai_recommendation" json:"ai_recommendation"`
	OriginalAIPayload  json.RawMessage       `db:"original_ai_payload" json:"original_ai_payload"`
	HumanEditedPayload json.RawMessage       `db:"human_edited_payload" json:"human_edited_payload"`
	Alternatives       json.RawMessage       `db:"alternatives" json:"alternatives"`
	Confidence         string                `db:"confidence" json:"confidence"`
	DataSufficiency    string                `db:"data_sufficiency" json:"data_sufficiency"`
	RiskLevel          string                `db:"risk_level" json:"risk_level"`
	IsReversible       bool                  `db:"is_reversible" json:"is_reversible"`
	DecisionStatus     HumanAIDecisionStatus `db:"decision_status" json:"decision_status"`
	HumanDecision      sql.NullString        `db:"human_decision" json:"human_decision"`
	DecisionReason     sql.NullString        `db:"decision_reason" json:"decision_reason"`
	DecidedByID        sql.NullInt64         `db:"decided_by_id" json:"decided_by_id"`
	DecidedByName      sql.NullString        `db:"decided_by_name" json:"decided_by_name"`
	DecidedAt          sql.NullTime          `db:"decided_at" json:"decided_at"`
	FeedbackType       sql.NullString        `db:"feedback_type" json:"feedback_type"`
	CreatedAt          time.Time             `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time             `db:"updated_at" json:"updated_at"`
}

func (d HumanAIDecision) MarshalJSON() ([]byte, error) {
	type Alias HumanAIDecision
	var humanDec, decReason, decidedByName, feedbackType, planID, stepID *string
	var approvalID, decidedByID *int64
	var decidedAt *time.Time

	if d.HumanDecision.Valid {
		humanDec = &d.HumanDecision.String
	}
	if d.DecisionReason.Valid {
		decReason = &d.DecisionReason.String
	}
	if d.DecidedByName.Valid {
		decidedByName = &d.DecidedByName.String
	}
	if d.FeedbackType.Valid {
		feedbackType = &d.FeedbackType.String
	}
	if d.PlanID.Valid {
		planID = &d.PlanID.String
	}
	if d.StepID.Valid {
		stepID = &d.StepID.String
	}
	if d.ApprovalID.Valid {
		approvalID = &d.ApprovalID.Int64
	}
	if d.DecidedByID.Valid {
		decidedByID = &d.DecidedByID.Int64
	}
	if d.DecidedAt.Valid {
		decidedAt = &d.DecidedAt.Time
	}

	return json.Marshal(&struct {
		Alias
		PlanID         *string    `json:"plan_id,omitempty"`
		StepID         *string    `json:"step_id,omitempty"`
		ApprovalID     *int64     `json:"approval_id,omitempty"`
		HumanDecision  *string    `json:"human_decision,omitempty"`
		DecisionReason *string    `json:"decision_reason,omitempty"`
		DecidedByID    *int64     `json:"decided_by_id,omitempty"`
		DecidedByName  *string    `json:"decided_by_name,omitempty"`
		DecidedAt      *time.Time `json:"decided_at,omitempty"`
		FeedbackType   *string    `json:"feedback_type,omitempty"`
	}{
		Alias:          Alias(d),
		PlanID:         planID,
		StepID:         stepID,
		ApprovalID:     approvalID,
		HumanDecision:  humanDec,
		DecisionReason: decReason,
		DecidedByID:    decidedByID,
		DecidedByName:  decidedByName,
		DecidedAt:      decidedAt,
		FeedbackType:   feedbackType,
	})
}

type CreateDecisionPointRequest struct {
	PlanID             string                 `json:"plan_id,omitempty"`
	StepID             string                 `json:"step_id,omitempty"`
	ApprovalID         *int64                 `json:"approval_id,omitempty"`
	Module             string                 `json:"module"`
	EntityType         string                 `json:"entity_type"`
	EntityID           string                 `json:"entity_id"`
	Title              string                 `json:"title,omitempty"`
	Facts              []string               `json:"facts"`
	Predictions        []string               `json:"predictions"`
	AuthoritativeState map[string]interface{} `json:"authoritative_state,omitempty"`
	AutonomyLevel      AutonomyLevel          `json:"autonomy_level,omitempty"`
	CorrelationID      string                 `json:"correlation_id,omitempty"`
}

type SubmitHumanDecisionRequest struct {
	DecisionType       string          `json:"decision_type"` // APPROVE, REJECT, EDIT_AND_APPROVE, OVERRIDE, STOP, ESCALATE
	Reason             string          `json:"reason,omitempty"`
	HumanEditedPayload json.RawMessage `json:"human_edited_payload,omitempty"`
	AlternativeID      string          `json:"alternative_id,omitempty"`
}

type StopWorkflowRequest struct {
	PlanID string `json:"plan_id"`
	Reason string `json:"reason"`
}

type HumanDecisionCenterSummary struct {
	PendingAwaitingUser int64 `json:"pending_awaiting_user"`
	HighRiskWorkflows   int64 `json:"high_risk_workflows"`
	LowConfidenceCount  int64 `json:"low_confidence_count"`
	ActiveEscalations   int64 `json:"active_escalations"`
	ModifiedByMeCount   int64 `json:"modified_by_me_count"`
	ApprovedByMeCount   int64 `json:"approved_by_me_count"`
	RejectedByMeCount   int64 `json:"rejected_by_me_count"`
	StoppedWorkflows    int64 `json:"stopped_workflows"`
	WaitingOnHumanCount int64 `json:"waiting_on_human_count"`
}

type SidecarDecisionAnalysisRequest struct {
	OrgID              int64                  `json:"org_id"`
	Module             string                 `json:"module"`
	EntityType         string                 `json:"entity_type"`
	EntityID           string                 `json:"entity_id"`
	Facts              []string               `json:"facts"`
	Predictions        []string               `json:"predictions"`
	AuthoritativeState map[string]interface{} `json:"authoritative_state,omitempty"`
	AutonomyLevel      AutonomyLevel          `json:"autonomy_level"`
	CorrelationID      string                 `json:"correlation_id,omitempty"`
}

type SidecarDecisionAnalysisResponse struct {
	DecisionID            string                 `json:"decision_id"`
	OperatingMode         string                 `json:"operating_mode"`
	Title                 string                 `json:"title"`
	ContextSummary        string                 `json:"context_summary"`
	Facts                 map[string]interface{} `json:"facts"`
	Predictions           map[string]interface{} `json:"predictions"`
	AIRecommendation      string                 `json:"ai_recommendation"`
	PreparedPayload       map[string]interface{} `json:"prepared_payload"`
	Alternatives          []map[string]interface{}`json:"alternatives"`
	Confidence            string                 `json:"confidence"`
	DataSufficiency       string                 `json:"data_sufficiency"`
	RiskLevel             string                 `json:"risk_level"`
	RequiresHumanApproval bool                   `json:"requires_human_approval"`
	IsReversible          bool                   `json:"is_reversible"`
	ProvenanceSummary     string                 `json:"provenance_summary"`
	CorrelationID         string                 `json:"correlation_id"`
}

type SidecarFeedbackAnalysisRequest struct {
	OrgID             int64   `json:"org_id"`
	DecisionID        string  `json:"decision_id"`
	HumanDecision     string  `json:"human_decision"`
	FeedbackType      string  `json:"feedback_type"`
	FeedbackReason    *string `json:"feedback_reason,omitempty"`
	OriginalAIPayload *string `json:"original_ai_payload,omitempty"`
	HumanFinalPayload *string `json:"human_final_payload,omitempty"`
}

type SidecarFeedbackAnalysisResponse struct {
	DecisionID        string                 `json:"decision_id"`
	LearnedPreference string                 `json:"learned_preference"`
	MemoryCandidate   map[string]interface{} `json:"memory_candidate,omitempty"`
	ReplanSuggested   bool                   `json:"replan_suggested"`
	Notes             string                 `json:"notes"`
	CorrelationID     string                 `json:"correlation_id"`
}

// Phase 5 Task 5.12: Autonomous Operations Command Center Models

type CommandCenterOverviewDTO struct {
	ActiveShipments       int            `json:"active_shipments"`
	ShipmentsAtRisk       int            `json:"shipments_at_risk"`
	ActiveExceptions      int            `json:"active_exceptions"`
	CriticalExceptions    int            `json:"critical_exceptions"`
	ActiveWorkflows       int            `json:"active_workflows"`
	WorkflowsWaitingHuman int            `json:"workflows_waiting_human"`
	PendingApprovals      int            `json:"pending_approvals"`
	EscalationsCount      int            `json:"escalations_count"`
	FailedActionsCount    int            `json:"failed_actions_count"`
	StalledPlansCount     int            `json:"stalled_plans_count"`
	ActualSummary         string         `json:"actual_summary"`
	PredictedSummary      string         `json:"predicted_summary"`
	AIAnalysisSummary     string         `json:"ai_analysis_summary"`
	AutonomyDistribution map[string]int `json:"autonomy_distribution"`
	SystemHealthStatus    string         `json:"system_health_status"`
	LastUpdated           time.Time      `json:"last_updated"`
	IsStale               bool           `json:"is_stale"`
}

type CriticalAttentionItemDTO struct {
	ID                string     `json:"id"`
	PriorityRank      int        `json:"priority_rank"`
	PriorityScore     float64    `json:"priority_score"`
	PriorityTier      string     `json:"priority_tier"`
	Severity          string     `json:"severity"`
	EntityType        string     `json:"entity_type"`
	EntityID          string     `json:"entity_id"`
	EntityReference   string     `json:"entity_reference"`
	Title             string     `json:"title"`
	IssueSummary      string     `json:"issue_summary"`
	WhyFlagged        string     `json:"why_flagged"`
	ActualFacts       string     `json:"actual_facts"`
	PredictedImpact   string     `json:"predicted_impact"`
	RecommendedAction string     `json:"recommended_action"`
	Impact            string     `json:"impact"`
	RequiredAction    string     `json:"required_action"`
	Owner             string     `json:"owner"`
	Deadline          *time.Time `json:"deadline,omitempty"`
	Urgency           string     `json:"urgency"`
	Source            string     `json:"source"`
	RequiresHuman     bool       `json:"requires_human"`
	CreatedAt         time.Time  `json:"created_at"`
}

type DomainRiskItemDTO struct {
	EntityID          string    `json:"entity_id"`
	EntityReference   string    `json:"entity_reference"`
	Issue             string    `json:"issue"`
	Severity          string    `json:"severity"`
	ActualFact        string    `json:"actual_fact"`
	PredictedRisk     string    `json:"predicted_risk"`
	RecommendedAction string    `json:"recommended_action"`
	Status            string    `json:"status"`
	HasActivePlan     bool      `json:"has_active_plan"`
	PlanID            *string   `json:"plan_id,omitempty"`
	LastEvent         string    `json:"last_event"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type DomainRiskSummaryDTO struct {
	Domain                  string              `json:"domain"`
	TotalAtRisk             int                 `json:"total_at_risk"`
	CriticalCount           int                 `json:"critical_count"`
	HighCount               int                 `json:"high_count"`
	AuthoritativeState      string              `json:"authoritative_state"`
	PredictedRisk           string              `json:"predicted_risk"`
	ActiveRecoveryWorkflows int                 `json:"active_recovery_workflows"`
	Items                   []DomainRiskItemDTO `json:"items"`
}

type CommandCenterActionDTO struct {
	ActionID      string    `json:"action_id"`
	PlanID        string    `json:"plan_id"`
	StepID        string    `json:"step_id"`
	Title         string    `json:"title"`
	ActionType    string    `json:"action_type"`
	Status        string    `json:"status"`
	ExecutionMode string    `json:"execution_mode"`
	ResultSummary string    `json:"result_summary"`
	Timestamp     time.Time `json:"timestamp"`
}

type CommandCenterReplanDTO struct {
	ReplanID           string    `json:"replan_id"`
	PlanID             string    `json:"plan_id"`
	EntityType         string    `json:"entity_type"`
	EntityID           string    `json:"entity_id"`
	TriggerReason      string    `json:"trigger_reason"`
	OldStatus          string    `json:"old_status"`
	NewStatus          string    `json:"new_status"`
	ChangedAssumptions string    `json:"changed_assumptions"`
	Timestamp          time.Time `json:"timestamp"`
}

type CommandCenterEscalationDTO struct {
	EscalationID           string     `json:"escalation_id"`
	EntityType             string     `json:"entity_type"`
	EntityID               string     `json:"entity_id"`
	Severity               string     `json:"severity"`
	Reason                 string     `json:"reason"`
	Owner                  string     `json:"owner"`
	Deadline               *time.Time `json:"deadline,omitempty"`
	RecommendedHumanAction string     `json:"recommended_human_action"`
	Timestamp              time.Time  `json:"timestamp"`
}

type CommandCenterActivityDTO struct {
	RecentActions    []CommandCenterActionDTO     `json:"recent_actions"`
	ReplanningEvents []CommandCenterReplanDTO     `json:"replanning_events"`
	Escalations      []CommandCenterEscalationDTO `json:"escalations"`
}

type SubsystemHealthDTO struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"` // HEALTHY, DEGRADED, UNAVAILABLE
	LatencyMs   int64     `json:"latency_ms"`
	Message     string    `json:"message"`
	LastChecked time.Time `json:"last_checked"`
}

type SystemHealthStatusDTO struct {
	OverallStatus string                        `json:"overall_status"`
	Subsystems    map[string]SubsystemHealthDTO `json:"subsystems"`
	LastCheckedAt time.Time                     `json:"last_checked_at"`
}

type SidecarPrioritizeRequest struct {
	OrgID         int64                    `json:"org_id"`
	Items         []map[string]interface{} `json:"items"`
	CorrelationID *string                  `json:"correlation_id,omitempty"`
}

type SidecarPrioritizedItemDTO struct {
	ID                string  `json:"id"`
	PriorityRank      int     `json:"priority_rank"`
	PriorityScore     float64 `json:"priority_score"`
	PriorityTier      string  `json:"priority_tier"`
	Severity          string  `json:"severity"`
	EntityType        string  `json:"entity_type"`
	EntityID          string  `json:"entity_id"`
	Title             string  `json:"title"`
	IssueSummary      string  `json:"issue_summary"`
	WhyFlagged        string  `json:"why_flagged"`
	ActualFacts       string  `json:"actual_facts"`
	PredictedImpact   string  `json:"predicted_impact"`
	RecommendedAction string  `json:"recommended_action"`
	Owner             string  `json:"owner"`
	Deadline          *string `json:"deadline,omitempty"`
	Urgency           string  `json:"urgency"`
	Source            string  `json:"source"`
	RequiresHuman     bool    `json:"requires_human"`
}

type SidecarPrioritizeResponse struct {
	Items            []SidecarPrioritizedItemDTO `json:"items"`
	CriticalCount    int                         `json:"critical_count"`
	HighCount        int                         `json:"high_count"`
	ExecutiveSummary string                      `json:"executive_summary"`
	CorrelationID    string                      `json:"correlation_id"`
}

// ============================================================================
// Phase 5 Task 5.13: Agent Memory and Learning from Outcomes Models
// ============================================================================

type AgentOutcome struct {
	ID                  int64           `db:"id" json:"id"`
	OrgID               int64           `db:"org_id" json:"org_id"`
	OutcomeID           string          `db:"outcome_id" json:"outcome_id"`
	SourceEntityType    string          `db:"source_entity_type" json:"source_entity_type"`
	SourceEntityID      string          `db:"source_entity_id" json:"source_entity_id"`
	WorkflowID          sql.NullString  `db:"workflow_id" json:"workflow_id"`
	PlanID              sql.NullString  `db:"plan_id" json:"plan_id"`
	PlanVersion         int             `db:"plan_version" json:"plan_version"`
	StepID              sql.NullString  `db:"step_id" json:"step_id"`
	ActionID            sql.NullString  `db:"action_id" json:"action_id"`
	ActionType          sql.NullString  `db:"action_type" json:"action_type"`
	OutcomeType         string          `db:"outcome_type" json:"outcome_type"`
	ExpectedResult      sql.NullString  `db:"expected_result" json:"expected_result"`
	ActualResult        sql.NullString  `db:"actual_result" json:"actual_result"`
	Status              string          `db:"status" json:"status"`
	IsVerified          bool            `db:"is_verified" json:"is_verified"`
	VerifiedAt          sql.NullTime    `db:"verified_at" json:"verified_at"`
	VerificationMethod  sql.NullString  `db:"verification_method" json:"verification_method"`
	TimeToResolutionSec sql.NullInt64   `db:"time_to_resolution_sec" json:"time_to_resolution_sec"`
	FailureCategory     string          `db:"failure_category" json:"failure_category"`
	Reason              sql.NullString  `db:"reason" json:"reason"`
	HumanInvolvement    string          `db:"human_involvement" json:"human_involvement"`
	DecidedByID         sql.NullInt64   `db:"decided_by_id" json:"decided_by_id"`
	DecidedByName       sql.NullString  `db:"decided_by_name" json:"decided_by_name"`
	ConfidenceScore     float64         `db:"confidence_score" json:"confidence_score"`
	CorrelationID       sql.NullString  `db:"correlation_id" json:"correlation_id"`
	Metadata            json.RawMessage `db:"metadata" json:"metadata,omitempty"`
	CreatedAt           time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time       `db:"updated_at" json:"updated_at"`
}

type LearnedPattern struct {
	ID                     int64     `db:"id" json:"id"`
	OrgID                  int64     `db:"org_id" json:"org_id"`
	PatternID              string    `db:"pattern_id" json:"pattern_id"`
	PatternType            string    `db:"pattern_type" json:"pattern_type"`
	EntityType             string    `db:"entity_type" json:"entity_type"`
	EntityIdentifier       string    `db:"entity_identifier" json:"entity_identifier"`
	Title                  string    `db:"title" json:"title"`
	Description            string    `db:"description" json:"description"`
	RecommendedStrategy    *string   `db:"recommended_strategy" json:"recommended_strategy,omitempty"`
	SupportingObservations int       `db:"supporting_observations" json:"supporting_observations"`
	SuccessRate            float64   `db:"success_rate" json:"success_rate"`
	Confidence             string    `db:"confidence" json:"confidence"`
	Scope                  string    `db:"scope" json:"scope"`
	IsActive               bool      `db:"is_active" json:"is_active"`
	LastObservedAt         time.Time `db:"last_observed_at" json:"last_observed_at"`
	CreatedAt              time.Time `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time `db:"updated_at" json:"updated_at"`
}

type ExtendedMemoryItem struct {
	ID                  int64           `db:"id" json:"id"`
	OrgID               int64           `db:"org_id" json:"org_id"`
	UserID              int64           `db:"user_id" json:"user_id"`
	Scope               string          `db:"scope" json:"scope"`
	Category            string          `db:"category" json:"category"`
	MemoryType          string          `db:"memory_type" json:"memory_type"`
	Title               string          `db:"title" json:"title"`
	Content             string          `db:"content" json:"content"`
	OriginalContent     sql.NullString  `db:"original_content" json:"original_content"`
	StructuredValue     json.RawMessage `db:"structured_value" json:"structured_value,omitempty"`
	EntityType          sql.NullString  `db:"entity_type" json:"entity_type"`
	EntityID            sql.NullString  `db:"entity_id" json:"entity_id"`
	OutcomeID           sql.NullString  `db:"outcome_id" json:"outcome_id"`
	Confidence          float64         `db:"confidence" json:"confidence"`
	RecencyWeight       float64         `db:"recency_weight" json:"recency_weight"`
	TimesObserved       int             `db:"times_observed" json:"times_observed"`
	TimesUsed           int             `db:"times_used" json:"times_used"`
	SuccessCount        int             `db:"success_count" json:"success_count"`
	FailureCount        int             `db:"failure_count" json:"failure_count"`
	IsStale             bool            `db:"is_stale" json:"is_stale"`
	InvalidatedAt       sql.NullTime    `db:"invalidated_at" json:"invalidated_at"`
	InvalidationReason  sql.NullString  `db:"invalidation_reason" json:"invalidation_reason"`
	InvalidatedByID     sql.NullInt64   `db:"invalidated_by_id" json:"invalidated_by_id"`
	ConflictStatus      string          `db:"conflict_status" json:"conflict_status"`
	SupersededByID      sql.NullInt64   `db:"superseded_by_id" json:"superseded_by_id"`
	ProvenanceType      string          `db:"provenance_type" json:"provenance_type"`
	SourceType          string          `db:"source_type" json:"source_type"`
	SourceReference     sql.NullString  `db:"source_reference" json:"source_reference"`
	Evidence            sql.NullString  `db:"evidence" json:"evidence"`
	ExplicitlyConfirmed bool            `db:"explicitly_confirmed" json:"explicitly_confirmed"`
	Status              string          `db:"status" json:"status"`
	ReviewAt            sql.NullTime    `db:"review_at" json:"review_at"`
	ExpiresAt           sql.NullTime    `db:"expires_at" json:"expires_at"`
	LastUsedAt          sql.NullTime    `db:"last_used_at" json:"last_used_at"`
	CreatedBy           sql.NullString  `db:"created_by" json:"created_by"`
	UpdatedBy           sql.NullString  `db:"updated_by" json:"updated_by"`
	CorrelationID       sql.NullString  `db:"correlation_id" json:"correlation_id"`
	CreatedAt           time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time       `db:"updated_at" json:"updated_at"`
}

func (o AgentOutcome) MarshalJSON() ([]byte, error) {
	type Alias AgentOutcome
	var workflowID, planID, stepID, actionID, actionType, expRes, actRes, verMeth, reason, decidedByName, corrID *string
	var verAt *time.Time
	var timeRes, decByID *int64

	if o.WorkflowID.Valid { workflowID = &o.WorkflowID.String }
	if o.PlanID.Valid { planID = &o.PlanID.String }
	if o.StepID.Valid { stepID = &o.StepID.String }
	if o.ActionID.Valid { actionID = &o.ActionID.String }
	if o.ActionType.Valid { actionType = &o.ActionType.String }
	if o.ExpectedResult.Valid { expRes = &o.ExpectedResult.String }
	if o.ActualResult.Valid { actRes = &o.ActualResult.String }
	if o.VerificationMethod.Valid { verMeth = &o.VerificationMethod.String }
	if o.Reason.Valid { reason = &o.Reason.String }
	if o.DecidedByName.Valid { decidedByName = &o.DecidedByName.String }
	if o.CorrelationID.Valid { corrID = &o.CorrelationID.String }
	if o.VerifiedAt.Valid { verAt = &o.VerifiedAt.Time }
	if o.TimeToResolutionSec.Valid { timeRes = &o.TimeToResolutionSec.Int64 }
	if o.DecidedByID.Valid { decByID = &o.DecidedByID.Int64 }

	return json.Marshal(&struct {
		Alias
		WorkflowID          *string    `json:"workflow_id,omitempty"`
		PlanID              *string    `json:"plan_id,omitempty"`
		StepID              *string    `json:"step_id,omitempty"`
		ActionID            *string    `json:"action_id,omitempty"`
		ActionType          *string    `json:"action_type,omitempty"`
		ExpectedResult      *string    `json:"expected_result,omitempty"`
		ActualResult        *string    `json:"actual_result,omitempty"`
		VerifiedAt          *time.Time `json:"verified_at,omitempty"`
		VerificationMethod  *string    `json:"verification_method,omitempty"`
		TimeToResolutionSec *int64     `json:"time_to_resolution_sec,omitempty"`
		Reason              *string    `json:"reason,omitempty"`
		DecidedByID         *int64     `json:"decided_by_id,omitempty"`
		DecidedByName       *string    `json:"decided_by_name,omitempty"`
		CorrelationID       *string    `json:"correlation_id,omitempty"`
	}{
		Alias:               Alias(o),
		WorkflowID:          workflowID,
		PlanID:              planID,
		StepID:              stepID,
		ActionID:            actionID,
		ActionType:          actionType,
		ExpectedResult:      expRes,
		ActualResult:        actRes,
		VerifiedAt:          verAt,
		VerificationMethod:  verMeth,
		TimeToResolutionSec: timeRes,
		Reason:              reason,
		DecidedByID:         decByID,
		DecidedByName:       decidedByName,
		CorrelationID:       corrID,
	})
}

func (m ExtendedMemoryItem) MarshalJSON() ([]byte, error) {
	type Alias ExtendedMemoryItem
	var origContent, entityType, entityID, outcomeID, invReason, srcRef, evidence, createdBy, updatedBy, corrID *string
	var invAt, revAt, expAt, usedAt *time.Time
	var invByID, supByID *int64

	if m.OriginalContent.Valid { origContent = &m.OriginalContent.String }
	if m.EntityType.Valid { entityType = &m.EntityType.String }
	if m.EntityID.Valid { entityID = &m.EntityID.String }
	if m.OutcomeID.Valid { outcomeID = &m.OutcomeID.String }
	if m.InvalidationReason.Valid { invReason = &m.InvalidationReason.String }
	if m.SourceReference.Valid { srcRef = &m.SourceReference.String }
	if m.Evidence.Valid { evidence = &m.Evidence.String }
	if m.CreatedBy.Valid { createdBy = &m.CreatedBy.String }
	if m.UpdatedBy.Valid { updatedBy = &m.UpdatedBy.String }
	if m.CorrelationID.Valid { corrID = &m.CorrelationID.String }

	if m.InvalidatedAt.Valid { invAt = &m.InvalidatedAt.Time }
	if m.ReviewAt.Valid { revAt = &m.ReviewAt.Time }
	if m.ExpiresAt.Valid { expAt = &m.ExpiresAt.Time }
	if m.LastUsedAt.Valid { usedAt = &m.LastUsedAt.Time }

	if m.InvalidatedByID.Valid { invByID = &m.InvalidatedByID.Int64 }
	if m.SupersededByID.Valid { supByID = &m.SupersededByID.Int64 }

	return json.Marshal(&struct {
		Alias
		OriginalContent    *string    `json:"original_content,omitempty"`
		EntityType         *string    `json:"entity_type,omitempty"`
		EntityID           *string    `json:"entity_id,omitempty"`
		OutcomeID          *string    `json:"outcome_id,omitempty"`
		InvalidatedAt      *time.Time `json:"invalidated_at,omitempty"`
		InvalidationReason *string    `json:"invalidation_reason,omitempty"`
		InvalidatedByID    *int64     `json:"invalidated_by_id,omitempty"`
		SupersededByID     *int64     `json:"superseded_by_id,omitempty"`
		SourceReference    *string    `json:"source_reference,omitempty"`
		Evidence           *string    `json:"evidence,omitempty"`
		ReviewAt           *time.Time `json:"review_at,omitempty"`
		ExpiresAt          *time.Time `json:"expires_at,omitempty"`
		LastUsedAt         *time.Time `json:"last_used_at,omitempty"`
		CreatedBy          *string    `json:"created_by,omitempty"`
		UpdatedBy          *string    `json:"updated_by,omitempty"`
		CorrelationID      *string    `json:"correlation_id,omitempty"`
	}{
		Alias:              Alias(m),
		OriginalContent:    origContent,
		EntityType:         entityType,
		EntityID:           entityID,
		OutcomeID:          outcomeID,
		InvalidatedAt:      invAt,
		InvalidationReason: invReason,
		InvalidatedByID:    invByID,
		SupersededByID:     supByID,
		SourceReference:    srcRef,
		Evidence:           evidence,
		ReviewAt:           revAt,
		ExpiresAt:          expAt,
		LastUsedAt:         usedAt,
		CreatedBy:          createdBy,
		UpdatedBy:          updatedBy,
		CorrelationID:      corrID,
	})
}


type RecordOutcomeRequest struct {
	SourceEntityType    string                 `json:"source_entity_type"`
	SourceEntityID      string                 `json:"source_entity_id"`
	WorkflowID          *string                `json:"workflow_id,omitempty"`
	PlanID              *string                `json:"plan_id,omitempty"`
	PlanVersion         int                    `json:"plan_version,omitempty"`
	StepID              *string                `json:"step_id,omitempty"`
	ActionID            *string                `json:"action_id,omitempty"`
	ActionType          *string                `json:"action_type,omitempty"`
	OutcomeType         string                 `json:"outcome_type"`
	ExpectedResult      string                 `json:"expected_result"`
	ActualResult        string                 `json:"actual_result"`
	Status              string                 `json:"status"` // SUCCESS, PARTIAL_SUCCESS, FAILED, etc.
	IsVerified          bool                   `json:"is_verified"`
	VerificationMethod  *string                `json:"verification_method,omitempty"`
	TimeToResolutionSec *int                   `json:"time_to_resolution_sec,omitempty"`
	FailureCategory     *string                `json:"failure_category,omitempty"`
	Reason              *string                `json:"reason,omitempty"`
	HumanInvolvement    *string                `json:"human_involvement,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
	CorrelationID       *string                `json:"correlation_id,omitempty"`
}

type VerifyOutcomeRequest struct {
	ActualResult       *string `json:"actual_result,omitempty"`
	VerificationMethod string  `json:"verification_method"`
	Status             *string `json:"status,omitempty"`
}

type CorrectMemoryRequest struct {
	CorrectedContent string `json:"corrected_content"`
	Reason           string `json:"reason"`
}

type InvalidateMemoryRequest struct {
	Reason string `json:"reason"`
}

type MemoryLearningSummaryDTO struct {
	TotalMemories           int                  `json:"total_memories"`
	ActiveMemories          int                  `json:"active_memories"`
	StaleMemories           int                  `json:"stale_memories"`
	InvalidatedMemories     int                  `json:"invalidated_memories"`
	TotalOutcomes           int                  `json:"total_outcomes"`
	VerifiedOutcomes        int                  `json:"verified_outcomes"`
	SuccessfulOutcomes      int                  `json:"successful_outcomes"`
	FailedOutcomes          int                  `json:"failed_outcomes"`
	DetectedPatterns        int                  `json:"detected_patterns"`
	OverallSuccessRate      float64              `json:"overall_success_rate"`
	RecommendationAcceptPct float64              `json:"recommendation_accept_pct"`
	MemoryCategoryCounts    map[string]int       `json:"memory_category_counts"`
	TopPatterns             []LearnedPattern     `json:"top_patterns"`
	RecentOutcomes          []AgentOutcome       `json:"recent_outcomes"`
}

// Sidecar DTOs for Memory & Learning
type SidecarOutcomeEvaluationRequest struct {
	OrgID               int64                  `json:"org_id"`
	SourceEntityType    string                 `json:"source_entity_type"`
	SourceEntityID      string                 `json:"source_entity_id"`
	OutcomeType         string                 `json:"outcome_type"`
	PlanID              *string                `json:"plan_id,omitempty"`
	StepID              *string                `json:"step_id,omitempty"`
	ActionType          *string                `json:"action_type,omitempty"`
	ExpectedResult      string                 `json:"expected_result"`
	ActualResult        string                 `json:"actual_result"`
	Status              string                 `json:"status"`
	TimeToResolutionSec *int                   `json:"time_to_resolution_sec,omitempty"`
	HumanInvolvement    string                 `json:"human_involvement"`
	DecidedByName       *string                `json:"decided_by_name,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
	CorrelationID       *string                `json:"correlation_id,omitempty"`
}

type SidecarMemoryCandidateDTO struct {
	OrgID          int64                  `json:"org_id"`
	Scope          string                 `json:"scope"`
	Category       string                 `json:"category"`
	MemoryType     string                 `json:"memory_type"`
	Title          string                 `json:"title"`
	Content        string                 `json:"content"`
	StructuredValue map[string]interface{} `json:"structured_value,omitempty"`
	EntityType     *string                `json:"entity_type,omitempty"`
	EntityID       *string                `json:"entity_id,omitempty"`
	OutcomeID      *string                `json:"outcome_id,omitempty"`
	Confidence     string                 `json:"confidence"`
	ConfidenceScore float64               `json:"confidence_score"`
	RecencyWeight  float64                `json:"recency_weight"`
	ProvenanceType string                 `json:"provenance_type"`
	SourceReference *string               `json:"source_reference,omitempty"`
	Evidence       *string                `json:"evidence,omitempty"`
	Status         string                 `json:"status"`
}

type SidecarOutcomeEvaluationResponse struct {
	OutcomeStatus      string                     `json:"outcome_status"`
	IsVerified         bool                       `json:"is_verified"`
	EvaluationSummary  string                     `json:"evaluation_summary"`
	FailureCategory    string                     `json:"failure_category"`
	Confidence         string                     `json:"confidence"`
	ConfidenceScore    float64                    `json:"confidence_score"`
	ShouldCreateMemory bool                       `json:"should_create_memory"`
	MemoryCandidate    *SidecarMemoryCandidateDTO `json:"memory_candidate,omitempty"`
	CorrelationID      string                     `json:"correlation_id"`
}

type SidecarMemoryRetrievalRequest struct {
	OrgID             int64                    `json:"org_id"`
	QueryContext      string                   `json:"query_context"`
	Module            *string                  `json:"module,omitempty"`
	Category          *string                  `json:"category,omitempty"`
	EntityType        *string                  `json:"entity_type,omitempty"`
	EntityID          *string                  `json:"entity_id,omitempty"`
	CarrierSCAC       *string                  `json:"carrier_scac,omitempty"`
	CustomerID        *string                  `json:"customer_id,omitempty"`
	Limit             int                      `json:"limit"`
	IncludeStale      bool                     `json:"include_stale"`
	CandidateMemories []map[string]interface{} `json:"candidate_memories"`
	CorrelationID     *string                  `json:"correlation_id,omitempty"`
}

type SidecarMemoryRetrievalResponse struct {
	RetrievedMemories   []map[string]interface{} `json:"retrieved_memories"`
	TotalFound          int                      `json:"total_found"`
	ContextSummary      string                   `json:"context_summary"`
	ProvenanceBreakdown map[string]int           `json:"provenance_breakdown"`
	HasConflicts        bool                     `json:"has_conflicts"`
	ConflictWarnings    []string                 `json:"conflict_warnings"`
	CorrelationID       string                   `json:"correlation_id"`
}

type SidecarPatternDetectionRequest struct {
	OrgID         int64                    `json:"org_id"`
	Outcomes      []map[string]interface{} `json:"outcomes"`
	Memories      []map[string]interface{} `json:"memories"`
	CorrelationID *string                  `json:"correlation_id,omitempty"`
}

type SidecarPatternDetectionResponse struct {
	DetectedPatterns []map[string]interface{} `json:"detected_patterns"`
	TotalPatterns    int                      `json:"total_patterns"`
	Summary          string                   `json:"summary"`
	CorrelationID    string                   `json:"correlation_id"`
}

type SidecarMemoryConflictRequest struct {
	OrgID            int64                    `json:"org_id"`
	NewObservation   string                   `json:"new_observation"`
	ExistingMemories []map[string]interface{} `json:"existing_memories"`
	CorrelationID    *string                  `json:"correlation_id,omitempty"`
}

type SidecarMemoryConflictResponse struct {
	HasConflict             bool    `json:"has_conflict"`
	ConflictingMemoryID     *int64  `json:"conflicting_memory_id,omitempty"`
	ConflictExplanation     string  `json:"conflict_explanation"`
	AuthoritativeResolution string  `json:"authoritative_resolution"`
	RecommendedAction       string  `json:"recommended_action"`
	CorrelationID           string  `json:"correlation_id"`
}










