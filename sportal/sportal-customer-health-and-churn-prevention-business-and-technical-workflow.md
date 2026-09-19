# SPortal Customer Health, Success Intelligence & Churn Prevention — Business & Technical Workflow

## 1. Executive Summary & Operational Mission
Task **S11** establishes the comprehensive **Customer Health & Churn Prevention Intelligence System** for **SPortal** (LogisticsHQ SaaS Internal Administration Portal). 

Prior to Task S11, customer retention evaluations relied on disconnected manual reports or subjective manager impressions. Task S11 provides authorized internal LogisticsHQ staff (Super Admins, Customer Success Leads, Logistics Operations Desk) with an authoritative, automated, 7-dimensional customer health scorecard computed directly from persistent, real-time MariaDB transactional records.

### Core Guarantees:
- **Zero Mock Numbers**: All metrics (active users, cargo consignments, exceptions, invoice amounts, AI tasks, and audit logs) derive directly from primary database tables (`org_members`, `shipments`, `shipment_exceptions`, `customer_invoices`, `contracts`, `ai_processing_tasks`, `ai_automations`, `external_integration_configs`, `audit_logs`).
- **Zero Shadow Schemas / DB Duplication**: Aggregations are executed directly within the primary `freel_db` schema without spinning up secondary analytics datastores.
- **Transparent 7-Dimensional Weighting**: Health scores (0–100) are computed across seven clearly weighted vectors with explicit factor breakdowns and confidence scores.
- **Truthful Historical Reporting**: If historical audit records are insufficient for statistical delta computation (e.g. baseline under 60 days), the system truthfully displays `"Insufficient history"` rather than synthesizing fabricated trends.
- **Actionable Grounded Playbooks**: Recommendations cite exact database records (such as specific overdue shipments, unconfigured EDI gateways, or missing transport references).
- **Persistent Internal Customer Success Journal**: LogisticsHQ internal team members can record and review collaborative health checks, QBR notes, and retention interventions directly on customer accounts.

---

## 2. Multi-Dimensional Health Scoring Architecture

The composite health score (0–100) is calculated via a mathematically sound, weighted model across seven operational and commercial dimensions:

| Dimension | Weight | Primary Data Sources | Evaluation Criteria |
| :--- | :---: | :--- | :--- |
| **Commercial & Subscription Health** | **20%** | `organization_subscriptions`, `subscription_plans`, `customer_invoices` | Subscription active status, billing delinquency, days to renewal term, auto-renewal commitment. |
| **Product & Module Adoption** | **20%** | Multi-module telemetry (14 modules), `org_members` | Percentage of core freight forwarding modules actively adopted, active user ratio. |
| **Operational Execution** | **20%** | `shipments`, `shipment_exceptions` | Moving cargo volume, open exceptions, presence of unresolved CRITICAL cargo disruptions. |
| **Team Engagement & Activity** | **15%** | `org_members`, `audit_logs` (30-day window) | Logged user activity, frequency of team operational actions recorded in immutable audit ledger. |
| **AI & Automation Adoption** | **10%** | `ai_processing_tasks`, `ai_automations` | Completed autonomous document extractions, active event automation workflows. |
| **External Connectivity & Gateways** | **5%** | `external_integration_configs`, `carrier_integrations` | Active carrier EDI/API connections, webhook gateways, tracking integrations. |
| **Contract & Compliance Health** | **10%** | `contracts`, `shipment_documents` | Valid master service agreements, contract expiration proximity (30 days), compliance document density. |

### Health State Thresholds:
- **HEALTHY** (Score 85–100): Optimal product engagement, strong team usage, zero critical operational blocks.
- **GOOD** (Score 70–84): Solid platform utilization, active operational movements, minor open disruptions.
- **WATCH** (Score 55–69): Approaching renewal or experiencing noticeable drop in usage or unresolved exceptions.
- **AT_RISK** (Score 40–54): Serious delinquency, high churn probability, low workforce participation.
- **CRITICAL** (Score 0–39): Subscription lapsed or immediate operational crisis.

---

## 3. Go Backend Technical Implementation

### Architecture Overview (`backend/internal/sportal/`)
1. **RBAC & Authorization (`rbac.go`)**:
   - `PermHealthView = "health:view"` enforces that only authorized LogisticsHQ internal personnel can inspect retention scores.
   - `PermHealthManage = "health:manage"` governs creating internal CS notes and executing retention playbooks.
   - Unauthorized customer-tenant users or unauthenticated calls receive HTTP `401 Unauthorized` or `403 Forbidden`.
   - Cross-tenant requests are validated against existing organization IDs, returning `404 Not Found` for invalid IDs.

