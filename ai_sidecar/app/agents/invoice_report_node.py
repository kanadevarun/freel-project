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
        "ai_summary": ai_summary,
    }

    token = os.getenv("INTERNAL_SERVICE_TOKEN", "internal-service-key-logisticshq")
    internal_headers = {
        "X-LogisticsHQ-Service-Key": token,
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
