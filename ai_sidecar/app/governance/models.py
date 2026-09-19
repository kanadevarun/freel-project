from typing import List, Optional, Dict, Any
from pydantic import BaseModel, Field

class SafetyCheckResult(BaseModel):
    passed: bool
    check_type: str  # PROMPT_INJECTION, SENSITIVE_DATA, UNGROUNDED_CLAIM, HALLUCINATION_RISK, UNSAFE_TOOL_INTENT, SCHEMA_COMPLIANCE
    severity: str    # LOW, MEDIUM, HIGH, CRITICAL
    details: str
    detected_patterns: List[str] = []

class InputInspectionRequest(BaseModel):
    org_id: int
    user_id: Optional[int] = None
    workflow_name: str
    input_text: str
    input_payload: Optional[Dict[str, Any]] = None
    correlation_id: str

class InputInspectionResponse(BaseModel):
    is_safe: bool
    prompt_injection_detected: bool
    injection_confidence: float = Field(default=0.0, ge=0.0, le=1.0)
    pii_detected: bool
    sanitized_text: str
    detected_pii_types: List[str] = []
    safety_checks: List[SafetyCheckResult] = []
    refusal_reason: Optional[str] = None
    correlation_id: str

class OutputInspectionRequest(BaseModel):
    org_id: int
    user_id: Optional[int] = None
    workflow_name: str
    output_text: str
    output_payload: Optional[Dict[str, Any]] = None
    source_facts: List[Dict[str, Any]] = []
    proposed_actions: List[Dict[str, Any]] = []
    correlation_id: str

class ActionSafetyEvaluation(BaseModel):
    action_type: str
    action_title: str
    is_consequential: bool
    requires_human_approval: bool
    risk_level: str  # SAFE, LOW, MEDIUM, HIGH, CRITICAL
    safety_reason: str

class OutputInspectionResponse(BaseModel):
    is_safe: bool
    grounding_score: float = Field(default=1.0, ge=0.0, le=1.0)
    hallucination_risk: str  # LOW, MEDIUM, HIGH
    unsupported_claims: List[str] = []
    action_evaluations: List[ActionSafetyEvaluation] = []
    requires_human_approval: bool
    safety_checks: List[SafetyCheckResult] = []
    refusal_reason: Optional[str] = None
    correlation_id: str

class QualityEvaluationRequest(BaseModel):
    org_id: int
    test_run_id: str
    workflow_name: str
    test_cases: List[Dict[str, Any]]
    correlation_id: str

class TestCaseResult(BaseModel):
    case_id: str
    passed: bool
    schema_valid: bool
    grounding_score: float
    latency_ms: int
    error_message: Optional[str] = None

class QualityEvaluationResponse(BaseModel):
    test_run_id: str
    workflow_name: str
    total_cases: int
    passed_cases: int
    pass_rate: float
    avg_grounding_score: float
    overall_status: str  # READY_FOR_PRODUCTION, NEEDS_REVIEW, BLOCKED
    results: List[TestCaseResult] = []
    summary: str
    correlation_id: str
