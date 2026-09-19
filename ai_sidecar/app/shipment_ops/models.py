"""
Shipment Operations Automation and Intelligent Exception Response Models
Phase 3 Task 3.5 for LogisticsHQ

Typed Pydantic domain models for shipment risk analysis, exception classification,
exception prioritization, operational recommendation synthesis, and communication drafting.
"""

from typing import List, Optional, Literal, Dict, Any
from pydantic import BaseModel, Field, field_validator


class EvidenceFact(BaseModel):
    source_module: str = Field(..., description="Originating module (e.g. shipments, milestones, exceptions, tracking, documents)")
    source_entity_id: int = Field(..., description="ID of source business record")
    source_ref: str = Field(..., description="Human-readable business reference code (e.g. SH-101, EX-204, MS-402, MBL-9812)")
    field_name: str = Field(..., description="Specific field evaluated")
    observed_value: str = Field(..., description="Factual value verified in database")
    description: str = Field(..., description="Contextual explanation of verified fact")


class SafetyWarning(BaseModel):
    warning_type: str = Field(..., description="Warning type, e.g. OVERDUE_MILESTONE, STALE_TRACKING, ACTIVE_CRITICAL_EXCEPTION, MISSING_DOCUMENTS")
    message: str = Field(..., description="Actionable explanation for human operator")
    severity: Literal["INFO", "WARNING", "HIGH", "CRITICAL"] = Field(default="WARNING")


class MilestoneFact(BaseModel):
    milestone_id: Optional[int] = None
    milestone_code: str = Field(..., description="BOOKED, CONTAINER_GATED_IN, VESSEL_DEPARTED, ARRIVED, DELIVERED, etc.")
    description: Optional[str] = None
    planned_date: Optional[str] = None
    actual_date: Optional[str] = None
    status: str = Field(default="PLANNED", description="PLANNED or COMPLETED")
    is_overdue: bool = False
    delay_hours: float = 0.0
    location: Optional[str] = None


class ExceptionFact(BaseModel):
    exception_id: Optional[int] = None
    exception_type: str = Field(..., description="SCHEDULE_DELAY, ETD_DELAY, ETA_DELAY, VESSEL_ROLLOVER, PORT_CONGESTION, CUSTOMS_HOLD, DOCUMENT_ISSUE, CARRIER_DELAY, ROUTE_DEVIATION, CONTAINER_ISSUE, OTHER")
    severity: Literal["LOW", "MEDIUM", "HIGH", "CRITICAL"] = "MEDIUM"
    status: Literal["OPEN", "ACKNOWLEDGED", "IN_PROGRESS", "RESOLVED", "DISMISSED"] = "OPEN"
    title: Optional[str] = "Operational Exception"
    description: Optional[str] = None
    resolved: bool = False
    resolution_notes: Optional[str] = None
    created_at: Optional[str] = None


class TrackingFact(BaseModel):
    event_id: str
    source_type: Optional[str] = None
    milestone_code: Optional[str] = None
    location: Optional[str] = None
    description: Optional[str] = None
    event_time: Optional[str] = None


class DocumentFact(BaseModel):
    document_id: Optional[int] = None
    doc_type: str
    file_name: Optional[str] = "unspecified"
    status: str = "PENDING"
    is_verified: bool = False
    is_missing: bool = False


class DeterministicShipmentSignals(BaseModel):
    """
    Authoritative calculations supplied strictly by Go backend.
    Python cannot override these signals.
    """
    has_overdue_milestone: bool = False
    overdue_milestones: Optional[List[str]] = Field(default_factory=list)
    approaching_milestones: Optional[List[str]] = Field(default_factory=list)

    has_active_exception: bool = False
    active_exceptions_count: int = 0
    critical_exceptions_count: int = 0

    is_tracking_stale: bool = False
    hours_since_last_tracking: float = 0.0

    is_delayed: bool = False
    delay_days: float = 0.0

    missing_documents_count: int = 0
    missing_documents: Optional[List[str]] = Field(default_factory=list)

    carrier_response_gap_hours: float = 0.0
    is_carrier_response_overdue: bool = False

    risk_score: float = Field(default=0.0, ge=0.0, le=100.0)
    risk_level: Literal["LOW", "MEDIUM", "HIGH", "CRITICAL"] = "LOW"
    requires_approval: bool = False

    @field_validator("overdue_milestones", "approaching_milestones", "missing_documents", mode="before")
    @classmethod
    def ensure_list_signals(cls, v):
        if v is None:
            return []
        return v


