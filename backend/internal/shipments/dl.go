package shipments

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/shipments/spec"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	CreateShipment(ctx context.Context, s *spec.Shipment) error
	CreateShipmentTx(ctx context.Context, tx *sqlx.Tx, s *spec.Shipment) error
	GetShipmentByID(ctx context.Context, orgID int64, id int64) (*spec.Shipment, error)
	GetShipmentByRFQID(ctx context.Context, rfqID int64) (*spec.Shipment, error)
	GetShipmentByBooking(ctx context.Context, orgID int64, bookingNumber string) (*spec.Shipment, error)
	GetShipmentByContainer(ctx context.Context, orgID int64, containerNumber string) (*spec.Shipment, error)
	GetShipmentByMBL(ctx context.Context, orgID int64, mblNumber string) (*spec.Shipment, error)
	GetShipmentByHBL(ctx context.Context, orgID int64, hblNumber string) (*spec.Shipment, error)

	FindShipmentsByBooking(ctx context.Context, orgID int64, bookingNumber string) ([]*spec.Shipment, error)
	FindShipmentsByContainer(ctx context.Context, orgID int64, containerNumber string) ([]*spec.Shipment, error)
	FindShipmentsByMBL(ctx context.Context, orgID int64, mblNumber string) ([]*spec.Shipment, error)
	FindShipmentsByHBL(ctx context.Context, orgID int64, hblNumber string) ([]*spec.Shipment, error)

	ListShipments(ctx context.Context, orgID int64) ([]*spec.Shipment, error)
	UpdateShipment(ctx context.Context, s *spec.Shipment) error

	GetMilestones(ctx context.Context, shipmentID int64) ([]*spec.ShipmentMilestone, error)
	CreateMilestone(ctx context.Context, m *spec.ShipmentMilestone) error
	CreateMilestoneTx(ctx context.Context, tx *sqlx.Tx, m *spec.ShipmentMilestone) error
	UpdateMilestone(ctx context.Context, m *spec.ShipmentMilestone) error

	GetExceptions(ctx context.Context, orgID int64, shipmentID int64) ([]*spec.ShipmentException, error)
	CreateException(ctx context.Context, ex *spec.ShipmentException) error
	GetExceptionByID(ctx context.Context, orgID int64, id int64) (*spec.ShipmentException, error)
	UpdateException(ctx context.Context, ex *spec.ShipmentException) error

	IsEventProcessed(ctx context.Context, eventID string) (bool, error)
	MarkEventProcessed(ctx context.Context, eventID string, shipmentID int64) error

	FindSCACByCarrierName(ctx context.Context, carrierName string) (string, error)

	InsertCarrierEvent(ctx context.Context, event *spec.CarrierTrackingEvent) (bool, error)
	UpdateCarrierEventStatus(ctx context.Context, eventID string, orgID int64, matchingStatus string, processingStatus string, shipmentID *int64) error
	CreateActivity(ctx context.Context, orgID int64, entityType string, entityID int64, action string, description string, actor string) error
	GetShipmentsWorkspace(ctx context.Context, orgID int64, filter spec.ShipmentListFilter) ([]*spec.Shipment, spec.ShipmentKPIs, int, error)
	UpdateClosureStatus(ctx context.Context, orgID int64, shipmentID int64, status string) error

	// Document operations (Task 16.7)
	GetShipmentDocuments(ctx context.Context, orgID int64, shipmentID int64) ([]*spec.ShipmentDocument, error)
	GetShipmentDocumentByID(ctx context.Context, orgID int64, shipmentID int64, docID int64) (*spec.ShipmentDocument, error)
	CreateShipmentDocument(ctx context.Context, orgID int64, doc *spec.ShipmentDocument) error
	UpdateShipmentDocument(ctx context.Context, orgID int64, doc *spec.ShipmentDocument) error
	DeleteShipmentDocument(ctx context.Context, orgID int64, shipmentID int64, docID int64) error
	GetUpstreamRFQDocuments(ctx context.Context, orgID int64, rfqID int64) ([]*spec.ShipmentDocument, error)
	GetShipmentDiscrepancies(ctx context.Context, orgID int64, shipmentID int64) ([]*spec.ShipmentDocumentDiscrepancy, error)

	// Financial operations (Task 16.8)
	GetShipmentCharges(ctx context.Context, orgID int64, shipmentID int64) ([]*spec.ShipmentFinancialCharge, error)
	GetShipmentChargeByID(ctx context.Context, orgID int64, shipmentID int64, chargeID int64) (*spec.ShipmentFinancialCharge, error)
	CreateShipmentCharge(ctx context.Context, orgID int64, charge *spec.ShipmentFinancialCharge) error
	UpdateShipmentCharge(ctx context.Context, orgID int64, charge *spec.ShipmentFinancialCharge) error
	DeleteShipmentCharge(ctx context.Context, orgID int64, shipmentID int64, chargeID int64) error
	GetQuoteCommercialSnapshot(ctx context.Context, quoteID int64) (*RFQQuoteCommercialSnapshot, error)
	GetCarrierInvoicesSnapshot(ctx context.Context, orgID int64, shipmentID int64) ([]*CarrierInvoiceSnapshot, error)
	GetCustomerInvoicesSnapshot(ctx context.Context, orgID int64, shipmentID int64) ([]*CustomerInvoiceSnapshot, error)
	SaveShipmentProfitabilitySnapshot(ctx context.Context, orgID int64, shipmentID int64, sellAmount float64, buyAmount float64, netProfit float64, marginPct float64) error

	// Real-time tracking position operations (Task 17.3)
	SaveTrackingPosition(ctx context.Context, pos *spec.TrackingPosition) error
	GetLatestPosition(ctx context.Context, orgID int64, shipmentID int64) (*spec.TrackingPosition, error)
	GetPositionHistory(ctx context.Context, orgID int64, shipmentID int64, limit int) ([]*spec.TrackingPosition, error)
	GetCarrierEventsForShipment(ctx context.Context, orgID int64, shipmentID int64) ([]*spec.CarrierTrackingEvent, error)

	// Persistent tracking alerts and monitoring operations (Task 17.5)
	GetTrackingAlerts(ctx context.Context, orgID int64, shipmentID int64, status string) ([]*spec.ShipmentTrackingAlertRecord, error)
	GetTrackingAlertByID(ctx context.Context, orgID int64, shipmentID int64, alertID int64) (*spec.ShipmentTrackingAlertRecord, error)
	GetTrackingAlertByKey(ctx context.Context, orgID int64, shipmentID int64, alertKey string) (*spec.ShipmentTrackingAlertRecord, error)
	CreateTrackingAlert(ctx context.Context, alert *spec.ShipmentTrackingAlertRecord) error
	UpdateTrackingAlert(ctx context.Context, alert *spec.ShipmentTrackingAlertRecord) error
	AcknowledgeTrackingAlert(ctx context.Context, orgID int64, shipmentID int64, alertID int64, userID *int64) error
	ResolveTrackingAlert(ctx context.Context, orgID int64, shipmentID int64, alertID int64, userID *int64) error
	SuppressTrackingAlert(ctx context.Context, orgID int64, shipmentID int64, alertID int64, userID *int64) error
	GetTrackingMonitoringSummary(ctx context.Context, orgID int64, shipmentID int64) (*spec.TrackingMonitoringSummary, error)

	// Tracking Refresh Runs & Monitoring (Task 17.7)
	CreateTrackingRefreshRun(ctx context.Context, run *spec.TrackingRefreshRunRecord) error
	UpdateTrackingRefreshRun(ctx context.Context, run *spec.TrackingRefreshRunRecord) error
	GetLatestTrackingRefresh(ctx context.Context, orgID int64, shipmentID int64) (*spec.TrackingRefreshRunRecord, error)
	GetTrackingRefreshHistory(ctx context.Context, orgID int64, shipmentID int64, limit int) ([]*spec.TrackingRefreshRunRecord, error)
	GetActiveShipmentsForRefresh(ctx context.Context) ([]*spec.Shipment, error)

	// Tracking Analytics & Operational Intelligence (Task 17.8)
	GetTrackingAnalyticsOverview(ctx context.Context, orgID int64) (*spec.TrackingAnalyticsOverview, error)
	GetTrackingAnalyticsTrends(ctx context.Context, orgID int64, days int) ([]spec.TrackingTrendDataPoint, error)
	GetCarrierTrackingPerformance(ctx context.Context, orgID int64) ([]spec.CarrierTrackingPerformance, error)
	GetRouteTrackingPerformance(ctx context.Context, orgID int64) ([]spec.RouteTrackingPerformance, error)
}

type repository struct {
	db *sqlx.DB
}

