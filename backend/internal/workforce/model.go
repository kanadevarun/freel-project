package workforce

import (
	"encoding/json"
	"time"
)

// AgentType defines the organizational role of the agent
type AgentType string

const (
	AgentTypeCoordinator AgentType = "COORDINATOR"
	AgentTypeSpecialist  AgentType = "SPECIALIST"
	AgentTypeEvaluator   AgentType = "EVALUATOR"
	AgentTypeAnalyst     AgentType = "ANALYST"
)

// TaskStatus defines the durable lifecycle of multi-agent tasks
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "PENDING"
	TaskStatusAssigned  TaskStatus = "ASSIGNED"
	TaskStatusRunning   TaskStatus = "RUNNING"
	TaskStatusWaiting   TaskStatus = "WAITING"
	TaskStatusBlocked   TaskStatus = "BLOCKED"
	TaskStatusCompleted TaskStatus = "COMPLETED"
	TaskStatusFailed    TaskStatus = "FAILED"
	TaskStatusCancelled TaskStatus = "CANCELLED"
	TaskStatusEscalated TaskStatus = "ESCALATED"
)

// TaskPriority defines execution priority
type TaskPriority string

const (
	PriorityLow    TaskPriority = "LOW"
	PriorityMedium TaskPriority = "MEDIUM"
	PriorityHigh   TaskPriority = "HIGH"
	PriorityUrgent TaskPriority = "URGENT"
)

// MessageType defines structured inter-agent communication types
type MessageType string

const (
	MessageTypeRequest    MessageType = "REQUEST"
	MessageTypeDelegation MessageType = "DELEGATION"
	MessageTypeContext    MessageType = "CONTEXT"
	MessageTypeResult     MessageType = "RESULT"
	MessageTypeStatus     MessageType = "STATUS"
	MessageTypeEscalation MessageType = "ESCALATION"
	MessageTypeError      MessageType = "ERROR"
	MessageTypeHandoff    MessageType = "HANDOFF"
)

// ContextItemType distinguishes epistemological categories in shared context
type ContextItemType string

const (
	ContextTypeFact          ContextItemType = "FACT"
	ContextTypePrediction    ContextItemType = "PREDICTION"
	ContextTypeRecommendation ContextItemType = "RECOMMENDATION"
	ContextTypeAgentResult   ContextItemType = "AGENT_RESULT"
	ContextTypeHumanDecision ContextItemType = "HUMAN_DECISION"
	ContextTypeSystemEvent   ContextItemType = "SYSTEM_EVENT"
)

// HandoffStatus tracks state of structured agent transfers
type HandoffStatus string

const (
	HandoffStatusInitiated HandoffStatus = "INITIATED"
	HandoffStatusAccepted  HandoffStatus = "ACCEPTED"
	HandoffStatusRejected  HandoffStatus = "REJECTED"
	HandoffStatusCompleted HandoffStatus = "COMPLETED"
)

// Standard Capability Constants
const (
	CapShipmentRead              = "shipment.read"
	CapShipmentAnalyze           = "shipment.analyze"
	CapShipmentPredict           = "shipment.predict"
	CapExceptionRead             = "exception.read"
	CapExceptionAnalyze          = "exception.analyze"
	CapExceptionRecommend        = "exception.recommend"
	CapCustomerRead              = "customer.read"
	CapCustomerAnalyze           = "customer.analyze"
	CapCustomerFollowupRecommend = "customer.followup_recommend"
	CapRFQRead                   = "rfq.read"
	CapPricingAnalyze            = "pricing.analyze"
	CapPricingRecommend          = "pricing.recommend"
	CapInvoiceRead               = "invoice.read"
	CapFinanceAnalyze            = "finance.analyze"
	CapFinanceRecommend          = "finance.recommend"
	CapContractRead              = "contract.read"
	CapContractAnalyze           = "contract.analyze"
	CapComplianceRead            = "compliance.read"
	CapComplianceAnalyze         = "compliance.analyze"
	CapPlanningCreate            = "planning.create"
	CapPlanningEvaluate          = "planning.evaluate"
	CapTaskDelegate              = "task.delegate"
	CapWorkforceObserve          = "workforce.observe"
	CapTaskMonitor               = "task.monitor"
	CapMonitoringObserve         = "monitoring.observe"
	CapMemoryRetrieve            = "memory.retrieve"
	CapMemoryAnalyze             = "memory.analyze"
	CapOutcomeRecord             = "outcome.record"
)

// Standard Autonomy Level Constants (Phase 6.9)
const (
	AutonomyLevel0Observe           = "LEVEL_0_OBSERVE"
	AutonomyLevel1Recommend         = "LEVEL_1_RECOMMEND"
	AutonomyLevel2Prepare           = "LEVEL_2_PREPARE"
	AutonomyLevel3ControlledExec    = "LEVEL_3_CONTROLLED_EXECUTION"
	AutonomyLevel3ControlledExecution = "LEVEL_3_CONTROLLED_EXECUTION"
	AutonomyLevel4GovernedMultiStep = "LEVEL_4_GOVERNED_MULTI_STEP"
)

// Standard Operational & Health Status Constants
const (
	AgentStatusActive   = "ACTIVE"
	AgentStatusPaused   = "PAUSED"
	AgentStatusDisabled = "DISABLED"
)

const (
	HealthStatusHealthy  = "HEALTHY"
	HealthStatusDegraded = "DEGRADED"
	HealthStatusFailing  = "FAILING"
	HealthStatusDisabled = "DISABLED"
	HealthStatusUnknown  = "UNKNOWN"
)

// Emergency Stop Scopes
const (
	EmergencyStopScopeWorkforce   = "WORKFORCE"
	EmergencyStopScopeAgent       = "AGENT"
	EmergencyStopScopeWorkflow    = "WORKFLOW"
	EmergencyStopScopeActionClass = "ACTION_CLASS"
)

