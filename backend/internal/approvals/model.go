package approvals

import (
	"time"
)

const (
	StatusDraft              = "Draft"
	StatusPendingApproval    = "Pending"
	StatusInReview           = "In Review"
	StatusApproved           = "Approved"
	StatusRejected           = "Rejected"
	StatusReturnedForChanges = "Returned for Changes"
	StatusExecuting          = "Executing"
	StatusCompleted          = "Completed"
	StatusFailed             = "Failed"
	StatusPartiallyCompleted = "Partially Completed"
	StatusCancelled          = "Cancelled"
	StatusExpired            = "Expired"
	StatusOverdue            = "Overdue"
)

type ApprovalRequest struct {
	ID                int64      `db:"id" json:"id"`
	OrgID             int64      `db:"org_id" json:"org_id"`
	RequestCode       string     `db:"request_code" json:"request_code"`
	Title             string     `db:"title" json:"title"`
	Category          string     `db:"category" json:"category"` // DOCUMENTS, COMMERCIAL, OPERATIONS, FINANCE
	Type              string     `db:"type" json:"type"`         // Document Approval, Commercial Approval, Operations Approval, Finance Approval, AI Action
	Status            string     `db:"status" json:"status"`     // Pending, In Review, Approved, Rejected, Returned for Changes, Executing, Completed, Failed, Cancelled, Expired
	Priority          string     `db:"priority" json:"priority"` // LOW, MEDIUM, HIGH, URGENT
	RelatedEntityType *string    `db:"related_entity_type" json:"related_entity_type,omitempty"`
	RelatedEntityID   *int64     `db:"related_entity_id" json:"related_entity_id,omitempty"`
	RelatedRef        *string    `db:"related_ref" json:"related_ref,omitempty"`
	CustomerName      *string    `db:"customer_name" json:"customer_name,omitempty"`
	CustomerID        *int64     `db:"customer_id" json:"customer_id,omitempty"`
	ShipmentID        *int64     `db:"shipment_id" json:"shipment_id,omitempty"`
	DocumentID        *int64     `db:"document_id" json:"document_id,omitempty"`
	BookingID         *int64     `db:"booking_id" json:"booking_id,omitempty"`
	RequestedByID     *int64     `db:"requested_by_id" json:"requested_by_id,omitempty"`
	RequestedByName   string     `db:"requested_by_name" json:"requested_by_name"`
	Department        *string    `db:"department" json:"department,omitempty"`
	Avatar            *string    `db:"avatar" json:"avatar,omitempty"`
	DueDate           *time.Time `db:"due_date" json:"due_date,omitempty"`
	DueText           *string    `db:"due_text" json:"due_text,omitempty"`
	AssignedTo        *string    `db:"assigned_to" json:"assigned_to,omitempty"`
	ApprovedBy        *string    `db:"approved_by" json:"approved_by,omitempty"`
	ApprovedAt        *time.Time `db:"approved_at" json:"approved_at,omitempty"`
	RejectedBy        *string    `db:"rejected_by" json:"rejected_by,omitempty"`
	RejectedAt        *time.Time `db:"rejected_at" json:"rejected_at,omitempty"`
	RejectionReason   *string    `db:"rejection_reason" json:"rejection_reason,omitempty"`
	Comments          *string    `db:"comments" json:"comments,omitempty"`
	Description       *string    `db:"description" json:"description,omitempty"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`

	// Unified AI HITL and Centralized Action Fields
	ActorType          string     `db:"actor_type" json:"actor_type"`
	Source             *string    `db:"source" json:"source,omitempty"`
	ActionName         *string    `db:"action_name" json:"action_name,omitempty"`
	RiskLevel          string     `db:"risk_level" json:"risk_level"`
	RequiredPermission *string    `db:"required_permission" json:"required_permission,omitempty"`
	AITaskID           *int64     `db:"ai_task_id" json:"ai_task_id,omitempty"`
	ThreadID           *string    `db:"thread_id" json:"thread_id,omitempty"`
	CheckpointID       *string    `db:"checkpoint_id" json:"checkpoint_id,omitempty"`
	ProposedPayload    *string    `db:"proposed_payload" json:"proposed_payload,omitempty"`
	ApprovalReference  *string    `db:"approval_reference" json:"approval_reference,omitempty"`
	CorrelationID      *string    `db:"correlation_id" json:"correlation_id,omitempty"`
	IdempotencyKey     *string    `db:"idempotency_key" json:"idempotency_key,omitempty"`
	ExpiresAt          *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	CancelledBy        *string    `db:"cancelled_by" json:"cancelled_by,omitempty"`
	CancelledAt        *time.Time `db:"cancelled_at" json:"cancelled_at,omitempty"`

	// Phase 2 Task 2.9 HITL Expansion Fields
	ExecutionStatus       string     `db:"execution_status" json:"execution_status"` // NOT_STARTED, EXECUTING, COMPLETED, FAILED, PARTIALLY_COMPLETED
	ExecutionResult       *string    `db:"execution_result" json:"execution_result,omitempty"`
	ExecutionError        *string    `db:"execution_error" json:"execution_error,omitempty"`
	ExecutionRetries      int        `db:"execution_retries" json:"execution_retries"`
	SourceModule          *string    `db:"source_module" json:"source_module,omitempty"`
	SourceRecordType      *string    `db:"source_record_type" json:"source_record_type,omitempty"`
	SourceRecordID        *string    `db:"source_record_id" json:"source_record_id,omitempty"`
	SourceRecordSnapshot  *string    `db:"source_record_snapshot" json:"source_record_snapshot,omitempty"`
	Evidence              *string    `db:"evidence" json:"evidence,omitempty"`
	ImpactSummary         *string    `db:"impact_summary" json:"impact_summary,omitempty"`
	IsReversible          bool       `db:"is_reversible" json:"is_reversible"`
	ExternalCommunication bool       `db:"external_communication" json:"external_communication"`
	RequiredApprovalLevel string     `db:"required_approval_level" json:"required_approval_level"`
	ReturnedBy            *string    `db:"returned_by" json:"returned_by,omitempty"`
	ReturnedAt            *time.Time `db:"returned_at" json:"returned_at,omitempty"`
	ReturnedReason        *string    `db:"returned_reason" json:"returned_reason,omitempty"`
}

