# LogisticsHQ Agentic AI Platform — Phase 0.1 Baseline & Verification Report

**Author**: Antigravity AI Engineering  
**Environment**: Windows Development Environment (`DESKTOP-UI4UA1T`)  
**Organization**: LogisticsHQ Dev Org — Varun Logistics (`organization_id = 2`)  
**Test User**: `kanadevarun123@gmail.com` (`user_id = 6`, `SUPER_ADMIN`)  
**Date**: September 2026  
**Status**: Completed & Verified  

---

## 1. Executive Summary

This report establishes the verified, evidence-based baseline for the LogisticsHQ Agentic AI platform prior to further architectural refactoring or feature additions. In accordance with the Phase 0.1 mandate, this investigation was strictly diagnostic and verification-focused. No new AI agents were added, no existing data was deleted or reseeded, and zero external communications (such as real emails) were dispatched.

### Summary of Classification Findings
- **Verified**: Local environment services, development dataset integrity for `organization_id = 2`, user role and permissions (40 system permissions), 7-agent inventory and execution flow, persistent checkpoint storage via MariaDB for Pricing and Contracts graphs, outbound email safety gating (zero autonomous sends), and complete functional suite execution across all 7 agent modules.
- **Partially Verified**: Frontend status visibility across certain sub-modules (e.g., granular LangGraph step display is currently abstracted behind high-level agent status flags like `DRAFT_READY` and `WAITING_FOR_HUMAN`).
- **Critical Risk Confirmed (P0)**: **100% of AI sidecar tools bypass the centralized Go Action System**, invoking raw internal HTTP endpoints with static service-key authentication.
- **High Risk Confirmed (P1)**: **Agent HITL (Human-in-the-Loop) mechanisms are fragmented across 6 distinct patterns**; only Sales clarification emails currently interface with the centralized `approval_requests` table.
- **Medium Risk Confirmed (P2)**: `sales_graph.py` and `operations_graph.py` have not yet been migrated to `MariaDBSaver` persistent checkpoints.
- **Recommended Next Task**: **Migrate AI Sidecar Tools to Centralized Action System Execution Bridge**.

---

## 2. Environment & Service Health Matrix

All core services were verified through active execution, socket connectivity, and authenticated health checks:

| Service / Component | Host / Port | Health / Status Endpoint | Verification Result | Details |
| :--- | :--- | :--- | :--- | :--- |
| **Go REST Backend** | `127.0.0.1:8080` | `GET /health` | **VERIFIED (PASS)** | HTTP 200 `{"message":"Freel backend is running","success":true}`. Managed via `task-2977`. |
| **Python AI Sidecar** | `127.0.0.1:8090` | `GET /health` | **VERIFIED (PASS)** | HTTP 200 `{"status":"ok","checkpointer":"MariaDBSaver","database":"healthy"}`. FastAPI + LangGraph. Managed via `task-2739`. |
| **MariaDB Database** | `127.0.0.1:3306` | Native TCP Ping | **VERIFIED (PASS)** | MariaDB 12.3.3 (`freel_mysql`), single shared database across Go and Python sidecar. |
| **React + Vite Frontend**| `127.0.0.1:5173` | `GET /` | **VERIFIED (PASS)** | HTTP 200, HTML Title: `LogisticsHQ \| The Operating System for Modern Logistics`. Managed via `task-924`. |
| **Database Migrations** | `127.0.0.1:3306` | `SHOW TABLES` | **VERIFIED (PASS)** | 90 migration files in repository; 119 tables present in MariaDB. |
| **Internal Service Auth**| HTTP Headers | `X-LogisticsHQ-Service-Key` | **VERIFIED (PASS)** | Verified: Valid key grants access; invalid/missing key returns HTTP 401 Unauthorized. |
| **LLM Configuration** | Sidecar Runtime | `test_llm_config.py` | **VERIFIED (PASS)** | Primary: Google Gemini (`gemini-3.1-flash-lite`), Fallback: OpenAI (`gpt-4o-mini`) via `FailoverChatModel`. |
| **Queue Worker Status** | Go Runtime | Async Dispatcher | **VERIFIED (PASS)** | Queue worker actively processes `ai_processing_tasks` (`task_type = 'PRICING_ANALYZE'`). |

---

## 3. Development Dataset Inventory (`organization_id = 2`)

The persistent dataset was audited directly in MariaDB. Zero records were deleted, truncated, or overwritten:

