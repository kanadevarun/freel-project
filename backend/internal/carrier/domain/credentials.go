package domain

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	bearerTokenRegex = regexp.MustCompile(`(?i)(bearer\s+)[A-Za-z0-9_\-\.]+`)
	sensitiveParamRegex = regexp.MustCompile(`(?i)(client_secret|api_secret|api_key|webhook_secret|secret|password)=([^&\s;]+)`)
	sensitiveHeaderRegex = regexp.MustCompile(`(?i)(Consumer-Key|X-API-Key|Ocp-Apim-Subscription-Key|X-IBM-Client-Secret|Authorization)([:=]\s*)[^\s,;]+`)
)

// DecryptedCredentials holds plaintext credentials in memory during adapter execution.
// It is NEVER returned in API responses, serialized to logs, or persisted in cleartext.
type DecryptedCredentials struct {
	APIKey         string                 `json:"api_key,omitempty"`
	APISecret      string                 `json:"api_secret,omitempty"`
	BaseURL        string                 `json:"base_url,omitempty"`
	AuthType       string                 `json:"auth_type,omitempty"`
	ClientID       string                 `json:"client_id,omitempty"`
	ClientSecret   string                 `json:"client_secret,omitempty"`
	WebhookSecret  string                 `json:"webhook_secret,omitempty"`
	AccountID      string                 `json:"account_id,omitempty"`
	CustomHeaders  map[string]string      `json:"custom_headers,omitempty"`
	ExtraOptions   map[string]interface{} `json:"extra_options,omitempty"`
}

// String implements fmt.Stringer to ensure raw secrets are never printed in logs if formatted with %v or %s.
func (d DecryptedCredentials) String() string {
	return "[PROTECTED_CARRIER_CREDENTIALS]"
}

// GoString implements fmt.GoStringer to protect against %#v log printing.
func (d DecryptedCredentials) GoString() string {
	return "[PROTECTED_CARRIER_CREDENTIALS]"
}

// GetConsumerKey returns the canonical API/Consumer Key, prioritizing APIKey over ClientID.
func (d DecryptedCredentials) GetConsumerKey() string {
	if d.APIKey != "" {
		return d.APIKey
	}
	return d.ClientID
}

// GetAPISecret returns the canonical API/Client Secret, prioritizing APISecret over ClientSecret.
func (d DecryptedCredentials) GetAPISecret() string {
	if d.APISecret != "" {
		return d.APISecret
	}
	return d.ClientSecret
}

// MaskCredentials generates a safe, obfuscated view of credentials for UI display.
// Strictly masks all secrets (client_secret, api_secret, webhook_secret, password, token)
// while allowing safe account identifiers (client_id, account_id) to show trailing characters for user verification.
func MaskCredentials(rawJSON string) map[string]string {
	masked := make(map[string]string)
	if rawJSON == "" {
		return masked
	}

	var rawMap map[string]interface{}
	if err := json.Unmarshal([]byte(rawJSON), &rawMap); err != nil {
		return masked
	}

	for k, v := range rawMap {
		valStr := fmt.Sprintf("%v", v)
		valLen := len(valStr)
		lowerKey := strings.ToLower(k)

		// Sensitive keys: ALWAYS mask fully with bullets, never reveal trailing characters
		if strings.Contains(lowerKey, "secret") ||
			strings.Contains(lowerKey, "password") ||
			strings.Contains(lowerKey, "token") ||
			strings.Contains(lowerKey, "key_secret") {
			masked[k] = "••••••••"
			continue
		}

		// Account IDs, Client IDs, or public keys: reveal at most last 4 characters if long enough
		if valLen <= 4 {
			masked[k] = "••••••••"
		} else {
			lastFour := valStr[valLen-4:]
			masked[k] = fmt.Sprintf("••••••••%s", lastFour)
		}
	}
	return masked
}

// SanitizeError strips sensitive tokens, authorization headers, or keys from raw provider error messages.
func SanitizeError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()

	// Redact Bearer tokens: replace the token after Bearer
	msg = bearerTokenRegex.ReplaceAllString(msg, "${1}[REDACTED]")

	// Redact key=value sensitive parameters
	msg = sensitiveParamRegex.ReplaceAllString(msg, "${1}=[REDACTED]")

	// Redact sensitive header values
	msg = sensitiveHeaderRegex.ReplaceAllString(msg, "${1}${2}[REDACTED]")

	return msg
}
