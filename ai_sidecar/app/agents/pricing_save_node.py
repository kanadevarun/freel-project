from app.state.pricing_state import PricingAgentState
from app.tools.pricing_tool import save_draft_quotes_tool

def pricing_save_node(state: PricingAgentState) -> dict:
    """
    Deterministic save node.
    Calls the save_draft_quotes_tool to persist suggested quotes to the Go backend.
    """
    rfq_id = state.get("rfq_id")
    org_id = state.get("org_id", 1)
    quotes = state.get("suggested_quotes", [])
    
    if not quotes:
        print("[Pricing Save Node] No suggested quotes found to save.")
        return {}
        
    print(f"[Pricing Save Node] Persisting {len(quotes)} quotes for RFQ #{rfq_id}...")
    
    # Invoke the underlying function of the LangChain tool directly
    result = save_draft_quotes_tool.func(rfq_id=rfq_id, quotes=quotes, org_id=org_id)
    print(f"[Pricing Save Node] Result: {result}")
    
    return {}
