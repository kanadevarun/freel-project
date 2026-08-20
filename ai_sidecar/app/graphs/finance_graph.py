from langgraph.graph import StateGraph, END
from app.state.finance_state import FinanceState
from app.agents.invoice_ocr_node import invoice_ocr_node
from app.agents.invoice_extract_node import invoice_extract_node
from app.agents.invoice_reconcile_node import invoice_reconcile_node
from app.agents.invoice_report_node import invoice_report_node


def create_finance_graph():
    workflow = StateGraph(FinanceState)

    # Register Nodes
    workflow.add_node("ocr", invoice_ocr_node)
    workflow.add_node("extract", invoice_extract_node)
    workflow.add_node("reconcile", invoice_reconcile_node)
    workflow.add_node("report", invoice_report_node)

    # Wire Transitions
    workflow.set_entry_point("ocr")
    workflow.add_edge("ocr", "extract")
    workflow.add_edge("extract", "reconcile")
    workflow.add_edge("reconcile", "report")
    workflow.add_edge("report", END)

    return workflow.compile()


finance_graph = create_finance_graph()
