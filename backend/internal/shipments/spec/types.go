package spec

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type JSONStringSlice []string

func (j *JSONStringSlice) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		if str, ok := value.(string); ok {
			bytes = []byte(str)
		} else {
			return fmt.Errorf("failed to unmarshal JSONStringSlice: %v", value)
		}
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONStringSlice) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

type Shipment struct {
	ID               int64           `db:"id" json:"id"`
	OrgID            int64           `db:"org_id" json:"org_id"`
	RFQID            *int64          `db:"rfq_id" json:"rfq_id,omitempty"`
	QuoteID          *int64          `db:"quote_id" json:"quote_id,omitempty"`
	BookingID        *int64          `db:"booking_id" json:"booking_id,omitempty"`
	CarrierSCAC      string          `db:"carrier_scac" json:"carrier_scac"`
	BookingNumber    *string         `db:"booking_number" json:"booking_number,omitempty"`
	MBLNumber        *string         `db:"mbl_number" json:"mbl_number,omitempty"`
	HBLNumber        *string         `db:"hbl_number" json:"hbl_number,omitempty"`
	ContainerNumbers JSONStringSlice `db:"container_numbers" json:"container_numbers"`
	Status           string          `db:"status" json:"status"`
	OriginPort       string          `db:"origin_port" json:"origin_port"`
	DestinationPort  string          `db:"destination_port" json:"destination_port"`
	VesselName       *string         `db:"vessel_name" json:"vessel_name,omitempty"`
	VoyageNumber     *string         `db:"voyage_number" json:"voyage_number,omitempty"`
	ETD              *time.Time      `db:"etd" json:"etd,omitempty"`
	ETA              *time.Time      `db:"eta" json:"eta,omitempty"`
	SourceQuotationID *string        `db:"source_quotation_id" json:"source_quotation_id,omitempty"`
	SourceBookingID   *string        `db:"source_booking_id" json:"source_booking_id,omitempty"`
	CustomerID       *int64          `db:"customer_id" json:"customer_id,omitempty"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at" json:"updated_at"`

	// Joined Metadata Fields
	RFQNumber    *string `db:"rfq_number" json:"rfq_number,omitempty"`
	CustomerName *string `db:"customer_name" json:"customer_name,omitempty"`
	CarrierName  *string `db:"carrier_name" json:"carrier_name,omitempty"`
	ActiveExceptionsCount int64 `db:"active_exceptions_count" json:"active_exceptions_count"`
	HighExceptionsCount   int64 `db:"high_exceptions_count" json:"high_exceptions_count"`
	ClosureStatus         string `db:"closure_status" json:"closure_status"`
}

type ShipmentKPIs struct {
	TotalShipments int64 `db:"total_shipments" json:"total_shipments"`
	BookingPending int64 `db:"booking_pending" json:"booking_pending"`
	Booked         int64 `db:"booked" json:"booked"`
	InTransit      int64 `db:"in_transit" json:"in_transit"`
	Arrived        int64 `db:"arrived" json:"arrived"`
	Delivered      int64 `db:"delivered" json:"delivered"`
	Exceptions     int64 `db:"exceptions" json:"exceptions"`
}

type ShipmentListFilter struct {
	Page   int     `json:"page"`
	Limit  int     `json:"limit"`
	Status *string `json:"status,omitempty"`
	Search *string `json:"search,omitempty"`
}

type ShipmentMilestone struct {
	ID            int64      `db:"id" json:"id"`
	ShipmentID    int64      `db:"shipment_id" json:"shipment_id"`
	MilestoneCode string     `db:"milestone_code" json:"milestone_code"` // BOOKED, DEPARTED, IN_TRANSIT, ARRIVED, DELIVERED
	Description   *string    `db:"description" json:"description,omitempty"`
	PlannedDate   *time.Time `db:"planned_date" json:"planned_date,omitempty"`
	ActualDate    *time.Time `db:"actual_date" json:"actual_date,omitempty"`
	Status        string     `db:"status" json:"status"` // PLANNED, COMPLETED
	Location      *string    `db:"location" json:"location,omitempty"`
	Notes         *string    `db:"notes" json:"notes,omitempty"`
	SourceEventID *string    `db:"source_event_id" json:"source_event_id,omitempty"`
}

