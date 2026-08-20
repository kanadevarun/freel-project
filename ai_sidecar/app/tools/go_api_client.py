import os
import httpx
from langchain_core.tools import tool

# Simple meaning:
#   This file defines a "Tool" in Python. In an Agentic AI platform, agents cannot access
#   databases directly. Instead, they are given "Tools" that they can invoke when they
#   need to communicate with external APIs or look up transactional data.

# Retrieve the backend location and security keys from environment parameters.
go_backend_url = os.getenv("GO_BACKEND_URL", "http://localhost:8080")
service_token = os.getenv("INTERNAL_SERVICE_TOKEN", "internal-service-key-logisticshq")

@tool
def normalize_port_tool(query: str) -> str:
    """
    Lookup UN/LOCODE port information from the LogisticsHQ backend.
    Use this tool to resolve free-text port names (like 'Nhava Sheva' or 'Hamburg') 
    to standard 5-character LOCODEs (like 'INNSA' or 'DEHAM').
    
    Simple meaning:
      The AI agent calls this function during layout extraction validation to convert
      messy port names to standardized codes.
      
    Example:
      Input: "jnpt" -> Output: "Success: Port 'jnpt' normalized to standard UN/LOCODE 'INNSA'."
    """
    # Target endpoint on the Go backend
    url = f"{go_backend_url}/internal/ports/normalize"
    # Supply internal authentication headers to bypass public Cognito gates
    headers = {"X-LogisticsHQ-Service-Key": service_token}
    params = {"query": query}
    
    try:
        # Perform HTTP GET request
        response = httpx.get(url, params=params, headers=headers, timeout=5.0)
        
        if response.status_code == 200:
            data = response.json()
            normalized = data.get("normalized")
            is_known = data.get("is_known", False)
            if is_known:
                return f"Success: Port '{query}' normalized to standard UN/LOCODE '{normalized}'."
            else:
                return f"Warning: Port '{query}' resolved to '{normalized}' but is not registered in the system."
        elif response.status_code == 401:
            return "Error: Unauthorized calling Go backend ports normalizer tool. Check service token."
            
        return f"Error: Backend API responded with status {response.status_code}"
    except Exception as e:
        return f"Failed to connect to Go backend normalizer API: {str(e)}"

@tool
def search_ports_tool(query: str) -> str:
    """
    Search UN/LOCODE port information from the LogisticsHQ backend.
    Use this tool to find all matching port aliases and standard 5-character LOCODEs
    containing the search query (e.g. searching 'India' or 'DE').
    
    Simple meaning:
      The AI agent calls this function when it needs to search or check multiple port options
      matching a prefix or keyword.
      
    Example:
      Input: "Nhava" -> Output: "{'NHAVA SHEVA': 'INNSA', 'NHAVASHEVA': 'INNSA', 'NAVI MUMBAI': 'INNSA'}"
    """
    url = f"{go_backend_url}/internal/ports/search"
    headers = {"X-LogisticsHQ-Service-Key": service_token}
    params = {"query": query}
    
    try:
        response = httpx.get(url, params=params, headers=headers, timeout=5.0)
        if response.status_code == 200:
            return response.text
        elif response.status_code == 401:
            return "Error: Unauthorized calling Go backend ports search tool. Check service token."
        return f"Error: Backend API responded with status {response.status_code}"
    except Exception as e:
        return f"Failed to connect to Go backend search API: {str(e)}"
