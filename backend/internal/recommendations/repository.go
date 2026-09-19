package recommendations

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// Repository handles database persistence for recommendations
type Repository interface {
	List(ctx context.Context, orgID int64, filter RecommendationFilter) ([]*Recommendation, int64, error)
	GetByID(ctx context.Context, orgID int64, id int64) (*Recommendation, error)
	GetByDedupHash(ctx context.Context, orgID int64, dedupHash string) (*Recommendation, error)
	ListBySource(ctx context.Context, orgID int64, sourceType string, sourceID int64) ([]*Recommendation, error)
	Create(ctx context.Context, rec *Recommendation) (*Recommendation, error)
	UpdateFreshnessAndEvidence(ctx context.Context, id int64, orgID int64, freshness time.Time, evidenceJSON string, confidence string, confidenceScore float64) error
	UpdateStatus(ctx context.Context, id int64, orgID int64, status string, userID *int64, reason *string) error
	Assign(ctx context.Context, id int64, orgID int64, assigneeID int64, assigneeName string) error
	Dismiss(ctx context.Context, id int64, orgID int64, dismissedByID int64, reason string) error
	MarkReviewed(ctx context.Context, id int64, orgID int64, reviewedByID int64) error
	GetStats(ctx context.Context, orgID int64) (*RecommendationStats, error)

	// Phase 2 Task 2.2: Customer Follow-Up Assistant methods
	SaveDraft(ctx context.Context, id int64, orgID int64, subject string, body string, status string) error
	CreateFollowupTask(ctx context.Context, task *CustomerFollowupTask) (*CustomerFollowupTask, error)
	GetFollowupTaskByRecID(ctx context.Context, orgID int64, recID int64) (*CustomerFollowupTask, error)
	ListFollowupTasks(ctx context.Context, orgID int64, customerID *int64, status string) ([]*CustomerFollowupTask, error)
	GetFollowupStats(ctx context.Context, orgID int64) (*FollowupStats, error)

	// Phase 2 Task 2.3: Link Approval
	LinkApproval(ctx context.Context, id int64, orgID int64, approvalID int64) error

	// Phase 2 Task 2.5: Invoice and Collections Assistant methods
	GetInvoiceEvidence(ctx context.Context, orgID int64, invoiceID int64) (*InvoiceEvidencePayload, error)
	GetCustomerCollectionSummary(ctx context.Context, orgID int64, customerID int64) (*CustomerCollectionSummaryPayload, error)

	// Phase 2 Task 2.6: Contract, Document, and Compliance Assistant methods
	GetContractEvidence(ctx context.Context, orgID int64, contractID int64) (*ContractEvidencePayload, error)
	GetDocumentEvidence(ctx context.Context, orgID int64, docID string) (*DocumentEvidencePayload, error)
	GetComplianceEvidence(ctx context.Context, orgID int64, compID int64) (*ComplianceEvidencePayload, error)
	GetContractComplianceSummary(ctx context.Context, orgID int64) (*ContractComplianceSummaryPayload, error)

	// Phase 2 Task 2.7: Workflow Automation methods
	UpdateExecutionLink(ctx context.Context, id int64, orgID int64, automationID int64, executionID int64) error
}

type repository struct {
	db *sqlx.DB
}

// NewRepository creates a new recommendation repository
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

const recSelectCols = `
	id, org_id, source_type, source_id, source_reference, title, description,
	category, priority, risk_level, confidence, confidence_score, evidence,
	recommended_action, action_type, status, assignee_id, assignee_name,
	created_at, updated_at, reviewed_at, reviewed_by_id, completed_at, completed_by_id,
	dismissed_at, dismissed_by_id, dismissed_reason, freshness, expires_at,
	requires_approval, approval_id, correlation_id, created_by, generated_by,
	rule_applied, dedup_hash, metadata,
	customer_id, customer_name, followup_type, suggested_owner_id, suggested_owner_name,
	draft_subject, draft_body, draft_status, draft_generated_at, followup_task_id,
	rfq_id, quotation_id,
	invoice_id,
	shipment_id, milestone_id, exception_id, booking_id,
	contract_id, document_id, compliance_id, carrier_id,
	automation_id, execution_id
`

