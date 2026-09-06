package rates

import (
	"math"

	"github.com/freel/backend/internal/rates/spec"
)

// CalculateRateChargeItem deterministically calculates the line amount for a given RateChargeItem
func CalculateRateChargeItem(baseRate float64, item spec.RateChargeItem) float64 {
	// If charge is already included in the base rate, its marginal additional cost to add to the base rate is 0
	if item.IncludedInBaseRate {
		return 0.0
	}

	var amount float64

	switch item.CalculationBasis {
	case CalculationBasisFlat:
		amount = item.UnitPrice
	case CalculationBasisPercentage:
		// Percent calculated against base rate
		if baseRate > 0 {
			amount = (item.UnitPrice / 100.0) * baseRate
		} else {
			amount = 0.0
		}
	case CalculationBasisPerContainer,
		CalculationBasisPerShipment,
		CalculationBasisPerWeight,
		CalculationBasisPerVolume,
		CalculationBasisPerUnit:
		// Quantity x UnitPrice
		if item.Quantity > 0 {
			amount = item.Quantity * item.UnitPrice
		} else {
			amount = 0.0
		}
	default:
		// Fallback to unit price if unrecognized basis
		amount = item.UnitPrice
	}

	// Prevent negative calculated totals
	if amount < 0 {
		amount = 0.0
	}

	// Apply minimum threshold constraint if defined
	if item.MinimumAmount != nil && *item.MinimumAmount > 0 {
		if amount < *item.MinimumAmount {
			amount = *item.MinimumAmount
		}
	}

	// Apply maximum ceiling constraint if defined
	if item.MaximumAmount != nil && *item.MaximumAmount > 0 {
		if amount > *item.MaximumAmount {
			amount = *item.MaximumAmount
		}
	}

	return roundFloat(amount, 2)
}

// CalculateRatePricing deterministically computes the full RatePricingSummary
func CalculateRatePricing(
	rateID int64,
	rateRef string,
	baseRate float64,
	baseCurrency string,
	charges []spec.RateChargeItem,
) spec.RatePricingSummary {
	if baseCurrency == "" {
		baseCurrency = "USD"
	}
	if baseRate < 0 {
		baseRate = 0.0
	}
	baseRate = roundFloat(baseRate, 2)

	// Currency tracking
	currenciesSeen := make(map[string]bool)
	currenciesSeen[baseCurrency] = true
	currencyBreakdown := make(map[string]float64)
	currencyBreakdown[baseCurrency] = baseRate

	// Category tracking
	categoryMap := make(map[string]*spec.CategoryTotal)

	var additionalChargesTotal float64
	evaluatedCharges := make([]spec.RateChargeItem, len(charges))

	for i, ch := range charges {
		curr := ch.Currency
		if curr == "" {
			curr = baseCurrency
			ch.Currency = curr
		}
		currenciesSeen[curr] = true

		// Calculate charge line amount
		calcAmount := CalculateRateChargeItem(baseRate, ch)
		ch.CalculatedAmount = calcAmount
		evaluatedCharges[i] = ch

		// Currency breakdown sum
		currencyBreakdown[curr] += calcAmount

		// If same currency as base rate, contribute to additionalChargesTotal
		if curr == baseCurrency {
			additionalChargesTotal += calcAmount
		}

		// Category subtotal aggregation
		catKey := ch.ChargeCategory
		if catKey == "" {
			catKey = ChargeCategoryOther
		}
		if _, exists := categoryMap[catKey]; !exists {
			categoryMap[catKey] = &spec.CategoryTotal{
				Category:    catKey,
				ChargeCount: 0,
				TotalAmount: 0,
				Currency:    curr,
			}
		}
		categoryMap[catKey].ChargeCount++
		categoryMap[catKey].TotalAmount += calcAmount
	}

	// Build sorted currencies list
	currenciesList := make([]string, 0, len(currenciesSeen))
	for c := range currenciesSeen {
		currenciesList = append(currenciesList, c)
	}

	isMultiCurrency := len(currenciesSeen) > 1

	// Category totals slice
	categoryTotals := make([]spec.CategoryTotal, 0, len(categoryMap))
	for _, ct := range categoryMap {
		ct.TotalAmount = roundFloat(ct.TotalAmount, 2)
		categoryTotals = append(categoryTotals, *ct)
	}

	// Commercial total
	var commercialTotal float64
	if !isMultiCurrency {
		commercialTotal = roundFloat(baseRate+additionalChargesTotal, 2)
	} else {
		// When multiple currencies exist, CommercialTotal represents base currency aggregate
		commercialTotal = roundFloat(baseRate+additionalChargesTotal, 2)
	}

	return spec.RatePricingSummary{
		RateID:            rateID,
		RateReference:     rateRef,
		BaseRate:          baseRate,
		BaseCurrency:      baseCurrency,
		AdditionalCharges: roundFloat(additionalChargesTotal, 2),
		CommercialTotal:   commercialTotal,
		ChargeCount:       len(charges),
		IsMultiCurrency:   isMultiCurrency,
		Currencies:        currenciesList,
		CurrencyBreakdown: currencyBreakdown,
		CategoryTotals:    categoryTotals,
		Charges:           evaluatedCharges,
	}
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
