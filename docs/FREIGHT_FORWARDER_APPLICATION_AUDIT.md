# Freel Freight Forwarder Application Audit
**Comprehensive Technical Linkage, Routing, Database & Backend Integration Report**
*Generated: August 15, 2026*

---

## 1. Executive Summary

This technical audit provides an exhaustive, source-code-verified inspection of the entire Freel Freight Forwarder (FF) web application and backend platform. Every finding, request flow, and table schema is grounded in direct codebase analysis.

### High-Level Summary Statistics

| Category | Count | Key Notes |
| :--- | :---: | :--- |
| **Total Sidebar Modules** | **18** | Across Operations, Commercial, Documents, Finance, and Admin |
| **Fully Connected Modules** | **5** | Dashboard, Leads, Reports, Users, Settings/Roles |
| **Partially Connected Modules** | **3** | RFQs (service URL bug), Shipments (shipment-level only), Customers/Companies |
| **Backend-Only Modules** | **2** | Rate Intelligence (`/api/v1/rates`), Contracts (`/api/v1/contracts`) |
| **Missing Frontend Routes** | **9** | Bookings, Tracking, Quotations, Rate Mgmt, Contracts, Templates, Approvals, Invoices, Payments |
| **Hardcoded UI Elements** | **1** | Sidebar Badge Counts (`12`, `8`, `23`, `2` are hardcoded in `Sidebar.jsx`) |
| **Critical Bugs Identified** | **2** | `Bookings` 404 navigation bug; `rfqService.js` missing `/api/v1` prefix |

---

## 2. Current Architecture Overview

Freel is architected as an **AI-powered Logistics Operating System** tailored for Freight Forwarders:
- **Frontend:** React 19 SPA powered by Vite 8, React Router v7, Context API for authentication & RBAC, and Lucide React iconography.
- **Backend:** Go 1.24 HTTP REST server structured using Go-Kit endpoint patterns and Chi / Gorilla Mux routing, with AWS Cognito JWT verification and PostgreSQL multi-tenant data isolation.
- **AI Sidecar:** Python 3.14 LangChain / LangGraph microservice coordinating Google Gemini (Primary), OpenAI (Failover), and deterministic fallbacks.

```
┌────────────────────────────────────────────────────────────────────────┐
│                          FREEL APPLICATION                             │
├──────────────────────────┬─────────────────────────┬───────────────────┤
│    FRONTEND (React 19)   │      BACKEND (Go)       │ AI SIDECAR (Py)   │
│  - AppShell / TopBar     │  - Chi + Mux Routing    │ - Pricing Agent   │
│  - Sidebar Navigation    │  - Cognito Auth Guard   │ - Sales Email Bot │
│  - New FF / Op Dashboard │  - Tenant Isolation     │ - Document OCR    │
│  - Module Workspaces     │  - PostgreSQL 16 DB     │ - Carrier Poller  │
└──────────────────────────┴─────────────────────────┴───────────────────┘
```

---

## 3. Frontend Architecture

