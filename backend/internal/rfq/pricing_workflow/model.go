package pricing_workflow

import (
	"encoding/json"
	"time"
)

// RFQRequirementsExtraction records normalized extraction facts persisted in MariaDB
type RFQRequirementsExtraction struct {
	ID                  int64           `json:"id" db:"id"`
	OrgID               int64           `json:"org_id" db:"org_id"`
	RFQID               int64           `json:"rfq_id" db:"rfq_id"`
	Status              string          `json:"status" db:"status"`
	ExtractedData       json.RawMessage `json:"extracted_data" db:"extracted_data"`
	MissingFields       json.RawMessage `json:"missing_fields" db:"missing_fields"`
	ClarificationNeeded bool            `json:"clarification_needed" db:"clarification_needed"`
	ConfidenceScore     float64         `json:"confidence_score" db:"confidence_score"`
	CorrelationID       string          `json:"correlation_id" db:"correlation_id"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
}

// QuotationDraft records an AI-prepared, operator-editable commercial proposal
type QuotationDraft struct {
	ID                 int64           `json:"id" db:"id"`
	OrgID              int64           `json:"org_id" db:"org_id"`
	RFQID              int64           `json:"rfq_id" db:"rfq_id"`
	QuotationID        *int64          `json:"quotation_id,omitempty" db:"quotation_id"`
	Status             string          `json:"status" db:"status"`
	Currency           string          `json:"currency" db:"currency"`
	BaseCost           float64         `json:"base_cost" db:"base_cost"`
	TotalCost          float64         `json:"total_cost" db:"total_cost"`
	BaseSell           float64         `json:"base_sell" db:"base_sell"`
	Surcharges         float64         `json:"surcharges" db:"surcharges"`
	Discounts          float64         `json:"discounts" db:"discounts"`
	TaxAmount          float64         `json:"tax_amount" db:"tax_amount"`
	TotalSellingPrice  float64         `json:"total_selling_price" db:"total_selling_price"`
	GrossMarginAmount  float64         `json:"gross_margin_amount" db:"gross_margin_amount"`
	GrossMarginPct     float64         `json:"gross_margin_pct" db:"gross_margin_pct"`
	MarginHealth       string          `json:"margin_health" db:"margin_health"`
	RateReferences     json.RawMessage `json:"rate_references,omitempty" db:"rate_references"`
	CostComponents     json.RawMessage `json:"cost_components,omitempty" db:"cost_components"`
	SellingComponents  json.RawMessage `json:"selling_components,omitempty" db:"selling_components"`
	TermsAndConditions string          `json:"terms_and_conditions" db:"terms_and_conditions"`
	InternalSummary    string          `json:"internal_summary" db:"internal_summary"`
	CustomerWording    string          `json:"customer_wording" db:"customer_wording"`
	PricingExplanation string          `json:"pricing_explanation" db:"pricing_explanation"`
	RecipientEmail     string          `json:"recipient_email" db:"recipient_email"`
	RecipientName      string          `json:"recipient_name" db:"recipient_name"`
	ValidityStart      *time.Time      `json:"validity_start,omitempty" db:"validity_start"`
	ValidityEnd        *time.Time      `json:"validity_end,omitempty" db:"validity_end"`
	RequiresApproval   bool            `json:"requires_approval" db:"requires_approval"`
	ApprovalID         *int64          `json:"approval_id,omitempty" db:"approval_id"`
	ActionProposalID   *string         `json:"action_proposal_id,omitempty" db:"action_proposal_id"`
	ExecutionID        *int64          `json:"execution_id,omitempty" db:"execution_id"`
	CreatedByUserID    *int64          `json:"created_by_user_id,omitempty" db:"created_by_user_id"`
	CorrelationID      string          `json:"correlation_id" db:"correlation_id"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at" db:"updated_at"`
}

