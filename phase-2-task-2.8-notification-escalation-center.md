# Phase 2 Task 2.8: Notification and Escalation Center Implementation Report

**Status:** Completed & Validated  
**Module:** Notification and Escalation Center  
**Environment:** Go Backend (Port 8080), MariaDB (Port 3306), React/Vite Frontend (Port 5173), Python/FastAPI AI Sidecar (Port 8090)  

---

## 1. Summary of Implementation

Phase 2 Task 2.8 establishes a centralized, enterprise-grade in-app Notification and Escalation Center for LogisticsHQ. The system enables users to discover high-priority AI recommendations, operational shipment exceptions, pending human-in-the-loop approvals, overdue invoice risks, upcoming document expirations, and automation failures without creating duplicate notification systems, parallel alert infrastructure, or uncontrolled outbound messaging.

Key capabilities delivered:
- **Real Database Persistence:** Replaced in-memory mock structures with real database tables (`notifications`, `notification_escalation_events`, `user_notification_preferences`) applied via migration `096_notifications_and_escalation_center.sql`.
- **Deterministic Evaluation & Synchronization:** Evaluates active business records (`ai_recommendations`, `approval_requests`, `ai_automation_executions`, `shipment_tracking_alerts`, `customer_invoices`, `contract_documents`) directly from MariaDB.
- **Strict Hash-Based Deduplication:** Enforces unique SHA256 hashes per organization, source record, and active issue type to prevent alert fatigue and noise from page reloads or repeated runs.
- **Deterministic SLA Escalation Engine:** Detects items exceeding SLA windows (e.g., 48h pending approval, 24h unresolved high-priority recommendation/shipment exception, 14d overdue invoice) and elevates severity to `CRITICAL` while recording auditable escalation events.
- **Centralized In-App UI:** Implemented a full-page Notification and Escalation Center (`/dashboard/notifications`) with interactive KPI metrics, severity and module filters, read/unread and dismissal controls, escalation audit trail, and user preferences.
- **TopBar Bell & Sidebar Integration:** Connected the global header bell popover and sidebar badge directly to the live notifications service.
- **Strict In-App Read-Only Safety:** Guaranteed zero automated outbound communications (emails, SMS, WhatsApp) and zero unauthorized state mutations.

---

## 2. Existing Architecture Reused

The implementation directly leverages and integrates with:
- **Go Backend & Chi Router:** `backend/internal/server/routes.go` and `backend/cmd/server/main.go`.
- **MariaDB/MySQL (`freel_mysql`):** Connection pool managed via `*sqlx.DB`, enforcing strict foreign keys, indexes, and ACID guarantees.
- **RBAC & Authentication Middleware:** `middleware.GetUserContext(ctx)` enforcing tenant (`org_id`) isolation and recipient authorization (`user_id` and `role_target`).
- **Recommendation Center (Task 2.1):** Consumes recommendations from `ai_recommendations` for high-priority and assigned action items.
- **Approval & Human-in-the-Loop System:** Queries `approval_requests` for pending actions and links directly to review workflows.
- **Workflow Automation Center (Task 2.7):** Evaluates `ai_automation_executions` for failed runs and retry escalation.
- **Event Bus:** Listens to operational events (`EventRFQCreated`, etc.) to trigger real-time synchronization.
- **LogisticsHQ Design System:** Reuses original light/white interface styling (`#ffffff`, `#f8fafc`, borders `#e2e8f0`, headings `#0f172a`, muted `#64748b`).

---

## 3. Notification and Escalation Data Model

Applied via migration `backend/migrations/096_notifications_and_escalation_center.sql`:

