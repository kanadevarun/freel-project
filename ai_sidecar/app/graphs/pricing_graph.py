import os
from langgraph.graph import StateGraph, START, END
from langgraph.prebuilt import ToolNode, tools_condition
from langgraph.checkpoint.postgres import PostgresSaver

from app.state.pricing_state import PricingAgentState
from app.agents.pricing_agent_node import pricing_agent_node, tools
from app.agents.pricing_validate_node import pricing_validate_node
from app.agents.pricing_save_node import pricing_save_node

# Assemble the StateGraph
pricing_builder = StateGraph(PricingAgentState)

# Add nodes
pricing_builder.add_node("agent", pricing_agent_node)
pricing_builder.add_node("tools", ToolNode(tools))
pricing_builder.add_node("validate", pricing_validate_node)
pricing_builder.add_node("save", pricing_save_node)

# Set starting edge
pricing_builder.add_edge(START, "agent")

# Wire ReAct condition logic: if LLM has tool calls -> tools, else -> validate
pricing_builder.add_conditional_edges(
    "agent",
    tools_condition,
    {
        "tools": "tools",
        END: "validate"
    }
)

# Loop back from tools to agent
pricing_builder.add_edge("tools", "agent")

# Route validation to save
pricing_builder.add_edge("validate", "save")
pricing_builder.add_edge("save", END)

# Helper function to get database url
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
print("[AI Sidecar Pricing] Successfully initialized PostgresSaver checkpointer.")

# Compile the Pricing Graph with an interrupt before the save node for manual reviews
pricing_graph = pricing_builder.compile(
    checkpointer=saver,
    interrupt_before=["save"]
)
