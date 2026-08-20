package spec

import "time"

// Lead models

type Lead struct {
	ID               int32     `json:"id" db:"id"`
	OrgID            int32     `json:"org_id" db:"org_id"`
	CompanyName      string    `json:"company_name" db:"company_name"`
	ContactName      *string   `json:"contact_name" db:"contact_name"`
	Email            *string   `json:"email" db:"email"`
	Phone            *string   `json:"phone" db:"phone"`
	Status           string    `json:"status" db:"status"` // NEW, QUALIFIED, CONVERTED, REJECTED
	Source           *string   `json:"source" db:"source"`
	AIScore          int32     `json:"ai_score" db:"ai_score"`
	AIResearchReport *string   `json:"ai_research_report" db:"ai_research_report"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
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
