package spec

// Leads paths
const (
	// swagger:operation GET /leads Leads ListLeads
	// ---
	// summary: List Leads
	ListURL = "/api/v1/leads"

	// swagger:operation POST /leads Leads CreateLead
	// ---
	// summary: Create Lead
	CreateURL = "/api/v1/leads"

	// swagger:operation POST /leads/import Leads ImportLeads
	// ---
	// summary: Import Leads
	ImportURL = "/api/v1/leads/import"

	// swagger:operation GET /leads/{id} Leads GetLead
	// ---
	// summary: Get Lead
	GetURL = "/api/v1/leads/{id:[0-9]+}"

	// swagger:operation PUT /leads/{id} Leads UpdateLead
	// ---
	// summary: Update Lead
	UpdateURL = "/api/v1/leads/{id:[0-9]+}"

	// swagger:operation DELETE /leads/{id} Leads DeleteLead
	// ---
	// summary: Delete Lead
	DeleteURL = "/api/v1/leads/{id:[0-9]+}"
)
