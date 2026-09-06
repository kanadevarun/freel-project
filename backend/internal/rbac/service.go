package rbac

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/common/events"
	"github.com/jmoiron/sqlx"
)

// Service defines methods for handling Role-Based Access Control logic.
type Service interface {
	HasPermission(ctx context.Context, roleName string, orgID int64, resource string, action string) (bool, error)
	SeedSystemPermissions(ctx context.Context) error
	SeedDefaultRolesForOrg(ctx context.Context, orgID int64) error
	GetRoles(ctx context.Context, orgID int64) ([]RoleResponse, error)
	GetRolePermissions(ctx context.Context, orgID, roleID int64) (*RolePermissionsResponse, error)
	UpdateRolePermissions(ctx context.Context, orgID, roleID int64, req UpdatePermissionsRequest) error
	GetStats(ctx context.Context, orgID int64) (*StatsResponse, error)
	CreateRole(ctx context.Context, orgID int64, req CreateRoleRequest) (*RoleResponse, error)
	UpdateRole(ctx context.Context, orgID, roleID int64, req UpdateRoleRequest) error
	DeleteRole(ctx context.Context, orgID, roleID int64) error
}

type service struct {
	db *sqlx.DB
}

func NewService(db *sqlx.DB, bus events.Bus) Service {
	s := &service{db: db}

	if bus != nil {
		bus.Subscribe(events.EventOrgCreated, func(e events.Event) {
			go func() {
				b, err := json.Marshal(e.Payload)
				if err != nil {
					return
				}
				var orgData struct {
					ID int64 `json:"id"`
				}
				if err := json.Unmarshal(b, &orgData); err == nil && orgData.ID != 0 {
					s.SeedDefaultRolesForOrg(context.Background(), orgData.ID)
				}
			}()
		})
	}

	return s
}

// coreResources is the canonical list of 10 resources in the RBAC system.
var coreResources = []string{
	ResourceCompanies, ResourceLeads, ResourceOpportunities,
	ResourceRFQs, ResourceOutreach, ResourceShipments,
	ResourceDocuments, ResourceFinance, ResourceUsers, ResourceSettings,
}

// coreActions is the canonical list of 4 actions.
var coreActions = []string{ActionCreate, ActionRead, ActionUpdate, ActionDelete}

// SeedSystemPermissions ensures the 40 canonical permissions exist (idempotent).
func (s *service) SeedSystemPermissions(ctx context.Context) error {
	descriptions := map[string]string{
		"COMPANIES.CREATE": "Create company records", "COMPANIES.READ": "View company records",
		"COMPANIES.UPDATE": "Edit company records", "COMPANIES.DELETE": "Delete company records",
		"LEADS.CREATE": "Create leads", "LEADS.READ": "View leads",
		"LEADS.UPDATE": "Edit leads", "LEADS.DELETE": "Delete leads",
		"OPPORTUNITIES.CREATE": "Create sales opportunities", "OPPORTUNITIES.READ": "View sales opportunities",
		"OPPORTUNITIES.UPDATE": "Edit sales opportunities", "OPPORTUNITIES.DELETE": "Delete sales opportunities",
		"RFQS.CREATE": "Create RFQs", "RFQS.READ": "View RFQs",
		"RFQS.UPDATE": "Edit RFQs", "RFQS.DELETE": "Delete RFQs",
		"OUTREACH.CREATE": "Create outreach campaigns", "OUTREACH.READ": "View outreach campaigns",
		"OUTREACH.UPDATE": "Edit outreach campaigns", "OUTREACH.DELETE": "Delete outreach campaigns",
		"SHIPMENTS.CREATE": "Create shipments", "SHIPMENTS.READ": "View shipments",
		"SHIPMENTS.UPDATE": "Edit shipments", "SHIPMENTS.DELETE": "Delete shipments",
		"DOCUMENTS.CREATE": "Create documents", "DOCUMENTS.READ": "View documents",
		"DOCUMENTS.UPDATE": "Edit documents", "DOCUMENTS.DELETE": "Delete documents",
		"FINANCE.CREATE": "Create finance records", "FINANCE.READ": "View finance records",
		"FINANCE.UPDATE": "Edit finance records", "FINANCE.DELETE": "Delete finance records",
		"USERS.CREATE": "Invite/create users", "USERS.READ": "View users and team",
		"USERS.UPDATE": "Edit user profiles and roles", "USERS.DELETE": "Remove users from organization",
		"SETTINGS.CREATE": "Create settings entries", "SETTINGS.READ": "View settings",
		"SETTINGS.UPDATE": "Edit settings", "SETTINGS.DELETE": "Delete settings entries",
	}

	for _, resource := range coreResources {
		for _, action := range coreActions {
			key := resource + "." + action
			desc := descriptions[key]
			_, err := s.db.ExecContext(ctx, `
				INSERT INTO permissions (resource, action, description) 
				VALUES (?, ?, ?) 
				ON DUPLICATE KEY UPDATE description = VALUES(description)
			`, resource, action, desc)
			if err != nil {
				return fmt.Errorf("failed to seed permission %s:%s - %w", resource, action, err)
			}
		}
	}
	return nil
}

