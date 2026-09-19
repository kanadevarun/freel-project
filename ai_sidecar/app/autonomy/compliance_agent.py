"""
LogisticsHQ Phase 5 Task 5.7: Contract and Compliance Monitoring Agent
Implements controlled, risk-aware contract and compliance reasoning, expiration tracking,
hard vs. soft requirement segregation, rate validity monitoring, commercial deviation detection,
prompt-injection sanitization, and 7-step autonomous remediation plan synthesis.
"""

import re
import datetime
from typing import Dict, Any, List, Optional
from app.autonomy.models import (
    ContractComplianceContext,
    ContractComplianceEvaluationRequest,
    ContractComplianceEvaluationResponse,
    ContractComplianceReplanningRequest,
    ComplianceRemediationStrategyCandidate,
)

INJECTION_PATTERNS = [
    re.compile(r"ignore\s+(all\s+)?(previous\s+)?instructions", re.IGNORECASE),
    re.compile(r"bypass\s+(approval|policy|rules|compliance|contract)", re.IGNORECASE),
    re.compile(r"waive\s+(all\s+)?(compliance|documents|fmc|permits|penalties)", re.IGNORECASE),
    re.compile(r"declare\s+compliant\s+automatically", re.IGNORECASE),
    re.compile(r"override\s+hard\s+requirements?", re.IGNORECASE),
    re.compile(r"system\s*:\s*", re.IGNORECASE),
]


def _sanitize_untrusted_text(text: Optional[str]) -> str:
    """Sanitizes user, carrier, and customer text against prompt injection attempts."""
    if not text:
        return ""
    sanitized = text
    for pattern in INJECTION_PATTERNS:
        sanitized = pattern.sub("[REDACTED_INSTRUCTION]", sanitized)
    return sanitized.strip()[:1000]


