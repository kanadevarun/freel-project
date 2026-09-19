package automations

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Supported Automation Types
const (
	AutomationTypeDailyOverdueInvoiceReview         = "DAILY_OVERDUE_INVOICE_REVIEW"
	AutomationTypeDailyContractDocumentExpiryReview = "DAILY_CONTRACT_DOCUMENT_EXPIRY_REVIEW"
	AutomationTypeShipmentExceptionReview           = "SHIPMENT_EXCEPTION_REVIEW"
	AutomationTypeRFQQuotationReview               = "RFQ_QUOTATION_REVIEW"
	AutomationTypeCustomerFollowupReview            = "CUSTOMER_FOLLOWUP_REVIEW"
	AutomationTypeDailyOperationalSummary           = "DAILY_OPERATIONAL_SUMMARY"
)

// Schedule Types
const (
	ScheduleTypeDaily  = "DAILY"
	ScheduleTypeWeekly = "WEEKLY"
	ScheduleTypeHourly = "HOURLY"
)

// Supported Trigger Types (Phase 3 Event-Driven & Scheduled Foundation)
const (
	TriggerTypeScheduled                 = "SCHEDULED"
	TriggerTypeManual                    = "MANUAL"
	TriggerTypeRecordCreated             = "RECORD_CREATED"
	TriggerTypeRecordUpdated             = "RECORD_UPDATED"
	TriggerTypeStatusChanged             = "STATUS_CHANGED"
	TriggerTypeMilestoneMissed           = "MILESTONE_MISSED"
	TriggerTypeShipmentExceptionDetected = "SHIPMENT_EXCEPTION_DETECTED"
	TriggerTypeInvoiceOverdue            = "INVOICE_OVERDUE"
	TriggerTypeContractExpiring          = "CONTRACT_EXPIRING"
	TriggerTypeRFQDeadlineApproaching    = "RFQ_DEADLINE_APPROACHING"
	TriggerTypeApprovalReturned          = "APPROVAL_RETURNED"
)

// Execution Statuses (Phase 3 Granular Lifecycle)
const (
	ExecutionStatusPending            = "PENDING"
	ExecutionStatusQueued             = "QUEUED"
	ExecutionStatusRunning            = "RUNNING"
	ExecutionStatusWaitingForApproval = "WAITING_FOR_APPROVAL"
	ExecutionStatusExecutingAction    = "EXECUTING_ACTION"
	ExecutionStatusCompleted          = "COMPLETED"
	ExecutionStatusPartiallyCompleted = "PARTIALLY_COMPLETED"
	ExecutionStatusFailed             = "FAILED"
	ExecutionStatusCancelled          = "CANCELLED"
	ExecutionStatusExpired            = "EXPIRED"
	ExecutionStatusRetrying           = "RETRYING"
	ExecutionStatusSkipped            = "SKIPPED"
)

// Execution Lifecycle Steps
const (
	StepInitializing       = "INITIALIZING"
	StepEvaluatingRules    = "EVALUATING_RULES"
	StepGeneratingInsights = "GENERATING_INSIGHTS"
	StepWaitingApproval    = "WAITING_FOR_APPROVAL"
	StepExecutingAction    = "EXECUTING_ACTION"
	StepVerifyingOutcome   = "VERIFYING_OUTCOME"
	StepCompleted          = "COMPLETED"
	StepFailed             = "FAILED"
	StepCancelled          = "CANCELLED"
)

// Approval Policies
const (
	ApprovalPolicyAlwaysRequire = "ALWAYS_REQUIRE"
	ApprovalPolicyAutoIfLowRisk = "AUTO_IF_LOW_RISK"
	ApprovalPolicyManualOnly    = "MANUAL_ONLY"
)

// Insight Severities
const (
	InsightSeverityInfo     = "INFO"
	InsightSeverityLow      = "LOW"
	InsightSeverityMedium   = "MEDIUM"
	InsightSeverityHigh     = "HIGH"
	InsightSeverityCritical = "CRITICAL"
)

// Insight Statuses
const (
	InsightStatusActive       = "ACTIVE"
	InsightStatusAcknowledged = "ACKNOWLEDGED"
	InsightStatusActioned     = "ACTIONED"
	InsightStatusDismissed    = "DISMISSED"
)

