import json
from app.state.compliance_state import ComplianceState

def doc_ocr_node(state: ComplianceState) -> ComplianceState:
    """
    Mock OCR Extraction Node.
    Extracts text layouts from the target document and preserves original layout text.
    """
    file_name = state.get("file_name", "document.pdf")
    doc_type = state.get("doc_type", "HBL")
    print(f"[Compliance Agent] OCR: Processing file {file_name} of type {doc_type}...")

    # Simulated OCR extraction representing typical HBL/MBL layout files
    mock_ocr = f"""
    SHIPPING DOCUMENT - TYPE: {doc_type}
    Carrier: Maersk Line (MAEU)
    Booking Ref: BK-12345
    Container Number: MSKU1234567
    Seal Number: SL-999888
    Gross Weight: 24500 KGS
    Package Count: 42 packages
    Shipper: Global Manufacturing Inc.
    Consignee: Freel Logistics Group
    Invoice Value: 75000 USD
    """
    
    # Intentionally plant a mismatch if HBL to verify discrepancy detections in tests
    if doc_type == "HBL" and "mismatched" in file_name:
        mock_ocr = mock_ocr.replace("Gross Weight: 24500 KGS", "Gross Weight: 21200 KGS")
        mock_ocr = mock_ocr.replace("MSKU1234567", "MSKU9999999")

    return {
        **state,
        "raw_ocr_text": mock_ocr.strip()
    }