type ShipmentException struct {
	ID              int64      `db:"id" json:"id"`
	OrgID           int64      `db:"org_id" json:"org_id"`
	ShipmentID      int64      `db:"shipment_id" json:"shipment_id"`
	ExceptionType   string     `db:"exception_type" json:"exception_type"`
	Severity        string     `db:"severity" json:"severity"`
	Status          string     `db:"status" json:"status"`
	Title           string     `db:"title" json:"title"`
	Description     *string    `db:"description" json:"description,omitempty"`
	Resolved        bool       `db:"resolved" json:"resolved"`
	ResolvedAt      *time.Time `db:"resolved_at" json:"resolved_at,omitempty"`
	ResolvedBy      *int64     `db:"resolved_by" json:"resolved_by,omitempty"`
	ResolutionNotes *string    `db:"resolution_notes" json:"resolution_notes,omitempty"`
	AISummary       *string    `db:"ai_summary" json:"ai_summary,omitempty"`
	SourceEventID   *string    `db:"source_event_id" json:"source_event_id,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// NormalizedTrackingEvent represents a standardized tracking event used across Go and Python
type NormalizedTrackingEvent struct {
	EventID         string          `json:"event_id"`
	SourceType      string          `json:"source_type"` // API | WEBHOOK | EMAIL | MANUAL | POLLING
	CarrierSCAC     string          `json:"carrier_scac"`
	BookingNumber   string          `json:"booking_number"`
	ContainerNumber string          `json:"container_number"`
	MBLNumber       string          `json:"mbl_number"`
	HBLNumber       string          `json:"hbl_number"`
	VesselName      string          `json:"vessel_name"`
	VoyageNumber    string          `json:"voyage_number"`
	MilestoneCode   string          `json:"milestone_code"`
	EventTime       time.Time       `json:"event_time"`
	Location        string          `json:"location"`
	Description     string          `json:"description"`
	RawPayload      json.RawMessage `json:"raw_payload,omitempty"`
	ReceivedAt      time.Time       `json:"received_at"`
}

// CarrierTrackingEvent DB representation of raw carrier updates
type CarrierTrackingEvent struct {
	ID                int64           `db:"id" json:"id"`
	OrgID             int64           `db:"org_id" json:"org_id"`
	EventID           string          `db:"event_id" json:"event_id"`
	SourceType        string          `db:"source_type" json:"source_type"`
	CarrierSCAC       string          `db:"carrier_scac" json:"carrier_scac"`
	BookingNumber     string          `db:"booking_number" json:"booking_number"`
	ContainerNumber   string          `db:"container_number" json:"container_number"`
	MBLNumber         string          `db:"mbl_number" json:"mbl_number"`
	HBLNumber         string          `db:"hbl_number" json:"hbl_number"`
	VesselName        string          `db:"vessel_name" json:"vessel_name"`
	VoyageNumber      string          `db:"voyage_number" json:"voyage_number"`
	MilestoneCode     string          `db:"milestone_code" json:"milestone_code"`
	EventTime         time.Time       `db:"event_time" json:"event_time"`
	Location          string          `db:"location" json:"location"`
	RawDescription    string          `db:"raw_description" json:"raw_description"`
	RawPayload        json.RawMessage `db:"raw_payload" json:"raw_payload"`
	ShipmentID        *int64          `db:"shipment_id" json:"shipment_id,omitempty"`
	MatchingStatus    string          `db:"matching_status" json:"matching_status"`
	ProcessingStatus  string          `db:"processing_status" json:"processing_status"`
	ReceivedAt        time.Time       `db:"received_at" json:"received_at"`
	UpdatedAt         time.Time       `db:"updated_at" json:"updated_at"`
}

type ShipmentTrackingSummary struct {
	ShipmentStatus            string     `json:"shipment_status"`
	TrackingState             string     `json:"tracking_state"`
	ProgressPercentage        int        `json:"progress_percentage"`
	PlannedETD                *time.Time `json:"planned_etd,omitempty"`
	ActualETD                 *time.Time `json:"actual_etd,omitempty"`
	PlannedETA                *time.Time `json:"planned_eta,omitempty"`
	ActualArrival             *time.Time `json:"actual_arrival,omitempty"`
	ScheduleVariance          *float64   `json:"schedule_variance,omitempty"` // in days
	ScheduleVarianceState     string     `json:"schedule_variance_state"`
	ActiveExceptionsCount     int64      `json:"active_exceptions_count"`
	HighestCompletedMilestone string     `json:"highest_completed_milestone"`
	ClosureStatus             string     `json:"closure_status"`
}

// ShipmentDocument DB representation for Task 16.9
type ShipmentDocument struct {
	ID              int64           `db:"id" json:"id"`
	OrgID           int64           `db:"org_id" json:"org_id"`
	ShipmentID      int64           `db:"shipment_id" json:"shipment_id"`
	DocType         string          `db:"doc_type" json:"doc_type"`
	DocumentName    *string         `db:"document_name" json:"document_name"`
	Category        string          `db:"category" json:"category"`
	Description     *string         `db:"description" json:"description"`
	S3Key           *string         `db:"s3_key" json:"s3_key,omitempty"`
	FileName        string          `db:"file_name" json:"file_name"`
	FileURL         *string         `db:"file_url" json:"file_url,omitempty"`
	FileSize        *int64          `db:"file_size" json:"file_size,omitempty"`
	MimeType        *string         `db:"mime_type" json:"mime_type,omitempty"`
	FileType        *string         `db:"file_type" json:"file_type,omitempty"`
	Status          string          `db:"status" json:"status"`
	UploadedBy      *string         `db:"uploaded_by" json:"uploaded_by,omitempty"`
	UploadedAt      *time.Time      `db:"uploaded_at" json:"uploaded_at,omitempty"`
	ReviewedBy      *string         `db:"reviewed_by" json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time      `db:"reviewed_at" json:"reviewed_at,omitempty"`
	RejectionReason *string         `db:"rejection_reason" json:"rejection_reason,omitempty"`
	ExpiresAt       *time.Time      `db:"expires_at" json:"expires_at,omitempty"`
	DocumentDate    *time.Time      `db:"document_date" json:"document_date,omitempty"`
	ReferenceNumber *string         `db:"reference_number" json:"reference_number,omitempty"`
	Source          string          `db:"source" json:"source"`
	SourceID        *int64          `db:"source_id" json:"source_id,omitempty"`
	ExtractedData   json.RawMessage `db:"extracted_data" json:"extracted_data,omitempty"`
	RawOcrText      *string         `db:"raw_ocr_text" json:"raw_ocr_text,omitempty"`
	AISummary       *string         `db:"ai_summary" json:"ai_summary,omitempty"`
	ValidityStatus  string          `json:"validity_status,omitempty"` // VALID, EXPIRING_SOON, EXPIRED
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
}

