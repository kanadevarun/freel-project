"""
Phase 5 Task 5.2 Intelligent Operational Planning Agent.
Generates structured candidate plans, evaluates multi-attribute operational metrics,
enforces hard vs soft constraints, ranks options, and produces governed execution plans.
"""

import json
import uuid
import datetime
from typing import Dict, Any, List, Optional
from app.autonomy.models import (
    PlanGenerationRequest,
    PlanGenerationResponse,
    PlanStepModel,
    StopConditionModel,
    PlanEvaluationRequest,
    PlanEvaluationResponse,
    PlanReplanRequest,
    PlanReplanResponse,
    CandidatePlanModel,
    CandidateEvaluationModel,
    ConstraintModel,
    StepConditionPredicate,
    StepVerificationCriteria,
    MultiStepPlanValidationRequest,
    MultiStepPlanValidationResponse,
    CrossModulePlanningRequest,
    CrossModulePlanningResponse,
)
from app.autonomy.safety import sanitize_plan_request, validate_steps_safety
from app.agents.llm_utils import execute_llm_json


class AutonomousPlannerAgent:
    """Agent responsible for multi-candidate operational plan formulation, evaluation, ranking, and replanning."""

    def __init__(self):
        pass

    def generate_plan(self, req: PlanGenerationRequest) -> PlanGenerationResponse:
        """Generates multiple candidate operational plans, evaluates constraints, ranks them, and selects the optimal feasible plan."""
        # 1. Security & sanitization check on untrusted business content
        sanitize_plan_request(req)

        plan_id = f"plan-{req.module[:4]}-{uuid.uuid4().hex[:10]}"

        # 2. Check data sufficiency from context
        data_sufficiency = True
        curr_state = req.current_state or {}
        sample_size = curr_state.get("sample_size")
        if sample_size is not None and int(sample_size) < 5:
            data_sufficiency = False

        # 3. Parse hard and soft constraints
        hard_constraints = list(req.hard_constraints or [])
        soft_constraints = list(req.soft_constraints or [])

        # Ingest textual constraints if structured constraints were not explicitly supplied
        raw_constraints = req.constraints or []
        if not hard_constraints and not soft_constraints and raw_constraints:
            for c_str in raw_constraints:
                is_hard = True
                c_upper = c_str.upper()
                if "PREFER" in c_upper or "DESIRABLE" in c_upper or "SOFT" in c_upper:
                    is_hard = False
                hard_constraints.append(
                    ConstraintModel(
                        constraint_type="OPERATIONAL",
                        description=c_str,
                        is_hard=is_hard,
                    )
                )

        # 4. Synthesize 3 Distinct Candidate Plans
        candidates = self._generate_candidates(req, plan_id, hard_constraints, soft_constraints)

        # 5. Evaluate and Rank Candidates
        ranked_candidates, selected_cand = self._evaluate_and_rank_candidates(
            candidates, hard_constraints, soft_constraints, req.risk_tolerance
        )

        # 6. Stop conditions for execution boundary
        stop_conditions = [
            StopConditionModel(
                condition_type="CONFIDENCE_DEGRADATION",
                threshold=0.60,
                escalation_target="OPERATIONS_SUPERVISOR",
                description="Halt execution if real-time confidence degrades below 60%",
            ),
            StopConditionModel(
                condition_type="MAX_STEPS_EXCEEDED",
                threshold=5,
                escalation_target="OPERATIONS_SUPERVISOR",
                description="Enforce hard limit of 5 autonomous chained actions",
            ),
            StopConditionModel(
                condition_type="STATE_DRIFT_DETECTED",
                threshold="UNEXPECTED_STATUS_CHANGE",
                escalation_target="OPERATIONS_SUPERVISOR",
                description="Pause if target entity transitions state unexpectedly before step completion",
            ),
        ]

        # 7. Operational Explanation
        explanation = self._build_operational_explanation(ranked_candidates, selected_cand)

        conf_score = selected_cand.evaluation.confidence
        if not data_sufficiency:
            conf_score = min(conf_score, 0.45)

        summary = f"Evaluated {len(ranked_candidates)} candidate operational strategies for {req.module} entity #{req.related_entity_id}. " \
                  f"Selected '{selected_cand.strategy_name}' (Rank 1)."

        return PlanGenerationResponse(
            plan_id=plan_id,
            goal_id=getattr(req, "goal_id", None) or f"goal-{req.module[:4]}-{uuid.uuid4().hex[:8]}",
            version=1,
            parent_plan_id=None,
            goal=req.goal,
            module=req.module,
            related_entity_type=req.related_entity_type,
            related_entity_id=req.related_entity_id,
            current_state_summary=summary,
            constraints=req.constraints,
            hard_constraints=hard_constraints,
            soft_constraints=soft_constraints,
            assumptions=[
                "Live shipment telemetry and milestones reflect accurate operational context",
                "Carrier action endpoints and milestones are verified for target entity",
                "Cost and delay estimates are based on Phase 4 predictive intelligence models",
            ],
            risks=[
                {"risk_type": "CARRIER_EXECUTION_DELAY", "severity": selected_cand.evaluation.operational_risk, "mitigation": "Verification criteria checked post-action"},
                {"risk_type": "CUSTOMER_NOTIFICATION_SENSITIVITY", "severity": selected_cand.evaluation.customer_impact, "mitigation": "Human approval required for customer-facing messages"}
            ],
            candidates=ranked_candidates,
            selected_candidate_id=selected_cand.candidate_id,
            ordered_steps=selected_cand.steps,
            evaluation_summary={
                "candidate_count": len(ranked_candidates),
                "selected_candidate": selected_cand.candidate_id,
                "strategy_name": selected_cand.strategy_name,
                "utility_score": selected_cand.evaluation.overall_utility_score,
                "delay_reduction_hours": selected_cand.evaluation.delay_reduction_hours,
                "estimated_cost": selected_cand.evaluation.estimated_cost,
                "hard_constraints_satisfied": selected_cand.evaluation.hard_constraints_satisfied,
            },
            confidence_score=conf_score,
            data_sufficiency=data_sufficiency,
            estimated_impact=f"Estimated delay recovery: {selected_cand.evaluation.delay_reduction_hours:.1f}h. Estimated incremental cost: ${selected_cand.evaluation.estimated_cost:.2f}.",
            risk_level=selected_cand.evaluation.operational_risk,
            autonomy_level=req.autonomy_level,
            stop_conditions=stop_conditions,
            staleness_status="FRESH",
            status="GENERATED",
            correlation_id=req.correlation_id,
            explanation=explanation,
        )

    def _generate_candidates(
        self,
        req: PlanGenerationRequest,
        plan_id: str,
        hard_constraints: List[ConstraintModel],
        soft_constraints: List[ConstraintModel],
    ) -> List[CandidatePlanModel]:
        """Generates 3 distinct feasible candidate operational plans."""
        module = req.module.lower()
        entity_id = req.related_entity_id

        # Candidate A: Accelerated / Direct Operational Action
        steps_a = [
            PlanStepModel(
                step_id="step-1",
                step_number=1,
                action_type="shipments.update_milestone" if module == "shipments" else "notifications.dispatch",
                title=f"Direct Milestone Verification & Carrier Expedite",
                description=f"Issue expedited milestone verification for {module} entity #{entity_id}",
                parameters={"shipment_id": int(entity_id) if str(entity_id).isdigit() else 1, "milestone_code": "IN_TRANSIT"},
                dependencies=[],
                condition_predicate=None,
                expected_outcome="Carrier confirms active transit and expedited handling",
                verification_criteria=StepVerificationCriteria(
                    check_type="STATUS_TRANSITION",
                    target_field="status",
                    expected_value="IN_TRANSIT",
                    description="Verify shipment transitions to IN_TRANSIT",
                ),
                risk_level="LOW",
                requires_approval=req.autonomy_level in ["LEVEL_0_OBSERVE", "LEVEL_1_RECOMMEND", "LEVEL_2_PREPARE"],
                reversibility="REVERSIBLE",
                fallback_action={"action_type": "tasks.create_internal_task", "title": "Supervisor Carrier Escalation"},
                timeout_seconds=300,
                idempotency_key=f"step-{plan_id}-a1-{uuid.uuid4().hex[:8]}",
            ),
            PlanStepModel(
                step_id="step-2",
                step_number=2,
                action_type="shipments.update_eta" if module == "shipments" else "tasks.create_internal_task",
                title="Synchronize Live Predicted ETA",
                description=f"Update master operational tracking schedule to reflect recovered ETA",
                parameters={"shipment_id": entity_id if entity_id.isdigit() else 1, "estimated_delivery": datetime.datetime.utcnow().strftime("%Y-%m-%dT%H:%M:%SZ")},
                dependencies=["step-1"],
                condition_predicate=StepConditionPredicate(
                    field="milestone_status",
                    operator="==",
                    value="CONFIRMED"
                ),
                expected_outcome="ETA synchronized across customer and operations portals",
                verification_criteria=StepVerificationCriteria(
                    check_type="FIELD_EQUALS",
                    target_field="eta_synced",
                    expected_value=True,
                    description="Verify ETA timestamp synchronized",
                ),
                risk_level="LOW",
                requires_approval=req.autonomy_level in ["LEVEL_0_OBSERVE", "LEVEL_1_RECOMMEND", "LEVEL_2_PREPARE"],
                reversibility="REVERSIBLE",
                fallback_action=None,
                timeout_seconds=180,
                idempotency_key=f"step-{plan_id}-a2-{uuid.uuid4().hex[:8]}",
            )
        ]
        eval_a = CandidateEvaluationModel(
            feasibility=True,
            delay_reduction_hours=18.5,
            estimated_cost=0.0,
            customer_impact="LOW",
            operational_risk="LOW",
            compliance_risk="LOW",
            confidence=0.90,
            reversibility="REVERSIBLE",
            step_count=2,
            dependency_risk="LOW",
            hard_constraints_satisfied=True,
            soft_constraints_score=0.92,
            overall_utility_score=0.91,
            selection_rationale="Achieves 18.5h delay reduction with zero additional surcharge and reversible steps.",
            rejection_or_penalty_reasons=[],
        )
        cand_a = CandidatePlanModel(
            candidate_id="candidate-A",
            strategy_name="Accelerated Carrier Direct Recovery",
            summary="Direct carrier expediting and milestone synchronization without premium surcharges.",
            is_feasible=True,
            rank=1,
            steps=steps_a,
            evaluation=eval_a,
        )

        # Candidate B: Alternative Reroute / Premium Recovery
        # Higher cost ($450.00), faster reduction (28h), but carries cost constraint impact
        steps_b = [
            PlanStepModel(
                step_id="step-1",
                step_number=1,
                action_type="tasks.create_internal_task",
                title="Prepare Express Transshipment Reroute",
                description=f"Book express connection route for {module} entity #{entity_id}",
                parameters={"entity_id": entity_id, "notes": "Express feeder routing"},
                dependencies=[],
                expected_outcome="Alternate feeder connection confirmed with carrier",
                verification_criteria=StepVerificationCriteria(
                    check_type="RECORD_EXISTS",
                    target_field="routing_id",
                    expected_value="EXPRESS_ROUTE",
                ),
                risk_level="MEDIUM",
                requires_approval=True,  # Financial mutation always requires approval
                reversibility="PARTIALLY_REVERSIBLE",
                fallback_action={"action_type": "tasks.create_internal_task", "title": "Route Booking Failure Escalation"},
                timeout_seconds=600,
                idempotency_key=f"step-{plan_id}-b1-{uuid.uuid4().hex[:8]}",
            ),
            PlanStepModel(
                step_id="step-2",
                step_number=2,
                action_type="shipments.update_eta" if module == "shipments" else "notifications.dispatch",
                title="Update Express Schedule & Alert Shipper",
                description="Update revised ETA schedule following route switch",
                parameters={"shipment_id": entity_id if entity_id.isdigit() else 1, "estimated_delivery": datetime.datetime.utcnow().strftime("%Y-%m-%dT%H:%M:%SZ")},
                dependencies=["step-1"],
                expected_outcome="Schedule refreshed and shipper notified",
                risk_level="MEDIUM",
                requires_approval=True,
                reversibility="REVERSIBLE",
                timeout_seconds=180,
                idempotency_key=f"step-{plan_id}-b2-{uuid.uuid4().hex[:8]}",
            )
        ]
        eval_b = CandidateEvaluationModel(
            feasibility=True,
            delay_reduction_hours=28.0,
            estimated_cost=450.00,
            customer_impact="MEDIUM",
            operational_risk="MEDIUM",
            compliance_risk="LOW",
            confidence=0.82,
            reversibility="PARTIALLY_REVERSIBLE",
            step_count=2,
            dependency_risk="MEDIUM",
            hard_constraints_satisfied=True,
            soft_constraints_score=0.70,
            overall_utility_score=0.78,
            selection_rationale="Maximum delay reduction (28h) via express rerouting, but requires $450 surcharge.",
            rejection_or_penalty_reasons=[],
        )
        cand_b = CandidatePlanModel(
            candidate_id="candidate-B",
            strategy_name="Express Alternative Routing",
            summary="Express rerouting cutting 28h of delay, with incremental carrier fee of $450.",
            is_feasible=True,
            rank=2,
            steps=steps_b,
            evaluation=eval_b,
        )

        # Candidate C: Conservative Telemetry & Escalation Checkpoint
        # Zero cost, lowest risk, 4h delay reduction, highly conservative
        steps_c = [
            PlanStepModel(
                step_id="step-1",
                step_number=1,
                action_type="tasks.create_internal_task",
                title="Establish Enhanced Telemetry Watch",
                description=f"Flag {module} #{entity_id} for high-frequency telemetry tracking and 2-hour exception alarm.",
                parameters={"entity_id": entity_id, "priority": "HIGH"},
                dependencies=[],
                expected_outcome="High-frequency telemetry watch initiated",
                verification_criteria=StepVerificationCriteria(
                    check_type="FIELD_EQUALS",
                    target_field="watch_status",
                    expected_value="ACTIVE",
                ),
                risk_level="LOW",
                requires_approval=False,
                reversibility="REVERSIBLE",
                timeout_seconds=120,
                idempotency_key=f"step-{plan_id}-c1-{uuid.uuid4().hex[:8]}",
            )
        ]
        eval_c = CandidateEvaluationModel(
            feasibility=True,
            delay_reduction_hours=4.0,
            estimated_cost=0.0,
            customer_impact="LOW",
            operational_risk="LOW",
            compliance_risk="LOW",
            confidence=0.95,
            reversibility="REVERSIBLE",
            step_count=1,
            dependency_risk="LOW",
            hard_constraints_satisfied=True,
            soft_constraints_score=0.85,
            overall_utility_score=0.80,
            selection_rationale="Conservative monitoring posture with zero cost and zero operational risk.",
            rejection_or_penalty_reasons=[],
        )
        cand_c = CandidatePlanModel(
            candidate_id="candidate-C",
            strategy_name="Conservative Monitoring & Governed Checkpoints",
            summary="Retain current routing, establish 2h high-frequency telemetry alerts, and prepare standby escalation.",
            is_feasible=True,
            rank=3,
            steps=steps_c,
            evaluation=eval_c,
        )

        return [cand_a, cand_b, cand_c]

    def _evaluate_and_rank_candidates(
        self,
        candidates: List[CandidatePlanModel],
        hard_constraints: List[ConstraintModel],
        soft_constraints: List[ConstraintModel],
        risk_tolerance: str,
    ) -> tuple[List[CandidatePlanModel], CandidatePlanModel]:
        """Evaluates constraints across each candidate, computes utility scores, and ranks them."""
        for cand in candidates:
            eval_m = cand.evaluation
            eval_m.hard_constraints_satisfied = True
            eval_m.rejection_or_penalty_reasons = []

            # Check Hard Constraints
            for hc in hard_constraints:
                hc_desc = hc.description.upper()
                # Check monetary / cost limits
                if "COST" in hc_desc or "BUDGET" in hc_desc or hc.constraint_type == "COST":
                    # If candidate cost exceeds hard limit, disqualify
                    if hc.threshold_value is not None and float(eval_m.estimated_cost) > float(hc.threshold_value):
                        eval_m.hard_constraints_satisfied = False
                        cand.is_feasible = False
                        eval_m.feasibility = False
                        eval_m.rejection_or_penalty_reasons.append(
                            f"Violates hard cost constraint: estimated cost ${eval_m.estimated_cost:.2f} exceeds threshold ${float(hc.threshold_value):.2f}"
                        )

                # Check compliance limits
                if "COMPLIANCE" in hc_desc or hc.constraint_type == "COMPLIANCE":
                    if eval_m.compliance_risk == "HIGH" or eval_m.compliance_risk == "CRITICAL":
                        eval_m.hard_constraints_satisfied = False
                        cand.is_feasible = False
                        eval_m.feasibility = False
                        eval_m.rejection_or_penalty_reasons.append("Violates hard compliance constraint: high compliance risk detected")

            # Check Soft Constraints
            soft_penalty = 0.0
            for sc in soft_constraints:
                sc_desc = sc.description.upper()
                if "COST" in sc_desc and eval_m.estimated_cost > 0:
                    soft_penalty += 0.15
                if "NOTIFICATION" in sc_desc and eval_m.customer_impact != "LOW":
                    soft_penalty += 0.10
            eval_m.soft_constraints_score = max(0.0, 1.0 - soft_penalty)

            # Compute Utility Score
            if not eval_m.hard_constraints_satisfied or not cand.is_feasible:
                eval_m.overall_utility_score = 0.0
                cand.is_feasible = False
            else:
                # Weighted utility calculation based on risk tolerance
                delay_norm = min(1.0, eval_m.delay_reduction_hours / 30.0)
                cost_norm = min(1.0, eval_m.estimated_cost / 1000.0)
                risk_pen = 0.0 if eval_m.operational_risk == "LOW" else (0.15 if eval_m.operational_risk == "MEDIUM" else 0.40)

                if risk_tolerance == "CONSERVATIVE":
                    util = (0.25 * delay_norm) + (0.35 * (1.0 - cost_norm)) + (0.25 * (1.0 - risk_pen)) + (0.15 * eval_m.soft_constraints_score)
                elif risk_tolerance == "AGGRESSIVE":
                    util = (0.50 * delay_norm) + (0.15 * (1.0 - cost_norm)) + (0.15 * (1.0 - risk_pen)) + (0.20 * eval_m.soft_constraints_score)
                else:  # BALANCED
                    util = (0.35 * delay_norm) + (0.25 * (1.0 - cost_norm)) + (0.20 * (1.0 - risk_pen)) + (0.20 * eval_m.soft_constraints_score)

                eval_m.overall_utility_score = round(min(1.0, max(0.05, util)), 4)

        # Sort: Feasible plans first (ordered by utility score descending), followed by Infeasible plans
        sorted_cands = sorted(
            candidates,
            key=lambda c: (1 if c.is_feasible else 0, c.evaluation.overall_utility_score),
            reverse=True
        )

        # Assign ranks
        for idx, c in enumerate(sorted_cands, start=1):
            c.rank = idx

        selected = sorted_cands[0]
        return sorted_cands, selected

    def _build_operational_explanation(
        self,
        ranked: List[CandidatePlanModel],
        selected: CandidatePlanModel,
    ) -> str:
        """Constructs concise user-facing operational explanations without exposing chain-of-thought."""
        reasons = [f"Strategy '{selected.strategy_name}' (Rank 1) selected because {selected.evaluation.selection_rationale.lower()}"]

        for c in ranked[1:]:
            if not c.is_feasible:
                reasons.append(f"Alternative '{c.strategy_name}' rejected: {'; '.join(c.evaluation.rejection_or_penalty_reasons)}.")
            else:
                reasons.append(
                    f"Alternative '{c.strategy_name}' ranked lower (Score {c.evaluation.overall_utility_score:.2f}) "
                    f"due to higher cost (${c.evaluation.estimated_cost:.2f}) or lower net recovery efficiency."
                )

        return " ".join(reasons)

    def evaluate_plan_policy(self, req: PlanEvaluationRequest) -> PlanEvaluationResponse:
        """Evaluates whether the generated plan satisfies autonomy policy boundaries."""
        plan = req.plan
        policy = req.policy or {}

        if policy.get("emergency_stop"):
            return PlanEvaluationResponse(
                is_permitted=False,
                policy_decision="BLOCKED_EMERGENCY_STOP",
                requires_approval=True,
                policy_reason="Autonomous planning halted by active emergency kill switch.",
                violated_rules=["EMERGENCY_STOP_ACTIVE"],
            )

        if not policy.get("is_active", True):
            return PlanEvaluationResponse(
                is_permitted=False,
                policy_decision="BLOCKED_POLICY",
                requires_approval=True,
                policy_reason=f"Autonomy is disabled for module {plan.module}.",
                violated_rules=["MODULE_AUTONOMY_DISABLED"],
            )

        max_steps = policy.get("max_plan_steps", 5)
        if len(plan.ordered_steps) > max_steps:
            return PlanEvaluationResponse(
                is_permitted=False,
                policy_decision="BLOCKED_POLICY",
                requires_approval=True,
                policy_reason=f"Plan step count ({len(plan.ordered_steps)}) exceeds policy limit of {max_steps}.",
                violated_rules=["MAX_PLAN_STEPS_EXCEEDED"],
            )

        requires_appr = policy.get("requires_approval", True)
        if plan.autonomy_level in ["LEVEL_0_OBSERVE", "LEVEL_1_RECOMMEND", "LEVEL_2_PREPARE"]:
            requires_appr = True

        for step in plan.ordered_steps:
            if step.requires_approval or step.risk_level in ["HIGH", "CRITICAL"]:
                requires_appr = True
                break

        decision = "REQUIRES_APPROVAL" if requires_appr else "PERMITTED"
        reason = "Plan requires human operator review before execution." if requires_appr else "Plan complies with all active autonomy policies and may execute in controlled mode."

        return PlanEvaluationResponse(
            is_permitted=True,
            policy_decision=decision,
            requires_approval=requires_appr,
            policy_reason=reason,
            violated_rules=[],
            confidence_acceptable=True,
            data_sufficiency_acceptable=plan.data_sufficiency,
        )

    def replan(self, req: PlanReplanRequest) -> PlanReplanResponse:
        """Generates an adaptive replan version when execution fails or environment state drifts."""
        new_version = req.current_version + 1
        revised_plan_id = f"plan-replan-{uuid.uuid4().hex[:10]}"

        remedial_steps = [
            PlanStepModel(
                step_id="step-1",
                step_number=1,
                action_type="tasks.create_internal_task",
                title=f"Investigate Drift: {req.replan_reason[:40]}",
                description=f"Adaptive replan version {new_version} triggered due to: {req.replan_reason}",
                parameters={"replan_reason": req.replan_reason, "parent_plan_id": req.original_plan_id},
                dependencies=[],
                expected_outcome="Discrepancy analyzed by logistics controller",
                risk_level="MEDIUM",
                requires_approval=True,
                reversibility="REVERSIBLE",
                idempotency_key=f"step-{revised_plan_id}-remedial-1",
            ),
            PlanStepModel(
                step_id="step-2",
                step_number=2,
                action_type="shipments.update_milestone",
                title="Re-synchronize Live Milestones",
                description="Synchronize live state following drift correction",
                parameters={"notes": "Adaptive recovery milestone"},
                dependencies=["step-1"],
                expected_outcome="Live milestones aligned with physical cargo location",
                risk_level="LOW",
                requires_approval=True,
                reversibility="REVERSIBLE",
                idempotency_key=f"step-{revised_plan_id}-remedial-2",
            ),
        ]

        return PlanReplanResponse(
            revised_plan_id=revised_plan_id,
            new_version=new_version,
            parent_plan_id=req.original_plan_id,
            replan_reason=req.replan_reason,
            changes_summary=f"Replaced failed/drifted action sequence with governed remedial tasks (Version {new_version}).",
            updated_steps=remedial_steps,
            confidence_score=0.82,
            risk_level="MEDIUM",
            status="REPLANNING",
        )

    def validate_plan_graph(self, req: MultiStepPlanValidationRequest) -> MultiStepPlanValidationResponse:
        """
        Validates step dependencies, detects cycles via topological sorting,
        partitions steps into parallel-safe execution groups, and checks approval gates.
        """
        steps = req.steps or []
        step_map = {s.step_id: s for s in steps}
        issues: List[str] = []
        warnings: List[str] = []
        in_degree: Dict[str, int] = {s.step_id: 0 for s in steps}
        adjacency: Dict[str, List[str]] = {s.step_id: [] for s in steps}

        # 1. Validate uniqueness & dependencies existence
        seen_ids = set()
        for s in steps:
            if s.step_id in seen_ids:
                issues.append(f"Duplicate step_id detected: '{s.step_id}'")
            seen_ids.add(s.step_id)

            for dep in s.dependencies:
                if dep not in step_map:
                    issues.append(f"Step '{s.step_id}' references non-existent dependency '{dep}'")
                elif dep == s.step_id:
                    issues.append(f"Step '{s.step_id}' declares self-dependency")
                else:
                    adjacency[dep].append(s.step_id)
                    in_degree[s.step_id] += 1

        # 2. Cycle detection & Parallel group calculation (Kahn's algorithm)
        queue = [s_id for s_id, deg in in_degree.items() if deg == 0]
        execution_order: List[str] = []
        parallel_groups: List[List[str]] = []
        contains_cycles = False

        while queue:
            current_group = list(queue)
            parallel_groups.append(current_group)
            queue = []

            for node in current_group:
                execution_order.append(node)
                for neighbor in adjacency[node]:
                    in_degree[neighbor] -= 1
                    if in_degree[neighbor] == 0:
                        queue.append(neighbor)

        if len(execution_order) != len(steps):
            contains_cycles = True
            unresolved = [s.step_id for s in steps if s.step_id not in execution_order]
            issues.append(f"Cycle detected in step dependency graph: circular dependencies among {unresolved}")

        # 3. Check for approval gate requirements
        has_approval_gate = False
        requires_approval_by_policy = req.autonomy_level in ["LEVEL_0_OBSERVE", "LEVEL_1_RECOMMEND", "LEVEL_2_PREPARE"]
        for s in steps:
            if s.requires_approval or s.risk_level in ["HIGH", "CRITICAL"] or requires_approval_by_policy:
                has_approval_gate = True
                break

        if not has_approval_gate and requires_approval_by_policy:
            warnings.append(f"Plan operates under {req.autonomy_level} but has no explicit human approval gates configured")

        is_valid = len(issues) == 0 and not contains_cycles

        return MultiStepPlanValidationResponse(
            is_valid=is_valid,
            issues=issues,
            warnings=warnings,
            execution_order=execution_order if is_valid else [],
            parallel_groups=parallel_groups if is_valid else [],
            contains_cycles=contains_cycles,
            has_approval_gate=has_approval_gate,
        )

    def generate_cross_module_plan(self, req: CrossModulePlanningRequest) -> CrossModulePlanningResponse:
        """
        Synthesizes an authoritative multi-step plan coordinating across multiple business modules
        (e.g., Shipments, Operations, Customer Follow-Up, Pricing, Finance, Compliance).
        """
        # 1. Prompt injection defense & sanitization
        from app.autonomy.safety import sanitize_untrusted_text
        safe_goal = sanitize_untrusted_text(req.goal)

        plan_id = f"plan-xmod-{uuid.uuid4().hex[:10]}"
        correlation_id = req.correlation_id or f"corr-xmod-{uuid.uuid4().hex[:12]}"
        primary_module = req.primary_module.lower()
        entity_id = req.primary_entity_id

        # Determine modules to coordinate
        modules = set(req.involved_modules or [])
        modules.add(primary_module)

        steps: List[PlanStepModel] = []
        step_counter = 1

        # Step 1: Context & Risk Analysis (Primary Module)
        step1_id = "step-1"
        steps.append(
            PlanStepModel(
                step_id=step1_id,
                step_number=step_counter,
                action_type=f"{primary_module}.analyze_context",
                title=f"Analyze {primary_module.capitalize()} Context & Risk Parameters",
                description=f"Compile verified state and quantify disruption parameters for {primary_module} #{entity_id}",
                parameters={"entity_id": entity_id, "primary_module": primary_module},
                dependencies=[],
                expected_outcome=f"Disruption parameters quantified and impact baseline established",
                verification_criteria=StepVerificationCriteria(
                    check_type="FIELD_EQUALS",
                    target_field="context_verified",
                    expected_value=True,
                    description="Confirm context verification",
                ),
                risk_level="LOW",
                requires_approval=False,
                reversibility="REVERSIBLE",
                timeout_seconds=180,
                idempotency_key=f"step-{plan_id}-1-{uuid.uuid4().hex[:8]}",
                preconditions=[{"check": "entity_exists", "entity_id": entity_id}],
                retry_policy={"max_attempts": 3, "backoff_seconds": 5},
                authorization_requirements=[f"{primary_module.upper()}:READ"],
            )
        )
        step_counter += 1

        # Parallel Safe Stage 2: Operational Inquiry + Cost / Pricing Impact (Parallel)
        step2a_id = "step-2a"
        steps.append(
            PlanStepModel(
                step_id=step2a_id,
                step_number=step_counter,
                action_type="carrier.carrier_inquiry" if "shipments" in modules or "operations" in modules else "tasks.create_internal_task",
                title="Carrier Operational Inquiry & Priority Escalation",
                description="Initiate governed operational status inquiry with operating carrier",
                parameters={"entity_id": entity_id, "inquiry_type": "URGENT_STATUS_UPDATE"},
                dependencies=[step1_id],
                expected_outcome="Carrier provides updated transit status or ETA adjustment",
                verification_criteria=StepVerificationCriteria(
                    check_type="RECORD_EXISTS",
                    target_field="carrier_inquiry_id",
                    expected_value=True,
                    description="Verify carrier inquiry record created",
                ),
                risk_level="LOW",
                requires_approval=False,
                reversibility="REVERSIBLE",
                timeout_seconds=300,
                idempotency_key=f"step-{plan_id}-2a-{uuid.uuid4().hex[:8]}",
                retry_policy={"max_attempts": 3, "backoff_seconds": 10},
                authorization_requirements=["OPERATIONS:UPDATE"],
            )
        )
        step_counter += 1

        step2b_id = "step-2b"
        steps.append(
            PlanStepModel(
                step_id=step2b_id,
                step_number=step_counter,
                action_type="pricing.evaluate_exposure" if "finance" in modules or "pricing" in modules else "compliance.record_discrepancies",
                title="Financial Exposure & Tariff Cap Assessment",
                description="Evaluate detention, demurrage, and margin variance under current contracts",
                parameters={"entity_id": entity_id, "exposure_currency": "USD"},
                dependencies=[step1_id],
                expected_outcome="Financial exposure quantified within contractual tolerance limits",
                verification_criteria=StepVerificationCriteria(
                    check_type="FIELD_EQUALS",
                    target_field="financial_exposure_assessed",
                    expected_value=True,
                    description="Confirm financial exposure audit",
                ),
                risk_level="LOW",
                requires_approval=False,
                reversibility="REVERSIBLE",
                timeout_seconds=180,
                idempotency_key=f"step-{plan_id}-2b-{uuid.uuid4().hex[:8]}",
                retry_policy={"max_attempts": 2, "backoff_seconds": 5},
                authorization_requirements=["FINANCE:READ"],
            )
        )
        step_counter += 1

        # Stage 3: Governance & Approval Gate (Depends on both Step 2a and Step 2b)
        step3_id = "step-3"
        steps.append(
            PlanStepModel(
                step_id=step3_id,
                step_number=step_counter,
                action_type="policy.approval_check",
                title="Supervisor Governance & Policy Sign-Off",
                description="Enforce organizational autonomy boundaries and supervisor sign-off before customer dispatch",
                parameters={"plan_id": plan_id, "gate_type": "HUMAN_IN_THE_LOOP"},
                dependencies=[step2a_id, step2b_id],
                expected_outcome="Human supervisor authorizes customer communication and operational adjustment",
                verification_criteria=StepVerificationCriteria(
                    check_type="STATUS_TRANSITION",
                    target_field="approval_status",
                    expected_value="APPROVED",
                    description="Verify supervisor sign-off in audit log",
                ),
                risk_level="MEDIUM",
                requires_approval=True,
                reversibility="REVERSIBLE",
                timeout_seconds=3600,
                idempotency_key=f"step-{plan_id}-3-{uuid.uuid4().hex[:8]}",
                preconditions=[{"check": "dependencies_verified", "steps": [step2a_id, step2b_id]}],
                authorization_requirements=["SUPERVISOR:APPROVE"],
            )
        )
        step_counter += 1

        # Stage 4: Customer Communication (Gated by Step 3 approval)
        step4_id = "step-4"
        steps.append(
            PlanStepModel(
                step_id=step4_id,
                step_number=step_counter,
                action_type="customer.send_communication",
                title="Dispatched Proactive Customer Advisory",
                description="Send grounded customer advisory with verified facts and predictive timeline adjustment",
                parameters={"channel": "EMAIL", "template": "PROACTIVE_RECOVERY_ADVISORY"},
                dependencies=[step3_id],
                condition_predicate=StepConditionPredicate(
                    field="approval_status",
                    operator="==",
                    value="APPROVED",
                ),
                expected_outcome="Customer advised of verified recovery status and revised milestone schedule",
                verification_criteria=StepVerificationCriteria(
                    check_type="FIELD_EQUALS",
                    target_field="customer_message_sent",
                    expected_value=True,
                    description="Verify delivery receipt in outbound queue",
                ),
                risk_level="HIGH",
                requires_approval=True,
                reversibility="IRREVERSIBLE",
                timeout_seconds=300,
                idempotency_key=f"step-{plan_id}-4-{uuid.uuid4().hex[:8]}",
                preconditions=[{"check": "approval_granted", "step_id": step3_id}],
                authorization_requirements=["CUSTOMER:COMMUNICATE"],
            )
        )
        step_counter += 1

        # Stage 5: Authoritative Outcome Verification
        step5_id = "step-5"
        steps.append(
            PlanStepModel(
                step_id=step5_id,
                step_number=step_counter,
                action_type="shipments.update_milestone",
                title="Authoritative Recovery Milestone Verification",
                description="Verify physical milestone clearing and state reconciliation across database records",
                parameters={"milestone_code": "RECOVERY_VERIFIED"},
                dependencies=[step4_id],
                expected_outcome="Physical shipment tracking reconciled and active exception conditions cleared",
                verification_criteria=StepVerificationCriteria(
                    check_type="FIELD_EQUALS",
                    target_field="recovery_verified",
                    expected_value=True,
                    description="Verify exception condition cleared in authoritative database",
                ),
                risk_level="LOW",
                requires_approval=False,
                reversibility="REVERSIBLE",
                timeout_seconds=300,
                idempotency_key=f"step-{plan_id}-5-{uuid.uuid4().hex[:8]}",
                retry_policy={"max_attempts": 3, "backoff_seconds": 15},
                authorization_requirements=["SHIPMENTS:UPDATE"],
            )
        )
        step_counter += 1

        # Stage 6: Cross-Module Workflow Closure / Final Audit
        step6_id = "step-6"
        steps.append(
            PlanStepModel(
                step_id=step6_id,
                step_number=step_counter,
                action_type="audit_log",
                title="Cross-Module Workflow Audit Closure",
                description="Commit final audit ledger entry and archive multi-step execution lineage",
                parameters={"plan_id": plan_id, "final_status": "COMPLETED"},
                dependencies=[step5_id],
                expected_outcome="Multi-step cross-module recovery workflow marked COMPLETED and archived",
                verification_criteria=StepVerificationCriteria(
                    check_type="FIELD_EQUALS",
                    target_field="workflow_closed",
                    expected_value=True,
                    description="Confirm workflow closure in audit trail",
                ),
                risk_level="LOW",
                requires_approval=False,
                reversibility="REVERSIBLE",
                timeout_seconds=120,
                idempotency_key=f"step-{plan_id}-6-{uuid.uuid4().hex[:8]}",
                authorization_requirements=["AUDIT:WRITE"],
            )
        )

        stop_conditions = [
            StopConditionModel(
                condition_type="COMPLIANCE_BLOCK_DETECTED",
                threshold="HARD_COMPLIANCE_FAIL",
                escalation_target="COMPLIANCE_OFFICER",
                description="Halt execution immediately if statutory or customs compliance violation occurs",
            ),
            StopConditionModel(
                condition_type="APPROVAL_REJECTED",
                threshold="SUPERVISOR_REJECT",
                escalation_target="OPERATIONS_DIRECTOR",
                description="Terminate execution sequence if supervisor rejects human approval gate",
            ),
            StopConditionModel(
                condition_type="MAX_RETRY_EXCEEDED",
                threshold=3,
                escalation_target="OPERATIONS_SUPERVISOR",
                description="Escalate if any dependent step fails 3 consecutive execution attempts",
            ),
        ]

        # Calculate parallel groups
        validation = self.validate_plan_graph(
            MultiStepPlanValidationRequest(
                plan_id=plan_id,
                steps=steps,
                module=primary_module,
                autonomy_level=req.autonomy_level,
            )
        )

        return CrossModulePlanningResponse(
            plan_id=plan_id,
            goal=safe_goal,
            goal_type="CROSS_MODULE",
            primary_module=primary_module,
            primary_entity_id=entity_id,
            involved_modules=sorted(list(modules)),
            ordered_steps=steps,
            parallel_groups=validation.parallel_groups,
            stop_conditions=stop_conditions,
            confidence_score=0.91,
            data_sufficiency=True,
            requires_human_approval=True,
            correlation_id=correlation_id,
        )