func autoMigrateDocumentAndFinancialSchema(db *sqlx.DB) {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS shipment_documents (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			org_id BIGINT NOT NULL,
			shipment_id BIGINT NOT NULL,
			doc_type VARCHAR(50) NOT NULL,
			document_name VARCHAR(255) NULL,
			category VARCHAR(100) NOT NULL DEFAULT 'TRANSPORT',
			description TEXT NULL,
			s3_key VARCHAR(1000) NOT NULL DEFAULT '',
			file_name VARCHAR(500) NOT NULL,
			file_url VARCHAR(1000) NULL,
			file_size BIGINT NULL,
			mime_type VARCHAR(100) NULL,
			file_type VARCHAR(20) DEFAULT 'PDF',
			status VARCHAR(50) NOT NULL DEFAULT 'UPLOADED',
			uploaded_by VARCHAR(255) NULL,
			uploaded_at DATETIME NULL,
			reviewed_by VARCHAR(255) NULL,
			reviewed_at DATETIME NULL,
			rejection_reason TEXT NULL,
			expires_at DATETIME NULL,
			document_date DATETIME NULL,
			reference_number VARCHAR(255) NULL,
			source VARCHAR(50) NOT NULL DEFAULT 'SHIPMENT',
			source_id BIGINT NULL,
			extracted_data JSON NULL,
			raw_ocr_text TEXT NULL,
			ai_summary TEXT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_ship_docs_shipment (org_id, shipment_id),
			INDEX idx_ship_docs_status (status)
		)`,
		`ALTER TABLE shipment_documents ADD COLUMN IF NOT EXISTS document_name VARCHAR(255) NULL`,
		`ALTER TABLE shipment_documents ADD COLUMN IF NOT EXISTS category VARCHAR(100) NOT NULL DEFAULT 'TRANSPORT'`,
		`ALTER TABLE shipment_documents ADD COLUMN IF NOT EXISTS description TEXT NULL`,
		`ALTER TABLE shipment_documents ADD COLUMN IF NOT EXISTS file_url VARCHAR(1000) NULL`,
		`ALTER TABLE shipment_documents ADD COLUMN IF NOT EXISTS file_size BIGINT NULL`,
		`ALTER TABLE shipment_documents ADD COLUMN IF NOT EXISTS mime_type VARCHAR(100) NULL`,
		`ALTER TABLE shipment_documents ADD COLUMN IF NOT EXISTS uploaded_by VARCHAR(255) NULL`,
		`ALTER TABLE shipment_documents ADD COLUMN IF NOT EXISTS uploaded_at DATETIME NULL`,
		`ALTER TABLE shipment_documents ADD COLUMN IF NOT EXISTS reviewed_by VARCHAR(255) NULL`,
		`ALTER TABLE shipment_documents ADD COLUMN IF NOT EXISTS reviewed_at DATETIME NULL`,
		`ALTER TABLE shipment_documents ADD COLUMN IF NOT EXISTS rejection_reason TEXT NULL`,
		`ALTER TABLE shipment_documents ADD COLUMN IF NOT EXISTS expires_at DATETIME NULL`,
		`ALTER TABLE shipment_documents ADD COLUMN IF NOT EXISTS document_date DATETIME NULL`,
		`ALTER TABLE shipment_documents ADD COLUMN IF NOT EXISTS reference_number VARCHAR(255) NULL`,
		`ALTER TABLE shipment_documents ADD COLUMN IF NOT EXISTS source VARCHAR(50) NOT NULL DEFAULT 'SHIPMENT'`,
		`ALTER TABLE shipment_documents ADD COLUMN IF NOT EXISTS source_id BIGINT NULL`,
		`CREATE TABLE IF NOT EXISTS shipment_financial_charges (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			org_id BIGINT NOT NULL,
			shipment_id BIGINT NOT NULL,
			booking_id BIGINT NULL,
			rfq_id BIGINT NULL,
			category VARCHAR(100) NOT NULL DEFAULT 'OCEAN_FREIGHT',
			charge_type VARCHAR(50) NOT NULL DEFAULT 'COST',
			description VARCHAR(255) NOT NULL,
			vendor_name VARCHAR(255) NULL,
			estimated_amount DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
			actual_amount DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
			currency VARCHAR(10) NOT NULL DEFAULT 'USD',
			reference_number VARCHAR(255) NULL,
			charge_date DATETIME NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'ESTIMATED',
			notes TEXT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_fin_charges_shipment (org_id, shipment_id),
			INDEX idx_fin_charges_type (charge_type)
		)`,
		`CREATE TABLE IF NOT EXISTS shipment_financial_summaries (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			org_id BIGINT NOT NULL,
			shipment_id BIGINT NOT NULL,
			currency VARCHAR(10) NOT NULL DEFAULT 'USD',
			estimated_revenue DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
			actual_revenue DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
			estimated_cost DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
			actual_cost DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
			estimated_margin DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
			actual_margin DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
			estimated_margin_percent DECIMAL(5, 2) NOT NULL DEFAULT 0.00,
			actual_margin_percent DECIMAL(5, 2) NOT NULL DEFAULT 0.00,
			variance_amount DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
			variance_percent DECIMAL(5, 2) NOT NULL DEFAULT 0.00,
			financial_status VARCHAR(50) NOT NULL DEFAULT 'ESTIMATED',
			total_charges_count INT NOT NULL DEFAULT 0,
			pending_charges_count INT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uq_ship_fin_summary (org_id, shipment_id)
		)`,
		`CREATE TABLE IF NOT EXISTS shipment_tracking_positions (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			org_id BIGINT NOT NULL,
			shipment_id BIGINT NOT NULL,
			vessel_name VARCHAR(255) NULL,
			latitude DECIMAL(10, 6) NOT NULL,
			longitude DECIMAL(10, 6) NOT NULL,
			speed_knots DECIMAL(6, 2) NOT NULL DEFAULT 0.00,
			heading_degrees DECIMAL(6, 2) NOT NULL DEFAULT 0.00,
			location_name VARCHAR(255) NULL,
			tracking_source VARCHAR(100) NOT NULL DEFAULT 'CARRIER_API',
			data_freshness VARCHAR(50) NOT NULL DEFAULT 'RECENT',
			recorded_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_stp_org_shipment (org_id, shipment_id, recorded_at DESC),
			INDEX idx_stp_recorded (recorded_at DESC)
		)`,
		`CREATE TABLE IF NOT EXISTS shipment_document_discrepancies (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			org_id BIGINT NOT NULL,
			shipment_id BIGINT NOT NULL,
			field_name VARCHAR(100) NOT NULL,
			expected_value TEXT NULL,
			actual_value TEXT NULL,
			source_document VARCHAR(50) NOT NULL,
			target_document VARCHAR(50) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'OPEN',
			resolved_by BIGINT NULL,
			resolved_at DATETIME NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_ship_disc_shipment (org_id, shipment_id)
		)`,
		`CREATE TABLE IF NOT EXISTS shipment_tracking_alerts (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			org_id BIGINT NOT NULL,
			shipment_id BIGINT NOT NULL,
			alert_key VARCHAR(150) NOT NULL,
			alert_type VARCHAR(100) NOT NULL,
			severity VARCHAR(30) NOT NULL,
			title VARCHAR(255) NOT NULL,
			description TEXT NULL,
			status VARCHAR(30) NOT NULL DEFAULT 'OPEN',
			first_detected_at DATETIME NOT NULL,
			last_detected_at DATETIME NOT NULL,
			acknowledged_at DATETIME NULL,
			acknowledged_by BIGINT NULL,
			resolved_at DATETIME NULL,
			resolved_by BIGINT NULL,
			suppressed_at DATETIME NULL,
			suppressed_by BIGINT NULL,
			notification_count INT NOT NULL DEFAULT 0,
			last_notified_at DATETIME NULL,
			metadata JSON NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uq_org_shipment_alert_key (org_id, shipment_id, alert_key),
			INDEX idx_sta_org_shipment_status (org_id, shipment_id, status),
			INDEX idx_sta_org_status (org_id, status)
		)`,
		`CREATE TABLE IF NOT EXISTS shipment_tracking_refresh_runs (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			org_id BIGINT NOT NULL,
			shipment_id BIGINT NOT NULL,
			provider_name VARCHAR(100) NULL,
			provider_type VARCHAR(50) NULL,
			trigger_type VARCHAR(50) NOT NULL DEFAULT 'MANUAL',
			status VARCHAR(50) NOT NULL DEFAULT 'STARTED',
			started_at DATETIME NOT NULL,
			completed_at DATETIME NULL,
			new_positions INT NOT NULL DEFAULT 0,
			new_events INT NOT NULL DEFAULT 0,
			data_freshness VARCHAR(50) NULL,
			used_fallback BOOLEAN NOT NULL DEFAULT FALSE,
			error_message TEXT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_trkr_org_shipment_started (org_id, shipment_id, started_at DESC),
			INDEX idx_trkr_org_status_started (org_id, status, started_at DESC)
		)`,
	}
	for _, stmt := range stmts {
		_, _ = db.Exec(stmt)
	}
}

func NewRepository(db *sqlx.DB) Repository {
	autoMigrateDocumentAndFinancialSchema(db)
	return &repository{db: db}
}

func (r *repository) CreateShipment(ctx context.Context, s *spec.Shipment) error {
	containersJSON, err := json.Marshal(s.ContainerNumbers)
	if err != nil {
		containersJSON = []byte("[]")
	}
	query := `
		INSERT INTO shipments (
			org_id, rfq_id, quote_id, booking_id, carrier_scac, booking_number, mbl_number, hbl_number,
			container_numbers, status, origin_port, destination_port, vessel_name, voyage_number, etd, eta, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
		)
	`
	res, err := r.db.ExecContext(ctx, query,
		s.OrgID, s.RFQID, s.QuoteID, s.BookingID, s.CarrierSCAC, s.BookingNumber, s.MBLNumber, s.HBLNumber,
		containersJSON, s.Status, s.OriginPort, s.DestinationPort, s.VesselName, s.VoyageNumber, s.ETD, s.ETA,
	)
	if err != nil {
		return err
	}
	s.ID, err = res.LastInsertId()
	return err
}

func (r *repository) CreateShipmentTx(ctx context.Context, tx *sqlx.Tx, s *spec.Shipment) error {
	containersJSON, err := json.Marshal(s.ContainerNumbers)
	if err != nil {
		containersJSON = []byte("[]")
	}
	query := `
		INSERT INTO shipments (
			org_id, rfq_id, quote_id, booking_id, carrier_scac, booking_number, mbl_number, hbl_number,
			container_numbers, status, origin_port, destination_port, vessel_name, voyage_number, etd, eta, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
		)
	`
	res, err := tx.ExecContext(ctx, query,
		s.OrgID, s.RFQID, s.QuoteID, s.BookingID, s.CarrierSCAC, s.BookingNumber, s.MBLNumber, s.HBLNumber,
		containersJSON, s.Status, s.OriginPort, s.DestinationPort, s.VesselName, s.VoyageNumber, s.ETD, s.ETA,
	)
	if err != nil {
		return err
	}
	s.ID, err = res.LastInsertId()
	return err
}

func (r *repository) GetShipmentByID(ctx context.Context, orgID int64, id int64) (*spec.Shipment, error) {
	query := `
		SELECT 
			s.id, s.org_id, s.rfq_id, s.quote_id, s.booking_id, s.booking_number,
			s.mbl_number, s.hbl_number, s.carrier_scac, s.vessel_name, s.voyage_number,
			s.origin_port, s.destination_port, s.container_numbers, s.status, s.etd, s.eta,
			s.created_at, s.updated_at, s.closure_status,
			r.rfq_number AS rfq_number,
			c.name AS customer_name,
			b.carrier_name AS carrier_name
		FROM shipments s
		LEFT JOIN rfqs r ON s.rfq_id = r.id AND r.org_id = s.org_id
		LEFT JOIN customers c ON r.customer_id = c.id
		LEFT JOIN bookings b ON s.booking_id = b.id AND b.org_id = s.org_id
		WHERE s.id = ? AND s.org_id = ?
	`
	var s spec.Shipment
	err := r.db.GetContext(ctx, &s, query, id, orgID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *repository) GetShipmentByRFQID(ctx context.Context, rfqID int64) (*spec.Shipment, error) {
	query := `SELECT * FROM shipments WHERE rfq_id = ?`
	var s spec.Shipment
	err := r.db.GetContext(ctx, &s, query, rfqID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *repository) GetShipmentByBooking(ctx context.Context, orgID int64, bookingNumber string) (*spec.Shipment, error) {
	query := `SELECT * FROM shipments WHERE booking_number = ? AND org_id = ?`
	var s spec.Shipment
	err := r.db.GetContext(ctx, &s, query, bookingNumber, orgID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *repository) GetShipmentByContainer(ctx context.Context, orgID int64, containerNumber string) (*spec.Shipment, error) {
	query := `SELECT * FROM shipments WHERE JSON_CONTAINS(container_numbers, JSON_QUOTE(?)) AND org_id = ?`
	var s spec.Shipment
	err := r.db.GetContext(ctx, &s, query, containerNumber, orgID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *repository) GetShipmentByMBL(ctx context.Context, orgID int64, mblNumber string) (*spec.Shipment, error) {
	query := `SELECT * FROM shipments WHERE mbl_number = ? AND org_id = ?`
	var s spec.Shipment
	err := r.db.GetContext(ctx, &s, query, mblNumber, orgID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *repository) GetShipmentByHBL(ctx context.Context, orgID int64, hblNumber string) (*spec.Shipment, error) {
	query := `SELECT * FROM shipments WHERE hbl_number = ? AND org_id = ?`
	var s spec.Shipment
	err := r.db.GetContext(ctx, &s, query, hblNumber, orgID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *repository) FindShipmentsByBooking(ctx context.Context, orgID int64, bookingNumber string) ([]*spec.Shipment, error) {
	query := `SELECT * FROM shipments WHERE booking_number = ? AND org_id = ?`
	var list []*spec.Shipment
	err := r.db.SelectContext(ctx, &list, query, bookingNumber, orgID)
	return list, err
}

func (r *repository) FindShipmentsByContainer(ctx context.Context, orgID int64, containerNumber string) ([]*spec.Shipment, error) {
	query := `SELECT * FROM shipments WHERE JSON_CONTAINS(container_numbers, JSON_QUOTE(?)) AND org_id = ?`
	var list []*spec.Shipment
	err := r.db.SelectContext(ctx, &list, query, containerNumber, orgID)
	return list, err
}

func (r *repository) FindShipmentsByMBL(ctx context.Context, orgID int64, mblNumber string) ([]*spec.Shipment, error) {
	query := `SELECT * FROM shipments WHERE mbl_number = ? AND org_id = ?`
	var list []*spec.Shipment
	err := r.db.SelectContext(ctx, &list, query, mblNumber, orgID)
	return list, err
}

func (r *repository) FindShipmentsByHBL(ctx context.Context, orgID int64, hblNumber string) ([]*spec.Shipment, error) {
	query := `SELECT * FROM shipments WHERE hbl_number = ? AND org_id = ?`
	var list []*spec.Shipment
	err := r.db.SelectContext(ctx, &list, query, hblNumber, orgID)
	return list, err
}

func (r *repository) ListShipments(ctx context.Context, orgID int64) ([]*spec.Shipment, error) {
	query := `SELECT * FROM shipments WHERE org_id = ? ORDER BY created_at DESC`
	var shipments []*spec.Shipment
	err := r.db.SelectContext(ctx, &shipments, query, orgID)
	return shipments, err
}

func (r *repository) UpdateShipment(ctx context.Context, s *spec.Shipment) error {
	containersJSON, err := json.Marshal(s.ContainerNumbers)
	if err != nil {
		containersJSON = []byte("[]")
	}
	query := `
		UPDATE shipments
		SET booking_number = ?,
			mbl_number = ?,
			hbl_number = ?,
			container_numbers = ?,
			status = ?,
			vessel_name = ?,
			voyage_number = ?,
			etd = ?,
			eta = ?,
			updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err = r.db.ExecContext(ctx, query,
		s.BookingNumber, s.MBLNumber, s.HBLNumber, containersJSON,
		s.Status, s.VesselName, s.VoyageNumber, s.ETD, s.ETA,
		s.ID, s.OrgID,
	)
	return err
}

