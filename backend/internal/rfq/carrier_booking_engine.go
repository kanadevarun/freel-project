package rfq

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	carrierDomain "github.com/freel/backend/internal/carrier/domain"
	carrierService "github.com/freel/backend/internal/carrier/service"
	"github.com/freel/backend/internal/rfq/spec"
)

// CarrierBookingEngine orchestrates live carrier bookings for LogisticsHQ bookings.
type CarrierBookingEngine struct {
	dl         Datalayer
	carrierSvc carrierService.CarrierService
}

// NewCarrierBookingEngine creates a new carrier booking engine.
func NewCarrierBookingEngine(dl Datalayer, carrierSvc carrierService.CarrierService) *CarrierBookingEngine {
	return &CarrierBookingEngine{
		dl:         dl,
		carrierSvc: carrierSvc,
	}
}

// ResolveCarrierSCAC attempts to map carrier name or input SCAC to standard 4-letter carrier SCAC.
func ResolveCarrierSCAC(scac, name string) string {
	cleanedSCAC := strings.ToUpper(strings.TrimSpace(scac))
	if len(cleanedSCAC) == 4 {
		return cleanedSCAC
	}

	cleanedName := strings.ToUpper(strings.TrimSpace(name))
	switch {
	case strings.Contains(cleanedName, "MAERSK") || cleanedSCAC == "MAEU":
		return "MAEU"
	case strings.Contains(cleanedName, "MSC") || strings.Contains(cleanedName, "MEDITERRANEAN") || cleanedSCAC == "MSCU":
		return "MSCU"
	case strings.Contains(cleanedName, "HAPAG") || cleanedSCAC == "HLCU":
		return "HLCU"
	case strings.Contains(cleanedName, "CMA") || strings.Contains(cleanedName, "CGM") || cleanedSCAC == "CMDU":
		return "CMDU"
	case strings.Contains(cleanedName, "ONE") || strings.Contains(cleanedName, "OCEAN NETWORK") || cleanedSCAC == "ONEY":
		return "ONEY"
	case strings.Contains(cleanedName, "EVERGREEN") || cleanedSCAC == "EGLV":
		return "EGLV"
	case strings.Contains(cleanedName, "COSCO") || cleanedSCAC == "COSU":
		return "COSU"
	case strings.Contains(cleanedName, "YANG MING") || cleanedSCAC == "YMLU":
		return "YMLU"
	case strings.Contains(cleanedName, "ZIM") || cleanedSCAC == "ZIMU":
		return "ZIMU"
	case strings.Contains(cleanedName, "HMM") || cleanedSCAC == "HDMU":
		return "HDMU"
	default:
		if cleanedSCAC != "" {
			return cleanedSCAC
		}
		return ""
	}
}

