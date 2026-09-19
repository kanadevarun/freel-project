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
	GetApprovalByReference(ctx context.Context, orgID int64, ref string) (*ApprovalRequest, error)
	GetPendingApprovalByActionAndThread(ctx context.Context, orgID int64, actionName string, threadID string) (*ApprovalRequest, error)
	CreateApproval(ctx context.Context, req *ApprovalRequest) error
	UpdateApprovalStatus(ctx context.Context, orgID int64, id int64, status string, actorName string, notes string, reason string) (*ApprovalRequest, error)
	CancelApproval(ctx context.Context, orgID int64, id int64, actorName string, notes string) (*ApprovalRequest, error)
	ReturnForChanges(ctx context.Context, orgID int64, id int64, actorName string, reason string, notes string) (*ApprovalRequest, error)
	ExpireStaleApprovals(ctx context.Context, orgID int64) (int64, error)
	GetApprovalStats(ctx context.Context, orgID int64) (*ApprovalStats, error)
	ResolveUserName(ctx context.Context, userID int64) (string, error)

	// Decision History & Execution Tracking
	RecordDecision(ctx context.Context, orgID int64, approvalID int64, actionName *string, decision string, actorID *int64, actorName string, reason *string, notes *string, correlationID *string) error
	GetDecisionHistory(ctx context.Context, orgID int64, approvalID int64) ([]*ApprovalDecision, error)
	UpdateExecutionState(ctx context.Context, orgID int64, id int64, status string, result *string, errStr *string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

const selectApprovalColumns = `
	id, org_id, request_code, title, category, type, status, priority,
	related_entity_type, related_entity_id, related_ref, customer_name,
	customer_id, shipment_id, document_id, booking_id,
	requested_by_id, requested_by_name, department, avatar,
	due_date, due_text, assigned_to, approved_by, approved_at,
	rejected_by, rejected_at, rejection_reason, comments, description,
	created_at, updated_at,
	actor_type, source, action_name, risk_level, required_permission,
	ai_task_id, thread_id, checkpoint_id, proposed_payload,
	approval_reference, correlation_id, idempotency_key,
	expires_at, cancelled_by, cancelled_at,
	execution_status, execution_result, execution_error, execution_retries,
	source_module, source_record_type, source_record_id, source_record_snapshot,
	evidence, impact_summary, is_reversible, external_communication,
	required_approval_level, returned_by, returned_at, returned_reason
`

func (r *repository) GetApprovalsByOrg(ctx context.Context, orgID int64) ([]*ApprovalRequest, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM approval_requests
		WHERE org_id = ?
		ORDER BY created_at DESC
	`, selectApprovalColumns)
	var list []*ApprovalRequest
	err := r.db.SelectContext(ctx, &list, query, orgID)
	return list, err
}

func (r *repository) GetApprovalByID(ctx context.Context, orgID int64, id int64) (*ApprovalRequest, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM approval_requests
		WHERE org_id = ? AND id = ?
	`, selectApprovalColumns)
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
	query := fmt.Sprintf(`
		SELECT %s
		FROM approval_requests
		WHERE org_id = ? AND request_code = ?
	`, selectApprovalColumns)
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

func (r *repository) GetApprovalByReference(ctx context.Context, orgID int64, ref string) (*ApprovalRequest, error) {
	if ref == "" {
		return nil, errors.New("approval reference is empty")
	}
	query := fmt.Sprintf(`
		SELECT %s
		FROM approval_requests
		WHERE org_id = ? AND approval_reference = ?
		LIMIT 1
	`, selectApprovalColumns)
	var req ApprovalRequest
	err := r.db.GetContext(ctx, &req, query, orgID, ref)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

func (r *repository) GetPendingApprovalByActionAndThread(ctx context.Context, orgID int64, actionName string, threadID string) (*ApprovalRequest, error) {
	if actionName == "" || threadID == "" {
		return nil, nil
	}
	query := fmt.Sprintf(`
		SELECT %s
		FROM approval_requests
		WHERE org_id = ? AND action_name = ? AND thread_id = ? 
		  AND (status = 'Pending' OR status = 'PENDING_APPROVAL' OR status = 'In Review' OR status = 'Returned for Changes')
		ORDER BY id DESC
		LIMIT 1
	`, selectApprovalColumns)
	var req ApprovalRequest
	err := r.db.GetContext(ctx, &req, query, orgID, actionName, threadID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

func (r *repository) CreateApproval(ctx context.Context, req *ApprovalRequest) error {
	actorType := req.ActorType
	if actorType == "" {
		actorType = "USER"
	}
	riskLevel := req.RiskLevel
	if riskLevel == "" {
		riskLevel = "MEDIUM"
	}
	execStatus := req.ExecutionStatus
	if execStatus == "" {
		execStatus = "NOT_STARTED"
	}
	reqLevel := req.RequiredApprovalLevel
	if reqLevel == "" {
		reqLevel = "MANAGER"
	}

	query := `
		INSERT INTO approval_requests (
			org_id, request_code, title, category, type, status, priority,
			related_entity_type, related_entity_id, related_ref, customer_name,
			customer_id, shipment_id, document_id, booking_id,
			requested_by_id, requested_by_name, department, avatar,
			due_date, due_text, description,
			actor_type, source, action_name, risk_level, required_permission,
			ai_task_id, thread_id, checkpoint_id, proposed_payload,
			approval_reference, correlation_id, idempotency_key, expires_at,
			execution_status, execution_result, execution_error, execution_retries,
			source_module, source_record_type, source_record_id, source_record_snapshot,
			evidence, impact_summary, is_reversible, external_communication,
			required_approval_level,
			created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?,
			NOW(), NOW()
		)
	`
	res, err := r.db.ExecContext(ctx, query,
		req.OrgID, req.RequestCode, req.Title, req.Category, req.Type, req.Status, req.Priority,
		req.RelatedEntityType, req.RelatedEntityID, req.RelatedRef, req.CustomerName,
		req.CustomerID, req.ShipmentID, req.DocumentID, req.BookingID,
		req.RequestedByID, req.RequestedByName, req.Department, req.Avatar,
		req.DueDate, req.DueText, req.Description,
		actorType, req.Source, req.ActionName, riskLevel, req.RequiredPermission,
		req.AITaskID, req.ThreadID, req.CheckpointID, req.ProposedPayload,
		req.ApprovalReference, req.CorrelationID, req.IdempotencyKey, req.ExpiresAt,
		execStatus, req.ExecutionResult, req.ExecutionError, req.ExecutionRetries,
		req.SourceModule, req.SourceRecordType, req.SourceRecordID, req.SourceRecordSnapshot,
		req.Evidence, req.ImpactSummary, req.IsReversible, req.ExternalCommunication,
		reqLevel,
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

	if status == "Approved" || status == "Completed" {
		updateQuery = `
			UPDATE approval_requests
			SET status = ?, due_text = ?, approved_by = ?, approved_at = ?, comments = ?, updated_at = NOW()
			WHERE org_id = ? AND id = ?
		`
		args = []interface{}{status, status, actorName, now, notes, orgID, id}
	} else if status == "Rejected" {
		updateQuery = `
			UPDATE approval_requests
			SET status = 'Rejected', due_text = 'Rejected', rejected_by = ?, rejected_at = ?, rejection_reason = ?, comments = ?, updated_at = NOW()
			WHERE org_id = ? AND id = ?
		`
		args = []interface{}{actorName, now, reason, notes, orgID, id}
	} else if status == "Returned for Changes" {
		updateQuery = `
			UPDATE approval_requests
			SET status = 'Returned for Changes', due_text = 'Action Required', returned_by = ?, returned_at = ?, returned_reason = ?, comments = ?, updated_at = NOW()
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

func (r *repository) ReturnForChanges(ctx context.Context, orgID int64, id int64, actorName string, reason string, notes string) (*ApprovalRequest, error) {
	now := time.Now()
	query := `
		UPDATE approval_requests
		SET status = 'Returned for Changes', due_text = 'Action Required', returned_by = ?, returned_at = ?, returned_reason = ?, comments = ?, updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	res, err := r.db.ExecContext(ctx, query, actorName, now, reason, notes, orgID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to return approval request for changes: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return nil, errors.New("approval request not found or update failed")
	}
	return r.GetApprovalByID(ctx, orgID, id)
}

func (r *repository) CancelApproval(ctx context.Context, orgID int64, id int64, actorName string, notes string) (*ApprovalRequest, error) {
	now := time.Now()
	query := `
		UPDATE approval_requests
		SET status = 'Cancelled', due_text = 'Cancelled', cancelled_by = ?, cancelled_at = ?, comments = ?, updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	res, err := r.db.ExecContext(ctx, query, actorName, now, notes, orgID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel approval request: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return nil, errors.New("approval request not found or cancel failed")
	}
	return r.GetApprovalByID(ctx, orgID, id)
}

func (r *repository) ExpireStaleApprovals(ctx context.Context, orgID int64) (int64, error) {
	query := `
		UPDATE approval_requests
		SET status = 'Expired', due_text = 'Expired', updated_at = NOW()
		WHERE org_id = ? AND (status = 'Pending' OR status = 'PENDING_APPROVAL' OR status = 'In Review' OR status = 'Returned for Changes')
		  AND expires_at IS NOT NULL AND expires_at < NOW()
	`
	res, err := r.db.ExecContext(ctx, query, orgID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *repository) GetApprovalStats(ctx context.Context, orgID int64) (*ApprovalStats, error) {
	stats := &ApprovalStats{
		PendingTrend:  "↑ 3 from last 7 days",
		ApprovedTrend: "↑ 8 from last 7 days",
		RejectedTrend: "↓ 2 from last 7 days",
	}

	_ = r.db.GetContext(ctx, &stats.Pending, "SELECT COUNT(*) FROM approval_requests WHERE org_id = ? AND (status = 'Pending' OR status = 'PENDING' OR status = 'PENDING_APPROVAL' OR status = 'In Review')", orgID)
	_ = r.db.GetContext(ctx, &stats.Approved, "SELECT COUNT(*) FROM approval_requests WHERE org_id = ? AND (status = 'Approved' OR status = 'APPROVED' OR status = 'Completed')", orgID)
	_ = r.db.GetContext(ctx, &stats.Rejected, "SELECT COUNT(*) FROM approval_requests WHERE org_id = ? AND (status = 'Rejected' OR status = 'REJECTED')", orgID)
	_ = r.db.GetContext(ctx, &stats.Returned, "SELECT COUNT(*) FROM approval_requests WHERE org_id = ? AND (status = 'Returned for Changes' OR status = 'RETURNED_FOR_CHANGES' OR status = 'Returned')", orgID)
	_ = r.db.GetContext(ctx, &stats.Overdue, "SELECT COUNT(*) FROM approval_requests WHERE org_id = ? AND (status = 'Overdue' OR status = 'OVERDUE' OR status = 'Expired' OR (due_date IS NOT NULL AND due_date < NOW() AND (status = 'Pending' OR status = 'PENDING_APPROVAL')))", orgID)

	return stats, nil
}

func (r *repository) RecordDecision(ctx context.Context, orgID int64, approvalID int64, actionName *string, decision string, actorID *int64, actorName string, reason *string, notes *string, correlationID *string) error {
	query := `
		INSERT INTO approval_decisions (
			org_id, approval_id, action_name, decision, actor_id, actor_name,
			reason, notes, correlation_id, created_at
		) VALUES (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, NOW()
		)
	`
	_, err := r.db.ExecContext(ctx, query, orgID, approvalID, actionName, decision, actorID, actorName, reason, notes, correlationID)
	return err
}

func (r *repository) GetDecisionHistory(ctx context.Context, orgID int64, approvalID int64) ([]*ApprovalDecision, error) {
	query := `
		SELECT id, org_id, approval_id, action_name, decision, actor_id, actor_name, reason, notes, correlation_id, created_at
		FROM approval_decisions
		WHERE org_id = ? AND approval_id = ?
		ORDER BY created_at ASC
	`
	var history []*ApprovalDecision
	err := r.db.SelectContext(ctx, &history, query, orgID, approvalID)
	if err != nil {
		return nil, err
	}
	return history, nil
}

func (r *repository) UpdateExecutionState(ctx context.Context, orgID int64, id int64, status string, result *string, errStr *string) error {
	var query string
	var args []interface{}

	if status == "FAILED" {
		query = `
			UPDATE approval_requests
			SET execution_status = ?, execution_error = ?, execution_retries = execution_retries + 1, updated_at = NOW()
			WHERE org_id = ? AND id = ?
		`
		args = []interface{}{status, errStr, orgID, id}
	} else {
		query = `
			UPDATE approval_requests
			SET execution_status = ?, execution_result = ?, execution_error = NULL, updated_at = NOW()
			WHERE org_id = ? AND id = ?
		`
		args = []interface{}{status, result, orgID, id}
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *repository) ResolveUserName(ctx context.Context, userID int64) (string, error) {
	if userID <= 0 {
		return "", nil
	}
	var name sql.NullString
	query := `
		SELECT NULLIF(TRIM(CONCAT(COALESCE(first_name, ''), ' ', COALESCE(last_name, ''))), '')
		FROM users
		WHERE id = ?
	`
	err := r.db.GetContext(ctx, &name, query, userID)
	if err == nil && name.Valid && name.String != "" {
		return name.String, nil
	}

	var email string
	if err := r.db.GetContext(ctx, &email, "SELECT email FROM users WHERE id = ?", userID); err == nil && email != "" {
		return email, nil
	}

	return "", nil
}
