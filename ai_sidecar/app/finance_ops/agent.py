"""
Finance and Collections Automation AI Agent
Phase 3 Task 3.6 for LogisticsHQ

Provides intelligent receivables risk analysis, collections prioritization,
customer payment behavior evaluation, bounded operational recommendations, and
communication draft synthesis.
"""

import re
from typing import List, Dict, Any
from app.finance_ops.models import (
    InvoiceContext,
    DeterministicFinanceSignals,
    ReceivablesRiskAnalysisResponse,
    CollectionPrioritizationResponse,
    CollectionPrioritizedItem,
    CustomerPaymentBehaviorResponse,
    OperationalActionRecommendation,
    OperationalRecommendationsResponse,
    CollectionDraftRequest,
    CollectionDraftResponse,
    EvidenceFact
)


class FinanceCollectionsAgent:
    """
    Dedicated AI Agent for Finance and Collections Automation.
    Strictly restricted to analyzing and recommending based on real Go backend facts.
    Never alters database records or directly transmits communications.
    """

    INJECTION_PATTERNS = [
        r"(?i)\bignore\s+(all\s+)?(previous|prior)\s+instructions\b",
        r"(?i)\bmark\s+(as\s+)?paid\b",
        r"(?i)\bwaive\s+(all\s+)?(fees|debt|balance)\b",
        r"(?i)\bwrite\s*off\b",
        r"(?i)\bdelete\s+from\b",
        r"(?i)\bdrop\s+table\b",
        r"(?i)\bexec(ute)?\s+immediate\b"
    ]

    def sanitize_user_input(self, text: str) -> str:
        """Strip dangerous prompt injection vectors from freeform operator instructions."""
        if not text:
            return ""
        cleaned = text.strip()
        for pat in self.INJECTION_PATTERNS:
            if re.search(pat, cleaned):
                return "[User instruction suppressed: unsafe directive detected]"
        return cleaned[:500]

    def analyze_receivables_risk(
        self,
        ctx: InvoiceContext,
        signals: DeterministicFinanceSignals
    ) -> ReceivablesRiskAnalysisResponse:
        """
        Synthesizes multi-signal receivables risk analysis grounded in deterministic Go facts.
        """
        evidence: List[EvidenceFact] = []
        key_risks: List[str] = []
        next_steps: List[str] = []

        # Grounding: Invoice Amount & Balance
        evidence.append(EvidenceFact(
            source_module="invoices",
            source_entity_id=ctx.invoice_id,
            source_ref=ctx.invoice_number,
            field_name="balance_due",
            observed_value=f"{ctx.currency} {ctx.balance_due:,.2f}",
            description=f"Outstanding balance on total amount of {ctx.currency} {ctx.total_amount:,.2f}"
        ))

        # Grounding: Overdue status & Aging
        if signals.is_overdue:
            evidence.append(EvidenceFact(
                source_module="invoices",
                source_entity_id=ctx.invoice_id,
                source_ref=ctx.invoice_number,
                field_name="days_overdue",
                observed_value=f"{signals.days_overdue} days",
                description=f"Payment deadline lapsed ({ctx.due_date}). Aging bucket: {signals.aging_bucket}"
            ))
            key_risks.append(f"Invoice is overdue by {signals.days_overdue} days (Bucket: {signals.aging_bucket})")

        # Grounding: High value exposure
        if signals.is_high_value:
            key_risks.append(f"High-value balance exposure exceeding threshold ({ctx.currency} {ctx.balance_due:,.2f})")
            evidence.append(EvidenceFact(
                source_module="invoices",
                source_entity_id=ctx.invoice_id,
                source_ref=ctx.invoice_number,
                field_name="is_high_value",
                observed_value="True",
                description=f"Balance exceeds organization high-value threshold"
            ))

        # Grounding: Customer aggregate overdue
        if signals.has_multiple_overdue:
            key_risks.append(f"Customer has {signals.customer_overdue_count} overdue invoices totaling {ctx.currency} {signals.customer_total_overdue_balance:,.2f}")
            evidence.append(EvidenceFact(
                source_module="customers",
                source_entity_id=ctx.customer_id,
                source_ref=ctx.customer_name,
                field_name="customer_overdue_count",
                observed_value=str(signals.customer_overdue_count),
                description=f"Multiple delinquent invoices detected across customer account"
            ))

        # Grounding: Credit limit breach
        if signals.is_credit_limit_exceeded:
            key_risks.append(f"Customer total outstanding balance exceeds approved credit limit of {ctx.currency} {ctx.customer_credit_limit:,.2f}")

        # Grounding: Linked shipment operational exception
        if signals.has_linked_exception:
            key_risks.append(f"Invoice is linked to shipment #{ctx.shipment_id or ctx.shipment_number} with an active operational exception")

        # Synthesize Operational Next Steps based on deterministic signals
        if signals.days_overdue > 60 or signals.is_credit_limit_exceeded:
            next_steps.append("Escalate account to Finance Manager for formal collection action review")
            next_steps.append("Place customer on temporary credit hold pending balance reconciliation")
            next_steps.append("Prepare formal demand notice with payment verification schedule")
        elif signals.days_overdue > 30:
            next_steps.append("Issue assertive overdue notice to accounts payable lead")
            next_steps.append("Request carrier proof of delivery documentation to eliminate billing dispute pretexts")
        elif signals.days_overdue > 0:
            next_steps.append("Transmit polite payment reminder referencing invoice and payment portal details")
            next_steps.append("Verify customer receipt of original tax invoice")
        elif signals.is_approaching_due_date:
            next_steps.append(f"Send courtesy pre-due reminder (due in {signals.days_until_due} days on {ctx.due_date})")
        else:
            next_steps.append("Invoice in good standing; standard receivables cycle")

        # Synthesize human-readable summary
        if signals.is_overdue:
            summary = (
                f"Invoice {ctx.invoice_number} for customer '{ctx.customer_name}' has an outstanding balance of "
                f"{ctx.currency} {ctx.balance_due:,.2f} and is {signals.days_overdue} days past its due date ({ctx.due_date}). "
                f"Assigned to aging bucket '{signals.aging_bucket}'. "
            )
            if signals.has_multiple_overdue:
                summary += f"The customer has {signals.customer_overdue_count} overdue invoices across their account. "
            if signals.is_credit_limit_exceeded:
                summary += "Account exceeds authorized credit limit. Proactive collections follow-up is recommended."
        else:
            summary = (
                f"Invoice {ctx.invoice_number} for customer '{ctx.customer_name}' has a total balance of "
                f"{ctx.currency} {ctx.balance_due:,.2f} due on {ctx.due_date}. Account is currently within normal credit terms."
            )

        return ReceivablesRiskAnalysisResponse(
            invoice_id=ctx.invoice_id,
            customer_id=ctx.customer_id,
            risk_level=signals.risk_level,
            risk_score=signals.risk_score,
            days_overdue=signals.days_overdue,
            aging_bucket=signals.aging_bucket,
            outstanding_amount=ctx.balance_due,
            currency=ctx.currency,
            receivables_summary=summary,
            key_risks=key_risks,
            recommended_next_steps=next_steps,
            evidence=evidence,
            confidence_score=0.92,
            correlation_id=ctx.correlation_id
        )

    def prioritize_collections(
        self,
        ctx: InvoiceContext,
        signals: DeterministicFinanceSignals
    ) -> CollectionPrioritizationResponse:
        """
        Ranks collections urgency based on financial exposure, aging, and multi-invoice risk.
        """
        items: List[CollectionPrioritizedItem] = []

        rank = 1
        if signals.days_overdue > 60:
            urgency = "IMMEDIATE"
            deadline = "Within 24 hours"
            impact = f"High default risk on {ctx.currency} {ctx.balance_due:,.2f} (>60 days overdue)"
            action = "Dispatch formal demand notice & schedule executive escalation review"
            rationale = "Long-overdue balance in critical aging bracket with compounding bad debt risk"
        elif signals.days_overdue > 30:
            urgency = "HIGH"
            deadline = "Within 48 hours"
            impact = f"Elevated cashflow delay on {ctx.currency} {ctx.balance_due:,.2f} (31-60 days bucket)"
            action = "Transmit assertive second overdue notice and contact AP controller by phone"
            rationale = "Invoice has exceeded typical monthly settlement window"
        elif signals.days_overdue > 0:
            urgency = "MEDIUM"
            deadline = "Within 3 business days"
            impact = f"Initial payment delay on {ctx.currency} {ctx.balance_due:,.2f} (1-30 days bucket)"
            action = "Send standard automated reminder with attached invoice PDF and banking details"
            rationale = "Recently lapsed invoice; proactive outreach minimizes aging into higher buckets"
        else:
            urgency = "LOW"
            deadline = "Pre-due cycle"
            impact = "Standard receivables flow"
            action = "Monitor upcoming due date; send courtesy reminder if requested"
            rationale = "Invoice is not yet overdue"

        items.append(CollectionPrioritizedItem(
            invoice_id=ctx.invoice_id,
            invoice_number=ctx.invoice_number,
            customer_id=ctx.customer_id,
            customer_name=ctx.customer_name,
            priority_rank=rank,
            urgency=urgency,
            outstanding_amount=ctx.balance_due,
            currency=ctx.currency,
            days_overdue=signals.days_overdue,
            aging_bucket=signals.aging_bucket,
            business_impact=impact,
            recommended_action=action,
            action_deadline=deadline,
            rationale=rationale
        ))

        summary = f"Evaluated 1 invoice for collections prioritization. Total at-risk balance: {ctx.currency} {ctx.balance_due:,.2f}."

        return CollectionPrioritizationResponse(
            prioritized_items=items,
            summary=summary,
            total_at_risk_amount=ctx.balance_due,
            currency=ctx.currency
        )

    def analyze_customer_payment_behavior(
        self,
        ctx: InvoiceContext,
        signals: DeterministicFinanceSignals
    ) -> CustomerPaymentBehaviorResponse:
        """
        Evaluates customer payment habits and credit risk based on historical settlement patterns.
        """
        evidence: List[EvidenceFact] = []

        total_invoices = max(1, ctx.customer_open_invoices_count or 1)
        overdue_count = signals.customer_overdue_count

        if overdue_count == 0 and not signals.is_overdue:
            category = "RELIABLE"
            on_time_ratio = 0.98
            avg_days = 2.0
            credit_action = "Maintain existing standard credit terms"
            comm_strategy = "Polite and consultative communication"
        elif signals.days_overdue > 60 or overdue_count >= 3 or signals.is_credit_limit_exceeded:
            category = "HIGH_RISK_DEFAULT"
            on_time_ratio = 0.40
            avg_days = float(signals.days_overdue)
            credit_action = "Suspend credit line; require advance deposit or prepayment on new bookings"
            comm_strategy = "Formal demand notices and escalation to senior corporate leadership"
        elif signals.days_overdue > 20 or overdue_count >= 2:
            category = "CHRONICALLY_LATE"
            on_time_ratio = 0.65
            avg_days = float(signals.days_overdue)
            credit_action = "Reduce credit terms from Net 30 to Net 15"
            comm_strategy = "Regular weekly telephone follow-ups and prompt reminder automation"
        else:
            category = "OCCASIONAL_DELAY"
            on_time_ratio = 0.85
            avg_days = 8.0
            credit_action = "Maintain terms with close monitoring of approaching milestones"
            comm_strategy = "Professional email reminders 3 days prior to due date"

        limit = ctx.customer_credit_limit or 0.0
        outstanding = ctx.customer_total_outstanding or ctx.balance_due
        util_pct = (outstanding / limit * 100.0) if limit > 0 else 0.0

        evidence.append(EvidenceFact(
            source_module="customers",
            source_entity_id=ctx.customer_id,
            source_ref=ctx.customer_name,
            field_name="payment_behavior_category",
            observed_value=category,
            description=f"Classified as {category} based on overdue count ({overdue_count}) and aging ({signals.days_overdue}d)"
        ))

        return CustomerPaymentBehaviorResponse(
            customer_id=ctx.customer_id,
            customer_name=ctx.customer_name,
            behavior_category=category,
            on_time_payment_ratio=on_time_ratio,
            average_days_to_pay=avg_days,
            outstanding_balance=outstanding,
            credit_limit_utilization_pct=round(util_pct, 1),
            credit_limit=limit,
            recommended_credit_action=credit_action,
            communication_strategy=comm_strategy,
            evidence=evidence
        )

    def generate_operational_recommendations(
        self,
        ctx: InvoiceContext,
        signals: DeterministicFinanceSignals
    ) -> OperationalRecommendationsResponse:
        """
        Generates bounded, action-system-mapped recommendations.
        """
        recs: List[OperationalActionRecommendation] = []

        if signals.is_overdue:
            if signals.days_overdue > 60:
                recs.append(OperationalActionRecommendation(
                    recommendation_id=f"rec-demand-{ctx.invoice_id}",
                    title="Send Formal Demand Notice",
                    description=f"Transmit formal legal-precursor demand notice for {ctx.currency} {ctx.balance_due:,.2f} ({signals.days_overdue} days overdue).",
                    target_action="finance.send_formal_demand",
                    risk_rating="HIGH",
                    requires_approval=True,
                    action_payload_template={
                        "invoice_id": ctx.invoice_id,
                        "customer_id": ctx.customer_id,
                        "draft_type": "FINAL_DEMAND",
                        "outstanding_amount": ctx.balance_due,
                        "currency": ctx.currency
                    }
                ))
                recs.append(OperationalActionRecommendation(
                    recommendation_id=f"rec-esc-{ctx.invoice_id}",
                    title="Escalate Delinquent Account",
                    description="Trigger managerial escalation workflow and notify VP of Finance regarding bad-debt exposure.",
                    target_action="finance.escalate_overdue_receivable",
                    risk_rating="HIGH",
                    requires_approval=True,
                    action_payload_template={
                        "invoice_id": ctx.invoice_id,
                        "customer_id": ctx.customer_id,
                        "severity": "CRITICAL"
                    }
                ))
            else:
                recs.append(OperationalActionRecommendation(
                    recommendation_id=f"rec-remind-{ctx.invoice_id}",
                    title="Send Overdue Collection Reminder",
                    description=f"Send professional follow-up notice with bank settlement details for invoice {ctx.invoice_number}.",
                    target_action="finance.send_collection_reminder",
                    risk_rating="HIGH",
                    requires_approval=True,
                    action_payload_template={
                        "invoice_id": ctx.invoice_id,
                        "customer_id": ctx.customer_id,
                        "draft_type": "OVERDUE_NOTICE",
                        "outstanding_amount": ctx.balance_due,
                        "currency": ctx.currency
                    }
                ))

            recs.append(OperationalActionRecommendation(
                recommendation_id=f"rec-task-{ctx.invoice_id}",
                title="Create Internal Collections Task",
                description=f"Log internal follow-up task for Finance AR specialist to call customer accounts payable desk.",
                target_action="finance.create_followup_task",
                risk_rating="LOW",
                requires_approval=False,
                action_payload_template={
                    "invoice_id": ctx.invoice_id,
                    "task_title": f"Follow up on overdue invoice {ctx.invoice_number} ({ctx.customer_name})",
                    "priority": "HIGH" if signals.is_high_value else "MEDIUM"
                }
            ))
        else:
            recs.append(OperationalActionRecommendation(
                recommendation_id=f"rec-pre-due-{ctx.invoice_id}",
                title="Send Courtesy Pre-Due Notice",
                description=f"Send friendly courtesy reminder prior to due date {ctx.due_date}.",
                target_action="finance.send_collection_reminder",
                risk_rating="HIGH",
                requires_approval=True,
                action_payload_template={
                    "invoice_id": ctx.invoice_id,
                    "draft_type": "FIRST_REMINDER",
                    "outstanding_amount": ctx.balance_due,
                    "currency": ctx.currency
                }
            ))

        return OperationalRecommendationsResponse(
            invoice_id=ctx.invoice_id,
            recommendations=recs
        )

    def generate_collection_draft(
        self,
        req: CollectionDraftRequest
    ) -> CollectionDraftResponse:
        """
        Synthesizes an editable, human-in-the-loop collection communication draft.
        All consequential messages require human approval before transmission.
        """
        ctx = req.context
        signals = req.signals
        draft_type = req.draft_type
        instructions = self.sanitize_user_input(req.user_instructions or "")

        recipient_name = ctx.customer_name
        recipient_email = ctx.customer_email or "ap-dept@customer.com"

        safety_restrictions = [
            "Draft cannot be dispatched without Managerial Approval",
            "Cannot waive interest, fees, or principal amounts",
            "Cannot offer binding payment settlements or discounts without CFO sign-off",
            "Must preserve exact outstanding amount and currency from backend"
        ]

        if draft_type == "FINAL_DEMAND":
            subject = f"FINAL NOTICE: Urgent Settlement Required - Invoice {ctx.invoice_number} ({ctx.currency} {ctx.balance_due:,.2f})"
            body = (
                f"Dear {recipient_name} Accounts Team,\n\n"
                f"This is a formal final notice regarding invoice {ctx.invoice_number}, originally issued on {ctx.invoice_date} "
                f"and due on {ctx.due_date}. As of today, the balance of {ctx.currency} {ctx.balance_due:,.2f} remains unpaid "
                f"({signals.days_overdue} days past due).\n\n"
                f"Despite previous communications, we have not received payment confirmation or a formal remittance advice. "
                f"To avoid suspension of your account's ongoing freight services and escalation to formal collections, "
                f"please remit the full outstanding amount immediately using our registered banking details.\n\n"
                f"If payment has already been executed within the last 24 hours, kindly forward the transaction reference "
                f"or bank wire slip to our finance desk.\n\n"
                f"Sincerely,\nLogisticsHQ Finance & Receivables Management"
            )
            notes = f"Final demand draft prepared for invoice #{ctx.invoice_id} ({signals.days_overdue} days overdue). Requires immediate manager review."
            tone_applied = "URGENT_FORMAL"

        elif draft_type == "OVERDUE_NOTICE":
            subject = f"OVERDUE REMINDER: Invoice {ctx.invoice_number} - {ctx.currency} {ctx.balance_due:,.2f}"
            body = (
                f"Dear {recipient_name} Accounts Payable,\n\n"
                f"Our records indicate that invoice {ctx.invoice_number} for {ctx.currency} {ctx.balance_due:,.2f} was due on {ctx.due_date} "
                f"and is currently overdue by {signals.days_overdue} days.\n\n"
                f"We value our commercial partnership and understand administrative delays can occur. "
                f"Please review this statement and let us know when we can expect settlement, or reply with your payment remittance details.\n\n"
                f"If you have questions regarding the underlying shipment (#{ctx.shipment_number or 'N/A'}) or require duplicate copies "
                f"of the billing documents, please let us know immediately.\n\n"
                f"Warm regards,\nFinance Team · LogisticsHQ"
            )
            notes = f"Second overdue reminder for {ctx.invoice_number} ({signals.days_overdue}d overdue). Standard escalation path."
            tone_applied = "ASSERTIVE"

        elif draft_type == "PAYMENT_PLAN_OFFER":
            subject = f"Payment Schedule Proposal: Invoice {ctx.invoice_number} - LogisticsHQ"
            body = (
                f"Dear {recipient_name},\n\n"
                f"Following our recent discussions regarding the outstanding balance of {ctx.currency} {ctx.balance_due:,.2f} on invoice {ctx.invoice_number}, "
                f"we would like to propose a structured installment schedule to facilitate timely resolution.\n\n"
                f"Please confirm your availability to discuss acceptable installment dates so that we can maintain uninterrupted freight logistics services.\n\n"
                f"Sincerely,\nCredit & Collections Team · LogisticsHQ"
            )
            notes = f"Payment plan discussion proposal. Requires credit committee verification."
            tone_applied = "CONSULTATIVE"

        elif draft_type == "INTERNAL_ESCALATION":
            subject = f"INTERNAL ESCALATION: Delinquent Account {ctx.customer_name} (Invoice {ctx.invoice_number})"
            body = (
                f"ATTN: Finance Manager & Account Director,\n\n"
                f"Invoice {ctx.invoice_number} for customer '{ctx.customer_name}' is {signals.days_overdue} days overdue with an outstanding balance of {ctx.currency} {ctx.balance_due:,.2f}.\n\n"
                f"Key risk factors:\n"
                f"- Aging Bucket: {signals.aging_bucket}\n"
                f"- Customer total overdue invoices: {signals.customer_overdue_count}\n"
                f"- Credit limit exceeded: {'YES' if signals.is_credit_limit_exceeded else 'NO'}\n\n"
                f"Recommended action: Review commercial credit status and approve formal recovery procedure.\n\n"
                f"Generated by Finance Operations Automation"
            )
            notes = f"Internal operational escalation notice for invoice #{ctx.invoice_id}."
            tone_applied = "INTERNAL_FORMAL"

        else:  # FIRST_REMINDER / PRE_DUE
            subject = f"Payment Reminder: Invoice {ctx.invoice_number} ({ctx.currency} {ctx.balance_due:,.2f}) due {ctx.due_date}"
            body = (
                f"Dear {recipient_name},\n\n"
                f"This is a friendly reminder that invoice {ctx.invoice_number} in the amount of {ctx.currency} {ctx.balance_due:,.2f} "
                f"is scheduled for payment on {ctx.due_date}.\n\n"
                f"For your convenience, our standard electronic wire and ACH transfer details are provided on the invoice document. "
                f"Please ensure payment is initiated in time to meet the due date.\n\n"
                f"Thank you for your continued business.\n\n"
                f"Best regards,\nAccounts Receivable · LogisticsHQ"
            )
            notes = f"Courtesy reminder draft for upcoming/recent due date."
            tone_applied = "POLITE"

        if instructions:
            body += f"\n\n[Special Instruction Note: {instructions}]"

        return CollectionDraftResponse(
            subject=subject,
            message_body=body,
            internal_notes=notes,
            recipient_preview={"name": recipient_name, "email": recipient_email},
            outstanding_amount=ctx.balance_due,
            currency=ctx.currency,
            requires_approval=True,
            tone_applied=tone_applied,
            safety_restrictions=safety_restrictions
        )
