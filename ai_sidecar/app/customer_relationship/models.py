"""
Customer Relationship Automation and Intelligent Follow-Up Models
Phase 3 Task 3.3 for LogisticsHQ

Typed Pydantic domain models for customer follow-up context,
health & prioritization signals, grounded evidence evaluation,
and communication drafting.
"""

from typing import List, Optional, Literal, Dict, Any
from pydantic import BaseModel, Field


class EvidenceFact(BaseModel):
    source_module: str = Field(..., description="Originating module, e.g. customers, rfqs, shipments, invoices, contracts")
    source_entity_id: int = Field(..., description="ID of source business record")
    source_ref: str = Field(..., description="Human-readable business reference code (e.g. CUST-10, INV-2026-003)")
    field_name: str = Field(..., description="Specific field evaluated")
    observed_value: str = Field(..., description="Factual value verified in database")
    description: str = Field(..., description="Contextual explanation of verified fact")


class SafetyWarning(BaseModel):
    warning_type: str = Field(..., description="Type of risk or missing information, e.g. MISSING_CONTACT_EMAIL, UNVERIFIED_CARRIER_STATUS")
    message: str = Field(..., description="Actionable warning explanation for human operator")
    severity: Literal["INFO", "WARNING", "CRITICAL"] = Field(default="WARNING")


class CustomerFollowupContext(BaseModel):
    """
    Sanitized, structured customer and relationship context supplied exclusively by Go backend.
    """
    org_id: int
    customer_id: int
    customer_name: str
    customer_code: Optional[str] = None
    account_owner_id: Optional[int] = None
    account_owner_name: Optional[str] = None
    primary_contact_name: Optional[str] = None
    primary_contact_email: Optional[str] = None
    
    # Deterministic activity metrics
    days_inactive: int = 0
    health_score: int = 50
    credit_status: Optional[str] = "NORMAL"
    outstanding_balance: float = 0.0
    overdue_balance: float = 0.0
    
    # Pipeline & operations counts
    rfq_count: int = 0
    open_quote_count: int = 0
    booking_count: int = 0
    active_shipment_count: int = 0
    delayed_shipment_count: int = 0
    unresolved_exception_count: int = 0
    contract_count: int = 0
    expiring_contract_count: int = 0
    missing_doc_count: int = 0
    
    recent_activity_date: Optional[str] = None
    source_module: Optional[str] = "customers"
    source_record_id: Optional[int] = None
    source_reference: Optional[str] = None
    correlation_id: Optional[str] = None


class CustomerPriorityResult(BaseModel):
    priority_score: int = Field(..., ge=0, le=100, description="Deterministic priority score (0-100)")
    priority_level: Literal["CRITICAL", "HIGH", "MEDIUM", "LOW"] = Field(..., description="Categorical priority level")
    risk_level: Literal["CRITICAL", "HIGH", "MEDIUM", "LOW"] = Field(..., description="Commercial/operational risk rating")
    primary_reason: str = Field(..., description="Clear transparent explanation for priority score")
    contributing_factors: List[str] = Field(default_factory=list, description="Explicit bulleted drivers of priority")
    confidence: float = Field(..., ge=0.0, le=1.0, description="Confidence in recommendation logic based on verified evidence")
    evidence: List[EvidenceFact] = Field(default_factory=list, description="Verified facts from MariaDB")
    correlation_id: Optional[str] = None


class FollowupRecommendationOutput(BaseModel):
    customer_id: int
    customer_name: str
    source_module: str
    source_record_id: int
    source_reference: str
    trigger_reason: str
    category: str = "CUSTOMER_FOLLOWUP"
    priority: Literal["CRITICAL", "HIGH", "MEDIUM", "LOW"]
    risk_level: Literal["CRITICAL", "HIGH", "MEDIUM", "LOW"]
    confidence: Literal["VERY_HIGH", "HIGH", "MEDIUM", "LOW"] = "HIGH"
    confidence_score: float = Field(..., ge=0.0, le=1.0)
    evidence: List[EvidenceFact] = Field(default_factory=list)
    recommended_action: str
    action_type: str
    followup_type: str
    suggested_owner_id: Optional[int] = None
    suggested_owner_name: Optional[str] = None
    requires_approval: bool = False
    correlation_id: str
    dedup_key: str


class CustomerEvaluationResponse(BaseModel):
    status: str = "SUCCESS"
    customer_id: int
    customer_name: str
    priority: CustomerPriorityResult
    recommendations: List[FollowupRecommendationOutput] = Field(default_factory=list)
    summary: str
    correlation_id: str


class DraftMessageRequest(BaseModel):
    """
    Explicit request from human operator to draft customer communication.
    Must contain only verified facts supplied by Go backend.
    """
    org_id: int
    customer_id: int
    customer_name: str
    primary_contact_name: Optional[str] = None
    primary_contact_email: Optional[str] = None
    operator_name: str = "LogisticsHQ Customer Success"
    draft_type: str = Field(
        ..., 
        description="Type of draft: CUSTOMER_FOLLOWUP_EMAIL, RFQ_CLARIFICATION_REQUEST, QUOTATION_FOLLOWUP, SHIPMENT_DELAY_EXPLANATION, SHIPMENT_STATUS_UPDATE, INVOICE_REMINDER_DRAFT, CONTRACT_RENEWAL_REMINDER, MISSING_DOCUMENT_REQUEST, INTERNAL_ESCALATION_NOTE"
    )
    source_module: str
    source_record_id: int
    source_reference: str
    description: str
    recommended_action: str
    evidence: List[EvidenceFact] = Field(default_factory=list)
    user_tone_preference: Optional[str] = "Professional and collaborative"
    correlation_id: str


class DraftMessageResponse(BaseModel):
    status: str = "SUCCESS"
    subject: str
    body: str
    suggested_recipients: List[str] = Field(default_factory=list)
    evidence: List[EvidenceFact] = Field(default_factory=list)
    missing_information: List[str] = Field(default_factory=list)
    safety_warnings: List[SafetyWarning] = Field(default_factory=list)
    confidence: float = Field(..., ge=0.0, le=1.0)
    requires_approval: bool = True  # External communication always requires approval
    correlation_id: str
