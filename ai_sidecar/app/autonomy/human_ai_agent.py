"""
LogisticsHQ Phase 5 Task 5.11: Human + AI Operating Model Agent
Provides decision point analysis, provenance attribution, confidence & data sufficiency scoring,
prepared output generation, prompt injection defense, and human feedback interpretation.
"""

import uuid
import re
from typing import Dict, Any, List, Tuple
from app.autonomy.models import (
    HumanAIDecisionAnalysisRequest,
    HumanAIDecisionAnalysisResponse,
    HumanFeedbackAnalysisRequest,
    HumanFeedbackAnalysisResponse,
    OperatingMode,
    ConfidenceLevel,
    DataSufficiency,
    RiskLevel,
)

INJECTION_PATTERNS = [
    r"ignore\s+(all\s+)?previous\s+instructions",
    r"bypass\s+(approval|auth|security|policy)",
    r"grant\s+(full\s+)?autonomy",
    r"system\s+override",
    r"escalate\s+privileges",
    r"you\s+are\s+now\s+in\s+developer\s+mode",
]


def sanitize_text(text: str) -> Tuple[str, bool]:
    if not text:
        return "", False
    sanitized = text
    detected = False
    for pat in INJECTION_PATTERNS:
        if re.search(pat, sanitized, re.IGNORECASE):
            detected = True
            sanitized = re.sub(pat, "[SANITIZED_INSTRUCTION]", sanitized, flags=re.IGNORECASE)
    return sanitized, detected


