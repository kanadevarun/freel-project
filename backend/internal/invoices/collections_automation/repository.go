package collections_automation

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Repository defines the persistence interface for collections automation
type Repository interface {
	GetLatestAnalysis(ctx context.Context, orgID, invoiceID int64) (*ReceivablesAnalysis, error)
	SaveAnalysis(ctx context.Context, a *ReceivablesAnalysis) error

	GetLatestDraft(ctx context.Context, orgID, invoiceID int64) (*CollectionDraft, error)
	GetDraftByID(ctx context.Context, orgID, draftID int64) (*CollectionDraft, error)
	ListDraftsByInvoice(ctx context.Context, orgID, invoiceID int64) ([]*CollectionDraft, error)
	SaveDraft(ctx context.Context, d *CollectionDraft) error
	UpdateDraft(ctx context.Context, d *CollectionDraft) error
	UpdateDraftStatus(ctx context.Context, orgID, draftID int64, status string, approvalID *int64, proposalID *string) error

	GetCustomerTotalOutstanding(ctx context.Context, orgID, customerID int64) (float64, error)
	GetCustomerOverdueInvoices(ctx context.Context, orgID, customerID int64) (int, float64, error)
	GetCustomerCreditLimit(ctx context.Context, orgID, customerID int64) (float64, error)
}

type repository struct {
	db *sqlx.DB
}

// NewRepository instantiates a new Repository instance
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetLatestAnalysis(ctx context.Context, orgID, invoiceID int64) (*ReceivablesAnalysis, error) {
	query := `
		SELECT id, org_id, invoice_id, customer_id, risk_level, risk_score,
		       days_overdue, aging_bucket, outstanding_amount, currency,
		       receivables_summary, deterministic_signals, key_risks,
		       recommended_next_steps, evidence, confidence_score,
		       correlation_id, created_at
		FROM ai_finance_receivables_analyses
		WHERE org_id = ? AND invoice_id = ?
		ORDER BY id DESC
		LIMIT 1
	`
	var a ReceivablesAnalysis
	if err := r.db.GetContext(ctx, &a, query, orgID, invoiceID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch latest receivables analysis: %w", err)
	}
	return &a, nil
}

func (r *repository) SaveAnalysis(ctx context.Context, a *ReceivablesAnalysis) error {
	query := `
		INSERT INTO ai_finance_receivables_analyses (
			org_id, invoice_id, customer_id, risk_level, risk_score,
			days_overdue, aging_bucket, outstanding_amount, currency,
			receivables_summary, deterministic_signals, key_risks,
			recommended_next_steps, evidence, confidence_score,
			correlation_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`
	res, err := r.db.ExecContext(ctx, query,
		a.OrgID, a.InvoiceID, a.CustomerID, a.RiskLevel, a.RiskScore,
		a.DaysOverdue, a.AgingBucket, a.OutstandingAmount, a.Currency,
		a.ReceivablesSummary, string(a.DeterministicSignals), string(a.KeyRisks),
		string(a.RecommendedNextSteps), string(a.Evidence), a.ConfidenceScore,
		a.CorrelationID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert receivables analysis: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		a.ID = id
	}
	return nil
}

func (r *repository) GetLatestDraft(ctx context.Context, orgID, invoiceID int64) (*CollectionDraft, error) {
	query := `
		SELECT id, org_id, invoice_id, customer_id, draft_type, subject,
		       message_body, internal_notes, recipient_name, recipient_email,
		       outstanding_amount, currency, status, requires_approval,
		       approval_id, action_proposal_id, created_by_user_id,
		       correlation_id, created_at, updated_at
		FROM ai_finance_collection_drafts
		WHERE org_id = ? AND invoice_id = ?
		ORDER BY id DESC
		LIMIT 1
	`
	var d CollectionDraft
	if err := r.db.GetContext(ctx, &d, query, orgID, invoiceID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch latest collection draft: %w", err)
	}
	return &d, nil
}

func (r *repository) GetDraftByID(ctx context.Context, orgID, draftID int64) (*CollectionDraft, error) {
	query := `
		SELECT id, org_id, invoice_id, customer_id, draft_type, subject,
		       message_body, internal_notes, recipient_name, recipient_email,
		       outstanding_amount, currency, status, requires_approval,
		       approval_id, action_proposal_id, created_by_user_id,
		       correlation_id, created_at, updated_at
		FROM ai_finance_collection_drafts
		WHERE org_id = ? AND id = ?
		LIMIT 1
	`
	var d CollectionDraft
	if err := r.db.GetContext(ctx, &d, query, orgID, draftID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch collection draft by id: %w", err)
	}
	return &d, nil
}

