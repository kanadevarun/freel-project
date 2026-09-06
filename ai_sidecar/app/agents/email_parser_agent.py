import os
import re
import json
import time
import httpx
from typing import Dict, Any, Optional
from app.state.sales_state import SalesAgentState
from app.agents.llm_utils import execute_llm_json, execute_llm_text
from app.tools.web_search import get_web_search_tool
from app.tools.leads_tool import create_rfq_from_email_tool

# Backend endpoint configurations
go_backend_url = os.getenv("GO_BACKEND_URL", "http://localhost:8080")
service_token = os.getenv("INTERNAL_SERVICE_TOKEN", "internal-service-key-logisticshq")

def resolve_locode(query: str) -> Optional[str]:
    """Helper to call standard UN/LOCODE normalizer in Go backend."""
    if not query:
        return None
    # If it's already a standard uppercase 5-letter code, return it.
    if len(query) == 5 and query.isupper() and query.isalpha():
        return query
        
    url = f"{go_backend_url}/internal/ports/normalize"
    headers = {"X-LogisticsHQ-Service-Key": service_token}
    params = {"query": query}
    try:
        resp = httpx.get(url, params=params, headers=headers, timeout=5.0)
        if resp.status_code == 200:
            data = resp.json()
            if data.get("normalized"):
                return data.get("normalized")
    except Exception as e:
        print(f"[AI Sidecar] Failed to resolve locode for '{query}': {e}")
    return None

def classify_and_parse_email_node(state: SalesAgentState) -> Dict[str, Any]:
    """Parses incoming shipper email content and extracts structural shipping requirements.
    
    When is_reply=True, the prior_rfq_context (structured fields only — NOT raw emails) is
    injected into the prompt so the LLM understands what is already known and focuses on
    extracting only the newly provided information.
    """
    print(f"[Sales Agent] Node: classify_and_parse_email_node starting...")
    body = state.get("email_body", "")
    subject = state.get("email_subject", "")
    sender = state.get("from_email", "")
    is_reply = state.get("is_reply", False)
    prior_context = state.get("prior_rfq_context") or {}

    today_str = time.strftime("%Y-%m-%d", time.localtime())

    # Build prior-context section for the prompt — structured JSON only, no raw emails
    prior_context_section = ""
    if is_reply and prior_context:
        known_fields = {k: v for k, v in prior_context.items() if v is not None}
        if known_fields:
            prior_context_section = f"""
IMPORTANT CONTEXT — This email is a REPLY to a previous incomplete quote request.
The following fields were already extracted from the earlier email in this conversation:
{json.dumps(known_fields, indent=2)}

Focus on extracting only the fields NOT yet known. Do NOT re-extract known fields unless
the customer explicitly changes them. If the customer provides a value for a known field,
use the new value.
"""

    prompt = f"""
    Analyze the following inbound email from a shipper and extract key metadata:
    
    Email From: {sender}
    Subject: {subject}
    Body:
    {body}
    {prior_context_section}
    Extract the following fields and return ONLY a JSON object:
    - intent: "RFQ_REQUEST" (if requesting a shipping quote), "QUESTION" (general inquiry), "MEETING" (scheduling), "UNSUBSCRIBE" (opt-out), "FOLLOW_UP" (reply to existing thread).
    - sentiment: "POSITIVE", "NEUTRAL", "NEGATIVE".
    - confidence: integer from 0 to 100 representing your confidence in intent classification.
    - lead_name: The sender's name if signed or mentioned, or null.
    - company_domain: The domain of the sender (e.g. extract from from_email like 'tataexports.com', but exclude generic domains like 'gmail.com', 'yahoo.com', 'outlook.com', etc.).
    - origin_port: The name of the origin port or city mentioned (e.g., "Nhava Sheva", "Mumbai", "INNSA"), or null.
    - destination_port: The name of the destination port or city mentioned (e.g., "Hamburg", "DEHAM"), or null.
    - incoterms: The Incoterms code mentioned (e.g. FOB, CIF, EXW, FCA), or null.
    - cargo_description: Text description of the commodity/goods, or null.
    - cargo_weight: Weight in KG (as float), or null.
    - cargo_volume: Volume in CBM (as float), or null.
    - target_date: Cargo ready date if mentioned (formatted as YYYY-MM-DD). If a relative date is mentioned (e.g., "next month", "in 2 weeks"), calculate the approximate YYYY-MM-DD date based on today's date ({today_str}) and use that, or null if completely unspecified.
    - ai_summary: A short, concise summary (1-2 sentences) of the email's request.
    
    Response JSON:
    """

    res = execute_llm_json(prompt)
    if not res:
        res = {
            "intent": "QUESTION",
            "sentiment": "NEUTRAL",
            "confidence": 50,
            "ai_summary": "Failed to parse email request due to LLM error."
        }

    raw_origin = res.get("origin_port")
    raw_destination = res.get("destination_port")
    
    origin_code = resolve_locode(raw_origin) if raw_origin else None
    dest_code = resolve_locode(raw_destination) if raw_destination else None

    updates = {
        "intent": res.get("intent", "QUESTION"),
        "sentiment": res.get("sentiment", "NEUTRAL"),
        "confidence_score": int(res.get("confidence", 80)),
        "lead_name": res.get("lead_name"),
        "company_domain": res.get("company_domain"),
        "origin_port": origin_code or raw_origin,
        "destination_port": dest_code or raw_destination,
        "incoterms": res.get("incoterms"),
        "cargo_description": res.get("cargo_description"),
        "cargo_weight": res.get("cargo_weight"),
        "cargo_volume": res.get("cargo_volume"),
        "target_date": res.get("target_date"),
        "ai_summary": res.get("ai_summary", "Analyzed inbound email.")
    }
    
    print(f"[Sales Agent] Parsing result: {updates}")
    return updates

