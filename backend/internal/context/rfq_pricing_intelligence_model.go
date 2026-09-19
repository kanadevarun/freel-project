package bcontext

import "time"

// RFQIdentitySummary holds identity, customer, routing, cargo, and completeness for an RFQ.
type RFQIdentitySummary struct {
	ID                    int64      `json:"id"`
	OrgID                 int64      `json:"org_id"`
	RFQNumber             string     `json:"rfq_number"`
	CustomerID            int64      `json:"customer_id"`
	CustomerName          string     `json:"customer_name"`
	CustomerCode          string     `json:"customer_code,omitempty"`
	Stage                 string     `json:"stage"`
	Origin                string     `json:"origin"`
	Destination           string     `json:"destination"`
	Incoterms             string     `json:"incoterms,omitempty"`
	TargetDate            *time.Time `json:"target_date,omitempty"`
	SalesAssigneeID       *int64     `json:"sales_assignee_id,omitempty"`
	SalesAssigneeName     string     `json:"sales_assignee_name,omitempty"`
	PricingAssigneeID     *int64     `json:"pricing_assignee_id,omitempty"`
	PricingAssigneeName   string     `json:"pricing_assignee_name,omitempty"`
	LeadID                *int64     `json:"lead_id,omitempty"`
	TotalItemsCount       int        `json:"total_items_count"`
	TotalWeightKg         float64    `json:"total_weight_kg"`
	TotalVolumeCbm        float64    `json:"total_volume_cbm"`
	CargoDescription      string     `json:"cargo_description,omitempty"`
	Currency              string     `json:"currency"`
	CompletenessScore     int        `json:"completeness_score"` // 0 - 100%
	IsComplete            bool       `json:"is_complete"`
	MissingRequiredFields []string   `json:"missing_required_fields"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// RFQQuotationItemSummary represents a commercial quotation row from `quotations`.
type RFQQuotationItemSummary struct {
	ID              int64      `json:"id"`
	QuotationNumber string     `json:"quotation_number"`
	Status          string     `json:"status"`
	Carrier         string     `json:"carrier,omitempty"`
	TotalAmount     float64    `json:"total_amount"`
	TotalCost       float64    `json:"total_cost"`
	GrossProfit     float64    `json:"gross_profit"`
	GrossMarginPct  float64    `json:"gross_margin_pct"`
	Currency        string     `json:"currency"`
	ValidFrom       *time.Time `json:"valid_from,omitempty"`
	ValidUntil      *time.Time `json:"valid_until,omitempty"`
	IsExpired       bool       `json:"is_expired"`
	CreatedAt       time.Time  `json:"created_at"`
}

// RFQCarrierQuoteItemSummary represents a carrier rate quote from `rfq_quotes`.
type RFQCarrierQuoteItemSummary struct {
	ID               int64     `json:"id"`
	CarrierName      string    `json:"carrier_name"`
	TransitTimeDays  int       `json:"transit_time_days"`
	BuyPrice         float64   `json:"buy_price"`
	SellPrice        float64   `json:"sell_price"`
	GrossMarginPct   float64   `json:"gross_margin_pct"`
	IsRecommended    bool      `json:"is_recommended"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}

// RFQQuotationComparison summarizes all quotation records associated with this RFQ.
type RFQQuotationComparison struct {
	TotalQuotationsRequested int                          `json:"total_quotations_requested"`
	TotalQuotationsReceived  int                          `json:"total_quotations_received"`
	ValidQuotationsCount     int                          `json:"valid_quotations_count"`
	ExpiredQuotationsCount   int                          `json:"expired_quotations_count"`
	RejectedQuotationsCount  int                          `json:"rejected_quotations_count"`
	LowestValidPrice         *float64                     `json:"lowest_valid_price,omitempty"`
	HighestValidPrice        *float64                     `json:"highest_valid_price,omitempty"`
	AverageValidPrice        *float64                     `json:"average_valid_price,omitempty"`
	MedianValidPrice         *float64                     `json:"median_valid_price,omitempty"`
	PriceSpread              *float64                     `json:"price_spread,omitempty"`     // Highest - Lowest
	PriceSpreadPct           *float64                     `json:"price_spread_pct,omitempty"` // (Spread / Lowest) * 100
	Currency                 string                       `json:"currency"`
	SelectedQuotationID      *int64                       `json:"selected_quotation_id,omitempty"`
	SelectedQuotationNumber  string                       `json:"selected_quotation_number,omitempty"`
	SelectedPrice            *float64                     `json:"selected_price,omitempty"`
	SelectedCarrier          string                       `json:"selected_carrier,omitempty"`
	TimeToFirstQuoteHours    *float64                     `json:"time_to_first_quote_hours,omitempty"`
	TimeToFinalSelectHours   *float64                     `json:"time_to_final_select_hours,omitempty"`
	Quotations               []RFQQuotationItemSummary    `json:"quotations"`
	CarrierQuotes            []RFQCarrierQuoteItemSummary `json:"carrier_quotes"`
}

