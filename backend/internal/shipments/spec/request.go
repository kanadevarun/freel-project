package spec

import "time"

// ListShipmentsRequest defines the input for listing shipments
type ListShipmentsRequest struct {
	OrgID     int64
	Page      int
	Limit     int
	Status    *string
	Search    *string
	Workspace bool
}

// CarrierEmailRequest is the raw payload structure parsed from incoming carrier emails
type CarrierEmailRequest struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	MessageID string `json:"message_id"`
}

// GetShipmentRequest defines the input for retrieving a single shipment
type GetShipmentRequest struct {
	ID    int64
	OrgID int64
}

// GetTrackingPositionsRequest defines the input for retrieving position history
type GetTrackingPositionsRequest struct {
	ID    int64
	OrgID int64
	Limit int
}

// CarrierUpdateRequest defines the input for manual carrier tracking event update
type CarrierUpdateRequest struct {
	ID          int64
	OrgID       int64
	EventID     string `json:"event_id"`
	Description string `json:"description"`
}

// ResolveExceptionRequest defines the input for resolving a shipment exception
type ResolveExceptionRequest struct {
	ID    int64
	OrgID int64
}

// InboundCarrierEmailRequest represents the inbound carrier email fields
type InboundCarrierEmailRequest struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	MessageID string `json:"message_id"`
	OrgID     int64  `json:"-"`
}

// InboundWebhookRequest represents a raw inbound carrier webhook request parsed from transport
type InboundWebhookRequest struct {
	Carrier       string
	IntegrationID int64
	OrgID         int64
	Body          []byte
	Headers       map[string]string
}

// GetShipmentInternalRequest is used by internal Python service checks
type GetShipmentInternalRequest struct {
	ID    int64
	OrgID int64
}

// UpdateMilestoneInternalRequest is used by internal Python tracking workers
type UpdateMilestoneInternalRequest struct {
	ID            int64      `json:"-"`
	MilestoneCode string     `json:"milestone_code"`
	ActualDate    time.Time  `json:"actual_date"`
	Location      *string    `json:"location"`
	Notes         *string    `json:"notes"`
	OrgID         *int64     `json:"org_id"`
}

// UpdateMilestoneRequest is used by manual operator UI updates
type UpdateMilestoneRequest struct {
	ID            int64      `json:"-"`
	OrgID         int64      `json:"-"`
	MilestoneCode string     `json:"milestone_code"`
	ActualDate    time.Time  `json:"actual_date"`
	Location      *string    `json:"location,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
}

// CreateExceptionInternalRequest is used by internal Python exception analyzers
type CreateExceptionInternalRequest struct {
	ID            int64      `json:"-"`
	ExceptionType string     `json:"exception_type"`
	Severity      string     `json:"severity"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	OrgID         *int64     `json:"org_id"`
	SourceEventID *string    `json:"source_event_id"`
}

// CallbackInternalRequest is the completion payload from the LLM parser callback
type CallbackInternalRequest struct {
	ShipmentID           int64  `json:"shipment_id"`
	OrgID                int64  `json:"org_id"`
	HasCriticalException bool   `json:"has_critical_exception"`
	AISummary            string `json:"ai_summary"`
	EventID              string `json:"event_id"`
}

// Exception mutation requests for Task 16.5
type CreateExceptionRequest struct {
	ShipmentID    int64   `json:"-"`
	OrgID         int64   `json:"-"`
	ExceptionType string  `json:"exception_type"`
	Severity      string  `json:"severity"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	SourceEventID *string `json:"source_event_id,omitempty"`
}

type UpdateExceptionRequest struct {
	ID         int64   `json:"-"`
	ShipmentID int64   `json:"-"`
	OrgID      int64   `json:"-"`
	Status     string  `json:"status"`
	Severity   string  `json:"severity"`
	Notes      *string `json:"notes,omitempty"`
}

type AcknowledgeExceptionRequest struct {
	ID         int64 `json:"-"`
	ShipmentID int64 `json:"-"`
	OrgID      int64 `json:"-"`
}

type ResolveShipmentExceptionRequest struct {
	ID              int64   `json:"-"`
	ShipmentID      int64   `json:"-"`
	OrgID           int64   `json:"-"`
	ResolutionNotes string  `json:"resolution_notes"`
	ResolvedBy      int64   `json:"-"`
}

type DismissExceptionRequest struct {
	ID         int64 `json:"-"`
	ShipmentID int64 `json:"-"`
	OrgID      int64 `json:"-"`
}

type EvaluateExceptionsRequest struct {
	ShipmentID int64 `json:"-"`
	OrgID      int64 `json:"-"`
}

// Document Requests for Task 16.7
type GetShipmentDocumentsRequest struct {
	ShipmentID int64  `json:"-"`
	OrgID      int64  `json:"-"`
	Category   string `json:"category,omitempty"`
	Status     string `json:"status,omitempty"`
}

type CreateShipmentDocumentRequest struct {
	ShipmentID      int64      `json:"-"`
	OrgID           int64      `json:"-"`
	DocType         string     `json:"doc_type"`
	DocumentName    string     `json:"document_name"`
	Category        string     `json:"category"`
	Description     string     `json:"description,omitempty"`
	FileName        string     `json:"file_name"`
	FileURL         string     `json:"file_url,omitempty"`
	FileSize        int64      `json:"file_size,omitempty"`
	MimeType        string     `json:"mime_type,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	DocumentDate    *time.Time `json:"document_date,omitempty"`
	ReferenceNumber string     `json:"reference_number,omitempty"`
	Source          string     `json:"source,omitempty"`
	SourceID        *int64     `json:"source_id,omitempty"`
}

