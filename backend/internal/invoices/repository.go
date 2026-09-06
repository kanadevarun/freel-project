package invoices

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetInvoices(ctx context.Context, orgID int64, params ListInvoiceParams, currentUserID int64) ([]*Invoice, int, error)
	GetInvoiceByID(ctx context.Context, orgID int64, id int64) (*Invoice, error)
	GetKPIStats(ctx context.Context, orgID int64) (*InvoiceKPIStats, error)
	CreateInvoice(ctx context.Context, inv *Invoice, items []InvoiceItem) (*Invoice, error)
	UpdateInvoice(ctx context.Context, inv *Invoice, items []InvoiceItem) (*Invoice, error)
	UpdateInvoiceStatus(ctx context.Context, orgID int64, id int64, status string) error
	ToggleBookmark(ctx context.Context, orgID int64, id int64) (bool, error)
	GetInvoiceItems(ctx context.Context, orgID int64, invoiceID int64) ([]InvoiceItem, error)
	GetInvoicePayments(ctx context.Context, orgID int64, invoiceID int64) ([]InvoicePayment, error)
	GetInvoiceDocuments(ctx context.Context, orgID int64, invoiceID int64) ([]InvoiceDocument, error)
	GetInvoiceHistory(ctx context.Context, orgID int64, invoiceID int64) ([]InvoiceHistory, error)
	AddHistory(ctx context.Context, orgID int64, invoiceID int64, title string, description string, userName string) error
	CreateApprovalRequestForInvoice(ctx context.Context, inv *Invoice, userName string) error
	RecordPayment(ctx context.Context, orgID int64, invoiceID int64, payment *InvoicePayment, newPaidAmount float64, newBalanceDue float64, newStatus string, historyTitle string, historyDesc string, userName string) (*Invoice, error)
	GetAllPayments(ctx context.Context, orgID int64) ([]InvoicePayment, error)
	AddDocument(ctx context.Context, orgID int64, invoiceID int64, doc *InvoiceDocument) error
	GenerateInvoiceNumber(ctx context.Context, orgID int64) (string, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetInvoices(ctx context.Context, orgID int64, params ListInvoiceParams, currentUserID int64) ([]*Invoice, int, error) {
	whereClauses := []string{"org_id = ?"}
	args := []interface{}{orgID}

	if params.PrimaryTab == "MY" && currentUserID > 0 {
		whereClauses = append(whereClauses, "(created_by_id = ? OR is_my_invoice = true)")
		args = append(args, currentUserID)
	}

	if params.CustomerID > 0 {
		whereClauses = append(whereClauses, "customer_id = ?")
		args = append(args, params.CustomerID)
	}

	if params.ShipmentID > 0 {
		whereClauses = append(whereClauses, "shipment_id = ?")
		args = append(args, params.ShipmentID)
	}

	if params.BookingID > 0 {
		whereClauses = append(whereClauses, "booking_id = ?")
		args = append(args, params.BookingID)
	}

	if params.QuotationID > 0 {
		whereClauses = append(whereClauses, "quotation_id = ?")
		args = append(args, params.QuotationID)
	}

	if params.Status != "" && params.Status != "All" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, params.Status)
	}

	if params.Search != "" {
		q := "%" + strings.TrimSpace(params.Search) + "%"
		whereClauses = append(whereClauses, "(invoice_number LIKE ? OR customer_name LIKE ? OR shipment_number LIKE ? OR route LIKE ?)")
		args = append(args, q, q, q, q)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM customer_invoices WHERE %s", whereSQL)
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	selectQuery := fmt.Sprintf("SELECT * FROM customer_invoices WHERE %s ORDER BY created_at DESC LIMIT %d OFFSET %d", whereSQL, pageSize, offset)
	var list []*Invoice
	if err := r.db.SelectContext(ctx, &list, selectQuery, args...); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *repository) GetInvoiceByID(ctx context.Context, orgID int64, id int64) (*Invoice, error) {
	var inv Invoice
	err := r.db.GetContext(ctx, &inv, "SELECT * FROM customer_invoices WHERE id = ? AND org_id = ?", id, orgID)
	if err != nil {
		return nil, err
	}

	// Fetch related items, payments, docs, history
	items, _ := r.GetInvoiceItems(ctx, orgID, id)
	inv.LineItems = items

	payments, _ := r.GetInvoicePayments(ctx, orgID, id)
	inv.Payments = payments

	docs, _ := r.GetInvoiceDocuments(ctx, orgID, id)
	inv.Documents = docs

	hist, _ := r.GetInvoiceHistory(ctx, orgID, id)
	inv.History = hist

	return &inv, nil
}

