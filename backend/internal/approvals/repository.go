package approvals

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetApprovalsByOrg(ctx context.Context, orgID int64) ([]*ApprovalRequest, error)
	GetApprovalByID(ctx context.Context, orgID int64, id int64) (*ApprovalRequest, error)
	GetApprovalByCode(ctx context.Context, orgID int64, code string) (*ApprovalRequest, error)
	CreateApproval(ctx context.Context, req *ApprovalRequest) error
	UpdateApprovalStatus(ctx context.Context, orgID int64, id int64, status string, actorName string, notes string, reason string) (*ApprovalRequest, error)
	GetApprovalStats(ctx context.Context, orgID int64) (*ApprovalStats, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetApprovalsByOrg(ctx context.Context, orgID int64) ([]*ApprovalRequest, error) {
	query := `
		SELECT 
			id, org_id, request_code, title, category, type, status, priority,
			related_entity_type, related_entity_id, related_ref, customer_name,
			customer_id, shipment_id, document_id, booking_id,
			requested_by_id, requested_by_name, department, avatar,
			due_date, due_text, assigned_to, approved_by, approved_at,
			rejected_by, rejected_at, rejection_reason, comments, description,
			created_at, updated_at
		FROM approval_requests
		WHERE org_id = ?
		ORDER BY created_at DESC
	`
	var list []*ApprovalRequest
	err := r.db.SelectContext(ctx, &list, query, orgID)
	return list, err
}

func (r *repository) GetApprovalByID(ctx context.Context, orgID int64, id int64) (*ApprovalRequest, error) {
	query := `
		SELECT 
			id, org_id, request_code, title, category, type, status, priority,
			related_entity_type, related_entity_id, related_ref, customer_name,
			customer_id, shipment_id, document_id, booking_id,
			requested_by_id, requested_by_name, department, avatar,
			due_date, due_text, assigned_to, approved_by, approved_at,
			rejected_by, rejected_at, rejection_reason, comments, description,
			created_at, updated_at
		FROM approval_requests
		WHERE org_id = ? AND id = ?
	`
	var req ApprovalRequest
	err := r.db.GetContext(ctx, &req, query, orgID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("approval request not found")
		}
		return nil, err
	}
	return &req, nil
}

func (r *repository) GetApprovalByCode(ctx context.Context, orgID int64, code string) (*ApprovalRequest, error) {
	query := `
		SELECT 
			id, org_id, request_code, title, category, type, status, priority,
			related_entity_type, related_entity_id, related_ref, customer_name,
			customer_id, shipment_id, document_id, booking_id,
			requested_by_id, requested_by_name, department, avatar,
			due_date, due_text, assigned_to, approved_by, approved_at,
			rejected_by, rejected_at, rejection_reason, comments, description,
			created_at, updated_at
		FROM approval_requests
		WHERE org_id = ? AND request_code = ?
	`
	var req ApprovalRequest
	err := r.db.GetContext(ctx, &req, query, orgID, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("approval request not found")
		}
		return nil, err
	}
	return &req, nil
}

func (r *repository) CreateApproval(ctx context.Context, req *ApprovalRequest) error {
	query := `
		INSERT INTO approval_requests (
			org_id, request_code, title, category, type, status, priority,
			related_entity_type, related_entity_id, related_ref, customer_name,
			customer_id, shipment_id, document_id, booking_id,
			requested_by_id, requested_by_name, department, avatar,
			due_date, due_text, description, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, NOW(), NOW()
		)
	`
	res, err := r.db.ExecContext(ctx, query,
		req.OrgID, req.RequestCode, req.Title, req.Category, req.Type, req.Status, req.Priority,
		req.RelatedEntityType, req.RelatedEntityID, req.RelatedRef, req.CustomerName,
		req.CustomerID, req.ShipmentID, req.DocumentID, req.BookingID,
		req.RequestedByID, req.RequestedByName, req.Department, req.Avatar,
		req.DueDate, req.DueText, req.Description,
	)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err == nil {
		req.ID = id
	}
	return nil
}

func (r *repository) UpdateApprovalStatus(ctx context.Context, orgID int64, id int64, status string, actorName string, notes string, reason string) (*ApprovalRequest, error) {
	now := time.Now()
	var updateQuery string
	var args []interface{}

	if status == "Approved" {
		updateQuery = `
			UPDATE approval_requests
			SET status = 'Approved', due_text = 'Approved', approved_by = ?, approved_at = ?, comments = ?, updated_at = NOW()
			WHERE org_id = ? AND id = ?
		`
		args = []interface{}{actorName, now, notes, orgID, id}
	} else if status == "Rejected" {
		updateQuery = `
			UPDATE approval_requests
			SET status = 'Rejected', due_text = 'Rejected', rejected_by = ?, rejected_at = ?, rejection_reason = ?, comments = ?, updated_at = NOW()
			WHERE org_id = ? AND id = ?
		`
		args = []interface{}{actorName, now, reason, notes, orgID, id}
	} else {
		updateQuery = `
			UPDATE approval_requests
			SET status = ?, comments = ?, updated_at = NOW()
			WHERE org_id = ? AND id = ?
		`
		args = []interface{}{status, notes, orgID, id}
	}

	res, err := r.db.ExecContext(ctx, updateQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update approval status: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return nil, errors.New("approval request not found or status update failed")
	}

	return r.GetApprovalByID(ctx, orgID, id)
}

func (r *repository) GetApprovalStats(ctx context.Context, orgID int64) (*ApprovalStats, error) {
	stats := &ApprovalStats{
		PendingTrend:  "↑ 3 from last 7 days",
		ApprovedTrend: "↑ 8 from last 7 days",
		RejectedTrend: "↓ 2 from last 7 days",
	}

	_ = r.db.GetContext(ctx, &stats.Pending, "SELECT COUNT(*) FROM approval_requests WHERE org_id = ? AND (status = 'Pending' OR status = 'PENDING' OR status = 'Overdue')", orgID)
	_ = r.db.GetContext(ctx, &stats.Approved, "SELECT COUNT(*) FROM approval_requests WHERE org_id = ? AND (status = 'Approved' OR status = 'APPROVED')", orgID)
	_ = r.db.GetContext(ctx, &stats.Rejected, "SELECT COUNT(*) FROM approval_requests WHERE org_id = ? AND (status = 'Rejected' OR status = 'REJECTED')", orgID)
	_ = r.db.GetContext(ctx, &stats.Overdue, "SELECT COUNT(*) FROM approval_requests WHERE org_id = ? AND (status = 'Overdue' OR status = 'OVERDUE' OR (due_date IS NOT NULL AND due_date < NOW() AND status = 'Pending'))", orgID)

	return stats, nil
}
