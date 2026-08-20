import httpx
import json
from app.state.compliance_state import ComplianceState

def doc_report_node(state: ComplianceState) -> ComplianceState:
    """
    Callback Node.
    Sends final parsed results, ocr content logs, and discrepancy arrays to Go callback.
    """
    org_id = state.get("org_id")
    if not org_id or org_id <= 0:
        raise ValueError("Missing or invalid org_id in ComplianceState")

    shipment_id = state.get("shipment_id")
    if not shipment_id or shipment_id <= 0:
        raise ValueError("Missing or invalid shipment_id in ComplianceState")

    doc_id = state.get("doc_id")
    if not doc_id:
        raise ValueError("Missing or invalid doc_id in ComplianceState")

    import os
    go_backend_url = os.getenv("GO_BACKEND_URL", "http://localhost:8080")
    callback_url = state.get("callback_url")
    if not callback_url or "localhost:8080" in callback_url:
        callback_url = f"{go_backend_url}/internal/compliance/callback"

    print(f"[Compliance Agent] Completing verification for document {doc_id}. Callback: {callback_url}")

    # Build status list payload mapping docID -> JSON details
    doc_payload = {
        "status": state.get("doc_status", "VERIFIED"),
        "extracted_data": state.get("extracted_data", {}),
        "raw_ocr_text": state.get("raw_ocr_text", "")
    }
    
    doc_status_list = {
        doc_id: json.dumps(doc_payload)
    }

    payload = {
        "org_id": org_id,
        "shipment_id": shipment_id,
        "doc_status_list": doc_status_list,
        "discrepancies": state.get("discrepancies", [])
    }

    import os
    token = os.getenv("INTERNAL_SERVICE_TOKEN", "internal-service-key-logisticshq")
    internal_headers = {
        "X-LogisticsHQ-Service-Key": token,
        "Content-Type": "application/json"
    }

    resp = httpx.post(callback_url, json=payload, headers=internal_headers, timeout=5.0)
    # Raise exception if status code is not 2xx so worker marks it failed / retries
    resp.raise_for_status()
    print(f"[Compliance Agent] Callback response: {resp.status_code}")

    return state
