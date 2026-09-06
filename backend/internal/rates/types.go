package rates

import (
	"github.com/freel/backend/internal/carrier"
	"github.com/freel/backend/internal/rates/spec"
	"github.com/jmoiron/sqlx"
)

// Re-export core types for backwards compatibility
type (
	Rate                   = spec.Rate
	RateListItem           = spec.RateListItem
	RateDetail             = spec.RateDetail
	CarrierCoverageSummary = spec.CarrierCoverageSummary
	RecentRateUpdate       = spec.RecentRateUpdate
	RateSummaryKPIs        = spec.RateSummaryKPIs
	ListRatesRequest       = spec.ListRatesRequest
	RateFilter             = spec.ListRatesRequest
	CreateRateRequest      = spec.CreateRateRequest
	UpdateRateRequest      = spec.UpdateRateRequest
	ArchiveRateRequest     = spec.ArchiveRateRequest
	RateListResponse       = spec.ListRatesResponse
	ListRatesResponse      = spec.ListRatesResponse
	RateQuery              = spec.RateQuery
	RateSearchResult       = spec.RateSearchResult
	CanonicalRate          = spec.CanonicalRate
	Surcharge              = spec.Surcharge
)

// Legacy type aliases
type RateStatus = string
type RateType = string
type RateSource = string
type ExtractionStatus = string
type SurchargeUnit = string
type Service = BusinessLogic
type Repository = Datalayer

// NewRepository wraps NewDataLayer for backward compatibility in existing tests
func NewRepository(db *sqlx.DB) Datalayer {
	return NewDataLayer(db)
}

// NewService wraps NewBusinessLogic for backward compatibility in existing tests
func NewService(dl Datalayer, normalizer SpotNormalizer, carrierSvc carrier.Service) BusinessLogic {
	return NewBusinessLogic(dl, normalizer, carrierSvc)
}

// ── Task 19.6: Rate Lifecycle Intelligence Types ──────────────────────────────

type RateLifecycleEvent struct {
	ID             int64       `db:"id"              json:"id"`
	OrgID          int64       `db:"org_id"          json:"org_id"`
	RateID         *int64      `db:"rate_id"         json:"rate_id,omitempty"`
	ContractID     *int64      `db:"contract_id"     json:"contract_id,omitempty"`
	EventType      string      `db:"event_type"      json:"event_type"`
	PreviousStatus string      `db:"previous_status" json:"previous_status,omitempty"`
	CurrentStatus  string      `db:"current_status"  json:"current_status"`
	Description    string      `db:"description"     json:"description"`
	MetadataJSON   string      `db:"metadata"        json:"-"`
	Metadata       interface{} `db:"-"               json:"metadata,omitempty"`
	CreatedAt      string      `db:"created_at"      json:"created_at"`
}

type RateLifecycleSummary struct {
	TotalRates               int `json:"total_rates"`
	ActiveRates              int `json:"active_rates"`
	ExpiringSoonRates        int `json:"expiring_soon_rates"`
	ExpiredRates             int `json:"expired_rates"`
	SupersededRates          int `json:"superseded_rates"`
	TotalContracts           int `json:"total_contracts"`
	ActiveContracts          int `json:"active_contracts"`
	ExpiringContracts        int `json:"expiring_contracts"`
	ExpiredContracts         int `json:"expired_contracts"`
	ContractsRequiringRenewal int `json:"contracts_requiring_renewal"`
	QuotationsAtRisk         int `json:"quotations_at_risk"`
}

type RateAttentionItem struct {
	RateID             int64   `db:"rate_id"             json:"rate_id"`
	CarrierName        string  `db:"carrier_name"        json:"carrier_name"`
	CarrierCode        string  `db:"carrier_code"        json:"carrier_code"`
	RateType           string  `db:"rate_type"           json:"rate_type"`
	VersionNumber      int     `db:"version_number"      json:"version_number"`
	Origin             string  `db:"origin_port"         json:"origin"`
	Destination        string  `db:"destination_port"    json:"destination"`
	TransportMode      string  `db:"transport_mode"      json:"transport_mode"`
	EquipmentType      string  `db:"equipment_type"      json:"equipment_type"`
	Currency           string  `db:"currency"            json:"currency"`
	BaseAmount         float64 `db:"base_amount"         json:"base_amount"`
	ValidFrom          *string `db:"valid_from"          json:"valid_from,omitempty"`
	ValidUntil         *string `db:"valid_until"         json:"valid_until,omitempty"`
	Status             string  `db:"status"              json:"status"`
	DaysRemaining      int     `db:"days_remaining"      json:"days_remaining"`
	AttentionBucket    string  `db:"attention_bucket"    json:"attention_bucket"` // 'EXPIRED', 'EXPIRING_7D', 'EXPIRING_30D', 'SUPERSEDED'
	AffectedQuotesCount int    `db:"affected_quotes"     json:"affected_quotes_count"`
	ContractCode       string  `db:"contract_code"       json:"contract_code,omitempty"`
}