// Automation represents a scheduled or event-driven workflow automation definition
type Automation struct {
	ID                     int64           `db:"id" json:"id"`
	OrgID                  int64           `db:"org_id" json:"org_id"`
	Name                   string          `db:"name" json:"name"`
	AutomationType         string          `db:"automation_type" json:"automation_type"`
	Description            *string         `db:"description" json:"description,omitempty"`
	IsEnabled              bool            `db:"is_enabled" json:"is_enabled"`
	TriggerType            string          `db:"trigger_type" json:"trigger_type"`
	TriggerConfigJSON      *string         `db:"trigger_config" json:"-"`
	TriggerConfig          json.RawMessage `db:"-" json:"trigger_config,omitempty"`
	Scope                  string          `db:"scope" json:"scope"`
	ApprovalPolicy         string          `db:"approval_policy" json:"approval_policy"`
	AllowedActionsJSON     *string         `db:"allowed_actions" json:"-"`
	AllowedActions         []string        `db:"-" json:"allowed_actions"`
	OwnerTeam              string          `db:"owner_team" json:"owner_team"`
	Priority               string          `db:"priority" json:"priority"`
	ScheduleType           string          `db:"schedule_type" json:"schedule_type"`
	ScheduleTime           string          `db:"schedule_time" json:"schedule_time"`
	ScheduleDays           *string         `db:"schedule_days" json:"schedule_days,omitempty"`
	Timezone               string          `db:"timezone" json:"timezone"`
	ExecutionWindowMinutes int             `db:"execution_window_minutes" json:"execution_window_minutes"`
	ConfigurationJSON      *string         `db:"configuration" json:"-"`
	Configuration          json.RawMessage `db:"-" json:"configuration,omitempty"`
	TargetModulesJSON      *string         `db:"target_modules" json:"-"`
	TargetModules          []string        `db:"-" json:"target_modules"`
	LastExecutionAt        *time.Time      `db:"last_execution_at" json:"last_execution_at,omitempty"`
	NextExecutionAt        *time.Time      `db:"next_execution_at" json:"next_execution_at,omitempty"`
	LastExecutionStatus    *string         `db:"last_execution_status" json:"last_execution_status,omitempty"`
	LastError              *string         `db:"last_error" json:"last_error,omitempty"`
	LastSuccessfulRun      *time.Time      `db:"last_successful_run" json:"last_successful_run,omitempty"`
	LastFailedRun          *time.Time      `db:"last_failed_run" json:"last_failed_run,omitempty"`
	RetryCount             int             `db:"retry_count" json:"retry_count"`
	MaxRetries             int             `db:"max_retries" json:"max_retries"`
	RetryPolicyJSON        *string         `db:"retry_policy" json:"-"`
	RetryPolicy            json.RawMessage `db:"-" json:"retry_policy,omitempty"`
	MaxExecutionDurationSec int            `db:"max_execution_duration_sec" json:"max_execution_duration_sec"`
	CorrelationID          *string         `db:"correlation_id" json:"correlation_id,omitempty"`
	CreatedBy              int64           `db:"created_by" json:"created_by"`
	UpdatedBy              int64           `db:"updated_by" json:"updated_by"`
	CreatedAt              time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time       `db:"updated_at" json:"updated_at"`

	// Derived fields for client display
	CreatorName       string `db:"creator_name" json:"creator_name,omitempty"`
	IsRunning         bool   `db:"-" json:"is_running"`
	ActiveExecutionID *int64 `db:"-" json:"active_execution_id,omitempty"`
}

// UnmarshalJSONFields parses JSON columns into Go representations
func (a *Automation) UnmarshalJSONFields() {
	if a.ConfigurationJSON != nil && *a.ConfigurationJSON != "" {
		a.Configuration = json.RawMessage(*a.ConfigurationJSON)
	} else {
		a.Configuration = json.RawMessage("{}")
	}

	if a.TriggerConfigJSON != nil && *a.TriggerConfigJSON != "" {
		a.TriggerConfig = json.RawMessage(*a.TriggerConfigJSON)
	} else {
		a.TriggerConfig = json.RawMessage("{}")
	}

	if a.RetryPolicyJSON != nil && *a.RetryPolicyJSON != "" {
		a.RetryPolicy = json.RawMessage(*a.RetryPolicyJSON)
	} else {
		a.RetryPolicy = json.RawMessage("{}")
	}

	if a.TargetModulesJSON != nil && *a.TargetModulesJSON != "" {
		var mods []string
		if err := json.Unmarshal([]byte(*a.TargetModulesJSON), &mods); err == nil {
			a.TargetModules = mods
		} else {
			a.TargetModules = []string{}
		}
	} else {
		a.TargetModules = []string{}
	}

	if a.AllowedActionsJSON != nil && *a.AllowedActionsJSON != "" {
		var acts []string
		if err := json.Unmarshal([]byte(*a.AllowedActionsJSON), &acts); err == nil {
			a.AllowedActions = acts
		} else {
			a.AllowedActions = []string{}
		}
	} else {
		a.AllowedActions = []string{}
	}
}

