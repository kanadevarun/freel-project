package predictions

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// InvoiceFinanceData holds deterministic financial and collections facts for an invoice
type InvoiceFinanceData struct {
	InvoiceID            int64
	OrgID                int64
	InvoiceNumber        string
	CustomerID           int64
	CustomerName         string
	ShipmentID           *int64
	ShipmentNumber       string
	TotalAmount          float64
	PaidAmount           float64
	BalanceDue           float64
	Status               string
	Currency             string
	InvoiceDate          string
	DueDate              string
	DaysLeft             int
	IsOverdue            bool
	DaysOverdue          int
	CustomerCreditRating string
	CustomerHealthScore  int
	PaymentHistoryCount  int
	DisputeStatus        string
	ContextMap           map[string]interface{}
}

// FinanceCollectionsDataProvider extracts authoritative financial data with tenant isolation
type FinanceCollectionsDataProvider struct {
	db *sql.DB
}

// NewFinanceCollectionsDataProvider instantiates the provider
func NewFinanceCollectionsDataProvider(db *sql.DB) *FinanceCollectionsDataProvider {
	return &FinanceCollectionsDataProvider{db: db}
}

// FetchInvoiceData retrieves invoice, debtor profile, and payment history from MariaDB
func (p *FinanceCollectionsDataProvider) FetchInvoiceData(ctx context.Context, orgID int64, invoiceID int64) (*InvoiceFinanceData, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	query := `
		SELECT 
			id, org_id, COALESCE(invoice_number, ''), COALESCE(customer_id, 0),
			COALESCE(customer_name, ''), shipment_id, COALESCE(shipment_number, ''),
			COALESCE(total_amount, 0.0), COALESCE(paid_amount, 0.0), COALESCE(balance_due, 0.0),
			COALESCE(status, 'Draft'), COALESCE(currency, 'USD'),
			COALESCE(invoice_date, ''), COALESCE(due_date, '')
		FROM customer_invoices
		WHERE id = ? AND org_id = ?
	`
	var (
		invID, oID, custID           int64
		invNum, custName, shipNum    string
		shipID                       sql.NullInt64
		totAmt, paidAmt, balDue      float64
		status, currency             string
		invDateRaw, dueDateRaw       string
	)

	err := p.db.QueryRowContext(ctx, query, invoiceID, orgID).Scan(
		&invID, &oID, &invNum, &custID,
		&custName, &shipID, &shipNum,
		&totAmt, &paidAmt, &balDue,
		&status, &currency,
		&invDateRaw, &dueDateRaw,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invoice %d not found for organization %d", invoiceID, orgID)
	} else if err != nil {
		return nil, fmt.Errorf("failed querying customer_invoices: %w", err)
	}

	// Deterministic balance due calculation
	if balDue == 0.0 && totAmt > paidAmt {
		balDue = totAmt - paidAmt
	}

	// Deterministic due date and aging calculation
	var daysLeft int
	var isOverdue bool
	var daysOverdue int

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if dueDateRaw != "" {
		if dueTime, err := time.Parse("2006-01-02", dueDateRaw); err == nil {
			diffHours := dueTime.Sub(today).Hours()
			daysLeft = int(diffHours / 24.0)
			if daysLeft < 0 {
				isOverdue = true
				daysOverdue = -daysLeft
			}
		}
	}

	if status == "Overdue" || status == "OVERDUE" {
		isOverdue = true
		if daysOverdue <= 0 {
			daysOverdue = 1
		}
	}

	// Fetch debtor customer rating and health score
	customerCreditRating := "GOOD"
	customerHealthScore := 80
	if custID > 0 {
		var creditRating sql.NullString
		var healthScore sql.NullInt64
		_ = p.db.QueryRowContext(ctx, "SELECT credit_status, health_score FROM customers WHERE id = ? AND org_id = ?", custID, orgID).Scan(&creditRating, &healthScore)
		if creditRating.Valid && creditRating.String != "" {
			customerCreditRating = creditRating.String
		}
		if healthScore.Valid && healthScore.Int64 > 0 {
			customerHealthScore = int(healthScore.Int64)
		}
	}

	// Fetch payment history count for this invoice
	var paymentCount int
	_ = p.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM customer_invoice_payments WHERE invoice_id = ? AND org_id = ?", invoiceID, orgID).Scan(&paymentCount)

	data := &InvoiceFinanceData{
		InvoiceID:            invID,
		OrgID:                oID,
		InvoiceNumber:        invNum,
		CustomerID:           custID,
		CustomerName:         custName,
		ShipmentNumber:       shipNum,
		TotalAmount:          totAmt,
		PaidAmount:           paidAmt,
		BalanceDue:           balDue,
		Status:               status,
		Currency:             currency,
		InvoiceDate:          invDateRaw,
		DueDate:              dueDateRaw,
		DaysLeft:             daysLeft,
		IsOverdue:            isOverdue,
		DaysOverdue:          daysOverdue,
		CustomerCreditRating: customerCreditRating,
		CustomerHealthScore:  customerHealthScore,
		PaymentHistoryCount:  paymentCount,
		DisputeStatus:        "NONE",
	}

	if shipID.Valid {
		sID := shipID.Int64
		data.ShipmentID = &sID
	}

	data.ContextMap = map[string]interface{}{
		"invoice_id":             data.InvoiceID,
		"invoice_number":         data.InvoiceNumber,
		"customer_id":            data.CustomerID,
		"customer_name":          data.CustomerName,
		"total_amount":           data.TotalAmount,
		"paid_amount":            data.PaidAmount,
		"balance_due":            data.BalanceDue,
		"status":                 data.Status,
		"currency":               data.Currency,
		"invoice_date":           data.InvoiceDate,
		"due_date":               data.DueDate,
		"days_left":              data.DaysLeft,
		"is_overdue":             data.IsOverdue,
		"days_overdue":           data.DaysOverdue,
		"customer_credit_rating": data.CustomerCreditRating,
		"customer_health_score":  data.CustomerHealthScore,
		"payment_history_count":  data.PaymentHistoryCount,
		"dispute_status":         data.DisputeStatus,
		"shipment_number":        data.ShipmentNumber,
	}

	return data, nil
}
