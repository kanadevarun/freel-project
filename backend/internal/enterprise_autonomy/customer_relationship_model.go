package enterprise_autonomy

import (
	"errors"
	"time"
)

// CustomerHealthCategory defines structured health categories
type CustomerHealthCategory string

const (
	CustomerHealthHealthy     CustomerHealthCategory = "HEALTHY"
	CustomerHealthStable      CustomerHealthCategory = "STABLE"
	CustomerHealthDeclining   CustomerHealthCategory = "DECLINING"
	CustomerHealthAtRisk      CustomerHealthCategory = "AT_RISK"
	CustomerHealthHighRisk    CustomerHealthCategory = "HIGH_RISK"
	CustomerHealthOpportunity CustomerHealthCategory = "OPPORTUNITY"
	CustomerHealthInactive    CustomerHealthCategory = "INACTIVE"
)

// CustomerLifecycleStage models continuous CRM workflow progression
type CustomerLifecycleStage string

const (
	StageCustomerMonitoring              CustomerLifecycleStage = "MONITORING"
	StageCustomerHealthAssessing         CustomerLifecycleStage = "HEALTH_ASSESSMENT"
	StageCustomerRiskOpportunityDetect   CustomerLifecycleStage = "RISK_OPPORTUNITY_DETECTION"
	StageCustomerMultiAgentInvestigation CustomerLifecycleStage = "INVESTIGATING"
	StageCustomerInterventionPlanning    CustomerLifecycleStage = "PLANNING_INTERVENTION"
	StageCustomerWaitingApproval         CustomerLifecycleStage = "WAITING_FOR_APPROVAL"
	StageCustomerExecutingIntervention   CustomerLifecycleStage = "EXECUTING"
	StageCustomerVerifying               CustomerLifecycleStage = "VERIFYING"
	StageCustomerOutcomeTracking         CustomerLifecycleStage = "OUTCOME_TRACKING"
	StageCustomerCompleted               CustomerLifecycleStage = "COMPLETED"
	StageCustomerEscalated               CustomerLifecycleStage = "ESCALATED"
)

var (
	ErrInvalidCustomerStageTransition = errors.New("invalid customer relationship lifecycle stage transition")
	ErrCustomerWorkflowNotFound       = errors.New("autonomous customer relationship workflow not found")
	ErrCustomerApprovalRequired       = errors.New("customer intervention requires human approval by enterprise autonomy policy")
	ErrCustomerVerificationFailed     = errors.New("customer intervention action unconfirmed by authoritative records")
	ErrDuplicateCustomerCommunication = errors.New("duplicate customer outreach suppressed by communication governance policy")
	ErrCustomerLoopDetected           = errors.New("customer intervention loop detected; circuit breaker activated")
)

// ValidateCustomerStageTransition enforces forward progression and governed adaptive intervention loops
func ValidateCustomerStageTransition(current, next CustomerLifecycleStage) error {
	if current == next {
		return nil
	}

	valid := false
	switch current {
	case StageCustomerMonitoring:
		valid = (next == StageCustomerHealthAssessing || next == StageCustomerEscalated || next == StageCustomerCompleted)
	case StageCustomerHealthAssessing:
		valid = (next == StageCustomerRiskOpportunityDetect || next == StageCustomerMonitoring || next == StageCustomerEscalated)
	case StageCustomerRiskOpportunityDetect:
		valid = (next == StageCustomerMultiAgentInvestigation || next == StageCustomerInterventionPlanning || next == StageCustomerMonitoring || next == StageCustomerEscalated)
	case StageCustomerMultiAgentInvestigation:
		valid = (next == StageCustomerInterventionPlanning || next == StageCustomerMonitoring || next == StageCustomerEscalated)
	case StageCustomerInterventionPlanning:
		valid = (next == StageCustomerWaitingApproval || next == StageCustomerExecutingIntervention || next == StageCustomerMonitoring || next == StageCustomerEscalated)
	case StageCustomerWaitingApproval:
		valid = (next == StageCustomerExecutingIntervention || next == StageCustomerInterventionPlanning || next == StageCustomerEscalated)
	case StageCustomerExecutingIntervention:
		valid = (next == StageCustomerVerifying || next == StageCustomerInterventionPlanning || next == StageCustomerEscalated)
	case StageCustomerVerifying:
		valid = (next == StageCustomerOutcomeTracking || next == StageCustomerInterventionPlanning || next == StageCustomerEscalated)
	case StageCustomerOutcomeTracking:
		valid = (next == StageCustomerMonitoring || next == StageCustomerCompleted || next == StageCustomerEscalated)
	case StageCustomerCompleted:
		valid = (next == StageCustomerMonitoring) // Can re-enter monitoring for subsequent account cycles
	case StageCustomerEscalated:
		valid = (next == StageCustomerInterventionPlanning || next == StageCustomerMonitoring || next == StageCustomerCompleted)
	default:
		valid = false
	}

	if !valid {
		return ErrInvalidCustomerStageTransition
	}
	return nil
}

