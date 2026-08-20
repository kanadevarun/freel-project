package shipments

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	CreateShipment(ctx context.Context, s *Shipment) error
	CreateShipmentTx(ctx context.Context, tx *sqlx.Tx, s *Shipment) error
	GetShipmentByID(ctx context.Context, orgID int64, id int64) (*Shipment, error)
	GetShipmentByRFQID(ctx context.Context, rfqID int64) (*Shipment, error)
	GetShipmentByBooking(ctx context.Context, orgID int64, bookingNumber string) (*Shipment, error)
	GetShipmentByContainer(ctx context.Context, orgID int64, containerNumber string) (*Shipment, error)
	GetShipmentByMBL(ctx context.Context, orgID int64, mblNumber string) (*Shipment, error)
	GetShipmentByHBL(ctx context.Context, orgID int64, hblNumber string) (*Shipment, error)
	
	FindShipmentsByBooking(ctx context.Context, orgID int64, bookingNumber string) ([]*Shipment, error)
	FindShipmentsByContainer(ctx context.Context, orgID int64, containerNumber string) ([]*Shipment, error)
	FindShipmentsByMBL(ctx context.Context, orgID int64, mblNumber string) ([]*Shipment, error)
	FindShipmentsByHBL(ctx context.Context, orgID int64, hblNumber string) ([]*Shipment, error)

	ListShipments(ctx context.Context, orgID int64) ([]*Shipment, error)
	UpdateShipment(ctx context.Context, s *Shipment) error

	GetMilestones(ctx context.Context, shipmentID int64) ([]*ShipmentMilestone, error)
	CreateMilestone(ctx context.Context, m *ShipmentMilestone) error
	CreateMilestoneTx(ctx context.Context, tx *sqlx.Tx, m *ShipmentMilestone) error
	UpdateMilestone(ctx context.Context, m *ShipmentMilestone) error

	GetExceptions(ctx context.Context, shipmentID int64) ([]*ShipmentException, error)
	CreateException(ctx context.Context, ex *ShipmentException) error
	ResolveException(ctx context.Context, orgID int64, id int64, resolvedAt time.Time) error

	IsEventProcessed(ctx context.Context, eventID string) (bool, error)
	MarkEventProcessed(ctx context.Context, eventID string, shipmentID int64) error

	FindSCACByCarrierName(ctx context.Context, carrierName string) (string, error)

	InsertCarrierEvent(ctx context.Context, event *CarrierTrackingEvent) (bool, error)
	UpdateCarrierEventStatus(ctx context.Context, eventID string, orgID int64, matchingStatus string, processingStatus string, shipmentID *int64) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateShipment(ctx context.Context, s *Shipment) error {
	containersJSON, err := json.Marshal(s.ContainerNumbers)
	if err != nil {
		containersJSON = []byte("[]")
	}
	query := `
		INSERT INTO shipments (
			org_id, rfq_id, quote_id, carrier_scac, booking_number, mbl_number, hbl_number,
			container_numbers, status, origin_port, destination_port, vessel_name, voyage_number, etd, eta, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
		)
	`
	res, err := r.db.ExecContext(ctx, query,
		s.OrgID, s.RFQID, s.QuoteID, s.CarrierSCAC, s.BookingNumber, s.MBLNumber, s.HBLNumber,
		containersJSON, s.Status, s.OriginPort, s.DestinationPort, s.VesselName, s.VoyageNumber, s.ETD, s.ETA,
	)
	if err != nil {
		return err
	}
	s.ID, err = res.LastInsertId()
	return err
}

func (r *repository) CreateShipmentTx(ctx context.Context, tx *sqlx.Tx, s *Shipment) error {
	containersJSON, err := json.Marshal(s.ContainerNumbers)
	if err != nil {
		containersJSON = []byte("[]")
	}
	query := `
		INSERT INTO shipments (
			org_id, rfq_id, quote_id, carrier_scac, booking_number, mbl_number, hbl_number,
			container_numbers, status, origin_port, destination_port, vessel_name, voyage_number, etd, eta, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
		)
	`
	res, err := tx.ExecContext(ctx, query,
		s.OrgID, s.RFQID, s.QuoteID, s.CarrierSCAC, s.BookingNumber, s.MBLNumber, s.HBLNumber,
		containersJSON, s.Status, s.OriginPort, s.DestinationPort, s.VesselName, s.VoyageNumber, s.ETD, s.ETA,
	)
	if err != nil {
		return err
	}
	s.ID, err = res.LastInsertId()
	return err
}

