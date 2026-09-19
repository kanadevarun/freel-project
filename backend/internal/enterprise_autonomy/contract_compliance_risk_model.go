package enterprise_autonomy

import (
	"errors"
	"time"
)

// RiskGovernanceStage represents governed stages for enterprise risk management
type RiskGovernanceStage string

const (
	StageRiskMonitoring             RiskGovernanceStage = "MONITORING"
	StageRiskDetection              RiskGovernanceStage = "RISK_DETECTION"
	StageRiskEvidenceCollection     RiskGovernanceStage = "EVIDENCE_COLLECTION"
	StageRiskMultiAgentAssessment   RiskGovernanceStage = "MULTI_AGENT_ASSESSMENT"
	StageRiskImpactAnalysis         RiskGovernanceStage = "IMPACT_ANALYSIS"
	StageRiskMitigationPlanning     RiskGovernanceStage = "MITIGATION_PLANNING"
	StageRiskWaitingApproval        RiskGovernanceStage = "WAITING_FOR_APPROVAL"
	StageRiskExecutingMitigation    RiskGovernanceStage = "EXECUTING_MITIGATION"
	StageRiskVerifying              RiskGovernanceStage = "VERIFYING"
	StageRiskActiveMonitoring       RiskGovernanceStage = "RISK_MONITORING"
	StageRiskResolved               RiskGovernanceStage = "RESOLVED"
	StageRiskEscalated              RiskGovernanceStage = "ESCALATED"
)

var (
	ErrInvalidRiskStageTransition      = errors.New("invalid risk governance lifecycle stage transition")
	ErrRiskWorkflowNotFound            = errors.New("autonomous contract, compliance or risk workflow not found")
	ErrCriticalRiskApprovalRequired    = errors.New("critical risk mitigation requires human executive approval by autonomy policy")
	ErrContractPolicyConflict          = errors.New("proposed commercial action conflicts with active customer contract or agreed rate card")
	ErrCompliancePolicyViolation       = errors.New("proposed operation violates statutory compliance controls or documentation readiness rules")
	ErrMitigationVerificationFailed    = errors.New("risk mitigation action could not be verified against authoritative records")
	ErrDuplicateMitigationAction       = errors.New("duplicate mitigation action dispatch rejected by idempotency filter")
)

// ValidateRiskGovernanceStageTransition enforces deterministic progression, loopbacks, and escalation
func ValidateRiskGovernanceStageTransition(current, next RiskGovernanceStage) error {
	if current == next {
		return nil
	}

	valid := false
	switch current {
	case StageRiskMonitoring:
		valid = (next == StageRiskDetection || next == StageRiskEvidenceCollection || next == StageRiskEscalated || next == StageRiskResolved)
	case StageRiskDetection:
		valid = (next == StageRiskEvidenceCollection || next == StageRiskMultiAgentAssessment || next == StageRiskMonitoring || next == StageRiskEscalated)
	case StageRiskEvidenceCollection:
		valid = (next == StageRiskMultiAgentAssessment || next == StageRiskImpactAnalysis || next == StageRiskMonitoring || next == StageRiskEscalated)
	case StageRiskMultiAgentAssessment:
		valid = (next == StageRiskImpactAnalysis || next == StageRiskMitigationPlanning || next == StageRiskMonitoring || next == StageRiskEscalated)
	case StageRiskImpactAnalysis:
		valid = (next == StageRiskMitigationPlanning || next == StageRiskEvidenceCollection || next == StageRiskMonitoring || next == StageRiskEscalated)
	case StageRiskMitigationPlanning:
		valid = (next == StageRiskWaitingApproval || next == StageRiskExecutingMitigation || next == StageRiskImpactAnalysis || next == StageRiskMonitoring || next == StageRiskEscalated)
	case StageRiskWaitingApproval:
		valid = (next == StageRiskExecutingMitigation || next == StageRiskMitigationPlanning || next == StageRiskEscalated)
	case StageRiskExecutingMitigation:
		valid = (next == StageRiskVerifying || next == StageRiskMitigationPlanning || next == StageRiskEscalated)
	case StageRiskVerifying:
		valid = (next == StageRiskActiveMonitoring || next == StageRiskResolved || next == StageRiskMitigationPlanning || next == StageRiskEscalated)
	case StageRiskActiveMonitoring:
		valid = (next == StageRiskResolved || next == StageRiskEvidenceCollection || next == StageRiskImpactAnalysis || next == StageRiskMonitoring || next == StageRiskEscalated)
	case StageRiskResolved:
		valid = (next == StageRiskMonitoring) // Return to continuous monitoring for new events
	case StageRiskEscalated:
		valid = (next == StageRiskMitigationPlanning || next == StageRiskMonitoring || next == StageRiskResolved)
	default:
		valid = false
	}

	if !valid {
		return ErrInvalidRiskStageTransition
	}
	return nil
}

