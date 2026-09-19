import os
import json
import time
import random
import hashlib
import httpx
from typing import Optional, Dict, Any

from app.tools.auth_utils import get_internal_service_token

go_backend_url = os.getenv("GO_BACKEND_URL", "http://localhost:8080")
RETRYABLE_STATUS_CODES = {429, 502, 503, 504}

def derive_idempotency_key(
    action_name: str,
    org_id: int,
    input_data: Dict[str, Any],
    task_id: Optional[str] = None,
    thread_id: Optional[str] = None
) -> str:
    """
    Derives a stable idempotency key from task ID, thread ID, action name,
    organization ID, and primary business record ID.
    """
    rec_id = (
        input_data.get("rfq_id")
        or input_data.get("shipment_id")
        or input_data.get("customer_id")
        or input_data.get("lead_id")
        or input_data.get("document_id")
        or input_data.get("invoice_id")
        or input_data.get("milestone_code")
        or ""
    )
    raw = f"{org_id}:{action_name}:{task_id or ''}:{thread_id or ''}:{rec_id}"
    digest = hashlib.sha256(raw.encode("utf-8")).hexdigest()[:24]
    return f"idem_{action_name.replace('.', '_')}_{digest}"

def execute_action(
    action_name: str,
    org_id: int,
    input_data: Dict[str, Any],
    acting_user_id: int = 0,
    actor_type: str = "AI_AGENT",
    source: str = "langgraph",
    task_id: Optional[str] = None,
    thread_id: Optional[str] = None,
    idempotency_key: Optional[str] = None,
    is_confirmed: bool = False,
    timeout: float = 15.0,
    max_retries: int = 3
) -> Dict[str, Any]:
    """
    Authoritative client bridge that executes AI-triggered business actions
    through the Go backend's Centralized Action System.
    
    Guarantees:
    - Service key authentication
    - Organization scoping
    - Actor context preservation
    - RBAC enforcement
    - Human confirmation protection for high-risk actions
    - Stable idempotency across retries
    - Universal audit logging
    """
    if not idempotency_key:
        idempotency_key = derive_idempotency_key(action_name, org_id, input_data, task_id, thread_id)

    url = f"{go_backend_url}/internal/actions/execute"
    headers = {
        "X-LogisticsHQ-Service-Key": get_internal_service_token(),
        "Content-Type": "application/json"
    }
    payload = {
        "action_name": action_name,
        "org_id": org_id,
        "acting_user_id": acting_user_id,
        "actor_type": actor_type,
        "source": source,
        "task_id": task_id or "",
        "thread_id": thread_id or "",
        "idempotency_key": idempotency_key,
        "is_confirmed": is_confirmed,
        "input": input_data
    }

    last_error: Optional[Exception] = None
    for attempt in range(max_retries):
        try:
            resp = httpx.post(url, json=payload, headers=headers, timeout=timeout)
            
            # Handle structured responses
            try:
                data = resp.json()
            except Exception:
                data = {"raw_text": resp.text}

            if resp.status_code in (200, 400, 403, 404, 409):
                return data

            if resp.status_code in RETRYABLE_STATUS_CODES:
                wait_time = (2 ** attempt) + random.uniform(0.1, 0.5)
                print(f"[ActionBridge] Retryable status {resp.status_code} for {action_name}. Retrying in {wait_time:.2f}s...")
                time.sleep(wait_time)
                continue

            # Non-retryable error status
            return {
                "success": False,
                "action_name": action_name,
                "error": {
                    "type": "HttpError",
                    "message": f"Action bridge responded with HTTP {resp.status_code}: {resp.text[:200]}"
                }
            }

        except httpx.RequestError as e:
            last_error = e
            wait_time = (2 ** attempt) + random.uniform(0.1, 0.5)
            print(f"[ActionBridge] Network error calling {action_name}: {e}. Retrying in {wait_time:.2f}s...")
            time.sleep(wait_time)

    return {
        "success": False,
        "action_name": action_name,
        "error": {
            "type": "NetworkFailure",
            "message": f"Failed to connect to Centralized Action System: {str(last_error)}"
        }
    }