func (r *repository) List(ctx context.Context, orgID int64, filter RecommendationFilter) ([]*Recommendation, int64, error) {
	where := []string{"org_id = ?"}
	args := []interface{}{orgID}

	if filter.Category != "" {
		where = append(where, "category = ?")
		args = append(args, filter.Category)
	}
	if filter.Priority != "" {
		where = append(where, "priority = ?")
		args = append(args, filter.Priority)
	}
	if filter.RiskLevel != "" {
		where = append(where, "risk_level = ?")
		args = append(args, filter.RiskLevel)
	}
	if filter.Status != "" {
		where = append(where, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.SourceType != "" {
		where = append(where, "source_type = ?")
		args = append(args, filter.SourceType)
	}
	if filter.RequiresApproval != nil {
		where = append(where, "requires_approval = ?")
		args = append(args, *filter.RequiresApproval)
	}
	if filter.AssigneeID != nil {
		where = append(where, "assignee_id = ?")
		args = append(args, *filter.AssigneeID)
	}
	if filter.CustomerID != nil {
		where = append(where, "customer_id = ?")
		args = append(args, *filter.CustomerID)
	}
	if filter.FollowupType != "" {
		where = append(where, "followup_type = ?")
		args = append(args, filter.FollowupType)
	}
	if filter.RFQID != nil {
		where = append(where, "(rfq_id = ? OR (source_type = 'RFQ' AND source_id = ?))")
		args = append(args, *filter.RFQID, *filter.RFQID)
	}
	if filter.QuotationID != nil {
		where = append(where, "(quotation_id = ? OR (source_type = 'QUOTATION' AND source_id = ?))")
		args = append(args, *filter.QuotationID, *filter.QuotationID)
	}
	if filter.MissingInfo != nil && *filter.MissingInfo {
		where = append(where, "rule_applied = 'RFQ_MISSING_INFO'")
	}
	if filter.ExpiringSoon != nil && *filter.ExpiringSoon {
		where = append(where, "(rule_applied = 'QUOTATION_EXPIRING_SOON' OR rule_applied = 'QUOTATION_EXPIRY_RISK')")
	}
	if filter.PricingConcern != nil && *filter.PricingConcern {
		where = append(where, "(category = 'pricing' OR rule_applied = 'QUOTATION_MARGIN_ALERT')")
	}
	if filter.ShipmentID != nil {
		where = append(where, "(shipment_id = ? OR (source_type = 'SHIPMENT' AND source_id = ?))")
		args = append(args, *filter.ShipmentID, *filter.ShipmentID)
	}
	if filter.MilestoneID != nil {
		where = append(where, "milestone_id = ?")
		args = append(args, *filter.MilestoneID)
	}
	if filter.ExceptionID != nil {
		where = append(where, "exception_id = ?")
		args = append(args, *filter.ExceptionID)
	}
	if filter.BookingID != nil {
		where = append(where, "booking_id = ?")
		args = append(args, *filter.BookingID)
	}
	if filter.DelayedMilestone != nil && *filter.DelayedMilestone {
		where = append(where, "(rule_applied = 'SHIPMENT_MILESTONE_DELAYED' OR rule_applied = 'SHIPMENT_OVERDUE_DELIVERY')")
	}
	if filter.ActiveException != nil && *filter.ActiveException {
		where = append(where, "(category = 'exception' OR rule_applied = 'SHIPMENT_EXCEPTION_ALERT' OR rule_applied = 'UNRESOLVED_SHIPMENT_EXCEPTIONS')")
	}
	if filter.MissingOpsInfo != nil && *filter.MissingOpsInfo {
		where = append(where, "rule_applied = 'SHIPMENT_MISSING_OPERATIONAL_INFO'")
	}
	// Phase 2 Task 2.5 Filters
	if filter.InvoiceID != nil {
		where = append(where, "(invoice_id = ? OR (source_type = 'INVOICE' AND source_id = ?))")
		args = append(args, *filter.InvoiceID, *filter.InvoiceID)
	}
	if filter.OverdueOnly != nil && *filter.OverdueOnly {
		where = append(where, "(rule_applied = 'OVERDUE_INVOICE_COLLECTIONS' OR rule_applied = 'HIGH_RISK_OVERDUE_INVOICE')")
	}
	if filter.DataQualityOnly != nil && *filter.DataQualityOnly {
		where = append(where, "rule_applied = 'INVOICE_DATA_QUALITY_ISSUE'")
	}
	if filter.UpcomingOnly != nil && *filter.UpcomingOnly {
		where = append(where, "rule_applied = 'UPCOMING_COLLECTION_PRIORITY'")
	}

	// Phase 2 Task 2.6: Contract, Document, and Compliance Assistant Filters
	if filter.ContractID != nil {
		where = append(where, "(contract_id = ? OR (source_type = 'CONTRACT' AND source_id = ?))")
		args = append(args, *filter.ContractID, *filter.ContractID)
	}
	if filter.DocumentID != nil {
		where = append(where, "(document_id = ? OR (source_type = 'DOCUMENT' AND source_reference = ?))")
		args = append(args, *filter.DocumentID, *filter.DocumentID)
	}
	if filter.ComplianceID != nil {
		where = append(where, "(compliance_id = ? OR (source_type = 'COMPLIANCE' AND source_id = ?))")
		args = append(args, *filter.ComplianceID, *filter.ComplianceID)
	}
	if filter.CarrierID != nil {
		where = append(where, "(carrier_id = ? OR (source_type = 'CARRIER' AND source_id = ?))")
		args = append(args, *filter.CarrierID, *filter.CarrierID)
	}
	if filter.ExpiredOnly != nil && *filter.ExpiredOnly {
		where = append(where, "(rule_applied IN ('CONTRACT_EXPIRED', 'CONTRACT_EXPIRING_SOON', 'DOCUMENT_EXPIRED', 'DOCUMENT_EXPIRING_SOON'))")
	}

	// Phase 2 Task 2.7: Workflow Automation Filters
	if filter.AutomationID != nil {
		where = append(where, "automation_id = ?")
		args = append(args, *filter.AutomationID)
	}
	if filter.ExecutionID != nil {
		where = append(where, "execution_id = ?")
		args = append(args, *filter.ExecutionID)
	}

	if filter.Search != "" {
		where = append(where, "(title LIKE ? OR description LIKE ? OR source_reference LIKE ? OR customer_name LIKE ?)")
		searchPattern := "%" + filter.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	whereClause := strings.Join(where, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ai_recommendations WHERE %s", whereClause)
	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count recommendations: %w", err)
	}

	// Safe sort
	sortCol := "created_at"
	switch strings.ToLower(filter.SortBy) {
	case "priority":
		sortCol = "FIELD(priority, 'critical', 'high', 'medium', 'low')"
	case "risk_level":
		sortCol = "FIELD(risk_level, 'critical', 'high', 'medium', 'low')"
	case "confidence":
		sortCol = "confidence_score"
	case "freshness":
		sortCol = "freshness"
	case "status":
		sortCol = "status"
	case "title":
		sortCol = "title"
	}

	sortOrder := "DESC"
	if strings.ToUpper(filter.SortDir) == "ASC" {
		sortOrder = "ASC"
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_recommendations
		WHERE %s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, recSelectCols, whereClause, sortCol, sortOrder)

	var list []*Recommendation
	args = append(args, limit, offset)
	if err := r.db.SelectContext(ctx, &list, query, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to query recommendations: %w", err)
	}

	for _, rec := range list {
		rec.UnmarshalDetails()
	}

	return list, total, nil
}

func (r *repository) GetByID(ctx context.Context, orgID int64, id int64) (*Recommendation, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_recommendations
		WHERE id = ? AND org_id = ?
		LIMIT 1
	`, recSelectCols)

	var rec Recommendation
	if err := r.db.GetContext(ctx, &rec, query, id, orgID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get recommendation: %w", err)
	}

	rec.UnmarshalDetails()
	return &rec, nil
}

func (r *repository) GetByDedupHash(ctx context.Context, orgID int64, dedupHash string) (*Recommendation, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_recommendations
		WHERE org_id = ? AND dedup_hash = ?
		LIMIT 1
	`, recSelectCols)

	var rec Recommendation
	if err := r.db.GetContext(ctx, &rec, query, orgID, dedupHash); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get recommendation by dedup hash: %w", err)
	}

	rec.UnmarshalDetails()
	return &rec, nil
}

func (r *repository) ListBySource(ctx context.Context, orgID int64, sourceType string, sourceID int64) ([]*Recommendation, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_recommendations
		WHERE org_id = ? AND source_type = ? AND source_id = ? AND status NOT IN ('dismissed', 'completed')
		ORDER BY created_at DESC
	`, recSelectCols)

	var list []*Recommendation
	if err := r.db.SelectContext(ctx, &list, query, orgID, sourceType, sourceID); err != nil {
		return nil, fmt.Errorf("failed to list recommendations by source: %w", err)
	}

	for _, rec := range list {
		rec.UnmarshalDetails()
	}

	return list, nil
}

func (r *repository) Create(ctx context.Context, rec *Recommendation) (*Recommendation, error) {
	query := `
		INSERT INTO ai_recommendations (
			org_id, source_type, source_id, source_reference, title, description,
			category, priority, risk_level, confidence, confidence_score, evidence,
			recommended_action, action_type, status, assignee_id, assignee_name,
			created_at, updated_at, freshness, expires_at, requires_approval,
			approval_id, correlation_id, created_by, generated_by, rule_applied,
			dedup_hash, metadata,
			customer_id, customer_name, followup_type, suggested_owner_id, suggested_owner_name,
			draft_subject, draft_body, draft_status, draft_generated_at, followup_task_id,
			rfq_id, quotation_id,
			invoice_id,
			shipment_id, milestone_id, exception_id, booking_id,
			contract_id, document_id, compliance_id, carrier_id,
			automation_id, execution_id
		) VALUES (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			NOW(), NOW(), NOW(), ?, ?,
			?, ?, ?, ?, ?,
			?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?,
			?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?
		)
	`
	draftStatus := rec.DraftStatus
	if draftStatus == "" {
		draftStatus = DraftStatusNotGenerated
	}

	var rfqID *int64 = rec.RFQID
	if rfqID == nil && rec.SourceType == SourceRFQ && rec.SourceID > 0 {
		id := rec.SourceID
		rfqID = &id
	}
	var quoteID *int64 = rec.QuotationID
	if quoteID == nil && rec.SourceType == SourceQuotation && rec.SourceID > 0 {
		id := rec.SourceID
		quoteID = &id
	}

	var shipID *int64 = rec.ShipmentID
	if shipID == nil && rec.SourceType == SourceShipment && rec.SourceID > 0 {
		id := rec.SourceID
		shipID = &id
	}

	var invID *int64 = rec.InvoiceID
	if invID == nil && rec.SourceType == SourceInvoice && rec.SourceID > 0 {
		id := rec.SourceID
		invID = &id
	}

	var contractID *int64 = rec.ContractID
	if contractID == nil && rec.SourceType == SourceContract && rec.SourceID > 0 {
		id := rec.SourceID
		contractID = &id
	}

	var compID *int64 = rec.ComplianceID
	if compID == nil && rec.SourceType == SourceCompliance && rec.SourceID > 0 {
		id := rec.SourceID
		compID = &id
	}

	res, err := r.db.ExecContext(ctx, query,
		rec.OrgID, rec.SourceType, rec.SourceID, rec.SourceReference, rec.Title, rec.Description,
		rec.Category, rec.Priority, rec.RiskLevel, rec.Confidence, rec.ConfidenceScore, rec.EvidenceJSON,
		rec.RecommendedAction, rec.ActionType, rec.Status, rec.AssigneeID, rec.AssigneeName,
		rec.ExpiresAt, rec.RequiresApproval, rec.ApprovalID, rec.CorrelationID, rec.CreatedBy,
		rec.GeneratedBy, rec.RuleApplied, rec.DedupHash, rec.MetadataJSON,
		rec.CustomerID, rec.CustomerName, rec.FollowupType, rec.SuggestedOwnerID, rec.SuggestedOwnerName,
		rec.DraftSubject, rec.DraftBody, draftStatus, rec.DraftGeneratedAt, rec.FollowupTaskID,
		rfqID, quoteID,
		invID,
		shipID, rec.MilestoneID, rec.ExceptionID, rec.BookingID,
		contractID, rec.DocumentID, compID, rec.CarrierID,
		rec.AutomationID, rec.ExecutionID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create recommendation: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return r.GetByID(ctx, rec.OrgID, id)
}

func (r *repository) UpdateExecutionLink(ctx context.Context, id int64, orgID int64, automationID int64, executionID int64) error {
	query := `
		UPDATE ai_recommendations
		SET automation_id = ?, execution_id = ?, updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, automationID, executionID, id, orgID)
	return err
}

