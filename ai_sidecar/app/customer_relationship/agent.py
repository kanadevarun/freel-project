"""
Customer Relationship Intelligence Agent
Phase 3 Task 3.3 for LogisticsHQ

Implements Python-only AI decision logic, signal detection,
transparent prioritization, grounded evidence evaluation,
and controlled communication drafting.
"""

import hashlib
import time
from typing import List, Optional, Tuple
from app.customer_relationship.models import (
    CustomerFollowupContext,
    CustomerPriorityResult,
    CustomerEvaluationResponse,
    FollowupRecommendationOutput,
    EvidenceFact,
    SafetyWarning,
    DraftMessageRequest,
    DraftMessageResponse,
)


class CustomerRelationshipAgent:
    """
    Dedicated AI Agent for Customer Relationship Automation & Follow-Up.
    Operates strictly on structured context supplied by the Go backend.
    Does not access MariaDB or execute mutations directly.
    """

    def evaluate_customer(self, ctx: CustomerFollowupContext) -> CustomerEvaluationResponse:
        priority_result = self.evaluate_priority(ctx)
        recommendations = self.generate_recommendations(ctx, priority_result)
        
        summary = (
            f"Customer relationship assessment for {ctx.customer_name} (ID #{ctx.customer_id}): "
            f"Priority Score {priority_result.priority_score}/100 ({priority_result.priority_level}). "
            f"Primary factor: {priority_result.primary_reason} with {len(recommendations)} recommended follow-up action(s)."
        )
        
        correlation_id = ctx.correlation_id or f"cr-eval-{ctx.customer_id}-{int(time.time()*1000)}"

        return CustomerEvaluationResponse(
            status="SUCCESS",
            customer_id=ctx.customer_id,
            customer_name=ctx.customer_name,
            priority=priority_result,
            recommendations=recommendations,
            summary=summary,
            correlation_id=correlation_id,
        )

    def evaluate_priority(self, ctx: CustomerFollowupContext) -> CustomerPriorityResult:
        """
        Computes a transparent, deterministic priority rating and risk level
        based exclusively on verifiable business facts in the supplied context.
        """
        score = 40  # Baseline neutral score
        factors: List[str] = []
        evidence: List[EvidenceFact] = []
        source_ref = ctx.source_reference or f"CUST-{ctx.customer_id}"

        # 1. Financial & Credit Signals
        if ctx.overdue_balance > 0:
            score += 25
            factors.append(f"Outstanding overdue invoice balance of ${ctx.overdue_balance:,.2f}")
            evidence.append(EvidenceFact(
                source_module="invoices",
                source_entity_id=ctx.customer_id,
                source_ref=source_ref,
                field_name="overdue_balance",
                observed_value=f"${ctx.overdue_balance:,.2f}",
                description="Unpaid past-due invoice balance recorded in accounts receivable",
            ))

        if ctx.credit_status in ("ON_HOLD", "CREDIT_HOLD"):
            score += 20
            factors.append("Commercial credit status is currently On Hold")
            evidence.append(EvidenceFact(
                source_module="customers",
                source_entity_id=ctx.customer_id,
                source_ref=source_ref,
                field_name="credit_status",
                observed_value=str(ctx.credit_status),
                description="Customer commercial credit is restricted",
            ))

        # 2. Operational Exceptions & Delays
        if ctx.unresolved_exception_count > 0:
            score += 25
            factors.append(f"{ctx.unresolved_exception_count} unresolved active shipment exception(s)")
            evidence.append(EvidenceFact(
                source_module="shipments",
                source_entity_id=ctx.customer_id,
                source_ref=source_ref,
                field_name="unresolved_exceptions",
                observed_value=str(ctx.unresolved_exception_count),
                description="Active exceptions on shipments requiring carrier or customer coordination",
            ))

        if ctx.delayed_shipment_count > 0:
            score += 15
            factors.append(f"{ctx.delayed_shipment_count} shipment(s) currently experiencing transit delays")
            evidence.append(EvidenceFact(
                source_module="shipments",
                source_entity_id=ctx.customer_id,
                source_ref=source_ref,
                field_name="delayed_shipments",
                observed_value=str(ctx.delayed_shipment_count),
                description="Active freight shipments delayed past estimated milestone ETA",
            ))

        # 3. Commercial Contracts & Pipeline
        if ctx.expiring_contract_count > 0:
            score += 15
            factors.append(f"{ctx.expiring_contract_count} commercial contract(s) approaching expiration")
            evidence.append(EvidenceFact(
                source_module="contracts",
                source_entity_id=ctx.customer_id,
                source_ref=source_ref,
                field_name="expiring_contracts",
                observed_value=str(ctx.expiring_contract_count),
                description="Master rate contract agreement renewal deadline approaching",
            ))

        if ctx.open_quote_count > 0:
            score += 10
            factors.append(f"{ctx.open_quote_count} submitted quotation(s) awaiting customer feedback")
            evidence.append(EvidenceFact(
                source_module="quotations",
                source_entity_id=ctx.customer_id,
                source_ref=source_ref,
                field_name="open_quotations",
                observed_value=str(ctx.open_quote_count),
                description="Quotation options pending commercial acceptance",
            ))

        # 4. Inactivity Detection
        if ctx.days_inactive >= 30 and ctx.active_shipment_count == 0 and ctx.rfq_count == 0:
            score += 15
            factors.append(f"Account inactive for {ctx.days_inactive} days with no new bookings or RFQs")
            evidence.append(EvidenceFact(
                source_module="customers",
                source_entity_id=ctx.customer_id,
                source_ref=source_ref,
                field_name="days_inactive",
                observed_value=f"{ctx.days_inactive} days",
                description="Customer has had no freight inquiries or active bookings over 30 days",
            ))

        # 5. Account Ownership & Contact Completeness
        if not ctx.account_owner_id:
            score += 5
            factors.append("No commercial account manager assigned")
            evidence.append(EvidenceFact(
                source_module="customers",
                source_entity_id=ctx.customer_id,
                source_ref=source_ref,
                field_name="account_owner",
                observed_value="Unassigned",
                description="Account has no primary commercial owner assigned in CRM",
            ))

        if not ctx.primary_contact_name or not ctx.primary_contact_email:
            score += 5
            factors.append("Missing primary contact details (name or email)")
            evidence.append(EvidenceFact(
                source_module="customers",
                source_entity_id=ctx.customer_id,
                source_ref=source_ref,
                field_name="primary_contact",
                observed_value="Incomplete",
                description="Primary commercial or operational contact information missing",
            ))

        # Normalize score
        score = max(0, min(100, score))

        if score >= 80:
            p_level = "CRITICAL"
            r_level = "CRITICAL"
        elif score >= 65:
            p_level = "HIGH"
            r_level = "HIGH"
        elif score >= 45:
            p_level = "MEDIUM"
            r_level = "MEDIUM"
        else:
            p_level = "LOW"
            r_level = "LOW"

        primary_reason = factors[0] if factors else "Standard commercial account monitoring"
        confidence = 0.94 if evidence else 0.70

        return CustomerPriorityResult(
            priority_score=score,
            priority_level=p_level,
            risk_level=r_level,
            primary_reason=primary_reason,
            contributing_factors=factors,
            confidence=confidence,
            evidence=evidence,
            correlation_id=ctx.correlation_id,
        )

    def generate_recommendations(
        self, ctx: CustomerFollowupContext, priority: CustomerPriorityResult
    ) -> List[FollowupRecommendationOutput]:
        """
        Synthesizes grounded follow-up recommendations based on evaluated signals.
        Each recommendation includes stable deduplication keys.
        """
        recs: List[FollowupRecommendationOutput] = []
        source_ref = ctx.source_reference or f"CUST-{ctx.customer_id}"
        corr = ctx.correlation_id or f"corr-rec-{ctx.customer_id}-{int(time.time()*1000)}"

        # 1. Overdue Invoice / Collections Follow-Up
        if ctx.overdue_balance > 0 or ctx.credit_status in ("ON_HOLD", "CREDIT_HOLD"):
            dedup = self._compute_dedup(ctx.org_id, "customers", ctx.customer_id, "CUSTOMER_INVOICE_FOLLOWUP")
            ev_list = [e for e in priority.evidence if e.source_module in ("invoices", "customers")]
            recs.append(FollowupRecommendationOutput(
                customer_id=ctx.customer_id,
                customer_name=ctx.customer_name,
                source_module="invoices",
                source_record_id=ctx.customer_id,
                source_reference=source_ref,
                trigger_reason=f"Overdue billing balance of ${ctx.overdue_balance:,.2f} requires commercial alignment",
                category="CUSTOMER_FOLLOWUP",
                priority="HIGH" if ctx.overdue_balance > 5000 or ctx.credit_status == "ON_HOLD" else "MEDIUM",
                risk_level="HIGH",
                confidence="VERY_HIGH",
                confidence_score=0.96,
                evidence=ev_list,
                recommended_action="Coordinate with finance and primary contact regarding overdue ledger status.",
                action_type="SCHEDULE_INVOICE_REMINDER",
                followup_type="PAYMENT_OVERDUE",
                suggested_owner_id=ctx.account_owner_id,
                suggested_owner_name=ctx.account_owner_name,
                requires_approval=True,
                correlation_id=corr,
                dedup_key=dedup,
            ))

        # 2. Shipment Exception / Delay Follow-Up
        if ctx.unresolved_exception_count > 0 or ctx.delayed_shipment_count > 0:
            dedup = self._compute_dedup(ctx.org_id, "customers", ctx.customer_id, "CUSTOMER_SHIPMENT_EXCEPTION_FOLLOWUP")
            ev_list = [e for e in priority.evidence if e.source_module == "shipments"]
            recs.append(FollowupRecommendationOutput(
                customer_id=ctx.customer_id,
                customer_name=ctx.customer_name,
                source_module="shipments",
                source_record_id=ctx.customer_id,
                source_reference=source_ref,
                trigger_reason=f"{ctx.unresolved_exception_count} exception(s) / {ctx.delayed_shipment_count} delayed shipment(s) pending customer advisory",
                category="CUSTOMER_FOLLOWUP",
                priority="CRITICAL" if ctx.unresolved_exception_count > 0 else "HIGH",
                risk_level="HIGH",
                confidence="HIGH",
                confidence_score=0.92,
                evidence=ev_list,
                recommended_action="Draft verified operational transit update explaining delay mitigation to shipping desk.",
                action_type="SEND_SHIPMENT_UPDATE",
                followup_type="SHIPMENT_EXCEPTION_ADVISORY",
                suggested_owner_id=ctx.account_owner_id,
                suggested_owner_name=ctx.account_owner_name,
                requires_approval=True,
                correlation_id=corr,
                dedup_key=dedup,
            ))

        # 3. Open Quotation Follow-Up
        if ctx.open_quote_count > 0:
            dedup = self._compute_dedup(ctx.org_id, "customers", ctx.customer_id, "CUSTOMER_QUOTE_FOLLOWUP")
            ev_list = [e for e in priority.evidence if e.source_module == "quotations"]
            recs.append(FollowupRecommendationOutput(
                customer_id=ctx.customer_id,
                customer_name=ctx.customer_name,
                source_module="quotations",
                source_record_id=ctx.customer_id,
                source_reference=source_ref,
                trigger_reason=f"{ctx.open_quote_count} active rate quote(s) awaiting customer commercial feedback",
                category="CUSTOMER_FOLLOWUP",
                priority="MEDIUM",
                risk_level="LOW",
                confidence="HIGH",
                confidence_score=0.90,
                evidence=ev_list,
                recommended_action="Follow up with client shipping manager to address route questions and confirm booking interest.",
                action_type="QUOTE_CHECKIN",
                followup_type="QUOTE_PENDING",
                suggested_owner_id=ctx.account_owner_id,
                suggested_owner_name=ctx.account_owner_name,
                requires_approval=True,
                correlation_id=corr,
                dedup_key=dedup,
            ))

        # 4. Inactivity Re-Engagement
        if ctx.days_inactive >= 30 and ctx.active_shipment_count == 0:
            dedup = self._compute_dedup(ctx.org_id, "customers", ctx.customer_id, "CUSTOMER_INACTIVITY_REENGAGEMENT")
            ev_list = [e for e in priority.evidence if e.field_name == "days_inactive"]
            recs.append(FollowupRecommendationOutput(
                customer_id=ctx.customer_id,
                customer_name=ctx.customer_name,
                source_module="customers",
                source_record_id=ctx.customer_id,
                source_reference=source_ref,
                trigger_reason=f"Account has had zero operational freight activity for {ctx.days_inactive} days",
                category="CUSTOMER_FOLLOWUP",
                priority="MEDIUM",
                risk_level="MEDIUM",
                confidence="HIGH",
                confidence_score=0.88,
                evidence=ev_list,
                recommended_action="Schedule relationship check-in call with logistics coordinator to review upcoming shipping lanes.",
                action_type="SCHEDULE_CHECKIN",
                followup_type="GENERAL_CHECKIN",
                suggested_owner_id=ctx.account_owner_id,
                suggested_owner_name=ctx.account_owner_name,
                requires_approval=False,
                correlation_id=corr,
                dedup_key=dedup,
            ))

        # 5. Contract Expiry Follow-Up
        if ctx.expiring_contract_count > 0:
            dedup = self._compute_dedup(ctx.org_id, "customers", ctx.customer_id, "CUSTOMER_CONTRACT_EXPIRY_FOLLOWUP")
            ev_list = [e for e in priority.evidence if e.source_module == "contracts"]
            recs.append(FollowupRecommendationOutput(
                customer_id=ctx.customer_id,
                customer_name=ctx.customer_name,
                source_module="contracts",
                source_record_id=ctx.customer_id,
                source_reference=source_ref,
                trigger_reason=f"{ctx.expiring_contract_count} rate agreement(s) nearing expiration renewal window",
                category="CUSTOMER_FOLLOWUP",
                priority="HIGH",
                risk_level="MEDIUM",
                confidence="VERY_HIGH",
                confidence_score=0.95,
                evidence=ev_list,
                recommended_action="Prepare commercial rate review and initiate renewal discussions with procurement.",
                action_type="CONTRACT_RENEWAL_REVIEW",
                followup_type="EXPIRING_CONTRACT",
                suggested_owner_id=ctx.account_owner_id,
                suggested_owner_name=ctx.account_owner_name,
                requires_approval=True,
                correlation_id=corr,
                dedup_key=dedup,
            ))

        return recs

    def draft_communication(self, req: DraftMessageRequest) -> DraftMessageResponse:
        """
        Drafts grounded customer communication ONLY when explicitly requested.
        Adheres strictly to supplied context with ZERO hallucinated values.
        """
        missing_info: List[str] = []
        warnings: List[SafetyWarning] = []
        recipients: List[str] = []

        if req.primary_contact_email:
            recipients.append(req.primary_contact_email)
        else:
            missing_info.append("Primary contact email address is not recorded on file")
            warnings.append(SafetyWarning(
                warning_type="MISSING_RECIPIENT_EMAIL",
                message="No verified recipient email on customer file. An operator must confirm the email address before sending.",
                severity="WARNING",
            ))

        salutation = f"Dear {req.primary_contact_name}," if req.primary_contact_name else f"Dear {req.customer_name} Team,"
        operator = req.operator_name or "LogisticsHQ Customer Success"
        
        # Build evidence text block
        evidence_lines = [f"• {ev.field_name}: {ev.observed_value} ({ev.description})" for ev in req.evidence]
        evidence_block = "\n".join(evidence_lines) if evidence_lines else "• Verified account record details on file."

        dt = req.draft_type.upper()
        if "INVOICE" in dt or "PAYMENT" in dt:
            subject = f"LogisticsHQ Statement Follow-Up: Account {req.customer_name} ({req.source_reference})"
            body = (
                f"{salutation}\n\n"
                f"I am reaching out from LogisticsHQ accounts receivable regarding your logistics account statement.\n\n"
                f"Account Condition:\n"
                f"{req.description}\n\n"
                f"Verified Record Details:\n"
                f"{evidence_block}\n\n"
                f"Action Requested:\n"
                f"{req.recommended_action}\n\n"
                f"If payment has already been processed, please forward the remittance advice so our finance team can update your ledger.\n\n"
                f"Sincerely,\n"
                f"{operator}\n"
                f"LogisticsHQ Financial Operations"
            )
        elif "DELAY" in dt or "SHIPMENT" in dt or "EXCEPTION" in dt:
            subject = f"Shipment Status Advisory: Consignment {req.source_reference} ({req.customer_name})"
            body = (
                f"{salutation}\n\n"
                f"We are providing an operational update regarding shipment {req.source_reference}.\n\n"
                f"Status Summary:\n"
                f"{req.description}\n\n"
                f"Verified Milestone Details:\n"
                f"{evidence_block}\n\n"
                f"Operational Next Steps:\n"
                f"{req.recommended_action}\n\n"
                f"Our operations control tower is actively monitoring this movement with the carrier. "
                f"We will provide further milestone confirmations as transit progresses.\n\n"
                f"Sincerely,\n"
                f"{operator}\n"
                f"LogisticsHQ Freight Operations Desk"
            )
            warnings.append(SafetyWarning(
                warning_type="UNCONFIRMED_FINAL_DELIVERY",
                message="Draft does not state an unverified ETA commitment. Verified milestones only.",
                severity="INFO",
            ))
        elif "QUOTE" in dt or "RFQ" in dt:
            subject = f"Quotation Follow-Up: Freight Inquiry {req.source_reference} for {req.customer_name}"
            body = (
                f"{salutation}\n\n"
                f"I hope your week is going well. I am following up on the freight quotation options prepared for reference {req.source_reference}.\n\n"
                f"Inquiry Overview:\n"
                f"{req.description}\n\n"
                f"Details on File:\n"
                f"{evidence_block}\n\n"
                f"Next Step:\n"
                f"{req.recommended_action}\n\n"
                f"Please let us know if you would like to adjust container specifications, transit schedules, or move forward with booking confirmation.\n\n"
                f"Sincerely,\n"
                f"{operator}\n"
                f"LogisticsHQ Commercial Desk"
            )
        elif "CONTRACT" in dt or "RENEWAL" in dt:
            subject = f"Commercial Agreement Review: {req.source_reference} ({req.customer_name})"
            body = (
                f"{salutation}\n\n"
                f"As part of our periodic commercial relationship review, we noted that agreement {req.source_reference} is approaching its renewal window.\n\n"
                f"Agreement Overview:\n"
                f"{req.description}\n\n"
                f"Current Record Observations:\n"
                f"{evidence_block}\n\n"
                f"Recommended Next Step:\n"
                f"{req.recommended_action}\n\n"
                f"We would welcome the opportunity to review trade lane volume commitments and rate structures for the upcoming period.\n\n"
                f"Sincerely,\n"
                f"{operator}\n"
                f"LogisticsHQ Commercial Management"
            )
        elif "DOCUMENT" in dt:
            subject = f"Action Required: Transport Documentation for {req.source_reference}"
            body = (
                f"{salutation}\n\n"
                f"Regarding active freight reference {req.source_reference}:\n\n"
                f"To ensure customs clearance and carrier compliance, our team requires the following documentation:\n\n"
                f"{evidence_block}\n\n"
                f"Recommended Action:\n"
                f"{req.recommended_action}\n\n"
                f"Please upload or reply with these documents at your earliest convenience to prevent cargo dwell or carrier holds.\n\n"
                f"Sincerely,\n"
                f"{operator}\n"
                f"LogisticsHQ Documentation Desk"
            )
        elif "INTERNAL" in dt or "NOTE" in dt:
            subject = f"Internal Note: Follow-Up Assessment for {req.customer_name} ({req.source_reference})"
            body = (
                f"LogisticsHQ Internal Customer Follow-Up Note\n\n"
                f"Account: {req.customer_name} (Reference: {req.source_reference})\n"
                f"Condition: {req.description}\n\n"
                f"Verified Observations:\n"
                f"{evidence_block}\n\n"
                f"Recommended Internal Action:\n"
                f"{req.recommended_action}\n\n"
                f"Logged by: {operator}"
            )
            # Internal notes do not need external recipients
            recipients = []
        else:  # General Check-in
            subject = f"LogisticsHQ Relationship Check-In: {req.customer_name}"
            body = (
                f"{salutation}\n\n"
                f"I am checking in on behalf of LogisticsHQ regarding your ongoing freight forwarding and supply chain requirements.\n\n"
                f"Account Status:\n"
                f"{req.description}\n\n"
                f"Recommended Next Step:\n"
                f"{req.recommended_action}\n\n"
                f"Please let us know if you have upcoming shipping inquiries, lane reviews, or rate inquiries where our team can assist.\n\n"
                f"Sincerely,\n"
                f"{operator}\n"
                f"LogisticsHQ Customer Success"
            )

        confidence = 0.95 if req.evidence else 0.82

        return DraftMessageResponse(
            status="SUCCESS",
            subject=subject,
            body=body,
            suggested_recipients=recipients,
            evidence=req.evidence,
            missing_information=missing_info,
            safety_warnings=warnings,
            confidence=confidence,
            requires_approval=True,  # Mandatory human approval gate for external communication
            correlation_id=req.correlation_id,
        )

    def _compute_dedup(self, org_id: int, source_type: str, source_id: int, rule_name: str) -> str:
        raw = f"{org_id}:{source_type}:{source_id}:{rule_name}"
        return hashlib.sha256(raw.encode("utf-8")).hexdigest()[:32]
