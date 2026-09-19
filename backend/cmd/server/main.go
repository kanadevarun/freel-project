package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"

	"github.com/freel/backend/internal/actions"
	"github.com/freel/backend/internal/activity"
	"github.com/freel/backend/internal/agent"
	"github.com/freel/backend/internal/ai"
	"github.com/freel/backend/internal/aitasks"
	"github.com/freel/backend/internal/approvals"
	auditPkg "github.com/freel/backend/internal/audit"
	auditRepoPkg "github.com/freel/backend/internal/audit/repository"
	auditSvcPkg "github.com/freel/backend/internal/audit/service"
	auditTransportPkg "github.com/freel/backend/internal/audit/transport"
	"github.com/freel/backend/internal/auth"
	"github.com/freel/backend/internal/automations"
	"github.com/freel/backend/internal/autonomy"
	"github.com/freel/backend/internal/billing"
	"github.com/freel/backend/internal/carrier"
	carrierRepoPkg "github.com/freel/backend/internal/carrier/repository"
	carrierSvcPkg "github.com/freel/backend/internal/carrier/service"
	carrierTransportPkg "github.com/freel/backend/internal/carrier/transport"
	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/config"
	bcontext "github.com/freel/backend/internal/context"
	"github.com/freel/backend/internal/contracts"
	"github.com/freel/backend/internal/contracts/contract_compliance_automation"
	"github.com/freel/backend/internal/copilot"
	"github.com/freel/backend/internal/customers"
	"github.com/freel/backend/internal/dashboard"
	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/documents"
	"github.com/freel/backend/internal/enterprise_autonomy"
	"github.com/freel/backend/internal/event_workflows"
	"github.com/freel/backend/internal/files"
	"github.com/freel/backend/internal/finance"
	"github.com/freel/backend/internal/governance"
	"github.com/freel/backend/internal/invoices"
	"github.com/freel/backend/internal/invoices/collections_automation"
	"github.com/freel/backend/internal/integrations"
	"github.com/freel/backend/internal/jobs"
	"github.com/freel/backend/internal/leads"
	"github.com/freel/backend/internal/memory"
	"github.com/freel/backend/internal/monitoring"
	"github.com/freel/backend/internal/notifications"
	"github.com/freel/backend/internal/orchestration"
	"github.com/freel/backend/internal/organization"
	"github.com/freel/backend/internal/outreach"
	"github.com/freel/backend/internal/predictions"
	"github.com/freel/backend/internal/pricing"
	"github.com/freel/backend/internal/quotations"
	"github.com/freel/backend/internal/rates"
	"github.com/freel/backend/internal/rbac"
	"github.com/freel/backend/internal/recommendations"
	"github.com/freel/backend/internal/reports"
	"github.com/freel/backend/internal/rfq"
	"github.com/freel/backend/internal/rfq/pricing_workflow"
	"github.com/freel/backend/internal/search"
	"github.com/freel/backend/internal/server"
	"github.com/freel/backend/internal/shipments"
	"github.com/freel/backend/internal/shipments/operations_automation"
	"github.com/freel/backend/internal/sportal"
	"github.com/freel/backend/internal/subscription"
	"github.com/freel/backend/internal/trade_intel"
	"github.com/freel/backend/internal/users"
	"github.com/freel/backend/internal/workflow"
	"github.com/freel/backend/internal/workforce"
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

	// Initialize AI Gateway routed through the Python AI Sidecar
	// Architectural rule: Go does not execute direct external LLM calls.
	// All AI reasoning, prompt orchestration, and model providers run inside Python AI Sidecar.
	sidecarClient := ai.NewSidecarClient("", "")
	sidecarProvider := ai.NewSidecarProvider(sidecarClient)
	aiProviders := map[string]ai.Provider{
		"mock":    ai.NewMockProvider(),
		"sidecar": sidecarProvider,
		"gemini":  sidecarProvider,
		"openai":  sidecarProvider,
	}
	aiRuntimeConfig, err := ai.LoadRuntimeConfigFromEnv()
	if err != nil {
		log.Printf("⚠️ AI Runtime Config warning: %v", err)
	}
	aiGateway := ai.NewGatewayWithConfig(aiProviders, aiRuntimeConfig, db)
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

	// Initialize Centralized Notification and Escalation Center (Phase 2 Task 2.8)
	notifRepo := notifications.NewRepository(db)
	notifEngine := notifications.NewEngine(db, notifRepo)
	notifSvc := notifications.NewService(notifRepo, notifEngine, eventBus, emailNotifSvc,
		notifications.WithEnvironment(cfg.AppEnv),
		notifications.WithServiceKey(cfg.InternalServiceToken),
	)
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
	approvalsSvc.SetDraftHandlers(
		func(ctx context.Context, orgID int64, draftID int64, actorName string, notes string) error {
			draft, err := leadsBL.GetDraftByID(ctx, orgID, draftID)
			if err != nil {
				return err
			}
			if draft == nil {
				return fmt.Errorf("draft ID %d not found", draftID)
			}
			_, err = leadsBL.ApproveClarificationDraft(ctx, orgID, draft.LeadID, draft.ParentInteractionID, actorName, notes)
			return err
		},
		func(ctx context.Context, orgID int64, draftID int64, actorName string, reason string) error {
			draft, err := leadsBL.GetDraftByID(ctx, orgID, draftID)
			if err != nil {
				return err
			}
			if draft == nil {
				return fmt.Errorf("draft ID %d not found", draftID)
			}
			return leadsBL.RejectClarificationDraft(ctx, orgID, draft.LeadID, draft.ParentInteractionID, actorName, reason)
		},
	)
	approvalsHandler := approvals.NewHandler(approvalsSvc)
	notifSvc.SetApprovalsService(approvalsSvc)

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

	// Initialize Centralized Action System (Task 0.5 Action Boundary)
	actionsRegistry := actions.NewRegistry()
	_ = actionsRegistry.Register(actions.NewGetShipmentAction(shipmentsSvc))
	_ = actionsRegistry.Register(actions.NewUpdateMilestoneAction(shipmentsSvc))
	_ = actionsRegistry.Register(actions.NewRefreshShipmentTrackingAction(shipmentsSvc))
	_ = actionsRegistry.Register(actions.NewCreateExceptionAction(shipmentsSvc))
	_ = actionsRegistry.Register(actions.NewUpdateETAAction(shipmentsSvc))
	_ = actionsRegistry.Register(actions.NewOperationsCallbackAction(shipmentsSvc))
	_ = actionsRegistry.Register(actions.NewSaveDraftQuotesAction(rfqBL))
	_ = actionsRegistry.Register(actions.NewApplySelectedRateAction(rfqBL))
	_ = actionsRegistry.Register(actions.NewCreateRFQFromEmailAction(leadsBL, rfqBL))
	_ = actionsRegistry.Register(actions.NewConvertLeadAction(leadsBL))
	_ = actionsRegistry.Register(actions.NewSendClarificationEmailAction(leadsBL))
	_ = actionsRegistry.Register(actions.NewIngestRatesAction(contractsSvc))
	_ = actionsRegistry.Register(actions.NewReviewExtractionAction(contractsSvc))
	_ = actionsRegistry.Register(actions.NewRecordComplianceDiscrepanciesAction(documentsSvc))
	_ = actionsRegistry.Register(actions.NewReconcileInvoiceAction(financeSvc))

	// Phase 5 Task 5.8: Autonomous Exception Resolution Actions
	_ = actionsRegistry.Register(actions.NewExceptionRecoveryAction("customs_broker_notification", "Notify customs broker to resolve compliance and hold discrepancies"))
	_ = actionsRegistry.Register(actions.NewExceptionRecoveryAction("carrier_inquiry", "Initiate high-priority carrier escalation and status inquiry"))
	_ = actionsRegistry.Register(actions.NewExceptionRecoveryAction("customer_advisory", "Send proactive advisory and ETA adjustment notice to customer"))
	_ = actionsRegistry.Register(actions.NewExceptionRecoveryAction("audit_log", "Record governed autonomous audit entry and supervisor notes"))
	_ = actionsRegistry.Register(actions.NewExceptionRecoveryAction("exceptions.execute_recovery", "Execute controlled recovery action for shipment exception"))

	// Unified Business Context & Intelligence Layer (Task 1.1, 1.2, 1.3, 1.4)
	contextSvc := bcontext.NewService(db, rbacSvc)
	_ = actionsRegistry.Register(actions.NewGetContextAction(contextSvc))
	_ = actionsRegistry.Register(actions.NewGetInsightAction(contextSvc))
	_ = actionsRegistry.Register(actions.NewGetCustomerIntelligenceAction(contextSvc))
	_ = actionsRegistry.Register(actions.NewGetRFQIntelligenceAction(contextSvc))
	_ = actionsRegistry.Register(actions.NewGetShipmentIntelligenceAction(contextSvc))
	_ = actionsRegistry.Register(actions.NewGetInvoiceIntelligenceAction(contextSvc))
	_ = actionsRegistry.Register(actions.NewGetContractIntelligenceAction(contextSvc))
	_ = actionsRegistry.Register(actions.NewGetCrossModuleInsightsAction(contextSvc))
	contextHandler := bcontext.NewHandler(contextSvc)

	actionsStore := actions.NewDBIdempotencyStore(db)
	actionsService := actions.NewService(actionsRegistry, actionsStore, rbacSvc, db)
	actionsHandler := actions.NewHandler(actionsService)

	// Wire Approvals and Centralized Action System (Task 0.6 Unified Approval & HITL Bridge)
	approvalsSvc.SetRBACService(rbacSvc)
	pricingHandler.SetApprovalsService(approvalsSvc)
	contractsSvc.SetApprovalsService(approvalsSvc)
	actionsService.SetApprovalsService(approvalsSvc)

	approvalsSvc.SetActionExecutor(func(ctx context.Context, orgID int64, actionName string, input []byte, actorName string, userID int64, approvalRef string) (map[string]interface{}, error) {
		var inputMap map[string]interface{}
		if len(input) > 0 {
			_ = json.Unmarshal(input, &inputMap)
		}
		execReq := actions.ActionExecutionRequest{
			ActionName:     actionName,
			OrgID:          orgID,
			Input:          inputMap,
			ActorType:      actions.ActorTypeUI,
			ActingUserID:   userID,
			IsConfirmed:    true, // Human approval satisfied the confirmation gate
			IdempotencyKey: fmt.Sprintf("approval-%s", approvalRef),
		}
		resp, err := actionsService.Execute(ctx, execReq)
		if err != nil {
			return nil, err
		}
		if !resp.Success {
			errMsg := "action execution failed"
			if resp.Error != nil && resp.Error.Message != "" {
				errMsg = resp.Error.Message
			}
			return nil, fmt.Errorf("%s", errMsg)
		}
		if resp.Data != nil {
			if dataMap, ok := resp.Data.(map[string]interface{}); ok {
				return dataMap, nil
			}
		}
		return map[string]interface{}{"status": "success"}, nil
	})

	approvalsSvc.SetResumeExecutor(func(ctx context.Context, orgID int64, threadID string, action string, notes string) error {
		if strings.HasPrefix(threadID, "rfq-") {
			rfqIDStr := strings.TrimPrefix(threadID, "rfq-")
			if action == "APPROVE" {
				payload := map[string]interface{}{
					"correlation_id": fmt.Sprintf("resume-%s-%d", threadID, time.Now().Unix()),
					"callback_url":   goBackendURL + "/internal/pricing/callback",
					"notes":          notes,
				}
				payloadBytes, _ := json.Marshal(payload)
				_, err := db.ExecContext(ctx, `
					INSERT INTO ai_processing_tasks (
						org_id, entity_type, entity_id, task_type, payload, status, created_at, updated_at
					) VALUES (
						?, 'RFQ', ?, 'PRICING_RESUME', ?, 'QUEUED', NOW(), NOW()
					)
				`, orgID, rfqIDStr, string(payloadBytes))
				return err
			}
			return nil
		}

		// Contract document resumption via sidecar aiBridge
		return aiBridge.TriggerResumption(ctx, contracts.ResumptionRequest{
			DocumentID:  threadID,
			OrgID:       orgID,
			Action:      action,
			Notes:       notes,
			CallbackURL: goBackendURL + "/internal/contracts/callback",
		})
	})

	// AI Tasks Subsystem (Task 0.7 Hardening)
	aiTasksRepo := aitasks.NewRepository(db)
	aiTasksSvc := aitasks.NewService(aiTasksRepo)
	aiTasksSvc.SetApprovalsDelegate(func(ctx context.Context, orgID int64, approvalID int64, actorName string, reason string) error {
		_, err := approvalsSvc.CancelRequest(ctx, orgID, approvalID, actorName, 0, reason)
		return err
	})
	aiTasksHandler := aitasks.NewHandler(aiTasksSvc)

	// AI Action & Recommendation Center Subsystem (Phase 2 Task 2.1 & 2.2)
	recommendationsRepo := recommendations.NewRepository(db)
	recommendationsGen := recommendations.NewGenerator(db, recommendationsRepo)
	recommendationsSvc := recommendations.NewService(recommendationsRepo, recommendationsGen, auditSvc)
	recommendationsSvc.SetApprovalsService(approvalsSvc)
	_ = actionsRegistry.Register(actions.NewListRecommendationsAction(recommendationsSvc))
	_ = actionsRegistry.Register(actions.NewGetRecommendationAction(recommendationsSvc))
	_ = actionsRegistry.Register(actions.NewGenerateFollowupDraftAction(recommendationsSvc))
	_ = actionsRegistry.Register(actions.NewCreateFollowupTaskAction(recommendationsSvc))
	recommendationsHandler := recommendations.NewHandler(recommendationsSvc)

	// Workflow Automation & Scheduled AI Jobs Subsystem (Phase 2 Task 2.7)
	automationsRepo := automations.NewRepository(db)
	automationsSvc := automations.NewService(automationsRepo, recommendationsGen, recommendationsRepo, auditSvc)
	automationsHandler := automations.NewHandler(automationsSvc)
	automationsScheduler := automations.NewScheduler(automationsRepo, automationsSvc, 30*time.Second)
	automationsScheduler.Start()

	// AI Memory & Personalization Subsystem (Phase 2 Task 2.10)
	memoryRepo := memory.NewRepository(db)
	memorySvc := memory.NewService(memoryRepo)
	memoryHandler := memory.NewHandler(memorySvc)

	// AI Performance, Cost, and Quality Monitoring Subsystem (Phase 2 Task 2.11)
	monitoringRepo := monitoring.NewRepository(db)
	monitoringSvc := monitoring.NewService(monitoringRepo)
	monitoringHandler := monitoring.NewHandler(monitoringSvc)

	// Controlled AI Workflow Execution and Action Orchestration (Phase 3 Task 3.2)
	orchestrationRepo := orchestration.NewRepository(db)
	orchestrationRegistry := orchestration.NewRegistry()
	orchestrationSidecar := orchestration.NewSidecarClient()
	orchestrationSvc := orchestration.NewService(orchestrationRepo, orchestrationRegistry, orchestrationSidecar, auditSvc, db)
	orchestrationSvc.SetActionsService(actionsService)
	orchestrationHandler := orchestration.NewHandler(orchestrationSvc)

	// RFQ-to-Quotation Automation and Intelligent Pricing Workflow (Phase 3 Task 3.4)
	pricingWorkflowRepo := pricing_workflow.NewRepository(db)
	pricingWorkflowSvc := pricing_workflow.NewService(db, pricingWorkflowRepo, rfqBL, approvalsSvc, orchestrationSvc, auditSvc)
	pricingWorkflowHandler := pricing_workflow.NewHandler(pricingWorkflowSvc)

	// Shipment Operations Automation and Intelligent Exception Response (Phase 3 Task 3.5)
	shipmentOpsRepo := operations_automation.NewRepository(db)
	shipmentOpsSvc := operations_automation.NewService(db, shipmentOpsRepo, shipmentsRepo, approvalsSvc, auditSvc)
	shipmentOpsHandler := operations_automation.NewHandler(shipmentOpsSvc)

	// Finance and Collections Automation and Intelligent Receivables Follow-Up (Phase 3 Task 3.6)
	collectionsAutomationRepo := collections_automation.NewRepository(db)
	collectionsAutomationSvc := collections_automation.NewService(db, collectionsAutomationRepo, invoicesRepo, approvalsSvc, auditSvc)
	collectionsAutomationHandler := collections_automation.NewHandler(collectionsAutomationSvc)

	// Contract and Compliance Automation and Intelligent Document Review (Phase 3 Task 3.7)
	contractComplianceRepo := contract_compliance_automation.NewRepository(db)
	contractComplianceSvc := contract_compliance_automation.NewService(db, contractComplianceRepo, approvalsSvc, auditSvc)
	contractComplianceHandler := contract_compliance_automation.NewHandler(contractComplianceSvc)

	// Event-Driven AI Workflows and Cross-Module Automation (Phase 3 Task 3.8)
	eventWorkflowsRepo := event_workflows.NewRepository(db)
	eventWorkflowsSvc := event_workflows.NewService(db, eventWorkflowsRepo, approvalsSvc, orchestrationRegistry, auditSvc)
	eventWorkflowsHandler := event_workflows.NewHandler(eventWorkflowsSvc)

	// AI Copilot Across Every Module (Phase 3 Task 3.10)
	copilotRepo := copilot.NewRepository(db)
	copilotSvc := copilot.NewService(copilotRepo)
	copilotHandler := copilot.NewHandler(copilotSvc)

	// Advanced Reporting and Forecasting Subsystem (Phase 3 Task 3.11)
	reportsRepo := reports.NewRepository(db)
	reportsSvc := reports.NewService(reportsRepo)
	reportsAdvancedHandler := reports.NewHandler(reportsSvc)

	// AI Governance and Production Controls Subsystem (Phase 3 Task 3.12)
	governanceRepo := governance.NewRepository(db)
	governanceSvc := governance.NewService(governanceRepo)
	governanceHandler := governance.NewHandler(governanceSvc)

	// Predictive Intelligence & Decision Support Foundation (Phase 4 Task 4.1)
	predictionsRepo := predictions.NewMySQLRepository(db.DB)
	predictionsSidecar := predictions.NewSidecarClient(aiSidecarURL, "")
	predictionsSvc := predictions.NewService(predictionsRepo, predictionsSidecar, db.DB)
	predictionsHandler := predictions.NewHandler(predictionsSvc)

	// Phase 5 Controlled Autonomy Foundation (Phase 5 Task 5.1)
	autonomyRepo := autonomy.NewMySQLRepository(db.DB)
	autonomySidecar := autonomy.NewSidecarClient(aiSidecarURL, "")
	autonomySvc := autonomy.NewService(autonomyRepo, autonomySidecar, actionsService, approvalsSvc, db.DB)
	autonomyHandler := autonomy.NewHandler(autonomySvc)

	// Phase 6.1 Multi-Agent Workforce Foundation
	workforceRepo := workforce.NewMySQLRepository(db.DB)
	_ = workforceRepo.SeedBaselineAgents(context.Background())
	workforceSidecar := workforce.NewSidecarClient(aiSidecarURL)
	workforceSvc := workforce.NewService(workforceRepo, workforceSidecar, actionsService, approvalsSvc, auditSvc)
	workforceHandler := workforce.NewHandler(workforceSvc)

	// Phase 7.1, 7.2, 7.3, 7.4, 7.5, 7.6 & 7.7 Enterprise Autonomous Platform Foundation
	enterpriseRepo := enterprise_autonomy.NewMySQLRepository(db)
	enterpriseSvc := enterprise_autonomy.NewService(enterpriseRepo, workforceSvc, actionsService, approvalsSvc, auditSvc)
	enterpriseResilienceSvc := enterprise_autonomy.NewEnterpriseResilienceService(enterpriseRepo, actionsService, approvalsSvc, auditSvc, db)
	enterpriseSvc.SetResilienceService(enterpriseResilienceSvc)
	shipmentLifecycleSvc := enterprise_autonomy.NewShipmentLifecycleService(enterpriseRepo, workforceSvc, actionsService, approvalsSvc, auditSvc, shipmentsSvc, predictionsSvc)
	commercialLifecycleSvc := enterprise_autonomy.NewCommercialLifecycleService(enterpriseRepo, workforceSvc, actionsService, approvalsSvc, auditSvc, predictionsSvc, shipmentLifecycleSvc, db)
	exceptionManagementSvc := enterprise_autonomy.NewExceptionManagementService(enterpriseRepo, workforceSvc, actionsService, approvalsSvc, auditSvc, predictionsSvc, db)
	customerRelationshipSvc := enterprise_autonomy.NewCustomerRelationshipService(enterpriseRepo, workforceSvc, actionsService, approvalsSvc, auditSvc, predictionsSvc, commercialLifecycleSvc, exceptionManagementSvc, db)
	revenueOptimizationSvc := enterprise_autonomy.NewRevenueOptimizationService(enterpriseRepo, workforceSvc, actionsService, approvalsSvc, auditSvc, predictionsSvc, commercialLifecycleSvc, exceptionManagementSvc, customerRelationshipSvc, db)
	contractComplianceRiskSvc := enterprise_autonomy.NewContractComplianceRiskService(enterpriseRepo, workforceSvc, actionsService, approvalsSvc, auditSvc, predictionsSvc, shipmentLifecycleSvc, commercialLifecycleSvc, exceptionManagementSvc, customerRelationshipSvc, revenueOptimizationSvc, db)
	eventMeshSvc := enterprise_autonomy.NewEnterpriseEventMeshService(enterpriseRepo, workforceSvc, actionsService, approvalsSvc, auditSvc, predictionsSvc, shipmentLifecycleSvc, commercialLifecycleSvc, exceptionManagementSvc, customerRelationshipSvc, revenueOptimizationSvc, contractComplianceRiskSvc, enterpriseResilienceSvc, db)
	enterpriseGovernanceSvc := enterprise_autonomy.NewEnterpriseGovernanceService(enterpriseRepo, actionsService, approvalsSvc, auditSvc, db)
	controlTowerSvc := enterprise_autonomy.NewEnterpriseControlTowerService(enterpriseRepo, workforceSvc, actionsService, approvalsSvc, auditSvc, predictionsSvc, shipmentLifecycleSvc, commercialLifecycleSvc, exceptionManagementSvc, customerRelationshipSvc, revenueOptimizationSvc, contractComplianceRiskSvc, eventMeshSvc, enterpriseGovernanceSvc, enterpriseResilienceSvc, db)
	enterpriseHandler := enterprise_autonomy.NewHandler(enterpriseSvc, shipmentLifecycleSvc, commercialLifecycleSvc, exceptionManagementSvc, customerRelationshipSvc, revenueOptimizationSvc, contractComplianceRiskSvc, eventMeshSvc, controlTowerSvc, enterpriseGovernanceSvc, enterpriseResilienceSvc)

	srv := server.NewServer(cfg, db, authService, rbacSvc, rbacHandler, usersHandler, orgHandler, leadsEndpoints, leadsEmailHandler, outreachEndpoints, rfqEndpoints, dashboardEndpoints, notifHandler, reportsEndpoints, ratesEndpoints, contractsHandler, pricingHandler, shipmentsEndpoints, shipmentsSvc, documentsHandler, financeHandler, billingHandler, subHandler, quotationsEndpoints, commercialContractsEndpoints, customersEndpoints, approvalsHandler, invoicesHandler, carrierHandler, searchHandler, auditHandler, actionsHandler, aiTasksHandler, contextHandler, recommendationsHandler, automationsHandler, memoryHandler, monitoringHandler, orchestrationHandler)

	srv.RegisterRFQPricingWorkflowRoutes(pricingWorkflowHandler)
	srv.RegisterShipmentOperationsAutomationRoutes(shipmentOpsHandler)
	srv.RegisterFinanceCollectionsAutomationRoutes(collectionsAutomationHandler)
	srv.RegisterContractComplianceAutomationRoutes(contractComplianceHandler)
	srv.RegisterEventWorkflowsRoutes(eventWorkflowsHandler)
	srv.RegisterCopilotRoutes(copilotHandler)
	srv.RegisterAdvancedReportsRoutes(reportsAdvancedHandler)
	srv.RegisterGovernanceRoutes(governanceHandler)
	srv.RegisterPredictionsRoutes(predictionsHandler)
	// External Integration Foundation & Gateway (Task 2.1)
	integrationsRepo := integrations.NewConfigRepository(db)
	integrationsIdempotency := integrations.NewIdempotencyManager(db)
	webhookSecCfg := integrations.DefaultWebhookSecurityConfig(cfg.AdminAPIKey)
	integrationsWebhookGateway := integrations.NewWebhookGateway(db, webhookSecCfg)
	integrationsHTTPClient := integrations.NewResilientHTTPClient(integrations.DefaultHTTPClientConfig())
	twilioProvider := integrations.NewTwilioNotificationProvider(db, integrationsRepo, integrationsHTTPClient)
	sesProvider := integrations.NewSESNotificationProvider(db, integrationsRepo, integrationsHTTPClient)
	compositeNotificationPv := integrations.NewCompositeNotificationProvider(twilioProvider, sesProvider)
	carrierTrackingPv := integrations.NewCarrierGatewayTrackingProvider(carrierIntegrationSvc, carrierIntegrationRepo)
	s3Provider := integrations.NewS3StorageProvider(db, integrationsRepo, integrationsHTTPClient)
	textractProvider := integrations.NewAWSTextractProvider(db, integrationsRepo, integrationsHTTPClient)
	integrationsSvc := integrations.NewGatewayService(
		db,
		integrationsRepo,
		compositeNotificationPv,
		carrierTrackingPv,
		s3Provider,
		integrationsWebhookGateway,
		integrationsIdempotency,
		integrationsHTTPClient,
		actionsService,
		textractProvider,
	)
	integrationsHandler := integrations.NewHandler(integrationsSvc)
	_ = actionsRegistry.Register(integrations.NewSendSMSAction(integrationsSvc))
	_ = actionsRegistry.Register(integrations.NewSendEmailAction(integrationsSvc))
	_ = actionsRegistry.Register(integrations.NewFetchTrackingAction(integrationsSvc))
	_ = actionsRegistry.Register(integrations.NewUploadDocumentAction(integrationsSvc))
	_ = actionsRegistry.Register(integrations.NewDownloadDocumentAction(integrationsSvc))
	_ = actionsRegistry.Register(integrations.NewExtractTextAction(integrationsSvc))

	srv.RegisterAutonomyRoutes(autonomyHandler)
	srv.RegisterWorkforceRoutes(workforceHandler)
	srv.RegisterEnterpriseAutonomyRoutes(enterpriseHandler)
	srv.RegisterIntegrationsRoutes(integrationsHandler)

	// SPortal Internal Platform Foundation (Task S1)
	sportalRepo := sportal.NewRepository(db)
	sportalSvc := sportal.NewServiceWithDeps(sportalRepo, cfg.Environment, filesSvc, emailNotifSvc, cfg.FrontendURL)
	sportalHandler := sportal.NewHandler(sportalSvc)
	srv.RegisterSPortalRoutes(sportalHandler)

	// Background non-blocking recovery of interrupted enterprise workflows
	go func() {
		report, err := enterpriseSvc.RecoverInterruptedWorkflows(context.Background())
		if err != nil {
			log.Printf("[EnterpriseAutonomy] Recovery warning: %v", err)
		} else if report != nil && report.ScannedCount > 0 {
			log.Printf("[EnterpriseAutonomy] Interrupted workflow recovery: %d scanned, %d recovered, %d blocked",
				report.ScannedCount, report.RecoveredCount, report.BlockedCount)
		}
	}()

	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
