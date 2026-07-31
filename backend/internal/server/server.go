package server

import (
	"log"
	"net/http"

	"github.com/freel/backend/internal/auth"
	"github.com/freel/backend/internal/config"
	"github.com/freel/backend/internal/dashboard"
	"github.com/freel/backend/internal/leads"
	"github.com/freel/backend/internal/notifications"
	"github.com/freel/backend/internal/outreach"
	"github.com/freel/backend/internal/reports"
	"github.com/freel/backend/internal/rfq"
	"github.com/freel/backend/internal/rbac"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	cfg             *config.Config
	router          *chi.Mux
	authService      *auth.Service
	rbacSvc          rbac.Service
	leadsEndpoints   leads.Endpoints
	outreachEndpoints outreach.Endpoints // Handles all /api/v1/outreach/* routes
	rfqEndpoints    rfq.Endpoints
	dashboardEndpoints dashboard.Endpoints
	notificationsHandler *notifications.Handler
	reportsEndpoints  reports.Endpoints
}

func NewServer(cfg *config.Config, authService *auth.Service, rbacSvc rbac.Service, leadsEndpoints leads.Endpoints, outreachEndpoints outreach.Endpoints, rfqEndpoints rfq.Endpoints, dashboardEndpoints dashboard.Endpoints, notifHandler *notifications.Handler, reportsEndpoints reports.Endpoints) *Server {
	s := &Server{
		cfg:             cfg,
		router:          chi.NewRouter(),
		authService:      authService,
		rbacSvc:          rbacSvc,
		leadsEndpoints:  leadsEndpoints,
		outreachEndpoints: outreachEndpoints,
		rfqEndpoints:    rfqEndpoints,
		dashboardEndpoints: dashboardEndpoints,
		notificationsHandler: notifHandler,
		reportsEndpoints:  reportsEndpoints,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) Start() error {
	log.Printf("Server starting on port %s", s.cfg.Port)
	return http.ListenAndServe(":"+s.cfg.Port, s.router)
}
