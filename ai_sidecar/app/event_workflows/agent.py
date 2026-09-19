"""
AI Agent implementation for Phase 3 Task 3.8: Event-Driven AI Workflows and Cross-Module Automation.
ALL AI CODE IS WRITTEN IN PYTHON ONLY.
"""

import os
import re
from typing import Dict, Any, List, Optional
from .models import (
    EventWorkflowRequest,
    EventAnalysisResponse,
    WorkflowRecommendation,
    WorkflowDraftRequest,
    WorkflowDraftResponse,
)


class EventWorkflowsAgent:
    """
    EventWorkflowsAgent evaluates domain events across LogisticsHQ modules (Shipments,
    Finance, Contracts, RFQs, Leads, Bookings), synthesizes cross-module business context,
    and produces grounded recommendations mapped to the Centralized Action System.
    """

    def __init__(self):
        self.gemini_api_key = os.getenv("GEMINI_API_KEY", "")

    def _sanitize(self, text: str) -> str:
        if not text:
            return ""
        forbidden = [
            r"<<<.*?>>>",
            r"system\s*prompt",
            r"ignore\s*previous\s*instructions",
            r"you\s*are\s*now\s*a",
            r"drop\s*table",
            r"delete\s*from",
        ]
        sanitized = text
        for pattern in forbidden:
            sanitized = re.sub(pattern, "[FILTERED]", sanitized, flags=re.IGNORECASE)
        return sanitized.strip()

    def analyze_event(self, req: EventWorkflowRequest) -> EventAnalysisResponse:
        """
        Analyzes an incoming domain event, synthesizes cross-module dependencies,
        and generates prioritized recommendations.
        """
        event_type = req.event_type.lower()
        facts = req.event_facts or {}
        cross_ctx = req.cross_module_context or {}

        # 1. Classify Workflow Type & Default Urgency
        if "shipment" in event_type or "milestone" in event_type or "exception" in event_type:
            workflow_type = "SHIPMENT_EXCEPTION_RESPONSE"
            urgency = "HIGH" if facts.get("severity") in ("CRITICAL", "HIGH") or facts.get("delay_hours", 0) > 24 else "MEDIUM"
            significance = 85.0 if urgency == "HIGH" else 60.0
            rec_actions = self._generate_shipment_recommendations(req, facts, cross_ctx)
            insights = self._generate_shipment_insights(facts, cross_ctx)
            summary = (
                f"Shipment exception event '{req.event_type}' detected for record {req.source_record_id}. "
                f"Operational impact assessed with {len(cross_ctx.get('active_shipments', []))} active related consignments "
                f"and customer tier '{cross_ctx.get('customer_tier', 'STANDARD')}'."
            )

        elif "invoice" in event_type or "payment" in event_type or "receivable" in event_type:
            workflow_type = "INVOICE_COLLECTION_ESCALATION"
            days_overdue = facts.get("days_overdue", 0)
            urgency = "CRITICAL" if days_overdue > 30 or facts.get("is_high_value") else "MEDIUM"
            significance = 90.0 if urgency == "CRITICAL" else 65.0
            rec_actions = self._generate_finance_recommendations(req, facts, cross_ctx)
            insights = self._generate_finance_insights(facts, cross_ctx)
            summary = (
                f"Receivables event '{req.event_type}' triggered for Invoice #{req.source_record_id}. "
                f"Outstanding balance: ${facts.get('outstanding_amount', 0):,.2f} ({days_overdue} days aging). "
                f"Customer total exposure: ${cross_ctx.get('customer_total_outstanding', 0):,.2f}."
            )

        elif "contract" in event_type or "compliance" in event_type or "document" in event_type:
            workflow_type = "CONTRACT_COMPLIANCE_RENEWAL"
            days_until_expiry = facts.get("days_until_expiry", 999)
            urgency = "HIGH" if days_until_expiry <= 30 or facts.get("missing_documents") else "MEDIUM"
            significance = 80.0 if urgency == "HIGH" else 50.0
            rec_actions = self._generate_contract_recommendations(req, facts, cross_ctx)
            insights = self._generate_contract_insights(facts, cross_ctx)
            summary = (
                f"Contract/compliance event '{req.event_type}' detected for agreement {req.source_record_id}. "
                f"Agreement horizon: {days_until_expiry} days to expiry. "
                f"Counterparty: {cross_ctx.get('party_name', 'Commercial Partner')}."
            )

        elif "rfq" in event_type or "quote" in event_type or "pricing" in event_type:
            workflow_type = "RFQ_QUOTATION_DISPATCH"
            urgency = "HIGH" if facts.get("deadline_approaching") else "MEDIUM"
            significance = 75.0
            rec_actions = self._generate_rfq_recommendations(req, facts, cross_ctx)
            insights = self._generate_rfq_insights(facts, cross_ctx)
            summary = (
                f"RFQ/Pricing lifecycle event '{req.event_type}' evaluated for RFQ #{req.source_record_id}. "
                f"Route: {facts.get('origin', 'ORIG')} -> {facts.get('destination', 'DEST')}. "
                f"Margin target: {cross_ctx.get('target_margin_pct', 15.0)}%."
            )

        elif "lead" in event_type:
            workflow_type = "LEAD_FOLLOWUP"
            urgency = "MEDIUM" if facts.get("score", 0) > 70 else "LOW"
            significance = 60.0
            rec_actions = self._generate_lead_recommendations(req, facts, cross_ctx)
            insights = [
                f"Lead qualification score: {facts.get('score', 50)}/100",
                f"Inquiry channel: {facts.get('channel', 'INBOUND_EMAIL')}",
                f"Estimated annual freight spend: ${cross_ctx.get('estimated_annual_spend', 0):,.2f}",
            ]
            summary = (
                f"Lead event '{req.event_type}' processed for Lead #{req.source_record_id}. "
                f"Engagement potential evaluated based on qualification profile."
            )

        else:
            workflow_type = "GENERAL_CROSS_MODULE_WORKFLOW"
            urgency = "LOW"
            significance = 40.0
            rec_actions = [
                WorkflowRecommendation(
                    recommendation_type="GENERAL_AUDIT",
                    title="Audit Event Notification",
                    description=f"Record cross-module event {req.event_type} in operations log.",
                    action_name="notifications.create",
                    action_payload={"title": f"Event {req.event_type}", "message": f"Processed event for {req.source_record_type} #{req.source_record_id}"},
                    requires_approval=False,
                    risk_level="LOW",
                    target_module="NOTIFICATIONS",
                )
            ]
            insights = [f"Event received with {len(facts)} payload facts."]
            summary = f"General domain event '{req.event_type}' recorded and evaluated."

        # Missing information checks
        missing_info = []
        if not facts.get("contact_email") and not cross_ctx.get("contact_email"):
            missing_info.append("Counterparty direct notification email address")
        if "shipment" in event_type and not facts.get("carrier_name"):
            missing_info.append("Assigned line-haul carrier confirmation")

        return EventAnalysisResponse(
            workflow_type=workflow_type,
            urgency=urgency,
            significance_score=significance,
            summary=summary,
            cross_module_insights=insights,
            recommended_actions=rec_actions,
            missing_information=missing_info,
            confidence_score=0.94,
            correlation_id=req.correlation_id,
        )

    def _generate_shipment_recommendations(self, req: EventWorkflowRequest, facts: Dict[str, Any], cross_ctx: Dict[str, Any]) -> List[WorkflowRecommendation]:
        recs = []
        shipment_id = int(req.source_record_id) if req.source_record_id.isdigit() else 1
        # Safe internal notification
        recs.append(
            WorkflowRecommendation(
                recommendation_type="OPERATIONS_INTERNAL_TASK",
                title=f"Exception Review Task for Shipment #{shipment_id}",
                description="Assign operational mitigation task to on-duty freight coordinator.",
                action_name="notifications.create",
                action_payload={
                    "title": f"Shipment #{shipment_id} Operational Exception",
                    "message": f"Delay or exception detected: {facts.get('exception_type', 'Milestone Delay')}. Immediate carrier check recommended.",
                    "severity": "WARNING",
                },
                requires_approval=False,
                risk_level="LOW",
                target_module="SHIPMENTS",
            )
        )
        # Consequential customer delay alert (Requires approval!)
        recs.append(
            WorkflowRecommendation(
                recommendation_type="EXTERNAL_COMMUNICATION",
                title=f"Dispatch Delay Notice for Shipment #{shipment_id}",
                description="Notify customer shipper of schedule deviation with revised ETA.",
                action_name="shipments.notify_delay",
                action_payload={
                    "shipment_id": shipment_id,
                    "revised_eta": facts.get("revised_eta", "TBD"),
                    "reason": facts.get("exception_type", "Port Congestion / Rail Delay"),
                },
                requires_approval=True,
                risk_level="HIGH",
                target_module="SHIPMENTS",
            )
        )
        return recs

    def _generate_shipment_insights(self, facts: Dict[str, Any], cross_ctx: Dict[str, Any]) -> List[str]:
        insights = []
        if facts.get("delay_hours", 0) > 0:
            insights.append(f"Cumulative schedule deviation: {facts.get('delay_hours')} hours past milestone window.")
        if cross_ctx.get("customer_tier") == "ENTERPRISE":
            insights.append("Counterparty is a Priority Enterprise Account: Strict SLA reporting applies.")
        if cross_ctx.get("active_shipments_count", 0) > 1:
            insights.append(f"Customer currently has {cross_ctx.get('active_shipments_count')} active consignments in transit.")
        return insights

    def _generate_finance_recommendations(self, req: EventWorkflowRequest, facts: Dict[str, Any], cross_ctx: Dict[str, Any]) -> List[WorkflowRecommendation]:
        recs = []
        inv_id = int(req.source_record_id) if req.source_record_id.isdigit() else 1
        days_overdue = facts.get("days_overdue", 0)

        # Internal review task
        recs.append(
            WorkflowRecommendation(
                recommendation_type="FINANCE_INTERNAL_REVIEW",
                title=f"Internal Receivables Review Task: Invoice #{inv_id}",
                description="Create finance review task to verify remittance advice or bank reconciliation.",
                action_name="finance.create_followup_task",
                action_payload={
                    "invoice_id": inv_id,
                    "task_title": f"Receivables Aging Follow-up for Invoice #{inv_id}",
                },
                requires_approval=False,
                risk_level="LOW",
                target_module="FINANCE",
            )
        )

        # External reminder (Requires approval!)
        if days_overdue > 15:
            recs.append(
                WorkflowRecommendation(
                    recommendation_type="FINANCE_ESCALATION",
                    title=f"Escalate Overdue Receivable: Invoice #{inv_id}",
                    description=f"Escalate collection notice for {days_overdue} days past due balance.",
                    action_name="finance.escalate_overdue_receivable",
                    action_payload={
                        "invoice_id": inv_id,
                        "escalation_level": "LEVEL_2",
                    },
                    requires_approval=True,
                    risk_level="HIGH",
                    target_module="FINANCE",
                )
            )
        else:
            recs.append(
                WorkflowRecommendation(
                    recommendation_type="FINANCE_REMINDER",
                    title=f"Send Collection Reminder: Invoice #{inv_id}",
                    description="Dispatch courteous payment reminder with wire remittance details.",
                    action_name="finance.send_collection_reminder",
                    action_payload={
                        "invoice_id": inv_id,
                        "reminder_type": "FRIENDLY_REMINDER",
                    },
                    requires_approval=True,
                    risk_level="HIGH",
                    target_module="FINANCE",
                )
            )
        return recs

    def _generate_finance_insights(self, facts: Dict[str, Any], cross_ctx: Dict[str, Any]) -> List[str]:
        insights = []
        if facts.get("days_overdue", 0) > 0:
            insights.append(f"Aging Bracket: {facts.get('days_overdue')} days past net payment terms.")
        if cross_ctx.get("customer_in_transit_value", 0) > 0:
            insights.append(f"Customer has ${cross_ctx.get('customer_in_transit_value', 0):,.2f} in active freight in transit.")
        if cross_ctx.get("credit_limit_exceeded"):
            insights.append("Customer commercial credit limit currently breached: Hold new booking approvals.")
        return insights

    def _generate_contract_recommendations(self, req: EventWorkflowRequest, facts: Dict[str, Any], cross_ctx: Dict[str, Any]) -> List[WorkflowRecommendation]:
        recs = []
        contract_id = int(req.source_record_id) if req.source_record_id.isdigit() else 1
        days_until_expiry = facts.get("days_until_expiry", 999)

        # Internal renewal task
        recs.append(
            WorkflowRecommendation(
                recommendation_type="CONTRACT_RENEWAL_TASK",
                title=f"Renewal Preparation Task: Contract #{contract_id}",
                description="Assign contract commercial renegotiation and rate extension task.",
                action_name="contracts.create_renewal_task",
                action_payload={
                    "contract_id": contract_id,
                    "task_title": f"Contract Renewal Review for Agreement #{contract_id}",
                },
                requires_approval=False,
                risk_level="LOW",
                target_module="CONTRACTS",
            )
        )

        # Missing doc request if compliance issue (Requires approval!)
        if facts.get("missing_documents") or facts.get("compliance_issue"):
            recs.append(
                WorkflowRecommendation(
                    recommendation_type="COMPLIANCE_DOCUMENT_REQUEST",
                    title=f"Request Missing Compliance Document: Contract #{contract_id}",
                    description="Request mandatory Certificate of Insurance or Customs POA.",
                    action_name="contracts.request_missing_document",
                    action_payload={
                        "contract_id": contract_id,
                        "subject": "ACTION REQUIRED: Missing Mandatory Compliance Documentation",
                    },
                    requires_approval=True,
                    risk_level="HIGH",
                    target_module="CONTRACTS",
                )
            )
        return recs

    def _generate_contract_insights(self, facts: Dict[str, Any], cross_ctx: Dict[str, Any]) -> List[str]:
        insights = []
        if facts.get("days_until_expiry", 999) <= 30:
            insights.append(f"Contract expires in {facts.get('days_until_expiry')} days: Renewal lead-time required.")
        if cross_ctx.get("active_shipments_under_contract", 0) > 0:
            insights.append(f"{cross_ctx.get('active_shipments_under_contract')} active freight shipments currently covered by this SLA.")
        return insights

    def _generate_rfq_recommendations(self, req: EventWorkflowRequest, facts: Dict[str, Any], cross_ctx: Dict[str, Any]) -> List[WorkflowRecommendation]:
        rfq_id = int(req.source_record_id) if req.source_record_id.isdigit() else 1
        return [
            WorkflowRecommendation(
                recommendation_type="RFQ_PRICING_REVIEW",
                title=f"Evaluate Carrier Sourcing for RFQ #{rfq_id}",
                description="Review contracted tariff rates and available spot carrier quotes.",
                action_name="notifications.create",
                action_payload={
                    "title": f"RFQ #{rfq_id} Pricing Ready",
                    "message": "AI pricing engine completed rate analysis. Ready for quotation review.",
                    "severity": "INFO",
                },
                requires_approval=False,
                risk_level="LOW",
                target_module="RFQ",
            )
        ]

    def _generate_rfq_insights(self, facts: Dict[str, Any], cross_ctx: Dict[str, Any]) -> List[str]:
        return [
            f"Mode: {facts.get('shipment_mode', 'OCEAN_FCL')} | Cargo: {facts.get('cargo_type', 'General Cargo')}",
            f"Historical lane win rate: {cross_ctx.get('lane_win_rate_pct', 35.0)}%",
        ]

    def _generate_lead_recommendations(self, req: EventWorkflowRequest, facts: Dict[str, Any], cross_ctx: Dict[str, Any]) -> List[WorkflowRecommendation]:
        lead_id = int(req.source_record_id) if req.source_record_id.isdigit() else 1
        return [
            WorkflowRecommendation(
                recommendation_type="LEAD_OUTREACH",
                title=f"Initiate Sales Follow-up for Lead #{lead_id}",
                description="Review qualification profile and schedule introductory commercial discovery.",
                action_name="notes.add",
                action_payload={
                    "entity_type": "LEAD",
                    "entity_id": lead_id,
                    "note": f"AI Event Workflow: High qualification lead ({facts.get('score', 75)}/100) queued for follow-up.",
                },
                requires_approval=False,
                risk_level="LOW",
                target_module="LEADS",
            )
        ]

    def generate_draft(self, req: WorkflowDraftRequest) -> WorkflowDraftResponse:
        """
        Synthesizes an editable draft message responding to a specific domain event.
        Drafts are explicitly marked as drafts and require human review.
        """
        event_type = req.event_type.lower()
        topic_clean = self._sanitize(req.topic)
        custom_clean = self._sanitize(req.custom_instructions or "")

        if "shipment" in event_type or "milestone" in event_type:
            subject = f"UPDATE: Shipment Schedule Advisory - Consignment #{req.source_record_id}"
            body = (
                f"Dear {req.recipient_name},\n\n"
                f"We are writing to inform you of an operational status update regarding Shipment #{req.source_record_id}.\n\n"
                f"Event Summary: {topic_clean or 'Schedule revision due to port terminal processing'}.\n"
                f"Current Status: Freight remains secure. Our operations dispatch is actively coordinating with carrier partners.\n\n"
                f"{custom_clean if custom_clean else 'We will provide the updated arrival telemetry as soon as terminal gate processing completes.'}\n\n"
                "Best regards,\n"
                "LogisticsHQ Freight Operations Team"
            )
            action_name = "shipments.notify_delay"

        elif "invoice" in event_type or "finance" in event_type:
            subject = f"STATEMENT NOTICE: Account Balance Inquiry - Invoice #{req.source_record_id}"
            body = (
                f"Dear {req.recipient_name},\n\n"
                f"This is a courteous reminder regarding Invoice #{req.source_record_id}.\n\n"
                f"Details: {topic_clean or 'Outstanding freight charges are pending remittance'}.\n"
                "Please verify payment status with your accounts payable department.\n\n"
                f"{custom_clean if custom_clean else 'Wire transfer instructions are attached to the original invoice. Please share remittance advice upon dispatch.'}\n\n"
                "Thank you for your business,\n"
                "LogisticsHQ Finance & Receivables Team"
            )
            action_name = "finance.send_collection_reminder"

        elif "contract" in event_type or "compliance" in event_type:
            subject = f"COMPLIANCE INQUIRY: Documentation Update for Agreement #{req.source_record_id}"
            body = (
                f"Dear {req.recipient_name},\n\n"
                f"In connection with Master Agreement #{req.source_record_id}, our compliance records require an updated filing.\n\n"
                f"Requirement: {topic_clean or 'Annual Certificate of Insurance or Partner Filing'}.\n\n"
                f"{custom_clean if custom_clean else 'Please reply with the requested document within 5 business days to avoid operational service holds.'}\n\n"
                "Sincerely,\n"
                "LogisticsHQ Commercial Compliance Office"
            )
            action_name = "contracts.request_missing_document"

        else:
            subject = f"Commercial Inquiry: Ref #{req.source_record_id}"
            body = (
                f"Dear {req.recipient_name},\n\n"
                f"Regarding Ref #{req.source_record_id}: {topic_clean}.\n\n"
                f"{custom_clean}\n\n"
                "LogisticsHQ Customer Service"
            )
            action_name = "notifications.create"

        return WorkflowDraftResponse(
            subject=subject,
            message_body=body,
            recommended_action_name=action_name,
            confidence_score=0.92,
            correlation_id=req.correlation_id,
        )
