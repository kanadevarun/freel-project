package autonomy

import (
	"encoding/json"
	"time"
)

// ActionRiskClass categorizes the potential harm/exposure of an action
type ActionRiskClass string

const (
	RiskClassLow      ActionRiskClass = "LOW"
	RiskClassMedium   ActionRiskClass = "MEDIUM"
	RiskClassHigh     ActionRiskClass = "HIGH"
	RiskClassCritical ActionRiskClass = "CRITICAL"
)

// ActionReversibility defines whether an action can be undone
type ActionReversibility string

const (
	ReversibilityReversible          ActionReversibility = "REVERSIBLE"
	ReversibilityPartiallyReversible ActionReversibility = "PARTIALLY_REVERSIBLE"
	ReversibilityIrreversible        ActionReversibility = "IRREVERSIBLE"
)

// ApprovalRequirement defines human approval policy for an action
type ApprovalRequirement string

const (
	ApprovalAlways                  ApprovalRequirement = "ALWAYS"
	ApprovalConditionalRiskThreshold ApprovalRequirement = "CONDITIONAL_RISK_THRESHOLD"
	ApprovalNever                   ApprovalRequirement = "NEVER"
)

// DataSufficiencyState defines completeness of operational context
type DataSufficiencyState string

const (
	SufficiencySufficient          DataSufficiencyState = "SUFFICIENT"
	SufficiencyPartiallySufficient DataSufficiencyState = "PARTIALLY_SUFFICIENT"
	SufficiencyInsufficient        DataSufficiencyState = "INSUFFICIENT"
)

// ConfidenceClass categorizes confidence
type ConfidenceClass string

const (
	ConfidenceClassHigh   ConfidenceClass = "HIGH"
	ConfidenceClassMedium ConfidenceClass = "MEDIUM"
	ConfidenceClassLow    ConfidenceClass = "LOW"
)

// GovernanceDecision represents the final authoritative Go verdict
type GovernanceDecision string

const (
	GovernanceDecisionAllow          GovernanceDecision = "ALLOW"
	GovernanceDecisionBlock          GovernanceDecision = "BLOCK"
	GovernanceDecisionRequireReview  GovernanceDecision = "REQUIRE_REVIEW"
	GovernanceDecisionUnavailable    GovernanceDecision = "GOVERNANCE_UNAVAILABLE"
)