// RiskSeverity defines normalized enterprise risk tiering
type RiskSeverity string

const (
	RiskSeverityLow      RiskSeverity = "LOW"
	RiskSeverityMedium   RiskSeverity = "MEDIUM"
	RiskSeverityHigh     RiskSeverity = "HIGH"
	RiskSeverityCritical RiskSeverity = "CRITICAL"
)

// RiskCategory categorizes domains across enterprise operations
type RiskCategory string

const (
	RiskCategoryContract    RiskCategory = "CONTRACT"
	RiskCategoryCompliance  RiskCategory = "COMPLIANCE"
	RiskCategoryOperational RiskCategory = "OPERATIONAL"
	RiskCategoryFinancial   RiskCategory = "FINANCIAL"
	RiskCategoryCustomer    RiskCategory = "CUSTOMER"
	RiskCategoryCrossDomain RiskCategory = "CROSS_DOMAIN"
)

// AuthoritativeRiskEntity holds authoritative business facts from database and ledger
type AuthoritativeRiskEntity struct {
	ContractID            string    `json:"contract_id,omitempty"`
	CustomerID            int64     `json:"customer_id,omitempty"`
	CustomerName          string    `json:"customer_name,omitempty"`
	ShipmentID            string    `json:"shipment_id,omitempty"`
	InvoiceID             string    `json:"invoice_id,omitempty"`
	AgreedSLAHours        int       `json:"agreed_sla_hours,omitempty"`
	ContractMarginFloor   float64   `json:"contract_margin_floor,omitempty"`
	ConfirmedExpiryDate   time.Time `json:"confirmed_expiry_date,omitempty"`
	HasCustomsFiling      bool      `json:"has_customs_filing"`
	VerifiedCarrierBonded bool      `json:"verified_carrier_bonded"`
	VerifiedAt            time.Time `json:"verified_at"`
}

// RiskEvidenceItem captures verified provenance distinguishing facts from inferences
type RiskEvidenceItem struct {
	EvidenceID            string       `json:"evidence_id"`
	SignalType            string       `json:"signal_type"` // e.g. "CONTRACT_EXPIRY_THRESHOLD", "CARRIER_TRANSIT_DELAY", "CUSTOMS_DOC_ABSENT"
	Category              RiskCategory `json:"category"`
	SourceEntity          string       `json:"source_entity"`
	SourceRecord          string       `json:"source_record"`
	SourceTimestamp       time.Time    `json:"source_timestamp"`
	AuthoritativeValue    string       `json:"authoritative_value"`
	DataFreshnessHours    float64      `json:"data_freshness_hours"`
	ReportingAgent        string       `json:"reporting_agent"`
	ConfidenceScore       float64      `json:"confidence_score"`
	IsAuthoritativeFact   bool         `json:"is_authoritative_fact"` // true for database facts, false for AI inferences
	ProvenanceDescription string       `json:"provenance_description"`
}

