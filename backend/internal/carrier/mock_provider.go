package carrier

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Ensure the local math.go Round2 is used.
var _ = Round2

// ──────────────────────────────────────────────────────────────────────────────
// MockProvider implements CarrierProvider with simulated data.
//
// For in-house testing (no FF partner onboarded yet), this generates
// realistic carrier rates based on the trade lane.  Each lane has a set of
// carriers with sensible defaults; random jitter is applied to prices so
// repeated calls to the same lane feel realistic.
//
// When the FF partner REST API is ready, replace this with a
// FFPartnerProvider that calls their endpoint.
// ──────────────────────────────────────────────────────────────────────────────
type MockProvider struct {
	// rng is seeded once on creation so tests are deterministic when using
	// a fixed seed, but random in production.
	rng *rand.Rand
}

func NewMockProvider() CarrierProvider {
	return &MockProvider{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// laneProfile defines realistic base rates and carrier options for a trade lane.
type laneProfile struct {
	carriers []laneCarrier
}

type laneCarrier struct {
	name                  string
	baseBuyPrice          float64 // USD per container (20ft equivalent)
	transitDays           int
	reliabilityScore      int
	historicalSuccessRate float64
	freeDays              int // Demurrage/Detention free days at destination port
}

// tradeLanes maps a canonical "ORIGIN→DESTINATION" key to realistic carrier profiles.
// If no exact match, we fall back to a default set.
var tradeLanes = map[string]laneProfile{
	// India → Europe (Hamburg)
	"INNSA→DEHAM": {carriers: []laneCarrier{
		{"Hapag-Lloyd", 1350, 22, 92, 91.0, 14},
		{"MSC", 1200, 25, 85, 88.5, 7},
		{"Evergreen", 1100, 28, 78, 84.0, 14},
		{"CMA CGM", 1280, 24, 88, 89.0, 10},
	}},
	// India → UAE
	"INBOM→AEAUH": {carriers: []laneCarrier{
		{"MSC", 620, 7, 91, 93.0, 14},
		{"Maersk", 720, 6, 95, 97.0, 14},
		{"Hapag-Lloyd", 680, 7, 89, 91.5, 7},
	}},
	// India → USA East Coast
	"INCCU→USNYC": {carriers: []laneCarrier{
		{"Maersk", 2100, 28, 94, 96.0, 7},
		{"MSC", 1950, 32, 86, 90.0, 7},
		{"CMA CGM", 2000, 30, 88, 92.0, 14},
		{"ONE", 1900, 35, 82, 87.0, 14},
	}},
	// India → Singapore
	"INBLR→SGSIN": {carriers: []laneCarrier{
		{"MSC", 480, 12, 90, 94.0, 14},
		{"Maersk", 520, 11, 96, 98.0, 14},
		{"Hapag-Lloyd", 500, 12, 88, 92.0, 7},
	}},
	// India → China Shanghai
	"INCCU→CNSHA": {carriers: []laneCarrier{
		{"COSCO", 980, 18, 88, 92.0, 21},
		{"Evergreen", 920, 20, 82, 87.0, 14},
		{"ONE", 950, 19, 85, 90.0, 14},
		{"MSC", 1000, 17, 91, 93.0, 7},
	}},
}

// GetRates returns simulated carrier rates for the given origin and destination.
//
// Lookup order:
//  1. Exact match on "ORIGIN→DESTINATION"
//  2. Reverse lookup "DESTINATION→ORIGIN" (same lane, mirrored direction)
//  3. Default set of 3 carriers with generic prices
//
// A ±10% random jitter is applied to each price so the UI looks realistic.
func (m *MockProvider) GetRates(ctx context.Context, origin, destination string, incoterms string, grossWeight float64, volumeCBM float64, commodity string) ([]CarrierRate, error) {
	// Simulate network latency (150-400ms)
	delay := time.Duration(150+m.rng.Intn(250)) * time.Millisecond
	time.Sleep(delay)

	key := fmt.Sprintf("%s→%s", origin, destination)
	profile, ok := tradeLanes[key]
	if !ok {
		// Try reverse direction as fallback
		reverseKey := fmt.Sprintf("%s→%s", destination, origin)
		profile, ok = tradeLanes[reverseKey]
	}
	if !ok {
		// Default: use a generic set so we always return something
		profile = defaultLaneProfile(origin, destination, m.rng)
	}

	// Apply random jitter ±10% to make repeated calls feel real
	rates := make([]CarrierRate, len(profile.carriers))
	for i, c := range profile.carriers {
		jitter := 1.0 + (m.rng.Float64()-0.5)*0.20 // ±10%
		basePrice := c.baseBuyPrice * jitter

		// 1. Incoterms Adjustments (e.g. DDP / DAP requires final destination delivery surcharges)
		ddpSurcharge := 0.0
		if incoterms == "DDP" || incoterms == "DAP" {
			ddpSurcharge = 180.0 // Add final delivery delivery surcharge
		}

		// 2. Gross Weight Surcharges (heavy container surcharge for cargo > 15 tonnes)
		weightSurcharge := 0.0
		if grossWeight > 15000.0 {
			weightSurcharge = 150.0
		} else if grossWeight > 0.0 && grossWeight < 500.0 {
			// Very light cargo/LCL gets a slight discount on base ocean freight
			basePrice = basePrice * 0.9
		}

		// 3. Commodity Surcharges (e.g. Hazardous or High-Value Machinery security fee)
		commoditySurcharge := 0.0
		if commodity != "" {
			// Simulating extra inspections or security escorts
			commoditySurcharge = 95.0
		}

		buyPrice := Round2(basePrice + ddpSurcharge + weightSurcharge + commoditySurcharge, 2)

		// Detailed price itemisation (adds up to buyPrice exactly, showing surcharges in export/import)
		oceanFreight := Round2(basePrice*0.80, 2)
		originCharges := Round2(basePrice*0.12 + weightSurcharge + (commoditySurcharge * 0.5), 2)
		destCharges := Round2(buyPrice-oceanFreight-originCharges, 2)

		// Vessel name generation E.g. "MSC CASSIOPEIA"
		vessels := []string{"CASSIOPEIA", "NEPTUNE", "VALIANT", "DISCOVERY", "SOVEREIGN", "PACIFIC"}
		vesselName := fmt.Sprintf("%s %s", c.name, vessels[m.rng.Intn(len(vessels))])

		// Service Code (AS1, ME3, INDEX)
		services := []string{"AS1", "AS2", "ME3", "INDEX", "EU1"}
		serviceCode := services[m.rng.Intn(len(services))]

		// Optional transshipment via port (50% chance)
		viaPort := ""
		if m.rng.Float64() < 0.6 {
			vias := []string{"SINGAPORE, SG", "COLOMBO, LK", "PORT KELANG, MY", "JEBEL ALI, AE"}
			viaPort = fmt.Sprintf("via %s", vias[m.rng.Intn(len(vias))])
		}

		// Nautical miles (estimated based on price)
		nauticalMiles := int(buyPrice * 2.1)

		// CO2 Emissions per TEU (tonnes)
		co2 := Round2(1.8+m.rng.Float64()*2.5, 2)

		rates[i] = CarrierRate{
			CarrierName:           c.name,
			BuyPrice:              buyPrice,
			TransitDays:           c.transitDays + m.rng.Intn(3) - 1, // ±1 day
			ReliabilityScore:      c.reliabilityScore,
			HistoricalSuccessRate: c.historicalSuccessRate,
			FreeDays:              c.freeDays,
			VesselName:            vesselName,
			ServiceCode:           serviceCode,
			ViaPort:               viaPort,
			CO2Emissions:          co2,
			NauticalMiles:         nauticalMiles,
			OceanFreight:          oceanFreight,
			OriginCharges:         originCharges,
			DestinationCharges:    destCharges,
		}
	}

	return rates, nil
}


// defaultLaneProfile returns a generic 3-carrier set when no specific lane is configured.
func defaultLaneProfile(origin, destination string, rng *rand.Rand) laneProfile {
	basePrice := 1200.0 + rng.Float64()*800.0
	return laneProfile{carriers: []laneCarrier{
		{"MSC", basePrice, 28, 85, 90.0, 14},
		{"Maersk", basePrice + 200, 22, 95, 97.0, 14},
		{"CMA CGM", basePrice - 100, 32, 78, 85.0, 7},
	}}
}