func (r *repository) GetShipmentByID(ctx context.Context, orgID int64, id int64) (*Shipment, error) {
	query := `SELECT * FROM shipments WHERE id = ? AND org_id = ?`
	var s Shipment
	err := r.db.GetContext(ctx, &s, query, id, orgID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *repository) GetShipmentByRFQID(ctx context.Context, rfqID int64) (*Shipment, error) {
	query := `SELECT * FROM shipments WHERE rfq_id = ?`
	var s Shipment
	err := r.db.GetContext(ctx, &s, query, rfqID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *repository) GetShipmentByBooking(ctx context.Context, orgID int64, bookingNumber string) (*Shipment, error) {
	query := `SELECT * FROM shipments WHERE booking_number = ? AND org_id = ?`
	var s Shipment
	err := r.db.GetContext(ctx, &s, query, bookingNumber, orgID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *repository) GetShipmentByContainer(ctx context.Context, orgID int64, containerNumber string) (*Shipment, error) {
	query := `SELECT * FROM shipments WHERE JSON_CONTAINS(container_numbers, JSON_QUOTE(?)) AND org_id = ?`
	var s Shipment
	err := r.db.GetContext(ctx, &s, query, containerNumber, orgID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *repository) GetShipmentByMBL(ctx context.Context, orgID int64, mblNumber string) (*Shipment, error) {
	query := `SELECT * FROM shipments WHERE mbl_number = ? AND org_id = ?`
	var s Shipment
	err := r.db.GetContext(ctx, &s, query, mblNumber, orgID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *repository) GetShipmentByHBL(ctx context.Context, orgID int64, hblNumber string) (*Shipment, error) {
	query := `SELECT * FROM shipments WHERE hbl_number = ? AND org_id = ?`
	var s Shipment
	err := r.db.GetContext(ctx, &s, query, hblNumber, orgID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *repository) FindShipmentsByBooking(ctx context.Context, orgID int64, bookingNumber string) ([]*Shipment, error) {
	query := `SELECT * FROM shipments WHERE booking_number = ? AND org_id = ?`
	var list []*Shipment
	err := r.db.SelectContext(ctx, &list, query, bookingNumber, orgID)
	return list, err
}

func (r *repository) FindShipmentsByContainer(ctx context.Context, orgID int64, containerNumber string) ([]*Shipment, error) {
	query := `SELECT * FROM shipments WHERE JSON_CONTAINS(container_numbers, JSON_QUOTE(?)) AND org_id = ?`
	var list []*Shipment
	err := r.db.SelectContext(ctx, &list, query, containerNumber, orgID)
	return list, err
}

func (r *repository) FindShipmentsByMBL(ctx context.Context, orgID int64, mblNumber string) ([]*Shipment, error) {
	query := `SELECT * FROM shipments WHERE mbl_number = ? AND org_id = ?`
	var list []*Shipment
	err := r.db.SelectContext(ctx, &list, query, mblNumber, orgID)
	return list, err
}

func (r *repository) FindShipmentsByHBL(ctx context.Context, orgID int64, hblNumber string) ([]*Shipment, error) {
	query := `SELECT * FROM shipments WHERE hbl_number = ? AND org_id = ?`
	var list []*Shipment
	err := r.db.SelectContext(ctx, &list, query, hblNumber, orgID)
	return list, err
}

func (r *repository) ListShipments(ctx context.Context, orgID int64) ([]*Shipment, error) {
	query := `SELECT * FROM shipments WHERE org_id = ? ORDER BY created_at DESC`
	var shipments []*Shipment
	err := r.db.SelectContext(ctx, &shipments, query, orgID)
	return shipments, err
}

func (r *repository) UpdateShipment(ctx context.Context, s *Shipment) error {
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

func (r *repository) GetMilestones(ctx context.Context, shipmentID int64) ([]*ShipmentMilestone, error) {
	query := `SELECT * FROM shipment_milestones WHERE shipment_id = $1 ORDER BY id ASC`
	var milestones []*ShipmentMilestone
	err := r.db.SelectContext(ctx, &milestones, query, shipmentID)
	return milestones, err
}

func (r *repository) CreateMilestone(ctx context.Context, m *ShipmentMilestone) error {
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

func (r *repository) CreateMilestoneTx(ctx context.Context, tx *sqlx.Tx, m *ShipmentMilestone) error {
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

func (r *repository) UpdateMilestone(ctx context.Context, m *ShipmentMilestone) error {
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

func (r *repository) GetExceptions(ctx context.Context, shipmentID int64) ([]*ShipmentException, error) {
	query := `SELECT * FROM shipment_exceptions WHERE shipment_id = ? ORDER BY created_at DESC`
	var exceptions []*ShipmentException
	err := r.db.SelectContext(ctx, &exceptions, query, shipmentID)
	return exceptions, err
}

func (r *repository) CreateException(ctx context.Context, ex *ShipmentException) error {
	query := `
		INSERT INTO shipment_exceptions (
			shipment_id, exception_type, severity, title, description, resolved, ai_summary, source_event_id, created_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, NOW()
		)
	`
	res, err := r.db.ExecContext(ctx, query, ex.ShipmentID, ex.ExceptionType, ex.Severity, ex.Title, ex.Description, ex.Resolved, ex.AISummary, ex.SourceEventID)
	if err != nil {
		return err
	}
	ex.ID, err = res.LastInsertId()
	return err
}

func (r *repository) ResolveException(ctx context.Context, orgID int64, id int64, resolvedAt time.Time) error {
	query := `
		UPDATE shipment_exceptions 
		SET resolved = true, resolved_at = ? 
		WHERE id = ? AND shipment_id IN (SELECT id FROM shipments WHERE org_id = ?)
	`
	_, err := r.db.ExecContext(ctx, query, resolvedAt, id, orgID)
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

func (r *repository) InsertCarrierEvent(ctx context.Context, event *CarrierTrackingEvent) (bool, error) {
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
	query := `SELECT scac FROM carriers WHERE name = ? OR (aliases IS NOT NULL AND JSON_CONTAINS(aliases, JSON_QUOTE(?))) LIMIT 1`
	var scac string
	err := r.db.GetContext(ctx, &scac, query, carrierName, carrierName)
	return scac, err
}
