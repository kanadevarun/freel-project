"""
context_tools.py — Read-only AI business context tools for LogisticsHQ agents.

Guarantees:
- Strictly read-only: No mutations to business records, invoices, shipments, quotations, or emails.
- Organization-isolated: Every tool requires and enforces `org_id`.
- Grounded: Every response includes supporting records, field citations, and data freshness.
- Safe error handling: Structured failure responses without leaking internals.
- Bridge routing: All context queries run via the Go backend's Centralized Action System
  and internal authenticated context endpoints.
"""

import os
import httpx
from typing import Optional, Dict, Any, List
from pydantic import BaseModel, Field

from app.tools.auth_utils import get_internal_service_token

try:
    from langchain_core.tools import tool
except ImportError:
    # Fallback decorator if langchain_core is mocked/not present in test
    def tool(name_or_fn=None, **kwargs):
        def decorator(fn):
            fn.name = getattr(fn, "__name__", "tool")
            return fn
        if callable(name_or_fn):
            return decorator(name_or_fn)
        return decorator

go_backend_url = os.getenv("GO_BACKEND_URL", "http://localhost:8080")


def _get_headers() -> Dict[str, str]:
    return {
        "X-LogisticsHQ-Service-Key": get_internal_service_token(),
        "Content-Type": "application/json",
    }


def _retrieve_context(org_id: int, primary_type: str, primary_id: int) -> Dict[str, Any]:
    """
    Calls the Go backend's internal context retrieval endpoint.
    """
    if org_id <= 0:
        return {"success": False, "error": "Invalid organization ID (must be > 0)"}
    if primary_id <= 0:
        return {"success": False, "error": "Invalid primary record ID (must be > 0)"}

    url = f"{go_backend_url}/internal/context/retrieve"
    payload = {
        "org_id": org_id,
        "primary_type": primary_type.upper(),
        "primary_id": primary_id,
    }

    try:
        with httpx.Client(timeout=12.0) as client:
            resp = client.post(url, headers=_get_headers(), json=payload)
            if resp.status_code == 200:
                data = resp.json()
                return {"success": True, "data": data.get("data", data)}
            elif resp.status_code == 404:
                return {"success": False, "error": f"{primary_type} #{primary_id} not found in organization #{org_id}"}
            else:
                return {"success": False, "error": f"Backend returned status {resp.status_code}: {resp.text}"}
    except Exception as e:
        return {"success": False, "error": f"Context retrieval failed: {str(e)}"}


def _retrieve_insight(org_id: int, entity_type: str, entity_id: int, question: str = "") -> Dict[str, Any]:
    """
    Calls the Go backend's internal intelligence insight endpoint.
    """
    if org_id <= 0:
        return {"success": False, "error": "Invalid organization ID"}
    if entity_id <= 0:
        return {"success": False, "error": "Invalid entity ID"}

    url = f"{go_backend_url}/internal/intelligence/insight"
    payload = {
        "org_id": org_id,
        "entity_type": entity_type.upper(),
        "entity_id": entity_id,
        "question": question,
    }

    try:
        with httpx.Client(timeout=15.0) as client:
            resp = client.post(url, headers=_get_headers(), json=payload)
            if resp.status_code == 200:
                data = resp.json()
                return {"success": True, "data": data.get("data", data)}
            elif resp.status_code == 404:
                return {"success": False, "error": f"{entity_type} #{entity_id} not found in organization"}
            else:
                return {"success": False, "error": f"Backend returned status {resp.status_code}: {resp.text}"}
    except Exception as e:
        return {"success": False, "error": f"Intelligence insight failed: {str(e)}"}


# ── LangChain Tool Schemas ───────────────────────────────────────────────────

class EntityContextInput(BaseModel):
    org_id: int = Field(..., description="Organization ID for multi-tenant isolation")
    record_id: int = Field(..., description="Primary record database ID")


