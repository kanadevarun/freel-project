"""
Fixed Action Registry for the Python AI Action Orchestrator.
Defines registered action types, risk classifications, approval rules,
and parameter validation mappings.
"""

from typing import Dict, Any, Tuple, Optional, Type
from pydantic import BaseModel, ValidationError

from app.orchestrator.models import (
    CreateInternalTaskParams,
    AssignInternalTaskParams,
    CreateRecommendationParams,
    CreateNotificationParams,
    UpdateRecommendationStatusParams,
    RequestApprovalParams,
    CreateInternalFollowupParams,
    AddInternalNoteParams,
    ScheduleInternalReviewParams,
    MarkAutomationAttentionParams,
    ShipmentStatusUpdateParams,
    ShipmentEtaUpdateParams,
    InvoiceAdjustAmountParams,
    RfqSendQuoteParams,
    RiskLevel,
    ReversibilityType
)

class RegisteredActionDef:
    def __init__(
        self,
        name: str,
        category: str,  # "SAFE_INTERNAL" or "HIGH_RISK"
        risk_level: RiskLevel,
        requires_approval_by_default: bool,
        is_reversible: ReversibilityType,
        param_model: Type[BaseModel],
        description: str
    ):
        self.name = name
        self.category = category
        self.risk_level = risk_level
        self.requires_approval_by_default = requires_approval_by_default
        self.is_reversible = is_reversible
        self.param_model = param_model
        self.description = description


# ── Fixed Allowlist of Registered Actions ─────────────────────────────────────

ACTION_REGISTRY: Dict[str, RegisteredActionDef] = {
    # Safe Internal Actions (Low/Medium Risk)
    "tasks.create": RegisteredActionDef(
        name="tasks.create",
        category="SAFE_INTERNAL",
        risk_level="LOW",
        requires_approval_by_default=False,
        is_reversible="REVERSIBLE",
        param_model=CreateInternalTaskParams,
        description="Creates an internal operational task for team review and execution."
    ),
    "tasks.assign": RegisteredActionDef(
        name="tasks.assign",
        category="SAFE_INTERNAL",
        risk_level="LOW",
        requires_approval_by_default=False,
        is_reversible="REVERSIBLE",
        param_model=AssignInternalTaskParams,
        description="Assigns an existing internal task to a specific team member."
    ),
    "recommendations.create": RegisteredActionDef(
        name="recommendations.create",
        category="SAFE_INTERNAL",
        risk_level="LOW",
        requires_approval_by_default=False,
        is_reversible="REVERSIBLE",
        param_model=CreateRecommendationParams,
        description="Creates a grounded recommendation in the AI Recommendation Center."
    ),
    "notifications.create": RegisteredActionDef(
        name="notifications.create",
        category="SAFE_INTERNAL",
        risk_level="LOW",
        requires_approval_by_default=False,
        is_reversible="REVERSIBLE",
        param_model=CreateNotificationParams,
        description="Sends an in-app operational notification to relevant operators."
    ),
    "recommendations.update_status": RegisteredActionDef(
        name="recommendations.update_status",
        category="SAFE_INTERNAL",
        risk_level="LOW",
        requires_approval_by_default=False,
        is_reversible="REVERSIBLE",
        param_model=UpdateRecommendationStatusParams,
        description="Updates the lifecycle status of an AI recommendation."
    ),
    "approvals.request": RegisteredActionDef(
        name="approvals.request",
        category="SAFE_INTERNAL",
        risk_level="MEDIUM",
        requires_approval_by_default=True,
        is_reversible="REVERSIBLE",
        param_model=RequestApprovalParams,
        description="Submits an operational decision to the Human-in-the-Loop Approval Queue."
    ),
    "followups.create": RegisteredActionDef(
        name="followups.create",
        category="SAFE_INTERNAL",
        risk_level="LOW",
        requires_approval_by_default=False,
        is_reversible="REVERSIBLE",
        param_model=CreateInternalFollowupParams,
        description="Logs an internal customer follow-up schedule without external messaging."
    ),
    "notes.add": RegisteredActionDef(
        name="notes.add",
        category="SAFE_INTERNAL",
        risk_level="LOW",
        requires_approval_by_default=False,
        is_reversible="REVERSIBLE",
        param_model=AddInternalNoteParams,
        description="Appends an internal operational note to a business entity."
    ),
    "reviews.schedule": RegisteredActionDef(
        name="reviews.schedule",
        category="SAFE_INTERNAL",
        risk_level="LOW",
        requires_approval_by_default=False,
        is_reversible="REVERSIBLE",
        param_model=ScheduleInternalReviewParams,
        description="Schedules an internal risk review meeting for a critical incident."
    ),
    "automations.flag_attention": RegisteredActionDef(
        name="automations.flag_attention",
        category="SAFE_INTERNAL",
        risk_level="MEDIUM",
        requires_approval_by_default=False,
        is_reversible="REVERSIBLE",
        param_model=MarkAutomationAttentionParams,
        description="Marks an automation workflow execution as requiring operator attention."
    ),

    # High-Risk Actions (Mandatory Human Approval, Never Executed Autonomously)
    "shipments.update_status": RegisteredActionDef(
        name="shipments.update_status",
        category="HIGH_RISK",
        risk_level="HIGH",
        requires_approval_by_default=True,
        is_reversible="PARTIALLY_REVERSIBLE",
        param_model=ShipmentStatusUpdateParams,
        description="Updates the operational milestone or status of a shipment (Requires Approval)."
    ),
    "shipments.update_eta": RegisteredActionDef(
        name="shipments.update_eta",
        category="HIGH_RISK",
        risk_level="HIGH",
        requires_approval_by_default=True,
        is_reversible="REVERSIBLE",
        param_model=ShipmentEtaUpdateParams,
        description="Updates the planned ETA date on a live shipment (Requires Approval)."
    ),
    "invoices.adjust_amount": RegisteredActionDef(
        name="invoices.adjust_amount",
        category="HIGH_RISK",
        risk_level="CRITICAL",
        requires_approval_by_default=True,
        is_reversible="IRREVERSIBLE",
        param_model=InvoiceAdjustAmountParams,
        description="Adjusts commercial financial amounts or credit limits on an invoice (Requires Approval)."
    ),
    "rfq.send_quote": RegisteredActionDef(
        name="rfq.send_quote",
        category="HIGH_RISK",
        risk_level="HIGH",
        requires_approval_by_default=True,
        is_reversible="IRREVERSIBLE",
        param_model=RfqSendQuoteParams,
        description="Dispatches formal quotation rates to a customer (Requires Approval)."
    ),
}


def get_registered_action(action_name: str) -> Optional[RegisteredActionDef]:
    return ACTION_REGISTRY.get(action_name)


def validate_action_proposal_parameters(action_name: str, raw_params: Dict[str, Any]) -> Tuple[bool, Optional[str], Optional[Dict[str, Any]]]:
    """
    Validates that the proposed action name is in the fixed allowlist,
    and that the parameters match the strict Pydantic model.
    """
    action_def = get_registered_action(action_name)
    if not action_def:
        return False, f"Action '{action_name}' is not registered in the fixed action allowlist.", None

    try:
        validated_model = action_def.param_model(**raw_params)
        return True, None, validated_model.model_dump()
    except ValidationError as e:
        return False, f"Parameter validation failed for action '{action_name}': {e.errors()}", None
