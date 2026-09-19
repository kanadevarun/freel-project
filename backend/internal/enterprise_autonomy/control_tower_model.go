package enterprise_autonomy

import (
	"time"
)

// ---------------------------------------------------------------------
// Phase 7.9: Enterprise Autonomous Control Tower Models & DTOs
// ---------------------------------------------------------------------

// ControlTowerAttentionSeverity defines priority rankings
type ControlTowerAttentionSeverity string

const (
	AttentionSeverityCritical ControlTowerAttentionSeverity = "CRITICAL"
	AttentionSeverityHigh     ControlTowerAttentionSeverity = "HIGH"
	AttentionSeverityMedium   ControlTowerAttentionSeverity = "MEDIUM"
	AttentionSeverityLow      ControlTowerAttentionSeverity = "LOW"
)

// ControlTowerAttentionItem represents a prioritized operational alert requiring human attention
type ControlTowerAttentionItem struct {
	ID                 string                        `json:"id"`
	Severity           ControlTowerAttentionSeverity `json:"severity"`
	Category           string                        `json:"category"` // SHIPMENT, EXCEPTION, COMMERCIAL, FINANCE, RISK, COMPLIANCE, WORKFLOW
	Title              string                        `json:"title"`
	AffectedEntity     string                        `json:"affected_entity"`
	EntityType         string                        `json:"entity_type"`
	EntityID           string                        `json:"entity_id"`
	Reason             string                        `json:"reason"`
	CurrentState       string                        `json:"current_state"`
	RequiredHumanAction string                       `json:"required_human_action"`
	Urgency            string                        `json:"urgency"` // IMMEDIATE, WITHIN_1_HOUR, TODAY, SCHEDULED
	WorkflowID         *string                       `json:"workflow_id,omitempty"`
	ApprovalID         *int64                        `json:"approval_id,omitempty"`
	Confidence         float64                       `json:"confidence"`
	IsAuthoritative    bool                          `json:"is_authoritative"`
	OccurredAt         time.Time                     `json:"occurred_at"`
	Timestamp          time.Time                     `json:"timestamp"`
}

// ControlTowerDomainSummary aggregates business counts and health across one operational pillar
type ControlTowerDomainSummary struct {
	DomainName         string  `json:"domain_name"`
	AuthoritativeCount int     `json:"authoritative_count"`
	AtRiskCount        int     `json:"at_risk_count"`
	ActiveWorkflows    int     `json:"active_workflows"`
	PendingApprovals   int     `json:"pending_approvals"`
	FinancialExposure  float64 `json:"financial_exposure"`
	StatusIndicator    string  `json:"status_indicator"` // OPTIMAL, WATCH, AT_RISK, BLOCKED
	KeyInsight         string  `json:"key_insight"`
}

// ControlTowerWorkforceHealth represents the status of the AI agent workforce
type ControlTowerWorkforceHealth struct {
	TotalAgents        int      `json:"total_agents"`
	HealthyAgents      int      `json:"healthy_agents"`
	DegradedAgents     int      `json:"degraded_agents"`
	ActiveWorkflows    int      `json:"active_workflows"`
	AutonomousLevelMax string   `json:"autonomous_level_max"`
	QueuePressure      string   `json:"queue_pressure"` // NOMINAL, ELEVATED, CRITICAL
	EmergencyStopActive bool    `json:"emergency_stop_active"`
	LastAuditedAt      time.Time `json:"last_audited_at"`
}

// ControlTowerComprehensiveView returns the full enterprise picture in one cohesive, lightweight payload
type ControlTowerComprehensiveView struct {
	OrgID               int64                                `json:"org_id"`
	LastSyncedAt        time.Time                            `json:"last_synced_at"`
	PlatformHealth      string                               `json:"platform_health"` // HEALTHY, DEGRADED, EMERGENCY_HALT
	HumanAttentionItems []ControlTowerAttentionItem          `json:"human_attention_items"`
	ActiveWorkflows     []*EnterpriseWorkflow                `json:"active_workflows"`
	DomainSummaries     map[string]ControlTowerDomainSummary `json:"domain_summaries"`
	WorkforceHealth     ControlTowerWorkforceHealth          `json:"workforce_health"`
	AutonomyOverview    map[string]int                       `json:"autonomy_overview"`
	RecentEscalations   []ControlTowerAttentionItem          `json:"recent_escalations"`
	GovernanceStatus    *GovernanceStatusSummary             `json:"governance_status,omitempty"`
	ResilienceHealth    *EnterprisePlatformHealthSummary     `json:"resilience_health,omitempty"`
}

// WorkflowTraceDetail exposes the complete auditable event-to-action lineage for operator drill-down
type WorkflowTraceDetail struct {
	WorkflowID         string                   `json:"workflow_id"`
	OrgID              int64                    `json:"org_id"`
	WorkflowType       EnterpriseWorkflowType   `json:"workflow_type"`
	CurrentState       EnterpriseWorkflowState  `json:"current_state"`
	InitiatingEvent    string                   `json:"initiating_event"`
	CorrelationID      string                   `json:"correlation_id"`
	AutonomyLevel      string                   `json:"autonomy_level"`
	RelatedEntityType  string                   `json:"related_entity_type"`
	RelatedEntityID    string                   `json:"related_entity_id"`
	Confidence         float64                  `json:"confidence"`
	Steps              []EnterpriseWorkflowStep `json:"steps"`
	PendingApprovals   []EnterprisePendingApproval `json:"pending_approvals,omitempty"`
	AuthoritativeFacts map[string]interface{}   `json:"authoritative_facts"`
	AIPredictions      map[string]interface{}   `json:"ai_predictions"`
	PolicyDecision     string                   `json:"policy_decision"`
	VerificationStatus string                   `json:"verification_status"`
	CreatedAt          time.Time                `json:"created_at"`
	UpdatedAt          time.Time                `json:"updated_at"`
}
