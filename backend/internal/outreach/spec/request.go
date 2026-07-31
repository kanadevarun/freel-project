package spec

type ListCampaignsRequest struct {
	OrgID  int32 `json:"-"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

type GetCampaignRequest struct {
	OrgID int32 `json:"-"`
	ID    int32 `json:"id"`
}

type CreateCampaignRequest struct {
	OrgID int32  `json:"-"`
	Name  string `json:"name" validate:"required"`
}

type ActivateCampaignRequest struct {
	OrgID int32 `json:"-"`
	ID    int32 `json:"id"`
}

type PauseCampaignRequest struct {
	OrgID int32 `json:"-"`
	ID    int32 `json:"id"`
}

type DeleteCampaignRequest struct {
	OrgID int32 `json:"-"`
	ID    int32 `json:"id"`
}

type GenerateEmailRequest struct {
	OrgID       int32  `json:"-"`
	CompanyName string `json:"company_name" validate:"required"`
	Industry    string `json:"industry"`
	Goal        string `json:"goal"`
}
