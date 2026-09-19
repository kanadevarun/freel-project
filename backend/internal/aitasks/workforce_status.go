package aitasks

import (
	"fmt"
	"strings"
	"time"
)

// Authoritative 13 Workforce Status Constants
const (
	WorkforceStatusQueued                = "queued"
	WorkforceStatusClaimed               = "claimed"
	WorkforceStatusProcessing            = "processing"
	WorkforceStatusWaitingForApproval    = "waiting_for_approval"
	WorkforceStatusPaused                = "paused"
	WorkforceStatusRetrying              = "retrying"
	WorkforceStatusCompleted             = "completed"
	WorkforceStatusCompletedWithFailover = "completed_with_failover"
	WorkforceStatusCompletedInMockMode   = "completed_in_mock_mode"
	WorkforceStatusFailed                = "failed"
	WorkforceStatusCancelled             = "cancelled"
	WorkforceStatusStale                 = "stale"
	WorkforceStatusUnknown               = "unknown"
)

// WorkforceStatusMetadata provides safe, standardized display and action semantics for each state.
type WorkforceStatusMetadata struct {
	Value            string `json:"value"`
	Label            string `json:"label"`
	Description      string `json:"description"`
	Severity         string `json:"severity"` // "info", "warning", "success", "error", "neutral", "purple"
	IsTerminal       bool   `json:"is_terminal"`
	IsActionable     bool   `json:"is_actionable"`
	RetryPermitted   bool   `json:"retry_permitted"`
	ApprovalRequired bool   `json:"approval_required"`
}

// Canonical status registry definitions
var statusMetadataRegistry = map[string]WorkforceStatusMetadata{
	WorkforceStatusQueued: {
		Value:            WorkforceStatusQueued,
		Label:            "Queued",
		Description:      "Task is staged in queue waiting for an available worker.",
		Severity:         "neutral",
		IsTerminal:       false,
		IsActionable:     true, // Can cancel
		RetryPermitted:   false,
		ApprovalRequired: false,
	},
	WorkforceStatusClaimed: {
		Value:            WorkforceStatusClaimed,
		Label:            "Claimed",
		Description:      "Task has been leased by an active worker thread and is initializing.",
		Severity:         "info",
		IsTerminal:       false,
		IsActionable:     false,
		RetryPermitted:   false,
		ApprovalRequired: false,
	},
	WorkforceStatusProcessing: {
		Value:            WorkforceStatusProcessing,
		Label:            "Processing",
		Description:      "Agent is actively executing workflows, calling tools, or querying rates.",
		Severity:         "info",
		IsTerminal:       false,
		IsActionable:     true, // Can cancel
		RetryPermitted:   false,
		ApprovalRequired: false,
	},
	WorkforceStatusWaitingForApproval: {
		Value:            WorkforceStatusWaitingForApproval,
		Label:            "Waiting for Sign-Off",
		Description:      "Autonomous execution is paused at a governance gate awaiting human review.",
		Severity:         "purple",
		IsTerminal:       false,
		IsActionable:     true, // Can open approval modal
		RetryPermitted:   false,
		ApprovalRequired: true,
	},
	WorkforceStatusPaused: {
		Value:            WorkforceStatusPaused,
		Label:            "Paused",
		Description:      "Graph execution paused at an interrupt node.",
		Severity:         "warning",
		IsTerminal:       false,
		IsActionable:     true,
		RetryPermitted:   false,
		ApprovalRequired: true,
	},
	WorkforceStatusRetrying: {
		Value:            WorkforceStatusRetrying,
		Label:            "Retrying",
		Description:      "A transient provider error occurred; task is backed off for automatic retry.",
		Severity:         "warning",
		IsTerminal:       false,
		IsActionable:     true, // Can cancel
		RetryPermitted:   false,
		ApprovalRequired: false,
	},
	WorkforceStatusCompleted: {
		Value:            WorkforceStatusCompleted,
		Label:            "Completed",
		Description:      "AI work finished successfully with authorized actions persisted.",
		Severity:         "success",
		IsTerminal:       true,
		IsActionable:     false,
		RetryPermitted:   false,
		ApprovalRequired: false,
	},
	WorkforceStatusCompletedWithFailover: {
		Value:            WorkforceStatusCompletedWithFailover,
		Label:            "Completed (Failover)",
		Description:      "Task finished successfully after primary provider failed and switched to failover.",
		Severity:         "info",
		IsTerminal:       true,
		IsActionable:     false,
		RetryPermitted:   false,
		ApprovalRequired: false,
	},
	WorkforceStatusCompletedInMockMode: {
		Value:            WorkforceStatusCompletedInMockMode,
		Label:            "Completed (Mock Dev)",
		Description:      "Task finished in development test mode using deterministic mock response.",
		Severity:         "warning",
		IsTerminal:       true,
		IsActionable:     false,
		RetryPermitted:   false,
		ApprovalRequired: false,
	},
	WorkforceStatusFailed: {
		Value:            WorkforceStatusFailed,
		Label:            "Failed",
		Description:      "Execution failed due to unrecoverable error or exceeded maximum retry attempts.",
		Severity:         "error",
		IsTerminal:       true,
		IsActionable:     true, // Can manual retry
		RetryPermitted:   true,
		ApprovalRequired: false,
	},
	WorkforceStatusCancelled: {
		Value:            WorkforceStatusCancelled,
		Label:            "Cancelled",
		Description:      "Task was safely cancelled by operator or reviewer.",
		Severity:         "neutral",
		IsTerminal:       true,
		IsActionable:     false,
		RetryPermitted:   false,
		ApprovalRequired: false,
	},
	WorkforceStatusStale: {
		Value:            WorkforceStatusStale,
		Label:            "Stale / Needs Attention",
		Description:      "Worker lease expired without heartbeats; queued for automatic recovery or retry.",
		Severity:         "error",
		IsTerminal:       false,
		IsActionable:     true, // Can cancel or retry
		RetryPermitted:   true,
		ApprovalRequired: false,
	},
	WorkforceStatusUnknown: {
		Value:            WorkforceStatusUnknown,
		Label:            "Unknown",
		Description:      "Task status could not be authoritatively resolved from database.",
		Severity:         "neutral",
		IsTerminal:       false,
		IsActionable:     false,
		RetryPermitted:   false,
		ApprovalRequired: false,
	},
}