// MissingDocumentRequirement captures missing or required operational document rules (Task 16.9)
type MissingDocumentRequirement struct {
	DocType          string `json:"doc_type"`
	Category         string `json:"category"`
	Name             string `json:"name"`
	RequirementLevel string `json:"requirement_level"` // CRITICAL, REQUIRED, OPTIONAL
	Reason           string `json:"reason"`
	IsMissing        bool   `json:"is_missing"`
	Status           string `json:"status"`            // MISSING, REJECTED, EXPIRED, UPLOADED, APPROVED
}

// ShipmentDocumentComplianceSummary represents authoritative compliance evaluation results (Task 16.9)
type ShipmentDocumentComplianceSummary struct {
	TotalRequired    int                           `json:"total_required"`
	Available        int                           `json:"available"`
	Missing          int                           `json:"missing"`
	UnderReview      int                           `json:"under_review"`
	Approved         int                           `json:"approved"`
	Rejected         int                           `json:"rejected"`
	Expired          int                           `json:"expired"`
	ExpiringSoon     int                           `json:"expiring_soon"`
	ComplianceState  string                        `json:"compliance_state"` // COMPLIANT, ATTENTION_REQUIRED, AT_RISK, NON_COMPLIANT, BLOCKED
	BlockerReason    string                        `json:"blocker_reason,omitempty"`
	MissingDocuments []*MissingDocumentRequirement `json:"missing_documents"`
}

