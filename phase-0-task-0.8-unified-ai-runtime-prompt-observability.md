# Phase 0 — Task 0.8: Unified AI Runtime, Model Configuration, Prompt Management, and Observability Foundation

**Date:** 2026-09-06  
**Status:** Completed & Validated  
**Scope:** Go Backend (`ai.Gateway`, `PromptManager`, Action Engine), Python AI Sidecar (`llm_factory.py`, `prompt_registry.py`, `runtime_config.py`, `tracer.py`), MariaDB Persistence (`ai_prompt_templates`, `ai_execution_traces`, `audit_logs`), Frontend Governance Modals.

---

## 1. Files Inspected

### Python AI Sidecar
- `ai_sidecar/app/tools/llm_factory.py`: Inspected primary/fallback initialization, provider wrappers, mock fallback heuristics, tool-binding mechanics, and error propagation.
- `ai_sidecar/app/agents/llm_utils.py`: Inspected direct LLM invocations, embedded prompt templates, JSON parsing logic, and string match mock dictionaries.
- `ai_sidecar/app/agents/nodes/pricing_agent_node.py`: Inspected LangGraph node definitions, tool schemas, and embedded pricing analyst prompts.
- `ai_sidecar/app/agents/nodes/ops_parse_node.py`: Inspected carrier milestone and tracking update parsing prompt logic.
- `ai_sidecar/app/agents/parser_agent.py`: Inspected email extraction prompts and port normalization flows.
- `ai_sidecar/app/agents/email_parser_agent.py`: Inspected inbound classification, intent parsing, and reply drafting prompts.
- `ai_sidecar/app/graphs/pricing_graph.py`, `contracts_graph.py`, `sales_graph.py`: Inspected state transitions, persistent checkpoint integration, and interrupt gates.
- `ai_sidecar/app/persistence/mariadb_saver.py`: Inspected connection pooling and checkpoint schema compliance.

### Go Backend
- `backend/internal/ai/gateway.go`: Inspected `Gateway` struct, provider registry, `ExecutePrompt` pipeline, and timeout enforcement.
- `backend/internal/ai/prompt_manager.go`: Inspected hardcoded prompt maps, template interpolation, and workflow isolation.
- `backend/internal/ai/gemini.go`: Inspected Gemini SDK calls, API key extraction, and model selection.
- `backend/internal/ai/chatgpt.go`: Inspected OpenAI chat completion wrapper and error parsing.
- `backend/cmd/server/main.go`: Inspected server bootstrapping, environment variable ingestion, and service wiring.

### Database & Migrations
- `database/migrations/`: Audited migrations 001 through 086, particularly 081 (`universal_audit_logs`), 084 (`ai_task_queue`), and 086 (`langgraph_checkpoints`).

---

## 2. Existing Architecture Findings

1. **Dual Competing AI Execution Paths:**
   - Go-native workflows (Lead Scoring, Outreach) utilized `ai.Gateway` with direct HTTP/SDK calls.
   - Python FastAPI sidecar utilized LangGraph state graphs with separate ChatModel instances.
   - Neither shared runtime configuration, error handling contracts, or execution telemetry.
2. **Unverified Model Identifiers:**
   - Both Go and Python hardcoded `gemini-3.1-flash-lite`, an unverified identifier that produced API warnings or unpredictable routing.
3. **Silent Mock Fallback Hazard in Production:**
   - `llm_factory.py` contained fallback heuristics that silently caught any LLM error (including quota exhaustion or auth failures) and returned hardcoded pricing recommendations (`$2,800 buy / $3,220 sell`) for all inquiries, without checking whether the environment was production.
4. **Embedded, Fragmented Prompts:**
   - System prompts and extraction guidelines were scattered across Python source files (`pricing_agent_node.py`, `parser_agent.py`, `email_parser_agent.py`, `ops_parse_node.py`) as multi-line strings, preventing centralized versioning, audit tracking, or runtime overrides.
5. **Inconsistent SDK Response Formats:**
   - The Google GenAI SDK frequently returned `AIMessage.content` as a list of content blocks (`[{'type': 'text', 'text': ...}]`), whereas OpenAI returned plain strings, causing silent type crashes in downstream JSON parsers.
