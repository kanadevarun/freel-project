package shipments

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
	carrierService "github.com/freel/backend/internal/carrier/service"
	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/shipments/spec"
	"github.com/freel/backend/internal/svcerror"
	"github.com/jmoiron/sqlx"
)

type Service interface {
	CreateFromRFQ(ctx context.Context, rfqID int64) (*spec.Shipment, error)
	GetShipmentByID(ctx context.Context, orgID int64, id int64) (*spec.Shipment, error)
	ListShipments(ctx context.Context, orgID int64) ([]*spec.Shipment, error)
	UpdateShipment(ctx context.Context, s *spec.Shipment) error
	GetMilestones(ctx context.Context, shipmentID int64) ([]*spec.ShipmentMilestone, error)
	UpdateMilestone(ctx context.Context, orgID int64, shipmentID int64, milestoneCode string, actualDate *time.Time, location *string, notes *string) error
	GetShipmentExceptions(ctx context.Context, orgID int64, shipmentID int64) ([]*spec.ShipmentException, error)
	CreateShipmentException(ctx context.Context, orgID int64, shipmentID int64, exType string, severity string, title string, description string, sourceEventID *string) error
	UpdateShipmentException(ctx context.Context, orgID int64, shipmentID int64, exceptionID int64, status string, severity string, notes *string) error
	AcknowledgeShipmentException(ctx context.Context, orgID int64, shipmentID int64, exceptionID int64) error
	ResolveShipmentException(ctx context.Context, orgID int64, shipmentID int64, exceptionID int64, notes string, resolvedBy int64) error
	ResolveException(ctx context.Context, orgID int64, exceptionID int64) error
	DismissShipmentException(ctx context.Context, orgID int64, shipmentID int64, exceptionID int64) error
	EvaluateShipmentExceptions(ctx context.Context, orgID int64, shipmentID int64) error
	HandleInboundCarrierEvent(ctx context.Context, orgID int64, event *spec.NormalizedTrackingEvent) error
	CompleteCarrierEvent(ctx context.Context, eventID string, orgID int64, shipmentID int64, hasCritical bool, aiSummary string) error
	GetShipmentsWorkspace(ctx context.Context, orgID int64, filter spec.ShipmentListFilter) ([]*spec.Shipment, spec.ShipmentKPIs, int, error)
	GetShipmentTracking(ctx context.Context, orgID int64, shipmentID int64) (*spec.ShipmentTrackingSummary, error)
	EvaluateClosure(ctx context.Context, orgID int64, shipmentID int64) (string, error)
	RequestClosure(ctx context.Context, orgID int64, shipmentID int64) error
	CompleteShipment(ctx context.Context, orgID int64, shipmentID int64) error
	ReopenShipment(ctx context.Context, orgID int64, shipmentID int64) error

	// Document Management (Task 16.7)
	GetShipmentDocuments(ctx context.Context, orgID int64, shipmentID int64) ([]*spec.ShipmentDocument, *spec.ShipmentDocumentComplianceSummary, []*spec.ShipmentDocumentDiscrepancy, error)
	CreateShipmentDocument(ctx context.Context, orgID int64, shipmentID int64, req spec.CreateShipmentDocumentRequest, uploader string, userID *int64) (*spec.ShipmentDocument, error)
	UpdateShipmentDocument(ctx context.Context, orgID int64, shipmentID int64, docID int64, req spec.UpdateShipmentDocumentRequest, reviewer string, userID *int64) (*spec.ShipmentDocument, error)
	ApproveShipmentDocument(ctx context.Context, orgID int64, shipmentID int64, docID int64, reviewer string, userID *int64) (*spec.ShipmentDocument, error)
	RejectShipmentDocument(ctx context.Context, orgID int64, shipmentID int64, docID int64, reason string, reviewer string, userID *int64) (*spec.ShipmentDocument, error)
	DeleteShipmentDocument(ctx context.Context, orgID int64, shipmentID int64, docID int64, userID *int64) error

	// Financial Operations (Task 16.8)
	GetShipmentFinancials(ctx context.Context, orgID int64, shipmentID int64) (*spec.ShipmentFinancialSummary, error)
	GetShipmentCharges(ctx context.Context, orgID int64, shipmentID int64) ([]*spec.ShipmentFinancialCharge, error)
	CreateShipmentCharge(ctx context.Context, orgID int64, shipmentID int64, req *spec.CreateShipmentChargeRequest, actor string) (*spec.ShipmentFinancialCharge, *spec.ShipmentFinancialSummary, error)
	UpdateShipmentCharge(ctx context.Context, orgID int64, shipmentID int64, chargeID int64, req *spec.UpdateShipmentChargeRequest, actor string) (*spec.ShipmentFinancialCharge, *spec.ShipmentFinancialSummary, error)
	DeleteShipmentCharge(ctx context.Context, orgID int64, shipmentID int64, chargeID int64, actor string) (*spec.ShipmentFinancialSummary, error)
	RecalculateShipmentFinancials(ctx context.Context, orgID int64, shipmentID int64, actor string) (*spec.ShipmentFinancialSummary, error)
	ReviewShipmentFinancials(ctx context.Context, orgID int64, shipmentID int64, status string, notes string, actor string) (*spec.ShipmentFinancialSummary, error)

	// Real-Time Tracking Telemetry & Routes (Task 17.3)
	GetLatestTrackingPosition(ctx context.Context, orgID int64, shipmentID int64) (*spec.TrackingPosition, error)
	GetTrackingPositionHistory(ctx context.Context, orgID int64, shipmentID int64, limit int) ([]spec.TrackingPosition, error)
	GetTrackingRoute(ctx context.Context, orgID int64, shipmentID int64) (*spec.TrackingRoute, error)
	GetTrackingEventsList(ctx context.Context, orgID int64, shipmentID int64) ([]spec.TrackingEventNormalized, error)

	// Tracking Event Intelligence & Operational Alerts (Task 17.4)
	GetShipmentTrackingIntelligence(ctx context.Context, orgID int64, shipmentID int64) (*spec.ShipmentTrackingIntelligence, error)

	// Tracking Automation, Alert Notifications & Operational Monitoring (Task 17.5)
	GetTrackingAlerts(ctx context.Context, orgID int64, shipmentID int64, status string) ([]*spec.ShipmentTrackingAlertRecord, error)
	GetTrackingMonitoringSummary(ctx context.Context, orgID int64, shipmentID int64) (*spec.TrackingMonitoringSummary, error)
	AcknowledgeTrackingAlert(ctx context.Context, orgID int64, shipmentID int64, alertID int64, userID *int64, actor string) error
	ResolveTrackingAlert(ctx context.Context, orgID int64, shipmentID int64, alertID int64, userID *int64, notes string, actor string) error
	SuppressTrackingAlert(ctx context.Context, orgID int64, shipmentID int64, alertID int64, userID *int64, reason string, actor string) error
	RefreshShipmentTracking(ctx context.Context, orgID int64, shipmentID int64, userID *int64, actor string) (*spec.TrackingRefreshResult, error)
	GetTrackingRefreshHistory(ctx context.Context, orgID int64, shipmentID int64, limit int) ([]*spec.TrackingRefreshRunRecord, error)

	// Tracking Analytics & Operational Intelligence (Task 17.8)
	GetTrackingAnalyticsOverview(ctx context.Context, orgID int64) (*spec.TrackingAnalyticsOverview, error)
	GetTrackingAnalyticsTrends(ctx context.Context, orgID int64, days int) ([]spec.TrackingTrendDataPoint, error)
	GetCarrierTrackingPerformance(ctx context.Context, orgID int64) ([]spec.CarrierTrackingPerformance, error)
	GetRouteTrackingPerformance(ctx context.Context, orgID int64) ([]spec.RouteTrackingPerformance, error)

	// Carrier Integration Engine (Task 4)
	SetCarrierService(carrierSvc carrierService.CarrierService)
}

type serviceImpl struct {
	repo                  Repository
	db                    *sqlx.DB
	eventBus              events.Bus
	backendBaseURL        string // e.g. "http://backend:8080" — no trailing slash
	carrierTrackingEngine *CarrierTrackingEngine
}

func NewService(repo Repository, db *sqlx.DB, eventBus events.Bus, backendBaseURL string) Service {
	return &serviceImpl{
		repo:           repo,
		db:             db,
		eventBus:       eventBus,
		backendBaseURL: backendBaseURL,
	}
}

func (s *serviceImpl) SetCarrierService(carrierSvc carrierService.CarrierService) {
	s.carrierTrackingEngine = NewCarrierTrackingEngine(s.db, s.repo, carrierSvc)
}

