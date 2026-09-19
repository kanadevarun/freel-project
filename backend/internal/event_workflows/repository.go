package event_workflows

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	SaveEvent(ctx context.Context, e *EventRecord) (*EventRecord, error)
	GetEventByDedupKey(ctx context.Context, orgID int64, dedupKey string) (*EventRecord, error)
	GetEventByID(ctx context.Context, orgID, eventID int64) (*EventRecord, error)
	ListEvents(ctx context.Context, orgID int64, limit, offset int) ([]*EventRecord, int, error)
	UpdateEventStatus(ctx context.Context, orgID, eventID int64, status string, failureReason *string) error

	SaveWorkflowInstance(ctx context.Context, w *WorkflowInstance) (*WorkflowInstance, error)
	GetWorkflowInstanceByDedupKey(ctx context.Context, orgID int64, dedupKey string) (*WorkflowInstance, error)
	GetWorkflowInstanceByID(ctx context.Context, orgID, id int64) (*WorkflowInstance, error)
	ListWorkflowInstances(ctx context.Context, orgID int64, limit, offset int) ([]*WorkflowInstance, int, error)
	UpdateWorkflowStatus(ctx context.Context, orgID, id int64, status, actionStatus string, approvalID *int64, errorMsg *string) error

	GetCustomerContext(ctx context.Context, orgID, customerID int64) (map[string]interface{}, error)
	GetShipmentContext(ctx context.Context, orgID, shipmentID int64) (map[string]interface{}, error)
	GetInvoiceContext(ctx context.Context, orgID, invoiceID int64) (map[string]interface{}, error)
	GetContractContext(ctx context.Context, orgID, contractID int64) (map[string]interface{}, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) SaveEvent(ctx context.Context, e *EventRecord) (*EventRecord, error) {
	query := `INSERT INTO ai_event_store 
		(org_id, event_type, source_module, source_record_type, source_record_id,
		 event_version, actor_type, actor_id, payload, dedup_key, status,
		 failure_reason, retry_count, correlation_id, causation_id, event_timestamp,
		 processed_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`

	res, err := r.db.ExecContext(ctx, query,
		e.OrgID, e.EventType, e.SourceModule, e.SourceRecordType, e.SourceRecordID,
		e.EventVersion, e.ActorType, e.ActorID, e.Payload, e.DedupKey, e.Status,
		e.FailureReason, e.RetryCount, e.CorrelationID, e.CausationID, e.EventTimestamp,
		e.ProcessedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert ai_event_store: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}
	e.ID = id
	return e, nil
}

