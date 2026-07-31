package organization

import "time"

// Organization represents a tenant in the system.
type Organization struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateOrgRequest represents the payload for creating an organization.
type CreateOrgRequest struct {
	Name string `json:"name" validate:"required"`
}

// UpdateOrgRequest represents the payload for updating an organization.
type UpdateOrgRequest struct {
	Name string `json:"name" validate:"required"`
}

// InviteRequest represents the payload for inviting a user.
type InviteRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required"`
}