func (s *serviceImpl) CreateFromRFQ(ctx context.Context, rfqID int64) (*spec.Shipment, error) {
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
		TargetDate  *time.Time `db:"target_date"`
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
	shipment := &spec.Shipment{
		OrgID:            rfqRow.OrgID,
		RFQID:            &rfqID,
		QuoteID:          &quoteRow.ID,
		CarrierSCAC:      scac,
		BookingNumber:    nil,
		MBLNumber:        nil,
		HBLNumber:        nil,
		ContainerNumbers: []string{},
		Status:           spec.BOOKING_PENDING,
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
		{spec.BOOKED, "Booking confirmed by shipping line", 0},
		{spec.DEPARTED, "Vessel departed origin port", 7},
		{spec.IN_TRANSIT, "Vessel in transit", 14},
		{spec.ARRIVED, "Vessel arrived at destination port", 7 + transitDays},
		{spec.DELIVERED, "Cargo delivered to final consignee", 9 + transitDays},
	}

	for _, m := range milestones {
		planned := time.Now().AddDate(0, 0, m.delay)
		ms := &spec.ShipmentMilestone{
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

	if s.eventBus != nil {
		s.eventBus.Publish(events.Event{
			Type:      "shipment.created",
			Payload:   map[string]interface{}{"shipment_id": shipment.ID, "rfq_id": rfqID, "org_id": rfqRow.OrgID},
			Timestamp: time.Now(),
		})
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        rfqRow.OrgID,
		Action:       domain.ActionCreate,
		Module:       domain.ModuleShipments,
		ResourceType: "SHIPMENT",
		ResourceID:   fmt.Sprintf("%d", shipment.ID),
		Description:  fmt.Sprintf("Created shipment #%d (%s → %s)", shipment.ID, origin, dest),
		Result:       domain.ResultSuccess,
	})

	return shipment, nil
}

func (s *serviceImpl) GetShipmentByID(ctx context.Context, orgID int64, id int64) (*spec.Shipment, error) {
	return s.repo.GetShipmentByID(ctx, orgID, id)
}

func (s *serviceImpl) ListShipments(ctx context.Context, orgID int64) ([]*spec.Shipment, error) {
	return s.repo.ListShipments(ctx, orgID)
}

func (s *serviceImpl) GetShipmentsWorkspace(ctx context.Context, orgID int64, filter spec.ShipmentListFilter) ([]*spec.Shipment, spec.ShipmentKPIs, int, error) {
	return s.repo.GetShipmentsWorkspace(ctx, orgID, filter)
}

func (s *serviceImpl) UpdateShipment(ctx context.Context, sh *spec.Shipment) error {
	err := s.repo.UpdateShipment(ctx, sh)
	if err == nil && sh != nil {
		_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
			OrgID:        sh.OrgID,
			Action:       domain.ActionUpdate,
			Module:       domain.ModuleShipments,
			ResourceType: "SHIPMENT",
			ResourceID:   fmt.Sprintf("%d", sh.ID),
			Description:  fmt.Sprintf("Updated shipment #%d (Status: %s)", sh.ID, sh.Status),
			Result:       domain.ResultSuccess,
		})
	}
	return err
}

func (s *serviceImpl) GetMilestones(ctx context.Context, shipmentID int64) ([]*spec.ShipmentMilestone, error) {
	return s.repo.GetMilestones(ctx, shipmentID)
}

func (s *serviceImpl) UpdateMilestone(ctx context.Context, orgID int64, shipmentID int64, milestoneCode string, actualDate *time.Time, location *string, notes *string) error {
	// 1. Validate shipment exists and belongs to the caller's organization
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return err
	}
	if sh == nil {
		return fmt.Errorf("shipment %d not found or access denied", shipmentID)
	}

	milestones, err := s.repo.GetMilestones(ctx, shipmentID)
	if err != nil {
		return err
	}

	var targetMilestone *spec.ShipmentMilestone
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
	var finalActualDate time.Time
	if actualDate != nil && !actualDate.IsZero() {
		finalActualDate = *actualDate
	} else {
		finalActualDate = time.Now()
	}

	if targetMilestone.ActualDate != nil && finalActualDate.Before(*targetMilestone.ActualDate) {
		log.Printf("[Shipment Service] Ignored milestone actual date update: incoming date %s is older than current milestone actual date %s",
			finalActualDate.Format(time.RFC3339), targetMilestone.ActualDate.Format(time.RFC3339))
		return nil
	}

	targetMilestone.ActualDate = &finalActualDate
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
	milestoneOrder := map[string]int{
		spec.BOOKING_PENDING: 0,
		spec.BOOKED:          1,
		spec.DEPARTED:        2,
		spec.IN_TRANSIT:      3,
		spec.ARRIVED:         4,
		spec.DELIVERED:       5,
	}

	currentRank := milestoneOrder[sh.Status]
	newRank, exists := milestoneOrder[milestoneCode]
	// Only advance status if the target milestone represents progression (prevents delays setting status backwards)
	if exists && newRank > currentRank {
		sh.Status = milestoneCode
		_ = s.repo.UpdateShipment(ctx, sh)

		// Record unified activity logs
		_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_STATUS_UPDATED,
			fmt.Sprintf("Shipment status updated to %s", milestoneCode),
			"Logistics Operations",
		)
		if sh.BookingID != nil {
			_ = s.repo.CreateActivity(ctx, orgID, "BOOKING", *sh.BookingID, spec.SHIPMENT_STATUS_UPDATED,
				fmt.Sprintf("Shipment status updated to %s", milestoneCode),
				"Logistics Operations",
			)
		}
	}

	return nil
}

func (s *serviceImpl) GetShipmentExceptions(ctx context.Context, orgID int64, shipmentID int64) ([]*spec.ShipmentException, error) {
	// First check shipment ownership
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	if sh == nil {
		return nil, fmt.Errorf("shipment %d not found or access denied", shipmentID)
	}
	return s.repo.GetExceptions(ctx, orgID, shipmentID)
}

func (s *serviceImpl) CreateShipmentException(ctx context.Context, orgID int64, shipmentID int64, exType string, severity string, title string, description string, sourceEventID *string) error {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return err
	}
	if sh == nil {
		return fmt.Errorf("shipment %d not found or access denied", shipmentID)
	}

	// Validate inputs
	if exType == "" || severity == "" || title == "" {
		return fmt.Errorf("missing required exception fields: type, severity, or title")
	}

	validSeverities := map[string]bool{"LOW": true, "MEDIUM": true, "HIGH": true, "CRITICAL": true}
	if !validSeverities[severity] {
		return fmt.Errorf("invalid severity level: %s", severity)
	}

	validCategories := map[string]bool{
		"SCHEDULE_DELAY": true, "ETD_DELAY": true, "ETA_DELAY": true, "VESSEL_ROLLOVER": true,
		"PORT_CONGESTION": true, "CUSTOMS_HOLD": true, "DOCUMENT_ISSUE": true, "CARRIER_DELAY": true,
		"ROUTE_DEVIATION": true, "CONTAINER_ISSUE": true, "OTHER": true,
	}
	if !validCategories[exType] {
		return fmt.Errorf("invalid exception category: %s", exType)
	}

	ex := &spec.ShipmentException{
		OrgID:         orgID,
		ShipmentID:    shipmentID,
		ExceptionType: exType,
		Severity:      severity,
		Status:        spec.ExceptionStatusOpen,
		Title:         title,
		Description:   &description,
		Resolved:      false,
		SourceEventID: sourceEventID,
	}

	err = s.repo.CreateException(ctx, ex)
	if err != nil {
		// Idempotency: Ignore duplicate exception updates gracefully
		if strings.Contains(err.Error(), "uq_exception_type_shipment") || strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "Duplicate entry") {
			log.Printf("[Shipment Service] Exception %s with source %v already recorded for shipment %d. Skipping duplicate.", exType, sourceEventID, shipmentID)
			return nil
		}
		return err
	}

	// We do NOT change the main shipment status to EXCEPTION automatically.
	// But we log the activity timeline action
	_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_EXCEPTION_CREATED,
		fmt.Sprintf("New %s Exception raised: %s", severity, title),
		"Logistics Operations",
	)

	return nil
}

func (s *serviceImpl) UpdateShipmentException(ctx context.Context, orgID int64, shipmentID int64, exceptionID int64, status string, severity string, notes *string) error {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return err
	}
	if sh == nil {
		return fmt.Errorf("shipment %d not found or access denied", shipmentID)
	}

	ex, err := s.repo.GetExceptionByID(ctx, orgID, exceptionID)
	if err != nil {
		return err
	}
	if ex == nil || ex.ShipmentID != shipmentID {
		return fmt.Errorf("exception %d not found or mismatch on shipment %d", exceptionID, shipmentID)
	}

	// Prevent updates on already resolved/dismissed exceptions
	if ex.Status == spec.ExceptionStatusResolved || ex.Status == spec.ExceptionStatusDismissed {
		return fmt.Errorf("cannot update an exception in final state: %s", ex.Status)
	}

	// Validate status transition
	validStatuses := map[string]bool{
		spec.ExceptionStatusOpen: true,
		spec.ExceptionStatusAcknowledged: true,
		spec.ExceptionStatusInProgress: true,
	}
	if status != "" && !validStatuses[status] {
		return fmt.Errorf("invalid status transition: %s", status)
	}

	if status != "" {
		ex.Status = status
	}
	if severity != "" {
		ex.Severity = severity
	}
	if notes != nil {
		ex.Description = notes
	}

	return s.repo.UpdateException(ctx, ex)
}

