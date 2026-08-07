package carrier

import "context"

// CarrierRate represents a single pricing option returned by a carrier API.
type CarrierRate struct {
	CarrierName           string
	BuyPrice              float64
	TransitDays           int
	ReliabilityScore      int
	HistoricalSuccessRate float64
	// FreeDays is the number of days a customer can store the container at the destination port
	// without incurring demurrage/detention penalties. E.g., 7, 14 days.
	FreeDays              int
	// Shipping line operational details from the rate sheet
	VesselName            string  // E.g., "CMA CGM CASSIOPEIA"
	ServiceCode           string  // E.g., "AS1" (service code/route tag)
	ViaPort               string  // Optional transshipment port (e.g., "SINGAPORE, SG")
	CO2Emissions          float64 // CO2e tonnes per TEU (e.g., 3.67)
	NauticalMiles         int     // Total nautical miles (e.g., 4168)
	// Price itemisation breakdown
	OceanFreight          float64 // Base ocean rate (e.g., 2800)
	OriginCharges         float64 // Export handling fees (e.g., terminal handling THC)
	DestinationCharges    float64 // Import handling and local terminal fees at dest
}

// CarrierProvider is the interface for fetching rates from carriers.
// In the future, this will be implemented by real MSC, Maersk, ONE API integrations.
type CarrierProvider interface {
	// GetRates retrieves carrier rates for a given lane. It takes logistics details like
	// Incoterms (DDP/FOB), Gross Weight (kg), Volume (CBM), and Commodity to simulate
	// a real shipping line quotation query.
	GetRates(ctx context.Context, origin, destination string, incoterms string, grossWeight float64, volumeCBM float64, commodity string) ([]CarrierRate, error)
}
