package rbac

// RoleResponse represents a role along with its high-level details, returned to the client.
type RoleResponse struct {
	ID          int64  `json:"id" db:"id"`
	OrgID       int64  `json:"org_id" db:"org_id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
}

// PermissionNode represents a specific action permission on a resource.
type PermissionNode struct {
	Resource string `json:"resource" db:"resource"`
	Action   string `json:"action" db:"action"`
}

// RolePermissionsResponse represents a role and a flat list of its current permissions.
type RolePermissionsResponse struct {
	RoleID      int64            `json:"role_id"`
	Permissions []PermissionNode `json:"permissions"`
}

// UpdatePermissionsRequest represents the payload from the client to overwrite a role's permissions.
type UpdatePermissionsRequest struct {
	// Permissions is a list of resource/action pairs the role should have.
	// Any permissions not in this list will be revoked for this role.
	Permissions []PermissionNode `json:"permissions" binding:"required"`
}
