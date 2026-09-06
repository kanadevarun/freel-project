package main

import (
	"context"
	"log"
	"os"
	"strings"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"

	"github.com/freel/backend/internal/activity"
	"github.com/freel/backend/internal/agent"
	"github.com/freel/backend/internal/ai"
	"github.com/freel/backend/internal/approvals"
	auditPkg "github.com/freel/backend/internal/audit"
	auditRepoPkg "github.com/freel/backend/internal/audit/repository"
	auditSvcPkg "github.com/freel/backend/internal/audit/service"
	auditTransportPkg "github.com/freel/backend/internal/audit/transport"
	"github.com/freel/backend/internal/auth"
	"github.com/freel/backend/internal/billing"
	"github.com/freel/backend/internal/carrier"
	carrierRepoPkg "github.com/freel/backend/internal/carrier/repository"
	carrierSvcPkg "github.com/freel/backend/internal/carrier/service"
	carrierTransportPkg "github.com/freel/backend/internal/carrier/transport"
	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/config"
	"github.com/freel/backend/internal/contracts"
	"github.com/freel/backend/internal/customers"
	"github.com/freel/backend/internal/dashboard"
	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/documents"
	"github.com/freel/backend/internal/files"
	"github.com/freel/backend/internal/finance"
	"github.com/freel/backend/internal/invoices"
	"github.com/freel/backend/internal/jobs"
	"github.com/freel/backend/internal/leads"
	"github.com/freel/backend/internal/notifications"
	"github.com/freel/backend/internal/organization"
	"github.com/freel/backend/internal/outreach"
	"github.com/freel/backend/internal/pricing"
	"github.com/freel/backend/internal/quotations"
	"github.com/freel/backend/internal/rates"
	"github.com/freel/backend/internal/rbac"
	"github.com/freel/backend/internal/reports"
	"github.com/freel/backend/internal/rfq"
	"github.com/freel/backend/internal/search"
	"github.com/freel/backend/internal/server"
	"github.com/freel/backend/internal/shipments"
	"github.com/freel/backend/internal/subscription"
	"github.com/freel/backend/internal/trade_intel"
	"github.com/freel/backend/internal/users"
	"github.com/freel/backend/internal/workflow"
)

