# Phase 2 — Task 2.2: Customer Follow-Up Assistant Implementation Report

## 1. Executive Summary
Phase 2 — Task 2.2 implements the **Customer Follow-Up Assistant** for LogisticsHQ freight-forwarding SaaS application. The assistant builds upon the foundation of Phase 2 Task 2.1 (AI Action & Recommendation Center) to help LogisticsHQ commercial and customer operations teams identify customers who require attention, understand why follow-up is recommended using verifiable source data, review attached evidence, and prepare controlled follow-up actions without turning into an autonomous sales agent.

The initial scope is strictly bounded:
- It evaluates business records deterministically using MariaDB data.
- It attaches explicit factual evidence to each follow-up opportunity.
- It allows authorized operators to review, assign, or dismiss recommendations with mandatory audit reasons.
- It provides an **explicit user-initiated draft message generation** workflow where editable drafts are created on demand.
- It guarantees **idempotent internal follow-up task creation** preventing duplicate work items.
- It prohibits automatic external communication (no auto-emails, SMS, or WhatsApp) and prevents unauthorized mutations to business records (shipments, invoices, quotations, contracts, or customer profiles).

---

## 2. Existing Architecture Reused
The Customer Follow-Up Assistant completely reuses and integrates natively with the existing LogisticsHQ systems:
- **MariaDB 12.3 Database**: Reuses `freel_mysql` persistent storage, extending `ai_recommendations` and creating `customer_followup_tasks` with schema foreign keys and unique constraints.
- **Go Chi HTTP Backend**: Reuses router, JWT auth, tenant isolation (`middleware.GetUserContext`), RBAC authorization, and error wrappers.
- **Unified Recommendation Center (Task 2.1)**: Extends recommendation generator, repository, service, and handlers.
- **Centralized Action System (Task 0.5)**: Reuses `actions.Registry` and `actions.Service`, adding `followups.generate_draft` and `followups.create_task`.
- **HITL Approvals System (Task 0.6)**: Follow-ups that touch credit terms or financial exposure enforce existing approval workflow gates.
- **Universal Audit Logging**: Audits all recommendation status transitions, draft generations, draft updates, and task creations via `auditService.Service`.
- **AI Workforce Monitoring & Queue Worker**: Maintains standard correlation ID tracking and queue worker reliability.
- **Original LogisticsHQ White/Light Theme**: Native CSS classes, enterprise layouts, navy sidebar, and zero dark AI panels.

---

## 3. Files Changed
### Database Migrations
- [`backend/internal/database/migrations/090_customer_followups.sql`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/database/migrations/090_customer_followups.sql): Applied repeatable migration adding follow-up and draft fields to `ai_recommendations` and creating `customer_followup_tasks`.

