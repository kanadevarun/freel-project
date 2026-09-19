package bcontext

import "time"

// ShipmentIdentitySummary encapsulates identity, routing, and lifecycle metadata for a shipment.
type ShipmentIdentitySummary struct {
	ShipmentID       int64      `json:"shipment_id"`
	OrgID            int64      `json:"org_id"`
	ShipmentNumber   string     `json:"shipment_number"`
	CustomerID       *int64     `json:"customer_id,omitempty"`
	CustomerName     string     `json:"customer_name,omitempty"`
	BookingID        *int64     `json:"booking_id,omitempty"`
	BookingNumber    string     `json:"booking_number,omitempty"`
	RFQID            *int64     `json:"rfq_id,omitempty"`
	RFQNumber        string     `json:"rfq_number,omitempty"`
	QuoteID          *int64     `json:"quote_id,omitempty"`
	OriginPort       string     `json:"origin_port"`
	DestinationPort  string     `json:"destination_port"`
	TransportMode    string     `json:"transport_mode"` // OCEAN, AIR, ROAD, RAIL
	CarrierSCAC      string     `json:"carrier_scac"`
	CarrierName      string     `json:"carrier_name"`
	VesselName       string     `json:"vessel_name,omitempty"`
	VoyageNumber     string     `json:"voyage_number,omitempty"`
	Status           string     `json:"status"`
	ClosureStatus    string     `json:"closure_status"`
	MBLNumber        string     `json:"mbl_number,omitempty"`
	HBLNumber        string     `json:"hbl_number,omitempty"`
	ContainerNumbers []string   `json:"container_numbers,omitempty"`
	ContainerCount   int        `json:"container_count"`
	ETD              *time.Time `json:"etd,omitempty"`
	ETA              *time.Time `json:"eta,omitempty"`
	ActualDeparture  *time.Time `json:"actual_departure,omitempty"`
	ActualArrival    *time.Time `json:"actual_arrival,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// ShipmentMilestoneItem represents a normalized milestone record.
type ShipmentMilestoneItem struct {
	ID            int64      `json:"id"`
	MilestoneCode string     `json:"milestone_code"`
	Description   string     `json:"description,omitempty"`
	PlannedDate   *time.Time `json:"planned_date,omitempty"`
	ActualDate    *time.Time `json:"actual_date,omitempty"`
	Status        string     `json:"status"` // PLANNED, COMPLETED
	Location      string     `json:"location,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	IsDelayed     bool       `json:"is_delayed"`
	DelayHours    float64    `json:"delay_hours,omitempty"`
}

// ShipmentMilestoneIntelligence captures milestone progress, timeliness, and tracking freshness.
type ShipmentMilestoneIntelligence struct {
	TotalMilestones             int                      `json:"total_milestones"`
	CompletedMilestones         int                      `json:"completed_milestones"`
	PendingMilestones           int                      `json:"pending_milestones"`
	OverdueMilestones           int                      `json:"overdue_milestones"`
	UpcomingMilestones          int                      `json:"upcoming_milestones"`
	MilestoneCompletionRate     float64                  `json:"milestone_completion_rate"`
	LastCompletedMilestoneCode  string                   `json:"last_completed_milestone_code,omitempty"`
	LastCompletedMilestoneDate  *time.Time               `json:"last_completed_milestone_date,omitempty"`
	NextExpectedMilestoneCode   string                   `json:"next_expected_milestone_code,omitempty"`
	NextExpectedMilestoneDate   *time.Time               `json:"next_expected_milestone_date,omitempty"`
	TimeSinceLastUpdateHours    *float64                 `json:"time_since_last_update_hours,omitempty"`
	NumberOfDelayedMilestones   int                      `json:"number_of_delayed_milestones"`
	MissingDateMilestonesCount  int                      `json:"missing_date_milestones_count"`
	StaleTrackingFlag           bool                     `json:"stale_tracking_flag"`
	CriticalMilestoneRisks      []string                 `json:"critical_milestone_risks"`
	MilestonesList              []*ShipmentMilestoneItem `json:"milestones_list"`
}

