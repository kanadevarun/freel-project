"""
ops_exception_node.py — OperationsAgent: Classify and Persist Exceptions

Applies deterministic severity rules (in Python, not LLM) to each detected exception
signal, then calls Go backend to persist each one.

Severity Rules:
  DELAY < 48h       → INFO
  DELAY 48-96h      → WARNING
  DELAY > 96h       → CRITICAL
  ROLLOVER          → CRITICAL
  CUSTOMS_HOLD      → CRITICAL
  PORT_CONGESTION   → WARNING
  WEATHER           → CRITICAL (always — weather events are critical)
"""

from typing import Dict, Any
from app.state.operations_state import OperationsAgentState
from app.tools.shipments_tool import create_exception


# Title templates per exception type
EXCEPTION_TITLES = {
    "DELAY": "Shipment Delayed",
    "ROLLOVER": "Shipment Rolled to Next Sailing",
    "CUSTOMS_HOLD": "Customs Hold",
    "PORT_CONGESTION": "Port Congestion Delay",
    "WEATHER": "Weather Disruption",
}


EXCEPTION_TYPE_MAP = {
    "DELAY": "ETA_DELAY",
    "ROLLOVER": "VESSEL_ROLLOVER",
    "CUSTOMS_HOLD": "CUSTOMS_HOLD",
    "PORT_CONGESTION": "PORT_CONGESTION",
    "WEATHER": "OTHER",
}

SEVERITY_MAP = {
    "CRITICAL": "CRITICAL",
    "WARNING": "HIGH",
    "INFO": "LOW",
}

def _classify_severity(exception_type: str, delay_hours: int) -> str:
    """
    Deterministic severity classification.
    LLM does NOT decide severity — this Go-equivalent rule is enforced here.
    """
    if exception_type == "ROLLOVER":
        return "CRITICAL"
    if exception_type == "CUSTOMS_HOLD":
        return "CRITICAL"
    if exception_type == "WEATHER":
        return "CRITICAL"
    if exception_type == "PORT_CONGESTION":
        return "HIGH"
    if exception_type == "DELAY":
        if delay_hours >= 96:
            return "CRITICAL"
        elif delay_hours >= 48:
            return "HIGH"
        else:
            return "LOW"
    return "LOW"


def ops_exception_node(state: OperationsAgentState) -> Dict[str, Any]:
    """
    LangGraph Node: Classify and persist detected exceptions to Go backend.
    Only runs if a shipment has been identified.
    """
    print(f"[Operations Agent] Node: ops_exception_node starting...")

    shipment_id = state.get("shipment_id")
    org_id = state.get("org_id", 1)
    detected_exceptions = state.get("detected_exceptions", [])

    if not shipment_id:
        print("[Operations Agent] ops_exception_node: No shipment identified, skipping exceptions.")
        return {"has_critical_exception": False}

    if not detected_exceptions:
        print("[Operations Agent] ops_exception_node: No exceptions detected.")
        return {"has_critical_exception": False}

    has_critical = False
    for exc in detected_exceptions:
        raw_exc_type = exc.get("type", "DELAY")
        delay_hours = exc.get("delay_hours", 0)
        details = exc.get("details", "")

        severity = _classify_severity(raw_exc_type, delay_hours)
        mapped_type = EXCEPTION_TYPE_MAP.get(raw_exc_type, "OTHER")
        title = EXCEPTION_TITLES.get(raw_exc_type, f"Shipment Exception: {raw_exc_type}")
        description = details or state.get("raw_description", "")[:500]

        success = create_exception(
            shipment_id=shipment_id,
            exception_type=mapped_type,
            severity=severity,
            title=title,
            description=description,
            org_id=org_id,
            source_event_id=state.get("event_id"),
        )

        if success:
            print(f"[Operations Agent] Exception {exc_type} ({severity}) created for shipment {shipment_id}.")
            if severity == "CRITICAL":
                has_critical = True

    print(f"[Operations Agent] Exception processing complete. Has critical: {has_critical}")
    return {"has_critical_exception": has_critical}
