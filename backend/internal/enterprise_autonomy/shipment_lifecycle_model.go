package enterprise_autonomy

import (
	"errors"
	"time"
)

// ShipmentLifecycleStage defines the continuous stages of an autonomous shipment
type ShipmentLifecycleStage string

const (
	StageShipmentCreated ShipmentLifecycleStage = "SHIPMENT_CREATED"
	StagePlanning        ShipmentLifecycleStage = "PLANNING"
	StageBooked          ShipmentLifecycleStage = "BOOKED"
	StageInTransit       ShipmentLifecycleStage = "IN_TRANSIT"
	StageMonitoring      ShipmentLifecycleStage = "MONITORING"
	StageDelivered       ShipmentLifecycleStage = "DELIVERED"
	StagePostDelivery    ShipmentLifecycleStage = "POST_DELIVERY"
	StageCompleted       ShipmentLifecycleStage = "COMPLETED"
)

var (
	ErrInvalidStageTransition       = errors.New("invalid shipment lifecycle stage transition")
	ErrShipmentWorkflowNotFound     = errors.New("autonomous shipment workflow not found")
	ErrLoopDetected                 = errors.New("autonomous cycle detected: suppressed recursive self-triggering event")
	ErrActionVerificationFailed     = errors.New("autonomous action execution could not be verified against authoritative data")
	ErrOperationallyInsignificant   = errors.New("ETA variance or event is below operational significance threshold")
	ErrUnresolvedPostDeliveryClaims = errors.New("post-delivery financial or customer claims remain unresolved")
)

// ValidateShipmentStageTransition enforces forward progression of the autonomous lifecycle
func ValidateShipmentStageTransition(current, next ShipmentLifecycleStage) error {
	if current == next {
		return nil
	}

	valid := false
	switch current {
	case StageShipmentCreated:
		valid = (next == StagePlanning || next == StageBooked || next == StageInTransit)
	case StagePlanning:
		valid = (next == StageBooked || next == StageInTransit || next == StageMonitoring)
	case StageBooked:
		valid = (next == StageInTransit || next == StageMonitoring || next == StagePlanning)
	case StageInTransit:
		valid = (next == StageMonitoring || next == StageDelivered || next == StagePlanning)
	case StageMonitoring:
		valid = (next == StageInTransit || next == StageDelivered || next == StagePlanning)
	case StageDelivered:
		valid = (next == StagePostDelivery || next == StageCompleted)
	case StagePostDelivery:
		valid = (next == StageCompleted)
	case StageCompleted:
		valid = false // Terminal state
	default:
		valid = false
	}

	if !valid {
		return ErrInvalidStageTransition
	}
	return nil
}

