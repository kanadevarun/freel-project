"""
runtime_config.py — Authoritative AI Runtime Configuration Contract for LogisticsHQ Python Sidecar

Enforces single source of truth for:
- Model & provider selection (Primary: Gemini, Failover: OpenAI)
- Explicit failover rules and timeout boundaries
- Environment-aware mock mode safety (strictly prohibited in production)
- Secret redaction and telemetry sanitization
"""

import os
from typing import Dict, Any, Optional

VERIFIED_MODELS = {
    "gemini-1.5-flash",
    "gemini-1.5-pro",
    "gemini-2.0-flash",
    "gemini-3.1-flash-lite",
    "gpt-4o",
    "gpt-4o-mini",
    "gpt-4-turbo",
    "mock-evaluator",
}


class AIRuntimeConfig:
    def __init__(
        self,
        environment: Optional[str] = None,
        primary_provider: Optional[str] = None,
        primary_model: Optional[str] = None,
        failover_provider: Optional[str] = None,
        failover_model: Optional[str] = None,
        allow_mock_fallback: Optional[bool] = None,
        max_retries: Optional[int] = None,
        provider_timeout_seconds: Optional[int] = None,
        tool_call_timeout_seconds: Optional[int] = None,
        max_output_tokens: Optional[int] = None,
        temperature: Optional[float] = None,
        streaming_enabled: Optional[bool] = None,
        default_prompt_version: Optional[str] = None,
        gemini_api_key: Optional[str] = None,
        openai_api_key: Optional[str] = None,
    ):
        self.environment = (
            environment or os.getenv("APP_ENV", "development").strip().lower()
        )
        self.primary_provider = (
            primary_provider
            or os.getenv("AI_PRIMARY_PROVIDER", "gemini").strip().lower()
        )
        self.primary_model = (
            primary_model
            or os.getenv("GEMINI_MODEL", "gemini-1.5-flash").strip()
        )
        self.failover_provider = (
            failover_provider
            or os.getenv("AI_FAILOVER_PROVIDER", "openai").strip().lower()
        )
        self.failover_model = (
            failover_model
            or os.getenv("OPENAI_MODEL", "gpt-4o-mini").strip()
        )

        if gemini_api_key is not None:
            self.gemini_api_key = gemini_api_key
        else:
            self.gemini_api_key = (
                os.getenv("GEMINI_API_KEY") or os.getenv("GOOGLE_API_KEY")
            )

        if openai_api_key is not None:
            self.openai_api_key = openai_api_key
        else:
            self.openai_api_key = os.getenv("OPENAI_API_KEY")

        # Mock fallback policy: disabled in production unless explicitly allowed for test
        if allow_mock_fallback is not None:
            self.allow_mock_fallback = allow_mock_fallback
        else:
            raw_mock = os.getenv("ALLOW_MOCK_FALLBACK")
            if raw_mock is not None:
                self.allow_mock_fallback = raw_mock.strip().lower() in ("true", "1", "yes")
            else:
                self.allow_mock_fallback = (self.environment != "production")

        self.max_retries = max_retries or int(os.getenv("AI_MAX_RETRIES", "2"))
        self.provider_timeout_seconds = provider_timeout_seconds or int(
            os.getenv("AI_PROVIDER_TIMEOUT_SECONDS", "30")
        )
        self.tool_call_timeout_seconds = tool_call_timeout_seconds or int(
            os.getenv("AI_TOOL_TIMEOUT_SECONDS", "30")
        )
        self.max_output_tokens = max_output_tokens or 2048
        self.temperature = (
            temperature
            if temperature is not None
            else float(os.getenv("AI_TEMPERATURE", "0.1"))
        )
        self.streaming_enabled = (
            streaming_enabled
            if streaming_enabled is not None
            else os.getenv("AI_STREAMING_ENABLED", "false").strip().lower() in ("true", "1")
        )
        self.default_prompt_version = (
            default_prompt_version
            or os.getenv("AI_DEFAULT_PROMPT_VERSION", "1.0.0").strip()
        )

        self.validate()

    def is_development(self) -> bool:
        return self.environment != "production"

    def is_production(self) -> bool:
        return self.environment == "production"

    def validate(self) -> None:
        """Validates configuration against environment and security boundaries."""
        if self.environment == "production":
            if self.allow_mock_fallback:
                raise ValueError(
                    "CRITICAL SECURITY VIOLATION: Production AI configuration error. "
                    "allow_mock_fallback cannot be enabled in production environment."
                )
            if self.primary_provider == "gemini":
                if not self.gemini_api_key:
                    raise ValueError(
                        "CRITICAL: missing required primary API key for provider 'gemini' in production."
                    )
            elif self.primary_provider == "openai":
                if not self.openai_api_key:
                    raise ValueError(
                        "CRITICAL: missing required primary API key for provider 'openai' in production."
                    )

        if self.primary_model and self.primary_model not in VERIFIED_MODELS:
            raise ValueError(
                f"Unverified or deprecated Gemini model: '{self.primary_model}'. "
                f"Must be one of: {sorted(list(VERIFIED_MODELS))}"
            )
        if self.failover_model and self.failover_model not in VERIFIED_MODELS:
            raise ValueError(
                f"Unverified or deprecated failover model: '{self.failover_model}'. "
                f"Must be one of: {sorted(list(VERIFIED_MODELS))}"
            )

    def to_safe_dict(self) -> Dict[str, Any]:
        """Returns clean telemetry-safe dictionary omitting API keys and secrets."""
        return {
            "environment": self.environment,
            "primary_provider": self.primary_provider,
            "primary_model": self.primary_model,
            "failover_provider": self.failover_provider,
            "failover_model": self.failover_model,
            "allow_mock_fallback": self.allow_mock_fallback,
            "max_retries": self.max_retries,
            "provider_timeout_seconds": self.provider_timeout_seconds,
            "temperature": self.temperature,
            "streaming_enabled": self.streaming_enabled,
            "default_prompt_version": self.default_prompt_version,
            "gemini_api_key": f"[CONFIGURED: {len(self.gemini_api_key)} chars]" if self.gemini_api_key else "[NOT_SET]",
            "openai_api_key": f"[CONFIGURED: {len(self.openai_api_key)} chars]" if self.openai_api_key else "[NOT_SET]",
        }


# Process-level singleton configuration
_runtime_config: Optional[AIRuntimeConfig] = None


def get_runtime_config() -> AIRuntimeConfig:
    global _runtime_config
    if _runtime_config is None:
        _runtime_config = AIRuntimeConfig()
    return _runtime_config


def set_runtime_config(config: AIRuntimeConfig) -> None:
    global _runtime_config
    _runtime_config = config