// WorkforceAgent represents a registered agent in the Agent Registry
type WorkforceAgent struct {
	ID                int64           `db:"id" json:"id"`
	OrgID             int64           `db:"org_id" json:"org_id"` // 0 = system baseline, >0 = tenant specific
	AgentID           string          `db:"agent_id" json:"agent_id"`
	AgentType         AgentType       `db:"agent_type" json:"agent_type"`
	Name              string          `db:"name" json:"name"`
	Description       string          `db:"description" json:"description"`
	Capabilities      json.RawMessage `db:"capabilities" json:"capabilities"`
	AllowedTasks      json.RawMessage `db:"allowed_tasks" json:"allowed_tasks"`
	AllowedEntities   json.RawMessage `db:"allowed_entities" json:"allowed_entities"`
	AutonomyLevel     string          `db:"autonomy_level" json:"autonomy_level"`
	IsEnabled         bool            `db:"is_enabled" json:"is_enabled"`
	OperationalStatus string          `db:"operational_status" json:"operational_status"`
	Version           string          `db:"version" json:"version"`
	PromptVersion     string          `db:"prompt_version" json:"prompt_version"`
	HealthStatus      string          `db:"health_status" json:"health_status"`
	LastExecutionAt   *time.Time      `db:"last_execution_at" json:"last_execution_at,omitempty"`
	LastExecutionInfo json.RawMessage `db:"last_execution_info" json:"last_execution_info,omitempty"`
	CreatedAt         time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time       `db:"updated_at" json:"updated_at"`
}

// WorkforceTask represents a durable unit of work assigned to an agent
type WorkforceTask struct {
	ID                   int64           `db:"id" json:"id"`
	TaskID               string          `db:"task_id" json:"task_id"`
	OrgID                int64           `db:"org_id" json:"org_id"`
	Objective            string          `db:"objective" json:"objective"`
	InitiatingEvent      *string         `db:"initiating_event" json:"initiating_event,omitempty"`
	ParentTaskID         *string         `db:"parent_task_id" json:"parent_task_id,omitempty"`
	RootTaskID           *string         `db:"root_task_id" json:"root_task_id,omitempty"`
	AssignedAgentID      string          `db:"assigned_agent_id" json:"assigned_agent_id"`
	Status               TaskStatus      `db:"status" json:"status"`
	Priority             TaskPriority    `db:"priority" json:"priority"`
	ContextReference     json.RawMessage `db:"context_reference" json:"context_reference,omitempty"`
	RequiredCapabilities json.RawMessage `db:"required_capabilities" json:"required_capabilities"`
	Dependencies         json.RawMessage `db:"dependencies" json:"dependencies,omitempty"`
	Result               json.RawMessage `db:"result" json:"result,omitempty"`
	Confidence           float64         `db:"confidence" json:"confidence"`
	ErrorCode            *string         `db:"error_code" json:"error_code,omitempty"`
	ErrorMessage         *string         `db:"error_message" json:"error_message,omitempty"`
	CorrelationID        string          `db:"correlation_id" json:"correlation_id"`
	StartedAt            *time.Time      `db:"started_at" json:"started_at,omitempty"`
	CompletedAt          *time.Time      `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt            time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time       `db:"updated_at" json:"updated_at"`
}

// WorkforceMessage represents an internal structured message between agents
type WorkforceMessage struct {
	ID                 int64           `db:"id" json:"id"`
	MessageID          string          `db:"message_id" json:"message_id"`
	OrgID              int64           `db:"org_id" json:"org_id"`
	TaskID             string          `db:"task_id" json:"task_id"`
	ParentTaskID       *string         `db:"parent_task_id" json:"parent_task_id,omitempty"`
	SenderAgentID      string          `db:"sender_agent_id" json:"sender_agent_id"`
	RecipientAgentID   string          `db:"recipient_agent_id" json:"recipient_agent_id"`
	MessageType        MessageType     `db:"message_type" json:"message_type"`
	Objective          string          `db:"objective" json:"objective"`
	RequestedCapability *string        `db:"requested_capability" json:"requested_capability,omitempty"`
	ContextReference   json.RawMessage `db:"context_reference" json:"context_reference,omitempty"`
	Payload            json.RawMessage `db:"payload" json:"payload"`
	Priority           TaskPriority    `db:"priority" json:"priority"`
	CorrelationID      string          `db:"correlation_id" json:"correlation_id"`
	CreatedAt          time.Time       `db:"created_at" json:"created_at"`
}

// WorkforceContextItem represents a typed item in the controlled shared context
type WorkforceContextItem struct {
	ID              int64           `db:"id" json:"id"`
	ContextID       string          `db:"context_id" json:"context_id"`
	OrgID           int64           `db:"org_id" json:"org_id"`
	TaskID          string          `db:"task_id" json:"task_id"`
	ItemType        ContextItemType `db:"item_type" json:"item_type"`
	EntityType      *string         `db:"entity_type" json:"entity_type,omitempty"`
	EntityID        *string         `db:"entity_id" json:"entity_id,omitempty"`
	SourceAgentID   *string         `db:"source_agent_id" json:"source_agent_id,omitempty"`
	ProvenanceID    *string         `db:"provenance_id" json:"provenance_id,omitempty"`
	Content         json.RawMessage `db:"content" json:"content"`
	Confidence      float64         `db:"confidence" json:"confidence"`
	IsAuthoritative bool            `db:"is_authoritative" json:"is_authoritative"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
}

// WorkforceHandoff represents a documented handoff from one agent to another
type WorkforceHandoff struct {
	ID                 int64           `db:"id" json:"id"`
	HandoffID          string          `db:"handoff_id" json:"handoff_id"`
	OrgID              int64           `db:"org_id" json:"org_id"`
	TaskID             string          `db:"task_id" json:"task_id"`
	OriginatingTaskID  string          `db:"originating_task_id" json:"originating_task_id"`
	SourceAgentID      string          `db:"source_agent_id" json:"source_agent_id"`
	DestinationAgentID string          `db:"destination_agent_id" json:"destination_agent_id"`
	Reason             string          `db:"reason" json:"reason"`
	Objective          string          `db:"objective" json:"objective"`
	RequiredCapability string          `db:"required_capability" json:"required_capability"`
	ContextReferences  json.RawMessage `db:"context_references" json:"context_references,omitempty"`
	CurrentFindings    json.RawMessage `db:"current_findings" json:"current_findings,omitempty"`
	ExpectedOutput     string          `db:"expected_output" json:"expected_output"`
	Confidence         float64         `db:"confidence" json:"confidence"`
	Status             HandoffStatus   `db:"status" json:"status"`
	CreatedAt          time.Time       `db:"created_at" json:"created_at"`
}