// RFQCommercialPerformance represents win rates, lane patterns, and conversions.
type RFQCommercialPerformance struct {
	HasQuotation               bool     `json:"has_quotation"`
	HasBooking                 bool     `json:"has_booking"`
	HasShipment                bool     `json:"has_shipment"`
	LinkedBookingID            *int64   `json:"linked_booking_id,omitempty"`
	LinkedShipmentID           *int64   `json:"linked_shipment_id,omitempty"`
	LaneAverageQuotedPrice     *float64 `json:"lane_average_quoted_price,omitempty"`
	LaneHistoricalQuotesCount  int      `json:"lane_historical_quotes_count"`
	LanePricingVolatilityNote  string   `json:"lane_pricing_volatility_note"`
	CustomerHistoricalWinRate  *float64 `json:"customer_historical_win_rate,omitempty"`
	CustomerTotalRFQs          int      `json:"customer_total_rfqs"`
	CarrierResponseRatePct     *float64 `json:"carrier_response_rate_pct,omitempty"`
	CarrierAverageResponseDays *float64 `json:"carrier_average_response_days,omitempty"`
}

// RFQMarginAndRiskIndicators evaluates commercial risk and pricing anomalies.
type RFQMarginAndRiskIndicators struct {
	QuotedRevenue      *float64 `json:"quoted_revenue,omitempty"`
	RecordedCost       *float64 `json:"recorded_cost,omitempty"`
	GrossMarginAmount  *float64 `json:"gross_margin_amount,omitempty"`
	GrossMarginPct     *float64 `json:"gross_margin_pct,omitempty"`
	MarginHealth       string   `json:"margin_health"`       // HEALTHY, THIN, NEGATIVE, UNKNOWN
	IsLowMargin        bool     `json:"is_low_margin"`        // < 10%
	IsHighPriceAnomaly bool     `json:"is_high_price_anomaly"` // > 50% above lane average
	IsMissingCostData  bool     `json:"is_missing_cost_data"`  // cost == 0 with revenue > 0
	IsRateExpired      bool     `json:"is_rate_expired"`
	RiskSeverity       string   `json:"risk_severity"` // LOW, MEDIUM, HIGH, CRITICAL
	RiskFactors        []string `json:"risk_factors"`
}

// RFQObservation holds a grounded evidence item with source traceability.
type RFQObservation struct {
	Module       string     `json:"module"` // "RFQ", "QUOTATION", "CARRIER", "BOOKING", "LANE"
	RecordType   string     `json:"record_type"`
	RecordID     string     `json:"record_id"`
	Finding      string     `json:"finding"`
	Significance string     `json:"significance"` // "POSITIVE", "ATTENTION", "CRITICAL", "NEUTRAL"
	Timestamp    *time.Time `json:"timestamp,omitempty"`
}

// RFQAISummary contains the synthesized read-only analysis.
type RFQAISummary struct {
	ExecutiveSummary         string           `json:"executive_summary"`
	QuotationSpreadAnalysis  string           `json:"quotation_spread_analysis"`
	CarrierResponseInsight   string           `json:"carrier_response_insight"`
	CommercialAndMarginRisks string           `json:"commercial_and_margin_risks"`
	AttentionItems           []string         `json:"attention_items"`
	SuggestedQuestions       []string         `json:"suggested_questions"`
	SupportingObservations   []RFQObservation `json:"supporting_observations"`
	ConfidenceLevel          string           `json:"confidence_level"` // "HIGH", "MEDIUM", "LOW"
	DataFreshness            string           `json:"data_freshness"`
	DataWarnings             []string         `json:"data_warnings"`
	Classification           string           `json:"classification"` // "READ_ONLY_INFORMATIONAL"
}

// RFQ360PricingIntelligence is the complete payload for RFQ and pricing intelligence.
type RFQ360PricingIntelligence struct {
	OrgID                 int64                      `json:"org_id"`
	RFQID                 int64                      `json:"rfq_id"`
	Identity              RFQIdentitySummary         `json:"identity"`
	QuotationComparison  RFQQuotationComparison     `json:"quotation_comparison"`
	CommercialPerformance RFQCommercialPerformance   `json:"commercial_performance"`
	MarginAndRisk         RFQMarginAndRiskIndicators `json:"margin_and_risk"`
	AISummary             RFQAISummary               `json:"ai_summary"`
	CorrelationID         string                     `json:"correlation_id"`
	DataFreshness         time.Time                  `json:"data_freshness"`
	IsReadOnly            bool                       `json:"is_read_only"`
}
