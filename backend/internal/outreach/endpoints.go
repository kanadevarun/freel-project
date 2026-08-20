package outreach

import (
	"context"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/outreach/spec"
	"github.com/freel/backend/internal/svcerror"
	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	ListCampaignsEP    endpoint.Endpoint
	CreateCampaignEP   endpoint.Endpoint
	GetCampaignEP      endpoint.Endpoint
	ActivateCampaignEP endpoint.Endpoint
	PauseCampaignEP    endpoint.Endpoint
	DeleteCampaignEP   endpoint.Endpoint
	GenerateEmailEP    endpoint.Endpoint
}

func NewAllOutreachEndpoints(bl BusinessLogic) Endpoints {
	return Endpoints{
		ListCampaignsEP:    makeListCampaignsEndpoint(bl),
		CreateCampaignEP:   makeCreateCampaignEndpoint(bl),
		GetCampaignEP:      makeGetCampaignEndpoint(bl),
		ActivateCampaignEP: makeActivateCampaignEndpoint(bl),
		PauseCampaignEP:    makePauseCampaignEndpoint(bl),
		DeleteCampaignEP:   makeDeleteCampaignEndpoint(bl),
		GenerateEmailEP:    makeGenerateEmailEndpoint(), // We'll implement this mock for now as it used AI gateway directly
	}
}

func getOrgID(ctx context.Context) (int32, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		return 0, svcerror.NewServiceError(svcerror.ErrInsufficientResourceAccess)
	}
	return int32(userCtx.OrgID), nil
}

func makeListCampaignsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ListCampaignsRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		resp, err := bl.ListCampaigns(ctx, *req)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}

func makeCreateCampaignEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateCampaignRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		resp, err := bl.CreateCampaign(ctx, *req)
		if err != nil {
			return nil, err
		}

		return &spec.CreateCampaignResponse{Data: resp}, nil
	}
}

func makeGetCampaignEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetCampaignRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := bl.GetCampaign(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}

		return &spec.GetCampaignResponse{Data: resp}, nil
	}
}

func makeActivateCampaignEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ActivateCampaignRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		resp, err := bl.ActivateCampaign(ctx, *req)
		if err != nil {
			return nil, err
		}

		return &spec.ActivateCampaignResponse{Data: resp}, nil
	}
}

func makePauseCampaignEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.PauseCampaignRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		resp, err := bl.PauseCampaign(ctx, *req)
		if err != nil {
			return nil, err
		}

		return &spec.PauseCampaignResponse{Data: resp}, nil
	}
}

func makeDeleteCampaignEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.DeleteCampaignRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		err = bl.DeleteCampaign(ctx, *req)
		if err != nil {
			return nil, err
		}

		return &spec.DeleteCampaignResponse{
			Data: map[string]interface{}{"success": true},
		}, nil
	}
}

func makeGenerateEmailEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		_ = request.(*spec.GenerateEmailRequest)
		_, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		// Mock implementation for GenerateEmail as it requires AI Gateway
		// If we need the actual implementation, we would pass the gateway into NewAllOutreachEndpoints.
		return &spec.GenerateEmailResponse{
			Data: map[string]interface{}{
				"subject": "Hello from Freel",
				"body":    "This is a generated email.",
			},
		}, nil
	}
}
