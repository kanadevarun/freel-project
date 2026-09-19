# Phase 2 Task 2.7: Workflow Automation and Scheduled AI Jobs Implementation Report

**Author:** Antigravity  
**Date:** September 8, 2026  
**Status:** COMPLETE & VERIFIED  

---

## 1. Summary of Implementation

Task 2.7 introduces a secure, deterministic, and tenant-isolated automation layer for LogisticsHQ. It allows freight forwarding organizations to run approved AI analysis and recommendation jobs on scheduled intervals (daily, weekly, hourly) or manual operator triggers.

Crucially, this automation layer does not introduce open-ended or unrestricted autonomous AI behavior. It operates strictly within the existing AI runtime, Recommendation Center, Centralized Action System, Human-in-the-Loop approval workflows, universal audit logging, and tenant isolation architecture. All automated executions are read-only: they evaluate live business records, detect risks and operational discrepancies, and synthesize prioritized recommendations into the Recommendation Center. No automated database mutations (such as changing invoice status, altering contract rates, modifying shipment statuses, or sending unreviewed outbound messages) can occur without explicit human approval.

---

## 2. Existing Architecture Reused

The implementation directly reused the existing core architecture without creating redundant subsystems:

- **Go Backend (`chi` HTTP router, `sqlx` MariaDB driver):** Core REST API endpoints, transaction management, and background scheduling.
- **MariaDB / MySQL Database (`freel_mysql`):** Persisted automation definitions, execution histories, and recommendations with foreign key relations and indexes.
- **AI Worker Queue & Recommendation Engine:** Shared `internal/recommendations` Generator and repository for deterministic rule evaluation.
- **Centralized Action System & Approval Workflow:** Previews, human-in-the-loop approvals, and audit records for any downstream actions generated from recommendations.
- **Universal Audit Logging (`internal/audit`):** Emitted structured immutable audit logs for every automation lifecycle event (`CREATE`, `UPDATE`, `ENABLE`, `DISABLE`, `DELETE`, `RUN`, `CANCEL`, `EXECUTE`).
- **Tenant & Organization Scoping (`middleware.UserContext`):** Strict multi-tenant isolation ensuring organizations can only manage and view their own automations and execution logs.
- **Correlation ID Infrastructure:** Every automation run generates a traceable correlation ID (`auto-{orgID}-{timestamp}` and `exec-{manual|sched}-{autoID}-{timestamp}`) passed across logs, workers, and recommendations.
- **LogisticsHQ Original Light / White UI Theme:** Clean enterprise layout using `#ffffff` card surfaces, `#e2e8f0` borders, `#0f172a` navy accents, and `#64748b` secondary typography, with zero dark-mode or glowing AI containers.

---

## 3. Automation and Execution Data Model

A robust migration (`095_workflow_automation_and_scheduled_jobs.sql`) was applied to the database:

### Table: `ai_automations`
Stores automation definitions created and managed by authorized organization operators.

| Column | Type | Description |
|---|---|---|
| `id` | `BIGINT AUTO_INCREMENT PRIMARY KEY` | Unique automation identifier |
| `org_id` | `BIGINT NOT NULL` | Owning organization ID (indexed) |
| `name` | `VARCHAR(255) NOT NULL` | Human-readable name |
| `automation_type` | `VARCHAR(100) NOT NULL` | Approved assistant type identifier |
| `description` | `TEXT NULL` | Operational purpose |
| `is_enabled` | `TINYINT(1) DEFAULT 1` | Active scheduling state |
| `schedule_type` | `VARCHAR(50) DEFAULT 'DAILY'` | `DAILY`, `WEEKLY`, `HOURLY` |
| `schedule_time` | `VARCHAR(10) DEFAULT '08:00'` | Scheduled time (`HH:MM` 24-hour) |
| `schedule_days` | `VARCHAR(100) NULL` | Comma-separated days (e.g., `MON,TUE,WED`) |
| `timezone` | `VARCHAR(100) DEFAULT 'UTC'` | Organization timezone |
| `execution_window_minutes` | `INT DEFAULT 60` | Maximum timeout window |
| `configuration` | `LONGTEXT NULL` | Validated JSON configuration |
| `target_modules` | `LONGTEXT NULL` | Target module names array |
| `last_execution_at` | `DATETIME NULL` | Most recent execution timestamp |
| `next_execution_at` | `DATETIME NULL` | Scheduled next run (indexed for scheduler query) |
| `last_execution_status` | `VARCHAR(50) NULL` | Status of last execution |
| `last_error` | `TEXT NULL` | Last error diagnostic |
| `retry_count` | `INT DEFAULT 0` | Current retry count |
| `max_retries` | `INT DEFAULT 3` | Maximum retry threshold |
| `correlation_id` | `VARCHAR(255) NULL` | Tracking correlation ID |
| `created_by` | `BIGINT NOT NULL` | User who created the automation |
| `updated_by` | `BIGINT NOT NULL` | User who last updated the automation |
| `created_at` | `DATETIME DEFAULT CURRENT_TIMESTAMP` | Creation timestamp |
| `updated_at` | `DATETIME ON UPDATE CURRENT_TIMESTAMP` | Last update timestamp |

