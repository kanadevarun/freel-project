# LogisticsHQ — Phase 5, Task 5.12: Autonomous Operations Command Center

## Executive Summary & System Architecture

The **Autonomous Operations Command Center** serves as the authoritative, real-time control surface for human and AI collaborative logistics operations across LogisticsHQ. It provides forwarders, dispatchers, finance officers, and executive operators immediate, cross-functional visibility into autonomous planning, continuous monitoring, adaptive replanning, multi-domain risks, pending approvals, human interventions, and active workflows.

```
+----------------------------------------------------------------------------------------------------+
|                                    LogisticsHQ Web Application                                      |
|                                     (/dashboard/command-center)                                    |
|                                                                                                    |
|  +----------------------------------------------------------------------------------------------+  |
|  | Subsystem Health Ribbon (Action System, Approvals, MySQL, Events, Go, Python, Autonomy Engine)|  |
|  +----------------------------------------------------------------------------------------------+  |
|  | Fact (Actual State)  |  Prediction (AI Forecasts & Risks)  |  AI Analysis (Bounded Guidance) |  |
|  +----------------------------------------------------------------------------------------------+  |
|  | 6 Authoritative KPI Cards: Active Shipments, Critical Exceptions, Active Workflows, Decisions|  |
|  +----------------------------------------------------------------------------------------------+  |
|  | Autonomy Level Distribution Strip: Level 1 Assist | Level 2 Prepare | Level 3 Policy Execute  |  |
|  +----------------------------------------------------------------------------------------------+  |
|  | Tab 1: Critical Attention Feed (Prioritized with "Why Flagged" + Actual/Predicted/Action)    |  |
|  | Tab 2: Active AI Workflows (Plan Progress, Autonomy Lvl, Step Progression Preview)           |  |
|  | Tab 3: Needs Your Decision (Human-in-the-Loop Approval Queue with 1-Click Action)             |  |
|  | Tab 4: Multi-Domain Risk Matrix (Shipment, Finance, Customer, Compliance)                    |  |
|  | Tab 5: AI Actions & Lineage (Recent Actions, Replanning Lineage, Escalation Registry)        |  |
|  | Tab 6: System Health (Authoritative Latency, Status, Uptime, Service Health)                |  |
|  +----------------------------------------------------------------------------------------------+  |
|                                                |                                                   |
|                        Authoritative REST API  |  Axios Client                                     |
|                                                v                                                   |
+----------------------------------------------------------------------------------------------------+
                                                 |
                                                 v
+----------------------------------------------------------------------------------------------------+
|                                     Go Core Backend (:8080)                                        |
|                          (Authoritative Business State, Action System, RBAC)                       |
|                                                                                                    |
|  - GET /api/v1/autonomy/command-center/overview          -> Overview KPIs & System State           |
|  - GET /api/v1/autonomy/command-center/critical-attention -> Sidecar-Prioritized Ranked Feed        |
|  - GET /api/v1/autonomy/command-center/workflows         -> Authoritative Active Plans & Progress  |
|  - GET /api/v1/autonomy/command-center/decisions         -> Pending Approvals & Human Gates        |
|  - GET /api/v1/autonomy/command-center/risks             -> Multi-Domain Risk Aggregations         |
|  - GET /api/v1/autonomy/command-center/activity          -> Actions, Replanning Events, Escalations|
|  - GET /api/v1/autonomy/command-center/system-health     -> Subsystem Ping Latencies & Status      |
|                                                |                                                   |
|                        Structured JSON (HTTP)  |  Sidecar Client                                   |
|                                                v                                                   |
+----------------------------------------------------------------------------------------------------+
                                                 |
                                                 v
+----------------------------------------------------------------------------------------------------+
|                                    Python AI Sidecar (:8090)                                       |
|                         (Reasoning, Prioritization Scoring, Prompt Sanitization)                   |
|                                                                                                    |
|  - POST /autonomy/command-center/prioritize                                                        |
|      * Multi-factor scoring (Safety/Compliance = 95, Operational = 85, Customer = 75, Finance = 65) |
|      * Prompt injection neutralization regex & instruction scrubbers                               |
|      * Concise operational explanations without internal chain-of-thought leaking                  |
+----------------------------------------------------------------------------------------------------+
```

---

## 1. Non-Negotiable AI Architecture
- **Separation of Concerns**:
  - **Python AI Sidecar**: Exclusively performs multi-factor priority scoring, cross-domain risk interpretation, prompt-injection sanitization, and concise operator summarization.
  - **Go Backend Core**: Authoritative database queries (`freel_mysql`), strict tenant isolation (`org_id`), Action System boundary governance, human approval lifecycle, audit trails, and execution state.
- **Zero Duplicate Engines**: Reuses existing Phase 5 domain engines (Multi-Step Planning, Adaptive Replanning, Exception Resolution, Pricing Optimization, Finance Collections, Contract Compliance, Continuous Monitoring, and Human-in-the-Loop Decision Center).
- **Graceful Fallback**: If the Python sidecar is temporarily degraded or unreachable, the Go backend falls back to deterministic rule-based priority scoring without interrupting operator workflows.

---