func (r *repository) GetMilestones(ctx context.Context, shipmentID int64) ([]*spec.ShipmentMilestone, error) {
	query := `
		SELECT id, shipment_id, milestone_code, description, planned_date, actual_date, status, location, notes, source_event_id
		FROM shipment_milestones
		WHERE shipment_id = ?
		ORDER BY id ASC
	`
	var milestones []*spec.ShipmentMilestone
	err := r.db.SelectContext(ctx, &milestones, query, shipmentID)
	return milestones, err
}

func (r *repository) CreateMilestone(ctx context.Context, m *spec.ShipmentMilestone) error {
	query := `
		INSERT INTO shipment_milestones (
			shipment_id, milestone_code, description, planned_date, actual_date, status, location, notes
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?
		)
	`
	res, err := r.db.ExecContext(ctx, query, m.ShipmentID, m.MilestoneCode, m.Description, m.PlannedDate, m.ActualDate, m.Status, m.Location, m.Notes)
	if err != nil {
		return err
	}
	m.ID, err = res.LastInsertId()
	return err
}

func (r *repository) CreateMilestoneTx(ctx context.Context, tx *sqlx.Tx, m *spec.ShipmentMilestone) error {
	query := `
		INSERT INTO shipment_milestones (
			shipment_id, milestone_code, description, planned_date, actual_date, status, location, notes
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?
		)
	`
	res, err := tx.ExecContext(ctx, query, m.ShipmentID, m.MilestoneCode, m.Description, m.PlannedDate, m.ActualDate, m.Status, m.Location, m.Notes)
	if err != nil {
		return err
	}
	m.ID, err = res.LastInsertId()
	return err
}

func (r *repository) UpdateMilestone(ctx context.Context, m *spec.ShipmentMilestone) error {
	query := `
		UPDATE shipment_milestones
		SET actual_date = ?,
			status = ?,
			location = ?,
			notes = ?,
			updated_at = NOW()
		WHERE id = ? AND shipment_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, m.ActualDate, m.Status, m.Location, m.Notes, m.ID, m.ShipmentID)
	return err
}

func (r *repository) GetExceptions(ctx context.Context, orgID int64, shipmentID int64) ([]*spec.ShipmentException, error) {
	query := `
		SELECT id, org_id, shipment_id, exception_type, severity, status, title, description, resolved, resolved_at, resolved_by, resolution_notes, ai_summary, source_event_id, created_at, updated_at
		FROM shipment_exceptions
		WHERE shipment_id = ? AND org_id = ?
		ORDER BY created_at DESC
	`
	var exceptions []*spec.ShipmentException
	err := r.db.SelectContext(ctx, &exceptions, query, shipmentID, orgID)
	return exceptions, err
}

func (r *repository) CreateException(ctx context.Context, ex *spec.ShipmentException) error {
	query := `
		INSERT INTO shipment_exceptions (
			org_id, shipment_id, exception_type, severity, status, title, description, resolved, resolved_at, resolved_by, resolution_notes, ai_summary, source_event_id, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
		)
	`
	res, err := r.db.ExecContext(ctx, query,
		ex.OrgID, ex.ShipmentID, ex.ExceptionType, ex.Severity, ex.Status, ex.Title, ex.Description,
		ex.Resolved, ex.ResolvedAt, ex.ResolvedBy, ex.ResolutionNotes, ex.AISummary, ex.SourceEventID,
	)
	if err != nil {
		return err
	}
	ex.ID, err = res.LastInsertId()
	return err
}

func (r *repository) GetExceptionByID(ctx context.Context, orgID int64, id int64) (*spec.ShipmentException, error) {
	query := `
		SELECT id, org_id, shipment_id, exception_type, severity, status, title, description, resolved, resolved_at, resolved_by, resolution_notes, ai_summary, source_event_id, created_at, updated_at
		FROM shipment_exceptions
		WHERE id = ? AND org_id = ?
	`
	var ex spec.ShipmentException
	err := r.db.GetContext(ctx, &ex, query, id, orgID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &ex, err
}

func (r *repository) UpdateException(ctx context.Context, ex *spec.ShipmentException) error {
	query := `
		UPDATE shipment_exceptions
		SET status = ?,
			severity = ?,
			resolved = ?,
			resolved_at = ?,
			resolved_by = ?,
			resolution_notes = ?,
			description = ?,
			updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		ex.Status, ex.Severity, ex.Resolved, ex.ResolvedAt, ex.ResolvedBy, ex.ResolutionNotes, ex.Description,
		ex.ID, ex.OrgID,
	)
	return err
}

func (r *repository) IsEventProcessed(ctx context.Context, eventID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM carrier_tracking_events WHERE event_id = ? AND processing_status = 'PROCESSED')`
	var exists bool
	err := r.db.GetContext(ctx, &exists, query, eventID)
	return exists, err
}

func (r *repository) MarkEventProcessed(ctx context.Context, eventID string, shipmentID int64) error {
	return nil
}

func (r *repository) InsertCarrierEvent(ctx context.Context, event *spec.CarrierTrackingEvent) (bool, error) {
	query := `
		INSERT INTO carrier_tracking_events (
			org_id, event_id, source_type, carrier_scac, booking_number, container_number,
			mbl_number, hbl_number, vessel_name, voyage_number, milestone_code, event_time,
			location, raw_description, raw_payload, shipment_id, matching_status, processing_status
		) VALUES (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?
		) ON DUPLICATE KEY UPDATE updated_at = NOW()
	`
	res, err := r.db.ExecContext(ctx, query,
		event.OrgID, event.EventID, event.SourceType, event.CarrierSCAC, event.BookingNumber, event.ContainerNumber,
		event.MBLNumber, event.HBLNumber, event.VesselName, event.VoyageNumber, event.MilestoneCode, event.EventTime,
		event.Location, event.RawDescription, event.RawPayload, event.ShipmentID, event.MatchingStatus, event.ProcessingStatus,
	)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	return affected > 0, err
}

func (r *repository) UpdateCarrierEventStatus(ctx context.Context, eventID string, orgID int64, matchingStatus string, processingStatus string, shipmentID *int64) error {
	query := `
		UPDATE carrier_tracking_events
		SET matching_status = ?, processing_status = ?, shipment_id = ?, updated_at = NOW()
		WHERE event_id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, matchingStatus, processingStatus, shipmentID, eventID, orgID)
	return err
}

func (r *repository) FindSCACByCarrierName(ctx context.Context, carrierName string) (string, error) {
	query := `SELECT scac FROM carriers WHERE name = ? OR scac = ? OR name LIKE ? LIMIT 1`
	var scac string
	err := r.db.GetContext(ctx, &scac, query, carrierName, carrierName, "%"+carrierName+"%")
	return scac, err
}

func (r *repository) CreateActivity(ctx context.Context, orgID int64, entityType string, entityID int64, action string, description string, actor string) error {
	var userID *int64
	if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.UserID > 0 {
		userID = &userCtx.UserID
	}
	query := `
		INSERT INTO activities (org_id, entity_type, entity_id, action, description, user_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW())
	`
	_, err := r.db.ExecContext(ctx, query, orgID, entityType, entityID, action, description, userID)
	return err
}