// AutomationExecution records an instance of an automation run with durable lifecycle
type AutomationExecution struct {
	ID                     int64           `db:"id" json:"id"`
	AutomationID           int64           `db:"automation_id" json:"automation_id"`
	OrgID                  int64           `db:"org_id" json:"org_id"`
	CorrelationID          string          `db:"correlation_id" json:"correlation_id"`
	TriggerType            string          `db:"trigger_type" json:"trigger_type"`
	TriggerEvent           *string         `db:"trigger_event" json:"trigger_event,omitempty"`
	InputRecordRef         *string         `db:"input_record_ref" json:"input_record_ref,omitempty"`
	CurrentStep            string          `db:"current_step" json:"current_step"`
	StepResultsJSON        *string         `db:"step_results" json:"-"`
	StepResults            json.RawMessage `db:"-" json:"step_results,omitempty"`
	TriggeredByUserID      *int64          `db:"triggered_by_user_id" json:"triggered_by_user_id,omitempty"`
	Status                 string          `db:"status" json:"status"`
	QueuedAt               time.Time       `db:"queued_at" json:"queued_at"`
	StartedAt              *time.Time      `db:"started_at" json:"started_at,omitempty"`
	CompletedAt            *time.Time      `db:"completed_at" json:"completed_at,omitempty"`
	DurationMs             int64           `db:"duration_ms" json:"duration_ms"`
	RecordsReviewed        int             `db:"records_reviewed" json:"records_reviewed"`
	RecommendationsCreated int             `db:"recommendations_created" json:"recommendations_created"`
	RecommendationsUpdated int             `db:"recommendations_updated" json:"recommendations_updated"`
	SummaryText            *string         `db:"summary_text" json:"summary_text,omitempty"`
	ErrorMessage           *string         `db:"error_message" json:"error_message,omitempty"`
	DetailsJSON            *string         `db:"details" json:"-"`
	Details                json.RawMessage `db:"-" json:"details,omitempty"`
	ApprovalID             *int64          `db:"approval_id" json:"approval_id,omitempty"`
	ActionID               *int64          `db:"action_id" json:"action_id,omitempty"`
	IdempotencyKey         *string         `db:"idempotency_key" json:"idempotency_key,omitempty"`
	RetryCount             int             `db:"retry_count" json:"retry_count"`
	CreatedAt              time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time       `db:"updated_at" json:"updated_at"`

	// Derived fields
	AutomationName  string  `db:"automation_name" json:"automation_name,omitempty"`
	AutomationType  string  `db:"automation_type" json:"automation_type,omitempty"`
	TriggeredByName *string `db:"triggered_by_name" json:"triggered_by_name,omitempty"`
}

// UnmarshalJSONFields parses execution JSON columns
func (e *AutomationExecution) UnmarshalJSONFields() {
	if e.DetailsJSON != nil && *e.DetailsJSON != "" {
		e.Details = json.RawMessage(*e.DetailsJSON)
	} else {
		e.Details = json.RawMessage("{}")
	}

	if e.StepResultsJSON != nil && *e.StepResultsJSON != "" {
		e.StepResults = json.RawMessage(*e.StepResultsJSON)
	} else {
		e.StepResults = json.RawMessage("{}")
	}
}

