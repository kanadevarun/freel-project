package invoices

import (
	"time"
)

// Invoice represents a persistent customer invoice in LogisticsHQ.
type Invoice struct {
	ID              int64     `db:"id" json:"id"`
	OrgID           int64     `db:"org_id" json:"org_id"`
	InvoiceNumber   string    `db:"invoice_number" json:"invoice_number"`
	CustomerID      int64     `db:"customer_id" json:"customer_id"`
	CustomerName    string    `db:"customer_name" json:"customer_name"`
	CustomerCountry string    `db:"customer_country" json:"customer_country"`
	ShipmentID      *int64    `db:"shipment_id" json:"shipment_id,omitempty"`
	ShipmentNumber  string    `db:"shipment_number" json:"shipment_number"`
	BookingID       *int64    `db:"booking_id" json:"booking_id,omitempty"`
	BookingNumber   string    `db:"booking_number" json:"booking_number"`
	QuotationID     *int64    `db:"quotation_id" json:"quotation_id,omitempty"`
	QuoteNumber     string    `db:"quote_number" json:"quote_number"`
	Route           string    `db:"route" json:"route"`
	Origin          string    `db:"origin" json:"origin"`
	Destination     string    `db:"destination" json:"destination"`
	InvoiceDate     time.Time `db:"invoice_date" json:"invoice_date"`
	DueDate         time.Time `db:"due_date" json:"due_date"`
	DaysLeft        string    `db:"days_left" json:"days_left"`
	Currency        string    `db:"currency" json:"currency"`
	Subtotal        float64   `db:"subtotal" json:"subtotal"`
	TaxAmount       float64   `db:"tax_amount" json:"tax_amount"`
	DiscountAmount  float64   `db:"discount_amount" json:"discount_amount"`
	TotalAmount     float64   `db:"total_amount" json:"total_amount"`
	PaidAmount      float64   `db:"paid_amount" json:"paid_amount"`
	BalanceDue      float64   `db:"balance_due" json:"balance_due"`
	Status          string    `db:"status" json:"status"` // Draft, Pending Approval, Issued, Partially Paid, Paid, Overdue, Cancelled
	Type            string    `db:"type" json:"type"`     // CUSTOMER_AR
	Bookmarked      bool      `db:"bookmarked" json:"bookmarked"`
	IsMyInvoice     bool      `db:"is_my_invoice" json:"is_my_invoice"`
	CreatorName     string    `db:"creator_name" json:"creator_name"`
	CreatedByID     *int64    `db:"created_by_id" json:"created_by_id,omitempty"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`

	// Relational details populated on GET by ID
	LineItems []InvoiceItem     `json:"line_items,omitempty"`
	Payments  []InvoicePayment  `json:"payments,omitempty"`
	Documents []InvoiceDocument `json:"documents,omitempty"`
	History   []InvoiceHistory  `json:"history,omitempty"`
}

