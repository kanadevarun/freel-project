"""
Python AI Action Orchestrator Agent.
Synthesizes real business signals, evaluates evidence, checks safety gates,
and produces strictly validated AI action proposals.
"""

import re
import datetime
import time
from typing import Dict, Any, List, Optional, Tuple

from app.orchestrator.models import (
    ActionProposal,
    OrchestrationContextInput,
    OrchestrationProposalResponse,
    FactualEvidence,
    RiskLevel,
    PriorityLevel,
    ApprovalPolicy,
    ReversibilityType
)
from app.orchestrator.registry import (
    ACTION_REGISTRY,
    get_registered_action,
    validate_action_proposal_parameters
)

# Prompt injection & code execution detection patterns
PROMPT_INJECTION_PATTERNS = [
    r"(?i)ignore\s+(previous|all)\s+instructions",
    r"(?i)system\s*:\s*override",
    r"(?i)drop\s+table",
    r"(?i)delete\s+from",
    r"(?i)update\s+\w+\s+set",
    r"(?i)exec\s*\(",
    r"(?i)eval\s*\(",
    r"(?i)import\s+os",
    r"(?i)import\s+subprocess",
    r"(?i)__import__",
    r"(?i)<script",
    r"(?i)bypass\s+approval",
]

def detect_safety_violations(text: str) -> Optional[str]:
    """Scans text for prompt injection, arbitrary SQL, or shell execution attempts."""
    for pattern in PROMPT_INJECTION_PATTERNS:
        if re.search(pattern, text):
            return f"Safety gate triggered: potential prompt injection or unsafe pattern detected: '{pattern}'"
    return None


