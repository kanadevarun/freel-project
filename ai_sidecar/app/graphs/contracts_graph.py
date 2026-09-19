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
from app.persistence.checkpointer import get_checkpointer

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

# Route validator to ingest (compiled with interrupt_before=["ingest"] for HITL review)
workflow_builder.add_edge("validator", "ingest")
workflow_builder.add_edge("ingest", END)

saver = get_checkpointer()
print("[AI Sidecar Contracts] Successfully initialized checkpointer.")

contracts_graph = workflow_builder.compile(checkpointer=saver, interrupt_before=["ingest"])