func (r *repository) GetEventByDedupKey(ctx context.Context, orgID int64, dedupKey string) (*EventRecord, error) {
	var e EventRecord
	query := `SELECT * FROM ai_event_store WHERE org_id = ? AND dedup_key = ? LIMIT 1`
	err := r.db.GetContext(ctx, &e, query, orgID, dedupKey)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *repository) GetEventByID(ctx context.Context, orgID, eventID int64) (*EventRecord, error) {
	var e EventRecord
	query := `SELECT * FROM ai_event_store WHERE org_id = ? AND id = ? LIMIT 1`
	err := r.db.GetContext(ctx, &e, query, orgID, eventID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *repository) ListEvents(ctx context.Context, orgID int64, limit, offset int) ([]*EventRecord, int, error) {
	if limit <= 0 {
		limit = 50
	}
	var total int
	err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM ai_event_store WHERE org_id = ?`, orgID)
	if err != nil {
		return nil, 0, err
	}

	var events []*EventRecord
	query := `SELECT * FROM ai_event_store WHERE org_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	err = r.db.SelectContext(ctx, &events, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func (r *repository) UpdateEventStatus(ctx context.Context, orgID, eventID int64, status string, failureReason *string) error {
	now := time.Now()
	query := `UPDATE ai_event_store 
		SET status = ?, failure_reason = ?, processed_at = ? 
		WHERE org_id = ? AND id = ?`
	_, err := r.db.ExecContext(ctx, query, status, failureReason, now, orgID, eventID)
	return err
}

func (r *repository) SaveWorkflowInstance(ctx context.Context, w *WorkflowInstance) (*WorkflowInstance, error) {
	query := `INSERT INTO ai_cross_module_workflows 
		(org_id, event_id, workflow_type, status, trigger_event_type,
		 source_module, source_record_type, source_record_id, urgency,
		 ai_summary, ai_analysis, recommendations, action_intent, approval_id,
		 action_name, action_status, draft_id, dedup_key, correlation_id,
		 retry_count, error_message, started_at, completed_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`

	res, err := r.db.ExecContext(ctx, query,
		w.OrgID, w.EventID, w.WorkflowType, w.Status, w.TriggerEventType,
		w.SourceModule, w.SourceRecordType, w.SourceRecordID, w.Urgency,
		w.AISummary, w.AIAnalysis, w.Recommendations, w.ActionIntent, w.ApprovalID,
		w.ActionName, w.ActionStatus, w.DraftID, w.DedupKey, w.CorrelationID,
		w.RetryCount, w.ErrorMessage, w.StartedAt, w.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert ai_cross_module_workflows: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow last insert id: %w", err)
	}
	w.ID = id
	return w, nil
}

func (r *repository) GetWorkflowInstanceByDedupKey(ctx context.Context, orgID int64, dedupKey string) (*WorkflowInstance, error) {
	var w WorkflowInstance
	query := `SELECT * FROM ai_cross_module_workflows WHERE org_id = ? AND dedup_key = ? LIMIT 1`
	err := r.db.GetContext(ctx, &w, query, orgID, dedupKey)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *repository) GetWorkflowInstanceByID(ctx context.Context, orgID, id int64) (*WorkflowInstance, error) {
	var w WorkflowInstance
	query := `SELECT * FROM ai_cross_module_workflows WHERE org_id = ? AND id = ? LIMIT 1`
	err := r.db.GetContext(ctx, &w, query, orgID, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *repository) ListWorkflowInstances(ctx context.Context, orgID int64, limit, offset int) ([]*WorkflowInstance, int, error) {
	if limit <= 0 {
		limit = 50
	}
	var total int
	err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM ai_cross_module_workflows WHERE org_id = ?`, orgID)
	if err != nil {
		return nil, 0, err
	}

	var workflows []*WorkflowInstance
	query := `SELECT * FROM ai_cross_module_workflows WHERE org_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	err = r.db.SelectContext(ctx, &workflows, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return workflows, total, nil
}

func (r *repository) UpdateWorkflowStatus(ctx context.Context, orgID, id int64, status, actionStatus string, approvalID *int64, errorMsg *string) error {
	var completedAt *time.Time
	if status == "COMPLETED" || status == "FAILED" || status == "CANCELLED" {
		now := time.Now()
		completedAt = &now
	}
	query := `UPDATE ai_cross_module_workflows 
		SET status = ?, action_status = ?, approval_id = COALESCE(?, approval_id), error_message = ?, completed_at = COALESCE(?, completed_at), updated_at = NOW() 
		WHERE org_id = ? AND id = ?`
	_, err := r.db.ExecContext(ctx, query, status, actionStatus, approvalID, errorMsg, completedAt, orgID, id)
	return err
}

// ── Cross-Module Aggregators ────────────────────────────────────────────────

func (r *repository) GetCustomerContext(ctx context.Context, orgID, customerID int64) (map[string]interface{}, error) {
	ctxMap := make(map[string]interface{})
	var name string
	var tier sql.NullString
	var creditLimit sql.NullFloat64
	err := r.db.QueryRowContext(ctx, `SELECT company_name, tier, credit_limit FROM customers WHERE org_id = ? AND id = ? LIMIT 1`, orgID, customerID).
		Scan(&name, &tier, &creditLimit)
	if err == nil {
		ctxMap["customer_name"] = name
		if tier.Valid {
			ctxMap["customer_tier"] = tier.String
		}
		if creditLimit.Valid {
			ctxMap["credit_limit"] = creditLimit.Float64
		}
	}

	// Calculate customer total outstanding
	var totalOutstanding sql.NullFloat64
	_ = r.db.QueryRowContext(ctx, `SELECT SUM(total_amount - paid_amount) FROM customer_invoices WHERE org_id = ? AND customer_id = ? AND status != 'PAID'`, orgID, customerID).
		Scan(&totalOutstanding)
	if totalOutstanding.Valid {
		ctxMap["customer_total_outstanding"] = totalOutstanding.Float64
	}

	// Active shipments count
	var activeCount int
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM shipments WHERE org_id = ? AND customer_id = ? AND status NOT IN ('DELIVERED', 'CANCELLED', 'ARCHIVED')`, orgID, customerID).
		Scan(&activeCount)
	ctxMap["active_shipments_count"] = activeCount

	return ctxMap, nil
}

func (r *repository) GetShipmentContext(ctx context.Context, orgID, shipmentID int64) (map[string]interface{}, error) {
	ctxMap := make(map[string]interface{})
	var bookingNum string
	var status string
	var customerID sql.NullInt64
	var origin, dest sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT booking_number, status, customer_id, origin_port, destination_port FROM shipments WHERE org_id = ? AND id = ? LIMIT 1`, orgID, shipmentID).
		Scan(&bookingNum, &status, &customerID, &origin, &dest)
	if err == nil {
		ctxMap["booking_number"] = bookingNum
		ctxMap["status"] = status
		if customerID.Valid {
			ctxMap["customer_id"] = customerID.Int64
			custCtx, _ := r.GetCustomerContext(ctx, orgID, customerID.Int64)
			for k, v := range custCtx {
				ctxMap[k] = v
			}
		}
		if origin.Valid {
			ctxMap["origin_port"] = origin.String
		}
		if dest.Valid {
			ctxMap["destination_port"] = dest.String
		}
	}
	return ctxMap, nil
}

func (r *repository) GetInvoiceContext(ctx context.Context, orgID, invoiceID int64) (map[string]interface{}, error) {
	ctxMap := make(map[string]interface{})
	var invNum string
	var total, paid float64
	var customerID sql.NullInt64
	var dueDate time.Time
	err := r.db.QueryRowContext(ctx, `SELECT invoice_number, total_amount, paid_amount, customer_id, due_date FROM customer_invoices WHERE org_id = ? AND id = ? LIMIT 1`, orgID, invoiceID).
		Scan(&invNum, &total, &paid, &customerID, &dueDate)
	if err == nil {
		ctxMap["invoice_number"] = invNum
		ctxMap["total_amount"] = total
		ctxMap["paid_amount"] = paid
		ctxMap["outstanding_amount"] = total - paid
		ctxMap["due_date"] = dueDate.Format(time.RFC3339)
		if customerID.Valid {
			ctxMap["customer_id"] = customerID.Int64
			custCtx, _ := r.GetCustomerContext(ctx, orgID, customerID.Int64)
			for k, v := range custCtx {
				ctxMap[k] = v
			}
		}
	}
	return ctxMap, nil
}

func (r *repository) GetContractContext(ctx context.Context, orgID, contractID int64) (map[string]interface{}, error) {
	ctxMap := make(map[string]interface{})
	var ref, name string
	var partyName sql.NullString
	var status string
	var expiryDate sql.NullTime
	err := r.db.QueryRowContext(ctx, `SELECT contract_reference, contract_name, party_name, status, expiry_date FROM contracts WHERE org_id = ? AND id = ? LIMIT 1`, orgID, contractID).
		Scan(&ref, &name, &partyName, &status, &expiryDate)
	if err == nil {
		ctxMap["contract_reference"] = ref
		ctxMap["contract_name"] = name
		ctxMap["status"] = status
		if partyName.Valid {
			ctxMap["party_name"] = partyName.String
		}
		if expiryDate.Valid {
			ctxMap["expiry_date"] = expiryDate.Time.Format(time.RFC3339)
		}
	}
	return ctxMap, nil
}
