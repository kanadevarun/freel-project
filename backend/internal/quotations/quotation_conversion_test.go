package quotations

import (
	"context"
	"testing"
	"time"

	"github.com/freel/backend/internal/rates"
)

// Unit tests for pure eligibility engine
func TestCanConvertQuotationToBooking(t *testing.T) {
	// Scenario 1: Non-accepted statuses must be blocked
	blockedStatuses := []string{
		QuotationStatusDraft,
		QuotationStatusReadyForReview,
		QuotationStatusApproved,
		QuotationStatusSent,
		QuotationStatusViewed,
		QuotationStatusDeclined,
		QuotationStatusExpired,
		QuotationStatusCancelled,
	}

	for _, st := range blockedStatuses {
		q := &Quotation{
			ID:              101,
			OrgID:           1,
			QuotationNumber: "QT-2026-0001",
			Status:          st,
			CustomerName:    "Acme Corp",
			Origin:          "Shanghai, China",
			Destination:     "Los Angeles, USA",
			ServiceType:     "FCL",
			TransportMode:   "Ocean Freight",
			TotalAmount:     5000.00,
		}
		canConvert, reasons := CanConvertQuotationToBooking(q, nil)
		if canConvert {
			t.Errorf("Expected status %s to be blocked from conversion, but CanConvert returned true", st)
		}
		if len(reasons) == 0 {
			t.Errorf("Expected blocking reasons for status %s, got none", st)
		}
	}

	// Scenario 2: Accepted quote with complete data should be eligible
	acceptedQ := &Quotation{
		ID:              102,
		OrgID:           1,
		QuotationNumber: "QT-2026-0002",
		Status:          QuotationStatusAccepted,
		CustomerName:    "Global Trade Ltd",
		Origin:          "Rotterdam, Netherlands",
		Destination:     "New York, USA",
		ServiceType:     "FCL",
		TransportMode:   "Ocean Freight",
		TotalAmount:     7200.00,
	}
	canConvert, reasons := CanConvertQuotationToBooking(acceptedQ, nil)
	if !canConvert {
		t.Errorf("Expected accepted quote to be convertible, but got blocked: %v", reasons)
	}

	// Scenario 3: Missing route details must block
	missingRouteQ := &Quotation{
		ID:              103,
		OrgID:           1,
		QuotationNumber: "QT-2026-0003",
		Status:          QuotationStatusAccepted,
		CustomerName:    "Global Trade Ltd",
		Origin:          "",
		Destination:     "",
		TotalAmount:     7200.00,
	}
	canConvert, reasons = CanConvertQuotationToBooking(missingRouteQ, nil)
	if canConvert {
		t.Errorf("Expected quote with missing route to be blocked, but CanConvert returned true")
	}

	// Scenario 4: Already converted quote must block duplicate conversion
	bookingID := int64(9901)
	alreadyConvertedQ := &Quotation{
		ID:                 104,
		OrgID:              1,
		QuotationNumber:    "QT-2026-0004",
		Status:             QuotationStatusAccepted,
		CustomerName:       "Global Trade Ltd",
		Origin:             "Rotterdam, Netherlands",
		Destination:        "New York, USA",
		ServiceType:        "FCL",
		TransportMode:      "Ocean Freight",
		TotalAmount:        7200.00,
		ConvertedBookingID: &bookingID,
		ConversionStatus:   QuotationConversionStatusConverted,
	}
	canConvert, reasons = CanConvertQuotationToBooking(alreadyConvertedQ, nil)
	if canConvert {
		t.Errorf("Expected already converted quote to be blocked, but CanConvert returned true")
	}
}

// Mock repository for Business Logic testing
type mockQuotationRepoForConversion struct {
	Repository
	q           *Quotation
	charges     []*QuotationChargeItem
	createdHist []*QuotationConversionHistory
	convBookingID int64
}

func (m *mockQuotationRepoForConversion) GetQuotationByID(ctx context.Context, orgID, quotationID int64) (*Quotation, error) {
	return m.q, nil
}

func (m *mockQuotationRepoForConversion) GetQuotationCharges(ctx context.Context, orgID, quotationID int64) ([]*QuotationChargeItem, error) {
	return m.charges, nil
}