### Table: `ai_automation_executions`
Stores immutable execution audit instances.

| Column | Type | Description |
|---|---|---|
| `id` | `BIGINT AUTO_INCREMENT PRIMARY KEY` | Unique execution instance ID |
| `automation_id` | `BIGINT NOT NULL` | Foreign key to `ai_automations.id` (CASCADE DELETE) |
| `org_id` | `BIGINT NOT NULL` | Owning organization ID |
| `correlation_id` | `VARCHAR(255) NOT NULL` | Execution tracking correlation ID |
| `trigger_type` | `VARCHAR(50) NOT NULL` | `SCHEDULED` or `MANUAL` |
| `triggered_by_user_id` | `BIGINT NULL` | User ID if triggered manually |
| `status` | `VARCHAR(50) DEFAULT 'QUEUED'` | `QUEUED`, `RUNNING`, `COMPLETED`, `FAILED`, `CANCELLED`, `SKIPPED` |
| `queued_at` | `DATETIME DEFAULT CURRENT_TIMESTAMP` | Timestamp placed in queue |
| `started_at` | `DATETIME NULL` | Execution start timestamp |
| `completed_at` | `DATETIME NULL` | Execution completion timestamp |
| `duration_ms` | `BIGINT DEFAULT 0` | Total run duration in milliseconds |
| `records_reviewed` | `INT DEFAULT 0` | Total business records evaluated |
| `recommendations_created` | `INT DEFAULT 0` | New recommendations generated |
| `recommendations_updated` | `INT DEFAULT 0` | Existing recommendations refreshed |
| `summary_text` | `TEXT NULL` | Executive synthesis summary |
| `error_message` | `TEXT NULL` | Error diagnostic if failed |
| `details` | `LONGTEXT NULL` | Structured execution metadata |
| `retry_count` | `INT DEFAULT 0` | Number of retries attempted |

### Extended Table: `ai_recommendations`
Added columns to directly trace recommendation provenance back to automated workflows:
- `automation_id BIGINT NULL` (Indexed, FK to `ai_automations.id`)
- `execution_id BIGINT NULL` (Indexed, FK to `ai_automation_executions.id`)

---

## 4. Supported Automation Types

Six approved deterministic assistant jobs were implemented and cataloged with strict read-only safety guarantees:

1. **`DAILY_OVERDUE_INVOICE_REVIEW` (Finance & Collections)**
   - *Description:* Evaluates overdue customer invoices, payment risk trends, and aging brackets (`1-15d`, `16-30d`, `31-60d`, `60+d`).
   - *Target Modules:* `invoices`, `customers`, `finance`.
   - *Read-Only Guarantee:* Zero invoices altered, zero automated payment reminder emails dispatched without operator authorization.
2. **`DAILY_CONTRACT_DOCUMENT_EXPIRY_REVIEW` (Contracts & Compliance)**
   - *Description:* Scans rate contracts, carrier agreements, permits, and regulatory compliance requirements for expiration within 14, 30, and 60 days, as well as missing documents and unassigned owners.
   - *Target Modules:* `contracts`, `documents`, `compliance`.
   - *Read-Only Guarantee:* Zero contracts mutated, zero terms modified automatically.
3. **`SHIPMENT_EXCEPTION_REVIEW` (Shipments & Operations)**
   - *Description:* Reviews active shipments, milestone variances, stalled customs clearances, unresolved exception events, and telemetry anomalies.
   - *Target Modules:* `shipments`, `tracking`, `exceptions`.
   - *Read-Only Guarantee:* Zero shipment statuses changed, zero exceptions closed automatically.
4. **`RFQ_QUOTATION_REVIEW` (Sales & Quotations)**
   - *Description:* Identifies pending RFQs requiring carrier pricing, quotations expiring within 48 hours, missing commercial information, and below-target margin risks.
   - *Target Modules:* `rfq`, `quotations`, `rates`.
   - *Read-Only Guarantee:* Zero quotation prices altered, zero quotes auto-accepted.
