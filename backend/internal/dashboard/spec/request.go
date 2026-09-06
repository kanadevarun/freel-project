package spec

type GetMissionControlRequest struct {
	OrgID     int64  `json:"-"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Preset    string `json:"preset,omitempty"`
}