func (s *serviceImpl) AcknowledgeShipmentException(ctx context.Context, orgID int64, shipmentID int64, exceptionID int64) error {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return err
	}
	if sh == nil {
		return fmt.Errorf("shipment %d not found or access denied", shipmentID)
	}

	ex, err := s.repo.GetExceptionByID(ctx, orgID, exceptionID)
	if err != nil {
		return err
	}
	if ex == nil || ex.ShipmentID != shipmentID {
		return fmt.Errorf("exception %d not found or mismatch on shipment %d", exceptionID, shipmentID)
	}

	if ex.Status == spec.ExceptionStatusResolved || ex.Status == spec.ExceptionStatusDismissed {
		return fmt.Errorf("cannot acknowledge an exception that is already %s", ex.Status)
	}

	ex.Status = spec.ExceptionStatusAcknowledged
	err = s.repo.UpdateException(ctx, ex)
	if err != nil {
		return err
	}

	_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_EXCEPTION_ACKNOWLEDGED,
		fmt.Sprintf("Exception acknowledged: %s", ex.Title),
		"Logistics Operations",
	)

	return nil
}

func (s *serviceImpl) ResolveShipmentException(ctx context.Context, orgID int64, shipmentID int64, exceptionID int64, notes string, resolvedBy int64) error {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return err
	}
	if sh == nil {
		return fmt.Errorf("shipment %d not found or access denied", shipmentID)
	}

	ex, err := s.repo.GetExceptionByID(ctx, orgID, exceptionID)
	if err != nil {
		return err
	}
	if ex == nil || ex.ShipmentID != shipmentID {
		return fmt.Errorf("exception %d not found or mismatch on shipment %d", exceptionID, shipmentID)
	}

	if ex.Status == spec.ExceptionStatusResolved || ex.Status == spec.ExceptionStatusDismissed {
		return fmt.Errorf("cannot resolve an exception that is already %s", ex.Status)
	}

	now := time.Now()
	ex.Status = spec.ExceptionStatusResolved
	ex.Resolved = true
	ex.ResolvedAt = &now
	ex.ResolvedBy = &resolvedBy
	ex.ResolutionNotes = &notes

	err = s.repo.UpdateException(ctx, ex)
	if err != nil {
		return err
	}

	_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_EXCEPTION_RESOLVED,
		fmt.Sprintf("Exception resolved: %s (Notes: %s)", ex.Title, notes),
		"Logistics Operations",
	)

	// Restore shipment status to the latest completed milestone if there are no other active exceptions
	milestones, err := s.repo.GetMilestones(ctx, shipmentID)
	if err == nil {
		status := spec.BOOKING_PENDING
		for _, m := range milestones {
			if m.Status == "COMPLETED" {
				status = m.MilestoneCode
			}
		}
		var activeCount int
		err = s.db.GetContext(ctx, &activeCount, `
			SELECT COUNT(*) FROM shipment_exceptions 
			WHERE shipment_id = ? AND status NOT IN ('RESOLVED', 'DISMISSED')
		`, shipmentID)
		if err == nil && activeCount == 0 {
			sh.Status = status
			_ = s.repo.UpdateShipment(ctx, sh)

			_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_STATUS_UPDATED,
				fmt.Sprintf("All operational exceptions resolved. Restored shipment status to %s", status),
				"Logistics Operations",
			)
		}
	}

	return nil
}

func (s *serviceImpl) ResolveException(ctx context.Context, orgID int64, exceptionID int64) error {
	var shipmentID int64
	err := s.db.GetContext(ctx, &shipmentID, `
		SELECT shipment_id FROM shipment_exceptions WHERE id = ? AND org_id = ?
	`, exceptionID, orgID)
	if err != nil {
		return err
	}
	userCtx, ok := middleware.GetUserContext(ctx)
	var userID int64
	if ok {
		userID = userCtx.UserID
	}
	return s.ResolveShipmentException(ctx, orgID, shipmentID, exceptionID, "Resolved via legacy interface", userID)
}

func (s *serviceImpl) DismissShipmentException(ctx context.Context, orgID int64, shipmentID int64, exceptionID int64) error {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return err
	}
	if sh == nil {
		return fmt.Errorf("shipment %d not found or access denied", shipmentID)
	}

	ex, err := s.repo.GetExceptionByID(ctx, orgID, exceptionID)
	if err != nil {
		return err
	}
	if ex == nil || ex.ShipmentID != shipmentID {
		return fmt.Errorf("exception %d not found or mismatch on shipment %d", exceptionID, shipmentID)
	}

	if ex.Status == spec.ExceptionStatusResolved || ex.Status == spec.ExceptionStatusDismissed {
		return fmt.Errorf("cannot dismiss an exception that is already %s", ex.Status)
	}

	now := time.Now()
	ex.Status = spec.ExceptionStatusDismissed
	ex.Resolved = true
	ex.ResolvedAt = &now

	err = s.repo.UpdateException(ctx, ex)
	if err != nil {
		return err
	}

	_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_EXCEPTION_DISMISSED,
		fmt.Sprintf("Exception dismissed: %s", ex.Title),
		"Logistics Operations",
	)

	// Restore shipment status to the latest completed milestone if there are no other active exceptions
	milestones, err := s.repo.GetMilestones(ctx, shipmentID)
	if err == nil {
		status := spec.BOOKING_PENDING
		for _, m := range milestones {
			if m.Status == "COMPLETED" {
				status = m.MilestoneCode
			}
		}
		var activeCount int
		err = s.db.GetContext(ctx, &activeCount, `
			SELECT COUNT(*) FROM shipment_exceptions 
			WHERE shipment_id = ? AND status NOT IN ('RESOLVED', 'DISMISSED')
		`, shipmentID)
		if err == nil && activeCount == 0 {
			sh.Status = status
			_ = s.repo.UpdateShipment(ctx, sh)

			_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_STATUS_UPDATED,
				fmt.Sprintf("All operational exceptions resolved/dismissed. Restored shipment status to %s", status),
				"Logistics Operations",
			)
		}
	}

	return nil
}

func (s *serviceImpl) EvaluateShipmentExceptions(ctx context.Context, orgID int64, shipmentID int64) error {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return err
	}
	if sh == nil {
		return fmt.Errorf("shipment %d not found or access denied", shipmentID)
	}

	milestones, err := s.repo.GetMilestones(ctx, shipmentID)
	if err != nil {
		return err
	}

	exceptions := EvaluateDeterministicExceptions(sh, milestones)
	for _, ex := range exceptions {
		_ = s.CreateShipmentException(ctx, orgID, shipmentID, ex.ExceptionType, ex.Severity, ex.Title, *ex.Description, ex.SourceEventID)
	}

	return nil
}

func (s *serviceImpl) HandleInboundCarrierEvent(ctx context.Context, orgID int64, event *spec.NormalizedTrackingEvent) error {
	// 1. Persist the raw carrier event first (ON CONFLICT DO NOTHING) to prevent data loss
	rawPayloadJSON, _ := event.RawPayload.MarshalJSON()
	rawEv := &spec.CarrierTrackingEvent{
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
	var matchedShipments []*spec.Shipment

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

func (s *serviceImpl) CompleteCarrierEvent(ctx context.Context, eventID string, orgID int64, shipmentID int64, hasCritical bool, aiSummary string) error {
	// Mark processing_status = PROCESSED in carrier_tracking_events and update associated shipment
	return s.repo.UpdateCarrierEventStatus(ctx, eventID, orgID, "MATCHED", "PROCESSED", &shipmentID)
}

func (s *serviceImpl) GetShipmentTracking(ctx context.Context, orgID int64, shipmentID int64) (*spec.ShipmentTrackingSummary, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	if sh == nil {
		return nil, fmt.Errorf("shipment %d not found", shipmentID)
	}

	milestones, err := s.repo.GetMilestones(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	exceptions, err := s.repo.GetExceptions(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}

	summary := CalculateTrackingSummary(sh, milestones, exceptions)
	return &summary, nil
}

func (s *serviceImpl) EvaluateClosure(ctx context.Context, orgID int64, shipmentID int64) (string, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return "", err
	}
	if sh == nil {
		return "", fmt.Errorf("shipment %d not found", shipmentID)
	}

	milestones, err := s.repo.GetMilestones(ctx, shipmentID)
	if err != nil {
		return "", err
	}

	exceptions, err := s.repo.GetExceptions(ctx, orgID, shipmentID)
	if err != nil {
		return "", err
	}

	docs, _ := s.repo.GetShipmentDocuments(ctx, orgID, shipmentID)
	docSummary := EvaluateDocumentCompliance(sh, docs)

	evalStatus := EvaluateShipmentClosure(sh, milestones, exceptions, docSummary)
	if evalStatus != sh.ClosureStatus {
		err = s.repo.UpdateClosureStatus(ctx, orgID, shipmentID, evalStatus)
		if err != nil {
			return "", err
		}
		
		// Record timeline update if state shifted
		_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_TRACKING_STATE_UPDATED,
			fmt.Sprintf("Shipment closure eligibility evaluated as %s", evalStatus),
			"System Engine",
		)
	}

	return evalStatus, nil
}

func (s *serviceImpl) RequestClosure(ctx context.Context, orgID int64, shipmentID int64) error {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return err
	}
	if sh == nil {
		return fmt.Errorf("shipment %d not found", shipmentID)
	}

	if sh.ClosureStatus == spec.ClosureStatusClosed {
		return fmt.Errorf("cannot request closure: shipment is already closed")
	}

	milestones, err := s.repo.GetMilestones(ctx, shipmentID)
	if err != nil {
		return err
	}

	exceptions, err := s.repo.GetExceptions(ctx, orgID, shipmentID)
	if err != nil {
		return err
	}

	docs, _ := s.repo.GetShipmentDocuments(ctx, orgID, shipmentID)
	docSummary := EvaluateDocumentCompliance(sh, docs)

	evalStatus := EvaluateShipmentClosure(sh, milestones, exceptions, docSummary)
	if evalStatus != spec.ClosureStatusReady {
		return fmt.Errorf("cannot request closure: shipment is not ready (status: %s)", evalStatus)
	}

	err = s.repo.UpdateClosureStatus(ctx, orgID, shipmentID, spec.ClosureStatusReady)
	if err != nil {
		return err
	}

	// Record timeline update
	_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_CLOSURE_REQUESTED,
		"Operator requested operational closure for shipment",
		"Logistics Operations",
	)

	return nil
}

