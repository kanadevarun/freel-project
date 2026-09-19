import os
import httpx
from typing import Optional, Dict, Any

from app.tools.auth_utils import get_internal_service_token

go_backend_url = os.getenv("GO_BACKEND_URL", "http://localhost:8080")

def get_headers():
    return {
        "X-LogisticsHQ-Service-Key": get_internal_service_token(),
        "Content-Type": "application/json"
    }


import random
import time

def _make_result(success: bool, status_code: int = 0, message: str = "", retryable: bool = False, data: Any = None) -> dict:
    return {
        "success": success,
        "status_code": status_code,
        "message": message,
        "retryable": retryable,
        "data": data,
    }

RETRYABLE_STATUS_CODES = {429, 502, 503, 504}

def _call_with_retry(fn, max_attempts: int = 3) -> dict:
    last_res = _make_result(False, message="No attempts made")
    for attempt in range(max_attempts):
        try:
            res = fn()
            if res["success"]:
                return res
            last_res = res
            if not res["retryable"]:
                return res
        except Exception as e:
            last_res = _make_result(False, message=str(e), retryable=True)
        
        wait = (2 ** attempt) + random.uniform(0, 1)
        print(f"[Shipments Tool] Retry {attempt+1}/{max_attempts} in {wait:.2f}s due to: {last_res['message']}")
        time.sleep(wait)
    return last_res

def get_shipment(shipment_id: int, org_id: int = 1) -> Optional[Dict[str, Any]]:
    """
    Fetch full shipment detail from Go backend including milestones.
    Returns None on failure.
    """
    url = f"{go_backend_url}/internal/shipments/{shipment_id}?org_id={org_id}"
    def _call():
        try:
            response = httpx.get(url, headers=get_headers(), timeout=10.0)
            if response.status_code == 200:
                return _make_result(True, data=response.json().get("data"))
            retry = response.status_code in RETRYABLE_STATUS_CODES
            return _make_result(False, response.status_code, response.text, retry)
        except Exception as e:
            return _make_result(False, message=str(e), retryable=True)

    res = _call_with_retry(_call)
    return res.get("data") if res["success"] else None


def update_milestone(shipment_id: int, milestone_code: str, actual_date: str,
                     location: Optional[str] = None, notes: Optional[str] = None,
                     org_id: int = 1) -> bool:
    """
    Update a specific milestone as COMPLETED with actual date through Centralized Action System.
    Returns True if successful.
    """
    from app.tools.action_bridge import execute_action

    input_data: Dict[str, Any] = {
        "shipment_id": shipment_id,
        "milestone_code": milestone_code,
        "actual_date": actual_date,
    }
    if location:
        input_data["location"] = location
    if notes:
        input_data["notes"] = notes

    res = execute_action(
        action_name="shipments.update_milestone",
        org_id=org_id,
        input_data=input_data,
        source="langgraph.operations"
    )
    return bool(res.get("success", False))


def create_exception(shipment_id: int, exception_type: str, severity: str,
                     title: str, description: str, org_id: int = 1, source_event_id: Optional[str] = None) -> bool:
    """
    Create a new exception entry for a shipment through Centralized Action System.
    exception_type: ROLLOVER | DELAY | CUSTOMS_HOLD | PORT_CONGESTION | WEATHER
    severity: INFO | WARNING | CRITICAL
    Returns True if successful.
    """
    from app.tools.action_bridge import execute_action

    input_data: Dict[str, Any] = {
        "shipment_id": shipment_id,
        "exception_type": exception_type,
        "severity": severity,
        "title": title,
        "description": description,
    }
    if source_event_id:
        input_data["source_event_id"] = source_event_id

    res = execute_action(
        action_name="shipments.create_exception",
        org_id=org_id,
        input_data=input_data,
        source="langgraph.operations"
    )
    return bool(res.get("success", False))


def send_operations_callback(callback_url: str, shipment_id: int, org_id: int,
                              has_critical: bool, ai_summary: str, event_id: Optional[str] = None) -> bool:
    """
    Send the final operations callback through Centralized Action System.
    """
    from app.tools.action_bridge import execute_action

    input_data: Dict[str, Any] = {
        "shipment_id": shipment_id,
        "has_critical_exception": has_critical,
        "ai_summary": ai_summary,
    }
    if event_id:
        input_data["event_id"] = event_id

    res = execute_action(
        action_name="operations.send_callback",
        org_id=org_id,
        input_data=input_data,
        source="langgraph.operations"
    )
    return bool(res.get("success", False))

