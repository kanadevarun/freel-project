"""
LogisticsHQ Phase 5: Controlled Autonomous Operations Models & Schemas
Defines typed models for Autonomy Levels, Operational Plans, Candidate Alternatives,
Multi-Attribute Evaluations, Hard/Soft Constraints, Stop Conditions, Replanning Lineages,
and Verification States.
"""

from typing import List, Optional, Dict, Any, Literal
from pydantic import BaseModel, Field, field_validator
import uuid
import datetime

# Autonomy Levels (Section 3 Requirement)
AutonomyLevel = Literal[
    "LEVEL_0_OBSERVE",
    "LEVEL_1_RECOMMEND",
    "LEVEL_2_PREPARE",
    "LEVEL_3_CONTROLLED_EXECUTION",
    "LEVEL_4_CONTROLLED_MULTI_STEP",
]

# Durable Plan Lifecycle Statuses (Section 6 Requirement)
PlanStatus = Literal[
    "DRAFT",
    "GENERATED",
    "VALIDATING",
    "REQUIRES_APPROVAL",
    "APPROVED",
    "REJECTED",
    "EXECUTING",
    "PAUSED",
    "WAITING",
    "COMPLETED",
    "PARTIALLY_COMPLETED",
    "FAILED",
    "CANCELLED",
    "EXPIRED",
    "REPLANNING",
]

# Step Lifecycle Statuses
StepStatus = Literal[
    "PENDING",
    "READY",
    "AWAITING_APPROVAL",
    "APPROVED",
    "REJECTED",
    "EXECUTING",
    "WAITING",
    "COMPLETED",
    "SUCCEEDED",
    "FAILED",
    "SKIPPED",
    "CANCELLED",
]

RiskLevel = Literal["LOW", "MEDIUM", "HIGH", "CRITICAL"]
ReversibilityType = Literal["REVERSIBLE", "PARTIALLY_REVERSIBLE", "IRREVERSIBLE"]
ConstraintType = Literal["TIME", "COST", "SERVICE", "CUSTOMER", "COMPLIANCE", "OPERATIONAL", "POLICY"]


class ConstraintModel(BaseModel):
    constraint_type: ConstraintType
    description: str = Field(..., max_length=500)
    is_hard: bool = Field(default=True, description="Hard constraints cannot be violated. Soft constraints can be traded off.")
    field_target: Optional[str] = None
    operator: Optional[str] = None  # "<=", ">=", "==", "!=", "<", ">", "IN", "NOT_IN"
    threshold_value: Optional[Any] = None


class StepConditionPredicate(BaseModel):
    field: Optional[str] = None
    operator: str = "<="  # "<=", ">=", "==", "!=", "<", ">"
    value: Optional[Any] = None


class StepVerificationCriteria(BaseModel):
    check_type: str = "FIELD_EQUALS"  # FIELD_EQUALS, RECORD_EXISTS, STATUS_TRANSITION
    target_field: Optional[str] = None
    expected_value: Optional[Any] = None
    description: Optional[str] = None


class PlanStepModel(BaseModel):
    step_id: str = Field(..., description="Unique step identifier e.g. step-1")
    step_number: int = Field(..., ge=1, description="1-indexed execution sequence order")
    action_type: str = Field(..., description="Registered Action System action name")
    title: str = Field(..., min_length=1, max_length=255)
    description: str = Field("", max_length=2000)
    parameters: Dict[str, Any] = Field(default_factory=dict)
    dependencies: List[str] = Field(default_factory=list, description="Prior step_ids required before execution")
    condition_predicate: Optional[StepConditionPredicate] = None
    expected_outcome: str = Field("", max_length=1000)
    verification_criteria: Optional[StepVerificationCriteria] = None
    risk_level: RiskLevel = "LOW"
    requires_approval: bool = False
    reversibility: ReversibilityType = "REVERSIBLE"
    fallback_action: Optional[Dict[str, Any]] = None
    timeout_seconds: int = 300
    idempotency_key: Optional[str] = None
    status: Optional[StepStatus] = "PENDING"
    preconditions: Optional[List[Dict[str, Any]]] = Field(default_factory=list)
    retry_policy: Optional[Dict[str, Any]] = None
    compensation_action: Optional[Dict[str, Any]] = None
    authorization_requirements: Optional[List[str]] = Field(default_factory=list)

    @field_validator("condition_predicate", mode="before")
    @classmethod
    def clean_condition_predicate(cls, v):
        if not v or (isinstance(v, dict) and (not v or "field" not in v)):
            return None
        return v

    @field_validator("verification_criteria", mode="before")
    @classmethod
    def clean_verification_criteria(cls, v):
        if not v or (isinstance(v, dict) and (not v or "target_field" not in v)):
            return None
        return v

    @field_validator("fallback_action", "compensation_action", "retry_policy", mode="before")
    @classmethod
    def clean_empty_dicts(cls, v):
        if not v or (isinstance(v, dict) and not v):
            return None
        return v

    @field_validator("dependencies", "preconditions", "authorization_requirements", mode="before")
    @classmethod
    def clean_empty_lists(cls, v):
        if v is None:
            return []
        return v

    @field_validator("description", "expected_outcome", mode="before")
    @classmethod
    def clean_empty_strings(cls, v):
        if v is None:
            return ""
        return v


class StopConditionModel(BaseModel):
    condition_type: str = Field(..., description="E.g. MAX_STEPS_EXCEEDED, CONFIDENCE_DEGRADATION, MONETARY_CAP_EXCEEDED, STATE_DRIFT")
    threshold: Any = Field(..., description="Value threshold triggering emergency stop or escalation")
    escalation_target: str = Field("OPERATIONS_SUPERVISOR", description="Target role or team notified")
    description: Optional[str] = None


class CandidateEvaluationModel(BaseModel):
    feasibility: bool = True
    delay_reduction_hours: float = 0.0
    estimated_cost: float = 0.0
    customer_impact: RiskLevel = "LOW"
    operational_risk: RiskLevel = "LOW"
    compliance_risk: RiskLevel = "LOW"
    confidence: float = Field(0.85, ge=0.0, le=1.0)
    reversibility: ReversibilityType = "REVERSIBLE"
    step_count: int = 1
    dependency_risk: RiskLevel = "LOW"
    hard_constraints_satisfied: bool = True
    soft_constraints_score: float = Field(0.8, ge=0.0, le=1.0)
    overall_utility_score: float = Field(0.8, ge=0.0, le=1.0)
    selection_rationale: str = ""
    rejection_or_penalty_reasons: List[str] = Field(default_factory=list)


