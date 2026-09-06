package quotations

import (
	"fmt"
	"strings"
	"time"
)

// RawOperationalBooking captures raw booking attributes from the DB for lineage & drift detection.
type RawOperationalBooking struct {
	ID                     int64      `db:"id"`
	OrgID                  int64      `db:"org_id"`
	BookingNumber          string     `db:"booking_number"`
	Status                 string     `db:"status"`
	CarrierName            *string    `db:"carrier_name"`
	CarrierSCAC            *string    `db:"carrier_scac"`
	OriginPort             string     `db:"origin_port"`
	DestinationPort        string     `db:"destination_port"`
	VesselName             *string    `db:"vessel_name"`
	VoyageNumber           *string    `db:"voyage_number"`
	ETD                    *time.Time `db:"etd"`
	ETA                    *time.Time `db:"eta"`
	CargoSummary           *string    `db:"cargo_summary"`
	SpecialInstructions    *string    `db:"special_instructions"`
	CommercialHandoverStatus string   `db:"commercial_handover_status"`
	CommercialSnapshotAt   *time.Time `db:"commercial_snapshot_at"`
	ConfirmedAt            *time.Time `db:"confirmed_at"`
	ConfirmedBy            *string    `db:"confirmed_by"`
	CreatedAt              time.Time  `db:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at"`
}

// RawOperationalShipment captures raw shipment attributes from the DB.
type RawOperationalShipment struct {
	ID                     int64      `db:"id"`
	OrgID                  int64      `db:"org_id"`
	BookingID              int64      `db:"booking_id"`
	BookingNumber          string     `db:"booking_number"`
	CarrierSCAC            string     `db:"carrier_scac"`
	OriginPort             string     `db:"origin_port"`
	DestinationPort        string     `db:"destination_port"`
	VesselName             *string    `db:"vessel_name"`
	VoyageNumber           *string    `db:"voyage_number"`
	ETD                    *time.Time `db:"etd"`
	ETA                    *time.Time `db:"eta"`
	Status                 string     `db:"status"`
	TrackingStatus         *string    `db:"tracking_status"`
	CreatedAt              time.Time  `db:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at"`
}

// BuildQuotationOperationalHandover synthesizes the complete lineage and status.
func BuildQuotationOperationalHandover(
	q *Quotation,
	booking *RawOperationalBooking,
	shipment *RawOperationalShipment,
) *QuotationOperationalHandover {
	if q == nil {
		return nil
	}

	snapshot := &QuotationCommercialSnapshot{
		QuotationID:     q.ID,
		QuotationNumber: q.QuotationNumber,
		CustomerID:      q.CustomerID,
		CustomerName:    q.CustomerName,
		AcceptedAt:      q.AcceptedAt,
		Currency:        q.Currency,
		Subtotal:        q.Subtotal,
		Surcharges:      q.Surcharges,
		Taxes:           q.Taxes,
		TotalAmount:     q.TotalAmount,
		PaymentTerms:    q.PaymentTerms,
		CommercialTerms: q.CommercialTerms,
		CustomerNotes:   q.CustomerNotes,
		Origin:          q.Origin,
		OriginCode:      q.OriginCode,
		Destination:     q.Destination,
		DestinationCode: q.DestinationCode,
		ServiceType:     q.ServiceType,
		TransportMode:   q.TransportMode,
	}

	handoverStatus := CalculateHandoverStatus(q, booking, shipment)
	changes := DetectOperationalChanges(q, booking, shipment)
	chain := BuildLineageChain(q, booking, shipment)

	var bID *int64
	var bNumber string
	var shID *int64
	var shNumber string
	var confirmedAt *time.Time
	var confirmedBy string
	var currCarrier string
	var currVessel string
	var currVoyage string
	var currETD *time.Time
	var currETA *time.Time
	var trackingStatus string

	if booking != nil {
		bID = &booking.ID
		bNumber = booking.BookingNumber
		confirmedAt = booking.ConfirmedAt
		if booking.ConfirmedBy != nil {
			confirmedBy = *booking.ConfirmedBy
		}
		if booking.CarrierName != nil {
			currCarrier = *booking.CarrierName
		} else if booking.CarrierSCAC != nil {
			currCarrier = *booking.CarrierSCAC
		}
		if booking.VesselName != nil {
			currVessel = *booking.VesselName
		}
		if booking.VoyageNumber != nil {
			currVoyage = *booking.VoyageNumber
		}
		currETD = booking.ETD
		currETA = booking.ETA
	}

	if shipment != nil {
		shID = &shipment.ID
		shNumber = fmt.Sprintf("SH-%d", shipment.ID)
		if shipment.TrackingStatus != nil && *shipment.TrackingStatus != "" {
			trackingStatus = *shipment.TrackingStatus
		} else {
			trackingStatus = shipment.Status
		}
		if currCarrier == "" && shipment.CarrierSCAC != "" {
			currCarrier = shipment.CarrierSCAC
		}
		if currVessel == "" && shipment.VesselName != nil {
			currVessel = *shipment.VesselName
		}
		if currVoyage == "" && shipment.VoyageNumber != nil {
			currVoyage = *shipment.VoyageNumber
		}
		if currETD == nil {
			currETD = shipment.ETD
		}
		if currETA == nil {
			currETA = shipment.ETA
		}
	}

	canConfirm := booking != nil && (booking.CommercialHandoverStatus == CommercialHandoverPending || booking.CommercialHandoverStatus == CommercialHandoverConverted || booking.CommercialHandoverStatus == "")
	var blockingReasons []string
	if booking == nil {
		canConfirm = false
		blockingReasons = append(blockingReasons, "Quotation has not yet been converted into an operational booking")
	} else if booking.CommercialHandoverStatus == CommercialHandoverBookingConfirmed || booking.CommercialHandoverStatus == CommercialHandoverCompleted {
		canConfirm = false
		blockingReasons = append(blockingReasons, "Booking handover is already confirmed")
	}

	var convertedByStr string
	if q.ConvertedBy != "" {
		convertedByStr = q.ConvertedBy
	}

	return &QuotationOperationalHandover{
		QuotationID:        q.ID,
		QuotationNumber:    q.QuotationNumber,
		RFQID:              q.RFQID,
		RFQNumber:          q.RFQNumber,
		BookingID:          bID,
		BookingNumber:      bNumber,
		ShipmentID:         shID,
		ShipmentNumber:     shNumber,
		CustomerName:       q.CustomerName,
		ConversionStatus:   q.ConversionStatus,
		HandoverStatus:     handoverStatus,
		CommercialSnapshot: snapshot,
		OperationalChanges: changes,
		LineageChain:       chain,
		ConvertedAt:        q.ConvertedAt,
		ConvertedBy:        convertedByStr,
		BookingConfirmedAt: confirmedAt,
		BookingConfirmedBy: confirmedBy,
		CanConfirmHandover: canConfirm,
		BlockingReasons:    blockingReasons,
		CurrentCarrier:     currCarrier,
		CurrentVessel:      currVessel,
		CurrentVoyage:      currVoyage,
		CurrentETD:         currETD,
		CurrentETA:         currETA,
		TrackingStatus:     trackingStatus,
	}
}