// TaskHierarchyNode represents a parent task and its nested children
type TaskHierarchyNode struct {
	Task     *WorkforceTask       `json:"task"`
	Children []*TaskHierarchyNode `json:"children"`
}

// DTO Requests & Responses

type RegisterAgentRequest struct {
	AgentID         string    `json:"agent_id"`
	AgentType       AgentType `json:"agent_type"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Capabilities    []string  `json:"capabilities"`
	AllowedTasks    []string  `json:"allowed_tasks"`
	AllowedEntities []string  `json:"allowed_entities"`
	AutonomyLevel   string    `json:"autonomy_level,omitempty"`
	Version         string    `json:"version,omitempty"`
	PromptVersion   string    `json:"prompt_version,omitempty"`
}

type UpdateAgentRequest struct {
	Name            *string    `json:"name,omitempty"`
	Description     *string    `json:"description,omitempty"`
	Capabilities    []string   `json:"capabilities,omitempty"`
	AllowedTasks    []string   `json:"allowed_tasks,omitempty"`
	AllowedEntities []string   `json:"allowed_entities,omitempty"`
	AutonomyLevel   *string    `json:"autonomy_level,omitempty"`
	IsEnabled       *bool      `json:"is_enabled,omitempty"`
	HealthStatus    *string    `json:"health_status,omitempty"`
}

type CreateTaskRequest struct {
	Objective            string                  `json:"objective"`
	InitiatingEvent      string                  `json:"initiating_event,omitempty"`
	AssignedAgentID      string                  `json:"assigned_agent_id"`
	Priority             TaskPriority            `json:"priority,omitempty"`
	RequiredCapabilities []string                `json:"required_capabilities,omitempty"`
	ContextReference     map[string]interface{}  `json:"context_reference,omitempty"`
	Dependencies         []string                `json:"dependencies,omitempty"`
	InitialContextItems  []AddContextItemRequest `json:"initial_context_items,omitempty"`
	CorrelationID        string                  `json:"correlation_id,omitempty"`
	IdempotencyKey       string                  `json:"idempotency_key,omitempty"`
}

type DelegateTaskRequest struct {
	TargetAgentID        string                 `json:"target_agent_id"`
	Objective            string                 `json:"objective"`
	RequiredCapabilities []string               `json:"required_capabilities,omitempty"`
	ExpectedOutput       string                 `json:"expected_output"`
	Priority             TaskPriority           `json:"priority,omitempty"`
	ContextReferences    []string               `json:"context_references,omitempty"`
	ContextReference     map[string]interface{} `json:"context_reference,omitempty"`
	Dependencies         []string               `json:"dependencies,omitempty"`
	IsOptional           bool                   `json:"is_optional,omitempty"`
	IdempotencyKey       string                 `json:"idempotency_key,omitempty"`
}

type PostMessageRequest struct {
	RecipientAgentID    string                 `json:"recipient_agent_id"`
	MessageType         MessageType            `json:"message_type"`
	Objective           string                 `json:"objective"`
	RequestedCapability string                 `json:"requested_capability,omitempty"`
	Payload             map[string]interface{} `json:"payload"`
	Priority            TaskPriority           `json:"priority,omitempty"`
	IdempotencyKey      string                 `json:"idempotency_key,omitempty"`
}

type AddContextItemRequest struct {
	ItemType      ContextItemType        `json:"item_type"`
	EntityType    string                 `json:"entity_type,omitempty"`
	EntityID      string                 `json:"entity_id,omitempty"`
	SourceAgentID string                 `json:"source_agent_id,omitempty"`
	ProvenanceID  string                 `json:"provenance_id,omitempty"`
	Content       map[string]interface{} `json:"content"`
	Confidence    float64                `json:"confidence,omitempty"`
}

type InitiateHandoffRequest struct {
	DestinationAgentID string                 `json:"destination_agent_id"`
	Reason             string                 `json:"reason"`
	Objective          string                 `json:"objective"`
	RequiredCapability string                 `json:"required_capability"`
	CurrentFindings    map[string]interface{} `json:"current_findings,omitempty"`
	ExpectedOutput     string                 `json:"expected_output"`
	Confidence         float64                `json:"confidence,omitempty"`
}

type TaskFilter struct {
	OrgID           int64
	Status          string
	AssignedAgentID string
	ParentTaskID    *string
	RootTaskID      *string
	CorrelationID   string
	Limit           int
	Offset          int
}

// Sidecar Integration DTOs
type SidecarContextReference struct {
	ContextID       string                 `json:"context_id"`
	ItemType        string                 `json:"item_type"`
	EntityType      string                 `json:"entity_type,omitempty"`
	EntityID        string                 `json:"entity_id,omitempty"`
	SourceAgentID   string                 `json:"source_agent_id,omitempty"`
	ProvenanceID    string                 `json:"provenance_id,omitempty"`
	Content         map[string]interface{} `json:"content"`
	Confidence      float64                `json:"confidence"`
	IsAuthoritative bool                   `json:"is_authoritative"`
}

type SidecarTaskRequest struct {
	TaskID               string                    `json:"task_id"`
	OrgID                int64                     `json:"org_id"`
	Objective            string                    `json:"objective"`
	InitiatingEvent      string                    `json:"initiating_event,omitempty"`
	ParentTaskID         string                    `json:"parent_task_id,omitempty"`
	RootTaskID           string                    `json:"root_task_id,omitempty"`
	AssignedAgentID      string                    `json:"assigned_agent_id"`
	Status               string                    `json:"status"`
	Priority             string                    `json:"priority"`
	RequiredCapabilities []string                  `json:"required_capabilities"`
	ContextReferences    []SidecarContextReference `json:"context_references"`
	Dependencies         []string                  `json:"dependencies"`
	CorrelationID        string                    `json:"correlation_id"`
}

type SidecarProposedDelegation struct {
	TargetAgentID        string   `json:"target_agent_id"`
	Objective            string   `json:"objective"`
	RequiredCapabilities []string `json:"required_capabilities"`
	ContextReferences    []string `json:"context_references"`
	ExpectedOutput       string   `json:"expected_output"`
	Priority             string   `json:"priority"`
}

type SidecarProposedAction struct {
	ActionType       string                 `json:"action_type"`
	EntityType       string                 `json:"entity_type"`
	EntityID         string                 `json:"entity_id"`
	Parameters       map[string]interface{} `json:"parameters"`
	RiskLevel        string                 `json:"risk_level"`
	RequiresApproval bool                   `json:"requires_approval"`
	Reasoning        string                 `json:"reasoning"`
}

type SidecarProposedHandoff struct {
	OriginatingTaskID  string                 `json:"originating_task_id"`
	SourceAgentID      string                 `json:"source_agent_id"`
	DestinationAgentID string                 `json:"destination_agent_id"`
	Reason             string                 `json:"reason"`
	Objective          string                 `json:"objective"`
	RequiredCapability string                 `json:"required_capability"`
	ContextReferences  []string               `json:"context_references"`
	CurrentFindings    map[string]interface{} `json:"current_findings"`
	ExpectedOutput     string                 `json:"expected_output"`
	Confidence         float64                `json:"confidence"`
}

type SidecarMessageItem struct {
	MessageID          string                 `json:"message_id"`
	TaskID             string                 `json:"task_id"`
	ParentTaskID       string                 `json:"parent_task_id,omitempty"`
	SenderAgentID      string                 `json:"sender_agent_id"`
	RecipientAgentID   string                 `json:"recipient_agent_id"`
	MessageType        string                 `json:"message_type"`
	Objective          string                 `json:"objective"`
	RequestedCapability string                `json:"requested_capability,omitempty"`
	Payload            map[string]interface{} `json:"payload"`
	Priority           string                 `json:"priority"`
	CorrelationID      string                 `json:"correlation_id"`
	CreatedAt          string                 `json:"created_at,omitempty"`
}

type SidecarTaskResponse struct {
	TaskID                  string                      `json:"task_id"`
	AgentID                 string                      `json:"agent_id"`
	Status                  string                      `json:"status"`
	Confidence              float64                     `json:"confidence"`
	Objective               string                      `json:"objective"`
	Findings                map[string]interface{}      `json:"findings"`
	Summary                 string                      `json:"summary"`
	Facts                   []map[string]interface{}    `json:"facts"`
	KnownFacts              []map[string]interface{}    `json:"known_facts,omitempty"`
	LikelyCauses            []map[string]interface{}    `json:"likely_causes,omitempty"`
	Predictions             []map[string]interface{}    `json:"predictions"`
	Recommendations         []map[string]interface{}    `json:"recommendations"`
	RecoveryOptions         []RecoveryOption            `json:"recovery_options,omitempty"`
	Evidence                []string                    `json:"evidence"`
	RequestedFollowUpAgents []string                    `json:"requested_follow_up_agents"`
	EscalationIndicator     bool                        `json:"escalation_indicator"`
	ProposedDelegations     []SidecarProposedDelegation `json:"proposed_delegations"`
	ProposedHandoff         *SidecarProposedHandoff     `json:"proposed_handoff,omitempty"`
	ProposedActions         []SidecarProposedAction     `json:"proposed_actions"`
	NewContextItems         []map[string]interface{}    `json:"new_context_items"`
	Messages                []SidecarMessageItem        `json:"messages"`
	DecisionRecord          *DecisionRecord             `json:"decision_record,omitempty"`
	CollaborativePlan       *CollaborativePlan          `json:"collaborative_plan,omitempty"`
	CrossModuleRisk         *CrossModuleRisk            `json:"cross_module_risk,omitempty"`
	RiskAssessments         []CrossModuleRisk           `json:"risk_assessments,omitempty"`
	ErrorCode               string                      `json:"error_code,omitempty"`
	ErrorMessage            string                      `json:"error_message,omitempty"`
	CreatedAt               string                      `json:"created_at,omitempty"`
}

// Phase 6.5: Structured Recovery Options
type RecoveryOption struct {
	OptionID             string   `json:"option_id"`
	Title                string   `json:"title"`
	Description          string   `json:"description"`
	ExpectedBenefit      string   `json:"expected_benefit"`
	OperationalImpact    string   `json:"operational_impact"`
	CustomerImpact       string   `json:"customer_impact"`
	FinancialImpact      string   `json:"financial_impact"`
	Risks                string   `json:"risks"`
	Confidence           float64  `json:"confidence"`
	RequiredCapabilities []string `json:"required_capabilities"`
	RequiresApproval     bool     `json:"requires_approval"`
}

// Phase 6.4 & 6.5: Collaborative Planning & Decision Making models

type SpecialistContribution struct {
	AgentID             string                   `json:"agent_id"`
	TaskID              string                   `json:"task_id"`
	Confidence          float64                  `json:"confidence"`
	Facts               []map[string]interface{} `json:"facts"`
	Predictions         []map[string]interface{} `json:"predictions"`
	Recommendations     []map[string]interface{} `json:"recommendations"`
	Evidence            []string                 `json:"evidence"`
	Warnings            []string                 `json:"warnings"`
	Errors              []string                 `json:"errors"`
	EscalationIndicator bool                     `json:"escalation_indicator"`
	Timestamp           string                   `json:"timestamp,omitempty"`
}

type ConflictRecord struct {
	ConflictID             string                 `json:"conflict_id"`
	TenantID               int64                  `json:"tenant_id,omitempty"`
	TaskID                 string                 `json:"task_id,omitempty"`
	ParentTaskID           string                 `json:"parent_task_id,omitempty"`
	Topic                  string                 `json:"topic,omitempty"`
	AgentA                 string                 `json:"agent_a,omitempty"`
	RecommendationA        map[string]interface{} `json:"recommendation_a,omitempty"`
	AgentB                 string                 `json:"agent_b,omitempty"`
	RecommendationB        map[string]interface{} `json:"recommendation_b,omitempty"`
	ParticipatingAgents    []string               `json:"participating_agents,omitempty"`
	ConflictingResults     map[string]interface{} `json:"conflicting_results,omitempty"`
	ConflictType           string                 `json:"conflict_type,omitempty"` // factual, prediction, recommendation, priority, business_constraint, incomplete_information, policy
	Evidence               []string               `json:"evidence,omitempty"`
	ConfidenceValues       map[string]float64     `json:"confidence_values,omitempty"`
	BusinessConstraints    []string               `json:"business_constraints,omitempty"`
	ResolutionStatus       string                 `json:"resolution_status,omitempty"` // RESOLVED, UNRESOLVED, ESCALATED_TO_HUMAN, PENDING_DECISION
	ResolutionStrategy     string                 `json:"resolution_strategy"`         // CONSENSUS, ESCALATE_TO_HUMAN, BUSINESS_RULE_PRIORITY, AUTHORITATIVE_DATA_OVERRIDE, DATA_FRESHNESS_PRECEDENCE
	IsResolved             bool                   `json:"is_resolved"`
	ResolvedRecommendation map[string]interface{} `json:"resolved_recommendation,omitempty"`
	SelectedOutcome        map[string]interface{} `json:"selected_outcome,omitempty"`
	UnresolvedReason       string                 `json:"unresolved_reason,omitempty"`
	EscalationRequirement  bool                   `json:"escalation_requirement,omitempty"`
	Reasoning              string                 `json:"reasoning,omitempty"`
	Severity               string                 `json:"severity,omitempty"` // LOW, MEDIUM, HIGH, CRITICAL
	CreatedAt              string                 `json:"created_at,omitempty"`
	ResolvedAt             string                 `json:"resolved_at,omitempty"`
}

type DecisionRecord struct {
	DecisionID              string                   `json:"decision_id"`
	Objective               string                   `json:"objective"`
	ParticipatingAgents     []string                 `json:"participating_agents"`
	SpecialistContributions []SpecialistContribution `json:"specialist_contributions"`
	Conflicts               []ConflictRecord         `json:"conflicts"`
	SelectedRecommendation  map[string]interface{}   `json:"selected_recommendation,omitempty"`
	ConsensusSummary        string                   `json:"consensus_summary"`
	OverallConfidence       float64                  `json:"overall_confidence"`
	ConfidenceRationale     string                   `json:"confidence_rationale"`
	UncertaintyIndicators   []string                 `json:"uncertainty_indicators"`
	Assumptions             []string                 `json:"assumptions"`
	RequiresHumanApproval   bool                     `json:"requires_human_approval"`
	ApprovalReason          string                   `json:"approval_reason,omitempty"`
	ProposedActions         []SidecarProposedAction  `json:"proposed_actions"`
	RecoveryOptions         []RecoveryOption         `json:"recovery_options,omitempty"`
	QuotationRecommendation map[string]interface{}   `json:"quotation_recommendation,omitempty"`
	CrossModuleRisk         *CrossModuleRisk         `json:"cross_module_risk,omitempty"`
	RiskAssessments         []CrossModuleRisk        `json:"risk_assessments,omitempty"`
	MissingData             []string                 `json:"missing_data,omitempty"`
	CorrelationID           string                   `json:"correlation_id"`
	CreatedAt               string                   `json:"created_at,omitempty"`
}

type CollaborativePlanStep struct {
	StepID               string                 `json:"step_id"`
	AgentID              string                 `json:"agent_id"`
	Objective            string                 `json:"objective"`
	Dependencies         []string               `json:"dependencies"`
	RequiredCapabilities []string               `json:"required_capabilities"`
	ExpectedOutput       string                 `json:"expected_output"`
	IsRequired           bool                   `json:"is_required"`
	Status               string                 `json:"status"` // PENDING, RUNNING, COMPLETED, FAILED, SKIPPED
	TaskID               string                 `json:"task_id,omitempty"`
	Result               map[string]interface{} `json:"result,omitempty"`
	Confidence           float64                `json:"confidence"`
	ErrorMessage         string                 `json:"error_message,omitempty"`
}

type CollaborativePlan struct {
	PlanID              string                  `json:"plan_id"`
	PlanningTaskID      string                  `json:"planning_task_id"`
	Objective           string                  `json:"objective"`
	CoordinatorAgentID  string                  `json:"coordinator_agent_id"`
	ParticipatingAgents []string                `json:"participating_agents"`
	Steps               []CollaborativePlanStep `json:"steps"`
	CompletedSteps      []string                `json:"completed_steps"`
	FailedSteps         []string                `json:"failed_steps"`
	Status              string                  `json:"status"` // CREATED, IN_PROGRESS, COMPLETED, FAILED, WAITING_APPROVAL
	FinalDecision       *DecisionRecord         `json:"final_decision,omitempty"`
	CrossModuleRisk     *CrossModuleRisk        `json:"cross_module_risk,omitempty"`
	RiskAssessments     []CrossModuleRisk       `json:"risk_assessments,omitempty"`
	OverallConfidence   float64                 `json:"overall_confidence"`
	EscalationState     bool                    `json:"escalation_state"`
	RequiresApproval    bool                    `json:"requires_approval"`
	Version             int                     `json:"version"`
	SupersededPlanID    string                  `json:"superseded_plan_id,omitempty"`
	TriggerEvent        string                  `json:"trigger_event,omitempty"`
	RecoveryOptions     []RecoveryOption        `json:"recovery_options,omitempty"`
	CreatedAt           string                  `json:"created_at,omitempty"`
}

type CreateCollaborativePlanRequest struct {
	Objective          string                  `json:"objective"`
	CoordinatorAgentID string                  `json:"coordinator_agent_id,omitempty"`
	Steps              []CollaborativePlanStep `json:"steps,omitempty"`
	ContextReferences  []string                `json:"context_references,omitempty"`
	ContextItems       []AddContextItemRequest `json:"context_items,omitempty"`
	CorrelationID      string                  `json:"correlation_id,omitempty"`
}

// Phase 6.5: Operational Request Contracts
type AssessShipmentHealthRequest struct {
	ShipmentID    string                 `json:"shipment_id"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	ExtraFacts    map[string]interface{} `json:"extra_facts,omitempty"`
}