// CustomerHealthProfile separates authoritative facts from predictions and recommendations
type CustomerHealthProfile struct {
	Category        CustomerHealthCategory `json:"category"`
	HealthScore     float64                `json:"health_score"` // 0.0 - 100.0
	ConfirmedFacts  []string               `json:"confirmed_facts"`
	AIPredictions   []string               `json:"ai_predictions"`
	Recommendations []string               `json:"recommendations"`
	Confidence      float64                `json:"confidence"`
	DataSufficiency string                 `json:"data_sufficiency"` // HIGH, SUFFICIENT, LOW
	EvaluatedAt     time.Time              `json:"evaluated_at"`
}

// CustomerRiskSignal models detected risk factors with provenance
type CustomerRiskSignal struct {
	SignalID        string   `json:"signal_id"`
	RiskType        string   `json:"risk_type"` // e.g. "CHURN_RISK", "VOLUME_DROP", "PAYMENT_DETERIORATION", "REPEAT_EXCEPTIONS"
	Severity        string   `json:"severity"`  // INFO, WARNING, CRITICAL
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Confidence      float64  `json:"confidence"`
	DataSufficiency string   `json:"data_sufficiency"`
	ConfirmedFacts  []string `json:"confirmed_facts"`
	LikelyCauses    []string `json:"likely_causes"`
	PossibleCauses  []string `json:"possible_causes"`
	Evidence        []string `json:"evidence"`
	DetectedAt      time.Time `json:"detected_at"`
}

