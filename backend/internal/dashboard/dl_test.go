package dashboard_test

import (
	"context"
	"testing"

	"github.com/freel/backend/internal/dashboard"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func TestDashboardMissionControlAggregation(t *testing.T) {
	db, err := sqlx.Open("mysql", "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC")
	if err != nil {
		t.Skip("MySQL not accessible for test:", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("MySQL ping failed:", err)
		return
	}

	dl := dashboard.NewDataLayer(db)
	bl := dashboard.NewBusinessLogic(dl)
	ctx := context.Background()

	// 1. Test Populated / Mature Org (Org 1)
	respOrg1, err := bl.GetMissionControl(ctx, 1, "", "", "7D")
	if err != nil {
		t.Fatalf("Failed to get mission control for Org 1: %v", err)
	}
	if respOrg1 == nil {
		t.Fatal("Expected non-nil response for Org 1")
	}

	t.Logf("Org 1 Maturity: %s, IsNewUser: %v, IsOperational: %v", respOrg1.Stats.AccountMaturity, respOrg1.Stats.IsNewUser, respOrg1.Stats.IsOperational)
	t.Logf("Org 1 Stats: Customers=%d, Leads=%d, RFQs=%d, Quotations=%d, Bookings=%d, Shipments=%d, Invoices=%d, Revenue=$%.2f",
		respOrg1.Stats.TotalCustomers, respOrg1.Stats.TotalLeads, respOrg1.Stats.TotalRFQs, respOrg1.Stats.TotalQuotations,
		respOrg1.Stats.TotalBookings, respOrg1.Stats.TotalShipments, respOrg1.Stats.TotalInvoices, respOrg1.Stats.TotalRevenue)
	t.Logf("Org 1 Attention Items: %d, Active Shipments: %d, Recent Invoices: %d, Reminders: %d",
		len(respOrg1.AttentionItems), len(respOrg1.ActiveShipments), len(respOrg1.InvoiceSummary.RecentInvoices), len(respOrg1.UpcomingReminders))

	if respOrg1.Stats.IsNewUser {
		t.Errorf("Expected Org 1 to be classified as active/operational user, got IsNewUser=true")
	}
	if respOrg1.Stats.AccountMaturity != "MATURE" && respOrg1.Stats.AccountMaturity != "OPERATIONAL" {
		t.Errorf("Expected Org 1 maturity MATURE/OPERATIONAL, got %s", respOrg1.Stats.AccountMaturity)
	}
	if respOrg1.Stats.TotalCustomers == 0 {
		t.Errorf("Expected Org 1 TotalCustomers > 0, got %d", respOrg1.Stats.TotalCustomers)
	}

	// 2. Test Low-Data Org (Org 8801 - only 1 customer)
	respOrg8801, err := bl.GetMissionControl(ctx, 8801, "", "", "7D")
	if err != nil {
		t.Fatalf("Failed to get mission control for Org 8801: %v", err)
	}
	t.Logf("Org 8801 Maturity: %s, IsNewUser: %v, IsOperational: %v", respOrg8801.Stats.AccountMaturity, respOrg8801.Stats.IsNewUser, respOrg8801.Stats.IsOperational)
	if !respOrg8801.Stats.IsNewUser {
		t.Errorf("Expected Org 8801 (1 customer, no operational records) to be IsNewUser=true, got %v", respOrg8801.Stats.IsNewUser)
	}

	// 3. Test Zero-Data Org (Org 1023)
	respOrg1023, err := bl.GetMissionControl(ctx, 1023, "", "", "7D")
	if err != nil {
		t.Fatalf("Failed to get mission control for Org 1023: %v", err)
	}
	t.Logf("Org 1023 Maturity: %s, IsNewUser: %v, IsOperational: %v", respOrg1023.Stats.AccountMaturity, respOrg1023.Stats.IsNewUser, respOrg1023.Stats.IsOperational)
	if !respOrg1023.Stats.IsNewUser {
		t.Errorf("Expected Org 1023 to be IsNewUser=true, got %v", respOrg1023.Stats.IsNewUser)
	}
	if respOrg1023.Stats.AccountMaturity != "NEW" {
		t.Errorf("Expected Org 1023 maturity NEW, got %s", respOrg1023.Stats.AccountMaturity)
	}
	if respOrg1023.Stats.TotalCustomers != 0 || respOrg1023.Stats.TotalInvoices != 0 {
		t.Errorf("Expected Org 1023 to have 0 counts, got customers=%d invoices=%d", respOrg1023.Stats.TotalCustomers, respOrg1023.Stats.TotalInvoices)
	}

	// 4. Verify Tenant Isolation (Org 1023 must not see any Org 1 data)
	if len(respOrg1023.AttentionItems) != 0 {
		t.Errorf("Tenant isolation breach: Org 1023 got %d attention items", len(respOrg1023.AttentionItems))
	}
	if len(respOrg1023.ActiveShipments) != 0 {
		t.Errorf("Tenant isolation breach: Org 1023 got %d active shipments", len(respOrg1023.ActiveShipments))
	}
	if len(respOrg1023.InvoiceSummary.RecentInvoices) != 0 {
		t.Errorf("Tenant isolation breach: Org 1023 got %d recent invoices", len(respOrg1023.InvoiceSummary.RecentInvoices))
	}
}