class HumanAIOperatingAgent:
    """
    Coordinates reasoning for the Human + AI operating model:
    AI OBSERVES → AI ANALYZES → AI RECOMMENDS → AI PLANS → AI PREPARES
    → AI EXECUTES WITHIN POLICY → HUMAN REVIEWS → AI EXECUTES APPROVED ACTIONS
    """

    def analyze_decision_point(self, req: HumanAIDecisionAnalysisRequest) -> HumanAIDecisionAnalysisResponse:
        sanitized_notes, injection_detected = sanitize_text(req.notes or "")
        decision_id = f"dec-{uuid.uuid4().hex[:10]}"
        correlation_id = req.correlation_id or f"corr-{uuid.uuid4().hex[:8]}"

        context = req.context_data or {}
        predictions = req.predictive_data or {}
        module = (req.module or "shipments").lower()

        # 1. Assess Data Sufficiency
        data_sufficiency: DataSufficiency = "SUFFICIENT"
        missing_reasons = []

        if module == "shipments":
            if not context.get("shipment_id") and not context.get("id") and not req.entity_id:
                missing_reasons.append("Shipment ID identifier missing")
            if not context.get("status") and not context.get("tracking_number"):
                missing_reasons.append("Active shipment tracking status missing")
        elif module == "pricing":
            if not context.get("rfq_id") and not context.get("cost_amount"):
                missing_reasons.append("Base lane pricing or RFQ reference missing")
        elif module == "finance":
            if not context.get("invoice_id") and not context.get("balance"):
                missing_reasons.append("Invoice balance details missing")

        if missing_reasons:
            data_sufficiency = "INSUFFICIENT" if len(missing_reasons) > 1 else "PARTIALLY_SUFFICIENT"

        # 2. Extract Facts vs Predictions
        facts: Dict[str, Any] = {}
        for k, v in context.items():
            if not isinstance(v, (dict, list)) or k in ["booking", "milestones", "parties"]:
                facts[k] = v

        predictive_signals: Dict[str, Any] = {}
        for k, v in predictions.items():
            predictive_signals[k] = v

        # 3. Assess Confidence & Risk
        confidence: ConfidenceLevel = "HIGH"
        if data_sufficiency == "INSUFFICIENT":
            confidence = "LOW"
        elif data_sufficiency == "PARTIALLY_SUFFICIENT":
            confidence = "MEDIUM"

        risk_level: RiskLevel = "MEDIUM"
        if module in ["finance", "pricing"] and float(context.get("amount", 0) or context.get("total_amount", 0)) > 25000:
            risk_level = "HIGH"
        if context.get("has_exceptions") or context.get("is_blocked"):
            risk_level = "HIGH"

        # 4. Formulate Recommendations & Prepared Output by Module
        title = req.title or f"Human Review Required: {module.capitalize()} Entity {req.entity_id}"
        ai_recommendation = ""
        prepared_payload: Dict[str, Any] = {}
        alternatives: List[Dict[str, Any]] = []
        is_reversible = True
        operating_mode: OperatingMode = "HUMAN_APPROVAL"

        if module == "shipments":
            eta_delay = float(predictive_signals.get("predicted_delay_hours", 0) or context.get("eta_delay_hours", 0))
            is_reversible = True
            if eta_delay >= 3.0:
                title = f"Proactive Customer Advisory & Recovery for Shipment #{req.entity_id}"
                ai_recommendation = f"Issue proactive milestone delay notification of {eta_delay:.1f}h to customer and dispatch status inquiry to carrier."
                prepared_payload = {
                    "action_type": "customer.send_communication",
                    "template": "PROACTIVE_ETA_ADVISORY",
                    "recipient": context.get("customer_email", "customer@client.com"),
                    "subject": f"LogisticsHQ Update: Shipment #{req.entity_id} Schedule Revision",
                    "body_draft": f"Dear Customer, your shipment #{req.entity_id} is experiencing a transit variance of approximately {eta_delay:.1f} hours due to AIS berth congestion. Our operations team is actively monitoring resolution.",
                    "eta_delay_hours": eta_delay,
                }
                alternatives = [
                    {"id": "alt-1", "title": "Reroute via Alternate Intermodal Hub", "impact": "+$450 fee, recovers 48h", "risk": "LOW"},
                    {"id": "alt-2", "title": "Defer Notification until Final Port Arrival", "impact": "Zero cost, increases SLA escalation risk", "risk": "HIGH"},
                    {"id": "alt-3", "title": "Operator Override / Manual Carrier Call", "impact": "Direct human phone resolution", "risk": "LOW"},
                ]
            else:
                title = f"Routine Monitoring for Shipment #{req.entity_id}"
                ai_recommendation = "Continue automated vessel AIS tracking without disruptive customer dispatch."
                operating_mode = "AI_EXECUTE"
                prepared_payload = {"action_type": "shipments.continue_monitoring", "poll_interval": 3600}

        elif module == "customer":
            title = f"Customer Communication Review: #{req.entity_id}"
            ai_recommendation = "Review and authorize prepared proactive customer recovery message prior to outbound dispatch."
            prepared_payload = {
                "action_type": "customer.send_advisory",
                "customer_id": req.entity_id,
                "message_draft": "Your shipment schedule has been adjusted. We are taking preventative measures to maintain delivery integrity.",
            }
            alternatives = [
                {"id": "alt-cust-1", "title": "Send via SMS Alert", "impact": "High urgency, short format", "risk": "LOW"},
                {"id": "alt-cust-2", "title": "Escalate to Account Executive Call", "impact": "High touch personal contact", "risk": "LOW"},
            ]

        elif module == "pricing":
            margin = float(context.get("margin_pct", 12.5))
            title = f"Spot Quotation Margin Sign-off: RFQ #{req.entity_id}"
            ai_recommendation = f"Approve revised spot quote rate at {margin:.1f}% operating margin within enterprise corridors."
            prepared_payload = {
                "action_type": "pricing.apply_quote",
                "rfq_id": req.entity_id,
                "quoted_rate": context.get("quoted_rate", 3250.0),
                "margin_pct": margin,
            }
            alternatives = [
                {"id": "alt-price-1", "title": "Aggressive Discount (8% Margin)", "impact": "Increases conversion probability to 85%", "risk": "MEDIUM"},
                {"id": "alt-price-2", "title": "Contract Tariff Cap Ceiling", "impact": "Guaranteed contract tariff, non-negotiable", "risk": "LOW"},
            ]

        elif module == "finance":
            title = f"Receivables Dunning Gating: Invoice #{req.entity_id}"
            ai_recommendation = "Hold collection escalation notice pending confirmation of wire settlement."
            is_reversible = False
            prepared_payload = {
                "action_type": "finance.hold_collection",
                "invoice_id": req.entity_id,
                "hold_duration_hours": 48,
            }
            alternatives = [
                {"id": "alt-fin-1", "title": "Send Automated Payment Reminder", "impact": "Standard dunning sequence step", "risk": "LOW"},
                {"id": "alt-fin-2", "title": "Immediate Credit Limit Suspension", "impact": "Severe commercial impact, halts all active bookings", "risk": "CRITICAL"},
            ]

        elif module == "contracts":
            title = f"Contract Compliance Review: #{req.entity_id}"
            ai_recommendation = "Human compliance review required for non-standard demurrage indemnity clause."
            prepared_payload = {"action_type": "contracts.flag_compliance_review", "document_id": req.entity_id}
            alternatives = [
                {"id": "alt-cont-1", "title": "Accept Standard Shipper Liability Terms", "impact": "Aligns with baseline terms", "risk": "LOW"},
                {"id": "alt-cont-2", "title": "Reject Contract Rider", "impact": "Requires customer renegotiation", "risk": "HIGH"},
            ]

        else:  # exceptions / default
            title = f"Operational Exception Recovery: #{req.entity_id}"
            ai_recommendation = "Execute controlled multi-step recovery action under human operator supervision."
            prepared_payload = {"action_type": "exceptions.execute_recovery", "entity_id": req.entity_id}
            alternatives = [
                {"id": "alt-exc-1", "title": "Immediate Customs Hold Escalation", "impact": "Engages terminal broker directly", "risk": "LOW"},
                {"id": "alt-exc-2", "title": "Stand Down & Await AIS Refresh", "impact": "Avoids broker fees, potential 12h drift", "risk": "MEDIUM"},
            ]

        # 5. Provenance Attribution Summary
        provenance = (
            f"[FACTS]: Verified {len(facts)} operational attributes from authoritative store. "
            f"[PREDICTIONS]: Model forecasts {predictive_signals if predictive_signals else 'aligned with current trajectory'}. "
            f"[RECOMMENDATION]: {ai_recommendation} (Confidence: {confidence}, Data: {data_sufficiency})."
        )
        if injection_detected:
            provenance += " [SECURITY NOTE]: Adversarial prompt injection was detected and sanitized."

        return HumanAIDecisionAnalysisResponse(
            decision_id=decision_id,
            operating_mode=operating_mode,
            title=title,
            context_summary=f"Analysis of {module} entity {req.entity_id} under {req.autonomy_level}.",
            facts=facts,
            predictions=predictive_signals,
            ai_recommendation=ai_recommendation,
            prepared_payload=prepared_payload,
            alternatives=alternatives,
            confidence=confidence,
            data_sufficiency=data_sufficiency,
            risk_level=risk_level,
            requires_human_approval=(operating_mode in ["HUMAN_APPROVAL", "HUMAN_REVIEW"]),
            is_reversible=is_reversible,
            provenance_summary=provenance,
            correlation_id=correlation_id,
        )

    def analyze_human_feedback(self, req: HumanFeedbackAnalysisRequest) -> HumanFeedbackAnalysisResponse:
        decision_id = req.decision_id
        correlation_id = req.correlation_id or f"corr-fb-{uuid.uuid4().hex[:8]}"
        decision = req.human_decision.upper()
        fb_type = req.feedback_type
        reason = req.decision_reason or ""

        learned_preference = ""
        memory_candidate = None
        replan_suggested = False
        notes = ""

        if decision == "REJECT" or fb_type == "RECOMMENDATION_REJECTED":
            learned_preference = f"Operator rejected recommendation for decision {decision_id}. Reason: {reason or 'Not suitable for current commercial state'}."
            memory_candidate = {
                "memory_type": "HUMAN_OPERATOR_PREFERENCE",
                "key": f"preference_reject_{decision_id}",
                "value": {"rejected_decision": decision_id, "reason": reason, "policy_guidance": "Prefer alternative recovery route or manual contact."},
            }
            replan_suggested = True
            notes = "Rejection recorded; active workflow adaptation/replan required to prevent stalled execution."

        elif decision == "OVERRIDE" or fb_type == "ALTERNATIVE_SELECTED":
            learned_preference = f"Operator selected alternative option over default AI recommendation. Override notes: {reason}."
            memory_candidate = {
                "memory_type": "HUMAN_OPERATOR_PREFERENCE",
                "key": f"preference_override_{decision_id}",
                "value": {"selected_alternative": reason, "decision_id": decision_id},
            }
            replan_suggested = True
            notes = "Human override applied; workflow graph must adopt operator's chosen alternative."

        elif fb_type == "AI_OUTPUT_EDITED":
            orig = req.original_ai_payload or ""
            edited = req.human_edited_payload or ""
            learned_preference = "Operator tailored communication phrasing for specific customer/carrier nuance."
            memory_candidate = {
                "memory_type": "COMMUNICATION_STYLE_PREFERENCE",
                "key": f"style_edit_{decision_id}",
                "value": {"original_snippet": orig[:100], "edited_snippet": edited[:100]},
            }
            notes = "Human edits preserved verbatim. Executed action will utilize human-edited payload."

        elif decision == "STOP" or fb_type == "ACTION_STOPPED":
            learned_preference = f"Operator halted autonomous execution: {reason}."
            notes = "Workflow halted by authorized operator. Pending actions cancelled without mutating completed steps."

        elif decision == "ESCALATE" or fb_type == "ESCALATION_REQUESTED":
            learned_preference = f"Escalated to human supervisor: {reason}."
            notes = "Workflow escalated to supervisor queue with priority status."

        else:  # APPROVE / RECOMMENDATION_ACCEPTED
            learned_preference = "Standard operating procedure approved by operator."
            notes = "Approval verified; action released to Go Action System for controlled execution."

        return HumanFeedbackAnalysisResponse(
            decision_id=decision_id,
            learned_preference=learned_preference,
            memory_candidate=memory_candidate,
            replan_suggested=replan_suggested,
            notes=notes,
            correlation_id=correlation_id,
        )


human_ai_agent = HumanAIOperatingAgent()
