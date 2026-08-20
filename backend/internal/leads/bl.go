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
	GetLeadByEmail(ctx context.Context, orgID int32, email string) (*spec.Lead, error)
	LogInteraction(ctx context.Context, orgID int32, inter *LeadInteraction) error
	ListInteractions(ctx context.Context, orgID int32, leadID int32) ([]*LeadInteraction, error)
	FindByThreadID(ctx context.Context, orgID int32, threadID string) ([]*LeadInteraction, error)
	CreateAITask(ctx context.Context, orgID int64, entityType, entityID, taskType string, payload map[string]interface{}) error
	UpdateInteractionAI(ctx context.Context, orgID int64, id int64, intent string, sentiment string, confidence int, linkedRFQID *int64, aiSummary string, draftedReply string) error
	UpdateInteractionContext(ctx context.Context, orgID int64, id int64, partialCtx map[string]interface{}) error
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
		oldStatus := lead.Status
		lead.Status = *req.Status

		if oldStatus != "CONVERTED" && lead.Status == "CONVERTED" {
			_ = b.dl.EnsureCustomerForLead(ctx, lead)
		}
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

func (b *businessLogic) GetLeadByEmail(ctx context.Context, orgID int32, email string) (*spec.Lead, error) {
	lead, err := b.dl.GetByEmail(ctx, orgID, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return lead, nil
}

func (b *businessLogic) LogInteraction(ctx context.Context, orgID int32, inter *LeadInteraction) error {
	inter.OrgID = int64(orgID)
	err := b.dl.LogInteraction(ctx, inter)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) ListInteractions(ctx context.Context, orgID int32, leadID int32) ([]*LeadInteraction, error) {
	list, err := b.dl.ListInteractions(ctx, orgID, leadID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return list, nil
}

func (b *businessLogic) FindByThreadID(ctx context.Context, orgID int32, threadID string) ([]*LeadInteraction, error) {
	list, err := b.dl.FindByThreadID(ctx, orgID, threadID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return list, nil
}

func (b *businessLogic) CreateAITask(ctx context.Context, orgID int64, entityType, entityID, taskType string, payload map[string]interface{}) error {
	err := b.dl.CreateAITask(ctx, orgID, entityType, entityID, taskType, payload)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) UpdateInteractionAI(ctx context.Context, orgID int64, id int64, intent string, sentiment string, confidence int, linkedRFQID *int64, aiSummary string, draftedReply string) error {
	err := b.dl.UpdateInteractionAI(ctx, orgID, id, intent, sentiment, confidence, linkedRFQID, aiSummary, draftedReply)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) UpdateInteractionContext(ctx context.Context, orgID int64, id int64, partialCtx map[string]interface{}) error {
	err := b.dl.UpdateInteractionContext(ctx, orgID, id, partialCtx)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}
