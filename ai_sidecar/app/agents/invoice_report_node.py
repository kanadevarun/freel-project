import os
import httpx
from app.state.finance_state import FinanceState


def invoice_report_node(state: FinanceState) -> FinanceState:
    """
    Invoice Report Callback Node.
    Sends final extracted line items, status, and discrepancies back to Go callback.
    """
    org_id = state.get("org_id")
    if not org_id or org_id <= 0:
        raise ValueError("Missing or invalid org_id in FinanceState")

    shipment_id = state.get("shipment_id")
    if not shipment_id or shipment_id <= 0:
        raise ValueError("Missing or invalid shipment_id in FinanceState")

    invoice_id = state.get("invoice_id")
    if not invoice_id:
        raise ValueError("Missing or invalid invoice_id in FinanceState")

    go_backend_url = os.getenv("GO_BACKEND_URL", "http://localhost:8080")
    callback_url = state.get("callback_url")
    if not callback_url or "localhost:8080" in callback_url:
        callback_url = f"{go_backend_url}/internal/finance/callback"

    print(f"[Finance Agent] Callback: Sending reconciliation results to {callback_url}...")

    # Generate a user-friendly AI summary narrative
    discrepancies = state.get("discrepancies", [])
    if len(discrepancies) == 0:
        ai_summary = f"Invoice successfully verified. 3-way match complete with 0 discrepancies."
    else:
        disc_details = ", ".join([f"{d['charge_code']} ({d['field_name']}: actual {d['actual_value']} vs expected {d['expected_value']})" for d in discrepancies])
        ai_summary = f"Reconciliation flagged {len(discrepancies)} discrepancies: {disc_details}."

    payload = {
        "org_id": org_id,
        "shipment_id": shipment_id,
        "invoice_id": invoice_id,
        "status": state.get("invoice_status", "APPROVED"),
        "items": state.get("extracted_items", []),
        "discrepancies": discrepancies,
    # Execute through Centralized Action System
    from app.tools.action_bridge import execute_action
    action_res = execute_action(
        action_name="finance.reconcile_invoice",
        org_id=org_id,
        input_data={
            "shipment_id": shipment_id,
            "invoice_id": invoice_id,
            "status": state.get("invoice_status", "APPROVED"),
            "items": state.get("extracted_items", []),
            "discrepancies": discrepancies,
            "ai_summary": ai_summary
        },
        source="langgraph.finance"
    )
    if action_res.get("success"):
        print(f"[Finance Agent] Action System execution success for invoice {invoice_id}")
        return {
            "discrepancies": discrepancies,
            "ai_summary": ai_summary,
            "invoice_status": state.get("invoice_status", "APPROVED")
        }

    # Fallback to direct callback if action did not execute
    from app.tools.auth_utils import get_internal_service_token
    internal_headers = {
        "X-LogisticsHQ-Service-Key": get_internal_service_token(),
        "Content-Type": "application/json"
    }

    resp = httpx.post(callback_url, json=payload, headers=internal_headers, timeout=5.0)
    # Ensure error propagation to worker queue
    resp.raise_for_status()

    print(f"[Finance Agent] Callback complete. Status code: {resp.status_code}")

    return {
        **state,
        "ai_summary": ai_summary,
    }