func (r *repository) GetShipmentsWorkspace(ctx context.Context, orgID int64, filter spec.ShipmentListFilter) ([]*spec.Shipment, spec.ShipmentKPIs, int, error) {
	// 1. Calculate operational KPIs (real MySQL persisted data only)
	var kpis spec.ShipmentKPIs
	kpiQuery := `
		SELECT 
			COUNT(*) AS total_shipments,
			COUNT(CASE WHEN status = 'BOOKING_PENDING' THEN 1 END) AS booking_pending,
			COUNT(CASE WHEN status = 'BOOKED' THEN 1 END) AS booked,
			COUNT(CASE WHEN status IN ('DEPARTED', 'IN_TRANSIT') THEN 1 END) AS in_transit,
			COUNT(CASE WHEN status = 'ARRIVED' THEN 1 END) AS arrived,
			COUNT(CASE WHEN status = 'DELIVERED' THEN 1 END) AS delivered,
			COUNT(CASE WHEN (SELECT COUNT(*) FROM shipment_exceptions WHERE shipment_id = id AND status NOT IN ('RESOLVED', 'DISMISSED')) > 0 THEN 1 END) AS exceptions
		FROM shipments
		WHERE org_id = ?
	`
	_ = r.db.GetContext(ctx, &kpis, kpiQuery, orgID)

	// 2. Build listing query with joins
	baseQuery := `
		FROM shipments s
		LEFT JOIN rfqs r ON s.rfq_id = r.id AND r.org_id = s.org_id
		LEFT JOIN customers c ON r.customer_id = c.id
		LEFT JOIN bookings b ON s.booking_id = b.id AND b.org_id = s.org_id
		WHERE s.org_id = ?
	`
	args := []interface{}{orgID}

	if filter.Status != nil && *filter.Status != "" && *filter.Status != "ALL" {
		baseQuery += " AND s.status = ?"
		args = append(args, *filter.Status)
	}

	if filter.Search != nil && *filter.Search != "" {
		searchPattern := "%" + *filter.Search + "%"
		baseQuery += ` AND (
			s.booking_number LIKE ? OR
			r.rfq_number LIKE ? OR
			c.name LIKE ? OR
			s.vessel_name LIKE ? OR
			s.voyage_number LIKE ? OR
			s.origin_port LIKE ? OR
			s.destination_port LIKE ?
		)`
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// Calculate total items matching filter
	var totalItems int
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.db.GetContext(ctx, &totalItems, countQuery, args...)
	if err != nil {
		return nil, kpis, 0, err
	}

	// Apply sorting and pagination
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	selectQuery := `
		SELECT 
			s.id, s.org_id, s.rfq_id, s.quote_id, s.booking_id, s.booking_number,
			s.mbl_number, s.hbl_number, s.carrier_scac, s.vessel_name, s.voyage_number,
			s.origin_port, s.destination_port, s.container_numbers, s.status, s.etd, s.eta,
			s.created_at, s.updated_at, s.closure_status,
			r.rfq_number AS rfq_number,
			c.name AS customer_name,
			b.carrier_name AS carrier_name,
			(SELECT COUNT(*) FROM shipment_exceptions WHERE shipment_id = s.id AND status NOT IN ('RESOLVED', 'DISMISSED')) AS active_exceptions_count,
			(SELECT COUNT(*) FROM shipment_exceptions WHERE shipment_id = s.id AND status NOT IN ('RESOLVED', 'DISMISSED') AND severity IN ('HIGH', 'CRITICAL')) AS high_exceptions_count
	` + baseQuery + `
		ORDER BY s.created_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)

	list := make([]*spec.Shipment, 0)
	err = r.db.SelectContext(ctx, &list, selectQuery, args...)
	if err != nil {
		return nil, kpis, 0, err
	}

	return list, kpis, totalItems, nil
}

func (r *repository) UpdateClosureStatus(ctx context.Context, orgID int64, shipmentID int64, status string) error {
	query := `UPDATE shipments SET closure_status = ?, updated_at = NOW() WHERE id = ? AND org_id = ?`
	_, err := r.db.ExecContext(ctx, query, status, shipmentID, orgID)
	return err
}

// ─── Shipment Document Repository Methods (Task 16.7) ───────────────────────

func (r *repository) GetShipmentDocuments(ctx context.Context, orgID int64, shipmentID int64) ([]*spec.ShipmentDocument, error) {
	query := `
		SELECT 
			id, org_id, shipment_id, doc_type, document_name, category, description,
			s3_key, file_name, file_url, file_size, mime_type, file_type, status,
			uploaded_by, uploaded_at, reviewed_by, reviewed_at, rejection_reason, expires_at,
			document_date, reference_number,
			source, source_id, extracted_data, raw_ocr_text, ai_summary, created_at, updated_at
		FROM shipment_documents
		WHERE shipment_id = ? AND org_id = ?
		ORDER BY created_at DESC
	`
	var list []*spec.ShipmentDocument
	err := r.db.SelectContext(ctx, &list, query, shipmentID, orgID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]*spec.ShipmentDocument, 0)
	}
	return list, nil
}

func (r *repository) GetShipmentDocumentByID(ctx context.Context, orgID int64, shipmentID int64, docID int64) (*spec.ShipmentDocument, error) {
	query := `
		SELECT 
			id, org_id, shipment_id, doc_type, document_name, category, description,
			s3_key, file_name, file_url, file_size, mime_type, file_type, status,
			uploaded_by, uploaded_at, reviewed_by, reviewed_at, rejection_reason, expires_at,
			document_date, reference_number,
			source, source_id, extracted_data, raw_ocr_text, ai_summary, created_at, updated_at
		FROM shipment_documents
		WHERE id = ? AND shipment_id = ? AND org_id = ?
		LIMIT 1
	`
	var doc spec.ShipmentDocument
	err := r.db.GetContext(ctx, &doc, query, docID, shipmentID, orgID)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *repository) CreateShipmentDocument(ctx context.Context, orgID int64, doc *spec.ShipmentDocument) error {
	query := `
		INSERT INTO shipment_documents (
			org_id, shipment_id, doc_type, document_name, category, description,
			s3_key, file_name, file_url, file_size, mime_type, file_type, status,
			uploaded_by, uploaded_at, reviewed_by, reviewed_at, rejection_reason, expires_at,
			document_date, reference_number,
			source, source_id, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?,
			?, ?, NOW(), NOW()
		)
	`
	res, err := r.db.ExecContext(ctx, query,
		orgID, doc.ShipmentID, doc.DocType, doc.DocumentName, doc.Category, doc.Description,
		doc.S3Key, doc.FileName, doc.FileURL, doc.FileSize, doc.MimeType, doc.FileType, doc.Status,
		doc.UploadedBy, doc.UploadedAt, doc.ReviewedBy, doc.ReviewedAt, doc.RejectionReason, doc.ExpiresAt,
		doc.DocumentDate, doc.ReferenceNumber,
		doc.Source, doc.SourceID,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil && id > 0 {
		doc.ID = id
	}
	return nil
}

func (r *repository) UpdateShipmentDocument(ctx context.Context, orgID int64, doc *spec.ShipmentDocument) error {
	query := `
		UPDATE shipment_documents SET
			document_name = ?,
			category = ?,
			description = ?,
			status = ?,
			reviewed_by = ?,
			reviewed_at = ?,
			rejection_reason = ?,
			expires_at = ?,
			document_date = ?,
			reference_number = ?,
			updated_at = NOW()
		WHERE id = ? AND shipment_id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		doc.DocumentName, doc.Category, doc.Description, doc.Status,
		doc.ReviewedBy, doc.ReviewedAt, doc.RejectionReason, doc.ExpiresAt,
		doc.DocumentDate, doc.ReferenceNumber,
		doc.ID, doc.ShipmentID, orgID,
	)
	return err
}

func (r *repository) DeleteShipmentDocument(ctx context.Context, orgID int64, shipmentID int64, docID int64) error {
	query := `DELETE FROM shipment_documents WHERE id = ? AND shipment_id = ? AND org_id = ?`
	_, err := r.db.ExecContext(ctx, query, docID, shipmentID, orgID)
	return err
}

func (r *repository) GetUpstreamRFQDocuments(ctx context.Context, orgID int64, rfqID int64) ([]*spec.ShipmentDocument, error) {
	query := `
		SELECT 
			id, org_id, rfq_id AS source_id, document_type AS doc_type, document_name, description,
			status, file_name, file_url, file_size, mime_type, uploaded_by, uploaded_at,
			reviewed_by, reviewed_at, rejection_reason, expires_at, created_at, updated_at
		FROM rfq_documents
		WHERE rfq_id = ? AND org_id = ?
		ORDER BY created_at DESC
	`
	type dbRFQDoc struct {
		ID              int64      `db:"id"`
		OrgID           int64      `db:"org_id"`
		SourceID        int64      `db:"source_id"`
		DocType         string     `db:"doc_type"`
		DocumentName    *string    `db:"document_name"`
		Description     *string    `db:"description"`
		Status          string     `db:"status"`
		FileName        *string    `db:"file_name"`
		FileURL         *string    `db:"file_url"`
		FileSize        *int64     `db:"file_size"`
		MimeType        *string    `db:"mime_type"`
		UploadedBy      *string    `db:"uploaded_by"`
		UploadedAt      *time.Time `db:"uploaded_at"`
		ReviewedBy      *string    `db:"reviewed_by"`
		ReviewedAt      *time.Time `db:"reviewed_at"`
		RejectionReason *string    `db:"rejection_reason"`
		ExpiresAt       *time.Time `db:"expires_at"`
		CreatedAt       time.Time  `db:"created_at"`
		UpdatedAt       time.Time  `db:"updated_at"`
	}

	var rfqDocs []dbRFQDoc
	err := r.db.SelectContext(ctx, &rfqDocs, query, rfqID, orgID)
	if err != nil {
		return make([]*spec.ShipmentDocument, 0), nil
	}

	res := make([]*spec.ShipmentDocument, 0, len(rfqDocs))
	for _, d := range rfqDocs {
		fName := ""
		if d.FileName != nil {
			fName = *d.FileName
		}
		doc := &spec.ShipmentDocument{
			ID:              d.ID,
			OrgID:           d.OrgID,
			DocType:         d.DocType,
			DocumentName:    d.DocumentName,
			Category:        InferCategory(d.DocType),
			Description:     d.Description,
			FileName:        fName,
			FileURL:         d.FileURL,
			FileSize:        d.FileSize,
			MimeType:        d.MimeType,
			Status:          d.Status,
			UploadedBy:      d.UploadedBy,
			UploadedAt:      d.UploadedAt,
			ReviewedBy:      d.ReviewedBy,
			ReviewedAt:      d.ReviewedAt,
			RejectionReason: d.RejectionReason,
			ExpiresAt:       d.ExpiresAt,
			Source:          spec.DocSourceRFQ,
			SourceID:        &d.SourceID,
			CreatedAt:       d.CreatedAt,
			UpdatedAt:       d.UpdatedAt,
		}
		res = append(res, doc)
	}
	return res, nil
}