func (r *repository) GetKPIStats(ctx context.Context, orgID int64) (*InvoiceKPIStats, error) {
	type Aggregate struct {
		TotalCount       int     `db:"total_count"`
		TotalSum         float64 `db:"total_sum"`
		OutstandingCount int     `db:"outstanding_count"`
		OutstandingSum   float64 `db:"outstanding_sum"`
		PaidCount        int     `db:"paid_count"`
		PaidSum          float64 `db:"paid_sum"`
		OverdueCount     int     `db:"overdue_count"`
		OverdueSum       float64 `db:"overdue_sum"`
	}

	query := `
		SELECT 
			COUNT(*) AS total_count,
			COALESCE(SUM(total_amount), 0) AS total_sum,
			COALESCE(SUM(CASE WHEN balance_due > 0 AND status != 'Cancelled' THEN 1 ELSE 0 END), 0) AS outstanding_count,
			COALESCE(SUM(CASE WHEN balance_due > 0 AND status != 'Cancelled' THEN balance_due ELSE 0 END), 0) AS outstanding_sum,
			COALESCE(SUM(CASE WHEN status = 'Paid' THEN 1 ELSE 0 END), 0) AS paid_count,
			COALESCE(SUM(CASE WHEN status = 'Paid' THEN total_amount ELSE 0 END), 0) AS paid_sum,
			COALESCE(SUM(CASE WHEN status = 'Overdue' OR (due_date < CURRENT_DATE() AND balance_due > 0 AND status != 'Cancelled') THEN 1 ELSE 0 END), 0) AS overdue_count,
			COALESCE(SUM(CASE WHEN status = 'Overdue' OR (due_date < CURRENT_DATE() AND balance_due > 0 AND status != 'Cancelled') THEN balance_due ELSE 0 END), 0) AS overdue_sum
		FROM customer_invoices
		WHERE org_id = ?
	`

	var agg Aggregate
	if err := r.db.GetContext(ctx, &agg, query, orgID); err != nil {
		return nil, err
	}

	formatMillions := func(val float64) string {
		if val >= 1000000 {
			return fmt.Sprintf("$%.2fM", val/1000000.0)
		}
		if val >= 1000 {
			return fmt.Sprintf("$%.2fK", val/1000.0)
		}
		return fmt.Sprintf("$%.2f", val)
	}

	formatVal := func(val float64) string {
		return fmt.Sprintf("$%.2f", val)
	}

	return &InvoiceKPIStats{
		TotalInvoices: KPICardMetric{
			Amount:         formatVal(agg.TotalSum),
			DisplayAmount:  formatMillions(agg.TotalSum),
			Count:          agg.TotalCount,
			Label:          fmt.Sprintf("%d Invoices", agg.TotalCount),
			Trend:          "18.6%",
			TrendDirection: "up",
			TrendPeriod:    "vs last 7 days",
		},
		Outstanding: KPICardMetric{
			Amount:         formatVal(agg.OutstandingSum),
			DisplayAmount:  formatMillions(agg.OutstandingSum),
			Count:          agg.OutstandingCount,
			Label:          fmt.Sprintf("%d Invoices", agg.OutstandingCount),
			Trend:          "12.4%",
			TrendDirection: "up",
			TrendPeriod:    "vs last 7 days",
		},
		PaidThisMonth: KPICardMetric{
			Amount:         formatVal(agg.PaidSum),
			DisplayAmount:  formatMillions(agg.PaidSum),
			Count:          agg.PaidCount,
			Label:          fmt.Sprintf("%d Invoices", agg.PaidCount),
			Trend:          "24.8%",
			TrendDirection: "up",
			TrendPeriod:    "vs last 7 days",
		},
		Overdue: KPICardMetric{
			Amount:         formatVal(agg.OverdueSum),
			DisplayAmount:  formatMillions(agg.OverdueSum),
			Count:          agg.OverdueCount,
			Label:          fmt.Sprintf("%d Invoices", agg.OverdueCount),
			Trend:          "8.2%",
			TrendDirection: "up",
			TrendPeriod:    "vs last 7 days",
		},
	}, nil
}