func (r *repository) UpdateFreshnessAndEvidence(ctx context.Context, id int64, orgID int64, freshness time.Time, evidenceJSON string, confidence string, confidenceScore float64) error {
	query := `
		UPDATE ai_recommendations
		SET freshness = ?, evidence = ?, confidence = ?, confidence_score = ?, updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, freshness, evidenceJSON, confidence, confidenceScore, id, orgID)
	return err
}

func (r *repository) UpdateStatus(ctx context.Context, id int64, orgID int64, status string, userID *int64, reason *string) error {
	var query string
	var args []interface{}

	switch status {
	case StatusReviewed:
		query = `
			UPDATE ai_recommendations
			SET status = ?, reviewed_at = NOW(), reviewed_by_id = ?, updated_at = NOW()
			WHERE id = ? AND org_id = ?
		`
		args = []interface{}{status, userID, id, orgID}
	case StatusCompleted:
		query = `
			UPDATE ai_recommendations
			SET status = ?, completed_at = NOW(), completed_by_id = ?, updated_at = NOW()
			WHERE id = ? AND org_id = ?
		`
		args = []interface{}{status, userID, id, orgID}
	case StatusDismissed:
		query = `
			UPDATE ai_recommendations
			SET status = ?, dismissed_at = NOW(), dismissed_by_id = ?, dismissed_reason = ?, updated_at = NOW()
			WHERE id = ? AND org_id = ?
		`
		args = []interface{}{status, userID, reason, id, orgID}
	default:
		query = `
			UPDATE ai_recommendations
			SET status = ?, updated_at = NOW()
			WHERE id = ? AND org_id = ?
		`
		args = []interface{}{status, id, orgID}
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *repository) Assign(ctx context.Context, id int64, orgID int64, assigneeID int64, assigneeName string) error {
	query := `
		UPDATE ai_recommendations
		SET assignee_id = ?, assignee_name = ?, status = CASE WHEN status IN ('new', 'reviewed') THEN 'assigned' ELSE status END, updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, assigneeID, assigneeName, id, orgID)
	return err
}

func (r *repository) Dismiss(ctx context.Context, id int64, orgID int64, dismissedByID int64, reason string) error {
	query := `
		UPDATE ai_recommendations
		SET status = 'dismissed', dismissed_at = NOW(), dismissed_by_id = ?, dismissed_reason = ?, updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, dismissedByID, reason, id, orgID)
	return err
}

func (r *repository) MarkReviewed(ctx context.Context, id int64, orgID int64, reviewedByID int64) error {
	query := `
		UPDATE ai_recommendations
		SET status = 'reviewed', reviewed_at = NOW(), reviewed_by_id = ?, updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, reviewedByID, id, orgID)
	return err
}

func (r *repository) GetStats(ctx context.Context, orgID int64) (*RecommendationStats, error) {
	query := `
		SELECT
			COUNT(CASE WHEN status NOT IN ('completed', 'dismissed', 'failed') THEN 1 END) AS total_active,
			COUNT(CASE WHEN priority = 'critical' AND status NOT IN ('completed', 'dismissed', 'failed') THEN 1 END) AS critical_count,
			COUNT(CASE WHEN priority = 'high' AND status NOT IN ('completed', 'dismissed', 'failed') THEN 1 END) AS high_count,
			COUNT(CASE WHEN priority = 'medium' AND status NOT IN ('completed', 'dismissed', 'failed') THEN 1 END) AS medium_count,
			COUNT(CASE WHEN priority = 'low' AND status NOT IN ('completed', 'dismissed', 'failed') THEN 1 END) AS low_count,
			COUNT(CASE WHEN status = 'new' THEN 1 END) AS requires_review,
			COUNT(CASE WHEN requires_approval = 1 AND status IN ('new', 'reviewed', 'assigned') THEN 1 END) AS requires_approval,
			COUNT(CASE WHEN status = 'assigned' THEN 1 END) AS assigned_count,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) AS completed_count,
			COUNT(CASE WHEN status = 'dismissed' THEN 1 END) AS dismissed_count
		FROM ai_recommendations
		WHERE org_id = ?
	`
	var stats RecommendationStats
	if err := r.db.GetContext(ctx, &stats, query, orgID); err != nil {
		return nil, fmt.Errorf("failed to get recommendation stats: %w", err)
	}
	return &stats, nil
}

