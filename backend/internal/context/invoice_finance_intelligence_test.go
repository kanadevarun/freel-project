package bcontext

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupInvoiceIntelligenceMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, *defaultService) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	svc := &defaultService{db: sqlxDB}
	return sqlxDB, mock, svc
}

func TestGetInvoice360FinanceIntelligence_Validation(t *testing.T) {
	sqlxDB, _, svc := setupInvoiceIntelligenceMockDB(t)
	defer sqlxDB.Close()

	ctx := context.Background()

	// 1. Invalid Org ID
	_, err := svc.GetInvoice360FinanceIntelligence(ctx, 0, 10, "corr-1", 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid organization id")

	// 2. Invalid Invoice ID
	_, err = svc.GetInvoice360FinanceIntelligence(ctx, 1, 0, "corr-1", 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid invoice id")
}

func TestGetInvoice360FinanceIntelligence_PaidInvoice(t *testing.T) {
	sqlxDB, mock, svc := setupInvoiceIntelligenceMockDB(t)
	defer sqlxDB.Close()

	ctx := context.Background()
	now := time.Now().UTC()
	dueDate := now.Add(48 * time.Hour)
	paidDate := now.Add(-2 * time.Hour)

	// 1. Invoice query
	invRows := sqlmock.NewRows([]string{
		"id", "org_id", "invoice_number", "customer_id", "customer_name", "customer_country",
		"shipment_id", "shipment_number", "booking_id", "booking_number", "quotation_id", "quote_number",
		"route", "origin", "destination", "invoice_date", "due_date", "currency", "subtotal", "tax_amount",
		"discount_amount", "total_amount", "paid_amount", "balance_due", "status", "type", "created_at", "updated_at",
	}).AddRow(
		1, 1, "INV-2026-001", 10, "Acme Logistics", "US",
		101, "SH-101", 201, "BK-201", 301, "QT-301",
		"INNSA-NLRTM", "INNSA", "NLRTM", now.Add(-24*time.Hour), dueDate, "USD", 5000.0, 500.0,
		0.0, 5500.0, 5500.0, 0.0, "Paid", "CUSTOMER_AR", now.Add(-24*time.Hour), now,
	)
	mock.ExpectQuery("SELECT (.+) FROM customer_invoices WHERE id = \\? AND org_id = \\?").
		WithArgs(int64(1), int64(1)).
		WillReturnRows(invRows)

	// 2. Line items
	itemRows := sqlmock.NewRows([]string{
		"id", "description", "service_category", "quantity", "unit_price", "total_amount",
	}).AddRow(1, "Ocean Freight", "FREIGHT", 1.0, 5000.0, 5000.0)
	mock.ExpectQuery("SELECT (.+) FROM customer_invoice_items WHERE invoice_id = \\? AND org_id = \\?").
		WithArgs(int64(1), int64(1)).
		WillReturnRows(itemRows)

	// 3. Payments
	payRows := sqlmock.NewRows([]string{
		"id", "payment_ref", "amount", "payment_method", "status", "payment_date", "notes",
	}).AddRow(1, "PAY-001", 5500.0, "Wire Transfer", "Completed", paidDate, "Full payment")
	mock.ExpectQuery("SELECT (.+) FROM customer_invoice_payments WHERE invoice_id = \\? AND org_id = \\?").
		WithArgs(int64(1), int64(1)).
		WillReturnRows(payRows)

	// 4. Customer AR totals
	custAggRows := sqlmock.NewRows([]string{
		"total_invoiced", "total_paid", "total_outstanding", "total_count", "paid_count", "open_count",
	}).AddRow(15000.0, 15000.0, 0.0, 3, 3, 0)
	mock.ExpectQuery("SELECT (.+) FROM customer_invoices WHERE customer_id = \\? AND org_id = \\?").
		WithArgs(int64(10), int64(1)).
		WillReturnRows(custAggRows)

	// 5. Customer Overdue totals
	overdueRows := sqlmock.NewRows([]string{
		"overdue_amount", "overdue_count",
	}).AddRow(0.0, 0)
	mock.ExpectQuery("SELECT (.+) FROM customer_invoices WHERE customer_id = \\? AND org_id = \\? AND status != 'Paid'").
		WithArgs(int64(10), int64(1), sqlmock.AnyArg()).
		WillReturnRows(overdueRows)

	// 6. Payment delays
	delayRows := sqlmock.NewRows([]string{"avg_delay", "late_count"}).AddRow(0.0, 0)
	mock.ExpectQuery("SELECT (.+) FROM customer_invoice_payments p (.+) WHERE i.customer_id = \\? AND i.org_id = \\?").
		WithArgs(int64(10), int64(1)).
		WillReturnRows(delayRows)

	// 7. Quotation cost
	quoteRows := sqlmock.NewRows([]string{
		"total_amount", "total_cost", "gross_profit", "gross_margin_pct",
	}).AddRow(5500.0, 4500.0, 1000.0, 18.18)
	mock.ExpectQuery("SELECT (.+) FROM quotations WHERE id = \\? AND org_id = \\?").
		WithArgs(int64(301), int64(1)).
		WillReturnRows(quoteRows)

	intel, err := svc.GetInvoice360FinanceIntelligence(ctx, 1, 1, "corr-test-1", 10)
	require.NoError(t, err)
	require.NotNil(t, intel)

	assert.Equal(t, int64(1), intel.InvoiceID)
	assert.Equal(t, "INV-2026-001", intel.Identity.InvoiceNumber)
	assert.True(t, intel.Identity.IsPaid)
	assert.False(t, intel.Identity.IsOverdue)
	assert.Equal(t, 0.0, intel.Identity.BalanceDue)
	assert.Equal(t, "NOT_DUE", intel.Identity.AgingBucket)
	assert.Equal(t, "LOW", intel.RiskIndicators.OverallRiskRating)
	assert.False(t, intel.RevenueCost.HasMissingCost)
	assert.NotNil(t, intel.RevenueCost.GrossMarginAmount)
	assert.Equal(t, 1000.0, *intel.RevenueCost.GrossMarginAmount)
	assert.True(t, intel.ReadOnly)
}

func TestGetInvoice360FinanceIntelligence_OverdueWithAgingBuckets(t *testing.T) {
	sqlxDB, mock, svc := setupInvoiceIntelligenceMockDB(t)
	defer sqlxDB.Close()

	ctx := context.Background()
	now := time.Now().UTC()
	// 45 days overdue -> 31-60_DAYS bucket
	dueDate := now.Add(-45 * 24 * time.Hour)

	invRows := sqlmock.NewRows([]string{
		"id", "org_id", "invoice_number", "customer_id", "customer_name", "customer_country",
		"shipment_id", "shipment_number", "booking_id", "booking_number", "quotation_id", "quote_number",
		"route", "origin", "destination", "invoice_date", "due_date", "currency", "subtotal", "tax_amount",
		"discount_amount", "total_amount", "paid_amount", "balance_due", "status", "type", "created_at", "updated_at",
	}).AddRow(
		2, 1, "INV-2026-002", 20, "Global Importers", "DE",
		nil, "", nil, "", nil, "",
		"", "", "", now.Add(-60*24*time.Hour), dueDate, "USD", 25000.0, 0.0,
		0.0, 25000.0, 5000.0, 20000.0, "Overdue", "CUSTOMER_AR", now.Add(-60*24*time.Hour), now,
	)
	mock.ExpectQuery("SELECT (.+) FROM customer_invoices WHERE id = \\? AND org_id = \\?").
		WithArgs(int64(2), int64(1)).
		WillReturnRows(invRows)

	// Line items
	itemRows := sqlmock.NewRows([]string{
		"id", "description", "service_category", "quantity", "unit_price", "total_amount",
	}).AddRow(1, "Air Charter Freight", "AIR", 1.0, 25000.0, 25000.0)
	mock.ExpectQuery("SELECT (.+) FROM customer_invoice_items WHERE invoice_id = \\? AND org_id = \\?").
		WithArgs(int64(2), int64(1)).
		WillReturnRows(itemRows)

	// Payments
	payRows := sqlmock.NewRows([]string{
		"id", "payment_ref", "amount", "payment_method", "status", "payment_date", "notes",
	}).AddRow(1, "PAY-PARTIAL", 5000.0, "Credit Card", "Completed", now.Add(-50*24*time.Hour), "Deposit")
	mock.ExpectQuery("SELECT (.+) FROM customer_invoice_payments WHERE invoice_id = \\? AND org_id = \\?").
		WithArgs(int64(2), int64(1)).
		WillReturnRows(payRows)

	// Customer AR totals
	custAggRows := sqlmock.NewRows([]string{
		"total_invoiced", "total_paid", "total_outstanding", "total_count", "paid_count", "open_count",
	}).AddRow(50000.0, 15000.0, 35000.0, 3, 1, 2)
	mock.ExpectQuery("SELECT (.+) FROM customer_invoices WHERE customer_id = \\? AND org_id = \\?").
		WithArgs(int64(20), int64(1)).
		WillReturnRows(custAggRows)

	// Customer Overdue
	overdueRows := sqlmock.NewRows([]string{
		"overdue_amount", "overdue_count",
	}).AddRow(20000.0, 1)
	mock.ExpectQuery("SELECT (.+) FROM customer_invoices WHERE customer_id = \\? AND org_id = \\? AND status != 'Paid'").
		WithArgs(int64(20), int64(1), sqlmock.AnyArg()).
		WillReturnRows(overdueRows)

	// Delay
	delayRows := sqlmock.NewRows([]string{"avg_delay", "late_count"}).AddRow(14.5, 3)
	mock.ExpectQuery("SELECT (.+) FROM customer_invoice_payments p (.+) WHERE i.customer_id = \\? AND i.org_id = \\?").
		WithArgs(int64(20), int64(1)).
		WillReturnRows(delayRows)

	intel, err := svc.GetInvoice360FinanceIntelligence(ctx, 1, 2, "corr-test-2", 10)
	require.NoError(t, err)
	require.NotNil(t, intel)

	assert.True(t, intel.Identity.IsOverdue)
	assert.True(t, intel.Identity.IsPartiallyPaid)
	assert.Equal(t, 20000.0, intel.Identity.BalanceDue)
	assert.Equal(t, "31-60_DAYS", intel.Identity.AgingBucket)
	assert.True(t, intel.RiskIndicators.HasLargeBalance)
	assert.True(t, intel.RiskIndicators.RepeatedCustomerLatePayer)
	assert.Equal(t, "CRITICAL", intel.RiskIndicators.OverallRiskRating)
	assert.True(t, intel.RevenueCost.IsUnlinkedInvoice)
	assert.True(t, intel.RevenueCost.HasMissingCost)
}

func TestGetInvoice360FinanceIntelligence_LineItemMismatch(t *testing.T) {
	sqlxDB, mock, svc := setupInvoiceIntelligenceMockDB(t)
	defer sqlxDB.Close()

	ctx := context.Background()
	now := time.Now().UTC()
	dueDate := now.Add(7 * 24 * time.Hour)

	invRows := sqlmock.NewRows([]string{
		"id", "org_id", "invoice_number", "customer_id", "customer_name", "customer_country",
		"shipment_id", "shipment_number", "booking_id", "booking_number", "quotation_id", "quote_number",
		"route", "origin", "destination", "invoice_date", "due_date", "currency", "subtotal", "tax_amount",
		"discount_amount", "total_amount", "paid_amount", "balance_due", "status", "type", "created_at", "updated_at",
	}).AddRow(
		3, 1, "INV-2026-003", 10, "Acme Logistics", "US",
		nil, "", nil, "", nil, "",
		"", "", "", now, dueDate, "USD", 10000.0, 0.0,
		0.0, 10000.0, 0.0, 10000.0, "Issued", "CUSTOMER_AR", now, now,
	)
	mock.ExpectQuery("SELECT (.+) FROM customer_invoices WHERE id = \\? AND org_id = \\?").
		WithArgs(int64(3), int64(1)).
		WillReturnRows(invRows)

	// Line items sum to 8000, while invoice subtotal is 10000 -> mismatch
	itemRows := sqlmock.NewRows([]string{
		"id", "description", "service_category", "quantity", "unit_price", "total_amount",
	}).AddRow(1, "Port Handling", "HANDLING", 1.0, 8000.0, 8000.0)
	mock.ExpectQuery("SELECT (.+) FROM customer_invoice_items WHERE invoice_id = \\? AND org_id = \\?").
		WithArgs(int64(3), int64(1)).
		WillReturnRows(itemRows)

	// Payments (empty)
	mock.ExpectQuery("SELECT (.+) FROM customer_invoice_payments WHERE invoice_id = \\? AND org_id = \\?").
		WithArgs(int64(3), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "payment_ref", "amount", "payment_method", "status", "payment_date", "notes"}))

	// Customer AR totals
	custAggRows := sqlmock.NewRows([]string{
		"total_invoiced", "total_paid", "total_outstanding", "total_count", "paid_count", "open_count",
	}).AddRow(10000.0, 0.0, 10000.0, 1, 0, 1)
	mock.ExpectQuery("SELECT (.+) FROM customer_invoices WHERE customer_id = \\? AND org_id = \\?").
		WithArgs(int64(10), int64(1)).
		WillReturnRows(custAggRows)

	// Customer Overdue
	mock.ExpectQuery("SELECT (.+) FROM customer_invoices WHERE customer_id = \\? AND org_id = \\? AND status != 'Paid'").
		WithArgs(int64(10), int64(1), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"overdue_amount", "overdue_count"}).AddRow(0.0, 0))

	// Delay
	mock.ExpectQuery("SELECT (.+) FROM customer_invoice_payments p (.+) WHERE i.customer_id = \\? AND i.org_id = \\?").
		WithArgs(int64(10), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"avg_delay", "late_count"}).AddRow(0.0, 0))

	intel, err := svc.GetInvoice360FinanceIntelligence(ctx, 1, 3, "corr-test-3", 10)
	require.NoError(t, err)
	require.NotNil(t, intel)

	assert.True(t, intel.RiskIndicators.LineItemsInconsistent)
	assert.Equal(t, 8000.0, intel.AuditedLineItemTotal)
	assert.Len(t, intel.Warnings, 1)
	assert.Contains(t, intel.Warnings[0], "Line items total (8000.00) differs from recorded invoice subtotal")
}

