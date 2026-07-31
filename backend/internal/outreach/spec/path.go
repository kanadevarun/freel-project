package spec

const (
	// swagger:operation GET /outreach/campaigns Outreach ListCampaigns
	// ---
	// summary: List Campaigns
	ListCampaignsURL = "/api/v1/outreach/campaigns"

	// swagger:operation POST /outreach/campaigns Outreach CreateCampaign
	// ---
	// summary: Create Campaign
	CreateCampaignURL = "/api/v1/outreach/campaigns"

	// swagger:operation GET /outreach/campaigns/{id} Outreach GetCampaign
	// ---
	// summary: Get Campaign
	GetCampaignURL = "/api/v1/outreach/campaigns/{id:[0-9]+}"

	// swagger:operation POST /outreach/campaigns/{id}/activate Outreach ActivateCampaign
	// ---
	// summary: Activate Campaign
	ActivateCampaignURL = "/api/v1/outreach/campaigns/{id:[0-9]+}/activate"

	// swagger:operation POST /outreach/campaigns/{id}/pause Outreach PauseCampaign
	// ---
	// summary: Pause Campaign
	PauseCampaignURL = "/api/v1/outreach/campaigns/{id:[0-9]+}/pause"

	// swagger:operation DELETE /outreach/campaigns/{id} Outreach DeleteCampaign
	// ---
	// summary: Delete Campaign
	DeleteCampaignURL = "/api/v1/outreach/campaigns/{id:[0-9]+}"

	// swagger:operation POST /outreach/generate-email Outreach GenerateEmail
	// ---
	// summary: Generate Email
	GenerateEmailURL = "/api/v1/outreach/generate-email"
)
