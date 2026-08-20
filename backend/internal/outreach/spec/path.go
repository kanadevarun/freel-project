package spec

// Outreach paths (relative to /api/v1/outreach mount)
const (
	ListCampaignsURL    = "/campaigns"
	CreateCampaignURL  = "/campaigns"
	GetCampaignURL     = "/campaigns/{id:[0-9]+}"
	ActivateCampaignURL = "/campaigns/{id:[0-9]+}/activate"
	PauseCampaignURL    = "/campaigns/{id:[0-9]+}/pause"
	DeleteCampaignURL   = "/campaigns/{id:[0-9]+}"
	GenerateEmailURL    = "/generate-email"
)
