package reports

import (
	"context"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/reports/spec"
	"github.com/freel/backend/internal/svcerror"
	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	GetMetricsEP endpoint.Endpoint
}

func NewAllReportsEndpoints(bl BusinessLogic) Endpoints {
	return Endpoints{
		GetMetricsEP: makeGetMetricsEndpoint(bl),
	}
}

func makeGetMetricsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetMetricsRequest)

		if req.OrgID == 0 {
			userCtx, ok := ctx.Value(middleware.UserContextKey).(*middleware.UserContext)
			if !ok {
				return nil, svcerror.NewServiceError(svcerror.ErrInsufficientResourceAccess)
			}
			req.OrgID = int32(userCtx.OrgID)
		}

		resp, err := bl.GetMetrics(ctx, req.OrgID)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}
