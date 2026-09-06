"""
operations_graph.py — LangGraph Operations Agent

Graph:
  START
    ↓
  parse_update          ← LLM + keyword fallback: extract milestones + exception signals
    ↓
  update_milestones     ← Call Go to mark milestones COMPLETED
    ↓
  detect_exceptions     ← Deterministic severity rules + call Go to persist exceptions
    ↓
  ops_action            ← Final callback to Go with summary + has_critical_exception flag
    ↓
  END

No HITL interrupt — normal tracking updates flow automatically.
Ops team is alerted via has_critical_exception flag in the callback.
"""

import os
from langgraph.graph import StateGraph, START, END

from app.state.operations_state import OperationsAgentState
from app.agents.ops_parse_node import ops_parse_node
from app.agents.ops_milestone_node import ops_milestone_node
from app.agents.ops_exception_node import ops_exception_node
from app.agents.ops_action_node import ops_action_node

# Assemble the StateGraph
ops_builder = StateGraph(OperationsAgentState)

# Add nodes
ops_builder.add_node("parse_update", ops_parse_node)
ops_builder.add_node("update_milestones", ops_milestone_node)
ops_builder.add_node("detect_exceptions", ops_exception_node)
ops_builder.add_node("ops_action", ops_action_node)

# Linear edges
ops_builder.add_edge(START, "parse_update")
ops_builder.add_edge("parse_update", "update_milestones")
ops_builder.add_edge("update_milestones", "detect_exceptions")
ops_builder.add_edge("detect_exceptions", "ops_action")
ops_builder.add_edge("ops_action", END)

from langgraph.checkpoint.memory import MemorySaver

saver = MemorySaver()
print("[AI Sidecar Ops] Successfully initialized MemorySaver checkpointer for OperationsAgent.")

# Compile the Operations Graph
operations_graph = ops_builder.compile(
    checkpointer=saver
)
