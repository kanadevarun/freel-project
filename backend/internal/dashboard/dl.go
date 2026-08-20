package dashboard

import (
	"context"
	"time"

	"github.com/freel/backend/internal/dashboard/spec"
	"github.com/jmoiron/sqlx"
)

type Datalayer interface {
	GetStats(ctx context.Context, orgID int32) (spec.Stats, error)
	GetApprovalQueue(ctx context.Context, orgID int32) ([]spec.PendingTask, error)
}

type dataLayer struct {
	db *sqlx.DB
}

func NewDataLayer(db *sqlx.DB) Datalayer {
	return &dataLayer{db: db}
}

func (d *dataLayer) GetStats(ctx context.Context, orgID int32) (spec.Stats, error) {
	var stats spec.Stats

	// ── Open RFQs (any stage except terminal WON/LOST) ────────────────────────
	err := d.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM rfqs WHERE org_id = ? AND stage NOT IN ('WON', 'LOST')`,
		orgID,
	).Scan(&stats.OpenRFQs)
	if err != nil {
		return stats, err
	}

	// ── Open Leads (customers still in pipeline, not yet converted) ───────────
	err = d.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM leads WHERE org_id = ? AND status NOT IN ('CONVERTED', 'LOST')`,
		orgID,
	).Scan(&stats.OpenLeads)
	if err != nil {
		return stats, err
	}

	// ── Total Revenue from WON RFQs ───────────────────────────────────────────
	// We SUM the sell_price of the recommended (approved) quote for every WON RFQ.
	// COALESCE handles the case where no WON RFQs exist yet → returns 0.
	err = d.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(q.sell_price), 0)
		FROM rfq_quotes q
		INNER JOIN rfqs r ON r.id = q.rfq_id
		WHERE r.org_id = ?
		  AND r.stage = 'WON'
		  AND q.status = 'APPROVED'
	`, orgID).Scan(&stats.TotalRevenue)
	if err != nil {
		// Non-fatal: revenue may be unavailable if quotes table schema differs
		stats.TotalRevenue = 0
	}

	// ── Win Rate ──────────────────────────────────────────────────────────────
	// WinRate = WON / (WON + LOST) * 100. Stored as a float 0-100.
	// We derive it in the BL layer from closed counts for simplicity.
	var wonCount, lostCount int
	d.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM rfqs WHERE org_id = ? AND stage = 'WON'`, orgID,
	).Scan(&wonCount)
	d.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM rfqs WHERE org_id = ? AND stage = 'LOST'`, orgID,
	).Scan(&lostCount)

	total := wonCount + lostCount
	if total > 0 {
		stats.WinRate = float64(wonCount) / float64(total) * 100
	}

	// ActiveShipments is currently not tracked in a dedicated table.
	// We approximate it as WON RFQs from the last 90 days.
	d.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM rfqs
		WHERE org_id = ? AND stage = 'WON' AND updated_at > NOW() - INTERVAL 90 DAY
	`, orgID).Scan(&stats.ActiveShipments)

	return stats, nil
}

func (d *dataLayer) GetApprovalQueue(ctx context.Context, orgID int32) ([]spec.PendingTask, error) {
	var queue []spec.PendingTask

	rows, err := d.db.QueryContext(ctx, `
		SELECT id, 'RFQ_QUOTE_DRAFT', CONCAT(origin, ' to ', destination), 'Draft Quote Awaiting Approval', created_at 
		FROM rfqs 
		WHERE org_id = ? AND agent_status = 'WAITING_FOR_HUMAN_REVIEW'
		ORDER BY created_at DESC LIMIT 5
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t spec.PendingTask
		var created time.Time
		if err := rows.Scan(&t.RefID, &t.Type, &t.Title, &t.Subtitle, &created); err != nil {
			return nil, err
		}
		t.ID = t.RefID // Use RFQ ID as task ID for now
		t.Timestamp = created.Format(time.RFC3339)
		queue = append(queue, t)
	}

	return queue, nil
}