// OperationalInsight represents a durable, deterministic operational signal detected by the system
type OperationalInsight struct {
	ID                    int64           `db:"id" json:"id"`
	OrgID                 int64           `db:"org_id" json:"org_id"`
	AutomationID          *int64          `db:"automation_id" json:"automation_id,omitempty"`
	ExecutionID           *int64          `db:"execution_id" json:"execution_id,omitempty"`
	SourceModule          string          `db:"source_module" json:"source_module"`
	SourceRecordID        int64           `db:"source_record_id" json:"source_record_id"`
	SourceRecordRef       *string         `db:"source_record_ref" json:"source_record_ref,omitempty"`
	InsightType           string          `db:"insight_type" json:"insight_type"`
	Severity              string          `db:"severity" json:"severity"`
	Title                 string          `db:"title" json:"title"`
	Description           string          `db:"description" json:"description"`
	EvidenceJSON          *string         `db:"evidence" json:"-"`
	Evidence              json.RawMessage `db:"-" json:"evidence,omitempty"`
	DetectionRule         string          `db:"detection_rule" json:"detection_rule"`
	Confidence            float64         `db:"confidence" json:"confidence"`
	Priority              string          `db:"priority" json:"priority"`
	RiskLevel             string          `db:"risk_level" json:"risk_level"`
	DataFreshness         string          `db:"data_freshness" json:"data_freshness"`
	RecommendedNextStep   *string         `db:"recommended_next_step" json:"recommended_next_step,omitempty"`
	IsApprovalRequired    bool            `db:"is_approval_required" json:"is_approval_required"`
	RecommendedActionType *string         `db:"recommended_action_type" json:"recommended_action_type,omitempty"`
	ActionPayloadJSON     *string         `db:"action_payload" json:"-"`
	ActionPayload         json.RawMessage `db:"-" json:"action_payload,omitempty"`
	Status                string          `db:"status" json:"status"`
	ActionID              *int64          `db:"action_id" json:"action_id,omitempty"`
	ApprovalID            *int64          `db:"approval_id" json:"approval_id,omitempty"`
	CorrelationID         string          `db:"correlation_id" json:"correlation_id"`
	CreatedAt             time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time       `db:"updated_at" json:"updated_at"`

	// Derived fields
	AutomationName string `db:"automation_name" json:"automation_name,omitempty"`
}

// UnmarshalJSONFields parses operational insight JSON columns
func (o *OperationalInsight) UnmarshalJSONFields() {
	if o.EvidenceJSON != nil && *o.EvidenceJSON != "" {
		o.Evidence = json.RawMessage(*o.EvidenceJSON)
	} else {
		o.Evidence = json.RawMessage("{}")
	}

	if o.ActionPayloadJSON != nil && *o.ActionPayloadJSON != "" {
		o.ActionPayload = json.RawMessage(*o.ActionPayloadJSON)
	} else {
		o.ActionPayload = json.RawMessage("{}")
	}
}

// CreateAutomationInput represents client request to create an automation definition
type CreateAutomationInput struct {
	Name                    string          `json:"name"`
	AutomationType          string          `json:"automation_type"`
	Description             *string         `json:"description,omitempty"`
	TriggerType             string          `json:"trigger_type,omitempty"`
	TriggerConfig           json.RawMessage `json:"trigger_config,omitempty"`
	Scope                   string          `json:"scope,omitempty"`
	ApprovalPolicy          string          `json:"approval_policy,omitempty"`
	AllowedActions          []string        `json:"allowed_actions,omitempty"`
	OwnerTeam               string          `json:"owner_team,omitempty"`
	Priority                string          `json:"priority,omitempty"`
	ScheduleType            string          `json:"schedule_type,omitempty"`
	ScheduleTime            string          `json:"schedule_time,omitempty"`
	ScheduleDays            *string         `json:"schedule_days,omitempty"`
	Timezone                string          `json:"timezone,omitempty"`
	ExecutionWindowMinutes  int             `json:"execution_window_minutes,omitempty"`
	Configuration           json.RawMessage `json:"configuration,omitempty"`
	TargetModules           []string        `json:"target_modules,omitempty"`
	IsEnabled               *bool           `json:"is_enabled,omitempty"`
	MaxExecutionDurationSec int             `json:"max_execution_duration_sec,omitempty"`
	RetryPolicy             json.RawMessage `json:"retry_policy,omitempty"`
}