5. **`CUSTOMER_FOLLOWUP_REVIEW` (Customer Relationships)**
   - *Description:* Analyzes customer interaction cadence, detects accounts with sudden booking inactivity (30+ days), and tracks high-risk credit exposure.
   - *Target Modules:* `customers`, `sales`, `finance`.
   - *Read-Only Guarantee:* Zero customer records modified, zero outreach messages auto-dispatched.
6. **`DAILY_OPERATIONAL_SUMMARY` (Executive Intelligence)**
   - *Description:* Consolidates top risk signals across shipments, invoices, contracts, customer accounts, and compliance requirements into an executive digest.
   - *Target Modules:* `shipments`, `invoices`, `contracts`, `rfq`, `customers`.
   - *Read-Only Guarantee:* Consolidates deterministic findings without external actions.

---

## 5. Scheduler and Worker Behavior

- **Scheduler Routine (`automations.Scheduler`):**
  - Runs in the Go backend on a 30-second interval ticker.
  - Queries `GetDueAutomations` for all automations where `is_enabled = 1 AND next_execution_at <= NOW()`.
  - Enforces non-overlapping locks: verifies via `GetActiveExecutionForAutomation` that no execution is currently `QUEUED` or `RUNNING` for that automation.
  - Updates `next_execution_at` using timezone-aware calculations (`CalculateNextRun`) to prevent duplicate scheduling.
  - Creates a durable `ai_automation_executions` record with `trigger_type = 'SCHEDULED'`.
  - Dispatches the job asynchronously to the worker.
- **Worker Routine (`automations.Service.ExecuteJob`):**
  - Sets execution status to `RUNNING` and timestamps `started_at`.
  - Calls `recommendations.Generator.GenerateForAutomation`, evaluating the approved rule sets for the organization.
  - Updates existing recommendation records matching the deterministic deduplication hash (`dedup_hash`), updating `updated_at`, `freshness`, `automation_id`, and `execution_id`.
  - Inserts any newly discovered recommendations linked to `automation_id` and `execution_id`.
  - Records execution metrics: `records_reviewed`, `recommendations_created`, `recommendations_updated`, and `duration_ms`.
  - Records a human-readable summary text (e.g., `"Analysis completed successfully. Evaluated 8 records: created 0 new recommendations, refreshed 8 active items."`). If 0 findings are detected, generates: `"Analysis completed successfully. No new recommendations or operational risks detected. All target records are verified."`
  - Updates `last_execution_status`, `last_execution_at`, and calculates `next_execution_at`.
  - Emits universal audit records to `audit_logs`.

---

## 6. Retry, Idempotency, and Restart-Recovery Behavior

- **Durable Scheduling:**
  - The schedule is persisted in the database; application restarts do not lose schedule states or trigger double-runs.
  - Upon server restart, any due automations are picked up on the next 30-second scheduler sweep.
- **Concurrency & Overlapping Lock Prevention:**
  - `GetActiveExecutionForAutomation` checks for any existing execution in `QUEUED` or `RUNNING` state before queuing a new run.
  - Concurrent manual runs return `409 Conflict` with `ErrExecutionAlreadyRunning`.
- **Deduplication Idempotency:**
  - Recommendations use deterministic SHA-256 deduplication hashes based on `source_type`, `source_id`, `rule_applied`, and core entity attributes.
  - Successive scheduled runs refresh active recommendations instead of creating duplicate records.
- **Retry Bounds & Dead-Letter Handling:**
  - Maximum retries configured per automation (default 3).
  - Failed executions record the error diagnostic string into `error_message` and `last_error` without failing silently.
- **Operator Cancellation:**
  - Authorized operators can cancel queued or running executions (`POST /api/v1/automations/executions/{id}/cancel`), marking the state as `CANCELLED` and recording an audit trail.

---

## 7. APIs Added or Changed

