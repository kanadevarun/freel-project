package users

import (
	"time"
)

// User represents a registered user in the system.
// This structure maps directly to the "users" database table.
// It contains core identity information such as the unique Cognito subject ID,
// the user's email address, and their first and last name.
type User struct {
	// ID is the unique auto-incrementing integer identifier for the user.
	ID int64 `json:"id" db:"id"`
	// CognitoSub is the unique identifier assigned to the user by AWS Cognito.
	CognitoSub string `json:"cognito_sub" db:"cognito_sub"`
	// Email is the user's unique email address used for login and notifications.
	Email string `json:"email" db:"email"`
	// FirstName is the user's given name.
	FirstName *string `json:"first_name" db:"first_name"`
	// LastName is the user's family name.
	LastName *string `json:"last_name" db:"last_name"`
	// CreatedAt is the timestamp when the user record was created.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	// UpdatedAt is the timestamp of the last update to the user record.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// OrgMember represents a user's membership within a specific organization.
// This structure maps directly to the "org_members" database table and joins with the roles table
// to provide the human-readable role name.
type OrgMember struct {
	// ID is the unique identifier for the membership record.
	ID int64 `json:"id" db:"id"`
	// OrgID is the identifier of the organization the user belongs to.
	OrgID int64 `json:"org_id" db:"org_id"`
	// UserID is the identifier of the user who holds this membership.
	UserID int64 `json:"user_id" db:"user_id"`
	// RoleID is the identifier of the role assigned to the user in this organization.
	RoleID int64 `json:"role_id" db:"role_id"`
	// RoleName is the human-readable name of the role (e.g., 'ADMIN', 'SALES').
	// This is typically joined from the "roles" table during retrieval.
	RoleName string `json:"role_name" db:"role_name"`
	// Status indicates the current state of the membership (e.g., 'ACTIVE', 'INACTIVE').
	Status string `json:"status" db:"status"`
	// CreatedAt is the timestamp when the membership was created.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// OrgMemberResponse is the structure returned to the client when listing organization members.
// It flattens the User and OrgMember data into a single clean API response.
type OrgMemberResponse struct {
	// UserID is the unique identifier of the user.
	UserID int64 `json:"user_id" db:"user_id"`
	// Email is the user's email address.
	Email string `json:"email" db:"email"`
	// FirstName is the user's given name.
	FirstName *string `json:"first_name" db:"first_name"`
	// LastName is the user's family name.
	LastName *string `json:"last_name" db:"last_name"`
	// RoleID is the identifier of the user's assigned role.
	RoleID int64 `json:"role_id" db:"role_id"`
	// RoleName is the human-readable name of the user's assigned role.
	RoleName string `json:"role_name" db:"role_name"`
	// Status is the current status of the user's membership in the organization.
	Status string `json:"status" db:"status"`
	// JoinedAt is the timestamp when the user became a member of the organization.
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
}

// Invitation represents a pending invitation for a new user to join an organization.
// This structure maps directly to the "invitations" database table.
type Invitation struct {
	// ID is the unique auto-incrementing integer identifier for the invitation.
	ID int64 `json:"id" db:"id"`
	// OrgID is the identifier of the organization the invitee will join.
	OrgID int64 `json:"org_id" db:"org_id"`
	// RoleID is the identifier of the role that will be assigned to the invitee upon acceptance.
	RoleID int64 `json:"role_id" db:"role_id"`
	// Email is the email address where the invitation was sent.
	Email string `json:"email" db:"email"`
	// Token is a secure, randomly generated string used to validate the invitation link.
	Token string `json:"token" db:"token"`
	// ExpiresAt is the timestamp when the invitation will expire and no longer be valid.
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	// CreatedAt is the timestamp when the invitation was generated.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	// UpdatedAt is the timestamp of the last update to the invitation record.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// InvitationResponse is the structure returned to the client when listing pending invitations.
// It exposes necessary details (like email and role) but omits the secure token.
type InvitationResponse struct {
	ID        int64     `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	RoleID    int64     `json:"role_id" db:"role_id"`
	RoleName  string    `json:"role_name" db:"role_name"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// InviteUserRequest represents the payload expected from the client when creating a new invitation.
type InviteUserRequest struct {
	// Email is the target email address to invite.
	Email string `json:"email" binding:"required,email"`
	// RoleID is the role that should be assigned to the user upon joining.
	RoleID int64 `json:"role_id" binding:"required"`
	// OrgName is the name of the organization sending the invite, used in the email template.
	OrgName string `json:"org_name"`
}

// UpdateRoleRequest represents the payload expected from the client to change an existing member's role.
type UpdateRoleRequest struct {
	// RoleID is the new role to assign to the user.
	RoleID int64 `json:"role_id" binding:"required"`
}

// AcceptInviteRequest represents the payload expected from the client when accepting an invitation during signup.
type AcceptInviteRequest struct {
	// Token is the secure invitation token extracted from the invite link.
	Token string `json:"token" binding:"required"`
	// Password is the new password the user wishes to set for their account.
	Password string `json:"password" binding:"required,min=8"`
	// FullName is the user's chosen display name.
	FullName string `json:"full_name" binding:"required"`
}
