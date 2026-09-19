package collections_automation

import (
	"encoding/json"
	"time"
)

// ReceivablesAnalysis represents a persisted AI analysis run for an invoice
type ReceivablesAnalysis struct {
	ID                   int64           `db:"id" json:"id"`
	OrgID                int64           `db:"org_id" json:"org_id"`
	InvoiceID            int64           `db:"invoice_id" json:"invoice_id"`
	CustomerID           int64           `db:"customer_id" json:"customer_id"`
	RiskLevel            string          `db:"risk_level" json:"risk_level"` // LOW, MEDIUM, HIGH, CRITICAL
	RiskScore            float64         `db:"risk_score" json:"risk_score"`
	DaysOverdue          int             `db:"days_overdue" json:"days_overdue"`
	AgingBucket          string          `db:"aging_bucket" json:"aging_bucket"`
	OutstandingAmount    float64         `db:"outstanding_amount" json:"outstanding_amount"`
	Currency             string          `db:"currency" json:"currency"`
	ReceivablesSummary   string          `db:"receivables_summary" json:"receivables_summary"`
	DeterministicSignals json.RawMessage `db:"deterministic_signals" json:"deterministic_signals"`
	KeyRisks             json.RawMessage `db:"key_risks" json:"key_risks"`
	RecommendedNextSteps json.RawMessage `db:"recommended_next_steps" json:"recommended_next_steps"`
	Evidence             json.RawMessage `db:"evidence" json:"evidence"`
	ConfidenceScore      float64         `db:"confidence_score" json:"confidence_score"`
	CorrelationID        string          `db:"correlation_id" json:"correlation_id"`
	CreatedAt            time.Time       `db:"created_at" json:"created_at"`
}

// Draft status constants
const (
	DraftStatusDraft           = "DRAFT"
	DraftStatusPendingApproval = "PENDING_APPROVAL"
	DraftStatusApproved        = "APPROVED"
	DraftStatusRejected        = "REJECTED"
	DraftStatusSent            = "SENT"
)

// CollectionDraft represents a persisted collection message draft with HITL approval lifecycle
type CollectionDraft struct {
	ID                int64     `db:"id" json:"id"`
	OrgID             int64     `db:"org_id" json:"org_id"`
	InvoiceID         int64     `db:"invoice_id" json:"invoice_id"`
	CustomerID        int64     `db:"customer_id" json:"customer_id"`
	DraftType         string    `db:"draft_type" json:"draft_type"` // FIRST_REMINDER, OVERDUE_NOTICE, FINAL_DEMAND, PAYMENT_PLAN_OFFER, INTERNAL_ESCALATION
	Subject           string    `db:"subject" json:"subject"`
	MessageBody       string    `db:"message_body" json:"message_body"`
	InternalNotes     *string   `db:"internal_notes" json:"internal_notes,omitempty"`
	RecipientName     string    `db:"recipient_name" json:"recipient_name"`
	RecipientEmail    string    `db:"recipient_email" json:"recipient_email"`
	OutstandingAmount float64   `db:"outstanding_amount" json:"outstanding_amount"`
	Currency          string    `db:"currency" json:"currency"`
	Status            string    `db:"status" json:"status"` // DRAFT, PENDING_APPROVAL, APPROVED, REJECTED, SENT
	RequiresApproval  bool      `db:"requires_approval" json:"requires_approval"`
	ApprovalID        *int64    `db:"approval_id" json:"approval_id,omitempty"`
	ActionProposalID  *string   `db:"action_proposal_id" json:"action_proposal_id,omitempty"`
	CreatedByUserID   *int64    `db:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CorrelationID     string    `db:"correlation_id" json:"correlation_id"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}