type ContractAttentionItem struct {
	ContractID         int64   `db:"contract_id"         json:"contract_id"`
	CarrierName        string  `db:"carrier_name"        json:"carrier_name"`
	ContractCode       string  `db:"contract_code"       json:"contract_code"`
	Title              string  `db:"contract_title"      json:"title"`
	StartDate          *string `db:"start_date"          json:"start_date,omitempty"`
	EndDate            *string `db:"end_date"            json:"end_date,omitempty"`
	Status             string  `db:"status"              json:"status"`
	RenewalStatus      string  `db:"renewal_status"      json:"renewal_status"`
	DaysRemaining      int     `db:"days_remaining"      json:"days_remaining"`
	LinkedRatesCount   int     `db:"linked_rates_count"  json:"linked_rates_count"`
	AffectedQuotesCount int    `db:"affected_quotes"     json:"affected_quotes_count"`
}


// ── Task 19.7: Rate Analytics & Procurement Intelligence Types ─────────────────

// CurrencyValue is a currency-safe amount pair. Analytics never sum across currencies.
type CurrencyValue struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

// RateAnalyticsOverview provides 17 management-level KPIs across all rate management domains.
type RateAnalyticsOverview struct {
	// Rate counts
	TotalRates        int `db:"total_rates"         json:"total_rates"`
	ActiveRates       int `db:"active_rates"        json:"active_rates"`
	ExpiringSoonRates int `db:"expiring_soon_rates" json:"expiring_soon_rates"`
	ExpiredRates      int `db:"expired_rates"       json:"expired_rates"`
	SupersededRates   int `db:"superseded_rates"    json:"superseded_rates"`
	DraftRates        int `db:"draft_rates"         json:"draft_rates"`

	// Contract counts
	TotalContracts            int `db:"total_contracts"              json:"total_contracts"`
	ActiveContracts           int `db:"active_contracts"             json:"active_contracts"`
	ContractsRequiringRenewal int `db:"contracts_requiring_renewal"  json:"contracts_requiring_renewal"`

	// Coverage
	TotalLanesCovered int `db:"total_lanes_covered" json:"total_lanes_covered"`
	TotalCarriers     int `db:"total_carriers"      json:"total_carriers"`

	// Spot sourcing
	TotalSpotRequests    int `db:"total_spot_requests"     json:"total_spot_requests"`
	SpotRequestsResponded int `db:"spot_requests_responded" json:"spot_requests_responded"`
	SpotRequestsSelected  int `db:"spot_requests_selected"  json:"spot_requests_selected"`
	SpotRequestsExpired   int `db:"spot_requests_expired"   json:"spot_requests_expired"`

	// Quotation integration
	QuoteToRateSelectionCount    int `db:"quote_rate_selection_count"   json:"quote_to_rate_selection_count"`
	QuotationCommercialRiskCount int `db:"quotation_risk_count"         json:"quotation_commercial_risk_count"`
}

// RateTrendDataPoint is a single-day snapshot of rate activity for time-series charts.
type RateTrendDataPoint struct {
	Date               string `db:"date"                 json:"date"`
	RatesCreated       int    `db:"rates_created"        json:"rates_created"`
	RatesActivated     int    `db:"rates_activated"      json:"rates_activated"`
	RatesExpired       int    `db:"rates_expired"        json:"rates_expired"`
	RatesSuperseded    int    `db:"rates_superseded"     json:"rates_superseded"`
	ContractsCreated   int    `db:"contracts_created"    json:"contracts_created"`
	SpotRequestsCreated int   `db:"spot_requests_created" json:"spot_requests_created"`
	SpotRequestsSelected int  `db:"spot_requests_selected" json:"spot_requests_selected"`
}

// CarrierRatePerformance captures per-carrier analytics for the leaderboard.
type CarrierRatePerformance struct {
	CarrierName        string  `db:"carrier_name"          json:"carrier_name"`
	CarrierCode        string  `db:"carrier_code"          json:"carrier_code"`
	TotalRates         int     `db:"total_rates"           json:"total_rates"`
	ActiveRates        int     `db:"active_rates"          json:"active_rates"`
	ExpiringRates      int     `db:"expiring_rates"        json:"expiring_rates"`
	ExpiredRates       int     `db:"expired_rates"         json:"expired_rates"`
	LanesCovered       int     `db:"lanes_covered"         json:"lanes_covered"`
	ContractsCount     int     `db:"contracts_count"       json:"contracts_count"`
	SpotResponsesCount int     `db:"spot_responses_count"  json:"spot_responses_count"`
	SpotSelections     int     `db:"spot_selections"       json:"spot_selections"`
	SelectionRate      float64 `db:"-"                     json:"selection_rate"`      // computed, not stored
	AverageTransitDays float64 `db:"avg_transit_days"      json:"average_transit_days"`
	// RateHealthStatus is one of: HEALTHY, ATTENTION, CRITICAL
	RateHealthStatus string `db:"-" json:"rate_health_status"`
}

