"""
LogisticsHQ Phase 5 Task 5.6: Adaptive Finance and Collections Agent
Implements controlled, risk-aware financial reasoning, collection prioritization,
multi-strategy generation, dispute handling, multi-invoice consolidation,
prompt-injection sanitization, and 7-step autonomous plan synthesis.
"""

import re
import datetime
from typing import Dict, Any, List, Optional
from app.autonomy.models import (
    FinanceInvoiceContext,
    FinanceCollectionEvaluationRequest,
    FinanceCollectionEvaluationResponse,
    FinanceCollectionReplanningRequest,
    CollectionStrategyCandidate,
)

INJECTION_PATTERNS = [
    re.compile(r"ignore\s+(all\s+)?(previous\s+)?instructions", re.IGNORECASE),
    re.compile(r"bypass\s+(approval|policy|rules|finance)", re.IGNORECASE),
    re.compile(r"waive\s+(all\s+)?(debt|fees|balances)", re.IGNORECASE),
    re.compile(r"write\s*off\s+(automatically|all)", re.IGNORECASE),
    re.compile(r"system\s*:\s*", re.IGNORECASE),
    re.compile(r"set\s+balance\s+to\s+0", re.IGNORECASE),
]


def _sanitize_untrusted_text(text: Optional[str]) -> str:
    """Sanitizes user and customer text against prompt injection attempts."""
    if not text:
        return ""
    sanitized = text
    for pattern in INJECTION_PATTERNS:
        sanitized = pattern.sub("[REDACTED_INSTRUCTION]", sanitized)
    return sanitized.strip()[:1000]


