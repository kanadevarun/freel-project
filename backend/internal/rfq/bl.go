package rfq

// This rfq business-logic file is currently connecting your Sales/RFQ flow, carrier-rate flow,
// quote lifecycle, and basic AI extraction together; your next architectural step is to move the
// reasoning-heavy multi-step coordination into LangGraph while keeping these Go business services
// as the controlled execution layer.

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/freel/backend/internal/ai"
	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/carrier"
	carrierService "github.com/freel/backend/internal/carrier/service"
	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/rates"
	"github.com/freel/backend/internal/rfq/spec"
	"github.com/freel/backend/internal/subscription"
	"github.com/freel/backend/internal/svcerror"
)


type BusinessLogic interface {
	CreateRFQ(ctx context.Context, req spec.CreateRFQRequest) (*spec.RFQ, error)
	GetRFQ(ctx context.Context, orgID, rfqID int32) (*spec.RFQ, error)
	GetRFQByLeadID(ctx context.Context, orgID int32, leadID int64) (*spec.RFQ, error)
	GetTimeline(ctx context.Context, orgID, rfqID int32) ([]spec.TimelineEvent, error)
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
	// GetRequirements evaluates the operational readiness of an RFQ using deterministic
	// business rules. Org isolation is enforced: the RFQ must belong to orgID.
	GetRequirements(ctx context.Context, orgID, rfqID int32) (*spec.GetRequirementsResponse, error)
	// GetActivity aggregates and normalizes the full chronological operational timeline
	// for an RFQ across Lead, interactions, AI tasks, requirements, and quotes.
	GetActivity(ctx context.Context, orgID, rfqID int32) (*spec.GetActivityResponse, error)

	// Documents (Task 12)
	GetDocuments(ctx context.Context, orgID, rfqID int32) (*spec.GetDocumentsResponse, error)
	CreateDocument(ctx context.Context, orgID, rfqID int32, req spec.CreateDocumentRequest, uploader string) (*spec.RFQDocument, error)
	UpdateDocumentStatus(ctx context.Context, orgID, rfqID int32, documentID int64, req spec.UpdateDocumentStatusRequest, reviewer string) (*spec.RFQDocument, error)
	DeleteDocument(ctx context.Context, orgID, rfqID int32, documentID int64) error

	// Quotes & Pricing (Task 13)
	GetQuotes(ctx context.Context, orgID, rfqID int32) (*spec.GetQuotesResponse, error)
	CreateRFQQuote(ctx context.Context, orgID, rfqID int32, req spec.CreateQuoteRequest, creator string) (*spec.RFQQuote, error)
	UpdateRFQQuote(ctx context.Context, orgID, rfqID int32, quoteID int64, req spec.UpdateQuoteRequest, updater string) (*spec.RFQQuote, error)
	UpdateRFQQuoteStatus(ctx context.Context, orgID, rfqID int32, quoteID int64, req spec.UpdateQuoteStatusRequest, updater string) (*spec.RFQQuote, error)
	RecommendRFQQuote(ctx context.Context, orgID, rfqID int32, quoteID int64, recommender string) (*spec.RFQQuote, error)
	ApproveRFQQuote(ctx context.Context, orgID, rfqID int32, quoteID int64, req spec.ApproveRFQQuoteRequest, approver string) (*spec.RFQQuote, error)
	SelectRFQQuoteForCustomer(ctx context.Context, orgID, rfqID int32, quoteID int64, selector string) (*spec.RFQQuote, error)
	DeleteRFQQuote(ctx context.Context, orgID, rfqID int32, quoteID int64) error

	// Bookings & Shipments (Task 14)
	GetBookingHandoff(ctx context.Context, orgID, rfqID int32) (*spec.GetBookingHandoffResponse, error)
	CreateBookingFromRFQ(ctx context.Context, orgID, rfqID int32, req spec.CreateBookingRequest, creator string) (*spec.RFQBooking, error)
	UpdateBookingStatus(ctx context.Context, orgID, rfqID int32, bookingID int64, req spec.UpdateBookingStatusRequest, updater string) (*spec.RFQBooking, error)
	GetShipmentHandoff(ctx context.Context, orgID, rfqID int32) (*spec.GetShipmentHandoffResponse, error)

	// Dedicated Booking Operations Workspace (Task 15)
	GetBookingsWorkspace(ctx context.Context, orgID int32, filter spec.BookingListFilter) (*spec.GetBookingsWorkspaceResponse, error)
	GetBookingWorkspaceDetail(ctx context.Context, orgID int32, bookingID int64) (*spec.BookingDetailResponse, error)
	DirectUpdateBookingStatus(ctx context.Context, orgID int32, bookingID int64, req spec.DirectUpdateBookingStatusRequest, updater string) (*spec.RFQBooking, error)
	GetEligibleRFQsForBooking(ctx context.Context, orgID int32) ([]spec.EligibleBookingRFQ, error)
	CreateShipmentFromBooking(ctx context.Context, orgID int32, bookingID int64, req spec.CreateShipmentFromBookingRequest, creator string) (*spec.RFQShipment, error)

	// Task 5: Live Carrier Booking Integration
	BookWithCarrier(ctx context.Context, orgID int32, bookingID int64, req spec.BookWithCarrierRequest, user string) (*spec.BookingDetailResponse, error)
	SyncCarrierBooking(ctx context.Context, orgID int32, bookingID int64, user string) (*spec.BookingDetailResponse, error)
	SetCarrierIntegrationService(carrierSvc carrierService.CarrierService)
}

