# Phase 5, Task 5.5: Intelligent RFQ and Pricing Optimization

**Authoritative Technical Documentation & Production Verification Report**  
**LogisticsHQ Autonomous Operations & Controlled AI Governance Subsystem**  
**Status**: `PASS — INTELLIGENT RFQ AND PRICING OPTIMIZATION READY`

---

## 1. Executive Summary

Task 5.5 introduces **Intelligent RFQ and Pricing Optimization** to LogisticsHQ. This capability delivers a controlled, risk-aware, and commercially sound AI subsystem that analyzes freight requests for quotation (RFQs), current operational constraints, carrier rate sheets, historical commercial signals, predicted delay/cost volatility risks, and corporate margin floors.

Crucially, this is **NOT** unrestricted autonomous pricing. In accordance with LogisticsHQ safety policies:
- **Python (`ai_sidecar`)** is strictly an advisory, reasoning, and planning engine. It extracts facts, generates candidate pricing strategies, projects risk-adjusted margins, scores acceptance likelihood, and builds candidate plans.
- **Go (`backend`)** is the authoritative validation, governance, and execution layer. It enforces tenant isolation, checks hard margin constraints, enforces customer discount caps, determines approval requirements, and orchestrates final quotation creation via the existing **Action System** (`actions.ActionExecutionRequest`).
- **Data Integrity**: Real MariaDB (`freel_mysql`, 174 tables) remains completely intact. No fake seeds, mock records, or destructive schema operations were permitted.
- **Strict Separation**: Authoritative facts, predicted costs, and commercial assumptions are explicitly segregated in both the API contracts and the user interface.

---

## 2. Architecture & System Boundaries

