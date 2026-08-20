"""
ops_milestone_node.py — OperationsAgent: Update Shipment Milestones

Calls Go internal endpoint to mark milestones as COMPLETED.
Milestone codes: BOOKED, DEPARTED, IN_TRANSIT, ARRIVED, DELIVERED
"""

from typing import Dict, Any
from app.state.operations_state import OperationsAgentState
from app.tools.shipments_tool import update_milestone, get_shipment


VALID_MILESTONE_CODES = {"BOOKED", "DEPARTED", "IN_TRANSIT", "ARRIVED", "DELIVERED"}


def ops_milestone_node(state: OperationsAgentState) -> Dict[str, Any]:
    """
    LangGraph Node: Push detected milestones to Go backend.
    Only runs if a shipment has been identified.
    """
    print(f"[Operations Agent] Node: ops_milestone_node starting...")

    shipment_id = state.get("shipment_id")
    org_id = state.get("org_id", 1)
    detected_milestones = state.get("detected_milestones", [])

    if not shipment_id:
        print("[Operations Agent] ops_milestone_node: No shipment identified, skipping milestone updates.")
        return {}

    if not detected_milestones:
        print("[Operations Agent] ops_milestone_node: No milestones detected, skipping.")
        return {}

    updated = []
    for m in detected_milestones:
        code = (m.get("code") or "").upper()
        if code not in VALID_MILESTONE_CODES:
            print(f"[Operations Agent] Skipping unknown milestone code: {code}")
            continue

        actual_date = m.get("date")
        location = m.get("location")
        notes = m.get("notes")

        # If no date was extracted, use event_time from state
        if not actual_date:
            actual_date = state.get("event_time") or "2026-01-01T00:00:00Z"

        success = update_milestone(
            shipment_id=shipment_id,
            milestone_code=code,
            actual_date=actual_date,
            location=location,
            notes=notes,
            org_id=org_id,
        )

        if success:
            updated.append(code)
            print(f"[Operations Agent] Milestone {code} marked COMPLETED for shipment {shipment_id}.")

    print(f"[Operations Agent] Updated {len(updated)} milestones: {updated}")
    return {}
