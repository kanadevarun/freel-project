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

	// New EPs
	GetCampaignSequenceEP    endpoint.Endpoint
	AddSequenceStepEP        endpoint.Endpoint
	UpdateSequenceStepEP     endpoint.Endpoint
	DeleteSequenceStepEP     endpoint.Endpoint
	ReorderSequenceEP        endpoint.Endpoint
	GetCampaignAudienceEP    endpoint.Endpoint
	AddCampaignAudienceEP    endpoint.Endpoint
	RemoveCampaignAudienceEP endpoint.Endpoint
	GetOutreachAnalyticsEP   endpoint.Endpoint
	GetCampaignAnalyticsEP   endpoint.Endpoint
	GetCampaignLeadsEP       endpoint.Endpoint
	GetCampaignInsightsEP     endpoint.Endpoint
	GetConversionFunnelEP    endpoint.Endpoint

	CreateActivityEP   endpoint.Endpoint
	GetActivityEP      endpoint.Endpoint
	UpdateActivityEP   endpoint.Endpoint
	CompleteActivityEP endpoint.Endpoint
	DeleteActivityEP   endpoint.Endpoint

	// Engagement & Prospects EPs
	GetCampaignRecipientsEP   endpoint.Endpoint
	GetCampaignActivityEP     endpoint.Endpoint
	GetProspectsEP            endpoint.Endpoint
	GetProspectEngagementEP   endpoint.Endpoint
	GetLeadOutreachActivityEP endpoint.Endpoint

	GetProspectDetailEP  endpoint.Endpoint
	EnrollProspectEP     endpoint.Endpoint
	UpdateProspectEP     endpoint.Endpoint
	PauseProspectEP      endpoint.Endpoint
	ResumeProspectEP     endpoint.Endpoint
	StopProspectEP       endpoint.Endpoint
	GetFollowUpsEP       endpoint.Endpoint
	CancelFollowUpEP     endpoint.Endpoint
	RescheduleFollowUpEP endpoint.Endpoint
}

func NewAllOutreachEndpoints(bl BusinessLogic) Endpoints {
	return Endpoints{
		ListCampaignsEP:          makeListCampaignsEndpoint(bl),
		CreateCampaignEP:         makeCreateCampaignEndpoint(bl),
		GetCampaignEP:            makeGetCampaignEndpoint(bl),
		ActivateCampaignEP:       makeActivateCampaignEndpoint(bl),
		PauseCampaignEP:          makePauseCampaignEndpoint(bl),
		DeleteCampaignEP:         makeDeleteCampaignEndpoint(bl),
		GenerateEmailEP:          makeGenerateEmailEndpoint(), // We'll implement this mock for now as it used AI gateway directly
		GetCampaignSequenceEP:    makeGetCampaignSequenceEndpoint(bl),
		AddSequenceStepEP:        makeAddSequenceStepEndpoint(bl),
		UpdateSequenceStepEP:     makeUpdateSequenceStepEndpoint(bl),
		DeleteSequenceStepEP:     makeDeleteSequenceStepEndpoint(bl),
		ReorderSequenceEP:        makeReorderSequenceEndpoint(bl),
		GetCampaignAudienceEP:    makeGetCampaignAudienceEndpoint(bl),
		AddCampaignAudienceEP:    makeAddCampaignAudienceEndpoint(bl),
		RemoveCampaignAudienceEP: makeRemoveCampaignAudienceEndpoint(bl),
		GetOutreachAnalyticsEP:   makeGetOutreachAnalyticsEndpoint(bl),
		GetCampaignAnalyticsEP:   makeGetCampaignAnalyticsEndpoint(bl),
		GetCampaignLeadsEP:       makeGetCampaignLeadsEndpoint(bl),
		GetCampaignInsightsEP:     makeGetCampaignInsightsEndpoint(bl),
		GetConversionFunnelEP:    makeGetConversionFunnelEndpoint(bl),
		CreateActivityEP:         makeCreateActivityEndpoint(bl),
		GetActivityEP:            makeGetActivityEndpoint(bl),
		UpdateActivityEP:         makeUpdateActivityEndpoint(bl),
		CompleteActivityEP:       makeCompleteActivityEndpoint(bl),
		DeleteActivityEP:         makeDeleteActivityEndpoint(bl),
		GetCampaignRecipientsEP:   makeGetCampaignRecipientsEndpoint(bl),
		GetCampaignActivityEP:     makeGetCampaignActivityEndpoint(bl),
		GetProspectsEP:            makeGetProspectsEndpoint(bl),
		GetProspectEngagementEP:   makeGetProspectEngagementEndpoint(bl),
		GetLeadOutreachActivityEP:  makeGetLeadOutreachActivityEndpoint(bl),
		GetProspectDetailEP:       makeGetProspectDetailEndpoint(bl),
		EnrollProspectEP:          makeEnrollProspectEndpoint(bl),
		UpdateProspectEP:          makeUpdateProspectEndpoint(bl),
		PauseProspectEP:           makePauseProspectEndpoint(bl),
		ResumeProspectEP:          makeResumeProspectEndpoint(bl),
		StopProspectEP:            makeStopProspectEndpoint(bl),
		GetFollowUpsEP:            makeGetFollowUpsEndpoint(bl),
		CancelFollowUpEP:          makeCancelFollowUpEndpoint(bl),
		RescheduleFollowUpEP:      makeRescheduleFollowUpEndpoint(bl),
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

func makeGetCampaignSequenceEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetCampaignRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		steps, err := bl.GetCampaignSequence(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}
		return steps, nil
	}
}

func makeAddSequenceStepEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.AddSequenceStepRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID
		step, err := bl.AddSequenceStep(ctx, *req)
		if err != nil {
			return nil, err
		}
		return step, nil
	}
}

func makeUpdateSequenceStepEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateSequenceStepRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID
		step, err := bl.UpdateSequenceStep(ctx, *req)
		if err != nil {
			return nil, err
		}
		return step, nil
	}
}

func makeDeleteSequenceStepEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.DeleteSequenceStepRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID
		err = bl.DeleteSequenceStep(ctx, *req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil
	}
}

func makeReorderSequenceEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ReorderSequenceRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID
		err = bl.ReorderSequence(ctx, *req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil
	}
}

func makeGetCampaignAudienceEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetCampaignAudienceRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		leads, err := bl.GetCampaignAudience(ctx, orgID, req.CampaignID)
		if err != nil {
			return nil, err
		}
		return &spec.GetCampaignAudienceResponse{Leads: leads}, nil
	}
}

func makeAddCampaignAudienceEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.AddCampaignAudienceRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID
		err = bl.AddCampaignAudience(ctx, *req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil
	}
}

func makeRemoveCampaignAudienceEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.RemoveCampaignAudienceRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID
		err = bl.RemoveCampaignAudience(ctx, *req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil
	}
}

func makeGetOutreachAnalyticsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		res, err := bl.GetOutreachAnalytics(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func makeGetCampaignAnalyticsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetCampaignRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		res, err := bl.GetCampaignAnalytics(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func makeGetCampaignLeadsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetCampaignRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		leads, err := bl.GetCampaignLeads(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}
		return &spec.CampaignLeadsResponse{Leads: leads}, nil
	}
}

func makeGetCampaignInsightsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetCampaignRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		insights, err := bl.GetCampaignInsights(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}
		return &spec.CampaignInsightsResponse{Insights: insights}, nil
	}
}

func makeGetConversionFunnelEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		funnel, err := bl.GetConversionFunnel(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return funnel, nil
	}
}

func makeCreateActivityEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateActivityRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		userCtx, ok := middleware.GetUserContext(ctx)
		if ok {
			userID := int64(userCtx.UserID)
			req.CreatedBy = &userID
		}

		id, err := bl.CreateActivity(ctx, req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true, "id": id}, nil
	}
}

func makeGetActivityEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetActivityRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		res, err := bl.GetActivity(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func makeUpdateActivityEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateActivityRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		err = bl.UpdateActivity(ctx, req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil
	}
}

func makeCompleteActivityEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetActivityRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		err = bl.CompleteActivity(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil
	}
}

func makeDeleteActivityEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetActivityRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		err = bl.DeleteActivity(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil
	}
}

func makeGetCampaignRecipientsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		campaignID := request.(int32)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		res, err := bl.GetCampaignRecipients(ctx, orgID, campaignID)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func makeGetCampaignActivityEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		campaignID := request.(int32)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		res, err := bl.GetCampaignActivity(ctx, orgID, campaignID)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func makeGetProspectsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		res, err := bl.GetProspects(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func makeGetProspectEngagementEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		leadID := request.(int64)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		res, err := bl.GetProspectEngagement(ctx, orgID, leadID)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func makeGetLeadOutreachActivityEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		leadID := request.(int64)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		res, err := bl.GetLeadOutreachActivity(ctx, orgID, leadID)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func makeGetProspectDetailEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		leadID := request.(int64)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		res, err := bl.GetProspectDetail(ctx, orgID, leadID)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func makeEnrollProspectEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.EnrollProspectRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		err = bl.EnrollProspect(ctx, orgID, req.CampaignID, req.LeadID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil
	}
}

func makeUpdateProspectEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateProspectRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID
		err = bl.UpdateProspect(ctx, req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil
	}
}

func makePauseProspectEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateProspectRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		err = bl.PauseProspect(ctx, orgID, req.LeadID, req.CampaignID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil
	}
}

func makeResumeProspectEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateProspectRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		err = bl.ResumeProspect(ctx, orgID, req.LeadID, req.CampaignID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil
	}
}

func makeStopProspectEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateProspectRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		err = bl.StopProspect(ctx, orgID, req.LeadID, req.CampaignID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil
	}
}

func makeGetFollowUpsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		filter := request.(string)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		res, err := bl.GetFollowUps(ctx, orgID, filter)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func makeCancelFollowUpEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		id := request.(int64)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		err = bl.CancelFollowUp(ctx, orgID, id)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil
	}
}

func makeRescheduleFollowUpEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.RescheduleFollowUpRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID
		err = bl.RescheduleFollowUp(ctx, req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil
	}
}