func (r *repository) SaveDraft(ctx context.Context, id int64, orgID int64, subject string, body string, status string) error {
	if status == "" {
		status = "SAVED"
	}
	query := `
		UPDATE ai_recommendations
		SET draft_subject = ?, draft_body = ?, draft_status = ?, draft_generated_at = NOW(), updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, subject, body, status, id, orgID)
	return err
}

func (r *repository) CreateFollowupTask(ctx context.Context, task *CustomerFollowupTask) (*CustomerFollowupTask, error) {
	// Idempotency check: if task already exists for this (org_id, recommendation_id), return existing
	existing, err := r.GetFollowupTaskByRecID(ctx, task.OrgID, task.RecommendationID)
	if err == nil && existing != nil {
		return existing, nil
	}

	query := `
		INSERT INTO customer_followup_tasks (
			org_id, recommendation_id, customer_id, customer_name, source_type,
			source_id, source_reference, followup_type, title, reason,
			suggested_action, priority, assignee_id, assignee_name, due_date,
			status, created_at, updated_at, created_by_id, created_by_name,
			correlation_id, notes
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, NOW(), NOW(), ?, ?,
			?, ?
		)
	`
	res, err := r.db.ExecContext(ctx, query,
		task.OrgID, task.RecommendationID, task.CustomerID, task.CustomerName, task.SourceType,
		task.SourceID, task.SourceReference, task.FollowupType, task.Title, task.Reason,
		task.SuggestedAction, task.Priority, task.AssigneeID, task.AssigneeName, task.DueDate,
		task.Status, task.CreatedByID, task.CreatedByName,
		task.CorrelationID, task.Notes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert customer followup task: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}
	task.ID = id

	// Update ai_recommendations with task linkage and transition status to assigned
	updateRecQuery := `
		UPDATE ai_recommendations
		SET followup_task_id = ?,
		    status = CASE WHEN status IN ('new', 'reviewed') THEN 'assigned' ELSE status END,
		    assignee_id = COALESCE(?, assignee_id),
		    assignee_name = COALESCE(?, assignee_name),
		    updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, _ = r.db.ExecContext(ctx, updateRecQuery, id, task.AssigneeID, task.AssigneeName, task.RecommendationID, task.OrgID)

	return task, nil
}

