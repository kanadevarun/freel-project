import re
import uuid
from typing import Dict, Any, List
from .models import (
    GovernanceContextEvaluationRequest,
    GovernanceContextEvaluationResponse,
    GovernancePlanPreviewRequest,
    GovernancePlanPreviewResponse,
)

INJECTION_PATTERNS = [
    r"(?i)ignore\s+(?:all\s+)?(?:previous\s+)?instructions",
    r"(?i)system\s+override",
    r"(?i)escalate\s+privileges?",
    r"(?i)bypass\s+(?:governance|approval|policy|controls?)",
    r"(?i)grant\s+(?:full\s+)?autonomy",
    r"(?i)set\s+autonomy\s+(?:to\s+)?(?:level\s+)?[34]",
    r"(?i)disable\s+(?:safety|kill[\s_-]?switch|limits?)",
    r"(?i)role:\s*admin",
    r"(?i)sudo\b",
]

HIGH_RISK_ACTIONS = {
    "shipment.cancel": (95.0, "CRITICAL", 1, "Irreversible shipment booking cancellation"),
    "invoice.credit_memo": (88.0, "HIGH", 1, "Direct financial write-off / credit memo issuance"),
    "pricing.apply_discount": (72.0, "HIGH", 2, "Commercial margin adjustment"),
    "invoice.dispute_hold": (68.0, "HIGH", 2, "Collections hold on disputed invoice"),
    "shipment.reroute": (75.0, "HIGH", 2, "Physical freight diversion to alternate corridor"),
}

MEDIUM_RISK_ACTIONS = {
    "carrier.escalate": (48.0, "MEDIUM", 3, "External operational escalation to carrier"),
    "customer.followup": (42.0, "MEDIUM", 3, "Customer-facing milestone advisory"),
    "customs.request_doc": (38.0, "MEDIUM", 3, "Statutory document submission request"),
}

LOW_RISK_ACTIONS = {
    "task.create": (15.0, "LOW", 4, "Internal operational task dispatch"),
    "notification.internal": (10.0, "LOW", 4, "Internal notification alert"),
}

