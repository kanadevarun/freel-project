package finance

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	InsertInvoiceTx(ctx context.Context, tx *sqlx.Tx, inv *Invoice) error
	InsertInvoiceItemsTx(ctx context.Context, tx *sqlx.Tx, items []InvoiceItem) error
	GetInvoicesByShipment(ctx context.Context, orgID int64, shipmentID int64) ([]*Invoice, error)
	GetInvoicesByOrg(ctx context.Context, orgID int64) ([]*Invoice, error)
	GetInvoiceByID(ctx context.Context, orgID int64, id string) (*Invoice, error)
	UpdateInvoiceStatusTx(ctx context.Context, tx *sqlx.Tx, orgID int64, id string, status string, summary string) error
	GetItemsByInvoice(ctx context.Context, orgID int64, invoiceID string) ([]*InvoiceItem, error)

	InsertDiscrepanciesTx(ctx context.Context, tx *sqlx.Tx, items []*FinanceDiscrepancy) error
	GetDiscrepanciesByShipment(ctx context.Context, orgID int64, shipmentID int64) ([]*FinanceDiscrepancy, error)
	GetDiscrepanciesByInvoice(ctx context.Context, orgID int64, invoiceID string) ([]*FinanceDiscrepancy, error)
	ResolveDiscrepancy(ctx context.Context, orgID int64, id int64, userID int64) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) InsertInvoiceTx(ctx context.Context, tx *sqlx.Tx, inv *Invoice) error {
	query := `
		INSERT INTO shipment_invoices (
			org_id, shipment_id, invoice_number, vendor_name, vendor_ref,
			s3_key, file_name, currency, total_amount, status, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, NOW(), NOW()
		) ON DUPLICATE KEY UPDATE
		s3_key = VALUES(s3_key), file_name = VALUES(file_name),
		status = 'PENDING_RECONCILIATION', updated_at = NOW()
	`
	res, err := tx.ExecContext(ctx, query,
		inv.OrgID, inv.ShipmentID, inv.InvoiceNumber, inv.VendorName, inv.VendorRef,
		inv.S3Key, inv.FileName, inv.Currency, inv.TotalAmount, inv.Status,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil && id > 0 {
		inv.ID = fmt.Sprintf("%d", id)
	}
	return nil
}

func (r *repository) InsertInvoiceItemsTx(ctx context.Context, tx *sqlx.Tx, items []InvoiceItem) error {
	query := `
		INSERT INTO shipment_invoice_items (
			org_id, invoice_id, charge_code, description, quantity, unit_price, total_amount, currency
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?
		)
	`
	for _, item := range items {
		if _, err := tx.ExecContext(ctx, query, item.OrgID, item.InvoiceID, item.ChargeCode, item.Description, item.Quantity, item.UnitPrice, item.TotalAmount, item.Currency); err != nil {
			return err
		}
	}
	return nil
}

func (r *repository) GetInvoicesByShipment(ctx context.Context, orgID int64, shipmentID int64) ([]*Invoice, error) {
	var list []*Invoice
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM shipment_invoices WHERE shipment_id = ? AND org_id = ? ORDER BY created_at DESC`,
		shipmentID, orgID,
	)
	return list, err
}

func (r *repository) GetInvoicesByOrg(ctx context.Context, orgID int64) ([]*Invoice, error) {
	var list []*Invoice
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM shipment_invoices WHERE org_id = ? ORDER BY created_at DESC`,
		orgID,
	)
	return list, err
}

func (r *repository) GetInvoiceByID(ctx context.Context, orgID int64, id string) (*Invoice, error) {
	var inv Invoice
	err := r.db.GetContext(ctx, &inv,
		`SELECT * FROM shipment_invoices WHERE id = ? AND org_id = ?`,
		id, orgID,
	)
	return &inv, err
}

func (r *repository) UpdateInvoiceStatusTx(ctx context.Context, tx *sqlx.Tx, orgID int64, id string, status string, summary string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE shipment_invoices SET status = ?, updated_at = NOW() WHERE id = ? AND org_id = ?`,
		status, id, orgID,
	)
	return err
}

func (r *repository) GetItemsByInvoice(ctx context.Context, orgID int64, invoiceID string) ([]*InvoiceItem, error) {
	var list []*InvoiceItem
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM shipment_invoice_items WHERE invoice_id = ?`,
		invoiceID,
	)
	return list, err
}

func (r *repository) InsertDiscrepanciesTx(ctx context.Context, tx *sqlx.Tx, items []*FinanceDiscrepancy) error {
	query := `
		INSERT INTO shipment_finance_discrepancies (
			invoice_id, charge_code, field_name,
			expected_value, actual_value, source, status, created_at, updated_at
		) VALUES (
			?, ?, ?,
			?, ?, ?, ?, NOW(), NOW()
		) ON DUPLICATE KEY UPDATE
		expected_value = VALUES(expected_value),
		actual_value = VALUES(actual_value),
		status = CASE WHEN status = 'RESOLVED' THEN 'RESOLVED' ELSE 'OPEN' END,
		updated_at = NOW()
	`
	for _, d := range items {
		if _, err := tx.ExecContext(ctx, query, d.InvoiceID, d.ChargeCode, d.FieldName, d.ExpectedValue, d.ActualValue, d.Source, d.Status); err != nil {
			return err
		}
	}
	return nil
}

func (r *repository) GetDiscrepanciesByShipment(ctx context.Context, orgID int64, shipmentID int64) ([]*FinanceDiscrepancy, error) {
	var list []*FinanceDiscrepancy
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM shipment_finance_discrepancies WHERE shipment_id = ? AND org_id = ? ORDER BY created_at DESC`,
		shipmentID, orgID,
	)
	return list, err
}

func (r *repository) GetDiscrepanciesByInvoice(ctx context.Context, orgID int64, invoiceID string) ([]*FinanceDiscrepancy, error) {
	var list []*FinanceDiscrepancy
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM shipment_finance_discrepancies WHERE invoice_id = ? AND org_id = ? ORDER BY created_at DESC`,
		invoiceID, orgID,
	)
	return list, err
}

func (r *repository) ResolveDiscrepancy(ctx context.Context, orgID int64, id int64, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE shipment_finance_discrepancies
		 SET status = 'RESOLVED', resolved_by = ?, resolved_at = NOW(), updated_at = NOW()
		 WHERE id = ? AND org_id = ?`,
		userID, id, orgID,
	)
	return err
}