// GetStatusMetadata returns the canonical metadata descriptor for a given workforce status.
func GetStatusMetadata(status string) WorkforceStatusMetadata {
	norm := strings.ToLower(strings.TrimSpace(status))
	if meta, exists := statusMetadataRegistry[norm]; exists {
		return meta
	}
	return statusMetadataRegistry[WorkforceStatusUnknown]
}

// AgentWorkforceStat represents high-level metrics for one agent/workflow type.
type AgentWorkforceStat struct {
	AgentKey         string `json:"agent_key"`        // "pricing", "sales", "operations", etc.
	DisplayName      string `json:"display_name"`     // "Pricing Analyst"
	Module           string `json:"module"`           // "PRICING", "SALES", etc.
	ActiveTasks      int64  `json:"active_tasks"`
	Completed24h     int64  `json:"completed_24h"`
	Failed24h        int64  `json:"failed_24h"`
	WaitingApprovals int64  `json:"waiting_approvals"`
	Status           string `json:"status"`           // "ACTIVE", "IDLE", "ATTENTION", "OFFLINE"
}

// FailureCategoryStat aggregates recent failures by safe category.
type FailureCategoryStat struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

// WorkforceHealth exposes system readiness and component availability.
type WorkforceHealth struct {
	OverallStatus         string     `json:"overall_status"`          // "healthy", "degraded", "unavailable"
	BackendStatus         string     `json:"backend_status"`          // "healthy"
	SidecarStatus         string     `json:"sidecar_status"`          // "healthy", "unavailable"
	WorkerStatus          string     `json:"worker_status"`           // "healthy", "degraded", "idle"
	QueueStatus           string     `json:"queue_status"`            // "healthy", "lagging", "blocked"
	CheckpointStatus      string     `json:"checkpoint_status"`       // "healthy", "degraded"
	PrimaryProviderStatus string     `json:"primary_provider_status"` // "healthy", "degraded"
	FailoverStatus        string     `json:"failover_status"`         // "ready", "degraded"
	LastWorkerHeartbeat   *time.Time `json:"last_worker_heartbeat,omitempty"`
}

// WorkforceSummary aggregates live metrics for an authenticated organization.
type WorkforceSummary struct {
	OrgID                   int64                         `json:"org_id"`
	TotalActiveTasks        int64                         `json:"total_active_tasks"`
	QueuedTasks             int64                         `json:"queued_tasks"`
	ProcessingTasks         int64                         `json:"processing_tasks"`
	WaitingForApprovalTasks int64                         `json:"waiting_for_approval_tasks"`
	RetryingTasks           int64                         `json:"retrying_tasks"`
	FailedTasks             int64                         `json:"failed_tasks"`
	StaleTasks              int64                         `json:"stale_tasks"`
	CompletedRecent24h      int64                         `json:"completed_recent_24h"`
	CompletedWithFailover   int64                         `json:"completed_with_failover_24h"`
	CompletedInMockMode     int64                         `json:"completed_in_mock_mode_24h"`
	AvgDurationMs           int64                         `json:"avg_duration_ms"`
	ActiveAgentsCount       int                           `json:"active_agents_count"`
	ByAgent                 map[string]AgentWorkforceStat `json:"by_agent"`
	RecentFailures          []FailureCategoryStat         `json:"recent_failures"`
	Health                  WorkforceHealth               `json:"health"`
	LastUpdated             time.Time                     `json:"last_updated"`
}

