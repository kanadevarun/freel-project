package spec

// Dashboard paths
const (
	// swagger:operation GET /dashboard/mission-control Dashboard GetMissionControl
	// ---
	// summary: Get Mission Control Data
	// description: Returns the aggregated mission control data for the dashboard.
	// responses:
	//   "200":
	//     description: Mission Control data returned successfully
	//     schema:
	//         $ref: '#/definitions/MissionControlResponse'
	//   "401":
	//     description: Unauthorized
	//   "500":
	//     description: Internal Server Error
	GetMissionControlURL = "/api/v1/dashboard/mission-control"
)
