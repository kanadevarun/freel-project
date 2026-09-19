package ai

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
)

// Standardized Error Categories Taxonomy (Part 8)
const (
	ErrorCategoryConfiguration     = "configuration_error"
	ErrorCategoryAuthentication    = "authentication_error"
	ErrorCategoryAuthorization     = "authorization_error"
	ErrorCategoryTenantIsolation   = "organization_isolation_error"
	ErrorCategoryValidation        = "validation_error"
	ErrorCategoryRateLimit         = "provider_rate_limit"
	ErrorCategoryTimeout           = "provider_timeout"
	ErrorCategoryUnavailable       = "provider_unavailable"
	ErrorCategoryModelError        = "model_error"
	ErrorCategoryToolError         = "tool_error"
	ErrorCategoryActionRejected    = "action_rejected"
	ErrorCategoryApprovalRequired  = "approval_required"
	ErrorCategoryCheckpointError   = "checkpoint_error"
	ErrorCategoryTaskRetryable     = "task_retryable"
	ErrorCategoryPermanentFailure  = "task_permanent_failure"
	ErrorCategoryUnknown           = "unknown_error"
)

var (
	ErrProvidersUnavailable = errors.New("all configured ai providers failed")
	ErrInvalidExecutionContext = errors.New("invalid execution context: organization_id is required")
)

// Secret redaction patterns
var (
	apiKeyRegex    = regexp.MustCompile(`(?i)(api[_-]?key|secret|token|password|bearer)\s*[:=]\s*["']?([a-zA-Z0-9_\-\.]{8,})["']?`)
	skOpenAIRegex  = regexp.MustCompile(`sk-[a-zA-Z0-9_\-]{20,}`)
	googleKeyRegex = regexp.MustCompile(`AIza[0-9A-Za-z-_]{10,}`)
	tokenRegex     = regexp.MustCompile(`lhq_sec_[a-zA-Z0-9_\-]{15,}`)
)

// RedactSecrets replaces sensitive tokens, passwords, and API keys with [REDACTED].
func RedactSecrets(text string) string {
	if text == "" {
		return text
	}
	s := skOpenAIRegex.ReplaceAllString(text, "sk-[REDACTED]")
	s = googleKeyRegex.ReplaceAllString(s, "AIza[REDACTED]")
	s = tokenRegex.ReplaceAllString(s, "lhq_sec_[REDACTED]")
	s = apiKeyRegex.ReplaceAllString(s, "$1:[REDACTED]")
	return s
}

// ClassifyError categorizes errors into the standard taxonomy for observability.
func ClassifyError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.ToLower(err.Error())

	// 1. Rate limits & Quotas
	if strings.Contains(msg, "429") || strings.Contains(msg, "rate limit") || strings.Contains(msg, "quota") || strings.Contains(msg, "insufficient_quota") {
		return ErrorCategoryRateLimit
	}

	// 2. Timeouts
	if strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded") || strings.Contains(msg, "context deadline") {
		return ErrorCategoryTimeout
	}

	// 3. Provider Availability (502, 503, 504, connection reset)
	if strings.Contains(msg, "502") || strings.Contains(msg, "503") || strings.Contains(msg, "504") ||
		strings.Contains(msg, "unavailable") || strings.Contains(msg, "connection refused") || strings.Contains(msg, "bad gateway") {
		return ErrorCategoryUnavailable
	}

	// 4. Authentication / API key
	if strings.Contains(msg, "401") || strings.Contains(msg, "unauthorized") || strings.Contains(msg, "invalid api key") || strings.Contains(msg, "key is missing") {
		return ErrorCategoryAuthentication
	}

	// 5. Authorization / Permissions
	if strings.Contains(msg, "403") || strings.Contains(msg, "forbidden") || strings.Contains(msg, "permission denied") {
		return ErrorCategoryAuthorization
	}

	// 6. Organization isolation
	if strings.Contains(msg, "tenant") || strings.Contains(msg, "organization mismatch") || strings.Contains(msg, "cross-organization") {
		return ErrorCategoryTenantIsolation
	}

	// 7. Validation / Schema
	if strings.Contains(msg, "400") || strings.Contains(msg, "validation") || strings.Contains(msg, "json parse") || strings.Contains(msg, "invalid json") {
		return ErrorCategoryValidation
	}

	// 8. Configuration
	if strings.Contains(msg, "config") || strings.Contains(msg, "unsupported model") {
		return ErrorCategoryConfiguration
	}

	// 9. Approvals
	if strings.Contains(msg, "approval required") || strings.Contains(msg, "confirmation required") {
		return ErrorCategoryApprovalRequired
	}

	return ErrorCategoryUnknown
}

// IsRetryableErrorCategory determines if the category is eligible for automated failover or retry.
func IsRetryableErrorCategory(cat string) bool {
	switch cat {
	case ErrorCategoryRateLimit, ErrorCategoryTimeout, ErrorCategoryUnavailable, ErrorCategoryTaskRetryable:
		return true
	default:
		return false
	}
}

// HTTPStatusCodeForCategory maps category to safe client HTTP status code.
func HTTPStatusCodeForCategory(cat string) int {
	switch cat {
	case ErrorCategoryRateLimit:
		return http.StatusTooManyRequests
	case ErrorCategoryTimeout:
		return http.StatusGatewayTimeout
	case ErrorCategoryUnavailable:
		return http.StatusBadGateway
	case ErrorCategoryAuthentication:
		return http.StatusUnauthorized
	case ErrorCategoryAuthorization, ErrorCategoryTenantIsolation:
		return http.StatusForbidden
	case ErrorCategoryValidation:
		return http.StatusBadRequest
	case ErrorCategoryApprovalRequired:
		return http.StatusAccepted
	default:
		return http.StatusInternalServerError
	}
}