// defaultRolePermissions defines least-privilege permission sets for each default role.
var defaultRolePermissions = map[string]map[string][]string{
	RoleSuperAdmin: {
		ResourceCompanies: coreActions, ResourceLeads: coreActions,
		ResourceOpportunities: coreActions, ResourceRFQs: coreActions,
		ResourceOutreach: coreActions, ResourceShipments: coreActions,
		ResourceDocuments: coreActions, ResourceFinance: coreActions,
		ResourceUsers: coreActions, ResourceSettings: coreActions,
	},
	RoleSales: {
		ResourceCompanies:     {ActionCreate, ActionRead, ActionUpdate},
		ResourceLeads:         {ActionCreate, ActionRead, ActionUpdate},
		ResourceOpportunities: {ActionCreate, ActionRead, ActionUpdate},
		ResourceRFQs:          {ActionCreate, ActionRead, ActionUpdate},
		ResourceOutreach:      {ActionCreate, ActionRead, ActionUpdate},
		ResourceShipments:     {ActionRead},
		ResourceDocuments:     {ActionRead},
		ResourceFinance:       {ActionRead},
	},
	RolePricing: {
		ResourceRFQs:          {ActionCreate, ActionRead, ActionUpdate},
		ResourceCompanies:     {ActionCreate, ActionRead, ActionUpdate},
		ResourceOpportunities: {ActionCreate, ActionRead, ActionUpdate},
		ResourceShipments:     {ActionRead},
		ResourceDocuments:     {ActionRead},
	},
	RoleOperations: {
		ResourceShipments: {ActionCreate, ActionRead, ActionUpdate},
		ResourceDocuments: {ActionCreate, ActionRead, ActionUpdate},
		ResourceCompanies: {ActionRead},
		ResourceRFQs:      {ActionRead},
	},
	RoleFinance: {
		ResourceFinance:   {ActionCreate, ActionRead, ActionUpdate},
		ResourceCompanies: {ActionRead},
		ResourceShipments: {ActionRead},
		ResourceDocuments: {ActionRead},
	},
	RoleDocumentation: {
		ResourceDocuments: {ActionCreate, ActionRead, ActionUpdate},
		ResourceShipments: {ActionRead},
		ResourceCompanies: {ActionRead},
	},
	RoleHR: {
		ResourceUsers:     {ActionCreate, ActionRead, ActionUpdate},
		ResourceCompanies: {ActionRead},
	},
}

