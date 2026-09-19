package enterprise_autonomy

import (
	"errors"
	"time"
)

// RevenueLifecycleStage represents governed stages for revenue and margin optimization
type RevenueLifecycleStage string

const (
	StageRevenueMonitoring             RevenueLifecycleStage = "MONITORING"
	StageRevenueCostAnalysis           RevenueLifecycleStage = "COST_MARGIN_ANALYSIS"
	StageRevenueCarrierEconomics       RevenueLifecycleStage = "CARRIER_ECONOMICS_EVALUATION"
	StageRevenueCustomerValue          RevenueLifecycleStage = "CUSTOMER_VALUE_ASSESSMENT"
	StageRevenueMultiAgentOptimization RevenueLifecycleStage = "OPTIMIZATION_REASONING"
	StageRevenueRecommendation         RevenueLifecycleStage = "RECOMMENDATION_FORMULATION"
	StageRevenueWaitingApproval        RevenueLifecycleStage = "WAITING_FOR_APPROVAL"
	StageRevenueExecutingAction        RevenueLifecycleStage = "EXECUTING"
	StageRevenueVerifying              RevenueLifecycleStage = "VERIFYING"
	StageRevenueOutcomeTracking        RevenueLifecycleStage = "OUTCOME_TRACKING"
	StageRevenueCompleted              RevenueLifecycleStage = "COMPLETED"
	StageRevenueEscalated              RevenueLifecycleStage = "ESCALATED"
)

var (
	ErrInvalidRevenueStageTransition = errors.New("invalid revenue/margin optimization lifecycle stage transition")
	ErrRevenueWorkflowNotFound       = errors.New("autonomous revenue optimization workflow not found")
	ErrPricingPolicyViolation        = errors.New("pricing recommendation violates minimum margin floor or maximum discount ceiling")
	ErrMarginThresholdBreached       = errors.New("actual or expected margin falls below enterprise margin risk threshold")
	ErrRevenueApprovalRequired       = ErrCommercialApprovalRequired
	ErrActionExecutionFailed         = errors.New("commercial action execution failed")
	ErrCommercialVerificationFailed  = errors.New("commercial optimization action unconfirmed by authoritative business records")
)

// ValidateRevenueStageTransition enforces governed forward progression and continuous re-evaluation loops
func ValidateRevenueStageTransition(current, next RevenueLifecycleStage) error {
	if current == next {
		return nil
	}

	valid := false
	switch current {
	case StageRevenueMonitoring:
		valid = (next == StageRevenueCostAnalysis || next == StageRevenueEscalated || next == StageRevenueCompleted)
	case StageRevenueCostAnalysis:
		valid = (next == StageRevenueCarrierEconomics || next == StageRevenueCustomerValue || next == StageRevenueMultiAgentOptimization || next == StageRevenueMonitoring || next == StageRevenueEscalated)
	case StageRevenueCarrierEconomics:
		valid = (next == StageRevenueCustomerValue || next == StageRevenueMultiAgentOptimization || next == StageRevenueMonitoring || next == StageRevenueEscalated)
	case StageRevenueCustomerValue:
		valid = (next == StageRevenueMultiAgentOptimization || next == StageRevenueRecommendation || next == StageRevenueMonitoring || next == StageRevenueEscalated)
	case StageRevenueMultiAgentOptimization:
		valid = (next == StageRevenueRecommendation || next == StageRevenueCostAnalysis || next == StageRevenueEscalated)
	case StageRevenueRecommendation:
		valid = (next == StageRevenueWaitingApproval || next == StageRevenueExecutingAction || next == StageRevenueMultiAgentOptimization || next == StageRevenueCostAnalysis || next == StageRevenueMonitoring || next == StageRevenueEscalated)
	case StageRevenueWaitingApproval:
		valid = (next == StageRevenueExecutingAction || next == StageRevenueRecommendation || next == StageRevenueEscalated)
	case StageRevenueExecutingAction:
		valid = (next == StageRevenueVerifying || next == StageRevenueRecommendation || next == StageRevenueEscalated)
	case StageRevenueVerifying:
		valid = (next == StageRevenueOutcomeTracking || next == StageRevenueRecommendation || next == StageRevenueEscalated)
	case StageRevenueOutcomeTracking:
		valid = (next == StageRevenueCompleted || next == StageRevenueMonitoring || next == StageRevenueEscalated)
	case StageRevenueCompleted:
		valid = (next == StageRevenueMonitoring) // Re-enter monitoring for subsequent lane/quote cycles
	case StageRevenueEscalated:
		valid = (next == StageRevenueRecommendation || next == StageRevenueMonitoring || next == StageRevenueCompleted)
	default:
		valid = false
	}

	if !valid {
		return ErrInvalidRevenueStageTransition
	}
	return nil
}

