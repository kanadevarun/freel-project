package quotations

import (
	"encoding/json"
	"time"
)

// Quotation is the core domain entity — a formal price offer to a customer.
type Quotation struct {
	ID     int64  `db:"id"     json:"id"`
	OrgID  int64  `db:"org_id" json:"org_id"`

	QuotationNumber string `db:"quotation_number" json:"quotation_number"`

	CustomerID   *int64  `db:"customer_id"   json:"customer_id,omitempty"`
	CustomerName string  `db:"customer_name" json:"customer_name"`
	RFQID        *int64  `db:"rfq_id"        json:"rfq_id,omitempty"`
	RFQNumber    string  `db:"rfq_number"    json:"rfq_number"`

	Status string `db:"status" json:"status"`

	Origin          string `db:"origin"           json:"origin"`
	OriginCode      string `db:"origin_code"      json:"origin_code"`
	Destination     string `db:"destination"      json:"destination"`
	DestinationCode string `db:"destination_code" json:"destination_code"`

	ServiceType   string `db:"service_type"   json:"service_type"`
	TransportMode string `db:"transport_mode" json:"transport_mode"`

	Currency       string  `db:"currency"         json:"currency"`
	PaymentTerms   string  `db:"payment_terms"    json:"payment_terms"`
	Subtotal       float64 `db:"subtotal"         json:"subtotal"`
	Surcharges     float64 `db:"surcharges"       json:"surcharges"`
	Taxes          float64 `db:"taxes"            json:"taxes"`
	TotalAmount    float64 `db:"total_amount"     json:"total_amount"`
	TotalCost      float64 `db:"total_cost"       json:"total_cost"`
	GrossProfit    float64 `db:"gross_profit"     json:"gross_profit"`
	GrossMarginPct float64 `db:"gross_margin_pct" json:"gross_margin_pct"`

	ValidFrom  *time.Time `db:"valid_from"  json:"valid_from,omitempty"`
	ValidUntil *time.Time `db:"valid_until" json:"valid_until,omitempty"`

	SubmittedForReviewAt   *time.Time `db:"submitted_for_review_at"   json:"submitted_for_review_at,omitempty"`
	SubmittedForReviewBy   string     `db:"submitted_for_review_by"   json:"submitted_for_review_by,omitempty"`
	ApprovedAt             *time.Time `db:"approved_at"               json:"approved_at,omitempty"`
	ApprovedBy             string     `db:"approved_by"               json:"approved_by,omitempty"`
	ApprovalNotes          string     `db:"approval_notes"            json:"approval_notes,omitempty"`
	ChangesRequestedAt     *time.Time `db:"changes_requested_at"     json:"changes_requested_at,omitempty"`
	ChangesRequestedBy     string     `db:"changes_requested_by"     json:"changes_requested_by,omitempty"`
	ChangesRequestedReason string     `db:"changes_requested_reason" json:"changes_requested_reason,omitempty"`

	SentAt         *time.Time `db:"sent_at"         json:"sent_at,omitempty"`
	SentBy         string     `db:"sent_by"         json:"sent_by,omitempty"`
	ViewedAt       *time.Time `db:"viewed_at"       json:"viewed_at,omitempty"`
	FirstViewedAt  *time.Time `db:"first_viewed_at" json:"first_viewed_at,omitempty"`
	LastViewedAt   *time.Time `db:"last_viewed_at"  json:"last_viewed_at,omitempty"`
	ViewCount      int        `db:"view_count"      json:"view_count"`
	AcceptedAt     *time.Time `db:"accepted_at"     json:"accepted_at,omitempty"`
	DeclinedAt     *time.Time `db:"declined_at"     json:"declined_at,omitempty"`
	DeclinedReason string     `db:"declined_reason" json:"declined_reason,omitempty"`
	RejectedAt     *time.Time `db:"rejected_at"     json:"rejected_at,omitempty"`
	ExpiredAt      *time.Time `db:"expired_at"      json:"expired_at,omitempty"`
	CancelledAt    *time.Time `db:"cancelled_at"    json:"cancelled_at,omitempty"`
	CancelledBy    string     `db:"cancelled_by"    json:"cancelled_by,omitempty"`
	CancelledReason string    `db:"cancelled_reason" json:"cancelled_reason,omitempty"`

	// Conversion tracking fields (Task 18.6)
	ConvertedAt         *time.Time `db:"converted_at"          json:"converted_at,omitempty"`
	ConvertedBy         string     `db:"converted_by"          json:"converted_by,omitempty"`
	ConvertedBookingID  *int64     `db:"converted_booking_id"  json:"converted_booking_id,omitempty"`
	ConvertedShipmentID *int64     `db:"converted_shipment_id" json:"converted_shipment_id,omitempty"`
	ConversionStatus    string     `db:"conversion_status"     json:"conversion_status"`
	ConversionNotes     string     `db:"conversion_notes"      json:"conversion_notes,omitempty"`

	Notes           string `db:"notes"            json:"notes"`
	CommercialTerms string `db:"commercial_terms" json:"commercial_terms"`
	CustomerNotes   string `db:"customer_notes"   json:"customer_notes"`
	InternalNotes   string `db:"internal_notes"   json:"internal_notes"`
	TemplateID      *int64 `db:"template_id"      json:"template_id,omitempty"`

	ValidityStatus string `json:"validity_status"`
	CanEdit        bool   `json:"can_edit"`
	CanConvert     bool   `json:"can_convert"`

	CreatedBy string `db:"created_by" json:"created_by"`
	UpdatedBy string `db:"updated_by" json:"updated_by"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// QuotationListItem is a lightweight projection for the quotation table.
type QuotationListItem struct {
	ID              int64      `db:"id"               json:"id"`
	QuotationNumber string     `db:"quotation_number" json:"quotation_number"`
	CustomerID      *int64     `db:"customer_id"      json:"customer_id,omitempty"`
	CustomerName    string     `db:"customer_name"    json:"customer_name"`
	Origin          string     `db:"origin"           json:"origin"`
	OriginCode      string     `db:"origin_code"      json:"origin_code"`
	Destination     string     `db:"destination"      json:"destination"`
	DestinationCode string     `db:"destination_code" json:"destination_code"`
	ServiceType     string     `db:"service_type"     json:"service_type"`
	TransportMode   string     `db:"transport_mode"   json:"transport_mode"`
	Currency        string     `db:"currency"         json:"currency"`
	PaymentTerms    string     `db:"payment_terms"    json:"payment_terms"`
	TotalAmount     float64    `db:"total_amount"     json:"total_amount"`
	TotalCost       float64    `db:"total_cost"       json:"total_cost"`
	GrossProfit     float64    `db:"gross_profit"     json:"gross_profit"`
	GrossMarginPct  float64    `db:"gross_margin_pct" json:"gross_margin_pct"`
	Status          string     `db:"status"           json:"status"`
	ConversionStatus string    `db:"conversion_status" json:"conversion_status"`
	ConvertedBookingID *int64  `db:"converted_booking_id" json:"converted_booking_id,omitempty"`
	ConvertedShipmentID *int64 `db:"converted_shipment_id" json:"converted_shipment_id,omitempty"`
	ValidityStatus  string     `json:"validity_status"`
	ValidFrom       *time.Time `db:"valid_from"       json:"valid_from,omitempty"`
	ValidUntil      *time.Time `db:"valid_until"      json:"valid_until,omitempty"`
	UpdatedAt       time.Time  `db:"updated_at"       json:"updated_at"`
}

// QuotationChargeItem represents a single commercial line item belonging to a quotation.
type QuotationChargeItem struct {
	ID               int64     `db:"id"                json:"id"`
	OrgID            int64     `db:"org_id"            json:"org_id"`
	QuotationID      int64     `db:"quotation_id"      json:"quotation_id"`
	ChargeCode       string    `db:"charge_code"       json:"charge_code"`
	ChargeName       string    `db:"charge_name"       json:"charge_name"`
	ChargeCategory   string    `db:"charge_category"   json:"charge_category"`
	ChargeType       string    `db:"charge_type"       json:"charge_type"`
	CalculationBasis string    `db:"calculation_basis" json:"calculation_basis"`
	Quantity         float64   `db:"quantity"          json:"quantity"`
	UnitPrice        float64   `db:"unit_price"        json:"unit_price"`
	CostAmount       float64   `db:"cost_amount"       json:"cost_amount"`
	SellAmount       float64   `db:"sell_amount"       json:"sell_amount"`
	Currency         string    `db:"currency"          json:"currency"`
	ExchangeRate     float64   `db:"exchange_rate"     json:"exchange_rate"`
	TaxRate          float64   `db:"tax_rate"          json:"tax_rate"`
	TaxAmount        float64   `db:"tax_amount"        json:"tax_amount"`
	DiscountType     string    `db:"discount_type"     json:"discount_type"`
	DiscountValue    float64   `db:"discount_value"    json:"discount_value"`
	DiscountAmount   float64   `db:"discount_amount"   json:"discount_amount"`
	TotalCost        float64   `db:"total_cost"        json:"total_cost"`
	TotalSell        float64   `db:"total_sell"        json:"total_sell"`
	DisplayOrder     int       `db:"display_order"     json:"display_order"`
	IsOptional       bool      `db:"is_optional"       json:"is_optional"`
	Notes            string    `db:"notes"             json:"notes"`
	CreatedBy        string    `db:"created_by"        json:"created_by"`
	UpdatedBy        string    `db:"updated_by"        json:"updated_by"`
	CreatedAt        time.Time `db:"created_at"        json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"        json:"updated_at"`
}

