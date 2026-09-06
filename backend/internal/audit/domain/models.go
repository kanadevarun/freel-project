package domain

import (
	"time"
)

// Actor Types representing who or what executed the operation.
const (
	ActorTypeUser    = "USER"
	ActorTypeSystem  = "SYSTEM"
	ActorTypeAIAgent = "AI_AGENT"
)

// Universal Module Categories covering all LogisticsHQ operations.
const (
	ModuleAuthentication      = "AUTHENTICATION"
	ModuleUsers               = "USERS"
	ModuleRolesPermissions    = "ROLES_PERMISSIONS"
	ModuleLeads               = "LEADS"
	ModuleRFQs                = "RFQS"
	ModuleQuotations          = "QUOTATIONS"
	ModuleBookings            = "BOOKINGS"
	ModuleShipments           = "SHIPMENTS"
	ModuleTracking            = "TRACKING"
	ModuleRateManagement      = "RATE_MANAGEMENT"
	ModuleContracts           = "CONTRACTS"
	ModuleCustomers           = "CUSTOMERS"
	ModuleOutreach            = "OUTREACH"
	ModuleDocuments           = "DOCUMENTS"
	ModuleApprovals           = "APPROVALS"
	ModuleInvoices            = "INVOICES"
	ModulePayments            = "PAYMENTS"
	ModuleCarrierIntegrations = "CARRIER_INTEGRATIONS"
	ModuleSettings            = "SETTINGS"
)

// Standard Audit Actions representing controlled operation verbs.
const (
	ActionCreate            = "CREATE"
	ActionUpdate            = "UPDATE"
	ActionDelete            = "DELETE"
	ActionLogin             = "LOGIN"
	ActionLogout            = "LOGOUT"
	ActionLoginFailed       = "LOGIN_FAILED"
	ActionInvite            = "INVITE"
	ActionRoleChanged       = "ROLE_CHANGED"
	ActionPermissionChanged = "PERMISSION_CHANGED"
	ActionApprove           = "APPROVE"
	ActionReject            = "REJECT"
	ActionSend              = "SEND"
	ActionFollowUpSent      = "FOLLOW_UP_SENT"
	ActionConnect           = "CONNECT"
	ActionDisconnect        = "DISCONNECT"
	ActionEnable            = "ENABLE"
	ActionDisable           = "DISABLE"
	ActionSync              = "SYNC"
	ActionPaymentRecorded   = "PAYMENT_RECORDED"
	ActionExport            = "EXPORT"
)

// Result status of the audit event.
const (
	ResultSuccess = "SUCCESS"
	ResultFailed  = "FAILED"
)

// FieldChange represents an atomic field-level diff (Before vs After).
type FieldChange struct {
	Field  string      `json:"field"`
	Before interface{} `json:"before"`
	After  interface{} `json:"after"`
}

// AuditLog represents a canonical, immutable historical audit record.
type AuditLog struct {
	ID           int64                  `json:"id" db:"id"`
	OrgID        int64                  `json:"org_id" db:"org_id"`
	ActorID      *int64                 `json:"actor_id,omitempty" db:"user_id"`
	ActorType    string                 `json:"actor_type" db:"actor_type"` // USER, SYSTEM, AI_AGENT
	ActorName    string                 `json:"actor_name" db:"actor_name"`
	ActorRole    string                 `json:"actor_role,omitempty" db:"actor_role"`
	Action       string                 `json:"action" db:"action"`
	Module       string                 `json:"module" db:"module"`
	ResourceType string                 `json:"resource_type" db:"resource_type"`
	ResourceID   string                 `json:"resource_id" db:"resource_id"`
	ResourceName string                 `json:"resource_name,omitempty" db:"resource_name"`
	Description  string                 `json:"description" db:"description"`
	Result       string                 `json:"result" db:"result"` // SUCCESS, FAILED
	ErrorMessage string                 `json:"error_message,omitempty" db:"error_message"`
	BeforeData   map[string]interface{} `json:"before_data,omitempty"`
	AfterData    map[string]interface{} `json:"after_data,omitempty"`
	Changes      []FieldChange          `json:"changes,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    string                 `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// CreateAuditLogParams provides the input payload for the central audit service.
type CreateAuditLogParams struct {
	OrgID        int64                  `json:"org_id"`
	ActorID      *int64                 `json:"actor_id,omitempty"`
	ActorType    string                 `json:"actor_type,omitempty"` // Default: USER if ActorID != nil else SYSTEM
	ActorName    string                 `json:"actor_name,omitempty"`
	ActorRole    string                 `json:"actor_role,omitempty"`
	Action       string                 `json:"action"`
	Module       string                 `json:"module"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	ResourceName string                 `json:"resource_name,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Result       string                 `json:"result,omitempty"` // Default: SUCCESS
	ErrorMessage string                 `json:"error_message,omitempty"`
	Before       map[string]interface{} `json:"before,omitempty"`
	After        map[string]interface{} `json:"after,omitempty"`
	Changes      []FieldChange          `json:"changes,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
}

// AuditLogFilter specifies query criteria for retrieving audit logs.
type AuditLogFilter struct {
	OrgID        int64      `json:"org_id"`
	ActorID      *int64     `json:"actor_id,omitempty"`
	ActorType    string     `json:"actor_type,omitempty"`
	Module       string     `json:"module,omitempty"`
	Action       string     `json:"action,omitempty"`
	ResourceType string     `json:"resource_type,omitempty"`
	ResourceID   string     `json:"resource_id,omitempty"`
	Result       string     `json:"result,omitempty"`
	Search       string     `json:"search,omitempty"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Page         int        `json:"page"`
	Limit        int        `json:"limit"`
}

// AuditLogListResponse defines the paginated list response payload.
type AuditLogListResponse struct {
	Items      []AuditLog `json:"items"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	TotalPages int        `json:"total_pages"`
}
