package agent

import (
	"context"
	"log"
	"strconv"

	"github.com/freel/backend/internal/ai"
	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/rates"
	"github.com/freel/backend/internal/rfq"
	"github.com/freel/backend/internal/rfq/spec"
)

// PricingAgent acts as a Senior Pricing Analyst.
// It subscribes to RFQ assignment events, fetches the best available rates
// from the Rate Intelligence Service (covering both spot and contract rates),
// and drafts a recommended quote for human review.
type PricingAgent struct {
	eventBus       events.Bus
	rfqService     rfq.BusinessLogic
	rateSvc        rates.Service // unified Rate Intelligence layer
	aiGateway      ai.Gateway
	promptManager  ai.PromptManager
	backendBaseURL string // e.g. "http://backend:8080" — no trailing slash
}

// NewPricingAgent creates a PricingAgent wired to the Rate Intelligence Service.
// backendBaseURL is the resolved base URL of the Go backend (e.g. GO_BACKEND_URL env var).
// The rateSvc replaces the old carrierProv — it transparently serves contract
// rates when available and falls back to live spot rates automatically.
func NewPricingAgent(eb events.Bus, rs rfq.BusinessLogic, rateSvc rates.Service, ag ai.Gateway, pm ai.PromptManager, backendBaseURL string) *PricingAgent {
	return &PricingAgent{
		eventBus:       eb,
		rfqService:     rs,
		rateSvc:        rateSvc,
		aiGateway:      ag,
		promptManager:  pm,
		backendBaseURL: backendBaseURL,
	}
}

// Start subscribes the agent to relevant events.
func (a *PricingAgent) Start() {
	// Listen for when an RFQ is assigned to pricing.
	a.eventBus.Subscribe(events.EventRFQAssigned, a.handleRFQAssigned)

	// Listen to RFQ Created as well to auto-assign for this demo
	a.eventBus.Subscribe(events.EventRFQCreated, a.handleRFQCreated)
}

func (a *PricingAgent) handleRFQCreated(e events.Event) {
	// For MVP, auto advance to Pricing Assigned to trigger the agent
	payload, ok := e.Payload.(map[string]interface{})
	if !ok {
		return
	}
	rfqID, _ := payload["rfq_id"].(int32)
	orgID, _ := payload["org_id"].(int32)

	// In a real app, Workflow engine would do this, but we'll do it here if Workflow isn't active
	_, _ = a.rfqService.AdvanceStage(context.Background(), orgID, rfqID, spec.StagePricingAssigned)
}

func (a *PricingAgent) handleRFQAssigned(e events.Event) {
	payload, ok := e.Payload.(map[string]interface{})
	if !ok {
		return
	}

	// Convert IDs cleanly
	var rfqID int32
	switch v := payload["rfq_id"].(type) {
	case int32:
		rfqID = v
	case int:
		rfqID = int32(v)
	case float64:
		rfqID = int32(v)
	}

	var orgID int32
	if val, ok := payload["org_id"].(int32); ok {
		orgID = val
	} else if val, ok := payload["org_id"].(int); ok {
		orgID = int32(val)
	} else if val, ok := payload["org_id"].(float64); ok {
		orgID = int32(val)
	} else {
		log.Printf("[PricingAgent] Error: org_id is missing in EventRFQAssigned payload: %v", payload)
		return
	}

	ctx := context.Background()

	// 1. Update status to COLLECTING_INFORMATION
	_ = a.rfqService.UpdateAgentStatus(ctx, orgID, rfqID, "COLLECTING_INFORMATION")

	// Get correlation ID if present in the event payload
	correlationID := ""
	if val, ok := payload["correlation_id"].(string); ok {
		correlationID = val
	}

	// Construct task payload
	taskPayload := map[string]interface{}{
		"rfq_id":         rfqID,
		"org_id":         orgID,
		"entity_type":    "RFQ",
		"entity_id":      strconv.Itoa(int(rfqID)),
		"correlation_id": correlationID,
		"callback_url":   a.backendBaseURL + "/internal/pricing/callback",
	}

	// Create AI task in generalized queue
	err := a.rfqService.CreateAITask(
		ctx,
		int64(orgID),
		"RFQ",
		strconv.Itoa(int(rfqID)),
		"PRICING_ANALYZE",
		taskPayload,
	)
	if err != nil {
		log.Printf("[PricingAgent] Failed to create AI task for RFQ %d: %v", rfqID, err)
		_ = a.rfqService.UpdateAgentStatus(ctx, orgID, rfqID, "ERROR")
		return
	}

	// 2. Set status to PROCESSING (meaning queued and worker will pick up shortly)
	_ = a.rfqService.UpdateAgentStatus(ctx, orgID, rfqID, "PROCESSING")
	log.Printf("[PricingAgent] Enqueued PRICING_ANALYZE task for RFQ %d in org %d", rfqID, orgID)
}
