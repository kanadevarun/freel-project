import httpx
from app.state.compliance_state import ComplianceState

def doc_reconcile_node(state: ComplianceState) -> ComplianceState:
    """
    Deterministic Cross-Reconciliation Node (Decision 6).
    Cross-checks extracted data against other active shipment documents.
    """
    org_id = state.get("org_id")
    if not org_id or org_id <= 0:
        raise ValueError("Missing or invalid org_id in ComplianceState")

    shipment_id = state.get("shipment_id")
    if not shipment_id or shipment_id <= 0:
        raise ValueError("Missing or invalid shipment_id in ComplianceState")

    doc_type = state.get("doc_type", "")
    extracted = state.get("extracted_data", {})
    discrepancies = []

    print(f"[Compliance Agent] Reconciling {doc_type} data against existing documents...")

    # Call internal shipments endpoint to fetch current shipment metadata
    # Secure token validation headers included
    import os
    token = os.getenv("INTERNAL_SERVICE_TOKEN", "internal-service-key-logisticshq")
    go_backend_url = os.getenv("GO_BACKEND_URL", "http://localhost:8080")
    internal_headers = {
        "X-LogisticsHQ-Service-Key": token,
        "Content-Type": "application/json"
    }

    try:
        # 1. Fetch current active documents already uploaded using protected internal endpoint (Org-isolated)
        resp = httpx.get(
            f"{go_backend_url}/internal/shipments/{shipment_id}/documents?org_id={org_id}",
            headers=internal_headers,
            timeout=5.0
        )
        if resp.status_code == 200:
            data = resp.json().get("data", {})
            existing_docs = data.get("documents", [])
            
            # Cross-check container number, weight, package counts against existing docs
            for doc in existing_docs:
                if doc["doc_type"] == doc_type:
                    continue # Skip comparing against self

                other_ext = doc.get("extracted_data") or {}
                if not other_ext:
                    continue

                # Cross weight mismatch check
                if "gross_weight" in extracted and "gross_weight" in other_ext:
                    w1 = extracted["gross_weight"]
                    w2 = other_ext["gross_weight"]
                    if w1 != w2 and w1 > 0 and w2 > 0:
                        discrepancies.append({
                            "field_name": "gross_weight",
                            "expected_value": str(w2),
                            "actual_value": str(w1),
                            "source_document": doc["doc_type"],
                            "target_document": doc_type,
                            "status": "OPEN"
                        })

                # Cross container number mismatch check
                if "container_number" in extracted and "container_number" in other_ext:
                    c1 = extracted["container_number"]
                    c2 = other_ext["container_number"]
                    if c1 != c2 and c1 != "" and c2 != "":
                        discrepancies.append({
                            "field_name": "container_number",
                            "expected_value": c2,
                            "actual_value": c1,
                            "source_document": doc["doc_type"],
                            "target_document": doc_type,
                            "status": "OPEN"
                        })

    except Exception as e:
        print(f"[Compliance Agent] Error querying internal records: {e}")

    status = "VERIFIED" if len(discrepancies) == 0 else "DISCREPANCY"

    return {
        **state,
        "discrepancies": discrepancies,
        "doc_status": status
    }
