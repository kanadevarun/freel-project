package carrier

import (
	"context"
	"time"
)

// CarrierRate represents a single pricing option returned by a carrier API.
type CarrierRate struct {
	CarrierName           string
	BuyPrice              float64
	TransitDays           int
	ReliabilityScore      int
	HistoricalSuccessRate float64
	// FreeDays is the number of days a customer can store the container at the destination port
	// without incurring demurrage/detention penalties. E.g., 7, 14 days.
	FreeDays              int
	// Shipping line operational details from the rate sheet
	VesselName            string  // E.g., "CMA CGM CASSIOPEIA"
	ServiceCode           string  // E.g., "AS1" (service code/route tag)
	ViaPort               string  // Optional transshipment port (e.g., "SINGAPORE, SG")
	CO2Emissions          float64 // CO2e tonnes per TEU (e.g., 3.67)
	NauticalMiles         int     // Total nautical miles (e.g., 4168)
	// Price itemisation breakdown
	OceanFreight          float64 // Base ocean rate (e.g., 2800)
	OriginCharges         float64 // Export handling fees (e.g., terminal handling THC)
	DestinationCharges    float64 // Import handling and local terminal fees at dest
}

// RateProvider is the rate extraction interface.
type RateProvider interface {
	GetRates(ctx context.Context, origin, destination string, incoterms string, grossWeight float64, volumeCBM float64, commodity string) ([]CarrierRate, error)
}

// CarrierProvider is kept as alias to RateProvider for backward compatibility.
type CarrierProvider = RateProvider

// TrackingRequest contains identifiers used to query a carrier for operational tracking events.
type TrackingRequest struct {
	BookingNumber   string
	ContainerNumber string
	MBLNumber       string
	CarrierSCAC     string
}

// TrackingEvent is a raw tracking update returned by a carrier adapter.
type TrackingEvent struct {
	EventID         string
	MilestoneCode   string
	EventTime       time.Time
	Location        string
	VesselName      string
	VoyageNumber    string
	Description     string
	RawPayload      []byte
}

// TrackingProvider is implemented by adapters that support active tracking event fetches.
type TrackingProvider interface {
	GetTrackingEvents(ctx context.Context, req TrackingRequest) ([]TrackingEvent, error)
}

// BookingStatus holds carrier booking confirmation data.
type BookingStatus struct {
	BookingNumber string
	Status        string
	VesselName    string
	VoyageNumber  string
	ETD           *time.Time
	ETA           *time.Time
}

// BookingProvider is implemented by adapters that support querying/confirming carrier bookings.
type BookingProvider interface {
	GetBooking(ctx context.Context, bookingNumber string) (*BookingStatus, error)
}

// WebhookProvider is implemented by adapters that handle inbound push notifications (webhooks).
type WebhookProvider interface {
	VerifyWebhookSignature(payload []byte, headers map[string]string) error
	ParseWebhookPayload(payload []byte) (*TrackingEvent, error)
}

