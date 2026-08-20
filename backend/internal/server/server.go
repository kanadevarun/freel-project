package server

import (
	"log"
	"net/http"

	"github.com/freel/backend/internal/auth"
	"github.com/freel/backend/internal/config"
	"github.com/freel/backend/internal/contracts"
	"github.com/freel/backend/internal/dashboard"
	"github.com/freel/backend/internal/leads"
	"github.com/freel/backend/internal/notifications"
	"github.com/freel/backend/internal/outreach"
	"github.com/freel/backend/internal/rates"
	"github.com/freel/backend/internal/pricing"
	"github.com/freel/backend/internal/reports"
	"github.com/freel/backend/internal/rfq"
	"github.com/freel/backend/internal/rbac"
	"github.com/freel/backend/internal/documents"
	"github.com/freel/backend/internal/finance"
	"github.com/freel/backend/internal/shipments"
	"github.com/freel/backend/internal/billing"
	"github.com/freel/backend/internal/users"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	cfg                  *config.Config
	db                   *sqlx.DB
	router               *chi.Mux
	authService          *auth.Service
	rbacSvc              rbac.Service
	rbacHandler          *rbac.Handler
	usersHandler         *users.Handler
	leadsEndpoints       leads.Endpoints
	leadsEmailHandler    *leads.EmailHandler
	outreachEndpoints    outreach.Endpoints
	rfqEndpoints         rfq.Endpoints
	dashboardEndpoints   dashboard.Endpoints
	notificationsHandler *notifications.Handler
	reportsEndpoints     reports.Endpoints
	ratesHandler         *rates.Handler
	contractsHandler     *contracts.Handler
	pricingHandler       *pricing.Handler
	shipmentsHandler     *shipments.Handler
	documentsHandler     *documents.Handler
	financeHandler       *finance.Handler
	billingHandler       *billing.Handler
}

func NewServer(cfg *config.Config, db *sqlx.DB, authService *auth.Service, rbacSvc rbac.Service, rbacHandler *rbac.Handler, usersHandler *users.Handler, leadsEndpoints leads.Endpoints, leadsEmailHandler *leads.EmailHandler, outreachEndpoints outreach.Endpoints, rfqEndpoints rfq.Endpoints, dashboardEndpoints dashboard.Endpoints, notifHandler *notifications.Handler, reportsEndpoints reports.Endpoints, ratesHandler *rates.Handler, contractsHandler *contracts.Handler, pricingHandler *pricing.Handler, shipmentsHandler *shipments.Handler, documentsHandler *documents.Handler, financeHandler *finance.Handler, billingHandler *billing.Handler) *Server {
	s := &Server{
		cfg:                  cfg,
		db:                   db,
		router:               chi.NewRouter(),
		authService:          authService,
		rbacSvc:              rbacSvc,
		rbacHandler:          rbacHandler,
		usersHandler:         usersHandler,
		leadsEndpoints:       leadsEndpoints,
		leadsEmailHandler:    leadsEmailHandler,
		outreachEndpoints:    outreachEndpoints,
		rfqEndpoints:         rfqEndpoints,
		dashboardEndpoints:   dashboardEndpoints,
		notificationsHandler: notifHandler,
		reportsEndpoints:     reportsEndpoints,
		ratesHandler:         ratesHandler,
		contractsHandler:     contractsHandler,
		pricingHandler:       pricingHandler,
		shipmentsHandler:     shipmentsHandler,
		documentsHandler:     documentsHandler,
		financeHandler:       financeHandler,
		billingHandler:       billingHandler,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) Start() error {
	log.Printf("Server starting on port %s", s.cfg.Port)
	return http.ListenAndServe(":"+s.cfg.Port, s.router)
}
