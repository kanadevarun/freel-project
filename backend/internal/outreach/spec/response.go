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
	ID        int32      `json:"id" db:"id"`
	OrgID     int32      `json:"org_id" db:"org_id"`
	Name      string     `json:"name" db:"name"`
	Status    string     `json:"status" db:"status"` // DRAFT | ACTIVE | PAUSED | COMPLETED
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	Sequences []Sequence `json:"sequences,omitempty"`
}

type Sequence struct {
	ID         int32     `json:"id" db:"id"`
	CampaignID int32     `json:"campaign_id" db:"campaign_id"`
	StepNumber int       `json:"step_number" db:"step_number"`
	Channel    string    `json:"channel" db:"channel"` // EMAIL | LINKEDIN
	Name       string    `json:"name" db:"name"`
	Subject    string    `json:"subject" db:"subject"`
	Body       string    `json:"body" db:"body"`
	Template   string    `json:"template" db:"template"`
	DelayDays  int       `json:"delay_days" db:"delay_days"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
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
