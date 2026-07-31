package spec

// Reports paths
const (
	// swagger:operation GET /reports/metrics Reports GetMetrics
	// ---
	// summary: Get Business Metrics
	// description: Returns the aggregated business metrics for the reports dashboard.
	// responses:
	//   "200":
	//     description: Metrics data returned successfully
	//     schema:
	//         $ref: '#/definitions/MetricsResponse'
	//   "401":
	//     description: Unauthorized
	//   "500":
	//     description: Internal Server Error
	GetMetricsURL = "/api/v1/reports/metrics"
)