// ContractRiskAssessment captures contract terms and rate risks
type ContractRiskAssessment struct {
	ContractID               string   `json:"contract_id"`
	ExpiredContract          bool     `json:"expired_contract"`
	ExpiryWarningDays        int      `json:"expiry_warning_days"`
	RateMismatchDetected     bool     `json:"rate_mismatch_detected"`
	AgreedContractRate       float64  `json:"agreed_contract_rate"`
	QuotedOrInvoicedRate     float64  `json:"quoted_or_invoiced_rate"`
	UnauthorizedRateVariance float64  `json:"unauthorized_rate_variance"`
	SLABreachLikelihoodPct   float64  `json:"sla_breach_likelihood_pct"`
	CustomerTermsConflict    bool     `json:"customer_terms_conflict"`
	CarrierAgreementConflict bool     `json:"carrier_agreement_conflict"`
	AssessmentSummary        string   `json:"assessment_summary"`
	ContractClausesAtRisk    []string `json:"contract_clauses_at_risk"`
}

// ComplianceRiskAssessment captures trade, regulatory, and documentation controls
type ComplianceRiskAssessment struct {
	MissingDocumentation     bool     `json:"missing_documentation"`
	MissingDocumentsList     []string `json:"missing_documents_list"`
	RestrictedOperation      bool     `json:"restricted_operation"`
	DocumentationInconsistent bool    `json:"documentation_inconsistent"`
	SanctionsScreeningPassed  bool     `json:"sanctions_screening_passed"`
	CustomsReadinessScore    float64  `json:"customs_readiness_score"` // 0.0 - 100.0
	RegulatoryStandard       string   `json:"regulatory_standard"`     // e.g. "CBP_19CFR", "IMO_SOLAS", "FMC_TARIFF"
	ComplianceDefects        []string `json:"compliance_defects"`
	SeverityTier             string   `json:"severity_tier"`
}

// CrossDomainRiskCorrelation tracks cascading risks propagating across departments
type CrossDomainRiskCorrelation struct {
	CorrelationID           string       `json:"correlation_id"`
	PrimaryRiskCategory     RiskCategory `json:"primary_risk_category"`
	CorrelatedDomains       []string     `json:"correlated_domains"` // e.g. ["OPERATIONAL", "CONTRACT", "FINANCIAL", "CUSTOMER"]
	PropagationChain        string       `json:"propagation_chain"`  // e.g. "Port Congestion -> Delay -> SLA Breach -> Liquidated Damages -> Churn"
	TotalFinancialExposure  float64      `json:"total_financial_exposure"`
	SLAPenaltyExposure      float64      `json:"sla_penalty_exposure"`
	CustomerRetentionRiskPct float64     `json:"customer_retention_risk_pct"`
	CombinedRiskSeverity    RiskSeverity `json:"combined_risk_severity"`
	SynthesisNotes          []string     `json:"synthesis_notes"`
}

// MitigationOption represents a structured, governed mitigation proposal
type MitigationOption struct {
	OptionID            string                 `json:"option_id"`
	Title               string                 `json:"title"`
	ActionType          string                 `json:"action_type"` // e.g. "contracts.renew_terms", "compliance.request_origin_certificate", "shipments.expedite_drayage", "finance.freeze_disputed_charge"
	Category            RiskCategory           `json:"category"`
	Reason              string                 `json:"reason"`
	EvidenceRefs        []string               `json:"evidence_refs"`
	ProposedRemediation string                 `json:"proposed_remediation"`
	ExpectedOutcome     string                 `json:"expected_outcome"`
	ResidualRiskLevel   RiskSeverity           `json:"residual_risk_level"`
	Confidence          float64                `json:"confidence"`
	RequiredCapability  string                 `json:"required_capability"`
	RequiredAutonomy    string                 `json:"required_autonomy"` // LEVEL_1 to LEVEL_4
	RequiresApproval    bool                   `json:"requires_approval"`
	ApprovalReason      string                 `json:"approval_reason,omitempty"`
	HandoffModule       string                 `json:"handoff_module,omitempty"` // e.g. "PHASE_7_2_SHIPMENT", "PHASE_7_3_QUOTE_TO_CASH", "PHASE_7_4_EXCEPTION", "PHASE_7_5_CUSTOMER", "PHASE_7_6_REVENUE"
	Parameters          map[string]interface{} `json:"parameters,omitempty"`
}