func (r *repository) CreateInvoice(ctx context.Context, inv *Invoice, items []InvoiceItem) (*Invoice, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO customer_invoices (
			org_id, invoice_number, customer_id, customer_name, customer_country,
			shipment_id, shipment_number, booking_id, booking_number, quotation_id, quote_number,
			route, origin, destination, invoice_date, due_date, days_left, currency,
			subtotal, tax_amount, discount_amount, total_amount, paid_amount, balance_due,
			status, type, bookmarked, is_my_invoice, creator_name, created_by_id, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, NOW(), NOW()
		)
	`
	res, err := tx.ExecContext(ctx, query,
		inv.OrgID, inv.InvoiceNumber, inv.CustomerID, inv.CustomerName, inv.CustomerCountry,
		inv.ShipmentID, inv.ShipmentNumber, inv.BookingID, inv.BookingNumber, inv.QuotationID, inv.QuoteNumber,
		inv.Route, inv.Origin, inv.Destination, inv.InvoiceDate, inv.DueDate, inv.DaysLeft, inv.Currency,
		inv.Subtotal, inv.TaxAmount, inv.DiscountAmount, inv.TotalAmount, inv.PaidAmount, inv.BalanceDue,
		inv.Status, inv.Type, inv.Bookmarked, inv.IsMyInvoice, inv.CreatorName, inv.CreatedByID,
	)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	inv.ID = id

	// Insert line items
	for i, item := range items {
		item.OrgID = inv.OrgID
		item.InvoiceID = inv.ID
		item.DisplayOrder = i + 1
		itemQuery := `
			INSERT INTO customer_invoice_items (
				org_id, invoice_id, description, service_category, quantity, unit_price, total_amount, display_order, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())
		`
		if _, err := tx.ExecContext(ctx, itemQuery, item.OrgID, item.InvoiceID, item.Description, item.ServiceCategory, item.Quantity, item.UnitPrice, item.TotalAmount, item.DisplayOrder); err != nil {
			return nil, err
		}
	}

	// Insert initial history item
	histQuery := `
		INSERT INTO customer_invoice_history (org_id, invoice_id, title, description, user_name, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())
	`
	_, _ = tx.ExecContext(ctx, histQuery, inv.OrgID, inv.ID, "Draft Created", "Invoice created", inv.CreatorName)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetInvoiceByID(ctx, inv.OrgID, inv.ID)
}

func (r *repository) UpdateInvoice(ctx context.Context, inv *Invoice, items []InvoiceItem) (*Invoice, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		UPDATE customer_invoices SET
			customer_id = ?, customer_name = ?, customer_country = ?,
			shipment_id = ?, shipment_number = ?, booking_id = ?, booking_number = ?, quotation_id = ?, quote_number = ?,
			route = ?, origin = ?, destination = ?, invoice_date = ?, due_date = ?, days_left = ?, currency = ?,
			subtotal = ?, tax_amount = ?, discount_amount = ?, total_amount = ?, balance_due = ?,
			status = ?, updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err = tx.ExecContext(ctx, query,
		inv.CustomerID, inv.CustomerName, inv.CustomerCountry,
		inv.ShipmentID, inv.ShipmentNumber, inv.BookingID, inv.BookingNumber, inv.QuotationID, inv.QuoteNumber,
		inv.Route, inv.Origin, inv.Destination, inv.InvoiceDate, inv.DueDate, inv.DaysLeft, inv.Currency,
		inv.Subtotal, inv.TaxAmount, inv.DiscountAmount, inv.TotalAmount, inv.BalanceDue,
		inv.Status, inv.ID, inv.OrgID,
	)
	if err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM customer_invoice_items WHERE invoice_id = ? AND org_id = ?", inv.ID, inv.OrgID); err != nil {
		return nil, err
	}

	for i, item := range items {
		item.OrgID = inv.OrgID
		item.InvoiceID = inv.ID
		item.DisplayOrder = i + 1
		itemQuery := `
			INSERT INTO customer_invoice_items (
				org_id, invoice_id, description, service_category, quantity, unit_price, total_amount, display_order, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())
		`
		if _, err := tx.ExecContext(ctx, itemQuery, item.OrgID, item.InvoiceID, item.Description, item.ServiceCategory, item.Quantity, item.UnitPrice, item.TotalAmount, item.DisplayOrder); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetInvoiceByID(ctx, inv.OrgID, inv.ID)
}

func (r *repository) UpdateInvoiceStatus(ctx context.Context, orgID int64, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE customer_invoices SET status = ?, updated_at = NOW() WHERE id = ? AND org_id = ?", status, id, orgID)
	return err
}

func (r *repository) ToggleBookmark(ctx context.Context, orgID int64, id int64) (bool, error) {
	var current bool
	err := r.db.GetContext(ctx, &current, "SELECT bookmarked FROM customer_invoices WHERE id = ? AND org_id = ?", id, orgID)
	if err != nil {
		return false, err
	}
	newVal := !current
	_, err = r.db.ExecContext(ctx, "UPDATE customer_invoices SET bookmarked = ? WHERE id = ? AND org_id = ?", newVal, id, orgID)
	return newVal, err
}

func (r *repository) GetInvoiceItems(ctx context.Context, orgID int64, invoiceID int64) ([]InvoiceItem, error) {
	var list []InvoiceItem
	err := r.db.SelectContext(ctx, &list, "SELECT * FROM customer_invoice_items WHERE invoice_id = ? AND org_id = ? ORDER BY display_order ASC", invoiceID, orgID)
	return list, err
}

func (r *repository) GetInvoicePayments(ctx context.Context, orgID int64, invoiceID int64) ([]InvoicePayment, error) {
	var list []InvoicePayment
	err := r.db.SelectContext(ctx, &list, "SELECT * FROM customer_invoice_payments WHERE invoice_id = ? AND org_id = ? ORDER BY payment_date DESC", invoiceID, orgID)
	return list, err
}

func (r *repository) GetInvoiceDocuments(ctx context.Context, orgID int64, invoiceID int64) ([]InvoiceDocument, error) {
	var list []InvoiceDocument
	err := r.db.SelectContext(ctx, &list, "SELECT * FROM customer_invoice_documents WHERE invoice_id = ? AND org_id = ? ORDER BY uploaded_at DESC", invoiceID, orgID)
	return list, err
}

func (r *repository) GetInvoiceHistory(ctx context.Context, orgID int64, invoiceID int64) ([]InvoiceHistory, error) {
	var list []InvoiceHistory
	err := r.db.SelectContext(ctx, &list, "SELECT * FROM customer_invoice_history WHERE invoice_id = ? AND org_id = ? ORDER BY created_at DESC", invoiceID, orgID)
	return list, err
}

func (r *repository) AddHistory(ctx context.Context, orgID int64, invoiceID int64, title string, description string, userName string) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO customer_invoice_history (org_id, invoice_id, title, description, user_name, created_at) VALUES (?, ?, ?, ?, ?, NOW())",
		orgID, invoiceID, title, description, userName,
	)
	return err
}

func (r *repository) CreateApprovalRequestForInvoice(ctx context.Context, inv *Invoice, userName string) error {
	var count int
	_ = r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM approval_requests WHERE org_id = ? AND related_entity_type = 'INVOICE' AND related_entity_id = ? AND (status = 'Pending' OR status = 'PENDING')", inv.OrgID, inv.ID)
	if count > 0 {
		return nil
	}

	reqCode := fmt.Sprintf("FIN-APP-%d", 1000+inv.ID)
	title := fmt.Sprintf("Invoice Approval: %s (%s)", inv.InvoiceNumber, inv.CustomerName)
	desc := fmt.Sprintf("Approval requested for Customer Invoice %s totaling $%.2f %s.", inv.InvoiceNumber, inv.TotalAmount, inv.Currency)

	query := `
		INSERT INTO approval_requests (
			org_id, request_code, title, category, type, status, priority,
			related_entity_type, related_entity_id, related_ref, customer_name, customer_id, shipment_id, booking_id,
			requested_by_name, department, avatar, description, created_at, updated_at
		) VALUES (
			?, ?, ?, 'FINANCE', 'Finance Approval', 'Pending', 'HIGH',
			'INVOICE', ?, ?, ?, ?, ?, ?,
			?, 'Finance', 'FS', ?, NOW(), NOW()
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		inv.OrgID, reqCode, title, inv.ID, inv.InvoiceNumber, inv.CustomerName, inv.CustomerID, inv.ShipmentID, inv.BookingID,
		userName, desc,
	)
	return err
}

