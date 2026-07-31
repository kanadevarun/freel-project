package reports

import (
	"context"

	"github.com/freel/backend/internal/reports/spec"
	"github.com/freel/backend/internal/svcerror"
)

type BusinessLogic interface {
	GetMetrics(ctx context.Context, orgID int32) (*spec.GetMetricsResponse, error)
}

type businessLogic struct {
	dl Datalayer
}

func NewBusinessLogic(dl Datalayer) BusinessLogic {
	return &businessLogic{dl: dl}
}

func (b *businessLogic) GetMetrics(ctx context.Context, orgID int32) (*spec.GetMetricsResponse, error) {
	if orgID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	metrics, err := b.dl.GetMetrics(ctx, orgID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	return metrics, nil
}