class InsightInput(BaseModel):
    org_id: int = Field(..., description="Organization ID for multi-tenant isolation")
    entity_type: str = Field(..., description="Entity type: RFQ, SHIPMENT, INVOICE, CUSTOMER")
    entity_id: int = Field(..., description="Entity primary key ID")
    question: Optional[str] = Field(default="", description="Specific question or context request")


# ── Read-Only Tools for Future AI Agents ─────────────────────────────────────

@tool("get_customer_context", args_schema=EntityContextInput)
def get_customer_context(org_id: int, record_id: int) -> Dict[str, Any]:
    """Retrieve 360-degree customer context: profile, credit health, linked RFQs, invoices, and quotations (Read-Only)."""
    return _retrieve_context(org_id, "CUSTOMER", record_id)


@tool("get_lead_context", args_schema=EntityContextInput)
def get_lead_context(org_id: int, record_id: int) -> Dict[str, Any]:
    """Retrieve lead context: company name, contact, status, and related trade inquiries (Read-Only)."""
    return _retrieve_context(org_id, "LEAD", record_id)


@tool("get_rfq_context", args_schema=EntityContextInput)
def get_rfq_context(org_id: int, record_id: int) -> Dict[str, Any]:
    """Retrieve RFQ context: origin, destination, cargo specs, customer, and generated quotations (Read-Only)."""
    return _retrieve_context(org_id, "RFQ", record_id)


@tool("get_quotation_context", args_schema=EntityContextInput)
def get_quotation_context(org_id: int, record_id: int) -> Dict[str, Any]:
    """Retrieve commercial quotation context: pricing breakdown, margin, validity dates, and linked RFQ (Read-Only)."""
    return _retrieve_context(org_id, "QUOTATION", record_id)


@tool("get_shipment_context", args_schema=EntityContextInput)
def get_shipment_context(org_id: int, record_id: int) -> Dict[str, Any]:
    """Retrieve operational shipment context: vessel, voyage, milestones, exceptions, and carrier SCAC (Read-Only)."""
    return _retrieve_context(org_id, "SHIPMENT", record_id)


@tool("get_invoice_context", args_schema=EntityContextInput)
def get_invoice_context(org_id: int, record_id: int) -> Dict[str, Any]:
    """Retrieve financial invoice context: amounts due, payment status, customer, and linked shipment/debit notes (Read-Only)."""
    return _retrieve_context(org_id, "INVOICE", record_id)


@tool("get_contract_context", args_schema=EntityContextInput)
def get_contract_context(org_id: int, record_id: int) -> Dict[str, Any]:
    """Retrieve commercial contract context: terms, parties, dates, and value (Read-Only)."""
    return _retrieve_context(org_id, "CONTRACT", record_id)


@tool("get_related_business_records")
def get_related_business_records(org_id: int, entity_type: str, entity_id: int) -> Dict[str, Any]:
    """Retrieve all related business records linked to a primary entity across all modules (Read-Only)."""
    res = _retrieve_context(org_id, entity_type, entity_id)
    if not res.get("success"):
        return res
    ctx_data = res.get("data", {})
    return {
        "success": True,
        "primary_record": ctx_data.get("primary_record"),
        "related_records": ctx_data.get("related_records", {}),
        "source_references": ctx_data.get("source_references", []),
        "is_read_only": True,
    }


@tool("get_recent_activity")
def get_recent_activity(org_id: int, entity_type: str, entity_id: int) -> Dict[str, Any]:
    """Retrieve recent milestone progression and exceptions for an operational record (Read-Only)."""
    res = _retrieve_context(org_id, entity_type, entity_id)
    if not res.get("success"):
        return res
    ctx_data = res.get("data", {})
    return {
        "success": True,
        "milestones": ctx_data.get("milestones", []),
        "operational_exceptions": ctx_data.get("operational_exceptions", []),
        "recent_activity": ctx_data.get("recent_activity", []),
        "is_read_only": True,
    }


