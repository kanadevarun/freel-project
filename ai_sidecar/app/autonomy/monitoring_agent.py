"""
LogisticsHQ Phase 5 Task 5.10: Continuous Monitoring and Replanning Agent
Responsible for:
- Interpreting meaningful business state changes
- Validating whether active AI plans remain valid
- Identifying invalidated assumptions
- Detecting prediction drift
- Formulating revised plan versions while strictly protecting completed steps
- Prompt injection defense
"""

import uuid
import datetime
from typing import List, Dict, Any, Tuple

from app.autonomy.models import (
    StateChangeEvaluationRequest,
    StateChangeEvaluationResponse,
    ContinuousReplanningRequest,
    ContinuousReplanningResponse,
    PlanStepModel,
    PlanHealthState,
    RecommendedAction,
)
from app.autonomy.planner import AutonomousPlannerAgent
from app.autonomy.safety import detect_prompt_injection, sanitize_untrusted_text


class ContinuousMonitoringAgent:
    """
    Python AI reasoning agent for continuous operational state evaluation and adaptive replanning.
    Adheres strictly to the architectural invariant: reasoning & graph formulation only;
    zero direct database mutations or action dispatches.
    """

    def __init__(self):
        self.planner = AutonomousPlannerAgent()

    def evaluate_state_change(self, req: StateChangeEvaluationRequest) -> StateChangeEvaluationResponse:
        """
        Evaluates an authoritative incoming event against an active plan.
        Determines plan health, assumption validity, invalidated steps, and next recommended action.
        """
        evt = req.event
        corr_id = req.correlation_id or f"corr-mon-{uuid.uuid4().hex[:12]}"

        # Defensively sanitize any raw text payload from external entities
        raw_payload_desc = str(evt.payload.get("description", "") or evt.payload.get("message", "") or "")
        sanitized_desc = raw_payload_desc
        is_inj, _ = detect_prompt_injection(raw_payload_desc)
        if is_inj:
            sanitized_desc = sanitize_untrusted_text(raw_payload_desc)
            evt.payload["sanitized_description"] = sanitized_desc
            evt.payload["prompt_injection_flagged"] = True

        event_type = evt.event_type.upper()
        current_plan = req.current_plan or {}
        steps = req.steps or []

        plan_health: PlanHealthState = "HEALTHY"
        is_plan_valid: bool = True
        materiality_analysis: str = "Event assessed as non-material; existing plan remains valid and active."
        invalidated_step_ids: List[str] = []
        changed_assumptions: List[str] = []
        recommended_action: RecommendedAction = "CONTINUE"
        replan_rationale = None
        escalation_details = None

        # ---------------------------------------------------------------------
        # 1. SHIPMENT ETA SLIP & SCHEDULE DISRUPTIONS
        # ---------------------------------------------------------------------
        if "ETA" in event_type or "SLIP" in event_type or "DELAY" in event_type:
            delay_hours = float(evt.payload.get("eta_delay_hours", 0.0) or evt.payload.get("delay_hours", 0.0))
            commitment_breached = bool(evt.payload.get("commitment_breached", False))

            if delay_hours >= 3.0 or commitment_breached:
                plan_health = "AT_RISK" if delay_hours < 24.0 else "REPLANNING"
                is_plan_valid = False
                changed_assumptions = [
                    "vessel_schedule_on_time_guarantee",
                    "delivery_within_customer_sla_commitment",
                ]
                # Invalidate steps that depend on timely arrival or old milestone dates
                for s in steps:
                    if s.status not in ["COMPLETED", "SUCCEEDED"]:
                        act_lower = s.action_type.lower()
                        if any(kw in act_lower for kw in [
                            "eta", "delay", "carrier", "drayage", "dock", "delivery", "appointment", "schedule", "dispatch", "hold_invoice"
                        ]):
                            invalidated_step_ids.append(s.step_id)

                recommended_action = "REPLAN"
                materiality_analysis = (
                    f"Material schedule deviation detected: ETA slipped by {delay_hours:.1f}h. "
                    f"Customer commitment breach: {commitment_breached}. Active execution steps require recalibration."
                )
                replan_rationale = "Replan needed to coordinate carrier expedited handling and revised customer delivery commitments."

        # ---------------------------------------------------------------------
        # 2. CARRIER HOLD, CANCELLATION & REJECTION
        # ---------------------------------------------------------------------
        elif "CARRIER" in event_type and ("HOLD" in event_type or "REJECT" in event_type or "CANCEL" in event_type):
            plan_health = "BLOCKED"
            is_plan_valid = False
            changed_assumptions = [
                "carrier_space_confirmed",
                "carrier_operational_readiness",
                "vessel_feeder_connection_available",
            ]
            for s in steps:
                if s.status not in ["COMPLETED", "SUCCEEDED"]:
                    invalidated_step_ids.append(s.step_id)

            recommended_action = "REPLAN"
            materiality_analysis = (
                f"Carrier event '{event_type}' represents an operational blockage. "
                "Assumptions regarding carrier allocation have been invalidated."
            )
            replan_rationale = "Replan required to engage alternative carrier routing or request priority terminal unblocking."

        # ---------------------------------------------------------------------
        # 3. CUSTOMER OBJECTION & COMMERCIAL REJECTION
        # ---------------------------------------------------------------------
        elif "CUSTOMER" in event_type:
            is_cust_rejection = ("REJECT" in event_type or "OBJECTION" in event_type or "ESCALAT" in event_type) or (
                any(neg in raw_payload_desc.lower() for neg in ["cannot accept", "reject", "disagree", "unacceptable", "cancel", "refuse", "not accept"])
            )
            if is_cust_rejection:
                plan_health = "REPLANNING"
                is_plan_valid = False
                changed_assumptions = [
                    "customer_consents_to_proposed_timeline",
                    "customer_accepts_cost_allocation",
                ]
                for s in steps:
                    if "customer" in s.action_type.lower() and s.status not in ["COMPLETED", "SUCCEEDED"]:
                        invalidated_step_ids.append(s.step_id)

                recommended_action = "REPLAN"
                materiality_analysis = (
                    f"Customer response indicates objection or terms rejection: '{sanitized_desc[:120]}'. "
                    "Current customer communication assumptions are no longer valid."
                )
                replan_rationale = "Replan required to formulate tailored commercial concessions and revised milestone schedule."
            else:
                materiality_analysis = "Routine customer interaction received without commercial objection."

        # ---------------------------------------------------------------------
        # 4. INVOICE PAYMENT (COMPLETE PREVIOUS WORKFLOW)
        # ---------------------------------------------------------------------
        elif "PAYMENT" in event_type or ("INVOICE" in event_type and "PAID" in event_type):
            plan_health = "COMPLETED"
            is_plan_valid = True
            changed_assumptions = ["invoice_remains_overdue"]
            # Stop any collection/dunning/finance steps immediately
            for s in steps:
                if s.status not in ["COMPLETED", "SUCCEEDED"] and (
                    "collection" in s.action_type.lower()
                    or "dunning" in s.action_type.lower()
                    or "finance" in s.action_type.lower()
                    or "reminder" in s.action_type.lower()
                    or "credit" in s.action_type.lower()
                    or "invoice" in s.action_type.lower()
                ):
                    invalidated_step_ids.append(s.step_id)

            recommended_action = "STOP"
            materiality_analysis = "Full payment received. Active collection actions must be safely halted to prevent duplicate messaging."
            replan_rationale = "Workflow goals achieved; halt pending dunning steps."

        # ---------------------------------------------------------------------
        # 5. INVOICE DISPUTE & FINANCIAL ESCALATION
        # ---------------------------------------------------------------------
        elif "DISPUTE" in event_type or ("INVOICE" in event_type and "OVERDUE" in event_type):
            plan_health = "ESCALATED"
            is_plan_valid = False
            changed_assumptions = ["invoice_terms_uncontested", "prompt_payment_commitment"]
            recommended_action = "ESCALATE"
            materiality_analysis = f"Financial dispute or severe delinquency event: {event_type}."
            escalation_details = "Customer raised billing dispute. Autonomous collections paused; routed to Finance Supervisor."

        # ---------------------------------------------------------------------
        # 6. RATE EXPIRY & PRICING INVALIDATION
        # ---------------------------------------------------------------------
        elif "RATE" in event_type and "EXPIR" in event_type:
            plan_health = "STALE"
            is_plan_valid = False
            changed_assumptions = ["spot_rate_commercial_validity", "carrier_pricing_lock"]
            for s in steps:
                if "quote" in s.action_type.lower() or "rate" in s.action_type.lower():
                    invalidated_step_ids.append(s.step_id)
            recommended_action = "REPLAN"
            materiality_analysis = "Carrier base rate or spot validity expired. Quotation plan is now stale."
            replan_rationale = "Replan required to request updated carrier spot rate and recompute margins."

        # ---------------------------------------------------------------------
        # 7. COMPLIANCE & CUSTOMS BLOCKS
        # ---------------------------------------------------------------------
        elif "COMPLIANCE" in event_type or "CUSTOMS_HOLD" in event_type or "SANCTION" in event_type:
            plan_health = "BLOCKED"
            is_plan_valid = False
            changed_assumptions = ["regulatory_clearance_granted", "entity_sanctions_cleared"]
            for s in steps:
                if s.status not in ["COMPLETED", "SUCCEEDED"]:
                    invalidated_step_ids.append(s.step_id)
            recommended_action = "STOP"
            materiality_analysis = "Strict compliance or customs hold triggered. All autonomous operations frozen."
            escalation_details = "Regulatory compliance violation or document hold detected. Human review mandated."

        # ---------------------------------------------------------------------
        # 8. PREDICTION DRIFT (PHASE 4 PREDICTIONS UPDATE)
        # ---------------------------------------------------------------------
        elif "PREDICTION" in event_type or "DRIFT" in event_type:
            prev_pred = req.previous_predictions or {}
            new_pred = req.new_predictions or {}
            prev_conf = float(prev_pred.get("confidence", 0.9))
            new_conf = float(new_pred.get("confidence", 0.5))

            if (prev_conf - new_conf) >= 0.25 or new_pred.get("risk_category") == "CRITICAL":
                plan_health = "AT_RISK"
                is_plan_valid = False
                changed_assumptions = ["predictive_model_high_confidence", "stable_environmental_conditions"]
                recommended_action = "REPLAN"
                materiality_analysis = (
                    f"Significant predictive drift detected: confidence dropped from {prev_conf:.2f} to {new_conf:.2f}. "
                    "Assumed operational margins are no longer statistically reliable."
                )
                replan_rationale = "Replan needed to incorporate wider buffer windows and secondary carrier contingencies."

        return StateChangeEvaluationResponse(
            plan_id=req.plan_id,
            plan_health=plan_health,
            is_plan_valid=is_plan_valid,
            materiality_analysis=materiality_analysis,
            invalidated_step_ids=invalidated_step_ids,
            changed_assumptions=changed_assumptions,
            recommended_action=recommended_action,
            replan_rationale=replan_rationale,
            escalation_details=escalation_details,
            confidence_score=0.88 if is_plan_valid else 0.65,
            correlation_id=corr_id,
        )

    def generate_adaptive_replan(self, req: ContinuousReplanningRequest) -> ContinuousReplanningResponse:
        """
        Generates a revised plan version that strictly PROTECTS completed steps
        (preventing accidental re-execution) while replacing invalidated steps with adapted actions.
        """
        parent_plan = req.current_plan or {}
        parent_plan_id = req.plan_id
        parent_version = int(parent_plan.get("version", 1))
        new_version = parent_version + 1
        new_plan_id = f"plan-rev-v{new_version}-{uuid.uuid4().hex[:8]}"
        corr_id = req.correlation_id or f"corr-replan-{uuid.uuid4().hex[:12]}"

        # 1. Protect completed steps: Keep them verbatim with COMPLETED status
        ordered_steps: List[PlanStepModel] = []
        protected_completed_steps: List[str] = []
        step_number = 1

        for cs in req.completed_steps:
            protected_completed_steps.append(cs.step_id)
            preserved_step = PlanStepModel(
                step_id=cs.step_id,
                step_number=step_number,
                action_type=cs.action_type,
                title=f"[COMPLETED] {cs.title}",
                description=cs.description,
                parameters=cs.parameters,
                dependencies=cs.dependencies,
                expected_outcome=cs.expected_outcome,
                risk_level=cs.risk_level,
                requires_approval=False,  # Already executed
                reversibility=cs.reversibility,
                status="COMPLETED",
                timeout_seconds=cs.timeout_seconds,
                idempotency_key=cs.idempotency_key,
            )
            ordered_steps.append(preserved_step)
            step_number += 1

        # 2. Formulate revised replacement steps based on the triggering event
        last_completed_id = protected_completed_steps[-1] if protected_completed_steps else None
        event_type = req.event.event_type.upper()
        changed_assumptions: List[str] = []

        if "ETA" in event_type or "SLIP" in event_type:
            changed_assumptions = ["vessel_schedule_on_time_guarantee", "delivery_within_customer_sla_commitment"]
            step_a = PlanStepModel(
                step_id=f"step-rev-{step_number}",
                step_number=step_number,
                action_type="carrier.expedited_inquiry",
                title="Carrier Expedited Feeder & Berth Prioritization",
                description=f"Engage carrier operations to expedite transshipment following ETA slip ({req.event.event_type})",
                parameters={"entity_id": req.event.entity_id, "priority": "EXPEDITED"},
                dependencies=[last_completed_id] if last_completed_id else [],
                expected_outcome="Expedited feeder connection confirmed with minimal dwell time",
                risk_level="MEDIUM",
                requires_approval=False,
                reversibility="REVERSIBLE",
                idempotency_key=f"idemp-{new_plan_id}-{step_number}",
            )
            ordered_steps.append(step_a)
            step_number += 1

            step_b = PlanStepModel(
                step_id=f"step-rev-{step_number}",
                step_number=step_number,
                action_type="customers.notify_delay",
                title="Proactive Customer SLA & Revised ETA Advisory",
                description="Issue authoritative advisory with updated predictive ETA and proactive concessions",
                parameters={"entity_id": req.event.entity_id, "notification_channel": "EMAIL_AND_PORTAL"},
                dependencies=[step_a.step_id],
                expected_outcome="Customer acknowledged revised arrival schedule with SLA penalty waiver",
                risk_level="LOW",
                requires_approval=False,
                reversibility="REVERSIBLE",
                idempotency_key=f"idemp-{new_plan_id}-{step_number}",
            )
            ordered_steps.append(step_b)
            step_number += 1

        elif "CARRIER" in event_type:
            changed_assumptions = ["carrier_space_confirmed", "carrier_operational_readiness"]
            step_alt = PlanStepModel(
                step_id=f"step-rev-{step_number}",
                step_number=step_number,
                action_type="carrier.request_rebooking",
                title="Secondary Carrier Fast-Track Rebooking",
                description="Secure guaranteed secondary carrier allocation bypassing congested terminal",
                parameters={"entity_id": req.event.entity_id, "strategy": "FAST_TRACK"},
                dependencies=[last_completed_id] if last_completed_id else [],
                expected_outcome="Confirmed replacement booking on partner carrier",
                risk_level="HIGH",
                requires_approval=True,  # Carrier switch requires approval
                reversibility="REVERSIBLE",
                idempotency_key=f"idemp-{new_plan_id}-{step_number}",
            )
            ordered_steps.append(step_alt)
            step_number += 1

        else:
            changed_assumptions = ["baseline_operational_assumptions_recalibrated"]
            step_gen = PlanStepModel(
                step_id=f"step-rev-{step_number}",
                step_number=step_number,
                action_type="shipments.analyze_context",
                title="Recalibrate Context & Assess Recovery Feasibility",
                description=f"Evaluate newly arrived event {event_type} and execute adjusted mitigation actions",
                parameters={"entity_id": req.event.entity_id},
                dependencies=[last_completed_id] if last_completed_id else [],
                expected_outcome="Mitigation actions updated and verified against policy",
                risk_level="LOW",
                requires_approval=False,
                reversibility="REVERSIBLE",
                idempotency_key=f"idemp-{new_plan_id}-{step_number}",
            )
            ordered_steps.append(step_gen)
            step_number += 1

        # 3. Compute parallel groups using Kahn's algorithm
        from app.autonomy.models import MultiStepPlanValidationRequest
        val_resp = self.planner.validate_plan_graph(
            MultiStepPlanValidationRequest(
                plan_id=new_plan_id,
                steps=ordered_steps,
                module=parent_plan.get("module", "shipments"),
                autonomy_level=req.autonomy_level,
            )
        )

        has_approval = any(s.requires_approval for s in ordered_steps)

        return ContinuousReplanningResponse(
            new_plan_id=new_plan_id,
            version=new_version,
            parent_plan_id=parent_plan_id,
            ordered_steps=ordered_steps,
            parallel_groups=val_resp.parallel_groups,
            protected_completed_steps=protected_completed_steps,
            changed_assumptions=changed_assumptions,
            replan_rationale=f"Plan adapted to Version {new_version} triggered by {event_type}. {len(protected_completed_steps)} completed steps protected.",
            confidence_score=0.91,
            requires_approval=has_approval,
            correlation_id=corr_id,
        )


_monitoring_agent_instance = ContinuousMonitoringAgent()

def evaluate_state_change(req: StateChangeEvaluationRequest) -> StateChangeEvaluationResponse:
    """Module-level entrypoint for state change evaluation."""
    return _monitoring_agent_instance.evaluate_state_change(req)

def replan_continuous_workflow(req: ContinuousReplanningRequest) -> ContinuousReplanningResponse:
    """Module-level entrypoint for continuous adaptive replanning."""
    return _monitoring_agent_instance.generate_adaptive_replan(req)