// SubmitCarrierBooking submits an existing LogisticsHQ booking to the connected carrier API.
func (e *CarrierBookingEngine) SubmitCarrierBooking(ctx context.Context, orgID int32, bookingID int64, req spec.BookWithCarrierRequest, user string) (*spec.BookingDetailResponse, error) {
	if e.carrierSvc == nil {
		return nil, fmt.Errorf("carrier integration service is not available")
	}

	// 1. Retrieve the existing booking
	booking, err := e.dl.GetBookingByIDOnly(ctx, orgID, bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking %d not found for organization", bookingID)
	}

	// 2. Resolve target carrier SCAC
	requestedSCAC := ""
	if req.CarrierSCAC != nil && *req.CarrierSCAC != "" {
		requestedSCAC = *req.CarrierSCAC
	} else if booking.CarrierSCAC != nil {
		requestedSCAC = *booking.CarrierSCAC
	}
	carrierSCAC := ResolveCarrierSCAC(requestedSCAC, booking.CarrierName)
	if carrierSCAC == "" {
		return nil, fmt.Errorf("unable to determine carrier SCAC for '%s'. Please specify carrier SCAC", booking.CarrierName)
	}

	// 3. Find active carrier integration for this tenant and SCAC
	integrations, err := e.carrierSvc.GetIntegrations(ctx, int64(orgID))
	if err != nil {
		return nil, fmt.Errorf("failed to check carrier integrations: %w", err)
	}

	var targetIntegration *carrierDomain.CarrierIntegrationView
	for i := range integrations {
		ci := &integrations[i]
		if strings.EqualFold(ci.CarrierSCAC, carrierSCAC) || strings.EqualFold(ci.CarrierName, booking.CarrierName) {
			targetIntegration = ci
			break
		}
	}

	if targetIntegration == nil {
		return nil, fmt.Errorf("carrier integration for %s (%s) is not configured. Connect this carrier in Settings > Carrier Integrations", booking.CarrierName, carrierSCAC)
	}

	if !targetIntegration.IsEnabled || targetIntegration.ConnectionStatus == carrierDomain.StatusDisabled {
		return nil, fmt.Errorf("carrier integration for %s is currently disabled. Enable it in Settings > Carrier Integrations", targetIntegration.CarrierName)
	}

	// 4. Verify BOOKING capability is enabled
	hasBookingCap := false
	for _, cap := range targetIntegration.Capabilities {
		if cap == carrierDomain.CapBooking {
			hasBookingCap = true
			break
		}
	}
	if !hasBookingCap {
		return nil, fmt.Errorf("booking capability is not enabled for carrier %s. Enable Booking in Settings > Carrier Integrations", targetIntegration.CarrierName)
	}

	// 5. Idempotency & Duplicate Protection
	if !req.ForceRetry && booking.CarrierBookingReference != nil && *booking.CarrierBookingReference != "" {
		if booking.CarrierBookingStatus != nil && (*booking.CarrierBookingStatus == "CONFIRMED" || *booking.CarrierBookingStatus == "PENDING_ALLOCATION") {
			log.Printf("[CarrierBookingEngine] Booking %d already booked with carrier %s (Ref: %s). Returning existing booking.", bookingID, carrierSCAC, *booking.CarrierBookingReference)
			return e.dl.GetBookingWorkspaceDetail(ctx, orgID, bookingID)
		}
	}

	// 6. Build normalized carrier BookingRequest
	now := time.Now().UTC()
	readyDate := now.AddDate(0, 0, 7)
	if req.CargoReadyDate != nil && !req.CargoReadyDate.IsZero() {
		readyDate = *req.CargoReadyDate
	} else if booking.ETD != nil && !booking.ETD.IsZero() {
		readyDate = *booking.ETD
	}

	eqType := "40HC"
	if req.EquipmentType != nil && *req.EquipmentType != "" {
		eqType = *req.EquipmentType
	}
	quantity := 1
	if req.Quantity != nil && *req.Quantity > 0 {
		quantity = *req.Quantity
	}

	contractNum := ""
	if req.ContractNumber != nil {
		contractNum = *req.ContractNumber
	}

	commodity := "General Cargo"
	if req.Commodity != nil && *req.Commodity != "" {
		commodity = *req.Commodity
	} else if booking.CargoSummary != nil && *booking.CargoSummary != "" {
		commodity = *booking.CargoSummary
	}

	shipper := "LogisticsHQ Shipper"
	if req.ShipperName != nil && *req.ShipperName != "" {
		shipper = *req.ShipperName
	}
	consignee := "LogisticsHQ Consignee"
	if req.ConsigneeName != nil && *req.ConsigneeName != "" {
		consignee = *req.ConsigneeName
	}

	carrierReq := carrierDomain.BookingRequest{
		CarrierSCAC:       carrierSCAC,
		ContractNumber:    contractNum,
		OriginPort:        strings.ToUpper(strings.TrimSpace(booking.OriginPort)),
		DestinationPort:   strings.ToUpper(strings.TrimSpace(booking.DestinationPort)),
		EquipmentType:     eqType,
		Quantity:          quantity,
		CargoReadyDate:    readyDate,
		Commodity:         commodity,
		ShipperName:       shipper,
		ConsigneeName:     consignee,
		CustomerReference: booking.BookingNumber,
	}

	// 7. Invoke carrier adapter
	bookingResult, err := e.carrierSvc.CreateBooking(ctx, int64(orgID), targetIntegration.ID, carrierReq)
	if err != nil {
		log.Printf("[CarrierBookingEngine] Carrier booking failed with %s: %v", carrierSCAC, err)
		errStr := err.Error()
		carrierStatus := "REJECTED"

		// Record rejection in database while preserving booking record
		_ = e.dl.UpdateCarrierBookingResult(ctx, orgID, bookingID, "", carrierStatus, nil, &errStr, nil, nil, nil, nil, nil, booking.Status)
		_ = e.dl.CreateActivity(ctx, orgID, "BOOKING", bookingID, "CARRIER_BOOKING_REJECTED",
			fmt.Sprintf("Carrier booking request rejected by %s: %s", targetIntegration.CarrierName, errStr), user)

		return nil, fmt.Errorf("carrier booking rejected by %s: %w", targetIntegration.CarrierName, err)
	}

	// 8. Carrier Confirmed: update booking record
	carrierRef := bookingResult.BookingNumber
	if carrierRef == "" {
		carrierRef = bookingResult.ConfirmationRef
	}
	if carrierRef == "" {
		carrierRef = fmt.Sprintf("%s-%s", carrierSCAC, booking.BookingNumber)
	}

	carrierStatus := bookingResult.Status
	if carrierStatus == "" {
		carrierStatus = "CONFIRMED"
	}

	var confRef *string
	if bookingResult.ConfirmationRef != "" {
		confRef = &bookingResult.ConfirmationRef
	}

	bookedAt := time.Now().UTC()

	var vesselName *string
	if bookingResult.VesselName != "" {
		vesselName = &bookingResult.VesselName
	}
	var voyageNum *string
	if bookingResult.VoyageNumber != "" {
		voyageNum = &bookingResult.VoyageNumber
	}

	etd := bookingResult.ETD
	eta := bookingResult.ETA

	newStatus := spec.BookingStatusConfirmed

	err = e.dl.UpdateCarrierBookingResult(ctx, orgID, bookingID, carrierRef, carrierStatus, confRef, nil, &bookedAt, vesselName, voyageNum, etd, eta, newStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to save carrier booking confirmation: %w", err)
	}

	// 9. Log auditable activity
	_ = e.dl.CreateActivity(ctx, orgID, "BOOKING", bookingID, "CARRIER_BOOKING_CONFIRMED",
		fmt.Sprintf("Carrier booking confirmed with %s (Ref: %s, Status: %s)", targetIntegration.CarrierName, carrierRef, carrierStatus), user)

	return e.dl.GetBookingWorkspaceDetail(ctx, orgID, bookingID)
}

