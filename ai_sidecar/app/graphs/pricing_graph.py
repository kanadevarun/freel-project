import os
from langgraph.graph import StateGraph, START, END
from langgraph.prebuilt import ToolNode, tools_condition
from langgraph.checkpoint.memory import MemorySaver

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

# Initialize in-memory checkpointer
saver = MemorySaver()
print("[AI Sidecar Pricing] Successfully initialized MemorySaver checkpointer.")

# Compile the Pricing Graph with an interrupt before the save node for manual reviews
pricing_graph = pricing_builder.compile(
    checkpointer=saver,
    interrupt_before=["save"]
)