class GovernanceAgent:
    """
    Python AI Sidecar Governance Agent for LogisticsHQ.
    Responsible for:
      - Multi-factor risk interpretation & scoring
      - Data sufficiency analysis (identifying missing parameters without fabrication)
      - Confidence categorization
      - Recommended autonomy tier assessment
      - Blast radius estimation & dry-run plan preview
      - Prompt injection sanitization
    """

    def _sanitize(self, text: str) -> str:
        if not text:
            return ""
        sanitized = text
        for pattern in INJECTION_PATTERNS:
            sanitized = re.sub(pattern, "[SANITIZED_INSTRUCTION]", sanitized)
        return sanitized

    def evaluate_context(self, req: GovernanceContextEvaluationRequest) -> GovernanceContextEvaluationResponse:
        corr_id = req.correlation_id or f"corr-gov-eval-{uuid.uuid4().hex[:12]}"
        
        # 1. Sanitize text
        sanitized_ctx = self._sanitize(req.context_text)
        
        # 2. Risk scoring based on action type
        act = req.action_type
        if act in HIGH_RISK_ACTIONS:
            base_score, risk_class, rec_tier, blast_desc = HIGH_RISK_ACTIONS[act]
        elif act in MEDIUM_RISK_ACTIONS:
            base_score, risk_class, rec_tier, blast_desc = MEDIUM_RISK_ACTIONS[act]
        elif act in LOW_RISK_ACTIONS:
            base_score, risk_class, rec_tier, blast_desc = LOW_RISK_ACTIONS[act]
        else:
            base_score, risk_class, rec_tier, blast_desc = (60.0, "MEDIUM", 2, "Uncataloged operational action")

        # Adjust for financial exposure in parameters
        params = req.parameters or {}
        fin_amount = float(params.get("amount", params.get("discount_amount", params.get("cost_impact", 0.0))))
        if fin_amount > 2000.0:
            base_score = min(100.0, base_score + 15.0)
            risk_class = "HIGH"
            rec_tier = min(rec_tier, 2)
        elif fin_amount > 500.0:
            base_score = min(100.0, base_score + 8.0)

        # 3. Data sufficiency check
        missing_fields = []
        if not req.entity_id or req.entity_id in ["0", "unknown"]:
            missing_fields.append("entity_id")
        if act in ["pricing.apply_discount", "invoice.credit_memo"] and "amount" not in params and "discount_amount" not in params:
            missing_fields.append("financial_amount")
        if act in ["customer.followup"] and not params.get("recipient_email") and not params.get("customer_email") and not req.entity_id:
            missing_fields.append("recipient_contact")

        if len(missing_fields) > 1:
            data_sufficiency = "INSUFFICIENT"
            rec_tier = min(rec_tier, 1)
        elif len(missing_fields) == 1:
            data_sufficiency = "PARTIALLY_SUFFICIENT"
            rec_tier = min(rec_tier, 2)
        else:
            data_sufficiency = "SUFFICIENT"

        # 4. Confidence classification
        if data_sufficiency == "SUFFICIENT" and base_score < 70.0:
            confidence_class = "HIGH"
            confidence_score = 0.92
        elif data_sufficiency == "PARTIALLY_SUFFICIENT" or (base_score >= 70.0 and base_score < 90.0):
            confidence_class = "MEDIUM"
            confidence_score = 0.74
        else:
            confidence_class = "LOW"
            confidence_score = 0.48

        # 5. Blast radius & explanation
        explanation = (
            f"Action '{act}' on {req.entity_type} #{req.entity_id} classified as {risk_class} risk "
            f"(Score: {base_score:.1f}/100, Sufficiency: {data_sufficiency}). "
            f"Recommended autonomy: Level {rec_tier}."
        )

        return GovernanceContextEvaluationResponse(
            risk_score=base_score,
            risk_class=risk_class,
            data_sufficiency=data_sufficiency,
            missing_data_fields=missing_fields,
            confidence_class=confidence_class,
            confidence_score=confidence_score,
            recommended_autonomy_tier=rec_tier,
            blast_radius_assessment=f"{blast_desc} (Financial Exposure: ${fin_amount:.2f})",
            explanation=explanation,
            sanitized_context=sanitized_ctx,
            correlation_id=corr_id
        )

    def preview_plan(self, req: GovernancePlanPreviewRequest) -> GovernancePlanPreviewResponse:
        corr_id = req.correlation_id or f"corr-gov-prev-{uuid.uuid4().hex[:12]}"
        
        steps = req.steps or []
        total_steps = len(steps)
        executable_steps = 0
        approval_required_steps = 0
        total_exposure = 0.0
        max_risk = "LOW"
        affected_entities = []

        risk_hierarchy = {"LOW": 1, "MEDIUM": 2, "HIGH": 3, "CRITICAL": 4}

        for step in steps:
            act = step.get("action_type", "")
            params = step.get("parameters", {})
            entity_type = step.get("entity_type", "unknown")
            entity_id = str(step.get("entity_id", ""))
            
            if entity_id and {"entity_type": entity_type, "entity_id": entity_id} not in affected_entities:
                affected_entities.append({"entity_type": entity_type, "entity_id": entity_id})

            cost = float(params.get("amount", params.get("cost_impact", 0.0)))
            total_exposure += cost

            if act in HIGH_RISK_ACTIONS:
                step_risk = HIGH_RISK_ACTIONS[act][1]
                approval_required_steps += 1
            elif act in MEDIUM_RISK_ACTIONS:
                step_risk = MEDIUM_RISK_ACTIONS[act][1]
                if step.get("requires_approval", False):
                    approval_required_steps += 1
                else:
                    executable_steps += 1
            else:
                step_risk = "LOW"
                executable_steps += 1

            if risk_hierarchy.get(step_risk, 1) > risk_hierarchy.get(max_risk, 1):
                max_risk = step_risk

        summary = (
            f"Dry-run simulation for Plan '{req.plan_id}' ({req.module}): {total_steps} total steps, "
            f"{executable_steps} autonomous candidates, {approval_required_steps} approval gates. "
            f"Max risk: {max_risk}, Estimated financial exposure: ${total_exposure:.2f}."
        )

        return GovernancePlanPreviewResponse(
            plan_id=req.plan_id,
            total_steps=total_steps,
            executable_steps=executable_steps,
            approval_required_steps=approval_required_steps,
            max_risk_class=max_risk,
            estimated_financial_exposure_usd=total_exposure,
            affected_entities=affected_entities,
            safety_summary=summary,
            correlation_id=corr_id
        )

governance_agent = GovernanceAgent()