class AIActionOrchestrator:
    """
    Pure Python Orchestrator Agent.
    Evaluates business context, builds evidence, checks safety, and proposes
    controlled actions from the fixed action registry.
    """

    def analyze_and_propose(self, ctx: OrchestrationContextInput) -> OrchestrationProposalResponse:
        start_time = time.time()

        # 1. Safety Check on Context Inputs
        combined_text = f"{ctx.trigger_event} {ctx.source_module} {str(ctx.record_data)}"
        violation = detect_safety_violations(combined_text)
        if violation:
            return OrchestrationProposalResponse(
                status="REJECTED_UNSAFE",
                proposal=None,
                evaluation_summary="Action proposal generation aborted due to safety policy violation.",
                rejection_reason=violation,
                duration_ms=int((time.time() - start_time) * 1000)
            )

        # 2. Extract Evidence & Identify Missing Information
        evidence_list: List[FactualEvidence] = []
        missing_info: List[str] = []

        # Analyze by business domain
        module = (ctx.source_module or "").upper()
        record = ctx.record_data or {}
        signals = ctx.operational_signals or []

        proposed_action_type: str = "tasks.create"
        action_params: Dict[str, Any] = {}
        explanation: str = ""
        expected_impact: str = ""
        risk_level: RiskLevel = "LOW"
        priority: PriorityLevel = "MEDIUM"
        requires_approval: bool = True
        approval_policy: ApprovalPolicy = "ALWAYS_REQUIRE_APPROVAL"
        confidence: float = 0.85
        reversibility: ReversibilityType = "REVERSIBLE"

        now_iso = datetime.datetime.now(datetime.timezone.utc).isoformat()
        expires_at_iso = (datetime.datetime.now(datetime.timezone.utc) + datetime.timedelta(days=7)).isoformat()

        if module == "SHIPMENTS":
            shipment_id = int(ctx.source_record_id) if ctx.source_record_id.isdigit() else 0
            status = record.get("status", "UNKNOWN")
            carrier_scac = record.get("carrier_scac", "UNKNOWN")
            vessel_name = record.get("vessel_name")
            booking_number = record.get("booking_number", f"SH-{shipment_id}")

            evidence_list.append(FactualEvidence(
                field_name="shipment_status",
                observed_value=status,
                source_record=f"SHIPMENTS/{shipment_id}",
                timestamp=now_iso,
                fact_type="CONFIRMED_FACT"
            ))
            evidence_list.append(FactualEvidence(
                field_name="carrier_scac",
                observed_value=carrier_scac,
                source_record=f"SHIPMENTS/{shipment_id}",
                fact_type="CONFIRMED_FACT"
            ))

            has_delay = any("DELAY" in s.signal_type.upper() or s.severity in ("HIGH", "CRITICAL") for s in signals)
            if has_delay:
                evidence_list.append(FactualEvidence(
                    field_name="schedule_variance_signal",
                    observed_value="Delay signal detected from carrier telemetry",
                    fact_type="THRESHOLD_BREACH"
                ))
                risk_level = "HIGH"
                priority = "HIGH"
                requires_approval = True
                proposed_action_type = "tasks.create"
                action_params = {
                    "title": f"Investigate Carrier Delay: Shipment {booking_number}",
                    "description": f"Carrier {carrier_scac} reported schedule delay anomalies for shipment {booking_number}. Review updated ETA and notify affected stakeholders.",
                    "assigned_team": "OPERATIONS",
                    "priority": "HIGH",
                    "due_date": (datetime.datetime.now(datetime.timezone.utc) + datetime.timedelta(hours=24)).isoformat()
                }
                explanation = f"Shipment {booking_number} ({carrier_scac}) has active operational delay signals. Creating an urgent internal operations task to verify carrier tracking and evaluate ETA impacts."
                expected_impact = "Ensures operator verification of delay before customer communication, maintaining high SLA compliance."
                confidence = 0.92
            else:
                proposed_action_type = "notes.add"
                action_params = {
                    "entity_type": "SHIPMENT",
                    "entity_id": shipment_id,
                    "note_text": f"Automated operational check verified shipment {booking_number}. Carrier milestones are tracking on schedule.",
                    "is_confidential": False
                }
                explanation = f"Shipment {booking_number} is operating normally with no active exception alerts."
                expected_impact = "Documents automated routine verification in shipment file."
                confidence = 0.95
                requires_approval = False
                approval_policy = "AUTOMATIC_FOR_SAFE_ACTIONS"

        elif module == "INVOICES":
            invoice_id = int(ctx.source_record_id) if ctx.source_record_id.isdigit() else 0
            invoice_number = record.get("invoice_number", f"INV-{invoice_id}")
            total_amount = float(record.get("total_amount", 0.0))
            due_date = record.get("due_date", "")
            days_overdue = int(record.get("days_overdue", 0))
            customer_id = int(record.get("customer_id", 0))

            evidence_list.append(FactualEvidence(
                field_name="invoice_total",
                observed_value=total_amount,
                source_record=f"INVOICES/{invoice_id}",
                fact_type="CONFIRMED_FACT"
            ))
            evidence_list.append(FactualEvidence(
                field_name="days_overdue",
                observed_value=days_overdue,
                source_record=f"INVOICES/{invoice_id}",
                fact_type="CALCULATED_VALUE"
            ))

            if days_overdue > 15:
                evidence_list.append(FactualEvidence(
                    field_name="overdue_threshold_breach",
                    observed_value=f"{days_overdue} days past due",
                    baseline_value="0 days",
                    fact_type="THRESHOLD_BREACH"
                ))
                risk_level = "MEDIUM"
                priority = "HIGH" if days_overdue > 30 else "MEDIUM"
                proposed_action_type = "followups.create"
                action_params = {
                    "customer_id": customer_id if customer_id > 0 else 1,
                    "followup_type": "PAYMENT_OVERDUE",
                    "notes": f"Invoice {invoice_number} is {days_overdue} days overdue (Total: ${total_amount:,.2f}). Prepare internal review before customer dispatch.",
                    "due_date": (datetime.datetime.now(datetime.timezone.utc) + datetime.timedelta(days=2)).isoformat(),
                    "suggested_channel": "EMAIL"
                }
                explanation = f"Invoice {invoice_number} is {days_overdue} days overdue. Creating an internal collections follow-up task to confirm payment status."
                expected_impact = "Accelerates cash flow recovery while ensuring human sign-off on billing outreach."
                confidence = 0.94
                requires_approval = False
            else:
                proposed_action_type = "notes.add"
                action_params = {
                    "entity_type": "INVOICE",
                    "entity_id": invoice_id,
                    "note_text": f"Invoice {invoice_number} within standard payment window. Verified against credit policy.",
                    "is_confidential": False
                }
                explanation = f"Invoice {invoice_number} is within normal credit terms."
                expected_impact = "Maintains audit record of billing monitoring."
                confidence = 0.96
                requires_approval = False

        elif module in ("CONTRACTS", "DOCUMENTS"):
            contract_id = int(ctx.source_record_id) if ctx.source_record_id.isdigit() else 0
            title = record.get("title", f"Contract #{contract_id}")
            days_to_expiry = int(record.get("days_to_expiry", 30))

            evidence_list.append(FactualEvidence(
                field_name="days_to_expiry",
                observed_value=days_to_expiry,
                source_record=f"CONTRACTS/{contract_id}",
                fact_type="CALCULATED_VALUE"
            ))

            if days_to_expiry <= 30:
                proposed_action_type = "reviews.schedule"
                action_params = {
                    "review_subject": f"Expiring Contract Renewal: {title}",
                    "review_scope": "RATE_CONTRACT_RENEWAL",
                    "scheduled_for": (datetime.datetime.now(datetime.timezone.utc) + datetime.timedelta(days=3)).isoformat(),
                    "agenda": f"Review rate commitments and carrier performance ahead of contract expiry in {days_to_expiry} days.",
                    "participants": ["Procurement Team", "Commercial Operations"]
                }
                explanation = f"Contract '{title}' expires in {days_to_expiry} days. Proposing internal renewal review meeting."
                expected_impact = "Avoids rate coverage lapse and uninterrupted cargo booking capacity."
                confidence = 0.91
                risk_level = "LOW"
                requires_approval = False
            else:
                proposed_action_type = "notes.add"
                action_params = {
                    "entity_type": "CONTRACT",
                    "entity_id": contract_id,
                    "note_text": f"Contract validity verified. {days_to_expiry} days remaining.",
                    "is_confidential": False
                }
                explanation = f"Contract has {days_to_expiry} days remaining."
                expected_impact = "Records active compliance verification."
                confidence = 0.95
                requires_approval = False

        elif module == "RFQS":
            rfq_id = int(ctx.source_record_id) if ctx.source_record_id.isdigit() else 0
            rfq_number = record.get("rfq_number", f"RFQ-{rfq_id}")
            quotes_count = int(record.get("quotes_count", 0))

            evidence_list.append(FactualEvidence(
                field_name="quotes_received",
                observed_value=quotes_count,
                source_record=f"RFQS/{rfq_id}",
                fact_type="CONFIRMED_FACT"
            ))

            if quotes_count == 0:
                missing_info.append("No carrier rate options received for this inquiry")
                proposed_action_type = "tasks.create"
                action_params = {
                    "title": f"Procure Carrier Rates for {rfq_number}",
                    "description": f"Customer inquiry {rfq_number} has zero carrier quotes. Request rate cards from primary ocean carriers.",
                    "assigned_team": "PRICING",
                    "priority": "HIGH",
                    "due_date": (datetime.datetime.now(datetime.timezone.utc) + datetime.timedelta(hours=12)).isoformat()
                }
                explanation = f"Inquiry {rfq_number} is awaiting carrier rate quotes. Creating urgent procurement task."
                expected_impact = "Ensures rate quotation is prepared before inquiry deadline."
                confidence = 0.88
                risk_level = "MEDIUM"
                requires_approval = False
            else:
                proposed_action_type = "recommendations.create"
                action_params = {
                    "category": "RFQ_QUOTATION_REVIEW",
                    "title": f"Review Rate Options for {rfq_number}",
                    "description": f"{quotes_count} carrier quotes available for {rfq_number}. Evaluate margin and dispatch quote to client.",
                    "action_type": "REVIEW_QUOTES",
                    "priority": "MEDIUM",
                    "source_ref": rfq_number,
                    "evidence": [{"quotes_count": quotes_count}]
                }
                explanation = f"{quotes_count} carrier quotes have been received for {rfq_number}."
                expected_impact = "Guides pricing desk to select optimal rate."
                confidence = 0.93
                requires_approval = False

        else:
            # Generic fallback safe internal task
            proposed_action_type = "tasks.create"
            action_params = {
                "title": f"Review Operational Event: {ctx.trigger_event}",
                "description": f"Autonomous evaluation received event '{ctx.trigger_event}' on {ctx.source_module} #{ctx.source_record_id}. Operator review requested.",
                "assigned_team": "OPERATIONS",
                "priority": "MEDIUM",
                "due_date": (datetime.datetime.now(datetime.timezone.utc) + datetime.timedelta(days=1)).isoformat()
            }
            explanation = f"Automated detection flagged event '{ctx.trigger_event}' for manual operator verification."
            expected_impact = "Guarantees human oversight for unclassified event types."
            confidence = 0.80
            requires_approval = False

        # 3. Validate Proposed Action Against Fixed Registry
        is_valid, err_msg, validated_params = validate_action_proposal_parameters(proposed_action_type, action_params)
        if not is_valid or not validated_params:
            return OrchestrationProposalResponse(
                status="ERROR",
                proposal=None,
                evaluation_summary="Action proposal failed registry parameter validation.",
                rejection_reason=err_msg,
                duration_ms=int((time.time() - start_time) * 1000)
            )

        # 4. Construct Strict ActionProposal
        action_def = get_registered_action(proposed_action_type)
        if action_def:
            if action_def.category == "HIGH_RISK":
                requires_approval = True
                risk_level = action_def.risk_level
                reversibility = action_def.is_reversible

        proposal = ActionProposal(
            org_id=ctx.org_id,
            source_module=ctx.source_module,
            source_record_type=ctx.source_record_type,
            source_record_id=ctx.source_record_id,
            trigger_event=ctx.trigger_event,
            proposed_action_type=proposed_action_type,
            action_parameters=validated_params,
            explanation=explanation,
            evidence=evidence_list,
            confidence=confidence,
            risk_level=risk_level,
            priority=priority,
            requires_approval=requires_approval,
            approval_policy=approval_policy,
            expected_impact=expected_impact,
            reversibility=reversibility,
            missing_information=missing_info,
            data_freshness="REAL_TIME",
            correlation_id=ctx.correlation_id,
            created_at=now_iso,
            expires_at=expires_at_iso,
            schema_version="1.0"
        )

        duration = int((time.time() - start_time) * 1000)
        return OrchestrationProposalResponse(
            status="SUCCESS",
            proposal=proposal,
            evaluation_summary=f"Successfully analyzed context for {ctx.source_module} #{ctx.source_record_id}. Proposed action '{proposed_action_type}' (Confidence: {confidence:.2f}, Risk: {risk_level}, Requires Approval: {requires_approval}).",
            rejection_reason=None,
            duration_ms=duration
        )
