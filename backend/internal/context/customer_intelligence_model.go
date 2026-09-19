package bcontext

import "time"

// CustomerIdentitySummary holds identity, ownership, and authorized contact info.
type CustomerIdentitySummary struct {
	ID             int64                 `json:"id"`
	OrgID          int64                 `json:"org_id"`
	Name           string                `json:"name"`
	CustomerCode   string                `json:"customer_code"`
	TradingName    string                `json:"trading_name,omitempty"`
	CustomerType   string                `json:"customer_type"`
	Industry       string                `json:"industry,omitempty"`
	Status         string                `json:"status"`
	Country        string                `json:"country,omitempty"`
	City           string                `json:"city,omitempty"`
	Website        string                `json:"website,omitempty"`
	Currency       string                `json:"currency"`
	PaymentTerms   string                `json:"payment_terms"`
	CreditLimit    float64               `json:"credit_limit"`
	CreditStatus   string                `json:"credit_status"`
	HealthScore    int                   `json:"health_score"`
	AccountOwnerID *int64                `json:"account_owner_id,omitempty"`
	AccountOwner   string                `json:"account_owner,omitempty"`
	Contacts       []CustomerContactItem `json:"contacts"`
	CreatedAt      *time.Time            `json:"created_at,omitempty"`
	UpdatedAt      *time.Time            `json:"updated_at,omitempty"`
}

// CustomerContactItem represents a contact person associated with the customer.
type CustomerContactItem struct {
	ID          int64  `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	JobTitle    string `json:"job_title,omitempty"`
	Department  string `json:"department,omitempty"`
	ContactRole string `json:"contact_role"`
	IsPrimary   bool   `json:"is_primary"`
}

// CommercialIntelligenceMetrics provides deterministic commercial metrics.
type CommercialIntelligenceMetrics struct {
	TotalRFQs              int      `json:"total_rfqs"`
	OpenRFQs               int      `json:"open_rfqs"`
	TotalQuotations        int      `json:"total_quotations"`
	OpenQuotations         int      `json:"open_quotations"`
	AcceptedQuotations     int      `json:"accepted_quotations"`
	TotalBookings          int      `json:"total_bookings"`
	ActiveBookings         int      `json:"active_bookings"`
	TotalContracts         int      `json:"total_contracts"`
	ActiveContracts        int      `json:"active_contracts"`
	TotalQuotedAmount      float64  `json:"total_quoted_amount"`
	AcceptedQuotedAmount   float64  `json:"accepted_quoted_amount"`
	Currency               string   `json:"currency"`
	QuoteToBookingConvRate *float64 `json:"quote_to_booking_conversion_rate,omitempty"` // nil if insufficient quotes
	AvgQuotationValue      *float64 `json:"avg_quotation_value,omitempty"`              // nil if 0 quotes
	ConversionExplanation  string   `json:"conversion_explanation"`
}

// OperationsIntelligenceMetrics provides deterministic operations metrics.
type OperationsIntelligenceMetrics struct {
	TotalShipments          int        `json:"total_shipments"`
	ActiveShipments         int        `json:"active_shipments"`
	DeliveredShipments      int        `json:"delivered_shipments"`
	DelayedShipments        int        `json:"delayed_shipments"`
	OpenExceptionsCount     int        `json:"open_exceptions_count"`
	TotalExceptionsCount    int        `json:"total_exceptions_count"`
	LastShipmentDate        *time.Time `json:"last_shipment_date,omitempty"`
	AvgTransitDays          *float64   `json:"avg_transit_days,omitempty"` // nil if insufficient completed shipments
	DeliveryPerformanceNote string     `json:"delivery_performance_note"`
}

// FinancialIntelligenceMetrics provides deterministic invoicing and balance metrics.
type FinancialIntelligenceMetrics struct {
	TotalInvoicedAmount   float64 `json:"total_invoiced_amount"`
	TotalPaidAmount       float64 `json:"total_paid_amount"`
	OutstandingBalance    float64 `json:"outstanding_balance"`
	OverdueInvoicesCount  int     `json:"overdue_invoices_count"`
	TotalInvoicesCount    int     `json:"total_invoices_count"`
	PaidInvoicesCount     int     `json:"paid_invoices_count"`
	Currency              string  `json:"currency"`
	PaymentComplianceRate float64 `json:"payment_compliance_rate"` // % of non-overdue invoices
}

// GovernanceAndActivityMetrics provides interaction and compliance tracking.
type GovernanceAndActivityMetrics struct {
	RecentInteractionsCount int        `json:"recent_interactions_count"`
	LastInteractionDate     *time.Time `json:"last_interaction_date,omitempty"`
	OpenApprovalsCount      int        `json:"open_approvals_count"`
	RecentAITasksCount      int        `json:"recent_ai_tasks_count"`
	RecentAuditLogsCount    int        `json:"recent_audit_logs_count"`
	EngagementTrend         string     `json:"engagement_trend"` // "HIGH", "MODERATE", "LOW", "INACTIVE"
	EngagementExplanation   string     `json:"engagement_explanation"`
}

// CustomerObservation is an individual grounded finding with source traceability.
type CustomerObservation struct {
	Module       string     `json:"module"` // "COMMERCIAL", "OPERATIONS", "FINANCIAL", "GOVERNANCE"
	RecordType   string     `json:"record_type"`
	RecordID     int64      `json:"record_id"`
	Reference    string     `json:"reference"`
	Finding      string     `json:"finding"`
	Significance string     `json:"significance"` // "POSITIVE", "ATTENTION", "CRITICAL", "NEUTRAL"
	Timestamp    *time.Time `json:"timestamp,omitempty"`
}

// CustomerAISummary holds the read-only AI summary sections.
type CustomerAISummary struct {
	ExecutiveSummary    string                `json:"executive_summary"`
	CommercialPosition  string                `json:"commercial_position"`
	OperationalPosition string                `json:"operational_position"`
	FinancialConcerns   string                `json:"financial_concerns"`
	AttentionItems      []string              `json:"attention_items"`
	RecommendedFollowUp []string              `json:"recommended_follow_up"`
	KeyObservations     []CustomerObservation `json:"key_observations"`
	ConfidenceLevel     string                `json:"confidence_level"` // "HIGH", "MEDIUM", "LOW"
	DataFreshness       string                `json:"data_freshness"`
	IsInformationalOnly bool                  `json:"is_informational_only"`
}

// Customer360Intelligence represents the complete assembled intelligence payload.
type Customer360Intelligence struct {
	OrgID                     int64                         `json:"org_id"`
	CustomerID                int64                         `json:"customer_id"`
	Customer                  CustomerIdentitySummary       `json:"customer"`
	CommercialMetrics         CommercialIntelligenceMetrics `json:"commercial_metrics"`
	OperationsMetrics         OperationsIntelligenceMetrics `json:"operations_metrics"`
	FinancialMetrics          FinancialIntelligenceMetrics  `json:"financial_metrics"`
	GovernanceMetrics         GovernanceAndActivityMetrics  `json:"governance_metrics"`
	AISummary                 CustomerAISummary             `json:"ai_summary"`
	SupportingRecords         []SourceReference             `json:"supporting_records"`
	SupportingFieldReferences []FieldReference             `json:"supporting_field_references"`
	CorrelationID             string                        `json:"correlation_id"`
	DataFreshness             time.Time                     `json:"data_freshness"`
	IsReadOnly                bool                          `json:"is_read_only"`
}
