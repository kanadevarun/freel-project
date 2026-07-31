package leads

import (
	"context"
	"database/sql"
	"errors"

	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/leads/spec"
	"github.com/freel/backend/internal/svcerror"
)

type BusinessLogic interface {
	CreateLead(ctx context.Context, req spec.CreateLeadRequest) (*spec.Lead, error)
	BulkCreate(ctx context.Context, req spec.ImportLeadsRequest) (int, error)
	GetLead(ctx context.Context, orgID int32, id int32) (*spec.Lead, error)
	ListLeads(ctx context.Context, req spec.ListLeadsRequest) (*spec.ListLeadsResponse, error)
	UpdateLead(ctx context.Context, req spec.UpdateLeadRequest) (*spec.Lead, error)
	DeleteLead(ctx context.Context, req spec.DeleteLeadRequest) error
}

type businessLogic struct {
	dl       Datalayer
	eventBus events.Bus
}

func NewBusinessLogic(dl Datalayer, eventBus events.Bus) BusinessLogic {
	return &businessLogic{
		dl:       dl,
		eventBus: eventBus,
	}
}

func (b *businessLogic) CreateLead(ctx context.Context, req spec.CreateLeadRequest) (*spec.Lead, error) {
	if req.CompanyName == "" {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	lead := &spec.Lead{
		OrgID:       req.OrgID,
		CompanyName: req.CompanyName,
		ContactName: req.ContactName,
		Email:       req.Email,
		Phone:       req.Phone,
		Status:      "NEW",
		Source:      req.Source,
	}

	err := b.dl.Create(ctx, lead)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	// Publish Event
	b.eventBus.Publish(events.Event{
		Type: events.EventLeadCreated,
		Payload: map[string]interface{}{
			"lead_id": lead.ID,
			"org_id":  lead.OrgID,
		},
	})

	return lead, nil
}

func (b *businessLogic) BulkCreate(ctx context.Context, req spec.ImportLeadsRequest) (int, error) {
	createdCount := 0
	for _, lReq := range req.Leads {
		lReq.OrgID = req.OrgID
		_, err := b.CreateLead(ctx, *lReq)
		if err == nil {
			createdCount++
		}
	}
	return createdCount, nil
}

func (b *businessLogic) GetLead(ctx context.Context, orgID int32, id int32) (*spec.Lead, error) {
	lead, err := b.dl.GetByID(ctx, orgID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return lead, nil
}

func (b *businessLogic) ListLeads(ctx context.Context, req spec.ListLeadsRequest) (*spec.ListLeadsResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	data, total, err := b.dl.List(ctx, req.OrgID, req.Limit, req.Offset, req.Status)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	if data == nil {
		data = []*spec.Lead{}
	}

	return &spec.ListLeadsResponse{
		Data:       data,
		TotalCount: total,
	}, nil
}

func (b *businessLogic) UpdateLead(ctx context.Context, req spec.UpdateLeadRequest) (*spec.Lead, error) {
	lead, err := b.dl.GetByID(ctx, req.OrgID, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	if req.CompanyName != nil {
		lead.CompanyName = *req.CompanyName
	}
	if req.ContactName != nil {
		lead.ContactName = req.ContactName
	}
	if req.Email != nil {
		lead.Email = req.Email
	}
	if req.Phone != nil {
		lead.Phone = req.Phone
	}
	if req.Status != nil {
		lead.Status = *req.Status
	}
	if req.Source != nil {
		lead.Source = req.Source
	}
	if req.AIScore != nil {
		lead.AIScore = *req.AIScore
	}
	if req.AIResearchReport != nil {
		lead.AIResearchReport = req.AIResearchReport
	}

	err = b.dl.Update(ctx, lead)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	return lead, nil
}

func (b *businessLogic) DeleteLead(ctx context.Context, req spec.DeleteLeadRequest) error {
	err := b.dl.Delete(ctx, req.OrgID, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}