type businessLogic struct {
	dl                   Datalayer
	eventBus             events.Bus
	rateSvc              rates.Service
	aiGateway            ai.Gateway
	promptManager        ai.PromptManager
	entitlementSvc       subscription.EntitlementService
	carrierBookingEngine *CarrierBookingEngine
}

func NewBusinessLogic(dl Datalayer, eventBus events.Bus, rateSvc rates.Service, aiGateway ai.Gateway, promptManager ai.PromptManager, entitlementSvc subscription.EntitlementService) BusinessLogic {
	return &businessLogic{
		dl:             dl,
		eventBus:       eventBus,
		rateSvc:        rateSvc,
		aiGateway:      aiGateway,
		promptManager:  promptManager,
		entitlementSvc: entitlementSvc,
	}
}

func (b *businessLogic) SetCarrierIntegrationService(carrierSvc carrierService.CarrierService) {
	b.carrierBookingEngine = NewCarrierBookingEngine(b.dl, carrierSvc)
}

func (b *businessLogic) CreateRFQ(ctx context.Context, req spec.CreateRFQRequest) (*spec.RFQ, error) {
	if req.OrgID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	// 0. Duplicate Protection: Check if an RFQ already exists for this Lead
	if req.LeadID != nil {
		existing, err := b.dl.GetRFQByLeadID(ctx, req.OrgID, *req.LeadID)
		if err == nil && existing != nil {
			log.Printf("[RFQ Service] RFQ already exists for Lead ID %d (idempotency): RFQ %s", *req.LeadID, existing.RFQNumber)
			// Return/reuse the existing RFQ, loading all its items/quotes
			fullRFQ, err := b.dl.GetRFQByID(ctx, req.OrgID, existing.ID)
			if err == nil {
				return fullRFQ, nil
			}
			return existing, nil
		}
	}

	// 1. Enforce Entitlement Limits for RFQs
	if b.entitlementSvc != nil {
		if err := b.entitlementSvc.CheckEntitlement(ctx, int64(req.OrgID), subscription.MetricRFQs); err != nil {
			if err == subscription.ErrLimitReached {
				// Typically we'd return a specific 402 Payment Required or 403 Forbidden.
				// Wrap it in a 403 for now.
				return nil, svcerror.WrapServiceError(svcerror.ErrInsufficientResourceAccess, err)
			}
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}
	}

	rfqNumber := fmt.Sprintf("RFQ-%s-%03d", time.Now().Format("20060102-150405"), time.Now().Nanosecond()/1000000)

	rfq := &spec.RFQ{

		OrgID:       req.OrgID,
		RFQNumber:   rfqNumber,
		CustomerID:  req.CustomerID,
		Stage:       spec.StageRFQCreated,
		Origin:      req.Origin,
		Destination: req.Destination,
		Incoterms:   req.Incoterms,
		TargetDate:  req.TargetDate,
		LeadID:      req.LeadID,
	}

	rfq.HealthScore = b.calculateHealthScore(rfq)

	if err := b.dl.CreateRFQ(ctx, rfq); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	// 2. Increment Entitlement Usage
	if b.entitlementSvc != nil {
		_ = b.entitlementSvc.IncrementUsage(ctx, int64(req.OrgID), subscription.MetricRFQs, 1)
	}

	// Update lead status if converted from a lead
	if req.LeadID != nil {
		if err := b.dl.ConvertLead(ctx, req.OrgID, *req.LeadID); err != nil {
			log.Printf("[RFQ Service] Failed to update lead status to CONVERTED: %v", err)
		} else {
			log.Printf("[RFQ Service] Converted lead ID %d to CONVERTED status successfully", *req.LeadID)
		}
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

	origin := ""
	if rfq.Origin != nil {
		origin = *rfq.Origin
	}
	dest := ""
	if rfq.Destination != nil {
		dest = *rfq.Destination
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        int64(rfq.OrgID),
		Action:       domain.ActionCreate,
		Module:       domain.ModuleRFQs,
		ResourceType: "RFQ",
		ResourceID:   fmt.Sprintf("%d", rfq.ID),
		ResourceName: fmt.Sprintf("RFQ-%d", rfq.ID),
		Description:  fmt.Sprintf("Created RFQ RFQ-%d (%s → %s)", rfq.ID, origin, dest),
		Result:       domain.ResultSuccess,
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

func (b *businessLogic) GetRFQByLeadID(ctx context.Context, orgID int32, leadID int64) (*spec.RFQ, error) {
	rfq, err := b.dl.GetRFQByLeadID(ctx, orgID, leadID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
	}
	return rfq, nil
}

func (b *businessLogic) GetTimeline(ctx context.Context, orgID, rfqID int32) ([]spec.TimelineEvent, error) {
	rfq, err := b.dl.GetRFQByID(ctx, orgID, rfqID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
	}

	events, err := b.dl.GetRFQTimeline(ctx, orgID, rfqID, rfq.LeadID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	return events, nil
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

	oldStage := rfq.Stage
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

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        int64(orgID),
		Action:       domain.ActionUpdate,
		Module:       domain.ModuleRFQs,
		ResourceType: "RFQ",
		ResourceID:   fmt.Sprintf("%d", rfq.ID),
		ResourceName: fmt.Sprintf("RFQ-%d", rfq.ID),
		Description:  fmt.Sprintf("Advanced RFQ RFQ-%d stage to %s", rfq.ID, newStage),
		Before:       map[string]interface{}{"stage": oldStage},
		After:        map[string]interface{}{"stage": newStage},
		Result:       domain.ResultSuccess,
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

// GetRequirements evaluates operational readiness for an RFQ.
// It enforces multi-tenant org isolation by loading via (orgID, rfqID),
// then delegates to the pure deterministic requirements engine.
func (b *businessLogic) GetRequirements(ctx context.Context, orgID, rfqID int32) (*spec.GetRequirementsResponse, error) {
	rfq, err := b.dl.GetRFQByID(ctx, orgID, rfqID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
	}
	// EvaluateRequirements is a pure function: no DB calls, no side effects.
	// It reads only from the already-loaded RFQ struct (items embedded by GetRFQByID).
	result := EvaluateRequirements(rfq)
	return result, nil
}

// GetActivity aggregates and normalizes the full chronological operational timeline
// for an RFQ across Lead, interactions, AI tasks, requirements, and quotes.
// Org isolation is enforced: the RFQ must belong to orgID.
func (b *businessLogic) GetActivity(ctx context.Context, orgID, rfqID int32) (*spec.GetActivityResponse, error) {
	rfq, err := b.dl.GetRFQByID(ctx, orgID, rfqID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
	}

	rawTimeline, err := b.dl.GetRFQTimeline(ctx, orgID, rfqID, rfq.LeadID)
	if err != nil {
		rawTimeline = []spec.TimelineEvent{}
	}

	reqEval := EvaluateRequirements(rfq)
	activityResp := BuildRFQActivity(rfq, rawTimeline, reqEval)
	return activityResp, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Document Management Business Logic (Task 12)
// ──────────────────────────────────────────────────────────────────────────────

// GetDocuments returns all resolved document requirements and active records for an RFQ.
func (b *businessLogic) GetDocuments(ctx context.Context, orgID, rfqID int32) (*spec.GetDocumentsResponse, error) {
	rfq, err := b.dl.GetRFQByID(ctx, orgID, rfqID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
	}

	docs, err := b.dl.GetRFQDocuments(ctx, orgID, rfqID)
	if err != nil {
		docs = []spec.RFQDocument{}
	}
	rfq.Documents = docs

	reqs := EvaluateRequirements(rfq)
	resp := BuildGetDocumentsResponse(rfq, reqs.DocumentRequirements, docs)
	return resp, nil
}

// CreateDocument creates or attaches a new document for an RFQ.
func (b *businessLogic) CreateDocument(ctx context.Context, orgID, rfqID int32, req spec.CreateDocumentRequest, uploader string) (*spec.RFQDocument, error) {
	_, err := b.dl.GetRFQByID(ctx, orgID, rfqID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
	}


	normType := NormalizeDocType(req.DocumentType)
	name := req.DocumentName
	if name == "" {
		name = normType
	}

	status := spec.DocStatusUploaded
	if req.FileName == nil || *req.FileName == "" {
		status = spec.DocStatusRequested
	}

	now := time.Now()
	uploaderName := uploader
	if uploaderName == "" {
		uploaderName = "Operations Team"
	}

	doc := &spec.RFQDocument{
		OrgID:        orgID,
		RFQID:        rfqID,
		DocumentType: normType,
		DocumentName: name,
		Description:  req.Description,
		Status:       status,
		FileName:     req.FileName,
		FileURL:      req.FileURL,
		FileSize:     req.FileSize,
		MimeType:     req.MimeType,
		UploadedBy:   &uploaderName,
		UploadedAt:   &now,
		ExpiresAt:    req.ExpiresAt,
		Metadata:     req.Metadata,
	}

	if err := b.dl.CreateRFQDocument(ctx, orgID, doc); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	// Log activity
	action := "DOCUMENT_UPLOADED"
	if status == spec.DocStatusRequested {
		action = "DOCUMENT_REQUESTED"
	}
	_ = b.dl.CreateActivity(ctx, orgID, "RFQ", int64(rfqID), action, fmt.Sprintf("%s was attached by %s.", doc.DocumentName, uploaderName), uploaderName)

	return doc, nil
}

// UpdateDocumentStatus transitions a document's status with validation.
func (b *businessLogic) UpdateDocumentStatus(ctx context.Context, orgID, rfqID int32, documentID int64, req spec.UpdateDocumentStatusRequest, reviewer string) (*spec.RFQDocument, error) {
	_, err := b.dl.GetRFQByID(ctx, orgID, rfqID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
	}

	doc, err := b.dl.GetRFQDocumentByID(ctx, orgID, rfqID, documentID)
	if err != nil || doc == nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("document not found"))
	}

	if err := ValidateStatusTransition(doc.Status, req.Status); err != nil {
		return nil, err
	}

	reviewerName := reviewer
	if reviewerName == "" {
		reviewerName = "Operations Team"
	}

	if err := b.dl.UpdateRFQDocumentStatus(ctx, orgID, rfqID, documentID, req.Status, reviewerName, req.RejectionReason); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	// Log activity
	action := "DOCUMENT_UPDATED"
	switch strings.ToUpper(req.Status) {
	case spec.DocStatusApproved:
		action = "DOCUMENT_APPROVED"
	case spec.DocStatusRejected:
		action = "DOCUMENT_REJECTED"
	case spec.DocStatusUnderReview:
		action = "DOCUMENT_REVIEW_STARTED"
	case spec.DocStatusUploaded:
		action = "DOCUMENT_UPLOADED"
	case spec.DocStatusRequested:
		action = "DOCUMENT_REQUESTED"
	}
	desc := fmt.Sprintf("%s status was updated to %s by %s.", doc.DocumentName, req.Status, reviewerName)
	if req.RejectionReason != nil && *req.RejectionReason != "" {
		desc = fmt.Sprintf("%s was rejected by %s: %s", doc.DocumentName, reviewerName, *req.RejectionReason)
	}
	_ = b.dl.CreateActivity(ctx, orgID, "RFQ", int64(rfqID), action, desc, reviewerName)

	updatedDoc, err := b.dl.GetRFQDocumentByID(ctx, orgID, rfqID, documentID)
	if err != nil {
		return nil, err
	}
	return updatedDoc, nil
}

// DeleteDocument removes a document from an RFQ.
func (b *businessLogic) DeleteDocument(ctx context.Context, orgID, rfqID int32, documentID int64) error {
	_, err := b.dl.GetRFQByID(ctx, orgID, rfqID)
	if err != nil {
		return svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
	}

	doc, err := b.dl.GetRFQDocumentByID(ctx, orgID, rfqID, documentID)
	if err != nil || doc == nil {
		return svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("document not found"))
	}

	if err := b.dl.DeleteRFQDocument(ctx, orgID, rfqID, documentID); err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	_ = b.dl.CreateActivity(ctx, orgID, "RFQ", int64(rfqID), "DOCUMENT_DELETED", fmt.Sprintf("%s was deleted.", doc.DocumentName), "Operations Team")
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Quotes & Pricing Management Implementation (Task 13)
// ──────────────────────────────────────────────────────────────────────────────

func (b *businessLogic) GetQuotes(ctx context.Context, orgID, rfqID int32) (*spec.GetQuotesResponse, error) {
	if orgID <= 0 || rfqID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	rfq, err := b.dl.GetRFQByID(ctx, orgID, rfqID)
	if err != nil || rfq == nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("rfq %d not found in org %d", rfqID, orgID))
	}

	quotes, err := b.dl.GetRFQQuotes(ctx, orgID, rfqID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	reqs, _ := b.GetRequirements(ctx, orgID, rfqID)

	resp := BuildQuotesResponse(rfq, quotes, reqs, time.Now())
	return resp, nil
}

func (b *businessLogic) CreateRFQQuote(ctx context.Context, orgID, rfqID int32, req spec.CreateQuoteRequest, creator string) (*spec.RFQQuote, error) {
	if orgID <= 0 || rfqID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	if strings.TrimSpace(req.CarrierName) == "" {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("carrier name is required"))
	}
	if req.BuyPrice <= 0 || req.SellPrice <= 0 {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("buy and sell price must be greater than 0"))
	}

	rfq, err := b.dl.GetRFQByID(ctx, orgID, rfqID)
	if err != nil || rfq == nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("rfq %d not found", rfqID))
	}

	if req.Currency == "" {
		req.Currency = "USD"
	}

	quote := &spec.RFQQuote{
		RFQID:                 rfqID,
		OrgID:                 orgID,
		CarrierID:             req.CarrierID,
		CarrierName:           strings.TrimSpace(req.CarrierName),
		QuoteReference:        req.QuoteReference,
		Status:                spec.QuoteStatusReceived,
		Currency:              req.Currency,
		BuyPrice:              req.BuyPrice,
		SellPrice:             req.SellPrice,
		OceanFreight:          0,
		OriginCharges:         0,
		DestinationCharges:    0,
		TotalBuyPrice:         req.BuyPrice,
		TransitTimeDays:       req.TransitTimeDays,
		FreeDays:              req.FreeDays,
		ValidFrom:             req.ValidFrom,
		ValidUntil:            req.ValidUntil,
		ETD:                   req.ETD,
		ETA:                   req.ETA,
		Notes:                 req.Notes,
		Charges:               req.Charges,
		IsRecommended:         false,
		ReliabilityScore:      85,
		HistoricalSuccessRate: 92.5,
	}

	if req.OceanFreight != nil {
		quote.OceanFreight = *req.OceanFreight
	}
	if req.OriginCharges != nil {
		quote.OriginCharges = *req.OriginCharges
	}
	if req.DestinationCharges != nil {
		quote.DestinationCharges = *req.DestinationCharges
	}

	if err := b.dl.CreateRFQQuote(ctx, orgID, quote); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	ref := "—"
	if quote.QuoteReference != nil && *quote.QuoteReference != "" {
		ref = *quote.QuoteReference
	}
	creatorName := creator
	if creatorName == "" {
		creatorName = "Operations Team"
	}

	_ = b.dl.CreateActivity(ctx, orgID, spec.EntityQuote, quote.ID, spec.ActionQuoteReceived,
		fmt.Sprintf("Carrier quotation received from %s (Ref: %s) for %s %.2f", quote.CarrierName, ref, quote.Currency, quote.BuyPrice), creatorName)

	return b.dl.GetRFQQuoteByID(ctx, orgID, rfqID, quote.ID)
}

