package operations_automation

import (
	"encoding/json"
	"time"
)

// ShipmentOperationsAnalysis represents an operational risk assessment record
type ShipmentOperationsAnalysis struct {
	ID                   int64           `db:"id" json:"id"`
	OrgID                int64           `db:"org_id" json:"org_id"`
	ShipmentID           int64           `db:"shipment_id" json:"shipment_id"`
	RiskLevel            string          `db:"risk_level" json:"risk_level"` // LOW, MEDIUM, HIGH, CRITICAL
	RiskScore            float64         `db:"risk_score" json:"risk_score"`
	OperationalSummary   string          `db:"operational_summary" json:"operational_summary"`
	DeterministicSignals json.RawMessage `db:"deterministic_signals" json:"deterministic_signals,omitempty"`
	KeyRisks             json.RawMessage `db:"key_risks" json:"key_risks,omitempty"`
	RecommendedNextSteps json.RawMessage `db:"recommended_next_steps" json:"recommended_next_steps,omitempty"`
	Evidence             json.RawMessage `db:"evidence" json:"evidence,omitempty"`
	ConfidenceScore      float64         `db:"confidence_score" json:"confidence_score"`
	CorrelationID        string          `db:"correlation_id" json:"correlation_id"`
	CreatedAt            time.Time       `db:"created_at" json:"created_at"`
}

// ShipmentCommunicationDraft represents an AI-generated, human-editable communication draft
type ShipmentCommunicationDraft struct {
	ID               int64     `db:"id" json:"id"`
	OrgID            int64     `db:"org_id" json:"org_id"`
	ShipmentID       int64     `db:"shipment_id" json:"shipment_id"`
	DraftType        string    `db:"draft_type" json:"draft_type"` // CARRIER_FOLLOWUP, CUSTOMER_UPDATE, INTERNAL_ESCALATION
	Subject          string    `db:"subject" json:"subject"`
	CustomerWording  string    `db:"customer_wording" json:"customer_wording"`
	InternalNotes    *string   `db:"internal_notes" json:"internal_notes,omitempty"`
	RecipientName    string    `db:"recipient_name" json:"recipient_name"`
	RecipientEmail   string    `db:"recipient_email" json:"recipient_email"`
	Status           string    `db:"status" json:"status"` // DRAFT, PENDING_APPROVAL, APPROVED, REJECTED, EXECUTED
	RequiresApproval bool      `db:"requires_approval" json:"requires_approval"`
	ApprovalID       *int64    `db:"approval_id" json:"approval_id,omitempty"`
	ActionProposalID *string   `db:"action_proposal_id" json:"action_proposal_id,omitempty"`
	ExecutionID      *int64    `db:"execution_id" json:"execution_id,omitempty"`
	CreatedByUserID  *int64    `db:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CorrelationID    string    `db:"correlation_id" json:"correlation_id"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// DeterministicSignalsDTO represents Go-calculated operational indicators
type DeterministicSignalsDTO struct {
	HasOverdueMilestone      bool     `json:"has_overdue_milestone"`
	OverdueMilestones        []string `json:"overdue_milestones"`
	ApproachingMilestones    []string `json:"approaching_milestones"`
	HasActiveException       bool     `json:"has_active_exception"`
	ActiveExceptionsCount    int      `json:"active_exceptions_count"`
	CriticalExceptionsCount  int      `json:"critical_exceptions_count"`
	IsTrackingStale          bool     `json:"is_tracking_stale"`
	HoursSinceLastTracking   float64  `json:"hours_since_last_tracking"`
	IsDelayed                bool     `json:"is_delayed"`
	DelayDays                float64  `json:"delay_days"`
	MissingDocumentsCount    int      `json:"missing_documents_count"`
	MissingDocuments         []string `json:"missing_documents"`
	CarrierResponseGapHours  float64  `json:"carrier_response_gap_hours"`
	IsCarrierResponseOverdue bool     `json:"is_carrier_response_overdue"`
	RiskScore                float64  `json:"risk_score"`
	RiskLevel                string   `json:"risk_level"`
	RequiresApproval         bool     `json:"requires_approval"`
}

// ShipmentOperationsOverview represents the consolidated operational response
type ShipmentOperationsOverview struct {
	ShipmentID           int64                       `json:"shipment_id"`
	CarrierSCAC          string                      `json:"carrier_scac"`
	CarrierName          string                      `json:"carrier_name"`
	Status               string                      `json:"status"`
	MBLNumber            string                      `json:"mbl_number"`
	BookingNumber        string                      `json:"booking_number"`
	OriginPort           string                      `json:"origin_port"`
	DestinationPort      string                      `json:"destination_port"`
	VesselName           string                      `json:"vessel_name"`
	VoyageNumber         string                      `json:"voyage_number"`
	ETD                  *string                     `json:"etd,omitempty"`
	ETA                  *string                     `json:"eta,omitempty"`
	CustomerName         string                      `json:"customer_name"`
	CustomerTier         string                      `json:"customer_tier"`
	Signals              DeterministicSignalsDTO     `json:"signals"`
	LatestAnalysis       *ShipmentOperationsAnalysis `json:"latest_analysis,omitempty"`
	LatestDraft          *ShipmentCommunicationDraft `json:"latest_draft,omitempty"`
	PrioritizedExceptions json.RawMessage             `json:"prioritized_exceptions,omitempty"`
	Recommendations      json.RawMessage             `json:"recommendations,omitempty"`
	PendingApproval      bool                        `json:"pending_approval"`
	ApprovalID           *int64                      `json:"approval_id,omitempty"`
	CorrelationID        string                      `json:"correlation_id"`
}

// GenerateDraftInput represents input for drafting an operational communication
type GenerateDraftInput struct {
	DraftType        string  `json:"draft_type"` // CARRIER_FOLLOWUP, CUSTOMER_UPDATE, INTERNAL_ESCALATION
	RecipientName    string  `json:"recipient_name,omitempty"`
	RecipientEmail   string  `json:"recipient_email,omitempty"`
	UserInstructions string  `json:"user_instructions,omitempty"`
	Tone             string  `json:"tone,omitempty"`
}

// UpdateDraftInput represents operator edits to a communication draft
type UpdateDraftInput struct {
	Subject         string `json:"subject"`
	CustomerWording string `json:"customer_wording"`
	InternalNotes   string `json:"internal_notes,omitempty"`
	RecipientName   string `json:"recipient_name,omitempty"`
	RecipientEmail  string `json:"recipient_email,omitempty"`
}

// SubmitDraftApprovalInput represents input when requesting managerial sign-off
type SubmitDraftApprovalInput struct {
	Reason string `json:"reason"`
}
