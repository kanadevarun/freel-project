# Phase 2 Task 2.11 Implementation Report: AI Performance, Cost, and Quality Monitoring

## 1. Executive Summary
Phase 2 Task 2.11 delivers a unified, production-grade AI Performance, Cost, and Quality Monitoring system for LogisticsHQ. It provides end-to-end observability across the Centralized AI Runtime, AI Workforce assistants, Recommendation Center, Action & Approval systems, AI Queue workers, AI Memory, and Safety/Security gates.

All telemetry and metrics are strictly backed by persistent MariaDB storage (`freel_mysql`), enforce multi-tenant isolation, sanitize sensitive customer data/credentials before storage and rendering, compute token and cost estimations via a centralized pricing catalog, and adhere to LogisticsHQ's clean white/light design system with zero dark/black AI panels.

---

## 2. Existing Architecture Reused
The implementation directly extends LogisticsHQ's proven architectural components without creating parallel logging, monitoring, or runtime frameworks:
- **Database**: MariaDB (`freel_mysql`) accessed via `database/sql` with transactional consistency, connection pooling, and standard schema migrations.
- **AI Runtime & Gateway**: `backend/internal/ai/gateway.go` instrumented to record real-time latency, token usage, cost, feature, assistant, module, correlation ID, safety flags, and grounding statuses.
- **Queue Worker**: Extends existing `internal/queue` queries to calculate queue age, backlog, running vs completed/failed jobs, and dead-letter/permanent failure categories.
- **Recommendations**: Reuses `internal/recommendations` table `ai_recommendations` to evaluate generation volume, severity distributions, user dispositions (accepted, dismissed, expired), conversion into actions, and resolution latency.
- **Actions & Approvals**: Reuses `internal/action` tables `action_records` and `action_approval_requests` to track execution outcomes, approval failure rates, duplicate conflict prevention, and decision durations.
- **AI Memory**: Reuses `internal/ai_memory` metadata counters (proposals, confirmations, rejections, deletions) to provide privacy-safe observability without exposing raw memory content.
- **Security & RBAC**: Integrated with Chi router middleware (`middleware.RequireAuth`, `middleware.GetUserContext`), strictly deriving `org_id` and user context from verified session tokens.

---

## 3. Monitoring Data Model
Structured entities introduced in `backend/internal/monitoring/model.go`:

1. **`ai_execution_traces` (extended)**:
   - Primary trace record capturing every runtime invocation.
   - Fields: `id`, `organization_id`, `user_id`, `feature`, `assistant`, `module`, `request_type`, `runtime_route`, `model_provider`, `model_name`, `status`, `started_at`, `completed_at`, `duration_ms`, `input_tokens`, `output_tokens`, `total_tokens`, `estimated_cost`, `cost_currency`, `retry_count`, `error_category`, `safety_status`, `grounding_status`, `is_memory_assisted`, `correlation_id`, `created_at`.
   - Indexed on `organization_id`, `feature`, `assistant`, `model_provider`, `status`, `started_at`, `correlation_id`.

2. **`ai_model_pricing`**:
   - Centralized pricing metadata per provider and model.
   - Fields: `id`, `provider`, `model_name`, `input_rate_per_mtoken`, `output_rate_per_mtoken`, `currency`, `effective_date`, `is_active`, `created_at`, `updated_at`.

3. **`ai_quality_evaluations`**:
   - Automated quality and grounding verification records.
   - Fields: `id`, `organization_id`, `trace_id`, `feature`, `evaluation_type`, `score`, `pass_status`, `evidence_status`, `safety_status`, `reviewer_type`, `reviewer_id`, `evaluation_notes`, `correlation_id`, `created_at`.

4. **`ai_health_thresholds`**:
   - Configurable operational boundaries for health evaluation.
   - Fields: `id`, `organization_id`, `metric_name`, `warning_threshold`, `critical_threshold`, `unit`, `is_active`, `updated_by`, `updated_at`.

