package jobs

import (
	"context"
	"fmt"
	"log"
	"strings"

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

	// 4. Delegate lead scoring, reasoning, and trade intelligence to the Python AI Sidecar
	corrID := fmt.Sprintf("lead-worker-%d-%d", orgID, leadID)
	volStr := fmt.Sprintf("%d TEU", intel.MonthlyShippingVolume)
	suppliersStr := strings.Join(intel.TopSuppliers, ", ")
	sidecarReq := &ai.ScoreLeadRequest{
		LeadID:                leadID,
		OrgID:                 orgID,
		CompanyName:           intel.Name,
		Industry:              &intel.Industry,
		EstimatedRevenue:      &intel.EstimatedRevenue,
		EmployeeCount:         &intel.EmployeeCount,
		MonthlyShippingVolume: &volStr,
		TopSuppliers:          &suppliersStr,
		IsExporter:            &intel.IsExporter,
		CorrelationID:         corrID,
	}

	sidecarClient := ai.NewSidecarClient("", "")
	scoreResp, err := sidecarClient.ScoreLead(ctx, sidecarReq)
	if err != nil {
		log.Printf("Lead Worker Error: Python AI Sidecar lead scoring failed: %v", err)
		return
	}

	result := struct {
		Score          int32
		ResearchReport string
	}{
		Score:          scoreResp.Score,
		ResearchReport: scoreResp.ResearchReport,
	}

	// 8. Update the Lead in the database with the new AI Score and Report!
	status := "NEW"
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
