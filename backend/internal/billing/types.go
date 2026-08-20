package billing

import (
	"time"
)

type CustomerInvoice struct {
	ID            string     `json:"id" db:"id"`
	OrgID         int64      `json:"org_id" db:"org_id"`
	ShipmentID    int64      `json:"shipment_id" db:"shipment_id"`
	InvoiceNumber string     `json:"invoice_number" db:"invoice_number"`
	Status        string     `json:"status" db:"status"` // DRAFT, APPROVED, SENT, PAID, OVERDUE
	DueDate       *time.Time `json:"due_date" db:"due_date"`
	Currency      string     `json:"currency" db:"currency"`
	TotalAmount   float64    `json:"total_amount" db:"total_amount"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type CustomerInvoiceItem struct {
	ID          int64     `json:"id" db:"id"`
	OrgID       int64     `json:"org_id" db:"org_id"`
	InvoiceID   string    `json:"invoice_id" db:"invoice_id"`
	ChargeCode  string    `json:"charge_code" db:"charge_code"`
	Description string    `json:"description" db:"description"`
	Quantity    float64   `json:"quantity" db:"quantity"`
	UnitPrice   float64   `json:"unit_price" db:"unit_price"`
	TotalAmount float64   `json:"total_amount" db:"total_amount"`
	Currency    string    `json:"currency" db:"currency"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type Profitability struct {
	ID                  int64     `json:"id" db:"id"`
	OrgID               int64     `json:"org_id" db:"org_id"`
	ShipmentID          int64     `json:"shipment_id" db:"shipment_id"`
	QuotedSellPrice     float64   `json:"quoted_sell_price" db:"quoted_sell_price"`
	QuotedBuyPrice      float64   `json:"quoted_buy_price" db:"quoted_buy_price"`
	ActualCarrierCost   float64   `json:"actual_carrier_cost" db:"actual_carrier_cost"`
	AdditionalCharges   float64   `json:"additional_charges" db:"additional_charges"`
	ActualRevenue       float64   `json:"actual_revenue" db:"actual_revenue"`
	ExpectedProfit      float64   `json:"expected_profit" db:"expected_profit"`
	ActualProfit        float64   `json:"actual_profit" db:"actual_profit"`
	ExpectedMarginPct   float64   `json:"expected_margin_pct" db:"expected_margin_pct"`
	ActualMarginPct     float64   `json:"actual_margin_pct" db:"actual_margin_pct"`
	Variance            float64   `json:"variance" db:"variance"`
	ProfitabilityStatus string    `json:"profitability_status" db:"profitability_status"` // PENDING, ON_TARGET, UNDER_TARGET, NEGATIVE_MARGIN
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

type ClosureCheckResult struct {
	RuleName    string `json:"rule_name"`
	Passed      bool   `json:"passed"`
	Description string `json:"description"`
}

type ShipmentClosureAudit struct {
	ShipmentID int64                `json:"shipment_id"`
	Ready      bool                 `json:"ready"`
	Checks     []ClosureCheckResult `json:"checks"`
}