func (b *businessLogic) UpdateRFQQuote(ctx context.Context, orgID, rfqID int32, quoteID int64, req spec.UpdateQuoteRequest, updater string) (*spec.RFQQuote, error) {
	if orgID <= 0 || rfqID <= 0 || quoteID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	existing, err := b.dl.GetRFQQuoteByID(ctx, orgID, rfqID, quoteID)
	if err != nil || existing == nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("quote %d not found", quoteID))
	}

	if req.CarrierName != nil && strings.TrimSpace(*req.CarrierName) != "" {
		existing.CarrierName = strings.TrimSpace(*req.CarrierName)
	}
	if req.CarrierID != nil {
		existing.CarrierID = req.CarrierID
	}
	if req.QuoteReference != nil {
		existing.QuoteReference = req.QuoteReference
	}
	if req.Currency != nil && *req.Currency != "" {
		existing.Currency = *req.Currency
	}
	if req.BuyPrice != nil && *req.BuyPrice > 0 {
		existing.BuyPrice = *req.BuyPrice
		existing.TotalBuyPrice = *req.BuyPrice
	}
	if req.SellPrice != nil && *req.SellPrice > 0 {
		existing.SellPrice = *req.SellPrice
	}
	if req.OceanFreight != nil {
		existing.OceanFreight = *req.OceanFreight
	}
	if req.OriginCharges != nil {
		existing.OriginCharges = *req.OriginCharges
	}
	if req.DestinationCharges != nil {
		existing.DestinationCharges = *req.DestinationCharges
	}
	if req.TransitTimeDays != nil {
		existing.TransitTimeDays = req.TransitTimeDays
	}
	if req.FreeDays != nil {
		existing.FreeDays = req.FreeDays
	}
	if req.ValidFrom != nil {
		existing.ValidFrom = req.ValidFrom
	}
	if req.ValidUntil != nil {
		existing.ValidUntil = req.ValidUntil
	}
	if req.ETD != nil {
		existing.ETD = req.ETD
	}
	if req.ETA != nil {
		existing.ETA = req.ETA
	}
	if req.Notes != nil {
		existing.Notes = req.Notes
	}
	if req.Charges != nil {
		existing.Charges = req.Charges
	}
	if req.Status != nil && *req.Status != "" && *req.Status != existing.Status {
		if err := ValidateQuoteTransition(existing.Status, *req.Status); err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}
		existing.Status = *req.Status
	}

	if err := b.dl.UpdateRFQQuote(ctx, orgID, existing); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	updaterName := updater
	if updaterName == "" {
		updaterName = "Pricing Manager"
	}
	_ = b.dl.CreateActivity(ctx, orgID, spec.EntityQuote, quoteID, spec.ActionQuoteUpdated,
		fmt.Sprintf("Quote from %s was updated with new commercial details.", existing.CarrierName), updaterName)

	return b.dl.GetRFQQuoteByID(ctx, orgID, rfqID, quoteID)
}

