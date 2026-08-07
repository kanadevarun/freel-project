package main

import (
	"context"
	"log"

	"github.com/freel/backend/internal/activity"
	"github.com/freel/backend/internal/agent"
	"github.com/freel/backend/internal/ai"
	"github.com/freel/backend/internal/auth"
	"github.com/freel/backend/internal/carrier"
	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/config"
	"github.com/freel/backend/internal/dashboard"
	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/jobs"
	"github.com/freel/backend/internal/leads"
	"github.com/freel/backend/internal/notifications"
	"github.com/freel/backend/internal/outreach"
	"github.com/freel/backend/internal/rbac"
	"github.com/freel/backend/internal/reports"
	"github.com/freel/backend/internal/rfq"
	"github.com/freel/backend/internal/server"
	"github.com/freel/backend/internal/trade_intel"
	"github.com/freel/backend/internal/workflow"
)

func main() {
	cfg := config.LoadConfig()

	// Initialize database connection
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	log.Println("Successfully connected to Postgres database!")

	// Initialize Event Bus
	eventBus := events.NewInProcessBus()

	// Initialize RBAC and seed system permissions
	rbacSvc := rbac.NewService(db, eventBus)
	if err := rbacSvc.SeedSystemPermissions(context.Background()); err != nil {
		log.Fatalf("Failed to seed system permissions: %v", err)
	}

	// Initialize Trade Intel Engine
	tradeIntelEngine := trade_intel.NewMockEngine()

	// Initialize AI Gateway (dynamically registers active keys from .env)
	aiProviders := map[string]ai.Provider{
		"mock": ai.NewMockProvider(),
	}
	if cfg.GeminiAPIKey != "" {
		log.Println("🤖 AI: Registering Google Gemini primary provider...")
		aiProviders["gemini"] = ai.NewGeminiProvider(cfg.GeminiAPIKey)
	}
	if cfg.OpenAIAPIKey != "" {
		log.Println("🤖 AI: Registering OpenAI ChatGPT failover provider...")
		aiProviders["openai"] = ai.NewOpenAIProvider(cfg.OpenAIAPIKey)
	}
	aiGateway := ai.NewGateway(aiProviders)
	promptManager := ai.NewPromptManager()

	// Initialize Leads Module
	leadsDL := leads.NewDataLayer(db)
	leadsBL := leads.NewBusinessLogic(leadsDL, eventBus)
	leadsEndpoints := leads.NewAllLeadsEndpoints(leadsBL)

	// Initialize Timeline Service (also registers event listeners)
	_ = activity.NewTimelineService(db, eventBus)

	// Initialize Background Workers
	leadWorker := jobs.NewLeadWorker(eventBus, tradeIntelEngine, aiGateway, promptManager, leadsBL)
	if err := leadWorker.Start(); err != nil {
		log.Fatalf("Failed to start lead worker: %v", err)
	}

	authService := auth.NewService(cfg, db)

	// Initialize Outreach Module
	outreachDL := outreach.NewDataLayer(db)
	outreachBL := outreach.NewBusinessLogic(outreachDL)
	outreachEndpoints := outreach.NewAllOutreachEndpoints(outreachBL)

	// Initialize Workflow Engine
	assigner := workflow.NewAssigner()
	workflowEngine := workflow.NewEngine(assigner)

	// Wire RFQCreated event to Workflow Engine
	eventBus.Subscribe(events.EventRFQCreated, func(e events.Event) {
		payload, ok := e.Payload.(map[string]interface{})
		if !ok {
			return
		}
		rfqIDFloat, ok := payload["rfq_id"].(float64)
		if !ok {
			// If it's passed directly as int32 or int, handle appropriately
			switch v := payload["rfq_id"].(type) {
			case int32:
				rfqIDFloat = float64(v)
			case int:
				rfqIDFloat = float64(v)
			}
		}
		
		err := workflowEngine.ProcessEvent(context.Background(), string(e.Type), payload)
		if err != nil {
			log.Printf("Workflow engine failed to process event %s: %v", e.Type, err)
			return
		}

		// In a full implementation, we'd take the assignee returned by ProcessEvent 
		_ = rfqIDFloat
		log.Printf("Workflow successfully evaluated routing for RFQ %v", payload["rfq_id"])
	})

	// Initialize RFQ Module
	// carrierService wraps the mock provider and adds FetchRates ranking logic.
	// When the FF partner API is ready, swap NewMockProvider() for the real adapter.
	carrierProvider := carrier.NewMockProvider()
	carrierService := carrier.NewService(carrierProvider)

	rfqDL := rfq.NewDataLayer(db)
	rfqBL := rfq.NewBusinessLogic(rfqDL, eventBus, carrierService)
	rfqEndpoints := rfq.NewAllRFQEndpoints(rfqBL)

	// In the workflow processor, the old rfqSvc was passed in.
	// Since the workflow processor is just using it right now for agent dispatch, 
	// we will need to update the agent setup if needed. The pricing agent requires an rfq service interface.
	// Actually, the pricing agent is using `rfq.Service` interface. We will need to see what that is,
	// or we can pass rfqBL to it instead since they share similar methods (Wait, pricing agent needs to get RFQ and Quotes).
	// Let's pass rfqBL for now and see if it compiles (we might need to adapt it).

	// We will create the Pricing Agent
	// The carrier provider is already initialized above as part of the RFQ module.
	pricingAgent := agent.NewPricingAgent(eventBus, rfqBL, carrierProvider, aiGateway, promptManager)
	pricingAgent.Start()

	// Initialize Dashboard Module
	dashboardDL := dashboard.NewDataLayer(db)
	dashboardBL := dashboard.NewBusinessLogic(dashboardDL)
	dashboardEndpoints := dashboard.NewAllDashboardEndpoints(dashboardBL)

	// Initialize Notifications Module
	notifSvc := notifications.NewMockInAppService(eventBus)
	notifHandler := notifications.NewHandler(notifSvc)

	// Initialize Reports Module
	reportsDL := reports.NewDataLayer(db)
	reportsBL := reports.NewBusinessLogic(reportsDL)
	reportsEndpoints := reports.NewAllReportsEndpoints(reportsBL)

	srv := server.NewServer(cfg, authService, rbacSvc, leadsEndpoints, outreachEndpoints, rfqEndpoints, dashboardEndpoints, notifHandler, reportsEndpoints)

	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
