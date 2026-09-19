package governance

import (
	"encoding/json"
	"time"
)

// GovernancePolicy represents tenant-specific or system-wide AI constraints.
type GovernancePolicy struct {
	ID                            int64     `db:"id" json:"id"`
	OrgID                         int64     `db:"org_id" json:"org_id"`
	AllowedWorkflowsJSON          string    `db:"allowed_workflows" json:"-"`
	AllowedModelsJSON             string    `db:"allowed_models" json:"-"`
	AllowedProvidersJSON          string    `db:"allowed_providers" json:"-"`
	AllowedActionsJSON            string    `db:"allowed_actions" json:"-"`
	AllowedWorkflows              []string  `json:"allowed_workflows"`
	AllowedModels                 []string  `json:"allowed_models"`
	AllowedProviders              []string  `json:"allowed_providers"`
	AllowedActions                []string  `json:"allowed_actions"`
	MaxInputChars                 int       `db:"max_input_chars" json:"max_input_chars"`
	MaxOutputChars                int       `db:"max_output_chars" json:"max_output_chars"`
	TimeoutSeconds                int       `db:"timeout_seconds" json:"timeout_seconds"`
	MaxRetries                    int       `db:"max_retries" json:"max_retries"`
	RateLimitRPM                  int       `db:"rate_limit_rpm" json:"rate_limit_rpm"`
	EnforcePromptInjectionCheck   bool      `db:"enforce_prompt_injection_check" json:"enforce_prompt_injection_check"`
	EnforceSensitiveDataRedaction bool      `db:"enforce_sensitive_data_redaction" json:"enforce_sensitive_data_redaction"`
	EnforceHITLApprovals          bool      `db:"enforce_hitl_approvals" json:"enforce_hitl_approvals"`
	IsActive                      bool      `db:"is_active" json:"is_active"`
	CreatedAt                     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                     time.Time `db:"updated_at" json:"updated_at"`
}

func (p *GovernancePolicy) UnpackJSON() {
	if p.AllowedWorkflowsJSON != "" {
		_ = json.Unmarshal([]byte(p.AllowedWorkflowsJSON), &p.AllowedWorkflows)
	}
	if p.AllowedModelsJSON != "" {
		_ = json.Unmarshal([]byte(p.AllowedModelsJSON), &p.AllowedModels)
	}
	if p.AllowedProvidersJSON != "" {
		_ = json.Unmarshal([]byte(p.AllowedProvidersJSON), &p.AllowedProviders)
	}
	if p.AllowedActionsJSON != "" {
		_ = json.Unmarshal([]byte(p.AllowedActionsJSON), &p.AllowedActions)
	}
}

