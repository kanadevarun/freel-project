package spec

import "time"

type ListRFQsRequest struct {
	OrgID  int32 `json:"-"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

type GetRFQRequest struct {
	OrgID int32 `json:"-"`
	ID    int32 `json:"id"` // Path param
}

type GetTimelineRequest struct {
	OrgID int32 `json:"-"`
	ID    int32 `json:"id"`
}

type GetAgentStatusRequest struct {
	OrgID int32 `json:"-"`
	ID    int32 `json:"id"`
}

type CreateRFQRequest struct {
	OrgID       int32      `json:"-"`
	CustomerID  int32      `json:"customer_id"`
	Origin      *string    `json:"origin"`
	Destination *string    `json:"destination"`
	Incoterms   *string    `json:"incoterms"`
	TargetDate  *time.Time `json:"target_date"`
	Items       []RFQItem  `json:"items"`
	LeadID      *int64     `json:"lead_id"`
}

type UpdateStageRequest struct {
	OrgID int32  `json:"-"`
	ID    int32  `json:"id"`
	Stage string `json:"stage"`
}

type ParseShipmentRequest struct {
	OrgID    int32  `json:"-"`
	RawEmail string `json:"raw_email"`
	RawText  string `json:"raw_text"`
}

type AddQuoteRequest struct {
	OrgID int32  `json:"-"`
	ID    int32  `json:"id"`
	Quote Quote  `json:"quote"`
}

// GetCarrierRatesRequest asks for all available carrier rates for a given RFQ.
// The RFQ's origin, destination, and target_date are used to call the carrier provider.
type GetCarrierRatesRequest struct {
	OrgID int32 `json:"-"`
	ID    int32 `json:"id"` // RFQ ID — path param
}

// ApproveQuoteRequest selects a specific quote and advances the RFQ to QUOTE_SENT.
// This is triggered when the Pricing team clicks "Approve & Send" in the UI.
type ApproveQuoteRequest struct {
	OrgID   int32 `json:"-"`
	ID      int32 `json:"id"`       // RFQ ID — path param
	QuoteID int32 `json:"quote_id"` // Which quote to approve
}

// GetRequirementsRequest is the decoded request for the requirements evaluation endpoint.
// OrgID is injected from the authenticated user context (never from the client).
type GetRequirementsRequest struct {
	OrgID int32 `json:"-"`
	ID    int32 `json:"id"` // RFQ ID — path param
}

// GetActivityRequest is the decoded request for the activity timeline endpoint.
// OrgID is injected from the authenticated user context (never from the client).
type GetActivityRequest struct {
	OrgID int32 `json:"-"`
	ID    int32 `json:"id"` // RFQ ID — path param
}

// ──────────────────────────────────────────────────────────────────────────────
// Document Management Requests (Task 12)
// ──────────────────────────────────────────────────────────────────────────────

// GetDocumentsRequest requests all resolved document requirements and records for an RFQ.
type GetDocumentsRequest struct {
	OrgID int32 `json:"-"`
	ID    int32 `json:"id"` // RFQ ID — path param
}

// CreateDocumentRequest creates a new persistent document record for an RFQ.
type CreateDocumentRequest struct {
	OrgID        int32                  `json:"-"`
	RFQID        int32                  `json:"-"`
	DocumentType string                 `json:"document_type"`
	DocumentName string                 `json:"document_name"`
	Description  *string                `json:"description,omitempty"`
	FileName     *string                `json:"file_name,omitempty"`
	FileURL      *string                `json:"file_url,omitempty"`
	FileSize     *int64                 `json:"file_size,omitempty"`
	MimeType     *string                `json:"mime_type,omitempty"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateDocumentStatusRequest transitions a document's lifecycle status.
type UpdateDocumentStatusRequest struct {
	OrgID           int32   `json:"-"`
	RFQID           int32   `json:"-"`
	DocumentID      int64   `json:"-"`
	Status          string  `json:"status"`
	RejectionReason *string `json:"rejection_reason,omitempty"`
}

// DeleteDocumentRequest deletes a document record belonging to an RFQ.
type DeleteDocumentRequest struct {
	OrgID      int32 `json:"-"`
	RFQID      int32 `json:"-"`
	DocumentID int64 `json:"document_id"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Quote Management Requests (Task 13)
// ──────────────────────────────────────────────────────────────────────────────


// GetQuotesRequest requests all quotes and commercial comparison intelligence for an RFQ.
type GetQuotesRequest struct {
	OrgID int32 `json:"-"`
	ID    int32 `json:"id"` // RFQ ID — path param
}

// CreateQuoteRequest creates a new persistent carrier quote for an RFQ.
type CreateQuoteRequest struct {
	OrgID              int32         `json:"-"`
	RFQID              int32         `json:"-"`
	CarrierName        string        `json:"carrier_name"`
	CarrierID          *string       `json:"carrier_id,omitempty"`
	QuoteReference     *string       `json:"quote_reference,omitempty"`
	Currency           string        `json:"currency"`
	BuyPrice           float64       `json:"buy_price"`
	SellPrice          float64       `json:"sell_price"`
	OceanFreight       *float64      `json:"ocean_freight,omitempty"`
	OriginCharges      *float64      `json:"origin_charges,omitempty"`
	DestinationCharges *float64      `json:"destination_charges,omitempty"`
	TransitTimeDays    *int          `json:"transit_time_days,omitempty"`
	FreeDays           *int          `json:"free_days,omitempty"`
	ValidFrom          *time.Time    `json:"valid_from,omitempty"`
	ValidUntil         *time.Time    `json:"valid_until,omitempty"`
	ETD                *time.Time    `json:"etd,omitempty"`
	ETA                *time.Time    `json:"eta,omitempty"`
	Notes              *string       `json:"notes,omitempty"`
	Charges            []QuoteCharge `json:"charges,omitempty"`
}

// UpdateQuoteRequest updates an existing quote.
type UpdateQuoteRequest struct {
	OrgID              int32         `json:"-"`
	RFQID              int32         `json:"-"`
	QuoteID            int64         `json:"-"`
	CarrierName        *string       `json:"carrier_name,omitempty"`
	CarrierID          *string       `json:"carrier_id,omitempty"`
	QuoteReference     *string       `json:"quote_reference,omitempty"`
	Currency           *string       `json:"currency,omitempty"`
	BuyPrice           *float64      `json:"buy_price,omitempty"`
	SellPrice          *float64      `json:"sell_price,omitempty"`
	OceanFreight       *float64      `json:"ocean_freight,omitempty"`
	OriginCharges      *float64      `json:"origin_charges,omitempty"`
	DestinationCharges *float64      `json:"destination_charges,omitempty"`
	TransitTimeDays    *int          `json:"transit_time_days,omitempty"`
	FreeDays           *int          `json:"free_days,omitempty"`
	ValidFrom          *time.Time    `json:"valid_from,omitempty"`
	ValidUntil         *time.Time    `json:"valid_until,omitempty"`
	ETD                *time.Time    `json:"etd,omitempty"`
	ETA                *time.Time    `json:"eta,omitempty"`
	Notes              *string       `json:"notes,omitempty"`
	Status             *string       `json:"status,omitempty"`
	Charges            []QuoteCharge `json:"charges,omitempty"`
}

// UpdateQuoteStatusRequest updates the lifecycle review status of a quote.
type UpdateQuoteStatusRequest struct {
	OrgID   int32   `json:"-"`
	RFQID   int32   `json:"-"`
	QuoteID int64   `json:"-"`
	Status  string  `json:"status"`
	Notes   *string `json:"notes,omitempty"`
}

// RecommendQuoteRequest marks a quote as the recommended option.
type RecommendQuoteRequest struct {
	OrgID   int32 `json:"-"`
	RFQID   int32 `json:"-"`
	QuoteID int64 `json:"-"`
}

// ApproveRFQQuoteRequest approves a quote for internal operations.
type ApproveRFQQuoteRequest struct {
	OrgID   int32   `json:"-"`
	RFQID   int32   `json:"-"`
	QuoteID int64   `json:"-"`
	Notes   *string `json:"notes,omitempty"`
}

// SelectQuoteRequest selects an approved quote for the customer proposal.
type SelectQuoteRequest struct {
	OrgID   int32 `json:"-"`
	RFQID   int32 `json:"-"`
	QuoteID int64 `json:"-"`
}

// DeleteQuoteRequest deletes or withdraws a quote record.
type DeleteQuoteRequest struct {
	OrgID   int32 `json:"-"`
	RFQID   int32 `json:"-"`
	QuoteID int64 `json:"-"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Task 14: Booking & Shipment Handoff Requests
// ──────────────────────────────────────────────────────────────────────────────

type GetBookingHandoffRequest struct {
	OrgID int32 `json:"-"`
	ID    int32 `json:"id"`
}

type CreateBookingRequest struct {
	OrgID               int32      `json:"-"`
	RFQID               int32      `json:"-"`
	QuoteID             *int64     `json:"quote_id,omitempty"`
	BookingNumber       *string    `json:"booking_number,omitempty"`
	CarrierID           *string    `json:"carrier_id,omitempty"`
	CarrierName         string     `json:"carrier_name"`
	CarrierSCAC         *string    `json:"carrier_scac,omitempty"`
	OriginPort          string     `json:"origin_port"`
	DestinationPort     string     `json:"destination_port"`
	VesselName          *string    `json:"vessel_name,omitempty"`
	VoyageNumber        *string    `json:"voyage_number,omitempty"`
	ETD                 *time.Time `json:"etd,omitempty"`
	ETA                 *time.Time `json:"eta,omitempty"`
	CargoSummary        *string    `json:"cargo_summary,omitempty"`
	SpecialInstructions *string    `json:"special_instructions,omitempty"`
}

type UpdateBookingStatusRequest struct {
	OrgID     int32   `json:"-"`
	RFQID     int32   `json:"-"`
	BookingID int64   `json:"-"`
	Status    string  `json:"status"`
	Notes     *string `json:"notes,omitempty"`
}

type GetShipmentHandoffRequest struct {
	OrgID int32 `json:"-"`
	ID    int32 `json:"id"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Task 15: Dedicated Booking Operations Workspace Requests
// ──────────────────────────────────────────────────────────────────────────────

type BookingListFilter struct {
	OrgID           int32      `json:"-"`
	Status          *string    `json:"status,omitempty"`
	Carrier         *string    `json:"carrier,omitempty"`
	OriginPort      *string    `json:"origin_port,omitempty"`
	DestinationPort *string    `json:"destination_port,omitempty"`
	ETDFrom         *time.Time `json:"etd_from,omitempty"`
	ETDTo           *time.Time `json:"etd_to,omitempty"`
	Search          *string    `json:"search,omitempty"`
	Page            int        `json:"page"`
	Limit           int        `json:"limit"`
	SortBy          string     `json:"sort_by"`
	SortDir         string     `json:"sort_dir"`
}

type GetBookingWorkspaceDetailRequest struct {
	OrgID     int32 `json:"-"`
	BookingID int64 `json:"booking_id"`
}

type DirectUpdateBookingStatusRequest struct {
	OrgID     int32   `json:"-"`
	BookingID int64   `json:"booking_id"`
	Status    string  `json:"status"`
	Notes     *string `json:"notes,omitempty"`
}

type CreateShipmentFromBookingRequest struct {
	OrgID            int32      `json:"-"`
	BookingID        int64      `json:"booking_id"`
	VesselName       *string    `json:"vessel_name,omitempty"`
	VoyageNumber     *string    `json:"voyage_number,omitempty"`
	ContainerNumbers []string   `json:"container_numbers,omitempty"`
	ETD              *time.Time `json:"etd,omitempty"`
	ETA              *time.Time `json:"eta,omitempty"`
	Notes            *string    `json:"notes,omitempty"`
}

type GetEligibleRFQsForBookingRequest struct {
	OrgID int32 `json:"-"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Task 5: Real Carrier Booking Integration
// ──────────────────────────────────────────────────────────────────────────────

type BookWithCarrierRequest struct {
	OrgID          int32      `json:"-"`
	BookingID      int64      `json:"booking_id"`
	CarrierSCAC    *string    `json:"carrier_scac,omitempty"`
	EquipmentType  *string    `json:"equipment_type,omitempty"`
	Quantity       *int       `json:"quantity,omitempty"`
	ContractNumber *string    `json:"contract_number,omitempty"`
	CargoReadyDate *time.Time `json:"cargo_ready_date,omitempty"`
	Commodity      *string    `json:"commodity,omitempty"`
	ShipperName    *string    `json:"shipper_name,omitempty"`
	ConsigneeName  *string    `json:"consignee_name,omitempty"`
	ForceRetry     bool       `json:"force_retry,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
}

type SyncCarrierBookingRequest struct {
	OrgID     int32 `json:"-"`
	BookingID int64 `json:"booking_id"`
}





