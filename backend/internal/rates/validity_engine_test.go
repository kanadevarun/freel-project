package rates

import (
	"testing"
	"time"
)

func TestCalculateRateValidity(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		effectiveDate   time.Time
		expiryDate      time.Time
		currentStatus   RateStatus
		expectedStatus  RateStatus
		expectedMinDays int
		expectedMaxDays int
	}{
		{
			name:            "Active rate valid for 90 days",
			effectiveDate:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
			expiryDate:      time.Date(2026, 11, 29, 0, 0, 0, 0, time.UTC),
			currentStatus:   RateStatusActive,
			expectedStatus:  RateStatusActive,
			expectedMinDays: 90,
			expectedMaxDays: 93,
		},
		{
			name:            "Expiring soon rate within 10 days",
			effectiveDate:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
			expiryDate:      time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC),
			currentStatus:   RateStatusActive,
			expectedStatus:  RateStatusExpiringSoon,
			expectedMinDays: 7,
			expectedMaxDays: 7,
		},
		{
			name:            "Expired rate yesterday",
			effectiveDate:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			expiryDate:      time.Date(2026, 8, 28, 0, 0, 0, 0, time.UTC),
			currentStatus:   RateStatusActive,
			expectedStatus:  RateStatusExpired,
			expectedMinDays: -1,
			expectedMaxDays: -1,
		},
		{
			name:            "Draft upcoming rate next month",
			effectiveDate:   time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC),
			expiryDate:      time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			currentStatus:   RateStatusDraft,
			expectedStatus:  RateStatusDraft,
			expectedMinDays: 100,
			expectedMaxDays: 130,
		},
		{
			name:            "Archived rate remains archived",
			effectiveDate:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
			expiryDate:      time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			currentStatus:   RateStatusArchived,
			expectedStatus:  RateStatusArchived,
			expectedMinDays: 100,
			expectedMaxDays: 130,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, days := CalculateRateValidity(tt.effectiveDate, tt.expiryDate, tt.currentStatus, now)
			if status != tt.expectedStatus {
				t.Errorf("got status %s, want %s", status, tt.expectedStatus)
			}
			if days < tt.expectedMinDays || days > tt.expectedMaxDays {
				t.Errorf("got days %d, want between %d and %d", days, tt.expectedMinDays, tt.expectedMaxDays)
			}
		})
	}
}
