"""
tracer.py — Structured AI Observability, Execution Tracing, and Audit Bridge for LogisticsHQ

Records:
- Lifecycle telemetry into MariaDB table ai_execution_traces
- Business-critical AI events into universal audit_logs table
- Secret redaction for API keys, tokens, and sensitive credentials
- Safe error category mapping
"""

import json
import time
import uuid
import pymysql
from typing import Dict, Any, Optional

from app.persistence.mariadb_saver import parse_db_url, get_db_url
from app.tools.llm_factory import redact_secrets, classify_error_category


def get_pymysql_connection():
    db_params = parse_db_url(get_db_url())
    return pymysql.connect(
        host=db_params["host"],
        port=db_params["port"],
        user=db_params["user"],
        password=db_params["password"],
        database=db_params["db"],
        autocommit=True,
    )


class AIObservabilityTracer:
    """Manages recording of AI execution traces and audit logs."""

    @staticmethod
    def record_lifecycle_event(
        org_id: int,
        event_type: str,
        workflow_name: str,
        task_id: Optional[int] = None,
        thread_id: Optional[str] = None,
        correlation_id: Optional[str] = None,
        actor_type: str = "AI_AGENT",
        actor_id: Optional[str] = None,
        details: Optional[Dict[str, Any]] = None,
        error_message: Optional[str] = None,
    ):
        """
        Records lifecycle events:
        - AI_EXECUTION_STARTED
        - PROMPT_RESOLVED
        - PROVIDER_REQUEST_ATTEMPTED
        - PROVIDER_FAILOVER
        - TOOL_ACTION_REQUESTED
        - TOOL_ACTION_COMPLETED
        - HUMAN_APPROVAL_REQUESTED
        - HUMAN_APPROVAL_COMPLETED
        - CHECKPOINT_RESUMED
        - TASK_RETRIED
        - TASK_COMPLETED
        - TASK_FAILED
        - AI_EXECUTION_CANCELLED
        """
        safe_details = {k: v for k, v in (details or {}).items() if "key" not in k.lower() and "token" not in k.lower()}
        sanitized_error = redact_secrets(error_message) if error_message else None

        try:
            conn = get_pymysql_connection()
            try:
                with conn.cursor() as cur:
                    result_status = "FAILED" if sanitized_error or "FAILED" in event_type else "SUCCESS"
                    cur.execute("""
                        INSERT INTO audit_logs
                        (org_id, actor_type, actor_name, action, module, resource_type, resource_id, result, error_message, description, metadata, created_at)
                        VALUES (%s, %s, %s, %s, 'AI_RUNTIME', 'ai_workflow', %s, %s, %s, %s, %s, NOW())
                    """, (
                        org_id,
                        actor_type,
                        actor_id or f"AI_{workflow_name.upper()}",
                        event_type,
                        str(task_id or thread_id or workflow_name),
                        result_status,
                        sanitized_error,
                        f"AI Lifecycle event '{event_type}' for workflow '{workflow_name}'",
                        json.dumps(safe_details) if safe_details else None,
                    ))
            finally:
                conn.close()
        except Exception as e:
            print(f"[Observability Tracer] Warning: Failed to record lifecycle event: {e}")

    @staticmethod
    def record_trace(
        org_id: int,
        workflow_name: str,
        request_id: Optional[str] = None,
        task_id: Optional[int] = None,
        thread_id: Optional[str] = None,
        correlation_id: Optional[str] = None,
        prompt_key: Optional[str] = None,
        prompt_version: Optional[str] = None,
        primary_provider: str = "gemini",
        primary_model: str = "gemini-1.5-flash",
        final_provider: str = "gemini",
        final_model: str = "gemini-1.5-flash",
        failover_occurred: bool = False,
        failover_reason: Optional[str] = None,
        is_mock: bool = False,
        status: str = "COMPLETED",
        duration_ms: Optional[int] = None,
        error_category: Optional[str] = None,
        error_message: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None,
    ) -> str:
        req_id = request_id or uuid.uuid4().hex[:12]
        sanitized_error = redact_secrets(error_message) if error_message else None
        safe_meta = {k: v for k, v in (metadata or {}).items() if "key" not in k.lower() and "token" not in k.lower()}

        try:
            conn = get_pymysql_connection()
            try:
                with conn.cursor() as cur:
                    cur.execute("""
                        INSERT INTO ai_execution_traces
                        (org_id, task_id, thread_id, request_id, correlation_id, workflow_name, prompt_key, prompt_version,
                         primary_provider, primary_model, final_provider, final_model, failover_occurred, failover_reason,
                         is_mock, status, duration_ms, error_category, error_message, execution_metadata, created_at, completed_at)
                        VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, NOW(), NOW())
                    """, (
                        org_id,
                        task_id,
                        thread_id,
                        req_id,
                        correlation_id,
                        workflow_name,
                        prompt_key,
                        prompt_version,
                        primary_provider,
                        primary_model,
                        final_provider,
                        final_model,
                        1 if failover_occurred else 0,
                        failover_reason,
                        1 if is_mock else 0,
                        status,
                        duration_ms,
                        error_category,
                        sanitized_error,
                        json.dumps(safe_meta),
                    ))

                    # If fatal failure, also record in universal audit_logs
                    if status == "FAILED":
                        cur.execute("""
                            INSERT INTO audit_logs
                            (org_id, actor_type, actor_name, action, module, resource_type, resource_id, result, error_message, description, created_at)
                            VALUES (%s, 'AI_AGENT', %s, 'AI_EXECUTION_FAILED', 'AI_RUNTIME', 'ai_execution', %s, 'FAILED', %s, %s, NOW())
                        """, (
                            org_id,
                            f"AI_{workflow_name.upper()}",
                            req_id,
                            sanitized_error,
                            f"AI execution failed for workflow '{workflow_name}' [{error_category or 'error'}]",
                        ))
            finally:
                conn.close()
        except Exception as e:
            print(f"[Observability Tracer] Warning: Failed to record trace: {e}")

        return req_id

