package outreach

import (
	"context"
	"database/sql"
	"errors"
	"strings"

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
	return nil
}
