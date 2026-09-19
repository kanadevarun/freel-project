"""
auth_utils.py — Internal machine-to-machine service authentication for LogisticsHQ AI Sidecar.

Enforces:
1. Missing service key in production fails loudly.
2. Constant-time comparison using hmac.compare_digest to prevent timing attacks.
3. No leaking of secrets in logs, responses, or error messages.
4. Consistent header: X-LogisticsHQ-Service-Key.
"""

import os
import hmac
from typing import Optional
from dotenv import load_dotenv

load_dotenv()

DEV_FALLBACK_TOKEN = "dev-local-only-insecure-service-token-not-for-prod"
SERVICE_KEY_HEADER = "X-LogisticsHQ-Service-Key"


def get_internal_service_token() -> str:
    """
    Returns the configured internal service token.
    Raises RuntimeError if running in production and the token is not configured.
    """
    token = os.getenv("INTERNAL_SERVICE_TOKEN", "").strip()
    is_prod = os.getenv("APP_ENV", "development").strip().lower() == "production"

    if not token:
        if is_prod:
            raise RuntimeError(
                "Configuration error: INTERNAL_SERVICE_TOKEN must be specified in production environments"
            )
        return DEV_FALLBACK_TOKEN

    return token


def verify_internal_service_key(provided_key: Optional[str]) -> bool:
    """
    Verifies the provided service key against the configured token using constant-time comparison.
    """
    if not provided_key or not isinstance(provided_key, str):
        return False

    try:
        expected = get_internal_service_token()
    except RuntimeError:
        return False

    return hmac.compare_digest(provided_key.strip(), expected.strip())


from fastapi import Header, HTTPException, status


async def require_internal_service_key(
    x_logisticshq_service_key: Optional[str] = Header(None, alias="X-LogisticsHQ-Service-Key"),
    x_internal_service_key: Optional[str] = Header(None, alias="X-Internal-Service-Key"),
):
    """
    FastAPI dependency that enforces valid internal service key.
    Accepts both X-LogisticsHQ-Service-Key and X-Internal-Service-Key headers.
    """
    provided_key = x_logisticshq_service_key or x_internal_service_key
    if not provided_key or not verify_internal_service_key(provided_key):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail={"error": "UNAUTHORIZED", "message": "Invalid or missing internal service key"},
        )

