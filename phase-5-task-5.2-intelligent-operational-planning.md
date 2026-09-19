# Phase 5, Task 5.2 — Intelligent Operational Planning Report

## 1. Executive Summary

Phase 5, Task 5.2 establishes the core **Intelligent Operational Planning** capability for LogisticsHQ, extending the Phase 5 autonomous operations foundation established in Task 5.1. 

Logistics operations frequently encounter disruptions—port congestion, customs documentation holds, vessel delays, and equipment shortages. Rather than relying on rigid static automation or unconstrained LLM execution, Task 5.2 equips LogisticsHQ with a **hybrid AI-Go operational planning engine**:
- **Python (AI Sidecar)** synthesizes operational strategies, reasons over operational goals and constraints, generates distinct candidate recovery plans, computes multi-attribute trade-off evaluations (cost, delay, customer impact, operational risk, reversibility), and ranks candidates using objective utility formulations.
- **Go (Backend Enforcement Boundary)** gathers authorized, tenant-isolated operational context, authoritatively enforces hard constraints, controls plan versioning and staleness, enforces approval gates, binds execution to supported Action System primitives, and audits every state change.
- **Frontend UI (LogisticsHQ Native)** seamlessly integrates operational goal formulation, side-by-side alternative plan comparison, step-by-step dependency inspection, and revalidation directly into Shipments and the AI Workforce Monitoring Dashboard without dark panels or visual dissonance.

The complete system was validated end-to-end against real persistent MariaDB data (`freel_mysql`), Python unit & integration suites, Go test suites, and Playwright browser testing across 7 viewports (320x800 to 1920x1080) and 6 zoom levels (80% to 150%).

---

## 2. Architecture

The architecture maintains a strict, non-negotiable separation of responsibilities:

```
+-----------------------------------------------------------------------------------+
|                            BUSINESS GOAL / EVENT                                  |
|   (User in UI or Business Event: Delay, Milestone Failure, Exception Detected)    |
+-----------------------------------------------------------------------------------+
                                         |
                                         v
+-----------------------------------------------------------------------------------+
|                        GO ENFORCEMENT BOUNDARY (Port 8080)                        |
|  - Authenticate User & Validate Tenant Isolation (OrgID scoping)                  |
|  - OperationalContextAssembler: Assemble live DB records, provenance, freshness   |
|  - Authoritative Policy Evaluation (Autonomy Level, Monetary Caps, Actions)       |
+-----------------------------------------------------------------------------------+
                                         |
                                         | HTTP (POST /autonomy/plan/generate)
                                         | Headers: X-LogisticsHQ-Service-Key, TenantID
                                         v
+-----------------------------------------------------------------------------------+
|                      PYTHON AI PLANNING ENGINE (Port 8090)                        |
|  - Input Sanitization & Prompt Injection Defense (safety.py)                      |
|  - Multi-Candidate Generation: Strategies A (Expedite), B (Reroute), C (Monitor)  |
|  - Constraint Engine: Hard vs Soft Constraint classification                      |
|  - Multi-Attribute Evaluation: Delay delta, Cost delta, Customer Impact, Risk     |
|  - Candidate Ranking & Selection of optimal feasible plan                         |
|  - Structured Step Model with Conditions & Verification Criteria                  |
+-----------------------------------------------------------------------------------+
                                         |
                                         | Return Candidate Plans + Ranked Steps
                                         v
+-----------------------------------------------------------------------------------+
|                        GO VALIDATION & PERSISTENCE                                |
|  - Validate Plan Schema & Reject Unknown Action Types                             |
|  - Hard-Constraint Gating: Infeasible candidates cannot be executed               |
|  - Persist Goal & Plan in MariaDB (planning_goals, autonomous_plans, steps)       |
|  - Enforce Approval Requirements (HITL gates before step dispatch)                |
+-----------------------------------------------------------------------------------+
                                         |
                                         v
+-----------------------------------------------------------------------------------+
|                       ACTION SYSTEM & STEP EXECUTION                              |
|  - Safe Conditional Branch Evaluation (Structured Predicates)                     |
|  - Step Dispatch via Action System (Idempotency, Retries, Timeouts)               |
|  - Result Verification Criteria (Field Equals, State Transition)                  |
|  - Audit Logging (autonomous_plan_audit_history)                                  |
+-----------------------------------------------------------------------------------+
```

### Prohibited Actions Confirmed
- Python **never** executes SQL, touches MariaDB directly, mutates business records, or calls external carrier APIs.
- Python **never** bypasses Go authorization or approval policies.
- Python **never** decides its own autonomy level.

---

