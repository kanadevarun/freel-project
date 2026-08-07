package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/freel/backend/internal/ai"
	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/leads"
	"github.com/freel/backend/internal/leads/spec"
	"github.com/freel/backend/internal/trade_intel"
)

// leadWorker processes new leads in the background.
// Simple meaning: Think of this as an automated digital assistant. 
// It waits around until it hears "Hey, a new lead was created!". 
// When it hears that, it goes and researches the company, asks the AI to score it, 
// and updates the lead record with the new information.
type leadWorker struct {
	eventBus      events.Bus
	intelEngine   trade_intel.Engine
	aiGateway     ai.Gateway
	promptManager ai.PromptManager
	leadsSvc      leads.BusinessLogic
}

// NewLeadWorker creates a new background worker for leads.
func NewLeadWorker(
	eb events.Bus, 
	intel trade_intel.Engine, 
	gw ai.Gateway, 
	pm ai.PromptManager, 
	ls leads.BusinessLogic,
) Worker {
	return &leadWorker{
		eventBus:      eb,
		intelEngine:   intel,
		aiGateway:     gw,
		promptManager: pm,
		leadsSvc:      ls,
	}
}

// Start tells the worker to start listening for "LeadCreated" events.
func (w *leadWorker) Start() error {
	w.eventBus.Subscribe(events.EventLeadCreated, w.handleLeadCreated)
	log.Println("Lead Worker started: Listening for new leads...")
	return nil
}

// Stop shuts down the worker. (We don't need any special teardown for this simple version).
func (w *leadWorker) Stop() error {
	log.Println("Lead Worker stopped.")
	return nil
}

// handleLeadCreated is the actual job that runs when a new lead is detected.
// Simple meaning: This is the step-by-step instruction set for the digital assistant.
func (w *leadWorker) handleLeadCreated(event events.Event) {
	ctx := context.Background() // Background context since this is an async job

	// 1. Extract the data sent in the event (lead_id and org_id)
	payload, ok := event.Payload.(map[string]interface{})
	if !ok {
		log.Printf("Lead Worker Error: Invalid event payload format")
		return
	}

	leadIDFloat, ok1 := payload["lead_id"].(float64) // JSON unmarshals numbers to float64 by default in some setups, but here we pass int32 from Go. Let's be safe.
	var leadID int32
	if ok1 {
		leadID = int32(leadIDFloat)
	} else if idInt, ok2 := payload["lead_id"].(int32); ok2 {
		leadID = idInt
	} else if idInt2, ok3 := payload["lead_id"].(int); ok3 {
		leadID = int32(idInt2)
	} else {
		log.Printf("Lead Worker Error: Could not parse lead_id")
		return
	}

	orgIDFloat, ok1 := payload["org_id"].(float64)
	var orgID int32
	if ok1 {
		orgID = int32(orgIDFloat)
	} else if orgInt, ok2 := payload["org_id"].(int32); ok2 {
		orgID = orgInt
	} else if orgInt2, ok3 := payload["org_id"].(int); ok3 {
		orgID = int32(orgInt2)
	} else {
		log.Printf("Lead Worker Error: Could not parse org_id")
		return
	}

	log.Printf("Lead Worker processing Lead ID: %d...", leadID)

	// 2. Fetch the full lead details from the database so we know the Company Name
	lead, err := w.leadsSvc.GetLead(ctx, orgID, leadID)
	if err != nil {
		log.Printf("Lead Worker Error: Failed to fetch lead %d: %v", leadID, err)
		return
	}

	// 3. Research the company using the Trade Intelligence Engine
	intel, err := w.intelEngine.EnrichCompany(ctx, lead.CompanyName)
	if err != nil {
		log.Printf("Lead Worker Error: Failed to enrich company %s: %v", lead.CompanyName, err)
		return
	}

	// 4. Put the research data into a nice map that the Prompt Manager can read
	vars := map[string]interface{}{
		"CompanyName":           intel.Name,
		"Industry":              intel.Industry,
		"EstimatedRevenue":      intel.EstimatedRevenue,
		"EmployeeCount":         intel.EmployeeCount,
		"MonthlyShippingVolume": intel.MonthlyShippingVolume,
		"TopSuppliers":          intel.TopSuppliers,
		"IsExporter":            intel.IsExporter,
	}

	// 5. Ask the Prompt Manager to write the fill-in-the-blank prompt for the AI
	promptText, err := w.promptManager.GetPrompt("score_lead", vars)
	if err != nil {
		log.Printf("Lead Worker Error: Failed to generate prompt: %v", err)
		return
	}

	// 6. Send the prompt to the AI Gateway (like OpenAI or Claude)
	// ── CALLING THE AI ─────────────────────────────────────────────────────────
	// This sends our generated lead research prompt to the AI router.
	// It will run on Google Gemini first, automatically failing over to OpenAI (ChatGPT)
	// if Gemini fails. If no keys are set, it returns safe mock text.
	aiResponseStr, err := w.aiGateway.ExecutePrompt(ctx, promptText)
	if err != nil {
		log.Printf("Lead Worker Error: AI Gateway failed: %v", err)
		return
	}

	// 7. The AI returns a JSON string containing the score and research report. We need to decode it.
	// We expect: {"score": 85, "research_report": "This company is..."}
	type aiResult struct {
		Score          int32  `json:"score"`
		ResearchReport string `json:"research_report"`
	}
	
	var result aiResult
	if err := json.Unmarshal([]byte(aiResponseStr), &result); err != nil {
		log.Printf("Lead Worker Error: Failed to parse AI response JSON: %v. Raw Response: %s", err, aiResponseStr)
		// Default to something safe if the AI hallucinated invalid JSON
		result.Score = 0
		result.ResearchReport = fmt.Sprintf("Failed to parse AI response. Raw output: %s", aiResponseStr)
	}

	// 8. Update the Lead in the database with the new AI Score and Report!
	status := "QUALIFIED" // Let's assume if we scored it, it's qualified. Real logic could threshold this (e.g. > 50).
	if result.Score < 50 {
		status = "REJECTED"
	}

	updateReq := spec.UpdateLeadRequest{
		OrgID:            orgID,
		ID:               leadID,
		Status:           &status,
		AIScore:          &result.Score,
		AIResearchReport: &result.ResearchReport,
	}

	_, err = w.leadsSvc.UpdateLead(ctx, updateReq)
	if err != nil {
		log.Printf("Lead Worker Error: Failed to update lead %d: %v", leadID, err)
		return
	}

	// 9. Emit an event so the Timeline knows the lead was enriched.
	w.eventBus.Publish(events.Event{
		Type: events.EventLeadEnriched,
		Payload: map[string]interface{}{
			"lead_id": leadID,
			"org_id":  orgID,
			"score":   result.Score,
		},
	})

	log.Printf("Lead Worker successfully processed and scored Lead ID: %d (Score: %d)", leadID, result.Score)
}
