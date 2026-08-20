from typing import List, Dict, Any
from typing_extensions import TypedDict

class ComplianceState(TypedDict, total=False):
    # org / shipment details
    org_id: int
    shipment_id: int
    doc_id: str
    doc_type: str
    s3_key: str
    file_name: str
    callback_url: str

    # OCR extraction steps
    raw_ocr_text: str
    extracted_data: Dict[str, Any]

    # Verification and cross-reconciliation output
    discrepancies: List[Dict[str, Any]]
    doc_status: str # VERIFIED or DISCREPANCY
