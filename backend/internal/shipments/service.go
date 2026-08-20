package shipments

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/freel/backend/internal/common/events"
	"github.com/jmoiron/sqlx"
)

type Service interface {
	CreateFromRFQ(ctx context.Context, rfqID int64) (*Shipment, error)
	GetShipmentByID(ctx context.Context, orgID int64, id int64) (*Shipment, error)
	ListShipments(ctx context.Context, orgID int64) ([]*Shipment, error)
	UpdateShipment(ctx context.Context, s *Shipment) error
	GetMilestones(ctx context.Context, shipmentID int64) ([]*ShipmentMilestone, error)
	UpdateMilestone(ctx context.Context, orgID int64, shipmentID int64, milestoneCode string, actualDate *time.Time, location *string, notes *string) error
	GetExceptions(ctx context.Context, shipmentID int64) ([]*ShipmentException, error)
	CreateException(ctx context.Context, orgID int64, shipmentID int64, exType string, severity string, title string, description string, sourceEventID *string) error
	ResolveException(ctx context.Context, orgID int64, exceptionID int64) error
	HandleInboundCarrierEvent(ctx context.Context, orgID int64, event *NormalizedTrackingEvent) error
	CompleteCarrierEvent(ctx context.Context, eventID string, orgID int64, shipmentID int64, hasCritical bool, aiSummary string) error
}

type service struct {
	repo           Repository
	db             *sqlx.DB
	eventBus       events.Bus
	backendBaseURL string // e.g. "http://backend:8080" — no trailing slash
}

func NewService(repo Repository, db *sqlx.DB, eventBus events.Bus, backendBaseURL string) Service {
	return &service{
		repo:           repo,
		db:             db,
		eventBus:       eventBus,
		backendBaseURL: backendBaseURL,
	}
}

