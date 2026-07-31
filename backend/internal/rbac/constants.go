package rbac

// Role names
const (
	RoleSuperAdmin = "SUPER_ADMIN"
	RoleSales      = "SALES"
	RolePricing    = "PRICING"
	RoleOperations = "OPERATIONS"
	RoleFinance    = "FINANCE"
	RoleCustomer   = "CUSTOMER"
)

// Resource names
const (
	ResourceCompanies     = "COMPANIES"
	ResourceLeads         = "LEADS"
	ResourceOpportunities = "OPPORTUNITIES"
	ResourceRFQs          = "RFQS"
	ResourceOutreach      = "OUTREACH"
	ResourceDashboard     = "DASHBOARD"
)

// Action names
const (
	ActionCreate = "CREATE"
	ActionRead   = "READ"
	ActionUpdate = "UPDATE"
	ActionDelete = "DELETE"
)
