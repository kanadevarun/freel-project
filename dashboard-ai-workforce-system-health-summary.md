# Dashboard Task 7: Compact AI Workforce and System Health Summary

## Executive Overview
In Dashboard Task 7, the developer-oriented AI telemetry, repetitive recommendations, stack traces, and verbose operational monitoring cards were consolidated into two compact, business-focused summary cards:
1. **AI Workforce**: Summarizes whether autonomous agents are operating, active background task volume, tasks requiring human sign-off, blocked or failed tasks, and 24h completed throughput.
2. **System Health**: Provides a calm, authoritative operational health status for core infrastructure components (Go API Backend, Python AI Sidecar, MariaDB Storage, Queue Workers), with freshness indicators and permission-gated audit access.

The redesign strictly adheres to the original **LogisticsHQ Light UI** design system:
- Preserved the `#0A1128` deep navy sidebar and global application shell.
- Preserved white cards (`#FFFFFF`), neutral borders (`#E2E8F0`), standard typography, and semantic status colors.
- Zero black or dark AI panels.
- Zero gradients, decorative charts, or artificial animations.
- 100% grounded in real persistent backend data; zero fake data or database resets.

---

## 1. Files and Components Changed

| File Path | Type | Modifications |
| :--- | :--- | :--- |
| `frontend/src/pages/dashboard/Home/OperationalDashboard.jsx` | React JSX | Replaced technical telemetry panels with compact "AI Workforce" and "System Health" cards; implemented resilient `Promise.allSettled` fetching; added permission gating (`canViewAIWorkforce`, `canViewAuditLogs`); mapped business status and subsystem health. |
| `frontend/src/pages/dashboard/Home/OperationalDashboard.css` | Stylesheet | Added compact card styling, semantic status chips (`.compact-health-badge`, `.health-status-chip`), subtle 4-column metric pill grid, and responsive wrapping for small laptop/mobile viewports. |
| `frontend/src/context/RBACContext.jsx` | React Context | Exported `RBACContext` so unit tests and nested components can provide custom role/permission states. |
| `frontend/src/__tests__/pages/dashboard/Home/OperationalDashboard.test.jsx` | Test | Updated test assertions to match compact headers and added 5 new unit tests covering business metric mapping, subsystem health indicators, partial service failures, and RBAC permission gating. |
| `scratch_verify_task7.py` | Verification Script | Automated Playwright script verifying element presence, metrics extraction, and 36-matrix responsive & zoom test (6 viewports × 6 zoom levels: 80% to 150%). |

---

## 2. Data Sources and APIs Reused

All metrics and statuses are derived directly from authoritative backend services and persistent MariaDB records:

1. **AI Workforce Summary**:
   - Endpoint: `GET /api/v1/ai/workforce/summary`
   - Client Service: `aiTaskService.getWorkforceSummary()`
   - Fields Consumed:
     - `total_active_tasks`: In-progress background tasks.
     - `waiting_for_approval_tasks`: Tasks gated on human operator review.
     - `failed_tasks` + `stale_tasks`: Failed or lease-expired tasks requiring operator attention.
     - `completed_recent_24h`: Tasks completed in the trailing 24 hours.
     - `health.overall_status`: High-level operational status (`operational`, `degraded`, `unavailable`).
     - `last_updated`: Freshness timestamp.
2. **System Health Summary**:
   - Endpoint: `GET /api/v1/monitoring/health`
   - Client Service: `monitoringService.getHealthSummary()`
   - Subsystems Evaluated:
     - `go_backend`: Core API server (port 8080).
     - `ai_sidecar`: Python LangGraph sidecar (port 8090).
     - `database`: MariaDB persistent storage (port 3306).
     - `queue_worker`: Background queue worker pool and backlog.
     - `evaluated_at`: Real-time health evaluation timestamp.
3. **Recommendation Stats (Fallback for Approvals)**:
   - Endpoint: `GET /api/v1/recommendations/stats`
   - Client Service: `recommendationService.getStats()`

---

## 3. AI Telemetry Removed or Consolidated

The following developer-focused telemetry and verbose outputs were removed from the main dashboard overview:
- **Raw prompts and completions**: Completely removed from the dashboard; inspected only in dedicated debug tools.
- **Model token counts and internal latency graphs**: Removed from the business dashboard; available on `/dashboard/ai-monitoring`.
- **Stack traces and error dumps**: Suppressed on dashboard surfaces; translated into calm status labels (`Attention needed` / `Requires attention`).
- **Internal Python module names and queue names**: Hidden from business users.
- **Duplicate recommendation blocks**: Consolidated into Priority Actions (Section 3) and high-level workforce counters.

---

## 4. Architecture & Responsibility Separation

### Python AI Sidecar Responsibilities (:8090)
- Executes LangGraph state machine workflows and agent reasoning steps.
- Generates structured JSON responses conforming to strict schemas.
- Classifies document types, RFQ risks, and pricing recommendations.
- **Strict Guardrail**: Python never connects directly to the production MariaDB database, never runs SQL queries, never bypasses Go authorization, and never mutates shipments, invoices, or contracts directly.