func (b *businessLogic) UpdateRFQQuoteStatus(ctx context.Context, orgID, rfqID int32, quoteID int64, req spec.UpdateQuoteStatusRequest, updater string) (*spec.RFQQuote, error) {
	if orgID <= 0 || rfqID <= 0 || quoteID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	existing, err := b.dl.GetRFQQuoteByID(ctx, orgID, rfqID, quoteID)
	if err != nil || existing == nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("quote %d not found", quoteID))
	}

	if err := ValidateQuoteTransition(existing.Status, req.Status); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}

	if err := b.dl.UpdateRFQQuoteStatus(ctx, orgID, rfqID, quoteID, req.Status); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	updaterName := updater
	if updaterName == "" {
		updaterName = "Pricing Operations"
	}
	_ = b.dl.CreateActivity(ctx, orgID, spec.EntityQuote, quoteID, spec.ActionQuoteUnderReview,
		fmt.Sprintf("Quote from %s transitioned to %s.", existing.CarrierName, req.Status), updaterName)

	return b.dl.GetRFQQuoteByID(ctx, orgID, rfqID, quoteID)
}

func (b *businessLogic) RecommendRFQQuote(ctx context.Context, orgID, rfqID int32, quoteID int64, recommender string) (*spec.RFQQuote, error) {
	if orgID <= 0 || rfqID <= 0 || quoteID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	existing, err := b.dl.GetRFQQuoteByID(ctx, orgID, rfqID, quoteID)
	if err != nil || existing == nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("quote %d not found", quoteID))
	}

	if err := CanRecommendQuote(existing, time.Now()); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}

	if err := b.dl.RecommendRFQQuote(ctx, orgID, rfqID, quoteID); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	recommenderName := recommender
	if recommenderName == "" {
		recommenderName = "Pricing Team"
	}
	_ = b.dl.CreateActivity(ctx, orgID, spec.EntityQuote, quoteID, spec.ActionQuoteRecommended,
		fmt.Sprintf("Quote from %s (%s %.2f) was selected as the recommended operational option.", existing.CarrierName, existing.Currency, existing.BuyPrice), recommenderName)

	return b.dl.GetRFQQuoteByID(ctx, orgID, rfqID, quoteID)
}

