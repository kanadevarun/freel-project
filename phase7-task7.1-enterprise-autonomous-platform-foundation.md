# Phase 7.1 — Enterprise Autonomous Platform Foundation

**Final Status:** **PASS — TASK 7.1 COMPLETE**

---

## 1. Implementation Summary

LogisticsHQ Phase 7.1 establishes the Enterprise Autonomous Operations Platform Foundation by connecting and hardening the capabilities built across Phases 1 through 6 (Multi-Agent Specialist AI Workforce, Action System boundary, Human-in-the-Loop Approval Gating, Governed Autonomy Levels 0–4, Memory, Predictive Intelligence, and Observability).

The core of Phase 7.1 is a centralized control plane (`internal/enterprise_autonomy`) that provides deterministic lifecycle management, durable persistence, cross-module orchestration, restart recovery, tenant isolation, and strict policy enforcement without creating duplicate queues, databases, or AI orchestration engines.

### Key Tenets Maintained
1. **Python Remains the Sole AI Intelligence Layer**: Specialist agents (planning, shipment, exception, customer, pricing, finance, contract, compliance, monitoring, memory), LangGraph workflows, prompt engineering, predictive reasoning, and LLM inferences remain exclusively inside Python (`ai_sidecar`).
2. **Go Remains the Sole Application-Control & Enforcement Boundary**: Authentication, tenant isolation (`WHERE org_id = ?`), RBAC, Action System invocation, HITL approval enforcement, autonomy policy checks, database persistence, event deduplication, and crash recovery are executed exclusively in Go.
3. **Python Has Zero Direct Database Mutation**: Python never executes raw SQL, never writes directly to business tables, never bypasses Go, and cannot elevate its own autonomy permissions.
4. **Idempotency & Safe Recovery**: Workflows survive process crashes, Go restarts, Python restarts, or worker disruptions. Recovery is idempotent, skips completed steps, and preserves correlation IDs and audit history.

---

## 2. Architecture Changes

