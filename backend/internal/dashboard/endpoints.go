package dashboard

import (
	"context"

	"github.com/freel/backend/internal/dashboard/spec"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/svcerror"
	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	GetMissionControlEP endpoint.Endpoint
}

func NewAllDashboardEndpoints(bl BusinessLogic) Endpoints {
	return Endpoints{
		GetMissionControlEP: makeGetMissionControlEndpoint(bl),
	}
}

func makeGetMissionControlEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetMissionControlRequest)

		if req.OrgID == 0 {
			userCtx, ok := middleware.GetUserContext(ctx)
			if !ok {
				return nil, svcerror.NewServiceError(svcerror.ErrInsufficientResourceAccess)
			}
			req.OrgID = userCtx.OrgID
		}

		resp, err := bl.GetMissionControl(ctx, req.OrgID, req.StartDate, req.EndDate, req.Preset)
		if err != nil {
			return nil, err
		}

		// Wrap in a data envelope for consistent API response
		return map[string]interface{}{"data": resp}, nil
	}
}
