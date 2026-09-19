"""
Contract and Compliance Automation Models
Phase 3 Task 3.7 for LogisticsHQ

Strict Pydantic domain models for contract document review, clause extraction,
compliance risk analysis, structured-record verification, and clarification drafting.
"""

from typing import List, Optional, Literal, Dict, Any
from pydantic import BaseModel, Field, field_validator, AliasChoices


class EvidenceFact(BaseModel):
    source_module: str = Field(..., description="Originating module (e.g. contracts, contract_documents, compliance, rates)")
    source_entity_id: int = Field(..., description="ID of source business record")
    source_ref: str = Field(..., description="Human-readable business reference code (e.g. CTR-TP-2026-01, DOC-102)")
    field_name: str = Field(..., description="Specific field evaluated")
    observed_value: str = Field(..., description="Factual value verified in database or document")
    description: str = Field(..., description="Contextual explanation of verified fact")


class ContractDocumentContext(BaseModel):
    """
    Authoritative contract and document context supplied strictly by the Go backend.
    """
    org_id: int
    contract_id: int
    contract_reference: str
    contract_name: str
    contract_type: str = "COMMERCIAL_AGREEMENT"
    party_id: Optional[int] = None
    party_name: str = "Commercial Counterparty"
    status: str = "ACTIVE"
    effective_date: Optional[str] = None
    expiry_date: Optional[str] = None
    contract_value: Optional[float] = 0.0
    currency: str = "USD"
    transport_mode: Optional[str] = "MULTIMODAL"
    owner: Optional[str] = "Commercial Operations"

    # Document details (if a specific document is being reviewed)
    document_id: Optional[str] = None
    document_name: Optional[str] = None
    file_type: Optional[str] = "PDF"
    document_text: Optional[str] = Field(default="", description="Authoritative extracted or OCR text of the contract document")

    # Structured database records to compare against document text
    structured_terms: Optional[List[Dict[str, Any]]] = Field(default_factory=list)
    structured_obligations: Optional[List[Dict[str, Any]]] = Field(default_factory=list)
    linked_documents: Optional[List[Dict[str, Any]]] = Field(default_factory=list)
    compliance_requirements: Optional[List[Dict[str, Any]]] = Field(default_factory=list)

    correlation_id: str = "corr-default"

    @field_validator("structured_terms", "structured_obligations", "linked_documents", "compliance_requirements", mode="before")
    @classmethod
    def ensure_list_fields(cls, v):
        if v is None:
            return []
        return v


class DeterministicComplianceSignals(BaseModel):
    """
    Authoritative compliance, lifecycle, and document signals evaluated deterministically in Go.
    """
    is_expired: bool = False
    is_nearing_expiry: bool = False
    days_until_expiry: int = 999
    missing_effective_date: bool = False
    missing_expiry_date: bool = False
    missing_required_documents: bool = False
    required_documents_missing: Optional[List[str]] = Field(default_factory=list)
    has_document_awaiting_review: bool = False
    has_rejected_document: bool = False
    has_version_conflict: bool = False
    has_incomplete_metadata: bool = False
    is_linked_to_inactive_agreement: bool = False
    missing_rate_information: bool = False
    missing_service_level_terms: bool = False
    missing_insurance_terms: bool = False
    has_structured_term_discrepancy: bool = False
    requires_manual_compliance_review: bool = False

    @field_validator("required_documents_missing", mode="before")
    @classmethod
    def ensure_docs_list(cls, v):
        if v is None:
            return []
        return v


class ExtractedClause(BaseModel):
    clause_id: str
    clause_type: Literal[
        "PAYMENT_TERMS",
        "LIABILITY_LIMIT",
        "INDEMNITY",
        "GOVERNING_LAW",
        "TERMINATION",
        "FORCE_MAJEURE",
        "SERVICE_LEVEL_AGREEMENT",
        "INSURANCE_REQUIREMENT",
        "RATE_STRUCTURE",
        "DEMURRAGE_DETENTION",
        "CUSTOMS_COMPLIANCE",
        "GENERAL_PROVISION"
    ]
    title: str
    extracted_text: str = Field(..., description="Verbatim or summarized text extracted directly from document")
    section_reference: Optional[str] = None
    page_number: Optional[int] = None
    risk_level: Literal["LOW", "MEDIUM", "HIGH", "CRITICAL"] = "LOW"
    risk_assessment: str
    structured_match_status: Literal["MATCH", "DISCREPANCY", "NOT_IN_STRUCTURED_RECORD", "AMBIGUOUS"] = "NOT_IN_STRUCTURED_RECORD"


