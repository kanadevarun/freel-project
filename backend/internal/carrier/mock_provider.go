package carrier

import (
	"context"
	"math/rand"
	"time"
)

// MockProvider implements CarrierProvider with simulated data.
type MockProvider struct{}

func NewMockProvider() CarrierProvider {
	return &MockProvider{}
}

func (m *MockProvider) GetRates(ctx context.Context, origin, destination string) ([]CarrierRate, error) {
	// Simulate API latency
	time.Sleep(2 * time.Second)

	// In a real scenario, this would call external APIs based on origin/dest.
	// Here we generate plausible mock data.
	
	basePrice := 1500.0 + rand.Float64()*1000.0 // random base price between $1500 and $2500

	rates := []CarrierRate{
		{
			CarrierName:           "MSC",
			BuyPrice:              basePrice,
			TransitDays:           28,
			ReliabilityScore:      85,
			HistoricalSuccessRate: 92.5,
		},
		{
			CarrierName:           "Maersk",
			BuyPrice:              basePrice + 300, // Premium but faster
			TransitDays:           22,
			ReliabilityScore:      95,
			HistoricalSuccessRate: 98.0,
		},
		{
			CarrierName:           "CMA CGM",
			BuyPrice:              basePrice - 150, // Cheaper but slower
			TransitDays:           35,
			ReliabilityScore:      70,
			HistoricalSuccessRate: 85.0,
		},
	}

	return rates, nil
}
