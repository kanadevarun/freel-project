package enterprise_autonomy

import (
	"errors"
	"time"
)

// ExceptionDomain represents the cross-enterprise business domains for exceptions
type ExceptionDomain string

const (
	DomainShipment   ExceptionDomain = "SHIPMENT"
	DomainCarrier    ExceptionDomain = "CARRIER"
	DomainCustomer   ExceptionDomain = "CUSTOMER"
	DomainCommercial ExceptionDomain = "COMMERCIAL"
	DomainFinance    ExceptionDomain = "FINANCE"
	DomainContract   ExceptionDomain = "CONTRACT"
	DomainCompliance ExceptionDomain = "COMPLIANCE"
)

// ExceptionSeverity defines standard business criticality levels
type ExceptionSeverity string

const (
	SeverityLow      ExceptionSeverity = "LOW"
	SeverityMedium   ExceptionSeverity = "MEDIUM"
	SeverityHigh     ExceptionSeverity = "HIGH"
	SeverityCritical ExceptionSeverity = "CRITICAL"
)

// ExceptionLifecycleStage defines the governed stages from detection to resolution and learning
type ExceptionLifecycleStage string

const (
	StageExceptionDetected        ExceptionLifecycleStage = "DETECTED"
	StageExceptionInvestigating   ExceptionLifecycleStage = "INVESTIGATING"
	StageExceptionImpactAssessing ExceptionLifecycleStage = "IMPACT_ASSESSMENT"
	StageExceptionPlanningRecovery ExceptionLifecycleStage = "PLANNING_RECOVERY"
	StageExceptionWaitingApproval ExceptionLifecycleStage = "WAITING_FOR_APPROVAL"
	StageExceptionExecuting       ExceptionLifecycleStage = "EXECUTING"
	StageExceptionVerifying       ExceptionLifecycleStage = "VERIFYING"
	StageExceptionMonitoring      ExceptionLifecycleStage = "MONITORING"
	StageExceptionResolved        ExceptionLifecycleStage = "RESOLVED"
	StageExceptionEscalated       ExceptionLifecycleStage = "ESCALATED"
	StageExceptionClosed          ExceptionLifecycleStage = "CLOSED"
)

var (
	ErrInvalidExceptionStageTransition = errors.New("invalid enterprise exception lifecycle stage transition")
	ErrExceptionWorkflowNotFound       = errors.New("autonomous enterprise exception workflow not found")
	ErrExceptionApprovalRequired       = errors.New("recovery action requires human approval by enterprise autonomy policy")
	ErrActionVerificationUnconfirmed   = errors.New("recovery action verification unconfirmed by authoritative business records")
	ErrResolutionEvidenceMissing       = errors.New("exception resolution requires authoritative verification proof")
	ErrMaxRecoveryRetriesExceeded      = errors.New("maximum adaptive recovery retries exceeded; escalating to human operations")
	ErrExceptionLoopDetected           = errors.New("autonomous exception loop detected; suppressed recursive event trigger")
)

// ValidateExceptionStageTransition enforces strict forward progression and governed adaptive loops
func ValidateExceptionStageTransition(current, next ExceptionLifecycleStage) error {
	if current == next {
		return nil
	}

	valid := false
	switch current {
	case StageExceptionDetected:
		valid = (next == StageExceptionInvestigating || next == StageExceptionEscalated || next == StageExceptionClosed)
	case StageExceptionInvestigating:
		valid = (next == StageExceptionImpactAssessing || next == StageExceptionPlanningRecovery || next == StageExceptionEscalated || next == StageExceptionClosed)
	case StageExceptionImpactAssessing:
		valid = (next == StageExceptionPlanningRecovery || next == StageExceptionEscalated || next == StageExceptionClosed)
	case StageExceptionPlanningRecovery:
		valid = (next == StageExceptionWaitingApproval || next == StageExceptionExecuting || next == StageExceptionEscalated || next == StageExceptionClosed)
	case StageExceptionWaitingApproval:
		// After approval, advance to Executing; if rejected, can Replan or Escalate or Close
		valid = (next == StageExceptionExecuting || next == StageExceptionPlanningRecovery || next == StageExceptionEscalated || next == StageExceptionClosed)
	case StageExceptionExecuting:
		valid = (next == StageExceptionVerifying || next == StageExceptionPlanningRecovery || next == StageExceptionEscalated)
	case StageExceptionVerifying:
		// If verification succeeds, advance to Monitoring or Resolved; if fails, adaptive replan or escalate
		valid = (next == StageExceptionMonitoring || next == StageExceptionResolved || next == StageExceptionPlanningRecovery || next == StageExceptionEscalated)
	case StageExceptionMonitoring:
		valid = (next == StageExceptionResolved || next == StageExceptionPlanningRecovery || next == StageExceptionEscalated)
	case StageExceptionResolved:
		valid = (next == StageExceptionClosed)
	case StageExceptionEscalated:
		// Human intervention can resume to Planning, or Close
		valid = (next == StageExceptionPlanningRecovery || next == StageExceptionExecuting || next == StageExceptionResolved || next == StageExceptionClosed)
	case StageExceptionClosed:
		valid = false // Terminal state
	default:
		valid = false
	}

	if !valid {
		return ErrInvalidExceptionStageTransition
	}
	return nil
}