class CandidatePlanModel(BaseModel):
    candidate_id: str = Field(..., description="e.g. candidate-A")
    strategy_name: str = Field(..., description="e.g. Accelerated Carrier Recovery")
    summary: str
    is_feasible: bool = True
    rank: int = 1
    steps: List[PlanStepModel] = Field(default_factory=list)
    evaluation: CandidateEvaluationModel


class PlanGenerationRequest(BaseModel):
    org_id: int
    user_id: Optional[int] = None
    goal: str = Field(..., min_length=5, max_length=1000)
    goal_type: str = "OPERATIONAL"
    module: str = Field(..., max_length=50)
    related_entity_type: str = Field(..., max_length=50)
    related_entity_id: str = Field(..., max_length=100)
    objective: Optional[str] = None
    priority: str = "MEDIUM"
    risk_tolerance: str = "BALANCED"
    current_state: Optional[Dict[str, Any]] = Field(default_factory=dict)
    predictions_context: Optional[Dict[str, Any]] = Field(default_factory=dict)
    cross_module_context: Optional[Dict[str, Any]] = Field(default_factory=dict)
    constraints: Optional[List[str]] = Field(default_factory=list)
    hard_constraints: Optional[List[ConstraintModel]] = Field(default_factory=list)
    soft_constraints: Optional[List[ConstraintModel]] = Field(default_factory=list)
    autonomy_level: AutonomyLevel = "LEVEL_2_PREPARE"
    correlation_id: str = Field(default_factory=lambda: f"corr-auto-{uuid.uuid4().hex[:12]}")


class PlanGenerationResponse(BaseModel):
    plan_id: str = Field(default_factory=lambda: f"plan-{uuid.uuid4().hex[:12]}")
    goal_id: Optional[str] = None
    version: int = 1
    parent_plan_id: Optional[str] = None
    goal: str
    goal_type: str = "OPERATIONAL"
    fallback_strategy: Optional[str] = None
    module: str
    related_entity_type: str
    related_entity_id: str
    current_state_summary: str
    constraints: Optional[List[str]] = Field(default_factory=list)
    hard_constraints: Optional[List[ConstraintModel]] = Field(default_factory=list)
    soft_constraints: Optional[List[ConstraintModel]] = Field(default_factory=list)
    assumptions: Optional[List[str]] = Field(default_factory=list)
    risks: Optional[List[Dict[str, Any]]] = Field(default_factory=list)
    candidates: Optional[List[CandidatePlanModel]] = Field(default_factory=list)
    selected_candidate_id: str = "candidate-A"
    ordered_steps: Optional[List[PlanStepModel]] = Field(default_factory=list)
    evaluation_summary: Optional[Dict[str, Any]] = Field(default_factory=dict)
    confidence_score: float = Field(..., ge=0.0, le=1.0)
    data_sufficiency: bool = True
    estimated_impact: str
    risk_level: RiskLevel = "MEDIUM"
    autonomy_level: AutonomyLevel = "LEVEL_2_PREPARE"
    stop_conditions: List[StopConditionModel] = Field(default_factory=list)
    staleness_status: str = "FRESH"
    waiting_state: Optional[str] = None
    waiting_until: Optional[str] = None
    escalation_reason: Optional[str] = None
    customer_commitment_date: Optional[str] = None
    predicted_eta: Optional[str] = None
    eta_deviation_hours: float = 0.0
    commitment_risk_severity: RiskLevel = "LOW"
    status: PlanStatus = "GENERATED"
    correlation_id: str
    explanation: str = ""


class PlanEvaluationRequest(BaseModel):
    plan: PlanGenerationResponse
    policy: Dict[str, Any] = Field(default_factory=dict)


class PlanEvaluationResponse(BaseModel):
    is_permitted: bool
    policy_decision: Literal["PERMITTED", "REQUIRES_APPROVAL", "BLOCKED_POLICY", "BLOCKED_EMERGENCY_STOP"]
    requires_approval: bool
    policy_reason: str
    violated_rules: List[str] = Field(default_factory=list)
    confidence_acceptable: bool = True
    data_sufficiency_acceptable: bool = True


class PlanReplanRequest(BaseModel):
    original_plan_id: str
    current_version: int
    executed_steps: Optional[List[Dict[str, Any]]] = Field(default_factory=list)
    failed_or_drifted_step: Optional[Dict[str, Any]] = None
    observed_new_state: Optional[Dict[str, Any]] = Field(default_factory=dict)
    replan_reason: str = Field(..., min_length=3, max_length=1000)
    triggering_event: Optional[str] = None
    correlation_id: str = Field(default_factory=lambda: f"corr-replan-{uuid.uuid4().hex[:12]}")

    @field_validator("executed_steps", mode="before")
    @classmethod
    def ensure_steps_list(cls, v):
        return v if v is not None else []

    @field_validator("observed_new_state", mode="before")
    @classmethod
    def ensure_state_dict(cls, v):
        return v if v is not None else {}


class PlanReplanResponse(BaseModel):
    revised_plan_id: str
    new_version: int
    parent_plan_id: str
    replan_reason: str
    changes_summary: str
    updated_steps: List[PlanStepModel]
    confidence_score: float
    risk_level: RiskLevel = "MEDIUM"
    status: PlanStatus = "REPLANNING"


# ── Task 5.3: Adaptive Shipment Management Models ───────────────────────────

WaitingStateType = Literal[
    "WAITING_FOR_CARRIER",
    "WAITING_FOR_APPROVAL",
    "WAITING_FOR_CUSTOMER",
    "WAITING_FOR_MILESTONE",
    "WAITING_FOR_EXTERNAL_EVENT",
    "WAITING_FOR_VERIFICATION",
]

EventDecisionType = Literal[
    "NO_ACTION",
    "CONTINUE_MONITORING",
    "RECOMMENDATION",
    "NEW_PLAN",
    "APPROVAL_REQUIRED",
    "CONTROLLED_EXECUTION",
    "ESCALATION",
    "REPLANNING",
]


