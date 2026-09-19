package bcontext

import (
	"time"
)

// CrossModuleInsightSeverity represents the urgency and business impact of an insight.
type CrossModuleInsightSeverity string

const (
	SeverityCritical CrossModuleInsightSeverity = "CRITICAL"
	SeverityHigh     CrossModuleInsightSeverity = "HIGH"
	SeverityModerate CrossModuleInsightSeverity = "MODERATE"
	SeverityLow      CrossModuleInsightSeverity = "LOW"
	SeverityInfo     CrossModuleInsightSeverity = "INFO"
)

// CrossModuleInsightType classifies the domains connected by the insight.
type CrossModuleInsightType string

const (
	TypeOperationalFinance   CrossModuleInsightType = "OPERATIONAL_FINANCE"
	TypeCommercialContract   CrossModuleInsightType = "COMMERCIAL_CONTRACT"
	TypeCustomerRisk         CrossModuleInsightType = "CUSTOMER_RISK"
	TypeComplianceOperations CrossModuleInsightType = "COMPLIANCE_OPERATIONS"
	TypeWorkflowAttention    CrossModuleInsightType = "WORKFLOW_ATTENTION"
)

// EntityReference identifies a source business record in a specific domain module.
type EntityReference struct {
	EntityType string `json:"entity_type"` // CUSTOMER, SHIPMENT, INVOICE, RFQ, QUOTATION, CONTRACT, COMPLIANCE
	EntityID   int64  `json:"entity_id"`
	Reference  string `json:"reference"` // Human readable identifier (e.g. SH-101, INV-2026-001)
	Module     string `json:"module"`    // shipments, finance, rfq, contracts, customers
}

// InsightEvidenceItem provides factual evidence for an insight traced back to a specific record.
type InsightEvidenceItem struct {
	SourceModule   string      `json:"source_module"`
	SourceEntityID int64       `json:"source_entity_id"`
	SourceRef      string      `json:"source_ref"`
	FieldName      string      `json:"field_name"`
	ObservedValue  interface{} `json:"observed_value"`
	Description    string      `json:"description"`
}

// CrossModuleInsight represents a single, deterministic or AI-synthesized connected business insight.
type CrossModuleInsight struct {
	InsightID            string                     `json:"insight_id"`
	InsightType          CrossModuleInsightType     `json:"insight_type"`
	Category             string                     `json:"category"`
	Title                string                     `json:"title"`
	Explanation          string                     `json:"explanation"`
	Severity             CrossModuleInsightSeverity `json:"severity"`
	PriorityScore        int                        `json:"priority_score"` // 1 - 100
	Confidence           string                     `json:"confidence"`     // HIGH, MEDIUM, LOW
	DataFreshness        time.Time                  `json:"data_freshness"`
	OrganizationScope    int64                      `json:"organization_scope"`
	PrimaryEntity        EntityReference            `json:"primary_entity"`
	RelatedEntities      []EntityReference          `json:"related_entities"`
	Evidence             []InsightEvidenceItem      `json:"evidence"`
	RuleApplied          string                     `json:"rule_applied"`
	IsDeterministic      bool                       `json:"is_deterministic"`
	LifecycleStatus      string                     `json:"lifecycle_status"` // ACTIVE, RECENTLY_OBSERVED, RESOLVED, STALE
	SuggestedHumanAction string                     `json:"suggested_human_action"`
	Limitations          []string                   `json:"limitations"`
	ReadOnly             bool                       `json:"read_only"`
}

// CrossModuleAISynthesis represents grounded AI analysis contextualizing multiple related signals.
type CrossModuleAISynthesis struct {
	ExecutiveSummary                string    `json:"executive_summary"`
	ConnectedSituationAnalysis      string    `json:"connected_situation_analysis"`
	TradeoffsAndPriorities          string    `json:"tradeoffs_and_priorities"`
	SuggestedInvestigationQuestions []string  `json:"suggested_investigation_questions"`
	Confidence                      string    `json:"confidence"`
	GeneratedAt                     time.Time `json:"generated_at"`
	Limitations                     []string  `json:"limitations"`
}

// CrossModuleInsightsResult encapsulates cross-module intelligence for an entity or organization query.
type CrossModuleInsightsResult struct {
	PrimaryEntity      *EntityReference        `json:"primary_entity,omitempty"`
	OrganizationScope  int64                   `json:"organization_scope"`
	CorrelationID      string                  `json:"correlation_id"`
	CalculatedAt       time.Time               `json:"calculated_at"`
	ReadOnly           bool                    `json:"read_only"`
	TotalInsightsCount int                     `json:"total_insights_count"`
	CriticalCount      int                     `json:"critical_count"`
	HighCount          int                     `json:"high_count"`
	ModerateCount      int                     `json:"moderate_count"`
	LowCount           int                     `json:"low_count"`
	ConnectedModules   []string                `json:"connected_modules"`
	Insights           []CrossModuleInsight    `json:"insights"`
	AISynthesis        *CrossModuleAISynthesis `json:"ai_synthesis,omitempty"`
	Warnings           []string                `json:"warnings"`
	Limitations        []string                `json:"limitations"`
}

// OrgCrossModuleSummary aggregates organization-level cross-module health indicators and top priorities.
type OrgCrossModuleSummary struct {
	OrganizationScope      int64                  `json:"organization_scope"`
	CorrelationID          string                 `json:"correlation_id"`
	CalculatedAt           time.Time              `json:"calculated_at"`
	ReadOnly               bool                   `json:"read_only"`
	TotalInsights          int                    `json:"total_insights"`
	CriticalInsights       int                    `json:"critical_insights"`
	HighInsights           int                    `json:"high_insights"`
	ModerateInsights       int                    `json:"moderate_insights"`
	LowInsights            int                    `json:"low_insights"`
	TopPrioritizedInsights []CrossModuleInsight   `json:"top_prioritized_insights"`
	InsightsByType         map[string]int         `json:"insights_by_type"`
	AffectedEntitiesCount  map[string]int         `json:"affected_entities_count"`
	ConnectedModules       []string               `json:"connected_modules"`
	DataFreshness          time.Time              `json:"data_freshness"`
	Limitations            []string               `json:"limitations"`
}
