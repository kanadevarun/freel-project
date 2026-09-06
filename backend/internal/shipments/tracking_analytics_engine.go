package shipments

import (
	"fmt"

	"github.com/freel/backend/internal/shipments/spec"
)

// GenerateOperationalInsights computes deterministic operational recommendations based on real MySQL data (Task 17.8)
func GenerateOperationalInsights(
	overview *spec.TrackingAnalyticsOverview,
	carriers []spec.CarrierTrackingPerformance,
	routes []spec.RouteTrackingPerformance,
) []spec.TrackingOperationalInsight {
	insights := make([]spec.TrackingOperationalInsight, 0)
	var insightIdx int

	nextID := func() string {
		insightIdx++
		return fmt.Sprintf("INS-%03d", insightIdx)
	}

	// 1. Critical Alert Urgency
	if overview.OpenCriticalAlerts > 0 {
		insights = append(insights, spec.TrackingOperationalInsight{
			ID:             nextID(),
			Category:       "ALERT_SPIKE",
			Severity:       "CRITICAL",
			Title:          fmt.Sprintf("%d Critical Tracking Alert%s Require Immediate Attention", overview.OpenCriticalAlerts, plural(overview.OpenCriticalAlerts)),
			Description:    "High-severity operational alerts (such as severe delays, route deviations, or missed port calls) are currently unaddressed in the active fleet.",
			Metric:         fmt.Sprintf("%d Critical / %d Total Open Alerts", overview.OpenCriticalAlerts, overview.TotalOpenAlerts),
			Recommendation: "Open the Tracking Alerts queue, review flagged exception milestones, and notify consignees with updated delivery schedules.",
		})
	}

	// 2. Carrier Reliability Bottlenecks
	for _, c := range carriers {
		if c.ShipmentsTracked >= 2 && (c.OnTimeRate < 75.0 || c.AverageDelayHours >= 24.0) {
			insights = append(insights, spec.TrackingOperationalInsight{
				ID:             nextID(),
				Category:       "CARRIER",
				Severity:       "HIGH",
				Title:          fmt.Sprintf("Carrier %s Exhibits Elevated Schedule Slippage", c.CarrierName),
				Description:    fmt.Sprintf("%s has an on-time delivery rate of %.1f%% with an average schedule delay of %.1f hours across %d tracked shipments.", c.CarrierName, c.OnTimeRate, c.AverageDelayHours, c.ShipmentsTracked),
				Metric:         fmt.Sprintf("%.1f%% On-Time | %.1fh Avg Delay", c.OnTimeRate, c.AverageDelayHours),
				Recommendation: fmt.Sprintf("Review booking allocations for %s and set proactive customer buffer times for subsequent voyages.", c.CarrierName),
			})
		}
	}

	// 3. High-Risk Corridor / Route Transit Variance
	for _, r := range routes {
		if r.ShipmentsCount >= 2 && (r.RiskLevel == "HIGH" || r.AvgTransitVarianceHours >= 24.0) {
			insights = append(insights, spec.TrackingOperationalInsight{
				ID:             nextID(),
				Category:       "ROUTE",
				Severity:       "MEDIUM",
				Title:          fmt.Sprintf("Corridor %s Experiences Systematic Transit Delays", r.RouteKey),
				Description:    fmt.Sprintf("Corridor %s has an average transit schedule variance of +%.1f hours across %d shipments with %.1f%% on-time reliability.", r.RouteKey, r.AvgTransitVarianceHours, r.ShipmentsCount, r.OnTimeRate),
				Metric:         fmt.Sprintf("+%.1fh Variance | %s Risk", r.AvgTransitVarianceHours, r.RiskLevel),
				Recommendation: fmt.Sprintf("Adjust standard transit estimates for %s and audit transshipment terminal queues at origin.", r.RouteKey),
			})
		}
	}

	// 4. Telemetry Freshness & Sensor Health
	if overview.ActiveShipments > 0 {
		staleOrMissing := overview.DataFreshnessStale + overview.DataFreshnessUnavailable
		staleRatio := float64(staleOrMissing) / float64(overview.ActiveShipments)
		if staleRatio > 0.25 {
			insights = append(insights, spec.TrackingOperationalInsight{
				ID:             nextID(),
				Category:       "DATA_FRESHNESS",
				Severity:       "MEDIUM",
				Title:          "Tracking Telemetry Freshness Is Degrading Across Fleet",
				Description:    fmt.Sprintf("%d of %d active shipments (%.0f%%) have stale or unavailable live AIS/carrier coordinates (>24h since last recorded position).", staleOrMissing, overview.ActiveShipments, staleRatio*100.0),
				Metric:         fmt.Sprintf("%d Stale/Unavailable of %d Active", staleOrMissing, overview.ActiveShipments),
				Recommendation: "Trigger manual tracking telemetry refresh across at-risk shipments and verify ocean carrier API connectivity.",
			})
		}
	}

	// 5. Background Refresh Reliability
	if overview.TotalRefreshes30d >= 5 && overview.RefreshSuccessRate < 90.0 {
		insights = append(insights, spec.TrackingOperationalInsight{
			ID:             nextID(),
			Category:       "PERFORMANCE",
			Severity:       "HIGH",
			Title:          "Telemetry Background Sync Failure Rate Exceeds Threshold",
			Description:    fmt.Sprintf("%.1f%% of background tracking sync cycles failed over the last 30 days (%d failed runs recorded).", 100.0-overview.RefreshSuccessRate, overview.FailedRefreshes30d),
			Metric:         fmt.Sprintf("%.1f%% Success Rate (%d Failures)", overview.RefreshSuccessRate, overview.FailedRefreshes30d),
			Recommendation: "Check external provider API rate limits and network latency in the Tracking Settings workspace.",
		})
	}

	// 6. Positive Fleet Performance (if operations are running smoothly)
	if overview.OnTimePercentage >= 90.0 && overview.OpenCriticalAlerts == 0 && overview.ActiveShipments > 0 {
		insights = append(insights, spec.TrackingOperationalInsight{
			ID:             nextID(),
			Category:       "PERFORMANCE",
			Severity:       "POSITIVE",
			Title:          "Fleet Operational Reliability Operating at High Efficiency",
			Description:    fmt.Sprintf("Active fleet is maintaining a %.1f%% on-time performance rate with zero critical schedule disruptions across %d active shipments.", overview.OnTimePercentage, overview.ActiveShipments),
			Metric:         fmt.Sprintf("%.1f%% On-Time Fleetwide", overview.OnTimePercentage),
			Recommendation: "Maintain current carrier allocation ratios and continue automated 60-minute telemetry monitoring cadences.",
		})
	}

	return insights
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
