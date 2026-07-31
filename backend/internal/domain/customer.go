package domain

import (
	"time"
)

// Customer represents a customer, typically associated with a Company.
type Customer struct {
	ID        int64     `json:"id" db:"id"`
	OrgID     int64     `json:"org_id" db:"org_id"`
	CompanyID int64     `json:"company_id" db:"company_id"`
	Status    string    `json:"status" db:"status"` // e.g., "ACTIVE", "CHURNED"
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	Company *Company `json:"company,omitempty" db:"-"`
}
