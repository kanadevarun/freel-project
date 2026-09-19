package operations_automation

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Repository handles database persistence for shipment operations analyses and drafts
type Repository interface {
	GetLatestAnalysis(ctx context.Context, orgID, shipmentID int64) (*ShipmentOperationsAnalysis, error)
	SaveAnalysis(ctx context.Context, analysis *ShipmentOperationsAnalysis) error

	GetLatestDraft(ctx context.Context, orgID, shipmentID int64) (*ShipmentCommunicationDraft, error)
	GetDraftByID(ctx context.Context, orgID, draftID int64) (*ShipmentCommunicationDraft, error)
	SaveDraft(ctx context.Context, draft *ShipmentCommunicationDraft) error
	UpdateDraft(ctx context.Context, draft *ShipmentCommunicationDraft) error
	UpdateDraftStatus(ctx context.Context, orgID, draftID int64, status string, approvalID *int64, proposalID *string) error
	ListDraftsByShipment(ctx context.Context, orgID, shipmentID int64) ([]*ShipmentCommunicationDraft, error)
}

type repository struct {
	db *sqlx.DB
}

// NewRepository creates a new Repository instance
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetLatestAnalysis(ctx context.Context, orgID, shipmentID int64) (*ShipmentOperationsAnalysis, error) {
	query := `
		SELECT id, org_id, shipment_id, risk_level, risk_score, operational_summary,
		       deterministic_signals, key_risks, recommended_next_steps, evidence,
		       confidence_score, correlation_id, created_at
		FROM ai_shipment_operations_analyses
		WHERE org_id = ? AND shipment_id = ?
		ORDER BY id DESC
		LIMIT 1
	`
	var a ShipmentOperationsAnalysis
	if err := r.db.GetContext(ctx, &a, query, orgID, shipmentID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query latest shipment analysis: %w", err)
	}
	return &a, nil
}

func (r *repository) SaveAnalysis(ctx context.Context, a *ShipmentOperationsAnalysis) error {
	query := `
		INSERT INTO ai_shipment_operations_analyses (
			org_id, shipment_id, risk_level, risk_score, operational_summary,
			deterministic_signals, key_risks, recommended_next_steps, evidence,
			confidence_score, correlation_id
		) VALUES (
			:org_id, :shipment_id, :risk_level, :risk_score, :operational_summary,
			:deterministic_signals, :key_risks, :recommended_next_steps, :evidence,
			:confidence_score, :correlation_id
		)
	`
	res, err := r.db.NamedExecContext(ctx, query, a)
	if err != nil {
		return fmt.Errorf("failed to insert shipment operations analysis: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		a.ID = id
	}
	return nil
}

func (r *repository) GetLatestDraft(ctx context.Context, orgID, shipmentID int64) (*ShipmentCommunicationDraft, error) {
	query := `
		SELECT id, org_id, shipment_id, draft_type, subject, customer_wording,
		       internal_notes, recipient_name, recipient_email, status,
		       requires_approval, approval_id, action_proposal_id, execution_id,
		       created_by_user_id, correlation_id, created_at, updated_at
		FROM ai_shipment_communication_drafts
		WHERE org_id = ? AND shipment_id = ?
		ORDER BY id DESC
		LIMIT 1
	`
	var d ShipmentCommunicationDraft
	if err := r.db.GetContext(ctx, &d, query, orgID, shipmentID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query latest shipment draft: %w", err)
	}
	return &d, nil
}

func (r *repository) GetDraftByID(ctx context.Context, orgID, draftID int64) (*ShipmentCommunicationDraft, error) {
	query := `
		SELECT id, org_id, shipment_id, draft_type, subject, customer_wording,
		       internal_notes, recipient_name, recipient_email, status,
		       requires_approval, approval_id, action_proposal_id, execution_id,
		       created_by_user_id, correlation_id, created_at, updated_at
		FROM ai_shipment_communication_drafts
		WHERE org_id = ? AND id = ?
		LIMIT 1
	`
	var d ShipmentCommunicationDraft
	if err := r.db.GetContext(ctx, &d, query, orgID, draftID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query shipment draft by id: %w", err)
	}
	return &d, nil
}

func (r *repository) SaveDraft(ctx context.Context, d *ShipmentCommunicationDraft) error {
	query := `
		INSERT INTO ai_shipment_communication_drafts (
			org_id, shipment_id, draft_type, subject, customer_wording,
			internal_notes, recipient_name, recipient_email, status,
			requires_approval, approval_id, action_proposal_id, execution_id,
			created_by_user_id, correlation_id
		) VALUES (
			:org_id, :shipment_id, :draft_type, :subject, :customer_wording,
			:internal_notes, :recipient_name, :recipient_email, :status,
			:requires_approval, :approval_id, :action_proposal_id, :execution_id,
			:created_by_user_id, :correlation_id
		)
	`
	res, err := r.db.NamedExecContext(ctx, query, d)
	if err != nil {
		return fmt.Errorf("failed to insert shipment communication draft: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		d.ID = id
	}
	return nil
}

func (r *repository) UpdateDraft(ctx context.Context, d *ShipmentCommunicationDraft) error {
	query := `
		UPDATE ai_shipment_communication_drafts
		SET subject = :subject,
		    customer_wording = :customer_wording,
		    internal_notes = :internal_notes,
		    recipient_name = :recipient_name,
		    recipient_email = :recipient_email,
		    status = :status,
		    requires_approval = :requires_approval,
		    approval_id = :approval_id,
		    action_proposal_id = :action_proposal_id,
		    updated_at = CURRENT_TIMESTAMP
		WHERE org_id = :org_id AND id = :id
	`
	_, err := r.db.NamedExecContext(ctx, query, d)
	if err != nil {
		return fmt.Errorf("failed to update shipment draft: %w", err)
	}
	return nil
}

func (r *repository) UpdateDraftStatus(ctx context.Context, orgID, draftID int64, status string, approvalID *int64, proposalID *string) error {
	query := `
		UPDATE ai_shipment_communication_drafts
		SET status = ?,
		    approval_id = ?,
		    action_proposal_id = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE org_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query, status, approvalID, proposalID, orgID, draftID)
	if err != nil {
		return fmt.Errorf("failed to update shipment draft status: %w", err)
	}
	return nil
}

func (r *repository) ListDraftsByShipment(ctx context.Context, orgID, shipmentID int64) ([]*ShipmentCommunicationDraft, error) {
	query := `
		SELECT id, org_id, shipment_id, draft_type, subject, customer_wording,
		       internal_notes, recipient_name, recipient_email, status,
		       requires_approval, approval_id, action_proposal_id, execution_id,
		       created_by_user_id, correlation_id, created_at, updated_at
		FROM ai_shipment_communication_drafts
		WHERE org_id = ? AND shipment_id = ?
		ORDER BY id DESC
	`
	var list []*ShipmentCommunicationDraft
	if err := r.db.SelectContext(ctx, &list, query, orgID, shipmentID); err != nil {
		return nil, fmt.Errorf("failed to list shipment drafts: %w", err)
	}
	return list, nil
}
