package contract_compliance_automation

import (
	"encoding/json"
	"time"
)

// Compliance review status constants
const (
	ComplianceStatusCompliant      = "COMPLIANT"
	ComplianceStatusReviewRequired = "REVIEW_REQUIRED"
	ComplianceStatusNonCompliant   = "NON_COMPLIANT"
	ComplianceStatusExpired        = "EXPIRED"
)

// Draft status constants
const (
	DraftStatusDraft           = "DRAFT"
	DraftStatusPendingApproval = "PENDING_APPROVAL"
	DraftStatusApproved        = "APPROVED"
	DraftStatusRejected        = "REJECTED"
	DraftStatusDispatched      = "DISPATCHED"
)

// ContractComplianceReview represents a persisted review analysis record in MariaDB
type ContractComplianceReview struct {
	ID                     int64           `db:"id" json:"id"`
	OrgID                  int64           `db:"org_id" json:"org_id"`
	ContractID             int64           `db:"contract_id" json:"contract_id"`
	DocumentID             *string         `db:"document_id" json:"document_id,omitempty"`
	RiskLevel              string          `db:"risk_level" json:"risk_level"` // LOW, MEDIUM, HIGH, CRITICAL
	RiskScore              float64         `db:"risk_score" json:"risk_score"` // 0.00 to 100.00
	ComplianceStatus       string          `db:"compliance_status" json:"compliance_status"`
	ExecutiveSummary       string          `db:"executive_summary" json:"executive_summary"`
	DeterministicSignals   json.RawMessage `db:"deterministic_signals" json:"deterministic_signals"`
	ExtractedClauses       json.RawMessage `db:"extracted_clauses" json:"extracted_clauses"`
	StructuredDiscrepancies json.RawMessage `db:"structured_discrepancies" json:"structured_discrepancies"`
	ComplianceObligations  json.RawMessage `db:"compliance_obligations" json:"compliance_obligations"`
	MissingInformation     json.RawMessage `db:"missing_information" json:"missing_information"`
	Recommendations        json.RawMessage `db:"recommendations" json:"recommendations"`
	Evidence               json.RawMessage `db:"evidence" json:"evidence"`
	ConfidenceScore        float64         `db:"confidence_score" json:"confidence_score"`
	CorrelationID          string          `db:"correlation_id" json:"correlation_id"`
	CreatedAt              time.Time       `db:"created_at" json:"created_at"`
}

// ContractComplianceDraft represents a persisted clarification or review draft with HITL approval lifecycle
type ContractComplianceDraft struct {
	ID               int64     `db:"id" json:"id"`
	OrgID            int64     `db:"org_id" json:"org_id"`
	ContractID       int64     `db:"contract_id" json:"contract_id"`
	DocumentID       *string   `db:"document_id" json:"document_id,omitempty"`
	DraftType        string    `db:"draft_type" json:"draft_type"` // MISSING_DOCUMENT_REQUEST, CLAUSE_CLARIFICATION, RENEWAL_NOTICE, INTERNAL_REVIEW_NOTE, COMPLIANCE_BREACH_ALERT
	Subject          string    `db:"subject" json:"subject"`
	MessageBody      string    `db:"message_body" json:"message_body"`
	InternalNotes    *string   `db:"internal_notes" json:"internal_notes,omitempty"`
	RecipientName    string    `db:"recipient_name" json:"recipient_name"`
	RecipientEmail   string    `db:"recipient_email" json:"recipient_email"`
	Status           string    `db:"status" json:"status"` // DRAFT, PENDING_APPROVAL, APPROVED, REJECTED, DISPATCHED
	RequiresApproval bool      `db:"requires_approval" json:"requires_approval"`
	ApprovalID       *int64    `db:"approval_id" json:"approval_id,omitempty"`
	ActionProposalID *string   `db:"action_proposal_id" json:"action_proposal_id,omitempty"`
	CreatedByUserID  *int64    `db:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CorrelationID    string    `db:"correlation_id" json:"correlation_id"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// DeterministicComplianceSignalsDTO captures authoritative facts calculated in Go
type DeterministicComplianceSignalsDTO struct {
	IsExpired                       bool     `json:"is_expired"`
	IsNearingExpiry                 bool     `json:"is_nearing_expiry"`
	DaysUntilExpiry                 int      `json:"days_until_expiry"`
	MissingEffectiveDate            bool     `json:"missing_effective_date"`
	MissingExpiryDate               bool     `json:"missing_expiry_date"`
	MissingRequiredDocuments        bool     `json:"missing_required_documents"`
	RequiredDocumentsMissing        []string `json:"required_documents_missing"`
	HasDocumentAwaitingReview       bool     `json:"has_document_awaiting_review"`
	HasRejectedDocument             bool     `json:"has_rejected_document"`
	HasVersionConflict              bool     `json:"has_version_conflict"`
	HasIncompleteMetadata           bool     `json:"has_incomplete_metadata"`
	IsLinkedToInactiveAgreement     bool     `json:"is_linked_to_inactive_agreement"`
	MissingRateInformation          bool     `json:"missing_rate_information"`
	MissingServiceLevelTerms        bool     `json:"missing_service_level_terms"`
	MissingInsuranceTerms           bool     `json:"missing_insurance_terms"`
	HasStructuredTermDiscrepancy    bool     `json:"has_structured_term_discrepancy"`
	RequiresManualComplianceReview   bool     `json:"requires_manual_compliance_review"`
}

// ContractOverviewDTO aggregates contract header, deterministic signals, and recent reviews
type ContractOverviewDTO struct {
	ContractID            int64                              `json:"contract_id"`
	OrgID                 int64                              `json:"org_id"`
	ContractReference     string                             `json:"contract_reference"`
	ContractName          string                             `json:"contract_name"`
	ContractType          string                             `json:"contract_type"`
	PartyName             string                             `json:"party_name"`
	Status                string                             `json:"status"`
	EffectiveDate         *string                            `json:"effective_date,omitempty"`
	ExpiryDate            *string                            `json:"expiry_date,omitempty"`
	ContractValue         float64                            `json:"contract_value"`
	Currency              string                             `json:"currency"`
	TransportMode         string                             `json:"transport_mode"`
	Owner                 string                             `json:"owner"`
	DeterministicSignals  DeterministicComplianceSignalsDTO `json:"deterministic_signals"`
	LatestReview          *ContractComplianceReview          `json:"latest_review,omitempty"`
	ExistingDrafts        []*ContractComplianceDraft         `json:"existing_drafts"`
	LinkedDocumentsCount  int                                `json:"linked_documents_count"`
	StructuredTermsCount  int                                `json:"structured_terms_count"`
}

// GenerateClarificationDraftInput payload from frontend
type GenerateClarificationDraftInput struct {
	DraftType          string `json:"draft_type"`
	Tone               string `json:"tone"`
	RecipientName      string `json:"recipient_name,omitempty"`
	RecipientEmail     string `json:"recipient_email,omitempty"`
	CustomInstructions string `json:"custom_instructions,omitempty"`
	TargetClauseID     string `json:"target_clause_id,omitempty"`
}

// UpdateClarificationDraftInput payload for editing draft
type UpdateClarificationDraftInput struct {
	Subject        string  `json:"subject"`
	MessageBody    string  `json:"message_body"`
	RecipientName  string  `json:"recipient_name"`
	RecipientEmail string  `json:"recipient_email"`
	InternalNotes  *string `json:"internal_notes,omitempty"`
}

// SubmitDraftApprovalInput payload for submitting into approval center
type SubmitDraftApprovalInput struct {
	Notes string `json:"notes,omitempty"`
}
