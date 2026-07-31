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