// UpdateAutomationInput represents client request to update an automation
type UpdateAutomationInput struct {
	Name                    string          `json:"name"`
	Description             *string         `json:"description,omitempty"`
	TriggerType             string          `json:"trigger_type,omitempty"`
	TriggerConfig           json.RawMessage `json:"trigger_config,omitempty"`
	Scope                   string          `json:"scope,omitempty"`
	ApprovalPolicy          string          `json:"approval_policy,omitempty"`
	AllowedActions          []string        `json:"allowed_actions,omitempty"`
	OwnerTeam               string          `json:"owner_team,omitempty"`
	Priority                string          `json:"priority,omitempty"`
	ScheduleType            string          `json:"schedule_type,omitempty"`
	ScheduleTime            string          `json:"schedule_time,omitempty"`
	ScheduleDays            *string         `json:"schedule_days,omitempty"`
	Timezone                string          `json:"timezone,omitempty"`
	ExecutionWindowMinutes  int             `json:"execution_window_minutes,omitempty"`
	Configuration           json.RawMessage `json:"configuration,omitempty"`
	TargetModules           []string        `json:"target_modules,omitempty"`
	IsEnabled               *bool           `json:"is_enabled,omitempty"`
	MaxExecutionDurationSec int             `json:"max_execution_duration_sec,omitempty"`
	RetryPolicy             json.RawMessage `json:"retry_policy,omitempty"`
}

// AutomationFilter represents list filtering options
type AutomationFilter struct {
	Page           int    `json:"page"`
	Limit          int    `json:"limit"`
	AutomationType string `json:"automation_type"`
	TriggerType    string `json:"trigger_type"`
	Scope          string `json:"scope"`
	OwnerTeam      string `json:"owner_team"`
	IsEnabled      *bool  `json:"is_enabled"`
	Search         string `json:"search"`
	SortBy         string `json:"sort_by"`
	SortDir        string `json:"sort_dir"`
}

// ExecutionFilter represents execution history filter options
type ExecutionFilter struct {
	Page         int    `json:"page"`
	Limit        int    `json:"limit"`
	AutomationID *int64 `json:"automation_id"`
	Status       string `json:"status"`
	TriggerType  string `json:"trigger_type"`
	CurrentStep  string `json:"current_step"`
	SortBy       string `json:"sort_by"`
	SortDir      string `json:"sort_dir"`
}

// OperationalInsightFilter represents query filters for operational insights
type OperationalInsightFilter struct {
	Page           int    `json:"page"`
	Limit          int    `json:"limit"`
	SourceModule   string `json:"source_module"`
	Status         string `json:"status"`
	Severity       string `json:"severity"`
	Priority       string `json:"priority"`
	AutomationID   *int64 `json:"automation_id"`
	ExecutionID    *int64 `json:"execution_id"`
	SourceRecordID *int64 `json:"source_record_id"`
	Search         string `json:"search"`
	SortBy         string `json:"sort_by"`
	SortDir        string `json:"sort_dir"`
}

// EvaluateEventInput represents real business event ingestion for automation trigger detection
type EvaluateEventInput struct {
	EventType      string          `json:"event_type"` // e.g. SHIPMENT_EXCEPTION_DETECTED, INVOICE_OVERDUE, etc.
	SourceModule   string          `json:"source_module"`
	SourceRecordID int64           `json:"source_record_id"`
	SourceRecordRef string         `json:"source_record_ref,omitempty"`
	Payload        json.RawMessage `json:"payload,omitempty"`
	IdempotencyKey string          `json:"idempotency_key,omitempty"`
}

// EvaluateEventResult represents the result of processing a real business event
type EvaluateEventResult struct {
	EventProcessed       bool                  `json:"event_processed"`
	SuppressedDuplicate  bool                  `json:"suppressed_duplicate"`
	MatchedAutomations   int                   `json:"matched_automations"`
	ExecutionsTriggered  []*AutomationExecution `json:"executions_triggered"`
	InsightsCreated      []*OperationalInsight `json:"insights_created"`
	CorrelationID        string                `json:"correlation_id"`
	ExecutionStatus      string                `json:"execution_status"`
	Summary              string                `json:"summary"`
}

// NextRunPreview represents next run preview info
type NextRunPreview struct {
	NextRunTime        time.Time `json:"next_run_time"`
	LocalTimeFormatted string    `json:"local_time_formatted"`
	Timezone           string    `json:"timezone"`
	DaysUntilRun       int       `json:"days_until_run"`
	HoursUntilRun      float64   `json:"hours_until_run"`
	HumanDescription   string    `json:"human_description"`
}