## 3. Planning Workflow

The operational planning lifecycle follows an explicit state progression:
1. **Goal Formulation**: A business goal is declared via API or UI with priority, risk tolerance, and explicit hard/soft constraints.
2. **Context Gathering**: Go pulls fresh operational records from `shipments`, `invoices`, and `shipment_exceptions` along with Phase 4 predictive intelligence.
3. **Candidate Synthesis**: Python formulates 3 distinct candidate operational strategies:
   - **Candidate A (Accelerated Direct Recovery)**: Immediate milestone validation and expedited carrier dispatch. High delay recovery, moderate cost.
   - **Candidate B (Express Route & Multi-Carrier Handoff)**: Alternate port or carrier rerouting. Maximum delay recovery, higher cost, requires approval.
   - **Candidate C (Conservative Continuous Monitoring)**: Enhanced polling interval and automated notification trigger. Zero additional cost, lower delay recovery.
4. **Multi-Attribute Evaluation & Ranking**: Each candidate is evaluated against hard constraints (disqualifying violations) and scored across soft constraints and utility.
5. **Candidate Selection**: The optimal candidate is marked selected by default, while operators can inspect and select alternatives in the comparison modal.
6. **Execution & Verification**: Executable steps enter the Action System with condition checks and verification criteria.

---

## 4. Goal Model

Planning goals are first-class persistent entities stored in `planning_goals`:

```sql
CREATE TABLE IF NOT EXISTS planning_goals (
    goal_id VARCHAR(64) PRIMARY KEY,
    org_id BIGINT NOT NULL,
    module VARCHAR(50) NOT NULL,
    related_entity_type VARCHAR(50) NOT NULL,
    related_entity_id VARCHAR(100) NOT NULL,
    objective TEXT NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'MEDIUM',
    risk_tolerance VARCHAR(20) NOT NULL DEFAULT 'BALANCED',
    autonomy_level VARCHAR(50) NOT NULL DEFAULT 'LEVEL_2_PREPARE',
    hard_constraints JSON NULL,
    soft_constraints JSON NULL,
    success_criteria JSON NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_planning_goals_org_module (org_id, module)
);
```

---

## 5. Context Model

The Go `OperationalContextAssembler` (`backend/internal/autonomy/context_assembler.go`) safely constructs operational context for Python:
- **Tenant Isolation**: Guarantees queries are filtered by `org_id`.
- **Freshness Classification**: Evaluates `updated_at` against staleness windows:
  - `FRESH`: Data modified within recent operational threshold.
  - `STALE`: Data unmodified for > 7 days.
  - `ESTIMATED`: No live DB record found; fallback estimates provided.
- **Data Sufficiency**: Signals whether sufficient data exists to formulate confident plans.
- **Provenance**: Preserves data source tags (`LIVE_OPERATIONAL_DATABASE`).
- **Predictive Context**: Integrates Phase 4 signals (`predicted_delay_risk`, `predicted_delay_hours`).

---

## 6. Constraint Model

The system explicitly distinguishes **Hard Constraints** from **Soft Constraints**:

| Constraint Dimension | Hard Constraint Example | Soft Constraint Example |
|---|---|---|
| **TIME** | Delivery deadline before customs warehouse demurrage | Preferred morning delivery window |
| **COST** | Additional budget capped at $250.00 | Minimize additional operational expense |
| **SERVICE** | Required temperature-controlled reefer handling | Preferred top-tier carrier SCAC |
| **CUSTOMER** | Contractual SLA with liquidated damages clause | Communication cadence (e.g. notify every 12h) |
| **COMPLIANCE** | Prohibited from transshipping through sanctioned ports | Preferred green-lane documentation customs |
| **POLICY** | Cannot execute autonomous refunds > $500 | Prefer low-step plans requiring fewer approvals |

---

## 7. Candidate Plan Generation

For any planning goal, Python synthesizes 3 distinct operational strategies rather than a single monolithic path:
1. **Candidate A**: Accelerated Carrier Recovery (direct expedite).
2. **Candidate B**: Alternative Express Route & Transshipment (express reroute).
3. **Candidate C**: Enhanced Milestone Monitoring & Proactive Escalation (conservative).

Unknown or arbitrary actions generated outside supported Action System capabilities (`shipments.*`, `notifications.*`, `documents.*`, `approvals.*`) are strictly rejected by the Go validation layer.

---

## 8. Plan Evaluation

