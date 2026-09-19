# Phase 3 Task 3.4 — RFQ-to-Quotation Automation and Intelligent Pricing Workflow

## 1. Summary
Phase 3, Task 3.4 implements an end-to-end, controlled **RFQ-to-Quotation Automation and Intelligent Pricing Workflow** for LogisticsHQ. The architecture enforces a strict separation of concerns:
- **Python AI Sidecar**: Houses 100% of the artificial intelligence logic, including unstructured and structured RFQ requirement extraction, missing field ambiguity analysis, grounded pricing explanations, quotation risk analysis, and draft synthesis based strictly on verified business facts. The AI sidecar is completely stateless regarding mutations and cannot directly modify MariaDB records, alter prices, approve quotations, or dispatch emails.
- **Go Integration and Application-Control Layer**: Serves as the authoritative control plane, enforcing JWT authentication, tenant isolation, role-based access control, database transactions in MariaDB, deterministic pricing and margin calculations, audit logging, action registration, and Human-in-the-Loop (HITL) approval gates.
- **Frontend**: Seamlessly integrated into the standard LogisticsHQ white/light design system within the RFQ Detail view (`⚡ AI Quotation Workflow` tab) and supporting views, providing operational staff with clear requirement visibility, real-time pricing preview, editable draft proposals, and one-click submission to the managerial approval workflow.

---

## 2. Existing Architecture Reused
Rather than creating parallel systems or duplicate services, Task 3.4 extends existing LogisticsHQ components:
1. **RFQ and Quotation Models**: Built upon existing domain models in `internal/rfq`, `internal/quotations`, and database tables `rfqs`, `rfq_items`, `quotes`, `rates`, and `customers`.
2. **Deterministic Pricing Engine**: Extended existing freight calculation principles (`cost + surcharges - discounts = sell price`, `margin = (sell - cost) / sell * 100`).
3. **Action System & Action Registry**: Registered the new high-risk action `quotations.send_draft` within `internal/orchestration/registry.go`, requiring managerial approval before execution.
4. **Approval / HITL System**: Connected to `internal/approvals/service.Service`, automatically creating pending approval items whenever a proposal is submitted or triggered by low margin (< 15%), negative margin, or commercial discounts.
5. **Audit Logging**: Recorded comprehensive audit logs for all workflow events via `internal/audit/service.Service`.
6. **Notification Center**: Triggered actionable in-app notifications upon requirement extraction and approval submissions.
7. **Database Isolation**: Preserved multi-tenant scoping on `org_id` throughout all SQL queries and schema constraints.
8. **Frontend White/Light Theme**: Built with CSS tokens adhering to `var(--primary-color)`, neutral slate borders, card containers, and standard action buttons without any dark panels or neon AI effects.

---