@tool("get_audit_context")
def get_audit_context(org_id: int, entity_type: str, entity_id: int) -> Dict[str, Any]:
    """Retrieve verifiable source references and audit citations for an entity (Read-Only)."""
    res = _retrieve_context(org_id, entity_type, entity_id)
    if not res.get("success"):
        return res
    ctx_data = res.get("data", {})
    return {
        "success": True,
        "source_references": ctx_data.get("source_references", []),
        "data_freshness": ctx_data.get("data_freshness"),
        "is_read_only": True,
    }


@tool("get_business_intelligence_insight", args_schema=InsightInput)
def get_business_intelligence_insight(org_id: int, entity_type: str, entity_id: int, question: str = "") -> Dict[str, Any]:
    """Generate deterministic, grounded business intelligence citing real records, field references, and operational warnings (Read-Only)."""
    return _retrieve_insight(org_id, entity_type, entity_id, question)


class CustomerIntelligenceInput(BaseModel):
    org_id: int = Field(description="Organization ID to scope the customer query strictly (server-enforced).")
    customer_id: int = Field(description="Primary customer record ID.")
    question: Optional[str] = Field(default="", description="Specific operational or commercial inquiry.")


def _retrieve_customer_intelligence(org_id: int, customer_id: int) -> Dict[str, Any]:
    """
    Calls the Go backend's internal customer 360 intelligence endpoint.
    """
    if org_id <= 0:
        return {"success": False, "error": "Invalid organization ID (must be > 0)"}
    if customer_id <= 0:
        return {"success": False, "error": "Invalid customer ID (must be > 0)"}

    url = f"{go_backend_url}/internal/customers/intelligence"
    payload = {
        "org_id": org_id,
        "customer_id": customer_id,
    }

    try:
        with httpx.Client(timeout=15.0) as client:
            resp = client.post(url, headers=_get_headers(), json=payload)
            if resp.status_code == 200:
                data = resp.json()
                return {"success": True, "data": data.get("data", {})}
            elif resp.status_code == 404:
                return {"success": False, "error": f"Customer {customer_id} not found in organization."}
            elif resp.status_code == 401:
                return {"success": False, "error": "Internal service authentication failure."}
            else:
                return {"success": False, "error": f"Backend returned HTTP {resp.status_code}: {resp.text}"}
    except httpx.RequestError as exc:
        return {"success": False, "error": f"Failed to connect to internal context service: {str(exc)}"}


@tool("get_customer_intelligence_360", args_schema=CustomerIntelligenceInput)
def get_customer_intelligence_360(org_id: int, customer_id: int, question: str = "") -> Dict[str, Any]:
    """Retrieve complete 360-degree customer intelligence including commercial metrics, operations, finances, governance, and grounded AI summary (Strictly Read-Only)."""
    return _retrieve_customer_intelligence(org_id, customer_id)


class RFQPricingIntelligenceInput(BaseModel):
    org_id: int = Field(description="Organization ID to scope the RFQ query strictly (server-enforced).")
    rfq_id: int = Field(description="Primary RFQ record ID.")
    question: Optional[str] = Field(default="", description="Specific pricing or margin inquiry.")


def _retrieve_rfq_pricing_intelligence(org_id: int, rfq_id: int) -> Dict[str, Any]:
    """
    Calls the Go backend's internal RFQ and pricing intelligence endpoint.
    """
    if org_id <= 0:
        return {"success": False, "error": "Invalid organization ID (must be > 0)"}
    if rfq_id <= 0:
        return {"success": False, "error": "Invalid RFQ ID (must be > 0)"}

    url = f"{go_backend_url}/internal/rfqs/intelligence"
    payload = {
        "org_id": org_id,
        "rfq_id": rfq_id,
    }

    try:
        with httpx.Client(timeout=15.0) as client:
            resp = client.post(url, headers=_get_headers(), json=payload)
            if resp.status_code == 200:
                data = resp.json()
                return {"success": True, "data": data.get("data", {})}
            elif resp.status_code == 404:
                return {"success": False, "error": f"RFQ {rfq_id} not found in organization."}
            elif resp.status_code == 401:
                return {"success": False, "error": "Internal service authentication failure."}
            else:
                return {"success": False, "error": f"Backend returned HTTP {resp.status_code}: {resp.text}"}
    except httpx.RequestError as exc:
        return {"success": False, "error": f"Failed to connect to internal context service: {str(exc)}"}