// CalculateHandoverStatus derives the handover status from quotation, booking, and shipment lifecycle state.
func CalculateHandoverStatus(q *Quotation, b *RawOperationalBooking, s *RawOperationalShipment) string {
	if q == nil {
		return CommercialHandoverPending
	}
	if q.ConvertedBookingID == nil && q.ConversionStatus != QuotationConversionStatusConverted {
		return CommercialHandoverPending
	}
	if b == nil {
		return CommercialHandoverConverted
	}
	if b.CommercialHandoverStatus == CommercialHandoverBookingConfirmed {
		if s != nil && (s.Status == "COMPLETED" || s.Status == "DELIVERED" || s.Status == "ARRIVED") {
			return CommercialHandoverCompleted
		}
		return CommercialHandoverBookingConfirmed
	}
	if b.CommercialHandoverStatus == CommercialHandoverCompleted {
		return CommercialHandoverCompleted
	}
	return CommercialHandoverConverted
}

// DetectOperationalChanges compares quotation baseline values against current operational booking/shipment data.
func DetectOperationalChanges(q *Quotation, b *RawOperationalBooking, s *RawOperationalShipment) []*OperationalChange {
	var changes []*OperationalChange
	if q == nil || b == nil {
		return changes
	}

	// 1. Origin Route comparison
	qOrigin := strings.TrimSpace(q.Origin)
	bOrigin := strings.TrimSpace(b.OriginPort)
	if bOrigin != "" && qOrigin != "" && !strings.EqualFold(qOrigin, bOrigin) && !strings.EqualFold(q.OriginCode, bOrigin) {
		changes = append(changes, &OperationalChange{
			Field:          "Origin Port",
			Category:       "ROUTING",
			BaselineValue:  qOrigin,
			CurrentValue:   bOrigin,
			ChangedAt:      b.UpdatedAt,
			ChangedBy:      "Operations",
			IsCommercial:   false,
			ImpactSeverity: "WARNING",
		})
	}

	// 2. Destination Route comparison
	qDest := strings.TrimSpace(q.Destination)
	bDest := strings.TrimSpace(b.DestinationPort)
	if bDest != "" && qDest != "" && !strings.EqualFold(qDest, bDest) && !strings.EqualFold(q.DestinationCode, bDest) {
		changes = append(changes, &OperationalChange{
			Field:          "Destination Port",
			Category:       "ROUTING",
			BaselineValue:  qDest,
			CurrentValue:   bDest,
			ChangedAt:      b.UpdatedAt,
			ChangedBy:      "Operations",
			IsCommercial:   false,
			ImpactSeverity: "WARNING",
		})
	}

	// 3. Vessel / Flight comparison
	if b.VesselName != nil && *b.VesselName != "" {
		if s != nil && s.VesselName != nil && *s.VesselName != "" && !strings.EqualFold(*b.VesselName, *s.VesselName) {
			changes = append(changes, &OperationalChange{
				Field:          "Vessel / Flight",
				Category:       "CARRIER",
				BaselineValue:  *b.VesselName,
				CurrentValue:   *s.VesselName,
				ChangedAt:      s.UpdatedAt,
				ChangedBy:      "Carrier Dispatch",
				IsCommercial:   false,
				ImpactSeverity: "INFO",
			})
		}
	}

	// 4. ETD Schedule comparison
	if b.ETD != nil && s != nil && s.ETD != nil {
		if !b.ETD.Equal(*s.ETD) {
			changes = append(changes, &OperationalChange{
				Field:          "Estimated Departure (ETD)",
				Category:       "SCHEDULE",
				BaselineValue:  b.ETD.Format("02 Jan 2006"),
				CurrentValue:   s.ETD.Format("02 Jan 2006"),
				ChangedAt:      s.UpdatedAt,
				ChangedBy:      "Carrier Tracking",
				IsCommercial:   false,
				ImpactSeverity: "WARNING",
			})
		}
	}

	// 5. ETA Schedule comparison
	if b.ETA != nil && s != nil && s.ETA != nil {
		if !b.ETA.Equal(*s.ETA) {
			changes = append(changes, &OperationalChange{
				Field:          "Estimated Arrival (ETA)",
				Category:       "SCHEDULE",
				BaselineValue:  b.ETA.Format("02 Jan 2006"),
				CurrentValue:   s.ETA.Format("02 Jan 2006"),
				ChangedAt:      s.UpdatedAt,
				ChangedBy:      "Carrier Tracking",
				IsCommercial:   false,
				ImpactSeverity: "WARNING",
			})
		}
	}

	return changes
}

