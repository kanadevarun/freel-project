package carrier

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// ──────────────────────────────────────────────────────────────────────────────
// RichCarrierRate extends CarrierRate with derived fields used by the
// frontend Pricing Workspace (recommended flag, AI reasoning, deadline check).
// ──────────────────────────────────────────────────────────────────────────────
type RichCarrierRate struct {
	CarrierName string  `json:"carrier_name"`
	BuyPrice    float64 `json:"buy_price"`
	// SellPrice is BuyPrice + the default margin (20%).
	// The pricing team can override this in the UI.
	SellPrice             float64 `json:"sell_price"`
	MarginPct             float64 `json:"margin_pct"`
	TransitDays           int     `json:"transit_days"`
	ReliabilityScore      int     `json:"reliability_score"`
	HistoricalSuccessRate float64 `json:"historical_success_rate"`
	// IsRecommended is true for the single carrier AI considers best.
	IsRecommended bool `json:"is_recommended"`
	// AIReasoning is a short human-readable explanation for IsRecommended.
	AIReasoning string `json:"ai_reasoning"`
	// MeetsDeadline is true when TransitDays fits before TargetDate.
	MeetsDeadline bool `json:"meets_deadline"`
	// DeadlineStatus is one of "on_time", "borderline", "missed".
	DeadlineStatus string `json:"deadline_status"`
	// FreeDays is the number of port demurrage/detention free days.
	FreeDays int `json:"free_days"`
	// Operational parameters
	VesselName         string  `json:"vessel_name"`
	ServiceCode        string  `json:"service_code"`
	ViaPort            string  `json:"via_port"` // Optional transshipment port
	CO2Emissions       float64 `json:"co2_emissions"`
	NauticalMiles      int     `json:"nautical_miles"`
	// Fee itemisation breakdowns
	OceanFreight       float64 `json:"ocean_freight"`
	OriginCharges      float64 `json:"origin_charges"`
	DestinationCharges float64 `json:"destination_charges"`
	// FetchedAt records when the rate was fetched, so the UI can show "valid for X hours".
	FetchedAt string `json:"fetched_at"`
}

