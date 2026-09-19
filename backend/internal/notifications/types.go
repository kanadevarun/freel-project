package notifications

import (
	"context"
	"time"

	"github.com/freel/backend/internal/approvals"
)

// Severity levels for notifications
const (
	SeverityInformational = "INFORMATIONAL"
	SeverityLow           = "LOW"
	SeverityMedium        = "MEDIUM"
	SeverityHigh          = "HIGH"
	SeverityCritical      = "CRITICAL"
)

// Priority levels for notifications
const (
	PriorityLow    = "LOW"
	PriorityMedium = "MEDIUM"
	PriorityHigh   = "HIGH"
	PriorityUrgent = "URGENT"
)

// Status states for notifications
const (
	StatusActive    = "ACTIVE"
	StatusResolved  = "RESOLVED"
	StatusDismissed = "DISMISSED"
	StatusExpired   = "EXPIRED"
)

// Source modules
const (
	ModuleShipments       = "SHIPMENTS"
	ModuleInvoices        = "INVOICES"
	ModuleContracts       = "CONTRACTS"
	ModuleApprovals       = "APPROVALS"
	ModuleAutomations     = "AUTOMATIONS"
	ModuleRecommendations = "RECOMMENDATIONS"
	ModuleSystem          = "SYSTEM"
)

// Notification types
const (
	TypeHighPriorityRecommendation = "HIGH_PRIORITY_RECOMMENDATION"
	TypeAssignedRecommendation     = "ASSIGNED_RECOMMENDATION"
	TypeCriticalShipmentException  = "CRITICAL_SHIPMENT_EXCEPTION"
	TypeMissedShipmentMilestone    = "MISSED_SHIPMENT_MILESTONE"
	TypeOverdueInvoiceRisk         = "OVERDUE_INVOICE_RISK"
	TypeExpiringContract           = "EXPIRING_CONTRACT"
	TypePendingApproval            = "PENDING_APPROVAL"
	TypeRejectedApproval           = "REJECTED_APPROVAL"
	TypeAutomationRunFailed        = "AUTOMATION_RUN_FAILED"
	TypeComplianceReviewRequired   = "COMPLIANCE_REVIEW_REQUIRED"
	TypeSystemAlert                = "SYSTEM_ALERT"
)

// Mail Providers
const (
	MailProviderSMTP = "smtp"
	MailProviderSES  = "ses"

	DefaultFromEmail   = "logisticshq26@gmail.com"
	DefaultTemplateDir = "internal/notifications/templates"
)

// Notification represents a centralized, real persisted in-app notification.
type Notification struct {
	ID                    int64      `json:"id" db:"id"`
	OrgID                 int64      `json:"org_id" db:"org_id"`
	UserID                *int64     `json:"user_id,omitempty" db:"user_id"`
	RoleTarget            *string    `json:"role_target,omitempty" db:"role_target"`
	SourceModule          string     `json:"source_module" db:"source_module"`
	SourceRecordType      string     `json:"source_record_type" db:"source_record_type"`
	SourceRecordID        string     `json:"source_record_id" db:"source_record_id"`
	RecommendationID      *int64     `json:"recommendation_id,omitempty" db:"recommendation_id"`
	ApprovalID            *int64     `json:"approval_id,omitempty" db:"approval_id"`
	AutomationExecutionID *int64     `json:"automation_execution_id,omitempty" db:"automation_execution_id"`
	NotificationType      string     `json:"notification_type" db:"notification_type"`
	Title                 string     `json:"title" db:"title"`
	Message               string     `json:"message" db:"message"`
	Severity              string     `json:"severity" db:"severity"`
	Priority              string     `json:"priority" db:"priority"`
	Status                string     `json:"status" db:"status"`
	DeliveryStatus        string     `json:"delivery_status" db:"delivery_status"`
	IsRead                bool       `json:"is_read" db:"is_read"`
	ReadAt                *time.Time `json:"read_at,omitempty" db:"read_at"`
	IsDismissed           bool       `json:"is_dismissed" db:"is_dismissed"`
	DismissedAt           *time.Time `json:"dismissed_at,omitempty" db:"dismissed_at"`
	IsAcknowledged        bool       `json:"is_acknowledged" db:"is_acknowledged"`
	AcknowledgedAt        *time.Time `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	AcknowledgedBy        *int64     `json:"acknowledged_by,omitempty" db:"acknowledged_by"`
	IsSnoozed             bool       `json:"is_snoozed" db:"is_snoozed"`
	SnoozedUntil          *time.Time `json:"snoozed_until,omitempty" db:"snoozed_until"`
	IsEscalated           bool       `json:"is_escalated" db:"is_escalated"`
	EscalationLevel       int        `json:"escalation_level" db:"escalation_level"`
	EscalatedAt           *time.Time `json:"escalated_at,omitempty" db:"escalated_at"`
	ActionRequired        bool       `json:"action_required" db:"action_required"`
	ActionURL             *string    `json:"action_url,omitempty" db:"action_url"`
	DedupHash             string     `json:"dedup_hash" db:"dedup_hash"`
	GroupKey              *string    `json:"group_key,omitempty" db:"group_key"`
	CorrelationID         string     `json:"correlation_id" db:"correlation_id"`
	AISummary             *string    `json:"ai_summary,omitempty" db:"ai_summary"`
	AIEscalationReason    *string    `json:"ai_escalation_reason,omitempty" db:"ai_escalation_reason"`
	AIPriorityScore       *float64   `json:"ai_priority_score,omitempty" db:"ai_priority_score"`
	ExpiresAt             *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	Metadata              *string    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`

	// Legacy compat fields
	Type string `json:"type,omitempty"`
}

