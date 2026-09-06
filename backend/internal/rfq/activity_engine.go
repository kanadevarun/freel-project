package rfq

// activity_engine.go — Task 11: Pure Deterministic RFQ Activity Aggregation & Normalization Engine
//
// BuildRFQActivity aggregates, normalizes, and classifies operational events from:
//   1. Real RFQ records (rfqs, rfq_items, rfq_quotes)
//   2. Real Lead records & conversion lineage (leads)
//   3. Real customer email & communication history (lead_interactions)
//   4. Real database audit events (activities)
//   5. Real Task 10 requirements readiness evaluations (EvaluateRequirements)
//
// Critical rules:
//   - ZERO fabricated or fake events. Every event is strictly derived from real data.
//   - Preserves complete Lead -> Interaction -> AI -> RFQ -> Requirements -> Quotes lineage.
//   - Multi-tenant organization isolation is enforced prior to calling this pure engine.

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/freel/backend/internal/rfq/spec"
)

// BuildRFQActivity is the top-level pure function aggregating and normalizing all
// operational events for a given RFQ into a unified GetActivityResponse.
func BuildRFQActivity(
	rfq *spec.RFQ,
	rawTimeline []spec.TimelineEvent,
	reqEval *spec.GetRequirementsResponse,
) *spec.GetActivityResponse {
	if rfq == nil {
		return &spec.GetActivityResponse{
			Summary: spec.ActivitySummary{},
			Events:  []spec.ActivityEvent{},
		}
	}

	eventsMap := make(map[string]spec.ActivityEvent)

	// 1. Normalize raw database events from activities & lead_interactions
	for _, raw := range rawTimeline {
		norm := normalizeRawTimelineEvent(raw, rfq)
		if norm != nil {
			eventsMap[norm.ID] = *norm
		}
	}

	// 2. Ensure RFQ Creation event exists
	rfqCreatedID := fmt.Sprintf("rfq-created-%d", rfq.ID)
	hasRFQCreated := false
	for _, ev := range eventsMap {
		if ev.Type == spec.ActivityRFQCreated {
			hasRFQCreated = true
			break
		}
	}
	if !hasRFQCreated {
		createdDesc := fmt.Sprintf("RFQ %s was created in the system.", rfq.RFQNumber)
		actorName := "Operations Team"
		sourceType := "RFQ"
		sourceID := fmt.Sprintf("%d", rfq.ID)
		if rfq.LeadID != nil && *rfq.LeadID > 0 {
			createdDesc = fmt.Sprintf("RFQ %s was created from Lead #%d.", rfq.RFQNumber, *rfq.LeadID)
			sourceType = "LEAD"
			sourceID = fmt.Sprintf("%d", *rfq.LeadID)
		}
		eventsMap[rfqCreatedID] = spec.ActivityEvent{
			ID:                rfqCreatedID,
			Type:              spec.ActivityRFQCreated,
			Category:          spec.ActivityCatOperations,
			Title:             "RFQ Created",
			Description:       createdDesc,
			Timestamp:         rfq.CreatedAt,
			ActorType:         spec.ActorOperations,
			ActorName:         actorName,
			SourceType:        sourceType,
			SourceID:          sourceID,
			IsImportant:       true,
			RequiresAction:    false,
			RelatedEntityType: "RFQ",
			RelatedEntityID:   fmt.Sprintf("%d", rfq.ID),
		}
	}

	// 3. Integrate real Task 10 requirements evaluation milestone
	if reqEval != nil {
		reqEventID := fmt.Sprintf("req-eval-%d", rfq.ID)
		readiness := reqEval.OperationalReadiness
		reqTitle := "Requirements Evaluated"
		reqDesc := fmt.Sprintf("Operational readiness: %s (%d%% readiness score, %d critical blockers).",
			formatStatusTitle(readiness.OverallStatus),
			readiness.ReadinessScore,
			readiness.BlockingCount,
		)
		isImportant := readiness.OverallStatus == spec.ReadinessReadyForQuotation || readiness.BlockingCount > 0
		requiresAction := readiness.BlockingCount > 0

		// Timestamp requirements slightly after RFQ creation or at UpdatedAt
		reqTimestamp := rfq.UpdatedAt
		if reqTimestamp.Before(rfq.CreatedAt) {
			reqTimestamp = rfq.CreatedAt.Add(1 * time.Minute)
		}

		eventsMap[reqEventID] = spec.ActivityEvent{
			ID:                reqEventID,
			Type:              spec.ActivityRequirementsEvaluated,
			Category:          spec.ActivityCatRequirements,
			Title:             reqTitle,
			Description:       reqDesc,
			Timestamp:         reqTimestamp,
			ActorType:         spec.ActorSystem,
			ActorName:         "Requirements Engine",
			SourceType:        "RFQ",
			SourceID:          fmt.Sprintf("%d", rfq.ID),
			IsImportant:       isImportant,
			RequiresAction:    requiresAction,
			RelatedEntityType: "RFQ",
			RelatedEntityID:   fmt.Sprintf("%d", rfq.ID),
			Metadata: map[string]interface{}{
				"overall_status":   readiness.OverallStatus,
				"readiness_score":  readiness.ReadinessScore,
				"blocking_count":   readiness.BlockingCount,
				"next_best_action": readiness.NextBestAction,
			},
		}
	}

	// 4. Integrate real carrier quotes (if any exist in rfq.Quotes)
	for _, q := range rfq.Quotes {
		quoteEventID := fmt.Sprintf("quote-gen-%d", q.ID)
		quoteDesc := fmt.Sprintf("Carrier quote from %s: Buy Price $%.2f, Sell Price $%.2f.",
			q.CarrierName, q.BuyPrice, q.SellPrice)
		if q.TransitTimeDays != nil && *q.TransitTimeDays > 0 {
			quoteDesc = fmt.Sprintf("Carrier quote from %s: %d days transit, Buy $%.2f, Sell $%.2f.",
				q.CarrierName, *q.TransitTimeDays, q.BuyPrice, q.SellPrice)
		}

		eventsMap[quoteEventID] = spec.ActivityEvent{
			ID:                quoteEventID,
			Type:              spec.ActivityQuoteGenerated,
			Category:          spec.ActivityCatQuotes,
			Title:             fmt.Sprintf("Quote Generated — %s", q.CarrierName),
			Description:       quoteDesc,
			Timestamp:         q.CreatedAt,
			ActorType:         spec.ActorOperations,
			ActorName:         "Pricing Engine",
			SourceType:        "QUOTE",
			SourceID:          fmt.Sprintf("%d", q.ID),
			IsImportant:       true,
			RequiresAction:    false,
			RelatedEntityType: "RFQ",
			RelatedEntityID:   fmt.Sprintf("%d", rfq.ID),
			Metadata: map[string]interface{}{
				"quote_id":     q.ID,
				"carrier_name": q.CarrierName,
				"buy_price":    q.BuyPrice,
				"sell_price":   q.SellPrice,
				"status":       q.Status,
			},
		}

		// If quote is approved / accepted
		if strings.ToUpper(q.Status) == "ACCEPTED" || strings.ToUpper(q.Status) == "APPROVED" {
			approvedID := fmt.Sprintf("quote-appr-%d", q.ID)
			eventsMap[approvedID] = spec.ActivityEvent{
				ID:                approvedID,
				Type:              spec.ActivityQuoteApproved,
				Category:          spec.ActivityCatQuotes,
				Title:             fmt.Sprintf("Quote Approved — %s", q.CarrierName),
				Description:       fmt.Sprintf("Quotation from %s for $%.2f was approved.", q.CarrierName, q.SellPrice),
				Timestamp:         q.UpdatedAt,
				ActorType:         spec.ActorOperations,
				ActorName:         "Pricing Operations",
				SourceType:        "QUOTE",
				SourceID:          fmt.Sprintf("%d", q.ID),
				IsImportant:       true,
				RequiresAction:    false,
				RelatedEntityType: "RFQ",
				RelatedEntityID:   fmt.Sprintf("%d", rfq.ID),
			}
		}
	}

	// 5. Flatten into slice and sort chronologically (descending: newest first)
	eventsList := make([]spec.ActivityEvent, 0, len(eventsMap))
	for _, ev := range eventsMap {
		eventsList = append(eventsList, ev)
	}

	sort.Slice(eventsList, func(i, j int) bool {
		return eventsList[i].Timestamp.After(eventsList[j].Timestamp)
	})

	// 6. Compute summary counts
	summary := computeActivitySummary(eventsList)

	return &spec.GetActivityResponse{
		Summary: summary,
		Events:  eventsList,
		LeadID:  rfq.LeadID,
	}
}

