package spec

type GetMissionControlResponse struct {
	Stats         Stats         `json:"stats"`
	ApprovalQueue []PendingTask `json:"approval_queue"`
	AIStatus      AIStatus      `json:"ai_status"`
}

type Stats struct {
	TotalRevenue    float64 `json:"total_revenue"`
	OpenRFQs        int     `json:"open_rfqs"`
	OpenLeads       int     `json:"open_leads"`
	ActiveShipments int     `json:"active_shipments"`
}

type PendingTask struct {
	ID        int32  `json:"id"`
	Type      string `json:"type"` // e.g., "RFQ_QUOTE_DRAFT"
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle"`
	Timestamp string `json:"timestamp"`
	RefID     int32  `json:"ref_id"` // RFQ ID or Lead ID
}

type AIStatus struct {
	ActiveAgents  int `json:"active_agents"`
	TasksFinished int `json:"tasks_finished"`
	HealthScore   int `json:"health_score"` // out of 100
}