// BuildLineageChain constructs the visual 5-step lineage chain.
func BuildLineageChain(q *Quotation, b *RawOperationalBooking, s *RawOperationalShipment) []*LineageStep {
	steps := []*LineageStep{
		{
			StepID:      "QUOTE",
			Name:        "Commercial Quote",
			Status:      "COMPLETED",
			ReferenceID: &q.ID,
			RefNumber:   q.QuotationNumber,
			URL:         fmt.Sprintf("/dashboard/quotations?id=%d", q.ID),
			Timestamp:   &q.CreatedAt,
			Actor:       q.CreatedBy,
		},
	}

	// Quote Acceptance Step
	acceptStatus := "PENDING"
	var acceptTime *time.Time
	if q.Status == QuotationStatusAccepted {
		acceptStatus = "COMPLETED"
		acceptTime = q.AcceptedAt
	}
	steps = append(steps, &LineageStep{
		StepID:    "ACCEPTANCE",
		Name:      "Customer Acceptance",
		Status:    acceptStatus,
		Timestamp: acceptTime,
	})

	// Booking Step
	bookingStatus := "PENDING"
	var bID *int64
	var bNum string
	var bURL string
	var bTime *time.Time
	if b != nil {
		bID = &b.ID
		bNum = b.BookingNumber
		bURL = "/dashboard/bookings"
		bTime = &b.CreatedAt
		if b.CommercialHandoverStatus == CommercialHandoverBookingConfirmed || b.CommercialHandoverStatus == CommercialHandoverCompleted {
			bookingStatus = "COMPLETED"
		} else {
			bookingStatus = "ACTIVE"
		}
	}
	steps = append(steps, &LineageStep{
		StepID:      "BOOKING",
		Name:        "Operational Booking",
		Status:      bookingStatus,
		ReferenceID: bID,
		RefNumber:   bNum,
		URL:         bURL,
		Timestamp:   bTime,
	})

	// Shipment Step
	shipmentStatus := "PENDING"
	var sID *int64
	var sNum string
	var sURL string
	var sTime *time.Time
	if s != nil {
		sID = &s.ID
		sNum = fmt.Sprintf("SH-%d", s.ID)
		sURL = "/dashboard/shipments"
		sTime = &s.CreatedAt
		shipmentStatus = "ACTIVE"
		if s.Status == "COMPLETED" || s.Status == "DELIVERED" {
			shipmentStatus = "COMPLETED"
		}
	}
	steps = append(steps, &LineageStep{
		StepID:      "SHIPMENT",
		Name:        "Active Shipment",
		Status:      shipmentStatus,
		ReferenceID: sID,
		RefNumber:   sNum,
		URL:         sURL,
		Timestamp:   sTime,
	})

	// Live Tracking Step
	trackingStatus := "PENDING"
	var tURL string
	if s != nil {
		tURL = "/dashboard/tracking"
		if s.TrackingStatus != nil && *s.TrackingStatus != "" {
			trackingStatus = "ACTIVE"
		}
	}
	steps = append(steps, &LineageStep{
		StepID: "TRACKING",
		Name:   "Milestone Tracking",
		Status: trackingStatus,
		URL:    tURL,
	})

	return steps
}
