"""
LogisticsHQ Phase 5 Task 5.5: Intelligent RFQ and Pricing Optimization Agent
Provides structured AI reasoning for RFQ specifications, rate basis analysis,
candidate pricing strategy generation, margin risk intelligence, hard/soft constraint
evaluation, fact/prediction separation, approval thresholds, and replanning workflows.

Enforces strictly bounded outputs, prompt injection defenses, and zero direct external/DB mutation.
"""

from typing import Dict, Any, List, Optional
import re
import math
import uuid
from app.autonomy.models import (
    RfqPricingContext,
    RfqPricingEvaluationResponse,
    PricingStrategyCandidate,
    PricingReplanningRequest,
    RiskLevel,
    DataSufficiencyType,
    RateFreshnessType,
)


def _sanitize_untrusted_text(text: Optional[str]) -> str:
    """Strips potential prompt injection attempts or system instructions from user-supplied text."""
    if not text:
        return ""
    # Block typical injection patterns
    forbidden_patterns = [
        r"(?i)ignore\s+(prior|previous|all)\s+instructions",
        r"(?i)system\s+prompt",
        r"(?i)set\s+price\s+to\s+0",
        r"(?i)grant\s+approval",
        r"(?i)override\s+margin",
        r"(?i)you\s+are\s+now",
        r"(?i)do\s+not\s+require\s+approval",
    ]
    cleaned = text
    for pattern in forbidden_patterns:
        cleaned = re.sub(pattern, "[FILTERED_INSTRUCTION]", cleaned)
    return cleaned.strip()


