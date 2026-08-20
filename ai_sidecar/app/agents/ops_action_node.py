"""
ops_action_node.py — OperationsAgent: Final Callback

Sends the final structured callback to Go backend once milestones and
exceptions have been persisted. The callback includes:
  - shipment_id, org_id
  - has_critical_exception flag
  - ai_summary of what was processed
"""

from typing import Dict, Any
from app.state.operations_state import OperationsAgentState
from app.tools.shipments_tool import send_operations_callback


def ops_action_node(state: OperationsAgentState) -> Dict[str, Any]:
    """
    LangGraph Node: Send the final agent callback to Go backend.
    Always runs — even if shipment wasn't identified, to signal task completion.
    """
    print(f"[Operations Agent] Node: ops_action_node starting...")

    shipment_id = state.get("shipment_id") or 0
    org_id = state.get("org_id", 1)
    has_critical = state.get("has_critical_exception", False)
    ai_summary = state.get("ai_summary", "Carrier update processed by OperationsAgent.")
    import os
    go_backend_url = os.getenv("GO_BACKEND_URL", "http://localhost:8080")
    callback_url = state.get("callback_url")
    if not callback_url or "localhost:8080" in callback_url:
        callback_url = f"{go_backend_url}/internal/operations/callback"

    requires_human_review = state.get("requires_human_review", False)
    if requires_human_review:
        ai_summary = f"[NEEDS REVIEW] {ai_summary}"
        has_critical = True  # Flag for ops team to handle

    if shipment_id > 0:
        success = send_operations_callback(
            callback_url=callback_url,
            shipment_id=shipment_id,
            org_id=org_id,
            has_critical=has_critical,
            ai_summary=ai_summary,
            event_id=state.get("event_id"),
        )
        if success:
            print(f"[Operations Agent] Callback sent successfully for shipment {shipment_id}.")
        else:
            print(f"[Operations Agent] Callback failed for shipment {shipment_id}.")
    else:
        print(f"[Operations Agent] No shipment identified — skipping callback.")

    return {
        "ai_summary": ai_summary,
        "has_critical_exception": has_critical,
        "error_message": None,
    }
