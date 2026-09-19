# Phase 5, Task 5.6: Adaptive Finance and Collections

**Status:** PASS — ADAPTIVE FINANCE AND COLLECTIONS READY  
**Author:** Antigravity AI  
**Date:** September 11, 2026  
**Repository:** `freel-project`  

---

## 1. Executive Summary
LogisticsHQ Phase 5, Task 5.6 implements **Adaptive Finance and Collections**, a controlled, enterprise-grade AI capability that monitors freight receivables, models customer payment risk, prioritizes collection opportunities, generates adaptive multi-strategy collection plans, drafts context-aware customer communications, and safely coordinates approved finance actions without compromising authoritative accounting boundaries.

In accordance with core enterprise governance rules:
- **Python AI Sidecar** exclusively handles financial reasoning, collections prioritization, risk driver synthesis, candidate strategy generation, multi-invoice consolidation analysis, prompt-injection defense, 7-step autonomous plan formulation, and adaptive event replanning.
- **Go Backend & Action System** remain strictly authoritative for the general ledger, balances, currencies, tax, approvals (HITL gates), permission checks, idempotency, audit trails, and execution boundaries.
- **Non-Negotiable Boundaries**: The AI cannot alter invoice balances, create credits, write off debt, issue refunds, waive penalties, or send customer messages directly.

---