// QuotationPricingSummary represents the calculated commercial totals for a quotation.
type QuotationPricingSummary struct {
	Currency              string  `json:"currency"`
	FreightAmount         float64 `json:"freight_amount"`
	OriginCharges         float64 `json:"origin_charges"`
	DestinationCharges    float64 `json:"destination_charges"`
	Surcharges            float64 `json:"surcharges"`
	DocumentationCharges  float64 `json:"documentation_charges"`
	CustomsCharges        float64 `json:"customs_charges"`
	InsuranceCharges      float64 `json:"insurance_charges"`
	OtherCharges          float64 `json:"other_charges"`
	Taxes                 float64 `json:"taxes"`
	Discounts             float64 `json:"discounts"`
	Subtotal              float64 `json:"subtotal"`
	TotalAmount           float64 `json:"total_amount"`
	TotalCost             float64 `json:"total_cost"`
	GrossProfit           float64 `json:"gross_profit"`
	GrossMarginPercentage float64 `json:"gross_margin_percentage"`
	MarginHealth          string  `json:"margin_health"`
	MultiCurrencyWarning  bool    `json:"multi_currency_warning"`
	ItemCount             int     `json:"item_count"`
}

// QuotationPricing wraps the full pricing response consumed by the quotation detail UI.
type QuotationPricing struct {
	QuotationID int64                    `json:"quotation_id"`
	Currency    string                   `json:"currency"`
	ChargeItems []*QuotationChargeItem   `json:"charge_items"`
	Summary     *QuotationPricingSummary `json:"summary"`
}

// ── Reusable Quotation Templates (Task 18.3) ──────────────────────────────────

// QuotationTemplate represents a reusable blueprint for creating quotations.
type QuotationTemplate struct {
	ID              int64     `db:"id"               json:"id"`
	OrgID           int64     `db:"org_id"           json:"org_id"`
	Name            string    `db:"name"             json:"name"`
	Description     string    `db:"description"      json:"description"`
	ShipmentMode    string    `db:"shipment_mode"    json:"shipment_mode"`
	TransportMode   string    `db:"transport_mode"   json:"transport_mode"`
	Origin          string    `db:"origin"           json:"origin"`
	Destination     string    `db:"destination"      json:"destination"`
	Currency        string    `db:"currency"         json:"currency"`
	ValidityDays    int       `db:"validity_days"    json:"validity_days"`
	PaymentTerms    string    `db:"payment_terms"    json:"payment_terms"`
	CommercialTerms string    `db:"commercial_terms" json:"commercial_terms"`
	CustomerNotes   string    `db:"customer_notes"   json:"customer_notes"`
	InternalNotes   string    `db:"internal_notes"   json:"internal_notes"`
	IsActive        bool      `db:"is_active"        json:"is_active"`
	CreatedBy       string    `db:"created_by"       json:"created_by"`
	CreatedAt       time.Time `db:"created_at"       json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"       json:"updated_at"`

	// ChargeCount holds the count of reusable charge items for list projections
	ChargeCount int `db:"charge_count" json:"charge_count"`
}

// QuotationTemplateChargeItem represents a pre-configured reusable charge definition.
type QuotationTemplateChargeItem struct {
	ID               int64     `db:"id"                json:"id"`
	OrgID            int64     `db:"org_id"            json:"org_id"`
	TemplateID       int64     `db:"template_id"       json:"template_id"`
	ChargeCategory   string    `db:"charge_category"   json:"charge_category"`
	ChargeCode       string    `db:"charge_code"       json:"charge_code"`
	ChargeName       string    `db:"charge_name"       json:"charge_name"`
	CalculationBasis string    `db:"calculation_basis" json:"calculation_basis"`
	Quantity         float64   `db:"quantity"          json:"quantity"`
	UnitPrice        float64   `db:"unit_price"        json:"unit_price"`
	CostAmount       float64   `db:"cost_amount"       json:"cost_amount"`
	DiscountType     string    `db:"discount_type"     json:"discount_type"`
	DiscountValue    float64   `db:"discount_value"    json:"discount_value"`
	TaxRate          float64   `db:"tax_rate"          json:"tax_rate"`
	Currency         string    `db:"currency"          json:"currency"`
	DisplayOrder     int       `db:"display_order"     json:"display_order"`
	IsOptional       bool      `db:"is_optional"       json:"is_optional"`
	Notes            string    `db:"notes"             json:"notes"`
	CreatedAt        time.Time `db:"created_at"        json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"        json:"updated_at"`
}

// QuotationTemplateDetail wraps a template with its charge definitions.
type QuotationTemplateDetail struct {
	Template *QuotationTemplate             `json:"template"`
	Charges  []*QuotationTemplateChargeItem `json:"charges"`
}

// QuotationCommercialTerms represents commercial conditions for a quotation.
type QuotationCommercialTerms struct {
	ValidFrom       *string `json:"valid_from,omitempty"`
	ValidUntil      *string `json:"valid_until,omitempty"`
	PaymentTerms    string  `json:"payment_terms"`
	CommercialTerms string  `json:"commercial_terms"`
	CustomerNotes   string  `json:"customer_notes"`
	InternalNotes   string  `json:"internal_notes"`
	ValidityStatus  string  `json:"validity_status"`
}

// ── Requests & Responses ──────────────────────────────────────────────────────

// CreateQuotationTemplateRequest is payload to create a new template.
type CreateQuotationTemplateRequest struct {
	Name            string                               `json:"name"`
	Description     string                               `json:"description"`
	ShipmentMode    string                               `json:"shipment_mode"`
	TransportMode   string                               `json:"transport_mode"`
	Origin          string                               `json:"origin"`
	Destination     string                               `json:"destination"`
	Currency        string                               `json:"currency"`
	ValidityDays    int                                  `json:"validity_days"`
	PaymentTerms    string                               `json:"payment_terms"`
	CommercialTerms string                               `json:"commercial_terms"`
	CustomerNotes   string                               `json:"customer_notes"`
	InternalNotes   string                               `json:"internal_notes"`
	Charges         []*CreateTemplateChargeItemRequest  `json:"charges,omitempty"`
}

// CreateTemplateChargeItemRequest defines a charge inside a template create/update request.
type CreateTemplateChargeItemRequest struct {
	ChargeCategory   string  `json:"charge_category"`
	ChargeCode       string  `json:"charge_code"`
	ChargeName       string  `json:"charge_name"`
	CalculationBasis string  `json:"calculation_basis"`
	Quantity         float64 `json:"quantity"`
	UnitPrice        float64 `json:"unit_price"`
	CostAmount       float64 `json:"cost_amount"`
	DiscountType     string  `json:"discount_type"`
	DiscountValue    float64 `json:"discount_value"`
	TaxRate          float64 `json:"tax_rate"`
	Currency         string  `json:"currency"`
	DisplayOrder     int     `json:"display_order"`
	IsOptional       bool    `json:"is_optional"`
	Notes            string  `json:"notes"`
}