@tool("get_rfq_pricing_intelligence", args_schema=RFQPricingIntelligenceInput)
def get_rfq_pricing_intelligence(org_id: int, rfq_id: int, question: str = "") -> Dict[str, Any]:
    """Retrieve complete RFQ and Pricing intelligence including quotation spread, commercial conversion, margin health, and grounded AI analysis (Strictly Read-Only)."""
    return _retrieve_rfq_pricing_intelligence(org_id, rfq_id)


class ShipmentOperationsIntelligenceInput(BaseModel):
    org_id: int = Field(description="Organization ID to scope the shipment query strictly (server-enforced).")
    shipment_id: int = Field(description="Primary shipment record ID.")
    question: Optional[str] = Field(default="", description="Specific operational or milestone inquiry.")


def _retrieve_shipment_operations_intelligence(org_id: int, shipment_id: int) -> Dict[str, Any]:
    """
    Calls the Go backend's internal shipment and operations intelligence endpoint.
    """
    if org_id <= 0:
        return {"success": False, "error": "Invalid organization ID (must be > 0)"}
    if shipment_id <= 0:
        return {"success": False, "error": "Invalid shipment ID (must be > 0)"}

    url = f"{go_backend_url}/internal/shipments/intelligence"
    payload = {
        "org_id": org_id,
        "shipment_id": shipment_id,
    }

    try:
        with httpx.Client(timeout=15.0) as client:
            resp = client.post(url, headers=_get_headers(), json=payload)
            if resp.status_code == 200:
                data = resp.json()
                return {"success": True, "data": data.get("data", {})}
            elif resp.status_code == 404:
                return {"success": False, "error": f"Shipment {shipment_id} not found in organization."}
            elif resp.status_code == 401:
                return {"success": False, "error": "Internal service authentication failure."}
            else:
                return {"success": False, "error": f"Backend returned HTTP {resp.status_code}: {resp.text}"}
    except httpx.RequestError as exc:
        return {"success": False, "error": f"Failed to connect to internal context service: {str(exc)}"}


@tool("get_shipment_operations_intelligence", args_schema=ShipmentOperationsIntelligenceInput)
def get_shipment_operations_intelligence(org_id: int, shipment_id: int, question: str = "") -> Dict[str, Any]:
    """Retrieve complete Shipment and Operations Intelligence including milestone progression, open exceptions, schedule adherence, and grounded AI operational analysis (Strictly Read-Only)."""
    return _retrieve_shipment_operations_intelligence(org_id, shipment_id)


class OrgOperationsSummaryInput(BaseModel):
    org_id: int = Field(description="Organization ID to aggregate operational health for.")


@tool("get_org_operations_summary", args_schema=OrgOperationsSummaryInput)
def get_org_operations_summary(org_id: int) -> Dict[str, Any]:
    """Retrieve organization-wide operational health summary including active shipments, delayed counts, critical exceptions, and risk distribution (Strictly Read-Only)."""
    if org_id <= 0:
        return {"success": False, "error": "Invalid organization ID (must be > 0)"}

    url = f"{go_backend_url}/internal/shipments/operations-summary"
    payload = {"org_id": org_id}

    try:
        with httpx.Client(timeout=15.0) as client:
            resp = client.post(url, headers=_get_headers(), json=payload)
            if resp.status_code == 200:
                data = resp.json()
                return {"success": True, "data": data.get("data", {})}
            elif resp.status_code == 401:
                return {"success": False, "error": "Internal service authentication failure."}
            else:
                return {"success": False, "error": f"Backend returned HTTP {resp.status_code}: {resp.text}"}
    except httpx.RequestError as exc:
        return {"success": False, "error": f"Failed to connect to internal context service: {str(exc)}"}