### Go Application Layer Responsibilities (:8080)
- Authenticates requests (JWT validation, session tokens).
- Enforces Role-Based Access Control (RBAC) and strict tenant isolation (`tenant_id = ?`).
- Implements the centralized Action System and Human-in-the-Loop approval enforcement.
- Persists all business entities, audit logs, and workflow run states to MariaDB.
- Communicates with the Python sidecar via authenticated HTTP RPC and handles retries, timeouts, and failovers.

---

## 5. Health Status Calculation & Freshness Behavior

### AI Workforce Status Logic
- **`Checking...`**: Initial non-blocking loading state.
- **`Operational`**: When backend reports overall status as operational and failed/blocked task count is 0.
- **`Attention needed`**: When backend reports degraded status or `failed_tasks + stale_tasks > 0`.
- **`Unavailable`**: When sidecar or workforce service returns unavailable or endpoint call fails.

### System Health Overall Status Logic
- **`Checking...`**: When health evaluation is in progress.
- **`Healthy`**: When Go API, Sidecar, MariaDB, and Queue Workers report healthy status.
- **`Attention needed`**: When any single subsystem reports degraded or lagging status.
- **`Unavailable`**: When critical services report offline or unreachable.
- **`Status unavailable`**: When the monitoring endpoint fails or cannot be reached.

### Resilient Partial-Failure Handling
- Data fetching uses `Promise.allSettled([monitoringService.getHealthSummary(), aiTaskService.getWorkforceSummary()])`.
- If the AI sidecar is offline or restarting, the System Health card preserves the operational status of Go Backend, MariaDB Storage, and Queue Workers.
- Temporary sidecar downtime never crashes or prevents access to the core LogisticsHQ freight management features.

---

## 6. RBAC Permission Enforcement

The dashboard adheres strictly to user permissions:
- **`View AI Workforce →`**: Rendered only if user has `Admin`, `Super Admin`, `Developer`, or `Executive` roles, or possesses `view_ai_workforce` or `view_audit_logs` permissions.
- **`Audit Logs →`**: Rendered only if user has administrative or compliance audit permissions.
- **Restricted Users**: For viewers and warehouse operators without AI permissions, the high-level status and counts remain visible for situational awareness, while links and drill-downs are cleanly omitted.
- **Approvals Gating**: AI recommendation sign-offs navigate directly to `/dashboard/approvals` rather than mutating state inline.

---

## 7. Testing & Verification Results

### Automated Go Tests
Command:
```bash
go test -v ./internal/monitoring/... ./internal/aitasks/... ./internal/dashboard/...
```
Result: **PASS** (100% passed across all monitoring, AI task orchestration, and dashboard aggregation suites).

### Automated Frontend Vitest Suites
Commands:
```bash
cmd /c npx vitest run src/__tests__/pages/dashboard/Home/OperationalDashboard.test.jsx
cmd /c npx vitest run
```
Results:
- `OperationalDashboard.test.jsx`: **16 passed (16 tests)**
- Full frontend test suite: **54 test files passed, 320 tests passed, 0 failures**.

### Production Build Check
Command:
```bash
cmd /c npm run build
```
Result: **Built cleanly in 16.03s** with zero syntax, type, or bundling errors.

### Browser-Based Playwright Verification
Script: `scratch_verify_task7.py` (Headless Chrome)
- **Live Data Extracted**:
  - AI Workforce Card: Visible
  - Active Tasks: 0 (in progress)
  - Awaiting Review: 0 (requires sign-off)
  - Blocked / Issues: 7 (real database records requiring attention)
  - Completed (24h): 0 (autonomous jobs)
  - Go API Backend: Healthy
  - Python AI Sidecar: Healthy
  - MariaDB Storage: Connected
  - Queue Workers: 0 Backlog
- **36-Matrix Verification (6 viewports × 6 zoom levels)**:
  - Viewports: `1366×768`, `1440×900`, `1920×1080`, `1024×768` (small laptop), `1200×800` (mid desktop), `390×844` (mobile).
  - Zoom levels: `80%`, `90%`, `100%`, `110%`, `125%`, `150%`.
  - **Results**: **36 / 36 passed (0 failures)**.
  - Verified: Independent scrolling, no card-induced horizontal overflow, stable navy sidebar, natural text wrapping, accessible action buttons.

---

## 8. Preserved Invariants & Non-Blocking Notes

1. **Persistent Data Integrity**: No mock data was introduced; no database tables were reset, truncated, or dropped.
2. **Design Language Integrity**: No dark AI cards or gradients were added. The UI maintains complete visual harmony with the LogisticsHQ design system.
3. **Approval Security**: Main dashboard provides read-only visibility; state changes must pass through the Go Action System and Approval Center.