def evaluate_finance_collection(req: FinanceCollectionEvaluationRequest) -> FinanceCollectionEvaluationResponse:
    """
    Evaluates an invoice context and synthesizes candidate collection strategies,
    fact/prediction segregation, risk/priority scoring, and a 7-step autonomous execution plan.
    """
    ctx = req.context

    # 1. Authoritative Facts
    balance = max(0.0, round(float(ctx.balance_due), 2))
    total_amount = max(0.0, round(float(ctx.total_amount), 2))
    days_overdue = int(ctx.days_overdue)
    currency = ctx.currency or "USD"
    invoice_num = ctx.invoice_number or f"INV-{ctx.invoice_id}"
    customer_name = ctx.customer_name or f"Customer {ctx.customer_id}"
    is_disputed = bool(ctx.is_disputed)

    # Aging calculation
    if days_overdue <= 0:
        aging_bucket = "CURRENT"
    elif days_overdue <= 15:
        aging_bucket = "1-15_DAYS"
    elif days_overdue <= 30:
        aging_bucket = "16-30_DAYS"
    elif days_overdue <= 60:
        aging_bucket = "31-60_DAYS"
    else:
        aging_bucket = "60+_DAYS"

    actual_facts = [
        f"Invoice #{invoice_num} has authoritative balance of {balance:.2f} {currency} out of {total_amount:.2f} {currency}.",
        f"Due date: {ctx.due_date or 'Not specified'} | Days overdue: {days_overdue} days ({aging_bucket}).",
        f"Current authoritative invoice status: {ctx.invoice_status} | Disputed: {is_disputed}.",
        f"Customer account tier: {ctx.account_tier} | Customer ID: {ctx.customer_id} ({customer_name}).",
    ]

    # 2. Predictions & Projections
    if balance <= 0.0:
        predicted_prob_collection = 1.0
        risk_score = 0.05
        risk_level = "LOW"
    elif is_disputed:
        predicted_prob_collection = 0.40
        risk_score = 0.70
        risk_level = "HIGH"
    elif days_overdue <= 0:
        predicted_prob_collection = 0.95
        risk_score = 0.15
        risk_level = "LOW"
    elif days_overdue <= 15:
        predicted_prob_collection = 0.85
        risk_score = 0.35
        risk_level = "MEDIUM"
    elif days_overdue <= 30:
        predicted_prob_collection = 0.70
        risk_score = 0.55
        risk_level = "MEDIUM"
    else:
        predicted_prob_collection = 0.45
        risk_score = 0.80
        risk_level = "HIGH"

    predictions = [
        f"Projected collection probability: {int(predicted_prob_collection * 100)}% based on commercial aging.",
        f"Receivable cash-flow risk score: {risk_score:.2f} ({risk_level}).",
        f"Estimated settlement window: {'Immediate (Paid)' if balance <= 0 else '7-14 days with standard follow-up'}.",
    ]

    # 3. Assumptions
    assumptions = [
        "Assumes primary customer billing contact is valid and operational.",
        "Assumes normal banking clearing cycle of 2-3 business days for international remittances.",
        "Assumes no unrecorded credit memos or dispute claims pending outside the system.",
    ]

    # Priority score (0-100)
    # Higher overdue days and higher balance increase priority
    raw_priority = (min(days_overdue, 90) / 90.0) * 50.0 + (min(balance, 50000.0) / 50000.0) * 35.0
    if is_disputed:
        raw_priority += 15.0
    priority_score = min(100.0, max(10.0, round(raw_priority, 1)))

    if priority_score >= 80.0:
        priority_level = "CRITICAL"
    elif priority_score >= 60.0:
        priority_level = "HIGH"
    elif priority_score >= 35.0:
        priority_level = "MEDIUM"
    else:
        priority_level = "LOW"

    # Multi-invoice consolidation check
    multi_invoice_summary = None
    if ctx.other_customer_invoices and len(ctx.other_customer_invoices) > 0:
        other_total = sum(float(inv.get("balance_due", 0.0)) for inv in ctx.other_customer_invoices)
        multi_invoice_summary = (
            f"Customer has {len(ctx.other_customer_invoices)} other active invoice(s) with total outstanding "
            f"balance of {other_total:.2f} {currency}. Consolidated statement recommended to prevent communication fatigue."
        )

    # 4. Stop Condition Detection
    stop_reason = None
    if balance <= 0.0 or ctx.invoice_status.upper() == "PAID":
        stop_reason = "Invoice has been fully settled. No collection action required."
    elif ctx.invoice_status.upper() in ["CANCELLED", "VOID", "REFUNDED"]:
        stop_reason = f"Invoice status is {ctx.invoice_status}. Collection activity stopped."

    # 5. Candidate Strategy Generation
    candidates: List[CollectionStrategyCandidate] = []

    # Strategy 1: Wait & Monitor (Appropriate if not yet overdue or fully paid)
    strat_monitor = CollectionStrategyCandidate(
        strategy_id="strat-wait-monitor",
        strategy_name="Wait & Routine Monitoring",
        description="Monitor invoice status without direct customer contact. Suitable for invoices not yet due or under standard terms.",
        recommended_action="WAIT_AND_MONITOR",
        priority_level="LOW",
        priority_score=20.0,
        urgency="ROUTINE",
        cooldown_days=5,
        requires_approval=False,
        expected_outcome="Allow customer normal payment cycle without premature contact.",
        score=0.90 if days_overdue <= 0 else 0.40,
        draft_subject=f"Notice: Invoice #{invoice_num} Schedule",
        draft_message=f"Dear {customer_name},\n\nThis is a courtesy update regarding invoice #{invoice_num} for {balance:.2f} {currency} due on {ctx.due_date}.\n\nBest regards,\nLogisticsHQ Accounts Receivable",
    )
    candidates.append(strat_monitor)

    # Strategy 2: Friendly Payment Reminder (Standard follow-up)
    strat_reminder = CollectionStrategyCandidate(
        strategy_id="strat-friendly-reminder",
        strategy_name="Friendly Payment Reminder",
        description="Polite, professional communication reminding customer of outstanding invoice and requesting payment schedule.",
        recommended_action="SEND_FRIENDLY_REMINDER",
        priority_level="MEDIUM",
        priority_score=55.0,
        urgency="NORMAL",
        cooldown_days=3,
        requires_approval=(balance >= 5000.0 or days_overdue >= 30),
        approval_reason="High balance or 30+ days aging requires finance review" if (balance >= 5000.0 or days_overdue >= 30) else None,
        expected_outcome="Prompt customer payment without escalating commercial relationship.",
        score=0.88 if (0 < days_overdue <= 15 and not is_disputed) else 0.60,
        draft_subject=f"Friendly Reminder: Outstanding Invoice #{invoice_num}",
        draft_message=(
            f"Dear {customer_name},\n\nWe hope this message finds you well. Our records show that invoice #{invoice_num} "
            f"for {balance:.2f} {currency} was due on {ctx.due_date} and is currently {days_overdue} days overdue.\n\n"
            f"Could you please confirm the payment status or share remittance details at your earliest convenience?\n\n"
            f"Thank you for your partnership.\n\nSincerely,\nLogisticsHQ Finance Department"
        ),
    )
    candidates.append(strat_reminder)

    # Strategy 3: Status & Remittance Confirmation (For 16-30 days overdue)
    strat_status = CollectionStrategyCandidate(
        strategy_id="strat-status-confirmation",
        strategy_name="Remittance Advice & Status Request",
        description="Direct request for bank wire reference or processing schedule for overdue receivable.",
        recommended_action="REQUEST_PAYMENT_STATUS",
        priority_level="HIGH",
        priority_score=70.0,
        urgency="URGENT",
        cooldown_days=3,
        requires_approval=(balance >= 5000.0 or days_overdue >= 20),
        approval_reason="High monetary threshold or aging requires oversight" if (balance >= 5000.0 or days_overdue >= 20) else None,
        expected_outcome="Obtain verified bank tracking reference or payment commitment date.",
        score=0.85 if (15 < days_overdue <= 30 and not is_disputed) else 0.55,
        draft_subject=f"Payment Status Inquiry: Invoice #{invoice_num} ({balance:.2f} {currency})",
        draft_message=(
            f"Dear {customer_name} Accounts Payable,\n\nWe are writing regarding overdue invoice #{invoice_num} "
            f"in the amount of {balance:.2f} {currency} (Due: {ctx.due_date}).\n\n"
            f"Please provide the estimated remittance date or wire confirmation so we can update our records.\n\n"
            f"Regards,\nLogisticsHQ Credit Control"
        ),
    )
    candidates.append(strat_status)

    # Strategy 4: Dispute Investigation & Resolution (If disputed)
    strat_dispute = CollectionStrategyCandidate(
        strategy_id="strat-dispute-resolution",
        strategy_name="Billing Dispute Investigation & Resolution",
        description="Pause collections and coordinate resolution of customer billing dispute. Do not pressure for immediate payment until resolved.",
        recommended_action="FOLLOW_UP_DISPUTE",
        priority_level="HIGH",
        priority_score=75.0,
        urgency="URGENT",
        cooldown_days=2,
        requires_approval=True,
        approval_reason="Disputed invoice requires finance manager review before communication",
        expected_outcome="Resolve operational or rate discrepancy and unblock customer payment.",
        score=0.95 if is_disputed else 0.20,
        draft_subject=f"Billing Dispute Resolution: Invoice #{invoice_num} ({customer_name})",
        draft_message=(
            f"Dear {customer_name},\n\nThank you for bringing your billing question regarding invoice #{invoice_num} to our attention. "
            f"Our operations and finance teams are actively reviewing the details. We will provide an updated reconciliation shortly.\n\n"
            f"Best regards,\nLogisticsHQ Customer Finance"
        ),
    )
    candidates.append(strat_dispute)

    # Strategy 5: Internal Finance Escalation (For severe overdue or large exposure)
    strat_escalation = CollectionStrategyCandidate(
        strategy_id="strat-finance-escalation",
        strategy_name="Internal Finance & Management Escalation",
        description="Escalate account to senior credit management. May involve credit limit hold or commercial review.",
        recommended_action="ESCALATE_TO_FINANCE",
        priority_level="CRITICAL",
        priority_score=90.0,
        urgency="IMMEDIATE",
        cooldown_days=1,
        requires_approval=True,
        approval_reason="Account escalation requires credit controller approval",
        expected_outcome="Internal decision on account terms, shipment holds, or structured settlement.",
        score=0.92 if (days_overdue > 30 or balance >= 25000.0) else 0.35,
        draft_subject=f"URGENT: Executive Escalation - Overdue Invoice #{invoice_num}",
        draft_message=(
            f"INTERNAL NOTICE FOR FINANCE MANAGEMENT:\n\nCustomer: {customer_name}\nInvoice: #{invoice_num}\n"
            f"Balance: {balance:.2f} {currency}\nDays Overdue: {days_overdue} days.\n"
            f"Recommended Action: Review credit facilities and initiate senior management discussion."
        ),
    )
    candidates.append(strat_escalation)

    # 6. Recommendation Selection Logic
    if balance <= 0.0:
        recommended = strat_monitor
    elif is_disputed:
        recommended = strat_dispute
    elif days_overdue > 30 or balance >= 25000.0:
        recommended = strat_escalation
    elif days_overdue > 15:
        recommended = strat_status
    elif days_overdue > 0:
        recommended = strat_reminder
    else:
        recommended = strat_monitor

    # 7. 7-Step Autonomous Execution Plan
    plan_steps = [
        {
            "step_id": "step-fin-1",
            "step_number": 1,
            "action_type": "VERIFY_INVOICE_BALANCE",
            "title": "Verify Authoritative Ledger Balance",
            "description": f"Confirm current outstanding balance of {balance:.2f} {currency} against MariaDB customer_invoices and payments.",
            "requires_approval": False,
            "expected_outcome": "Guarantees no hallucinated or stale financial balance is acted upon.",
        },
        {
            "step_id": "step-fin-2",
            "step_number": 2,
            "action_type": "CHECK_PAYMENT_STATUS_AND_DISPUTES",
            "title": "Inspect Dispute State & Recent Clearing",
            "description": "Verify whether invoice is disputed or recent unallocated payments exist.",
            "requires_approval": False,
            "expected_outcome": "Prevents inappropriate collection harassment on disputed accounts.",
        },
        {
            "step_id": "step-fin-3",
            "step_number": 3,
            "action_type": "EVALUATE_AGING_AND_PRIORITY",
            "title": "Evaluate Aging Bucket & Exposure Risk",
            "description": f"Assess aging ({aging_bucket}) and priority score ({priority_score:.1f}).",
            "requires_approval": False,
            "expected_outcome": "Calibrates urgency and communication tone.",
        },
        {
            "step_id": "step-fin-4",
            "step_number": 4,
            "action_type": "SELECT_COLLECTION_STRATEGY",
            "title": "Select Governed Collection Strategy",
            "description": f"Select {recommended.strategy_name} ({recommended.recommended_action}).",
            "requires_approval": False,
            "expected_outcome": "Strategy approved within corporate autonomy policy boundaries.",
        },
        {
            "step_id": "step-fin-5",
            "step_number": 5,
            "action_type": "DRAFT_CUSTOMER_COMMUNICATION",
            "title": "Prepare Verified Financial Communication",
            "description": "Synthesize draft message using strict authoritative facts (invoice #, balance, due date).",
            "requires_approval": False,
            "expected_outcome": "Clean, audit-ready message draft without artificial penalties.",
        },
        {
            "step_id": "step-fin-6",
            "step_number": 6,
            "action_type": "OBTAIN_FINANCE_APPROVAL",
            "title": "Enforce Human-in-the-Loop Finance Gate",
            "description": "Require finance controller approval if balance exceeds policy threshold or invoice is disputed.",
            "requires_approval": recommended.requires_approval,
            "expected_outcome": "Authoritative sign-off before customer-facing transmission.",
        },
        {
            "step_id": "step-fin-7",
            "step_number": 7,
            "action_type": "DISPATCH_ACTION_SYSTEM_FOLLOWUP",
            "title": "Execute via Go Action System Boundary",
            "description": "Dispatch action execution request to existing Action System for auditable delivery.",
            "requires_approval": False,
            "expected_outcome": "Auditable execution with immutable correlation ID tracking.",
        },
    ]

    return FinanceCollectionEvaluationResponse(
        invoice_id=ctx.invoice_id,
        invoice_number=invoice_num,
        customer_name=customer_name,
        currency=currency,
        total_amount=total_amount,
        balance_due=balance,
        days_overdue=days_overdue,
        aging_bucket=aging_bucket,
        priority_level=priority_level,
        priority_score=priority_score,
        risk_level=risk_level,
        risk_score=risk_score,
        actual_facts=actual_facts,
        predictions=predictions,
        assumptions=assumptions,
        candidate_strategies=candidates,
        recommended_strategy_id=recommended.strategy_id,
        recommended_action=recommended.recommended_action,
        draft_subject=recommended.draft_subject or f"Invoice #{invoice_num} Follow-Up",
        draft_message=recommended.draft_message or "",
        requires_approval=recommended.requires_approval,
        approval_reason=recommended.approval_reason,
        stop_reason=stop_reason,
        confidence_score=0.90 if balance > 0 else 0.98,
        data_sufficiency="COMPLETE",
        plan_steps=plan_steps,
        multi_invoice_summary=multi_invoice_summary,
        correlation_id=req.correlation_id,
    )


