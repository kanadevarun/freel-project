package rbac

// Role names
const (
	RoleSuperAdmin   = "SUPER_ADMIN"
	RoleSales        = "SALES"
	RolePricing      = "PRICING"
	RoleOperations   = "OPERATIONS"
	RoleFinance      = "FINANCE"
	RoleDocumentation = "DOCUMENTATION"
	RoleHR           = "HR"
)

// Resource names
const (
	ResourceCompanies     = "COMPANIES"
	ResourceLeads         = "LEADS"
	ResourceOpportunities = "OPPORTUNITIES"
	ResourceRFQs          = "RFQS"
	ResourceOutreach      = "OUTREACH"
	ResourceShipments     = "SHIPMENTS"
	ResourceDocuments     = "DOCUMENTS"
	ResourceFinance       = "FINANCE"
	ResourceUsers         = "USERS"
	ResourceSettings      = "SETTINGS"
	ResourceDashboard     = "DASHBOARD"
)

// Action names
const (
	ActionCreate = "CREATE"
	ActionRead   = "READ"
	ActionUpdate = "UPDATE"
	ActionDelete = "DELETE"
)