# ── Invoice Finance Intelligence Tool (Task 1.5) ───────────────────────────────

class InvoiceFinanceIntelligenceInput(BaseModel):
    org_id: int = Field(description="Organization ID that owns the invoice.")
    invoice_id: int = Field(description="Database ID of the invoice to evaluate.")
    question: Optional[str] = Field(default="", description="Specific finance or receivables question.")


def _retrieve_invoice_finance_intelligence(org_id: int, invoice_id: int) -> Dict[str, Any]:
    if org_id <= 0:
        return {"success": False, "error": "Invalid organization ID (must be > 0)"}
    if invoice_id <= 0:
        return {"success": False, "error": "Invalid invoice ID (must be > 0)"}

    url = f"{go_backend_url}/internal/invoices/intelligence"
    payload = {"org_id": org_id, "invoice_id": invoice_id}

    try:
        with httpx.Client(timeout=15.0) as client:
            resp = client.post(url, headers=_get_headers(), json=payload)
            if resp.status_code == 200:
                data = resp.json()
                return {"success": True, "data": data.get("data", {})}
            elif resp.status_code == 404:
                return {"success": False, "error": f"Invoice {invoice_id} not found in organization."}
            elif resp.status_code == 401:
                return {"success": False, "error": "Internal service authentication failure."}
            else:
                return {"success": False, "error": f"Backend returned HTTP {resp.status_code}: {resp.text}"}
    except httpx.RequestError as exc:
        return {"success": False, "error": f"Failed to connect to internal context service: {str(exc)}"}


@tool("get_invoice_finance_intelligence", args_schema=InvoiceFinanceIntelligenceInput)
def get_invoice_finance_intelligence(org_id: int, invoice_id: int, question: str = "") -> Dict[str, Any]:
    """Retrieve complete Invoice and Finance Intelligence including aging buckets, receivables exposure, line item audit, margin review, and grounded AI finance analysis (Strictly Read-Only)."""
    return _retrieve_invoice_finance_intelligence(org_id, invoice_id)


class OrgFinanceSummaryInput(BaseModel):
    org_id: int = Field(description="Organization ID to aggregate accounts receivable and financial health for.")


@tool("get_org_finance_summary", args_schema=OrgFinanceSummaryInput)
def get_org_finance_summary(org_id: int) -> Dict[str, Any]:
    """Retrieve organization-wide accounts receivable summary including open invoices, overdue amounts, aging buckets, and top customer exposures (Strictly Read-Only)."""
    if org_id <= 0:
        return {"success": False, "error": "Invalid organization ID (must be > 0)"}

    url = f"{go_backend_url}/internal/invoices/finance-summary"
    payload = {"org_id": org_id}

    try:
        with httpx.Client(timeout=15.0) as client:
            resp = client.post(url, headers=_get_headers(), json=payload)
            if resp.status_code == 200:
                data = resp.json()
                return {"success": True, "data": data.get("data", {})}
            elif resp.status_code == 401:
                return {"success": False, "error": "Internal service authentication failure."}
            else:
                return {"success": False, "error": f"Backend returned HTTP {resp.status_code}: {resp.text}"}
    except httpx.RequestError as exc:
        return {"success": False, "error": f"Failed to connect to internal context service: {str(exc)}"}


# ── Contract & Compliance Intelligence Tools (Task 1.6) ───────────────────────

class ContractComplianceIntelligenceInput(BaseModel):
    org_id: int = Field(description="Organization ID that owns the contract.")
    contract_id: int = Field(description="Database ID of the contract to evaluate.")
    question: Optional[str] = Field(default="", description="Specific contract or compliance question.")


