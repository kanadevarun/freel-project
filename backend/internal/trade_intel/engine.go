package trade_intel

import "context"

// CompanyIntel represents the enriched data we gather about a company.
// Simple meaning: This is a standardized box where we put all the information
// we find out about a company (like how much money they make, how many people
// work there, and how many shipping containers they move).
type CompanyIntel struct {
	// Name is the legal or known name of the company.
	Name string `json:"name"`
	
	// Domain is the website address (e.g., "apple.com").
	Domain string `json:"domain"`
	
	// Industry describes what the company does (e.g., "Electronics", "Logistics").
	Industry string `json:"industry"`
	
	// EstimatedRevenue string representing how much money they make (e.g., "$10M - $50M").
	EstimatedRevenue string `json:"estimated_revenue"`
	
	// EmployeeCount represents the estimated number of workers.
	EmployeeCount int `json:"employee_count"`
	
	// MonthlyShippingVolume estimates how many TEUs (Twenty-foot Equivalent Units - standard shipping containers) they ship per month.
	MonthlyShippingVolume int `json:"monthly_shipping_volume"`
	
	// TopSuppliers is a list of company names they buy from.
	TopSuppliers []string `json:"top_suppliers"`
	
	// IsExporter is true if they export goods, false otherwise.
	IsExporter bool `json:"is_exporter"`
}

// Engine defines the blueprint for our Trade Intelligence service.
// Simple meaning: Think of this as an instruction manual. It says that any
// Trade Intelligence tool we plug into our system MUST have an "EnrichCompany" function.
// This allows us to easily swap out data providers (like Apollo, ZoomInfo, or ImportYeti) later.
type Engine interface {
	// EnrichCompany takes a company name or domain and tries to find all available data about it.
	// Simple meaning: You give it a company name (like "Tesla"), and it returns a filled-out
	// CompanyIntel box with all the details it could find on the internet.
	// Example: intel, err := engine.EnrichCompany(ctx, "Tesla")
	EnrichCompany(ctx context.Context, companyName string) (*CompanyIntel, error)
}
