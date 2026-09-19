# Phase 3 Task 3.8: Event-Driven AI Workflows and Cross-Module Automation

## Executive Summary
**Phase 3 Task 3.8: Event-Driven AI Workflows and Cross-Module Automation** has been successfully designed, implemented, integrated, and verified for LogisticsHQ. 

This implementation provides an enterprise-grade, event-driven automation layer that responds to business events across core operational and commercial modules (Shipments, Finance, Invoicing, Contracts, Compliance, Leads, and RFQ-to-Quotation). The architecture rigorously enforces the project's core boundary: **ALL AI reasoning, prompt orchestration, cross-module synthesis, classification, recommendations, and communication drafting are implemented strictly in Python (`ai_sidecar/app/event_workflows/`)**, while **Go remains the sole integration and application-control layer** governing authentication, organization isolation, deterministic eligibility, event deduplication, database transactions, idempotency, HITL approval enforcement, and audit logging.

---

## 1. Objective
The objective of Task 3.8 is to build an event-driven automation pipeline capable of:
1. Ingesting domain business events across LogisticsHQ modules.
2. Deduplicating repeated, delayed, or out-of-order events.
3. Applying deterministic backend eligibility rules in Go before invoking AI.
4. Aggregating authorized cross-module business context from MariaDB.
5. Invoking Python AI sidecar models for multi-domain reasoning, risk evaluation, root cause analysis, and actionable next-step recommendations.
6. Enforcing Human-In-The-Loop (HITL) approval gating for consequential operational, commercial, or external communication actions via the Centralized Action System and Approvals Center.
7. Providing complete auditability, persistence, and an intuitive user interface adhering to LogisticsHQ's light theme.

---

## 2. Existing Architecture Inspected
Before making changes, the following existing components were inspected and reused:
- **MariaDB Schema & Migration Infrastructure**: Checked `backend/internal/database/migrations/` and existing table structures (`shipments`, `invoices`, `contracts`, `leads`, `rfqs`, `action_proposals`, `approval_requests`).
- **Python AI Sidecar Runtime**: Inspected `ai_sidecar/main.py`, virtual environment dependencies (`langchain`, `pydantic`, `fastapi`), and existing operational agents (`pricing_agent.py`, `operations_agent.py`, `collections_agent.py`, `contract_compliance_agent.py`).
- **Centralized Action System & Approvals**: Inspected `backend/internal/approvals/service.go` and `backend/internal/orchestration/registry.go` to seamlessly route high-risk recommendations into real `approval_requests` rows.
- **Workflow Automations Frontend**: Inspected `frontend/src/pages/dashboard/Automations/WorkflowAutomationsPage.jsx` and `ActionOrchestrationTab.jsx` to ensure visual and architectural consistency.

---

## 3. Files Created & Changed