class ShipmentEventEvaluationRequest(BaseModel):
    org_id: int
    shipment_id: int
    event_type: str = Field(..., description="e.g. ETA_DETERIORATION, MILESTONE_DELAY, MISSED_MILESTONE, CARRIER_UPDATE, EXCEPTION_LOGGED")
    severity: RiskLevel = "MEDIUM"
    current_state: Dict[str, Any] = Field(default_factory=dict)
    previous_state: Optional[Dict[str, Any]] = None
    active_plan: Optional[Dict[str, Any]] = None
    customer_commitment_date: Optional[str] = None
    predicted_eta: Optional[str] = None
    raw_event_payload: Optional[Dict[str, Any]] = None
    correlation_id: str = Field(default_factory=lambda: f"corr-ship-eval-{uuid.uuid4().hex[:12]}")


class ShipmentEventEvaluationResponse(BaseModel):
    shipment_id: int
    event_type: str
    decision: EventDecisionType
    decision_reason: str
    is_meaningful_change: bool
    eta_deviation_hours: float = 0.0
    commitment_risk_severity: RiskLevel = "LOW"
    recommended_action_type: Optional[str] = None
    escalation_reason: Optional[str] = None
    requires_new_plan: bool = False
    requires_replan: bool = False
    active_plan_id: Optional[str] = None
    waiting_state: Optional[str] = None
    correlation_id: str


# =============================================================================
# Phase 5 Task 5.4: Autonomous Customer Follow-Up Models
# =============================================================================

FollowupDecisionType = Literal[
    "NO_ACTION",
    "MONITOR",
    "INTERNAL_TASK",
    "RECOMMEND_FOLLOW_UP",
    "PREPARE_FOLLOW_UP",
    "REQUIRE_APPROVAL",
    "CONTROLLED_SEND",
    "WAIT_FOR_RESPONSE",
    "ESCALATE",
    "STOP",
]

ResponseClassificationType = Literal[
    "POSITIVE",
    "NEGATIVE",
    "REQUEST_FOR_INFORMATION",
    "REQUEST_FOR_ACTION",
    "CLARIFICATION",
    "COMPLAINT",
    "ESCALATION",
    "CONFIRMATION_APPROVAL",
    "OPT_OUT_STOP",
    "UNCERTAIN",
]


class CustomerCommunicationPreferencesModel(BaseModel):
    preferred_channel: str = "EMAIL"
    opt_out: bool = False
    opt_out_reason: Optional[str] = None
    contact_restrictions: str = "NONE"
    business_hours_only: bool = True
    designated_contact_id: Optional[int] = None
    max_followups_per_incident: int = 3
    min_followup_interval_hours: int = 24


class VerifiedContactModel(BaseModel):
    contact_id: int
    first_name: str
    last_name: str
    email: str
    phone: Optional[str] = None
    job_title: Optional[str] = None
    is_primary: bool = True


class CustomerFollowupDraftModel(BaseModel):
    subject: str
    actual_facts: List[str] = Field(default_factory=list, description="Verified historical facts strictly from database")
    predictions: List[str] = Field(default_factory=list, description="Machine learning projections explicitly labeled as estimates")
    recommendations: List[str] = Field(default_factory=list, description="Suggested customer actions or next steps")
    full_body: str = Field(..., description="Formatted email/message body separating facts and predictions")
    channel: str = "EMAIL"


class CustomerFollowupEvaluationRequest(BaseModel):
    org_id: int
    customer_id: int
    customer_name: str
    account_tier: str = "STANDARD"
    event_type: str = Field(..., description="e.g. SHIPMENT_DELAY, QUOTATION_EXPIRING, UNRESOLVED_INQUIRY, PAYMENT_REMINDER, BOOKING_UPDATE")
    event_payload: Dict[str, Any] = Field(default_factory=dict)
    preferences: Optional[CustomerCommunicationPreferencesModel] = None
    contact: Optional[VerifiedContactModel] = None
    recent_followups_count: int = 0
    last_followup_hours_ago: Optional[float] = None
    active_plan: Optional[Dict[str, Any]] = None
    correlation_id: str = Field(default_factory=lambda: f"corr-cust-eval-{uuid.uuid4().hex[:12]}")


class CustomerFollowupEvaluationResponse(BaseModel):
    customer_id: int
    event_type: str
    decision: FollowupDecisionType
    decision_reason: str
    urgency: RiskLevel = "MEDIUM"
    draft: Optional[CustomerFollowupDraftModel] = None
    channel: str = "EMAIL"
    requires_approval: bool = True
    approval_reason: Optional[str] = None
    recommended_action_type: str = "SEND_CUSTOMER_COMMUNICATION"
    stop_conditions: List[str] = Field(default_factory=list)
    confidence: float = 0.90
    correlation_id: str


class ClassifyCustomerResponseRequest(BaseModel):
    org_id: int
    customer_id: int
    contact_name: str = "Customer"
    message_text: str = Field(..., min_length=1, max_length=5000)
    followup_context: Optional[Dict[str, Any]] = Field(default_factory=dict)
    correlation_id: str = Field(default_factory=lambda: f"corr-resp-class-{uuid.uuid4().hex[:12]}")


class ClassifyCustomerResponseResponse(BaseModel):
    customer_id: int
    classification: ResponseClassificationType
    sentiment: Literal["POSITIVE", "NEUTRAL", "NEGATIVE"] = "NEUTRAL"
    action_requested: Optional[str] = None
    recommended_next_step: Literal["RESOLVE_AND_STOP", "PROPOSE_OPERATIONAL_REPLAN", "ESCALATE_TO_HUMAN", "SCHEDULE_REPLY"] = "RESOLVE_AND_STOP"
    confidence: float = 0.92
    correlation_id: str


# -----------------------------------------------------------------------------
# Phase 5 Task 5.5: Intelligent RFQ and Pricing Optimization Models
# -----------------------------------------------------------------------------

PricingStrategyType = Literal[
    "COMPETITIVE_STANDARD",
    "RISK_ADJUSTED_PREMIUM",
    "CUSTOMER_RETENTION_VOLUME",
    "EXPRESS_PREMIUM",
    "CONSERVATIVE_HOLD",
]

DataSufficiencyType = Literal["COMPLETE", "SUFFICIENT", "PARTIAL", "INSUFFICIENT"]
RateFreshnessType = Literal["FRESH", "EXPIRING_SOON", "STALE", "EXPIRED"]


