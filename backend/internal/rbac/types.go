package rbac

// RoleResponse represents a role along with its high-level details, returned to the client.
type RoleResponse struct {
	ID              int64  `json:"id" db:"id"`
	OrgID           int64  `json:"org_id" db:"org_id"`
	Name            string `json:"name" db:"name"`
	Description     string `json:"description" db:"description"`
	PermissionCount int    `json:"permission_count" db:"permission_count"`
}

// PermissionNode represents a specific action permission on a resource.
type PermissionNode struct {
	ID       int64  `json:"id,omitempty" db:"id"`
	Resource string `json:"resource" db:"resource"`
	Action   string `json:"action" db:"action"`
}

// RolePermissionsResponse represents a role and a flat list of its current permissions.
type RolePermissionsResponse struct {
	RoleID      int64            `json:"role_id"`
	RoleName    string           `json:"role_name"`
	Permissions []PermissionNode `json:"permissions"`
}

// UpdatePermissionsRequest represents the payload from the client to overwrite a role's permissions.
type UpdatePermissionsRequest struct {
	// Permissions is a list of resource/action pairs the role should have.
	// Any permissions not in this list will be revoked for this role.
	Permissions []PermissionNode `json:"permissions" binding:"required"`
}

// CreateRoleRequest is the payload for creating a new custom role with initial permissions.
type CreateRoleRequest struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Permissions []PermissionNode `json:"permissions"`
}

// UpdateRoleRequest is the payload for updating a custom role's metadata.
type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// RoleCountEntry holds user count for a single role, used in stats.
type RoleCountEntry struct {
	RoleName string `json:"role" db:"role_name"`
	Count    int    `json:"count" db:"member_count"`
}

// StatsResponse is returned by GET /api/v1/roles/stats.
type StatsResponse struct {
	TotalRoles       int              `json:"total_roles"`
	TotalPermissions int              `json:"total_permissions"`
	ActiveMembers    int              `json:"active_members"`
	SystemCoverage   int              `json:"system_coverage"`
	RoleCounts       []RoleCountEntry `json:"role_counts"`
}
