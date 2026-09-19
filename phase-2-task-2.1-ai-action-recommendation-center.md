# Phase 2 — Task 2.1: AI Action and Recommendation Center Implementation Report

## 1. Executive Summary
Phase 2 — Task 2.1 introduces the centralized **AI Action and Recommendation Center** to LogisticsHQ. The system provides operational operators, logistics coordinators, and executives with a single, unified, auditable command surface to discover, review, prioritize, assign, and track grounded AI recommendations.

Crucially, in accordance with the Phase 2 safety guidelines, the recommendation center operates in a **strictly read-only initial scope** regarding business records. AI generates recommendations and identifies high-risk workflows, but cannot autonomously mutate business entities (such as dispatching emails, altering shipment milestones, mutating invoices, or issuing bookings) without human review and explicit approval.

---

## 2. Existing Architecture Reused
The implementation directly builds on and extends existing LogisticsHQ production abstractions without duplicating subsystems:
- **MariaDB Database**: Applied migration `089_ai_recommendations.sql` using standard MariaDB DDL with idempotent index generation.
- **Tenant Isolation & Auth**: Context extracted via `middleware.GetUserContext(c)` enforcing `org_id` and operator scoping.
- **Centralized Action System**: Registered new read actions (`recommendations.list`, `recommendations.get`) under `backend/internal/actions/` conforming to `Action`, `ActionHandler`, and `ActionMetadata`.
- **Audit System**: Integrated directly with `backend/internal/audit/` (`audit.Service`), ensuring every status transition, review, assignment, and dismissal is persistently audited.
- **Approval / HITL System**: Integrated with existing approval structures; recommendations requiring approval cannot be marked completed directly by users or workers without passing the safety gate.
- **Frontend Design System**: Reused 100% white/light theme, typography, badges, modals, drawers, and Lucide icons matching existing LogisticsHQ modules.

---

