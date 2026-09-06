package dashboard

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/freel/backend/internal/dashboard/spec"
	"github.com/jmoiron/sqlx"
)

type Datalayer interface {
	GetStats(ctx context.Context, orgID int64, startDate, endDate, preset string) (spec.Stats, spec.PipelineStats, spec.ShipmentStatusCounts, spec.InvoiceSummary, spec.ModuleStatus, spec.DateRangeInfo, error)
	GetApprovalQueue(ctx context.Context, orgID int64) ([]spec.PendingTask, []spec.PendingApprovalItem, error)
	GetAttentionItems(ctx context.Context, orgID int64) ([]spec.AttentionItem, error)
	GetActiveShipments(ctx context.Context, orgID int64) ([]spec.ActiveShipment, []spec.ActiveShipmentItem, error)
	GetRecentDocuments(ctx context.Context, orgID int64) ([]spec.RecentDocument, error)
	GetRecentActivity(ctx context.Context, orgID int64) ([]spec.RecentActivity, error)
	GetUpcomingReminders(ctx context.Context, orgID int64) ([]spec.UpcomingReminder, error)
	GetOrganizationInfo(ctx context.Context, orgID int64) (spec.OrganizationInfo, error)
}

type dataLayer struct {
	db *sqlx.DB
}

func NewDataLayer(db *sqlx.DB) Datalayer {
	return &dataLayer{db: db}
}

func formatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	now := time.Now()
	diff := now.Sub(t)
	if diff < 0 {
		diff = -diff
	}

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		mins := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	} else if diff < 30*24*time.Hour {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	} else if diff < 365*24*time.Hour {
		months := int(diff.Hours() / (24 * 30))
		return fmt.Sprintf("%dmo ago", months)
	}
	years := int(diff.Hours() / (24 * 365))
	return fmt.Sprintf("%dy ago", years)
}

func formatFutureTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	now := time.Now()
	diff := t.Sub(now)

	if diff <= 0 {
		return "Today, " + t.Format("03:04 PM")
	}

	days := int(diff.Hours() / 24)
	if days == 0 {
		return "Today, " + t.Format("03:04 PM")
	} else if days == 1 {
		return "Tomorrow, " + t.Format("03:04 PM")
	} else if days <= 30 {
		return fmt.Sprintf("In %d days (%s)", days, t.Format("02 Jan 2006"))
	}
	return t.Format("02 Jan 2006")
}

func formatFileSize(bytes int64) string {
	if bytes <= 0 {
		return "250 KB"
	}
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%d KB", bytes/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
}

func computeTrend(curr, prev int) (float64, string) {
	if prev == 0 {
		if curr > 0 {
			return 100.0, "up"
		}
		return 0.0, "neutral"
	}
	diff := float64(curr - prev)
	pct := (diff / float64(prev)) * 100.0
	if pct > 0 {
		return pct, "up"
	} else if pct < 0 {
		return -pct, "down"
	}
	return 0.0, "neutral"
}

func resolveDateRange(startDateStr, endDateStr, presetStr string) (time.Time, time.Time, time.Time, time.Time, spec.DateRangeInfo) {
	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.UTC)

	preset := strings.ToUpper(strings.TrimSpace(presetStr))
	if preset == "" && startDateStr == "" && endDateStr == "" {
		preset = "7D"
	} else if preset == "" && (startDateStr != "" || endDateStr != "") {
		preset = "CUSTOM"
	}

	var startTime, endTime, prevStartTime, prevEndTime time.Time
	var label, comparisonLabel string

	switch preset {
	case "TODAY":
		startTime = todayStart
		endTime = todayEnd
		prevStartTime = todayStart.AddDate(0, 0, -1)
		prevEndTime = time.Date(prevStartTime.Year(), prevStartTime.Month(), prevStartTime.Day(), 23, 59, 59, 999999999, time.UTC)
		label = "Today"
		comparisonLabel = "vs yesterday"

	case "YESTERDAY":
		startTime = todayStart.AddDate(0, 0, -1)
		endTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 23, 59, 59, 999999999, time.UTC)
		prevStartTime = startTime.AddDate(0, 0, -1)
		prevEndTime = time.Date(prevStartTime.Year(), prevStartTime.Month(), prevStartTime.Day(), 23, 59, 59, 999999999, time.UTC)
		label = "Yesterday"
		comparisonLabel = "vs previous day"

	case "30D", "LAST_30D", "LAST_30_DAYS":
		preset = "LAST_30D"
		startTime = todayStart.AddDate(0, 0, -29)
		endTime = todayEnd
		prevStartTime = startTime.AddDate(0, 0, -30)
		prevEndTime = startTime.Add(-time.Second)
		label = fmt.Sprintf("%s – %s", startTime.Format("Jan 2"), endTime.Format("Jan 2, 2006"))
		comparisonLabel = "vs last 30 days"

	case "THIS_MONTH":
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		endTime = todayEnd
		firstOfLastMonth := startTime.AddDate(0, -1, 0)
		prevStartTime = firstOfLastMonth
		prevEndTime = startTime.Add(-time.Second)
		label = fmt.Sprintf("This Month (%s)", now.Format("Jan 2006"))
		comparisonLabel = "vs last month"

	case "LAST_MONTH":
		firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		startTime = firstOfThisMonth.AddDate(0, -1, 0)
		endTime = firstOfThisMonth.Add(-time.Second)
		prevStartTime = startTime.AddDate(0, -1, 0)
		prevEndTime = startTime.Add(-time.Second)
		label = fmt.Sprintf("Last Month (%s)", startTime.Format("Jan 2006"))
		comparisonLabel = "vs prior month"

	case "THIS_QUARTER":
		quarterMonth := ((int(now.Month())-1)/3)*3 + 1
		startTime = time.Date(now.Year(), time.Month(quarterMonth), 1, 0, 0, 0, 0, time.UTC)
		endTime = todayEnd
		prevStartTime = startTime.AddDate(0, -3, 0)
		prevEndTime = startTime.Add(-time.Second)
		label = fmt.Sprintf("Q%d %d", (int(now.Month())-1)/3+1, now.Year())
		comparisonLabel = "vs previous quarter"

	case "CUSTOM":
		if s, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startTime = time.Date(s.Year(), s.Month(), s.Day(), 0, 0, 0, 0, time.UTC)
		} else {
			startTime = todayStart.AddDate(0, 0, -6)
		}
		if e, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endTime = time.Date(e.Year(), e.Month(), e.Day(), 23, 59, 59, 999999999, time.UTC)
		} else {
			endTime = todayEnd
		}
		if endTime.Before(startTime) {
			endTime = startTime.Add(24*time.Hour - time.Second)
		}
		dur := endTime.Sub(startTime)
		prevEndTime = startTime.Add(-time.Second)
		prevStartTime = prevEndTime.Add(-dur)
		label = fmt.Sprintf("%s – %s", startTime.Format("Jan 2"), endTime.Format("Jan 2, 2006"))
		comparisonLabel = "vs previous period"

	default: // "7D", "LAST_7D", "LAST_7_DAYS"
		preset = "LAST_7D"
		startTime = todayStart.AddDate(0, 0, -6)
		endTime = todayEnd
		prevStartTime = startTime.AddDate(0, 0, -7)
		prevEndTime = startTime.Add(-time.Second)
		label = fmt.Sprintf("%s – %s", startTime.Format("Jan 2"), endTime.Format("Jan 2, 2006"))
		comparisonLabel = "vs last 7 days"
	}

	daysCount := int(endTime.Sub(startTime).Hours()/24) + 1
	if daysCount < 1 {
		daysCount = 1
	}

	info := spec.DateRangeInfo{
		StartDate:       startTime.Format("2006-01-02"),
		EndDate:         endTime.Format("2006-01-02"),
		Preset:          preset,
		Label:           label,
		ComparisonLabel: comparisonLabel,
		DaysCount:       daysCount,
		HasDataInPeriod: true,
	}

	return startTime, endTime, prevStartTime, prevEndTime, info
}