// PricingPreviewRequest requests a deterministic pricing calculation
type PricingPreviewRequest struct {
	RateID          *int64   `json:"rate_id,omitempty"`
	TargetMarginPct *float64 `json:"target_margin_pct,omitempty"`
	DiscountAmount  *float64 `json:"discount_amount,omitempty"`
}

// PricingPreviewResponse returns the deterministic pricing calculation
type PricingPreviewResponse struct {
	Currency           string          `json:"currency"`
	BaseCost           float64         `json:"base_cost"`
	Surcharges         float64         `json:"surcharges"`
	TotalCost          float64         `json:"total_cost"`
	BaseSell           float64         `json:"base_sell"`
	Discounts          float64         `json:"discounts"`
	TaxAmount          float64         `json:"tax_amount"`
	TotalSellingPrice  float64         `json:"total_selling_price"`
	GrossProfit        float64         `json:"gross_profit"`
	GrossMarginPct     float64         `json:"gross_margin_pct"`
	MarginHealth       string          `json:"margin_health"`
	AppliedRateID      *int64          `json:"applied_rate_id,omitempty"`
	AppliedCarrierName string          `json:"applied_carrier_name"`
	RateIsExpired      bool            `json:"rate_is_expired"`
	CostBreakdown      json.RawMessage `json:"cost_breakdown,omitempty"`
	SellBreakdown      json.RawMessage `json:"sell_breakdown,omitempty"`
	CorrelationID      string          `json:"correlation_id"`
}

// GenerateDraftInput inputs for generating a quotation draft
type GenerateDraftInput struct {
	RateID           *int64   `json:"rate_id,omitempty"`
	BaseSellOverride *float64 `json:"base_sell_override,omitempty"`
	DiscountAmount   *float64 `json:"discount_amount,omitempty"`
	RecipientName    string   `json:"recipient_name"`
	RecipientEmail   string   `json:"recipient_email"`
	ValidityDays     int      `json:"validity_days"`
	UserPromptNotes  string   `json:"user_prompt_notes,omitempty"`
}

// UpdateDraftInput inputs for updating an existing quotation draft
type UpdateDraftInput struct {
	CustomerWording    string  `json:"customer_wording"`
	InternalSummary    string  `json:"internal_summary"`
	TermsAndConditions string  `json:"terms_and_conditions"`
	RecipientName      string  `json:"recipient_name"`
	RecipientEmail     string  `json:"recipient_email"`
	ValidityDays       *int    `json:"validity_days,omitempty"`
	BaseSell           *float64 `json:"base_sell,omitempty"`
	Discounts          *float64 `json:"discounts,omitempty"`
}

// SubmitApprovalInput inputs for submitting draft to HITL manager approval
type SubmitApprovalInput struct {
	Reason string `json:"reason"`
}

// WorkflowOverviewResponse response for the full RFQ Pricing Workflow overview
type WorkflowOverviewResponse struct {
	RFQID               int64                      `json:"rfq_id"`
	RFQNumber           string                     `json:"rfq_number"`
	CustomerName        string                     `json:"customer_name"`
	Origin              string                     `json:"origin"`
	Destination         string                     `json:"destination"`
	ShipmentMode        string                     `json:"shipment_mode"`
	Incoterms           string                     `json:"incoterms"`
	Extraction          *RFQRequirementsExtraction `json:"extraction,omitempty"`
	LatestDraft         *QuotationDraft            `json:"latest_draft,omitempty"`
	PricingPreview      *PricingPreviewResponse    `json:"pricing_preview,omitempty"`
	AIExplanation       json.RawMessage            `json:"ai_explanation,omitempty"`
	RiskAnalysis        json.RawMessage            `json:"risk_analysis,omitempty"`
	PendingApproval     bool                       `json:"pending_approval"`
	ApprovalID          *int64                     `json:"approval_id,omitempty"`
	AvailableRatesCount int                        `json:"available_rates_count"`
	CorrelationID       string                     `json:"correlation_id"`
}