## 3. Files Changed
### Backend
- [`backend/internal/database/migrations/089_ai_recommendations.sql`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/database/migrations/089_ai_recommendations.sql): DDL for table `ai_recommendations`.
- [`backend/internal/recommendations/model.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/recommendations/model.go): Data domain model, evidence item structs, allowed status transitions map, and KPI statistics structs.
- [`backend/internal/recommendations/repository.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/recommendations/repository.go): SQL queries for listing, filtering, pagination, stats aggregation, CRUD, evidence queries, and deduplication.
- [`backend/internal/recommendations/generator.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/recommendations/generator.go): Deterministic recommendation generator evaluating 7 database rules across invoices, shipments, RFQs, quotations, contracts, customers, and leads.
- [`backend/internal/recommendations/service.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/recommendations/service.go): Business logic enforcing valid status transitions, approval safety gate, mandatory dismissal reasons, and universal audit logging.
- [`backend/internal/recommendations/handler.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/recommendations/handler.go): HTTP handler layer providing RESTful endpoints.
- [`backend/internal/recommendations/recommendations_test.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/recommendations/recommendations_test.go): Comprehensive test suite covering handler security, transitions, deduplication, lifecycle, approval gate, tenant isolation, and read-only safety.
- [`backend/internal/actions/recommendation_actions.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/actions/recommendation_actions.go): Centralized action registry handlers for recommendation queries.
- [`backend/internal/server/server.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/server/server.go) & [`backend/internal/server/routes.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/server/routes.go): Server wire-up and route registration under `/api/v1/recommendations`.
- [`backend/cmd/server/main.go`](file:///c:/Users/Sai/go/src/freel-project/backend/cmd/server/main.go): Dependency injection and centralized action registration.

### Frontend
- [`frontend/src/services/recommendationService.js`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/services/recommendationService.js): Client service wrapping HTTP endpoints.
- [`frontend/src/pages/dashboard/Recommendations/RecommendationCenterPage.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/Recommendations/RecommendationCenterPage.jsx): Centralized management dashboard with KPI cards, multi-facet filtering, sorting, pagination, review drawers, assign/dismiss modals, and trigger generation.
- [`frontend/src/pages/dashboard/Recommendations/RecommendationCenterPage.css`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/Recommendations/RecommendationCenterPage.css): Light theme styling strictly conforming to LogisticsHQ visual guidelines.
- [`frontend/src/components/recommendations/ModuleRecommendationsWidget.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/components/recommendations/ModuleRecommendationsWidget.jsx): Reusable contextual widget for entity detail pages.
- [`frontend/src/components/dashboard/DashboardRecommendationsWidget.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/components/dashboard/DashboardRecommendationsWidget.jsx): Executive dashboard urgent recommendation overview widget.
- [`frontend/src/components/layout/Sidebar.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/components/layout/Sidebar.jsx): Navigation link for "Recommendations" under the AI Operations category.
- [`frontend/src/pages/dashboard/Customers/CustomerDetailsPage.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/Customers/CustomerDetailsPage.jsx): Embedded module recommendations widget.
- [`frontend/src/pages/dashboard/Shipments/ShipmentDetail.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/Shipments/ShipmentDetail.jsx): Embedded module recommendations widget.
- [`frontend/src/pages/dashboard/Home/OperationalDashboard.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/Home/OperationalDashboard.jsx): Embedded dashboard widget.
- [`frontend/src/App.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/App.jsx): Route `/dashboard/recommendations` registered.
- [`frontend/src/__tests__/pages/Recommendations/RecommendationCenterPage.test.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/__tests__/pages/Recommendations/RecommendationCenterPage.test.jsx): Frontend unit tests for recommendation center workflows.
- [`frontend/src/__tests__/components/RecommendationsWidget.test.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/__tests__/components/RecommendationsWidget.test.jsx): Frontend tests for contextual and dashboard widgets.

---

## 4. Database Migrations
Migration `089_ai_recommendations.sql` created the `ai_recommendations` table with the following schema:
```sql
CREATE TABLE IF NOT EXISTS ai_recommendations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    source_type VARCHAR(64) NOT NULL,
    source_id BIGINT NOT NULL,
    source_reference VARCHAR(128) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(64) NOT NULL,
    priority VARCHAR(32) NOT NULL DEFAULT 'medium',
    risk_level VARCHAR(32) NOT NULL DEFAULT 'medium',
    confidence VARCHAR(32) NOT NULL DEFAULT 'HIGH',
    confidence_score DECIMAL(5,2) NOT NULL DEFAULT 0.85,
    evidence LONGTEXT NOT NULL,
    recommended_action TEXT NOT NULL,
    action_type VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'new',
    assignee_id BIGINT NULL,
    assignee_name VARCHAR(128) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    reviewed_at DATETIME NULL,
    reviewed_by_id BIGINT NULL,
    completed_at DATETIME NULL,
    completed_by_id BIGINT NULL,
    dismissed_at DATETIME NULL,
    dismissed_by_id BIGINT NULL,
    dismissed_reason TEXT NULL,
    freshness DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NULL,
    requires_approval BOOLEAN NOT NULL DEFAULT FALSE,
    approval_id BIGINT NULL,
    correlation_id VARCHAR(128) NOT NULL,
    created_by VARCHAR(128) NOT NULL DEFAULT 'SYSTEM',
    generated_by VARCHAR(128) NOT NULL DEFAULT 'DETERMINISTIC_RULES',
    rule_applied VARCHAR(128) NOT NULL,
    dedup_hash VARCHAR(64) NOT NULL,
    metadata LONGTEXT NULL,
    INDEX idx_ai_rec_org_status (org_id, status),
    INDEX idx_ai_rec_org_priority (org_id, priority),
    INDEX idx_ai_rec_source (org_id, source_type, source_id),
    INDEX idx_ai_rec_dedup (org_id, dedup_hash),
    INDEX idx_ai_rec_approval (org_id, requires_approval, approval_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

---

## 5. Data Model
Key conceptual properties supported by `ai_recommendations`:
- `id`, `org_id`: Tenant identifier guaranteeing isolation.
- `source_type`, `source_id`, `source_reference`: Operational entity linkage (e.g., `SHIPMENT`, `INVOICE`, `RFQ`, `CUSTOMER`).
- `title`, `description`, `category`, `priority`, `risk_level`: Clear business categorization and severity classification.
- `confidence`, `confidence_score`: Explicit calibration (e.g., 0.85–0.95).
- `evidence`: Serialized JSON array of structured observations (`source_module`, `source_entity_id`, `source_ref`, `field_name`, `observed_value`, `description`).
- `recommended_action`, `action_type`: Descriptive next steps without auto-execution.
- `status`: Lifecycle state (`new`, `reviewed`, `assigned`, `approved`, `rejected`, `dismissed`, `completed`, `failed`).
- `requires_approval`, `approval_id`: Guardrail preventing completion of high-risk actions without explicit authorization.
- `dedup_hash`: Cryptographic SHA-256 fingerprint for idempotency.

---

## 6. APIs Added or Changed
All endpoints require JWT authentication and derive `org_id` strictly from the verified session:
- `GET /api/v1/recommendations`: List recommendations with multi-parameter filtering (`category`, `priority`, `risk_level`, `status`, `source_type`, `requires_approval`, `assignee_id`, `search`), sorting, and pagination.
- `GET /api/v1/recommendations/stats`: Summary counts of total, new, reviewed, assigned, critical, and approval-requiring recommendations.
- `GET /api/v1/recommendations/:id`: Retrieve full recommendation details.
- `POST /api/v1/recommendations/generate`: Trigger deterministic rule evaluation and recommendation generation cycle.
- `PATCH /api/v1/recommendations/:id/status`: Update recommendation status according to allowed transition state machine.
- `POST /api/v1/recommendations/:id/assign`: Assign recommendation to an operator (transitions `new`/`reviewed` to `assigned`).
- `POST /api/v1/recommendations/:id/dismiss`: Dismiss recommendation with mandatory reason string.
- `POST /api/v1/recommendations/:id/review`: Mark recommendation as reviewed (transitions `new` to `reviewed`).
- `GET /api/v1/recommendations/:id/evidence`: Return parsed factual evidence items.
- `GET /api/v1/recommendations/source/:type/:id`: Retrieve active recommendations linked to a specific operational record.

---

## 7. Recommendation Generation Rules
The recommendation engine deterministically inspects existing database tables:
1. `OVERDUE_INVOICE_COLLECTIONS`: Flags invoices past due date with outstanding balances > 0.
2. `UNRESOLVED_SHIPMENT_EXCEPTIONS`: Evaluates shipments with active unresolved exceptions (customs holds, weather disruptions, port congestion); high severity triggers `requires_approval = true`.
3. `RFQ_AWAITING_PRICING`: Identifies RFQs in draft or submitted stage without carrier quotes.
4. `QUOTATION_EXPIRING_SOON`: Alerts for accepted/sent quotes expiring within 3 days to expedite booking.
5. `CONTRACT_EXPIRY_OR_RENEWAL`: Detects active contracts with < 30 days remaining before expiration.
6. `HIGH_RISK_CUSTOMER_CREDIT`: Highlights customers with overdue receivables or multiple operational holds.
7. `LEAD_UNATTENDED`: Detects leads in `new` stage with no outreach recorded for > 5 days.

---

## 8. Deduplication Strategy
To guarantee idempotency across page reloads, worker retries, and server restarts, each recommendation computes a stable SHA-256 hash:
```
dedup_hash = SHA256(org_id + ":" + source_type + ":" + source_id + ":" + rule_applied + ":" + active_issue_identifier)
```
When generating recommendations:
- The database is checked for an existing record with matching `(org_id, dedup_hash)`.
- If found and still open (`new`, `reviewed`, `assigned`), the existing record's `freshness`, `evidence`, and `updated_at` are refreshed in-place without generating a duplicate row.
- If previously resolved or dismissed, no duplicate is created while the condition remains unchanged.

---

## 9. Status Transition Rules
Valid transitions are strictly enforced in `service.go`:
- `new` → `reviewed`, `assigned`, `dismissed`
- `reviewed` → `assigned`, `dismissed`
- `assigned` → `approved`, `rejected`, `dismissed`, `reviewed`
- `approved` → `completed`, `failed`
- `rejected` → `dismissed`
- `dismissed` → (terminal state, requires mandatory reason)
- `completed` → (terminal state)

### High-Risk Approval Gate
If `requires_approval == true` or `risk_level == "critical"`, attempting to transition directly to `completed` returns an error (`http.StatusConflict`):
> `"recommendation requires formal approval before completion: risk level critical"`

---

## 10. Permission and Tenant-Isolation Controls
- `org_id` is extracted exclusively from `middleware.GetUserContext(c)` and never accepted from request parameters or body payloads.
- Repository queries enforce `WHERE org_id = ? AND id = ?`. Cross-tenant attempts return `404 Not Found`.
- Source record queries verify ownership of underlying records before attaching evidence.
- Tested specifically in `TestTenantIsolationAndCrossTenantAccess`.

---

## 11. Audit and Observability Changes
Every status update, assignment, review, and dismissal writes an immutable entry to `audit_logs`:
- `action`: `ai_recommendation.status_change`, `ai_recommendation.assigned`, `ai_recommendation.dismissed`, `ai_recommendation.reviewed`
- `actor_id`, `actor_name`: Authenticated operator details
- `entity_type`: `ai_recommendation`
- `entity_id`: Recommendation ID
- `metadata`: Old status, new status, assignee, dismissal reason, correlation ID.

---

## 12. Worker Behavior
The recommendation generator is designed for worker and cron job execution:
- Bounded batch queries preventing database connection exhaustion.
- Fully idempotent: multiple runs with identical conditions do not create duplicate rows.
- Failures in single rule evaluations are logged to observability channels without crashing the generation cycle.

---

## 13. UI Changes
- **Recommendation Center (`/dashboard/recommendations`)**:
  - Full-featured command center with 6 KPI summary stat cards (Total Active, Unreviewed, Assigned, Critical Risk, Approval Required, High Priority).
  - Search input with category, priority, risk, status, and approval filters.
  - Sorting controls (priority, risk, confidence, created date).
  - Detail drawer showing full title, description, rule applied, correlation ID, recommended action, and collapsible evidence table.
  - Operator assignment modal with input validation.
  - Dismissal modal requiring a valid explanation before submission.
- **Module Recommendations Widget**:
  - Clean, compact card embedded in `CustomerDetailsPage.jsx` and `ShipmentDetail.jsx` showing active recommendations linked directly to the open entity.
- **Dashboard Widget**:
  - Embedded in `OperationalDashboard.jsx` highlighting top urgent recommendations.
- **Sidebar Integration**:
  - Added "Recommendations" item with `Compass` icon under AI Operations.

---

## 14. Theme Preservation: Original White/Light Theme
The Recommendation Center strictly adheres to the LogisticsHQ white/light theme:
- Background: `#ffffff` for cards and drawers, `#f8fafc` for page canvas.
- Typography: `#0f172a` primary text, `#475569` secondary text, `#64748b` muted labels.
- Borders: `#e2e8f0` crisp hairline borders with `8px` and `12px` border radii.
- Badges: Semantic pastel badge backgrounds (`#fef2f2` for critical, `#fffbeb` for high, `#eff6ff` for info).

---

## 15. Confirmation: Zero Black or Dark AI Panels
- **No dark panels**: Confirmed 0 dark containers or black overlays.
- **No dark gradients**: Pure solid white and slate surfaces.
- **No glowing neon**: All indicators use standard LogisticsHQ CSS design tokens.

---

## 16. Confirmation: No Fake Data or Database Reset
- **No demo seeds**: All recommendations generated from real operational rows in `freel_mysql`.
- **No database reset / truncation**: Existing customer, shipment, invoice, RFQ, quotation, and contract rows preserved intact.
- **Persistence verified**: Recommendations survive backend restarts, frontend reloads, and container recycles.

---

## 17. Confirmation: Read-Only Initial Scope
- Recommended actions (e.g., "Contact customer accounts payable", "Escalate customs hold") are displayed as actionable guidance for human operators.
- The system executes zero automated email dispatches, status mutations, price changes, or invoice adjustments.
- High-risk items explicitly trigger approval requirements.

---

## 18. Tests Executed
### Backend Tests (Go)
- `internal/recommendations/...`:
  - `TestHandlerSecurityAndWorkflows/Unauthenticated_Returns401` — **PASS**
  - `TestHandlerSecurityAndWorkflows/Generate_Authenticated` — **PASS**
  - `TestHandlerSecurityAndWorkflows/CRUD_Workflow_Endpoints` — **PASS**
  - `TestRecommendationStatusTransitions` — **PASS**
  - `TestDeduplicationHashStability` — **PASS**
  - `TestRecommendationLifecycleAndIdempotency` — **PASS**
  - `TestHighRiskApprovalSafetyGate` — **PASS**
  - `TestTenantIsolationAndCrossTenantAccess` — **PASS**
  - `TestReadOnlySafetyUnderlyingTablesUnchanged` — **PASS**
- `internal/actions/...`:
  - 13 actions tests — **ALL PASS**

### Python Tests (FastAPI / LangGraph Sidecar)
- `ai_sidecar/tests/`:
  - 24 tests passed, 1 skipped (0 failures).

### Frontend Tests (Vitest)
- `RecommendationCenterPage.test.jsx`: 6/6 tests passed.
- `RecommendationsWidget.test.jsx`: 3/3 tests passed.
- Full frontend test suite: **37 test files passed, 208/208 tests passed**.
- Production build: `vite build` completed with 0 errors (`dist/index.html` generated).

---

## 19. Browser & System QA Results
- Dev server running at `http://localhost:5173/` (HTTP 200).
- Go backend running at `http://localhost:8080/` (HTTP 200).
- MariaDB running on port 3306 with verified persistent records.
- Verified real MariaDB rows in `ai_recommendations` generated deterministically:
  - `BK-2026-DEV-001 Exception: Port Congestion Warning`
  - `BK-2026-DEV-001 Exception: Weather Disruption` (requires approval)
  - `BK-2026-DEV-003 Exception: Customs Hold: Discrepancy in HS Code declarations` (requires approval)
  - `RFQ-2026-DEV-002 Awaiting Commercial Pricing`
- Evidence columns store structured, verifiable operational context.

---

## 20. Known Limitations
- Automatic execution of recommended actions is intentionally disabled in this task; operators review and execute through approved workflows.
- Asynchronous AI sidecar summarization is available through context tools; offline deterministic rule generation serves as the primary high-reliability generation pipeline.

---

## 21. Final Acceptance Status
**ACCEPTED AND COMPLETE**
- All 15 core recommendation outcomes implemented.
- Pure LogisticsHQ white/light theme preserved.
- Zero fake data introduced.
- Strict read-only initial safety verified.
- All Go, Python, and frontend tests pass. Production build succeeds.