class StructuredDiscrepancy(BaseModel):
    field_name: str
    structured_value: str
    document_extracted_value: str
    severity: Literal["LOW", "MEDIUM", "HIGH", "CRITICAL"]
    discrepancy_explanation: str
    recommendation: str


class OperationalObligationSummary(BaseModel):
    obligation_type: str
    responsible_party: str
    description: str
    due_trigger: str
    penalty_or_consequence: Optional[str] = None


class ComplianceReviewResponse(BaseModel):
    contract_id: int
    document_id: Optional[str] = None
    risk_level: Literal["LOW", "MEDIUM", "HIGH", "CRITICAL"]
    risk_score: float = Field(..., ge=0.0, le=100.0, description="Risk score from 0.00 to 100.00")
    compliance_status: Literal["COMPLIANT", "REVIEW_REQUIRED", "NON_COMPLIANT", "EXPIRED"]
    executive_summary: str
    extracted_clauses: List[ExtractedClause] = Field(default_factory=list)
    structured_discrepancies: List[StructuredDiscrepancy] = Field(default_factory=list)
    missing_information: List[str] = Field(default_factory=list)
    compliance_obligations: List[OperationalObligationSummary] = Field(default_factory=list)
    recommendations: List[Dict[str, Any]] = Field(default_factory=list)
    evidence: List[EvidenceFact] = Field(default_factory=list)
    confidence_score: float = Field(..., ge=0.0, le=1.0)
    correlation_id: str


class ClauseExtractionResponse(BaseModel):
    contract_id: int
    document_id: Optional[str] = None
    total_clauses_extracted: int
    clauses: List[ExtractedClause] = Field(default_factory=list)
    high_risk_clauses_count: int = 0
    confidence_score: float = 0.90
    correlation_id: str


class StructuredTermsVerificationResponse(BaseModel):
    contract_id: int
    is_fully_consistent: bool
    total_fields_verified: int
    matches_count: int
    discrepancies_count: int
    discrepancies: List[StructuredDiscrepancy] = Field(default_factory=list)
    summary: str
    correlation_id: str


class ComplianceChecklistResponse(BaseModel):
    contract_id: int
    overall_compliance: Literal["PASSED", "WARNING", "FAILED"]
    checklist_items: List[Dict[str, Any]] = Field(default_factory=list)
    missing_mandatory_documents: List[str] = Field(default_factory=list)
    expiring_certifications: List[Dict[str, Any]] = Field(default_factory=list)
    recommended_remedies: List[str] = Field(default_factory=list)
    correlation_id: str


class ClarificationDraftRequest(BaseModel):
    context: ContractDocumentContext
    signals: DeterministicComplianceSignals
    draft_type: Literal[
        "MISSING_DOCUMENT_REQUEST",
        "CLAUSE_CLARIFICATION",
        "RENEWAL_NOTICE",
        "INTERNAL_REVIEW_NOTE",
        "COMPLIANCE_BREACH_ALERT"
    ] = "CLAUSE_CLARIFICATION"
    tone: Literal["POLITE", "ASSERTIVE", "URGENT", "FORMAL", "PROFESSIONAL", "FIRM_FORMAL", "ACCOMMODATING"] = "PROFESSIONAL"
    custom_instructions: Optional[str] = None
    target_clause_id: Optional[str] = None
    correlation_id: str = "corr-default"


class ClarificationDraftResponse(BaseModel):
    draft_type: str
    subject: str
    message_body: str
    internal_review_notes: Optional[str] = None
    recipient_role: str
    suggested_recipient_name: str
    suggested_recipient_email: str
    key_clauses_referenced: List[str] = Field(default_factory=list)
    missing_items_requested: List[str] = Field(default_factory=list)
    requires_approval: bool = True
    confidence_score: float = 0.92
    correlation_id: str
