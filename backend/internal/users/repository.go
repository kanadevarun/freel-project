package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Repository defines the interface for data access operations related to the users module.
// It includes methods for managing organization members and pending invitations.
type Repository interface {
	// CreateInvitation inserts a new invitation record into the database.
	// It returns an error if the insertion fails (e.g., due to a unique constraint violation on email/org).
	CreateInvitation(ctx context.Context, inv *Invitation) error

	// GetInvitationByToken retrieves an invitation record using the secure token string.
	// It returns the invitation if found, or an error if no matching record exists.
	GetInvitationByToken(ctx context.Context, token string) (*Invitation, error)

	// DeleteInvitation removes an invitation record from the database by its ID.
	// This is typically called after an invitation is successfully accepted or explicitly canceled.
	DeleteInvitation(ctx context.Context, id int64) error

	// ListInvitations retrieves all pending invitations for a given organization.
	ListInvitations(ctx context.Context, orgID int64) ([]InvitationResponse, error)

	// ListOrgMembers retrieves all active members for a given organization ID.
	// It performs a join between the org_members, users, and roles tables to return a comprehensive OrgMemberResponse array.
	ListOrgMembers(ctx context.Context, orgID int64) ([]OrgMemberResponse, error)

	// UpdateMemberRole modifies the assigned role for a specific user within an organization.
	// It requires the organization ID, user ID, and the new role ID.
	UpdateMemberRole(ctx context.Context, orgID, userID, roleID int64) error

	// RemoveMember deletes an organization membership record, effectively removing the user's access to the organization.
	// Note: This does not delete the user record itself, only their association with the organization.
	RemoveMember(ctx context.Context, orgID, userID int64) error
	
	// CreateUser inserts a newly registered user into the users table.
	// It returns the ID of the newly created user.
	CreateUser(ctx context.Context, user *User) (int64, error)
	
	// GetUserByEmail retrieves a user record by their email address.
	GetUserByEmail(ctx context.Context, email string) (*User, error)

	// AddUserToOrg creates a new membership record linking a user to an organization with a specific role.
	AddUserToOrg(ctx context.Context, member *OrgMember) error
}

// repositoryImpl is the concrete implementation of the Repository interface using sqlx.
type repositoryImpl struct {
	db *sqlx.DB
}

// NewRepository creates and returns a new instance of the users repository implementation.
// It requires a connected sqlx.DB instance.
func NewRepository(db *sqlx.DB) Repository {
	return &repositoryImpl{db: db}
}

// CreateInvitation inserts a new invitation into the 'invitations' table.
// It uses named parameters from the Invitation struct to execute the query.
func (r *repositoryImpl) CreateInvitation(ctx context.Context, inv *Invitation) error {
	query := `
		INSERT INTO invitations (org_id, role_id, email, token, expires_at)
		VALUES (:org_id, :role_id, :email, :token, :expires_at)
		RETURNING id, created_at, updated_at
	`
	rows, err := r.db.NamedQueryContext(ctx, query, inv)
	if err != nil {
		return fmt.Errorf("failed to insert invitation: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&inv.ID, &inv.CreatedAt, &inv.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan returned invitation ID: %w", err)
		}
	}
	return nil
}

// GetInvitationByToken queries the 'invitations' table for a record matching the provided secure token.
// It returns an error (specifically sql.ErrNoRows) if the token is invalid or the invitation does not exist.
func (r *repositoryImpl) GetInvitationByToken(ctx context.Context, token string) (*Invitation, error) {
	var inv Invitation
	query := `SELECT id, org_id, role_id, email, token, expires_at, created_at FROM invitations WHERE token = $1`
	err := r.db.GetContext(ctx, &inv, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invitation not found")
		}
		return nil, fmt.Errorf("failed to fetch invitation by token: %w", err)
	}
	return &inv, nil
}

