package rates

import (
	"testing"
	"time"

	"github.com/freel/backend/internal/rates/spec"
)

func TestCalculateSpotRateRequestStatus(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)

	req := &spec.SpotRateRequest{
		ID:               1,
		RequestReference: "SPOT-2026-001",
		RequiredByDate:   "2026-09-05",
		Status:           SpotRequestDraft,
	}

	// 1. Draft with no responses
	status := CalculateSpotRateRequestStatus(req, nil, now)
	if status != SpotRequestDraft {
		t.Errorf("Expected draft status, got %s", status)
	}

	// 2. Sent with 1 response -> Partially Responded
	req.Status = SpotRequestSent
	resp1 := &spec.SpotRateResponse{
		ID:          10,
		CarrierName: "MSC",
		Status:      SpotResponseReceived,
		TotalAmount: 2200,
	}
	status = CalculateSpotRateRequestStatus(req, []*spec.SpotRateResponse{resp1}, now)
	if status != SpotRequestPartiallyResponded {
		t.Errorf("Expected PARTIALLY_RESPONDED, got %s", status)
	}

	// 3. Sent with 2 responses -> Responded
	resp2 := &spec.SpotRateResponse{
		ID:          11,
		CarrierName: "Maersk",
		Status:      SpotResponseReceived,
		TotalAmount: 2400,
	}
	status = CalculateSpotRateRequestStatus(req, []*spec.SpotRateResponse{resp1, resp2}, now)
	if status != SpotRequestResponded {
		t.Errorf("Expected RESPONDED, got %s", status)
	}

	// 4. One response marked preferred -> Selected
	resp1.IsPreferred = true
	status = CalculateSpotRateRequestStatus(req, []*spec.SpotRateResponse{resp1, resp2}, now)
	if status != SpotRequestSelected {
		t.Errorf("Expected SELECTED, got %s", status)
	}

	// 5. Expired date
	reqPast := &spec.SpotRateRequest{
		ID:             2,
		RequiredByDate: "2026-08-01",
		Status:         SpotRequestSent,
	}
	status = CalculateSpotRateRequestStatus(reqPast, nil, now)
	if status != SpotRequestExpired {
		t.Errorf("Expected EXPIRED, got %s", status)
	}
}

func TestCalculateSpotRateComparison(t *testing.T) {
	req := &spec.SpotRateRequest{
		ID:               100,
		RequestReference: "SPOT-2026-NYC-IND",
		OriginPort:       "USNYC",
		DestinationPort:  "INNSA",
		TransportMode:    "Ocean FCL",
		TargetCurrency:   "USD",
	}

	transit22 := 22
	transit18 := 18
	transit28 := 28

	responses := []*spec.SpotRateResponse{
		{
			ID:                   1,
			CarrierName:          "MSC Mediterranean Shipping",
			Currency:             "USD",
			BaseAmount:           2100,
			TotalAmount:          2350,
			TransitDays:          &transit22,
			FreeDaysDestination: 14,
			ValidUntil:           "2026-09-30",
			Status:               SpotResponseReceived,
		},
		{
			ID:                   2,
			CarrierName:          "Maersk Line",
			Currency:             "USD",
			BaseAmount:           2400,
			TotalAmount:          2700,
			TransitDays:          &transit18,
			FreeDaysDestination: 7,
			ValidUntil:           "2026-09-30",
			Status:               SpotResponseReceived,
		},
		{
			ID:                   3,
			CarrierName:          "CMA CGM",
			Currency:             "USD",
			BaseAmount:           1900,
			TotalAmount:          2150,
			TransitDays:          &transit28,
			FreeDaysDestination: 10,
			ValidUntil:           "2026-09-30",
			Status:               SpotResponseReceived,
			IsPreferred:          true,
		},
	}

	comp := CalculateSpotRateComparison(req, responses)

	// CMA CGM (2150) is cheapest
	if comp.CheapestCarrierName == nil || *comp.CheapestCarrierName != "CMA CGM" {
		t.Errorf("Expected CMA CGM as cheapest, got %v", comp.CheapestCarrierName)
	}

	// Maersk (18 days) is fastest
	if comp.FastestCarrierName == nil || *comp.FastestCarrierName != "Maersk Line" {
		t.Errorf("Expected Maersk as fastest, got %v", comp.FastestCarrierName)
	}

	// Preferred is CMA CGM
	if comp.PreferredCarrierName == nil || *comp.PreferredCarrierName != "CMA CGM" {
		t.Errorf("Expected CMA CGM as preferred, got %v", comp.PreferredCarrierName)
	}

	if comp.IsMultiCurrency {
		t.Errorf("Expected single currency USD comparison")
	}
}

func TestCalculateSpotRateComparison_MultiCurrencySafety(t *testing.T) {
	req := &spec.SpotRateRequest{
		ID:               101,
		RequestReference: "SPOT-EUR-USD",
		OriginPort:       "NLRTM",
		DestinationPort:  "SGSIN",
		TargetCurrency:   "USD",
	}

	transit25 := 25
	responses := []*spec.SpotRateResponse{
		{
			ID:          1,
			CarrierName: "Hapag-Lloyd",
			Currency:    "EUR",
			TotalAmount: 2000,
			TransitDays: &transit25,
			Status:      SpotResponseReceived,
		},
		{
			ID:          2,
			CarrierName: "ONE Ocean Network Express",
			Currency:    "USD",
			TotalAmount: 2400,
			TransitDays: &transit25,
			Status:      SpotResponseReceived,
		},
	}

	comp := CalculateSpotRateComparison(req, responses)
	if !comp.IsMultiCurrency {
		t.Errorf("Expected IsMultiCurrency = true for EUR vs USD responses")
	}
}
