package spec

import "time"

// RateQuery is input for canonical / intelligence rate search
type RateQuery struct {
	OrgID           int64      `json:"-"`
	OriginPort      string     `json:"origin"`
	DestinationPort string     `json:"destination"`
	EquipmentType   string     `json:"equipment_type"`
	TargetDate      *time.Time `json:"target_date,omitempty"`
	CarrierSCACs    []string   `json:"carrier_scacs,omitempty"`
	Sources         []string   `json:"sources,omitempty"`
	MaxResults      int        `json:"max_results,omitempty"`
	Incoterms       string     `json:"incoterms,omitempty"`
}

// ListRatesRequest is the filter input for listing rate cards
type ListRatesRequest struct {
	OrgID         int64      `json:"-"`
	Search        string     `json:"search"`
	Status        string     `json:"status"`
	RateType      string     `json:"rate_type"`
	TransportMode string     `json:"transport_mode"`
	ServiceType   string     `json:"service_type"`
	EquipmentType string     `json:"equipment_type"`
	CarrierName   string     `json:"carrier_name"`
	Origin        string     `json:"origin"`
	Destination   string     `json:"destination"`
	ValidDate     *time.Time `json:"valid_date,omitempty"`
	Page          int        `json:"page"`
	Limit         int        `json:"limit"`
	SortBy        string     `json:"sort_by"`
	SortOrder     string     `json:"sort_order"`
}

// GetRateRequest payload for fetching rate by ID
type GetRateRequest struct {
	OrgID int64  `json:"-"`
	ID    string `json:"id"`
}

// CreateRateRequest is payload for POST /api/v1/rates
type CreateRateRequest struct {
	OrgID             int64   `json:"-"`
	Author            string  `json:"-"`
	RateReference     string  `json:"rate_reference"`
	CarrierName       string  `json:"carrier_name"`
	CarrierCode       *string `json:"carrier_code,omitempty"`
	ServiceProvider   *string `json:"service_provider,omitempty"`
	RateType          string  `json:"rate_type"`
	TransportMode     string  `json:"transport_mode"`
	ServiceType       string  `json:"service_type"`
	EquipmentType     *string `json:"equipment_type,omitempty"`
	OriginPort        string  `json:"origin_port"`
	OriginCode        *string `json:"origin_code,omitempty"`
	DestinationPort   string  `json:"destination_port"`
	DestinationCode   *string `json:"destination_code,omitempty"`
	Currency          string  `json:"currency"`
	BaseAmount        float64 `json:"base_amount"`
	EffectiveDate     string  `json:"effective_date"` // YYYY-MM-DD
	ExpiryDate        string  `json:"expiry_date"`    // YYYY-MM-DD
	CarrierReference  *string `json:"carrier_reference,omitempty"`
	ContractReference *string `json:"contract_reference,omitempty"`
	Notes             *string `json:"notes,omitempty"`
}

// UpdateRateRequest is payload for PUT /api/v1/rates/{id}
type UpdateRateRequest struct {
	OrgID             int64    `json:"-"`
	ID                int64    `json:"id"`
	Updater           string   `json:"-"`
	CarrierName       *string  `json:"carrier_name,omitempty"`
	CarrierCode       *string  `json:"carrier_code,omitempty"`
	ServiceProvider   *string  `json:"service_provider,omitempty"`
	RateType          *string  `json:"rate_type,omitempty"`
	TransportMode     *string  `json:"transport_mode,omitempty"`
	ServiceType       *string  `json:"service_type,omitempty"`
	EquipmentType     *string  `json:"equipment_type,omitempty"`
	OriginPort        *string  `json:"origin_port,omitempty"`
	OriginCode        *string  `json:"origin_code,omitempty"`
	DestinationPort   *string  `json:"destination_port,omitempty"`
	DestinationCode   *string  `json:"destination_code,omitempty"`
	Currency          *string  `json:"currency,omitempty"`
	BaseAmount        *float64 `json:"base_amount,omitempty"`
	EffectiveDate     *string  `json:"effective_date,omitempty"` // YYYY-MM-DD
	ExpiryDate        *string  `json:"expiry_date,omitempty"`    // YYYY-MM-DD
	Status            *string  `json:"status,omitempty"`
	CarrierReference  *string  `json:"carrier_reference,omitempty"`
	ContractReference *string  `json:"contract_reference,omitempty"`
	Notes             *string  `json:"notes,omitempty"`
}