// ShipmentDocumentDiscrepancy for automated OCR cross-document checks
type ShipmentDocumentDiscrepancy struct {
	ID             int64      `db:"id" json:"id"`
	ShipmentID     int64      `db:"shipment_id" json:"shipment_id"`
	FieldName      string     `db:"field_name" json:"field_name"`
	SourceDocument string     `db:"source_document" json:"source_document"`
	TargetDocument string     `db:"target_document" json:"target_document"`
	SourceValue    *string    `db:"source_value" json:"source_value,omitempty"`
	TargetValue    *string    `db:"target_value" json:"target_value,omitempty"`
	Severity       string     `db:"severity" json:"severity"`
	Resolved       bool       `db:"resolved" json:"resolved"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
}

// ShipmentFinancialSummary represents aggregated financial performance and health (Task 16.8)
type ShipmentFinancialSummary struct {
	ShipmentID             int64    `json:"shipment_id"`
	OrgID                  int64    `json:"org_id"`
	Currency               string   `json:"currency"`
	EstimatedRevenue       float64  `json:"estimated_revenue"`
	ActualRevenue          float64  `json:"actual_revenue"`
	EstimatedCost          float64  `json:"estimated_cost"`
	ActualCost             float64  `json:"actual_cost"`
	EstimatedMargin        float64  `json:"estimated_margin"`
	ActualMargin           float64  `json:"actual_margin"`
	EstimatedMarginPercent float64  `json:"estimated_margin_percent"`
	ActualMarginPercent    float64  `json:"actual_margin_percent"`
	VarianceAmount         float64  `json:"variance_amount"`
	VariancePercent        float64  `json:"variance_percent"`
	FinancialStatus        string   `json:"financial_status"` // ESTIMATED, IN_PROGRESS, PENDING_REVIEW, PROFITABLE, LOW_MARGIN, LOSS, FINANCIALLY_CLOSED
	TotalChargesCount      int      `json:"total_charges_count"`
	PendingChargesCount    int      `json:"pending_charges_count"`
	InvoicedCarrierCount   int      `json:"invoiced_carrier_count"`
	InvoicedCustomerCount  int      `json:"invoiced_customer_count"`
	RFQID                  *int64   `json:"rfq_id,omitempty"`
	QuoteID                *int64   `json:"quote_id,omitempty"`
	BookingID              *int64   `json:"booking_id,omitempty"`
}

// ShipmentFinancialCharge represents operational financial line item costs/revenues (Task 16.8)
type ShipmentFinancialCharge struct {
	ID              int64      `db:"id" json:"id"`
	OrgID           int64      `db:"org_id" json:"org_id"`
	ShipmentID      int64      `db:"shipment_id" json:"shipment_id"`
	BookingID       *int64     `db:"booking_id" json:"booking_id,omitempty"`
	RFQID           *int64     `db:"rfq_id" json:"rfq_id,omitempty"`
	Category        string     `db:"category" json:"category"`
	ChargeType      string     `db:"charge_type" json:"charge_type"` // COST or REVENUE
	Description     string     `db:"description" json:"description"`
	VendorName      *string    `db:"vendor_name" json:"vendor_name,omitempty"`
	EstimatedAmount float64    `db:"estimated_amount" json:"estimated_amount"`
	ActualAmount    float64    `db:"actual_amount" json:"actual_amount"`
	Currency        string     `db:"currency" json:"currency"`
	ReferenceNumber *string    `db:"reference_number" json:"reference_number,omitempty"`
	ChargeDate      *time.Time `db:"charge_date" json:"charge_date,omitempty"`
	Status          string     `db:"status" json:"status"` // ESTIMATED, INVOICED, APPROVED, DISPUTED, PAID
	Notes           *string    `db:"notes" json:"notes,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// TrackingPosition represents authoritative geographic coordinates and vessel telemetry (Task 17.3)
type TrackingPosition struct {
	ID             int64     `db:"id" json:"id"`
	OrgID          int64     `db:"org_id" json:"org_id"`
	ShipmentID     int64     `db:"shipment_id" json:"shipment_id"`
	VesselName     *string   `db:"vessel_name" json:"vessel_name,omitempty"`
	Latitude       float64   `db:"latitude" json:"latitude"`
	Longitude      float64   `db:"longitude" json:"longitude"`
	SpeedKnots     float64   `db:"speed_knots" json:"speed_knots"`
	HeadingDegrees float64   `db:"heading_degrees" json:"heading_degrees"`
	LocationName   string    `db:"location_name" json:"location_name"`
	TrackingSource string    `db:"tracking_source" json:"tracking_source"`
	DataFreshness  string    `db:"data_freshness" json:"data_freshness"` // LIVE, RECENT, STALE, UNAVAILABLE
	RecordedAt     time.Time `db:"recorded_at" json:"recorded_at"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// TrackingCoordinates represents basic latitude/longitude point
type TrackingCoordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name,omitempty"`
	Code      string  `json:"code,omitempty"`
}

// TrackingWaypoint represents a navigational corridor point
type TrackingWaypoint struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Sequence  int     `json:"sequence"`
	Passed    bool    `json:"passed"`
	PassedAt  *string `json:"passed_at,omitempty"`
}