## 2. Architecture
```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                                 FRONTEND (React + Vite)                                 │
│  - Invoices Workspace (`/dashboard/invoices`)                                          │
│  - Collections Intelligence Banner                                                      │
│  - Adaptive Collections Drawer (`FinanceCollectionsAdaptiveDrawer.jsx`)                │
│    * [Actual] Ledger vs [Predicted] Risk vs [Recommended] Strategy vs Autonomy Strip   │
│    * 4 Subtabs: Strategies, Facts vs Predictions, 7-Step Autonomous Plan, Replanning   │
└────────────────────────────────────────┬───────────────────────────────────────────────┘
                                         │ REST API (Bearer JWT / Org Context)
                                         ▼
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                              GO BACKEND (Core Engine)                                  │
│  - `internal/autonomy/handler.go` (`/api/v1/autonomy/finance/invoices/{id}/*`)         │
│  - `internal/autonomy/service.go` (Policy enforcement, approval gates, replanning)     │
│  - `internal/autonomy/repository.go` (Authoritative invoice context, plan persistence)  │
│  - `internal/actions/service.go` (Idempotent Action System execution boundary)         │
│  - MariaDB (`finance_collection_plans`, `finance_collection_versions`)                │
└────────────────────────────────────────┬───────────────────────────────────────────────┘
                                         │ Mutual Service Key Auth (`X-LogisticsHQ-Service-Key`)
                                         ▼
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                          PYTHON AI SIDECAR (FastAPI + LangGraph)                        │
│  - `app/autonomy/finance_agent.py`                                                     │
│    * `POST /autonomy/finance/evaluate-collection`                                       │
│    * `POST /autonomy/finance/replan-collection`                                        │
│  - 5 Candidate Strategies (`strat-wait-monitor`, `strat-friendly-reminder`, etc.)      │
│  - Fact / Prediction / Assumption Segregation                                         │
│  - Multi-Invoice Account Aggregation                                                   │
│  - Adaptive Event Replanning (Payments, Partial Payments, Disputes, Customer Replies)  │
└────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Financial Authority
Authoritative financial numbers remain strictly controlled in Go and MariaDB:
- **Total Amount, Paid Amount, Balance Due**: Derived exclusively from verified transactional invoice records.
- **Currency**: Preserved explicitly (e.g. `USD`, `EUR`). No ad-hoc currency cross-mixing.
- **Status Authority**: Transitions (e.g., `Draft` → `Issued` → `Paid` / `Overdue` / `Disputed`) can only be executed by Go business logic.
- **AI Separation**: All AI figures are explicitly flagged as predictions (`[Predicted]`) or recommendations (`[Recommended]`) and never mutate the ledger.

---

## 4. Finance Context
The Go backend assembles a comprehensive, verified `FinanceInvoiceContext` passed to the Python sidecar:
- **Invoice Facts**: `invoice_id`, `invoice_number`, `customer_id`, `customer_name`, `total_amount`, `paid_amount`, `balance_due`, `currency`, `issue_date`, `due_date`, `days_overdue`, `status`, `dispute_status`.
- **Aging Analysis**: Categorized into standardized aging buckets (`CURRENT`, `1-15_DAYS`, `16-30_DAYS`, `31-60_DAYS`, `61-90_DAYS`, `90+_DAYS`).
- **Account Health & Other Invoices**: Aggregates all open, overdue, and disputed invoices for the customer to prevent spamming multiple disconnected messages.
- **Payment Behavior Profile**: Historical late payments, reminder responsiveness, average payment lag.
- **Commercial Relationship**: Account tier (`STRATEGIC`, `ENTERPRISE`, `STANDARD`), freight volume, open shipments.

---

## 5. Invoice States
The system recognizes and safely respects all LogisticsHQ authoritative invoice states:
- `DRAFT`: AI collections disabled.
- `ISSUED` / `SENT`: Monitored; friendly reminders scheduled as due date approaches.
- `PARTIALLY_PAID`: Re-evaluated based on remaining balance due.
- `PAID`: Active collection plans immediately transition to `STOPPED` / `RESOLVED`.
- `OVERDUE`: Candidate strategies generated based on aging severity.
- `DISPUTED`: Automatic freeze of automated outbound reminders; routes to dispute resolution.
- `ON_HOLD` / `CANCELLED` / `CREDITED` / `REFUNDED`: Collections stopped immediately.

---

## 6. Collection Prioritization
Prioritization is calculated using a multi-factor risk and urgency model:
- **Financial Exposure**: High balance due significantly elevates score.
- **Aging Severity**: Overdue duration scaling exponentially past 30 days.
- **Customer Relationship Tier**: Strategic accounts receive tailored, collaborative follow-ups rather than rigid automation.
- **Dispute Presence**: Moves priority to collaborative resolution rather than aggressive collections.
- **Cash Flow Impact**: Impact rating (`LOW`, `MEDIUM`, `HIGH`, `CRITICAL`).

---

## 7. Risk Intelligence
Rather than a single opaque number, collection risk is decomposed into observable, auditable drivers:
- `OVERDUE_DURATION`: Impact of days past agreed payment terms.
- `LARGE_EXPOSURE`: Threshold exceeding $5,000 requiring senior controller oversight.
- `ACTIVE_DISPUTE`: Freight damage or rate discrepancy pauses automated follow-ups.
- `PAYMENT_BEHAVIOR`: Past delays or bounce history.
- `COMMUNICATION_FATIGUE`: Excessive reminders without customer acknowledgment.

---

## 8. Collection Strategies
The Python AI Sidecar generates 5 structured candidate strategies for each invoice:
1. `strat-wait-monitor`: **Wait and Monitor Receivables** (For invoices not yet due or within grace period).
2. `strat-friendly-reminder`: **Friendly Courtesy Reminder** (For invoices approaching due date or 1–7 days past due).
3. `strat-status-confirmation`: **Formal Payment Status Request** (For invoices 8–30 days overdue).
4. `strat-dispute-resolution`: **Dispute Reconciliation Review** (When an active billing or operational dispute is detected).
5. `strat-finance-escalation`: **Senior Finance Escalation** (For severely overdue or high-balance accounts).

Each strategy specifies recommended actions, tone, human approval requirements, and rationale.

---

## 9. Due-Date Awareness
The system strictly distinguishes current, upcoming, and overdue invoices:
- Invoices due in > 3 days: Default strategy is `strat-wait-monitor`.
- Invoices approaching due date (≤ 3 days): Courtesy heads-up drafted without pressure.
- Invoices overdue 1–15 days: Professional status check.
- Invoices severely overdue (> 30 days): Internal finance escalation recommended.

---

## 10. Dispute Awareness
When an invoice is marked `DISPUTED` or contains an active freight claim:
- Automated reminder delivery is halted.
- The AI selects `strat-dispute-resolution`.
- Tone switches to collaborative fact-finding.
- Mandatory Human-In-The-Loop approval is enforced before any customer communication.

---

## 11. Customer Payment Behavior
Analyzes historical payment performance without making unsupported deterministic claims:
- Expresses uncertainty (confidence scores between 0.0 and 1.0).
- Identifies typical settlement windows (e.g. "Customer typically processes payments on the 25th of the month").
- Recommends timing reminders to align with customer accounts payable batch runs.

---

## 12. Collection Plan
Every evaluated invoice produces an immutable, structured 7-step plan persisted in MariaDB:
1. `PLAN_INIT`: Verification of authoritative ledger balance and payment terms.
2. `RISK_EVAL`: Multi-driver payment risk and cash flow impact modeling.
3. `STRATEGY_SELECT`: Candidate strategy synthesis and rule-based selection.
4. `POLICY_GATE`: Autonomy level and monetary threshold validation.
5. `APPROVAL_CHECK`: Human review determination (flags > $5,000 or disputes).
6. `ACTION_DISPATCH`: Dispatching action to Go Action System boundary.
7. `REPLAN_OBSERVE`: Monitoring for payment events or customer replies.

---

## 13. Multi-Invoice Customer Planning
When an account has multiple open invoices:
- Python aggregates all related invoices into `multi_invoice_summary`.
- Consolidates context to avoid sending separate disconnected reminders.
- Acknowledges paid and disputed invoices while addressing the overdue balances in a unified statement.

---

## 14. Customer Communication
- **Drafting Only in Python**: The AI drafts suggested messages based on verified invoice facts.
- **Go Execution Boundary**: All message dispatches must be executed via `ExecuteFinanceCollectionAction` through the Go Action System.
- **Audit Logging**: Every dispatched message records correlation IDs, user IDs, and timestamps.

---

## 15. Financial Communication Safety
Customer-facing drafts strictly prohibit fabricated figures:
- Late fees, interest penalties, or discounts are never introduced by the AI.
- Balance due, invoice number, and due date are injected directly from Go's authoritative context.

---

## 16. Collection Message Types
Supported structured communication categories:
- `UPCOMING_DUE_REMINDER`
- `OVERDUE_COURTESY_REMINDER`
- `PAYMENT_STATUS_REQUEST`
- `DISPUTE_CLARIFICATION`
- `INTERNAL_FINANCE_ESCALATION`
- `STATEMENT_CONSOLIDATION`

---

## 17. Discount / Settlement Safety
- The AI may suggest that a settlement discount could accelerate recovery.
- It is strictly forbidden from applying or committing to any discount.
- Any commercial concession requires Go authorization and authorized finance manager sign-off.

---

## 18. Write-Off Safety
- Severe debt may be flagged as a candidate for bad-debt review.
- The AI has zero write-off authority. Write-offs must be processed manually in the General Ledger by finance controllers.

---

## 19. Refund Safety
- No automated refunds can be triggered by collections workflows.
- Overpayments route to human billing specialists for reconciliation.

---

## 20. Payment Application
- Payments are matched and recorded exclusively by Go payment reconciliation handlers.
- The AI collection agent detects payment events and adapts its plans accordingly.

---

## 21. Currency Safety
- Amounts are bound to their ISO currency code.
- Cross-currency sums or conversions are strictly blocked unless authoritative exchange rates are provided by Go.

---

## 22. Financial Numeric Safety
- Database schema uses `DECIMAL(12,2)` for financial precision.
- Zero floating-point rounding errors in ledger checks.
- Sanitized date handling to avoid RFC3339 timestamp mismatches on MariaDB `DATE` columns.

---

## 23. Approval Requirements (HITL)
Human-In-The-Loop approval is mandatory when:
- Invoice balance exceeds $5,000 (`MaxMonetaryThreshold`).
- Invoice is in `DISPUTED` status.
- Invoice is severely overdue (> 30 days).
- Autonomy level is below Level 3.

---

## 24. Autonomy Levels
Adheres to LogisticsHQ 5-level autonomy architecture:
- **Level 0 (Observe)**: Monitors receivables and aging without generating active plans.
- **Level 1 (Recommend)**: Suggests collection strategies and drafts; requires manual review.
- **Level 2 (Prepare)**: Synthesizes complete 7-step plans and drafts; awaits approval.
- **Level 3 (Bounded Execution)**: Autonomously executes low-risk reminders (≤ $5,000, non-disputed).
- **Level 4 (Full Workflow)**: Executes multi-step workflows with autonomous replanning under strict policy.

---

## 25. Stop Conditions
A collection workflow terminates immediately if:
- Invoice balance reaches $0.00 (`PAID`).
- Invoice is cancelled or credited.
- Customer opens a formal dispute.
- Customer explicitly requests contact pause / human escalation.
- Maximum reminder count (3) is reached.
- Emergency stop or kill switch is triggered.

---

## 26. Follow-Up Limits
- Maximum automated reminders: **3 follow-ups**.
- Minimum cooldown interval: **72 hours** between contacts.
- Escalation threshold: **14 days** without customer response triggers internal finance escalation.

---

## 27. Payment Event Adaptation
When payment events are received via `/api/v1/autonomy/finance/invoices/{id}/replan`:
- `PAYMENT_RECEIVED`: If balance is $0, plan transitions to `STOPPED` / `RESOLVED`.
- `PARTIAL_PAYMENT`: Recalculates balance due; generates updated version (e.g. `v2`) with revised collection target.
- `PAYMENT_FAILED`: Immediately notifies finance and drafts payment failure clarification.

---

## 28. Invoice Change Adaptation
If invoice due date, total amount, or line items change:
- Existing active collection plan is invalidated.
- A new evaluation is triggered with the updated authoritative facts.
- Audit history retains previous versions with change reasons.

---

## 29. Plan Versioning
Every plan change creates an immutable record in `finance_collection_versions`:
- Version number increments (`v1` → `v2` → `v3`).
- Stores `trigger_event`, `reason`, `plan_snapshot`, and timestamp.
- Allows complete retrospective auditing by finance teams.

---

## 30. Event-Driven Collections
The system listens to core business events:
- `INVOICE_ISSUED`
- `INVOICE_OVERDUE`
- `PAYMENT_CONFIRMED`
- `PARTIAL_PAYMENT_RECORDED`
- `DISPUTE_RAISED`
- `CUSTOMER_COMMUNICATION_RECEIVED`

Deduplication keys prevent recursive event loops.

---

## 31. Collection Escalation
Escalation creates an internal task for the credit control team containing:
- Full aging and payment history breakdown.
- Log of all sent reminders and customer responses.
- Recommended escalation route (e.g. telephone contact, credit hold).

---

## 32. Contract-Aware Collections
- Incorporates agreed payment terms (e.g. Net 30, Net 60) from customer contracts.
- Grace periods are respected before any overdue flag is applied.

---

## 33. Customer Relationship Awareness
- Strategic and key enterprise accounts are flagged.
- Aggressive language is barred; collaborative account executive touches are scheduled instead.

---

## 34. Collection Plan Verification
- Every dispatched action verifies the resulting audit record in Go.
- Status is only updated to `EXECUTED` when the Action System confirms successful dispatch.

---

## 35. Action System Integration
- The Action System is the sole mechanism for outbound side-effects.
- Every action payload includes `action_type`, `invoice_id`, `plan_id`, `idempotency_key`, and `correlation_id`.

---

## 36. Idempotency
- Uses UUID-based idempotency keys: `fin-col-{invoiceId}-{version}-{step}`.
- Prevents duplicate reminders even during network retries or worker restarts.

---

## 37. Memory
- Historical customer responsiveness is recorded.
- Customer preferences (e.g., preferred contact time, AP email) are stored in tenant-isolated memory.

---

## 38. UI Design
- Integrated into `/dashboard/invoices` and `/dashboard/finance/invoices`.
- Includes Quick Banner and action buttons (`btn-ai-collection-{id}`).
- Full-featured **Adaptive Collections Drawer** (`FinanceCollectionsAdaptiveDrawer.jsx`):
  - Clean LogisticsHQ light theme (slate, blue, emerald, amber, rose).
  - Explicit KPI strip segregating `[Actual]` Ledger, `[Predicted]` Risk, `[Recommended]` Strategy, and `[Governed]` Autonomy.
  - Subtabs: Strategies, Facts vs Predictions, 7-Step Plan, Adaptation/Replanning.

---

## 39. Security & Permissions
- Strict multi-tenant isolation (`org_id` verified on every query).
- RBAC validation (`FINANCE_WRITE`, `AUTONOMY_EXECUTE`).
- Prompt injection protection sanitizes free-form invoice notes and customer replies.

---

## 40. Failure & Recovery
- Python Sidecar fallback: Go falls back to standard rule-based aging if AI is unreachable.
- Action System timeouts prevent orphaned state.
- Database reconnection resilience verified.

---

## 41. Test Results Summary

| Test Suite | Command | Result |
| :--- | :--- | :--- |
| **Go Backend Unit Tests** | `go test -v ./internal/autonomy -run TestFinance` | **PASS (5/5 tests)** |
| **Python Sidecar Tests** | `pytest tests/test_finance_agent.py` | **PASS (8/8 tests)** |
| **Live Integration Suite** | `python scripts/test_task56_finance_collections.py` | **PASS (11/11 tests)** |
| **Browser & UI Suite** | `python scripts/test_task56_browser_ui.py` | **PASS (All Criteria)** |
| **Responsive Viewports** | 7 Viewports (320px to 1920px) | **PASS (0 overflow)** |
| **Zoom Levels** | 6 Zoom Levels (80% to 150%) | **PASS (0 overflow)** |
| **Core Workspaces** | 8 Core Workspaces Regressed | **PASS** |

---

## 42. Production Readiness
Task 5.6 meets all enterprise acceptance criteria. The adaptive finance and collections system is fully operational, safely governed, and ready for production deployment.