Plans are evaluated across 14 multi-attribute dimensions:
- `feasibility`: Boolean feasibility flag.
- `hard_constraints_satisfied`: Boolean check against all hard boundaries.
- `delay_reduction_hours`: Expected delay hours mitigated.
- `estimated_cost`: Expected incremental monetary expense.
- `customer_impact`: Risk level (`LOW`, `MEDIUM`, `HIGH`, `CRITICAL`).
- `operational_risk`: Risk level of plan execution.
- `compliance_risk`: Risk level regarding regulatory/contractual exposure.
- `confidence`: Confidence score (0.0 to 1.0).
- `reversibility`: `REVERSIBLE`, `PARTIALLY_REVERSIBLE`, `IRREVERSIBLE`.
- `step_count`: Number of steps (simplicity preference).
- `dependency_risk`: Risk from chained dependencies.
- `soft_constraints_score`: Normalized score for soft constraint satisfaction.
- `overall_utility_score`: Multi-objective utility rating (0.0 to 1.0).
- `selection_rationale`: Human-readable operational explanation.

---

## 9. Plan Ranking

Candidate plans are ranked deterministically:
1. **Hard Constraint Gating**: Any candidate violating a hard constraint receives `is_feasible = False`, `overall_utility_score = 0.0`, and is placed at the bottom of the rank list.
2. **Utility Function**: Feasible candidates are ranked by weighted multi-attribute utility:
   $$\text{Utility} = 0.35 \times \text{DelayRecovery} - 0.25 \times \text{CostPenalty} - 0.15 \times \text{RiskPenalty} + 0.15 \times \text{SoftScore} + 0.10 \times \text{Simplicity}$$
3. **Optimal Selection**: The top-ranked feasible candidate becomes the active plan.

---

## 10. Risk-Aware Planning

The engine incorporates Phase 4 risk intelligence:
- Integrates predicted delay risk and exception probabilities into candidate generation.
- Flags high-risk steps with `requires_approval = True` regardless of autonomy level.
- Rejects candidate selection if predicted risk exceeds tenant policy limits.

---

## 11. Cost-Aware Planning

- Evaluates current expected cost versus incremental strategy cost.
- Checks hard cost caps (e.g. `Max budget $250`).
- If cost information is unavailable, plans lower confidence and flag human review.
- Never fabricates financial numbers.

---

## 12. Customer-Impact Awareness

- Incorporates customer SLA commitments and priority.
- External notifications are restricted under policy: if autonomy level < 3, customer notifications require explicit HITL approval before dispatch.

---

## 13. Autonomy Integration

Implements the 5 standard autonomy levels:
- **Level 0 (Observe)**: Telemetry and monitoring only; no execution.
- **Level 1 (Recommend)**: Proposal only; requires user initiation.
- **Level 2 (Prepare)**: Complete executable plan generated; all steps require human approval.
- **Level 3 (Controlled)**: Pre-approved low-risk actions execute autonomously under policy.
- **Level 4 (Multi-Step)**: Chained dependent steps execute under strict policy boundaries.

---

## 14. Approval Integration

- Every plan evaluates approval requirements per step and per policy.
- Steps requiring approval pause in `AWAITING_APPROVAL`.
- Approvals record audit logs with user ID, timestamp, and rationale.

---

## 15. Execution Boundary

The Go backend enforces the execution boundary:
- Validates plan schema and action types before persisting.
- Gating checks: Disallows selection or execution of infeasible candidates.
- Dispatches steps through the Action System with idempotency keys.

---

## 16. Verification Criteria

Every step supports structured verification criteria:
```json
{
  "check_type": "STATUS_TRANSITION",
  "target_field": "status",
  "expected_value": "IN_TRANSIT",
  "description": "Verify shipment status transitioned to IN_TRANSIT"
}
```
Steps are only marked `COMPLETED` when verification passes.

---

## 17. Fallback Planning

Each step can define a structured `fallback_action`:
- If an expedite action fails or times out, the fallback triggers operations escalation.
- Fallback actions must also pass authorization and policy checks.
- Recursive fallback loops are strictly disallowed.

---

## 18. Staleness & Revalidation

Operational plans can become stale if business conditions shift:
- Plan age > 24 hours triggers `REVALIDATION_REQUIRED`.
- The `/api/v1/autonomy/plans/{id}/revalidate` endpoint gathers fresh context and evaluates whether plan assumptions still hold.
- Stale plans cannot execute without revalidation.

---

## 19. Event Integration

- Business events (e.g. exception detected, ETA slippage) trigger operational planning.
- Deduplication keys prevent redundant plan generation storms.
- Event ID and correlation ID are tracked end-to-end.

---

## 20. UI Implementation

