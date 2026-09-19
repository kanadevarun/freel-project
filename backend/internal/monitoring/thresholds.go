package monitoring

import (
	"fmt"
	"time"
)

// DefaultThresholds provides sensible production defaults for AI operations.
var DefaultThresholds = map[string]struct {
	Value       float64
	Unit        string
	Severity    string
	Description string
}{
	"max_p95_latency_ms": {
		Value:       5000.0,
		Unit:        "ms",
		Severity:    "CRITICAL",
		Description: "Maximum acceptable P95 latency for AI completions",
	},
	"warn_p95_latency_ms": {
		Value:       3000.0,
		Unit:        "ms",
		Severity:    "WARNING",
		Description: "Latency threshold triggering performance degradation warnings",
	},
	"max_failure_rate_pct": {
		Value:       10.0,
		Unit:        "percentage",
		Severity:    "CRITICAL",
		Description: "Maximum failure percentage before declaring subsystem degraded",
	},
	"warn_failure_rate_pct": {
		Value:       5.0,
		Unit:        "percentage",
		Severity:    "WARNING",
		Description: "Failure rate threshold triggering operational warnings",
	},
	"max_queue_backlog": {
		Value:       50.0,
		Unit:        "count",
		Severity:    "CRITICAL",
		Description: "Queue backlog threshold indicating task execution bottleneck",
	},
	"max_queue_age_seconds": {
		Value:       300.0,
		Unit:        "seconds",
		Severity:    "WARNING",
		Description: "Maximum acceptable age of oldest pending queue item",
	},
	"max_daily_cost_usd": {
		Value:       100.0,
		Unit:        "usd",
		Severity:    "WARNING",
		Description: "Daily AI spend budget threshold warning",
	},
	"max_security_events_24h": {
		Value:       5.0,
		Unit:        "count",
		Severity:    "CRITICAL",
		Description: "Security events threshold indicating potential active abuse",
	},
}

