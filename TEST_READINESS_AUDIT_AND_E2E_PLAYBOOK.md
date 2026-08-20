# LogisticsHQ: Complete Repository-Wide Test Readiness Audit & E2E Testing Playbook

---

## EXECUTIVE SUMMARY

This document provides the authoritative, code-level **Test Readiness Audit and End-to-End Testing Playbook** for the LogisticsHQ codebase (`freel-project`). It is derived exclusively from the actual source code across the Go backend, PostgreSQL schema migrations, Python AI Sidecar (LangGraph), and React frontend.

---

## PART 1 — COMPLETE ARCHITECTURE TRACE

Trace of the active source code across the end-to-end lifecycle:

```
[Inbound Email] ──► [SalesAgent] ──► [RFQ Created] ──► [PricingAgent / Rate Engine]
       │                                                         │
[Shipment Created] ◄── [RFQ Won (EventBus)] ◄── [Quote Approved] ◄
       │
[Carrier Adapter / Webhook] ──► [OperationsAgent] ──► [Milestones & Exceptions]
       │
[Document Upload] ──► [ComplianceAgent] ──► [OCR / Extraction / Discrepancy HITL]
       │
[Carrier Invoice] ──► [FinanceAgent] ──► [3-Way Reconciliation / Discrepancy]
       │
[Customer Invoice Generation] ──► [Profitability Recalculation] ──► [4-Rule Closure Audit] ──► [Shipment Closed]
```

### Detailed Lifecycle Stage Breakdown

| Stage | Entry API / Event | Go Handler | Go Service | Repository / DB Tables | `ai_processing_tasks` Type | Python Handler | LangGraph Graph / Nodes | Python → Go Callback | React Page / Component | Key State Transitions | HITL Points | External Dependencies & Current Mocks | Current Known Limitations |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **1. Customer Email** | `POST /api/v1/emails/inbound` | `leads.EmailHandler.InboundEmailWebhook` | `leads.BusinessLogic` | `leads`, `lead_interactions`, `ai_processing_tasks` | `EMAIL_PARSE` | `run_email_parse_pipeline` | `sales_graph` (`classify` → `merge_context` → `check_completeness` → `research` → `draft_rfq` → `callback`) | `POST /internal/rfqs/from-email`, `POST /internal/sales/callback` | `src/pages/dashboard/Leads/LeadsPage.jsx` | Interaction: `INBOUND` → `PROCESSED`; Lead: `NEW` → `ACTIVE` | If incomplete info: drafts email response for human review | Direct HTTP (SES not required). Web search falls back from Tavily to DuckDuckGo. | Webhook assumes Org ID 1 by default if not passed in query/body |
| **2. Sales & RFQ Ingestion** | `POST /internal/rfqs/from-email` or `POST /api/v1/rfqs` | `rfq.Endpoints.CreateRFQ` | `rfq.BusinessLogic` | `rfqs`, `rfq_items`, `customers`, `companies` | None | None | None | None | `src/pages/dashboard/RFQ/RFQList.jsx`, `CreateRFQModal.jsx` | RFQ: `DRAFT` → `RFQ_CREATED` | Manual RFQ form submission or AI draft review | In-memory event bus fires `EventRFQCreated`. | Multi-commodity items stored as flat list |
| **3. Pricing & Rate Intelligence** | Event `EventRFQCreated` or `POST /internal/pricing/rules` | `agent.PricingAgent.HandleRFQ` | `pricing.Service`, `rates.Service` | `rate_entries`, `pricing_rules`, `rfq_quotes`, `ai_processing_tasks` | `PRICING_ANALYZE` | `run_pricing_pipeline` | `pricing_graph` (`agent` ⇄ `tools` → `validate` → `save`) | `POST /internal/pricing/quotes/draft`, `POST /internal/pricing/callback` | `src/pages/dashboard/RFQ/RFQDetail.jsx` (Quotes Tab) | RFQ: `PRICING_ASSIGNED` → `QUOTE_GENERATED` | Anomaly/Low Margin triggers `is_anomaly=true`, pausing before `save` | Rates: `rates.NewSpotNormalizer` + Mock Carrier Provider fallback; LLM fallback to deterministic calculation if keys absent. | Contract rates take precedence over spot rates |
| **4. Quotation Approval & Won** | `PUT /api/v1/rfqs/{id}/quotes/{qid}/approve` | `rfq.Endpoints.ApproveQuote` | `rfq.BusinessLogic` | `rfqs`, `rfq_quotes` | None (or `PRICING_RESUME` if resuming review) | `run_pricing_resume_pipeline` | Resumes `pricing_graph` | `POST /internal/pricing/callback` | `src/pages/dashboard/RFQ/RFQDetail.jsx` | RFQ: `QUOTE_GENERATED` → `QUOTE_SENT` → `WON` (Quote: `APPROVED`) | Sales / Pricing manager manual approval | Event bus fires `EventRFQWon` synchronously. | Single approved quote per RFQ |
| **5. Shipment Creation** | Handled synchronously by `EventRFQWon` subscriber | `shipments.Service.CreateFromRFQ` | `shipments.Service` | `shipments`, `shipment_milestones` | None | None | None | None | `src/pages/dashboard/Shipments/ShipmentsPage.jsx` | Shipment: `BOOKING_PENDING`, Milestones: `BOOKED`, `DEPARTED`, `IN_TRANSIT`, `ARRIVED`, `DELIVERED` (`PLANNED`) | None | Auto-populates origin, destination, carrier SCAC from approved quote. | One shipment per WON RFQ |
| **6. Carrier Integration & Tracking** | `POST /webhooks/carriers/{carrier}` or `{carrier}/{integration_id}` | `shipments.Handler.InboundWebhook` | `shipments.Service` | `carrier_tracking_events`, `shipment_processed_events`, `ai_processing_tasks` | `CARRIER_UPDATE_PARSE` | `run_operations_pipeline` | `operations_graph` (`parse_update` → `update_milestones` → `detect_exceptions` → `ops_action`) | `POST /internal/shipments/{id}/milestones`, `POST /internal/shipments/{id}/exceptions`, `POST /internal/operations/callback` | `src/pages/dashboard/Shipments/ShipmentDetail.jsx` (Operations Timeline) | `carrier_tracking_events`: `RECEIVED` → `PROCESSED`, Shipment: `IN_TRANSIT` / `EXCEPTION` | Critical exception flagged for Operations Manager | Poller runs background cron; Mock carrier adapter generates webhook payloads. | Unmatched tracking events stored in raw audit log table |
| **7. Documents & Compliance** | `POST /api/v1/shipments/{id}/documents/upload` | `documents.Handler.UploadDocument` | `documents.Service` | `shipment_documents`, `shipment_document_discrepancies`, `ai_processing_tasks` | `DOC_VERIFY` | `run_compliance_pipeline` | `compliance_graph` (`ocr` → `extract` → `reconcile` → `report`) | `POST /internal/compliance/callback` | `src/pages/dashboard/Shipments/DocumentWorkspace.jsx` | `shipment_documents`: `PENDING_VERIFICATION` → `VERIFIED` or `DISCREPANCY` | Human discrepancy resolution (`POST /api/v1/shipments/discrepancies/{id}/resolve`) | Files stored in `./uploads` locally or AWS S3. Local PyPDF fallback for OCR. | Enforces 1 document per type (`HBL`, `MBL`, `PACKING_LIST`, `COMMERCIAL_INVOICE`) per shipment |
| **8. Carrier Invoice & Finance** | `POST /api/v1/shipments/{id}/finance/invoices/upload` | `finance.Handler.IngestInvoice` | `finance.Service` | `shipment_invoices`, `shipment_invoice_items`, `shipment_finance_discrepancies`, `ai_processing_tasks` | `BILL_RECONCILE` | `run_finance_pipeline` | `finance_graph` (`ocr` → `extract` → `reconcile` → `report`) | `POST /internal/finance/callback` | `src/pages/dashboard/Shipments/FinanceWorkspace.jsx` | `shipment_invoices`: `PENDING_RECONCILIATION` → `APPROVED` / `DISCREPANCY` | Human discrepancy resolution & invoice approval | 3-way check against `rate_entries` / accepted `rfq_quotes`. | Enforces unique `invoice_number` per organization |
| **9. Customer Billing & Profitability** | `POST /api/v1/shipments/{id}/billing/invoices/generate` | `billing.Handler.GenerateInvoice` | `billing.Service` | `shipment_customer_invoices`, `shipment_customer_invoice_items`, `shipment_finance_profitability` | None | None | None | None | `src/pages/dashboard/Shipments/BillingWorkspace.jsx` | Invoice: `DRAFT` → `APPROVED` → `PAID`; Profitability: `PENDING` → `ON_TARGET` / `UNDER_TARGET` / `NEGATIVE_MARGIN` | Manual invoice item modification & approval | Deterministic Go math engine computing variance against quoted margins. | Default Net 30 payment terms |
| **10. Shipment Closure** | `POST /api/v1/shipments/{id}/close` | `billing.Handler.CloseShipment` | `billing.Service` | `shipments`, `shipment_finance_profitability` | None | None | None | None | `src/pages/dashboard/Shipments/BillingWorkspace.jsx` | Shipment: `DELIVERED` → `CLOSED` | Enforces 4 strict prerequisite rules | Fails with 400 Bad Request if any prerequisite is unfulfilled. | Terminal state: closed shipments cannot be reopened |