func (s *service) CreateFromRFQ(ctx context.Context, rfqID int64) (*Shipment, error) {
	log.Printf("[Shipment Service] Creating shipment automatically from Won RFQ #%d", rfqID)

	// Check idempotency: check if shipment already exists for this RFQ
	existing, err := s.repo.GetShipmentByRFQID(ctx, rfqID)
	if err == nil && existing != nil {
		log.Printf("[Shipment Service] Shipment already exists for RFQ #%d (ID=%d). Returning existing.", rfqID, existing.ID)
		return existing, nil
	}

	// 1. Fetch RFQ details (using direct query to avoid org_id mismatch on system-level event trigger)
	var rfqRow struct {
		OrgID       int64   `db:"org_id"`
		Origin      *string `db:"origin"`
		Destination *string `db:"destination"`
		Incoterms   *string `db:"incoterms"`
		TargetDate  *time.Time
	}
	err = s.db.GetContext(ctx, &rfqRow, "SELECT org_id, origin, destination, incoterms, target_date FROM rfqs WHERE id = ?", rfqID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rfq: %w", err)
	}

	// 2. Fetch the APPROVED quote for this RFQ
	var quoteRow struct {
		ID                int64   `db:"id"`
		CarrierName       string  `db:"carrier_name"`
		TransitTimeDays   *int    `db:"transit_time_days"`
		SellPrice         float64 `db:"sell_price"`
		VesselName        *string `db:"vessel_name"`
		VoyageNumber      *string `db:"voyage_number"`
		FreeDays          *int    `db:"free_days"`
		ContainerQuantity *int    `db:"container_quantity"`
	}

	// Find the approved quote
	err = s.db.GetContext(ctx, &quoteRow, "SELECT id, carrier_name, transit_time_days, sell_price FROM rfq_quotes WHERE rfq_id = ? AND status = 'APPROVED' LIMIT 1", rfqID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no approved quote found for rfq %d", rfqID)
		}
		return nil, fmt.Errorf("failed to fetch approved quote: %w", err)
	}

	// Lookup carrier SCAC (Fail Hard as per User instruction)
	scac, err := s.repo.FindSCACByCarrierName(ctx, quoteRow.CarrierName)
	if err != nil {
		return nil, fmt.Errorf("unable to auto-create shipment: carrier '%s' is not mapped to any known carrier SCAC in carriers database: %w", quoteRow.CarrierName, err)
	}

	origin := ""
	if rfqRow.Origin != nil {
		origin = *rfqRow.Origin
	}
	dest := ""
	if rfqRow.Destination != nil {
		dest = *rfqRow.Destination
	}

	transitDays := 14
	if quoteRow.TransitTimeDays != nil {
		transitDays = *quoteRow.TransitTimeDays
	}

	// Booking starts with NULL values for carrier allocation variables until confirmed
	shipment := &Shipment{
		OrgID:            rfqRow.OrgID,
		RFQID:            &rfqID,
		QuoteID:          &quoteRow.ID,
		CarrierSCAC:      scac,
		BookingNumber:    nil,
		MBLNumber:        nil,
		HBLNumber:        nil,
		ContainerNumbers: []string{},
		Status:           "BOOKING_PENDING",
		OriginPort:       origin,
		DestinationPort:  dest,
		VesselName:       nil,
		VoyageNumber:     nil,
		ETD:              nil,
		ETA:              nil,
	}

	// Perform database operations within a single transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	err = s.repo.CreateShipmentTx(ctx, tx, shipment)
	if err != nil {
		return nil, fmt.Errorf("failed to create shipment in transaction: %w", err)
	}

	// Seed 5 standard milestones: BOOKED, DEPARTED, IN_TRANSIT, ARRIVED, DELIVERED
	milestones := []struct {
		code  string
		desc  string
		delay int
	}{
		{"BOOKED", "Booking confirmed by shipping line", 0},
		{"DEPARTED", "Vessel departed origin port", 7},
		{"IN_TRANSIT", "Vessel in transit", 14},
		{"ARRIVED", "Vessel arrived at destination port", 7 + transitDays},
		{"DELIVERED", "Cargo delivered to final consignee", 9 + transitDays},
	}

	for _, m := range milestones {
		planned := time.Now().AddDate(0, 0, m.delay)
		ms := &ShipmentMilestone{
			ShipmentID:    shipment.ID,
			MilestoneCode: m.code,
			Description:   &m.desc,
			PlannedDate:   &planned,
			Status:        "PLANNED",
		}
		err = s.repo.CreateMilestoneTx(ctx, tx, ms)
		if err != nil {
			return nil, fmt.Errorf("failed to seed milestone %s in transaction: %w", m.code, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[Shipment Service] Shipment #%d successfully created and seeded with 5 milestones.", shipment.ID)

	s.eventBus.Publish(events.Event{
		Type:      "shipment.created",
		Payload:   map[string]interface{}{"shipment_id": shipment.ID, "rfq_id": rfqID, "org_id": rfqRow.OrgID},
		Timestamp: time.Now(),
	})

	return shipment, nil
}

func (s *service) GetShipmentByID(ctx context.Context, orgID int64, id int64) (*Shipment, error) {
	return s.repo.GetShipmentByID(ctx, orgID, id)
}

func (s *service) ListShipments(ctx context.Context, orgID int64) ([]*Shipment, error) {
	return s.repo.ListShipments(ctx, orgID)
}

func (s *service) UpdateShipment(ctx context.Context, sh *Shipment) error {
	return s.repo.UpdateShipment(ctx, sh)
}

func (s *service) GetMilestones(ctx context.Context, shipmentID int64) ([]*ShipmentMilestone, error) {
	return s.repo.GetMilestones(ctx, shipmentID)
}

func (s *service) UpdateMilestone(ctx context.Context, orgID int64, shipmentID int64, milestoneCode string, actualDate *time.Time, location *string, notes *string) error {
	milestones, err := s.repo.GetMilestones(ctx, shipmentID)
	if err != nil {
		return err
	}

	var targetMilestone *ShipmentMilestone
	for _, m := range milestones {
		if m.MilestoneCode == milestoneCode {
			targetMilestone = m
			break
		}
	}

	if targetMilestone == nil {
		return fmt.Errorf("milestone %s not found for shipment %d", milestoneCode, shipmentID)
	}

	// Decision 2: Prevent milestone actual-date regression. Preserve raw event but do not overwrite newer date.
	if actualDate != nil && targetMilestone.ActualDate != nil && actualDate.Before(*targetMilestone.ActualDate) {
		log.Printf("[Shipment Service] Ignored milestone actual date update: incoming date %s is older than current milestone actual date %s",
			actualDate.Format(time.RFC3339), targetMilestone.ActualDate.Format(time.RFC3339))
		return nil
	}

	targetMilestone.ActualDate = actualDate
	targetMilestone.Status = "COMPLETED"
	if location != nil {
		targetMilestone.Location = location
	}
	if notes != nil {
		targetMilestone.Notes = notes
	}

	err = s.repo.UpdateMilestone(ctx, targetMilestone)
	if err != nil {
		return err
	}

	// Update shipment status based on latest milestone code progression rules
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err == nil && sh != nil {
		milestoneOrder := map[string]int{
			"BOOKING_PENDING": 0,
			"BOOKED":          1,
			"DEPARTED":        2,
			"IN_TRANSIT":      3,
			"ARRIVED":         4,
			"DELIVERED":       5,
		}

		currentRank := milestoneOrder[sh.Status]
		newRank, exists := milestoneOrder[milestoneCode]
		// Only advance status if the target milestone represents progression (prevents delays setting status backwards)
		if exists && newRank > currentRank {
			sh.Status = milestoneCode
			_ = s.repo.UpdateShipment(ctx, sh)
		}
	}

	return nil
}

func (s *service) GetExceptions(ctx context.Context, shipmentID int64) ([]*ShipmentException, error) {
	return s.repo.GetExceptions(ctx, shipmentID)
}

func (s *service) CreateException(ctx context.Context, orgID int64, shipmentID int64, exType string, severity string, title string, description string, sourceEventID *string) error {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil || sh == nil {
		return fmt.Errorf("shipment %d not found: %w", shipmentID, err)
	}

	ex := &ShipmentException{
		ShipmentID:    shipmentID,
		ExceptionType: exType,
		Severity:      severity,
		Title:         title,
		Description:   &description,
		Resolved:      false,
		SourceEventID: sourceEventID,
	}

	err = s.repo.CreateException(ctx, ex)
	if err != nil {
		// 8. Exception idempotency: Ignore duplicate exception updates gracefully
		if strings.Contains(err.Error(), "uq_exception_event") || strings.Contains(err.Error(), "unique constraint") {
			log.Printf("[Shipment Service] Exception %s already recorded for shipment %d. Skipping duplicate.", exType, shipmentID)
			return nil
		}
		return err
	}

	// Update shipment status to EXCEPTION for operational visibility
	sh.Status = "EXCEPTION"
	_ = s.repo.UpdateShipment(ctx, sh)

	return nil
}

func (s *service) ResolveException(ctx context.Context, orgID int64, exceptionID int64) error {
	// First resolve the exception enforcing org_id ownership
	err := s.repo.ResolveException(ctx, orgID, exceptionID, time.Now())
	if err != nil {
		return err
	}

	// Get shipment details to restore its status
	var shipmentID int64
	err = s.db.GetContext(ctx, &shipmentID, `
		SELECT shipment_id FROM shipment_exceptions 
		WHERE id = ? AND shipment_id IN (SELECT id FROM shipments WHERE org_id = ?)
	`, exceptionID, orgID)
	if err == nil {
		// Reset shipment status to the latest completed milestone or BOOKED
		milestones, err := s.repo.GetMilestones(ctx, shipmentID)
		if err == nil {
			status := "BOOKING_PENDING"
			for _, m := range milestones {
				if m.Status == "COMPLETED" {
					status = m.MilestoneCode
				}
			}
			sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
			if err == nil && sh != nil {
				sh.Status = status
				_ = s.repo.UpdateShipment(ctx, sh)
			}
		}
	}

	return nil
}

func (s *service) HandleInboundCarrierEvent(ctx context.Context, orgID int64, event *NormalizedTrackingEvent) error {
	// 1. Persist the raw carrier event first (ON CONFLICT DO NOTHING) to prevent data loss
	rawPayloadJSON, _ := event.RawPayload.MarshalJSON()
	rawEv := &CarrierTrackingEvent{
		OrgID:            orgID,
		EventID:          event.EventID,
		SourceType:       event.SourceType,
		CarrierSCAC:      event.CarrierSCAC,
		BookingNumber:    event.BookingNumber,
		ContainerNumber:  event.ContainerNumber,
		MBLNumber:        event.MBLNumber,
		HBLNumber:        event.HBLNumber,
		VesselName:       event.VesselName,
		VoyageNumber:     event.VoyageNumber,
		MilestoneCode:    event.MilestoneCode,
		EventTime:        event.EventTime,
		Location:         event.Location,
		RawDescription:   event.Description,
		RawPayload:       rawPayloadJSON,
		MatchingStatus:   "UNMATCHED",
		ProcessingStatus: "RECEIVED",
		ReceivedAt:       time.Now(),
		UpdatedAt:        time.Now(),
	}

	// 6. DB transaction atomic block for Event Ingestion + Task Queueing
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Persist the raw carrier event first inside tx context to prevent data loss
	queryInsert := `
		INSERT INTO carrier_tracking_events (
			org_id, event_id, source_type, carrier_scac, booking_number, container_number,
			mbl_number, hbl_number, vessel_name, voyage_number, milestone_code, event_time,
			location, raw_description, raw_payload, shipment_id, matching_status, processing_status, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
		) ON DUPLICATE KEY UPDATE updated_at = NOW()
	`
	res, err := tx.ExecContext(ctx, queryInsert,
		rawEv.OrgID, rawEv.EventID, rawEv.SourceType, rawEv.CarrierSCAC, rawEv.BookingNumber, rawEv.ContainerNumber,
		rawEv.MBLNumber, rawEv.HBLNumber, rawEv.VesselName, rawEv.VoyageNumber, rawEv.MilestoneCode, rawEv.EventTime,
		rawEv.Location, rawEv.RawDescription, rawEv.RawPayload, rawEv.ShipmentID, rawEv.MatchingStatus, rawEv.ProcessingStatus,
	)
	if err != nil {
		return fmt.Errorf("failed to persist raw carrier event inside transaction: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		log.Printf("[Shipment Service] Event %s already received in transaction. Skipping.", event.EventID)
		return nil
	}

	// 2. Identify Shipment using prioritized matching Strategy with proper AMBIGUOUS/UNMATCHED handling
	var matchedShipments []*Shipment

	// Priority 1: Booking Number
	if event.BookingNumber != "" {
		matchedShipments, err = s.repo.FindShipmentsByBooking(ctx, orgID, event.BookingNumber)
		if err != nil {
			return err
		}
	}

	// Priority 2: Container Number (fallback)
	if len(matchedShipments) == 0 && event.ContainerNumber != "" {
		matchedShipments, err = s.repo.FindShipmentsByContainer(ctx, orgID, event.ContainerNumber)
		if err != nil {
			return err
		}
	}

	// Priority 3: MBL (fallback)
	if len(matchedShipments) == 0 && event.MBLNumber != "" {
		matchedShipments, err = s.repo.FindShipmentsByMBL(ctx, orgID, event.MBLNumber)
		if err != nil {
			return err
		}
	}

	// Priority 4: HBL (fallback)
	if len(matchedShipments) == 0 && event.HBLNumber != "" {
		matchedShipments, err = s.repo.FindShipmentsByHBL(ctx, orgID, event.HBLNumber)
		if err != nil {
			return err
		}
	}

	var shipmentID *int64
	matchingStatus := "UNMATCHED"
	entityID := ""

	// 4. Hardened Matching check
	if len(matchedShipments) == 1 {
		shipment := matchedShipments[0]
		shipmentID = &shipment.ID
		matchingStatus = "MATCHED"
		entityID = fmt.Sprintf("%d", shipment.ID)
	} else if len(matchedShipments) > 1 {
		matchingStatus = "AMBIGUOUS"
		log.Printf("[Shipment Service] Ambiguous matches (%d shipments) found for event %s. Holding event.", len(matchedShipments), event.EventID)
	}

	// Update carrier tracking event status to MATCHED/AMBIGUOUS/UNMATCHED in transaction
	queryUpdate := `
		UPDATE carrier_tracking_events
		SET matching_status = ?, processing_status = ?, shipment_id = ?, updated_at = NOW()
		WHERE event_id = ? AND org_id = ?
	`
	_, err = tx.ExecContext(ctx, queryUpdate, matchingStatus, "QUEUED", shipmentID, event.EventID, orgID)
	if err != nil {
		return fmt.Errorf("failed to update event status: %w", err)
	}

	// 3. Queue task in ai_processing_tasks (Only if matched cleanly, else preserve for HITL re-matching review)
	if matchingStatus == "MATCHED" {
		payload := map[string]interface{}{
			"event_id":         event.EventID,
			"org_id":           orgID,
			"source_type":      event.SourceType,
			"carrier_scac":     event.CarrierSCAC,
			"booking_number":   event.BookingNumber,
			"container_number": event.ContainerNumber,
			"mbl_number":       event.MBLNumber,
			"hbl_number":       event.HBLNumber,
			"vessel_name":      event.VesselName,
			"voyage_number":    event.VoyageNumber,
			"milestone_code":   event.MilestoneCode,
			"event_time":       event.EventTime.Format(time.RFC3339),
			"location":         event.Location,
			"description":      event.Description,
			"callback_url":     s.backendBaseURL + "/internal/operations/callback",
		}

		queryQueue := `
			INSERT INTO ai_processing_tasks (org_id, entity_type, entity_id, task_type, payload, status, created_at, updated_at)
			VALUES (?, 'SHIPMENT', ?, 'CARRIER_UPDATE_PARSE', ?, 'QUEUED', NOW(), NOW())
		`
		payloadJSON, _ := json.Marshal(payload)
		_, err = tx.ExecContext(ctx, queryQueue, orgID, entityID, string(payloadJSON))
		if err != nil {
			return fmt.Errorf("failed to queue task: %w", err)
		}
	} else {
		// Event remains as UNMATCHED/AMBIGUOUS, processing_status stays RECEIVED/HELD to prevent processing wrong shipment
		queryHeld := `
			UPDATE carrier_tracking_events
			SET processing_status = 'HELD', updated_at = NOW()
			WHERE event_id = ? AND org_id = ?
		`
		_, _ = tx.ExecContext(ctx, queryHeld, event.EventID, orgID)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	log.Printf("[Shipment Service] Handled event %s. Matching Status: %s, Entity: %s", event.EventID, matchingStatus, entityID)
	return nil
}

// 17. Clean handler callback structures using CompleteCarrierEvent
func (s *service) CompleteCarrierEvent(ctx context.Context, eventID string, orgID int64, shipmentID int64, hasCritical bool, aiSummary string) error {
	// Mark processing_status = PROCESSED in carrier_tracking_events and update associated shipment
	return s.repo.UpdateCarrierEventStatus(ctx, eventID, orgID, "MATCHED", "PROCESSED", &shipmentID)
}
