from typing import List, Optional, Dict, Any
from pydantic import BaseModel, Field, ConfigDict


class CopilotMessage(BaseModel):
    role: str = Field(..., description="user, assistant, or system")
    content: str = Field(..., description="Message text content")
    timestamp: Optional[str] = Field(None, description="ISO timestamp")


class CopilotSourceRef(BaseModel):
    record_type: str = Field(..., description="SHIPMENT, INVOICE, RFQ, QUOTATION, CONTRACT, LEAD, CUSTOMER, APPROVAL, etc.")
    record_id: str = Field(..., description="Unique record identifier")
    title: str = Field(..., description="Record headline or title")
    url: Optional[str] = Field(None, description="Frontend navigation route")
    snippet: Optional[str] = Field(None, description="Key fact or excerpt")


class CopilotActionProposal(BaseModel):
    action_type: str = Field(
        ...,
        description="CREATE_TASK, CREATE_RECOMMENDATION, REQUEST_DRAFT, REQUEST_APPROVAL, NAVIGATE, ESCALATE, REVIEW_DOCUMENT"
    )
    action_title: str = Field(..., description="Human-readable title of proposed action")
    description: str = Field(..., description="Rationale and context for this action")
    payload: Dict[str, Any] = Field(default_factory=dict, description="Structured parameters for execution")
    requires_approval: bool = Field(True, description="Consequential mutations require HITL approval")


class CopilotContextPayload(BaseModel):
    org_id: int = Field(..., description="Tenant organization ID")
    user_id: int = Field(..., description="Authenticated user ID")
    user_role: str = Field(..., description="User RBAC role (e.g. ADMIN, OPERATIONS, FINANCE, COMPLIANCE)")
    current_route: str = Field(..., description="Active frontend route (e.g. /dashboard/shipments/42)")
    current_module: str = Field(..., description="Active domain module (e.g. SHIPMENTS, INVOICES, DASHBOARD)")
    current_record_id: Optional[str] = Field(None, description="ID of active record on screen if viewing detail")
    active_filters: Optional[Dict[str, Any]] = Field(default_factory=dict, description="Current filter parameters")
    authorized_records: List[Dict[str, Any]] = Field(default_factory=list, description="Sanitized records authorized by Go backend")
    summary_metrics: Optional[Dict[str, Any]] = Field(default_factory=dict, description="Active KPI figures for context")


class CopilotChatRequest(BaseModel):
    context: CopilotContextPayload = Field(..., description="Context provided and verified by Go integration layer")
    query: str = Field(..., description="User question or instruction")
    conversation_history: List[CopilotMessage] = Field(default_factory=list, description="Prior conversation turns in session")
    correlation_id: str = Field(..., description="Request correlation identifier")


class CopilotChatResponse(BaseModel):
    answer: str = Field(..., description="Synthesized, grounded response explaining the page or answering query")
    confirmed_facts: List[str] = Field(default_factory=list, description="Verified data points from Go backend")
    source_references: List[CopilotSourceRef] = Field(default_factory=list, description="References to business records in context")
    signals: List[str] = Field(default_factory=list, description="Operational signals (delays, overdue amounts, SLA risks)")
    ai_interpretation: str = Field(..., description="Analytical breakdown and operational significance")
    recommendations: List[str] = Field(default_factory=list, description="Concrete next steps for operator")
    suggested_followups: List[str] = Field(default_factory=list, description="Suggested follow-up queries")
    draft_content: Optional[str] = Field(None, description="Generated communication draft if requested")
    draft_type: Optional[str] = Field(None, description="Type of draft (e.g. CUSTOMER_UPDATE, CARRIER_FOLLOW_UP, COLLECTION_NOTICE)")
    action_proposals: List[CopilotActionProposal] = Field(default_factory=list, description="Controlled actions ready for review")
    confidence: float = Field(..., ge=0.0, le=1.0, description="Confidence score")
    missing_information: List[str] = Field(default_factory=list, description="Unknown parameters or records not in authorized context")
    requires_approval: bool = Field(False, description="True if any proposed action requires human sign-off")
    safety_restrictions: List[str] = Field(default_factory=list, description="Safety guardrails applied to this query")
    correlation_id: str = Field(..., description="Correlation ID")
