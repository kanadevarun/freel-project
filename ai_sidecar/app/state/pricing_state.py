from typing import TypedDict, Annotated, List, Optional, Any
from langgraph.graph.message import add_messages

class PricingAgentState(TypedDict):
    """
    State for the Pricing Agent LangGraph workflow.
    Uses Message-Based Agentic Architecture with structured metadata fields.
    """
    # ── Message history (accumulates via add_messages reducer) ───────────────
    messages: Annotated[List[Any], add_messages]
    
    # ── Structured context fields ────────────────────────────────────────────
    org_id: int
    rfq_id: int
    origin: str
    destination: str
    incoterms: str
    equipment_type: str
    gross_weight: float
    volume_cbm: float
    commodity: str
    target_date: Optional[str]
    
    # ── Ingested database parameters ─────────────────────────────────────────
    pricing_rules: List[dict]
    raw_rates: List[dict]
    
    # ── Derived outputs ──────────────────────────────────────────────────────
    suggested_quotes: List[dict]
    overall_reasoning: str
    is_anomaly: bool
    confidence_score: int
    error_message: Optional[str]
