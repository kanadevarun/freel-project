import os
import time
import json
import socket
import uuid
import asyncio
import traceback
import aiomysql
import httpx
from urllib.parse import urlparse
from typing import Dict, Any, Optional, Callable, List

# Task Status Constants
STATUS_QUEUED = "QUEUED"
STATUS_PROCESSING = "PROCESSING"
STATUS_WAITING_FOR_APPROVAL = "WAITING_FOR_APPROVAL"
STATUS_RETRYING = "RETRYING"
STATUS_COMPLETED = "COMPLETED"
STATUS_FAILED = "FAILED"
STATUS_REJECTED = "REJECTED"
STATUS_CANCELLED = "CANCELLED"

TERMINAL_STATUSES = {STATUS_COMPLETED, STATUS_FAILED, STATUS_REJECTED, STATUS_CANCELLED}

VALID_TRANSITIONS = {
    STATUS_QUEUED: {STATUS_PROCESSING, STATUS_CANCELLED, STATUS_FAILED},
    STATUS_PROCESSING: {STATUS_COMPLETED, STATUS_WAITING_FOR_APPROVAL, STATUS_RETRYING, STATUS_FAILED, STATUS_CANCELLED},
    STATUS_WAITING_FOR_APPROVAL: {STATUS_PROCESSING, STATUS_COMPLETED, STATUS_REJECTED, STATUS_CANCELLED},
    STATUS_RETRYING: {STATUS_QUEUED, STATUS_CANCELLED, STATUS_FAILED},
    STATUS_COMPLETED: set(),
    STATUS_REJECTED: set(),
    STATUS_CANCELLED: set(),
    STATUS_FAILED: {STATUS_QUEUED}, # Only via explicit retry!
}

def validate_transition(from_status: str, to_status: str, is_explicit_retry: bool = False) -> bool:
    from_s = from_status.upper() if from_status else ""
    to_s = to_status.upper() if to_status else ""
    if from_s == to_s:
        return True
    if from_s == STATUS_FAILED and to_s == STATUS_QUEUED:
        return is_explicit_retry
    allowed = VALID_TRANSITIONS.get(from_s, set())
    return to_s in allowed

def is_retryable_error(exc: Exception) -> bool:
    """
    Classifies errors into retryable (transient infrastructure/rate limits)
    vs non-retryable (validation, auth, business logic, permanent failures).
    """
    if isinstance(exc, (ConnectionError, TimeoutError, asyncio.TimeoutError)):
        return True
    if isinstance(exc, (aiomysql.OperationalError, aiomysql.InternalError)):
        return True
    if isinstance(exc, (httpx.ConnectError, httpx.TimeoutException, httpx.NetworkError)):
        return True
    if isinstance(exc, httpx.HTTPStatusError):
        code = exc.response.status_code
        # 429 Too Many Requests, 502 Bad Gateway, 503 Service Unavailable, 504 Gateway Timeout
        if code in (429, 502, 503, 504):
            return True
        # Client errors are permanently non-retryable
        if 400 <= code < 500:
            return False

    err_str = str(exc).lower()
    if any(term in err_str for term in ["connection refused", "broken pipe", "timeout", "timed out", "rate limit", "temporarily unavailable"]):
        return True

    # Validation errors, missing keys, type errors are non-retryable
    if isinstance(exc, (ValueError, KeyError, TypeError, json.JSONDecodeError)):
        return False

    return False

from app.persistence.db_config import get_db_url, parse_db_url, mask_db_config, ALLOWED_AI_TABLES