5. **`ai_security_events`**:
   - Audit trail for AI safety, injection, and authorization anomalies.
   - Fields: `id`, `organization_id`, `event_type`, `severity`, `feature`, `assistant`, `correlation_id`, `details`, `detected_at`.

---

## 4. Database Migrations
Applied migration: `backend/migrations/098_ai_performance_cost_and_quality_monitoring.sql`
- Safely added trace columns (`user_id`, `feature`, `assistant`, `module`, `request_type`, `runtime_route`, `input_tokens`, `output_tokens`, `total_tokens`, `estimated_cost`, `cost_currency`, `retry_count`, `safety_status`, `grounding_status`, `is_memory_assisted`) to `ai_execution_traces` via schema checks (`information_schema.COLUMNS`).
- Added composite indexes for high-speed tenant querying:
  - `idx_traces_org_feature`
  - `idx_traces_org_assistant`
  - `idx_traces_org_provider`
  - `idx_traces_org_status`
  - `idx_traces_org_started`
- Created tables `ai_model_pricing`, `ai_quality_evaluations`, `ai_health_thresholds`, `ai_security_events`.
- Preserved 100% of existing operational, RFQ, shipment, contract, recommendation, and trace records.

---

## 5. API Endpoints
All endpoints are secured under `/api/v1/monitoring`, requiring authenticated tenant sessions:

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/v1/monitoring/health` | Comprehensive multi-system health summary with deterministic state (`HEALTHY`, `DEGRADED`, `UNAVAILABLE`) and active alerts |
| `GET` | `/api/v1/monitoring/traces` | Paginated execution traces with filtering by `feature`, `assistant`, `provider`, `model`, `status`, `from`, `to`, `correlation_id` |
| `GET` | `/api/v1/monitoring/traces/{id}` | Detailed trace information with sanitized error diagnostics |
| `GET` | `/api/v1/monitoring/cost` | Aggregated token counts, request counts, and estimated cost breakdowns by day, feature, assistant, model, and provider |
| `GET` | `/api/v1/monitoring/performance` | Latency distribution (Avg, P50, P95, P99), throughput, error rate, and retry counts |
| `GET` | `/api/v1/monitoring/quality` | Grounding pass rates, evidence validity, automated evaluation scores, and review distributions |
| `GET` | `/api/v1/monitoring/queue` | Queue worker metrics: pending, running, completed, failed, dead-letter, queue age, and execution duration |
| `GET` | `/api/v1/monitoring/recommendations` | Generation counts, severity breakdown, acceptance/dismissal rates, action conversion, and resolution latency |
| `GET` | `/api/v1/monitoring/approvals` | Action proposal counts, approval rate, rejection rate, decision durations, execution failures, and idempotency conflicts |
| `GET` | `/api/v1/monitoring/memory` | Safe memory personalization metrics: proposals, confirmations, rejections, deletions, retrieval failures |
| `GET` | `/api/v1/monitoring/security` | Safety incidents, prompt injection attempts, cross-tenant rejection events |
| `GET` | `/api/v1/monitoring/pricing` | Active model pricing rates catalog |
| `POST` | `/api/v1/monitoring/pricing` | Upsert model pricing rate (Admin authorized) |
| `GET` | `/api/v1/monitoring/thresholds` | Active operational threshold boundaries |
| `PUT` | `/api/v1/monitoring/thresholds` | Update operational threshold boundaries (Admin authorized) |

---

## 6. Runtime Instrumentation
In `backend/internal/ai/gateway.go`:
- Invocations are intercepted by `recordTelemetry()`.
- Captures start and end timestamps (`duration_ms = end.Sub(start).Milliseconds()`).
- Calculates token usage: uses provider response tokens when available, or applies heuristic estimation (`1 token ≈ 4 characters` for prompts and completions).
- Computes estimated cost in real time via `monitoring.CalculateCost()`.
- Records safety status (`PASSED` vs `REJECTED`) and grounding validation status (`GROUNDED`, `UNGROUNDED`, `UNVERIFIED`).
- Propagates tenant context, user identity, correlation IDs, and runtime route (`/v1/chat`, `/v1/recommendations/generate`, `/v1/actions/execute`).
- Asynchronous DB writes guarantee that monitoring persistence failures never block or crash the primary AI request path.

---

## 7. Worker and Scheduler Instrumentation
In `backend/internal/monitoring/repository.go`:
- Tracks queue backlog and worker execution by querying the persistent queue records.
- Computes oldest queued job age (`TIMESTAMPDIFF(SECOND, MIN(created_at), NOW())`).
- Aggregates execution duration averages and failure categories.
- Detects dead-letter and permanently failed jobs (`status = 'FAILED' AND attempts >= max_attempts`).
- Re-execution idempotency: traces track `retry_count` without creating duplicate records for retried jobs.

---

## 8. Cost Calculation Approach
Implemented in `backend/internal/monitoring/pricing.go`:
- Built-in pricing catalog covering Gemini (Flash & Pro), OpenAI (GPT-4o, GPT-4o-mini), Anthropic (Claude 3.5 Sonnet, Haiku), and Mock models.
- Database overrides in `ai_model_pricing` allow dynamic rate updates without code changes.
- Exact vs Estimated labeling: when provider usage tokens are missing, token counts are estimated from character length and marked as estimated.
- Cost formula:
  $$\text{Cost} = \left(\frac{\text{Input Tokens}}{1,000,000} \times \text{Input Rate}\right) + \left(\frac{\text{Output Tokens}}{1,000,000} \times \text{Output Rate}\right)$$
- Costs are aggregated strictly within the authenticated tenant's scope across dimensions: day, feature, assistant, model, and provider.

---

## 9. Quality and Grounding Checks
Implemented in `backend/internal/monitoring/service.go` and `repository.go`:
- Grounding verification: Evaluates whether AI outputs reference existing LogisticsHQ database records (shipments, RFQs, rate contracts, invoices).
- Evidence completeness: Verifies presence of supporting facts, confidence scores, and data freshness timestamps.
- Quality evaluations stored in `ai_quality_evaluations` with deterministic pass/fail statuses.
- Automated checks labeled clearly as "Automated Check" (never as subjective ground truth).

---

## 10. Safety and Security Monitoring
Implemented in `backend/internal/monitoring/redaction.go` and `repository.go`:
- Cross-tenant access attempts are detected, blocked, and logged to `ai_security_events`.
- Prompt injection detection: attempts to override system prompts or bypass safety gates are recorded with correlation IDs.
- Sensitive-content rejection: triggers on regex patterns matching API keys, JWT tokens, credit cards, or private secrets.
- Redaction: Error diagnostics and trace summaries are sanitized via `SanitizeDetails()` and `RedactError()`, replacing secrets with `[REDACTED]`.

---

## 11. Health Thresholds
Default operational boundaries defined in `backend/internal/monitoring/thresholds.go`:

| Metric | Warning | Critical | Unit | Subsystem Affected |
|---|---|---|---|---|
| `failure_rate` | > 5.0 | > 15.0 | % | AI Runtime |
| `p95_latency_ms` | > 3000 | > 8000 | ms | AI Runtime |
| `queue_backlog` | > 20 | > 100 | jobs | AI Queue Worker |
| `oldest_queued_sec`| > 300 | > 900 | s | AI Queue Worker |
| `worker_failure_rate`| > 5.0 | > 20.0 | % | AI Queue Worker |
| `security_alerts_24h`| > 1 | > 5 | events | AI Safety & Security |
| `grounding_failure_rate`| > 10.0 | > 25.0 | % | Quality & Grounding |

Health synthesis combines threshold breaches with direct connectivity checks for MariaDB, Go Backend, AI Sidecar, and Runtime Gateway to produce a deterministic state: `HEALTHY`, `DEGRADED`, or `UNAVAILABLE`.

---

## 12. Retention and Performance Behavior
- All monitoring queries are strictly bounded (`LIMIT`, `OFFSET`) with mandatory tenant filtering (`organization_id = ?`).
- Pagination enforced across traces (default 25 records per page, max 100).
- Indexed columns (`organization_id`, `started_at`, `status`, `feature`) prevent full table scans.
- Traces and evaluations are preserved without deleting core business or audit records.

---

## 13. UI Implementation
- **Design Theme**: 100% adheres to LogisticsHQ's light/white UI theme (`#ffffff` background cards, `#f8fafc` outer canvas, `#e2e8f0` borders, `#0f172a` slate typography).
- **Zero Dark Panels**: No dark backgrounds, glowing neon borders, or black cards.
- **Top KPI Cards**: 6 summary cards with trend indicators and status colors (Overall Health, Success Rate, P95 Latency, Total Invocations, Est. Cost, Active Alerts).
- **Subsystem Health Matrix**: Real-time status for AI Runtime, Sidecar, MariaDB, Queue Worker, and Safety Gates with latency and message badges.
- **Tabs**:
  1. *Runtime & Executions*: Paginated execution trace table with multi-filter bar (feature, status, provider, correlation ID) and click-to-view trace detail modal.
  2. *Cost & Token Breakdown*: Cost by provider, model, feature, and daily timeline with token counts and clear "Estimated" disclaimers.
  3. *Quality & Grounding Checks*: Grounding pass rates, evidence validation metrics, and automated evaluation audit trail.
  4. *Queue & Worker Status*: Queue backlog gauge, oldest job age, worker health, and failure category breakdown.
  5. *Recommendations & Approvals*: Acceptance rates, conversion to actions, decision duration distributions, and execution failure counts.
  6. *Pricing & Thresholds*: View and edit model pricing rates and operational health threshold boundaries.