type UpdateShipmentDocumentRequest struct {
	ID              int64      `json:"-"`
	ShipmentID      int64      `json:"-"`
	OrgID           int64      `json:"-"`
	DocumentName    *string    `json:"document_name,omitempty"`
	Category        *string    `json:"category,omitempty"`
	Description     *string    `json:"description,omitempty"`
	Status          *string    `json:"status,omitempty"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	DocumentDate    *time.Time `json:"document_date,omitempty"`
	ReferenceNumber *string    `json:"reference_number,omitempty"`
}

type ApproveShipmentDocumentRequest struct {
	ID         int64 `json:"-"`
	ShipmentID int64 `json:"-"`
	OrgID      int64 `json:"-"`
}

type RejectShipmentDocumentRequest struct {
	ID              int64  `json:"-"`
	ShipmentID      int64  `json:"-"`
	OrgID           int64  `json:"-"`
	RejectionReason string `json:"rejection_reason"`
}

type DeleteShipmentDocumentRequest struct {
	ID         int64 `json:"-"`
	ShipmentID int64 `json:"-"`
	OrgID      int64 `json:"-"`
}

// Financial Requests for Task 16.8
type GetShipmentChargesRequest struct {
	ShipmentID int64  `json:"-"`
	OrgID      int64  `json:"-"`
	Category   string `json:"category,omitempty"`
	ChargeType string `json:"charge_type,omitempty"`
}

type CreateShipmentChargeRequest struct {
	ShipmentID      int64      `json:"-"`
	OrgID           int64      `json:"-"`
	Category        string     `json:"category"`
	ChargeType      string     `json:"charge_type"`
	Description     string     `json:"description"`
	VendorName      string     `json:"vendor_name,omitempty"`
	EstimatedAmount float64    `json:"estimated_amount"`
	ActualAmount    float64    `json:"actual_amount"`
	Currency        string     `json:"currency,omitempty"`
	ReferenceNumber string     `json:"reference_number,omitempty"`
	ChargeDate      *time.Time `json:"charge_date,omitempty"`
	Status          string     `json:"status,omitempty"`
	Notes           string     `json:"notes,omitempty"`
}

type UpdateShipmentChargeRequest struct {
	ID              int64      `json:"-"`
	ShipmentID      int64      `json:"-"`
	OrgID           int64      `json:"-"`
	Category        *string    `json:"category,omitempty"`
	ChargeType      *string    `json:"charge_type,omitempty"`
	Description     *string    `json:"description,omitempty"`
	VendorName      *string    `json:"vendor_name,omitempty"`
	EstimatedAmount *float64   `json:"estimated_amount,omitempty"`
	ActualAmount    *float64   `json:"actual_amount,omitempty"`
	Currency        *string    `json:"currency,omitempty"`
	ReferenceNumber *string    `json:"reference_number,omitempty"`
	ChargeDate      *time.Time `json:"charge_date,omitempty"`
	Status          *string    `json:"status,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
}

type DeleteShipmentChargeRequest struct {
	ID         int64 `json:"-"`
	ShipmentID int64 `json:"-"`
	OrgID      int64 `json:"-"`
}

type ReviewShipmentFinancialsRequest struct {
	ShipmentID      int64  `json:"-"`
	OrgID           int64  `json:"-"`
	FinancialStatus string `json:"financial_status"`
	Notes           string `json:"notes,omitempty"`
}



