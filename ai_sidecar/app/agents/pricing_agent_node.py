import json
import re
from typing import Any
from langchain_core.messages import SystemMessage, HumanMessage, AIMessage
from app.tools.llm_factory import get_chat_model
from app.tools.pricing_tool import get_rfq_details_tool, search_rates_tool, get_pricing_rules_tool, save_draft_quotes_tool
from app.tools.web_search import get_web_search_tool
from app.state.pricing_state import PricingAgentState

tools = [
    get_rfq_details_tool, 
    search_rates_tool, 
    get_pricing_rules_tool, 
    save_draft_quotes_tool, 
    get_web_search_tool()
]

def content_to_text(content: Any) -> str:
    """
    Normalize LangChain/Gemini message content into plain text.

    Gemini may return response.content as either:
      - a plain string
      - a list of structured content blocks
    """
    if isinstance(content, str):
        return content

    if isinstance(content, list):
        parts = []

        for item in content:
            if isinstance(item, str):
                parts.append(item)
            elif isinstance(item, dict):
                text = item.get("text")
                if text:
                    parts.append(str(text))

        return "\n".join(parts)

    return str(content)
    
llm_with_tools = get_chat_model(tools=tools)

SYSTEM_PROMPT = """You are a Senior Pricing Analyst Agent for a Freight Forwarder.
Your goal is to recommend the best quotation options (up to 3 carrier quote options) for an RFQ based on:
1. Candidate rates (contract and spot rates returned by search_rates_tool)
2. Active markup and minimum margin rules (returned by get_pricing_rules_tool)
3. Live port congestion, carrier General Rate Increases (GRIs), or seasonal surcharges (using the search tool to verify if needed).

To compute the sell price:
- Apply the appropriate markup percentage (from pricing rules) to the buy price. E.g. if markup is 20%, sell = buy * 1.20.
- Verify that the resulting margin (Sell - Buy) / Sell meets or exceeds the min_margin_pct defined in the rules.
- If multiple rules match, prioritize LANE rules over DEFAULT rules, and CUSTOMER_TIER rules. Higher priority values always apply first.

When you are ready to conclude and return the recommended options:
Your final AIMessage MUST contain a JSON block representing the suggested quotes list in the following format:
```json
[
  {
    "carrier_name": "Maersk",
    "transit_time_days": 14,
    "buy_price": 2800.00,
    "sell_price": 3360.00,
    "is_recommended": true,
    "reliability_score": 92,
    "historical_success_rate": 0.95,
    "ai_reasoning": "Standard contract rate. Safe transit duration."
  }
]
```
Also, summarize your overall reasoning outside the JSON block.
"""

def pricing_agent_node(state: PricingAgentState) -> dict:
    messages = list(state.get("messages", []))
    is_new = False
    
    # 1. Initialize messages if empty
    if not messages:
        context_prompt = f"""Process pricing for RFQ #{state['rfq_id']} in Org #{state['org_id']}.
Origin: {state.get('origin', '')}
Destination: {state.get('destination', '')}
Incoterms: {state.get('incoterms', '')}
Equipment Type: {state.get('equipment_type', '40GP')}
Weight: {state.get('gross_weight', 0.0)} KG
Volume: {state.get('volume_cbm', 0.0)} CBM
Commodity: {state.get('commodity', '')}
Target Date: {state.get('target_date', '')}

Steps to follow:
1. Call get_rfq_details_tool to load any additional info.
2. Call search_rates_tool to find candidate contracts and spot rates.
3. Call get_pricing_rules_tool to fetch matching markup parameters.
4. If candidate rates are missing or anomalies exist, perform a web search to check for active General Rate Increases (GRIs) or port congestion.
5. Compute the buy/sell options and output the final suggested quotes list in the required JSON block.
"""
        messages = [
            SystemMessage(content=SYSTEM_PROMPT),
            HumanMessage(content=context_prompt)
        ]
        is_new = True
        
    # 2. Invoke LLM with current messages
    response = llm_with_tools.invoke(messages)
    
    if is_new:
        # Prepend the system and human messages to the updates list
        updates = {"messages": messages + [response]}
    else:
        updates = {"messages": [response]}
    
    if not response.tool_calls:
        # Extract json block
        text = content_to_text(response.content)
        json_match = re.search(
            r"```json\s*(.*?)\s*```",
            text,
            re.DOTALL,
        )
        if json_match:
            try:
                quotes = json.loads(json_match.group(1).strip())
                updates["suggested_quotes"] = quotes
                updates["overall_reasoning"] = text.replace(json_match.group(0), "").strip()
            except Exception as e:
                print(f"[Pricing Agent] Error parsing JSON from final content: {e}")
                
    return updates
