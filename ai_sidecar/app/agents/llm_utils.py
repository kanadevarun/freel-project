"""
llm_utils.py — Unified LLM Execution, Observability, and JSON Parsing Utilities

Integrates:
- Authoritative get_chat_model() from llm_factory
- Prompt versioning and registry from app.prompts
- Environment-aware mock mode (active only when allow_mock_fallback=True)
- Telemetry tracing and secret redaction
"""

import re
import json
import time
import uuid
from typing import Optional, Dict, Any

from app.tools.llm_factory import get_chat_model, redact_secrets, classify_error_category
from app.config.runtime_config import get_runtime_config
from app.prompts.prompt_registry import get_prompt, get_prompt_definition


def parse_json_garbage(text: str) -> Optional[dict]:
    """Cleans markdown fencing and extracts valid JSON objects."""
    if not text:
        return None
    cleaned = re.sub(r"^```json\s*", "", text.strip(), flags=re.IGNORECASE)
    cleaned = re.sub(r"\s*```$", "", cleaned.strip())
    cleaned = cleaned.strip()

    try:
        return json.loads(cleaned)
    except Exception:
        # Match outermost curly braces
        match = re.search(r"\{.*\}", cleaned, re.DOTALL)
        if match:
            try:
                return json.loads(match.group(0))
            except Exception:
                pass
        print(f"[AI Sidecar] Failed to parse JSON from response text: {redact_secrets(cleaned[:200])}...")
        return None