// ArchiveRateRequest is payload for POST /api/v1/rates/{id}/archive
type ArchiveRateRequest struct {
	OrgID int64  `json:"-"`
	ID    int64  `json:"id"`
	User  string `json:"-"`
}

// RefreshSpotRatesRequest is payload for POST /api/v1/rates/spot/refresh
type RefreshSpotRatesRequest struct {
	OrgID         int64  `json:"-"`
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	EquipmentType string `json:"equipment_type"`
}

// ── Task 19.2: Rate Charge Requests ──────────────────────────────────────────

// CreateRateChargeRequest is payload for POST /api/v1/rates/{id}/charges
type CreateRateChargeRequest struct {
	OrgID              int64    `json:"-"`
	RateID             int64    `json:"rate_id"`
	ChargeCategory     string   `json:"charge_category"`
	ChargeCode         string   `json:"charge_code"`
	ChargeName         string   `json:"charge_name"`
	CalculationBasis   string   `json:"calculation_basis"`
	Quantity           float64  `json:"quantity"`
	UnitPrice          float64  `json:"unit_price"`
	Currency           string   `json:"currency"`
	MinimumAmount      *float64 `json:"minimum_amount,omitempty"`
	MaximumAmount      *float64 `json:"maximum_amount,omitempty"`
	IncludedInBaseRate bool     `json:"included_in_base_rate"`
	DisplayOrder       int      `json:"display_order"`
	Notes              *string  `json:"notes,omitempty"`
}

// UpdateRateChargeRequest is payload for PUT /api/v1/rates/{id}/charges/{chargeId}
type UpdateRateChargeRequest struct {
	OrgID              int64    `json:"-"`
	RateID             int64    `json:"rate_id"`
	ChargeID           int64    `json:"charge_id"`
	ChargeCategory     *string  `json:"charge_category,omitempty"`
	ChargeCode         *string  `json:"charge_code,omitempty"`
	ChargeName         *string  `json:"charge_name,omitempty"`
	CalculationBasis   *string  `json:"calculation_basis,omitempty"`
	Quantity           *float64 `json:"quantity,omitempty"`
	UnitPrice          *float64 `json:"unit_price,omitempty"`
	Currency           *string  `json:"currency,omitempty"`
	MinimumAmount      *float64 `json:"minimum_amount,omitempty"`
	MaximumAmount      *float64 `json:"maximum_amount,omitempty"`
	IncludedInBaseRate *bool    `json:"included_in_base_rate,omitempty"`
	DisplayOrder       *int     `json:"display_order,omitempty"`
	Notes              *string  `json:"notes,omitempty"`
}

// DeleteRateChargeRequest is payload for DELETE /api/v1/rates/{id}/charges/{chargeId}
type DeleteRateChargeRequest struct {
	OrgID    int64 `json:"-"`
	RateID   int64 `json:"rate_id"`
	ChargeID int64 `json:"charge_id"`
}

// ReorderRateChargesRequest is payload for POST /api/v1/rates/{id}/charges/reorder
type ReorderRateChargesRequest struct {
	OrgID     int64   `json:"-"`
	RateID    int64   `json:"rate_id"`
	ChargeIDs []int64 `json:"charge_ids"`
}

// ── Task 19.3: Rate Contract & Versioning Requests ────────────────────────────