func TestGetInvoice360FinanceIntelligence_CrossTenantIsolation(t *testing.T) {
	sqlxDB, mock, svc := setupInvoiceIntelligenceMockDB(t)
	defer sqlxDB.Close()

	ctx := context.Background()

	// Org 99 attempts to query Invoice 1 belonging to Org 1
	mock.ExpectQuery("SELECT (.+) FROM customer_invoices WHERE id = \\? AND org_id = \\?").
		WithArgs(int64(1), int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{"id"})) // No rows

	_, err := svc.GetInvoice360FinanceIntelligence(ctx, 99, 1, "corr-cross-tenant", 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invoice not found or unauthorized")
}

func TestGetOrgFinanceSummary_Success(t *testing.T) {
	sqlxDB, mock, svc := setupInvoiceIntelligenceMockDB(t)
	defer sqlxDB.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	// 1. Overall Aggregates
	aggRows := sqlmock.NewRows([]string{
		"active_count", "open_count", "paid_count", "cancelled_count", "total_invoiced", "total_paid", "total_outstanding",
	}).AddRow(10, 4, 5, 1, 100000.0, 60000.0, 40000.0)
	mock.ExpectQuery("SELECT (.+) FROM customer_invoices WHERE org_id = \\?").
		WithArgs(int64(1)).
		WillReturnRows(aggRows)

	// 2. Open invoices for aging
	openRows := sqlmock.NewRows([]string{
		"id", "customer_id", "customer_name", "currency", "balance_due", "due_date",
	}).
		AddRow(1, 10, "Acme", "USD", 10000.0, now.Add(-10*24*time.Hour)). // 1-30 days overdue
		AddRow(2, 10, "Acme", "USD", 15000.0, now.Add(-40*24*time.Hour)). // 31-60 days overdue
		AddRow(3, 20, "Global", "USD", 15000.0, now.Add(5*24*time.Hour))  // Not due, approaching (5 days)
	mock.ExpectQuery("SELECT (.+) FROM customer_invoices WHERE org_id = \\? AND status != 'Paid'").
		WithArgs(int64(1)).
		WillReturnRows(openRows)

	summary, err := svc.GetOrgFinanceSummary(ctx, 1, "corr-org-fin", 10)
	require.NoError(t, err)
	require.NotNil(t, summary)

	assert.Equal(t, 10, summary.ActiveInvoicesCount)
	assert.Equal(t, 4, summary.OpenInvoicesCount)
	assert.Equal(t, 2, summary.OverdueInvoicesCount)
	assert.Equal(t, 1, summary.InvoicesApproachingDueCount)
	assert.Equal(t, 40000.0, summary.TotalOutstandingReceivables)
	assert.Equal(t, 25000.0, summary.TotalOverdueReceivables)
	assert.Equal(t, 10000.0, summary.AgingBreakdown.Days1To30)
	assert.Equal(t, 15000.0, summary.AgingBreakdown.Days31To60)
	assert.Equal(t, 15000.0, summary.AgingBreakdown.NotDue)
	assert.Len(t, summary.TopCustomerExposures, 2)
	assert.True(t, summary.ReadOnly)
}
