package leads

import (
	"context"

	"github.com/freel/backend/internal/leads/spec"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/svcerror"
	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	ListLeadsEP   endpoint.Endpoint
	CreateLeadEP  endpoint.Endpoint
	ImportLeadsEP endpoint.Endpoint
	GetLeadEP     endpoint.Endpoint
	UpdateLeadEP  endpoint.Endpoint
	DeleteLeadEP  endpoint.Endpoint
}

func NewAllLeadsEndpoints(bl BusinessLogic) Endpoints {
	return Endpoints{
		ListLeadsEP:   makeListLeadsEndpoint(bl),
		CreateLeadEP:  makeCreateLeadEndpoint(bl),
		ImportLeadsEP: makeImportLeadsEndpoint(bl),
		GetLeadEP:     makeGetLeadEndpoint(bl),
		UpdateLeadEP:  makeUpdateLeadEndpoint(bl),
		DeleteLeadEP:  makeDeleteLeadEndpoint(bl),
	}
}

func getOrgID(ctx context.Context) (int32, error) {
	userCtx, ok := ctx.Value(middleware.UserContextKey).(*middleware.UserContext)
	if !ok {
		return 0, svcerror.NewServiceError(svcerror.ErrInsufficientResourceAccess)
	}
	return int32(userCtx.OrgID), nil
}

func makeListLeadsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ListLeadsRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		resp, err := bl.ListLeads(ctx, *req)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}

func makeCreateLeadEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateLeadRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		resp, err := bl.CreateLead(ctx, *req)
		if err != nil {
			return nil, err
		}

		return &spec.CreateLeadResponse{Data: resp}, nil
	}
}

func makeImportLeadsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ImportLeadsRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		createdCount, err := bl.BulkCreate(ctx, *req)
		if err != nil {
			return nil, err
		}

		return &spec.ImportLeadsResponse{
			Data: map[string]interface{}{"created_count": createdCount},
		}, nil
	}
}

func makeGetLeadEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetLeadRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := bl.GetLead(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}

		return &spec.GetLeadResponse{Data: resp}, nil
	}
}

func makeUpdateLeadEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateLeadRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		resp, err := bl.UpdateLead(ctx, *req)
		if err != nil {
			return nil, err
		}

		return &spec.UpdateLeadResponse{Data: resp}, nil
	}
}

func makeDeleteLeadEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.DeleteLeadRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		err = bl.DeleteLead(ctx, *req)
		if err != nil {
			return nil, err
		}

		return &spec.DeleteLeadResponse{
			Data: map[string]interface{}{"success": true},
		}, nil
	}
}