## 2. Core Answers Provided by the Command Center
1. **WHAT IS HAPPENING?** Real-time KPI strip and Active AI Workflows tab showing all in-flight plans across shipments, exceptions, pricing, finance, and contracts.
2. **WHAT IS AT RISK?** Multi-Domain Risk Matrix aggregating cross-functional exposure across shipments (delays, temperature excursion), finance (overdue receivables, margin leakage), customer SLAs, and regulatory compliance holds.
3. **WHAT IS AI DOING?** AI Actions & Lineage tab displaying recently executed tool dispatches, execution mode (`AUTONOMOUS` vs `HUMAN_APPROVED`), and step progression.
4. **WHAT IS WAITING?** Waiting state indicators (e.g. `WAITING_FOR_CARRIER`, `WAITING_FOR_DOCUMENT`, `WAITING_FOR_ETA`) with step-level timestamps.
5. **WHAT NEEDS A HUMAN?** "Needs Your Decision" tab with human-in-the-loop approvals, override buttons, and reason inputs.
6. **WHAT IS EXECUTING?** Active workflow execution table displaying plan health (`HEALTHY`, `AT_RISK`, `BLOCKED`, `STALLED`).
7. **WHAT FAILED?** Escalation Registry displaying blocked plans, policy violations, and recommended dispatcher recovery actions.
8. **WHAT CHANGED?** Replanning Lineage table displaying state transitions (`old_status -> new_status`), trigger reasons, and changed assumptions.
9. **WHAT WILL HAPPEN NEXT?** Multi-step progression nodes displaying sequential milestones (Observe -> Prepare -> Review -> Execute -> Verify).
10. **WHAT NEEDS ATTENTION NOW?** Critical Attention feed ranked by severity, business urgency, operational impact, and deadlines.

---

## 3. Provenance Strip: Fact vs. Prediction vs. AI Analysis
Following Section 29 requirements, every operational entity clearly distinguishes its evidentiary basis:
- **ACTUAL (Authoritative Facts)**: Blue tag indicating verified records loaded from the MySQL database (e.g., recorded milestone departures, issued invoices, filed customs documents).
- **PREDICTED (AI Forecasts & Risks)**: Amber tag indicating statistical/ML estimations (e.g., port congestion probability, delivery slip hours, payment default risk).
- **AI ANALYSIS (Operating Bounds & Reasoning)**: Purple tag indicating LangGraph policy bounds and recommended operator recovery strategies.

---

## 4. Multi-Domain Risk Matrix
Aggregates four core operational pillars with interactive sub-tabs:
1. **Shipment & Transit Health**: Monitors vessel AIS telemetry, port congestion, weather disruptions, customs holds, and milestone deviations.
2. **Finance & Collections**: Tracks overdue receivables, credit ceiling breaches, dispute holds, and high-exposure invoices.
3. **Customer Commitment Radar**: Quantifies SLA breach probability, sensitive customer tier accounts, and proactive notification status.
4. **Contract & Compliance**: Evaluates GDP pharma certifications, dangerous goods declarations, sanctioned lane checks, and expiring rate agreements.

---

## 5. Subsystems Health Ribbon
The top-level operational ribbon queries and aggregates live latencies and statuses across 7 vital subsystems:
- `action_system`: Safe boundary executor (Latency: ~1ms, Status: HEALTHY)
- `approval_service`: Dual-approval and governance registry (Latency: ~1ms, Status: HEALTHY)
- `database`: MariaDB / MySQL cluster (Latency: ~2ms, Status: HEALTHY)
- `event_processing`: Ingestion and deduplication engine (Latency: ~1ms, Status: HEALTHY)
- `go_backend`: Core authoritative service (Latency: ~0ms, Status: HEALTHY)
- `python_ai_service`: LangGraph & Priority reasoning sidecar (Latency: ~2ms, Status: HEALTHY)
- `worker_daemon`: Background async event consumers (Latency: ~1ms, Status: HEALTHY)

---

## 6. Prompt Injection Defense & Sanitization
The Python sidecar (`app/autonomy/command_center_agent.py`) intercepts all incoming items and scrubs instruction overrides, system role manipulations, and jailbreak patterns before priority scoring:
- Regex neutralizing `ignore previous instructions`, `system override`, `escalate privileges`, and `format: raw`.
- Content replaced with `[SANITIZED_INSTRUCTION]` tokens.
- Structured output guaranteed via Pydantic model `CommandCenterPrioritizeResponse`.

---

## 7. Responsive & Zoom Verification
The Command Center has been verified across 7 responsive viewport resolutions and 6 zoom levels using Playwright:
- **320x800 (Mobile Mini)**: PASS — Responsive layout, horizontal scroll for tables, touch-friendly touch targets.
- **375x812 (iPhone SE)**: PASS — Stacked KPI grid, collapsible filters.
- **768x1024 (Tablet Portrait)**: PASS — 2-column KPI grid, accessible tabs.
- **1024x768 (Tablet Landscape)**: PASS — Standard tablet layout.
- **1280x720 (HD Laptop)**: PASS — Desktop grid.
- **1440x900 (Desktop)**: PASS — Full-featured multi-column workspace.
- **1920x1080 (Full HD)**: PASS — Edge-to-edge operational control room view.
- **Zoom Levels (80%, 90%, 100%, 110%, 125%, 150%)**: PASS — Text scaling intact, zero overlapping controls.

---

## 8. Real Persistent Business Data Integrity
- Real records for **Shipment 101 (`BK-2026-DEV-001`)** in Org 2 remain completely intact.
- Verified 3 active milestones (Order Placed, Cargo Received, Vessel Departed) and 3 operational exceptions (`exc-101`, `exc-102`, `exc-103`).
- Weather Disruption `exc-103` on Shipment 101 was ranked as Priority #1 (Score 85, Immediate Urgency) without any synthetic data modification or database reset.
