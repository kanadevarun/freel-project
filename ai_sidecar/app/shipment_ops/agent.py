"""
Shipment Operations Automation and Intelligent Exception Response Agent
Phase 3 Task 3.5 for LogisticsHQ

Pure Python AI Agent providing structured shipment risk analysis, exception classification & prioritization,
actionable operational recommendations, and grounded communication drafting.
Adheres strictly to the requirement: Zero direct database mutations;
authoritative operational states, tracking, and approval decisions remain governed by Go backend.
"""

from datetime import datetime
import re
from typing import List, Dict, Any, Optional

from app.shipment_ops.models import (
    ShipmentContext,
    DeterministicShipmentSignals,
    MilestoneFact,
    ExceptionFact,
    EvidenceFact,
    SafetyWarning,
    PrioritizedExceptionItem,
    ExceptionPrioritizationResponse,
    OperationalRecommendationItem,
    OperationalRecommendationsResponse,
    ShipmentRiskAnalysisResponse,
    CommunicationDraftRequest,
    CommunicationDraftResponse,
)


class ShipmentOpsAgent:
    """
    Dedicated AI agent for shipment risk synthesis, exception triage, and communication drafting.
    """

    INJECTION_PATTERNS = [
        re.compile(r"ignore\s+(all\s+)?previous\s+instructions", re.IGNORECASE),
        re.compile(r"bypass\s+(manager\s+)?approval", re.IGNORECASE),
        re.compile(r"system\s*:\s*you\s+are", re.IGNORECASE),
        re.compile(r"cancel\s+(all\s+)?shipments?", re.IGNORECASE),
        re.compile(r"mark\s+(all\s+)?delivered", re.IGNORECASE),
        re.compile(r"set\s+status\s+to", re.IGNORECASE),
    ]

    def _sanitize_input(self, text: Optional[str]) -> str:
        if not text:
            return ""
        sanitized = text.strip()
        for pattern in self.INJECTION_PATTERNS:
            if pattern.search(sanitized):
                sanitized = pattern.sub("[FILTERED_UNSAFE_INSTRUCTION]", sanitized)
        return sanitized

    def analyze_shipment_risks(
        self,
        ctx: ShipmentContext,
        signals: DeterministicShipmentSignals
    ) -> ShipmentRiskAnalysisResponse:
        """
        Synthesizes multi-signal operational risks grounded strictly in verified Go backend facts.
        """
        evidence: List[EvidenceFact] = []
        warnings: List[SafetyWarning] = []
        key_signals: List[str] = []
        recommended_next_steps: List[str] = []

        # 1. Evaluate Milestone Risks
        if signals.has_overdue_milestone:
            for ms_code in signals.overdue_milestones:
                key_signals.append(f"Overdue operational milestone: {ms_code}")
                warnings.append(SafetyWarning(
                    warning_type="OVERDUE_MILESTONE",
                    message=f"Milestone '{ms_code}' has exceeded its planned completion timestamp.",
                    severity="HIGH"
                ))
            evidence.append(EvidenceFact(
                source_module="milestones",
                source_entity_id=ctx.shipment_id,
                source_ref=f"SH-{ctx.shipment_id}",
                field_name="overdue_milestones",
                observed_value=", ".join(signals.overdue_milestones),
                description="Deterministic delay detected against sequential transit schedule"
            ))
            recommended_next_steps.append("Request carrier clarification regarding delayed milestone progression.")

        # 2. Evaluate Active Exceptions
        if signals.has_active_exception:
            key_signals.append(f"{signals.active_exceptions_count} active exception(s) open ({signals.critical_exceptions_count} critical)")
            sev = "CRITICAL" if signals.critical_exceptions_count > 0 else "WARNING"
            warnings.append(SafetyWarning(
                warning_type="ACTIVE_EXCEPTION",
                message=f"Shipment has {signals.active_exceptions_count} unresolved exception(s) requiring operations triage.",
                severity=sev
            ))
            for ex in ctx.exceptions:
                if not ex.resolved and ex.status != "DISMISSED":
                    evidence.append(EvidenceFact(
                        source_module="exceptions",
                        source_entity_id=ex.exception_id,
                        source_ref=f"EX-{ex.exception_id}",
                        field_name="severity",
                        observed_value=ex.severity,
                        description=f"Active exception '{ex.title}' ({ex.exception_type})"
                    ))
            recommended_next_steps.append("Prioritize resolution of open exceptions with carrier and port agents.")

        # 3. Evaluate Tracking Freshness
        if signals.is_tracking_stale:
            key_signals.append(f"Tracking data is stale ({signals.hours_since_last_tracking:.1f} hours without carrier updates)")
            warnings.append(SafetyWarning(
                warning_type="STALE_TRACKING",
                message=f"No telematics or carrier status update received in {signals.hours_since_last_tracking:.1f} hours.",
                severity="WARNING"
            ))
            evidence.append(EvidenceFact(
                source_module="tracking",
                source_entity_id=ctx.shipment_id,
                source_ref=f"SH-{ctx.shipment_id}",
                field_name="hours_since_last_tracking",
                observed_value=f"{signals.hours_since_last_tracking:.1f}h",
                description="Lack of fresh telemetry from ocean/air carrier polling"
            ))
            recommended_next_steps.append("Initiate automated carrier tracking poll or manual dispatch inquiry.")

        # 4. Evaluate Documentation & Compliance
        if signals.missing_documents_count > 0:
            key_signals.append(f"Missing mandatory shipping documents: {', '.join(signals.missing_documents)}")
            warnings.append(SafetyWarning(
                warning_type="MISSING_DOCUMENTS",
                message=f"Shipment is missing {signals.missing_documents_count} required document(s) for clearance.",
                severity="WARNING"
            ))
            evidence.append(EvidenceFact(
                source_module="documents",
                source_entity_id=ctx.shipment_id,
                source_ref=f"SH-{ctx.shipment_id}",
                field_name="missing_documents",
                observed_value=", ".join(signals.missing_documents),
                description="Trade compliance / clearance documentation gap"
            ))
            recommended_next_steps.append("Request missing shipping documents from customer or freight forwarder.")

        # 5. Evaluate Carrier Response Gap
        if signals.is_carrier_response_overdue:
            key_signals.append(f"Carrier response is overdue by {signals.carrier_response_gap_hours:.1f} hours")
            warnings.append(SafetyWarning(
                warning_type="CARRIER_RESPONSE_OVERDUE",
                message="Carrier has failed to respond to operational inquiry within SLA window.",
                severity="HIGH"
            ))
            recommended_next_steps.append("Escalate issue to carrier key account manager.")

        # Determine Root Cause Narration
        root_causes = []
        if signals.critical_exceptions_count > 0:
            root_causes.append("Critical operational disruption on transit leg")
        if signals.has_overdue_milestone:
            root_causes.append("Sequential milestone schedule slippage")
        if signals.missing_documents_count > 0:
            root_causes.append("Documentation hold impacting cargo release")
        if signals.is_tracking_stale:
            root_causes.append("Carrier EDI/telemetry blackout")

        likely_root_cause = " & ".join(root_causes) if root_causes else "Normal operational progression within expected tolerances."

        # Operational Summary
        containers_str = ", ".join(ctx.container_numbers) if ctx.container_numbers else "No container assigned"
        summary = (
            f"Shipment #{ctx.shipment_id} ({ctx.carrier_scac}, MBL: {ctx.mbl_number or 'N/A'}) on lane "
            f"{ctx.origin_port} -> {ctx.destination_port} is currently in status '{ctx.status}' with risk level '{signals.risk_level}' "
            f"(Score: {signals.risk_score:.1f}/100). {len(key_signals)} active operational signals identified. "
            f"Cargo: {containers_str}. {likely_root_cause}"
        )

        return ShipmentRiskAnalysisResponse(
            shipment_id=ctx.shipment_id,
            risk_level=signals.risk_level,
            risk_score=signals.risk_score,
            operational_summary=summary,
            key_signals=key_signals,
            likely_root_cause=likely_root_cause,
            recommended_next_steps=recommended_next_steps or ["Continue routine operational monitoring."],
            evidence=evidence,
            warnings=warnings,
            confidence_score=0.96,
            correlation_id=ctx.correlation_id
        )

    def prioritize_exceptions(
        self,
        ctx: ShipmentContext,
        signals: DeterministicShipmentSignals
    ) -> ExceptionPrioritizationResponse:
        """
        Ranks active exceptions by operational urgency, financial exposure, and downstream impact.
        """
        prioritized_items: List[PrioritizedExceptionItem] = []
        escalations: List[str] = []
        evidence: List[EvidenceFact] = []
        warnings: List[SafetyWarning] = []

        # Sort exceptions: CRITICAL first, then HIGH, then MEDIUM, then LOW
        severity_weight = {"CRITICAL": 4, "HIGH": 3, "MEDIUM": 2, "LOW": 1}
        active_exceptions = [e for e in ctx.exceptions if not e.resolved and e.status != "DISMISSED"]

        sorted_exceptions = sorted(
            active_exceptions,
            key=lambda x: severity_weight.get(x.severity, 0),
            reverse=True
        )

        for rank, ex in enumerate(sorted_exceptions, start=1):
            if ex.severity == "CRITICAL":
                urgency = "IMMEDIATE"
                impact = "Vessel rollover or customs detainment risk; imminent port demurrage and SLA breach."
                deadline = "Within 4 hours"
                rec_action = "Escalate to Operations Director and initiate direct carrier line escalation."
                escalations.append(f"Immediate managerial escalation required for Exception #{ex.exception_id} ({ex.title})")
            elif ex.severity == "HIGH":
                urgency = "HIGH"
                impact = "Significant schedule variance (>48h) or transshipment connection miss."
                deadline = "Within 12 hours"
                rec_action = "Request carrier operational follow-up and prepare proactive customer update."
            elif ex.severity == "MEDIUM":
                urgency = "MEDIUM"
                impact = "Minor schedule delay or pending document reconciliation."
                deadline = "Within 24 hours"
                rec_action = "Acknowledge exception and follow up with terminal or broker."
            else:
                urgency = "LOW"
                impact = "Informational variance with minimal downstream cargo impact."
                deadline = "Within 48 hours"
                rec_action = "Monitor next automated milestone event."

            prioritized_items.append(PrioritizedExceptionItem(
                exception_id=ex.exception_id,
                priority_rank=rank,
                urgency=urgency,
                business_impact=impact,
                action_deadline=deadline,
                rationale=f"Exception '{ex.title}' ({ex.exception_type}) holds severity '{ex.severity}' on active shipment #{ctx.shipment_id}.",
                recommended_action=rec_action
            ))

            evidence.append(EvidenceFact(
                source_module="exceptions",
                source_entity_id=ex.exception_id,
                source_ref=f"EX-{ex.exception_id}",
                field_name="severity",
                observed_value=ex.severity,
                description=f"Rank {rank}: {ex.title}"
            ))

        if escalations:
            warnings.append(SafetyWarning(
                warning_type="EXCEPTION_ESCALATION_REQUIRED",
                message=f"{len(escalations)} high-priority exception(s) require escalation protocol.",
                severity="CRITICAL"
            ))

        return ExceptionPrioritizationResponse(
            shipment_id=ctx.shipment_id,
            prioritized_exceptions=prioritized_items,
            recommended_escalations=escalations,
            evidence=evidence,
            warnings=warnings,
            confidence_score=0.98,
            correlation_id=ctx.correlation_id
        )

    def generate_operational_recommendations(
        self,
        ctx: ShipmentContext,
        signals: DeterministicShipmentSignals
    ) -> OperationalRecommendationsResponse:
        """
        Generates actionable recommendations mapped directly to the centralized Action System.
        """
        recs: List[OperationalRecommendationItem] = []
        evidence: List[EvidenceFact] = []
        warnings: List[SafetyWarning] = []

        # 1. Exception Escalation (High Risk, Requires Approval)
        if signals.critical_exceptions_count > 0:
            crit_ex = next((e for e in ctx.exceptions if e.severity == "CRITICAL" and not e.resolved), None)
            ex_id = crit_ex.exception_id if crit_ex else (ctx.exceptions[0].exception_id if ctx.exceptions else 0)
            recs.append(OperationalRecommendationItem(
                action_type="shipments.escalate_exception",
                title="Escalate Critical Shipment Exception to Operations Lead",
                description=f"Formal escalation for unresolved critical disruption on shipment #{ctx.shipment_id}.",
                target_role="Operations Manager",
                priority="CRITICAL",
                requires_approval=True,
                suggested_payload={
                    "shipment_id": ctx.shipment_id,
                    "exception_id": ex_id,
                    "reason": "Compounding delay risk exceeding acceptable operational tolerance",
                    "carrier_scac": ctx.carrier_scac,
                    "mbl_number": ctx.mbl_number
                },
                rationale="Critical exception threatens cargo delivery commitment and requires formal supervisor intervention."
            ))

        # 2. Carrier Follow-up (High Risk, Requires Approval)
        if signals.has_overdue_milestone or signals.is_carrier_response_overdue or signals.is_tracking_stale:
            recs.append(OperationalRecommendationItem(
                action_type="shipments.request_carrier_followup",
                title=f"Request Urgent Carrier Status Clarification from {ctx.carrier_name or ctx.carrier_scac}",
                description="Draft and dispatch formal carrier inquiry regarding overdue milestone progression.",
                target_role="Carrier Coordinator",
                priority="HIGH",
                requires_approval=True,
                suggested_payload={
                    "shipment_id": ctx.shipment_id,
                    "carrier_scac": ctx.carrier_scac,
                    "mbl_number": ctx.mbl_number,
                    "booking_number": ctx.booking_number,
                    "overdue_milestones": signals.overdue_milestones
                },
                rationale="Missing carrier updates prevent accurate transit estimation and risk customer dissatisfaction."
            ))

        # 3. Customer Update (High Risk, Requires Approval)
        if signals.risk_level in ["HIGH", "CRITICAL"] or signals.is_delayed:
            recs.append(OperationalRecommendationItem(
                action_type="shipments.send_customer_update",
                title=f"Dispatch Proactive Status Update to {ctx.customer_name}",
                description=f"Send approved operational advisory regarding revised transit schedule ({signals.delay_days:.1f}d variance).",
                target_role="Customer Service Specialist",
                priority="HIGH",
                requires_approval=True,
                suggested_payload={
                    "shipment_id": ctx.shipment_id,
                    "customer_id": ctx.customer_id,
                    "customer_name": ctx.customer_name,
                    "delay_days": signals.delay_days,
                    "current_status": ctx.status
                },
                rationale="Proactive communication preserves customer relationship and prevents unassisted escalation."
            ))

        # 4. Document Review (Safe Internal)
        if signals.missing_documents_count > 0:
            recs.append(OperationalRecommendationItem(
                action_type="shipments.request_document_review",
                title="Request Expedited Document Review from Compliance Team",
                description=f"Resolve {signals.missing_documents_count} missing documentation requirements prior to port arrival.",
                target_role="Trade Compliance Auditor",
                priority="MEDIUM",
                requires_approval=False,
                suggested_payload={
                    "shipment_id": ctx.shipment_id,
                    "missing_documents": signals.missing_documents
                },
                rationale="Missing shipping documents can trigger customs holds and container detention fees."
            ))

        # 5. Internal Operations Task (Safe Internal)
        recs.append(OperationalRecommendationItem(
            action_type="shipments.create_internal_task",
            title=f"Monitor Vessel ETA Variance for {ctx.vessel_name or 'Shipment Leg'}",
            description="Verify AIS position and scheduled discharge window 48 hours prior to arrival.",
            target_role="Operations Specialist",
            priority="LOW",
            requires_approval=False,
            suggested_payload={
                "shipment_id": ctx.shipment_id,
                "vessel_name": ctx.vessel_name,
                "eta": ctx.eta
            },
            rationale="Standard operational diligence for ocean/air freight in transit."
        ))

        evidence.append(EvidenceFact(
            source_module="shipments",
            source_entity_id=ctx.shipment_id,
            source_ref=f"SH-{ctx.shipment_id}",
            field_name="risk_score",
            observed_value=f"{signals.risk_score:.1f}",
            description=f"Generated {len(recs)} bounded operational actions for risk score {signals.risk_score:.1f}"
        ))

        return OperationalRecommendationsResponse(
            shipment_id=ctx.shipment_id,
            recommendations=recs,
            evidence=evidence,
            warnings=warnings,
            confidence_score=0.97,
            correlation_id=ctx.correlation_id
        )

    def generate_communication_draft(
        self,
        req: CommunicationDraftRequest
    ) -> CommunicationDraftResponse:
        """
        Synthesizes an editable, strictly grounded communication draft.
        Never fabricates dates or carrier commitments.
        """
        ctx = req.context
        signals = req.signals
        sanitized_instructions = self._sanitize_input(req.user_instructions)

        recipient_name = req.recipient_name or ctx.customer_name or "Operations Partner"
        recipient_email = req.recipient_email or "operations@logisticshq.internal"

        evidence: List[EvidenceFact] = []
        warnings: List[SafetyWarning] = []
        proposed_actions: List[str] = []

        mbl = ctx.mbl_number or "Pending"
        booking = ctx.booking_number or "N/A"
        origin = ctx.origin_port
        destination = ctx.destination_port
        vessel = f"{ctx.vessel_name or 'Vessel'} {ctx.voyage_number or ''}".strip()
        etd_str = ctx.etd[:10] if ctx.etd else "Scheduled"
        eta_str = ctx.eta[:10] if ctx.eta else "Pending Confirmation"

        if req.draft_type == "CARRIER_FOLLOWUP":
            subject = f"URGENT: Tracking & Milestone Update Request - MBL: {mbl} / Booking: {booking} - [{ctx.carrier_scac}]"
            wording = (
                f"Dear {ctx.carrier_name or ctx.carrier_scac} Operations Team,\n\n"
                f"We are requesting an immediate operational update regarding shipment #{ctx.shipment_id} under "
                f"Master Bill of Lading #{mbl} (Booking Ref: {booking}).\n\n"
                f"Routing Details:\n"
                f"  • Origin Port: {origin}\n"
                f"  • Destination Port: {destination}\n"
                f"  • Vessel / Voyage: {vessel}\n"
                f"  • Scheduled ETA: {eta_str}\n\n"
                f"Our automated monitoring indicates the following item(s) require your immediate attention:\n"
            )
            if signals.overdue_milestones:
                wording += f"  • Overdue Milestone(s): {', '.join(signals.overdue_milestones)}\n"
            if signals.is_tracking_stale:
                wording += f"  • Tracking Stale: {signals.hours_since_last_tracking:.1f} hours without telemetry update\n"
            if signals.has_active_exception:
                wording += f"  • Active Exception Count: {signals.active_exceptions_count}\n"

            if sanitized_instructions:
                wording += f"\nOperator Note: {sanitized_instructions}\n"

            wording += (
                f"\nPlease provide confirmed container status, current position, and validated ETA at your earliest convenience.\n\n"
                f"Sincerely,\n"
                f"LogisticsHQ Global Freight Operations"
            )
            internal_notes = "Grounded carrier inquiry. Requires dispatcher verification before dispatch."
            proposed_actions.append("shipments.request_carrier_followup")

        elif req.draft_type == "CUSTOMER_UPDATE":
            subject = f"LogisticsHQ Shipment Advisory: Status Update for {mbl} ({origin} to {destination})"
            wording = (
                f"Dear {recipient_name},\n\n"
                f"We are writing to provide a proactive transit update regarding your consignment (Ref: {mbl}).\n\n"
                f"Shipment Summary:\n"
                f"  • Current Status: {ctx.status}\n"
                f"  • Origin: {origin}\n"
                f"  • Destination: {destination}\n"
                f"  • Carrier / Vessel: {ctx.carrier_name} ({vessel})\n"
                f"  • Estimated Arrival: {eta_str}\n\n"
            )
            if signals.is_delayed:
                wording += (
                    f"Operational Advisory:\n"
                    f"Our tracking systems have flagged an operational variance of approximately {signals.delay_days:.1f} day(s) "
                    f"resulting from upstream transit delays. Our operations specialists are actively liaising with the carrier "
                    f"to expedite port handling upon arrival.\n\n"
                )
            else:
                wording += (
                    f"Cargo is progressing through the planned routing. All primary milestones are being tracked actively "
                    f"by our 24/7 operations monitoring center.\n\n"
                )

            if sanitized_instructions:
                wording += f"Additional Context: {sanitized_instructions}\n\n"

            wording += (
                f"We will notify you immediately as soon as the vessel berths and discharge commences.\n\n"
                f"Warm regards,\n"
                f"LogisticsHQ Dedicated Customer Care"
            )
            internal_notes = "Customer-facing transit advisory. Cost and internal operational comments have been excluded."
            proposed_actions.append("shipments.send_customer_update")

        else:  # INTERNAL_ESCALATION
            subject = f"INTERNAL ESCALATION: Shipment #{ctx.shipment_id} - Risk Level: {signals.risk_level} (Score: {signals.risk_score:.1f})"
            wording = (
                f"Operations Management Briefing:\n"
                f"Shipment ID: #{ctx.shipment_id}\n"
                f"Customer: {ctx.customer_name} ({ctx.customer_tier} Tier)\n"
                f"Carrier: {ctx.carrier_name} ({ctx.carrier_scac})\n"
                f"MBL: {mbl} | Booking: {booking}\n"
                f"Lane: {origin} -> {destination}\n\n"
                f"Compounding Risk Signals:\n"
                f"  • Risk Score: {signals.risk_score:.1f} / 100 ({signals.risk_level})\n"
                f"  • Active Exceptions: {signals.active_exceptions_count} ({signals.critical_exceptions_count} Critical)\n"
                f"  • Overdue Milestones: {', '.join(signals.overdue_milestones) if signals.overdue_milestones else 'None'}\n"
                f"  • Stale Telemetry: {'Yes' if signals.is_tracking_stale else 'No'} ({signals.hours_since_last_tracking:.1f}h)\n"
                f"  • Missing Documents: {signals.missing_documents_count}\n\n"
                f"Immediate Recommendation:\n"
                f"Escalate to Carrier Key Account Manager and dispatch notice to port agent to secure priority discharge.\n"
            )
            if sanitized_instructions:
                wording += f"\nEscalation Notes: {sanitized_instructions}\n"
            internal_notes = "Managerial escalation brief. Restrict to internal team."
            proposed_actions.append("shipments.escalate_exception")

        evidence.append(EvidenceFact(
            source_module="shipments",
            source_entity_id=ctx.shipment_id,
            source_ref=f"SH-{ctx.shipment_id}",
            field_name="draft_type",
            observed_value=req.draft_type,
            description=f"Synthesized {req.draft_type} draft based on verified shipment #{ctx.shipment_id} parameters"
        ))

        warnings.append(SafetyWarning(
            warning_type="DRAFT_REQUIRES_APPROVAL",
            message="This operational message is in DRAFT state and requires explicit managerial approval before transmission.",
            severity="WARNING"
        ))

        return CommunicationDraftResponse(
            shipment_id=ctx.shipment_id,
            draft_type=req.draft_type,
            subject=subject,
            customer_wording=wording,
            internal_notes=internal_notes,
            recipient_preview={
                "name": recipient_name,
                "email": recipient_email
            },
            proposed_actions=proposed_actions,
            requires_approval=True,
            evidence=evidence,
            warnings=warnings,
            confidence_score=0.99,
            correlation_id=req.correlation_id
        )
