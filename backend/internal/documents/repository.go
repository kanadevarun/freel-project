package documents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	InsertDocument(ctx context.Context, doc *ShipmentDocument) error
	InsertDocumentTx(ctx context.Context, tx *sqlx.Tx, doc *ShipmentDocument) error
	GetDocumentsByShipment(ctx context.Context, orgID int64, shipmentID int64) ([]*ShipmentDocument, error)
	GetDocumentsByOrg(ctx context.Context, orgID int64) ([]*ShipmentDocument, error)
	GetDocumentByID(ctx context.Context, orgID int64, id string) (*ShipmentDocument, error)
	UpdateDocumentData(ctx context.Context, docID string, orgID int64, extractedData json.RawMessage, rawOcrText string, status string) error
	
	InsertDiscrepancies(ctx context.Context, orgID int64, shipmentID int64, discrepancies []*ShipmentDocumentDiscrepancy) error
	InsertDiscrepanciesTx(ctx context.Context, tx *sqlx.Tx, orgID int64, shipmentID int64, discrepancies []*ShipmentDocumentDiscrepancy) error
	GetDiscrepancies(ctx context.Context, orgID int64, shipmentID int64) ([]*ShipmentDocumentDiscrepancy, error)
	ResolveDiscrepancy(ctx context.Context, orgID int64, id int64, userID int64) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) InsertDocument(ctx context.Context, doc *ShipmentDocument) error {
	query := `
		INSERT INTO shipment_documents (
			org_id, shipment_id, doc_type, s3_key, file_name, file_type, status, extracted_data, raw_ocr_text, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
		) ON DUPLICATE KEY UPDATE
		s3_key = VALUES(s3_key), file_name = VALUES(file_name), status = 'PENDING_VERIFICATION', updated_at = NOW()
	`
	res, err := r.db.ExecContext(ctx, query, doc.OrgID, doc.ShipmentID, doc.DocType, doc.S3Key, doc.FileName, doc.FileType, doc.Status, doc.ExtractedData, doc.RawOcrText)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil && id > 0 {
		doc.ID = fmt.Sprintf("%d", id)
	}
	return nil
}

func (r *repository) GetDocumentsByShipment(ctx context.Context, orgID int64, shipmentID int64) ([]*ShipmentDocument, error) {
	query := `SELECT * FROM shipment_documents WHERE shipment_id = ? AND org_id = ? ORDER BY created_at DESC`
	var list []*ShipmentDocument
	err := r.db.SelectContext(ctx, &list, query, shipmentID, orgID)
	return list, err
}

func (r *repository) GetDocumentsByOrg(ctx context.Context, orgID int64) ([]*ShipmentDocument, error) {
	query := `SELECT * FROM shipment_documents WHERE org_id = ? ORDER BY created_at DESC`
	var list []*ShipmentDocument
	err := r.db.SelectContext(ctx, &list, query, orgID)
	return list, err
}

func (r *repository) GetDocumentByID(ctx context.Context, orgID int64, id string) (*ShipmentDocument, error) {
	query := `SELECT * FROM shipment_documents WHERE id = ? AND org_id = ?`
	var doc ShipmentDocument
	err := r.db.GetContext(ctx, &doc, query, id, orgID)
	return &doc, err
}

func (r *repository) UpdateDocumentData(ctx context.Context, docID string, orgID int64, extractedData json.RawMessage, rawOcrText string, status string) error {
	query := `
		UPDATE shipment_documents
		SET extracted_data = ?, raw_ocr_text = ?, status = ?, updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, extractedData, rawOcrText, status, docID, orgID)
	return err
}

func (r *repository) InsertDocumentTx(ctx context.Context, tx *sqlx.Tx, doc *ShipmentDocument) error {
	query := `
		INSERT INTO shipment_documents (
			org_id, shipment_id, doc_type, s3_key, file_name, file_type, status, extracted_data, raw_ocr_text, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
		) ON DUPLICATE KEY UPDATE
		s3_key = VALUES(s3_key), file_name = VALUES(file_name), status = 'PENDING_VERIFICATION', updated_at = NOW()
	`
	res, err := tx.ExecContext(ctx, query, doc.OrgID, doc.ShipmentID, doc.DocType, doc.S3Key, doc.FileName, doc.FileType, doc.Status, doc.ExtractedData, doc.RawOcrText)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil && id > 0 {
		doc.ID = fmt.Sprintf("%d", id)
	}
	return nil
}

func (r *repository) InsertDiscrepancies(ctx context.Context, orgID int64, shipmentID int64, discrepancies []*ShipmentDocumentDiscrepancy) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = r.InsertDiscrepanciesTx(ctx, tx, orgID, shipmentID, discrepancies)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *repository) InsertDiscrepanciesTx(ctx context.Context, tx *sqlx.Tx, orgID int64, shipmentID int64, discrepancies []*ShipmentDocumentDiscrepancy) error {
	query := `
		INSERT INTO shipment_document_discrepancies (
			org_id, shipment_id, field_name, expected_value, actual_value, source_document, target_document, status
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?
		) ON DUPLICATE KEY UPDATE
		expected_value = VALUES(expected_value),
		actual_value = VALUES(actual_value),
		status = CASE WHEN status = 'RESOLVED' THEN 'RESOLVED' ELSE 'OPEN' END,
		updated_at = NOW()
	`
	for _, d := range discrepancies {
		d.OrgID = orgID
		d.ShipmentID = shipmentID
		_, err := tx.ExecContext(ctx, query, d.OrgID, d.ShipmentID, d.FieldName, d.ExpectedValue, d.ActualValue, d.SourceDocument, d.TargetDocument, d.Status)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *repository) GetDiscrepancies(ctx context.Context, orgID int64, shipmentID int64) ([]*ShipmentDocumentDiscrepancy, error) {
	query := `SELECT * FROM shipment_document_discrepancies WHERE shipment_id = ? AND org_id = ? ORDER BY created_at DESC`
	var list []*ShipmentDocumentDiscrepancy
	err := r.db.SelectContext(ctx, &list, query, shipmentID, orgID)
	return list, err
}

func (r *repository) ResolveDiscrepancy(ctx context.Context, orgID int64, id int64, userID int64) error {
	query := `
		UPDATE shipment_document_discrepancies
		SET status = 'RESOLVED', resolved_by = ?, resolved_at = NOW(), updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, userID, id, orgID)
	return err
}
