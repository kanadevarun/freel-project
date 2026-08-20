package finance

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

type Service interface {
	IngestInvoice(ctx context.Context, orgID int64, shipmentID int64, invoiceNumber string, vendorName string, s3Key string, fileName string) (*Invoice, error)
	GetInvoicesByShipment(ctx context.Context, orgID int64, shipmentID int64) ([]*Invoice, error)
	GetInvoicesByOrg(ctx context.Context, orgID int64) ([]*Invoice, error)
	GetInvoiceByID(ctx context.Context, orgID int64, id string) (*Invoice, error)
	GetItemsByInvoice(ctx context.Context, orgID int64, invoiceID string) ([]*InvoiceItem, error)
	GetDiscrepanciesByShipment(ctx context.Context, orgID int64, shipmentID int64) ([]*FinanceDiscrepancy, error)
	ResolveDiscrepancy(ctx context.Context, orgID int64, id int64, userID int64) error
	CompleteReconciliation(ctx context.Context, req *FinanceCallbackRequest) error
	GetFinanceWorkspaceInternal(ctx context.Context, orgID int64, shipmentID int64) (map[string]interface{}, error)
	ApproveInvoice(ctx context.Context, orgID int64, id string) error
}

type service struct {
	repo           Repository
	db             *sqlx.DB
	backendBaseURL string // e.g. "http://backend:8080" — no trailing slash
}

func NewService(repo Repository, db *sqlx.DB, backendBaseURL string) Service {
	return &service{repo: repo, db: db, backendBaseURL: backendBaseURL}
}