func (r *repository) GetShipmentDiscrepancies(ctx context.Context, orgID int64, shipmentID int64) ([]*spec.ShipmentDocumentDiscrepancy, error) {
	query := `
		SELECT 
			id, shipment_id, field_name, source_document, target_document,
			COALESCE(expected_value, '') AS source_value,
			COALESCE(actual_value, '') AS target_value,
			'WARNING' AS severity,
			CASE WHEN status = 'RESOLVED' THEN true ELSE false END AS resolved,
			created_at
		FROM shipment_document_discrepancies
		WHERE shipment_id = ? AND org_id = ?
		ORDER BY created_at DESC
	`
	var list []*spec.ShipmentDocumentDiscrepancy
	err := r.db.SelectContext(ctx, &list, query, shipmentID, orgID)
	if err != nil {
		// Fallback query if org_id is not in discrepancies table
		fallbackQuery := `
			SELECT 
				id, shipment_id, field_name, source_document, target_document,
				COALESCE(expected_value, '') AS source_value,
				COALESCE(actual_value, '') AS target_value,
				'WARNING' AS severity,
				CASE WHEN status = 'RESOLVED' THEN true ELSE false END AS resolved,
				created_at
			FROM shipment_document_discrepancies
			WHERE shipment_id = ?
			ORDER BY created_at DESC
		`
		if fbErr := r.db.SelectContext(ctx, &list, fallbackQuery, shipmentID); fbErr != nil {
			return make([]*spec.ShipmentDocumentDiscrepancy, 0), nil
		}
	}
	if list == nil {
		list = make([]*spec.ShipmentDocumentDiscrepancy, 0)
	}
	return list, nil
}

// ─── Financial Operations Repository Methods (Task 16.8) ────────────────────

func (r *repository) GetShipmentCharges(ctx context.Context, orgID int64, shipmentID int64) ([]*spec.ShipmentFinancialCharge, error) {
	query := `
		SELECT 
			id, org_id, shipment_id, booking_id, rfq_id, category, charge_type,
			description, vendor_name, estimated_amount, actual_amount, currency,
			reference_number, charge_date, status, notes, created_at, updated_at
		FROM shipment_financial_charges
		WHERE shipment_id = ? AND org_id = ?
		ORDER BY created_at ASC
	`
	var list []*spec.ShipmentFinancialCharge
	err := r.db.SelectContext(ctx, &list, query, shipmentID, orgID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]*spec.ShipmentFinancialCharge, 0)
	}
	return list, nil
}

func (r *repository) GetShipmentChargeByID(ctx context.Context, orgID int64, shipmentID int64, chargeID int64) (*spec.ShipmentFinancialCharge, error) {
	query := `
		SELECT 
			id, org_id, shipment_id, booking_id, rfq_id, category, charge_type,
			description, vendor_name, estimated_amount, actual_amount, currency,
			reference_number, charge_date, status, notes, created_at, updated_at
		FROM shipment_financial_charges
		WHERE id = ? AND shipment_id = ? AND org_id = ?
		LIMIT 1
	`
	var ch spec.ShipmentFinancialCharge
	err := r.db.GetContext(ctx, &ch, query, chargeID, shipmentID, orgID)
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *repository) CreateShipmentCharge(ctx context.Context, orgID int64, charge *spec.ShipmentFinancialCharge) error {
	query := `
		INSERT INTO shipment_financial_charges (
			org_id, shipment_id, booking_id, rfq_id, category, charge_type,
			description, vendor_name, estimated_amount, actual_amount, currency,
			reference_number, charge_date, status, notes, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, NOW(), NOW()
		)
	`
	res, err := r.db.ExecContext(ctx, query,
		orgID, charge.ShipmentID, charge.BookingID, charge.RFQID, charge.Category, charge.ChargeType,
		charge.Description, charge.VendorName, charge.EstimatedAmount, charge.ActualAmount, charge.Currency,
		charge.ReferenceNumber, charge.ChargeDate, charge.Status, charge.Notes,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil && id > 0 {
		charge.ID = id
	}
	return nil
}

func (r *repository) UpdateShipmentCharge(ctx context.Context, orgID int64, charge *spec.ShipmentFinancialCharge) error {
	query := `
		UPDATE shipment_financial_charges SET
			category = ?,
			charge_type = ?,
			description = ?,
			vendor_name = ?,
			estimated_amount = ?,
			actual_amount = ?,
			currency = ?,
			reference_number = ?,
			charge_date = ?,
			status = ?,
			notes = ?,
			updated_at = NOW()
		WHERE id = ? AND shipment_id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		charge.Category, charge.ChargeType, charge.Description, charge.VendorName,
		charge.EstimatedAmount, charge.ActualAmount, charge.Currency, charge.ReferenceNumber,
		charge.ChargeDate, charge.Status, charge.Notes,
		charge.ID, charge.ShipmentID, orgID,
	)
	return err
}

func (r *repository) DeleteShipmentCharge(ctx context.Context, orgID int64, shipmentID int64, chargeID int64) error {
	query := `DELETE FROM shipment_financial_charges WHERE id = ? AND shipment_id = ? AND org_id = ?`
	_, err := r.db.ExecContext(ctx, query, chargeID, shipmentID, orgID)
	return err
}

func (r *repository) GetQuoteCommercialSnapshot(ctx context.Context, quoteID int64) (*RFQQuoteCommercialSnapshot, error) {
	query := `
		SELECT 
			id, rfq_id, carrier_name, buy_price, sell_price, ocean_freight,
			origin_charges, destination_charges, total_buy_price, currency, status
		FROM rfq_quotes
		WHERE id = ?
		LIMIT 1
	`
	var q RFQQuoteCommercialSnapshot
	err := r.db.GetContext(ctx, &q, query, quoteID)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *repository) GetCarrierInvoicesSnapshot(ctx context.Context, orgID int64, shipmentID int64) ([]*CarrierInvoiceSnapshot, error) {
	query := `
		SELECT id, total_amount, currency, status
		FROM shipment_invoices
		WHERE shipment_id = ? AND org_id = ?
	`
	var list []*CarrierInvoiceSnapshot
	err := r.db.SelectContext(ctx, &list, query, shipmentID, orgID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]*CarrierInvoiceSnapshot, 0)
	}
	return list, nil
}

func (r *repository) GetCustomerInvoicesSnapshot(ctx context.Context, orgID int64, shipmentID int64) ([]*CustomerInvoiceSnapshot, error) {
	query := `
		SELECT id, total_amount, currency, status
		FROM shipment_customer_invoices
		WHERE shipment_id = ? AND org_id = ?
	`
	var list []*CustomerInvoiceSnapshot
	err := r.db.SelectContext(ctx, &list, query, shipmentID, orgID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]*CustomerInvoiceSnapshot, 0)
	}
	return list, nil
}

func (r *repository) SaveShipmentProfitabilitySnapshot(ctx context.Context, orgID int64, shipmentID int64, sellAmount float64, buyAmount float64, netProfit float64, marginPct float64) error {
	query := `
		INSERT INTO shipment_finance_profitability (
			shipment_id, org_id, total_sell_amount, total_buy_amount, net_profit, profit_margin_pct, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, NOW()
		)
		ON DUPLICATE KEY UPDATE
			total_sell_amount = VALUES(total_sell_amount),
			total_buy_amount = VALUES(total_buy_amount),
			net_profit = VALUES(net_profit),
			profit_margin_pct = VALUES(profit_margin_pct),
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query, shipmentID, orgID, sellAmount, buyAmount, netProfit, marginPct)
	return err
}

func (r *repository) SaveTrackingPosition(ctx context.Context, pos *spec.TrackingPosition) error {
	// Idempotency: check if identical position was already recorded recently for this shipment
	var existingID int64
	checkQuery := `
		SELECT id FROM shipment_tracking_positions
		WHERE org_id = ? AND shipment_id = ?
		  AND ABS(latitude - ?) < 0.0001 AND ABS(longitude - ?) < 0.0001
		  AND recorded_at >= ?
		LIMIT 1
	`
	timeThreshold := pos.RecordedAt.Add(-10 * time.Minute)
	err := r.db.GetContext(ctx, &existingID, checkQuery, pos.OrgID, pos.ShipmentID, pos.Latitude, pos.Longitude, timeThreshold)
	if err == nil && existingID > 0 {
		pos.ID = existingID
		return nil // Reuse existing record without duplicate insertion
	}

	query := `
		INSERT INTO shipment_tracking_positions (
			org_id, shipment_id, vessel_name, latitude, longitude,
			speed_knots, heading_degrees, location_name, tracking_source,
			data_freshness, recorded_at, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, NOW(), NOW()
		)
	`
	res, err := r.db.ExecContext(ctx, query,
		pos.OrgID, pos.ShipmentID, pos.VesselName, pos.Latitude, pos.Longitude,
		pos.SpeedKnots, pos.HeadingDegrees, pos.LocationName, pos.TrackingSource,
		pos.DataFreshness, pos.RecordedAt,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		pos.ID = id
	}
	return nil
}

func (r *repository) GetLatestPosition(ctx context.Context, orgID int64, shipmentID int64) (*spec.TrackingPosition, error) {
	query := `
		SELECT id, org_id, shipment_id, vessel_name, latitude, longitude,
		       speed_knots, heading_degrees, location_name, tracking_source,
		       data_freshness, recorded_at, created_at, updated_at
		FROM shipment_tracking_positions
		WHERE org_id = ? AND shipment_id = ?
		ORDER BY recorded_at DESC
		LIMIT 1
	`
	var pos spec.TrackingPosition
	err := r.db.GetContext(ctx, &pos, query, orgID, shipmentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &pos, nil
}

func (r *repository) GetPositionHistory(ctx context.Context, orgID int64, shipmentID int64, limit int) ([]*spec.TrackingPosition, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	query := `
		SELECT id, org_id, shipment_id, vessel_name, latitude, longitude,
		       speed_knots, heading_degrees, location_name, tracking_source,
		       data_freshness, recorded_at, created_at, updated_at
		FROM shipment_tracking_positions
		WHERE org_id = ? AND shipment_id = ?
		ORDER BY recorded_at DESC
		LIMIT ?
	`
	var list []*spec.TrackingPosition
	err := r.db.SelectContext(ctx, &list, query, orgID, shipmentID, limit)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]*spec.TrackingPosition, 0)
	}
	return list, nil
}

func (r *repository) GetCarrierEventsForShipment(ctx context.Context, orgID int64, shipmentID int64) ([]*spec.CarrierTrackingEvent, error) {
	query := `
		SELECT id, org_id, event_id, source_type, carrier_scac,
		       booking_number, container_number, mbl_number, hbl_number,
		       vessel_name, voyage_number, milestone_code, event_time,
		       location, raw_description, raw_payload, shipment_id,
		       matching_status, processing_status, received_at, updated_at
		FROM carrier_tracking_events
		WHERE org_id = ? AND shipment_id = ?
		ORDER BY event_time DESC, received_at DESC
	`
	var list []*spec.CarrierTrackingEvent
	err := r.db.SelectContext(ctx, &list, query, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]*spec.CarrierTrackingEvent, 0)
	}
	return list, nil
}

// ── Persistent Tracking Alerts (Task 17.5) ──────────────────────────────────

