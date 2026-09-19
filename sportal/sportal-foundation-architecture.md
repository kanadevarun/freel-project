# SPortal Foundation Architecture Specification
## LogisticsHQ Internal SaaS Administration Platform

---

## 1. Executive Summary & Purpose

**SPortal** is the dedicated internal SaaS administration portal for the **LogisticsHQ** platform. While **CPortal** serves external freight-forwarding customer organizations and their operational teams, **SPortal** is the internal control center used by the LogisticsHQ engineering, operations, executive, customer success, and finance teams to manage customer organizations, onboarding, subscriptions, platform telemetry, billing, user security, and enterprise AI controls.

This architecture establishes SPortal as an independent application parallel to `frontend/` (CPortal) and `backend/` (Go API), adhering to the principle of a **shared authoritative source of truth**:
- **Single Source of Truth**: SPortal and CPortal operate against the exact same authoritative MariaDB schema (`freel_mysql`) via the unified Go backend API.
- **Zero Architectural Duplication**: No duplicate backend servers, no separate databases, and no duplicate entity tables.
- **Strict Separation of Concerns**: Customer organization users cannot access internal SPortal routes; internal staff identities are strictly distinguished from tenant organization roles.

---

## 2. Product Distinction: SPortal vs. CPortal

| Dimension | CPortal (`frontend/`) | SPortal (`sportal/`) |
| :--- | :--- | :--- |
| **Primary Audience** | LogisticsHQ Customers (Freight Forwarders, 3PLs, Brokers, Cargo Shippers) | LogisticsHQ Internal Team (Founders, Support, Engineering, Operations, Customer Success) |
| **Operational Domain** | Day-to-day international freight forwarding operations | Global multi-tenant SaaS business & platform administration |
| **Key Functional Areas** | Leads, Customers, RFQs, Quotations, Bookings, Shipments, Milestones, Exceptions, Invoices, Contracts, Compliance, AI Workforce | Tenant Directory, Customer Onboarding, SaaS Subscriptions, Platform Revenue, Global Users, Customer 360, Usage Metering, Customer Health, Global Integrations, Audit Trail, SPortal AI, Global Settings |
| **Data Scope** | Single Tenant Isolation (`WHERE org_id = ?`) | Cross-Tenant Aggregation & Direct Tenant Administration |
| **Development URL** | `http://localhost:5173` | `http://localhost:5174` |
| **Production Target** | `https://app.logisticshq.in` | `https://sportal.logisticshq.in` |

---

## 3. Repository Structure

The repository architecture places `sportal/` at the top level parallel to `frontend/` and `backend/`:

```
LogisticsHQ/
├── backend/                     # Authoritative Go API server (Chi v5, sqlx, MariaDB 12.3)
│   ├── cmd/server/main.go       # Server entry point with SPortal route registration
│   ├── internal/
│   │   ├── sportal/             # SPortal backend domain module
│   │   │   ├── handler.go       # HTTP handler for SPortal endpoints
│   │   │   ├── repository.go    # Data access querying authoritative MariaDB tables
│   │   │   ├── service.go       # Business logic & internal staff role validation
│   │   │   ├── service_test.go  # Unit test suite
│   │   │   └── types.go         # Domain entities & API contract definitions
│   │   ├── server/              # Server configuration, middleware & route mounting
│   │   └── ...                  # Existing 58 Go domain packages
├── frontend/                    # Existing CPortal application (Vite, React 19, Port 5173)
├── sportal/                     # NEW SPortal application (Vite, React 19, Port 5174)
│   ├── public/                  # Favicons, branding assets
│   ├── src/
│   │   ├── app/
│   │   │   ├── config/env.js    # SPortal environment and route configuration
│   │   │   ├── providers/       # AppProviders, AuthProvider, Toaster
│   │   │   └── routes/index.jsx # Complete 14-module routing architecture
│   │   ├── components/
│   │   │   ├── common/          # StatusBadge, LoadingSpinner, HonestPlaceholder
│   │   │   ├── layout/          # SPortalLayout (Dark navy sidebar, Header, Footer)
│   │   │   └── navigation/      # Sidebar (14 items), Header (Search, Profile)
│   │   ├── contexts/            # AuthContext (internal staff session)
│   │   ├── features/            # Feature modules (Dashboard, Organizations, Subscriptions, etc.)
│   │   │   └── dashboard/       # DashboardOverview with live MariaDB KPI integration
│   │   ├── services/            # api.js, sportalService.js
│   │   ├── styles/              # Tailwind CSS v4 design system
│   │   ├── App.jsx              # Application root
│   │   ├── main.jsx             # React entry point
│   │   └── index.html           # HTML template
│   ├── package.json             # Independent package definitions (React 19, Vite, Lucide, Recharts)
│   ├── vite.config.js           # Dedicated Vite config running on Port 5174
│   ├── sportal-foundation-architecture.md
│   └── sportal-foundation-report.md
└── ai_sidecar/                  # Python AI runtime
```

---

## 4. Frontend Architecture

### 4.1 Design Language & Visual Hierarchy
SPortal adopts the **LogisticsHQ Professional SaaS Control Center** aesthetic, faithfully implementing the visual composition established in `sportalDashboard.png`:
- **Deep Navy Sidebar (`#0B192C`)**: High-contrast, focused navigation housing all 14 administrative modules with distinct active indicators (`#1D4ED8`) and bottom branding.
- **Light Off-White Canvas (`#F8FAFC`)**: Clean, distraction-free content viewport with crisp borders (`#E2E8F0`) and subtle shadows.
- **Top Administration Header**: Features portal branding, system search with `Ctrl+K` shortcut, notification alerts, and active administrator identity (`Vaidanshi - Super Admin`).
- **Typography**: Responsive typography utilizing `Outfit` and `Plus Jakarta Sans`.