type InvestigateExceptionRequest struct {
	ShipmentID    string                 `json:"shipment_id"`
	ExceptionID   string                 `json:"exception_id"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	ExtraFacts    map[string]interface{} `json:"extra_facts,omitempty"`
}

type ReplanOperationRequest struct {
	OriginalPlanID string                 `json:"original_plan_id"`
	TriggerEvent   string                 `json:"trigger_event"`
	NewFacts       map[string]interface{} `json:"new_facts,omitempty"`
	CorrelationID  string                 `json:"correlation_id,omitempty"`
}

type ShipmentEventWorkflowRequest struct {
	EventType     string                 `json:"event_type"` // e.g. MILESTONE_DELAY, ETA_RISK_INCREASE, EXCEPTION_CREATED, EXCEPTION_SEVERITY_CHANGED, CARRIER_UPDATE
	ShipmentID    string                 `json:"shipment_id"`
	EventData     map[string]interface{} `json:"event_data,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
}

// Phase 6.6: Commercial Request Contracts
type AssessCustomerRequest struct {
	CustomerID    string                 `json:"customer_id"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	ExtraFacts    map[string]interface{} `json:"extra_facts,omitempty"`
}

type EvaluateLeadRequest struct {
	LeadID        string                 `json:"lead_id"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	ExtraFacts    map[string]interface{} `json:"extra_facts,omitempty"`
}

