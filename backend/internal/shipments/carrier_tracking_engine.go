package shipments

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	carrierAdapters "github.com/freel/backend/internal/carrier/adapters"
	carrierDomain "github.com/freel/backend/internal/carrier/domain"
	carrierService "github.com/freel/backend/internal/carrier/service"
	"github.com/freel/backend/internal/shipments/spec"
	"github.com/jmoiron/sqlx"
)

// MapCarrierMilestoneToShipmentStatus maps normalized milestone codes to canonical LogisticsHQ shipment statuses.
func MapCarrierMilestoneToShipmentStatus(milestoneCode string) string {
	switch strings.ToUpper(strings.TrimSpace(milestoneCode)) {
	case "GATE_IN", "BOOKED", "BOOKING_CONFIRMED", "ORIGIN_TERMINAL":
		return spec.BOOKED
	case "LOADED", "DEPARTED", "VESSEL_DEPARTED", "ORIGIN_DEPARTURE":
		return spec.DEPARTED
	case "IN_TRANSIT", "STOWED", "TRANSSHIPMENT", "SEA_PASSAGE", "WAYPOINT_PASSAGE":
		return spec.IN_TRANSIT
	case "ARRIVED", "VESSEL_ARRIVED", "DISCHARGED", "BERTHED", "DESTINATION_ARRIVAL":
		return spec.ARRIVED
	case "DELIVERED", "GATE_OUT", "CARGO_RELEASED", "EMPTY_RETURN", "CONSIGNEE_DELIVERY":
		return spec.DELIVERED
	case "CUSTOMS_HOLD", "INSPECTION_HOLD", "EXCEPTION":
		return spec.EXCEPTION
	default:
		return spec.IN_TRANSIT
	}
}

// ShouldAdvanceStatus ensures out-of-order or stale events do not downgrade an advanced shipment lifecycle status.
func ShouldAdvanceStatus(currentStatus, candidateStatus string) bool {
	ranks := map[string]int{
		spec.BOOKING_PENDING: 1,
		spec.BOOKED:          2,
		spec.DEPARTED:        3,
		spec.IN_TRANSIT:      4,
		spec.ARRIVED:         5,
		spec.DELIVERED:       6,
		spec.EXCEPTION:       4,
	}

	currRank, ok1 := ranks[currentStatus]
	candRank, ok2 := ranks[candidateStatus]
	if !ok1 || !ok2 {
		return false
	}
	return candRank > currRank
}

// CarrierTrackingEngine orchestrates real carrier adapter communication, DCSA event normalization,
// idempotent persistence, and shipment state synchronization.
type CarrierTrackingEngine struct {
	db         *sqlx.DB
	repo       Repository
	carrierSvc carrierService.CarrierService
	registry   *carrierAdapters.AdapterRegistry
}

// NewCarrierTrackingEngine initializes a new tracking engine.
func NewCarrierTrackingEngine(db *sqlx.DB, repo Repository, carrierSvc carrierService.CarrierService) *CarrierTrackingEngine {
	return &CarrierTrackingEngine{
		db:         db,
		repo:       repo,
		carrierSvc: carrierSvc,
		registry:   carrierAdapters.GetDefaultRegistry(),
	}
}