// RootCauseAnalysis separates confirmed facts from likely and possible causes with confidence scoring
type RootCauseAnalysis struct {
	ConfirmedFacts      []string `json:"confirmed_facts"`
	LikelyCauses        []string `json:"likely_causes"`
	PossibleCauses      []string `json:"possible_causes"`
	PrimaryHypothesis   string   `json:"primary_hypothesis"`
	Confidence          float64  `json:"confidence"`
	Evidence            []string `json:"evidence"`
	ConflictingEvidence []string `json:"conflicting_evidence,omitempty"`
	DataSufficiency     string   `json:"data_sufficiency"` // HIGH, SUFFICIENT, LOW
}

// CrossModuleImpactAssessment details downstream operational, customer, financial, contractual, and compliance exposures
type CrossModuleImpactAssessment struct {
	// Operational Impact
	OperationalDelayHours  float64 `json:"operational_delay_hours"`
	RouteDeviated          bool    `json:"route_deviated"`
	EquipmentBlocked       bool    `json:"equipment_blocked"`
	OperationalSummary     string  `json:"operational_summary"`

	// Customer Impact
	CustomerServiceImpact  string  `json:"customer_service_impact"` // MINOR, MODERATE, SEVERE
	CustomerNotificationNeeded bool `json:"customer_notification_needed"`
	PredictedChurnRisk     string  `json:"predicted_churn_risk"`
	CustomerSummary        string  `json:"customer_summary"`

	// Financial Impact
	EstimatedCostImpact    float64 `json:"estimated_cost_impact"`
	MarginExposurePct      float64 `json:"margin_exposure_pct"`
	InvoiceDisputeRisk     string  `json:"invoice_dispute_risk"` // LOW, MEDIUM, HIGH
	FinancialSummary       string  `json:"financial_summary"`

	// Contract & SLA Impact
	SLABreached            bool    `json:"sla_breached"`
	ContractPenaltyRisk    float64 `json:"contract_penalty_risk"`
	ContractSummary        string  `json:"contract_summary"`

	// Compliance Impact
	RegulatoryRisk         string  `json:"regulatory_risk"` // NONE, LOW, HIGH, CRITICAL
	MissingDocumentation   []string `json:"missing_documentation,omitempty"`
	ComplianceSummary      string  `json:"compliance_summary"`

	OverallImpactScore     float64 `json:"overall_impact_score"` // 0.0 - 100.0
}

// ExceptionRecoveryOption models structured recovery options with required autonomy and governance flags
type ExceptionRecoveryOption struct {
	OptionID           string                 `json:"option_id"`
	Title              string                 `json:"title"`
	ActionType         string                 `json:"action_type"`
	Reason             string                 `json:"reason"`
	Evidence           []string               `json:"evidence"`
	ExpectedOutcome    string                 `json:"expected_outcome"`
	RiskLevel          string                 `json:"risk_level"` // LOW, MEDIUM, HIGH
	Confidence         float64                `json:"confidence"`
	RequiredCapability string                 `json:"required_capability"`
	RequiredAutonomy   string                 `json:"required_autonomy"` // LEVEL_1, LEVEL_2, LEVEL_3, LEVEL_4
	RequiresApproval   bool                   `json:"requires_approval"`
	Parameters         map[string]interface{} `json:"parameters,omitempty"`
	VerificationMethod string                 `json:"verification_method"`
}

// ExceptionVerificationResult provides authoritative verification proof of executed recovery actions
type ExceptionVerificationResult struct {
	StepID              string                 `json:"step_id"`
	ActionType          string                 `json:"action_type"`
	VerifiedSuccess     bool                   `json:"verified_success"`
	AuthoritativeSource string                 `json:"authoritative_source"` // e.g. "DATABASE_SHIPMENTS", "NOTIFICATION_SERVICE", "CARRIER_INTEGRATION"
	ProofData           map[string]interface{} `json:"proof_data,omitempty"`
	VerificationMessage string                 `json:"verification_message"`
	VerifiedAt          time.Time              `json:"verified_at"`
}