// AuthoritativeCommercialMetrics holds verified business facts from ledger and operational records
type AuthoritativeCommercialMetrics struct {
	ActualQuotedPrice      float64 `json:"actual_quoted_price"`
	ActualCarrierCost      float64 `json:"actual_carrier_cost"`
	ActualOperationalCost  float64 `json:"actual_operational_cost"`
	ActualGrossMargin      float64 `json:"actual_gross_margin"`
	ActualMarginPercentage float64 `json:"actual_margin_percentage"`
	LaneCode               string  `json:"lane_code"`
	VolumeTEU              float64 `json:"volume_teu"`
	ConfirmedAt            time.Time `json:"confirmed_at"`
}

// CarrierEconomicsProfile evaluates carrier cost reliability and historical margin contribution
type CarrierEconomicsProfile struct {
	CarrierID                    int64    `json:"carrier_id"`
	CarrierName                  string   `json:"carrier_name"`
	CarrierBaseCost              float64  `json:"carrier_base_cost"`
	OnTimeReliabilityPct         float64  `json:"on_time_reliability_pct"`
	DelayFrequencyPct            float64  `json:"delay_frequency_pct"`
	ExceptionFrequencyPct        float64  `json:"exception_frequency_pct"`
	HistoricalProfitabilityScore float64  `json:"historical_profitability_score"` // 0.0 - 100.0
	RecommendedCarrierAction     string   `json:"recommended_carrier_action"`
	EvaluationNotes              []string `json:"evaluation_notes"`
}

// CustomerValueScorecard measures multi-dimensional customer commercial value
type CustomerValueScorecard struct {
	CustomerID             int64    `json:"customer_id"`
	CustomerName           string   `json:"customer_name"`
	AnnualVolumeTEU        float64  `json:"annual_volume_teu"`
	GrossRevenueYTD        float64  `json:"gross_revenue_ytd"`
	AverageMarginPct       float64  `json:"average_margin_pct"`
	PaymentDSO             int      `json:"payment_dso"`
	ChurnRiskTier          string   `json:"churn_risk_tier"` // LOW, MEDIUM, HIGH
	LifetimeValueScore     float64  `json:"lifetime_value_score"` // 0.0 - 100.0
	NegotiationFlexibility string   `json:"negotiation_flexibility"` // LOW, MODERATE, HIGH
	CommercialNotes        []string `json:"commercial_notes"`
}

// PricingRecommendation represents a structured, versioned recommendation produced by Python reasoning
type PricingRecommendation struct {
	Version                   int       `json:"version"`
	RecommendedPrice          float64   `json:"recommended_price"`
	PriceRangeMin             float64   `json:"price_range_min"`
	PriceRangeMax             float64   `json:"price_range_max"`
	ExpectedCarrierCost       float64   `json:"expected_carrier_cost"`
	ExpectedOperationalCost   float64   `json:"expected_operational_cost"`
	ExpectedMarginPct         float64   `json:"expected_margin_pct"`
	WinProbabilityPct         float64   `json:"win_probability_pct"`
	Confidence                float64   `json:"confidence"`
	Assumptions               []string  `json:"assumptions"`
	PricingConstraintsApplied []string  `json:"pricing_constraints_applied"`
	PolicyCompliant           bool      `json:"policy_compliant"`
	RequiresApproval          bool      `json:"requires_approval"`
	ReasonForRevision         string    `json:"reason_for_revision,omitempty"`
	GeneratedAt               time.Time `json:"generated_at"`
}

// MarginRiskSignal captures identified risks of margin erosion
type MarginRiskSignal struct {
	SignalID          string   `json:"signal_id"`
	RiskType          string   `json:"risk_type"` // e.g. "BELOW_MARGIN_THRESHOLD", "CARRIER_COST_SPIKE", "EXCEPTION_COST_EROSION"
	ThresholdPct      float64  `json:"threshold_pct"`
	ActualMarginPct   float64  `json:"actual_margin_pct"`
	ExpectedMarginPct float64  `json:"expected_margin_pct"`
	CostVariance      float64  `json:"cost_variance"`
	Severity          string   `json:"severity"` // INFO, WARNING, CRITICAL
	IdentifiedCauses  []string `json:"identified_causes"`
	DetectedAt        time.Time `json:"detected_at"`
}

