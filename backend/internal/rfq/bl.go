package rfq

import (
	"context"
	"fmt"
	"time"

	"github.com/freel/backend/internal/carrier"
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
	// GetCarrierRates fetches ranked carrier rate options for a given RFQ.
	// It reads the RFQ's origin/destination/target_date and delegates to the carrier service.
	GetCarrierRates(ctx context.Context, orgID, rfqID int32) (*carrier.FetchRatesResponse, error)
	// ApproveQuote marks the selected quote as APPROVED and advances the RFQ to QUOTE_SENT.
	// This is the "Approve & Send" action in the Pricing Workspace.
	ApproveQuote(ctx context.Context, orgID, rfqID, quoteID int32) (*spec.RFQ, error)
}

type businessLogic struct {
	dl             Datalayer
	eventBus       events.Bus
	// carrierService is used by GetCarrierRates to fetch and rank carrier options.
	carrierService carrier.Service
}

func NewBusinessLogic(dl Datalayer, eventBus events.Bus, carrierSvc carrier.Service) BusinessLogic {
	return &businessLogic{
		dl:             dl,
		eventBus:       eventBus,
		carrierService: carrierSvc,
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

// GetCarrierRates fetches carrier rate options for the given RFQ.
// It resolves the RFQ's origin, destination, and target_date, then delegates
// to the carrier service which calls the FF partner API (or mock in dev).
func (b *businessLogic) GetCarrierRates(ctx context.Context, orgID, rfqID int32) (*carrier.FetchRatesResponse, error) {
	// Load the RFQ to get route information
	rfq, err := b.dl.GetRFQByID(ctx, orgID, rfqID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
	}

	// Both origin and destination are required to fetch rates
	if rfq.Origin == nil || rfq.Destination == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	// ── CARGO LOGISTICS & CARRIER SHEET INTEGRATION ────────────────────────────
	// We extract operational variables directly from the RFQ cargo details
	// to feed them to the carrier API comparison system.
	incoterms := ""
	if rfq.Incoterms != nil {
		incoterms = *rfq.Incoterms
	}

	grossWeight := 0.0
	volumeCBM := 0.0
	commodity := ""

	// We calculate the gross cargo weight and cubic volume directly by summing
	// all cargo line items associated with this RFQ.
	if len(rfq.Items) > 0 {
		commodity = rfq.Items[0].Description // Represent commodity using the first line-item's name
		for _, item := range rfq.Items {
			if item.WeightKG != nil {
				grossWeight += *item.WeightKG
			}
			if item.VolumeCBM != nil {
				volumeCBM += *item.VolumeCBM
			}
		}
	}

	resp, err := b.carrierService.FetchRates(ctx, *rfq.Origin, *rfq.Destination, rfq.TargetDate, incoterms, grossWeight, volumeCBM, commodity)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	return resp, nil
}

// ApproveQuote marks the selected quote as APPROVED and advances the RFQ stage to QUOTE_SENT.
// This represents the Pricing Manager's action in the Pricing Workspace: clicking "Approve & Send".
func (b *businessLogic) ApproveQuote(ctx context.Context, orgID, rfqID, quoteID int32) (*spec.RFQ, error) {
	// ── STEP 1: SECURITY CHECK (MULTI-TENANT ENFORCEMENT) ──────────────────────
	// Before making any changes, fetch the RFQ using both the orgID (from the authenticated user)
	// and the rfqID. This guarantees a user from Organization A cannot view or approve quotes 
	// belonging to Organization B. If the record isn't found under this orgID, we return a 404.
	rfq, err := b.dl.GetRFQByID(ctx, orgID, rfqID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
	}

	// ── STEP 2: DATABASE UPDATE (TRANSACTIONAL APPROVAL) ────────────────────────
	// Update the database records. Under the hood, this executes a SQL transaction that:
	//   a) Sets the status of the selected quote (quoteID) to 'APPROVED'.
	//   b) Sets all other alternate carrier quotes for this RFQ to 'REJECTED'.
	// We run this as a single database transaction to prevent partial updates.
	if err := b.dl.ApproveQuote(ctx, rfqID, quoteID); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	// ── STEP 3: STATE MACHINE TRANSITION (STAGE ADVANCEMENT) ─────────────────────
	// Advance the RFQ's operational lifecycle stage to 'QUOTE_SENT'.
	// The AdvanceStage function updates the status column in the database and prepares
	// the state machine to notify downstream sales operations.
	rfq, err = b.AdvanceStage(ctx, orgID, rfqID, spec.StageQuoteSent)
	if err != nil {
		return nil, err
	}

	// ── STEP 4: ASYNCHRONOUS EVENT DISPATCHING (EVENT BUS) ──────────────────────
	// Publish the EventQuoteSent message onto the event bus.
	// This lets other decoupled system modules know the quote is out. For example:
	//   - The Notification Service will capture this and send an email/WhatsApp to the client.
	//   - The Audit Logger will write a permanent activity trace.
	//   - The Outreach Bot can start scheduling follow-up reminders.
	b.eventBus.Publish(events.Event{
		Type: events.EventQuoteSent,
		Payload: map[string]interface{}{
			"rfq_id":   rfqID,
			"quote_id": quoteID,
			"org_id":   orgID,
		},
		Timestamp: time.Now(),
	})

	// ── STEP 5: RETURN RESPONSE ────────────────────────────────────────────────
	// Return the updated RFQ struct back to the transport (HTTP) layer so it can
	// be serialized to JSON and displayed immediately to the user in the React frontend.
	return rfq, nil
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