// KillSwitchRecord represents an administrative kill switch blocking AI execution.
type KillSwitchRecord struct {
	ID               int64     `db:"id" json:"id"`
	OrgID            int64     `db:"org_id" json:"org_id"`
	Scope            string    `db:"scope" json:"scope"` // GLOBAL, WORKFLOW, MODEL, PROVIDER, ACTION
	TargetIdentifier string    `db:"target_identifier" json:"target_identifier"`
	IsKilled         bool      `db:"is_killed" json:"is_killed"`
	Reason           string    `db:"reason" json:"reason"`
	KilledByUserID   *int64    `db:"killed_by_user_id" json:"killed_by_user_id,omitempty"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// SafetyViolationRecord stores security and policy rejection events.
type SafetyViolationRecord struct {
	ID                   int64     `db:"id" json:"id"`
	OrgID                int64     `db:"org_id" json:"org_id"`
	UserID               *int64    `db:"user_id" json:"user_id,omitempty"`
	WorkflowName         string    `db:"workflow_name" json:"workflow_name"`
	ViolationType        string    `db:"violation_type" json:"violation_type"`
	Severity             string    `db:"severity" json:"severity"`
	ActionTaken          string    `db:"action_taken" json:"action_taken"`
	Details              string    `db:"details" json:"details"`
	InputSnippetRedacted *string   `db:"input_snippet_redacted" json:"input_snippet_redacted,omitempty"`
	CorrelationID        string    `db:"correlation_id" json:"correlation_id"`
	CreatedAt            time.Time `db:"created_at" json:"created_at"`
}

// DTOs for Go <-> Python Sidecar HTTP communication

type SidecarInputInspectionReq struct {
	OrgID         int64                  `json:"org_id"`
	UserID        *int64                 `json:"user_id,omitempty"`
	WorkflowName  string                 `json:"workflow_name"`
	InputText     string                 `json:"input_text"`
	InputPayload  map[string]interface{} `json:"input_payload,omitempty"`
	CorrelationID string                 `json:"correlation_id"`
}

type SidecarInputInspectionResp struct {
	IsSafe                   bool          `json:"is_safe"`
	PromptInjectionDetected  bool          `json:"prompt_injection_detected"`
	InjectionConfidence      float64       `json:"injection_confidence"`
	PIIDetected              bool          `json:"pii_detected"`
	SanitizedText            string        `json:"sanitized_text"`
	DetectedPIITypes         []string      `json:"detected_pii_types"`
	SafetyChecks             []interface{} `json:"safety_checks"`
	RefusalReason            *string       `json:"refusal_reason,omitempty"`
	CorrelationID            string        `json:"correlation_id"`
}

type SidecarOutputInspectionReq struct {
	OrgID           int64                    `json:"org_id"`
	UserID          *int64                   `json:"user_id,omitempty"`
	WorkflowName    string                   `json:"workflow_name"`
	OutputText      string                   `json:"output_text"`
	OutputPayload   map[string]interface{}   `json:"output_payload,omitempty"`
	SourceFacts     []map[string]interface{} `json:"source_facts"`
	ProposedActions []map[string]interface{} `json:"proposed_actions"`
	CorrelationID   string                   `json:"correlation_id"`
}

type ActionSafetyEvaluationDTO struct {
	ActionType            string `json:"action_type"`
	ActionTitle           string `json:"action_title"`
	IsConsequential       bool   `json:"is_consequential"`
	RequiresHumanApproval bool   `json:"requires_human_approval"`
	RiskLevel             string `json:"risk_level"`
	SafetyReason          string `json:"safety_reason"`
}

type SidecarOutputInspectionResp struct {
	IsSafe                bool                        `json:"is_safe"`
	GroundingScore        float64                     `json:"grounding_score"`
	HallucinationRisk     string                      `json:"hallucination_risk"`
	UnsupportedClaims     []string                    `json:"unsupported_claims"`
	ActionEvaluations     []ActionSafetyEvaluationDTO `json:"action_evaluations"`
	RequiresHumanApproval bool                        `json:"requires_human_approval"`
	SafetyChecks          []interface{}               `json:"safety_checks"`
	RefusalReason         *string                     `json:"refusal_reason,omitempty"`
	CorrelationID         string                      `json:"correlation_id"`
}

type SidecarQualityEvalReq struct {
	OrgID         int64                    `json:"org_id"`
	TestRunID     string                   `json:"test_run_id"`
	WorkflowName  string                   `json:"workflow_name"`
	TestCases     []map[string]interface{} `json:"test_cases"`
	CorrelationID string                   `json:"correlation_id"`
}

type SidecarQualityEvalResp struct {
	TestRunID         string                   `json:"test_run_id"`
	WorkflowName      string                   `json:"workflow_name"`
	TotalCases        int                      `json:"total_cases"`
	PassedCases       int                      `json:"passed_cases"`
	PassRate          float64                  `json:"pass_rate"`
	AvgGroundingScore float64                  `json:"avg_grounding_score"`
	OverallStatus     string                   `json:"overall_status"`
	Results           []map[string]interface{} `json:"results"`
	Summary           string                   `json:"summary"`
	CorrelationID     string                   `json:"correlation_id"`
}

// User Request / Response DTOs

type PreExecutionCheckRequest struct {
	WorkflowName string                 `json:"workflow_name"`
	ModelName    string                 `json:"model_name"`
	ProviderName string                 `json:"provider_name"`
	InputText    string                 `json:"input_text"`
	InputPayload map[string]interface{} `json:"input_payload,omitempty"`
}

type PreExecutionCheckResult struct {
	Allowed            bool    `json:"allowed"`
	BlockReason        *string `json:"block_reason,omitempty"`
	SanitizedInput     string  `json:"sanitized_input"`
	Sanitized          bool    `json:"sanitized"`
	PromptInjection    bool    `json:"prompt_injection"`
	EnforceApproval    bool    `json:"enforce_approval"`
	TimeoutSeconds     int     `json:"timeout_seconds"`
	MaxRetries         int     `json:"max_retries"`
	CorrelationID      string  `json:"correlation_id"`
}

type PostExecutionCheckRequest struct {
	WorkflowName    string                   `json:"workflow_name"`
	OutputText      string                   `json:"output_text"`
	OutputPayload   map[string]interface{}   `json:"output_payload,omitempty"`
	SourceFacts     []map[string]interface{} `json:"source_facts,omitempty"`
	ProposedActions []map[string]interface{} `json:"proposed_actions,omitempty"`
}

type PostExecutionCheckResult struct {
	Allowed               bool                        `json:"allowed"`
	BlockReason           *string                     `json:"block_reason,omitempty"`
	GroundingScore        float64                     `json:"grounding_score"`
	HallucinationRisk     string                      `json:"hallucination_risk"`
	RequiresHumanApproval bool                        `json:"requires_human_approval"`
	ActionEvaluations     []ActionSafetyEvaluationDTO `json:"action_evaluations"`
	CorrelationID         string                      `json:"correlation_id"`
}

type SetKillSwitchInput struct {
	Scope            string `json:"scope"`
	TargetIdentifier string `json:"target_identifier"`
	IsKilled         bool   `json:"is_killed"`
	Reason           string `json:"reason"`
}

type GovernanceOverviewDTO struct {
	SystemStatus        string                 `json:"system_status"` // HEALTHY, DEGRADED, BLOCKED
	GlobalKillSwitch    bool                   `json:"global_kill_switch"`
	ActiveKillSwitches  int                    `json:"active_kill_switches"`
	AllowedWorkflows    []string               `json:"allowed_workflows"`
	AllowedProviders    []string               `json:"allowed_providers"`
	AllowedModels       []string               `json:"allowed_models"`
	Violations24h       int                    `json:"violations_24h"`
	RecentViolations    []SafetyViolationRecord`json:"recent_violations"`
	HITLApprovalsActive bool                   `json:"hitl_approvals_active"`
	ProductionReadiness string                 `json:"production_readiness"` // PRODUCTION_READY, REVIEW_REQUIRED
}
