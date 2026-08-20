import os
from langgraph.graph import StateGraph, START, END
from langgraph.checkpoint.postgres import PostgresSaver

from app.state.sales_state import SalesAgentState
from app.agents.email_parser_agent import (
    classify_and_parse_email_node,
    merge_context_node,
    check_completeness_node,
    lead_research_node,
    draft_rfq_node,
    save_and_callback_node
)

# Assemble the StateGraph for Sales Agent
sales_builder = StateGraph(SalesAgentState)

# Add nodes
sales_builder.add_node("classify", classify_and_parse_email_node)
sales_builder.add_node("merge_context", merge_context_node)    # NEW: fills gaps from prior turns
sales_builder.add_node("check_completeness", check_completeness_node)
sales_builder.add_node("research", lead_research_node)
sales_builder.add_node("draft_rfq", draft_rfq_node)
sales_builder.add_node("callback", save_and_callback_node)

# ── Routing functions ──────────────────────────────────────────────────────

def route_by_intent(state: SalesAgentState) -> str:
    """Routes based on the classified email intent.
    
    RFQ_REQUEST always goes through merge_context (which is a no-op if not a reply),
    then check_completeness. All other intents go straight to callback.
    """
    intent = state.get("intent", "QUESTION")
    print(f"[Sales Graph Router] Intent classified as: {intent}")
    if intent == "RFQ_REQUEST":
        return "merge_context"
    return "callback"

def route_by_completeness(state: SalesAgentState) -> str:
    """Routes based on whether mandatory RFQ fields are complete after merging context."""
    intent = state.get("intent", "QUESTION")
    print(f"[Sales Graph Router] Post-check completeness intent: {intent}")
    if intent == "RFQ_REQUEST":
        # All fields present → proceed to research + create RFQ
        return "research"
    # RFQ_REQUEST_INCOMPLETE → draft a clarification email and callback
    return "callback"

# ── Graph edges ────────────────────────────────────────────────────────────

# Entry point
sales_builder.add_edge(START, "classify")

# After classify: branch on intent
sales_builder.add_conditional_edges(
    "classify",
    route_by_intent,
    {
        "merge_context": "merge_context",   # RFQ_REQUEST → merge prior context first
        "callback": "callback"              # Everything else (QUESTION, MEETING, etc.)
    }
)

# After merging context: always check completeness
# (merge_context is a pass-through no-op for non-reply emails)
sales_builder.add_edge("merge_context", "check_completeness")

# After completeness check: branch on whether all fields are present
sales_builder.add_conditional_edges(
    "check_completeness",
    route_by_completeness,
    {
        "research": "research",     # Complete → enrich lead + create RFQ
        "callback": "callback"      # Incomplete → drafted reply sent in callback
    }
)

# Complete RFQ path: research → draft → callback → done
sales_builder.add_edge("research", "draft_rfq")
sales_builder.add_edge("draft_rfq", "callback")
sales_builder.add_edge("callback", END)

# ── Checkpointer setup ─────────────────────────────────────────────────────

def get_db_url() -> str:
    url = os.getenv("DB_URL", "postgres://user:password@localhost:5432/freel?sslmode=disable")
    if "host.docker.internal" in url:
        url = url.replace("host.docker.internal", "localhost")
    return url

# Initialize database-backed checkpointer
db_url = get_db_url()
_saver_context = PostgresSaver.from_conn_string(db_url)
saver = _saver_context.__enter__()
saver.setup()
print("[AI Sidecar Sales] Successfully initialized PostgresSaver checkpointer.")

# Compile the Sales Graph
sales_graph = sales_builder.compile(
    checkpointer=saver
)