// AutomationStats provides KPI metrics across all automations in the organization
type AutomationStats struct {
	TotalAutomations         int     `json:"total_automations"`
	ActiveAutomations        int     `json:"active_automations"`
	TotalExecutions          int     `json:"total_executions"`
	SuccessfulExecutions     int     `json:"successful_executions"`
	FailedExecutions         int     `json:"failed_executions"`
	WaitingApprovalCount     int     `json:"waiting_approval_count"`
	ActiveInsightsCount      int     `json:"active_insights_count"`
	RecommendationsGenerated int     `json:"recommendations_generated"`
	RecommendationsUpdated   int     `json:"recommendations_updated"`
	RecentSuccessRate        float64 `json:"recent_success_rate"`
}

// SupportedAutomationTypeInfo provides metadata about approved automation definitions
type SupportedAutomationTypeInfo struct {
	Type                string   `json:"type"`
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	Category            string   `json:"category"`
	DefaultTriggerType  string   `json:"default_trigger_type"`
	DefaultScheduleType string   `json:"default_schedule_type"`
	DefaultScheduleTime string   `json:"default_schedule_time"`
	TargetModules       []string `json:"target_modules"`
	ReadOnlyGuarantee   string   `json:"read_only_guarantee"`
	RecommendedAction   string   `json:"recommended_action"`
}

// GetSupportedAutomationTypes returns the list of approved read-only assistant jobs
func GetSupportedAutomationTypes() []SupportedAutomationTypeInfo {
	return []SupportedAutomationTypeInfo{
		{
			Type:                AutomationTypeDailyOverdueInvoiceReview,
			Name:                "Daily Overdue Invoice & Receivables Review",
			Description:         "Evaluates overdue customer invoices, payment risk trends, and aging brackets (1-15d, 16-30d, 31-60d, 60+d). Generates prioritized collections follow-up recommendations without sending automatic messages.",
			Category:            "Finance & Collections",
			DefaultTriggerType:  TriggerTypeScheduled,
			DefaultScheduleType: ScheduleTypeDaily,
			DefaultScheduleTime: "08:00",
			TargetModules:       []string{"invoices", "customers", "finance"},
			ReadOnlyGuarantee:   "Strictly read-only. Zero invoices altered, zero automated payment reminder emails dispatched without operator approval.",
			RecommendedAction:   "Review collections recommendations and prepare follow-up message drafts.",
		},
		{
			Type:                AutomationTypeDailyContractDocumentExpiryReview,
			Name:                "Daily Contract, Document & Compliance Expiry Review",
			Description:         "Scans rate contracts, carrier agreements, permits, and regulatory compliance requirements for expiration within 14, 30, and 60 days, as well as missing documents and unassigned owners.",
			Category:            "Contracts & Compliance",
			DefaultTriggerType:  TriggerTypeScheduled,
			DefaultScheduleType: ScheduleTypeDaily,
			DefaultScheduleTime: "08:30",
			TargetModules:       []string{"contracts", "documents", "compliance"},
			ReadOnlyGuarantee:   "Strictly read-only. Zero contracts mutated, zero terms modified automatically.",
			RecommendedAction:   "Review expiring agreements and generate renewal reminders or internal review drafts.",
		},
		{
			Type:                AutomationTypeShipmentExceptionReview,
			Name:                "Shipment Exception & Delayed Milestone Review",
			Description:         "Reviews all active operational shipments, milestone variances, stalled customs clearances, unresolved exception events, and telemetry anomalies.",
			Category:            "Shipments & Operations",
			DefaultTriggerType:  TriggerTypeScheduled,
			DefaultScheduleType: ScheduleTypeDaily,
			DefaultScheduleTime: "07:30",
			TargetModules:       []string{"shipments", "tracking", "exceptions"},
			ReadOnlyGuarantee:   "Strictly read-only. Zero shipment statuses changed, zero exceptions closed automatically.",
			RecommendedAction:   "Review carrier coordination recommendations and update affected customer shipments.",
		},
		{
			Type:                AutomationTypeRFQQuotationReview,
			Name:                "RFQ & Quotation Commercial Health Review",
			Description:         "Identifies pending RFQs requiring carrier pricing, quotations expiring within 48 hours, missing commercial information, and below-target margin risks.",
			Category:            "Sales & Quotations",
			DefaultTriggerType:  TriggerTypeScheduled,
			DefaultScheduleType: ScheduleTypeDaily,
			DefaultScheduleTime: "09:00",
			TargetModules:       []string{"rfq", "quotations", "rates"},
			ReadOnlyGuarantee:   "Strictly read-only. Zero quotation prices altered, zero quotes auto-accepted.",
			RecommendedAction:   "Prioritize pending RFQ quotes and review carrier benchmark margins.",
		},
		{
			Type:                AutomationTypeCustomerFollowupReview,
			Name:                "Customer Inactivity & Credit Health Review",
			Description:         "Analyzes customer commercial interaction cadence, detects accounts with sudden booking inactivity (30+ days), and tracks high-risk credit exposure.",
			Category:            "Customer Relationships",
			DefaultTriggerType:  TriggerTypeScheduled,
			DefaultScheduleType: ScheduleTypeWeekly,
			DefaultScheduleTime: "09:30",
			TargetModules:       []string{"customers", "sales", "finance"},
			ReadOnlyGuarantee:   "Strictly read-only. Zero customer records modified, zero outreach messages auto-dispatched.",
			RecommendedAction:   "Prepare commercial check-in drafts for assigned account executives.",
		},
		{
			Type:                AutomationTypeDailyOperationalSummary,
			Name:                "Daily Enterprise Operational & Risk Synthesis",
			Description:         "Consolidates top risk signals across shipments, invoices, contracts, customer accounts, and compliance requirements into an executive digest.",
			Category:            "Executive Intelligence",
			DefaultTriggerType:  TriggerTypeScheduled,
			DefaultScheduleType: ScheduleTypeDaily,
			DefaultScheduleTime: "07:00",
			TargetModules:       []string{"shipments", "invoices", "contracts", "rfq", "customers"},
			ReadOnlyGuarantee:   "Strictly read-only. Consolidates deterministic findings without external actions.",
			RecommendedAction:   "Review daily synthesis in Recommendation Center and Mission Control.",
		},
	}
}

