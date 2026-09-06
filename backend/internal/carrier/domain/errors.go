package domain

import (
	"fmt"
)

// Standard Error Codes for Carrier Integrations
const (
	ErrCodeAuthFailed     = "AUTHENTICATION_FAILED"
	ErrCodeForbidden      = "AUTHORIZATION_FAILED"
	ErrCodeRateLimited    = "RATE_LIMITED"
	ErrCodeTimeout        = "TIMEOUT"
	ErrCodeUnavailable    = "CARRIER_UNAVAILABLE"
	ErrCodeUnsupported    = "UNSUPPORTED_OPERATION"
	ErrCodeNotFound       = "RESOURCE_NOT_FOUND"
	ErrCodeInvalidRequest = "INVALID_REQUEST"
	ErrCodeInvalidConfig  = "INVALID_CONFIGURATION"
	ErrCodeInternal       = "CARRIER_INTERNAL_ERROR"
)

// IntegrationError represents a structured, normalized error returned by a carrier adapter.
type IntegrationError struct {
	Provider      string `json:"provider"`
	Operation     string `json:"operation"`
	ErrorCode     string `json:"error_code"`
	UserMessage   string `json:"user_message"`
	Retryable     bool   `json:"retryable"`
	HTTPStatus    int    `json:"http_status"`
	CorrelationID string `json:"correlation_id,omitempty"`
	InternalErr   error  `json:"-"`
}

func (e *IntegrationError) Error() string {
	if e.UserMessage != "" {
		return fmt.Sprintf("[%s:%s] %s (%s)", e.Provider, e.Operation, e.UserMessage, e.ErrorCode)
	}
	if e.InternalErr != nil {
		return fmt.Sprintf("[%s:%s] %v (%s)", e.Provider, e.Operation, e.InternalErr, e.ErrorCode)
	}
	return fmt.Sprintf("[%s:%s] carrier integration error (%s)", e.Provider, e.Operation, e.ErrorCode)
}

func (e *IntegrationError) Unwrap() error {
	return e.InternalErr
}

// NewIntegrationError constructs a standardized IntegrationError.
func NewIntegrationError(provider, operation, code, userMsg string, httpStatus int, retryable bool, internalErr error) *IntegrationError {
	return &IntegrationError{
		Provider:    provider,
		Operation:   operation,
		ErrorCode:   code,
		UserMessage: userMsg,
		HTTPStatus:  httpStatus,
		Retryable:   retryable,
		InternalErr: internalErr,
	}
}