def merge_context_node(state: SalesAgentState) -> Dict[str, Any]:
    """Merges prior conversation context with newly extracted fields from a reply email.
    
    This node fires only when is_reply=True. It takes the structured prior_rfq_context
    (cumulative extracted fields from previous turns) and fills in any fields that the
    classify_and_parse_email_node did not extract from the current email.
    
    Decision rule: current turn values always win; prior context fills gaps only.
    This ensures the customer can correct previously stated values.
    """
    print(f"[Sales Agent] Node: merge_context_node starting...")
    
    if not state.get("is_reply"):
        print(f"[Sales Agent] merge_context_node: not a reply, nothing to merge.")
        return {}
    
    prior = state.get("prior_rfq_context") or {}
    if not prior:
        print(f"[Sales Agent] merge_context_node: no prior context available.")
        return {}

    merged = {}
    rfq_fields = [
        "origin_port", "destination_port", "incoterms",
        "cargo_description", "cargo_weight", "cargo_volume",
        "target_date", "lead_name"
    ]
    
    for field in rfq_fields:
        current_val = state.get(field)
        prior_val = prior.get(field)
        # Fill from prior only if current turn didn't extract this field
        if (current_val is None or current_val == "") and prior_val is not None:
            merged[field] = prior_val
            print(f"[Sales Agent] merge_context_node: restored '{field}' = {prior_val!r} from prior context")

    if merged:
        print(f"[Sales Agent] merge_context_node: merged {len(merged)} field(s) from prior context into current state.")
    else:
        print(f"[Sales Agent] merge_context_node: all fields already present in current email, no merge needed.")
    
    return merged

