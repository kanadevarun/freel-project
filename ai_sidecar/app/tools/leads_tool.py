import os
import httpx
from langchain_core.tools import tool

# Retrieve the backend location and security keys from environment parameters.
go_backend_url = os.getenv("GO_BACKEND_URL", "http://localhost:8080")
service_token = os.getenv("INTERNAL_SERVICE_TOKEN", "internal-service-key-logisticshq")

@tool
def create_rfq_from_email_tool(org_id: int, customer_id: int, origin: str, destination: str, incoterms: str, target_date: str, items: list) -> str:
    """
    Create a new draft RFQ in the Go backend based on details parsed from a customer's inbound email.
    
    Parameters:
      org_id (int): Organization ID (typically 5).
      customer_id (int): The ID of the Lead (customer) associated with the email.
      origin (str): Standard 5-letter UN/LOCODE of origin port (e.g. "INNSA").
      destination (str): Standard 5-letter UN/LOCODE of destination port (e.g. "DEHAM").
      incoterms (str): Standard 3-letter Incoterms (e.g. "FOB", "CIF", "EXW").
      target_date (str): Expected cargo ready date formatted as YYYY-MM-DD.
      items (list): List of cargo items, each matching:
        {
          "description": "Text description of cargo",
          "quantity": 1,
          "weight_kg": 5000.0,
          "volume_cbm": 12.5
        }
    """
    url = f"{go_backend_url}/internal/rfqs/from-email"
    headers = {
        "X-LogisticsHQ-Service-Key": service_token,
        "Content-Type": "application/json"
    }
    payload = {
        "org_id": org_id,
        "customer_id": customer_id,
        "origin": origin,
        "destination": destination,
        "incoterms": incoterms,
        "target_date": target_date,
        "items": items
    }
    
    try:
        response = httpx.post(url, json=payload, headers=headers, timeout=10.0)
        if response.status_code == 200:
            return response.text
        return f"Error: Backend API responded with status {response.status_code}: {response.text}"
    except Exception as e:
        return f"Failed to connect to Go backend RFQ creation tool: {str(e)}"