func (s *serviceImpl) CompleteShipment(ctx context.Context, orgID int64, shipmentID int64) error {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return err
	}
	if sh == nil {
		return fmt.Errorf("shipment %d not found", shipmentID)
	}

	if sh.ClosureStatus == spec.ClosureStatusClosed {
		// Idempotency check: already closed
		return nil
	}

	milestones, err := s.repo.GetMilestones(ctx, shipmentID)
	if err != nil {
		return err
	}

	exceptions, err := s.repo.GetExceptions(ctx, orgID, shipmentID)
	if err != nil {
		return err
	}

	docs, _ := s.repo.GetShipmentDocuments(ctx, orgID, shipmentID)
	docSummary := EvaluateDocumentCompliance(sh, docs)

	evalStatus := EvaluateShipmentClosure(sh, milestones, exceptions, docSummary)
	// Must be ready for closure
	if evalStatus != spec.ClosureStatusReady && evalStatus != spec.ClosureStatusClosed {
		return fmt.Errorf("cannot complete shipment: required milestones incomplete or blocked by document/exceptions (status: %s)", evalStatus)
	}

	err = s.repo.UpdateClosureStatus(ctx, orgID, shipmentID, spec.ClosureStatusClosed)
	if err != nil {
		return err
	}

	// Make sure shipment status is DELIVERED
	if sh.Status != spec.DELIVERED {
		sh.Status = spec.DELIVERED
		_ = s.repo.UpdateShipment(ctx, sh)
	}

	// Record timeline update
	_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_COMPLETED,
		"Operator successfully completed and closed shipment",
		"Logistics Operations",
	)

	return nil
}

func (s *serviceImpl) ReopenShipment(ctx context.Context, orgID int64, shipmentID int64) error {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return err
	}
	if sh == nil {
		return fmt.Errorf("shipment %d not found", shipmentID)
	}

	if sh.ClosureStatus != spec.ClosureStatusClosed {
		return fmt.Errorf("cannot reopen shipment: shipment is not closed (status: %s)", sh.ClosureStatus)
	}

	// Reset to ACTIVE (which will be re-evaluated as needed)
	err = s.repo.UpdateClosureStatus(ctx, orgID, shipmentID, spec.ClosureStatusActive)
	if err != nil {
		return err
	}

	// Record timeline update
	_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_REOPENED,
		"Operator reopened closed shipment for modifications",
		"Logistics Operations",
	)

	return nil
}

// ─── Shipment Document Service Methods (Task 16.9) ──────────────────────────

// GetShipmentDocuments retrieves all operational documents belonging to a shipment as well as upstream linked documents (from RFQs/Bookings).
// In simple terms, this function:
// 1. Checks that the shipment exists and is owned by the organization.
// 2. Loads all direct shipment documents (e.g. Master B/L, Delivery Order).
// 3. Loads any upstream documents from the associated RFQ/Booking without duplicating files.
// 4. Fetches cross-document discrepancy alerts generated by OCR checks.
// 5. Evaluates document compliance, expiry risks (valid, expiring soon, expired), and missing required documents.
// 6. Returns the full document register, compliance summary, and any discrepancy notices.
func (s *serviceImpl) GetShipmentDocuments(ctx context.Context, orgID int64, shipmentID int64) ([]*spec.ShipmentDocument, *spec.ShipmentDocumentComplianceSummary, []*spec.ShipmentDocumentDiscrepancy, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, nil, nil, err
	}
	if sh == nil {
		return nil, nil, nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	// 1. Fetch direct shipment documents
	shipDocs, err := s.repo.GetShipmentDocuments(ctx, orgID, shipmentID)
	if err != nil {
		return nil, nil, nil, err
	}

	// 2. Fetch upstream RFQ documents if shipment is linked to an RFQ
	allDocs := make([]*spec.ShipmentDocument, 0, len(shipDocs))
	allDocs = append(allDocs, shipDocs...)

	if sh.RFQID != nil && *sh.RFQID > 0 {
		upstreamDocs, err := s.repo.GetUpstreamRFQDocuments(ctx, orgID, *sh.RFQID)
		if err == nil && len(upstreamDocs) > 0 {
			// Avoid duplicates if direct shipment document has same type
			existingTypes := make(map[string]bool)
			for _, d := range shipDocs {
				existingTypes[NormalizeDocType(d.DocType)] = true
			}
			for _, ud := range upstreamDocs {
				if !existingTypes[NormalizeDocType(ud.DocType)] {
					allDocs = append(allDocs, ud)
				}
			}
		}
	}

	// 3. Fetch automated OCR cross-document discrepancies
	discrepancies, err := s.repo.GetShipmentDiscrepancies(ctx, orgID, shipmentID)
	if err != nil {
		discrepancies = make([]*spec.ShipmentDocumentDiscrepancy, 0)
	}

	// 4. Compute authoritative compliance summary and missing documents queue
	compliance := EvaluateDocumentCompliance(sh, allDocs)

	return allDocs, compliance, discrepancies, nil
}

// CreateShipmentDocument uploads or links a new document file to a shipment.
// In simple terms, this function:
// 1. Validates the shipment and standardizes the document type and category.
// 2. Stores file metadata (filename, S3 storage key, reference number, document date, expiry date).
// 3. Inserts the document record with status UPLOADED.
// 4. Queues a background OCR verification task for automated data extraction.
// 5. Records a "Document Uploaded" event in the shipment's activity timeline.
func (s *serviceImpl) CreateShipmentDocument(ctx context.Context, orgID int64, shipmentID int64, req spec.CreateShipmentDocumentRequest, uploader string, userID *int64) (*spec.ShipmentDocument, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	if sh == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	normDocType := NormalizeDocType(req.DocType)
	cat := req.Category
	if cat == "" {
		cat = InferCategory(normDocType)
	}

	docName := req.DocumentName
	if docName == "" {
		docName = req.FileName
	}

	now := time.Now()
	s3Key := fmt.Sprintf("uploads/shipments/%d/%s_%d_%s", shipmentID, normDocType, now.Unix(), req.FileName)
	fileType := "PDF"
	if req.MimeType != "" {
		parts := strings.Split(req.MimeType, "/")
		if len(parts) > 1 {
			fileType = strings.ToUpper(parts[1])
		}
	}

	var refNum *string
	if req.ReferenceNumber != "" {
		r := strings.TrimSpace(req.ReferenceNumber)
		refNum = &r
	}

	doc := &spec.ShipmentDocument{
		OrgID:           orgID,
		ShipmentID:      shipmentID,
		DocType:         normDocType,
		DocumentName:    &docName,
		Category:        cat,
		Description:     &req.Description,
		S3Key:           &s3Key,
		FileName:        req.FileName,
		FileURL:         &req.FileURL,
		FileSize:        &req.FileSize,
		MimeType:        &req.MimeType,
		FileType:        &fileType,
		Status:          spec.DocStatusUploaded,
		UploadedBy:      &uploader,
		UploadedAt:      &now,
		ExpiresAt:       req.ExpiresAt,
		DocumentDate:    req.DocumentDate,
		ReferenceNumber: refNum,
		Source:          spec.DocSourceShipment,
	}
	if req.Source != "" {
		doc.Source = req.Source
	}
	if req.SourceID != nil {
		doc.SourceID = req.SourceID
	}

	err = s.repo.CreateShipmentDocument(ctx, orgID, doc)
	if err != nil {
		return nil, err
	}

	// Queue AI Compliance OCR verification task
	payload := map[string]interface{}{
		"org_id":       orgID,
		"shipment_id":  shipmentID,
		"doc_id":       doc.ID,
		"doc_type":     normDocType,
		"s3_key":       s3Key,
		"file_name":    req.FileName,
		"callback_url": s.backendBaseURL + "/internal/compliance/callback",
	}
	payloadJSON, _ := json.Marshal(payload)
	queryTask := `
		INSERT INTO ai_processing_tasks (org_id, entity_type, entity_id, task_type, payload, status, created_at, updated_at)
		VALUES (?, 'SHIPMENT_DOCUMENT', ?, 'DOC_VERIFY', ?, 'QUEUED', NOW(), NOW())
	`
	_, _ = s.db.ExecContext(ctx, queryTask, orgID, fmt.Sprintf("%d", doc.ID), string(payloadJSON))

	// Log activity
	_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_DOCUMENT_UPLOADED,
		fmt.Sprintf("%s (%s) uploaded by %s", docName, normDocType, uploader),
		uploader,
	)

	return doc, nil
}

