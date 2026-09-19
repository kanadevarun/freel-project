package bcontext

import (
	"time"
)

// ── Invoice Identity & Overview Models ────────────────────────────────────────

// InvoiceIdentitySummary provides core commercial identification for an invoice.
type InvoiceIdentitySummary struct {
	InvoiceID        int64      `json:"invoice_id"`
	InvoiceNumber    string     `json:"invoice_number"`
	CustomerID       int64      `json:"customer_id"`
	CustomerName     string     `json:"customer_name"`
	CustomerCountry  string     `json:"customer_country"`
	ShipmentID       *int64     `json:"shipment_id,omitempty"`
	ShipmentNumber   string     `json:"shipment_number,omitempty"`
	BookingID        *int64     `json:"booking_id,omitempty"`
	BookingNumber    string     `json:"booking_number,omitempty"`
	QuotationID      *int64     `json:"quotation_id,omitempty"`
	QuoteNumber      string     `json:"quote_number,omitempty"`
	Route            string     `json:"route,omitempty"`
	Origin           string     `json:"origin,omitempty"`
	Destination      string     `json:"destination,omitempty"`
	InvoiceStatus    string     `json:"invoice_status"`
	PaymentStatus    string     `json:"payment_status"`
	Currency         string     `json:"currency"`
	IssueDate        time.Time  `json:"issue_date"`
	DueDate          *time.Time `json:"due_date,omitempty"`
	PaymentDate      *time.Time `json:"payment_date,omitempty"`
	Subtotal         float64    `json:"subtotal"`
	TaxAmount        float64    `json:"tax_amount"`
	DiscountAmount   float64    `json:"discount_amount"`
	TotalAmount      float64    `json:"total_amount"`
	PaidAmount       float64    `json:"paid_amount"`
	BalanceDue       float64    `json:"balance_due"`
	OverdueAmount    float64    `json:"overdue_amount"`
	DaysUntilDue     int        `json:"days_until_due"`
	DaysOverdue      int        `json:"days_overdue"`
	InvoiceAgeDays   int        `json:"invoice_age_days"`
	AgingBucket      string     `json:"aging_bucket"` // NOT_DUE, 1-30_DAYS, 31-60_DAYS, 61-90_DAYS, OVER_90_DAYS
	IsOverdue        bool       `json:"is_overdue"`
	IsPaid           bool       `json:"is_paid"`
	IsPartiallyPaid  bool       `json:"is_partially_paid"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// InvoiceLineItemSummary represents an audited line item.
type InvoiceLineItemSummary struct {
	ID              int64   `json:"id"`
	Description     string  `json:"description"`
	ServiceCategory string  `json:"service_category"`
	Quantity        float64 `json:"quantity"`
	UnitPrice       float64 `json:"unit_price"`
	TotalAmount     float64 `json:"total_amount"`
}

// InvoicePaymentRecordSummary represents an audited payment transaction.
type InvoicePaymentRecordSummary struct {
	ID            int64     `json:"id"`
	PaymentRef    string    `json:"payment_ref"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	Status        string    `json:"status"`
	PaymentDate   time.Time `json:"payment_date"`
	Notes         string    `json:"notes,omitempty"`
}

// ── Receivables & Commercial Exposure Models ─────────────────────────────────

// CustomerReceivablesSummary captures the customer's overall AR posture.
type CustomerReceivablesSummary struct {
	CustomerID                 int64   `json:"customer_id"`
	CustomerName               string  `json:"customer_name"`
	TotalInvoicedAmount        float64 `json:"total_invoiced_amount"`
	TotalPaidAmount            float64 `json:"total_paid_amount"`
	TotalOutstandingAmount     float64 `json:"total_outstanding_amount"`
	TotalOverdueAmount         float64 `json:"total_overdue_amount"`
	OpenInvoicesCount          int     `json:"open_invoices_count"`
	OverdueInvoicesCount       int     `json:"overdue_invoices_count"`
	PaidInvoicesCount          int     `json:"paid_invoices_count"`
	PaymentCompletionRate      float64 `json:"payment_completion_rate"`
	AveragePaymentDelayDays    float64 `json:"average_payment_delay_days"`
	HasRepeatedLatePayments    bool    `json:"has_repeated_late_payments"`
	ExposureRating             string  `json:"exposure_rating"` // LOW, MODERATE, ELEVATED, SEVERE
}

// ── Revenue and Cost Visibility Models ───────────────────────────────────────

// RevenueCostVisibility tracks commercial profitability, margin, and cost parity.
type RevenueCostVisibility struct {
	InvoiceRevenue          float64  `json:"invoice_revenue"`
	QuotationAmount         *float64 `json:"quotation_amount,omitempty"`
	CommercialCostAmount    *float64 `json:"commercial_cost_amount,omitempty"`
	GrossMarginAmount       *float64 `json:"gross_margin_amount,omitempty"`
	GrossMarginPercentage   *float64 `json:"gross_margin_percentage,omitempty"`
	RevenueCostVariance     *float64 `json:"revenue_cost_variance,omitempty"`
	HasMissingCost          bool     `json:"has_missing_cost"`
	MissingCostReason       string   `json:"missing_cost_reason,omitempty"`
	HasCommercialDiscrepancy bool    `json:"has_commercial_discrepancy"`
	DiscrepancyReason       string   `json:"discrepancy_reason,omitempty"`
	Currency                string   `json:"currency"`
	IsUnlinkedInvoice       bool     `json:"is_unlinked_invoice"` // True if no linked shipment or booking
}

// ── Finance Risk Indicators ──────────────────────────────────────────────────