def _retrieve_contract_compliance_intelligence(org_id: int, contract_id: int) -> Dict[str, Any]:
    if org_id <= 0:
        return {"success": False, "error": "Invalid organization ID (must be > 0)"}
    if contract_id <= 0:
        return {"success": False, "error": "Invalid contract ID (must be > 0)"}

    url = f"{go_backend_url}/internal/contracts/intelligence"
    payload = {"org_id": org_id, "contract_id": contract_id}

    try:
        with httpx.Client(timeout=15.0) as client:
            resp = client.post(url, headers=_get_headers(), json=payload)
            if resp.status_code == 200:
                data = resp.json()
                return {"success": True, "data": data.get("data", {})}
            elif resp.status_code == 404:
                return {"success": False, "error": f"Contract {contract_id} not found in organization."}
            elif resp.status_code == 401:
                return {"success": False, "error": "Internal service authentication failure."}
            else:
                return {"success": False, "error": f"Backend returned HTTP {resp.status_code}: {resp.text}"}
    except httpx.RequestError as exc:
        return {"success": False, "error": f"Failed to connect to internal context service: {str(exc)}"}


@tool("get_contract_compliance_intelligence", args_schema=ContractComplianceIntelligenceInput)
def get_contract_compliance_intelligence(org_id: int, contract_id: int, question: str = "") -> Dict[str, Any]:
    """Retrieve complete Contract & Compliance Intelligence including lifecycle dates, commercial terms, obligations, compliance requirements, risk score, and grounded AI analysis (Strictly Read-Only)."""
    return _retrieve_contract_compliance_intelligence(org_id, contract_id)


class ContractCoverageInput(BaseModel):
    org_id: int = Field(description="Organization ID for multi-tenant isolation.")
    entity_type: str = Field(description="Entity type to check coverage for: SHIPMENT, INVOICE, QUOTATION.")
    entity_id: int = Field(description="Database primary key ID of the entity.")
    question: Optional[str] = Field(default="", description="Specific contract coverage question.")


def _retrieve_contract_coverage(org_id: int, entity_type: str, entity_id: int) -> Dict[str, Any]:
    if org_id <= 0:
        return {"success": False, "error": "Invalid organization ID (must be > 0)"}
    if entity_id <= 0:
        return {"success": False, "error": "Invalid entity ID (must be > 0)"}

    url = f"{go_backend_url}/internal/contracts/coverage"
    payload = {"org_id": org_id, "entity_type": entity_type.upper(), "entity_id": entity_id}

    try:
        with httpx.Client(timeout=15.0) as client:
            resp = client.post(url, headers=_get_headers(), json=payload)
            if resp.status_code == 200:
                data = resp.json()
                return {"success": True, "data": data.get("data", {})}
            elif resp.status_code == 404:
                return {"success": False, "error": f"{entity_type} {entity_id} not found in organization."}
            elif resp.status_code == 401:
                return {"success": False, "error": "Internal service authentication failure."}
            else:
                return {"success": False, "error": f"Backend returned HTTP {resp.status_code}: {resp.text}"}
    except httpx.RequestError as exc:
        return {"success": False, "error": f"Failed to connect to internal context service: {str(exc)}"}


@tool("get_contract_coverage", args_schema=ContractCoverageInput)
def get_contract_coverage(org_id: int, entity_type: str, entity_id: int, question: str = "") -> Dict[str, Any]:
    """Evaluate whether an operational shipment, invoice, or quotation is backed by a valid, active contract and identify coverage gaps (Strictly Read-Only)."""
    return _retrieve_contract_coverage(org_id, entity_type, entity_id)


class OrgContractComplianceSummaryInput(BaseModel):
    org_id: int = Field(description="Organization ID to aggregate contract and compliance health for.")


