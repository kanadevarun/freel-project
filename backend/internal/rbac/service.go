package rbac

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/freel/backend/internal/common/events"
	"github.com/jmoiron/sqlx"
)

// Service defines methods for handling Role-Based Access Control logic.
type Service interface {
	// HasPermission checks if a specific role has a certain permission.
	// Simple meaning: It asks the database "is this person allowed to do this action on this page?"
	// Example: allowed, err := svc.HasPermission(ctx, "SALES", 123, "COMPANIES", "CREATE")
	HasPermission(ctx context.Context, roleName string, orgID int64, resource string, action string) (bool, error)

	// SeedSystemPermissions inserts the default 20 permissions if they don't exist.
	// Simple meaning: It loads all the basic rules into the system database when the server starts.
	// Example: err := rbacSvc.SeedSystemPermissions(ctx)
	SeedSystemPermissions(ctx context.Context) error

	// SeedDefaultRolesForOrg creates the standard roles for a new organization.
	// Simple meaning: When a new company signs up, it sets up their default job roles.
	// Example: err := rbacSvc.SeedDefaultRolesForOrg(ctx, 123)
	SeedDefaultRolesForOrg(ctx context.Context, orgID int64) error

	// GetRoles returns all roles for a specific organization.
	GetRoles(ctx context.Context, orgID int64) ([]RoleResponse, error)

	// GetRolePermissions returns all active permissions assigned to a specific role.
	GetRolePermissions(ctx context.Context, orgID, roleID int64) (*RolePermissionsResponse, error)

	// UpdateRolePermissions overwrites all permissions for a specific role with the new provided set.
	UpdateRolePermissions(ctx context.Context, orgID, roleID int64, req UpdatePermissionsRequest) error
}

type service struct {
	db *sqlx.DB
}

