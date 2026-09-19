"""
LogisticsHQ Phase 5 Task 5.12: Autonomous Operations Command Center Agent
Provides cross-domain operational reasoning, multi-factor prioritization,
provenance-preserving fact vs. prediction separation, and concise business explanations.
"""

import uuid
import re
from typing import Dict, Any, List, Tuple
from app.autonomy.models import (
    CommandCenterPrioritizeRequest,
    CommandCenterPrioritizeResponse,
    CommandCenterPrioritizedItem,
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


class CommandCenterAgent:
    """
    Coordinates operational intelligence for the Autonomous Operations Command Center:
    Prioritizes issues across safety/compliance, operational disruption, customer commitments,
    financial risk, approval deadlines, and workflow stalls.
    """

    def prioritize_operations(self, req: CommandCenterPrioritizeRequest) -> CommandCenterPrioritizeResponse:
        correlation_id = req.correlation_id or f"cc-{uuid.uuid4().hex[:8]}"
        prioritized: List[CommandCenterPrioritizedItem] = []

        for idx, item in enumerate(req.items or []):
            item_id = str(item.get("id") or f"item-{idx+1}")
            raw_title = str(item.get("title") or item.get("issue") or "Operational Issue")
            sanitized_title, _ = sanitize_text(raw_title)

            raw_notes = str(item.get("notes") or item.get("description") or "")
            sanitized_notes, _ = sanitize_text(raw_notes)

            severity = str(item.get("severity") or "MEDIUM").upper()
            entity_type = str(item.get("entity_type") or "OPERATIONS").upper()
            entity_id = str(item.get("entity_id") or "")
            module = str(item.get("module") or entity_type).lower()
            category = str(item.get("category") or "").upper()
            deadline = item.get("deadline")
            owner = str(item.get("owner") or "Operations Team")

            # Multi-factor priority score calculation (Section 23)
            # 1. Critical safety/compliance: 90 - 100
            # 2. Critical operational disruption: 80 - 89
            # 3. Severe customer impact: 70 - 79
            # 4. Major financial impact: 60 - 69
            # 5. Approval deadlines: 50 - 59
            # 6. High-risk exceptions: 40 - 49
            # 7. Stalled workflows: 30 - 39
            # 8. Normal operational items: 10 - 29

            priority_score = 20.0
            priority_tier = "NORMAL_OPERATIONAL"
            urgency = "LOW"

            if category in ["COMPLIANCE", "SAFETY"] or "compliance" in module or severity == "CRITICAL" and "compliance" in sanitized_title.lower():
                priority_tier = "CRITICAL_SAFETY_COMPLIANCE"
                priority_score = 95.0 if severity == "CRITICAL" else 90.0
                urgency = "IMMEDIATE"
            elif severity == "CRITICAL" and (category == "DISRUPTION" or "shipment" in module or "operation" in module):
                priority_tier = "CRITICAL_OPERATIONAL"
                priority_score = 85.0
                urgency = "IMMEDIATE"
            elif category == "CUSTOMER" or "customer" in module or "sla" in sanitized_title.lower():
                priority_tier = "SEVERE_CUSTOMER"
                priority_score = 75.0 if severity in ["CRITICAL", "HIGH"] else 70.0
                urgency = "HIGH"
            elif category in ["FINANCE", "INVOICE", "PAYMENT"] or "finance" in module:
                priority_tier = "MAJOR_FINANCIAL"
                priority_score = 65.0 if severity in ["CRITICAL", "HIGH"] else 60.0
                urgency = "HIGH" if severity == "CRITICAL" else "MEDIUM"
            elif category in ["APPROVAL", "DECISION"] or "approval" in sanitized_title.lower():
                priority_tier = "APPROVAL_DEADLINE"
                priority_score = 55.0
                urgency = "HIGH" if deadline else "MEDIUM"
            elif severity in ["CRITICAL", "HIGH"] or category == "EXCEPTION":
                priority_tier = "HIGH_RISK_EXCEPTION"
                priority_score = 45.0
                urgency = "HIGH" if severity == "CRITICAL" else "MEDIUM"
            elif category in ["STALLED", "STALE", "BLOCKED"] or "stalled" in sanitized_title.lower():
                priority_tier = "STALLED_WORKFLOW"
                priority_score = 35.0
                urgency = "MEDIUM"
            else:
                priority_tier = "NORMAL_OPERATIONAL"
                priority_score = 20.0
                urgency = "LOW"

            # Urgency booster if deadline is imminent
            if deadline and urgency in ["MEDIUM", "LOW"]:
                urgency = "HIGH"
                priority_score += 4.0

            # Separation of facts vs. predictions vs. recommendations (Section 29)
            actual_facts = str(item.get("actual_facts") or item.get("current_state") or f"Authoritative state recorded for {entity_type} {entity_id}.")
            predicted_impact = str(item.get("predicted_impact") or item.get("impact") or "AI prediction: Operational delay or escalation if unaddressed.")
            recommended_action = str(item.get("recommended_action") or item.get("required_action") or "Review item and authorize planned recovery steps.")

            # Concise business explanation (Section 30: no chain-of-thought)
            why_flagged = (
                f"Flagged under {priority_tier.replace('_', ' ').lower()}: "
                f"{sanitized_title}. {sanitized_notes[:140]}"
            ).strip()
            if not why_flagged.endswith("."):
                why_flagged += "."

            prioritized.append(
                CommandCenterPrioritizedItem(
                    id=item_id,
                    priority_rank=0,  # assigned after sorting
                    priority_score=round(priority_score, 1),
                    priority_tier=priority_tier,
                    severity=severity,
                    entity_type=entity_type,
                    entity_id=entity_id,
                    title=sanitized_title,
                    issue_summary=sanitized_notes[:200] if sanitized_notes else sanitized_title,
                    why_flagged=why_flagged,
                    actual_facts=actual_facts,
                    predicted_impact=predicted_impact,
                    recommended_action=recommended_action,
                    owner=owner,
                    deadline=deadline,
                    urgency=urgency,
                    source=str(item.get("source") or "AUTONOMY_ENGINE"),
                    requires_human=bool(item.get("requires_human", True)),
                )
            )

        # Sort descending by priority score
        prioritized.sort(key=lambda x: x.priority_score, reverse=True)

        critical_count = 0
        high_count = 0
        for rank, p in enumerate(prioritized, 1):
            p.priority_rank = rank
            if p.severity == "CRITICAL" or p.priority_tier == "CRITICAL_SAFETY_COMPLIANCE":
                critical_count += 1
            elif p.severity == "HIGH" or p.urgency == "HIGH":
                high_count += 1

        summary = (
            f"Evaluated {len(prioritized)} operational items across business modules. "
            f"{critical_count} critical safety/disruption item(s) requiring immediate intervention, "
            f"{high_count} high-priority item(s) awaiting action."
        )

        return CommandCenterPrioritizeResponse(
            items=prioritized,
            critical_count=critical_count,
            high_count=high_count,
            executive_summary=summary,
            correlation_id=correlation_id,
        )


command_center_agent = CommandCenterAgent()
