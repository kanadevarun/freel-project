package rates

import "fmt"

// ── Task 19.7: Rate Analytics Insight Engine ──────────────────────────────────
//
// GenerateRateCommercialInsights applies deterministic business rules to
// pre-fetched analytics data and returns actionable CommercialImpactInsight cards.
// No database calls, no side effects — pure functional logic.
//
// Rules:
//  A. Expiry Exposure
//  B. Contract Renewal
//  C. Carrier Concentration (from lane data)
//  D. Spot Sourcing Performance
//  E. Commercial Risk Exposure (quotations with unresolved risk)
//  F. Rate Coverage Gaps
//  G. Positive Performance Signals
func GenerateRateCommercialInsights(
	overview *RateAnalyticsOverview,
	lifecycle *RateLifecycleAnalytics,
	spot *SpotSourcingPerformance,
	carriers []CarrierRatePerformance,
	lanes []LaneRatePerformance,
	riskExposure int,
) []CommercialImpactInsight {
	if overview == nil {
		overview = &RateAnalyticsOverview{}
	}
	if lifecycle == nil {
		lifecycle = &RateLifecycleAnalytics{}
	}
	if spot == nil {
		spot = &SpotSourcingPerformance{}
	}

	insights := make([]CommercialImpactInsight, 0, 10)

	// ── A. Expiry Exposure ──────────────────────────────────────────────────────

	if lifecycle.Expired > 0 {
		insights = append(insights, CommercialImpactInsight{
			Category:          "EXPIRY",
			Severity:          "CRITICAL",
			Headline:          fmt.Sprintf("%d rate(s) have already expired", lifecycle.Expired),
			Description:       "Expired rates are no longer commercially valid and cannot be selected for new quotations. Renew or replace them to maintain full lane coverage.",
			MetricValue:       fmt.Sprintf("%d expired", lifecycle.Expired),
			RecommendedAction: "Review expired rates and create new versions",
			RelatedEntityType: "rate",
		})
	}

	total := lifecycle.Active + lifecycle.ExpiringSoon
	if lifecycle.ExpiringSoon > 0 && total > 0 {
		ratio := float64(lifecycle.ExpiringSoon) / float64(total)
		sev := "INFO"
		if ratio >= 0.2 {
			sev = "WARNING"
		}
		insights = append(insights, CommercialImpactInsight{
			Category:          "EXPIRY",
			Severity:          sev,
			Headline:          fmt.Sprintf("%d rate(s) expire within 30 days", lifecycle.ExpiringSoon),
			Description:       fmt.Sprintf("%.0f%% of your active rate portfolio is approaching expiry. Proactively renewing these rates avoids last-minute procurement gaps.", ratio*100),
			MetricValue:       fmt.Sprintf("%d / %d rates expiring", lifecycle.ExpiringSoon, total),
			RecommendedAction: "Open Rate Lifecycle workspace to review expiring rates",
			RelatedEntityType: "rate",
		})
	}

	// ── B. Contract Renewal ─────────────────────────────────────────────────────

	if overview.ContractsRequiringRenewal > 0 {
		insights = append(insights, CommercialImpactInsight{
			Category:          "CONTRACT",
			Severity:          "WARNING",
			Headline:          fmt.Sprintf("%d carrier contract(s) require renewal", overview.ContractsRequiringRenewal),
			Description:       "Active contracts expiring within 30 days with no renewal action started. Lapsing contracts may reduce the carrier rate pool available for quotations.",
			MetricValue:       fmt.Sprintf("%d contract(s)", overview.ContractsRequiringRenewal),
			RecommendedAction: "Go to Carrier Contracts workspace to initiate renewal",
			RelatedEntityType: "contract",
		})
	}

	// ── C. Carrier Concentration Risk ──────────────────────────────────────────

	singleCarrierLanes := 0
	for _, lane := range lanes {
		if lane.ActiveRates > 0 && lane.CarrierCount == 1 {
			singleCarrierLanes++
		}
	}
	if singleCarrierLanes > 0 {
		sev := "INFO"
		if singleCarrierLanes >= 3 {
			sev = "WARNING"
		}
		insights = append(insights, CommercialImpactInsight{
			Category:          "CARRIER_CONCENTRATION",
			Severity:          sev,
			Headline:          fmt.Sprintf("%d lane(s) depend on a single carrier", singleCarrierLanes),
			Description:       "Lanes with only one active carrier have no commercial fallback if that carrier withdraws. Consider sourcing additional carrier rates for these lanes.",
			MetricValue:       fmt.Sprintf("%d / %d active lanes single-carrier", singleCarrierLanes, len(lanes)),
			RecommendedAction: "Review Lane Coverage table and source additional carrier rates",
			RelatedEntityType: "rate",
		})
	}

	// ── D. Spot Sourcing Performance ────────────────────────────────────────────

	if spot.TotalRequests > 0 {
		responseRate := float64(spot.FullyResponded) / float64(spot.TotalRequests)
		if responseRate < 0.5 {
			insights = append(insights, CommercialImpactInsight{
				Category:          "SPOT_SOURCING",
				Severity:          "WARNING",
				Headline:          fmt.Sprintf("Only %.0f%% of spot requests have carrier responses", responseRate*100),
				Description:       "Low carrier response rates reduce competitive benchmarking and may delay quotation commercial approvals.",
				MetricValue:       fmt.Sprintf("%d / %d requests responded", spot.FullyResponded, spot.TotalRequests),
				RecommendedAction: "Review and follow up on unresponded spot requests",
				RelatedEntityType: "spot_request",
			})
		}
	}

	if spot.Expired > 0 {
		insights = append(insights, CommercialImpactInsight{
			Category:          "SPOT_SOURCING",
			Severity:          "WARNING",
			Headline:          fmt.Sprintf("%d spot request(s) expired without a carrier selection", spot.Expired),
			Description:       "Spot requests that expire unselected represent missed procurement opportunities.",
			MetricValue:       fmt.Sprintf("%d expired requests", spot.Expired),
			RecommendedAction: "Reopen expired spot requests or create new ones",
			RelatedEntityType: "spot_request",
		})
	}

	// ── E. Commercial Risk Exposure ─────────────────────────────────────────────

	if riskExposure > 0 {
		sev := "WARNING"
		if riskExposure >= 5 {
			sev = "CRITICAL"
		}
		insights = append(insights, CommercialImpactInsight{
			Category:          "QUOTATION_RISK",
			Severity:          sev,
			Headline:          fmt.Sprintf("%d quotation(s) have unresolved rate risk events", riskExposure),
			Description:       "Quotations with expired or superseded rate snapshots may expose the organisation to commercial mispricing.",
			MetricValue:       fmt.Sprintf("%d open risk events", riskExposure),
			RecommendedAction: "Open Quotations and review Commercial Risk & Rate Health panels",
			RelatedEntityType: "quotation",
		})
	}

	// ── F. Coverage Gaps ────────────────────────────────────────────────────────

	uncoveredLanes := 0
	for _, lane := range lanes {
		if lane.ActiveRates == 0 && lane.AvailableRates > 0 {
			uncoveredLanes++
		}
	}
	if uncoveredLanes > 0 {
		insights = append(insights, CommercialImpactInsight{
			Category:          "COVERAGE_GAP",
			Severity:          "CRITICAL",
			Headline:          fmt.Sprintf("%d lane(s) have no active rates", uncoveredLanes),
			Description:       "These lanes have historical rate records but no currently valid active rate. Quotations for these lanes cannot be commercially priced.",
			MetricValue:       fmt.Sprintf("%d uncovered lanes", uncoveredLanes),
			RecommendedAction: "Create new rates or renew expired rates for uncovered lanes",
			RelatedEntityType: "rate",
		})
	}

	// ── G. Positive Performance Signals ─────────────────────────────────────────

	if spot.TotalRequests >= 3 && spot.SelectionRate >= 70 {
		insights = append(insights, CommercialImpactInsight{
			Category:          "PERFORMANCE",
			Severity:          "SUCCESS",
			Headline:          fmt.Sprintf("Strong spot procurement: %.0f%% selection rate", spot.SelectionRate),
			Description:       "A high spot quote selection rate indicates effective carrier engagement and a strong competitive rate pool.",
			MetricValue:       fmt.Sprintf("%.0f%% (%d / %d)", spot.SelectionRate, spot.Selected, spot.TotalRequests),
			RecommendedAction: "Continue growing the carrier network to maintain strong competition",
			RelatedEntityType: "spot_request",
		})
	}

	if overview.TotalRates > 0 {
		activeRatio := float64(overview.ActiveRates) / float64(overview.TotalRates)
		if activeRatio >= 0.8 {
			insights = append(insights, CommercialImpactInsight{
				Category:          "PERFORMANCE",
				Severity:          "SUCCESS",
				Headline:          fmt.Sprintf("%.0f%% of rates are active and commercially valid", activeRatio*100),
				Description:       "Excellent rate portfolio health. The majority of your indexed rates are within their validity window.",
				MetricValue:       fmt.Sprintf("%d / %d active rates", overview.ActiveRates, overview.TotalRates),
				RecommendedAction: "Monitor expiring rates to maintain this coverage level",
				RelatedEntityType: "rate",
			})
		}
	}

	var topCarrier *CarrierRatePerformance
	for i := range carriers {
		c := &carriers[i]
		if c.SpotResponsesCount >= 2 && (topCarrier == nil || c.SelectionRate > topCarrier.SelectionRate) {
			topCarrier = c
		}
	}
	if topCarrier != nil && topCarrier.SelectionRate >= 50 {
		insights = append(insights, CommercialImpactInsight{
			Category:          "PERFORMANCE",
			Severity:          "SUCCESS",
			Headline:          fmt.Sprintf("%s has the highest spot selection rate at %.0f%%", topCarrier.CarrierName, topCarrier.SelectionRate),
			Description:       "This carrier demonstrates strong commercial competitiveness across your spot request portfolio.",
			MetricValue:       fmt.Sprintf("%.0f%% (%d/%d selections)", topCarrier.SelectionRate, topCarrier.SpotSelections, topCarrier.SpotResponsesCount),
			RecommendedAction: "Prioritise this carrier for future spot rate sourcing",
			RelatedEntityType: "rate",
		})
	}

	return insights
}
