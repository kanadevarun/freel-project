package adapters

import (
	"errors"
	"fmt"
)

var (
	ErrAdapterNotFound          = errors.New("carrier adapter not found in registry")
	ErrCapabilityNotImplemented = errors.New("capability not implemented for this carrier adapter")
	ErrInvalidCredentials       = errors.New("invalid or missing carrier API credentials")
	ErrCarrierAPIUnavailable    = errors.New("carrier API is temporarily unreachable")
	ErrAuthenticationFailed     = errors.New("carrier authentication failed: check API key / secret")
	ErrRateLimitExceeded        = errors.New("carrier API rate limit exceeded")
)

// AdapterError represents a structured, safely-formattable error from a carrier adapter.
type AdapterError struct {
	CarrierCode string
	Capability  string
	HTTPStatus  int
	InternalErr error
	SafeMessage string
}

func (e *AdapterError) Error() string {
	if e.SafeMessage != "" {
		return fmt.Sprintf("[%s:%s] %s", e.CarrierCode, e.Capability, e.SafeMessage)
	}
	if e.InternalErr != nil {
		return fmt.Sprintf("[%s:%s] %v", e.CarrierCode, e.Capability, e.InternalErr)
	}
	return fmt.Sprintf("[%s:%s] carrier adapter error", e.CarrierCode, e.Capability)
}

func (e *AdapterError) Unwrap() error {
	return e.InternalErr
}

func NewCapabilityNotImplementedError(carrierCode, capability string) error {
	return &AdapterError{
		CarrierCode: carrierCode,
		Capability:  capability,
		HTTPStatus:  501,
		SafeMessage: fmt.Sprintf("Carrier %s does not currently support %s operation via API adapter", carrierCode, capability),
		InternalErr: ErrCapabilityNotImplemented,
	}
}

func NewAdapterNotConfiguredError(carrierCode string) error {
	return &AdapterError{
		CarrierCode: carrierCode,
		HTTPStatus:  400,
		SafeMessage: fmt.Sprintf("Carrier %s adapter is not configured with valid endpoints or credentials", carrierCode),
		InternalErr: ErrInvalidCredentials,
	}
}
