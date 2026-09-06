package quotations

import (
	"testing"
	"time"
)

func TestGenerateQuotationOperationalInsights_ZeroData(t *testing.T) {
	overview := &QuotationAnalyticsOverview{
		Currency: "USD",
	}

	insights := GenerateQuotationOperationalInsights(overview, nil, nil)
	// Zero data organization should produce no false/fabricated insights
	if len(insights) != 0 {
		t.Errorf("expected 0 insights for empty org, got %d", len(insights))
	}
}

func TestGenerateQuotationOperationalInsights_Rules(t *testing.T) {
	now := time.Now()

	overview := &QuotationAnalyticsOverview{
		TotalQuotes:                  10,
		SentQuotes:                   4,
		UnviewedSentQuotes:           3,
		StuckInReviewCount:           2,
		ExpiringSoonCount:            3,
		AcceptedQuotes:               1,
		DeclinedQuotes:               5,
		AcceptanceRate:               16.7, // Low conversion
		QuoteToBookingConversionRate: 100.0,
		AverageGrossMarginPct:        5.5, // Compressed margin
		Currency:                     "USD",
	}

	risks := []*QuotationRiskItem{
		{
			QuotationID:       1,
			QuotationNumber:   "QT-001",
			TotalAmount:       60000,
			RiskCategory:      "EXPIRING_SOON",
			DaysUntilExpiry:   3,
			ValidUntil:        &now,
			RecommendedAction: "Follow up immediately",
		},
	}

	custs := []*CustomerQuotationPerformance{
		{
			CustomerName:      "Titan Logistics",
			AcceptedQuotes:    3,
			AcceptedValue:     45000,
			AverageQuoteValue: 15000,
		},
	}

	insights := GenerateQuotationOperationalInsights(overview, custs, risks)

	if len(insights) == 0 {
		t.Fatalf("expected multiple operational insights, got 0")
	}

	var foundExpiry, foundUnviewed, foundReview, foundLowWin, foundMargin, foundTopCust bool
	for _, ins := range insights {
		switch ins.Category {
		case "EXPIRY":
			foundExpiry = true
			if ins.Severity != "CRITICAL" {
				t.Errorf("expected CRITICAL severity for >$50k expiring value, got %s", ins.Severity)
			}
		case "CONVERSION":
			if ins.Headline == "Unviewed Customer Proposals" {
				foundUnviewed = true
			}
			if ins.Headline == "Low Win Rate / Commercial Friction" {
				foundLowWin = true
			}
		case "APPROVAL":
			foundReview = true
		case "MARGIN":
			foundMargin = true
		case "CUSTOMER":
			foundTopCust = true
		}
	}

	if !foundExpiry {
		t.Errorf("expected Expiry insight")
	}
	if !foundUnviewed {
		t.Errorf("expected Unviewed insight")
	}
	if !foundReview {
		t.Errorf("expected Approval Bottleneck insight")
	}
	if !foundLowWin {
		t.Errorf("expected Low Win Rate insight")
	}
	if !foundMargin {
		t.Errorf("expected Margin Risk insight")
	}
	if !foundTopCust {
		t.Errorf("expected Top Customer insight")
	}
}