### A. `notifications` Table
| Column | Type | Constraints / Purpose |
|---|---|---|
| `id` | BIGINT | Auto-increment primary key |
| `org_id` | BIGINT | Tenant isolation key (NOT NULL, Indexed) |
| `user_id` | BIGINT | Specific recipient user ID; NULL for broadcast/role |
| `role_target` | VARCHAR(50) | Target role (e.g. `SUPER_ADMIN`, `OPERATIONS`, `FINANCE`) |
| `source_module` | VARCHAR(50) | `SHIPMENTS`, `INVOICES`, `CONTRACTS`, `APPROVALS`, `AUTOMATIONS`, `RECOMMENDATIONS`, `SYSTEM` |
| `source_record_type` | VARCHAR(50) | `SHIPMENT_ALERT`, `INVOICE`, `CONTRACT_DOCUMENT`, `APPROVAL`, `AUTOMATION_EXECUTION`, `RECOMMENDATION` |
| `source_record_id` | VARCHAR(100) | Unique identifier of underlying business record |
| `recommendation_id` | BIGINT | Foreign reference to `ai_recommendations` (nullable) |
| `approval_id` | BIGINT | Foreign reference to `approval_requests` (nullable) |
| `automation_execution_id` | BIGINT | Foreign reference to `ai_automation_executions` (nullable) |
| `notification_type` | VARCHAR(100) | Event type constant |
| `title` | VARCHAR(255) | Concise, descriptive title |
| `message` | TEXT | Grounded, factual explanation |
| `severity` | VARCHAR(30) | `INFORMATIONAL`, `LOW`, `MEDIUM`, `HIGH`, `CRITICAL` |
| `priority` | VARCHAR(30) | `LOW`, `MEDIUM`, `HIGH`, `URGENT` |
| `status` | VARCHAR(50) | `ACTIVE`, `RESOLVED`, `DISMISSED`, `EXPIRED` |
| `is_read` | TINYINT(1) | Read status flag (0 = unread, 1 = read) |
| `read_at` | DATETIME | Timestamp when user marked notification as read |
| `is_dismissed` | TINYINT(1) | Dismissed status flag (0 = visible, 1 = dismissed) |
| `dismissed_at` | DATETIME | Timestamp when dismissed |
| `is_escalated` | TINYINT(1) | Escalation flag (0 = normal, 1 = escalated) |
| `escalation_level` | INT | Escalation level (0 = baseline, 1+ = escalated) |
| `escalated_at` | DATETIME | Timestamp of escalation |
| `action_required` | TINYINT(1) | Flag indicating required human action |
| `action_url` | VARCHAR(255) | Deep-link to the source record or detail panel |
| `dedup_hash` | VARCHAR(64) | SHA256 deduplication hash (UNIQUE per org + dedup_hash) |
| `correlation_id` | VARCHAR(255) | Tracing identifier across system boundaries |
| `created_at` | DATETIME | Insertion timestamp |
| `updated_at` | DATETIME | Last modification timestamp |

### B. `notification_escalation_events` Table
| Column | Type | Purpose |
|---|---|---|
| `id` | BIGINT | Primary key |
| `notification_id` | BIGINT | Reference to escalated notification |
| `org_id` | BIGINT | Tenant ID |
| `escalation_level` | INT | Escalation tier (e.g. 1, 2) |
| `escalation_reason` | VARCHAR(255) | Factual explanation of breach or SLA trigger |
| `previous_severity` | VARCHAR(30) | Prior severity level |
| `new_severity` | VARCHAR(30) | Elevated severity level |
| `trigger_type` | VARCHAR(50) | `TIME_THRESHOLD`, `REPEATED_FAILURE`, `MANUAL` |
| `threshold_hours` | INT | SLA threshold hours evaluated |
| `correlation_id` | VARCHAR(255) | Audit correlation identifier |
| `created_at` | DATETIME | Transition timestamp |

### C. `user_notification_preferences` Table
| Column | Type | Purpose |
|---|---|---|
| `id` | BIGINT | Primary key |
| `user_id` | BIGINT | User ID |
| `org_id` | BIGINT | Organization ID |
| `min_severity` | VARCHAR(32) | Minimum severity (`INFORMATIONAL`, `LOW`, `MEDIUM`, `HIGH`, `CRITICAL`) |
| `in_app_enabled` | TINYINT(1) | Master toggle for in-app delivery |
| `assigned_only` | TINYINT(1) | Filter for assigned tasks only |
| `approvals_enabled` | TINYINT(1) | Module toggle for approvals |
| `automations_enabled` | TINYINT(1) | Module toggle for AI automations |
| `recommendations_enabled` | TINYINT(1) | Module toggle for AI recommendations |
| `finance_enabled` | TINYINT(1) | Module toggle for invoices/finance |
| `operations_enabled` | TINYINT(1) | Module toggle for shipments/tracking |
| `compliance_enabled` | TINYINT(1) | Module toggle for contracts/compliance |