func main() {
	cfg := config.LoadConfig()

	// Resolve the Go backend's own public/internal URL.
	// In production this should be the internal service address (e.g. http://backend:8080).
	// Falls back to localhost for local development.
	// Trailing slash is stripped so callback URLs never become "//internal/..."
	goBackendURL := strings.TrimRight(os.Getenv("GO_BACKEND_URL"), "/")
	if goBackendURL == "" {
		goBackendURL = "http://localhost:8080"
	}

	// Initialize database connection
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Successfully connected to MySQL database!")

	// Initialize Universal Audit Logs Module (Task 1 & Task 2)
	auditRepo := auditRepoPkg.NewMySQLRepository(db)
	auditSvc := auditSvcPkg.NewService(auditRepo, db)
	auditPkg.SetDefaultService(auditSvc)
	auditHandler := auditTransportPkg.NewHandler(auditSvc)

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

	orgRepo := organization.NewRepository(db)
	gmailProvider := organization.NewGmailProvider()

	// Initialize Leads Module
	leadsDL := leads.NewDataLayer(db)
	leadsBL := leads.NewBusinessLogic(leadsDL, eventBus, orgRepo, gmailProvider, aiGateway)
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

	carrierProvider := carrier.NewMockProvider()
	carrierService := carrier.NewService(carrierProvider)

	// ── Rate Intelligence & Management Service ───────────────────────────────
	// The rates package is the unified rate layer for the entire platform.
	// Structured identically to leads (spec, bl, dl, endpoints, transport).
	ratesDL := rates.NewDataLayer(db)
	spotNormalizer := rates.NewSpotNormalizer()
	rateSvc := rates.NewBusinessLogic(ratesDL, spotNormalizer, carrierService)
	ratesEndpoints := rates.NewAllRatesEndpoints(rateSvc)

	// ── Contract Intelligence Service ─────────────────────────────────────────
	var filesSvc files.Service
	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket != "" {
		log.Printf("Initializing S3 File Service with bucket: %s", s3Bucket)
		awsCfg, err := awsConfig.LoadDefaultConfig(context.Background(), awsConfig.WithRegion(cfg.AWSRegion))
		if err != nil {
			log.Fatalf("failed to load AWS config for S3: %v", err)
		}
		s3Client := s3.NewFromConfig(awsCfg)
		filesSvc = files.NewS3Service(s3Client, s3Bucket)
	} else {
		log.Println("Initializing Local File Service (uploads stored locally)")
		filesSvc = files.NewLocalService("./uploads", goBackendURL+"/uploads")
	}

	// AI sidecar URL — default to localhost for development
	aiSidecarURL := os.Getenv("AI_SIDECAR_URL")
	if aiSidecarURL == "" {
		aiSidecarURL = "http://localhost:8090"
	}

	contractsRepo := contracts.NewRepository(db)
	aiBridge := contracts.NewAIBridge(aiSidecarURL)
	contractsSvc := contracts.NewService(contractsRepo, filesSvc, aiBridge, rateSvc, goBackendURL+"/internal/contracts/callback")
	contractsHandler := contracts.NewHandler(contractsSvc)

	// Commercial Contracts (Task 20.1)
	commercialContractsDL := contracts.NewDataLayer(db)
	commercialContractsBL := contracts.NewBusinessLogic(commercialContractsDL)
	commercialContractsEndpoints := contracts.MakeServerEndpoints(commercialContractsBL)

	// Initialize Subscription (SaaS) Module first so it can be injected
	stripeClient := subscription.NewStripeClient(cfg)
	subRepo := subscription.NewRepository(db)
	subSvc := subscription.NewService(subRepo, stripeClient)
	entitlementSvc := subscription.NewEntitlementService(subRepo)
	subHandler := subscription.NewHandler(subSvc)

	rfqDL := rfq.NewDataLayer(db)
	rfqBL := rfq.NewBusinessLogic(rfqDL, eventBus, rateSvc, aiGateway, promptManager, entitlementSvc)
	rfqEndpoints := rfq.NewAllRFQEndpoints(rfqBL)

	leadsEmailHandler := leads.NewEmailHandler(leadsBL, rfqBL, goBackendURL)

	// In the workflow processor, the old rfqSvc was passed in.
	// Since the workflow processor is just using it right now for agent dispatch,
	// we will need to update the agent setup if needed. The pricing agent requires an rfq service interface.
	// Actually, the pricing agent is using `rfq.Service` interface. We will need to see what that is,
	// or we can pass rfqBL to it instead since they share similar methods (Wait, pricing agent needs to get RFQ and Quotes).
	// Let's pass rfqBL for now and see if it compiles (we might need to adapt it).

	// We will create the Pricing Agent
	// The rateSvc (Rate Intelligence Service) is the single source of rates.
	// It transparently serves contract rates when available, falling back to live spot.
	pricingAgent := agent.NewPricingAgent(eventBus, rfqBL, rateSvc, aiGateway, promptManager, goBackendURL)
	pricingAgent.Start()

	// Initialize Dashboard Module
	dashboardDL := dashboard.NewDataLayer(db)
	dashboardBL := dashboard.NewBusinessLogic(dashboardDL)
	dashboardEndpoints := dashboard.NewAllDashboardEndpoints(dashboardBL)

	// Initialize Notifications Module
	var emailNotifSvc notifications.Service
	if cfg.MailProvider == notifications.MailProviderSMTP {
		log.Println("📧 Notifications: Using SMTP Mail Provider")
		emailNotifSvc = notifications.NewSMTPService(
			cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword,
			notifications.DefaultFromEmail, notifications.DefaultTemplateDir, cfg.FrontendURL,
		)
	} else if cfg.MailProvider == notifications.MailProviderSES {
		log.Println("📧 Notifications: Using AWS SES Mail Provider")
		awsCfg, err := awsConfig.LoadDefaultConfig(context.Background(), awsConfig.WithRegion(cfg.AWSRegion))
		if err != nil {
			log.Fatalf("failed to load AWS config for SES: %v", err)
		}
		sesClient := ses.NewFromConfig(awsCfg)
		emailNotifSvc = notifications.NewSESService(sesClient, cfg.SESFromEmail, notifications.DefaultTemplateDir)
	} else {
		log.Println("📧 Notifications: Using Mock Provider")
		emailNotifSvc = notifications.NewMockInAppService(eventBus)
	}

	notifSvc := notifications.NewMockInAppService(eventBus)
	notifHandler := notifications.NewHandler(notifSvc)

	// Initialize Pricing Module
	pricingSvc := pricing.NewService(db)
	pricingHandler := pricing.NewHandler(pricingSvc, rfqBL, rateSvc)

	// Initialize Reports Module
	reportsDL := reports.NewDataLayer(db)
	reportsBL := reports.NewBusinessLogic(reportsDL)
	reportsEndpoints := reports.NewAllReportsEndpoints(reportsBL)

	// Initialize Shipments (Operations) Module
	shipmentsRepo := shipments.NewRepository(db)
	shipmentsSvc := shipments.NewService(shipmentsRepo, db, eventBus, goBackendURL)
	shipmentsEndpoints := shipments.NewAllShipmentsEndpoints(shipmentsSvc)

	// Initialize Tracking Auto-Refresh Scheduler (Task 17.7)
	trackingScheduler := shipments.NewTrackingRefreshScheduler(shipmentsSvc, shipmentsRepo)
	trackingScheduler.Start(context.Background())
	defer trackingScheduler.Stop()

	// Wire RFQWon event to auto-create Shipment synchronously (No goroutine, Group 4 fix)
	eventBus.Subscribe(events.EventRFQWon, func(e events.Event) {
		payload, ok := e.Payload.(map[string]interface{})
		if !ok {
			return
		}
		var rfqID int64
		switch v := payload["rfq_id"].(type) {
		case int64:
			rfqID = v
		case int32:
			rfqID = int64(v)
		case int:
			rfqID = int64(v)
		case float64:
			rfqID = int64(v)
		}
		if rfqID > 0 {
			_, err := shipmentsSvc.CreateFromRFQ(context.Background(), rfqID)
			if err != nil {
				log.Printf("[EventRFQWon Handler] Error creating shipment from RFQ %d: %v", rfqID, err)
			}
		}
	})

	// Initialize Documents (Compliance) Module
	documentsRepo := documents.NewRepository(db)
	documentsSvc := documents.NewService(documentsRepo, db, goBackendURL, filesSvc)
	documentsHandler := documents.NewHandler(documentsSvc)

	// Initialize Finance (Reconciliation) Module
	financeRepo := finance.NewRepository(db)
	financeSvc := finance.NewService(financeRepo, db, goBackendURL)
	financeHandler := finance.NewHandler(financeSvc)

	// Initialize Billing (Customer Invoicing & Margins) Module
	billingRepo := billing.NewRepository(db)
	billingSvc := billing.NewService(billingRepo)
	billingHandler := billing.NewHandler(billingSvc)

	// Initialize Quotations Module (Task 18)
	quotationRepo := quotations.NewRepository(db)
	quotationSvc := quotations.NewService(quotationRepo, rateSvc)
	quotationsEndpoints := quotations.NewAllQuotationEndpoints(quotationSvc)

	// Initialize Users & RBAC Handlers
	usersRepo := users.NewRepository(db)
	usersSvc := users.NewService(usersRepo, emailNotifSvc)
	usersHandler := users.NewHandler(usersSvc)

	rbacHandler := rbac.NewHandler(rbacSvc)

	// Initialize Organization Module
	orgSvc := organization.NewService(orgRepo, eventBus, filesSvc)
	orgHandler := organization.NewHandler(orgSvc)

	// Start Mailbox Sync Worker
	mailboxSyncWorker := jobs.NewMailboxSyncWorker(db, leadsBL, orgRepo)
	orgSvc.SetSyncNowFunc(mailboxSyncWorker.SyncMailboxNow)
	mailboxSyncWorker.Start()

	// Start Carrier Poller background scheduler
	carrierPoller := jobs.NewCarrierPoller(db, shipmentsSvc)
	carrierPoller.Start()

	// Initialize Customers Module (Task 21.1)
	customersDL := customers.NewDataLayer(db)
	customersBL := customers.NewBusinessLogic(customersDL)
	customersEndpoints := customers.MakeEndpoints(customersBL)

	// Initialize Approvals Module
	approvalsRepo := approvals.NewRepository(db)
	approvalsSvc := approvals.NewService(approvalsRepo)
	approvalsHandler := approvals.NewHandler(approvalsSvc)

	// Initialize Invoices Module
	invoicesRepo := invoices.NewRepository(db)
	invoicesSvc := invoices.NewService(invoicesRepo)
	invoicesHandler := invoices.NewHandler(invoicesSvc)

	// Initialize Carrier Integrations Module (Foundation Architecture)
	carrierIntegrationRepo := carrierRepoPkg.NewCarrierRepository(db)
	carrierIntegrationSvc := carrierSvcPkg.NewCarrierService(carrierIntegrationRepo, cfg.MailboxEncryptionKey)
	carrierHandler := carrierTransportPkg.NewCarrierHandler(carrierIntegrationSvc)

	// Wire Carrier Integration Engine into Modules (Task 4, 5, 6)
	carrierIntegrationSvc.SetDB(db)
	shipmentsSvc.SetCarrierService(carrierIntegrationSvc)
	rateSvc.SetCarrierIntegrationService(carrierIntegrationSvc)
	rfqBL.SetCarrierIntegrationService(carrierIntegrationSvc)

	// Wire synchronization delegates for Tracking & Bookings
	carrierIntegrationSvc.SetTrackingSyncer(func(ctx context.Context, orgID int64, shipmentID int64) (int, error) {
		res, err := shipmentsSvc.RefreshShipmentTracking(ctx, orgID, shipmentID, nil, "CARRIER_SYNC_ENGINE")
		if err != nil {
			return 0, err
		}
		if res != nil {
			return res.NewEvents, nil
		}
		return 0, nil
	})
	carrierIntegrationSvc.SetBookingSyncer(func(ctx context.Context, orgID int64, bookingID int64) error {
		_, err := rfqBL.SyncCarrierBooking(ctx, int32(orgID), bookingID, "CARRIER_SYNC_ENGINE")
		return err
	})

	// Start Carrier Sync Worker (Task 6 Background Sync)
	carrierSyncWorker := jobs.NewCarrierSyncWorker(db, carrierIntegrationSvc)
	carrierSyncWorker.Start()

	// Initialize Global Search Module
	searchRepo := search.NewRepository(db)
	searchSvc := search.NewService(searchRepo)
	searchHandler := search.NewHandler(searchSvc)

	srv := server.NewServer(cfg, db, authService, rbacSvc, rbacHandler, usersHandler, orgHandler, leadsEndpoints, leadsEmailHandler, outreachEndpoints, rfqEndpoints, dashboardEndpoints, notifHandler, reportsEndpoints, ratesEndpoints, contractsHandler, pricingHandler, shipmentsEndpoints, shipmentsSvc, documentsHandler, financeHandler, billingHandler, subHandler, quotationsEndpoints, commercialContractsEndpoints, customersEndpoints, approvalsHandler, invoicesHandler, carrierHandler, searchHandler, auditHandler)

	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