// SyncCarrierBooking queries the carrier API for updated allocation / voyage details.
func (e *CarrierBookingEngine) SyncCarrierBooking(ctx context.Context, orgID int32, bookingID int64, user string) (*spec.BookingDetailResponse, error) {
	if e.carrierSvc == nil {
		return nil, fmt.Errorf("carrier integration service is not available")
	}

	booking, err := e.dl.GetBookingByIDOnly(ctx, orgID, bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking %d not found for organization", bookingID)
	}

	if booking.CarrierBookingReference == nil || *booking.CarrierBookingReference == "" {
		return nil, fmt.Errorf("booking has not been submitted to a carrier yet. Use 'Book with Carrier' first")
	}

	carrierSCAC := ResolveCarrierSCAC("", booking.CarrierName)
	if booking.CarrierSCAC != nil && *booking.CarrierSCAC != "" {
		carrierSCAC = *booking.CarrierSCAC
	}

	integrations, err := e.carrierSvc.GetIntegrations(ctx, int64(orgID))
	if err != nil {
		return nil, fmt.Errorf("failed to check carrier integrations: %w", err)
	}

	var targetIntegration *carrierDomain.CarrierIntegrationView
	for i := range integrations {
		ci := &integrations[i]
		if strings.EqualFold(ci.CarrierSCAC, carrierSCAC) || strings.EqualFold(ci.CarrierName, booking.CarrierName) {
			targetIntegration = ci
			break
		}
	}

	if targetIntegration == nil {
		return nil, fmt.Errorf("carrier integration for %s is not configured", booking.CarrierName)
	}

	bookingResult, err := e.carrierSvc.GetBooking(ctx, int64(orgID), targetIntegration.ID, *booking.CarrierBookingReference)
	if err != nil {
		log.Printf("[CarrierBookingEngine] Failed to sync booking with %s: %v", targetIntegration.CarrierName, err)
		return nil, fmt.Errorf("failed to fetch latest carrier booking state from %s: %w", targetIntegration.CarrierName, err)
	}

	carrierStatus := bookingResult.Status
	if carrierStatus == "" {
		carrierStatus = "CONFIRMED"
	}

	var confRef *string
	if bookingResult.ConfirmationRef != "" {
		confRef = &bookingResult.ConfirmationRef
	}

	var vesselName *string
	if bookingResult.VesselName != "" {
		vesselName = &bookingResult.VesselName
	}
	var voyageNum *string
	if bookingResult.VoyageNumber != "" {
		voyageNum = &bookingResult.VoyageNumber
	}

	_ = e.dl.UpdateCarrierBookingResult(ctx, orgID, bookingID, *booking.CarrierBookingReference, carrierStatus, confRef, nil, booking.CarrierBookedAt, vesselName, voyageNum, bookingResult.ETD, bookingResult.ETA, booking.Status)
	_ = e.dl.CreateActivity(ctx, orgID, "BOOKING", bookingID, "CARRIER_BOOKING_SYNCED",
		fmt.Sprintf("Synchronized latest carrier booking status with %s (Status: %s)", targetIntegration.CarrierName, carrierStatus), user)

	return e.dl.GetBookingWorkspaceDetail(ctx, orgID, bookingID)
}