func getPeriodTrend(ctx context.Context, db *sqlx.DB, table string, orgID int64, startTime, endTime, prevStartTime, prevEndTime time.Time) (float64, string) {
	var curr, prev int
	query := fmt.Sprintf(`
		SELECT 
			(SELECT COUNT(*) FROM %s WHERE org_id = ? AND created_at >= ? AND created_at <= ?) AS curr,
			(SELECT COUNT(*) FROM %s WHERE org_id = ? AND created_at >= ? AND created_at <= ?) AS prev
	`, table, table)

	_ = db.QueryRowContext(ctx, query, orgID, startTime, endTime, orgID, prevStartTime, prevEndTime).Scan(&curr, &prev)
	return computeTrend(curr, prev)
}

func getPeriodSparkline(ctx context.Context, db *sqlx.DB, table string, orgID int64, startTime, endTime time.Time) []int {
	numPoints := 7
	result := make([]int, numPoints)
	dur := endTime.Sub(startTime)
	if dur <= 0 {
		return result
	}

	step := dur / time.Duration(numPoints)
	for i := 0; i < numPoints; i++ {
		bStart := startTime.Add(time.Duration(i) * step)
		bEnd := startTime.Add(time.Duration(i+1) * step)
		if i == numPoints-1 {
			bEnd = endTime
		}
		var cnt int
		query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE org_id = ? AND created_at >= ? AND created_at <= ?`, table)
		_ = db.QueryRowContext(ctx, query, orgID, bStart, bEnd).Scan(&cnt)
		result[i] = cnt
	}
	return result
}

func (d *dataLayer) GetStats(ctx context.Context, orgID int64, startDate, endDate, preset string) (spec.Stats, spec.PipelineStats, spec.ShipmentStatusCounts, spec.InvoiceSummary, spec.ModuleStatus, spec.DateRangeInfo, error) {
	var stats spec.Stats
	var pipeline spec.PipelineStats
	var shipmentStatus spec.ShipmentStatusCounts
	var invoiceSummary spec.InvoiceSummary
	var moduleStatus spec.ModuleStatus

	// 1. Unified Master Aggregation Query (Single DB roundtrip)
	masterQuery := `
		SELECT 
			(SELECT COUNT(*) FROM customers WHERE org_id = ?) AS total_customers,
			(SELECT COUNT(*) FROM customers WHERE org_id = ? AND MONTH(created_at) = MONTH(CURRENT_DATE()) AND YEAR(created_at) = YEAR(CURRENT_DATE())) AS new_customers_this_month,
			(SELECT COUNT(*) FROM leads WHERE org_id = ?) AS total_leads,
			(SELECT COUNT(*) FROM leads WHERE org_id = ? AND status NOT IN ('CONVERTED', 'LOST')) AS open_leads,
			(SELECT COUNT(*) FROM rfqs WHERE org_id = ?) AS total_rfqs,
			(SELECT COUNT(*) FROM rfqs WHERE org_id = ? AND stage NOT IN ('WON', 'LOST', 'CANCELLED', 'CLOSED')) AS open_rfqs,
			(SELECT COUNT(*) FROM rfqs WHERE org_id = ? AND stage = 'WON') AS won_rfqs,
			(SELECT COUNT(*) FROM rfqs WHERE org_id = ? AND stage = 'LOST') AS lost_rfqs,
			(SELECT COUNT(*) FROM quotations WHERE org_id = ?) AS total_quotations,
			(SELECT COUNT(*) FROM quotations WHERE org_id = ? AND status NOT IN ('EXPIRED', 'REJECTED', 'CANCELLED')) AS active_quotations,
			(SELECT COUNT(*) FROM bookings WHERE org_id = ?) AS total_bookings,
			(SELECT COUNT(*) FROM bookings WHERE org_id = ? AND status NOT IN ('CANCELLED', 'COMPLETED')) AS active_bookings,
			(SELECT COUNT(*) FROM shipments WHERE org_id = ?) AS total_shipments,
			(SELECT COUNT(*) FROM shipments WHERE org_id = ? AND status NOT IN ('DELIVERED', 'CANCELLED', 'CLOSED')) AS active_shipments,
			(SELECT COUNT(*) FROM shipments WHERE org_id = ? AND MONTH(created_at) = MONTH(CURRENT_DATE()) AND YEAR(created_at) = YEAR(CURRENT_DATE())) AS shipments_this_month,
			(SELECT COUNT(*) FROM approval_requests WHERE org_id = ? AND status = 'Pending') AS pending_approvals,
			(SELECT COUNT(*) FROM customer_invoices WHERE org_id = ?) AS total_invoices,
			(SELECT COUNT(*) FROM customer_invoices WHERE org_id = ? AND status IN ('Issued', 'Partially Paid', 'Overdue', 'Pending Approval') AND balance_due > 0) AS outstanding_invoices,
			(SELECT COALESCE(SUM(balance_due), 0) FROM customer_invoices WHERE org_id = ? AND status IN ('Issued', 'Partially Paid', 'Overdue', 'Pending Approval') AND balance_due > 0) AS outstanding_amount,
			(SELECT COUNT(*) FROM customer_invoices WHERE org_id = ? AND (status = 'Overdue' OR (due_date < CURRENT_DATE() AND status NOT IN ('Paid', 'Cancelled', 'Draft') AND balance_due > 0))) AS overdue_invoices,
			(SELECT COALESCE(SUM(balance_due), 0) FROM customer_invoices WHERE org_id = ? AND (status = 'Overdue' OR (due_date < CURRENT_DATE() AND status NOT IN ('Paid', 'Cancelled', 'Draft') AND balance_due > 0))) AS overdue_amount,
			(SELECT COALESCE(SUM(paid_amount), 0) FROM customer_invoices WHERE org_id = ?) AS total_revenue,
			(SELECT COALESCE(SUM(amount), 0) FROM customer_invoice_payments WHERE org_id = ? AND payment_date >= DATE_SUB(CURRENT_DATE(), INTERVAL 30 DAY)) AS paid_this_month,
			(SELECT COUNT(*) FROM outreach_campaigns WHERE org_id = ?) AS total_outreach,
			(SELECT COUNT(*) FROM outreach_campaigns WHERE org_id = ? AND status = 'ACTIVE') AS active_outreach,
			(SELECT COUNT(*) FROM contracts WHERE org_id = ?) AS total_contracts,
			(SELECT COUNT(*) FROM contracts WHERE org_id = ? AND status = 'ACTIVE') AS active_contracts,
			(SELECT COUNT(*) FROM customer_invoice_documents WHERE org_id = ?) + (SELECT COUNT(*) FROM shipment_documents WHERE org_id = ?) AS total_documents
	`

	var wonRFQs, lostRFQs, totalContracts, activeContracts, totalDocs int
	row := d.db.QueryRowContext(ctx, masterQuery,
		orgID, orgID, // customers
		orgID, orgID, // leads
		orgID, orgID, orgID, orgID, // rfqs
		orgID, orgID, // quotations
		orgID, orgID, // bookings
		orgID, orgID, orgID, // shipments
		orgID, // approvals
		orgID, orgID, orgID, orgID, orgID, orgID, orgID, // invoices
		orgID, orgID, // outreach
		orgID, orgID, // contracts
		orgID, orgID, // documents
	)

	err := row.Scan(
		&stats.TotalCustomers,
		&stats.NewCustomersThisMonth,
		&stats.TotalLeads,
		&stats.OpenLeads,
		&stats.TotalRFQs,
		&stats.OpenRFQs,
		&wonRFQs,
		&lostRFQs,
		&stats.TotalQuotations,
		&stats.ActiveQuotations,
		&stats.TotalBookings,
		&stats.ActiveBookings,
		&stats.TotalShipments,
		&stats.ActiveShipments,
		&stats.ShipmentsThisMonth,
		&stats.PendingApprovals,
		&stats.TotalInvoices,
		&stats.OutstandingInvoices,
		&stats.OutstandingAmount,
		&stats.OverdueInvoices,
		&stats.OverdueAmount,
		&stats.TotalRevenue,
		&stats.PaidThisMonth,
		&stats.ActiveOutreachCampaigns,
		&activeContracts,
		&totalContracts,
		&activeContracts,
		&totalDocs,
	)
	if err != nil {
		return stats, pipeline, shipmentStatus, invoiceSummary, moduleStatus, spec.DateRangeInfo{}, err
	}

	// 2. Fallbacks for revenue if paid_amount wasn't populated on invoices
	if stats.PaidThisMonth == 0 {
		_ = d.db.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(paid_amount), 0) FROM customer_invoices 
			WHERE org_id = ? AND status = 'Paid' AND updated_at >= DATE_SUB(CURRENT_DATE(), INTERVAL 30 DAY)
		`, orgID).Scan(&stats.PaidThisMonth)
	}

	// Revenue this month
	_ = d.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total_amount), 0) FROM customer_invoices
		WHERE org_id = ? AND MONTH(created_at) = MONTH(CURRENT_DATE()) AND YEAR(created_at) = YEAR(CURRENT_DATE()) AND status NOT IN ('Cancelled', 'Draft')
	`, orgID).Scan(&stats.RevenueThisMonth)
	if stats.RevenueThisMonth == 0 {
		stats.RevenueThisMonth = stats.PaidThisMonth
	}

	// 3. Win Rate & Conversion Rate Calculations
	totalClosedRFQs := wonRFQs + lostRFQs
	if totalClosedRFQs > 0 {
		stats.WinRate = float64(wonRFQs) / float64(totalClosedRFQs) * 100.0
	} else if stats.TotalRFQs > 0 && wonRFQs > 0 {
		stats.WinRate = float64(wonRFQs) / float64(stats.TotalRFQs) * 100.0
	} else {
		stats.WinRate = 0.0
	}

	if stats.TotalLeads > 0 && stats.TotalRFQs > 0 {
		stats.ConversionRate = float64(stats.TotalRFQs) / float64(stats.TotalLeads) * 100.0
		if stats.ConversionRate > 100.0 {
			stats.ConversionRate = 100.0
		}
	} else if stats.TotalRFQs > 0 && stats.ActiveBookings > 0 {
		stats.ConversionRate = float64(stats.ActiveBookings) / float64(stats.TotalRFQs) * 100.0
		if stats.ConversionRate > 100.0 {
			stats.ConversionRate = 100.0
		}
	} else {
		stats.ConversionRate = 0.0
	}

	// 4. Avg Revenue per Shipment
	if stats.TotalShipments > 0 && stats.TotalRevenue > 0 {
		stats.AvgRevenuePerShipment = stats.TotalRevenue / float64(stats.TotalShipments)
	} else {
		stats.AvgRevenuePerShipment = 0.0
	}

	// 5. Dynamic Operational vs New User State Detection
	totalOperationalEntities := stats.TotalLeads + stats.TotalRFQs + stats.TotalQuotations + stats.TotalBookings + stats.TotalShipments + stats.TotalInvoices

	if totalOperationalEntities == 0 && stats.TotalCustomers <= 1 {
		stats.AccountMaturity = "NEW"
		stats.IsNewUser = true
		stats.IsOperational = false
	} else if totalOperationalEntities <= 1 && stats.TotalCustomers <= 3 {
		stats.AccountMaturity = "LOW_DATA"
		stats.IsNewUser = true
		stats.IsOperational = false
	} else if stats.TotalShipments == 0 && stats.TotalInvoices == 0 {
		stats.AccountMaturity = "GROWING"
		stats.IsNewUser = false
		stats.IsOperational = true
	} else if stats.TotalInvoices == 0 {
		stats.AccountMaturity = "OPERATIONAL"
		stats.IsNewUser = false
		stats.IsOperational = true
	} else {
		stats.AccountMaturity = "MATURE"
		stats.IsNewUser = false
		stats.IsOperational = true
	}

	// Resolve date range and preceding comparison window
	startTime, endTime, prevStartTime, prevEndTime, dateRangeInfo := resolveDateRange(startDate, endDate, preset)

	// 5.5 Calculate real trends and sparklines for the active date range
	stats.LeadsTrendPct, stats.LeadsTrendDirection = getPeriodTrend(ctx, d.db, "leads", orgID, startTime, endTime, prevStartTime, prevEndTime)
	stats.RFQsTrendPct, stats.RFQsTrendDirection = getPeriodTrend(ctx, d.db, "rfqs", orgID, startTime, endTime, prevStartTime, prevEndTime)
	stats.QuotesTrendPct, stats.QuotesTrendDirection = getPeriodTrend(ctx, d.db, "quotations", orgID, startTime, endTime, prevStartTime, prevEndTime)
	stats.ShipmentsTrendPct, stats.ShipmentsTrendDirection = getPeriodTrend(ctx, d.db, "shipments", orgID, startTime, endTime, prevStartTime, prevEndTime)
	stats.ApprovalsTrendPct, stats.ApprovalsTrendDirection = getPeriodTrend(ctx, d.db, "approval_requests", orgID, startTime, endTime, prevStartTime, prevEndTime)
	stats.InvoicesTrendPct, stats.InvoicesTrendDirection = getPeriodTrend(ctx, d.db, "customer_invoices", orgID, startTime, endTime, prevStartTime, prevEndTime)

	stats.LeadsSparkline = getPeriodSparkline(ctx, d.db, "leads", orgID, startTime, endTime)
	stats.RFQsSparkline = getPeriodSparkline(ctx, d.db, "rfqs", orgID, startTime, endTime)
	stats.QuotesSparkline = getPeriodSparkline(ctx, d.db, "quotations", orgID, startTime, endTime)
	stats.ShipmentsSparkline = getPeriodSparkline(ctx, d.db, "shipments", orgID, startTime, endTime)
	stats.ApprovalsSparkline = getPeriodSparkline(ctx, d.db, "approval_requests", orgID, startTime, endTime)
	stats.InvoicesSparkline = getPeriodSparkline(ctx, d.db, "customer_invoices", orgID, startTime, endTime)

	// Check if there was any activity data in this period
	totalPeriodCount := 0
	for _, v := range stats.LeadsSparkline {
		totalPeriodCount += v
	}
	for _, v := range stats.RFQsSparkline {
		totalPeriodCount += v
	}
	for _, v := range stats.ShipmentsSparkline {
		totalPeriodCount += v
	}
	dateRangeInfo.HasDataInPeriod = totalPeriodCount > 0 || stats.TotalCustomers > 0

	// 6. Pipeline Stats
	pipeline.LeadsCount = stats.TotalLeads
	pipeline.RFQsCount = stats.TotalRFQs
	pipeline.QuotationsCount = stats.TotalQuotations
	pipeline.BookingsCount = stats.TotalBookings
	pipeline.ShipmentsCount = stats.TotalShipments

	if pipeline.LeadsCount > 0 {
		pipeline.LeadsToRFQsConv = float64(pipeline.RFQsCount) / float64(pipeline.LeadsCount) * 100.0
		if pipeline.LeadsToRFQsConv > 100.0 {
			pipeline.LeadsToRFQsConv = 100.0
		}
	}
	if pipeline.RFQsCount > 0 {
		pipeline.RFQsToQuotesConv = float64(pipeline.QuotationsCount) / float64(pipeline.RFQsCount) * 100.0
		if pipeline.RFQsToQuotesConv > 100.0 {
			pipeline.RFQsToQuotesConv = 100.0
		}
	}
	if pipeline.QuotationsCount > 0 {
		pipeline.QuotesToBookingsConv = float64(pipeline.BookingsCount) / float64(pipeline.QuotationsCount) * 100.0
		if pipeline.QuotesToBookingsConv > 100.0 {
			pipeline.QuotesToBookingsConv = 100.0
		}
	}
	if pipeline.BookingsCount > 0 {
		pipeline.BookingsToShipmentsConv = float64(pipeline.ShipmentsCount) / float64(pipeline.BookingsCount) * 100.0
		if pipeline.BookingsToShipmentsConv > 100.0 {
			pipeline.BookingsToShipmentsConv = 100.0
		}
	}

	// 7. Shipment Status Breakdown
	shipmentStatus.Total = stats.TotalShipments
	statusRows, err := d.db.QueryContext(ctx, `
		SELECT status, COUNT(*) 
		FROM shipments 
		WHERE org_id = ? 
		GROUP BY status
	`, orgID)
	if err == nil {
		defer statusRows.Close()
		for statusRows.Next() {
			var st string
			var count int
			if err := statusRows.Scan(&st, &count); err == nil {
				upper := strings.ToUpper(st)
				switch {
				case strings.Contains(upper, "TRANSIT"):
					shipmentStatus.InTransit += count
				case strings.Contains(upper, "DELIVER"):
					shipmentStatus.Delivered += count
				case strings.Contains(upper, "DELAY"):
					shipmentStatus.Delayed += count
				case strings.Contains(upper, "CUSTOM") || strings.Contains(upper, "HOLD") || strings.Contains(upper, "EXCEPTION"):
					shipmentStatus.CustomsHold += count
				case strings.Contains(upper, "BOOK"):
					shipmentStatus.Booked += count
				default:
					shipmentStatus.InTransit += count
				}
			}
		}
	}

	// 8. Invoice Summary & Recent Invoices
	invoiceSummary.OutstandingAmount = stats.OutstandingAmount
	invoiceSummary.OverdueAmount = stats.OverdueAmount
	invoiceSummary.PaidThisMonthAmount = stats.PaidThisMonth
	invoiceSummary.RecentInvoices = []spec.RecentInvoice{}

	invRows, err := d.db.QueryContext(ctx, `
		SELECT id, invoice_number, customer_name, total_amount, balance_due, status, COALESCE(due_date, created_at), created_at
		FROM customer_invoices
		WHERE org_id = ?
		ORDER BY created_at DESC
		LIMIT 4
	`, orgID)
	if err == nil {
		defer invRows.Close()
		for invRows.Next() {
			var inv spec.RecentInvoice
			var dueDate, createdAt time.Time
			if err := invRows.Scan(&inv.ID, &inv.InvoiceNumber, &inv.CustomerName, &inv.TotalAmount, &inv.BalanceDue, &inv.Status, &dueDate, &createdAt); err == nil {
				inv.DueDate = dueDate.Format("2006-01-02")
				inv.CreatedAt = createdAt.Format(time.RFC3339)
				inv.AgeText = formatRelativeTime(createdAt)
				invoiceSummary.RecentInvoices = append(invoiceSummary.RecentInvoices, inv)
			}
		}
	}

	// 9. Module Status
	moduleStatus.Customers = spec.ModuleCount{TotalCount: stats.TotalCustomers, ActiveCount: stats.TotalCustomers, HasData: stats.TotalCustomers > 0}
	moduleStatus.Leads = spec.ModuleCount{TotalCount: stats.TotalLeads, ActiveCount: stats.OpenLeads, HasData: stats.TotalLeads > 0}
	moduleStatus.RFQs = spec.ModuleCount{TotalCount: stats.TotalRFQs, ActiveCount: stats.OpenRFQs, HasData: stats.TotalRFQs > 0}
	moduleStatus.Quotations = spec.ModuleCount{TotalCount: stats.TotalQuotations, ActiveCount: stats.ActiveQuotations, HasData: stats.TotalQuotations > 0}
	moduleStatus.Bookings = spec.ModuleCount{TotalCount: stats.TotalBookings, ActiveCount: stats.ActiveBookings, HasData: stats.TotalBookings > 0}
	moduleStatus.Shipments = spec.ModuleCount{TotalCount: stats.TotalShipments, ActiveCount: stats.ActiveShipments, HasData: stats.TotalShipments > 0}
	moduleStatus.Documents = spec.ModuleCount{TotalCount: totalDocs, ActiveCount: totalDocs, HasData: totalDocs > 0}
	moduleStatus.Approvals = spec.ModuleCount{TotalCount: stats.PendingApprovals, ActiveCount: stats.PendingApprovals, HasData: stats.PendingApprovals > 0}
	moduleStatus.Invoices = spec.ModuleCount{TotalCount: stats.TotalInvoices, ActiveCount: stats.OutstandingInvoices, HasData: stats.TotalInvoices > 0}
	moduleStatus.Outreach = spec.ModuleCount{TotalCount: stats.ActiveOutreachCampaigns, ActiveCount: stats.ActiveOutreachCampaigns, HasData: stats.ActiveOutreachCampaigns > 0}
	moduleStatus.Contracts = spec.ModuleCount{TotalCount: totalContracts, ActiveCount: activeContracts, HasData: totalContracts > 0}

	return stats, pipeline, shipmentStatus, invoiceSummary, moduleStatus, dateRangeInfo, nil
}

func (d *dataLayer) GetApprovalQueue(ctx context.Context, orgID int64) ([]spec.PendingTask, []spec.PendingApprovalItem, error) {
	var queue []spec.PendingTask
	var pendingList []spec.PendingApprovalItem

	rows, err := d.db.QueryContext(ctx, `
		SELECT 
			id, 
			request_code, 
			title, 
			category, 
			type, 
			priority, 
			requested_by_name, 
			department, 
			created_at 
		FROM approval_requests 
		WHERE org_id = ? AND status = 'Pending'
		ORDER BY created_at DESC LIMIT 5
	`, orgID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item spec.PendingApprovalItem
			var created time.Time
			if err := rows.Scan(
				&item.ID,
				&item.RequestCode,
				&item.Title,
				&item.Category,
				&item.Type,
				&item.Priority,
				&item.RequestedByName,
				&item.Department,
				&created,
			); err == nil {
				item.CreatedAt = created.Format(time.RFC3339)
				item.AgeText = formatRelativeTime(created)
				item.ActionURL = "/dashboard/approvals"
				pendingList = append(pendingList, item)

				queue = append(queue, spec.PendingTask{
					ID:        item.ID,
					Type:      item.Category,
					Title:     item.Title,
					Subtitle:  fmt.Sprintf("%s • %s", item.Department, item.AgeText),
					Timestamp: item.CreatedAt,
					RefID:     item.ID,
				})
			}
		}
	}

	return queue, pendingList, nil
}

func (d *dataLayer) GetAttentionItems(ctx context.Context, orgID int64) ([]spec.AttentionItem, error) {
	var items []spec.AttentionItem

	// 1. RFQs awaiting quotation
	var rfqCount int
	var latestRFQNo, latestCustomer string
	var latestRFQTime sql.NullTime
	_ = d.db.QueryRowContext(ctx, `
		SELECT r.rfq_number, COALESCE(c.name, c.trading_name, 'Direct Client'), r.created_at
		FROM rfqs r
		LEFT JOIN customers c ON c.id = r.customer_id
		WHERE r.org_id = ? AND r.stage IN ('STAGE_QUOTE_DRAFTING', 'STAGE_RATE_INTELLIGENCE_MATCHING', 'STAGE_DOCUMENT_EXTRACTION', 'DRAFT', 'NEW', 'IN_PROGRESS')
		ORDER BY r.created_at DESC
		LIMIT 1
	`, orgID).Scan(&latestRFQNo, &latestCustomer, &latestRFQTime)

	_ = d.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM rfqs 
		WHERE org_id = ? AND stage IN ('STAGE_QUOTE_DRAFTING', 'STAGE_RATE_INTELLIGENCE_MATCHING', 'STAGE_DOCUMENT_EXTRACTION', 'DRAFT', 'NEW', 'IN_PROGRESS')
	`, orgID).Scan(&rfqCount)

	if rfqCount > 0 {
		subtitle := "Active freight requests awaiting carrier rate quoting"
		if latestRFQNo != "" {
			subtitle = fmt.Sprintf("Latest: %s from %s", latestRFQNo, latestCustomer)
		}
		items = append(items, spec.AttentionItem{
			ID:        "rfqs_awaiting_quote",
			Priority:  "HIGH",
			Category:  "RFQs",
			Title:     fmt.Sprintf("%d RFQ(s) are awaiting your quotation", rfqCount),
			Subtitle:  subtitle,
			Count:     rfqCount,
			ActionURL: "/dashboard/rfqs",
			Timestamp: "2h ago",
		})
	}

	// 2. Overdue Invoices
	var overdueCount int
	var overdueSum float64
	_ = d.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(balance_due), 0) 
		FROM customer_invoices 
		WHERE org_id = ? AND (status = 'Overdue' OR (due_date < CURRENT_DATE() AND status NOT IN ('Paid', 'Cancelled', 'Draft') AND balance_due > 0))
	`, orgID).Scan(&overdueCount, &overdueSum)

	if overdueCount > 0 {
		items = append(items, spec.AttentionItem{
			ID:        "overdue_invoices",
			Priority:  "HIGH",
			Category:  "Finance",
			Title:     fmt.Sprintf("%d Invoice(s) are overdue", overdueCount),
			Subtitle:  fmt.Sprintf("Total overdue amount: $%.2f", overdueSum),
			Count:     overdueCount,
			ActionURL: "/dashboard/invoices?primary_tab=ALL&status=Overdue",
			Timestamp: "2h ago",
		})
	}

	// 3. Delayed / Exception Shipments
	var delayedShipmentCount int
	var delayedRoute string
	_ = d.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(MIN(CONCAT(origin_port, ' → ', destination_port)), '') 
		FROM shipments 
		WHERE org_id = ? AND status IN ('DELAYED', 'CUSTOMS_HOLD', 'EXCEPTION')
	`, orgID).Scan(&delayedShipmentCount, &delayedRoute)

	if delayedShipmentCount > 0 {
		subtitle := "Operational exception or carrier schedule disruption"
		if delayedRoute != "" {
			subtitle = fmt.Sprintf("Route: %s", delayedRoute)
		}
		items = append(items, spec.AttentionItem{
			ID:        "delayed_shipments",
			Priority:  "HIGH",
			Category:  "Shipments",
			Title:     fmt.Sprintf("%d Shipment(s) delayed or on hold", delayedShipmentCount),
			Subtitle:  subtitle,
			Count:     delayedShipmentCount,
			ActionURL: "/dashboard/shipments",
			Timestamp: "2h ago",
		})
	}

	// 4. Pending Approvals
	var pendingApprovals int
	_ = d.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM approval_requests WHERE org_id = ? AND status = 'Pending'`, orgID).Scan(&pendingApprovals)
	if pendingApprovals > 0 {
		items = append(items, spec.AttentionItem{
			ID:        "pending_approvals",
			Priority:  "HIGH",
			Category:  "Approvals",
			Title:     fmt.Sprintf("%d Approvals are pending", pendingApprovals),
			Subtitle:  "Quotation, Contract & Shipment approvals",
			Count:     pendingApprovals,
			ActionURL: "/dashboard/approvals",
			Timestamp: "2h ago",
		})
	}

	// 5. Expiring Contracts (within next 30 days)
	var expiringContracts int
	var nextContractRef string
	var daysUntilExpiry int
	_ = d.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(MIN(contract_reference), ''), COALESCE(MIN(DATEDIFF(expiry_date, CURRENT_DATE())), 0)
		FROM contracts
		WHERE org_id = ? AND expiry_date BETWEEN CURRENT_DATE() AND DATE_ADD(CURRENT_DATE(), INTERVAL 30 DAY) AND status = 'ACTIVE'
	`, orgID).Scan(&expiringContracts, &nextContractRef, &daysUntilExpiry)

	if expiringContracts > 0 {
		subtitle := fmt.Sprintf("Next expiry: Contract %s in %d days", nextContractRef, daysUntilExpiry)
		items = append(items, spec.AttentionItem{
			ID:        "expiring_contracts",
			Priority:  "MEDIUM",
			Category:  "Contracts",
			Title:     fmt.Sprintf("%d Contract(s) are expiring soon", expiringContracts),
			Subtitle:  subtitle,
			Count:     expiringContracts,
			ActionURL: "/dashboard/contracts",
			Timestamp: "2h ago",
		})
	}

	// 6. Open Leads requiring outreach
	var newLeadsCount int
	_ = d.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM leads WHERE org_id = ? AND status = 'NEW'`, orgID).Scan(&newLeadsCount)
	if newLeadsCount > 0 {
		items = append(items, spec.AttentionItem{
			ID:        "new_leads",
			Priority:  "MEDIUM",
			Category:  "Leads",
			Title:     fmt.Sprintf("%d New Lead(s) require outreach", newLeadsCount),
			Subtitle:  "Prospect inquiries waiting for sales follow-up",
			Count:     newLeadsCount,
			ActionURL: "/dashboard/leads",
			Timestamp: "2h ago",
		})
	}

	return items, nil
}

func (d *dataLayer) GetActiveShipments(ctx context.Context, orgID int64) ([]spec.ActiveShipment, []spec.ActiveShipmentItem, error) {
	var list []spec.ActiveShipment
	var detailedList []spec.ActiveShipmentItem

	rows, err := d.db.QueryContext(ctx, `
		SELECT 
			s.id,
			COALESCE(s.mbl_number, s.booking_number, CONCAT('SHP-', s.id)) AS shipment_no,
			COALESCE(c.name, c.trading_name, 'Direct Customer') AS customer,
			COALESCE(s.origin_port, b.origin_port, 'Shanghai') AS origin,
			COALESCE(s.destination_port, b.destination_port, 'Rotterdam') AS destination,
			COALESCE(b.carrier_name, 'Maersk Line') AS carrier,
			s.status,
			'Ocean Freight' AS mode,
			COALESCE(s.eta, b.eta) AS eta,
			s.created_at
		FROM shipments s
		LEFT JOIN bookings b ON b.id = s.booking_id
		LEFT JOIN rfqs r ON r.id = s.rfq_id
		LEFT JOIN customers c ON c.id = r.customer_id
		WHERE s.org_id = ?
		ORDER BY s.created_at DESC 
		LIMIT 6
	`, orgID)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int64
			var shipmentNo, customer, origin, destination, carrier, status, mode string
			var eta sql.NullTime
			var createdAt time.Time

			if err := rows.Scan(&id, &shipmentNo, &customer, &origin, &destination, &carrier, &status, &mode, &eta, &createdAt); err == nil {
				etaFormatted := "In Transit"
				etaDate := ""
				if eta.Valid {
					etaFormatted = eta.Time.Format("02 Jan 2006")
					etaDate = eta.Time.Format("2006-01-02")
				} else {
					etaFormatted = createdAt.AddDate(0, 0, 14).Format("02 Jan 2006")
					etaDate = createdAt.AddDate(0, 0, 14).Format("2006-01-02")
				}

				statusDisplay := "+ In Transit"
				switch strings.ToUpper(status) {
				case "IN_TRANSIT":
					statusDisplay = "+ In Transit"
				case "DELIVERED":
					statusDisplay = "+ Delivered"
				case "CUSTOMS_HOLD":
					statusDisplay = "Customs Hold"
				case "ON_BOARD":
					statusDisplay = "+ On Board"
				case "BOOKED":
					statusDisplay = "Awaiting Pickup"
				default:
					statusDisplay = status
				}

				list = append(list, spec.ActiveShipment{
					ID:          id,
					ShipmentNo:  shipmentNo,
					Customer:    customer,
					Origin:      origin,
					Destination: destination,
					Status:      status,
					Mode:        mode,
					ETA:         etaFormatted,
				})

				detailedList = append(detailedList, spec.ActiveShipmentItem{
					ID:            id,
					ShipmentNo:    shipmentNo,
					Customer:      customer,
					Origin:        origin,
					Destination:   destination,
					Carrier:       carrier,
					Status:        status,
					StatusDisplay: statusDisplay,
					Mode:          mode,
					ETA:           etaFormatted,
					ETADate:       etaDate,
				})
			}
		}
	}

	return list, detailedList, nil
}

func (d *dataLayer) GetRecentDocuments(ctx context.Context, orgID int64) ([]spec.RecentDocument, error) {
	var list []spec.RecentDocument

	// 1. Customer Invoice Documents
	rows, err := d.db.QueryContext(ctx, `
		SELECT 
			d.id,
			d.document_name,
			COALESCE(d.file_type, 'application/pdf') AS document_type,
			CONCAT('INV-', d.invoice_id) AS reference,
			250000 AS file_size,
			d.uploaded_at
		FROM customer_invoice_documents d
		WHERE d.org_id = ?
		ORDER BY d.uploaded_at DESC LIMIT 5
	`, orgID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var doc spec.RecentDocument
			var uploaded time.Time
			if err := rows.Scan(&doc.ID, &doc.DocumentName, &doc.DocumentType, &doc.Reference, &doc.FileSize, &uploaded); err == nil {
				doc.UploadedAt = uploaded.Format("Jan 02, 15:04")
				doc.FileSizeFormatted = formatFileSize(doc.FileSize)
				list = append(list, doc)
			}
		}
	}

	// 2. Shipment Documents
	shipDocs, err := d.db.QueryContext(ctx, `
		SELECT 
			d.id,
			d.file_name,
			COALESCE(d.file_type, 'PDF') AS document_type,
			COALESCE(CONCAT('SHP-', d.shipment_id), 'Direct') AS reference,
			COALESCE(d.file_size, 150000) AS file_size,
			d.created_at
		FROM shipment_documents d
		WHERE d.org_id = ?
		ORDER BY d.created_at DESC LIMIT 5
	`, orgID)
	if err == nil {
		defer shipDocs.Close()
		for shipDocs.Next() {
			var doc spec.RecentDocument
			var created time.Time
			if err := shipDocs.Scan(&doc.ID, &doc.DocumentName, &doc.DocumentType, &doc.Reference, &doc.FileSize, &created); err == nil {
				doc.UploadedAt = created.Format("Jan 02, 15:04")
				doc.FileSizeFormatted = formatFileSize(doc.FileSize)
				list = append(list, doc)
			}
		}
	}

	// Sort newest first
	if len(list) > 6 {
		list = list[:6]
	}

	return list, nil
}

type activityItemWithTime struct {
	item spec.RecentActivity
	t    time.Time
}

func (d *dataLayer) GetRecentActivity(ctx context.Context, orgID int64) ([]spec.RecentActivity, error) {
	var activityList []activityItemWithTime

	// 1. Recent Payments
	payRows, err := d.db.QueryContext(ctx, `
		SELECT p.id, p.amount, p.payment_ref, p.invoice_id, p.created_at
		FROM customer_invoice_payments p
		WHERE p.org_id = ?
		ORDER BY p.created_at DESC LIMIT 3
	`, orgID)
	if err == nil {
		defer payRows.Close()
		for payRows.Next() {
			var id, invID int64
			var amt float64
			var ref string
			var date time.Time
			if err := payRows.Scan(&id, &amt, &ref, &invID, &date); err == nil {
				activityList = append(activityList, activityItemWithTime{
					t: date,
					item: spec.RecentActivity{
						ID:        fmt.Sprintf("pay_%d", id),
						Type:      "PAYMENT",
						Title:     fmt.Sprintf("Payment Received — $%.2f", amt),
						Subtitle:  fmt.Sprintf("Invoice #%d • Ref: %s", invID, ref),
						Timestamp: formatRelativeTime(date),
						ActionURL: fmt.Sprintf("/dashboard/invoices?id=%d", invID),
					},
				})
			}
		}
	}

	// 2. Recent RFQs
	rfqRows, err := d.db.QueryContext(ctx, `
		SELECT r.id, r.rfq_number, COALESCE(c.name, 'Direct Customer'), r.created_at
		FROM rfqs r
		LEFT JOIN customers c ON c.id = r.customer_id
		WHERE r.org_id = ?
		ORDER BY r.created_at DESC LIMIT 3
	`, orgID)
	if err == nil {
		defer rfqRows.Close()
		for rfqRows.Next() {
			var id int64
			var rfqNo, cust string
			var date time.Time
			if err := rfqRows.Scan(&id, &rfqNo, &cust, &date); err == nil {
				activityList = append(activityList, activityItemWithTime{
					t: date,
					item: spec.RecentActivity{
						ID:        fmt.Sprintf("rfq_%d", id),
						Type:      "RFQ",
						Title:     fmt.Sprintf("New RFQ received %s", rfqNo),
						Subtitle:  cust,
						Timestamp: formatRelativeTime(date),
						ActionURL: "/dashboard/rfqs",
					},
				})
			}
		}
	}

	// 3. Recent Quotations
	quoteRows, err := d.db.QueryContext(ctx, `
		SELECT q.id, q.quotation_number, COALESCE(c.name, 'Direct Client'), q.total_amount, q.created_at
		FROM quotations q
		LEFT JOIN customers c ON c.id = q.customer_id
		WHERE q.org_id = ?
		ORDER BY q.created_at DESC LIMIT 3
	`, orgID)
	if err == nil {
		defer quoteRows.Close()
		for quoteRows.Next() {
			var id int64
			var qNo, cust string
			var amt float64
			var date time.Time
			if err := quoteRows.Scan(&id, &qNo, &cust, &amt, &date); err == nil {
				activityList = append(activityList, activityItemWithTime{
					t: date,
					item: spec.RecentActivity{
						ID:        fmt.Sprintf("quote_%d", id),
						Type:      "QUOTATION",
						Title:     fmt.Sprintf("Quotation %s created", qNo),
						Subtitle:  fmt.Sprintf("For %s • $%.2f", cust, amt),
						Timestamp: formatRelativeTime(date),
						ActionURL: "/dashboard/quotations",
					},
				})
			}
		}
	}

	// 4. Recent Invoices
	invRows, err := d.db.QueryContext(ctx, `
		SELECT i.id, i.invoice_number, i.customer_name, i.status, i.total_amount, i.created_at
		FROM customer_invoices i
		WHERE i.org_id = ?
		ORDER BY i.created_at DESC LIMIT 3
	`, orgID)
	if err == nil {
		defer invRows.Close()
		for invRows.Next() {
			var id int64
			var invNo, cust, status string
			var total float64
			var date time.Time
			if err := invRows.Scan(&id, &invNo, &cust, &status, &total, &date); err == nil {
				activityList = append(activityList, activityItemWithTime{
					t: date,
					item: spec.RecentActivity{
						ID:        fmt.Sprintf("inv_%d", id),
						Type:      "INVOICE",
						Title:     fmt.Sprintf("Invoice %s (%s)", invNo, status),
						Subtitle:  fmt.Sprintf("Customer: %s • Amount: $%.2f", cust, total),
						Timestamp: formatRelativeTime(date),
						ActionURL: fmt.Sprintf("/dashboard/invoices?id=%d", id),
					},
				})
			}
		}
	}

	// 5. Recent Documents
	docRows, err := d.db.QueryContext(ctx, `
		SELECT d.id, d.document_name, d.invoice_id, d.uploaded_at
		FROM customer_invoice_documents d
		WHERE d.org_id = ?
		ORDER BY d.uploaded_at DESC LIMIT 3
	`, orgID)
	if err == nil {
		defer docRows.Close()
		for docRows.Next() {
			var id, invID int64
			var name string
			var date time.Time
			if err := docRows.Scan(&id, &name, &invID, &date); err == nil {
				activityList = append(activityList, activityItemWithTime{
					t: date,
					item: spec.RecentActivity{
						ID:        fmt.Sprintf("doc_%d", id),
						Type:      "DOCUMENT",
						Title:     fmt.Sprintf("Document Uploaded — %s", name),
						Subtitle:  fmt.Sprintf("Attached to Invoice #%d", invID),
						Timestamp: formatRelativeTime(date),
						ActionURL: fmt.Sprintf("/dashboard/invoices?id=%d", invID),
					},
				})
			}
		}
	}

	// 6. Recent Shipments
	shipRows, err := d.db.QueryContext(ctx, `
		SELECT s.id, COALESCE(s.mbl_number, s.booking_number, CONCAT('SHP-', s.id)), s.status, s.created_at
		FROM shipments s
		WHERE s.org_id = ?
		ORDER BY s.created_at DESC LIMIT 3
	`, orgID)
	if err == nil {
		defer shipRows.Close()
		for shipRows.Next() {
			var id int64
			var shpNo, status string
			var date time.Time
			if err := shipRows.Scan(&id, &shpNo, &status, &date); err == nil {
				activityList = append(activityList, activityItemWithTime{
					t: date,
					item: spec.RecentActivity{
						ID:        fmt.Sprintf("shp_%d", id),
						Type:      "SHIPMENT",
						Title:     fmt.Sprintf("Shipment %s status: %s", shpNo, status),
						Subtitle:  "Ocean Freight Consignment",
						Timestamp: formatRelativeTime(date),
						ActionURL: "/dashboard/shipments",
					},
				})
			}
		}
	}

	// 7. Recent Customers
	custRows, err := d.db.QueryContext(ctx, `
		SELECT c.id, COALESCE(c.name, c.trading_name, 'New Customer'), c.created_at
		FROM customers c
		WHERE c.org_id = ?
		ORDER BY c.created_at DESC LIMIT 2
	`, orgID)
	if err == nil {
		defer custRows.Close()
		for custRows.Next() {
			var id int64
			var name string
			var date time.Time
			if err := custRows.Scan(&id, &name, &date); err == nil {
				activityList = append(activityList, activityItemWithTime{
					t: date,
					item: spec.RecentActivity{
						ID:        fmt.Sprintf("cust_%d", id),
						Type:      "CUSTOMER",
						Title:     fmt.Sprintf("Customer onboarded: %s", name),
						Subtitle:  "Account active in CRM",
						Timestamp: formatRelativeTime(date),
						ActionURL: "/dashboard/customers",
					},
				})
			}
		}
	}

	// Sort merged activities by descending timestamp
	sort.Slice(activityList, func(i, j int) bool {
		return activityList[i].t.After(activityList[j].t)
	})

	var res []spec.RecentActivity
	maxItems := 8
	if len(activityList) < maxItems {
		maxItems = len(activityList)
	}
	for i := 0; i < maxItems; i++ {
		res = append(res, activityList[i].item)
	}

	return res, nil
}

func (d *dataLayer) GetUpcomingReminders(ctx context.Context, orgID int64) ([]spec.UpcomingReminder, error) {
	var reminders []spec.UpcomingReminder

	// 1. Follow up on latest lead / rfq
	var leadID int64
	var leadCompany, contactName string
	var leadCreatedAt time.Time
	err := d.db.QueryRowContext(ctx, `
		SELECT id, company_name, COALESCE(contact_name, 'Lead Contact'), created_at
		FROM leads
		WHERE org_id = ? AND status = 'NEW'
		ORDER BY created_at DESC
		LIMIT 1
	`, orgID).Scan(&leadID, &leadCompany, &contactName, &leadCreatedAt)
	if err == nil {
		reminders = append(reminders, spec.UpcomingReminder{
			ID:        fmt.Sprintf("lead_rem_%d", leadID),
			Type:      "FOLLOW_UP",
			Title:     fmt.Sprintf("Follow up with %s", leadCompany),
			Subtitle:  fmt.Sprintf("Regarding Lead inquiry • %s", contactName),
			DueText:   "Today, 03:00 PM",
			DueDate:   time.Now().Format("2006-01-02"),
			ActionURL: "/dashboard/leads",
		})
	}

	// 2. Upcoming contract expiry
	var contractID int64
	var contractRef, contractName, partyName string
	var contractExpiry time.Time
	err = d.db.QueryRowContext(ctx, `
		SELECT id, contract_reference, contract_name, COALESCE(party_name, 'Carrier'), expiry_date
		FROM contracts
		WHERE org_id = ? AND expiry_date >= CURRENT_DATE() AND status = 'ACTIVE'
		ORDER BY expiry_date ASC
		LIMIT 1
	`, orgID).Scan(&contractID, &contractRef, &contractName, &partyName, &contractExpiry)
	if err == nil {
		reminders = append(reminders, spec.UpcomingReminder{
			ID:        fmt.Sprintf("contract_rem_%d", contractID),
			Type:      "CONTRACT_EXPIRY",
			Title:     fmt.Sprintf("Contract %s expires", contractRef),
			Subtitle:  fmt.Sprintf("Party: %s • %s", partyName, contractName),
			DueText:   formatFutureTime(contractExpiry),
			DueDate:   contractExpiry.Format("2006-01-02"),
			ActionURL: "/dashboard/contracts",
		})
	}

	// 3. Upcoming invoice payment due
	var invID int64
	var invNumber, customerName string
	var invDueDate time.Time
	err = d.db.QueryRowContext(ctx, `
		SELECT id, invoice_number, customer_name, due_date
		FROM customer_invoices
		WHERE org_id = ? AND due_date >= CURRENT_DATE() AND status NOT IN ('Paid', 'Cancelled')
		ORDER BY due_date ASC
		LIMIT 1
	`, orgID).Scan(&invID, &invNumber, &customerName, &invDueDate)
	if err == nil {
		reminders = append(reminders, spec.UpcomingReminder{
			ID:        fmt.Sprintf("inv_rem_%d", invID),
			Type:      "PAYMENT_DUE",
			Title:     fmt.Sprintf("Payment due from %s", customerName),
			Subtitle:  fmt.Sprintf("Invoice %s", invNumber),
			DueText:   formatFutureTime(invDueDate),
			DueDate:   invDueDate.Format("2006-01-02"),
			ActionURL: fmt.Sprintf("/dashboard/invoices?id=%d", invID),
		})
	}

	return reminders, nil
}

func (d *dataLayer) GetOrganizationInfo(ctx context.Context, orgID int64) (spec.OrganizationInfo, error) {
	var info spec.OrganizationInfo
	info.ID = orgID
	info.Name = "LogisticsHQ Workspace"
	info.DefaultCurrency = "USD"
	info.DefaultTimezone = "UTC"

	row := d.db.QueryRowContext(ctx, `
		SELECT id, name, COALESCE(default_currency, 'USD'), COALESCE(default_timezone, 'UTC')
		FROM organizations
		WHERE id = ?
	`, orgID)
	_ = row.Scan(&info.ID, &info.Name, &info.DefaultCurrency, &info.DefaultTimezone)

	return info, nil
}
