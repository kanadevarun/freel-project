import re
from typing import List, Dict, Any
from app.state.finance_state import FinanceState


def invoice_extract_node(state: FinanceState) -> FinanceState:
    """
    Invoice Extraction Node.
    Parses the raw OCR text and extracts structured line items, totals, currency, and vendor.

    Phase 4: deterministic regex-based extraction.
    Real implementation should use an LLM for varied document layouts.

    Contract:
        Input:  raw_ocr_text
        Output: extracted_items, extracted_total, extracted_currency, extracted_vendor
    """
    raw_text = state.get("raw_ocr_text", "")
    print(f"[Finance Agent] Extract: Parsing invoice line items from OCR text...")

    items: List[Dict[str, Any]] = []
    extracted_total = 0.0
    extracted_currency = "USD"
    extracted_vendor = state.get("vendor_name", "")

    # Extract line items: "CHARGE_CODE | Description | QTY: x | UNIT: y | TOTAL: z USD"
    line_pattern = re.compile(
        r"^([A-Z_]+)\s*\|\s*(.+?)\s*\|\s*QTY:\s*([\d.]+)\s*\|\s*UNIT:\s*([\d.]+)\s*\|\s*TOTAL:\s*([\d.]+)\s*([A-Z]{3})",
        re.MULTILINE,
    )
    currency_pattern = re.compile(r"CURRENCY:\s*([A-Z]{3})")
    total_pattern = re.compile(r"TOTAL AMOUNT:\s*([\d.]+)\s*([A-Z]{3})")
    vendor_pattern = re.compile(r"Vendor:\s*(.+)")

    # Extract currency
    currency_match = currency_pattern.search(raw_text)
    if currency_match:
        extracted_currency = currency_match.group(1)

    # Extract grand total
    total_match = total_pattern.search(raw_text)
    if total_match:
        extracted_total = float(total_match.group(1))
        extracted_currency = total_match.group(2)

    # Extract vendor
    vendor_match = vendor_pattern.search(raw_text)
    if vendor_match:
        extracted_vendor = vendor_match.group(1).strip()

    # Extract line items
    for match in line_pattern.finditer(raw_text):
        charge_code = match.group(1).strip()
        description = match.group(2).strip()
        quantity = float(match.group(3))
        unit_price = float(match.group(4))
        total_amount = float(match.group(5))
        currency = match.group(6)
        items.append({
            "charge_code": charge_code,
            "description": description,
            "quantity": quantity,
            "unit_price": unit_price,
            "total_amount": total_amount,
            "currency": currency,
        })

    print(f"[Finance Agent] Extract: Found {len(items)} line items. Total={extracted_total} {extracted_currency}")

    return {
        **state,
        "extracted_items": items,
        "extracted_total": extracted_total,
        "extracted_currency": extracted_currency,
        "extracted_vendor": extracted_vendor,
    }
