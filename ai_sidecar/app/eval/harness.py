"""
harness.py — Authoritative Deterministic AI Test Harness for LogisticsHQ

Provides an explicit test-only AI runtime layer that intercepts LLM invocations and provides:
- Deterministic provider responses
- Deterministic tool calls
- Deterministic structured outputs
- Deterministic failure injection (rate limit, timeout, unavailable, malformed JSON)
- Deterministic retry & failover simulation
- Strict prohibition against running in production environments
- Clear marking of all test data with `_eval_meta`
- Zero dependence on live external providers (Gemini/OpenAI), external web search, real emails, or carrier APIs.
"""

import os
import json
import time
from typing import Any, Dict, List, Optional, Callable
from langchain_core.messages import AIMessage, BaseMessage

from app.config.runtime_config import get_runtime_config, AIRuntimeConfig


class DeterministicChatModel:
    """
    Drop-in mock chat model for LangChain/LangGraph workflows during deterministic testing.
    Intercepts prompts, checks for registered scenario behaviors or returns deterministic baseline JSON.
    """

    def __init__(self, harness: "DeterministicTestHarness", tools: Optional[List[Any]] = None):
        self.harness = harness
        self.tools = tools or []

    def bind_tools(self, tools: List[Any], **kwargs) -> "DeterministicChatModel":
        return DeterministicChatModel(self.harness, tools=tools)

    def invoke(self, input: Any, config: Optional[Dict[str, Any]] = None, **kwargs) -> AIMessage:
        return self.harness.resolve_response(input, config=config, tools=self.tools)

    async def ainvoke(self, input: Any, config: Optional[Dict[str, Any]] = None, **kwargs) -> AIMessage:
        return self.invoke(input, config=config, **kwargs)


