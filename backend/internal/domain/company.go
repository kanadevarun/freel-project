package domain

import (
	"time"
)

// Company represents a business entity (can be a lead, customer, or partner).
type Company struct {
	ID        int64      `json:"id" db:"id"`
	OrgID     int64      `json:"org_id" db:"org_id"` // Tenant ID
	Name      string     `json:"name" db:"name"`
	Domain    *string    `json:"domain,omitempty" db:"domain"`
	Industry  *string    `json:"industry,omitempty" db:"industry"`
	AddressID *int64     `json:"address_id,omitempty" db:"address_id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`

	Address *Address `json:"address,omitempty" db:"-"`
}
