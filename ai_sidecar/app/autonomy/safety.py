"""
Phase 5 AI Safety & Policy Guardrails for Autonomous Operations.
Enforces untrusted input sanitization, prohibited action filters,
stop conditions, and risk boundaries before returning generated plans to Go.
"""

import re
from typing import List, Dict, Any, Tuple, Optional
from app.autonomy.models import PlanGenerationRequest, PlanStepModel

# Allowed Action Types in LogisticsHQ Action System
ALLOWED_ACTION_TYPES = {
    # Shipments & Operations
    "shipments.get",
    "shipments.update_milestone",
    "shipments.create_exception",
    "shipments.operations_callback",
    "shipments.request_carrier_update",
    "shipments.add_tracking_milestone",
    "shipments.flag_exception",
    "shipments.schedule_dock_appointment",
    "shipments.update_eta",
    # Commercial & CRM
    "rfq.save_draft_quotes",
    "rfq.apply_selected_rate",
    "leads.create_rfq_from_email",
    "leads.convert_lead",
    "leads.send_clarification_email",
    "leads.schedule_followup",
    "leads.assign_representative",
    "leads.update_score",
    "rfq.request_supplier_rate",
    "rfq.generate_draft_quote",
    "rfq.extend_validity",
    # Rates, Contracts & Compliance
    "contracts.ingest_rates",
    "contracts.review_extraction",
    "documents.record_compliance_discrepancies",
    "contracts.schedule_renewal_review",
    "contracts.flag_compliance_deviation",
    "contracts.request_document_resubmission",
    # Finance & Receivables
    "finance.reconcile_invoice",
    "finance.issue_payment_reminder",
    "finance.create_discrepancy_note",
    "finance.schedule_credit_review",
    "finance.flag_overdue",
    # Intelligence, Context & Notifications
    "context.get_context",
    "context.get_insight",
    "context.get_customer_intelligence",
    "context.get_rfq_intelligence",
    "context.get_shipment_intelligence",
    "context.get_invoice_intelligence",
    "context.get_contract_intelligence",
    "context.get_cross_module_insights",
    "notifications.dispatch",
    "notifications.create_task",
    "notifications.escalate",
    "notifications.send_alert",
    "tasks.create_internal_task",
    "tasks.assign_team",
    "recommendations.create",
    "recommendations.update_status",
}

# Strict Prohibited Actions (Never allowed to execute autonomously)
PROHIBITED_ACTION_PATTERNS = [
    re.compile(r"(shell|exec|cmd|bash|powershell|system)", re.IGNORECASE),
    re.compile(r"(raw_sql|drop_table|truncate|delete_from|alter_table)", re.IGNORECASE),
    re.compile(r"(bypass_approval|grant_admin|elevate_privilege)", re.IGNORECASE),
    re.compile(r"(direct_wire_transfer|drain_account|refund_all)", re.IGNORECASE),
    re.compile(r"(send_unreviewed_blast|mass_spam)", re.IGNORECASE),
]

# Prompt injection patterns
PROMPT_INJECTION_PATTERNS = [
    re.compile(r"(ignore\s+(all\s+)?(previous\s+)?instructions)", re.IGNORECASE),
    re.compile(r"(forget\s+(your\s+)?(all\s+)?(rules|instructions|constraints))", re.IGNORECASE),
    re.compile(r"(system\s+override)", re.IGNORECASE),
    re.compile(r"(you\s+are\s+now\s+in\s+developer\s+mode|you\s+are\s+now\s+dan)", re.IGNORECASE),
    re.compile(r"(bypass\s+(all\s+)?(policy|policies|approval|approvals|security|safeguards))", re.IGNORECASE),
    re.compile(r"(auto[-_\s]?resolve\s+(all\s+)?)", re.IGNORECASE),
    re.compile(r"(grant\s+admin|elevate\s+privilege)", re.IGNORECASE),
    re.compile(r"(reveal\s+(the\s+)?system\s+prompt)", re.IGNORECASE),
    re.compile(r"(print\s+(the\s+)?(env|db_password|secret|api_key))", re.IGNORECASE),
]


def detect_prompt_injection(text: str) -> Tuple[bool, Optional[str]]:
    """Scans text for prompt injection attempts."""
    for pattern in PROMPT_INJECTION_PATTERNS:
        if pattern.search(text):
            return True, f"Prompt injection signature detected: '{pattern.pattern}'"
    return False, None


def sanitize_plan_request(req: PlanGenerationRequest) -> PlanGenerationRequest:
    """Validates and sanitizes plan generation inputs."""
    is_inj, reason = detect_prompt_injection(req.goal)
    if is_inj:
        raise ValueError(f"Plan generation rejected due to security policy: {reason}")

    # Check text fields in current state
    for k, v in req.current_state.items():
        if isinstance(v, str):
            is_inj, reason = detect_prompt_injection(v)
            if is_inj:
                raise ValueError(f"State property '{k}' contains malicious injection pattern: {reason}")

    return req


def sanitize_prompt_input(text: str) -> str:
    """Strips control characters and caps length to avoid prompt injection buffer exploits."""
    if not text:
        return ""
    # Strip null bytes and control chars except newlines and tabs
    sanitized = re.sub(r'[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]', '', text)
    return sanitized.strip()[:4000]


def sanitize_untrusted_text(text: str) -> str:
    """Sanitizes untrusted textual input by neutralizing prompt-injection signatures and control characters."""
    if not text:
        return ""
    clean = text
    for pattern in PROMPT_INJECTION_PATTERNS:
        clean = pattern.sub("[REDACTED_SECURITY]", clean)
    clean = re.sub(r'[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]', '', clean)
    return clean.strip()[:4000]


def validate_action_type_safety(action_type: str) -> bool:
    """Checks if an action type is in the allowed registry and not in the prohibited list."""
    if action_type not in ALLOWED_ACTION_TYPES:
        return False
    for pattern in PROHIBITED_ACTION_PATTERNS:
        if pattern.search(action_type):
            return False
    return True


def validate_monetary_amount(amount: float, max_limit: float = 500.0) -> Tuple[bool, Optional[str]]:
    """Checks if monetary action amounts exceed policy ceiling."""
    if amount > max_limit:
        return False, f"Amount ${amount:.2f} exceeds policy ceiling ${max_limit:.2f}"
    return True, None


def validate_steps_safety(steps: List[PlanStepModel]) -> Tuple[bool, List[str]]:
    """Ensures all generated steps conform strictly to the allowed action registry."""
    violations = []
    for step in steps:
        if step.action_type not in ALLOWED_ACTION_TYPES:
            violations.append(f"Step {step.step_id} requested unregistered action type: '{step.action_type}'")

        for pattern in PROHIBITED_ACTION_PATTERNS:
            if pattern.search(step.action_type):
                violations.append(f"Step {step.step_id} matches critical prohibited pattern: '{step.action_type}'")

        amount = step.parameters.get("amount") or step.parameters.get("cost") or step.parameters.get("total_sell_price")
        if amount and isinstance(amount, (int, float)) and amount > 500.0:
            step.requires_approval = True

    return len(violations) == 0, violations