---

## PART 2 — DEVELOPMENT ENVIRONMENT SPECIFICATION

### Required Tooling Versions

*   **Go Version:** `1.25.x` (Defined in `backend/go.mod`: `go 1.25.2`)
*   **Python Version:** `3.11.x` or `3.12.x` (Virtual environment in `ai_sidecar/venv`)
*   **Node.js Version:** `v20.x` or `v22.x` (LTS)
*   **Package Manager:** `npm` (Lockfile `frontend/package-lock.json` present)
*   **PostgreSQL Version:** `15.x` or `16.x`
*   **Docker:** Required if running PostgreSQL in a container

### Process and Port Allocation

| Component | Working Directory | Runtime Command | Default Port | Internal Dependencies |
| :--- | :--- | :--- | :--- | :--- |
| **PostgreSQL** | Docker / Daemon | `docker run -p 5432:5432 ...` | `5432` | None |
| **Go Backend Server** | `backend/` | `go run ./cmd/server` | `8080` | PostgreSQL (5432), AI Sidecar (8090) |
| **Python AI Sidecar** | `ai_sidecar/` | `uvicorn main:app --port 8090` | `8090` | PostgreSQL (5432), Go Backend (8080) |
| **React Frontend** | `frontend/` | `npm run dev` | `5173` | Go Backend (8080) |

---

## PART 3 — ENVIRONMENT VARIABLES AUDIT

| Variable | Component | Required? | Development Value | Production Value / Source | Purpose |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `APP_ENV` | Go, Python | **Required** | `development` | `production` | Controls dev auth bypass and strict secret enforcement |
| `PORT` | Go | Optional | `8080` | `8080` | HTTP listening port for Go backend |
| `DB_URL` / `DATABASE_URL` | Go, Python | **Required** | `postgres://user:password@localhost:5432/freel?sslmode=disable` | Secret manager / RDS URL | PostgreSQL connection string |
| `GO_BACKEND_URL` | Go, Python | **Required** | `http://localhost:8080` (Python: `http://127.0.0.1:8080`) | `http://backend:8080` | Base URL used for internal callbacks and tool execution |
| `AI_SIDECAR_URL` | Go | Optional | `http://localhost:8090` | `http://ai-sidecar:8090` | Python FastAPI sidecar endpoint for contract processing |
| `INTERNAL_SERVICE_TOKEN` | Go, Python | **Required** | `internal-service-key-logisticshq` | 64+ char random hex string | Shared secret for `X-LogisticsHQ-Service-Key` header |
| `FRONTEND_URL` | Go | Optional | `http://localhost:5173` | `https://app.logisticshq.in` | CORS allowed origin in development |
| `FRONTEND_PROD_URL` | Go | Optional | `https://logisticshq.in` | `https://logisticshq.in` | Production CORS allowed origin |
| `AWS_REGION` | Go | Optional | `ap-south-1` | `ap-south-1` | AWS region for Cognito & S3 |
| `COGNITO_USER_POOL_ID` | Go | Optional (Dev) | `ap-south-1_fMCgUxYeV` | AWS Cognito User Pool ID | JWKS token validation endpoint |
| `COGNITO_CLIENT_ID` | Go | Optional (Dev) | `3isa6qmgr2sc452ta5bf2okkjg` | AWS Cognito Client ID | Auth client identification |
| `COGNITO_CLIENT_SECRET` | Go | Optional (Dev) | `1pfilr2onu2...` | AWS Cognito Client Secret | Auth client secret calculation |
| `S3_BUCKET` | Go | Optional (Dev) | *empty* (uses local `./uploads`) | `logisticshq-prod-documents` | Activates AWS S3 storage instead of local disk |
| `GEMINI_API_KEY` / `GOOGLE_API_KEY` | Go, Python | Optional | *Active Gemini Key* | Google AI Studio Key | Primary LLM engine (Gemini 2.0 Flash) |
| `OPENAI_API_KEY` | Go, Python | Optional | *Active OpenAI Key* | OpenAI API Key | Failover fallback LLM (GPT-4o-mini) |
| `TAVILY_API_KEY` | Python | Optional | *empty* | Tavily API Key | Web search agent (falls back to DuckDuckGo if empty) |
| `LANGCHAIN_TRACING_V2` | Python | Optional | `false` | `true` | Enables LangSmith distributed tracing |
| `LANGCHAIN_API_KEY` | Python | Optional | *empty* | LangSmith API Key | LangSmith telemetry authorization |
| `LANGCHAIN_PROJECT` | Python | Optional | `logisticshq-agents` | `logisticshq-production` | LangSmith project container name |
| `LANGCHAIN_ENDPOINT` | Python | Optional | `https://api.smith.langchain.com` | `https://api.smith.langchain.com` | LangSmith ingest endpoint |
| `VITE_API_BASE_URL` | Frontend | **Required** | `http://localhost:8080` | `https://api.logisticshq.in` | Frontend target URL for backend API requests |

---

## PART 4 — LOCAL DEVELOPMENT & TEST STRATEGY

1.  **Fully Local Implementations (Zero External Cloud Dependencies):**
    *   **Task Queue:** The system uses `ai_processing_tasks` in PostgreSQL with `SELECT FOR UPDATE SKIP LOCKED`. SQS/Kafka are not used.
    *   **File Storage:** When `S3_BUCKET` is empty, `files.NewLocalService` stores files in `./uploads` and serves them statically at `http://localhost:8080/uploads/`.
    *   **Email Processing:** Inbound emails are sent directly to `POST /api/v1/emails/inbound`. SES is bypassed.
    *   **Carrier Integrations:** Mock carrier provider in `internal/carrier/mock.go` supplies tracking and spot rate responses.
2.  **Authentication Strategy:**
    *   In development (`APP_ENV=development`), requests with `Authorization: Bearer test-token` are authenticated as `UserID: 1, OrgID: 1, Role: ADMIN` without contacting AWS Cognito.
3.  **LLM Strategy:**
    *   `llm_factory.py` implements a 3-tier cascade:
        1. Google Gemini 2.0 Flash (`GEMINI_API_KEY` / `GOOGLE_API_KEY`)
        2. OpenAI GPT-4o-mini (`OPENAI_API_KEY`)
        3. Local deterministic fallback (extracts lane from prompt and produces valid quote JSON)

