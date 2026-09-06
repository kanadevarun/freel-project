package outreach

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/outreach/spec"
	"github.com/freel/backend/internal/svcerror"
)

type BusinessLogic interface {
	CreateCampaign(ctx context.Context, req spec.CreateCampaignRequest) (*spec.Campaign, error)
	GetCampaign(ctx context.Context, orgID, id int32) (*spec.Campaign, error)
	ListCampaigns(ctx context.Context, req spec.ListCampaignsRequest) (*spec.ListCampaignsResponse, error)
	ActivateCampaign(ctx context.Context, req spec.ActivateCampaignRequest) (*spec.Campaign, error)
	PauseCampaign(ctx context.Context, req spec.PauseCampaignRequest) (*spec.Campaign, error)
	DeleteCampaign(ctx context.Context, req spec.DeleteCampaignRequest) error

	// New methods
	GetCampaignSequence(ctx context.Context, orgID, campaignID int32) ([]*spec.Sequence, error)
	AddSequenceStep(ctx context.Context, req spec.AddSequenceStepRequest) (*spec.Sequence, error)
	UpdateSequenceStep(ctx context.Context, req spec.UpdateSequenceStepRequest) (*spec.Sequence, error)
	DeleteSequenceStep(ctx context.Context, req spec.DeleteSequenceStepRequest) error
	ReorderSequence(ctx context.Context, req spec.ReorderSequenceRequest) error

	GetCampaignAudience(ctx context.Context, orgID, campaignID int32) ([]*spec.AudienceLead, error)
	AddCampaignAudience(ctx context.Context, req spec.AddCampaignAudienceRequest) error
	RemoveCampaignAudience(ctx context.Context, req spec.RemoveCampaignAudienceRequest) error

	GetOutreachAnalytics(ctx context.Context, orgID int32) (*spec.OutreachDashboardResponse, error)
	GetCampaignAnalytics(ctx context.Context, orgID, campaignID int32) (*spec.OutreachAnalyticsResponse, error)
	GetCampaignLeads(ctx context.Context, orgID, campaignID int32) ([]*spec.GeneratedLead, error)
	GetCampaignInsights(ctx context.Context, orgID, campaignID int32) ([]spec.CampaignInsight, error)
	GetConversionFunnel(ctx context.Context, orgID int32) (*spec.ConversionFunnelResponse, error)

	// Activity CRUD Methods
	CreateActivity(ctx context.Context, req *spec.CreateActivityRequest) (int64, error)
	GetActivity(ctx context.Context, orgID int32, id int64) (*spec.OutreachActivityDetail, error)
	UpdateActivity(ctx context.Context, req *spec.UpdateActivityRequest) error
	CompleteActivity(ctx context.Context, orgID int32, id int64) error
	DeleteActivity(ctx context.Context, orgID int32, id int64) error

	// Engagement & Prospects Methods
	GetCampaignRecipients(ctx context.Context, orgID, campaignID int32) ([]*spec.CampaignRecipient, error)
	GetCampaignActivity(ctx context.Context, orgID, campaignID int32) ([]*spec.CampaignActivityEvent, error)
	GetProspects(ctx context.Context, orgID int32) ([]*spec.CampaignRecipient, error)
	GetProspectEngagement(ctx context.Context, orgID int32, leadID int64) (*spec.ProspectEngagementResponse, error)
	GetLeadOutreachActivity(ctx context.Context, orgID int32, leadID int64) ([]*spec.CampaignActivityEvent, error)

	GetProspectDetail(ctx context.Context, orgID int32, leadID int64) (*spec.ProspectDetailResponse, error)
	EnrollProspect(ctx context.Context, orgID int32, campaignID int32, leadID int64) error
	UpdateProspect(ctx context.Context, req *spec.UpdateProspectRequest) error
	PauseProspect(ctx context.Context, orgID int32, leadID int64, campaignID int32) error
	ResumeProspect(ctx context.Context, orgID int32, leadID int64, campaignID int32) error
	StopProspect(ctx context.Context, orgID int32, leadID int64, campaignID int32) error
	GetFollowUps(ctx context.Context, orgID int32, filter string) ([]*spec.OutreachActivityDetail, error)
	CancelFollowUp(ctx context.Context, orgID int32, id int64) error
	RescheduleFollowUp(ctx context.Context, req *spec.RescheduleFollowUpRequest) error
}

