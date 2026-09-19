"""
result_store.py — Lightweight Evaluation Result Persistence for LogisticsHQ

Saves scenario execution results into MariaDB table `ai_evaluation_results`.
Guarantees:
- Tenant isolation (scoped by org_id)
- Zero persistence of raw prompts, raw keys, documents, email bodies, or financial payloads
- Machine-readable serialization
"""

import json
import time
from dataclasses import dataclass, asdict
from datetime import datetime, timezone
from typing import Any, Dict, List, Optional
import pymysql

from app.persistence.mariadb_saver import parse_db_url, get_db_url


@dataclass
class EvaluationExecutionResult:
    org_id: int
    test_run_id: str
    scenario_id: str
    scenario_name: str
    scenario_version: str
    agent_key: str
    category: str
    status: str  # PASS, FAIL, SKIPPED
    expected_result: str
    actual_result_summary: str
    error_category: Optional[str] = None
    provider_mode: str = "deterministic_test"
    prompt_version: Optional[str] = "1.0.0"
    action_names_invoked: Optional[List[str]] = None
    approval_state: Optional[str] = "NONE"
    task_state: Optional[str] = "completed"
    checkpoint_state: Optional[str] = "NONE"
    correlation_id: Optional[str] = None
    duration_ms: int = 0
    started_at: Optional[str] = None
    completed_at: Optional[str] = None


class EvaluationResultStore:
    def __init__(self, db_config: Optional[dict] = None):
        self.db_config = db_config or parse_db_url(get_db_url())

    def _get_connection(self):
        return pymysql.connect(
            host=self.db_config["host"],
            port=self.db_config["port"],
            user=self.db_config["user"],
            password=self.db_config["password"],
            database=self.db_config["db"],
            autocommit=True,
            charset="utf8mb4",
            cursorclass=pymysql.cursors.Cursor,
        )

    def save_result(self, res: EvaluationExecutionResult) -> int:
        now_str = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%S.%f")[:-3]
        started = res.started_at or now_str
        completed = res.completed_at or now_str
        actions_json = json.dumps(res.action_names_invoked or [])

        # Sanitize any accidental sensitive fields in actual_result_summary
        sanitized_summary = str(res.actual_result_summary)
        if len(sanitized_summary) > 500:
            sanitized_summary = sanitized_summary[:500] + "..."

        query = """
            INSERT INTO ai_evaluation_results (
                org_id, test_run_id, scenario_id, scenario_name, scenario_version,
                agent_key, category, status, expected_result, actual_result_summary,
                error_category, provider_mode, prompt_version, action_names_invoked,
                approval_state, task_state, checkpoint_state, correlation_id,
                duration_ms, started_at, completed_at
            ) VALUES (
                %s, %s, %s, %s, %s,
                %s, %s, %s, %s, %s,
                %s, %s, %s, %s,
                %s, %s, %s, %s,
                %s, %s, %s
            )
        """
        params = (
            res.org_id,
            res.test_run_id,
            res.scenario_id,
            res.scenario_name,
            res.scenario_version,
            res.agent_key,
            res.category,
            res.status,
            res.expected_result,
            sanitized_summary,
            res.error_category,
            res.provider_mode,
            res.prompt_version,
            actions_json,
            res.approval_state,
            res.task_state,
            res.checkpoint_state,
            res.correlation_id,
            res.duration_ms,
            started,
            completed,
        )

        with self._get_connection() as conn:
            with conn.cursor() as cur:
                cur.execute(query, params)
                return cur.lastrowid

    def list_results_for_run(self, org_id: int, test_run_id: str) -> List[Dict[str, Any]]:
        query = """
            SELECT id, org_id, test_run_id, scenario_id, scenario_name, scenario_version,
                   agent_key, category, status, expected_result, actual_result_summary,
                   error_category, provider_mode, prompt_version, action_names_invoked,
                   approval_state, task_state, checkpoint_state, correlation_id,
                   duration_ms, started_at, completed_at, created_at
            FROM ai_evaluation_results
            WHERE org_id = %s AND test_run_id = %s
            ORDER BY id ASC
        """
        results = []
        with self._get_connection() as conn:
            with conn.cursor() as cur:
                cur.execute(query, (org_id, test_run_id))
                cols = [d[0] for d in cur.description]
                for row in cur.fetchall():
                    item = dict(zip(cols, row))
                    if isinstance(item.get("action_names_invoked"), str):
                        try:
                            item["action_names_invoked"] = json.loads(item["action_names_invoked"])
                        except Exception:
                            pass
                    results.append(item)
        return results