// UpdateShipmentDocument updates metadata or status for an existing shipment document.
// In simple terms, this function:
// 1. Confirms the document exists and belongs to the specified shipment and organization.
// 2. Validates any requested lifecycle status changes against allowed transitions.
// 3. Updates fields such as document name, category, reference number, document date, expiry date, or rejection reason.
// 4. Records a "Document Updated" event in the shipment's activity timeline.
func (s *serviceImpl) UpdateShipmentDocument(ctx context.Context, orgID int64, shipmentID int64, docID int64, req spec.UpdateShipmentDocumentRequest, reviewer string, userID *int64) (*spec.ShipmentDocument, error) {
	doc, err := s.repo.GetShipmentDocumentByID(ctx, orgID, shipmentID, docID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	if req.DocumentName != nil {
		doc.DocumentName = req.DocumentName
	}
	if req.Category != nil {
		doc.Category = *req.Category
	}
	if req.Description != nil {
		doc.Description = req.Description
	}
	if req.ExpiresAt != nil {
		doc.ExpiresAt = req.ExpiresAt
	}
	if req.DocumentDate != nil {
		doc.DocumentDate = req.DocumentDate
	}
	if req.ReferenceNumber != nil {
		r := strings.TrimSpace(*req.ReferenceNumber)
		doc.ReferenceNumber = &r
	}

	if req.Status != nil && *req.Status != "" && *req.Status != doc.Status {
		if err := ValidateDocumentTransition(doc.Status, *req.Status); err != nil {
			return nil, err
		}
		doc.Status = *req.Status
		now := time.Now()
		doc.ReviewedBy = &reviewer
		doc.ReviewedAt = &now
		if req.RejectionReason != nil {
			doc.RejectionReason = req.RejectionReason
		}
	}

	err = s.repo.UpdateShipmentDocument(ctx, orgID, doc)
	if err != nil {
		return nil, err
	}

	_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_DOCUMENT_UPDATED,
		fmt.Sprintf("Document %s updated by %s", doc.FileName, reviewer),
		reviewer,
	)

	return doc, nil
}

// ApproveShipmentDocument marks a document as verified and approved for operational and customs compliance.
// In simple terms, this function:
// 1. Checks that the document is in a valid state to be approved (e.g. UPLOADED or UNDER_REVIEW).
// 2. Sets the status to APPROVED, records the reviewer name and approval timestamp, and clears any previous rejection reason.
// 3. Saves the changes to the database.
// 4. Logs a "Document Approved" entry in the shipment's activity timeline.
func (s *serviceImpl) ApproveShipmentDocument(ctx context.Context, orgID int64, shipmentID int64, docID int64, reviewer string, userID *int64) (*spec.ShipmentDocument, error) {
	doc, err := s.repo.GetShipmentDocumentByID(ctx, orgID, shipmentID, docID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	if err := ValidateDocumentTransition(doc.Status, spec.DocStatusApproved); err != nil {
		return nil, err
	}

	now := time.Now()
	doc.Status = spec.DocStatusApproved
	doc.ReviewedBy = &reviewer
	doc.ReviewedAt = &now
	doc.RejectionReason = nil

	err = s.repo.UpdateShipmentDocument(ctx, orgID, doc)
	if err != nil {
		return nil, err
	}

	_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_DOCUMENT_APPROVED,
		fmt.Sprintf("Document %s approved by %s", doc.FileName, reviewer),
		reviewer,
	)

	return doc, nil
}

// RejectShipmentDocument rejects a document file with an explicit operational reason.
// In simple terms, this function:
// 1. Ensures a non-empty rejection reason is provided (e.g. "Consignee address mismatch with customs entry").
// 2. Sets the status to REJECTED and records the reason, reviewer, and timestamp.
// 3. Updates the database record.
// 4. Logs a "Document Rejected" alert in the shipment activity timeline.
func (s *serviceImpl) RejectShipmentDocument(ctx context.Context, orgID int64, shipmentID int64, docID int64, reason string, reviewer string, userID *int64) (*spec.ShipmentDocument, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("rejection reason is required"))
	}

	doc, err := s.repo.GetShipmentDocumentByID(ctx, orgID, shipmentID, docID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	if err := ValidateDocumentTransition(doc.Status, spec.DocStatusRejected); err != nil {
		return nil, err
	}

	now := time.Now()
	doc.Status = spec.DocStatusRejected
	doc.ReviewedBy = &reviewer
	doc.ReviewedAt = &now
	doc.RejectionReason = &reason

	err = s.repo.UpdateShipmentDocument(ctx, orgID, doc)
	if err != nil {
		return nil, err
	}

	_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_DOCUMENT_REJECTED,
		fmt.Sprintf("Document %s rejected by %s: %s", doc.FileName, reviewer, reason),
		reviewer,
	)

	return doc, nil
}

// DeleteShipmentDocument permanently removes a document record from a shipment.
// In simple terms, this function:
// 1. Verifies that the document exists and belongs to the organization and shipment.
// 2. Deletes the database record.
// 3. Logs a "Document Deleted" event in the shipment's activity timeline.
func (s *serviceImpl) DeleteShipmentDocument(ctx context.Context, orgID int64, shipmentID int64, docID int64, userID *int64) error {
	doc, err := s.repo.GetShipmentDocumentByID(ctx, orgID, shipmentID, docID)
	if err != nil {
		return err
	}
	if doc == nil {
		return svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	err = s.repo.DeleteShipmentDocument(ctx, orgID, shipmentID, docID)
	if err != nil {
		return err
	}

	_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_DOCUMENT_DELETED,
		fmt.Sprintf("Document %s deleted", doc.FileName),
		"Operator",
	)

	return nil
}

// ─── Financial Operations Service Methods (Task 16.8) ───────────────────────

// GetShipmentFinancials aggregates all financial data for a shipment.
// In simple terms, this function:
// 1. Checks that the shipment exists and belongs to the requesting organization.
// 2. Pulls the original accepted commercial quote (estimated sell/buy prices).
// 3. Pulls all direct operational line items (costs like drayage, customs, detention).
// 4. Pulls audited carrier invoices (actual costs) and customer invoices (actual revenue).
// 5. Passes everything to the calculation engine to compute real revenues, costs, margins, and variances.
// 6. Caches the latest profitability numbers in the database and returns the summary.
func (s *serviceImpl) GetShipmentFinancials(ctx context.Context, orgID int64, shipmentID int64) (*spec.ShipmentFinancialSummary, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	if sh == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	// 1. Fetch quote commercial terms if linked
	var quote *RFQQuoteCommercialSnapshot
	if sh.QuoteID != nil && *sh.QuoteID > 0 {
		q, err := s.repo.GetQuoteCommercialSnapshot(ctx, *sh.QuoteID)
		if err == nil {
			quote = q
		}
	}

	// 2. Fetch direct shipment charges
	charges, err := s.repo.GetShipmentCharges(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}

	// 3. Fetch carrier invoices & customer invoices
	carrierInvoices, _ := s.repo.GetCarrierInvoicesSnapshot(ctx, orgID, shipmentID)
	customerInvoices, _ := s.repo.GetCustomerInvoicesSnapshot(ctx, orgID, shipmentID)

	// 4. Calculate summary
	summary := CalculateFinancialSummary(sh, quote, charges, carrierInvoices, customerInvoices)

	// 5. Persist to shipment_finance_profitability for cross-system cache
	_ = s.repo.SaveShipmentProfitabilitySnapshot(ctx, orgID, shipmentID, summary.ActualRevenue, summary.ActualCost, summary.ActualMargin, summary.ActualMarginPercent)

	return summary, nil
}

// GetShipmentCharges retrieves all individual cost and revenue line items for a shipment.
// For example: ocean freight base, origin drayage, terminal handling, customs brokerage, etc.
func (s *serviceImpl) GetShipmentCharges(ctx context.Context, orgID int64, shipmentID int64) ([]*spec.ShipmentFinancialCharge, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	if sh == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	return s.repo.GetShipmentCharges(ctx, orgID, shipmentID)
}

