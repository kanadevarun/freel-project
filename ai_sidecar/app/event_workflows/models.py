"""
Pydantic schemas for Phase 3 Task 3.8: Event-Driven AI Workflows and Cross-Module Automation.
ALL AI DATA CONTRACTS ARE STRICTLY DEFINED AND VALIDATED IN PYTHON ONLY.
"""

from typing import List, Dict, Any, Optional
from pydantic import BaseModel, Field


class EvidenceFact(BaseModel):
    source_module: str
    record_type: str
    record_id: str
    fact_key: str
    fact_value: Any


class WorkflowRecommendation(BaseModel):
    recommendation_type: str = Field(description="CATEGORY e.g. OPERATIONS_ALERT, INVOICE_REMINDER, CONTRACT_RENEWAL")
    title: str
    description: str
    action_name: str = Field(description="Registered Centralized Action System name")
    action_payload: Dict[str, Any] = Field(default_factory=dict)
    requires_approval: bool = Field(default=False)
    risk_level: str = Field(default="LOW", description="LOW | MEDIUM | HIGH | CRITICAL")
    target_module: str = Field(description="Target business domain e.g. SHIPMENTS, FINANCE, CONTRACTS, LEADS, RFQ")


class EventWorkflowRequest(BaseModel):
    event_id: int
    org_id: int
    event_type: str = Field(description="Domain event type e.g. shipment.milestone_missed, invoice.overdue, lead.created")
    source_module: str = Field(description="LEADS | RFQ | QUOTATIONS | BOOKINGS | SHIPMENTS | FINANCE | CONTRACTS | DOCUMENTS")
    source_record_type: str = Field(description="LEAD | RFQ | QUOTATION | BOOKING | SHIPMENT | INVOICE | CONTRACT | DOCUMENT")
    source_record_id: str
    event_facts: Dict[str, Any] = Field(default_factory=dict)
    cross_module_context: Dict[str, Any] = Field(default_factory=dict)
    correlation_id: str


class EventAnalysisResponse(BaseModel):
    workflow_type: str = Field(description="Mapped workflow e.g. LEAD_FOLLOWUP, SHIPMENT_EXCEPTION_RESPONSE, INVOICE_COLLECTION_ESCALATION")
    urgency: str = Field(description="LOW | MEDIUM | HIGH | CRITICAL")
    significance_score: float = Field(ge=0.0, le=100.0)
    summary: str
    cross_module_insights: List[str] = Field(default_factory=list)
    recommended_actions: List[WorkflowRecommendation] = Field(default_factory=list)
    missing_information: List[str] = Field(default_factory=list)
    confidence_score: float = Field(ge=0.0, le=1.0, default=0.92)
    correlation_id: str


class WorkflowDraftRequest(BaseModel):
    event_type: str
    source_record_id: str
    recipient_name: str
    recipient_email: str
    recipient_role: str = Field(default="CUSTOMER", description="CUSTOMER | CARRIER | INTERNAL_OPS")
    topic: str
    tone: str = Field(default="PROFESSIONAL", description="PROFESSIONAL | URGENT | FORMAL")
    event_facts: Dict[str, Any] = Field(default_factory=dict)
    cross_module_context: Dict[str, Any] = Field(default_factory=dict)
    custom_instructions: Optional[str] = None
    correlation_id: str


class WorkflowDraftResponse(BaseModel):
    subject: str
    message_body: str
    recommended_action_name: str
    confidence_score: float = Field(ge=0.0, le=1.0, default=0.92)
    correlation_id: str