### Database Migrations
- [`backend/internal/database/migrations/105_phase3_event_driven_ai_workflows.sql`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/database/migrations/105_phase3_event_driven_ai_workflows.sql) (Applied to MariaDB)
- [`backend/migrations/105_phase3_event_driven_ai_workflows.sql`](file:///c:/Users/Sai/go/src/freel-project/backend/migrations/105_phase3_event_driven_ai_workflows.sql)

### Python AI Sidecar (`ai_sidecar/app/event_workflows/`)
- [`ai_sidecar/app/event_workflows/__init__.py`](file:///c:/Users/Sai/go/src/freel-project/ai_sidecar/app/event_workflows/__init__.py)
- [`ai_sidecar/app/event_workflows/models.py`](file:///c:/Users/Sai/go/src/freel-project/ai_sidecar/app/event_workflows/models.py): Pydantic input/output schemas.
- [`ai_sidecar/app/event_workflows/agent.py`](file:///c:/Users/Sai/go/src/freel-project/ai_sidecar/app/event_workflows/agent.py): `EventWorkflowsAgent` featuring multi-domain cross-module reasoning, draft generator, and prompt-injection defenses.
- [`ai_sidecar/main.py`](file:///c:/Users/Sai/go/src/freel-project/ai_sidecar/main.py): Registered `/event-workflows/analyze-event` and `/event-workflows/generate-draft`.
- [`ai_sidecar/tests/test_event_workflows.py`](file:///c:/Users/Sai/go/src/freel-project/ai_sidecar/tests/test_event_workflows.py): 7/7 unit tests verifying Pydantic validation, sanitization, missing-info detection, and cross-module synthesis.

### Go Backend Integration Layer (`backend/internal/event_workflows/`)
- [`backend/internal/event_workflows/model.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/event_workflows/model.go): Domain records, `RawJSON` custom sql.Scanner, workflow instance structs.
- [`backend/internal/event_workflows/repository.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/event_workflows/repository.go): MariaDB persistence, deduplication queries, cross-module context collectors.
- [`backend/internal/event_workflows/service.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/event_workflows/service.go): Deterministic eligibility engine, sidecar caller, approval router, retry/cancel management.
- [`backend/internal/event_workflows/handler.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/event_workflows/handler.go): Authenticated HTTP handlers with organization context enforcement.
- [`backend/internal/event_workflows/service_test.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/event_workflows/service_test.go): Unit tests for eligibility, deduplication, and workflow lifecycle transitions.
- [`backend/internal/server/server.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/server/server.go): Route registration.
- [`backend/cmd/server/main.go`](file:///c:/Users/Sai/go/src/freel-project/backend/cmd/server/main.go): Service & handler initialization.

### Frontend Integration (`frontend/src/`)
- [`frontend/src/services/eventWorkflowsService.js`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/services/eventWorkflowsService.js): API client for event workflows.
- [`frontend/src/pages/dashboard/Automations/EventWorkflowsTab.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/Automations/EventWorkflowsTab.jsx): Interactive UI tab with KPI cards, workflow table, detail drawer, event ledger, and one-click event simulator.
- [`frontend/src/pages/dashboard/Automations/EventWorkflowsTab.css`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/Automations/EventWorkflowsTab.css): LogisticsHQ light theme styling.
- [`frontend/src/pages/dashboard/Automations/WorkflowAutomationsPage.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/Automations/WorkflowAutomationsPage.jsx): Mounted tab into main Automations page.
- [`frontend/src/__tests__/pages/Automations/EventWorkflowsTab.test.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/__tests__/pages/Automations/EventWorkflowsTab.test.jsx): 5/5 Vitest unit tests.

---

## 4. Python AI Components Implemented
Implemented in `ai_sidecar/app/event_workflows/`:
1. **`EventWorkflowRequest` & `EventAnalysisResponse`**: Strict Pydantic models enforcing typing for event metadata, raw payloads, cross-module context, and confidence scoring.
2. **`EventWorkflowsAgent`**:
   - Multi-domain event classification and urgency scoring (`LOW`, `MEDIUM`, `HIGH`, `CRITICAL`).
   - Cross-module context synthesis (e.g. correlating shipment delays with invoice credit terms or contract SLAs).
   - Missing-information detection (e.g., detecting missing revised ETAs, missing proof of delivery, or missing revised credit limits).
   - Structured recommendation generation with explicit risk levels (`LOW`, `MEDIUM`, `HIGH`, `CRITICAL`) and `requires_approval` flags.
   - Prompt-injection defense that neutralizes control characters and adversarial overrides embedded in event notes.
3. **`WorkflowDraftResponse` Generator**: Produces customer update, carrier inquiry, or internal escalation drafts with explicit draft markers.

---

## 5. Go Integration & Application Control Layer
Implemented in `backend/internal/event_workflows/`:
1. **Tenant & Organization Isolation**: Every query (`SELECT`, `INSERT`, `UPDATE`) strictly enforces `WHERE org_id = ?`. Cross-tenant queries return 404 Not Found.
2. **Deterministic Eligibility Rules**: Evaluated in Go before any sidecar call:
   - `shipment.milestone_missed` / `shipment.exception_created` -> `SHIPMENT_EXCEPTION_RESPONSE`
   - `invoice.overdue` / `invoice.due_soon` -> `INVOICE_COLLECTION_ESCALATION`
   - `contract.approaching_expiry` / `document.rejected` -> `CONTRACT_COMPLIANCE_RENEWAL`
   - `lead.created` / `lead.qualified` -> `LEAD_FOLLOWUP`
   - `rfq.submitted` -> `RFQ_QUOTATION_DISPATCH`
   - Ineligible events are marked `IGNORED` with zero LLM overhead.
3. **Event Ingestion & Deduplication**: Generates a deterministic deduplication key:
   `evt:{org_id}:{event_type}:{source_record_type}:{source_record_id}:{correlation_id}`
   Duplicate events return the existing workflow instance and do not trigger duplicate AI evaluations or duplicate actions.
4. **Approval & Action Gating**: If any AI recommendation specifies `requires_approval = true` or `risk_level in ("HIGH", "CRITICAL")`, Go halts automatic execution, sets the workflow status to `AWAITING_APPROVAL`, and creates a pending approval record in `approval_requests`.
5. **MariaDB Persistence**: Handled via `ai_event_store` and `ai_cross_module_workflows`. Custom type `RawJSON` safely scans NULL values in JSON columns.

---

## 6. Supported Event Types and Workflows

| Domain | Event Type | Workflow Type | Action Intent | Approval Required? |
|---|---|---|---|---|
| **Shipments** | `shipment.milestone_missed` | `SHIPMENT_EXCEPTION_RESPONSE` | `shipments.notify_delay` | **Yes (High Risk)** |
| **Shipments** | `shipment.exception_created` | `SHIPMENT_EXCEPTION_RESPONSE` | `shipments.carrier_escalation` | **Yes (High Risk)** |
| **Finance** | `invoice.overdue` | `INVOICE_COLLECTION_ESCALATION` | `finance.send_dunning_reminder` | **Yes (High Risk)** |
| **Finance** | `invoice.due_soon` | `INVOICE_COLLECTION_ESCALATION` | `finance.courtesy_notice` | Low Risk / Internal |
| **Contracts** | `contract.approaching_expiry` | `CONTRACT_COMPLIANCE_RENEWAL` | `contracts.initiate_renewal` | **Yes (High Risk)** |
| **Contracts** | `document.rejected` | `CONTRACT_COMPLIANCE_RENEWAL` | `documents.request_reupload` | Low Risk / Internal |
| **CRM / Leads**| `lead.qualified` | `LEAD_FOLLOWUP` | `leads.schedule_call` | Low Risk / Internal |
| **Pricing** | `rfq.submitted` | `RFQ_QUOTATION_DISPATCH` | `rfq.draft_quotation` | **Yes (High Risk)** |

---

## 7. Workflow Lifecycle
1. **Received**: Ingested domain event authenticated and scoped to current `org_id`.
2. **Deduplication Checked**: Query MariaDB `ai_event_store` using unique dedup key. If duplicate, return existing workflow.
3. **Eligibility Evaluated**: Deterministic business rules in Go determine if the event warrants cross-module AI processing.
4. **Running**: Workflow initialized in `ai_cross_module_workflows`.
5. **Context Aggregated**: Go retrieves authorized cross-module customer, shipment, or invoice context.
6. **Sidecar Analysis**: Sidecar calculates urgency, synthesizes cross-module signals, and suggests actions.
7. **Approval Gated**: If action has high risk, set status to `AWAITING_APPROVAL` and insert into `approval_requests`.
8. **Completed / Cancelled / Retried**: Fully auditable with timestamps and error diagnostics.

---

## 8. API Endpoints Implemented

| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `GET` | `/api/v1/event-workflows/overview` | Overall KPIs (ingested, active, awaiting approval, deduplicated) | Yes (`org_id`) |
| `GET` | `/api/v1/event-workflows/events` | List ingested domain events | Yes (`org_id`) |
| `GET` | `/api/v1/event-workflows/events/:id` | Get single event record with raw payload | Yes (`org_id`) |
| `GET` | `/api/v1/event-workflows/workflows` | List cross-module AI workflows | Yes (`org_id`) |
| `GET` | `/api/v1/event-workflows/workflows/:id` | Get detailed workflow record & recommendations | Yes (`org_id`) |
| `POST` | `/api/v1/event-workflows/workflows/:id/retry` | Retry failed or stalled workflow instance | Yes (`org_id`) |
| `POST` | `/api/v1/event-workflows/workflows/:id/cancel` | Cancel active or awaiting-approval workflow | Yes (`org_id`) |
| `POST` | `/api/v1/event-workflows/simulate-event` | Ingest and evaluate domain business event | Yes (`org_id`) |

---

## 9. Verification & Testing Results

### 1. Python AI Sidecar Tests (`pytest`)
Ran `pytest tests/test_event_workflows.py -v`:
- `test_shipment_exception_event_analysis`: **PASSED**
- `test_invoice_overdue_event_analysis`: **PASSED**
- `test_contract_expiring_event_analysis`: **PASSED**
- `test_lead_created_event_analysis`: **PASSED**
- `test_workflow_draft_generation`: **PASSED**
- `test_prompt_injection_sanitization`: **PASSED**
- `test_missing_info_detection`: **PASSED**
**Result: 7/7 tests passed (100%) in 0.10s.**

### 2. Go Unit Tests
Ran `go test -v ./internal/event_workflows/...`:
- `TestEligibilityEvaluation`: **PASS**
- `TestEventDeduplicationAndIdempotency`: **PASS**
- `TestWorkflowInstanceLifecycleState`: **PASS**
**Result: 3/3 tests passed (100%) in 0.47s.**

### 3. End-to-End Live Integration Verification (`verify_phase3_task38_live.py`)
Tested against running daemons (Go server on port 8080, Sidecar on port 8090, MariaDB):
- **Sidecar Direct**: `/event-workflows/analyze-event` (200 OK), `/event-workflows/generate-draft` (200 OK).
- **Go Endpoints**: `/overview` (200 OK), `/simulate-event` milestone delay (201 Created), `/simulate-event` invoice overdue (201 Created).
- **Deduplication Check**: Re-sent identical event; verified exact existing workflow ID returned with no duplicate database insertions.
- **Tenant Isolation**: Cross-tenant query by Organization 2 for Organization 1's workflow returned 404 Not Found.
- **MariaDB Persistence**: Verified 3 rows in `ai_event_store`, 3 rows in `ai_cross_module_workflows`, and pending Approval Request ID #210 created in `approval_requests`.
**Result: 100% Passed.**

### 4. Frontend Unit Tests (`vitest`)
Ran `vitest run src/__tests__/pages/Automations/EventWorkflowsTab.test.jsx`:
- `renders KPI overview counters accurately`: **PASS**
- `renders workflows list with governance badges and summaries`: **PASS**
- `opens detail drawer when clicking Details button`: **PASS**
- `switches to Event Store Ledger sub-tab`: **PASS**
- `opens event simulation modal and triggers simulation`: **PASS**
**Result: 5/5 tests passed (100%) in 22.78s.**

### 5. Frontend Production Build (`npm run build`)
Ran `vite build`:
- Cleanly compiled 3,155 modules into `dist/` without errors or bundle warnings.

---

## 10. Compliance with Core Directives
1. **ALL AI CODE IN PYTHON ONLY**: All intelligence, agent prompts, Pydantic schemas, and reasoning reside exclusively in `ai_sidecar/app/event_workflows/`. Zero AI or LLM logic was written in Go or JavaScript.
2. **Go Integration Layer Only**: Go governs tenant isolation, MariaDB transactions, eligibility evaluation, and approval routing.
3. **No Direct DB Mutations by Sidecar**: Sidecar never touches SQL, never mutates business records, and never sends messages.
4. **Approval Required for Consequential Actions**: External notifications and escalations require human approval via `approval_requests`.
5. **Real Persistent Data Preserved**: MariaDB `freel_mysql` was preserved without table drops or destructive seeds.
6. **Consistent UI Design**: The interface uses crisp white cards, standard navy sidebar styling, standard Lucide icons, and zero dark or glowing neon widgets.

---

## 11. Final Implementation Status
**Status: COMPLETED (100% Tested & Verified)**.
Phase 3 Task 3.8 is fully operational and integrated into the LogisticsHQ enterprise platform.