// ActionAllowlistItem represents a validated action entry in the tenant's allowlist
type ActionAllowlistItem struct {
	ID                   int64               `db:"id" json:"id"`
	OrgID                int64               `db:"org_id" json:"org_id"`
	ActionType           string              `db:"action_type" json:"action_type"`
	ActionName           string              `db:"action_name" json:"action_name"`
	Module               string              `db:"module" json:"module"`
	RiskClass            ActionRiskClass     `db:"risk_class" json:"risk_class"`
	AllowedAutonomyLevels json.RawMessage    `db:"allowed_autonomy_levels" json:"allowed_autonomy_levels"`
	ApprovalRequirement  ApprovalRequirement `db:"approval_requirement" json:"approval_requirement"`
	RequiredPermission   string              `db:"required_permission" json:"required_permission"`
	Reversibility        ActionReversibility `db:"reversibility" json:"reversibility"`
	MaxFinancialLimit    float64             `db:"max_financial_limit" json:"max_financial_limit"`
	IsEnabled            bool                `db:"is_enabled" json:"is_enabled"`
	Description          string              `db:"description" json:"description"`
	CreatedAt            time.Time           `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time           `db:"updated_at" json:"updated_at"`
}

// GovernanceFeatureFlag represents a granular autonomous capability toggle
type GovernanceFeatureFlag struct {
	ID               int64     `db:"id" json:"id"`
	OrgID            int64     `db:"org_id" json:"org_id"`
	FlagKey          string    `db:"flag_key" json:"flag_key"`
	FlagName         string    `db:"flag_name" json:"flag_name"`
	IsEnabled        bool      `db:"is_enabled" json:"is_enabled"`
	MaxAutonomyLevel int       `db:"max_autonomy_level" json:"max_autonomy_level"`
	RequiresApproval bool      `db:"requires_approval" json:"requires_approval"`
	Description      string    `db:"description" json:"description"`
	UpdatedByID      *int64    `db:"updated_by_id" json:"updated_by_id"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// TenantGovernanceLimits represents tenant-wide blast radius, rates, and kill switch controls
type TenantGovernanceLimits struct {
	ID                              int64      `db:"id" json:"id"`
	OrgID                           int64      `db:"org_id" json:"org_id"`
	MaxTenantAutonomy               int        `db:"max_tenant_autonomy" json:"max_tenant_autonomy"`
	KillSwitchActive                bool       `db:"kill_switch_active" json:"kill_switch_active"`
	KillSwitchReason                *string    `db:"kill_switch_reason" json:"kill_switch_reason"`
	KillSwitchByID                  *int64     `db:"kill_switch_by_id" json:"kill_switch_by_id"`
	KillSwitchAt                    *time.Time `db:"kill_switch_at" json:"kill_switch_at"`
	MaxActionsPerHour               int        `db:"max_actions_per_hour" json:"max_actions_per_hour"`
	MaxFinancialExposurePerWorkflow float64    `db:"max_financial_exposure_per_workflow" json:"max_financial_exposure_per_workflow"`
	MaxRetriesPerStep               int        `db:"max_retries_per_step" json:"max_retries_per_step"`
	MaxReplansPerPlan               int        `db:"max_replans_per_plan" json:"max_replans_per_plan"`
	EnforceFourEyes                 bool       `db:"enforce_four_eyes" json:"enforce_four_eyes"`
	CreatedAt                       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt                       time.Time  `db:"updated_at" json:"updated_at"`
}

// PolicyEvaluationRecord logs each authoritative governance check
type PolicyEvaluationRecord struct {
	ID                int64                `db:"id" json:"id"`
	OrgID             int64                `db:"org_id" json:"org_id"`
	UserID            *int64               `db:"user_id" json:"user_id"`
	EntityType        string               `db:"entity_type" json:"entity_type"`
	EntityID          string               `db:"entity_id" json:"entity_id"`
	ActionType        string               `db:"action_type" json:"action_type"`
	Module            string               `db:"module" json:"module"`
	RequestedAutonomy int                  `db:"requested_autonomy" json:"requested_autonomy"`
	EffectiveAutonomy int                  `db:"effective_autonomy" json:"effective_autonomy"`
	Decision          GovernanceDecision   `db:"decision" json:"decision"`
	Reasons           json.RawMessage      `db:"reasons" json:"reasons"`
	RiskLevel         ActionRiskClass      `db:"risk_level" json:"risk_level"`
	DataSufficiency   DataSufficiencyState `db:"data_sufficiency" json:"data_sufficiency"`
	Confidence        ConfidenceClass      `db:"confidence" json:"confidence"`
	ApprovalRequired  bool                 `db:"approval_required" json:"approval_required"`
	FourEyesRequired  bool                 `db:"four_eyes_required" json:"four_eyes_required"`
	PolicyVersion     int                  `db:"policy_version" json:"policy_version"`
	CorrelationID     string               `db:"correlation_id" json:"correlation_id"`
	CreatedAt         time.Time            `db:"created_at" json:"created_at"`
}

// PolicyAuditLog tracks administrative changes to governance configurations
type PolicyAuditLog struct {
	ID         int64           `db:"id" json:"id"`
	OrgID      int64           `db:"org_id" json:"org_id"`
	UserID     int64           `db:"user_id" json:"user_id"`
	ChangeType string          `db:"change_type" json:"change_type"`
	TargetType string          `db:"target_type" json:"target_type"`
	TargetID   string          `db:"target_id" json:"target_id"`
	OldValue   json.RawMessage `db:"old_value" json:"old_value"`
	NewValue   json.RawMessage `db:"new_value" json:"new_value"`
	Reason     *string         `db:"reason" json:"reason"`
	CreatedAt  time.Time       `db:"created_at" json:"created_at"`
}

// GovernanceEvaluationRequest is input to the central Go policy engine
type GovernanceEvaluationRequest struct {
	OrgID              int64                  `json:"org_id"`
	UserID             int64                  `json:"user_id"`
	UserRole           string                 `json:"user_role"`
	UserPermissions    []string               `json:"user_permissions"`
	PreparerUserID     int64                  `json:"preparer_user_id"`
	Module             string                 `json:"module"`
	ActionType         string                 `json:"action_type"`
	EntityType         string                 `json:"entity_type"`
	EntityID           string                 `json:"entity_id"`
	EntityState        string                 `json:"entity_state"`
	Parameters         map[string]interface{} `json:"parameters"`
	ContextText        string                 `json:"context_text"`
	RequestedAutonomy  int                    `json:"requested_autonomy"`
	CorrelationID      string                 `json:"correlation_id"`
	IsApproved         bool                   `json:"is_approved"`
	ApprovalApproverID int64                  `json:"approval_approver_id"`
	ComplianceStatus   string                 `json:"compliance_status"`
}

// GovernanceEvaluationResponse is output from the central Go policy engine
type GovernanceEvaluationResponse struct {
	Decision          GovernanceDecision   `json:"decision"`
	EffectiveAutonomy int                  `json:"effective_autonomy"`
	RiskClass         ActionRiskClass      `json:"risk_class"`
	DataSufficiency   DataSufficiencyState `json:"data_sufficiency"`
	ConfidenceClass   ConfidenceClass      `json:"confidence_class"`
	RequiresApproval  bool                 `json:"requires_approval"`
	FourEyesEnforced  bool                 `json:"four_eyes_enforced"`
	Reversibility     ActionReversibility  `json:"reversibility"`
	Reasons           []string             `json:"reasons"`
	PolicyVersion     int                  `json:"policy_version"`
	CorrelationID     string               `json:"correlation_id"`
}

// SidecarGovernanceEvalRequest is DTO for python sidecar context evaluation
type SidecarGovernanceEvalRequest struct {
	OrgID               int64                  `json:"org_id"`
	Module              string                 `json:"module"`
	ActionType          string                 `json:"action_type"`
	EntityType          string                 `json:"entity_type"`
	EntityID            string                 `json:"entity_id"`
	Parameters          map[string]interface{} `json:"parameters"`
	ContextText         string                 `json:"context_text"`
	RequestedAutonomy   int                    `json:"requested_autonomy"`
	ConfiguredAutonomy  int                    `json:"configured_autonomy"`
	HistoricalMemories  []map[string]interface{}`json:"historical_memories"`
	CorrelationID       string                 `json:"correlation_id"`
}

// SidecarGovernanceEvalResponse is DTO returned by python sidecar
type SidecarGovernanceEvalResponse struct {
	RiskScore               float64  `json:"risk_score"`
	RiskClass               string   `json:"risk_class"`
	DataSufficiency         string   `json:"data_sufficiency"`
	MissingDataFields       []string `json:"missing_data_fields"`
	ConfidenceClass         string   `json:"confidence_class"`
	ConfidenceScore         float64  `json:"confidence_score"`
	RecommendedAutonomyTier int      `json:"recommended_autonomy_tier"`
	BlastRadiusAssessment   string   `json:"blast_radius_assessment"`
	Explanation             string   `json:"explanation"`
	SanitizedContext        string   `json:"sanitized_context"`
	CorrelationID           string   `json:"correlation_id"`
}

// SidecarPlanPreviewRequest is DTO for python sidecar plan dry-run
type SidecarPlanPreviewRequest struct {
	OrgID         int64                    `json:"org_id"`
	PlanID        string                   `json:"plan_id"`
	Goal          string                   `json:"goal"`
	Module        string                   `json:"module"`
	Steps         []map[string]interface{} `json:"steps"`
	CorrelationID string                   `json:"correlation_id"`
}

// SidecarPlanPreviewResponse is DTO returned by python sidecar
type SidecarPlanPreviewResponse struct {
	PlanID                       string              `json:"plan_id"`
	TotalSteps                   int                 `json:"total_steps"`
	ExecutableSteps              int                 `json:"executable_steps"`
	ApprovalRequiredSteps        int                 `json:"approval_required_steps"`
	MaxRiskClass                 string              `json:"max_risk_class"`
	EstimatedFinancialExposureUSD float64             `json:"estimated_financial_exposure_usd"`
	AffectedEntities             []map[string]string `json:"affected_entities"`
	SafetySummary                string              `json:"safety_summary"`
	CorrelationID                string              `json:"correlation_id"`
}

// GovernanceTelemetrySummary represents high-level governance telemetry
type GovernanceTelemetrySummary struct {
	TotalEvaluations    int                      `json:"total_evaluations"`
	AllowedCount        int                      `json:"allowed_count"`
	BlockedCount        int                      `json:"blocked_count"`
	ReviewRequiredCount int                      `json:"review_required_count"`
	KillSwitchActive    bool                     `json:"kill_switch_active"`
	MaxTenantAutonomy   int                      `json:"max_tenant_autonomy"`
	ActiveFlagsCount    int                      `json:"active_flags_count"`
	AllowlistCount      int                      `json:"allowlist_count"`
	RecentEvaluations   []PolicyEvaluationRecord `json:"recent_evaluations"`
}

// AutonomyLevelToInt converts an AutonomyLevel enum string to its 0-4 integer tier
func AutonomyLevelToInt(level AutonomyLevel) int {
	switch level {
	case Level0Observe:
		return 0
	case Level1Recommend:
		return 1
	case Level2Prepare:
		return 2
	case Level3ControlledExecution:
		return 3
	case Level4ControlledMultiStep:
		return 4
	default:
		return 2
	}
}
