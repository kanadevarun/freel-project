package pricing_workflow

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Repository defines data access methods for RFQ extractions and quotation drafts
type Repository interface {
	SaveExtraction(ctx context.Context, ext *RFQRequirementsExtraction) error
	GetLatestExtractionByRFQ(ctx context.Context, orgID, rfqID int64) (*RFQRequirementsExtraction, error)
	SaveDraft(ctx context.Context, draft *QuotationDraft) error
	GetDraftByID(ctx context.Context, orgID, draftID int64) (*QuotationDraft, error)
	GetLatestDraftByRFQ(ctx context.Context, orgID, rfqID int64) (*QuotationDraft, error)
	ListDraftsByRFQ(ctx context.Context, orgID, rfqID int64) ([]*QuotationDraft, error)
	UpdateDraft(ctx context.Context, draft *QuotationDraft) error
	UpdateDraftStatus(ctx context.Context, orgID, draftID int64, status string, approvalID *int64, proposalID *string) error
}

type mysqlRepository struct {
	db *sqlx.DB
}

// NewRepository creates a new MariaDB repository instance
func NewRepository(db *sqlx.DB) Repository {
	return &mysqlRepository{db: db}
}

func (r *mysqlRepository) SaveExtraction(ctx context.Context, ext *RFQRequirementsExtraction) error {
	query := `
		INSERT INTO ai_rfq_requirements_extractions (
			org_id, rfq_id, status, extracted_data, missing_fields,
			clarification_needed, confidence_score, correlation_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	res, err := r.db.ExecContext(ctx, query,
		ext.OrgID, ext.RFQID, ext.Status, ext.ExtractedData, ext.MissingFields,
		ext.ClarificationNeeded, ext.ConfidenceScore, ext.CorrelationID,
	)
	if err != nil {
		return fmt.Errorf("failed to save rfq requirements extraction: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		ext.ID = id
	}
	return nil
}

func (r *mysqlRepository) GetLatestExtractionByRFQ(ctx context.Context, orgID, rfqID int64) (*RFQRequirementsExtraction, error) {
	query := `
		SELECT id, org_id, rfq_id, status, extracted_data, missing_fields,
		       clarification_needed, confidence_score, correlation_id, created_at, updated_at
		FROM ai_rfq_requirements_extractions
		WHERE org_id = ? AND rfq_id = ?
		ORDER BY id DESC LIMIT 1
	`
	var ext RFQRequirementsExtraction
	err := r.db.GetContext(ctx, &ext, query, orgID, rfqID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query latest rfq extraction: %w", err)
	}
	return &ext, nil
}

func (r *mysqlRepository) SaveDraft(ctx context.Context, d *QuotationDraft) error {
	query := `
		INSERT INTO ai_quotation_drafts (
			org_id, rfq_id, quotation_id, status, currency,
			base_cost, total_cost, base_sell, surcharges, discounts,
			tax_amount, total_selling_price, gross_margin_amount, gross_margin_pct, margin_health,
			rate_references, cost_components, selling_components, terms_and_conditions,
			internal_summary, customer_wording, pricing_explanation, recipient_email, recipient_name,
			validity_start, validity_end, requires_approval, approval_id, action_proposal_id,
			execution_id, created_by_user_id, correlation_id, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, NOW(), NOW()
		)
	`
	res, err := r.db.ExecContext(ctx, query,
		d.OrgID, d.RFQID, d.QuotationID, d.Status, d.Currency,
		d.BaseCost, d.TotalCost, d.BaseSell, d.Surcharges, d.Discounts,
		d.TaxAmount, d.TotalSellingPrice, d.GrossMarginAmount, d.GrossMarginPct, d.MarginHealth,
		d.RateReferences, d.CostComponents, d.SellingComponents, d.TermsAndConditions,
		d.InternalSummary, d.CustomerWording, d.PricingExplanation, d.RecipientEmail, d.RecipientName,
		d.ValidityStart, d.ValidityEnd, d.RequiresApproval, d.ApprovalID, d.ActionProposalID,
		d.ExecutionID, d.CreatedByUserID, d.CorrelationID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert quotation draft: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		d.ID = id
	}
	return nil
}

func (r *mysqlRepository) GetDraftByID(ctx context.Context, orgID, draftID int64) (*QuotationDraft, error) {
	query := `
		SELECT id, org_id, rfq_id, quotation_id, status, currency,
		       base_cost, total_cost, base_sell, surcharges, discounts,
		       tax_amount, total_selling_price, gross_margin_amount, gross_margin_pct, margin_health,
		       rate_references, cost_components, selling_components, terms_and_conditions,
		       internal_summary, customer_wording, pricing_explanation, recipient_email, recipient_name,
		       validity_start, validity_end, requires_approval, approval_id, action_proposal_id,
		       execution_id, created_by_user_id, correlation_id, created_at, updated_at
		FROM ai_quotation_drafts
		WHERE org_id = ? AND id = ?
	`
	var d QuotationDraft
	err := r.db.GetContext(ctx, &d, query, orgID, draftID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query quotation draft by id: %w", err)
	}
	return &d, nil
}

func (r *mysqlRepository) GetLatestDraftByRFQ(ctx context.Context, orgID, rfqID int64) (*QuotationDraft, error) {
	query := `
		SELECT id, org_id, rfq_id, quotation_id, status, currency,
		       base_cost, total_cost, base_sell, surcharges, discounts,
		       tax_amount, total_selling_price, gross_margin_amount, gross_margin_pct, margin_health,
		       rate_references, cost_components, selling_components, terms_and_conditions,
		       internal_summary, customer_wording, pricing_explanation, recipient_email, recipient_name,
		       validity_start, validity_end, requires_approval, approval_id, action_proposal_id,
		       execution_id, created_by_user_id, correlation_id, created_at, updated_at
		FROM ai_quotation_drafts
		WHERE org_id = ? AND rfq_id = ?
		ORDER BY id DESC LIMIT 1
	`
	var d QuotationDraft
	err := r.db.GetContext(ctx, &d, query, orgID, rfqID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query latest quotation draft: %w", err)
	}
	return &d, nil
}

func (r *mysqlRepository) ListDraftsByRFQ(ctx context.Context, orgID, rfqID int64) ([]*QuotationDraft, error) {
	query := `
		SELECT id, org_id, rfq_id, quotation_id, status, currency,
		       base_cost, total_cost, base_sell, surcharges, discounts,
		       tax_amount, total_selling_price, gross_margin_amount, gross_margin_pct, margin_health,
		       rate_references, cost_components, selling_components, terms_and_conditions,
		       internal_summary, customer_wording, pricing_explanation, recipient_email, recipient_name,
		       validity_start, validity_end, requires_approval, approval_id, action_proposal_id,
		       execution_id, created_by_user_id, correlation_id, created_at, updated_at
		FROM ai_quotation_drafts
		WHERE org_id = ? AND rfq_id = ?
		ORDER BY id DESC
	`
	var list []*QuotationDraft
	err := r.db.SelectContext(ctx, &list, query, orgID, rfqID)
	if err != nil {
		return nil, fmt.Errorf("failed to list quotation drafts: %w", err)
	}
	return list, nil
}

func (r *mysqlRepository) UpdateDraft(ctx context.Context, d *QuotationDraft) error {
	query := `
		UPDATE ai_quotation_drafts
		SET base_sell = ?, surcharges = ?, discounts = ?, total_selling_price = ?,
		    gross_margin_amount = ?, gross_margin_pct = ?, margin_health = ?,
		    internal_summary = ?, customer_wording = ?, terms_and_conditions = ?,
		    recipient_name = ?, recipient_email = ?, validity_end = ?,
		    requires_approval = ?, updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		d.BaseSell, d.Surcharges, d.Discounts, d.TotalSellingPrice,
		d.GrossMarginAmount, d.GrossMarginPct, d.MarginHealth,
		d.InternalSummary, d.CustomerWording, d.TermsAndConditions,
		d.RecipientName, d.RecipientEmail, d.ValidityEnd,
		d.RequiresApproval, d.OrgID, d.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update quotation draft: %w", err)
	}
	return nil
}

func (r *mysqlRepository) UpdateDraftStatus(ctx context.Context, orgID, draftID int64, status string, approvalID *int64, proposalID *string) error {
	query := `
		UPDATE ai_quotation_drafts
		SET status = ?, approval_id = COALESCE(?, approval_id), action_proposal_id = COALESCE(?, action_proposal_id), updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query, status, approvalID, proposalID, orgID, draftID)
	if err != nil {
		return fmt.Errorf("failed to update draft status: %w", err)
	}
	return nil
}