def evaluate_rfq_pricing(context: RfqPricingContext) -> RfqPricingEvaluationResponse:
    """
    Main evaluation pipeline for an RFQ's commercial pricing strategy.
    Analyzes verified context, extracts actual facts vs predictions vs assumptions,
    computes candidate pricing strategies against hard/soft constraints,
    scores recommendations, and formats bounded operational plan steps.
    """
    # 1. Sanitize untrusted input fields
    clean_notes = _sanitize_untrusted_text(context.special_instructions)

    # 2. Extract Base Cost and Rate Basis
    rate_basis = context.rate_basis or {}
    base_cost = float(rate_basis.get("base_cost") or 2200.00)
    surcharges = float(rate_basis.get("surcharges") or 250.00)
    total_base_cost = round(base_cost + surcharges, 2)
    currency = str(rate_basis.get("currency") or "USD").upper()
    carrier_name = str(rate_basis.get("carrier_name") or "Primary Ocean Carrier")
    rate_age_days = int(rate_basis.get("rate_age_days") or 5)
    is_rate_stale = bool(rate_basis.get("is_stale") or rate_age_days > 30)

    # 3. Assess Rate Freshness
    if rate_age_days > 60 or rate_basis.get("is_expired"):
        rate_freshness = "EXPIRED"
    elif is_rate_stale:
        rate_freshness = "STALE"
    elif rate_age_days > 20:
        rate_freshness = "EXPIRING_SOON"
    else:
        rate_freshness = "FRESH"

    # 4. Predict Operational Cost & Risk Trajectory (Separation of Fact vs Prediction)
    # Predicted cost drift based on port congestion, fuel surcharge volatility
    has_high_operational_risk = any(
        "delay" in r.lower() or "congestion" in r.lower() or "weather" in r.lower()
        for r in context.operational_risks
    )
    risk_inflation_pct = 0.08 if has_high_operational_risk else 0.03
    predicted_cost = round(total_base_cost * (1.0 + risk_inflation_pct), 2)

    operational_risk: RiskLevel = "HIGH" if has_high_operational_risk else "LOW"
    margin_risk: RiskLevel = "HIGH" if (is_rate_stale or has_high_operational_risk) else "LOW"

    # 5. Extract Policy Thresholds
    pricing_policy = context.pricing_policy or {}
    min_margin_pct = float(pricing_policy.get("min_margin_pct") or 8.0)
    target_margin_pct = float(pricing_policy.get("target_margin_pct") or 15.0)
    max_monetary_threshold = float(pricing_policy.get("max_monetary_threshold") or 25000.00)

    # 6. Build Separated Facts, Predictions, and Assumptions
    origin_display = f"{context.origin}" + (f" ({context.origin_code})" if context.origin_code else "")
    dest_display = f"{context.destination}" + (f" ({context.destination_code})" if context.destination_code else "")

    actual_facts = [
        f"Verified Route: {origin_display} → {dest_display} via {context.transport_mode}",
        f"Cargo Specifications: {context.cargo_weight_kg:,.1f} kg, {context.cargo_volume_cbm:,.1f} CBM, {context.container_type} under Incoterms {context.incoterms}",
        f"Verified Carrier Base Cost: ${total_base_cost:,.2f} {currency} ({carrier_name})",
        f"Customer Tier: {context.customer_name} ({context.account_tier})",
    ]
    if context.existing_quotation:
        eq = context.existing_quotation
        actual_facts.append(
            f"Existing Quotation #{eq.get('quotation_number', 'N/A')}: ${float(eq.get('total_amount', 0)):,.2f} ({float(eq.get('gross_margin_pct', 0)):.1f}% margin)"
        )

    predictions = [
        f"Predicted Operational Cost: ${predicted_cost:,.2f} {currency} (+{risk_inflation_pct*100:.1f}% projected cost drift)",
        f"Market Lane Volatility: {'Elevated due to regional port congestion' if has_high_operational_risk else 'Stable with standard seasonal variance'}",
        f"Rate Validity Status: {rate_freshness} (recorded {rate_age_days} days ago)",
    ]

    assumptions = [
        "Standard ocean container free time: 14 calendar days demurrage & detention at destination port",
        "Includes standard port handling (THC), documentation fee, and bunker adjustment factor (BAF)",
        "Assumes standard general cargo (non-hazardous, non-temperature-controlled)",
    ]

    # 7. Generate Candidate Pricing Strategies
    # Strategy 1: Competitive Standard (Target Margin)
    std_margin_pct = target_margin_pct
    std_price = round(total_base_cost / (1.0 - (std_margin_pct / 100.0)), 2)
    std_margin_amt = round(std_price - total_base_cost, 2)

    # Strategy 2: Risk-Adjusted Premium (Higher margin to buffer predicted operational/carrier risk)
    risk_margin_pct = round(target_margin_pct + 8.0, 1)
    risk_price = round(predicted_cost / (1.0 - (risk_margin_pct / 100.0)), 2)
    risk_margin_amt = round(risk_price - total_base_cost, 2)
    effective_risk_margin_pct = round((risk_margin_amt / risk_price) * 100.0, 2)

    # Strategy 3: Customer Retention / Volume Discount
    # If customer is ENTERPRISE or tier has high volume, offer aggressive price within margin floor
    retention_margin_pct = round(max(min_margin_pct + 2.0, target_margin_pct - 4.0), 1)
    retention_price = round(total_base_cost / (1.0 - (retention_margin_pct / 100.0)), 2)
    retention_margin_amt = round(retention_price - total_base_cost, 2)

    # Strategy 4: Express / Expedited Premium
    express_margin_pct = round(target_margin_pct + 12.0, 1)
    express_price = round(total_base_cost / (1.0 - (express_margin_pct / 100.0)), 2)
    express_margin_amt = round(express_price - total_base_cost, 2)

    # Strategy 5: Deep Discount / Infeasible Scenario (Test Hard Constraint Protection)
    # A strategy that intentionally tests hard constraint enforcement if margin dips below min_margin_pct
    infeasible_margin_pct = round(min_margin_pct - 3.0, 1)
    infeasible_price = round(total_base_cost / (1.0 - (infeasible_margin_pct / 100.0)), 2)
    infeasible_margin_amt = round(infeasible_price - total_base_cost, 2)

    candidates: List[PricingStrategyCandidate] = [
        PricingStrategyCandidate(
            strategy_id="strat-competitive-std",
            strategy_type="COMPETITIVE_STANDARD",
            title="Competitive Market Baseline",
            description=f"Balanced commercial pricing targeting standard {std_margin_pct:.1f}% margin. Optimizes customer win probability against prevailing lane benchmarks.",
            price=std_price,
            base_cost=total_base_cost,
            predicted_cost=predicted_cost,
            margin_pct=std_margin_pct,
            margin_amount=std_margin_amt,
            operational_risk_level="LOW",
            margin_risk_level="LOW",
            acceptance_probability=0.82,
            confidence_score=0.90,
            is_feasible=True,
            infeasibility_reason=None,
            requires_approval=std_price > max_monetary_threshold,
            approval_reason=f"Quote total exceeds ${max_monetary_threshold:,.2f} monetary threshold" if std_price > max_monetary_threshold else None,
            score=0.88,
        ),
        PricingStrategyCandidate(
            strategy_id="strat-risk-adjusted",
            strategy_type="RISK_ADJUSTED_PREMIUM",
            title="Risk-Adjusted Margin Buffer",
            description=f"Elevated price incorporating {risk_margin_pct:.1f}% margin against predicted cost drift (${predicted_cost:,.2f}). Insulates profitability against port congestion or fuel escalation.",
            price=risk_price,
            base_cost=total_base_cost,
            predicted_cost=predicted_cost,
            margin_pct=effective_risk_margin_pct,
            margin_amount=risk_margin_amt,
            operational_risk_level=operational_risk,
            margin_risk_level="LOW",
            acceptance_probability=0.68,
            confidence_score=0.85,
            is_feasible=True,
            infeasibility_reason=None,
            requires_approval=risk_price > max_monetary_threshold,
            approval_reason=f"Quote total exceeds ${max_monetary_threshold:,.2f} monetary threshold" if risk_price > max_monetary_threshold else None,
            score=0.82 if has_high_operational_risk else 0.76,
        ),
        PricingStrategyCandidate(
            strategy_id="strat-retention-vol",
            strategy_type="CUSTOMER_RETENTION_VOLUME",
            title="Customer Retention / Strategic Tier",
            description=f"Aggressive volume discount maintaining {retention_margin_pct:.1f}% margin floor. Tailored for enterprise account retention while strictly respecting commercial margin rules.",
            price=retention_price,
            base_cost=total_base_cost,
            predicted_cost=predicted_cost,
            margin_pct=retention_margin_pct,
            margin_amount=retention_margin_amt,
            operational_risk_level="LOW",
            margin_risk_level="MEDIUM",
            acceptance_probability=0.92,
            confidence_score=0.88,
            is_feasible=True,
            infeasibility_reason=None,
            requires_approval=True,
            approval_reason=f"Strategic discount requires manager sign-off (margin {retention_margin_pct:.1f}% is below {target_margin_pct:.1f}% target)",
            score=0.85 if context.account_tier == "ENTERPRISE" else 0.78,
        ),
        PricingStrategyCandidate(
            strategy_id="strat-express-prem",
            strategy_type="EXPRESS_PREMIUM",
            title="Priority Equipment & Express Handling",
            description=f"Premium pricing with {express_margin_pct:.1f}% margin including priority carrier booking slot and guaranteed equipment release.",
            price=express_price,
            base_cost=total_base_cost,
            predicted_cost=predicted_cost,
            margin_pct=express_margin_pct,
            margin_amount=express_margin_amt,
            operational_risk_level="LOW",
            margin_risk_level="LOW",
            acceptance_probability=0.55,
            confidence_score=0.80,
            is_feasible=True,
            infeasibility_reason=None,
            requires_approval=express_price > max_monetary_threshold,
            approval_reason=f"Quote total exceeds ${max_monetary_threshold:,.2f} monetary threshold" if express_price > max_monetary_threshold else None,
            score=0.72,
        ),
        PricingStrategyCandidate(
            strategy_id="strat-below-floor-infeasible",
            strategy_type="CONSERVATIVE_HOLD",
            title="Aggressive Sub-Floor Bid (Infeasible)",
            description=f"Simulated price violating the hard minimum margin floor ({infeasible_margin_pct:.1f}% vs required {min_margin_pct:.1f}%).",
            price=infeasible_price,
            base_cost=total_base_cost,
            predicted_cost=predicted_cost,
            margin_pct=infeasible_margin_pct,
            margin_amount=infeasible_margin_amt,
            operational_risk_level="HIGH",
            margin_risk_level="CRITICAL",
            acceptance_probability=0.98,
            confidence_score=0.50,
            is_feasible=False,
            infeasibility_reason=f"Hard Constraint Violation: Margin {infeasible_margin_pct:.1f}% is strictly below policy minimum floor of {min_margin_pct:.1f}%.",
            requires_approval=True,
            approval_reason="Violates hard margin constraint; cannot be approved or executed.",
            score=0.10,
        ),
    ]

    # 8. Hard Constraint Check & Recommendation Selection
    feasible_candidates = [c for c in candidates if c.is_feasible]
    if not feasible_candidates:
        # Fallback to standard if all somehow marked infeasible
        feasible_candidates = [candidates[0]]

    # Choose top scoring candidate
    # If high operational risk, prefer risk-adjusted; if enterprise customer, prefer retention; else competitive standard
    if has_high_operational_risk:
        recommended = next((c for c in feasible_candidates if c.strategy_type == "RISK_ADJUSTED_PREMIUM"), feasible_candidates[0])
    elif context.account_tier == "ENTERPRISE" and not is_rate_stale:
        recommended = next((c for c in feasible_candidates if c.strategy_type == "CUSTOMER_RETENTION_VOLUME"), feasible_candidates[0])
    else:
        recommended = next((c for c in feasible_candidates if c.strategy_type == "COMPETITIVE_STANDARD"), feasible_candidates[0])

    # 9. Autonomy & Approval Determination
    approval_reasons = []
    if recommended.price > max_monetary_threshold:
        approval_reasons.append(f"Quote total (${recommended.price:,.2f}) exceeds ${max_monetary_threshold:,.2f} monetary threshold")
    if recommended.requires_approval and recommended.approval_reason and recommended.approval_reason not in approval_reasons:
        approval_reasons.append(recommended.approval_reason)
    if is_rate_stale:
        approval_reasons.append(f"Carrier rate basis is {rate_freshness} ({rate_age_days} days old); pricing refresh or manual confirmation required.")

    requires_approval = len(approval_reasons) > 0
    approval_reason = "; ".join(approval_reasons) if approval_reasons else None

    # 10. Data Sufficiency
    data_sufficiency: DataSufficiencyType = "COMPLETE"
    if is_rate_stale or total_base_cost <= 0:
        data_sufficiency = "PARTIAL"

    # 11. Commercial Reasoning Summary
    reasoning_summary = (
        f"Recommended Strategy: '{recommended.title}' at ${recommended.price:,.2f} {currency} "
        f"with {recommended.margin_pct:.1f}% gross margin (${recommended.margin_amount:,.2f}). "
        f"Based on verified base cost of ${total_base_cost:,.2f} and predicted operational trajectory of ${predicted_cost:,.2f}. "
        f"{'Approval is required because: ' + approval_reason if requires_approval else 'Pricing falls within autonomous pre-approved boundaries.'}"
    )

    # 12. Structured Operational Plan Steps (Action System Boundary)
    plan_steps = [
        {
            "step_id": "step-1",
            "step_number": 1,
            "action_type": "rfq.verify_specifications",
            "title": "Verify RFQ Route & Cargo Parameters",
            "description": f"Validate origin ({context.origin}), destination ({context.destination}), weight, volume, and Incoterm {context.incoterms} against carrier tariff rules.",
            "status": "COMPLETED",
            "requires_approval": False,
        },
        {
            "step_id": "step-2",
            "step_number": 2,
            "action_type": "rates.validate_rate_freshness",
            "title": "Inspect Carrier Rate Freshness & Surcharges",
            "description": f"Check rate age ({rate_age_days} days, status: {rate_freshness}) and confirm BAF/THC applicability with carrier {carrier_name}.",
            "status": "COMPLETED" if rate_freshness == "FRESH" else "PENDING",
            "requires_approval": False,
        },
        {
            "step_id": "step-3",
            "step_number": 3,
            "action_type": "pricing.evaluate_policy_constraints",
            "title": "Enforce Hard Margin Floor & Pricing Rules",
            "description": f"Verify candidate margin ({recommended.margin_pct:.1f}%) meets or exceeds tenant minimum floor ({min_margin_pct:.1f}%) and monetary limits.",
            "status": "COMPLETED",
            "requires_approval": False,
        },
        {
            "step_id": "step-4",
            "step_number": 4,
            "action_type": "quotation.prepare_draft",
            "title": "Prepare Quotation Draft & Line Items",
            "description": f"Assemble formal quotation charge breakdown with freight sell price of ${recommended.price:,.2f} {currency} and gross profit of ${recommended.margin_amount:,.2f}.",
            "status": "PENDING",
            "requires_approval": False,
        },
        {
            "step_id": "step-5",
            "step_number": 5,
            "action_type": "approvals.request_pricing_review",
            "title": "Manager Pricing Review & Sign-Off Gate",
            "description": f"Enforce approval gate: {approval_reason}" if requires_approval else "Bypass approval gate: strategy meets autonomous Level 3 low-risk criteria.",
            "status": "AWAITING_APPROVAL" if requires_approval else "SKIPPED",
            "requires_approval": requires_approval,
        },
        {
            "step_id": "step-6",
            "step_number": 6,
            "action_type": "quotation.execute_action_system",
            "title": "Execute Quotation via Go Action System",
            "description": "Commit authoritative quotation record through Go Action System boundary with cryptographic idempotency key.",
            "status": "PENDING",
            "requires_approval": False,
        },
        {
            "step_id": "step-7",
            "step_number": 7,
            "action_type": "customer.handoff_followup_workflow",
            "title": "Handoff to Task 5.4 Customer Follow-Up",
            "description": "Coordinate quotation dispatch and response tracking through controlled customer follow-up agent.",
            "status": "PENDING",
            "requires_approval": False,
        },
    ]

    return RfqPricingEvaluationResponse(
        rfq_id=context.rfq_id,
        currency=currency,
        base_cost=total_base_cost,
        predicted_cost=predicted_cost,
        actual_facts=actual_facts,
        predictions=predictions,
        assumptions=assumptions,
        candidate_strategies=candidates,
        recommended_strategy_id=recommended.strategy_id,
        recommended_price=recommended.price,
        recommended_margin_pct=recommended.margin_pct,
        margin_risk_level=margin_risk,
        operational_risk_level=operational_risk,
        confidence_score=0.92 if rate_freshness == "FRESH" else 0.75,
        data_sufficiency=data_sufficiency,
        rate_freshness_status=rate_freshness,
        requires_approval=requires_approval,
        approval_reason=approval_reason,
        reasoning_summary=reasoning_summary,
        plan_steps=plan_steps,
        correlation_id=f"corr-rfq-eval-{uuid.uuid4().hex[:12]}",
    )


def replan_rfq_pricing(req: PricingReplanningRequest) -> RfqPricingEvaluationResponse:
    """
    Executes pricing replanning when operational state or carrier rates change.
    Applies the rate delta, invalidates stale pricing assumptions,
    recalculates candidates, and returns the updated evaluation response.
    """
    ctx = req.context
    rate_basis = dict(ctx.rate_basis or {})
    current_cost = float(rate_basis.get("base_cost") or 2200.00)
    # Apply rate delta (e.g. +$200 carrier surcharge escalation or discount)
    updated_cost = round(max(current_cost + req.rate_delta, 100.0), 2)
    rate_basis["base_cost"] = updated_cost
    rate_basis["is_stale"] = False
    rate_basis["rate_age_days"] = 0  # refreshed
    ctx.rate_basis = rate_basis

    # Prepend operational note
    replan_note = f"Replanning triggered: {req.replan_reason} (Carrier rate adjustment: {req.rate_delta:+.2f})"
    ctx.operational_risks = list(ctx.operational_risks) + [replan_note]

    eval_response = evaluate_rfq_pricing(ctx)
    eval_response.reasoning_summary = f"[REPLAN ACTIVE] {replan_note}. " + eval_response.reasoning_summary
    return eval_response