// TrackingRoute represents planned oceanic route geometry and corridor coordinates (Task 17.3)
type TrackingRoute struct {
	Origin                 string                `json:"origin"`
	Destination            string                `json:"destination"`
	OriginCoordinates      TrackingCoordinates   `json:"origin_coordinates"`
	DestinationCoordinates TrackingCoordinates   `json:"destination_coordinates"`
	Waypoints              []TrackingWaypoint    `json:"waypoints"`
	PlannedDistanceNM      float64               `json:"planned_distance_nm"`
	DistanceRemainingNM    float64               `json:"distance_remaining_nm"`
	TransitDurationDays    float64               `json:"transit_duration_days"`
	RoutePolyline          []TrackingCoordinates `json:"route_polyline"`
}

// TrackingEventNormalized represents normalized chronological operational event (Task 17.3)
type TrackingEventNormalized struct {
	ID            int64      `json:"id"`
	EventID       string     `json:"event_id"`
	MilestoneCode string     `json:"milestone_code"`
	Category      string     `json:"category"` // CARRIER, AIS, TERMINAL, SYSTEM
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Location      string     `json:"location"`
	Latitude      *float64   `json:"latitude,omitempty"`
	Longitude     *float64   `json:"longitude,omitempty"`
	EventTime     time.Time  `json:"event_time"`
	Source        string     `json:"source"`
	RawPayload    *string    `json:"raw_payload,omitempty"`
}

// TrackingJourneyMilestone represents an evaluated lifecycle milestone in the operational journey (Task 17.4)
type TrackingJourneyMilestone struct {
	Code        string     `json:"code"`
	Label       string     `json:"label"`
	State       string     `json:"state"` // COMPLETED, CURRENT, UPCOMING, DELAYED
	PlannedDate *time.Time `json:"planned_date,omitempty"`
	ActualDate  *time.Time `json:"actual_date,omitempty"`
	Location    string     `json:"location"`
	IsCompleted bool       `json:"is_completed"`
	IsCurrent   bool       `json:"is_current"`
	IsDelayed   bool       `json:"is_delayed"`
	DelayDays   float64    `json:"delay_days,omitempty"`
	Notes       *string    `json:"notes,omitempty"`
}

// TrackingJourney represents full normalized journey lifecycle progression (Task 17.4)
type TrackingJourney struct {
	Milestones            []TrackingJourneyMilestone `json:"milestones"`
	CurrentMilestoneCode  string                     `json:"current_milestone_code"`
	CurrentMilestoneLabel string                     `json:"current_milestone_label"`
	TotalMilestones       int                        `json:"total_milestones"`
	CompletedMilestones   int                        `json:"completed_milestones"`
	CriticalPathDelayDays float64                    `json:"critical_path_delay_days"`
}

// TrackingSchedule represents evaluated planned vs actual transit schedule and variance (Task 17.4)
type TrackingSchedule struct {
	PlannedETD            *time.Time `json:"planned_etd,omitempty"`
	ActualETD             *time.Time `json:"actual_etd,omitempty"`
	PlannedETA            *time.Time `json:"planned_eta,omitempty"`
	EstimatedArrival      *time.Time `json:"estimated_arrival,omitempty"`
	DepartureVarianceDays float64    `json:"departure_variance_days"`
	ArrivalVarianceDays   float64    `json:"arrival_variance_days"`
	DepartureState        string     `json:"departure_state"` // ON_SCHEDULE, DELAYED, EARLY, AWAITING_DATA
	ArrivalState          string     `json:"arrival_state"`   // ON_SCHEDULE, DELAYED, EARLY, AT_RISK, AWAITING_DATA
	OverallVarianceState  string     `json:"overall_variance_state"`
}