6. **Missing Observability & Secret Leakage Risks:**
   - AI execution traces were not recorded to database tables; API keys (`AIzaSy...`, `sk-proj-...`) were vulnerable to appearing in raw error logs during timeout or 429 exceptions.

---

## 3. Configuration Changes

### Go Backend (`backend/internal/ai/config.go`)
- Introduced authoritative `RuntimeConfig` struct with strict validation:
  - `Environment`: Differentiates development from production.
  - `PrimaryProvider` (`gemini`) & `PrimaryModel` (`gemini-1.5-flash`).
  - `FailoverProvider` (`openai`) & `FailoverModel` (`gpt-4o-mini`).
  - `AllowMockFallback`: Strictly prohibited in production (`is_production=true`). Startup fails fast if mock mode is requested in production.
  - `ProviderTimeout` (default 30s) & `MaxRetries` (default 2).
  - Telemetry-safe string representation redacting sensitive API keys.

### Python Sidecar (`ai_sidecar/app/config/runtime_config.py`)
- Created `AIRuntimeConfig` mirroring the Go runtime contract:
  - Enforces `VERIFIED_MODELS` whitelist (`gemini-1.5-flash`, `gemini-1.5-pro`, `gemini-2.0-flash`, `gpt-4o`, `gpt-4o-mini`).
  - Rejects deprecated or unverified model identifiers at instantiation.
  - Rejects missing primary API keys in production with explicit `ValueError`.
  - Rejects `allow_mock_fallback=True` in production environments with `CRITICAL SECURITY VIOLATION`.
  - Provides `to_safe_dict()` for telemetry and health check endpoints.

---

## 4. Provider and Model Changes

1. **Model Standardization:**
   - Migrated all primary Gemini invocations from `gemini-3.1-flash-lite` to verified, high-throughput `gemini-1.5-flash`.
   - Migrated failover OpenAI invocations to `gpt-4o-mini`.
2. **Unified Failover Chat Model (`ai_sidecar/app/tools/llm_factory.py`):**
   - Implemented `UnifiedFailoverChatModel` wrapping LangChain base models for both sync (`invoke`) and async (`ainvoke`).
   - Normalizes response content across SDKs via `normalize_ai_content()` (flattens content block lists to strings).
   - Enriches `AIMessage.response_metadata` with `primary_provider`, `final_provider`, `failover_occurred`, and redacted `failover_reason`.
   - Explicitly logs provider failover events with classified error categories.
   - Enforces mock fallback boundaries: mock mode triggers **only** when `allow_mock_fallback=True` in non-production environments. In production or strict mode, throws `RuntimeError("All configured AI providers failed. Mock fallback is disabled.")`.

---

## 5. Prompt Management Changes

### MariaDB Schema Migration (`087_unified_ai_runtime_prompts_observability.sql`)
- Created table `ai_prompt_templates`:
  - `prompt_key` VARCHAR(100), `version` VARCHAR(20), `workflow_name` VARCHAR(50), `purpose` VARCHAR(255), `template` MEDIUMTEXT, `input_variables` JSON, `output_format` VARCHAR(50), `is_active` TINYINT(1).
  - Unique composite index `(prompt_key, version)`.

### Seeded Canonical Prompt Templates (v1.0.0)
Seeded 10 production prompt templates across all 8 business workflows:
1. `pricing.analyst`: Senior Pricing Analyst Agent system prompt.
2. `sales.email_classification`: Inbound customer email classifier and entity extractor.
3. `sales.reply_draft`: Clarification reply drafting prompt for incomplete customer inquiries.
4. `operations.tracking_parse`: Unstructured carrier tracking milestone and anomaly extractor.
5. `contracts.rate_extraction`: Ocean and air contract rate card extractor.
6. `compliance.doc_verification`: Shipping document discrepancy auditor.
7. `finance.invoice_audit`: Carrier vendor invoice reconciliation prompt.
8. `leads.lead_scoring`: Prospective shipper commercial fit and potential evaluator.
9. `outreach.cold_email`: Personalized B2B freight cold outreach generator.
10. `rfq.extract_shipment_request`: Shipment parameter extraction from raw text.