// EvaluateHealthState examines operational metrics against active thresholds.
func EvaluateHealthState(
	perf PerformanceMetrics,
	queue QueueWorkerMetrics,
	cost CostSummary,
	sec SecurityMetrics,
	customThresholds map[string]HealthThresholdRecord,
) (SystemHealthState, []OperationalAlert) {
	var alerts []OperationalAlert
	overall := HealthStateHealthy

	// Helper to get threshold value
	getThreshold := func(key string, defaultVal float64) float64 {
		if customThresholds != nil {
			if rec, ok := customThresholds[key]; ok {
				return rec.ThresholdValue
			}
		}
		if def, ok := DefaultThresholds[key]; ok {
			return def.Value
		}
		return defaultVal
	}

	// 1. Failure rate check
	maxFailRate := getThreshold("max_failure_rate_pct", 10.0)
	warnFailRate := getThreshold("warn_failure_rate_pct", 5.0)
	if perf.TotalRequests >= 10 {
		failRate := (float64(perf.FailureCount) / float64(perf.TotalRequests)) * 100.0
		if failRate >= maxFailRate {
			overall = HealthStateDegraded
			alerts = append(alerts, OperationalAlert{
				ID:             "alert-failrate-critical",
				Severity:       "CRITICAL",
				Subsystem:      "AI Runtime",
				Title:          "High AI Failure Rate",
				Message:        fmt.Sprintf("AI request failure rate is %.1f%% (threshold: %.1f%%)", failRate, maxFailRate),
				MetricKey:      "failure_rate_pct",
				CurrentValue:   failRate,
				ThresholdValue: maxFailRate,
				Unit:           "percentage",
				TriggeredAt:    time.Now(),
			})
		} else if failRate >= warnFailRate {
			alerts = append(alerts, OperationalAlert{
				ID:             "alert-failrate-warning",
				Severity:       "WARNING",
				Subsystem:      "AI Runtime",
				Title:          "Elevated AI Failure Rate",
				Message:        fmt.Sprintf("AI request failure rate is elevated at %.1f%% (warning threshold: %.1f%%)", failRate, warnFailRate),
				MetricKey:      "failure_rate_pct",
				CurrentValue:   failRate,
				ThresholdValue: warnFailRate,
				Unit:           "percentage",
				TriggeredAt:    time.Now(),
			})
		}
	}

	// 2. Latency check
	maxP95 := getThreshold("max_p95_latency_ms", 5000.0)
	warnP95 := getThreshold("warn_p95_latency_ms", 3000.0)
	if perf.Latency.P95Ms > int64(maxP95) {
		if overall == HealthStateHealthy {
			overall = HealthStateDegraded
		}
		alerts = append(alerts, OperationalAlert{
			ID:             "alert-latency-critical",
			Severity:       "CRITICAL",
			Subsystem:      "AI Runtime",
			Title:          "Critical Latency Spike",
			Message:        fmt.Sprintf("P95 latency is %dms (threshold: %.0fms)", perf.Latency.P95Ms, maxP95),
			MetricKey:      "p95_latency_ms",
			CurrentValue:   float64(perf.Latency.P95Ms),
			ThresholdValue: maxP95,
			Unit:           "ms",
			TriggeredAt:    time.Now(),
		})
	} else if perf.Latency.P95Ms > int64(warnP95) {
		alerts = append(alerts, OperationalAlert{
			ID:             "alert-latency-warning",
			Severity:       "WARNING",
			Subsystem:      "AI Runtime",
			Title:          "High Latency Warning",
			Message:        fmt.Sprintf("P95 latency is %dms (warning threshold: %.0fms)", perf.Latency.P95Ms, warnP95),
			MetricKey:      "p95_latency_ms",
			CurrentValue:   float64(perf.Latency.P95Ms),
			ThresholdValue: warnP95,
			Unit:           "ms",
			TriggeredAt:    time.Now(),
		})
	}

	// 3. Queue backlog check
	maxBacklog := getThreshold("max_queue_backlog", 50.0)
	if float64(queue.QueuedJobs) >= maxBacklog {
		if overall == HealthStateHealthy {
			overall = HealthStateDegraded
		}
		alerts = append(alerts, OperationalAlert{
			ID:             "alert-queue-backlog",
			Severity:       "CRITICAL",
			Subsystem:      "AI Queue Worker",
			Title:          "Queue Worker Backlog",
			Message:        fmt.Sprintf("Queued background jobs (%d) exceeded capacity threshold (%.0f)", queue.QueuedJobs, maxBacklog),
			MetricKey:      "queue_backlog",
			CurrentValue:   float64(queue.QueuedJobs),
			ThresholdValue: maxBacklog,
			Unit:           "count",
			TriggeredAt:    time.Now(),
		})
	}

	// 4. Queue age check
	maxAge := getThreshold("max_queue_age_seconds", 300.0)
	if float64(queue.OldestQueuedAgeSec) > maxAge && queue.QueuedJobs > 0 {
		alerts = append(alerts, OperationalAlert{
			ID:             "alert-queue-age",
			Severity:       "WARNING",
			Subsystem:      "AI Queue Worker",
			Title:          "Oldest Queue Job Stagnant",
			Message:        fmt.Sprintf("Oldest queued job waiting for %d seconds (threshold: %.0fs)", queue.OldestQueuedAgeSec, maxAge),
			MetricKey:      "oldest_queued_age_sec",
			CurrentValue:   float64(queue.OldestQueuedAgeSec),
			ThresholdValue: maxAge,
			Unit:           "seconds",
			TriggeredAt:    time.Now(),
		})
	}

	// 5. Daily cost warning check
	maxDailyCost := getThreshold("max_daily_cost_usd", 100.0)
	if cost.TotalEstimatedCost > maxDailyCost {
		alerts = append(alerts, OperationalAlert{
			ID:             "alert-cost-budget",
			Severity:       "WARNING",
			Subsystem:      "Billing & Cost",
			Title:          "Daily AI Spend Threshold Exceeded",
			Message:        fmt.Sprintf("Total estimated AI spend is $%.2f (budget threshold: $%.2f)", cost.TotalEstimatedCost, maxDailyCost),
			MetricKey:      "daily_cost_usd",
			CurrentValue:   cost.TotalEstimatedCost,
			ThresholdValue: maxDailyCost,
			Unit:           "usd",
			TriggeredAt:    time.Now(),
		})
	}

	// 6. Security events check
	maxSecEvents := getThreshold("max_security_events_24h", 5.0)
	if float64(sec.TotalSecurityEvents24h) >= maxSecEvents {
		alerts = append(alerts, OperationalAlert{
			ID:             "alert-security-events",
			Severity:       "CRITICAL",
			Subsystem:      "Safety & Security",
			Title:          "Elevated Security Event Volume",
			Message:        fmt.Sprintf("%d security/safety events recorded in 24 hours (threshold: %.0f)", sec.TotalSecurityEvents24h, maxSecEvents),
			MetricKey:      "security_events_24h",
			CurrentValue:   float64(sec.TotalSecurityEvents24h),
			ThresholdValue: maxSecEvents,
			Unit:           "count",
			TriggeredAt:    time.Now(),
		})
	}

	return overall, alerts
}