func (r *repository) GetTrackingAlerts(ctx context.Context, orgID int64, shipmentID int64, status string) ([]*spec.ShipmentTrackingAlertRecord, error) {
	query := `
		SELECT sta.id, sta.org_id, sta.shipment_id, sta.alert_key, sta.alert_type,
		       sta.severity, sta.title, sta.description, sta.status,
		       sta.first_detected_at, sta.last_detected_at,
		       sta.acknowledged_at, sta.acknowledged_by,
		       sta.resolved_at, sta.resolved_by,
		       sta.suppressed_at, sta.suppressed_by,
		       sta.notification_count, sta.last_notified_at, sta.metadata,
		       sta.created_at, sta.updated_at,
		       CONCAT(COALESCE(u_ack.first_name, ''), ' ', COALESCE(u_ack.last_name, '')) AS acknowledged_by_name,
		       CONCAT(COALESCE(u_res.first_name, ''), ' ', COALESCE(u_res.last_name, '')) AS resolved_by_name
		FROM shipment_tracking_alerts sta
		LEFT JOIN users u_ack ON u_ack.id = sta.acknowledged_by
		LEFT JOIN users u_res ON u_res.id = sta.resolved_by
		WHERE sta.org_id = ? AND sta.shipment_id = ?
	`
	args := []interface{}{orgID, shipmentID}
	if status != "" && status != "ALL" {
		query += " AND sta.status = ?"
		args = append(args, status)
	}
	query += `
		ORDER BY CASE sta.status 
		           WHEN 'OPEN' THEN 1 
		           WHEN 'ACKNOWLEDGED' THEN 2 
		           WHEN 'SUPPRESSED' THEN 3 
		           ELSE 4 
		         END, 
		         sta.last_detected_at DESC
	`
	var list []*spec.ShipmentTrackingAlertRecord
	err := r.db.SelectContext(ctx, &list, query, args...)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]*spec.ShipmentTrackingAlertRecord, 0)
	}
	return list, nil
}

func (r *repository) GetTrackingAlertByID(ctx context.Context, orgID int64, shipmentID int64, alertID int64) (*spec.ShipmentTrackingAlertRecord, error) {
	query := `
		SELECT sta.id, sta.org_id, sta.shipment_id, sta.alert_key, sta.alert_type,
		       sta.severity, sta.title, sta.description, sta.status,
		       sta.first_detected_at, sta.last_detected_at,
		       sta.acknowledged_at, sta.acknowledged_by,
		       sta.resolved_at, sta.resolved_by,
		       sta.suppressed_at, sta.suppressed_by,
		       sta.notification_count, sta.last_notified_at, sta.metadata,
		       sta.created_at, sta.updated_at,
		       CONCAT(COALESCE(u_ack.first_name, ''), ' ', COALESCE(u_ack.last_name, '')) AS acknowledged_by_name,
		       CONCAT(COALESCE(u_res.first_name, ''), ' ', COALESCE(u_res.last_name, '')) AS resolved_by_name
		FROM shipment_tracking_alerts sta
		LEFT JOIN users u_ack ON u_ack.id = sta.acknowledged_by
		LEFT JOIN users u_res ON u_res.id = sta.resolved_by
		WHERE sta.id = ? AND sta.org_id = ? AND sta.shipment_id = ?
	`
	var record spec.ShipmentTrackingAlertRecord
	err := r.db.GetContext(ctx, &record, query, alertID, orgID, shipmentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func (r *repository) GetTrackingAlertByKey(ctx context.Context, orgID int64, shipmentID int64, alertKey string) (*spec.ShipmentTrackingAlertRecord, error) {
	query := `
		SELECT sta.id, sta.org_id, sta.shipment_id, sta.alert_key, sta.alert_type,
		       sta.severity, sta.title, sta.description, sta.status,
		       sta.first_detected_at, sta.last_detected_at,
		       sta.acknowledged_at, sta.acknowledged_by,
		       sta.resolved_at, sta.resolved_by,
		       sta.suppressed_at, sta.suppressed_by,
		       sta.notification_count, sta.last_notified_at, sta.metadata,
		       sta.created_at, sta.updated_at
		FROM shipment_tracking_alerts sta
		WHERE sta.org_id = ? AND sta.shipment_id = ? AND sta.alert_key = ?
	`
	var record spec.ShipmentTrackingAlertRecord
	err := r.db.GetContext(ctx, &record, query, orgID, shipmentID, alertKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func (r *repository) CreateTrackingAlert(ctx context.Context, alert *spec.ShipmentTrackingAlertRecord) error {
	query := `
		INSERT INTO shipment_tracking_alerts (
			org_id, shipment_id, alert_key, alert_type, severity, title, description,
			status, first_detected_at, last_detected_at, notification_count, metadata
		) VALUES (
			:org_id, :shipment_id, :alert_key, :alert_type, :severity, :title, :description,
			:status, :first_detected_at, :last_detected_at, :notification_count, :metadata
		)
		ON DUPLICATE KEY UPDATE
			last_detected_at = VALUES(last_detected_at),
			severity = VALUES(severity),
			title = VALUES(title),
			description = VALUES(description),
			metadata = VALUES(metadata)
	`
	res, err := r.db.NamedExecContext(ctx, query, alert)
	if err != nil {
		return err
	}
	if alert.ID == 0 {
		id, err := res.LastInsertId()
		if err == nil {
			alert.ID = id
		}
	}
	return nil
}

func (r *repository) UpdateTrackingAlert(ctx context.Context, alert *spec.ShipmentTrackingAlertRecord) error {
	query := `
		UPDATE shipment_tracking_alerts
		SET severity = :severity,
		    title = :title,
		    description = :description,
		    status = :status,
		    last_detected_at = :last_detected_at,
		    metadata = :metadata
		WHERE id = :id AND org_id = :org_id AND shipment_id = :shipment_id
	`
	_, err := r.db.NamedExecContext(ctx, query, alert)
	return err
}

func (r *repository) AcknowledgeTrackingAlert(ctx context.Context, orgID int64, shipmentID int64, alertID int64, userID *int64) error {
	query := `
		UPDATE shipment_tracking_alerts
		SET status = 'ACKNOWLEDGED',
		    acknowledged_at = NOW(),
		    acknowledged_by = ?
		WHERE id = ? AND org_id = ? AND shipment_id = ? AND status = 'OPEN'
	`
	_, err := r.db.ExecContext(ctx, query, userID, alertID, orgID, shipmentID)
	return err
}

func (r *repository) ResolveTrackingAlert(ctx context.Context, orgID int64, shipmentID int64, alertID int64, userID *int64) error {
	query := `
		UPDATE shipment_tracking_alerts
		SET status = 'RESOLVED',
		    resolved_at = NOW(),
		    resolved_by = ?
		WHERE id = ? AND org_id = ? AND shipment_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, userID, alertID, orgID, shipmentID)
	return err
}

func (r *repository) SuppressTrackingAlert(ctx context.Context, orgID int64, shipmentID int64, alertID int64, userID *int64) error {
	query := `
		UPDATE shipment_tracking_alerts
		SET status = 'SUPPRESSED',
		    suppressed_at = NOW(),
		    suppressed_by = ?
		WHERE id = ? AND org_id = ? AND shipment_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, userID, alertID, orgID, shipmentID)
	return err
}

func (r *repository) GetTrackingMonitoringSummary(ctx context.Context, orgID int64, shipmentID int64) (*spec.TrackingMonitoringSummary, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN status = 'OPEN' THEN 1 ELSE 0 END), 0) AS open_alerts,
			COALESCE(SUM(CASE WHEN status = 'OPEN' AND severity IN ('CRITICAL', 'HIGH') THEN 1 ELSE 0 END), 0) AS critical_alerts,
			COALESCE(SUM(CASE WHEN status = 'ACKNOWLEDGED' THEN 1 ELSE 0 END), 0) AS acknowledged_alerts,
			COALESCE(SUM(CASE WHEN status = 'SUPPRESSED' THEN 1 ELSE 0 END), 0) AS suppressed_alerts,
			COALESCE(SUM(CASE WHEN status = 'RESOLVED' THEN 1 ELSE 0 END), 0) AS resolved_alerts
		FROM shipment_tracking_alerts
		WHERE org_id = ? AND shipment_id = ?
	`
	type counts struct {
		Open         int `db:"open_alerts"`
		Critical     int `db:"critical_alerts"`
		Acknowledged int `db:"acknowledged_alerts"`
		Suppressed   int `db:"suppressed_alerts"`
		Resolved     int `db:"resolved_alerts"`
	}
	var c counts
	err := r.db.GetContext(ctx, &c, query, orgID, shipmentID)
	if err != nil {
		return nil, err
	}

	summary := &spec.TrackingMonitoringSummary{
		ShipmentID:         shipmentID,
		OpenAlerts:         c.Open,
		CriticalAlerts:     c.Critical,
		AcknowledgedAlerts: c.Acknowledged,
		SuppressedAlerts:   c.Suppressed,
		ResolvedAlerts:     c.Resolved,
	}
	return summary, nil
}

// ─── Tracking Refresh Runs (Task 17.7) ────────────────────────────────────────

func (r *repository) CreateTrackingRefreshRun(ctx context.Context, run *spec.TrackingRefreshRunRecord) error {
	query := `
		INSERT INTO shipment_tracking_refresh_runs (
			org_id, shipment_id, provider_name, provider_type, trigger_type,
			status, started_at, completed_at, new_positions, new_events,
			data_freshness, used_fallback, error_message, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, NOW(), NOW()
		)
	`
	res, err := r.db.ExecContext(ctx, query,
		run.OrgID, run.ShipmentID, run.ProviderName, run.ProviderType, run.TriggerType,
		run.Status, run.StartedAt, run.CompletedAt, run.NewPositions, run.NewEvents,
		run.DataFreshness, run.UsedFallback, run.ErrorMessage,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		run.ID = id
	}
	return nil
}

func (r *repository) UpdateTrackingRefreshRun(ctx context.Context, run *spec.TrackingRefreshRunRecord) error {
	query := `
		UPDATE shipment_tracking_refresh_runs
		SET provider_name = ?,
		    provider_type = ?,
		    status = ?,
		    completed_at = ?,
		    new_positions = ?,
		    new_events = ?,
		    data_freshness = ?,
		    used_fallback = ?,
		    error_message = ?,
		    updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		run.ProviderName, run.ProviderType, run.Status, run.CompletedAt,
		run.NewPositions, run.NewEvents, run.DataFreshness, run.UsedFallback,
		run.ErrorMessage, run.ID, run.OrgID,
	)
	return err
}

