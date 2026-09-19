import os
import httpx
from langchain_core.tools import tool
from app.tools.auth_utils import get_internal_service_token

# Retrieve the backend location from environment parameters.
go_backend_url = os.getenv("GO_BACKEND_URL", "http://localhost:8080")

@tool
def create_rfq_from_email_tool(org_id: int, customer_id: int, origin: str, destination: str, incoterms: str, target_date: str, items: list) -> str:
    """
    Create a new draft RFQ in the Go backend based on details parsed from a customer's inbound email.
    Executes through the Centralized Action System.
    
    Parameters:
      org_id (int): Organization ID.
      customer_id (int): The ID of the Lead (customer) associated with the email.
      origin (str): Standard 5-letter UN/LOCODE of origin port (e.g. "INNSA").
      destination (str): Standard 5-letter UN/LOCODE of destination port (e.g. "DEHAM").
      incoterms (str): Standard 3-letter Incoterms (e.g. "FOB", "CIF", "EXW").
      target_date (str): Expected cargo ready date formatted as YYYY-MM-DD.
      items (list): List of cargo items.
    """
    import json
    from app.tools.action_bridge import execute_action

    input_data = {
        "customer_id": customer_id,
        "origin": origin,
        "destination": destination,
        "incoterms": incoterms,
        "target_date": target_date,
        "items": items
    }

    res = execute_action(
        action_name="sales.create_rfq_from_email",
        org_id=org_id,
        input_data=input_data,
        source="langgraph.sales"
    )

    if res.get("success"):
        return json.dumps(res)
    err_msg = res.get("error", {}).get("message", "Unknown error")
    return f"Error: {err_msg}"
