package subscription

import (
	"encoding/json"
	"time"
)

const (
	MetricTeamMembers       = "team_members"
	MetricAIEmails          = "ai_email_processing"
	MetricRFQs              = "rfqs"
	MetricShipments         = "shipments"
	MetricCarrierConnection = "carrier_connections"
	MetricStorage           = "storage_gb"
)

type SubscriptionPlan struct {
	ID                int64           `json:"id" db:"id"`
	Name              string          `json:"name" db:"name"`
	Description       string          `json:"description" db:"description"`
	PriceMonthly      float64         `json:"price_monthly" db:"price_monthly"`
	PriceAnnual       float64         `json:"price_annual" db:"price_annual"`
	Features          json.RawMessage `json:"features" db:"features"`
	Limits                  json.RawMessage `json:"limits" db:"limits"`
	ProviderProductID       *string         `json:"provider_product_id" db:"provider_product_id"`
	ProviderPriceIDMonthly  *string         `json:"provider_price_id_monthly" db:"provider_price_id_monthly"`
	ProviderPriceIDAnnual   *string         `json:"provider_price_id_annual" db:"provider_price_id_annual"`
	IsActive                bool            `json:"is_active" db:"is_active"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
}

type OrganizationSubscription struct {
	ID                     int64      `json:"id" db:"id"`
	OrgID                  int64      `json:"org_id" db:"org_id"`
	PlanID                 int64      `json:"plan_id" db:"plan_id"`
	Status                 string     `json:"status" db:"status"`
	BillingCycle           string     `json:"billing_cycle" db:"billing_cycle"`
	CurrentPeriodStart     *time.Time `json:"current_period_start" db:"current_period_start"`
	CurrentPeriodEnd       *time.Time `json:"current_period_end" db:"current_period_end"`
	CancelAtPeriodEnd      bool       `json:"cancel_at_period_end" db:"cancel_at_period_end"`
	ProviderSubscriptionID *string    `json:"provider_subscription_id" db:"provider_subscription_id"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
}

type BillingCustomer struct {
	ID                 int64           `json:"id" db:"id"`
	OrgID              int64           `json:"org_id" db:"org_id"`
	ProviderCustomerID string          `json:"provider_customer_id" db:"provider_customer_id"`
	BillingEmail       *string         `json:"billing_email" db:"billing_email"`
	TaxID              *string         `json:"tax_id" db:"tax_id"`
	BillingAddress     json.RawMessage `json:"billing_address" db:"billing_address"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at" db:"updated_at"`
}

type BillingPaymentMethod struct {
	ID                      int64     `json:"id" db:"id"`
	OrgID                   int64     `json:"org_id" db:"org_id"`
	ProviderPaymentMethodID string    `json:"provider_payment_method_id" db:"provider_payment_method_id"`
	CardBrand               *string   `json:"card_brand" db:"card_brand"`
	CardLast4               *string   `json:"card_last4" db:"card_last4"`
	ExpMonth                *int      `json:"exp_month" db:"exp_month"`
	ExpYear                 *int      `json:"exp_year" db:"exp_year"`
	IsDefault               bool      `json:"is_default" db:"is_default"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time `json:"updated_at" db:"updated_at"`
}

type SubscriptionUsage struct {
	ID           int64      `json:"id" db:"id"`
	OrgID        int64      `json:"org_id" db:"org_id"`
	MetricName   string     `json:"metric_name" db:"metric_name"`
	CurrentUsage int        `json:"current_usage" db:"current_usage"`
	LimitAmount  *int       `json:"limit_amount" db:"limit_amount"`
	PeriodStart  *time.Time `json:"period_start" db:"period_start"`
	PeriodEnd    *time.Time `json:"period_end" db:"period_end"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`

	// Enriched fields for the frontend
	Remaining  int  `json:"remaining" db:"-"`
	Percentage int  `json:"percentage" db:"-"`
	Unlimited  bool `json:"unlimited" db:"-"`
}

type Invoice struct {
	ID                int64      `json:"id" db:"id"`
	OrgID             int64      `json:"org_id" db:"org_id"`
	ProviderInvoiceID *string    `json:"provider_invoice_id" db:"provider_invoice_id"`
	Number            *string    `json:"number" db:"number"`
	AmountDue         float64    `json:"amount_due" db:"amount_due"`
	AmountPaid        float64    `json:"amount_paid" db:"amount_paid"`
	Status            string     `json:"status" db:"status"`
	InvoicePdfUrl     *string    `json:"invoice_pdf_url" db:"invoice_pdf_url"`
	IssuedAt          *time.Time `json:"issued_at" db:"issued_at"`
	PaidAt            *time.Time `json:"paid_at" db:"paid_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

type SubscriptionAddon struct {
	ID                         int64     `json:"id" db:"id"`
	OrgID                      int64     `json:"org_id" db:"org_id"`
	AddonConfigID              *int64    `json:"addon_config_id" db:"addon_config_id"`
	AddonName                  string    `json:"addon_name" db:"addon_name"`
	Quantity                   int       `json:"quantity" db:"quantity"`
	PricePerUnit               float64   `json:"price_per_unit" db:"price_per_unit"`
	ProviderSubscriptionItemID *string   `json:"provider_subscription_item_id" db:"provider_subscription_item_id"`
	CreatedAt                  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at" db:"updated_at"`
}

type AddonConfig struct {
	ID                int64           `json:"id" db:"id"`
	Name              string          `json:"name" db:"name"`
	Description       string          `json:"description" db:"description"`
	PricingModel      string          `json:"pricing_model" db:"pricing_model"` // e.g. "per_unit", "flat_fee"
	UnitPrice         float64         `json:"unit_price" db:"unit_price"`
	IsRecurring       bool            `json:"is_recurring" db:"is_recurring"`
	AvailableForPlans json.RawMessage `json:"available_for_plans" db:"available_for_plans"`
	ProviderProductID *string         `json:"provider_product_id" db:"provider_product_id"`
	ProviderPriceID   *string         `json:"provider_price_id" db:"provider_price_id"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
}

// API Requests / Responses

type ChangePlanRequest struct {
	PlanID       int64  `json:"plan_id"`
	BillingCycle string `json:"billing_cycle"` // "monthly" or "annual"
}

type PlanChangePreviewResponse struct {
	CurrentPlanName     string  `json:"current_plan_name"`
	NewPlanName         string  `json:"new_plan_name"`
	BillingCycle        string  `json:"billing_cycle"`
	NewPrice            float64 `json:"new_price"`
	ProratedCharge      float64 `json:"prorated_charge"` // Positive means charge, negative means credit
	EffectiveDate       string  `json:"effective_date"`
}

type UpdateAddonsRequest struct {
	AddonConfigID int64 `json:"addon_config_id"`
	Quantity      int   `json:"quantity"` // 0 means disable/remove
}

type SubscriptionWorkspaceResponse struct {
	Subscription  *OrganizationSubscription `json:"subscription"`
	CurrentPlan   *SubscriptionPlan         `json:"current_plan"`
	Customer      *BillingCustomer          `json:"customer"`
	PaymentMethod *BillingPaymentMethod     `json:"payment_method"`
	Usage         []SubscriptionUsage       `json:"usage"`
	Addons        []SubscriptionAddon       `json:"addons"`
	Invoices      []Invoice                 `json:"invoices"`
}
