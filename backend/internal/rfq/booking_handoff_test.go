package rfq_test

import (
	"context"
	"testing"
	"time"

	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/rfq"
	"github.com/freel/backend/internal/rfq/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 1. Pure Deterministic Eligibility & State Machine Tests
func TestBooking_Eligibility_Evaluation(t *testing.T) {
	origin := "INNSA"
	dest := "DEHAM"

	rfqObj := &spec.RFQ{
		ID:          100,
		RFQNumber:   "RFQ-TEST-100",
		Origin:      &origin,
		Destination: &dest,
	}

	// Case A: No quotes -> NOT eligible
	quotes := []spec.RFQQuote{}
	reqs := &spec.OperationalReadiness{BlockingCount: 0}
	docsSummary := &spec.DocumentSummary{MissingDocuments: 0}

	elig := rfq.EvaluateBookingEligibility(rfqObj, quotes, reqs, docsSummary)
	assert.False(t, elig.IsEligible)
	assert.Contains(t, elig.MissingPrerequisites, "Commercial Quote Approval Required")

	// Case B: Has Draft quote -> NOT eligible
	quotes = []spec.RFQQuote{
		{ID: 1, CarrierName: "Maersk", Status: spec.QuoteStatusReceived},
	}
	elig = rfq.EvaluateBookingEligibility(rfqObj, quotes, reqs, docsSummary)
	assert.False(t, elig.IsEligible)

	// Case C: Quote is APPROVED -> ELIGIBLE
	quotes = []spec.RFQQuote{
		{ID: 1, CarrierName: "Maersk", Status: spec.QuoteStatusApproved},
	}
	elig = rfq.EvaluateBookingEligibility(rfqObj, quotes, reqs, docsSummary)
	assert.True(t, elig.IsEligible)
	require.NotNil(t, elig.ApprovedQuoteID)
	assert.Equal(t, int64(1), *elig.ApprovedQuoteID)
	require.NotNil(t, elig.ApprovedCarrier)
	assert.Equal(t, "Maersk", *elig.ApprovedCarrier)

	// Case D: Quote is APPROVED but Trade Requirements have blockers -> NOT eligible
	reqsWithBlocker := &spec.OperationalReadiness{BlockingCount: 2}
	elig = rfq.EvaluateBookingEligibility(rfqObj, quotes, reqsWithBlocker, docsSummary)
	assert.False(t, elig.IsEligible)
	assert.Contains(t, elig.MissingPrerequisites[0], "Blockers")
}

func TestBooking_Lifecycle_Transitions(t *testing.T) {
	// Valid transitions
	assert.NoError(t, rfq.ValidateBookingTransition(spec.BookingStatusDraft, spec.BookingStatusRequested))
	assert.NoError(t, rfq.ValidateBookingTransition(spec.BookingStatusDraft, spec.BookingStatusCancelled))
	assert.NoError(t, rfq.ValidateBookingTransition(spec.BookingStatusRequested, spec.BookingStatusPendingConfirmation))
	assert.NoError(t, rfq.ValidateBookingTransition(spec.BookingStatusRequested, spec.BookingStatusConfirmed))
	assert.NoError(t, rfq.ValidateBookingTransition(spec.BookingStatusPendingConfirmation, spec.BookingStatusConfirmed))
	assert.NoError(t, rfq.ValidateBookingTransition(spec.BookingStatusConfirmed, spec.BookingStatusCompleted))
	assert.NoError(t, rfq.ValidateBookingTransition(spec.BookingStatusConfirmed, spec.BookingStatusCancelled))

	// Invalid transitions
	assert.Error(t, rfq.ValidateBookingTransition(spec.BookingStatusCancelled, spec.BookingStatusConfirmed))
	assert.Error(t, rfq.ValidateBookingTransition(spec.BookingStatusCompleted, spec.BookingStatusDraft))
	assert.Error(t, rfq.ValidateBookingTransition("INVALID", spec.BookingStatusConfirmed))
}

// 2. Full MySQL Workflow & Multi-Tenant Isolation Test
func TestBooking_FullWorkflowAndOrgIsolation(t *testing.T) {
	db, err := database.Connect("root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC&multiStatements=true")
	if err != nil {
		t.Skip("Skipping DB test: MySQL unavailable")
	}

	// Seed test organizations, customers, and carriers
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (8801, 'Booking Org 8801', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)")
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (8802, 'Booking Org 8802', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)")
	_, _ = db.Exec("INSERT INTO customers (id, org_id, name, contact_email, created_at, updated_at) VALUES (8801, 8801, 'Apex Imports', 'apex@example.com', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)")
	_, _ = db.Exec("INSERT INTO carriers (scac, name, is_active) VALUES ('HLCU', 'Hapag-Lloyd', 1) ON DUPLICATE KEY UPDATE name=VALUES(name)")

	dl := rfq.NewDataLayer(db)
	bus := events.NewInProcessBus()
	bl := rfq.NewBusinessLogic(dl, bus, nil, nil, nil, nil)
	ctx := context.Background()

	orgID := int32(8801)
	otherOrgID := int32(8802)

	// 1. Create RFQ with complete operational trade parameters
	origin := "INNSA"
	dest := "DEHAM"
	incoterms := "FOB"
	targetDate := time.Now().Add(14 * 24 * time.Hour)
	cargoWeight := 18500.0
	cargoVol := 45.0
	rfqCreated, err := bl.CreateRFQ(ctx, spec.CreateRFQRequest{
		OrgID:       orgID,
		CustomerID:  8801,
		Origin:      &origin,
		Destination: &dest,
		Incoterms:   &incoterms,
		TargetDate:  &targetDate,
		Items: []spec.RFQItem{
			{
				Description: "2x 40HC Industrial Machinery",
				Quantity:    2,
				WeightKG:    &cargoWeight,
				VolumeCBM:   &cargoVol,
			},
		},
	})
	require.NoError(t, err)

	// 2. Check initial booking handoff (not eligible yet)
	initialHandoff, err := bl.GetBookingHandoff(ctx, orgID, rfqCreated.ID)
	require.NoError(t, err)
	assert.False(t, initialHandoff.Eligibility.IsEligible)
	assert.Equal(t, 0, initialHandoff.Summary.TotalBookings)

	// 3. Add and approve a quote
	quote, err := bl.CreateRFQQuote(ctx, orgID, rfqCreated.ID, spec.CreateQuoteRequest{
		CarrierName: "Hapag-Lloyd",
		BuyPrice:    4500.0,
		SellPrice:   5500.0,
	}, "Pricing Agent")
	require.NoError(t, err)

	_, err = bl.ApproveRFQQuote(ctx, orgID, rfqCreated.ID, quote.ID, spec.ApproveRFQQuoteRequest{}, "Commercial Director")
	require.NoError(t, err)

	// 4. Verify eligibility now true
	eligibleHandoff, err := bl.GetBookingHandoff(ctx, orgID, rfqCreated.ID)
	require.NoError(t, err)
	assert.True(t, eligibleHandoff.Eligibility.IsEligible)
	require.NotNil(t, eligibleHandoff.Eligibility.ApprovedCarrier)
	assert.Equal(t, "Hapag-Lloyd", *eligibleHandoff.Eligibility.ApprovedCarrier)

	// 5. Create Booking from RFQ
	bookingNum := "BK-2026-TEST8801"
	vessel := "Hamburg Express"
	voyage := "HE-042W"
	booking, err := bl.CreateBookingFromRFQ(ctx, orgID, rfqCreated.ID, spec.CreateBookingRequest{
		BookingNumber: &bookingNum,
		CarrierName:   "Hapag-Lloyd",
		VesselName:    &vessel,
		VoyageNumber:  &voyage,
		OriginPort:    "INNSA",
		DestinationPort: "DEHAM",
	}, "Operations Specialist")
	require.NoError(t, err)
	assert.Equal(t, "BK-2026-TEST8801", booking.BookingNumber)
	assert.Equal(t, spec.BookingStatusDraft, booking.Status)

	// 6. Update Booking Status: DRAFT -> REQUESTED -> CONFIRMED
	bookingReq, err := bl.UpdateBookingStatus(ctx, orgID, rfqCreated.ID, booking.ID, spec.UpdateBookingStatusRequest{
		Status: spec.BookingStatusRequested,
	}, "Ops Lead")
	require.NoError(t, err)
	assert.Equal(t, spec.BookingStatusRequested, bookingReq.Status)

	bookingConf, err := bl.UpdateBookingStatus(ctx, orgID, rfqCreated.ID, booking.ID, spec.UpdateBookingStatusRequest{
		Status: spec.BookingStatusConfirmed,
	}, "Carrier Agent")
	require.NoError(t, err)
	assert.Equal(t, spec.BookingStatusConfirmed, bookingConf.Status)

	// 7. Verify RFQ Activity Timeline contains real BOOKING_CREATED and BOOKING_CONFIRMED events
	actResp, err := bl.GetActivity(ctx, orgID, rfqCreated.ID)
	require.NoError(t, err)
	hasBookingCreated := false
	hasBookingConfirmed := false
	for _, ev := range actResp.Events {
		if ev.Type == spec.ActivityBookingCreated {
			hasBookingCreated = true
		}
		if ev.Type == spec.ActivityBookingConfirmed {
			hasBookingConfirmed = true
		}
	}
	assert.True(t, hasBookingCreated, "Timeline must contain real BOOKING_CREATED event")
	assert.True(t, hasBookingConfirmed, "Timeline must contain real BOOKING_CONFIRMED event")

	// 8. Create a shipment linked to this booking
	_, err = db.Exec(`
		INSERT INTO shipments (org_id, rfq_id, quote_id, booking_id, booking_number, carrier_scac, origin_port, destination_port, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'BOOKED', NOW(), NOW())
	`, orgID, rfqCreated.ID, quote.ID, booking.ID, booking.BookingNumber, "HLCU", "INNSA", "DEHAM")
	require.NoError(t, err)

	// 9. Verify Shipment Handoff returns the linked shipment
	shipHandoff, err := bl.GetShipmentHandoff(ctx, orgID, rfqCreated.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, shipHandoff.Summary.TotalShipments)
	require.NotNil(t, shipHandoff.Summary.ActiveShipment)
	assert.Equal(t, "BOOKED", shipHandoff.Summary.ActiveShipment.Status)
	assert.Equal(t, "Hapag-Lloyd", shipHandoff.Summary.ActiveShipment.CarrierName)

	// 10. Multi-Tenant Organization Isolation Verification
	// Other org cannot access Org 8801's bookings or shipments
	_, err = bl.GetBookingHandoff(ctx, otherOrgID, rfqCreated.ID)
	assert.Error(t, err, "Org 8802 must not access Org 8801 booking handoff")

	_, err = bl.GetShipmentHandoff(ctx, otherOrgID, rfqCreated.ID)
	assert.Error(t, err, "Org 8802 must not access Org 8801 shipment handoff")
}
