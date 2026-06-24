package utils

import (
	"errors"
	"net/mail"
	"strings"
)

// ValidateEmail checks if the email is provided and roughly valid.
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email is required")
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("invalid email format")
	}
	return nil
}

// ValidatePassword checks if a password is provided and meets minimum length.
func ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}

// ValidateRequired checks if a string field is not empty.
func ValidateRequired(field, fieldName string) error {
	if strings.TrimSpace(field) == "" {
		return errors.New(fieldName + " is required")
	}
	return nil
}
