# SPortal Executive Dashboard — Business & Technical Architecture Workflow

## 1. Executive & Business Overview

### 1.1 Purpose of the SPortal Primary Dashboard
The **LogisticsHQ SPortal Executive Dashboard** serves as the command center and primary internal landing page for LogisticsHQ operations, executive leadership, customer success, finance, and engineering teams.

Unlike the customer-facing **CPortal** (where individual tenant organizations manage their shipments, documents, and billing), the **SPortal** is an internal SaaS control plane. Its singular objective is to give LogisticsHQ staff total portfolio visibility and operational command over every customer organization, active subscription, onboarding stage, platform subsystem, and customer-health signal across the entire platform.

---

### 1.2 Dashboard Information Hierarchy & Sections

The dashboard layout adheres to an 8-column / 4-column master-detail grid with light theme styling (`#F8FAFC` slate canvas, `#FFFFFF` rounded card panels, and `#0B192C` navy sidebar navigation) matching the authoritative visual reference.

```
+---------------------------------------------------------------------------------------------------+
|  HEADER: Brand Logo | Dashboard Search | Live Platform Status Indicator | Internal Staff Profile  |
+---------------------------------------------------------------------------------------------------+
|  WELCOME ROW: Greeting & Date | Operational Quote | Live Timezone (Local)                        |
+---------------------------------------------------------------------------------------------------+
|  TOP KPI ROW:                                                                                     |
|  [ Total Customers ]    [ Active Customers ]    [ Customers Onboarding ]    [ Monthly Rec. Rev ]  |
+-----------------------------------------------------------------+---------------------------------+
|  PRIMARY 8-COL WORKSPACE                                        |  SIDEBAR 4-COL COMMAND          |
|  - Customer Portfolio Growth (Real Dual-Bar Metric)             |  - Customer Health Donut        |
|  - Revenue & Commercial Run-Rate (MRR vs ARR Projection)        |    (Healthy / Watch / At Risk)    |
|  - Recent Organizations Registry (Tier, Users, Renew, Health)   |  - Priority Attention Items     |
|  - Operational Command Snapshot (Shipments, Exceptions, Autos)  |    (Real Actionable Tasks)      |
|  - Documents & Compliance Verification Tracker                  |  - Upcoming Subscription Renew  |
|  - AI Workforce Intelligence Banner                             |  - Quick Command Actions        |
+-----------------------------------------------------------------+---------------------------------+
```

#### Section Breakdown & Business Value:
1. **Welcome Row & Top KPIs:**
   - **Total Customers (34):** The complete universe of client organizations registered on LogisticsHQ.
   - **Active Customers (1):** Organizations with verified active status and active production subscriptions.
   - **Onboarding Pipeline (1):** Organizations actively progressing through the structured S5 onboarding stages.
   - **Monthly Recurring Revenue ($1,297 / ARR $15,564):** Authoritative commercial recurring revenue from active subscription plans (Apex Freight Enterprise @ $999/mo, Freel Global Standard @ $199/mo, LogisticsHQ Dev Org @ $99/mo).
2. **Customer Portfolio Growth & Trajectory:**
   - Visualizes month-over-month customer acquisition and retention based on real registration dates from the database.
3. **Revenue Overview:**
   - Displays real commercial trajectory, tracking active monthly recurring cashflow and annual contract run-rates.
4. **Customer Health Matrix (S11):**
   - Portfolio-level distribution of customer stability categorized into **Healthy**, **Watch**, **At Risk**, **Critical**, or **Insufficient Data**.
   - Pulls directly from the automated customer health evaluation engine.
5. **Recent Organizations Registry (S4 / S6 / S9):**
   - Comprehensive customer index displaying company name, primary domain, subscription tier, seat utilization, next renewal date, and health status.
   - Clicking any customer navigates directly into that organization's **Customer 360** cockpit (`/customer-360/:id`).