@tool("get_org_contract_compliance_summary", args_schema=OrgContractComplianceSummaryInput)
def get_org_contract_compliance_summary(org_id: int) -> Dict[str, Any]:
    """Retrieve organization-wide contract health and compliance summary including expiring agreements, missing documentation, and high-severity compliance issues (Strictly Read-Only)."""
    if org_id <= 0:
        return {"success": False, "error": "Invalid organization ID (must be > 0)"}

    url = f"{go_backend_url}/internal/contracts/compliance-summary"
    payload = {"org_id": org_id}

    try:
        with httpx.Client(timeout=15.0) as client:
            resp = client.post(url, headers=_get_headers(), json=payload)
            if resp.status_code == 200:
                data = resp.json()
                return {"success": True, "data": data.get("data", {})}
            elif resp.status_code == 401:
                return {"success": False, "error": "Internal service authentication failure."}
            else:
                return {"success": False, "error": f"Backend returned HTTP {resp.status_code}: {resp.text}"}
    except httpx.RequestError as exc:
        return {"success": False, "error": f"Failed to connect to internal context service: {str(exc)}"}


class CrossModuleInsightsInput(BaseModel):
    org_id: int = Field(description="Organization ID to scope the cross-module insights query.")
    entity_type: str = Field(default="", description="Optional entity type (CUSTOMER, SHIPMENT, INVOICE, CONTRACT, RFQ, QUOTATION). Leave empty for organization-wide insights.")
    entity_id: int = Field(default=0, description="Optional entity ID to analyze. Leave 0 for organization-wide insights.")
    question: str = Field(default="", description="Specific cross-module investigation question from the operator.")


@tool("get_cross_module_insights", args_schema=CrossModuleInsightsInput)
def get_cross_module_insights(org_id: int, entity_type: str = "", entity_id: int = 0, question: str = "") -> Dict[str, Any]:
    """Assemble cross-module connected business signals, relationship evidence, and grounded risk insights across customers, shipments, invoices, contracts, and RFQs (Strictly Read-Only)."""
    if org_id <= 0:
        return {"success": False, "error": "Invalid organization ID (must be > 0)"}

    url = f"{go_backend_url}/internal/insights/cross-module"
    payload = {
        "org_id": org_id,
        "entity_type": entity_type.strip().upper() if entity_type else "",
        "entity_id": entity_id if entity_id > 0 else 0,
    }

    try:
        with httpx.Client(timeout=15.0) as client:
            resp = client.post(url, headers=_get_headers(), json=payload)
            if resp.status_code == 200:
                data = resp.json()
                return {"success": True, "data": data.get("data", {})}
            elif resp.status_code == 404:
                return {"success": False, "error": f"Entity {entity_type} #{entity_id} not found in organization."}
            elif resp.status_code == 401:
                return {"success": False, "error": "Internal service authentication failure."}
            else:
                return {"success": False, "error": f"Backend returned HTTP {resp.status_code}: {resp.text}"}
    except httpx.RequestError as exc:
        return {"success": False, "error": f"Failed to connect to internal context service: {str(exc)}"}


class OrgCrossModuleSummaryInput(BaseModel):
    org_id: int = Field(description="Organization ID to aggregate cross-module insights and priorities for.")


@tool("get_org_cross_module_summary", args_schema=OrgCrossModuleSummaryInput)
def get_org_cross_module_summary(org_id: int) -> Dict[str, Any]:
    """Retrieve organization-level prioritized cross-module business insights, connected module health, and top risk signals (Strictly Read-Only)."""
    if org_id <= 0:
        return {"success": False, "error": "Invalid organization ID (must be > 0)"}

    url = f"{go_backend_url}/internal/insights/summary"
    payload = {"org_id": org_id}

    try:
        with httpx.Client(timeout=15.0) as client:
            resp = client.post(url, headers=_get_headers(), json=payload)
            if resp.status_code == 200:
                data = resp.json()
                return {"success": True, "data": data.get("data", {})}
            elif resp.status_code == 401:
                return {"success": False, "error": "Internal service authentication failure."}
            else:
                return {"success": False, "error": f"Backend returned HTTP {resp.status_code}: {resp.text}"}
    except httpx.RequestError as exc:
        return {"success": False, "error": f"Failed to connect to internal context service: {str(exc)}"}