// FinanceRiskIndicators evaluates deterministic balance, date, and audit risks.
type FinanceRiskIndicators struct {
	OverallRiskRating         string   `json:"overall_risk_rating"` // LOW, MODERATE, HIGH, CRITICAL
	OverallRiskScore          int      `json:"overall_risk_score"`  // 0 - 100
	IsOverdue                 bool     `json:"is_overdue"`
	IsApproachingDueDate      bool     `json:"is_approaching_due_date"`
	HasLargeBalance           bool     `json:"has_large_balance"`
	IsPartiallyPaid           bool     `json:"is_partially_paid"`
	MissingDueDate            bool     `json:"missing_due_date"`
	MissingCustomer           bool     `json:"missing_customer"`
	MissingShipment           bool     `json:"missing_shipment"`
	CurrencyMismatch          bool     `json:"currency_mismatch"`
	LineItemsInconsistent     bool     `json:"line_items_inconsistent"`
	RevenueWithoutCost        bool     `json:"revenue_without_cost"`
	RepeatedCustomerLatePayer bool     `json:"repeated_customer_late_payer"`
	RiskFactors               []string `json:"risk_factors"`
}

// ── Grounded AI Operational Finance Summary ──────────────────────────────────

// InvoiceAIFinanceSummary provides evidence-grounded financial synthesis.
type InvoiceAIFinanceSummary struct {
	ExecutiveSummary           string   `json:"executive_summary"`
	OutstandingExposureExplanation string `json:"outstanding_exposure_explanation"`
	AgingPositionExplanation   string   `json:"aging_position_explanation"`
	RevenueCostObservations    string   `json:"revenue_cost_observations"`
	PaymentBehaviorObservations string  `json:"payment_behavior_observations"`
	ActionableAttentionItems   []string `json:"actionable_attention_items"`
	SuggestedFinanceInquiries  []string `json:"suggested_finance_inquiries"`
	Confidence                 string   `json:"confidence"` // HIGH, MEDIUM, LOW
	FreshnessTimestamp         time.Time `json:"freshness_timestamp"`
	Citations                  []string `json:"citations"`
}

// ── Complete Invoice 360 Finance Intelligence Aggregate ─────────────────────

// Invoice360FinanceIntelligence represents the holistic financial view.
type Invoice360FinanceIntelligence struct {
	InvoiceID            int64                         `json:"invoice_id"`
	Identity             InvoiceIdentitySummary        `json:"identity"`
	CustomerAR           CustomerReceivablesSummary    `json:"customer_receivables"`
	RevenueCost          RevenueCostVisibility         `json:"revenue_cost"`
	LineItems            []InvoiceLineItemSummary      `json:"line_items"`
	Payments             []InvoicePaymentRecordSummary `json:"payments"`
	RiskIndicators       FinanceRiskIndicators         `json:"risk_indicators"`
	AIFinanceSummary     InvoiceAIFinanceSummary       `json:"ai_summary"`
	AuditedLineItemTotal float64                       `json:"audited_line_item_total"`
	CalculatedAt         time.Time                     `json:"calculated_at"`
	CorrelationID        string                        `json:"correlation_id"`
	OrganizationScope    int64                         `json:"organization_scope"`
	ReadOnly             bool                          `json:"read_only"`
	Warnings             []string                      `json:"warnings,omitempty"`
}

// ── Organization-Level Operational Finance Summary ───────────────────────────

// AgingBucketBreakdown summarizes outstanding balances by aging period.
type AgingBucketBreakdown struct {
	NotDue     float64 `json:"not_due"`
	Days1To30  float64 `json:"days_1_30"`
	Days31To60 float64 `json:"days_31_60"`
	Days61To90 float64 `json:"days_61_90"`
	Over90Days float64 `json:"over_90_days"`
}

// CustomerExposureItem highlights high-exposure debtor accounts.
type CustomerExposureItem struct {
	CustomerID        int64   `json:"customer_id"`
	CustomerName      string  `json:"customer_name"`
	TotalOutstanding float64 `json:"total_outstanding"`
	TotalOverdue     float64 `json:"total_overdue"`
	OpenInvoicesCount int     `json:"open_invoices_count"`
	Currency          string  `json:"currency"`
}

// OrgFinanceSummary provides an authenticated organization AR and exposure summary.
type OrgFinanceSummary struct {
	ActiveInvoicesCount         int                    `json:"active_invoices_count"`
	OpenInvoicesCount           int                    `json:"open_invoices_count"`
	OverdueInvoicesCount        int                    `json:"overdue_invoices_count"`
	PaidInvoicesCount           int                    `json:"paid_invoices_count"`
	CancelledInvoicesCount      int                    `json:"cancelled_invoices_count"`
	InvoicesApproachingDueCount int                    `json:"invoices_approaching_due_count"`
	TotalInvoicedAmount         float64                `json:"total_invoiced_amount"`
	TotalPaidAmount             float64                `json:"total_paid_amount"`
	TotalOutstandingReceivables float64                `json:"total_outstanding_receivables"`
	TotalOverdueReceivables     float64                `json:"total_overdue_receivables"`
	AgingBreakdown              AgingBucketBreakdown   `json:"aging_breakdown"`
	TopCustomerExposures        []CustomerExposureItem `json:"top_customer_exposures"`
	PrimaryCurrency             string                 `json:"primary_currency"`
	ReceivablesHealthRating     string                 `json:"receivables_health_rating"` // EXCELLENT, GOOD, MODERATE, HIGH_RISK, CRITICAL
	CalculatedAt                time.Time              `json:"calculated_at"`
	OrganizationScope           int64                  `json:"organization_scope"`
	ReadOnly                    bool                   `json:"read_only"`
}
