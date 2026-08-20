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