// AutonomousShipmentWorkflow represents a durable, governed end-to-end shipment workflow
type AutonomousShipmentWorkflow struct {
	WorkflowID                      string                 `json:"workflow_id"`
	ShipmentID                      int64                  `json:"shipment_id"`
	OrgID                           int64                  `json:"org_id"`
	CurrentStage                    ShipmentLifecycleStage `json:"current_stage"`
	WorkflowState                   EnterpriseWorkflowState `json:"workflow_state"`
	AuthoritativeStatus             string                 `json:"authoritative_status"` // Real business status: BOOKED, IN_TRANSIT, DELIVERED, etc.
	CarrierSCAC                     string                 `json:"carrier_scac"`
	OriginPort                      string                 `json:"origin_port"`
	DestinationPort                 string                 `json:"destination_port"`
	ScheduledETA                    *time.Time             `json:"scheduled_eta,omitempty"`
	PredictedETA                    *time.Time             `json:"predicted_eta,omitempty"`
	ActualDeliveryDate              *time.Time             `json:"actual_delivery_date,omitempty"`
	ETAConfidence                   float64                `json:"eta_confidence"`
	PredictedDelayHours             float64                `json:"predicted_delay_hours"`
	IsDelayOperationallySignificant bool                   `json:"is_delay_operationally_significant"`
	ActiveExceptionsCount           int                    `json:"active_exceptions_count"`
	ExceptionSeverity               string                 `json:"exception_severity,omitempty"`
	ExceptionRootCause              string                 `json:"exception_root_cause,omitempty"`
	AssignedSpecialists             []string               `json:"assigned_specialists"`
	CurrentStepID                   string                 `json:"current_step_id,omitempty"`
	Steps                           []EnterpriseWorkflowStep `json:"steps,omitempty"`
	PendingApprovalsCount           int                    `json:"pending_approvals_count"`
	LastExecutedAction              string                 `json:"last_executed_action,omitempty"`
	LastActionVerified              bool                   `json:"last_action_verified"`
	LastActionVerificationMessage   string                 `json:"last_action_verification_message,omitempty"`
	ReplanVersion                   int                    `json:"replan_version"`
	ParentWorkflowID                *string                `json:"parent_workflow_id,omitempty"`
	PostDeliverySummary             map[string]interface{} `json:"post_delivery_summary,omitempty"`
	LearningOutcomesRecorded        bool                   `json:"learning_outcomes_recorded"`
	CorrelationID                   string                 `json:"correlation_id"`
	CreatedAt                       time.Time              `json:"created_at"`
	UpdatedAt                       time.Time              `json:"updated_at"`
}

// ShipmentLifecycleEvent models events triggering or progressing autonomous shipment operations
type ShipmentLifecycleEvent struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"` // e.g. BOOKING_CONFIRMED, CARRIER_ASSIGNED, DEPARTURE, ARRIVAL, MILESTONE_DELAY, ETA_CHANGE, CARRIER_EXCEPTION, DELIVERED
	ShipmentID    int64                  `json:"shipment_id"`
	MilestoneCode string                 `json:"milestone_code,omitempty"`
	Payload       map[string]interface{} `json:"payload"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Source        string                 `json:"source,omitempty"` // "carrier_integration", "iot_sensor", "manual", "enterprise_autonomous_platform"
}

// ActionVerificationResult encapsulates authoritative proof of executed action
type ActionVerificationResult struct {
	ActionID            string                 `json:"action_id"`
	ActionType          string                 `json:"action_type"`
	WorkflowID          string                 `json:"workflow_id"`
	ShipmentID          int64                  `json:"shipment_id"`
	VerifiedSuccess     bool                   `json:"verified_success"`
	AuthoritativeSource string                 `json:"authoritative_source"`
	VerificationMessage string                 `json:"verification_message"`
	AuthoritativeState  map[string]interface{} `json:"authoritative_state,omitempty"`
	VerifiedAt          time.Time              `json:"verified_at"`
}

// OutcomeRecordResult tracks memory & learning feedback
type OutcomeRecordResult struct {
	OutcomeID       string    `json:"outcome_id"`
	WorkflowID      string    `json:"workflow_id"`
	ShipmentID      int64     `json:"shipment_id"`
	MetricName      string    `json:"metric_name"`
	PredictedValue  string    `json:"predicted_value"`
	ActualValue     string    `json:"actual_value"`
	Variance        float64   `json:"variance"`
	LearningApplied bool      `json:"learning_applied"`
	RecordedAt      time.Time `json:"recorded_at"`
}

// Request DTOs
type InitiateShipmentLifecycleRequest struct {
	ShipmentID    int64  `json:"shipment_id"`
	CorrelationID string `json:"correlation_id,omitempty"`
}

type ReplanShipmentLifecycleRequest struct {
	Reason string `json:"reason"`
}

type RecordShipmentOutcomeRequest struct {
	MetricName     string  `json:"metric_name"`
	PredictedValue string  `json:"predicted_value"`
	ActualValue    string  `json:"actual_value"`
	Feedback       string  `json:"feedback,omitempty"`
}