// InvoiceItem represents a line item within a customer invoice.
type InvoiceItem struct {
	ID              int64     `db:"id" json:"id"`
	OrgID           int64     `db:"org_id" json:"org_id"`
	InvoiceID       int64     `db:"invoice_id" json:"invoice_id"`
	Description     string    `db:"description" json:"description"`
	ServiceCategory string    `db:"service_category" json:"service_category"`
	Quantity        float64   `db:"quantity" json:"quantity"`
	UnitPrice       float64   `db:"unit_price" json:"unit_price"`
	TotalAmount     float64   `db:"total_amount" json:"total_amount"`
	DisplayOrder    int       `db:"display_order" json:"display_order"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

// InvoicePayment represents a recorded payment transaction.
type InvoicePayment struct {
	ID            int64     `db:"id" json:"id"`
	OrgID         int64     `db:"org_id" json:"org_id"`
	InvoiceID     int64     `db:"invoice_id" json:"invoice_id"`
	PaymentRef    string    `db:"payment_ref" json:"payment_ref"`
	Amount        float64   `db:"amount" json:"amount"`
	PaymentMethod string    `db:"payment_method" json:"payment_method"`
	Status        string    `db:"status" json:"status"`
	PaymentDate   time.Time `db:"payment_date" json:"payment_date"`
	Notes         *string   `db:"notes" json:"notes,omitempty"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

// InvoiceDocument represents an attached document record.
type InvoiceDocument struct {
	ID           int64     `db:"id" json:"id"`
	OrgID        int64     `db:"org_id" json:"org_id"`
	InvoiceID    int64     `db:"invoice_id" json:"invoice_id"`
	DocumentName string    `db:"document_name" json:"document_name"`
	FileSize     string    `db:"file_size" json:"file_size"`
	FileType     string    `db:"file_type" json:"file_type"`
	S3Key        *string   `db:"s3_key" json:"s3_key,omitempty"`
	UploadedAt   time.Time `db:"uploaded_at" json:"uploaded_at"`
}

// InvoiceHistory represents an audit log entry for invoice lifecycle.
type InvoiceHistory struct {
	ID          int64     `db:"id" json:"id"`
	OrgID       int64     `db:"org_id" json:"org_id"`
	InvoiceID   int64     `db:"invoice_id" json:"invoice_id"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	UserName    string    `db:"user_name" json:"user_name"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// InvoiceKPIStats represents tenant-scoped invoice dashboard summary metrics.
type InvoiceKPIStats struct {
	TotalInvoices   KPICardMetric `json:"total_invoices"`
	Outstanding     KPICardMetric `json:"outstanding"`
	PaidThisMonth   KPICardMetric `json:"paid_this_month"`
	Overdue         KPICardMetric `json:"overdue"`
}

type KPICardMetric struct {
	Amount         string `json:"amount"`
	DisplayAmount  string `json:"display_amount"`
	Count          int    `json:"count"`
	Label          string `json:"label"`
	Trend          string `json:"trend"`
	TrendDirection string `json:"trend_direction"`
	TrendPeriod    string `json:"trend_period"`
}

// CreateInvoiceInput represents input for invoice creation.
type CreateInvoiceInput struct {
	CustomerID      int64              `json:"customer_id"`
	CustomerName    string             `json:"customer_name"`
	CustomerCountry string             `json:"customer_country"`
	ShipmentID      *int64             `json:"shipment_id"`
	ShipmentNumber  string             `json:"shipment_number"`
	BookingID       *int64             `json:"booking_id"`
	BookingNumber   string             `json:"booking_number"`
	QuotationID     *int64             `json:"quotation_id"`
	QuoteNumber     string             `json:"quote_number"`
	Route           string             `json:"route"`
	Origin          string             `json:"origin"`
	Destination     string             `json:"destination"`
	InvoiceDate     string             `json:"invoice_date"`
	DueDate         string             `json:"due_date"`
	Currency        string             `json:"currency"`
	Subtotal        float64            `json:"subtotal"`
	TaxAmount       float64            `json:"tax_amount"`
	DiscountAmount  float64            `json:"discount_amount"`
	Status          string             `json:"status"`
	LineItems       []CreateItemInput  `json:"line_items"`
}

type CreateItemInput struct {
	Description     string  `json:"description"`
	ServiceCategory string  `json:"service_category"`
	Quantity        float64 `json:"quantity"`
	UnitPrice       float64 `json:"unit_price"`
}

// RecordPaymentInput represents input for recording a customer payment.
type RecordPaymentInput struct {
	Amount        float64 `json:"amount"`
	PaymentDate   string  `json:"payment_date"`
	PaymentMethod string  `json:"payment_method"`
	PaymentRef    string  `json:"payment_ref"`
	Notes         string  `json:"notes"`
}

// ListInvoiceParams represents filter parameters for listing invoices.
type ListInvoiceParams struct {
	PrimaryTab  string // 'ALL' or 'MY'
	Status      string // 'All', 'Draft', 'Pending Approval', 'Issued', 'Partially Paid', 'Paid', 'Overdue', 'Cancelled'
	Search      string
	CustomerID  int64
	ShipmentID  int64
	BookingID   int64
	QuotationID int64
	Page        int
	PageSize    int
}
