package predictions

import (
	"encoding/json"
	"time"
)

// PredictionStatus defines the lifecycle states of a prediction
type PredictionStatus string

const (
	StatusGenerated       PredictionStatus = "GENERATED"
	StatusValidated       PredictionStatus = "VALIDATED"
	StatusPublished       PredictionStatus = "PUBLISHED"
	StatusAcknowledged   PredictionStatus = "ACKNOWLEDGED"
	StatusInReview        PredictionStatus = "IN_REVIEW"
	StatusAccepted        PredictionStatus = "ACCEPTED"
	StatusDismissed       PredictionStatus = "DISMISSED"
	StatusActionRequested PredictionStatus = "ACTION_REQUESTED"
	StatusAwaitingApproval PredictionStatus = "AWAITING_APPROVAL"
	StatusActionExecuted  PredictionStatus = "ACTION_EXECUTED"
	StatusExpired         PredictionStatus = "EXPIRED"
	StatusSuperseded      PredictionStatus = "SUPERSEDED"
	StatusFailed          PredictionStatus = "FAILED"
)

// Severity defines risk impact levels
type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

// ConfidenceBand defines prediction certainty brackets
type ConfidenceBand string

const (
	ConfidenceLow    ConfidenceBand = "LOW"
	ConfidenceMedium ConfidenceBand = "MEDIUM"
	ConfidenceHigh   ConfidenceBand = "HIGH"
)

// OutcomeStatus defines post-event evaluation
type OutcomeStatus string

const (
	OutcomePending      OutcomeStatus = "PENDING"
	OutcomeCorrect      OutcomeStatus = "CORRECT"
	OutcomeIncorrect    OutcomeStatus = "INCORRECT"
	OutcomeInconclusive OutcomeStatus = "INCONCLUSIVE"
)

// SourceReference tracks where evidence originated
type SourceReference struct {
	SourceModule         string `json:"source_module"`
	SourceRecordID       string `json:"source_record_id"`
	SourceField          string `json:"source_field"`
	SourceTimestamp      string `json:"source_timestamp"`
	DocumentReference    string `json:"document_reference,omitempty"`
	ClauseReference      string `json:"clause_reference,omitempty"`
	PageNumber           int    `json:"page_number,omitempty"`
	SectionHeading       string `json:"section_heading,omitempty"`
	MilestoneReference   string `json:"milestone_reference,omitempty"`
	CutoffReference      string `json:"cutoff_reference,omitempty"`
	ComparisonPeriod     string `json:"comparison_period,omitempty"`
	SampleSize           int    `json:"sample_size,omitempty"`
	LaneReference        string `json:"lane_reference,omitempty"`
	CarrierReference     string  `json:"carrier_reference,omitempty"`
	CustomerReference    string  `json:"customer_reference,omitempty"`
	WorkloadType         string  `json:"workload_type,omitempty"`
	PendingCount         int     `json:"pending_count,omitempty"`
	CapacityLimit        int     `json:"capacity_limit,omitempty"`
	UtilizationRate      float64 `json:"utilization_rate,omitempty"`
	BottleneckType       string  `json:"bottleneck_type,omitempty"`
	AffectedStage        string  `json:"affected_stage,omitempty"`
	AssignedOwner        string  `json:"assigned_owner,omitempty"`
	QueueDwellHours      float64 `json:"queue_dwell_hours,omitempty"`
	SourceType           string  `json:"source_type,omitempty"`
	SourceURL            string  `json:"source_url,omitempty"`
	PortReference        string  `json:"port_reference,omitempty"`
	DataFreshnessSeconds int     `json:"data_freshness_seconds,omitempty"`
}

// SupportingSignal tracks quantitative drivers
type SupportingSignal struct {
	SignalName       string      `json:"signal_name"`
	ObservedValue    interface{} `json:"observed_value"`
	BaselineValue    interface{} `json:"baseline_value,omitempty"`
	ImportanceWeight float64     `json:"importance_weight"`
}

