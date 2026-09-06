package domain

import (
	"time"
)

// DCSA Event Types
const (
	DCSAEventShipment  = "SHIPMENT"
	DCSAEventTransport = "TRANSPORT"
	DCSAEventEquipment = "EQUIPMENT"
)

// DCSA Event Classifiers
const (
	DCSAClassifierActual    = "ACT" // Actual event
	DCSAClassifierEstimated = "EST" // Estimated future event
	DCSAClassifierPlanned   = "PLN" // Planned schedule event
	DCSAClassifierRequested = "REQ" // Requested event
)

// DCSA Standard Event Milestone Codes
const (
	DCSAMilestoneGateIn    = "GTIN" // Gate in
	DCSAMilestoneGateOut   = "GTOT" // Gate out
	DCSAMilestoneLoaded    = "LOAD" // Loaded on vessel / rail
	DCSAMilestoneDischarge = "DISC" // Discharged from vessel
	DCSAMilestoneDeparture = "DEPA" // Vessel departed
	DCSAMilestoneArrival   = "ARRI" // Vessel arrived
	DCSAMilestoneStowage   = "STOW" // Stowed in container yard
	DCSAMilestonePickUp    = "PICK" // Picked up by consignee/trucker
	DCSAMilestoneDelivery  = "DROP" // Delivered / dropped off
	DCSAMilestoneCustoms   = "CUST" // Customs released
	DCSAMilestoneInspect   = "INSP" // Inspection hold
)

// DCSATransportCall represents a port call or facility stop in DCSA Track & Trace.
type DCSATransportCall struct {
	TransportCallID       string         `json:"transportCallID,omitempty"`
	CarrierServiceCode    string         `json:"carrierServiceCode,omitempty"`
	VesselPartnerRef      string         `json:"vesselPartnerCarrierServiceRef,omitempty"`
	Vessel                *DCSAVessel    `json:"vessel,omitempty"`
	Location              *DCSALocation  `json:"location,omitempty"`
	FacilityCode          string         `json:"facilityCode,omitempty"`
	FacilityCodeListProv  string         `json:"facilityCodeListProvider,omitempty"`
	UNLocationCode        string         `json:"UNLocationCode,omitempty"`
	CarrierVoyageNumber   string         `json:"carrierVoyageNumber,omitempty"`
}

// DCSAVessel represents a maritime vessel in DCSA specifications.
type DCSAVessel struct {
	VesselIMONumber string `json:"vesselIMONumber,omitempty"`
	VesselName      string `json:"vesselName,omitempty"`
	VesselFlag      string `json:"vesselFlag,omitempty"`
	CallSign        string `json:"vesselCallSignNumber,omitempty"`
}

// DCSALocation represents a geographic or facility location.
type DCSALocation struct {
	LocationName   string  `json:"locationName,omitempty"`
	UNLocationCode string  `json:"UNLocationCode,omitempty"`
	Latitude       float64 `json:"latitude,omitempty"`
	Longitude      float64 `json:"longitude,omitempty"`
	Address        string  `json:"address,omitempty"`
}

// DCSAEvent represents a raw Track & Trace event conforming to DCSA OpenAPI spec.
type DCSAEvent struct {
	EventID                string             `json:"eventID"`
	EventType              string             `json:"eventType"` // "SHIPMENT", "TRANSPORT", "EQUIPMENT"
	EventDateTime          string             `json:"eventDateTime"`
	EventClassifierCode    string             `json:"eventClassifierCode"` // "ACT", "EST", "PLN"
	EventTypeCode          string             `json:"eventTypeCode,omitempty"`
	EquipmentEventTypeCode string             `json:"equipmentEventTypeCode,omitempty"` // "LOAD", "DISC", "GTIN", "GTOT"
	TransportEventTypeCode string             `json:"transportEventTypeCode,omitempty"` // "DEPA", "ARRI"
	ShipmentEventTypeCode  string             `json:"shipmentEventTypeCode,omitempty"`
	EquipmentReference     string             `json:"equipmentReference,omitempty"` // Container number
	ISOEquipmentCode       string             `json:"ISOEquipmentCode,omitempty"`   // e.g. "45G1", "22G1"
	EmptyIndicatorCode     string             `json:"emptyIndicatorCode,omitempty"` // "EMPTY", "LADEN"
	EventLocation          *DCSALocation      `json:"eventLocation,omitempty"`
	TransportCall          *DCSATransportCall `json:"transportCall,omitempty"`
	CarrierBookingRef      string             `json:"carrierBookingReference,omitempty"`
	DocumentReferences     []DCSADocRef       `json:"documentReferences,omitempty"`
}

// DCSADocRef represents document links (B/L, Booking, Shipping Instruction).
type DCSADocRef struct {
	DocumentType      string `json:"documentReferenceType"`
	DocumentReference string `json:"documentReferenceValue"`
}

// NormalizeDCSAMilestone translates DCSA event codes into LogisticsHQ canonical milestone codes.
func NormalizeDCSAMilestone(eventType, milestoneCode string) string {
	switch milestoneCode {
	case DCSAMilestoneGateIn:
		return "GATE_IN"
	case DCSAMilestoneGateOut:
		return "GATE_OUT"
	case DCSAMilestoneLoaded:
		return "LOADED"
	case DCSAMilestoneDischarge:
		return "DISCHARGED"
	case DCSAMilestoneDeparture:
		return "DEPARTED"
	case DCSAMilestoneArrival:
		return "ARRIVED"
	case DCSAMilestoneStowage:
		return "STOWED"
	case DCSAMilestonePickUp:
		return "PICKED_UP"
	case DCSAMilestoneDelivery:
		return "DELIVERED"
	case DCSAMilestoneCustoms:
		return "CUSTOMS_CLEARED"
	case DCSAMilestoneInspect:
		return "CUSTOMS_HOLD"
	default:
		if milestoneCode != "" {
			return milestoneCode
		}
		if eventType != "" {
			return eventType
		}
		return "IN_TRANSIT"
	}
}

// ParseDCSATime parses ISO8601 timestamps standard in carrier APIs.
func ParseDCSATime(rawTime string) time.Time {
	if rawTime == "" {
		return time.Time{}
	}
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, rawTime); err == nil {
			return t
		}
	}
	return time.Time{}
}
