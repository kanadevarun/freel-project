"""
RFQ-to-Quotation Automation and Intelligent Pricing Workflow Agent
Phase 3 Task 3.4 for LogisticsHQ

Pure Python AI Agent providing structured RFQ requirement extraction,
grounded pricing explanations, commercial risk detection, and quotation drafting.
Adheres strictly to the requirement: Zero direct database mutations;
authoritative pricing and approval decisions remain governed by Go backend.
"""

from datetime import datetime, timedelta
import re
from typing import List, Dict, Any, Optional

from app.rfq_pricing.models import (
    RFQContext,
    EvidenceFact,
    SafetyWarning,
    ExtractedRequirementField,
    ExtractedRFQRequirements,
    DeterministicPricingFacts,
    PricingExplanation,
    QuotationRiskAnalysis,
    QuotationDraftRequest,
    QuotationDraftResponse,
)


class RFQPricingWorkflowAgent:
    """
    Dedicated AI agent for RFQ intake, pricing intelligence, and quotation preparation.
    """

    MANDATORY_FIELDS = [
        ("origin", "Origin Port / POL", "Port of loading is required for carrier route selection and tariff lookup."),
        ("destination", "Destination Port / POD", "Port of discharge is required for route determination and customs routing."),
        ("incoterms", "Incoterms", "Commercial incoterms (e.g. FOB, CIF, EXW) dictate risk and cost allocation."),
        ("cargo_description", "Cargo Description", "Accurate commodity description is required for carrier acceptance and classification."),
        ("cargo_weight", "Cargo Weight", "Gross weight in KG is required for vessel space, payload, and weight band pricing."),
    ]

    INJECTION_PATTERNS = [
        re.compile(r"ignore\s+(all\s+)?previous\s+instructions", re.IGNORECASE),
        re.compile(r"override\s+(all\s+)?pricing|set\s+price\s+to\s+0", re.IGNORECASE),
        re.compile(r"bypass\s+(manager\s+)?approval", re.IGNORECASE),
        re.compile(r"system\s*:\s*you\s+are", re.IGNORECASE),
        re.compile(r"grant\s+100%\s+discount", re.IGNORECASE),
    ]

    def _sanitize_input(self, text: Optional[str]) -> str:
        if not text:
            return ""
        sanitized = text.strip()
        for pattern in self.INJECTION_PATTERNS:
            if pattern.search(sanitized):
                sanitized = pattern.sub("[FILTERED_UNSAFE_INSTRUCTION]", sanitized)
        return sanitized

    def extract_requirements(self, ctx: RFQContext) -> ExtractedRFQRequirements:
        """
        Deterministically evaluates RFQ context and normalizes requirements.
        Never fabricates missing values; explicitly flags missing items.
        """
        fields: Dict[str, ExtractedRequirementField] = {}
        missing_mandatory: List[str] = []
        missing_optional: List[str] = []
        evidence: List[EvidenceFact] = []

        # 1. Origin
        if ctx.origin and ctx.origin.strip():
            fields["origin"] = ExtractedRequirementField(
                field_name="origin",
                value=ctx.origin.strip(),
                status="VERIFIED",
                confidence=0.98,
                source="rfqs.origin",
                evidence=f"Specified as POL {ctx.origin}"
            )
            evidence.append(EvidenceFact(
                source_module="rfqs",
                source_entity_id=ctx.rfq_id,
                source_ref=ctx.rfq_number,
                field_name="origin",
                observed_value=ctx.origin,
                description="Origin port specified in RFQ"
            ))
        else:
            fields["origin"] = ExtractedRequirementField(
                field_name="origin",
                value=None,
                status="MISSING",
                confidence=1.0,
                source="rfqs.origin",
                missing_impact="Cannot identify carrier lanes or freight tariffs without origin."
            )
            missing_mandatory.append("origin")

        # 2. Destination
        if ctx.destination and ctx.destination.strip():
            fields["destination"] = ExtractedRequirementField(
                field_name="destination",
                value=ctx.destination.strip(),
                status="VERIFIED",
                confidence=0.98,
                source="rfqs.destination",
                evidence=f"Specified as POD {ctx.destination}"
            )
            evidence.append(EvidenceFact(
                source_module="rfqs",
                source_entity_id=ctx.rfq_id,
                source_ref=ctx.rfq_number,
                field_name="destination",
                observed_value=ctx.destination,
                description="Destination port specified in RFQ"
            ))
        else:
            fields["destination"] = ExtractedRequirementField(
                field_name="destination",
                value=None,
                status="MISSING",
                confidence=1.0,
                source="rfqs.destination",
                missing_impact="Cannot price discharge port handling or ocean freight without destination."
            )
            missing_mandatory.append("destination")

        # 3. Incoterms
        if ctx.incoterms and ctx.incoterms.strip():
            fields["incoterms"] = ExtractedRequirementField(
                field_name="incoterms",
                value=ctx.incoterms.strip().upper(),
                status="VERIFIED",
                confidence=0.95,
                source="rfqs.incoterms",
                evidence=f"Commercial terms: {ctx.incoterms.upper()}"
            )
            evidence.append(EvidenceFact(
                source_module="rfqs",
                source_entity_id=ctx.rfq_id,
                source_ref=ctx.rfq_number,
                field_name="incoterms",
                observed_value=ctx.incoterms,
                description="Commercial trade terms defined"
            ))
        else:
            fields["incoterms"] = ExtractedRequirementField(
                field_name="incoterms",
                value=None,
                status="MISSING",
                confidence=1.0,
                source="rfqs.incoterms",
                missing_impact="Required to establish buyer vs seller liability and origin surcharge coverage."
            )
            missing_mandatory.append("incoterms")

        # 4. Cargo Items (Description & Weight)
        total_weight = sum(item.weight_kg for item in ctx.items if item.weight_kg > 0)
        total_volume = sum(item.volume_cbm for item in ctx.items if item.volume_cbm > 0)
        cargo_descs = [item.description for item in ctx.items if item.description]

        if cargo_descs:
            desc_val = ", ".join(cargo_descs)
            fields["cargo_description"] = ExtractedRequirementField(
                field_name="cargo_description",
                value=desc_val,
                status="EXTRACTED",
                confidence=0.92,
                source="rfq_items.description",
                evidence=f"Item descriptions: {desc_val}"
            )
        else:
            fields["cargo_description"] = ExtractedRequirementField(
                field_name="cargo_description",
                value=None,
                status="MISSING",
                confidence=1.0,
                source="rfq_items",
                missing_impact="Missing commodity details prevent accurate tariff classification and dangerous goods check."
            )
            missing_mandatory.append("cargo_description")

        if total_weight > 0:
            fields["cargo_weight"] = ExtractedRequirementField(
                field_name="cargo_weight",
                value=f"{total_weight:.2f} KG",
                status="VERIFIED",
                confidence=0.95,
                source="rfq_items.weight_kg",
                evidence=f"Sum of item weights: {total_weight:.2f} KG"
            )
        else:
            fields["cargo_weight"] = ExtractedRequirementField(
                field_name="cargo_weight",
                value=None,
                status="MISSING",
                confidence=1.0,
                source="rfq_items.weight_kg",
                missing_impact="Gross weight is needed to determine payload and heavy-lift surcharges."
            )
            missing_mandatory.append("cargo_weight")

        # Optional fields
        if total_volume > 0:
            fields["cargo_volume"] = ExtractedRequirementField(
                field_name="cargo_volume",
                value=f"{total_volume:.2f} CBM",
                status="VERIFIED",
                confidence=0.90,
                source="rfq_items.volume_cbm",
                evidence=f"Total volume: {total_volume:.2f} CBM"
            )
        else:
            missing_optional.append("cargo_volume")

        if ctx.target_date:
            fields["target_date"] = ExtractedRequirementField(
                field_name="target_date",
                value=ctx.target_date,
                status="VERIFIED",
                confidence=0.95,
                source="rfqs.target_date"
            )
        else:
            missing_optional.append("target_date")

        # Overall Status
        if missing_mandatory:
            status = "CLARIFICATION_REQUIRED"
            recommendation = (
                f"RFQ #{ctx.rfq_number} is missing mandatory commercial details: "
                f"{', '.join(missing_mandatory)}. An inquiry clarification request should be dispatched to the customer."
            )
            confidence = 0.85
        elif missing_optional:
            status = "INCOMPLETE"
            recommendation = (
                f"Core mandatory parameters verified for RFQ #{ctx.rfq_number}. "
                f"Optional fields missing ({', '.join(missing_optional)}), but quotation calculation can proceed."
            )
            confidence = 0.94
        else:
            status = "COMPLETE"
            recommendation = f"All operational and commercial requirements fully verified for RFQ #{ctx.rfq_number}."
            confidence = 0.99

        return ExtractedRFQRequirements(
            rfq_id=ctx.rfq_id,
            rfq_number=ctx.rfq_number,
            status=status,
            fields=fields,
            missing_mandatory=missing_mandatory,
            missing_optional=missing_optional,
            clarification_recommendation=recommendation,
            confidence_score=confidence,
            evidence=evidence,
            correlation_id=ctx.correlation_id,
        )

    def explain_pricing(self, ctx: RFQContext, pricing: DeterministicPricingFacts) -> PricingExplanation:
        """
        Explains deterministic pricing and cost/sell breakdown calculated by Go.
        Never modifies authoritative numbers.
        """
        warnings: List[SafetyWarning] = []
        evidence: List[EvidenceFact] = []

        # Base Cost & Surcharges Evidence
        evidence.append(EvidenceFact(
            source_module="pricing",
            source_entity_id=ctx.rfq_id,
            source_ref=ctx.rfq_number,
            field_name="total_cost",
            observed_value=f"{pricing.currency} {pricing.total_cost:.2f}",
            description="Authoritative total cost calculated by Go deterministic engine"
        ))

        evidence.append(EvidenceFact(
            source_module="pricing",
            source_entity_id=ctx.rfq_id,
            source_ref=ctx.rfq_number,
            field_name="total_selling_price",
            observed_value=f"{pricing.currency} {pricing.total_selling_price:.2f}",
            description="Authoritative total selling price calculated by Go deterministic engine"
        ))

        evidence.append(EvidenceFact(
            source_module="pricing",
            source_entity_id=ctx.rfq_id,
            source_ref=ctx.rfq_number,
            field_name="gross_margin_pct",
            observed_value=f"{pricing.gross_margin_pct:.2f}%",
            description=f"Calculated gross profit margin with health rating {pricing.margin_health}"
        ))

        # Check margin risks
        if pricing.margin_health == "NEGATIVE":
            warnings.append(SafetyWarning(
                warning_type="NEGATIVE_MARGIN",
                message=f"Quotation has negative gross profit ({pricing.currency} {pricing.gross_profit:.2f}, {pricing.gross_margin_pct:.2f}%). Execution blocked without senior manager signoff.",
                severity="CRITICAL"
            ))
        elif pricing.margin_health == "LOW":
            warnings.append(SafetyWarning(
                warning_type="LOW_MARGIN",
                message=f"Gross margin is thin ({pricing.gross_margin_pct:.2f}% < 15.0%). Manager approval required.",
                severity="WARNING"
            ))

        # Check rate expiry
        if pricing.rate_is_expired:
            warnings.append(SafetyWarning(
                warning_type="EXPIRED_RATE",
                message="Applied carrier tariff or rate contract has expired. Spot re-validation recommended.",
                severity="WARNING"
            ))

        carrier_name = pricing.applied_carrier_name or "Standard Carrier Tariff"
        cost_explanation = (
            f"Total carrier cost is {pricing.currency} {pricing.total_cost:,.2f}, comprised of base freight "
            f"({pricing.currency} {pricing.base_cost:,.2f}) and verified operational surcharges "
            f"({pricing.currency} {pricing.surcharges:,.2f}) from carrier {carrier_name}."
        )

        sell_notes = (
            f"Target selling price is {pricing.currency} {pricing.total_selling_price:,.2f}, yielding a gross profit of "
            f"{pricing.currency} {pricing.gross_profit:,.2f} ({pricing.gross_margin_pct:.2f}% margin). "
        )
        if pricing.discounts > 0:
            sell_notes += f"Reflects approved commercial discount of {pricing.currency} {pricing.discounts:,.2f}."

        commercial_nuance = (
            f"Pricing aligned for customer {ctx.customer_name} ({ctx.customer_tier} Tier, {ctx.payment_terms}) "
            f"on lane {ctx.origin or 'POL'} -> {ctx.destination or 'POD'} under {ctx.incoterms or 'standard'} incoterms."
        )

        return PricingExplanation(
            rfq_id=ctx.rfq_id,
            carrier_rate_selected=carrier_name,
            cost_explanation=cost_explanation,
            sell_pricing_notes=sell_notes,
            commercial_nuance=commercial_nuance,
            pricing_warnings=warnings,
            confidence_score=0.96,
            evidence=evidence,
            correlation_id=ctx.correlation_id,
        )

    def analyze_quotation_risks(self, ctx: RFQContext, pricing: DeterministicPricingFacts) -> QuotationRiskAnalysis:
        """
        Evaluates quotation risk profile and triggers Human-in-the-Loop safety requirements.
        """
        triggers: List[str] = []
        risk_reasons: List[str] = []
        warnings: List[SafetyWarning] = []
        evidence: List[EvidenceFact] = []
        next_steps: List[str] = []

        # 1. Margin Health Trigger
        if pricing.margin_health == "NEGATIVE":
            triggers.append("NEGATIVE_MARGIN_APPROVAL")
            risk_reasons.append(f"Negative gross margin ({pricing.gross_margin_pct:.2f}%)")
            warnings.append(SafetyWarning(
                warning_type="NEGATIVE_MARGIN",
                message="Loss-making quotation detected. Commercial director approval mandatory.",
                severity="CRITICAL"
            ))
            next_steps.append("Request commercial director signoff or increase freight sell price.")
        elif pricing.margin_health == "LOW":
            triggers.append("LOW_MARGIN_APPROVAL")
            risk_reasons.append(f"Gross margin below 15% threshold ({pricing.gross_margin_pct:.2f}%)")
            warnings.append(SafetyWarning(
                warning_type="LOW_MARGIN",
                message="Thin gross margin below 15% standard threshold.",
                severity="WARNING"
            ))
            next_steps.append("Submit quotation draft for manager pricing approval.")

        # 2. Rate Expiry Trigger
        if pricing.rate_is_expired:
            triggers.append("EXPIRED_RATE_APPROVAL")
            risk_reasons.append("Applied carrier rate contract is past validity date.")
            warnings.append(SafetyWarning(
                warning_type="RATE_EXPIRED",
                message="Carrier rate source expired. Confirm carrier spot rate before finalizing.",
                severity="WARNING"
            ))
            next_steps.append("Verify active carrier spot quote with carrier procurement.")

        # 3. Discount Trigger
        if pricing.discounts > 0:
            triggers.append("COMMERCIAL_DISCOUNT_APPROVAL")
            risk_reasons.append(f"Special commercial discount applied ({pricing.currency} {pricing.discounts:.2f})")
            next_steps.append("Ensure commercial discount complies with account guidelines.")

        # 4. Customer Credit Check
        if ctx.outstanding_balance > ctx.credit_limit and ctx.credit_limit > 0:
            triggers.append("CREDIT_LIMIT_EXCEEDED")
            risk_reasons.append(f"Customer outstanding balance exceeds credit limit (${ctx.outstanding_balance:,.2f} > ${ctx.credit_limit:,.2f})")
            warnings.append(SafetyWarning(
                warning_type="CREDIT_HOLD",
                message="Customer credit exposure exceeds limit. Finance approval required.",
                severity="CRITICAL"
            ))
            next_steps.append("Coordinate with Finance to review credit hold status.")

        # 5. Missing Mandatory Information
        if not ctx.origin or not ctx.destination or not ctx.incoterms:
            risk_reasons.append("Incomplete inquiry details")
            warnings.append(SafetyWarning(
                warning_type="INCOMPLETE_RFQ",
                message="Mandatory routing or incoterms missing from RFQ record.",
                severity="WARNING"
            ))
            next_steps.append("Request missing shipment parameters from customer.")

        # Overall Risk Level Determination
        if any(w.severity == "CRITICAL" for w in warnings):
            risk_level = "CRITICAL"
        elif len(triggers) > 0 or len(warnings) > 0:
            risk_level = "HIGH" if len(triggers) > 1 else "MEDIUM"
        else:
            risk_level = "LOW"
            next_steps.append("Generate draft quotation and review customer presentation.")

        requires_approval = len(triggers) > 0 or risk_level in ["HIGH", "CRITICAL"]

        evidence.append(EvidenceFact(
            source_module="quotations",
            source_entity_id=ctx.rfq_id,
            source_ref=ctx.rfq_number,
            field_name="risk_evaluation",
            observed_value=risk_level,
            description=f"Automated risk evaluation completed with {len(triggers)} approval triggers"
        ))

        return QuotationRiskAnalysis(
            rfq_id=ctx.rfq_id,
            risk_level=risk_level,
            requires_approval=requires_approval,
            approval_triggers=triggers,
            risk_reasons=risk_reasons,
            margin_health_evaluation=f"Margin health is {pricing.margin_health} ({pricing.gross_margin_pct:.2f}%)",
            rate_freshness_evaluation="Expired" if pricing.rate_is_expired else "Active",
            deadline_urgency="Target ready date: " + (ctx.target_date or "Immediate"),
            suggested_next_steps=next_steps,
            safety_warnings=warnings,
            confidence_score=0.97,
            evidence=evidence,
            correlation_id=ctx.correlation_id,
        )

    def generate_quotation_draft(self, req: QuotationDraftRequest) -> QuotationDraftResponse:
        """
        Synthesizes a grounded quotation draft based exclusively on Go-calculated financial facts.
        Prompt injection attempts in notes are sanitized.
        """
        sanitized_notes = self._sanitize_input(req.user_prompt_notes)

        # Dates
        now = datetime.utcnow()
        validity_start = now.strftime("%Y-%m-%d")
        validity_end = (now + timedelta(days=req.validity_days or 14)).strftime("%Y-%m-%d")

        p = req.pricing
        title = f"Formal Quotation — {req.origin} to {req.destination} ({req.shipment_mode})"

        # Internal Summary
        internal_summary = (
            f"Quotation for RFQ #{req.rfq_number} ({req.customer_name}). "
            f"Routing: {req.origin} -> {req.destination} under {req.incoterms} terms. "
            f"Total carrier cost: {p.currency} {p.total_cost:,.2f} | Total sell: {p.currency} {p.total_selling_price:,.2f}. "
            f"Projected gross profit: {p.currency} {p.gross_profit:,.2f} ({p.gross_margin_pct:.2f}%). "
            f"Validity period: {validity_start} to {validity_end} ({req.validity_days} days)."
        )
        if sanitized_notes:
            internal_summary += f" Commercial Notes: {sanitized_notes}"

        # Customer-facing professional wording
        recipient_name = req.contact_name or req.customer_name
        customer_wording = (
            f"Dear {recipient_name},\n\n"
            f"Thank you for your inquiry (Ref: #{req.rfq_number}). LogisticsHQ is pleased to present "
            f"our competitive rate proposal for your upcoming shipment from {req.origin} to {req.destination}.\n\n"
            f"Cargo Summary: {req.items_summary or 'General Commercial Freight'}\n"
            f"Shipment Mode: {req.shipment_mode}\n"
            f"Incoterms: {req.incoterms}\n\n"
            f"Commercial Quotation Summary ({p.currency}):\n"
            f"  • Base Freight: {p.currency} {p.base_sell:,.2f}\n"
            f"  • Destination / Ancillary Surcharges: {p.currency} {p.surcharges:,.2f}\n"
        )
        if p.discounts > 0:
            customer_wording += f"  • Promotional Discount: -{p.currency} {p.discounts:,.2f}\n"
        if p.tax_amount > 0:
            customer_wording += f"  • Taxes / VAT: {p.currency} {p.tax_amount:,.2f}\n"
        customer_wording += (
            f"  ──────────────────────────────────────────\n"
            f"  TOTAL ALL-IN RATE: {p.currency} {p.total_selling_price:,.2f}\n\n"
            f"This proposal is valid through {validity_end}. To confirm booking or request schedule options, "
            f"please reply directly to this communication or access your LogisticsHQ portal.\n\n"
            f"Best regards,\nLogisticsHQ Commercial Team"
        )

        pricing_explanation = (
            f"Base freight calculated at {p.currency} {p.base_sell:,.2f}. Verified surcharges totaling "
            f"{p.currency} {p.surcharges:,.2f} encompass terminal handling, documentation, and security fees. "
            f"Net margin is {p.gross_margin_pct:.2f}%."
        )

        terms = (
            f"1. Rates subject to equipment availability and space at time of carrier booking.\n"
            f"2. Rate valid from {validity_start} through {validity_end}.\n"
            f"3. Surcharges reflect current tariff levels and are subject to floating bunker and port adjustments.\n"
            f"4. Standard demurrage and detention free time applies per carrier schedule.\n"
            f"5. Payment terms: Subject to credit agreement ({p.currency})."
        )

        # Triggers
        triggers: List[str] = []
        warnings: List[SafetyWarning] = []
        if p.margin_health in ["LOW", "NEGATIVE"]:
            triggers.append(f"MARGIN_{p.margin_health}")
            warnings.append(SafetyWarning(
                warning_type="MARGIN_HEALTH",
                message=f"Quotation margin {p.gross_margin_pct:.2f}% requires managerial approval prior to release.",
                severity="WARNING"
            ))
        if p.discounts > 0:
            triggers.append("DISCOUNT_APPLIED")
        if p.rate_is_expired:
            triggers.append("EXPIRED_RATE")

        requires_approval = len(triggers) > 0

        evidence = [
            EvidenceFact(
                source_module="rfqs",
                source_entity_id=req.rfq_id,
                source_ref=req.rfq_number,
                field_name="quotation_draft",
                observed_value=f"{p.currency} {p.total_selling_price:.2f}",
                description="Synthesized quotation draft grounded on verified Go calculations"
            )
        ]

        recipient_preview = {
            "name": recipient_name,
            "email": req.contact_email or "pending-verification@customer.com",
            "organization": req.customer_name
        }

        return QuotationDraftResponse(
            rfq_id=req.rfq_id,
            rfq_number=req.rfq_number,
            quotation_title=title,
            internal_summary=internal_summary,
            customer_wording=customer_wording,
            pricing_explanation=pricing_explanation,
            terms_and_conditions=terms,
            validity_start=validity_start,
            validity_end=validity_end,
            recipient_preview=recipient_preview,
            requires_approval=requires_approval,
            approval_triggers=triggers,
            confidence_score=0.98,
            evidence=evidence,
            warnings=warnings,
            correlation_id=req.correlation_id,
        )