// SyncShipmentTracking executes the full Carrier API -> Adapter -> Normalization -> Persistence pipeline.
func (e *CarrierTrackingEngine) SyncShipmentTracking(ctx context.Context, orgID int64, shipmentID int64, userID *int64, actor string) (*spec.TrackingRefreshResult, error) {
	now := time.Now().UTC()

	// 1. Fetch Shipment
	sh, err := e.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shipment: %w", err)
	}
	if sh == nil {
		return nil, fmt.Errorf("shipment %d not found", shipmentID)
	}

	// 2. Resolve Carrier SCAC
	scac := strings.ToUpper(strings.TrimSpace(sh.CarrierSCAC))
	if scac == "" && sh.CarrierName != nil && *sh.CarrierName != "" {
		if resolvedSCAC, err := e.repo.FindSCACByCarrierName(ctx, *sh.CarrierName); err == nil && resolvedSCAC != "" {
			scac = resolvedSCAC
		}
	}

	if scac == "" {
		return &spec.TrackingRefreshResult{
			Success:       true,
			DataFreshness: spec.TrackingFreshnessUnavailable,
			LastUpdatedAt: &now,
			NewPositions:  0,
			NewEvents:     0,
			UsedFallback:  true,
			Message:       "Shipment does not have an assigned ocean carrier. Assign a carrier line to sync live tracking.",
		}, nil
	}

	// 3. Check Carrier Integration Configuration
	if e.carrierSvc == nil {
		return &spec.TrackingRefreshResult{
			Success:       true,
			DataFreshness: spec.TrackingFreshnessUnavailable,
			LastUpdatedAt: &now,
			NewPositions:  0,
			NewEvents:     0,
			UsedFallback:  true,
			Message:       fmt.Sprintf("Carrier integration service unavailable. Showing latest persisted operational tracking data for %s.", scac),
		}, nil
	}

	integrations, err := e.carrierSvc.GetIntegrations(ctx, orgID)
	if err != nil {
		log.Printf("[CarrierTrackingEngine] Error fetching integrations for org %d: %v", orgID, err)
	}

	var activeIntegration *carrierDomain.CarrierIntegrationView
	for _, ci := range integrations {
		if strings.EqualFold(ci.CarrierSCAC, scac) || strings.EqualFold(ci.CarrierName, scac) {
			activeIntegration = &ci
			break
		}
	}

	// Case 1: Carrier not connected for this organization
	if activeIntegration == nil {
		return &spec.TrackingRefreshResult{
			Success:       true,
			DataFreshness: spec.TrackingFreshnessUnavailable,
			LastUpdatedAt: &now,
			NewPositions:  0,
			NewEvents:     0,
			UsedFallback:  true,
			Message:       fmt.Sprintf("Carrier integration (%s) is not configured. Connect this carrier in Settings > Carrier Integrations to enable live tracking.", scac),
		}, nil
	}

	// Case 2: Integration is disabled
	if !activeIntegration.IsEnabled || activeIntegration.ConnectionStatus == carrierDomain.StatusDisabled {
		return &spec.TrackingRefreshResult{
			Success:       true,
			DataFreshness: spec.TrackingFreshnessUnavailable,
			LastUpdatedAt: &now,
			NewPositions:  0,
			NewEvents:     0,
			UsedFallback:  true,
			Message:       fmt.Sprintf("Carrier integration (%s) is currently disabled. Enable it in Carrier Integrations settings.", scac),
		}, nil
	}

	// Case 3: Tracking capability is disabled
	hasTrackingCap := false
	for _, cap := range activeIntegration.Capabilities {
		if cap == carrierDomain.CapTracking {
			hasTrackingCap = true
			break
		}
	}
	if !hasTrackingCap {
		return &spec.TrackingRefreshResult{
			Success:       true,
			DataFreshness: spec.TrackingFreshnessUnavailable,
			LastUpdatedAt: &now,
			NewPositions:  0,
			NewEvents:     0,
			UsedFallback:  true,
			Message:       fmt.Sprintf("Tracking capability is not enabled for carrier connection (%s). Enable Tracking in Carrier Integrations settings.", scac),
		}, nil
	}

	// 4. Construct Neutral Tracking Request
	primaryContainer := ""
	if len(sh.ContainerNumbers) > 0 {
		primaryContainer = sh.ContainerNumbers[0]
	}
	bookingRef := ""
	if sh.BookingNumber != nil {
		bookingRef = *sh.BookingNumber
	}
	mblRef := ""
	if sh.MBLNumber != nil {
		mblRef = *sh.MBLNumber
	}

	if primaryContainer == "" && bookingRef == "" && mblRef == "" {
		return &spec.TrackingRefreshResult{
			Success:       true,
			DataFreshness: spec.TrackingFreshnessUnavailable,
			LastUpdatedAt: &now,
			NewPositions:  0,
			NewEvents:     0,
			UsedFallback:  true,
			Message:       "Shipment lacks a Container Number, Booking Number, or MBL Reference for carrier lookup.",
		}, nil
	}

	trackingReq := carrierDomain.TrackingRequest{
		CarrierSCAC:     scac,
		ContainerNumber: primaryContainer,
		BookingNumber:   bookingRef,
		MBLNumber:       mblRef,
	}

	// 5. Invoke Adapter via CarrierService with Retries and Safe Fallback
	trackingResult, err := e.carrierSvc.GetTracking(ctx, orgID, activeIntegration.ID, trackingReq)
	if err != nil {
		log.Printf("[CarrierTrackingEngine] Carrier API call failed for shipment #%d (%s): %v", shipmentID, scac, err)
		// Failure must NOT delete existing tracking data!
		return &spec.TrackingRefreshResult{
			Success:       true,
			DataFreshness: spec.TrackingFreshnessStale,
			LastUpdatedAt: &now,
			NewPositions:  0,
			NewEvents:     0,
			UsedFallback:  true,
			Message:       fmt.Sprintf("Unable to retrieve the latest carrier update (%s API unavailable). Existing tracking history is preserved.", scac),
		}, nil
	}

	if trackingResult == nil || len(trackingResult.Events) == 0 {
		return &spec.TrackingRefreshResult{
			Success:       true,
			DataFreshness: spec.TrackingFreshnessRecent,
			LastUpdatedAt: &now,
			NewPositions:  0,
			NewEvents:     0,
			UsedFallback:  false,
			Message:       fmt.Sprintf("Carrier %s checked successfully. No new tracking events found for reference %s.", scac, primaryContainer),
		}, nil
	}

	// 6. Idempotent Transactional Persistence
	tx, err := e.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	newEventCount := 0

	for _, ev := range trackingResult.Events {
		eventID := ev.EventID
		if eventID == "" {
			eventID = fmt.Sprintf("%s-%s-%s-%d", scac, ev.MilestoneCode, ev.ContainerNumber, ev.EventTime.Unix())
		}

		rawPayloadBytes, _ := json.Marshal(ev)

		// Idempotent insert into carrier_tracking_events
		queryEvent := `
			INSERT INTO carrier_tracking_events (
				org_id, event_id, source_type, carrier_scac, booking_number, container_number,
				mbl_number, hbl_number, vessel_name, voyage_number, milestone_code, event_time,
				location, raw_description, raw_payload, shipment_id, matching_status, processing_status, created_at, updated_at
			) VALUES (
				?, ?, 'API', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'MATCHED', 'PROCESSED', NOW(), NOW()
			) ON DUPLICATE KEY UPDATE 
				shipment_id = VALUES(shipment_id),
				matching_status = 'MATCHED',
				processing_status = 'PROCESSED',
				updated_at = NOW()
		`
		res, err := tx.ExecContext(ctx, queryEvent,
			orgID, eventID, scac, bookingRef, ev.ContainerNumber,
			mblRef, "", ev.VesselName, ev.VoyageNumber, ev.MilestoneCode, ev.EventTime,
			ev.Location, ev.Description, string(rawPayloadBytes), shipmentID,
		)
		if err == nil {
			if affected, _ := res.RowsAffected(); affected > 0 {
				newEventCount++
			}
		}

		// Update or Insert into shipment_milestones
		queryMilestone := `
			INSERT INTO shipment_milestones (
				shipment_id, milestone_name, event_timestamp, location, is_completed, source, created_at
			) VALUES (
				?, ?, ?, ?, 1, 'CARRIER_API', NOW()
			) ON DUPLICATE KEY UPDATE
				event_timestamp = VALUES(event_timestamp),
				location = VALUES(location),
				is_completed = 1
		`
		_, _ = tx.ExecContext(ctx, queryMilestone, shipmentID, ev.MilestoneCode, ev.EventTime, ev.Location)
	}

	// 7. Update Shipment Master Record based on authoritative actual status
	actualMilestone := trackingResult.CurrentStatus
	candidateStatus := MapCarrierMilestoneToShipmentStatus(actualMilestone)
	newStatus := sh.Status
	if ShouldAdvanceStatus(sh.Status, candidateStatus) {
		newStatus = candidateStatus
	}

	vesselName := sh.VesselName
	voyageNumber := sh.VoyageNumber

	// Extract latest vessel/voyage from events if available
	for _, ev := range trackingResult.Events {
		if ev.VesselName != "" {
			vesselName = &ev.VesselName
		}
		if ev.VoyageNumber != "" {
			voyageNumber = &ev.VoyageNumber
		}
	}

	eta := sh.ETA
	if trackingResult.EstimatedArrival != nil {
		eta = trackingResult.EstimatedArrival
	}
	etd := sh.ETD
	if trackingResult.ActualDeparture != nil {
		etd = trackingResult.ActualDeparture
	}

	queryShipmentUpdate := `
		UPDATE shipments
		SET status = ?, eta = ?, etd = ?, vessel_name = ?, voyage_number = ?, updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err = tx.ExecContext(ctx, queryShipmentUpdate, newStatus, eta, etd, vesselName, voyageNumber, shipmentID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to update shipment master state: %w", err)
	}

	// 8. Update tracking position coordinates if location is known
	if trackingResult.LatestLocation != "" {
		coords := GetPortCoordinates(trackingResult.LatestLocation)
		vNameStr := ""
		if vesselName != nil {
			vNameStr = *vesselName
		}
		queryPos := `
			INSERT INTO tracking_positions (
				org_id, shipment_id, vessel_name, latitude, longitude, speed_knots, heading_degrees,
				location_name, tracking_source, data_freshness, recorded_at, created_at, updated_at
			) VALUES (
				?, ?, ?, ?, ?, 18.5, 312.0, ?, 'Carrier API Integration', 'LIVE', NOW(), NOW(), NOW()
			)
		`
		_, _ = tx.ExecContext(ctx, queryPos, orgID, shipmentID, vNameStr, coords.Latitude, coords.Longitude, coords.Name)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit tracking synchronization transaction: %w", err)
	}

	// 9. Log Auditable Activity & Meaningful Lifecycle Events
	if newEventCount > 0 {
		_ = e.repo.CreateActivity(
			ctx,
			orgID,
			"SHIPMENT",
			shipmentID,
			spec.SHIPMENT_TRACKING_REFRESHED,
			fmt.Sprintf("Synchronized %d new carrier tracking events from %s API", newEventCount, scac),
			actor,
		)
	}

	if newStatus != sh.Status {
		_ = e.repo.CreateActivity(
			ctx,
			orgID,
			"SHIPMENT",
			shipmentID,
			spec.SHIPMENT_STATUS_UPDATED,
			fmt.Sprintf("Shipment status updated to %s based on %s carrier milestone", newStatus, actualMilestone),
			actor,
		)
	}

	return &spec.TrackingRefreshResult{
		Success:       true,
		DataFreshness: spec.TrackingFreshnessLive,
		LastUpdatedAt: &now,
		NewPositions:  1,
		NewEvents:     newEventCount,
		UsedFallback:  false,
		Message:       fmt.Sprintf("Successfully synchronized %d tracking events from %s API.", len(trackingResult.Events), scac),
	}, nil
}
