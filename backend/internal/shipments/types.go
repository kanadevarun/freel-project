package shipments

import (
	"encoding/json"
	"time"
)

type Shipment struct {
	ID               int64      `db:"id" json:"id"`
	OrgID            int64      `db:"org_id" json:"org_id"`
	RFQID            *int64     `db:"rfq_id" json:"rfq_id,omitempty"`
	QuoteID          *int64     `db:"quote_id" json:"quote_id,omitempty"`
	CarrierSCAC      string     `db:"carrier_scac" json:"carrier_scac"`
	BookingNumber    *string    `db:"booking_number" json:"booking_number,omitempty"`
	MBLNumber        *string    `db:"mbl_number" json:"mbl_number,omitempty"`
	HBLNumber        *string    `db:"hbl_number" json:"hbl_number,omitempty"`
	ContainerNumbers []string   `db:"container_numbers" json:"container_numbers"`
	Status           string     `db:"status" json:"status"`
	OriginPort       string         `db:"origin_port" json:"origin_port"`
	DestinationPort  string         `db:"destination_port" json:"destination_port"`
	VesselName       *string        `db:"vessel_name" json:"vessel_name,omitempty"`
	VoyageNumber     *string        `db:"voyage_number" json:"voyage_number,omitempty"`
	ETD              *time.Time     `db:"etd" json:"etd,omitempty"`
	ETA              *time.Time     `db:"eta" json:"eta,omitempty"`
	CreatedAt        time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at" json:"updated_at"`
}

type ShipmentMilestone struct {
	ID            int64      `db:"id" json:"id"`
	ShipmentID    int64      `db:"shipment_id" json:"shipment_id"`
	MilestoneCode string     `db:"milestone_code" json:"milestone_code"` // BOOKED, DEPARTED, IN_TRANSIT, ARRIVED, DELIVERED
	Description   *string    `db:"description" json:"description,omitempty"`
	PlannedDate   *time.Time `db:"planned_date" json:"planned_date,omitempty"`
	ActualDate    *time.Time `db:"actual_date" json:"actual_date,omitempty"`
	Status        string     `db:"status" json:"status"` // PLANNED, COMPLETED
	Location      *string    `db:"location" json:"location,omitempty"`
	Notes         *string    `db:"notes" json:"notes,omitempty"`
	SourceEventID *string    `db:"source_event_id" json:"source_event_id,omitempty"`
}

type ShipmentException struct {
	ID            int64      `db:"id" json:"id"`
	ShipmentID    int64      `db:"shipment_id" json:"shipment_id"`
	ExceptionType string     `db:"exception_type" json:"exception_type"` // ROLLOVER, DELAY, CUSTOMS_HOLD, PORT_CONGESTION, WEATHER
	Severity      string     `db:"severity" json:"severity"`             // INFO, WARNING, CRITICAL
	Title         string     `db:"title" json:"title"`
	Description   *string    `db:"description" json:"description,omitempty"`
	Resolved      bool       `db:"resolved" json:"resolved"`
	ResolvedAt    *time.Time `db:"resolved_at" json:"resolved_at,omitempty"`
	AISummary     *string    `db:"ai_summary" json:"ai_summary,omitempty"`
	SourceEventID *string    `db:"source_event_id" json:"source_event_id,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
}

// NormalizedTrackingEvent represents a standardized tracking event used across Go and Python
type NormalizedTrackingEvent struct {
	EventID         string          `json:"event_id"`
	SourceType      string          `json:"source_type"` // API | WEBHOOK | EMAIL | MANUAL | POLLING
	CarrierSCAC     string          `json:"carrier_scac"`
	BookingNumber   string          `json:"booking_number"`
	ContainerNumber string          `json:"container_number"`
	MBLNumber       string          `json:"mbl_number"`
	HBLNumber       string          `json:"hbl_number"`
	VesselName      string          `json:"vessel_name"`
	VoyageNumber    string          `json:"voyage_number"`
	MilestoneCode   string          `json:"milestone_code"`
	EventTime       time.Time       `json:"event_time"`
	Location        string          `json:"location"`
	Description     string          `json:"description"`
	RawPayload      json.RawMessage `json:"raw_payload,omitempty"`
	ReceivedAt      time.Time       `json:"received_at"`
}

// CarrierTrackingEvent DB representation of raw carrier updates
type CarrierTrackingEvent struct {
	ID                int64           `db:"id" json:"id"`
	OrgID             int64           `db:"org_id" json:"org_id"`
	EventID           string          `db:"event_id" json:"event_id"`
	SourceType        string          `db:"source_type" json:"source_type"`
	CarrierSCAC       string          `db:"carrier_scac" json:"carrier_scac"`
	BookingNumber     string          `db:"booking_number" json:"booking_number"`
	ContainerNumber   string          `db:"container_number" json:"container_number"`
	MBLNumber         string          `db:"mbl_number" json:"mbl_number"`
	HBLNumber         string          `db:"hbl_number" json:"hbl_number"`
	VesselName        string          `db:"vessel_name" json:"vessel_name"`
	VoyageNumber      string          `db:"voyage_number" json:"voyage_number"`
	MilestoneCode     string          `db:"milestone_code" json:"milestone_code"`
	EventTime         time.Time       `db:"event_time" json:"event_time"`
	Location          string          `db:"location" json:"location"`
	RawDescription    string          `db:"raw_description" json:"raw_description"`
	RawPayload        json.RawMessage `db:"raw_payload" json:"raw_payload"`
	ShipmentID        *int64          `db:"shipment_id" json:"shipment_id,omitempty"`
	MatchingStatus    string          `db:"matching_status" json:"matching_status"`
	ProcessingStatus  string          `db:"processing_status" json:"processing_status"`
	ReceivedAt        time.Time       `db:"received_at" json:"received_at"`
	UpdatedAt         time.Time       `db:"updated_at" json:"updated_at"`
}