func (r *repository) GetLatestTrackingRefresh(ctx context.Context, orgID int64, shipmentID int64) (*spec.TrackingRefreshRunRecord, error) {
	query := `
		SELECT id, org_id, shipment_id, provider_name, provider_type, trigger_type,
		       status, started_at, completed_at, new_positions, new_events,
		       data_freshness, used_fallback, error_message, created_at, updated_at
		FROM shipment_tracking_refresh_runs
		WHERE org_id = ? AND shipment_id = ?
		ORDER BY started_at DESC
		LIMIT 1
	`
	var run spec.TrackingRefreshRunRecord
	err := r.db.GetContext(ctx, &run, query, orgID, shipmentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &run, nil
}

func (r *repository) GetTrackingRefreshHistory(ctx context.Context, orgID int64, shipmentID int64, limit int) ([]*spec.TrackingRefreshRunRecord, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	query := `
		SELECT id, org_id, shipment_id, provider_name, provider_type, trigger_type,
		       status, started_at, completed_at, new_positions, new_events,
		       data_freshness, used_fallback, error_message, created_at, updated_at
		FROM shipment_tracking_refresh_runs
		WHERE org_id = ? AND shipment_id = ?
		ORDER BY started_at DESC
		LIMIT ?
	`
	var runs []*spec.TrackingRefreshRunRecord
	err := r.db.SelectContext(ctx, &runs, query, orgID, shipmentID, limit)
	if err != nil {
		return nil, err
	}
	return runs, nil
}

func (r *repository) GetActiveShipmentsForRefresh(ctx context.Context) ([]*spec.Shipment, error) {
	query := `
		SELECT id, org_id, rfq_id, quote_id, booking_id, booking_number,
		       mbl_number, hbl_number, carrier_scac, vessel_name, voyage_number,
		       origin_port, destination_port, container_numbers, status,
		       closure_status, etd, eta, created_at, updated_at
		FROM shipments
		WHERE status NOT IN ('DELIVERED', 'CANCELLED')
		  AND (closure_status IS NULL OR closure_status != 'CLOSED')
		ORDER BY updated_at ASC
	`
	var shipments []*spec.Shipment
	err := r.db.SelectContext(ctx, &shipments, query)
	if err != nil {
		return nil, err
	}
	return shipments, nil
}

// ─── Tracking Analytics & Operational Intelligence (Task 17.8) ────────────────

func (r *repository) GetTrackingAnalyticsOverview(ctx context.Context, orgID int64) (*spec.TrackingAnalyticsOverview, error) {
	overview := &spec.TrackingAnalyticsOverview{
		Insights: make([]spec.TrackingOperationalInsight, 0),
	}

	// 1. Shipment Performance & Schedule Variance
	type shipStats struct {
		TotalShipments    int     `db:"total_shipments"`
		ActiveShipments   int     `db:"active_shipments"`
		OnTimeShipments   int     `db:"on_time_shipments"`
		DelayedShipments  int     `db:"delayed_shipments"`
		EarlyShipments    int     `db:"early_shipments"`
		AvgDelayHours     float64 `db:"avg_delay_hours"`
		AvgVarianceHours  float64 `db:"avg_variance_hours"`
	}
	var ss shipStats
	shipQuery := `
		SELECT 
			COUNT(*) AS total_shipments,
			COALESCE(SUM(CASE WHEN status NOT IN ('DELIVERED', 'CANCELLED') THEN 1 ELSE 0 END), 0) AS active_shipments,
			COALESCE(SUM(CASE WHEN status = 'DELIVERED' OR (eta IS NOT NULL AND eta >= NOW()) THEN 1 ELSE 0 END), 0) AS on_time_shipments,
			COALESCE(SUM(CASE WHEN status NOT IN ('DELIVERED', 'CANCELLED') AND eta IS NOT NULL AND eta < NOW() THEN 1 ELSE 0 END), 0) AS delayed_shipments,
			COALESCE(SUM(CASE WHEN status NOT IN ('DELIVERED', 'CANCELLED') AND eta IS NOT NULL AND TIMESTAMPDIFF(HOUR, NOW(), eta) > 48 THEN 1 ELSE 0 END), 0) AS early_shipments,
			COALESCE(AVG(CASE WHEN eta IS NOT NULL AND eta < NOW() AND status NOT IN ('DELIVERED', 'CANCELLED') THEN TIMESTAMPDIFF(HOUR, eta, NOW()) ELSE 0 END), 0) AS avg_delay_hours,
			COALESCE(AVG(CASE WHEN eta IS NOT NULL AND etd IS NOT NULL THEN ABS(TIMESTAMPDIFF(HOUR, etd, eta)) / 24.0 ELSE 0 END), 0) AS avg_variance_hours
		FROM shipments
		WHERE org_id = ?
	`
	_ = r.db.GetContext(ctx, &ss, shipQuery, orgID)

	overview.TotalTrackedShipments = ss.TotalShipments
	overview.ActiveShipments = ss.ActiveShipments
	overview.DelayedShipments = ss.DelayedShipments
	overview.EarlyShipments = ss.EarlyShipments
	overview.OnScheduleShipments = ss.OnTimeShipments - ss.DelayedShipments
	if overview.OnScheduleShipments < 0 {
		overview.OnScheduleShipments = 0
	}
	overview.AverageDelayHours = ss.AvgDelayHours
	overview.AverageEtaVarianceHours = ss.AvgVarianceHours

	if ss.TotalShipments > 0 {
		overview.OnTimePercentage = float64(ss.TotalShipments-ss.DelayedShipments) / float64(ss.TotalShipments) * 100.0
		if overview.OnTimePercentage < 0 {
			overview.OnTimePercentage = 0
		}
	} else {
		overview.OnTimePercentage = 100.0
	}

	// 2. Alert Statistics
	type alertStats struct {
		TotalOpenAlerts    int `db:"total_open_alerts"`
		OpenCriticalAlerts int `db:"open_critical_alerts"`
	}
	var as alertStats
	alertQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN status = 'OPEN' THEN 1 ELSE 0 END), 0) AS total_open_alerts,
			COALESCE(SUM(CASE WHEN status = 'OPEN' AND severity IN ('CRITICAL', 'HIGH') THEN 1 ELSE 0 END), 0) AS open_critical_alerts
		FROM shipment_tracking_alerts
		WHERE org_id = ?
	`
	_ = r.db.GetContext(ctx, &as, alertQuery, orgID)
	overview.TotalOpenAlerts = as.TotalOpenAlerts
	overview.OpenCriticalAlerts = as.OpenCriticalAlerts

	// 3. Data Freshness Distribution
	type freshStats struct {
		Live   int `db:"freshness_live"`
		Recent int `db:"freshness_recent"`
		Stale  int `db:"freshness_stale"`
	}
	var fs freshStats
	freshQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN p.recorded_at >= NOW() - INTERVAL 6 HOUR THEN 1 ELSE 0 END), 0) AS freshness_live,
			COALESCE(SUM(CASE WHEN p.recorded_at < NOW() - INTERVAL 6 HOUR AND p.recorded_at >= NOW() - INTERVAL 24 HOUR THEN 1 ELSE 0 END), 0) AS freshness_recent,
			COALESCE(SUM(CASE WHEN p.recorded_at < NOW() - INTERVAL 24 HOUR THEN 1 ELSE 0 END), 0) AS freshness_stale
		FROM shipments s
		JOIN (
			SELECT shipment_id, MAX(recorded_at) AS recorded_at
			FROM shipment_tracking_positions
			WHERE org_id = ?
			GROUP BY shipment_id
		) p ON s.id = p.shipment_id
		WHERE s.org_id = ? AND s.status NOT IN ('DELIVERED', 'CANCELLED')
	`
	_ = r.db.GetContext(ctx, &fs, freshQuery, orgID, orgID)
	overview.DataFreshnessLive = fs.Live
	overview.DataFreshnessRecent = fs.Recent
	overview.DataFreshnessStale = fs.Stale

	trackedActive := fs.Live + fs.Recent + fs.Stale
	if overview.ActiveShipments > trackedActive {
		overview.DataFreshnessUnavailable = overview.ActiveShipments - trackedActive
	} else {
		overview.DataFreshnessUnavailable = 0
	}

	// 4. Refresh Success / Reliability Analytics (last 30 days)
	type refreshStats struct {
		TotalRefreshes      int `db:"total_refreshes"`
		SuccessfulRefreshes int `db:"successful_refreshes"`
		FailedRefreshes     int `db:"failed_refreshes"`
	}
	var rs refreshStats
	refreshQuery := `
		SELECT 
			COUNT(*) AS total_refreshes,
			COALESCE(SUM(CASE WHEN status IN ('SUCCESS', 'PARTIAL') THEN 1 ELSE 0 END), 0) AS successful_refreshes,
			COALESCE(SUM(CASE WHEN status = 'FAILED' THEN 1 ELSE 0 END), 0) AS failed_refreshes
		FROM shipment_tracking_refresh_runs
		WHERE org_id = ? AND started_at >= NOW() - INTERVAL 30 DAY
	`
	_ = r.db.GetContext(ctx, &rs, refreshQuery, orgID)
	overview.TotalRefreshes30d = rs.TotalRefreshes
	overview.FailedRefreshes30d = rs.FailedRefreshes
	if rs.TotalRefreshes > 0 {
		overview.RefreshSuccessRate = float64(rs.SuccessfulRefreshes) / float64(rs.TotalRefreshes) * 100.0
	} else {
		overview.RefreshSuccessRate = 100.0
	}

	return overview, nil
}