// UpdateQuotationTemplateRequest is payload to update an existing template.
type UpdateQuotationTemplateRequest struct {
	Name            *string                              `json:"name,omitempty"`
	Description     *string                              `json:"description,omitempty"`
	ShipmentMode    *string                              `json:"shipment_mode,omitempty"`
	TransportMode   *string                              `json:"transport_mode,omitempty"`
	Origin          *string                              `json:"origin,omitempty"`
	Destination     *string                              `json:"destination,omitempty"`
	Currency        *string                              `json:"currency,omitempty"`
	ValidityDays    *int                                 `json:"validity_days,omitempty"`
	PaymentTerms    *string                              `json:"payment_terms,omitempty"`
	CommercialTerms *string                              `json:"commercial_terms,omitempty"`
	CustomerNotes   *string                              `json:"customer_notes,omitempty"`
	InternalNotes   *string                              `json:"internal_notes,omitempty"`
	IsActive        *bool                                `json:"is_active,omitempty"`
	Charges         []*CreateTemplateChargeItemRequest  `json:"charges,omitempty"`
}

// ApplyQuotationTemplateRequest is payload to apply a template onto a quotation.
type ApplyQuotationTemplateRequest struct {
	TemplateID      int64 `json:"template_id"`
	OverrideCharges bool  `json:"override_charges"` // true = replace existing charges; false = append
}

// CreateTemplateFromQuotationRequest is payload to save an existing quotation as a template.
type CreateTemplateFromQuotationRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	ValidityDays int    `json:"validity_days"`
}

// UpdateQuotationCommercialTermsRequest is payload to update quotation terms and notes.
type UpdateQuotationCommercialTermsRequest struct {
	ValidFrom       *string `json:"valid_from,omitempty"`
	ValidUntil      *string `json:"valid_until,omitempty"`
	PaymentTerms    *string `json:"payment_terms,omitempty"`
	CommercialTerms *string `json:"commercial_terms,omitempty"`
	CustomerNotes   *string `json:"customer_notes,omitempty"`
	InternalNotes   *string `json:"internal_notes,omitempty"`
}

// CreateQuotationChargeRequest is the payload to add a new line item charge.
type CreateQuotationChargeRequest struct {
	ChargeCode       string   `json:"charge_code"`
	ChargeName       string   `json:"charge_name"`
	ChargeCategory   string   `json:"charge_category"`
	ChargeType       string   `json:"charge_type"`
	CalculationBasis string   `json:"calculation_basis"`
	Quantity         float64  `json:"quantity"`
	UnitPrice        float64  `json:"unit_price"`
	CostAmount       float64  `json:"cost_amount"`
	Currency         string   `json:"currency"`
	TaxRate          float64  `json:"tax_rate"`
	DiscountType     string   `json:"discount_type"`
	DiscountValue    float64  `json:"discount_value"`
	DisplayOrder     *int     `json:"display_order,omitempty"`
	IsOptional       bool     `json:"is_optional"`
	Notes            string   `json:"notes"`
}

// UpdateQuotationChargeRequest is the payload to update an existing line item charge.
type UpdateQuotationChargeRequest struct {
	ChargeCode       *string  `json:"charge_code,omitempty"`
	ChargeName       *string  `json:"charge_name,omitempty"`
	ChargeCategory   *string  `json:"charge_category,omitempty"`
	ChargeType       *string  `json:"charge_type,omitempty"`
	CalculationBasis *string  `json:"calculation_basis,omitempty"`
	Quantity         *float64 `json:"quantity,omitempty"`
	UnitPrice        *float64 `json:"unit_price,omitempty"`
	CostAmount       *float64 `json:"cost_amount,omitempty"`
	Currency         *string  `json:"currency,omitempty"`
	TaxRate          *float64 `json:"tax_rate,omitempty"`
	DiscountType     *string  `json:"discount_type,omitempty"`
	DiscountValue    *float64 `json:"discount_value,omitempty"`
	DisplayOrder     *int     `json:"display_order,omitempty"`
	IsOptional       *bool    `json:"is_optional,omitempty"`
	Notes            *string  `json:"notes,omitempty"`
}

// ReorderQuotationChargesRequest specifies the new ordered list of charge IDs.
type ReorderQuotationChargesRequest struct {
	ChargeIDs []int64 `json:"charge_ids"`
}

// ImportRateChargesRequest instructs the backend to copy line items from a Rate Management rate.
type ImportRateChargesRequest struct {
	RateID           string   `json:"rate_id"`
	ChargeCategories []string `json:"charge_categories,omitempty"`
}

// RateCandidate provides rate surface information for the rate import drawer.
type RateCandidate struct {
	ID                 string                   `json:"id"`
	Source             string                   `json:"source"`
	CarrierSCAC        string                   `json:"carrier_scac"`
	CarrierName        string                   `json:"carrier_name"`
	OriginPort         string                   `json:"origin_port"`
	DestinationPort    string                   `json:"destination_port"`
	EquipmentType      string                   `json:"equipment_type"`
	OceanFreight       float64                  `json:"ocean_freight"`
	OriginCharges      float64                  `json:"origin_charges"`
	DestinationCharges float64                  `json:"destination_charges"`
	TotalBuyPrice      float64                  `json:"total_buy_price"`
	Currency           string                   `json:"currency"`
	TransitDays        *int                     `json:"transit_days,omitempty"`
	ConfidenceScore    int                      `json:"confidence_score"`
	Surcharges         []RateCandidateSurcharge `json:"surcharges"`
}

// RateCandidateSurcharge is a surcharge breakdown on a rate candidate.
type RateCandidateSurcharge struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Unit        string  `json:"unit"`
	Included    bool    `json:"included"`
}

// QuotationActivity is a timeline event for a quotation.
type QuotationActivity struct {
	ID           int64     `db:"id"            json:"id"`
	OrgID        int64     `db:"org_id"        json:"org_id"`
	QuotationID  int64     `db:"quotation_id"  json:"quotation_id"`
	ActivityType string    `db:"activity_type" json:"activity_type"`
	Description  string    `db:"description"   json:"description"`
	Actor        string    `db:"actor"         json:"actor"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
}

// QuotationCustomerInfo holds resolved customer details for the detail panel.
type QuotationCustomerInfo struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	CustomerCode string `json:"customer_code"`
	ContactPhone string `json:"contact_phone"`
	ContactEmail string `json:"contact_email"`
}

// QuotationDetail is the full detail response for the right-side panel.
type QuotationDetail struct {
	Quotation       *Quotation                `json:"quotation"`
	Pricing         *QuotationPricing         `json:"pricing,omitempty"`
	CommercialTerms *QuotationCommercialTerms `json:"commercial_terms,omitempty"`
	Customer        *QuotationCustomerInfo    `json:"customer,omitempty"`
	Activity        []QuotationActivity       `json:"activity"`
}

// QuotationSummary holds fleet-level KPI counts for the summary strip.
type QuotationSummary struct {
	TotalQuotations int `db:"total_quotations" json:"total_quotations"`
	DraftCount      int `db:"draft_count"      json:"draft_count"`
	SentCount       int `db:"sent_count"       json:"sent_count"`
	ViewedCount     int `db:"viewed_count"     json:"viewed_count"`
	AcceptedCount   int `db:"accepted_count"   json:"accepted_count"`
	RejectedCount   int `db:"rejected_count"   json:"rejected_count"`
	ExpiredCount    int `db:"expired_count"    json:"expired_count"`
}

// CreateQuotationRequest holds the fields required to create a new quotation.
type CreateQuotationRequest struct {
	CustomerID      *int64  `json:"customer_id"`
	RFQID           *int64  `json:"rfq_id"`
	Origin          string  `json:"origin"`
	OriginCode      string  `json:"origin_code"`
	Destination     string  `json:"destination"`
	DestinationCode string  `json:"destination_code"`
	ServiceType     string  `json:"service_type"`
	TransportMode   string  `json:"transport_mode"`
	Currency        string  `json:"currency"`
	PaymentTerms    string  `json:"payment_terms"`
	ValidFrom       *string `json:"valid_from"`   // YYYY-MM-DD
	ValidUntil      *string `json:"valid_until"`  // YYYY-MM-DD
	Notes           string  `json:"notes"`
	CommercialTerms string  `json:"commercial_terms"`
	CustomerNotes   string  `json:"customer_notes"`
	InternalNotes   string  `json:"internal_notes"`
	TemplateID      *int64  `json:"template_id,omitempty"`
}

// UpdateQuotationRequest holds the fields that can be mutated on a DRAFT quotation.
type UpdateQuotationRequest struct {
	CustomerID      *int64  `json:"customer_id,omitempty"`
	Origin          *string `json:"origin,omitempty"`
	OriginCode      *string `json:"origin_code,omitempty"`
	Destination     *string `json:"destination,omitempty"`
	DestinationCode *string `json:"destination_code,omitempty"`
	ServiceType     *string `json:"service_type,omitempty"`
	TransportMode   *string `json:"transport_mode,omitempty"`
	Currency        *string `json:"currency,omitempty"`
	PaymentTerms    *string `json:"payment_terms,omitempty"`
	ValidFrom       *string `json:"valid_from,omitempty"`
	ValidUntil      *string `json:"valid_until,omitempty"`
	Notes           *string `json:"notes,omitempty"`
	CommercialTerms *string `json:"commercial_terms,omitempty"`
	CustomerNotes   *string `json:"customer_notes,omitempty"`
	InternalNotes   *string `json:"internal_notes,omitempty"`
}

// QuotationListFilters represents the filtering options for listing quotations.
type QuotationListFilters struct {
	OrgID      int64  `json:"org_id"`
	Search     string `json:"search"`
	Status     string `json:"status"`
	CustomerID *int64 `json:"customer_id"`
	Validity   string `json:"validity"` // ALL | VALID | EXPIRING_SOON | EXPIRED
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
}

// QuotationsListResponse wraps list results with pagination.
type QuotationsListResponse struct {
	Quotations []*QuotationListItem `json:"quotations"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
}

