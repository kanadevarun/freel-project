"""
RFQ-to-Quotation Automation and Intelligent Pricing Workflow Models
Phase 3 Task 3.4 for LogisticsHQ

Typed Pydantic domain models for RFQ requirement extraction, deterministic pricing explanation,
quotation risk analysis, quotation drafting, and safety evaluation.
"""

from typing import List, Optional, Literal, Dict, Any
from pydantic import BaseModel, Field


class EvidenceFact(BaseModel):
    source_module: str = Field(..., description="Originating module, e.g. rfqs, quotations, rates, customers")
    source_entity_id: int = Field(..., description="ID of source business record")
    source_ref: str = Field(..., description="Human-readable business reference code (e.g. RFQ-102, QUOTE-204, RATE-501)")
    field_name: str = Field(..., description="Specific field evaluated")
    observed_value: str = Field(..., description="Factual value verified in database")
    description: str = Field(..., description="Contextual explanation of verified fact")


class SafetyWarning(BaseModel):
    warning_type: str = Field(..., description="Warning type, e.g. MISSING_MANDATORY_FIELD, LOW_MARGIN, RATE_EXPIRED")
    message: str = Field(..., description="Actionable explanation for human operator")
    severity: Literal["INFO", "WARNING", "CRITICAL"] = Field(default="WARNING")


class RFQItemContext(BaseModel):
    item_id: Optional[int] = None
    description: Optional[str] = None
    cargo_type: Optional[str] = None
    package_type: Optional[str] = None
    quantity: float = 1.0
    weight_kg: float = 0.0
    volume_cbm: float = 0.0
    dimensions: Optional[str] = None


class CarrierQuoteFact(BaseModel):
    quote_id: int
    carrier_id: Optional[int] = None
    carrier_name: str
    service_level: Optional[str] = None
    transit_time_days: Optional[int] = None
    total_cost: float
    currency: str = "USD"
    validity_date: Optional[str] = None
    expired: bool = False


class RateRecordFact(BaseModel):
    rate_id: int
    carrier_name: str
    origin_port: str
    destination_port: str
    mode: str
    equipment_type: Optional[str] = None
    base_ocean_rate: float = 0.0
    bunker_fuel_surcharge: float = 0.0
    security_fee: float = 0.0
    terminal_handling: float = 0.0
    total_cost: float = 0.0
    currency: str = "USD"
    effective_date: Optional[str] = None
    expiry_date: Optional[str] = None
    is_expired: bool = False


class RFQContext(BaseModel):
    """
    Structured RFQ and commercial context provided exclusively by Go backend.
    """
    org_id: int
    rfq_id: int
    rfq_number: str
    customer_id: int
    customer_name: str
    customer_tier: Optional[str] = "STANDARD"
    payment_terms: Optional[str] = "NET_30"
    credit_limit: float = 0.0
    outstanding_balance: float = 0.0

    origin: Optional[str] = None
    destination: Optional[str] = None
    shipment_mode: Optional[str] = "OCEAN_FCL"
    incoterms: Optional[str] = None
    target_date: Optional[str] = None
    deadline_date: Optional[str] = None

    special_instructions: Optional[str] = None
    commodity: Optional[str] = None
    customs_required: bool = False
    insurance_required: bool = False

    items: List[RFQItemContext] = Field(default_factory=list)
    carrier_quotes: List[CarrierQuoteFact] = Field(default_factory=list)
    available_rates: List[RateRecordFact] = Field(default_factory=list)
    correlation_id: str


class ExtractedRequirementField(BaseModel):
    field_name: str
    value: Optional[str] = None
    status: Literal["VERIFIED", "EXTRACTED", "MISSING", "AMBIGUOUS"]
    confidence: float = Field(..., ge=0.0, le=1.0)
    source: str
    evidence: Optional[str] = None
    missing_impact: Optional[str] = None


class ExtractedRFQRequirements(BaseModel):
    rfq_id: int
    rfq_number: str
    status: Literal["COMPLETE", "INCOMPLETE", "CLARIFICATION_REQUIRED"]
    fields: Dict[str, ExtractedRequirementField]
    missing_mandatory: List[str] = Field(default_factory=list)
    missing_optional: List[str] = Field(default_factory=list)
    clarification_recommendation: Optional[str] = None
    confidence_score: float = Field(..., ge=0.0, le=1.0)
    evidence: List[EvidenceFact] = Field(default_factory=list)
    correlation_id: str


class DeterministicPricingFacts(BaseModel):
    """
    Authoritative calculations supplied strictly by Go engine.
    Python cannot override these values.
    """
    currency: str = "USD"
    base_cost: float = 0.0
    total_cost: float = 0.0
    base_sell: float = 0.0
    surcharges: float = 0.0
    discounts: float = 0.0
    tax_amount: float = 0.0
    total_selling_price: float = 0.0
    gross_profit: float = 0.0
    gross_margin_pct: float = 0.0
    margin_health: Literal["HEALTHY", "LOW", "NEGATIVE"] = "HEALTHY"
    applied_rate_id: Optional[int] = None
    applied_carrier_name: Optional[str] = None
    rate_is_expired: bool = False


class PricingExplanation(BaseModel):
    rfq_id: int
    carrier_rate_selected: Optional[str] = None
    cost_explanation: str
    sell_pricing_notes: str
    commercial_nuance: str
    pricing_warnings: List[SafetyWarning] = Field(default_factory=list)
    confidence_score: float = Field(..., ge=0.0, le=1.0)
    evidence: List[EvidenceFact] = Field(default_factory=list)
    correlation_id: str


class QuotationRiskAnalysis(BaseModel):
    rfq_id: int
    risk_level: Literal["LOW", "MEDIUM", "HIGH", "CRITICAL"]
    requires_approval: bool
    approval_triggers: List[str] = Field(default_factory=list)
    risk_reasons: List[str] = Field(default_factory=list)
    margin_health_evaluation: str
    rate_freshness_evaluation: str
    deadline_urgency: str
    suggested_next_steps: List[str] = Field(default_factory=list)
    safety_warnings: List[SafetyWarning] = Field(default_factory=list)
    confidence_score: float = Field(..., ge=0.0, le=1.0)
    evidence: List[EvidenceFact] = Field(default_factory=list)
    correlation_id: str


class QuotationDraftRequest(BaseModel):
    rfq_id: int
    rfq_number: str
    customer_id: int
    customer_name: str
    contact_name: Optional[str] = None
    contact_email: Optional[str] = None
    origin: str
    destination: str
    shipment_mode: str
    incoterms: str
    pricing: DeterministicPricingFacts
    items_summary: str
    customs_required: bool = False
    insurance_required: bool = False
    validity_days: int = 14
    user_prompt_notes: Optional[str] = None
    correlation_id: str


class QuotationDraftResponse(BaseModel):
    rfq_id: int
    rfq_number: str
    quotation_title: str
    internal_summary: str
    customer_wording: str
    pricing_explanation: str
    terms_and_conditions: str
    validity_start: str
    validity_end: str
    recipient_preview: Dict[str, str]
    requires_approval: bool
    approval_triggers: List[str] = Field(default_factory=list)
    confidence_score: float = Field(..., ge=0.0, le=1.0)
    evidence: List[EvidenceFact] = Field(default_factory=list)
    warnings: List[SafetyWarning] = Field(default_factory=list)
    correlation_id: str
