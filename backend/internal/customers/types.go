package customers

import (
	"time"
)

// Customer Status Constants
const (
	StatusActive     = "ACTIVE"
	StatusInactive   = "INACTIVE"
	StatusOnboarding = "ONBOARDING"
	StatusAtRisk     = "AT_RISK"
	StatusChurned    = "CHURNED"
)

// Customer Type / Account Role Constants
const (
	TypeShipper     = "SHIPPER"
	TypeImporter    = "IMPORTER"
	TypeExporter    = "EXPORTER"
	TypeTrader      = "TRADER"
	TypeManufacturer= "MANUFACTURER"
	TypeLogisticsCo = "LOGISTICS_CO"
	TypeConsignee   = "CONSIGNEE"
	TypeBuyer       = "BUYER"
	TypeSeller      = "SELLER"
	TypeBroker      = "BROKER"
	TypeOther       = "OTHER"
)

// Address Type Constants
const (
	AddressTypeBilling          = "BILLING"
	AddressTypeShipping         = "SHIPPING"
	AddressTypeWarehouse        = "WAREHOUSE"
	AddressTypeRegisteredOffice = "REGISTERED_OFFICE"
	AddressTypePortFacility     = "PORT_FACILITY"
)

// Credit Status Constants
const (
	CreditStatusGoodStanding   = "GOOD_STANDING"
	CreditStatusReviewRequired = "REVIEW_REQUIRED"
	CreditStatusOnHold         = "ON_HOLD"
	CreditStatusNoLimit        = "NO_CREDIT_LIMIT"
)

// Contact Role Constants
const (
	ContactRoleCommercial    = "COMMERCIAL"
	ContactRoleOperations    = "OPERATIONS"
	ContactRoleLogistics     = "LOGISTICS"
	ContactRoleFinance       = "FINANCE"
	ContactRoleManagement    = "MANAGEMENT"
	ContactRoleDecisionMaker = "DECISION_MAKER"
	ContactRoleBilling       = "BILLING"
	ContactRoleOther         = "OTHER"
)

