package rfq

// This rfq business-logic file is currently connecting your Sales/RFQ flow, carrier-rate flow,
// quote lifecycle, and basic AI extraction together; your next architectural step is to move the
// reasoning-heavy multi-step coordination into LangGraph while keeping these Go business services
// as the controlled execution layer.

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/freel/backend/internal/ai"
	"github.com/freel/backend/internal/carrier"
	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/rates"
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
	ParseShipmentRequest(ctx context.Context, rawText string) (*spec.ParseShipmentResponse, error)
	CreateAITask(ctx context.Context, orgID int64, entityType string, entityID string, taskType string, payload map[string]interface{}) error
}

type businessLogic struct {
	dl            Datalayer
	eventBus      events.Bus
	rateSvc       rates.Service
	aiGateway     ai.Gateway
	promptManager ai.PromptManager
}

func NewBusinessLogic(dl Datalayer, eventBus events.Bus, rateSvc rates.Service, aiGateway ai.Gateway, promptManager ai.PromptManager) BusinessLogic {
	return &businessLogic{
		dl:            dl,
		eventBus:      eventBus,
		rateSvc:       rateSvc,
		aiGateway:     aiGateway,
		promptManager: promptManager,
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
		Payload:   map[string]interface{}{"rfq_id": rfq.ID, "new_stage": newStage, "org_id": orgID},
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

	incoterms := ""
	if rfq.Incoterms != nil {
		incoterms = *rfq.Incoterms
	}

	rateResult, err := b.rateSvc.SearchRates(ctx, rates.RateQuery{
		OrgID:           int64(orgID),
		OriginPort:      *rfq.Origin,
		DestinationPort: *rfq.Destination,
		EquipmentType:   "40GP",
		TargetDate:      rfq.TargetDate,
		Incoterms:       incoterms,
	})
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	richRates := make([]carrier.RichCarrierRate, 0, len(rateResult.Rates))
	for i, r := range rateResult.Rates {
		meetsDeadline := true
		deadlineStatus := "on_time"

		if rfq.TargetDate != nil && r.TransitDays != nil {
			daysUntilTarget := int(time.Until(*rfq.TargetDate).Hours() / 24)
			if *r.TransitDays > daysUntilTarget {
				meetsDeadline = false
				deadlineStatus = "missed"
			} else if daysUntilTarget-*r.TransitDays <= 3 {
				deadlineStatus = "borderline"
			}
		}

		sellPrice := r.TotalBuyPrice * 1.2
		marginPct := 20.0

		transitDaysVal := 0
		if r.TransitDays != nil {
			transitDaysVal = *r.TransitDays
		}

		isRecommended := (i == rateResult.RecommendedIdx)

		richRates = append(richRates, carrier.RichCarrierRate{
			CarrierName:           r.CarrierName,
			BuyPrice:              r.TotalBuyPrice,
			SellPrice:             sellPrice,
			MarginPct:             marginPct,
			TransitDays:           transitDaysVal,
			ReliabilityScore:      r.ConfidenceScore,
			HistoricalSuccessRate: float64(r.ConfidenceScore),
			IsRecommended:         isRecommended,
			AIReasoning:           "",
			MeetsDeadline:         meetsDeadline,
			DeadlineStatus:        deadlineStatus,
			FreeDays:              r.FreeDaysDestination,
			VesselName:            r.VesselName,
			ServiceCode:           r.ServiceCode,
			ViaPort:               r.ViaPort,
			CO2Emissions:          r.CO2PerTEU,
			NauticalMiles:         r.NauticalMiles,
			OceanFreight:          r.OceanFreight,
			OriginCharges:         r.OriginCharges,
			DestinationCharges:    r.DestinationCharges,
			FetchedAt:             rateResult.SearchedAt.Format(time.RFC3339),
		})
	}

	// Update recommended AI reasonings in the first item if there are rates
	if len(richRates) > 0 && rateResult.RecommendedIdx < len(richRates) {
		richRates[rateResult.RecommendedIdx].AIReasoning = rateResult.OverallReasoning
	}

	return &carrier.FetchRatesResponse{
		Rates:            richRates,
		OverallReasoning: rateResult.OverallReasoning,
		RecommendedIdx:   rateResult.RecommendedIdx,
		FetchedAt:        rateResult.SearchedAt.Format(time.RFC3339),
	}, nil
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

// ParseShipmentRequest parses unstructured shipment requests using AI Gateway.
//
// Simple meaning:
//
//	Constructs the extract_shipment_request prompt template, executes it via
//	the LLM API, parses the JSON payload, and returns the structured extraction result.
func (b *businessLogic) ParseShipmentRequest(ctx context.Context, rawText string) (*spec.ParseShipmentResponse, error) {
	if rawText == "" {
		return nil, fmt.Errorf("raw text is required for AI parsing")
	}

	// 1. Fetch and format the AI extraction prompt template
	prompt, err := b.promptManager.GetPrompt("extract_shipment_request", map[string]interface{}{
		"RawText": rawText,
	})
	if err != nil {
		return nil, fmt.Errorf("prompt format: %w", err)
	}

	// 2. Call the AI Gateway to execute prompt and get structured JSON output
	aiResponseStr, err := b.aiGateway.ExecutePrompt(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("ai execution: %w", err)
	}

	// 3. Deserialize the structured LLM output
	var result struct {
		Data struct {
			Origin      *string `json:"origin"`
			Destination *string `json:"destination"`
			Incoterms   *string `json:"incoterms"`
			Weight      *string `json:"weight"`
			Volume      *string `json:"volume"`
		} `json:"data"`
		ConfidenceScore int      `json:"confidence_score"`
		MissingFields   []string `json:"missing_fields"`
	}

	if err := json.Unmarshal([]byte(aiResponseStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON response: %w (raw response: %s)", err, aiResponseStr)
	}

	// 4. Return formatted response matching react frontend expectations
	return &spec.ParseShipmentResponse{
		Data: map[string]interface{}{
			"data": map[string]interface{}{
				"origin":      result.Data.Origin,
				"destination": result.Data.Destination,
				"incoterms":   result.Data.Incoterms,
				"weight":      result.Data.Weight,
				"volume":      result.Data.Volume,
			},
			"confidence_score": result.ConfidenceScore,
			"missing_fields":   result.MissingFields,
		},
	}, nil
}

func (b *businessLogic) CreateAITask(ctx context.Context, orgID int64, entityType string, entityID string, taskType string, payload map[string]interface{}) error {
	return b.dl.CreateAITask(ctx, orgID, entityType, entityID, taskType, payload)
}
