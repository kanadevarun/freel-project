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
	authGuard := s.newAuthGuard()
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
		r.Post("/leads/{id:[0-9]+}/interactions/{interaction_id:[0-9]+}/approve-draft", s.leadsEmailHandler.ApproveDraft)
		r.Post("/leads/{id:[0-9]+}/interactions/{interaction_id:[0-9]+}/reject-draft", s.leadsEmailHandler.RejectDraft)

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

		// Notifications & Escalation Center Endpoints
		r.Route("/notifications", func(r chi.Router) {
			r.Get("/", s.notificationsHandler.GetNotifications)
			r.Get("/unread", s.notificationsHandler.GetUnread)
			r.Get("/unread-count", s.notificationsHandler.GetUnreadCount)
			r.Get("/stats", s.notificationsHandler.GetStats)
			r.Get("/escalations", s.notificationsHandler.GetEscalations)
			r.Get("/preferences", s.notificationsHandler.GetPreferences)
			r.Put("/preferences", s.notificationsHandler.UpdatePreferences)
			r.Post("/evaluate", s.notificationsHandler.Evaluate)
			r.Post("/read-all", s.notificationsHandler.MarkAllAsRead)
			r.Get("/{id}", s.notificationsHandler.GetNotificationByID)
			r.Post("/{id}/read", s.notificationsHandler.MarkAsRead)
			r.Put("/{id}/read", s.notificationsHandler.MarkAsRead)
			r.Post("/{id}/unread", s.notificationsHandler.MarkAsUnread)
			r.Post("/{id}/dismiss", s.notificationsHandler.Dismiss)
			r.Post("/{id:[0-9]+}/acknowledge", s.notificationsHandler.Acknowledge)
			r.Post("/{id:[0-9]+}/snooze", s.notificationsHandler.Snooze)
			r.Post("/{id:[0-9]+}/escalate", s.notificationsHandler.Escalate)
			r.Post("/{id:[0-9]+}/analyze-ai", s.notificationsHandler.AnalyzeAI)
			r.Post("/{id:[0-9]+}/generate-draft", s.notificationsHandler.GenerateDraft)
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
		r.Post("/documents", s.documentsHandler.UploadGeneralDocument)
		r.Post("/documents/upload", s.documentsHandler.UploadGeneralDocument)
		r.Get("/documents/{id}", s.documentsHandler.GetDocument)
		r.Get("/documents/{id}/download", s.documentsHandler.DownloadDocument)
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

		// Debit Notes Module Endpoints
		r.Route("/debit-notes", func(r chi.Router) {
			r.Get("/", s.invoicesHandler.ListDebitNotes)
			r.Get("/kpi-stats", s.invoicesHandler.GetDebitNoteKPIStats)
			r.Post("/", s.invoicesHandler.CreateDebitNote)
			r.Get("/{id:[0-9]+}", s.invoicesHandler.GetDebitNote)
			r.Post("/{id:[0-9]+}/issue", s.invoicesHandler.IssueDebitNote)
			r.Post("/{id:[0-9]+}/void", s.invoicesHandler.VoidDebitNote)
		})



		// Approvals Endpoints
		r.Route("/approvals", func(r chi.Router) {
			r.Get("/", s.approvalsHandler.ListApprovals)
			r.Get("/stats", s.approvalsHandler.GetApprovalStats)
			r.Get("/requirements", s.approvalsHandler.GetApprovalRequirements)
			r.Post("/", s.approvalsHandler.CreateApproval)
			r.Get("/{id}", s.approvalsHandler.GetApprovalByID)
			r.Get("/{id}/preview", s.approvalsHandler.GetActionPreview)
			r.Get("/{id}/history", s.approvalsHandler.GetDecisionHistory)
			r.Get("/{id}/execution-status", s.approvalsHandler.GetExecutionStatus)
			r.Post("/{id}/retry-execution", s.approvalsHandler.RetryExecution)
			r.Get("/{id}/recommendation", s.approvalsHandler.GetRelatedRecommendation)
			r.Get("/{id}/source-record", s.approvalsHandler.GetRelatedSourceRecord)
			r.Get("/{id}/audit", s.approvalsHandler.GetAuditHistory)
			r.Post("/{id}/approve", s.approvalsHandler.ApproveRequest)
			r.Post("/{id}/reject", s.approvalsHandler.RejectRequest)
			r.Post("/{id}/return", s.approvalsHandler.ReturnRequest)
			r.Post("/{id}/cancel", s.approvalsHandler.CancelRequest)
		})

		// AI Tasks & Workforce Monitoring Endpoints
		if s.aiTasksHandler != nil {
			r.Route("/ai/tasks", func(r chi.Router) {
				r.Get("/", s.aiTasksHandler.ListTasks)
				r.Get("/stats", s.aiTasksHandler.GetTaskStats)
				r.Get("/{id:[0-9]+}", s.aiTasksHandler.GetTaskByID)
				r.Post("/{id:[0-9]+}/cancel", s.aiTasksHandler.CancelTask)
				r.Post("/{id:[0-9]+}/retry", s.aiTasksHandler.RetryTask)
			})
			r.Route("/ai/workforce", func(r chi.Router) {
				r.Get("/summary", s.aiTasksHandler.GetWorkforceSummary)
				r.Get("/tasks", s.aiTasksHandler.ListWorkforceTasks)
				r.Get("/health", s.aiTasksHandler.GetWorkforceHealth)
			})
		}

		// AI Action & Recommendation Center & Customer Follow-Up Assistant (Phase 2 Tasks 2.1 & 2.2)
		if s.recommendationsHandler != nil {
			r.Route("/recommendations", func(r chi.Router) {
				r.Get("/", s.recommendationsHandler.List)
				r.Get("/stats", s.recommendationsHandler.GetStats)
				r.Post("/generate", s.recommendationsHandler.Generate)
				r.Get("/{id:[0-9]+}", s.recommendationsHandler.GetByID)
				r.Patch("/{id:[0-9]+}/status", s.recommendationsHandler.UpdateStatus)
				r.Post("/{id:[0-9]+}/assign", s.recommendationsHandler.Assign)
				r.Post("/{id:[0-9]+}/dismiss", s.recommendationsHandler.Dismiss)
				r.Post("/{id:[0-9]+}/review", s.recommendationsHandler.MarkReviewed)
				r.Get("/{id:[0-9]+}/evidence", s.recommendationsHandler.GetEvidence)
				r.Get("/source/{sourceType}/{sourceId:[0-9]+}", s.recommendationsHandler.ListBySource)
				// Task 2.2 Customer Follow-Up actions on recommendations
				r.Post("/{id:[0-9]+}/draft", s.recommendationsHandler.GenerateDraft)
				r.Patch("/{id:[0-9]+}/draft", s.recommendationsHandler.SaveDraft)
				r.Post("/{id:[0-9]+}/task", s.recommendationsHandler.CreateFollowupTask)
				r.Get("/tasks", s.recommendationsHandler.ListFollowupTasks)
				r.Get("/followups/stats", s.recommendationsHandler.GetFollowupStats)
				// Task 2.3 RFQ & Quotation Assistant actions on recommendations
				r.Get("/{id:[0-9]+}/action-preview", s.recommendationsHandler.GetActionPreview)
				r.Post("/{id:[0-9]+}/request-approval", s.recommendationsHandler.RequestApproval)
				// Task 2.4 Shipment Exception & Operations Copilot evidence routes
				r.Get("/shipments/{shipmentId:[0-9]+}/evidence", s.recommendationsHandler.GetShipmentEvidence)
				r.Get("/milestones/{milestoneId:[0-9]+}/evidence", s.recommendationsHandler.GetMilestoneEvidence)
				r.Get("/exceptions/{exceptionId:[0-9]+}/evidence", s.recommendationsHandler.GetExceptionEvidence)
				// Task 2.5 Invoice & Collections Assistant routes
				r.Get("/invoices/{invoiceId:[0-9]+}/evidence", s.recommendationsHandler.GetInvoiceEvidence)
				r.Get("/customers/{customerId:[0-9]+}/collection-summary", s.recommendationsHandler.GetCustomerCollectionSummary)
				// Task 2.6 Contract, Document & Compliance Assistant routes
				r.Get("/contracts/{contractId:[0-9]+}/evidence", s.recommendationsHandler.GetContractEvidence)
				r.Get("/documents/{documentId}/evidence", s.recommendationsHandler.GetDocumentEvidence)
				r.Get("/compliance/{complianceId:[0-9]+}/evidence", s.recommendationsHandler.GetComplianceEvidence)
				r.Get("/contracts/compliance-summary", s.recommendationsHandler.GetContractComplianceSummary)
			})

			r.Route("/followups", func(r chi.Router) {
				r.Get("/", s.recommendationsHandler.List)
				r.Get("/stats", s.recommendationsHandler.GetFollowupStats)
				r.Get("/tasks", s.recommendationsHandler.ListFollowupTasks)
				r.Get("/{id:[0-9]+}", s.recommendationsHandler.GetByID)
				r.Post("/{id:[0-9]+}/draft", s.recommendationsHandler.GenerateDraft)
				r.Patch("/{id:[0-9]+}/draft", s.recommendationsHandler.SaveDraft)
				r.Post("/{id:[0-9]+}/task", s.recommendationsHandler.CreateFollowupTask)
				r.Post("/{id:[0-9]+}/assign", s.recommendationsHandler.Assign)
				r.Post("/{id:[0-9]+}/dismiss", s.recommendationsHandler.Dismiss)
				r.Post("/{id:[0-9]+}/review", s.recommendationsHandler.MarkReviewed)
			})
		}

		// Workflow Automation & Scheduled AI Jobs (Phase 2 Task 2.7)
		if s.automationsHandler != nil {
			r.Route("/automations", func(r chi.Router) {
				r.Get("/", s.automationsHandler.ListAutomations)
				r.Post("/", s.automationsHandler.CreateAutomation)
				r.Get("/supported-types", s.automationsHandler.GetSupportedTypes)
				r.Get("/stats", s.automationsHandler.GetStats)
				r.Get("/executions", s.automationsHandler.ListExecutions)
				r.Get("/executions/{executionId:[0-9]+}", s.automationsHandler.GetExecution)
				r.Post("/executions/{executionId:[0-9]+}/cancel", s.automationsHandler.CancelExecution)
				r.Post("/executions/{executionId:[0-9]+}/retry", s.automationsHandler.RetryExecution)
				r.Get("/executions/{executionId:[0-9]+}/recommendations", s.automationsHandler.GetExecutionRecommendations)
				r.Get("/executions/{executionId:[0-9]+}/insights", s.automationsHandler.GetExecutionInsights)

				// Operational Intelligence & Insights Feed (Phase 3 Task 3.1)
				r.Get("/insights", s.automationsHandler.ListInsights)
				r.Get("/insights/{id:[0-9]+}", s.automationsHandler.GetInsight)
				r.Post("/insights/{id:[0-9]+}/acknowledge", s.automationsHandler.AcknowledgeInsight)
				r.Post("/insights/{id:[0-9]+}/dismiss", s.automationsHandler.DismissInsight)
				r.Post("/evaluate", s.automationsHandler.EvaluateEvent)

				r.Get("/{id:[0-9]+}", s.automationsHandler.GetAutomation)
				r.Put("/{id:[0-9]+}", s.automationsHandler.UpdateAutomation)
				r.Delete("/{id:[0-9]+}", s.automationsHandler.DeleteAutomation)
				r.Post("/{id:[0-9]+}/enable", s.automationsHandler.EnableAutomation)
				r.Post("/{id:[0-9]+}/disable", s.automationsHandler.DisableAutomation)
				r.Get("/{id:[0-9]+}/preview-next", s.automationsHandler.PreviewNextRun)
				r.Post("/{id:[0-9]+}/preview-next", s.automationsHandler.PreviewNextRun)
				r.Post("/{id:[0-9]+}/run", s.automationsHandler.TriggerManualRun)
				r.Get("/{id:[0-9]+}/executions", s.automationsHandler.ListExecutions)
			})
		}

		// AI Memory & Personalization (Phase 2 Task 2.10)
		if s.memoryHandler != nil {
			r.Route("/memory", func(r chi.Router) {
				r.Get("/", s.memoryHandler.ListMemories)
				r.Post("/", s.memoryHandler.CreateMemory)
				r.Post("/propose", s.memoryHandler.ProposeMemory)
				r.Get("/stats", s.memoryHandler.GetStats)
				r.Get("/settings", s.memoryHandler.GetUserSettings)
				r.Put("/settings", s.memoryHandler.UpdateUserSettings)
				r.Post("/toggle", s.memoryHandler.TogglePersonalization)
				r.Post("/clear-personal", s.memoryHandler.ClearPersonalMemories)
				r.Get("/preferences", s.memoryHandler.ListPreferences)
				r.Put("/preferences", s.memoryHandler.SetPreference)
				r.Delete("/preferences/{key}", s.memoryHandler.DeletePreference)
				r.Post("/runtime-context", s.memoryHandler.GetRuntimeContext)
				r.Get("/audit", s.memoryHandler.ListAuditEvents)

				r.Get("/{id:[0-9]+}", s.memoryHandler.GetMemory)
				r.Put("/{id:[0-9]+}", s.memoryHandler.UpdateMemory)
				r.Delete("/{id:[0-9]+}", s.memoryHandler.DeleteMemory)
				r.Post("/{id:[0-9]+}/disable", s.memoryHandler.DisableMemory)
				r.Post("/{id:[0-9]+}/enable", s.memoryHandler.EnableMemory)
			})
		}

		// AI Performance, Cost, and Quality Monitoring (Phase 2 Task 2.11)
		if s.monitoringHandler != nil {
			r.Route("/monitoring", func(r chi.Router) {
				s.monitoringHandler.RegisterRoutes(r)
			})
		}

		// Unified Business Context & Read-Only Intelligence Endpoints (Task 1.1, 1.2, 1.3, 1.4, 1.5)
		if s.contextHandler != nil {
			r.Get("/customers/{id:[0-9]+}/intelligence", s.contextHandler.GetCustomerIntelligence)
			r.Get("/rfqs/{id:[0-9]+}/intelligence", s.contextHandler.GetRFQIntelligence)
			r.Get("/shipments/{id:[0-9]+}/intelligence", s.contextHandler.GetShipmentIntelligence)
			r.Get("/shipments/operations-summary", s.contextHandler.GetOrgOperationsSummary)
			r.Get("/invoices/{id:[0-9]+}/intelligence", s.contextHandler.GetInvoiceIntelligence)
			r.Get("/invoices/finance-summary", s.contextHandler.GetOrgFinanceSummary)
			r.Get("/contracts/{id:[0-9]+}/intelligence", s.contextHandler.GetContractIntelligence)
			r.Get("/contracts/compliance-summary", s.contextHandler.GetOrgContractComplianceSummary)
			r.Get("/contracts/coverage-check", s.contextHandler.GetContractCoverage)
			r.Get("/insights/cross-module", s.contextHandler.GetCrossModuleInsights)
			r.Get("/insights/summary", s.contextHandler.GetOrgCrossModuleSummary)
			r.Get("/context/{type}/{id:[0-9]+}", s.contextHandler.GetContext)
			r.Post("/intelligence/insight", s.contextHandler.GetInsight)
		}

		// Controlled AI Workflow Execution and Action Orchestration (Phase 3 Task 3.2)
		if s.orchestrationHandler != nil {
			r.Route("/orchestration", func(r chi.Router) {
				r.Post("/proposals", s.orchestrationHandler.GenerateProposal)
				r.Get("/proposals", s.orchestrationHandler.ListProposals)
				r.Get("/proposals/{id}", s.orchestrationHandler.GetProposal)
				r.Post("/proposals/{id}/execute", s.orchestrationHandler.ExecuteProposal)

				r.Get("/executions", s.orchestrationHandler.ListExecutions)
				r.Get("/executions/{id}", s.orchestrationHandler.GetExecution)
				r.Post("/executions/{id}/cancel", s.orchestrationHandler.CancelExecution)
				r.Post("/executions/{id}/retry", s.orchestrationHandler.RetryExecution)

				r.Get("/actions", s.orchestrationHandler.ListRegisteredActions)
			})
		}

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

		// Approvals internal bridge
		r.Post("/approvals/propose", s.approvalsHandler.ProposeAIApproval)

		// Centralized Action System (Task 0.5 Bridge)
		if s.actionsHandler != nil {
			r.Post("/actions/execute", s.actionsHandler.ExecuteAction)
			r.Get("/actions", s.actionsHandler.ListActions)
		}

		// AI Task Queue & Worker Lifecycle Bridge (Task 0.7)
		if s.aiTasksHandler != nil {
			r.Post("/ai/tasks/claim", s.aiTasksHandler.InternalClaimTask)
			r.Post("/ai/tasks/{id:[0-9]+}/heartbeat", s.aiTasksHandler.InternalHeartbeatTask)
			r.Post("/ai/tasks/{id:[0-9]+}/status", s.aiTasksHandler.InternalUpdateStatus)
			r.Post("/ai/tasks/recover-stale", s.aiTasksHandler.InternalRecoverStale)
			r.Get("/ai/tasks/{id:[0-9]+}", s.aiTasksHandler.InternalGetTask)
		}

		// Unified Business Context & Intelligence internal bridge (Task 1.1, 1.2, 1.3, 1.4, 1.5)
		if s.contextHandler != nil {
			r.Post("/context/retrieve", s.contextHandler.InternalGetContext)
			r.Post("/intelligence/insight", s.contextHandler.InternalGetInsight)
			r.Post("/customers/intelligence", s.contextHandler.InternalGetCustomerIntelligence)
			r.Post("/rfqs/intelligence", s.contextHandler.InternalGetRFQIntelligence)
			r.Post("/shipments/intelligence", s.contextHandler.InternalGetShipmentIntelligence)
			r.Post("/shipments/operations-summary", s.contextHandler.InternalGetOrgOperationsSummary)
			r.Post("/invoices/intelligence", s.contextHandler.InternalGetInvoiceIntelligence)
			r.Post("/invoices/finance-summary", s.contextHandler.InternalGetOrgFinanceSummary)
			r.Post("/contracts/intelligence", s.contextHandler.InternalGetContractIntelligence)
			r.Post("/contracts/compliance-summary", s.contextHandler.InternalGetOrgContractComplianceSummary)
			r.Post("/contracts/coverage", s.contextHandler.InternalGetContractCoverage)
			r.Post("/insights/cross-module", s.contextHandler.InternalGetCrossModuleInsights)
			r.Post("/insights/summary", s.contextHandler.InternalGetOrgCrossModuleSummary)
		}

		// AI Memory Internal Bridge (Phase 2 Task 2.10)
		if s.memoryHandler != nil {
			r.Post("/memory/runtime-context", s.memoryHandler.GetRuntimeContext)
		}
	})

	// ── Serve uploads statically (for local dev visual review) ─────────────────
	s.router.Mount("/uploads", http.StripPrefix("/uploads", http.FileServer(http.Dir("./uploads"))))
}
