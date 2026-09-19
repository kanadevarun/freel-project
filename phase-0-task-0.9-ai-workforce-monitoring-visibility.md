# Phase 0 — Task 0.9: AI Workforce Status, Monitoring, and Operational Visibility Foundation

## 1. Files Inspected

### Backend
- [`backend/internal/aitasks/workforce_status.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/aitasks/workforce_status.go): Defined the authoritative 13-state workforce status contract, metadata, agent mappings, and telemetry structures.
- [`backend/internal/aitasks/repository.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/aitasks/repository.go): Implemented `GetWorkforceSummary`, `ListWorkforceTasks`, and `CheckWorkforceHealth` with strict tenant boundaries and bounded queries.
- [`backend/internal/aitasks/service.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/aitasks/service.go): Enforced organization validation (`orgID <= 0` rejection) and wrapped repository operations.
- [`backend/internal/aitasks/handler.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/aitasks/handler.go): Registered HTTP endpoints for summary, task lists, and operational health under `/api/v1/ai/workforce`.
- [`backend/internal/server/routes.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/server/routes.go): Mounted the `/api/v1/ai/workforce` router group with JWT and RBAC middleware.
- [`backend/internal/aitasks/workforce_test.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/aitasks/workforce_test.go): Comprehensive unit tests for status contracts, tenant isolation, bounds, agent mappings, and action eligibility.