// normalizeRawTimelineEvent maps a raw database timeline event into a rich ActivityEvent.
func normalizeRawTimelineEvent(raw spec.TimelineEvent, rfq *spec.RFQ) *spec.ActivityEvent {
	action := strings.ToUpper(raw.Action)
	category := strings.ToUpper(raw.Category)

	var evType spec.ActivityEventType
	var actCategory string
	var actorType string
	var actorName string
	var title string
	var isImportant bool
	var requiresAction bool

	actorName = raw.Actor
	if actorName == "" {
		actorName = "System"
	}

	switch {
	// Lead creation
	case action == spec.ActionCreated && raw.EntityType == spec.EntityLead:
		evType = spec.ActivityLeadCreated
		actCategory = spec.ActivityCatOperations
		actorType = spec.ActorSystem
		title = "Lead Created"
		isImportant = true

	// Lead converted to RFQ
	case action == spec.ActionConverted || action == spec.ActionLeadConverted:
		evType = spec.ActivityLeadConverted
		actCategory = spec.ActivityCatOperations
		actorType = spec.ActorOperations
		title = "Lead Converted to RFQ"
		isImportant = true

	// Inbound customer email / inquiry
	case action == spec.ActionEmailInbound || (category == spec.EntityEmail && strings.Contains(action, "INBOUND")):
		evType = spec.ActivityCustomerInquiry
		actCategory = spec.ActivityCatCustomer
		actorType = spec.ActorCustomer
		title = "Customer Email Received"
		isImportant = true

	// Outbound email / reply to customer
	case action == spec.ActionEmailOutbound || (category == spec.EntityEmail && strings.Contains(action, "OUTBOUND")):
		evType = spec.ActivityClarificationRequested
		actCategory = spec.ActivityCatOperations
		actorType = spec.ActorOperations
		title = "Email Sent to Customer"

	// AI Parsing / Extraction
	case action == spec.ActionAIParsed || action == spec.ActionAIExtracted || action == spec.ActionAIEnriched || category == spec.EntityAI:
		evType = spec.ActivityAIExtraction
		actCategory = spec.ActivityCatAI
		actorType = spec.ActorAI
		actorName = "AI Assistant"
		title = "AI Extracted Shipment Details"
		isImportant = true

	// RFQ Created
	case action == spec.ActionRFQCreated:
		evType = spec.ActivityRFQCreated
		actCategory = spec.ActivityCatOperations
		actorType = spec.ActorOperations
		title = "RFQ Created"
		isImportant = true

	// Quote lifecycle events
	case action == spec.ActionQuoteReceived:
		evType = spec.ActivityQuoteGenerated
		actCategory = spec.ActivityCatQuotes
		actorType = spec.ActorOperations
		title = "Carrier Quote Received"
		isImportant = true

	case action == spec.ActionQuoteCreated:
		evType = spec.ActivityQuoteGenerated
		actCategory = spec.ActivityCatQuotes
		actorType = spec.ActorOperations
		title = "Carrier Quote Created"
		isImportant = true

	case action == spec.ActionQuoteUnderReview:
		evType = spec.ActivityQuoteGenerated
		actCategory = spec.ActivityCatQuotes
		actorType = spec.ActorOperations
		title = "Carrier Quote Under Review"

	case action == spec.ActionQuoteRecommended:
		evType = spec.ActivityQuoteRecommended
		actCategory = spec.ActivityCatQuotes
		actorType = spec.ActorOperations
		title = "Carrier Quote Recommended"
		isImportant = true

	case action == spec.ActionQuoteApproved:
		evType = spec.ActivityQuoteApproved
		actCategory = spec.ActivityCatQuotes
		actorType = spec.ActorOperations
		title = "Carrier Quote Approved"
		isImportant = true

	case action == spec.ActionQuoteSelected:
		evType = spec.ActivityQuoteSelected
		actCategory = spec.ActivityCatQuotes
		actorType = spec.ActorOperations
		title = "Quote Selected for Customer"
		isImportant = true

	case action == spec.ActionQuoteRejected:
		evType = spec.ActivityQuoteRejected
		actCategory = spec.ActivityCatQuotes
		actorType = spec.ActorOperations
		title = "Carrier Quote Rejected"

	case action == spec.ActionQuoteWithdrawn:
		evType = spec.ActivityQuoteWithdrawn
		actCategory = spec.ActivityCatQuotes
		actorType = spec.ActorOperations
		title = "Carrier Quote Withdrawn"

	case action == spec.ActionQuoteGenerated || category == spec.EntityQuote:
		evType = spec.ActivityQuoteGenerated
		actCategory = spec.ActivityCatQuotes
		actorType = spec.ActorOperations
		title = "Carrier Quote Event"
		isImportant = true


	// Document events
	case action == spec.ActionDocumentUploaded:
		evType = spec.ActivityDocumentUploaded
		actCategory = spec.ActivityCatDocuments
		actorType = spec.ActorOperations
		title = "Document Uploaded"
		isImportant = true

	case action == spec.ActionDocumentApproved:
		evType = spec.ActivityDocumentApproved
		actCategory = spec.ActivityCatDocuments
		actorType = spec.ActorOperations
		title = "Document Approved"
		isImportant = true

	case action == spec.ActionDocumentRejected:
		evType = spec.ActivityDocumentRejected
		actCategory = spec.ActivityCatDocuments
		actorType = spec.ActorOperations
		title = "Document Rejected"
		isImportant = true
		requiresAction = true

	case action == spec.ActionDocumentRequested:
		evType = spec.ActivityDocumentRequested
		actCategory = spec.ActivityCatDocuments
		actorType = spec.ActorOperations
		title = "Document Requested"
		requiresAction = true

	case action == spec.ActionDocumentReviewStarted:
		evType = spec.ActivityDocumentReviewStarted
		actCategory = spec.ActivityCatDocuments
		actorType = spec.ActorOperations
		title = "Document Review Started"

	case action == spec.ActionDocumentExpired:
		evType = spec.ActivityDocumentExpired
		actCategory = spec.ActivityCatDocuments
		actorType = spec.ActorOperations
		title = "Document Expired"
		requiresAction = true

	case category == spec.EntityDocument || category == spec.EntityDocuments:
		evType = spec.ActivityDocumentUploaded
		actCategory = spec.ActivityCatDocuments
		actorType = spec.ActorOperations
		title = "Document Activity"

	// Booking events (Task 14)
	case action == spec.ActionBookingCreated:
		evType = spec.ActivityBookingCreated
		actCategory = spec.ActivityCatOperations
		actorType = spec.ActorOperations
		title = "Carrier Booking Created"
		isImportant = true

	case action == spec.ActionBookingRequested:
		evType = spec.ActivityBookingRequested
		actCategory = spec.ActivityCatOperations
		actorType = spec.ActorOperations
		title = "Carrier Booking Requested"
		isImportant = true

	case action == spec.ActionBookingConfirmed:
		evType = spec.ActivityBookingConfirmed
		actCategory = spec.ActivityCatOperations
		actorType = spec.ActorOperations
		title = "Carrier Booking Confirmed"
		isImportant = true

	case action == spec.ActionBookingCancelled:
		evType = spec.ActivityBookingCancelled
		actCategory = spec.ActivityCatOperations
		actorType = spec.ActorOperations
		title = "Carrier Booking Cancelled"
		requiresAction = true

	case category == spec.EntityBooking:
		evType = spec.ActivityBookingCreated
		actCategory = spec.ActivityCatOperations
		actorType = spec.ActorOperations
		title = "Carrier Booking Activity"

	// Shipment events (Task 14)
	case action == spec.ActionShipmentCreated:
		evType = spec.ActivityShipmentCreated
		actCategory = spec.ActivityCatOperations
		actorType = spec.ActorOperations
		title = "Shipment Created"
		isImportant = true

	case action == spec.ActionShipmentDeparted:
		evType = spec.ActivityShipmentDeparted
		actCategory = spec.ActivityCatOperations
		actorType = spec.ActorOperations
		title = "Vessel Departed Origin"
		isImportant = true

	case action == spec.ActionShipmentArrived:
		evType = spec.ActivityShipmentArrived
		actCategory = spec.ActivityCatOperations
		actorType = spec.ActorOperations
		title = "Vessel Arrived at Destination"
		isImportant = true

	case action == spec.ActionShipmentCompleted || action == spec.ActionShipmentDelivered:
		evType = spec.ActivityShipmentCompleted
		actCategory = spec.ActivityCatOperations
		actorType = spec.ActorOperations
		title = "Shipment Completed"
		isImportant = true

	case category == spec.EntityShipment:
		evType = spec.ActivityShipmentCreated
		actCategory = spec.ActivityCatOperations
		actorType = spec.ActorOperations
		title = "Shipment Activity"

	default:
		evType = spec.ActivityUserAction
		actCategory = spec.ActivityCatOperations
		actorType = spec.ActorUser
		title = formatStatusTitle(raw.Action)
	}

	sourceType := raw.EntityType
	if sourceType == "" {
		sourceType = spec.EntityTimeline
	}

	sourceID := fmt.Sprintf("%d", raw.EntityID)

	return &spec.ActivityEvent{
		ID:                raw.ID,
		Type:              evType,
		Category:          actCategory,
		Title:             title,
		Description:       raw.Description,
		Timestamp:         raw.Timestamp,
		ActorType:         actorType,
		ActorName:         actorName,
		SourceType:        sourceType,
		SourceID:          sourceID,
		Metadata:          raw.Metadata,
		IsImportant:       isImportant,
		RequiresAction:    requiresAction,
		RelatedEntityType: "RFQ",
		RelatedEntityID:   fmt.Sprintf("%d", rfq.ID),
	}
}

