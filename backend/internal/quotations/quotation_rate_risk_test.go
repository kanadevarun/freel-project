package quotations

import (
	"testing"
	"time"
)

func TestCommercialImpactAnalysis(t *testing.T) {
	// Test commercial impact calculation logic
	currentSnapshot := &QuotationRateSnapshot{
		ID:              1,
		CarrierName:     "MSC",
		Currency:        "USD",
		CommercialTotal: 2500.00,
	}

	// 1. Cheaper candidate
	cheaperCand := &QuotationRateCandidate{
		CarrierName:     "CMA CGM",
		Currency:        "USD",
		CommercialTotal: 2200.00,
		TransitDays:     20,
	}

	diffAmount := cheaperCand.CommercialTotal - currentSnapshot.CommercialTotal
	diffPct := (diffAmount / currentSnapshot.CommercialTotal) * 100.0

	if diffAmount != -300.00 {
		t.Errorf("expected price diff of -300.00, got %.2f", diffAmount)
	}
	if diffPct >= 0 || diffPct < -15.0 {
		t.Errorf("expected price pct around -12.0%%, got %.2f%%", diffPct)
	}

	// 2. Cross-currency candidate
	eurCand := &QuotationRateCandidate{
		CarrierName:     "Hapag-Lloyd",
		Currency:        "EUR",
		CommercialTotal: 2400.00,
	}

	if eurCand.Currency == currentSnapshot.Currency {
		t.Errorf("expected currency mismatch between USD and EUR")
	}
}

func TestQuotationRiskStructure(t *testing.T) {
	now := time.Now()
	risk := &QuotationRateRisk{
		OrgID:             1,
		QuotationID:       10,
		RiskType:          "RATE_EXPIRING_SOON",
		Severity:          "WARNING",
		Headline:          "Carrier rate expires in 5 days",
		Description:       "Underlying rate expiring soon.",
		RecommendedAction: "Review replacement options.",
		CreatedAt:         now,
	}

	if risk.RiskType != "RATE_EXPIRING_SOON" || risk.Severity != "WARNING" {
		t.Errorf("unexpected risk type or severity: %s / %s", risk.RiskType, risk.Severity)
	}
}