func (b *businessLogic) ApproveRFQQuote(ctx context.Context, orgID, rfqID int32, quoteID int64, req spec.ApproveRFQQuoteRequest, approver string) (*spec.RFQQuote, error) {
	if orgID <= 0 || rfqID <= 0 || quoteID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	existing, err := b.dl.GetRFQQuoteByID(ctx, orgID, rfqID, quoteID)
	if err != nil || existing == nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("quote %d not found", quoteID))
	}

	if err := CanApproveQuote(existing, time.Now()); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}

	approverName := approver
	if approverName == "" {
		approverName = "Operations Manager"
	}

	if err := b.dl.ApproveRFQQuote(ctx, orgID, rfqID, quoteID, approverName); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	_ = b.dl.CreateActivity(ctx, orgID, spec.EntityQuote, quoteID, spec.ActionQuoteApproved,
		fmt.Sprintf("Quote from %s was officially APPROVED by %s for customer quotation.", existing.CarrierName, approverName), approverName)

	return b.dl.GetRFQQuoteByID(ctx, orgID, rfqID, quoteID)
}

func (b *businessLogic) SelectRFQQuoteForCustomer(ctx context.Context, orgID, rfqID int32, quoteID int64, selector string) (*spec.RFQQuote, error) {
	if orgID <= 0 || rfqID <= 0 || quoteID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	existing, err := b.dl.GetRFQQuoteByID(ctx, orgID, rfqID, quoteID)
	if err != nil || existing == nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("quote %d not found", quoteID))
	}

	if err := b.dl.SelectRFQQuoteForCustomer(ctx, orgID, rfqID, quoteID); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	selectorName := selector
	if selectorName == "" {
		selectorName = "Commercial Operations"
	}
	_ = b.dl.CreateActivity(ctx, orgID, spec.EntityQuote, quoteID, spec.ActionQuoteSelected,
		fmt.Sprintf("Quote from %s was selected for customer quotation presentation.", existing.CarrierName), selectorName)

	return b.dl.GetRFQQuoteByID(ctx, orgID, rfqID, quoteID)
}