// CreateShipmentCharge adds a new operational charge line item to the shipment.
// In simple terms, this function:
// 1. Validates the shipment and standardizes the inputs (category, currency, estimated/actual amounts).
// 2. Inserts the new charge record into the database.
// 3. Immediately triggers a recalculation so the shipment's overall profit margin updates in real-time.
// 4. Logs an entry to the shipment's activity timeline (e.g. "Added COST line item: USD 350.00").
func (s *serviceImpl) CreateShipmentCharge(ctx context.Context, orgID int64, shipmentID int64, req *spec.CreateShipmentChargeRequest, actor string) (*spec.ShipmentFinancialCharge, *spec.ShipmentFinancialSummary, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, nil, err
	}
	if sh == nil {
		return nil, nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	category := strings.ToUpper(strings.TrimSpace(req.Category))
	if category == "" {
		category = spec.CostCategoryOther
	}
	chargeType := strings.ToUpper(strings.TrimSpace(req.ChargeType))
	if chargeType == "" {
		chargeType = spec.ChargeTypeCost
	}
	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status == "" {
		status = spec.ChargeStatusEstimated
	}
	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "USD"
	}

	var vendorName *string
	if req.VendorName != "" {
		v := strings.TrimSpace(req.VendorName)
		vendorName = &v
	}
	var refNum *string
	if req.ReferenceNumber != "" {
		r := strings.TrimSpace(req.ReferenceNumber)
		refNum = &r
	}
	var notes *string
	if req.Notes != "" {
		n := strings.TrimSpace(req.Notes)
		notes = &n
	}

	charge := &spec.ShipmentFinancialCharge{
		OrgID:           orgID,
		ShipmentID:      shipmentID,
		BookingID:       sh.BookingID,
		RFQID:           sh.RFQID,
		Category:        category,
		ChargeType:      chargeType,
		Description:     strings.TrimSpace(req.Description),
		VendorName:      vendorName,
		EstimatedAmount: req.EstimatedAmount,
		ActualAmount:    req.ActualAmount,
		Currency:        currency,
		ReferenceNumber: refNum,
		ChargeDate:      req.ChargeDate,
		Status:          status,
		Notes:           notes,
	}

	err = s.repo.CreateShipmentCharge(ctx, orgID, charge)
	if err != nil {
		return nil, nil, err
	}

	// Recalculate summary
	summary, err := s.GetShipmentFinancials(ctx, orgID, shipmentID)
	if err != nil {
		return charge, nil, nil
	}

	if actor == "" {
		actor = "Operations Team"
	}

	_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_COST_ADDED,
		fmt.Sprintf("Added %s line item (%s): %s %.2f", chargeType, charge.Description, currency, charge.EstimatedAmount),
		actor,
	)

	return charge, summary, nil
}

// UpdateShipmentCharge updates an existing financial line item.
// In simple terms:
// 1. Verifies that the charge item exists and belongs to the shipment and organization.
// 2. Modifies any fields provided (such as updating an estimated cost to the actual billed amount).
// 3. Saves the changes to the database.
// 4. Recalculates the shipment's profit summary and logs the update to the activity timeline.
func (s *serviceImpl) UpdateShipmentCharge(ctx context.Context, orgID int64, shipmentID int64, chargeID int64, req *spec.UpdateShipmentChargeRequest, actor string) (*spec.ShipmentFinancialCharge, *spec.ShipmentFinancialSummary, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, nil, err
	}
	if sh == nil {
		return nil, nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	charge, err := s.repo.GetShipmentChargeByID(ctx, orgID, shipmentID, chargeID)
	if err != nil {
		return nil, nil, err
	}
	if charge == nil {
		return nil, nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	if req.Category != nil {
		charge.Category = strings.ToUpper(strings.TrimSpace(*req.Category))
	}
	if req.ChargeType != nil {
		charge.ChargeType = strings.ToUpper(strings.TrimSpace(*req.ChargeType))
	}
	if req.Description != nil {
		charge.Description = strings.TrimSpace(*req.Description)
	}
	if req.VendorName != nil {
		v := strings.TrimSpace(*req.VendorName)
		charge.VendorName = &v
	}
	if req.EstimatedAmount != nil {
		charge.EstimatedAmount = *req.EstimatedAmount
	}
	if req.ActualAmount != nil {
		charge.ActualAmount = *req.ActualAmount
	}
	if req.Currency != nil {
		charge.Currency = strings.ToUpper(strings.TrimSpace(*req.Currency))
	}
	if req.ReferenceNumber != nil {
		r := strings.TrimSpace(*req.ReferenceNumber)
		charge.ReferenceNumber = &r
	}
	if req.ChargeDate != nil {
		charge.ChargeDate = req.ChargeDate
	}
	if req.Status != nil {
		charge.Status = strings.ToUpper(strings.TrimSpace(*req.Status))
	}
	if req.Notes != nil {
		n := strings.TrimSpace(*req.Notes)
		charge.Notes = &n
	}

	err = s.repo.UpdateShipmentCharge(ctx, orgID, charge)
	if err != nil {
		return nil, nil, err
	}

	summary, err := s.GetShipmentFinancials(ctx, orgID, shipmentID)
	if err != nil {
		return charge, nil, nil
	}

	if actor == "" {
		actor = "Operations Team"
	}

	_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_COST_UPDATED,
		fmt.Sprintf("Updated %s line item (%s): %s %.2f actual", charge.ChargeType, charge.Description, charge.Currency, charge.ActualAmount),
		actor,
	)

	return charge, summary, nil
}

// DeleteShipmentCharge deletes a financial charge from the shipment.
// In simple terms:
// 1. Verifies ownership and existence of the charge record.
// 2. Deletes the charge from the database.
// 3. Recalculates the shipment's profit summary without this charge.
// 4. Logs the removal action in the shipment's activity timeline.
func (s *serviceImpl) DeleteShipmentCharge(ctx context.Context, orgID int64, shipmentID int64, chargeID int64, actor string) (*spec.ShipmentFinancialSummary, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	if sh == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	charge, err := s.repo.GetShipmentChargeByID(ctx, orgID, shipmentID, chargeID)
	if err != nil {
		return nil, err
	}
	if charge == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	err = s.repo.DeleteShipmentCharge(ctx, orgID, shipmentID, chargeID)
	if err != nil {
		return nil, err
	}

	summary, err := s.GetShipmentFinancials(ctx, orgID, shipmentID)

	if actor == "" {
		actor = "Operations Team"
	}

	_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_COST_REMOVED,
		fmt.Sprintf("Removed %s charge item: %s", charge.ChargeType, charge.Description),
		actor,
	)

	return summary, err
}

// RecalculateShipmentFinancials forces a live refresh and recalculation of all shipment financial numbers.
// In simple terms:
// 1. Re-evaluates all revenue sources and cost items against current database records.
// 2. Computes the updated margin and variance amounts.
// 3. Logs an event showing the new margin percentage on the shipment's timeline.
func (s *serviceImpl) RecalculateShipmentFinancials(ctx context.Context, orgID int64, shipmentID int64, actor string) (*spec.ShipmentFinancialSummary, error) {
	summary, err := s.GetShipmentFinancials(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}

	if actor == "" {
		actor = "Operations System"
	}

	_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_FINANCIAL_RECALCULATED,
		fmt.Sprintf("Shipment margin recalculated: %.1f%% margin (%s %.2f)", summary.ActualMarginPercent, summary.Currency, summary.ActualMargin),
		actor,
	)

	return summary, nil
}

// ReviewShipmentFinancials records a management review or financial closure of the shipment.
// In simple terms:
// 1. Confirms the shipment exists and is owned by the organization.
// 2. Fetches the latest financial summary.
// 3. Sets the financial status (e.g. FINANCIALLY_CLOSED or REVIEWED).
// 4. Logs a timeline event recording who performed the review and any operational notes provided.
func (s *serviceImpl) ReviewShipmentFinancials(ctx context.Context, orgID int64, shipmentID int64, status string, notes string, actor string) (*spec.ShipmentFinancialSummary, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	if sh == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	summary, err := s.GetShipmentFinancials(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}

	targetStatus := strings.ToUpper(strings.TrimSpace(status))
	if targetStatus == "" {
		targetStatus = spec.FinancialStatusFinanciallyClosed
	}
	summary.FinancialStatus = targetStatus

	if actor == "" {
		actor = "Finance Manager"
	}

	_ = s.repo.CreateActivity(ctx, orgID, "SHIPMENT", shipmentID, spec.SHIPMENT_FINANCIAL_REVIEWED,
		fmt.Sprintf("Financial review completed: status set to %s. Notes: %s", targetStatus, notes),
		actor,
	)

	return summary, nil
}