// QuotationApprovalHistory represents an auditable record of a state transition or review action.
type QuotationApprovalHistory struct {
	ID             int64     `db:"id"              json:"id"`
	OrgID          int64     `db:"org_id"          json:"org_id"`
	QuotationID    int64     `db:"quotation_id"    json:"quotation_id"`
	Action         string    `db:"action"          json:"action"`
	PreviousStatus string    `db:"previous_status" json:"previous_status"`
	NewStatus      string    `db:"new_status"      json:"new_status"`
	ActorUserID    *int64    `db:"actor_user_id"   json:"actor_user_id,omitempty"`
	ActorName      string    `db:"actor_name"      json:"actor_name"`
	Comments       string    `db:"comments"        json:"comments"`
	CreatedAt      time.Time `db:"created_at"      json:"created_at"`
}

// CustomerQuotationPreview is the dedicated, safe model returned for customer preview.
// Internal costs, profits, gross margin percentage, internal notes, and rate intelligence are NEVER included.
type CustomerQuotationPreview struct {
	QuotationID     int64                        `json:"quotation_id"`
	QuotationNumber string                       `json:"quotation_number"`
	Status          string                       `json:"status"`
	CustomerID      *int64                       `json:"customer_id,omitempty"`
	CustomerName    string                       `json:"customer_name"`
	Origin          string                       `json:"origin"`
	OriginCode      string                       `json:"origin_code"`
	Destination     string                       `json:"destination"`
	DestinationCode string                       `json:"destination_code"`
	ServiceType     string                       `json:"service_type"`
	TransportMode   string                       `json:"transport_mode"`
	Currency        string                       `json:"currency"`
	PaymentTerms    string                       `json:"payment_terms"`
	ValidFrom       *time.Time                   `json:"valid_from,omitempty"`
	ValidUntil      *time.Time                   `json:"valid_until,omitempty"`
	ValidityStatus  string                       `json:"validity_status"`
	Subtotal        float64                      `json:"subtotal"`
	DiscountTotal   float64                      `json:"discount_total"`
	TaxTotal        float64                      `json:"tax_total"`
	TotalAmount     float64                      `json:"total_amount"`
	CommercialTerms string                       `json:"commercial_terms"`
	CustomerNotes   string                       `json:"customer_notes"`
	SentAt          *time.Time                   `json:"sent_at,omitempty"`
	ViewedAt        *time.Time                   `json:"viewed_at,omitempty"`
	AcceptedAt      *time.Time                   `json:"accepted_at,omitempty"`
	DeclinedAt      *time.Time                   `json:"declined_at,omitempty"`
	Charges         []CustomerQuotationChargeItem `json:"charges"`
	CompanyName     string                       `json:"company_name"`
	CompanyAddress  string                       `json:"company_address"`
	CompanyContact  string                       `json:"company_contact"`
}

// CustomerQuotationChargeItem represents a single charge item formatted safely for customer display.
type CustomerQuotationChargeItem struct {
	ID               int64   `json:"id"`
	ChargeCode       string  `json:"charge_code"`
	ChargeName       string  `json:"charge_name"`
	ChargeCategory   string  `json:"charge_category"`
	CalculationBasis string  `json:"calculation_basis"`
	Quantity         float64 `json:"quantity"`
	UnitPrice        float64 `json:"unit_price"`
	DiscountType     string  `json:"discount_type"`
	DiscountValue    float64 `json:"discount_value"`
	DiscountAmount   float64 `json:"discount_amount"`
	TaxRate          float64 `json:"tax_rate"`
	TaxAmount        float64 `json:"tax_amount"`
	FinalAmount      float64 `json:"final_amount"`
	Currency         string  `json:"currency"`
	IsOptional       bool    `json:"is_optional"`
	SortOrder        int     `json:"sort_order"`
}

// QuotationApprovalStatus provides computed capabilities and review state for the current quotation.
type QuotationApprovalStatus struct {
	Status             string `json:"status"`
	ApprovalRequired   bool   `json:"approval_required"`
	CanSubmitForReview bool   `json:"can_submit_for_review"`
	CanApprove         bool   `json:"can_approve"`
	CanRequestChanges  bool   `json:"can_request_changes"`
	CanSend            bool   `json:"can_send"`
	CanAccept          bool   `json:"can_accept"`
	CanDecline         bool   `json:"can_decline"`
	CanCancel          bool   `json:"can_cancel"`
	CanEdit            bool   `json:"can_edit"`
}

// SubmitQuotationForReviewRequest represents a request to submit a draft quotation for management review.
type SubmitQuotationForReviewRequest struct {
	Comments string `json:"comments"`
}

// ApproveQuotationRequest represents a request to approve a quotation.
type ApproveQuotationRequest struct {
	ApprovalNotes string `json:"approval_notes"`
}

// RequestQuotationChangesRequest represents a request to ask for modifications on a submitted quotation.
type RequestQuotationChangesRequest struct {
	Reason string `json:"reason"`
}

// SendQuotationRequest represents a request to mark a quotation as sent to the customer.
type SendQuotationRequest struct {
	Comments       string `json:"comments"`
	RecipientEmail string `json:"recipient_email"`
	SendCopy       bool   `json:"send_copy"`
}

// MarkQuotationViewedRequest represents a request to record a quotation view event.
type MarkQuotationViewedRequest struct {
	ViewerName  string `json:"viewer_name"`
	ViewerEmail string `json:"viewer_email"`
	IPAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
}

// AcceptQuotationRequest represents a customer acceptance payload.
type AcceptQuotationRequest struct {
	AcceptedBy string `json:"accepted_by"`
	Comments   string `json:"comments"`
}

// DeclineQuotationRequest represents a customer decline payload.
type DeclineQuotationRequest struct {
	DeclinedBy string `json:"declined_by"`
	Reason     string `json:"reason"`
}

// CancelQuotationRequest represents an administrative cancellation payload.
type CancelQuotationRequest struct {
	Reason string `json:"reason"`
}

// ── Document & Public Sharing Models (Task 18.5) ─────────────────────────────

