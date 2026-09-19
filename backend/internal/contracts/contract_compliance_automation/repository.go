package contract_compliance_automation

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/freel/backend/internal/contracts"
	"github.com/jmoiron/sqlx"
)

// Repository defines data access methods for contract compliance automation
type Repository interface {
	GetContractByID(ctx context.Context, orgID, contractID int64) (*contracts.Contract, error)
	GetContractTerms(ctx context.Context, orgID, contractID int64) ([]map[string]interface{}, error)
	GetContractObligations(ctx context.Context, orgID, contractID int64) ([]map[string]interface{}, error)
	GetContractComplianceRequirements(ctx context.Context, orgID, contractID int64) ([]map[string]interface{}, error)
	GetContractDocuments(ctx context.Context, orgID, contractID int64) ([]map[string]interface{}, error)
	SaveReview(ctx context.Context, review *ContractComplianceReview) (*ContractComplianceReview, error)
	GetLatestReview(ctx context.Context, orgID, contractID int64) (*ContractComplianceReview, error)
	CreateDraft(ctx context.Context, draft *ContractComplianceDraft) (*ContractComplianceDraft, error)
	GetDraft(ctx context.Context, orgID, draftID int64) (*ContractComplianceDraft, error)
	ListDrafts(ctx context.Context, orgID, contractID int64) ([]*ContractComplianceDraft, error)
	UpdateDraft(ctx context.Context, draft *ContractComplianceDraft) (*ContractComplianceDraft, error)
	UpdateDraftApproval(ctx context.Context, orgID, draftID, approvalID int64, status string) error
}

type repository struct {
	db *sqlx.DB
}

// NewRepository creates a new Repository instance
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetContractByID(ctx context.Context, orgID, contractID int64) (*contracts.Contract, error) {
	query := `
		SELECT 
			id, org_id, contract_reference, contract_name, contract_type, party_id, party_name,
			transport_mode, status, currency, contract_value, effective_date, expiry_date,
			owner, description, notes, created_by, updated_by, created_at, updated_at, archived_at
		FROM contracts
		WHERE id = ? AND org_id = ?
	`
	var c contracts.Contract
	err := r.db.GetContext(ctx, &c, query, contractID, orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get contract: %w", err)
	}
	return &c, nil
}

