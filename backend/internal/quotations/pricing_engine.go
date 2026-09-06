package quotations

import (
	"fmt"
	"math"
	"strings"
)

// round2 rounds a float64 to 2 decimal places for financial calculations.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// round4 rounds a float64 to 4 decimal places (for percentages/rates).
func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}

// CalculateChargeItem executes the pricing calculation for a single line item.
// It computes sell_amount, discount_amount, tax_amount, total_cost, and total_sell.
// This is a pure function without external or circular dependencies.
func CalculateChargeItem(item *QuotationChargeItem) error {
	if item == nil {
		return fmt.Errorf("charge item is nil")
	}

	if item.Quantity <= 0 {
		item.Quantity = 1.0
	}
	if item.ExchangeRate <= 0 {
		item.ExchangeRate = 1.0
	}

	// 1. Calculate Base Sell Amount
	var baseSell float64
	if strings.ToUpper(item.CalculationBasis) == QuotationChargeBasisPercentage {
		// When basis is PERCENTAGE, UnitPrice represents the percentage rate (e.g., 2.5 for 2.5%).
		// Quantity represents the base amount (e.g. declared cargo value).
		baseSell = item.Quantity * (item.UnitPrice / 100.0)
	} else {
		baseSell = item.Quantity * item.UnitPrice
	}
	baseSell = round2(baseSell)

	// 2. Calculate Discount Amount
	var discountAmt float64
	switch strings.ToUpper(item.DiscountType) {
	case QuotationDiscountTypePercentage:
		if item.DiscountValue > 0 {
			discountAmt = baseSell * (item.DiscountValue / 100.0)
		}
	case QuotationDiscountTypeFixed:
		discountAmt = item.DiscountValue
	default:
		discountAmt = 0.0
	}
	discountAmt = round2(discountAmt)

	// Discount cannot exceed baseSell
	if discountAmt > baseSell && baseSell > 0 {
		discountAmt = baseSell
	}
	if discountAmt < 0 {
		discountAmt = 0.0
	}
	item.DiscountAmount = discountAmt

	// 3. Sell Amount after discount, before tax
	sellAfterDiscount := round2(baseSell - discountAmt)
	if sellAfterDiscount < 0 {
		sellAfterDiscount = 0.0
	}
	item.SellAmount = sellAfterDiscount

	// 4. Calculate Tax Amount
	var taxAmt float64
	if item.TaxRate > 0 {
		taxAmt = round2(sellAfterDiscount * (item.TaxRate / 100.0))
	}
	item.TaxAmount = taxAmt

	// 5. Total Sell (Final sell amount including tax)
	item.TotalSell = round2(sellAfterDiscount + taxAmt)

	// 6. Total Cost (Unit cost * Quantity)
	item.TotalCost = round2(item.Quantity * item.CostAmount)

	return nil
}

// CalculateQuotationPricing aggregates all line items into a unified financial summary.
// It is the single centralized source of truth for quotation commercial calculations.
func CalculateQuotationPricing(quotation *Quotation, charges []*QuotationChargeItem) (*QuotationPricingSummary, error) {
	quoteCurrency := "USD"
	if quotation != nil && quotation.Currency != "" {
		quoteCurrency = quotation.Currency
	}

	summary := &QuotationPricingSummary{
		Currency:     quoteCurrency,
		MarginHealth: MarginHealthHealthy,
		ItemCount:    len(charges),
	}

	if len(charges) == 0 {
		return summary, nil
	}

	var (
		freightAmount float64
		originCharges float64
		destCharges   float64
		surcharges    float64
		docCharges    float64
		customsCharges float64
		insCharges    float64
		otherCharges  float64
		taxTotal      float64
		discountTotal float64
		subtotal      float64
		totalAmount   float64
		totalCost     float64
		multiCurrency bool
	)

	for _, item := range charges {
		if item == nil {
			continue
		}

		// Calculate item fields
		_ = CalculateChargeItem(item)

		// Multi-currency check
		if item.Currency != "" && !strings.EqualFold(item.Currency, quoteCurrency) {
			multiCurrency = true
		}

		// If charge is optional, it is an add-on and does not inflate base quotation total
		if item.IsOptional {
			continue
		}

		// Bucket by category
		cat := strings.ToUpper(item.ChargeCategory)
		switch cat {
		case QuotationChargeCategoryFreight:
			freightAmount += item.SellAmount
		case QuotationChargeCategoryOrigin:
			originCharges += item.SellAmount
		case QuotationChargeCategoryDestination:
			destCharges += item.SellAmount
		case QuotationChargeCategorySurcharge:
			surcharges += item.SellAmount
		case QuotationChargeCategoryDocumentation:
			docCharges += item.SellAmount
		case QuotationChargeCategoryCustoms:
			customsCharges += item.SellAmount
		case QuotationChargeCategoryInsurance:
			insCharges += item.SellAmount
		default:
			otherCharges += item.SellAmount
		}

		discountTotal += item.DiscountAmount
		taxTotal += item.TaxAmount
		subtotal += item.SellAmount
		totalAmount += item.TotalSell
		totalCost += item.TotalCost
	}

	summary.FreightAmount = round2(freightAmount)
	summary.OriginCharges = round2(originCharges)
	summary.DestinationCharges = round2(destCharges)
	summary.Surcharges = round2(surcharges)
	summary.DocumentationCharges = round2(docCharges)
	summary.CustomsCharges = round2(customsCharges)
	summary.InsuranceCharges = round2(insCharges)
	summary.OtherCharges = round2(otherCharges)
	summary.Discounts = round2(discountTotal)
	summary.Taxes = round2(taxTotal)
	summary.Subtotal = round2(subtotal)
	summary.TotalAmount = round2(totalAmount)
	summary.TotalCost = round2(totalCost)
	summary.MultiCurrencyWarning = multiCurrency

	// Margin Calculations
	grossProfit := round2(totalAmount - totalCost)
	summary.GrossProfit = grossProfit

	var marginPct float64
	if totalAmount > 0 {
		marginPct = round4((grossProfit / totalAmount) * 100.0)
	}
	summary.GrossMarginPercentage = marginPct

	// Determine Margin Health
	if totalCost == 0 && totalAmount >= 0 {
		summary.MarginHealth = MarginHealthHealthy
	} else if marginPct >= 15.0 {
		summary.MarginHealth = MarginHealthHealthy
	} else if marginPct >= 0.0 {
		summary.MarginHealth = MarginHealthLow
	} else {
		summary.MarginHealth = MarginHealthNegative
	}

	return summary, nil
}