### Backend (Go)
- [`backend/internal/recommendations/model.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/recommendations/model.go): Added Customer Follow-Up types, Draft statuses (`NOT_GENERATED`, `DRAFTED`, `EDITED`), `CustomerFollowupTask`, `CreateFollowupTaskInput`, `SaveDraftInput`, and `FollowupStats`.
- [`backend/internal/recommendations/repository.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/recommendations/repository.go): Extended column projections, added `SaveDraft`, `CreateFollowupTask`, `GetFollowupTaskByRecID`, `ListFollowupTasks`, and `GetFollowupStats`.
- [`backend/internal/recommendations/generator.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/recommendations/generator.go): Integrated customer identity resolution, overdue invoice customer checks (`customer_invoices`), quotation response reminders, and customer inactivity check-ins.
- [`backend/internal/recommendations/service.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/recommendations/service.go): Implemented deterministic draft synthesis based on verified records, draft persistence, task creation idempotency, and audit logging.
- [`backend/internal/recommendations/handler.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/recommendations/handler.go): Added HTTP handlers for draft generation, draft editing/saving, task creation, and stats retrieval.
- [`backend/internal/server/routes.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/server/routes.go): Registered `/api/v1/recommendations/{id}/draft`, `/api/v1/recommendations/{id}/task`, `/api/v1/recommendations/tasks`, and `/api/v1/followups` endpoints.
- [`backend/internal/actions/recommendation_actions.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/actions/recommendation_actions.go): Registered `followups.generate_draft` and `followups.create_task` business actions.
- [`backend/cmd/server/main.go`](file:///c:/Users/Sai/go/src/freel-project/backend/cmd/server/main.go): Initialized and registered follow-up actions into the centralized action registry.
- [`backend/internal/recommendations/recommendations_test.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/recommendations/recommendations_test.go): Added unit and integration tests for follow-up signals, draft generation, and task idempotency.
- [`backend/internal/recommendations/security_test.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/recommendations/security_test.go): Added cross-tenant isolation and read-only safety tests for follow-ups, drafts, and tasks.

### Frontend (React/Vite)
- [`frontend/src/services/recommendationService.js`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/services/recommendationService.js): Added client methods `generateDraft`, `saveDraft`, `createFollowupTask`, `listFollowupTasks`, and `getFollowupStats`.
- [`frontend/src/services/recommendationService.test.js`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/services/recommendationService.test.js): Unit test suite for new service methods.
- [`frontend/src/pages/dashboard/Recommendations/RecommendationCenterPage.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/Recommendations/RecommendationCenterPage.jsx): Added Follow-Up Type dropdown filter, "Draft Message" / "Create Task" controls, Draft Communication modal, and Create Follow-Up Task modal.
- [`frontend/src/components/recommendations/ModuleRecommendationsWidget.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/components/recommendations/ModuleRecommendationsWidget.jsx): Displayed follow-up type badges, owner suggestions, draft indicators, and direct link to Recommendation Center on Customer Details page.
- [`frontend/src/__tests__/pages/Recommendations/RecommendationCenterPage.test.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/__tests__/pages/Recommendations/RecommendationCenterPage.test.jsx): Frontend component test verifying modal interactions and draft generation.

---

## 4. Database Migrations
Migration script: `090_customer_followups.sql`
1. Added columns to `ai_recommendations`:
   - `customer_id BIGINT NULL`
   - `customer_name VARCHAR(255) NULL`
   - `followup_type VARCHAR(64) NULL`
   - `suggested_owner_id BIGINT NULL`
   - `suggested_owner_name VARCHAR(255) NULL`
   - `draft_subject VARCHAR(500) NULL`
   - `draft_body TEXT NULL`
   - `draft_status VARCHAR(32) NOT NULL DEFAULT 'NOT_GENERATED'`
   - `draft_generated_at DATETIME NULL`
   - `followup_task_id BIGINT NULL`
2. Created table `customer_followup_tasks`:
   - Enforces tenant isolation via `org_id`.
   - Guaranteed idempotency via `UNIQUE KEY uq_rec_task (org_id, recommendation_id)`.
   - Tracks `customer_id`, `source_type`, `source_id`, `followup_type`, `title`, `reason`, `suggested_action`, `priority`, `assignee_id`, `due_date`, `status`, and `correlation_id`.

---

## 5. Follow-Up Signal Rules
Signals are strictly deterministic and traceable to real LogisticsHQ records:
1. `LEAD_SLA_BREACH`: Uncontacted inbound lead exceeding SLA thresholds.
2. `UNANSWERED_RFQ`: Open RFQ pending customer quote response > 48h.
3. `QUOTE_APPROACHING_EXPIRY`: Customer quotation valid but expiring within 72h.
4. `DELIVERED_SHIPMENT_CHECKIN`: Delivered shipment where customer check-in follow-up is beneficial.
5. `UNRESOLVED_SHIPMENT_EXCEPTIONS`: Active shipment exceptions impacting customer freight.
6. `OVERDUE_INVOICE_COLLECTIONS`: Invoices with overdue balance approaching 30+ days.
7. `CONTRACT_EXPIRY_OR_RENEWAL`: Commercial customer contracts expiring within 45 days.
8. `CUSTOMER_INACTIVITY_CHECKIN`: Customer with no recorded shipment or RFQ activity in 60 days.

