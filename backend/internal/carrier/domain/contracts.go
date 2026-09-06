package domain

import (
	"time"
)

// TrackingRequest contains identifiers used to query a carrier for operational tracking events.
type TrackingRequest struct {
	BookingNumber   string `json:"booking_number,omitempty"`
	ContainerNumber string `json:"container_number,omitempty"`
	MBLNumber       string `json:"mbl_number,omitempty"`
	CarrierSCAC     string `json:"carrier_scac"`
}

// NormalizedTrackingEvent represents a canonical telemetry milestone returned by a carrier adapter.
type NormalizedTrackingEvent struct {
	EventID         string    `json:"event_id"`
	MilestoneCode   string    `json:"milestone_code"` // e.g., "GATE_IN", "LOADED", "DEPARTED", "ARRIVED", "DISCHARGED", "GATE_OUT", "DELIVERED"
	EventTime       time.Time `json:"event_time"`
	Location        string    `json:"location"`
	VesselName      string    `json:"vessel_name,omitempty"`
	VoyageNumber    string    `json:"voyage_number,omitempty"`
	ContainerNumber string    `json:"container_number,omitempty"`
	Description     string    `json:"description"`
	RawPayload      []byte    `json:"raw_payload,omitempty"`
}

// NormalizedTrackingResult represents the aggregated tracking status returned by an adapter.
type NormalizedTrackingResult struct {
	CarrierSCAC       string                    `json:"carrier_scac"`
	BookingNumber     string                    `json:"booking_number,omitempty"`
	ContainerNumber   string                    `json:"container_number,omitempty"`
	CurrentStatus     string                    `json:"current_status"`
	LatestLocation    string                    `json:"latest_location,omitempty"`
	EstimatedArrival  *time.Time                `json:"estimated_arrival,omitempty"`
	ActualDeparture   *time.Time                `json:"actual_departure,omitempty"`
	Events            []NormalizedTrackingEvent `json:"events"`
	IsDelivered       bool                      `json:"is_delivered"`
	FetchedAt         time.Time                 `json:"fetched_at"`
}

// RateRequest represents parameters to query carrier pricing engine.
type RateRequest struct {
	OriginPort      string  `json:"origin_port"`      // UN/LOCODE e.g. "INNSA"
	DestinationPort string  `json:"destination_port"` // UN/LOCODE e.g. "NLRTM"
	EquipmentType   string  `json:"equipment_type"`   // e.g. "20GP", "40HC"
	Commodity       string  `json:"commodity,omitempty"`
	GrossWeightKG   float64 `json:"gross_weight_kg,omitempty"`
	VolumeCBM       float64 `json:"volume_cbm,omitempty"`
	Incoterms       string  `json:"incoterms,omitempty"`
	ValidDate       time.Time `json:"valid_date"`
}

// ContractRateRequest represents parameters to query contracted carrier rates.
type ContractRateRequest struct {
	RateRequest
	ContractNumber string `json:"contract_number,omitempty"`
	AccountID      string `json:"account_id,omitempty"`
}

// SpotRateRequest represents parameters to query live carrier spot pricing.
type SpotRateRequest struct {
	RateRequest
	CargoReadyDate time.Time `json:"cargo_ready_date"`
}

// NormalizedCarrierRate represents a canonical rate quotation returned by a carrier adapter.
type NormalizedCarrierRate struct {
	RateID                string    `json:"rate_id"`
	CarrierSCAC           string    `json:"carrier_scac"`
	CarrierName           string    `json:"carrier_name"`
	OriginPort            string    `json:"origin_port"`
	DestinationPort       string    `json:"destination_port"`
	ViaPort               string    `json:"via_port,omitempty"`
	ServiceCode           string    `json:"service_code,omitempty"`
	VesselName            string    `json:"vessel_name,omitempty"`
	EquipmentType         string    `json:"equipment_type"`
	Currency              string    `json:"currency"`
	OceanFreight          float64   `json:"ocean_freight"`
	OriginCharges         float64   `json:"origin_charges"`
	DestinationCharges    float64   `json:"destination_charges"`
	TotalBuyPrice         float64   `json:"total_buy_price"`
	TransitDays           int       `json:"transit_days"`
	FreeDays              int       `json:"free_days"`
	ValidFrom             time.Time `json:"valid_from"`
	ValidUntil            time.Time `json:"valid_until"`
	CO2Emissions          float64   `json:"co2_emissions,omitempty"`
	HistoricalSuccessRate float64   `json:"historical_success_rate,omitempty"`
	IsContractRate        bool      `json:"is_contract_rate"`
}

// BookingRequest represents a structured request to reserve carrier space.
type BookingRequest struct {
	CarrierSCAC       string    `json:"carrier_scac"`
	ContractNumber    string    `json:"contract_number,omitempty"`
	OriginPort        string    `json:"origin_port"`
	DestinationPort   string    `json:"destination_port"`
	EquipmentType     string    `json:"equipment_type"`
	Quantity          int       `json:"quantity"`
	CargoReadyDate    time.Time `json:"cargo_ready_date"`
	Commodity         string    `json:"commodity"`
	ShipperName       string    `json:"shipper_name"`
	ConsigneeName     string    `json:"consignee_name"`
	CustomerReference string    `json:"customer_reference,omitempty"`
}

// NormalizedBookingResult holds carrier booking response and allocation state.
type NormalizedBookingResult struct {
	BookingNumber        string     `json:"booking_number"`
	CarrierSCAC          string     `json:"carrier_scac"`
	Status               string     `json:"status"` // e.g. "CONFIRMED", "PENDING_ALLOCATION", "REJECTED"
	ConfirmationRef      string     `json:"confirmation_ref,omitempty"`
	VesselName           string     `json:"vessel_name,omitempty"`
	VoyageNumber         string     `json:"voyage_number,omitempty"`
	ETD                  *time.Time `json:"etd,omitempty"`
	ETA                  *time.Time `json:"eta,omitempty"`
	CutOffDate           *time.Time `json:"cut_off_date,omitempty"`
	VGMDeadline          *time.Time `json:"vgm_deadline,omitempty"`
	EmptyPickupLocation  string     `json:"empty_pickup_location,omitempty"`
	AllocatedContainers  []string   `json:"allocated_containers,omitempty"`
}

// DocumentRequest represents a query to fetch carrier-issued transport documents.
type DocumentRequest struct {
	CarrierSCAC   string `json:"carrier_scac"`
	BookingNumber string `json:"booking_number"`
	DocumentType  string `json:"document_type,omitempty"` // e.g., "BILL_OF_LADING", "ARRIVAL_NOTICE", "BOOKING_CONFIRMATION"
}

// NormalizedDocumentResult holds a transport document provided by a carrier API.
type NormalizedDocumentResult struct {
	DocumentID   string    `json:"document_id"`
	CarrierSCAC  string    `json:"carrier_scac"`
	DocumentType string    `json:"document_type"`
	Reference    string    `json:"reference"`
	DocumentURL  string    `json:"document_url"`
	MimeType     string    `json:"mime_type"`
	CreatedAt    time.Time `json:"created_at"`
}

// TestConnectionResult represents the response from testing an integration adapter.
type TestConnectionResult struct {
	Success            bool         `json:"success"`
	Message            string       `json:"message"`
	LatencyMs          int64        `json:"latency_ms"`
	TestedCapabilities []Capability `json:"tested_capabilities"`
	TestedEnvironment  Environment  `json:"tested_environment"`
	ErrorCode          string       `json:"error_code,omitempty"`
	HTTPStatus         int          `json:"http_status,omitempty"`
}