func (s *serviceImpl) GetLatestTrackingPosition(ctx context.Context, orgID int64, shipmentID int64) (*spec.TrackingPosition, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	if sh == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	provider := ResolveTrackingProvider()
	pos, err := provider.GetLatestPosition(ctx, orgID, sh)
	if err != nil {
		return nil, err
	}

	if pos != nil {
		_ = s.repo.SaveTrackingPosition(ctx, pos)
	}

	return pos, nil
}

func (s *serviceImpl) GetTrackingPositionHistory(ctx context.Context, orgID int64, shipmentID int64, limit int) ([]spec.TrackingPosition, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	if sh == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	provider := ResolveTrackingProvider()
	return provider.GetPositionHistory(ctx, orgID, sh, limit)
}

func (s *serviceImpl) GetTrackingRoute(ctx context.Context, orgID int64, shipmentID int64) (*spec.TrackingRoute, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	if sh == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	provider := ResolveTrackingProvider()
	return provider.GetRoute(ctx, orgID, sh)
}

func (s *serviceImpl) GetTrackingEventsList(ctx context.Context, orgID int64, shipmentID int64) ([]spec.TrackingEventNormalized, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	if sh == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	provider := ResolveTrackingProvider()
	return provider.GetTrackingEvents(ctx, orgID, sh)
}

func (s *serviceImpl) GetShipmentTrackingIntelligence(ctx context.Context, orgID int64, shipmentID int64) (*spec.ShipmentTrackingIntelligence, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	if sh == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	milestones, _ := s.repo.GetMilestones(ctx, shipmentID)
	exceptions, _ := s.repo.GetExceptions(ctx, orgID, shipmentID)

	provider := ResolveTrackingProvider()
	providerMeta := provider.GetMetadata()

	// Data Fallback Hierarchy: Provider -> Persisted Database Telemetry -> Unavailable
	pos, _ := provider.GetLatestPosition(ctx, orgID, sh)
	if pos == nil {
		// Fallback to persisted database position
		pos, _ = s.repo.GetLatestPosition(ctx, orgID, shipmentID)
	}

	events, _ := provider.GetTrackingEvents(ctx, orgID, sh)
	if len(events) == 0 {
		// Fallback to database carrier events if provider returned none
		dbEvents, err := s.repo.GetCarrierEventsForShipment(ctx, orgID, shipmentID)
		if err == nil && len(dbEvents) > 0 {
			events = NormalizeCarrierEvents(dbEvents)
		}
	}

	summary := CalculateTrackingSummary(sh, milestones, exceptions)
	schedule := BuildTrackingSchedule(sh, milestones)
	journey := BuildTrackingJourney(sh, milestones)
	alerts := GenerateTrackingAlerts(sh, milestones, exceptions, pos, &schedule)

	var activeExceptionsCount int64
	for _, ex := range exceptions {
		if !ex.Resolved && ex.Status != spec.ExceptionStatusDismissed {
			activeExceptionsCount++
		}
	}

	freshness := spec.TrackingFreshnessUnavailable
	trackingSource := "UNAVAILABLE"
	if pos != nil {
		freshness = pos.DataFreshness
		trackingSource = pos.TrackingSource
	}

	now := time.Now()

	lineage := spec.TrackingLineage{
		RFQID:       sh.RFQID,
		QuoteID:     sh.QuoteID,
		BookingID:   sh.BookingID,
		CarrierSCAC: sh.CarrierSCAC,
	}
	if sh.BookingNumber != nil {
		lineage.BookingNumber = *sh.BookingNumber
	}
	if len(sh.ContainerNumbers) > 0 {
		lineage.ContainerNumber = sh.ContainerNumbers[0]
	}
	if sh.MBLNumber != nil {
		lineage.MBLNumber = *sh.MBLNumber
	}
	if sh.VesselName != nil {
		lineage.VesselName = *sh.VesselName
	}
	if sh.VoyageNumber != nil {
		lineage.VoyageNumber = *sh.VoyageNumber
	}

	// Reconcile and persist tracking alerts for lifecycle monitoring (Task 17.5)
	_, _ = SyncTrackingAlerts(ctx, s.repo, sh, alerts, "System Engine")

	return &spec.ShipmentTrackingIntelligence{
		ShipmentID:            sh.ID,
		ShipmentNumber:        fmt.Sprintf("SH-%d", sh.ID),
		ShipmentStatus:        sh.Status,
		TrackingState:         summary.TrackingState,
		ClosureStatus:         sh.ClosureStatus,
		ProgressPercentage:    summary.ProgressPercentage,
		DataFreshness:         freshness,
		TrackingSource:        trackingSource,
		IsLiveTracking:        providerMeta.IsLive,
		ProviderMetadata:      providerMeta,
		LastUpdatedAt:         &now,
		Journey:               journey,
		Schedule:              schedule,
		Alerts:                alerts,
		Events:                events,
		ActiveExceptionsCount: activeExceptionsCount,
		LatestPosition:        pos,
		Lineage:               lineage,
	}, nil
}

func (s *serviceImpl) GetTrackingAlerts(ctx context.Context, orgID int64, shipmentID int64, status string) ([]*spec.ShipmentTrackingAlertRecord, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	if sh == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	return s.repo.GetTrackingAlerts(ctx, orgID, shipmentID, status)
}

func (s *serviceImpl) GetTrackingMonitoringSummary(ctx context.Context, orgID int64, shipmentID int64) (*spec.TrackingMonitoringSummary, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	if sh == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	provider := ResolveTrackingProvider()
	providerMeta := provider.GetMetadata()
	pos, _ := provider.GetLatestPosition(ctx, orgID, sh)
	if pos == nil {
		pos, _ = s.repo.GetLatestPosition(ctx, orgID, shipmentID)
	}

	freshness := spec.TrackingFreshnessUnavailable
	var lastRecorded *time.Time
	if pos != nil {
		freshness = pos.DataFreshness
		lastRecorded = &pos.RecordedAt
	}

	summary, err := BuildTrackingMonitoringSummary(ctx, s.repo, orgID, shipmentID, freshness, lastRecorded)
	if err != nil {
		return nil, err
	}
	summary.TrackingProvider = providerMeta.ProviderName
	summary.ProviderMetadata = providerMeta
	summary.IsLiveTracking = providerMeta.IsLive

	// ── Task 17.7: Populate Automation and Scheduling Information ────────────
	latestRun, _ := s.repo.GetLatestTrackingRefresh(ctx, orgID, shipmentID)
	intervalMinutes := getRefreshIntervalForShipment(sh.Status)
	autoRefreshEnabled := strings.ToLower(os.Getenv("TRACKING_AUTO_REFRESH_ENABLED")) != "false"

	summary.AutomaticRefreshEnabled = autoRefreshEnabled && intervalMinutes > 0
	summary.RefreshIntervalMinutes = intervalMinutes
	summary.RefreshStatus = "READY"

	if latestRun != nil {
		summary.LastRefreshAt = &latestRun.StartedAt
		summary.LastRefreshStatus = latestRun.Status
		summary.LastRefreshError = latestRun.ErrorMessage
		summary.RefreshStatus = latestRun.Status
		if intervalMinutes > 0 && autoRefreshEnabled {
			next := latestRun.StartedAt.Add(time.Duration(intervalMinutes) * time.Minute)
			summary.NextRefreshAt = &next
		}
	} else if intervalMinutes > 0 && autoRefreshEnabled {
		next := sh.UpdatedAt.Add(time.Duration(intervalMinutes) * time.Minute)
		summary.NextRefreshAt = &next
	}

	return summary, nil
}

func getRefreshIntervalForShipment(status string) int {
	switch status {
	case spec.BOOKED:
		return 360 // 6 hours
	case spec.DEPARTED:
		return 120 // 2 hours
	case spec.IN_TRANSIT:
		return 60 // 1 hour
	case spec.ARRIVED:
		return 180 // 3 hours
	case spec.DELIVERED:
		return 0 // disabled
	default:
		return 60
	}
}