def replan_finance_collection(req: FinanceCollectionReplanningRequest) -> FinanceCollectionEvaluationResponse:
    """
    Replanning adaptation: adapts the active collection plan when payment events,
    customer responses, dispute events, or balance changes occur.
    """
    ctx = req.context
    payload = req.event_payload or {}
    trigger = req.trigger_event.upper()

    # Adapt context based on trigger event
    if trigger == "PAYMENT_RECEIVED":
        amount_paid = float(payload.get("amount_paid", ctx.balance_due))
        ctx.balance_due = max(0.0, round(float(ctx.balance_due) - amount_paid, 2))
        ctx.paid_amount = float(ctx.paid_amount) + amount_paid
        if ctx.balance_due <= 0.0:
            ctx.invoice_status = "PAID"
    elif trigger == "PARTIAL_PAYMENT":
        amount_paid = float(payload.get("amount_paid", 0.0))
        ctx.balance_due = max(0.0, round(float(ctx.balance_due) - amount_paid, 2))
        ctx.paid_amount = float(ctx.paid_amount) + amount_paid
        ctx.invoice_status = "PARTIALLY_PAID"
    elif trigger == "DISPUTE_OPENED":
        ctx.is_disputed = True
        ctx.dispute_reason = payload.get("dispute_reason", "Customer flagged billing discrepancy")
        ctx.invoice_status = "DISPUTED"
    elif trigger == "CUSTOMER_RESPONSE":
        response_note = payload.get("response_text", "Customer promised payment next week")
        ctx.special_notes = f"Customer response: {response_note}"

    # Re-evaluate with updated context
    eval_req = FinanceCollectionEvaluationRequest(context=ctx, correlation_id=req.correlation_id)
    return evaluate_finance_collection(eval_req)