class PricingStrategyCandidate(BaseModel):
    strategy_id: str
    strategy_type: PricingStrategyType
    title: str
    description: str
    price: float
    base_cost: float
    predicted_cost: float
    margin_pct: float
    margin_amount: float
    operational_risk_level: RiskLevel = "LOW"
    margin_risk_level: RiskLevel = "LOW"
    acceptance_probability: float = 0.75
    confidence_score: float = 0.85
    is_feasible: bool = True
    infeasibility_reason: Optional[str] = None
    requires_approval: bool = False
    approval_reason: Optional[str] = None
    score: float = 0.80


class RfqPricingContext(BaseModel):
    org_id: int
    rfq_id: int
    rfq_number: str
    customer_id: int
    customer_name: str
    account_tier: str = "STANDARD"
    origin: str
    origin_code: Optional[str] = None
    destination: str
    destination_code: Optional[str] = None
    transport_mode: str = "OCEAN"
    incoterms: str = "FOB"
    cargo_weight_kg: float = 0.0
    cargo_volume_cbm: float = 0.0
    container_type: str = "40GP"
    target_date: Optional[str] = None
    rate_basis: Optional[Dict[str, Any]] = Field(default_factory=dict)
    pricing_policy: Optional[Dict[str, Any]] = Field(default_factory=dict)
    existing_quotation: Optional[Dict[str, Any]] = None
    customer_history: Optional[Dict[str, Any]] = Field(default_factory=dict)
    operational_risks: List[str] = Field(default_factory=list)
    special_instructions: Optional[str] = None


class RfqPricingEvaluationRequest(BaseModel):
    context: RfqPricingContext
    correlation_id: str = Field(default_factory=lambda: f"corr-rfq-eval-{uuid.uuid4().hex[:12]}")


class RfqPricingEvaluationResponse(BaseModel):
    rfq_id: int
    currency: str = "USD"
    base_cost: float
    predicted_cost: float
    actual_facts: List[str] = Field(default_factory=list)
    predictions: List[str] = Field(default_factory=list)
    assumptions: List[str] = Field(default_factory=list)
    candidate_strategies: List[PricingStrategyCandidate] = Field(default_factory=list)
    recommended_strategy_id: str
    recommended_price: float
    recommended_margin_pct: float
    margin_risk_level: RiskLevel = "LOW"
    operational_risk_level: RiskLevel = "LOW"
    confidence_score: float = 0.85
    data_sufficiency: DataSufficiencyType = "COMPLETE"
    rate_freshness_status: RateFreshnessType = "FRESH"
    requires_approval: bool = False
    approval_reason: Optional[str] = None
    reasoning_summary: str
    plan_steps: List[Dict[str, Any]] = Field(default_factory=list)
    correlation_id: str


class PricingReplanningRequest(BaseModel):
    context: RfqPricingContext
    replan_reason: str
    rate_delta: float = 0.0
    correlation_id: str = Field(default_factory=lambda: f"corr-rfq-replan-{uuid.uuid4().hex[:12]}")


# ==============================================================================
# Task 5.6: Adaptive Finance and Collections Models
# ==============================================================================

class CollectionStrategyCandidate(BaseModel):
    strategy_id: str = Field(..., description="Unique strategy identifier e.g. strat-friendly-reminder")
    strategy_name: str = Field(..., min_length=3, max_length=150)
    description: str = Field(..., max_length=1000)
    recommended_action: str = Field(..., description="Action name e.g. WAIT_AND_MONITOR, SEND_FRIENDLY_REMINDER, REQUEST_PAYMENT_STATUS, FOLLOW_UP_DISPUTE, ESCALATE_TO_FINANCE, HIGH_PRIORITY_WORKFLOW")
    priority_level: str = "MEDIUM"  # "LOW", "MEDIUM", "HIGH", "CRITICAL"
    priority_score: float = 50.0   # 0.0 to 100.0
    urgency: str = "NORMAL"         # "ROUTINE", "NORMAL", "URGENT", "IMMEDIATE"
    cooldown_days: int = 3
    requires_approval: bool = False
    approval_reason: Optional[str] = None
    expected_outcome: str
    score: float = 0.75
    draft_subject: Optional[str] = None
    draft_message: Optional[str] = None
    is_feasible: bool = True
    infeasibility_reason: Optional[str] = None


class FinanceInvoiceContext(BaseModel):
    org_id: int
    invoice_id: int
    invoice_number: str
    customer_id: int
    customer_name: str
    account_tier: str = "STANDARD"
    currency: str = "USD"
    total_amount: float
    paid_amount: float = 0.0
    balance_due: float
    due_date: Optional[str] = None
    days_overdue: int = 0
    aging_bucket: str = "CURRENT"
    invoice_status: str = "ISSUED"
    is_disputed: bool = False
    dispute_reason: Optional[str] = None
    customer_payment_behavior: Optional[Dict[str, Any]] = Field(default_factory=dict)
    other_customer_invoices: List[Dict[str, Any]] = Field(default_factory=list)
    communication_history: List[Dict[str, Any]] = Field(default_factory=list)
    special_notes: Optional[str] = None


class FinanceCollectionEvaluationRequest(BaseModel):
    context: FinanceInvoiceContext
    correlation_id: str = Field(default_factory=lambda: f"corr-fin-eval-{uuid.uuid4().hex[:12]}")


class FinanceCollectionEvaluationResponse(BaseModel):
    invoice_id: int
    invoice_number: str
    customer_name: str
    currency: str = "USD"
    total_amount: float
    balance_due: float
    days_overdue: int
    aging_bucket: str
    priority_level: str
    priority_score: float
    risk_level: RiskLevel = "LOW"
    risk_score: float = 0.20
    actual_facts: List[str] = Field(default_factory=list)
    predictions: List[str] = Field(default_factory=list)
    assumptions: List[str] = Field(default_factory=list)
    candidate_strategies: List[CollectionStrategyCandidate] = Field(default_factory=list)
    recommended_strategy_id: str
    recommended_action: str
    draft_subject: str
    draft_message: str
    requires_approval: bool = False
    approval_reason: Optional[str] = None
    stop_reason: Optional[str] = None
    confidence_score: float = 0.85
    data_sufficiency: DataSufficiencyType = "COMPLETE"
    plan_steps: List[Dict[str, Any]] = Field(default_factory=list)
    multi_invoice_summary: Optional[str] = None
    correlation_id: str


