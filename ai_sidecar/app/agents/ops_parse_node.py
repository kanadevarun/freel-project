"""
ops_parse_node.py — OperationsAgent: Parse Carrier Update

This LLM node receives raw unstructured carrier update text and extracts:
  - Detected milestones (code, date, location)
  - Potential exception signals (delay, rollover, customs hold, etc.)

The LLM interprets the free-form description. Downstream nodes apply Go-side
deterministic business rules for severity classification and DB writes.
"""

import json
import re
from typing import Dict, Any

from app.state.operations_state import OperationsAgentState
from app.agents.llm_utils import execute_llm_json


PARSE_SYSTEM_PROMPT = """You are an expert freight logistics operations analyst.
Your job is to parse a raw carrier tracking update and extract structured information.

Given a raw description from a carrier update, extract:

1. milestones: A list of shipping milestones detected. Each entry must have:
   - code: one of BOOKED, DEPARTED, IN_TRANSIT, ARRIVED, DELIVERED (map appropriately)
   - date: ISO8601 date string (e.g. "2026-08-15T00:00:00Z"), null if not mentioned
   - location: port or place name if mentioned, null otherwise
   - notes: brief note about this event

2. exception_signals: A list of anomalies detected. Each entry must have:
   - type: one of DELAY, ROLLOVER, CUSTOMS_HOLD, PORT_CONGESTION, WEATHER
   - detected: true/false
   - details: brief text explaining what triggered this signal
   - delay_hours: estimated hours of delay if applicable, 0 otherwise

3. needs_human_review: true if the update is ambiguous or contains a serious unrecognized event

4. ai_summary: A single plain English sentence summarizing what happened to this shipment.

Return ONLY a JSON object with these 4 keys. No markdown, no explanation.

Example:
{
  "milestones": [{"code": "DEPARTED", "date": "2026-08-12T00:00:00Z", "location": "INNSA", "notes": "Vessel departed on schedule"}],
  "exception_signals": [{"type": "DELAY", "detected": false, "details": "", "delay_hours": 0}],
  "needs_human_review": false,
  "ai_summary": "Vessel departed Nhava Sheva on August 12 as scheduled."
}
"""

# Keywords for exception detection fallback (used when LLM is unavailable)
ROLLOVER_KEYWORDS = ["rolled", "rollover", "next sailing", "vessel change", "alternative vessel"]
DELAY_KEYWORDS = ["delay", "delayed", "postponed", "pushed back", "revised eta", "revised etd", "behind schedule"]
CUSTOMS_KEYWORDS = ["customs hold", "customs examination", "detention", "customs inspection", "held at customs"]
CONGESTION_KEYWORDS = ["port congestion", "berth delay", "anchorage", "congestion", "port delay"]
WEATHER_KEYWORDS = ["typhoon", "cyclone", "storm", "weather", "adverse conditions", "bad weather"]


def _extract_delay_hours(text: str) -> int:
    """
    Dynamically extract explicit delay hours from the raw description.
    Returns 0 if no delay or if the duration is unknown (Group 5 logic).
    """
    text_lower = text.lower()
    # E.g. "delay of 3 days", "delayed 2 days", "48 hours delay"
    patterns = [
        r'(\d+)\s*day',
        r'(\d+)\s*hour',
    ]
    for pattern in patterns:
        match = re.search(pattern, text_lower)
        if match:
            num = int(match.group(1))
            if 'day' in pattern:
                return num * 24
            return num
    return 0  # 0 indicates unknown or unspecified delay duration

def _mock_parse(raw_description: str) -> Dict[str, Any]:
    """
    Deterministic keyword-based fallback parser when LLM is unavailable.
    Used for testing under rate limits and ensures E2E coverage.
    """
    desc_lower = raw_description.lower()

    milestones = []
    exception_signals = []

    if any(k in desc_lower for k in ["depart", "sailed", "atd"]):
        milestones.append({"code": "DEPARTED", "date": None, "location": None, "notes": "Departure detected"})
    if any(k in desc_lower for k in ["arriv", "ata", "discharg"]):
        milestones.append({"code": "ARRIVED", "date": None, "location": None, "notes": "Arrival detected"})
    if any(k in desc_lower for k in ["deliver", "final delivery"]):
        milestones.append({"code": "DELIVERED", "date": None, "location": None, "notes": "Delivery detected"})
    if any(k in desc_lower for k in ["book", "booking confirm"]):
        milestones.append({"code": "BOOKED", "date": None, "location": None, "notes": "Booking confirmed"})

    # Exception detection
    for exc_type, keywords in [
        ("ROLLOVER", ROLLOVER_KEYWORDS),
        ("DELAY", DELAY_KEYWORDS),
        ("CUSTOMS_HOLD", CUSTOMS_KEYWORDS),
        ("PORT_CONGESTION", CONGESTION_KEYWORDS),
        ("WEATHER", WEATHER_KEYWORDS),
    ]:
        detected = any(k in desc_lower for k in keywords)
        
        delay_hours = 0
        if detected and exc_type in ("DELAY", "ROLLOVER"):
            delay_hours = _extract_delay_hours(raw_description)

        exception_signals.append({
            "type": exc_type,
            "detected": detected,
            "details": f"Keyword matched in: '{raw_description[:80]}'" if detected else "",
            "delay_hours": delay_hours,
        })

    ai_summary = f"Carrier update processed: {raw_description[:120]}"
    needs_human_review = not milestones and not any(e["detected"] for e in exception_signals)

    return {
        "milestones": milestones,
        "exception_signals": exception_signals,
        "needs_human_review": needs_human_review,
        "ai_summary": ai_summary,
    }


def ops_parse_node(state: OperationsAgentState) -> Dict[str, Any]:
    """
    LangGraph Node: Parse the raw carrier update description using LLM.
    Falls back to keyword-based parsing under rate limits.
    """
    print(f"[Operations Agent] Node: ops_parse_node starting...")

    raw_description = state.get("raw_description", "")
    carrier_scac = state.get("carrier_scac", "")
    vessel_name = state.get("vessel_name", "")
    booking_number = state.get("booking_number", "")
    event_time = state.get("event_time", "")

    if not raw_description:
        # Nothing to parse — just pass through
        return {
            "detected_milestones": [],
            "detected_exceptions": [],
            "has_critical_exception": False,
            "requires_human_review": False,
            "ai_summary": "No carrier update text provided.",
        }

    prompt = f"""{PARSE_SYSTEM_PROMPT}

CARRIER UPDATE TO PARSE:
Carrier: {carrier_scac}
Vessel: {vessel_name}
Booking: {booking_number}
Event Time: {event_time}
Description: {raw_description}
"""

    result = execute_llm_json(prompt)

    if result is None:
        print("[Operations Agent] ops_parse_node: LLM unavailable, using keyword fallback.")
        result = _mock_parse(raw_description)

    # Extract milestones and exception signals
    detected_milestones = result.get("milestones", [])
    exception_signals = result.get("exception_signals", [])
    ai_summary = result.get("ai_summary", raw_description[:120])
    needs_human_review = result.get("needs_human_review", False)

    # Convert exception signals to detected_exceptions list (only detected ones)
    detected_exceptions = [s for s in exception_signals if s.get("detected")]

    print(f"[Operations Agent] Parsed {len(detected_milestones)} milestones, {len(detected_exceptions)} exception signals.")

    return {
        "detected_milestones": detected_milestones,
        "detected_exceptions": detected_exceptions,
        "requires_human_review": needs_human_review,
        "ai_summary": ai_summary,
        "error_message": None,
    }