// WorkforceTaskItem is a safe, enriched representation of an AI task for operational views.
type WorkforceTaskItem struct {
	ID            int64                   `json:"id"`
	OrgID         int64                   `json:"org_id"`
	TaskType      string                  `json:"task_type"`
	AgentKey      string                  `json:"agent_key"`
	AgentName     string                  `json:"agent_name"`
	Module        string                  `json:"module"`
	RelatedRef    string                  `json:"related_ref"`
	RelatedID     string                  `json:"related_id"`
	Status        string                  `json:"status"` // Authoritative 13-status value
	StatusMeta    WorkforceStatusMetadata `json:"status_meta"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
	StartedAt     *time.Time              `json:"started_at,omitempty"`
	CompletedAt   *time.Time              `json:"completed_at,omitempty"`
	DurationMs    *int64                  `json:"duration_ms,omitempty"`
	RetryCount    int                     `json:"retry_count"`
	MaxRetries    int                     `json:"max_retries"`
	IsStale       bool                    `json:"is_stale"`
	ApprovalID    *int64                  `json:"approval_id,omitempty"`
	ErrorCategory *string                 `json:"error_category,omitempty"`
	ErrorMessage  *string                 `json:"error_message,omitempty"` // Secret-redacted
	CanRetry      bool                    `json:"can_retry"`
	CanCancel     bool                    `json:"can_cancel"`
	NavigationURL string                  `json:"navigation_url"`
	ThreadID      *string                 `json:"thread_id,omitempty"`
	CorrelationID *string                 `json:"correlation_id,omitempty"`
}

// WorkforceTaskFilter specifies filtering criteria for querying AI workforce tasks.
type WorkforceTaskFilter struct {
	Status           string `json:"status,omitempty"`
	AgentKey         string `json:"agent_key,omitempty"`
	Module           string `json:"module,omitempty"`
	RequiresApproval *bool  `json:"requires_approval,omitempty"`
	FailedOrStale    *bool  `json:"failed_or_stale,omitempty"`
	Search           string `json:"search,omitempty"`
	Limit            int    `json:"limit,omitempty"`
	Offset           int    `json:"offset,omitempty"`
}

// MapTaskTypeToAgent maps raw task_type strings to canonical agent identity, module, and navigation.
func MapTaskTypeToAgent(taskType string) (agentKey, agentName, module, defaultNav string) {
	upper := strings.ToUpper(strings.TrimSpace(taskType))
	switch {
	case strings.Contains(upper, "PRICING"):
		return "pricing", "Pricing Analyst", "PRICING", "/dashboard/rfqs"
	case strings.Contains(upper, "EMAIL") || strings.Contains(upper, "SALES"):
		return "sales", "Sales Email Parser", "SALES", "/dashboard/leads"
	case strings.Contains(upper, "CARRIER") || strings.Contains(upper, "OPERATIONS") || strings.Contains(upper, "TRACKING"):
		return "operations", "Operations Carrier Tracker", "OPERATIONS", "/dashboard/shipments"
	case strings.Contains(upper, "CONTRACT") || strings.Contains(upper, "DOC_PROCESS") || strings.Contains(upper, "RATE_EXTRACT"):
		return "contracts", "Contracts Intelligence", "CONTRACTS", "/dashboard/contracts"
	case strings.Contains(upper, "COMPLIANCE") || strings.Contains(upper, "DOC_VERIFY"):
		return "compliance", "Compliance Auditor", "COMPLIANCE", "/dashboard/shipments"
	case strings.Contains(upper, "FINANCE") || strings.Contains(upper, "BILL_RECONCILE") || strings.Contains(upper, "INVOICE"):
		return "finance", "Finance Invoice Auditor", "FINANCE", "/dashboard/invoices"
	case strings.Contains(upper, "LEAD"):
		return "leads", "Lead Scoring Specialist", "LEADS", "/dashboard/leads"
	case strings.Contains(upper, "OUTREACH"):
		return "outreach", "Outreach Copywriter", "OUTREACH", "/dashboard/outreach"
	default:
		return "general", "General AI Assistant", "OPERATIONS", "/dashboard"
	}
}

// ResolveTaskWorkforceStatus maps a task row and execution telemetry into one authoritative status.
func ResolveTaskWorkforceStatus(task *Task, isFailover bool, isMock bool) (string, WorkforceStatusMetadata) {
	if task == nil {
		return WorkforceStatusUnknown, GetStatusMetadata(WorkforceStatusUnknown)
	}

	upperStatus := strings.ToUpper(strings.TrimSpace(task.Status))

	// 1. Check for Stale processing condition (lease expired without heartbeat)
	if upperStatus == StatusProcessing && task.LeaseExpiresAt != nil && task.LeaseExpiresAt.Before(time.Now()) {
		return WorkforceStatusStale, GetStatusMetadata(WorkforceStatusStale)
	}

	// 2. Map standard task states
	switch upperStatus {
	case StatusQueued:
		if task.WorkerID != nil && *task.WorkerID != "" {
			return WorkforceStatusClaimed, GetStatusMetadata(WorkforceStatusClaimed)
		}
		return WorkforceStatusQueued, GetStatusMetadata(WorkforceStatusQueued)

	case StatusProcessing:
		return WorkforceStatusProcessing, GetStatusMetadata(WorkforceStatusProcessing)

	case StatusWaitingForApproval:
		return WorkforceStatusWaitingForApproval, GetStatusMetadata(WorkforceStatusWaitingForApproval)

	case StatusRetrying:
		return WorkforceStatusRetrying, GetStatusMetadata(WorkforceStatusRetrying)

	case StatusCompleted:
		if isMock {
			return WorkforceStatusCompletedInMockMode, GetStatusMetadata(WorkforceStatusCompletedInMockMode)
		}
		if isFailover {
			return WorkforceStatusCompletedWithFailover, GetStatusMetadata(WorkforceStatusCompletedWithFailover)
		}
		return WorkforceStatusCompleted, GetStatusMetadata(WorkforceStatusCompleted)

	case StatusFailed:
		return WorkforceStatusFailed, GetStatusMetadata(WorkforceStatusFailed)

	case StatusCancelled, StatusRejected:
		return WorkforceStatusCancelled, GetStatusMetadata(WorkforceStatusCancelled)

	default:
		return WorkforceStatusUnknown, GetStatusMetadata(WorkforceStatusUnknown)
	}
}

// ResolveRelatedRef resolves display label, related entity ID, and safe navigation link.
func ResolveRelatedRef(task *Task) (label, id, navURL string) {
	if task == nil {
		return "Unknown Task", "", "/dashboard"
	}

	// Check EntityType and EntityID first
	if task.EntityType != nil && task.EntityID != nil && *task.EntityID != "" {
		eType := strings.ToLower(strings.TrimSpace(*task.EntityType))
		eID := *task.EntityID
		switch eType {
		case "rfq":
			return fmt.Sprintf("RFQ #%s", eID), eID, fmt.Sprintf("/dashboard/rfqs/%s", eID)
		case "lead":
			return fmt.Sprintf("Lead #%s", eID), eID, fmt.Sprintf("/dashboard/leads/%s", eID)
		case "shipment":
			return fmt.Sprintf("Shipment #%s", eID), eID, fmt.Sprintf("/dashboard/shipments/%s", eID)
		case "invoice":
			return fmt.Sprintf("Invoice #%s", eID), eID, "/dashboard/invoices"
		case "contract", "document":
			return fmt.Sprintf("Contract #%s", eID), eID, fmt.Sprintf("/dashboard/contracts/%s", eID)
		default:
			return fmt.Sprintf("%s #%s", strings.ToUpper(eType), eID), eID, "/dashboard"
		}
	}

	// Check ResourceID
	if task.ResourceID != nil && *task.ResourceID > 0 {
		_, _, _, defNav := MapTaskTypeToAgent(task.TaskType)
		return fmt.Sprintf("Record #%d", *task.ResourceID), fmt.Sprintf("%d", *task.ResourceID), defNav
	}

	// Check DocumentID
	if task.DocumentID != nil && *task.DocumentID != "" {
		return fmt.Sprintf("Doc #%s", *task.DocumentID), *task.DocumentID, "/dashboard/contracts"
	}

	// Fallback to Task ID
	_, _, _, defNav := MapTaskTypeToAgent(task.TaskType)
	return fmt.Sprintf("Task #%d", task.ID), fmt.Sprintf("%d", task.ID), defNav
}
