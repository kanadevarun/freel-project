package sportal

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// PlatformOverview summarizes system-wide health and tenant statistics for internal SPortal administrators.
type PlatformOverview struct {
	// 1. Core Portfolio KPIs (preserved for backwards-compatibility)
	TotalOrganizations  int       `json:"total_organizations"`
	ActiveCustomers     int       `json:"active_customers"`
	PendingOnboarding   int       `json:"pending_onboarding"`
	TotalUsers          int       `json:"total_users"`
	ActiveSubscriptions int       `json:"active_subscriptions"`
	PlatformStatus      string    `json:"platform_status"`
	DatabaseStatus      string    `json:"database_status"`
	GeneratedAt         time.Time `json:"generated_at"`

	// 2. Commercial & Revenue Overview (RBAC-protected)
	MonthlyRecurringRev       float64                `json:"monthly_recurring_rev"`
	AnnualRunRate             float64                `json:"annual_run_rate"`
	TrialingSubscriptions     int                    `json:"trialing_subscriptions"`
	PastDueSubscriptions      int                    `json:"past_due_subscriptions"`
	ExpiringIn30Days          int                    `json:"expiring_in_30_days"`
	AutoRenewPercentage       float64                `json:"auto_renew_percentage"`
	PlanDistribution          []PlanDistributionItem `json:"plan_distribution"`
	UpcomingRenewals          []UpcomingRenewalItem  `json:"upcoming_renewals"`
	OutstandingInvoicesCount  int                    `json:"outstanding_invoices_count"`
	OutstandingInvoicesAmount float64                `json:"outstanding_invoices_amount"`
	PaidInvoicesAmount        float64                `json:"paid_invoices_amount"`
	Currency                  string                 `json:"currency"`

	// 3. Customer Health Aggregate Overview (S11)
	CustomerHealth PortfolioHealthSummary `json:"customer_health"`

	// 4. Operational Command & Activity Overview
	Operations OperationsSummary `json:"operations"`

	// 5. Adoption & Usage Overview (S10)
	Adoption AdoptionSummary `json:"adoption"`

	// 6. Integrations & External Gateways (S12)
	Integrations IntegrationsSummary `json:"integrations"`

	// 7. Documents & Compliance (S13)
	Documents DocumentsComplianceSummary `json:"documents"`

	// 8. AI Workforce & Automations
	AiWorkforce AiWorkforceSummary `json:"ai_workforce"`

	// 9. Priority Attention Items (genuine alerts with drill-downs)
	AttentionItems []DashboardAttentionItem `json:"attention_items"`

	// 10. Historical Trends for Charts (Customer Growth & Revenue over last 6 months)
	GrowthTrend  []MonthlyGrowthItem  `json:"growth_trend"`
	RevenueTrend []MonthlyRevenueItem `json:"revenue_trend"`

	// 11. Platform Services Health Matrix
	PlatformHealth PlatformHealthMatrix `json:"platform_health"`
}

// PlanDistributionItem represents the breakdown of active customer subscriptions across plan tiers.
type PlanDistributionItem struct {
	PlanName string  `json:"plan_name"`
	Count    int     `json:"count"`
	Revenue  float64 `json:"revenue"`
}

// UpcomingRenewalItem details a customer subscription approaching renewal.
type UpcomingRenewalItem struct {
	OrgID            int64     `json:"org_id"`
	OrgName          string    `json:"org_name"`
	PlanName         string    `json:"plan_name"`
	Amount           float64   `json:"amount"`
	Currency         string    `json:"currency"`
	CurrentPeriodEnd time.Time `json:"current_period_end"`
	DaysLeft         int       `json:"days_left"`
	AutoRenew        bool      `json:"auto_renew"`
}

// PortfolioHealthSummary aggregates health states across the customer portfolio.
type PortfolioHealthSummary struct {
	HealthyCount          int     `json:"healthy_count"`
	WatchCount            int     `json:"watch_count"`
	AtRiskCount           int     `json:"at_risk_count"`
	CriticalCount         int     `json:"critical_count"`
	InsufficientDataCount int     `json:"insufficient_data_count"`
	AverageHealthScore    float64 `json:"average_health_score"`
}

// OperationsSummary provides system-wide operational snapshot metrics.
type OperationsSummary struct {
	TotalShipments     int `json:"total_shipments"`
	ActiveShipments    int `json:"active_shipments"`
	OpenExceptions     int `json:"open_exceptions"`
	CriticalExceptions int `json:"critical_exceptions"`
	TotalRFQs          int `json:"total_rfqs"`
	TotalQuotations    int `json:"total_quotations"`
	TotalBookings      int `json:"total_bookings"`
}

// AdoptionSummary details feature adoption across customer organizations.
type AdoptionSummary struct {
	ActiveUsersCount       int `json:"active_users_count"`
	ActiveForwardersCount  int `json:"active_forwarders_count"`
	CoreModulesActiveCount int `json:"core_modules_active_count"`
	AiAdoptionCount        int `json:"ai_adoption_count"`
	IntegrationsCount      int `json:"integrations_count"`
}

// IntegrationsSummary details external gateway and provider connectivity.
type IntegrationsSummary struct {
	TotalConfigured  int    `json:"total_configured"`
	ActiveConnected  int    `json:"active_connected"`
	DegradedCount    int    `json:"degraded_count"`
	ErrorCount       int    `json:"error_count"`
	DeadLettersCount int    `json:"dead_letters_count"`
	GatewayStatus    string `json:"gateway_status"`
}

// DocumentsComplianceSummary provides platform document and contract compliance status.
type DocumentsComplianceSummary struct {
	TotalDocuments     int     `json:"total_documents"`
	VerifiedDocuments  int     `json:"verified_documents"`
	DiscrepancyCount   int     `json:"discrepancy_count"`
	ExpiringSoonCount  int     `json:"expiring_soon_count"`
	TotalContracts     int     `json:"total_contracts"`
	ActiveContracts    int     `json:"active_contracts"`
	AvgComplianceScore float64 `json:"avg_compliance_score"`
}

// AiWorkforceSummary details autonomous agents, tasks, and recommendations.
type AiWorkforceSummary struct {
	TotalAgents          int `json:"total_agents"`
	ActiveAgents         int `json:"active_agents"`
	TotalAutomations     int `json:"total_automations"`
	ActiveAutomations    int `json:"active_automations"`
	CompletedTasks       int `json:"completed_tasks"`
	PendingTasks         int `json:"pending_tasks"`
	TotalRecommendations int `json:"total_recommendations"`
}

// DashboardAttentionItem represents a verified operational, commercial, or compliance issue requiring action.
type DashboardAttentionItem struct {
	ID          string    `json:"id"`
	OrgID       int64     `json:"org_id"`
	OrgName     string    `json:"org_name"`
	Title       string    `json:"title"`
	Issue       string    `json:"issue"`
	Severity    string    `json:"severity"` // "CRITICAL", "HIGH", "MEDIUM", "INFO"
	Source      string    `json:"source"`   // "Subscription", "Billing", "Customer Health", "Operations", "Integration", "Documents", "Onboarding"
	TargetRoute string    `json:"target_route"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MonthlyGrowthItem represents customer additions and cumulative portfolio size for a month.
type MonthlyGrowthItem struct {
	Month          string `json:"month"`
	NewCustomers   int    `json:"new_customers"`
	TotalCustomers int    `json:"total_customers"`
}

// MonthlyRevenueItem represents monthly revenue trajectory for executive charts.
type MonthlyRevenueItem struct {
	Month        string  `json:"month"`
	MRR          float64 `json:"mrr"`
	ARRProjected float64 `json:"arr_projected"`
}

// ServiceHealthStatus represents individual platform microservice health.
type ServiceHealthStatus struct {
	Name      string `json:"name"`
	Status    string `json:"status"` // "OPERATIONAL", "DEGRADED", "ERROR", "UNKNOWN"
	Endpoint  string `json:"endpoint"`
	Message   string `json:"message"`
	LatencyMs int64  `json:"latency_ms"`
}

// PlatformHealthMatrix aggregates multi-subsystem platform health checks.
type PlatformHealthMatrix struct {
	OverallStatus string                `json:"overall_status"`
	Services      []ServiceHealthStatus `json:"services"`
}


// OrganizationSummary provides high-level tenant details for the SPortal recent organizations view.
type OrganizationSummary struct {
	ID           int64      `json:"id" db:"id"`
	Name         string     `json:"name" db:"name"`
	LegalName    *string    `json:"legal_name" db:"legal_name"`
	Status       string     `json:"status" db:"status"`
	PlanName     string     `json:"plan_name" db:"plan_name"`
	UserCount    int        `json:"user_count" db:"user_count"`
	RenewalDate  *time.Time `json:"renewal_date" db:"renewal_date"`
	HealthStatus string     `json:"health_status" db:"health_status"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// SPortalMetaResponse provides discovery and module capability metadata to the SPortal frontend.
type SPortalMetaResponse struct {
	PortalName       string    `json:"portal_name"`
	PortalVersion    string    `json:"portal_version"`
	Environment      string    `json:"environment"`
	AvailableModules []string  `json:"available_modules"`
	ApiBaseURL       string    `json:"api_base_url"`
	ServerTime       time.Time `json:"server_time"`
}

// SPortalLoginRequest contains credentials submitted on the SPortal login page.
type SPortalLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SPortalUserInfo represents the authenticated internal staff user.
type SPortalUserInfo struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Status    string `json:"status"`
}

// SPortalOrgInfo represents the internal organization context.
type SPortalOrgInfo struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	IsInternal bool   `json:"is_internal"`
}

// SPortalRoleInfo represents the internal role and its authorized SPortal permissions.
type SPortalRoleInfo struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Permissions []string `json:"permissions"`
}

// SPortalLoginResponseData is returned on successful SPortal authentication.
type SPortalLoginResponseData struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token,omitempty"`
	ExpiresIn    int32           `json:"expires_in"`
	User         SPortalUserInfo `json:"user"`
	Org          SPortalOrgInfo  `json:"org"`
	Role         SPortalRoleInfo `json:"role"`
	IsInternal   bool            `json:"is_internal"`
}

// SPortalCurrentUserResponseData is returned by GET /api/v1/sportal/auth/me.
type SPortalCurrentUserResponseData struct {
	User       SPortalUserInfo `json:"user"`
	Org        SPortalOrgInfo  `json:"org"`
	Role       SPortalRoleInfo `json:"role"`
	IsInternal bool            `json:"is_internal"`
	SessionAt  time.Time       `json:"session_at"`
}

// SensitiveFinancialInfoResponse represents sensitive platform billing data guarded by billing:sensitive_view.
type SensitiveFinancialInfoResponse struct {
	GatewayAccountID   string    `json:"gateway_account_id"`
	BankSettlementNode string    `json:"bank_settlement_node"`
	TaxRegistrationPAN string    `json:"tax_registration_pan"`
	GSTSecretStatus    string    `json:"gst_secret_status"`
	AuditNotice        string    `json:"audit_notice"`
	AuthorizedViewer   string    `json:"authorized_viewer"`
	Timestamp          time.Time `json:"timestamp"`
}

// --- Task S3: Organizations & Customer 360 Types ---

// OrganizationListParams holds search, filtering, and pagination parameters.
type OrganizationListParams struct {
	Search    string
	Status    string
	Plan      string
	SortBy    string
	SortOrder string
	Page      int
	Limit     int
}

// OrganizationListItem represents an organization entry in the SPortal management table.
type OrganizationListItem struct {
	ID               int64     `json:"id" db:"id"`
	Name             string    `json:"name" db:"name"`
	LegalName        *string   `json:"legal_name" db:"legal_name"`
	RegistrationNum  *string   `json:"registration_number" db:"registration_number"`
	TaxNumber        *string   `json:"tax_number" db:"tax_number"`
	Status           string    `json:"status" db:"status"`
	OnboardingStatus string    `json:"onboarding_status" db:"onboarding_status"`
	PrimaryEmail     *string   `json:"primary_email" db:"primary_email"`
	PhoneNumber      *string   `json:"phone_number" db:"phone_number"`
	City             *string   `json:"city" db:"city"`
	State            *string   `json:"state" db:"state"`
	Country          *string   `json:"country" db:"country"`
	PlanName         string    `json:"plan_name" db:"plan_name"`
	PlanStatus       string    `json:"plan_status" db:"plan_status"`
	UserCount        int       `json:"user_count" db:"user_count"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// OrganizationListResult represents the paginated result list returned to SPortal.
type OrganizationListResult struct {
	Items      []OrganizationListItem `json:"items"`
	Total      int                    `json:"total"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	TotalPages int                    `json:"total_pages"`
}