// Prediction is the persistent business record for Phase 4 forecasts
type Prediction struct {
	ID                   int64            `db:"id" json:"id"`
	OrgID                int64            `db:"org_id" json:"org_id"`
	UserID               *int64           `db:"user_id" json:"user_id,omitempty"`
	PredictionID         string           `db:"prediction_id" json:"prediction_id"`
	IdempotencyKey       string           `db:"idempotency_key" json:"idempotency_key"`
	Module               string           `db:"module" json:"module"`
	PredictionType       string           `db:"prediction_type" json:"prediction_type"`
	Status               PredictionStatus `db:"status" json:"status"`
	Severity             Severity         `db:"severity" json:"severity"`
	ConfidenceScore      float64          `db:"confidence_score" json:"confidence_score"`
	ConfidenceBand       ConfidenceBand   `db:"confidence_band" json:"confidence_band"`
	RelatedRecordType    string           `db:"related_record_type" json:"related_record_type"`
	RelatedRecordID      string           `db:"related_record_id" json:"related_record_id"`
	PredictionStatement  string           `db:"prediction_statement" json:"prediction_statement"`
	PredictedValue       *string          `db:"predicted_value" json:"predicted_value,omitempty"`
	TimeHorizon          *string          `db:"time_horizon" json:"time_horizon,omitempty"`
	TargetDate           *time.Time       `db:"target_date" json:"target_date,omitempty"`
	Explanation          string           `db:"explanation" json:"explanation"`
	SupportingSignalsRaw string           `db:"supporting_signals" json:"-"`
	SupportingSignals    []SupportingSignal `json:"supporting_signals"`
	SourceReferencesRaw  string           `db:"source_references" json:"-"`
	SourceReferences     []SourceReference `json:"source_references"`
	SourceTimestamp      time.Time        `db:"source_timestamp" json:"source_timestamp"`
	RecommendedAction    *string          `db:"recommended_action" json:"recommended_action,omitempty"`
	ActionType           *string          `db:"action_type" json:"action_type,omitempty"`
	IsActionRequired     bool             `db:"is_action_required" json:"is_action_required"`
	RequiresApproval     bool             `db:"requires_approval" json:"requires_approval"`
	ActionProposalID     *string          `db:"action_proposal_id" json:"action_proposal_id,omitempty"`
	ReviewStatus         string           `db:"review_status" json:"review_status"`
	ReviewedBy           *int64           `db:"reviewed_by" json:"reviewed_by,omitempty"`
	ReviewedAt           *time.Time       `db:"reviewed_at" json:"reviewed_at,omitempty"`
	ReviewNotes          *string          `db:"review_notes" json:"review_notes,omitempty"`
	ActualOutcomeStatus  OutcomeStatus    `db:"actual_outcome_status" json:"actual_outcome_status"`
	ActualOutcomeValue   *string          `db:"actual_outcome_value" json:"actual_outcome_value,omitempty"`
	FeedbackNotes        *string          `db:"feedback_notes" json:"feedback_notes,omitempty"`
	ModelVersion         string           `db:"model_version" json:"model_version"`
	ExpiresAt            *time.Time       `db:"expires_at" json:"expires_at,omitempty"`
	DocumentReference    *string          `db:"-" json:"document_reference,omitempty"`
	ClauseReference      *string          `db:"-" json:"clause_reference,omitempty"`
	PageNumber           *int             `db:"-" json:"page_number,omitempty"`
	SectionHeading       *string          `db:"-" json:"section_heading,omitempty"`
	MilestoneReference   *string          `db:"-" json:"milestone_reference,omitempty"`
	CutoffReference      *string          `db:"-" json:"cutoff_reference,omitempty"`
	ComparisonPeriod     *string          `db:"-" json:"comparison_period,omitempty"`
	SampleSize           *int             `db:"-" json:"sample_size,omitempty"`
	LaneReference        *string          `db:"-" json:"lane_reference,omitempty"`
	CarrierReference     *string          `db:"-" json:"carrier_reference,omitempty"`
	CustomerReference    *string          `db:"-" json:"customer_reference,omitempty"`
	WorkloadType         *string          `db:"-" json:"workload_type,omitempty"`
	PendingCount         *int             `db:"-" json:"pending_count,omitempty"`
	CapacityLimit        *int             `db:"-" json:"capacity_limit,omitempty"`
	UtilizationRate      *float64         `db:"-" json:"utilization_rate,omitempty"`
	BottleneckType       *string          `db:"-" json:"bottleneck_type,omitempty"`
	AffectedStage        *string          `db:"-" json:"affected_stage,omitempty"`
	AssignedOwner        *string          `db:"-" json:"assigned_owner,omitempty"`
	QueueDwellHours      *float64         `db:"-" json:"queue_dwell_hours,omitempty"`
	SourceType           *string          `db:"-" json:"source_type,omitempty"`
	SourceURL            *string          `db:"-" json:"source_url,omitempty"`
	PortReference        *string          `db:"-" json:"port_reference,omitempty"`
	LinkedExceptionID    *int64           `db:"-" json:"linked_exception_id,omitempty"`
	InsufficientData     bool             `json:"insufficient_data"`
	InsufficientReason   *string          `json:"insufficient_data_reason,omitempty"`
	CreatedAt            time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time        `db:"updated_at" json:"updated_at"`
}

