"""
Pydantic Schemas for Phase 3 Task 3.9: Advanced Notifications and Escalations AI capabilities.
All models enforce strict data validation, confidence scores, and safety restrictions.
"""
from typing import Dict, Any, List, Optional
from pydantic import BaseModel, Field


class NotificationAnalysisRequest(BaseModel):
    org_id: int = Field(..., description="Organization ID for tenant isolation")
    notification_id: Optional[int] = Field(None, description="Existing notification ID if re-evaluating")
    source_module: str = Field(..., description="SHIPMENTS, INVOICES, CONTRACTS, APPROVALS, AUTOMATIONS, RECOMMENDATIONS, SYSTEM")
    source_record_type: str = Field(..., description="SHIPMENT, INVOICE, CONTRACT, DOCUMENT, APPROVAL, LEAD, RFQ")
    source_record_id: str = Field(..., description="Identifier of the source record")
    notification_type: str = Field(..., description="Notification event type")
    title: str = Field(..., description="Notification headline")
    message: str = Field(..., description="Notification message body")
    severity: str = Field("INFORMATIONAL", description="INFORMATIONAL, LOW, MEDIUM, HIGH, CRITICAL")
    current_priority: str = Field("MEDIUM", description="LOW, MEDIUM, HIGH, URGENT")
    age_hours: float = Field(0.0, description="Hours elapsed since issue detection")
    escalation_level: int = Field(0, description="Current escalation tier (0=Initial, 1=Level 1, 2=Level 2, 3=Executive)")
    unresolved_duration_hours: float = Field(0.0, description="Hours issue has remained unresolved or unacknowledged")
    context_payload: Dict[str, Any] = Field(default_factory=dict, description="Authorized domain context attributes")
    correlation_id: str = Field(..., description="Tracing correlation ID")


class NotificationAnalysisResponse(BaseModel):
    ai_summary: str = Field(..., description="Executive operational summary of the notification")
    ai_escalation_reason: str = Field(..., description="AI explanation of why this issue warrants escalation or immediate focus")
    recommended_priority: str = Field(..., description="Evaluated priority: LOW, MEDIUM, HIGH, URGENT")
    priority_score: float = Field(..., ge=0.0, le=100.0, description="Calculated priority score (0.0 to 100.0)")
    recommended_role_target: str = Field(..., description="Target organizational role: OPERATIONS, FINANCE, COMPLIANCE, SALES, MANAGEMENT")
    suggested_action: str = Field(..., description="Specific recommended next-step operational action")
    requires_approval: bool = Field(False, description="True if recommended action or communication requires HITL approval")
    risk_level: str = Field("LOW", description="Risk level: LOW, MEDIUM, HIGH, CRITICAL")
    group_key: str = Field(..., description="Semantic cluster key for grouping related notifications and reducing alert fatigue")
    overload_reduction_advice: str = Field(..., description="Advice on digest aggregation, suppression, or deduplication")
    confidence_score: float = Field(..., ge=0.0, le=1.0, description="Model confidence score")
    correlation_id: str = Field(..., description="Echo of correlation ID")


class EscalationDraftRequest(BaseModel):
    org_id: int = Field(..., description="Organization ID for tenant isolation")
    notification_id: int = Field(..., description="Notification ID being escalated")
    draft_type: str = Field(..., description="INTERNAL_ESCALATION, MANAGEMENT_ALERT, CARRIER_FOLLOWUP, CLIENT_ADVISORY")
    source_module: str = Field(..., description="Source module")
    source_record_id: str = Field(..., description="Source record ID")
    recipient_role: str = Field(..., description="Target recipient role or audience")
    escalation_level: int = Field(1, description="Escalation tier: 1=Supervisor, 2=Manager, 3=Executive")
    key_findings: List[str] = Field(default_factory=list, description="Extracted operational facts or failure points")
    correlation_id: str = Field(..., description="Tracing correlation ID")


class EscalationDraftResponse(BaseModel):
    subject: str = Field(..., description="Generated draft subject line")
    body_text: str = Field(..., description="Generated draft communication body")
    recommended_channels: List[str] = Field(default_factory=list, description="Recommended delivery channels: IN_APP, EMAIL_DIGEST, URGENT_SMS")
    is_external: bool = Field(False, description="Whether this communication targets external clients or vendors")
    requires_approval: bool = Field(True, description="Consequential drafts must be approved before delivery")
    confidence_score: float = Field(..., ge=0.0, le=1.0, description="Model confidence score")
    correlation_id: str = Field(..., description="Echo of correlation ID")