---

## 4. Supported Notification Sources

The deterministic engine evaluates real records across six core modules:
1. **AI Recommendations (`ai_recommendations`):**
   - High-priority recommendations (`priority IN ('HIGH', 'URGENT')` or `risk_level IN ('HIGH', 'CRITICAL')`).
   - Assigned recommendations (`assignee_id IS NOT NULL`).
2. **Human-in-the-Loop Approvals (`approval_requests`):**
   - Pending approvals requiring role-based or user-assigned review (`status = 'Pending'`).
3. **AI Workflow Automations (`ai_automation_executions`):**
   - Failed background runs (`status = 'FAILED'`) with error details and retry counts.
4. **Shipments & Tracking Exceptions (`shipment_tracking_alerts`):**
   - Critical and high-severity tracking anomalies (`severity IN ('CRITICAL', 'HIGH')` and `status = 'OPEN'`).
5. **Customer Invoices & Collections (`customer_invoices`):**
   - Overdue customer receivables (`due_date < CURRENT_DATE()` and `balance_due > 0`).
6. **Contracts & Compliance (`contract_documents`):**
   - Contract documents requiring extraction review (`status = 'PENDING_REVIEW'`).

---

## 5. Routing and Recipient Authorization

All notification queries enforce recipient scoping derived securely from the authenticated token:
- **Organization Isolation:** Always filtered by `org_id = ?`.
- **Recipient Checks:**
  - Standard users: `(user_id IS NULL OR user_id = ?) AND (role_target IS NULL OR role_target = ?)`
  - Administrators (`SUPER_ADMIN`, `ADMIN`): `(user_id IS NULL OR user_id = ?)` with access to organization-wide broadcasts and role targets.
- **Client Input Safety:** Neither `org_id` nor `user_id` supplied by query parameters are trusted. Access is anchored strictly in server-validated session context.

---

## 6. Deduplication and Noise-Control Behavior

Deduplication uses a SHA256 digest:
$$\text{dedup\_hash} = \text{SHA256}(\text{org\_id} : \text{source\_module} : \text{source\_record\_type} : \text{source\_record\_id} : \text{notification\_type})$$

- **Unique Constraint:** The database enforces `UNIQUE INDEX idx_notif_dedup (org_id, dedup_hash)`.
- **Idempotent Synchronization:** When the evaluation engine scans records, it queries `GetByDedupHash`. If an active notification exists for that record condition, no new record is inserted.
- **State Changes:**
  - If the underlying issue is resolved (e.g. invoice paid, approval signed), the engine transitions the notification to `RESOLVED`.
  - If an issue is dismissed or read, re-evaluations do not reopen or duplicate it unless severity increases.

---

## 7. Escalation Rules

Escalation is deterministic, auditable, and rate-limited:
1. **Pending Approvals:** If an approval remains `Pending` for $\ge 48$ hours, the engine:
   - Sets `is_escalated = 1`, `escalated_at = NOW()`, `escalation_level = escalation_level + 1`.
   - Elevates severity to `CRITICAL`.
   - Inserts an audit record in `notification_escalation_events` with reason `"Approval pending beyond 48-hour SLA window"`.
2. **High-Priority Recommendations:** If unresolved for $\ge 24$ hours, escalates to `CRITICAL` with reason `"High priority AI recommendation unresolved after 24 hours"`.
3. **Critical Shipment Tracking Exceptions:** If open for $\ge 24$ hours, escalates to `CRITICAL` with reason `"Shipment exception unresolved after 24 hours"`.
4. **Overdue Invoices:** If payment is overdue by $\ge 14$ days, escalates to `CRITICAL` with reason `"Invoice X days past due threshold"`.
5. **Repeated Automation Failures:** If an automated AI job fails with `retry_count >= 2`, escalates to `CRITICAL` with reason `"Automation job repeatedly failed after retries"`.

---

## 8. Notification Preference Behavior

