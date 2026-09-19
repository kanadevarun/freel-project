package server

import (
	"log"
	"net/http"

	"github.com/freel/backend/internal/actions"
	"github.com/freel/backend/internal/aitasks"
	"github.com/freel/backend/internal/approvals"
	auditTransport "github.com/freel/backend/internal/audit/transport"
	"github.com/freel/backend/internal/auth"
	"github.com/freel/backend/internal/automations"
	"github.com/freel/backend/internal/autonomy"
	"github.com/freel/backend/internal/billing"
	carrierTransport "github.com/freel/backend/internal/carrier/transport"
	"github.com/freel/backend/internal/config"
	bcontext "github.com/freel/backend/internal/context"
	"github.com/freel/backend/internal/contracts"
	"github.com/freel/backend/internal/contracts/contract_compliance_automation"
	"github.com/freel/backend/internal/copilot"
	"github.com/freel/backend/internal/customers"
	"github.com/freel/backend/internal/dashboard"
	"github.com/freel/backend/internal/documents"
	"github.com/freel/backend/internal/enterprise_autonomy"
	"github.com/freel/backend/internal/event_workflows"
	"github.com/freel/backend/internal/finance"
	"github.com/freel/backend/internal/governance"
	"github.com/freel/backend/internal/invoices"
	"github.com/freel/backend/internal/invoices/collections_automation"
	"github.com/freel/backend/internal/integrations"
	"github.com/freel/backend/internal/leads"
	"github.com/freel/backend/internal/memory"
	"github.com/freel/backend/internal/middleware"
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
	"github.com/freel/backend/internal/shipments"
	"github.com/freel/backend/internal/shipments/operations_automation"
	"github.com/freel/backend/internal/sportal"
	"github.com/freel/backend/internal/subscription"
	"github.com/freel/backend/internal/users"
	"github.com/freel/backend/internal/workforce"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	cfg                          *config.Config
	db                           *sqlx.DB
	router                       *chi.Mux
	authService                  *auth.Service
	rbacSvc                      rbac.Service
	rbacHandler                  *rbac.Handler
	usersHandler                 *users.Handler
	sportalHandler               *sportal.Handler
	orgHandler                   *organization.Handler
	leadsEndpoints               leads.Endpoints
	leadsEmailHandler            *leads.EmailHandler
	outreachEndpoints            outreach.Endpoints
	rfqEndpoints                 rfq.Endpoints
	dashboardEndpoints           dashboard.Endpoints
	notificationsHandler         *notifications.Handler
	reportsEndpoints             reports.Endpoints
	ratesEndpoints               rates.Endpoints
	contractsHandler             *contracts.Handler
	pricingHandler               *pricing.Handler
	shipmentsEndpoints           shipments.Endpoints
	shipmentsSvc                 shipments.Service
	documentsHandler             *documents.Handler
	financeHandler               *finance.Handler
	billingHandler               *billing.Handler
	subscriptionHandler          *subscription.Handler
	quotationsEndpoints          quotations.Endpoints
	commercialContractsEndpoints contracts.Endpoints
	customersEndpoints           customers.Endpoints
	approvalsHandler             *approvals.Handler
	invoicesHandler              *invoices.Handler
	carrierHandler               *carrierTransport.CarrierHandler
	searchHandler                *search.Handler
	auditHandler                 *auditTransport.Handler
	actionsHandler               *actions.Handler
	aiTasksHandler               *aitasks.Handler
	contextHandler               *bcontext.Handler
	recommendationsHandler       *recommendations.Handler
	automationsHandler           *automations.Handler
	memoryHandler                *memory.Handler
	monitoringHandler           *monitoring.Handler
	orchestrationHandler        *orchestration.Handler
}

