package bcontext

import (
	"time"
)

// EntityType denotes the canonical LogisticsHQ business record type.
type EntityType string

const (
	EntityTypeCustomer  EntityType = "CUSTOMER"
	EntityTypeLead      EntityType = "LEAD"
	EntityTypeRFQ       EntityType = "RFQ"
	EntityTypeQuotation EntityType = "QUOTATION"
	EntityTypeShipment  EntityType = "SHIPMENT"
	EntityTypeInvoice   EntityType = "INVOICE"
	EntityTypeContract  EntityType = "CONTRACT"
	EntityTypeBooking   EntityType = "BOOKING"
)

// ContextRequest specifies which primary entity and organization to build context for.
type ContextRequest struct {
	OrgID         int64      `json:"org_id"`
	UserID        int64      `json:"user_id"`
	UserRole      string     `json:"user_role"`
	PrimaryType   EntityType `json:"primary_type"`
	PrimaryID     int64      `json:"primary_id"`
	CorrelationID string     `json:"correlation_id,omitempty"`
}

// RecordSummary represents a sanitized, organization-scoped summary of a business record.
type RecordSummary struct {
	EntityType      EntityType             `json:"entity_type"`
	ID              int64                  `json:"id"`
	ReferenceNumber string                 `json:"reference_number"`
	Title           string                 `json:"title"`
	Status          string                 `json:"status"`
	Stage           string                 `json:"stage,omitempty"`
	CreatedAt       *time.Time             `json:"created_at,omitempty"`
	UpdatedAt       *time.Time             `json:"updated_at,omitempty"`
	KeyDates        map[string]string      `json:"key_dates,omitempty"`
	FinancialValues map[string]interface{} `json:"financial_values,omitempty"`
	KeyAttributes   map[string]interface{} `json:"key_attributes,omitempty"`
	DeepLink        string                 `json:"deep_link,omitempty"`
}

// ExceptionSummary captures an operational or data discrepancy for a business entity.
type ExceptionSummary struct {
	ID          int64      `json:"id"`
	Type        string     `json:"type"`
	Severity    string     `json:"severity"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
}

// MilestoneSummary captures a transit milestone or lifecycle progress checkpoint.
type MilestoneSummary struct {
	ID            int64      `json:"id"`
	Code          string     `json:"code"`
	Description   string     `json:"description"`
	Status        string     `json:"status"`
	PlannedDate   *time.Time `json:"planned_date,omitempty"`
	ActualDate    *time.Time `json:"actual_date,omitempty"`
	Location      string     `json:"location,omitempty"`
}

// ActivitySummary captures audit or timeline events related to the entity.
type ActivitySummary struct {
	ID        int64      `json:"id"`
	Action    string     `json:"action"`
	Actor     string     `json:"actor"`
	Summary   string     `json:"summary"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

// BusinessContext is the unified, read-only graph of business records for an entity.
type BusinessContext struct {
	OrgID                 int64                          `json:"org_id"`
	RequestingUserID      int64                          `json:"requesting_user_id"`
	RequestingUserRole    string                         `json:"requesting_user_role"`
	PrimaryType           EntityType                     `json:"primary_type"`
	PrimaryID             int64                          `json:"primary_id"`
	PrimaryRecord         *RecordSummary                 `json:"primary_record"`
	RelatedRecords        map[EntityType][]RecordSummary `json:"related_records"`
	OperationalExceptions []ExceptionSummary             `json:"operational_exceptions"`
	Milestones            []MilestoneSummary             `json:"milestones"`
	RecentActivity        []ActivitySummary              `json:"recent_activity"`
	SourceReferences      []SourceReference              `json:"source_references"`
	CorrelationID         string                         `json:"correlation_id"`
	DataFreshness         time.Time                      `json:"data_freshness"`
	IsReadOnly            bool                           `json:"is_read_only"`
}

// SourceReference explicitly identifies a supporting record for grounding and UI deep linking.
type SourceReference struct {
	EntityType      EntityType             `json:"entity_type"`
	EntityID        int64                  `json:"entity_id"`
	ReferenceNumber string                 `json:"reference_number"`
	Label           string                 `json:"label"`
	Status          string                 `json:"status"`
	Path            string                 `json:"path"`
	KeyFields       map[string]interface{} `json:"key_fields,omitempty"`
}

// FieldReference tracks a specific field key and value cited in an AI insight.
type FieldReference struct {
	SourceRecord string `json:"source_record"`
	FieldName    string `json:"field_name"`
	FieldValue   string `json:"field_value"`
}

// InsightRequest specifies what business intelligence insight is needed for a record.
type InsightRequest struct {
	OrgID         int64      `json:"org_id"`
	UserID        int64      `json:"user_id"`
	UserRole      string     `json:"user_role"`
	EntityType    EntityType `json:"entity_type"`
	EntityID      int64      `json:"entity_id"`
	Question      string     `json:"question,omitempty"`
	CorrelationID string     `json:"correlation_id,omitempty"`
}

// IntelligenceInsight provides a grounded, read-only AI analysis citing real records and fields.
type IntelligenceInsight struct {
	Title                      string            `json:"title"`
	Summary                    string            `json:"summary"`
	ConfidenceLevel            string            `json:"confidence_level"` // "HIGH", "MEDIUM", "LOW"
	IsInformationalOnly        bool              `json:"is_informational_only"`
	SupportingRecords          []SourceReference `json:"supporting_records"`
	SupportingFieldReferences  []FieldReference  `json:"supporting_field_references"`
	KeyHighlights              []string          `json:"key_highlights"`
	Warnings                   []string          `json:"warnings"`
	RecommendedFollowUp        []string          `json:"recommended_follow_up"`
	DataFreshness              string            `json:"data_freshness"`
	CorrelationID              string            `json:"correlation_id"`
	DataCompletenessPercentage int               `json:"data_completeness_percentage"`
}