def execute_llm_json(
    prompt: str,
    prompt_key: Optional[str] = None,
    prompt_version: Optional[str] = None,
    exec_ctx: Optional[Dict[str, Any]] = None,
) -> Optional[dict]:
    """
    Executes LLM completion and parses the output as JSON.
    Observes failover and logs execution telemetry safely.
    """
    cfg = get_runtime_config()
    start_time = time.time()
    req_id = (exec_ctx or {}).get("request_id") or uuid.uuid4().hex[:8]

    # Deterministic Test Harness Hook (Priority when active)
    try:
        from app.eval.harness import DeterministicTestHarness
        harness = DeterministicTestHarness.get_instance()
        if harness.is_enabled():
            chat = harness.get_chat_model()
            resp = chat.invoke(prompt)
            text = str(resp.content) if resp and hasattr(resp, "content") else ""
            return parse_json_garbage(text)
    except ImportError:
        pass

    # Deterministic Mock Fallback for Development/Testing ONLY
    if cfg.allow_mock_fallback and cfg.environment != "production":
        if "Can you send us your general brochure and catalog" in prompt:
            return {
                "intent": "QUESTION",
                "sentiment": "NEUTRAL",
                "confidence": 90,
                "lead_name": "New Customer",
                "company_domain": "freel-testing.local",
                "ai_summary": "A new potential client is requesting general company information."
            }
        if "industrial valves from Nhava Sheva" in prompt:
            return {
                "intent": "RFQ_REQUEST",
                "sentiment": "NEUTRAL",
                "confidence": 100,
                "lead_name": "Alex Mercer",
                "company_domain": "freel-testing.local",
                "origin_port": "Nhava Sheva (INNSA)",
                "destination_port": "Hamburg (DEHAM)",
                "incoterms": "FOB",
                "cargo_description": "industrial valves",
                "cargo_weight": 20000.0,
                "cargo_volume": 32.0,
                "target_date": "2026-10-15",
                "ai_summary": "Inbound ocean freight quote for 20 tons of industrial valves from INNSA to DEHAM."
            }
        if "shipment of machinery parts ready 2026-11-20" in prompt:
            return {
                "intent": "RFQ_REQUEST",
                "sentiment": "NEUTRAL",
                "confidence": 95,
                "lead_name": "Commercial Shipper",
                "company_domain": "freel-testing.local",
                "origin_port": "Nhava Sheva (INNSA)",
                "destination_port": "Hamburg (DEHAM)",
                "incoterms": "FOB",
                "cargo_description": "machinery parts",
                "cargo_weight": None,
                "cargo_volume": None,
                "target_date": "2026-11-20",
                "ai_summary": "Incomplete Quote Request. Missing mandatory fields: Cargo Weight, Cargo Volume."
            }
        if "Thanks for the response. The weight is 12,000 kg" in prompt:
            return {
                "intent": "FOLLOW_UP",
                "sentiment": "NEUTRAL",
                "confidence": 98,
                "lead_name": "Commercial Shipper",
                "company_domain": "freel-testing.local",
                "origin_port": None,
                "destination_port": None,
                "incoterms": None,
                "cargo_description": None,
                "cargo_weight": 12000.0,
                "cargo_volume": 18.0,
                "target_date": None,
                "ai_summary": "Customer confirmed weight of 12,000 kg and volume of 18 CBM."
            }
        if "packing list updated" in prompt or "15,000 kg" in prompt:
            return {
                "intent": "RFQ_REQUEST",
                "sentiment": "NEUTRAL",
                "confidence": 98,
                "lead_name": "Commercial Shipper",
                "company_domain": "freel-testing.local",
                "origin_port": "Nhava Sheva (INNSA)",
                "destination_port": "Hamburg (DEHAM)",
                "incoterms": "FOB",
                "cargo_description": "machinery parts",
                "cargo_weight": 15000.0,
                "cargo_volume": 18.0,
                "target_date": "2026-11-20",
                "ai_summary": "The shipper corrected the cargo weight to 15,000 kg."
            }
        if "Weekly Logistics Newsletter" in prompt:
            return {
                "intent": "QUESTION",
                "sentiment": "NEUTRAL",
                "confidence": 90,
                "lead_name": None,
                "company_domain": "spam-domain.com",
                "ai_summary": "Weekly Logistics Newsletter."
            }
        if "Duplicate quote check" in prompt:
            return {
                "intent": "RFQ_REQUEST",
                "sentiment": "NEUTRAL",
                "confidence": 100,
                "lead_name": "Duplicate Shipper",
                "company_domain": "freel-testing.local",
                "origin_port": "Nhava Sheva (INNSA)",
                "destination_port": "Hamburg (DEHAM)",
                "incoterms": "FOB",
                "cargo_description": "general cargo",
                "cargo_weight": 20000.0,
                "cargo_volume": 32.0,
                "target_date": "2026-10-15",
                "ai_summary": "Duplicate quote check email."
            }
        if "shipper_complete@tata-exports.local" in prompt or "automobile parts from Nhava Sheva to Hamburg" in prompt:
            return {
                "intent": "RFQ_REQUEST",
                "sentiment": "NEUTRAL",
                "confidence": 100,
                "lead_name": "Commercial Shipper",
                "company_domain": "tata-exports.local",
                "origin_port": "INNSA",
                "destination_port": "DEHAM",
                "incoterms": "FOB",
                "cargo_description": "automobile parts",
                "cargo_weight": 20000.0,
                "cargo_volume": 32.0,
                "target_date": "2026-09-01",
                "ai_summary": "Tata Exports complete RFQ."
            }
        if ("shipper_incomplete@tata-exports.local" in prompt or "Quick rate query: Mumbai to Hamburg" in prompt) and "Re:" not in prompt and "prior_rfq_context" not in prompt:
            return {
                "intent": "RFQ_REQUEST",
                "sentiment": "NEUTRAL",
                "confidence": 95,
                "lead_name": "Commercial Shipper",
                "company_domain": "tata-exports.local",
                "origin_port": "MUMBAI",
                "destination_port": "DEHAM",
                "incoterms": None,
                "cargo_description": "steel parts",
                "cargo_weight": None,
                "cargo_volume": None,
                "target_date": "2026-09-10",
                "ai_summary": "Incomplete Quote Request from MUMBAI to DEHAM."
            }
        if "shipper_conversation@tata-exports.local" in prompt or "shipper_incomplete@tata-exports.local" in prompt or "steel parts from Mumbai to Hamburg" in prompt:
            if "Re:" in prompt or "Sorry, forgot the details" in prompt or "prior_rfq_context" in prompt:
                return {
                    "intent": "RFQ_REQUEST",
                    "sentiment": "NEUTRAL",
                    "confidence": 98,
                    "lead_name": "Commercial Shipper",
                    "company_domain": "tata-exports.local",
                    "origin_port": None,
                    "destination_port": None,
                    "incoterms": "FOB",
                    "cargo_description": "steel parts",
                    "cargo_weight": 15000.0,
                    "cargo_volume": 24.0,
                    "target_date": None,
                    "ai_summary": "Customer provided missing fields in reply."
                }
            else:
                return {
                    "intent": "RFQ_REQUEST",
                    "sentiment": "NEUTRAL",
                    "confidence": 95,
                    "lead_name": "Commercial Shipper",
                    "company_domain": "tata-exports.local",
                    "origin_port": "MUMBAI",
                    "destination_port": "DEHAM",
                    "incoterms": None,
                    "cargo_description": "steel parts",
                    "cargo_weight": None,
                    "cargo_volume": None,
                    "target_date": "2026-09-10",
                    "ai_summary": "Conversational Turn 1."
                }

    # Execute Real AI Generation via Authoritative Chat Model
    try:
        chat = get_chat_model()
        resp = chat.invoke(prompt)
        text = str(resp.content) if resp and hasattr(resp, "content") else ""
        duration_ms = int((time.time() - start_time) * 1000)
        print(f"[AI Runtime] [{req_id}] LLM JSON execution completed in {duration_ms}ms")
        return parse_json_garbage(text)
    except Exception as e:
        safe_err = redact_secrets(str(e))
        err_cat = classify_error_category(e)
        print(f"[AI Runtime] [{req_id}] LLM execution failed [{err_cat}]: {safe_err}")
        if cfg.environment == "production":
            raise

    return None


