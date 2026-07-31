package rfq

import (
	"context"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/rfq/spec"
	"github.com/freel/backend/internal/svcerror"
	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	ListRFQsEP             endpoint.Endpoint
	GetRFQEP               endpoint.Endpoint
	GetTimelineEP          endpoint.Endpoint
	GetAgentStatusEP       endpoint.Endpoint
	CreateRFQEP            endpoint.Endpoint
	UpdateStageEP          endpoint.Endpoint
	ParseShipmentRequestEP endpoint.Endpoint
	AddQuoteEP             endpoint.Endpoint
}

func NewAllRFQEndpoints(bl BusinessLogic) Endpoints {
	return Endpoints{
		ListRFQsEP:             makeListRFQsEndpoint(bl),
		GetRFQEP:               makeGetRFQEndpoint(bl),
		GetTimelineEP:          makeGetTimelineEndpoint(bl),
		GetAgentStatusEP:       makeGetAgentStatusEndpoint(bl),
		CreateRFQEP:            makeCreateRFQEndpoint(bl),
		UpdateStageEP:          makeUpdateStageEndpoint(bl),
		ParseShipmentRequestEP: makeParseShipmentRequestEndpoint(bl),
		AddQuoteEP:             makeAddQuoteEndpoint(bl),
	}
}

func getOrgID(ctx context.Context) (int32, error) {
	userCtx, ok := ctx.Value(middleware.UserContextKey).(*middleware.UserContext)
	if !ok {
		return 0, svcerror.NewServiceError(svcerror.ErrInsufficientResourceAccess)
	}
	return int32(userCtx.OrgID), nil
}

func makeListRFQsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ListRFQsRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		resp, err := bl.ListRFQs(ctx, *req)
		if err != nil {
			return nil, err
		}

		return resp, nil // Returns Data and TotalCount
	}
}

func makeGetRFQEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetRFQRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := bl.GetRFQ(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}

		return &spec.GetRFQResponse{Data: *resp}, nil
	}
}

func makeGetTimelineEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetTimelineRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		_, err = bl.GetRFQ(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}

		// Mock implementation
		return &spec.GetTimelineResponse{Data: []interface{}{}}, nil
	}
}

func makeGetAgentStatusEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetAgentStatusRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		rfq, err := bl.GetRFQ(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}

		return &spec.GetAgentStatusResponse{
			Data: map[string]interface{}{"agent_status": rfq.AgentStatus},
		}, nil
	}
}

func makeCreateRFQEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateRFQRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		resp, err := bl.CreateRFQ(ctx, *req)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeUpdateStageEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateStageRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := bl.AdvanceStage(ctx, orgID, req.ID, req.Stage)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeParseShipmentRequestEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ParseShipmentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		// Mock AI parser
		return &spec.ParseShipmentResponse{
			Data: map[string]interface{}{
				"origin":      "Shanghai",
				"destination": "Los Angeles",
				"incoterms":   "FOB",
				"items": []map[string]interface{}{
					{"description": "Electronics", "quantity": 10},
				},
				"confidence":     92,
				"missing_fields": []string{"Target Date", "Weight"},
			},
		}, nil
	}
}

func makeAddQuoteEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.AddQuoteRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		req.Quote.RFQID = req.ID

		err = bl.AddQuote(ctx, orgID, &req.Quote)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": req.Quote}, nil
	}
}