func (b *businessLogic) DeleteRFQQuote(ctx context.Context, orgID, rfqID int32, quoteID int64) error {
	if orgID <= 0 || rfqID <= 0 || quoteID <= 0 {
		return svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	existing, err := b.dl.GetRFQQuoteByID(ctx, orgID, rfqID, quoteID)
	if err != nil || existing == nil {
		return svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("quote not found"))
	}

	if err := b.dl.DeleteRFQQuote(ctx, orgID, rfqID, quoteID); err != nil {
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	_ = b.dl.CreateActivity(ctx, orgID, spec.EntityQuote, quoteID, spec.ActionQuoteWithdrawn,
		fmt.Sprintf("Carrier quote from %s was withdrawn.", existing.CarrierName), "Operations Team")
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Task 14: Booking & Shipment Handoff Implementations
// ──────────────────────────────────────────────────────────────────────────────

func (b *businessLogic) GetBookingHandoff(ctx context.Context, orgID, rfqID int32) (*spec.GetBookingHandoffResponse, error) {
	if orgID <= 0 || rfqID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	rfq, err := b.dl.GetRFQByID(ctx, orgID, rfqID)
	if err != nil || rfq == nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("rfq %d not found", rfqID))
	}

	quotes, err := b.dl.GetRFQQuotes(ctx, orgID, rfqID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	reqsResp, _ := b.GetRequirements(ctx, orgID, rfqID)
	var operationalReadiness *spec.OperationalReadiness
	var docReqs []spec.DocumentRequirement
	if reqsResp != nil {
		operationalReadiness = &reqsResp.OperationalReadiness
		docReqs = reqsResp.DocumentRequirements
	}

	docs, _ := b.dl.GetRFQDocuments(ctx, orgID, rfqID)
	docsResp := BuildGetDocumentsResponse(rfq, docReqs, docs)
	docSummary := docsResp.Summary

	eligibility := EvaluateBookingEligibility(rfq, quotes, operationalReadiness, &docSummary)

	bookings, err := b.dl.GetRFQBookings(ctx, orgID, rfqID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	var activeBooking *spec.RFQBooking
	var latestStatus *string
	if len(bookings) > 0 {
		for i := range bookings {
			if bookings[i].Status == spec.BookingStatusConfirmed {
				activeBooking = &bookings[i]
				break
			}
		}
		if activeBooking == nil {
			activeBooking = &bookings[0]
		}
		latestStatus = &activeBooking.Status
	}

	return &spec.GetBookingHandoffResponse{
		Eligibility: eligibility,
		Summary: spec.BookingSummary{
			TotalBookings: len(bookings),
			ActiveBooking: activeBooking,
			LatestStatus:  latestStatus,
		},
		Bookings: bookings,
	}, nil
}

func (b *businessLogic) CreateBookingFromRFQ(ctx context.Context, orgID, rfqID int32, req spec.CreateBookingRequest, creator string) (*spec.RFQBooking, error) {
	if orgID <= 0 || rfqID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	rfq, err := b.dl.GetRFQByID(ctx, orgID, rfqID)
	if err != nil || rfq == nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("rfq %d not found", rfqID))
	}

	quotes, err := b.dl.GetRFQQuotes(ctx, orgID, rfqID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	reqsResp, _ := b.GetRequirements(ctx, orgID, rfqID)
	var operationalReadiness *spec.OperationalReadiness
	var docReqs []spec.DocumentRequirement
	if reqsResp != nil {
		operationalReadiness = &reqsResp.OperationalReadiness
		docReqs = reqsResp.DocumentRequirements
	}

	docs, _ := b.dl.GetRFQDocuments(ctx, orgID, rfqID)
	docsResp := BuildGetDocumentsResponse(rfq, docReqs, docs)
	docSummary := docsResp.Summary

	eligibility := EvaluateBookingEligibility(rfq, quotes, operationalReadiness, &docSummary)
	if !eligibility.IsEligible {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("booking creation blocked: %s", strings.Join(eligibility.MissingPrerequisites, ", ")))
	}

	// 1. Generate Booking Number if not provided
	bookingNumber := ""
	if req.BookingNumber != nil && strings.TrimSpace(*req.BookingNumber) != "" {
		bookingNumber = strings.TrimSpace(*req.BookingNumber)
	} else {
		bookingNumber = fmt.Sprintf("BK-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)
	}

	// 2. Resolve carrier info
	carrierName := req.CarrierName
	var quoteID *int64 = req.QuoteID
	var carrierID *string = req.CarrierID
	var carrierSCAC *string = req.CarrierSCAC

	if carrierName == "" && eligibility.ApprovedCarrier != nil {
		carrierName = *eligibility.ApprovedCarrier
	}
	if quoteID == nil && eligibility.ApprovedQuoteID != nil {
		quoteID = eligibility.ApprovedQuoteID
	}

	// SCAC fallback if possible
	if carrierSCAC == nil {
		scacMap := map[string]string{
			"Maersk": "MAEU", "Maersk Line": "MAEU",
			"Hapag-Lloyd": "HLCU",
			"MSC": "MSCU",
			"CMA CGM": "CMDU",
			"Evergreen": "EGLV",
			"Cosco": "COSU",
		}
		if scac, exists := scacMap[carrierName]; exists {
			carrierSCAC = &scac
		}
	}

	// 3. Resolve Origin & Destination
	originPort := req.OriginPort
	if originPort == "" && rfq.Origin != nil {
		originPort = *rfq.Origin
	}
	destPort := req.DestinationPort
	if destPort == "" && rfq.Destination != nil {
		destPort = *rfq.Destination
	}

	creatorName := creator
	if creatorName == "" {
		creatorName = "Operations Team"
	}

	booking := spec.RFQBooking{
		OrgID:               int64(orgID),
		RFQID:               int64(rfqID),
		QuoteID:             quoteID,
		BookingNumber:       bookingNumber,
		CarrierID:           carrierID,
		CarrierName:         carrierName,
		CarrierSCAC:         carrierSCAC,
		Status:              spec.BookingStatusDraft,
		OriginPort:          originPort,
		DestinationPort:     destPort,
		VesselName:          req.VesselName,
		VoyageNumber:        req.VoyageNumber,
		ETD:                 req.ETD,
		ETA:                 req.ETA,
		CargoSummary:        req.CargoSummary,
		SpecialInstructions: req.SpecialInstructions,
		CreatedBy:           &creatorName,
	}

	if err := b.dl.CreateRFQBooking(ctx, orgID, &booking); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	// Log activity timeline event
	_ = b.dl.CreateActivity(ctx, orgID, spec.EntityBooking, booking.ID, spec.ActionBookingCreated,
		fmt.Sprintf("Booking %s was created with %s for route %s → %s.", booking.BookingNumber, booking.CarrierName, booking.OriginPort, booking.DestinationPort), creatorName)

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        int64(orgID),
		Action:       domain.ActionCreate,
		Module:       domain.ModuleBookings,
		ResourceType: "BOOKING",
		ResourceID:   fmt.Sprintf("%d", booking.ID),
		ResourceName: booking.BookingNumber,
		Description:  fmt.Sprintf("Created booking %s with carrier %s", booking.BookingNumber, booking.CarrierName),
		Result:       domain.ResultSuccess,
	})

	return &booking, nil
}

