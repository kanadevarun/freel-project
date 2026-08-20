import time
from typing import Dict, Any
from app.state.contract_state import ContractExtractionState, StateLogEntry
from app.agents.llm_utils import execute_llm_json

def classify_node(state: ContractExtractionState) -> Dict[str, Any]:
    """Identifies the carrier and SCAC code using LLMs."""
    logs = list(state.processing_log)
    timestamp = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
    
    logs.append(StateLogEntry(
        step="CLASSIFICATION",
        timestamp=timestamp,
        message="Analyzing carrier formats, metadata and validity headers..."
    ))

    classifier_prompt = f"""
Analyze the following shipping document text and classify:
1. carrier_scac: e.g. MAEU (Maersk), MSCU (MSC), CMDU (CMA CGM), ONEY (ONE), HLCU (Hapag-Lloyd).
2. carrier_name: display name.
3. contract_type: "OCEAN_CONTRACT" or "RATE_CIRCULAR" or "UNKNOWN".
4. validity_range: format {{"from": "YYYY-MM-DD", "until": "YYYY-MM-DD"}}.

Document Text:
{state.raw_text[:1500]}

Return ONLY a JSON object:
{{"carrier_scac": "SCAC", "carrier_name": "Name", "contract_type": "TYPE", "validity": {{"from": "YYYY-MM-DD", "until": "YYYY-MM-DD"}}}}
"""
    classification = execute_llm_json(classifier_prompt)
    if classification:
        carrier_scac = classification.get("carrier_scac", "MAEU")
        carrier_name = classification.get("carrier_name", "Maersk")
        contract_type = classification.get("contract_type", "OCEAN_CONTRACT")
        logs.append(StateLogEntry(
            step="CLASSIFICATION",
            timestamp=time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
            message=f"AI identified carrier as {carrier_name} ({carrier_scac}) and document as {contract_type}."
        ))
    else:
        carrier_scac = "MAEU"
        carrier_name = "Maersk"
        logs.append(StateLogEntry(
            step="CLASSIFICATION",
            timestamp=time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
            message="Classifier call failed. Defaulting to Maersk (MAEU) layout context."
        ))

    return {"carrier_scac": carrier_scac, "carrier_name": carrier_name, "processing_log": logs}