// QuotationDocument represents a generated document (PDF, Customer Copy) for a quotation.
type QuotationDocument struct {
	ID           int64     `db:"id"            json:"id"`
	OrgID        int64     `db:"org_id"        json:"org_id"`
	QuotationID  int64     `db:"quotation_id"  json:"quotation_id"`
	DocumentType string    `db:"document_type" json:"document_type"`
	FileName     string    `db:"file_name"     json:"file_name"`
	FilePath     string    `db:"file_path"     json:"file_path"`
	Version      int       `db:"version"       json:"version"`
	GeneratedAt  time.Time `db:"generated_at"  json:"generated_at"`
	GeneratedBy  *int64    `db:"generated_by"  json:"generated_by,omitempty"`
	CreatedAt    time.Time `db:"created_at"    json:"created_at"`
}

// QuotationPublicLink represents a cryptographically secure sharing token for public quotation access.
type QuotationPublicLink struct {
	ID               int64      `db:"id"                json:"id"`
	OrgID            int64      `db:"org_id"            json:"org_id"`
	QuotationID      int64      `db:"quotation_id"      json:"quotation_id"`
	PublicToken      string     `db:"public_token"      json:"public_token"`
	Status           string     `db:"status"            json:"status"`
	ExpiresAt        *time.Time `db:"expires_at"        json:"expires_at,omitempty"`
	CreatedBy        int64      `db:"created_by"        json:"created_by"`
	RevokedAt        *time.Time `db:"revoked_at"        json:"revoked_at,omitempty"`
	RevokedBy        *int64     `db:"revoked_by"        json:"revoked_by,omitempty"`
	RevocationReason *string    `db:"revocation_reason" json:"revocation_reason,omitempty"`
	LastAccessedAt   *time.Time `db:"last_accessed_at"  json:"last_accessed_at,omitempty"`
	AccessCount      int        `db:"access_count"      json:"access_count"`
	CreatedAt        time.Time  `db:"created_at"        json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"        json:"updated_at"`
}

// QuotationPublicViewResponse is returned by public quotation retrieval.
type QuotationPublicViewResponse struct {
	PublicToken string                    `json:"public_token"`
	Status      string                    `json:"status"`
	ExpiresAt   *time.Time                `json:"expires_at,omitempty"`
	Quotation   *CustomerQuotationPreview `json:"quotation"`
	AccessCount int                       `json:"access_count"`
	CanAccept   bool                      `json:"can_accept"`
	CanDecline  bool                      `json:"can_decline"`
}

// CreateQuotationPublicLinkRequest defines payload to create a sharing link.
type CreateQuotationPublicLinkRequest struct {
	ValidityDays *int    `json:"validity_days,omitempty"`
	ExpiresAt    *string `json:"expires_at,omitempty"`
}

// RevokeQuotationPublicLinkRequest defines payload to revoke a public link.
type RevokeQuotationPublicLinkRequest struct {
	Reason string `json:"reason,omitempty"`
}

// PublicAcceptQuotationRequest is submitted by customer on the public quote page.
type PublicAcceptQuotationRequest struct {
	AcceptedBy string `json:"accepted_by"`
	Comments   string `json:"comments,omitempty"`
}

// PublicDeclineQuotationRequest is submitted by customer on the public quote page.
type PublicDeclineQuotationRequest struct {
	DeclinedBy string `json:"declined_by"`
	Reason     string `json:"reason,omitempty"`
}

// APIResponse is the standard envelope used across the quotation transport layer.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ─── Quotation-to-Booking Conversion Models (Task 18.6) ───────────────────────

// QuotationConversionPreview displays what commercial and shipment data will transfer to operations.
type QuotationConversionPreview struct {
	QuotationID     int64                     `json:"quotation_id"`
	QuotationNumber string                    `json:"quotation_number"`
	CustomerID      *int64                    `json:"customer_id,omitempty"`
	CustomerName    string                    `json:"customer_name"`
	CustomerCode    string                    `json:"customer_code,omitempty"`
	Origin          string                    `json:"origin"`
	OriginCode      string                    `json:"origin_code"`
	Destination     string                    `json:"destination"`
	DestinationCode string                    `json:"destination_code"`
	ServiceType     string                    `json:"service_type"`
	TransportMode   string                    `json:"transport_mode"`
	Equipment       string                    `json:"equipment"`
	Currency        string                    `json:"currency"`
	PaymentTerms    string                    `json:"payment_terms"`
	Subtotal        float64                   `json:"subtotal"`
	Surcharges      float64                   `json:"surcharges"`
	Taxes           float64                   `json:"taxes"`
	TotalAmount     float64                   `json:"total_amount"`
	ValidUntil      *time.Time                `json:"valid_until,omitempty"`
	AcceptedAt      *time.Time                `json:"accepted_at,omitempty"`
	CommercialTerms string                    `json:"commercial_terms"`
	CustomerNotes   string                    `json:"customer_notes"`
	SelectedCharges []*QuotationChargeItem    `json:"selected_charges"`
	CanConvert      bool                      `json:"can_convert"`
	BlockingReasons []string                  `json:"blocking_reasons"`
	ConversionStatus string                   `json:"conversion_status"`
	ExistingBookingID *int64                  `json:"existing_booking_id,omitempty"`
	ExistingShipmentID *int64                 `json:"existing_shipment_id,omitempty"`
}

// ConvertQuotationToBookingRequest allows the operator to provide/override operational parameters.
type ConvertQuotationToBookingRequest struct {
	BookingNumber       *string    `json:"booking_number,omitempty"`
	CarrierID           *string    `json:"carrier_id,omitempty"`
	CarrierName         string     `json:"carrier_name"`
	CarrierSCAC         *string    `json:"carrier_scac,omitempty"`
	OriginPort          *string    `json:"origin_port,omitempty"`
	DestinationPort     *string    `json:"destination_port,omitempty"`
	VesselName          *string    `json:"vessel_name,omitempty"`
	VoyageNumber        *string    `json:"voyage_number,omitempty"`
	ETD                 *time.Time `json:"etd,omitempty"`
	ETA                 *time.Time `json:"eta,omitempty"`
	CargoSummary        *string    `json:"cargo_summary,omitempty"`
	SpecialInstructions *string    `json:"special_instructions,omitempty"`
	OperationalNotes    string     `json:"operational_notes,omitempty"`
	CreateShipmentImmediately bool `json:"create_shipment_immediately"`
}

// QuotationConversionResult returns the handover confirmation details.
type QuotationConversionResult struct {
	Success          bool      `json:"success"`
	QuotationID      int64     `json:"quotation_id"`
	QuotationNumber  string    `json:"quotation_number"`
	BookingID        int64     `json:"booking_id"`
	BookingNumber    string    `json:"booking_number"`
	ShipmentID       *int64    `json:"shipment_id,omitempty"`
	ShipmentNumber   *string   `json:"shipment_number,omitempty"`
	Message          string    `json:"message"`
	ConversionStatus string    `json:"conversion_status"`
	ConvertedAt      time.Time `json:"converted_at"`
	AlreadyConverted bool      `json:"already_converted"`
}