// computeActivitySummary calculates aggregate counts from the normalized event list.
func computeActivitySummary(events []spec.ActivityEvent) spec.ActivitySummary {
	var (
		total          int
		customerCount  int
		opsCount       int
		aiCount        int
		reqCount       int
		docCount       int
		quoteCount     int
		actionReqCount int
		latestTime     *time.Time
	)

	total = len(events)
	for i, ev := range events {
		if i == 0 {
			t := ev.Timestamp
			latestTime = &t
		}

		switch ev.Category {
		case spec.ActivityCatCustomer:
			customerCount++
		case spec.ActivityCatOperations:
			opsCount++
		case spec.ActivityCatAI:
			aiCount++
		case spec.ActivityCatRequirements:
			reqCount++
		case spec.ActivityCatDocuments:
			docCount++
		case spec.ActivityCatQuotes:
			quoteCount++
		}

		if ev.RequiresAction {
			actionReqCount++
		}
	}

	return spec.ActivitySummary{
		TotalEvents:         total,
		CustomerEvents:      customerCount,
		OperationalEvents:   opsCount,
		AIEvents:            aiCount,
		RequirementsEvents:  reqCount,
		DocumentEvents:      docCount,
		QuoteEvents:         quoteCount,
		ActionRequiredCount: actionReqCount,
		LatestActivityAt:    latestTime,
	}
}

func formatStatusTitle(raw string) string {
	parts := strings.Split(raw, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		}
	}
	return strings.Join(parts, " ")
}
