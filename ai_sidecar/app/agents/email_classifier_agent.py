import re
from typing import Optional, Dict, Any
from pydantic import BaseModel, Field
from app.prompts.prompt_registry import get_prompt
from app.agents.llm_utils import execute_llm_json

class EmailClassificationRequest(BaseModel):
    org_id: int
    user_id: Optional[int] = None
    from_email: Optional[str] = None
    sender: Optional[str] = None
    subject: str = ""
    body: str = ""
    correlation_id: str

    def get_effective_sender(self) -> str:
        return self.from_email or self.sender or "unknown@domain.local"

class EmailClassificationResponse(BaseModel):
    is_logistics_related: bool
    intent: str  # RFQ_INQUIRY, GENERAL_LOGISTICS, NON_LOGISTICS, BANK_ALERT, SPAM, MEETING, FOLLOW_UP
    sentiment: str = "NEUTRAL"  # POSITIVE, NEUTRAL, URGENT, NEGATIVE
    reasoning: str
    confidence: float = Field(default=0.90, ge=0.0, le=1.0)
    extracted_entities: Dict[str, Any] = {}
    correlation_id: str

class EmailClassifierAgent:
    """
    Authoritative Python Agent for Inbound Freight Shipper Email Classification and Triage.
    Migrated from Go to Python to preserve agentic reasoning in Python AI runtime.
    """

    LOGISTICS_KEYWORDS = [
        "freight", "rate", "quote", "shipping", "container", "fcl", "lcl",
        "teu", "bl", "bill of lading", "carrier", "customs", "origin",
        "destination", "transit", "demurrage", "detention", "port", "vessel",
        "rfq", "cargo", "ocean", "air freight", "inquiry", "shipment", "transport"
    ]

    NON_LOGISTICS_KEYWORDS = [
        "otp", "password reset", "statement of account", "bank transfer receipt",
        "wire transfer confirmation", "job application", "resume", "newsletter",
        "subscription receipt", "unsubscribe", "security alert"
    ]

    def classify_email(self, req: EmailClassificationRequest) -> EmailClassificationResponse:
        effective_sender = req.get_effective_sender()
        vars_dict = {
            "Sender": effective_sender,
            "Subject": req.subject,
            "Body": req.body,
            "PriorContext": "",
        }

        prompt_str = get_prompt("sales.email_classification", variables=vars_dict)
        if not prompt_str:
            prompt_str = f"""Analyze this inbound email:
From: {req.from_email}
Subject: {req.subject}
Body: {req.body[:1500]}

Determine if it is freight/logistics related and extract intent.
Return JSON: {{"is_logistics_related": bool, "intent": str, "sentiment": str, "reasoning": str}}"""

        exec_ctx = {
            "org_id": req.org_id,
            "user_id": req.user_id,
            "correlation_id": req.correlation_id,
            "workflow_name": "sales",
            "request_id": req.correlation_id,
        }

        output = execute_llm_json(
            prompt=prompt_str,
            prompt_key="sales.email_classification",
            prompt_version="1.0.0",
            exec_ctx=exec_ctx,
        )

        is_logistics = False
        intent = "NON_LOGISTICS"
        sentiment = "NEUTRAL"
        reasoning = "Automated classification of inbound communication."
        confidence = 0.85

        combined = f"{req.subject} {req.body}".lower()
        has_non_log = any(k in combined for k in self.NON_LOGISTICS_KEYWORDS)
        has_log = any(k in combined for k in self.LOGISTICS_KEYWORDS)
        is_reply = "re:" in req.subject.lower()

        if output and isinstance(output, dict):
            if "is_logistics_related" in output:
                is_logistics = bool(output["is_logistics_related"])
            elif "intent" in output:
                is_logistics = output["intent"] in ("RFQ_INQUIRY", "RFQ_REQUEST", "GENERAL_LOGISTICS", "FOLLOW_UP", "REPLY")

            if "intent" in output:
                intent = str(output["intent"]).upper()
            if "sentiment" in output:
                sentiment = str(output["sentiment"]).upper()
            if "reasoning" in output:
                reasoning = str(output["reasoning"])
            if "confidence" in output:
                try:
                    raw_conf = float(output["confidence"])
                    if raw_conf > 1.0:
                        raw_conf = raw_conf / 100.0
                    confidence = max(0.0, min(1.0, raw_conf))
                except Exception:
                    pass

            # If subject is a reply to an RFQ or has logistics keywords, ensure is_logistics is true
            if (is_reply and has_log) or ("rfq" in combined):
                is_logistics = True
                if intent == "NON_LOGISTICS":
                    intent = "RFQ_INQUIRY"
        else:
            # Deterministic heuristic classifier fallback
            if has_non_log and not has_log and not is_reply:
                is_logistics = False
                intent = "NON_LOGISTICS"
                reasoning = "Email contains administrative, banking, or non-freight patterns."
            elif has_log or (is_reply and not has_non_log):
                is_logistics = True
                if "quote" in combined or "rate" in combined or "rfq" in combined:
                    intent = "RFQ_INQUIRY"
                elif is_reply:
                    intent = "FOLLOW_UP"
                else:
                    intent = "GENERAL_LOGISTICS"
                reasoning = "Email matches commercial shipping, rate inquiry, or customer conversation reply."
            else:
                is_logistics = False
                intent = "GENERAL_LOGISTICS"
                reasoning = "General business communication without explicit logistics terms."

        return EmailClassificationResponse(
            is_logistics_related=is_logistics,
            intent=intent,
            sentiment=sentiment,
            reasoning=reasoning,
            confidence=confidence,
            correlation_id=req.correlation_id,
        )

    classify = classify_email