class DeterministicTestHarness:
    """
    Authoritative test-only runtime coordinator.
    Manages deterministic outputs, simulated failure injections, and execution tracking.
    """

    _instance: Optional["DeterministicTestHarness"] = None

    def __init__(self):
        self._enabled: bool = False
        self._active_scenario_id: Optional[str] = None
        self._registered_responses: Dict[str, Any] = {}
        self._failure_injections: Dict[str, Dict[str, Any]] = {}
        self._attempt_counters: Dict[str, int] = {}
        self._invocations_log: List[Dict[str, Any]] = []

    @classmethod
    def get_instance(cls) -> "DeterministicTestHarness":
        if cls._instance is None:
            cls._instance = DeterministicTestHarness()
        return cls._instance

    def enable(self, scenario_id: Optional[str] = None) -> None:
        """Enables deterministic test mode with strict production safeguard."""
        cfg = get_runtime_config()
        if cfg.environment == "production":
            raise RuntimeError(
                "CRITICAL SECURITY VIOLATION: Deterministic test harness is strictly "
                "prohibited in production environment (APP_ENV=production)."
            )
        self._enabled = True
        if scenario_id:
            self._active_scenario_id = scenario_id

    def disable(self) -> None:
        self._enabled = False
        self._active_scenario_id = None
        self._registered_responses.clear()
        self._failure_injections.clear()
        self._attempt_counters.clear()
        self._invocations_log.clear()

    def is_enabled(self) -> bool:
        return self._enabled

    def set_active_scenario(self, scenario_id: str) -> None:
        self._active_scenario_id = scenario_id

    def register_response(self, key_or_scenario: str, response: Any) -> None:
        """Registers a scenario-specific response or tool call structure."""
        self._registered_responses[key_or_scenario] = response

    def inject_failure(
        self,
        key_or_scenario: str,
        error_category: str = "provider_rate_limit",
        error_message: str = "Simulated provider error",
        fail_times: int = 1,
    ) -> None:
        """Configures a temporary or permanent failure injection for retry/failover testing."""
        self._failure_injections[key_or_scenario] = {
            "error_category": error_category,
            "error_message": error_message,
            "fail_times": fail_times,
            "failed_so_far": 0,
        }

    def get_chat_model(self, tools: Optional[List[Any]] = None) -> DeterministicChatModel:
        return DeterministicChatModel(self, tools=tools)

    def resolve_response(
        self, input_data: Any, config: Optional[Dict[str, Any]] = None, tools: Optional[List[Any]] = None
    ) -> AIMessage:
        prompt_text = str(input_data)
        scenario_id = self._active_scenario_id or "default"
        attempts = self._attempt_counters.get(scenario_id, 0) + 1
        self._attempt_counters[scenario_id] = attempts

        self._invocations_log.append({
            "scenario_id": scenario_id,
            "attempt": attempts,
            "timestamp": time.time(),
            "prompt_snippet": prompt_text[:200]
        })

        # Check for simulated failures
        if scenario_id in self._failure_injections:
            fail_cfg = self._failure_injections[scenario_id]
            if fail_cfg["failed_so_far"] < fail_cfg["fail_times"]:
                fail_cfg["failed_so_far"] += 1
                cat = fail_cfg["error_category"]
                msg = fail_cfg["error_message"]
                if cat == "malformed_json":
                    return AIMessage(
                        content="<<<INVALID_JSON>>>{'corrupted': true, missing_quotes}",
                        response_metadata={"is_test": True, "provider": "deterministic_test", "scenario_id": scenario_id}
                    )
                raise RuntimeError(f"[{cat.upper()}] {msg} (attempt {attempts})")

        # Check explicit registered response
        if scenario_id in self._registered_responses:
            val = self._registered_responses[scenario_id]
            if isinstance(val, AIMessage):
                return val
            if isinstance(val, dict):
                content = json.dumps(val)
                return AIMessage(
                    content=f"```json\n{content}\n```",
                    response_metadata={"is_test": True, "provider": "deterministic_test", "scenario_id": scenario_id}
                )
            return AIMessage(
                content=str(val),
                response_metadata={"is_test": True, "provider": "deterministic_test", "scenario_id": scenario_id}
            )

        # Baseline deterministic generator by domain
        return self._generate_baseline_response(prompt_text, scenario_id)

    def _generate_baseline_response(self, prompt_text: str, scenario_id: str) -> AIMessage:
        meta = {
            "is_test": True,
            "provider": "deterministic_test",
            "scenario_id": scenario_id,
            "generated_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
        }

        # Pricing analyst baseline
        if "pricing.analyst" in prompt_text or "Process pricing for RFQ" in prompt_text or "RFQ #" in prompt_text:
            quotes = [
                {
                    "carrier_name": "Hapag-Lloyd",
                    "transit_time_days": 18,
                    "buy_price": 2800.0,
                    "sell_price": 3220.0,
                    "is_recommended": True,
                    "reliability_score": 95,
                    "historical_success_rate": 0.96,
                    "ai_reasoning": "Deterministic standard margin rate applied."
                }
            ]
            content = f"```json\n{json.dumps(quotes, indent=2)}\n```\nOverall reasoning: Evaluated candidate trade lane rates."
            return AIMessage(content=content, response_metadata=meta)

        # Lead scoring baseline
        if "score_lead" in prompt_text or "lead_scoring" in prompt_text or "score" in prompt_text.lower():
            res = {
                "score": 85,
                "confidence": 92,
                "reasoning": "Strong shipper revenue and consistent trade volume profile.",
                "action_recommended": "APPROVE_OUTREACH",
                "fit_category": "TIER_1_ENTERPRISE",
                "_eval_meta": meta
            }
            return AIMessage(content=f"```json\n{json.dumps(res)}\n```", response_metadata=meta)

        # Outreach cold email baseline
        if "cold outreach email" in prompt_text.lower() or "outreach.cold_email" in prompt_text:
            res = {
                "subject": "Optimizing Ocean Freight for Your Trade Lanes",
                "body": "Hello,\n\nWe noticed your substantial shipping operations and would welcome the opportunity to benchmark your ocean freight rates.\n\nBest regards,\nLogisticsHQ Team",
                "target_persona": "VP Supply Chain",
                "_eval_meta": meta
            }
            return AIMessage(content=f"```json\n{json.dumps(res)}\n```", response_metadata=meta)

        # Operations / carrier update baseline
        if "tracking" in prompt_text.lower() or "milestone" in prompt_text.lower():
            res = {
                "milestone": "DEPARTED_PORT",
                "location": "INNSA",
                "timestamp": "2026-10-01T12:00:00Z",
                "has_exception": False,
                "_eval_meta": meta
            }
            return AIMessage(content=f"```json\n{json.dumps(res)}\n```", response_metadata=meta)

        # Contracts / Rate sheet extraction baseline
        if "rate-sheet" in prompt_text.lower() or "contract" in prompt_text.lower():
            res = {
                "contract_code": "CTR-2026-EVAL",
                "carrier_name": "Maersk Line",
                "effective_from": "2026-01-01",
                "effective_to": "2026-12-31",
                "rates": [
                    {
                        "origin_port": "INNSA",
                        "destination_port": "DEHAM",
                        "equipment_type": "40GP",
                        "rate_amount": 2950.0,
                        "currency": "USD"
                    }
                ],
                "_eval_meta": meta
            }
            return AIMessage(content=f"```json\n{json.dumps(res)}\n```", response_metadata=meta)

        # Default fallback
        res = {
            "status": "DETERMINISTIC_SUCCESS",
            "message": "Deterministic test response evaluated successfully.",
            "_eval_meta": meta
        }
        return AIMessage(content=f"```json\n{json.dumps(res)}\n```", response_metadata=meta)