def execute_llm_text(
    prompt: str,
    prompt_key: Optional[str] = None,
    prompt_version: Optional[str] = None,
    exec_ctx: Optional[Dict[str, Any]] = None,
) -> str:
    """Executes LLM generation and returns the plain text output."""
    cfg = get_runtime_config()
    start_time = time.time()
    req_id = (exec_ctx or {}).get("request_id") or uuid.uuid4().hex[:8]

    # Deterministic Test Harness Hook (Priority when active)
    try:
        from app.eval.harness import DeterministicTestHarness
        harness = DeterministicTestHarness.get_instance()
        if harness.is_enabled():
            chat = harness.get_chat_model()
            resp = chat.invoke(prompt)
            return str(resp.content).strip() if resp and hasattr(resp, "content") else ""
    except ImportError:
        pass

    # Development Mock Fallback ONLY
    if cfg.allow_mock_fallback and cfg.environment != "production":
        if "Draft a polite, professional, and concise email reply" in prompt or "Missing Mandatory Fields" in prompt:
            return """Dear Valued Shipper,

Thank you for reaching out for a shipping quote from Mumbai to Hamburg. We would be happy to assist you with transporting your cargo.

To provide you with an accurate and competitive rate, could you please provide the following missing details?

* Incoterms (e.g., FOB, CIF, EXW)
* Cargo Weight
* Cargo Volume

Once we have this information, we will promptly prepare and send over your quotation.

Best regards,
LogisticsHQ Sales Team"""

    try:
        chat = get_chat_model()
        resp = chat.invoke(prompt)
        text = str(resp.content) if resp and hasattr(resp, "content") else ""
        duration_ms = int((time.time() - start_time) * 1000)
        print(f"[AI Runtime] [{req_id}] LLM text execution completed in {duration_ms}ms")
        return text.strip()
    except Exception as e:
        safe_err = redact_secrets(str(e))
        err_cat = classify_error_category(e)
        print(f"[AI Runtime] [{req_id}] LLM text execution failed [{err_cat}]: {safe_err}")
        if cfg.environment == "production":
            raise

    return "LLM service unavailable."
