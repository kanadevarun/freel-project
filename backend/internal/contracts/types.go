package contracts

import "time"

// ProcessingStatus represents the status of a contract document as it moves through the AI extraction pipeline.
type ProcessingStatus string

const (
	// StatusQueued indicates the file has been uploaded and is waiting for processing.
	StatusQueued        ProcessingStatus = "QUEUED"
	// StatusOCRProcessing indicates the document is running through layout and optical text detection.
	StatusOCRProcessing ProcessingStatus = "OCR_PROCESSING"
	// StatusAIExtracting indicates that LLM agents are parsing tables and footnotes.
	StatusAIExtracting  ProcessingStatus = "AI_EXTRACTING"
	// StatusPendingReview indicates that rate anomalies have been flagged and require operator approval.
	StatusPendingReview ProcessingStatus = "PENDING_REVIEW"
	// StatusConfirmed indicates that all rates are successfully extracted, validated, and ingested.
	StatusConfirmed     ProcessingStatus = "CONFIRMED"
	// StatusFailed indicates an error occurred during text detection or AI agent parsing.
	StatusFailed        ProcessingStatus = "FAILED"
)

// LogEntry is a single entry in a document's processing history logs.
type LogEntry struct {
	// Step is the module identifier that wrote the log (e.g. UPLOAD, CLASSIFICATION).
	Step      string    `json:"step"`
	// Timestamp is the UTC execution time.
	Timestamp time.Time `json:"timestamp"`
	// Message is the detailed log statement content.
	Message   string    `json:"message"`
}

