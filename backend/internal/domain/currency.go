package domain

import (
	"time"
)

// Currency represents a currency code and symbol.
type Currency struct {
	ID        int64     `json:"id" db:"id"`
	Code      string    `json:"code" db:"code"`     // e.g., "USD", "INR"
	Symbol    string    `json:"symbol" db:"symbol"` // e.g., "$", "₹"
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