2. **Data Models (`types.go`)**:
   - `CustomerHealthDetail`: Root aggregate model containing composite score, state, days to renewal, renewal risk, churn probability, dimension list, contributing signals, observed facts, predictive models, recommendations, and CS notes.
   - `HealthDimensionScore`: Structural breakdown per dimension (key, label, score 0-100, weight, status, summary, and key metrics map).
   - `HealthSignalItem`: Grounded observation classified by type (`FACT`, `CALCULATED_SIGNAL`, `PREDICTIVE_SIGNAL`) and impact (`POSITIVE`, `WARNING`, `NEUTRAL`, `CRITICAL`).
   - `PredictiveRiskItem`: Retention and churn models with confidence scores.
   - `CustomerSuccessRecommendation`: Actionable intervention items with grounded evidence and suggested owners.
   - `CustomerNoteItem`: Persistent internal journal entry.

3. **Repository Aggregation (`repository.go`)**:
   - `GetCustomerHealth(ctx, orgID)`: Executes concurrent, non-blocking queries against MariaDB for subscriptions, `org_members`, `shipments`, `shipment_exceptions`, `customer_invoices`, `contracts`, `shipment_documents`, `ai_processing_tasks`, `ai_automations`, `external_integration_configs`, `audit_logs`, and `sportal_customer_notes`.
   - `CreateCustomerNote(ctx, orgID, authorID, authorName, noteType, content)`: Inserts persistent notes into `sportal_customer_notes` and writes an audit log entry.
   - `GetCustomerNotes(ctx, orgID)`: Retrieves chronological history of internal CS notes.

4. **HTTP Endpoints (`handler.go` & `server.go`)**:
   - `GET /api/v1/sportal/organizations/{id}/health`: Detailed health scorecard for customer account.
   - `GET /api/v1/sportal/customer-health`: Platform-wide or default anchor customer health view.
   - `POST /api/v1/sportal/organizations/{id}/health/notes`: Submits internal customer note.
   - `GET /api/v1/sportal/organizations/{id}/health/notes`: Retrieves list of customer notes.

---

## 4. Frontend User Experience (`sportal/src/`)

### Architecture Overview
1. **Dedicated Customer Health Page (`/customer-health` & `/organizations/:organizationId/health`)**:
   - Organization Switcher dropdown allowing instant traversal between customer accounts.
   - Breadcrumb navigation (`SPortal > Customer Health & Retention Intelligence`).
   - Master Health Header with large score badge (e.g. `87/100`), status pill (`HEALTHY`), renewal countdown, churn probability, and quick action buttons.

2. **Customer 360 Workspace Integration**:
   - Seamlessly integrated as a dedicated tab: `Health & Retention` inside `OrganizationDetailPage.jsx` (`/organizations/:id`).
   - Displays identical real-time telemetry alongside Company Profile, Usage, Shipments, Contracts, and Documents.

3. **Signals, Facts & Predictions Matrix**:
   - Interactive sub-tab selector (`All`, `Observed Facts`, `Contributing Signals`, `Predictions`).
   - Authoritative facts verified from MariaDB (e.g. "Workforce Activity: 5 of 5 provisioned seats active", "Cargo Consignments: 4 active shipments").

4. **Actionable Recommendations Engine**:
   - Highlighting grounded evidence from real shipment records (e.g. Overdue ETA on shipment BK-2026-ORG1-001).
   - Priority badges (`CRITICAL`, `HIGH`, `MEDIUM`).
   - Clear assignment to CS Lead or Operations Desk.

5. **Customer Success Notes Journal & Modal**:
   - Clean interactive modal for authorized staff to log notes classified by `HEALTH_CHECK`, `BUSINESS_REVIEW`, `RISK_MITIGATION`, `ONBOARDING`, or `GENERAL`.
   - Responsive grid of notes displaying author, timestamp, classification badge, and text content.

6. **Cross-Module Deep Links**:
   - Direct shortcuts to `/organizations/:id` (Customer 360), `/organizations/:id/usage` (Usage & Quotas), `/organizations/:id/users` (User Directory), and `/subscriptions` (Subscription Management).

---

## 5. Security, Multi-Tenancy & Zero-Trust Governance
- **Strict Tenant Separation**: All MariaDB queries scope by `WHERE org_id = ?` or `WHERE om.org_id = ?`.
- **Internal Role Guard**: Customer tenant staff attempting to hit `/api/v1/sportal/*` endpoints are rejected with `401 Unauthorized` / `403 Forbidden`.
- **Audit Trail**: Every customer note logged triggers an immutable audit log entry in `audit_logs` under module `CUSTOMERS` and action `CUSTOMER_NOTE_CREATED`.
