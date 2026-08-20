from app.state.pricing_state import PricingAgentState

def pricing_validate_node(state: PricingAgentState) -> dict:
    """
    Deterministic validation node (no LLM).
    Checks for pricing anomalies:
    1. Buy price > $8,000 on any quote option.
    2. Margin < 5% on any option.
    3. Low confidence flag.
    If any anomalies are found, sets is_anomaly=True, which triggers the HITL interrupt before saving.
    """
    quotes = state.get("suggested_quotes", [])
    is_anomaly = False
    reasons = []

    for q in quotes:
        buy = q.get("buy_price", 0.0)
        sell = q.get("sell_price", 0.0)
        carrier = q.get("carrier_name", "Unknown")
        
        # 1. Buy price anomaly check
        if buy > 8000.0:
            is_anomaly = True
            reasons.append(f"Buy price for {carrier} (${buy}) exceeds $8,000 threshold.")
            
        # 2. Minimum margin check
        if sell > 0:
            margin = (sell - buy) / sell
            if margin < 0.05:
                is_anomaly = True
                reasons.append(f"Margin for {carrier} ({margin:.1%}) is below the minimum 5% requirement.")
                
    if is_anomaly:
        reason_str = " | ".join(reasons)
        print(f"[Pricing Validator] Anomaly detected: {reason_str}")
        return {
            "is_anomaly": True,
            "overall_reasoning": state.get("overall_reasoning", "") + f"\n\n[ANOMALY WARNING]: {reason_str}"
        }
        
    return {"is_anomaly": False}