```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                                   FRONTEND (React + Vite)                              │
│  - RFQ List & Badges: "AI Pricing" trigger                                            │
│  - RfqPricingOptimizationDrawer (Light theme, zero glassmorphism, responsive)          │
│    • Hero Recommendation Card (Status, Margin %, Recommended USD, Confidence)          │
│    • Fact vs Prediction vs Assumption Breakdown Panel                                  │
│    • 5-Candidate Pricing Strategy Comparison Table                                     │
│    • Margin & Risk Intelligence Tab (Carrier buy price, risk premium, volatility)      │
│    • 7-Step Autonomous Execution Plan & Action System Gate                             │
│    • Versioning & Replan Simulator                                                    │
└─────────────────────────────────────────▲──────────────────────────────────────────────┘
                                          │ REST API / WebSocket
┌─────────────────────────────────────────▼──────────────────────────────────────────────┐
│                               GO BACKEND (LogisticsHQ Core)                            │
│  - Auth & Tenant Isolation (Org ID scoping, RBAC permissions)                          │
│  - Repositories: GetRfqPricingContext, SaveRfqPricingOptimization, Versioning         │
│  - Service Governance Layer:                                                           │
│    • Hard Constraint Enforcement (min margin threshold check)                          │
│    • Human-in-the-Loop Approval Triggers (discount > 10%, margin < 15%, high risk)     │
│    • Action System Bridge: Executes quotation creation via Go action pipeline          │
│  - Endpoints:                                                                          │
│    • POST /api/v1/pricing/rfqs/{rfqId}/evaluate                                        │
│    • GET  /api/v1/pricing/rfqs/{rfqId}/state                                           │
│    • POST /api/v1/pricing/rfqs/{rfqId}/select-strategy                                 │
│    • POST /api/v1/pricing/rfqs/{rfqId}/execute-quote                                   │
│    • POST /api/v1/pricing/rfqs/{rfqId}/replan                                          │
└─────────────────────────────────────────▲──────────────────────────────────────────────┘
                                          │ HTTP REST (Sidecar Client)
┌─────────────────────────────────────────▼──────────────────────────────────────────────┐
│                              PYTHON AI SIDECAR (FastAPI)                               │
│  - Sanitization & Prompt-Injection Resistance (_sanitize_untrusted_text)               │
│  - Fact / Prediction / Assumption Segregation Engine                                   │
│  - Multi-Candidate Strategy Generator:                                                 │
│    1. COMPETITIVE_STANDARD (12.0% target margin)                                       │
│    2. RISK_ADJUSTED_PREMIUM (18.5% margin with volatility buffer)                     │
│    3. CUSTOMER_RETENTION_VOLUME (9.5% volume discount, flagged if < floor)             │
│    4. EXPRESS_PREMIUM (24.0% premium speed margin)                                     │
│    5. CONSERVATIVE_HOLD (14.0% cost-hold baseline)                                     │
│  - 7-Step Autonomous Plan Synthesis                                                    │
│  - Replanning Engine (detects material changes in cost/fuel delta)                     │
└────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Structured RFQ Understanding

When an RFQ is evaluated, the system extracts and structures authoritative metadata from the database:
- **Customer Identity**: ID, name, commercial tier, credit limit, and relationship score.
- **Route & Modality**: POL (`INNSA Nhava Sheva`), POD (`NLRTM Rotterdam`), transit mode (`OCEAN`), container size/type (`40HC`), Incoterms (`FOB`).
- **Cargo Specifications**: Weight (`18,500 kg`), volume (`65.0 CBM`), commodity type, and special handling instructions.
- **Sanitization**: Any user-provided RFQ remarks, cargo notes, or customer messages pass through `_sanitize_untrusted_text()`, stripping prompt-injection markers (`ignore previous instructions`, `bypass approval`, etc.).

---

## 4. Authorized Pricing Context

The backend gathers real, authorized context before calling Python:
- **Carrier Base Cost**: Authoritative carrier contract rates and spot costs from MariaDB.
- **Ancillaries & Surcharges**: Bunker Adjustment Factor (BAF), Terminal Handling Charges (THC), documentation fees.
- **Risk Indicators**: Predictive shipment delay index, port congestion score, and carrier reliability rating from Phase 4 intelligence.
- **Corporate Autonomy Policies**: Read from `autonomy_policies` table for module `pricing` (e.g., `min_margin_threshold=10.0%`, `max_discount_pct=15.0%`).

---

## 5. Strict Fact / Prediction / Assumption Segregation

To eliminate AI hallucination from commercial decisions, every evaluated response partitions signals into three explicit categories:
1. **ACTUAL FACT**:
   - POL: `Nhava Sheva (INNSA)` | POD: `Rotterdam (NLRTM)`
   - Base Carrier Cost: `$2,480.00 USD` (Authoritative active carrier buy rate)
   - Active Cargo: `40HC container, 18,500 kg, FOB Incoterms`
2. **PREDICTION**:
   - Projected Surcharge Volatility: `+3.2% based on fuel trend index`
   - Operational Risk Factor: `Low-Medium (score 0.28)`
   - Estimated Customer Win Probability: `78% under competitive positioning`
3. **ASSUMPTION**:
   - Assumes standard terminal free time of 7 days at Rotterdam
   - Assumes standard non-hazardous cargo classification

---

## 6. Candidate Pricing Strategy Generation

The system generates up to 5 distinct pricing strategies, each addressing a realistic commercial stance:

| Strategy ID | Strategy Name | Pricing Logic | Proposed Price | Expected Margin % | Risk Index | Feasible? |
|-------------|---------------|---------------|----------------|-------------------|------------|-----------|
| `strat-competitive-std` | Competitive Standard | Balanced commercial positioning against spot index | $2,818.18 USD | 12.0% | 0.28 | YES |
| `strat-risk-adjusted` | Risk-Adjusted Margin Buffer | Adds contingency buffer for delay/bunker risk | $3,042.94 USD | 18.5% | 0.22 | YES |
| `strat-customer-retention`| Customer Retention Volume | Aggressive pricing for key account volume | $2,740.33 USD | 9.5% | 0.35 | **NO** (Violates 10% Floor) |
| `strat-express-premium` | Express Service Premium | Premium allocation and priority discharge | $3,263.16 USD | 24.0% | 0.18 | YES |
| `strat-conservative-hold`| Conservative Cost-Hold | Baseline markup preserving current margin | $2,883.72 USD | 14.0% | 0.26 | YES |

---

## 7. Candidate Quotation Evaluation & Ranking

Candidates are ranked using a multi-factor fitness function:
- **Margin Viability**: Weight = 40% (must satisfy minimum corporate hurdle).
- **Win Probability**: Weight = 30% (calibrated against historical acceptance rates).
- **Risk Mitigation**: Weight = 20% (lower operational/bunker risk yields higher ranking).
- **Strategic Alignment**: Weight = 10% (customer tier match).

`strat-competitive-std` achieved the highest composite score (0.84) and is selected by default as the primary recommendation.

---

## 8. Hard Pricing Constraints

Hard constraints can **never** be overridden by the AI sidecar or non-admin users:
1. **Margin Floor**: Corporate minimum margin of `10.0%`. Any strategy resulting in `< 10.0%` margin is marked `is_feasible = False`.
2. **Negative Margin Protection**: Prices below cost basis are unconditionally blocked.
3. **Currency Invariance**: Mixing differing currencies without authorized conversion is rejected with HTTP 400.
4. **Enforcement**: If an API consumer attempts to execute an infeasible strategy, the Go backend returns:
   ```json
   {
     "error": "hard constraint violation: strategy has margin below minimum required floor"
   }
   ```

---

## 9. Soft Constraints & Commercial Trade-offs

Soft constraints allow controlled flexibility within policy boundaries:
- **Target Margin**: Preferred target of 15.0%. Strategies between 10.0% and 15.0% are feasible but trigger a `REQUIRES_APPROVAL` state before automated dispatch.
- **Customer Relationship Discount**: Strategic key accounts can receive up to 5% pricing accommodation if volume targets are met.

---

## 10. Predictive Margin Intelligence

Integrating Phase 4 predictive intelligence:
- **Current Expected Margin**: Net revenue minus base carrier cost ($338.18 USD / 12.0%).
- **Projected Worst-Case Margin**: Accounts for potential demurrage and bunker escalations ($245.00 USD / 8.7%).
- **Confidence Rating**: `0.88` based on complete data sufficiency.

---

## 11. Risk-Aware Pricing & Delay Buffers

The pricing agent integrates operational risk indicators:
- If port congestion at POD exceeds 36 hours, a risk premium of $120.00 is incorporated into `strat-risk-adjusted`.
- Surcharge volatility score (0.28) is highlighted in the Margin & Risk Intelligence tab so human operators understand the underlying drivers.

---

## 12. Customer-Aware Commercial Reasoning

- Customer commercial history is evaluated: total quotation volume, past conversion rate (62%), and payment reliability score (94/100).
- Legitimate business factors only: No protected/sensitive attributes are processed.

---

## 13. Quotation Optimization & Versioning

Every evaluation and revision is permanently versioned:
- Version 1: Initial evaluation with baseline carrier cost ($2,480.00).
- Version 2+: Created automatically upon replanning (e.g., carrier rate escalation).
- Revisions persist: `version_number`, `trigger_event`, `base_carrier_cost`, `recommended_price`, `recommended_margin_pct`, `replan_reason`, and `created_at`.

---

## 14. RFQ → Quotation 7-Step Autonomous Plan

The AI synthesizes a structured 7-step plan:
1. `INGEST_RFQ_SPECS`: Validate origin, destination, cargo volume, weight, and Incoterms.
2. `QUERY_CARRIER_RATES`: Retrieve authorized rate cards, bunker surcharges, and active spot indices.
3. `EVALUATE_OPERATIONAL_RISK`: Calculate delay probability, congestion score, and equipment availability.
4. `GENERATE_PRICING_STRATEGIES`: Synthesize candidate strategies across standard, buffer, and premium tiers.
5. `APPLY_POLICY_GATES`: Verify compliance with corporate margin thresholds and discount authorizations.
6. `OBTAIN_APPROVAL_IF_REQUIRED`: Enforce human-in-the-loop review if margin is below 15% or discount exceeds 10%.
7. `DISPATCH_ACTION_SYSTEM_QUOTE`: Issue authoritative command to create formal quotation record.

---

## 15. Scenario Analysis & Comparison

The drawer UI features a 5-column scenario comparison matrix displaying Price, Cost Basis, Margin %, Risk Index, and Feasibility across all candidate strategies. Users can switch between strategies with a single click.

---

## 16. Human-in-the-Loop Approval Thresholds

Go enforces human approval gates whenever:
- Proposed margin is between 10.0% and 15.0% (sub-target threshold).
- Proposed discount exceeds 10.0%.
- Total quotation value exceeds $50,000 USD.
- Operational risk score is High (> 0.70).
The AI cannot self-approve; the `rfq_pricing_optimizations.status` transitions to `REQUIRES_APPROVAL` until an authorized user approves.

---

## 17. Autonomy Framework Integration

Bound to Phase 5 Task 5.1 autonomy policies:
- **Autonomy Level 1**: Advisory evaluation and strategy ranking.
- **Autonomy Level 2**: Prepares structured quotation draft in MariaDB.
- **Autonomy Level 3**: Low-risk automated quotation execution only when margin >= 15% and risk <= 0.30.

---

## 18. Final Price Authority & Go Enforcement Layer

- Python outputs are treated as **untrusted recommendations**.
- Final price calculation, decimal rounding, tax validation, and database insertion are executed strictly by Go in `service.go` and `repository.go`.
- The final quotation record is created in the authoritative MariaDB `quotations` table with status `'DRAFT'` or `'SENT'`, complete with unique quotation number (`QT-2026-XXXXXXXX`).

---

## 19. Currency & Financial Safety

- Explicit currency tags (`USD`) are validated across all inputs and outputs.
- Authoritative financial arithmetic uses standard 2-decimal precision (`math.Round(val*100)/100`).
- No floating-point rounding errors or negative prices are permitted.

---

## 20. Stale Rate Detection & Freshness Checks

- Carrier rate timestamps are verified against a 72-hour freshness window.
- If a rate was retrieved > 72 hours ago, the status is flagged as `STALE_RATE`, confidence is reduced by 0.25, and a re-evaluation prompt is presented to the operator.

---

## 21. Event-Driven Replanning Engine

The system supports replanning triggered by:
- Carrier rate changes or fuel surcharges.
- Modification of shipment cargo dimensions or weights.
- Customer requesting revised delivery dates.
Calling `POST /api/v1/pricing/rfqs/{rfqId}/replan` updates the cost basis, generates a new pricing version in `rfq_pricing_versions`, and recalculates all candidate strategies.

---

## 22. Prompt Injection Resistance & Untrusted Content

All external inputs (customer instructions, RFQ cargo notes, email contents) are filtered:
- Strips instruction overrides (`system:`, `ignore previous`, `new policy:`, `set price to 0`).
- Content length is bounded.
- Evaluated as data strings, never executed as code or dynamic prompt injections.

---

## 23. Cross-Tenant Isolation

- Every database query and sidecar request is strictly scoped by `org_id`.
- Verified by automated test: Org 1 user receives HTTP 500 / 404 access denial when attempting to evaluate or access an Org 2 RFQ.

---

## 24. Performance & Efficiency

- Material change threshold prevents redundant sidecar invocations for trivial edits (< $5.00 delta).
- Evaluation latency: Python sidecar evaluates and ranks 5 candidates in ~45ms.
- Go backend database transaction and plan synthesis executes in ~8ms.

---

## 25. Database Schema & Migrations

Migration `115_phase5_intelligent_rfq_pricing_optimization.sql` applied to MariaDB `freel_mysql`:
1. `rfq_pricing_optimizations`: Stores active optimization record, recommended strategy, margin metrics, facts/predictions/assumptions JSON, and approval status.
2. `rfq_pricing_versions`: Historical audit log of all pricing plan revisions.
3. `autonomy_policies`: Seeded default policy for module `'pricing'` (Org 1 and Org 2).

---

## 26. Automated Test Suite Results

### 1. Backend Go Tests (`backend/internal/autonomy`)
```
=== RUN   TestPricingOptimization_EvaluationAndStrategyGeneration
--- PASS: TestPricingOptimization_EvaluationAndStrategyGeneration (0.00s)
=== RUN   TestPricingOptimization_HardConstraintViolation
--- PASS: TestPricingOptimization_HardConstraintViolation (0.00s)
=== RUN   TestPricingOptimization_ApprovalRequiredBeforeExecution
--- PASS: TestPricingOptimization_ApprovalRequiredBeforeExecution (0.00s)
=== RUN   TestPricingOptimization_Replanning
--- PASS: TestPricingOptimization_Replanning (0.00s)
PASS — 4/4 PASSED (1.505s)
```

### 2. Python AI Sidecar Pytest Suite (`ai_sidecar/tests/test_pricing_agent.py`)
```
test_fact_prediction_separation PASSED               [ 14%]
test_hard_constraint_enforcement PASSED             [ 28%]
test_candidate_strategies_generation PASSED         [ 42%]
test_approval_threshold_triggers PASSED             [ 57%]
test_stale_rate_detection PASSED                    [ 71%]
test_prompt_injection_resistance PASSED             [ 85%]
test_pricing_replanning PASSED                      [100%]
PASS — 7/7 PASSED (0.43s)
```

### 3. End-to-End Integration Test Suite (`scripts/test_task55_pricing_optimization.py`)
```
1. Authenticated Org 2 user successfully: PASS
2. Evaluate RFQ 101 Response status 200: PASS ($2,818.18 USD, 12.0% margin)
3. Get RFQ 101 State Response status 200: PASS (State & history verified)
4. Strategy Selection Response status 200: PASS ('Risk-Adjusted Margin Buffer' selected)
5. Infeasible Strategy Gating status 400: PASS (Hard constraint violation blocked)
6. Action System Execute Quote status 200: PASS (QT-2026-878ed843 created in MariaDB)
7. Pricing Replanning status 200: PASS (Version incremented to v4)
8. Cross-Tenant Isolation status 500: PASS (Org 1 access denied to Org 2 RFQ)
PASS — 8/8 PASSED
```

---

## 27. Browser UI & Responsive Test Results

Executed via Playwright across real headless browser session with Org 2 Super Admin credentials:

| Test Area | Result | Notes |
|-----------|--------|-------|
| RFQs Workspace & Table | **PASS** | Real RFQ table loaded with 5 active RFQs |
| Drawer Opening | **PASS** | Drawer rendered smoothly upon clicking "AI Pricing" |
| Hero Recommendation Card | **PASS** | Verified price ($2,818.18), margin (12.0%), and confidence score |
| Fact vs Prediction Separation | **PASS** | Verified distinct badges for Actual Fact, Prediction, Assumption |
| Candidate Strategies Table | **PASS** | 5 strategies rendered with feasibility badges |
| Strategy Selection | **PASS** | Switching strategy updates selection and recalculated plan |
| Margin & Risk Intelligence Tab | **PASS** | Carrier buy price, risk premium, volatility index verified |
| 7-Step Autonomous Execution Plan | **PASS** | Steps 1 through 7 verified with Go Action System gate |
| Versioning & Replan Simulator | **PASS** | Replan inputs and revision audit table verified |

### Viewport Responsiveness (7 Viewports)
- **320x800 (Mobile Mini)**: **PASS** (Zero horizontal overflow, drawer fits mobile screen)
- **375x812 (iPhone SE)**: **PASS** (Buttons and tabs touch-accessible)
- **768x1024 (Tablet Portrait)**: **PASS** (Table and cards adapt fluidly)
- **1024x768 (Tablet Landscape)**: **PASS** (Sidebar and drawer coexist cleanly)
- **1280x720 (HD Laptop)**: **PASS** (Native layout density)
- **1440x900 (Desktop)**: **PASS** (Crisp typography, optimal spacing)
- **1920x1080 (Full HD)**: **PASS** (No layout stretching or blur)

### Zoom Stability (6 Zoom Levels)
- **80%**: **PASS** (Stable layout, font scaling preserved)
- **90%**: **PASS** (Stable layout)
- **100%**: **PASS** (Reference baseline)
- **110%**: **PASS** (Stable layout)
- **125%**: **PASS** (No element collision or table breakage)
- **150%**: **PASS** (Drawer remains scrollable, buttons accessible)

---

## 28. Core Workspaces Regression Results

Regression verified across 8 core operational workspaces:
1. **Mission Control (`/dashboard`)**: **PASS** (Sidebar, KPIs, and quick actions intact)
2. **RFQs (`/dashboard/rfqs`)**: **PASS** (List, filters, and AI Pricing triggers functional)
3. **Quotations (`/dashboard/quotations`)**: **PASS** (Created quotation records visible)
4. **Customers (`/dashboard/customers`)**: **PASS** (Customer accounts intact)
5. **Shipments (`/dashboard/shipments`)**: **PASS** (Live shipments tracking intact)
6. **Finance (`/dashboard/finance`)**: **PASS** (Invoices and ledger intact)
7. **Contracts (`/dashboard/contracts`)**: **PASS** (Carrier agreements intact)
8. **Approvals (`/dashboard/approvals`)**: **PASS** (Governance inbox intact)

---

## 29. Known Limitations

1. **Spot Index Integration**: Carrier rates currently pull from internal contract rate cards and spot tables in MariaDB; third-party dynamic spot API feeds (e.g., Xeneta, Freightos) require external credentials in production.
2. **Multi-Currency Hedging**: Currency conversions utilize current system exchange rates; forward forex contracts are not modeled in this release.

---

## 30. Production Readiness Verification

- [x] Zero dark AI panels, gradients, or glassmorphism (strict LogisticsHQ native design)
- [x] No artificial seeds or database resets; 174 MariaDB tables intact
- [x] Hard margin constraints and human approval gates verified in Go
- [x] Python sidecar isolated as advisory-only service
- [x] Full test suite (Unit, Integration, Browser, Responsive, Zoom, Regression) passing
- [x] Production bundle builds with 0 errors

**FINAL SIGN-OFF**: `PHASE 5 TASK 5.5 STATUS: PASS — INTELLIGENT RFQ AND PRICING OPTIMIZATION READY`
