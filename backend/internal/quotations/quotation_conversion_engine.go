package quotations

import (
	"fmt"
	"strings"
	"time"
)

// CanConvertQuotationToBooking evaluates business rules to determine if a quotation can be converted to an operational booking.
func CanConvertQuotationToBooking(q *Quotation, charges []*QuotationChargeItem) (bool, []string) {
	var blockingReasons []string

	if q == nil {
		return false, []string{"Quotation record is missing or nil"}
	}

	// 1. Organization & Lifecycle Status Check
	if q.OrgID <= 0 {
		blockingReasons = append(blockingReasons, "Invalid organization ownership")
	}

	if q.Status != QuotationStatusAccepted {
		blockingReasons = append(blockingReasons, fmt.Sprintf("Quotation must be in ACCEPTED status to convert (current status: %s)", q.Status))
	}

	// 2. Already Converted Check
	if q.ConvertedBookingID != nil || q.ConversionStatus == QuotationConversionStatusConverted {
		blockingReasons = append(blockingReasons, fmt.Sprintf("Quotation has already been converted to Booking #%d", *q.ConvertedBookingID))
	}

	// 3. Customer Information Check
	if strings.TrimSpace(q.CustomerName) == "" && (q.CustomerID == nil || *q.CustomerID <= 0) {
		blockingReasons = append(blockingReasons, "Customer information is required for commercial conversion")
	}

	// 4. Origin & Destination Route Check
	origin := strings.TrimSpace(q.Origin)
	if origin == "" {
		origin = strings.TrimSpace(q.OriginCode)
	}
	if origin == "" {
		blockingReasons = append(blockingReasons, "Origin port or facility is required for booking handover")
	}

	dest := strings.TrimSpace(q.Destination)
	if dest == "" {
		dest = strings.TrimSpace(q.DestinationCode)
	}
	if dest == "" {
		blockingReasons = append(blockingReasons, "Destination port or facility is required for booking handover")
	}

	// 5. Transport & Service Mode Check
	if strings.TrimSpace(q.TransportMode) == "" && strings.TrimSpace(q.ServiceType) == "" {
		blockingReasons = append(blockingReasons, "Transport mode or service type must be specified")
	}

	// 6. Pricing & Commercial Totals Check
	if q.TotalAmount < 0 {
		blockingReasons = append(blockingReasons, "Commercial total amount cannot be negative")
	}

	// 7. Cancellation or Expiry Check
	if q.CancelledAt != nil {
		blockingReasons = append(blockingReasons, "Quotation has been cancelled and cannot be converted")
	}

	if q.ValidUntil != nil && q.ValidUntil.Before(time.Now().Truncate(24*time.Hour)) && q.Status != QuotationStatusAccepted {
		blockingReasons = append(blockingReasons, "Quotation validity period has expired")
	}

	return len(blockingReasons) == 0, blockingReasons
}

// BuildQuotationConversionPreview constructs a read-only snapshot preview of the handover.
func BuildQuotationConversionPreview(q *Quotation, charges []*QuotationChargeItem, cust *QuotationCustomerInfo) *QuotationConversionPreview {
	if q == nil {
		return nil
	}

	canConvert, blockingReasons := CanConvertQuotationToBooking(q, charges)

	equipment := "Standard"
	if strings.Contains(strings.ToUpper(q.ServiceType), "FCL") || strings.Contains(strings.ToUpper(q.TransportMode), "OCEAN") {
		equipment = "40' Standard Container (40GP)"
	} else if strings.Contains(strings.ToUpper(q.ServiceType), "LCL") {
		equipment = "LCL Consolidate"
	} else if strings.Contains(strings.ToUpper(q.TransportMode), "AIR") {
		equipment = "Air Freight Unit (ULD)"
	}

	customerCode := ""
	if cust != nil {
		customerCode = cust.CustomerCode
	}

	convStatus := q.ConversionStatus
	if convStatus == "" {
		if canConvert {
			convStatus = QuotationConversionStatusReady
		} else {
			convStatus = QuotationConversionStatusNotConverted
		}
	}

	preview := &QuotationConversionPreview{
		QuotationID:        q.ID,
		QuotationNumber:    q.QuotationNumber,
		CustomerID:         q.CustomerID,
		CustomerName:       q.CustomerName,
		CustomerCode:       customerCode,
		Origin:             q.Origin,
		OriginCode:         q.OriginCode,
		Destination:        q.Destination,
		DestinationCode:    q.DestinationCode,
		ServiceType:        q.ServiceType,
		TransportMode:      q.TransportMode,
		Equipment:          equipment,
		Currency:           q.Currency,
		PaymentTerms:       q.PaymentTerms,
		Subtotal:           q.Subtotal,
		Surcharges:         q.Surcharges,
		Taxes:              q.Taxes,
		TotalAmount:        q.TotalAmount,
		ValidUntil:         q.ValidUntil,
		AcceptedAt:         q.AcceptedAt,
		CommercialTerms:    q.CommercialTerms,
		CustomerNotes:      q.CustomerNotes,
		SelectedCharges:    charges,
		CanConvert:         canConvert,
		BlockingReasons:    blockingReasons,
		ConversionStatus:   convStatus,
		ExistingBookingID:  q.ConvertedBookingID,
		ExistingShipmentID: q.ConvertedShipmentID,
	}

	return preview
}
