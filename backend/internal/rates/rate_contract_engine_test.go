package rates

import (
	"testing"
	"time"
)

func TestCalculateRateContractStatus(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name          string
		effectiveDate time.Time
		expiryDate    time.Time
		currentStatus string
		expected      string
	}{
		{
			name:          "Active contract valid for 180 days",
			effectiveDate: now.AddDate(0, -1, 0),
			expiryDate:    now.AddDate(0, 5, 0),
			currentStatus: ContractStatusActive,
			expected:      ContractStatusActive,
		},
		{
			name:          "Expiring soon within 15 days",
			effectiveDate: now.AddDate(0, -3, 0),
			expiryDate:    now.AddDate(0, 0, 15),
			currentStatus: ContractStatusActive,
			expected:      ContractStatusExpiringSoon,
		},
		{
			name:          "Expired contract 5 days ago",
			effectiveDate: now.AddDate(0, -6, 0),
			expiryDate:    now.AddDate(0, 0, -5),
			currentStatus: ContractStatusActive,
			expected:      ContractStatusExpired,
		},
		{
			name:          "Draft future contract starting next month",
			effectiveDate: now.AddDate(0, 1, 0),
			expiryDate:    now.AddDate(0, 7, 0),
			currentStatus: ContractStatusActive,
			expected:      ContractStatusDraft,
		},
		{
			name:          "Archived contract remains archived regardless of dates",
			effectiveDate: now.AddDate(0, -1, 0),
			expiryDate:    now.AddDate(0, 5, 0),
			currentStatus: ContractStatusArchived,
			expected:      ContractStatusArchived,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateRateContractStatus(tt.effectiveDate, tt.expiryDate, tt.currentStatus)
			if got != tt.expected {
				t.Errorf("CalculateRateContractStatus() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCanModifyRateDirectly(t *testing.T) {
	tests := []struct {
		name          string
		isReferenced  bool
		status        string
		versionStatus string
		expected      bool
	}{
		{
			name:          "Unreferenced active rate allows direct modification",
			isReferenced:  false,
			status:        RateStatusActive,
			versionStatus: VersionStatusCurrent,
			expected:      true,
		},
		{
			name:          "Referenced active rate prohibits direct modification (requires new version)",
			isReferenced:  true,
			status:        RateStatusActive,
			versionStatus: VersionStatusCurrent,
			expected:      false,
		},
		{
			name:          "Superseded rate prohibits direct modification",
			isReferenced:  false,
			status:        RateStatusActive,
			versionStatus: VersionStatusSuperseded,
			expected:      false,
		},
		{
			name:          "Archived rate prohibits direct modification",
			isReferenced:  false,
			status:        RateStatusArchived,
			versionStatus: VersionStatusCurrent,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanModifyRateDirectly(tt.isReferenced, tt.status, tt.versionStatus)
			if got != tt.expected {
				t.Errorf("CanModifyRateDirectly() = %v, want %v", got, tt.expected)
			}
		})
	}
}
