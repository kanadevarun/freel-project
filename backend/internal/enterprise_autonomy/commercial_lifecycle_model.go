package enterprise_autonomy

import (
	"errors"
	"time"
)

// CommercialLifecycleStage defines the sequential & governed stages of the autonomous quote-to-cash lifecycle
type CommercialLifecycleStage string

const (
	StageCommercialIntake     CommercialLifecycleStage = "COMMERCIAL_INTAKE"
	StageRFQExtraction        CommercialLifecycleStage = "RFQ_EXTRACTION"
	StageQualification        CommercialLifecycleStage = "QUALIFICATION"
	StagePricingMargin        CommercialLifecycleStage = "PRICING_MARGIN_ANALYSIS"
	StageContractCompliance   CommercialLifecycleStage = "CONTRACT_COMPLIANCE_CHECK"
	StageQuotationPrep        CommercialLifecycleStage = "QUOTATION_PREPARATION"
	StageCustomerNegotiation  CommercialLifecycleStage = "CUSTOMER_NEGOTIATION"
	StageQuoteAccepted        CommercialLifecycleStage = "QUOTE_ACCEPTED"
	StageBookingHandoff       CommercialLifecycleStage = "BOOKING_HANDOFF"
	StageShipmentHandoff      CommercialLifecycleStage = "SHIPMENT_HANDOFF"
	StageInvoiceGeneration    CommercialLifecycleStage = "INVOICE_GENERATION"
	StageCollectionMonitoring CommercialLifecycleStage = "COLLECTION_MONITORING"
	StageCommercialOutcome    CommercialLifecycleStage = "OUTCOME_LEARNING"
	StageCommercialCompleted  CommercialLifecycleStage = "COMPLETED"
)

var (
	ErrInvalidCommercialStageTransition = errors.New("invalid commercial quote-to-cash lifecycle stage transition")
	ErrCommercialWorkflowNotFound       = errors.New("autonomous commercial quote-to-cash workflow not found")
	ErrMissingRFQRequirements           = errors.New("mandatory RFQ requirements missing: origin, destination, or commodity")
	ErrCommercialApprovalRequired       = errors.New("quotation or commercial change requires human approval by autonomy policy")
	ErrDuplicateCommercialAction        = errors.New("duplicate commercial action suppressed by idempotency protection")
	ErrContractConflictDetected         = errors.New("proposed commercial terms conflict with active customer contract")
	ErrComplianceConflictDetected       = errors.New("proposed commercial routing violates trade compliance regulations")
	ErrPromptInjectionDetected          = errors.New("untrusted instruction or prompt injection attempt detected in input")
)

// ValidateCommercialStageTransition enforces forward progression with governed loops for negotiation
func ValidateCommercialStageTransition(current, next CommercialLifecycleStage) error {
	if current == next {
		return nil
	}

	valid := false
	switch current {
	case StageCommercialIntake:
		valid = (next == StageRFQExtraction || next == StageQualification)
	case StageRFQExtraction:
		valid = (next == StageQualification || next == StagePricingMargin)
	case StageQualification:
		valid = (next == StagePricingMargin || next == StageContractCompliance || next == StageQuotationPrep || next == StageCommercialCompleted)
	case StagePricingMargin:
		valid = (next == StageContractCompliance || next == StageQuotationPrep)
	case StageContractCompliance:
		valid = (next == StageQuotationPrep || next == StageQualification)
	case StageQuotationPrep:
		valid = (next == StageCustomerNegotiation || next == StageQuoteAccepted || next == StageCommercialCompleted)
	case StageCustomerNegotiation:
		// Customer negotiation can loop back to Pricing or Quotation Preparation for revision, or advance to Accepted or Completed
		valid = (next == StagePricingMargin || next == StageQuotationPrep || next == StageQuoteAccepted || next == StageCommercialCompleted)
	case StageQuoteAccepted:
		valid = (next == StageBookingHandoff || next == StageShipmentHandoff)
	case StageBookingHandoff:
		valid = (next == StageShipmentHandoff || next == StageInvoiceGeneration)
	case StageShipmentHandoff:
		valid = (next == StageInvoiceGeneration || next == StageCollectionMonitoring || next == StageCommercialOutcome)
	case StageInvoiceGeneration:
		valid = (next == StageCollectionMonitoring || next == StageCommercialOutcome || next == StageCommercialCompleted)
	case StageCollectionMonitoring:
		valid = (next == StageCommercialOutcome || next == StageCommercialCompleted)
	case StageCommercialOutcome:
		valid = (next == StageCommercialCompleted)
	case StageCommercialCompleted:
		valid = false // Terminal state
	default:
		valid = false
	}

	if !valid {
		return ErrInvalidCommercialStageTransition
	}
	return nil
}