// ContractDocument represents an uploaded carrier contract document database model.
type ContractDocument struct {
	// ID is the unique UUID string (Primary Key).
	ID            string           `json:"id" db:"id"`
	// OrgID is the multi-tenant scope identifier.
	OrgID         int64            `json:"org_id" db:"org_id"`
	// CarrierSCAC holds the SCAC registry identifier (e.g., MAEU).
	CarrierSCAC   *string          `json:"carrier_scac" db:"carrier_scac"`
	// CarrierName is the matched display name (e.g., Maersk).
	CarrierName   *string          `json:"carrier_name" db:"carrier_name"`
	// FileName is the original local file name.
	FileName      string           `json:"file_name" db:"file_name"`
	// S3Key references the saved file name on storage.
	S3Key         string           `json:"s3_key" db:"s3_key"`
	// FileType is either PDF or XLSX.
	FileType      string           `json:"file_type" db:"file_type"`
	// FileSize is the file size in bytes.
	FileSize      int64            `json:"file_size_bytes" db:"file_size_bytes"`
	// PageCount is the number of parsed document pages.
	PageCount     *int             `json:"page_count" db:"page_count"`
	// Status is the current ProcessingStatus enum.
	Status        ProcessingStatus `json:"status" db:"status"`

	// Extraction Counts
	// ExtractedRateCount is the sum of confirmed rates and flagged review items.
	ExtractedRateCount int `json:"extracted_rate_count" db:"extracted_rate_count"`
	// ConfirmedRateCount represents the number of auto-ingested clean rates.
	ConfirmedRateCount int `json:"confirmed_rate_count" db:"confirmed_rate_count"`
	// PendingReviewCount represents the number of flagged items currently in the review queue.
	PendingReviewCount int `json:"pending_review_count" db:"pending_review_count"`
	// FailedRateCount represents failed extractions (currently not active, reserved for future use).
	FailedRateCount    int `json:"failed_rate_count" db:"failed_rate_count"`

	// Processing Metadata
	// ProcessingStartedAt records when the document moved out of QUEUED state.
	ProcessingStartedAt   *time.Time `json:"processing_started_at,omitempty" db:"processing_started_at"`
	// ProcessingCompletedAt records when the callback returned completion values.
	ProcessingCompletedAt *time.Time `json:"processing_completed_at,omitempty" db:"processing_completed_at"`
	// ProcessingLogRaw stores the log history in raw JSONB byte representation.
	ProcessingLogRaw      []byte     `json:"-" db:"processing_log"`
	// ProcessingLog is the deserialized slice of LogEntry structs.
	ProcessingLog         []LogEntry `json:"processing_log" db:"-"`
	// AIDocumentSummary contains the Gemini summarized terms.
	AIDocumentSummary     *string    `json:"ai_document_summary,omitempty" db:"ai_document_summary"`

	// Human Review
	// ReviewedBy holds the user ID of the operator who verified the document.
	ReviewedBy  *int64     `json:"reviewed_by,omitempty" db:"reviewed_by"`
	// ReviewedAt holds the date of review.
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
	// ReviewNotes holds the manual comments left by the reviewer.
	ReviewNotes *string    `json:"review_notes,omitempty" db:"review_notes"`

	// CreatedBy is the user ID who uploaded the file.
	CreatedBy int64     `json:"created_by" db:"created_by"`
	// CreatedAt is the record creation timestamp.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	// UpdatedAt is the record modification timestamp.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ReviewStatus represents the review progression status of an item in the human review queue.
type ReviewStatus string

const (
	// ReviewStatusPending indicates the rate is flagged and waiting for verification.
	ReviewStatusPending   ReviewStatus = "PENDING"
	// ReviewStatusApproved indicates the rate was accepted exactly as extracted.
	ReviewStatusApproved  ReviewStatus = "APPROVED"
	// ReviewStatusRejected indicates the rate was discarded by the reviewer.
	ReviewStatusRejected  ReviewStatus = "REJECTED"
	// ReviewStatusCorrected indicates the rate was adjusted and saved with manual changes.
	ReviewStatusCorrected ReviewStatus = "CORRECTED"
)

// RateReviewItem represents an item in the human verification queue database model.
type RateReviewItem struct {
	// ID is the unique rate review item UUID string (Primary Key).
	ID             string       `json:"id" db:"id"`
	// OrgID is the multi-tenant scope identifier.
	OrgID          int64        `json:"org_id" db:"org_id"`
	// ContractDocID is the foreign key back to the contract document.
	ContractDocID  string       `json:"contract_doc_id" db:"contract_doc_id"`
	// ExtractedData is the raw JSON representation of the proposed CanonicalRate.
	ExtractedData  []byte       `json:"extracted_data" db:"extracted_data"`
	// Confidence is the extraction quality score (0-100).
	Confidence     int          `json:"confidence_score" db:"confidence_score"`
	// ReviewFlags contains slice descriptors of flags (e.g. PRICE_ANOMALY).
	ReviewFlags    []string     `json:"review_flags" db:"-"`
	// ReviewFlagsRaw contains raw pq string array bytes from database.
	ReviewFlagsRaw []byte       `json:"-" db:"review_flags"`
	// AIReasoning explains why the rate validator flagged this record.
	AIReasoning    *string      `json:"ai_reasoning,omitempty" db:"ai_reasoning"`
	// SourcePage index.
	SourcePage     *int         `json:"source_page,omitempty" db:"source_page"`
	// SourceText details.
	SourceText     *string      `json:"source_text,omitempty" db:"source_text"`
	// SourceImageURL screenshot.
	SourceImageURL *string      `json:"source_image_url,omitempty" db:"source_image_url"`
	// Status represents the ReviewStatus state.
	Status         ReviewStatus `json:"status" db:"status"`
	// ReviewedBy holds the reviewer user ID.
	ReviewedBy     *int64       `json:"reviewed_by,omitempty" db:"reviewed_by"`
	// ReviewedAt is when review occurred.
	ReviewedAt     *time.Time   `json:"reviewed_at,omitempty" db:"reviewed_at"`
	// CorrectedData holds the corrected JSON rate representation.
	CorrectedData  []byte       `json:"corrected_data,omitempty" db:"corrected_data"`
	// ReviewNotes holds reviewer notes.
	ReviewNotes    *string      `json:"review_notes,omitempty" db:"review_notes"`
	// CreatedAt is the database creation time.
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
	// UpdatedAt is the database update time.
	UpdatedAt      time.Time    `json:"updated_at" db:"updated_at"`
}

// UploadRequest is the request payload to upload contract documents.
type UploadRequest struct {
	// CarrierSCAC is the optional carrier filter.
	CarrierSCAC *string `json:"carrier_scac"`
	// FileName is the local file name.
	FileName    string  `json:"file_name"`
	// FileType is either PDF or XLSX.
	FileType    string  `json:"file_type"`
}


// ── Commercial Contracts (Task 20.1) ───────────────────────────────────────

type ContractParty struct {
	ID           int64      `json:"id" db:"id"`
	OrgID        int64      `json:"org_id" db:"org_id"`
	PartyName    string     `json:"party_name" db:"party_name"`
	PartyType    PartyType  `json:"party_type" db:"party_type"`
	CustomerID   *int64     `json:"customer_id,omitempty" db:"customer_id"`
	CarrierID    *int64     `json:"carrier_id,omitempty" db:"carrier_id"`
	VendorID     *int64     `json:"vendor_id,omitempty" db:"vendor_id"`
	ContactName  *string    `json:"contact_name,omitempty" db:"contact_name"`
	ContactEmail *string    `json:"contact_email,omitempty" db:"contact_email"`
	ContactPhone *string    `json:"contact_phone,omitempty" db:"contact_phone"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

type Contract struct {
	ID                int64          `json:"id" db:"id"`
	OrgID             int64          `json:"org_id" db:"org_id"`
	ContractReference string         `json:"contract_reference" db:"contract_reference"`
	ContractName      string         `json:"contract_name" db:"contract_name"`
	ContractType      string         `json:"contract_type" db:"contract_type"`
	PartyID           int64          `json:"party_id" db:"party_id"`
	PartyName         string         `json:"party_name" db:"party_name"`
	TransportMode     *string        `json:"transport_mode,omitempty" db:"transport_mode"`
	Status            ContractStatus `json:"status" db:"status"`
	Currency          *string        `json:"currency,omitempty" db:"currency"`
	ContractValue     *float64       `json:"contract_value,omitempty" db:"contract_value"`
	EffectiveDate     *string        `json:"effective_date,omitempty" db:"effective_date"`
	ExpiryDate        *string        `json:"expiry_date,omitempty" db:"expiry_date"`
	Owner             *string        `json:"owner,omitempty" db:"owner"`
	Description       *string        `json:"description,omitempty" db:"description"`
	Notes             *string        `json:"notes,omitempty" db:"notes"`
	CreatedBy         *string        `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy         *string        `json:"updated_by,omitempty" db:"updated_by"`
	CreatedAt         time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at" db:"updated_at"`
	ArchivedAt        *time.Time     `json:"archived_at,omitempty" db:"archived_at"`
	SourceDocumentID  *string        `json:"source_document_id,omitempty" db:"source_document_id"`
}

type ContractLifecycleEvent struct {
	ID             int64      `json:"id" db:"id"`
	OrgID          int64      `json:"org_id" db:"org_id"`
	ContractID     int64      `json:"contract_id" db:"contract_id"`
	PreviousStatus *string    `json:"previous_status,omitempty" db:"previous_status"`
	NewStatus      string     `json:"new_status" db:"new_status"`
	EventType      EventType  `json:"event_type" db:"event_type"`
	Description    *string    `json:"description,omitempty" db:"description"`
	PerformedBy    *string    `json:"performed_by,omitempty" db:"performed_by"`
	Metadata       []byte     `json:"metadata,omitempty" db:"metadata"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// ── API Payload Models ─────────────────────────────────────────────────────

type CreateContractRequest struct {
	ContractReference string         `json:"contract_reference"`
	ContractName      string         `json:"contract_name"`
	ContractType      string         `json:"contract_type"`
	PartyID           int64          `json:"party_id"`
	TransportMode     *string        `json:"transport_mode,omitempty"`
	Status            ContractStatus `json:"status"`
	Currency          *string        `json:"currency,omitempty"`
	ContractValue     *float64       `json:"contract_value,omitempty"`
	EffectiveDate     *string        `json:"effective_date,omitempty"`
	ExpiryDate        *string        `json:"expiry_date,omitempty"`
	Owner             *string        `json:"owner,omitempty"`
	Description       *string        `json:"description,omitempty"`
	Notes             *string        `json:"notes,omitempty"`
}

type UpdateContractRequest struct {
	ContractName  *string        `json:"contract_name,omitempty"`
	ContractType  *string        `json:"contract_type,omitempty"`
	PartyID       *int64         `json:"party_id,omitempty"`
	TransportMode *string        `json:"transport_mode,omitempty"`
	Status        *ContractStatus `json:"status,omitempty"`
	Currency      *string        `json:"currency,omitempty"`
	ContractValue *float64       `json:"contract_value,omitempty"`
	EffectiveDate *string        `json:"effective_date,omitempty"`
	ExpiryDate    *string        `json:"expiry_date,omitempty"`
	Owner         *string        `json:"owner,omitempty"`
	Description   *string        `json:"description,omitempty"`
	Notes         *string        `json:"notes,omitempty"`
}

type UpdateContractLifecycleRequest struct {
	NewStatus   ContractStatus `json:"new_status"`
	Description *string        `json:"description,omitempty"`
}

type ContractOverview struct {
	TotalContracts   int     `json:"total_contracts" db:"total_contracts"`
	ActiveContracts  int     `json:"active_contracts" db:"active_contracts"`
	ExpiringSoon     int     `json:"expiring_soon" db:"expiring_soon"`
	ExpiredContracts int     `json:"expired_contracts" db:"expired_contracts"`
	DraftContracts   int     `json:"draft_contracts" db:"draft_contracts"`
	TotalValue       float64 `json:"total_value" db:"total_value"`
}

type ListContractsRequest struct {
	Page         int
	Limit        int
	Search       string
	Status       string
	PartyID      int64
	ContractType string
}

type ListContractsResponse struct {
	Data       []*Contract `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	TotalPages int         `json:"total_pages"`
}

// ── Contract Linking Models ──────────────────────────────────────────────────

// ContractLink represents a relationship between a contract and another entity.
type ContractLink struct {
	ID               int64            `json:"id" db:"id"`
	OrgID            int64            `json:"org_id" db:"org_id"`
	ContractID       int64            `json:"contract_id" db:"contract_id"`
	LinkedEntityType LinkedEntityType `json:"linked_entity_type" db:"linked_entity_type"`
	LinkedEntityID   int64            `json:"linked_entity_id" db:"linked_entity_id"`
	LinkType         LinkType         `json:"link_type" db:"link_type"`
	IsPrimary        bool             `json:"is_primary" db:"is_primary"`
	Notes            *string          `json:"notes" db:"notes"`
	CreatedBy        string           `json:"created_by" db:"created_by"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at" db:"updated_at"`
}

// ContractLinkHistory represents an immutable audit log entry for links.
type ContractLinkHistory struct {
	ID               int64            `json:"id" db:"id"`
	OrgID            int64            `json:"org_id" db:"org_id"`
	ContractID       int64            `json:"contract_id" db:"contract_id"`
	ContractLinkID   int64            `json:"contract_link_id" db:"contract_link_id"`
	LinkedEntityType LinkedEntityType `json:"linked_entity_type" db:"linked_entity_type"`
	LinkedEntityID   int64            `json:"linked_entity_id" db:"linked_entity_id"`
	LinkType         LinkType         `json:"link_type" db:"link_type"`
	Action           string           `json:"action" db:"action"`
	PreviousMetadata *string          `json:"previous_metadata" db:"previous_metadata"`
	PerformedBy      string           `json:"performed_by" db:"performed_by"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
}

// ContractLinkedRecord is the frontend-friendly hydrated representation of a link.
type ContractLinkedRecord struct {
	ContractLink
	// Resolved display metadata
	ReferenceName string  `json:"reference_name"`
	EntityStatus  string  `json:"entity_status"`
	PartyName     *string `json:"party_name,omitempty"`
	ValidityStart *string `json:"validity_start,omitempty"`
	ValidityEnd   *string `json:"validity_end,omitempty"`
	ExtraInfo     *string `json:"extra_info,omitempty"`
}

// ContractRelationshipSummary groups linked records by category for the UI.
type ContractRelationshipSummary struct {
	Parties          []ContractLinkedRecord `json:"parties"`
	CommercialRates  []ContractLinkedRecord `json:"commercial_rates"`
	Quotations       []ContractLinkedRecord `json:"quotations"`
	SpotRateActivity []ContractLinkedRecord `json:"spot_rate_activity"`
}

// CreateContractLinkRequest payload.
type CreateContractLinkRequest struct {
	LinkedEntityType LinkedEntityType `json:"linked_entity_type"`
	LinkedEntityID   int64            `json:"linked_entity_id"`
	LinkType         LinkType         `json:"link_type"`
	IsPrimary        bool             `json:"is_primary"`
	Notes            *string          `json:"notes"`
}

// UpdateContractLinkRequest payload.
type UpdateContractLinkRequest struct {
	LinkType  *LinkType `json:"link_type,omitempty"`
	IsPrimary *bool     `json:"is_primary,omitempty"`
	Notes     *string   `json:"notes,omitempty"`
}

// ContractLifecycleIntelligenceEvent represents an immutable audit entry for lifecycle state changes and detections.
type ContractLifecycleIntelligenceEvent struct {
	ID            int64                 `json:"id" db:"id"`
	OrgID         int64                 `json:"org_id" db:"org_id"`
	ContractID    int64                 `json:"contract_id" db:"contract_id"`
	EventType     IntelligenceEventType `json:"event_type" db:"event_type"`
	PreviousState *string               `json:"previous_state,omitempty" db:"previous_state"`
	NewState      *string               `json:"new_state,omitempty" db:"new_state"`
	Severity      RiskSeverity          `json:"severity" db:"severity"`
	Description   *string               `json:"description,omitempty" db:"description"`
	Metadata      *string               `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time             `json:"created_at" db:"created_at"`
}

// ContractRenewalTracking represents renewal workflow management state.
type ContractRenewalTracking struct {
	ID                   int64         `json:"id" db:"id"`
	OrgID                int64         `json:"org_id" db:"org_id"`
	ContractID           int64         `json:"contract_id" db:"contract_id"`
	RenewalStatus        RenewalStatus `json:"renewal_status" db:"renewal_status"`
	RenewalStartDate     *string       `json:"renewal_start_date,omitempty" db:"renewal_start_date"`
	TargetCompletionDate *string       `json:"target_completion_date,omitempty" db:"target_completion_date"`
	SuccessorContractID  *int64        `json:"successor_contract_id,omitempty" db:"successor_contract_id"`
	SuccessorName        *string       `json:"successor_name,omitempty"`
	SuccessorReference   *string       `json:"successor_reference,omitempty"`
	Owner                *string       `json:"owner,omitempty" db:"owner"`
	Notes                *string       `json:"notes,omitempty" db:"notes"`
	CreatedBy            string        `json:"created_by" db:"created_by"`
	CreatedAt            time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at" db:"updated_at"`
}

// ContractRiskEvent represents an actionable commercial or lifecycle risk.
type ContractRiskEvent struct {
	ID              int64        `json:"id" db:"id"`
	OrgID           int64        `json:"org_id" db:"org_id"`
	ContractID      int64        `json:"contract_id" db:"contract_id"`
	RiskType        RiskType     `json:"risk_type" db:"risk_type"`
	Severity        RiskSeverity `json:"severity" db:"severity"`
	Description     string       `json:"description" db:"description"`
	IsResolved      bool         `json:"is_resolved" db:"is_resolved"`
	ResolvedBy      *string      `json:"resolved_by,omitempty" db:"resolved_by"`
	ResolvedAt      *time.Time   `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolutionNotes *string      `json:"resolution_notes,omitempty" db:"resolution_notes"`
	Metadata        *string      `json:"metadata,omitempty" db:"metadata"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at" db:"updated_at"`
}

// ContractAttentionItem represents a prioritized contract requiring commercial attention.
type ContractAttentionItem struct {
	ContractID          int64              `json:"contract_id"`
	ContractName        string             `json:"contract_name"`
	ContractReference   string             `json:"contract_reference"`
	PartyName           string             `json:"party_name"`
	Status              ContractStatus     `json:"status"`
	Condition           LifecycleCondition `json:"condition"`
	ExpiryDate          *string            `json:"expiry_date,omitempty"`
	DaysRemaining       int                `json:"days_remaining"`
	Severity            RiskSeverity       `json:"severity"`
	RiskType            RiskType           `json:"risk_type"`
	RiskDescription     string             `json:"risk_description"`
	LinkedRecordsCount  int                `json:"linked_records_count"`
	RenewalStatus       RenewalStatus      `json:"renewal_status"`
	SuccessorContractID *int64             `json:"successor_contract_id,omitempty"`
}

// ContractCommercialImpactSummary analyzes the downstream records affected by a contract.
type ContractCommercialImpactSummary struct {
	ContractID              int64   `json:"contract_id"`
	LinkedRatesCount        int     `json:"linked_rates_count"`
	ActiveRatesCount        int     `json:"active_rates_count"`
	LinkedQuotationsCount   int     `json:"linked_quotations_count"`
	DraftQuotationsCount    int     `json:"draft_quotations_count"`
	AcceptedQuotationsCount int     `json:"accepted_quotations_count"`
	SpotRequestsCount       int     `json:"spot_requests_count"`
	SpotResponsesCount      int     `json:"spot_responses_count"`
	AffectedPartiesCount    int     `json:"affected_parties_count"`
	TotalCommercialExposure float64 `json:"total_commercial_exposure"`
}

// ContractLifecycleSummary provides aggregated portfolio intelligence metrics.
type ContractLifecycleSummary struct {
	TotalContracts         int `json:"total_contracts"`
	ActiveHealthy          int `json:"active_healthy"`
	Expiring7Days          int `json:"expiring_7_days"`
	Expiring30Days         int `json:"expiring_30_days"`
	Expiring60Days         int `json:"expiring_60_days"`
	Expiring90Days         int `json:"expiring_90_days"`
	ExpiredCount           int `json:"expired_count"`
	RenewalRequiredCount   int `json:"renewal_required_count"`
	RenewalInProgressCount int `json:"renewal_in_progress_count"`
	SupersededCount        int `json:"superseded_count"`
	CriticalRisksCount     int `json:"critical_risks_count"`
	WarningRisksCount      int `json:"warning_risks_count"`
}

// ContractLifecycleIntelligenceDetail aggregates full intelligence for a single contract.
type ContractLifecycleIntelligenceDetail struct {
	Contract         Contract                             `json:"contract"`
	Condition        LifecycleCondition                   `json:"condition"`
	DaysRemaining    int                                  `json:"days_remaining"`
	ExpiryThreshold  string                               `json:"expiry_threshold"`
	HealthLabel      string                               `json:"health_label"`
	HealthProgress   int                                  `json:"health_progress"`
	RenewalTracking  *ContractRenewalTracking             `json:"renewal_tracking,omitempty"`
	CommercialImpact ContractCommercialImpactSummary      `json:"commercial_impact"`
	ActiveRisks      []ContractRiskEvent                  `json:"active_risks"`
	RecentEvents     []ContractLifecycleIntelligenceEvent `json:"recent_events"`
}

// StartRenewalRequest payload.
type StartRenewalRequest struct {
	TargetCompletionDate *string `json:"target_completion_date,omitempty"`
	Owner                *string `json:"owner,omitempty"`
	Notes                *string `json:"notes,omitempty"`
}

// UpdateRenewalRequest payload.
type UpdateRenewalRequest struct {
	RenewalStatus        *RenewalStatus `json:"renewal_status,omitempty"`
	TargetCompletionDate *string        `json:"target_completion_date,omitempty"`
	SuccessorContractID  *int64         `json:"successor_contract_id,omitempty"`
	Owner                *string        `json:"owner,omitempty"`
	Notes                *string        `json:"notes,omitempty"`
}

// ResolveRiskRequest payload.
type ResolveRiskRequest struct {
	ResolutionNotes *string `json:"resolution_notes,omitempty"`
}

// ── Task 20.4: Contract Versioning, Amendments & Approvals Models ────────────

type ContractVersion struct {
	ID               string        `json:"id" db:"id"`
	OrgID            string        `json:"org_id" db:"org_id"`
	ContractID       int64         `json:"contract_id" db:"contract_id"`
	VersionNumber    int           `json:"version_number" db:"version_number"`
	VersionLabel     string        `json:"version_label" db:"version_label"`
	Status           VersionStatus `json:"status" db:"status"`
	EffectiveDate    *string       `json:"effective_date,omitempty" db:"effective_date"`
	ExpiryDate       *string       `json:"expiry_date,omitempty" db:"expiry_date"`
	ContractSnapshot []byte        `json:"contract_snapshot" db:"contract_snapshot"`
	ChangeSummary    *string       `json:"change_summary,omitempty" db:"change_summary"`
	CreatedBy        *string       `json:"created_by,omitempty" db:"created_by"`
	ApprovedBy       *string       `json:"approved_by,omitempty" db:"approved_by"`
	ApprovedAt       *time.Time    `json:"approved_at,omitempty" db:"approved_at"`
	SupersededAt     *time.Time    `json:"superseded_at,omitempty" db:"superseded_at"`
	CreatedAt        time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at" db:"updated_at"`
}

type ContractVersionSummary struct {
	ID            string        `json:"id"`
	ContractID    int64         `json:"contract_id"`
	VersionNumber int           `json:"version_number"`
	VersionLabel  string        `json:"version_label"`
	Status        VersionStatus `json:"status"`
	EffectiveDate *string       `json:"effective_date,omitempty"`
	ExpiryDate    *string       `json:"expiry_date,omitempty"`
	ChangeSummary *string       `json:"change_summary,omitempty"`
	CreatedBy     *string       `json:"created_by,omitempty"`
	ApprovedBy    *string       `json:"approved_by,omitempty"`
	ApprovedAt    *time.Time    `json:"approved_at,omitempty"`
	SupersededAt  *time.Time    `json:"superseded_at,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}

type ContractVersionComparison struct {
	BaseVersionID    string                   `json:"base_version_id"`
	BaseVersionNum   int                      `json:"base_version_num"`
	TargetVersionID  string                   `json:"target_version_id"`
	TargetVersionNum int                      `json:"target_version_num"`
	Changes          []ContractFieldChangeDiff `json:"changes"`
}

type ContractFieldChangeDiff struct {
	FieldName     string  `json:"field_name"`
	PreviousValue *string `json:"previous_value"`
	ProposedValue *string `json:"proposed_value"`
	ChangeType    string  `json:"change_type"` // ADD, MODIFY, REMOVE
}

type ContractAmendment struct {
	ID                    string          `json:"id" db:"id"`
	OrgID                 string          `json:"org_id" db:"org_id"`
	ContractID            int64           `json:"contract_id" db:"contract_id"`
	BaseVersionID         *string         `json:"base_version_id,omitempty" db:"base_version_id"`
	AmendmentReference    string          `json:"amendment_reference" db:"amendment_reference"`
	AmendmentType         AmendmentType   `json:"amendment_type" db:"amendment_type"`
	Title                 string          `json:"title" db:"title"`
	Description           *string         `json:"description,omitempty" db:"description"`
	ChangeSummary         *string         `json:"change_summary,omitempty" db:"change_summary"`
	Status                AmendmentStatus `json:"status" db:"status"`
	ProposedEffectiveDate *string         `json:"proposed_effective_date,omitempty" db:"proposed_effective_date"`
	CreatedBy             *string         `json:"created_by,omitempty" db:"created_by"`
	SubmittedAt           *time.Time      `json:"submitted_at,omitempty" db:"submitted_at"`
	ApprovedAt            *time.Time      `json:"approved_at,omitempty" db:"approved_at"`
	RejectedAt            *time.Time      `json:"rejected_at,omitempty" db:"rejected_at"`
	CreatedAt             time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at" db:"updated_at"`
	Changes               []ContractAmendmentChange `json:"changes,omitempty"`
}

type ContractAmendmentChange struct {
	ID            string    `json:"id" db:"id"`
	OrgID         string    `json:"org_id" db:"org_id"`
	AmendmentID   string    `json:"amendment_id" db:"amendment_id"`
	FieldName     string    `json:"field_name" db:"field_name"`
	PreviousValue *string   `json:"previous_value,omitempty" db:"previous_value"`
	ProposedValue *string   `json:"proposed_value,omitempty" db:"proposed_value"`
	ChangeType    string    `json:"change_type" db:"change_type"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type ContractApprovalRequest struct {
	ID              string         `json:"id" db:"id"`
	OrgID           string         `json:"org_id" db:"org_id"`
	ContractID      int64          `json:"contract_id" db:"contract_id"`
	VersionID       *string        `json:"version_id,omitempty" db:"version_id"`
	AmendmentID     *string        `json:"amendment_id,omitempty" db:"amendment_id"`
	ApprovalType    ApprovalType   `json:"approval_type" db:"approval_type"`
	Status          ApprovalStatus `json:"status" db:"status"`
	RequestedBy     *string        `json:"requested_by,omitempty" db:"requested_by"`
	AssignedTo      *string        `json:"assigned_to,omitempty" db:"assigned_to"`
	DecisionBy      *string        `json:"decision_by,omitempty" db:"decision_by"`
	DecisionComment *string        `json:"decision_comment,omitempty" db:"decision_comment"`
	RequestedAt     time.Time      `json:"requested_at" db:"requested_at"`
	DecidedAt       *time.Time     `json:"decided_at,omitempty" db:"decided_at"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
}

// ── Versioning & Amendment API Request DTOs ────────────────────────────────

type CreateContractVersionRequest struct {
	VersionLabel  *string `json:"version_label,omitempty"`
	ChangeSummary *string `json:"change_summary,omitempty"`
}

type CreateContractAmendmentRequest struct {
	AmendmentType         AmendmentType                  `json:"amendment_type"`
	Title                 string                         `json:"title"`
	Description           *string                        `json:"description,omitempty"`
	ChangeSummary         *string                        `json:"change_summary,omitempty"`
	ProposedEffectiveDate *string                        `json:"proposed_effective_date,omitempty"`
	Changes               []CreateAmendmentChangeRequest `json:"changes,omitempty"`
}

type CreateAmendmentChangeRequest struct {
	FieldName     string  `json:"field_name"`
	PreviousValue *string `json:"previous_value,omitempty"`
	ProposedValue *string `json:"proposed_value,omitempty"`
	ChangeType    string  `json:"change_type,omitempty"`
}

type UpdateContractAmendmentRequest struct {
	Title                 *string                        `json:"title,omitempty"`
	Description           *string                        `json:"description,omitempty"`
	ChangeSummary         *string                        `json:"change_summary,omitempty"`
	ProposedEffectiveDate *string                        `json:"proposed_effective_date,omitempty"`
	Changes               []CreateAmendmentChangeRequest `json:"changes,omitempty"`
}

type SubmitContractAmendmentRequest struct {
	AssignedTo *string `json:"assigned_to,omitempty"`
	Notes      *string `json:"notes,omitempty"`
}

type ApproveContractRequest struct {
	Comment string `json:"comment,omitempty"`
}

type RejectContractRequest struct {
	Reason string `json:"reason"`
}

type CancelApprovalRequest struct {
	Reason string `json:"reason,omitempty"`
}

// ═══════════════════════════════════════════════════════════════════════════
// Tasks 20.5 & 20.6: Terms, Obligations, Compliance, Performance Models
// ═══════════════════════════════════════════════════════════════════════════

// ContractAgreementDocument represents an attached agreement document or annexure
type ContractAgreementDocument struct {
	ID                   string     `json:"id" db:"id"`
	OrgID                int64      `json:"org_id" db:"org_id"`
	ContractID           *int64     `json:"contract_id" db:"contract_id"`
	ContractVersionID    *int64     `json:"contract_version_id" db:"contract_version_id"`
	DocumentType         string     `json:"document_type" db:"document_type"`
	DocumentName         *string    `json:"document_name" db:"document_name"`
	FileName             string     `json:"file_name" db:"file_name"`
	S3Key                string     `json:"s3_key" db:"s3_key"`
	FileType             string     `json:"file_type" db:"file_type"`
	FileSizeBytes        int64      `json:"file_size_bytes" db:"file_size_bytes"`
	Status               string     `json:"status" db:"status"`
	DocumentStatus       string     `json:"document_status" db:"document_status"`
	IsCurrent            bool       `json:"is_current" db:"is_current"`
	SupersedesDocumentID *string    `json:"supersedes_document_id" db:"supersedes_document_id"`
	EffectiveDate        *string    `json:"effective_date" db:"effective_date"`
	ExpiryDate           *string    `json:"expiry_date" db:"expiry_date"`
	Description          *string    `json:"description" db:"description"`
	CreatedBy            *int64     `json:"created_by" db:"created_by"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateAgreementDocumentRequest struct {
	DocumentType         string  `json:"document_type"`
	DocumentName         string  `json:"document_name"`
	FileName             string  `json:"file_name"`
	S3Key                string  `json:"s3_key"`
	FileType             string  `json:"file_type"`
	FileSizeBytes        int64   `json:"file_size_bytes"`
	ContractVersionID    *int64  `json:"contract_version_id,omitempty"`
	EffectiveDate        *string `json:"effective_date,omitempty"`
	ExpiryDate           *string `json:"expiry_date,omitempty"`
	Description          *string `json:"description,omitempty"`
	SupersedesDocumentID *string `json:"supersedes_document_id,omitempty"`
}

// ContractTerm represents a structured commercial, payment, or operational term
type ContractTerm struct {
	ID                int64     `json:"id" db:"id"`
	OrgID             int64     `json:"org_id" db:"org_id"`
	ContractID        int64     `json:"contract_id" db:"contract_id"`
	ContractVersionID *int64    `json:"contract_version_id" db:"contract_version_id"`
	TermCategory      string    `json:"term_category" db:"term_category"`
	TermKey           string    `json:"term_key" db:"term_key"`
	TermTitle         string    `json:"term_title" db:"term_title"`
	TermValue         string    `json:"term_value" db:"term_value"`
	ValueType         string    `json:"value_type" db:"value_type"`
	Currency          *string   `json:"currency" db:"currency"`
	EffectiveDate     *string   `json:"effective_date" db:"effective_date"`
	ExpiryDate        *string   `json:"expiry_date" db:"expiry_date"`
	DisplayOrder      int       `json:"display_order" db:"display_order"`
	IsCritical        bool      `json:"is_critical" db:"is_critical"`
	CreatedBy         *int64    `json:"created_by" db:"created_by"`
	UpdatedBy         *int64    `json:"updated_by" db:"updated_by"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

type CreateContractTermRequest struct {
	ContractVersionID *int64  `json:"contract_version_id,omitempty"`
	TermCategory      string  `json:"term_category"`
	TermKey           string  `json:"term_key"`
	TermTitle         string  `json:"term_title"`
	TermValue         string  `json:"term_value"`
	ValueType         string  `json:"value_type,omitempty"`
	Currency          *string `json:"currency,omitempty"`
	EffectiveDate     *string `json:"effective_date,omitempty"`
	ExpiryDate        *string `json:"expiry_date,omitempty"`
	DisplayOrder      int     `json:"display_order,omitempty"`
	IsCritical        bool    `json:"is_critical,omitempty"`
}

type UpdateContractTermRequest struct {
	TermCategory  *string `json:"term_category,omitempty"`
	TermTitle     *string `json:"term_title,omitempty"`
	TermValue     *string `json:"term_value,omitempty"`
	ValueType     *string `json:"value_type,omitempty"`
	Currency      *string `json:"currency,omitempty"`
	EffectiveDate *string `json:"effective_date,omitempty"`
	ExpiryDate    *string `json:"expiry_date,omitempty"`
	DisplayOrder  *int    `json:"display_order,omitempty"`
	IsCritical    *bool   `json:"is_critical,omitempty"`
}

// ContractObligation represents a contractual commitment or operational requirement
type ContractObligation struct {
	ID                  int64      `json:"id" db:"id"`
	OrgID               int64      `json:"org_id" db:"org_id"`
	ContractID          int64      `json:"contract_id" db:"contract_id"`
	ContractVersionID   *int64     `json:"contract_version_id" db:"contract_version_id"`
	ObligationReference string     `json:"obligation_reference" db:"obligation_reference"`
	Title               string     `json:"title" db:"title"`
	Description         *string    `json:"description" db:"description"`
	ObligationType      string     `json:"obligation_type" db:"obligation_type"`
	Category            string     `json:"category" db:"category"`
	ResponsibleParty    string     `json:"responsible_party" db:"responsible_party"`
	Owner               *string    `json:"owner" db:"owner"`
	Priority            string     `json:"priority" db:"priority"`
	Status              string     `json:"status" db:"status"`
	EffectiveDate       *string    `json:"effective_date" db:"effective_date"`
	DueDate             *string    `json:"due_date" db:"due_date"`
	CompletionDate      *time.Time `json:"completion_date" db:"completion_date"`
	IsRecurring         bool       `json:"is_recurring" db:"is_recurring"`
	RecurrenceType      string     `json:"recurrence_type" db:"recurrence_type"`
	TargetValue         *float64   `json:"target_value" db:"target_value"`
	TargetUnit          *string    `json:"target_unit" db:"target_unit"`
	CurrentValue        float64    `json:"current_value" db:"current_value"`
	WarningThreshold    *float64   `json:"warning_threshold" db:"warning_threshold"`
	CriticalThreshold   *float64   `json:"critical_threshold" db:"critical_threshold"`
	SourceDocumentID    *string    `json:"source_document_id" db:"source_document_id"`
	SourceTermID        *int64     `json:"source_term_id" db:"source_term_id"`
	Notes               *string    `json:"notes" db:"notes"`
	CreatedBy           *int64     `json:"created_by" db:"created_by"`
	FulfilledBy         *int64     `json:"fulfilled_by" db:"fulfilled_by"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateContractObligationRequest struct {
	ContractVersionID   *int64   `json:"contract_version_id,omitempty"`
	ObligationReference string   `json:"obligation_reference"`
	Title               string   `json:"title"`
	Description         *string  `json:"description,omitempty"`
	ObligationType      string   `json:"obligation_type"`
	Category            string   `json:"category,omitempty"`
	ResponsibleParty    string   `json:"responsible_party"`
	Owner               *string  `json:"owner,omitempty"`
	Priority            string   `json:"priority,omitempty"`
	EffectiveDate       *string  `json:"effective_date,omitempty"`
	DueDate             *string  `json:"due_date,omitempty"`
	IsRecurring         bool     `json:"is_recurring,omitempty"`
	RecurrenceType      string   `json:"recurrence_type,omitempty"`
	TargetValue         *float64 `json:"target_value,omitempty"`
	TargetUnit          *string  `json:"target_unit,omitempty"`
	WarningThreshold    *float64 `json:"warning_threshold,omitempty"`
	CriticalThreshold   *float64 `json:"critical_threshold,omitempty"`
	Notes               *string  `json:"notes,omitempty"`
}

type UpdateContractObligationRequest struct {
	Title             *string  `json:"title,omitempty"`
	Description       *string  `json:"description,omitempty"`
	Owner             *string  `json:"owner,omitempty"`
	Priority          *string  `json:"priority,omitempty"`
	Status            *string  `json:"status,omitempty"`
	DueDate           *string  `json:"due_date,omitempty"`
	CurrentValue      *float64 `json:"current_value,omitempty"`
	TargetValue       *float64 `json:"target_value,omitempty"`
	TargetUnit        *string  `json:"target_unit,omitempty"`
	WarningThreshold  *float64 `json:"warning_threshold,omitempty"`
	CriticalThreshold *float64 `json:"critical_threshold,omitempty"`
	Notes             *string  `json:"notes,omitempty"`
}

type FulfillObligationRequest struct {
	Notes string `json:"notes,omitempty"`
}

type WaiveObligationRequest struct {
	Reason string `json:"reason"`
}

// ContractComplianceEvent represents a detected compliance risk, SLA breach, or fulfillment event
type ContractComplianceEvent struct {
	ID                   int64      `json:"id" db:"id"`
	OrgID                int64      `json:"org_id" db:"org_id"`
	ContractID           int64      `json:"contract_id" db:"contract_id"`
	ContractObligationID *int64     `json:"contract_obligation_id" db:"contract_obligation_id"`
	RelatedEntityType    *string    `json:"related_entity_type" db:"related_entity_type"`
	RelatedEntityID      *int64     `json:"related_entity_id" db:"related_entity_id"`
	EventType            string     `json:"event_type" db:"event_type"`
	Severity             string     `json:"severity" db:"severity"`
	Status               string     `json:"status" db:"status"`
	Title                string     `json:"title" db:"title"`
	Description          *string    `json:"description" db:"description"`
	DetectedAt           time.Time  `json:"detected_at" db:"detected_at"`
	ResolvedAt           *time.Time `json:"resolved_at" db:"resolved_at"`
	ResolvedBy           *int64     `json:"resolved_by" db:"resolved_by"`
	ResolutionNotes      *string    `json:"resolution_notes" db:"resolution_notes"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

type ResolveComplianceEventRequest struct {
	ResolutionNotes string `json:"resolution_notes"`
}

// ContractComplianceRequirement represents a formal certificate, insurance, or regulatory mandate
type ContractComplianceRequirement struct {
	ID                 int64      `json:"id" db:"id"`
	OrgID              int64      `json:"org_id" db:"org_id"`
	ContractID         int64      `json:"contract_id" db:"contract_id"`
	RequirementType    string     `json:"requirement_type" db:"requirement_type"`
	Title              string     `json:"title" db:"title"`
	Description        *string    `json:"description" db:"description"`
	ResponsibleParty   string     `json:"responsible_party" db:"responsible_party"`
	ValidFrom          *string    `json:"valid_from" db:"valid_from"`
	ValidUntil         *string    `json:"valid_until" db:"valid_until"`
	Status             string     `json:"status" db:"status"`
	EvidenceDocumentID *string    `json:"evidence_document_id" db:"evidence_document_id"`
	VerificationDate   *time.Time `json:"verification_date" db:"verification_date"`
	VerifiedBy         *int64     `json:"verified_by" db:"verified_by"`
	RiskSeverity       string     `json:"risk_severity" db:"risk_severity"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateComplianceRequirementRequest struct {
	RequirementType    string  `json:"requirement_type"`
	Title              string  `json:"title"`
	Description        *string `json:"description,omitempty"`
	ResponsibleParty   string  `json:"responsible_party"`
	ValidFrom          *string `json:"valid_from,omitempty"`
	ValidUntil         *string `json:"valid_until,omitempty"`
	EvidenceDocumentID *string `json:"evidence_document_id,omitempty"`
	RiskSeverity       string  `json:"risk_severity,omitempty"`
}

type VerifyComplianceRequest struct {
	Status             string  `json:"status"` // COMPLIANT, NON_COMPLIANT, WAIVED
	EvidenceDocumentID *string `json:"evidence_document_id,omitempty"`
	Notes              *string `json:"notes,omitempty"`
}

// ContractPerformanceMetrics aggregates real operational activity connected to this contract
type ContractPerformanceMetrics struct {
	ContractID                      int64     `json:"contract_id"`
	LinkedShipmentsCount            int       `json:"linked_shipments_count"`
	LinkedBookingsCount             int       `json:"linked_bookings_count"`
	LinkedRatesCount                int       `json:"linked_rates_count"`
	LinkedQuotationsCount           int       `json:"linked_quotations_count"`
	OnTimePerformancePercent        float64   `json:"on_time_performance_percent"`
	AverageTransitDays              float64   `json:"average_transit_days"`
	TransitDeviationDays            float64   `json:"transit_deviation_days"`
	CancellationRatePercent         float64   `json:"cancellation_rate_percent"`
	VolumeCommitmentTarget          float64   `json:"volume_commitment_target"`
	VolumeCommitmentActual          float64   `json:"volume_commitment_actual"`
	VolumeCommitmentProgressPercent float64   `json:"volume_commitment_progress_percent"`
	RevenueOrContractValueUtilized  float64   `json:"revenue_or_contract_value_utilized"`
	CalculatedAt                    time.Time `json:"calculated_at"`
}

// ContractIntelligenceSummary provides dashboard-level intelligence
type ContractIntelligenceSummary struct {
	TotalContracts       int `json:"total_contracts"`
	ActiveContracts      int `json:"active_contracts"`
	ExpiringSoonContracts int `json:"expiring_soon_contracts"`
	TotalObligations     int `json:"total_obligations"`
	ActiveObligations    int `json:"active_obligations"`
	OverdueObligations   int `json:"overdue_obligations"`
	BreachedObligations  int `json:"breached_obligations"`
	DueSoonObligations   int `json:"due_soon_obligations"`
	TotalComplianceEvents int `json:"total_compliance_events"`
	OpenComplianceRisks  int `json:"open_compliance_risks"`
	CriticalRisksCount   int `json:"critical_risks_count"`
	WarningRisksCount    int `json:"warning_risks_count"`
}

// ── Contract Document Import & AI-Assisted Contract Creation ─────────────────

type ExtractedAgreementTermDraft struct {
	TermCategory string  `json:"term_category"`
	TermKey      string  `json:"term_key"`
	TermTitle    string  `json:"term_title"`
	TermValue    string  `json:"term_value"`
	ValueType    string  `json:"value_type"`
	Currency     *string `json:"currency,omitempty"`
	IsCritical   bool    `json:"is_critical"`
	Notes        *string `json:"notes,omitempty"`
}

type ExtractedAgreementObligationDraft struct {
	ObligationTitle  string   `json:"obligation_title"`
	ObligationType   string   `json:"obligation_type"`
	PartyResponsible string   `json:"party_responsible"`
	TargetMetric     *string  `json:"target_metric,omitempty"`
	TargetValue      *float64 `json:"target_value,omitempty"`
	MetricUnit       *string  `json:"metric_unit,omitempty"`
	PenaltyTerms     *string  `json:"penalty_terms,omitempty"`
	DueDate          *string  `json:"due_date,omitempty"`
}

type ExtractedContractDraft struct {
	DocumentID           string                              `json:"document_id"`
	FileName             string                              `json:"file_name"`
	ContractName         string                              `json:"contract_name"`
	ContractReference    string                              `json:"contract_reference"`
	ContractType         string                              `json:"contract_type"`
	PartyName            string                              `json:"party_name"`
	PartyType            string                              `json:"party_type"`
	MatchedPartyID       *int64                              `json:"matched_party_id,omitempty"`
	CarrierSCAC          *string                             `json:"carrier_scac,omitempty"`
	TransportMode        string                              `json:"transport_mode"`
	Currency             string                              `json:"currency"`
	ContractValue        *float64                            `json:"contract_value,omitempty"`
	EffectiveDate        *string                             `json:"effective_date,omitempty"`
	ExpiryDate           *string                             `json:"expiry_date,omitempty"`
	PaymentTerms         *string                             `json:"payment_terms,omitempty"`
	FreeDaysOrigin       int                                 `json:"free_days_origin"`
	FreeDaysDestination  int                                 `json:"free_days_destination"`
	TransitTimeDays      *int                                `json:"transit_time_days,omitempty"`
	Description          *string                             `json:"description,omitempty"`
	Notes                *string                             `json:"notes,omitempty"`
	AISummary            string                              `json:"ai_summary"`
	OverallConfidence    int                                 `json:"overall_confidence"`
	FieldConfidences     map[string]int                      `json:"field_confidences"`
	ExtractedTerms       []ExtractedAgreementTermDraft       `json:"extracted_terms"`
	ExtractedObligations []ExtractedAgreementObligationDraft `json:"extracted_obligations"`
	DuplicateWarning     *string                             `json:"duplicate_warning,omitempty"`
}

type ImportContractDocumentResponse struct {
	DocumentID         string                  `json:"document_id"`
	Status             string                  `json:"status"`
	ExtractionStatus   string                  `json:"extraction_status"`
	ExtractedDraft     *ExtractedContractDraft `json:"extracted_draft"`
	CandidateParties   []MatchedPartyCandidate `json:"candidate_parties"`
	DuplicateDetection *ContractDuplicateWarning `json:"duplicate_detection,omitempty"`
}

type MatchedPartyCandidate struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	PartyType string `json:"party_type"`
	SCAC      string `json:"scac,omitempty"`
	Code      string `json:"code,omitempty"`
}

type ContractDuplicateWarning struct {
	IsDuplicate        bool   `json:"is_duplicate"`
	ExistingContractID *int64 `json:"existing_contract_id,omitempty"`
	ExistingReference  string `json:"existing_reference,omitempty"`
	ExistingName       string `json:"existing_name,omitempty"`
	Message            string `json:"message"`
}

type ConfirmContractImportRequest struct {
	DocumentID           string                              `json:"document_id"`
	ContractReference    string                              `json:"contract_reference"`
	ContractName         string                              `json:"contract_name"`
	ContractType         string                              `json:"contract_type"`
	PartyID              int64                               `json:"party_id"`
	PartyName            string                              `json:"party_name"`
	TransportMode        *string                             `json:"transport_mode,omitempty"`
	Status               string                              `json:"status"` // Default DRAFT
	Currency             *string                             `json:"currency,omitempty"`
	ContractValue        *float64                            `json:"contract_value,omitempty"`
	EffectiveDate        *string                             `json:"effective_date,omitempty"`
	ExpiryDate           *string                             `json:"expiry_date,omitempty"`
	Owner                *string                             `json:"owner,omitempty"`
	Description          *string                             `json:"description,omitempty"`
	Notes                *string                             `json:"notes,omitempty"`
	IncludedTerms        []ExtractedAgreementTermDraft       `json:"included_terms,omitempty"`
	IncludedObligations  []ExtractedAgreementObligationDraft `json:"included_obligations,omitempty"`
}



