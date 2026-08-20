from langgraph.graph import StateGraph, END
from app.state.compliance_state import ComplianceState
from app.agents.doc_ocr_node import doc_ocr_node
from app.agents.doc_extract_node import doc_extract_node
from app.agents.doc_reconcile_node import doc_reconcile_node
from app.agents.doc_report_node import doc_report_node

def create_compliance_graph():
    workflow = StateGraph(ComplianceState)

    # Register Nodes
    workflow.add_node("ocr", doc_ocr_node)
    workflow.add_node("extract", doc_extract_node)
    workflow.add_node("reconcile", doc_reconcile_node)
    workflow.add_node("report", doc_report_node)

    # Establish Transitions
    workflow.set_entry_point("ocr")
    workflow.add_edge("ocr", "extract")
    workflow.add_edge("extract", "reconcile")
    workflow.add_edge("reconcile", "report")
    workflow.add_edge("report", END)

    return workflow.compile()

compliance_graph = create_compliance_graph()