### Python Sidecar & Persistence
- [`ai_sidecar/app/persistence/checkpointer.py`](file:///c:/Users/Sai/go/src/freel-project/ai_sidecar/app/persistence/checkpointer.py): Audited checkpoint table schemas and health endpoints on port 8090.
- [`ai_sidecar/main.py`](file:///c:/Users/Sai/go/src/freel-project/ai_sidecar/main.py): Inspected sidecar `/health` output ensuring safe availability reporting without exposing network addresses or secrets.

### Frontend
- [`frontend/src/utils/workforceStatus.js`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/utils/workforceStatus.js): Canonical frontend contract mirroring backend status enums, metadata, colors, and normalization helpers.
- [`frontend/src/services/aiTaskService.js`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/services/aiTaskService.js): Added client methods `getWorkforceSummary()`, `getWorkforceTasks(params)`, and `getWorkforceHealth()`.
- [`frontend/src/components/agent/AgentStatusBadge.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/components/agent/AgentStatusBadge.jsx): Upgraded to support all 13 canonical statuses, pulsing indicators, and safe action buttons.
- [`frontend/src/components/dashboard/MissionControl/AIWorkforceWidget.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/components/dashboard/MissionControl/AIWorkforceWidget.jsx): Re-implemented as a high-density, live-connected command center with 8-agent breakdown, system health, task drawer, safe retry/cancellation, and tab visibility detection.
- [`frontend/src/components/dashboard/MissionControl/AIWorkforceWidget.css`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/components/dashboard/MissionControl/AIWorkforceWidget.css): Dedicated styling sheet for responsiveness, dark-mode tokens, and modal dialogues.
- [`frontend/src/pages/dashboard/Home/OperationalDashboard.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/Home/OperationalDashboard.jsx): Integrated `AIWorkforceWidget` directly between the top KPI cards and operational columns.
- [`frontend/src/pages/dashboard/Finance/InvoicesPage.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/Finance/InvoicesPage.jsx): Added live AI Finance & Reconciliation telemetry strip with safe retry and sign-off actions.
- [`frontend/src/pages/dashboard/Finance/components/InvoiceTable.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/Finance/components/InvoiceTable.jsx): Added per-invoice AI task indicators matching live task references.
- [`frontend/src/pages/dashboard/Contracts/ContractCompliancePanel.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/Contracts/ContractCompliancePanel.jsx): Added live AI Compliance & SLA Monitor strip with retry and status reporting.

---

## 2. Existing Status and Monitoring Architecture

Prior to Task 0.9:
1. **Frontend Disconnection**: `AIWorkforceWidget.jsx` was a static stub expecting hardcoded props (`active_agents`, `tasks_finished`, `health_score`) and was not mounted on any dashboard.
2. **Missing Visibility in Key Modules**: Finance (`InvoicesPage`) and Compliance (`ContractCompliancePanel`) lacked any AI execution telemetry or background task status.
3. **Inconsistent Status Models**: Some parts of the frontend looked for `QUEUED`, others `PROCESSING`, `FAILED`, `ERROR`, or graph-specific states (`COLLECTING_INFORMATION`, `WAITING_FOR_LLM`), without a unified schema.
4. **Un-audited Tenant Boundaries**: Existing AI task queries did not uniformly prevent client query parameters from overriding the authenticated organization context.

---

## 3. Authoritative AI Workforce Status Contract

Thirteen canonical statuses were defined across backend Go structs, frontend JavaScript contracts, and database mapping logic:

| Machine Status | Human Label | Severity | Terminal | Actionable | Retry Permitted | Requires Sign-Off |
|---|---|---|---|---|---|---|
| `queued` | Queued | Neutral | No | Yes (Cancel) | No | No |
| `claimed` | Claimed | Info | No | Yes (Cancel) | No | No |
| `processing` | Processing | Info | No | Yes (Cancel) | No | No |
| `waiting_for_approval` | Awaiting Sign-Off | Warning | No | Yes (Sign-Off) | No | Yes |
| `paused` | Paused | Neutral | No | Yes (Cancel) | No | No |
| `retrying` | Retrying | Warning | No | Yes (Cancel) | No | No |
| `completed` | Completed | Success | Yes | No | No | No |
| `completed_with_failover` | Completed (Failover) | Success | Yes | No | No | No |
| `completed_in_mock_mode` | Completed (Mock) | Neutral | Yes | No | No | No |
| `failed` | Failed | Danger | Yes | Yes (Retry) | Yes | No |
| `cancelled` | Cancelled | Neutral | Yes | No | No | No |
| `stale` | Stale / Interrupted | Danger | No | Yes (Retry) | Yes | No |
| `unknown` | Unknown | Neutral | No | No | No | No |

---

## 4. Status Mapping Rules

State resolution follows `ResolveTaskWorkforceStatus(task, isFailover, isMock)`:
1. **Stale Lease Preemption**: If a task has status `PROCESSING` but its `lease_expires_at` is in the past, it resolves to `stale`.
2. **Claimed vs Queued**: If status is `QUEUED` but `worker_id` is non-nil, it resolves to `claimed`.
3. **HITL Interrupt**: If status is `WAITING_FOR_APPROVAL` or an approval record is referenced, it resolves to `waiting_for_approval`.
4. **Failover Execution**: If status is `COMPLETED` and execution telemetry indicates `provider_failover = true`, it resolves to `completed_with_failover`.
5. **Mock Mode Execution**: If status is `COMPLETED` and development mode mock was active, it resolves to `completed_in_mock_mode`.
6. **Permanent Failure vs Retrying**: If attempts < max_attempts and backoff is active, it resolves to `retrying`; if attempts exhausted, it resolves to `failed`.
7. **Canonical Agent Mapping**: Maps tasks by `task_type` across the 8 primary business agents:
   - `pricing`: `PRICING_ANALYZE`, `PRICING_RESUME` (Module: `PRICING`)
   - `sales`: `EMAIL_PARSE`, `SALES_INBOUND` (Module: `SALES`)
   - `operations`: `CARRIER_UPDATE_PARSE`, `OPERATIONS_TRACKING` (Module: `OPERATIONS`)
   - `contracts`: `CONTRACT_PARSE`, `DOC_PROCESS`, `RATE_EXTRACT` (Module: `CONTRACTS`)
   - `compliance`: `COMPLIANCE_AUDIT`, `DOC_VERIFY` (Module: `COMPLIANCE`)
   - `finance`: `BILL_RECONCILE`, `INVOICE_AUDIT` (Module: `FINANCE`)
   - `leads`: `LEAD_SCORING` (Module: `LEADS`)
   - `outreach`: `OUTREACH_GEN` (Module: `OUTREACH`)

---

## 5. Organization and RBAC Enforcement

1. **Strict Server-Side Context**:
   - Every database query enforces `WHERE org_id = ?`.
   - `orgID` is extracted strictly from the validated JWT token via `auth.GetOrgID(c)`.
   - Any client-supplied `organization_id` in URL parameters or query strings is completely ignored and cannot override authenticated context.
2. **Task Ownership Verification**:
   - Before returning task details or permitting retry/cancel actions, the service validates `task.OrgID == user.OrgID`. Mismatches return `ErrTenantMismatch` (HTTP 403/404).
3. **Secret Redaction**:
   - Internal stack traces, raw prompts, full LLM messages, database connection strings, and document/financial payloads are omitted from operational responses.
   - Task error messages undergo sanitization (`redactSensitiveText`).

---

## 6. AIWorkforceWidget Integration Details

Located at [`frontend/src/components/dashboard/MissionControl/AIWorkforceWidget.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/components/dashboard/MissionControl/AIWorkforceWidget.jsx):
1. **Live Metrics Grid**:
   - Active Workflows (Running + Queued)
   - Awaiting Sign-Off (HITL Interrupts)
   - Attention Required (Failed + Stale)
   - Completed within 24h with Average Execution Duration
2. **8-Agent Operational Matrix**:
   - Real-time status pills (`ACTIVE`, `ATTENTION`, `IDLE`).
   - Interactive filtering: clicking an agent card filters the operational tasks table to that specific agent.
3. **Infrastructure Health Strip**:
   - Live checks for Python sidecar, database checkpoint table, worker heartbeat, and provider readiness.
4. **Interactive Tasks Table**:
   - Filter tabs: All Recent, Attention Needed, Sign-Off Required, Active Workflows.
   - Direct actions: `Retry` for eligible failed/stale tasks, `Cancel` for non-terminal tasks, `Sign Off` navigating to `/dashboard/approvals`, and `Go to Record` navigating to the related business entity.
5. **Operational Diagnostics Modal**:
   - Safe inspection of task duration, retry counts, error category, and failover status without exposing raw prompts or keys.
6. **Tab Visibility Lifecycle**:
   - Uses `document.visibilityState`: pauses background polling (30s) when the tab is hidden and refreshes immediately when the user returns.

---

## 7. Module-Level Status Integration

1. **Finance (`InvoicesPage.jsx` & `InvoiceTable.jsx`)**:
   - Top-level AI Finance & Reconciliation banner displaying current audit status, running task counts, safe retry action, and approval navigation.
   - Per-row `AgentStatusBadge` in the invoices table showing matching AI reconciliation tasks.
2. **Compliance (`ContractCompliancePanel.jsx`)**:
   - Live AI Compliance & SLA Monitor strip tracking legal clause validation, risk exception alerts, and automated review states.
3. **Home Dashboard (`OperationalDashboard.jsx`)**:
   - Embedded `AIWorkforceWidget` between Row 1 KPI metrics and Row 2 operational columns, giving operators instant visibility upon login.

---

## 8. Operational Action Behavior

1. **Retry Action**:
   - Authorized via `POST /api/v1/ai/tasks/:id/retry`.
   - Permitted only for tasks with `failed` or `stale` status.
   - Idempotent: resets lease expiry and increments attempt count; emits `RETRY_TASK` universal audit log entry.
2. **Cancel Action**:
   - Authorized via `POST /api/v1/ai/tasks/:id/cancel`.
   - Permitted only for non-terminal tasks (`queued`, `claimed`, `processing`, `retrying`).
   - Emits `CANCEL_TASK` audit log entry with caller attribution.
3. **Approval Sign-Off**:
   - Deep-links directly to the existing unified approval modal (`/dashboard/approvals`).

---

## 9. Health and Degradation Behavior

Endpoint: `GET /api/v1/ai/workforce/health`
- **Sidecar Health**: Probes Python sidecar on `:8090/health` (returns `healthy` when reachable with persistent `MariaDBSaver`, `degraded` or `unavailable` if connection refused).
- **Checkpointer Health**: Executes `SELECT 1 FROM ai_checkpoints LIMIT 1` to verify MariaDB checkpoint storage.
- **Worker Health**: Evaluates latest worker heartbeat timestamp; flags `idle` or `degraded` if overdue.
- **Provider Readiness**: Checks primary and secondary provider configuration. If failover is active, system status is marked `degraded` rather than `unavailable`.

---

## 10. Tests Added or Updated

- `backend/internal/aitasks/workforce_test.go`:
  - `TestWorkforceStatusContract`: Validates 13 status invariants, labels, terminal flags, and retry eligibility.
  - `TestResolveTaskWorkforceStatus`: 12 subtests covering queued, claimed, processing, expired lease (stale), sign-off, failover, mock mode, failed, and cancelled states.
  - `TestMapTaskTypeToAgent`: Validates canonical mapping of all 8 business agents.
  - `TestResolveRelatedRef`: Tests entity label and navigation generation.
  - `TestListWorkforceTasks_IsolationAndFiltering`: Tests organization boundaries, result bounding (LIMIT/OFFSET), and column scans.
  - `TestOperationalActionEligibility`: Tests retry and cancellation status eligibility constraints.
- `test_workforce_endpoints.py`:
  - Integration test verifying JWT authentication, organization isolation against query param tampering, workforce health, summary aggregations, and tasks list.
- Frontend Build:
  - Vite production build (`npm run build`) verified with 0 syntax or compilation errors.

---

## 11. Test Results

### Go Unit Tests
```
=== RUN   TestWorkforceStatusContract
--- PASS: TestWorkforceStatusContract (0.00s)
=== RUN   TestResolveTaskWorkforceStatus
    --- PASS: TestResolveTaskWorkforceStatus/Queued_Unclaimed (0.00s)
    --- PASS: TestResolveTaskWorkforceStatus/Queued_Claimed (0.00s)
    --- PASS: TestResolveTaskWorkforceStatus/Active_Processing (0.00s)
    --- PASS: TestResolveTaskWorkforceStatus/Stale_Processing_(Expired_Lease) (0.00s)
    --- PASS: TestResolveTaskWorkforceStatus/Waiting_for_Sign-Off (0.00s)
    --- PASS: TestResolveTaskWorkforceStatus/Retrying_Backoff (0.00s)
    --- PASS: TestResolveTaskWorkforceStatus/Standard_Completed (0.00s)
    --- PASS: TestResolveTaskWorkforceStatus/Completed_with_Failover (0.00s)
    --- PASS: TestResolveTaskWorkforceStatus/Completed_in_Mock_Mode (0.00s)
    --- PASS: TestResolveTaskWorkforceStatus/Permanent_Failure (0.00s)
    --- PASS: TestResolveTaskWorkforceStatus/Cancelled_Task (0.00s)
    --- PASS: TestResolveTaskWorkforceStatus/Nil_Task (0.00s)
=== RUN   TestMapTaskTypeToAgent
--- PASS: TestMapTaskTypeToAgent (0.00s)
=== RUN   TestResolveRelatedRef
--- PASS: TestResolveRelatedRef (0.00s)
=== RUN   TestListWorkforceTasks_IsolationAndFiltering
--- PASS: TestListWorkforceTasks_IsolationAndFiltering (0.00s)
=== RUN   TestOperationalActionEligibility
--- PASS: TestOperationalActionEligibility (0.00s)
PASS
ok      github.com/freel/backend/internal/aitasks       2.228s
```

### Integration Smoke Test
```
--- 1. Public Health ---
Health: 200 {"message":"Freel backend is running","success":true}

--- 2. Login to get JWT ---
Login successful! Expected Org ID: 2, Token present: True

--- 3. Workforce Health ---
Workforce Health (200):
{
  "overall_status": "healthy",
  "backend_status": "healthy",
  "sidecar_status": "healthy",
  "worker_status": "healthy",
  "queue_status": "healthy",
  "checkpoint_status": "healthy",
  "primary_provider_status": "healthy",
  "failover_status": "ready"
}

--- 4. Workforce Summary ---
Workforce Summary (200):
{
  "org_id": 2,
  "failed_tasks": 4,
  "health": { "overall_status": "healthy" }
}

--- 6. Test Organization Isolation ---
Passed ?organization_id=9999, returned org_id in payload: 2
SUCCESS: Organization isolation verified. Client query cannot override server JWT context.
```

### Frontend Build
```
✓ 3092 modules transformed.
dist/index.html                           2.97 kB │ gzip:   0.94 kB
dist/assets/index-Be124Plt.css        1,455.16 kB │ gzip: 224.27 kB
dist/assets/index-hq_eSUmu.js         2,785.74 kB │ gzip: 553.16 kB
✓ built in 26.24s
```

---

## 12. Limitations

1. Task duration averages rely on recorded `completed_at` timestamps; in-flight tasks do not contribute to historical average duration calculations until finished.
2. High task volume (>10,000 tasks/org) pagination is bounded at `limit=50` to safeguard MariaDB query performance.

---

## 13. Deployment and Configuration Requirements

1. **Environment Variables**:
   - `AI_SIDECAR_URL`: URL of the Python AI sidecar (defaults to `http://127.0.0.1:8090`).
   - `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`: MariaDB connection credentials.
2. **Database Schema**:
   - `ai_processing_tasks`, `ai_execution_traces`, and `ai_checkpoints` tables must exist (already provisioned and verified in MariaDB).