func (s *serviceImpl) AcknowledgeTrackingAlert(ctx context.Context, orgID int64, shipmentID int64, alertID int64, userID *int64, actor string) error {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return err
	}
	if sh == nil {
		return svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	alert, err := s.repo.GetTrackingAlertByID(ctx, orgID, shipmentID, alertID)
	if err != nil {
		return err
	}
	if alert == nil {
		return svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	if err := s.repo.AcknowledgeTrackingAlert(ctx, orgID, shipmentID, alertID, userID); err != nil {
		return err
	}

	_ = s.repo.CreateActivity(
		ctx,
		orgID,
		"SHIPMENT",
		shipmentID,
		spec.SHIPMENT_TRACKING_ALERT_ACKNOWLEDGED,
		fmt.Sprintf("Operator acknowledged tracking alert: %s", alert.Title),
		actor,
	)
	return nil
}

func (s *serviceImpl) ResolveTrackingAlert(ctx context.Context, orgID int64, shipmentID int64, alertID int64, userID *int64, notes string, actor string) error {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return err
	}
	if sh == nil {
		return svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	alert, err := s.repo.GetTrackingAlertByID(ctx, orgID, shipmentID, alertID)
	if err != nil {
		return err
	}
	if alert == nil {
		return svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	if err := s.repo.ResolveTrackingAlert(ctx, orgID, shipmentID, alertID, userID); err != nil {
		return err
	}

	desc := fmt.Sprintf("Operator resolved tracking alert: %s", alert.Title)
	if notes != "" {
		desc += fmt.Sprintf(" (Notes: %s)", notes)
	}

	_ = s.repo.CreateActivity(
		ctx,
		orgID,
		"SHIPMENT",
		shipmentID,
		spec.SHIPMENT_TRACKING_ALERT_RESOLVED,
		desc,
		actor,
	)
	return nil
}

func (s *serviceImpl) SuppressTrackingAlert(ctx context.Context, orgID int64, shipmentID int64, alertID int64, userID *int64, reason string, actor string) error {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return err
	}
	if sh == nil {
		return svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	alert, err := s.repo.GetTrackingAlertByID(ctx, orgID, shipmentID, alertID)
	if err != nil {
		return err
	}
	if alert == nil {
		return svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	if err := s.repo.SuppressTrackingAlert(ctx, orgID, shipmentID, alertID, userID); err != nil {
		return err
	}

	desc := fmt.Sprintf("Operator suppressed tracking alert: %s", alert.Title)
	if reason != "" {
		desc += fmt.Sprintf(" (Reason: %s)", reason)
	}

	_ = s.repo.CreateActivity(
		ctx,
		orgID,
		"SHIPMENT",
		shipmentID,
		spec.SHIPMENT_TRACKING_ALERT_SUPPRESSED,
		desc,
		actor,
	)
	return nil
}

func (s *serviceImpl) RefreshShipmentTracking(ctx context.Context, orgID int64, shipmentID int64, userID *int64, actor string) (*spec.TrackingRefreshResult, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	if sh == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	// Determine Trigger Type
	triggerType := spec.TrackingTriggerManual
	if strings.Contains(strings.ToLower(actor), "scheduler") || strings.Contains(strings.ToLower(actor), "background") || strings.Contains(strings.ToLower(actor), "poller") {
		triggerType = spec.TrackingTriggerScheduled
	}

	// 1. Authoritative: If Carrier Tracking Engine is wired and shipment has a carrier, execute real carrier integration flow
	if s.carrierTrackingEngine != nil && (sh.CarrierSCAC != "" || (sh.CarrierName != nil && *sh.CarrierName != "")) {
		startTime := time.Now()
		provName := "Carrier Integration Adapter"
		provType := spec.ProviderTypeCarrier

		runRecord := &spec.TrackingRefreshRunRecord{
			OrgID:        orgID,
			ShipmentID:   shipmentID,
			ProviderName: &provName,
			ProviderType: &provType,
			TriggerType:  triggerType,
			Status:       spec.TrackingRunStatusStarted,
			StartedAt:    startTime,
		}
		_ = s.repo.CreateTrackingRefreshRun(ctx, runRecord)

		res, err := s.carrierTrackingEngine.SyncShipmentTracking(ctx, orgID, shipmentID, userID, actor)
		if err == nil && res != nil {
			// Recalculate intelligence based on updated database state
			intel, _ := s.GetShipmentTrackingIntelligence(ctx, orgID, shipmentID)
			if intel != nil {
				res.Intelligence = intel
				res.DataFreshness = intel.DataFreshness
			}

			completedTime := time.Now()
			runRecord.CompletedAt = &completedTime
			runRecord.NewPositions = res.NewPositions
			runRecord.NewEvents = res.NewEvents
			runRecord.DataFreshness = &res.DataFreshness
			runRecord.UsedFallback = res.UsedFallback
			if res.Success {
				runRecord.Status = spec.TrackingRunStatusSuccess
			} else {
				runRecord.Status = spec.TrackingRunStatusFailed
				runRecord.ErrorMessage = &res.Message
			}
			_ = s.repo.UpdateTrackingRefreshRun(ctx, runRecord)

			return res, nil
		}
	}

	provider := ResolveTrackingProvider()
	providerMeta := provider.GetMetadata()

	// 1. Authoritative: Create persistent TrackingRefreshRunRecord (STARTED)
	startTime := time.Now()
	runRecord := &spec.TrackingRefreshRunRecord{
		OrgID:        orgID,
		ShipmentID:   shipmentID,
		ProviderName: &providerMeta.ProviderName,
		ProviderType: &providerMeta.ProviderType,
		TriggerType:  triggerType,
		Status:       spec.TrackingRunStatusStarted,
		StartedAt:    startTime,
	}
	_ = s.repo.CreateTrackingRefreshRun(ctx, runRecord)

	// 2. Execute Provider Refresh Call (or fallback)
	refreshRes, err := provider.Refresh(ctx, orgID, sh)
	if err != nil || refreshRes == nil {
		now := time.Now()
		refreshRes = &spec.TrackingRefreshResult{
			Success:       true,
			Provider:      providerMeta,
			DataFreshness: spec.TrackingFreshnessUnavailable,
			LastUpdatedAt: &now,
			UsedFallback:  true,
			Message:       "Provider refresh completed using latest persisted operational data.",
		}
	}

	// 3. Fetch and persist any updated position idempotently
	pos, err := provider.GetLatestPosition(ctx, orgID, sh)
	if err == nil && pos != nil {
		_ = s.repo.SaveTrackingPosition(ctx, pos)
	}

	// 4. Recalculate complete intelligence & synchronize alerts
	intel, err := s.GetShipmentTrackingIntelligence(ctx, orgID, shipmentID)
	if err == nil && intel != nil {
		refreshRes.Intelligence = intel
		refreshRes.DataFreshness = intel.DataFreshness
	}

	// 5. Complete and Update TrackingRefreshRunRecord
	completedTime := time.Now()
	runRecord.CompletedAt = &completedTime
	runRecord.NewPositions = refreshRes.NewPositions
	runRecord.NewEvents = refreshRes.NewEvents
	runRecord.DataFreshness = &refreshRes.DataFreshness
	runRecord.UsedFallback = refreshRes.UsedFallback

	if refreshRes.Success {
		if refreshRes.UsedFallback && providerMeta.ProviderType != spec.ProviderTypeDemo && providerMeta.ProviderType != spec.ProviderTypeDatabase {
			runRecord.Status = spec.TrackingRunStatusPartial
		} else {
			runRecord.Status = spec.TrackingRunStatusSuccess
		}
	} else {
		runRecord.Status = spec.TrackingRunStatusFailed
		runRecord.ErrorMessage = &refreshRes.Message
	}
	_ = s.repo.UpdateTrackingRefreshRun(ctx, runRecord)

	// 6. Log Auditable Activity
	activityDesc := fmt.Sprintf("Tracking refreshed successfully using %s (%s)", providerMeta.ProviderName, providerMeta.ProviderType)
	if refreshRes.UsedFallback {
		activityDesc = "Tracking refresh completed using latest persisted operational data"
	}

	_ = s.repo.CreateActivity(
		ctx,
		orgID,
		"SHIPMENT",
		shipmentID,
		spec.SHIPMENT_TRACKING_REFRESHED,
		activityDesc,
		actor,
	)

	return refreshRes, nil
}

func (s *serviceImpl) GetTrackingRefreshHistory(ctx context.Context, orgID int64, shipmentID int64, limit int) ([]*spec.TrackingRefreshRunRecord, error) {
	sh, err := s.repo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	if sh == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	return s.repo.GetTrackingRefreshHistory(ctx, orgID, shipmentID, limit)
}

// ─── Tracking Analytics & Operational Intelligence (Task 17.8) ────────────────

func (s *serviceImpl) GetTrackingAnalyticsOverview(ctx context.Context, orgID int64) (*spec.TrackingAnalyticsOverview, error) {
	overview, err := s.repo.GetTrackingAnalyticsOverview(ctx, orgID)
	if err != nil {
		return nil, err
	}

	carriers, _ := s.repo.GetCarrierTrackingPerformance(ctx, orgID)
	routes, _ := s.repo.GetRouteTrackingPerformance(ctx, orgID)

	// Compute deterministic operational insights based on real aggregated metrics
	overview.Insights = GenerateOperationalInsights(overview, carriers, routes)

	return overview, nil
}

func (s *serviceImpl) GetTrackingAnalyticsTrends(ctx context.Context, orgID int64, days int) ([]spec.TrackingTrendDataPoint, error) {
	return s.repo.GetTrackingAnalyticsTrends(ctx, orgID, days)
}

func (s *serviceImpl) GetCarrierTrackingPerformance(ctx context.Context, orgID int64) ([]spec.CarrierTrackingPerformance, error) {
	return s.repo.GetCarrierTrackingPerformance(ctx, orgID)
}

func (s *serviceImpl) GetRouteTrackingPerformance(ctx context.Context, orgID int64) ([]spec.RouteTrackingPerformance, error) {
	return s.repo.GetRouteTrackingPerformance(ctx, orgID)
}