// DeterministicFinanceSignalsDTO captures authoritative financial facts calculated in Go
type DeterministicFinanceSignalsDTO struct {
	IsOverdue                    bool    `json:"is_overdue"`
	DaysOverdue                  int     `json:"days_overdue"`
	AgingBucket                  string  `json:"aging_bucket"`
	IsApproachingDueDate         bool    `json:"is_approaching_due_date"`
	DaysUntilDue                 int     `json:"days_until_due"`
	IsHighValue                  bool    `json:"is_high_value"`
	IsPartiallyPaid              bool    `json:"is_partially_paid"`
	HasMultipleOverdue           bool    `json:"has_multiple_overdue"`
	CustomerOverdueCount         int     `json:"customer_overdue_count"`
	CustomerTotalOverdueBalance  float64 `json:"customer_total_overdue_balance"`
	IsCreditLimitExceeded        bool    `json:"is_credit_limit_exceeded"`
	IsDisputed                   bool    `json:"is_disputed"`
	HasLinkedException           bool    `json:"has_linked_exception"`
	RiskScore                    float64 `json:"risk_score"`
	RiskLevel                    string  `json:"risk_level"`
	RequiresApproval             bool    `json:"requires_approval"`
}

// FinanceCollectionsOverview represents the complete operational state for an invoice
type FinanceCollectionsOverview struct {
	InvoiceID                int64                           `json:"invoice_id"`
	OrgID                    int64                           `json:"org_id"`
	InvoiceNumber            string                          `json:"invoice_number"`
	CustomerID               int64                           `json:"customer_id"`
	CustomerName             string                          `json:"customer_name"`
	CustomerEmail            string                          `json:"customer_email"`
	InvoiceDate              string                          `json:"invoice_date"`
	DueDate                  string                          `json:"due_date"`
	Currency                 string                          `json:"currency"`
	TotalAmount              float64                         `json:"total_amount"`
	PaidAmount               float64                         `json:"paid_amount"`
	BalanceDue               float64                         `json:"balance_due"`
	Status                   string                          `json:"status"`
	ShipmentID               *int64                          `json:"shipment_id,omitempty"`
	ShipmentNumber           string                          `json:"shipment_number,omitempty"`
	BookingNumber            string                          `json:"booking_number,omitempty"`
	Signals                  DeterministicFinanceSignalsDTO  `json:"signals"`
	LatestAnalysis           *ReceivablesAnalysis            `json:"latest_analysis,omitempty"`
	LatestDraft              *CollectionDraft                `json:"latest_draft,omitempty"`
	PrioritizedItems         json.RawMessage                 `json:"prioritized_items,omitempty"`
	Recommendations          json.RawMessage                 `json:"recommendations,omitempty"`
	PendingApproval          bool                            `json:"pending_approval"`
	ApprovalID               *int64                          `json:"approval_id,omitempty"`
	CorrelationID            string                          `json:"correlation_id"`
}

// GenerateCollectionDraftInput represents user request to generate a collection draft
type GenerateCollectionDraftInput struct {
	DraftType        string `json:"draft_type"` // FIRST_REMINDER, OVERDUE_NOTICE, FINAL_DEMAND, PAYMENT_PLAN_OFFER, INTERNAL_ESCALATION
	Tone             string `json:"tone"`       // POLITE, ASSERTIVE, URGENT, FORMAL
	RecipientName    string `json:"recipient_name,omitempty"`
	RecipientEmail   string `json:"recipient_email,omitempty"`
	UserInstructions string `json:"user_instructions,omitempty"`
}

// UpdateCollectionDraftInput represents user modifications to an existing draft
type UpdateCollectionDraftInput struct {
	Subject        string `json:"subject"`
	MessageBody    string `json:"message_body"`
	InternalNotes  string `json:"internal_notes,omitempty"`
	RecipientName  string `json:"recipient_name,omitempty"`
	RecipientEmail string `json:"recipient_email,omitempty"`
}

// SubmitCollectionDraftApprovalInput provides operational reason when submitting for approval
type SubmitCollectionDraftApprovalInput struct {
	Reason string `json:"reason"`
}