def lead_research_node(state: SalesAgentState) -> Dict[str, Any]:
    """Runs a Tavily web search on the lead's company domain to gather context about scale and products."""
    print(f"[Sales Agent] Node: lead_research_node starting...")
    domain = state.get("company_domain")
    if not domain or domain in ["gmail.com", "yahoo.com", "outlook.com"]:
        return {"company_enrichment": "No commercial company domain available for lookup."}

    search_tool = get_web_search_tool()
    query = f"{domain} company overview scale products cargo shipping"
    print(f"[Sales Agent] Searching web for lead enrichment info: {query}")
    try:
        search_results = search_tool.invoke(query)
        results_str = str(search_results)
        
        prompt = f"""
        Below is search results about the company '{domain}'. Summarize the key facts about this company:
        1. What do they manufacture, trade, or export?
        2. What is their general industry type and scale?
        
        Search Results:
        {results_str[:4000]}
        
        Return a concise summary paragraph (no markdown headers, plain text).
        """
        summary = execute_llm_text(prompt)
        print(f"[Sales Agent] Lead enrichment report: {summary}")
        return {"company_enrichment": summary}
    except Exception as e:
        print(f"[Sales Agent] Web search research failed: {e}")
        return {"company_enrichment": f"Web search research failed: {str(e)}"}

def check_completeness_node(state: SalesAgentState) -> Dict[str, Any]:
    """Validates that all mandatory fields are present in the parsed email.
    
    By the time this node runs, merge_context_node has already filled any gaps
    from prior conversation turns, so this check evaluates the full merged state.
    """
    print(f"[Sales Agent] Node: check_completeness_node starting...")
    origin = state.get("origin_port")
    destination = state.get("destination_port")
    incoterms = state.get("incoterms")
    target_date = state.get("target_date")
    cargo_desc = state.get("cargo_description")
    cargo_weight = state.get("cargo_weight")
    cargo_volume = state.get("cargo_volume")

    missing_fields = []
    if not origin:
        missing_fields.append("Origin Port")
    if not destination:
        missing_fields.append("Destination Port")
    if not incoterms:
        missing_fields.append("Incoterms")
    if not cargo_desc:
        missing_fields.append("Cargo Description")
    if cargo_weight is None or str(cargo_weight).strip() == "":
        missing_fields.append("Cargo Weight")
    if cargo_volume is None or str(cargo_volume).strip() == "":
        missing_fields.append("Cargo Volume")
    if not target_date:
        missing_fields.append("Target Date")

    if missing_fields:
        print(f"[Sales Agent] Missing mandatory fields for RFQ: {missing_fields}")
        is_reply = state.get("is_reply", False)
        summary_msg = f"Incomplete Quote Request from {origin or 'unknown'} to {destination or 'unknown'}. Missing mandatory fields: {', '.join(missing_fields)}."
        
        # Call LLM to draft a polite response email asking the customer for the missing fields
        lead_name = state.get("lead_name")
        greeting = f"Dear {lead_name}" if lead_name else "Dear Customer"
        
        # Tailor the prompt slightly for follow-up replies
        reply_context = ""
        if is_reply:
            reply_context = "Note: The customer has already replied once with some information. Acknowledge the information they provided and politely ask only for what is still missing."
        
        email_prompt = f"""
        Draft a polite, professional, and concise email reply to a customer who sent a quote request but missed some mandatory details.
        
        Original Email Subject: {state.get("email_subject", "")}
        Original Email Body:
        {state.get("email_body", "")}
        
        Today's Date: {time.strftime("%Y-%m-%d", time.localtime())}
        Missing Mandatory Fields that you MUST request: {', '.join(missing_fields)}
        {reply_context}
        Write only the email body. Do not include subject line or header fields. Start with the greeting "{greeting}" and sign off professionally as "LogisticsHQ Sales Team".
        """
        drafted_email = execute_llm_text(email_prompt)
        print(f"[Sales Agent] Drafted email reply for missing info: {drafted_email}")
        
        return {
            "intent": "RFQ_REQUEST_INCOMPLETE",
            "ai_summary": summary_msg,
            "drafted_reply": drafted_email
        }

    print(f"[Sales Agent] All mandatory fields are present.")
    return {"intent": "RFQ_REQUEST"}