// QuotationConversionHistory tracks immutable audit records of handover attempts.
type QuotationConversionHistory struct {
	ID          int64     `db:"id"           json:"id"`
	OrgID       int64     `db:"org_id"       json:"org_id"`
	QuotationID int64     `db:"quotation_id" json:"quotation_id"`
	BookingID   *int64    `db:"booking_id"   json:"booking_id,omitempty"`
	ShipmentID  *int64    `db:"shipment_id"  json:"shipment_id,omitempty"`
	Action      string    `db:"action"       json:"action"`
	Status      string    `db:"status"       json:"status"`
	Message     string    `db:"message"      json:"message"`
	PerformedBy string    `db:"performed_by" json:"performed_by"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
}

// ─── Booking Confirmation & Traceability Models (Task 18.7) ─────────────────

// QuotationCommercialSnapshot represents the frozen accepted commercial record.
type QuotationCommercialSnapshot struct {
	QuotationID     int64      `json:"quotation_id"`
	QuotationNumber string     `json:"quotation_number"`
	CustomerID      *int64     `json:"customer_id,omitempty"`
	CustomerName    string     `json:"customer_name"`
	AcceptedAt      *time.Time `json:"accepted_at,omitempty"`
	AcceptedBy      string     `json:"accepted_by,omitempty"`
	Currency        string     `json:"currency"`
	Subtotal        float64    `json:"subtotal"`
	Surcharges      float64    `json:"surcharges"`
	Taxes           float64    `json:"taxes"`
	TotalAmount     float64    `json:"total_amount"`
	PaymentTerms    string     `json:"payment_terms"`
	CommercialTerms string     `json:"commercial_terms"`
	CustomerNotes   string     `json:"customer_notes"`
	Origin          string     `json:"origin"`
	OriginCode      string     `json:"origin_code"`
	Destination     string     `json:"destination"`
	DestinationCode string     `json:"destination_code"`
	ServiceType     string     `json:"service_type"`
	TransportMode   string     `json:"transport_mode"`
}

// OperationalChange represents a detected modification in operational booking/shipment data vs the commercial quote baseline.
type OperationalChange struct {
	Field          string    `json:"field"`
	Category       string    `json:"category"` // e.g. "ROUTING", "CARRIER", "SCHEDULE", "EQUIPMENT"
	BaselineValue  string    `json:"baseline_value"`
	CurrentValue   string    `json:"current_value"`
	ChangedAt      time.Time `json:"changed_at"`
	ChangedBy      string    `json:"changed_by"`
	IsCommercial   bool      `json:"is_commercial"` // Always false: operational changes do not alter commercial quote
	ImpactSeverity string    `json:"impact_severity"` // "INFO", "WARNING", "CRITICAL"
}

// LineageStep represents a single node in the end-to-end lineage chain.
type LineageStep struct {
	StepID      string     `json:"step_id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"` // "COMPLETED", "ACTIVE", "PENDING", "SKIPPED"
	ReferenceID *int64     `json:"reference_id,omitempty"`
	RefNumber   string     `json:"ref_number,omitempty"`
	URL         string     `json:"url,omitempty"`
	Timestamp   *time.Time `json:"timestamp,omitempty"`
	Actor       string     `json:"actor,omitempty"`
}

// QuotationOperationalHandover contains the unified end-to-end lineage and operational status.
type QuotationOperationalHandover struct {
	QuotationID             int64                        `json:"quotation_id"`
	QuotationNumber         string                       `json:"quotation_number"`
	RFQID                   *int64                       `json:"rfq_id,omitempty"`
	RFQNumber               string                       `json:"rfq_number,omitempty"`
	BookingID               *int64                       `json:"booking_id,omitempty"`
	BookingNumber           string                       `json:"booking_number,omitempty"`
	ShipmentID              *int64                       `json:"shipment_id,omitempty"`
	ShipmentNumber          string                       `json:"shipment_number,omitempty"`
	CustomerName            string                       `json:"customer_name"`
	ConversionStatus        string                       `json:"conversion_status"`
	HandoverStatus          string                       `json:"handover_status"`
	CommercialSnapshot      *QuotationCommercialSnapshot `json:"commercial_snapshot"`
	OperationalChanges      []*OperationalChange         `json:"operational_changes"`
	LineageChain            []*LineageStep               `json:"lineage_chain"`
	ConvertedAt             *time.Time                   `json:"converted_at,omitempty"`
	ConvertedBy             string                       `json:"converted_by,omitempty"`
	BookingConfirmedAt      *time.Time                   `json:"booking_confirmed_at,omitempty"`
	BookingConfirmedBy      string                       `json:"booking_confirmed_by,omitempty"`
	CanConfirmHandover      bool                         `json:"can_confirm_handover"`
	BlockingReasons         []string                     `json:"blocking_reasons,omitempty"`
	CurrentCarrier          string                       `json:"current_carrier,omitempty"`
	CurrentVessel           string                       `json:"current_vessel,omitempty"`
	CurrentVoyage           string                       `json:"current_voyage,omitempty"`
	CurrentETD              *time.Time                   `json:"current_etd,omitempty"`
	CurrentETA              *time.Time                   `json:"current_eta,omitempty"`
	TrackingStatus          string                       `json:"tracking_status,omitempty"`
}

// QuotationOperationalHandoverHistory represents an immutable event in the handover lifecycle.
type QuotationOperationalHandoverHistory struct {
	ID          int64           `db:"id"           json:"id"`
	OrgID       int64           `db:"org_id"       json:"org_id"`
	QuotationID int64           `db:"quotation_id" json:"quotation_id"`
	BookingID   *int64          `db:"booking_id"   json:"booking_id,omitempty"`
	ShipmentID  *int64          `db:"shipment_id"  json:"shipment_id,omitempty"`
	EventType   string          `db:"event_type"   json:"event_type"`
	Description string          `db:"description"  json:"description"`
	Metadata    json.RawMessage `db:"metadata"     json:"metadata,omitempty"`
	PerformedBy string          `db:"performed_by" json:"performed_by"`
	CreatedAt   time.Time       `db:"created_at"   json:"created_at"`
}

// ConfirmQuotationHandoverRequest payload for confirming booking handover to operations.
type ConfirmQuotationHandoverRequest struct {
	ConfirmationNotes string `json:"confirmation_notes"`
	NotifyOperations  bool   `json:"notify_operations"`
}

// ─── Quotation Analytics, Performance & Intelligence (Task 18.8) ────────────

// QuotationAnalyticsOverview represents the high-level management executive KPI dashboard.
type QuotationAnalyticsOverview struct {
	// Volume metrics
	TotalQuotes          int `json:"total_quotes"`
	DraftQuotes          int `json:"draft_quotes"`
	ReadyForReviewQuotes int `json:"ready_for_review_quotes"`
	ApprovedQuotes       int `json:"approved_quotes"`
	SentQuotes           int `json:"sent_quotes"`
	ViewedQuotes         int `json:"viewed_quotes"`
	AcceptedQuotes       int `json:"accepted_quotes"`
	DeclinedQuotes       int `json:"declined_quotes"`
	ExpiredQuotes        int `json:"expired_quotes"`
	CancelledQuotes      int `json:"cancelled_quotes"`

	// Commercial Value metrics
	PipelineValue         float64 `json:"pipeline_value"`
	AcceptedValue         float64 `json:"accepted_value"`
	ConvertedBookingValue float64 `json:"converted_booking_value"`
	AverageQuoteValue     float64 `json:"average_quote_value"`
	AverageGrossMarginPct float64 `json:"average_gross_margin_pct"`
	TotalGrossProfit      float64 `json:"total_gross_profit"`
	Currency              string  `json:"currency"`

	// Performance Conversion rates
	AcceptanceRate                   float64 `json:"acceptance_rate"`
	DeclineRate                      float64 `json:"decline_rate"`
	QuoteToBookingConversionRate     float64 `json:"quote_to_booking_conversion_rate"`
	AverageApprovalTimeHours         float64 `json:"average_approval_time_hours"`
	AverageCustomerResponseTimeHours float64 `json:"average_customer_response_time_hours"`

	// Risk metrics
	ExpiringSoonCount    int `json:"expiring_soon_count"`
	StuckInReviewCount   int `json:"stuck_in_review_count"`
	UnviewedSentQuotes   int `json:"unviewed_sent_quotes"`

	// Management Intelligence Insights
	Insights []*QuotationOperationalInsight `json:"insights"`
}

// QuotationTrendDataPoint represents time-series commercial activity for charts.
type QuotationTrendDataPoint struct {
	Date           string  `json:"date"`
	QuotesCreated  int     `json:"quotes_created"`
	QuotesSent     int     `json:"quotes_sent"`
	QuotesAccepted int     `json:"quotes_accepted"`
	QuotesDeclined int     `json:"quotes_declined"`
	QuotesExpired  int     `json:"quotes_expired"`
	PipelineValue  float64 `json:"pipeline_value"`
	AcceptedValue  float64 `json:"accepted_value"`
	AverageMargin  float64 `json:"average_margin"`
}

// CustomerQuotationPerformance represents aggregated quotation performance per customer.
type CustomerQuotationPerformance struct {
	CustomerID        *int64  `json:"customer_id,omitempty"`
	CustomerName      string  `json:"customer_name"`
	QuoteCount        int     `json:"quote_count"`
	AcceptedQuotes    int     `json:"accepted_quotes"`
	DeclinedQuotes    int     `json:"declined_quotes"`
	AcceptanceRate    float64 `json:"acceptance_rate"`
	PipelineValue     float64 `json:"pipeline_value"`
	AcceptedValue     float64 `json:"accepted_value"`
	AverageQuoteValue float64 `json:"average_quote_value"`
	ConvertedBookings int     `json:"converted_bookings"`
}

// QuotationPerformanceByMode represents aggregated quotation metrics broken down by transport/service mode.
type QuotationPerformanceByMode struct {
	TransportMode     string  `json:"transport_mode"`
	ServiceType       string  `json:"service_type"`
	QuoteCount        int     `json:"quote_count"`
	AcceptedCount     int     `json:"accepted_count"`
	AcceptanceRate    float64 `json:"acceptance_rate"`
	PipelineValue     float64 `json:"pipeline_value"`
	AcceptedValue     float64 `json:"accepted_value"`
	AverageQuoteValue float64 `json:"average_quote_value"`
	AverageMarginPct  float64 `json:"average_margin_pct"`
}

// QuotationRiskItem represents a specific quotation requiring attention or at risk of expiring/stalling.
type QuotationRiskItem struct {
	QuotationID       int64      `json:"quotation_id"`
	QuotationNumber   string     `json:"quotation_number"`
	CustomerName      string     `json:"customer_name"`
	TotalAmount       float64    `json:"total_amount"`
	Currency          string     `json:"currency"`
	Status            string     `json:"status"`
	ValidUntil        *time.Time `json:"valid_until,omitempty"`
	DaysUntilExpiry   int        `json:"days_until_expiry"`
	RiskCategory      string     `json:"risk_category"` // "EXPIRING_SOON", "EXPIRED", "UNVIEWED_SENT", "STUCK_IN_REVIEW"
	RecommendedAction string     `json:"recommended_action"`
}

// QuotationOperationalInsight represents a rule-based deterministic recommendation.
type QuotationOperationalInsight struct {
	Category          string `json:"category"`           // "CONVERSION", "APPROVAL", "EXPIRY", "CUSTOMER", "MARGIN", "OPERATIONS"
	Severity          string `json:"severity"`           // "INFO", "WARNING", "CRITICAL", "SUCCESS"
	Headline          string `json:"headline"`
	Description       string `json:"description"`
	MetricValue       string `json:"metric_value"`
	RecommendedAction string `json:"recommended_action"`
}

// ── Task 19.5: Rate-to-Quotation Domain Models ────────────────────────────────

// QuotationRateSelection represents an active or historical rate selection for a quotation.
type QuotationRateSelection struct {
	ID                 int64      `db:"id"                    json:"id"`
	OrgID              int64      `db:"org_id"                json:"org_id"`
	QuotationID        int64      `db:"quotation_id"          json:"quotation_id"`
	RateID             *int64     `db:"rate_id"               json:"rate_id,omitempty"`
	SpotRateRequestID  *int64     `db:"spot_rate_request_id"  json:"spot_rate_request_id,omitempty"`
	SpotRateResponseID *int64     `db:"spot_rate_response_id" json:"spot_rate_response_id,omitempty"`
	RateSourceType     string     `db:"rate_source_type"      json:"rate_source_type"` // MANAGED_RATE, SPOT_RATE, CUSTOM_RATE
	SelectedBy         string     `db:"selected_by"           json:"selected_by"`
	SelectedAt         time.Time  `db:"selected_at"           json:"selected_at"`
	IsActive           bool       `db:"is_active"             json:"is_active"`
	Notes              string     `db:"notes"                 json:"notes"`
	CreatedAt          time.Time  `db:"created_at"            json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"            json:"updated_at"`
}

