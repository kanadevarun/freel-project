package domain

import (
	"encoding/json"
	"fmt"
	"strings"
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

// MaskCredentials generates a safe, obfuscated view of credentials for UI display.
// e.g. "api_key": "••••••••3456", "api_secret": "••••••••"
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
	// Replace potential token patterns or sensitive keys
	msg = strings.ReplaceAll(msg, "Bearer ", "Bearer [REDACTED]")
	msg = strings.ReplaceAll(msg, "api_key=", "api_key=[REDACTED]")
	msg = strings.ReplaceAll(msg, "secret=", "secret=[REDACTED]")
	return msg
}
