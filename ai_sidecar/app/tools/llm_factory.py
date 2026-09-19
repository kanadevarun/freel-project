"""
llm_factory.py — Authoritative AI Runtime & Provider Failover Foundation for LogisticsHQ AI Sidecar

Enforces:
- Single source of truth via AIRuntimeConfig
- Transparent failover between Primary (Gemini) and Fallback (OpenAI)
- Consistent content normalization across SDK versions (string vs content blocks)
- Structured execution telemetry and observability
- Strict prohibition against silent mock fallback in production
"""

import os
import time
import uuid
import traceback
from typing import Any, Dict, List, Optional
from langchain_core.messages import AIMessage, BaseMessage
from langchain_core.language_models.chat_models import BaseChatModel

from app.config.runtime_config import get_runtime_config, AIRuntimeConfig

try:
    from langchain_google_genai import ChatGoogleGenerativeAI
except ImportError:
    ChatGoogleGenerativeAI = None  # type: ignore

try:
    from langchain_openai import ChatOpenAI
except ImportError:
    ChatOpenAI = None  # type: ignore


def normalize_ai_content(content: Any) -> str:
    """
    Normalizes diverse LLM output content formats (string vs list of content blocks)
    into a clean, consistent plain text string.
    """
    if isinstance(content, str):
        return content
    if isinstance(content, list):
        parts = []
        for block in content:
            if isinstance(block, str):
                parts.append(block)
            elif isinstance(block, dict):
                text = block.get("text")
                if text:
                    parts.append(str(text))
            elif hasattr(block, "text"):
                parts.append(str(block.text))
        return "\n".join(parts) if parts else str(content)
    return str(content) if content is not None else ""


def redact_secrets(text: Any) -> str:
    """Sanitizes sensitive tokens, keys, and credentials from logs."""
    import re
    if not text:
        return ""
    s = str(text)
    s = re.sub(r"sk-[a-zA-Z0-9_\-]{15,}", "[REDACTED_OPENAI_KEY]", s)
    s = re.sub(r"AIza[0-9A-Za-z-_]{10,}", "[REDACTED_GEMINI_KEY]", s)
    s = re.sub(r"lhq_sec_[a-zA-Z0-9_\-]{15,}", "[REDACTED_SECRET]", s)
    return s


def classify_error_category(exc_or_msg: Any) -> str:
    """Classifies exceptions or error strings into the standard error taxonomy."""
    msg = str(exc_or_msg).lower()
    if "isolation" in msg or "tenant" in msg:
        return "organization_isolation_error"
    if "429" in msg or "rate limit" in msg or "quota" in msg or "insufficient_quota" in msg:
        return "provider_rate_limit"
    if "timeout" in msg or "deadline" in msg or "504" in msg:
        return "provider_timeout"
    if "502" in msg or "503" in msg or "unavailable" in msg or "connection refused" in msg:
        return "provider_unavailable"
    if "401" in msg or "unauthorized" in msg or "invalid api key" in msg or "api key" in msg:
        return "authentication_error"
    if "403" in msg or "forbidden" in msg:
        return "authorization_error"
    if "validation" in msg or "value_error" in msg:
        return "validation_error"
    return "unknown_error"