---

## PART 5 — DATABASE SETUP & MIGRATIONS

### Migration Sequence

The database schema is managed via 24 sequential SQL migration files in `backend/internal/database/migrations/`:

```
001_rbac.sql                      ──► Base organizations, users, roles, org_members, permissions
002_companies_leads.sql           ──► addresses, companies, contacts, customers
003_outreach.sql                  ──► outreach_campaigns, outreach_sequences
004_rfq.sql                       ──► rfqs, rfq_items
005_opportunities.sql             ──► opportunities
005_rfq_enhancements.sql          ──► health_score on rfqs, reliability on rfq_quotes
006_audit_activity.sql            ──► audit_logs, activities
006_pricing_agent.sql             ──► agent_status, ai_reasoning on rfqs/quotes
007_invitations.sql               ──► invitations
008_leads.sql                     ──► leads
009_rfq_free_days.sql             ──► free_days on rfq_quotes
010_rate_intelligence.sql         ──► carriers, rate_entries
011_contracts.sql                 ──► contract_documents, contract_review_items
012_ai_processing_queue.sql       ──► ai_processing_tasks
013_generalize_ai_queue.sql       ──► entity_type, entity_id on ai_processing_tasks
013b_pricing_rules.sql            ──► pricing_rules
013c_sales_crm.sql                ──► lead_interactions
013d_interaction_ai_response.sql  ──► ai_summary, drafted_reply on lead_interactions
013e_thread_context.sql           ──► parent_interaction_id, partial_rfq_context
014_operations_shipments.sql      ──► shipments, shipment_milestones, shipment_exceptions
015_operations_hardening.sql      ──► carrier_tracking_events, queue lock columns
016_shipment_documents.sql        ──► shipment_documents, shipment_document_discrepancies
017_shipment_invoices.sql         ──► shipment_invoices, shipment_invoice_items, discrepancies
018_finance_billing_closure.sql   ──► shipment_customer_invoices, items, profitability
```

### Database Initialization Commands

```bash
# 1. Start PostgreSQL (Docker)
docker run -d --name freel-postgres \
  -e POSTGRES_USER=user \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=freel \
  -p 5432:5432 postgres:16-alpine

# 2. Run all migrations in sequence
cd backend
for file in $(ls internal/database/migrations/*.sql | sort); do
  echo "Applying migration: $file"
  psql "postgres://user:password@localhost:5432/freel?sslmode=disable" -f "$file"
done

# 3. Seed Development Data
go run ./cmd/seed
```

### Seed Data Verification

The `cmd/seed` binary populates:
*   **Organization:** `Freel Global Logistics Pvt Ltd` (ID: `1`)
*   **Users:**
    *   `ceo@freel-demo.local` (Role: `CEO`, ID: `1`)
    *   `sales@freel-demo.local` (Role: `SALES`, ID: `2`)
    *   `pricing@freel-demo.local` (Role: `PRICING`, ID: `3`)
    *   `customer@tata-exports.local` (Role: `CUSTOMER_CONTACT`, ID: `4`)
*   **Customers:** `Tata Exports Ltd`, `Sun Pharma`, `Mahindra Auto`, `Reliance Petro`, `Adani Ports`
*   **RFQs:** 6 RFQs covering stages `RFQ_CREATED`, `PRICING_ASSIGNED`, `QUOTE_GENERATED`, `QUOTE_SENT`, `WON`, `LOST`
*   **Quotes:** Seeded carrier quotes from `Maersk`, `MSC`, `CMA CGM`, `Hapag-Lloyd`, `Evergreen`
*   **Pricing Rules:** Default 20% markup, INNSA-DEHAM 12% promo, Enterprise Tier 10% discount

---

## PART 6 — AUTHENTICATION & MULTI-TENANCY DEEP DIVE

### Authentication Middleware Architecture

```
Incoming Request
      │
      ├── Header: 'Authorization: Bearer <token>' missing? ──► 401 Unauthorized
      │
      ├── Token == 'test-token'?
      │        ├── YES ──► Injects UserContext(UserID: 1, OrgID: 1, Role: 'ADMIN') ──► Next Handler
      │        └── NO  ──► Validates JWT via AWS Cognito JWKS Cache
      │                         │
      │                         ├── Invalid / Expired? ──► 401 Unauthorized
      │                         └── Valid JWT ──► Extract cognito_sub
      │                                                │
      │                                                ▼
      │                                     Query PostgreSQL:
      │                                     SELECT u.id, om.org_id, r.name
      │                                     FROM users u
      │                                     JOIN org_members om ON u.id = om.user_id
      │                                     JOIN roles r ON om.role_id = r.id
      │                                     WHERE u.cognito_sub = $1 AND om.status = 'ACTIVE'
      │                                                │
      │                                                ├── Not Found / Inactive? ──► 401 Unauthorized
      │                                                └── Found ──► Injects UserContext ──► Next Handler
```

### Dev Auth Test Scenarios Matrix

| Scenario | Token Header | Cognito Status | Database User Record | Expected Status | Result |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **1. Dev Bypass** | `Bearer test-token` | Ignored | Ignored | `200 OK` | Authenticated as User 1, Org 1, ADMIN |
| **2. Valid Cognito User** | `Bearer <valid_jwt>` | Valid Signature & Claims | Active user & active org_member | `200 OK` | Authenticated as DB user & org |
| **3. Unknown cognito_sub** | `Bearer <valid_jwt>` | Valid Signature | No matching `cognito_sub` in `users` | `401 Unauthorized` | "User context not found or inactive" |
| **4. Inactive Org Member** | `Bearer <valid_jwt>` | Valid Signature | `org_members.status = 'SUSPENDED'` | `401 Unauthorized` | "User context not found or inactive" |
| **5. Missing Header** | *None* | N/A | N/A | `401 Unauthorized` | "Missing or invalid authorization header" |
| **6. Invalid Signature** | `Bearer fake.jwt.blob` | JWKS verify fails | N/A | `401 Unauthorized` | "Invalid token" |
| **7. Cross-Tenant Access** | `Bearer token_org_a` | Valid Org A | Org B resource ID requested | `404 Not Found` | Query filtered by `org_id = 1` returns 0 rows |

---

## PART 7 — PYTHON / LANGGRAPH AGENT REGISTRY

| Agent | Task Type | Graph Identifier | Node Pipeline | Primary LLM | Tools Used | Go Backend APIs Invoked | Python → Go Callback Endpoint | Checkpointer & HITL Interrupts |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **SalesAgent** | `EMAIL_PARSE` | `sales_graph` | `classify` → `merge_context` → `check_completeness` → `research` → `draft_rfq` → `callback` | Gemini 2.0 Flash (OpenAI failover) | `web_search_tool`, `get_lead_details_tool`, `save_lead_research_tool` | `POST /internal/rfqs/from-email`, `GET /internal/leads/{id}` | `POST /internal/sales/callback` | `PostgresSaver`. If incomplete RFQ, routes to callback with draft reply (no hard interrupt). |
| **PricingAgent** | `PRICING_ANALYZE` / `PRICING_RESUME` | `pricing_graph` | `agent` ⇄ `tools` → `validate` → `save` | Gemini 2.0 Flash (OpenAI failover) | `get_rfq_details_tool`, `search_rates_tool`, `get_pricing_rules_tool`, `create_draft_quotes_tool` | `GET /internal/rfqs/{id}`, `GET /internal/rates/search`, `GET /internal/pricing/rules`, `POST /internal/pricing/quotes/draft` | `POST /internal/pricing/callback` | `PostgresSaver`. Interrupts before `save` node if `is_anomaly=true` (Low margin / high buy price). |
| **OperationsAgent** | `CARRIER_UPDATE_PARSE` | `operations_graph` | `parse_update` → `update_milestones` → `detect_exceptions` → `ops_action` | Gemini 2.0 Flash + Regex keyword fallback | `update_milestone_tool`, `create_exception_tool` | `GET /internal/shipments/{id}`, `POST /internal/shipments/{id}/milestones`, `POST /internal/shipments/{id}/exceptions` | `POST /internal/operations/callback` | `PostgresSaver`. Fully automated. Flags `has_critical_exception=true` in callback. |
| **ComplianceAgent** | `DOC_VERIFY` | `compliance_graph` | `ocr` → `extract` → `reconcile` → `report` | Gemini 2.0 Flash (Structured JSON output) | Local PyPDF / Regex normalizer | `GET /internal/shipments/{id}/documents` | `POST /internal/compliance/callback` | Ephemeral `StateGraph`. Generates discrepancies on field mismatch for Go HITL resolution. |
| **FinanceAgent** | `BILL_RECONCILE` | `finance_graph` | `ocr` → `extract` → `reconcile` → `report` | Gemini 2.0 Flash (Structured JSON output) | Local PyPDF / Regex normalizer | `GET /internal/shipments/{id}/finance` | `POST /internal/finance/callback` | Ephemeral `StateGraph`. Compares invoice items against quote/rates; flags overcharges for Go approval. |
| **ContractReader** | `PROCESS` / `RESUME` | `contracts_graph` | `ocr` → `classify` → `parser` → `validator` → `ingest` | Gemini 2.0 Flash | Port Normalizer tools | `GET /internal/ports/normalize`, `GET /internal/ports/search` | `POST /internal/contracts/callback` | `PostgresSaver`. Interrupts before `ingest` if rate anomaly detected. |

