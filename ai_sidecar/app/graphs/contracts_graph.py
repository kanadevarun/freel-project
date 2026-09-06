import os
import time
from typing import Dict, Any

from app.state.contract_state import ContractExtractionState
from langgraph.graph import StateGraph, END

# Import modular node functions from agents package
#
# Simple meaning:
#   Instead of keeping all the logic for OCR, classification, and parsing
#   in one large graph file, we move each node function into its own file
#   in the agents module to keep the code modular and clean.
from app.agents.ocr_agent import ocr_node
from app.agents.classifier_agent import classify_node
from app.agents.parser_agent import parser_node
from app.agents.validator_agent import validator_node
from app.agents.ingest_agent import ingest_node
from langgraph.checkpoint.memory import MemorySaver

# --- State Graph Assembly ---
workflow_builder = StateGraph(ContractExtractionState)
workflow_builder.add_node("ocr", ocr_node)
workflow_builder.add_node("classify", classify_node)
workflow_builder.add_node("parser", parser_node)
workflow_builder.add_node("validator", validator_node)
workflow_builder.add_node("ingest", ingest_node)

workflow_builder.set_entry_point("ocr")
workflow_builder.add_edge("ocr", "classify")
workflow_builder.add_edge("classify", "parser")
workflow_builder.add_edge("parser", "validator")

def route_validator(state: ContractExtractionState):
    if state.is_anomaly_detected:
        return END  # Pauses for manual review (compiled with interrupt_before=["ingest"])
    return "ingest"

workflow_builder.add_conditional_edges("validator", route_validator)
workflow_builder.add_edge("ingest", END)

saver = MemorySaver()
print("[AI Sidecar] Successfully initialized MemorySaver checkpointer.")

contracts_graph = workflow_builder.compile(checkpointer=saver, interrupt_before=["ingest"])