// IngestInvoice stores the invoice record and atomically queues an INVOICE_AUDIT task
// in the outbox table so the Python FinanceAgent sidecar picks it up.
func (s *service) IngestInvoice(ctx context.Context, orgID int64, shipmentID int64, invoiceNumber string, vendorName string, s3Key string, fileName string) (*Invoice, error) {
	inv := &Invoice{
		OrgID:         orgID,
		ShipmentID:    shipmentID,
		InvoiceNumber: invoiceNumber,
		VendorName:    vendorName,
		S3Key:         s3Key,
		FileName:      fileName,
		Currency:      "USD",
		Status:        "PENDING_RECONCILIATION",
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.repo.InsertInvoiceTx(ctx, tx, inv); err != nil {
		return nil, fmt.Errorf("failed to insert invoice: %w", err)
	}

	// Enqueue BILL_RECONCILE task as part of same transaction (outbox guarantee)
	payload := map[string]interface{}{
		"org_id":         orgID,
		"shipment_id":    shipmentID,
		"invoice_id":     inv.ID,
		"invoice_number": invoiceNumber,
		"vendor_name":    vendorName,
		"s3_key":         s3Key,
		"file_name":      fileName,
		"callback_url":   s.backendBaseURL + "/internal/finance/callback",
	}
	payloadJSON, _ := json.Marshal(payload)

	queryTask := `
		INSERT INTO ai_processing_tasks (org_id, entity_type, entity_id, task_type, payload, status, created_at, updated_at)
		VALUES (?, 'SHIPMENT_INVOICE', ?, 'BILL_RECONCILE', ?, 'QUEUED', NOW(), NOW())
	`
	if _, err := tx.ExecContext(ctx, queryTask, orgID, inv.ID, string(payloadJSON)); err != nil {
		return nil, fmt.Errorf("failed to queue finance audit task: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	log.Printf("[Finance Service] Invoice %s (%s) ingested for shipment %d, BILL_RECONCILE task queued", inv.InvoiceNumber, inv.ID, shipmentID)
	return inv, nil
}

func (s *service) GetInvoicesByShipment(ctx context.Context, orgID int64, shipmentID int64) ([]*Invoice, error) {
	return s.repo.GetInvoicesByShipment(ctx, orgID, shipmentID)
}

func (s *service) GetInvoicesByOrg(ctx context.Context, orgID int64) ([]*Invoice, error) {
	return s.repo.GetInvoicesByOrg(ctx, orgID)
}

func (s *service) GetInvoiceByID(ctx context.Context, orgID int64, id string) (*Invoice, error) {
	return s.repo.GetInvoiceByID(ctx, orgID, id)
}

func (s *service) GetItemsByInvoice(ctx context.Context, orgID int64, invoiceID string) ([]*InvoiceItem, error) {
	return s.repo.GetItemsByInvoice(ctx, orgID, invoiceID)
}

func (s *service) GetDiscrepanciesByShipment(ctx context.Context, orgID int64, shipmentID int64) ([]*FinanceDiscrepancy, error) {
	return s.repo.GetDiscrepanciesByShipment(ctx, orgID, shipmentID)
}

func (s *service) ResolveDiscrepancy(ctx context.Context, orgID int64, id int64, userID int64) error {
	return s.repo.ResolveDiscrepancy(ctx, orgID, id, userID)
}

// CompleteReconciliation is called by the internal callback handler when the FinanceAgent finishes.
// It atomically saves invoice status, line items, discrepancies, and optionally flags the shipment.
func (s *service) CompleteReconciliation(ctx context.Context, req *FinanceCallbackRequest) error {
	if req.OrgID <= 0 || req.ShipmentID <= 0 || req.InvoiceID == "" {
		return fmt.Errorf("missing required fields in finance callback request")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update invoice status and AI summary
	if err := s.repo.UpdateInvoiceStatusTx(ctx, tx, req.OrgID, req.InvoiceID, req.Status, req.AISummary); err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	// 2. Persist extracted line items
	for i := range req.Items {
		req.Items[i].OrgID = req.OrgID
		req.Items[i].InvoiceID = req.InvoiceID
	}
	if len(req.Items) > 0 {
		if err := s.repo.InsertInvoiceItemsTx(ctx, tx, req.Items); err != nil {
			return fmt.Errorf("failed to insert invoice items: %w", err)
		}
	}

	// 3. Persist discrepancies (idempotent upsert — preserves RESOLVED status)
	for i := range req.Discrepancies {
		req.Discrepancies[i].OrgID = req.OrgID
		req.Discrepancies[i].ShipmentID = req.ShipmentID
		req.Discrepancies[i].InvoiceID = req.InvoiceID
	}
	if len(req.Discrepancies) > 0 {
		if err := s.repo.InsertDiscrepanciesTx(ctx, tx, req.Discrepancies); err != nil {
			return fmt.Errorf("failed to insert finance discrepancies: %w", err)
		}
	}

	// 4. Preserve shipment status; only escalate to EXCEPTION if discrepancies found
	if len(req.Discrepancies) > 0 {
		_, _ = tx.ExecContext(ctx,
			`UPDATE shipments SET status = 'EXCEPTION', updated_at = NOW() WHERE id = ? AND org_id = ?`,
			req.ShipmentID, req.OrgID,
		)
	}

	return tx.Commit()
}

func (s *service) GetFinanceWorkspaceInternal(ctx context.Context, orgID int64, shipmentID int64) (map[string]interface{}, error) {
	invoices, err := s.repo.GetInvoicesByShipment(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}

	discrepancies, err := s.repo.GetDiscrepanciesByShipment(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}

	var shipment struct {
		ID              int64  `db:"id" json:"id"`
		OrgID           int64  `db:"org_id" json:"org_id"`
		RFQID           *int64 `db:"rfq_id" json:"rfq_id"`
		QuoteID         *int64 `db:"quote_id" json:"quote_id"`
		CarrierSCAC     string `db:"carrier_scac" json:"carrier_scac"`
		OriginPort      string `db:"origin_port" json:"origin_port"`
		DestinationPort string `db:"destination_port" json:"destination_port"`
		Status          string `db:"status" json:"status"`
	}

	err = s.db.GetContext(ctx, &shipment,
		`SELECT id, org_id, rfq_id, quote_id, carrier_scac, origin_port, destination_port, status
		 FROM shipments WHERE id = ? AND org_id = ?`,
		shipmentID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shipment: %w", err)
	}

	var quote *struct {
		ID          int32   `db:"id" json:"id"`
		RFQID       int32   `db:"rfq_id" json:"rfq_id"`
		CarrierName string  `db:"carrier_name" json:"carrier_name"`
		BuyPrice    float64 `db:"buy_price" json:"buy_price"`
		SellPrice   float64 `db:"sell_price" json:"sell_price"`
		Status      string  `db:"status" json:"status"`
	}

	if shipment.QuoteID != nil {
		var q struct {
			ID          int32   `db:"id" json:"id"`
			RFQID       int32   `db:"rfq_id" json:"rfq_id"`
			CarrierName string  `db:"carrier_name" json:"carrier_name"`
			BuyPrice    float64 `db:"buy_price" json:"buy_price"`
			SellPrice   float64 `db:"sell_price" json:"sell_price"`
			Status      string  `db:"status" json:"status"`
		}
		err = s.db.GetContext(ctx, &q,
			`SELECT id, rfq_id, carrier_name, buy_price, sell_price, status
			 FROM rfq_quotes WHERE id = ?`,
			*shipment.QuoteID,
		)
		if err == nil {
			quote = &q
		}
	}

	type DBRateEntry struct {
		ID              string  `db:"id" json:"id"`
		Source          string  `db:"source" json:"source"`
		OriginPort      string  `db:"origin_port" json:"origin_port"`
		DestinationPort string  `db:"destination_port" json:"destination_port"`
		CarrierSCAC     string  `db:"carrier_scac" json:"carrier_scac"`
		OceanFreight    float64 `db:"ocean_freight" json:"ocean_freight"`
		Surcharges      string  `db:"surcharges" json:"surcharges"`
	}
	var rateEntries []DBRateEntry
	err = s.db.SelectContext(ctx, &rateEntries,
		`SELECT id, source, origin_port, destination_port, carrier_scac, ocean_freight, CAST(surcharges AS CHAR) AS surcharges
		 FROM rate_entries
		 WHERE org_id = ? AND origin_port = ? AND destination_port = ? AND carrier_scac = ? AND extraction_status = 'CONFIRMED'`,
		orgID, shipment.OriginPort, shipment.DestinationPort, shipment.CarrierSCAC,
	)
	if err != nil {
		log.Printf("[Finance Service] Rate entries search error: %v", err)
	}

	return map[string]interface{}{
		"invoices":       invoices,
		"discrepancies":  discrepancies,
		"shipment":       shipment,
		"quote":          quote,
		"contract_rates": rateEntries,
	}, nil
}

func (s *service) ApproveInvoice(ctx context.Context, orgID int64, id string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = s.repo.UpdateInvoiceStatusTx(ctx, tx, orgID, id, "APPROVED", "Invoice manually approved by user.")
	if err != nil {
		return err
	}

	return tx.Commit()
}
