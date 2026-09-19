"""
Pydantic schemas and typed definitions for the Python AI Action Orchestrator.
Adheres strictly to LogisticsHQ Phase 3 Task 3.2 requirements.
"""

from typing import List, Optional, Dict, Any, Union, Literal
from pydantic import BaseModel, Field, field_validator
import datetime
import uuid

# Allowed risk levels
RiskLevel = Literal["LOW", "MEDIUM", "HIGH", "CRITICAL"]

# Allowed priorities
PriorityLevel = Literal["LOW", "MEDIUM", "HIGH", "URGENT"]

# Allowed approval policies
ApprovalPolicy = Literal["ALWAYS_REQUIRE_APPROVAL", "THRESHOLD_BASED", "AUTOMATIC_FOR_SAFE_ACTIONS"]

# Allowed reversibility ratings
ReversibilityType = Literal["REVERSIBLE", "PARTIALLY_REVERSIBLE", "IRREVERSIBLE"]

# Data freshness ratings
DataFreshnessType = Literal["REAL_TIME", "RECENT", "STALE", "HISTORICAL"]


# ── Action Parameter Models for Fixed Allowlist ──────────────────────────────

class CreateInternalTaskParams(BaseModel):
    title: str = Field(..., min_length=3, max_length=255)
    description: str = Field(..., max_length=2000)
    assigned_to_user_id: Optional[int] = None
    assigned_team: Optional[str] = Field("OPERATIONS", max_length=64)
    priority: PriorityLevel = "MEDIUM"
    due_date: Optional[str] = None  # ISO8601 string


class AssignInternalTaskParams(BaseModel):
    task_id: int
    assignee_user_id: int
    reassignment_reason: Optional[str] = Field(None, max_length=500)


class CreateRecommendationParams(BaseModel):
    category: str = Field(..., max_length=64)
    title: str = Field(..., min_length=3, max_length=255)
    description: str = Field(..., max_length=2000)
    action_type: str = Field(..., max_length=64)
    priority: PriorityLevel = "MEDIUM"
    customer_id: Optional[int] = None
    shipment_id: Optional[int] = None
    source_ref: Optional[str] = None
    evidence: List[Dict[str, Any]] = Field(default_factory=list)


class CreateNotificationParams(BaseModel):
    recipient_user_id: Optional[int] = None
    recipient_role: Optional[str] = None
    title: str = Field(..., min_length=3, max_length=255)
    message: str = Field(..., max_length=2000)
    severity: Literal["INFO", "WARNING", "CRITICAL"] = "INFO"
    link_url: Optional[str] = None


class UpdateRecommendationStatusParams(BaseModel):
    recommendation_id: int
    target_status: Literal["ACCEPTED", "DISMISSED", "IN_PROGRESS", "COMPLETED"]
    notes: Optional[str] = None


class RequestApprovalParams(BaseModel):
    title: str = Field(..., min_length=3, max_length=255)
    category: Literal["OPERATIONS", "COMMERCIAL", "DOCUMENTS", "FINANCE"]
    action_name: str
    target_payload: Dict[str, Any]
    required_role: str = "MANAGER"
    reason: str = Field(..., max_length=1000)


class CreateInternalFollowupParams(BaseModel):
    customer_id: int
    followup_type: Literal["PAYMENT_OVERDUE", "EXPIRING_CONTRACT", "QUOTE_PENDING", "SERVICE_FEEDBACK"]
    notes: str = Field(..., max_length=2000)
    due_date: Optional[str] = None
    suggested_channel: Literal["EMAIL", "PHONE", "PORTAL"] = "EMAIL"


class AddInternalNoteParams(BaseModel):
    entity_type: Literal["SHIPMENT", "INVOICE", "CONTRACT", "RFQ", "CUSTOMER"]
    entity_id: int
    note_text: str = Field(..., min_length=2, max_length=2000)
    is_confidential: bool = False


class ScheduleInternalReviewParams(BaseModel):
    review_subject: str = Field(..., min_length=3, max_length=255)
    review_scope: str = Field("OPERATIONAL_RISK", max_length=64)
    scheduled_for: str  # ISO8601
    agenda: str = Field(..., max_length=2000)
    participants: List[str] = Field(default_factory=list)