func (r *repository) GetFollowupTaskByRecID(ctx context.Context, orgID int64, recID int64) (*CustomerFollowupTask, error) {
	query := `
		SELECT id, org_id, recommendation_id, customer_id, customer_name, source_type,
		       source_id, source_reference, followup_type, title, reason,
		       suggested_action, priority, assignee_id, assignee_name, due_date,
		       status, created_at, updated_at, completed_at, created_by_id, created_by_name,
		       correlation_id, notes
		FROM customer_followup_tasks
		WHERE org_id = ? AND recommendation_id = ?
		LIMIT 1
	`
	var task CustomerFollowupTask
	if err := r.db.GetContext(ctx, &task, query, orgID, recID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

func (r *repository) ListFollowupTasks(ctx context.Context, orgID int64, customerID *int64, status string) ([]*CustomerFollowupTask, error) {
	where := []string{"org_id = ?"}
	args := []interface{}{orgID}
	if customerID != nil {
		where = append(where, "customer_id = ?")
		args = append(args, *customerID)
	}
	if status != "" {
		where = append(where, "status = ?")
		args = append(args, status)
	}
	query := fmt.Sprintf(`
		SELECT id, org_id, recommendation_id, customer_id, customer_name, source_type,
		       source_id, source_reference, followup_type, title, reason,
		       suggested_action, priority, assignee_id, assignee_name, due_date,
		       status, created_at, updated_at, completed_at, created_by_id, created_by_name,
		       correlation_id, notes
		FROM customer_followup_tasks
		WHERE %s
		ORDER BY created_at DESC
		LIMIT 100
	`, strings.Join(where, " AND "))

	var tasks []*CustomerFollowupTask
	if err := r.db.SelectContext(ctx, &tasks, query, args...); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *repository) GetFollowupStats(ctx context.Context, orgID int64) (*FollowupStats, error) {
	query := `
		SELECT 
			COUNT(CASE WHEN category IN ('customer', 'sales') AND status NOT IN ('dismissed', 'completed', 'failed') THEN 1 END) AS total_followups,
			COUNT(CASE WHEN category IN ('customer', 'sales') AND priority = 'critical' AND status NOT IN ('dismissed', 'completed', 'failed') THEN 1 END) AS critical_count,
			COUNT(CASE WHEN category IN ('customer', 'sales') AND priority = 'high' AND status NOT IN ('dismissed', 'completed', 'failed') THEN 1 END) AS high_count,
			COUNT(CASE WHEN category IN ('customer', 'sales') AND status = 'new' THEN 1 END) AS requires_review,
			COUNT(CASE WHEN followup_task_id IS NOT NULL THEN 1 END) AS tasks_created_count,
			COUNT(CASE WHEN draft_status IN ('DRAFTED', 'SAVED') THEN 1 END) AS drafts_ready_count
		FROM ai_recommendations
		WHERE org_id = ?
	`
	var stats FollowupStats
	if err := r.db.GetContext(ctx, &stats, query, orgID); err != nil {
		return nil, fmt.Errorf("failed to get followup stats: %w", err)
	}
	return &stats, nil
}

func (r *repository) LinkApproval(ctx context.Context, id int64, orgID int64, approvalID int64) error {
	query := `
		UPDATE ai_recommendations
		SET approval_id = ?, requires_approval = 1, updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, approvalID, id, orgID)
	return err
}

// GetInvoiceEvidence retrieves complete evidence for a specific invoice (Phase 2 Task 2.5)
func (r *repository) GetInvoiceEvidence(ctx context.Context, orgID int64, invoiceID int64) (*InvoiceEvidencePayload, error) {
	var inv struct {
		ID             int64      `db:"id"`
		InvoiceNumber  string     `db:"invoice_number"`
		CustomerID     int64      `db:"customer_id"`
		CustomerName   string     `db:"customer_name"`
		TotalAmount    float64    `db:"total_amount"`
		PaidAmount     float64    `db:"paid_amount"`
		BalanceDue     float64    `db:"balance_due"`
		Currency       string     `db:"currency"`
		Status         string     `db:"status"`
		InvoiceDate    time.Time  `db:"invoice_date"`
		DueDate        *time.Time `db:"due_date"`
		Subtotal       float64    `db:"subtotal"`
		TaxAmount      float64    `db:"tax_amount"`
		DiscountAmount float64    `db:"discount_amount"`
	}

	query := `
		SELECT id, invoice_number, customer_id, customer_name, total_amount, paid_amount, balance_due,
		       currency, status, invoice_date, due_date, subtotal, tax_amount, discount_amount
		FROM customer_invoices
		WHERE id = ? AND org_id = ?
		LIMIT 1
	`
	if err := r.db.GetContext(ctx, &inv, query, invoiceID, orgID); err != nil {
		return nil, fmt.Errorf("invoice not found or access denied: %w", err)
	}

	type payRow struct {
		ID            int64     `db:"id"`
		PaymentRef    string    `db:"payment_ref"`
		Amount        float64   `db:"amount"`
		PaymentMethod string    `db:"payment_method"`
		Status        string    `db:"status"`
		PaymentDate   time.Time `db:"payment_date"`
		Notes         *string   `db:"notes"`
	}
	var payRows []payRow
	payQuery := `
		SELECT id, payment_ref, amount, payment_method, status, payment_date, notes
		FROM customer_invoice_payments
		WHERE invoice_id = ? AND org_id = ?
		ORDER BY payment_date DESC
	`
	_ = r.db.SelectContext(ctx, &payRows, payQuery, invoiceID, orgID)

	payments := make([]map[string]interface{}, 0, len(payRows))
	for _, p := range payRows {
		note := ""
		if p.Notes != nil {
			note = *p.Notes
		}
		payments = append(payments, map[string]interface{}{
			"id":             p.ID,
			"payment_ref":    p.PaymentRef,
			"amount":         p.Amount,
			"payment_method": p.PaymentMethod,
			"status":         p.Status,
			"payment_date":   p.PaymentDate.Format("2006-01-02"),
			"notes":          note,
		})
	}

	daysOverdue := 0
	agingBand := "NOT_DUE"
	dueDateStr := ""
	if inv.DueDate != nil {
		dueDateStr = inv.DueDate.Format("2006-01-02")
		diff := time.Since(*inv.DueDate).Hours() / 24
		if diff > 0 && inv.BalanceDue > 0 {
			daysOverdue = int(diff)
			if daysOverdue <= 15 {
				agingBand = "1-15_DAYS"
			} else if daysOverdue <= 30 {
				agingBand = "16-30_DAYS"
			} else if daysOverdue <= 60 {
				agingBand = "31-60_DAYS"
			} else {
				agingBand = "OVER_60_DAYS"
			}
		}
	}

	var custAR struct {
		OpenCount int     `db:"open_count"`
		TotalAR   float64 `db:"total_ar"`
	}
	custQuery := `
		SELECT COUNT(id) AS open_count, COALESCE(SUM(balance_due), 0) AS total_ar
		FROM customer_invoices
		WHERE org_id = ? AND customer_id = ? AND balance_due > 0 AND status NOT IN ('Paid', 'Cancelled')
	`
	_ = r.db.GetContext(ctx, &custAR, custQuery, orgID, inv.CustomerID)

	var riskFactors []string
	var dataQualityIssues []string

	if daysOverdue > 0 {
		riskFactors = append(riskFactors, fmt.Sprintf("Invoice is %d days overdue (Aging band: %s)", daysOverdue, agingBand))
	}
	if inv.BalanceDue >= 10000 {
		riskFactors = append(riskFactors, fmt.Sprintf("High-value exposure: Outstanding balance %s %.2f exceeds $10,000 threshold", inv.Currency, inv.BalanceDue))
	}
	if custAR.OpenCount > 1 {
		riskFactors = append(riskFactors, fmt.Sprintf("Customer has %d active unpaid invoices with total AR exposure of %s %.2f", custAR.OpenCount, inv.Currency, custAR.TotalAR))
	}
	if len(payRows) > 0 && inv.BalanceDue > 0 {
		riskFactors = append(riskFactors, fmt.Sprintf("Partial payment recorded (%s %.2f paid), remaining balance %s %.2f is pending", inv.Currency, inv.PaidAmount, inv.Currency, inv.BalanceDue))
	}

	if inv.DueDate == nil {
		dataQualityIssues = append(dataQualityIssues, "Missing required invoice maturity due date")
	}
	if inv.Currency == "" {
		dataQualityIssues = append(dataQualityIssues, "Missing invoice currency declaration")
	}
	computedTotal := inv.Subtotal + inv.TaxAmount - inv.DiscountAmount
	if inv.TotalAmount > 0 && (computedTotal-inv.TotalAmount > 0.02 || inv.TotalAmount-computedTotal > 0.02) {
		dataQualityIssues = append(dataQualityIssues, fmt.Sprintf("Line item totals discrepancy: subtotal (%.2f) + tax (%.2f) - discount (%.2f) = %.2f, but total_amount is %.2f", inv.Subtotal, inv.TaxAmount, inv.DiscountAmount, computedTotal, inv.TotalAmount))
	}
	if inv.BalanceDue < 0 {
		dataQualityIssues = append(dataQualityIssues, "Negative balance due detected")
	}
	if inv.BalanceDue > inv.TotalAmount+0.01 {
		dataQualityIssues = append(dataQualityIssues, "Balance due exceeds total invoiced amount")
	}
	if strings.EqualFold(inv.Status, "Paid") && inv.BalanceDue > 0.01 {
		dataQualityIssues = append(dataQualityIssues, fmt.Sprintf("Status mismatch: Invoice marked as 'Paid' but balance_due remains %.2f", inv.BalanceDue))
	}

	return &InvoiceEvidencePayload{
		InvoiceID:         inv.ID,
		InvoiceNumber:     inv.InvoiceNumber,
		CustomerID:        inv.CustomerID,
		CustomerName:      inv.CustomerName,
		TotalAmount:       inv.TotalAmount,
		PaidAmount:        inv.PaidAmount,
		BalanceDue:        inv.BalanceDue,
		Currency:          inv.Currency,
		Status:            inv.Status,
		InvoiceDate:       inv.InvoiceDate.Format("2006-01-02"),
		DueDate:           dueDateStr,
		DaysOverdue:       daysOverdue,
		AgingBand:         agingBand,
		IsHighValue:       inv.BalanceDue >= 10000,
		PaymentsCount:     len(payRows),
		Payments:          payments,
		CustomerOpenCount: custAR.OpenCount,
		CustomerTotalAR:   custAR.TotalAR,
		RiskFactors:       riskFactors,
		DataQualityIssues: dataQualityIssues,
	}, nil
}

// GetCustomerCollectionSummary retrieves collections overview for a customer (Phase 2 Task 2.5)
func (r *repository) GetCustomerCollectionSummary(ctx context.Context, orgID int64, customerID int64) (*CustomerCollectionSummaryPayload, error) {
	var cust struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}
	if err := r.db.GetContext(ctx, &cust, "SELECT id, name FROM customers WHERE id = ? AND org_id = ?", customerID, orgID); err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	type invListRow struct {
		ID            int64      `db:"id"`
		InvoiceNumber string     `db:"invoice_number"`
		InvoiceDate   time.Time  `db:"invoice_date"`
		DueDate       *time.Time `db:"due_date"`
		Currency      string     `db:"currency"`
		TotalAmount   float64    `db:"total_amount"`
		PaidAmount    float64    `db:"paid_amount"`
		BalanceDue    float64    `db:"balance_due"`
		Status        string     `db:"status"`
	}
	var rows []invListRow
	q := `
		SELECT id, invoice_number, invoice_date, due_date, currency, total_amount, paid_amount, balance_due, status
		FROM customer_invoices
		WHERE customer_id = ? AND org_id = ?
		ORDER BY due_date DESC
	`
	_ = r.db.SelectContext(ctx, &rows, q, customerID, orgID)

	var totalInvoiced, totalPaid, totalOutstanding, totalOverdue float64
	var openCount, overdueCount, upcomingCount int
	var invoiceMaps []map[string]interface{}

	now := time.Now()
	for _, row := range rows {
		totalInvoiced += row.TotalAmount
		totalPaid += row.PaidAmount

		isPaid := strings.EqualFold(row.Status, "Paid") || row.BalanceDue <= 0
		dueDateStr := ""
		isOverdue := false
		isUpcoming := false

		if row.DueDate != nil {
			dueDateStr = row.DueDate.Format("2006-01-02")
			if !isPaid {
				if row.DueDate.Before(now) {
					isOverdue = true
					overdueCount++
					totalOverdue += row.BalanceDue
				} else if row.DueDate.Before(now.AddDate(0, 0, 7)) {
					isUpcoming = true
					upcomingCount++
				}
			}
		}

		if !isPaid {
			openCount++
			totalOutstanding += row.BalanceDue
		}

		invoiceMaps = append(invoiceMaps, map[string]interface{}{
			"id":             row.ID,
			"invoice_number": row.InvoiceNumber,
			"invoice_date":   row.InvoiceDate.Format("2006-01-02"),
			"due_date":       dueDateStr,
			"currency":       row.Currency,
			"total_amount":   row.TotalAmount,
			"paid_amount":    row.PaidAmount,
			"balance_due":    row.BalanceDue,
			"status":         row.Status,
			"is_overdue":     isOverdue,
			"is_upcoming":    isUpcoming,
		})
	}

	completionRate := 0.0
	if totalInvoiced > 0 {
		completionRate = (totalPaid / totalInvoiced) * 100.0
	}

	var riskSignals []string
	if overdueCount > 1 {
		riskSignals = append(riskSignals, fmt.Sprintf("%d invoices currently past due", overdueCount))
	} else if overdueCount == 1 {
		riskSignals = append(riskSignals, "1 invoice currently past due")
	}
	if totalOverdue > 10000 {
		riskSignals = append(riskSignals, fmt.Sprintf("High overdue balance exposure: $%.2f", totalOverdue))
	}
	if completionRate < 50 && totalInvoiced > 5000 {
		riskSignals = append(riskSignals, fmt.Sprintf("Low payment completion rate (%.1f%%)", completionRate))
	}

	return &CustomerCollectionSummaryPayload{
		CustomerID:              cust.ID,
		CustomerName:            cust.Name,
		TotalInvoiced:           totalInvoiced,
		TotalPaid:               totalPaid,
		TotalOutstanding:        totalOutstanding,
		TotalOverdue:            totalOverdue,
		OpenInvoicesCount:       openCount,
		OverdueInvoicesCount:    overdueCount,
		UpcomingDueCount:        upcomingCount,
		PaymentCompletionRate:   completionRate,
		AveragePaymentDelayDays: 0,
		RiskSignals:             riskSignals,
		Invoices:                invoiceMaps,
	}, nil
}

// GetContractEvidence retrieves complete verifiable evidence for a contract (Phase 2 Task 2.6)
func (r *repository) GetContractEvidence(ctx context.Context, orgID int64, contractID int64) (*ContractEvidencePayload, error) {
	var c struct {
		ID                int64      `db:"id"`
		ContractReference string     `db:"contract_reference"`
		ContractName      string     `db:"contract_name"`
		ContractType      string     `db:"contract_type"`
		PartyID           *int64     `db:"party_id"`
		PartyName         string     `db:"party_name"`
		TransportMode     *string    `db:"transport_mode"`
		Status            string     `db:"status"`
		Currency          *string    `db:"currency"`
		ContractValue     *float64   `db:"contract_value"`
		EffectiveDate     *time.Time `db:"effective_date"`
		ExpiryDate        *time.Time `db:"expiry_date"`
		Owner             *string    `db:"owner"`
	}
	err := r.db.GetContext(ctx, &c, `
		SELECT id, contract_reference, contract_name, contract_type, party_id, party_name,
		       transport_mode, status, currency, contract_value, effective_date, expiry_date, owner
		FROM contracts
		WHERE id = ? AND org_id = ?
	`, contractID, orgID)
	if err != nil {
		return nil, fmt.Errorf("contract not found: %w", err)
	}

	transportMode := ""
	if c.TransportMode != nil {
		transportMode = *c.TransportMode
	}
	currency := "USD"
	if c.Currency != nil {
		currency = *c.Currency
	}
	val := 0.0
	if c.ContractValue != nil {
		val = *c.ContractValue
	}
	effectiveStr := ""
	if c.EffectiveDate != nil {
		effectiveStr = c.EffectiveDate.Format("2006-01-02")
	}
	expiryStr := ""
	daysUntilExpiry := 0
	isExpired := false
	isExpiringSoon := false
	now := time.Now()
	if c.ExpiryDate != nil {
		expiryStr = c.ExpiryDate.Format("2006-01-02")
		diffHours := c.ExpiryDate.Sub(now).Hours()
		daysUntilExpiry = int(diffHours / 24)
		if c.ExpiryDate.Before(now) {
			isExpired = true
		} else if daysUntilExpiry <= 30 {
			isExpiringSoon = true
		}
	}
	ownerStr := ""
	hasMissingOwner := true
	if c.Owner != nil && strings.TrimSpace(*c.Owner) != "" {
		ownerStr = *c.Owner
		hasMissingOwner = false
	}

	// Fetch obligations
	type obRow struct {
		ID                  int64      `db:"id"`
		ObligationReference string     `db:"obligation_reference"`
		Title               string     `db:"title"`
		ObligationType      string     `db:"obligation_type"`
		ResponsibleParty    string     `db:"responsible_party"`
		Owner               *string    `db:"owner"`
		Priority            string     `db:"priority"`
		Status              string     `db:"status"`
		DueDate             *time.Time `db:"due_date"`
	}
	var obRows []obRow
	_ = r.db.SelectContext(ctx, &obRows, `
		SELECT id, obligation_reference, title, obligation_type, responsible_party, owner, priority, status, due_date
		FROM contract_obligations
		WHERE contract_id = ? AND org_id = ?
		ORDER BY due_date ASC
	`, contractID, orgID)

	var obligations []map[string]interface{}
	for _, o := range obRows {
		dueStr := ""
		if o.DueDate != nil {
			dueStr = o.DueDate.Format("2006-01-02")
		}
		own := ""
		if o.Owner != nil {
			own = *o.Owner
		}
		obligations = append(obligations, map[string]interface{}{
			"id":                   o.ID,
			"obligation_reference": o.ObligationReference,
			"title":                o.Title,
			"obligation_type":      o.ObligationType,
			"responsible_party":    o.ResponsibleParty,
			"owner":                own,
			"priority":             o.Priority,
			"status":               o.Status,
			"due_date":             dueStr,
		})
	}

	// Fetch compliance requirements
	type compRow struct {
		ID               int64      `db:"id"`
		RequirementType  string     `db:"requirement_type"`
		Title            string     `db:"title"`
		ResponsibleParty string     `db:"responsible_party"`
		ValidUntil       *time.Time `db:"valid_until"`
		Status           string     `db:"status"`
		RiskSeverity     string     `db:"risk_severity"`
	}
	var compRows []compRow
	_ = r.db.SelectContext(ctx, &compRows, `
		SELECT id, requirement_type, title, responsible_party, valid_until, status, risk_severity
		FROM contract_compliance_requirements
		WHERE contract_id = ? AND org_id = ?
		ORDER BY valid_until ASC
	`, contractID, orgID)

	pendingCompCount := 0
	var compReqs []map[string]interface{}
	for _, cr := range compRows {
		validUntilStr := ""
		if cr.ValidUntil != nil {
			validUntilStr = cr.ValidUntil.Format("2006-01-02")
		}
		if !strings.EqualFold(cr.Status, "COMPLIANT") && !strings.EqualFold(cr.Status, "VERIFIED") {
			pendingCompCount++
		}
		compReqs = append(compReqs, map[string]interface{}{
			"id":                cr.ID,
			"requirement_type":  cr.RequirementType,
			"title":             cr.Title,
			"responsible_party": cr.ResponsibleParty,
			"valid_until":       validUntilStr,
			"status":            cr.Status,
			"risk_severity":     cr.RiskSeverity,
		})
	}

	// Fetch documents (contract_documents where carrier_name matches party or linked)
	type docRow struct {
		ID                 string `db:"id"`
		FileName           string `db:"file_name"`
		FileType           string `db:"file_type"`
		Status             string `db:"status"`
		PendingReviewCount int    `db:"pending_review_count"`
		ConfirmedRateCount int    `db:"confirmed_rate_count"`
	}
	var docRows []docRow
	_ = r.db.SelectContext(ctx, &docRows, `
		SELECT id, file_name, file_type, status, pending_review_count, confirmed_rate_count
		FROM contract_documents
		WHERE org_id = ? AND (carrier_name = ? OR ? = '')
		LIMIT 10
	`, orgID, c.PartyName, c.PartyName)

	var docMaps []map[string]interface{}
	for _, d := range docRows {
		docMaps = append(docMaps, map[string]interface{}{
			"id":                   d.ID,
			"file_name":            d.FileName,
			"file_type":            d.FileType,
			"status":               d.Status,
			"pending_review_count": d.PendingReviewCount,
			"confirmed_rate_count": d.ConfirmedRateCount,
		})
	}

	var riskSignals []string
	if isExpired {
		riskSignals = append(riskSignals, fmt.Sprintf("Contract expired on %s", expiryStr))
	} else if isExpiringSoon {
		riskSignals = append(riskSignals, fmt.Sprintf("Contract expires in %d days (%s)", daysUntilExpiry, expiryStr))
	}
	if hasMissingOwner {
		riskSignals = append(riskSignals, "Contract has no assigned owner")
	}
	if pendingCompCount > 0 {
		riskSignals = append(riskSignals, fmt.Sprintf("%d compliance requirement(s) pending verification", pendingCompCount))
	}

	return &ContractEvidencePayload{
		ContractID:             c.ID,
		ContractReference:      c.ContractReference,
		ContractName:           c.ContractName,
		ContractType:           c.ContractType,
		PartyID:                c.PartyID,
		PartyName:              c.PartyName,
		TransportMode:          transportMode,
		Status:                 c.Status,
		Currency:               currency,
		ContractValue:          val,
		EffectiveDate:          effectiveStr,
		ExpiryDate:             expiryStr,
		DaysUntilExpiry:        daysUntilExpiry,
		IsExpired:              isExpired,
		IsExpiringSoon:         isExpiringSoon,
		Owner:                  ownerStr,
		HasMissingOwner:        hasMissingOwner,
		DocumentsCount:         len(docMaps),
		ObligationsCount:       len(obligations),
		ComplianceReqCount:     len(compReqs),
		PendingComplianceCount: pendingCompCount,
		Documents:              docMaps,
		Obligations:            obligations,
		ComplianceReqs:         compReqs,
		RiskSignals:            riskSignals,
		DataFreshness:          time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// GetDocumentEvidence retrieves verifiable evidence for a document (Phase 2 Task 2.6)
func (r *repository) GetDocumentEvidence(ctx context.Context, orgID int64, docID string) (*DocumentEvidencePayload, error) {
	// Try contract_documents first (UUID id)
	var cd struct {
		ID                 string  `db:"id"`
		FileName           string  `db:"file_name"`
		FileType           string  `db:"file_type"`
		Status             string  `db:"status"`
		AIDocSummary       *string `db:"ai_document_summary"`
		PendingReviewCount int     `db:"pending_review_count"`
		FailedRateCount    int     `db:"failed_rate_count"`
		ReviewNotes        *string `db:"review_notes"`
	}
	err := r.db.GetContext(ctx, &cd, `
		SELECT id, file_name, file_type, status, ai_document_summary, pending_review_count, failed_rate_count, review_notes
		FROM contract_documents
		WHERE id = ? AND org_id = ?
	`, docID, orgID)
	if err == nil {
		var riskSignals []string
		requiresHuman := false
		if cd.PendingReviewCount > 0 {
			riskSignals = append(riskSignals, fmt.Sprintf("%d extracted rates pending human review", cd.PendingReviewCount))
			requiresHuman = true
		}
		if cd.FailedRateCount > 0 {
			riskSignals = append(riskSignals, fmt.Sprintf("%d extracted rates failed validation", cd.FailedRateCount))
			requiresHuman = true
		}
		if strings.EqualFold(cd.Status, "PENDING_EXTRACTION") || strings.EqualFold(cd.Status, "PENDING_REVIEW") {
			riskSignals = append(riskSignals, "Document extraction pending review")
			requiresHuman = true
		}
		notes := ""
		if cd.ReviewNotes != nil {
			notes = *cd.ReviewNotes
		}
		confidence := 0.85
		if cd.FailedRateCount > 0 || cd.PendingReviewCount > 0 {
			confidence = 0.65
		}
		return &DocumentEvidencePayload{
			DocumentID:           cd.ID,
			DocumentType:         cd.FileType,
			FileName:             cd.FileName,
			Status:               cd.Status,
			ExtractionStatus:     cd.Status,
			ExtractionConfidence: confidence,
			DiscrepancyNotes:     notes,
			RequiresHumanReview:  requiresHuman,
			RiskSignals:          riskSignals,
			DataFreshness:        time.Now().UTC().Format(time.RFC3339),
		}, nil
	}

	// Try shipment_documents (numeric id)
	var sd struct {
		ID           int64      `db:"id"`
		ShipmentID   *int64     `db:"shipment_id"`
		DocType      string     `db:"doc_type"`
		FileName     string     `db:"file_name"`
		Status       string     `db:"status"`
		DocumentDate *time.Time `db:"document_date"`
		ExpiresAt    *time.Time `db:"expires_at"`
		RawOcrText   *string    `db:"raw_ocr_text"`
		AISummary    *string    `db:"ai_summary"`
	}
	err2 := r.db.GetContext(ctx, &sd, `
		SELECT id, shipment_id, doc_type, file_name, status, document_date, expires_at, raw_ocr_text, ai_summary
		FROM shipment_documents
		WHERE (id = ? OR file_name = ?) AND org_id = ?
	`, docID, docID, orgID)
	if err2 != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	var effStr, expStr *string
	var daysUntil *int
	isExpired := false
	now := time.Now()
	if sd.DocumentDate != nil {
		s := sd.DocumentDate.Format("2006-01-02")
		effStr = &s
	}
	if sd.ExpiresAt != nil {
		s := sd.ExpiresAt.Format("2006-01-02")
		expStr = &s
		d := int(sd.ExpiresAt.Sub(now).Hours() / 24)
		daysUntil = &d
		if sd.ExpiresAt.Before(now) {
			isExpired = true
		}
	}

	var riskSignals []string
	requiresHuman := false
	if isExpired {
		riskSignals = append(riskSignals, "Document has passed its expiration date")
		requiresHuman = true
	} else if daysUntil != nil && *daysUntil <= 30 {
		riskSignals = append(riskSignals, fmt.Sprintf("Document expires in %d days", *daysUntil))
	}
	if strings.EqualFold(sd.Status, "MISSING") {
		riskSignals = append(riskSignals, "Required document is missing from shipment file")
		requiresHuman = true
	} else if strings.EqualFold(sd.Status, "DISCREPANCY") {
		riskSignals = append(riskSignals, "Document discrepancy detected against shipment records")
		requiresHuman = true
	}

	discNotes := ""
	if sd.AISummary != nil {
		discNotes = *sd.AISummary
	}

	return &DocumentEvidencePayload{
		DocumentID:           fmt.Sprintf("%d", sd.ID),
		DocumentType:         sd.DocType,
		FileName:             sd.FileName,
		Status:               sd.Status,
		ShipmentID:           sd.ShipmentID,
		EffectiveDate:        effStr,
		ExpiryDate:           expStr,
		DaysUntilExpiry:      daysUntil,
		IsExpired:            isExpired,
		ExtractionStatus:     sd.Status,
		ExtractionConfidence: 0.90,
		DiscrepancyNotes:     discNotes,
		RequiresHumanReview:  requiresHuman,
		RiskSignals:          riskSignals,
		DataFreshness:        time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// GetComplianceEvidence retrieves verifiable evidence for a compliance requirement (Phase 2 Task 2.6)
func (r *repository) GetComplianceEvidence(ctx context.Context, orgID int64, compID int64) (*ComplianceEvidencePayload, error) {
	var cr struct {
		ID                int64      `db:"id"`
		RequirementType   string     `db:"requirement_type"`
		Title             string     `db:"title"`
		Description       *string    `db:"description"`
		ContractID        int64      `db:"contract_id"`
		ContractReference string     `db:"contract_reference"`
		ResponsibleParty  string     `db:"responsible_party"`
		ValidUntil        *time.Time `db:"valid_until"`
		Status            string     `db:"status"`
		RiskSeverity      string     `db:"risk_severity"`
	}
	err := r.db.GetContext(ctx, &cr, `
		SELECT r.id, r.requirement_type, r.title, r.description, r.contract_id,
		       c.contract_reference, r.responsible_party, r.valid_until, r.status, r.risk_severity
		FROM contract_compliance_requirements r
		JOIN contracts c ON c.id = r.contract_id
		WHERE r.id = ? AND r.org_id = ?
	`, compID, orgID)
	if err != nil {
		return nil, fmt.Errorf("compliance requirement not found: %w", err)
	}

	desc := cr.Title
	if cr.Description != nil && *cr.Description != "" {
		desc = *cr.Description
	}

	var dueDateStr *string
	var daysUntilDue *int
	isOverdue := false
	now := time.Now()
	if cr.ValidUntil != nil {
		s := cr.ValidUntil.Format("2006-01-02")
		dueDateStr = &s
		d := int(cr.ValidUntil.Sub(now).Hours() / 24)
		daysUntilDue = &d
		if cr.ValidUntil.Before(now) {
			isOverdue = true
		}
	}

	var riskSignals []string
	if isOverdue {
		riskSignals = append(riskSignals, "Compliance validity has expired")
	} else if daysUntilDue != nil && *daysUntilDue <= 30 {
		riskSignals = append(riskSignals, fmt.Sprintf("Compliance validity expiring in %d days", *daysUntilDue))
	}
	if !strings.EqualFold(cr.Status, "COMPLIANT") && !strings.EqualFold(cr.Status, "VERIFIED") {
		riskSignals = append(riskSignals, fmt.Sprintf("Requirement status is %s (verification incomplete)", cr.Status))
	}

	return &ComplianceEvidencePayload{
		ComplianceID:      cr.ID,
		RequirementType:   cr.RequirementType,
		Description:       desc,
		ContractID:        cr.ContractID,
		ContractReference: cr.ContractReference,
		Mandatory:         true,
		Status:            cr.Status,
		DueDate:           dueDateStr,
		DaysUntilDue:      daysUntilDue,
		IsOverdue:         isOverdue,
		RiskSignals:       riskSignals,
		DataFreshness:     time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// GetContractComplianceSummary returns overall metrics and active recommendations for contracts, documents & compliance
func (r *repository) GetContractComplianceSummary(ctx context.Context, orgID int64) (*ContractComplianceSummaryPayload, error) {
	summary := &ContractComplianceSummaryPayload{}

	// 1. Contract counts
	var cStats struct {
		Total        int `db:"total"`
		Active       int `db:"active"`
		Expired      int `db:"expired"`
		ExpiringSoon int `db:"expiring_soon"`
		MissingOwner int `db:"missing_owner"`
	}
	_ = r.db.GetContext(ctx, &cStats, `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'ACTIVE' THEN 1 END) as active,
			COUNT(CASE WHEN expiry_date < NOW() THEN 1 END) as expired,
			COUNT(CASE WHEN expiry_date >= NOW() AND expiry_date <= DATE_ADD(NOW(), INTERVAL 30 DAY) THEN 1 END) as expiring_soon,
			COUNT(CASE WHEN owner IS NULL OR TRIM(owner) = '' THEN 1 END) as missing_owner
		FROM contracts
		WHERE org_id = ?
	`, orgID)
	summary.TotalContracts = cStats.Total
	summary.ActiveContracts = cStats.Active
	summary.ExpiredContracts = cStats.Expired
	summary.ExpiringSoonContracts = cStats.ExpiringSoon
	summary.MissingOwnerContracts = cStats.MissingOwner

	// 2. Document counts (shipment_documents + contract_documents)
	var sDocStats struct {
		Total        int `db:"total"`
		Expired      int `db:"expired"`
		ExpiringSoon int `db:"expiring_soon"`
		Discrepancy  int `db:"discrepancy"`
	}
	_ = r.db.GetContext(ctx, &sDocStats, `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN expires_at < NOW() THEN 1 END) as expired,
			COUNT(CASE WHEN expires_at >= NOW() AND expires_at <= DATE_ADD(NOW(), INTERVAL 30 DAY) THEN 1 END) as expiring_soon,
			COUNT(CASE WHEN status = 'DISCREPANCY' THEN 1 END) as discrepancy
		FROM shipment_documents
		WHERE org_id = ?
	`, orgID)

	var cDocCount int
	_ = r.db.GetContext(ctx, &cDocCount, "SELECT COUNT(*) FROM contract_documents WHERE org_id = ?", orgID)

	summary.TotalDocuments = sDocStats.Total + cDocCount
	summary.ExpiredDocuments = sDocStats.Expired
	summary.ExpiringSoonDocuments = sDocStats.ExpiringSoon
	summary.DiscrepancyDocuments = sDocStats.Discrepancy

	// 3. Pending compliance requirements
	var pendingComp int
	_ = r.db.GetContext(ctx, &pendingComp, `
		SELECT COUNT(*) 
		FROM contract_compliance_requirements 
		WHERE org_id = ? AND status NOT IN ('COMPLIANT', 'VERIFIED')
	`, orgID)
	summary.PendingComplianceReqs = pendingComp

	// 4. Active recommendations in CategoryContract and CategoryCompliance
	recs, _, err := r.List(ctx, orgID, RecommendationFilter{
		Limit:   10,
		SortBy:  "priority",
		SortDir: "DESC",
	})
	if err == nil {
		var filtered []*Recommendation
		for _, rec := range recs {
			if rec.Category == CategoryContract || rec.Category == CategoryCompliance ||
				rec.SourceType == SourceContract || rec.SourceType == SourceDocument || rec.SourceType == SourceCompliance {
				filtered = append(filtered, rec)
			}
		}
		summary.ActiveRecommendations = filtered
		summary.TotalActiveRisks = len(filtered)
	}

	return summary, nil
}



