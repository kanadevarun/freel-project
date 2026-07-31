package domain

import (
	"time"
)

// Address represents a physical address.
type Address struct {
	ID         int64   `json:"id" db:"id"`
	Line1      string     `json:"line1" db:"line1"`
	Line2      *string    `json:"line2,omitempty" db:"line2"`
	City       string     `json:"city" db:"city"`
	State      string     `json:"state" db:"state"`
	PostalCode string    `json:"postal_code" db:"postal_code"`
	CountryID  int64     `json:"country_id" db:"country_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`

	Country *Country `json:"country,omitempty" db:"-"`
}