// LaneRatePerformance captures per-lane analytics with currency-safe price grouping.
type LaneRatePerformance struct {
	Origin          string          `db:"origin_port"       json:"origin"`
	Destination     string          `db:"destination_port"  json:"destination"`
	TransportMode   string          `db:"transport_mode"    json:"transport_mode"`
	ServiceType     string          `db:"service_type"      json:"service_type"`
	EquipmentType   string          `db:"equipment_type"    json:"equipment_type"`
	AvailableRates  int             `db:"available_rates"   json:"available_rates"`
	ActiveRates     int             `db:"active_rates"      json:"active_rates"`
	CarrierCount    int             `db:"carrier_count"     json:"carrier_count"`
	SpotRequestCount int            `db:"spot_request_count" json:"spot_request_count"`
	SelectedRateCount int           `db:"selected_rate_count" json:"selected_rate_count"`
	// CurrencyBreakdown groups min/avg/max prices per currency — never mixed.
	CurrencyBreakdown []LaneCurrencyBreakdown `db:"-" json:"currency_breakdown"`
	// CoverageStatus: COVERED, LIMITED, UNCOVERED
	CoverageStatus string `db:"-" json:"coverage_status"`
}

// LaneCurrencyBreakdown holds min/avg/max for a single currency on a lane.
type LaneCurrencyBreakdown struct {
	Currency    string  `db:"currency"     json:"currency"`
	CheapestRate float64 `db:"cheapest_rate" json:"cheapest_rate"`
	AverageRate  float64 `db:"average_rate"  json:"average_rate"`
	HighestRate  float64 `db:"highest_rate"  json:"highest_rate"`
}

// RateLifecycleAnalytics is a status-distribution count breakdown.
type RateLifecycleAnalytics struct {
	Active                    int `db:"active_count"          json:"active"`
	ExpiringSoon              int `db:"expiring_soon_count"   json:"expiring_soon"`
	Expired                   int `db:"expired_count"         json:"expired"`
	Superseded                int `db:"superseded_count"      json:"superseded"`
	Archived                  int `db:"archived_count"        json:"archived"`
	Draft                     int `db:"draft_count"           json:"draft"`
	TotalRates                int `db:"-"                     json:"total_rates"`
	ContractRenewalRequired   int `db:"contract_renewal_count" json:"contract_renewal_required"`
	CommercialRiskEvents      int `db:"risk_events_count"     json:"commercial_risk_events"`
}

// SpotSourcingPerformance captures the spot request-to-selection funnel.
type SpotSourcingPerformance struct {
	TotalRequests             int     `db:"total_requests"              json:"total_requests"`
	AwaitingResponses         int     `db:"awaiting_responses"          json:"awaiting_responses"`
	FullyResponded            int     `db:"fully_responded"             json:"fully_responded"`
	Selected                  int     `db:"selected"                    json:"selected"`
	Expired                   int     `db:"expired"                     json:"expired"`
	Cancelled                 int     `db:"cancelled"                   json:"cancelled"`
	AverageResponsesPerRequest float64 `db:"avg_responses_per_request"   json:"average_responses_per_request"`
	SelectionRate             float64 `db:"-"                           json:"selection_rate"`  // computed
	ResponseRate              float64 `db:"-"                           json:"response_rate"`   // computed
	// CarrierParticipation top 5 spot responders
	CarrierParticipation []SpotCarrierParticipation `db:"-" json:"carrier_participation"`
}

// SpotCarrierParticipation summarises a carrier's spot market activity.
type SpotCarrierParticipation struct {
	CarrierName    string `db:"carrier_name"     json:"carrier_name"`
	ResponsesCount int    `db:"responses_count"  json:"responses_count"`
	SelectionsCount int   `db:"selections_count" json:"selections_count"`
}

// CommercialImpactInsight is a deterministic, rule-based intelligence card.
type CommercialImpactInsight struct {
	Category          string `json:"category"`
	Severity          string `json:"severity"` // INFO, WARNING, CRITICAL, SUCCESS
	Headline          string `json:"headline"`
	Description       string `json:"description"`
	MetricValue       string `json:"metric_value"`
	RecommendedAction string `json:"recommended_action"`
	RelatedEntityType string `json:"related_entity_type,omitempty"` // "rate", "contract", "lane", "spot_request", "quotation"
	RelatedEntityID   int64  `json:"related_entity_id,omitempty"`
}