func NewServer(cfg *config.Config, db *sqlx.DB, authService *auth.Service, rbacSvc rbac.Service, rbacHandler *rbac.Handler, usersHandler *users.Handler, orgHandler *organization.Handler, leadsEndpoints leads.Endpoints, leadsEmailHandler *leads.EmailHandler, outreachEndpoints outreach.Endpoints, rfqEndpoints rfq.Endpoints, dashboardEndpoints dashboard.Endpoints, notifHandler *notifications.Handler, reportsEndpoints reports.Endpoints, ratesEndpoints rates.Endpoints, contractsHandler *contracts.Handler, pricingHandler *pricing.Handler, shipmentsEndpoints shipments.Endpoints, shipmentsSvc shipments.Service, documentsHandler *documents.Handler, financeHandler *finance.Handler, billingHandler *billing.Handler, subscriptionHandler *subscription.Handler, quotationsEndpoints quotations.Endpoints, commercialContractsEndpoints contracts.Endpoints, customersEndpoints customers.Endpoints, approvalsHandler *approvals.Handler, invoicesHandler *invoices.Handler, carrierHandler *carrierTransport.CarrierHandler, searchHandler *search.Handler, auditHandler *auditTransport.Handler, actionsHandler *actions.Handler, aiTasksHandler *aitasks.Handler, contextHandler *bcontext.Handler, recommendationsHandler *recommendations.Handler, automationsHandler *automations.Handler, memoryHandler *memory.Handler, monitoringHandler *monitoring.Handler, orchestrationHandler *orchestration.Handler) *Server {
	s := &Server{
		cfg:                          cfg,
		db:                           db,
		router:                       chi.NewRouter(),
		authService:                  authService,
		rbacSvc:                      rbacSvc,
		rbacHandler:                  rbacHandler,
		usersHandler:                 usersHandler,
		orgHandler:                   orgHandler,
		leadsEndpoints:               leadsEndpoints,
		leadsEmailHandler:            leadsEmailHandler,
		outreachEndpoints:            outreachEndpoints,
		rfqEndpoints:                 rfqEndpoints,
		dashboardEndpoints:           dashboardEndpoints,
		notificationsHandler:         notifHandler,
		reportsEndpoints:             reportsEndpoints,
		ratesEndpoints:               ratesEndpoints,
		contractsHandler:             contractsHandler,
		pricingHandler:               pricingHandler,
		shipmentsEndpoints:           shipmentsEndpoints,
		shipmentsSvc:                 shipmentsSvc,
		documentsHandler:             documentsHandler,
		financeHandler:               financeHandler,
		billingHandler:               billingHandler,
		subscriptionHandler:          subscriptionHandler,
		quotationsEndpoints:          quotationsEndpoints,
		commercialContractsEndpoints: commercialContractsEndpoints,
		customersEndpoints:           customersEndpoints,
		approvalsHandler:             approvalsHandler,
		invoicesHandler:              invoicesHandler,
		carrierHandler:               carrierHandler,
		searchHandler:                searchHandler,
		auditHandler:                 auditHandler,
		actionsHandler:               actionsHandler,
		aiTasksHandler:               aiTasksHandler,
		contextHandler:               contextHandler,
		recommendationsHandler:       recommendationsHandler,
		automationsHandler:           automationsHandler,
		memoryHandler:                memoryHandler,
		monitoringHandler:           monitoringHandler,
		orchestrationHandler:        orchestrationHandler,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) Start() error {
	log.Printf("Server starting on port %s", s.cfg.Port)
	return http.ListenAndServe(":"+s.cfg.Port, s.router)
}

func (s *Server) newAuthGuard() *middleware.AuthMiddleware {
	env := s.cfg.AppEnv
	if env == "" {
		env = s.cfg.Environment
	}
	return middleware.NewAuthMiddleware(s.cfg.AWSRegion, s.cfg.CognitoUserPoolID, s.db, middleware.WithEnvironment(env))
}

// RegisterRFQPricingWorkflowRoutes mounts the Phase 3 Task 3.4 RFQ pricing workflow endpoints
func (s *Server) RegisterRFQPricingWorkflowRoutes(h *pricing_workflow.Handler) {
	if h == nil {
		return
	}
	authGuard := s.newAuthGuard()
	s.router.Route("/api/v1/rfqs/{id:[0-9]+}/pricing-workflow", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		h.RegisterRoutes(r)
	})
}

// RegisterShipmentOperationsAutomationRoutes mounts the Phase 3 Task 3.5 Shipment Operations Automation endpoints
func (s *Server) RegisterShipmentOperationsAutomationRoutes(h *operations_automation.Handler) {
	if h == nil {
		return
	}
	authGuard := s.newAuthGuard()
	s.router.Route("/api/v1/shipments/{id:[0-9]+}/operations-automation", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		h.RegisterRoutes(r)
	})
}