// CreateRateContractRequest is payload for POST /api/v1/rates/contracts
type CreateRateContractRequest struct {
	OrgID             int64   `json:"-"`
	Author            string  `json:"-"`
	ContractReference string  `json:"contract_reference"`
	CarrierName       string  `json:"carrier_name"`
	CarrierCode       *string `json:"carrier_code,omitempty"`
	ContractName      string  `json:"contract_name"`
	ContractType      string  `json:"contract_type"`
	TransportMode     *string `json:"transport_mode,omitempty"`
	Currency          *string `json:"currency,omitempty"`
	EffectiveDate     string  `json:"effective_date"` // YYYY-MM-DD
	ExpiryDate        string  `json:"expiry_date"`    // YYYY-MM-DD
	RenewalOwner      *string `json:"renewal_owner,omitempty"`
	Notes             *string `json:"notes,omitempty"`
}

// UpdateRateContractRequest is payload for PUT /api/v1/rates/contracts/{id}
type UpdateRateContractRequest struct {
	OrgID             int64   `json:"-"`
	ID                int64   `json:"id"`
	Updater           string  `json:"-"`
	ContractReference *string `json:"contract_reference,omitempty"`
	CarrierName       *string `json:"carrier_name,omitempty"`
	CarrierCode       *string `json:"carrier_code,omitempty"`
	ContractName      *string `json:"contract_name,omitempty"`
	ContractType      *string `json:"contract_type,omitempty"`
	TransportMode     *string `json:"transport_mode,omitempty"`
	Currency          *string `json:"currency,omitempty"`
	EffectiveDate     *string `json:"effective_date,omitempty"` // YYYY-MM-DD
	ExpiryDate        *string `json:"expiry_date,omitempty"`    // YYYY-MM-DD
	Status            *string `json:"status,omitempty"`
	RenewalStatus     *string `json:"renewal_status,omitempty"`
	RenewalOwner      *string `json:"renewal_owner,omitempty"`
	Notes             *string `json:"notes,omitempty"`
}

// RenewRateContractRequest is payload for POST /api/v1/rates/contracts/{id}/renew
type RenewRateContractRequest struct {
	OrgID         int64   `json:"-"`
	ID            int64   `json:"id"`
	User          string  `json:"-"`
	NewExpiryDate string  `json:"new_expiry_date"` // YYYY-MM-DD
	RenewalStatus string  `json:"renewal_status"`  // RENEWED, IN_PROGRESS
	Notes         *string `json:"notes,omitempty"`
}

// ListRateContractsRequest is filter query for GET /api/v1/rates/contracts
type ListRateContractsRequest struct {
	OrgID         int64  `json:"-"`
	Search        string `json:"search"`
	CarrierName   string `json:"carrier_name"`
	ContractType  string `json:"contract_type"`
	Status        string `json:"status"`
	RenewalStatus string `json:"renewal_status"`
	TransportMode string `json:"transport_mode"`
	Page          int    `json:"page"`
	Limit         int    `json:"limit"`
	SortBy        string `json:"sort_by"`
	SortOrder     string `json:"sort_order"`
}

// CreateRateVersionRequest is payload for POST /api/v1/rates/{id}/versions
type CreateRateVersionRequest struct {
	OrgID             int64                  `json:"-"`
	RateID            int64                  `json:"rate_id"`
	User              string                 `json:"-"`
	BaseAmount        *float64               `json:"base_amount,omitempty"`
	Currency          *string                `json:"currency,omitempty"`
	EffectiveDate     *string                `json:"effective_date,omitempty"`
	ExpiryDate        *string                `json:"expiry_date,omitempty"`
	CarrierReference  *string                `json:"carrier_reference,omitempty"`
	ContractReference *string                `json:"contract_reference,omitempty"`
	Notes             *string                `json:"notes,omitempty"`
	ChargeUpdates     []CreateRateChargeRequest `json:"charge_updates,omitempty"`
	RevisionReason    string                 `json:"revision_reason"`
}