### Registry Implementations
- **Go (`backend/internal/ai/prompt_manager.go`):** Upgraded `PromptManager` to resolve canonical dot-notation keys (`pricing.analyst`, `leads.lead_scoring`), supporting legacy aliases (`score_lead`, `generate_email`) for backward compatibility.
- **Python (`ai_sidecar/app/prompts/prompt_registry.py`):** Created `PromptRegistry` with `get_prompt(key, version, vars)`, `get_prompt_definition()`, and `list_prompts_for_workflow()`.
- **Node Migrations:** Replaced embedded strings in `pricing_agent_node.py`, `ops_parse_node.py`, `parser_agent.py`, and `email_parser_agent.py` with calls to `get_prompt()`.

---

## 6. Execution-Context Changes

Standardized `ExecutionContext` across Go and Python:
- **Tenant Isolation:** Requires `org_id` on all execution requests.
- **Actor Attribution:** Preserves `actor_type=AI_AGENT` and `acting_user_id`.
- **Tracing Context:** Tracks `task_id`, `thread_id`, `request_id`, and `correlation_id`.
- **Workflow Metadata:** Scopes prompt key, prompt version, primary provider, final provider, and duration.
- **Thread Safety:** Context is passed explicitly per call; no global mutable state.

---

## 7. Observability and Redaction Changes

### Database Telemetry (`ai_execution_traces`)
- Recorded per-invocation metadata: `org_id`, `task_id`, `thread_id`, `request_id`, `correlation_id`, `workflow_name`, `prompt_key`, `prompt_version`, `primary_provider`, `final_provider`, `failover_occurred`, `failover_reason`, `is_mock`, `status`, `duration_ms`, `error_category`, `error_message`, and `execution_metadata`.

### Secret Redaction Engine
- Implemented `redact_secrets()` in Go (`backend/internal/ai/errors.go`) and Python (`ai_sidecar/app/tools/llm_factory.py`):
  - Redacts OpenAI API keys (`sk-[a-zA-Z0-9_-]{20,}` -> `[REDACTED_OPENAI_KEY]`).
  - Redacts Google/Gemini API keys (`AIza[0-9A-Za-z-_]{15,}` -> `[REDACTED_GEMINI_KEY]`).
  - Redacts internal tokens and secrets (`lhq_sec_...` -> `[REDACTED_SECRET]`).

### Universal Audit Log Integration
- Connected `AIObservabilityTracer.record_lifecycle_event()` and fatal execution failures to `audit_logs` table:
  - `actor_type`: `AI_AGENT`.
  - `module`: `AI_RUNTIME`.
  - `action`: `AI_EXECUTION_STARTED`, `PROVIDER_FAILOVER`, `AI_EXECUTION_FAILED`.
  - Redacts sensitive payload parameters before writing to audit logs.

---

## 8. Error Taxonomy

Standardized across Go, Python, Queue Workers, and Audit Logs:

| Error Category | Description | Retryable |
| :--- | :--- | :---: |
| `configuration_error` | Missing required config, invalid model, invalid provider | No |
| `authentication_error` | Invalid API keys, expired tokens (HTTP 401) | No |
| `authorization_error` | Insufficient permissions, forbidden access (HTTP 403) | No |
| `organization_isolation_error` | Multi-tenant boundary violation, org mismatch | No |
| `validation_error` | Invalid payload schema, missing mandatory variables | No |
| `provider_rate_limit` | LLM provider quota/rate limits (HTTP 429) | Yes |
| `provider_timeout` | Gateway or provider deadline exceeded (HTTP 504) | Yes |
| `provider_unavailable` | Provider service outage (HTTP 502/503) | Yes |
| `model_error` | Context window exceeded, bad output format | No |
| `tool_error` | Action tool execution error | Context-dependent |
| `action_rejected` | Centralized action system rejection | No |
| `approval_required` | Execution paused for human sign-off | No (Workflow paused) |
| `checkpoint_error` | Checkpoint persistence read/write failure | Yes |
| `task_retryable` | Generic retryable background worker failure | Yes |
| `task_permanent_failure` | Maximum retries exhausted or fatal business error | No |
| `unknown_error` | Unclassified runtime exceptions | No |

---

## 9. Integration Points

