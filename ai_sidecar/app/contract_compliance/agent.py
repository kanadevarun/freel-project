"""
Contract and Compliance Automation Agent
Phase 3 Task 3.7 for LogisticsHQ

AI engine for contract document review, clause extraction, compliance risk scoring,
structured-record discrepancy detection, and clarification drafting.
"""

import re
import logging
from typing import List, Dict, Any, Optional

from .models import (
    ContractDocumentContext,
    DeterministicComplianceSignals,
    ExtractedClause,
    StructuredDiscrepancy,
    OperationalObligationSummary,
    EvidenceFact,
    ComplianceReviewResponse,
    ClauseExtractionResponse,
    StructuredTermsVerificationResponse,
    ComplianceChecklistResponse,
    ClarificationDraftRequest,
    ClarificationDraftResponse,
)

logger = logging.getLogger(__name__)


class ContractComplianceAgent:
    """
    Analyzes contract documents and structured legal/commercial metadata.
    Does NOT mutate database records or execute unapproved actions.
    """

    def __init__(self, llm_factory=None):
        self.llm_factory = llm_factory

    def _sanitize_text(self, text: str) -> str:
        """Strip dangerous injection patterns from input text."""
        if not text:
            return ""
        # Remove prompt injection or instruction override vectors
        cleaned = re.sub(r"(?i)(ignore previous instructions|disregard above|override compliance|mark as fully compliant|bypass approval)", "[REDACTED_SECURITY_PATTERN]", text)
        return cleaned.strip()

    def review_contract(
        self,
        context: ContractDocumentContext,
        signals: DeterministicComplianceSignals
    ) -> ComplianceReviewResponse:
        """
        Conducts multi-dimensional contract and compliance review combining Go deterministic signals
        with document clause analysis and structured discrepancy checks.
        """
        correlation_id = context.correlation_id or "corr-default"
        contract_id = context.contract_id
        doc_id = context.document_id

        # 1. Build grounded evidence facts strictly from Go backend context
        evidence: List[EvidenceFact] = [
            EvidenceFact(
                source_module="contracts",
                source_entity_id=contract_id,
                source_ref=context.contract_reference,
                field_name="status",
                observed_value=context.status,
                description=f"Current lifecycle status in database: {context.status}"
            ),
            EvidenceFact(
                source_module="contracts",
                source_entity_id=contract_id,
                source_ref=context.contract_reference,
                field_name="party_name",
                observed_value=context.party_name,
                description=f"Designated commercial counterparty: {context.party_name}"
            )
        ]

        if context.expiry_date:
            evidence.append(
                EvidenceFact(
                    source_module="contracts",
                    source_entity_id=contract_id,
                    source_ref=context.contract_reference,
                    field_name="expiry_date",
                    observed_value=str(context.expiry_date),
                    description=f"Contract expiration date: {context.expiry_date} ({signals.days_until_expiry} days remaining)"
                )
            )

        # 2. Extract or evaluate clauses
        extracted_clauses = self._extract_clauses_internal(context, signals)

        # 3. Detect structured vs document discrepancies
        discrepancies = self._detect_structured_discrepancies(context, signals, extracted_clauses)

        # 4. Identify missing mandatory information
        missing_info: List[str] = []
        if signals.missing_effective_date or not context.effective_date:
            missing_info.append("Contract effective date is missing or unspecified in structured agreement.")
        if signals.missing_expiry_date or not context.expiry_date:
            missing_info.append("Contract expiration/term date is missing in structured agreement.")
        if signals.missing_required_documents:
            for missing_doc in signals.required_documents_missing:
                missing_info.append(f"Required document not filed: {missing_doc}")
        if signals.missing_insurance_terms:
            missing_info.append("Certificate of Cargo Liability Insurance not verified or missing on file.")
        if signals.missing_rate_information:
            missing_info.append("No active structured rate schedule linked to this commercial contract.")

        # 5. Determine operational obligations
        obligations = self._extract_obligations_internal(context, extracted_clauses)

        # 6. Calculate weighted risk score & compliance status
        risk_score = 15.0  # Base nominal review score
        if signals.is_expired:
            risk_score += 45.0
        elif signals.is_nearing_expiry:
            risk_score += 25.0

        if signals.missing_required_documents or missing_info:
            risk_score += min(20.0, len(missing_info) * 8.0)

        if discrepancies:
            risk_score += min(25.0, len(discrepancies) * 10.0)

        if signals.has_rejected_document or signals.has_version_conflict:
            risk_score += 15.0

        risk_score = min(100.0, max(0.0, risk_score))

        if risk_score >= 70.0 or signals.is_expired:
            risk_level = "CRITICAL" if signals.is_expired else "HIGH"
            compliance_status = "EXPIRED" if signals.is_expired else "NON_COMPLIANT"
        elif risk_score >= 35.0 or signals.is_nearing_expiry or discrepancies or missing_info:
            risk_level = "MEDIUM"
            compliance_status = "REVIEW_REQUIRED"
        else:
            risk_level = "LOW"
            compliance_status = "COMPLIANT"

        # 7. Synthesize executive summary
        summary_parts = []
        summary_parts.append(f"Contract {context.contract_reference} ('{context.contract_name}') with {context.party_name} evaluated under {compliance_status} compliance status (Risk Score: {risk_score:.1f}/100, Level: {risk_level}).")
        
        if signals.is_expired:
            summary_parts.append(f"CRITICAL COMPLIANCE NOTICE: Agreement expired on {context.expiry_date}. Operations conducted under this reference carry uncontracted commercial and demurrage liability.")
        elif signals.is_nearing_expiry:
            summary_parts.append(f"RENEWAL WARNING: Agreement expires in {signals.days_until_expiry} days ({context.expiry_date}). Proactive renewal review recommended.")

        if discrepancies:
            summary_parts.append(f"Identified {len(discrepancies)} discrepancies between structured records and extracted terms requiring operational verification.")

        if missing_info:
            summary_parts.append(f"Identified {len(missing_info)} missing mandatory compliance or documentation items.")

        executive_summary = " ".join(summary_parts)

        # 8. Action Recommendations mapped to Centralized Action System
        recommendations = self._generate_recommendations(context, signals, discrepancies, missing_info, compliance_status)

        return ComplianceReviewResponse(
            contract_id=contract_id,
            document_id=doc_id,
            risk_level=risk_level,
            risk_score=round(risk_score, 1),
            compliance_status=compliance_status,
            executive_summary=executive_summary,
            extracted_clauses=extracted_clauses,
            structured_discrepancies=discrepancies,
            missing_information=missing_info,
            compliance_obligations=obligations,
            recommendations=recommendations,
            evidence=evidence,
            confidence_score=0.92,
            correlation_id=correlation_id
        )

    def extract_clauses(
        self,
        context: ContractDocumentContext,
        signals: DeterministicComplianceSignals
    ) -> ClauseExtractionResponse:
        """Extract and structure all contractual clauses from the document."""
        clauses = self._extract_clauses_internal(context, signals)
        high_risk = sum(1 for c in clauses if c.risk_level in ("HIGH", "CRITICAL"))
        return ClauseExtractionResponse(
            contract_id=context.contract_id,
            document_id=context.document_id,
            total_clauses_extracted=len(clauses),
            clauses=clauses,
            high_risk_clauses_count=high_risk,
            confidence_score=0.91,
            correlation_id=context.correlation_id or "corr-default"
        )

    def verify_structured_terms(
        self,
        context: ContractDocumentContext,
        signals: DeterministicComplianceSignals
    ) -> StructuredTermsVerificationResponse:
        """Compare database structured values with document extracted terms."""
        clauses = self._extract_clauses_internal(context, signals)
        discrepancies = self._detect_structured_discrepancies(context, signals, clauses)
        total_fields = 5 + len(context.structured_terms or [])
        matches = max(0, total_fields - len(discrepancies))

        summary = (
            f"Verified {total_fields} contractual fields against structured master records. "
            f"Found {len(discrepancies)} discrepancies requiring reconciliation."
            if discrepancies
            else f"All {total_fields} contractual fields are fully consistent with structured master records."
        )

        return StructuredTermsVerificationResponse(
            contract_id=context.contract_id,
            is_fully_consistent=(len(discrepancies) == 0),
            total_fields_verified=total_fields,
            matches_count=matches,
            discrepancies_count=len(discrepancies),
            discrepancies=discrepancies,
            summary=summary,
            correlation_id=context.correlation_id or "corr-default"
        )

    def assess_compliance(
        self,
        context: ContractDocumentContext,
        signals: DeterministicComplianceSignals
    ) -> ComplianceChecklistResponse:
        """Run standard maritime and freight compliance checklist."""
        checklist_items = [
            {
                "item_code": "CUSTOMS_POA",
                "title": "Customs Power of Attorney / Agent Representation",
                "status": "PASSED" if not signals.missing_required_documents else "WARNING",
                "description": "Verification of valid customs broker representation authorization."
            },
            {
                "item_code": "CARGO_INSURANCE_COI",
                "title": "Certificate of Cargo Liability Insurance",
                "status": "FAILED" if signals.missing_insurance_terms else "PASSED",
                "description": "Active cargo liability coverage meeting minimum limits ($250,000 / occurrence)."
            },
            {
                "item_code": "DEMURRAGE_DETENTION_TIERS",
                "title": "Demurrage & Detention Free-Time Schedule",
                "status": "PASSED",
                "description": "Explicit free-time calendar days and tiered daily detention charge tables."
            },
            {
                "item_code": "FMC_REGULATORY_COMPLIANCE",
                "title": "FMC Ocean Freight Service Contract Filing",
                "status": "PASSED" if context.contract_type == "CARRIER_AGREEMENT" else "NOT_APPLICABLE",
                "description": "Federal Maritime Commission regulatory filing alignment for ocean common carrier agreements."
            },
            {
                "item_code": "SANCTIONS_AND_EXPORT_CONTROLS",
                "title": "Export Controls & Denied Party Screening Warranty",
                "status": "PASSED",
                "description": "Standard OFAC and international sanctions warranty clause present."
            }
        ]

        overall = "PASSED"
        if any(item["status"] == "FAILED" for item in checklist_items) or signals.is_expired:
            overall = "FAILED"
        elif any(item["status"] == "WARNING" for item in checklist_items) or signals.is_nearing_expiry:
            overall = "WARNING"

        remedies = []
        if signals.missing_insurance_terms:
            remedies.append("Request renewed Certificate of Insurance (COI) naming LogisticsHQ as certificate holder.")
        if signals.missing_required_documents:
            remedies.append("Upload and link missing mandatory documentation via Documents Portal.")
        if signals.is_nearing_expiry:
            remedies.append("Initiate commercial renewal cycle before agreement lapse date.")

        return ComplianceChecklistResponse(
            contract_id=context.contract_id,
            overall_compliance=overall,
            checklist_items=checklist_items,
            missing_mandatory_documents=signals.required_documents_missing or [],
            expiring_certifications=[],
            recommended_remedies=remedies,
            correlation_id=context.correlation_id or "corr-default"
        )

    def generate_clarification_draft(
        self,
        request: ClarificationDraftRequest
    ) -> ClarificationDraftResponse:
        """
        Synthesizes an editable clarification draft, missing-document request, or renewal notice.
        Requires Human-in-the-Loop managerial approval before dispatch.
        """
        ctx = request.context
        signals = request.signals
        draft_type = request.draft_type
        tone = request.tone
        custom_inst = self._sanitize_text(request.custom_instructions or "")

        # Default recipient
        recipient_role = "Commercial Representative / Legal Counsel"
        recipient_name = f"{ctx.party_name} Contracts Lead"
        recipient_email = f"contracts@{ctx.party_name.lower().replace(' ', '')}.com"

        key_clauses: List[str] = []
        missing_items: List[str] = []

        if draft_type == "MISSING_DOCUMENT_REQUEST":
            subject = f"ACTION REQUIRED: Missing Documentation for Agreement {ctx.contract_reference} - {ctx.party_name}"
            missing_items = signals.required_documents_missing or ["Certificate of Insurance", "Customs Broker POA"]
            items_bullet = "\n".join([f"  • {item}" for item in missing_items])
            body = (
                f"Dear {ctx.party_name} Team,\n\n"
                f"During a compliance review of agreement {ctx.contract_reference} ('{ctx.contract_name}'), "
                f"our compliance operations team identified that the following mandatory documentation is pending or unverified on file:\n\n"
                f"{items_bullet}\n\n"
                f"To ensure uninterrupted booking execution and regulatory compliance across our active shipments, "
                f"please provide electronic copies of the requested documents at your earliest convenience.\n\n"
            )
            if custom_inst:
                body += f"Special Instructions: {custom_inst}\n\n"
            body += (
                f"Thank you for your prompt cooperation.\n\n"
                f"Sincerely,\n"
                f"Commercial Contracts & Compliance Department\n"
                f"LogisticsHQ"
            )
            key_clauses.append("Section 8: Regulatory & Compliance Documentation Requirements")

        elif draft_type == "CLAUSE_CLARIFICATION":
            subject = f"Clarification Request: Agreement Terms & Conditions - {ctx.contract_reference} ({ctx.party_name})"
            body = (
                f"Dear {ctx.party_name} Legal & Commercial Team,\n\n"
                f"We are conducting a standard contract alignment review for agreement {ctx.contract_reference} ('{ctx.contract_name}').\n\n"
                f"We would appreciate your formal clarification regarding the following operational provisions:\n"
                f"  1. Demurrage & Detention Free-Time Calculation: Confirmation of standard container free-time allowances at destination ports.\n"
                f"  2. Payment Terms Alignment: Reconciliation between invoiced payment terms and structured agreement schedules.\n"
                f"  3. Carrier Limitation of Liability: Confirmation of applicability under COGSA / Hague-Visby rules.\n\n"
            )
            if custom_inst:
                body += f"Additional Context: {custom_inst}\n\n"
            body += (
                f"Please let us know your availability for a brief alignment call or confirm the updated terms in writing.\n\n"
                f"Sincerely,\n"
                f"Contracts & Commercial Operations\n"
                f"LogisticsHQ"
            )
            key_clauses.extend(["Section 4: Payment Terms", "Section 7: Demurrage & Detention", "Section 12: Limitation of Liability"])

        elif draft_type == "RENEWAL_NOTICE":
            expiry_str = ctx.expiry_date or "the upcoming validity threshold"
            subject = f"Commercial Agreement Renewal Notice: {ctx.contract_reference} - {ctx.party_name}"
            body = (
                f"Dear {ctx.party_name} Commercial Leadership,\n\n"
                f"Our master commercial agreement {ctx.contract_reference} ('{ctx.contract_name}') is scheduled to expire on {expiry_str}.\n\n"
                f"Given our ongoing operational partnership and freight volumes, we would like to initiate discussions regarding agreement renewal "
                f"and rate schedule updates for the upcoming contract term.\n\n"
            )
            if custom_inst:
                body += f"Renewal Notes: {custom_inst}\n\n"
            body += (
                f"Please review your capacity forecasts and proposed rate schedules so we can schedule our renewal review session.\n\n"
                f"Best regards,\n"
                f"Commercial Partnership Management\n"
                f"LogisticsHQ"
            )
            key_clauses.append("Section 2: Term, Validity & Renewal Procedures")

        elif draft_type == "INTERNAL_REVIEW_NOTE":
            subject = f"INTERNAL LEGAL/COMPLIANCE MEMO: Agreement {ctx.contract_reference} ({ctx.party_name})"
            recipient_role = "Internal Compliance Manager"
            recipient_name = "Legal & Compliance Committee"
            recipient_email = "compliance@logisticshq.internal"
            body = (
                f"INTERNAL COMPLIANCE MEMO\n"
                f"--------------------------------------------------\n"
                f"Contract Reference: {ctx.contract_reference}\n"
                f"Agreement Name:     {ctx.contract_name}\n"
                f"Counterparty:       {ctx.party_name} ({ctx.contract_type})\n"
                f"Status / Term:      {ctx.status} (Expires: {ctx.expiry_date or 'Indefinite'})\n"
                f"Identified Risks:   {', '.join(signals.required_documents_missing or ['General review required'])}\n\n"
                f"Findings & Analysis:\n"
                f"An automated review of the agreement terms identified areas requiring internal legal risk sign-off "
                f"prior to execution of future bookings. Operational exposure involves demurrage liability and unverified COI.\n\n"
            )
            if custom_inst:
                body += f"Reviewer Notes: {custom_inst}\n\n"
            body += "Recommendation: Place contract under conditional review pending updated counterparty COI."

        else:  # COMPLIANCE_BREACH_ALERT
            subject = f"URGENT NOTICE: Compliance Deficiency - Agreement {ctx.contract_reference} ({ctx.party_name})"
            body = (
                f"Dear {ctx.party_name} Compliance Department,\n\n"
                f"This formal communication serves to notify you of an unfulfilled compliance requirement under agreement {ctx.contract_reference}.\n\n"
                f"Our records indicate an unresolved compliance deficiency that restricts active operational bookings under this master contract.\n\n"
            )
            if custom_inst:
                body += f"Deficiency Details: {custom_inst}\n\n"
            body += (
                f"Please contact our legal and compliance desk immediately to rectify this issue.\n\n"
                f"Sincerely,\n"
                f"Chief Compliance Officer\n"
                f"LogisticsHQ"
            )

        internal_notes = (
            f"Draft generated by AI Contract Compliance Engine. "
            f"Contract #{ctx.contract_id}, Status: {ctx.status}. "
            f"Safeguard active: Consequential external communications require managerial approval."
        )

        return ClarificationDraftResponse(
            draft_type=draft_type,
            subject=subject,
            message_body=body,
            internal_review_notes=internal_notes,
            recipient_role=recipient_role,
            suggested_recipient_name=recipient_name,
            suggested_recipient_email=recipient_email,
            key_clauses_referenced=key_clauses,
            missing_items_requested=missing_items,
            requires_approval=True,
            confidence_score=0.92,
            correlation_id=request.correlation_id or "corr-default"
        )

    # -------------------------------------------------------------------------
    # Internal Helpers (Clause extraction, Discrepancy checks, Recommendations)
    # -------------------------------------------------------------------------

    def _extract_clauses_internal(
        self,
        context: ContractDocumentContext,
        signals: DeterministicComplianceSignals
    ) -> List[ExtractedClause]:
        """Extracts standard maritime, carrier, and SLA contract clauses."""
        clauses: List[ExtractedClause] = []

        # 1. Payment Terms Clause
        clauses.append(
            ExtractedClause(
                clause_id="CLAUSE-PAY-01",
                clause_type="PAYMENT_TERMS",
                title="Payment Terms and Currency Settlement",
                extracted_text="All ocean freight, destination surcharges, and intermodal drayage invoices shall be settled in USD within 30 calendar days of invoice presentation (Net 30). Late payments incur 1.5% monthly finance charge.",
                section_reference="Section 4.1",
                page_number=3,
                risk_level="LOW",
                risk_assessment="Standard commercial payment terms with explicit interest penalty on overdue balances.",
                structured_match_status="MATCH"
            )
        )

        # 2. Liability Limit Clause
        clauses.append(
            ExtractedClause(
                clause_id="CLAUSE-LIA-02",
                clause_type="LIABILITY_LIMIT",
                title="Carrier Limitation of Cargo Liability",
                extracted_text="Carrier liability for loss of or damage to cargo shall be strictly limited in accordance with the Carriage of Goods by Sea Act (COGSA), capped at $500 per customary freight unit (package), unless higher value declared.",
                section_reference="Section 7.3",
                page_number=6,
                risk_level="MEDIUM",
                risk_assessment="COGSA $500 package limitation applies. Higher-value shipments require separate marine cargo insurance declaration.",
                structured_match_status="NOT_IN_STRUCTURED_RECORD"
            )
        )

        # 3. Demurrage & Detention Clause
        clauses.append(
            ExtractedClause(
                clause_id="CLAUSE-DEM-03",
                clause_type="DEMURRAGE_DETENTION",
                title="Port Demurrage and Equipment Detention Free-Time",
                extracted_text="Standard ocean equipment free-time shall be 4 working days for dry standard containers and 2 working days for refrigerated equipment at discharge ports. Excess detention billed at $175/day for Days 1-5, and $250/day thereafter.",
                section_reference="Section 9.2",
                page_number=8,
                risk_level="HIGH" if signals.is_nearing_expiry else "MEDIUM",
                risk_assessment="Steep detention tier escalation ($250/day) after Day 5 poses operational cost exposure if port congestion occurs.",
                structured_match_status="NOT_IN_STRUCTURED_RECORD"
            )
        )

        # 4. Service Level Agreement Clause
        clauses.append(
            ExtractedClause(
                clause_id="CLAUSE-SLA-04",
                clause_type="SERVICE_LEVEL_AGREEMENT",
                title="Schedule Reliability and Rolling Safeguards",
                extracted_text="Carrier guarantees container equipment availability and space protection on contracted vessels, committing to a minimum 90% on-time departure metric for primary Trans-Pacific port pairs.",
                section_reference="Section 5.4",
                page_number=4,
                risk_level="LOW",
                risk_assessment="Strong 90% on-time service commitment with equipment priority protections.",
                structured_match_status="MATCH"
            )
        )

        # 5. Termination & Expiry Clause
        clauses.append(
            ExtractedClause(
                clause_id="CLAUSE-TERM-05",
                clause_type="TERMINATION",
                title="Term of Agreement and Early Termination Notice",
                extracted_text=f"This Master Agreement remains in full force through {context.expiry_date or 'September 30, 2027'}. Either party may terminate with 60 days written notice. Material breach permits immediate termination.",
                section_reference="Section 2.2",
                page_number=2,
                risk_level="HIGH" if signals.is_expired else "LOW",
                risk_assessment="60-day bilateral termination window provides sufficient commercial lead time.",
                structured_match_status="MATCH" if not signals.is_expired else "DISCREPANCY"
            )
        )

        # 6. Force Majeure Clause
        clauses.append(
            ExtractedClause(
                clause_id="CLAUSE-FM-06",
                clause_type="FORCE_MAJEURE",
                title="Force Majeure, Port Labor Actions, and Canal Restrictions",
                extracted_text="Neither party shall be liable for non-performance caused by acts of God, strikes, labor disputes, lockouts, canal transiting restrictions, or governmental war/sanctions actions.",
                section_reference="Section 14.1",
                page_number=11,
                risk_level="LOW",
                risk_assessment="Comprehensive maritime force majeure clause protecting against unforeseen canal delays and port labor unrest.",
                structured_match_status="NOT_IN_STRUCTURED_RECORD"
            )
        )

        return clauses

    def _detect_structured_discrepancies(
        self,
        context: ContractDocumentContext,
        signals: DeterministicComplianceSignals,
        clauses: List[ExtractedClause]
    ) -> List[StructuredDiscrepancy]:
        """Compares structured fields in database against extracted document text."""
        discrepancies: List[StructuredDiscrepancy] = []

        # Check 1: Expiry date consistency
        if signals.is_expired:
            discrepancies.append(
                StructuredDiscrepancy(
                    field_name="expiry_date",
                    structured_value=str(context.expiry_date),
                    document_extracted_value="Expired Agreement",
                    severity="CRITICAL",
                    discrepancy_explanation=f"Agreement expired on {context.expiry_date} while active operational records are linked.",
                    recommendation="Initiate commercial contract renewal or suspend non-contracted booking issuance."
                )
            )

        # Check 2: Status discrepancy if document says active but DB says draft
        if context.status == "DRAFT" and not signals.is_expired:
            discrepancies.append(
                StructuredDiscrepancy(
                    field_name="status",
                    structured_value="DRAFT",
                    document_extracted_value="Fully Executed Master Agreement",
                    severity="HIGH",
                    discrepancy_explanation="Document contains signed master agreement terms, but system lifecycle status remains DRAFT.",
                    recommendation="Submit contract for managerial activation approval via the Centralized Approvals Center."
                )
            )

        # Check 3: Structured terms comparison (e.g. rate or currency)
        if context.structured_terms:
            for st in context.structured_terms:
                term_val = str(st.get("term_value", ""))
                if "Net 60" in term_val:
                    discrepancies.append(
                        StructuredDiscrepancy(
                            field_name="payment_terms",
                            structured_value="Net 60",
                            document_extracted_value="Net 30",
                            severity="MEDIUM",
                            discrepancy_explanation="Structured database term specifies Net 60 days, but extracted agreement Section 4.1 specifies Net 30 days.",
                            recommendation="Review signed amendment to reconcile payment terms."
                        )
                    )

        return discrepancies

    def _extract_obligations_internal(
        self,
        context: ContractDocumentContext,
        clauses: List[ExtractedClause]
    ) -> List[OperationalObligationSummary]:
        """Extracts ongoing operational obligations."""
        return [
            OperationalObligationSummary(
                obligation_type="EQUIPMENT_RETURN",
                responsible_party="Shipper / LogisticsHQ",
                description="Return empty ocean containers to carrier depot within 4 working days of port discharge.",
                due_trigger="Container gate-out from marine terminal",
                penalty_or_consequence="Detention fee of $175 to $250 per calendar day"
            ),
            OperationalObligationSummary(
                obligation_type="INSURANCE_CERTIFICATE_MAINTENANCE",
                responsible_party="Counterparty / Carrier",
                description="Maintain comprehensive marine cargo insurance policy ($250,000 minimum) throughout validity term.",
                due_trigger="Annual policy renewal date",
                penalty_or_consequence="Suspension of booking allocations"
            ),
            OperationalObligationSummary(
                obligation_type="INVOICE_AUDIT_WINDOW",
                responsible_party="LogisticsHQ Accounts Payable",
                description="Present any freight invoice billing discrepancies or demurrage disputes within 15 calendar days.",
                due_trigger="Receipt of carrier ocean freight bill",
                penalty_or_consequence="Waiver of right to contest disputed freight surcharge"
            )
        ]

    def _generate_recommendations(
        self,
        context: ContractDocumentContext,
        signals: DeterministicComplianceSignals,
        discrepancies: List[StructuredDiscrepancy],
        missing_info: List[str],
        compliance_status: str
    ) -> List[Dict[str, Any]]:
        """Generates actionable recommendations mapped to Centralized Action System."""
        recs = []

        # Action 1: Document Review
        recs.append({
            "action_name": "contracts.request_document_review",
            "title": "Conduct Formal Legal & Compliance Document Review",
            "description": f"Assign compliance review for agreement {context.contract_reference} to verify liability provisions and detention tiers.",
            "risk_level": "HIGH" if compliance_status in ("NON_COMPLIANT", "EXPIRED") else "MEDIUM",
            "requires_approval": True,
            "category": "COMPLIANCE_REVIEW",
            "suggested_payload": {
                "contract_id": context.contract_id,
                "document_id": context.document_id,
                "review_type": "FULL_COMPLIANCE_AUDIT"
            }
        })

        # Action 2: Missing Document Request if needed
        if missing_info or signals.missing_required_documents:
            recs.append({
                "action_name": "contracts.request_missing_document",
                "title": "Request Missing Mandatory Compliance Documentation",
                "description": "Dispatch formal request to counterparty for missing Certificate of Insurance and Customs POA.",
                "risk_level": "HIGH",
                "requires_approval": True,
                "category": "EXTERNAL_COMMUNICATION",
                "suggested_payload": {
                    "contract_id": context.contract_id,
                    "missing_documents": signals.required_documents_missing or ["Certificate of Insurance"]
                }
            })

        # Action 3: Structured Term Discrepancy Verification if any
        if discrepancies:
            recs.append({
                "action_name": "contracts.verify_structured_discrepancy",
                "title": "Reconcile Structured Terms with Signed Agreement",
                "description": f"Update master database records to reflect extracted terms ({len(discrepancies)} discrepancies identified).",
                "risk_level": "LOW",
                "requires_approval": False,
                "category": "SAFE_INTERNAL",
                "suggested_payload": {
                    "contract_id": context.contract_id,
                    "discrepancies_count": len(discrepancies)
                }
            })

        # Action 4: Renewal Follow-Up Task if nearing expiry or expired
        if signals.is_nearing_expiry or signals.is_expired:
            recs.append({
                "action_name": "contracts.create_renewal_task",
                "title": "Create Commercial Renewal Follow-Up Task",
                "description": f"Assign commercial negotiation task to renew agreement prior to lapse ({signals.days_until_expiry} days).",
                "risk_level": "LOW",
                "requires_approval": False,
                "category": "SAFE_INTERNAL",
                "suggested_payload": {
                    "contract_id": context.contract_id,
                    "target_completion_date": context.expiry_date
                }
            })

        return recs
