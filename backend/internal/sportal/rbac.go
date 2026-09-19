package sportal

import (
	"strings"
)

// SPortal Canonical Permission Identifiers
const (
	// Organizations
	PermOrgsView    = "organizations:view"
	PermOrgsCreate  = "organizations:create"
	PermOrgsUpdate  = "organizations:update"
	PermOrgsArchive = "organizations:archive"

	PermOrganizationsView   = PermOrgsView
	PermOrganizationsCreate = PermOrgsCreate
	PermOrganizationsUpdate = PermOrgsUpdate
	PermOrganizationsDelete = PermOrgsArchive

	// Onboarding
	PermOnboardingView     = "onboarding:view"
	PermOnboardingCreate   = "onboarding:create"
	PermOnboardingUpdate   = "onboarding:update"
	PermOnboardingComplete = "onboarding:complete"

	// Subscriptions
	PermSubscriptionsView       = "subscriptions:view"
	PermSubscriptionsCreate     = "subscriptions:create"
	PermSubscriptionsUpdate     = "subscriptions:update"
	PermSubscriptionsCancel     = "subscriptions:cancel"
	PermSubscriptionsRenew      = "subscriptions:renew"
	PermSubscriptionsPlanManage = "subscriptions:plan_manage"

	// Billing & Finance
	PermBillingView          = "billing:view"
	PermBillingManage        = "billing:manage"
	PermBillingSensitiveView = "billing:sensitive_view" // Sensitive: bank details, tax GSTIN/PAN secrets

	// Users & Roles
	PermUsersView       = "users:view"
	PermUsersCreate     = "users:create"
	PermUsersUpdate     = "users:update"
	PermUsersDisable    = "users:disable"
	PermUsersInvite     = "users:invite"
	PermUsersReactivate = "users:reactivate"
	PermRolesView       = "roles:view"

	// Customer Operations
	PermCustomerOpsView = "customer_operations:view"

	// Usage & Analytics
	PermUsageView = "usage:view"

	// Customer Health
	PermCustomerHealthView   = "customer_health:view"
	PermCustomerHealthManage = "customer_health:manage"

	// Integrations
	PermIntegrationsView   = "integrations:view"
	PermIntegrationsManage = "integrations:manage"

	// Documents & Compliance
	PermDocumentsView   = "documents:view"
	PermDocumentsManage = "documents:manage"

	// Support & Activity
	PermSupportView   = "support:view"
	PermSupportManage = "support:manage"

	// Audit
	PermAuditView = "audit:view"

	// SPortal Settings
	PermSettingsView   = "settings:view"
	PermSettingsManage = "settings:manage"

	// SPortal AI
	PermAIView   = "ai:view"
	PermAIUse    = "ai:use"
	PermAIManage = "ai:manage"

	// Platform Health & Ops
	PermPlatformHealthView = "platform_health:view"
	PermPlatformAdmin      = "platform_admin"
)

// Internal Staff Role Constants
const (
	RoleSuperAdmin      = "SUPER_ADMIN"
	RoleCEO             = "CEO"
	RoleOwner           = "OWNER"
	RoleAdmin           = "ADMIN"
	RoleSportalAdmin    = "SPORTAL_ADMIN"
	RoleCustomerSuccess = "CUSTOMER_SUCCESS"
	RoleFinance         = "FINANCE"
	RoleSupport         = "SUPPORT"
	RoleOperations      = "OPERATIONS"
	RoleTechnical       = "TECHNICAL"
)

// AllSPortalPermissions returns the complete master list of all canonical SPortal permissions.
func AllSPortalPermissions() []string {
	return []string{
		PermOrgsView, PermOrgsCreate, PermOrgsUpdate, PermOrgsArchive,
		PermOnboardingView, PermOnboardingCreate, PermOnboardingUpdate, PermOnboardingComplete,
		PermSubscriptionsView, PermSubscriptionsCreate, PermSubscriptionsUpdate, PermSubscriptionsCancel, PermSubscriptionsRenew, PermSubscriptionsPlanManage,
		PermBillingView, PermBillingManage, PermBillingSensitiveView,
		PermUsersView, PermUsersCreate, PermUsersUpdate, PermUsersDisable, PermUsersInvite, PermUsersReactivate, PermRolesView,
		PermCustomerOpsView,
		PermUsageView,
		PermCustomerHealthView, PermCustomerHealthManage,
		PermIntegrationsView, PermIntegrationsManage,
		PermDocumentsView, PermDocumentsManage,
		PermSupportView, PermSupportManage,
		PermAuditView,
		PermSettingsView, PermSettingsManage,
		PermAIView, PermAIUse, PermAIManage,
		PermPlatformHealthView, PermPlatformAdmin,
	}
}