func (r *repository) GetTrackingAnalyticsTrends(ctx context.Context, orgID int64, days int) ([]spec.TrackingTrendDataPoint, error) {
	if days <= 0 || days > 90 {
		days = 14
	}

	query := `
		SELECT 
			DATE(s.updated_at) AS d_date,
			COUNT(s.id) AS total_shipments,
			COALESCE(SUM(CASE WHEN s.status = 'DELIVERED' OR (s.eta IS NOT NULL AND s.eta >= s.updated_at) THEN 1 ELSE 0 END), 0) AS on_time_count,
			COALESCE(SUM(CASE WHEN s.status NOT IN ('DELIVERED', 'CANCELLED') AND s.eta IS NOT NULL AND s.eta < s.updated_at THEN 1 ELSE 0 END), 0) AS delayed_count
		FROM shipments s
		WHERE s.org_id = ? AND s.updated_at >= NOW() - INTERVAL ? DAY
		GROUP BY DATE(s.updated_at)
		ORDER BY d_date ASC
	`
	type row struct {
		Date         string `db:"d_date"`
		Total        int    `db:"total_shipments"`
		OnTimeCount  int    `db:"on_time_count"`
		DelayedCount int    `db:"delayed_count"`
	}
	var rows []row
	_ = r.db.SelectContext(ctx, &rows, query, orgID, days)

	// Fetch alerts by date
	type alertRow struct {
		Date  string `db:"d_date"`
		Count int    `db:"alert_count"`
	}
	var alertRows []alertRow
	alertQuery := `
		SELECT DATE(created_at) AS d_date, COUNT(*) AS alert_count
		FROM shipment_tracking_alerts
		WHERE org_id = ? AND created_at >= NOW() - INTERVAL ? DAY
		GROUP BY DATE(created_at)
	`
	_ = r.db.SelectContext(ctx, &alertRows, alertQuery, orgID, days)
	alertMap := make(map[string]int)
	for _, ar := range alertRows {
		alertMap[ar.Date] = ar.Count
	}

	// Fetch refresh success by date
	type refRow struct {
		Date     string `db:"d_date"`
		Total    int    `db:"total_refreshes"`
		Success  int    `db:"successful_refreshes"`
	}
	var refRows []refRow
	refQuery := `
		SELECT 
			DATE(started_at) AS d_date,
			COUNT(*) AS total_refreshes,
			COALESCE(SUM(CASE WHEN status IN ('SUCCESS', 'PARTIAL') THEN 1 ELSE 0 END), 0) AS successful_refreshes
		FROM shipment_tracking_refresh_runs
		WHERE org_id = ? AND started_at >= NOW() - INTERVAL ? DAY
		GROUP BY DATE(started_at)
	`
	_ = r.db.SelectContext(ctx, &refRows, refQuery, orgID, days)
	refMap := make(map[string]float64)
	for _, rr := range refRows {
		if rr.Total > 0 {
			refMap[rr.Date] = float64(rr.Success) / float64(rr.Total) * 100.0
		} else {
			refMap[rr.Date] = 100.0
		}
	}

	// Build contiguous daily trend series
	points := make([]spec.TrackingTrendDataPoint, 0, days)
	rowMap := make(map[string]row)
	for _, r := range rows {
		rowMap[r.Date] = r
	}

	now := time.Now()
	for i := days - 1; i >= 0; i-- {
		t := now.AddDate(0, 0, -i)
		dateStr := t.Format("2006-01-02")

		pt := spec.TrackingTrendDataPoint{
			Date:               dateStr,
			OnTimeRate:         100.0,
			DelayedCount:       0,
			AlertCount:         alertMap[dateStr],
			RefreshSuccessRate: 100.0,
			TotalShipments:     0,
		}

		if r, ok := rowMap[dateStr]; ok {
			pt.TotalShipments = r.Total
			pt.DelayedCount = r.DelayedCount
			if r.Total > 0 {
				pt.OnTimeRate = float64(r.OnTimeCount) / float64(r.Total) * 100.0
			}
		}

		if rate, ok := refMap[dateStr]; ok {
			pt.RefreshSuccessRate = rate
		}

		points = append(points, pt)
	}

	return points, nil
}

func (r *repository) GetCarrierTrackingPerformance(ctx context.Context, orgID int64) ([]spec.CarrierTrackingPerformance, error) {
	query := `
		SELECT 
			COALESCE(NULLIF(TRIM(s.carrier_scac), ''), 'UNKNOWN') AS carrier_scac,
			COALESCE(NULLIF(TRIM(s.carrier_scac), ''), 'Unknown Carrier') AS carrier_name,
			COUNT(s.id) AS shipments_tracked,
			COALESCE(SUM(CASE WHEN s.status = 'DELIVERED' OR (s.eta IS NOT NULL AND s.eta >= NOW()) THEN 1 ELSE 0 END), 0) AS on_time_count,
			COALESCE(SUM(CASE WHEN s.status NOT IN ('DELIVERED', 'CANCELLED') AND s.eta IS NOT NULL AND s.eta < NOW() THEN 1 ELSE 0 END), 0) AS delayed_count,
			COALESCE(AVG(CASE WHEN s.eta IS NOT NULL AND s.eta < NOW() AND s.status NOT IN ('DELIVERED', 'CANCELLED') THEN TIMESTAMPDIFF(HOUR, s.eta, NOW()) ELSE 0 END), 0) AS average_delay_hours
		FROM shipments s
		WHERE s.org_id = ?
		GROUP BY s.carrier_scac
		ORDER BY shipments_tracked DESC
	`
	type carrierRow struct {
		CarrierSCAC       string  `db:"carrier_scac"`
		CarrierName       string  `db:"carrier_name"`
		ShipmentsTracked  int     `db:"shipments_tracked"`
		OnTimeCount       int     `db:"on_time_count"`
		DelayedCount      int     `db:"delayed_count"`
		AverageDelayHours float64 `db:"average_delay_hours"`
	}
	var rows []carrierRow
	err := r.db.SelectContext(ctx, &rows, query, orgID)
	if err != nil {
		return nil, err
	}

	// Fetch alert counts by carrier
	type carrierAlertRow struct {
		CarrierSCAC   string `db:"carrier_scac"`
		AlertCount    int    `db:"alert_count"`
		CriticalCount int    `db:"critical_count"`
	}
	var alertRows []carrierAlertRow
	alertQuery := `
		SELECT 
			COALESCE(NULLIF(TRIM(s.carrier_scac), ''), 'UNKNOWN') AS carrier_scac,
			COUNT(a.id) AS alert_count,
			COALESCE(SUM(CASE WHEN a.severity IN ('CRITICAL', 'HIGH') THEN 1 ELSE 0 END), 0) AS critical_count
		FROM shipment_tracking_alerts a
		JOIN shipments s ON a.shipment_id = s.id AND a.org_id = s.org_id
		WHERE a.org_id = ? AND a.status = 'OPEN'
		GROUP BY s.carrier_scac
	`
	_ = r.db.SelectContext(ctx, &alertRows, alertQuery, orgID)
	alertMap := make(map[string]carrierAlertRow)
	for _, ar := range alertRows {
		alertMap[ar.CarrierSCAC] = ar
	}

	results := make([]spec.CarrierTrackingPerformance, 0, len(rows))
	for _, row := range rows {
		item := spec.CarrierTrackingPerformance{
			CarrierSCAC:       row.CarrierSCAC,
			CarrierName:       getCarrierDisplayName(row.CarrierSCAC),
			ShipmentsTracked:  row.ShipmentsTracked,
			OnTimeCount:       row.OnTimeCount,
			DelayedCount:      row.DelayedCount,
			AverageDelayHours: row.AverageDelayHours,
		}

		if row.ShipmentsTracked > 0 {
			item.OnTimeRate = float64(row.OnTimeCount) / float64(row.ShipmentsTracked) * 100.0
			if item.OnTimeRate > 100.0 {
				item.OnTimeRate = 100.0
			}
		} else {
			item.OnTimeRate = 100.0
		}

		if ar, ok := alertMap[row.CarrierSCAC]; ok {
			item.AlertCount = ar.AlertCount
			item.CriticalAlertCount = ar.CriticalCount
		}

		// Calculate deterministic Reliability Score (0 - 100)
		score := item.OnTimeRate
		if item.CriticalAlertCount > 0 {
			score -= float64(item.CriticalAlertCount * 8)
		}
		if item.AverageDelayHours > 24 {
			score -= 10.0
		}
		if score < 0 {
			score = 0
		}
		if score > 100 {
			score = 100
		}
		item.ReliabilityScore = score

		if score >= 90 {
			item.ReliabilityTier = "EXCELLENT"
		} else if score >= 75 {
			item.ReliabilityTier = "GOOD"
		} else if score >= 60 {
			item.ReliabilityTier = "FAIR"
		} else {
			item.ReliabilityTier = "AT_RISK"
		}

		results = append(results, item)
	}

	return results, nil
}

func getCarrierDisplayName(scac string) string {
	switch strings.ToUpper(strings.TrimSpace(scac)) {
	case "MAEU":
		return "Maersk Line"
	case "MSCU":
		return "MSC (Mediterranean Shipping Co.)"
	case "CMDU":
		return "CMA CGM"
	case "HLCU":
		return "Hapag-Lloyd"
	case "ONEY":
		return "Ocean Network Express (ONE)"
	case "COSU":
		return "COSCO Shipping"
	case "EMCU":
		return "Evergreen Marine"
	case "YMLU":
		return "Yang Ming Line"
	case "ZIMU":
		return "ZIM Integrated Shipping"
	case "HMMU":
		return "HMM (Hyundai Merchant Marine)"
	default:
		if scac == "UNKNOWN" || scac == "" {
			return "Direct / Unassigned Carrier"
		}
		return scac
	}
}

func (r *repository) GetRouteTrackingPerformance(ctx context.Context, orgID int64) ([]spec.RouteTrackingPerformance, error) {
	query := `
		SELECT 
			COALESCE(NULLIF(TRIM(s.origin_port), ''), 'Origin') AS origin_port,
			COALESCE(NULLIF(TRIM(s.destination_port), ''), 'Destination') AS destination_port,
			COUNT(s.id) AS shipments_count,
			COALESCE(SUM(CASE WHEN s.status = 'DELIVERED' OR (s.eta IS NOT NULL AND s.eta >= NOW()) THEN 1 ELSE 0 END), 0) AS on_time_count,
			COALESCE(AVG(CASE WHEN s.etd IS NOT NULL AND s.eta IS NOT NULL THEN ABS(TIMESTAMPDIFF(HOUR, s.etd, s.eta)) ELSE 240 END), 240) AS avg_transit_hours,
			COALESCE(AVG(CASE WHEN s.eta IS NOT NULL AND s.eta < NOW() AND s.status NOT IN ('DELIVERED', 'CANCELLED') THEN TIMESTAMPDIFF(HOUR, s.eta, NOW()) ELSE 0 END), 0) AS avg_variance_hours
		FROM shipments s
		WHERE s.org_id = ?
		GROUP BY s.origin_port, s.destination_port
		ORDER BY shipments_count DESC
		LIMIT 20
	`
	type routeRow struct {
		OriginPort       string  `db:"origin_port"`
		DestinationPort  string  `db:"destination_port"`
		ShipmentsCount   int     `db:"shipments_count"`
		OnTimeCount      int     `db:"on_time_count"`
		AvgTransitHours  float64 `db:"avg_transit_hours"`
		AvgVarianceHours float64 `db:"avg_variance_hours"`
	}
	var rows []routeRow
	err := r.db.SelectContext(ctx, &rows, query, orgID)
	if err != nil {
		return nil, err
	}

	results := make([]spec.RouteTrackingPerformance, 0, len(rows))
	for _, row := range rows {
		item := spec.RouteTrackingPerformance{
			RouteKey:                fmt.Sprintf("%s → %s", row.OriginPort, row.DestinationPort),
			OriginPort:              row.OriginPort,
			DestinationPort:         row.DestinationPort,
			ShipmentsCount:          row.ShipmentsCount,
			AvgTransitHours:         row.AvgTransitHours,
			AvgTransitVarianceHours: row.AvgVarianceHours,
		}

		if row.ShipmentsCount > 0 {
			item.OnTimeRate = float64(row.OnTimeCount) / float64(row.ShipmentsCount) * 100.0
		} else {
			item.OnTimeRate = 100.0
		}

		if item.OnTimeRate >= 85 && item.AvgTransitVarianceHours < 12 {
			item.RiskLevel = "LOW"
		} else if item.OnTimeRate >= 70 && item.AvgTransitVarianceHours < 36 {
			item.RiskLevel = "MODERATE"
		} else {
			item.RiskLevel = "HIGH"
		}

		results = append(results, item)
	}

	return results, nil
}