// RegisterFinanceCollectionsAutomationRoutes mounts the Phase 3 Task 3.6 Finance & Collections Automation endpoints
func (s *Server) RegisterFinanceCollectionsAutomationRoutes(h *collections_automation.Handler) {
	if h == nil {
		return
	}
	authGuard := s.newAuthGuard()
	s.router.Route("/api/v1/invoices/{id:[0-9]+}/collections-automation", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		h.RegisterRoutes(r)
	})
}

// RegisterContractComplianceAutomationRoutes mounts the Phase 3 Task 3.7 Contract & Compliance Automation endpoints
func (s *Server) RegisterContractComplianceAutomationRoutes(h *contract_compliance_automation.Handler) {
	if h == nil {
		return
	}
	authGuard := s.newAuthGuard()
	s.router.Route("/api/v1/contracts/{id:[0-9]+}/compliance-automation", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		h.RegisterRoutes(r)
	})
}

// RegisterEventWorkflowsRoutes mounts the Phase 3 Task 3.8 Event-Driven AI Workflows endpoints
func (s *Server) RegisterEventWorkflowsRoutes(h *event_workflows.Handler) {
	if h == nil {
		return
	}
	authGuard := s.newAuthGuard()
	s.router.Route("/api/v1/event-workflows", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		h.RegisterRoutes(r)
	})
}

// RegisterCopilotRoutes mounts the Phase 3 Task 3.10 AI Copilot Across Every Module endpoints
func (s *Server) RegisterCopilotRoutes(h *copilot.Handler) {
	if h == nil {
		return
	}
	authGuard := s.newAuthGuard()
	s.router.Route("/api/v1/copilot", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		h.RegisterRoutes(r)
	})
}

// RegisterAdvancedReportsRoutes mounts the Phase 3 Task 3.11 Advanced Reporting & Forecasting endpoints
func (s *Server) RegisterAdvancedReportsRoutes(h *reports.Handler) {
	if h == nil {
		return
	}
	authGuard := s.newAuthGuard()
	s.router.Route("/api/v1/reports", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		h.RegisterRoutes(r)
	})
}

// RegisterGovernanceRoutes mounts the Phase 3 Task 3.12 AI Governance and Production Controls endpoints
func (s *Server) RegisterGovernanceRoutes(h *governance.Handler) {
	if h == nil {
		return
	}
	authGuard := s.newAuthGuard()
	s.router.Route("/api/v1/governance", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		h.RegisterRoutes(r)
	})
}

