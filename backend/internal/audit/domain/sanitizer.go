package domain

import (
	"reflect"
	"strings"
)

// SensitiveKeyPatterns contains lowercased keywords that must be redacted automatically.
var SensitiveKeyPatterns = []string{
	"password",
	"passwd",
	"pass",
	"hash",
	"salt",
	"secret",
	"client_secret",
	"api_secret",
	"signing_secret",
	"webhook_secret",
	"token",
	"access_token",
	"refresh_token",
	"id_token",
	"oauth_token",
	"session_token",
	"session_secret",
	"session_id",
	"api_key",
	"apikey",
	"auth_key",
	"authorization",
	"auth_header",
	"bearer",
	"cookie",
	"set-cookie",
	"credential",
	"credentials",
	"card_number",
	"credit_card",
	"cvv",
	"cvc",
	"expiry",
	"expiration",
	"private_key",
	"private",
	"smtp_password",
	"pin",
	"ssn",
	"bank_account",
	"bank_account_number",
	"routing_number",
	"iban",
	"swift",
}

// IsSensitiveKey checks whether a key name contains sensitive information markers.
func IsSensitiveKey(key string) bool {
	lowerKey := strings.ToLower(strings.TrimSpace(key))
	for _, pattern := range SensitiveKeyPatterns {
		if strings.Contains(lowerKey, pattern) {
			return true
		}
	}
	return false
}

// SanitizeMap recursively cleanses map keys and values to prevent secret leakage.
func SanitizeMap(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return nil
	}

	sanitized := make(map[string]interface{}, len(input))
	for k, v := range input {
		if IsSensitiveKey(k) {
			sanitized[k] = "[REDACTED]"
			continue
		}
		sanitized[k] = SanitizeValue(v)
	}
	return sanitized
}

// SanitizeValue handles generic values, recursing into nested maps, slices, and structs.
func SanitizeValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case map[string]interface{}:
		return SanitizeMap(val)
	case []interface{}:
		sanitizedSlice := make([]interface{}, len(val))
		for i, item := range val {
			sanitizedSlice[i] = SanitizeValue(item)
		}
		return sanitizedSlice
	case []map[string]interface{}:
		sanitizedSlice := make([]map[string]interface{}, len(val))
		for i, item := range val {
			sanitizedSlice[i] = SanitizeMap(item)
		}
		return sanitizedSlice
	default:
		return v
	}
}

// ComputeChanges calculates field-level differences between before and after maps.
func ComputeChanges(before, after map[string]interface{}) []FieldChange {
	if before == nil && after == nil {
		return nil
	}

	cleanBefore := SanitizeMap(before)
	cleanAfter := SanitizeMap(after)

	allKeys := make(map[string]bool)
	for k := range before {
		allKeys[k] = true
	}
	for k := range after {
		allKeys[k] = true
	}

	var changes []FieldChange
	for k := range allKeys {
		rawBeforeVal, inBefore := before[k]
		rawAfterVal, inAfter := after[k]

		// If key was not present in before
		if !inBefore {
			changes = append(changes, FieldChange{
				Field:  k,
				Before: nil,
				After:  cleanAfter[k],
			})
			continue
		}

		// If key was deleted in after
		if !inAfter {
			changes = append(changes, FieldChange{
				Field:  k,
				Before: cleanBefore[k],
				After:  nil,
			})
			continue
		}

		// If values differ in raw data
		if !reflect.DeepEqual(rawBeforeVal, rawAfterVal) {
			changes = append(changes, FieldChange{
				Field:  k,
				Before: cleanBefore[k],
				After:  cleanAfter[k],
			})
		}
	}

	return changes
}
