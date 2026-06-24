package models

import "time"

type User struct {
	ID             string    `json:"id"`
	FullName       string    `json:"full_name"`
	Email          string    `json:"email"`
	Role           string    `json:"role"`
	OrganizationID string    `json:"organization_id"`
	CognitoSub     string    `json:"cognito_sub"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