---

## 6. Priority Calculation
Priority calculation is transparent and deterministic:
- **Critical Priority**: High overdue invoices (> $10,000 balance / > 60 days overdue), critical shipment exceptions (customs hold, cargo damage), or high-value contracts expiring within 14 days.
- **High Priority**: Unanswered RFQs > 48h, quotation expiring within 72h, or active transit exceptions.
- **Medium Priority**: General check-in after delivery or customer inactivity > 60 days.
- **Low Priority**: General routine account reviews.

---

## 7. Recommendation and Evidence Structure
Every follow-up recommendation encapsulates:
- `id`, `org_id`, `customer_id`, `customer_name`, `source_type`, `source_id`, `source_reference`
- `title`, `description`, `category` (`customer` or `sales`)
- `priority`, `confidence`, `confidence_score`
- `evidence`: Array of `{ source_module, source_entity_id, source_ref, field_name, observed_value, description }`
- `followup_type`: One of 9 standardized categories
- `suggested_owner_id`, `suggested_owner_name`: Account manager or operations assigned to the customer
- `recommended_action`: Specific next business step
- `status`: Lifecycle state (`new`, `reviewed`, `assigned`, `approved`, `completed`, `dismissed`)
- `draft_status`: `NOT_GENERATED`, `DRAFTED`, or `EDITED`
- `followup_task_id`: Linked task if created

---

## 8. Deduplication Strategy
- Generates a stable SHA-256 hash using `(org_id, source_type, source_id, category, rule_applied)`.
- If an existing recommendation exists in `new`, `reviewed`, or `assigned` status, the generator refreshes its freshness timestamp and updates evidence rather than inserting duplicate records.
- Refreshing the page or running repeated worker cycles produces zero duplicate recommendations.

---

## 9. Freshness Behavior
- `freshness` records the exact UTC timestamp when underlying records were last evaluated.
- Cards and drawers display freshness relative to current system time.
- If source issues are resolved, status transitions to `completed`.

---

## 10. Draft Message Behavior
- **Explicit User Action Only**: Drafts are NEVER pre-generated during automatic background cycles.
- Initiated via `POST /api/v1/recommendations/{id}/draft`.
- Grounded strictly in verified records (Customer Name, Contact Name, RFQ/Quote/Shipment reference, actual dates, and factual statuses).
- Stored as editable text in `draft_subject` and `draft_body`.
- User edits are persisted via `PATCH /api/v1/recommendations/{id}/draft`.
- **Zero Automatic Sending**: The assistant cannot dispatch messages automatically. External communication requires manual user transmission or the existing approval gate.

---

## 11. Follow-Up Task Behavior
- Initiated via `POST /api/v1/recommendations/{id}/task`.
- Creates a persisted record in `customer_followup_tasks`.
- Idempotency is enforced at both service and database levels (`UNIQUE KEY uq_rec_task (org_id, recommendation_id)`). Repeated requests return the existing task without duplicate inserts.
- Task status and ID are linked to the recommendation and displayed in the UI.

---

## 12. API Changes
The following authenticated endpoints were added/extended:
- `POST /api/v1/recommendations/{id}/draft`: Generate verified draft message.
- `PATCH /api/v1/recommendations/{id}/draft`: Edit and save draft message.
- `POST /api/v1/recommendations/{id}/task`: Create internal follow-up task.
- `GET /api/v1/recommendations/tasks`: List internal follow-up tasks.
- `GET /api/v1/recommendations/followups/stats`: Get follow-up pipeline metrics.
- `GET /api/v1/followups/*`: Dedicated follow-up route alias.

---