6. **Priority Attention Items (What Requires Action Now):**
   - Live triage inbox surfacing legitimate customer risks: approaching renewals, billing disputes, open operational exceptions, expiring compliance documents, and failed integrations.
7. **Upcoming Renewals (S8):**
   - Proactive 30/60/90-day renewal watchlist showing company, plan tier, renewal date, and annual value to prevent involuntary churn.
8. **Operational Command Snapshot (Control Tower / TMS):**
   - Portfolio-level count of active freight shipments, open exception escalations, connected API integrations, and autonomous background workflows.
9. **Documents & Compliance Snapshot (S13):**
   - Real-time audit tracking verified customer contracts, tax certificates, power-of-attorney documents, and highlighting expired or expiring filings.
10. **SPortal AI & Workforce Telemetry (S10 / AI Workforce):**
    - Live operational status of internal AI agents (Freight Rerouting Agent, Customs Classifier, Document OCR) and total automated tasks processed.
11. **Platform Health & Subsystem Matrix (Platform Health):**
    - Real-time heartbeat probe verifying all 6 platform tiers: Web Frontend, Go Backend API, MariaDB Relational Database, Python AI Sidecar, Autonomous Worker, and Event Mesh.

---

## 2. Technical Architecture & Data Lineage

### 2.1 Component Architecture & Information Flow

```mermaid
flowchart TD
    subgraph BrowserClient["SPortal Frontend (React 19 / Vite)"]
        UI_Dash["DashboardOverview.jsx"]
        UI_Modal["PlatformHealthModal"]
        UI_Nav["Sidebar / Header Nav"]
    end

    subgraph GoBackend["LogisticsHQ Core Control Plane (Go 1.24)"]
        HTTP_Handler["SPortalHandler.GetOverview"]
        RBAC_Middleware["RequireRole / CheckPermission"]
        Service_Layer["SPortalService.GetPlatformOverview"]
        Repo_Layer["SPortalRepository.GetPlatformOverview"]
        Health_Prober["ProbePlatformHealth"]
    end

    subgraph RelationalDB["MariaDB 12.3 (freel_mysql)"]
        T_Org["organizations"]
        T_Sub["subscriptions & subscription_plans"]
        T_Inv["invoices"]
        T_Ship["shipments & exceptions"]
        T_Doc["documents"]
        T_Int["integrations & webhook_events"]
        T_Health["customer_health_evaluations"]
        T_Agent["ai_workforce_agents & automation_tasks"]
    end

    subgraph Sidecar["Python AI Intelligence (Port 8090)"]
        Py_FastAPI["FastAPI Health & Telemetry"]
    end

    UI_Dash -->|GET /api/v1/sportal/overview with Bearer Token| HTTP_Handler
    HTTP_Handler --> RBAC_Middleware
    RBAC_Middleware --> Service_Layer
    Service_Layer --> Repo_Layer
    Service_Layer --> Health_Prober
    
    Repo_Layer -->|SELECT COUNT, SUM| T_Org
    Repo_Layer -->|SELECT plan, price, renewal| T_Sub
    Repo_Layer -->|SELECT invoice status| T_Inv
    Repo_Layer -->|SELECT shipment status, exceptions| T_Ship
    Repo_Layer -->|SELECT doc status, expires_at| T_Doc
    Repo_Layer -->|SELECT provider, health, dead_letters| T_Int
    Repo_Layer -->|SELECT health_score, status| T_Health
    Repo_Layer -->|SELECT agents, task counts| T_Agent

    Health_Prober -->|Ping 8090 /health| Py_FastAPI
    Health_Prober -->|DB Ping| RelationalDB

    Service_Layer -->|Apply RBAC Masking (Strip Financials if Unauthorized)| HTTP_Handler
    HTTP_Handler -->|JSON Response (PlatformOverview)| UI_Dash
```

---

### 2.2 MariaDB Authoritative Tables & Aggregations

