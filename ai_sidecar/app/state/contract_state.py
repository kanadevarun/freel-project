from pydantic import BaseModel, Field
from typing import List, Optional, Dict, Any

# Simple meaning:
#   These are data structures defined using Pydantic. They act as "type safety guards"
#   for the variables that are passed between the nodes in the LangGraph workflow.

class StateSurcharge(BaseModel):
    """Represents a carrier surcharge fee block (e.g. BAF Fuel charge)."""
    code: str                  # e.g., "BAF", "CAF"
    description: str           # e.g., "Bunker Adjustment Factor"
    amount: float              # e.g., 350.0
    unit: str                  # e.g., "PER_TEU", "PER_CONTAINER"
    included: bool             # e.g., True (already in ocean freight) or False (additional)

class StateRateDraft(BaseModel):
    """Represents a single parsed port-pair rate row before it is stored in PostgreSQL."""
    origin_port: str           # Standard UN/LOCODE, e.g. "INNSA"
    destination_port: str      # Standard UN/LOCODE, e.g. "DEHAM"
    via_port: Optional[str] = None
    service_code: Optional[str] = None
    carrier_scac: str          # Carrier SCAC identifier, e.g. "MAEU"
    carrier_name: str          # Carrier display name, e.g. "Maersk"
    vessel_name: Optional[str] = None
    equipment_type: str        # e.g. "40GP", "20GP"
    ocean_freight: float       # Base rate, e.g. 2800.0
    origin_charges: float = 0.0
    destination_charges: float = 0.0
    surcharges: List[StateSurcharge] = []
    total_buy_price: float     # Combined base + surcharges
    currency_original: str = "USD"
    exchange_rate_used: float = 1.0
    included_charges: List[str] = []
    excluded_charges: List[str] = []
    free_days_origin: int = 0
    free_days_destination: int = 14
    transit_days: Optional[int] = None
    incoterms: Optional[str] = None
    commodity_restrictions: List[str] = []
    routing_conditions: Optional[str] = None
    valid_from: str            # RFC3339 timestamp, e.g. "2026-09-01T00:00:00Z"
    valid_until: str           # RFC3339 timestamp, e.g. "2026-12-31T23:59:59Z"
    confidence_score: int      # AI extraction confidence estimation (0-100)

class StateReviewItemDraft(BaseModel):
    """Represents an anomalous rate row flagged by the validator node for operator review."""
    extracted_data: StateRateDraft  # The parsed rate draft details
    confidence_score: int           # e.g., 62 (low confidence)
    review_flags: List[str]         # Reason for flagging, e.g., ["PRICE_ANOMALY"] or ["PORT_UNKNOWN"]
    ai_reasoning: str               # Explanatory prompt message diagnostic
    source_page: int                # Document page index where this was found
    source_text: str                # Note quote snippet containing the rate
    source_image_url: str           # Visual screenshot locator

class StateLogEntry(BaseModel):
    """Represents an agent execution history log step."""
    step: str                  # e.g., "OCR_PROCESSING", "CLASSIFICATION"
    timestamp: str             # UTC execution time string
    message: str               # Log detail text

class ContractExtractionState(BaseModel):
    """
    The main global state dictionary passed from node to node in LangGraph.
    
    Any node in the graph can read this state and update its key parameters.
    """
    document_id: str           # Unique contract UUID reference
    org_id: int                # Organization owner identifier
    s3_key: str                # File path reference
    file_type: str             # "PDF" or "XLSX"
    callback_url: str          # Webhook callback url to post final results
    correlation_id: str = ""   # Trace correlation identifier (correlation_id)
    raw_text: str = ""         # Extracted document text block
    carrier_name: Optional[str] = None
    carrier_scac: Optional[str] = None
    extracted_rates: List[StateRateDraft] = []  # List of clean confirmed rates
    flagged_items: List[StateReviewItemDraft] = [] # List of flagged anomalous items
    processing_log: List[StateLogEntry] = []       # Graph execution logs
    ai_summary: str = ""       # Gemini terms summary abstract
    is_anomaly_detected: bool = False             # Set to True to trigger interrupt pauses
    status: str = "QUEUED"     # Current pipeline status