// ──────────────────────────────────────────────────────────────────────────────
// FetchRatesResponse is the full response returned by the carrier service.
// ──────────────────────────────────────────────────────────────────────────────
type FetchRatesResponse struct {
	Rates            []RichCarrierRate `json:"rates"`
	OverallReasoning string            `json:"overall_reasoning"`
	// RecommendedIdx is the index into Rates of the recommended carrier.
	RecommendedIdx int    `json:"recommended_idx"`
	FetchedAt      string `json:"fetched_at"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Service is the business-logic interface for the carrier package.
// The HTTP handler calls FetchRates; in future sprints we can add
// CacheRates, RefreshRates, etc. without changing the interface.
// ──────────────────────────────────────────────────────────────────────────────
type Service interface {
	// FetchRates retrieves carrier rates for a given trade lane.
	// It applies AI ranking logic and returns a ranked, enriched slice.
	//
	// Simple meaning: You tell us where the cargo is going; we find all the
	// shipping lines that operate that lane, compare them, and tell you
	// which one to use and why.
	//
	// Example:
	//   resp, err := svc.FetchRates(ctx, "INNSA", "DEHAM", nil, "DDP", 257.0, 0.0, "JEWELLERY MACHINE AND PARTS")
	FetchRates(ctx context.Context, origin, destination string, targetDate *time.Time, incoterms string, grossWeight float64, volumeCBM float64, commodity string) (*FetchRatesResponse, error)
}

type service struct {
	provider CarrierProvider
}

// NewService creates a new carrier Service backed by the given provider.
// In production this will be the live FF partner API adapter;
// in development / in-house testing we use MockProvider.
func NewService(provider CarrierProvider) Service {
	return &service{provider: provider}
}

// FetchRates fetches, enriches, ranks, and returns carrier options.
func (s *service) FetchRates(ctx context.Context, origin, destination string, targetDate *time.Time, incoterms string, grossWeight float64, volumeCBM float64, commodity string) (*FetchRatesResponse, error) {
	// 1. Fetch raw rates from the provider (FF partner API or mock).
	raw, err := s.provider.GetRates(ctx, origin, destination, incoterms, grossWeight, volumeCBM, commodity)
	if err != nil {
		return nil, fmt.Errorf("carrier provider: %w", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("no carrier rates available for %s → %s", origin, destination)
	}

	// 2. Enrich each rate with sell price, margin, and deadline check.
	//    Default margin is 20%; the pricing team overrides it in the UI.
	const defaultMarginPct = 20.0
	now := time.Now().UTC()
	fetchedAt := now.Format(time.RFC3339)

	rich := make([]RichCarrierRate, len(raw))
	// Loop through each raw carrier rate returned by our provider (e.g. MSC, Maersk)
	// and transform it into a "rich" rate containing calculated fields for the UI.
	//
	// In each iteration, we:
	//  1. Calculate the suggested Sell Price by applying our standard markup (e.g., $1,000 buy + 20% margin = $1,200 sell).
	//  2. Estimate the arrival date (ETA) by adding transit days to today's date.
	//  3. If a target deadline is set, compare the ETA against it and tag the option as "on_time", "borderline", or "missed".
	for i, r := range raw {
		sellPrice := r.BuyPrice * (1 + defaultMarginPct/100)
		marginAmt := sellPrice - r.BuyPrice

		// Deadline check: assume cargo departs today. Add TransitDays and compare to targetDate.
		eta := now.AddDate(0, 0, r.TransitDays)
		meetsDeadline := true
		deadlineStatus := "on_time"
		if targetDate != nil {
			daysBuffer := targetDate.Sub(eta).Hours() / 24
			switch {
			case daysBuffer >= 0:
				deadlineStatus = "on_time"
			case daysBuffer >= -3:
				// Within 3 days late → borderline
				deadlineStatus = "borderline"
				meetsDeadline = false
			default:
				deadlineStatus = "missed"
				meetsDeadline = false
			}
		}
		// marginAmt represents the dollar-value markup applied to the carrier rate (SellPrice - BuyPrice).
		// In future sprints, this will be integrated with a Dynamic Margin Rule Engine.
		// Instead of applying a flat 20% margin, the engine will compute lane-specific,
		// customer-specific, and commodity-specific markups.
		//
		// Concrete Examples of future Margin Rules:
		//
		// 1. Lane-Specific Rules:
		//    - Highly competitive corridors (e.g., INNSA -> DEHAM) might use a tighter margin of 10-12%
		//      to remain price-competitive in the market.
		//    - Underutilized or complex corridors (e.g., INCCU -> USNYC) might use a higher margin of 20-25%.
		//
		// 2. Customer Volume/Tier Rules:
		//    - A high-volume enterprise customer (e.g., Tata Exports shipping >100 TEUs/month) could trigger a
		//      discount rule, reducing the margin to a flat 12% across all lanes.
		//    - Spot-market or new customers would start at a baseline 20% margin.
		//
		// 3. Equipment & Commodity Rules:
		//    - Hazardous cargo (HAZMAT) or temperature-controlled cargo (Reefers) require extra documentation
		//      and risk management. The rule engine would apply a flat surcharge (e.g., +$150) or a percentage
		//      bump (e.g., +5% margin) to the sell price.
		//
		// 4. Dynamic AI/Demand-based Rules:
		//    - If historical data shows a high likelihood of winning this lane at a higher price (based on AI win-rate analytics),
		//      or if there is an urgent timeline/capacity constraint, the AI could recommend a dynamically adjusted margin.
		_ = marginAmt

		rich[i] = RichCarrierRate{
			CarrierName:           r.CarrierName,
			BuyPrice:              r.BuyPrice,
			SellPrice:             sellPrice,
			MarginPct:             defaultMarginPct,
			TransitDays:           r.TransitDays,
			ReliabilityScore:      r.ReliabilityScore,
			HistoricalSuccessRate: r.HistoricalSuccessRate,
			MeetsDeadline:         meetsDeadline,
			DeadlineStatus:        deadlineStatus,
			FreeDays:              r.FreeDays,
			VesselName:            r.VesselName,
			ServiceCode:           r.ServiceCode,
			ViaPort:               r.ViaPort,
			CO2Emissions:          r.CO2Emissions,
			NauticalMiles:         r.NauticalMiles,
			OceanFreight:          r.OceanFreight,
			OriginCharges:         r.OriginCharges,
			DestinationCharges:    r.DestinationCharges,
			FetchedAt:             fetchedAt,
		}
	}

	// 3. Rank carriers using a weighted score:
	//    • Meeting the deadline is a hard gate (skips non-qualifying carriers first)
	//    • Among qualifying: reliability (50%) + cost efficiency (30%) + transit (20%)
	//
	// We score each carrier out of 100 using normalized values.
	maxReliability := 0
	minPrice := rich[0].BuyPrice
	maxPrice := rich[0].BuyPrice
	minTransit := rich[0].TransitDays
	maxTransit := rich[0].TransitDays

	for _, r := range rich {
		if r.ReliabilityScore > maxReliability {
			maxReliability = r.ReliabilityScore
		}
		if r.BuyPrice < minPrice {
			minPrice = r.BuyPrice
		}
		if r.BuyPrice > maxPrice {
			maxPrice = r.BuyPrice
		}
		if r.TransitDays < minTransit {
			minTransit = r.TransitDays
		}
		if r.TransitDays > maxTransit {
			maxTransit = r.TransitDays
		}
	}

	// scoreCarrier returns a composite 0-100 score. Higher is better.
	scoreCarrier := func(r RichCarrierRate) float64 {
		// Reliability: normalise to 0-1, weight 50%
		relScore := 0.0
		if maxReliability > 0 {
			relScore = float64(r.ReliabilityScore) / float64(maxReliability)
		}

		// Cost: cheaper = better; normalise inversely
		costScore := 0.0
		if maxPrice != minPrice {
			costScore = (maxPrice - r.BuyPrice) / (maxPrice - minPrice)
		}

		// Transit: faster = better; normalise inversely
		transitScore := 0.0
		if maxTransit != minTransit {
			transitScore = float64(maxTransit-r.TransitDays) / float64(maxTransit-minTransit)
		}

		// Deadline penalty: if a carrier misses the deadline, cut its score in half
		deadlinePenalty := 1.0
		if !r.MeetsDeadline {
			deadlinePenalty = 0.5
		}

		return (relScore*0.50 + costScore*0.30 + transitScore*0.20) * 100 * deadlinePenalty
	}

	// Sort by composite score descending
	sort.Slice(rich, func(i, j int) bool {
		return scoreCarrier(rich[i]) > scoreCarrier(rich[j])
	})

	// 4. Mark the top carrier as recommended and generate a short AI reasoning string.
	recommendedIdx := 0
	top := &rich[0]
	top.IsRecommended = true
	top.AIReasoning = buildReasoning(*top, origin, destination)

	// 5. Build the overall summary explanation shown at the top of Pricing Workspace.
	overallReasoning := fmt.Sprintf(
		"%s is recommended for %s → %s based on the highest composite score. "+
			"Reliability: %d/100, Transit: %d days, Buy rate: $%.2f/container. "+
			"Sell price at %d%% margin: $%.2f/container.",
		top.CarrierName, origin, destination,
		top.ReliabilityScore, top.TransitDays,
		top.BuyPrice, int(top.MarginPct), top.SellPrice,
	)

	return &FetchRatesResponse{
		Rates:            rich,
		OverallReasoning: overallReasoning,
		RecommendedIdx:   recommendedIdx,
		FetchedAt:        fetchedAt,
	}, nil
}

// buildReasoning produces a concise human-readable explanation for a recommended carrier.
func buildReasoning(r RichCarrierRate, origin, destination string) string {
	deadlineNote := ""
	switch r.DeadlineStatus {
	case "on_time":
		deadlineNote = "✓ meets customer deadline"
	case "borderline":
		deadlineNote = "⚠ borderline — ETA is very close to target date"
	case "missed":
		deadlineNote = "✗ misses customer deadline — recommend checking alternatives"
	}

	return fmt.Sprintf(
		"%s is recommended for the %s → %s trade lane. "+
			"Reliability score %d/100 (historical success rate %.1f%%), "+
			"transit time %d days, buy rate $%.2f/container. %s.",
		r.CarrierName, origin, destination,
		r.ReliabilityScore, r.HistoricalSuccessRate,
		r.TransitDays, r.BuyPrice, deadlineNote,
	)
}
