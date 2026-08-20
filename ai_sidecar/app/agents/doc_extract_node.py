import re
from app.state.compliance_state import ComplianceState

def doc_extract_node(state: ComplianceState) -> ComplianceState:
    """
    LLM/Regex Document Extraction Node.
    Maps fields such as container lists, weights, counts, and values.
    """
    raw_text = state.get("raw_ocr_text", "")
    print(f"[Compliance Agent] Extracting fields from OCR text...")

    # Extraction patterns
    gross_weight = 0
    package_count = 0
    container_number = ""
    seal_number = ""

    weight_match = re.search(r"Gross Weight:\s*(\d+)", raw_text)
    if weight_match:
        gross_weight = int(weight_match.group(1))

    count_match = re.search(r"Package Count:\s*(\d+)", raw_text)
    if count_match:
        package_count = int(count_match.group(1))

    container_match = re.search(r"Container Number:\s*([A-Z0-9]+)", raw_text)
    if container_match:
        container_number = container_match.group(1)

    seal_match = re.search(r"Seal Number:\s*([A-Za-z0-9\-]+)", raw_text)
    if seal_match:
        seal_number = seal_match.group(1)

    extracted = {
        "gross_weight": gross_weight,
        "package_count": package_count,
        "container_number": container_number,
        "seal_number": seal_number
    }

    return {
        **state,
        "extracted_data": extracted
    }