```
┌────────────────────────────────────────────────────────────────────────┐
│                          Go Control Plane                              │
│                                                                        │
│   HTTP Layer: /api/v1/enterprise/workflows                             │
│               /workflows/{id}/[pause|resume|cancel]                    │
│               /steps/{step_id}/[approve|reject]                        │
│               /workflows/recover                                       │
│               /events/trigger                                          │
│               /emergency-control                                       │
│               /overview                                                │
│                                                                        │
│   Enterprise Autonomy Service (internal/enterprise_autonomy)           │
│   ├── Policy Context & Autonomy Level Enforcement (Level 0-4)          │
│   ├── Cross-Module Orchestration Pipeline Builders                    │
│   │   ├── Shipment Recovery (Shipment→Exception→Customer→Fin→Comp)     │
│   │   ├── Commercial Cycle (Lead→RFQ→Pricing→Customer→Quotation)      │
│   │   ├── Financial Collection (Invoice→Customer→Finance→Contract)     │
│   │   └── Cross-Module Risk (Shipment→Compliance→Contract→Finance)    │
│   ├── State Transition Validator (Deterministic FSM)                   │
│   ├── Event Deduplication Filter (action_idempotency_keys)             │
│   └── Interrupted Workflow Recovery Engine                             │
│                                                                        │
│   Enforcement Boundaries                                               │
│   ├── Server-Side Multi-Tenant Isolation (WHERE org_id = ?)           │
│   ├── Action System (internal/actions) -> IdempotencyStore             │
│   ├── Approval Service (internal/approvals) -> HITL Gating             │
│   └── Universal Audit System (internal/audit/service)                  │
└────────────────────────────────────┬───────────────────────────────────┘
                                     │ JSON RPC / HTTP Client
                                     ▼
┌────────────────────────────────────────────────────────────────────────┐
│                   Python AI Intelligence (ai_sidecar)                  │
│                                                                        │
│   - LangGraph Workflow Graph & State Graphs                            │
│   - 10 Specialized Workforce Agents                                    │
│   - LLM Reasoning, Risk Assessment, and Anomaly Detection              │
│   - Output Structured Plan DTOs -> Returned to Go for Validation       │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Existing Infrastructure Reused

Rather than introducing redundant services or tables, Phase 7.1 extended and connected existing Phase 1–6 infrastructure:
- **`autonomous_plans` Table**: Serves as the primary durable record for enterprise workflows (`plan_id`, `org_id`, `module`, `goal`, `status`, `autonomy_level`, `policy_decision`, `correlation_id`, `related_entity_type`, `related_entity_id`).
- **`autonomous_plan_steps` Table**: Serves as the durable step ledger tracking step state, assigned specialist agents, action types, dependencies, risk levels, approval requirements, and action system execution IDs.
- **`autonomy_policies` Table**: Stores tenant-scoped autonomy rules, allowed/prohibited actions, monetary limits, confidence thresholds, and emergency pause flags.
- **`action_idempotency_keys` Table**: Reused for both Action System operations and enterprise event deduplication (24-hour sliding TTL).
- **`actions.Service`**: Reused for executing permitted business-changing operations (`exceptions.execute_recovery`, `carrier_inquiry`, `customer_advisory`, `customs_broker_notification`, etc.) under strict audit and idempotency guarantees.
- **`approvals.Service`**: Reused for human-in-the-loop gating when steps require supervisor authorization.
- **`auditSvc.Service`**: Reused for universal audit event recording across all workflow state transitions and security violation detections.
- **`workforce.Service`**: Reused to interact with Phase 6 specialist domain agents.

---

## 4. New & Extended Components

### 1. `backend/internal/enterprise_autonomy/model.go`
- Defined `EnterpriseWorkflowType`:
  - `SHIPMENT_RECOVERY`
  - `COMMERCIAL_CYCLE`
  - `FINANCIAL_COLLECTION`
  - `CROSS_MODULE_RISK`
  - `ADAPTIVE_REPLANNING`
  - `COMPLIANCE_AUDIT`
  - `OPERATIONAL_RECOVERY`
- Defined `EnterpriseWorkflowState` FSM with `ValidateStateTransition(current, next)`:
  - `PENDING` -> `RUNNING`, `WAITING`, `PAUSED`, `CANCELLED`
  - `RUNNING` -> `WAITING`, `WAITING_FOR_APPROVAL`, `PAUSED`, `BLOCKED`, `ESCALATED`, `COMPLETED`, `FAILED`, `CANCELLED`
  - `WAITING_FOR_APPROVAL` -> `RUNNING`, `BLOCKED`, `ESCALATED`, `FAILED`, `CANCELLED`
  - `PAUSED` -> `RUNNING`, `CANCELLED`, `FAILED`
  - Terminal states (`COMPLETED`, `FAILED`, `CANCELLED`) strictly prohibit transition.
- Defined DTOs: `EnterprisePolicyContext`, `EnterpriseWorkflow`, `EnterpriseWorkflowStep`, `StartWorkflowRequest`, `PauseWorkflowRequest`, `EmergencyControlRequest`, `RecoveryReport`, and `EnterprisePlatformOverview`.

### 2. `backend/internal/enterprise_autonomy/repository.go`
- Implemented `Repository` interface backed by MariaDB:
  - `CreateWorkflow`, `GetWorkflow`, `UpdateWorkflowState`, `UpdateWorkflowStop`
  - `ListWorkflows`, `SaveWorkflowSteps`, `GetWorkflowSteps`, `UpdateWorkflowStep`
  - `GetActiveOrInterruptedWorkflows`, `GetPolicyContext`, `UpdatePolicyContext`
  - `CheckAndRecordEventDedup`, `GetPlatformOverview`
- Every query enforces strict server-side tenant isolation (`WHERE org_id = ?`).

### 3. `backend/internal/enterprise_autonomy/service.go`
- Orchestrates multi-step workflows across specialist agents:
  - `buildShipmentRecoverySteps`: Shipment Telemetry -> Exception Triage -> Customer Notification (HITL) -> Demurrage Audit -> Compliance Audit.
  - `buildCommercialCycleSteps`: RFQ Extraction -> Pricing Margin Optimization -> Credit Check (HITL) -> Contract Quotation Draft.
  - `buildFinancialCollectionSteps`: Aged Receivables Audit -> Collection Reminder (HITL) -> Contract Lien Clauses -> Balance Dispute Reconciliation.
  - `buildCrossModuleRiskSteps`: Transit Risk -> Customs & Sanctions Compliance -> Free Days & Force Majeure Terms -> Financial Exposure Impact (HITL).
- Autonomy gating:
  - `LEVEL_0_OBSERVE`: Refuses modifying workflows.
  - `LEVEL_1_RECOMMEND`: Generates recommendations only.
  - `LEVEL_2_PREPARE`: Prepares actions and requires approval.
  - `LEVEL_3_CONTROLLED_EXECUTION`: Executes permitted registered actions via Action System boundary with idempotency keys.
  - `LEVEL_4_GOVERNED_MULTI_STEP`: Multi-step governed autonomy.
  - Self-elevation attempts are rejected server-side and logged as security violations.
- Interrupted workflow recovery: `RecoverInterruptedWorkflows` identifies active/interrupted workflows on startup, validates policy and emergency stop, skips completed steps, and resumes execution safely.

### 4. `backend/internal/enterprise_autonomy/handler.go`
- REST HTTP API registered on router:
  - `POST /api/v1/enterprise/workflows`
  - `GET /api/v1/enterprise/workflows`
  - `GET /api/v1/enterprise/workflows/{id}`
  - `POST /api/v1/enterprise/workflows/{id}/pause`
  - `POST /api/v1/enterprise/workflows/{id}/resume`
  - `POST /api/v1/enterprise/workflows/{id}/cancel`
  - `POST /api/v1/enterprise/workflows/{id}/steps/{step_id}/approve`
  - `POST /api/v1/enterprise/workflows/{id}/steps/{step_id}/reject`
  - `POST /api/v1/enterprise/workflows/recover`
  - `POST /api/v1/enterprise/events/trigger`
  - `POST /api/v1/enterprise/emergency-control`
  - `GET /api/v1/enterprise/overview`

### 5. `frontend/src/services/enterpriseService.js`
- Clean client service providing API bindings to all `/api/v1/enterprise` endpoints.

### 6. `frontend/src/components/workforce/AIWorkforceCommandCenter.jsx`
- Added the **Enterprise Workflows** tab to the existing Command Center.
- Displays active enterprise workflows, current state badges, assigned specialist agents, and action buttons (Inspect, Pause, Resume, Cancel).
- Header provides a 1-click **Recover Interrupted Workflows** control.
- Detailed step inspection modal displays step progression, agent assignments, risk ratings, and execution results.

---

## 5. Workflow Lifecycle & State FSM

```
              ┌─────────┐
              │ PENDING │
              └────┬────┘
                   │
                   ▼
              ┌─────────┐       Requires Approval       ┌──────────────────────┐
              │ RUNNING ├──────────────────────────────►│ WAITING_FOR_APPROVAL │
              └──┬──┬─┬─┘                               └───┬──────────────┬───┘
                 │  │ │                                     │              │
      Paused by  │  │ │ Finished All Steps      Approved by │   Rejected by│
      Operator   │  │ └───────────────────────┐ Human       │   Human      │
                 ▼  │                         ▼             ▼              ▼
           ┌────────┴┐                  ┌───────────┐ ┌─────────┐   ┌─────────────┐
           │ PAUSED  │                  │ COMPLETED │ │ RUNNING │   │   BLOCKED   │
           └────┬────┘                  └───────────┘ └─────────┘   └──────┬──────┘
                │ Resumed by Operator                                      │ Escalated
                └─────────────────────────────┐                            ▼
                                              ▼                     ┌─────────────┐
                                        ┌───────────┐               │  ESCALATED  │
                                        │  RUNNING  │               └─────────────┘
                                        └───────────┘
