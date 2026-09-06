package search

import "time"

type Category string

const (
	CategoryShipment  Category = "SHIPMENT"
	CategoryBooking   Category = "BOOKING"
	CategoryRFQ       Category = "RFQ"
	CategoryQuotation Category = "QUOTATION"
	CategoryCustomer  Category = "CUSTOMER"
	CategoryInvoice   Category = "INVOICE"
	CategoryLead      Category = "LEAD"
	CategoryContract  Category = "CONTRACT"
	CategoryDocument  Category = "DOCUMENT"
	CategoryTracking  Category = "TRACKING"
)

type SearchItem struct {
	ID        string                 `json:"id"`
	Category  Category               `json:"category"`
	Title     string                 `json:"title"`
	Subtitle  string                 `json:"subtitle"`
	Badge     string                 `json:"badge,omitempty"`
	BadgeType string                 `json:"badge_type,omitempty"` // "success", "info", "warning", "neutral", "danger"
	URL       string                 `json:"url"`
	CreatedAt *time.Time             `json:"created_at,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type SearchResultGroup struct {
	Category      Category     `json:"category"`
	CategoryLabel string       `json:"category_label"`
	Icon          string       `json:"icon"`
	Count         int          `json:"count"`
	Items         []SearchItem `json:"items"`
}

type SearchResponse struct {
	Query        string              `json:"query"`
	TotalMatches int                 `json:"total_matches"`
	Groups       []SearchResultGroup `json:"groups"`
}