// RegisterPredictionsRoutes mounts the Phase 4 Predictive Intelligence & Decision Support endpoints
func (s *Server) RegisterPredictionsRoutes(h *predictions.Handler) {
	if h == nil {
		return
	}
	authGuard := s.newAuthGuard()
	s.router.Route("/api/v1/predictions", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		h.RegisterRoutes(r)
	})
	s.router.Route("/api/v1/shipments/{id:[0-9]+}/predicted-eta", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetShipmentPredictedETA)
		r.Post("/refresh", h.HandleRefreshShipmentPredictedETA)
	})
	s.router.Route("/api/v1/shipments/{id:[0-9]+}/predicted-exceptions", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetShipmentPredictedExceptions)
		r.Post("/refresh", h.HandleRefreshShipmentPredictedExceptions)
	})
	s.router.Route("/api/v1/shipments/{id:[0-9]+}/predicted-readiness", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetShipmentPredictedReadiness)
		r.Post("/refresh", h.HandleRefreshShipmentPredictedReadiness)
	})
	s.router.Route("/api/v1/leads/{id:[0-9]+}/predicted-intelligence", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetLeadPredictedIntelligence)
		r.Post("/refresh", h.HandleRefreshLeadPredictedIntelligence)
	})
	s.router.Route("/api/v1/customers/{id:[0-9]+}/predicted-intelligence", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetCustomerPredictedIntelligence)
		r.Post("/refresh", h.HandleRefreshCustomerPredictedIntelligence)
	})
	s.router.Route("/api/v1/rfqs/{id:[0-9]+}/predicted-margin", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetRFQPredictedMargin)
		r.Post("/refresh", h.HandleRefreshRFQPredictedMargin)
	})
	s.router.Route("/api/v1/contracts/{id:[0-9]+}/predicted-rate-pressure", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetContractPredictedRatePressure)
		r.Post("/refresh", h.HandleRefreshContractPredictedRatePressure)
	})
	s.router.Route("/api/v1/contracts/{id:[0-9]+}/predicted-compliance", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetContractPredictedComplianceRisk)
		r.Post("/refresh", h.HandleRefreshContractPredictedComplianceRisk)
	})
	s.router.Route("/api/v1/invoices/{id:[0-9]+}/predicted-collection", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetInvoicePredictedCollection)
		r.Post("/refresh", h.HandleRefreshInvoicePredictedCollection)
	})
	s.router.Route("/api/v1/carriers/{scac:[A-Za-z0-9_]+}/predicted-performance", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetCarrierPredictedPerformance)
		r.Post("/refresh", h.HandleRefreshCarrierPredictedPerformance)
	})
	s.router.Route("/api/v1/network/lanes/{laneCode:[A-Za-z0-9_-]+}/predicted-performance", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetLanePredictedPerformance)
		r.Post("/refresh", h.HandleRefreshLanePredictedPerformance)
	})
	s.router.Route("/api/v1/customers/{id:[0-9]+}/predicted-performance", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetCustomerPredictedServicePerformance)
		r.Post("/refresh", h.HandleRefreshCustomerPredictedServicePerformance)
	})
	s.router.Route("/api/v1/workload/predicted-workload", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetPredictedWorkload)
		r.Post("/refresh", h.HandleRefreshPredictedWorkload)
	})
	s.router.Route("/api/v1/capacity/predicted-capacity", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetPredictedCapacity)
		r.Post("/refresh", h.HandleRefreshPredictedCapacity)
	})
	s.router.Route("/api/v1/demand/predicted-demand", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetPredictedDemand)
		r.Post("/refresh", h.HandleRefreshPredictedDemand)
	})
	s.router.Route("/api/v1/planning/workload-capacity-intelligence", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetWorkloadCapacitySummary)
	})
	s.router.Route("/api/v1/bottlenecks/predicted-bottlenecks", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetOperationalBottleneck)
		r.Post("/refresh", h.HandleRefreshOperationalBottleneck)
	})
	s.router.Route("/api/v1/resources/predicted-allocation", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetResourceAllocation)
		r.Post("/refresh", h.HandleRefreshResourceAllocation)
	})
	s.router.Route("/api/v1/planning/resource-bottleneck-intelligence", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Get("/", h.HandleGetResourceBottleneckSummary)
	})
}

func (s *Server) RegisterAutonomyRoutes(h *autonomy.Handler) {
	if h == nil {
		return
	}
	authGuard := s.newAuthGuard()
	s.router.Route("/api/v1/autonomy", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		h.RegisterRoutes(r)
	})
}

func (s *Server) RegisterWorkforceRoutes(h *workforce.Handler) {
	if h == nil {
		return
	}
	authGuard := s.newAuthGuard()
	s.router.Route("/api/v1/workforce", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		h.RegisterRoutes(r)
	})
}

func (s *Server) RegisterEnterpriseAutonomyRoutes(h *enterprise_autonomy.Handler) {
	if h == nil {
		return
	}
	authGuard := s.newAuthGuard()
	s.router.Route("/api/v1/enterprise", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		h.RegisterRoutes(r)
	})
}

func (s *Server) RegisterIntegrationsRoutes(h *integrations.Handler) {
	if h == nil {
		return
	}
	authGuard := s.newAuthGuard()
	s.router.Route("/api/v1/integrations", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		h.RegisterRoutes(r)
	})
	// Ingress for webhooks (public route with HMAC signature verification)
	h.RegisterWebhookRoutes(s.router)
}