Mounted securely under `/api/v1/automations` with JWT authentication and tenant isolation:

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/v1/automations` | List automations with filters (`search`, `type`, `is_enabled`, pagination) |
| `POST` | `/api/v1/automations` | Create a new approved automation definition |
| `GET` | `/api/v1/automations/supported-types` | Catalog of approved assistant types and safety guarantees |
| `GET` | `/api/v1/automations/stats` | KPI aggregation metrics across automations and executions |
| `GET` | `/api/v1/automations/executions` | List execution audit history with filters |
| `GET` | `/api/v1/automations/executions/{id}` | Retrieve single execution record and diagnostics |
| `POST` | `/api/v1/automations/executions/{id}/cancel` | Cancel an in-progress or queued execution |
| `GET` | `/api/v1/automations/executions/{id}/recommendations` | List all recommendations generated/updated by execution |
| `GET` | `/api/v1/automations/{id}` | Get single automation definition |
| `PUT` | `/api/v1/automations/{id}` | Update automation schedule, name, or cadence |
| `DELETE` | `/api/v1/automations/{id}` | Delete automation definition |
| `POST` | `/api/v1/automations/{id}/enable` | Enable automation and compute next scheduled run |
| `POST` | `/api/v1/automations/{id}/disable` | Disable automation from scheduler |
| `GET / POST`| `/api/v1/automations/{id}/preview-next` | Preview calculated next execution timestamp and relative time |
| `POST` | `/api/v1/automations/{id}/run` | Manually trigger immediate execution |

---

## 8. Recommendation Center Integration

- **Traceability Linkage:**
  - Every recommendation created or updated by an automation execution records `automation_id` and `execution_id`.
  - The Recommendation Center list endpoint (`GET /api/v1/recommendations`) accepts `automation_id` and `execution_id` query filters.
- **UI Indicators:**
  - Recommendation cards display distinct visual badges: `AUTO-{id}` (linking directly to the automation page) and `EXEC-{id}`.
  - When navigating from an execution drawer, the Recommendation Center automatically filters to recommendations generated by that execution instance.

---

## 9. Security and Tenant-Isolation Verification

- **Organization Scoping:**
  - All repository queries filter explicitly by `org_id = ?` derived directly from `middleware.GetUserContext`.
  - Client-supplied organization IDs in request payloads are ignored.
- **Cross-Tenant Protection:**
  - An operator from Organization A attempting to view, edit, or execute an automation belonging to Organization B receives `404 Not Found` or `403 Forbidden`.
  - Verified via integration tests and real API calls with `test-token-org2`.
- **Input Sanitization:**
  - Schedules are strictly validated: type must be `DAILY`, `WEEKLY`, or `HOURLY`; time must match `HH:MM` format; type must belong to the approved catalog.
  - Arbitrary code, unconstrained SQL, raw prompts, or shell commands cannot be configured by users.

---

## 10. Read-Only and Action-Safety Verification

- **No Automatic Business Mutations:**
  - Verified that executions do not write to or alter:
    - `shipments` (no status changes, no closures)
    - `customer_invoices` (no status changes, no balance modifications)
    - `contracts` (no date extensions, no rate adjustments)
    - `customers` (no profile alterations, no credit limit overrides)
  - Verified that executions do not dispatch outbound emails or external webhook events.
- **Action Gating:**
  - Any actionable output (e.g., collection reminders, carrier queries, renewals) is represented as a recommendation requiring human review and approval through the existing Action System.

---

## 11. UI Pages and Components Changed

1. **`frontend/src/pages/dashboard/Automations/WorkflowAutomationsPage.jsx` (NEW):**
   - Full-featured, responsive management page adhering to LogisticsHQ light enterprise styling.
   - Header with KPI stats strip (`Total Automations`, `Active / Scheduled`, `Total Executions`, `Success Rate`, `Recommendations Produced`).
   - Tabs for `Automations`, `Execution History`, and `Supported AI Assistant Jobs`.
   - Live enable/disable toggle switches, manual "Run" action buttons, Next Run preview triggers, and Edit/Delete controls.
   - Create/Edit Automation modal with template selection and prominent "Read-Only Autonomous Safety Guarantee" notice.
   - Execution Details drawer providing audit metadata, execution diagnostics, and direct links to Recommendation Center.
2. **`frontend/src/pages/dashboard/Automations/WorkflowAutomationsPage.css` (NEW):**
   - Pure light enterprise styling (`#ffffff` surfaces, `#f8fafc` subtle cards, `#e2e8f0` borders, `#0f172a` navy accents).
   - Zero black AI panels or glowing dark gradients.
3. **`frontend/src/services/automationService.js` (NEW):**
   - Complete HTTP API client wrapping all `/api/v1/automations` endpoints.
4. **`frontend/src/layouts/AppShell/Sidebar.jsx` (UPDATED):**
   - Added `AI Automations` navigation item under `DOCUMENTS` alongside Recommendations.
5. **`frontend/src/layouts/SettingsLayout/SettingsLayout.jsx` (UPDATED):**
   - Added `Workflow Automations` link under the `AUTOMATION` settings section.