## 13. Permission and Tenant-Isolation Controls
- Context isolation enforced via `userCtx.OrgID`.
- Direct ID manipulation or cross-tenant lookups (`orgA` trying to access `orgB`'s recommendation or task) return 404/denial.
- Actions enforce RBAC permissions (`COMPANIES:READ` for drafts, `COMPANIES:UPDATE` for tasks).

---

## 14. Audit and Observability Changes
Universal audit events are logged via `auditService.Service`:
- `CUSTOMER_FOLLOWUP_DRAFT_GENERATED`
- `CUSTOMER_FOLLOWUP_DRAFT_SAVED`
- `CUSTOMER_FOLLOWUP_TASK_CREATED`
- `RECOMMENDATION_STATUS_UPDATED`
- `RECOMMENDATION_ASSIGNED`
- `RECOMMENDATION_DISMISSED`
Includes `org_id`, `actor_id`, `actor_name`, `resource_id`, `correlation_id`, `timestamp`, and `status`.

---

## 15. Worker Behavior
- Integrates into existing task queue without competing workers.
- Generation is bounded, idempotent, and non-blocking.
- Survives daemon and sidecar restarts.

---

## 16. UI Changes
- **Recommendation Center**:
  - Added Follow-Up Type dropdown filter with 9 standard types.
  - Added "Draft Message" / "Edit Draft" and "Create Task" buttons to relevant cards.
  - Integrated Draft Communication modal with verified context, editable subject, and body.
  - Integrated Create Follow-Up Task modal with Assignee, Priority, Due Date, and Notes.
- **Customer Details Page**:
  - Extended `ModuleRecommendationsWidget` to display customer follow-up type badges, suggested owner, draft indicators, and direct links to the Recommendation Center.

---

## 17. Confirmation of Design & Theme Compliance
- **100% White/Light LogisticsHQ Theme**: Preserved white card backgrounds, light slate borders (`#e2e8f0`), neutral typography (`#0f172a`, `#475569`), navy sidebar, and light inputs.
- **Zero Black/Dark AI Panels**: No black backgrounds, dark modals, glowing borders, or neon accents introduced.

---

## 18. Confirmation of Data Safety
- **No Fake Data**: Real persisted MariaDB data used exclusively.
- **No Database Reset or Record Mutation**: Existing shipments, invoices, quotations, contracts, and customer master records were not modified.
- **Read-Only Intelligence**: Evaluated deterministically without side effects on core business tables.

---

## 19. Confirmation of Communication Safety
- **Zero Automatic Communication**: The assistant does not send emails, SMS, or WhatsApp messages.
- External dispatch remains strictly under operator control.

---

## 20. Tests Executed
### Backend Tests (Go)
- `TestHandlerSecurityAndWorkflows`: PASS
- `TestRecommendationStatusTransitions`: PASS
- `TestDeduplicationHashStability`: PASS
- `TestRecommendationLifecycleAndIdempotency`: PASS
- `TestHighRiskApprovalSafetyGate`: PASS
- `TestCustomerFollowupDraftAndTaskIdempotency`: PASS
- `TestTenantIsolationAndCrossTenantAccess`: PASS
- `TestReadOnlySafetyUnderlyingTablesUnchanged`: PASS
**Result**: `ok github.com/freel/backend/internal/recommendations 2.355s`

### Python AI Sidecar Tests
- `pytest test_phase02_production_readiness.py`: 10 passed in 70.80s (100% PASS)

### Frontend Tests (Vitest)
- `src/services/recommendationService.test.js`: 5 passed
- `src/__tests__/pages/Recommendations/RecommendationCenterPage.test.jsx`: 7 passed
- Full suite execution: 172 tests passed across 30 test files

### Production Build
- `npm run build`: Vite v8.0.12 production build passed cleanly (`✓ built in 23.59s`).

---

## 21. Known Limitations
- Draft messages are generated using template heuristics and deterministic context; complex multi-language translations will be enhanced in future language model expansions.
- Communication dispatch is intentionally human-operated to satisfy enterprise HITL compliance.

---

## 22. Final Acceptance Status
**Status: COMPLETE AND ACCEPTED.**
All requirements of Phase 2 — Task 2.2 have been implemented, tested, verified, and integrated into LogisticsHQ.
