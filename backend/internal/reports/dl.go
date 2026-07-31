package reports

import (
	"context"

	"github.com/freel/backend/internal/reports/spec"
	"github.com/jmoiron/sqlx"
)

type Datalayer interface {
	GetMetrics(ctx context.Context, orgID int32) (*spec.GetMetricsResponse, error)
}

type dataLayer struct {
	db *sqlx.DB
}

func NewDataLayer(db *sqlx.DB) Datalayer {
	return &dataLayer{db: db}
}

func (d *dataLayer) GetMetrics(ctx context.Context, orgID int32) (*spec.GetMetricsResponse, error) {
	// For MVP, returning mocked aggregated metrics.
	// In production, these would be complex aggregation queries.
	return &spec.GetMetricsResponse{
		LeadConversion: 24.5,
		RFQConversion:  68.2,
		WinRate:        42.1,
		Revenue:        125000.00,
	}, nil
}