// AutonomousRiskWorkflow models durable enterprise contract, compliance and cross-domain risk orchestration
type AutonomousRiskWorkflow struct {
	WorkflowID            string                      `json:"workflow_id"`
	OrgID                 int64                       `json:"org_id"`
	EntityType            string                      `json:"entity_type"` // "CONTRACT", "SHIPMENT", "QUOTATION", "CUSTOMER", "INVOICE"
	EntityID              string                      `json:"entity_id"`
	CurrentStage          RiskGovernanceStage         `json:"current_stage"`
	WorkflowState         EnterpriseWorkflowState     `json:"workflow_state"`
	OverallSeverity       RiskSeverity                `json:"overall_severity"`
	AuthoritativeEntity   *AuthoritativeRiskEntity    `json:"authoritative_entity,omitempty"`
	EvidenceRecords       []RiskEvidenceItem          `json:"evidence_records,omitempty"`
	ContractRisk          *ContractRiskAssessment     `json:"contract_risk,omitempty"`
	ComplianceRisk        *ComplianceRiskAssessment   `json:"compliance_risk,omitempty"`
	CrossDomainRisk       *CrossDomainRiskCorrelation `json:"cross_domain_risk,omitempty"`
	AvailableMitigations  []MitigationOption          `json:"available_mitigations,omitempty"`
	SelectedMitigation    *MitigationOption           `json:"selected_mitigation,omitempty"`
	ExecutionPlan         []EnterpriseWorkflowStep    `json:"execution_plan,omitempty"`
	PendingApprovalsCount int                         `json:"pending_approvals_count"`
	ApprovalRequestID     *int64                      `json:"approval_request_id,omitempty"`
	VerificationResult    map[string]interface{}      `json:"verification_result,omitempty"`
	ReplanVersion         int                         `json:"replan_version"`
	RetryCount            int                         `json:"retry_count"`
	MaxRetries            int                         `json:"max_retries"`
	OutcomeDetails        map[string]interface{}      `json:"outcome_details,omitempty"`
	CorrelationID         string                      `json:"correlation_id"`
	CreatedAt             time.Time                   `json:"created_at"`
	UpdatedAt             time.Time                   `json:"updated_at"`
}

// RiskLifecycleEvent models incoming operational, commercial, or compliance risk triggers
type RiskLifecycleEvent struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"` // e.g. "CONTRACT_EXPIRING_SOON", "RATE_MISMATCH_DETECTED", "CUSTOMS_HOLD_DECLARED", "SLA_BREACH_ANTICIPATED", "UNAUTHORIZED_RATE_CHANGE"
	OrgID         int64                  `json:"org_id"`
	EntityType    string                 `json:"entity_type"`
	EntityID      string                 `json:"entity_id"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Source        string                 `json:"source"`
	Payload       map[string]interface{} `json:"payload"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
}

// RiskOutcomeFeedback records realization of risk for workforce memory and policy refinement
type RiskOutcomeFeedback struct {
	WorkflowID             string       `json:"workflow_id"`
	EntityType             string       `json:"entity_type"`
	EntityID               string       `json:"entity_id"`
	PredictedSeverity      RiskSeverity `json:"predicted_severity"`
	ActualSeverity         RiskSeverity `json:"actual_severity"`
	MitigationEffective    bool         `json:"mitigation_effective"`
	ResidualRiskResolved   bool         `json:"residual_risk_resolved"`
	FalsePositive          bool         `json:"false_positive"`
	FalseNegative          bool         `json:"false_negative"`
	FinancialLossPrevented float64      `json:"financial_loss_prevented"`
	ExecutionDurationSec   int          `json:"execution_duration_sec"`
	LessonsLearned         string       `json:"lessons_learned,omitempty"`
}