Users can customize in-app delivery preferences via `user_notification_preferences`:
- **Minimum Severity Filter:** Only notifications meeting or exceeding the chosen severity (`INFORMATIONAL`, `LOW`, `MEDIUM`, `HIGH`, `CRITICAL`) are shown.
- **Category Toggles:** Granular enablement for Approvals, Automations, Recommendations, Finance, Operations, and Compliance.
- **Assigned-Only Switch:** Restricts notifications to items specifically assigned to the user.
- **Mandatory Policy Protection:** Critical security and compliance alerts remain visible regardless of preference overrides.

---

## 9. APIs Added or Changed

Mounted at `/api/v1/notifications` in `backend/internal/server/routes.go`:

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/v1/notifications` | List notifications with severity, module, read state, search, and pagination |
| `GET` | `/api/v1/notifications/{id}` | Get single notification ensuring organization tenant isolation |
| `GET` | `/api/v1/notifications/unread-count` | Get count of unread notifications for badge rendering |
| `GET` | `/api/v1/notifications/stats` | Aggregate counts: total, unread, action-required, escalated, critical |
| `POST` | `/api/v1/notifications/{id}/read` | Mark specific notification as read |
| `POST` | `/api/v1/notifications/{id}/unread` | Mark specific notification as unread |
| `POST` | `/api/v1/notifications/{id}/dismiss` | Soft-dismiss notification |
| `POST` | `/api/v1/notifications/read-all` | Mark all visible unread notifications as read |
| `GET` | `/api/v1/notifications/escalations` | List auditable escalation transitions |
| `GET` | `/api/v1/notifications/preferences` | Retrieve user notification preferences |
| `PUT` | `/api/v1/notifications/preferences` | Update user notification preferences |
| `POST` | `/api/v1/notifications/evaluate` | Trigger on-demand synchronization sweep |

---

## 10. UI Pages and Components Changed

1. **New Page:** `frontend/src/pages/dashboard/Notifications/NotificationCenterPage.jsx` & `NotificationCenterPage.css`
   - 5 KPI metric cards with live counts and SLA breach indicator.
   - 3 tabs: Active Notifications, Escalations Audit, Notification Preferences.
   - Interactive search, severity dropdown, module filter, and read status filter.
   - Clean notification cards with left severity stripe, escalation badge, action button, mark-read toggle, and dismiss button.
   - Escalation audit table with transition arrows, trigger SLA, and correlation ID.
2. **TopBar Global Header:** `frontend/src/layouts/AppShell/TopBar.jsx`
   - Real-time unread badge count on notification bell.
   - Popover displaying real notifications, severity indicators, and direct link to source record.
   - "Mark all read" quick action in popover header.
   - Footer button navigating to `/dashboard/notifications`.
3. **Sidebar Navigation:** `frontend/src/layouts/AppShell/Sidebar.jsx`
   - Added "Notifications" entry with Bell icon under the `DOCUMENTS` section with dynamic unread count pill.
4. **Routing:** `frontend/src/App.jsx`
   - Registered `/dashboard/notifications` under protected app shell routes.
5. **Frontend Service:** `frontend/src/services/notificationCenterService.js`
   - Complete API client methods for all 12 notification endpoints.

---

## 11. Security and Tenant-Isolation Verification

- **Multi-Tenant Scoping:** Every SQL query in `repository.go` explicitly enforces `org_id = ?`.
- **Recipient Context Protection:** Handlers extract `UserContext` from middleware; browser-supplied tenant or user IDs in payloads are rejected or overwritten with authenticated session values.
- **Cross-Tenant Isolation Test:** Automated test `TestNotificationCRUDAndIsolation` creates notifications in Org 8881 and validates that Org 8882 queries return zero results.
- **Injection Safety:** All database queries utilize parameterized SQL statements via `sqlx`.

---

## 12. Audit and Observability Behavior

- **Escalation Events Logged:** Every severity promotion records an entry in `notification_escalation_events` capturing `notification_id`, `org_id`, `escalation_level`, `escalation_reason`, `previous_severity`, `new_severity`, `trigger_type`, `threshold_hours`, and `correlation_id`.
- **Audit Correlation IDs:** All generated notifications carry correlation IDs linking them back to the original source entity (e.g. `corr-inv-8`, `corr-appr-12`, `corr-rec-5`).
- **Read & Dismiss Timestamps:** All user actions update `read_at` and `dismissed_at` timestamps for forensic tracking.

---

## 13. Tests Executed and Results

### A. Backend Go Tests (`backend/internal/notifications`)
Command: `go test -v ./internal/notifications/...`
```
=== RUN   TestDeduplicationHash
--- PASS: TestDeduplicationHash (0.00s)
=== RUN   TestNotificationCRUDAndIsolation
--- PASS: TestNotificationCRUDAndIsolation (0.02s)
=== RUN   TestEscalationTransitionAndAudit
--- PASS: TestEscalationTransitionAndAudit (0.01s)
=== RUN   TestNotificationPreferences
--- PASS: TestNotificationPreferences (0.00s)
=== RUN   TestEvaluationEngineDedup
Engine generated 27 notifications on initial run, 0 on re-evaluation (deduplication verified)
--- PASS: TestEvaluationEngineDedup (0.06s)
PASS
ok  	github.com/freel/backend/internal/notifications	0.732s
```

### B. Frontend Vitest Tests
Command: `cmd /c npx vitest run src/services/notificationCenterService.test.js`
```
 ✓ src/services/notificationCenterService.test.js (11 tests) 11ms
 Test Files  1 passed (1)
      Tests  11 passed (11)