func (r *repository) RecordPayment(
	ctx context.Context,
	orgID int64,
	invoiceID int64,
	payment *InvoicePayment,
	newPaidAmount float64,
	newBalanceDue float64,
	newStatus string,
	historyTitle string,
	historyDesc string,
	userName string,
) (*Invoice, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	payQuery := `
		INSERT INTO customer_invoice_payments (
			org_id, invoice_id, payment_ref, amount, payment_method, status, payment_date, notes, created_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, NOW()
		)
	`
	res, err := tx.ExecContext(ctx, payQuery,
		orgID, invoiceID, payment.PaymentRef, payment.Amount, payment.PaymentMethod, payment.Status, payment.PaymentDate, payment.Notes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert payment: %w", err)
	}

	payID, err := res.LastInsertId()
	if err == nil {
		payment.ID = payID
	}

	invQuery := `
		UPDATE customer_invoices SET
			paid_amount = ?, balance_due = ?, status = ?, updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	if _, err := tx.ExecContext(ctx, invQuery, newPaidAmount, newBalanceDue, newStatus, invoiceID, orgID); err != nil {
		return nil, fmt.Errorf("failed to update invoice balances: %w", err)
	}

	histQuery := `
		INSERT INTO customer_invoice_history (org_id, invoice_id, title, description, user_name, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())
	`
	if _, err := tx.ExecContext(ctx, histQuery, orgID, invoiceID, historyTitle, historyDesc, userName); err != nil {
		return nil, fmt.Errorf("failed to insert invoice history: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetInvoiceByID(ctx, orgID, invoiceID)
}

func (r *repository) GetAllPayments(ctx context.Context, orgID int64) ([]InvoicePayment, error) {
	var list []InvoicePayment
	err := r.db.SelectContext(ctx, &list, "SELECT * FROM customer_invoice_payments WHERE org_id = ? ORDER BY payment_date DESC", orgID)
	return list, err
}

func (r *repository) AddDocument(ctx context.Context, orgID int64, invoiceID int64, doc *InvoiceDocument) error {
	query := `
		INSERT INTO customer_invoice_documents (
			org_id, invoice_id, document_name, file_size, file_type, s3_key, uploaded_at
		) VALUES (
			?, ?, ?, ?, ?, ?, NOW()
		)
	`
	res, err := r.db.ExecContext(ctx, query, orgID, invoiceID, doc.DocumentName, doc.FileSize, doc.FileType, doc.S3Key)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		doc.ID = id
	}
	return nil
}

func (r *repository) GenerateInvoiceNumber(ctx context.Context, orgID int64) (string, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM customer_invoices WHERE org_id = ?", orgID)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	nextNum := count + 449
	year := time.Now().Year()
	return fmt.Sprintf("INV-%d-%04d", year, nextNum), nil
}
