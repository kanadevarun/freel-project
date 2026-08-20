import os
import time
from typing import Dict, Any
from pypdf import PdfReader
from app.state.contract_state import ContractExtractionState, StateLogEntry

def ocr_node(state: ContractExtractionState) -> Dict[str, Any]:
    """Extracts raw text from PDF files or falls back to mock template."""
    logs = list(state.processing_log)
    timestamp = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
    
    logs.append(StateLogEntry(
        step="OCR_PROCESSING",
        timestamp=timestamp,
        message=f"Reading contract document text (key: {state.s3_key})..."
    ))

    local_path = f"/Users/varun.kanade/go/src/freel/freel-project/backend/uploads/{state.s3_key}"
    raw_text = ""
    pages_count = 0

    if os.path.exists(local_path):
        try:
            reader = PdfReader(local_path)
            pages_count = len(reader.pages)
            extracted_pages = []
            for i in range(min(5, pages_count)):
                extracted_pages.append(f"--- PAGE {i+1} ---\n" + reader.pages[i].extract_text())
            raw_text = "\n".join(extracted_pages)
            logs.append(StateLogEntry(
                step="OCR_PROCESSING",
                timestamp=time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
                message=f"Successfully extracted text from {pages_count} pages locally."
            ))
        except Exception as e:
            logs.append(StateLogEntry(
                step="OCR_PROCESSING",
                timestamp=time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
                message=f"Failed to parse PDF file: {e}. Falling back to default mock text."
            ))
    else:
        logs.append(StateLogEntry(
            step="OCR_PROCESSING",
            timestamp=time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
            message="File not found locally. Simulating OCR extraction on mock template."
        ))

    if not raw_text:
        raw_text = """
MAERSK LINE SERVICE CONTRACT
Contract Party: Freel Global Logistics
Contract Ref: SC-99201-2026
Carrier Code: MAEU

Rate Tables - Validity: 01-Sep-2026 to 31-Dec-2026
Route: Nhava Sheva (INNSA) to Hamburg (DEHAM)
Ocean Freight: USD 2,800 per 40GP standard dry
Origin THC: USD 180
Destination THC: USD 220
Subject to: BAF (Bunker Adjustment Factor) USD 350 (Included in ocean freight)
Subject to: CAF (Currency Adjustment Factor) USD 45 (Included in ocean freight)
Transit Time: 22 days via Singapore
Free Time at destination: 14 days standard

Route: Nhava Sheva (INNSA) to Rotterdam (NLRTM)
Ocean Freight: USD 2,900 per 40GP standard dry
Origin THC: USD 180
Destination THC: USD 210
Subject to: BAF USD 350 (Included in ocean freight)
Transit Time: 21 days
Free Time: 14 days

Route: Nhava Sheva (INNSA) to Jebel Ali (AEJEA)
Ocean Freight: USD 12,400 per 40GP standard dry
Origin THC: USD 150
Destination THC: USD 180
Transit Time: 8 days
Free Time: 14 days
        """
    
    return {"raw_text": raw_text, "processing_log": logs}
