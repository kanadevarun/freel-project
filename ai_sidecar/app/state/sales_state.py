from typing import TypedDict, Annotated, List, Optional, Any, Dict
from langgraph.graph.message import add_messages

class SalesAgentState(TypedDict):
    """
    State for the Sales/Email Lead Agent LangGraph workflow.
    Uses Message-Based Agentic Architecture with structured metadata fields.
    """
    # ── Message history (accumulates via add_messages reducer) ───────────────
    messages: Annotated[List[Any], add_messages]
    
    # ── Context metadata ─────────────────────────────────────────────────────
    org_id: int
    interaction_id: int
    lead_id: int
    from_email: str
    email_subject: str
    email_body: str
    callback_url: str
    
    # ── Thread-awareness fields (closed-loop reply tracking) ─────────────────
    thread_id: Optional[str]             # Thread identifier for conversation tracking
    is_reply: bool                       # True if this email is a reply in an existing thread
    parent_interaction_id: Optional[int] # ID of the previous interaction this replies to
    prior_rfq_context: Optional[Dict[str, Any]]  # Structured fields extracted in previous turns
                                                  # ONLY structured keys, NOT raw email text

    # ── Lead Research / Enrichment data ──────────────────────────────────────
    company_domain: Optional[str]
    company_enrichment: Optional[str]
    lead_name: Optional[str]
    
    # ── Extracted Cargo/RFQ Parameters ───────────────────────────────────────
    origin_port: Optional[str]
    destination_port: Optional[str]
    incoterms: Optional[str]
    cargo_description: Optional[str]
    cargo_weight: Optional[float]
    cargo_volume: Optional[float]
    target_date: Optional[str]
    
    # ── Classification & Action Outputs ──────────────────────────────────────
    intent: str          # "RFQ_REQUEST" | "QUESTION" | "MEETING" | "UNSUBSCRIBE" | "FOLLOW_UP" | "RFQ_REQUEST_INCOMPLETE"
    sentiment: str       # "POSITIVE" | "NEUTRAL" | "NEGATIVE"
    confidence_score: int
    linked_rfq_id: Optional[int]
    ai_summary: str
    drafted_reply: Optional[str]
    error_message: Optional[str]
