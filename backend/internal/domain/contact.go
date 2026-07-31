package domain

import (
	"time"
)

// Contact represents an individual person at a Company.
type Contact struct {
	ID        int64     `json:"id" db:"id"`
	OrgID     int64     `json:"org_id" db:"org_id"`
	CompanyID int64     `json:"company_id" db:"company_id"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Email     string    `json:"email" db:"email"`
	Phone     *string   `json:"phone,omitempty" db:"phone"`
	JobTitle  *string   `json:"job_title,omitempty" db:"job_title"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	Company *Company `json:"company,omitempty" db:"-"`
}