1. **Go Backend Server (`backend/cmd/server/main.go`):**
   - Bootstraps `ai.RuntimeConfig` and binds MySQL DB to `ai.NewGatewayWithConfig(db, aiRuntimeConfig)`.
2. **Centralized Action System:**
   - Retains `actor_type=AI_AGENT` execution context and links AI tasks to audit logs.
3. **Python AI Sidecar (`ai_sidecar/app/tools/llm_factory.py`):**
   - Provides unified chat models with observable failover and persistent MariaDB checkpointer.
4. **LangGraph Graphs:**
   - Pricing, Sales, Contracts, and Operations graphs consume versioned prompts from `prompt_registry`.
5. **Frontend Governance (`ApprovalDetailsModal.jsx`):**
   - Displays AI runtime state, provider model badges (`Gemini 1.5 Flash`, `Failover OpenAI`, `Mock Dev`), prompt version, and safe diagnostic categories.

---

## 10. Tests Added and Updated

1. **Go Unit Tests (`backend/internal/ai/ai_runtime_test.go`):**
   - `TestRuntimeConfig_Validation`: Dev configuration, production mock mode rejection, missing key rejection, invalid model rejection.
   - `TestPromptManager_VersioningAndAliases`: Canonical key resolution, legacy alias resolution, loud failure on missing prompt, workflow listing.
   - `TestErrorClassificationAndRedaction`: Secret redaction, rate limit classification, gateway timeout classification, unavailable classification.
   - `TestGateway_FailoverBehavior`: Primary success, primary failure triggering failover, both providers failing without mock fallback returning classified error.
2. **Python Test Suite (`ai_sidecar/test_ai_runtime_observability.py`):**
   - 25 automated test cases covering config validation, mock mode prohibition, prompt registry resolution, variable substitution, provider failover, content normalization, secret redaction, telemetry recording, audit log writing, and smoke tests across all 8 workflows.
3. **Functional Baseline Suite (`ai_sidecar/run_baseline_functional_suite.py`):**
   - End-to-end multi-agent evaluation on persistent MariaDB database.

---

## 11. Test Results

### Go Unit Tests
```
=== RUN   TestRuntimeConfig_Validation
--- PASS: TestRuntimeConfig_Validation (0.00s)
=== RUN   TestPromptManager_VersioningAndAliases
--- PASS: TestPromptManager_VersioningAndAliases (0.00s)
=== RUN   TestErrorClassificationAndRedaction
--- PASS: TestErrorClassificationAndRedaction (0.00s)
=== RUN   TestGateway_FailoverBehavior
--- PASS: TestGateway_FailoverBehavior (0.00s)
PASS
ok  	github.com/freel/backend/internal/ai	0.637s
```

### Python Runtime Test Suite (`test_ai_runtime_observability.py`)
```
test_invalid_model_identifier_rejected (__main__.TestAIRuntimeConfig) ... ok
test_production_rejects_missing_primary_key (__main__.TestAIRuntimeConfig) ... ok
test_production_rejects_mock_fallback (__main__.TestAIRuntimeConfig) ... ok
test_safe_dict_redacts_keys (__main__.TestAIRuntimeConfig) ... ok
test_valid_development_config (__main__.TestAIRuntimeConfig) ... ok
test_record_execution_trace (__main__.TestObservabilityAndAuditIntegration) ... ok
test_record_lifecycle_event_to_audit_logs (__main__.TestObservabilityAndAuditIntegration) ... ok
test_all_eight_workflows_covered (__main__.TestPromptRegistry) ... ok
test_missing_prompt_fails_loudly (__main__.TestPromptRegistry) ... ok
test_pricing_analyst_system_prompt (__main__.TestPromptRegistry) ... ok
test_prompt_metadata_integrity (__main__.TestPromptRegistry) ... ok
test_resolve_canonical_prompt_with_variables (__main__.TestPromptRegistry) ... ok
test_error_classification (__main__.TestProviderBehaviorAndFailover) ... ok
test_failover_success_when_primary_fails (__main__.TestProviderBehaviorAndFailover) ... ok
test_mock_fallback_rejected_when_disallowed (__main__.TestProviderBehaviorAndFailover) ... ok
test_normalize_ai_content (__main__.TestProviderBehaviorAndFailover) ... ok
test_secret_redaction (__main__.TestProviderBehaviorAndFailover) ... ok
test_compliance_smoke (__main__.TestWorkflowSmokeTests) ... ok
test_contracts_smoke (__main__.TestWorkflowSmokeTests) ... ok
test_finance_smoke (__main__.TestWorkflowSmokeTests) ... ok
test_lead_scoring_smoke (__main__.TestWorkflowSmokeTests) ... ok
test_operations_smoke (__main__.TestWorkflowSmokeTests) ... ok
test_outreach_smoke (__main__.TestWorkflowSmokeTests) ... ok
test_pricing_smoke (__main__.TestWorkflowSmokeTests) ... ok
test_sales_smoke (__main__.TestWorkflowSmokeTests) ... ok

----------------------------------------------------------------------
Ran 25 tests in 0.464s

OK
```

