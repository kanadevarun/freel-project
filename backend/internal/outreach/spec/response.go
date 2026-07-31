package spec

import "time"

// Constants

const (
	CampaignStatusDraft     = "DRAFT"
	CampaignStatusActive    = "ACTIVE"
	CampaignStatusPaused    = "PAUSED"
	CampaignStatusCompleted = "COMPLETED"

	ChannelEmail    = "EMAIL"
	ChannelLinkedIn = "LINKEDIN"
)

// Models

type Campaign struct {
	ID        int32      `json:"id"`
	OrgID     int32      `json:"org_id"`
	Name      string     `json:"name"`
	Status    string     `json:"status"` // DRAFT | ACTIVE | PAUSED | COMPLETED
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Sequences []Sequence `json:"sequences,omitempty"`
}

type Sequence struct {
	ID         int32     `json:"id"`
	CampaignID int32     `json:"campaign_id"`
	StepNumber int       `json:"step_number"`
	Channel    string    `json:"channel"` // EMAIL | LINKEDIN
	Template   string    `json:"template"`
	DelayDays  int       `json:"delay_days"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Responses

type ListCampaignsResponse struct {
	Data       []*Campaign `json:"data"`
	TotalCount int         `json:"total_count"`
}

type GetCampaignResponse struct {
	Data *Campaign `json:"data"`
}

type CreateCampaignResponse struct {
	Data *Campaign `json:"data"`
}

type ActivateCampaignResponse struct {
	Data *Campaign `json:"data"`
}

type PauseCampaignResponse struct {
	Data *Campaign `json:"data"`
}

type DeleteCampaignResponse struct {
	Data map[string]interface{} `json:"data"`
}

type GenerateEmailResponse struct {
	Data map[string]interface{} `json:"data"` // Contains subject and body
}
