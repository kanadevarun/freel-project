package quotations

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestBuildQuotationOperationalHandover(t *testing.T) {
	now := time.Now()
	bID := int64(101)
	sID := int64(201)
	cName := "Maersk Line"
	vName := "Maersk Mc-Kinney"
	voyage := "2601E"

	q := &Quotation{
		ID:                 1,
		OrgID:              10,
		QuotationNumber:    "QT-2026-0001",
		CustomerName:       "Acme Corp",
		Status:             QuotationStatusAccepted,
		Origin:             "Nhava Sheva",
		OriginCode:         "INNSA",
		Destination:        "Rotterdam",
		DestinationCode:    "NLRTM",
		Currency:           "USD",
		TotalAmount:        2930.00,
		ConvertedBookingID: &bID,
		ConvertedShipmentID: &sID,
		ConversionStatus:   QuotationConversionStatusConverted,
		ConvertedAt:        &now,
		ConvertedBy:        "Operations User",
		CreatedAt:          now.Add(-24 * time.Hour),
		AcceptedAt:         &now,
	}

	booking := &RawOperationalBooking{
		ID:                       101,
		OrgID:                    10,
		BookingNumber:            "BK-2026-0001",
		Status:                   "CONFIRMED",
		CarrierName:              &cName,
		OriginPort:               "Nhava Sheva",
		DestinationPort:          "Rotterdam",
		VesselName:               &vName,
		VoyageNumber:             &voyage,
		CommercialHandoverStatus: CommercialHandoverBookingConfirmed,
		ConfirmedAt:              &now,
		CreatedAt:                now,
	}

	shipment := &RawOperationalShipment{
		ID:              201,
		OrgID:           10,
		BookingID:       101,
		BookingNumber:   "BK-2026-0001",
		CarrierSCAC:     "MAEU",
		OriginPort:      "Nhava Sheva",
		DestinationPort: "Rotterdam",
		VesselName:      &vName,
		VoyageNumber:    &voyage,
		Status:          "BOOKED",
		CreatedAt:       now,
	}

	handover := BuildQuotationOperationalHandover(q, booking, shipment)

	if handover == nil {
		t.Fatalf("expected non-nil handover")
	}
	if handover.QuotationNumber != "QT-2026-0001" {
		t.Errorf("expected QT-2026-0001, got %s", handover.QuotationNumber)
	}
	if handover.BookingNumber != "BK-2026-0001" {
		t.Errorf("expected BK-2026-0001, got %s", handover.BookingNumber)
	}
	if handover.HandoverStatus != CommercialHandoverBookingConfirmed {
		t.Errorf("expected BOOKING_CONFIRMED, got %s", handover.HandoverStatus)
	}
	if len(handover.LineageChain) != 5 {
		t.Errorf("expected 5 lineage steps, got %d", len(handover.LineageChain))
	}
	if handover.CommercialSnapshot.TotalAmount != 2930.00 {
		t.Errorf("expected 2930.00 commercial total snapshot, got %f", handover.CommercialSnapshot.TotalAmount)
	}
}

func TestDetectOperationalChanges(t *testing.T) {
	now := time.Now()
	v1 := "Vessel Alpha"
	v2 := "Vessel Beta"

	q := &Quotation{
		ID:              1,
		OrgID:           10,
		QuotationNumber: "QT-2026-0001",
		Origin:          "Nhava Sheva",
		OriginCode:      "INNSA",
		Destination:     "Rotterdam",
		DestinationCode: "NLRTM",
	}

	// Booking has different origin port and schedule
	booking := &RawOperationalBooking{
		ID:              101,
		OrgID:           10,
		BookingNumber:   "BK-2026-0001",
		OriginPort:      "Mundra", // Changed from Nhava Sheva
		DestinationPort: "Rotterdam",
		VesselName:      &v1,
		ETD:             &now,
		UpdatedAt:       now,
	}

	later := now.Add(48 * time.Hour)
	shipment := &RawOperationalShipment{
		ID:              201,
		OrgID:           10,
		BookingID:       101,
		VesselName:      &v2, // Changed to Vessel Beta
		ETD:             &later, // Delayed by 48h
		UpdatedAt:       later,
	}

	changes := DetectOperationalChanges(q, booking, shipment)

	if len(changes) == 0 {
		t.Fatalf("expected operational changes to be detected")
	}

	var foundOriginChange, foundVesselChange, foundETDChange bool
	for _, ch := range changes {
		if ch.Field == "Origin Port" && ch.BaselineValue == "Nhava Sheva" && ch.CurrentValue == "Mundra" {
			foundOriginChange = true
		}
		if ch.Field == "Vessel / Flight" && ch.BaselineValue == "Vessel Alpha" && ch.CurrentValue == "Vessel Beta" {
			foundVesselChange = true
		}
		if ch.Field == "Estimated Departure (ETD)" {
			foundETDChange = true
		}
		if ch.IsCommercial {
			t.Errorf("expected is_commercial to be false for operational changes")
		}
	}

	if !foundOriginChange {
		t.Errorf("expected Origin Port change to be detected")
	}
	if !foundVesselChange {
		t.Errorf("expected Vessel change to be detected")
	}
	if !foundETDChange {
		t.Errorf("expected ETD schedule change to be detected")
	}
}

