package rfq

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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
	// ApproveQuote sets the chosen quote to APPROVED and rejects all others for this RFQ.
	ApproveQuote(ctx context.Context, rfqID, quoteID int32) error
	CreateAITask(ctx context.Context, orgID int64, entityType string, entityID string, taskType string, payload map[string]interface{}) error
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
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	rfq.CreatedAt = time.Now()
	rfq.UpdatedAt = time.Now()
	if rfq.Stage == "" {
		rfq.Stage = spec.StageRFQCreated
	}

	res, err := d.db.ExecContext(ctx, query, rfq.OrgID, rfq.RFQNumber, rfq.CustomerID, rfq.Stage, rfq.Origin, rfq.Destination, rfq.Incoterms, rfq.TargetDate)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	rfq.ID = int32(id)
	return nil
}

func (d *dataLayer) GetRFQByID(ctx context.Context, orgID, rfqID int32) (*spec.RFQ, error) {
	var rfq spec.RFQ
	query := `SELECT * FROM rfqs WHERE id = ? AND org_id = ?`
	err := d.db.GetContext(ctx, &rfq, query, rfqID, orgID)
	if err != nil {
		return nil, err
	}
	
	// Load items
	d.db.SelectContext(ctx, &rfq.Items, `SELECT * FROM rfq_items WHERE rfq_id = ?`, rfq.ID)
	
	// Load quotes
	d.db.SelectContext(ctx, &rfq.Quotes, `SELECT * FROM rfq_quotes WHERE rfq_id = ?`, rfq.ID)

	return &rfq, nil
}

func (d *dataLayer) ListRFQs(ctx context.Context, orgID int32, limit, offset int) ([]spec.RFQ, int, error) {
	var rfqs []spec.RFQ

	// We LEFT JOIN customers so that RFQs without a customer record still show up.
	// COALESCE ensures customer_name is never null in the JSON output.
	query := `
		SELECT
			r.id,
			r.org_id,
			r.rfq_number,
			r.customer_id,
			COALESCE(comp.name, '') AS customer_name,
			r.stage,
			r.origin,
			r.destination,
			r.incoterms,
			r.target_date,
			r.sales_assignee_id,
			r.pricing_assignee_id,
			r.health_score,
			r.agent_status,
			r.created_at,
			r.updated_at
		FROM rfqs r
		LEFT JOIN customers c ON c.id = r.customer_id
		LEFT JOIN companies comp ON comp.id = c.company_id
		WHERE r.org_id = ?
		ORDER BY r.created_at DESC
		LIMIT ? OFFSET ?
	`
	err := d.db.SelectContext(ctx, &rfqs, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var total int
	d.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM rfqs WHERE org_id = ?`, orgID)

	return rfqs, total, nil
}

func (d *dataLayer) UpdateStage(ctx context.Context, orgID, rfqID int32, stage string) error {
	query := `UPDATE rfqs SET stage = ?, updated_at = NOW() WHERE id = ? AND org_id = ?`
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
	query := `UPDATE rfqs SET agent_status = ?, updated_at = NOW() WHERE id = ? AND org_id = ?`
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
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	
	res, err := d.db.ExecContext(ctx, query, item.RFQID, item.Description, item.Quantity, item.WeightKG, item.VolumeCBM)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	item.ID = int32(id)
	return nil
}

func (d *dataLayer) CreateQuote(ctx context.Context, quote *spec.Quote) error {
	query := `
		INSERT INTO rfq_quotes (rfq_id, carrier_name, transit_time_days, buy_price, sell_price, is_recommended, reliability_score, historical_success_rate, ai_reasoning, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	quote.CreatedAt = time.Now()
	quote.UpdatedAt = time.Now()
	if quote.Status == "" {
		quote.Status = "DRAFT"
	}
	
	res, err := d.db.ExecContext(ctx, query,
		quote.RFQID, quote.CarrierName, quote.TransitTimeDays, quote.BuyPrice, quote.SellPrice,
		quote.IsRecommended, quote.ReliabilityScore, quote.HistoricalSuccessRate, quote.AiReasoning, quote.Status,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	quote.ID = int32(id)
	return nil
}

func (d *dataLayer) GetQuotesByRFQ(ctx context.Context, rfqID int32) ([]spec.Quote, error) {
	var quotes []spec.Quote
	query := `SELECT * FROM rfq_quotes WHERE rfq_id = ? ORDER BY created_at ASC`
	err := d.db.SelectContext(ctx, &quotes, query, rfqID)
	return quotes, err
}

// ApproveQuote atomically marks the chosen quote as APPROVED and
// marks all other quotes for the same RFQ as REJECTED.
// We do this in a transaction so the DB is never left in a partially-approved state.
func (d *dataLayer) ApproveQuote(ctx context.Context, rfqID, quoteID int32) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Reject all other quotes for this RFQ first
	_, err = tx.ExecContext(ctx,
		`UPDATE rfq_quotes SET status = 'REJECTED', updated_at = NOW() WHERE rfq_id = ? AND id != ?`,
		rfqID, quoteID,
	)
	if err != nil {
		return err
	}

	// Approve the selected quote
	res, err := tx.ExecContext(ctx,
		`UPDATE rfq_quotes SET status = 'APPROVED', is_recommended = TRUE, updated_at = NOW() WHERE id = ? AND rfq_id = ?`,
		quoteID, rfqID,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("quote %d not found for rfq %d", quoteID, rfqID)
	}

	return tx.Commit()
}

func (d *dataLayer) CreateAITask(ctx context.Context, orgID int64, entityType string, entityID string, taskType string, payload map[string]interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal task payload: %w", err)
	}

	const query = `
INSERT INTO ai_processing_tasks (
    org_id, entity_type, entity_id, task_type, payload, status, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, 'QUEUED', NOW(), NOW()
)
`
	_, err = d.db.ExecContext(ctx, query, orgID, entityType, entityID, taskType, payloadJSON)
	return err
}