class UnifiedFailoverChatModel:
    """
    Authoritative chat model wrapper providing observable failover between primary
    and fallback providers, with strict environment-aware mock fallback policy.
    """

    def __init__(
        self,
        primary: Any,
        fallback: Optional[Any] = None,
        config: Optional[AIRuntimeConfig] = None,
    ):
        self.primary = primary
        self.fallback = fallback
        self.config = config or get_runtime_config()

    def _normalize_response(self, response: Any) -> Any:
        if isinstance(response, AIMessage):
            response.content = normalize_ai_content(response.content)
        return response

    def _generate_dev_mock_response(self, input_data: Any) -> AIMessage:
        """
        Generates a development/test mock response ONLY when allow_mock_fallback is True.
        """
        import re
        input_str = ""
        if isinstance(input_data, list):
            input_str = "\n".join([str(getattr(m, "content", m)) for m in input_data])
        else:
            input_str = str(input_data)

        origin = "INNSA"
        dest = "DEHAM"
        orig_match = re.search(r"Origin:\s*([A-Za-z0-9_-]+)", input_str, re.IGNORECASE)
        dest_match = re.search(r"Destination:\s*([A-Za-z0-9_-]+)", input_str, re.IGNORECASE)
        if orig_match:
            origin = orig_match.group(1).strip()
        if dest_match:
            dest = dest_match.group(1).strip()

        buy = 2800.0 if origin == "INNSA" and dest == "DEHAM" else 3200.0
        sell = round(buy * 1.15, 2)
        mock_json = f"""```json
[
  {{
    "carrier_name": "Maersk",
    "transit_time_days": 18,
    "buy_price": {buy},
    "sell_price": {sell},
    "is_recommended": true,
    "reliability_score": 95,
    "historical_success_rate": 0.96,
    "ai_reasoning": "Lane contract rate applied with standard margin."
  }}
]
```
Overall reasoning: Evaluated available trade lane rates for {origin} to {dest}."""
        return AIMessage(content=mock_json)

    def invoke(self, input: Any, config: Optional[Dict[str, Any]] = None, **kwargs) -> Any:
        start_time = time.time()
        req_id = uuid.uuid4().hex[:8]
        primary_err = None
        failover_err = None

        # 1. Attempt Primary Provider
        if self.primary is not None:
            try:
                raw_resp = self.primary.invoke(input, config=config, **kwargs)
                if isinstance(raw_resp, AIMessage):
                    if not hasattr(raw_resp, "response_metadata") or raw_resp.response_metadata is None:
                        raw_resp.response_metadata = {}
                    raw_resp.response_metadata["primary_provider"] = self.config.primary_provider
                    raw_resp.response_metadata["final_provider"] = self.config.primary_provider
                    raw_resp.response_metadata["failover_occurred"] = False
                return self._normalize_response(raw_resp)
            except Exception as e1:
                primary_err = e1
                err_cat = classify_error_category(e1)
                safe_err = redact_secrets(str(e1))
                print(f"[AI Runtime] [{req_id}] Primary ({self.config.primary_provider}) failed [{err_cat}]: {safe_err}")

        # 2. Attempt Failover Provider
        if self.fallback is not None:
            try:
                print(f"[AI Runtime] [{req_id}] Initiating failover to '{self.config.failover_provider}' (model: {self.config.failover_model})...")
                raw_resp = self.fallback.invoke(input, config=config, **kwargs)
                if isinstance(raw_resp, AIMessage):
                    if not hasattr(raw_resp, "response_metadata") or raw_resp.response_metadata is None:
                        raw_resp.response_metadata = {}
                    raw_resp.response_metadata["primary_provider"] = self.config.primary_provider
                    raw_resp.response_metadata["final_provider"] = self.config.failover_provider
                    raw_resp.response_metadata["failover_occurred"] = True
                    raw_resp.response_metadata["failover_reason"] = redact_secrets(str(primary_err))
                return self._normalize_response(raw_resp)
            except Exception as e2:
                failover_err = e2
                err_cat = classify_error_category(e2)
                safe_err = redact_secrets(str(e2))
                print(f"[AI Runtime] [{req_id}] Failover provider ({self.config.failover_provider}) failed [{err_cat}]: {safe_err}")

        # 3. Check Mock Policy (Non-production test environments ONLY)
        if self.config.allow_mock_fallback and self.config.environment != "production":
            print(f"[AI Runtime] [{req_id}] [MOCK] Both live providers failed. Serving development mock response (allow_mock_fallback=True).")
            return self._generate_dev_mock_response(input)

        # 4. Production or Strict Mode: Fail fast loudly with classified diagnostics
        final_msg = (
            f"All configured AI providers failed. Mock fallback is disabled. "
            f"Primary ({self.config.primary_provider}): {redact_secrets(str(primary_err))}; "
            f"Fallback ({self.config.failover_provider}): {redact_secrets(str(failover_err))}"
        )
        raise RuntimeError(final_msg)

    async def ainvoke(self, input: Any, config: Optional[Dict[str, Any]] = None, **kwargs) -> Any:
        start_time = time.time()
        req_id = uuid.uuid4().hex[:8]
        primary_err = None
        failover_err = None

        # 1. Attempt Primary Provider async
        if self.primary is not None:
            try:
                if hasattr(self.primary, "ainvoke"):
                    raw_resp = await self.primary.ainvoke(input, config=config, **kwargs)
                else:
                    raw_resp = self.primary.invoke(input, config=config, **kwargs)
                if isinstance(raw_resp, AIMessage):
                    if not hasattr(raw_resp, "response_metadata") or raw_resp.response_metadata is None:
                        raw_resp.response_metadata = {}
                    raw_resp.response_metadata["primary_provider"] = self.config.primary_provider
                    raw_resp.response_metadata["final_provider"] = self.config.primary_provider
                    raw_resp.response_metadata["failover_occurred"] = False
                return self._normalize_response(raw_resp)
            except Exception as e1:
                primary_err = e1
                err_cat = classify_error_category(e1)
                safe_err = redact_secrets(str(e1))
                print(f"[AI Runtime] [{req_id}] Primary async ({self.config.primary_provider}) failed [{err_cat}]: {safe_err}")

        # 2. Attempt Failover Provider async
        if self.fallback is not None:
            try:
                print(f"[AI Runtime] [{req_id}] Initiating async failover to '{self.config.failover_provider}' (model: {self.config.failover_model})...")
                if hasattr(self.fallback, "ainvoke"):
                    raw_resp = await self.fallback.ainvoke(input, config=config, **kwargs)
                else:
                    raw_resp = self.fallback.invoke(input, config=config, **kwargs)
                if isinstance(raw_resp, AIMessage):
                    if not hasattr(raw_resp, "response_metadata") or raw_resp.response_metadata is None:
                        raw_resp.response_metadata = {}
                    raw_resp.response_metadata["primary_provider"] = self.config.primary_provider
                    raw_resp.response_metadata["final_provider"] = self.config.failover_provider
                    raw_resp.response_metadata["failover_occurred"] = True
                    raw_resp.response_metadata["failover_reason"] = redact_secrets(str(primary_err))
                return self._normalize_response(raw_resp)
            except Exception as e2:
                failover_err = e2
                err_cat = classify_error_category(e2)
                safe_err = redact_secrets(str(e2))
                print(f"[AI Runtime] [{req_id}] Failover async ({self.config.failover_provider}) failed [{err_cat}]: {safe_err}")

        # 3. Mock fallback for development/test
        if self.config.allow_mock_fallback and self.config.environment != "production":
            print(f"[AI Runtime] [{req_id}] [MOCK] Both live providers failed. Serving async development mock response.")
            return self._generate_dev_mock_response(input)

        final_msg = (
            f"All configured AI providers failed. Mock fallback is disabled. "
            f"Primary ({self.config.primary_provider}): {redact_secrets(str(primary_err))}; "
            f"Fallback ({self.config.failover_provider}): {redact_secrets(str(failover_err))}"
        )
        raise RuntimeError(final_msg)

    def bind_tools(self, tools: List[Any], **kwargs) -> "UnifiedFailoverChatModel":
        bound_primary = (
            self.primary.bind_tools(tools, **kwargs)
            if hasattr(self.primary, "bind_tools")
            else self.primary
        )
        bound_fallback = (
            self.fallback.bind_tools(tools, **kwargs)
            if self.fallback is not None and hasattr(self.fallback, "bind_tools")
            else self.fallback
        )
        return UnifiedFailoverChatModel(bound_primary, bound_fallback, config=self.config)