// DeleteInvitation removes an invitation from the 'invitations' table by its primary key ID.
func (r *repositoryImpl) DeleteInvitation(ctx context.Context, id int64) error {
	query := `DELETE FROM invitations WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete invitation: %w", err)
	}
	return nil
}

// ListInvitations fetches all active invitations for the given orgID,
// joining with the roles table to get the role name.
func (r *repositoryImpl) ListInvitations(ctx context.Context, orgID int64) ([]InvitationResponse, error) {
	var invites []InvitationResponse
	query := `
		SELECT 
			i.id, 
			i.email, 
			i.role_id, 
			r.name as role_name, 
			i.expires_at, 
			i.created_at
		FROM invitations i
		JOIN roles r ON i.role_id = r.id
		WHERE i.org_id = $1
		ORDER BY i.created_at DESC
	`
	err := r.db.SelectContext(ctx, &invites, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}
	return invites, nil
}

// ListOrgMembers retrieves a list of all users belonging to the specified organization.
// It executes a SQL JOIN to combine user details (email, names) with membership details (status, join date)
// and role details (role name).
func (r *repositoryImpl) ListOrgMembers(ctx context.Context, orgID int64) ([]OrgMemberResponse, error) {
	var members []OrgMemberResponse
	query := `
		SELECT 
			u.id as user_id, 
			u.email, 
			u.first_name, 
			u.last_name, 
			om.role_id, 
			r.name as role_name, 
			om.status, 
			om.created_at as joined_at
		FROM org_members om
		JOIN users u ON om.user_id = u.id
		JOIN roles r ON om.role_id = r.id
		WHERE om.org_id = $1
		ORDER BY om.created_at DESC
	`
	err := r.db.SelectContext(ctx, &members, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list organization members: %w", err)
	}
	return members, nil
}

// UpdateMemberRole updates the role_id of an existing organization membership record.
// It ensures that only the specified user's role within the specified organization is modified.
func (r *repositoryImpl) UpdateMemberRole(ctx context.Context, orgID, userID, roleID int64) error {
	query := `UPDATE org_members SET role_id = $1, updated_at = NOW() WHERE org_id = $2 AND user_id = $3`
	result, err := r.db.ExecContext(ctx, query, roleID, orgID, userID)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("member not found in organization")
	}
	return nil
}

// RemoveMember deletes the membership association between a user and an organization.
// The underlying user record remains intact in the 'users' table.
func (r *repositoryImpl) RemoveMember(ctx context.Context, orgID, userID int64) error {
	query := `DELETE FROM org_members WHERE org_id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, orgID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove member from organization: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("member not found in organization")
	}
	return nil
}

// CreateUser inserts a newly registered user into the 'users' table.
// It returns the auto-incremented ID generated by the database.
func (r *repositoryImpl) CreateUser(ctx context.Context, user *User) (int64, error) {
	query := `
		INSERT INTO users (cognito_sub, email, first_name, last_name)
		VALUES (:cognito_sub, :email, :first_name, :last_name)
		RETURNING id
	`
	rows, err := r.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&user.ID)
		if err != nil {
			return 0, fmt.Errorf("failed to scan returned user ID: %w", err)
		}
	}
	return user.ID, nil
}

// GetUserByEmail retrieves a user's record from the database using their unique email address.
func (r *repositoryImpl) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := `SELECT id, cognito_sub, email, first_name, last_name, created_at, updated_at FROM users WHERE email = $1`
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user by email: %w", err)
	}
	return &user, nil
}

// AddUserToOrg inserts a new record into the 'org_members' table, associating a user with an organization and role.
func (r *repositoryImpl) AddUserToOrg(ctx context.Context, member *OrgMember) error {
	query := `
		INSERT INTO org_members (org_id, user_id, role_id, status)
		VALUES (:org_id, :user_id, :role_id, :status)
		RETURNING id, created_at
	`
	rows, err := r.db.NamedQueryContext(ctx, query, member)
	if err != nil {
		return fmt.Errorf("failed to add user to organization: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&member.ID, &member.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to scan returned member ID: %w", err)
		}
	}
	return nil
}