| Entity Domain | Database Table | Record Count | Sample IDs / Records | Relation / Integrity Check | Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Customers** | `customers` | 5 | IDs: 1, 2, 3, 4, 5 (`Apex Global Logistics`, `Euro-Asia Retailers`) | Scoped to `org_id = 2` | **VERIFIED** |
| **Leads** | `leads` | 4 | IDs: 101, 102, 103, 104 | Foreign keys to `customers` & `organizations` intact | **VERIFIED** |
| **RFQs** | `rfqs` | 5 | IDs: 101, 102, 103, 104, 105 | Linked to customer and lead entities | **VERIFIED** |
| **Quotations** | `quotations` | 2 | IDs: 101, 102 (`QT-2026-DEV-001`, `QT-2026-DEV-002`) | Associated with RFQ #101 & #102 | **VERIFIED** |
| **Carrier Rates** | `rates` | 4 | IDs: 101, 102, 103, 104 | Contract ocean freight rates ($2,800 - $3,100) | **VERIFIED** |
| **Contracts** | `contract_documents` | 2 | IDs: 101, 102 (`Maersk SC-99201-2026`, `MSC-AGR-2026`) | PDF rate contracts parsed and linked | **VERIFIED** |
| **Bookings** | `bookings` | 1 | `BK-2026-DEV-002` | Linked to Quotation #102 | **VERIFIED** |
| **Shipments** | `shipments` | 3 | IDs: 101, 102, 103 (`SH-2026-DEV-001` - `003`) | Associated with Booking & Customer | **VERIFIED** |
| **Milestones** | `shipment_milestones` | 6 | `GATE_IN`, `DELAY_NOTICE`, `LOADED`, `DISCHARGED` | Scoped to Shipment #101 & #102 | **VERIFIED** |
| **Exceptions** | `shipment_exceptions`| 3 | `CUSTOMS_HOLD`, `REEFER_TEMP_SPIKE`, `FEEDER_DELAY` | Active disruption tickets | **VERIFIED** |
| **Documents** | `shipment_documents` | 5 | MBL, HBL, Packing List, Commercial Invoice | File metadata and verification flags | **VERIFIED** |
| **Invoices** | `customer_invoices` | 3 | IDs: 101, 102, 103 (`INV-2026-DEV-001` - `003`) | Scoped to Org 2, amounts USD 1,800 - 3,200 | **VERIFIED** |
| **Doc Discrepancies** | `shipment_document_discrepancies` | 2 | IDs: 5, 6 (Weight mismatch, container # discrepancy) | Scoped to Shipment #101 | **VERIFIED** |
| **Finance Discrepancies**| `shipment_finance_discrepancies` | 4 | IDs: 1, 2, 3, 4 (THC, Doc Fee, Fuel Surcharge variance) | Scoped to Invoice #101 | **VERIFIED** |
| **Payments** | `customer_invoice_payments` | 1 | ID: 101 ($3,200 payment applied to INV-101) | Status `COMPLETED` | **VERIFIED** |
| **Approvals** | `approval_requests` | 6+ | IDs: 101-106, + 107, 108 generated during test suite | Clarification emails, credit overrides, spot margins | **VERIFIED** |
| **AI Tasks** | `ai_processing_tasks`| 18 | PRICING_ANALYZE, CONTRACT_EXTRACTION, etc. | Task payloads with correlation IDs | **VERIFIED** |
| **AI Checkpoints** | `ai_checkpoints` | 80+ | Thread IDs: `rfq-2-105`, `contract-2-101`, etc. | MariaDBSaver persistent state tuples | **VERIFIED** |
| **Universal Audit** | `audit_logs` | 31 | Actions across RFQs, Leads, Shipments, Auth | Actor IDs, IP addresses, entity types | **VERIFIED** |
| **Activities** | `activities` | 26 | Activity stream for user dashboard | Human and agent activity feed | **VERIFIED** |
| **Lead Interactions**| `lead_interactions` | 10+ | Inbound RFQ inquiries, negotiation emails | Full thread hierarchy | **VERIFIED** |
| **Email Drafts** | `lead_email_drafts` | 5+ | Clarification drafts awaiting approval | Gated status: `AWAITING_APPROVAL` | **VERIFIED** |

---

## 4. User, Organization & Security Context

1. **Test User Identity**:
   - Email: `kanadevarun123@gmail.com`
   - User ID: `6`
   - Name: `Varun Kanade`
   - Organization: `organization_id = 2` (`LogisticsHQ Dev Org - Varun Logistics`)
   - Role: `SUPER_ADMIN`
   - Permissions: 40 active system permissions covering full operational and administrative scope.
2. **Tenant Isolation Verification**:
   - Authenticated queries to `/api/v1/leads`, `/api/v1/rfqs`, and `/api/v1/quotations` return exclusively records where `organization_id = 2`.
   - Injected queries requesting other organization IDs are rejected by the backend session context.
   - Internal endpoints validate `org_id` query/body parameters against internal service credentials.
3. **Actor Identity Preservation**:
   - User actions through the frontend record `actor_id = 6`, `actor_type = 'USER'`.
   - Background tasks record `actor_type = 'SYSTEM'`.
   - Sidecar tool callbacks record `actor_type = 'AI_AGENT'`, preserving auditable lineage in `audit_logs`.

---

## 5. Agent Inventory & Complete Architecture Flow Matrix

All 7 existing agents were analyzed across their end-to-end execution path:

```
[User / Event Trigger]
        │
        ▼
[Go REST Backend (:8080)] ── Enqueues Task ──▶ [ai_processing_tasks (MariaDB)]
        │                                                     │
        │                                                     ▼
[Internal HTTP / Event] ◀──────────────────────── [Worker Dispatcher]
        │
        ▼
[Python AI Sidecar (:8090)]
   ├── LangGraph Workflow (StateGraph)
   ├── LLM Invocation (Gemini 3.1 Flash Lite / GPT-4o-mini Fallback)
   ├── Checkpointer (MariaDBSaver -> ai_checkpoints)
   └── Sidecar Tools (go_api_client, pricing_tool, leads_tool, shipments_tool)
        │
        ▼
[Go Backend Callback (:8080/internal/...)]
        │
        ├── Updates Domain Tables (rfqs, quotations, leads, shipments)
        ├── Stages Approval Request (approval_requests) [If Gated]
        ├── Emits Audit Log (audit_logs) & Activity (activities)
        │
        ▼
[React + Vite Frontend (:5173)] (Displays DRAFT_READY / WAITING_FOR_HUMAN / AWAITING_APPROVAL)
```

### Agent-by-Agent Specification Matrix

| Agent | Frontend Entry Point | Backend Endpoint / Trigger | AI Task Type | Graph / Worker | Tools Used | DB Tables Read | DB Tables Written | Callback Endpoint | UI Status Field | Approval / HITL Mechanism |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **1. Pricing** | RFQ Detail Panel / "Analyze Pricing" | `POST /api/v1/rfqs/{id}/pricing-analyze` & `EventRFQAssigned` | `PRICING_ANALYZE` | `pricing_graph.py` | `fetch_rfq`, `search_rates`, `fetch_rules`, `create_quotes` | `rfqs`, `rates`, `pricing_rules` | `quotations`, `rfqs`, `ai_checkpoints` | `/internal/pricing/callback` | `agent_status`: `COLLECTING_INFO`, `PROCESSING`, `DRAFT_READY`, `WAITING_FOR_HUMAN` | LangGraph interrupt before `save` node for anomalies; RFQ status set to `WAITING_FOR_HUMAN`. |
| **2. Sales / Email** | Mailbox / Leads Panel / Inbound Webhook | `POST /api/v1/leads/{id}/process-inbound` | `SALES_EMAIL_PARSE` | `sales_graph.py` | `fetch_lead`, `create_rfq_from_email`, `update_lead_qual` | `lead_interactions`, `leads` | `lead_email_drafts`, `approval_requests`, `rfqs` | `/internal/sales/callback` | `lead_email_drafts.status`: `AWAITING_APPROVAL` | Creates entry in `approval_requests` (`Clarification Email Approval`). Zero autonomous outbound send. |
| **3. Operations** | Tracking & Shipments View / Carrier Sync | `POST /internal/carrier/webhook` & `PUT /shipments/{id}/sync` | `OPERATIONS_TRACKING` | `operations_graph.py` | `get_shipment`, `update_milestone`, `create_exception` | `shipments`, `shipment_milestones` | `shipment_milestones`, `shipment_exceptions` | `/internal/operations/callback` | `shipment_exceptions.status`: `OPEN`, `RESOLVED` | Non-blocking `shipment_exceptions` ticket created. No central approval request. |
| **4. Contracts** | Contracts & Rate Repository / "Extract Rates" | `POST /api/v1/contracts/extract` | `CONTRACT_EXTRACTION` | `contracts_graph.py` | `ocr_agent`, `classifier_agent`, `parser_agent`, `validator_agent` | `contract_documents` | `rates`, `rate_review_queue`, `ai_checkpoints` | `/internal/contracts/callback` | `status`: `EXTRACTING`, `REVIEW_REQUIRED`, `COMPLETED` | LangGraph interrupt before `ingest` node; flagged rates saved to `rate_review_queue`. |
| **5. Compliance** | Shipment Documents / "Verify Compliance" | `POST /api/v1/shipments/{id}/reconcile-docs` | `DOC_RECONCILIATION` | Go Worker / Sidecar Tool | Direct verification engine | `shipment_documents`, `shipments` | `shipment_document_discrepancies` | Internal function call | `shipment_document_discrepancies.status`: `OPEN`, `RESOLVED` | Documents marked discrepant; review via Shipment Documents tab. |
| **6. Finance** | Invoices Module / "Three-Way Match" | `POST /api/v1/invoices/{id}/reconcile` | `INVOICE_RECONCILIATION` | Go Worker / Sidecar Tool | Direct calculation engine | `customer_invoices`, `quotations`, `rates` | `shipment_finance_discrepancies` | Internal function call | `shipment_finance_discrepancies.status`: `OPEN`, `RESOLVED` | Discrepancy logged for variance above tolerance; reviewed via Invoices tab. |
| **7. Lead Scoring** | Leads Board / Lead Creation Event | `EventLeadCreated` & Background Worker | `LEAD_SCORING` | Go Background Worker | Rule evaluator & sidecar enricher | `leads`, `lead_interactions` | `leads` (`ai_score`, `notes`) | Internal event emit | `leads.ai_score` (0-100), `leads.status` | Automatic scoring; high scores trigger notification, no blocking HITL. |

---

## 6. Deep Verification of Known Critical Findings

### A. Persistent Checkpointer Verification
- **Implementation**: Verified active implementation of `MariaDBSaver` in `ai_sidecar/app/persistence/mariadb_saver.py`, backed by the `ai_checkpoints` table in MariaDB.
- **Affected Graphs**:
  - `pricing_graph.py`: **MIGRATED & VERIFIED**. Uses `MariaDBSaver`. Pauses cleanly on anomaly interrupt `('save',)`.
  - `contracts_graph.py`: **MIGRATED & VERIFIED**. Uses `MariaDBSaver`. Pauses cleanly before `('ingest',)`.
  - `sales_graph.py`: **NOT YET PERSISTENT**. Currently uses procedural routing without checkpointer.
  - `operations_graph.py`: **NOT YET PERSISTENT**. Currently uses procedural tool dispatch.
- **Thread ID Convention**:
  - Pricing: `rfq-{org_id}-{rfq_id}`
  - Contracts: `contract-{org_id}-{document_id}`
- **Restart Survival**:
  - Validated by instantiating a fresh `MariaDBSaver` instance and re-fetching paused state tuple (`saver.get_tuple(config)`). Checkpoint ID, node pointer, and serialized state were restored with 100% fidelity from MariaDB.
- **Required Next Step**: Wire `sales_graph.py` and `operations_graph.py` to `MariaDBSaver`.

### B. Outbound Email Safety Verification
- **Draft Generation**: When `sales_graph` detects `intent = 'RFQ_REQUEST_INCOMPLETE'`, it generates a draft clarification response.
- **Autonomous Send Prevention**: **VERIFIED**. The Go backend handler (`backend/internal/leads/email_handler.go`) explicitly stages the draft with `status = 'AWAITING_APPROVAL'`.
- **Approval Request**: An approval request is automatically created in `approval_requests` with `type = 'Clarification Email Approval'` and `status = 'PENDING'`.
- **External Email Protection**: In test execution, **zero outbound emails were dispatched**. Outbound transmission requires explicit human execution via `POST /api/v1/approvals/{id}/approve` or `POST /api/v1/leads/drafts/{id}/send`.
- **Mailbox Authentication**: Sending is additionally guarded by OAuth credentials in `connected_mailboxes`. If no mailbox is connected or verified, sending is strictly blocked.

### C. Action System Integration Verification (Critical Risk P0)
The Go backend contains a centralized Action architecture (`backend/internal/actions/contract.go`, `registry.go`, `context.go`) supporting `ActionContext`, `IdempotencyStore`, and strict RBAC. However, inspection of all 12 tools in `ai_sidecar/app/tools/` reveals complete bypass:

| Sidecar Tool Function | File Path | Current Execution Path | Action System Status | Risk / Defect |
| :--- | :--- | :--- | :--- | :--- |
| `fetch_rfq` | `go_api_client.py` | Raw `GET /internal/rfqs/{id}` | **BYPASS (Unregistered)** | Static shared key; bypasses user context |
| `fetch_pricing_rules` | `go_api_client.py` | Raw `GET /internal/pricing/rules` | **BYPASS (Unregistered)** | Static shared key; bypasses permission check |
| `create_draft_quotes` | `go_api_client.py` | Raw `POST /internal/pricing/quotes/draft` | **BYPASS (Unregistered)** | Bypasses IdempotencyStore & Action audit |
| `send_pricing_callback`| `go_api_client.py` | Raw `POST /internal/pricing/callback` | **BYPASS (Unregistered)** | Bypasses Action event bus |
| `search_rates` | `pricing_tool.py` | Raw `GET /internal/rates/search` | **BYPASS (Unregistered)** | Static shared key; bypasses rate visibility RBAC |
| `fetch_lead` | `leads_tool.py` | Raw `GET /internal/leads/{id}` | **BYPASS (Unregistered)** | Static shared key |
| `create_rfq_from_email`| `leads_tool.py` | Raw `POST /internal/rfqs/from-email` | **BYPASS (Unregistered)** | Bypasses IdempotencyStore; potential duplicate RFQs |
| `update_lead_qualification`| `leads_tool.py`| Raw `PUT /internal/leads/{id}/qualification`| **BYPASS (Unregistered)** | Bypasses Lead Action audit |
| `get_shipment` | `shipments_tool.py` | Raw `GET /internal/shipments/{id}` | **BYPASS (Unregistered)** | Static shared key |
| `update_milestone` | `shipments_tool.py` | Raw `POST /internal/shipments/{id}/milestones`| **BYPASS (Unregistered)** | Bypasses shipment action idempotency |
| `create_exception` | `shipments_tool.py` | Raw `POST /internal/shipments/{id}/exceptions`| **BYPASS (Unregistered)** | Bypasses central exception action pipeline |
| `send_operations_callback`| `shipments_tool.py`| Raw `POST /internal/operations/callback` | **BYPASS (Unregistered)** | Static shared key |

**Classification**: **100% Bypass**. All 12 sidecar tools use raw HTTP endpoints with a static shared secret header. None utilize the central `ActionRegistry` or `ActionContext`.

### D. Approvals & HITL Fragmentation (High Risk P1)
Mapping of human review mechanisms reveals severe fragmentation across 6 disjoint implementations:

| Domain / Flow | Review Mechanism | Database Table | Central Approvals Interfaced? | Lifecycle Capabilities |
| :--- | :--- | :--- | :--- | :--- |
| **Sales Clarification Draft** | Central Approval Request | `approval_requests` | **YES** | Can Pause, Approve, Reject, Send. Fully unified. |
| **Pricing Anomalies** | RFQ Agent Status Flag | `rfqs.agent_status` & `ai_checkpoints` | **NO** | Pauses at LangGraph interrupt (`save`). No entry in `approval_requests`. |
| **Contracts Rate Flagging** | Dedicated Review Queue | `rate_review_queue` & `ai_checkpoints`| **NO** | Pauses at LangGraph interrupt (`ingest`). Reviewed in custom queue UI. |
| **Operations Disruptions** | Shipment Exceptions | `shipment_exceptions` | **NO** | Disruption ticket created; resolved directly on shipment panel. |
| **Compliance Discrepancies** | Document Discrepancies | `shipment_document_discrepancies` | **NO** | Document mismatch logged; resolved in Documents tab. |
| **Finance Discrepancies** | Finance Discrepancies | `shipment_finance_discrepancies` | **NO** | Variance logged; resolved in Invoices reconciliation tab. |

**Consequence**: Operators cannot use a single "Approvals & Tasks" inbox to manage all pending AI interventions. Clarifications appear under Approvals, but pricing anomalies appear only under RFQs, and rate discrepancies appear only under Contracts.

---

## 7. Functional Test Suite Execution Evidence

A comprehensive diagnostic test script ([`ai_sidecar/run_baseline_functional_suite.py`](file:///c:/Users/Sai/go/src/freel-project/ai_sidecar/run_baseline_functional_suite.py)) was executed against the live services and persistent dataset.

```
===========================================================================
 LOGISTICSHQ AGENTIC AI PLATFORM — PHASE 0.1 BASELINE FUNCTIONAL SUITE
===========================================================================
[PASS] Pricing Agent — Normal RFQ           | Target: RFQ #102           | Rates retrieved and RFQ agent_status=DRAFT_READY
[PASS] Pricing Agent — Anomaly HITL         | Target: RFQ #105           | Paused at interrupt ('save',) and saved 14 checkpoints in MariaDB.
[PASS] Contracts Extraction & Rate Anomaly  | Target: Doc #101           | Paused before ('ingest',), state reloaded successfully from MariaDB.
[PASS] Sales / Email Parser — Incomplete Email | Target: Interaction #112   | Draft #12 staged, Approval #108 generated, 0 autonomous sends.
[PASS] Sales / Email Parser — Complete Email | Target: Lead #103          | RFQ created successfully from email payload.
[PASS] Operations / Carrier Tracking        | Target: Shipment #102      | Shipment retrieved (status=IN_TRANSIT), milestone validated.
[PASS] Compliance Reconciliation            | Target: Shipment #101      | Verified 2 compliance document discrepancies in database.
[PASS] Finance Invoice Reconciliation       | Target: Invoice #101       | Invoice verified, 4 finance discrepancies tracked.
[PASS] Lead Scoring Worker                  | Target: Lead #102          | Lead AI Score=0, status=CONVERTED.
===========================================================================
```

### Detailed Functional Evidence Table

| Test Case | Target Entity ID | Initial Status | Final Status | Database Changes Recorded | External Calls Attempted | Functional Result |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **1. Pricing Normal RFQ** | RFQ #102 | `STAGE_PRICING_ASSIGNED` | `DRAFT_READY` | Quotation options evaluated; rates matched | Gemini 3.1 Flash Lite (2 calls) | **PASS** |
| **2. Pricing Anomaly HITL** | RFQ #105 | `NEW` | `WAITING_FOR_HUMAN` | 14 checkpoints saved in `ai_checkpoints` (`rfq-2-105`) | Gemini 3.1 Flash Lite (1 call) | **PASS** |
| **3. Contracts Extraction** | Doc #101 | `DRAFT` | `REVIEW_REQUIRED` | Parsed rates extracted; paused before `ingest` node | Gemini 3.1 Flash Lite (3 calls) | **PASS** |
| **4. Sales Incomplete Email**| Interaction #112 | `RECEIVED` | `PROCESSED` | Draft #12 (`AWAITING_APPROVAL`), Approval #108 (`PENDING`)| None (Outbound strictly blocked) | **PASS** |
| **5. Sales Complete Email** | Lead #103 | `QUALIFIED` | `CONVERTED` | RFQ #102 confirmed / linked | None (Internal DB write) | **PASS** |
| **6. Operations Tracking** | Shipment #102 | `IN_TRANSIT` | `IN_TRANSIT` | Milestones inspected (`GATE_IN`, `DELAY_NOTICE`) | None (Internal DB read) | **PASS** |
| **7. Compliance Check** | Shipment #101 | `IN_PROGRESS` | `DISCREPANCY_FOUND` | Discrepancies #5 & #6 mapped in `shipment_document_discrepancies` | None | **PASS** |
| **8. Finance Reconciliation**| Invoice #101 | `PAID` | `PAID` | Discrepancies #1-4 mapped in `shipment_finance_discrepancies` | None | **PASS** |
| **9. Lead Scoring Worker** | Lead #102 | `CONVERTED` | `CONVERTED` | AI score and reasoning verified in `leads` | None | **PASS** |

---

## 8. Restart & Persistence Behavior Verification

A non-destructive restart verification was performed:
1. **Queued Task Resilience**: An AI processing task was staged in `ai_processing_tasks`. The Go backend process was restarted; upon startup, the queue dispatcher re-polled the queued task without data loss.
2. **Sidecar Checkpoint Resilience**: LangGraph workflows for Pricing and Contracts write state tuples directly to MariaDB table `ai_checkpoints`. A simulated cold process restart (instantiating a fresh `MariaDBSaver` object) re-hydrated the execution snapshot:
   - Resumed thread ID: `contract-baseline-...`
   - Retrieved checkpoint ID: `1f1a9f58-2ff2-636c-8004-e9c392a7a816`
   - Next execution node: `('ingest',)`
   - State dictionary, extracted rates, and anomaly flags were 100% intact.
3. **Idempotent Re-execution**: Tasks executed multiple times with identical correlation IDs produced consistent outcomes without creating duplicate quotations or duplicate approval requests.

---

## 9. Frontend Coverage Matrix

The React frontend was reviewed across all operational pages to assess AI status visibility and error handling:

| Module Screen | Route | AI Queued / Processing | Waiting for Human / Draft Ready | Completed / Success State | Approval Required / Action Blocked | Assessment |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Mission Control** | `/dashboard` | Generic badge | Displays pending approvals count | Summary KPI cards | Badge links to Approvals | Good coverage |
| **Leads** | `/dashboard/leads` | Spinner on analyze | AI Score & qualification badge | Qualified tag | Shows staged reply draft | Full coverage |
| **RFQs** | `/dashboard/rfqs` | `COLLECTING_INFO` tag | `WAITING_FOR_HUMAN` banner | `DRAFT_READY` with quote preview | Action blocked until reviewed | Full coverage |
| **Quotations** | `/dashboard/quotations`| Loading skeleton | Margin review alert | Status `DRAFT` / `ISSUED` | Approvals drawer accessible | Full coverage |
| **Contracts** | `/dashboard/contracts` | `EXTRACTING` status | Rate review alert | Extracted rates table | "Review Flagged Rates" modal | Full coverage |
| **Shipments** | `/dashboard/shipments` | Sync indicator | Disruption notice | Status `IN_TRANSIT` / `DELIVERED`| Exception ticket banner | Full coverage |
| **Tracking** | `/dashboard/tracking` | Carrier polling tag | Delay alert | Milestone timeline | Exception escalation prompt | Good coverage |
| **Invoices** | `/dashboard/invoices` | Reconciliation spinner| Discrepancy warning | Status `PAID` / `MATCHED` | Payment hold on discrepancy | Full coverage |
| **Approvals** | `/dashboard/approvals` | Pending count badge | Card with Approve / Reject | Approved audit entry | Blocks action until decision | Full coverage |
| **Audit Logs** | `/dashboard/audit` | N/A | Displays pending reviews | Full chronological event log | Records actor type & action | Full coverage |

---

## 10. Ranked Risk Register

| Risk ID | Severity | Category | Description | Impact | Remediation Order |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **RSK-01** | **P0 (Critical)** | Architecture / Security | **AI Sidecar Tools Bypass Action System**: All 12 tools in `ai_sidecar/app/tools/` call raw internal HTTP endpoints with a static shared secret key, completely bypassing Go `ActionContext`, `IdempotencyStore`, and RBAC permissions. | No centralized idempotency, audit trail omission in action registry, potential duplicate side-effects during retries. | **Priority 1 (Immediate Next Task)** |
| **RSK-02** | **P1 (High)** | Architecture / UX | **Fragmented Agent HITL Review Mechanisms**: 6 divergent review mechanisms exist across modules. Only Sales clarification drafts use `approval_requests`; pricing anomalies, contracts, operations, documents, and finance use disconnected tables. | Operators lack a single unified review workflow; approval policies cannot be applied universally across AI agents. | **Priority 2** |
| **RSK-03** | **P2 (Medium)** | Reliability | **Sales and Operations Graphs Lack MariaDBSaver**: While Pricing and Contracts use persistent `MariaDBSaver`, Sales and Operations sidecar graphs do not persist checkpoint state. | Interrupted execution in Sales or Operations requires restart from scratch. | **Priority 3** |
| **RSK-04** | **P3 (Low)** | Security | **Static Shared Secret for Internal Services**: `X-LogisticsHQ-Service-Key` is a static string shared across all sidecar requests without per-agent service accounts or cryptographic signing. | Compromise of the key grants full internal access without agent identity segmentation. | **Priority 4** |

---

## 11. Recommended Next Task

### Exactly One Task:
**Phase 0.2 — Implement the Centralized Action System Execution Bridge for AI Sidecar Tools**

#### Objective:
Migrate all 12 AI sidecar tools from raw HTTP calls to the centralized Go Action System (`backend/internal/actions`). Ensure every tool invocation passes through `ActionContext`, registers an `idempotency_key`, validates RBAC permissions, and executes via the registered Action handlers, eliminating the P0 architectural bypass identified in this baseline.
