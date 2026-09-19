package integrations

import (
	"fmt"
	"net/http"
)

// Standard normalized error codes across all external integrations.
const (
	ErrCodeProviderNotConfigured   = "provider_not_configured"
	ErrCodeAuthenticationFailed     = "authentication_failed"
	ErrCodeAuthorizationFailed      = "authorization_failed"
	ErrCodeRateLimited             = "rate_limited"
	ErrCodeTimeout                 = "timeout"
	ErrCodeConnectionFailed        = "connection_failed"
	ErrCodeInvalidRequest          = "invalid_request"
	ErrCodeProviderUnavailable     = "provider_unavailable"
	ErrCodeDuplicateRequest        = "duplicate_request"
	ErrCodeWebhookSignatureInvalid = "webhook_signature_invalid"
	ErrCodeWebhookReplayDetected   = "webhook_replay_detected"
	ErrCodeObjectNotFound          = "object_not_found"
	ErrCodeUnsupportedDocument     = "unsupported_document"
	ErrCodeProviderError           = "provider_error"
)

// IntegrationError represents a sanitized, normalized error returned by external integrations.
// It explicitly never leaks provider secrets, raw authorization headers, or sensitive payloads.
type IntegrationError struct {
	CodeValue       string `json:"code"`
	MessageValue    string `json:"message"`
	HTTPStatusValue int    `json:"status"`
	ProviderValue   string `json:"provider,omitempty"`
	RetryableValue  bool   `json:"retryable"`
	CorrelationID   string `json:"correlation_id,omitempty"`
}

func (e *IntegrationError) Error() string {
	if e.ProviderValue != "" {
		return fmt.Sprintf("[%s:%s] %s", e.ProviderValue, e.CodeValue, e.MessageValue)
	}
	return fmt.Sprintf("[%s] %s", e.CodeValue, e.MessageValue)
}

func (e *IntegrationError) Code() string {
	return e.CodeValue
}

func (e *IntegrationError) Message() string {
	return e.MessageValue
}

func (e *IntegrationError) HTTPStatus() int {
	if e.HTTPStatusValue > 0 {
		return e.HTTPStatusValue
	}
	return http.StatusInternalServerError
}

func (e *IntegrationError) Provider() string {
	return e.ProviderValue
}

func (e *IntegrationError) Retryable() bool {
	return e.RetryableValue
}

// Pre-defined constructor helpers for normalized integration errors.

func NewProviderNotConfiguredError(provider string, integrationType string) *IntegrationError {
	return &IntegrationError{
		CodeValue:       ErrCodeProviderNotConfigured,
		MessageValue:    fmt.Sprintf("%s provider (%s) is not configured", integrationType, provider),
		HTTPStatusValue: http.StatusServiceUnavailable,
		ProviderValue:   provider,
		RetryableValue:  false,
	}
}

func NewAuthenticationFailedError(provider string, msg string) *IntegrationError {
	if msg == "" {
		msg = "external provider authentication failed"
	}
	return &IntegrationError{
		CodeValue:       ErrCodeAuthenticationFailed,
		MessageValue:    msg,
		HTTPStatusValue: http.StatusBadGateway,
		ProviderValue:   provider,
		RetryableValue:  false,
	}
}

func NewRateLimitedError(provider string, retryAfterSeconds int) *IntegrationError {
	msg := "external provider rate limit exceeded"
	if retryAfterSeconds > 0 {
		msg = fmt.Sprintf("external provider rate limit exceeded; retry after %d seconds", retryAfterSeconds)
	}
	return &IntegrationError{
		CodeValue:       ErrCodeRateLimited,
		MessageValue:    msg,
		HTTPStatusValue: http.StatusTooManyRequests,
		ProviderValue:   provider,
		RetryableValue:  true,
	}
}

func NewTimeoutError(provider string, duration string) *IntegrationError {
	return &IntegrationError{
		CodeValue:       ErrCodeTimeout,
		MessageValue:    fmt.Sprintf("request to %s timed out after %s", provider, duration),
		HTTPStatusValue: http.StatusGatewayTimeout,
		ProviderValue:   provider,
		RetryableValue:  true,
	}
}