// EscalationEvent records auditable escalation transitions for a notification.
type EscalationEvent struct {
	ID                  int64      `json:"id" db:"id"`
	NotificationID      int64      `json:"notification_id" db:"notification_id"`
	OrgID               int64      `json:"org_id" db:"org_id"`
	EscalationLevel     int        `json:"escalation_level" db:"escalation_level"`
	EscalationReason    string     `json:"escalation_reason" db:"escalation_reason"`
	PreviousSeverity    string     `json:"previous_severity" db:"previous_severity"`
	NewSeverity         string     `json:"new_severity" db:"new_severity"`
	TriggerType         string     `json:"trigger_type" db:"trigger_type"`
	ThresholdHours      int        `json:"threshold_hours" db:"threshold_hours"`
	AIEscalationSummary *string    `json:"ai_escalation_summary,omitempty" db:"ai_escalation_summary"`
	RecommendedAction   *string    `json:"recommended_action,omitempty" db:"recommended_action"`
	ActionProposalID    *string    `json:"action_proposal_id,omitempty" db:"action_proposal_id"`
	ApprovalID          *int64     `json:"approval_id,omitempty" db:"approval_id"`
	AcknowledgedAt      *time.Time `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	AcknowledgedBy      *int64     `json:"acknowledged_by,omitempty" db:"acknowledged_by"`
	ResolvedAt          *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolvedBy          *int64     `json:"resolved_by,omitempty" db:"resolved_by"`
	CorrelationID       string     `json:"correlation_id" db:"correlation_id"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
}

