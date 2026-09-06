package spec

type GetMissionControlResponse struct {
	Stats             Stats                 `json:"stats"`
	Pipeline          PipelineStats         `json:"pipeline"`
	ShipmentStatus    ShipmentStatusCounts  `json:"shipment_status"`
	InvoiceSummary    InvoiceSummary        `json:"invoice_summary"`
	ApprovalQueue     []PendingTask         `json:"approval_queue"`
	PendingApprovals  []PendingApprovalItem `json:"pending_approvals"`
	AttentionItems    []AttentionItem       `json:"attention_items"`
	RecentShipments   []ActiveShipment      `json:"recent_shipments"`
	ActiveShipments   []ActiveShipmentItem  `json:"active_shipments"`
	RecentDocuments   []RecentDocument      `json:"recent_documents"`
	RecentActivity    []RecentActivity      `json:"recent_activity"`
	UpcomingReminders []UpcomingReminder    `json:"upcoming_reminders"`
	ModuleStatus      ModuleStatus          `json:"module_status"`
	Organization      OrganizationInfo      `json:"organization"`
	AIStatus          AIStatus              `json:"ai_status"`
	DateRange         DateRangeInfo         `json:"date_range"`
}

type AttentionItem struct {
	ID        string `json:"id"`
	Priority  string `json:"priority"` // "HIGH", "MEDIUM", "INFO"
	Category  string `json:"category"` // "Finance", "Approvals", "RFQs", "Shipments", "Contracts", "Leads"
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle"`
	Count     int    `json:"count"`
	ActionURL string `json:"action_url"`
	Timestamp string `json:"timestamp"`
}

type Stats struct {
	TotalCustomers          int     `json:"total_customers"`
	NewCustomersThisMonth   int     `json:"new_customers_this_month"`
	OpenLeads               int     `json:"open_leads"`
	TotalLeads              int     `json:"total_leads"`
	OpenRFQs                int     `json:"open_rfqs"`
	TotalRFQs               int     `json:"total_rfqs"`
	ActiveQuotations        int     `json:"active_quotations"`
	TotalQuotations         int     `json:"total_quotations"`
	ActiveBookings          int     `json:"active_bookings"`
	TotalBookings           int     `json:"total_bookings"`
	ActiveShipments         int     `json:"active_shipments"`
	TotalShipments          int     `json:"total_shipments"`
	ShipmentsThisMonth      int     `json:"shipments_this_month"`
	PendingApprovals        int     `json:"pending_approvals"`
	TotalInvoices           int     `json:"total_invoices"`
	OutstandingInvoices     int     `json:"outstanding_invoices"`
	OutstandingAmount       float64 `json:"outstanding_amount"`
	OverdueInvoices         int     `json:"overdue_invoices"`
	OverdueAmount           float64 `json:"overdue_amount"`
	PaidThisMonth           float64 `json:"paid_this_month"`
	TotalRevenue            float64 `json:"total_revenue"`
	RevenueThisMonth        float64 `json:"revenue_this_month"`
	AvgRevenuePerShipment   float64 `json:"avg_revenue_per_shipment"`
	WinRate                 float64 `json:"win_rate"`
	ConversionRate          float64 `json:"conversion_rate"`
	ActiveOutreachCampaigns int     `json:"active_outreach_campaigns"`
	IsNewUser               bool    `json:"is_new_user"`
	IsOperational           bool    `json:"is_operational"`
	AccountMaturity         string  `json:"account_maturity"` // "NEW", "LOW_DATA", "GROWING", "OPERATIONAL", "MATURE"

	// 7-day trends and direction ("up", "down", "neutral")
	LeadsTrendPct           float64 `json:"leads_trend_pct"`
	LeadsTrendDirection     string  `json:"leads_trend_direction"`
	RFQsTrendPct            float64 `json:"rfqs_trend_pct"`
	RFQsTrendDirection      string  `json:"rfqs_trend_direction"`
	QuotesTrendPct          float64 `json:"quotes_trend_pct"`
	QuotesTrendDirection    string  `json:"quotes_trend_direction"`
	ShipmentsTrendPct       float64 `json:"shipments_trend_pct"`
	ShipmentsTrendDirection string  `json:"shipments_trend_direction"`
	ApprovalsTrendPct       float64 `json:"approvals_trend_pct"`
	ApprovalsTrendDirection string  `json:"approvals_trend_direction"`
	InvoicesTrendPct        float64 `json:"invoices_trend_pct"`
	InvoicesTrendDirection  string  `json:"invoices_trend_direction"`

	// Real 7-day sparklines (distribution of counts across past 7 days)
	LeadsSparkline          []int   `json:"leads_sparkline"`
	RFQsSparkline           []int   `json:"rfqs_sparkline"`
	QuotesSparkline         []int   `json:"quotes_sparkline"`
	ShipmentsSparkline      []int   `json:"shipments_sparkline"`
	ApprovalsSparkline      []int   `json:"approvals_sparkline"`
	InvoicesSparkline       []int   `json:"invoices_sparkline"`
}

type PipelineStats struct {
	LeadsCount              int     `json:"leads_count"`
	RFQsCount               int     `json:"rfqs_count"`
	QuotationsCount         int     `json:"quotations_count"`
	BookingsCount           int     `json:"bookings_count"`
	ShipmentsCount          int     `json:"shipments_count"`
	LeadsToRFQsConv         float64 `json:"leads_to_rfqs_conv"`
	RFQsToQuotesConv        float64 `json:"rfqs_to_quotes_conv"`
	QuotesToBookingsConv    float64 `json:"quotes_to_bookings_conv"`
	BookingsToShipmentsConv float64 `json:"bookings_to_shipments_conv"`
}