// AutonomousExceptionWorkflow is the durable entity orchestrating enterprise exception management
type AutonomousExceptionWorkflow struct {
	WorkflowID             string                         `json:"workflow_id"`
	OrgID                  int64                          `json:"org_id"`
	ExceptionID            string                         `json:"exception_id"`
	Domain                 ExceptionDomain                `json:"domain"`
	ExceptionType          string                         `json:"exception_type"`
	Severity               ExceptionSeverity              `json:"severity"`
	CurrentStage           ExceptionLifecycleStage        `json:"current_stage"`
	WorkflowState          EnterpriseWorkflowState        `json:"workflow_state"`
	RelatedEntityType      string                         `json:"related_entity_type"` // e.g. "SHIPMENT", "RFQ", "INVOICE", "CONTRACT"
	RelatedEntityID        string                         `json:"related_entity_id"`
	CorrelationID          string                         `json:"correlation_id"`
	ParentWorkflowID       *string                        `json:"parent_workflow_id,omitempty"`
	AssignedSpecialists    []string                       `json:"assigned_specialists"`
	RootCauseAnalysis      *RootCauseAnalysis             `json:"root_cause_analysis,omitempty"`
	ImpactAssessment       *CrossModuleImpactAssessment   `json:"impact_assessment,omitempty"`
	RecoveryOptions        []ExceptionRecoveryOption      `json:"recovery_options,omitempty"`
	SelectedOption         *ExceptionRecoveryOption       `json:"selected_option,omitempty"`
	RecoveryPlan           []EnterpriseWorkflowStep       `json:"recovery_plan,omitempty"`
	CurrentStepID          string                         `json:"current_step_id,omitempty"`
	PendingApprovalsCount  int                            `json:"pending_approvals_count"`
	ApprovalRequestID      *int64                         `json:"approval_request_id,omitempty"`
	ReplanVersion          int                            `json:"replan_version"`
	RetryCount             int                            `json:"retry_count"`
	MaxRetries             int                            `json:"max_retries"`
	LastVerificationResult *ExceptionVerificationResult   `json:"last_verification_result,omitempty"`
	MonitoringMilestone    string                         `json:"monitoring_milestone,omitempty"`
	MonitoringStable       bool                           `json:"monitoring_stable"`
	ResolutionProof        string                         `json:"resolution_proof,omitempty"`
	OutcomeRecorded        bool                           `json:"outcome_recorded"`
	OutcomeDetails         map[string]interface{}         `json:"outcome_details,omitempty"`
	CreatedAt              time.Time                      `json:"created_at"`
	UpdatedAt              time.Time                      `json:"updated_at"`
}

// EnterpriseExceptionEvent models incoming triggers across all modules
type EnterpriseExceptionEvent struct {
	EventID           string                 `json:"event_id"`
	EventType         string                 `json:"event_type"` // e.g. "SHIPMENT_DELAY_DETECTED", "CARRIER_FAILURE", "INVOICE_DISPUTE", "CONTRACT_SLA_BREACH", "CUSTOMS_DOC_MISSING"
	Domain            ExceptionDomain        `json:"domain"`
	RelatedEntityType string                 `json:"related_entity_type"`
	RelatedEntityID   string                 `json:"related_entity_id"`
	Severity          ExceptionSeverity      `json:"severity"`
	Title             string                 `json:"title"`
	Description       string                 `json:"description"`
	Payload           map[string]interface{} `json:"payload"`
	CorrelationID     string                 `json:"correlation_id,omitempty"`
	Source            string                 `json:"source,omitempty"`
}

// ExceptionOutcomeFeedback records real business resolution results to train workforce memory
type ExceptionOutcomeFeedback struct {
	WorkflowID          string  `json:"workflow_id"`
	ExceptionID         string  `json:"exception_id"`
	ResolvedSuccessfully bool    `json:"resolved_successfully"`
	OperationalRecoveryTimeHours float64 `json:"operational_recovery_time_hours"`
	ActualFinancialCost float64 `json:"actual_financial_cost"`
	CustomerSatisfaction string  `json:"customer_satisfaction,omitempty"`
	HumanIntervention   bool    `json:"human_intervention"`
	LessonLearned       string  `json:"lesson_learned,omitempty"`
}
