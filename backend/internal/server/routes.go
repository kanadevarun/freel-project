package server

import (
	"net/http"

	"github.com/freel/backend/internal/auth"
	"github.com/freel/backend/internal/dashboard"
	"github.com/freel/backend/internal/health"
	"github.com/freel/backend/internal/leads"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/outreach"
	"github.com/freel/backend/internal/rbac"
	"github.com/freel/backend/internal/reports"
	"github.com/freel/backend/internal/rfq"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/mux"
)

func (s *Server) setupRoutes() {
	// Health
	s.router.Get("/health", health.HealthCheck)

	// Auth Handlers
	authHandler := auth.NewHandler(s.authService)

	s.router.Route("/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.Signup)
		r.Post("/verify-email", authHandler.VerifyEmail)
		r.Post("/login", authHandler.Login)
		r.Post("/forgot-password", authHandler.ForgotPassword)
		r.Post("/reset-password", authHandler.ResetPassword)
		r.Get("/me", authHandler.GetMe)
	})

	// Protected API Routes
	authGuard := middleware.NewAuthMiddleware(s.cfg.AWSRegion, s.cfg.CognitoUserPoolID)
	rbacGuard := middleware.NewRBACMiddleware(s.rbacSvc)

	s.router.Route("/api/v1", func(r chi.Router) {
		// Enforce JWT Auth for all /api/v1/ routes
		r.Use(authGuard.RequireAuth)

		// Example Protected Route for testing RBAC
		// Simple meaning: Only people with "Read Companies" permission can access this route.
		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/companies", func(w http.ResponseWriter, req *http.Request) {
				w.Write([]byte("Companies data (Protected by JWT and RBAC)"))
			})
		// Leads Endpoints
		leadsMux := mux.NewRouter()
		leads.AddLeadsHandlers(leadsMux, s.leadsEndpoints, authGuard.RequireAuth)
		r.Mount("/leads", leadsMux)

		// Outreach Endpoints — Campaign management + AI email generation
		// Outreach Endpoints — Campaign management + AI email generation
		outreachMux := mux.NewRouter()
		outreach.AddOutreachHandlers(outreachMux, s.outreachEndpoints, authGuard.RequireAuth)
		r.Mount("/outreach", outreachMux)

		// RFQ Endpoints — Shipment Initiation Platform
		rfqMux := mux.NewRouter()
		rfq.AddRFQHandlers(rfqMux, s.rfqEndpoints, authGuard.RequireAuth)
		r.Mount("/rfqs", rfqMux)

		// Dashboard Endpoints — Mission Control
		dashboardMux := mux.NewRouter()
		dashboard.AddDashboardHandlers(dashboardMux, s.dashboardEndpoints, authGuard.RequireAuth)
		r.Mount("/dashboard", dashboardMux)

		// Notifications Endpoints
		r.Route("/notifications", func(r chi.Router) {
			r.Get("/unread", s.notificationsHandler.GetUnread)
			r.Put("/{id}/read", s.notificationsHandler.MarkAsRead)
		})

		// Reports Endpoints
		reportsMux := mux.NewRouter()
		reports.AddReportsHandlers(reportsMux, s.reportsEndpoints, authGuard.RequireAuth)
		r.Mount("/reports", reportsMux)
	})
}