class FinanceCollectionReplanningRequest(BaseModel):
    context: FinanceInvoiceContext
    trigger_event: str
    event_payload: Optional[Dict[str, Any]] = Field(default_factory=dict)
    correlation_id: str = Field(default_factory=lambda: f"corr-fin-replan-{uuid.uuid4().hex[:12]}")


class ComplianceRemediationStrategyCandidate(BaseModel):
    strategy_id: str = Field(..., description="Unique strategy candidate ID e.g. strat-compliance-ok-monitor")
    strategy_name: str = Field(..., description="Human readable strategy name")
    strategy_type: str = Field(..., description="E.g. MONITORING, RENEWAL_PREP, DOCUMENT_REQUEST, COMMERCIAL_RECONCILIATION, HARD_COMPLIANCE_ESCALATION")
    description: str
    recommended_action: str = Field(..., description="Action name e.g. CONTINUE_MONITORING, PREPARE_RENEWAL_TASK, REQUEST_MISSING_DOCUMENT, RECONCILE_COMMERCIAL_TERMS, ESCALATE_COMPLIANCE_BLOCK")
    severity_level: str = "LOW"  # "LOW", "MEDIUM", "HIGH", "CRITICAL"
    urgency: str = "NORMAL"         # "ROUTINE", "NORMAL", "URGENT", "IMMEDIATE"
    requires_approval: bool = False
    approval_reason: Optional[str] = None
    expected_outcome: str
    score: float = 0.75
    draft_subject: Optional[str] = None
    draft_message: Optional[str] = None
    is_feasible: bool = True
    infeasibility_reason: Optional[str] = None


class ContractComplianceContext(BaseModel):
    org_id: int
    contract_id: int
    contract_reference: str
    contract_name: str
    contract_type: str = "CUSTOMER_SLA"
    party_id: Optional[int] = None
    party_name: str
    transport_mode: Optional[str] = "MULTIMODAL"
    status: str = "ACTIVE"
    currency: str = "USD"
    contract_value: float = 0.0
    effective_date: Optional[str] = None
    expiry_date: Optional[str] = None
    days_until_expiration: int = 0
    terms: List[Dict[str, Any]] = Field(default_factory=list)
    compliance_requirements: List[Dict[str, Any]] = Field(default_factory=list)
    documents: List[Dict[str, Any]] = Field(default_factory=list)
    active_shipments_count: int = 0
    rate_deviations: List[Dict[str, Any]] = Field(default_factory=list)
    operational_notes: Optional[str] = None


class ContractComplianceEvaluationRequest(BaseModel):
    context: ContractComplianceContext
    correlation_id: str = Field(default_factory=lambda: f"corr-ccm-eval-{uuid.uuid4().hex[:12]}")


class ContractComplianceEvaluationResponse(BaseModel):
    contract_id: int
    contract_reference: str
    contract_name: str
    party_name: str
    contract_type: str
    status: str
    effective_date: Optional[str] = None
    expiry_date: Optional[str] = None
    days_until_expiration: int
    expiration_status: str  # "CURRENT", "EXPIRING_SOON", "EXPIRED", "RENEWAL_PENDING", "REVIEW_REQUIRED"
    compliance_status: str  # "COMPLIANT", "WARNING", "NON_COMPLIANT", "BLOCKED"
    hard_requirement_count: int = 0
    soft_requirement_count: int = 0
    hard_violations_count: int = 0
    soft_deviations_count: int = 0
    missing_documents_count: int = 0
    expired_documents_count: int = 0
    risk_level: RiskLevel = "LOW"
    risk_score: float = 0.20
    authoritative_facts: List[str] = Field(default_factory=list)
    extracted_terms: List[str] = Field(default_factory=list)
    predictions: List[str] = Field(default_factory=list)
    assumptions: List[str] = Field(default_factory=list)
    deviations: List[Dict[str, Any]] = Field(default_factory=list)
    candidate_strategies: List[ComplianceRemediationStrategyCandidate] = Field(default_factory=list)
    recommended_strategy_id: str
    recommended_action: str
    draft_subject: str
    draft_message: str
    requires_approval: bool = False
    approval_reason: Optional[str] = None
    stop_reason: Optional[str] = None
    confidence_score: float = 0.85
    data_sufficiency: DataSufficiencyType = "COMPLETE"
    remediation_plan_steps: List[Dict[str, Any]] = Field(default_factory=list)
    correlation_id: str


class ContractComplianceReplanningRequest(BaseModel):
    context: ContractComplianceContext
    trigger_event: str
    event_payload: Optional[Dict[str, Any]] = Field(default_factory=dict)
    correlation_id: str = Field(default_factory=lambda: f"corr-ccm-replan-{uuid.uuid4().hex[:12]}")


class ExceptionCandidateRecoveryStrategy(BaseModel):
    strategy_id: str = Field(..., description="Candidate strategy ID e.g. strat-carrier-escalation")
    strategy_name: str = Field(..., description="Human readable strategy name")
    strategy_type: str = Field(..., description="E.g. CARRIER_ESCALATION, DOCUMENT_REMEDY, CUSTOMER_ADVISORY, ALTERNATE_CORRIDOR, OPERATIONS_ESCALATION")
    description: str
    recommended_action: str = Field(..., description="Action name e.g. ESCALATE_CARRIER, SUBMIT_CUSTOMS_CORRECTION, ISSUE_CUSTOMER_ADVISORY, DISPATCH_FEEDER_REROUTE, ESCALATE_OPS_DIRECTOR")
    expected_resolution_prob: float = 0.85
    time_to_resolution: str = "4-12 hours"
    cost_impact: float = 0.0
    margin_impact: str = "NEGLIGIBLE"  # NEGLIGIBLE, LOW, MODERATE, HIGH
    reversibility: str = "HIGH"        # HIGH, MODERATE, LOW, IRREVERSIBLE
    execution_complexity: str = "LOW"  # LOW, MEDIUM, HIGH
    requires_approval: bool = False
    approval_reason: Optional[str] = None
    expected_outcome: str = ""
    score: float = 0.75
    is_feasible: bool = True
    infeasibility_reason: Optional[str] = None