// CustomerOpportunitySignal models positive growth and expansion vectors
type CustomerOpportunitySignal struct {
	SignalID          string   `json:"signal_id"`
	OpportunityType   string   `json:"opportunity_type"` // e.g. "VOLUME_EXPANSION", "CROSS_SELL_AIR", "CONTRACT_RENEWAL", "NEW_LANE"
	PotentialValue    float64  `json:"potential_value"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	RecommendedAction string   `json:"recommended_action"`
	Confidence        float64  `json:"confidence"`
	SupportingEvidence []string `json:"supporting_evidence"`
	DetectedAt        time.Time `json:"detected_at"`
}

// CustomerInterventionOption represents structured, policy-governed interventions
type CustomerInterventionOption struct {
	OptionID           string                 `json:"option_id"`
	Title              string                 `json:"title"`
	ActionType         string                 `json:"action_type"` // e.g. "customer.send_retention_email", "pricing.review_account_rates", "exception.service_recovery"
	Reason             string                 `json:"reason"`
	TargetAudience     string                 `json:"target_audience"`
	DraftMessage       string                 `json:"draft_message,omitempty"`
	Evidence           []string               `json:"evidence"`
	ExpectedOutcome    string                 `json:"expected_outcome"`
	RiskLevel          string                 `json:"risk_level"` // LOW, MEDIUM, HIGH
	Confidence         float64                `json:"confidence"`
	RequiredCapability string                 `json:"required_capability"`
	RequiredAutonomy   string                 `json:"required_autonomy"` // LEVEL_1 to LEVEL_4
	RequiresApproval   bool                   `json:"requires_approval"`
	VerificationMethod string                 `json:"verification_method"`
	HandoffModule      string                 `json:"handoff_module,omitempty"` // e.g. "PHASE_7_3_QUOTE_TO_CASH", "PHASE_7_4_EXCEPTION_MANAGEMENT"
	Parameters         map[string]interface{} `json:"parameters,omitempty"`
}

// CustomerResponseClassification represents classified incoming communication
type CustomerResponseClassification struct {
	RawMessage          string                 `json:"raw_message"`
	Intent              string                 `json:"intent"` // e.g. "POSITIVE", "COMPLAINT", "NEGOTIATION", "SUPPORT_INQUIRY", "QUOTE_REQUEST"
	Sentiment           string                 `json:"sentiment"` // POSITIVE, NEUTRAL, NEGATIVE, HIGHLY_DISSATISFIED
	Urgency             string                 `json:"urgency"`   // LOW, MEDIUM, HIGH, IMMEDIATE
	Confidence          float64                `json:"confidence"`
	ExtractedEntities   map[string]interface{} `json:"extracted_entities,omitempty"`
	RecommendedNextStep string                 `json:"recommended_next_step"`
	ProcessedAt         time.Time              `json:"processed_at"`
}

// AutonomousCustomerWorkflow represents the durable enterprise CRM orchestrator
type AutonomousCustomerWorkflow struct {
	WorkflowID             string                          `json:"workflow_id"`
	OrgID                  int64                           `json:"org_id"`
	CustomerID             int64                           `json:"customer_id"`
	CustomerName           string                          `json:"customer_name"`
	CustomerCode           string                          `json:"customer_code"`
	CurrentStage           CustomerLifecycleStage          `json:"current_stage"`
	WorkflowState          EnterpriseWorkflowState         `json:"workflow_state"`
	HealthProfile          *CustomerHealthProfile          `json:"health_profile,omitempty"`
	DetectedRisks          []CustomerRiskSignal            `json:"detected_risks,omitempty"`
	DetectedOpportunities  []CustomerOpportunitySignal      `json:"detected_opportunities,omitempty"`
	AssignedSpecialists    []string                        `json:"assigned_specialists"`
	InterventionOptions    []CustomerInterventionOption    `json:"intervention_options,omitempty"`
	SelectedOption         *CustomerInterventionOption     `json:"selected_option,omitempty"`
	InterventionPlan       []EnterpriseWorkflowStep        `json:"intervention_plan,omitempty"`
	PendingApprovalsCount  int                             `json:"pending_approvals_count"`
	ApprovalRequestID      *int64                          `json:"approval_request_id,omitempty"`
	LastResponse           *CustomerResponseClassification `json:"last_response,omitempty"`
	VerificationResult     map[string]interface{}          `json:"verification_result,omitempty"`
	ReplanVersion          int                             `json:"replan_version"`
	RetryCount             int                             `json:"retry_count"`
	MaxRetries             int                             `json:"max_retries"`
	LastOutreachAt         *time.Time                      `json:"last_outreach_at,omitempty"`
	OutcomeDetails         map[string]interface{}          `json:"outcome_details,omitempty"`
	CorrelationID          string                          `json:"correlation_id"`
	CreatedAt              time.Time                       `json:"created_at"`
	UpdatedAt              time.Time                       `json:"updated_at"`
}

// CustomerLifecycleEvent models incoming customer business triggers
type CustomerLifecycleEvent struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"` // e.g. "CUSTOMER_INQUIRY_RECEIVED", "QUOTE_ACCEPTED", "SHIPMENT_DELAY_NOTIFIED", "INVOICE_OVERDUE", "CUSTOMER_COMPLAINT"
	OrgID         int64                  `json:"org_id"`
	CustomerID    int64                  `json:"customer_id"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Source        string                 `json:"source"`
	Payload       map[string]interface{} `json:"payload"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
}

// CustomerOutcomeFeedback records real-world intervention outcome for workforce memory
type CustomerOutcomeFeedback struct {
	WorkflowID                string  `json:"workflow_id"`
	CustomerID                int64   `json:"customer_id"`
	InterventionSuccess       bool    `json:"intervention_success"`
	CustomerRetentionStatus   string  `json:"customer_retention_status"` // RETAINED, EXPANDED, CHURNED, STABLE
	RevenueImpact             float64 `json:"revenue_impact"`
	CustomerSatisfactionScore string  `json:"customer_satisfaction_score,omitempty"` // HIGH, MEDIUM, LOW
	LessonsLearned            string  `json:"lessons_learned,omitempty"`
}
