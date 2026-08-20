# LogisticsHQ — Freight Forwarder Live Module Connectivity Matrix
**Complete 11-Stage Verification Chain Across All 18 Workspaces**
*Updated: August 15, 2026*

| # | Sidebar Path | React Route (`App.jsx`) | Page Component | Frontend Service | HTTP API | Go Route (`server/routes.go`) | Go Handler | Service | Repository | PostgreSQL Tables | Status |
| :-: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **1** | `/dashboard` | `/dashboard` | `DashboardHome` | `dashboardService.js` | `GET /api/v1/dashboard/mission-control` | `r.Mount("/dashboard", ...)` | `dashboard.Endpoints` | `dashboard.Service` | `dashboard.Repository` | `rfqs`, `leads`, `shipments`, `organizations` | **FULLY CONNECTED** |
| **2** | `/dashboard/leads` | `/dashboard/leads` | `LeadsPage` | `leadsService.js` | `GET/POST /api/v1/leads`, `/import` | `r.Mount("/leads", ...)` | `leads.Endpoints` | `leads.Service` | `leads.Repository` | `leads`, `interactions`, `lead_notes` | **FULLY CONNECTED** |
| **3** | `/dashboard/rfqs` | `/dashboard/rfqs` | `RFQPage` | `rfqService.js` | `GET/POST /api/v1/rfqs`, `/{id}/stage` | `r.Mount("/rfqs", ...)` | `rfq.Endpoints` | `rfq.Service` | `rfq.Repository` | `rfqs`, `rfq_items`, `customers` | **FULLY CONNECTED** |
| **4** | `/dashboard/shipments` | `/dashboard/shipments` | `ShipmentsPage` | `api.js` | `GET /api/v1/shipments`, `/{id}` | `r.Mount("/shipments", ...)` | `shipments.Handler` | `shipments.Service` | `shipments.Repository` | `shipments`, `milestones`, `containers` | **FULLY CONNECTED** |
| **5** | `/dashboard/bookings` | `/dashboard/bookings` | `ShipmentsPage` (Bookings mode) | `api.js` | `GET /api/v1/shipments` (`status=BOOKED`) | `r.Mount("/shipments", ...)` | `shipments.Handler` | `shipments.Service` | `shipments.Repository` | `shipments` (`status='BOOKED'`) | **FULLY CONNECTED** |
| **6** | `/dashboard/tracking` | `/dashboard/tracking` | `TrackingPage` | `api.js` | `GET /api/v1/shipments`, `/{id}` | `r.Mount("/shipments", ...)` | `shipments.Handler` | `shipments.Service` | `shipments.Repository` | `shipments`, `milestones` | **FULLY CONNECTED** |
| **7** | `/dashboard/quotations` | `/dashboard/quotations` | `QuotationsPage` | `rfqService.js` | `GET /api/v1/rfqs`, `/{id}/quotes` | `r.Mount("/rfqs", ...)` | `pricing.Handler` / `rfq.Endpoints` | `pricing.Service` | `pricing.Repository` | `rfqs`, `rfq_pricing_options` | **FULLY CONNECTED** |
| **8** | `/dashboard/rate-management` | `/dashboard/rate-management` | `RateManagementPage` | `api.js` | `GET /api/v1/rates/search`, `/spot/refresh` | `r.Mount("/rates", ...)` | `rates.Handler` | `rates.Service` | `rates.Repository` | `ports`, `contract_rates`, `spot_rates` | **FULLY CONNECTED** |
| **9** | `/dashboard/contracts` | `/dashboard/contracts` | `ContractsPage` | `contractsService.js` | `GET/POST /api/v1/contracts` | `r.Mount("/contracts", ...)` | `contracts.Handler` | `contracts.Service` | `contracts.Repository` | `contract_documents`, `rate_entries` | **FULLY CONNECTED** |
| **10**| `/dashboard/companies` | `/dashboard/companies` | `CustomersPage` | `api.js` | `GET /api/v1/companies` | `r.Get("/companies", ...)` | `server.routes.go` (SQL handler) | — | Direct SQL query | `companies`, `contacts`, `customers` | **FULLY CONNECTED** |
| **11**| `/dashboard/documents` | `/dashboard/documents` | `DocumentsPage` | `api.js` | `GET /api/v1/shipments` (docs attached) | `r.Mount("/shipments", ...)` | `documents.Handler` | `documents.Service` | `documents.Repository` | `shipment_documents`, `discrepancies` | **FULLY CONNECTED** |
| **12**| `/dashboard/templates` | `/dashboard/templates` | `TemplatesPage` | — | — | — | — | — | — | Template Studio UI (Standard formats) | **FRONTEND ONLY (Planned)** |
| **13**| `/dashboard/approvals` | `/dashboard/approvals` | `ApprovalsPage` | `contractsService.js` | `GET /api/v1/contracts/review?status=PENDING` | `r.Mount("/contracts", ...)` | `contracts.Handler` | `contracts.Service` | `contracts.Repository` | `contract_review_items`, `rate_entries` | **FULLY CONNECTED** |
| **14**| `/dashboard/invoices` | `/dashboard/invoices` | `InvoicesPage` | `api.js` | `GET /api/v1/shipments` (invoices) | `r.Mount("/shipments", ...)` | `finance.Handler` / `billing` | `finance.Service` | `finance.Repository` | `shipment_invoices`, `cust_invoices` | **FULLY CONNECTED** |
| **15**| `/dashboard/payments` | `/dashboard/payments` | `PaymentsPage` | `api.js` | `GET /api/v1/shipments`, `/billing/.../pay` | `r.Mount("/shipments", ...)` | `billing.Handler` | `billing.Service` | `billing.Repository` | `customer_invoices`, `shipments` | **FULLY CONNECTED** |
| **16**| `/dashboard/reports` | `/dashboard/reports` | `ReportsPage` | `api.js` | `GET /api/v1/reports/metrics` | `r.Mount("/reports", ...)` | `reports.Endpoints` | `reports.Service` | `reports.Repository` | `rfqs`, `leads`, `shipments` | **FULLY CONNECTED** |
| **17**| `/dashboard/users` | `/dashboard/users` | `UsersPage` | `api.js` | `GET /api/v1/users`, `/invite` | `r.Mount("/users", ...)` | `users.Endpoints` / `auth` | `users.Service` | `users.Repository` | `users`, `org_members`, `roles` | **FULLY CONNECTED** |
| **18**| `/dashboard/settings` | `/dashboard/settings` | `RolesPage` | `api.js` | `GET /api/v1/roles`, `/permissions` | `r.Mount("/roles", ...)` | `rbac.Endpoints` | `rbac.Service` | `rbac.Repository` | `roles`, `permissions`, `role_permissions`| **FULLY CONNECTED** |
| **—** | `/dashboard/*` | `/dashboard/*` | `DashboardNotFound` | — | — | — | — | — | — | Client-side AppShell 404 Catch-All | **AUTHENTICATED FALLBACK** |

---

### Tenant Isolation & Security Summary
- **Authentication:** All `/api/v1/*` Go endpoints are guarded by `authGuard.RequireAuth` with AWS Cognito JWT validation.
- **Tenant Scoping:** Handlers and repositories derive `org_id` exclusively from `middleware.GetUserContext(ctx)` and append `WHERE org_id = $1` to all SQL queries.
- **RBAC:** Guarded by `rbacGuard.RequirePermission(Resource, Action)` mapping to database-backed roles (`SUPER_ADMIN`, `OPERATIONS_MANAGER`, `CUSTOMS_BROKER`, etc.).