func (m *mockQuotationRepoForConversion) GetCustomerInfo(ctx context.Context, orgID, customerID int64) (*QuotationCustomerInfo, error) {
	return &QuotationCustomerInfo{
		ID:           customerID,
		Name:         m.q.CustomerName,
		CustomerCode: "CUST-001",
	}, nil
}

func (m *mockQuotationRepoForConversion) CreateActivity(ctx context.Context, act *QuotationActivity) error {
	return nil
}

func (m *mockQuotationRepoForConversion) CreateQuotationConversionHistory(ctx context.Context, history *QuotationConversionHistory) error {
	m.createdHist = append(m.createdHist, history)
	return nil
}

func (m *mockQuotationRepoForConversion) MarkQuotationConverted(ctx context.Context, orgID, quotationID, bookingID int64, shipmentID *int64, convertedBy, notes string) error {
	m.q.ConvertedBookingID = &bookingID
	m.q.ConvertedShipmentID = shipmentID
	m.q.ConversionStatus = QuotationConversionStatusConverted
	now := time.Now()
	m.q.ConvertedAt = &now
	m.q.ConvertedBy = convertedBy
	m.q.ConversionNotes = notes
	return nil
}

func (m *mockQuotationRepoForConversion) CreateBookingFromQuotationTx(ctx context.Context, orgID int64, q *Quotation, req *ConvertQuotationToBookingRequest, creator string) (int64, string, *int64, *string, error) {
	bID := m.convBookingID
	if bID == 0 {
		bID = 8801
	}
	bNum := "BK-TEST-001"
	var shID *int64
	var shNum *string
	if req.CreateShipmentImmediately {
		sID := int64(7701)
		sNum := "SH-TEST-001"
		shID = &sID
		shNum = &sNum
	}
	return bID, bNum, shID, shNum, nil
}

func TestConvertQuotationToBooking_SuccessAndIdempotency(t *testing.T) {
	ctx := context.Background()
	acceptedQ := &Quotation{
		ID:              501,
		OrgID:           10,
		QuotationNumber: "QT-2026-0501",
		Status:          QuotationStatusAccepted,
		CustomerName:    "Apex Logistics",
		Origin:          "Hamburg, Germany",
		Destination:     "Singapore",
		ServiceType:     "FCL",
		TransportMode:   "Ocean Freight",
		TotalAmount:     12500.00,
	}

	mockRepo := &mockQuotationRepoForConversion{
		q:             acceptedQ,
		convBookingID: 9955,
	}

	svc := NewService(mockRepo, rates.Service(nil))

	// 1. Initial conversion attempt
	req := &ConvertQuotationToBookingRequest{
		CarrierName:                 "Hapag-Lloyd",
		CarrierSCAC:                 ptrString("HLCU"),
		CreateShipmentImmediately:   true,
	}

	res, err := svc.ConvertQuotationToBooking(ctx, 10, 501, 100, req)
	if err != nil {
		t.Fatalf("Unexpected error during conversion: %v", err)
	}
	if !res.Success {
		t.Errorf("Expected conversion success, got false")
	}
	if res.BookingID != 9955 {
		t.Errorf("Expected booking ID 9955, got %d", res.BookingID)
	}
	if res.ShipmentID == nil || *res.ShipmentID != 7701 {
		t.Errorf("Expected shipment ID 7701, got %v", res.ShipmentID)
	}
	if res.AlreadyConverted {
		t.Errorf("Expected initial conversion already_converted to be false")
	}

	// 2. Repeat conversion request (Idempotency test)
	res2, err := svc.ConvertQuotationToBooking(ctx, 10, 501, 100, req)
	if err != nil {
		t.Fatalf("Unexpected error on duplicate conversion: %v", err)
	}
	if !res2.Success {
		t.Errorf("Expected idempotent repeat conversion to succeed")
	}
	if !res2.AlreadyConverted {
		t.Errorf("Expected res2.AlreadyConverted to be true")
	}
	if res2.BookingID != 9955 {
		t.Errorf("Expected existing booking ID 9955 on duplicate request, got %d", res2.BookingID)
	}
}

func ptrString(s string) *string {
	return &s
}
