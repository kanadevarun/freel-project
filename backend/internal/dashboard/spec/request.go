package spec

type GetMissionControlRequest struct {
	OrgID int32 `json:"-"` // Extracted from context/auth
}
