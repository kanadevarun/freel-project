import os
import time
import json
import asyncio
import traceback
import psycopg
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
    url = os.getenv("DB_URL", "postgres://user:password@localhost:5432/freel?sslmode=disable")
    if "host.docker.internal" in url:
        url = url.replace("host.docker.internal", "localhost")
    return url

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
                # Open an async connection to PostgreSQL
                async with await psycopg.AsyncConnection.connect(self.db_url) as conn:
                    # Select and claim a task atomically using SELECT FOR UPDATE SKIP LOCKED.
                    # This ensures that if multiple Python worker instances are running,
                    # only one worker will process a given task.
                    query = """
                    UPDATE ai_processing_tasks
                    SET status = 'PROCESSING', 
                        updated_at = NOW(), 
                        attempts = attempts + 1
                    WHERE id = (
                        SELECT id
                        FROM ai_processing_tasks
                        WHERE status = 'QUEUED' AND attempts < max_attempts
                        ORDER BY created_at ASC
                        LIMIT 1
                        FOR UPDATE SKIP LOCKED
                    )
                    RETURNING id, org_id, document_id, entity_type, entity_id, task_type, payload, attempts, max_attempts;
                    """
                    
                    async with conn.cursor() as cur:
                        await cur.execute(query)
                        row = await cur.fetchone()
                        
                        if row:
                            task_id, org_id, document_id, entity_type, entity_id, task_type, payload_raw, attempts, max_attempts = row
                            # Commit the transaction immediately to lock status as PROCESSING.
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
                                async with conn.cursor() as update_cur:
                                    await update_cur.execute(
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
                                    backoff_sec = 5 * attempts
                                    print(f"[AI Sidecar Worker] Retrying task {task_id} in {backoff_sec} seconds...")
                                    await asyncio.sleep(backoff_sec)
                                else:
                                    status_to_set = "FAILED"
                                    print(f"[AI Sidecar Worker] Task {task_id} exceeded max retries. Marking as FAILED.")
                                
                                async with conn.cursor() as err_cur:
                                    await err_cur.execute(
                                        "UPDATE ai_processing_tasks SET status = %s, last_error = %s, updated_at = NOW() WHERE id = %s",
                                        (status_to_set, error_msg, task_id)
                                    )
                                await conn.commit()
                        else:
                            # If no tasks are queued, close connection and wait
                            pass
                            
            except Exception as conn_err:
                print(f"[AI Sidecar Worker] Database connection or polling error: {conn_err}")
                
            # Sleep for 2 seconds before checking for new tasks again
            await asyncio.sleep(2.0)