// ValidateSupportedType validates if an automation type is supported
func ValidateSupportedType(autoType string) bool {
	for _, item := range GetSupportedAutomationTypes() {
		if item.Type == autoType {
			return true
		}
	}
	return false
}

// ValidateTriggerType validates trigger type and applies default
func ValidateTriggerType(triggerType string) string {
	tt := strings.ToUpper(strings.TrimSpace(triggerType))
	switch tt {
	case TriggerTypeScheduled, TriggerTypeManual, TriggerTypeRecordCreated,
		TriggerTypeRecordUpdated, TriggerTypeStatusChanged, TriggerTypeMilestoneMissed,
		TriggerTypeShipmentExceptionDetected, TriggerTypeInvoiceOverdue,
		TriggerTypeContractExpiring, TriggerTypeRFQDeadlineApproaching, TriggerTypeApprovalReturned:
		return tt
	default:
		return TriggerTypeScheduled
	}
}

// ValidateApprovalPolicy validates approval policy
func ValidateApprovalPolicy(policy string) string {
	p := strings.ToUpper(strings.TrimSpace(policy))
	switch p {
	case ApprovalPolicyAlwaysRequire, ApprovalPolicyAutoIfLowRisk, ApprovalPolicyManualOnly:
		return p
	default:
		return ApprovalPolicyAlwaysRequire
	}
}

