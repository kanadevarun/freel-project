package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/freel/backend/internal/auth"
	"github.com/freel/backend/internal/dashboard"
	"github.com/freel/backend/internal/health"
	"github.com/freel/backend/internal/leads"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/outreach"
	"github.com/freel/backend/internal/rbac"
	"github.com/freel/backend/internal/reports"
	"github.com/freel/backend/internal/rfq"
	"github.com/freel/backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

func (s *Server) setupRoutes() {
	// Health
	s.router.Get("/health", health.HealthCheck)

	// Auth Handlers & Middleware
	authHandler := auth.NewHandler(s.authService)
	authGuard := middleware.NewAuthMiddleware(s.cfg.AWSRegion, s.cfg.CognitoUserPoolID, s.db)
	rbacGuard := middleware.NewRBACMiddleware(s.rbacSvc)

	s.router.Route("/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.Signup)
		r.Post("/verify-email", authHandler.VerifyEmail)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/forgot-password", authHandler.ForgotPassword)
		r.Post("/reset-password", authHandler.ResetPassword)
		r.With(authGuard.RequireAuth).Get("/me", authHandler.GetMe)
	})

	// Public Inbound Email Webhook (Bypasses Cognito AuthGuard)
	s.router.Post("/api/v1/emails/inbound", s.leadsEmailHandler.InboundEmailWebhook)

	s.router.Route("/api/v1", func(r chi.Router) {
		// Enforce JWT Auth for all /api/v1/ routes
		r.Use(authGuard.RequireAuth)

		// Companies & Customer Directory (Protected by JWT and RBAC)
		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/companies", func(w http.ResponseWriter, req *http.Request) {
				userCtx, ok := middleware.GetUserContext(req.Context())
				if !ok || userCtx.OrgID <= 0 {
					utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
					return
				}
				type CompanyRow struct {
					ID           int64   `db:"id" json:"id"`
					OrgID        int64   `db:"org_id" json:"org_id"`
					Name         string  `db:"name" json:"name"`
					Domain       *string `db:"domain" json:"domain,omitempty"`
					Industry     *string `db:"industry" json:"industry,omitempty"`
					ContactName  string  `db:"contact_name" json:"contact_name"`
					ContactEmail string  `db:"contact_email" json:"contact_email"`
					ContactPhone string  `db:"contact_phone" json:"contact_phone"`
					Status       string  `db:"status" json:"status"`
					CreatedAt    string  `db:"created_at" json:"created_at"`
				}
				var rows []CompanyRow
				query := `
					SELECT 
						c.id, c.org_id, c.name, COALESCE(c.domain, '') as domain, COALESCE(c.industry, '') as industry, CAST(c.created_at AS CHAR) as created_at,
						COALESCE(CONCAT(cnt.first_name, ' ', cnt.last_name), '') as contact_name,
						COALESCE(cnt.email, '') as contact_email,
						COALESCE(cnt.phone, '') as contact_phone,
						cust.status as status
					FROM customers cust
					JOIN companies c ON cust.company_id = c.id AND cust.org_id = c.org_id
					LEFT JOIN contacts cnt ON cnt.company_id = c.id AND cnt.org_id = c.org_id
					WHERE cust.org_id = ?
					ORDER BY cust.created_at DESC`
				err := s.db.SelectContext(req.Context(), &rows, query, userCtx.OrgID)
				if err != nil || rows == nil {
					utils.Success(w, http.StatusOK, "Companies retrieved successfully", []CompanyRow{})
					return
				}
				utils.Success(w, http.StatusOK, "Companies retrieved successfully", rows)
			})

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionCreate)).
			Post("/companies", func(w http.ResponseWriter, req *http.Request) {
				userCtx, ok := middleware.GetUserContext(req.Context())
				if !ok || userCtx.OrgID <= 0 {
					utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
					return
				}
				var body struct {
					Name         string `json:"name"`
					Domain       string `json:"domain"`
					Industry     string `json:"industry"`
					ContactName  string `json:"contact_name"`
					ContactEmail string `json:"contact_email"`
					ContactPhone string `json:"contact_phone"`
				}
				if err := json.NewDecoder(req.Body).Decode(&body); err != nil || body.Name == "" {
					utils.Error(w, http.StatusBadRequest, "Company name is required", "INVALID_INPUT")
					return
				}

				tx, err := s.db.BeginTxx(req.Context(), nil)
				if err != nil {
					utils.Error(w, http.StatusInternalServerError, "Failed to start transaction", "SERVER_ERROR")
					return
				}
				defer tx.Rollback()

				// 1. Create company entity record
				res, err := tx.ExecContext(req.Context(), `
					INSERT INTO companies (org_id, name, domain, industry, created_at, updated_at)
					VALUES (?, ?, ?, ?, NOW(), NOW())
				`, userCtx.OrgID, body.Name, body.Domain, body.Industry)
				if err != nil {
					utils.Error(w, http.StatusInternalServerError, "Failed to create company: "+err.Error(), "SERVER_ERROR")
					return
				}
				companyID, err := res.LastInsertId()
				if err != nil {
					utils.Error(w, http.StatusInternalServerError, "Failed to get company ID", "SERVER_ERROR")
					return
				}

				// 2. Create customer relationship record
				_, err = tx.ExecContext(req.Context(), `
					INSERT INTO customers (org_id, company_id, status, created_at, updated_at)
					VALUES (?, ?, 'ACTIVE', NOW(), NOW())
				`, userCtx.OrgID, companyID)
				if err != nil {
					utils.Error(w, http.StatusInternalServerError, "Failed to create customer record: "+err.Error(), "SERVER_ERROR")
					return
				}

				// 3. Create primary contact record if provided
				if body.ContactName != "" || body.ContactEmail != "" {
					parts := strings.SplitN(body.ContactName, " ", 2)
					firstName := parts[0]
					lastName := ""
					if len(parts) > 1 {
						lastName = parts[1]
					}
					_, _ = tx.ExecContext(req.Context(), `
						INSERT INTO contacts (org_id, company_id, first_name, last_name, email, phone, created_at, updated_at)
						VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
					`, userCtx.OrgID, companyID, firstName, lastName, body.ContactEmail, body.ContactPhone)
				}

				if err := tx.Commit(); err != nil {
					utils.Error(w, http.StatusInternalServerError, "Failed to commit customer creation", "SERVER_ERROR")
					return
				}

				utils.Success(w, http.StatusCreated, "Customer created successfully", map[string]interface{}{
					"id":     companyID,
					"name":   body.Name,
					"org_id": userCtx.OrgID,
					"status": "ACTIVE",
				})
			})
		// Leads Endpoints
		r.Route("/leads", func(r chi.Router) {
			leads.AddLeadsHandlers(r, s.leadsEndpoints, authGuard.RequireAuth)
		})

		// Leads Interactions Endpoints (Cognito Protected)
		r.Get("/leads/{id:[0-9]+}/interactions", s.leadsEmailHandler.GetInteractions)
		r.Post("/leads/{id:[0-9]+}/interactions", s.leadsEmailHandler.CreateInteraction)

		// Outreach Endpoints — Campaign management + AI email generation
		r.Route("/outreach", func(r chi.Router) {
			outreach.AddOutreachHandlers(r, s.outreachEndpoints, authGuard.RequireAuth)
		})

		// RFQ Endpoints — Shipment Initiation Platform
		r.Route("/rfqs", func(r chi.Router) {
			rfq.AddRFQHandlers(r, s.rfqEndpoints, authGuard.RequireAuth)
		})

		// Dashboard Endpoints — Mission Control
		r.Route("/dashboard", func(r chi.Router) {
			dashboard.AddDashboardHandlers(r, s.dashboardEndpoints, authGuard.RequireAuth)
		})

		// Notifications Endpoints
		r.Route("/notifications", func(r chi.Router) {
			r.Get("/unread", s.notificationsHandler.GetUnread)
			r.Put("/{id}/read", s.notificationsHandler.MarkAsRead)
		})

		// Reports Endpoints
		r.Route("/reports", func(r chi.Router) {
			reports.AddReportsHandlers(r, s.reportsEndpoints, authGuard.RequireAuth)
		})

		// ── Rate Intelligence Endpoints ───────────────────────────────────────
		// GET  /api/v1/rates/search?origin=INNSA&destination=DEHAM&equipment=40GP
		// GET  /api/v1/rates/{id}
		// POST /api/v1/rates/spot/refresh
		r.Route("/rates", func(r chi.Router) {
			r.Get("/search", s.ratesHandler.SearchRates)
			r.Post("/spot/refresh", s.ratesHandler.RefreshSpotRates)
			r.Get("/{id}", s.ratesHandler.GetRate)
		})

		// ── Contract Intelligence Endpoints ────────────────────────────────────
		// POST /api/v1/contracts/upload (Multipart)
		// GET  /api/v1/contracts/ (List)
		// GET  /api/v1/contracts/{id} (Get)
		// POST /api/v1/contracts/{id}/reprocess
		// GET  /api/v1/contracts/review (Get Review Items list)
		// PUT  /api/v1/contracts/review/{id}/approve (Approve extraction)
		// PUT  /api/v1/contracts/review/{id}/reject (Reject extraction)
		r.Route("/contracts", func(r chi.Router) {
			r.Post("/upload", s.contractsHandler.Upload)
			r.Get("/", s.contractsHandler.List)
			r.Get("/review", s.contractsHandler.ListReview)
			r.Put("/review/{id}/approve", s.contractsHandler.ApproveReview)
			r.Put("/review/{id}/reject", s.contractsHandler.RejectReview)
			r.Get("/{id}", s.contractsHandler.Get)
			r.Post("/{id}/reprocess", s.contractsHandler.Reprocess)
		})

		// ── Shipment Operations Endpoints ─────────────────────────────────────
		r.Route("/shipments", func(r chi.Router) {
			r.Get("/", s.shipmentsHandler.ListShipments)
			r.Get("/{id:[0-9]+}", s.shipmentsHandler.GetShipment)
			r.Post("/{id:[0-9]+}/carrier-update", s.shipmentsHandler.CarrierUpdate)
			r.Post("/exceptions/{id:[0-9]+}/resolve", s.shipmentsHandler.ResolveException)
			
			// Phase 4: Document Compliance routes
			r.Post("/{id:[0-9]+}/documents/upload", s.documentsHandler.UploadDocument)
			r.Get("/{id:[0-9]+}/documents", s.documentsHandler.ListDocuments)
			r.Post("/discrepancies/{id:[0-9]+}/resolve", s.documentsHandler.ResolveDiscrepancy)

			// Phase 5: Finance routes
			r.Post("/{id:[0-9]+}/finance/invoices/upload", s.financeHandler.IngestInvoice)
			r.Get("/{id:[0-9]+}/finance", s.financeHandler.GetFinanceWorkspace)

			// Gap-Closure: Customer Billing & Closure routes
			r.Post("/{id:[0-9]+}/billing/invoices/generate", s.billingHandler.GenerateInvoice)
			r.Get("/{id:[0-9]+}/billing", s.billingHandler.GetBillingWorkspace)
			r.Post("/{id:[0-9]+}/close", s.billingHandler.CloseShipment)
		})
		// Phase 5: Finance discrepancy resolution and manual approval
		r.Post("/finance/discrepancies/{id:[0-9]+}/resolve", s.financeHandler.ResolveDiscrepancy)
		r.Post("/finance/invoices/{id}/approve", s.financeHandler.ApproveInvoice)
		r.Post("/emails/carrier-inbound", s.shipmentsHandler.InboundCarrierEmailWebhook)

		// Customer invoice overrides
		r.Post("/billing/invoices/{id}/approve", s.billingHandler.ApproveInvoice)
		r.Post("/billing/invoices/{id}/pay", s.billingHandler.PayInvoice)

		// Aggregate Document & Invoice routes
		r.Get("/documents", s.documentsHandler.ListAllDocuments)
		r.Get("/invoices", s.financeHandler.ListAllInvoices)

		// Users & Team Management
		r.Route("/users", func(r chi.Router) {
			r.Get("/", s.usersHandler.ListUsers)
			r.Get("/invites", s.usersHandler.ListInvitations)
			r.Post("/invite", s.usersHandler.InviteUser)
			r.Delete("/invites/{id}", s.usersHandler.CancelInvitation)
			r.Patch("/{id}/role", s.usersHandler.UpdateRole)
		})

		// Roles & RBAC Settings
		r.Route("/roles", func(r chi.Router) {
			r.Get("/", s.rbacHandler.GetRoles)
			r.Get("/{id}/permissions", s.rbacHandler.GetRolePermissions)
			r.Put("/{id}/permissions", s.rbacHandler.UpdateRolePermissions)
		})
	})

	// ── Carrier Webhook endpoint (Public signature verified) ───────────────────
	s.router.Post("/webhooks/carriers/{carrier}", s.shipmentsHandler.InboundWebhook)
	s.router.Post("/webhooks/carriers/{carrier}/{integration_id}", s.shipmentsHandler.InboundWebhook)

	// ── Internal Callbacks (Protected by InternalServiceAuthMiddleware) ─────────
	s.router.Route("/internal", func(r chi.Router) {
		r.Use(middleware.InternalServiceAuthMiddleware)

		r.Post("/contracts/callback", s.contractsHandler.Callback)
		r.Get("/ports/normalize", s.ratesHandler.NormalizePort)
		r.Get("/ports/search", s.ratesHandler.SearchPorts)
		r.Get("/pricing/rules", s.pricingHandler.GetRules)
		r.Get("/rfqs/{id:[0-9]+}", s.pricingHandler.GetRFQDetails)
		r.Get("/rates/search", s.pricingHandler.SearchRates)
		r.Post("/pricing/quotes/draft", s.pricingHandler.CreateDraftQuotes)
		r.Post("/pricing/callback", s.pricingHandler.Callback)
		r.Post("/rfqs/from-email", s.leadsEmailHandler.CreateRFQFromEmail)
		r.Post("/sales/callback", s.leadsEmailHandler.SalesCallback)
		r.Get("/shipments/{id:[0-9]+}", s.shipmentsHandler.GetShipmentInternal)
		r.Post("/shipments/{id:[0-9]+}/milestones", s.shipmentsHandler.UpdateMilestoneInternal)
		r.Post("/shipments/{id:[0-9]+}/exceptions", s.shipmentsHandler.CreateExceptionInternal)
		r.Post("/operations/callback", s.shipmentsHandler.CallbackInternal)
		r.Post("/compliance/callback", s.documentsHandler.CallbackInternal)
		r.Get("/shipments/{id:[0-9]+}/documents", s.documentsHandler.ListDocumentsInternal)

		// Phase 5: Finance internal routes
		r.Post("/finance/callback", s.financeHandler.CallbackInternal)
		r.Get("/shipments/{id:[0-9]+}/finance", s.financeHandler.GetFinanceWorkspaceInternal)
	})

	// ── Serve uploads statically (for local dev visual review) ─────────────────
	s.router.Mount("/uploads", http.StripPrefix("/uploads", http.FileServer(http.Dir("./uploads"))))
}
