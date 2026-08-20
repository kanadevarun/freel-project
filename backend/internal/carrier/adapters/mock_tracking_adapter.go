package adapters

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/freel/backend/internal/carrier"
)

// MockTrackingAdapter implements TrackingProvider, BookingProvider, WebhookProvider
type MockTrackingAdapter struct {
	SCAC   string
	Config *carrier.IntegrationConfig
}

func NewMockTrackingAdapter(scac string, cfg *carrier.IntegrationConfig) *MockTrackingAdapter {
	return &MockTrackingAdapter{
		SCAC:   scac,
		Config: cfg,
	}
}

func (a *MockTrackingAdapter) checkConfig(capability string) error {
	if a.Config == nil {
		return fmt.Errorf("carrier not configured for this organization")
	}
	if a.Config.APIKey == "" || a.Config.APIKey == "invalid" {
		return fmt.Errorf("API authentication failure: invalid credentials")
	}
	if !a.Config.Capabilities[capability] {
		return fmt.Errorf("unsupported capability: %s", capability)
	}
	return nil
}

func (a *MockTrackingAdapter) GetTrackingEvents(ctx context.Context, req carrier.TrackingRequest) ([]carrier.TrackingEvent, error) {
	if err := a.checkConfig("TRACKING"); err != nil {
		return nil, err
	}

	// Simulated error cases for negative testing
	if req.BookingNumber == "timeout" {
		return nil, fmt.Errorf("API timeout")
	}
	if req.BookingNumber == "auth-failure" {
		return nil, fmt.Errorf("API authentication failure: invalid credentials")
	}
	if req.BookingNumber == "malformed-payload" {
		return nil, fmt.Errorf("malformed carrier response")
	}

	now := time.Now()
	eventID := fmt.Sprintf("MOCK-EV-%s-%d", req.BookingNumber, now.Unix())
	
	var rawPayload []byte
	var milestone string
	var location string
	var desc string

	scacUpper := strings.ToUpper(a.SCAC)
	if scacUpper == "MAEU" || scacUpper == "MSK" {
		milestone = "DEPARTED"
		location = "Nhava Sheva"
		desc = "Vessel departed origin port (Nhava Sheva)."
		rawPayload, _ = json.Marshal(map[string]string{
			"container": "MSKU1234567",
			"status":    "VESSEL DEPARTED",
			"eventDate": now.Format(time.RFC3339),
			"location":  location,
		})
	} else if scacUpper == "MSC" {
		milestone = "IN_TRANSIT"
		location = "INNSA"
		desc = "Vessel departed transshipment port (SGSIN)."
		rawPayload, _ = json.Marshal(map[string]string{
			"equipmentNo": "MSKU1234567",
			"milestone":   "DEP",
			"timestamp":   now.Format(time.RFC3339),
			"port":        location,
		})
	} else {
		milestone = "IN_TRANSIT"
		location = "SGSIN"
		desc = "Vessel departed transshipment port (SGSIN)."
		rawPayload, _ = json.Marshal(map[string]string{
			"booking": req.BookingNumber,
			"status":  "in_transit",
			"carrier": a.SCAC,
		})
	}

	return []carrier.TrackingEvent{
		{
			EventID:       eventID,
			MilestoneCode: milestone,
			EventTime:     now,
			Location:      location,
			VesselName:    "MOCK VESSEL",
			VoyageNumber:  "2601E",
			Description:   desc,
			RawPayload:    rawPayload,
		},
	}, nil
}

func (a *MockTrackingAdapter) GetBooking(ctx context.Context, bookingNumber string) (*carrier.BookingStatus, error) {
	if err := a.checkConfig("BOOKING"); err != nil {
		return nil, err
	}

	etd := time.Now().AddDate(0, 0, 5)
	eta := etd.AddDate(0, 0, 14)
	return &carrier.BookingStatus{
		BookingNumber: bookingNumber,
		Status:        "CONFIRMED",
		VesselName:    "MOCK VESSEL",
		VoyageNumber:  "2601E",
		ETD:           &etd,
		ETA:           &eta,
	}, nil
}

func (a *MockTrackingAdapter) VerifyWebhookSignature(payload []byte, headers map[string]string) error {
	if err := a.checkConfig("WEBHOOK"); err != nil {
		return err
	}

	sig := headers["X-Mock-Signature"]
	expected := fmt.Sprintf("%x", sha256.Sum256(payload))
	if sig != "" && sig != expected {
		return fmt.Errorf("invalid mock webhook signature")
	}
	return nil
}

func (a *MockTrackingAdapter) ParseWebhookPayload(payload []byte) (*carrier.TrackingEvent, error) {
	// Try parsing Maersk format
	var maerskBody struct {
		Container string `json:"container"`
		Status    string `json:"status"`
		EventDate string `json:"event_date"`
		Location  string `json:"location"`
	}
	if err := json.Unmarshal(payload, &maerskBody); err == nil && maerskBody.Container != "" && maerskBody.Status != "" {
		t, parseErr := time.Parse(time.RFC3339, maerskBody.EventDate)
		if parseErr != nil {
			t = time.Now()
		}
		milestone := "IN_TRANSIT"
		if strings.Contains(strings.ToUpper(maerskBody.Status), "DEPARTED") {
			milestone = "DEPARTED"
		}
		return &carrier.TrackingEvent{
			EventID:       fmt.Sprintf("MAEU-WK-%d", time.Now().UnixNano()),
			MilestoneCode: milestone,
			EventTime:     t,
			Location:      maerskBody.Location,
			Description:   fmt.Sprintf("Maersk Event: %s at %s", maerskBody.Status, maerskBody.Location),
			RawPayload:    payload,
		}, nil
	}

	// Try parsing MSC format
	var mscBody struct {
		EquipmentNo string `json:"equipmentNo"`
		Milestone   string `json:"milestone"`
		Timestamp   string `json:"timestamp"`
		Port        string `json:"port"`
	}
	if err := json.Unmarshal(payload, &mscBody); err == nil && mscBody.EquipmentNo != "" && mscBody.Milestone != "" {
		t, parseErr := time.Parse(time.RFC3339, mscBody.Timestamp)
		if parseErr != nil {
			t = time.Now()
		}
		milestone := "IN_TRANSIT"
		if mscBody.Milestone == "DEP" {
			milestone = "DEPARTED"
		}
		return &carrier.TrackingEvent{
			EventID:       fmt.Sprintf("MSC-WK-%d", time.Now().UnixNano()),
			MilestoneCode: milestone,
			EventTime:     t,
			Location:      mscBody.Port,
			Description:   fmt.Sprintf("MSC Milestone: %s at %s", mscBody.Milestone, mscBody.Port),
			RawPayload:    payload,
		}, nil
	}

	// Fallback to original generic payload schema for backward compatibility
	var body struct {
		EventID     string `json:"event_id"`
		BookingNum  string `json:"booking_number"`
		Milestone   string `json:"milestone"`
		Description string `json:"description"`
		Location    string `json:"location"`
	}
	if err := json.Unmarshal(payload, &body); err == nil && body.EventID != "" {
		return &carrier.TrackingEvent{
			EventID:       body.EventID,
			MilestoneCode: body.Milestone,
			EventTime:     time.Now(),
			Location:      body.Location,
			Description:   body.Description,
			RawPayload:    payload,
		}, nil
	}

	return nil, fmt.Errorf("malformed carrier response / normalization failure")
}