// ExtractedRFQRequirements captures parsed logistics criteria with confidence & missing field audit
type ExtractedRFQRequirements struct {
	OriginPort             string   `json:"origin_port"`
	DestinationPort        string   `json:"destination_port"`
	Commodity              string   `json:"commodity"`
	EquipmentType          string   `json:"equipment_type"`
	WeightKg               float64  `json:"weight_kg"`
	VolumeCbm              float64  `json:"volume_cbm"`
	TargetShipDate         string   `json:"target_ship_date,omitempty"`
	Incoterms              string   `json:"incoterms,omitempty"`
	HazardousClass         string   `json:"hazardous_class,omitempty"`
	TemperatureControl     string   `json:"temperature_control,omitempty"`
	MissingMandatoryFields []string `json:"missing_mandatory_fields"`
	OperationalConstraints []string `json:"operational_constraints"`
	CanSafelyProceed       bool     `json:"can_safely_proceed"`
	Confidence             float64  `json:"confidence"`
}

// CustomerIntelligenceContext separates verifiable FACTS from PREDICTIONS and RECOMMENDATIONS
type CustomerIntelligenceContext struct {
	CustomerID            int64    `json:"customer_id"`
	CustomerName          string   `json:"customer_name"`
	// Authoritative Facts
	AuthoritativeTier     string   `json:"authoritative_tier"`
	AuthoritativeStatus   string   `json:"authoritative_status"`
	TotalHistoricalShipments int   `json:"total_historical_shipments"`
	AveragePaymentDays    float64  `json:"average_payment_days"`
	ActiveContractsCount  int      `json:"active_contracts_count"`
	// Predictions
	PredictedLeadScore    float64  `json:"predicted_lead_score"`
	PredictedWinRate      float64  `json:"predicted_win_rate"`
	PredictedChurnRisk    string   `json:"predicted_churn_risk"`
	PredictedPaymentRisk  string   `json:"predicted_payment_risk"`
	// Recommendations
	RecommendedDiscountPct float64 `json:"recommended_discount_pct"`
	RecommendedSalesAction string  `json:"recommended_sales_action"`
	ReasoningSummary      string   `json:"reasoning_summary"`
}

// PricingMarginRecommendation provides transparent commercial decision support
type PricingMarginRecommendation struct {
	RecommendedPrice      float64                `json:"recommended_price"`
	BaseCarrierCost       float64                `json:"base_carrier_cost"`
	ExpectedMarginPct     float64                `json:"expected_margin_pct"`
	ExpectedGrossProfit   float64                `json:"expected_gross_profit"`
	WinProbability        float64                `json:"win_probability"`
	RiskLevel             string                 `json:"risk_level"`
	Confidence            float64                `json:"confidence"`
	Evidence              []string               `json:"evidence"`
	Assumptions           []string               `json:"assumptions"`
	AlternativeOptions    []map[string]interface{} `json:"alternative_options"`
	RequiresPolicyApproval bool                  `json:"requires_policy_approval"`
	ApprovalReason        string                 `json:"approval_reason,omitempty"`
}

// ContractComplianceAssessment details legal and regulatory checks
type ContractComplianceAssessment struct {
	ContractID            *int64   `json:"contract_id,omitempty"`
	ContractReference     string   `json:"contract_reference,omitempty"`
	ContractCompliant     bool     `json:"contract_compliant"`
	ComplianceCompliant   bool     `json:"compliance_compliant"`
	RateAgreedInContract  bool     `json:"rate_agreed_in_contract"`
	ContractedRateAmount  float64  `json:"contracted_rate_amount,omitempty"`
	FlaggedConflicts      []string `json:"flagged_conflicts"`
	RequiredDocumentation []string `json:"required_documentation"`
	RequiresLegalApproval bool     `json:"requires_legal_approval"`
}

// NegotiationOption models structured counter-offer choices
type NegotiationOption struct {
	OptionID           string  `json:"option_id"` // OPTION_A, OPTION_B, OPTION_C, OPTION_D
	Label              string  `json:"label"`     // Maintain Price, Controlled Discount, etc.
	ProposedPrice      float64 `json:"proposed_price"`
	ExpectedMarginPct  float64 `json:"expected_margin_pct"`
	WinProbability     float64 `json:"win_probability"`
	RiskLevel          string  `json:"risk_level"`
	Confidence         float64 `json:"confidence"`
	PolicyRequirement  string  `json:"policy_requirement"`
	ApprovalRequired   bool    `json:"approval_required"`
	Rationale          string  `json:"rationale"`
}