type EvaluateRFQRequest struct {
	RFQID         string                 `json:"rfq_id"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	ExtraFacts    map[string]interface{} `json:"extra_facts,omitempty"`
}

type AssessCollectionsRequest struct {
	CustomerID    string                 `json:"customer_id,omitempty"`
	InvoiceID     string                 `json:"invoice_id,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	ExtraFacts    map[string]interface{} `json:"extra_facts,omitempty"`
}

type CommercialEventWorkflowRequest struct {
	EventType     string                 `json:"event_type"` // e.g. NEW_RFQ, QUOTATION_CREATED, QUOTATION_AGING, INVOICE_OVERDUE, PAYMENT_RECEIVED, LEAD_STATUS_CHANGED
	EntityID      string                 `json:"entity_id"`
	EventData     map[string]interface{} `json:"event_data,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
}

// Phase 6.7: Cross-Module Risk Model & Requests

type CrossModuleRisk struct {
	RiskID                string                   `json:"risk_id"`
	TenantID              *int64                   `json:"tenant_id,omitempty"`
	OriginatingTaskID     string                   `json:"originating_task_id"`
	RiskType              string                   `json:"risk_type"` // operational, shipment, exception, customer, pricing, financial, contract, compliance, documentation, commercial, cross_module
	Severity              string                   `json:"severity"`  // low, medium, high, critical
	Probability           float64                  `json:"probability"`
	Impact                string                   `json:"impact"`
	AffectedModules       []string                 `json:"affected_modules"`
	AffectedEntities      map[string]interface{}   `json:"affected_entities"`
	ContributingAgents    []string                 `json:"contributing_agents"`
	EvidenceReferences   []string                 `json:"evidence_references"`
	Facts                 []map[string]interface{} `json:"facts"`
	Predictions           []map[string]interface{} `json:"predictions"`
	Recommendations       []map[string]interface{} `json:"recommendations"`
	Confidence            float64                  `json:"confidence"`
	Uncertainty           []string                 `json:"uncertainty"`
	MissingData           []string                 `json:"missing_data"`
	Status                string                   `json:"status"` // IDENTIFIED, UNDER_REVIEW, ESCALATED, RESOLVED, SUPERSEDED
	EscalationRequirement bool                     `json:"escalation_requirement"`
	ApprovalRequirement   bool                     `json:"approval_requirement"`
	ParentRiskID          *string                  `json:"parent_risk_id,omitempty"`
	Version               int                      `json:"version"`
	SupersededRiskID      *string                  `json:"superseded_risk_id,omitempty"`
	CreatedAt             string                   `json:"created_at,omitempty"`
	UpdatedAt             string                   `json:"updated_at,omitempty"`
}

type AssessContractRiskRequest struct {
	ContractID    string                 `json:"contract_id"`
	ShipmentID    string                 `json:"shipment_id,omitempty"`
	Objective     string                 `json:"objective,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	ExtraFacts    map[string]interface{} `json:"extra_facts,omitempty"`
}

