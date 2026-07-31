package trade_intel

import (
	"context"
	"strings"
	"time"
)

// mockEngine is a fake version of our Trade Intelligence Engine.
// Simple meaning: Since we don't have real paid subscriptions to data providers
// like Apollo or ImportYeti yet, we use this "mock" engine. It pretends to go
// out to the internet, waits a tiny bit, and returns fake but realistic data.
type mockEngine struct{}

// NewMockEngine creates and returns a new fake Trade Intelligence Engine.
// Simple meaning: Think of this as turning on our fake data machine. We will
// wire this up in main.go so the rest of the application doesn't even know
// it's talking to a fake engine. It just sees the "Engine" interface.
func NewMockEngine() Engine {
	return &mockEngine{}
}

// EnrichCompany pretends to look up a company and returns simulated data.
// Simple meaning: When given a company name (e.g., "Tech Logistics"), it simulates
// a network delay (like it's actually searching the web) and then returns some
// hardcoded fake data depending on keywords in the name.
func (m *mockEngine) EnrichCompany(ctx context.Context, companyName string) (*CompanyIntel, error) {
	// Simulate the time it takes to call an external API over the internet (e.g., 500 milliseconds)
	time.Sleep(500 * time.Millisecond)

	// Create a default response box to fill out.
	intel := &CompanyIntel{
		Name:                  companyName,
		Domain:                strings.ToLower(strings.ReplaceAll(companyName, " ", "")) + ".com",
		Industry:              "General Trading",
		EstimatedRevenue:      "$1M - $5M",
		EmployeeCount:         25,
		MonthlyShippingVolume: 10,
		TopSuppliers:          []string{"Unknown Supplier A", "Unknown Supplier B"},
		IsExporter:            false,
	}

	// Make the fake data a little smarter based on keywords in the company name.
	nameLower := strings.ToLower(companyName)

	if strings.Contains(nameLower, "tech") {
		intel.Industry = "Technology"
		intel.EstimatedRevenue = "$10M - $50M"
		intel.EmployeeCount = 200
		intel.MonthlyShippingVolume = 5
		intel.TopSuppliers = []string{"Foxconn", "TSMC", "Intel"}
	} else if strings.Contains(nameLower, "logistics") || strings.Contains(nameLower, "freight") {
		intel.Industry = "Logistics & Supply Chain"
		intel.EstimatedRevenue = "$50M - $100M"
		intel.EmployeeCount = 500
		intel.MonthlyShippingVolume = 1000
		intel.TopSuppliers = []string{"Maersk", "MSC", "CMA CGM"}
		intel.IsExporter = true
	} else if strings.Contains(nameLower, "export") || strings.Contains(nameLower, "trade") {
		intel.Industry = "International Trade"
		intel.EstimatedRevenue = "$5M - $20M"
		intel.EmployeeCount = 50
		intel.MonthlyShippingVolume = 150
		intel.TopSuppliers = []string{"Global Farms Ltd", "Steel Corp"}
		intel.IsExporter = true
	}

	// Return the filled-out fake data box, with no errors.
	return intel, nil
}
