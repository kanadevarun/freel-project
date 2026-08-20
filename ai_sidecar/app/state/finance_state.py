from typing import List, Dict, Any, Optional
from typing_extensions import TypedDict


class FinanceState(TypedDict, total=False):
    # org / shipment / invoice context (all mandatory for graph execution)
    org_id: int
    shipment_id: int
    invoice_id: str
    invoice_number: str
    vendor_name: str
    s3_key: str
    file_name: str
    callback_url: str

    # OCR / parsing step
    raw_ocr_text: str

    # Extracted invoice data
    extracted_items: List[Dict[str, Any]]   # list of { charge_code, description, quantity, unit_price, total_amount }
    extracted_total: float
    extracted_currency: str
    extracted_vendor: str

    # Reference data fetched from Go backend
    contracted_rates: List[Dict[str, Any]]  # [{charge_code, contracted_price, ...}]

    # Reconciliation output
    discrepancies: List[Dict[str, Any]]     # [{charge_code, field_name, expected, actual, source}]
    invoice_status: str                     # APPROVED | DISCREPANCY

    # AI narrative summary
    ai_summary: str

    # Error propagation
    error: Optional[str]