---

## PART 8 — LLM & EXTERNAL SERVICE REQUIREMENTS

### Deterministic vs. LLM Workflows

*   **Deterministic (No LLM Required):** Milestone normalizer, Exception severity engine, 3-Way reconciliation math checks, Customer invoice generation, Margin & variance calculation, 4-Rule shipment closure audit.
*   **LLM-Powered:** SalesAgent email parsing, Lead enrichment strategy, Complex PDF OCR extraction, Shipping document cross-reconciliation, Anomaly reasoning generation.
*   **Minimum API Keys for Testing:** `GEMINI_API_KEY` (or zero keys if utilizing the built-in mock fallback).

---

## PART 9 — EMAIL & SALES INGESTION TEST SUITE

| ID | Test Case | Inbound Email Body / Condition | Expected DB State & Task | Expected SalesAgent Action | Expected Callback / Result | Expected UI State |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **EM-01** | Valid RFQ Email | Full route (INNSA→DEHAM), cargo (24t auto parts), EXW, target date | Lead created/resolved; interaction logged (`INTENT=RFQ_REQUEST`); `ai_processing_tasks` row inserted | Classifies `RFQ_REQUEST`, checks completeness (PASSED), calls `POST /internal/rfqs/from-email` | Callback returns `linked_rfq_id`; Lead status `ACTIVE` | RFQ appears on `/dashboard/rfqs` in `RFQ_CREATED` stage |
| **EM-02** | Missing Origin | "Quote 2x40GP to Hamburg (DEHAM) EXW" | Lead interaction logged (`INTENT=RFQ_REQUEST`); task queued | Detects missing origin; marks `RFQ_REQUEST_INCOMPLETE`; drafts clarification email | Callback returns `drafted_reply`, `partial_rfq_context` saved with destination & equipment | Shows interaction in Lead drawer with "Draft Reply" pending |
| **EM-03** | Missing Destination | "Quote 2x40GP from Nhava Sheva (INNSA)" | Interaction logged; task queued | Detects missing destination; marks `RFQ_REQUEST_INCOMPLETE`; drafts reply asking for POD | Callback persists partial context; no RFQ created | Lead timeline shows incomplete request |
| **EM-04** | Missing Cargo/Weight | "Quote shipping from INNSA to DEHAM" | Interaction logged; task queued | Completeness check fails; drafts reply asking for equipment & cargo details | Callback stores context; `linked_rfq_id=null` | Lead timeline shows draft response |
| **EM-05** | Non-RFQ Inquiry | "Can you send me your company brochure?" | Interaction logged; task queued | Classifies intent as `QUESTION`; generates general response | Callback updates interaction `sentiment=NEUTRAL`, `intent=QUESTION` | No RFQ created; interaction visible in CRM |
| **EM-06** | Unsubscribe Request | "Unsubscribe me from your outreach emails" | Interaction logged; task queued | Classifies intent as `UNSUBSCRIBE` | Updates lead status to `REJECTED` / unsubscribed | Outreach engine excludes this lead |
| **EM-07** | Malformed Payload | Missing `from` or `body` | No DB insert | N/A | HTTP `400 Bad Request` (`MISSING_PARAMS`) | Validation error displayed |
| **EM-08** | Duplicate Email | Same `message_id` sent twice | Unique constraint on `raw_email_id` / thread handling | Skips task re-queue | HTTP `200 OK` (Idempotent response) | No duplicate RFQs |
| **EM-09** | Thread Follow-Up | Reply to EM-02: "Origin port is INNSA" | Parent interaction linked via `thread_id`; `is_reply=true` | Merges prior context (`DEHAM`) with new origin (`INNSA`), completeness check passes, creates RFQ | Callback returns `linked_rfq_id` | RFQ created from conversational thread |
| **EM-10** | Lead Auto-Creation | Email from brand new domain `contact@newbuyer.com` | Automatically inserts company and lead in `companies` and `leads` | Enriches company domain via `web_search_tool` | Lead assigned `ai_score` and research report | New lead appears in Leads list |

---

## PART 10 — PRICING & QUOTATION TEST SUITE

| ID | Test Scenario | Rates Available in DB | Applied Pricing Rules | Expected AI / Engine Action | Expected RFQ Stage & DB Result | HITL Action Required |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **PR-01** | Contract Rate Match | Matching contract rate in `rate_entries` (Buy: $2,000) | Default 20% markup rule | Calculates Sell: $2,400 (Margin: 16.7%). Auto-saves draft quotes. | RFQ stage: `QUOTE_GENERATED`, Quote: `DRAFT` ($2,400) | None (Auto-saved) |
| **PR-02** | Spot Rate Fallback | No contract rate; Spot rate returned via Carrier Service (Buy: $2,500) | Default 20% markup | Calculates Sell: $3,000. Creates quote marked source `SPOT_API`. | RFQ stage: `QUOTE_GENERATED`, Quote: `DRAFT` ($3,000) | None |
| **PR-03** | Both Rates Available | Contract ($2,000) and Spot ($2,200) available | Default 20% markup | Recommends Contract rate (higher reliability score 95 vs 88) | RFQ stage: `QUOTE_GENERATED`, 2 quotes created, Contract marked `is_recommended=true` | None |
| **PR-04** | No Rate Available | No contract, live spot fetch fails | N/A | Fails rate search; logs error; requests manual pricing | RFQ stage: `PRICING_ASSIGNED`, `agent_status=FAILED` | Manual pricing entry required |
| **PR-05** | Lane Promo Markup | Matching contract ($2,000) on `INNSA` → `DEHAM` | Lane promo rule (12% markup, Priority 10) overrides Default (20%) | Applies 12% markup: Sell: $2,240 (Margin: 10.7%) | Quote created with Sell: $2,240 | None |
| **PR-06** | Normal Margin | Buy: $1,000, Sell: $1,250 (Margin: 20%) | Min margin rule (5.0%) | Validates margin >= 5.0%; `is_anomaly=false` | Auto-persists quote; RFQ: `QUOTE_GENERATED` | None |
| **PR-07** | Low Margin Anomaly | Buy: $2,000, Sell: $2,050 (Margin: 2.4%) | Min margin rule (5.0%) | Anomaly detected (`Margin 2.4% < Min 5.0%`); pauses before save | RFQ: `PRICING_ASSIGNED`, callback status `NEEDS_REVIEW` | Human must approve via `PRICING_RESUME` |
| **PR-08** | High Buy Price Anomaly | Buy: $12,000 (Market average $2,500) | Anomaly threshold check | Flags high cost anomaly; `is_anomaly=true` | RFQ: `PRICING_ASSIGNED`, review item flagged | Human must confirm rate |
| **PR-09** | Quote Draft Review | Draft quote generated for $2,400 | N/A | Human reviews quote in UI | No DB change | Review in UI |
| **PR-10** | Quote Approval | Quote in `DRAFT` status | N/A | `PUT /api/v1/rfqs/{id}/quotes/{qid}/approve` | Quote status `APPROVED`; RFQ stage `QUOTE_SENT` | Manager click Approve |
| **PR-11** | Quote Rejection | Quote in `DRAFT` status | N/A | `PUT /api/v1/rfqs/{id}/quotes/{qid}/reject` | Quote status `REJECTED`; RFQ stage remains `PRICING_ASSIGNED` | Manager click Reject |
| **PR-12** | Pricing Graph Resume | Paused `pricing_graph` thread | Human approved override | `run_pricing_resume_pipeline` sets `is_anomaly=false`, executes save | Quote saved; RFQ stage `QUOTE_GENERATED` | Triggered by review approval |

