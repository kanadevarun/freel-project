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
	stats.ActiveShipments = 0
	stats.TotalRevenue = 0.0

	// Active RFQs
	err := d.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM rfqs WHERE org_id = $1 AND status NOT IN ('WON', 'LOST')", orgID).Scan(&stats.OpenRFQs)
	if err != nil {
		return stats, err
	}

	// Active Leads
	err = d.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM customers WHERE org_id = $1 AND status != 'CUSTOMER'", orgID).Scan(&stats.OpenLeads)
	if err != nil {
		return stats, err
	}

	return stats, nil
}

func (d *dataLayer) GetApprovalQueue(ctx context.Context, orgID int32) ([]spec.PendingTask, error) {
	var queue []spec.PendingTask

	rows, err := d.db.QueryContext(ctx, `
		SELECT id, 'RFQ_QUOTE_DRAFT', origin || ' to ' || destination, 'Draft Quote Awaiting Approval', created_at 
		FROM rfqs 
		WHERE org_id = $1 AND agent_status = 'WAITING_FOR_HUMAN_REVIEW'
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
