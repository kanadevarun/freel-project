import re
from typing import Optional, List, Dict, Any
from pydantic import BaseModel, Field
from app.prompts.prompt_registry import get_prompt
from app.agents.llm_utils import execute_llm_json

class ExtractedShipmentData(BaseModel):
    origin: Optional[str] = None
    destination: Optional[str] = None
    incoterms: Optional[str] = None
    weight: Optional[str] = None
    volume: Optional[str] = None

class ShipmentParseRequest(BaseModel):
    org_id: int
    user_id: Optional[int] = None
    raw_text: Optional[str] = None
    text: Optional[str] = None
    correlation_id: str

    def get_effective_text(self) -> str:
        return self.raw_text or self.text or ""

class ShipmentParseResponse(BaseModel):
    data: ExtractedShipmentData
    confidence_score: int = Field(default=85, ge=0, le=100)
    missing_fields: List[str] = []
    correlation_id: str

    @property
    def extracted_data(self) -> ExtractedShipmentData:
        return self.data

    @property
    def confidence(self) -> float:
        return float(self.confidence_score) / 100.0

class RFQParserAgent:
    """
    Authoritative Python Agent for RFQ Free-Text Shipment Extraction and Parameter Resolution.
    Migrated from Go to Python to preserve agentic extraction reasoning in Python AI runtime.
    """

    def parse_request(self, req: ShipmentParseRequest) -> ShipmentParseResponse:
        effective_text = req.get_effective_text()
        vars_dict = {
            "RawText": effective_text,
        }

        prompt_str = get_prompt("rfq.extract_shipment_request", variables=vars_dict)
        if not prompt_str:
            prompt_str = f"""Extract freight shipping details from this text:
{effective_text}

Extract:
- origin: port or city of origin
- destination: port or city of delivery
- incoterms: FOB, CIF, EXW, DDP, etc.
- weight: cargo weight
- volume: container equipment or CBM

Return JSON:
{{
  "data": {{
    "origin": str or null,
    "destination": str or null,
    "incoterms": str or null,
    "weight": str or null,
    "volume": str or null
  }},
  "confidence_score": int,
  "missing_fields": []
}}"""

        exec_ctx = {
            "org_id": req.org_id,
            "user_id": req.user_id,
            "correlation_id": req.correlation_id,
            "workflow_name": "rfq",
            "request_id": req.correlation_id,
        }

        output = execute_llm_json(
            prompt=prompt_str,
            prompt_key="rfq.extract_shipment_request",
            prompt_version="1.0.0",
            exec_ctx=exec_ctx,
        )

        extracted = ExtractedShipmentData()
        confidence = 80
        missing: List[str] = []

        if output and isinstance(output, dict):
            data_dict = output.get("data", output)
            if isinstance(data_dict, dict):
                extracted.origin = data_dict.get("origin")
                extracted.destination = data_dict.get("destination")
                extracted.incoterms = data_dict.get("incoterms")
                extracted.weight = data_dict.get("weight")
                extracted.volume = data_dict.get("volume")

            if "confidence_score" in output:
                try:
                    confidence = int(output["confidence_score"])
                except Exception:
                    pass
            if "missing_fields" in output and isinstance(output["missing_fields"], list):
                missing = [str(f) for f in output["missing_fields"]]
        else:
            # Deterministic heuristic regex extraction fallback
            text = req.raw_text

            # Origin / Destination patterns
            from_match = re.search(r"(?i)\b(?:from|origin|pol)\s*[:\-]?\s*([A-Za-z\s,]+?)(?:\bto\b|\bdest|\n|$)", text)
            if from_match:
                extracted.origin = from_match.group(1).strip()

            to_match = re.search(r"(?i)\b(?:to|dest|destination|pod)\s*[:\-]?\s*([A-Za-z\s,]+?)(?:\bterms|\bincoterms|\n|$)", text)
            if to_match:
                extracted.destination = to_match.group(1).strip()

            # Incoterms pattern
            inco_match = re.search(r"\b(FOB|CIF|EXW|DDP|DAP|CFR|FCA|CPT|CIP)\b", text, re.IGNORECASE)
            if inco_match:
                extracted.incoterms = inco_match.group(1).upper()

            # Weight pattern
            weight_match = re.search(r"(\d+(?:\.\d+)?\s*(?:kg|tons?|lbs|mt))", text, re.IGNORECASE)
            if weight_match:
                extracted.weight = weight_match.group(1)

            # Volume / Container pattern
            vol_match = re.search(r"(\d+\s*(?:x\s*)?(?:20(?:ft|gp|dc)?|40(?:ft|hc|gp|dc)?|cbm|teu))", text, re.IGNORECASE)
            if vol_match:
                extracted.volume = vol_match.group(1)

        # Check for missing fields
        if not extracted.origin:
            missing.append("origin")
        if not extracted.destination:
            missing.append("destination")
        if not extracted.incoterms:
            missing.append("incoterms")
        if not extracted.volume and not extracted.weight:
            missing.append("volume_or_weight")

        if missing:
            confidence = max(40, confidence - len(missing) * 15)

        return ShipmentParseResponse(
            data=extracted,
            confidence_score=confidence,
            missing_fields=missing,
            correlation_id=req.correlation_id,
        )
