package spec

import (
	"encoding/json"
	"time"
)


// ──────────────────────────────────────────────────────────────────────────────
// Requirements Engine — Spec Types
// All evaluation is derived from real existing schema fields:
//   rfqs: origin, destination, incoterms, target_date, agent_status, lead_id, stage
//   rfq_items: description, weight_kg, volume_cbm
//   customers: contact_name, contact_email, contact_phone
//   leads: email, phone, ai_score
//   lead_interactions: intent, ai_confidence, partial_rfq_context
// ──────────────────────────────────────────────────────────────────────────────

// Requirement is a single evaluated operational requirement.

type Requirement struct {
	ID              string `json:"id"`
	Category        string `json:"category"`
	Type            string `json:"type"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	Status          string `json:"status"`
	Severity        string `json:"severity"`
	Value           string `json:"value,omitempty"`        // Current field value, if available
	IsConditional   bool   `json:"is_conditional"`
	ConditionReason string `json:"condition_reason,omitempty"` // Why this conditional applies
	SourceContext   string `json:"source_context,omitempty"`   // e.g. "From Lead #123 email thread"
}

// AIFinding is an intelligence finding from the existing agent_status and lead_interactions
// data. It does NOT override deterministic requirements — it supplements them.
type AIFinding struct {
	ID                  string `json:"id"`
	Title               string `json:"title"`
	Description         string `json:"description"`
	Confidence          string `json:"confidence"` // HIGH | MEDIUM | LOW
	Recommendation      string `json:"recommendation"`
	RequiresHumanReview bool   `json:"requires_human_review"`
	SourceContext       string `json:"source_context,omitempty"` // e.g. "From lead_interaction ai_confidence=87"
}

// DocumentRequirement tracks a specific document's readiness for this shipment.
// Stage-aware: future documents are NOT_APPLICABLE at RFQ stage and must
// not appear as blockers during quotation.
type DocumentRequirement struct {
	DocType         string     `json:"doc_type"`
	Title           string     `json:"title"`
	Status          string     `json:"status"`           // PENDING | SATISFIED | MISSING | NOT_APPLICABLE | UNDER_REVIEW
	ApplicableStage string     `json:"applicable_stage"` // When this document is actually required
	IsRequired      bool       `json:"is_required"`      // Required at the CURRENT stage?
	IsConditional   bool       `json:"is_conditional"`
	Reason          string     `json:"reason,omitempty"`
	DocumentID      *int64     `json:"document_id,omitempty"`
	DocumentStatus  string     `json:"document_status,omitempty"` // UPLOADED | UNDER_REVIEW | APPROVED | REJECTED
	FileName        *string    `json:"file_name,omitempty"`
	FileURL         *string    `json:"file_url,omitempty"`
	UploadedAt      *time.Time `json:"uploaded_at,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Document Management & Lifecycle — Spec Types (Task 12)
// ──────────────────────────────────────────────────────────────────────────────

// RFQDocument represents a persistent document record associated with an RFQ.
type RFQDocument struct {

	ID              int64                  `json:"id" db:"id"`
	OrgID           int32                  `json:"org_id" db:"org_id"`
	RFQID           int32                  `json:"rfq_id" db:"rfq_id"`
	DocumentType    string                 `json:"document_type" db:"document_type"`
	DocumentName    string                 `json:"document_name" db:"document_name"`
	Description     *string                `json:"description" db:"description"`
	Status          string                 `json:"status" db:"status"`
	FileName        *string                `json:"file_name" db:"file_name"`
	FileURL         *string                `json:"file_url" db:"file_url"`
	FileSize        *int64                 `json:"file_size" db:"file_size"`
	MimeType        *string                `json:"mime_type" db:"mime_type"`
	UploadedBy      *string                `json:"uploaded_by" db:"uploaded_by"`
	UploadedAt      *time.Time             `json:"uploaded_at" db:"uploaded_at"`
	ReviewedBy      *string                `json:"reviewed_by" db:"reviewed_by"`
	ReviewedAt      *time.Time             `json:"reviewed_at" db:"reviewed_at"`
	RejectionReason *string                `json:"rejection_reason" db:"rejection_reason"`
	ExpiresAt       *time.Time             `json:"expires_at" db:"expires_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" db:"-"`
	MetadataRaw     []byte                 `json:"-" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// UnmarshalMetadata deserializes raw JSON metadata from the DB into the Metadata map.
func (d *RFQDocument) UnmarshalMetadata() {
	if len(d.MetadataRaw) > 0 && string(d.MetadataRaw) != "{}" && string(d.MetadataRaw) != "null" {
		_ = json.Unmarshal(d.MetadataRaw, &d.Metadata)
	}
}


// DocumentSummary provides operational aggregate counters for RFQ documents.
type DocumentSummary struct {
	TotalDocuments       int `json:"total_documents"`
	RequiredDocuments    int `json:"required_documents"`
	ReceivedDocuments    int `json:"received_documents"`
	MissingDocuments     int `json:"missing_documents"`
	UnderReviewDocuments int `json:"under_review_documents"`
	ApprovedDocuments    int `json:"approved_documents"`
	RejectedDocuments    int `json:"rejected_documents"`
	FutureStageDocuments int `json:"future_stage_documents"`
	ReadinessPercentage  int `json:"readiness_percentage"`
}

// ResolvedDocumentRequirement pairs a requirement rule with its real matching document record.
type ResolvedDocumentRequirement struct {
	DocumentRequirement
	DocumentRecord *RFQDocument `json:"document_record,omitempty"`
}

// GetDocumentsResponse is the structured payload for GET /api/v1/rfqs/:id/documents.
type GetDocumentsResponse struct {
	Summary               DocumentSummary               `json:"summary"`
	CurrentStageDocuments []ResolvedDocumentRequirement `json:"current_stage_documents"`
	ConditionalDocuments  []ResolvedDocumentRequirement `json:"conditional_documents"`
	FutureStageDocuments  []ResolvedDocumentRequirement `json:"future_stage_documents"`
	AllDocuments          []RFQDocument                 `json:"all_documents"`
}


// RequirementGroup is a named category of evaluated requirements.
type RequirementGroup struct {
	Category      string        `json:"category"`
	Title         string        `json:"title"`
	Icon          string        `json:"icon"`
	CompleteCount int           `json:"complete_count"`
	TotalCount    int           `json:"total_count"`
	Status        string        `json:"status"` // group-level status: COMPLETE | INCOMPLETE | ATTENTION
	Requirements  []Requirement `json:"requirements"`
}

// OperationalReadiness is the top-level summary returned to the frontend.
type OperationalReadiness struct {
	OverallStatus           string `json:"overall_status"`
	BlockingCount           int    `json:"blocking_count"`
	MissingRequiredCount    int    `json:"missing_required_count"`
	ConditionalAttentionCount int  `json:"conditional_attention_count"`
	CompleteCount           int    `json:"complete_count"`
	TotalCount              int    `json:"total_count"`
	ReadinessScore          int    `json:"readiness_score"` // 0–100
	NextBestAction          string `json:"next_best_action"`
}

// GetRequirementsResponse is the full response for GET /rfqs/:id/requirements.
type GetRequirementsResponse struct {
	OperationalReadiness OperationalReadiness  `json:"operational_readiness"`
	Groups               []RequirementGroup    `json:"groups"`
	DocumentRequirements []DocumentRequirement `json:"document_requirements"`
	AIFindings           []AIFinding           `json:"ai_findings"`
	LeadID               *int64                `json:"lead_id,omitempty"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Activity Timeline & Audit Trail Engine — Spec Types
// Aggregates real records across RFQ, Lead, Emails, AI Tasks, Requirements, Quotes.
// ──────────────────────────────────────────────────────────────────────────────

// ActivityEvent is a normalized operational event displayed on the timeline.

type ActivityEvent struct {
	ID                string                 `json:"id"`
	Type              ActivityEventType      `json:"type"`
	Category          string                 `json:"category"` // CUSTOMER, OPERATIONS, AI, REQUIREMENTS, DOCUMENTS, QUOTES, SYSTEM
	Title             string                 `json:"title"`
	Description       string                 `json:"description"`
	Timestamp         time.Time              `json:"timestamp"`
	ActorType         string                 `json:"actor_type"`
	ActorName         string                 `json:"actor_name"`
	SourceType        string                 `json:"source_type,omitempty"` // LEAD, RFQ, INTERACTION, TASK, QUOTE
	SourceID          string                 `json:"source_id,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	IsImportant       bool                   `json:"is_important"`
	RequiresAction    bool                   `json:"requires_action"`
	RelatedEntityType string                 `json:"related_entity_type,omitempty"`
	RelatedEntityID   string                 `json:"related_entity_id,omitempty"`
}

// ActivitySummary provides high-level count aggregations for the timeline header.
type ActivitySummary struct {
	TotalEvents         int        `json:"total_events"`
	CustomerEvents      int        `json:"customer_events"`
	OperationalEvents   int        `json:"operational_events"`
	AIEvents            int        `json:"ai_events"`
	RequirementsEvents  int        `json:"requirements_events"`
	DocumentEvents      int        `json:"document_events"`
	QuoteEvents         int        `json:"quote_events"`
	ActionRequiredCount int        `json:"action_required_count"`
	LatestActivityAt    *time.Time `json:"latest_activity_at,omitempty"`
}

// GetActivityResponse is the HTTP response for GET /api/v1/rfqs/:id/activity.
type GetActivityResponse struct {
	Summary ActivitySummary `json:"summary"`
	Events  []ActivityEvent `json:"events"`
	LeadID  *int64          `json:"lead_id,omitempty"`
}


// RFQ models

type RFQ struct {
	ID                int32      `json:"id" db:"id"`
	OrgID             int32      `json:"org_id" db:"org_id"`
	RFQNumber         string     `json:"rfq_number" db:"rfq_number"`
	CustomerID        int32      `json:"customer_id" db:"customer_id"`
	// CustomerName is populated via JOIN with the customers table in ListRFQs and GetRFQByID.
	// It is NOT stored in the rfqs table — it is a read-only display field.
	CustomerName        string     `json:"customer_name" db:"customer_name"`
	CustomerEmail       *string    `json:"customer_email,omitempty" db:"customer_email"`
	CustomerPhone       *string    `json:"customer_phone,omitempty" db:"customer_phone"`
	CustomerContactName *string    `json:"customer_contact_name,omitempty" db:"customer_contact_name"`
	Stage               string     `json:"stage" db:"stage"`
	Origin              *string    `json:"origin" db:"origin"`
	Destination         *string    `json:"destination" db:"destination"`
	Incoterms           *string    `json:"incoterms" db:"incoterms"`
	TargetDate          *time.Time `json:"target_date" db:"target_date"`
	SalesAssigneeID     *int32     `json:"sales_assignee_id" db:"sales_assignee_id"`
	PricingAssigneeID   *int32     `json:"pricing_assignee_id" db:"pricing_assignee_id"`
	HealthScore         int        `json:"health_score" db:"health_score"`
	AgentStatus         string     `json:"agent_status" db:"agent_status"`
	Status              string     `json:"status" db:"status"`
	LeadID              *int64        `json:"lead_id" db:"lead_id"`
	CreatedAt           time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at" db:"updated_at"`
	Items               []RFQItem     `json:"items,omitempty"`
	Quotes              []Quote       `json:"quotes,omitempty"`
	Documents           []RFQDocument `json:"documents,omitempty"`
}



type TimelineEvent struct {
	ID          string                 `json:"id"`
	EntityType  string                 `json:"entity_type"` // RFQ, LEAD, EMAIL, AI, QUOTE, SYSTEM
	EntityID    int64                  `json:"entity_id"`
	Category    string                 `json:"category"`    // LEAD, EMAIL, AI, RFQ, DOCUMENT, QUOTE, BOOKING, SYSTEM
	Action      string                 `json:"action"`
	Description string                 `json:"description"`
	Actor       string                 `json:"actor"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type RFQItem struct {
	ID          int32     `json:"id" db:"id"`
	RFQID       int32     `json:"rfq_id" db:"rfq_id"`
	Description string    `json:"description" db:"description"`
	Quantity    int       `json:"quantity" db:"quantity"`
	WeightKG    *float64  `json:"weight_kg" db:"weight_kg"`
	VolumeCBM   *float64  `json:"volume_cbm" db:"volume_cbm"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Quote Management & Commercial Decision Intelligence — Spec Types (Task 13)
// ──────────────────────────────────────────────────────────────────────────────

// QuoteCharge represents an itemized freight charge.
type QuoteCharge struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
}

// RFQQuote represents a full persistent carrier quote with commercial lifecycle data.
type RFQQuote struct {
	ID                    int64         `json:"id" db:"id"`
	RFQID                 int32         `json:"rfq_id" db:"rfq_id"`
	OrgID                 int32         `json:"org_id" db:"org_id"`
	CarrierID             *string       `json:"carrier_id,omitempty" db:"carrier_id"`
	CarrierName           string        `json:"carrier_name" db:"carrier_name"`
	QuoteReference        *string       `json:"quote_reference,omitempty" db:"quote_reference"`
	Status                string        `json:"status" db:"status"`
	Currency              string        `json:"currency" db:"currency"`
	BuyPrice              float64       `json:"buy_price" db:"buy_price"`
	SellPrice             float64       `json:"sell_price" db:"sell_price"`
	MarginAmount          float64       `json:"margin_amount"`
	MarginPercentage      float64       `json:"margin_percentage"`
	OceanFreight          float64       `json:"ocean_freight" db:"ocean_freight"`
	OriginCharges         float64       `json:"origin_charges" db:"origin_charges"`
	DestinationCharges    float64       `json:"destination_charges" db:"destination_charges"`
	TotalBuyPrice         float64       `json:"total_buy_price" db:"total_buy_price"`
	TransitTimeDays       *int          `json:"transit_time_days,omitempty" db:"transit_time_days"`
	FreeDays              *int          `json:"free_days,omitempty" db:"free_days"`
	ValidFrom             *time.Time    `json:"valid_from,omitempty" db:"valid_from"`
	ValidUntil            *time.Time    `json:"valid_until,omitempty" db:"valid_until"`
	ValidityStatus        string        `json:"validity_status"`
	DaysUntilExpiry       *int          `json:"days_until_expiry,omitempty"`
	ETD                   *time.Time    `json:"etd,omitempty" db:"etd"`
	ETA                   *time.Time    `json:"eta,omitempty" db:"eta"`
	Notes                 *string       `json:"notes,omitempty" db:"notes"`
	ApprovedBy            *string       `json:"approved_by,omitempty" db:"approved_by"`
	ApprovedAt            *time.Time    `json:"approved_at,omitempty" db:"approved_at"`
	Charges               []QuoteCharge `json:"charges" db:"-"`
	ChargesRaw            []byte        `json:"-" db:"charges"`
	IsRecommended         bool          `json:"is_recommended" db:"is_recommended"`
	ReliabilityScore      float64       `json:"reliability_score" db:"reliability_score"`
	HistoricalSuccessRate float64       `json:"historical_success_rate" db:"historical_success_rate"`
	AiReasoning           *string       `json:"ai_reasoning,omitempty" db:"ai_reasoning"`
	CreatedAt             time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time     `json:"updated_at" db:"updated_at"`
}

// UnmarshalCharges deserializes the JSON charges column from the DB.
func (q *RFQQuote) UnmarshalCharges() {
	if len(q.ChargesRaw) > 0 && string(q.ChargesRaw) != "null" && string(q.ChargesRaw) != "" {
		_ = json.Unmarshal(q.ChargesRaw, &q.Charges)
	}
	if q.Charges == nil {
		q.Charges = make([]QuoteCharge, 0)
	}
}

// QuoteComparison represents comparison intelligence metrics for a quote.
type QuoteComparison struct {
	QuoteID              int64      `json:"quote_id"`
	CarrierName          string     `json:"carrier_name"`
	QuoteReference       *string    `json:"quote_reference,omitempty"`
	Status               string     `json:"status"`
	Currency             string     `json:"currency"`
	BuyPrice             float64    `json:"buy_price"`
	SellPrice            float64    `json:"sell_price"`
	MarginAmount         float64    `json:"margin_amount"`
	MarginPercentage     float64    `json:"margin_percentage"`
	TransitTimeDays      *int       `json:"transit_time_days,omitempty"`
	ValidUntil           *time.Time `json:"valid_until,omitempty"`
	ValidityStatus       string     `json:"validity_status"`
	IsLowestCost         bool       `json:"is_lowest_cost"`
	IsHighestMargin      bool       `json:"is_highest_margin"`
	IsFastest            bool       `json:"is_fastest"`
	IsRecommended        bool       `json:"is_recommended"`
	IsApproved           bool       `json:"is_approved"`
	Score                float64    `json:"score"`
	RecommendationReason string     `json:"recommendation_reason"`
}

// QuoteSummary aggregates high-level commercial counts and best options.
type QuoteSummary struct {
	TotalQuotes             int      `json:"total_quotes"`
	DraftQuotes             int      `json:"draft_quotes"`
	RequestedQuotes         int      `json:"requested_quotes"`
	ReceivedQuotes          int      `json:"received_quotes"`
	UnderReviewQuotes       int      `json:"under_review_quotes"`
	RecommendedQuotes       int      `json:"recommended_quotes"`
	ApprovedQuotes          int      `json:"approved_quotes"`
	SelectedQuotes          int      `json:"selected_quotes"`
	ExpiredQuotes           int      `json:"expired_quotes"`
	QuotesExpiringSoon      int      `json:"quotes_expiring_soon"`
	LowestBuyAmount         *float64 `json:"lowest_buy_amount,omitempty"`
	HighestMarginAmount     *float64 `json:"highest_margin_amount,omitempty"`
	HighestMarginPercentage *float64 `json:"highest_margin_percentage,omitempty"`
	FastestTransitDays      *int     `json:"fastest_transit_days,omitempty"`
	RecommendedQuoteID      *int64   `json:"recommended_quote_id,omitempty"`
	ApprovedQuoteID         *int64   `json:"approved_quote_id,omitempty"`
	PrimaryCurrency         string   `json:"primary_currency"`
	HasMixedCurrencies      bool     `json:"has_mixed_currencies"`
}

// GetQuotesResponse is the authoritative payload for GET /api/v1/rfqs/:id/quotes.
type GetQuotesResponse struct {
	Summary          QuoteSummary         `json:"summary"`
	RFQReadiness     OperationalReadiness `json:"rfq_readiness"`
	Quotes           []RFQQuote           `json:"quotes"`
	Comparison       []QuoteComparison    `json:"comparison"`
	RecommendedQuote *RFQQuote            `json:"recommended_quote,omitempty"`
	ApprovedQuote    *RFQQuote            `json:"approved_quote,omitempty"`
}

type Quote struct {
	ID                    int32     `json:"id" db:"id"`
	RFQID                 int32     `json:"rfq_id" db:"rfq_id"`
	CarrierName           string    `json:"carrier_name" db:"carrier_name"`
	TransitTimeDays       *int      `json:"transit_time_days" db:"transit_time_days"`
	BuyPrice              float64   `json:"buy_price" db:"buy_price"`
	SellPrice             float64   `json:"sell_price" db:"sell_price"`
	IsRecommended         bool      `json:"is_recommended" db:"is_recommended"`
	ReliabilityScore      int       `json:"reliability_score" db:"reliability_score"`
	HistoricalSuccessRate float64   `json:"historical_success_rate" db:"historical_success_rate"`
	AiReasoning           *string   `json:"ai_reasoning" db:"ai_reasoning"`
	Status                string    `json:"status" db:"status"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}


// Responses

type ListRFQsResponse struct {
	Data       []RFQ `json:"data"`
	TotalCount int   `json:"total_count"`
}

type GetRFQResponse struct {
	Data RFQ `json:"data"`
}

type GetTimelineResponse struct {
	Data []TimelineEvent `json:"data"`
}

type GetAgentStatusResponse struct {
	Data interface{} `json:"data"`
}

type ParseShipmentResponse struct {
	Data interface{} `json:"data"`
}

// GetCarrierRatesResponse wraps the carrier service response for the HTTP layer.
// The data field contains a ranked list of carrier options with AI reasoning.
type GetCarrierRatesResponse struct {
	Data interface{} `json:"data"`
}

// ApproveQuoteResponse confirms the stage advance and returns the updated RFQ.
type ApproveQuoteResponse struct {
	Data interface{} `json:"data"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Task 14: Booking & Shipment Handoff Models
// ──────────────────────────────────────────────────────────────────────────────

type RFQBooking struct {
	ID                           int64      `json:"id" db:"id"`
	OrgID                        int64      `json:"org_id" db:"org_id"`
	RFQID                        int64      `json:"rfq_id" db:"rfq_id"`
	QuoteID                      *int64     `json:"quote_id,omitempty" db:"quote_id"`
	BookingNumber                string     `json:"booking_number" db:"booking_number"`
	CarrierID                    *string    `json:"carrier_id,omitempty" db:"carrier_id"`
	CarrierName                  string     `json:"carrier_name" db:"carrier_name"`
	CarrierSCAC                  *string    `json:"carrier_scac,omitempty" db:"carrier_scac"`
	CarrierBookingReference      *string    `json:"carrier_booking_reference,omitempty" db:"carrier_booking_reference"`
	CarrierBookingStatus         *string    `json:"carrier_booking_status,omitempty" db:"carrier_booking_status"`
	CarrierConfirmationReference *string    `json:"carrier_confirmation_reference,omitempty" db:"carrier_confirmation_reference"`
	CarrierBookingError          *string    `json:"carrier_booking_error,omitempty" db:"carrier_booking_error"`
	CarrierBookedAt              *time.Time `json:"carrier_booked_at,omitempty" db:"carrier_booked_at"`
	Status                       string     `json:"status" db:"status"`
	OriginPort                   string     `json:"origin_port" db:"origin_port"`
	DestinationPort              string     `json:"destination_port" db:"destination_port"`
	VesselName                   *string    `json:"vessel_name,omitempty" db:"vessel_name"`
	VoyageNumber                 *string    `json:"voyage_number,omitempty" db:"voyage_number"`
	ETD                          *time.Time `json:"etd,omitempty" db:"etd"`
	ETA                          *time.Time `json:"eta,omitempty" db:"eta"`
	CargoSummary                 *string    `json:"cargo_summary,omitempty" db:"cargo_summary"`
	SpecialInstructions          *string    `json:"special_instructions,omitempty" db:"special_instructions"`
	CreatedBy                    *string    `json:"created_by,omitempty" db:"created_by"`
	CreatedAt                    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                    time.Time  `json:"updated_at" db:"updated_at"`
}

type RFQShipment struct {
	ID               int64      `json:"id" db:"id"`
	OrgID            int64      `json:"org_id" db:"org_id"`
	RFQID            *int64     `json:"rfq_id,omitempty" db:"rfq_id"`
	QuoteID          *int64     `json:"quote_id,omitempty" db:"quote_id"`
	BookingID        *int64     `json:"booking_id,omitempty" db:"booking_id"`
	BookingNumber    *string    `json:"booking_number,omitempty" db:"booking_number"`
	CarrierSCAC      string     `json:"carrier_scac" db:"carrier_scac"`
	CarrierName      string     `json:"carrier_name"`
	Status           string     `json:"status" db:"status"`
	OriginPort       string     `json:"origin_port" db:"origin_port"`
	DestinationPort  string     `json:"destination_port" db:"destination_port"`
	VesselName       *string    `json:"vessel_name,omitempty" db:"vessel_name"`
	VoyageNumber     *string    `json:"voyage_number,omitempty" db:"voyage_number"`
	ContainerNumbers []string   `json:"container_numbers"`
	ETD              *time.Time `json:"etd,omitempty" db:"etd"`
	ETA              *time.Time `json:"eta,omitempty" db:"eta"`
	CurrentMilestone *string    `json:"current_milestone,omitempty"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
	ActiveExceptionsCount int64      `json:"active_exceptions_count" db:"active_exceptions_count"`
}

type BookingEligibility struct {
	IsEligible           bool     `json:"is_eligible"`
	Reasons              []string `json:"reasons"`
	MissingPrerequisites []string `json:"missing_prerequisites,omitempty"`
	ApprovedQuoteID      *int64   `json:"approved_quote_id,omitempty"`
	ApprovedCarrier      *string  `json:"approved_carrier,omitempty"`
	QuoteStatus          *string  `json:"quote_status,omitempty"`
	ReadinessScore       int      `json:"readiness_score"`
}

type BookingSummary struct {
	TotalBookings int         `json:"total_bookings"`
	ActiveBooking *RFQBooking `json:"active_booking,omitempty"`
	LatestStatus  *string     `json:"latest_status,omitempty"`
}

type GetBookingHandoffResponse struct {
	Eligibility BookingEligibility `json:"eligibility"`
	Summary     BookingSummary     `json:"summary"`
	Bookings    []RFQBooking       `json:"bookings"`
}

type ShipmentSummary struct {
	TotalShipments int          `json:"total_shipments"`
	ActiveShipment *RFQShipment `json:"active_shipment,omitempty"`
	LatestStatus   *string      `json:"latest_status,omitempty"`
}

type GetShipmentHandoffResponse struct {
	SourceBooking *RFQBooking     `json:"source_booking,omitempty"`
	Summary       ShipmentSummary `json:"summary"`
	Shipments     []RFQShipment   `json:"shipments"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Task 15: Dedicated Booking Workspace & Operations Models
// ──────────────────────────────────────────────────────────────────────────────

type BookingWorkspaceItem struct {
	ID                           int64      `json:"id" db:"id"`
	OrgID                        int64      `json:"org_id" db:"org_id"`
	RFQID                        int64      `json:"rfq_id" db:"rfq_id"`
	RFQNumber                    string     `json:"rfq_number" db:"rfq_number"`
	CustomerID                   *int64     `json:"customer_id,omitempty" db:"customer_id"`
	CustomerName                 string     `json:"customer_name" db:"customer_name"`
	QuoteID                      *int64     `json:"quote_id,omitempty" db:"quote_id"`
	QuoteReference               *string    `json:"quote_reference,omitempty" db:"quote_reference"`
	QuoteSellPrice               *float64   `json:"quote_sell_price,omitempty" db:"quote_sell_price"`
	Currency                     string     `json:"currency" db:"currency"`
	BookingNumber                string     `json:"booking_number" db:"booking_number"`
	CarrierID                    *string    `json:"carrier_id,omitempty" db:"carrier_id"`
	CarrierName                  string     `json:"carrier_name" db:"carrier_name"`
	CarrierSCAC                  *string    `json:"carrier_scac,omitempty" db:"carrier_scac"`
	CarrierBookingReference      *string    `json:"carrier_booking_reference,omitempty" db:"carrier_booking_reference"`
	CarrierBookingStatus         *string    `json:"carrier_booking_status,omitempty" db:"carrier_booking_status"`
	CarrierConfirmationReference *string    `json:"carrier_confirmation_reference,omitempty" db:"carrier_confirmation_reference"`
	CarrierBookingError          *string    `json:"carrier_booking_error,omitempty" db:"carrier_booking_error"`
	CarrierBookedAt              *time.Time `json:"carrier_booked_at,omitempty" db:"carrier_booked_at"`
	Status                       string     `json:"status" db:"status"`
	OriginPort                   string     `json:"origin_port" db:"origin_port"`
	DestinationPort              string     `json:"destination_port" db:"destination_port"`
	VesselName                   *string    `json:"vessel_name,omitempty" db:"vessel_name"`
	VoyageNumber                 *string    `json:"voyage_number,omitempty" db:"voyage_number"`
	ETD                          *time.Time `json:"etd,omitempty" db:"etd"`
	ETA                          *time.Time `json:"eta,omitempty" db:"eta"`
	CargoSummary                 *string    `json:"cargo_summary,omitempty" db:"cargo_summary"`
	SpecialInstructions          *string    `json:"special_instructions,omitempty" db:"special_instructions"`
	CreatedBy                    *string    `json:"created_by,omitempty" db:"created_by"`
	ShipmentID                   *int64     `json:"shipment_id,omitempty" db:"shipment_id"`
	ShipmentStatus               *string    `json:"shipment_status,omitempty" db:"shipment_status"`
	CreatedAt                    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                    time.Time  `json:"updated_at" db:"updated_at"`
}

type BookingKPIs struct {
	TotalBookings       int `json:"total_bookings" db:"total_bookings"`
	Draft               int `json:"draft" db:"draft"`
	Requested           int `json:"requested" db:"requested"`
	PendingConfirmation int `json:"pending_confirmation" db:"pending_confirmation"`
	Confirmed           int `json:"confirmed" db:"confirmed"`
	Cancelled           int `json:"cancelled" db:"cancelled"`
	Completed           int `json:"completed" db:"completed"`
	DepartingSoon       int `json:"departing_soon" db:"departing_soon"`
}

type BookingPagination struct {
	CurrentPage int `json:"current_page"`
	PageSize    int `json:"page_size"`
	TotalItems  int `json:"total_items"`
	TotalPages  int `json:"total_pages"`
}

type BookingCarrierInfo struct {
	CarrierName string `json:"carrier_name" db:"carrier_name"`
	CarrierSCAC string `json:"carrier_scac" db:"carrier_scac"`
}

type GetBookingsWorkspaceResponse struct {
	Bookings   []BookingWorkspaceItem `json:"bookings"`
	KPIs       BookingKPIs            `json:"kpis"`
	Pagination BookingPagination      `json:"pagination"`
	Carriers   []BookingCarrierInfo   `json:"carriers"`
}

type EligibleBookingRFQ struct {
	RFQID            int64      `json:"rfq_id" db:"rfq_id"`
	RFQNumber        string     `json:"rfq_number" db:"rfq_number"`
	CustomerID       *int64     `json:"customer_id,omitempty" db:"customer_id"`
	CustomerName     string     `json:"customer_name" db:"customer_name"`
	OriginPort       string     `json:"origin_port" db:"origin_port"`
	DestinationPort  string     `json:"destination_port" db:"destination_port"`
	ApprovedQuoteID  int64      `json:"approved_quote_id" db:"approved_quote_id"`
	QuoteReference   *string    `json:"quote_reference,omitempty" db:"quote_reference"`
	CarrierName      string     `json:"carrier_name" db:"carrier_name"`
	CarrierSCAC      *string    `json:"carrier_scac,omitempty" db:"carrier_scac"`
	Currency         string     `json:"currency" db:"currency"`
	SellPrice        float64    `json:"sell_price" db:"sell_price"`
	BuyPrice         float64    `json:"buy_price" db:"buy_price"`
	TransitTimeDays  *int       `json:"transit_time_days,omitempty" db:"transit_time_days"`
	CargoDescription string     `json:"cargo_description" db:"cargo_description"`
	PackageCount     int        `json:"package_count" db:"package_count"`
	TotalWeightKg    float64    `json:"total_weight_kg" db:"total_weight_kg"`
	TotalVolumeCbm   float64    `json:"total_volume_cbm" db:"total_volume_cbm"`
	TargetDate       *time.Time `json:"target_date,omitempty" db:"target_date"`
}

type BookingDetailCargoSummary struct {
	ItemsCount     int     `json:"items_count"`
	TotalWeightKg  float64 `json:"total_weight_kg"`
	TotalVolumeCbm float64 `json:"total_volume_cbm"`
	CargoType      string  `json:"cargo_type"`
	Commodity      string  `json:"commodity"`
	PackagingType  string  `json:"packaging_type"`
}

type BookingDetailCommercialQuote struct {
	ID             int64      `json:"id"`
	CarrierName    string     `json:"carrier_name"`
	CarrierSCAC    *string    `json:"carrier_scac,omitempty"`
	QuoteReference *string    `json:"quote_reference,omitempty"`
	Currency       string     `json:"currency"`
	BuyPrice       float64    `json:"buy_price"`
	SellPrice      float64    `json:"sell_price"`
	MarginAmount   float64    `json:"margin_amount"`
	MarginPercent  float64    `json:"margin_percent"`
	Status         string     `json:"status"`
	ValidUntil     *time.Time `json:"valid_until,omitempty"`
	ApprovedAt     *time.Time `json:"approved_at,omitempty"`
	ApprovedBy     *string    `json:"approved_by,omitempty"`
}

type BookingDetailSourceRFQ struct {
	ID              int64      `json:"id" db:"id"`
	RFQNumber       string     `json:"rfq_number" db:"rfq_number"`
	LeadID          *int64     `json:"lead_id,omitempty" db:"lead_id"`
	CustomerID      *int64     `json:"customer_id,omitempty" db:"customer_id"`
	CustomerName    string     `json:"customer_name" db:"customer_name"`
	OriginPort      string     `json:"origin_port" db:"origin_port"`
	DestinationPort string     `json:"destination_port" db:"destination_port"`
	Status          string     `json:"status" db:"status"`
	Stage           string     `json:"stage" db:"stage"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

type BookingDetailResponse struct {
	Booking         RFQBooking                    `json:"booking"`
	SourceRFQ       BookingDetailSourceRFQ        `json:"source_rfq"`
	CommercialQuote *BookingDetailCommercialQuote `json:"commercial_quote,omitempty"`
	CargoSummary    BookingDetailCargoSummary     `json:"cargo_summary"`
	LinkedShipment  *RFQShipment                  `json:"linked_shipment,omitempty"`
	ActivityEvents  []ActivityEvent               `json:"activity_events"`
	AllowedActions  []string                      `json:"allowed_actions"`
}