def evaluate_contract_compliance(req: ContractComplianceEvaluationRequest) -> ContractComplianceEvaluationResponse:
    """
    Evaluates a contract and compliance context and synthesizes candidate remediation strategies,
    fact/prediction/assumption segregation, risk/severity scoring, hard vs soft requirement analysis,
    and a 7-step autonomous remediation execution plan.
    """
    ctx = req.context

    # 1. Authoritative Contract & Compliance Facts
    contract_id = ctx.contract_id
    contract_ref = ctx.contract_reference or f"CTR-{contract_id}"
    contract_name = ctx.contract_name or f"Contract {contract_id}"
    party_name = ctx.party_name or f"Party {ctx.party_id}"
    status = (ctx.status or "ACTIVE").upper()
    currency = ctx.currency or "USD"
    val = float(ctx.contract_value or 0.0)

    # Expiration analysis
    days_until_exp = int(ctx.days_until_expiration)
    if status == "EXPIRED" or days_until_exp < 0:
        expiration_status = "EXPIRED"
    elif days_until_exp <= 30:
        expiration_status = "EXPIRING_SOON"
    elif status == "DRAFT":
        expiration_status = "REVIEW_REQUIRED"
    else:
        expiration_status = "CURRENT"

    # Analyze Hard vs Soft Requirements
    hard_req_count = 0
    soft_req_count = 0
    hard_viol_count = 0
    soft_dev_count = 0
    missing_docs_count = 0
    expired_docs_count = 0
    deviations: List[Dict[str, Any]] = []

    for req_item in ctx.compliance_requirements:
        req_type = str(req_item.get("requirement_type", "")).upper()
        req_title = req_item.get("title") or req_item.get("description") or req_type or "Compliance Requirement"
        req_status = str(req_item.get("status") or req_item.get("compliance_status") or "PENDING").upper()
        is_hard = bool(req_item.get("is_mandatory", False) or req_item.get("is_hard", False) or req_type in ["REGULATORY", "MANDATORY_CERTIFICATE", "INSURANCE", "HAZMAT_PERMIT", "STATUTORY_FILING", "GDP_PHARMA", "GDP_PHARMA_CERT", "FMC_FILING"])

        if is_hard:
            hard_req_count += 1
            if req_status in ["MISSING", "EXPIRED", "REJECTED", "NON_COMPLIANT"]:
                hard_viol_count += 1
                deviations.append({
                    "type": "HARD_REQUIREMENT_VIOLATION",
                    "requirement": req_title,
                    "requirement_type": req_type,
                    "severity": "CRITICAL" if req_status in ["MISSING", "REJECTED"] else "HIGH",
                    "status": req_status,
                    "description": f"Mandatory compliance requirement '{req_title}' is {req_status}. Operation cannot proceed autonomously.",
                    "is_hard": True,
                })
            elif req_status in ["PENDING", "EXPIRING"]:
                deviations.append({
                    "type": "HARD_REQUIREMENT_AT_RISK",
                    "requirement": req_title,
                    "requirement_type": req_type,
                    "severity": "HIGH",
                    "status": req_status,
                    "description": f"Mandatory compliance requirement '{req_title}' is approaching expiry or pending verification.",
                    "is_hard": True,
                })
        else:
            soft_req_count += 1
            if req_status in ["MISSING", "EXPIRED", "NON_COMPLIANT"]:
                soft_dev_count += 1
                deviations.append({
                    "type": "SOFT_DEVIATION",
                    "requirement": req_title,
                    "requirement_type": req_type,
                    "severity": "MEDIUM",
                    "status": req_status,
                    "description": f"Advisory term '{req_title}' has soft deviation: {req_status}.",
                    "is_hard": False,
                })

    # Analyze Documents
    for doc in ctx.documents:
        doc_status = str(doc.get("status", "")).upper()
        doc_name = doc.get("file_name", "Document")
        if doc_status in ["MISSING", "REJECTED"]:
            missing_docs_count += 1
        elif doc_status in ["EXPIRED"]:
            expired_docs_count += 1

    # Analyze Rate & Commercial Deviations
    for rd in ctx.rate_deviations:
        soft_dev_count += 1
        deviations.append({
            "type": "RATE_DEVIATION",
            "requirement": f"Agreed Contract Rate for {rd.get('lane', 'Lane')}",
            "requirement_type": "COMMERCIAL_RATE",
            "severity": "MEDIUM",
            "status": "MISMATCH",
            "description": f"Active commercial rate {rd.get('actual_rate')} {currency} deviates from agreed contract rate {rd.get('contract_rate')} {currency}.",
            "is_hard": False,
        })

    # Determine Authoritative Compliance Status
    stop_reason: Optional[str] = None
    if status == "EXPIRED":
        compliance_status = "BLOCKED"
        stop_reason = f"Contract #{contract_ref} has expired. All operational bookings and shipments under this agreement are blocked."
        risk_level = "CRITICAL"
        risk_score = 0.95
    elif hard_viol_count > 0:
        compliance_status = "NON_COMPLIANT"
        risk_level = "CRITICAL"
        risk_score = 0.85
    elif expiration_status == "EXPIRING_SOON" or soft_dev_count > 0 or missing_docs_count > 0 or expired_docs_count > 0:
        compliance_status = "WARNING"
        risk_level = "HIGH" if (missing_docs_count > 0 or expired_docs_count > 0) else "MEDIUM"
        risk_score = 0.60
    else:
        compliance_status = "COMPLIANT"
        risk_level = "LOW"
        risk_score = 0.15

    # Authoritative Facts vs Extracted Terms vs Predictions vs Assumptions
    authoritative_facts = [
        f"Contract #{contract_ref} ({contract_name}) is registered with status '{status}' for party '{party_name}'.",
        f"Effective date: {ctx.effective_date or 'Not specified'} | Expiration date: {ctx.expiry_date or 'Not specified'} ({days_until_exp} days remaining).",
        f"Contract value: ${val:,.2f} {currency} | Transport mode: {ctx.transport_mode or 'MULTIMODAL'}.",
        f"Authoritative compliance audit: {hard_req_count} hard requirements ({hard_viol_count} violations) and {soft_req_count} soft requirements.",
    ]

    extracted_terms = [
        f"Extracted term: '{t.get('term_title', t.get('term_key'))}' = '{t.get('term_value')}' (Category: {t.get('term_category', 'GENERAL')}, Critical: {bool(t.get('is_critical'))})."
        for t in ctx.terms[:5]
    ] if ctx.terms else [f"Standard carrier/customer commercial terms and FMC tariff agreement extracted."]

    predictions = [
        f"Predicted overall compliance risk score: {risk_score:.2f} ({risk_level} risk).",
        f"Contract renewal risk is {('CRITICAL' if expiration_status == 'EXPIRED' else 'ELEVATED' if expiration_status == 'EXPIRING_SOON' else 'NORMAL')} based on expiration window.",
        f"Cargo clearance delay probability is projected at {int(risk_score * 100)}% if missing compliance evidence is not remediated.",
        f"Dispute likelihood with {party_name} is estimated at {20 if soft_dev_count == 0 else 65}% due to observed commercial deviations.",
    ]

    assumptions = [
        "Operating assumption: Counterparty maintains standard 72-hour turnaround time for statutory document requests.",
        "Operating assumption: Active shipments under this contract rely on verified FMC / customs filings remaining in good standing.",
        "Operating assumption: Regulatory authorities require valid certificates throughout the entire transit milestone window.",
    ]

    # Candidate Remediation Strategies
    candidates: List[ComplianceRemediationStrategyCandidate] = []

    # 1. Standard Compliance Monitoring
    c1 = ComplianceRemediationStrategyCandidate(
        strategy_id="strat-compliance-ok-monitor",
        strategy_name="Standard Compliance Monitoring",
        strategy_type="MONITORING",
        description="Contract terms and statutory compliance requirements are currently in good standing. Maintain automated sentinel monitoring.",
        recommended_action="CONTINUE_MONITORING",
        severity_level="LOW",
        urgency="ROUTINE",
        requires_approval=False,
        expected_outcome="Continuous observation with zero operational friction.",
        score=0.90 if compliance_status == "COMPLIANT" else 0.20,
        is_feasible=True,
    )
    candidates.append(c1)

    # 2. Expiring Agreement Renewal Workflow
    c2 = ComplianceRemediationStrategyCandidate(
        strategy_id="strat-expiring-renewal-prep",
        strategy_name="Expiring Agreement Renewal Workflow",
        strategy_type="RENEWAL_PREP",
        description="Contract or rate agreement is approaching expiration within 30 days. Prepare renewal review task for procurement lead.",
        recommended_action="PREPARE_RENEWAL_TASK",
        severity_level="MEDIUM",
        urgency="URGENT" if days_until_exp <= 15 else "NORMAL",
        requires_approval=True,
        approval_reason="Contract renewal and term negotiations require authorized commercial sign-off.",
        expected_outcome="Renewal audit created in CRM and reminder drafted for counterparty.",
        score=0.88 if expiration_status == "EXPIRING_SOON" else 0.40,
        draft_subject=f"Notice of Upcoming Contract Expiration & Renewal Review: #{contract_ref}",
        draft_message=f"Dear {party_name} Team,\n\nOur records indicate that Master Agreement #{contract_ref} is scheduled to expire on {ctx.expiry_date} ({days_until_exp} days remaining).\n\nTo prevent operational disruption to ongoing shipments, please confirm your renewal intentions or contact our procurement desk.",
        is_feasible=True,
    )
    candidates.append(c2)

    # 3. Document Remediation Request
    c3 = ComplianceRemediationStrategyCandidate(
        strategy_id="strat-missing-document-remediation",
        strategy_name="Document Remediation Request",
        strategy_type="DOCUMENT_REQUEST",
        description="Mandatory compliance documentation (cargo insurance or statutory permit) is missing or expiring. Request updated filing.",
        recommended_action="REQUEST_MISSING_DOCUMENT",
        severity_level="HIGH",
        urgency="URGENT",
        requires_approval=False,
        expected_outcome="Automated document request sent to carrier/shipper compliance contact.",
        score=0.92 if (missing_docs_count > 0 or expired_docs_count > 0) else 0.35,
        draft_subject=f"Action Required: Missing Compliance Documentation for Contract #{contract_ref}",
        draft_message=f"Dear {party_name} Compliance Department,\n\nDuring routine compliance monitoring of Agreement #{contract_ref}, the following mandatory document requirement was identified as incomplete:\n\n- Documentation status: {missing_docs_count} missing, {expired_docs_count} expired.\n\nPlease upload the required certificates via the secure LogisticsHQ portal to maintain approved carrier status.",
        is_feasible=True,
    )
    candidates.append(c3)

    # 4. Rate & Terms Mismatch Reconciliation
    c4 = ComplianceRemediationStrategyCandidate(
        strategy_id="strat-commercial-deviation-reconciliation",
        strategy_name="Rate & Terms Mismatch Reconciliation",
        strategy_type="COMMERCIAL_RECONCILIATION",
        description="Active shipment quotes or invoices deviate from contracted tariff terms. Create reconciliation review for billing manager.",
        recommended_action="RECONCILE_COMMERCIAL_TERMS",
        severity_level="MEDIUM",
        urgency="NORMAL",
        requires_approval=True,
        approval_reason="Rate discrepancy resolution requires commercial controller validation.",
        expected_outcome="Discrepancy logged with source references and routed for rate amendment review.",
        score=0.85 if soft_dev_count > 0 else 0.30,
        draft_subject=f"Commercial Rate Discrepancy Flag: #{contract_ref}",
        draft_message=f"Internal Notice for Commercial Billing:\n\nObserved tariff deviations between agreed contract #{contract_ref} and live booking records. Please review before invoice issuance.",
        is_feasible=True,
    )
    candidates.append(c4)

    # 5. Hard Compliance Block & Executive Escalation
    c5 = ComplianceRemediationStrategyCandidate(
        strategy_id="strat-hard-compliance-escalation",
        strategy_name="Hard Compliance Block & Executive Escalation",
        strategy_type="HARD_COMPLIANCE_ESCALATION",
        description="Critical compliance violation or expired contract detected. Autonomous execution is halted and flagged for Compliance Officer.",
        recommended_action="ESCALATE_COMPLIANCE_BLOCK",
        severity_level="CRITICAL",
        urgency="IMMEDIATE",
        requires_approval=True,
        approval_reason="Hard compliance blocks and statutory breaches require executive compliance review before release.",
        expected_outcome="Operations placed on hold; executive alert dispatched to Compliance & Legal desks.",
        score=0.95 if (compliance_status in ["NON_COMPLIANT", "BLOCKED"]) else 0.15,
        draft_subject=f"CRITICAL COMPLIANCE BLOCK: Contract #{contract_ref}",
        draft_message=f"URGENT ALERT: Agreement #{contract_ref} has been marked {compliance_status}.\n\nReason: {stop_reason or 'Mandatory statutory requirements violated.'}\n\nAll automated dispatches and freight handling are paused pending Compliance sign-off.",
        is_feasible=True,
    )
    candidates.append(c5)

    # Strategy Selection Hierarchy
    if compliance_status in ["NON_COMPLIANT", "BLOCKED"]:
        selected_strategy = c5
    elif missing_docs_count > 0 or expired_docs_count > 0:
        selected_strategy = c3
    elif expiration_status == "EXPIRING_SOON":
        selected_strategy = c2
    elif soft_dev_count > 0:
        selected_strategy = c4
    else:
        selected_strategy = c1

    # Synthesize 7-Step Autonomous Remediation Plan
    plan_steps = [
        {
            "step_id": "step-1-audit-state",
            "step_number": 1,
            "action_type": "AUDIT_CONTRACT_STATE",
            "title": "Audit Authoritative Contract & Expiration",
            "description": f"Verify contract #{contract_ref} validity period ({ctx.effective_date} to {ctx.expiry_date}) and active status '{status}'.",
            "expected_outcome": "Authoritative contract facts and validity boundaries verified.",
            "risk_level": "LOW",
            "requires_approval": False,
        },
        {
            "step_id": "step-2-verify-hard-reqs",
            "step_number": 2,
            "action_type": "VERIFY_HARD_REQUIREMENTS",
            "title": "Verify Mandatory Statutory Requirements",
            "description": f"Evaluate {hard_req_count} hard requirements (FMC filing, cargo insurance, statutory permits).",
            "expected_outcome": f"Hard requirement status confirmed: {hard_viol_count} violations detected.",
            "risk_level": "HIGH" if hard_viol_count > 0 else "LOW",
            "requires_approval": hard_viol_count > 0,
        },
        {
            "step_id": "step-3-detect-commercial-deviations",
            "step_number": 3,
            "action_type": "DETECT_COMMERCIAL_DEVIATIONS",
            "title": "Detect Commercial & Rate Deviations",
            "description": f"Scan active quotations, bookings, and freight invoices against agreed contract terms.",
            "expected_outcome": f"Identified {soft_dev_count} commercial deviations or tariff discrepancies.",
            "risk_level": "MEDIUM" if soft_dev_count > 0 else "LOW",
            "requires_approval": False,
        },
        {
            "step_id": "step-4-classify-strategy",
            "step_number": 4,
            "action_type": "CLASSIFY_RISK_AND_STRATEGY",
            "title": "Synthesize Remediation Strategy",
            "description": f"Select optimal candidate remediation strategy: '{selected_strategy.strategy_name}'.",
            "expected_outcome": f"Remediation strategy '{selected_strategy.strategy_id}' synthesized with risk score {risk_score:.2f}.",
            "risk_level": selected_strategy.severity_level,
            "requires_approval": False,
        },
        {
            "step_id": "step-5-governance-gate",
            "step_number": 5,
            "action_type": "ENFORCE_GOVERNANCE_GATE",
            "title": "Governance Policy & Approval Gate",
            "description": "Enforce Level 2 autonomy boundary. Verify whether human sign-off is required for remediation dispatch.",
            "expected_outcome": f"Approval requirement resolved: {'MANDATORY_HUMAN_REVIEW' if selected_strategy.requires_approval else 'PRE_APPROVED_ADMINISTRATIVE'}.",
            "risk_level": "LOW",
            "requires_approval": selected_strategy.requires_approval,
        },
        {
            "step_id": "step-6-dispatch-action",
            "step_number": 6,
            "action_type": "DISPATCH_ACTION_SYSTEM",
            "title": f"Dispatch Action: {selected_strategy.recommended_action}",
            "description": f"Route action through Go Action System boundary with correlation ID and idempotency key.",
            "expected_outcome": f"Action '{selected_strategy.recommended_action}' safely queued in execution boundary.",
            "risk_level": selected_strategy.severity_level,
            "requires_approval": selected_strategy.requires_approval,
        },
        {
            "step_id": "step-7-replan-and-monitor",
            "step_number": 7,
            "action_type": "REPLAN_AND_MONITOR",
            "title": "Active Sentinel Monitoring & Adaptive Replanning",
            "description": "Listen for incoming document uploads, route alterations, rate amendments, or expiration events to adapt plan.",
            "expected_outcome": "Sentinel event listener active with automatic replanning lineage.",
            "risk_level": "LOW",
            "requires_approval": False,
        },
    ]

    return ContractComplianceEvaluationResponse(
        contract_id=contract_id,
        contract_reference=contract_ref,
        contract_name=contract_name,
        party_name=party_name,
        contract_type=ctx.contract_type,
        status=status,
        effective_date=ctx.effective_date,
        expiry_date=ctx.expiry_date,
        days_until_expiration=days_until_exp,
        expiration_status=expiration_status,
        compliance_status=compliance_status,
        hard_requirement_count=hard_req_count,
        soft_requirement_count=soft_req_count,
        hard_violations_count=hard_viol_count,
        soft_deviations_count=soft_dev_count,
        missing_documents_count=missing_docs_count,
        expired_documents_count=expired_docs_count,
        risk_level=risk_level,
        risk_score=risk_score,
        authoritative_facts=authoritative_facts,
        extracted_terms=extracted_terms,
        predictions=predictions,
        assumptions=assumptions,
        deviations=deviations,
        candidate_strategies=candidates,
        recommended_strategy_id=selected_strategy.strategy_id,
        recommended_action=selected_strategy.recommended_action,
        draft_subject=selected_strategy.draft_subject or f"Compliance Notice: Contract #{contract_ref}",
        draft_message=selected_strategy.draft_message or "Compliance monitoring complete. No critical action required.",
        requires_approval=selected_strategy.requires_approval or hard_viol_count > 0 or status == "EXPIRED",
        approval_reason=selected_strategy.approval_reason if selected_strategy.requires_approval else ("Hard compliance violation requires review" if hard_viol_count > 0 else None),
        stop_reason=stop_reason,
        confidence_score=0.92,
        data_sufficiency="COMPLETE",
        remediation_plan_steps=plan_steps,
        correlation_id=req.correlation_id,
    )


