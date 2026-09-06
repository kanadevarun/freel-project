package quotations

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateSecurePublicToken generates a cryptographically secure 64-character hex token (32 random bytes).
func GenerateSecurePublicToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// IsPublicLinkExpired checks if a public link has passed its expiration time.
func IsPublicLinkExpired(link *QuotationPublicLink) bool {
	if link == nil {
		return true
	}
	if link.Status == QuotationPublicLinkRevoked {
		return false // It's revoked, not expired
	}
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return true
	}
	return false
}

// IsQuotationSharable checks whether a quotation has progressed far enough in its lifecycle to be shared.
// Only APPROVED, SENT, VIEWED, ACCEPTED, and DECLINED quotations may be shared publicly.
func IsQuotationSharable(status string) bool {
	switch status {
	case QuotationStatusApproved,
		QuotationStatusSent,
		QuotationStatusViewed,
		QuotationStatusAccepted,
		QuotationStatusDeclined:
		return true
	default:
		return false
	}
}