type ApprovalDecision struct {
	ID            int64     `db:"id" json:"id"`
	OrgID         int64     `db:"org_id" json:"org_id"`
	ApprovalID    int64     `db:"approval_id" json:"approval_id"`
	ActionName    *string   `db:"action_name" json:"action_name,omitempty"`
	Decision      string    `db:"decision" json:"decision"` // APPROVE, REJECT, RETURN_FOR_CHANGES, CANCEL, EXPIRE
	ActorID       *int64    `db:"actor_id" json:"actor_id,omitempty"`
	ActorName     string    `db:"actor_name" json:"actor_name"`
	Reason        *string   `db:"reason" json:"reason,omitempty"`
	Notes         *string   `db:"notes" json:"notes,omitempty"`
	CorrelationID *string   `db:"correlation_id" json:"correlation_id,omitempty"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type ActionPreview struct {
	ApprovalID            int64                  `json:"approval_id"`
	ActionName            string                 `json:"action_name"`
	ActionDescription     string                 `json:"action_description"`
	TargetRecordType      string                 `json:"target_record_type"`
	TargetRecordID        string                 `json:"target_record_id"`
	TargetRecordTitle     string                 `json:"target_record_title,omitempty"`
	CurrentValue          map[string]interface{} `json:"current_value,omitempty"`
	ProposedValue         map[string]interface{} `json:"proposed_value,omitempty"`
	RequestedBy           string                 `json:"requested_by"`
	Reason                string                 `json:"reason"`
	Evidence              string                 `json:"evidence"`
	ExpectedImpact        string                 `json:"expected_impact"`
	RiskLevel             string                 `json:"risk_level"`
	ExternalCommunication bool                   `json:"external_communication"`
	DataDomainsAffected   []string               `json:"data_domains_affected"`
	IsReversible          bool                   `json:"is_reversible"`
	RequiredApprovalLevel string                 `json:"required_approval_level"`
	IsStale               bool                   `json:"is_stale"`
	StaleReason           string                 `json:"stale_reason,omitempty"`
	MessagePreview        *MessagePreviewDetails `json:"message_preview,omitempty"`
}

type MessagePreviewDetails struct {
	Recipient   string   `json:"recipient"`
	Channel     string   `json:"channel"` // EMAIL, SMS, WHATSAPP, SYSTEM
	Subject     string   `json:"subject,omitempty"`
	Body        string   `json:"body"`
	Attachments []string `json:"attachments,omitempty"`
}

type ApprovalRequirements struct {
	ActionName            string `json:"action_name"`
	Category              string `json:"category"`
	RiskLevel             string `json:"risk_level"`
	RequiresConfirmation  bool   `json:"requires_confirmation"`
	RequiredPermission    string `json:"required_permission"`
	RequiredApprovalLevel string `json:"required_approval_level"`
	SeparationOfDuties    bool   `json:"separation_of_duties"`
	ApprovalPolicySummary string `json:"approval_policy_summary"`
}

type ApprovalStats struct {
	Pending       int    `json:"pending"`
	PendingTrend  string `json:"pending_trend"`
	Approved      int    `json:"approved"`
	ApprovedTrend string `json:"approved_trend"`
	Rejected      int    `json:"rejected"`
	RejectedTrend string `json:"rejected_trend"`
	Returned      int    `json:"returned"`
	Overdue       int    `json:"overdue"`
}

type CreateApprovalInput struct {
	Title             string `json:"title"`
	Category          string `json:"category"`
	Type              string `json:"type"`
	Priority          string `json:"priority"`
	RelatedRef        string `json:"related_ref"`
	RelatedEntityType string `json:"related_entity_type"`
	RelatedEntityID   int64  `json:"related_entity_id"`
	CustomerName      string `json:"customer_name"`
	CustomerID        int64  `json:"customer_id"`
	ShipmentID        int64  `json:"shipment_id"`
	DocumentID        int64  `json:"document_id"`
	BookingID         int64  `json:"booking_id"`
	RequestedByID     int64  `json:"requested_by_id"`
	RequestedByName   string `json:"requested_by_name"`
	Department        string `json:"department"`
	DueDate           string `json:"due_date"`
	Description       string `json:"description"`

	// AI Fields optional on creation
	ActorType          string  `json:"actor_type,omitempty"`
	Source             string  `json:"source,omitempty"`
	ActionName         string  `json:"action_name,omitempty"`
	RiskLevel          string  `json:"risk_level,omitempty"`
	RequiredPermission string  `json:"required_permission,omitempty"`
	AITaskID           int64   `json:"ai_task_id,omitempty"`
	ThreadID           string  `json:"thread_id,omitempty"`
	CheckpointID       string  `json:"checkpoint_id,omitempty"`
	ProposedPayload    string  `json:"proposed_payload,omitempty"`
	ApprovalReference  string  `json:"approval_reference,omitempty"`
	CorrelationID      string  `json:"correlation_id,omitempty"`
	IdempotencyKey     string  `json:"idempotency_key,omitempty"`
	ExpiresAt          *string `json:"expires_at,omitempty"`

	// Extended preview & tracking metadata
	SourceModule          string `json:"source_module,omitempty"`
	SourceRecordType      string `json:"source_record_type,omitempty"`
	SourceRecordID        string `json:"source_record_id,omitempty"`
	SourceRecordSnapshot  string `json:"source_record_snapshot,omitempty"`
	Evidence              string `json:"evidence,omitempty"`
	ImpactSummary         string `json:"impact_summary,omitempty"`
	IsReversible          bool   `json:"is_reversible,omitempty"`
	ExternalCommunication bool   `json:"external_communication,omitempty"`
	RequiredApprovalLevel string `json:"required_approval_level,omitempty"`
}

type ProposeAIApprovalInput struct {
	Title              string  `json:"title"`
	Category           string  `json:"category"`
	Type               string  `json:"type"`
	Priority           string  `json:"priority"`
	RelatedEntityType  string  `json:"related_entity_type"`
	RelatedEntityID    int64   `json:"related_entity_id"`
	RelatedRef         string  `json:"related_ref"`
	CustomerName       string  `json:"customer_name"`
	CustomerID         int64   `json:"customer_id"`
	ShipmentID         int64   `json:"shipment_id"`
	RequestedByID      int64   `json:"requested_by_id"`
	RequestedByName    string  `json:"requested_by_name"`
	Department         string  `json:"department"`
	Description        string  `json:"description"`
	ActorType          string  `json:"actor_type"`
	Source             string  `json:"source"`
	ActionName         string  `json:"action_name"`
	RiskLevel          string  `json:"risk_level"`
	RequiredPermission string  `json:"required_permission"`
	AITaskID           int64   `json:"ai_task_id,omitempty"`
	ThreadID           string  `json:"thread_id,omitempty"`
	CheckpointID       string  `json:"checkpoint_id,omitempty"`
	ProposedPayload    string  `json:"proposed_payload,omitempty"`
	ApprovalReference  string  `json:"approval_reference"`
	CorrelationID      string  `json:"correlation_id,omitempty"`
	IdempotencyKey     string  `json:"idempotency_key,omitempty"`
	ExpiresInHours     int     `json:"expires_in_hours,omitempty"`

	// Extended preview & governance metadata
	SourceModule          string `json:"source_module,omitempty"`
	SourceRecordType      string `json:"source_record_type,omitempty"`
	SourceRecordID        string `json:"source_record_id,omitempty"`
	SourceRecordSnapshot  string `json:"source_record_snapshot,omitempty"`
	Evidence              string `json:"evidence,omitempty"`
	ImpactSummary         string `json:"impact_summary,omitempty"`
	IsReversible          bool   `json:"is_reversible,omitempty"`
	ExternalCommunication bool   `json:"external_communication,omitempty"`
	RequiredApprovalLevel string `json:"required_approval_level,omitempty"`
}

type ActionApprovalInput struct {
	Action string `json:"action"` // APPROVE, REJECT, RETURN, CANCEL
	Reason string `json:"reason"`
	Notes  string `json:"notes"`
}

type ReturnApprovalInput struct {
	Reason string `json:"reason"`
	Notes  string `json:"notes"`
}