func (r *repository) ListDraftsByInvoice(ctx context.Context, orgID, invoiceID int64) ([]*CollectionDraft, error) {
	query := `
		SELECT id, org_id, invoice_id, customer_id, draft_type, subject,
		       message_body, internal_notes, recipient_name, recipient_email,
		       outstanding_amount, currency, status, requires_approval,
		       approval_id, action_proposal_id, created_by_user_id,
		       correlation_id, created_at, updated_at
		FROM ai_finance_collection_drafts
		WHERE org_id = ? AND invoice_id = ?
		ORDER BY id DESC
	`
	var drafts []*CollectionDraft
	if err := r.db.SelectContext(ctx, &drafts, query, orgID, invoiceID); err != nil {
		return nil, fmt.Errorf("failed to list collection drafts: %w", err)
	}
	return drafts, nil
}

func (r *repository) SaveDraft(ctx context.Context, d *CollectionDraft) error {
	query := `
		INSERT INTO ai_finance_collection_drafts (
			org_id, invoice_id, customer_id, draft_type, subject,
			message_body, internal_notes, recipient_name, recipient_email,
			outstanding_amount, currency, status, requires_approval,
			approval_id, action_proposal_id, created_by_user_id,
			correlation_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	res, err := r.db.ExecContext(ctx, query,
		d.OrgID, d.InvoiceID, d.CustomerID, d.DraftType, d.Subject,
		d.MessageBody, d.InternalNotes, d.RecipientName, d.RecipientEmail,
		d.OutstandingAmount, d.Currency, d.Status, d.RequiresApproval,
		d.ApprovalID, d.ActionProposalID, d.CreatedByUserID, d.CorrelationID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert collection draft: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		d.ID = id
	}
	return nil
}

func (r *repository) UpdateDraft(ctx context.Context, d *CollectionDraft) error {
	query := `
		UPDATE ai_finance_collection_drafts
		SET subject = ?, message_body = ?, internal_notes = ?,
		    recipient_name = ?, recipient_email = ?, updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		d.Subject, d.MessageBody, d.InternalNotes,
		d.RecipientName, d.RecipientEmail, d.OrgID, d.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update collection draft: %w", err)
	}
	return nil
}

func (r *repository) UpdateDraftStatus(
	ctx context.Context,
	orgID, draftID int64,
	status string,
	approvalID *int64,
	proposalID *string,
) error {
	query := `
		UPDATE ai_finance_collection_drafts
		SET status = ?, approval_id = ?, action_proposal_id = ?, updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query, status, approvalID, proposalID, orgID, draftID)
	if err != nil {
		return fmt.Errorf("failed to update collection draft status: %w", err)
	}
	return nil
}

func (r *repository) GetCustomerTotalOutstanding(ctx context.Context, orgID, customerID int64) (float64, error) {
	query := `
		SELECT COALESCE(SUM(balance_due), 0.00)
		FROM customer_invoices
		WHERE org_id = ? AND customer_id = ? AND status != 'Paid' AND status != 'Cancelled'
	`
	var total float64
	if err := r.db.GetContext(ctx, &total, query, orgID, customerID); err != nil {
		return 0.0, err
	}
	return total, nil
}

func (r *repository) GetCustomerOverdueInvoices(ctx context.Context, orgID, customerID int64) (int, float64, error) {
	query := `
		SELECT COUNT(*), COALESCE(SUM(balance_due), 0.00)
		FROM customer_invoices
		WHERE org_id = ? AND customer_id = ? AND status != 'Paid' AND status != 'Cancelled' AND due_date < CURDATE()
	`
	var count int
	var balance float64
	row := r.db.QueryRowContext(ctx, query, orgID, customerID)
	if err := row.Scan(&count, &balance); err != nil {
		return 0, 0.0, err
	}
	return count, balance, nil
}

func (r *repository) GetCustomerCreditLimit(ctx context.Context, orgID, customerID int64) (float64, error) {
	query := `
		SELECT COALESCE(credit_limit, 0.00)
		FROM customers
		WHERE org_id = ? AND id = ?
	`
	var limit float64
	if err := r.db.GetContext(ctx, &limit, query, orgID, customerID); err != nil {
		if err == sql.ErrNoRows {
			return 0.0, nil
		}
		return 0.0, err
	}
	return limit, nil
}