func (b *businessLogic) UpdateBookingStatus(ctx context.Context, orgID, rfqID int32, bookingID int64, req spec.UpdateBookingStatusRequest, updater string) (*spec.RFQBooking, error) {
	if orgID <= 0 || rfqID <= 0 || bookingID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	existing, err := b.dl.GetRFQBookingByID(ctx, orgID, rfqID, bookingID)
	if err != nil || existing == nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("booking %d not found", bookingID))
	}

	targetStatus := strings.ToUpper(strings.TrimSpace(req.Status))
	if err := ValidateBookingTransition(existing.Status, targetStatus); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}

	if err := b.dl.UpdateRFQBookingStatus(ctx, orgID, rfqID, bookingID, targetStatus); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	updaterName := updater
	if updaterName == "" {
		updaterName = "Carrier Operations"
	}

	action := spec.ActionBookingUpdated
	desc := fmt.Sprintf("Booking %s status changed from %s to %s.", existing.BookingNumber, existing.Status, targetStatus)
	if targetStatus == spec.BookingStatusConfirmed {
		action = spec.ActionBookingConfirmed
		desc = fmt.Sprintf("Booking %s was CONFIRMED with carrier %s.", existing.BookingNumber, existing.CarrierName)
	} else if targetStatus == spec.BookingStatusRequested {
		action = spec.ActionBookingRequested
		desc = fmt.Sprintf("Booking %s was requested from carrier %s.", existing.BookingNumber, existing.CarrierName)
	} else if targetStatus == spec.BookingStatusCancelled {
		action = spec.ActionBookingCancelled
		desc = fmt.Sprintf("Booking %s was CANCELLED.", existing.BookingNumber)
	}

	_ = b.dl.CreateActivity(ctx, orgID, spec.EntityBooking, bookingID, action, desc, updaterName)

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        int64(orgID),
		Action:       domain.ActionUpdate,
		Module:       domain.ModuleBookings,
		ResourceType: "BOOKING",
		ResourceID:   fmt.Sprintf("%d", bookingID),
		ResourceName: existing.BookingNumber,
		Description:  desc,
		Before:       map[string]interface{}{"status": existing.Status},
		After:        map[string]interface{}{"status": targetStatus},
		Result:       domain.ResultSuccess,
	})

	return b.dl.GetRFQBookingByID(ctx, orgID, rfqID, bookingID)
}