---

## PART 11 — OPERATIONS & CARRIER TRACKING TEST SUITE

| ID | Test Scenario | Webhook Endpoint / Headers | Payload Identifiers | Expected DB Actions & State | Expected OperationsAgent Result |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **OP-01** | Valid Carrier Webhook | `POST /webhooks/carriers/MAEU` | Booking: `BKG-1001`, Milestone: `DEPARTED`, Vessel: `MAERSK MC-KINNEY` | Insert `carrier_tracking_events` (`MATCHED`, `QUEUED`); enqueue task | Updates milestone `DEPARTED` to `COMPLETED`; Shipment status `DEPARTED` |
| **OP-02** | Valid Integration ID | `POST /webhooks/carriers/MAEU/101` | Matching active integration in DB | Resolves `org_id` from integration; processes event | Updates shipment for that specific organization |
| **OP-03** | Wrong Integration ID | `POST /webhooks/carriers/MAEU/99999` | Non-existent integration ID | Returns `404 Not Found` | No task queued; rejection logged |
| **OP-04** | Inactive Integration | `POST /webhooks/carriers/MAEU/102` | Integration with `is_active=false` | Returns `403 Forbidden` | Processing rejected |
| **OP-05** | Wrong Carrier SCAC | `POST /webhooks/carriers/INVALID` | Unknown carrier | Returns `400 Bad Request` | No event stored |
| **OP-06** | Invalid Webhook Signature | `POST /webhooks/carriers/MAEU` + Bad HMAC | Any payload | Returns `401 Unauthorized` | Rejected before DB insert |
| **OP-07** | Malformed JSON Payload | `POST /webhooks/carriers/MAEU` | `{ invalid_json ` | Returns `400 Bad Request` | No crash; rejection response |
| **OP-08** | Duplicate Tracking Event | Same `event_id` sent twice | `event_id: "evt-001"` | Deduplication triggered: second insert ignored via unique constraint | Idempotent `200 OK`, no duplicate task queued |
| **OP-09** | Vessel Delay Exception | Valid webhook | Description: "Vessel delayed by 5 days due to weather" | `OperationsAgent` detects `DELAY` severity `WARNING` | Creates row in `shipment_exceptions`; Shipment status `EXCEPTION` |
| **OP-10** | Normal Milestone Progression | Sequential webhooks | `BOOKED` → `DEPARTED` → `IN_TRANSIT` → `ARRIVED` → `DELIVERED` | Updates corresponding `shipment_milestones` to `COMPLETED` | Shipment status moves to `DELIVERED` |
| **OP-11** | Stale / Out-of-Order Milestone | Webhook for `DEPARTED` received after `ARRIVED` | Milestone `DEPARTED` with older timestamp | Milestone marked completed without regressing shipment status | Shipment status remains `ARRIVED` |
| **OP-12** | Ambiguous Shipment Match | Payload contains container `MSCU1234567` matching 2 active shipments | Container shared or duplicate | `matching_status` set to `AMBIGUOUS` in `carrier_tracking_events` | Task flags review; no milestone updated automatically |
| **OP-13** | Unmatched Shipment | Booking number not found in any shipment | `booking_number: "UNKNOWN-999"` | `matching_status` set to `UNMATCHED` | Stored in event log for audit; no task queued |
| **OP-14** | Multi-Org Same Carrier | Org 1 and Org 2 both have shipments on Maersk | Distinct integration IDs | Events isolated strictly to respective org's shipment | Zero cross-tenant data leakage |
| **OP-15** | Carrier Poller Cycle | Background `CarrierPoller` runs every 5m | Active shipments with `status IN ('BOOKED', 'IN_TRANSIT')` | Queries carrier API for tracking updates; ingests new events | Automatic milestone updates |
| **OP-16** | Manual Status Override | `POST /api/v1/shipments/{id}/carrier-update` | Operator manually sets `milestone_code: "DELIVERED"` | Milestone `DELIVERED` updated; Shipment status `DELIVERED` | Audit log record created |

---

## PART 12 — DOCUMENT COMPLIANCE TEST SUITE

| ID | Test Scenario | Document Type Uploaded | Test Payload / OCR Text | Expected Extraction & Reconciliation | Expected Result & Discrepancy Record |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **DC-01** | HBL Upload | `HBL` (House Bill of Lading) | MBL: 24,000kg, HBL: 24,000kg, Seal: `SL-991` | Extracts weight, pieces, container, seal | Document status: `VERIFIED` |
| **DC-02** | MBL Upload | `MBL` (Master Bill of Lading) | Carrier: Maersk, Vessel: Mc-Kinney, 24,000kg | Extracts carrier, vessel, voyage, container | Document status: `VERIFIED` |
| **DC-03** | Packing List Upload | `PACKING_LIST` | 500 cartons, Gross Weight: 24,000kg | Extracts package count, weight | Document status: `VERIFIED` |
| **DC-04** | Commercial Invoice | `COMMERCIAL_INVOICE` | Total: $48,000 USD, 500 cartons | Extracts invoice total, currency, items | Document status: `VERIFIED` |
| **DC-05** | Perfect Document Set | All 4 documents uploaded | All weights (24,000kg) and containers (`MSCU1234567`) match | Reconciles across all 4 documents with 0 mismatches | All documents `VERIFIED`; ready for closure |
| **DC-06** | Gross Weight Mismatch | HBL vs MBL | MBL says `24,000 kg`, HBL says `26,500 kg` | ComplianceAgent flags weight variance | Creates row in `shipment_document_discrepancies` (`field_name='gross_weight'`, `expected='24,000 kg'`, `actual='26,500 kg'`, `status='OPEN'`) |
| **DC-07** | Container ID Mismatch | HBL vs Packing List | HBL: `MSCU1234567`, Packing List: `MSCU9999999` | Flags container number mismatch | Creates discrepancy record (`field_name='container_numbers'`) |
| **DC-08** | Seal Number Mismatch | MBL vs HBL | MBL: `SEAL-001`, HBL: `SEAL-002` | Flags seal number mismatch | Creates discrepancy record (`field_name='seal_numbers'`) |
| **DC-09** | Duplicate Document | Uploading HBL when HBL already exists | New HBL PDF uploaded | Replaces or versions document; triggers re-verification | Updates existing record; re-runs reconciliation |
| **DC-10** | Discrepancy Resolution | OPEN Discrepancy exists | `POST /api/v1/shipments/discrepancies/{id}/resolve` | Sets `status='RESOLVED'`, `resolved_by=UserID`, `resolved_at=NOW()` | Discrepancy marked resolved; document unblocked |
| **DC-11** | Callback Retry Failure | Go server restarts during callback | Python callback receives 503 | Queue worker catches error, backs off, retries | Task completed successfully on retry |
| **DC-12** | Multi-Tenant Doc Isolation | Org 2 uploads document to Org 1 shipment | `POST /api/v1/shipments/{org1_shipment_id}/documents/upload` | Go handler checks `shipment.org_id == userCtx.OrgID` | Returns `404 Not Found` / `403 Forbidden` |