class ExceptionResolutionContext(BaseModel):
    org_id: int
    exception_id: int
    shipment_id: int
    exception_type: str = "ETA_DELAY"
    severity: str = "MEDIUM"
    title: str = ""
    description: Optional[str] = None
    status: str = "OPEN"
    source_event_id: Optional[str] = None
    shipment_details: Optional[Dict[str, Any]] = Field(default_factory=dict)
    customer_details: Optional[Dict[str, Any]] = Field(default_factory=dict)
    active_milestones: List[Dict[str, Any]] = Field(default_factory=list)
    related_exceptions: List[Dict[str, Any]] = Field(default_factory=list)
    documents: List[Dict[str, Any]] = Field(default_factory=list)
    notes: Optional[str] = None


class ExceptionEvaluationRequest(BaseModel):
    context: ExceptionResolutionContext
    correlation_id: str = Field(default_factory=lambda: f"corr-exc-eval-{uuid.uuid4().hex[:12]}")


class ExceptionEvaluationResponse(BaseModel):
    exception_id: int
    shipment_id: int
    exception_type: str
    severity: str
    lifecycle_status: str = "PLAN_READY"
    waiting_state: Optional[str] = None
    symptom: str = ""
    likely_root_cause: str = ""
    contributing_factors: List[str] = Field(default_factory=list)
    evidence: List[str] = Field(default_factory=list)
    confidence_score: float = 0.85
    unknown_factors: List[str] = Field(default_factory=list)
    impact_assessment: Dict[str, Any] = Field(default_factory=dict)
    hard_constraints: List[str] = Field(default_factory=list)
    candidate_strategies: List[ExceptionCandidateRecoveryStrategy] = Field(default_factory=list)
    selected_strategy_id: str
    selected_strategy_name: str
    recovery_plan_steps: List[Dict[str, Any]] = Field(default_factory=list)
    verification_criteria: Dict[str, Any] = Field(default_factory=dict)
    requires_approval: bool = False
    approval_reason: Optional[str] = None
    stop_reason: Optional[str] = None
    escalation_reason: Optional[str] = None
    data_sufficiency: DataSufficiencyType = "COMPLETE"
    correlation_id: str


class ExceptionReplanningRequest(BaseModel):
    context: ExceptionResolutionContext
    current_plan_version: int = 1
    trigger_event: str = "CARRIER_UPDATE"
    event_payload: Optional[Dict[str, Any]] = Field(default_factory=dict)
    correlation_id: str = Field(default_factory=lambda: f"corr-exc-replan-{uuid.uuid4().hex[:12]}")


# ==============================================================================
# Phase 5 Task 5.9: Multi-Step Planning & Execution Models
# ==============================================================================

class MultiStepPlanValidationRequest(BaseModel):
    plan_id: str
    steps: List[PlanStepModel]
    module: str = "general"
    autonomy_level: AutonomyLevel = "LEVEL_2_PREPARE"


class MultiStepPlanValidationResponse(BaseModel):
    is_valid: bool
    issues: List[str] = Field(default_factory=list)
    warnings: List[str] = Field(default_factory=list)
    execution_order: List[str] = Field(default_factory=list)
    parallel_groups: List[List[str]] = Field(default_factory=list)
    contains_cycles: bool = False
    has_approval_gate: bool = False


class CrossModulePlanningRequest(BaseModel):
    org_id: int
    goal: str = Field(..., min_length=5)
    primary_module: str  # e.g., "shipments", "exceptions", "rfqs", "finance"
    primary_entity_id: str
    involved_modules: List[str] = Field(default_factory=list)
    context: Dict[str, Any] = Field(default_factory=dict)
    hard_constraints: Optional[List[ConstraintModel]] = None
    soft_constraints: Optional[List[ConstraintModel]] = None
    autonomy_level: AutonomyLevel = "LEVEL_2_PREPARE"
    correlation_id: Optional[str] = None


class CrossModulePlanningResponse(BaseModel):
    plan_id: str
    goal: str
    goal_type: str = "CROSS_MODULE"
    primary_module: str
    primary_entity_id: str
    involved_modules: List[str]
    ordered_steps: List[PlanStepModel]
    parallel_groups: List[List[str]]
    stop_conditions: List[StopConditionModel]
    confidence_score: float
    data_sufficiency: bool
    requires_human_approval: bool
    correlation_id: str


# ==============================================================================
# Phase 5 Task 5.10: Continuous Monitoring & Replanning Models
# ==============================================================================

PlanHealthState = Literal[
    "HEALTHY",
    "AT_RISK",
    "STALE",
    "BLOCKED",
    "FAILED",
    "COMPLETED",
    "REPLANNING",
    "ESCALATED",
    "WAITING",
]

RecommendedAction = Literal[
    "CONTINUE",
    "PAUSE",
    "REPLAN",
    "ESCALATE",
    "STOP",
]


class MonitoringEventModel(BaseModel):
    event_id: str
    event_type: str  # e.g., "SHIPMENT_ETA_SLIP", "CARRIER_HOLD", "CUSTOMER_OBJECTION", "INVOICE_PAID", "EXCEPTION_ESCALATED"
    entity_type: str  # "SHIPMENT", "INVOICE", "CUSTOMER", "RFQ", "CONTRACT"
    entity_id: str
    source: str = "SYSTEM"
    payload: Dict[str, Any] = Field(default_factory=dict)
    timestamp: Optional[str] = None
    correlation_id: Optional[str] = None


class StateChangeEvaluationRequest(BaseModel):
    org_id: int
    plan_id: str
    current_plan: Dict[str, Any]
    steps: List[PlanStepModel] = Field(default_factory=list)
    event: MonitoringEventModel
    authoritative_state: Dict[str, Any] = Field(default_factory=dict)
    previous_predictions: Optional[Dict[str, Any]] = None
    new_predictions: Optional[Dict[str, Any]] = None
    autonomy_level: AutonomyLevel = "LEVEL_2_PREPARE"
    correlation_id: Optional[str] = None


class StateChangeEvaluationResponse(BaseModel):
    plan_id: str
    plan_health: PlanHealthState
    is_plan_valid: bool
    materiality_analysis: str
    invalidated_step_ids: List[str] = Field(default_factory=list)
    changed_assumptions: List[str] = Field(default_factory=list)
    recommended_action: RecommendedAction
    replan_rationale: Optional[str] = None
    escalation_details: Optional[str] = None
    confidence_score: float = 0.85
    correlation_id: str


