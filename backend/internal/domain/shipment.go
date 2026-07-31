package domain

import (
	"time"
)

// Shipment represents a freight shipment.
type Shipment struct {
	ID             int64      `json:"id" db:"id"`
	OrgID          int64      `json:"org_id" db:"org_id"`
	CustomerID     int64      `json:"customer_id" db:"customer_id"`
	OriginID       int64      `json:"origin_id" db:"origin_id"`
	DestinationID  int64      `json:"destination_id" db:"destination_id"`
	Status         string     `json:"status" db:"status"` // e.g., "PENDING", "IN_TRANSIT", "DELIVERED"
	Incoterm       string     `json:"incoterm" db:"incoterm"`
	ModeOfTranport string     `json:"mode_of_transport" db:"mode_of_transport"` // e.g., "OCEAN", "AIR", "ROAD"
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`

	Customer    *Customer `json:"customer,omitempty" db:"-"`
	Origin      *Address  `json:"origin,omitempty" db:"-"`
	Destination *Address  `json:"destination,omitempty" db:"-"`
}