class MarkAutomationAttentionParams(BaseModel):
    automation_id: int
    execution_id: Optional[int] = None
    attention_reason: str = Field(..., min_length=5, max_length=1000)
    severity: Literal["LOW", "MEDIUM", "HIGH", "CRITICAL"] = "MEDIUM"


# High-Risk Parameter Models (Strictly require Human-in-the-Loop approval)
class ShipmentStatusUpdateParams(BaseModel):
    shipment_id: int
    target_status: str
    reason: str


class ShipmentEtaUpdateParams(BaseModel):
    shipment_id: int
    new_planned_eta: str
    reason: str
    delay_hours: float


class InvoiceAdjustAmountParams(BaseModel):
    invoice_id: int
    original_amount: float
    adjusted_amount: float
    adjustment_reason: str


class RfqSendQuoteParams(BaseModel):
    rfq_id: int
    quote_id: int
    customer_email: str
    total_sell_price: float
    currency: str = "USD"


# ── Factual Evidence Item ───────────────────────────────────────────────────

class FactualEvidence(BaseModel):
    field_name: str
    observed_value: Any
    baseline_value: Optional[Any] = None
    source_record: Optional[str] = None
    timestamp: Optional[str] = None
    fact_type: Literal["CONFIRMED_FACT", "CALCULATED_VALUE", "INFERENCE", "THRESHOLD_BREACH"] = "CONFIRMED_FACT"


# ── Structured AI Action Proposal (Section 3 Requirement) ───────────────────

class ActionProposal(BaseModel):
    proposal_id: str = Field(default_factory=lambda: f"prop-{uuid.uuid4().hex[:12]}")
    org_id: int
    source_module: str = Field(..., max_length=64)
    source_record_type: str = Field(..., max_length=64)
    source_record_id: str = Field(..., max_length=128)
    trigger_event: str = Field(..., max_length=128)
    proposed_action_type: str = Field(..., max_length=128)
    action_parameters: Dict[str, Any]
    explanation: str = Field(..., min_length=5, max_length=3000)
    evidence: List[FactualEvidence] = Field(default_factory=list)
    confidence: float = Field(..., ge=0.0, le=1.0)
    risk_level: RiskLevel = "LOW"
    priority: PriorityLevel = "MEDIUM"
    requires_approval: bool = True
    approval_policy: ApprovalPolicy = "ALWAYS_REQUIRE_APPROVAL"
    expected_impact: str = Field(..., min_length=3, max_length=1000)
    reversibility: ReversibilityType = "REVERSIBLE"
    missing_information: List[str] = Field(default_factory=list)
    data_freshness: DataFreshnessType = "REAL_TIME"
    correlation_id: str
    created_at: str = Field(default_factory=lambda: datetime.datetime.now(datetime.timezone.utc).isoformat())
    expires_at: Optional[str] = None
    schema_version: str = "1.0"

    @field_validator("confidence")
    @classmethod
    def validate_confidence(cls, v: float) -> float:
        if v < 0.0 or v > 1.0:
            raise ValueError("Confidence must be between 0.0 and 1.0")
        return round(v, 4)


# ── Incoming Context from Go Backend ────────────────────────────────────────

class OperationalSignal(BaseModel):
    signal_type: str
    signal_value: Any
    severity: str = "MEDIUM"
    detected_at: Optional[str] = None


class OrchestrationContextInput(BaseModel):
    org_id: int
    user_id: int = 0
    source_module: str
    source_record_type: str
    source_record_id: str
    trigger_event: str
    record_data: Dict[str, Any] = Field(default_factory=dict)
    operational_signals: List[OperationalSignal] = Field(default_factory=list)
    correlation_id: str
    preferred_action_types: Optional[List[str]] = None


class OrchestrationProposalResponse(BaseModel):
    status: Literal["SUCCESS", "NO_ACTION_REQUIRED", "REJECTED_UNSAFE", "ERROR"]
    proposal: Optional[ActionProposal] = None
    evaluation_summary: str
    rejection_reason: Optional[str] = None
    duration_ms: int = 0
