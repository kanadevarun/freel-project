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
}

type UpdateStageRequest struct {
	OrgID int32  `json:"-"`
	ID    int32  `json:"id"`
	Stage string `json:"stage"`
}

type ParseShipmentRequest struct {
	OrgID       int32  `json:"-"`
	RawEmail    string `json:"raw_email"`
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
	ID      int32 `json:"id"`      // RFQ ID — path param
	QuoteID int32 `json:"quote_id"` // Which quote to approve
}
