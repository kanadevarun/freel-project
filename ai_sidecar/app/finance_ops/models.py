"""
Finance and Collections Automation Models
Phase 3 Task 3.6 for LogisticsHQ

Typed Pydantic domain models for receivables risk analysis, collection prioritization,
customer payment behavior analysis, operational recommendations, and communication drafting.
"""

from typing import List, Optional, Literal, Dict, Any
from pydantic import BaseModel, Field, field_validator, AliasChoices


class EvidenceFact(BaseModel):
    source_module: str = Field(..., description="Originating module (e.g. invoices, payments, customers, shipments)")
    source_entity_id: int = Field(..., description="ID of source business record")
    source_ref: str = Field(..., description="Human-readable business reference code (e.g. INV-2026-0454, CUST-102)")
    field_name: str = Field(..., description="Specific field evaluated")
    observed_value: str = Field(..., description="Factual value verified in database")
    description: str = Field(..., description="Contextual explanation of verified fact")


class InvoiceContext(BaseModel):
    """
    Authoritative invoice and customer data supplied strictly by Go backend.
    """
    org_id: int
    invoice_id: int
    invoice_number: str
    customer_id: int
    customer_name: str
    customer_email: Optional[str] = None
    customer_country: Optional[str] = None
    shipment_id: Optional[int] = None
    shipment_number: Optional[str] = None
    booking_id: Optional[int] = None
    booking_number: Optional[str] = None
    invoice_date: Optional[str] = "2026-01-01"
    due_date: Optional[str] = "2026-02-01"
    currency: str = "USD"
    total_amount: float = Field(default=0.0, validation_alias=AliasChoices("total_amount", "invoice_amount"))
    paid_amount: float = 0.0
    balance_due: float = Field(default=0.0, validation_alias=AliasChoices("balance_due", "outstanding_amount"))
    status: str = "OVERDUE"

    line_items: Optional[List[Dict[str, Any]]] = Field(default_factory=list)
    payment_history: Optional[List[Dict[str, Any]]] = Field(default_factory=list)

    customer_credit_limit: Optional[float] = 0.0
    customer_payment_terms: Optional[str] = "Net 30"
    customer_open_invoices_count: Optional[int] = 1
    customer_total_outstanding: Optional[float] = 0.0
    linked_shipment_status: Optional[str] = None
    has_linked_exception: Optional[bool] = False
    correlation_id: str = "corr-default"

    @field_validator("line_items", "payment_history", mode="before")
    @classmethod
    def ensure_list_fields(cls, v):
        if v is None:
            return []
        return v


class DeterministicFinanceSignals(BaseModel):
    """
    Authoritative calculations supplied strictly by Go backend.
    Python cannot override these signals.
    """
    is_overdue: bool = False
    days_overdue: int = 0
    aging_bucket: str = "CURRENT"  # CURRENT, 1_30, 31_60, 61_90, 90_PLUS

    is_approaching_due_date: bool = False
    days_until_due: int = 0

    is_high_value: bool = False
    is_partially_paid: bool = False
    has_multiple_overdue: bool = False
    customer_overdue_count: int = 0
    customer_total_overdue_balance: float = 0.0
    is_credit_limit_exceeded: bool = False
    is_disputed: bool = False
    has_linked_exception: bool = False

    risk_score: float = Field(default=0.0, ge=0.0, le=100.0)
    risk_level: Literal["LOW", "MEDIUM", "HIGH", "CRITICAL"] = "LOW"
    requires_approval: bool = True


class ReceivablesRiskAnalysisResponse(BaseModel):
    invoice_id: int
    customer_id: int
    risk_level: Literal["LOW", "MEDIUM", "HIGH", "CRITICAL"]
    risk_score: float
    days_overdue: int
    aging_bucket: str
    outstanding_amount: float
    currency: str
    receivables_summary: str
    key_risks: List[str] = Field(default_factory=list)
    recommended_next_steps: List[str] = Field(default_factory=list)
    evidence: List[EvidenceFact] = Field(default_factory=list)
    confidence_score: float = Field(default=0.85, ge=0.0, le=1.0)
    correlation_id: str


class CollectionPrioritizedItem(BaseModel):
    invoice_id: int
    invoice_number: str
    customer_id: int
    customer_name: str
    priority_rank: int
    urgency: Literal["IMMEDIATE", "HIGH", "MEDIUM", "LOW"]
    outstanding_amount: float
    currency: str
    days_overdue: int
    aging_bucket: str
    business_impact: str
    recommended_action: str
    action_deadline: str
    rationale: str


class CollectionPrioritizationResponse(BaseModel):
    prioritized_items: List[CollectionPrioritizedItem] = Field(default_factory=list)
    summary: str
    total_at_risk_amount: float
    currency: str = "USD"


class CustomerPaymentBehaviorResponse(BaseModel):
    customer_id: int
    customer_name: str
    behavior_category: Literal["RELIABLE", "OCCASIONAL_DELAY", "CHRONICALLY_LATE", "HIGH_RISK_DEFAULT"]
    on_time_payment_ratio: float = Field(default=1.0, ge=0.0, le=1.0)
    average_days_to_pay: float = 0.0
    outstanding_balance: float = 0.0
    credit_limit_utilization_pct: float = 0.0
    credit_limit: float = 0.0
    recommended_credit_action: str
    communication_strategy: str
    evidence: List[EvidenceFact] = Field(default_factory=list)


class OperationalActionRecommendation(BaseModel):
    recommendation_id: str
    title: str
    description: str
    target_action: str
    risk_rating: Literal["LOW", "MEDIUM", "HIGH", "CRITICAL"]
    requires_approval: bool
    action_payload_template: Dict[str, Any] = Field(default_factory=dict)


class OperationalRecommendationsResponse(BaseModel):
    invoice_id: int
    recommendations: List[OperationalActionRecommendation] = Field(default_factory=list)


class CollectionDraftRequest(BaseModel):
    context: InvoiceContext
    signals: DeterministicFinanceSignals
    draft_type: Literal["FIRST_REMINDER", "OVERDUE_NOTICE", "FINAL_DEMAND", "PAYMENT_PLAN_OFFER", "INTERNAL_ESCALATION"] = "FIRST_REMINDER"
    tone: Literal["POLITE", "ASSERTIVE", "URGENT", "FORMAL", "PROFESSIONAL", "FIRM_FORMAL", "ACCOMMODATING"] = "POLITE"
    user_instructions: Optional[str] = None
    correlation_id: str = "corr-default"


class CollectionDraftResponse(BaseModel):
    subject: str
    message_body: str
    internal_notes: str
    recipient_preview: Dict[str, str] = Field(default_factory=dict)
    outstanding_amount: float
    currency: str
    requires_approval: bool = True
    tone_applied: str
    safety_restrictions: List[str] = Field(default_factory=list)
