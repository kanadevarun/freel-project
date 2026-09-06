package spec

import "time"

// Rate represents a persisted rate entity
type Rate struct {
	ID                 int64      `json:"id" db:"id"`
	OrgID              int64      `json:"org_id" db:"org_id"`
	RateReference      string     `json:"rate_reference" db:"rate_reference"`
	CarrierName        string     `json:"carrier_name" db:"carrier_name"`
	CarrierCode        *string    `json:"carrier_code,omitempty" db:"carrier_code"`
	ServiceProvider    *string    `json:"service_provider,omitempty" db:"service_provider"`
	RateType           string     `json:"rate_type" db:"rate_type"`
	TransportMode      string     `json:"transport_mode" db:"transport_mode"`
	ServiceType        string     `json:"service_type" db:"service_type"`
	EquipmentType      *string    `json:"equipment_type,omitempty" db:"equipment_type"`
	OriginPort         string     `json:"origin_port" db:"origin_port"`
	OriginCode         *string    `json:"origin_code,omitempty" db:"origin_code"`
	DestinationPort    string     `json:"destination_port" db:"destination_port"`
	DestinationCode    *string    `json:"destination_code,omitempty" db:"destination_code"`
	Currency           string     `json:"currency" db:"currency"`
	BaseAmount         float64    `json:"base_amount" db:"base_amount"`
	EffectiveDate      time.Time  `json:"effective_date" db:"effective_date"`
	ExpiryDate         time.Time  `json:"expiry_date" db:"expiry_date"`
	Status             string     `json:"status" db:"status"`
	CarrierReference   *string    `json:"carrier_reference,omitempty" db:"carrier_reference"`
	ContractReference  *string    `json:"contract_reference,omitempty" db:"contract_reference"`
	ContractID         *int64     `json:"contract_id,omitempty" db:"contract_id"`
	VersionNumber      int        `json:"version_number" db:"version_number"`
	VersionStatus      string     `json:"version_status" db:"version_status"`
	SupersedesRateID   *int64     `json:"supersedes_rate_id,omitempty" db:"supersedes_rate_id"`
	SupersededByRateID *int64     `json:"superseded_by_rate_id,omitempty" db:"superseded_by_rate_id"`
	VersionCreatedAt   *time.Time `json:"version_created_at,omitempty" db:"version_created_at"`
	Notes              *string    `json:"notes,omitempty" db:"notes"`
	CreatedBy          *string    `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy          *string    `json:"updated_by,omitempty" db:"updated_by"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// RateListItem is the formatted response for listing rates in UI tables
type RateListItem struct {
	ID                 int64     `json:"id" db:"id"`
	RateReference      string    `json:"rate_reference" db:"rate_reference"`
	CarrierName        string    `json:"carrier_name" db:"carrier_name"`
	CarrierCode        *string   `json:"carrier_code,omitempty" db:"carrier_code"`
	RateType           string    `json:"rate_type" db:"rate_type"`
	TransportMode      string    `json:"transport_mode" db:"transport_mode"`
	ServiceType        string    `json:"service_type" db:"service_type"`
	EquipmentType      *string   `json:"equipment_type,omitempty" db:"equipment_type"`
	OriginPort         string    `json:"origin_port" db:"origin_port"`
	OriginCode         *string   `json:"origin_code,omitempty" db:"origin_code"`
	DestinationPort    string    `json:"destination_port" db:"destination_port"`
	DestinationCode    *string   `json:"destination_code,omitempty" db:"destination_code"`
	Currency           string    `json:"currency" db:"currency"`
	BaseAmount         float64   `json:"base_amount" db:"base_amount"`
	EffectiveDate      string    `json:"effective_date" db:"effective_date"`
	ExpiryDate         string    `json:"expiry_date" db:"expiry_date"`
	Status             string    `json:"status" db:"status"`
	ContractID         *int64    `json:"contract_id,omitempty" db:"contract_id"`
	VersionNumber      int       `json:"version_number" db:"version_number"`
	VersionStatus      string    `json:"version_status" db:"version_status"`
	CarrierReference   *string   `json:"carrier_reference,omitempty" db:"carrier_reference"`
	ContractReference  *string   `json:"contract_reference,omitempty" db:"contract_reference"`
	DaysUntilExpiry    int       `json:"days_until_expiry"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// RateDetail is the response for rate slide-over drawers / details
type RateDetail struct {
	Rate
	DaysUntilExpiry int      `json:"days_until_expiry"`
	IsExpired       bool     `json:"is_expired"`
	IsExpiringSoon  bool     `json:"is_expiring_soon"`
	LaneDisplay     string   `json:"lane_display"`
	ValidityText    string   `json:"validity_text"`
	Tags            []string `json:"tags"`
}

// CarrierCoverageSummary provides lane statistics per carrier
type CarrierCoverageSummary struct {
	CarrierName string  `json:"carrier_name" db:"carrier_name"`
	CarrierCode *string `json:"carrier_code,omitempty" db:"carrier_code"`
	LaneCount   int     `json:"lane_count" db:"lane_count"`
	RateCount   int     `json:"rate_count" db:"rate_count"`
	SharePct    float64 `json:"share_pct"`
}

// RecentRateUpdate summary for sidebar widgets
type RecentRateUpdate struct {
	ID            int64     `json:"id" db:"id"`
	RateReference string    `json:"rate_reference" db:"rate_reference"`
	CarrierName   string    `json:"carrier_name" db:"carrier_name"`
	OriginPort    string    `json:"origin_port" db:"origin_port"`
	DestPort      string    `json:"dest_port" db:"destination_port"`
	BaseAmount    float64   `json:"base_amount" db:"base_amount"`
	Currency      string    `json:"currency" db:"currency"`
	Status        string    `json:"status" db:"status"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// RateSummaryKPIs captures the 5 top cards + sidebar widgets
type RateSummaryKPIs struct {
	TotalRates        int                      `json:"total_rates"`
	ActiveRates       int                      `json:"active_rates"`
	ExpiringSoonRates int                      `json:"expiring_soon_rates"`
	ExpiredRates      int                      `json:"expired_rates"`
	LanesCovered      int                      `json:"lanes_covered"`
	TopCarriers       []CarrierCoverageSummary `json:"top_carriers"`
	RecentUpdates     []RecentRateUpdate       `json:"recent_updates"`
	ActivePct         float64                  `json:"active_pct"`
	ExpiringSoonPct   float64                  `json:"expiring_soon_pct"`
	ExpiredPct        float64                  `json:"expired_pct"`
}

// ListRatesResponse is the paginated response for listing rates
type ListRatesResponse struct {
	Rates      []*RateListItem `json:"rates"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}

// Surcharge describes additional line item charges on canonical rate
type Surcharge struct {
	Code        string  `json:"code" db:"code"`
	Description string  `json:"description" db:"description"`
	Amount      float64 `json:"amount" db:"amount"`
	Unit        string  `json:"unit" db:"unit"`
	Included    bool    `json:"included" db:"included"`
}

// CanonicalRate object for intelligence engine
type CanonicalRate struct {
	ID                    string      `json:"id" db:"id"`
	OrgID                 int64       `json:"org_id" db:"org_id"`
	Source                string      `json:"source" db:"source"`
	SourceRef             string      `json:"source_ref" db:"source_ref"`
	ContractDocID         *string     `json:"contract_doc_id,omitempty" db:"contract_doc_id"`
	OriginPort            string      `json:"origin_port" db:"origin_port"`
	DestinationPort       string      `json:"destination_port" db:"destination_port"`
	ViaPort               string      `json:"via_port,omitempty" db:"via_port"`
	ServiceCode           string      `json:"service_code,omitempty" db:"service_code"`
	CarrierSCAC           string      `json:"carrier_scac" db:"carrier_scac"`
	CarrierName           string      `json:"carrier_name" db:"carrier_name"`
	VesselName            string      `json:"vessel_name,omitempty" db:"vessel_name"`
	EquipmentType         string      `json:"equipment_type" db:"equipment_type"`
	OceanFreight          float64     `json:"ocean_freight" db:"ocean_freight"`
	OriginCharges         float64     `json:"origin_charges" db:"origin_charges"`
	DestinationCharges    float64     `json:"destination_charges" db:"destination_charges"`
	Surcharges            []Surcharge `json:"surcharges" db:"-"`
	TotalBuyPrice         float64     `json:"total_buy_price" db:"total_buy_price"`
	CurrencyOriginal      string      `json:"currency_original" db:"currency_original"`
	ExchangeRateUsed      float64     `json:"exchange_rate_used" db:"exchange_rate_used"`
	IncludedCharges       []string    `json:"included_charges" db:"included_charges"`
	ExcludedCharges       []string    `json:"excluded_charges" db:"excluded_charges"`
	FreeDaysOrigin        int         `json:"free_days_origin" db:"free_days_origin"`
	FreeDaysDestination   int         `json:"free_days_destination" db:"free_days_destination"`
	TransitDays           *int        `json:"transit_days,omitempty" db:"transit_days"`
	Incoterms             string      `json:"incoterms,omitempty" db:"incoterms"`
	CommodityRestrictions []string    `json:"commodity_restrictions" db:"commodity_restrictions"`
	RoutingConditions     string      `json:"routing_conditions,omitempty" db:"routing_conditions"`
	ValidFrom             time.Time   `json:"valid_from" db:"valid_from"`
	ValidUntil            time.Time   `json:"valid_until" db:"valid_until"`
	ConfidenceScore       int         `json:"confidence_score" db:"confidence_score"`
	ExtractionStatus      string      `json:"extraction_status" db:"extraction_status"`
	ExtractedBy           string      `json:"extracted_by" db:"extracted_by"`
	NauticalMiles         int         `json:"nautical_miles,omitempty" db:"nautical_miles"`
	CO2PerTEU             float64     `json:"co2_per_teu,omitempty" db:"co2_per_teu"`
	CreatedAt             time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time   `json:"updated_at" db:"updated_at"`
}

// RateSearchResult is the response from canonical rate search
type RateSearchResult struct {
	Rates             []CanonicalRate `json:"rates"`
	TotalCount        int             `json:"total_count"`
	SpotRateCount     int             `json:"spot_rate_count"`
	ContractRateCount int             `json:"contract_rate_count"`
	RecommendedIdx    int             `json:"recommended_idx"`
	OverallReasoning  string          `json:"overall_reasoning"`
	SearchedAt        time.Time       `json:"searched_at"`
}

// ── Task 19.2: Rate Charges & Commercial Pricing Models ──────────────────────

// RateChargeItem represents a persisted charge item belonging to a rate
type RateChargeItem struct {
	ID                 int64     `json:"id" db:"id"`
	OrgID              int64     `json:"org_id" db:"org_id"`
	RateID             int64     `json:"rate_id" db:"rate_id"`
	ChargeCategory     string    `json:"charge_category" db:"charge_category"`
	ChargeCode         string    `json:"charge_code" db:"charge_code"`
	ChargeName         string    `json:"charge_name" db:"charge_name"`
	CalculationBasis   string    `json:"calculation_basis" db:"calculation_basis"`
	Quantity           float64   `json:"quantity" db:"quantity"`
	UnitPrice          float64   `json:"unit_price" db:"unit_price"`
	Currency           string    `json:"currency" db:"currency"`
	MinimumAmount      *float64  `json:"minimum_amount,omitempty" db:"minimum_amount"`
	MaximumAmount      *float64  `json:"maximum_amount,omitempty" db:"maximum_amount"`
	IncludedInBaseRate bool      `json:"included_in_base_rate" db:"included_in_base_rate"`
	DisplayOrder       int       `json:"display_order" db:"display_order"`
	Notes              *string   `json:"notes,omitempty" db:"notes"`
	CalculatedAmount   float64   `json:"calculated_amount" db:"-"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// CategoryTotal captures the subtotal for a given charge category
type CategoryTotal struct {
	Category    string  `json:"category"`
	ChargeCount int     `json:"charge_count"`
	TotalAmount float64 `json:"total_amount"`
	Currency    string  `json:"currency"`
}

// RatePricingSummary represents the deterministic commercial breakdown of a rate
type RatePricingSummary struct {
	RateID              int64                    `json:"rate_id"`
	RateReference       string                   `json:"rate_reference"`
	BaseRate            float64                  `json:"base_rate"`
	BaseCurrency        string                   `json:"base_currency"`
	AdditionalCharges   float64                  `json:"additional_charges"`
	CommercialTotal     float64                  `json:"commercial_total"`
	ChargeCount         int                      `json:"charge_count"`
	IsMultiCurrency     bool                     `json:"is_multi_currency"`
	Currencies          []string                 `json:"currencies"`
	CurrencyBreakdown   map[string]float64       `json:"currency_breakdown"`
	CategoryTotals      []CategoryTotal          `json:"category_totals"`
	Charges             []RateChargeItem         `json:"charges"`
}

// RatePricingResponse is the API response payload for GET /api/v1/rates/{id}/pricing
type RatePricingResponse struct {
	Pricing RatePricingSummary `json:"pricing"`
}

// ── Task 19.3: Rate Contract & Versioning Models ──────────────────────────────

// RateContract represents a persisted carrier rate contract
type RateContract struct {
	ID                int64     `json:"id" db:"id"`
	OrgID             int64     `json:"org_id" db:"org_id"`
	ContractReference string    `json:"contract_reference" db:"contract_reference"`
	CarrierName       string    `json:"carrier_name" db:"carrier_name"`
	CarrierCode       *string   `json:"carrier_code,omitempty" db:"carrier_code"`
	ContractName      string    `json:"contract_name" db:"contract_name"`
	ContractType      string    `json:"contract_type" db:"contract_type"`
	TransportMode     *string   `json:"transport_mode,omitempty" db:"transport_mode"`
	Currency          *string   `json:"currency,omitempty" db:"currency"`
	EffectiveDate     time.Time `json:"effective_date" db:"effective_date"`
	ExpiryDate        time.Time `json:"expiry_date" db:"expiry_date"`
	Status            string    `json:"status" db:"status"`
	RenewalStatus     string    `json:"renewal_status" db:"renewal_status"`
	RenewalOwner      *string   `json:"renewal_owner,omitempty" db:"renewal_owner"`
	Notes             *string   `json:"notes,omitempty" db:"notes"`
	CreatedBy         *string   `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy         *string   `json:"updated_by,omitempty" db:"updated_by"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// RateContractListItem represents contract row in UI table
type RateContractListItem struct {
	ID                int64     `json:"id" db:"id"`
	ContractReference string    `json:"contract_reference" db:"contract_reference"`
	CarrierName       string    `json:"carrier_name" db:"carrier_name"`
	CarrierCode       *string   `json:"carrier_code,omitempty" db:"carrier_code"`
	ContractName      string    `json:"contract_name" db:"contract_name"`
	ContractType      string    `json:"contract_type" db:"contract_type"`
	TransportMode     string    `json:"transport_mode" db:"transport_mode"`
	Currency          string    `json:"currency" db:"currency"`
	EffectiveDate     string    `json:"effective_date" db:"effective_date"`
	ExpiryDate        string    `json:"expiry_date" db:"expiry_date"`
	Status            string    `json:"status" db:"status"`
	RenewalStatus     string    `json:"renewal_status" db:"renewal_status"`
	RenewalOwner      *string   `json:"renewal_owner,omitempty" db:"renewal_owner"`
	LinkedRateCount   int       `json:"linked_rate_count" db:"linked_rate_count"`
	DaysUntilExpiry   int       `json:"days_until_expiry"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// RateContractSummary captures top KPIs for Contracts Workspace
type RateContractSummary struct {
	TotalContracts       int                    `json:"total_contracts"`
	ActiveContracts      int                    `json:"active_contracts"`
	ExpiringSoonContracts int                   `json:"expiring_soon_contracts"`
	ExpiredContracts     int                    `json:"expired_contracts"`
	RenewalRequired      int                    `json:"renewal_required"`
	TotalLinkedRates     int                    `json:"total_linked_rates"`
	ExpiringSoonList     []RateContractListItem `json:"expiring_soon_list"`
}

// ListRateContractsResponse is paginated response for GET /api/v1/rates/contracts
type ListRateContractsResponse struct {
	Contracts  []*RateContractListItem `json:"contracts"`
	TotalCount int                     `json:"total_count"`
	Page       int                     `json:"page"`
	Limit      int                     `json:"limit"`
	TotalPages int                     `json:"total_pages"`
}

// RateVersionHistory represents audit log record for rate revisions
type RateVersionHistory struct {
	ID             int64      `json:"id" db:"id"`
	OrgID          int64      `json:"org_id" db:"org_id"`
	RateID         int64      `json:"rate_id" db:"rate_id"`
	VersionNumber  int        `json:"version_number" db:"version_number"`
	Action         string     `json:"action" db:"action"`
	PreviousRateID *int64     `json:"previous_rate_id,omitempty" db:"previous_rate_id"`
	NewRateID      *int64     `json:"new_rate_id,omitempty" db:"new_rate_id"`
	Description    string     `json:"description" db:"description"`
	PerformedBy    *string    `json:"performed_by,omitempty" db:"performed_by"`
	Metadata       *string    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// RateVersionChainItem represents historical version in version timeline
type RateVersionChainItem struct {
	RateID           int64     `json:"rate_id"`
	RateReference    string    `json:"rate_reference"`
	VersionNumber    int       `json:"version_number"`
	VersionStatus    string    `json:"version_status"`
	BaseAmount       float64   `json:"base_amount"`
	Currency         string    `json:"currency"`
	EffectiveDate    string    `json:"effective_date"`
	ExpiryDate       string    `json:"expiry_date"`
	Status           string    `json:"status"`
	CommercialTotal  float64   `json:"commercial_total"`
	ChargeCount      int       `json:"charge_count"`
	SupersededByRate *int64    `json:"superseded_by_rate_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// ── Task 19.4: Spot Rate Requests, Responses & Comparison Models ─────────────

// SpotRateRequest represents a spot freight sourcing request
type SpotRateRequest struct {
	ID                int64      `json:"id" db:"id"`
	OrgID             int64      `json:"org_id" db:"org_id"`
	RequestReference  string     `json:"request_reference" db:"request_reference"`
	CustomerID        *int64     `json:"customer_id,omitempty" db:"customer_id"`
	CustomerName      *string    `json:"customer_name,omitempty" db:"customer_name"`
	OriginPort        string     `json:"origin_port" db:"origin_port"`
	OriginCode        *string    `json:"origin_code,omitempty" db:"origin_code"`
	DestinationPort   string     `json:"destination_port" db:"destination_port"`
	DestinationCode   *string    `json:"destination_code,omitempty" db:"destination_code"`
	TransportMode     string     `json:"transport_mode" db:"transport_mode"`
	ServiceType       string     `json:"service_type" db:"service_type"`
	EquipmentType     *string    `json:"equipment_type,omitempty" db:"equipment_type"`
	Commodity         *string    `json:"commodity,omitempty" db:"commodity"`
	CargoWeight       *float64   `json:"cargo_weight,omitempty" db:"cargo_weight"`
	CargoVolume       *float64   `json:"cargo_volume,omitempty" db:"cargo_volume"`
	ContainerQuantity int        `json:"container_quantity" db:"container_quantity"`
	ReadyDate         string     `json:"ready_date" db:"ready_date"`
	TargetCurrency    string     `json:"target_currency" db:"target_currency"`
	RequiredByDate    string     `json:"required_by_date" db:"required_by_date"`
	Status            string     `json:"status" db:"status"`
	Notes             *string    `json:"notes,omitempty" db:"notes"`
	CreatedBy         *string    `json:"created_by,omitempty" db:"created_by"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// SpotRateRequestListItem represents row in spot requests table
type SpotRateRequestListItem struct {
	ID                int64   `json:"id" db:"id"`
	RequestReference  string  `json:"request_reference" db:"request_reference"`
	CustomerName      *string `json:"customer_name,omitempty" db:"customer_name"`
	OriginPort        string  `json:"origin_port" db:"origin_port"`
	OriginCode        *string `json:"origin_code,omitempty" db:"origin_code"`
	DestinationPort   string  `json:"destination_port" db:"destination_port"`
	DestinationCode   *string `json:"destination_code,omitempty" db:"destination_code"`
	TransportMode     string  `json:"transport_mode" db:"transport_mode"`
	ServiceType       string  `json:"service_type" db:"service_type"`
	EquipmentType     *string `json:"equipment_type,omitempty" db:"equipment_type"`
	ContainerQuantity int     `json:"container_quantity" db:"container_quantity"`
	ReadyDate         string  `json:"ready_date" db:"ready_date"`
	RequiredByDate    string  `json:"required_by_date" db:"required_by_date"`
	Status            string  `json:"status" db:"status"`
	TargetCurrency    string  `json:"target_currency" db:"target_currency"`
	ResponseCount     int     `json:"response_count" db:"response_count"`
	HasPreferred      bool    `json:"has_preferred" db:"has_preferred"`
	DaysUntilRequired int     `json:"days_until_required"`
	CreatedAt         string  `json:"created_at" db:"created_at"`
}

// ListSpotRateRequestsResponse is paginated response
type ListSpotRateRequestsResponse struct {
	Requests   []*SpotRateRequestListItem `json:"requests"`
	TotalCount int                        `json:"total_count"`
	Page       int                        `json:"page"`
	Limit      int                        `json:"limit"`
	TotalPages int                        `json:"total_pages"`
}

// SpotRateResponseCharge represents itemized charge for a carrier response
type SpotRateResponseCharge struct {
	ID                 int64     `json:"id" db:"id"`
	OrgID              int64     `json:"org_id" db:"org_id"`
	SpotRateResponseID int64     `json:"spot_rate_response_id" db:"spot_rate_response_id"`
	ChargeCategory     string    `json:"charge_category" db:"charge_category"`
	ChargeName         string    `json:"charge_name" db:"charge_name"`
	CalculationBasis   string    `json:"calculation_basis" db:"calculation_basis"`
	Quantity           float64   `json:"quantity" db:"quantity"`
	UnitPrice          float64   `json:"unit_price" db:"unit_price"`
	Currency           string    `json:"currency" db:"currency"`
	TotalChargeAmount  float64   `json:"total_charge_amount"`
	DisplayOrder       int       `json:"display_order" db:"display_order"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// SpotRateResponse represents a carrier/supplier rate quote for a spot request
type SpotRateResponse struct {
	ID                    int64                    `json:"id" db:"id"`
	OrgID                 int64                    `json:"org_id" db:"org_id"`
	SpotRateRequestID     int64                    `json:"spot_rate_request_id" db:"spot_rate_request_id"`
	CarrierName           string                   `json:"carrier_name" db:"carrier_name"`
	CarrierCode           *string                  `json:"carrier_code,omitempty" db:"carrier_code"`
	SupplierName          *string                  `json:"supplier_name,omitempty" db:"supplier_name"`
	RateID                *int64                   `json:"rate_id,omitempty" db:"rate_id"`
	Currency              string                   `json:"currency" db:"currency"`
	BaseAmount            float64                  `json:"base_amount" db:"base_amount"`
	TotalAmount           float64                  `json:"total_amount" db:"total_amount"`
	TransitDays           *int                     `json:"transit_days,omitempty" db:"transit_days"`
	FreeDaysOrigin        int                      `json:"free_days_origin" db:"free_days_origin"`
	FreeDaysDestination   int                      `json:"free_days_destination" db:"free_days_destination"`
	ValidFrom             string                   `json:"valid_from" db:"valid_from"`
	ValidUntil            string                   `json:"valid_until" db:"valid_until"`
	RoutingNotes          *string                  `json:"routing_notes,omitempty" db:"routing_notes"`
	ResponseNotes         *string                  `json:"response_notes,omitempty" db:"response_notes"`
	Status                string                   `json:"status" db:"status"`
	IsPreferred           bool                     `json:"is_preferred" db:"is_preferred"`
	RespondedAt           time.Time                `json:"responded_at" db:"responded_at"`
	CreatedBy             *string                  `json:"created_by,omitempty" db:"created_by"`
	CreatedAt             time.Time                `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time                `json:"updated_at" db:"updated_at"`
	Charges               []SpotRateResponseCharge `json:"charges,omitempty"`
}

// SpotRateComparisonItem represents a row in the comparison matrix
type SpotRateComparisonItem struct {
	ResponseID            int64    `json:"response_id"`
	CarrierName           string   `json:"carrier_name"`
	CarrierCode           *string  `json:"carrier_code,omitempty"`
	SupplierName          *string  `json:"supplier_name,omitempty"`
	Currency              string   `json:"currency"`
	BaseAmount            float64  `json:"base_amount"`
	TotalCommercialAmount float64  `json:"total_commercial_amount"`
	TransitDays           *int     `json:"transit_days,omitempty"`
	FreeDaysDestination   int      `json:"free_days_destination"`
	ValidUntil            string   `json:"valid_until"`
	Status                string   `json:"status"`
	IsPreferred           bool     `json:"is_preferred"`
	ChargeCount           int      `json:"charge_count"`
	RecommendationTags    []string `json:"recommendation_tags"`
	ValueScore            float64  `json:"value_score"`
}

// SpotRateComparison contains side-by-side comparison analytics
type SpotRateComparison struct {
	RequestID               int64                    `json:"request_id"`
	RequestReference        string                   `json:"request_reference"`
	Lane                    string                   `json:"lane"`
	TransportMode           string                   `json:"transport_mode"`
	EquipmentType           string                   `json:"equipment_type"`
	TargetCurrency          string                   `json:"target_currency"`
	Responses               []SpotRateComparisonItem `json:"responses"`
	TotalResponsesCount     int                      `json:"total_responses_count"`
	CheapestResponseID      *int64                   `json:"cheapest_response_id,omitempty"`
	CheapestCarrierName     *string                  `json:"cheapest_carrier_name,omitempty"`
	CheapestAmount          *float64                 `json:"cheapest_amount,omitempty"`
	FastestResponseID       *int64                   `json:"fastest_response_id,omitempty"`
	FastestCarrierName      *string                  `json:"fastest_carrier_name,omitempty"`
	FastestTransitDays      *int                     `json:"fastest_transit_days,omitempty"`
	BestValueResponseID     *int64                   `json:"best_value_response_id,omitempty"`
	BestValueCarrierName    *string                  `json:"best_value_carrier_name,omitempty"`
	PreferredResponseID     *int64                   `json:"preferred_response_id,omitempty"`
	PreferredCarrierName    *string                  `json:"preferred_carrier_name,omitempty"`
	IsMultiCurrency         bool                     `json:"is_multi_currency"`
	MultiCurrencyCurrencies []string                 `json:"multi_currency_currencies,omitempty"`
	ComparisonSummaryNote   string                   `json:"comparison_summary_note"`
}

// SpotRateRequestSummary provides top-level KPI metrics
type SpotRateRequestSummary struct {
	TotalRequests      int                        `json:"total_requests"`
	OpenRequests       int                        `json:"open_requests"`
	AwaitingResponses  int                        `json:"awaiting_responses"`
	FullyResponded     int                        `json:"fully_responded"`
	SelectedRequests   int                        `json:"selected_requests"`
	ExpiringSoon       int                        `json:"expiring_soon"`
	RecentSpotRequests []*SpotRateRequestListItem `json:"recent_spot_requests"`
}

// CarrierRateComparisonItem represents a normalized rate option from a connected carrier (Task 5).
type CarrierRateComparisonItem struct {
	RateID             string    `json:"rate_id"`
	CarrierSCAC        string    `json:"carrier_scac"`
	CarrierName        string    `json:"carrier_name"`
	OriginPort         string    `json:"origin_port"`
	DestinationPort    string    `json:"destination_port"`
	EquipmentType      string    `json:"equipment_type"`
	ServiceCode        string    `json:"service_code,omitempty"`
	VesselName         string    `json:"vessel_name,omitempty"`
	Currency           string    `json:"currency"`
	OceanFreight       float64   `json:"ocean_freight"`
	OriginCharges      float64   `json:"origin_charges"`
	DestinationCharges float64   `json:"destination_charges"`
	TotalBuyPrice      float64   `json:"total_buy_price"`
	TransitDays        int       `json:"transit_days"`
	FreeDays           int       `json:"free_days"`
	ValidFrom          time.Time `json:"valid_from"`
	ValidUntil         time.Time `json:"valid_until"`
	IsContractRate     bool      `json:"is_contract_rate"`
	IsCheapest         bool      `json:"is_cheapest"`
	IsFastest          bool      `json:"is_fastest"`
	IsBestValue        bool      `json:"is_best_value"`
}

// CarrierRateSearchResponse represents the result of querying live carrier rates (Task 5).
type CarrierRateSearchResponse struct {
	Success         bool                        `json:"success"`
	Message         string                      `json:"message"`
	OriginPort      string                      `json:"origin_port"`
	DestinationPort string                      `json:"destination_port"`
	EquipmentType   string                      `json:"equipment_type"`
	Rates           []CarrierRateComparisonItem `json:"rates"`
	TotalRatesCount int                         `json:"total_rates_count"`
	CarriersQueried int                         `json:"carriers_queried"`
	CheapestCarrier *string                     `json:"cheapest_carrier,omitempty"`
	CheapestAmount  *float64                    `json:"cheapest_amount,omitempty"`
	FastestCarrier  *string                     `json:"fastest_carrier,omitempty"`
	FastestTransit  *int                        `json:"fastest_transit_days,omitempty"`
	SearchedAt      time.Time                   `json:"searched_at"`
}


