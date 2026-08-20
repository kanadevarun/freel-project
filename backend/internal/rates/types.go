package rates

import "time"

// RateSource identifies where a CanonicalRate originated.
// The Quotation Engine treats all sources identically — this field is only
// used for display, analytics, and priority ranking.
type RateSource string

const (
	RateSourceSpotAPI     RateSource = "SPOT_API"
	RateSourceContractPDF RateSource = "CONTRACT_PDF"
	RateSourceManual      RateSource = "MANUAL"
	RateSourceEmail       RateSource = "EMAIL"
)

// ExtractionStatus tracks the validation state of a rate.
// The Quotation Engine ONLY reads rates with status == ExtractionStatusConfirmed.
type ExtractionStatus string

const (
	ExtractionStatusConfirmed     ExtractionStatus = "CONFIRMED"
	ExtractionStatusPendingReview ExtractionStatus = "PENDING_REVIEW"
	ExtractionStatusFlagged       ExtractionStatus = "FLAGGED"
	ExtractionStatusRejected      ExtractionStatus = "REJECTED"
)

// SurchargeUnit defines how a surcharge amount is applied.
type SurchargeUnit string

const (
	SurchargeUnitPerTEU       SurchargeUnit = "PER_TEU"
	SurchargeUnitPerContainer SurchargeUnit = "PER_CONTAINER"
	SurchargeUnitPerShipment  SurchargeUnit = "PER_SHIPMENT"
	SurchargeUnitPercent      SurchargeUnit = "PERCENT"
)

// Surcharge represents a single named additional charge on a rate.
// Examples: BAF (Bunker Adjustment Factor), CAF (Currency Adjustment Factor),
// PSS (Peak Season Surcharge), OHC (Origin Handling Charge).
type Surcharge struct {
	// Code is the normalized industry-standard charge code.
	// e.g., "BAF", "CAF", "OHC", "DHC", "PSS", "WRS"
	Code string `json:"code" db:"code"`

	// Description is the human-readable charge name as it appears in the contract.
	Description string `json:"description" db:"description"`

	// Amount is the charge value in USD (normalized from original currency at ingestion).
	Amount float64 `json:"amount" db:"amount"`

	// Unit determines how the amount is applied.
	Unit SurchargeUnit `json:"unit" db:"unit"`

	// Included is true when this surcharge is already factored into OceanFreight
	// (i.e., the carrier offers an "all-in" rate for this charge).
	Included bool `json:"included" db:"included"`
}

