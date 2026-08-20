import re
import time
from typing import Dict, Any
from app.state.contract_state import (
    ContractExtractionState,
    StateLogEntry,
    StateReviewItemDraft
)
from app.tools.go_api_client import normalize_port_tool

def validator_node(state: ContractExtractionState) -> Dict[str, Any]:
    """
    Validates port names and detects anomalies in the extracted rates.
    
    Simple meaning:
      This node iterates through every single rate that was parsed from the PDF.
      For each rate:
        1. It calls the 'normalize_port_tool' (which calls the Go backend API)
           to make sure the origin/destination ports are valid standard UN/LOCODEs.
           Example: If the PDF text said "Jnpt", the Go API translates it to "INNSA".
        2. It checks for a pricing anomaly: if the ocean freight exceeds $8,000,
           it flags it with a "PRICE_ANOMALY" tag.
        3. If any issues are found (unknown ports or high prices), it marks 'is_anomaly_detected = True'.
           This will trigger the graph execution pause (interrupt) before final ingestion.
    """
    logs = list(state.processing_log)
    timestamp = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
    
    logs.append(StateLogEntry(
        step="VALIDATION",
        timestamp=timestamp,
        message="Validating rates and performing anomaly detection checks..."
    ))

    confirmed_rates = []
    flagged_items = []
    is_anomaly = False

    for draft in state.extracted_rates:
        # A. Call the Go tool to normalize the origin port name.
        res_origin = normalize_port_tool.invoke(draft.origin_port)
        match_orig = re.search(r"UN/LOCODE '([A-Z]{5})'", res_origin)
        if match_orig:
            draft.origin_port = match_orig.group(1)
        
        # B. Call the Go tool to normalize the destination port name.
        res_dest = normalize_port_tool.invoke(draft.destination_port)
        match_dest = re.search(r"UN/LOCODE '([A-Z]{5})'", res_dest)
        if match_dest:
            draft.destination_port = match_dest.group(1)

        # C. Anomaly Evaluation logic
        review_flags = []
        reasoning_parts = []

        # If the port wasn't found in the Go backend ports database, flag it.
        if not match_orig:
            review_flags.append("PORT_UNKNOWN")
            reasoning_parts.append(f"Origin port '{draft.origin_port}' is unrecognized.")
        if not match_dest:
            review_flags.append("PORT_UNKNOWN")
            reasoning_parts.append(f"Destination port '{draft.destination_port}' is unrecognized.")

        # Price threshold check: Flag rates exceeding $8,000 USD.
        if draft.ocean_freight > 8000.0:
            review_flags.append("PRICE_ANOMALY")
            reasoning_parts.append(f"Ocean Freight is USD {draft.ocean_freight} which is exceptionally high.")

        # D. Route to appropriate queue list
        if len(review_flags) > 0:
            is_anomaly = True
            # Flagged items require human verification.
            flagged = StateReviewItemDraft(
                extracted_data=draft,
                confidence_score=50 if "PORT_UNKNOWN" in review_flags else 62,
                review_flags=review_flags,
                ai_reasoning="; ".join(reasoning_parts),
                source_page=3,
                source_text=f"High ocean freight or unrecognized ports for {draft.origin_port} -> {draft.destination_port}",
                source_image_url="http://localhost:8080/uploads/pages/page3.png"
            )
            flagged_items.append(flagged)
        else:
            # Confirmed rates bypass human review and are ingested directly.
            confirmed_rates.append(draft)

    logs.append(StateLogEntry(
        step="VALIDATION",
        timestamp=time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        message=f"Validation finished: confirmed {len(confirmed_rates)} rates, flagged {len(flagged_items)} rate anomaly."
    ))

    return {
        "extracted_rates": confirmed_rates,
        "flagged_items": flagged_items,
        "is_anomaly_detected": is_anomaly,
        "processing_log": logs
    }
