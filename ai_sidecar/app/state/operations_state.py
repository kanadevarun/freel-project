from typing import TypedDict, Annotated, List, Optional, Any, Dict

class OperationsAgentState(TypedDict):
    """
    State for the Operations Agent LangGraph workflow.
    Processes inbound carrier tracking events and updates shipment milestones/exceptions.
    """
    # ── Core identifiers ─────────────────────────────────────────────────────
    org_id: int
    entity_id: str            # Will be the shipment_id (as string) once identified
    callback_url: str

    # ── Raw carrier event data ───────────────────────────────────────────────
    event_id: str
    carrier_scac: str
    booking_number: str
    container_number: str
    vessel_name: str
    voyage_number: str
    milestone_code: str       # Pre-parsed code if deterministically available
    event_time: str           # ISO8601 string
    location: str
    raw_description: str      # Free-text description from carrier (can be unstructured)

    # ── Identified shipment (after lookup) ──────────────────────────────────
    shipment_id: Optional[int]
    shipment_data: Optional[Dict[str, Any]]   # Full shipment detail from Go
    identification_confident: bool            # True = deterministic match found

    # ── Parsed LLM output ────────────────────────────────────────────────────
    detected_milestones: List[Dict[str, Any]]   # [{code, date, location, notes}]
    detected_exceptions: List[Dict[str, Any]]   # [{type, severity, title, description}]

    # ── Control flags ────────────────────────────────────────────────────────
    has_critical_exception: bool        # If True → alert ops team in callback
    requires_human_review: bool         # HITL: True → flag for manual ops resolution
    ai_summary: str
    error_message: Optional[str]