The user experience is fully native to LogisticsHQ:
- **Autonomy Tab in AI Monitoring Dashboard**: Shows active plans, autonomy guide, and a prominent "+ New Operational Goal" button.
- **Formulate Operational Goal Modal**: Allows specifying module, entity ID, objective, and budget caps.
- **PlanLifecycleCard**: Renders selected candidate, delay reduction, estimated cost, staleness badge, and step dependency list.
- **PlanComparisonModal**: Side-by-side comparison of Candidates A, B, and C with feasibility badges, utility scores, and trade-off metrics.
- **Shipments Integration**: Contextual "AI Operational Plan" trigger in shipment row actions dropdown.

---

## 21. Security

- **Multi-Tenant Isolation**: Tested and verified. Tenant 2 cannot access Tenant 1 goals or candidates (`403/404`).
- **Authorization**: Goal creation and candidate selection require valid JWT bearer tokens.
- **Prompt Injection Defense**: Tested with prompt injection attacks (`DROP TABLE`, `sudo rm -rf`); sidecar sanitized inputs and maintained strict schema integrity.

---

## 22. Failure Handling

- **Sidecar Unavailable**: Go returns structured fallback error; does not crash.
- **Infeasible Candidate Selection**: Rejected with `ErrCandidateInfeasible` (HTTP 400).
- **Stale Context**: Flagged with `DataFreshness: STALE` and triggers revalidation.

---

## 23. Performance

- Python candidate generation and ranking executes in < 50ms.
- Go context assembly executes in < 5ms.
- Complete plan generation cycle takes < 60ms.

---

## 24. Tests

### Go Unit Tests (`internal/autonomy`)
- `TestOperationalPlanning_GoalCreationAndCandidateGeneration`: PASS
- `TestOperationalPlanning_HardConstraintSelectionGating`: PASS
- `TestOperationalPlanning_StalenessRevalidation`: PASS
- `TestOperationalPlanning_MultiTenantIsolation`: PASS
- Total: 9/9 PASS

### Python AI Sidecar Tests (`tests/test_operational_planning.py`)
- `test_candidate_plan_generation`: PASS
- `test_hard_constraint_violation_disqualification`: PASS
- `test_insufficient_data_handling`: PASS
- `test_prompt_injection_sanitization`: PASS
- `test_multi_objective_ranking_order`: PASS
- Total: 5/5 PASS

### End-to-End Integration Suite (`scripts/test_task52_operational_planning.py`)
- `context_assembly`: PASS
- `goal_formulation`: PASS
- `candidate_evaluation`: PASS
- `infeasible_candidate_rejection`: PASS
- `candidate_selection`: PASS
- `staleness_revalidation`: PASS
- `tenant_isolation`: PASS
- `prompt_injection_defense`: PASS
- Total: 8/8 PASS

---

## 25. Browser Results

Automated browser testing via Playwright (`scripts/test_task52_browser_ui.py`):
- Operational Goal Modal: PASS
- Plan Generation & Card Rendering: PASS
- Candidate Comparison Modal: PASS
- Revalidation Trigger: PASS
- Viewports:
  - 320x800 (Mobile Mini): PASS (0 horizontal overflow)
  - 375x812 (iPhone SE): PASS (0 horizontal overflow)
  - 768x1024 (Tablet Portrait): PASS (0 horizontal overflow)
  - 1024x768 (Tablet Landscape): PASS (0 horizontal overflow)
  - 1280x720 (HD Laptop): PASS (0 horizontal overflow)
  - 1440x900 (Desktop): PASS (0 horizontal overflow)
  - 1920x1080 (Full HD): PASS (0 horizontal overflow)
- Zoom Levels:
  - 80%: PASS
  - 90%: PASS
  - 100%: PASS
  - 110%: PASS
  - 125%: PASS
  - 150%: PASS

---

## 26. Regression Results

Core modules verified:
- `/dashboard`: PASS (1872ms, healthy)
- `/dashboard/shipments`: PASS (2308ms, healthy)
- `/dashboard/approvals`: PASS (2059ms, healthy)
- `/dashboard/recommendations`: PASS (2288ms, healthy)
- `/dashboard/invoices`: PASS (1815ms, healthy)
- `/dashboard/contracts`: PASS (1922ms, healthy)

---

## 27. Known Limitations

- Multi-carrier dynamic handoff in Candidate B currently relies on SCAC-based API routing definitions; additional live carrier booking adapters will expand in Task 5.3.
- Historical staleness threshold is fixed at 7 days for shipment updates; future work will allow tenant-specific configurable staleness policies.

---

## 28. Production Readiness

All Phase 5 Task 5.2 acceptance criteria have been verified and passed. The intelligent operational planning engine is production-ready.