// TrackingAlert represents calculated operational intelligence alerts (Task 17.4)
type TrackingAlert struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`     // DELAYED_DEPARTURE, ETA_AT_RISK, CRITICAL_EXCEPTION, STALE_TRACKING, MILESTONE_OVERDUE
	Severity    string    `json:"severity"` // CRITICAL, WARNING, INFO
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	ActionURL   *string   `json:"action_url,omitempty"`
}

// TrackingLineage represents upstream commercial and equipment references
type TrackingLineage struct {
	RFQID           *int64  `json:"rfq_id,omitempty"`
	QuoteID         *int64  `json:"quote_id,omitempty"`
	BookingID       *int64  `json:"booking_id,omitempty"`
	BookingNumber   string  `json:"booking_number"`
	ContainerNumber string  `json:"container_number"`
	MBLNumber       string  `json:"mbl_number"`
	CarrierSCAC     string  `json:"carrier_scac"`
	VesselName      string  `json:"vessel_name"`
	VoyageNumber    string  `json:"voyage_number"`
}

// TrackingProviderMetadata represents capabilities and configuration of active tracking provider (Task 17.6)
type TrackingProviderMetadata struct {
	ProviderName      string `json:"provider_name"`
	ProviderType      string `json:"provider_type"` // DEMO, CARRIER, AIS, DATABASE, UNAVAILABLE
	IsLive            bool   `json:"is_live"`
	IsConfigured      bool   `json:"is_configured"`
	SupportsPositions bool   `json:"supports_positions"`
	SupportsHistory   bool   `json:"supports_history"`
	SupportsEvents    bool   `json:"supports_events"`
	SupportsRefresh   bool   `json:"supports_refresh"`
	EnvironmentMode   string `json:"environment_mode,omitempty"` // development, staging, production
}

// TrackingRefreshResult represents the auditable normalized result of a provider telemetry refresh (Task 17.6)
type TrackingRefreshResult struct {
	Success       bool                          `json:"success"`
	Provider      TrackingProviderMetadata      `json:"provider"`
	DataFreshness string                        `json:"data_freshness"`
	LastUpdatedAt *time.Time                    `json:"last_updated_at,omitempty"`
	NewPositions  int                           `json:"new_positions"`
	NewEvents     int                           `json:"new_events"`
	UsedFallback  bool                          `json:"used_fallback"`
	Message       string                        `json:"message"`
	Intelligence  *ShipmentTrackingIntelligence `json:"intelligence,omitempty"`
}

// ShipmentTrackingIntelligence represents authoritative unified tracking intelligence (Task 17.4 & 17.6)
type ShipmentTrackingIntelligence struct {
	ShipmentID            int64                     `json:"shipment_id"`
	ShipmentNumber        string                    `json:"shipment_number"`
	ShipmentStatus        string                    `json:"shipment_status"`
	TrackingState         string                    `json:"tracking_state"`
	ClosureStatus         string                    `json:"closure_status"`
	ProgressPercentage    int                       `json:"progress_percentage"`
	DataFreshness         string                    `json:"data_freshness"`
	TrackingSource        string                    `json:"tracking_source"`
	IsLiveTracking        bool                      `json:"is_live_tracking"`
	ProviderMetadata      TrackingProviderMetadata  `json:"provider_metadata"`
	LastUpdatedAt         *time.Time                `json:"last_updated_at,omitempty"`
	Journey               TrackingJourney           `json:"journey"`
	Schedule              TrackingSchedule          `json:"schedule"`
	Alerts                []TrackingAlert           `json:"alerts"`
	Events                []TrackingEventNormalized `json:"events"`
	ActiveExceptionsCount int64                     `json:"active_exceptions_count"`
	LatestPosition        *TrackingPosition         `json:"latest_position,omitempty"`
	Lineage               TrackingLineage           `json:"lineage"`
}

// ShipmentTrackingAlertRecord represents persisted operational alert lifecycle record (Task 17.5)
type ShipmentTrackingAlertRecord struct {
	ID                int64      `db:"id" json:"id"`
	OrgID             int64      `db:"org_id" json:"org_id"`
	ShipmentID        int64      `db:"shipment_id" json:"shipment_id"`
	AlertKey          string     `db:"alert_key" json:"alert_key"`
	AlertType         string     `db:"alert_type" json:"alert_type"`
	Severity          string     `db:"severity" json:"severity"`
	Title             string     `db:"title" json:"title"`
	Description       *string    `db:"description" json:"description,omitempty"`
	Status            string     `db:"status" json:"status"` // OPEN, ACKNOWLEDGED, RESOLVED, SUPPRESSED
	FirstDetectedAt   time.Time  `db:"first_detected_at" json:"first_detected_at"`
	LastDetectedAt    time.Time  `db:"last_detected_at" json:"last_detected_at"`
	AcknowledgedAt    *time.Time `db:"acknowledged_at" json:"acknowledged_at,omitempty"`
	AcknowledgedBy    *int64     `db:"acknowledged_by" json:"acknowledged_by,omitempty"`
	AcknowledgedByName *string   `db:"acknowledged_by_name" json:"acknowledged_by_name,omitempty"`
	ResolvedAt        *time.Time `db:"resolved_at" json:"resolved_at,omitempty"`
	ResolvedBy        *int64     `db:"resolved_by" json:"resolved_by,omitempty"`
	ResolvedByName    *string    `db:"resolved_by_name" json:"resolved_by_name,omitempty"`
	SuppressedAt      *time.Time `db:"suppressed_at" json:"suppressed_at,omitempty"`
	SuppressedBy      *int64     `db:"suppressed_by" json:"suppressed_by,omitempty"`
	NotificationCount int        `db:"notification_count" json:"notification_count"`
	LastNotifiedAt    *time.Time `db:"last_notified_at" json:"last_notified_at,omitempty"`
	Metadata          *string    `db:"metadata" json:"metadata,omitempty"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