No duplicate analytics tables or artificial caching databases were created. Every metric is computed dynamically from the authoritative tables:

| Dashboard Metric | Authoritative MariaDB Tables | Aggregation Logic |
| :--- | :--- | :--- |
| **Total Organizations** | `organizations` | `COUNT(*) WHERE deleted_at IS NULL` |
| **Active Customers** | `organizations` | `COUNT(*) WHERE status = 'ACTIVE'` |
| **Customers Onboarding** | `organizations` | `COUNT(*) WHERE status IN ('ONBOARDING', 'PENDING_VERIFICATION')` |
| **Monthly Recurring Rev (MRR)** | `subscriptions` s JOIN `subscription_plans` p | `SUM(p.price_monthly) WHERE s.status = 'ACTIVE'` |
| **Annual Run Rate (ARR)** | `subscriptions` s JOIN `subscription_plans` p | Calculated as `MRR * 12` |
| **Customer Health Distribution** | `customer_health_evaluations` | Aggregates latest `health_status` per organization (`HEALTHY`, `WATCH`, `AT_RISK`, `CRITICAL`, or `INSUFFICIENT_DATA`) |
| **Upcoming Renewals** | `subscriptions` s JOIN `organizations` o | `SELECT s.*, o.name WHERE s.current_period_end BETWEEN NOW() AND DATE_ADD(NOW(), INTERVAL 90 DAY)` |
| **Active Shipments & Exceptions** | `shipments`, `shipment_exceptions` | Real-time counts of active freight movement and open exceptions |
| **Compliance Verification** | `documents` | Verified vs Expiring vs Expired document counts |
| **Connected Integrations** | `integrations`, `webhook_events` | Active integration configurations and dead-letter webhook tracking |
| **AI Workforce Telemetry** | `ai_workforce_agents`, `automation_tasks` | Active agent count and lifetime processed automation executions |

---

### 2.3 Security, Authorization & Server-Side RBAC

The SPortal Executive Dashboard aggregates highly sensitive commercial and operational data. Access is governed strictly server-side:

1. **Authentication Boundary:**
   - Requires a valid SPortal internal JWT session issued by `/api/v1/sportal/auth/login`.
   - Customer-tier credentials (e.g. CPortal shipper/carrier accounts) are rejected at the SPortal authentication middleware.
2. **Financial Data Protection (Field-Level RBAC):**
   - The Go service inspects the authenticated actor's permissions (`PermBillingView`).
   - If an internal staff member possesses an operational role (e.g. `CUSTOMER_SUCCESS`, `SUPPORT_AGENT`, or `OPERATIONS_COORDINATOR`) that lacks `billing:view`:
     - `MonthlyRecurringRev` and `AnnualRunRate` are forced to `$0.00`.
     - Invoices count and outstanding amounts are sanitized to `0`.
     - Plan pricing on upcoming renewals is masked (`$0.00`).
   - This ensures sensitive platform financials cannot be inspected by examining browser DevTools or intercepted JSON payloads.
3. **Tenant & Customer Data Isolation:**
   - Although SPortal is a global control plane, aggregate queries do not cross-contaminate tenant operational records.
   - Drill-down links pass strongly-typed, immutable UUIDs (`/customer-360/:id`) checked against MariaDB foreign keys.

---

### 2.4 Graceful Degradation & Partial Failure Resilience

To prevent catastrophic white-screen crashes, the dashboard implements resilient error boundaries:
- **Subsystem Isolation:** If the Python AI sidecar or integration gateway experiences transient latency or unavailability, the Go backend reports that specific subsystem as `DEGRADED` or `UNHEALTHY` while continuing to return all database-backed customer, financial, and operational metrics.
- **Frontend Safe Render:** Visual charts, KPI cards, and data tables employ null-safe default structures. If a subsystem payload is empty or masked, informative empty-state placeholders are displayed instead of breaking the React render tree.