// ShipmentExceptionItem represents a detailed exception record.
type ShipmentExceptionItem struct {
	ID              int64      `json:"id"`
	ExceptionType   string     `json:"exception_type"`
	Severity        string     `json:"severity"` // LOW, MEDIUM, HIGH, CRITICAL
	Status          string     `json:"status"`   // OPEN, ACKNOWLEDGED, RESOLVED
	Title           string     `json:"title"`
	Description     string     `json:"description,omitempty"`
	Resolved        bool       `json:"resolved"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	ResolutionNotes string     `json:"resolution_notes,omitempty"`
	HoursOpen       float64    `json:"hours_open"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ShipmentExceptionIntelligence provides operational exception tracking and prioritization.
type ShipmentExceptionIntelligence struct {
	TotalExceptions            int                      `json:"total_exceptions"`
	OpenExceptions             int                      `json:"open_exceptions"`
	ResolvedExceptions         int                      `json:"resolved_exceptions"`
	CriticalExceptions         int                      `json:"critical_exceptions"`
	HighSeverityExceptions     int                      `json:"high_severity_exceptions"`
	OldestUnresolvedTitle      string                   `json:"oldest_unresolved_title,omitempty"`
	OldestUnresolvedHoursOpen  *float64                 `json:"oldest_unresolved_hours_open,omitempty"`
	ExceptionsByCategory       map[string]int           `json:"exceptions_by_category"`
	ExceptionsList             []*ShipmentExceptionItem `json:"exceptions_list"`
}

// ShipmentOperationalPerformance represents schedule adherence and transit cycle analytics.
type ShipmentOperationalPerformance struct {
	OnTimeDeparture         *bool    `json:"on_time_departure,omitempty"`
	DepartureVarianceHours  *float64 `json:"departure_variance_hours,omitempty"`
	OnTimeArrival           *bool    `json:"on_time_arrival,omitempty"`
	ArrivalVarianceHours    *float64 `json:"arrival_variance_hours,omitempty"`
	PlannedTransitDays      *float64 `json:"planned_transit_days,omitempty"`
	ActualTransitDays       *float64 `json:"actual_transit_days,omitempty"`
	DelayDurationHours      *float64 `json:"delay_duration_hours,omitempty"`
	NumberOfDelays          int      `json:"number_of_delays"`
	CycleTimeDays           *float64 `json:"cycle_time_days,omitempty"`
	CarrierUpdateFreshness  string   `json:"carrier_update_freshness"` // LIVE, RECENT, STALE, UNAVAILABLE
}

// ShipmentRiskIndicators provides proactive operational flags and health assessment.
type ShipmentRiskIndicators struct {
	OverallRiskRating         string   `json:"overall_risk_rating"` // LOW, MODERATE, HIGH, CRITICAL
	RiskScore                 int      `json:"risk_score"`          // 0 to 100
	ShipmentDelayed           bool     `json:"shipment_delayed"`
	ETAMissingOrStale         bool     `json:"eta_missing_or_stale"`
	CriticalMilestoneOverdue  bool     `json:"critical_milestone_overdue"`
	UnresolvedHighException   bool     `json:"unresolved_high_exception"`
	StaleTrackingData         bool     `json:"stale_tracking_data"`
	ConflictingData           bool     `json:"conflicting_data"`
	MissingCarrierInfo        bool     `json:"missing_carrier_info"`
	MissingScheduleData       bool     `json:"missing_schedule_data"`
	DataIncomplete            bool     `json:"data_incomplete"`
	ActiveRiskFactors         []string `json:"active_risk_factors"`
	MissingDataReasons        []string `json:"missing_data_reasons"`
}

// ShipmentAIOperationsSummary provides grounded, read-only operational narrative and actions.
type ShipmentAIOperationsSummary struct {
	ExecutiveSummary            string   `json:"executive_summary"`
	CurrentStatusExplanation    string   `json:"current_status_explanation"`
	MilestoneProgressExplanation string  `json:"milestone_progress_explanation"`
	LikelyOperationalRisks      string   `json:"likely_operational_risks"`
	DelayCausesEvidence         string   `json:"delay_causes_evidence,omitempty"`
	ExceptionPrioritization     string   `json:"exception_prioritization,omitempty"`
	ActionableAttentionItems    []string `json:"actionable_attention_items"`
	SuggestedOperatorInquiries  []string `json:"suggested_operator_inquiries"`
	ConfidenceScore             string   `json:"confidence_score"` // HIGH, MEDIUM, LOW
	VerifiableSourceCitations   []string `json:"verifiable_source_citations"`
}

// ShipmentObservation represents a grounded, traceable observation for auditability.
type ShipmentObservation struct {
	Category string `json:"category"` // STATUS, MILESTONE, EXCEPTION, SCHEDULE, RISK
	Severity string `json:"severity"` // INFO, WARNING, CRITICAL
	Message  string `json:"message"`
	Evidence string `json:"evidence"`
}

// Shipment360OperationsIntelligence is the comprehensive read-only operational intelligence view.
type Shipment360OperationsIntelligence struct {
	ShipmentID             int64                            `json:"shipment_id"`
	OrgID                  int64                            `json:"org_id"`
	CorrelationID          string                           `json:"correlation_id"`
	DataFreshnessTimestamp time.Time                        `json:"data_freshness_timestamp"`
	Identity               *ShipmentIdentitySummary         `json:"identity"`
	Milestones             *ShipmentMilestoneIntelligence   `json:"milestones"`
	Exceptions             *ShipmentExceptionIntelligence   `json:"exceptions"`
	Performance            *ShipmentOperationalPerformance  `json:"performance"`
	RiskIndicators         *ShipmentRiskIndicators          `json:"risk_indicators"`
	AISummary              *ShipmentAIOperationsSummary     `json:"ai_summary"`
	GroundedObservations   []*ShipmentObservation           `json:"grounded_observations"`
}

// OrgOperationsSummary represents an aggregate operational health snapshot across an organization.
type OrgOperationsSummary struct {
	OrgID                  int64          `json:"org_id"`
	CorrelationID          string         `json:"correlation_id"`
	DataFreshnessTimestamp time.Time      `json:"data_freshness_timestamp"`
	TotalShipments         int            `json:"total_shipments"`
	ActiveShipments        int            `json:"active_shipments"`
	CompletedShipments     int            `json:"completed_shipments"`
	DelayedShipments       int            `json:"delayed_shipments"`
	TotalOpenExceptions    int            `json:"total_open_exceptions"`
	CriticalOpenExceptions int            `json:"critical_open_exceptions"`
	StaleTrackingShipments int            `json:"stale_tracking_shipments"`
	OnTimeRatePercentage   *float64       `json:"on_time_rate_percentage,omitempty"`
	RiskDistribution       map[string]int `json:"risk_distribution"` // LOW, MODERATE, HIGH, CRITICAL
}
