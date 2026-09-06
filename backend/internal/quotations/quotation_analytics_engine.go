package quotations

import (
	"fmt"
)

// GenerateQuotationOperationalInsights evaluates deterministic business rules to yield management insights.
func GenerateQuotationOperationalInsights(
	overview *QuotationAnalyticsOverview,
	customers []*CustomerQuotationPerformance,
	risks []*QuotationRiskItem,
) []*QuotationOperationalInsight {
	var insights []*QuotationOperationalInsight
	if overview == nil {
		return insights
	}

	// 1. Pipeline Expiry Risk
	var expiringValue float64
	for _, r := range risks {
		if r.RiskCategory == "EXPIRING_SOON" {
			expiringValue += r.TotalAmount
		}
	}
	if overview.ExpiringSoonCount > 0 {
		severity := "WARNING"
		if expiringValue > 50000 || overview.ExpiringSoonCount >= 5 {
			severity = "CRITICAL"
		}
		insights = append(insights, &QuotationOperationalInsight{
			Category:          "EXPIRY",
			Severity:          severity,
			Headline:          "Active Pipeline Expiry Risk",
			Description:       fmt.Sprintf("%d quotation(s) totaling %s %s are expiring within the next 7 days.", overview.ExpiringSoonCount, overview.Currency, fmt.Sprintf("%.2f", expiringValue)),
			MetricValue:       fmt.Sprintf("%d quotes / %s %.0f", overview.ExpiringSoonCount, overview.Currency, expiringValue),
			RecommendedAction: "Prioritize customer follow-up to secure acceptance before rates and validity expire.",
		})
	}

	// 2. Unviewed Proposals
	if overview.UnviewedSentQuotes > 0 {
		insights = append(insights, &QuotationOperationalInsight{
			Category:          "CONVERSION",
			Severity:          "WARNING",
			Headline:          "Unviewed Customer Proposals",
			Description:       fmt.Sprintf("%d sent quotation(s) have not yet been opened or viewed by customers.", overview.UnviewedSentQuotes),
			MetricValue:       fmt.Sprintf("%d unviewed", overview.UnviewedSentQuotes),
			RecommendedAction: "Verify recipient email accuracy or resend quotation link via customer portal.",
		})
	}

	// 3. Approval Bottleneck Detection
	if overview.StuckInReviewCount > 0 {
		insights = append(insights, &QuotationOperationalInsight{
			Category:          "APPROVAL",
			Severity:          "WARNING",
			Headline:          "Internal Review Bottleneck",
			Description:       fmt.Sprintf("%d quotation(s) are awaiting management review and pricing authorization.", overview.StuckInReviewCount),
			MetricValue:       fmt.Sprintf("%d in review", overview.StuckInReviewCount),
			RecommendedAction: "Review pending submissions to unblock sales teams and minimize quote turnaround time.",
		})
	}

	// 4. Conversion Rate Health
	decidedCount := overview.AcceptedQuotes + overview.DeclinedQuotes
	if decidedCount >= 3 {
		if overview.AcceptanceRate < 25.0 {
			insights = append(insights, &QuotationOperationalInsight{
				Category:          "CONVERSION",
				Severity:          "WARNING",
				Headline:          "Low Win Rate / Commercial Friction",
				Description:       fmt.Sprintf("Quotation acceptance rate is currently %.1f%% across %d decided commercial offers.", overview.AcceptanceRate, decidedCount),
				MetricValue:       fmt.Sprintf("%.1f%% Win Rate", overview.AcceptanceRate),
				RecommendedAction: "Benchmark rate competitiveness and assess if carrier buy rates need renegotiation.",
			})
		} else if overview.AcceptanceRate >= 60.0 {
			insights = append(insights, &QuotationOperationalInsight{
				Category:          "CONVERSION",
				Severity:          "SUCCESS",
				Headline:          "Strong Commercial Win Rate",
				Description:       fmt.Sprintf("Exceptional quotation acceptance rate of %.1f%% across decided proposals.", overview.AcceptanceRate),
				MetricValue:       fmt.Sprintf("%.1f%% Win Rate", overview.AcceptanceRate),
				RecommendedAction: "Sales pricing strategy is well calibrated to current freight market demand.",
			})
		}
	}

	// 5. Quote-to-Booking Conversion Progress
	if overview.AcceptedQuotes > 0 {
		if overview.QuoteToBookingConversionRate >= 75.0 {
			insights = append(insights, &QuotationOperationalInsight{
				Category:          "OPERATIONS",
				Severity:          "SUCCESS",
				Headline:          "High Operational Conversion Efficiency",
				Description:       fmt.Sprintf("%.1f%% of accepted commercial quotes successfully converted into active bookings.", overview.QuoteToBookingConversionRate),
				MetricValue:       fmt.Sprintf("%.1f%% Converted", overview.QuoteToBookingConversionRate),
				RecommendedAction: "Smooth commercial handover between sales and operational booking execution.",
			})
		} else if overview.QuoteToBookingConversionRate < 40.0 {
			insights = append(insights, &QuotationOperationalInsight{
				Category:          "OPERATIONS",
				Severity:          "WARNING",
				Headline:          "Pending Operational Conversion Backlog",
				Description:       fmt.Sprintf("%.1f%% of accepted quotations have been converted to operational bookings.", overview.QuoteToBookingConversionRate),
				MetricValue:       fmt.Sprintf("%.1f%% Converted", overview.QuoteToBookingConversionRate),
				RecommendedAction: "Ensure operators promptly convert accepted commercial quotes into operational bookings.",
			})
		}
	}

	// 6. Margin Health & Profitability Risk
	if overview.TotalQuotes > 0 && overview.AverageGrossMarginPct < 8.0 && overview.AverageGrossMarginPct > 0 {
		insights = append(insights, &QuotationOperationalInsight{
			Category:          "MARGIN",
			Severity:          "WARNING",
			Headline:          "Compressed Gross Margins",
			Description:       fmt.Sprintf("Average gross margin on quotations is %.1f%%, below the standard 12%% freight target.", overview.AverageGrossMarginPct),
			MetricValue:       fmt.Sprintf("%.1f%% Margin", overview.AverageGrossMarginPct),
			RecommendedAction: "Review surcharge markups and evaluate carrier contract procurement discounts.",
		})
	}

	// 7. Top Client Account Performance
	if len(customers) > 0 {
		topCust := customers[0]
		if topCust.AcceptedQuotes >= 2 && topCust.AcceptedValue > 10000 {
			insights = append(insights, &QuotationOperationalInsight{
				Category:          "CUSTOMER",
				Severity:          "INFO",
				Headline:          fmt.Sprintf("Primary Account: %s", topCust.CustomerName),
				Description:       fmt.Sprintf("Leading customer account with %d accepted quotes and %s %.2f in commercial booking value.", topCust.AcceptedQuotes, overview.Currency, topCust.AcceptedValue),
				MetricValue:       fmt.Sprintf("%s %.0f", overview.Currency, topCust.AcceptedValue),
				RecommendedAction: "Maintain priority vessel space allocation and dedicated account service.",
			})
		}
	}

	return insights
}