// TrackingRefreshRunRecord represents a persistent, auditable tracking refresh execution (Task 17.7)
type TrackingRefreshRunRecord struct {
	ID            int64      `db:"id" json:"id"`
	OrgID         int64      `db:"org_id" json:"org_id"`
	ShipmentID    int64      `db:"shipment_id" json:"shipment_id"`
	ProviderName  *string    `db:"provider_name" json:"provider_name,omitempty"`
	ProviderType  *string    `db:"provider_type" json:"provider_type,omitempty"`
	TriggerType   string     `db:"trigger_type" json:"trigger_type"` // MANUAL, SCHEDULED, SYSTEM_RETRY
	Status        string     `db:"status" json:"status"`             // STARTED, SUCCESS, PARTIAL, FAILED, SKIPPED
	StartedAt     time.Time  `db:"started_at" json:"started_at"`
	CompletedAt   *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	NewPositions  int        `db:"new_positions" json:"new_positions"`
	NewEvents     int        `db:"new_events" json:"new_events"`
	DataFreshness *string    `db:"data_freshness" json:"data_freshness,omitempty"`
	UsedFallback  bool       `db:"used_fallback" json:"used_fallback"`
	ErrorMessage  *string    `db:"error_message" json:"error_message,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

// TrackingRefreshConfig represents state-based refresh scheduling intervals (Task 17.7)
type TrackingRefreshConfig struct {
	AutoRefreshEnabled bool          `json:"auto_refresh_enabled"`
	IntervalBooked     time.Duration `json:"interval_booked"`
	IntervalDeparted   time.Duration `json:"interval_departed"`
	IntervalInTransit  time.Duration `json:"interval_in_transit"`
	IntervalArrived    time.Duration `json:"interval_arrived"`
	CheckInterval      time.Duration `json:"check_interval"`
}

// TrackingMonitoringSummary represents fleet-wide or shipment-specific operational health summary (Task 17.5 & 17.6 & 17.7)
type TrackingMonitoringSummary struct {
	ShipmentID              int64                    `json:"shipment_id"`
	OpenAlerts              int                      `json:"open_alerts"`
	CriticalAlerts          int                      `json:"critical_alerts"`
	AcknowledgedAlerts      int                      `json:"acknowledged_alerts"`
	SuppressedAlerts        int                      `json:"suppressed_alerts"`
	ResolvedAlerts          int                      `json:"resolved_alerts"`
	LastTrackingRefresh     *time.Time               `json:"last_tracking_refresh,omitempty"`
	LastRefreshAt           *time.Time               `json:"last_refresh_at,omitempty"`
	NextRefreshAt           *time.Time               `json:"next_refresh_at,omitempty"`
	RefreshStatus           string                   `json:"refresh_status"`
	LastRefreshStatus       string                   `json:"last_refresh_status"`
	LastRefreshError        *string                  `json:"last_refresh_error,omitempty"`
	AutomaticRefreshEnabled bool                     `json:"automatic_refresh_enabled"`
	RefreshIntervalMinutes  int                      `json:"refresh_interval_minutes"`
	TrackingFreshness       string                   `json:"tracking_freshness"`
	TrackingProvider        string                   `json:"tracking_provider"`
	ProviderMetadata        TrackingProviderMetadata `json:"provider_metadata"`
	IsLiveTracking          bool                     `json:"is_live_tracking"`
	NextRecommendedAction   string                   `json:"next_recommended_action"`
}

// Tracking Alert Mutation Requests
type AcknowledgeTrackingAlertRequest struct {
	ShipmentID int64  `json:"shipment_id"`
	AlertID    int64  `json:"alert_id"`
	Notes      string `json:"notes,omitempty"`
}

type ResolveTrackingAlertRequest struct {
	ShipmentID      int64  `json:"shipment_id"`
	AlertID         int64  `json:"alert_id"`
	ResolutionNotes string `json:"resolution_notes,omitempty"`
}

type SuppressTrackingAlertRequest struct {
	ShipmentID int64  `json:"shipment_id"`
	AlertID    int64  `json:"alert_id"`
	Reason     string `json:"reason,omitempty"`
}

type GetTrackingAlertsRequest struct {
	ShipmentID int64  `json:"shipment_id"`
	OrgID      int64  `json:"org_id"`
	Status     string `json:"status,omitempty"`
}

// ─── Tracking Analytics & Performance Intelligence (Task 17.8) ────────────────

// TrackingOperationalInsight represents deterministic operational guidance based on real telemetry & milestone metrics
type TrackingOperationalInsight struct {
	ID             string `json:"id"`
	Category       string `json:"category"` // CARRIER, ROUTE, DATA_FRESHNESS, ALERT_SPIKE, PERFORMANCE
	Severity       string `json:"severity"` // CRITICAL, HIGH, MEDIUM, LOW, POSITIVE
	Title          string `json:"title"`
	Description    string `json:"description"`
	Metric         string `json:"metric"`
	Recommendation string `json:"recommendation"`
}

// TrackingAnalyticsOverview represents executive KPIs and fleet performance aggregations
type TrackingAnalyticsOverview struct {
	ActiveShipments          int                           `json:"active_shipments"`
	TotalTrackedShipments    int                           `json:"total_tracked_shipments"`
	OnTimePercentage         float64                       `json:"on_time_percentage"`
	DelayedShipments         int                           `json:"delayed_shipments"`
	EarlyShipments           int                           `json:"early_shipments"`
	OnScheduleShipments      int                           `json:"on_schedule_shipments"`
	AverageDelayHours        float64                       `json:"average_delay_hours"`
	AverageEtaVarianceHours  float64                       `json:"average_eta_variance_hours"`
	OpenCriticalAlerts       int                           `json:"open_critical_alerts"`
	TotalOpenAlerts          int                           `json:"total_open_alerts"`
	DataFreshnessLive        int                           `json:"data_freshness_live"`
	DataFreshnessRecent      int                           `json:"data_freshness_recent"`
	DataFreshnessStale       int                           `json:"data_freshness_stale"`
	DataFreshnessUnavailable int                           `json:"data_freshness_unavailable"`
	RefreshSuccessRate       float64                       `json:"refresh_success_rate"`
	TotalRefreshes30d        int                           `json:"total_refreshes_30d"`
	FailedRefreshes30d       int                           `json:"failed_refreshes_30d"`
	Insights                 []TrackingOperationalInsight  `json:"insights"`
}

// TrackingTrendDataPoint represents time-series performance metrics for charts
type TrackingTrendDataPoint struct {
	Date               string  `json:"date"`
	OnTimeRate         float64 `json:"on_time_rate"`
	DelayedCount       int     `json:"delayed_count"`
	AlertCount         int     `json:"alert_count"`
	RefreshSuccessRate float64 `json:"refresh_success_rate"`
	TotalShipments     int     `json:"total_shipments"`
}

// CarrierTrackingPerformance represents aggregated reliability and tracking performance by ocean/air carrier
type CarrierTrackingPerformance struct {
	CarrierSCAC         string  `json:"carrier_scac"`
	CarrierName         string  `json:"carrier_name"`
	ShipmentsTracked    int     `json:"shipments_tracked"`
	OnTimeCount         int     `json:"on_time_count"`
	DelayedCount        int     `json:"delayed_count"`
	OnTimeRate          float64 `json:"on_time_rate"`
	AverageDelayHours   float64 `json:"average_delay_hours"`
	AlertCount          int     `json:"alert_count"`
	CriticalAlertCount  int     `json:"critical_alert_count"`
	ReliabilityScore    float64 `json:"reliability_score"` // 0 - 100
	ReliabilityTier     string  `json:"reliability_tier"`  // EXCELLENT, GOOD, FAIR, AT_RISK
}

// RouteTrackingPerformance represents corridor-level performance, transit variance, and risk rating
type RouteTrackingPerformance struct {
	RouteKey                string  `json:"route_key"` // e.g. "INNSA -> NLRTM"
	OriginPort              string  `json:"origin_port"`
	DestinationPort         string  `json:"destination_port"`
	ShipmentsCount          int     `json:"shipments_count"`
	OnTimeRate              float64 `json:"on_time_rate"`
	AvgTransitHours         float64 `json:"avg_transit_hours"`
	AvgTransitVarianceHours float64 `json:"avg_transit_variance_hours"`
	AlertCount              int     `json:"alert_count"`
	RiskLevel               string  `json:"risk_level"` // LOW, MODERATE, HIGH
}


