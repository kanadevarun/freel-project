package server

import (
	"encoding/json"
	"net/http"

	"github.com/freel/backend/internal/auth"
	"github.com/freel/backend/internal/contracts"
	"github.com/freel/backend/internal/customers"
	"github.com/freel/backend/internal/dashboard"
	"github.com/freel/backend/internal/health"
	"github.com/freel/backend/internal/leads"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/outreach"
	"github.com/freel/backend/internal/quotations"
	"github.com/freel/backend/internal/rates"
	"github.com/freel/backend/internal/rbac"
	"github.com/freel/backend/internal/reports"
	"github.com/freel/backend/internal/rfq"
	"github.com/freel/backend/internal/shipments"
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
		r.Get("/invite/validate", authHandler.ValidateInvite)
		r.Post("/invite/accept", authHandler.AcceptInvite)
		r.With(authGuard.RequireAuth).Get("/me", authHandler.GetMe)
		r.With(authGuard.RequireAuth).Post("/logout", authHandler.Logout)
	})

	// Public Inbound Email Webhook (Bypasses Cognito AuthGuard)
	s.router.Post("/api/v1/emails/inbound", s.leadsEmailHandler.InboundEmailWebhook)
	
	// Public Google OAuth Callback
	s.router.Get("/api/v1/organizations/mailboxes/connect/gmail/callback", s.orgHandler.HandleGmailOAuthCallback)
	
	// Public Stripe Webhook
	s.router.Post("/api/v1/subscription/webhook", s.subscriptionHandler.Webhook)
	s.router.Get("/api/v1/subscription/plans_public", s.subscriptionHandler.GetPlans)

	s.router.Route("/api/v1", func(r chi.Router) {
		// Enforce JWT Auth for all /api/v1/ routes
		r.Use(authGuard.RequireAuth)

		// ── Customers Module (Task 21.1) ──────────────────────────────────
		customers.AddCustomerHandlers(r, s.customersEndpoints, authGuard.RequireAuth, rbacGuard)
		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/companies", func(w http.ResponseWriter, req *http.Request) {
				userCtx, ok := middleware.GetUserContext(req.Context())
				if !ok || userCtx.OrgID <= 0 {
					utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
					return
				}
				type CustomerRow struct {
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
				var rows []CustomerRow
				query := `
					SELECT 
						id, org_id, 
						COALESCE(name, '') AS name,
						domain, industry,
						COALESCE(contact_name, '') AS contact_name,
						COALESCE(contact_email, '') AS contact_email,
						COALESCE(contact_phone, '') AS contact_phone,
						COALESCE(status, 'ACTIVE') AS status,
						CAST(created_at AS CHAR) AS created_at
					FROM customers
					WHERE org_id = ?
					ORDER BY created_at DESC`
				err := s.db.SelectContext(req.Context(), &rows, query, userCtx.OrgID)
				if err != nil || rows == nil {
					utils.Success(w, http.StatusOK, "Customers retrieved successfully", []CustomerRow{})
					return
				}
				utils.Success(w, http.StatusOK, "Customers retrieved successfully", rows)
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
					utils.Error(w, http.StatusBadRequest, "Customer name is required", "INVALID_INPUT")
					return
				}

				// Insert directly into customers (no companies table)
				res, err := s.db.ExecContext(req.Context(), `
					INSERT INTO customers 
						(org_id, name, domain, industry, contact_name, contact_email, contact_phone, status, created_at, updated_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, 'ACTIVE', NOW(), NOW())
				`, userCtx.OrgID, body.Name, body.Domain, body.Industry,
					body.ContactName, body.ContactEmail, body.ContactPhone)
				if err != nil {
					utils.Error(w, http.StatusInternalServerError, "Failed to create customer: "+err.Error(), "SERVER_ERROR")
					return
				}
				customerID, err := res.LastInsertId()
				if err != nil {
					utils.Error(w, http.StatusInternalServerError, "Failed to get customer ID", "SERVER_ERROR")
					return
				}

				utils.Success(w, http.StatusCreated, "Customer created successfully", map[string]interface{}{
					"id":     customerID,
					"name":   body.Name,
					"org_id": userCtx.OrgID,
					"status": "ACTIVE",
				})
			})

		// Organization Settings & Profile
		r.Route("/organizations", func(r chi.Router) {
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
				Get("/profile", s.orgHandler.GetProfile)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
				Put("/profile", s.orgHandler.UpdateProfile)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
				Post("/profile/logo", s.orgHandler.UploadLogo)

			// Notification Preferences
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
				Get("/notifications", s.orgHandler.GetNotificationPreferences)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
				Put("/notifications", s.orgHandler.UpdateNotificationPreferences)

			// Email Settings
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
				Get("/email-settings", s.orgHandler.HandleGetEmailSettings)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
				Put("/email-settings", s.orgHandler.HandleUpdateEmailSettings)
			
			// Connected Mailboxes
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
				Get("/mailboxes", s.orgHandler.HandleGetConnectedMailboxes)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
				Get("/mailboxes/{id}", s.orgHandler.HandleGetConnectedMailboxByID)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
				Get("/mailboxes/connect/gmail", s.orgHandler.HandleStartGmailOAuth)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
				Post("/mailboxes", s.orgHandler.HandleConnectMailbox)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
				Put("/mailboxes/{id}", s.orgHandler.HandleUpdateMailbox)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionDelete)).
				Delete("/mailboxes/{id}", s.orgHandler.HandleRemoveMailbox)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
				Post("/mailboxes/{id}/sync", s.orgHandler.HandleSyncMailbox)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
				Post("/mailboxes/{id}/toggle-processing", s.orgHandler.HandleToggleMailboxProcessing)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
				Post("/mailboxes/{id}/disconnect", s.orgHandler.HandleDisconnectMailbox)

			// Carrier Integrations (Settings Sub-route)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
				Get("/carrier-providers", s.carrierHandler.HandleListProviders)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
				Get("/carrier-integrations", s.carrierHandler.HandleListIntegrations)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
				Get("/carrier-integrations/{id}", s.carrierHandler.HandleGetIntegration)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
				Post("/carrier-integrations", s.carrierHandler.HandleConnectCarrier)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
				Put("/carrier-integrations/{id}", s.carrierHandler.HandleUpdateCarrier)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
				Patch("/carrier-integrations/{id}/toggle", s.carrierHandler.HandleToggleCarrier)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionDelete)).
				Delete("/carrier-integrations/{id}", s.carrierHandler.HandleDisconnectCarrier)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
				Post("/carrier-integrations/{id}/test", s.carrierHandler.HandleTestConnection)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
				Post("/carrier-integrations/test-direct", s.carrierHandler.HandleTestDirectConnection)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
				Post("/carrier-integrations/{id}/sync", s.carrierHandler.HandleSyncCarrier)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
				Get("/carrier-integrations/{id}/sync-history", s.carrierHandler.HandleGetSyncHistory)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
				Get("/carrier-integrations/{id}/sync-history/{syncId}", s.carrierHandler.HandleGetSyncJob)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
				Get("/carrier-integrations/{id}/health", s.carrierHandler.HandleGetIntegrationHealth)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
				Post("/carrier-integrations/{id}/tracking", s.carrierHandler.HandleGetTracking)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
				Post("/carrier-integrations/{id}/rates", s.carrierHandler.HandleGetRates)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
				Post("/carrier-integrations/{id}/booking", s.carrierHandler.HandleCreateBooking)

			// Universal Audit Logs (Task 1)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
				Get("/audit-logs", s.auditHandler.ListAuditLogs)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
				Get("/audit-logs/{id}", s.auditHandler.GetAuditLogByID)
		})

		// Settings Route Group (Task 1)
		r.Route("/settings", func(r chi.Router) {
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
				Get("/audit-logs", s.auditHandler.ListAuditLogs)
			r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
				Get("/audit-logs/{id}", s.auditHandler.GetAuditLogByID)
		})

		// Direct /api/v1/audit-logs routes
		r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
			Get("/audit-logs", s.auditHandler.ListAuditLogs)
		r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
			Get("/audit-logs/{id}", s.auditHandler.GetAuditLogByID)

		// Direct /api/v1/carrier-* routes
		r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
			Get("/carrier-providers", s.carrierHandler.HandleListProviders)
		r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
			Get("/carrier-integrations", s.carrierHandler.HandleListIntegrations)
		r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
			Get("/carrier-integrations/{id}", s.carrierHandler.HandleGetIntegration)
		r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
			Post("/carrier-integrations", s.carrierHandler.HandleConnectCarrier)
		r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
			Put("/carrier-integrations/{id}", s.carrierHandler.HandleUpdateCarrier)
		r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
			Patch("/carrier-integrations/{id}/toggle", s.carrierHandler.HandleToggleCarrier)
		r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionDelete)).
			Delete("/carrier-integrations/{id}", s.carrierHandler.HandleDisconnectCarrier)
		r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
			Post("/carrier-integrations/{id}/test", s.carrierHandler.HandleTestConnection)
		r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
			Post("/carrier-integrations/test-direct", s.carrierHandler.HandleTestDirectConnection)
		r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
			Post("/carrier-integrations/{id}/sync", s.carrierHandler.HandleSyncCarrier)
		r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
			Get("/carrier-integrations/{id}/sync-history", s.carrierHandler.HandleGetSyncHistory)
		r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
			Get("/carrier-integrations/{id}/sync-history/{syncId}", s.carrierHandler.HandleGetSyncJob)
		r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
			Get("/carrier-integrations/{id}/health", s.carrierHandler.HandleGetIntegrationHealth)
		r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
			Post("/carrier-integrations/{id}/tracking", s.carrierHandler.HandleGetTracking)
		r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionRead)).
			Post("/carrier-integrations/{id}/rates", s.carrierHandler.HandleGetRates)
		r.With(rbacGuard.RequirePermission(rbac.ResourceSettings, rbac.ActionUpdate)).
			Post("/carrier-integrations/{id}/booking", s.carrierHandler.HandleCreateBooking)

		// Leads Endpoints
		r.Route("/leads", func(r chi.Router) {
			leads.AddLeadsHandlers(r, s.leadsEndpoints, authGuard.RequireAuth)
		})

		// Leads Interactions Endpoints (Cognito Protected)
		r.Get("/leads/{id:[0-9]+}/interactions", s.leadsEmailHandler.GetInteractions)
		r.Post("/leads/{id:[0-9]+}/interactions", s.leadsEmailHandler.CreateInteraction)
		r.Post("/leads/{id:[0-9]+}/interactions/{interaction_id:[0-9]+}/retry-clarification", s.leadsEmailHandler.RetryClarificationEmail)
		r.Post("/leads/{id:[0-9]+}/interactions/{interaction_id:[0-9]+}/retry", s.leadsEmailHandler.RetryEmailInteraction)
		r.Post("/leads/{id:[0-9]+}/interactions/{interaction_id:[0-9]+}/reply", s.leadsEmailHandler.ReplyToInteraction)
		r.Get("/leads/{id:[0-9]+}/interactions/{interaction_id:[0-9]+}/draft", s.leadsEmailHandler.GetDraft)
		r.Put("/leads/{id:[0-9]+}/interactions/{interaction_id:[0-9]+}/draft", s.leadsEmailHandler.SaveDraft)
		r.Delete("/leads/{id:[0-9]+}/interactions/{interaction_id:[0-9]+}/draft", s.leadsEmailHandler.DeleteDraft)

		// Outreach Endpoints — Campaign management + AI email generation
		r.Route("/outreach", func(r chi.Router) {
			outreach.AddOutreachHandlers(r, s.outreachEndpoints, authGuard.RequireAuth)
		})

		// RFQ Endpoints — Shipment Initiation Platform
		r.Route("/rfqs", func(r chi.Router) {
			rfq.AddRFQHandlers(r, s.rfqEndpoints, authGuard.RequireAuth)
		})

		// Dedicated Bookings Workspace Endpoints (Task 15)
		r.Route("/bookings", func(r chi.Router) {
			rfq.AddBookingsWorkspaceHandlers(r, s.rfqEndpoints, authGuard.RequireAuth)
		})

		// Global Search Endpoint
		r.Get("/search", s.searchHandler.HandleGlobalSearch)

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

		// ── Rate Management & Rate Intelligence Endpoints ────────────────────
		// Managed via AddRatesHandlers in internal/rates/transport.go
		r.Route("/rates", func(r chi.Router) {
			rates.AddRatesHandlers(r, s.ratesEndpoints, authGuard.RequireAuth)
		})

		// ── Contract Intelligence Endpoints ────────────────────────────────────
		// POST /api/v1/contracts/upload (Multipart)
		// GET  /api/v1/contracts/ (List)
		// GET  /api/v1/contracts/{id} (Get)
		// POST /api/v1/contracts/{id}/reprocess
		// GET  /api/v1/contracts/review (Get Review Items list)
		// PUT  /api/v1/contracts/review/{id}/approve (Approve extraction)
		// PUT  /api/v1/contracts/review/{id}/reject (Reject extraction)
		// Contract Documents (AI Parsing)
		r.Route("/contract-documents", func(r chi.Router) {
			r.Post("/upload", s.contractsHandler.Upload)
			r.Get("/", s.contractsHandler.List)
			r.Get("/review", s.contractsHandler.ListReview)
			r.Put("/review/{id}/approve", s.contractsHandler.ApproveReview)
			r.Put("/review/{id}/reject", s.contractsHandler.RejectReview)
			r.Get("/{id}", s.contractsHandler.Get)
			r.Post("/{id}/reprocess", s.contractsHandler.Reprocess)
		})

		// Commercial Contracts
		r.Route("/contracts", func(r chi.Router) {
			contracts.AddContractsHandlers(r, s.commercialContractsEndpoints, authGuard.RequireAuth)
		})

		// ── Shipment Operations Endpoints ─────────────────────────────────────
		r.Route("/shipments", func(r chi.Router) {
			shipments.AddShipmentHandlers(r, s.shipmentsEndpoints, s.shipmentsSvc, authGuard.RequireAuth)
			
			// Phase 4: Document Compliance routes
			r.Post("/{id:[0-9]+}/documents/upload", s.documentsHandler.UploadDocument)
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
		shipments.AddShipmentEmailWebhookHandler(r, s.shipmentsEndpoints, s.shipmentsSvc, authGuard.RequireAuth)

		// ── Tracking Analytics & Operational Intelligence (Task 17.8) ────────
		r.Route("/tracking", func(r chi.Router) {
			shipments.AddTrackingAnalyticsHandlers(r, s.shipmentsEndpoints, s.shipmentsSvc, authGuard.RequireAuth)
		})

		// Customer invoice overrides
		r.Post("/billing/invoices/{id}/approve", s.billingHandler.ApproveInvoice)
		r.Post("/billing/invoices/{id}/pay", s.billingHandler.PayInvoice)

		// Aggregate Document & Invoice routes
		r.Get("/documents", s.documentsHandler.ListAllDocuments)
		r.Post("/documents/upload", s.documentsHandler.UploadGeneralDocument)
		r.Delete("/documents/{id}", s.documentsHandler.DeleteDocument)
		// Invoices Module Endpoints
		r.Route("/invoices", func(r chi.Router) {
			r.Get("/", s.invoicesHandler.ListInvoices)
			r.Get("/kpi-stats", s.invoicesHandler.GetKPIStats)
			r.Get("/payments", s.invoicesHandler.ListAllPayments)
			r.Post("/", s.invoicesHandler.CreateInvoice)
			r.Get("/{id:[0-9]+}", s.invoicesHandler.GetInvoiceByID)
			r.Put("/{id:[0-9]+}", s.invoicesHandler.UpdateDraftInvoice)
			r.Post("/{id:[0-9]+}/issue", s.invoicesHandler.IssueInvoice)
			r.Post("/{id:[0-9]+}/submit-approval", s.invoicesHandler.SubmitForApproval)
			r.Put("/{id:[0-9]+}/status", s.invoicesHandler.UpdateInvoiceStatus)
			r.Post("/{id:[0-9]+}/bookmark", s.invoicesHandler.ToggleBookmark)
			r.Post("/{id:[0-9]+}/cancel", s.invoicesHandler.CancelInvoice)
			r.Post("/{id:[0-9]+}/payments", s.invoicesHandler.RecordPayment)
			r.Get("/{id:[0-9]+}/payments", s.invoicesHandler.GetInvoicePayments)
			r.Post("/{id:[0-9]+}/documents", s.invoicesHandler.UploadDocument)
		})

		// Approvals Endpoints
		r.Route("/approvals", func(r chi.Router) {
			r.Get("/", s.approvalsHandler.ListApprovals)
			r.Get("/stats", s.approvalsHandler.GetApprovalStats)
			r.Post("/", s.approvalsHandler.CreateApproval)
			r.Get("/{id}", s.approvalsHandler.GetApprovalByID)
			r.Post("/{id}/approve", s.approvalsHandler.ApproveRequest)
			r.Post("/{id}/reject", s.approvalsHandler.RejectRequest)
		})

		// SaaS Subscription & Billing Endpoints
		r.Route("/subscription", func(r chi.Router) {
			r.Get("/", s.subscriptionHandler.GetWorkspace)
			r.Get("/plans", s.subscriptionHandler.GetPlans)
			r.Post("/plan/preview", s.subscriptionHandler.PreviewPlanChange)
			r.Post("/change", s.subscriptionHandler.ChangePlan)
			r.Post("/checkout", s.subscriptionHandler.Checkout)
			r.Post("/cancel", s.subscriptionHandler.CancelSubscription)
			r.Post("/reactivate", s.subscriptionHandler.ReactivateSubscription)
			r.Post("/portal", s.subscriptionHandler.CreateCustomerPortal)
			r.Get("/addons/config", s.subscriptionHandler.GetAddonConfigs)
			r.Post("/addons", s.subscriptionHandler.UpdateAddons)
		})

		// Users & Team Management
		r.Route("/users", func(r chi.Router) {
			r.Get("/", s.usersHandler.ListUsers)
			r.Get("/invites", s.usersHandler.ListInvitations)
			r.Post("/invite", s.usersHandler.InviteUser)
			r.Delete("/invites/{id}", s.usersHandler.CancelInvitation)
			r.Patch("/{id}/role", s.usersHandler.UpdateRole)
			r.Delete("/{id}", s.usersHandler.RemoveUser)
		})

		// Roles & RBAC Settings
		r.Route("/roles", func(r chi.Router) {
			r.Get("/stats", s.rbacHandler.GetStats)
			r.Get("/", s.rbacHandler.GetRoles)
			r.Post("/", s.rbacHandler.CreateRole)
			r.Get("/{id}/permissions", s.rbacHandler.GetRolePermissions)
			r.Put("/{id}/permissions", s.rbacHandler.UpdateRolePermissions)
			r.Put("/{id}", s.rbacHandler.UpdateRole)
			r.Delete("/{id}", s.rbacHandler.DeleteRole)
		})
	})

	// ── Quotations Module (Task 18) ────────────────────────────────────────────
	s.router.Route("/api/v1/quotations", func(r chi.Router) {
		r.Use(authGuard.RequireAuth)
		quotations.AddQuotationHandlers(r, s.quotationsEndpoints, func(h http.Handler) http.Handler { return h })
	})
	quotations.AddPublicQuotationHandlers(s.router, s.quotationsEndpoints)

	// ── Carrier Webhook endpoint (Public signature verified) ───────────────────
	shipments.AddCarrierWebhookHandlers(s.router, s.shipmentsEndpoints, s.shipmentsSvc)
	s.router.Post("/api/v1/carrier-integrations/webhooks/{providerCode}", s.carrierHandler.HandleInboundWebhook)

	// ── Internal Callbacks (Protected by InternalServiceAuthMiddleware) ─────────
	s.router.Route("/internal", func(r chi.Router) {
		r.Use(middleware.InternalServiceAuthMiddleware)

		r.Post("/contracts/callback", s.contractsHandler.Callback)
		rates.AddPortInternalHandlers(r)
		r.Get("/pricing/rules", s.pricingHandler.GetRules)
		r.Get("/rfqs/{id:[0-9]+}", s.pricingHandler.GetRFQDetails)
		r.Get("/rates/search", s.pricingHandler.SearchRates)
		r.Post("/pricing/quotes/draft", s.pricingHandler.CreateDraftQuotes)
		r.Post("/pricing/callback", s.pricingHandler.Callback)
		r.Post("/rfqs/from-email", s.leadsEmailHandler.CreateRFQFromEmail)
		r.Post("/sales/callback", s.leadsEmailHandler.SalesCallback)
		shipments.AddShipmentInternalHandlers(r, s.shipmentsEndpoints)
		r.Post("/compliance/callback", s.documentsHandler.CallbackInternal)
		r.Get("/shipments/{id:[0-9]+}/documents", s.documentsHandler.ListDocumentsInternal)

		// Phase 5: Finance internal routes
		r.Post("/finance/callback", s.financeHandler.CallbackInternal)
		r.Get("/shipments/{id:[0-9]+}/finance", s.financeHandler.GetFinanceWorkspaceInternal)
	})

	// ── Serve uploads statically (for local dev visual review) ─────────────────
	s.router.Mount("/uploads", http.StripPrefix("/uploads", http.FileServer(http.Dir("./uploads"))))
}
