package quotations

import (
	"testing"
)

func TestValidateQuotationForRateSelection(t *testing.T) {
	qDraft := &Quotation{Status: QuotationStatusDraft}
	if err := ValidateQuotationForRateSelection(qDraft); err != nil {
		t.Errorf("expected draft quotation to be editable, got: %v", err)
	}

	qChanges := &Quotation{Status: QuotationStatusChangesRequested}
	if err := ValidateQuotationForRateSelection(qChanges); err != nil {
		t.Errorf("expected changes requested quotation to be editable, got: %v", err)
	}

	qAccepted := &Quotation{Status: QuotationStatusAccepted}
	if err := ValidateQuotationForRateSelection(qAccepted); err == nil {
		t.Errorf("expected accepted quotation rate selection to fail")
	}

	qSent := &Quotation{Status: QuotationStatusSent}
	if err := ValidateQuotationForRateSelection(qSent); err == nil {
		t.Errorf("expected sent quotation rate selection to fail")
	}
}

func TestEvaluateCandidateRecommendations(t *testing.T) {
	candidates := []QuotationRateCandidate{
		{
			CarrierName:     "MSC",
			Currency:        "USD",
			CommercialTotal: 2200.0,
			TransitDays:     22,
		},
		{
			CarrierName:     "CMA CGM",
			Currency:        "USD",
			CommercialTotal: 1950.0,
			TransitDays:     28,
		},
		{
			CarrierName:     "Maersk",
			Currency:        "USD",
			CommercialTotal: 2450.0,
			TransitDays:     18,
		},
	}

	res := EvaluateCandidateRecommendations(candidates, "USD")

	// CMA CGM should be CHEAPEST
	cma := res[1]
	hasCheapest := false
	for _, tag := range cma.RecommendationTags {
		if tag == "CHEAPEST" {
			hasCheapest = true
		}
	}
	if !hasCheapest {
		t.Errorf("expected CMA CGM to be tagged as CHEAPEST, got: %v", cma.RecommendationTags)
	}

	// Maersk should be FASTEST
	maersk := res[2]
	hasFastest := false
	for _, tag := range maersk.RecommendationTags {
		if tag == "FASTEST" {
			hasFastest = true
		}
	}
	if !hasFastest {
		t.Errorf("expected Maersk to be tagged as FASTEST, got: %v", maersk.RecommendationTags)
	}
}

func TestBuildQuotationRateSnapshot(t *testing.T) {
	candidate := &QuotationRateCandidate{
		SourceType:      "MANAGED_RATE",
		RateID:          ptrInt64(101),
		CarrierName:     "MSC Mediterranean",
		CarrierCode:     "MSCU",
		RateType:        "CONTRACT",
		RateVersion:     2,
		ContractID:      ptrInt64(50),
		Origin:          "USNYC",
		Destination:     "INNSA",
		TransportMode:   "Ocean FCL",
		ServiceType:     "FCL",
		EquipmentType:   "40GP",
		Currency:        "USD",
		BaseRate:        2000.0,
		TotalCharges:    200.0,
		CommercialTotal: 2200.0,
		TransitDays:     22,
	}

	snap, err := BuildQuotationRateSnapshot(1, 55, 12, candidate, "test_user")
	if err != nil {
		t.Fatalf("unexpected snapshot build error: %v", err)
	}

	if snap.OrgID != 1 || snap.QuotationID != 55 || snap.QuotationRateSelectionID != 12 {
		t.Errorf("snapshot metadata mismatch")
	}
	if snap.CarrierName != "MSC Mediterranean" || snap.CommercialTotal != 2200.0 {
		t.Errorf("snapshot pricing mismatch: got %s total %.2f", snap.CarrierName, snap.CommercialTotal)
	}
	if snap.SourceRateVersion == nil || *snap.SourceRateVersion != 2 {
		t.Errorf("expected source rate version 2")
	}
	if snap.PricingSnapshotJSON == "" {
		t.Errorf("expected non-empty pricing snapshot JSON")
	}
}

func ptrInt64(v int64) *int64 { return &v }
