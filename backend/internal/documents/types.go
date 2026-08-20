package documents

import (
	"encoding/json"
	"time"
)

type ShipmentDocument struct {
	ID            string          `json:"id" db:"id"`
	OrgID         int64           `json:"org_id" db:"org_id"`
	ShipmentID    int64           `json:"shipment_id" db:"shipment_id"`
	DocType       string          `json:"doc_type" db:"doc_type"`
	S3Key         string          `json:"s3_key" db:"s3_key"`
	FileName      string          `json:"file_name" db:"file_name"`
	FileType      string          `json:"file_type" db:"file_type"`
	Status        string          `json:"status" db:"status"`
	ExtractedData json.RawMessage `json:"extracted_data" db:"extracted_data"`
	RawOcrText    *string         `json:"raw_ocr_text" db:"raw_ocr_text"`
	AISummary     *string         `json:"ai_summary" db:"ai_summary"`
	VerifiedBy    *int64          `json:"verified_by" db:"verified_by"`
	VerifiedAt    *time.Time      `json:"verified_at" db:"verified_at"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

type ShipmentDocumentDiscrepancy struct {
	ID             int64      `json:"id" db:"id"`
	OrgID          int64      `json:"org_id" db:"org_id"`
	ShipmentID     int64      `json:"shipment_id" db:"shipment_id"`
	FieldName      string     `json:"field_name" db:"field_name"`
	ExpectedValue  *string    `json:"expected_value" db:"expected_value"`
	ActualValue    *string    `json:"actual_value" db:"actual_value"`
	SourceDocument string     `json:"source_document" db:"source_document"`
	TargetDocument string     `json:"target_document" db:"target_document"`
	Status         string     `json:"status" db:"status"`
	ResolvedBy     *int64     `json:"resolved_by" db:"resolved_by"`
	ResolvedAt     *time.Time `json:"resolved_at" db:"resolved_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}
