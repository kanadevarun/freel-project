package rfq

import (
	"context"
	"database/sql"
	"time"

	"github.com/freel/backend/internal/rfq/spec"
	"github.com/jmoiron/sqlx"
)

type Datalayer interface {
	CreateRFQ(ctx context.Context, rfq *spec.RFQ) error
	GetRFQByID(ctx context.Context, orgID, rfqID int32) (*spec.RFQ, error)
	ListRFQs(ctx context.Context, orgID int32, limit, offset int) ([]spec.RFQ, int, error)
	UpdateStage(ctx context.Context, orgID, rfqID int32, stage string) error
	UpdateAgentStatus(ctx context.Context, orgID, rfqID int32, status string) error
	
	// Items & Quotes
	CreateRFQItem(ctx context.Context, item *spec.RFQItem) error
	CreateQuote(ctx context.Context, quote *spec.Quote) error
	GetQuotesByRFQ(ctx context.Context, rfqID int32) ([]spec.Quote, error)
}

type dataLayer struct {
	db *sqlx.DB
}

func NewDataLayer(db *sqlx.DB) Datalayer {
	return &dataLayer{db: db}
}

func (d *dataLayer) CreateRFQ(ctx context.Context, rfq *spec.RFQ) error {
	query := `
		INSERT INTO rfqs (org_id, rfq_number, customer_id, stage, origin, destination, incoterms, target_date, created_at, updated_at)
		VALUES (:org_id, :rfq_number, :customer_id, :stage, :origin, :destination, :incoterms, :target_date, :created_at, :updated_at)
		RETURNING id
	`
	rfq.CreatedAt = time.Now()
	rfq.UpdatedAt = time.Now()
	if rfq.Stage == "" {
		rfq.Stage = spec.StageRFQCreated
	}

	rows, err := d.db.NamedQueryContext(ctx, query, rfq)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&rfq.ID)
	}
	return nil
}

func (d *dataLayer) GetRFQByID(ctx context.Context, orgID, rfqID int32) (*spec.RFQ, error) {
	var rfq spec.RFQ
	query := `SELECT * FROM rfqs WHERE id = $1 AND org_id = $2`
	err := d.db.GetContext(ctx, &rfq, query, rfqID, orgID)
	if err != nil {
		return nil, err
	}
	
	// Load items
	d.db.SelectContext(ctx, &rfq.Items, `SELECT * FROM rfq_items WHERE rfq_id = $1`, rfq.ID)
	
	// Load quotes
	d.db.SelectContext(ctx, &rfq.Quotes, `SELECT * FROM rfq_quotes WHERE rfq_id = $1`, rfq.ID)

	return &rfq, nil
}

func (d *dataLayer) ListRFQs(ctx context.Context, orgID int32, limit, offset int) ([]spec.RFQ, int, error) {
	var rfqs []spec.RFQ
	query := `
		SELECT * FROM rfqs 
		WHERE org_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`
	err := d.db.SelectContext(ctx, &rfqs, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var total int
	d.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM rfqs WHERE org_id = $1`, orgID)

	return rfqs, total, nil
}

func (d *dataLayer) UpdateStage(ctx context.Context, orgID, rfqID int32, stage string) error {
	query := `UPDATE rfqs SET stage = $1, updated_at = NOW() WHERE id = $2 AND org_id = $3`
	res, err := d.db.ExecContext(ctx, query, stage, rfqID, orgID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (d *dataLayer) UpdateAgentStatus(ctx context.Context, orgID, rfqID int32, status string) error {
	query := `UPDATE rfqs SET agent_status = $1, updated_at = NOW() WHERE id = $2 AND org_id = $3`
	res, err := d.db.ExecContext(ctx, query, status, rfqID, orgID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (d *dataLayer) CreateRFQItem(ctx context.Context, item *spec.RFQItem) error {
	query := `
		INSERT INTO rfq_items (rfq_id, description, quantity, weight_kg, volume_cbm, created_at, updated_at)
		VALUES (:rfq_id, :description, :quantity, :weight_kg, :volume_cbm, :created_at, :updated_at)
		RETURNING id
	`
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	
	rows, err := d.db.NamedQueryContext(ctx, query, item)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return rows.Scan(&item.ID)
	}
	return nil
}

func (d *dataLayer) CreateQuote(ctx context.Context, quote *spec.Quote) error {
	query := `
		INSERT INTO rfq_quotes (rfq_id, carrier_name, transit_time_days, buy_price, sell_price, is_recommended, reliability_score, historical_success_rate, ai_reasoning, status, created_at, updated_at)
		VALUES (:rfq_id, :carrier_name, :transit_time_days, :buy_price, :sell_price, :is_recommended, :reliability_score, :historical_success_rate, :ai_reasoning, :status, :created_at, :updated_at)
		RETURNING id
	`
	quote.CreatedAt = time.Now()
	quote.UpdatedAt = time.Now()
	if quote.Status == "" {
		quote.Status = "DRAFT"
	}
	
	rows, err := d.db.NamedQueryContext(ctx, query, quote)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return rows.Scan(&quote.ID)
	}
	return nil
}

func (d *dataLayer) GetQuotesByRFQ(ctx context.Context, rfqID int32) ([]spec.Quote, error) {
	var quotes []spec.Quote
	query := `SELECT * FROM rfq_quotes WHERE rfq_id = $1 ORDER BY created_at ASC`
	err := d.db.SelectContext(ctx, &quotes, query, rfqID)
	return quotes, err
}
