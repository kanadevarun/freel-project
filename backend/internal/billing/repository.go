package billing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	CreateInvoice(ctx context.Context, invoice *CustomerInvoice, items []*CustomerInvoiceItem) error
	GetInvoicesByShipment(ctx context.Context, orgID int64, shipmentID int64) ([]*CustomerInvoice, error)
	GetInvoiceByID(ctx context.Context, orgID int64, id string) (*CustomerInvoice, []*CustomerInvoiceItem, error)
	UpdateInvoiceStatus(ctx context.Context, orgID int64, id string, status string) error
	SaveProfitability(ctx context.Context, p *Profitability) error
	GetProfitability(ctx context.Context, orgID int64, shipmentID int64) (*Profitability, error)
	GetShipmentQuoteAndRateEntry(ctx context.Context, orgID int64, shipmentID int64) (buyPrice float64, sellPrice float64, rateEntrySurcharges string, oceanFreight float64, err error)
	GetClosureChecksData(ctx context.Context, orgID int64, shipmentID int64) (shipmentStatus string, docCount int, unverifiedDocs int, carrierInvCount int, unapprovedCarrierInvs int, err error)
	UpdateShipmentStatus(ctx context.Context, orgID int64, shipmentID int64, status string) error
	GetApprovedCarrierInvoiceTotal(ctx context.Context, orgID int64, shipmentID int64) (float64, error)
	GetApprovedCustomerInvoiceTotal(ctx context.Context, orgID int64, shipmentID int64) (float64, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateInvoice(ctx context.Context, invoice *CustomerInvoice, items []*CustomerInvoiceItem) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert invoice
	queryInv := `
		INSERT INTO shipment_customer_invoices (id, org_id, shipment_id, invoice_number, status, due_date, currency, total_amount)
		VALUES (:id, :org_id, :shipment_id, :invoice_number, :status, :due_date, :currency, :total_amount)
	`
	_, err = tx.NamedExecContext(ctx, queryInv, invoice)
	if err != nil {
		return fmt.Errorf("failed to insert customer invoice: %w", err)
	}

	// Insert items
	queryItem := `
		INSERT INTO shipment_customer_invoice_items (org_id, invoice_id, charge_code, description, quantity, unit_price, total_amount, currency)
		VALUES (:org_id, :invoice_id, :charge_code, :description, :quantity, :unit_price, :total_amount, :currency)
	`
	for _, item := range items {
		_, err = tx.NamedExecContext(ctx, queryItem, item)
		if err != nil {
			return fmt.Errorf("failed to insert customer invoice item: %w", err)
		}
	}

	return tx.Commit()
}