func (b *businessLogic) GetShipmentHandoff(ctx context.Context, orgID, rfqID int32) (*spec.GetShipmentHandoffResponse, error) {
	if orgID <= 0 || rfqID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	rfq, err := b.dl.GetRFQByID(ctx, orgID, rfqID)
	if err != nil || rfq == nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("rfq %d not found in org %d", rfqID, orgID))
	}

	bookings, err := b.dl.GetRFQBookings(ctx, orgID, rfqID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	var sourceBooking *spec.RFQBooking
	for i := range bookings {
		if bookings[i].Status == spec.BookingStatusConfirmed {
			sourceBooking = &bookings[i]
			break
		}
	}
	if sourceBooking == nil && len(bookings) > 0 {
		sourceBooking = &bookings[0]
	}

	shipments, err := b.dl.GetRFQShipments(ctx, orgID, rfqID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	var activeShipment *spec.RFQShipment
	var latestStatus *string
	if len(shipments) > 0 {
		activeShipment = &shipments[0]
		latestStatus = &activeShipment.Status
	}

	return &spec.GetShipmentHandoffResponse{
		SourceBooking: sourceBooking,
		Summary: spec.ShipmentSummary{
			TotalShipments: len(shipments),
			ActiveShipment: activeShipment,
			LatestStatus:  latestStatus,
		},
		Shipments: shipments,
	}, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Task 15: Dedicated Booking Workspace Business Logic
// ──────────────────────────────────────────────────────────────────────────────

func (b *businessLogic) GetBookingsWorkspace(ctx context.Context, orgID int32, filter spec.BookingListFilter) (*spec.GetBookingsWorkspaceResponse, error) {
	if orgID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	filter.OrgID = orgID
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	items, kpis, totalItems, err := b.dl.GetBookingsWorkspace(ctx, orgID, filter)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	carriers, err := b.dl.GetUniqueCarriersForBookings(ctx, orgID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	totalPages := (totalItems + filter.Limit - 1) / filter.Limit
	if totalPages == 0 {
		totalPages = 1
	}

	return &spec.GetBookingsWorkspaceResponse{
		Bookings: items,
		KPIs:     kpis,
		Pagination: spec.BookingPagination{
			CurrentPage: filter.Page,
			PageSize:    filter.Limit,
			TotalItems:  totalItems,
			TotalPages:  totalPages,
		},
		Carriers: carriers,
	}, nil
}

func (b *businessLogic) GetBookingWorkspaceDetail(ctx context.Context, orgID int32, bookingID int64) (*spec.BookingDetailResponse, error) {
	if orgID <= 0 || bookingID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	detail, err := b.dl.GetBookingWorkspaceDetail(ctx, orgID, bookingID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	if detail == nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("booking %d not found in org %d", bookingID, orgID))
	}

	return detail, nil
}

func (b *businessLogic) DirectUpdateBookingStatus(ctx context.Context, orgID int32, bookingID int64, req spec.DirectUpdateBookingStatusRequest, updater string) (*spec.RFQBooking, error) {
	if orgID <= 0 || bookingID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	existing, err := b.dl.GetBookingByIDOnly(ctx, orgID, bookingID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	if existing == nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("booking %d not found in org %d", bookingID, orgID))
	}

	targetStatus := strings.ToUpper(strings.TrimSpace(req.Status))
	if err := ValidateBookingTransition(existing.Status, targetStatus); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}

	if err := b.dl.UpdateBookingStatusDirect(ctx, orgID, bookingID, targetStatus); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	updaterName := updater
	if updaterName == "" {
		updaterName = "Carrier Operations"
	}

	action := spec.ActionBookingUpdated
	desc := fmt.Sprintf("Booking %s status changed from %s to %s.", existing.BookingNumber, existing.Status, targetStatus)
	if targetStatus == spec.BookingStatusConfirmed {
		action = spec.ActionBookingConfirmed
		desc = fmt.Sprintf("Booking %s was CONFIRMED with carrier %s.", existing.BookingNumber, existing.CarrierName)
	} else if targetStatus == spec.BookingStatusRequested {
		action = spec.ActionBookingRequested
		desc = fmt.Sprintf("Booking %s was requested from carrier %s.", existing.BookingNumber, existing.CarrierName)
	} else if targetStatus == spec.BookingStatusCancelled {
		action = spec.ActionBookingCancelled
		desc = fmt.Sprintf("Booking %s was CANCELLED.", existing.BookingNumber)
	}

	_ = b.dl.CreateActivity(ctx, orgID, spec.EntityBooking, bookingID, action, desc, updaterName)

	return b.dl.GetBookingByIDOnly(ctx, orgID, bookingID)
}

func (b *businessLogic) GetEligibleRFQsForBooking(ctx context.Context, orgID int32) ([]spec.EligibleBookingRFQ, error) {
	if orgID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	items, err := b.dl.GetEligibleRFQsForBooking(ctx, orgID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return items, nil
}

func (b *businessLogic) CreateShipmentFromBooking(ctx context.Context, orgID int32, bookingID int64, req spec.CreateShipmentFromBookingRequest, creator string) (*spec.RFQShipment, error) {
	if orgID <= 0 || bookingID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	shipment, err := b.dl.CreateShipmentFromBookingTx(ctx, orgID, bookingID, req, creator)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}

	return shipment, nil
}

func (b *businessLogic) BookWithCarrier(ctx context.Context, orgID int32, bookingID int64, req spec.BookWithCarrierRequest, user string) (*spec.BookingDetailResponse, error) {
	if orgID <= 0 || bookingID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	if b.carrierBookingEngine == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInternal)
	}
	return b.carrierBookingEngine.SubmitCarrierBooking(ctx, orgID, bookingID, req, user)
}

func (b *businessLogic) SyncCarrierBooking(ctx context.Context, orgID int32, bookingID int64, user string) (*spec.BookingDetailResponse, error) {
	if orgID <= 0 || bookingID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	if b.carrierBookingEngine == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInternal)
	}
	return b.carrierBookingEngine.SyncCarrierBooking(ctx, orgID, bookingID, user)
}