type businessLogic struct {
	dl Datalayer
}

func NewBusinessLogic(dl Datalayer) BusinessLogic {
	return &businessLogic{dl: dl}
}

func (b *businessLogic) CreateCampaign(ctx context.Context, req spec.CreateCampaignRequest) (*spec.Campaign, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	c := &spec.Campaign{
		OrgID:  req.OrgID,
		Name:   strings.TrimSpace(req.Name),
		Status: spec.CampaignStatusDraft,
	}

	if err := b.dl.CreateCampaign(ctx, c); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        int64(c.OrgID),
		Action:       domain.ActionCreate,
		Module:       domain.ModuleOutreach,
		ResourceType: "CAMPAIGN",
		ResourceID:   fmt.Sprintf("%d", c.ID),
		ResourceName: c.Name,
		Description:  fmt.Sprintf("Created outreach campaign %s", c.Name),
		Result:       domain.ResultSuccess,
	})

	return c, nil
}

func (b *businessLogic) GetCampaign(ctx context.Context, orgID, id int32) (*spec.Campaign, error) {
	c, err := b.dl.GetCampaignByID(ctx, orgID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return c, nil
}

func (b *businessLogic) ListCampaigns(ctx context.Context, req spec.ListCampaignsRequest) (*spec.ListCampaignsResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	campaigns, total, err := b.dl.ListCampaigns(ctx, req.OrgID, req.Limit, req.Offset)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	if campaigns == nil {
		campaigns = []*spec.Campaign{}
	}
	return &spec.ListCampaignsResponse{Data: campaigns, TotalCount: total}, nil
}

func (b *businessLogic) ActivateCampaign(ctx context.Context, req spec.ActivateCampaignRequest) (*spec.Campaign, error) {
	c, err := b.dl.GetCampaignByID(ctx, req.OrgID, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	if c.Status != spec.CampaignStatusDraft && c.Status != spec.CampaignStatusPaused {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	c.Status = spec.CampaignStatusActive
	if err := b.dl.UpdateCampaign(ctx, c); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        int64(c.OrgID),
		Action:       domain.ActionEnable,
		Module:       domain.ModuleOutreach,
		ResourceType: "CAMPAIGN",
		ResourceID:   fmt.Sprintf("%d", c.ID),
		ResourceName: c.Name,
		Description:  fmt.Sprintf("Activated outreach campaign %s", c.Name),
		Result:       domain.ResultSuccess,
	})

	return c, nil
}

func (b *businessLogic) PauseCampaign(ctx context.Context, req spec.PauseCampaignRequest) (*spec.Campaign, error) {
	c, err := b.dl.GetCampaignByID(ctx, req.OrgID, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	if c.Status != spec.CampaignStatusActive {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	c.Status = spec.CampaignStatusPaused
	if err := b.dl.UpdateCampaign(ctx, c); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        int64(c.OrgID),
		Action:       domain.ActionDisable,
		Module:       domain.ModuleOutreach,
		ResourceType: "CAMPAIGN",
		ResourceID:   fmt.Sprintf("%d", c.ID),
		ResourceName: c.Name,
		Description:  fmt.Sprintf("Paused outreach campaign %s", c.Name),
		Result:       domain.ResultSuccess,
	})

	return c, nil
}

func (b *businessLogic) DeleteCampaign(ctx context.Context, req spec.DeleteCampaignRequest) error {
	err := b.dl.DeleteCampaign(ctx, req.OrgID, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        int64(req.OrgID),
		Action:       domain.ActionDelete,
		Module:       domain.ModuleOutreach,
		ResourceType: "CAMPAIGN",
		ResourceID:   fmt.Sprintf("%d", req.ID),
		Description:  fmt.Sprintf("Deleted campaign #%d", req.ID),
		Result:       domain.ResultSuccess,
	})

	return nil
}

