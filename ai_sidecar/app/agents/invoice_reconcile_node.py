import json
import os
import httpx
from app.state.finance_state import FinanceState


def invoice_reconcile_node(state: FinanceState) -> FinanceState:
    """
    Deterministic Three-Way Reconciliation Node.
    Cross-checks invoice line items against agreed buy rates (from won quote)
    and active contracts (from rate entries).
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

    extracted_items = state.get("extracted_items", [])
    discrepancies = []

    print(f"[Finance Agent] Reconciling invoice {invoice_id} line items...")

    # Fetch dynamic service authorization token
    token = os.getenv("INTERNAL_SERVICE_TOKEN", "internal-service-key-logisticshq")
    go_backend_url = os.getenv("GO_BACKEND_URL", "http://localhost:8080")
    internal_headers = {
        "X-LogisticsHQ-Service-Key": token,
        "Content-Type": "application/json"
    }

    try:
        # Fetch shipment, quote, and active contract rates from internal Go endpoint
        resp = httpx.get(
            f"{go_backend_url}/internal/shipments/{shipment_id}/finance?org_id={org_id}",
            headers=internal_headers,
            timeout=5.0
        )
        resp.raise_for_status()
        
        workspace_data = resp.json().get("data", {})
        shipment = workspace_data.get("shipment") or {}
        quote = workspace_data.get("quote")
        contract_rates = workspace_data.get("contract_rates") or []

        # Build lookup maps for contracted rates/surcharges
        contracted_surcharges = {}
        contracted_ocean_freight = 0.0

        for rate in contract_rates:
            if rate.get("ocean_freight"):
                contracted_ocean_freight = float(rate["ocean_freight"])
            
            # Parse surcharges JSON
            surcharges_raw = rate.get("surcharges", "[]")
            try:
                surcharges_list = json.loads(surcharges_raw)
                for item in surcharges_list:
                    code = item.get("code")
                    amount = item.get("amount")
                    if code:
                        contracted_surcharges[code] = float(amount)
            except Exception as e:
                print(f"[Finance Agent] Error parsing contracted surcharges: {e}")

        # Reconcile each extracted item
        for item in extracted_items:
            charge_code = item.get("charge_code")
            total_amount = float(item.get("total_amount", 0.0))
            unit_price = float(item.get("unit_price", 0.0))

            if charge_code == "OCEAN_FREIGHT":
                # Check against won quote first, fallback to contract rates
                expected_ocean = contracted_ocean_freight
                source = "CONTRACT"
                
                if quote:
                    expected_ocean = float(quote.get("buy_price", 0.0))
                    source = "QUOTE"

                if expected_ocean > 0.0 and total_amount > expected_ocean:
                    discrepancies.append({
                        "charge_code": "OCEAN_FREIGHT",
                        "field_name": "total_amount",
                        "expected_value": f"{expected_ocean:.2f}",
                        "actual_value": f"{total_amount:.2f}",
                        "source": source,
                        "status": "OPEN"
                    })
            else:
                # Surcharge charge code comparison
                expected_price = contracted_surcharges.get(charge_code)

                if expected_price is None:
                    # Unauthorized charge (not present in contract or quote)
                    discrepancies.append({
                        "charge_code": charge_code,
                        "field_name": "charge_code",
                        "expected_value": "UNAUTHORIZED_CHARGE",
                        "actual_value": charge_code,
                        "source": "CONTRACT",
                        "status": "OPEN"
                    })
                elif unit_price > expected_price:
                    # Overcharge discrepancy
                    discrepancies.append({
                        "charge_code": charge_code,
                        "field_name": "unit_price",
                        "expected_value": f"{expected_price:.2f}",
                        "actual_value": f"{unit_price:.2f}",
                        "source": "CONTRACT",
                        "status": "OPEN"
                    })

    except Exception as e:
        print(f"[Finance Agent] Reconcile error: {e}")
        # Re-raise so the pipeline fails and queues can retry
        raise e

    invoice_status = "APPROVED" if len(discrepancies) == 0 else "DISCREPANCY"
    print(f"[Finance Agent] Reconcile finished. Status={invoice_status}, Discrepancies={len(discrepancies)}")

    return {
        **state,
        "discrepancies": discrepancies,
        "invoice_status": invoice_status,
    }
