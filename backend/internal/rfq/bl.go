package rfq

import (
	"context"
	"fmt"
	"time"

	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/rfq/spec"
	"github.com/freel/backend/internal/svcerror"
)

type BusinessLogic interface {
	CreateRFQ(ctx context.Context, req spec.CreateRFQRequest) (*spec.RFQ, error)
	GetRFQ(ctx context.Context, orgID, rfqID int32) (*spec.RFQ, error)
	ListRFQs(ctx context.Context, req spec.ListRFQsRequest) (*spec.ListRFQsResponse, error)
	AdvanceStage(ctx context.Context, orgID, rfqID int32, newStage string) (*spec.RFQ, error)
	AddQuote(ctx context.Context, orgID int32, quote *spec.Quote) error
	UpdateAgentStatus(ctx context.Context, orgID, rfqID int32, status string) error
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

func (b *businessLogic) CreateRFQ(ctx context.Context, req spec.CreateRFQRequest) (*spec.RFQ, error) {
	if req.OrgID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	rfqNumber := fmt.Sprintf("RFQ-%s", time.Now().Format("20060102-150405"))

	rfq := &spec.RFQ{
		OrgID:       req.OrgID,
		RFQNumber:   rfqNumber,
		CustomerID:  req.CustomerID,
		Stage:       spec.StageRFQCreated,
		Origin:      req.Origin,
		Destination: req.Destination,
		Incoterms:   req.Incoterms,
		TargetDate:  req.TargetDate,
	}

	rfq.HealthScore = b.calculateHealthScore(rfq)

	if err := b.dl.CreateRFQ(ctx, rfq); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	for _, reqItem := range req.Items {
		item := reqItem
		item.RFQID = rfq.ID
		if err := b.dl.CreateRFQItem(ctx, &item); err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}
		rfq.Items = append(rfq.Items, item)
	}

	b.eventBus.Publish(events.Event{
		Type:      events.EventRFQCreated,
		Payload:   map[string]interface{}{"rfq_id": rfq.ID, "org_id": rfq.OrgID},
		Timestamp: time.Now(),
	})

	return rfq, nil
}

func (b *businessLogic) GetRFQ(ctx context.Context, orgID, rfqID int32) (*spec.RFQ, error) {
	rfq, err := b.dl.GetRFQByID(ctx, orgID, rfqID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
	}
	return rfq, nil
}

func (b *businessLogic) ListRFQs(ctx context.Context, req spec.ListRFQsRequest) (*spec.ListRFQsResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 50
	}
	rfqs, total, err := b.dl.ListRFQs(ctx, req.OrgID, req.Limit, req.Offset)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	if rfqs == nil {
		rfqs = []spec.RFQ{}
	}

	return &spec.ListRFQsResponse{
		Data:       rfqs,
		TotalCount: total,
	}, nil
}

func (b *businessLogic) AdvanceStage(ctx context.Context, orgID, rfqID int32, newStage string) (*spec.RFQ, error) {
	rfq, err := b.dl.GetRFQByID(ctx, orgID, rfqID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
	}

	validStages := map[string]bool{
		spec.StageRFQCreated:      true,
		spec.StagePricingAssigned: true,
		spec.StageQuoteGenerated:  true,
		spec.StageQuoteSent:       true,
		spec.StageNegotiation:     true,
		spec.StageWon:             true,
		spec.StageLost:            true,
		spec.StageShipmentCreated: true,
	}

	if !validStages[newStage] {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	if err := b.dl.UpdateStage(ctx, orgID, rfqID, newStage); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	rfq.Stage = newStage

	var eventType events.EventType
	switch newStage {
	case spec.StagePricingAssigned:
		eventType = events.EventRFQAssigned
	case spec.StageQuoteGenerated:
		eventType = events.EventQuoteGenerated
	case spec.StageQuoteSent:
		eventType = events.EventQuoteSent
	case spec.StageWon:
		eventType = events.EventRFQWon
	case spec.StageLost:
		eventType = events.EventRFQLost
	default:
		eventType = events.EventRFQUpdated
	}

	b.eventBus.Publish(events.Event{
		Type:      eventType,
		Payload:   map[string]interface{}{"rfq_id": rfq.ID, "new_stage": newStage},
		Timestamp: time.Now(),
	})

	return rfq, nil
}

func (b *businessLogic) AddQuote(ctx context.Context, orgID int32, quote *spec.Quote) error {
	rfq, err := b.dl.GetRFQByID(ctx, orgID, quote.RFQID)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
	}

	if err := b.dl.CreateQuote(ctx, quote); err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	if rfq.Stage == spec.StageRFQCreated || rfq.Stage == spec.StagePricingAssigned {
		_, _ = b.AdvanceStage(ctx, orgID, rfq.ID, spec.StageQuoteGenerated)
	}

	b.eventBus.Publish(events.Event{
		Type:      events.EventQuoteGenerated,
		Payload:   map[string]interface{}{"rfq_id": rfq.ID, "quote_id": quote.ID},
		Timestamp: time.Now(),
	})

	return nil
}

func (b *businessLogic) UpdateAgentStatus(ctx context.Context, orgID, rfqID int32, status string) error {
	if err := b.dl.UpdateAgentStatus(ctx, orgID, rfqID, status); err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) calculateHealthScore(rfq *spec.RFQ) int {
	score := 100
	if rfq.Origin == nil || *rfq.Origin == "" {
		score -= 20
	}
	if rfq.Destination == nil || *rfq.Destination == "" {
		score -= 20
	}
	if rfq.Incoterms == nil || *rfq.Incoterms == "" {
		score -= 10
	}
	if rfq.TargetDate == nil {
		score -= 10
	}
	
	if len(rfq.Items) == 0 {
		score -= 20
	}
	
	if score < 0 {
		return 0
	}
	return score
}
