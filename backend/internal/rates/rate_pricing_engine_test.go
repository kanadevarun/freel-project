package rates

import (
	"testing"

	"github.com/freel/backend/internal/rates/spec"
)

func TestCalculateRateChargeItem(t *testing.T) {
	minVal := 50.0
	maxVal := 200.0

	tests := []struct {
		name           string
		baseRate       float64
		item           spec.RateChargeItem
		expectedAmount float64
	}{
		{
			name:     "Flat charge",
			baseRate: 2000.0,
			item: spec.RateChargeItem{
				CalculationBasis: CalculationBasisFlat,
				UnitPrice:        150.0,
			},
			expectedAmount: 150.0,
		},
		{
			name:     "Per container with quantity 2",
			baseRate: 2000.0,
			item: spec.RateChargeItem{
				CalculationBasis: CalculationBasisPerContainer,
				Quantity:         2,
				UnitPrice:        350.0,
			},
			expectedAmount: 700.0,
		},
		{
			name:     "Percentage charge 5% on 2000 base rate",
			baseRate: 2000.0,
			item: spec.RateChargeItem{
				CalculationBasis: CalculationBasisPercentage,
				UnitPrice:        5.0, // 5%
			},
			expectedAmount: 100.0,
		},
		{
			name:     "Included in base rate returns 0 additional",
			baseRate: 2000.0,
			item: spec.RateChargeItem{
				CalculationBasis:   CalculationBasisFlat,
				UnitPrice:          150.0,
				IncludedInBaseRate: true,
			},
			expectedAmount: 0.0,
		},
		{
			name:     "Zero quantity results in 0",
			baseRate: 2000.0,
			item: spec.RateChargeItem{
				CalculationBasis: CalculationBasisPerContainer,
				Quantity:         0,
				UnitPrice:        350.0,
			},
			expectedAmount: 0.0,
		},
		{
			name:     "Minimum amount enforcement",
			baseRate: 2000.0,
			item: spec.RateChargeItem{
				CalculationBasis: CalculationBasisPerUnit,
				Quantity:         1,
				UnitPrice:        30.0,
				MinimumAmount:    &minVal, // min 50.0
			},
			expectedAmount: 50.0,
		},
		{
			name:     "Maximum amount enforcement",
			baseRate: 2000.0,
			item: spec.RateChargeItem{
				CalculationBasis: CalculationBasisPerContainer,
				Quantity:         5,
				UnitPrice:        100.0, // 500.0
				MaximumAmount:    &maxVal, // max 200.0
			},
			expectedAmount: 200.0,
		},
		{
			name:     "Negative price protected to zero",
			baseRate: 2000.0,
			item: spec.RateChargeItem{
				CalculationBasis: CalculationBasisFlat,
				UnitPrice:        -50.0,
			},
			expectedAmount: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateRateChargeItem(tt.baseRate, tt.item)
			if got != tt.expectedAmount {
				t.Errorf("CalculateRateChargeItem() = %v, want %v", got, tt.expectedAmount)
			}
		})
	}
}

func TestCalculateRatePricing(t *testing.T) {
	charges := []spec.RateChargeItem{
		{
			ChargeCategory:   ChargeCategoryOrigin,
			ChargeName:       "Origin THC",
			CalculationBasis: CalculationBasisPerContainer,
			Quantity:         1,
			UnitPrice:        200.0,
			Currency:         "USD",
		},
		{
			ChargeCategory:   ChargeCategoryDocumentation,
			ChargeName:       "Bill of Lading",
			CalculationBasis: CalculationBasisFlat,
			Quantity:         1,
			UnitPrice:        75.0,
			Currency:         "USD",
		},
		{
			ChargeCategory:   ChargeCategorySurcharge,
			ChargeName:       "Bunker Adjustment Factor",
			CalculationBasis: CalculationBasisPercentage,
			UnitPrice:        10.0, // 10% of 2500 = 250
			Currency:         "USD",
		},
	}

	pricing := CalculateRatePricing(101, "RATE-2026-001", 2500.0, "USD", charges)

	if pricing.BaseRate != 2500.0 {
		t.Errorf("expected BaseRate 2500.0, got %v", pricing.BaseRate)
	}
	if pricing.AdditionalCharges != 525.0 { // 200 + 75 + 250
		t.Errorf("expected AdditionalCharges 525.0, got %v", pricing.AdditionalCharges)
	}
	if pricing.CommercialTotal != 3025.0 { // 2500 + 525
		t.Errorf("expected CommercialTotal 3025.0, got %v", pricing.CommercialTotal)
	}
	if pricing.IsMultiCurrency {
		t.Errorf("expected IsMultiCurrency false, got true")
	}
	if len(pricing.CategoryTotals) != 3 {
		t.Errorf("expected 3 CategoryTotals, got %d", len(pricing.CategoryTotals))
	}
}

func TestCalculateRatePricing_MultiCurrency(t *testing.T) {
	charges := []spec.RateChargeItem{
		{
			ChargeCategory:   ChargeCategoryOrigin,
			ChargeName:       "Origin Local Drayage",
			CalculationBasis: CalculationBasisFlat,
			UnitPrice:        15000.0,
			Currency:         "INR",
		},
		{
			ChargeCategory:   ChargeCategoryDocumentation,
			ChargeName:       "Export Filing",
			CalculationBasis: CalculationBasisFlat,
			UnitPrice:        50.0,
			Currency:         "USD",
		},
	}

	pricing := CalculateRatePricing(102, "RATE-2026-002", 2000.0, "USD", charges)

	if !pricing.IsMultiCurrency {
		t.Errorf("expected IsMultiCurrency true for mixed USD and INR")
	}
	if pricing.CurrencyBreakdown["USD"] != 2050.0 {
		t.Errorf("expected USD subtotal 2050.0, got %v", pricing.CurrencyBreakdown["USD"])
	}
	if pricing.CurrencyBreakdown["INR"] != 15000.0 {
		t.Errorf("expected INR subtotal 15000.0, got %v", pricing.CurrencyBreakdown["INR"])
	}
}