### Baseline Functional Suite (`run_baseline_functional_suite.py`)
```
===========================================================================
 BASELINE FUNCTIONAL SUITE EXECUTION SUMMARY
===========================================================================
 [PASS] Pricing Agent — Normal RFQ           | Target: RFQ #102           | Rates retrieved and RFQ agent_status=DRAFT_READY
 [PASS] Pricing Agent — Anomaly HITL         | Target: RFQ #105           | Paused at interrupt ('save',) and saved 12 checkpoints in MariaDB.
 [PASS] Contracts Extraction & Rate Anomaly  | Target: Doc #101           | Paused before ('ingest',), state reloaded successfully from MariaDB.
 [PASS] Sales / Email Parser — Incomplete Email | Target: Interaction #119   | Draft #26 staged, Approval #139 generated, 0 autonomous sends.
 [PASS] Sales / Email Parser — Complete Email | Target: Lead #103          | RFQ created successfully from email payload.
 [PASS] Operations / Carrier Tracking        | Target: Shipment #102      | Shipment retrieved (status=IN_TRANSIT), milestone updated.
 [PASS] Compliance Reconciliation            | Target: Shipment #101      | Verified 2 compliance document discrepancies in database.
 [PASS] Finance Invoice Reconciliation       | Target: Invoice #101       | Invoice verified, 4 finance discrepancies tracked.
 [PASS] Lead Scoring Worker                  | Target: Lead #102          | Lead AI Score=0, status=CONVERTED.
===========================================================================
```

### Frontend Build Verification
```
vite v8.0.12 building client environment for production...
transforming...✓ 3087 modules transformed.
rendering chunks...
✓ built in 21.15s
```

---

## 12. Remaining Limitations

1. **OpenAI Quota Status:** The development environment's configured OpenAI key currently returns HTTP 429 (`insufficient_quota`). Failover logic correctly catches this error, logs the diagnostic category `provider_rate_limit`, and switches to explicit mock mode only in development.
2. **Live Dynamic Prompt Editing UI:** While prompt templates are now completely database-backed and versioned in `ai_prompt_templates`, an administrative UI for editing prompt templates at runtime is scheduled for subsequent phases.

---

## 13. Configuration and Migration Steps for Deployment

1. **Apply Migration 087:**
   ```bash
   mysql -u <user> -p<password> freel_mysql < database/migrations/087_unified_ai_runtime_prompts_observability.sql
   ```
2. **Seed Prompt Templates:**
   ```bash
   python ai_sidecar/app/prompts/seed_prompt_templates.py
   ```
3. **Environment Configuration (`.env`):**
   - Set `APP_ENV=production` in production.
   - Configure `GOOGLE_API_KEY` with a verified Gemini API key.
   - Configure `OPENAI_API_KEY` for failover provider support.
   - Ensure `ALLOW_MOCK_FALLBACK` is NOT set to `true` in production (validated and enforced at startup).
   - Verify `GEMINI_MODEL=gemini-1.5-flash` and `OPENAI_MODEL=gpt-4o-mini`.
4. **Compile & Restart Services:**
   - Recompile Go binary: `go build -o server.exe ./cmd/server` and restart backend.
   - Restart Python AI sidecar: `uvicorn main:app --host 0.0.0.0 --port 8090`.