class QueueWorker:
    def __init__(
        self,
        run_process_fn: Callable,
        run_resume_fn: Callable,
        processing_request_cls,
        resume_request_cls,
        task_handlers: Optional[Dict[str, Callable]] = None,
        lease_seconds: int = 300,
        heartbeat_interval: int = 25
    ):
        self.db_url = get_db_url()
        self.stop_event = asyncio.Event()
        self.poll_task = None
        self.run_process_fn = run_process_fn
        self.run_resume_fn = run_resume_fn
        self.processing_request_cls = processing_request_cls
        self.resume_request_cls = resume_request_cls
        self.task_handlers = task_handlers or {}
        self.lease_seconds = lease_seconds
        self.heartbeat_interval = heartbeat_interval
        self.worker_id = f"worker-{socket.gethostname()}-{uuid.uuid4().hex[:8]}"
        self.current_in_flight_task_id: Optional[int] = None
        self._heartbeat_task: Optional[asyncio.Task] = None

    async def get_connection(self):
        db_params = parse_db_url(self.db_url)
        return await aiomysql.connect(
            host=db_params["host"],
            port=db_params["port"],
            user=db_params["user"],
            password=db_params["password"],
            db=db_params["db"],
            autocommit=False
        )

    async def recover_stale_tasks(self):
        """
        Recovers tasks left in 'PROCESSING' state due to worker restart or crash.
        Checks lease_expires_at against database clock.
        Requeues retryable tasks (< max_retries) to 'QUEUED' and marks exhausted as 'FAILED'.
        """
        try:
            conn = await self.get_connection()
            try:
                async with conn.cursor(aiomysql.DictCursor) as cur:
                    # Select stale tasks where lease expired
                    stale_query = """
                    SELECT id, org_id, task_type, retry_count, max_retries, error_message
                    FROM ai_processing_tasks
                    WHERE status = 'PROCESSING'
                      AND lease_expires_at IS NOT NULL
                      AND lease_expires_at < NOW()
                    """
                    await cur.execute(stale_query)
                    stale_rows = await cur.fetchall()

                    recovered_cnt = 0
                    failed_cnt = 0

                    for r in stale_rows:
                        task_id = r["id"]
                        attempts = r.get("retry_count", 0)
                        max_retries = r.get("max_retries") or 3

                        if attempts < max_retries:
                            await cur.execute("""
                                UPDATE ai_processing_tasks
                                SET status = 'QUEUED',
                                    worker_id = NULL,
                                    lease_expires_at = NULL,
                                    available_at = NOW(),
                                    error_message = CONCAT(COALESCE(error_message, ''), '\n[Recovery] Lease expired; reset to QUEUED for retry.'),
                                    updated_at = NOW()
                                WHERE id = %s AND status = 'PROCESSING'
                            """, (task_id,))
                            recovered_cnt += 1
                        else:
                            await cur.execute("""
                                UPDATE ai_processing_tasks
                                SET status = 'FAILED',
                                    last_error_code = 'LEASE_EXPIRED',
                                    worker_id = NULL,
                                    lease_expires_at = NULL,
                                    completed_at = NOW(),
                                    error_message = CONCAT(COALESCE(error_message, ''), '\n[Recovery] Task failed: lease expired and max retries exceeded.'),
                                    updated_at = NOW()
                                WHERE id = %s AND status = 'PROCESSING'
                            """, (task_id,))
                            failed_cnt += 1

                    await conn.commit()
                    if recovered_cnt > 0 or failed_cnt > 0:
                        print(f"[AI Sidecar Worker] Stale task recovery: {recovered_cnt} reset to QUEUED, {failed_cnt} marked FAILED.")
            finally:
                conn.close()
        except Exception as e:
            print(f"[AI Sidecar Worker] Stale task recovery warning: {e}")

    async def _heartbeat_loop(self, task_id: int):
        """Sends periodic heartbeats to maintain active lease while long LLM operations execute."""
        try:
            while not self.stop_event.is_set() and self.current_in_flight_task_id == task_id:
                await asyncio.sleep(self.heartbeat_interval)
                if self.current_in_flight_task_id != task_id:
                    break
                try:
                    conn = await self.get_connection()
                    try:
                        async with conn.cursor() as cur:
                            await cur.execute("""
                                UPDATE ai_processing_tasks
                                SET heartbeat_at = NOW(),
                                    lease_expires_at = DATE_ADD(NOW(), INTERVAL %s SECOND),
                                    updated_at = NOW()
                                WHERE id = %s AND status = 'PROCESSING' AND worker_id = %s
                            """, (self.lease_seconds, task_id, self.worker_id))
                            await conn.commit()
                    finally:
                        conn.close()
                except Exception as hb_err:
                    print(f"[AI Sidecar Worker] Heartbeat error for task #{task_id}: {hb_err}")
        except asyncio.CancelledError:
            pass

    async def _claim_task(self):
        """
        Atomically claims the oldest eligible task from ai_processing_tasks using
        SELECT FOR UPDATE SKIP LOCKED and releases the lock immediately.
        """
        conn = await self.get_connection()
        try:
            async with conn.cursor(aiomysql.DictCursor) as cur:
                supported_types = list(self.task_handlers.keys()) + ["PROCESS", "RESUME"]
                if supported_types:
                    placeholders = ", ".join(["%s"] * len(supported_types))
                    type_filter = f"AND task_type IN ({placeholders})"
                    params = tuple(supported_types)
                else:
                    type_filter = ""
                    params = ()

                select_query = f"""
                SELECT id, org_id, document_id, entity_type, entity_id, task_type, payload, retry_count, max_retries, thread_id, correlation_id
                FROM ai_processing_tasks
                WHERE (status = 'QUEUED' OR (status = 'RETRYING' AND (available_at IS NULL OR available_at <= NOW())))
                  AND (available_at IS NULL OR available_at <= NOW())
                  AND retry_count < max_retries
                  {type_filter}
                ORDER BY created_at ASC
                LIMIT 1
                FOR UPDATE SKIP LOCKED
                """
                await cur.execute(select_query, params)
                row = await cur.fetchone()

                if not row:
                    return None

                task_id = row["id"]
                update_query = """
                UPDATE ai_processing_tasks
                SET status = 'PROCESSING',
                    worker_id = %s,
                    started_at = NOW(),
                    heartbeat_at = NOW(),
                    lease_expires_at = DATE_ADD(NOW(), INTERVAL %s SECOND),
                    retry_count = retry_count + 1,
                    updated_at = NOW()
                WHERE id = %s AND (status = 'QUEUED' OR status = 'RETRYING')
                """
                await cur.execute(update_query, (self.worker_id, self.lease_seconds, task_id))
                await conn.commit()
                return row
        finally:
            conn.close()

    async def _get_current_task_status(self, task_id: int) -> Optional[str]:
        """Queries the current status of the task from the database."""
        conn = await self.get_connection()
        try:
            async with conn.cursor() as cur:
                await cur.execute("SELECT status FROM ai_processing_tasks WHERE id = %s", (task_id,))
                row = await cur.fetchone()
                return row[0] if row else None
        finally:
            conn.close()

    async def _complete_task(self, task_id: int):
        """
        Marks task COMPLETED only if current state is still PROCESSING.
        If task transitioned to WAITING_FOR_APPROVAL or CANCELLED, preserves that status.
        """
        current_status = await self._get_current_task_status(task_id)
        if current_status == STATUS_CANCELLED:
            print(f"[AI Sidecar Worker] Task #{task_id} was CANCELLED during execution. Suppressing COMPLETED transition.")
            return
        if current_status == STATUS_WAITING_FOR_APPROVAL:
            print(f"[AI Sidecar Worker] Task #{task_id} is WAITING_FOR_APPROVAL. Preserving approval gate.")
            return

        conn = await self.get_connection()
        try:
            async with conn.cursor() as cur:
                await cur.execute("""
                    UPDATE ai_processing_tasks
                    SET status = 'COMPLETED',
                        completed_at = NOW(),
                        lease_expires_at = NULL,
                        updated_at = NOW()
                    WHERE id = %s AND status = 'PROCESSING'
                """, (task_id,))
                await conn.commit()
                print(f"[AI Sidecar Worker] Task #{task_id} completed successfully.")
        finally:
            conn.close()

    async def _handle_task_failure(self, task_id: int, attempts: int, max_retries: int, exc: Exception):
        """
        Handles task failure by checking retryability and retry count bounds.
        Applies exponential backoff for retryable errors.
        """
        current_status = await self._get_current_task_status(task_id)
        if current_status == STATUS_CANCELLED:
            print(f"[AI Sidecar Worker] Task #{task_id} is already CANCELLED. Suppressing failure transition.")
            return

        retryable = is_retryable_error(exc)
        error_msg = f"{type(exc).__name__}: {str(exc)}\n{traceback.format_exc()}"
        last_error_code = type(exc).__name__

        conn = await self.get_connection()
        try:
            async with conn.cursor() as cur:
                if retryable and attempts < max_retries:
                    backoff_sec = min(120, (2 ** attempts) * 5)
                    print(f"[AI Sidecar Worker] Task #{task_id} transient error. Scheduling retry #{attempts}/{max_retries} in {backoff_sec}s.")
                    await cur.execute("""
                        UPDATE ai_processing_tasks
                        SET status = 'RETRYING',
                            available_at = DATE_ADD(NOW(), INTERVAL %s SECOND),
                            worker_id = NULL,
                            lease_expires_at = NULL,
                            last_error_code = %s,
                            error_message = %s,
                            updated_at = NOW()
                        WHERE id = %s
                    """, (backoff_sec, last_error_code, error_msg, task_id))
                else:
                    code = "MAX_RETRIES_EXCEEDED" if attempts >= max_retries else last_error_code
                    print(f"[AI Sidecar Worker] Task #{task_id} permanently failed (retryable={retryable}, attempts={attempts}/{max_retries}). Marking FAILED.")
                    await cur.execute("""
                        UPDATE ai_processing_tasks
                        SET status = 'FAILED',
                            completed_at = NOW(),
                            worker_id = NULL,
                            lease_expires_at = NULL,
                            last_error_code = %s,
                            error_message = %s,
                            updated_at = NOW()
                        WHERE id = %s
                    """, (code, error_msg, task_id))
                await conn.commit()
        finally:
            conn.close()

    async def _execute_task(self, row: dict):
        """Executes a claimed task and manages heartbeat and completion."""
        task_id = row["id"]
        org_id = row["org_id"]
        document_id = row.get("document_id")
        entity_type = row.get("entity_type")
        entity_id = row.get("entity_id")
        task_type = row["task_type"]
        attempts = row.get("retry_count", 1)
        max_retries = row.get("max_retries") or 3
        payload_raw = row.get("payload")

        payload = payload_raw if isinstance(payload_raw, dict) else (json.loads(payload_raw) if payload_raw else {})

        self.current_in_flight_task_id = task_id
        self._heartbeat_task = asyncio.create_task(self._heartbeat_loop(task_id))

        print(f"[AI Sidecar Worker] [{self.worker_id}] Executing Task #{task_id} (type: {task_type}, entity: {entity_type}:{entity_id}, attempt: {attempts}/{max_retries})")

        try:
            if task_type in self.task_handlers:
                await self.task_handlers[task_type](org_id, entity_id, payload)
            elif task_type == "PROCESS":
                req = self.processing_request_cls(
                    document_id=str(document_id),
                    org_id=int(org_id),
                    s3_key=payload.get("s3_key", ""),
                    file_type=payload.get("file_type", "PDF"),
                    callback_url=payload.get("callback_url", ""),
                    correlation_id=payload.get("correlation_id")
                )
                await self.run_process_fn(req)
            elif task_type == "RESUME":
                req = self.resume_request_cls(
                    document_id=str(document_id),
                    org_id=int(org_id),
                    action=payload.get("action", "APPROVE"),
                    corrected_rates=payload.get("corrected_rates"),
                    notes=payload.get("notes"),
                    callback_url=payload.get("callback_url", ""),
                    correlation_id=payload.get("correlation_id")
                )
                await self.run_resume_fn(req)
            else:
                raise ValueError(f"Unknown task_type '{task_type}'")

            await self._complete_task(task_id)

        except Exception as task_err:
            await self._handle_task_failure(task_id, attempts, max_retries, task_err)

        finally:
            self.current_in_flight_task_id = None
            if self._heartbeat_task:
                self._heartbeat_task.cancel()
                try:
                    await self._heartbeat_task
                except asyncio.CancelledError:
                    pass
                self._heartbeat_task = None

    async def start(self):
        """Starts the worker, recovering stale tasks and launching the poll loop."""
        print(f"[AI Sidecar Worker] Initializing worker '{self.worker_id}'...")
        await self.recover_stale_tasks()
        self.stop_event.clear()
        self.poll_task = asyncio.create_task(self._poll_loop())

    async def stop(self):
        """Gracefully terminates the worker without interrupting in-flight operations prematurely."""
        print(f"[AI Sidecar Worker] Graceful shutdown initiated for worker '{self.worker_id}'...")
        self.stop_event.set()

        # Await poll loop termination
        if self.poll_task:
            try:
                await asyncio.wait_for(asyncio.shield(self.poll_task), timeout=5.0)
            except (asyncio.TimeoutError, asyncio.CancelledError):
                self.poll_task.cancel()

        # Cancel heartbeat if active
        if self._heartbeat_task:
            self._heartbeat_task.cancel()

        print(f"[AI Sidecar Worker] Worker '{self.worker_id}' stopped safely.")

    async def _poll_loop(self):
        """Continuously polls database for eligible tasks with backoff when queue is idle."""
        recovery_timer = 0
        while not self.stop_event.is_set():
            try:
                # Run stale task recovery once every 60 seconds
                recovery_timer += 2
                if recovery_timer >= 60:
                    recovery_timer = 0
                    await self.recover_stale_tasks()

                claimed = await self._claim_task()
                if claimed:
                    await self._execute_task(claimed)
                    # Yield briefly to event loop
                    await asyncio.sleep(0.05)
                else:
                    await asyncio.sleep(2.0)
            except asyncio.CancelledError:
                break
            except Exception as loop_err:
                print(f"[AI Sidecar Worker] Polling error: {loop_err}")
                await asyncio.sleep(2.0)