// SeedDefaultRolesForOrg creates the 7 default roles for a new organization and assigns permissions.
func (s *service) SeedDefaultRolesForOrg(ctx context.Context, orgID int64) error {
	defaultRoles := []struct {
		Name string
		Desc string
	}{
		{RoleSuperAdmin, "Full access to everything"},
		{RoleSales, "Manage leads, RFQs and customers"},
		{RolePricing, "Manage pricing and quotations"},
		{RoleOperations, "Manage shipments and tracking"},
		{RoleFinance, "Manage invoices and payments"},
		{RoleDocumentation, "Manage documents and compliance"},
		{RoleHR, "Manage HR and team data"},
	}

	for _, r := range defaultRoles {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO roles (org_id, name, description) 
			VALUES (?, ?, ?) 
			ON DUPLICATE KEY UPDATE description = VALUES(description)
		`, orgID, r.Name, r.Desc)
		if err != nil {
			return fmt.Errorf("failed to seed role %s - %w", r.Name, err)
		}

		var roleID int64
		err = s.db.QueryRowContext(ctx, `SELECT id FROM roles WHERE org_id = ? AND name = ? LIMIT 1`, orgID, r.Name).Scan(&roleID)
		if err != nil {
			return fmt.Errorf("failed to get role id for %s - %w", r.Name, err)
		}

		permMap, ok := defaultRolePermissions[r.Name]
		if !ok || roleID == 0 {
			continue
		}

		// Build the desired set of permission IDs from the canonical definition
		var desiredIDs []int64
		for resource, actions := range permMap {
			for _, action := range actions {
				var permID int64
				err = s.db.QueryRowContext(ctx,
					`SELECT id FROM permissions WHERE resource = ? AND action = ?`,
					resource, action).Scan(&permID)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						continue // permission not in catalog yet — skip
					}
					return fmt.Errorf("failed to resolve permission %s:%s - %w", resource, action, err)
				}
				desiredIDs = append(desiredIDs, permID)
			}
		}

		if len(desiredIDs) == 0 {
			continue
		}

		// SYNC STEP 1: Remove any existing permissions NOT in the desired set.
		// This ensures that if the definition shrinks (e.g., SALES loses FINANCE.READ),
		// the DB reflects the updated definition on re-seed.
		q, args, err := sqlx.In(
			`DELETE FROM role_permissions WHERE role_id = ? AND permission_id NOT IN (?)`,
			roleID, desiredIDs)
		if err != nil {
			return fmt.Errorf("failed to build sync delete for %s: %w", r.Name, err)
		}
		q = s.db.Rebind(q)
		if _, err = s.db.ExecContext(ctx, q, args...); err != nil {
			return fmt.Errorf("failed to sync delete permissions for %s: %w", r.Name, err)
		}

		// SYNC STEP 2: Insert any permissions in the desired set not already assigned.
		for _, permID := range desiredIDs {
			_, err = s.db.ExecContext(ctx,
				`INSERT IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)`,
				roleID, permID)
			if err != nil {
				return fmt.Errorf("failed to assign permission %d to %s: %w", permID, r.Name, err)
			}
		}
	}
	return nil
}

// HasPermission checks whether a named role in an organization has a specific resource/action permission.
func (s *service) HasPermission(ctx context.Context, roleName string, orgID int64, resource string, action string) (bool, error) {
	query := `
		SELECT 1 
		FROM roles r
		JOIN role_permissions rp ON r.id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE r.name = ? AND r.org_id = ? AND p.resource = ? AND p.action = ?
	`
	var exists int
	err := s.db.GetContext(ctx, &exists, query, roleName, orgID, resource, action)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return exists == 1, nil
}

// GetRoles returns all roles for an organization with dynamic permission_count.
func (s *service) GetRoles(ctx context.Context, orgID int64) ([]RoleResponse, error) {
	var roles []RoleResponse
	query := `
		SELECT r.id, r.org_id, r.name, COALESCE(r.description, '') AS description,
		       COUNT(rp.permission_id) AS permission_count
		FROM roles r
		LEFT JOIN role_permissions rp ON r.id = rp.role_id
		WHERE r.org_id = ?
		GROUP BY r.id, r.org_id, r.name, r.description
		ORDER BY r.id ASC
	`
	err := s.db.SelectContext(ctx, &roles, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch roles: %w", err)
	}
	if roles == nil {
		roles = []RoleResponse{}
	}
	return roles, nil
}

// GetRolePermissions returns the permissions assigned to a role, verified to belong to the org.
func (s *service) GetRolePermissions(ctx context.Context, orgID, roleID int64) (*RolePermissionsResponse, error) {
	var roleName string
	err := s.db.QueryRowContext(ctx, `SELECT name FROM roles WHERE id = ? AND org_id = ?`, roleID, orgID).Scan(&roleName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("role not found in organization")
		}
		return nil, fmt.Errorf("failed to verify role ownership: %w", err)
	}

	var perms []PermissionNode
	query := `
		SELECT p.id, p.resource, p.action 
		FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		WHERE rp.role_id = ?
		ORDER BY p.resource, p.action
	`
	err = s.db.SelectContext(ctx, &perms, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch role permissions: %w", err)
	}
	if perms == nil {
		perms = []PermissionNode{}
	}

	return &RolePermissionsResponse{
		RoleID:      roleID,
		RoleName:    roleName,
		Permissions: perms,
	}, nil
}

// UpdateRolePermissions replaces a role's permissions atomically within a transaction.
func (s *service) UpdateRolePermissions(ctx context.Context, orgID, roleID int64, req UpdatePermissionsRequest) error {
	var roleName string
	err := s.db.GetContext(ctx, &roleName, `SELECT name FROM roles WHERE id = ? AND org_id = ?`, roleID, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("role not found in organization")
		}
		return fmt.Errorf("failed to verify role ownership: %w", err)
	}

	if roleName == RoleSuperAdmin {
		return fmt.Errorf("cannot modify permissions for SUPER_ADMIN role")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = ?`, roleID)
	if err != nil {
		return fmt.Errorf("failed to clear existing permissions: %w", err)
	}

	if len(req.Permissions) > 0 {
		insertQuery := `
			INSERT INTO role_permissions (role_id, permission_id)
			SELECT ?, id FROM permissions WHERE resource = ? AND action = ?
		`
		stmt, err := tx.PrepareContext(ctx, insertQuery)
		if err != nil {
			return fmt.Errorf("failed to prepare insert statement: %w", err)
		}
		defer stmt.Close()

		for _, p := range req.Permissions {
			_, err = stmt.ExecContext(ctx, roleID, p.Resource, p.Action)
			if err != nil {
				return fmt.Errorf("failed to map permission %s:%s - %w", p.Resource, p.Action, err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit permission updates: %w", err)
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		Action:       domain.ActionPermissionChanged,
		Module:       domain.ModuleRolesPermissions,
		ResourceType: "ROLE",
		ResourceID:   fmt.Sprintf("%d", roleID),
		ResourceName: roleName,
		Description:  fmt.Sprintf("Updated permissions for role %s", roleName),
		Result:       domain.ResultSuccess,
	})

	return nil
}

// GetStats returns dashboard statistics for the Roles & Permissions page.
func (s *service) GetStats(ctx context.Context, orgID int64) (*StatsResponse, error) {
	var totalRoles int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM roles WHERE org_id = ?`, orgID).Scan(&totalRoles); err != nil {
		return nil, fmt.Errorf("failed to count roles: %w", err)
	}

	var totalPermissions int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM permissions WHERE resource IN ('COMPANIES','LEADS','OPPORTUNITIES','RFQS','OUTREACH','SHIPMENTS','DOCUMENTS','FINANCE','USERS','SETTINGS') AND action IN ('CREATE','READ','UPDATE','DELETE')`).Scan(&totalPermissions); err != nil {
		return nil, fmt.Errorf("failed to count permissions: %w", err)
	}

	var activeMembers int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM org_members WHERE org_id = ? AND status = 'ACTIVE'`, orgID).Scan(&activeMembers); err != nil {
		return nil, fmt.Errorf("failed to count active members: %w", err)
	}

	// System Coverage: % of the 40 canonical permissions that are assigned
	// to at least one role in this org. This gives a real measure of how
	// completely the permission catalog is covered by configured roles.
	var coveredPermissions int
	coverageQuery := `
		SELECT COUNT(DISTINCT rp.permission_id)
		FROM role_permissions rp
		JOIN roles r ON rp.role_id = r.id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE r.org_id = ?
		  AND p.resource IN ('COMPANIES','LEADS','OPPORTUNITIES','RFQS','OUTREACH','SHIPMENTS','DOCUMENTS','FINANCE','USERS','SETTINGS')
		  AND p.action IN ('CREATE','READ','UPDATE','DELETE')
	`
	if err := s.db.QueryRowContext(ctx, coverageQuery, orgID).Scan(&coveredPermissions); err != nil {
		return nil, fmt.Errorf("failed to calculate system coverage: %w", err)
	}
	systemCoverage := 0
	if totalPermissions > 0 {
		systemCoverage = (coveredPermissions * 100) / totalPermissions
	}

	var roleCounts []RoleCountEntry
	query := `
		SELECT r.name AS role_name, COUNT(om.id) AS member_count
		FROM roles r
		LEFT JOIN org_members om ON om.role_id = r.id AND om.org_id = ? AND om.status = 'ACTIVE'
		WHERE r.org_id = ?
		GROUP BY r.id, r.name
		ORDER BY r.id ASC
	`
	if err := s.db.SelectContext(ctx, &roleCounts, query, orgID, orgID); err != nil {
		return nil, fmt.Errorf("failed to fetch role counts: %w", err)
	}
	if roleCounts == nil {
		roleCounts = []RoleCountEntry{}
	}

	return &StatsResponse{
		TotalRoles:       totalRoles,
		TotalPermissions: totalPermissions,
		ActiveMembers:    activeMembers,
		SystemCoverage:   systemCoverage,
		RoleCounts:       roleCounts,
	}, nil
}

// CreateRole creates a new custom role within an organization and assigns its initial permissions.
func (s *service) CreateRole(ctx context.Context, orgID int64, req CreateRoleRequest) (*RoleResponse, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("role name is required")
	}

	var duplicateCount int
	err := s.db.GetContext(ctx, &duplicateCount, `SELECT COUNT(*) FROM roles WHERE org_id = ? AND name = ?`, orgID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicate role name: %w", err)
	}
	if duplicateCount > 0 {
		return nil, fmt.Errorf("role with this name already exists in the organization")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO roles (org_id, name, description) VALUES (?, ?, ?)
	`, orgID, req.Name, req.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	roleID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get new role id: %w", err)
	}

	for _, p := range req.Permissions {
		_, err = tx.ExecContext(ctx, `
			INSERT IGNORE INTO role_permissions (role_id, permission_id)
			SELECT ?, id FROM permissions WHERE resource = ? AND action = ?
		`, roleID, p.Resource, p.Action)
		if err != nil {
			return nil, fmt.Errorf("failed to assign permission %s:%s - %w", p.Resource, p.Action, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit role creation: %w", err)
	}

	// Query actual inserted permission count from DB — safer than trusting
	// len(req.Permissions) which may include invalid/nonexistent resource+action pairs
	// that were silently ignored by INSERT IGNORE.
	var actualPermCount int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM role_permissions WHERE role_id = ?`, roleID,
	).Scan(&actualPermCount); err != nil {
		// Non-fatal: fall back to len(req.Permissions) if the count query fails
		actualPermCount = len(req.Permissions)
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		Action:       domain.ActionCreate,
		Module:       domain.ModuleRolesPermissions,
		ResourceType: "ROLE",
		ResourceID:   fmt.Sprintf("%d", roleID),
		ResourceName: req.Name,
		Description:  fmt.Sprintf("Created role %s", req.Name),
		Result:       domain.ResultSuccess,
	})

	return &RoleResponse{
		ID:              roleID,
		OrgID:           orgID,
		Name:            req.Name,
		Description:     req.Description,
		PermissionCount: actualPermCount,
	}, nil
}

// isSystemRole checks if a role name is one of the protected system roles
func isSystemRole(roleName string) bool {
	systemRoles := map[string]bool{
		RoleSuperAdmin:    true,
		RoleSales:         true,
		RolePricing:       true,
		RoleOperations:    true,
		RoleFinance:       true,
		RoleDocumentation: true,
		RoleHR:            true,
	}
	return systemRoles[roleName]
}

// UpdateRole updates the metadata (name, description) of a custom role.
func (s *service) UpdateRole(ctx context.Context, orgID, roleID int64, req UpdateRoleRequest) error {
	if req.Name == "" {
		return fmt.Errorf("role name is required")
	}

	var currentName string
	err := s.db.GetContext(ctx, &currentName, `SELECT name FROM roles WHERE id = ? AND org_id = ?`, roleID, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("role not found in organization")
		}
		return fmt.Errorf("failed to fetch role: %w", err)
	}

	if isSystemRole(currentName) {
		return fmt.Errorf("cannot modify metadata of a system role")
	}

	var duplicateCount int
	err = s.db.GetContext(ctx, &duplicateCount, `SELECT COUNT(*) FROM roles WHERE org_id = ? AND name = ? AND id != ?`, orgID, req.Name, roleID)
	if err != nil {
		return fmt.Errorf("failed to check duplicate role name: %w", err)
	}
	if duplicateCount > 0 {
		return fmt.Errorf("role with this name already exists in the organization")
	}

	_, err = s.db.ExecContext(ctx, `UPDATE roles SET name = ?, description = ? WHERE id = ?`, req.Name, req.Description, roleID)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		Action:       domain.ActionUpdate,
		Module:       domain.ModuleRolesPermissions,
		ResourceType: "ROLE",
		ResourceID:   fmt.Sprintf("%d", roleID),
		ResourceName: req.Name,
		Description:  fmt.Sprintf("Updated role %s", req.Name),
		Before:       map[string]interface{}{"name": currentName},
		After:        map[string]interface{}{"name": req.Name},
		Result:       domain.ResultSuccess,
	})

	return nil
}

// DeleteRole removes a custom role and its permissions if it's not assigned to users.
func (s *service) DeleteRole(ctx context.Context, orgID, roleID int64) error {
	var currentName string
	err := s.db.GetContext(ctx, &currentName, `SELECT name FROM roles WHERE id = ? AND org_id = ?`, roleID, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("role not found in organization")
		}
		return fmt.Errorf("failed to fetch role: %w", err)
	}

	if isSystemRole(currentName) {
		return fmt.Errorf("cannot delete a system role")
	}

	var assignedCount int
	err = s.db.GetContext(ctx, &assignedCount, `SELECT COUNT(*) FROM org_members WHERE role_id = ?`, roleID)
	if err != nil {
		return fmt.Errorf("failed to check role assignments: %w", err)
	}
	if assignedCount > 0 {
		return fmt.Errorf("role is assigned to users")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = ?`, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role permissions: %w", err)
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM roles WHERE id = ?`, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit role deletion: %w", err)
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		Action:       domain.ActionDelete,
		Module:       domain.ModuleRolesPermissions,
		ResourceType: "ROLE",
		ResourceID:   fmt.Sprintf("%d", roleID),
		ResourceName: currentName,
		Description:  fmt.Sprintf("Deleted role %s", currentName),
		Result:       domain.ResultSuccess,
	})

	return nil
}