// QuotationRateSnapshot is the immutable commercial rate record linked to a quotation selection.
type QuotationRateSnapshot struct {
	ID                       int64      `db:"id"                           json:"id"`
	OrgID                    int64      `db:"org_id"                       json:"org_id"`
	QuotationID              int64      `db:"quotation_id"                 json:"quotation_id"`
	QuotationRateSelectionID int64      `db:"quotation_rate_selection_id" json:"quotation_rate_selection_id"`
	SourceRateID             *int64     `db:"source_rate_id"               json:"source_rate_id,omitempty"`
	SourceRateVersion        *int       `db:"source_rate_version"          json:"source_rate_version,omitempty"`
	SourceContractID         *int64     `db:"source_contract_id"           json:"source_contract_id,omitempty"`
	SourceSpotRateRequestID  *int64     `db:"source_spot_rate_request_id"  json:"source_spot_rate_request_id,omitempty"`
	SourceSpotRateResponseID *int64     `db:"source_spot_rate_response_id" json:"source_spot_rate_response_id,omitempty"`
	CarrierName              string     `db:"carrier_name"                 json:"carrier_name"`
	CarrierReference         string     `db:"carrier_reference"            json:"carrier_reference"`
	TransportMode            string     `db:"transport_mode"               json:"transport_mode"`
	ServiceType              string     `db:"service_type"                 json:"service_type"`
	EquipmentType            string     `db:"equipment_type"               json:"equipment_type"`
	Origin                   string     `db:"origin"                       json:"origin"`
	Destination              string     `db:"destination"                  json:"destination"`
	Currency                 string     `db:"currency"                     json:"currency"`
	BaseRate                 float64    `db:"base_rate"                    json:"base_rate"`
	AdditionalCharges        float64    `db:"additional_charges"           json:"additional_charges"`
	CommercialTotal          float64    `db:"commercial_total"             json:"commercial_total"`
	PricingSnapshotJSON      string     `db:"pricing_snapshot"             json:"-"`
	PricingSnapshot          interface{} `json:"pricing_snapshot,omitempty"`
	ValidFrom                *time.Time `db:"valid_from"                   json:"valid_from,omitempty"`
	ValidUntil               *time.Time `db:"valid_until"                  json:"valid_until,omitempty"`
	SnapshotCreatedAt        time.Time  `db:"snapshot_created_at"          json:"snapshot_created_at"`
	CreatedBy                string     `db:"created_by"                   json:"created_by"`
	CreatedAt                time.Time  `db:"created_at"                   json:"created_at"`
	UpdatedAt                time.Time  `db:"updated_at"                   json:"updated_at"`
}

// QuotationRateSelectionHistory tracks immutable audit timeline events for rate selections.
type QuotationRateSelectionHistory struct {
	ID                  int64     `db:"id"                    json:"id"`
	OrgID               int64     `db:"org_id"                json:"org_id"`
	QuotationID         int64     `db:"quotation_id"          json:"quotation_id"`
	EventType           string    `db:"event_type"            json:"event_type"`
	PreviousSelectionID *int64    `db:"previous_selection_id" json:"previous_selection_id,omitempty"`
	NewSelectionID      *int64    `db:"new_selection_id"      json:"new_selection_id,omitempty"`
	Description         string    `db:"description"           json:"description"`
	MetadataJSON        string    `db:"metadata"              json:"-"`
	Metadata            interface{} `json:"metadata,omitempty"`
	PerformedBy         string    `db:"performed_by"          json:"performed_by"`
	CreatedAt           time.Time `db:"created_at"            json:"created_at"`
}

// QuotationRateCandidate represents an available managed or spot rate for quotation selection.
type QuotationRateCandidate struct {
	SourceType             string     `json:"source_type"` // MANAGED_RATE, SPOT_RATE
	RateID                 *int64     `json:"rate_id,omitempty"`
	SpotRateRequestID      *int64     `json:"spot_rate_request_id,omitempty"`
	SpotRateResponseID     *int64     `json:"spot_rate_response_id,omitempty"`
	CarrierName            string     `json:"carrier_name"`
	CarrierCode            string     `json:"carrier_code,omitempty"`
	RateType               string     `json:"rate_type"`
	RateVersion            int        `json:"rate_version"`
	ContractID             *int64     `json:"contract_id,omitempty"`
	ContractCode           string     `json:"contract_code,omitempty"`
	Origin                 string     `json:"origin"`
	Destination            string     `json:"destination"`
	TransportMode          string     `json:"transport_mode"`
	ServiceType            string     `json:"service_type"`
	EquipmentType          string     `json:"equipment_type"`
	Currency               string     `json:"currency"`
	BaseRate               float64    `json:"base_rate"`
	TotalCharges           float64    `json:"total_charges"`
	CommercialTotal        float64    `json:"commercial_total"`
	TransitDays            int        `json:"transit_days"`
	FreeDaysOrigin         int        `json:"free_days_origin"`
	FreeDaysDestination    int        `json:"free_days_destination"`
	ValidFrom              *time.Time `json:"valid_from,omitempty"`
	ValidUntil             *time.Time `json:"valid_until,omitempty"`
	Status                 string     `json:"status"`
	RecommendationTags     []string   `json:"recommendation_tags,omitempty"`
	RiskWarnings           []string   `json:"risk_warnings,omitempty"`
	Charges                []interface{} `json:"charges,omitempty"`
}