- **AI Workforce Integration**: `AIWorkforceWidget.jsx` updated with an AI Observability badge showing real-time health, success rate, and direct navigation to the monitoring center.
- **Settings Navigation**: `SettingsLayout.jsx` updated with an `Activity` icon linking to `AI Monitoring & Quality`.

---

## 14. Tenant-Isolation Validation
- Verified that all queries enforce `WHERE organization_id = ?`.
- Verified via integration tests that Organization 1 cannot view Organization 2's traces, metrics, cost summaries, or security alerts.
- Request-supplied `organization_id` or `user_id` query parameters are strictly ignored; scope is extracted from authenticated user context.

---

## 15. Privacy and Redaction Validation
- Raw prompts, responses, customer message bodies, and documents are excluded from monitoring tables.
- `SanitizeDetails()` regex scrubbing removes Bearer tokens, API keys, password parameters, and credit card numbers from error logs and metadata.
- Memory observability records only aggregate action counters; zero raw memory statements are persisted in monitoring tables.

---

## 16. Verification & Test Results
- **Go Unit Tests** (`backend/internal/monitoring/...`): **10/10 PASSED** (Coverage: pricing calculations, threshold health evaluation, redaction, and catalog merging).
- **Python Integration Suite** (`ai_sidecar/test_phase2_task211_monitoring.py`): **31/31 PASSED (100%)**
  - Trace persistence and retrieval
  - Multi-tenant boundary enforcement
  - Health summary calculation and threshold triggers
  - Token and cost estimation accuracy
  - Recommendation and approval metrics aggregation
  - Memory metadata safety
  - Security incident logging and error redaction
- **Python Regression Suite** (`ai_sidecar/test_phase2_task210_memory.py`): **13/13 PASSED (100%)**
- **Frontend Vitest Suite**: **43 test files, 243 tests PASSED (100%)**
  - Includes dedicated `AIMonitoringDashboard.test.jsx` (2/2 passed)
- **Production Build**: `npm run build` completed successfully in 15.70s with zero errors.

---

## 17. Build & Integrity Confirmation
- No fake or hardcoded mock data was created.
- No database tables were reset, truncated, or dropped.
- No existing business records were modified or deleted.
- All monitoring data is persisted in real MariaDB tables.
- The UI matches the existing LogisticsHQ light/white theme without any dark/black AI panels.