func (b *businessLogic) GetCampaignSequence(ctx context.Context, orgID, campaignID int32) ([]*spec.Sequence, error) {
	steps, err := b.dl.GetCampaignSequence(ctx, orgID, campaignID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return steps, nil
}

func (b *businessLogic) AddSequenceStep(ctx context.Context, req spec.AddSequenceStepRequest) (*spec.Sequence, error) {
	s := &spec.Sequence{
		CampaignID: req.CampaignID,
		Channel:    req.Channel,
		Name:       req.Name,
		Subject:    req.Subject,
		Body:       req.Body,
		Template:   req.Body,
		DelayDays:  req.DelayDays,
	}
	if s.Channel == "" {
		s.Channel = "EMAIL"
	}
	err := b.dl.AddSequenceStep(ctx, s)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return s, nil
}

func (b *businessLogic) UpdateSequenceStep(ctx context.Context, req spec.UpdateSequenceStepRequest) (*spec.Sequence, error) {
	s := &spec.Sequence{
		ID:         req.StepID,
		CampaignID: req.CampaignID,
		Channel:    req.Channel,
		Name:       req.Name,
		Subject:    req.Subject,
		Body:       req.Body,
		Template:   req.Body,
		DelayDays:  req.DelayDays,
	}
	err := b.dl.UpdateSequenceStep(ctx, s)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return s, nil
}

func (b *businessLogic) DeleteSequenceStep(ctx context.Context, req spec.DeleteSequenceStepRequest) error {
	err := b.dl.DeleteSequenceStep(ctx, req.CampaignID, req.StepID)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) ReorderSequence(ctx context.Context, req spec.ReorderSequenceRequest) error {
	err := b.dl.ReorderSequence(ctx, req.CampaignID, req.StepIDs)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) GetCampaignAudience(ctx context.Context, orgID, campaignID int32) ([]*spec.AudienceLead, error) {
	leads, err := b.dl.GetCampaignAudience(ctx, orgID, campaignID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return leads, nil
}

func (b *businessLogic) AddCampaignAudience(ctx context.Context, req spec.AddCampaignAudienceRequest) error {
	err := b.dl.AddCampaignAudience(ctx, req.CampaignID, req.LeadIDs)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) RemoveCampaignAudience(ctx context.Context, req spec.RemoveCampaignAudienceRequest) error {
	err := b.dl.RemoveCampaignAudience(ctx, req.CampaignID, req.LeadID)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) GetOutreachAnalytics(ctx context.Context, orgID int32) (*spec.OutreachDashboardResponse, error) {
	res, err := b.dl.GetOutreachAnalytics(ctx, orgID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return res, nil
}

func (b *businessLogic) GetCampaignAnalytics(ctx context.Context, orgID, campaignID int32) (*spec.OutreachAnalyticsResponse, error) {
	res, err := b.dl.GetCampaignAnalytics(ctx, orgID, campaignID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return res, nil
}

func (b *businessLogic) GetCampaignLeads(ctx context.Context, orgID, campaignID int32) ([]*spec.GeneratedLead, error) {
	leads, err := b.dl.GetCampaignLeads(ctx, orgID, campaignID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return leads, nil
}

func (b *businessLogic) GetCampaignInsights(ctx context.Context, orgID, campaignID int32) ([]spec.CampaignInsight, error) {
	insights, err := b.dl.GetCampaignInsights(ctx, orgID, campaignID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return insights, nil
}

func (b *businessLogic) GetConversionFunnel(ctx context.Context, orgID int32) (*spec.ConversionFunnelResponse, error) {
	funnel, err := b.dl.GetConversionFunnel(ctx, orgID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return funnel, nil
}

func (b *businessLogic) CreateActivity(ctx context.Context, req *spec.CreateActivityRequest) (int64, error) {
	id, err := b.dl.CreateActivity(ctx, req)
	if err != nil {
		return 0, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return id, nil
}

func (b *businessLogic) GetActivity(ctx context.Context, orgID int32, id int64) (*spec.OutreachActivityDetail, error) {
	act, err := b.dl.GetActivity(ctx, orgID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return act, nil
}

func (b *businessLogic) UpdateActivity(ctx context.Context, req *spec.UpdateActivityRequest) error {
	err := b.dl.UpdateActivity(ctx, req)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) CompleteActivity(ctx context.Context, orgID int32, id int64) error {
	err := b.dl.CompleteActivity(ctx, orgID, id)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) DeleteActivity(ctx context.Context, orgID int32, id int64) error {
	err := b.dl.DeleteActivity(ctx, orgID, id)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) GetCampaignRecipients(ctx context.Context, orgID, campaignID int32) ([]*spec.CampaignRecipient, error) {
	recipients, err := b.dl.GetCampaignRecipients(ctx, orgID, campaignID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return recipients, nil
}

func (b *businessLogic) GetCampaignActivity(ctx context.Context, orgID, campaignID int32) ([]*spec.CampaignActivityEvent, error) {
	events, err := b.dl.GetCampaignActivity(ctx, orgID, campaignID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return events, nil
}

func (b *businessLogic) GetProspects(ctx context.Context, orgID int32) ([]*spec.CampaignRecipient, error) {
	prospects, err := b.dl.GetProspects(ctx, orgID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return prospects, nil
}

func (b *businessLogic) GetProspectEngagement(ctx context.Context, orgID int32, leadID int64) (*spec.ProspectEngagementResponse, error) {
	resp, err := b.dl.GetProspectEngagement(ctx, orgID, leadID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return resp, nil
}

func (b *businessLogic) GetLeadOutreachActivity(ctx context.Context, orgID int32, leadID int64) ([]*spec.CampaignActivityEvent, error) {
	timeline, err := b.dl.GetLeadOutreachActivity(ctx, orgID, leadID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return timeline, nil
}

func (b *businessLogic) GetProspectDetail(ctx context.Context, orgID int32, leadID int64) (*spec.ProspectDetailResponse, error) {
	detail, err := b.dl.GetProspectDetail(ctx, orgID, leadID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return detail, nil
}

func (b *businessLogic) EnrollProspect(ctx context.Context, orgID int32, campaignID int32, leadID int64) error {
	err := b.dl.EnrollProspect(ctx, orgID, campaignID, leadID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) UpdateProspect(ctx context.Context, req *spec.UpdateProspectRequest) error {
	err := b.dl.UpdateProspect(ctx, req)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) PauseProspect(ctx context.Context, orgID int32, leadID int64, campaignID int32) error {
	err := b.dl.PauseProspect(ctx, orgID, leadID, campaignID)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) ResumeProspect(ctx context.Context, orgID int32, leadID int64, campaignID int32) error {
	err := b.dl.ResumeProspect(ctx, orgID, leadID, campaignID)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) StopProspect(ctx context.Context, orgID int32, leadID int64, campaignID int32) error {
	err := b.dl.StopProspect(ctx, orgID, leadID, campaignID)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) GetFollowUps(ctx context.Context, orgID int32, filter string) ([]*spec.OutreachActivityDetail, error) {
	activities, err := b.dl.GetFollowUps(ctx, orgID, filter)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return activities, nil
}

func (b *businessLogic) CancelFollowUp(ctx context.Context, orgID int32, id int64) error {
	err := b.dl.CancelFollowUp(ctx, orgID, id)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) RescheduleFollowUp(ctx context.Context, req *spec.RescheduleFollowUpRequest) error {
	err := b.dl.RescheduleFollowUp(ctx, req)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