func (r *repository) GetContractTerms(ctx context.Context, orgID, contractID int64) ([]map[string]interface{}, error) {
	query := `
		SELECT id, term_category, term_key, term_title, term_value, value_type, currency, is_critical
		FROM contract_terms
		WHERE contract_id = ? AND org_id = ?
		ORDER BY display_order ASC, id ASC
	`
	rows, err := r.db.QueryxContext(ctx, query, contractID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var terms []map[string]interface{}
	for rows.Next() {
		entry := make(map[string]interface{})
		if err := rows.MapScan(entry); err == nil {
			terms = append(terms, entry)
		}
	}
	return terms, nil
}

func (r *repository) GetContractObligations(ctx context.Context, orgID, contractID int64) ([]map[string]interface{}, error) {
	query := `
		SELECT id, obligation_reference, title, description, obligation_type, category, responsible_party, priority, status
		FROM contract_obligations
		WHERE contract_id = ? AND org_id = ?
		ORDER BY id ASC
	`
	rows, err := r.db.QueryxContext(ctx, query, contractID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var obligations []map[string]interface{}
	for rows.Next() {
		entry := make(map[string]interface{})
		if err := rows.MapScan(entry); err == nil {
			obligations = append(obligations, entry)
		}
	}
	return obligations, nil
}

func (r *repository) GetContractComplianceRequirements(ctx context.Context, orgID, contractID int64) ([]map[string]interface{}, error) {
	query := `
		SELECT id, requirement_type, title, description, responsible_party, valid_from, valid_until, status, risk_severity
		FROM contract_compliance_requirements
		WHERE contract_id = ? AND org_id = ?
		ORDER BY id ASC
	`
	rows, err := r.db.QueryxContext(ctx, query, contractID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requirements []map[string]interface{}
	for rows.Next() {
		entry := make(map[string]interface{})
		if err := rows.MapScan(entry); err == nil {
			requirements = append(requirements, entry)
		}
	}
	return requirements, nil
}

func (r *repository) GetContractDocuments(ctx context.Context, orgID, contractID int64) ([]map[string]interface{}, error) {
	// Query contract_documents joined via contract_links or direct carrier_scac
	query := `
		SELECT cd.id, cd.carrier_scac, cd.carrier_name, cd.file_name, cd.file_type, cd.status, cd.ai_document_summary
		FROM contract_documents cd
		WHERE cd.org_id = ?
		ORDER BY cd.created_at DESC
		LIMIT 10
	`
	rows, err := r.db.QueryxContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []map[string]interface{}
	for rows.Next() {
		entry := make(map[string]interface{})
		if err := rows.MapScan(entry); err == nil {
			docs = append(docs, entry)
		}
	}
	return docs, nil
}

func (r *repository) SaveReview(ctx context.Context, review *ContractComplianceReview) (*ContractComplianceReview, error) {
	query := `
		INSERT INTO ai_contract_compliance_reviews (
			org_id, contract_id, document_id, risk_level, risk_score, compliance_status,
			executive_summary, deterministic_signals, extracted_clauses, structured_discrepancies,
			compliance_obligations, missing_information, recommendations, evidence,
			confidence_score, correlation_id, created_at
		) VALUES (
			:org_id, :contract_id, :document_id, :risk_level, :risk_score, :compliance_status,
			:executive_summary, :deterministic_signals, :extracted_clauses, :structured_discrepancies,
			:compliance_obligations, :missing_information, :recommendations, :evidence,
			:confidence_score, :correlation_id, NOW()
		)
	`
	res, err := r.db.NamedExecContext(ctx, query, review)
	if err != nil {
		return nil, fmt.Errorf("failed to insert contract compliance review: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	review.ID = id
	review.CreatedAt = time.Now()
	return review, nil
}

func (r *repository) GetLatestReview(ctx context.Context, orgID, contractID int64) (*ContractComplianceReview, error) {
	query := `
		SELECT 
			id, org_id, contract_id, document_id, risk_level, risk_score, compliance_status,
			executive_summary, deterministic_signals, extracted_clauses, structured_discrepancies,
			compliance_obligations, missing_information, recommendations, evidence,
			confidence_score, correlation_id, created_at
		FROM ai_contract_compliance_reviews
		WHERE contract_id = ? AND org_id = ?
		ORDER BY id DESC
		LIMIT 1
	`
	var rev ContractComplianceReview
	err := r.db.GetContext(ctx, &rev, query, contractID, orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest review: %w", err)
	}
	return &rev, nil
}

func (r *repository) CreateDraft(ctx context.Context, draft *ContractComplianceDraft) (*ContractComplianceDraft, error) {
	query := `
		INSERT INTO ai_contract_compliance_drafts (
			org_id, contract_id, document_id, draft_type, subject, message_body, internal_notes,
			recipient_name, recipient_email, status, requires_approval, approval_id,
			action_proposal_id, created_by_user_id, correlation_id, created_at, updated_at
		) VALUES (
			:org_id, :contract_id, :document_id, :draft_type, :subject, :message_body, :internal_notes,
			:recipient_name, :recipient_email, :status, :requires_approval, :approval_id,
			:action_proposal_id, :created_by_user_id, :correlation_id, NOW(), NOW()
		)
	`
	res, err := r.db.NamedExecContext(ctx, query, draft)
	if err != nil {
		return nil, fmt.Errorf("failed to insert compliance draft: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	draft.ID = id
	draft.CreatedAt = time.Now()
	draft.UpdatedAt = time.Now()
	return draft, nil
}

func (r *repository) GetDraft(ctx context.Context, orgID, draftID int64) (*ContractComplianceDraft, error) {
	query := `
		SELECT 
			id, org_id, contract_id, document_id, draft_type, subject, message_body, internal_notes,
			recipient_name, recipient_email, status, requires_approval, approval_id,
			action_proposal_id, created_by_user_id, correlation_id, created_at, updated_at
		FROM ai_contract_compliance_drafts
		WHERE id = ? AND org_id = ?
	`
	var d ContractComplianceDraft
	err := r.db.GetContext(ctx, &d, query, draftID, orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get draft: %w", err)
	}
	return &d, nil
}

func (r *repository) ListDrafts(ctx context.Context, orgID, contractID int64) ([]*ContractComplianceDraft, error) {
	query := `
		SELECT 
			id, org_id, contract_id, document_id, draft_type, subject, message_body, internal_notes,
			recipient_name, recipient_email, status, requires_approval, approval_id,
			action_proposal_id, created_by_user_id, correlation_id, created_at, updated_at
		FROM ai_contract_compliance_drafts
		WHERE contract_id = ? AND org_id = ?
		ORDER BY id DESC
	`
	var drafts []*ContractComplianceDraft
	err := r.db.SelectContext(ctx, &drafts, query, contractID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list drafts: %w", err)
	}
	return drafts, nil
}

func (r *repository) UpdateDraft(ctx context.Context, draft *ContractComplianceDraft) (*ContractComplianceDraft, error) {
	query := `
		UPDATE ai_contract_compliance_drafts
		SET 
			subject = :subject,
			message_body = :message_body,
			recipient_name = :recipient_name,
			recipient_email = :recipient_email,
			internal_notes = :internal_notes,
			updated_at = NOW()
		WHERE id = :id AND org_id = :org_id
	`
	_, err := r.db.NamedExecContext(ctx, query, draft)
	if err != nil {
		return nil, fmt.Errorf("failed to update draft: %w", err)
	}
	draft.UpdatedAt = time.Now()
	return draft, nil
}

func (r *repository) UpdateDraftApproval(ctx context.Context, orgID, draftID, approvalID int64, status string) error {
	query := `
		UPDATE ai_contract_compliance_drafts
		SET approval_id = ?, status = ?, updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, approvalID, status, draftID, orgID)
	return err
}
