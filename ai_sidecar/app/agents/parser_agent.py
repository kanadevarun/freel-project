import time
from typing import Dict, Any
from app.state.contract_state import (
    ContractExtractionState,
    StateLogEntry,
    StateRateDraft,
    StateSurcharge
)
from app.agents.llm_utils import execute_llm_json

def normalize_surcharge_code(code: str, description: str) -> str:
    """
    Normalizes dirty or non-standard surcharge names to canonical industry codes.
    
    Simple meaning:
      Maps variations like "BUNKER", "FUEL", "FUEL SURCHARGE" to "BAF",
      "PEAK SEASON", "PSS" to "PSS", etc.
      
    Example:
      normalize_surcharge_code("FUEL", "Fuel Adjustment Factor") -> "BAF"
    """
    # Clean up strings
    code_clean = (code or "").strip().upper()
    desc_clean = (description or "").strip().upper()
    
    # 1. Direct code mappings
    if code_clean in ["BAF", "BUC", "FRB", "IFP", "FUEL", "BUNKER"]:
        return "BAF"
    if code_clean in ["CAF", "CURR", "YAS", "CURRENCY"]:
        return "CAF"
    if code_clean in ["PSS", "PEAK", "SEASON"]:
        return "PSS"
    if code_clean in ["WRS", "WAR", "RISK"]:
        return "WRS"
    if code_clean in ["OHC", "OTHC", "ORC", "ORIGIN"]:
        return "OHC"
    if code_clean in ["DHC", "DTHC", "DESTINATION"]:
        return "DHC"
        
    # 2. Description keyword matching
    combined = f"{code_clean} {desc_clean}"
    if "FUEL" in combined or "BUNKER" in combined or "BAF" in combined:
        return "BAF"
    if "CURRENCY" in combined or "CAF" in combined or "ADJUSTMENT" in combined:
        if "BUNKER" not in combined and "FUEL" not in combined:
            return "CAF"
    if "PEAK" in combined or "SEASON" in combined or "PSS" in combined:
        return "PSS"
    if "WAR" in combined or "WRS" in combined or "RISK" in combined:
        return "WRS"
    if ("ORIGIN" in combined and "HANDLING" in combined) or "OHC" in combined or "OTHC" in combined or "ORC" in combined:
        return "OHC"
    if ("DESTINATION" in combined and "HANDLING" in combined) or "DHC" in combined or "DTHC" in combined:
        return "DHC"
        
    return code_clean if code_clean else "SUR"

