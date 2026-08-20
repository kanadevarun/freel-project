package finance

import (
	"encoding/json"
	"time"
)

type Invoice struct {
	ID            string          `json:"id" db:"id"`
	OrgID         int64           `json:"org_id" db:"org_id"`
	ShipmentID    int64           `json:"shipment_id" db:"shipment_id"`
	InvoiceNumber string          `json:"invoice_number" db:"invoice_number"`
	VendorName    string          `json:"vendor_name" db:"vendor_name"`
	VendorRef     string          `json:"vendor_ref" db:"vendor_ref"`
	S3Key         string          `json:"s3_key" db:"s3_key"`
	FileName      string          `json:"file_name" db:"file_name"`
	Currency      string          `json:"currency" db:"currency"`
	TotalAmount   float64         `json:"total_amount" db:"total_amount"`
	Status        string          `json:"status" db:"status"`
	AISummary     *string         `json:"ai_summary" db:"ai_summary"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

type InvoiceItem struct {
	ID          int64   `json:"id" db:"id"`
	OrgID       int64   `json:"org_id" db:"org_id"`
	InvoiceID   string  `json:"invoice_id" db:"invoice_id"`
	ChargeCode  string  `json:"charge_code" db:"charge_code"`
	Description string  `json:"description" db:"description"`
	Quantity    float64 `json:"quantity" db:"quantity"`
	UnitPrice   float64 `json:"unit_price" db:"unit_price"`
	TotalAmount float64 `json:"total_amount" db:"total_amount"`
	Currency    string  `json:"currency" db:"currency"`
}

type FinanceDiscrepancy struct {
	ID            int64      `json:"id" db:"id"`
	OrgID         int64      `json:"org_id" db:"org_id"`
	ShipmentID    int64      `json:"shipment_id" db:"shipment_id"`
	InvoiceID     string     `json:"invoice_id" db:"invoice_id"`
	ChargeCode    string     `json:"charge_code" db:"charge_code"`
	FieldName     string     `json:"field_name" db:"field_name"`
	ExpectedValue string     `json:"expected_value" db:"expected_value"`
	ActualValue   string     `json:"actual_value" db:"actual_value"`
	Source        string     `json:"source" db:"source"`
	Status        string     `json:"status" db:"status"`
	ResolvedBy    *int64     `json:"resolved_by" db:"resolved_by"`
	ResolvedAt    *time.Time `json:"resolved_at" db:"resolved_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// FinanceCallbackRequest is the payload sent by the Python FinanceAgent sidecar
// to POST /internal/finance/callback after completing invoice reconciliation.
type FinanceCallbackRequest struct {
	OrgID         int64                 `json:"org_id"`
	ShipmentID    int64                 `json:"shipment_id"`
	InvoiceID     string                `json:"invoice_id"`
	Status        string                `json:"status"`
	Items         []InvoiceItem         `json:"items"`
	Discrepancies []*FinanceDiscrepancy  `json:"discrepancies"`
	AISummary     string                `json:"ai_summary"`
}

// RawInvoiceItems is used to unmarshal JSON from the Python sidecar's item payload.
type RawInvoiceItems struct {
	Items json.RawMessage `json:"items"`
}