func (s *Server) RegisterSPortalRoutes(h *sportal.Handler) {
	if h == nil {
		return
	}
	s.sportalHandler = h
	authGuard := s.newAuthGuard()

	// Public SPortal Health check & Demo Request submission
	s.router.Get("/api/v1/sportal/health", h.GetHealth)
	s.router.Post("/api/v1/sportal/demo-requests", h.SubmitDemoRequest)
	s.router.Post("/api/v1/demo-requests", h.SubmitDemoRequest)

	// SPortal Authentication Endpoints
	s.router.Route("/api/v1/sportal/auth", func(r chi.Router) {
		r.Post("/login", h.Login)
		r.With(authGuard.RequireAuth, sportal.RequireInternalStaff).Get("/me", h.GetMe)
		r.With(authGuard.RequireAuth, sportal.RequireInternalStaff).Post("/logout", h.Logout)
		r.With(authGuard.RequireAuth, sportal.RequireInternalStaff).Get("/permissions", h.GetPermissions)
	})

	// Protected SPortal Administration endpoints
	s.router.Route("/api/v1/sportal", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		r.Use(sportal.RequireInternalStaff)
		r.Get("/meta", h.GetMeta)
		r.Get("/overview", h.GetOverview)
		r.Get("/organizations/recent", h.ListRecentOrganizations)

		// Organizations & Customer 360 Foundation
		r.With(sportal.RequirePermission(sportal.PermOrganizationsView)).Get("/organizations", h.ListOrganizations)
		r.With(sportal.RequirePermission(sportal.PermOrganizationsView)).Get("/organizations/{id}", h.GetOrganizationDetails)
		r.With(sportal.RequirePermission(sportal.PermOrganizationsCreate)).Post("/organizations", h.CreateOrganization)
		r.With(sportal.RequirePermission(sportal.PermOrganizationsUpdate)).Patch("/organizations/{id}", h.UpdateOrganization)
		r.With(sportal.RequirePermission(sportal.PermOrganizationsUpdate)).Post("/organizations/{id}/logo", h.UploadOrganizationLogo)

		// Protected sensitive financial endpoint requiring billing:sensitive_view
		r.With(sportal.RequirePermission(sportal.PermBillingSensitiveView)).Get("/finance/sensitive", h.GetSensitiveFinancialData)

		// Task S5: Subscriptions & Commercial Plan Administration
		r.With(sportal.RequirePermission(sportal.PermSubscriptionsView)).Get("/subscriptions/plans", h.ListPlans)
		r.With(sportal.RequirePermission(sportal.PermSubscriptionsView)).Get("/subscriptions/plans/{id}", h.GetPlanByID)
		r.With(sportal.RequirePermission(sportal.PermSubscriptionsPlanManage)).Post("/subscriptions/plans", h.CreatePlan)
		r.With(sportal.RequirePermission(sportal.PermSubscriptionsPlanManage)).Patch("/subscriptions/plans/{id}", h.UpdatePlan)

		r.With(sportal.RequirePermission(sportal.PermSubscriptionsView)).Get("/subscriptions", h.ListSubscriptions)
		r.With(sportal.RequirePermission(sportal.PermSubscriptionsView)).Get("/organizations/{id}/subscription", h.GetOrganizationSubscription)
		r.With(sportal.RequirePermission(sportal.PermSubscriptionsCreate)).Post("/organizations/{id}/subscription", h.AssignOrganizationSubscription)
		r.With(sportal.RequirePermission(sportal.PermSubscriptionsUpdate)).Patch("/organizations/{id}/subscription/plan", h.ChangeOrganizationPlan)
		r.With(sportal.RequirePermission(sportal.PermSubscriptionsUpdate)).Patch("/organizations/{id}/subscription/auto-renew", h.ToggleOrganizationAutoRenew)
		r.With(sportal.RequirePermission(sportal.PermSubscriptionsRenew)).Post("/organizations/{id}/subscription/renew", h.RenewOrganizationSubscription)
		r.With(sportal.RequirePermission(sportal.PermSubscriptionsCancel)).Post("/organizations/{id}/subscription/cancel", h.CancelOrganizationSubscription)

		// Task S6: Customer Organization Users, Roles, Invitations & Access Lifecycle
		r.With(sportal.RequirePermission(sportal.PermUsersView)).Get("/users", h.ListCustomerUsers)
		r.With(sportal.RequirePermission(sportal.PermUsersView)).Get("/users/roles", h.ListCustomerRoles)
		r.With(sportal.RequirePermission(sportal.PermUsersView)).Get("/organizations/{id}/users", h.ListCustomerUsers)
		r.With(sportal.RequirePermission(sportal.PermUsersView)).Get("/organizations/{id}/users/{userId}", h.GetCustomerUserDetail)
		r.With(sportal.RequirePermission(sportal.PermUsersView)).Get("/organizations/{id}/users/summary", h.GetOrgUserSummary)
		r.With(sportal.RequirePermission(sportal.PermUsersCreate)).Post("/organizations/{id}/users/invite", h.InviteCustomerUser)
		r.With(sportal.RequirePermission(sportal.PermUsersCreate)).Post("/users/invitations/{invitationId}/resend", h.ResendCustomerInvitation)
		r.With(sportal.RequirePermission(sportal.PermUsersDisable)).Delete("/users/invitations/{invitationId}", h.RevokeCustomerInvitation)
		r.With(sportal.RequirePermission(sportal.PermUsersUpdate)).Patch("/organizations/{id}/users/{userId}/status", h.UpdateCustomerUserStatus)

		// Task S7: Customer Roles, Permission Matrix & Access Administration
		r.With(sportal.RequirePermission(sportal.PermRolesView)).Get("/roles/matrix", h.GetPermissionMatrix)
		r.With(sportal.RequirePermission(sportal.PermUsersUpdate)).Patch("/organizations/{id}/users/{userId}/role", h.UpdateCustomerUserRole)

		// Task S9: Customer 360 Cross-Module Business Intelligence
		r.With(sportal.RequirePermission(sportal.PermOrganizationsView)).Get("/organizations/{id}/shipments", h.GetCustomerShipments)
		r.With(sportal.RequirePermission(sportal.PermOrganizationsView)).Get("/organizations/{id}/invoices", h.GetCustomerInvoices)
		r.With(sportal.RequirePermission(sportal.PermOrganizationsView)).Get("/organizations/{id}/contracts", h.GetCustomerContracts)
		r.With(sportal.RequirePermission(sportal.PermOrganizationsView)).Get("/organizations/{id}/exceptions", h.GetCustomerExceptions)
		r.With(sportal.RequirePermission(sportal.PermIntegrationsView)).Get("/organizations/{id}/integrations", h.GetCustomerIntegrations)
		r.With(sportal.RequirePermission(sportal.PermIntegrationsView)).Get("/organizations/{id}/integrations/webhooks", h.GetCustomerWebhooks)
		r.With(sportal.RequirePermission(sportal.PermIntegrationsView)).Get("/organizations/{id}/integrations/sync-jobs", h.GetCustomerSyncJobs)
		r.With(sportal.RequirePermission(sportal.PermIntegrationsManage)).Post("/organizations/{id}/integrations/toggle", h.ToggleCustomerIntegration)
		r.With(sportal.RequirePermission(sportal.PermIntegrationsManage)).Post("/organizations/{id}/integrations/test", h.TestCustomerIntegrationConnection)
		r.With(sportal.RequirePermission(sportal.PermIntegrationsView)).Get("/integrations", h.GetPlatformIntegrations)
		r.With(sportal.RequirePermission(sportal.PermOrganizationsView)).Get("/organizations/{id}/documents", h.GetCustomerDocuments)
		r.With(sportal.RequirePermission(sportal.PermOrganizationsView)).Get("/organizations/{id}/ai-summary", h.GetCustomerAiSummary)

		// Task S10: Customer Usage, Platform Analytics, Consumption & Adoption
		r.With(sportal.RequirePermission(sportal.PermUsageView)).Get("/organizations/{id}/usage", h.GetCustomerUsageAnalytics)
		r.With(sportal.RequirePermission(sportal.PermUsageView)).Get("/usage", h.GetPlatformUsageAnalytics)

		// Task S11: Customer Health, Customer Success Intelligence & Risk Signals
		r.With(sportal.RequirePermission(sportal.PermCustomerHealthView)).Get("/organizations/{id}/health", h.GetCustomerHealth)
		r.With(sportal.RequirePermission(sportal.PermCustomerHealthView)).Get("/organizations/{id}/health/notes", h.GetCustomerNotes)
		r.With(sportal.RequirePermission(sportal.PermCustomerHealthManage)).Post("/organizations/{id}/health/notes", h.CreateCustomerNote)
		r.With(sportal.RequirePermission(sportal.PermCustomerHealthView)).Get("/customer-health", h.GetPlatformHealth)
		r.With(sportal.RequirePermission(sportal.PermCustomerHealthView)).Get("/health", h.GetPlatformHealth)

		// Task S13: Customer Documents, Compliance, Contracts & Customer Records
		r.With(sportal.RequirePermission(sportal.PermDocumentsView)).Get("/documents/overview", h.GetPlatformDocumentsOverview)
		r.With(sportal.RequirePermission(sportal.PermDocumentsView)).Get("/organizations/{id}/documents/search", h.GetCustomerDocumentsPaginated)
		r.With(sportal.RequirePermission(sportal.PermDocumentsView)).Get("/organizations/{id}/documents/{docId}", h.GetCustomerDocumentDetail)
		r.With(sportal.RequirePermission(sportal.PermDocumentsManage)).Patch("/organizations/{id}/documents/{docId}/status", h.UpdateCustomerDocumentStatus)
		r.With(sportal.RequirePermission(sportal.PermDocumentsView)).Get("/organizations/{id}/documents/{docId}/download", h.DownloadCustomerDocument)
		r.With(sportal.RequirePermission(sportal.PermDocumentsView)).Get("/organizations/{id}/compliance-overview", h.GetCustomerComplianceOverview)
		r.With(sportal.RequirePermission(sportal.PermDocumentsView)).Get("/organizations/{id}/compliance", h.GetCustomerComplianceOverview)
		r.With(sportal.RequirePermission(sportal.PermDocumentsView)).Get("/organizations/{id}/contracts-overview", h.GetCustomerContractsOverview)
		r.With(sportal.RequirePermission(sportal.PermDocumentsView)).Get("/organizations/{id}/contracts", h.GetCustomerContractsOverview)

		// Task S16: SPortal AI, LogisticsHQ Internal Intelligence & Governed AI Operations
		r.Post("/ai/query", h.QueryAi)
		r.Post("/ai/action", h.ExecuteAiAction)
		r.Get("/ai/workforce", h.GetAiWorkforceOverview)
		r.Get("/ai/recommendations", h.ListAiRecommendations)
		r.Get("/organizations/{id}/ai-context", h.GetCustomerAiContext)

		// Task S17: SPortal Settings, Platform Administration & Operational Controls
		r.Route("/settings", func(sr chi.Router) {
			sr.Use(sportal.RequirePermission(sportal.PermSettingsView))
			sr.Get("/", h.GetSettingsOverview)
			sr.Get("/overview", h.GetSettingsOverview)
			sr.Get("/profile", h.GetInternalUserProfile)
			sr.Patch("/profile", h.UpdateInternalUserProfile)
			sr.Get("/platform", h.GetPlatformSettings)
			sr.With(sportal.RequirePermission(sportal.PermSettingsManage)).Patch("/platform/{key}", h.UpdatePlatformSetting)
			sr.Get("/feature-flags", h.GetFeatureFlags)
			sr.With(sportal.RequirePermission(sportal.PermSettingsManage)).Patch("/feature-flags/{key}", h.UpdateFeatureFlag)
			sr.Get("/autonomy", h.GetAutonomyPolicies)
			sr.Post("/autonomy/emergency-halt", h.TriggerEmergencyHalt)
			sr.Get("/integrations", h.GetIntegrationSettings)
			sr.With(sportal.RequirePermission(sportal.PermIntegrationsManage)).Patch("/integrations/{type}/toggle", h.ToggleIntegrationSetting)
			sr.Get("/audit", h.GetRecentAdministrativeAudits)
			sr.Get("/operations", h.GetOperationsHealthSummary)
		})

		// P12: Support Center, Activity Timeline, Notifications & Forensic Audit
		r.With(sportal.RequirePermission(sportal.PermSupportView)).Get("/support/cases", h.GetSupportCases)
		r.With(sportal.RequirePermission(sportal.PermSupportView)).Get("/support/cases/{id}", h.GetSupportCaseDetail)
		r.With(sportal.RequirePermission(sportal.PermSupportManage)).Patch("/support/cases/{id}/status", h.UpdateSupportCaseStatus)
		r.With(sportal.RequirePermission(sportal.PermSupportView)).Post("/support/cases/{id}/notes", h.AddSupportCaseNote)
		r.With(sportal.RequirePermission(sportal.PermSupportManage)).Post("/support/cases", h.CreateSupportCase)

		r.Get("/notifications", h.GetNotificationsList)
		r.Patch("/notifications/{id}/read", h.MarkNotificationRead)
		r.Post("/notifications/mark-all-read", h.MarkAllNotificationsRead)
		r.Patch("/notifications/{id}/acknowledge", h.AcknowledgeNotification)

		r.Get("/activity/timeline", h.GetUnifiedActivityTimeline)
		r.With(sportal.RequirePermission(sportal.PermAuditView)).Get("/audit/search", h.SearchAuditLogs)

		// Demo Requests Management
		r.Get("/demo-requests", h.ListDemoRequests)
		r.Get("/demo-requests/{id}", h.GetDemoRequest)
		r.Patch("/demo-requests/{id}", h.UpdateDemoRequest)
	})
}