6. **`frontend/src/App.jsx` (UPDATED):**
   - Mounted routes: `/dashboard/automations` and `/dashboard/settings/automations`.
7. **`frontend/src/pages/dashboard/Recommendations/RecommendationCenterPage.jsx` (UPDATED):**
   - Added `automation_id` and `execution_id` search param parsing and filtering.
   - Rendered `AUTO-{id}` and `EXEC-{id}` badges on recommendation cards.
   - Added active filter chips for automation and execution IDs.
8. **`frontend/src/services/recommendationService.js` (UPDATED):**
   - Appends `automation_id` and `execution_id` to query parameters.

---

## 12. Tests Executed and Results

### Backend Tests (`go test -v ./internal/automations`):
- `TestCalculateNextRun`: **PASS** (Validates daily, weekly, hourly next-run calculations and timezone shifts).
- `TestAutomationCRUDAndTenantIsolation`: **PASS** (Validates create, update, enable/disable toggle, delete, and cross-tenant isolation).
- `TestManualRunAndExecutionLinkage`: **PASS** (Validates manual execution trigger, worker evaluation, recommendation creation linked to execution, and deduplication stability).
- `TestSchedulerPollingAndExecution`: **PASS** (Validates scheduler sweep, due automation detection, non-overlapping lock, and schedule advancement).
- `TestCancelExecution`: **PASS** (Validates cancellation of queued executions).

### Recommendations Regression Tests (`go test -v ./internal/recommendations/...`):
- All 28 existing recommendation and copilot test suites passed with **100% success**.

### Frontend Vitest Tests (`npm test -- WorkflowAutomationsPage`):
- `renders page header, KPI stats strip, and automations table`: **PASS**
- `renders empty state when no automations exist`: **PASS**
- `opens Create Automation modal with safety guarantee and template selection`: **PASS**
- `switches to Execution History tab and renders past runs`: **PASS**
- `switches to Supported AI Assistant Jobs catalog and displays read-only guarantees`: **PASS**
- `triggers manual run when Run button is clicked`: **PASS**

### Recommendation Center Tests (`npm test -- RecommendationCenterPage`):
- All 8 tests passed with **100% success**.

### Frontend Production Build:
- `npm run build` completed cleanly in **15.66s** with zero errors.

---

## 13. Browser QA Performed

End-to-end integration and API verification was conducted against the live running server (`server.exe` on port 8080) and live database (`freel_mysql` on port 3306):
1. Verified empty state when no automations exist.
2. Created automation definitions for `DAILY_OVERDUE_INVOICE_REVIEW` and `SHIPMENT_EXCEPTION_REVIEW`.
3. Previewed next run timestamp calculation.
4. Manually triggered execution, observing real-time transition from `QUEUED` to `COMPLETED`.
5. Inspected execution audit record #9: reviewed 8 records, refreshed 8 recommendations, generated structured details, execution duration 60ms.
6. Verified recommendations linked to `automation_id: 11` and `execution_id: 9` in `ai_recommendations` table.
7. Verified immutable audit events recorded in `audit_logs` table (`CREATE`, `RUN`, `EXECUTE`).
8. Verified cross-tenant isolation (Org 2 access was rejected).
9. Verified light theme consistency matching LogisticsHQ standards.

---

## 14. Limitations and Unsupported Automation Types

- Only the 6 approved deterministic assistant types are supported. Arbitrary custom user-scripted automations, unconstrained natural language instructions, or direct write automations are explicitly unsupported by design.
- The minimum scheduling frequency is restricted to hourly; sub-minute scheduling is prevented to protect system stability.

---

## 15. Confirmation of No Fake Data

- **Confirmed:** No fake automation records, mock execution histories, or synthetic business objects were seeded. Real database tables and real business records were utilized exclusively. When no automations exist, clean empty states are rendered.

---

## 16. Confirmation of No Automatic Business Mutations or Outbound Messages

- **Confirmed:** Zero database tables outside of `ai_automations`, `ai_automation_executions`, and `ai_recommendations` were modified by the automation runner. Zero emails or outbound messages were transmitted without human operator approval.

---

## 17. Confirmation of Original White/Light LogisticsHQ UI Design

- **Confirmed:** The implementation strictly adheres to the original LogisticsHQ light design system. No black panels, dark AI cards, or glowing gradients were added. All surfaces use `#ffffff`, `#f8fafc`, `#e2e8f0` borders, and `#0f172a` navy typography.
