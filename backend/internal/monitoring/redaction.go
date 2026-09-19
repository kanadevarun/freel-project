package monitoring

import (
	"regexp"
	"strings"
)

var (
	bearerTokenRegex = regexp.MustCompile(`(?i)(bearer\s+)[a-zA-Z0-9_\-\.]{15,}`)
	apiKeyRegex      = regexp.MustCompile(`(?i)(api[_-]?key|secret[_-]?key|token)\s*[:=]\s*["']?([a-zA-Z0-9_\-\.]{12,})["']?`)
	creditCardRegex  = regexp.MustCompile(`\b(?:\d{4}[ -]?){3}\d{4}\b`)
	ssnRegex         = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	passwordRegex    = regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*["']?([^"'\s,;]+)["']?`)
)

// SanitizeDetails scrubs credentials, secrets, tokens, cards, and SSNs from string payloads.
func SanitizeDetails(input string) string {
	if input == "" {
		return ""
	}

	sanitized := bearerTokenRegex.ReplaceAllString(input, "${1}[REDACTED_TOKEN]")
	sanitized = apiKeyRegex.ReplaceAllString(sanitized, "${1}: [REDACTED_KEY]")
	sanitized = creditCardRegex.ReplaceAllString(sanitized, "[REDACTED_CARD_NUMBER]")
	sanitized = ssnRegex.ReplaceAllString(sanitized, "[REDACTED_SSN]")
	sanitized = passwordRegex.ReplaceAllString(sanitized, "${1}: [REDACTED_PASSWORD]")

	// Strip potential authorization headers
	if strings.Contains(strings.ToLower(sanitized), "authorization:") {
		authRegex := regexp.MustCompile(`(?i)authorization:\s*[^\r\n,]+`)
		sanitized = authRegex.ReplaceAllString(sanitized, "Authorization: [REDACTED]")
	}

	return sanitized
}

// RedactError ensures error strings do not leak credentials or connection strings with passwords.
func RedactError(err error) string {
	if err == nil {
		return ""
	}
	return SanitizeDetails(err.Error())
}
