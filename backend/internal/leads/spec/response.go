package spec

import "time"

// Lead models

type Lead struct {
	ID               int32      `json:"id" db:"id"`
	OrgID            int32      `json:"org_id" db:"org_id"`
	CompanyName      string     `json:"company_name" db:"company_name"`
	ContactName      *string    `json:"contact_name" db:"contact_name"`
	Email            *string    `json:"email" db:"email"`
	Phone            *string    `json:"phone" db:"phone"`
	Status           string     `json:"status" db:"status"` // NEW, QUALIFIED, CONVERTED, REJECTED
	Source           *string    `json:"source" db:"source"`
	AIScore          int32      `json:"ai_score" db:"ai_score"`
	AIResearchReport *string    `json:"ai_research_report" db:"ai_research_report"`
	Notes            *string    `json:"notes" db:"notes"`
	Location         *string    `json:"location" db:"location"`
	AssignedTo       *int64     `json:"assigned_to" db:"assigned_to"`
	AssignedAt       *time.Time `json:"assigned_at" db:"assigned_at"`
	AssignedToName   *string    `json:"assigned_to_name" db:"assigned_to_name"`
	Tags             []string   `json:"tags"`
	LinkedRFQID              *int32     `json:"linked_rfq_id,omitempty" db:"linked_rfq_id"`
	LinkedRFQNumber          *string    `json:"linked_rfq_number,omitempty" db:"linked_rfq_number"`
	CampaignID               *int64     `json:"campaign_id,omitempty" db:"campaign_id"`
	ConvertedFromOutreachAt  *time.Time `json:"converted_from_outreach_at,omitempty" db:"converted_from_outreach_at"`
	CampaignName             *string    `json:"campaign_name,omitempty" db:"campaign_name"`
	CreatedAt                time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at" db:"updated_at"`
}



// Responses

type ListLeadsResponse struct {
	Data       []*Lead `json:"data"`
	TotalCount int     `json:"total_count"`
}

type GetLeadResponse struct {
	Data *Lead `json:"data"`
}

type CreateLeadResponse struct {
	Data *Lead `json:"data"`
}

type UpdateLeadResponse struct {
	Data *Lead `json:"data"`
}

type ImportLeadsResponse struct {
	Data map[string]interface{} `json:"data"`
}

type DeleteLeadResponse struct {
	Data map[string]interface{} `json:"data"`
}

type BulkUpdateReport struct {
	SuccessIDs []int32          `json:"success_ids"`
	FailedIDs  map[string]string `json:"failed_ids"` // string keys for JSON object compatibility
}

type BulkUpdateLeadsResponse struct {
	Data BulkUpdateReport `json:"data"`
}

type TimelineEvent struct {
	Action        string     `json:"action"`
	Description   string     `json:"description"`
	Actor         string     `json:"actor"`
	Source        string     `json:"source"`
	Timestamp     time.Time  `json:"timestamp"`
	InteractionID *int64     `json:"interaction_id,omitempty"`
}

type GetTimelineResponse struct {
	Data []TimelineEvent `json:"data"`
}
