package server

import (
	"log"
	"net/http"

	"github.com/freel/backend/internal/approvals"
	auditTransport "github.com/freel/backend/internal/audit/transport"
	"github.com/freel/backend/internal/auth"
	"github.com/freel/backend/internal/billing"
	carrierTransport "github.com/freel/backend/internal/carrier/transport"
	"github.com/freel/backend/internal/config"
	"github.com/freel/backend/internal/contracts"
	"github.com/freel/backend/internal/customers"
	"github.com/freel/backend/internal/dashboard"
	"github.com/freel/backend/internal/documents"
	"github.com/freel/backend/internal/finance"
	"github.com/freel/backend/internal/invoices"
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
	"github.com/freel/backend/internal/shipments"
	"github.com/freel/backend/internal/subscription"
	"github.com/freel/backend/internal/users"
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
}

func NewServer(cfg *config.Config, db *sqlx.DB, authService *auth.Service, rbacSvc rbac.Service, rbacHandler *rbac.Handler, usersHandler *users.Handler, orgHandler *organization.Handler, leadsEndpoints leads.Endpoints, leadsEmailHandler *leads.EmailHandler, outreachEndpoints outreach.Endpoints, rfqEndpoints rfq.Endpoints, dashboardEndpoints dashboard.Endpoints, notifHandler *notifications.Handler, reportsEndpoints reports.Endpoints, ratesEndpoints rates.Endpoints, contractsHandler *contracts.Handler, pricingHandler *pricing.Handler, shipmentsEndpoints shipments.Endpoints, shipmentsSvc shipments.Service, documentsHandler *documents.Handler, financeHandler *finance.Handler, billingHandler *billing.Handler, subscriptionHandler *subscription.Handler, quotationsEndpoints quotations.Endpoints, commercialContractsEndpoints contracts.Endpoints, customersEndpoints customers.Endpoints, approvalsHandler *approvals.Handler, invoicesHandler *invoices.Handler, carrierHandler *carrierTransport.CarrierHandler, searchHandler *search.Handler, auditHandler *auditTransport.Handler) *Server {
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
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) Start() error {
	log.Printf("Server starting on port %s", s.cfg.Port)
	return http.ListenAndServe(":"+s.cfg.Port, s.router)
}