// OrganizationProfile represents complete organization company details for Customer 360.
type OrganizationProfile struct {
	ID                 int64      `json:"id" db:"id"`
	Name               string     `json:"name" db:"name"`
	LegalName          *string    `json:"legal_name" db:"legal_name"`
	RegistrationNumber *string    `json:"registration_number" db:"registration_number"`
	TaxNumber          *string    `json:"tax_number" db:"tax_number"`
	Website            *string    `json:"website" db:"website"`
	PrimaryEmail       *string    `json:"primary_email" db:"primary_email"`
	PhoneNumber        *string    `json:"phone_number" db:"phone_number"`
	SupportEmail       *string    `json:"support_email" db:"support_email"`
	Address            *string    `json:"address" db:"address"`
	City               *string    `json:"city" db:"city"`
	State              *string    `json:"state" db:"state"`
	Country            *string    `json:"country" db:"country"`
	PostalCode         *string    `json:"postal_code" db:"postal_code"`
	Industry           *string    `json:"industry" db:"industry"`
	CompanyType        *string    `json:"company_type" db:"company_type"`
	DefaultCurrency    *string    `json:"default_currency" db:"default_currency"`
	DefaultTimezone    *string    `json:"default_timezone" db:"default_timezone"`
	DateFormat         *string    `json:"date_format" db:"date_format"`
	LogoURL            *string    `json:"logo_url" db:"logo_url"`
	Status             string     `json:"status" db:"status"`
	OnboardingStatus   string     `json:"onboarding_status" db:"onboarding_status"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// OrganizationSubscriptionSummary summarizes active plan and billing status.
type OrganizationSubscriptionSummary struct {
	PlanName         string     `json:"plan_name" db:"plan_name"`
	PlanCode         string     `json:"plan_code" db:"plan_code"`
	Status           string     `json:"status" db:"status"`
	BillingCycle     string     `json:"billing_cycle" db:"billing_cycle"`
	CurrentPeriodEnd *time.Time `json:"current_period_end" db:"current_period_end"`
	MonthlyPrice     float64    `json:"monthly_price" db:"monthly_price"`
}

// OrganizationMemberSummary represents a user belonging to the organization.
type OrganizationMemberSummary struct {
	UserID    int64     `json:"user_id" db:"user_id"`
	Email     string    `json:"email" db:"email"`
	FullName  string    `json:"full_name" db:"full_name"`
	RoleName  string    `json:"role_name" db:"role_name"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// OrganizationBusinessStats aggregates operational telemetry for Customer 360.
type OrganizationBusinessStats struct {
	ShipmentsCount           int     `json:"shipments_count"`
	ActiveShipmentsCount     int     `json:"active_shipments_count"`
	CompletedShipments30d    int     `json:"completed_shipments_30d"`
	RFQsCount                int     `json:"rfqs_count"`
	RFQs30d                  int     `json:"rfqs_30d"`
	QuotesCount              int     `json:"quotes_count"`
	Quotes30d                int     `json:"quotes_30d"`
	BookingsCount            int     `json:"bookings_count"`
	Bookings30d              int     `json:"bookings_30d"`
	InvoicesCount            int     `json:"invoices_count"`
	OutstandingInvoicesCount int     `json:"outstanding_invoices_count"`
	OutstandingInvoicesAmount float64 `json:"outstanding_invoices_amount"`
	CustomersCount           int     `json:"customers_count"`
	ExceptionsCount          int     `json:"exceptions_count"`
	OpenExceptionsCount      int     `json:"open_exceptions_count"`
	ExpiringContractsCount   int     `json:"expiring_contracts_count"`
	PendingInvitationsCount  int     `json:"pending_invitations_count"`
	IntegrationIssuesCount   int     `json:"integration_issues_count"`
	AIActionItemsCount       int     `json:"ai_action_items_count"`
}

// OrganizationActivityItem represents an audit entry for this customer organization.
type OrganizationActivityItem struct {
	ID          int64     `json:"id" db:"id"`
	Action      string    `json:"action" db:"action"`
	Module      string    `json:"module" db:"module"`
	Description string    `json:"description" db:"description"`
	Result      string    `json:"result" db:"result"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// CustomerHealthSummary represents honest customer health assessment based on real operational signals.
type CustomerHealthSummary struct {
	Status         string   `json:"status"` // "Good", "Needs Attention", "At Risk"
	Score          int      `json:"score"`  // 0-100
	Summary        string   `json:"summary"`
	RiskIndicators []string `json:"risk_indicators"`
}

// ShipmentTrendItem represents real monthly shipment creation and completion metrics.
type ShipmentTrendItem struct {
	Month          string `json:"month" db:"month_name"`
	CreatedCount   int    `json:"created_count" db:"created_count"`
	CompletedCount int    `json:"completed_count" db:"completed_count"`
}

// CustomerAlertItem represents an operational item requiring attention in Customer 360.
type CustomerAlertItem struct {
	Type       string `json:"type"`
	Title      string `json:"title"`
	Count      int    `json:"count"`
	Severity   string `json:"severity"` // "high", "medium", "low"
	LinkModule string `json:"link_module"`
}

// CustomerTeamSummary represents workforce breakdown for the customer.
type CustomerTeamSummary struct {
	TotalUsers     int                         `json:"total_users"`
	ActiveUsers    int                         `json:"active_users"`
	PendingUsers   int                         `json:"pending_users"`
	SuspendedUsers int                         `json:"suspended_users"`
	PrimaryAdmin   *OrganizationMemberSummary  `json:"primary_admin"`
	OtherKeyUsers  []OrganizationMemberSummary `json:"other_key_users"`
}

// Customer360Details is the aggregate payload returned by GET /api/v1/sportal/organizations/:id.
type Customer360Details struct {
	Organization   OrganizationProfile              `json:"organization"`
	Subscription   *OrganizationSubscriptionSummary `json:"subscription"`
	Users          []OrganizationMemberSummary      `json:"users"`
	Stats          OrganizationBusinessStats        `json:"stats"`
	RecentActivity []OrganizationActivityItem       `json:"recent_activity"`
	Health         CustomerHealthSummary            `json:"health"`
	ShipmentTrends []ShipmentTrendItem              `json:"shipment_trends"`
	OpenAlerts     []CustomerAlertItem              `json:"open_alerts"`
	TeamSummary    CustomerTeamSummary              `json:"team_summary"`
}

// CustomerShipmentItem represents a customer shipment in Customer 360.
type CustomerShipmentItem struct {
	ID             int64      `json:"id" db:"id"`
	BookingNumber  string     `json:"booking_number" db:"booking_number"`
	CarrierSCAC    string     `json:"carrier_scac" db:"carrier_scac"`
	VesselName     string     `json:"vessel_name" db:"vessel_name"`
	VoyageNumber   string     `json:"voyage_number" db:"voyage_number"`
	OriginPort     string     `json:"origin_port" db:"origin_port"`
	DestinationPort string    `json:"destination_port" db:"destination_port"`
	Status         string     `json:"status" db:"status"`
	ETD            *time.Time `json:"etd" db:"etd"`
	ETA            *time.Time `json:"eta" db:"eta"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// CustomerInvoiceItem represents an invoice in Customer 360.
type CustomerInvoiceItem struct {
	ID            int64      `json:"id" db:"id"`
	InvoiceNumber string     `json:"invoice_number" db:"invoice_number"`
	CustomerName  string     `json:"customer_name" db:"customer_name"`
	TotalAmount   float64    `json:"total_amount" db:"total_amount"`
	PaidAmount    float64    `json:"paid_amount" db:"paid_amount"`
	BalanceDue    float64    `json:"balance_due" db:"balance_due"`
	Currency      string     `json:"currency" db:"currency"`
	Status        string     `json:"status" db:"status"`
	DueDate       *time.Time `json:"due_date" db:"due_date"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// CustomerContractItem represents a contract in Customer 360.
type CustomerContractItem struct {
	ID                int64      `json:"id" db:"id"`
	ContractReference string     `json:"contract_reference" db:"contract_reference"`
	ContractName      string     `json:"contract_name" db:"contract_name"`
	ContractType      string     `json:"contract_type" db:"contract_type"`
	PartyID           *int64     `json:"party_id,omitempty" db:"party_id"`
	PartyName         string     `json:"party_name" db:"party_name"`
	TransportMode     string     `json:"transport_mode,omitempty" db:"transport_mode"`
	Status            string     `json:"status" db:"status"`
	ContractValue     float64    `json:"contract_value" db:"contract_value"`
	Currency          string     `json:"currency" db:"currency"`
	EffectiveDate     *time.Time `json:"effective_date" db:"effective_date"`
	ExpiryDate        *time.Time `json:"expiry_date" db:"expiry_date"`
	DaysToExpiry      *int       `json:"days_to_expiry,omitempty"`
	IsExpiringSoon    bool       `json:"is_expiring_soon"`
	IsExpired         bool       `json:"is_expired"`
	Owner             string     `json:"owner,omitempty" db:"owner"`
	Description       string     `json:"description,omitempty" db:"description"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// CustomerExceptionItem represents an operational exception in Customer 360.
type CustomerExceptionItem struct {
	ID            int64     `json:"id" db:"id"`
	ShipmentID    int64     `json:"shipment_id" db:"shipment_id"`
	ExceptionType string    `json:"exception_type" db:"exception_type"`
	Severity      string    `json:"severity" db:"severity"`
	Title         string    `json:"title" db:"title"`
	Description   string    `json:"description" db:"description"`
	Status        string    `json:"status" db:"status"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// CustomerIntegrationItem represents an external or carrier integration for the customer.
type CustomerIntegrationItem struct {
	ID               int64      `json:"id" db:"id"`
	CarrierName      string     `json:"carrier_name" db:"carrier_name"`
	CarrierSCAC      string     `json:"carrier_scac" db:"carrier_scac"`
	ConnectionMethod string     `json:"connection_method" db:"connection_method"`
	ConnectionStatus string     `json:"connection_status" db:"connection_status"`
	IsActive         bool       `json:"is_active" db:"is_active"`
	SyncStatus       string     `json:"sync_status" db:"sync_status"`
	LastSyncedAt     *time.Time `json:"last_synced_at" db:"last_synced_at"`
	LastError        string     `json:"last_error" db:"last_error"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// CustomerDocumentItem represents a shipment or contract document in Customer 360 and Documents & Compliance.
type CustomerDocumentItem struct {
	ID                 int64      `json:"id" db:"id"`
	OrgID              int64      `json:"org_id" db:"org_id"`
	ShipmentID         *int64     `json:"shipment_id,omitempty" db:"shipment_id"`
	ShipmentRef        string     `json:"shipment_ref,omitempty" db:"shipment_ref"`
	CustomerID         *int64     `json:"customer_id,omitempty" db:"customer_id"`
	CustomerName       string     `json:"customer_name,omitempty" db:"customer_name"`
	BookingID          *int64     `json:"booking_id,omitempty" db:"booking_id"`
	BookingRef         string     `json:"booking_ref,omitempty" db:"booking_ref"`
	DocType            string     `json:"doc_type" db:"doc_type"`
	DocumentName       string     `json:"document_name" db:"document_name"`
	FileName           string     `json:"file_name" db:"file_name"`
	OriginalFileName   string     `json:"original_file_name,omitempty" db:"original_file_name"`
	FileSize           int64      `json:"file_size" db:"file_size"`
	MIMEType           string     `json:"mime_type,omitempty" db:"mime_type"`
	Status             string     `json:"status" db:"status"`
	DiscrepanciesCount int        `json:"discrepancies_count" db:"discrepancies_count"`
	HasDiscrepancy     bool       `json:"has_discrepancy" db:"has_discrepancy"`
	DaysToExpiry       int        `json:"days_to_expiry,omitempty" db:"days_to_expiry"`
	IsExpiringSoon     bool       `json:"is_expiring_soon" db:"is_expiring_soon"`
	IsExpired          bool       `json:"is_expired" db:"is_expired"`
	DownloadURL        string     `json:"download_url,omitempty" db:"download_url"`
	UploadedBy         string     `json:"uploaded_by,omitempty" db:"uploaded_by"`
	UploadedAt         *time.Time `json:"uploaded_at,omitempty" db:"uploaded_at"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// CustomerAiSummary represents safe AI Workforce and automation intelligence for the customer.
type CustomerAiSummary struct {
	ActiveAutomationsCount   int                  `json:"active_automations_count"`
	PendingTasksCount        int                  `json:"pending_tasks_count"`
	CompletedTasksCount      int                  `json:"completed_tasks_count"`
	FailedTasksCount         int                  `json:"failed_tasks_count"`
	RecentTasks              []CustomerAiTaskItem `json:"recent_tasks"`
	SafetyGovernanceStatus   string               `json:"safety_governance_status"`
	PersonalizationSummary   string               `json:"personalization_summary"`
}

// CustomerAiTaskItem represents an individual AI task execution.
type CustomerAiTaskItem struct {
	ID        int64      `json:"id" db:"id"`
	TaskType  string     `json:"task_type" db:"task_type"`
	Status    string     `json:"status" db:"status"`
	StartedAt *time.Time `json:"started_at" db:"started_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// CreateOrganizationRequest is submitted to create a new freight-forwarding customer organization.
type CreateOrganizationRequest struct {
	Name               string  `json:"name"`
	LegalName          string  `json:"legal_name"`
	RegistrationNumber string  `json:"registration_number"`
	TaxNumber          string  `json:"tax_number"`
	Website            string  `json:"website"`
	PrimaryEmail       string  `json:"primary_email"`
	PhoneNumber        string  `json:"phone_number"`
	Address            string  `json:"address"`
	City               string  `json:"city"`
	State              string  `json:"state"`
	Country            string  `json:"country"`
	PostalCode         string  `json:"postal_code"`
	Industry           string  `json:"industry"`
	CompanyType        string  `json:"company_type"`
	InitialPlan        string  `json:"initial_plan"`
}

// --- Task S5: Subscription & Plan Management Types ---

// PlanItem represents a commercial SaaS plan in SPortal.
type PlanItem struct {
	ID                   int64           `json:"id" db:"id"`
	Name                 string          `json:"name" db:"name"`
	Description          string          `json:"description" db:"description"`
	PriceMonthly         float64         `json:"price_monthly" db:"price_monthly"`
	PriceAnnual          float64         `json:"price_annual" db:"price_annual"`
	Features             json.RawMessage `json:"features" db:"features"`
	Limits               json.RawMessage `json:"limits" db:"limits"`
	ProviderProductID    *string         `json:"provider_product_id" db:"provider_product_id"`
	IsActive             bool            `json:"is_active" db:"is_active"`
	ActiveCustomersCount int             `json:"active_customers_count" db:"active_customers_count"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at" db:"updated_at"`
}

// CreatePlanRequest holds parameters to define a new plan tier.
type CreatePlanRequest struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	PriceMonthly float64                `json:"price_monthly"`
	PriceAnnual  float64                `json:"price_annual"`
	Features     []string               `json:"features"`
	Limits       map[string]interface{} `json:"limits"`
	IsActive     bool                   `json:"is_active"`
}

// UpdatePlanRequest holds parameters to modify an existing plan tier.
type UpdatePlanRequest struct {
	Name         *string                `json:"name"`
	Description  *string                `json:"description"`
	PriceMonthly *float64               `json:"price_monthly"`
	PriceAnnual  *float64               `json:"price_annual"`
	Features     []string               `json:"features"`
	Limits       map[string]interface{} `json:"limits"`
	IsActive     *bool                  `json:"is_active"`
}

// CustomerSubscriptionListItem represents an organization's subscription in the SPortal table.
type CustomerSubscriptionListItem struct {
	OrgID              int64      `json:"org_id" db:"org_id"`
	OrgName            string     `json:"org_name" db:"org_name"`
	LegalName          *string    `json:"legal_name" db:"legal_name"`
	PrimaryEmail       *string    `json:"primary_email" db:"primary_email"`
	Country            *string    `json:"country" db:"country"`
	SubscriptionID     *int64     `json:"subscription_id" db:"subscription_id"`
	PlanID             *int64     `json:"plan_id" db:"plan_id"`
	PlanName           *string    `json:"plan_name" db:"plan_name"`
	Status             string     `json:"status" db:"status"` // ACTIVE, TRIALING, PAST_DUE, CANCELED, NOT_CONFIGURED
	BillingCycle       *string    `json:"billing_cycle" db:"billing_cycle"` // monthly, annual
	Amount             float64    `json:"amount" db:"amount"`
	Currency           string     `json:"currency" db:"currency"`
	CurrentPeriodStart *time.Time `json:"current_period_start" db:"current_period_start"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end" db:"current_period_end"` // Renewal Date
	AutoRenew          bool       `json:"auto_renew" db:"auto_renew"` // true if cancel_at_period_end == 0
	PaymentStatus      string     `json:"payment_status" db:"payment_status"` // CURRENT, PENDING, FAILED, NOT_CONFIGURED
	DaysUntilRenewal   *int       `json:"days_until_renewal"`
	CreatedAt          *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at" db:"updated_at"`
}

// CustomerSubscriptionListParams contains filter/search query params.
type CustomerSubscriptionListParams struct {
	Search    string
	Status    string
	PlanID    int64
	AutoRenew *bool
	SortBy    string
	SortOrder string
	Page      int
	Limit     int
}

// SubscriptionDashboardMetrics summarizes commercial KPI totals.
type SubscriptionDashboardMetrics struct {
	TotalOrganizations  int     `json:"total_organizations"`
	ActiveSubscriptions int     `json:"active_subscriptions"`
	TrialingCount       int     `json:"trialing_count"`
	PastDueCount        int     `json:"past_due_count"`
	NotConfiguredCount  int     `json:"not_configured_count"`
	MonthlyRecurringRev float64 `json:"monthly_recurring_rev"`
	AnnualRunRate       float64 `json:"annual_run_rate"`
	AutoRenewPercentage float64 `json:"auto_renew_percentage"`
	ExpiringIn30Days    int     `json:"expiring_in_30_days"`
}

// CustomerSubscriptionListResult is the paginated response for subscriptions table.
type CustomerSubscriptionListResult struct {
	Items      []CustomerSubscriptionListItem `json:"items"`
	Total      int                            `json:"total"`
	Page       int                            `json:"page"`
	Limit      int                            `json:"limit"`
	TotalPages int                            `json:"total_pages"`
	Metrics    SubscriptionDashboardMetrics   `json:"metrics"`
}

// SubscriptionUsageSummary displays current consumption against plan limits.
type SubscriptionUsageSummary struct {
	MetricName   string `json:"metric_name"`
	CurrentUsage int    `json:"current_usage"`
	LimitAmount  *int   `json:"limit_amount"`
	Unlimited    bool   `json:"unlimited"`
	Remaining    int    `json:"remaining"`
	Percentage   int    `json:"percentage"`
}

// SubscriptionHistoryItem tracks historical changes from the audit ledger.
type SubscriptionHistoryItem struct {
	ID          int64     `json:"id" db:"id"`
	Action      string    `json:"action" db:"action"`
	Description string    `json:"description" db:"description"`
	Actor       string    `json:"actor" db:"actor"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// SubscriptionDetailView represents the comprehensive subscription profile for an organization.
type SubscriptionDetailView struct {
	OrgID                  int64                      `json:"org_id"`
	OrgName                string                     `json:"org_name"`
	LegalName              *string                    `json:"legal_name"`
	PrimaryEmail           *string                    `json:"primary_email"`
	SubscriptionID         *int64                     `json:"subscription_id"`
	PlanID                 *int64                     `json:"plan_id"`
	PlanName               string                     `json:"plan_name"`
	PlanDescription        string                     `json:"plan_description"`
	Status                 string                     `json:"status"`
	BillingCycle           string                     `json:"billing_cycle"`
	Amount                 float64                    `json:"amount"`
	Currency               string                     `json:"currency"`
	CurrentPeriodStart     *time.Time                 `json:"current_period_start"`
	CurrentPeriodEnd       *time.Time                 `json:"current_period_end"`
	DaysUntilRenewal       *int                       `json:"days_until_renewal"`
	AutoRenew              bool                       `json:"auto_renew"`
	CancelAtPeriodEnd      bool                       `json:"cancel_at_period_end"`
	PaymentStatus          string                     `json:"payment_status"`
	ProviderSubscriptionID *string                    `json:"provider_subscription_id"`
	RenewalURL             *string                    `json:"renewal_url"` // Truthful link or nil if not configured
	CustomerPortalURL      *string                    `json:"customer_portal_url"`
	Features               []string                   `json:"features"`
	Limits                 map[string]interface{}     `json:"limits"`
	Usage                  []SubscriptionUsageSummary `json:"usage"`
	History                []SubscriptionHistoryItem  `json:"history"`
	CreatedAt              *time.Time                 `json:"created_at"`
	UpdatedAt              *time.Time                 `json:"updated_at"`
}

// AssignSubscriptionRequest assigns an initial subscription to an organization.
type AssignSubscriptionRequest struct {
	PlanID       int64  `json:"plan_id"`
	BillingCycle string `json:"billing_cycle"` // monthly, annual
	AutoRenew    bool   `json:"auto_renew"`
	StartDate    string `json:"start_date,omitempty"`
}

// ChangeCustomerPlanRequest changes an organization's plan tier or billing frequency.
type ChangeCustomerPlanRequest struct {
	PlanID       int64  `json:"plan_id"`
	BillingCycle string `json:"billing_cycle"` // monthly, annual
	Notes        string `json:"notes,omitempty"`
}

// ToggleAutoRenewRequest updates auto-renew configuration.
type ToggleAutoRenewRequest struct {
	AutoRenew bool   `json:"auto_renew"`
	Reason    string `json:"reason,omitempty"`
}

// RenewSubscriptionRequest manually extends a subscription period.
type RenewSubscriptionRequest struct {
	ExtendMonths int    `json:"extend_months"` // e.g. 1, 3, 12
	Notes        string `json:"notes,omitempty"`
}

// CancelCustomerSubscriptionRequest terminates a subscription.
type CancelCustomerSubscriptionRequest struct {
	Immediate bool   `json:"immediate"`
	Reason    string `json:"reason"`
}

// UpdateOrganizationRequest is submitted to update organization metadata.
type UpdateOrganizationRequest struct {
	Name               *string `json:"name"`
	LegalName          *string `json:"legal_name"`
	RegistrationNumber *string `json:"registration_number"`
	TaxNumber          *string `json:"tax_number"`
	Website            *string `json:"website"`
	PrimaryEmail       *string `json:"primary_email"`
	PhoneNumber        *string `json:"phone_number"`
	Address            *string `json:"address"`
	City               *string `json:"city"`
	State              *string `json:"state"`
	Country            *string `json:"country"`
	PostalCode         *string `json:"postal_code"`
	Industry           *string `json:"industry"`
	CompanyType        *string `json:"company_type"`
	LogoURL            *string `json:"logo_url"`
}

// Task S6: Customer Organization Users, Roles, Invitations & Access Lifecycle

// CustomerUserListItem represents a customer organization user or invited user in the SPortal directory.
type CustomerUserListItem struct {
	UserID           int64      `json:"user_id" db:"user_id"`
	OrgID            int64      `json:"org_id" db:"org_id"`
	OrgName          string     `json:"org_name" db:"org_name"`
	FirstName        string     `json:"first_name" db:"first_name"`
	LastName         string     `json:"last_name" db:"last_name"`
	FullName         string     `json:"full_name" db:"full_name"`
	Email            string     `json:"email" db:"email"`
	Phone            string     `json:"phone" db:"phone"`
	RoleID           int64      `json:"role_id" db:"role_id"`
	RoleName         string     `json:"role_name" db:"role_name"`
	RoleDescription  string     `json:"role_description" db:"role_description"`
	Status           string     `json:"status" db:"status"` // ACTIVE, INACTIVE, SUSPENDED, DISABLED, PENDING
	InvitationStatus string     `json:"invitation_status" db:"invitation_status"` // ACCEPTED, PENDING, SENT, EXPIRED, NOT_INVITED, FAILED
	InvitationID     *int64     `json:"invitation_id,omitempty" db:"invitation_id"`
	LastLogin        *time.Time `json:"last_login" db:"last_login"`
	MFAStatus        string     `json:"mfa_status" db:"mfa_status"` // NOT_ENABLED, ENABLED
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// CustomerUserMetrics aggregates high-level customer workforce statistics.
type CustomerUserMetrics struct {
	TotalUsers         int `json:"total_users" db:"total_users"`
	ActiveUsers        int `json:"active_users" db:"active_users"`
	InactiveUsers      int `json:"inactive_users" db:"inactive_users"`
	PendingInvitations int `json:"pending_invitations" db:"pending_invitations"`
	SuperAdminsCount   int `json:"super_admins_count" db:"super_admins_count"`
}

// CustomerUserListParams specifies filtering and pagination criteria for the customer user directory.
type CustomerUserListParams struct {
	Search           string `json:"search"`
	OrgID            int64  `json:"org_id"`
	RoleName         string `json:"role_name"`
	Status           string `json:"status"`
	InvitationStatus string `json:"invitation_status"`
	Page             int    `json:"page"`
	Limit            int    `json:"limit"`
	SortBy           string `json:"sort_by"`
	SortOrder        string `json:"sort_order"`
}

// CustomerUserListResult contains paginated customer user records and summary metrics.
type CustomerUserListResult struct {
	Items      []CustomerUserListItem `json:"items"`
	Total      int                    `json:"total"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	TotalPages int                    `json:"total_pages"`
	Metrics    CustomerUserMetrics    `json:"metrics"`
}

// OrgUserRoleBreakdown provides the count of users assigned to each specific role in an organization.
type OrgUserRoleBreakdown struct {
	RoleID      int64   `json:"role_id" db:"role_id"`
	RoleName    string  `json:"role_name" db:"role_name"`
	Description string  `json:"description" db:"description"`
	Count       int     `json:"count" db:"count"`
	Percentage  float64 `json:"percentage"`
}

// OrgUserSummary provides the complete Customer 360 User breakdown for an organization.
type OrgUserSummary struct {
	OrgID         int64                  `json:"org_id"`
	OrgName       string                 `json:"org_name"`
	TotalUsers    int                    `json:"total_users"`
	ActiveUsers   int                    `json:"active_users"`
	RoleBreakdown []OrgUserRoleBreakdown `json:"role_breakdown"`
}

// CustomerUserDetailView provides deep dossier, permission visibility and audit activity for a customer user.
type CustomerUserDetailView struct {
	CustomerUserListItem
	Permissions     []string                   `json:"permissions"`
	EffectiveAccess *EffectiveAccessSummary    `json:"effective_access"`
	AuditActivity   []OrganizationActivityItem `json:"audit_activity"`
}

// -----------------------------------------------------------------------------
// TASK S7: Customer Users, Roles, Permission Visibility & Access Administration
// -----------------------------------------------------------------------------

// EffectiveModuleAccess explains the rights a role/user has for a specific platform module.
type EffectiveModuleAccess struct {
	Module         string   `json:"module"`          // e.g. "Shipments & Operations"
	Resource       string   `json:"resource"`        // e.g. "SHIPMENTS"
	AllowedActions []string `json:"allowed_actions"` // ["CREATE", "READ", "UPDATE"]
	DeniedActions  []string `json:"denied_actions"`  // ["DELETE"]
	AccessLevel    string   `json:"access_level"`    // "FULL", "MANAGE", "VIEW_ONLY", "NONE"
	Explanation    string   `json:"explanation"`     // Clear human-readable summary
}

// EffectiveAccessSummary provides a complete explanation of a user's rights.
type EffectiveAccessSummary struct {
	RoleName               string                  `json:"role_name"`
	RoleDescription        string                  `json:"role_description"`
	IsSuperAdmin           bool                    `json:"is_super_admin"`
	IsInternalStaff        bool                    `json:"is_internal_staff"`
	TenantScope            string                  `json:"tenant_scope"`
	CanAdministerUsers     bool                    `json:"can_administer_users"`
	CanAccessPlatformAdmin bool                    `json:"can_access_platform_admin"`
	SPortalBoundaryNotice  string                  `json:"sportal_boundary_notice"`
	CPortalResponsibility  string                  `json:"cportal_responsibility"`
	Modules                []EffectiveModuleAccess `json:"modules"`
	KeyCapabilities        []string                `json:"key_capabilities"`
	ExplicitDenials        []string                `json:"explicit_denials"`
}

// PlatformResourceCatalog defines a platform area with its available actions.
type PlatformResourceCatalog struct {
	Resource    string           `json:"resource"`     // e.g. "SHIPMENTS"
	DisplayName string           `json:"display_name"` // e.g. "Shipments & Tracking"
	Description string           `json:"description"`
	Category    string           `json:"category"` // "Operations", "Commercial", "Finance", "Governance"
	Actions     []ResourceAction `json:"actions"`
}

// ResourceAction defines an action on a resource with its canonical code.
type ResourceAction struct {
	Action      string `json:"action"`       // "CREATE", "READ", "UPDATE", "DELETE"
	DisplayName string `json:"display_name"` // "Create Shipments"
	Description string `json:"description"`
	Code        string `json:"code"` // "SHIPMENTS.CREATE"
}

// CustomerRoleDetail represents a role with full metadata and permission matrix for SPortal.
type CustomerRoleDetail struct {
	ID              int64                   `json:"id" db:"id"`
	OrgID           int64                   `json:"org_id" db:"org_id"`
	OrgName         string                  `json:"org_name" db:"org_name"`
	Name            string                  `json:"name" db:"name"`
	Description     string                  `json:"description" db:"description"`
	IsSystem        bool                    `json:"is_system"`       // true for 7 default roles
	IsProtected     bool                    `json:"is_protected"`    // true for SUPER_ADMIN
	IsConfigurable  bool                    `json:"is_configurable"` // true for custom roles
	UserCount       int                     `json:"user_count" db:"user_count"`
	PermissionCount int                     `json:"permission_count"`
	Permissions     []string                `json:"permissions"` // Flat list of "RESOURCE.ACTION"
	EffectiveAccess []EffectiveModuleAccess `json:"effective_access"`
}

// PermissionMatrixCatalog defines canonical resources and role permissions.
type PermissionMatrixCatalog struct {
	Resources        []PlatformResourceCatalog `json:"resources"`
	Roles            []CustomerRoleDetail      `json:"roles"`
	TotalRoles       int                       `json:"total_roles"`
	TotalPermissions int                       `json:"total_permissions"`
}

// UpdateCustomerUserRoleRequest specifies payload to reassign a customer user's role from SPortal.
type UpdateCustomerUserRoleRequest struct {
	RoleID   int64  `json:"role_id"`
	RoleName string `json:"role_name"`
	Role     string `json:"role"`
	Reason   string `json:"reason"`
}

// UnmarshalJSON enables accepting role_id either as integer (e.g. 7) or string (e.g. "OPERATIONS" or "7").
func (u *UpdateCustomerUserRoleRequest) UnmarshalJSON(data []byte) error {
	type Alias UpdateCustomerUserRoleRequest
	aux := struct {
		RawRoleID interface{} `json:"role_id"`
		*Alias
	}{
		Alias: (*Alias)(u),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.RawRoleID != nil {
		switch v := aux.RawRoleID.(type) {
		case float64:
			u.RoleID = int64(v)
		case string:
			clean := strings.TrimSpace(v)
			if id, err := strconv.ParseInt(clean, 10, 64); err == nil && id > 0 {
				u.RoleID = id
			} else if clean != "" {
				u.RoleName = clean
			}
		}
	}
	if u.RoleName == "" && u.Role != "" {
		u.RoleName = strings.TrimSpace(u.Role)
	}
	return nil
}

// InviteCustomerUserRequest specifies payload to invite a new customer user or initial Super Admin.
type InviteCustomerUserRequest struct {
	OrgID     int64  `json:"org_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	RoleID    int64  `json:"role_id"`
	RoleName  string `json:"role_name"`
}

// UpdateCustomerUserStatusRequest specifies payload to deactivate, reactivate or suspend a user.
type UpdateCustomerUserStatusRequest struct {
	Status string `json:"status"` // ACTIVE, INACTIVE, DISABLED, SUSPENDED
	Reason string `json:"reason"`
}

// CustomerRoleItem describes a tenant role available for assignment.
type CustomerRoleItem struct {
	ID          int64  `json:"id" db:"id"`
	OrgID       int64  `json:"org_id" db:"org_id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	UserCount   int    `json:"user_count" db:"user_count"`
}

// InvitationRecord represents an authoritative record in the invitations table.
type InvitationRecord struct {
	ID        int64     `json:"id" db:"id"`
	OrgID     int64     `json:"org_id" db:"org_id"`
	OrgName   string    `json:"org_name" db:"org_name"`
	RoleID    int64     `json:"role_id" db:"role_id"`
	RoleName  string    `json:"role_name" db:"role_name"`
	Email     string    `json:"email" db:"email"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	Status         string    `json:"status"` // PENDING, EXPIRED, ACCEPTED
	ActivationLink string    `json:"activation_link,omitempty"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// ============================================================================
// TASK S10: CUSTOMER USAGE, PLATFORM ANALYTICS, CONSUMPTION & ADOPTION
// ============================================================================

// UsageQuotaLimitItem represents consumption against legitimate subscription plan limits.
type UsageQuotaLimitItem struct {
	MetricKey      string `json:"metric_key"`
	Label          string `json:"label"`
	CurrentUsage   int    `json:"current_usage"`
	LimitAmount    *int   `json:"limit_amount"`
	Unlimited      bool   `json:"unlimited"`
	Remaining      int    `json:"remaining"`
	UtilizationPct int    `json:"utilization_pct"`
	Status         string `json:"status"` // NORMAL, APPROACHING_LIMIT, EXCEEDED, UNLIMITED
	Unit           string `json:"unit"`
}

// UsageThresholdAlert highlights approaching or exceeded quota limits.
type UsageThresholdAlert struct {
	MetricKey      string `json:"metric_key"`
	Severity       string `json:"severity"` // warning, critical
	Message        string `json:"message"`
	CurrentUsage   int    `json:"current_usage"`
	LimitAmount    int    `json:"limit_amount"`
	UtilizationPct int    `json:"utilization_pct"`
}

// ModuleAdoptionItem tracks customer adoption across each of the 15 LogisticsHQ platform modules.
type ModuleAdoptionItem struct {
	ModuleKey      string     `json:"module_key"`
	ModuleName     string     `json:"module_name"`
	Category       string     `json:"category"` // Commercial, Operations, Finance, Intelligence, Platform
	TotalActivity  int        `json:"total_activity"`
	PeriodActivity int        `json:"period_activity"`
	ActiveActors   int        `json:"active_actors"`
	Status         string     `json:"status"` // ACTIVE, LOW_ACTIVITY, NO_RECORDED_ACTIVITY, NOT_CONFIGURED
	LastActivityAt *time.Time `json:"last_activity_at"`
	DrillDownURL   string     `json:"drill_down_url"`
}

// AdoptionJourneyMilestone represents a step in the customer's onboarding and adoption progression.
type AdoptionJourneyMilestone struct {
	MilestoneKey  string     `json:"milestone_key"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Completed     bool       `json:"completed"`
	CompletedAt   *time.Time `json:"completed_at"`
	SequenceOrder int        `json:"sequence_order"`
}

// UsageTrendItem tracks monthly operational and activity trends over time.
type UsageTrendItem struct {
	MonthKey         string `json:"month_key"`  // "2026-03"
	MonthLabel       string `json:"month_label"` // "Mar 2026"
	ShipmentsCount   int    `json:"shipments_count"`
	RFQsCount        int    `json:"rfqs_count"`
	QuotesCount      int    `json:"quotes_count"`
	AITasksCount     int    `json:"ai_tasks_count"`
	AuditEventsCount int    `json:"audit_events_count"`
}

// AITaskTypeBreakdown classifies safe business-level AI execution records.
type AITaskTypeBreakdown struct {
	TaskType       string `json:"task_type"`
	TotalTasks     int    `json:"total_tasks"`
	CompletedTasks int    `json:"completed_tasks"`
	FailedTasks    int    `json:"failed_tasks"`
}

// AutomationUsageItem summarizes configured automation workflow status.
type AutomationUsageItem struct {
	ID                  int64      `json:"id"`
	Name                string     `json:"name"`
	AutomationType      string     `json:"automation_type"`
	IsEnabled           bool       `json:"is_enabled"`
	LastExecutionAt     *time.Time `json:"last_execution_at"`
	LastExecutionStatus string     `json:"last_execution_status"`
}

// CustomerUsageAnalytics is the complete root payload for customer usage and platform analytics.
type CustomerUsageAnalytics struct {
	OrgID                  int64                      `json:"org_id"`
	OrgName                string                     `json:"org_name"`
	PlanName               string                     `json:"plan_name"`
	PlanCode               string                     `json:"plan_code"`
	SubscriptionStatus     string                     `json:"subscription_status"`
	BillingCycle           string                     `json:"billing_cycle"`
	Period                 string                     `json:"period"` // current_month, last_30_days, last_90_days, ytd, all_time
	PeriodStartDate        time.Time                  `json:"period_start_date"`
	PeriodEndDate          time.Time                  `json:"period_end_date"`
	UsageStatus            string                     `json:"usage_status"` // NORMAL, HIGH_CONSUMPTION, APPROACHING_LIMIT, EXCEEDED

	// Core KPI Summary (All truth-backed from DB)
	ActiveUsersCount       int                        `json:"active_users_count"`
	TotalUsersCount        int                        `json:"total_users_count"`
	PendingInvitesCount    int                        `json:"pending_invites_count"`
	RFQsCount              int                        `json:"rfqs_count"`
	QuotesCount            int                        `json:"quotes_count"`
	BookingsCount          int                        `json:"bookings_count"`
	ShipmentsCount         int                        `json:"shipments_count"`
	ActiveShipmentsCount   int                        `json:"active_shipments_count"`
	ExceptionsCount        int                        `json:"exceptions_count"`
	InvoicesCount          int                        `json:"invoices_count"`
	TotalInvoicedAmount    float64                    `json:"total_invoiced_amount"`
	DocumentsCount         int                        `json:"documents_count"`
	AITasksCount           int                        `json:"ai_tasks_count"`
	CompletedAITasksCount  int                        `json:"completed_ai_tasks_count"`
	AutomationsCount       int                        `json:"automations_count"`
	ActiveAutomationsCount int                        `json:"active_automations_count"`
	IntegrationsCount      int                        `json:"integrations_count"`
	TotalAuditEvents       int                        `json:"total_audit_events"`
	ActiveCustomersCount   int                        `json:"active_customers_count,omitempty"`
	TotalCustomersCount    int                        `json:"total_customers_count,omitempty"`

	// Quotas, Utilization and Threshold Warnings
	QuotaLimits            []UsageQuotaLimitItem      `json:"quota_limits"`
	ThresholdAlerts        []UsageThresholdAlert      `json:"threshold_alerts"`

	// Module Adoption Matrix & Score
	ModuleAdoption         []ModuleAdoptionItem       `json:"module_adoption"`
	AdoptionScore          int                        `json:"adoption_score"` // Percentage of 15 modules adopted
	ActiveModulesCount     int                        `json:"active_modules_count"`
	TotalModulesCount      int                        `json:"total_modules_count"`

	// Customer Adoption Journey
	AdoptionJourney        []AdoptionJourneyMilestone `json:"adoption_journey"`

	// Truthful Historical Trends
	MonthlyTrends          []UsageTrendItem           `json:"monthly_trends"`
	HasSufficientTrendData bool                       `json:"has_sufficient_trend_data"`

	// AI & Automation Telemetry
	AITasksBreakdown       []AITaskTypeBreakdown      `json:"ai_tasks_breakdown"`
	RecentAutomations      []AutomationUsageItem      `json:"recent_automations"`

	// Customer Health Signals
	HealthSignals          []string                   `json:"health_signals"`
	HealthStatus           string                     `json:"health_status"`
}

// ---------------------------------------------------------------------------
// TASK S11: CUSTOMER HEALTH, CUSTOMER SUCCESS INTELLIGENCE & RISK SIGNALS
// ---------------------------------------------------------------------------

// HealthDimensionScore represents one of the 7 evaluated health dimensions.
type HealthDimensionScore struct {
	Dimension   string                 `json:"dimension"`
	Label       string                 `json:"label"`
	Score       int                    `json:"score"` // 0-100
	Weight      float64                `json:"weight"`
	Status      string                 `json:"status"` // EXCELLENT, HEALTHY, NEEDS_ATTENTION, AT_RISK, UNCONFIGURED
	Summary     string                 `json:"summary"`
	KeyMetrics  map[string]interface{} `json:"key_metrics"`
}

// HealthSignalItem represents an observed fact, calculated health signal, or predictive alert.
type HealthSignalItem struct {
	SignalKey    string    `json:"signal_key"`
	Type         string    `json:"type"` // FACT, CALCULATED_SIGNAL, PREDICTION, RECOMMENDATION
	Impact       string    `json:"impact"` // POSITIVE, NEUTRAL, WARNING, CRITICAL
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	SourceModule string    `json:"source_module"`
	ObservedAt   time.Time `json:"observed_at"`
}

// AIPredictionSignal represents a Phase 4 predictive intelligence model output.
type AIPredictionSignal struct {
	PredictionID    int64     `json:"prediction_id"`
	PredictionType  string    `json:"prediction_type"`
	RiskLevel       string    `json:"risk_level"`
	ConfidenceScore float64   `json:"confidence_score"`
	Statement       string    `json:"statement"`
	GeneratedAt     time.Time `json:"generated_at"`
}

// CustomerSuccessRecommendation represents a grounded recommendation for CS action.
type CustomerSuccessRecommendation struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Category         string    `json:"category"`
	Priority         string    `json:"priority"` // CRITICAL, HIGH, MEDIUM, LOW
	ActionType       string    `json:"action_type"`
	RequiresApproval bool      `json:"requires_approval"`
	GroundedEvidence string    `json:"grounded_evidence"`
	SuggestedOwner   string    `json:"suggested_owner"`
	CreatedAt        time.Time `json:"created_at"`
}

// CustomerSuccessActionHistoryItem records past interventions and actions executed.
type CustomerSuccessActionHistoryItem struct {
	ID            int64     `json:"id"`
	ActionType    string    `json:"action_type"`
	Status        string    `json:"status"` // PENDING, APPROVED, EXECUTED, REJECTED
	ActorName     string    `json:"actor_name"`
	Description   string    `json:"description"`
	ResultOutcome string    `json:"result_outcome"`
	ExecutedAt    time.Time `json:"executed_at"`
}

// CustomerNoteItem represents an internal customer success note.
type CustomerNoteItem struct {
	ID         int64     `json:"id" db:"id"`
	OrgID      int64     `json:"org_id" db:"org_id"`
	AuthorID   int64     `json:"author_id" db:"author_id"`
	AuthorName string    `json:"author_name" db:"author_name"`
	NoteType   string    `json:"note_type" db:"note_type"` // ONBOARDING, BUSINESS_REVIEW, SUPPORT_ESCALATION, RISK_INTERVENTION, GENERAL
	Content    string    `json:"content" db:"content"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// CreateCustomerNoteRequest payload for adding internal notes.
type CreateCustomerNoteRequest struct {
	NoteType string `json:"note_type"`
	Content  string `json:"content"`
}

// CustomerHealthDetail represents the comprehensive Customer Health & CS Intelligence payload.
type CustomerHealthDetail struct {
	OrgID                  int64                              `json:"org_id"`
	OrgName                string                             `json:"org_name"`
	PlanName               string                             `json:"plan_name"`
	PlanCode               string                             `json:"plan_code"`
	SubscriptionStatus     string                             `json:"subscription_status"`
	RenewalDate            *time.Time                         `json:"renewal_date"`
	AutoRenew              bool                               `json:"auto_renew"`
	DaysToRenewal          int                                `json:"days_to_renewal"`
	
	// Overall Health Assessment
	HealthState            string                             `json:"health_state"` // HEALTHY, GOOD, WATCH, AT_RISK, CRITICAL, INSUFFICIENT_DATA
	HealthScore            int                                `json:"health_score"` // 0-100
	PreviousHealthScore    int                                `json:"previous_health_score"`
	HealthTrend            string                             `json:"health_trend"` // IMPROVING, STABLE, DECLINING, INSUFFICIENT_DATA
	RiskLevel              string                             `json:"risk_level"` // LOW, MEDIUM, HIGH, CRITICAL
	RenewalRisk            string                             `json:"renewal_risk"` // LOW, ELEVATED, HIGH
	ChurnProbabilityPct    float64                            `json:"churn_probability_pct"`
	DataSufficiency        string                             `json:"data_sufficiency"` // HIGH, MODERATE, LIMITED, INSUFFICIENT
	ConfidenceScore        float64                            `json:"confidence_score"`
	Summary                string                             `json:"summary"`
	EvaluatedAt            time.Time                          `json:"evaluated_at"`

	// 7 Detailed Dimensions
	Dimensions             []HealthDimensionScore             `json:"dimensions"`

	// Why is the customer in this state? (Contributing Signals)
	ContributingSignals    []HealthSignalItem                 `json:"contributing_signals"`
	ObservedFacts          []string                           `json:"observed_facts"`

	// Predictive Risk Models
	PredictiveRiskSignals  []AIPredictionSignal               `json:"predictive_risk_signals"`

	// Grounded Customer Success Recommendations
	Recommendations        []CustomerSuccessRecommendation    `json:"recommendations"`

	// Action System & Execution History
	ActionHistory          []CustomerSuccessActionHistoryItem `json:"action_history"`

	// Internal Customer Success Notes
	Notes                  []CustomerNoteItem                 `json:"notes"`
}

// ==============================================================================
// TASK S12: SPORTAL CUSTOMER INTEGRATIONS & CONNECTIVITY MANAGEMENT
// ==============================================================================

// CustomerIntegrationsOverview represents the complete integrations and connectivity state for a customer.
type CustomerIntegrationsOverview struct {
	OrgID                int64                      `json:"org_id"`
	OrgName              string                     `json:"org_name"`
	TotalConfigured      int                        `json:"total_configured"`
	ActiveConnected      int                        `json:"active_connected"`
	DegradedCount        int                        `json:"degraded_count"`
	ErrorCount           int                        `json:"error_count"`
	OverallHealthScore   int                        `json:"overall_health_score"` // 0-100
	OverallHealthStatus  string                     `json:"overall_health_status"` // HEALTHY, DEGRADED, NEEDS_ATTENTION, UNCONFIGURED
	Items                []CustomerIntegrationDetail `json:"items"`
	WebhooksSummary      CustomerWebhookSummary     `json:"webhooks_summary"`
	RecentWebhooks       []CustomerWebhookEventItem `json:"recent_webhooks"`
	RecentSyncJobs       []CustomerSyncJobItem      `json:"recent_sync_jobs"`
	AvailableProviders   []AvailableProviderItem    `json:"available_providers"`
	CarrierCatalog       []AvailableProviderItem    `json:"carrier_catalog"`
	EvaluatedAt          time.Time                  `json:"evaluated_at"`
}

// CustomerIntegrationDetail represents full operational state for an individual integration.
type CustomerIntegrationDetail struct {
	ID                 int64                  `json:"id"`
	Category           string                 `json:"category"` // CARRIER, EMAIL, SMS, STORAGE, TEXTRACT, WEBHOOK
	ProviderName       string                 `json:"provider_name"`
	DisplayName        string                 `json:"display_name"`
	Description        string                 `json:"description"`
	BusinessPurpose    string                 `json:"business_purpose"`
	DependentWorkflows []string               `json:"dependent_workflows"`
	Status             string                 `json:"status"` // CONNECTED, HEALTHY, DEGRADED, ERROR, CONFIGURATION_REQUIRED, DISABLED, NOT_CONFIGURED
	IsEnabled          bool                   `json:"is_enabled"`
	IsConfigured       bool                   `json:"is_configured"`
	Environment        string                 `json:"environment"` // PRODUCTION, SANDBOX, SYSTEM_DEFAULT, NOT_CONFIGURED
	CredentialState    string                 `json:"credential_state"` // CONFIGURED, MASKED, MISSING, SYSTEM_DEFAULT
	SafeConfig         map[string]interface{} `json:"safe_config"`
	LastSyncedAt       *time.Time             `json:"last_synced_at"`
	LastSuccessAt      *time.Time             `json:"last_success_at"`
	LastFailureAt      *time.Time             `json:"last_failure_at"`
	LastError          string                 `json:"last_error"`
	SyncStatus         string                 `json:"sync_status"` // SUCCESS, FAILED, SYNCING, IDLE
	EventCount30d      int                    `json:"event_count_30d"`
	FailureCount30d    int                    `json:"failure_count_30d"`
	HealthScore        int                    `json:"health_score"` // 0-100
	HealthMessage      string                 `json:"health_message"`
	CanTest            bool                   `json:"can_test"`
	CanSync            bool                   `json:"can_sync"`
	CanToggle          bool                   `json:"can_toggle"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// CustomerWebhookSummary represents 30-day webhook activity metrics for the customer.
type CustomerWebhookSummary struct {
	TotalReceived30d int `json:"total_received_30d"`
	ProcessedCount   int `json:"processed_count"`
	VerifiedCount    int `json:"verified_count"`
	DuplicateCount   int `json:"duplicate_count"`
	FailedCount      int `json:"failed_count"`
	DeadLetterCount  int `json:"dead_letter_count"`
}

// CustomerWebhookEventItem represents an individual ingested webhook event.
type CustomerWebhookEventItem struct {
	ID             int64      `json:"id"`
	Source         string     `json:"source"` // EXTERNAL, CARRIER
	Provider       string     `json:"provider"`
	EventType      string     `json:"event_type"`
	Status         string     `json:"status"` // PROCESSED, VERIFIED, RECEIVED, REJECTED, DUPLICATE_IGNORED, DEAD_LETTER
	RejectionReason string    `json:"rejection_reason,omitempty"`
	CorrelationID  string     `json:"correlation_id"`
	PayloadPreview string     `json:"payload_preview"`
	ReceivedAt     time.Time  `json:"received_at"`
	ProcessedAt    *time.Time `json:"processed_at,omitempty"`
}

// CustomerSyncJobItem represents a carrier synchronization job.
type CustomerSyncJobItem struct {
	ID               int64      `json:"id"`
	CarrierSCAC      string     `json:"carrier_scac"`
	CarrierName      string     `json:"carrier_name"`
	Operation        string     `json:"operation"`
	Status           string     `json:"status"`
	RecordsProcessed int        `json:"records_processed"`
	RecordsCreated   int        `json:"records_created"`
	RecordsUpdated   int        `json:"records_updated"`
	RecordsFailed    int        `json:"records_failed"`
	ErrorCode        string     `json:"error_code,omitempty"`
	ErrorMessage     string     `json:"error_message,omitempty"`
	CorrelationID    string     `json:"correlation_id"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

// AvailableProviderItem represents a system-supported integration provider.
type AvailableProviderItem struct {
	Code                  string   `json:"code"`
	Name                  string   `json:"name"`
	Category              string   `json:"category"`
	Modes                 []string `json:"modes,omitempty"`
	SupportedCapabilities []string `json:"supported_capabilities"`
	Description           string   `json:"description"`
	IsConfigured          bool     `json:"is_configured"`
}

// IntegrationActionRequest represents a mutation request (e.g. toggle or test).
type IntegrationActionRequest struct {
	IntegrationType string `json:"integration_type"`
	ProviderName    string `json:"provider_name"`
	Action          string `json:"action"` // TOGGLE, TEST_CONNECTION, SYNC
	Enabled         *bool  `json:"enabled,omitempty"`
}

// IntegrationActionResult represents the outcome of an integration operation.
type IntegrationActionResult struct {
	Success      bool                   `json:"success"`
	Message      string                 `json:"message"`
	Status       string                 `json:"status"`
	TestedAt     time.Time              `json:"tested_at"`
	LatencyMs    int64                  `json:"latency_ms,omitempty"`
	AuditLogID   int64                  `json:"audit_log_id,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// =========================================================================
// Task S13: Customer Documents, Compliance, Contracts & Customer Records
// =========================================================================

// DocumentListParams specifies filtering and pagination for customer documents.
type DocumentListParams struct {
	Search       string `json:"search"`
	DocType      string `json:"doc_type"`
	Status       string `json:"status"`
	ExpiryFilter string `json:"expiry_filter"` // all, valid, expiring_soon, expired
	Page         int    `json:"page"`
	Limit        int    `json:"limit"`
	SortBy       string `json:"sort_by"`
	SortDir      string `json:"sort_dir"`
}

// CustomerDocumentDetail represents complete metadata, OCR, and discrepancy data for a document.
type CustomerDocumentDetail struct {
	CustomerDocumentItem
	ExtractedData      map[string]interface{}     `json:"extracted_data,omitempty"`
	RawOcrText         string                     `json:"raw_ocr_text,omitempty"`
	AISummary          string                     `json:"ai_summary,omitempty"`
	Discrepancies      []DocumentDiscrepancyItem  `json:"discrepancies,omitempty"`
	AuditLog           []DocumentAuditItem        `json:"audit_log,omitempty"`
}

// DocumentAuditItem represents an audit entry for a document lifecycle event.
type DocumentAuditItem struct {
	Action    string    `json:"action"`
	ActorName string    `json:"actor_name"`
	Timestamp time.Time `json:"timestamp"`
	Details   string    `json:"details"`
}

// DocumentDiscrepancyItem represents a specific field mismatch detected during OCR/Textract verification.
type DocumentDiscrepancyItem struct {
	ID              int64     `json:"id" db:"id"`
	OrgID           int64     `json:"org_id" db:"org_id"`
	ShipmentID      int64     `json:"shipment_id" db:"shipment_id"`
	DocumentID      int64     `json:"document_id" db:"document_id"`
	DiscrepancyType string    `json:"discrepancy_type" db:"discrepancy_type"`
	Severity        string    `json:"severity" db:"severity"`
	Field           string    `json:"field" db:"field"`
	FieldName       string    `json:"field_name" db:"field_name"`
	ExpectedValue   string    `json:"expected_value" db:"expected_value"`
	ActualValue     string    `json:"actual_value" db:"actual_value"`
	Description     string    `json:"description" db:"description"`
	SourceDocument  string    `json:"source_document" db:"source_document"`
	TargetDocument  string    `json:"target_document" db:"target_document"`
	Status          string    `json:"status" db:"status"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// CustomerDocumentsResponse represents paginated documents for an organization.
type CustomerDocumentsResponse struct {
	OrgID            int64                    `json:"org_id"`
	OrganizationID   int64                    `json:"organization_id,omitempty"`
	OrgName          string                   `json:"org_name"`
	OrganizationName string                   `json:"organization_name,omitempty"`
	Items            []CustomerDocumentDetail `json:"items,omitempty"`
	Documents        []CustomerDocumentItem   `json:"documents,omitempty"`
	Total            int                      `json:"total"`
	Page             int                      `json:"page"`
	Limit            int                      `json:"limit"`
	TotalPages       int                      `json:"total_pages"`
	Summary          CustomerDocumentsSummary `json:"summary"`
}

// CustomerDocumentsSummary summarizes customer document counts and health states.
type CustomerDocumentsSummary struct {
	TotalDocuments     int            `json:"total_documents"`
	VerifiedCount      int            `json:"verified_count"`
	PendingReviewCount int            `json:"pending_review_count"`
	DiscrepanciesCount int            `json:"discrepancies_count"`
	ExpiringSoonCount  int            `json:"expiring_soon_count"`
	ExpiredCount       int            `json:"expired_count"`
	OCRProcessedCount  int            `json:"ocr_processed_count"`
	ByTypeBreakdown    map[string]int `json:"by_type_breakdown,omitempty"`
	ByStatusBreakdown  map[string]int `json:"by_status_breakdown,omitempty"`
}

// DocumentSummaryStats alias for CustomerDocumentsSummary
type DocumentSummaryStats = CustomerDocumentsSummary

// CustomerComplianceRequirementItem represents an individual legal or operational compliance requirement.
type CustomerComplianceRequirementItem struct {
	ID                 int64      `json:"id"`
	OrgID              int64      `json:"org_id"`
	ContractID         int64      `json:"contract_id"`
	ContractRef        string     `json:"contract_ref"`
	RequirementType    string     `json:"requirement_type"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	ResponsibleParty   string     `json:"responsible_party"`
	ValidFrom          *time.Time `json:"valid_from,omitempty"`
	ValidUntil         *time.Time `json:"valid_until,omitempty"`
	Status             string     `json:"status"`
	EvidenceDocumentID string     `json:"evidence_document_id,omitempty"`
	VerificationDate   *time.Time `json:"verification_date,omitempty"`
	VerifiedBy         int64      `json:"verified_by,omitempty"`
	RiskSeverity       string     `json:"risk_severity"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DaysToExpiry       int        `json:"days_to_expiry"`
	IsExpiringSoon     bool       `json:"is_expiring_soon"`
	IsExpired          bool       `json:"is_expired"`
}

// CustomerComplianceOverview provides an aggregated view of customer compliance status and requirements.
type CustomerComplianceOverview struct {
	OrgID               int64                                  `json:"org_id"`
	OrgName             string                                 `json:"org_name"`
	ComplianceScore     float64                                `json:"compliance_score"` // 0 - 100
	OverallStatus       string                                 `json:"overall_status"`   // COMPLIANT, REVIEW_REQUIRED, NON_COMPLIANT
	TotalRequirements   int                                    `json:"total_requirements"`
	ValidCount          int                                    `json:"valid_count"`
	ExpiringSoonCount   int                                    `json:"expiring_soon_count"`
	ExpiredCount        int                                    `json:"expired_count"`
	MissingCount        int                                    `json:"missing_count"`
	PendingReviewCount  int                                    `json:"pending_review_count"`
	DiscrepanciesCount  int                                    `json:"discrepancies_count"`
	Requirements        []CustomerComplianceRequirementItem    `json:"requirements"`
	AIReviews           []CustomerComplianceAIReviewItem       `json:"ai_reviews"`
	MonitoringPlans     []CustomerComplianceMonitoringPlanItem `json:"monitoring_plans"`
	GeneratedAt         time.Time                              `json:"generated_at"`
}

// CustomerComplianceItem represents an individual legal or operational compliance requirement.
type CustomerComplianceItem struct {
	ID              int64      `json:"id" db:"id"`
	ContractID      *int64     `json:"contract_id,omitempty" db:"contract_id"`
	ContractRef     string     `json:"contract_ref,omitempty"`
	RequirementCode string     `json:"requirement_code" db:"requirement_code"`
	Title           string     `json:"title" db:"title"`
	RequirementType string     `json:"requirement_type" db:"requirement_type"`
	Status          string     `json:"status" db:"status"` // VALID, EXPIRING_SOON, EXPIRED, PENDING, MISSING
	Priority        string     `json:"priority" db:"priority"` // LOW, MEDIUM, HIGH, CRITICAL
	DueDate         *time.Time `json:"due_date,omitempty" db:"due_date"`
	EffectiveDate   *time.Time `json:"effective_date,omitempty" db:"effective_date"`
	DaysToDue       *int       `json:"days_to_due,omitempty"`
	ActionRequired  string     `json:"action_required,omitempty"`
}

// CustomerComplianceAIReviewItem represents an AI-evaluated contract compliance check.
type CustomerComplianceAIReviewItem struct {
	ID                   int64     `json:"id" db:"id"`
	ContractID           int64     `json:"contract_id" db:"contract_id"`
	ContractRef          string    `json:"contract_ref,omitempty"`
	RiskLevel            string    `json:"risk_level" db:"risk_level"`
	RiskScore            float64   `json:"risk_score" db:"risk_score"`
	ComplianceStatus     string    `json:"compliance_status" db:"compliance_status"`
	ExecutiveSummary     string    `json:"executive_summary" db:"executive_summary"`
	DeterministicSignals string    `json:"deterministic_signals,omitempty"`
	Recommendations      string    `json:"recommendations,omitempty"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
}

// CustomerComplianceMonitoringPlanItem represents an active compliance monitoring schedule.
type CustomerComplianceMonitoringPlanItem struct {
	ID            int64      `json:"id" db:"id"`
	PlanName      string     `json:"plan_name"`
	Scope         string     `json:"scope"`
	Frequency     string     `json:"frequency"`
	Status        string     `json:"status"`
	LastAuditDate *time.Time `json:"last_audit_date,omitempty"`
}

// CustomerContractsOverview summarizes customer legal and commercial contracts.
type CustomerContractsOverview struct {
	OrgID               int64                  `json:"org_id"`
	OrgName             string                 `json:"org_name"`
	TotalContracts      int                    `json:"total_contracts"`
	ActiveCount         int                    `json:"active_count"`
	ExpiringSoonCount   int                    `json:"expiring_soon_count"`
	ExpiredCount        int                    `json:"expired_count"`
	DraftCount          int                    `json:"draft_count"`
	TotalValueUSD       float64                `json:"total_value_usd"`
	Items               []CustomerContractItem `json:"items"`
}

// OrganizationDocumentHealthItem represents an organization's document health in platform overview.
type OrganizationDocumentHealthItem struct {
	OrganizationID     int64   `json:"organization_id"`
	OrganizationName   string  `json:"organization_name"`
	TotalDocuments     int     `json:"total_documents"`
	VerifiedCount      int     `json:"verified_count"`
	DiscrepanciesCount int     `json:"discrepancies_count"`
	ExpiringSoonCount  int     `json:"expiring_soon_count"`
	ExpiredCount       int     `json:"expired_count"`
	TotalContracts     int     `json:"total_contracts"`
	ComplianceScore    float64 `json:"compliance_score"`
}

// PlatformDocumentsOverview represents global platform metrics for documents, contracts, and compliance.
type PlatformDocumentsOverview struct {
	TotalDocuments         int                              `json:"total_documents"`
	VerifiedDocuments      int                              `json:"verified_documents"`
	DiscrepantDocuments    int                              `json:"discrepant_documents"`
	ExpiringSoonDocuments  int                              `json:"expiring_soon_documents"`
	ExpiredDocuments       int                              `json:"expired_documents"`
	OCRProcessedCount      int                              `json:"ocr_processed_count"`
	TotalContracts         int                              `json:"total_contracts"`
	ActiveContracts        int                              `json:"active_contracts"`
	AvgComplianceScore     float64                          `json:"avg_compliance_score"`
	AverageComplianceScore float64                          `json:"average_compliance_score"`
	Organizations          []OrganizationSummary            `json:"organizations,omitempty"`
	ByOrganization         []OrganizationDocumentHealthItem `json:"by_organization,omitempty"`
}

// ============================================================================
// TASK S16: SPORTAL AI, INTERNAL INTELLIGENCE & GOVERNED AI OPERATIONS TYPES
// ============================================================================

// SPortalAiQueryRequest represents an internal natural-language query from an authorized staff member.
type SPortalAiQueryRequest struct {
	Query          string                 `json:"query"`
	OrganizationID *int64                 `json:"organization_id,omitempty"`
	SessionID      string                 `json:"session_id,omitempty"`
	Route          string                 `json:"route,omitempty"`
	FilterContext  map[string]interface{} `json:"filter_context,omitempty"`
}

// SPortalAiSourceRef is a verified clickable link to an authoritative business record.
type SPortalAiSourceRef struct {
	RecordType string `json:"record_type"`
	RecordID   string `json:"record_id"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	Snippet    string `json:"snippet"`
}

// SPortalAiPredictionItem represents a forward-looking predictive signal.
type SPortalAiPredictionItem struct {
	SignalType      string  `json:"signal_type"`
	TargetEntity    string  `json:"target_entity"`
	RiskLevel       string  `json:"risk_level"`
	ConfidenceScore float64 `json:"confidence_score"`
	TimeHorizon     string  `json:"time_horizon"`
	SupportingFacts string  `json:"supporting_facts"`
}

// SPortalAiRecommendationItem represents an actionable recommendation for CSM or Operations.
type SPortalAiRecommendationItem struct {
	ID               int64   `json:"id,omitempty"`
	Category         string  `json:"category"`
	Title            string  `json:"title"`
	Description      string  `json:"description"`
	Priority         string  `json:"priority"`
	TargetOrgID      *int64  `json:"target_org_id,omitempty"`
	TargetOrgName    string  `json:"target_org_name,omitempty"`
	Confidence       float64 `json:"confidence"`
	RequiresApproval bool    `json:"requires_approval"`
	SuggestedAction  string  `json:"suggested_action,omitempty"`
}

// SPortalAiDraft represents an AI-generated draft communication or notice.
type SPortalAiDraft struct {
	DraftType   string `json:"draft_type"`
	Subject     string `json:"subject"`
	Recipient   string `json:"recipient,omitempty"`
	Body        string `json:"body"`
	Disclaimer  string `json:"disclaimer"`
	IsExecuted  bool   `json:"is_executed"`
}

// SPortalAiActionProposal represents a proposed action ready for Human Review (HITL).
type SPortalAiActionProposal struct {
	ActionType       string                 `json:"action_type"`
	ActionTitle      string                 `json:"action_title"`
	Description      string                 `json:"description"`
	Payload          map[string]interface{} `json:"payload"`
	RequiresApproval bool                   `json:"requires_approval"`
}

// SPortalAiQueryResponse represents the grounded, structured intelligence response.
type SPortalAiQueryResponse struct {
	SessionID          string                        `json:"session_id"`
	Answer             string                        `json:"answer"`
	ConfirmedFacts     []string                      `json:"confirmed_facts"`
	AiInterpretation   string                        `json:"ai_interpretation"`
	Predictions        []SPortalAiPredictionItem     `json:"predictions"`
	Recommendations    []SPortalAiRecommendationItem `json:"recommendations"`
	SourceReferences   []SPortalAiSourceRef          `json:"source_references"`
	Draft              *SPortalAiDraft               `json:"draft,omitempty"`
	ActionProposals    []SPortalAiActionProposal     `json:"action_proposals"`
	Confidence         float64                       `json:"confidence"`
	DataSufficiency    string                        `json:"data_sufficiency"`
	MissingInformation []string                      `json:"missing_information"`
	SafetyStatus       string                        `json:"safety_status"`
	SuggestedFollowups []string                      `json:"suggested_followups"`
	CorrelationID      string                        `json:"correlation_id"`
}

// SPortalAiActionRequest represents an action execution or approval submission.
type SPortalAiActionRequest struct {
	ActionType     string                 `json:"action_type"`
	ActionTitle    string                 `json:"action_title"`
	OrganizationID int64                  `json:"organization_id"`
	Payload        map[string]interface{} `json:"payload"`
}

// SPortalAiActionResponse represents the result of an action execution or approval gate.
type SPortalAiActionResponse struct {
	ActionID      string `json:"action_id"`
	Status        string `json:"status"` // EXECUTED | PENDING_APPROVAL | DRAFT_SAVED
	Summary       string `json:"summary"`
	ApprovalID    *int64 `json:"approval_id,omitempty"`
	CorrelationID string `json:"correlation_id"`
}

// SPortalAiWorkforceAgent represents an autonomous specialist agent in SPortal.
type SPortalAiWorkforceAgent struct {
	AgentID       string `json:"agent_id" db:"agent_id"`
	AgentType     string `json:"agent_type" db:"agent_type"`
	AutonomyLevel string `json:"autonomy_level" db:"autonomy_level"`
	HealthStatus  string `json:"health_status" db:"health_status"`
	IsEnabled     bool   `json:"is_enabled" db:"is_enabled"`
	Description   string `json:"description"`
}

// SPortalAiWorkforceOverview represents the live operational status of the AI workforce.
type SPortalAiWorkforceOverview struct {
	TotalAgents                int                           `json:"total_agents"`
	ActiveAgents               int                           `json:"active_agents"`
	Agents                     []SPortalAiWorkforceAgent     `json:"agents"`
	TotalTasksProcessed        int                           `json:"total_tasks_processed"`
	ActiveRecommendationsCount int                           `json:"active_recommendations_count"`
	PendingApprovalsCount      int                           `json:"pending_approvals_count"`
	RecentRecommendations      []SPortalAiRecommendationItem `json:"recent_recommendations"`
	GovernanceSafetyStatus     string                        `json:"governance_safety_status"`
}

// SPortalPlatformSetting represents a platform-wide configuration setting.
type SPortalPlatformSetting struct {
	SettingKey   string    `json:"setting_key" db:"setting_key"`
	SettingValue string    `json:"setting_value" db:"setting_value"`
	Category     string    `json:"category" db:"category"`
	DataType     string    `json:"data_type" db:"data_type"`
	Description  *string   `json:"description,omitempty" db:"description"`
	IsSensitive  bool      `json:"is_sensitive" db:"is_sensitive"`
	UpdatedBy    *string   `json:"updated_by,omitempty" db:"updated_by"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// SPortalPlatformSettingUpdateRequest contains parameters for updating a platform setting.
type SPortalPlatformSettingUpdateRequest struct {
	SettingValue string `json:"setting_value"`
}

// SPortalFeatureFlag represents an AI or system governance feature flag.
type SPortalFeatureFlag struct {
	ID                int64     `json:"id" db:"id"`
	OrgID             int64     `json:"org_id" db:"org_id"`
	FlagKey           string    `json:"flag_key" db:"flag_key"`
	FlagName          string    `json:"flag_name" db:"flag_name"`
	IsEnabled         bool      `json:"is_enabled" db:"is_enabled"`
	MaxAutonomyLevel  int       `json:"max_autonomy_level" db:"max_autonomy_level"`
	RequiresApproval  bool      `json:"requires_approval" db:"requires_approval"`
	Description       *string   `json:"description,omitempty" db:"description"`
	UpdatedByID       *int64    `json:"updated_by_id,omitempty" db:"updated_by_id"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// SPortalFeatureFlagUpdateRequest contains parameters for toggling or updating a feature flag.
type SPortalFeatureFlagUpdateRequest struct {
	IsEnabled        bool `json:"is_enabled"`
	RequiresApproval bool `json:"requires_approval"`
	MaxAutonomyLevel *int `json:"max_autonomy_level,omitempty"`
}

// SPortalAutonomyPolicy represents an operational module autonomy policy.
type SPortalAutonomyPolicy struct {
	ID                     int64     `json:"id" db:"id"`
	OrgID                  int64     `json:"org_id" db:"org_id"`
	Module                 string    `json:"module" db:"module"`
	AutonomyLevel          string    `json:"autonomy_level" db:"autonomy_level"`
	RequiresApproval       bool      `json:"requires_approval" db:"requires_approval"`
	MaxMonetaryThreshold   float64   `json:"max_monetary_threshold" db:"max_monetary_threshold"`
	CustomerImpactLevel    string    `json:"customer_impact_threshold" db:"customer_impact_threshold"`
	MinConfidenceThreshold float64   `json:"min_confidence_threshold" db:"min_confidence_threshold"`
	EmergencyStop          bool      `json:"emergency_stop" db:"emergency_stop"`
	IsActive               bool      `json:"is_active" db:"is_active"`
	PolicyVersion          int       `json:"policy_version" db:"policy_version"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

// SPortalEmergencyHaltRequest contains parameters for triggering or clearing emergency stop.
type SPortalEmergencyHaltRequest struct {
	HaltActive bool    `json:"halt_active"`
	Module     *string `json:"module,omitempty"` // nil = all modules
	Reason     string  `json:"reason"`
}

// SPortalIntegrationSetting represents safe public status of an external integration (NO SECRETS).
type SPortalIntegrationSetting struct {
	ID              int64      `json:"id" db:"id"`
	OrgID           int64      `json:"org_id" db:"org_id"`
	IntegrationType string     `json:"integration_type" db:"integration_type"`
	ProviderName    string     `json:"provider_name" db:"provider_name"`
	IsEnabled       bool       `json:"is_enabled" db:"is_enabled"`
	Status          string     `json:"status" db:"status"` // ENABLED, DISABLED, NOT_CONFIGURED, CONNECTED, ERROR
	LastHealthCheck *time.Time `json:"last_health_check,omitempty" db:"last_health_check"`
	HealthMessage   *string    `json:"health_message,omitempty" db:"health_message"`
	MaskedIdentity  string     `json:"masked_identity"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// SPortalIntegrationToggleRequest parameters for toggling an integration.
type SPortalIntegrationToggleRequest struct {
	IsEnabled bool   `json:"is_enabled"`
	Reason    string `json:"reason,omitempty"`
}

// SPortalUserNotificationPreferences represents internal user notification settings.
type SPortalUserNotificationPreferences struct {
	ID                   int64     `json:"id" db:"id"`
	UserID               int64     `json:"user_id" db:"user_id"`
	OrgID                int64     `json:"org_id" db:"org_id"`
	MinSeverity          string    `json:"min_severity" db:"min_severity"`
	InAppEnabled         bool      `json:"in_app_enabled" db:"in_app_enabled"`
	AssignedOnly         bool      `json:"assigned_only" db:"assigned_only"`
	ApprovalsEnabled     bool      `json:"approvals_enabled" db:"approvals_enabled"`
	AutomationsEnabled   bool      `json:"automations_enabled" db:"automations_enabled"`
	RecommendationsEnabled bool    `json:"recommendations_enabled" db:"recommendations_enabled"`
	FinanceEnabled       bool      `json:"finance_enabled" db:"finance_enabled"`
	OperationsEnabled    bool      `json:"operations_enabled" db:"operations_enabled"`
	ComplianceEnabled    bool      `json:"compliance_enabled" db:"compliance_enabled"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// SPortalProfileUpdateRequest contains permitted profile fields for internal staff.
type SPortalProfileUpdateRequest struct {
	FirstName              *string `json:"first_name,omitempty"`
	LastName               *string `json:"last_name,omitempty"`
	MinSeverity            *string `json:"min_severity,omitempty"`
	InAppEnabled           *bool   `json:"in_app_enabled,omitempty"`
	ApprovalsEnabled       *bool   `json:"approvals_enabled,omitempty"`
	AutomationsEnabled     *bool   `json:"automations_enabled,omitempty"`
	RecommendationsEnabled *bool   `json:"recommendations_enabled,omitempty"`
	FinanceEnabled         *bool   `json:"finance_enabled,omitempty"`
	OperationsEnabled      *bool   `json:"operations_enabled,omitempty"`
	ComplianceEnabled      *bool   `json:"compliance_enabled,omitempty"`
}

// SPortalAuditLogEntry represents an administrative audit record.
type SPortalAuditLogEntry struct {
	ID            int64     `json:"id" db:"id"`
	OrgID         int64     `json:"org_id" db:"org_id"`
	ActorID       *int64    `json:"actor_id,omitempty" db:"user_id"`
	ActorName     string    `json:"actor_name" db:"actor_name"`
	ActorRole     string    `json:"actor_role" db:"actor_role"`
	Action        string    `json:"action" db:"action"`
	Module        string    `json:"module" db:"module"`
	ResourceType  string    `json:"resource_type" db:"resource_type"`
	ResourceID    *string   `json:"resource_id,omitempty" db:"resource_id"`
	Description   string    `json:"description" db:"description"`
	Result        string    `json:"result" db:"result"`
	IPAddress     *string   `json:"ip_address,omitempty" db:"ip_address"`
	CorrelationID *string   `json:"correlation_id,omitempty"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// SPortalOperationsHealthSummary represents truthful operational status.
type SPortalOperationsHealthSummary struct {
	BackendStatus        string    `json:"backend_status"`
	DatabaseStatus       string    `json:"database_status"`
	AiSidecarStatus      string    `json:"ai_sidecar_status"`
	EventMeshStatus      string    `json:"event_mesh_status"`
	DeadLetterCount      int       `json:"dead_letter_count"`
	ActiveWorkersCount   int       `json:"active_workers_count"`
	PendingTaskCount     int       `json:"pending_task_count"`
	ActiveIntegrations   int       `json:"active_integrations"`
	UptimeSeconds        int64     `json:"uptime_seconds"`
	LastEvaluatedAt      time.Time `json:"last_evaluated_at"`
}

// SPortalSettingsOverviewResponse aggregates full settings control panel state.
type SPortalSettingsOverviewResponse struct {
	User                   SPortalUserInfo                     `json:"user"`
	Role                   SPortalRoleInfo                     `json:"role"`
	Preferences            *SPortalUserNotificationPreferences `json:"preferences,omitempty"`
	PlatformSettings       []SPortalPlatformSetting            `json:"platform_settings"`
	FeatureFlags           []SPortalFeatureFlag                `json:"feature_flags"`
	AutonomyPolicies       []SPortalAutonomyPolicy             `json:"autonomy_policies"`
	GlobalEmergencyHalt    bool                                `json:"global_emergency_halt"`
	Integrations           []SPortalIntegrationSetting         `json:"integrations"`
	OperationsHealth       SPortalOperationsHealthSummary      `json:"operations_health"`
	RecentAdministrativeAudits []SPortalAuditLogEntry          `json:"recent_audits"`
}

// -----------------------------------------------------------------------------
// P12: Support Center, Activity Timeline, Notifications & Forensic Audit
// -----------------------------------------------------------------------------

// SPortalSupportCase represents an authoritative operational support case / ticket.
type SPortalSupportCase struct {
	ID              int64                    `json:"id" db:"id"`
	OrgID           int64                    `json:"org_id" db:"org_id"`
	OrgName         string                   `json:"org_name" db:"org_name"`
	ShipmentID      *int64                   `json:"shipment_id,omitempty" db:"shipment_id"`
	ShipmentRef     string                   `json:"shipment_ref" db:"shipment_ref"`
	CaseCode        string                   `json:"case_code" db:"case_code"`
	Subject         string                   `json:"subject" db:"subject"`
	Description     string                   `json:"description" db:"description"`
	Category        string                   `json:"category" db:"category"`
	Priority        string                   `json:"priority" db:"priority"`
	Status          string                   `json:"status" db:"status"`
	AssignedOwner   string                   `json:"assigned_owner" db:"assigned_owner"`
	ResolutionNotes *string                  `json:"resolution_notes,omitempty" db:"resolution_notes"`
	ResolvedAt      *time.Time               `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolvedBy      *string                  `json:"resolved_by,omitempty" db:"resolved_by"`
	AiSummary       *string                  `json:"ai_summary,omitempty" db:"ai_summary"`
	SLAState        string                   `json:"sla_state" db:"sla_state"`
	LatestActivity  time.Time                `json:"latest_activity" db:"latest_activity"`
	CreatedAt       time.Time                `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time                `json:"updated_at" db:"updated_at"`
	InternalNotes   []CustomerNoteItem       `json:"internal_notes,omitempty"`
	AuditHistory    []OrganizationActivityItem `json:"audit_history,omitempty"`
}

// SPortalSupportCasesOverview aggregates support cases with truthful KPI metrics.
type SPortalSupportCasesOverview struct {
	Items              []SPortalSupportCase `json:"items"`
	TotalCases         int                  `json:"total_cases"`
	OpenCases          int                  `json:"open_cases"`
	CriticalCases      int                  `json:"critical_cases"`
	ResolvedCases      int                  `json:"resolved_cases"`
	AvgResolutionHours float64              `json:"avg_resolution_hours"`
}

// SPortalUpdateCaseStatusRequest defines payload for updating case lifecycle.
type SPortalUpdateCaseStatusRequest struct {
	Status          string `json:"status"`
	Severity        string `json:"severity,omitempty"`
	ResolutionNotes string `json:"resolution_notes,omitempty"`
}

// SPortalAddCaseNoteRequest defines payload for adding internal case note.
type SPortalAddCaseNoteRequest struct {
	Content        string `json:"content"`
	IsInternalOnly bool   `json:"is_internal_only"`
}

// SPortalCreateCaseRequest defines payload for creating a new support case.
type SPortalCreateCaseRequest struct {
	OrgID         int64  `json:"org_id"`
	ShipmentID    int64  `json:"shipment_id"`
	ExceptionType string `json:"exception_type"`
	Severity      string `json:"severity"`
	Title         string `json:"title"`
	Description   string `json:"description"`
}

// SPortalNotificationItem represents a real notification entry.
type SPortalNotificationItem struct {
	ID               int64      `json:"id" db:"id"`
	OrgID            int64      `json:"org_id" db:"org_id"`
	OrgName          string     `json:"org_name" db:"org_name"`
	SourceModule     string     `json:"source_module" db:"source_module"`
	SourceRecordType string     `json:"source_record_type" db:"source_record_type"`
	SourceRecordID   string     `json:"source_record_id" db:"source_record_id"`
	NotificationType string     `json:"notification_type" db:"notification_type"`
	Title            string     `json:"title" db:"title"`
	Message          string     `json:"message" db:"message"`
	Severity         string     `json:"severity" db:"severity"`
	Priority         string     `json:"priority" db:"priority"`
	Status           string     `json:"status" db:"status"`
	DeliveryStatus   string     `json:"delivery_status" db:"delivery_status"`
	IsRead           bool       `json:"is_read" db:"is_read"`
	ReadAt           *time.Time `json:"read_at,omitempty" db:"read_at"`
	IsAcknowledged   bool       `json:"is_acknowledged" db:"is_acknowledged"`
	AcknowledgedAt   *time.Time `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	ActionRequired   bool       `json:"action_required" db:"action_required"`
	ActionURL        string     `json:"action_url" db:"action_url"`
	AiSummary        string     `json:"ai_summary" db:"ai_summary"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// SPortalNotificationsOverview aggregates notifications with KPI counters.
type SPortalNotificationsOverview struct {
	Items               []SPortalNotificationItem `json:"items"`
	TotalCount          int                       `json:"total_count"`
	UnreadCount         int                       `json:"unread_count"`
	CriticalCount       int                       `json:"critical_count"`
	ActionRequiredCount int                       `json:"action_required_count"`
}

// SPortalUnifiedActivityItem represents an entry in the unified activity stream.
type SPortalUnifiedActivityItem struct {
	ID            string    `json:"id" db:"id"`
	OrgID         int64     `json:"org_id" db:"org_id"`
	OrgName       string    `json:"org_name" db:"org_name"`
	ActorName     string    `json:"actor_name" db:"actor_name"`
	ActivityType  string    `json:"activity_type" db:"activity_type"`
	Category      string    `json:"category" db:"category"`
	Action        string    `json:"action" db:"action"`
	Description   string    `json:"description" db:"description"`
	Entity        string    `json:"entity" db:"entity"`
	Result        string    `json:"result" db:"result"`
	CorrelationID string    `json:"correlation_id" db:"correlation_id"`
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`
}

// SPortalAuditSearchFilter defines filtering parameters for audit logs.
type SPortalAuditSearchFilter struct {
	OrgID     int64  `json:"org_id"`
	Module    string `json:"module"`
	Action    string `json:"action"`
	Actor     string `json:"actor"`
	Result    string `json:"result"`
	Search    string `json:"search"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

// ── Demo Requests ────────────────────────────────────────────────────────────

// DemoRequest represents an inbound demo request from the public website.
type DemoRequest struct {
	ID             int64      `json:"id"              db:"id"`
	FullName       string     `json:"full_name"       db:"full_name"`
	Email          string     `json:"email"           db:"email"`
	CompanyName    string     `json:"company_name"    db:"company_name"`
	Phone          *string    `json:"phone"           db:"phone"`
	Country        *string    `json:"country"         db:"country"`
	CompanySize    *string    `json:"company_size"    db:"company_size"`
	Message        *string    `json:"message"         db:"message"`
	ShipmentVolume *string    `json:"shipment_volume" db:"shipment_volume"`
	Services       *string    `json:"services"        db:"services"`
	Status         string     `json:"status"          db:"status"`
	Notes          *string    `json:"notes"           db:"notes"`
	AssignedTo     *string    `json:"assigned_to"     db:"assigned_to"`
	Source         string     `json:"source"          db:"source"`
	IPAddress      *string    `json:"ip_address"      db:"ip_address"`
	UserAgent      *string    `json:"user_agent"      db:"user_agent"`
	ContactedAt    *time.Time `json:"contacted_at"    db:"contacted_at"`
	ConvertedAt    *time.Time `json:"converted_at"    db:"converted_at"`
	CreatedAt      time.Time  `json:"created_at"      db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"      db:"updated_at"`
}

// CreateDemoRequestPayload is the public-facing submission payload.
type CreateDemoRequestPayload struct {
	FullName       string `json:"full_name"`
	Email          string `json:"email"`
	CompanyName    string `json:"company_name"`
	Phone          string `json:"phone,omitempty"`
	Country        string `json:"country,omitempty"`
	CompanySize    string `json:"company_size,omitempty"`
	Message        string `json:"message,omitempty"`
	ShipmentVolume string `json:"shipment_volume,omitempty"`
	Services       string `json:"services,omitempty"`
}

// DemoRequestListParams defines filtering/pagination for listing demo requests.
type DemoRequestListParams struct {
	Search string `json:"search"`
	Status string `json:"status"`
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
}

// DemoRequestListResult is the paginated list response for demo requests.
type DemoRequestListResult struct {
	Items      []DemoRequest `json:"items"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

// UpdateDemoRequestPayload is for SPortal staff to update demo request status/notes.
type UpdateDemoRequestPayload struct {
	Status     *string `json:"status,omitempty"`
	Notes      *string `json:"notes,omitempty"`
	AssignedTo *string `json:"assigned_to,omitempty"`
}