# Alias for backward compatibility across test suites
FailoverChatModel = UnifiedFailoverChatModel


def get_chat_model(tools: Optional[List[Any]] = None) -> Any:
    """
    Returns the authoritative unified chat model configured according to AIRuntimeConfig.
    Initializes primary (Gemini) and fallback (OpenAI) with bind_tools support and explicit failover.
    When deterministic test mode is active (or environment == 'test'), routes to DeterministicTestHarness.
    """
    config = get_runtime_config()

    # Deterministic Test Harness Hook (Test Mode Only)
    try:
        from app.eval.harness import DeterministicTestHarness
        harness = DeterministicTestHarness.get_instance()
        if harness.is_enabled() or config.environment == "test":
            if config.environment == "production":
                raise RuntimeError(
                    "CRITICAL SECURITY VIOLATION: Deterministic test harness is prohibited in production environment."
                )
            return harness.get_chat_model(tools=tools)
    except ImportError:
        pass

    google_key = os.getenv("GOOGLE_API_KEY") or os.getenv("GEMINI_API_KEY")
    openai_key = os.getenv("OPENAI_API_KEY")

    gemini_model = None
    openai_model = None

    if google_key and ChatGoogleGenerativeAI is not None:
        print(f"[LLM Factory] Initializing ChatGoogleGenerativeAI ({config.primary_model})")
        gemini_model = ChatGoogleGenerativeAI(
            model=config.primary_model,
            google_api_key=google_key,
            temperature=config.temperature,
            max_retries=0,  # Fail fast to trigger observable failover
        )
        if tools:
            gemini_model = gemini_model.bind_tools(tools)

    if openai_key and ChatOpenAI is not None:
        print(f"[LLM Factory] Initializing ChatOpenAI ({config.failover_model})")
        openai_model = ChatOpenAI(
            model=config.failover_model,
            api_key=openai_key,
            temperature=config.temperature,
            max_retries=config.max_retries,
        )
        if tools:
            openai_model = openai_model.bind_tools(tools)

    if gemini_model and openai_model:
        print("[LLM Factory] Setting up Gemini primary with OpenAI fallback failover.")
        return UnifiedFailoverChatModel(gemini_model, openai_model, config=config)
    elif gemini_model:
        return UnifiedFailoverChatModel(gemini_model, None, config=config)
    elif openai_model:
        return UnifiedFailoverChatModel(openai_model, None, config=config)
    else:
        if config.allow_mock_fallback and config.environment != "production":
            print("[LLM Factory] Warning: No API keys configured. Using Mock AI model for development.")
            return UnifiedFailoverChatModel(None, None, config=config)
        raise RuntimeError("No LLM API Key found in environment. Set GOOGLE_API_KEY or OPENAI_API_KEY.")