// Customer represents a customer entity in global freight forwarding
type Customer struct {
	ID               int64      `db:"id" json:"id"`
	OrgID            int64      `db:"org_id" json:"org_id"`
	CustomerCode     string     `db:"customer_code" json:"customer_code"`
	Name             string     `db:"name" json:"name"`
	TradingName      *string    `db:"trading_name" json:"trading_name,omitempty"`
	CustomerType     string     `db:"customer_type" json:"customer_type"`
	Domain           *string    `db:"domain" json:"domain,omitempty"`
	Industry         *string    `db:"industry" json:"industry,omitempty"`
	TaxID            *string    `db:"tax_id" json:"tax_id,omitempty"`
	PANNumber        *string    `db:"pan_number" json:"pan_number,omitempty"`
	EORINumber       *string    `db:"eori_number" json:"eori_number,omitempty"`
	Currency         string     `db:"currency" json:"currency"`
	PaymentTerms     string     `db:"payment_terms" json:"payment_terms"`
	CreditLimit      float64    `db:"credit_limit" json:"credit_limit"`
	CreditStatus     string     `db:"credit_status" json:"credit_status"`
	CommercialNotes  *string    `db:"commercial_notes" json:"commercial_notes,omitempty"`
	HealthScore      int        `db:"health_score" json:"health_score"`
	HealthStatus     string     `db:"-" json:"health_status"`
	ActivityTrend    string     `db:"-" json:"activity_trend"`
	AccountOwnerID   *int64     `db:"account_owner_id" json:"account_owner_id,omitempty"`
	SecondaryOwnerID *int64     `db:"secondary_owner_id" json:"secondary_owner_id,omitempty"`
	Website          *string    `db:"website" json:"website,omitempty"`
	Country          *string    `db:"country" json:"country,omitempty"`
	City             *string    `db:"city" json:"city,omitempty"`
	ContactName      *string    `db:"contact_name" json:"contact_name,omitempty"`
	ContactEmail     *string    `db:"contact_email" json:"contact_email,omitempty"`
	ContactPhone     *string    `db:"contact_phone" json:"contact_phone,omitempty"`
	Notes            *string    `db:"notes" json:"notes,omitempty"`
	Status           string     `db:"status" json:"status"`
	CompanyID        *int64     `db:"company_id" json:"company_id,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
	ArchivedAt       *time.Time `db:"archived_at" json:"archived_at,omitempty"`

	// Derived / Joined Fields
	AccountOwnerName   *string            `db:"account_owner_name" json:"account_owner_name,omitempty"`
	SecondaryOwnerName *string            `db:"secondary_owner_name" json:"secondary_owner_name,omitempty"`
	ActiveContracts    int                `db:"active_contracts" json:"active_contracts"`
	YTDRevenue         float64            `db:"ytd_revenue" json:"ytd_revenue"`
	LastActivity       *time.Time         `db:"last_activity" json:"last_activity,omitempty"`
	Contacts           []CustomerContact  `db:"-" json:"contacts,omitempty"`
	Addresses          []CustomerAddress  `db:"-" json:"addresses,omitempty"`
	LeadLinks          []CustomerLeadLink `db:"-" json:"lead_links,omitempty"`
}

// CustomerContact represents an individual contact under a customer account
type CustomerContact struct {
	ID          int64     `db:"id" json:"id"`
	OrgID       int64     `db:"org_id" json:"org_id"`
	CustomerID  int64     `db:"customer_id" json:"customer_id"`
	FirstName   string    `db:"first_name" json:"first_name"`
	LastName    string    `db:"last_name" json:"last_name"`
	Email       *string   `db:"email" json:"email,omitempty"`
	Phone       *string   `db:"phone" json:"phone,omitempty"`
	Mobile      *string   `db:"mobile" json:"mobile,omitempty"`
	JobTitle    *string   `db:"job_title" json:"job_title,omitempty"`
	Department  *string   `db:"department" json:"department,omitempty"`
	ContactRole string    `db:"contact_role" json:"contact_role"`
	IsPrimary   bool      `db:"is_primary" json:"is_primary"`
	Notes       *string   `db:"notes" json:"notes,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// CustomerAddress represents an operational location or address
type CustomerAddress struct {
	ID                int64     `db:"id" json:"id"`
	OrgID             int64     `db:"org_id" json:"org_id"`
	CustomerID        int64     `db:"customer_id" json:"customer_id"`
	AddressType       string    `db:"address_type" json:"address_type"`
	Label             *string   `db:"label" json:"label,omitempty"`
	AddressLine1      string    `db:"address_line_1" json:"address_line_1"`
	AddressLine2      *string   `db:"address_line_2" json:"address_line_2,omitempty"`
	City              string    `db:"city" json:"city"`
	State             *string   `db:"state" json:"state,omitempty"`
	PostalCode        *string   `db:"postal_code" json:"postal_code,omitempty"`
	CountryCode       string    `db:"country_code" json:"country_code"`
	Country           string    `db:"country" json:"country"`
	IsPrimaryBilling  bool      `db:"is_primary_billing" json:"is_primary_billing"`
	IsPrimaryShipping bool      `db:"is_primary_shipping" json:"is_primary_shipping"`
	ContactName       *string   `db:"contact_name" json:"contact_name,omitempty"`
	ContactPhone      *string   `db:"contact_phone" json:"contact_phone,omitempty"`
	ContactEmail      *string   `db:"contact_email" json:"contact_email,omitempty"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}

// CustomerLeadLink records the audit linkage between a original lead and resulting customer account
type CustomerLeadLink struct {
	ID                int64     `db:"id" json:"id"`
	OrgID             int64     `db:"org_id" json:"org_id"`
	CustomerID        int64     `db:"customer_id" json:"customer_id"`
	LeadID            int64     `db:"lead_id" json:"lead_id"`
	ConvertedByUserID *int64    `db:"converted_by_user_id" json:"converted_by_user_id,omitempty"`
	ConversionNotes   *string   `db:"conversion_notes" json:"conversion_notes,omitempty"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
}

// CustomerKPIs represents top summary aggregate indicators for Customer Directory header
type CustomerKPIs struct {
	TotalCustomers     int     `json:"total_customers"`
	ActiveCustomers    int     `json:"active_customers"`
	WithActiveContracts int    `json:"with_active_contracts"`
	TotalRevenueYTD    float64 `json:"total_revenue_ytd"`
	RequiringAttention int     `json:"requiring_attention"`
	TotalCustomersTrend float64 `json:"total_customers_trend"`
	ActiveCustomersTrend float64 `json:"active_customers_trend"`
	ContractsTrend     float64 `json:"contracts_trend"`
	RevenueTrend       float64 `json:"revenue_trend"`
	AttentionTrend     float64 `json:"attention_trend"`
}

// ListFilterParams holds query filters for Customer Directory
type ListFilterParams struct {
	Search         string `json:"search"`
	Status         string `json:"status"`
	CustomerType   string `json:"customer_type"`
	Country        string `json:"country"`
	AccountOwnerID int64  `json:"account_owner_id"`
	IncludeArchived bool   `json:"include_archived"`
	SortBy         string `json:"sort_by"`
	SortOrder      string `json:"sort_order"`
	Page           int    `json:"page"`
	Limit          int    `json:"limit"`
}

// CreateCustomerReq represents creation request payload
type CreateCustomerReq struct {
	Name           string  `json:"name"`
	TradingName    *string `json:"trading_name,omitempty"`
	CustomerType   string  `json:"customer_type"`
	Domain         *string `json:"domain,omitempty"`
	Industry       *string `json:"industry,omitempty"`
	TaxID          *string `json:"tax_id,omitempty"`
	PANNumber      *string `json:"pan_number,omitempty"`
	EORINumber     *string `json:"eori_number,omitempty"`
	Currency       string  `json:"currency"`
	PaymentTerms   string  `json:"payment_terms"`
	CreditLimit    float64 `json:"credit_limit"`
	AccountOwnerID *int64  `json:"account_owner_id,omitempty"`
	Website        *string `json:"website,omitempty"`
	Country        *string `json:"country,omitempty"`
	City           *string `json:"city,omitempty"`
	ContactName    *string `json:"contact_name,omitempty"`
	ContactEmail   *string `json:"contact_email,omitempty"`
	ContactPhone   *string `json:"contact_phone,omitempty"`
	Notes          *string `json:"notes,omitempty"`

	// Optional initial contact and address
	PrimaryContact *CreateContactReq `json:"primary_contact,omitempty"`
	BillingAddress *CreateAddressReq `json:"billing_address,omitempty"`
}

// UpdateCustomerReq represents update payload
type UpdateCustomerReq struct {
	Name           *string  `json:"name,omitempty"`
	TradingName    *string  `json:"trading_name,omitempty"`
	CustomerType   *string  `json:"customer_type,omitempty"`
	Domain         *string  `json:"domain,omitempty"`
	Industry       *string  `json:"industry,omitempty"`
	TaxID          *string  `json:"tax_id,omitempty"`
	PANNumber      *string  `json:"pan_number,omitempty"`
	EORINumber     *string  `json:"eori_number,omitempty"`
	Currency       *string  `json:"currency,omitempty"`
	PaymentTerms   *string  `json:"payment_terms,omitempty"`
	CreditLimit    *float64 `json:"credit_limit,omitempty"`
	HealthScore    *int     `json:"health_score,omitempty"`
	AccountOwnerID *int64   `json:"account_owner_id,omitempty"`
	Website        *string  `json:"website,omitempty"`
	Country        *string  `json:"country,omitempty"`
	City           *string  `json:"city,omitempty"`
	ContactName    *string  `json:"contact_name,omitempty"`
	ContactEmail   *string  `json:"contact_email,omitempty"`
	ContactPhone   *string  `json:"contact_phone,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
	Status         *string  `json:"status,omitempty"`
}

// CreateContactReq payload
type CreateContactReq struct {
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	Email       *string `json:"email,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	Mobile      *string `json:"mobile,omitempty"`
	JobTitle    *string `json:"job_title,omitempty"`
	Department  *string `json:"department,omitempty"`
	ContactRole string  `json:"contact_role,omitempty"`
	IsPrimary   bool    `json:"is_primary"`
	Notes       *string `json:"notes,omitempty"`
}

// CreateAddressReq payload
type CreateAddressReq struct {
	AddressType       string  `json:"address_type"`
	Label             *string `json:"label,omitempty"`
	AddressLine1      string  `json:"address_line_1"`
	AddressLine2      *string `json:"address_line_2,omitempty"`
	City              string  `json:"city"`
	State             *string `json:"state,omitempty"`
	PostalCode        *string `json:"postal_code,omitempty"`
	CountryCode       string  `json:"country_code"`
	Country           string  `json:"country"`
	IsPrimaryBilling  bool    `json:"is_primary_billing"`
	IsPrimaryShipping bool    `json:"is_primary_shipping"`
	ContactName       *string `json:"contact_name,omitempty"`
	ContactPhone      *string `json:"contact_phone,omitempty"`
	ContactEmail      *string `json:"contact_email,omitempty"`
}

// CheckDuplicateReq payload for evaluating potential customer duplicate candidates
type CheckDuplicateReq struct {
	Name    string  `json:"name"`
	Email   *string `json:"email,omitempty"`
	Phone   *string `json:"phone,omitempty"`
	Domain  *string `json:"domain,omitempty"`
	TaxID   *string `json:"tax_id,omitempty"`
	LeadID  *int64  `json:"lead_id,omitempty"`
}

// DuplicateMatchResult represents duplicate scoring evaluation output
type DuplicateMatchResult struct {
	CustomerID      int64   `json:"customer_id"`
	CustomerCode    string  `json:"customer_code"`
	CustomerName    string  `json:"customer_name"`
	ConfidenceScore int     `json:"confidence_score"` // 0-100
	MatchReason     string  `json:"match_reason"`     // e.g. "Tax ID exact match", "Domain & Name match"
	ExistingStatus  string  `json:"existing_status"`
	PrimaryContact  string  `json:"primary_contact"`
}

// ConvertLeadReq payload for deterministic Lead -> Customer conversion
type ConvertLeadReq struct {
	LeadID                 int64   `json:"lead_id"`
	LinkToExistingCustomerID *int64 `json:"link_to_existing_customer_id,omitempty"`
	ForceCreateNew         bool    `json:"force_create_new"`
	CustomerType           string  `json:"customer_type"`
	TaxID                  *string `json:"tax_id,omitempty"`
	PaymentTerms           string  `json:"payment_terms"`
	AccountOwnerID         *int64  `json:"account_owner_id,omitempty"`
	Notes                  *string `json:"notes,omitempty"`
}

// ConvertLeadResp output
type ConvertLeadResp struct {
	CustomerID   int64  `json:"customer_id"`
	CustomerCode string `json:"customer_code"`
	LeadID       int64  `json:"lead_id"`
	IsNewAccount bool   `json:"is_new_account"`
	Message      string `json:"message"`
}

// ── Customer 360 Degree Dashboard & Sub-Resource Types (Task 22) ───────────

type Customer360KPIs struct {
	TotalRFQs          int `json:"total_rfqs"`
	ActiveRFQs         int `json:"active_rfqs"`
	TotalQuotations    int `json:"total_quotations"`
	OpenQuotations     int `json:"open_quotations"`
	AcceptedQuotations int `json:"accepted_quotations"`
	ActiveBookings     int `json:"active_bookings"`
	ActiveShipments    int `json:"active_shipments"`
	LinkedContracts    int `json:"linked_contracts"`
}

type CustomerRFQ struct {
	ID              int64     `db:"id" json:"id"`
	RFQNumber       string    `db:"rfq_number" json:"rfq_number"`
	Status          string    `db:"status" json:"status"`
	Stage           string    `db:"stage" json:"stage"`
	Origin          string    `db:"origin" json:"origin"`
	Destination     string    `db:"destination" json:"destination"`
	ModeOfTransport string    `db:"mode_of_transport" json:"mode_of_transport"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

type CustomerQuotation struct {
	ID              int64      `db:"id" json:"id"`
	QuotationNumber string     `db:"quotation_number" json:"quotation_number"`
	Status          string     `db:"status" json:"status"`
	GrandTotal      float64    `db:"grand_total" json:"grand_total"`
	Currency        string     `db:"currency" json:"currency"`
	ValidUntil      *time.Time `db:"valid_until" json:"valid_until,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
}

type CustomerBooking struct {
	ID            int64     `db:"id" json:"id"`
	BookingNumber string    `db:"booking_number" json:"booking_number"`
	Status        string    `db:"status" json:"status"`
	CarrierName   string    `db:"carrier_name" json:"carrier_name"`
	OriginPort    string    `db:"origin_port" json:"origin_port"`
	DestinationPort string  `db:"destination_port" json:"destination_port"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type CustomerShipment struct {
	ID              int64     `db:"id" json:"id"`
	ShipmentNumber  string    `db:"shipment_number" json:"shipment_number"`
	Status          string    `db:"status" json:"status"`
	ModeOfTransport string    `db:"mode_of_transport" json:"mode_of_transport"`
	Origin          string    `db:"origin" json:"origin"`
	Destination     string    `db:"destination" json:"destination"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

type CustomerContract struct {
	ID                int64      `db:"id" json:"id"`
	ContractReference string     `db:"contract_reference" json:"contract_reference"`
	ContractName      string     `db:"contract_name" json:"contract_name"`
	ContractType      string     `db:"contract_type" json:"contract_type"`
	Status            string     `db:"status" json:"status"`
	ValidFrom         *time.Time `db:"valid_from" json:"valid_from,omitempty"`
	ValidTo           *time.Time `db:"valid_to" json:"valid_to,omitempty"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
}

type CustomerTimelineEvent struct {
	ID                string    `json:"id"`
	EventType         string    `json:"event_type"` // e.g. "CUSTOMER_CREATED", "LEAD_CONVERTED", "RFQ_CREATED", "QUOTATION_CREATED", "BOOKING_CREATED", "SHIPMENT_CREATED", "CONTRACT_LINKED"
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	RelatedRecordType string    `json:"related_record_type,omitempty"` // "RFQ", "QUOTATION", "BOOKING", "SHIPMENT", "CONTRACT", "LEAD"
	RelatedRecordID   int64     `json:"related_record_id,omitempty"`
	RelatedRecordCode string    `json:"related_record_code,omitempty"`
	Timestamp         time.Time `json:"timestamp"`
	ActorUser         string    `json:"actor_user,omitempty"`
}

type Customer360Dashboard struct {
	Customer         Customer                `json:"customer"`
	KPIs             Customer360KPIs         `json:"kpis"`
	RecentRFQs       []CustomerRFQ           `json:"recent_rfqs"`
	RecentQuotations []CustomerQuotation     `json:"recent_quotations"`
	RecentBookings   []CustomerBooking       `json:"recent_bookings"`
	RecentShipments  []CustomerShipment      `json:"recent_shipments"`
	RecentContracts  []CustomerContract      `json:"recent_contracts"`
	Timeline         []CustomerTimelineEvent `json:"timeline"`
}

// ── Customer Financial & Relationship Management Types (Task 23) ────────────

type CustomerFinancialProfile struct {
	CustomerID      int64     `json:"customer_id"`
	Currency        string    `json:"currency"`
	PaymentTerms    string    `json:"payment_terms"`
	CreditLimit     float64   `json:"credit_limit"`
	CreditStatus    string    `json:"credit_status"`
	CommercialNotes *string   `json:"commercial_notes,omitempty"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type UpdateFinancialProfileReq struct {
	Currency        *string  `json:"currency,omitempty"`
	PaymentTerms    *string  `json:"payment_terms,omitempty"`
	CreditLimit     *float64 `json:"credit_limit,omitempty"`
	CreditStatus    *string  `json:"credit_status,omitempty"`
	CommercialNotes *string  `json:"commercial_notes,omitempty"`
}

type CustomerCommercialMetrics struct {
	TotalQuotationValue    float64 `json:"total_quotation_value"`
	OpenQuotationValue     float64 `json:"open_quotation_value"`
	AcceptedQuotationValue float64 `json:"accepted_quotation_value"`
	ExpiredQuotationValue  float64 `json:"expired_quotation_value"`
	ActiveContractValue    float64 `json:"active_contract_value"`
	ExpiringContractValue  float64 `json:"expiring_contract_value"`
	TotalRFQs              int     `json:"total_rfqs"`
	TotalQuotations        int     `json:"total_quotations"`
	AcceptedQuotations     int     `json:"accepted_quotations"`
	QuoteConversionRate    float64 `json:"quote_conversion_rate"`
}

type CustomerAccountOwnership struct {
	CustomerID         int64   `json:"customer_id"`
	PrimaryOwnerID     *int64  `json:"primary_owner_id,omitempty"`
	PrimaryOwnerName   *string `json:"primary_owner_name,omitempty"`
	SecondaryOwnerID   *int64  `json:"secondary_owner_id,omitempty"`
	SecondaryOwnerName *string `json:"secondary_owner_name,omitempty"`
}

type UpdateOwnershipReq struct {
	PrimaryOwnerID   *int64  `json:"primary_owner_id,omitempty"`
	SecondaryOwnerID *int64  `json:"secondary_owner_id,omitempty"`
	ChangeReason     *string `json:"change_reason,omitempty"`
}

type CustomerOwnershipHistoryItem struct {
	ID                int64     `db:"id" json:"id"`
	OrgID             int64     `db:"org_id" json:"org_id"`
	CustomerID        int64     `db:"customer_id" json:"customer_id"`
	PreviousOwnerID   *int64    `db:"previous_owner_id" json:"previous_owner_id,omitempty"`
	PreviousOwnerName *string   `db:"previous_owner_name" json:"previous_owner_name,omitempty"`
	NewOwnerID        int64     `db:"new_owner_id" json:"new_owner_id"`
	NewOwnerName      string    `db:"new_owner_name" json:"new_owner_name"`
	OwnershipType     string    `db:"ownership_type" json:"ownership_type"`
	ChangedByUserID   *int64    `db:"changed_by_user_id" json:"changed_by_user_id,omitempty"`
	ChangedByUserName *string   `db:"changed_by_user_name" json:"changed_by_user_name,omitempty"`
	ChangeReason      *string   `db:"change_reason" json:"change_reason,omitempty"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
}

type CustomerRelationshipSummary struct {
	PrimaryContact *CustomerContact             `json:"primary_contact,omitempty"`
	ContactsByRole map[string][]CustomerContact `json:"contacts_by_role"`
}

// ── Customer Intelligence, Health & Risk Management Types (Task 24) ─────────

type CustomerHealthEvaluation struct {
	ID                  int64     `db:"id" json:"id"`
	OrgID               int64     `db:"org_id" json:"org_id"`
	CustomerID          int64     `db:"customer_id" json:"customer_id"`
	HealthStatus        string    `db:"health_status" json:"health_status"`
	HealthScore         int       `db:"health_score" json:"health_score"`
	ContributingFactors []string  `json:"contributing_factors"`
	EvaluatedAt         time.Time `db:"evaluated_at" json:"evaluated_at"`
}

type CustomerRiskEvent struct {
	ID             int64      `db:"id" json:"id"`
	OrgID          int64      `db:"org_id" json:"org_id"`
	CustomerID     int64      `db:"customer_id" json:"customer_id"`
	CustomerName   string     `db:"customer_name" json:"customer_name"`
	CustomerCode   string     `db:"customer_code" json:"customer_code"`
	RiskType       string     `db:"risk_type" json:"risk_type"`
	Severity       string     `db:"severity" json:"severity"`
	Title          string     `db:"title" json:"title"`
	Description    string     `db:"description" json:"description"`
	DetectedAt     time.Time  `db:"detected_at" json:"detected_at"`
	IsResolved     bool       `db:"is_resolved" json:"is_resolved"`
	ResolvedAt     *time.Time `db:"resolved_at" json:"resolved_at,omitempty"`
	ResolvedBy     *int64     `db:"resolved_by" json:"resolved_by,omitempty"`
	ResolvedByName *string    `db:"resolved_by_name" json:"resolved_by_name,omitempty"`
	ResolutionNote *string    `db:"resolution_note" json:"resolution_note,omitempty"`
}

type CustomerOpportunityEvent struct {
	ID                int64     `db:"id" json:"id"`
	OrgID             int64     `db:"org_id" json:"org_id"`
	CustomerID        int64     `db:"customer_id" json:"customer_id"`
	CustomerName      string    `db:"customer_name" json:"customer_name"`
	CustomerCode      string    `db:"customer_code" json:"customer_code"`
	OpportunityType   string    `db:"opportunity_type" json:"opportunity_type"`
	Priority          string    `db:"priority" json:"priority"`
	Title             string    `db:"title" json:"title"`
	Reason            string    `db:"reason" json:"reason"`
	SuggestedAction   string    `db:"suggested_action" json:"suggested_action"`
	RelatedRecordCode *string   `db:"related_record_code" json:"related_record_code,omitempty"`
	DetectedAt        time.Time `db:"detected_at" json:"detected_at"`
}

type CustomerAttentionItem struct {
	CustomerID       int64     `db:"customer_id" json:"customer_id"`
	CustomerName     string    `db:"customer_name" json:"customer_name"`
	CustomerCode     string    `db:"customer_code" json:"customer_code"`
	HealthStatus     string    `db:"health_status" json:"health_status"`
	HealthScore      int       `db:"health_score" json:"health_score"`
	Severity         string    `db:"severity" json:"severity"`
	Title            string    `db:"title" json:"title"`
	Reason           string    `db:"reason" json:"reason"`
	SuggestedAction  string    `db:"suggested_action" json:"suggested_action"`
	AccountOwnerName *string   `db:"account_owner_name" json:"account_owner_name,omitempty"`
	DetectedAt       time.Time `db:"detected_at" json:"detected_at"`
}

type CustomerIntelligenceSummary struct {
	HealthyCount          int `json:"healthy_count"`
	WatchCount            int `json:"watch_count"`
	AtRiskCount           int `json:"at_risk_count"`
	CriticalCount         int `json:"critical_count"`
	InsufficientDataCount int `json:"insufficient_data_count"`
	TotalRisks            int `json:"total_risks"`
	TotalOpportunities    int `json:"total_opportunities"`
	TotalAttentionItems   int `json:"total_attention_items"`
}

type CustomerIntelligenceProfile struct {
	Customer             Customer                   `json:"customer"`
	Health               CustomerHealthEvaluation   `json:"health"`
	OpenRisks            []CustomerRiskEvent        `json:"open_risks"`
	DetectedOpportunities []CustomerOpportunityEvent `json:"detected_opportunities"`
	ActivityTrend        string                     `json:"activity_trend"`
}

type ResolveRiskReq struct {
	ResolutionNote string `json:"resolution_note"`
}



