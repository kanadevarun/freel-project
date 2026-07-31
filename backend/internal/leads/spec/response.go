package spec

import "time"

// Lead models

type Lead struct {
	ID               int32     `json:"id"`
	OrgID            int32     `json:"org_id"`
	CompanyName      string    `json:"company_name"`
	ContactName      *string   `json:"contact_name"`
	Email            *string   `json:"email"`
	Phone            *string   `json:"phone"`
	Status           string    `json:"status"` // NEW, QUALIFIED, CONVERTED, REJECTED
	Source           *string   `json:"source"`
	AIScore          int32     `json:"ai_score"`
	AIResearchReport *string   `json:"ai_research_report"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
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
