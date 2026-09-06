import os
import time
import json
import asyncio
import traceback
import aiomysql
from urllib.parse import urlparse
from typing import Dict, Any, Optional, Callable

# Simple meaning:
#   This file implements the "Task Worker Loop".
#   Instead of processing uploads immediately inside HTTP requests, this worker
#   continuously checks the database table 'ai_processing_tasks' for new work.
#   It locks a single task atomically (so other worker processes don't run it),
#   runs the LangGraph pipeline, sends the webhook callback, and updates the task status.
#
# Dependency Injection:
#   To prevent circular imports with main.py, we pass the execution handler functions
#   and validation classes when initializing the QueueWorker.

def get_db_url() -> str:
    """
    Returns a normalized PostgreSQL connection string.
    
    Simple meaning:
      Reads the database connection URL from the environment.
      If it contains 'host.docker.internal' (which is used when running inside docker),
      it translates it to 'localhost' so that it works when running directly on the host machine.
    """
    url = os.getenv("DB_URL", "mysql://root:@127.0.0.1:3306/freel_mysql")
    if "host.docker.internal" in url:
        url = url.replace("host.docker.internal", "127.0.0.1")
    return url

def parse_db_url(url: str):
    parsed = urlparse(url)
    return {
        "host": parsed.hostname or "127.0.0.1",
        "port": parsed.port or 3306,
        "user": parsed.username or "root",
        "password": parsed.password or "",
        "db": parsed.path.lstrip('/') or "freel_mysql"
    }

class QueueWorker:
    def __init__(self, run_process_fn: Callable, run_resume_fn: Callable, processing_request_cls, resume_request_cls, task_handlers: Optional[Dict[str, Callable]] = None):
        self.db_url = get_db_url()
        self.stop_event = asyncio.Event()
        self.task = None
        self.run_process_fn = run_process_fn
        self.run_resume_fn = run_resume_fn
        self.processing_request_cls = processing_request_cls
        self.resume_request_cls = resume_request_cls
        self.task_handlers = task_handlers or {}

    async def start(self):
        """Starts the background asyncio polling loop."""
        print("[AI Sidecar Worker] Starting database task queue worker loop...")
        self.stop_event.clear()
        self.task = asyncio.create_task(self._poll_loop())

    async def stop(self):
        """Stops the worker polling loop gracefully."""
        print("[AI Sidecar Worker] Shutting down database task queue worker...")
        self.stop_event.set()
        if self.task:
            try:
                await self.task
            except asyncio.CancelledError:
                pass
            print("[AI Sidecar Worker] Background task stopped.")

    async def _poll_loop(self):
        """
        Main polling loop.
        Checks for QUEUED tasks every 2 seconds, claims them, and executes them.
        """
        while not self.stop_event.is_set():
            try:
                db_params = parse_db_url(self.db_url)
                conn = await aiomysql.connect(
                    host=db_params["host"],
                    port=db_params["port"],
                    user=db_params["user"],
                    password=db_params["password"],
                    db=db_params["db"],
                    autocommit=False
                )
                try:
                    async with conn.cursor() as cur:
                        # Select and claim a task atomically using SELECT FOR UPDATE SKIP LOCKED
                        # In MySQL, we select the ID first, then update it.
                        select_query = """
                        SELECT id, org_id, document_id, entity_type, entity_id, task_type, payload, retry_count
                        FROM ai_processing_tasks
                        WHERE status = 'QUEUED' AND retry_count < %s
                        ORDER BY created_at ASC
                        LIMIT 1
                        FOR UPDATE SKIP LOCKED
                        """
                        max_attempts = 3
                        await cur.execute(select_query, (max_attempts,))
                        row = await cur.fetchone()
                        
                        if row:
                            task_id, org_id, document_id, entity_type, entity_id, task_type, payload_raw, attempts = row
                            
                            # Mark as processing
                            update_query = """
                            UPDATE ai_processing_tasks
                            SET status = 'PROCESSING', 
                                updated_at = NOW(), 
                                retry_count = retry_count + 1
                            WHERE id = %s
                            """
                            await cur.execute(update_query, (task_id,))
                            await conn.commit()
                            
                            print(f"[AI Sidecar Worker] Claimed task {task_id} (type: {task_type}) for entity {entity_type}:{entity_id}")
                            
                            # Parse payload from database JSON
                            payload = payload_raw if isinstance(payload_raw, dict) else json.loads(payload_raw)
                            
                            try:
                                # Execute the corresponding LangGraph pipeline task
                                if task_type in self.task_handlers:
                                    # Route to the registered custom task handler dynamically
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
                                
                                # Mark task status as COMPLETED in DB upon successful completion
                                await cur.execute(
                                    "UPDATE ai_processing_tasks SET status = 'COMPLETED', updated_at = NOW() WHERE id = %s",
                                    (task_id,)
                                )
                                await conn.commit()
                                print(f"[AI Sidecar Worker] Task {task_id} processed successfully.")
                                
                            except Exception as task_err:
                                # If processing fails, capture stack trace and determine retry status.
                                #
                                # Simple meaning:
                                #   If the processing crashed (e.g. Gemini rate limit, S3 download failed),
                                #   we check if we can try again. If the current attempts is less than max_attempts,
                                #   we reschedule it by setting status back to 'QUEUED'.
                                #   Otherwise, we mark it as permanently 'FAILED' with the error details.
                                error_msg = traceback.format_exc()
                                print(f"[AI Sidecar Worker] Task {task_id} failed on attempt {attempts}/{max_attempts}: {task_err}\n{error_msg}")
                                
                                if attempts < max_attempts:
                                    status_to_set = "QUEUED"
                                    new_status = "QUEUED"
                                    backoff_sec = 5 * attempts
                                    print(f"[AI Sidecar Worker] Retrying task {task_id} in {backoff_sec} seconds...")
                                    await asyncio.sleep(backoff_sec)
                                else:
                                    new_status = "FAILED"
                                    print(f"[AI Sidecar Worker] Task {task_id} exceeded max retries. Marking as FAILED.")
                                # Mark task status as FAILED or QUEUED
                                await cur.execute(
                                    "UPDATE ai_processing_tasks SET status = %s, error_message = %s, updated_at = NOW() WHERE id = %s",
                                    (new_status, str(task_err) + "\n" + error_msg, task_id)
                                )
                                await conn.commit()
                finally:
                    conn.close()
            except Exception as conn_err:
                print(f"[AI Sidecar Worker] Database connection or polling error: {conn_err}")
                
            # Sleep for 2 seconds before checking for new tasks again
            await asyncio.sleep(2.0)