type AssessComplianceRiskRequest struct {
	ShipmentID    string                 `json:"shipment_id"`
	DocumentID    string                 `json:"document_id,omitempty"`
	Objective     string                 `json:"objective,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	ExtraFacts    map[string]interface{} `json:"extra_facts,omitempty"`
}

type AssessCrossModuleRiskRequest struct {
	ShipmentID    string                 `json:"shipment_id"`
	ExceptionID   string                 `json:"exception_id,omitempty"`
	ContractID    string                 `json:"contract_id,omitempty"`
	Objective     string                 `json:"objective,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	ExtraFacts    map[string]interface{} `json:"extra_facts,omitempty"`
}

type ReassessRiskRequest struct {
	OriginalPlanID string                 `json:"original_plan_id"`
	PriorRiskID    string                 `json:"prior_risk_id,omitempty"`
	TriggerEvent   string                 `json:"trigger_event"`
	NewFacts       map[string]interface{} `json:"new_facts,omitempty"`
	CorrelationID  string                 `json:"correlation_id,omitempty"`
}

// ----------------------------------------------------------------------
// Phase 6.8: Conflict Resolution, Memory & Learning
// ----------------------------------------------------------------------

type AgentOutcome struct {
	ID               int64                  `json:"id,omitempty"`
	OutcomeID        string                 `json:"outcome_id"`
	TenantID         int64                  `json:"tenant_id"`
	SourceTaskID     string                 `json:"source_task_id,omitempty"`
	SourceAgentID    string                 `json:"source_agent_id,omitempty"`
	SourceEntityType string                 `json:"source_entity_type"`
	SourceEntityID   string                 `json:"source_entity_id"`
	WorkflowID       string                 `json:"workflow_id,omitempty"`
	PlanID           string                 `json:"plan_id,omitempty"`
	Objective        string                 `json:"objective,omitempty"`
	Recommendation   map[string]interface{} `json:"recommendation,omitempty"`
	ActionType       string                 `json:"action_type,omitempty"`
	HumanDecision    string                 `json:"human_decision,omitempty"`
	HumanFeedback    string                 `json:"human_feedback,omitempty"`
	ActualOutcome    string                 `json:"actual_outcome,omitempty"`
	Status           string                 `json:"status"` // SUCCESS, FAILED, PARTIAL_SUCCESS, UNVERIFIED
	IsVerified       bool                   `json:"is_verified"`
	SuccessIndicator bool                   `json:"success_indicator"`
	Evidence         []string               `json:"evidence,omitempty"`
	Lesson           string                 `json:"lesson,omitempty"`
	Confidence       float64                `json:"confidence"`
	CorrelationID    string                 `json:"correlation_id,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        string                 `json:"created_at,omitempty"`
}

type MemoryItemDTO struct {
	ID           int64                  `json:"id"`
	OrgID        int64                  `json:"org_id"`
	MemoryType   string                 `json:"memory_type"`
	Category     string                 `json:"category"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	Confidence   float64                `json:"confidence"`
	Status       string                 `json:"status"`
	EntityType   string                 `json:"entity_type,omitempty"`
	EntityID     string                 `json:"entity_id,omitempty"`
	TimesUsed    int                    `json:"times_used"`
	SuccessCount int                    `json:"success_count"`
	FailureCount int                    `json:"failure_count"`
	IsStale      bool                   `json:"is_stale"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    string                 `json:"created_at"`
}

type QueryMemoryRequest struct {
	Domain     string `json:"domain,omitempty"`      // shipment, pricing, customer, finance, contract, compliance
	EntityType string `json:"entity_type,omitempty"` // SHIPMENT, EXCEPTION, INVOICE, RFQ, CONTRACT
	EntityID   string `json:"entity_id,omitempty"`
	Query      string `json:"query,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

type MemoryQueryResponse struct {
	Domain        string          `json:"domain"`
	TotalFound    int             `json:"total_found"`
	Items         []MemoryItemDTO `json:"items"`
	PrecedenceMsg string          `json:"precedence_msg"`
}

type ResolveConflictRequest struct {
	ConflictID       string                 `json:"conflict_id,omitempty"`
	ConflictType     string                 `json:"conflict_type"`
	Participating    []string               `json:"participating_agents"`
	Findings         map[string]interface{} `json:"findings"`
	AuthoritativeKey string                 `json:"authoritative_key,omitempty"`
	HumanDecision    string                 `json:"human_decision,omitempty"`
	Reason           string                 `json:"reason,omitempty"`
}

type RecordOutcomeRequest struct {
	SourceTaskID     string                 `json:"source_task_id,omitempty"`
	SourceAgentID    string                 `json:"source_agent_id,omitempty"`
	SourceEntityType string                 `json:"source_entity_type"`
	SourceEntityID   string                 `json:"source_entity_id"`
	WorkflowID       string                 `json:"workflow_id,omitempty"`
	PlanID           string                 `json:"plan_id,omitempty"`
	Objective        string                 `json:"objective,omitempty"`
	Recommendation   map[string]interface{} `json:"recommendation,omitempty"`
	ActionType       string                 `json:"action_type,omitempty"`
	HumanDecision    string                 `json:"human_decision,omitempty"`
	HumanFeedback    string                 `json:"human_feedback,omitempty"`
	ActualOutcome    string                 `json:"actual_outcome,omitempty"`
	Status           string                 `json:"status"`
	IsVerified       bool                   `json:"is_verified"`
	SuccessIndicator bool                   `json:"success_indicator"`
	Evidence         []string               `json:"evidence,omitempty"`
	Lesson           string                 `json:"lesson,omitempty"`
	Confidence       float64                `json:"confidence"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ----------------------------------------------------------------------
// Phase 6.9: Governed Multi-Agent Autonomy & Workforce Command Center
// ----------------------------------------------------------------------

type ActionPolicyDecision struct {
	AgentID            string `json:"agent_id"`
	AutonomyLevel      string `json:"autonomy_level"`
	ActionType         string `json:"action_type"`
	Allowed            bool   `json:"allowed"`
	CanRecommend       bool   `json:"can_recommend"`
	CanPrepare         bool   `json:"can_prepare"`
	CanAutoExecute     bool   `json:"can_auto_execute"`
	RequiresApproval   bool   `json:"requires_approval"`
	RequiresEscalation bool   `json:"requires_escalation"`
	Decision           string `json:"decision"`
	Reason             string `json:"reason"`
	PolicyReason       string `json:"policy_reason,omitempty"`
	RiskLevel          string `json:"risk_level"` // LOW, MEDIUM, HIGH, CRITICAL
	Timestamp          string `json:"timestamp"`
}

type AgentWorkloadMetrics struct {
	AgentID           string     `json:"agent_id"`
	Name              string     `json:"name"`
	AgentType         string     `json:"agent_type"`
	AutonomyLevel     string     `json:"autonomy_level"`
	OperationalStatus string     `json:"operational_status"` // ACTIVE, PAUSED, DISABLED
	HealthStatus      string     `json:"health_status"`      // HEALTHY, DEGRADED, FAILING, DISABLED, UNKNOWN
	PendingTasks      int        `json:"pending_tasks"`
	RunningTasks      int        `json:"running_tasks"`
	WaitingTasks      int        `json:"waiting_tasks"`
	CompletedTasks    int        `json:"completed_tasks"`
	FailedTasks       int        `json:"failed_tasks"`
	AverageLatencyMs  float64    `json:"average_latency_ms"`
	FailureCount      int        `json:"failure_count"`
	CurrentTaskID     string     `json:"current_task_id,omitempty"`
	LastExecutionAt   *time.Time `json:"last_execution_at,omitempty"`
	IsBottleneck      bool       `json:"is_bottleneck"`
	IsOverloaded      bool       `json:"is_overloaded"`
}

type WorkforceHealthSummary struct {
	OverallStatus       string  `json:"overall_status"` // HEALTHY, DEGRADED, CRITICAL, EMERGENCY_STOPPED
	ActiveAgentsCount   int     `json:"active_agents_count"`
	PausedAgentsCount   int     `json:"paused_agents_count"`
	DisabledAgentsCount int     `json:"disabled_agents_count"`
	TotalTasks          int     `json:"total_tasks"`
	PendingTasks        int     `json:"pending_tasks"`
	RunningTasks        int     `json:"running_tasks"`
	WaitingTasks        int     `json:"waiting_tasks"`
	BlockedTasks        int     `json:"blocked_tasks"`
	FailedTasks         int     `json:"failed_tasks"`
	CompletedTasks      int     `json:"completed_tasks"`
	ApprovalBacklog     int     `json:"approval_backlog"`
	EscalationBacklog   int     `json:"escalation_backlog"`
	AverageLatencyMs    float64 `json:"average_latency_ms"`
	EmergencyStopActive bool    `json:"emergency_stop_active"`
	EmergencyStopReason string  `json:"emergency_stop_reason,omitempty"`
}

type SingleStop struct {
	Scope       string `json:"scope"` // WORKFORCE, AGENT, WORKFLOW, ACTION_CLASS
	Target      string `json:"target,omitempty"`
	TargetID    string `json:"target_id,omitempty"`
	Reason      string `json:"reason"`
	TriggeredBy string `json:"triggered_by"`
	TriggeredAt string `json:"triggered_at"`
}

type EmergencyStopStatus struct {
	OrgID            int64        `json:"org_id"`
	IsActive         bool         `json:"is_active"`
	WorkforceStopped bool         `json:"workforce_stopped"`
	Scope            string       `json:"scope,omitempty"`
	Target           string       `json:"target,omitempty"`
	TargetID         string       `json:"target_id,omitempty"`
	Reason           string       `json:"reason,omitempty"`
	TriggeredBy      string       `json:"triggered_by,omitempty"`
	TriggeredAt      string       `json:"triggered_at,omitempty"`
	ActiveStops      []SingleStop `json:"active_stops"`
}

type EmergencyStopRequest struct {
	Action   string `json:"action"` // STOP or RESUME
	Scope    string `json:"scope"`  // WORKFORCE, AGENT, WORKFLOW, ACTION_CLASS
	Target   string `json:"target,omitempty"`
	TargetID string `json:"target_id,omitempty"`
	Reason   string `json:"reason"`
}

type AgentControlRequest struct {
	Action           string  `json:"action"` // ENABLE, DISABLE, PAUSE, RESUME, SET_AUTONOMY
	AutonomyLevel    string  `json:"autonomy_level,omitempty"`
	NewAutonomyLevel *string `json:"new_autonomy_level,omitempty"`
	Reason           string  `json:"reason"`
}

type WorkflowControlRequest struct {
	Command string `json:"command"` // PAUSE, RESUME, STOP
	Action  string `json:"action"`  // alias for command
	Reason  string `json:"reason"`
}

type WorkforceApprovalItem struct {
	ApprovalID          int64                   `json:"approval_id"`
	PlanID              string                  `json:"plan_id,omitempty"`
	WorkflowID          string                  `json:"workflow_id,omitempty"`
	TaskID              string                  `json:"task_id,omitempty"`
	AgentID             string                  `json:"agent_id"`
	InitiatingAgent     string                  `json:"initiating_agent,omitempty"`
	ProposedAction      string                  `json:"proposed_action"`
	Objective           string                  `json:"objective,omitempty"`
	Reason              string                  `json:"reason"`
	Confidence          float64                 `json:"confidence"`
	Risk                string                  `json:"risk,omitempty"`
	AffectedEntity      string                  `json:"affected_entity,omitempty"`
	ApprovalRequirement string                  `json:"approval_requirement,omitempty"`
	ProposedActions     []SidecarProposedAction `json:"proposed_actions,omitempty"`
	Status              string                  `json:"status"`
	CreatedAt           string                  `json:"created_at"`
}

type WorkforceEscalationItem struct {
	EscalationID   string                 `json:"escalation_id"`
	WorkflowID     string                 `json:"workflow_id,omitempty"`
	PlanID         string                 `json:"plan_id,omitempty"`
	TaskID         string                 `json:"task_id,omitempty"`
	AgentID        string                 `json:"agent_id"`
	Reason         string                 `json:"reason"`
	Severity       string                 `json:"severity"` // MEDIUM, HIGH, CRITICAL
	Status         string                 `json:"status"`   // OPEN, RESOLVED, DISMISSED, BLOCKED
	Evidence       map[string]interface{} `json:"evidence,omitempty"`
	ConflictRecord *ConflictRecord        `json:"conflict_record,omitempty"`
	CreatedAt      string                 `json:"created_at"`
}

type WorkforceActivityItem struct {
	ActivityID string `json:"activity_id"`
	Timestamp  string `json:"timestamp"`
	EventType  string `json:"event_type"`
	AgentID    string `json:"agent_id"`
	ActorID    string `json:"actor_id,omitempty"`
	TaskID     string `json:"task_id,omitempty"`
	WorkflowID string `json:"workflow_id,omitempty"`
	Summary    string `json:"summary"`
	Status     string `json:"status,omitempty"`
	Severity   string `json:"severity,omitempty"`
}

type WorkflowInspectionDetail struct {
	PlanID                string                  `json:"plan_id"`
	Objective             string                  `json:"objective"`
	CoordinatorAgentID    string                  `json:"coordinator_agent_id"`
	ParticipatingAgents   []string                `json:"participating_agents"`
	Status                string                  `json:"status"`
	OverallConfidence     float64                 `json:"overall_confidence"`
	Steps                 []CollaborativePlanStep `json:"steps"`
	FinalDecision         *DecisionRecord         `json:"final_decision,omitempty"`
	RequiresApproval      bool                    `json:"requires_approval"`
	ApprovalStatus        *string                 `json:"approval_status,omitempty"`
	CreatedAt             string                  `json:"created_at"`
}

type CommandCenterOverview struct {
	Health           *WorkforceHealthSummary   `json:"health"`
	AgentWorkload    []AgentWorkloadMetrics    `json:"agent_workload"`
	EmergencyStop    *EmergencyStopStatus      `json:"emergency_stop"`
	ActiveWorkflows  int                       `json:"active_workflows"`
	WaitingApprovals []WorkforceApprovalItem   `json:"waiting_approvals"`
	Escalations      []WorkforceEscalationItem `json:"escalations"`
	RecentActivity   []WorkforceActivityItem   `json:"recent_activity"`
	AutonomySummary  map[string]int            `json:"autonomy_summary"`
}