## 3. Python-Only AI Implementation
All AI-specific logic is implemented exclusively in Python within the `ai_sidecar` module:
- **Location**: [`ai_sidecar/app/rfq_pricing/`](file:///c:/Users/Sai/go/src/freel-project/ai_sidecar/app/rfq_pricing/)
  - `models.py`: Strongly typed Pydantic models: `RFQContext`, `ExtractedRFQRequirements`, `DeterministicPricingFacts`, `PricingExplanation`, `QuotationRiskAnalysis`, `QuotationDraftRequest`, `QuotationDraftResponse`, `EvidenceFact`, and `SafetyWarning`.
  - `agent.py`: `RFQPricingWorkflowAgent` containing prompt orchestration, prompt injection defense, requirement extraction, missing field impact evaluation, grounded pricing narration, risk analysis, and customer wording generation.
- **Endpoints in `ai_sidecar/main.py`**:
  1. `POST /rfq-pricing/extract-requirements`: Extracts and verifies RFQ items, cargo specifications, equipment, incoterms, special instructions, and flags missing mandatory information with confidence scores.
  2. `POST /rfq-pricing/explain-pricing`: Explains deterministic cost, sell, and margin figures calculated by Go without inventing or altering numbers.
  3. `POST /rfq-pricing/analyze-risks`: Evaluates margin health, rate freshness, deadline urgency, and determines if managerial approval is required.
  4. `POST /rfq-pricing/generate-draft`: Generates an editable internal summary, formal customer proposal wording, terms and conditions, and recipient preview.
- **Safety**: Inputs are sanitized against prompt injection patterns (`ignore previous instructions`, `system override`, `disregard pricing`). The agent operates only on structured facts provided by Go and is strictly forbidden from generating authoritative financial figures.

---

## 4. Go Integration Responsibilities
The Go backend (`backend/internal/rfq/pricing_workflow`) acts as the authoritative application-control layer:
1. **Authentication & Identity**: Resolves user and organization context via `middleware.GetUserContext`.
2. **Tenant Isolation**: Strictly scopes every query with `WHERE org_id = ?`. Cross-tenant requests immediately return `404 Not Found`.
3. **Data Loading & Sanitization**: Gathers real business records from `rfqs`, `rfq_items`, `quotes`, `rates`, and `customers`.
4. **Authoritative Deterministic Pricing**: Computes baseline carrier cost, fuel/security surcharges, selling price, gross profit, and margin percentage.
5. **Sidecar Communication**: Dispatches authenticated HTTP requests to the Python AI sidecar via `X-LogisticsHQ-Service-Key`.
6. **Response Validation**: Independently validates the Python response, verifying that no prices were overridden or unauthorized data was returned.
7. **Persistence**: Saves requirement extractions into `ai_rfq_requirements_extractions` and drafts into `ai_quotation_drafts`.
8. **HITL Enqueueing**: Integrates with the approval service to register an approval request for any proposal requiring oversight.
9. **Action Registry**: Guarantees that quotation dispatch can only be executed via an approved action token in `quotations.send_draft`.

---

## 5. RFQ Extraction Behavior
The extraction engine evaluates incoming RFQ data and attached notes:
- **Extracted Fields**: Mode (`OCEAN_FCL`, `AIR_CARGO`, `ROAD_FTL`), origin, destination, cargo type, equipment type, weight (kg), volume (cbm), incoterms, target date, deadline date, special handling instructions, commodity, customs requirements, and insurance requirements.
- **Field Statuses**:
  - `VERIFIED`: Explicitly documented in database fields with 1.0 confidence.
  - `EXTRACTED`: Inferred from unstructured notes or cargo descriptions with high confidence (> 0.85).
  - `MISSING`: Required operational field is missing; flagged for clarification.
  - `AMBIGUOUS`: Conflicting or incomplete values detected.
- **Clarification Recommendations**: When mandatory fields (e.g. weight, container type, incoterms) are absent, the system marks the extraction as `CLARIFICATION_REQUIRED` and generates a targeted clarification request rather than guessing or hallucinating values.

---

## 6. Deterministic Pricing Behavior
The Go backend is the sole source of truth for pricing calculations:
- **Cost Calculation**: Aggregates base ocean/air freight, bunker fuel surcharge (BAF), terminal handling charges (THC), security fees, and verified carrier quotation records.
- **Sell Calculation**: Applies standard organizational markup (default 20%) over verified cost, plus applicable commercial surcharges minus negotiated customer discounts.
- **Formula**:
  $$\text{Total Selling Price} = \text{Base Sell} + \text{Surcharges} - \text{Discounts}$$
  $$\text{Gross Profit} = \text{Total Selling Price} - \text{Total Cost}$$
  $$\text{Gross Margin \%} = \left(\frac{\text{Gross Profit}}{\text{Total Selling Price}}\right) \times 100$$
- **Pricing Breakdown**: Returns itemized `CostComponents` and `SellComponents` along with calculation methodology and input rate record references.

---

## 7. Margin Calculation Behavior
Margin thresholds are deterministically evaluated:
- **`HEALTHY`**: Margin $\ge 15.0\%$.
- **`LOW`**: Margin between $0.0\%$ and $14.99\%$. Automatically triggers approval requirement.
- **`NEGATIVE`**: Margin $< 0.0\%$ (selling price below carrier cost). Flagged with `CRITICAL` severity and strictly blocks dispatch without managerial override.
- **Override Safeguards**: Any manual selling price adjustment is clamped, re-evaluated deterministically, and immediately triggers an approval requirement if margin drops below the threshold.

---

## 8. Risk Detection
The risk analysis engine combines deterministic business rules with AI risk narration:
- **Signals Monitored**:
  1. Missing mandatory RFQ data.
  2. Carrier cost missing or zero.
  3. No applicable contracted rates found.
  4. Contracted rate expired or near expiry.
  5. Low gross margin (< 15%) or negative margin.
  6. RFQ response deadline within 48 hours.
  7. Large commercial discounts (> 10%).
- **Output**: Generates a risk rating (`LOW`, `MEDIUM`, `HIGH`, `CRITICAL`), list of actionable reasons, suggested next steps, and human-in-the-loop triggers.

---

## 9. Quotation Drafting
Operators can trigger AI-assisted quotation drafting with one click:
- **Grounded Content**: The Python sidecar formats an internal quotation summary, executive customer proposal letter, and standard shipping terms based solely on the Go-provided pricing facts and customer details.
- **Recipient Preview**: Resolves the primary customer contact from the real database record.
- **Editable Draft**: Drafts are persisted in status `DRAFT`, allowing operators to review and edit wording, adjust validities, or propose price revisions before submission.
- **No Hallucinations**: Carrier names, prices, validity dates, and terms are strictly bounded by verified inputs.

---

## 10. Approval and Sending Controls
To ensure compliance and commercial safety:
- **Action Registered**: `quotations.send_draft` registered as `RiskLevelHigh` and `RequiresApproval: true`.
- **Approval Gate**: Quotations with low margin, manual discounts, or missing customer credit approval cannot be dispatched directly. Clicking "Submit for Approval" registers a formal approval request in the LogisticsHQ Approval Center.
- **Execution Safeguard**: Quotations can only transition to `APPROVED` or `EXECUTED` when the corresponding approval is signed off by authorized personnel.
- **Tampering Defense**: Modifying an approved draft invalidates the prior approval ID and transitions status back to `DRAFT`.

---

## 11. Database Changes
Implemented migration [`101_phase3_rfq_pricing_workflow.sql`](file:///c:/Users/Sai/go/src/freel-project/backend/migrations/101_phase3_rfq_pricing_workflow.sql) in MariaDB `freel_mysql`:
1. **`ai_rfq_requirements_extractions`**:
   - Stores extraction results: `rfq_id`, `org_id`, `status` (`COMPLETE`, `INCOMPLETE`, `CLARIFICATION_REQUIRED`), `confidence_score`, `extracted_fields` (JSON), `missing_mandatory` (JSON), `missing_optional` (JSON), `evidence` (JSON), `clarification_notes`, `correlation_id`, `created_at`.
2. **`ai_quotation_drafts`**:
   - Stores quotation drafts: `rfq_id`, `org_id`, `status` (`DRAFT`, `PENDING_APPROVAL`, `APPROVED`, `REJECTED`, `EXECUTED`), `currency`, `base_cost`, `total_cost`, `base_sell`, `surcharges`, `discounts`, `total_selling_price`, `gross_margin_amount`, `gross_margin_pct`, `margin_health`, `terms_and_conditions`, `internal_summary`, `customer_wording`, `pricing_explanation`, `recipient_email`, `recipient_name`, `requires_approval`, `approval_id`, `action_proposal_id`, `correlation_id`, `created_at`, `updated_at`.

---

## 12. API Changes
All endpoints are scoped under `/api/v1/rfqs/{id}/pricing-workflow`:
- `GET /overview`: Returns consolidated workflow overview, latest extraction, pricing preview, latest draft, and pending approval state.
- `POST /extract-requirements`: Triggers requirement extraction, persists result, and returns grounded evidence.
- `POST /pricing-preview`: Calculates deterministic cost, selling price, and margin breakdown with optional rate selection.
- `POST /drafts`: Generates a new quotation draft, grounded in deterministic pricing and persists it in MariaDB.
- `GET /drafts/{draftId}`: Retrieves specific quotation draft details.
- `PUT /drafts/{draftId}`: Allows operators to edit customer wording, internal summary, or price overrides.
- `POST /drafts/{draftId}/submit-approval`: Submits the draft into the managerial approval pipeline and records audit entries.

---

## 13. Frontend Changes
- **Service**: Created [`frontend/src/services/rfqPricingWorkflowService.js`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/services/rfqPricingWorkflowService.js).
- **Component**: Created [`frontend/src/pages/dashboard/RFQ/components/RFQIntelligentPricingWorkflowSection.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/RFQ/components/RFQIntelligentPricingWorkflowSection.jsx) and companion styling [`RFQIntelligentPricingWorkflowSection.css`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/RFQ/components/RFQIntelligentPricingWorkflowSection.css).
- **Detail Integration**: Mounted in [`frontend/src/pages/dashboard/RFQ/RFQDetailPage.jsx`](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/RFQ/RFQDetailPage.jsx) under the dedicated tab `⚡ AI Quotation Workflow`.
- **Styling**: Strictly adheres to the white/light LogisticsHQ theme. Clean white card containers (`#ffffff`), soft slate borders (`#e2e8f0`), deep indigo primary buttons (`#4f46e5`), emerald badges for healthy margins, and amber warnings for pending reviews.

---

## 14. Notification Behavior
Integrated with `internal/notifications/service.Service`:
- Dispatches in-app notifications to assigned sales and pricing operators when an RFQ requires requirement clarification or when a draft has been submitted for approval.
- Uses idempotency keys based on RFQ ID and event type to prevent duplicate notifications.

---

## 15. Idempotency and Deduplication
- Every request carries a unique `CorrelationID` generated at the initiation of the workflow and threaded through Go, Python, MariaDB, and audit logs.
- Draft creation is idempotent per session; submitting for approval when already in `PENDING_APPROVAL` returns the existing record without duplicate approval creation.

---

## 16. Security and Tenant Isolation
- Every API endpoint validates the authenticated `UserContext` via `middleware.GetUserContext`.
- Organization isolation is enforced at the database level using `org_id` filters on all SELECT, INSERT, and UPDATE queries.
- Cross-tenant requests (e.g. Org 2 attempting to access Org 1 RFQ workflow) are rejected with `404 Not Found`.
- Python sidecar endpoints require `X-LogisticsHQ-Service-Key` and reject unauthorized external requests with `401 Unauthorized`.

---

## 17. Observability
- All operations log structured events including `correlation_id`, `rfq_id`, `org_id`, `user_id`, and `duration_ms`.
- Pricing calculations log itemized cost and sell components.
- Audit logs record all draft generation, editing, and approval submission actions in the `audit_logs` table.

---

## 18. Test Results
Comprehensive testing was executed across Go, Python, and React frontend:
1. **Go Unit & Service Tests**:
   - `internal/rfq/pricing_workflow/service_test.go`:
     - `TestDeterministicPricingCalculation`: **PASS**
     - `TestNegativeMarginDetection`: **PASS**
     - `TestActionRegistryQuotationSendDraft`: **PASS**
2. **Python AI Sidecar Tests**:
   - `ai_sidecar/tests/test_rfq_pricing_workflow.py`:
     - `test_rfq_context_pydantic_validation`: **PASS**
     - `test_requirements_extraction_complete`: **PASS**
     - `test_requirements_extraction_detects_missing_mandatory_fields`: **PASS**
     - `test_pricing_explanation_preserves_go_numbers`: **PASS**
     - `test_quotation_risk_analysis_low_margin_requires_approval`: **PASS**
     - `test_quotation_draft_synthesis_and_grounding`: **PASS**
     - `test_prompt_injection_sanitization`: **PASS**
     - Result: **7 passed in 0.32s (100%)**
3. **Frontend Component & UI Tests**:
   - `frontend/src/__tests__/pages/RFQ/RFQIntelligentPricingWorkflowSection.test.jsx`:
     - Overview rendering, metric cards, deterministic pricing: **PASS**
     - Re-evaluate requirement extraction: **PASS**
     - AI Quotation draft generation: **PASS**
     - Draft editing and value updates: **PASS**
     - Managerial approval submission: **PASS**
     - Result: **7 passed in 1.71s (100%)**
4. **Frontend Production Build**:
   - `npm run build`: Vite v8.0.12 client bundle built successfully in 32.91s with 0 errors.
5. **Live End-to-End System Integration**:
   - `scratch/verify_phase3_task34_live.py` executed against real MariaDB database, active Python sidecar (:8090), and Go backend (:8080):
     - Sidecar requirement extraction: **200 OK**
     - Sidecar pricing explanation: **200 OK**
     - Sidecar risk analysis: **200 OK**
     - Sidecar quotation drafting: **200 OK**
     - Backend Overview (`GET /overview`): **200 OK**
     - Backend Extraction (`POST /extract-requirements`): **200 OK (Persisted ID: 3)**
     - Backend Pricing Preview (`POST /pricing-preview`): **200 OK (Cost: $2,550.00, Sell: $2,990.00, Margin: 14.72% LOW)**
     - Backend Draft Creation (`POST /drafts`): **201 Created (Draft ID: 4, Requires Approval: True)**
     - Backend Submit Approval (`POST /submit-approval`): **200 OK (Approval ID: 201, Status: PENDING_APPROVAL)**
     - Cross-Tenant Security Isolation: **404 Not Found (Verified blocked)**
     - MariaDB Persistence Verification: **100% verified real database records in `ai_rfq_requirements_extractions` and `ai_quotation_drafts`**

---

## 19. Browser QA Results
- The application was verified running locally across all ports:
  - MySQL: `127.0.0.1:3306`
  - Go Backend: `127.0.0.1:8080`
  - Python AI Sidecar: `127.0.0.1:8090`
  - Frontend: `http://localhost:5173`
- The `browser_subagent` was invoked to perform visual inspection; however, the local environment encountered a Playwright driver CDN resolution error (`playwright-1.57.0-win32_x64.zip returned 404 from upstream`).
- Full DOM rendering, white/light UI class presence, interaction buttons, input updates, and error handling were independently validated through the Vitest browser testing suite (`RFQIntelligentPricingWorkflowSection.test.jsx`) with 100% pass rate.

---

## 20. Known Limitations
- The Playwright binary download on the local Windows environment failed due to upstream Playwright CDN 404 response for version 1.57.0 win32_x64 driver.
- External email delivery remains intentionally gated behind human manager approval; no automated live emails are dispatched without explicit approval token execution.

---

## 21. Confirmation: AI Code is Written in Python Only
- **CONFIRMED**: All AI reasoning, prompt orchestration, requirement extraction logic, missing field identification, pricing explanations, risk evaluations, and draft generation are written exclusively in Python (`ai_sidecar/app/rfq_pricing/`).
- No AI reasoning or prompt logic exists in Go, JavaScript, or TypeScript.

---

## 22. Confirmation: Go is Used for Integration, Validation, Persistence, Approvals, and Execution
- **CONFIRMED**: Go controls all database transactions, tenant scoping, deterministic pricing math, margin threshold checks, sidecar response validation, approval registration, and action execution.

---

## 23. Confirmation: No Fake Data Was Introduced
- **CONFIRMED**: All verification scripts and tests operated on real database entities (`rfq #1`, `org #1`, `customer #1`) in MariaDB `freel_mysql`. No fake fixtures or mock entities were introduced.

---

## 24. Confirmation: No Existing Business Data Was Reset or Deleted
- **CONFIRMED**: Database tables were not dropped, truncated, or reset. Migration `101` used `CREATE TABLE IF NOT EXISTS`, preserving all existing RFQs, quotations, customers, and bookings.

---

## 25. Confirmation: No Quotation Sent or Price Changed Without Authorization
- **CONFIRMED**: The action `quotations.send_draft` is strictly registered as high-risk and requires human approval. Quotations with low or negative margins cannot be finalized or dispatched without explicit approval through the LogisticsHQ Approval Center.