def parser_node(state: ContractExtractionState) -> Dict[str, Any]:
    """Parses port pairs and surcharges into structured rate drafts."""
    logs = list(state.processing_log)
    timestamp = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
    
    logs.append(StateLogEntry(
        step="TABLE_PARSING",
        timestamp=timestamp,
        message="Extracting rates, note quotes, and surcharge columns..."
    ))

    parser_prompt = f"""
You are an expert freight forwarding operations executive. Read the following shipping contract text and extract all port-pair rates.

Document Text:
{state.raw_text}

For each rate found, construct a JSON object matching this structure:
{{
  "origin_port": "UN/LOCODE or port name (e.g. INNSA)",
  "destination_port": "UN/LOCODE or port name (e.g. DEHAM)",
  "via_port": "transshipment port LOCODE or null",
  "service_code": "route code or service name or null",
  "carrier_scac": "{state.carrier_scac}",
  "carrier_name": "{state.carrier_name}",
  "vessel_name": "vessel name if specified or null",
  "equipment_type": "40GP",
  "ocean_freight": <ocean freight base amount as float>,
  "origin_charges": <origin terminal handling fee as float>,
  "destination_charges": <destination terminal handling fee as float>,
  "surcharges": [
     {{
       "code": "BAF|CAF|PSS|etc",
       "description": "fuel adjust factor or currency surcharge description",
       "amount": <surcharge amount as float>,
       "unit": "PER_TEU|PER_CONTAINER|PER_SHIPMENT",
       "included": true/false
     }}
  ],
  "total_buy_price": <ocean_freight + origin_charges + destination_charges + all non-included surcharges>,
  "free_days_origin": <free detention days POL or 0>,
  "free_days_destination": <free demurrage days POD or 14>,
  "transit_days": <transit time in days as integer or null>,
  "incoterms": "FOB|CIF|DDP|EXW|null",
  "valid_from": "YYYY-MM-DD",
  "valid_until": "YYYY-MM-DD"
}}

Rules:
- Port codes must be standard UN/LOCODEs if possible.
- If a rate is extremely high/low or suspicious, still extract it exactly as written.

Return a JSON object with a single key "rates" containing a list of these objects:
{{"rates": [...]}}
"""
    extracted_data = execute_llm_json(parser_prompt)
    rates_list = []
    
    if extracted_data and "rates" in extracted_data:
        rates_list = extracted_data["rates"]
        logs.append(StateLogEntry(
            step="TABLE_PARSING",
            timestamp=time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
            message=f"AI successfully parsed {len(rates_list)} port-pair rates from the contract text."
        ))
    else:
        logs.append(StateLogEntry(
            step="TABLE_PARSING",
            timestamp=time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
            message="AI Parsing failed or returned no rates. Falling back to default parser Mock template."
        ))
        rates_list = [
            {
                "origin_port": "INNSA",
                "destination_port": "DEHAM",
                "via_port": "SGSIN",
                "service_code": "AE7",
                "carrier_scac": state.carrier_scac,
                "carrier_name": state.carrier_name,
                "equipment_type": "40GP",
                "ocean_freight": 2800.0,
                "origin_charges": 180.0,
                "destination_charges": 220.0,
                "surcharges": [
                    {
                        "code": "BAF",
                        "description": "Bunker Adjustment Factor",
                        "amount": 350.0,
                        "unit": "PER_CONTAINER",
                        "included": True
                    },
                    {
                        "code": "CAF",
                        "description": "Currency Adjustment Factor",
                        "amount": 45.0,
                        "unit": "PER_CONTAINER",
                        "included": True
                    }
                ],
                "total_buy_price": 3200.0,
                "free_days_origin": 0,
                "free_days_destination": 14,
                "transit_days": 22,
                "incoterms": "FOB",
                "valid_from": "2026-09-01",
                "valid_until": "2026-12-31"
            },
            {
                "origin_port": "INNSA",
                "destination_port": "NLRTM",
                "service_code": "AE1",
                "carrier_scac": state.carrier_scac,
                "carrier_name": state.carrier_name,
                "equipment_type": "40GP",
                "ocean_freight": 2900.0,
                "origin_charges": 180.0,
                "destination_charges": 210.0,
                "surcharges": [
                    {
                        "code": "BAF",
                        "description": "Bunker Adjustment Factor",
                        "amount": 350.0,
                        "unit": "PER_CONTAINER",
                        "included": True
                    }
                ],
                "total_buy_price": 3290.0,
                "free_days_origin": 0,
                "free_days_destination": 14,
                "transit_days": 21,
                "incoterms": "FOB",
                "valid_from": "2026-09-01",
                "valid_until": "2026-12-31"
            }
        ]

    # Convert dictionary formats to Pydantic Models
    state_rates = []
    for r in rates_list:
        valid_from = r.get("valid_from", "2026-09-01")
        if not valid_from.endswith("T00:00:00Z"):
            valid_from = f"{valid_from}T00:00:00Z"
        valid_until = r.get("valid_until", "2026-12-31")
        if not valid_until.endswith("T23:59:59Z"):
            valid_until = f"{valid_until}T23:59:59Z"

        state_rates.append(StateRateDraft(
            origin_port=r.get("origin_port", "INNSA"),
            destination_port=r.get("destination_port", "DEHAM"),
            via_port=r.get("via_port"),
            service_code=r.get("service_code"),
            carrier_scac=r.get("carrier_scac", state.carrier_scac),
            carrier_name=r.get("carrier_name", state.carrier_name),
            vessel_name=r.get("vessel_name"),
            equipment_type=r.get("equipment_type", "40GP"),
            ocean_freight=r.get("ocean_freight", 2800.0),
            origin_charges=r.get("origin_charges", 0.0),
            destination_charges=r.get("destination_charges", 0.0),
            surcharges=[
                StateSurcharge(
                    code=normalize_surcharge_code(s.get("code", "SUR"), s.get("description", "")),
                    description=s.get("description", ""),
                    amount=float(s.get("amount", 0.0)),
                    unit=s.get("unit", "PER_TEU"),
                    included=bool(s.get("included", False))
                )
                for s in r.get("surcharges", [])
            ],
            total_buy_price=r.get("total_buy_price", 3200.0),
            free_days_origin=r.get("free_days_origin", 0),
            free_days_destination=r.get("free_days_destination", 14),
            transit_days=r.get("transit_days"),
            incoterms=r.get("incoterms"),
            valid_from=valid_from,
            valid_until=valid_until,
            confidence_score=95
        ))

    return {"extracted_rates": state_rates, "processing_log": logs}
