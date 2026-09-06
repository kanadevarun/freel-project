package rates

import (
	"fmt"
	"time"

	"github.com/freel/backend/internal/carrier"
	"github.com/freel/backend/internal/rates/spec"
	"github.com/google/uuid"
)

// carrierSCACMap maps the CarrierName strings used by the existing carrier package
// to their standard SCAC codes stored in the carriers table.
var carrierSCACMap = map[string]string{
	"Maersk":      "MAEU",
	"MSC":         "MSCU",
	"CMA CGM":     "CMDU",
	"ONE":         "ONEY",
	"Hapag-Lloyd": "HLCU",
	"Evergreen":   "EGLV",
	"COSCO":       "COSU",
	"ZIM":         "ZIMU",
	"Yang Ming":   "YMLU",
	"HMM":         "HDMU",
}

// spotRateValidityHours is how long a spot rate from a carrier API is considered
// fresh before the search falls back to a new fetch. Spot rates are real-time
// market prices and become stale quickly; 4 hours balances API cost vs. accuracy.
const spotRateValidityHours = 4

// SpotNormalizer converts a carrier.RichCarrierRate into a spec.CanonicalRate.
type SpotNormalizer interface {
	Normalize(r carrier.RichCarrierRate, orgID int64, originPort, destinationPort string) spec.CanonicalRate
}

type spotNormalizer struct{}

// NewSpotNormalizer creates a new SpotNormalizer.
func NewSpotNormalizer() SpotNormalizer {
	return &spotNormalizer{}
}

// Normalize converts a carrier.RichCarrierRate into a CanonicalRate.
//
// Key transformations:
//   - Port names are normalized via NormalizePort() → UN/LOCODE.
//   - SCAC code is looked up in carrierSCACMap; falls back to "UNKN".
//   - total_buy_price = ocean_freight + origin_charges + destination_charges.
//     (Spot rates from the mock / real provider don't have separate surcharge
//     line items yet; they are fully rolled into the three price fields.)
//   - valid_from = now(); valid_until = now() + 4 hours.
//     Spot rates expire quickly; the search will trigger a fresh fetch if stale.
//   - confidence_score = 100 (API data is authoritative).
//   - extraction_status = CONFIRMED (no human review needed for API data).
func (n *spotNormalizer) Normalize(r carrier.RichCarrierRate, orgID int64, originPort, destinationPort string) spec.CanonicalRate {
	now := time.Now().UTC()

	scac, ok := carrierSCACMap[r.CarrierName]
	if !ok {
		scac = "UNKN"
	}

	// Normalize port representations to UN/LOCODE
	origin := NormalizePort(originPort)
	dest := NormalizePort(destinationPort)

	total := r.OceanFreight + r.OriginCharges + r.DestinationCharges
	surcharges := buildSpotSurcharges(r)
	equipmentType := "40GP"
	transitDays := r.TransitDays

	return spec.CanonicalRate{
		ID:              uuid.NewString(),
		OrgID:           orgID,
		Source:          RateSourceSpotAPI,
		SourceRef:       fmt.Sprintf("carrier-api:%s", scac),
		OriginPort:      origin,
		DestinationPort: dest,
		ViaPort:         NormalizePort(r.ViaPort),
		ServiceCode:     r.ServiceCode,
		CarrierSCAC:     scac,
		CarrierName:     r.CarrierName,
		VesselName:      r.VesselName,
		EquipmentType:   equipmentType,

		OceanFreight:       r.OceanFreight,
		OriginCharges:      r.OriginCharges,
		DestinationCharges: r.DestinationCharges,
		Surcharges:         surcharges,
		TotalBuyPrice:      total,
		CurrencyOriginal:   "USD",
		ExchangeRateUsed:   1.0,

		IncludedCharges: []string{},
		ExcludedCharges: []string{"Customs", "Inland Haulage", "D&D"},

		FreeDaysOrigin:      0,
		FreeDaysDestination: r.FreeDays,
		TransitDays:         &transitDays,
		Incoterms:           "",
		CommodityRestrictions: []string{},
		RoutingConditions:   "",

		ValidFrom:  now,
		ValidUntil: now.Add(spotRateValidityHours * time.Hour),

		ConfidenceScore:  100, // API data = authoritative
		ExtractionStatus: ExtractionStatusConfirmed,
		ExtractedBy:      "spot-api",

		NauticalMiles: r.NauticalMiles,
		CO2PerTEU:     r.CO2Emissions,

		CreatedAt: now,
		UpdatedAt: now,
	}
}

// buildSpotSurcharges converts the origin/destination charge breakdown from a
// carrier.RichCarrierRate into named Surcharge items.
func buildSpotSurcharges(r carrier.RichCarrierRate) []spec.Surcharge {
	var surcharges []spec.Surcharge
	if r.OriginCharges > 0 {
		surcharges = append(surcharges, spec.Surcharge{
			Code:        "OHC",
			Description: "Origin Handling Charge",
			Amount:      r.OriginCharges,
			Unit:        SurchargeUnitPerContainer,
			Included:    false,
		})
	}
	if r.DestinationCharges > 0 {
		surcharges = append(surcharges, spec.Surcharge{
			Code:        "DHC",
			Description: "Destination Handling Charge",
			Amount:      r.DestinationCharges,
			Unit:        SurchargeUnitPerContainer,
			Included:    false,
		})
	}
	return surcharges
}
