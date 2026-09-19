"""
LogisticsHQ Phase 5 Task 5.3: Adaptive Shipment Management Agent
Handles continuous shipment state reasoning, event evaluation, meaningful change detection,
customer commitment protection, waiting states, and adaptive multi-candidate plan synthesis.
"""

from typing import List, Optional, Dict, Any, Tuple
import datetime
import uuid
from app.autonomy.models import (
    PlanGenerationRequest,
    PlanGenerationResponse,
    PlanStepModel,
    CandidatePlanModel,
    CandidateEvaluationModel,
    ConstraintModel,
    StepConditionPredicate,
    StepVerificationCriteria,
    StopConditionModel,
    ShipmentEventEvaluationRequest,
    ShipmentEventEvaluationResponse,
    RiskLevel,
    WaitingStateType,
    EventDecisionType,
)
from app.autonomy.safety import sanitize_plan_request, validate_steps_safety


class AdaptiveShipmentAgent:
    """Agent responsible for intelligent, adaptive operational reasoning on shipments."""

    def evaluate_event(self, req: ShipmentEventEvaluationRequest) -> ShipmentEventEvaluationResponse:
        """Evaluates incoming shipment events to determine whether they represent operationally meaningful changes."""
        curr = req.current_state or {}
        event_type = (req.event_type or "").upper()
        severity = req.severity or "MEDIUM"

        # 1. Parse ETA and Customer Commitment Date
        curr_eta_str = curr.get("eta") or curr.get("estimated_arrival")
        pred_eta_str = req.predicted_eta or curr.get("predicted_eta") or curr_eta_str
        commitment_str = req.customer_commitment_date or curr.get("customer_commitment_date")

        eta_deviation_hours = 0.0
        commitment_breach = False
        commitment_risk_severity: RiskLevel = "LOW"

        curr_eta_dt = self._parse_iso_date(curr_eta_str)
        pred_eta_dt = self._parse_iso_date(pred_eta_str)
        commitment_dt = self._parse_iso_date(commitment_str)

        if curr_eta_dt and pred_eta_dt:
            deviation_sec = (pred_eta_dt - curr_eta_dt).total_seconds()
            eta_deviation_hours = round(deviation_sec / 3600.0, 2)

        if commitment_dt and pred_eta_dt:
            if pred_eta_dt > commitment_dt:
                commitment_breach = True
                overdue_sec = (pred_eta_dt - commitment_dt).total_seconds()
                overdue_hours = overdue_sec / 3600.0
                if overdue_hours > 24:
                    commitment_risk_severity = "CRITICAL"
                elif overdue_hours > 6:
                    commitment_risk_severity = "HIGH"
                else:
                    commitment_risk_severity = "MEDIUM"

        # 2. Check for active plan and its status
        active_plan = req.active_plan or {}
        active_plan_id = active_plan.get("plan_id")
        active_plan_status = active_plan.get("status")

        # 3. Operational decision logic
        is_meaningful = True
        decision: EventDecisionType = "CONTINUE_MONITORING"
        decision_reason = ""
        recommended_action = None
        escalation_reason = None
        requires_new_plan = False
        requires_replan = False
        waiting_state = None

        # Check payload delay_hours if eta_deviation_hours is 0
        raw_payload = req.raw_event_payload or {}
        payload_delay = float(raw_payload.get("delay_hours", 0.0))
        if payload_delay != 0.0 and eta_deviation_hours == 0.0:
            eta_deviation_hours = payload_delay
            if eta_deviation_hours > 6.0 and not commitment_breach:
                commitment_risk_severity = "HIGH"

        # Check for compliance blockage (Hard boundary)
        exceptions = curr.get("unresolved_exceptions") or []
        has_customs_blockage = any(
            "CUSTOMS" in str(e.get("exception_type", "")).upper() or "EMBARGO" in str(e.get("title", "")).upper()
            for e in exceptions if isinstance(e, dict)
        )

        if has_customs_blockage or "CUSTOMS" in event_type or "EMBARGO" in event_type or "COMPLIANCE" in event_type:
            decision = "ESCALATION"
            is_meaningful = True
            escalation_reason = "Regulatory or customs hold detected. Requires human customs clearance specialist."
            decision_reason = "Compliance boundary prohibits autonomous clearance. Immediate human escalation dispatched."
            recommended_action = "approvals.request"
            waiting_state = "WAITING_FOR_APPROVAL"

        elif "ETA" in event_type or "PORT" in event_type or "DISRUPTION" in event_type or "DELAY" in event_type or "ROUTING" in event_type:
            if abs(eta_deviation_hours) < 2.0 and not commitment_breach and severity == "LOW":
                # Insignificant fluctuation
                is_meaningful = False
                decision = "CONTINUE_MONITORING"
                decision_reason = f"ETA deviation of {eta_deviation_hours:+.1f}h is within acceptable operating tolerance (< 2.0h). Continued telemetry monitoring."
            elif commitment_breach or abs(eta_deviation_hours) >= 6.0 or severity in ["HIGH", "CRITICAL"] or "DISRUPTION" in event_type:
                is_meaningful = True
                decision_reason = f"Operational disruption / delay of {eta_deviation_hours:+.1f}h detected ({commitment_risk_severity} commitment risk). Formulating adaptive recovery plan."
                if active_plan_id and active_plan_status in ["GENERATED", "APPROVED", "EXECUTING"]:
                    decision = "REPLANNING"
                    requires_replan = True
                else:
                    decision = "NEW_PLAN"
                    requires_new_plan = True
                recommended_action = "shipments.request_carrier_update"
                waiting_state = "WAITING_FOR_CARRIER"
            else:
                is_meaningful = True
                decision_reason = f"Material ETA slippage of {eta_deviation_hours:+.1f}h detected. Proactive recovery strategy formulated."
                decision = "NEW_PLAN"
                requires_new_plan = True
                recommended_action = "shipments.update_eta"

        elif event_type in ["MISSED_MILESTONE", "MILESTONE_DELAY"]:
            is_meaningful = True
            decision_reason = f"Shipment missed tracking milestone cutoff. Operational delay recovery required."
            if active_plan_id and active_plan_status in ["APPROVED", "EXECUTING"]:
                decision = "REPLANNING"
                requires_replan = True
            else:
                decision = "NEW_PLAN"
                requires_new_plan = True
            recommended_action = "shipments.request_carrier_update"
            waiting_state = "WAITING_FOR_CARRIER"

        elif event_type in ["EXCEPTION_LOGGED", "CRITICAL_EXCEPTION"]:
            is_meaningful = True
            decision_reason = f"Operational exception logged with severity '{severity}'. Formulating multi-step recovery options."
            decision = "NEW_PLAN"
            requires_new_plan = True
            recommended_action = "shipments.create_exception"

        elif event_type == "CARRIER_UPDATE":
            if commitment_breach or abs(eta_deviation_hours) >= 6.0 or severity in ["HIGH", "CRITICAL"]:
                is_meaningful = True
                decision_reason = f"Carrier update reported operational disruption with delay of {eta_deviation_hours:.1f}h. Formulating adaptive recovery plan."
                if active_plan_id:
                    decision = "REPLANNING"
                    requires_replan = True
                else:
                    decision = "NEW_PLAN"
                    requires_new_plan = True
                waiting_state = "WAITING_FOR_CARRIER"
            elif active_plan.get("waiting_state") == "WAITING_FOR_CARRIER":
                is_meaningful = True
                decision = "CONTROLLED_EXECUTION"
                decision_reason = "Carrier response received. Transitioning from waiting state to next planned action."
            else:
                is_meaningful = False
                decision = "CONTINUE_MONITORING"
                decision_reason = "Routine carrier telemetry update received. Current trajectory remains valid."

        elif event_type == "PLAN_STALE":
            is_meaningful = True
            decision = "REPLANNING"
            requires_replan = True
            decision_reason = "Operational plan has aged beyond freshness threshold. Revalidating context against current carrier status."

        else:
            is_meaningful = False
            decision = "CONTINUE_MONITORING"
            decision_reason = f"Event '{event_type}' noted. Shipment remains within operating parameters."

        return ShipmentEventEvaluationResponse(
            shipment_id=req.shipment_id,
            event_type=req.event_type,
            decision=decision,
            decision_reason=decision_reason,
            is_meaningful_change=is_meaningful,
            eta_deviation_hours=eta_deviation_hours,
            commitment_risk_severity=commitment_risk_severity,
            recommended_action_type=recommended_action,
            escalation_reason=escalation_reason,
            requires_new_plan=requires_new_plan,
            requires_replan=requires_replan,
            active_plan_id=active_plan_id,
            waiting_state=waiting_state,
            correlation_id=req.correlation_id,
        )

    def generate_adaptive_plan(self, req: PlanGenerationRequest) -> PlanGenerationResponse:
        """Synthesizes structured adaptive shipment recovery options, evaluates constraints, and produces a controlled plan."""
        sanitize_plan_request(req)

        plan_id = f"plan-ship-{uuid.uuid4().hex[:10]}"
        curr = req.current_state or {}
        shipment_id = req.related_entity_id

        # Parse commitments and ETA
        curr_eta_str = curr.get("eta") or curr.get("estimated_arrival")
        pred_eta_str = req.predictions_context.get("predicted_eta") or curr.get("predicted_eta") or curr_eta_str
        commitment_str = curr.get("customer_commitment_date")

        curr_eta_dt = self._parse_iso_date(curr_eta_str)
        pred_eta_dt = self._parse_iso_date(pred_eta_str)
        commitment_dt = self._parse_iso_date(commitment_str)

        eta_deviation_hours = 0.0
        if curr_eta_dt and pred_eta_dt:
            eta_deviation_hours = round((pred_eta_dt - curr_eta_dt).total_seconds() / 3600.0, 2)

        commitment_risk: RiskLevel = "LOW"
        if commitment_dt and pred_eta_dt and pred_eta_dt > commitment_dt:
            overdue_hours = (pred_eta_dt - commitment_dt).total_seconds() / 3600.0
            commitment_risk = "CRITICAL" if overdue_hours > 24 else "HIGH" if overdue_hours > 6 else "MEDIUM"

        hard_constraints = list(req.hard_constraints or [])
        soft_constraints = list(req.soft_constraints or [])

        # ── 1. Synthesize 4 Distinct Candidate Operational Strategies ─────────────
        candidates: List[CandidatePlanModel] = []

        # Candidate A: Accelerated Milestone Recovery (Direct Expedite)
        cand_a_steps = [
            PlanStepModel(
                step_id="step-1",
                step_number=1,
                action_type="shipments.get",
                title="Carrier Status & Port Telemetry Verification",
                description=f"Query fresh carrier telemetry for shipment #{shipment_id} to verify active vessel position.",
                parameters={"shipment_id": int(shipment_id) if str(shipment_id).isdigit() else 1},
                dependencies=[],
                condition_predicate=None,
                expected_outcome="Carrier API confirms current vessel position and ETA",
                verification_criteria=StepVerificationCriteria(
                    check_type="RECORD_EXISTS",
                    target_field="id",
                    expected_value=int(shipment_id) if str(shipment_id).isdigit() else 1,
                    description="Confirm shipment record retrieved",
                ),
                risk_level="LOW",
                requires_approval=False,
                reversibility="REVERSIBLE",
                timeout_seconds=120,
            ),
            PlanStepModel(
                step_id="step-2",
                step_number=2,
                action_type="shipments.update_milestone",
                title="Expedited Terminal Processing Push",
                description=f"Issue expedited terminal milestone priority for shipment #{shipment_id}.",
                parameters={"shipment_id": int(shipment_id) if str(shipment_id).isdigit() else 1, "milestone_code": "IN_TRANSIT"},
                dependencies=["step-1"],
                condition_predicate=None,
                expected_outcome="Milestone updated to IN_TRANSIT with expedited flag",
                verification_criteria=StepVerificationCriteria(
                    check_type="STATUS_TRANSITION",
                    target_field="status",
                    expected_value="IN_TRANSIT",
                    description="Verify milestone status transitioned to IN_TRANSIT",
                ),
                risk_level="MEDIUM",
                requires_approval=req.autonomy_level in ["LEVEL_0_OBSERVE", "LEVEL_1_RECOMMEND", "LEVEL_2_PREPARE"],
                reversibility="REVERSIBLE",
                timeout_seconds=300,
            ),
        ]
        candidates.append(
            CandidatePlanModel(
                candidate_id="candidate-A",
                strategy_name="Accelerated Carrier Milestone Push",
                summary="Direct carrier dispatch and expedited terminal processing. High recovery, low cost.",
                is_feasible=True,
                rank=1,
                steps=cand_a_steps,
                evaluation=CandidateEvaluationModel(
                    feasibility=True,
                    delay_reduction_hours=14.0,
                    estimated_cost=80.0,
                    customer_impact="LOW",
                    operational_risk="LOW",
                    compliance_risk="LOW",
                    confidence=0.88,
                    reversibility="REVERSIBLE",
                    step_count=len(cand_a_steps),
                    dependency_risk="LOW",
                    hard_constraints_satisfied=True,
                    soft_constraints_score=0.88,
                    overall_utility_score=0.85,
                    selection_rationale="Fastest turnaround with negligible incremental expense and full reversibility.",
                ),
            )
        )

        # Candidate B: Transshipment Hub Alternative Route
        cand_b_steps = [
            PlanStepModel(
                step_id="step-1",
                step_number=1,
                action_type="shipments.create_exception",
                title="Record Transshipment Reroute Advisory",
                description=f"Log operational schedule delay exception for shipment #{shipment_id}.",
                parameters={
                    "shipment_id": int(shipment_id) if str(shipment_id).isdigit() else 1,
                    "exception_type": "SCHEDULE_DELAY",
                    "severity": "HIGH",
                    "title": "Transshipment Congestion Reroute Required",
                    "description": "Port congestion at primary hub requires secondary bypass.",
                },
                dependencies=[],
                expected_outcome="Exception logged and visible on operational dashboard",
                verification_criteria=StepVerificationCriteria(
                    check_type="STATUS_TRANSITION",
                    target_field="status",
                    expected_value="OPEN",
                    description="Exception created with OPEN status",
                ),
                risk_level="MEDIUM",
                requires_approval=False,
                reversibility="REVERSIBLE",
                timeout_seconds=180,
            ),
            PlanStepModel(
                step_id="step-2",
                step_number=2,
                action_type="shipments.update_eta",
                title="Update Revised ETA via Alternate Hub",
                description=f"Record recalculated ETA reflecting express transshipment corridor for shipment #{shipment_id}.",
                parameters={
                    "shipment_id": int(shipment_id) if str(shipment_id).isdigit() else 1,
                    "estimated_arrival": pred_eta_str or "2026-09-20T12:00:00Z",
                    "notes": "Revised via adaptive express routing",
                },
                dependencies=["step-1"],
                expected_outcome="Shipment ETA record reflects expedited route schedule",
                verification_criteria=StepVerificationCriteria(
                    check_type="FIELD_EQUALS",
                    target_field="status",
                    expected_value="ETA_UPDATED",
                    description="Verify ETA update status",
                ),
                risk_level="HIGH",
                requires_approval=True,  # Always requires human approval due to cost
                reversibility="PARTIALLY_REVERSIBLE",
                timeout_seconds=300,
            ),
        ]
        candidates.append(
            CandidatePlanModel(
                candidate_id="candidate-B",
                strategy_name="Alternate Express Transshipment Route",
                summary="Bypass congested hub via alternate express feeder corridor. Maximum recovery, higher expense.",
                is_feasible=True,
                rank=2,
                steps=cand_b_steps,
                evaluation=CandidateEvaluationModel(
                    feasibility=True,
                    delay_reduction_hours=26.0,
                    estimated_cost=280.0,
                    customer_impact="LOW",
                    operational_risk="MEDIUM",
                    compliance_risk="LOW",
                    confidence=0.82,
                    reversibility="PARTIALLY_REVERSIBLE",
                    step_count=len(cand_b_steps),
                    dependency_risk="MEDIUM",
                    hard_constraints_satisfied=True,
                    soft_constraints_score=0.75,
                    overall_utility_score=0.78,
                    selection_rationale="Maximum delay reduction (26h) but incurs additional rerouting cost ($280).",
                ),
            )
        )

        # Candidate C: Customer SLA Communication & Buffer Realignment
        cand_c_steps = [
            PlanStepModel(
                step_id="step-1",
                step_number=1,
                action_type="shipments.update_eta",
                title="Calibrate Realized Milestone Delivery Window",
                description=f"Calibrate operational ETA for shipment #{shipment_id} to ensure billing and SLA alignment.",
                parameters={
                    "shipment_id": int(shipment_id) if str(shipment_id).isdigit() else 1,
                    "estimated_arrival": pred_eta_str or "2026-09-20T12:00:00Z",
                    "notes": "Calibrated based on terminal carrier report",
                },
                dependencies=[],
                expected_outcome="Shipment ETA aligned with carrier predictions",
                verification_criteria=StepVerificationCriteria(
                    check_type="FIELD_EQUALS",
                    target_field="status",
                    expected_value="ETA_UPDATED",
                    description="Verify ETA updated",
                ),
                risk_level="LOW",
                requires_approval=req.autonomy_level in ["LEVEL_0_OBSERVE", "LEVEL_1_RECOMMEND", "LEVEL_2_PREPARE"],
                reversibility="REVERSIBLE",
                timeout_seconds=180,
            ),
        ]
        candidates.append(
            CandidatePlanModel(
                candidate_id="candidate-C",
                strategy_name="Proactive Customer SLA Notice & Timeline Alignment",
                summary="Align customer delivery expectations proactively without route alterations. Zero incremental cost.",
                is_feasible=True,
                rank=3,
                steps=cand_c_steps,
                evaluation=CandidateEvaluationModel(
                    feasibility=True,
                    delay_reduction_hours=4.0,
                    estimated_cost=0.0,
                    customer_impact="MEDIUM",
                    operational_risk="LOW",
                    compliance_risk="LOW",
                    confidence=0.92,
                    reversibility="REVERSIBLE",
                    step_count=len(cand_c_steps),
                    dependency_risk="LOW",
                    hard_constraints_satisfied=True,
                    soft_constraints_score=0.85,
                    overall_utility_score=0.74,
                    selection_rationale="Zero incremental cost; mitigates client dissatisfaction through early transparent notification.",
                ),
            )
        )

        # ── 2. Evaluate Hard Constraints ──────────────────────────────────────────
        for cand in candidates:
            hard_sat = True
            for hc in hard_constraints:
                c_type = hc.constraint_type.upper()
                if c_type == "COST" and hc.threshold_value is not None:
                    if cand.evaluation.estimated_cost > float(hc.threshold_value):
                        hard_sat = False
                        cand.is_feasible = False
                        cand.evaluation.feasibility = False
                        cand.evaluation.hard_constraints_satisfied = False
                        cand.evaluation.overall_utility_score = 0.0
                        cand.evaluation.rejection_or_penalty_reasons.append(
                            f"Estimated cost ${cand.evaluation.estimated_cost:.2f} exceeds hard budget cap of ${float(hc.threshold_value):.2f}"
                        )
                elif c_type == "COMPLIANCE":
                    if cand.evaluation.compliance_risk == "CRITICAL":
                        hard_sat = False
                        cand.is_feasible = False
                        cand.evaluation.feasibility = False
                        cand.evaluation.hard_constraints_satisfied = False
                        cand.evaluation.overall_utility_score = 0.0
                        cand.evaluation.rejection_or_penalty_reasons.append("Violates compliance constraint")

        # ── 3. Rank Candidates ───────────────────────────────────────────────────
        feasible_cands = [c for c in candidates if c.is_feasible]
        infeasible_cands = [c for c in candidates if not c.is_feasible]

        feasible_cands.sort(key=lambda c: c.evaluation.overall_utility_score, reverse=True)
        ranked = feasible_cands + infeasible_cands
        for idx, c in enumerate(ranked):
            c.rank = idx + 1

        selected_cand = ranked[0] if ranked else candidates[0]

        # ── 4. Determine Initial Waiting State ─────────────────────────────────────
        waiting_state = None
        first_step = selected_cand.steps[0] if selected_cand.steps else None
        if first_step:
            if first_step.requires_approval:
                waiting_state = "WAITING_FOR_APPROVAL"
            elif "carrier" in first_step.action_type.lower():
                waiting_state = "WAITING_FOR_CARRIER"

        # ── 5. Stop Conditions ────────────────────────────────────────────────────
        stop_conditions = [
            StopConditionModel(
                condition_type="CONFIDENCE_DEGRADATION",
                threshold=0.60,
                escalation_target="OPERATIONS_SUPERVISOR",
                description="Halt execution if real-time tracking confidence drops below 60%",
            ),
            StopConditionModel(
                condition_type="MAX_STEPS_EXCEEDED",
                threshold=5,
                escalation_target="OPERATIONS_SUPERVISOR",
                description="Bound plan chain to 5 actions to prevent runaway execution",
            ),
            StopConditionModel(
                condition_type="UNEXPECTED_STATUS_CHANGE",
                threshold="CANCELLED_OR_DELIVERED",
                escalation_target="OPERATIONS_SUPERVISOR",
                description="Immediately halt recovery plan if shipment status changes out of sequence",
            ),
        ]

        explanation = (
            f"Evaluated {len(ranked)} adaptive shipment alternatives. Selected '{selected_cand.strategy_name}' (Rank 1). "
            f"Expected delay mitigation: {selected_cand.evaluation.delay_reduction_hours:.1f}h. "
            f"Incremental expense: ${selected_cand.evaluation.estimated_cost:.2f}. "
            f"Customer commitment risk: {commitment_risk}."
        )

        return PlanGenerationResponse(
            plan_id=plan_id,
            goal_id=getattr(req, "goal_id", None) or f"goal-ship-{uuid.uuid4().hex[:8]}",
            version=1,
            goal=req.goal or f"Adaptive recovery for shipment #{shipment_id}",
            module="shipments",
            related_entity_type="SHIPMENT",
            related_entity_id=str(shipment_id),
            current_state_summary=f"Shipment #{shipment_id} - ETA Deviation: {eta_deviation_hours:+.1f}h. Commitment Risk: {commitment_risk}.",
            constraints=req.constraints or [],
            hard_constraints=hard_constraints,
            soft_constraints=soft_constraints,
            assumptions=["Carrier API status accurate within 2h", "Alternate routing capacity available"],
            risks=[{"type": "COMMITMENT_BREACH", "severity": commitment_risk, "deviation_hours": eta_deviation_hours}],
            candidates=ranked,
            selected_candidate_id=selected_cand.candidate_id,
            ordered_steps=selected_cand.steps,
            evaluation_summary={
                "candidate_count": len(ranked),
                "selected_candidate": selected_cand.candidate_id,
                "strategy_name": selected_cand.strategy_name,
                "utility_score": selected_cand.evaluation.overall_utility_score,
                "delay_reduction_hours": selected_cand.evaluation.delay_reduction_hours,
                "estimated_cost": selected_cand.evaluation.estimated_cost,
                "hard_constraints_satisfied": selected_cand.evaluation.hard_constraints_satisfied,
                "commitment_risk": commitment_risk,
            },
            confidence_score=selected_cand.evaluation.confidence,
            data_sufficiency=True,
            estimated_impact=f"Estimated delay recovery: {selected_cand.evaluation.delay_reduction_hours:.1f}h. Incremental cost: ${selected_cand.evaluation.estimated_cost:.2f}.",
            risk_level=selected_cand.evaluation.operational_risk,
            autonomy_level=req.autonomy_level,
            stop_conditions=stop_conditions,
            staleness_status="FRESH",
            waiting_state=waiting_state,
            customer_commitment_date=commitment_str,
            predicted_eta=pred_eta_str,
            eta_deviation_hours=eta_deviation_hours,
            commitment_risk_severity=commitment_risk,
            status="GENERATED",
            correlation_id=req.correlation_id,
            explanation=explanation,
        )

    def _parse_iso_date(self, date_str: Optional[str]) -> Optional[datetime.datetime]:
        """Safely parses ISO / RFC3339 date strings into UTC datetime objects."""
        if not date_str or not isinstance(date_str, str):
            return None
        cleaned = date_str.replace("Z", "+00:00").strip()
        for fmt in [
            "%Y-%m-%dT%H:%M:%S%z",
            "%Y-%m-%d %H:%M:%S",
            "%Y-%m-%d",
        ]:
            try:
                dt = datetime.datetime.strptime(cleaned, fmt)
                if dt.tzinfo is None:
                    dt = dt.replace(tzinfo=datetime.timezone.utc)
                return dt
            except ValueError:
                continue
        try:
            return datetime.datetime.fromisoformat(cleaned)
        except Exception:
            return None
