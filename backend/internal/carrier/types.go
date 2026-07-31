package carrier

import "context"

// CarrierRate represents a single pricing option returned by a carrier API.
type CarrierRate struct {
	CarrierName           string
	BuyPrice              float64
	TransitDays           int
	ReliabilityScore      int
	HistoricalSuccessRate float64
}

// CarrierProvider is the interface for fetching rates from carriers.
// In the future, this will be implemented by real MSC, Maersk, ONE API integrations.
type CarrierProvider interface {
	GetRates(ctx context.Context, origin, destination string) ([]CarrierRate, error)
}