type ShipmentStatusCounts struct {
	InTransit   int `json:"in_transit"`
	Delivered   int `json:"delivered"`
	Delayed     int `json:"delayed"`
	CustomsHold int `json:"customs_hold"`
	Booked      int `json:"booked"`
	Total       int `json:"total"`
}

type InvoiceSummary struct {
	OutstandingAmount   float64         `json:"outstanding_amount"`
	OverdueAmount       float64         `json:"overdue_amount"`
	PaidThisMonthAmount float64         `json:"paid_this_month_amount"`
	RecentInvoices      []RecentInvoice `json:"recent_invoices"`
}

type RecentInvoice struct {
	ID            int64   `json:"id"`
	InvoiceNumber string  `json:"invoice_number"`
	CustomerName  string  `json:"customer_name"`
	TotalAmount   float64 `json:"total_amount"`
	BalanceDue    float64 `json:"balance_due"`
	Status        string  `json:"status"` // "Issued", "Paid", "Overdue", "Partial", "Draft"
	DueDate       string  `json:"due_date"`
	CreatedAt     string  `json:"created_at"`
	AgeText       string  `json:"age_text"` // e.g. "2d ago"
}

type PendingApprovalItem struct {
	ID              int64  `json:"id"`
	RequestCode     string `json:"request_code"`
	Title           string `json:"title"`
	Category        string `json:"category"`
	Type            string `json:"type"`
	Priority        string `json:"priority"` // "HIGH", "MEDIUM", "LOW"
	RequestedByName string `json:"requested_by_name"`
	Department      string `json:"department"`
	CreatedAt       string `json:"created_at"`
	AgeText         string `json:"age_text"`
	ActionURL       string `json:"action_url"`
}

type PendingTask struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"` // e.g., "RFQ_QUOTE_DRAFT", "FINANCE"
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle"`
	Timestamp string `json:"timestamp"`
	RefID     int64  `json:"ref_id"` // RFQ ID or Lead ID or Request ID
}

type AIStatus struct {
	ActiveAgents  int `json:"active_agents"`
	TasksFinished int `json:"tasks_finished"`
	HealthScore   int `json:"health_score"` // out of 100
}

type ActiveShipment struct {
	ID          int64  `json:"id"`
	ShipmentNo  string `json:"shipment_no"`
	Customer    string `json:"customer"`
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Status      string `json:"status"`
	Mode        string `json:"mode"`
	ETA         string `json:"eta"`
}

type ActiveShipmentItem struct {
	ID            int64  `json:"id"`
	ShipmentNo    string `json:"shipment_no"`
	Customer      string `json:"customer"`
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	Carrier       string `json:"carrier"`
	Status        string `json:"status"`
	StatusDisplay string `json:"status_display"` // e.g. "+ In Transit", "+ On Board", "Customs Hold", "Awaiting Pickup"
	Mode          string `json:"mode"`
	ETA           string `json:"eta"` // e.g. "08 Jun 2025"
	ETADate       string `json:"eta_date"`
}

type RecentDocument struct {
	ID                int64  `json:"id"`
	DocumentName      string `json:"document_name"`
	DocumentType      string `json:"document_type"`
	Reference         string `json:"reference"`
	FileSize          int64  `json:"file_size"`
	FileSizeFormatted string `json:"file_size_formatted"`
	UploadedAt        string `json:"uploaded_at"`
}

type RecentActivity struct {
	ID        string `json:"id"`
	Type      string `json:"type"` // "PAYMENT", "DOCUMENT", "INVOICE", "APPROVAL", "RFQ", "CUSTOMER", "LEAD"
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle"`
	Timestamp string `json:"timestamp"`
	ActionURL string `json:"action_url"`
}

type UpcomingReminder struct {
	ID        string `json:"id"`
	Type      string `json:"type"` // "FOLLOW_UP", "CONTRACT_EXPIRY", "PAYMENT_DUE", "SHIPMENT_ETA"
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle"`
	DueText   string `json:"due_text"` // e.g. "Today, 03:00 PM", "In 9 days (11 Jun 2025)"
	DueDate   string `json:"due_date"`
	ActionURL string `json:"action_url"`
}

type ModuleCount struct {
	TotalCount  int  `json:"total_count"`
	ActiveCount int  `json:"active_count"`
	HasData     bool `json:"has_data"`
}

type ModuleStatus struct {
	Customers  ModuleCount `json:"customers"`
	Leads      ModuleCount `json:"leads"`
	RFQs       ModuleCount `json:"rfqs"`
	Quotations ModuleCount `json:"quotations"`
	Bookings   ModuleCount `json:"bookings"`
	Shipments  ModuleCount `json:"shipments"`
	Documents  ModuleCount `json:"documents"`
	Approvals  ModuleCount `json:"approvals"`
	Invoices   ModuleCount `json:"invoices"`
	Outreach   ModuleCount `json:"outreach"`
	Contracts  ModuleCount `json:"contracts"`
}

type OrganizationInfo struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	DefaultCurrency string `json:"default_currency"`
	DefaultTimezone string `json:"default_timezone"`
}

type DateRangeInfo struct {
	StartDate        string `json:"start_date"`
	EndDate          string `json:"end_date"`
	Preset           string `json:"preset"`
	Label            string `json:"label"`
	ComparisonLabel  string `json:"comparison_label"`
	DaysCount        int    `json:"days_count"`
	HasDataInPeriod  bool   `json:"has_data_in_period"`
}