func (r *repository) GetInvoicesByShipment(ctx context.Context, orgID int64, shipmentID int64) ([]*CustomerInvoice, error) {
	var invoices []*CustomerInvoice
	query := `
		SELECT id, org_id, shipment_id, invoice_number, status, due_date, currency, total_amount, created_at, updated_at
		FROM shipment_customer_invoices
		WHERE org_id = ? AND shipment_id = ?
		ORDER BY created_at DESC
	`
	err := r.db.SelectContext(ctx, &invoices, query, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	return invoices, nil
}

func (r *repository) GetInvoiceByID(ctx context.Context, orgID int64, id string) (*CustomerInvoice, []*CustomerInvoiceItem, error) {
	var invoice CustomerInvoice
	queryInv := `
		SELECT id, org_id, shipment_id, invoice_number, status, due_date, currency, total_amount, created_at, updated_at
		FROM shipment_customer_invoices
		WHERE org_id = ? AND id = ?
	`
	err := r.db.GetContext(ctx, &invoice, queryInv, orgID, id)
	if err != nil {
		return nil, nil, err
	}

	var items []*CustomerInvoiceItem
	queryItems := `
		SELECT id, org_id, invoice_id, charge_code, description, quantity, unit_price, total_amount, currency, created_at
		FROM shipment_customer_invoice_items
		WHERE org_id = ? AND invoice_id = ?
		ORDER BY id ASC
	`
	err = r.db.SelectContext(ctx, &items, queryItems, orgID, id)
	if err != nil {
		return nil, nil, err
	}

	return &invoice, items, nil
}

func (r *repository) UpdateInvoiceStatus(ctx context.Context, orgID int64, id string, status string) error {
	query := `
		UPDATE shipment_customer_invoices
		SET status = ?, updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	_, err := r.db.ExecContext(ctx, query, status, orgID, id)
	return err
}

func (r *repository) SaveProfitability(ctx context.Context, p *Profitability) error {
	query := `
		INSERT INTO shipment_finance_profitability (
			org_id, shipment_id, quoted_sell_price, quoted_buy_price, actual_carrier_cost,
			additional_charges, actual_revenue, expected_profit, actual_profit,
			expected_margin_pct, actual_margin_pct, variance, profitability_status, updated_at
		) VALUES (
			:org_id, :shipment_id, :quoted_sell_price, :quoted_buy_price, :actual_carrier_cost,
			:additional_charges, :actual_revenue, :expected_profit, :actual_profit,
			:expected_margin_pct, :actual_margin_pct, :variance, :profitability_status, NOW()
		)
		ON DUPLICATE KEY UPDATE
			quoted_sell_price = VALUES(quoted_sell_price),
			quoted_buy_price = VALUES(quoted_buy_price),
			actual_carrier_cost = VALUES(actual_carrier_cost),
			additional_charges = VALUES(additional_charges),
			actual_revenue = VALUES(actual_revenue),
			expected_profit = VALUES(expected_profit),
			actual_profit = VALUES(actual_profit),
			expected_margin_pct = VALUES(expected_margin_pct),
			actual_margin_pct = VALUES(actual_margin_pct),
			variance = VALUES(variance),
			profitability_status = VALUES(profitability_status),
			updated_at = NOW()
	`
	_, err := r.db.NamedExecContext(ctx, query, p)
	return err
}

func (r *repository) GetProfitability(ctx context.Context, orgID int64, shipmentID int64) (*Profitability, error) {
	var p Profitability
	query := `
		SELECT id, org_id, shipment_id, quoted_sell_price, quoted_buy_price, actual_carrier_cost,
		       additional_charges, actual_revenue, expected_profit, actual_profit,
		       expected_margin_pct, actual_margin_pct, variance, profitability_status, updated_at
		FROM shipment_finance_profitability
		WHERE org_id = ? AND shipment_id = ?
	`
	err := r.db.GetContext(ctx, &p, query, orgID, shipmentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *repository) GetShipmentQuoteAndRateEntry(ctx context.Context, orgID int64, shipmentID int64) (buyPrice float64, sellPrice float64, rateEntrySurcharges string, oceanFreight float64, err error) {
	query := `
		SELECT 
			COALESCE(q.buy_price, 0.0) AS buy_price,
			COALESCE(q.sell_price, 0.0) AS sell_price,
			COALESCE(CAST(re.surcharges AS CHAR), '[]') AS rate_surcharges,
			COALESCE(re.ocean_freight, 0.0) AS ocean_freight
		FROM shipments s
		JOIN rfq_quotes q ON s.quote_id = q.id
		LEFT JOIN rate_entries re ON re.org_id = s.org_id
			AND re.origin_port = s.origin_port
			AND re.destination_port = s.destination_port
			AND re.carrier_scac = s.carrier_scac
			AND re.extraction_status = 'CONFIRMED'
		WHERE s.id = ? AND s.org_id = ?
		LIMIT 1
	`
	err = r.db.QueryRowContext(ctx, query, shipmentID, orgID).Scan(&buyPrice, &sellPrice, &rateEntrySurcharges, &oceanFreight)
	return
}

func (r *repository) GetClosureChecksData(ctx context.Context, orgID int64, shipmentID int64) (shipmentStatus string, docCount int, unverifiedDocs int, carrierInvCount int, unapprovedCarrierInvs int, err error) {
	queryShipment := `SELECT status FROM shipments WHERE id = ? AND org_id = ?`
	err = r.db.QueryRowContext(ctx, queryShipment, shipmentID, orgID).Scan(&shipmentStatus)
	if err != nil {
		return
	}

	queryDocs := `
		SELECT 
			COUNT(*),
			SUM(CASE WHEN status != 'VERIFIED' THEN 1 ELSE 0 END)
		FROM shipment_documents
		WHERE shipment_id = ? AND org_id = ?
	`
	var sumUnverified sql.NullInt64
	err = r.db.QueryRowContext(ctx, queryDocs, shipmentID, orgID).Scan(&docCount, &sumUnverified)
	if err != nil {
		return
	}
	if sumUnverified.Valid {
		unverifiedDocs = int(sumUnverified.Int64)
	}

	queryInvoices := `
		SELECT 
			COUNT(*),
			SUM(CASE WHEN status != 'APPROVED' THEN 1 ELSE 0 END)
		FROM shipment_invoices
		WHERE shipment_id = ? AND org_id = ?
	`
	var sumUnapproved sql.NullInt64
	err = r.db.QueryRowContext(ctx, queryInvoices, shipmentID, orgID).Scan(&carrierInvCount, &sumUnapproved)
	if err != nil {
		return
	}
	if sumUnapproved.Valid {
		unapprovedCarrierInvs = int(sumUnapproved.Int64)
	}

	return
}

func (r *repository) UpdateShipmentStatus(ctx context.Context, orgID int64, shipmentID int64, status string) error {
	query := `
		UPDATE shipments
		SET status = ?, updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, status, shipmentID, orgID)
	return err
}

func (r *repository) GetApprovedCarrierInvoiceTotal(ctx context.Context, orgID int64, shipmentID int64) (float64, error) {
	var total sql.NullFloat64
	query := `
		SELECT SUM(total_amount)
		FROM shipment_invoices
		WHERE org_id = ? AND shipment_id = ? AND status = 'APPROVED'
	`
	err := r.db.QueryRowContext(ctx, query, orgID, shipmentID).Scan(&total)
	if err != nil {
		return 0, err
	}
	if total.Valid {
		return total.Float64, nil
	}
	return 0, nil
}

func (r *repository) GetApprovedCustomerInvoiceTotal(ctx context.Context, orgID int64, shipmentID int64) (float64, error) {
	var total sql.NullFloat64
	query := `
		SELECT SUM(total_amount)
		FROM shipment_customer_invoices
		WHERE org_id = ? AND shipment_id = ? AND status IN ('APPROVED', 'SENT', 'PAID')
	`
	err := r.db.QueryRowContext(ctx, query, orgID, shipmentID).Scan(&total)
	if err != nil {
		return 0, err
	}
	if total.Valid {
		return total.Float64, nil
	}
	return 0, nil
}