```

### C. Full Project Frontend Test Suite
Command: `cmd /c npx vitest run`
```
 Test Files  41 passed (41)
      Tests  239 passed (239)
   Duration  43.73s
```

### D. Production Frontend Build
Command: `cmd /c npm run build`
```
✓ built in 23.40s with exit code 0
```

---

## 14. Browser QA Performed

Using live services against real persisted MariaDB data:
1. **Notification Stats API:** Returned real aggregate stats: Total=27, Unread=27, ActionRequired=27, Escalated=3, Critical=3.
2. **Unread Count API:** Returned 27. When notification #23 was marked as read, count dropped to 26. When reverted to unread, count restored to 27.
3. **Severity Filtering:** Filtered by `CRITICAL` returning exactly 3 overdue customer invoice risk notifications (`INV-2026-0449`, `INV-2026-0450`, `INV-2026-0451`).
4. **Module Filtering:** Filtered by `APPROVALS` returning 9 real pending document, credit limit, and extraction approvals.
5. **Escalation Audit Trail:** Returned 3 auditable escalation events with Level 2 transitions and SLA breach explanations.
6. **Preferences Persistence:** Updated `min_severity` to `HIGH` and verified immediate persistence and retrieval.
7. **Deduplication Verification:** Triggered manual evaluation sweep via `POST /api/v1/notifications/evaluate`, confirming 0 duplicate notifications created.

---

## 15. Any Limitations or Unsupported Delivery Channels

- **In-App Delivery Focus:** Per specification Section 8, outbound channels (SMTP email, AWS SES, Twilio SMS, WhatsApp) are intentionally restricted from autonomous transmission. Outbound notifications remain in-app only to ensure operational safety and prevent uncontrolled external communication.

---

## 16. Confirmation That No Fake Data Was Added

- No mock alerts, demo notifications, or synthetic records were inserted.
- All 27 initial notifications were deterministically derived from real existing records in `freel_mysql` (`customer_invoices`, `approval_requests`, `ai_recommendations`, `ai_automation_executions`, `shipment_tracking_alerts`, `contract_documents`).

---

## 17. Confirmation That No Automatic Outbound Messages or Business Mutations Were Performed

- Zero emails, SMS, WhatsApp messages, carrier updates, or customer notices were dispatched.
- All evaluation sweeps and escalation checks are strictly read-only relative to underlying business entities (no invoices modified, no shipments rescheduled, no contracts altered).

---

## 18. Confirmation That All AI UI Follows the Original White/Light LogisticsHQ Design

- The Notification & Escalation Center interface strictly uses the LogisticsHQ light palette: background `#f8fafc`, container card backgrounds `#ffffff`, borders `#e2e8f0`, headings `#0f172a`, and text `#475569`.
- No dark AI containers, black panels, dark gradients, or neon glows were introduced. All severity and escalation badges match native design standards.
