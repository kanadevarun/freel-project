package models

import "time"

type Onboarding struct {
	ID             string                 `json:"id"`
	OrganizationID string                 `json:"organization_id"`
	Role           string                 `json:"role"`
	Answers        map[string]interface{} `json:"answers"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}