def draft_rfq_node(state: SalesAgentState) -> Dict[str, Any]:
    """Drafts a standard RFQ in the Go backend. Assumes all fields have been verified by completeness checker."""
    print(f"[Sales Agent] Node: draft_rfq_node starting...")
    org_id = state.get("org_id", 5)
    lead_id = state.get("lead_id")
    origin = state.get("origin_port")
    destination = state.get("destination_port")
    incoterms = state.get("incoterms")
    target_date = state.get("target_date")
    cargo_desc = state.get("cargo_description")
    cargo_weight = state.get("cargo_weight")
    cargo_volume = state.get("cargo_volume")

    items = [
        {
            "description": cargo_desc,
            "quantity": 1,
            "weight_kg": float(cargo_weight),
            "volume_cbm": float(cargo_volume)
        }
    ]

    print(f"[Sales Agent] Invoking create_rfq_from_email_tool: origin={origin}, dest={destination}, incoterms={incoterms}, date={target_date}")
    try:
        res_str = create_rfq_from_email_tool.func(
            org_id=int(org_id),
            customer_id=int(lead_id),
            origin=origin,
            destination=destination,
            incoterms=incoterms,
            target_date=target_date,
            items=items
        )
        print(f"[Sales Agent] creation tool response: {res_str}")
        
        if "rfq_id" in res_str:
            data = json.loads(res_str)
            rfq_data = data.get("data", {})
            rfq_id = rfq_data.get("rfq_id")
            print(f"[Sales Agent] Successfully drafted RFQ #{rfq_id}")
            return {"linked_rfq_id": rfq_id}
            
    except Exception as e:
        print(f"[Sales Agent] Failed to draft RFQ: {e}")
        
    return {"linked_rfq_id": None}

def save_and_callback_node(state: SalesAgentState) -> Dict[str, Any]:
    """Posts back lead classification and RFQ link details to Go callback webhook.
    
    When the intent is RFQ_REQUEST_INCOMPLETE, this node also sends partial_rfq_context
    — a structured JSON of all fields extracted so far (across all turns) — back to Go
    so it can be persisted and passed to the agent on the customer's next reply.
    """
    print(f"[Sales Agent] Node: save_and_callback_node starting...")
    org_id = state.get("org_id", 5)
    interaction_id = state.get("interaction_id")
    lead_id = state.get("lead_id")
    intent = state.get("intent", "QUESTION")
    
    payload = {
        "interaction_id": int(interaction_id),
        "org_id": int(org_id),
        "lead_id": int(lead_id),
        "sentiment": state.get("sentiment", "NEUTRAL"),
        "intent": intent,
        "confidence": int(state.get("confidence_score", 100)),
        "linked_rfq_id": state.get("linked_rfq_id"),
        "summary": state.get("ai_summary", ""),
        "drafted_reply": state.get("drafted_reply")
    }

    # When info is still missing, persist the cumulative context (structured fields only)
    # so the next reply can restore it without re-reading raw emails.
    if intent in ["RFQ_REQUEST", "RFQ_REQUEST_INCOMPLETE"]:
        partial_context = {
            "origin_port": state.get("origin_port"),
            "destination_port": state.get("destination_port"),
            "incoterms": state.get("incoterms"),
            "cargo_description": state.get("cargo_description"),
            "cargo_weight": state.get("cargo_weight"),
            "cargo_volume": state.get("cargo_volume"),
            "target_date": state.get("target_date"),
            "lead_name": state.get("lead_name"),
        }
        # Only include fields that have a value (don't save None entries)
        payload["partial_rfq_context"] = {k: v for k, v in partial_context.items() if v is not None}
        print(f"[Sales Agent] Sending partial_rfq_context for next reply: {payload['partial_rfq_context']}")

    callback_url = state.get("callback_url", f"{go_backend_url}/internal/sales/callback")
    print(f"[Sales Agent] Sending callback to Go backend: url={callback_url}, payload={payload}")
    
    token = os.getenv("INTERNAL_SERVICE_TOKEN", "internal-service-key-logisticshq")
    headers = {
        "X-LogisticsHQ-Service-Key": token,
        "Content-Type": "application/json"
    }
    
    try:
        resp = httpx.post(callback_url, json=payload, headers=headers, timeout=15.0)
        print(f"[Sales Agent] Callback response: status={resp.status_code}, body={resp.text}")
    except Exception as e:
        print(f"[Sales Agent] Failed to send callback request to Go backend: {e}")
        
    return {}