// UserNotificationPreferences stores user-specific notification preferences.
type UserNotificationPreferences struct {
	ID                     int64     `json:"id" db:"id"`
	UserID                 int64     `json:"user_id" db:"user_id"`
	OrgID                  int64     `json:"org_id" db:"org_id"`
	MinSeverity            string    `json:"min_severity" db:"min_severity"`
	InAppEnabled           bool      `json:"in_app_enabled" db:"in_app_enabled"`
	AssignedOnly           bool      `json:"assigned_only" db:"assigned_only"`
	ApprovalsEnabled       bool      `json:"approvals_enabled" db:"approvals_enabled"`
	AutomationsEnabled     bool      `json:"automations_enabled" db:"automations_enabled"`
	RecommendationsEnabled bool      `json:"recommendations_enabled" db:"recommendations_enabled"`
	FinanceEnabled         bool      `json:"finance_enabled" db:"finance_enabled"`
	OperationsEnabled      bool      `json:"operations_enabled" db:"operations_enabled"`
	ComplianceEnabled      bool      `json:"compliance_enabled" db:"compliance_enabled"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

// NotificationStats provides aggregate counts for the UI badge and dashboard.
type NotificationStats struct {
	Total          int `json:"total"`
	Unread         int `json:"unread"`
	ActionRequired int `json:"action_required"`
	Escalated      int `json:"escalated"`
	Acknowledged   int `json:"acknowledged"`
	Snoozed        int `json:"snoozed"`
	CriticalCount  int `json:"critical_count"`
	HighCount      int `json:"high_count"`
}

// NotificationFilter specifies query filters for listing notifications.
type NotificationFilter struct {
	OrgID          int64
	UserID         int64
	UserRole       string
	Severity       string
	SourceModule   string
	DeliveryStatus string
	IsRead         *bool
	IsDismissed    *bool
	IsAcknowledged *bool
	IsSnoozed      *bool
	ActionRequired *bool
	IsEscalated    *bool
	GroupKey       string
	Search         string
	Limit          int
	Offset         int
}

// Python AI Sidecar Request & Response DTOs
type NotificationAnalysisRequestDTO struct {
	OrgID                   int64                  `json:"org_id"`
	NotificationID          *int64                 `json:"notification_id,omitempty"`
	SourceModule            string                 `json:"source_module"`
	SourceRecordType        string                 `json:"source_record_type"`
	SourceRecordID          string                 `json:"source_record_id"`
	NotificationType        string                 `json:"notification_type"`
	Title                   string                 `json:"title"`
	Message                 string                 `json:"message"`
	Severity                string                 `json:"severity"`
	CurrentPriority         string                 `json:"current_priority"`
	AgeHours                float64                `json:"age_hours"`
	EscalationLevel         int                    `json:"escalation_level"`
	UnresolvedDurationHours float64                `json:"unresolved_duration_hours"`
	ContextPayload          map[string]interface{} `json:"context_payload"`
	CorrelationID           string                 `json:"correlation_id"`
}

type NotificationAnalysisResponseDTO struct {
	AISummary              string  `json:"ai_summary"`
	AIEscalationReason     string  `json:"ai_escalation_reason"`
	RecommendedPriority    string  `json:"recommended_priority"`
	PriorityScore          float64 `json:"priority_score"`
	RecommendedRoleTarget  string  `json:"recommended_role_target"`
	SuggestedAction        string  `json:"suggested_action"`
	RequiresApproval       bool    `json:"requires_approval"`
	RiskLevel              string  `json:"risk_level"`
	GroupKey               string  `json:"group_key"`
	OverloadReductionAdvice string `json:"overload_reduction_advice"`
	ConfidenceScore        float64 `json:"confidence_score"`
	CorrelationID          string  `json:"correlation_id"`
}

type EscalationDraftRequestDTO struct {
	OrgID          int64    `json:"org_id"`
	NotificationID int64    `json:"notification_id"`
	DraftType      string   `json:"draft_type"`
	SourceModule   string   `json:"source_module"`
	SourceRecordID string   `json:"source_record_id"`
	RecipientRole  string   `json:"recipient_role"`
	EscalationLevel int     `json:"escalation_level"`
	KeyFindings    []string `json:"key_findings"`
	CorrelationID  string   `json:"correlation_id"`
}

type EscalationDraftResponseDTO struct {
	Subject             string   `json:"subject"`
	BodyText            string   `json:"body_text"`
	RecommendedChannels []string `json:"recommended_channels"`
	IsExternal          bool     `json:"is_external"`
	RequiresApproval    bool     `json:"requires_approval"`
	ConfidenceScore     float64  `json:"confidence_score"`
	CorrelationID       string   `json:"correlation_id"`
}

// Service defines the notification and escalation operations.
type Service interface {
	// Outbound legacy operations
	SendEmail(ctx context.Context, to string, subject string, body string) error
	SendWhatsApp(ctx context.Context, phone string, message string) error
	SendInviteEmail(ctx context.Context, toEmail, token, orgName string) error

	// In-App Notification & Escalation Center operations
	GetUnreadNotifications(ctx context.Context, orgID int32) ([]Notification, error)
	ListNotifications(ctx context.Context, filter NotificationFilter) ([]Notification, int, error)
	GetNotification(ctx context.Context, orgID int64, id int64) (*Notification, error)
	GetUnreadCount(ctx context.Context, orgID int64, userID int64, role string) (int, error)
	GetStats(ctx context.Context, orgID int64, userID int64, role string) (*NotificationStats, error)
	MarkAsRead(ctx context.Context, orgID int32, notifID int32) error
	MarkRead(ctx context.Context, orgID int64, id int64, read bool) error
	MarkAllAsRead(ctx context.Context, orgID int64, userID int64, role string) error
	Dismiss(ctx context.Context, orgID int64, id int64) error
	Acknowledge(ctx context.Context, orgID int64, id int64, userID int64) error
	Snooze(ctx context.Context, orgID int64, id int64, durationMinutes int) error
	EscalateWithAI(ctx context.Context, orgID int64, id int64, userID int64, reason string) (*EscalationEvent, error)
	GenerateDraftWithAI(ctx context.Context, orgID int64, id int64, draftType string) (*EscalationDraftResponseDTO, error)
	AnalyzeWithAI(ctx context.Context, orgID int64, id int64) (*NotificationAnalysisResponseDTO, error)
	ListEscalations(ctx context.Context, orgID int64, limit int) ([]EscalationEvent, error)
	GetPreferences(ctx context.Context, orgID int64, userID int64) (*UserNotificationPreferences, error)
	UpdatePreferences(ctx context.Context, prefs *UserNotificationPreferences) error
	EvaluateNotifications(ctx context.Context, orgID int64) (int, error)
	SetApprovalsService(appr approvals.Service)
}
