package rbac

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

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

func (s *service) SeedSystemPermissions(ctx context.Context) error {
	permissions := []struct {
		Resource string
		Action   string
		Desc     string
	}{
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
			VALUES (?, ?, ?) 
			ON DUPLICATE KEY UPDATE description = VALUES(description)
		`, p.Resource, p.Action, p.Desc)
		if err != nil {
			return fmt.Errorf("failed to seed permission %s:%s - %w", p.Resource, p.Action, err)
		}
	}
	return nil
}

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
			q, args, err := sqlx.In(`
				INSERT IGNORE INTO role_permissions (role_id, permission_id)
				SELECT ?, id FROM permissions WHERE resource IN (?)
			`, roleID, resources)
			if err != nil {
				return fmt.Errorf("failed to build role permission query: %w", err)
			}
			q = s.db.Rebind(q)
			_, err = s.db.ExecContext(ctx, q, args...)
			if err != nil {
				return fmt.Errorf("failed to attach permissions to %s - %w", r.Name, err)
			}
		}
	}
	return nil
}

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

func (s *service) GetRoles(ctx context.Context, orgID int64) ([]RoleResponse, error) {
	var roles []RoleResponse
	query := `SELECT id, org_id, name, description FROM roles WHERE org_id = ? ORDER BY id ASC`
	err := s.db.SelectContext(ctx, &roles, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch roles: %w", err)
	}
	if roles == nil {
		roles = []RoleResponse{}
	}
	return roles, nil
}

func (s *service) GetRolePermissions(ctx context.Context, orgID, roleID int64) (*RolePermissionsResponse, error) {
	var exists int
	err := s.db.GetContext(ctx, &exists, `SELECT 1 FROM roles WHERE id = ? AND org_id = ?`, roleID, orgID)
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
		WHERE rp.role_id = ?
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

func (s *service) UpdateRolePermissions(ctx context.Context, orgID, roleID int64, req UpdatePermissionsRequest) error {
	var exists int
	err := s.db.GetContext(ctx, &exists, `SELECT 1 FROM roles WHERE id = ? AND org_id = ?`, roleID, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("role not found in organization")
		}
		return fmt.Errorf("failed to verify role ownership: %w", err)
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

	return nil
}