func NewConnectionFailedError(provider string, details string) *IntegrationError {
	return &IntegrationError{
		CodeValue:       ErrCodeConnectionFailed,
		MessageValue:    fmt.Sprintf("connection to %s failed: %s", provider, details),
		HTTPStatusValue: http.StatusBadGateway,
		ProviderValue:   provider,
		RetryableValue:  true,
	}
}

func NewInvalidRequestError(provider string, details string) *IntegrationError {
	return &IntegrationError{
		CodeValue:       ErrCodeInvalidRequest,
		MessageValue:    details,
		HTTPStatusValue: http.StatusBadRequest,
		ProviderValue:   provider,
		RetryableValue:  false,
	}
}

func NewProviderUnavailableError(provider string, details string) *IntegrationError {
	return &IntegrationError{
		CodeValue:       ErrCodeProviderUnavailable,
		MessageValue:    fmt.Sprintf("%s is currently unavailable: %s", provider, details),
		HTTPStatusValue: http.StatusServiceUnavailable,
		ProviderValue:   provider,
		RetryableValue:  true,
	}
}

func NewDuplicateRequestError(correlationID string) *IntegrationError {
	return &IntegrationError{
		CodeValue:       ErrCodeDuplicateRequest,
		MessageValue:    "duplicate action execution detected",
		HTTPStatusValue: http.StatusConflict,
		RetryableValue:  false,
		CorrelationID:   correlationID,
	}
}

func NewWebhookSignatureInvalidError(provider string) *IntegrationError {
	return &IntegrationError{
		CodeValue:       ErrCodeWebhookSignatureInvalid,
		MessageValue:    "webhook signature verification failed",
		HTTPStatusValue: http.StatusUnauthorized,
		ProviderValue:   provider,
		RetryableValue:  false,
	}
}

func NewWebhookReplayDetectedError(provider string, eventID string) *IntegrationError {
	return &IntegrationError{
		CodeValue:       ErrCodeWebhookReplayDetected,
		MessageValue:    fmt.Sprintf("replay attack detected or duplicate webhook event %s", eventID),
		HTTPStatusValue: http.StatusConflict,
		ProviderValue:   provider,
		RetryableValue:  false,
	}
}

func NewProviderDisabledError(provider string) *IntegrationError {
	return &IntegrationError{
		CodeValue:       "provider_disabled",
		MessageValue:    fmt.Sprintf("%s provider is disabled by configuration", provider),
		HTTPStatusValue: http.StatusServiceUnavailable,
		ProviderValue:   provider,
		RetryableValue:  false,
	}
}

func NewObjectNotFoundError(provider string, key string) *IntegrationError {
	return &IntegrationError{
		CodeValue:       ErrCodeObjectNotFound,
		MessageValue:    fmt.Sprintf("requested object not found in %s: %s", provider, key),
		HTTPStatusValue: http.StatusNotFound,
		ProviderValue:   provider,
		RetryableValue:  false,
	}
}

func NewUnsupportedDocumentError(provider string, format string) *IntegrationError {
	return &IntegrationError{
		CodeValue:       ErrCodeUnsupportedDocument,
		MessageValue:    fmt.Sprintf("unsupported document format for %s: %s", provider, format),
		HTTPStatusValue: http.StatusBadRequest,
		ProviderValue:   provider,
		RetryableValue:  false,
	}
}

func NewAuthorizationFailedError(provider, details string) *IntegrationError {
	return &IntegrationError{
		CodeValue:       ErrCodeAuthorizationFailed,
		MessageValue:    details,
		HTTPStatusValue: http.StatusForbidden,
		ProviderValue:   provider,
		RetryableValue:  false,
	}
}

func NewProviderError(provider, details string, statusCode int) *IntegrationError {
	if statusCode <= 0 {
		statusCode = http.StatusBadGateway
	}
	return &IntegrationError{
		CodeValue:       ErrCodeProviderError,
		MessageValue:    details,
		HTTPStatusValue: statusCode,
		ProviderValue:   provider,
		RetryableValue:  false,
	}
}
