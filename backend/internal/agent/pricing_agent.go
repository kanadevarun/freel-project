package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/freel/backend/internal/ai"
	"github.com/freel/backend/internal/carrier"
	"github.com/freel/backend/internal/common"
	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/rfq"
	"github.com/freel/backend/internal/rfq/spec"
)

// PricingAgent acts as a Senior Pricing Analyst.
type PricingAgent struct {
	eventBus      events.Bus
	rfqService    rfq.BusinessLogic
	carrierProv   carrier.CarrierProvider
	aiGateway     ai.Gateway
	promptManager ai.PromptManager
}

func NewPricingAgent(eb events.Bus, rs rfq.BusinessLogic, cp carrier.CarrierProvider, ag ai.Gateway, pm ai.PromptManager) *PricingAgent {
	return &PricingAgent{
		eventBus:      eb,
		rfqService:    rs,
		carrierProv:   cp,
		aiGateway:     ag,
		promptManager: pm,
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

	// Hardcode orgID 5 for MVP if not in payload (since it's not emitted in AdvanceStage)
	orgID := int32(5)

	ctx := context.Background()

	// 1. COLLECTING_INFORMATION
	_ = a.rfqService.UpdateAgentStatus(ctx, orgID, rfqID, StateCollectingInformation)
	
	rfqData, err := a.rfqService.GetRFQ(ctx, orgID, rfqID)
	if err != nil {
		log.Printf("[PricingAgent] Failed to fetch RFQ %d: %v", rfqID, err)
		_ = a.rfqService.UpdateAgentStatus(ctx, orgID, rfqID, StateError)
		return
	}

	origin := ""
	dest := ""
	if rfqData.Origin != nil { origin = *rfqData.Origin }
	if rfqData.Destination != nil { dest = *rfqData.Destination }

	// Fetch Carrier Rates
	rates, err := a.carrierProv.GetRates(ctx, origin, dest)
	if err != nil {
		log.Printf("[PricingAgent] Failed to fetch rates: %v", err)
		_ = a.rfqService.UpdateAgentStatus(ctx, orgID, rfqID, StateError)
		return
	}

	// 2. ANALYZING_DATA
	_ = a.rfqService.UpdateAgentStatus(ctx, orgID, rfqID, StateAnalyzing)
	
	ratesJSON, _ := json.Marshal(rates)
	itemsJSON, _ := json.Marshal(rfqData.Items)
	incoterms := ""
	if rfqData.Incoterms != nil { incoterms = *rfqData.Incoterms }
	targetDate := ""
	if rfqData.TargetDate != nil { targetDate = rfqData.TargetDate.Format("2006-01-02") }

	promptVars := map[string]interface{}{
		"Origin":      origin,
		"Destination": dest,
		"Incoterms":   incoterms,
		"TargetDate":  targetDate,
		"Items":       string(itemsJSON),
		"CarrierRates": string(ratesJSON),
	}

	prompt, err := a.promptManager.GetPrompt("pricing_analyst", promptVars)
	if err != nil {
		_ = a.rfqService.UpdateAgentStatus(ctx, orgID, rfqID, StateError)
		return
	}

	// 3. WAITING_FOR_LLM
	_ = a.rfqService.UpdateAgentStatus(ctx, orgID, rfqID, StateWaitingForLLM)

	responseStr, err := a.aiGateway.ExecutePrompt(ctx, prompt)
	if err != nil {
		_ = a.rfqService.UpdateAgentStatus(ctx, orgID, rfqID, StateError)
		return
	}

	// 4. GENERATING_DRAFT
	_ = a.rfqService.UpdateAgentStatus(ctx, orgID, rfqID, StateGeneratingDraft)

	// Parse AI Response
	type AIResponse struct {
		Recommendation common.AIRecommendation `json:"recommendation"`
		DraftQuote     struct {
			CarrierName           string  `json:"carrier_name"`
			BuyPrice              float64 `json:"buy_price"`
			SellPrice             float64 `json:"sell_price"`
			TransitTimeDays       int     `json:"transit_time_days"`
			ReliabilityScore      int     `json:"reliability_score"`
			HistoricalSuccessRate float64 `json:"historical_success_rate"`
		} `json:"draft_quote"`
	}

	var aiResp AIResponse
	if err := json.Unmarshal([]byte(responseStr), &aiResp); err != nil {
		log.Printf("[PricingAgent] Failed to parse JSON: %v", err)
		_ = a.rfqService.UpdateAgentStatus(ctx, orgID, rfqID, StateError)
		return
	}

	transitDays := aiResp.DraftQuote.TransitTimeDays
	reasoning := aiResp.Recommendation.Reason

	quote := &spec.Quote{
		RFQID:                 rfqID,
		CarrierName:           aiResp.DraftQuote.CarrierName,
		BuyPrice:              aiResp.DraftQuote.BuyPrice,
		SellPrice:             aiResp.DraftQuote.SellPrice,
		TransitTimeDays:       &transitDays,
		ReliabilityScore:      aiResp.DraftQuote.ReliabilityScore,
		HistoricalSuccessRate: aiResp.DraftQuote.HistoricalSuccessRate,
		IsRecommended:         true,
		AiReasoning:           &reasoning,
		Status:                "DRAFT", // Important! AI only drafts
	}

	err = a.rfqService.AddQuote(ctx, orgID, quote)
	if err != nil {
		_ = a.rfqService.UpdateAgentStatus(ctx, orgID, rfqID, StateError)
		return
	}

	// Emit custom agent event (Draft Quote Generated)
	a.eventBus.Publish(events.Event{
		Type:      events.EventType("agent.pricing.draft_ready"),
		Payload:   map[string]interface{}{"rfq_id": rfqID, "quote_id": quote.ID, "recommendation": aiResp.Recommendation},
	})

	// 5. WAITING_FOR_HUMAN
	_ = a.rfqService.UpdateAgentStatus(ctx, orgID, rfqID, StateWaitingForHuman)
	fmt.Printf("[PricingAgent] Successfully generated draft quote for RFQ %d\n", rfqID)
}
