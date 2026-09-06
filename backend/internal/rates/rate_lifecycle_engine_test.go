package rates

import (
	"testing"
	"time"
)

func TestEvaluateRateLifecycleStatus(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)

	// 1. Rate valid for 60 days -> ACTIVE
	future60 := now.AddDate(0, 0, 60)
	status, days := EvaluateRateLifecycleStatus("ACTIVE", nil, &future60, now)
	if status != RateStatusActive || days <= 30 {
		t.Errorf("expected ACTIVE with >30 days, got %s with %d days", status, days)
	}

	// 2. Rate valid for 10 days -> EXPIRING_SOON
	future10 := now.AddDate(0, 0, 10)
	status, days = EvaluateRateLifecycleStatus("ACTIVE", nil, &future10, now)
	if status != RateStatusExpiringSoon || days != 10 {
		t.Errorf("expected EXPIRING_SOON with 10 days, got %s with %d days", status, days)
	}

	// 3. Rate expired 2 days ago -> EXPIRED
	past2 := now.AddDate(0, 0, -2)
	status, days = EvaluateRateLifecycleStatus("ACTIVE", nil, &past2, now)
	if status != RateStatusExpired || days >= 0 {
		t.Errorf("expected EXPIRED with negative days, got %s with %d days", status, days)
	}

	// 4. Rate explicitly archived -> ARCHIVED
	status, _ = EvaluateRateLifecycleStatus(RateStatusArchived, nil, &future60, now)
	if status != RateStatusArchived {
		t.Errorf("expected ARCHIVED to remain ARCHIVED, got %s", status)
	}

	// 5. Rate superseded -> SUPERSEDED
	status, _ = EvaluateRateLifecycleStatus("SUPERSEDED", nil, &future60, now)
	if status != "SUPERSEDED" {
		t.Errorf("expected SUPERSEDED to remain SUPERSEDED, got %s", status)
	}
}

func TestEvaluateContractLifecycleStatus(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)

	// 1. Contract ending in 90 days -> ACTIVE, NOT_STARTED
	future90 := now.AddDate(0, 0, 90)
	status, renewal, days := EvaluateContractLifecycleStatus(ContractStatusActive, RenewalStatusNotStarted, &future90, now)
	if status != ContractStatusActive || renewal != RenewalStatusNotStarted || days <= 30 {
		t.Errorf("expected ACTIVE with NOT_STARTED, got %s, %s, %d", status, renewal, days)
	}

	// 2. Contract ending in 15 days -> EXPIRING_SOON, IN_PROGRESS
	future15 := now.AddDate(0, 0, 15)
	status, renewal, days = EvaluateContractLifecycleStatus(ContractStatusActive, RenewalStatusNotStarted, &future15, now)
	if status != ContractStatusExpiringSoon || renewal != RenewalStatusInProgress || days != 15 {
		t.Errorf("expected EXPIRING_SOON with IN_PROGRESS, got %s, %s, %d", status, renewal, days)
	}

	// 3. Contract expired 5 days ago -> EXPIRED
	past5 := now.AddDate(0, 0, -5)
	status, _, days = EvaluateContractLifecycleStatus(ContractStatusActive, RenewalStatusInProgress, &past5, now)
	if status != ContractStatusExpired || days >= 0 {
		t.Errorf("expected EXPIRED with negative days, got %s, %d", status, days)
	}
}

func TestDetectQuotationRateRisks(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	rateID := int64(100)
	contractID := int64(200)

	// Case 1: Draft quotation with expired rate -> Critical Risk
	pastDate := now.AddDate(0, 0, -3)
	input1 := RateRiskEvaluationInput{
		QuotationID:             1,
		QuotationNumber:         "QT-2026-0001",
		QuotationStatus:         "DRAFT",
		SnapshotCarrierName:     "Maersk Line",
		SnapshotCommercialTotal: 2500,
		SourceRateID:            &rateID,
		SourceRateStatus:        "EXPIRED",
		SourceRateValidUntil:    &pastDate,
		SourceRateVersion:       1,
		LatestRateVersion:       1,
	}

	risks1 := DetectQuotationRateRisks(input1, now)
	if len(risks1) == 0 {
		t.Fatalf("expected at least 1 risk for expired rate")
	}
	if risks1[0].RiskType != "RATE_EXPIRED" || risks1[0].Severity != SeverityCritical {
		t.Errorf("expected RATE_EXPIRED with CRITICAL severity for draft quote, got %s / %s", risks1[0].RiskType, risks1[0].Severity)
	}

	// Case 2: Accepted quotation with expired rate -> Warning Risk (Preserved pricing)
	input2 := input1
	input2.QuotationStatus = "ACCEPTED"
	risks2 := DetectQuotationRateRisks(input2, now)
	if len(risks2) == 0 || risks2[0].Severity != SeverityWarning {
		t.Errorf("expected WARNING severity for accepted quotation with expired source rate, got %+v", risks2)
	}

	// Case 3: Rate superseded by newer version
	futureDate := now.AddDate(0, 0, 60)
	input3 := RateRiskEvaluationInput{
		QuotationID:             2,
		QuotationNumber:         "QT-2026-0002",
		QuotationStatus:         "DRAFT",
		SnapshotCarrierName:     "MSC",
		SnapshotCommercialTotal: 3000,
		SourceRateID:            &rateID,
		SourceRateStatus:        "ACTIVE",
		SourceRateValidUntil:    &futureDate,
		SourceRateVersion:       1,
		LatestRateVersion:       2,
		SourceContractID:        &contractID,
		SourceContractCode:      "MSC-CON-2026",
		SourceContractStatus:    "ACTIVE",
		SourceContractEndDate:   &futureDate,
	}

	risks3 := DetectQuotationRateRisks(input3, now)
	foundSuperseded := false
	for _, r := range risks3 {
		if r.RiskType == "RATE_SUPERSEDED" {
			foundSuperseded = true
			break
		}
	}
	if !foundSuperseded {
		t.Errorf("expected RATE_SUPERSEDED risk when latest version > source version")
	}
}