```

---

## 6. Persistence & Crash Recovery Behavior

1. **State Persistence**:
   - Workflows are persisted in `autonomous_plans` immediately upon initiation with a durable UUID/ID and correlation ID.
   - Discrete steps are persisted in `autonomous_plan_steps` with prerequisites and unique idempotency keys.
   - Updates to workflow state and step execution results are committed to MariaDB before proceeding to subsequent steps.
2. **Recovery Across Restart**:
   - `RecoverInterruptedWorkflows` queries for workflows in `RUNNING`, `PENDING`, or `WAITING` states.
   - Checks organization policy to ensure emergency pause has not been engaged.
   - Loads completed steps and resumes only pending/unexecuted steps.
   - Reuses existing idempotency keys to ensure no duplicate actions are executed in external systems or the Action System.
   - If a step was waiting for approval prior to crash, it safely remains in `WAITING_FOR_APPROVAL` without state corruption.

---

## 7. Autonomy Enforcement & Security Controls

| Autonomy Level | Permitted Behaviors | Enforcement Boundary |
|---|---|---|
| **Level 0 — Observe** | Read-only telemetry, anomaly logging. Autonomous execution blocked. | Go server rejects `StartWorkflow` (`ErrAutonomyRestricted`). |
| **Level 1 — Recommend** | AI generates plans and recommendations; no modifying actions. | Go executes steps in recommendation/advisory mode only. |
| **Level 2 — Prepare** | Prepares payload, stages quotes/notifications; requires approval before execution. | Go transitions step to `WAITING_FOR_APPROVAL` and halts until supervisor confirms. |
| **Level 3 — Controlled Execution** | Executes approved and registered actions within defined thresholds via Action System. | Go validates action registry, checks idempotency store, logs universal audit. |
| **Level 4 — Governed Multi-Step** | Multi-step chained actions under policy thresholds with automatic pause on anomaly. | Go enforces step dependencies, monetary ceilings, and confidence minimums. |

### Security Defenses Verified
- **Self-Elevation Defense**: If Python or a client requests `LEVEL_4` when the organization policy is `LEVEL_2`, Go rejects the request and logs `AUTONOMY_ELEVATION_REJECTED` in universal audit logs.
- **Tenant Isolation Defense**: Cross-tenant workflow inspection or control requests return `ErrWorkflowNotFound` or `ErrUnauthorizedTenant`. Server-side SQL queries enforce `WHERE org_id = ?`.
- **Duplicate Event Protection**: Repeated business events with identical event IDs or idempotency keys are detected via `action_idempotency_keys` and rejected (`ErrDuplicateEventTrigger`).
- **Emergency Halt / Kill-Switch**: Activating emergency stop at the tenant or global scope instantly halts workflow execution, blocks new workflow creation, and prevents restart recovery.

---

## 8. Test Execution & Verification

Comprehensive unit, security, and live integration tests were authored in `backend/internal/enterprise_autonomy/enterprise_test.go`:

```
=== RUN   TestValidateStateTransition
--- PASS: TestValidateStateTransition (0.00s)
=== RUN   TestEnterpriseTenantIsolation
--- PASS: TestEnterpriseTenantIsolation (0.00s)
=== RUN   TestCrossModuleOrchestration
--- PASS: TestCrossModuleOrchestration (0.00s)
=== RUN   TestAutonomyEnforcementAndElevationDefense
--- PASS: TestAutonomyEnforcementAndElevationDefense (0.00s)
=== RUN   TestApprovalWorkflowGating
--- PASS: TestApprovalWorkflowGating (0.00s)
=== RUN   TestInterruptedWorkflowRecovery
--- PASS: TestInterruptedWorkflowRecovery (0.00s)
=== RUN   TestEventDeduplication
--- PASS: TestEventDeduplication (0.00s)
=== RUN   TestEnterpriseEmergencyControl
--- PASS: TestEnterpriseEmergencyControl (0.00s)
=== RUN   TestPlatformOverview
--- PASS: TestPlatformOverview (0.00s)
=== RUN   TestPythonContractValidation
--- PASS: TestPythonContractValidation (0.00s)
=== RUN   TestLiveMySQLRepository
--- PASS: TestLiveMySQLRepository (0.02s)
PASS
ok      github.com/freel/backend/internal/enterprise_autonomy   0.762s
```

### Workforce Phase 6 Regression Verification
```
=== RUN   TestWorkforceAgentRegistryAndCapabilities
--- PASS: TestWorkforceAgentRegistryAndCapabilities (0.00s)
=== RUN   TestPhase69_CommandCenterOverview_TenantIsolation
--- PASS: TestPhase69_CommandCenterOverview_TenantIsolation (0.00s)
PASS
ok      github.com/freel/backend/internal/workforce             0.798s
```

### Full Backend Server Build
```powershell
go build -v ./cmd/server
# Result: Code 0 — Clean build with 0 compile errors
```

### Frontend Build Verification
```powershell
npm.cmd run build
# Result: vite v8.0.12 built in 19.20s with Code 0 — 0 errors
```

### Sidecar Liveness
```powershell
Invoke-RestMethod -Uri 'http://127.0.0.1:8090/health'
# Result: status: "ok", checkpointer: "MariaDBSaver", persistent: true
```

---

## 9. Files Changed / Created

| File | Status | Description |
|---|---|---|
| `backend/internal/enterprise_autonomy/model.go` | Created | Defined enterprise workflow types, FSM states, policy context, step models, and DTOs. |
| `backend/internal/enterprise_autonomy/repository.go` | Created | SQL repository implementing tenant-isolated queries on `autonomous_plans`, `autonomous_plan_steps`, and `action_idempotency_keys`. |
| `backend/internal/enterprise_autonomy/service.go` | Created | Enterprise orchestration service, step builders, autonomy enforcement, Action System bridge, and restart recovery engine. |
| `backend/internal/enterprise_autonomy/handler.go` | Created | REST HTTP handlers for enterprise operations, approvals, emergency stop, and recovery. |
| `backend/internal/enterprise_autonomy/enterprise_test.go` | Created | 11 comprehensive unit, security, and live MariaDB integration tests. |
| `backend/internal/server/server.go` | Modified | Wired `RegisterEnterpriseAutonomyRoutes` into the router with JWT authentication. |
| `backend/cmd/server/main.go` | Modified | Instantiated repository, service, handler, and registered endpoints. |
| `frontend/src/services/enterpriseService.js` | Created | Frontend API client for enterprise autonomy operations. |
| `frontend/src/components/workforce/AIWorkforceCommandCenter.jsx` | Modified | Integrated Enterprise Autonomous Workflows tab, step inspection modal, and restart recovery trigger. |

---

## 10. Known Limitations & Remaining Follow-Up Items

1. **Follow-Up for Phase 7.2 (Enterprise Control Tower)**: Phase 7.1 establishes the robust architectural and operational foundation. Phase 7.2 will expand visual observability with real-time multi-agent graph telemetry, SLA tracking, and advanced cross-module decision support.
2. **Scheduled Cron Trigger Integration**: Event-triggered workflows and API-triggered workflows are fully operational. Scheduled periodic crons can be linked to existing `workflow_automation` scheduled jobs in subsequent phases.

---

## 11. Final Status

**PASS — TASK 7.1 COMPLETE**