---

## PART 13 — FINANCE & 3-WAY RECONCILIATION TEST SUITE

| ID | Test Scenario | Uploaded Invoice Items | Expected Contract / Quote Benchmark | Expected FinanceAgent Action | Expected DB Records & Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **FN-01** | Normal Matching Invoice | Ocean Freight: $2,000, BAF: $300 (Total: $2,300) | Quoted Buy Price: $2,300 | Matches 100% against accepted quote items | `shipment_invoices.status = 'APPROVED'`; items inserted |
| **FN-02** | Base Freight Overcharge | Ocean Freight: $2,400 (Quoted: $2,000) | Quoted Buy Price: $2,000 | Flags $400 base freight discrepancy | `shipment_invoices.status = 'DISCREPANCY'`; discrepancy row inserted (`field_name='ocean_freight'`, `expected='2000'`, `actual='2400'`) |
| **FN-03** | Surcharge Overcharge | BAF Surcharge: $500 (Contracted: $300) | Contract rate surcharge: $300 | Flags $200 surcharge overcharge | Discrepancy row inserted (`field_name='BAF'`, `expected='300'`, `actual='500'`) |
| **FN-04** | Unauthorized Charge | Demurrage charge: $450 | Free days valid; no demurrage authorized | Flags unauthorized charge code `DEMURRAGE` | Discrepancy row inserted (`charge_code='DEMURRAGE'`, `source='QUOTE'`) |
| **FN-05** | Duplicate Invoice | Same `invoice_number` uploaded twice | Unique index `uq_invoice_number_org` | Rejects duplicate invoice number | Returns HTTP `400 / 409 Conflict` |
| **FN-06** | Discrepancy Resolution | Open finance discrepancy | `POST /api/v1/finance/discrepancies/{id}/resolve` | Updates discrepancy `status = 'RESOLVED'` | Discrepancy cleared |
| **FN-07** | Manual Invoice Approval | Invoice in `DISCREPANCY` status | `POST /api/v1/finance/invoices/{id}/approve` | Overrides discrepancy; sets `status = 'APPROVED'` | Invoice approved; triggers profitability recalculation |
| **FN-08** | Multi-Tenant Invoice Isolation | Org 1 user attempts to view Org 2 finance workspace | `GET /api/v1/shipments/{org2_id}/finance` | Query filtered by `org_id = 1` | Returns `404 Not Found` |

---

## PART 14 — CUSTOMER BILLING & PROFITABILITY TEST SUITE

| ID | Test Scenario | Input / Action | Calculation & Expected Values | Expected Profitability Status |
| :--- | :--- | :--- | :--- | :--- |
| **BL-01** | Generate Customer Invoice | `POST /api/v1/shipments/{id}/billing/invoices/generate` | Quoted Buy: $2,000, Sell: $2,400. Creates invoice with items totaling $2,400. | Invoice: `DRAFT`, Profitability: `PENDING` |
| **BL-02** | Customer Invoice Approval | `POST /api/v1/billing/invoices/{id}/approve` | Sets invoice `status = 'APPROVED'`. Actual Revenue becomes $2,400. | Triggers `RecalculateProfitability` |
| **BL-03** | Customer Invoice Payment | `POST /api/v1/billing/invoices/{id}/pay` | Sets invoice `status = 'PAID'` | Invoice `PAID` |
| **BL-04** | On-Target Profitability | Quoted Sell: $2,400, Quoted Buy: $2,000, Actual Carrier Cost: $2,000, Actual Revenue: $2,400 | Expected Profit: $400 (16.7%), Actual Profit: $400 (16.7%), Variance: $0 | `ON_TARGET` |
| **BL-05** | Under-Target Margin | Quoted Sell: $2,400, Actual Carrier Cost: $2,200 (extra carrier fee), Actual Revenue: $2,400 | Expected Profit: $400, Actual Profit: $200 (8.3%), Variance: -$200 | `UNDER_TARGET` |
| **BL-06** | Negative Margin Alert | Quoted Sell: $2,400, Actual Carrier Cost: $2,700, Actual Revenue: $2,400 | Expected Profit: $400, Actual Profit: -$300 (-12.5%), Variance: -$700 | `NEGATIVE_MARGIN` |
| **BL-07** | Duplicate Invoice Prevention | Calling generate invoice twice on same shipment | Creates new version or updates existing draft | Returns active customer invoice |

---

## PART 15 — SHIPMENT CLOSURE AUDIT (4-RULE VERIFICATION)

| ID | Shipment Status | Compliance Docs Status | Carrier Invoices Status | Customer Invoice Status | Expected Audit Result | API Response |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **CL-01** | `IN_TRANSIT` | All `VERIFIED` | All `APPROVED` | `APPROVED` | **BLOCKED** (Rule 1 fails: cargo not delivered) | `400 Bad Request` |
| **CL-02** | `DELIVERED` | 0 Documents uploaded | All `APPROVED` | `APPROVED` | **BLOCKED** (Rule 2 fails: no compliance docs) | `400 Bad Request` |
| **CL-03** | `DELIVERED` | 1 document `PENDING_VERIFICATION` | All `APPROVED` | `APPROVED` | **BLOCKED** (Rule 2 fails: unverified doc) | `400 Bad Request` |
| **CL-04** | `DELIVERED` | 1 document has `OPEN` discrepancy | All `APPROVED` | `APPROVED` | **BLOCKED** (Rule 2 fails: open discrepancy) | `400 Bad Request` |
| **CL-05** | `DELIVERED` | All `VERIFIED` | 0 Carrier Invoices | `APPROVED` | **BLOCKED** (Rule 3 fails: no carrier invoice) | `400 Bad Request` |
| **CL-06** | `DELIVERED` | All `VERIFIED` | 1 invoice `DISCREPANCY` | `APPROVED` | **BLOCKED** (Rule 3 fails: unapproved carrier invoice) | `400 Bad Request` |
| **CL-07** | `DELIVERED` | All `VERIFIED` | All `APPROVED` | 0 Customer Invoices | **BLOCKED** (Rule 4 fails: no customer invoice) | `400 Bad Request` |
| **CL-08** | `DELIVERED` | All `VERIFIED` | All `APPROVED` | Customer invoice `DRAFT` | **BLOCKED** (Rule 4 fails: invoice not approved) | `400 Bad Request` |
| **CL-09** | `DELIVERED` | All `VERIFIED` (HBL, MBL) | All `APPROVED` | `APPROVED` or `PAID` | **PASSED** (All 4 conditions satisfied) | `200 OK` (Shipment status → `CLOSED`) |

---

## PART 16 — FRONTEND TESTING & MANUAL QA CHECKLIST