// GetRolePermissions maps an internal role to its authorized SPortal permissions.
func GetRolePermissions(role string) []string {
	cleanRole := strings.ToUpper(strings.TrimSpace(role))

	switch cleanRole {
	case RoleSuperAdmin, RoleCEO, RoleOwner, "INTERNAL_STAFF":
		// Full access including sensitive financial records and platform administration
		return AllSPortalPermissions()

	case RoleAdmin, RoleSportalAdmin:
		// Administrator access to all operational and customer administration modules,
		// excluding raw sensitive billing secrets unless explicitly granted
		return []string{
			PermOrgsView, PermOrgsCreate, PermOrgsUpdate,
			PermOnboardingView, PermOnboardingCreate, PermOnboardingUpdate, PermOnboardingComplete,
			PermSubscriptionsView, PermSubscriptionsCreate, PermSubscriptionsUpdate, PermSubscriptionsCancel, PermSubscriptionsRenew, PermSubscriptionsPlanManage,
			PermBillingView, PermBillingManage,
			PermUsersView, PermUsersCreate, PermUsersUpdate, PermUsersDisable, PermUsersInvite, PermUsersReactivate, PermRolesView,
			PermCustomerOpsView,
			PermUsageView,
			PermCustomerHealthView, PermCustomerHealthManage,
			PermIntegrationsView, PermIntegrationsManage,
			PermDocumentsView, PermDocumentsManage,
			PermSupportView, PermSupportManage,
			PermAuditView,
			PermSettingsView, PermSettingsManage,
			PermAIView, PermAIUse, PermAIManage,
			PermPlatformHealthView,
		}

	case RoleCustomerSuccess:
		return []string{
			PermOrgsView,
			PermOnboardingView, PermOnboardingCreate, PermOnboardingUpdate,
			PermSubscriptionsView,
			PermUsersView, PermRolesView,
			PermCustomerOpsView,
			PermUsageView,
			PermCustomerHealthView, PermCustomerHealthManage,
			PermIntegrationsView,
			PermDocumentsView,
			PermSupportView,
			PermAIView, PermAIUse,
		}

	case RoleFinance:
		return []string{
			PermOrgsView,
			PermSubscriptionsView, PermSubscriptionsCreate, PermSubscriptionsUpdate, PermSubscriptionsCancel, PermSubscriptionsRenew, PermSubscriptionsPlanManage,
			PermBillingView, PermBillingManage, PermBillingSensitiveView,
			PermUsersView, PermRolesView,
			PermAuditView,
			PermDocumentsView,
		}

	case RoleSupport:
		return []string{
			PermOrgsView,
			PermUsersView, PermRolesView,
			PermCustomerOpsView,
			PermSupportView, PermSupportManage,
			PermDocumentsView,
			PermAuditView,
			PermUsageView,
			PermAIView, PermAIUse,
		}

	case RoleOperations:
		return []string{
			PermOrgsView,
			PermUsersView, PermRolesView,
			PermCustomerOpsView,
			PermUsageView,
			PermIntegrationsView,
			PermDocumentsView,
			PermAuditView,
		}

	case RoleTechnical:
		return []string{
			PermPlatformHealthView,
			PermIntegrationsView, PermIntegrationsManage,
			PermAIView, PermAIUse, PermAIManage,
			PermAuditView,
		}

	default:
		// Customer roles (e.g. SALES, PRICING, CUSTOMER_CONTACT, forwarder roles) get 0 SPortal permissions
		return []string{}
	}
}

// RoleHasPermission checks whether a given role possesses the specified permission.
func RoleHasPermission(role, permission string) bool {
	cleanRole := strings.ToUpper(strings.TrimSpace(role))
	if cleanRole == RoleSuperAdmin || cleanRole == RoleCEO || cleanRole == RoleOwner {
		return true
	}

	perms := GetRolePermissions(cleanRole)
	for _, p := range perms {
		if p == permission {
			return true
		}
		// Alias checks
		if permission == PermUsersInvite && p == PermUsersCreate {
			return true
		}
		if permission == PermUsersReactivate && p == PermUsersUpdate {
			return true
		}
		if permission == PermRolesView && p == PermUsersView {
			return true
		}
	}
	return false
}

// IsInternalStaffRole verifies if a role qualifies as internal LogisticsHQ personnel.
func IsInternalStaffRole(role string) bool {
	cleanRole := strings.ToUpper(strings.TrimSpace(role))
	switch cleanRole {
	case RoleSuperAdmin, RoleCEO, RoleOwner, RoleAdmin, RoleSportalAdmin,
		RoleCustomerSuccess, RoleFinance, RoleSupport, RoleOperations, RoleTechnical, "INTERNAL_STAFF":
		return true
	default:
		return false
	}
}

// IsInternalOrganization checks whether an org ID represents the internal LogisticsHQ organization.
// By platform convention, Org ID 1 is the primary LogisticsHQ internal operating entity.
func IsInternalOrganization(orgID int64) bool {
	return orgID == 1
}