class ContinuousReplanningRequest(BaseModel):
    org_id: int
    plan_id: str
    current_plan: Dict[str, Any]
    completed_steps: List[PlanStepModel] = Field(default_factory=list)
    invalidated_step_ids: List[str] = Field(default_factory=list)
    event: MonitoringEventModel
    authoritative_state: Dict[str, Any] = Field(default_factory=dict)
    hard_constraints: Optional[List[ConstraintModel]] = None
    soft_constraints: Optional[List[ConstraintModel]] = None
    autonomy_level: AutonomyLevel = "LEVEL_2_PREPARE"
    correlation_id: Optional[str] = None


class ContinuousReplanningResponse(BaseModel):
    new_plan_id: str
    version: int
    parent_plan_id: str
    ordered_steps: List[PlanStepModel]
    parallel_groups: List[List[str]] = Field(default_factory=list)
    protected_completed_steps: List[str] = Field(default_factory=list)
    changed_assumptions: List[str] = Field(default_factory=list)
    replan_rationale: str
    confidence_score: float = 0.85
    requires_approval: bool = False
    correlation_id: str


# =============================================================================
# Phase 5 Task 5.11: Human + AI Operating Model Models & Schemas
# =============================================================================

OperatingMode = Literal[
    "AI_OBSERVE",
    "AI_RECOMMEND",
    "AI_PREPARE",
    "AI_EXECUTE",
    "HUMAN_REVIEW",
    "HUMAN_APPROVAL",
    "HUMAN_OVERRIDE",
    "AI_VERIFY",
    "AI_ESCALATE",
]

ConfidenceLevel = Literal["HIGH", "MEDIUM", "LOW"]
DataSufficiency = Literal["SUFFICIENT", "PARTIALLY_SUFFICIENT", "INSUFFICIENT"]

DecisionStatus = Literal[
    "PENDING",
    "APPROVED",
    "REJECTED",
    "EXPIRED",
    "CANCELLED",
    "SUPERSEDED",
    "REQUIRES_REAPPROVAL",
    "OVERRIDDEN",
    "STOPPED",
    "ESCALATED",
]

HumanFeedbackType = Literal[
    "RECOMMENDATION_ACCEPTED",
    "RECOMMENDATION_REJECTED",
    "ALTERNATIVE_SELECTED",
    "AI_OUTPUT_EDITED",
    "ACTION_STOPPED",
    "PLAN_CHANGED",
    "ESCALATION_REQUESTED",
]


class HumanAIDecisionAnalysisRequest(BaseModel):
    org_id: int
    module: str = Field(..., description="shipments, pricing, finance, customer, contracts, exceptions")
    entity_type: str
    entity_id: str
    correlation_id: Optional[str] = None
    plan_id: Optional[str] = None
    step_id: Optional[str] = None
    title: Optional[str] = None
    context_data: Dict[str, Any] = Field(default_factory=dict, description="Authoritative business records & facts")
    predictive_data: Optional[Dict[str, Any]] = Field(default_factory=dict, description="Forecasts, predictions, delay estimates")
    autonomy_level: AutonomyLevel = "LEVEL_2_PREPARE"
    notes: Optional[str] = None


class HumanAIDecisionAnalysisResponse(BaseModel):
    decision_id: str
    operating_mode: OperatingMode
    title: str
    context_summary: str
    facts: Dict[str, Any]
    predictions: Dict[str, Any]
    ai_recommendation: str
    prepared_payload: Dict[str, Any] = Field(default_factory=dict)
    alternatives: List[Dict[str, Any]] = Field(default_factory=list)
    confidence: ConfidenceLevel = "HIGH"
    data_sufficiency: DataSufficiency = "SUFFICIENT"
    risk_level: RiskLevel = "MEDIUM"
    requires_human_approval: bool = True
    is_reversible: bool = True
    provenance_summary: str
    correlation_id: str


class HumanFeedbackAnalysisRequest(BaseModel):
    org_id: int
    decision_id: str
    feedback_type: HumanFeedbackType
    human_decision: str  # APPROVE, REJECT, OVERRIDE, STOP, ESCALATE
    decision_reason: Optional[str] = None
    original_ai_payload: Optional[str] = None
    human_edited_payload: Optional[str] = None
    correlation_id: Optional[str] = None


class HumanFeedbackAnalysisResponse(BaseModel):
    decision_id: str
    learned_preference: str
    memory_candidate: Optional[Dict[str, Any]] = None
    replan_suggested: bool = False
    notes: str
    correlation_id: str


class CommandCenterPrioritizedItem(BaseModel):
    id: str
    priority_rank: int
    priority_score: float
    priority_tier: str
    severity: str
    entity_type: str
    entity_id: str
    title: str
    issue_summary: str
    why_flagged: str
    actual_facts: str
    predicted_impact: str
    recommended_action: str
    owner: str = "Unassigned"
    deadline: Optional[str] = None
    urgency: str = "MEDIUM"
    source: str = "AUTONOMY_ENGINE"
    requires_human: bool = True


class CommandCenterPrioritizeRequest(BaseModel):
    org_id: int
    items: List[Dict[str, Any]] = Field(default_factory=list)
    correlation_id: Optional[str] = None


class CommandCenterPrioritizeResponse(BaseModel):
    items: List[CommandCenterPrioritizedItem]
    critical_count: int = 0
    high_count: int = 0
    executive_summary: str
    correlation_id: str


# -----------------------------------------------------------------------------
# Phase 5 Task 5.13: Agent Memory and Learning from Outcomes Models
# -----------------------------------------------------------------------------

OutcomeType = Literal[
    "RECOMMENDATION_OUTCOME",
    "PLAN_EXECUTION_OUTCOME",
    "EXCEPTION_RECOVERY_OUTCOME",
    "CUSTOMER_COMMUNICATION_OUTCOME",
    "CARRIER_RESPONSE_OUTCOME",
    "PRICING_DECISION_OUTCOME",
    "COLLECTION_OUTCOME",
    "COMPLIANCE_REMEDIATION_OUTCOME",
    "HUMAN_DECISION_OUTCOME",
]

OutcomeStatus = Literal[
    "SUCCESS",
    "PARTIAL_SUCCESS",
    "FAILED",
    "CANCELLED",
    "REJECTED",
    "SUPERSEDED",
    "ESCALATED",
    "UNVERIFIED",
    "TIMEOUT",
]

MemoryCategory = Literal[
    "OPERATIONAL",
    "CUSTOMER",
    "PRICING",
    "FINANCE",
    "CONTRACT_COMPLIANCE",
    "WORKFLOW",
    "HUMAN_FEEDBACK",
    "SYSTEM",
]