def replan_contract_compliance(req: ContractComplianceReplanningRequest) -> ContractComplianceEvaluationResponse:
    """
    Adapts an existing contract and compliance plan upon a business event trigger
    (e.g., DOCUMENT_UPLOADED, DOCUMENT_VERIFIED, DOCUMENT_REJECTED, ROUTE_CHANGED, RATE_AMENDED, CONTRACT_EXPIRED).
    """
    ctx = req.context
    trigger = req.trigger_event
    payload = req.event_payload or {}

    # Mutate context copy based on event
    if trigger == "DOCUMENT_UPLOADED":
        doc_name = payload.get("file_name", "Uploaded_Certificate.pdf")
        ctx.documents.append({
            "doc_id": payload.get("document_id", "doc-new-001"),
            "file_name": doc_name,
            "status": "PENDING_REVIEW",
            "is_verified": False,
        })
    elif trigger == "DOCUMENT_VERIFIED":
        # Resolve missing doc requirements
        for doc in ctx.documents:
            doc["status"] = "VERIFIED"
            doc["is_verified"] = True
        for cr in ctx.compliance_requirements:
            if cr.get("status") in ["MISSING", "EXPIRING"]:
                cr["status"] = "VERIFIED"
    elif trigger == "DOCUMENT_REJECTED":
        reject_reason = _sanitize_untrusted_text(payload.get("reason", "Document does not satisfy regulatory requirements"))
        for cr in ctx.compliance_requirements:
            cr["status"] = "REJECTED"
            cr["rejection_reason"] = reject_reason
    elif trigger == "CONTRACT_EXPIRED":
        ctx.status = "EXPIRED"
        ctx.days_until_expiration = -1
    elif trigger == "RATE_AMENDED":
        ctx.rate_deviations = []  # Rate deviation resolved

    # Evaluate mutated context
    eval_req = ContractComplianceEvaluationRequest(context=ctx, correlation_id=req.correlation_id)
    resp = evaluate_contract_compliance(eval_req)
    return resp