// CommercialOptimizationOption represents governed commercial action options
type CommercialOptimizationOption struct {
	OptionID           string                 `json:"option_id"`
	Title              string                 `json:"title"`
	ActionType         string                 `json:"action_type"` // e.g. "quotations.apply_optimized_price", "pricing.revise_lane_tariff", "carrier.switch_preferred_carrier"
	Reason             string                 `json:"reason"`
	TargetEntity       string                 `json:"target_entity"`
	RecommendedPrice   float64                `json:"recommended_price"`
	ExpectedMarginPct  float64                `json:"expected_margin_pct"`
	WinProbability     float64                `json:"win_probability"`
	RiskLevel          string                 `json:"risk_level"` // LOW, MEDIUM, HIGH
	Confidence         float64                `json:"confidence"`
	RequiredCapability string                 `json:"required_capability"`
	RequiredAutonomy   string                 `json:"required_autonomy"` // LEVEL_1 to LEVEL_4
	RequiresApproval   bool                   `json:"requires_approval"`
	VerificationMethod string                 `json:"verification_method"`
	HandoffModule      string                 `json:"handoff_module,omitempty"` // e.g. "PHASE_7_3_QUOTE_TO_CASH", "PHASE_7_5_CUSTOMER_MANAGEMENT"
	Parameters         map[string]interface{} `json:"parameters,omitempty"`
}

// AutonomousRevenueWorkflow models durable enterprise revenue and margin optimization orchestration
type AutonomousRevenueWorkflow struct {
	WorkflowID             string                          `json:"workflow_id"`
	OrgID                  int64                           `json:"org_id"`
	EntityType             string                          `json:"entity_type"` // e.g. "RFQ", "QUOTATION", "LANE", "CUSTOMER"
	EntityID               string                          `json:"entity_id"`
	CurrentStage           RevenueLifecycleStage           `json:"current_stage"`
	WorkflowState          EnterpriseWorkflowState         `json:"workflow_state"`
	Metrics                *AuthoritativeCommercialMetrics `json:"metrics,omitempty"`
	CarrierEconomics       *CarrierEconomicsProfile        `json:"carrier_economics,omitempty"`
	CustomerValue          *CustomerValueScorecard         `json:"customer_value,omitempty"`
	RecommendationHistory  []PricingRecommendation         `json:"recommendation_history,omitempty"`
	CurrentRecommendation  *PricingRecommendation          `json:"current_recommendation,omitempty"`
	DetectedMarginRisks    []MarginRiskSignal              `json:"detected_margin_risks,omitempty"`
	AvailableOptions       []CommercialOptimizationOption  `json:"available_options,omitempty"`
	SelectedOption         *CommercialOptimizationOption   `json:"selected_option,omitempty"`
	ExecutionPlan          []EnterpriseWorkflowStep        `json:"execution_plan,omitempty"`
	PendingApprovalsCount  int                             `json:"pending_approvals_count"`
	ApprovalRequestID      *int64                          `json:"approval_request_id,omitempty"`
	VerificationResult     map[string]interface{}          `json:"verification_result,omitempty"`
	ReplanVersion          int                             `json:"replan_version"`
	RetryCount             int                             `json:"retry_count"`
	MaxRetries             int                             `json:"max_retries"`
	OutcomeDetails         map[string]interface{}          `json:"outcome_details,omitempty"`
	CorrelationID          string                          `json:"correlation_id"`
	CreatedAt              time.Time                       `json:"created_at"`
	UpdatedAt              time.Time                       `json:"updated_at"`
}

// RevenueLifecycleEvent models incoming commercial and operational economic triggers
type RevenueLifecycleEvent struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"` // e.g. "RFQ_PRICING_REQUESTED", "CARRIER_RATE_INCREASED", "CUSTOMER_DISCOUNT_REQUESTED", "MARGIN_RISK_TRIGGERED"
	OrgID         int64                  `json:"org_id"`
	EntityType    string                 `json:"entity_type"`
	EntityID      string                 `json:"entity_id"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Source        string                 `json:"source"`
	Payload       map[string]interface{} `json:"payload"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
}

// RevenueOutcomeFeedback records realized economics for workforce memory and pricing model refinement
type RevenueOutcomeFeedback struct {
	WorkflowID        string  `json:"workflow_id"`
	EntityType        string  `json:"entity_type"`
	EntityID          string  `json:"entity_id"`
	QuotedPrice       float64 `json:"quoted_price"`
	ActualRevenue     float64 `json:"actual_revenue"`
	ActualCost        float64 `json:"actual_cost"`
	RealizedMarginPct float64 `json:"realized_margin_pct"`
	DealWon           bool    `json:"deal_won"`
	NegotiationRounds int     `json:"negotiation_rounds"`
	LessonsLearned    string  `json:"lessons_learned,omitempty"`
}