// DTO Requests and Responses
type SelectQuotationRateRequest struct {
	OrgID              int64  `json:"-"`
	QuotationID        int64  `json:"-"`
	RateID             *int64 `json:"rate_id,omitempty"`
	SpotRateResponseID *int64 `json:"spot_rate_response_id,omitempty"`
	SourceType         string `json:"source_type"` // MANAGED_RATE, SPOT_RATE
	Notes              string `json:"notes"`
	User               string `json:"-"`
}

type ReplaceQuotationRateRequest struct {
	OrgID              int64  `json:"-"`
	QuotationID        int64  `json:"-"`
	RateID             *int64 `json:"rate_id,omitempty"`
	SpotRateResponseID *int64 `json:"spot_rate_response_id,omitempty"`
	SourceType         string `json:"source_type"`
	Notes              string `json:"notes"`
	User               string `json:"-"`
}

type QuotationRateCandidatesResponse struct {
	QuotationID        int64                    `json:"quotation_id"`
	Origin             string                   `json:"origin"`
	Destination        string                   `json:"destination"`
	TransportMode      string                   `json:"transport_mode"`
	ServiceType        string                   `json:"service_type"`
	EquipmentType      string                   `json:"equipment_type"`
	Candidates         []QuotationRateCandidate `json:"candidates"`
	TotalCandidates    int                      `json:"total_candidates"`
	CheapestCandidate  *QuotationRateCandidate  `json:"cheapest_candidate,omitempty"`
	FastestCandidate   *QuotationRateCandidate  `json:"fastest_candidate,omitempty"`
	BestValueCandidate *QuotationRateCandidate  `json:"best_value_candidate,omitempty"`
}

// ── Task 19.6: Quotation Rate Risk & Commercial Impact Types ──────────────────

type QuotationRateRisk struct {
	ID                       int64       `db:"id"                          json:"id"`
	OrgID                    int64       `db:"org_id"                      json:"org_id"`
	QuotationID              int64       `db:"quotation_id"                json:"quotation_id"`
	QuotationRateSnapshotID  *int64      `db:"quotation_rate_snapshot_id"  json:"quotation_rate_snapshot_id,omitempty"`
	SourceRateID             *int64      `db:"source_rate_id"              json:"source_rate_id,omitempty"`
	SourceContractID         *int64      `db:"source_contract_id"          json:"source_contract_id,omitempty"`
	SourceSpotRateResponseID *int64      `db:"source_spot_rate_response_id" json:"source_spot_rate_response_id,omitempty"`
	RiskType                 string      `db:"risk_type"                   json:"risk_type"`
	Severity                 string      `db:"severity"                    json:"severity"` // 'INFO', 'WARNING', 'CRITICAL'
	Headline                 string      `db:"headline"                    json:"headline"`
	Description              string      `db:"description"                 json:"description"`
	RecommendedAction        string      `db:"recommended_action"          json:"recommended_action,omitempty"`
	IsResolved               bool        `db:"is_resolved"                 json:"is_resolved"`
	ResolvedBy               *string     `db:"resolved_by"                 json:"resolved_by,omitempty"`
	ResolvedAt               *time.Time  `db:"resolved_at"                 json:"resolved_at,omitempty"`
	MetadataJSON             string      `db:"metadata"                    json:"-"`
	Metadata                 interface{} `db:"-"                           json:"metadata,omitempty"`
	CreatedAt                time.Time   `db:"created_at"                  json:"created_at"`
	UpdatedAt                time.Time   `db:"updated_at"                  json:"updated_at"`
}

type QuotationRateRiskSummary struct {
	QuotationID        int64                `json:"quotation_id"`
	QuotationNumber    string               `json:"quotation_number"`
	Status             string               `json:"status"`
	HasActiveSnapshot  bool                 `json:"has_active_snapshot"`
	TotalRisks         int                  `json:"total_risks"`
	CriticalRisks      int                  `json:"critical_risks"`
	WarningRisks       int                  `json:"warning_risks"`
	InfoRisks          int                  `json:"info_risks"`
	Risks              []*QuotationRateRisk `json:"risks"`
	ReplacementCount   int                  `json:"replacement_count"`
}

type RateReplacementCandidate struct {
	SourceType         string   `json:"source_type"` // 'MANAGED_RATE', 'SPOT_RATE'
	RateID             *int64   `json:"rate_id,omitempty"`
	SpotRateResponseID *int64   `json:"spot_rate_response_id,omitempty"`
	CarrierName        string   `json:"carrier_name"`
	CarrierCode        string   `json:"carrier_code,omitempty"`
	RateType           string   `json:"rate_type"`
	VersionNumber      int      `json:"version_number"`
	Currency           string   `json:"currency"`
	BaseRate           float64  `json:"base_rate"`
	CommercialTotal    float64  `json:"commercial_total"`
	TransitDays        int      `json:"transit_days"`
	ValidUntil         *string  `json:"valid_until,omitempty"`
	RecommendationTags []string `json:"recommendation_tags,omitempty"`
}

type CommercialImpactAnalysis struct {
	QuotationID               int64   `json:"quotation_id"`
	QuotationNumber           string  `json:"quotation_number"`
	QuotationStatus           string  `json:"quotation_status"`
	CurrentCarrierName        string  `json:"current_carrier_name"`
	CurrentCommercialTotal    float64 `json:"current_commercial_total"`
	CurrentCurrency           string  `json:"current_currency"`
	CurrentTransitDays        int     `json:"current_transit_days"`
	CurrentValidUntil         *string `json:"current_valid_until,omitempty"`
	ReplacementSourceType     string  `json:"replacement_source_type"`
	ReplacementCarrierName    string  `json:"replacement_carrier_name"`
	ReplacementCommercialTotal float64 `json:"replacement_commercial_total"`
	ReplacementCurrency       string  `json:"replacement_currency"`
	ReplacementTransitDays    int     `json:"replacement_transit_days"`
	ReplacementValidUntil     *string `json:"replacement_valid_until,omitempty"`
	PriceDifferenceAmount     float64 `json:"price_difference_amount"`
	PriceDifferencePercentage float64 `json:"price_difference_percentage"`
	TransitDifferenceDays     int     `json:"transit_difference_days"`
	IsCheaper                 bool    `json:"is_cheaper"`
	IsFaster                  bool    `json:"is_faster"`
	CurrencyMismatch          bool    `json:"currency_mismatch"`
	ImpactSummary             string  `json:"impact_summary"`
}

type QuotationRateSelectionDetail struct {
	QuotationID              int64      `db:"quotation_id"`
	QuotationNumber          string     `db:"quotation_number"`
	QuotationStatus          string     `db:"quotation_status"`
	SnapshotID               int64      `db:"snapshot_id"`
	SnapshotCarrierName      string     `db:"carrier_name"`
	SnapshotCurrency         string     `db:"currency"`
	SnapshotCommercialTotal  float64    `db:"commercial_total"`
	SnapshotValidUntil       *time.Time `db:"valid_until"`
	SourceRateID             *int64     `db:"source_rate_id"`
	SourceRateStatus         string     `db:"source_rate_status"`
	SourceRateValidUntil     *time.Time `db:"source_rate_valid_until"`
	SourceRateVersion        int        `db:"source_rate_version"`
	LatestRateVersion        int        `db:"latest_rate_version"`
	SourceContractID         *int64     `db:"source_contract_id"`
	SourceContractCode       string     `db:"source_contract_code"`
	SourceContractStatus     string     `db:"source_contract_status"`
	SourceContractEndDate    *time.Time `db:"source_contract_end_date"`
	SourceSpotRateResponseID *int64     `db:"source_spot_rate_response_id"`
	SourceSpotValidUntil     *time.Time `db:"source_spot_valid_until"`
	SourceSpotStatus         string     `db:"source_spot_status"`
}