func (p *Prediction) UnpackJSON() {
	if p.SupportingSignalsRaw != "" {
		_ = json.Unmarshal([]byte(p.SupportingSignalsRaw), &p.SupportingSignals)
	}
	if p.SourceReferencesRaw != "" {
		_ = json.Unmarshal([]byte(p.SourceReferencesRaw), &p.SourceReferences)
	}
	for _, src := range p.SourceReferences {
		if src.DocumentReference != "" && p.DocumentReference == nil {
			doc := src.DocumentReference
			p.DocumentReference = &doc
		}
		if src.ClauseReference != "" && p.ClauseReference == nil {
			cl := src.ClauseReference
			p.ClauseReference = &cl
		}
		if src.PageNumber > 0 && p.PageNumber == nil {
			pg := src.PageNumber
			p.PageNumber = &pg
		}
		if src.SectionHeading != "" && p.SectionHeading == nil {
			sec := src.SectionHeading
			p.SectionHeading = &sec
		}
		if src.MilestoneReference != "" && p.MilestoneReference == nil {
			m := src.MilestoneReference
			p.MilestoneReference = &m
		}
		if src.CutoffReference != "" && p.CutoffReference == nil {
			c := src.CutoffReference
			p.CutoffReference = &c
		}
		if src.ComparisonPeriod != "" && p.ComparisonPeriod == nil {
			cp := src.ComparisonPeriod
			p.ComparisonPeriod = &cp
		}
		if src.SampleSize > 0 && p.SampleSize == nil {
			ss := src.SampleSize
			p.SampleSize = &ss
		}
		if src.LaneReference != "" && p.LaneReference == nil {
			lr := src.LaneReference
			p.LaneReference = &lr
		}
		if src.CarrierReference != "" && p.CarrierReference == nil {
			cr := src.CarrierReference
			p.CarrierReference = &cr
		}
		if src.CustomerReference != "" && p.CustomerReference == nil {
			custr := src.CustomerReference
			p.CustomerReference = &custr
		}
		if src.WorkloadType != "" && p.WorkloadType == nil {
			wt := src.WorkloadType
			p.WorkloadType = &wt
		}
		if src.PendingCount > 0 && p.PendingCount == nil {
			pc := src.PendingCount
			p.PendingCount = &pc
		}
		if src.CapacityLimit > 0 && p.CapacityLimit == nil {
			cl := src.CapacityLimit
			p.CapacityLimit = &cl
		}
		if src.UtilizationRate > 0 && p.UtilizationRate == nil {
			ur := src.UtilizationRate
			p.UtilizationRate = &ur
		}
		if src.BottleneckType != "" && p.BottleneckType == nil {
			bt := src.BottleneckType
			p.BottleneckType = &bt
		}
		if src.AffectedStage != "" && p.AffectedStage == nil {
			as := src.AffectedStage
			p.AffectedStage = &as
		}
		if src.AssignedOwner != "" && p.AssignedOwner == nil {
			ao := src.AssignedOwner
			p.AssignedOwner = &ao
		}
		if src.QueueDwellHours > 0 && p.QueueDwellHours == nil {
			dh := src.QueueDwellHours
			p.QueueDwellHours = &dh
		}
		if src.SourceType != "" && p.SourceType == nil {
			st := src.SourceType
			p.SourceType = &st
		}
		if src.SourceURL != "" && p.SourceURL == nil {
			su := src.SourceURL
			p.SourceURL = &su
		}
		if src.PortReference != "" && p.PortReference == nil {
			pr := src.PortReference
			p.PortReference = &pr
		}
	}
	if p.ReviewStatus == "INSUFFICIENT_DATA" || (p.PredictedValue != nil && (*p.PredictedValue == "INSUFFICIENT_DATA" || *p.PredictedValue == "UNPRICED")) {
		p.InsufficientData = true
	}
}