// CalculateNextRun calculates the next execution timestamp based on schedule parameters
func CalculateNextRun(scheduleType string, scheduleTime string, scheduleDays *string, timezone string, fromTime time.Time) (time.Time, error) {
	if timezone == "" {
		timezone = "UTC"
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	if scheduleTime == "" {
		scheduleTime = "08:00"
	}
	var hour, minute int
	_, err = fmt.Sscanf(scheduleTime, "%d:%d", &hour, &minute)
	if err != nil || hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		hour = 8
		minute = 0
	}

	nowInLoc := fromTime.In(loc)

	switch strings.ToUpper(scheduleType) {
	case ScheduleTypeHourly:
		next := nowInLoc.Add(1 * time.Hour)
		return time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), minute, 0, 0, loc).UTC(), nil

	case ScheduleTypeWeekly:
		targetDays := map[time.Weekday]bool{}
		if scheduleDays != nil && *scheduleDays != "" {
			parts := strings.Split(*scheduleDays, ",")
			for _, p := range parts {
				switch strings.ToUpper(strings.TrimSpace(p)) {
				case "SUN":
					targetDays[time.Sunday] = true
				case "MON":
					targetDays[time.Monday] = true
				case "TUE":
					targetDays[time.Tuesday] = true
				case "WED":
					targetDays[time.Wednesday] = true
				case "THU":
					targetDays[time.Thursday] = true
				case "FRI":
					targetDays[time.Friday] = true
				case "SAT":
					targetDays[time.Saturday] = true
				}
			}
		}
		if len(targetDays) == 0 {
			targetDays[time.Monday] = true
			targetDays[time.Tuesday] = true
			targetDays[time.Wednesday] = true
			targetDays[time.Thursday] = true
			targetDays[time.Friday] = true
		}

		for i := 0; i <= 14; i++ {
			candidateDate := nowInLoc.AddDate(0, 0, i)
			candidateRun := time.Date(candidateDate.Year(), candidateDate.Month(), candidateDate.Day(), hour, minute, 0, 0, loc)
			if targetDays[candidateRun.Weekday()] && candidateRun.After(nowInLoc) {
				return candidateRun.UTC(), nil
			}
		}
		return nowInLoc.AddDate(0, 0, 1).UTC(), nil

	case ScheduleTypeDaily:
		fallthrough
	default:
		todayRun := time.Date(nowInLoc.Year(), nowInLoc.Month(), nowInLoc.Day(), hour, minute, 0, 0, loc)
		if todayRun.After(nowInLoc) {
			return todayRun.UTC(), nil
		}
		tomorrowRun := todayRun.AddDate(0, 0, 1)
		return tomorrowRun.UTC(), nil
	}
}

// FormatNextRunPreview returns structured preview details
func FormatNextRunPreview(nextRun time.Time, timezone string) *NextRunPreview {
	if timezone == "" {
		timezone = "UTC"
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	localTime := nextRun.In(loc)
	now := time.Now()
	diff := nextRun.Sub(now)

	daysUntil := int(diff.Hours() / 24)
	if daysUntil < 0 {
		daysUntil = 0
	}

	humanDesc := fmt.Sprintf("Runs at %s %s (%s)", localTime.Format("15:04"), timezone, localTime.Format("Jan 02, 2006"))

	return &NextRunPreview{
		NextRunTime:        nextRun,
		LocalTimeFormatted: localTime.Format("Jan 02, 2006 15:04 MST"),
		Timezone:           timezone,
		DaysUntilRun:       daysUntil,
		HoursUntilRun:      diff.Hours(),
		HumanDescription:   humanDesc,
	}
}

// ValidateSchedule validates schedule parameters and applies defaults
func ValidateSchedule(scheduleType string, scheduleTime string, scheduleDays *string, timezone string) (string, string, *string, string, error) {
	if scheduleType == "" {
		scheduleType = ScheduleTypeDaily
	}
	st := strings.ToUpper(strings.TrimSpace(scheduleType))
	if st != ScheduleTypeDaily && st != ScheduleTypeWeekly && st != ScheduleTypeHourly {
		return "", "", nil, "", fmt.Errorf("%w: unrecognized schedule_type '%s'", ErrInvalidSchedule, scheduleType)
	}

	if scheduleTime == "" {
		scheduleTime = "08:00"
	}
	var hour, minute int
	_, err := fmt.Sscanf(scheduleTime, "%d:%d", &hour, &minute)
	if err != nil || hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return "", "", nil, "", fmt.Errorf("%w: schedule_time must be HH:MM in 24h format", ErrInvalidSchedule)
	}

	if timezone == "" {
		timezone = "UTC"
	}
	_, err = time.LoadLocation(timezone)
	if err != nil {
		timezone = "UTC"
	}

	return st, scheduleTime, scheduleDays, timezone, nil
}

// Errors
var (
	ErrAutomationNotFound       = errors.New("automation not found")
	ErrExecutionNotFound        = errors.New("automation execution not found")
	ErrInsightNotFound          = errors.New("operational insight not found")
	ErrInvalidAutomationType    = errors.New("unsupported or invalid automation type")
	ErrUnsupportedType          = ErrInvalidAutomationType
	ErrInvalidSchedule          = errors.New("invalid schedule configuration")
	ErrExecutionAlreadyRunning  = errors.New("an execution is already currently queued or running for this automation")
	ErrAutomationDisabled       = errors.New("cannot run a disabled automation")
	ErrUnauthorizedAutomationOp = errors.New("unauthorized automation operation")
	ErrDuplicateExecution       = errors.New("duplicate execution detected for event")
)