### 3.1 Entry Point & Layouts
- **Entry File:** [`frontend/src/main.jsx`](file:///Users/varun.kanade/go/src/freel/freel-project/frontend/src/main.jsx) mounts `<App />`.
- **Root Router:** [`frontend/src/App.jsx`](file:///Users/varun.kanade/go/src/freel/freel-project/frontend/src/App.jsx) configures `BrowserRouter`, `AuthProvider`, and `RBACProvider`.
- **Layout Shell:** [`frontend/src/layouts/AppShell/AppShell.jsx`](file:///Users/varun.kanade/go/src/freel/freel-project/frontend/src/layouts/AppShell/AppShell.jsx) wraps all protected dashboard pages with:
  - [`Sidebar.jsx`](file:///Users/varun.kanade/go/src/freel/freel-project/frontend/src/layouts/AppShell/Sidebar.jsx): Deep navy (`#0A1128`) navigation sidebar.
  - [`TopBar.jsx`](file:///Users/varun.kanade/go/src/freel/freel-project/frontend/src/layouts/AppShell/TopBar.jsx): Top navigation with search, ⌘K palette, notification bell, and date selector.
  - `<Outlet />`: Main content workspace.

### 3.2 State & Route Guards
- **`AuthProvider` ([`AuthContext.jsx`](file:///Users/varun.kanade/go/src/freel/freel-project/frontend/src/context/AuthContext.jsx)):** Manages `user`, `org`, `memberRole`, `isBooting`, and background `/auth/me` profile hydration.
- **`RBACProvider` ([`RBACContext.jsx`](file:///Users/varun.kanade/go/src/freel/freel-project/frontend/src/context/RBACContext.jsx)):** Grants fine-grained permission checks via `can(resource, action)` based on `role.permissions[]`.
- **`ProtectedRoute` ([`ProtectedRoute.jsx`](file:///Users/varun.kanade/go/src/freel/freel-project/frontend/src/routes/ProtectedRoute.jsx)):** Redirects unauthenticated sessions to `/login`, and handles unauthorized access via `/dashboard/unauthorized`.

---

## 4. Backend Architecture

### 4.1 Server Routing & Middleware
- **Entry File:** [`backend/cmd/server/main.go`](file:///Users/varun.kanade/go/src/freel/freel-project/backend/cmd/server/main.go)
- **Route Registration:** [`backend/internal/server/routes.go`](file:///Users/varun.kanade/go/src/freel/freel-project/backend/internal/server/routes.go)
  - Public routes: `/health`, `/auth/signup`, `/auth/login`, `/auth/refresh`, `/auth/verify-email`, `/api/v1/emails/inbound`.
  - Protected routes (`/api/v1/*`): Enforced by `authGuard.RequireAuth`.
  - Internal RPC routes (`/internal/*`): Enforced by `InternalServiceAuthMiddleware`.

### 4.2 Auth & Tenant Middleware Pipeline
```
HTTP Request with Bearer Token
  ↓
[middleware.RequireAuth]
  - Validates Cognito JWT signature & claims (cognito_sub, email)
  - Queries DB: SELECT u.*, om.org_id, om.role_id FROM users u JOIN org_members om ...
  - Attaches UserContext { UserID, OrgID, RoleID, Email } to request context
  ↓
[middleware.GetUserContext(ctx)]
  - Safely extracts UserContext (handles value vs pointer context keys)
  ↓
[Handler / Endpoint]
  - Extracts OrgID and injects into Business Logic Request: req.OrgID = userCtx.OrgID
  ↓
[Repository Query]
  - Appends WHERE org_id = $1 (Strict Tenant Isolation)
```

---

## 5. Database Architecture & Schema Map

PostgreSQL contains 24 migration layers. Key operational tables and tenant scoping:

```
┌────────────────────────────────────────────────────────────────────────┐
│                        ORGANIZATION (Tenant Root)                      │
│                        `organizations` (id, name)                      │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │
         ┌──────────────────────────┼──────────────────────────┐
         ▼                          ▼                          ▼
┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐
│  USER / MEMBERS  │      │     COMMERCIAL   │      │    OPERATIONS    │
│  - `users`       │      │  - `companies`   │      │  - `shipments`   │
│  - `org_members` │      │  - `customers`   │      │  - `milestones`  │
│  - `roles`       │      │  - `contacts`    │      │  - `exceptions`  │
│  - `permissions` │      │  - `leads`       │      │  - `documents`   │
│  - `role_perms`  │      │  - `rfqs`        │      │  - `invoices`    │
│  - `invitations` │      │  - `rates`       │      │  - `activities`  │
└──────────────────┘      └──────────────────┘      └──────────────────┘
```

### Table Isolation Summary:
- **Tenant Root:** `organizations` (`id`, `name`, `created_at`).
- **All operational tables** (`leads`, `rfqs`, `shipments`, `companies`, `customers`, `contacts`, `contracts`, `addresses`, `audit_logs`, `activities`) enforce `org_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE`.

---

## 6. Complete Sidebar Module Audit

| # | Module | Sidebar Link | Frontend Route in `App.jsx` | Component File | API Endpoint | Backend Handler / Package | Database Tables | Auth & RBAC | Live Data Status | Overall Status |
| :-: | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **1** | **Dashboard** | `/dashboard` | `/dashboard` | `DashboardHome.jsx` | `GET /api/v1/dashboard/mission-control` | `dashboard/endpoints.go` | `rfqs`, `leads`, `shipments` | `RequireAuth` | Real (0 records = New FF) | **FULLY CONNECTED** |
| **2** | **Leads** | `/dashboard/leads` | `/dashboard/leads` | `LeadsPage.jsx` | `GET/POST /api/v1/leads` | `leads/endpoints.go` | `leads`, `lead_interactions` | `RequireAuth`, `LEADS.READ` | Real (0 leads in DB) | **FULLY CONNECTED** |
| **3** | **RFQs** | `/dashboard/rfqs` | `/dashboard/rfqs` | `RFQPage.jsx` | `GET/POST /api/v1/rfqs` | `rfq/endpoints.go` | `rfqs`, `rfq_items`, `customers` | `RequireAuth`, `RFQS.READ` | Real (Service URL prefix bug) | **PARTIALLY CONNECTED** |
| **4** | **Shipments** | `/dashboard/shipments` | `/dashboard/shipments` | `ShipmentsPage.jsx` | `GET /api/v1/shipments` | `shipments/handler.go` | `shipments`, `milestones` | `RequireAuth`, `SHIPMENTS.READ`| Real (0 records in DB) | **FULLY CONNECTED** |
| **5** | **Bookings** | `/dashboard/bookings` | ❌ *Missing* | ❌ *None* | ❌ *No dedicated endpoint* | Maps to `shipments` (status `BOOKED`) | `shipments` | `RequireAuth` | ❌ None | **ROUTE MISSING (404 Bug)** |
| **6** | **Tracking** | `/dashboard/tracking` | ❌ *Missing* | ❌ *None* | `GET /api/v1/shipments/{id}` | `shipments/handler.go` | `shipments`, `milestones` | `RequireAuth` | ❌ None | **ROUTE MISSING** |
| **7** | **Quotations** | `/dashboard/quotations` | ❌ *Missing* | `PricingWorkspace.jsx` (Sub) | `/api/v1/rfqs/{id}/quotes` | `pricing/handler.go` | `rfq_pricing_options` | `RequireAuth` | ❌ Unrouted | **ROUTE MISSING** |
| **8** | **Rate Mgmt** | `/dashboard/rate-management` | ❌ *Missing* | ❌ *None* | `GET /api/v1/rates/search` | `rates/handler.go` | `ports`, `contract_rates`, `spot_rates` | `RequireAuth` | Backend exists, UI missing | **BACKEND ONLY** |
| **9** | **Contracts** | `/dashboard/contracts` | ❌ *Missing* | `ContractsPage.jsx` | `GET/POST /api/v1/contracts` | `contracts/handler.go` | `carrier_contracts`, `reviews` | `RequireAuth` | Page exists, unrouted | **BACKEND ONLY / UNROUTED** |
| **10**| **Customers** | `/dashboard/companies` | `/dashboard/companies` | `PlaceholderPage` | `GET /api/v1/companies` (stub) | `routes.go:48` | `companies`, `customers` | `RequireAuth`, `COMPANIES.READ` | Placeholder | **PARTIALLY CONNECTED** |
| **11**| **Documents** | `/dashboard/documents` | `/dashboard/documents` | `PlaceholderPage` | `GET /api/v1/shipments/{id}/documents` | `documents/handler.go` | `shipment_documents` | `RequireAuth`, `DOCUMENTS.READ`| Standalone UI placeholder | **PARTIALLY CONNECTED** |
| **12**| **Templates** | `/dashboard/templates` | ❌ *Missing* | ❌ *None* | ❌ *None* | ❌ *None* | ❌ *None* | ❌ *None* | ❌ None | **NOT IMPLEMENTED** |
| **13**| **Approvals** | `/dashboard/approvals` | ❌ *Missing* | `ReviewModal.jsx` (Sub) | `/api/v1/contracts/review` | `contracts/handler.go`, `finance` | `reviews`, `discrepancies` | `RequireAuth` | Multi-module backend stubs | **ROUTE MISSING** |
| **14**| **Invoices** | `/dashboard/invoices` | ❌ *Missing* | `FinanceWorkspace.jsx` (Sub) | `GET /api/v1/shipments/{id}/finance` | `finance/handler.go`, `billing` | `shipment_invoices`, `cust_invoices` | `RequireAuth` | Shipment sub-view only | **ROUTE MISSING** |
| **15**| **Payments** | `/dashboard/payments` | ❌ *Missing* | `BillingWorkspace.jsx` (Sub) | `POST /api/v1/billing/invoices/{id}/pay` | `billing/handler.go` | `customer_invoices` | `RequireAuth` | Shipment sub-view only | **ROUTE MISSING** |
| **16**| **Reports** | `/dashboard/reports` | `/dashboard/reports` | `ReportsPage.jsx` | `GET /api/v1/reports/metrics` | `reports/endpoints.go` | `rfqs`, `leads`, `shipments` | `RequireAuth`, `DASHBOARD.READ`| Real (0 counts) | **FULLY CONNECTED** |
| **17**| **Users** | `/dashboard/users` | `/dashboard/users` | `UsersPage.jsx` | `GET /api/v1/users` | `users/endpoints.go` | `users`, `org_members`, `roles` | `RequireAuth`, `USERS.READ` | Real (returns Super Admin) | **FULLY CONNECTED** |
| **18**| **Settings** | `/dashboard/settings` | `/dashboard/settings` | `RolesPage.jsx` | `GET /api/v1/roles` | `rbac/endpoints.go` | `roles`, `permissions` | `RequireAuth`, `SETTINGS.READ` | Real (38 permissions) | **FULLY CONNECTED** |

---

## 7. Complete Route Audit (`App.jsx`)

| Route Path | Route Component | Protected | Required Module / Action | Exists in Code? | Status |
| :--- | :--- | :---: | :---: | :---: | :--- |
| `/` | `RootRedirect` (Landing or `/dashboard`) | No | — | Yes | **Active** |
| `/login` | `LoginPage` | Public Only | — | Yes | **Active** |
| `/signup` | `SignupPage` | Public Only | — | Yes | **Active** |
| `/verify-email` | `VerifyEmailPage` | Public Only | — | Yes | **Active** |
| `/forgot-password` | `ForgotPasswordPage` | Public Only | — | Yes | **Active** |
| `/reset-password` | `ResetPasswordPage` | Public Only | — | Yes | **Active** |
| `/dashboard` | `DashboardHome` | Yes | — | Yes | **Active** |
| `/dashboard/reports` | `ReportsPage` | Yes | `DASHBOARD.READ` | Yes | **Active** |
| `/dashboard/leads` | `LeadsPage` | Yes | `LEADS.READ` | Yes | **Active** |
| `/dashboard/rfqs` | `RFQPage` | Yes | `RFQS.READ` | Yes | **Active** |
| `/dashboard/outreach`| `OutreachPage` | Yes | `OUTREACH.READ` | Yes | **Active** |
| `/dashboard/shipments`| `ShipmentsPage` | Yes | — | Yes | **Active** |
| `/dashboard/shipments/:id` | `ShipmentDetail` | Yes | — | Yes | **Active** |
| `/dashboard/companies` | `PlaceholderPage` | Yes | `COMPANIES.READ` | Yes | **Placeholder** |
| `/dashboard/documents` | `PlaceholderPage` | Yes | `DOCUMENTS.READ` | Yes | **Placeholder** |
| `/dashboard/users` | `UsersPage` | Yes | `USERS.READ` | Yes | **Active** |
| `/dashboard/settings` | `RolesPage` | Yes | `USERS.READ` / `SETTINGS` | Yes | **Active** |
| `/dashboard/bookings` | ❌ *None* | — | — | **NO** | **404 Page Not Found** |
| `/dashboard/tracking` | ❌ *None* | — | — | **NO** | **404 Page Not Found** |
| `/dashboard/quotations`| ❌ *None* | — | — | **NO** | **404 Page Not Found** |
| `/dashboard/rate-management` | ❌ *None* | — | — | **NO** | **404 Page Not Found** |
| `/dashboard/contracts`| ❌ *None* (`ContractsPage` exists) | — | — | **NO** | **404 Page Not Found** |
| `/dashboard/templates`| ❌ *None* | — | — | **NO** | **404 Page Not Found** |
| `/dashboard/approvals`| ❌ *None* | — | — | **NO** | **404 Page Not Found** |
| `/dashboard/invoices` | ❌ *None* | — | — | **NO** | **404 Page Not Found** |
| `/dashboard/payments` | ❌ *None* | — | — | **NO** | **404 Page Not Found** |

---

## 8. Complete Backend API Inventory

### Auth Subsystem ([`backend/internal/auth/handler.go`](file:///Users/varun.kanade/go/src/freel/freel-project/backend/internal/auth/handler.go))
- `POST /auth/signup` — Registers user with Cognito + returns confirmation requirement.
- `POST /auth/verify-email` — Confirms Cognito email code.
- `POST /auth/login` — Authenticates with Cognito, auto-provisions Org/User/Role in DB, returns JWT.
- `POST /auth/refresh` — Computes Cognito `SECRET_HASH` and rotates tokens.
- `POST /auth/forgot-password` — Triggers Cognito password reset code.
- `POST /auth/reset-password` — Confirms password reset.
- `GET /auth/me` (Protected) — Hydrates authenticated user, org, and permissions.

### Leads & Sales CRM ([`backend/internal/leads/`](file:///Users/varun.kanade/go/src/freel/freel-project/backend/internal/leads/))
- `GET /api/v1/leads` — Lists leads for `UserContext.OrgID`.
- `POST /api/v1/leads` — Creates a lead and spawns background AI lead qualification job.
- `GET /api/v1/leads/{id}` — Gets lead details with AI score and interactions.
- `PUT /api/v1/leads/{id}` — Updates lead status / notes.
- `DELETE /api/v1/leads/{id}` — Deletes lead scoped to org.
- `POST /api/v1/leads/import` — Multipart CSV ingestion with automated async lead qualification.
- `GET /api/v1/leads/{id}/interactions` — Lists communications and AI auto-drafted responses.
- `POST /api/v1/leads/{id}/interactions` — Manual or outbound email logging.
- `POST /api/v1/emails/inbound` — Public webhook for parsing inbound customer emails into leads/RFQs.

### RFQ & Pricing Engine ([`backend/internal/rfq/`](file:///Users/varun.kanade/go/src/freel/freel-project/backend/internal/rfq/) & [`backend/internal/pricing/`](file:///Users/varun.kanade/go/src/freel/freel-project/backend/internal/pricing/))
- `GET /api/v1/rfqs` — Lists RFQs with stage, origin/dest, and customer metadata.
- `POST /api/v1/rfqs` — Initiates an RFQ, items, and kicks off async rate lookup.
- `GET /api/v1/rfqs/{id}` — Returns RFQ items, quotes, and timeline.
- `PUT /api/v1/rfqs/{id}/stage` — Advances RFQ through stages (`CREATED` → `PRICING` → `QUOTED` → `WON`/`LOST`).
- `POST /api/v1/rfqs/parse-shipment-request` — AI OCR/email text extraction into structured RFQ.
- `POST /api/v1/rfqs/{id}/quotes` — Adds carrier quotes.
- `GET /api/v1/rfqs/{id}/timeline` — Unified activity stream.

### Rate Intelligence & Contracts ([`backend/internal/rates/`](file:///Users/varun.kanade/go/src/freel/freel-project/backend/internal/rates/) & [`backend/internal/contracts/`](file:///Users/varun.kanade/go/src/freel/freel-project/backend/internal/contracts/))
- `GET /api/v1/rates/search` — Queries contract and spot rates matching origin/dest/equipment.
- `POST /api/v1/rates/spot/refresh` — Triggers spot rate refresh for trade lanes.
- `GET /api/v1/rates/{id}` — Single rate card detail.
- `POST /api/v1/contracts/upload` — Multipart PDF/XLSX contract upload + AI extraction queue.
- `GET /api/v1/contracts` — Lists carrier contract documents.
- `GET /api/v1/contracts/review` — Lists low-confidence extractions requiring human review.
- `PUT /api/v1/contracts/review/{id}/approve` — Commits reviewed contract rate cards to DB.
- `PUT /api/v1/contracts/review/{id}/reject` — Rejects extraction with notes.

### Shipment Operations & Milestone Tracking ([`backend/internal/shipments/`](file:///Users/varun.kanade/go/src/freel/freel-project/backend/internal/shipments/))
- `GET /api/v1/shipments` — Lists active/delivered shipments.
- `GET /api/v1/shipments/{id}` — Shipment tracking, milestones, container statuses, and exceptions.
- `POST /api/v1/shipments/{id}/carrier-update` — Ingests carrier EDI/API milestone payload.
- `POST /api/v1/shipments/exceptions/{id}/resolve` — Resolves operational exception flag.
- `POST /webhooks/carriers/{carrier}` — Public carrier webhook receiver.

### Documents, Finance & Billing ([`backend/internal/documents/`](file:///Users/varun.kanade/go/src/freel/freel-project/backend/internal/documents/), [`finance/`](file:///Users/varun.kanade/go/src/freel/freel-project/backend/internal/finance/), [`billing/`](file:///Users/varun.kanade/go/src/freel/freel-project/backend/internal/billing/))
- `POST /api/v1/shipments/{id}/documents/upload` — Uploads HBL, MBL, Invoice, Packing List.
- `GET /api/v1/shipments/{id}/documents` — Lists shipment documents with compliance validation status.
- `POST /api/v1/shipments/discrepancies/{id}/resolve` — Resolves document discrepancies.
- `POST /api/v1/shipments/{id}/finance/invoices/upload` — Ingests carrier AP invoices + 3-way matching.
- `GET /api/v1/shipments/{id}/finance` — AP invoice ledger, variance analysis, margin checks.
- `POST /api/v1/shipments/{id}/billing/invoices/generate` — Generates AR invoice for customer.
- `GET /api/v1/shipments/{id}/billing` — Customer billing workspace.
- `POST /api/v1/billing/invoices/{id}/approve` — Approves customer invoice.
- `POST /api/v1/billing/invoices/{id}/pay` — Records customer payment against invoice.
- `POST /api/v1/shipments/{id}/close` — Closes shipment after AR/AP settlement.

---

## 9. Specific Root Cause Investigations

### 9.1 Bookings 404 Bug (Root Cause)
- **Investigation:** When clicking "Bookings" in the sidebar, the browser navigates to `/dashboard/bookings`.
- **Root Cause:** In [`frontend/src/App.jsx`](file:///Users/varun.kanade/go/src/freel/freel-project/frontend/src/App.jsx), `<Route path="/dashboard/bookings" ... />` is **completely omitted** from the route tree.
- **Result:** React Router cannot match `/dashboard/bookings` under `<AppShell />`, bubbles up to the wildcard `<Route path="*" element={<NotFoundPage />} />` in `PublicLayout`, rendering the public 404 page.

### 9.2 Sidebar Badges (Real vs Hardcoded)
- **Investigation:** Sidebar renders badges `12` (Leads), `8` (RFQs), `23` (Shipments), and `2` (Bookings).
- **Root Cause:** In [`frontend/src/layouts/AppShell/Sidebar.jsx`](file:///Users/varun.kanade/go/src/freel/freel-project/frontend/src/layouts/AppShell/Sidebar.jsx), lines 10–13 hardcode:
  ```javascript
  { path: '/dashboard/leads', label: 'Leads', Icon: Target, badge: '12' },
  { path: '/dashboard/rfqs', label: 'RFQs', Icon: FileText, badge: '8' },
  { path: '/dashboard/shipments', label: 'Shipments', Icon: Ship, badge: '23' },
  { path: '/dashboard/bookings', label: 'Bookings', Icon: Package, badge: '2' },
  ```
- **Conclusion:** These badge numbers are **100% hardcoded static constants**. No dynamic API call was connected to hydrate them from the backend.

### 9.3 Why Leads Page Shows 0
- **Investigation:** Leads page calls `listLeads()` (`GET /api/v1/leads?limit=100`).
- **Backend Flow:** `makeListLeadsEndpoint` → `service.ListLeads` → `SELECT * FROM leads WHERE org_id = 2`.
- **Conclusion:** Count `0` is **completely legitimate and accurate**. The database was recently wiped for fresh end-to-end testing, and ABC Logistics has not created any leads yet.

### 9.4 Why RFQs Page Shows 0 / Failed to Load
- **Investigation:** In [`frontend/src/services/rfqService.js`](file:///Users/varun.kanade/go/src/freel/freel-project/frontend/src/services/rfqService.js), line 11 calls `api.get('/rfqs?limit=...')` instead of `api.get('/api/v1/rfqs?limit=...')`.
- **Result:** Calling `/rfqs` hits `http://localhost:8080/rfqs` which returns a backend 404 error, causing the page to toast `"Failed to load RFQs"` and default state to `[]`.
- **Database Status:** In addition, the PostgreSQL `rfqs` table for org 2 currently has 0 records.

---

## 10. Security & Tenant Isolation Findings

1. **Strict Tenant Scoping Active:** All operational endpoints in `leads`, `rfq`, `shipments`, `contracts`, `rates`, `reports`, `notifications`, `documents`, and `finance` successfully extract `OrgID` via `middleware.GetUserContext(ctx)` and filter SQL queries with `WHERE org_id = $1`.
2. **Permission RBAC Active:** Routes in `server/routes.go` and components in `App.jsx` enforce RBAC permissions (`LEADS.READ`, `RFQS.READ`, `SHIPMENTS.READ`, `USERS.READ`, `SETTINGS.READ`, `DASHBOARD.READ`).
3. **No Cross-Tenant Leaks Found:** No queries were found using hardcoded organization IDs or global un-scoped `SELECT * FROM ...` without `WHERE org_id = $1`.

---

## 11. Backend → Frontend Gap Analysis

```
┌──────────────────────────────────────────────┬──────────────────────────────────────────────┐
│        BACKEND READY (NO DEDICATED UI)       │       FRONTEND READY (UNROUTED / BUGGED)     │
├──────────────────────────────────────────────┼──────────────────────────────────────────────┤
│ 1. Rate Intelligence Engine (/api/v1/rates)  │ 1. Contracts Page (ContractsPage.jsx exists) │
│ 2. Contract Review Queue (/contracts/review) │ 2. RFQ Service API Prefix (/api/v1/rfqs)     │
│ 3. Billing & Invoicing Generation API        │ 3. Bookings Route Missing (App.jsx)          │
│ 4. Document Discrepancy Resolution API       │ 4. Standalone Quotations Route               │
│ 5. Carrier Webhook Status Receiver           │ 5. Dynamic Sidebar Badge Hydration           │
└──────────────────────────────────────────────┴──────────────────────────────────────────────┘
```

---

## 12. Prioritized Implementation Plan

### P0 — Immediate Fixes (Bugs & Unrouted Existing Code)
1. **Fix Bookings 404:** Add `<Route path="/dashboard/bookings" element={<ShipmentsPage filter="BOOKED" />} />` in `App.jsx`.
2. **Fix `rfqService.js` Prefix:** Update `/rfqs` → `/api/v1/rfqs` across `rfqService.js`.
3. **Register `ContractsPage.jsx`:** Add `<Route path="/dashboard/contracts" element={<ContractsPage />} />` in `App.jsx`.
4. **Dynamic Sidebar Badges:** Connect sidebar badges to dynamic counts returned by `GET /api/v1/dashboard/mission-control` or remove hardcoded numbers when counts are 0.

### P1 — Required for Freight Forwarder MVP Workflow
1. **Leads Management:** Verify Add Lead and CSV Import to create real customer leads.
2. **RFQ Creation & Sourcing:** Enable FF to create an RFQ and link to customer records.
3. **Rate Intelligence & Quotations UI:** Connect Rate Search (`/api/v1/rates/search`) and Pricing Workspace to generate quotes.
4. **Shipment Booking & Milestones:** Transition Won RFQ into an active shipment.

### P2 — Operational & Compliance Depth
1. **Contract Extraction Review UI:** Wire `ReviewModal.jsx` to `/api/v1/contracts/review`.
2. **Dedicated Invoices & Billing Page:** Expose top-level AR/AP invoice ledger.
3. **Stand-alone Document Hub:** Connect Document Generator to `/api/v1/shipments/{id}/documents`.

### P3 — Future & Advanced Capabilities
1. **Automated Carrier API Integrations:** Live API polling for Ocean/Air carriers.
2. **Customs Clearance & Trade Finance Automation.**
3. **Multi-carrier EDI 204/315/214 Integration.**

---

## 13. Final Module Status Matrix

| Module | UI Component | Route Registered | Frontend Service | Backend Route | DB Schema | Auth Scoped | Org Filtered | Status |
| :--- | :--- | :---: | :---: | :---: | :---: | :---: | :---: | :--- |
| **Dashboard** | `DashboardHome.jsx` | ✅ | `dashboardService.js` | `/api/v1/dashboard/*` | `rfqs`, `leads`, `shipments` | ✅ | ✅ | **FULLY CONNECTED** |
| **Leads** | `LeadsPage.jsx` | ✅ | `leadsService.js` | `/api/v1/leads/*` | `leads`, `interactions` | ✅ | ✅ | **FULLY CONNECTED** |
| **RFQs** | `RFQPage.jsx` | ✅ | `rfqService.js` | `/api/v1/rfqs/*` | `rfqs`, `rfq_items` | ✅ | ✅ | **PARTIALLY CONNECTED** |
| **Shipments** | `ShipmentsPage.jsx` | ✅ | `api.js` | `/api/v1/shipments/*` | `shipments`, `milestones` | ✅ | ✅ | **FULLY CONNECTED** |
| **Bookings** | ❌ Missing | ❌ | ❌ Missing | `/api/v1/shipments?status=BOOKED` | `shipments` | ✅ | ✅ | **ROUTE MISSING** |
| **Tracking** | ❌ Missing | ❌ | ❌ Missing | `/api/v1/shipments/{id}` | `shipments`, `milestones` | ✅ | ✅ | **ROUTE MISSING** |
| **Quotations** | `PricingWorkspace.jsx` | ❌ | `rfqService.js` | `/api/v1/rfqs/{id}/quotes` | `rfq_pricing_options` | ✅ | ✅ | **ROUTE MISSING** |
| **Rate Mgmt** | ❌ Missing | ❌ | ❌ Missing | `/api/v1/rates/*` | `ports`, `rates` | ✅ | ✅ | **BACKEND ONLY** |
| **Contracts** | `ContractsPage.jsx` | ❌ | `contractsService.js` | `/api/v1/contracts/*` | `carrier_contracts` | ✅ | ✅ | **UNROUTED** |
| **Customers** | `PlaceholderPage` | ✅ | ❌ Missing | `/api/v1/companies` (stub) | `companies`, `customers` | ✅ | ✅ | **PARTIALLY CONNECTED** |
| **Documents** | `PlaceholderPage` | ✅ | ❌ Missing | `/api/v1/shipments/{id}/documents` | `shipment_documents` | ✅ | ✅ | **PARTIALLY CONNECTED** |
| **Templates** | ❌ Missing | ❌ | ❌ Missing | ❌ Missing | ❌ Missing | — | — | **NOT IMPLEMENTED** |
| **Approvals** | `ReviewModal.jsx` | ❌ | `contractsService.js` | `/api/v1/contracts/review` | `reviews`, `discrepancies` | ✅ | ✅ | **ROUTE MISSING** |
| **Invoices** | `FinanceWorkspace.jsx` | ❌ | ❌ Missing | `/api/v1/shipments/{id}/finance` | `shipment_invoices` | ✅ | ✅ | **ROUTE MISSING** |
| **Payments** | `BillingWorkspace.jsx` | ❌ | ❌ Missing | `/api/v1/billing/*` | `customer_invoices` | ✅ | ✅ | **ROUTE MISSING** |
| **Reports** | `ReportsPage.jsx` | ✅ | `api.js` | `/api/v1/reports/*` | `rfqs`, `leads`, `shipments` | ✅ | ✅ | **FULLY CONNECTED** |
| **Users** | `UsersPage.jsx` | ✅ | `api.js` | `/api/v1/users/*` | `users`, `org_members` | ✅ | ✅ | **FULLY CONNECTED** |
| **Settings** | `RolesPage.jsx` | ✅ | `api.js` | `/api/v1/roles/*` | `roles`, `permissions` | ✅ | ✅ | **FULLY CONNECTED** |

---

## 14. Files Inspected

### Frontend Source Files
- `frontend/src/App.jsx`
- `frontend/src/main.jsx`
- `frontend/src/context/AuthContext.jsx`
- `frontend/src/context/RBACContext.jsx`
- `frontend/src/routes/ProtectedRoute.jsx`
- `frontend/src/layouts/AppShell/AppShell.jsx`
- `frontend/src/layouts/AppShell/Sidebar.jsx`
- `frontend/src/layouts/AppShell/TopBar.jsx`
- `frontend/src/pages/dashboard/Home/DashboardHome.jsx`
- `frontend/src/pages/dashboard/Home/NewFFDashboard.jsx`
- `frontend/src/pages/dashboard/Home/OperationalDashboard.jsx`
- `frontend/src/pages/dashboard/Leads/LeadsPage.jsx`
- `frontend/src/pages/dashboard/RFQ/RFQPage.jsx`
- `frontend/src/pages/dashboard/Contracts/ContractsPage.jsx`
- `frontend/src/pages/dashboard/Shipments/ShipmentsPage.jsx`
- `frontend/src/pages/dashboard/Shipments/ShipmentDetail.jsx`
- `frontend/src/pages/dashboard/Settings/UsersPage.jsx`
- `frontend/src/pages/dashboard/Settings/RolesPage.jsx`
- `frontend/src/pages/dashboard/Reports/ReportsPage.jsx`
- `frontend/src/services/api.js`
- `frontend/src/services/leadsService.js`
- `frontend/src/services/rfqService.js`
- `frontend/src/services/contractsService.js`
- `frontend/src/services/dashboardService.js`
- `frontend/src/services/outreachService.js`

### Backend Source Files
- `backend/cmd/server/main.go`
- `backend/internal/server/routes.go`
- `backend/internal/middleware/auth.go`
- `backend/internal/middleware/rbac.go`
- `backend/internal/auth/handler.go`
- `backend/internal/auth/service.go`
- `backend/internal/dashboard/endpoints.go`
- `backend/internal/dashboard/service.go`
- `backend/internal/leads/endpoints.go`
- `backend/internal/leads/service.go`
- `backend/internal/leads/repository.go`
- `backend/internal/rfq/endpoints.go`
- `backend/internal/rfq/service.go`
- `backend/internal/rates/handler.go`
- `backend/internal/rates/service.go`
- `backend/internal/contracts/handler.go`
- `backend/internal/contracts/service.go`
- `backend/internal/shipments/handler.go`
- `backend/internal/shipments/service.go`
- `backend/internal/documents/handler.go`
- `backend/internal/finance/handler.go`
- `backend/internal/billing/handler.go`
- `backend/internal/reports/endpoints.go`
- `backend/internal/database/migrations/*` (24 migration files)
- `backend/up.sql`
