package outreach

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/freel/backend/internal/outreach/spec"
	"github.com/jmoiron/sqlx"
)

type Datalayer interface {
	CreateCampaign(ctx context.Context, c *spec.Campaign) error
	GetCampaignByID(ctx context.Context, orgID, id int32) (*spec.Campaign, error)
	ListCampaigns(ctx context.Context, orgID int32, limit, offset int) ([]*spec.Campaign, int, error)
	UpdateCampaign(ctx context.Context, c *spec.Campaign) error
	DeleteCampaign(ctx context.Context, orgID, id int32) error
	AddSequence(ctx context.Context, s *spec.Sequence) error
}

type dataLayer struct {
	db *sqlx.DB
}

func NewDataLayer(db *sqlx.DB) Datalayer {
	return &dataLayer{db: db}
}

func (d *dataLayer) CreateCampaign(ctx context.Context, c *spec.Campaign) error {
	query := `
		INSERT INTO outreach_campaigns (org_id, name, status, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
	`
	res, err := d.db.ExecContext(ctx, query, c.OrgID, c.Name, c.Status)
	if err != nil {
		return fmt.Errorf("outreach.CreateCampaign: insert failed: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("outreach.CreateCampaign: get id failed: %w", err)
	}
	c.ID = int32(id)
	return nil
}

func (d *dataLayer) GetCampaignByID(ctx context.Context, orgID, id int32) (*spec.Campaign, error) {
	c := &spec.Campaign{}
	query := `
		SELECT id, org_id, name, status, created_at, updated_at
		FROM outreach_campaigns
		WHERE id = ? AND org_id = ?
	`
	err := d.db.GetContext(ctx, c, query, id, orgID)
	if err != nil {
		return nil, sql.ErrNoRows
	}

	// Fetch sequences
	seqQuery := `
		SELECT id, campaign_id, step_number, channel, template, delay_days, created_at, updated_at
		FROM outreach_sequences
		WHERE campaign_id = ?
		ORDER BY step_number ASC
	`
	err = d.db.SelectContext(ctx, &c.Sequences, seqQuery, c.ID)
	if err != nil {
		return nil, err
	}
	if c.Sequences == nil {
		c.Sequences = []spec.Sequence{}
	}

	return c, nil
}

func (d *dataLayer) ListCampaigns(ctx context.Context, orgID int32, limit, offset int) ([]*spec.Campaign, int, error) {
	var campaigns []*spec.Campaign
	var total int

	countQuery := `SELECT COUNT(*) FROM outreach_campaigns WHERE org_id = ?`
	if err := d.db.GetContext(ctx, &total, countQuery, orgID); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*spec.Campaign{}, 0, nil
	}

	listQuery := `
		SELECT id, org_id, name, status, created_at, updated_at
		FROM outreach_campaigns
		WHERE org_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	if err := d.db.SelectContext(ctx, &campaigns, listQuery, orgID, limit, offset); err != nil {
		return nil, 0, err
	}

	return campaigns, total, nil
}

func (d *dataLayer) UpdateCampaign(ctx context.Context, c *spec.Campaign) error {
	query := `
		UPDATE outreach_campaigns
		SET name = ?, status = ?, updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := d.db.ExecContext(ctx, query, c.Name, c.Status, c.ID, c.OrgID)
	return err
}

func (d *dataLayer) DeleteCampaign(ctx context.Context, orgID, id int32) error {
	query := `DELETE FROM outreach_campaigns WHERE id = ? AND org_id = ?`
	res, err := d.db.ExecContext(ctx, query, id, orgID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (d *dataLayer) AddSequence(ctx context.Context, s *spec.Sequence) error {
	query := `
		INSERT INTO outreach_sequences (campaign_id, step_number, channel, template, delay_days, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`
	res, err := d.db.ExecContext(ctx, query, s.CampaignID, s.StepNumber, s.Channel, s.Template, s.DelayDays)
	if err != nil {
		return fmt.Errorf("outreach.AddSequence: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("outreach.AddSequence get id: %w", err)
	}
	s.ID = int32(id)
	return nil
}
