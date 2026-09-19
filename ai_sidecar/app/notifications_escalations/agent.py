"""
Notifications and Escalations AI Agent (Python Only)
Implements Phase 3 Task 3.9:
- Notification summarization & alert fatigue reduction
- Intelligent prioritization and risk scoring
- Escalation reasoning & multi-tier trajectory
- Semantic issue clustering (group_key)
- Role-aware recipient suggestion
- Draft notification & escalation message generation
- Prompt-injection sanitization
"""
import re
from typing import Dict, Any, List, Optional
from .models import (
    NotificationAnalysisRequest,
    NotificationAnalysisResponse,
    EscalationDraftRequest,
    EscalationDraftResponse,
)


class NotificationsEscalationsAgent:
    """
    Dedicated AI Agent for intelligent notification processing,
    deduplication grouping, priority optimization, and escalation reasoning.
    """

    def __init__(self):
        pass

    def sanitize_text(self, text: str) -> str:
        """Sanitize incoming text against prompt-injection and control characters."""
        if not text:
            return ""
        # Remove prompt injection override patterns
        cleaned = re.sub(r"(?i)(ignore previous instructions|system prompt|bypass safety|you are now)", "[REDACTED]", text)
        # Strip control characters
        cleaned = re.sub(r"[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]", "", cleaned)
        return cleaned.strip()

    def analyze_and_prioritize(self, req: NotificationAnalysisRequest) -> NotificationAnalysisResponse:
        """
        Evaluate notification context, calculate priority score,
        determine escalation rationale, and generate overload reduction advice.
        """
        clean_title = self.sanitize_text(req.title)
        clean_message = self.sanitize_text(req.message)
        payload = req.context_payload or {}

        # 1. Base Priority Score from Severity
        sev_upper = (req.severity or "INFORMATIONAL").upper()
        base_score = 20.0
        if sev_upper == "CRITICAL":
            base_score = 80.0
        elif sev_upper == "HIGH":
            base_score = 60.0
        elif sev_upper == "MEDIUM":
            base_score = 40.0
        elif sev_upper == "LOW":
            base_score = 25.0
        else:
            base_score = 15.0

        # 2. Time & Unresolved Duration Multipliers
        # Aging unacknowledged alerts gain urgency over time
        duration_factor = min(req.unresolved_duration_hours * 1.5, 15.0)
        escalation_factor = min(req.escalation_level * 5.0, 15.0)

        # 3. Domain Impact Modifiers
        domain_factor = 0.0
        suggested_role = "OPERATIONS"
        suggested_action = "Acknowledge notification and inspect source record"
        requires_approval = False
        risk_level = "LOW"

        mod_upper = req.source_module.upper()
        if mod_upper == "SHIPMENTS":
            suggested_role = "OPERATIONS"
            delay_hours = payload.get("delay_hours", 0) or payload.get("dwell_hours", 0)
            if delay_hours > 24 or "berth" in clean_title.lower() or "customs" in clean_title.lower() or "delay" in clean_title.lower() or req.severity.upper() == "CRITICAL":
                domain_factor += 10.0
                suggested_action = "Review carrier delay logs and request revised ETA"
                risk_level = "HIGH"
                requires_approval = True
        elif mod_upper == "INVOICES" or mod_upper == "FINANCE":
            suggested_role = "FINANCE"
            amount = payload.get("outstanding_amount", 0)
            if amount > 10000 or payload.get("days_overdue", 0) > 14:
                domain_factor += 12.0
                suggested_action = "Initiate formal payment reminder or credit review"
                risk_level = "HIGH"
                requires_approval = True
        elif mod_upper == "CONTRACTS":
            suggested_role = "COMPLIANCE"
            days_left = payload.get("days_until_expiry", 30)
            if days_left <= 14 or payload.get("requires_compliance_audit", False):
                domain_factor += 10.0
                suggested_action = "Commence contract renewal audit and SLA renegotiation"
                risk_level = "MEDIUM"
        elif mod_upper == "APPROVALS":
            suggested_role = "MANAGEMENT"
            domain_factor += 8.0
            suggested_action = "Review pending commercial action in Approvals Center"
            risk_level = "HIGH"
        elif mod_upper == "LEADS" or mod_upper == "RFQ":
            suggested_role = "SALES"
            domain_factor += 5.0
            suggested_action = "Engage prospect before competitive expiration window closes"

        # Calculate final priority score
        final_score = min(max(base_score + duration_factor + escalation_factor + domain_factor, 0.0), 100.0)

        # Determine recommended priority label
        if final_score >= 75.0:
            rec_priority = "URGENT"
        elif final_score >= 50.0:
            rec_priority = "HIGH"
        elif final_score >= 25.0:
            rec_priority = "MEDIUM"
        else:
            rec_priority = "LOW"

        # 4. Generate Semantic Cluster Grouping Key (to avoid alert fatigue)
        # Groups related events on the same entity together
        group_key = f"cluster:{req.source_module.lower()}:{req.source_record_type.lower()}:{req.source_record_id}"

        # 5. Formulate AI Summary
        summary = (
            f"Operational signal on {req.source_record_type} #{req.source_record_id} ({clean_title}). "
            f"Evaluated as {rec_priority} priority (Score {final_score:.1f}/100) for the {suggested_role} team."
        )

        # 6. Formulate Escalation Reasoning
        if req.escalation_level > 0 or final_score >= 60.0:
            escalation_reason = (
                f"Issue unresolved for {req.unresolved_duration_hours:.1f}h exceeding standard {suggested_role} SLA. "
                f"Escalation Level {req.escalation_level} triggered due to potential downstream schedule or financial breach."
            )
        else:
            escalation_reason = "Standard priority notice; within normal operational resolution window."

        # 7. Overload Reduction Advice
        if final_score < 40.0:
            overload_advice = "Low-impact notice: batch into daily digest to reduce operator cognitive fatigue."
        elif final_score < 70.0:
            overload_advice = "Moderate-priority: route to active in-app notifications without buzzer or SMS."
        else:
            overload_advice = "High-urgency event: surface at top of Notification Center and notify active duty coordinator."

        return NotificationAnalysisResponse(
            ai_summary=summary,
            ai_escalation_reason=escalation_reason,
            recommended_priority=rec_priority,
            priority_score=round(final_score, 1),
            recommended_role_target=suggested_role,
            suggested_action=suggested_action,
            requires_approval=requires_approval,
            risk_level=risk_level,
            group_key=group_key,
            overload_reduction_advice=overload_advice,
            confidence_score=0.93,
            correlation_id=req.correlation_id,
        )

    def generate_escalation_draft(self, req: EscalationDraftRequest) -> EscalationDraftResponse:
        """
        Synthesize formal internal escalation memos or external client advisories.
        Ensures clear [AI DRAFT] demarcation and approval gating.
        """
        clean_type = self.sanitize_text(req.draft_type).upper()
        findings_bullets = "\n".join([f"• {self.sanitize_text(f)}" for f in req.key_findings]) or "• Operational threshold breach requiring managerial intervention."

        if clean_type in ("CLIENT_ADVISORY", "CUSTOMER_UPDATE", "CARRIER_FOLLOW_UP"):
            subject = f"[AI DRAFT] URGENT ADVISORY: Schedule Adjustment Notice for {req.source_module} #{req.source_record_id}"
            body_text = (
                f"Dear Valued Client,\n\n"
                f"We are providing an urgent operational update regarding {req.source_module} reference #{req.source_record_id}.\n\n"
                f"Operational Context & Root Cause:\n{findings_bullets}\n\n"
                f"Remediation Plan:\n"
                f"Our operations team is actively coordinating with carriers to minimize impact and expedite clearance. "
                f"A revised timeline will be shared as soon as verified.\n\n"
                f"Sincerely,\nLogisticsHQ Operational Command"
            )
            is_external = True
            requires_approval = True
            recommended_channels = ["IN_APP", "EMAIL_DIGEST"]
        elif clean_type in ("MANAGEMENT_ALERT", "EXECUTIVE_ESCALATION"):
            subject = f"[AI DRAFT] INTERNAL ESCALATION (Tier {req.escalation_level}): {req.source_module} #{req.source_record_id}"
            body_text = (
                f"Attention: {req.recipient_role} Leadership,\n\n"
                f"An operational issue on {req.source_module} #{req.source_record_id} has exceeded resolution thresholds "
                f"and has been escalated to Tier {req.escalation_level}.\n\n"
                f"Key Escalation Findings:\n{findings_bullets}\n\n"
                f"Recommended Next Action:\n"
                f"Please review this case in the Centralized Approvals Center or assign a senior lead to intervene immediately.\n\n"
                f"LogisticsHQ Intelligent Monitoring Engine"
            )
            is_external = False
            requires_approval = True
            recommended_channels = ["IN_APP", "EMAIL_DIGEST", "URGENT_SMS"]
        else: # OPERATIONAL_ALERT / INTERNAL_ESCALATION
            subject = f"[AI DRAFT] Remediation Request: {req.source_module} #{req.source_record_id}"
            body_text = (
                f"Team {req.recipient_role},\n\n"
                f"Please take notice of the following unresolved operational signal for #{req.source_record_id}:\n\n"
                f"{findings_bullets}\n\n"
                f"Action Required: Acknowledge notification and update corrective status in the workspace."
            )
            is_external = False
            requires_approval = True if req.escalation_level >= 2 else False
            recommended_channels = ["IN_APP"]

        return EscalationDraftResponse(
            subject=subject,
            body_text=body_text,
            recommended_channels=recommended_channels,
            is_external=is_external,
            requires_approval=requires_approval,
            confidence_score=0.95,
            correlation_id=req.correlation_id,
        )
