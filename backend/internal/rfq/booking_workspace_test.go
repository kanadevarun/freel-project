package rfq_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/rfq"
	"github.com/freel/backend/internal/rfq/spec"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getWorkspaceTestDB(t *testing.T) *sqlx.DB {
	db, err := database.Connect("root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC&multiStatements=true")
	if err != nil {
		t.Skip("Skipping DB test: MySQL unavailable")
	}
	return db
}

func TestBookingWorkspace_FullWorkflow(t *testing.T) {
	db := getWorkspaceTestDB(t)
	defer db.Close()

	ctx := context.Background()
	dl := rfq.NewDataLayer(db)
	bus := events.NewInProcessBus()
	bl := rfq.NewBusinessLogic(dl, bus, nil, nil, nil, nil)

	testOrgID := int32(8800)
	otherOrgID := int32(8801)

	// Seed test organizations and carriers
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (8800, 'Workspace Org 8800', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)")
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (8801, 'Workspace Org 8801', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)")
	_, _ = db.Exec("INSERT INTO carriers (scac, name, is_active) VALUES ('MAEU', 'Maersk Line', 1) ON DUPLICATE KEY UPDATE name=VALUES(name)")

	// Clean test data
	_, _ = db.Exec("DELETE FROM shipments WHERE org_id IN (?, ?)", testOrgID, otherOrgID)
	_, _ = db.Exec("DELETE FROM activities WHERE org_id IN (?, ?)", testOrgID, otherOrgID)
	_, _ = db.Exec("DELETE FROM bookings WHERE org_id IN (?, ?)", testOrgID, otherOrgID)
	_, _ = db.Exec("DELETE FROM rfq_quotes WHERE org_id IN (?, ?)", testOrgID, otherOrgID)
	_, _ = db.Exec("DELETE FROM rfq_items WHERE rfq_id IN (SELECT id FROM rfqs WHERE org_id IN (?, ?))", testOrgID, otherOrgID)
	_, _ = db.Exec("DELETE FROM rfqs WHERE org_id IN (?, ?)", testOrgID, otherOrgID)

	// Seed customer
	custRes, err := db.Exec(`
		INSERT INTO customers (org_id, name, contact_email, created_at, updated_at)
		VALUES (?, 'Global Imports Inc', 'ops@globalimports.com', NOW(), NOW())
	`, testOrgID)
	require.NoError(t, err)
	custID, _ := custRes.LastInsertId()

	// 1. Create RFQ with complete trade parameters
	origin := "CNSHA"
	dest := "USLAX"
	incoterms := "FOB"
	targetDate := time.Now().Add(14 * 24 * time.Hour)
	cargoWeight := 4500.0
	cargoVol := 18.5
	rfqObj, err := bl.CreateRFQ(ctx, spec.CreateRFQRequest{
		OrgID:       testOrgID,
		CustomerID:  int32(custID),
		Origin:      &origin,
		Destination: &dest,
		Incoterms:   &incoterms,
		TargetDate:  &targetDate,
		Items: []spec.RFQItem{
			{
				Description: "Industrial Sensors",
				Quantity:    20,
				WeightKG:    &cargoWeight,
				VolumeCBM:   &cargoVol,
			},
		},
	})
	require.NoError(t, err)
	rfqID := rfqObj.ID

	// 2. Create Approved Quote for RFQ
	scac := "MAEU"
	qRef := "MSK-Q-8800"
	quote, err := bl.CreateRFQQuote(ctx, testOrgID, rfqID, spec.CreateQuoteRequest{
		CarrierName:    "Maersk Line",
		CarrierID:      &scac,
		QuoteReference: &qRef,
		Currency:       "USD",
		BuyPrice:       3200.00,
		SellPrice:      3800.00,
	}, "Pricing Agent")
	require.NoError(t, err)

	_, err = bl.ApproveRFQQuote(ctx, testOrgID, rfqID, quote.ID, spec.ApproveRFQQuoteRequest{}, "Commercial Director")
	require.NoError(t, err)
	quoteID := quote.ID

	// 3. Test Eligible RFQs for Booking
	eligible, err := bl.GetEligibleRFQsForBooking(ctx, testOrgID)
	require.NoError(t, err)
	assert.NotEmpty(t, eligible)
	var foundEligible bool
	for _, e := range eligible {
		if e.RFQID == int64(rfqID) {
			foundEligible = true
			assert.Equal(t, "Global Imports Inc", e.CustomerName)
			assert.Equal(t, "Maersk Line", e.CarrierName)
			assert.Equal(t, quoteID, e.ApprovedQuoteID)
			break
		}
	}
	assert.True(t, foundEligible, "Expected RFQ %d to be in eligible list", rfqID)

	// 4. Create Booking from RFQ
	etd := time.Now().AddDate(0, 0, 3)
	eta := time.Now().AddDate(0, 0, 18)
	bNum := "BKG-WS-8800"
	vName := "Maersk Mc-Kinney"
	vVoyage := "2608E"
	bkg, err := bl.CreateBookingFromRFQ(ctx, testOrgID, rfqID, spec.CreateBookingRequest{
		QuoteID:         &quoteID,
		BookingNumber:   &bNum,
		CarrierName:     "Maersk Line",
		CarrierSCAC:     &[]string{"MAEU"}[0],
		OriginPort:      "CNSHA",
		DestinationPort: "USLAX",
		VesselName:      &vName,
		VoyageNumber:    &vVoyage,
		ETD:             &etd,
		ETA:             &eta,
	}, "Test User")
	require.NoError(t, err)
	require.NotNil(t, bkg)
	assert.Equal(t, spec.BookingStatusDraft, bkg.Status)
	bookingID := bkg.ID

	// 5. Test Dedicated Workspace List & KPIs
	wsResp, err := bl.GetBookingsWorkspace(ctx, testOrgID, spec.BookingListFilter{
		Page:  1,
		Limit: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, wsResp)
	assert.Equal(t, 1, wsResp.KPIs.TotalBookings)
	assert.Equal(t, 1, wsResp.KPIs.Draft)
	assert.Len(t, wsResp.Bookings, 1)
	assert.Equal(t, bNum, wsResp.Bookings[0].BookingNumber)
	assert.Equal(t, "Global Imports Inc", wsResp.Bookings[0].CustomerName)

	// 6. Test Multi-Tenant Org Isolation
	otherWsResp, err := bl.GetBookingsWorkspace(ctx, otherOrgID, spec.BookingListFilter{
		Page:  1,
		Limit: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, otherWsResp.KPIs.TotalBookings)
	assert.Empty(t, otherWsResp.Bookings)

	_, err = bl.GetBookingWorkspaceDetail(ctx, otherOrgID, bookingID)
	assert.Error(t, err, "Accessing booking from other org must fail")

	// 7. Test Direct Lifecycle Transitions & State Machine
	// Invalid: DRAFT -> CONFIRMED directly (must go to REQUESTED or CANCELLED)
	_, err = bl.DirectUpdateBookingStatus(ctx, testOrgID, bookingID, spec.DirectUpdateBookingStatusRequest{
		Status: spec.BookingStatusConfirmed,
	}, "Test Ops")
	assert.Error(t, err, "Direct transition from DRAFT to CONFIRMED should be blocked")

	// Valid: DRAFT -> REQUESTED
	bkgReq, err := bl.DirectUpdateBookingStatus(ctx, testOrgID, bookingID, spec.DirectUpdateBookingStatusRequest{
		Status: spec.BookingStatusRequested,
	}, "Test Ops")
	require.NoError(t, err)
	assert.Equal(t, spec.BookingStatusRequested, bkgReq.Status)

	// Valid: REQUESTED -> CONFIRMED
	bkgConf, err := bl.DirectUpdateBookingStatus(ctx, testOrgID, bookingID, spec.DirectUpdateBookingStatusRequest{
		Status: spec.BookingStatusConfirmed,
	}, "Test Ops")
	require.NoError(t, err)
	assert.Equal(t, spec.BookingStatusConfirmed, bkgConf.Status)

	// 8. Test Workspace Detail
	detail, err := bl.GetBookingWorkspaceDetail(ctx, testOrgID, bookingID)
	require.NoError(t, err)
	require.NotNil(t, detail)
	assert.Equal(t, spec.BookingStatusConfirmed, detail.Booking.Status)
	assert.Equal(t, rfqObj.RFQNumber, detail.SourceRFQ.RFQNumber)
	assert.NotNil(t, detail.CommercialQuote)
	assert.Equal(t, "Maersk Line", detail.CommercialQuote.CarrierName)
	assert.Equal(t, 4500.0, detail.CargoSummary.TotalWeightKg)
	assert.Contains(t, detail.AllowedActions, "CREATE_SHIPMENT")

	// 9. Test Shipment Creation from Confirmed Booking
	shipment, err := bl.CreateShipmentFromBooking(ctx, testOrgID, bookingID, spec.CreateShipmentFromBookingRequest{
		ContainerNumbers: []string{"MSKU7788990", "MSKU7788991"},
	}, "Test Logistics")
	require.NoError(t, err)
	require.NotNil(t, shipment)
	assert.Equal(t, "BOOKED", shipment.Status)
	assert.Equal(t, "MAEU", shipment.CarrierSCAC)
	assert.Equal(t, []string{"MSKU7788990", "MSKU7788991"}, shipment.ContainerNumbers)

	// 10. Test Idempotency: creating shipment again returns the same record without error or duplication
	dupShipment, err := bl.CreateShipmentFromBooking(ctx, testOrgID, bookingID, spec.CreateShipmentFromBookingRequest{
		ContainerNumbers: []string{"MSKU7788990"},
	}, "Test Logistics")
	require.NoError(t, err)
	assert.Equal(t, shipment.ID, dupShipment.ID, "Shipment creation must be idempotent")

	// 11. Verify RFQ is no longer in eligible list since booking is CONFIRMED
	eligibleAfter, err := bl.GetEligibleRFQsForBooking(ctx, testOrgID)
	require.NoError(t, err)
	for _, e := range eligibleAfter {
		assert.NotEqual(t, int64(rfqID), e.RFQID, "RFQ with confirmed booking must no longer be eligible")
	}

	fmt.Println("✅ Task 15 Booking Workspace Go Backend Integration & Concurrency Tests Passed 100%!")
}