// AutonomousCommercialWorkflow is the durable entity tracking quote-to-cash AI orchestration
type AutonomousCommercialWorkflow struct {
	WorkflowID               string                         `json:"workflow_id"`
	OrgID                    int64                          `json:"org_id"`
	RFQID                    string                         `json:"rfq_id"`
	LeadID                   *int64                         `json:"lead_id,omitempty"`
	CustomerID               *int64                         `json:"customer_id,omitempty"`
	QuotationID              *int64                         `json:"quotation_id,omitempty"`
	BookingID                *int64                         `json:"booking_id,omitempty"`
	ShipmentID               *int64                         `json:"shipment_id,omitempty"`
	InvoiceID                *int64                         `json:"invoice_id,omitempty"`
	CurrentStage             CommercialLifecycleStage       `json:"current_stage"`
	WorkflowState            EnterpriseWorkflowState        `json:"workflow_state"`
	Objective                string                         `json:"objective"`
	CorrelationID            string                         `json:"correlation_id"`
	ParentWorkflowID         *string                        `json:"parent_workflow_id,omitempty"`
	ChildShipmentWorkflowID  *string                        `json:"child_shipment_workflow_id,omitempty"`
	ExtractedRequirements    *ExtractedRFQRequirements      `json:"extracted_requirements,omitempty"`
	CustomerIntelligence     *CustomerIntelligenceContext   `json:"customer_intelligence,omitempty"`
	PricingRecommendation    *PricingMarginRecommendation   `json:"pricing_recommendation,omitempty"`
	ContractCompliance       *ContractComplianceAssessment  `json:"contract_compliance,omitempty"`
	NegotiationOptions       []NegotiationOption            `json:"negotiation_options,omitempty"`
	SelectedNegotiationOption *NegotiationOption            `json:"selected_negotiation_option,omitempty"`
	AssignedSpecialists      []string                       `json:"assigned_specialists"`
	CurrentStepID            string                         `json:"current_step_id,omitempty"`
	Steps                    []EnterpriseWorkflowStep       `json:"steps,omitempty"`
	PendingApprovalsCount    int                            `json:"pending_approvals_count"`
	ApprovalRequestID        *int64                         `json:"approval_request_id,omitempty"`
	LastExecutedAction       string                         `json:"last_executed_action,omitempty"`
	LastActionVerified       bool                           `json:"last_action_verified"`
	ReplanVersion            int                            `json:"replan_version"`
	OutcomeRecorded          bool                           `json:"outcome_recorded"`
	OutcomeDetails           map[string]interface{}         `json:"outcome_details,omitempty"`
	CreatedAt                time.Time                      `json:"created_at"`
	UpdatedAt                time.Time                      `json:"updated_at"`
}

// CommercialLifecycleEvent models triggers for commercial workflow lifecycle progression
type CommercialLifecycleEvent struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"` // e.g. RFQ_CREATED, QUOTE_PREPARED, NEGOTIATION_RESPONSE, QUOTE_ACCEPTED, BOOKING_CREATED, SHIPMENT_DELIVERED, INVOICE_GENERATED, INVOICE_OVERDUE
	RFQID         string                 `json:"rfq_id,omitempty"`
	QuotationID   *int64                 `json:"quotation_id,omitempty"`
	CustomerID    *int64                 `json:"customer_id,omitempty"`
	BookingID     *int64                 `json:"booking_id,omitempty"`
	ShipmentID    *int64                 `json:"shipment_id,omitempty"`
	InvoiceID     *int64                 `json:"invoice_id,omitempty"`
	Payload       map[string]interface{} `json:"payload"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Source        string                 `json:"source,omitempty"`
}

// CommercialOutcomeFeedback contains authoritative business data to train workforce memory
type CommercialOutcomeFeedback struct {
	WorkflowID         string  `json:"workflow_id"`
	Won                bool    `json:"won"`
	ActualRevenue      float64 `json:"actual_revenue"`
	ActualCost         float64 `json:"actual_cost"`
	ActualMarginPct    float64 `json:"actual_margin_pct"`
	CustomerResponse   string  `json:"customer_response"`
	PaidOnTime         bool    `json:"paid_on_time"`
	DisputeRaised      bool    `json:"dispute_raised"`
	FeedbackNotes      string  `json:"feedback_notes,omitempty"`
}
