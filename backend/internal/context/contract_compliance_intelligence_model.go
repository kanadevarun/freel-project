package bcontext

import (
	"time"
)

// ── Contract Identity & Overview Models ───────────────────────────────────────

// ContractIdentitySummary provides commercial identification and lifecycle metrics.
type ContractIdentitySummary struct {
	ContractID           int64      `json:"contract_id"`
	ContractReference    string     `json:"contract_reference"`
	ContractName         string     `json:"contract_name"`
	ContractType         string     `json:"contract_type"` // e.g. CUSTOMER_AGREEMENT, CARRIER_SERVICE, ANNUAL_SERVICE
	PartyID              int64      `json:"party_id"`
	PartyName            string     `json:"party_name"`
	PartyType            string     `json:"party_type"` // CUSTOMER or CARRIER
	TransportMode        string     `json:"transport_mode"`
	Status               string     `json:"status"` // DRAFT, ACTIVE, EXPIRED, TERMINATED, ARCHIVED
	Currency             string     `json:"currency"`
	ContractValue        float64    `json:"contract_value"`
	EffectiveDate        *time.Time `json:"effective_date,omitempty"`
	ExpiryDate           *time.Time `json:"expiry_date,omitempty"`
	RenewalDate          *time.Time `json:"renewal_date,omitempty"`
	DaysUntilExpiry      int        `json:"days_until_expiry"`
	DaysExpired          int        `json:"days_expired"`
	Owner                string     `json:"owner"`
	DocumentCount        int        `json:"document_count"`
	CompletenessScore    int        `json:"completeness_score"` // 0 - 100% based on presence of essential fields & docs
	IsActive             bool       `json:"is_active"`
	IsExpired            bool       `json:"is_expired"`
	IsExpiringSoon       bool       `json:"is_expiring_soon"` // <= 30 days
	IsCriticalExpiry     bool       `json:"is_critical_expiry"` // <= 7 days
	HasMissingDocument   bool       `json:"has_missing_document"`
	HasMissingEffective  bool       `json:"has_missing_effective"`
	HasMissingExpiry     bool       `json:"has_missing_expiry"`
	HasConflictingDates  bool       `json:"has_conflicting_dates"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// ── Commercial Terms & Obligations Models ────────────────────────────────────

// ContractTermSummary represents a single structured contractual term or rate clause.
type ContractTermSummary struct {
	ID           int64   `json:"id"`
	TermCategory string  `json:"term_category"` // COMMERCIAL, OPERATIONAL, PAYMENT, LIABILITY, SLA
	TermKey      string  `json:"term_key"`
	TermTitle    string  `json:"term_title"`
	TermValue    string  `json:"term_value"`
	ValueType    string  `json:"value_type"`
	Currency     *string `json:"currency,omitempty"`
	IsCritical   bool    `json:"is_critical"`
}

// ContractCommercialTermsSummary groups terms and clauses.
type ContractCommercialTermsSummary struct {
	TotalTermsCount         int                   `json:"total_terms_count"`
	CommercialTermsCount    int                   `json:"commercial_terms_count"`
	OperationalTermsCount   int                   `json:"operational_terms_count"`
	PaymentTermsCount       int                   `json:"payment_terms_count"`
	LiabilityTermsCount     int                   `json:"liability_terms_count"`
	CriticalTermsCount      int                   `json:"critical_terms_count"`
	FreeTimeDemurrageTerms  []string              `json:"free_time_demurrage_terms,omitempty"`
	PaymentCreditTerms      []string              `json:"payment_credit_terms,omitempty"`
	ServiceLevelTerms       []string              `json:"service_level_terms,omitempty"`
	LiabilityTerms          []string              `json:"liability_terms,omitempty"`
	CancellationTerms       []string              `json:"cancellation_terms,omitempty"`
	Terms                   []ContractTermSummary `json:"terms,omitempty"`
}

// ContractObligationSummary represents a contractual commitment or operational requirement.
type ContractObligationSummary struct {
	ID                  int64      `json:"id"`
	ObligationReference string     `json:"obligation_reference"`
	Title               string     `json:"title"`
	ObligationType      string     `json:"obligation_type"`
	ResponsibleParty    string     `json:"responsible_party"`
	Owner               string     `json:"owner,omitempty"`
	Priority            string     `json:"priority"`
	Status              string     `json:"status"` // ACTIVE, FULFILLED, WAIVED, BREACHED
	DueDate             *time.Time `json:"due_date,omitempty"`
	IsOverdue           bool       `json:"is_overdue"`
	IsRecurring         bool       `json:"is_recurring"`
}

// ContractObligationsSummary aggregates obligations and SLAs.
type ContractObligationsSummary struct {
	TotalObligations     int                         `json:"total_obligations"`
	ActiveObligations    int                         `json:"active_obligations"`
	FulfilledObligations int                         `json:"fulfilled_obligations"`
	OverdueObligations   int                         `json:"overdue_obligations"`
	BreachedObligations  int                         `json:"breached_obligations"`
	HasUnownedObligation bool                        `json:"has_unowned_obligation"`
	Obligations          []ContractObligationSummary `json:"obligations,omitempty"`
}

// ── Compliance & Regulatory Requirements Models ──────────────────────────────

// ComplianceRequirementSummary represents a regulatory mandate, certificate, or insurance item.
type ComplianceRequirementSummary struct {
	ID               int64      `json:"id"`
	RequirementType  string     `json:"requirement_type"` // INSURANCE, PERMIT, LICENSE, CUSTOMS, SECURITY
	Title            string     `json:"title"`
	ResponsibleParty string     `json:"responsible_party"`
	Status           string     `json:"status"` // VERIFIED, PENDING, EXPIRED, WAIVED
	RiskSeverity     string     `json:"risk_severity"` // LOW, MEDIUM, HIGH, CRITICAL
	ValidFrom        *time.Time `json:"valid_from,omitempty"`
	ValidUntil       *time.Time `json:"valid_until,omitempty"`
	IsExpired        bool       `json:"is_expired"`
	IsExpiringSoon   bool       `json:"is_expiring_soon"` // <= 30 days
	EvidenceDocID    *string    `json:"evidence_doc_id,omitempty"`
}

// ComplianceEventSummary represents a recorded compliance risk or audit finding.
type ComplianceEventSummary struct {
	ID          int64     `json:"id"`
	EventType   string    `json:"event_type"`
	Severity    string    `json:"severity"` // LOW, MEDIUM, HIGH, CRITICAL
	Status      string    `json:"status"`   // OPEN, UNDER_REVIEW, RESOLVED
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	DetectedAt  time.Time `json:"detected_at"`
	IsResolved  bool      `json:"is_resolved"`
}

// ContractComplianceSummary aggregates compliance requirements and audit events.
type ContractComplianceSummary struct {
	TotalRequirements        int                            `json:"total_requirements"`
	VerifiedRequirements     int                            `json:"verified_requirements"`
	PendingRequirements      int                            `json:"pending_requirements"`
	ExpiredRequirements      int                            `json:"expired_requirements"`
	ExpiringSoonRequirements int                            `json:"expiring_soon_requirements"`
	MissingEvidenceCount     int                            `json:"missing_evidence_count"`
	OpenEventsCount          int                            `json:"open_events_count"`
	HighSeverityEventsCount  int                            `json:"high_severity_events_count"`
	ComplianceStatus         string                         `json:"compliance_status"` // COMPLIANT, PENDING_VERIFICATION, ACTION_REQUIRED, NON_COMPLIANT, UNKNOWN
	Requirements             []ComplianceRequirementSummary `json:"requirements,omitempty"`
	Events                   []ComplianceEventSummary       `json:"events,omitempty"`
}

// ── Coverage Scope & Risk Models ─────────────────────────────────────────────

// ContractCoverageScope tracks connected parties, links, and operational utilization.
type ContractCoverageScope struct {
	CoveredParties   []string `json:"covered_parties,omitempty"`
	CoveredModes     []string `json:"covered_modes,omitempty"`
	LinkedQuotations int      `json:"linked_quotations"`
	LinkedShipments  int      `json:"linked_shipments"`
	LinkedInvoices   int      `json:"linked_invoices"`
}

// ContractRiskIndicators captures multi-factor risk scores and flags.
type ContractRiskIndicators struct {
	OverallRiskRating           string   `json:"overall_risk_rating"` // LOW, MODERATE, ELEVATED, CRITICAL
	OverallRiskScore            int      `json:"overall_risk_score"`  // 0 - 100
	IsExpired                   bool     `json:"is_expired"`
	IsExpiringSoon              bool     `json:"is_expiring_soon"`
	HasMissingDocument          bool     `json:"has_missing_document"`
	HasMissingEffective         bool     `json:"has_missing_effective"`
	HasMissingExpiry            bool     `json:"has_missing_expiry"`
	HasConflictingDates         bool     `json:"has_conflicting_dates"`
	HasUnresolvedComplianceRisk bool     `json:"has_unresolved_compliance_risk"`
	HasOverdueObligation        bool     `json:"has_overdue_obligation"`
	RiskFactors                 []string `json:"risk_factors"`
}

// ContractAIComplianceSummary contains deterministic rule-based AI analysis citing real records.
type ContractAIComplianceSummary struct {
	ExecutiveSummary     string   `json:"executive_summary"`
	CoverageAssessment   string   `json:"coverage_assessment"`
	ComplianceAssessment string   `json:"compliance_assessment"`
	SuggestedReviewAreas []string `json:"suggested_review_areas"`
	Confidence           string   `json:"confidence"` // HIGH, MEDIUM, LOW
	Citations            []string `json:"citations"`
	Limitations          []string `json:"limitations"`
}

// Contract360ComplianceIntelligence is the primary read-only intelligence model for a single contract.
type Contract360ComplianceIntelligence struct {
	ContractID     int64                          `json:"contract_id"`
	Identity       ContractIdentitySummary        `json:"identity"`
	Commercial     ContractCommercialTermsSummary `json:"commercial"`
	Obligations    ContractObligationsSummary     `json:"obligations"`
	Compliance     ContractComplianceSummary      `json:"compliance"`
	Coverage       ContractCoverageScope          `json:"coverage"`
	RiskIndicators ContractRiskIndicators         `json:"risk_indicators"`
	AISummary      ContractAIComplianceSummary    `json:"ai_summary"`
	ReadOnly       bool                           `json:"read_only"`
	CorrelationID  string                         `json:"correlation_id"`
	DataFreshness  time.Time                      `json:"data_freshness"`
}

// ── Contract Coverage Evaluation (Cross-Module) ─────────────────────────────

// ContractCoverageEvaluation models whether a shipment, invoice, or quote is covered by a contract.
type ContractCoverageEvaluation struct {
	TargetEntityType       string                      `json:"target_entity_type"` // SHIPMENT, INVOICE, QUOTATION
	TargetEntityID         int64                       `json:"target_entity_id"`
	TargetReference        string                      `json:"target_reference"`
	CustomerOrCarrierID    int64                       `json:"customer_or_carrier_id"`
	CustomerOrCarrierName  string                      `json:"customer_or_carrier_name"`
	TargetMode             string                      `json:"target_mode,omitempty"`
	TargetCurrency         string                      `json:"target_currency,omitempty"`
	TargetDate             *time.Time                  `json:"target_date,omitempty"`
	HasApplicableContract  bool                        `json:"has_applicable_contract"`
	MatchedContractID      *int64                      `json:"matched_contract_id,omitempty"`
	MatchedContractRef     string                      `json:"matched_contract_ref,omitempty"`
	MatchedContractName    string                      `json:"matched_contract_name,omitempty"`
	CoverageStatus         string                      `json:"coverage_status"` // FULLY_COVERED, PARTIALLY_COVERED, OUTSIDE_COVERAGE, EXPIRED_COVERAGE, UNKNOWN
	PartyMatches           bool                        `json:"party_matches"`
	ModeMatches            bool                        `json:"mode_matches"`
	CurrencyMatches        bool                        `json:"currency_matches"`
	DateWithinValidity     bool                        `json:"date_within_validity"`
	CoverageGaps           []string                    `json:"coverage_gaps,omitempty"`
	AISummary              ContractAIComplianceSummary `json:"ai_summary"`
	ReadOnly               bool                        `json:"read_only"`
	CorrelationID          string                      `json:"correlation_id"`
	DataFreshness          time.Time                   `json:"data_freshness"`
}

// ── Organization-Wide Contract & Compliance Summary ──────────────────────────

// OrgContractComplianceSummary provides executive overview across all contracts in an org.
type OrgContractComplianceSummary struct {
	TotalContracts              int                        `json:"total_contracts"`
	ActiveContracts             int                        `json:"active_contracts"`
	DraftContracts              int                        `json:"draft_contracts"`
	ExpiringContracts30d        int                        `json:"expiring_contracts_30d"`
	CriticalExpiring7d          int                        `json:"critical_expiring_7d"`
	ExpiredContracts            int                        `json:"expired_contracts"`
	ContractsMissingDocuments   int                        `json:"contracts_missing_documents"`
	ContractsMissingExpiryDate  int                        `json:"contracts_missing_expiry_date"`
	TotalComplianceRequirements int                        `json:"total_compliance_requirements"`
	VerifiedRequirements        int                        `json:"verified_requirements"`
	PendingRequirements         int                        `json:"pending_requirements"`
	OverdueRequirements         int                        `json:"overdue_requirements"`
	OpenComplianceEventsCount   int                        `json:"open_compliance_events_count"`
	HighSeverityEventsCount     int                        `json:"high_severity_events_count"`
	OverallComplianceHealth     string                     `json:"overall_compliance_health"` // EXCELLENT, GOOD, MODERATE, HIGH_RISK, CRITICAL
	ExpiringContracts           []ContractIdentitySummary  `json:"expiring_contracts,omitempty"`
	HighRiskComplianceIssues    []ComplianceEventSummary   `json:"high_risk_compliance_issues,omitempty"`
	ReadOnly                    bool                       `json:"read_only"`
	CorrelationID               string                     `json:"correlation_id"`
	DataFreshness               time.Time                  `json:"data_freshness"`
}
