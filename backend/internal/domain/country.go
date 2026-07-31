package domain

import (
	"time"
)

// Country represents a country code and name.
type Country struct {
	ID        int64     `json:"id" db:"id"`
	Code      string    `json:"code" db:"code"` // e.g., "US", "IN"
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
