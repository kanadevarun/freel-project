package contracts

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Repository encapsulates all database read and write operations for contract documents and rate review items.
type Repository interface {
	CreateDocument(ctx context.Context, doc *ContractDocument) error
	GetDocumentByID(ctx context.Context, orgID int64, id string) (*ContractDocument, error)
	ListDocuments(ctx context.Context, orgID int64, status *ProcessingStatus) ([]ContractDocument, error)
	UpdateDocumentStatus(ctx context.Context, orgID int64, id string, status ProcessingStatus, logs []LogEntry) error
	UpdateDocumentSummaryAndCounts(ctx context.Context, orgID int64, id string, summary string, extracted, confirmed, pending, failed int) error

	CreateReviewItem(ctx context.Context, item *RateReviewItem) error
	GetReviewItemByID(ctx context.Context, orgID int64, id string) (*RateReviewItem, error)
	ListReviewItems(ctx context.Context, orgID int64, status *ReviewStatus) ([]RateReviewItem, error)
	UpdateReviewItemStatus(ctx context.Context, orgID int64, id string, status ReviewStatus, reviewerID int64, correctedData []byte, notes string) error
	CreateAITask(ctx context.Context, orgID int64, docID string, taskType string, payload map[string]interface{}) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateDocument(ctx context.Context, doc *ContractDocument) error {
	logsJSON, err := json.Marshal(doc.ProcessingLog)
	if err != nil {
		return fmt.Errorf("marshal processing log: %w", err)
	}

	const query = `
INSERT INTO contract_documents (
    id, org_id, carrier_scac, carrier_name, file_name, s3_key, file_type,
    file_size_bytes, page_count, status,
    extracted_rate_count, confirmed_rate_count, pending_review_count, failed_rate_count,
    processing_log, created_by, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
)
`
	_, err = r.db.ExecContext(ctx, query,
		doc.ID, doc.OrgID, doc.CarrierSCAC, doc.CarrierName, doc.FileName, doc.S3Key, doc.FileType,
		doc.FileSize, doc.PageCount, doc.Status,
		doc.ExtractedRateCount, doc.ConfirmedRateCount, doc.PendingReviewCount, doc.FailedRateCount,
		logsJSON, doc.CreatedBy,
	)
	return err
}

func (r *repository) GetDocumentByID(ctx context.Context, orgID int64, id string) (*ContractDocument, error) {
	const query = `
SELECT id, org_id, carrier_scac, carrier_name, file_name, s3_key, file_type, file_size_bytes, page_count, status,
    extracted_rate_count, confirmed_rate_count, pending_review_count, failed_rate_count,
    processing_started_at, processing_completed_at, processing_log, ai_document_summary,
    reviewed_by, reviewed_at, review_notes, created_by, created_at, updated_at
FROM contract_documents
WHERE id = ? AND org_id = ?
`
	var doc ContractDocument
	err := r.db.GetContext(ctx, &doc, query, id, orgID)
	if err != nil {
		return nil, err
	}

	if len(doc.ProcessingLogRaw) > 0 {
		_ = json.Unmarshal(doc.ProcessingLogRaw, &doc.ProcessingLog)
	}
	return &doc, nil
}

func (r *repository) ListDocuments(ctx context.Context, orgID int64, status *ProcessingStatus) ([]ContractDocument, error) {
	var query string
	var args []interface{}

	if status != nil {
		query = `
SELECT id, org_id, carrier_scac, carrier_name, file_name, s3_key, file_type, file_size_bytes, page_count, status,
    extracted_rate_count, confirmed_rate_count, pending_review_count, failed_rate_count,
    processing_started_at, processing_completed_at, processing_log, ai_document_summary,
    reviewed_by, reviewed_at, review_notes, created_by, created_at, updated_at
FROM contract_documents
WHERE org_id = ? AND status = ?
ORDER BY created_at DESC
`
		args = []interface{}{orgID, string(*status)}
	} else {
		query = `
SELECT id, org_id, carrier_scac, carrier_name, file_name, s3_key, file_type, file_size_bytes, page_count, status,
    extracted_rate_count, confirmed_rate_count, pending_review_count, failed_rate_count,
    processing_started_at, processing_completed_at, processing_log, ai_document_summary,
    reviewed_by, reviewed_at, review_notes, created_by, created_at, updated_at
FROM contract_documents
WHERE org_id = ?
ORDER BY created_at DESC
`
		args = []interface{}{orgID}
	}

	var docs []ContractDocument
	err := r.db.SelectContext(ctx, &docs, query, args...)
	if err != nil {
		return nil, err
	}

	for i := range docs {
		if len(docs[i].ProcessingLogRaw) > 0 {
			_ = json.Unmarshal(docs[i].ProcessingLogRaw, &docs[i].ProcessingLog)
		}
	}
	return docs, nil
}

func (r *repository) UpdateDocumentStatus(ctx context.Context, orgID int64, id string, status ProcessingStatus, logs []LogEntry) error {
	logsJSON, err := json.Marshal(logs)
	if err != nil {
		return fmt.Errorf("marshal logs: %w", err)
	}

	statusStr := string(status)
	const query = `
UPDATE contract_documents
SET status = ?,
    processing_log = ?,
    processing_started_at = COALESCE(processing_started_at, CASE WHEN ? = 'OCR_PROCESSING' OR ? = 'AI_EXTRACTING' THEN NOW() ELSE processing_started_at END),
    processing_completed_at = CASE WHEN ? = 'CONFIRMED' OR ? = 'PENDING_REVIEW' OR ? = 'FAILED' THEN NOW() ELSE processing_completed_at END,
    updated_at = NOW()
WHERE id = ? AND org_id = ?
`
	_, err = r.db.ExecContext(ctx, query, statusStr, logsJSON, statusStr, statusStr, statusStr, statusStr, statusStr, id, orgID)
	return err
}

func (r *repository) UpdateDocumentSummaryAndCounts(ctx context.Context, orgID int64, id string, summary string, extracted, confirmed, pending, failed int) error {
	const query = `
UPDATE contract_documents
SET ai_document_summary = ?,
    extracted_rate_count = ?,
    confirmed_rate_count = ?,
    pending_review_count = ?,
    failed_rate_count = ?,
    updated_at = NOW()
WHERE id = ? AND org_id = ?
`
	_, err := r.db.ExecContext(ctx, query, summary, extracted, confirmed, pending, failed, id, orgID)
	return err
}

func (r *repository) CreateReviewItem(ctx context.Context, item *RateReviewItem) error {
	flagsJSON, err := json.Marshal(item.ReviewFlags)
	if err != nil {
		return fmt.Errorf("marshal review flags: %w", err)
	}

	const query = `
INSERT INTO rate_review_queue (
    id, org_id, contract_doc_id, extracted_data, confidence_score, review_flags, ai_reasoning,
    source_page, source_text, source_image_url, status, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
)
`
	_, err = r.db.ExecContext(ctx, query,
		item.ID, item.OrgID, item.ContractDocID, item.ExtractedData, item.Confidence, flagsJSON, item.AIReasoning,
		item.SourcePage, item.SourceText, item.SourceImageURL, string(item.Status),
	)
	return err
}

func (r *repository) GetReviewItemByID(ctx context.Context, orgID int64, id string) (*RateReviewItem, error) {
	const query = `
SELECT id, org_id, contract_doc_id, extracted_data, confidence_score, review_flags, ai_reasoning,
    source_page, source_text, source_image_url, status, reviewed_by, reviewed_at, corrected_data, review_notes, created_at, updated_at
FROM rate_review_queue
WHERE id = ? AND org_id = ?
`
	var item RateReviewItem
	err := r.db.GetContext(ctx, &item, query, id, orgID)
	if err != nil {
		return nil, err
	}
	if len(item.ReviewFlagsRaw) > 0 {
		_ = json.Unmarshal(item.ReviewFlagsRaw, &item.ReviewFlags)
	}
	return &item, nil
}

func (r *repository) ListReviewItems(ctx context.Context, orgID int64, status *ReviewStatus) ([]RateReviewItem, error) {
	var query string
	var args []interface{}

	if status != nil {
		query = `
SELECT id, org_id, contract_doc_id, extracted_data, confidence_score, review_flags, ai_reasoning,
    source_page, source_text, source_image_url, status, reviewed_by, reviewed_at, corrected_data, review_notes, created_at, updated_at
FROM rate_review_queue
WHERE org_id = ? AND status = ?
ORDER BY created_at ASC
`
		args = []interface{}{orgID, string(*status)}
	} else {
		query = `
SELECT id, org_id, contract_doc_id, extracted_data, confidence_score, review_flags, ai_reasoning,
    source_page, source_text, source_image_url, status, reviewed_by, reviewed_at, corrected_data, review_notes, created_at, updated_at
FROM rate_review_queue
WHERE org_id = ?
ORDER BY created_at ASC
`
		args = []interface{}{orgID}
	}

	var items []RateReviewItem
	err := r.db.SelectContext(ctx, &items, query, args...)
	if err != nil {
		return nil, err
	}

	for i := range items {
		if len(items[i].ReviewFlagsRaw) > 0 {
			_ = json.Unmarshal(items[i].ReviewFlagsRaw, &items[i].ReviewFlags)
		}
	}
	return items, nil
}

func (r *repository) UpdateReviewItemStatus(ctx context.Context, orgID int64, id string, status ReviewStatus, reviewerID int64, correctedData []byte, notes string) error {
	var notesVal sql.NullString
	if notes != "" {
		notesVal = sql.NullString{String: notes, Valid: true}
	}

	var query string
	var args []interface{}

	if correctedData != nil {
		query = `
UPDATE rate_review_queue
SET status = ?,
    reviewed_by = ?,
    reviewed_at = NOW(),
    corrected_data = ?,
    review_notes = ?,
    updated_at = NOW()
WHERE id = ? AND org_id = ?
`
		args = []interface{}{string(status), reviewerID, correctedData, notesVal, id, orgID}
	} else {
		query = `
UPDATE rate_review_queue
SET status = ?,
    reviewed_by = ?,
    reviewed_at = NOW(),
    review_notes = ?,
    updated_at = NOW()
WHERE id = ? AND org_id = ?
`
		args = []interface{}{string(status), reviewerID, notesVal, id, orgID}
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *repository) CreateAITask(ctx context.Context, orgID int64, docID string, taskType string, payload map[string]interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal task payload: %w", err)
	}

	var docIDVal interface{} = docID
	if docID == "" {
		docIDVal = nil
	}

	entityType := "CONTRACT"
	if val, ok := payload["entity_type"].(string); ok {
		entityType = val
	}

	var entityIDVal interface{} = docID
	if docID == "" {
		entityIDVal = nil
	}
	if val, ok := payload["entity_id"].(string); ok {
		entityIDVal = val
	}

	const query = `
INSERT INTO ai_processing_tasks (
    org_id, document_id, entity_type, entity_id, task_type, payload, status, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, 'QUEUED', NOW(), NOW()
)
`
	_, err = r.db.ExecContext(ctx, query, orgID, docIDVal, entityType, entityIDVal, taskType, payloadJSON)
	return err
}