*   [ ] **1. Authentication:** Open `http://localhost:5173/login`. Check that dev mock login button / `test-token` allows immediate access to the dashboard.
*   [ ] **2. Lead & Inbound Email Inspection:** Navigate to `/dashboard/leads`. Open the detail drawer for a lead. Verify that simulated email threads show the extracted intent and AI summary.
*   [ ] **3. RFQ & Pricing Review:** Open an RFQ in `QUOTE_GENERATED` stage. Confirm carrier options (`Maersk`, `MSC`) render with buy price, sell price, margin %, and AI reasoning. Click **Approve Quote** and confirm RFQ moves to `WON`.
*   [ ] **4. Shipment Operations Timeline:** Go to `/dashboard/shipments`. Open the newly created shipment. Verify milestone tracker (`BOOKED`, `DEPARTED`, `IN_TRANSIT`, `ARRIVED`, `DELIVERED`).
*   [ ] **5. Document Compliance Verification:** On the shipment detail page, switch to the **Documents** tab. Upload a sample Bill of Lading. Confirm that OCR extracted fields populate and that any intentional discrepancy renders with an action button to Resolve.
*   [ ] **6. Finance 3-Way Reconciliation:** Switch to the **Finance** tab. Upload a carrier invoice. Verify line item reconciliation against the accepted quote.
*   [ ] **7. Billing & Profitability:** Switch to the **Billing** tab. Click **Generate Invoice**. Verify the sell price, gross profit, and margin % KPI cards. Click **Approve Invoice**.
*   [ ] **8. Shipment Closure Gate:** Attempt to click **Close Shipment** before delivering the shipment; verify that the UI displays the blocking rule. Complete milestones to `DELIVERED` and verify that **Close Shipment** succeeds.

---

## PART 17 — COMPLETE HAPPY-PATH E2E API COLLECTION

```bash
# 1. Submit Inbound RFQ Email
curl -X POST http://localhost:8080/api/v1/emails/inbound \
  -H "Content-Type: application/json" \
  -d '{
    "from": "procurement@tata-exports.local",
    "to": "quotes@freel.in",
    "subject": "RFQ: 2x40GP INNSA to DEHAM EXW",
    "body": "Please quote 2x40GP from INNSA to DEHAM, EXW, 24000kg auto components ready 2026-09-20.",
    "message_id": "msg-e2e-001@tata.com",
    "thread_id": "thread-e2e-001"
  }'

# 2. Check RFQ & Quotes
curl -X GET http://localhost:8080/api/v1/rfqs \
  -H "Authorization: Bearer test-token"

# 3. Approve Quote (Auto-Creates Shipment)
curl -X PUT http://localhost:8080/api/v1/rfqs/1/quotes/1/approve \
  -H "Authorization: Bearer test-token"

# 4. Ingest Carrier Milestones
curl -X POST http://localhost:8080/webhooks/carriers/MAEU \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "evt-e2e-del-001",
    "carrier_scac": "MAEU",
    "booking_number": "BKG-1",
    "milestone_code": "DELIVERED",
    "event_time": "2026-10-10T14:00:00Z",
    "location": "DEHAM"
  }'

# 5. Upload Compliance Document
curl -X POST http://localhost:8080/api/v1/shipments/1/documents/upload \
  -H "Authorization: Bearer test-token" \
  -F "doc_type=HBL" \
  -F "file=@/path/to/sample_hbl.pdf"

# 6. Ingest & Approve Carrier Invoice
curl -X POST http://localhost:8080/api/v1/shipments/1/finance/invoices/upload \
  -H "Authorization: Bearer test-token" \
  -F "invoice_number=INV-MAEU-991" \
  -F "vendor_name=Maersk Line" \
  -F "total_amount=2000.00" \
  -F "currency=USD" \
  -F "file=@/path/to/carrier_invoice.pdf"

curl -X POST http://localhost:8080/api/v1/finance/invoices/{invoice_uuid}/approve \
  -H "Authorization: Bearer test-token"

# 7. Generate & Approve Customer Invoice
curl -X POST http://localhost:8080/api/v1/shipments/1/billing/invoices/generate \
  -H "Authorization: Bearer test-token"

curl -X POST http://localhost:8080/api/v1/billing/invoices/{customer_invoice_uuid}/approve \
  -H "Authorization: Bearer test-token"

# 8. Close Shipment
curl -X POST http://localhost:8080/api/v1/shipments/1/close \
  -H "Authorization: Bearer test-token"
```

---

## PART 18 — FAILURE, CONCURRENCY & RETRY TESTING

| Scenario | Injected Condition | Expected Behavior |
| :--- | :--- | :--- |
| **Python Worker Crash** | Process killed during graph execution | Task remains locked; retried automatically on reboot (`attempts < max_attempts`) |
| **Go Backend 503** | Server down during callback | Python catches HTTP failure, backs off (`5 * attempts` seconds), retries |
| **PostgreSQL Outage** | Database restarts | Connection pools auto-reconnect without crashing server or worker |
| **LLM Rate Limit (429)** | Gemini quota exceeded | `llm_factory.py` automatically fails over to OpenAI GPT-4o-mini |
| **Duplicate Webhook** | 5 duplicate webhooks sent concurrently | `shipment_processed_events` + unique constraint guarantees single execution |
| **Concurrent Workers** | Multiple Python worker processes running | `SELECT FOR UPDATE SKIP LOCKED` guarantees zero duplicate processing |

---

## PART 19 — MULTI-TENANT SECURITY AUDIT

| Endpoint | Org A Action on Org B Data | Expected Result |
| :--- | :--- | :--- |
| `GET /api/v1/rfqs/{id}` | Access Org B RFQ | `404 Not Found` (Zero Leakage) |
| `GET /api/v1/shipments/{id}` | Access Org B Shipment | `404 Not Found` (Zero Leakage) |
| `GET /api/v1/shipments/{id}/documents` | Access Org B Documents | `404 Not Found` (Zero Leakage) |
| `GET /api/v1/shipments/{id}/finance` | Access Org B Carrier Costs | `404 Not Found` (Zero Leakage) |
| `GET /api/v1/shipments/{id}/billing` | Access Org B Margins | `404 Not Found` (Zero Leakage) |
| `GET /internal/*` | Missing `X-LogisticsHQ-Service-Key` | `401 Unauthorized` |

---

## PART 20 — OBSERVABILITY & TRACE PROPAGATION

Trace propagation path for a single request across the distributed stack:

```
[React Client]
  │ Header: 'X-Correlation-ID: e2e-trace-991'
  ▼
[Go Middleware: TraceMiddleware]
  │ Extracts / Generates CorrelationID ──► Context Key: CorrelationIDKey
  │ Header: 'X-Correlation-ID: e2e-trace-991' set on Response
  ▼
[Go Service / Queue Producer]
  │ Writes CorrelationID into task payload: {"correlation_id": "e2e-trace-991", ...}
  ▼
[PostgreSQL: ai_processing_tasks]
  ▼
[Python Sidecar: QueueWorker]
  │ Extracts payload['correlation_id']
  │ Logs: "[AI Sidecar][Correlation ID: e2e-trace-991] Starting LangGraph pipeline..."
  │ Injects correlation_id into LangGraph State & Config metadata
  ▼
[LangGraph Agent Execution]
  ▼
[Python → Go Callback]
  │ Sends JSON payload containing "correlation_id": "e2e-trace-991"
  ▼
[Go Callback Handler]
  │ Logs completion with Correlation ID and audits activity
```

---

## PART 21 — MASTER TEST MATRIX