MemoryConfidence = Literal["HIGH", "MEDIUM", "LOW"]
MemoryScope = Literal["TENANT", "CUSTOMER", "SHIPMENT", "CARRIER", "WORKFLOW", "EXCEPTION", "USER", "TEAM", "GLOBAL_SYSTEM"]
ProvenanceType = Literal["HUMAN_ENTERED", "SYSTEM_DERIVED", "AI_DERIVED", "EXTERNAL_SOURCE"]

PatternType = Literal[
    "CARRIER_DISRUPTION_PATTERN",
    "EXCEPTION_RECOVERY_STRATEGY",
    "CUSTOMER_PREFERENCE_PATTERN",
    "PRICING_CONVERSION_PATTERN",
    "COLLECTION_FRICTION_PATTERN",
    "COMPLIANCE_BOTTLENECK_PATTERN",
    "WORKFLOW_FAILURE_PATTERN",
]


class AgentMemoryItemModel(BaseModel):
    id: Optional[int] = None
    org_id: int
    scope: str = "TENANT"
    category: str = "OPERATIONAL"
    memory_type: str
    title: str
    content: str
    structured_value: Optional[Dict[str, Any]] = None
    entity_type: Optional[str] = None
    entity_id: Optional[str] = None
    outcome_id: Optional[str] = None
    confidence: str = "HIGH"
    confidence_score: float = 0.90
    recency_weight: float = 1.00
    times_observed: int = 1
    times_used: int = 0
    success_count: int = 1
    failure_count: int = 0
    is_stale: bool = False
    conflict_status: str = "NONE"
    provenance_type: str = "SYSTEM_DERIVED"
    source_reference: Optional[str] = None
    evidence: Optional[str] = None
    status: str = "ACTIVE"
    created_at: Optional[str] = None
    updated_at: Optional[str] = None


class OutcomeEvaluationRequest(BaseModel):
    org_id: int
    source_entity_type: str
    source_entity_id: str
    outcome_type: str
    plan_id: Optional[str] = None
    step_id: Optional[str] = None
    action_type: Optional[str] = None
    expected_result: str
    actual_result: str
    status: str = "UNVERIFIED"
    time_to_resolution_sec: Optional[int] = None
    human_involvement: str = "NONE"
    decided_by_name: Optional[str] = None
    metadata: Optional[Dict[str, Any]] = Field(default_factory=dict)
    correlation_id: Optional[str] = None


class OutcomeEvaluationResponse(BaseModel):
    outcome_status: str
    is_verified: bool
    evaluation_summary: str
    failure_category: str = "NONE"
    confidence: str = "HIGH"
    confidence_score: float = 0.95
    should_create_memory: bool = False
    memory_candidate: Optional[AgentMemoryItemModel] = None
    correlation_id: str


class MemoryRetrievalRequest(BaseModel):
    org_id: int
    query_context: str
    module: Optional[str] = None
    category: Optional[str] = None
    entity_type: Optional[str] = None
    entity_id: Optional[str] = None
    carrier_scac: Optional[str] = None
    customer_id: Optional[str] = None
    limit: int = 5
    include_stale: bool = False
    candidate_memories: List[Dict[str, Any]] = Field(default_factory=list)
    correlation_id: Optional[str] = None


class MemoryRetrievalResponse(BaseModel):
    retrieved_memories: List[Dict[str, Any]]
    total_found: int
    context_summary: str
    provenance_breakdown: Dict[str, int]
    has_conflicts: bool = False
    conflict_warnings: List[str] = Field(default_factory=list)
    correlation_id: str


class PatternDetectionRequest(BaseModel):
    org_id: int
    outcomes: List[Dict[str, Any]] = Field(default_factory=list)
    memories: List[Dict[str, Any]] = Field(default_factory=list)
    correlation_id: Optional[str] = None


class PatternDetectionResponse(BaseModel):
    detected_patterns: List[Dict[str, Any]]
    total_patterns: int
    summary: str
    correlation_id: str


class MemoryConflictRequest(BaseModel):
    org_id: int
    new_observation: str
    existing_memories: List[Dict[str, Any]] = Field(default_factory=list)
    correlation_id: Optional[str] = None


class MemoryConflictResponse(BaseModel):
    has_conflict: bool
    conflicting_memory_id: Optional[int] = None
    conflict_explanation: str
    authoritative_resolution: str
    recommended_action: str  # SUPERSEDE_OLD, MARK_STALE, RETAIN_BOTH, NO_CONFLICT
    correlation_id: str


# ==============================================================================
# Phase 5 Task 5.14: Governance for Controlled Autonomy Models
# ==============================================================================

class GovernanceContextEvaluationRequest(BaseModel):
    org_id: int
    module: str
    action_type: str
    entity_type: str
    entity_id: str
    parameters: Dict[str, Any] = Field(default_factory=dict)
    context_text: str = ""
    requested_autonomy: int = 2
    configured_autonomy: int = 3
    historical_memories: List[Dict[str, Any]] = Field(default_factory=list)
    correlation_id: Optional[str] = None


class GovernanceContextEvaluationResponse(BaseModel):
    risk_score: float = Field(..., ge=0.0, le=100.0)
    risk_class: str = "MEDIUM"  # LOW, MEDIUM, HIGH, CRITICAL
    data_sufficiency: str = "SUFFICIENT"  # SUFFICIENT, PARTIALLY_SUFFICIENT, INSUFFICIENT
    missing_data_fields: List[str] = Field(default_factory=list)
    confidence_class: str = "HIGH"  # HIGH, MEDIUM, LOW
    confidence_score: float = 0.85
    recommended_autonomy_tier: int = 2
    blast_radius_assessment: str = ""
    explanation: str = ""
    sanitized_context: str = ""
    correlation_id: str


class GovernancePlanPreviewRequest(BaseModel):
    org_id: int
    plan_id: str
    goal: str
    module: str
    steps: List[Dict[str, Any]] = Field(default_factory=list)
    correlation_id: Optional[str] = None


class GovernancePlanPreviewResponse(BaseModel):
    plan_id: str
    total_steps: int
    executable_steps: int
    approval_required_steps: int
    max_risk_class: str = "LOW"
    estimated_financial_exposure_usd: float = 0.0
    affected_entities: List[Dict[str, str]] = Field(default_factory=list)
    safety_summary: str = ""
    correlation_id: str