func TestConfirmQuotationBookingHandover_Validation(t *testing.T) {
	bID := int64(501)
	booking := &RawOperationalBooking{
		ID:                       501,
		OrgID:                    10,
		BookingNumber:            "BK-501",
		Status:                   "DRAFT",
		CommercialHandoverStatus: CommercialHandoverPending,
	}

	qAccepted := &Quotation{
		ID:                 1,
		OrgID:              10,
		QuotationNumber:    "QT-2026-0001",
		Status:             QuotationStatusAccepted,
		ConvertedBookingID: &bID,
		ConversionStatus:   QuotationConversionStatusConverted,
	}

	mockRepo := &mockHandoverRepo{
		q:       qAccepted,
		booking: booking,
	}

	svc := &service{
		repo: mockRepo,
	}

	// 1. Successful confirmation
	res, err := svc.ConfirmQuotationBookingHandover(context.Background(), 10, 1, 1, &ConfirmQuotationHandoverRequest{
		ConfirmationNotes: "Handed over to ops successfully",
	})
	if err != nil {
		t.Fatalf("expected successful confirmation, got error: %v", err)
	}
	if res.HandoverStatus != CommercialHandoverBookingConfirmed {
		t.Errorf("expected status %s, got %s", CommercialHandoverBookingConfirmed, res.HandoverStatus)
	}

	// 2. Duplicate confirmation attempt must be blocked
	_, err = svc.ConfirmQuotationBookingHandover(context.Background(), 10, 1, 1, &ConfirmQuotationHandoverRequest{
		ConfirmationNotes: "Duplicate attempt",
	})
	if err == nil {
		t.Fatalf("expected error on duplicate handover confirmation, got nil")
	}
}

type mockHandoverRepo struct {
	Repository
	q       *Quotation
	booking *RawOperationalBooking
	hist    []*QuotationOperationalHandoverHistory
}

func (m *mockHandoverRepo) GetQuotationByID(ctx context.Context, orgID, quotationID int64) (*Quotation, error) {
	if m.q != nil && m.q.ID == quotationID {
		return m.q, nil
	}
	return nil, sql.ErrNoRows
}

func (m *mockHandoverRepo) GetOperationalBooking(ctx context.Context, orgID, bookingID int64) (*RawOperationalBooking, error) {
	if m.booking != nil && m.booking.ID == bookingID {
		return m.booking, nil
	}
	return nil, sql.ErrNoRows
}

func (m *mockHandoverRepo) GetOperationalBookingByQuotationID(ctx context.Context, orgID, quotationID int64) (*RawOperationalBooking, error) {
	if m.booking != nil {
		return m.booking, nil
	}
	return nil, sql.ErrNoRows
}

func (m *mockHandoverRepo) GetOperationalShipment(ctx context.Context, orgID, shipmentID int64) (*RawOperationalShipment, error) {
	return nil, sql.ErrNoRows
}

func (m *mockHandoverRepo) GetOperationalShipmentByBookingID(ctx context.Context, orgID, bookingID int64) (*RawOperationalShipment, error) {
	return nil, sql.ErrNoRows
}

func (m *mockHandoverRepo) UpdateBookingHandoverStatus(ctx context.Context, orgID, bookingID int64, status, confirmedBy string) error {
	if m.booking != nil {
		m.booking.CommercialHandoverStatus = status
		m.booking.ConfirmedBy = &confirmedBy
		now := time.Now()
		m.booking.ConfirmedAt = &now
	}
	return nil
}

func (m *mockHandoverRepo) CreateOperationalHandoverHistory(ctx context.Context, h *QuotationOperationalHandoverHistory) error {
	m.hist = append(m.hist, h)
	return nil
}

func (m *mockHandoverRepo) CreateActivity(ctx context.Context, act *QuotationActivity) error {
	return nil
}