| ID | Module | Test Name | Type | Setup | Action | Expected Result | DB Verification | Pri |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **TM-01** | RBAC | Dev Auth Bypass | Unit/Integ | Postgres seeded | `GET /api/v1/companies` with `Bearer test-token` | Returns 200 OK with User 1, Org 1 context | `users`, `org_members` checked | **P0** |
| **TM-02** | RBAC | Invalid Token Block | Security | None | Request with `Bearer bad-token` | Returns 401 Unauthorized | No DB queries run | **P0** |
| **TM-03** | Email | Inbound RFQ Ingestion | Integration | Clean DB | `POST /api/v1/emails/inbound` | Queues `EMAIL_PARSE` task; logs interaction | `lead_interactions`, `ai_processing_tasks` | **P0** |
| **TM-04** | Sales | SalesAgent Parsing | E2E | Sidecar running | Queue worker executes `EMAIL_PARSE` | Extracts cargo, calls `CreateRFQFromEmail`, returns callback | `rfqs`, `rfq_items` created | **P0** |
| **TM-05** | Pricing | Rate Search & Calculation | Integration | Seed rates | `GET /api/v1/rates/search?origin=INNSA&dest=DEHAM` | Returns matching canonical rates in USD | `rate_entries` queried | **P0** |
| **TM-06** | Pricing | PricingAgent Graph | E2E | RFQ created | Sidecar executes `PRICING_ANALYZE` | Calculates markup, creates draft quote | `rfq_quotes` created | **P0** |
| **TM-07** | Pricing | Quote Approval Flow | Integration | Quote draft | `PUT /api/v1/rfqs/{id}/quotes/{qid}/approve` | Quote status `APPROVED`, RFQ stage `QUOTE_SENT` | `rfqs.stage = 'QUOTE_SENT'` | **P0** |
| **TM-08** | RFQ | RFQ Won Shipment Auto-Create | Integration | Approved quote | Mark RFQ `WON` | Synchronous event bus creates shipment record | `shipments` row created with RFQ link | **P0** |
| **TM-09** | Ops | Carrier Webhook Ingest | Integration | Shipment exists | `POST /webhooks/carriers/MAEU` | Enqueues `CARRIER_UPDATE_PARSE` task | `carrier_tracking_events` inserted | **P0** |
| **TM-10** | Ops | Milestone Update | E2E | Task queued | Sidecar executes `OperationsAgent` | Calls Go internal API; marks milestone COMPLETED | `shipment_milestones.status = 'COMPLETED'` | **P0** |
| **TM-11** | Ops | Exception Detection | E2E | Delay payload | Ingest webhook with delay text | Exception created; severity `WARNING` | `shipment_exceptions` created | **P1** |
| **TM-12** | Docs | Document Upload & Task | Integration | Shipment active | `POST /api/v1/shipments/{id}/documents/upload` | Saves file locally; enqueues `DOC_VERIFY` task | `shipment_documents` inserted | **P0** |
| **TM-13** | Docs | Compliance Reconciliation | E2E | Task queued | Sidecar executes `ComplianceAgent` | Reconciles HBL vs MBL; calls callback | `shipment_documents.status = 'VERIFIED'` | **P0** |
| **TM-14** | Docs | Discrepancy Flagging | E2E | Mismatched docs | Upload HBL with different weight | Flags discrepancy; status `DISCREPANCY` | `shipment_document_discrepancies` row created | **P1** |
| **TM-15** | Docs | Discrepancy Resolve | Integration | Discrepancy open| `POST /api/v1/shipments/discrepancies/{id}/resolve` | Discrepancy marked `RESOLVED` | `shipment_document_discrepancies.status = 'RESOLVED'` | **P1** |
| **TM-16** | Finance | Invoice Ingest & OCR | Integration | Shipment active | `POST /api/v1/shipments/{id}/finance/invoices/upload` | Ingests invoice; enqueues `BILL_RECONCILE` | `shipment_invoices` inserted | **P0** |
| **TM-17** | Finance | 3-Way Reconciliation | E2E | Task queued | Sidecar executes `FinanceAgent` | Compares items with quote; flags variances | `shipment_invoices.status = 'APPROVED'` | **P0** |
| **TM-18** | Finance | Invoice Approval | Integration | Invoice pending | `POST /api/v1/finance/invoices/{id}/approve` | Invoice marked `APPROVED`; updates profitability | `shipment_invoices.status = 'APPROVED'` | **P0** |
| **TM-19** | Billing | Customer Invoice Gen | Integration | Shipment active | `POST /api/v1/shipments/{id}/billing/invoices/generate` | Creates invoice breakdown with proportional markup | `shipment_customer_invoices`, `items` created | **P0** |
| **TM-20** | Billing | Margin Calculation | Integration | Invoices ready | Recalculate profitability | Computes actual profit, variance, status | `shipment_finance_profitability` updated | **P0** |
| **TM-21** | Closure | Blocker Enforcement | Integration | Shipment not delivered | `POST /api/v1/shipments/{id}/close` | Fails with 400 Bad Request; closure blocked | `shipments.status` unchanged | **P0** |
| **TM-22** | Closure | Complete Closure | Integration | All 4 rules pass | `POST /api/v1/shipments/{id}/close` | Returns 200 OK; status updated to `CLOSED` | `shipments.status = 'CLOSED'` | **P0** |
| **TM-23** | Security | Cross-Org Isolation | Security | Org A / Org B | Org A requests Org B shipment | Returns 404 Not Found; zero data leak | Verified in DB | **P0** |
| **TM-24** | Resilience | Python Retry Backoff | Resilience | Sidecar offline | Enqueue task; verify attempt counter | Task retried with backoff; marked FAILED on max | `ai_processing_tasks.attempts` increments | **P1** |
| **TM-25** | UI | Full Manual UI Walkthrough | Manual UI | All services up | Operator walks through Leads → RFQ → Shipment → Close | UI reacts smoothly, modals work, tables update | Full DB lifecycle completed | **P0** |

---

## PART 22 — EXACT DAY-1 EXECUTION SEQUENCE

```bash
# STEP 1: Start PostgreSQL
docker run -d --name freel-postgres \
  -e POSTGRES_USER=user \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=freel \
  -p 5432:5432 postgres:16-alpine

# STEP 2: Apply Migrations
cd backend
for file in $(ls internal/database/migrations/*.sql | sort); do
  psql "postgres://user:password@localhost:5432/freel?sslmode=disable" -f "$file"
done

# STEP 3: Seed Database
go run ./cmd/seed

# STEP 4: Start Go Backend Server (Port 8080)
go run ./cmd/server

# STEP 5: Start Python AI Sidecar (Port 8090)
cd ../ai_sidecar
source venv/bin/activate
uvicorn main:app --port 8090

# STEP 6: Start React Frontend (Port 5173)
cd ../frontend
npm run dev

# STEP 7: Run Automated Test Suites
cd ../backend && go test ./...
cd ../frontend && npm test
```

---

## PART 23 — PREREQUISITE AUDIT CHECKLIST

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      FINAL PREREQUISITE STATUS AUDIT                        │
├──────────────────────────────────────┬──────────────────────────────────────┤
│ ITEM                                 │ AUDIT STATUS                         │
├──────────────────────────────────────┼──────────────────────────────────────┤
│ Git, Go 1.25, Python 3.11+, Node 20+ │ REQUIRED NOW                         │
│ PostgreSQL 16 Instance               │ REQUIRED NOW                         │
│ Database Migrations (001 to 018)     │ REQUIRED NOW                         │
│ Database Seed Execution (cmd/seed)   │ REQUIRED NOW                         │
│ Go Backend Running (Port 8080)       │ REQUIRED NOW                         │
│ Python Sidecar Running (Port 8090)   │ REQUIRED NOW                         │
│ React Frontend Running (Port 5173)   │ REQUIRED NOW                         │
│ INTERNAL_SERVICE_TOKEN Configured    │ REQUIRED NOW                         │
│ Google Gemini API Key                │ OPTIONAL NOW (Has mock fallback)     │
│ OpenAI API Key                       │ OPTIONAL NOW (Has mock fallback)     │
│ Tavily API Key                       │ OPTIONAL NOW (Has DDG fallback)      │
│ LangSmith API Key                    │ OPTIONAL NOW                         │
│ Local Storage (./uploads)            │ REQUIRED NOW                         │
│ AWS S3 Bucket                        │ REQUIRED LATER (Prod only)           │
│ AWS Cognito User Pool                │ REQUIRED LATER (test-token for dev)  │
│ AWS SES Inbound Email                │ NOT REQUIRED (Direct webhook active) │
│ AWS SQS / Kafka                      │ NOT REQUIRED (Postgres queue active) │
│ Real Carrier Direct APIs             │ NOT REQUIRED (Mock adapter active)   │
└──────────────────────────────────────┴──────────────────────────────────────┘
```
