package documents

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/freel/backend/internal/files"
	"github.com/jmoiron/sqlx"
)

type Service interface {
	UploadDocument(ctx context.Context, orgID int64, shipmentID int64, docType string, s3Key string, fileName string, fileType string) (*ShipmentDocument, error)
	UploadGeneralDocument(ctx context.Context, orgID int64, doc *ShipmentDocument, fileReader io.Reader) (*ShipmentDocument, error)
	GetDocumentsByShipment(ctx context.Context, orgID int64, shipmentID int64) ([]*ShipmentDocument, error)
	GetDocumentsByOrg(ctx context.Context, orgID int64) ([]*ShipmentDocument, error)
	GetDiscrepancies(ctx context.Context, orgID int64, shipmentID int64) ([]*ShipmentDocumentDiscrepancy, error)
	ResolveDiscrepancy(ctx context.Context, orgID int64, id int64, userID int64) error
	DeleteDocument(ctx context.Context, orgID int64, id string) error

	CompleteVerification(ctx context.Context, orgID int64, shipmentID int64, docStatusList map[string]string, discrepancies []*ShipmentDocumentDiscrepancy) error
}

type service struct {
	repo           Repository
	db             *sqlx.DB
	backendBaseURL string // e.g. "http://backend:8080" — no trailing slash
	filesSvc       files.Service
}

func NewService(repo Repository, db *sqlx.DB, backendBaseURL string, filesSvc files.Service) Service {
	return &service{
		repo:           repo,
		db:             db,
		backendBaseURL: backendBaseURL,
		filesSvc:       filesSvc,
	}
}

func (s *service) UploadGeneralDocument(ctx context.Context, orgID int64, doc *ShipmentDocument, fileReader io.Reader) (*ShipmentDocument, error) {
	doc.OrgID = orgID
	if doc.Status == "" {
		doc.Status = "VERIFIED"
	}
	if len(doc.ExtractedData) == 0 {
		doc.ExtractedData = json.RawMessage("{}")
	}

	if fileReader != nil && s.filesSvc != nil {
		s3Key, err := s.filesSvc.UploadFile(ctx, doc.FileName, fileReader)
		if err != nil {
			return nil, fmt.Errorf("failed to save document file: %w", err)
		}
		doc.S3Key = s3Key
		filePath, _ := s.filesSvc.GetFileURL(ctx, s3Key)
		doc.FilePath = &filePath
	} else if doc.S3Key == "" {
		doc.S3Key = "docs/" + doc.FileName
	}

	err := s.repo.InsertDocument(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}

	log.Printf("[Documents Service] Uploaded general document %s (%s) for org %d", doc.FileName, doc.DocType, orgID)
	return doc, nil
}

func (s *service) UploadDocument(ctx context.Context, orgID int64, shipmentID int64, docType string, s3Key string, fileName string, fileType string) (*ShipmentDocument, error) {
	doc := &ShipmentDocument{
		OrgID:         orgID,
		ShipmentID:    &shipmentID,
		DocType:       docType,
		S3Key:         s3Key,
		FileName:      fileName,
		FileType:      fileType,
		Status:        "PENDING_VERIFICATION",
		ExtractedData: json.RawMessage("{}"),
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	err = s.repo.InsertDocumentTx(ctx, tx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}

	// Queue DOC_VERIFY task in ai_processing_tasks inside transactional block (idempotency, outbox atomic guarantee)
	payload := map[string]interface{}{
		"org_id":       orgID,
		"shipment_id":  shipmentID,
		"doc_id":       doc.ID,
		"doc_type":     docType,
		"s3_key":       s3Key,
		"file_name":    fileName,
		"callback_url": s.backendBaseURL + "/internal/compliance/callback",
	}
	payloadJSON, _ := json.Marshal(payload)

	queryTask := `
		INSERT INTO ai_processing_tasks (org_id, entity_type, entity_id, task_type, payload, status, created_at, updated_at)
		VALUES (?, 'SHIPMENT_DOCUMENT', ?, 'DOC_VERIFY', ?, 'QUEUED', NOW(), NOW())
	`
	_, err = tx.ExecContext(ctx, queryTask, orgID, doc.ID, string(payloadJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to queue compliance task: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	log.Printf("[Documents Service] Uploaded document %s (%s) for shipment %d, enqueued compliance verify task", fileName, docType, shipmentID)
	return doc, nil
}

func (s *service) GetDocumentsByShipment(ctx context.Context, orgID int64, shipmentID int64) ([]*ShipmentDocument, error) {
	return s.repo.GetDocumentsByShipment(ctx, orgID, shipmentID)
}

func (s *service) GetDocumentsByOrg(ctx context.Context, orgID int64) ([]*ShipmentDocument, error) {
	return s.repo.GetDocumentsByOrg(ctx, orgID)
}

func (s *service) GetDiscrepancies(ctx context.Context, orgID int64, shipmentID int64) ([]*ShipmentDocumentDiscrepancy, error) {
	return s.repo.GetDiscrepancies(ctx, orgID, shipmentID)
}

func (s *service) ResolveDiscrepancy(ctx context.Context, orgID int64, id int64, userID int64) error {
	return s.repo.ResolveDiscrepancy(ctx, orgID, id, userID)
}

func (s *service) DeleteDocument(ctx context.Context, orgID int64, id string) error {
	return s.repo.DeleteDocument(ctx, orgID, id)
}

// CompleteVerification is called by the internal callback handler when ComplianceAgent finishes validation.
func (s *service) CompleteVerification(ctx context.Context, orgID int64, shipmentID int64, docStatusList map[string]string, discrepancies []*ShipmentDocumentDiscrepancy) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update each document's status and extracted data
	for docID, ocrPayloadRaw := range docStatusList {
		var ocrData struct {
			Status        string          `json:"status"`
			ExtractedData json.RawMessage `json:"extracted_data"`
			RawOcrText    string          `json:"raw_ocr_text"`
		}
		if err := json.Unmarshal([]byte(ocrPayloadRaw), &ocrData); err == nil {
			queryDoc := `
				UPDATE shipment_documents
				SET status = ?, extracted_data = ?, raw_ocr_text = ?, updated_at = NOW()
				WHERE id = ? AND org_id = ?
			`
			_, _ = tx.ExecContext(ctx, queryDoc, ocrData.Status, ocrData.ExtractedData, ocrData.RawOcrText, docID, orgID)
		}
	}

	// Insert discrepancy records transactionally
	err = s.repo.InsertDiscrepanciesTx(ctx, tx, orgID, shipmentID, discrepancies)
	if err != nil {
		return fmt.Errorf("failed to insert discrepancies: %w", err)
	}

	// 2. Preserve existing shipment status instead of regressing to IN_TRANSIT
	if len(discrepancies) > 0 {
		queryShipment := `UPDATE shipments SET status = 'EXCEPTION', updated_at = NOW() WHERE id = ? AND org_id = ?`
		_, _ = tx.ExecContext(ctx, queryShipment, shipmentID, orgID)
	}

	return tx.Commit()
}