// ── Task 19.4: Spot Rate Requests & Responses Request DTOs ───────────────────

// CreateSpotRateRequestRequest is payload for POST /api/v1/rates/spot-requests
type CreateSpotRateRequestRequest struct {
	OrgID             int64    `json:"-"`
	User              string   `json:"-"`
	RequestReference  string   `json:"request_reference"`
	CustomerID        *int64   `json:"customer_id,omitempty"`
	CustomerName      *string  `json:"customer_name,omitempty"`
	OriginPort        string   `json:"origin_port"`
	OriginCode        *string  `json:"origin_code,omitempty"`
	DestinationPort   string   `json:"destination_port"`
	DestinationCode   *string  `json:"destination_code,omitempty"`
	TransportMode     string   `json:"transport_mode"`
	ServiceType       string   `json:"service_type"`
	EquipmentType     *string  `json:"equipment_type,omitempty"`
	Commodity         *string  `json:"commodity,omitempty"`
	CargoWeight       *float64 `json:"cargo_weight,omitempty"`
	CargoVolume       *float64 `json:"cargo_volume,omitempty"`
	ContainerQuantity int      `json:"container_quantity"`
	ReadyDate         string   `json:"ready_date"` // YYYY-MM-DD
	TargetCurrency    string   `json:"target_currency"`
	RequiredByDate    string   `json:"required_by_date"` // YYYY-MM-DD
	Notes             *string  `json:"notes,omitempty"`
}

// UpdateSpotRateRequestRequest is payload for PUT /api/v1/rates/spot-requests/{id}
type UpdateSpotRateRequestRequest struct {
	OrgID             int64    `json:"-"`
	ID                int64    `json:"id"`
	User              string   `json:"-"`
	CustomerName      *string  `json:"customer_name,omitempty"`
	TransportMode     *string  `json:"transport_mode,omitempty"`
	ServiceType       *string  `json:"service_type,omitempty"`
	EquipmentType     *string  `json:"equipment_type,omitempty"`
	Commodity         *string  `json:"commodity,omitempty"`
	CargoWeight       *float64 `json:"cargo_weight,omitempty"`
	CargoVolume       *float64 `json:"cargo_volume,omitempty"`
	ContainerQuantity *int     `json:"container_quantity,omitempty"`
	ReadyDate         *string  `json:"ready_date,omitempty"`
	TargetCurrency    *string  `json:"target_currency,omitempty"`
	RequiredByDate    *string  `json:"required_by_date,omitempty"`
	Status            *string  `json:"status,omitempty"`
	Notes             *string  `json:"notes,omitempty"`
}

// ListSpotRateRequestsRequest is filter query for GET /api/v1/rates/spot-requests
type ListSpotRateRequestsRequest struct {
	OrgID         int64  `json:"-"`
	Search        string `json:"search"`
	Status        string `json:"status"`
	TransportMode string `json:"transport_mode"`
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	Page          int    `json:"page"`
	Limit         int    `json:"limit"`
	SortBy        string `json:"sort_by"`
	SortOrder     string `json:"sort_order"`
}

// CreateSpotRateResponseChargeRequest is itemized charge inside a carrier response
type CreateSpotRateResponseChargeRequest struct {
	ChargeCategory   string  `json:"charge_category"`
	ChargeName       string  `json:"charge_name"`
	CalculationBasis string  `json:"calculation_basis"`
	Quantity         float64 `json:"quantity"`
	UnitPrice        float64 `json:"unit_price"`
	Currency         string  `json:"currency"`
	DisplayOrder     int     `json:"display_order"`
}