// CanonicalRate is the single unified rate object used across the entire platform.
//
// Both the Spot Rate engine (from carrier APIs) and the Contract Rate engine
// (from AI-parsed PDFs) produce CanonicalRate objects. The Quotation Engine
// reads ONLY CanonicalRate — it is completely blind to the original source.
//
// Invariants enforced at ingestion time:
//   - All monetary fields are in USD.
//   - OriginPort and DestinationPort are UN/LOCODE (5-char, e.g., "INNSA").
//   - TotalBuyPrice = OceanFreight + OriginCharges + DestinationCharges
//     (included surcharges are already baked into OceanFreight).
//   - Only rows with ExtractionStatus == "CONFIRMED" are returned by SearchRates.
type CanonicalRate struct {
	ID     string `json:"id" db:"id"`
	OrgID  int64  `json:"org_id" db:"org_id"`

	// Source
	Source        RateSource `json:"source" db:"source"`
	SourceRef     string     `json:"source_ref" db:"source_ref"`
	ContractDocID *string    `json:"contract_doc_id,omitempty" db:"contract_doc_id"`

	// Route
	OriginPort      string `json:"origin_port" db:"origin_port"`
	DestinationPort string `json:"destination_port" db:"destination_port"`
	ViaPort         string `json:"via_port,omitempty" db:"via_port"`
	ServiceCode     string `json:"service_code,omitempty" db:"service_code"`

	// Carrier
	CarrierSCAC string `json:"carrier_scac" db:"carrier_scac"`
	CarrierName string `json:"carrier_name" db:"carrier_name"`
	VesselName  string `json:"vessel_name,omitempty" db:"vessel_name"`

	// Equipment
	EquipmentType string `json:"equipment_type" db:"equipment_type"`

	// Pricing — all USD
	OceanFreight       float64     `json:"ocean_freight" db:"ocean_freight"`
	OriginCharges      float64     `json:"origin_charges" db:"origin_charges"`
	DestinationCharges float64     `json:"destination_charges" db:"destination_charges"`
	Surcharges         []Surcharge `json:"surcharges" db:"-"`
	// SurchargesRaw is used for DB scan; Surcharges is the parsed form.
	SurchargesRaw    []byte   `json:"-" db:"surcharges"`
	TotalBuyPrice    float64  `json:"total_buy_price" db:"total_buy_price"`
	CurrencyOriginal string   `json:"currency_original" db:"currency_original"`
	ExchangeRateUsed float64  `json:"exchange_rate_used" db:"exchange_rate_used"`

	// Included / excluded charge codes for display
	IncludedCharges []string `json:"included_charges" db:"included_charges"`
	ExcludedCharges []string `json:"excluded_charges" db:"excluded_charges"`

	// Conditions
	FreeDaysOrigin        int      `json:"free_days_origin" db:"free_days_origin"`
	FreeDaysDestination   int      `json:"free_days_destination" db:"free_days_destination"`
	TransitDays           *int     `json:"transit_days,omitempty" db:"transit_days"`
	Incoterms             string   `json:"incoterms,omitempty" db:"incoterms"`
	CommodityRestrictions []string `json:"commodity_restrictions" db:"commodity_restrictions"`
	RoutingConditions     string   `json:"routing_conditions,omitempty" db:"routing_conditions"`

	// Validity
	ValidFrom  time.Time `json:"valid_from" db:"valid_from"`
	ValidUntil time.Time `json:"valid_until" db:"valid_until"`

	// Data quality
	ConfidenceScore  int              `json:"confidence_score" db:"confidence_score"`
	ExtractionStatus ExtractionStatus `json:"extraction_status" db:"extraction_status"`
	ExtractedBy      string           `json:"extracted_by" db:"extracted_by"`
	ReviewFlags      []string         `json:"review_flags,omitempty" db:"review_flags"`
	ReviewedBy       *int64           `json:"reviewed_by,omitempty" db:"reviewed_by"`
	ReviewedAt       *time.Time       `json:"reviewed_at,omitempty" db:"reviewed_at"`

	// Operational
	NauticalMiles int     `json:"nautical_miles,omitempty" db:"nautical_miles"`
	CO2PerTEU     float64 `json:"co2_per_teu,omitempty" db:"co2_per_teu"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// RateQuery is the input to rates.Service.SearchRates.
// The Quotation Engine and the Pricing Agent both use this struct.
type RateQuery struct {
	OrgID           int64
	OriginPort      string     // UN/LOCODE preferred; port_normalizer handles aliases
	DestinationPort string     // UN/LOCODE preferred
	EquipmentType   string     // "40GP" | "40HC" | "20GP" | "REEFER" — defaults to "40GP" if empty
	TargetDate      *time.Time // If set, only rates valid on this date are returned
	CarrierSCACs    []string   // Optional whitelist; empty = all carriers
	Sources         []RateSource // Optional whitelist; empty = all sources (spot + contract)
	MaxResults      int          // 0 = default of 20
	Incoterms       string
}

// RateSearchResult is the full response from rates.Service.SearchRates.
// The recommended rate is at index RecommendedIdx.
type RateSearchResult struct {
	Rates              []CanonicalRate `json:"rates"`
	TotalCount         int             `json:"total_count"`
	SpotRateCount      int             `json:"spot_rate_count"`
	ContractRateCount  int             `json:"contract_rate_count"`
	RecommendedIdx     int             `json:"recommended_idx"`
	OverallReasoning   string          `json:"overall_reasoning"`
	SearchedAt         time.Time       `json:"searched_at"`
}
