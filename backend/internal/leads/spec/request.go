package spec

type ListLeadsRequest struct {
	OrgID  int32   `json:"-"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
	Status *string `json:"status"`
}

type GetLeadRequest struct {
	OrgID int32 `json:"-"`
	ID    int32 `json:"id"`
}

type CreateLeadRequest struct {
	OrgID       int32   `json:"-"`
	CompanyName string  `json:"company_name" validate:"required"`
	ContactName *string `json:"contact_name"`
	Email       *string `json:"email" validate:"omitempty,email"`
	Phone       *string `json:"phone"`
	Source      *string `json:"source"`
}

type UpdateLeadRequest struct {
	OrgID            int32   `json:"-"`
	ID               int32   `json:"id"`
	CompanyName      *string `json:"company_name"`
	ContactName      *string `json:"contact_name"`
	Email            *string `json:"email" validate:"omitempty,email"`
	Phone            *string `json:"phone"`
	Status           *string `json:"status"`
	Source           *string `json:"source"`
	AIScore          *int32  `json:"ai_score"`
	AIResearchReport *string `json:"ai_research_report"`
}

type DeleteLeadRequest struct {
	OrgID int32 `json:"-"`
	ID    int32 `json:"id"`
}

type ImportLeadsRequest struct {
	OrgID int32                `json:"-"`
	Leads []*CreateLeadRequest `json:"leads"`
}