// PredictionAuditHistory tracks status changes
type PredictionAuditHistory struct {
	ID             int64     `db:"id" json:"id"`
	PredictionID   string    `db:"prediction_id" json:"prediction_id"`
	OrgID          int64     `db:"org_id" json:"org_id"`
	UserID         *int64    `db:"user_id" json:"user_id,omitempty"`
	PreviousStatus *string   `db:"previous_status" json:"previous_status,omitempty"`
	NewStatus      string    `db:"new_status" json:"new_status"`
	Action         string    `db:"action" json:"action"`
	Notes          *string   `db:"notes" json:"notes,omitempty"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

// FilterParams for prediction queries
type FilterParams struct {
	Module            string
	Severity          string
	ConfidenceBand    string
	Status            string
	RelatedRecordType string
	RelatedRecordID   string
	Limit             int
	Offset            int
}

// GenerateRequest initiates a prediction from Go to Python
type GenerateRequest struct {
	OrgID             int64                  `json:"org_id"`
	Module            string                 `json:"module"`
	PredictionType    string                 `json:"prediction_type"`
	RelatedRecordType string                 `json:"related_record_type"`
	RelatedRecordID   string                 `json:"related_record_id"`
	RecordContext     map[string]interface{} `json:"record_context"`
	TimeHorizon       string                 `json:"time_horizon,omitempty"`
	CorrelationID     string                 `json:"correlation_id,omitempty"`
}

// SidecarPredictionResponse matches the Python Pydantic contract
type SidecarPredictionResponse struct {
	PredictionID        string             `json:"prediction_id"`
	OrgID               int64              `json:"org_id"`
	Module              string             `json:"module"`
	PredictionType      string             `json:"prediction_type"`
	RelatedRecordType   string             `json:"related_record_type"`
	RelatedRecordID     string             `json:"related_record_id"`
	PredictionStatement string             `json:"prediction_statement"`
	PredictedValue      *string            `json:"predicted_value,omitempty"`
	PredictionCategory  *string            `json:"prediction_category,omitempty"`
	DisruptionCategory  *string            `json:"disruption_category,omitempty"`
	LinkedExceptionID   *int64             `json:"linked_exception_id,omitempty"`
	TimeHorizon         *string            `json:"time_horizon,omitempty"`
	TargetDate          *string            `json:"target_date,omitempty"`
	Severity            string             `json:"severity"`
	ConfidenceScore     float64            `json:"confidence_score"`
	ConfidenceBand      string             `json:"confidence_band"`
	Explanation         string             `json:"explanation"`
	SupportingSignals   []SupportingSignal `json:"supporting_signals"`
	SourceReferences    []SourceReference  `json:"source_references"`
	SourceTimestamp     string             `json:"source_timestamp"`
	RecommendedAction   *string            `json:"recommended_action,omitempty"`
	ActionType          *string            `json:"action_type,omitempty"`
	IsActionRequired    bool               `json:"is_action_required"`
	RequiresApproval    bool               `json:"requires_approval"`
	DocumentReference   *string            `json:"document_reference,omitempty"`
	ClauseReference     *string            `json:"clause_reference,omitempty"`
	PageNumber          *int               `json:"page_number,omitempty"`
	SectionHeading      *string            `json:"section_heading,omitempty"`
	MilestoneReference  *string            `json:"milestone_reference,omitempty"`
	CutoffReference     *string            `json:"cutoff_reference,omitempty"`
	ComparisonPeriod    *string            `json:"comparison_period,omitempty"`
	SampleSize          *int               `json:"sample_size,omitempty"`
	LaneReference       *string            `json:"lane_reference,omitempty"`
	CarrierReference    *string            `json:"carrier_reference,omitempty"`
	CustomerReference   *string            `json:"customer_reference,omitempty"`
	WorkloadType        *string            `json:"workload_type,omitempty"`
	PendingCount        *int               `json:"pending_count,omitempty"`
	CapacityLimit       *int               `json:"capacity_limit,omitempty"`
	UtilizationRate     *float64           `json:"utilization_rate,omitempty"`
	BottleneckType      *string            `json:"bottleneck_type,omitempty"`
	AffectedStage       *string            `json:"affected_stage,omitempty"`
	AssignedOwner       *string            `json:"assigned_owner,omitempty"`
	QueueDwellHours     *float64           `json:"queue_dwell_hours,omitempty"`
	SourceType          *string            `json:"source_type,omitempty"`
	SourceURL           *string            `json:"source_url,omitempty"`
	PortReference       *string            `json:"port_reference,omitempty"`
	InsufficientData    bool               `json:"insufficient_data"`
	InsufficientReason  *string            `json:"insufficient_data_reason,omitempty"`
	ModelVersion        string             `json:"model_version"`
	CreatedAt           string             `json:"created_at"`
}