### 4.2 Application Shell & Module Directory
The application shell establishes routes for the planned 14 SPortal functional areas:
1. **Dashboard** (`/`): Real-time platform KPI overview, revenue metrics, growth charts, health donut, and authoritative recent organizations list.
2. **Organizations** (`/organizations`): Directory of tenant accounts, company profiles, and status controls (Task S3).
3. **Onboarding** (`/onboarding`): Multi-step customer activation pipeline and KYC verification (Task S4).
4. **Subscriptions** (`/subscriptions`): Tier management, feature flags, and custom quotas (Task S5).
5. **Billing & Finance** (`/billing`): MRR/ARR tracking, platform invoices, and payment gateway status (Task S6).
6. **Users & Roles** (`/users`): Cross-tenant user directory and internal staff access controls (Task S7).
7. **Customer 360** (`/customer-360`): Comprehensive single-pane view of customer activity and history (Task S8).
8. **Usage & Analytics** (`/usage`): Metering, API call tracking, and resource overages (Task S9).
9. **Customer Health** (`/customer-health`): Predictive retention scoring and churn early-warning triggers (Task S10).
10. **Integrations** (`/integrations`): Global status of third-party connectors (SES, Twilio, S3, Textract, carriers) (Task S11).
11. **Documents & Compliance** (`/documents`): Document storage metrics, OCR queues, and sanctions screening (Task S12).
12. **Support & Activity** (`/support`): Support ticketing and universal audit log inspection (Task S13).
13. **SPortal AI** (`/ai`): Natural language SaaS intelligence assistant (Task S14).
14. **Settings** (`/settings`): Platform maintenance flags, backup status, and global policies (Task S15).

---

## 5. Backend Integration & Data Model

### 5.1 Shared Authoritative Database
SPortal introduces **zero duplicate tables**. All metrics and administrative actions read and write to authoritative MariaDB 12.3 tables:
- `organizations`: 33 tenant records.
- `organization_subscriptions`: Subscriptions and plan assignments.
- `subscription_plans`: Plan definitions and entitlements.
- `users`: User profiles.
- `org_members`: Organization membership mappings.
- `audit_logs`: 7,915+ universal system events.
- `customer_health_evaluations`: Health scoring.
- `external_integrations`: Connector configurations.

### 5.2 SPortal Backend Endpoints
The backend establishes an explicit SPortal API group in Chi:
- `GET /api/v1/sportal/health`: Public subsystem health probe.
- `GET /api/v1/sportal/meta`: Discovery endpoint returning portal version and module list.
- `GET /api/v1/sportal/overview`: Returns platform aggregates (`total_organizations`, `active_customers`, `total_users`, `active_subscriptions`, `platform_status`).
- `GET /api/v1/sportal/organizations/recent?limit=N`: Returns recent customer organizations with live member counts and plan assignments.

### 5.3 Cross-Origin Resource Sharing (CORS)
Backend CORS middleware was updated in `internal/server/middleware.go` to explicitly allow:
- `http://localhost:5174`
- `http://127.0.0.1:5174`
- `https://sportal.logisticshq.in`

---

## 6. Access Boundary & Security Model

```
                    ┌─────────────────────────┐
                    │     Client Request      │
                    └────────────┬────────────┘
                                 │
                   Bearer JWT / Dev Token
                                 ▼
                    ┌─────────────────────────┐
                    │   authGuard.RequireAuth │
                    └────────────┬────────────┘
                                 │
                        Extract UserContext
                                 ▼
             ┌───────────────────────────────────────┐
             │       Is Internal Staff Role?         │
             │ (SUPER_ADMIN, CEO, SPORTAL_ADMIN)     │
             └───────────┬───────────────┬───────────┘
                         │               │
                     YES │               │ NO
                         ▼               ▼
          ┌─────────────────────┐   ┌───────────────────────────┐
          │ Execute SPortal API │   │ HTTP 403 Forbidden        │
          │ (Cross-Tenant Data) │   │ (Customer Tenant Blocked) │
          └─────────────────────┘   └───────────────────────────┘
```

1. **Authentication Enforcement**: SPortal API endpoints require valid JWT credentials. Unauthenticated requests receive `HTTP 401 Unauthorized`.
2. **Internal Role Gating**: Requests to SPortal endpoints are validated via `IsInternalStaffRole(actorRole)`. Normal freight forwarder users (e.g. `SALES`, `PRICING`, `CUSTOMER_CONTACT`) attempting to access SPortal APIs are rejected with `HTTP 403 Forbidden`.
3. **No Frontend-Only Bypasses**: The backend server is strictly authoritative.
4. **Credential Security**: Zero plaintext API keys or secrets are returned in SPortal metadata.

---

## 7. Future Production URL Architecture

In production, the platform architecture will map as follows:
- `https://logisticshq.in`: Public marketing website.
- `https://app.logisticshq.in`: Customer freight forwarding portal (CPortal).
- `https://sportal.logisticshq.in`: Internal SaaS administration portal (SPortal).
- `https://api.logisticshq.in`: Shared authoritative Go backend API.

Local development maps to:
- `http://localhost:5173`: CPortal
- `http://localhost:5174`: SPortal
- `http://localhost:8080`: Shared Go Backend API
- `127.0.0.1:3306`: Shared MariaDB 12.3 Database

---

## 8. Development & Running Instructions

### Starting the SPortal Frontend
```bash
cd sportal
cmd /c npm run dev
# SPortal will be live at http://localhost:5174/
```

### Building for Production
```bash
cd sportal
cmd /c npm run build
# Generates production bundle in sportal/dist/
```

### Running SPortal Backend Tests
```bash
cd backend
go test -v ./internal/sportal/...
```
