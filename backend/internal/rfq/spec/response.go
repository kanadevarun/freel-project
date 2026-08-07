package spec

import "time"

// RFQ models

type RFQ struct {
	ID                int32      `json:"id" db:"id"`
	OrgID             int32      `json:"org_id" db:"org_id"`
	RFQNumber         string     `json:"rfq_number" db:"rfq_number"`
	CustomerID        int32      `json:"customer_id" db:"customer_id"`
	// CustomerName is populated via JOIN with the customers table in ListRFQs.
	// It is NOT stored in the rfqs table — it is a read-only display field.
	CustomerName      string     `json:"customer_name" db:"customer_name"`
	Stage             string     `json:"stage" db:"stage"`
	Origin            *string    `json:"origin" db:"origin"`
	Destination       *string    `json:"destination" db:"destination"`
	Incoterms         *string    `json:"incoterms" db:"incoterms"`
	TargetDate        *time.Time `json:"target_date" db:"target_date"`
	SalesAssigneeID   *int32     `json:"sales_assignee_id" db:"sales_assignee_id"`
	PricingAssigneeID *int32     `json:"pricing_assignee_id" db:"pricing_assignee_id"`
	HealthScore       int        `json:"health_score" db:"health_score"`
	AgentStatus       string     `json:"agent_status" db:"agent_status"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
	Items             []RFQItem  `json:"items,omitempty"`
	Quotes            []Quote    `json:"quotes,omitempty"`
}

type RFQItem struct {
	ID          int32     `json:"id"`
	RFQID       int32     `json:"rfq_id"`
	Description string    `json:"description"`
	Quantity    int       `json:"quantity"`
	WeightKG    *float64  `json:"weight_kg"`
	VolumeCBM   *float64  `json:"volume_cbm"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Quote struct {
	ID                    int32     `json:"id"`
	RFQID                 int32     `json:"rfq_id"`
	CarrierName           string    `json:"carrier_name"`
	TransitTimeDays       *int      `json:"transit_time_days"`
	BuyPrice              float64   `json:"buy_price"`
	SellPrice             float64   `json:"sell_price"`
	IsRecommended         bool      `json:"is_recommended"`
	ReliabilityScore      int       `json:"reliability_score"`
	HistoricalSuccessRate float64   `json:"historical_success_rate"`
	AiReasoning           *string   `json:"ai_reasoning"`
	Status                string    `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// Responses

type ListRFQsResponse struct {
	Data       []RFQ `json:"data"`
	TotalCount int   `json:"total_count"`
}

type GetRFQResponse struct {
	Data RFQ `json:"data"`
}

type GetTimelineResponse struct {
	Data []interface{} `json:"data"` // simplified for now
}

type GetAgentStatusResponse struct {
	Data interface{} `json:"data"`
}

type ParseShipmentResponse struct {
	Data interface{} `json:"data"`
}

// GetCarrierRatesResponse wraps the carrier service response for the HTTP layer.
// The data field contains a ranked list of carrier options with AI reasoning.
type GetCarrierRatesResponse struct {
	Data interface{} `json:"data"`
}

// ApproveQuoteResponse confirms the stage advance and returns the updated RFQ.
type ApproveQuoteResponse struct {
	Data interface{} `json:"data"`
}