// CreateSpotRateResponseRequest is payload for POST /api/v1/rates/spot-requests/{id}/responses
type CreateSpotRateResponseRequest struct {
	OrgID                int64                                 `json:"-"`
	SpotRateRequestID    int64                                 `json:"spot_rate_request_id"`
	User                 string                                `json:"-"`
	CarrierName          string                                `json:"carrier_name"`
	CarrierCode          *string                               `json:"carrier_code,omitempty"`
	SupplierName         *string                               `json:"supplier_name,omitempty"`
	RateID               *int64                                `json:"rate_id,omitempty"`
	Currency             string                                `json:"currency"`
	BaseAmount           float64                               `json:"base_amount"`
	TransitDays          *int                                  `json:"transit_days,omitempty"`
	FreeDaysOrigin       int                                   `json:"free_days_origin"`
	FreeDaysDestination  int                                   `json:"free_days_destination"`
	ValidFrom            string                                `json:"valid_from"`  // YYYY-MM-DD
	ValidUntil           string                                `json:"valid_until"` // YYYY-MM-DD
	RoutingNotes         *string                               `json:"routing_notes,omitempty"`
	ResponseNotes        *string                               `json:"response_notes,omitempty"`
	Charges              []CreateSpotRateResponseChargeRequest `json:"charges,omitempty"`
}

// UpdateSpotRateResponseRequest is payload for PUT /api/v1/rates/spot-requests/{id}/responses/{responseId}
type UpdateSpotRateResponseRequest struct {
	OrgID                int64                                 `json:"-"`
	SpotRateRequestID    int64                                 `json:"spot_rate_request_id"`
	ResponseID           int64                                 `json:"response_id"`
	User                 string                                `json:"-"`
	CarrierName          *string                               `json:"carrier_name,omitempty"`
	CarrierCode          *string                               `json:"carrier_code,omitempty"`
	SupplierName         *string                               `json:"supplier_name,omitempty"`
	Currency             *string                               `json:"currency,omitempty"`
	BaseAmount           *float64                              `json:"base_amount,omitempty"`
	TransitDays          *int                                  `json:"transit_days,omitempty"`
	FreeDaysOrigin       *int                                  `json:"free_days_origin,omitempty"`
	FreeDaysDestination  *int                                  `json:"free_days_destination,omitempty"`
	ValidFrom            *string                               `json:"valid_from,omitempty"`
	ValidUntil           *string                               `json:"valid_until,omitempty"`
	Status               *string                               `json:"status,omitempty"`
	RoutingNotes         *string                               `json:"routing_notes,omitempty"`
	ResponseNotes        *string                               `json:"response_notes,omitempty"`
	Charges              []CreateSpotRateResponseChargeRequest `json:"charges,omitempty"`
}

// SelectPreferredSpotRateRequest is payload for POST /api/v1/rates/spot-requests/{id}/responses/{responseId}/select
type SelectPreferredSpotRateRequest struct {
	OrgID             int64   `json:"-"`
	SpotRateRequestID int64   `json:"spot_rate_request_id"`
	ResponseID        int64   `json:"response_id"`
	User              string  `json:"-"`
	SelectionNotes    *string `json:"selection_notes,omitempty"`
}

// CarrierRateSearchRequest is payload for live multi-carrier rate search (Task 5).
type CarrierRateSearchRequest struct {
	OrgID           int64   `json:"-"`
	OriginPort      string  `json:"origin_port"`                 // UN/LOCODE e.g. "INNSA"
	DestinationPort string  `json:"destination_port"`            // UN/LOCODE e.g. "NLRTM"
	EquipmentType   string  `json:"equipment_type"`              // e.g. "40HC", "20GP"
	CarrierSCAC     string  `json:"carrier_scac,omitempty"`      // optional filter for specific carrier
	Commodity       string  `json:"commodity,omitempty"`
	CargoReadyDate  string  `json:"cargo_ready_date,omitempty"`  // YYYY-MM-DD
	ContractNumber  string  `json:"contract_number,omitempty"`
	RateType        string  `json:"rate_type,omitempty"`         // "ALL", "SPOT", "CONTRACT"
	Currency        string  `json:"currency,omitempty"`
}

