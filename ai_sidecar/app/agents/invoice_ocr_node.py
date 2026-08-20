from app.state.finance_state import FinanceState


def invoice_ocr_node(state: FinanceState) -> FinanceState:
    """
    Invoice OCR Node.
    Parses the raw invoice document layout text.
    Phase 4 placeholder: returns a structured mock invoice layout.
    Real implementation should use a document-intelligence OCR service.

    Contract:
        Input:  invoice_id, file_name, s3_key (document reference)
        Output: raw_ocr_text (raw invoice layout text for the extraction node)
    """
    file_name = state.get("file_name", "invoice.pdf")
    invoice_number = state.get("invoice_number", "INV-UNKNOWN")
    vendor_name = state.get("vendor_name", "Unknown Vendor")

    print(f"[Finance Agent] OCR: Processing invoice {file_name} ({invoice_number}) from {vendor_name}...")

    # Mock OCR output representing a typical carrier/broker invoice layout.
    # TODO: Replace with real OCR/document-intelligence service call in production.
    mock_ocr = f"""
CARRIER FREIGHT INVOICE
Invoice Number: {invoice_number}
Vendor: {vendor_name}
Date: 2024-03-15

LINE ITEMS:
OCEAN_FREIGHT | Ocean Base Freight | QTY: 1 | UNIT: 2500.00 | TOTAL: 2500.00 USD
ORIGIN_THC    | Origin Terminal Handling | QTY: 1 | UNIT: 150.00 | TOTAL: 150.00 USD
DEST_THC      | Destination Terminal Handling | QTY: 1 | UNIT: 175.00 | TOTAL: 175.00 USD
DOC_FEE       | Documentation Fee | QTY: 1 | UNIT: 75.00 | TOTAL: 75.00 USD
FUEL_SURCHARGE| Fuel Surcharge (BAF) | QTY: 1 | UNIT: 320.00 | TOTAL: 320.00 USD

TOTAL AMOUNT: 3220.00 USD
CURRENCY: USD
PAYMENT TERMS: NET 30
"""

    # Inject a mismatch for test invoices to verify discrepancy detection
    if "mismatch" in file_name.lower():
        # Simulate overcharging on ocean freight to trigger a discrepancy
        mock_ocr = mock_ocr.replace(
            "OCEAN_FREIGHT | Ocean Base Freight | QTY: 1 | UNIT: 2500.00 | TOTAL: 2500.00 USD",
            "OCEAN_FREIGHT | Ocean Base Freight | QTY: 1 | UNIT: 3100.00 | TOTAL: 3100.00 USD"
        ).replace("TOTAL AMOUNT: 3220.00 USD", "TOTAL AMOUNT: 3820.00 USD")

    return {
        **state,
        "raw_ocr_text": mock_ocr.strip(),
    }
