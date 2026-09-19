"""
checkpointer.py — Abstract Checkpointer Factory for LogisticsHQ Agentic AI Workflows

Enforces strict separation between local development and production environments:
- Development: May use in-memory MemorySaver via LANGGRAPH_CHECKPOINTER=memory.
- Production: Strictly mandates a persistent checkpointer (e.g., LANGGRAPH_CHECKPOINTER=mariadb).
- Production startup fails loudly if MemorySaver is selected.
- Silent fallback to MemorySaver is strictly prohibited.
- Preserves MariaDB business database and AI task queue.
"""

import os
import sys
from typing import Optional, Dict, Any
from langgraph.checkpoint.base import BaseCheckpointSaver

# Singleton cache for checkpointer instance within the process
_cached_checkpointer: Optional[BaseCheckpointSaver] = None
_cached_checkpointer_type: Optional[str] = None


def get_checkpointer_type() -> str:
    """Returns the configured checkpointer type string ('mariadb', 'memory', etc.). Defaults to 'mariadb'."""
    return os.getenv("LANGGRAPH_CHECKPOINTER", "mariadb").strip().lower()


def get_environment() -> str:
    """Returns the application environment ('development', 'production', 'test', etc.)."""
    return os.getenv("APP_ENV", "development").strip().lower()


def validate_checkpointer_config() -> None:
    """
    Validates the checkpointer configuration against the current environment.
    
    Raises:
        RuntimeError: If APP_ENV=production and LANGGRAPH_CHECKPOINTER is 'memory' or invalid.
    """
    env = get_environment()
    ck_type = get_checkpointer_type()

    if env == "production":
        if ck_type == "memory":
            raise RuntimeError(
                "CRITICAL: Production startup failed. "
                "LANGGRAPH_CHECKPOINTER='memory' (MemorySaver) is not permitted in production (APP_ENV=production). "
                "In-memory checkpointer loses all workflow state (WAITING_FOR_HUMAN, PENDING_REVIEW) "
                "upon process restart, crash, redeployment, or worker replacement. "
                "A persistent checkpointer (e.g., LANGGRAPH_CHECKPOINTER=mariadb) is required for production."
            )
        elif ck_type not in ["mariadb"]:
            raise RuntimeError(
                f"CRITICAL: Unknown or non-persistent checkpointer '{ck_type}' configured for production. "
                "A verified persistent checkpointer backend (such as 'mariadb') is required."
            )
    else:
        # Development / test mode notice
        if ck_type == "memory":
            print(
                "[AI Sidecar Persistence] NOTICE: Using in-memory MemorySaver checkpointer for development "
                "(LANGGRAPH_CHECKPOINTER=memory). NOTE: Paused workflow states will NOT survive sidecar "
                "restarts, crashes, or worker replacements. Persistent storage is required for production."
            )


def get_checkpointer(force_refresh: bool = False) -> BaseCheckpointSaver:
    """
    Returns the configured checkpointer instance.
    
    Shares a single instance across all graphs in the process so they share the same
    in-memory storage (in development) or connection pool (in persistent mode).
    """
    global _cached_checkpointer, _cached_checkpointer_type

    validate_checkpointer_config()
    ck_type = get_checkpointer_type()

    if _cached_checkpointer is not None and _cached_checkpointer_type == ck_type and not force_refresh:
        return _cached_checkpointer

    if ck_type == "memory":
        from langgraph.checkpoint.memory import MemorySaver
        _cached_checkpointer = MemorySaver()
        _cached_checkpointer_type = "memory"
        print("[AI Sidecar Persistence] Initialized shared MemorySaver checkpointer.")
        return _cached_checkpointer

    elif ck_type == "mariadb":
        from app.persistence.mariadb_saver import get_checkpointer as get_mariadb_checkpointer
        _cached_checkpointer = get_mariadb_checkpointer()
        _cached_checkpointer_type = "mariadb"
        print("[AI Sidecar Persistence] Initialized persistent MariaDBSaver checkpointer.")
        return _cached_checkpointer

    else:
        raise ValueError(f"Unsupported LANGGRAPH_CHECKPOINTER type: '{ck_type}'. Supported: 'memory', 'mariadb'.")


def get_checkpointer_info() -> Dict[str, Any]:
    """Returns metadata describing the active checkpointer for health / telemetry."""
    env = get_environment()
    ck_type = get_checkpointer_type()
    is_persistent = (ck_type != "memory")
    return {
        "checkpointer": "MemorySaver" if ck_type == "memory" else "MariaDBSaver",
        "checkpointer_type": ck_type,
        "environment": env,
        "persistent": is_persistent,
        "production_ready": is_persistent and (env == "production" or ck_type == "mariadb"),
        "warning": None if is_persistent else "In-memory state will be lost on process restart"
    }