// NewService creates a new RBAC service.
// Simple meaning: Sets up the tool that checks user permissions against the database.
// Example: rbacSvc := NewService(db)
func NewService(db *sqlx.DB, bus events.Bus) Service {
	s := &service{db: db}

	// Listen for new orgs being created so we can seed default roles
	if bus != nil {
		bus.Subscribe(events.EventOrgCreated, func(e events.Event) {
			// e.Payload is the Organization struct from the organization module.
			// Instead of importing it, we can decode it via JSON or reflection,
			// but a simpler way is to just expect a map if the producer serializes it, 
			// or use JSON to safely extract the ID.
			// We'll run a background context for the DB inserts.
			go func() {
				// We assume the payload will serialize to JSON with an "id" field
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

// SeedSystemPermissions inserts the default 20 permissions if they don't exist.
func (s *service) SeedSystemPermissions(ctx context.Context) error {
	permissions := []struct {
		Resource string
		Action   string
		Desc     string
	}{
		// Add 20 system permissions
		{ResourceCompanies, ActionCreate, "Create companies"},
		{ResourceCompanies, ActionRead, "Read companies"},
		{ResourceCompanies, ActionUpdate, "Update companies"},
		{ResourceCompanies, ActionDelete, "Delete companies"},
		{ResourceLeads, ActionCreate, "Create leads"},
		{ResourceLeads, ActionRead, "Read leads"},
		{ResourceLeads, ActionUpdate, "Update leads"},
		{ResourceLeads, ActionDelete, "Delete leads"},
		{ResourceOpportunities, ActionCreate, "Create opportunities"},
		{ResourceOpportunities, ActionRead, "Read opportunities"},
		{ResourceOpportunities, ActionUpdate, "Update opportunities"},
		{ResourceOpportunities, ActionDelete, "Delete opportunities"},
		{ResourceRFQs, ActionCreate, "Create RFQs"},
		{ResourceRFQs, ActionRead, "Read RFQs"},
		{ResourceRFQs, ActionUpdate, "Update RFQs"},
		{ResourceRFQs, ActionDelete, "Delete RFQs"},
		{ResourceOutreach, ActionCreate, "Create campaigns"},
		{ResourceOutreach, ActionRead, "Read campaigns"},
		{ResourceOutreach, ActionUpdate, "Update campaigns"},
		{ResourceOutreach, ActionDelete, "Delete campaigns"},
	}

	for _, p := range permissions {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO permissions (resource, action, description) 
			VALUES ($1, $2, $3) 
			ON CONFLICT (resource, action) DO NOTHING
		`, p.Resource, p.Action, p.Desc)
		if err != nil {
			return fmt.Errorf("failed to seed permission %s:%s - %w", p.Resource, p.Action, err)
		}
	}
	return nil
}

// SeedDefaultRolesForOrg creates the standard roles for a new organization.
func (s *service) SeedDefaultRolesForOrg(ctx context.Context, orgID int64) error {
	roles := []struct {
		Name string
		Desc string
	}{
		{RoleSuperAdmin, "Full access to everything"},
		{RoleSales, "Can manage leads and opportunities"},
		{RolePricing, "Can manage pricing and quotes"},
		{RoleOperations, "Can manage operations and shipments"},
		{RoleFinance, "Can manage invoices and billing"},
		{RoleCustomer, "External customer portal access"},
	}

	for _, r := range roles {
		var roleID int64
		// Insert role and return ID
		err := s.db.QueryRowContext(ctx, `
			INSERT INTO roles (org_id, name, description) 
			VALUES ($1, $2, $3) 
			ON CONFLICT (org_id, name) DO NOTHING
			RETURNING id
		`, orgID, r.Name, r.Desc).Scan(&roleID)
		
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Role already existed
				continue
			}
			return fmt.Errorf("failed to seed role %s - %w", r.Name, err)
		}

		var resources []string
		switch r.Name {
		case RoleSuperAdmin:
			resources = []string{ResourceCompanies, ResourceLeads, ResourceOpportunities, ResourceRFQs, ResourceOutreach}
		case RoleSales:
			resources = []string{ResourceCompanies, ResourceLeads, ResourceOpportunities}
		case RolePricing:
			resources = []string{ResourceCompanies, ResourceRFQs}
		case RoleOperations:
			resources = []string{ResourceCompanies} 
		case RoleFinance:
			resources = []string{ResourceCompanies} 
		}

		if len(resources) > 0 && roleID != 0 {
			query := `
				INSERT INTO role_permissions (role_id, permission_id)
				SELECT $1, id FROM permissions WHERE resource = ANY($2)
				ON CONFLICT DO NOTHING
			`
			_, err = s.db.ExecContext(ctx, query, roleID, pq.Array(resources))
			if err != nil {
				return fmt.Errorf("failed to attach permissions to %s - %w", r.Name, err)
			}
		}
	}
	return nil
}

// HasPermission checks if a specific role has a certain permission.
// Simple meaning: It asks the database "is this person allowed to do this action on this page?"
// Example: allowed, err := svc.HasPermission(ctx, "SALES", 123, "COMPANIES", "CREATE")
func (s *service) HasPermission(ctx context.Context, roleName string, orgID int64, resource string, action string) (bool, error) {
	query := `
		SELECT 1 
		FROM roles r
		JOIN role_permissions rp ON r.id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE r.name = $1 AND r.org_id = $2 AND p.resource = $3 AND p.action = $4
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

// GetRoles fetches all roles associated with a specific organization.
func (s *service) GetRoles(ctx context.Context, orgID int64) ([]RoleResponse, error) {
	var roles []RoleResponse
	query := `SELECT id, org_id, name, description FROM roles WHERE org_id = $1 ORDER BY id ASC`
	err := s.db.SelectContext(ctx, &roles, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch roles: %w", err)
	}
	if roles == nil {
		roles = []RoleResponse{}
	}
	return roles, nil
}

// GetRolePermissions fetches all permissions currently attached to a given role.
// It verifies the role actually belongs to the given orgID to prevent unauthorized access.
func (s *service) GetRolePermissions(ctx context.Context, orgID, roleID int64) (*RolePermissionsResponse, error) {
	// First, verify the role belongs to the org
	var exists int
	err := s.db.GetContext(ctx, &exists, `SELECT 1 FROM roles WHERE id = $1 AND org_id = $2`, roleID, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("role not found in organization")
		}
		return nil, fmt.Errorf("failed to verify role ownership: %w", err)
	}

	var perms []PermissionNode
	query := `
		SELECT p.resource, p.action 
		FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		WHERE rp.role_id = $1
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
		Permissions: perms,
	}, nil
}

// UpdateRolePermissions completely overwrites a role's permissions using a transaction.
// It deletes all existing mappings for the role, then inserts the new ones.
func (s *service) UpdateRolePermissions(ctx context.Context, orgID, roleID int64, req UpdatePermissionsRequest) error {
	// 1. Verify role belongs to org
	var exists int
	err := s.db.GetContext(ctx, &exists, `SELECT 1 FROM roles WHERE id = $1 AND org_id = $2`, roleID, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("role not found in organization")
		}
		return fmt.Errorf("failed to verify role ownership: %w", err)
	}

	// 2. Begin transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 3. Delete existing role_permissions
	_, err = tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID)
	if err != nil {
		return fmt.Errorf("failed to clear existing permissions: %w", err)
	}

	// 4. Insert new role_permissions if any are provided
	if len(req.Permissions) > 0 {
		// Prepare a batch insert by looking up the permission IDs based on resource/action
		insertQuery := `
			INSERT INTO role_permissions (role_id, permission_id)
			SELECT $1, id FROM permissions WHERE resource = $2 AND action = $3
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

	// 5. Commit
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit permission updates: %w", err)
	}

	return nil
}
