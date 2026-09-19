package bcontext

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func setupMockInsightsDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, *defaultService) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	dbx := sqlx.NewDb(mockDB, "sqlmock")
	svc := &defaultService{
		db:      dbx,
		rbacSvc: nil,
	}
	return dbx, mock, svc
}

func TestCrossModuleInsights_ValidationErrors(t *testing.T) {
	db, _, svc := setupMockInsightsDB(t)
	defer db.Close()

	ctx := context.Background()

	// 1. Invalid org ID
	_, err := svc.GetCrossModuleInsights(ctx, 0, "CUSTOMER", 1, "test-corr", 1)
	if err == nil {
		t.Errorf("expected error for orgID=0, got nil")
	}

	// 2. Unsupported entity type
	_, err = svc.GetCrossModuleInsights(ctx, 1, "UNKNOWN_TYPE", 1, "test-corr", 1)
	if err == nil {
		t.Errorf("expected error for unsupported entity type, got nil")
	}
}

func TestCrossModuleInsights_CustomerDelinquentWithActiveOps(t *testing.T) {
	db, mock, svc := setupMockInsightsDB(t)
	defer db.Close()

	ctx := context.Background()
	orgID := int64(1)
	customerID := int64(42)

	// 1. Query customer
	custRows := sqlmock.NewRows([]string{"id", "name", "code"}).
		AddRow(customerID, "Acme Imports Corp", "ACME")
	mock.ExpectQuery(`SELECT id, name, code FROM customers WHERE id = \? AND org_id = \?`).
		WithArgs(customerID, orgID).
		WillReturnRows(custRows)

	// 2. Query overdue stats
	overdueRows := sqlmock.NewRows([]string{"overdue_count", "overdue_sum"}).
		AddRow(3, 24500.0)
	mock.ExpectQuery(`SELECT COUNT\(\*\) as overdue_count, COALESCE\(SUM\(balance_due\), 0\) as overdue_sum FROM customer_invoices`).
		WithArgs(orgID, customerID, sqlmock.AnyArg()).
		WillReturnRows(overdueRows)

	// 3. Query active shipments count
	activeShipRows := sqlmock.NewRows([]string{"count"}).AddRow(4)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM shipments s`).
		WithArgs(orgID, customerID).
		WillReturnRows(activeShipRows)

	// 4. Query active contracts count
	contractsRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM contracts`).
		WithArgs(orgID, customerID, sqlmock.AnyArg()).
		WillReturnRows(contractsRows)

	res, err := svc.GetCrossModuleInsights(ctx, orgID, "CUSTOMER", customerID, "corr-cust-test", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TotalInsightsCount != 2 {
		t.Errorf("expected 2 insights, got %d", res.TotalInsightsCount)
	}

	if res.CriticalCount != 1 {
		t.Errorf("expected 1 critical insight, got %d", res.CriticalCount)
	}

	if res.PrimaryEntity.Reference != "Acme Imports Corp" {
		t.Errorf("expected primary entity reference 'Acme Imports Corp', got '%s'", res.PrimaryEntity.Reference)
	}

	if res.AISynthesis == nil {
		t.Fatalf("expected AI synthesis, got nil")
	}

	if res.AISynthesis.Confidence != "HIGH" {
		t.Errorf("expected HIGH confidence, got '%s'", res.AISynthesis.Confidence)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled mock expectations: %v", err)
	}
}

func TestCrossModuleInsights_ShipmentDelayedWithOverdueInvoice(t *testing.T) {
	db, mock, svc := setupMockInsightsDB(t)
	defer db.Close()

	ctx := context.Background()
	orgID := int64(1)
	shipmentID := int64(101)

	// 1. Query shipment
	eta := time.Now().Add(-48 * time.Hour)
	shRows := sqlmock.NewRows([]string{"id", "carrier_scac", "closure_status", "status", "eta", "created_at", "customer_id", "customer_name"}).
		AddRow(shipmentID, "MAEU", "ACTIVE", "IN_TRANSIT", eta, time.Now().Add(-120*time.Hour), int64(10), "Global Forwarding LLC")
	mock.ExpectQuery(`SELECT s\.id, s\.carrier_scac, s\.closure_status, s\.status, s\.eta, s\.created_at`).
		WithArgs(shipmentID, orgID).
		WillReturnRows(shRows)

	// 2. Query invoices for shipment
	invRows := sqlmock.NewRows([]string{"id", "invoice_number", "due_date", "balance_due", "status"}).
		AddRow(int64(201), "INV-2026-0042", time.Now().Add(-72*time.Hour), 8950.0, "ISSUED")
	mock.ExpectQuery(`SELECT id, invoice_number, due_date, balance_due, status FROM customer_invoices`).
		WithArgs(orgID, shipmentID).
		WillReturnRows(invRows)

	// 3. Query total invoices count for unbilled check
	totalInvs := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM customer_invoices WHERE org_id = \? AND shipment_id = \?`).
		WithArgs(orgID, shipmentID).
		WillReturnRows(totalInvs)

	res, err := svc.GetCrossModuleInsights(ctx, orgID, "SHIPMENT", shipmentID, "corr-sh-test", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TotalInsightsCount != 1 {
		t.Fatalf("expected 1 insight, got %d", res.TotalInsightsCount)
	}

	ins := res.Insights[0]
	if ins.InsightType != TypeOperationalFinance {
		t.Errorf("expected OPERATIONAL_FINANCE, got %s", ins.InsightType)
	}

	if ins.Severity != SeverityHigh {
		t.Errorf("expected HIGH severity, got %s", ins.Severity)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled mock expectations: %v", err)
	}
}

func TestCrossModuleInsights_OrgWideEmpty(t *testing.T) {
	db, mock, svc := setupMockInsightsDB(t)
	defer db.Close()

	ctx := context.Background()
	orgID := int64(1)

	// 1. Delayed shipments with overdue invoices: 0 rows
	mock.ExpectQuery(`SELECT s\.id as shipment_id, inv\.id as invoice_id`).
		WithArgs(orgID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"shipment_id", "invoice_id", "invoice_number", "balance_due", "due_date", "eta", "status"}))

	// 2. Delinquent customers with active shipments: 0 rows
	mock.ExpectQuery(`SELECT c\.id as customer_id, c\.name as customer_name`).
		WithArgs(orgID, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"customer_id", "customer_name", "overdue_count", "total_overdue", "active_shipments_count"}))

	summary, err := svc.GetOrgCrossModuleSummary(ctx, orgID, "corr-org-test", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.TotalInsights != 0 {
		t.Errorf("expected 0 insights for empty org, got %d", summary.TotalInsights)
	}

	if summary.CriticalInsights != 0 || summary.HighInsights != 0 {
		t.Errorf("expected 0 critical and 0 high insights")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled mock expectations: %v", err)
	}
}