class ShipmentContext(BaseModel):
    """
    Structured shipment and operational context provided exclusively by Go backend.
    """
    org_id: int
    shipment_id: int
    rfq_id: Optional[int] = None
    booking_id: Optional[int] = None
    carrier_scac: str
    carrier_name: Optional[str] = "Commercial Carrier"
    booking_number: Optional[str] = None
    mbl_number: Optional[str] = None
    hbl_number: Optional[str] = None
    container_numbers: Optional[List[str]] = Field(default_factory=list)
    status: str
    origin_port: str
    destination_port: str
    vessel_name: Optional[str] = None
    voyage_number: Optional[str] = None
    etd: Optional[str] = None
    eta: Optional[str] = None
    customer_id: Optional[int] = None
    customer_name: Optional[str] = "Commercial Cargo Customer"
    customer_tier: Optional[str] = "STANDARD"

    milestones: Optional[List[MilestoneFact]] = Field(default_factory=list)
    exceptions: Optional[List[ExceptionFact]] = Field(default_factory=list)
    recent_tracking: Optional[List[TrackingFact]] = Field(default_factory=list)
    documents: Optional[List[DocumentFact]] = Field(default_factory=list)

    correlation_id: str = "corr-default"

    @field_validator("container_numbers", "milestones", "exceptions", "recent_tracking", "documents", mode="before")
    @classmethod
    def ensure_list_context(cls, v):
        if v is None:
            return []
        return v


class PrioritizedExceptionItem(BaseModel):
    exception_id: int
    priority_rank: int
    urgency: Literal["IMMEDIATE", "HIGH", "MEDIUM", "LOW"]
    business_impact: str
    action_deadline: str
    rationale: str
    recommended_action: str


class ExceptionPrioritizationResponse(BaseModel):
    shipment_id: int
    prioritized_exceptions: List[PrioritizedExceptionItem] = Field(default_factory=list)
    recommended_escalations: List[str] = Field(default_factory=list)
    evidence: List[EvidenceFact] = Field(default_factory=list)
    warnings: List[SafetyWarning] = Field(default_factory=list)
    confidence_score: float = Field(default=1.0, ge=0.0, le=1.0)
    correlation_id: str


class OperationalRecommendationItem(BaseModel):
    action_type: str = Field(..., description="shipments.escalate_exception, shipments.request_carrier_followup, shipments.send_customer_update, shipments.create_internal_task, shipments.request_document_review, shipments.acknowledge_exception")
    title: str
    description: str
    target_role: str = Field(default="Operations Specialist")
    priority: Literal["CRITICAL", "HIGH", "MEDIUM", "LOW"] = "MEDIUM"
    requires_approval: bool = False
    suggested_payload: Dict[str, Any] = Field(default_factory=dict)
    rationale: str


class OperationalRecommendationsResponse(BaseModel):
    shipment_id: int
    recommendations: List[OperationalRecommendationItem] = Field(default_factory=list)
    evidence: List[EvidenceFact] = Field(default_factory=list)
    warnings: List[SafetyWarning] = Field(default_factory=list)
    confidence_score: float = Field(default=1.0, ge=0.0, le=1.0)
    correlation_id: str


class ShipmentRiskAnalysisResponse(BaseModel):
    shipment_id: int
    risk_level: Literal["LOW", "MEDIUM", "HIGH", "CRITICAL"]
    risk_score: float
    operational_summary: str
    key_signals: List[str] = Field(default_factory=list)
    likely_root_cause: str
    recommended_next_steps: List[str] = Field(default_factory=list)
    evidence: List[EvidenceFact] = Field(default_factory=list)
    warnings: List[SafetyWarning] = Field(default_factory=list)
    confidence_score: float = Field(default=1.0, ge=0.0, le=1.0)
    correlation_id: str


class CommunicationDraftRequest(BaseModel):
    shipment_id: int
    draft_type: Literal["CARRIER_FOLLOWUP", "CUSTOMER_UPDATE", "INTERNAL_ESCALATION"]
    recipient_name: Optional[str] = None
    recipient_email: Optional[str] = None
    context: ShipmentContext
    signals: DeterministicShipmentSignals
    tone: Optional[str] = "Professional and Proactive"
    user_instructions: Optional[str] = None
    correlation_id: str


class CommunicationDraftResponse(BaseModel):
    shipment_id: int
    draft_type: str
    subject: str
    customer_wording: str
    internal_notes: str
    recipient_preview: Dict[str, str]
    proposed_actions: List[str] = Field(default_factory=list)
    requires_approval: bool = True
    evidence: List[EvidenceFact] = Field(default_factory=list)
    warnings: List[SafetyWarning] = Field(default_factory=list)
    confidence_score: float = Field(default=1.0, ge=0.0, le=1.0)
    correlation_id: str
