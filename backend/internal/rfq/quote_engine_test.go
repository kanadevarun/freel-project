package rfq

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/rfq/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 1. Lifecycle State Machine Valid & Invalid Transitions
func TestQuote_Lifecycle_Transitions(t *testing.T) {
	// Valid Transitions
	assert.NoError(t, ValidateQuoteTransition(spec.QuoteStatusDraft, spec.QuoteStatusReceived))
	assert.NoError(t, ValidateQuoteTransition(spec.QuoteStatusReceived, spec.QuoteStatusUnderReview))
	assert.NoError(t, ValidateQuoteTransition(spec.QuoteStatusUnderReview, spec.QuoteStatusRecommended))
	assert.NoError(t, ValidateQuoteTransition(spec.QuoteStatusRecommended, spec.QuoteStatusApproved))
	assert.NoError(t, ValidateQuoteTransition(spec.QuoteStatusApproved, spec.QuoteStatusSelectedForCustomer))
	assert.NoError(t, ValidateQuoteTransition(spec.QuoteStatusReceived, spec.QuoteStatusRejected))
	assert.NoError(t, ValidateQuoteTransition(spec.QuoteStatusReceived, spec.QuoteStatusWithdrawn))
	assert.NoError(t, ValidateQuoteTransition(spec.QuoteStatusDraft, spec.QuoteStatusDraft)) // Idempotent

	// Invalid Transitions
	assert.Error(t, ValidateQuoteTransition(spec.QuoteStatusDraft, spec.QuoteStatusApproved))
	assert.Error(t, ValidateQuoteTransition(spec.QuoteStatusDraft, spec.QuoteStatusSelectedForCustomer))
	assert.Error(t, ValidateQuoteTransition(spec.QuoteStatusSelectedForCustomer, spec.QuoteStatusDraft))
}

// 2. Deterministic & Safe Margin Calculations
func TestQuote_MarginCalculations_Safe(t *testing.T) {
	// Standard margin
	marginAmt, marginPct := CalculateQuoteMargin(4200.0, 5000.0)
	assert.Equal(t, 800.0, marginAmt)
	assert.Equal(t, 16.0, marginPct)

	// Zero sell price (Safe division by zero avoidance)
	marginAmt, marginPct = CalculateQuoteMargin(4200.0, 0.0)
	assert.Equal(t, -4200.0, marginAmt)
	assert.Equal(t, 0.0, marginPct)

	// Negative sell price
	marginAmt, marginPct = CalculateQuoteMargin(4200.0, -100.0)
	assert.Equal(t, -4300.0, marginAmt)
	assert.Equal(t, 0.0, marginPct)

	// Negative margin (buy > sell)
	marginAmt, marginPct = CalculateQuoteMargin(5000.0, 4500.0)
	assert.Equal(t, -500.0, marginAmt)
	assert.Equal(t, -11.1, marginPct)
}

// 3. Validity Evaluation (Valid, Expiring Soon, Expired)
func TestQuote_ValidityEvaluation(t *testing.T) {
	now := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)

	// Valid (> 7 days)
	futureDate := now.Add(15 * 24 * time.Hour)
	status, days := EvaluateQuoteValidity(&futureDate, now)
	assert.Equal(t, spec.ValidityValid, status)
	assert.NotNil(t, days)
	assert.Equal(t, 15, *days)

	// Expiring Soon (3 days)
	soonDate := now.Add(3 * 24 * time.Hour)
	status, days = EvaluateQuoteValidity(&soonDate, now)
	assert.Equal(t, spec.ValidityExpiringSoon, status)
	assert.NotNil(t, days)
	assert.Equal(t, 3, *days)

	// Expired (2 days ago)
	pastDate := now.Add(-2 * 24 * time.Hour)
	status, days = EvaluateQuoteValidity(&pastDate, now)
	assert.Equal(t, spec.ValidityExpired, status)
	assert.NotNil(t, days)
	assert.True(t, *days <= 0)

	// Nil validUntil
	status, days = EvaluateQuoteValidity(nil, now)
	assert.Equal(t, spec.ValidityValid, status)
	assert.Nil(t, days)
}

// 4. Multi-Carrier Comparison Intelligence (Lowest Cost, Best Margin, Fastest Transit)
func TestQuote_Comparison_LowestCost_BestMargin_FastestTransit(t *testing.T) {
	now := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)
	validDate := now.Add(10 * 24 * time.Hour)
	transitA := 24
	transitB := 18
	transitC := 28

	quotes := []spec.RFQQuote{
		{
			ID:              101,
			RFQID:           1,
			CarrierName:     "Maersk Line",
			Currency:        "USD",
			BuyPrice:        4200.0,
			SellPrice:       5000.0,
			TransitTimeDays: &transitA,
			ValidUntil:      &validDate,
			Status:          spec.QuoteStatusReceived,
		},
		{
			ID:              102,
			RFQID:           1,
			CarrierName:     "Hapag-Lloyd",
			Currency:        "USD",
			BuyPrice:        3900.0, // Lowest Buy
			SellPrice:       4900.0, // Margin: 1000 (20.4%)
			TransitTimeDays: &transitB, // Fastest Transit
			ValidUntil:      &validDate,
			Status:          spec.QuoteStatusReceived,
		},
		{
			ID:              103,
			RFQID:           1,
			CarrierName:     "CMA CGM",
			Currency:        "USD",
			BuyPrice:        4500.0,
			SellPrice:       5800.0, // Margin: 1300 (22.4%) - Highest Margin
			TransitTimeDays: &transitC,
			ValidUntil:      &validDate,
			Status:          spec.QuoteStatusReceived,
		},
	}

	resp := BuildQuotesResponse(&spec.RFQ{ID: 1}, quotes, nil, now)
	require.NotNil(t, resp)
	assert.Equal(t, 3, resp.Summary.TotalQuotes)
	assert.Equal(t, 3, resp.Summary.ReceivedQuotes)

	require.NotNil(t, resp.Summary.LowestBuyAmount)
	assert.Equal(t, 3900.0, *resp.Summary.LowestBuyAmount)

	require.NotNil(t, resp.Summary.HighestMarginAmount)
	assert.Equal(t, 1300.0, *resp.Summary.HighestMarginAmount)

	require.NotNil(t, resp.Summary.FastestTransitDays)
	assert.Equal(t, 18, *resp.Summary.FastestTransitDays)

	// Check comparison items
	require.Len(t, resp.Comparison, 3)

	// Quote 102 should be marked lowest cost and fastest
	assert.False(t, resp.Comparison[0].IsLowestCost)
	assert.True(t, resp.Comparison[1].IsLowestCost)
	assert.True(t, resp.Comparison[1].IsFastest)
	assert.False(t, resp.Comparison[1].IsHighestMargin)

	// Quote 103 should be marked highest margin
	assert.True(t, resp.Comparison[2].IsHighestMargin)
	assert.False(t, resp.Comparison[2].IsLowestCost)
	assert.False(t, resp.Comparison[2].IsFastest)
}

// 5. Expired / Rejected Quotes Excluded from Comparison Benchmarks
func TestQuote_ExclusionOfExpiredRejectedWithdrawn(t *testing.T) {
	now := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)
	pastDate := now.Add(-5 * 24 * time.Hour)
	futureDate := now.Add(10 * 24 * time.Hour)
	transit := 14

	quotes := []spec.RFQQuote{
		{
			ID:              201,
			CarrierName:     "Expired Carrier",
			Currency:        "USD",
			BuyPrice:        1000.0, // Extremely low buy, but expired!
			SellPrice:       3000.0,
			TransitTimeDays: &transit,
			ValidUntil:      &pastDate,
			Status:          spec.QuoteStatusReceived,
		},
		{
			ID:              202,
			CarrierName:     "Rejected Carrier",
			Currency:        "USD",
			BuyPrice:        1200.0, // Low buy, but rejected!
			SellPrice:       2500.0,
			TransitTimeDays: &transit,
			ValidUntil:      &futureDate,
			Status:          spec.QuoteStatusRejected,
		},
		{
			ID:              203,
			CarrierName:     "Valid Carrier",
			Currency:        "USD",
			BuyPrice:        4000.0,
			SellPrice:       4800.0,
			TransitTimeDays: &transit,
			ValidUntil:      &futureDate,
			Status:          spec.QuoteStatusReceived,
		},
	}

	resp := BuildQuotesResponse(&spec.RFQ{ID: 1}, quotes, nil, now)
	require.NotNil(t, resp)

	// Lowest buy should be 4000 (Valid Carrier), NOT 1000 or 1200
	require.NotNil(t, resp.Summary.LowestBuyAmount)
	assert.Equal(t, 4000.0, *resp.Summary.LowestBuyAmount)

	assert.Equal(t, 1, resp.Summary.ExpiredQuotes)
}

// 6. Currency Safety in Comparison
func TestQuote_MixedCurrencySafety(t *testing.T) {
	now := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)
	futureDate := now.Add(10 * 24 * time.Hour)

	quotes := []spec.RFQQuote{
		{
			ID:          301,
			CarrierName: "Carrier USD 1",
			Currency:    "USD",
			BuyPrice:    4000.0,
			SellPrice:   4800.0,
			ValidUntil:  &futureDate,
			Status:      spec.QuoteStatusReceived,
		},
		{
			ID:          302,
			CarrierName: "Carrier EUR 1",
			Currency:    "EUR",
			BuyPrice:    3500.0, // EUR, different currency
			SellPrice:   4200.0,
			ValidUntil:  &futureDate,
			Status:      spec.QuoteStatusReceived,
		},
	}

	resp := BuildQuotesResponse(&spec.RFQ{ID: 1}, quotes, nil, now)
	require.NotNil(t, resp)
	assert.True(t, resp.Summary.HasMixedCurrencies)
	assert.Equal(t, "USD", resp.Summary.PrimaryCurrency)

	// Lowest buy only compares USD
	require.NotNil(t, resp.Summary.LowestBuyAmount)
	assert.Equal(t, 4000.0, *resp.Summary.LowestBuyAmount)
}

// 7. Full MySQL Database Workflow & Org Isolation Integration Test
func TestQuote_DatabaseWorkflow_OrgIsolation(t *testing.T) {
	db, err := database.Connect("root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC&multiStatements=true")
	if err != nil {
		t.Skip("Skipping DB test: MySQL connection unavailable")
		return
	}
	defer db.Close()

	orgA := int32(8801)
	orgB := int32(8802)
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (8801, 'Test Org 8801', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)")
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (8802, 'Test Org 8802', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)")

	dl := NewDataLayer(db)
	ctx := context.Background()


	// 1. Create RFQ in Org A
	rfqA := &spec.RFQ{
		OrgID:      orgA,
		RFQNumber:  fmt.Sprintf("RFQ-QT-%d", time.Now().UnixNano()),
		CustomerID: 1,
		Stage:      spec.StageRFQCreated,
	}
	err = dl.CreateRFQ(ctx, rfqA)
	require.NoError(t, err)
	require.True(t, rfqA.ID > 0)

	// 2. Create Carrier Quote in Org A
	refA := "MSK-QT-801"
	transit := 21
	validFuture := time.Now().Add(14 * 24 * time.Hour)
	quoteA := &spec.RFQQuote{
		RFQID:           rfqA.ID,
		OrgID:           orgA,
		CarrierName:     "Maersk",
		QuoteReference:  &refA,
		Currency:        "USD",
		BuyPrice:        4500.0,
		SellPrice:       5400.0,
		TransitTimeDays: &transit,
		ValidUntil:      &validFuture,
		Status:          spec.QuoteStatusReceived,
		Charges: []spec.QuoteCharge{
			{Type: spec.ChargeTypeFreight, Description: "Base Ocean Freight", Amount: 4000.0, Currency: "USD"},
			{Type: spec.ChargeTypeFuel, Description: "Bunker Adjustment Factor", Amount: 500.0, Currency: "USD"},
		},
	}
	err = dl.CreateRFQQuote(ctx, orgA, quoteA)
	require.NoError(t, err)
	require.True(t, quoteA.ID > 0)

	// 3. Org A can retrieve the quote
	quotesA, err := dl.GetRFQQuotes(ctx, orgA, rfqA.ID)
	require.NoError(t, err)
	require.Len(t, quotesA, 1)
	assert.Equal(t, "Maersk", quotesA[0].CarrierName)
	assert.Equal(t, 2, len(quotesA[0].Charges))

	// 4. Strict Org Isolation: Org B CANNOT read Org A quotes
	quotesB, err := dl.GetRFQQuotes(ctx, orgB, rfqA.ID)
	require.NoError(t, err)
	assert.Len(t, quotesB, 0)

	// Org B CANNOT get quote by ID
	quoteBByID, err := dl.GetRFQQuoteByID(ctx, orgB, rfqA.ID, quoteA.ID)
	assert.Error(t, err)
	assert.Nil(t, quoteBByID)

	// Org B CANNOT update Org A quote status
	err = dl.UpdateRFQQuoteStatus(ctx, orgB, rfqA.ID, quoteA.ID, spec.QuoteStatusUnderReview)
	assert.Error(t, err)

	// Org B CANNOT recommend Org A quote
	err = dl.RecommendRFQQuote(ctx, orgB, rfqA.ID, quoteA.ID)
	assert.Error(t, err)

	// 5. Org A Recommends Quote A
	err = dl.RecommendRFQQuote(ctx, orgA, rfqA.ID, quoteA.ID)
	require.NoError(t, err)

	updatedQuote, err := dl.GetRFQQuoteByID(ctx, orgA, rfqA.ID, quoteA.ID)
	require.NoError(t, err)
	assert.True(t, updatedQuote.IsRecommended)
	assert.Equal(t, spec.QuoteStatusRecommended, updatedQuote.Status)

	// 6. Org A Approves Quote A
	err = dl.ApproveRFQQuote(ctx, orgA, rfqA.ID, quoteA.ID, "Lead Pricing Manager")
	require.NoError(t, err)

	approvedQuote, err := dl.GetRFQQuoteByID(ctx, orgA, rfqA.ID, quoteA.ID)
	require.NoError(t, err)
	assert.Equal(t, spec.QuoteStatusApproved, approvedQuote.Status)
	assert.NotNil(t, approvedQuote.ApprovedBy)
	assert.Equal(t, "Lead Pricing Manager", *approvedQuote.ApprovedBy)
	assert.NotNil(t, approvedQuote.ApprovedAt)

	// 7. Verify RFQ Stage advanced to STAGE_QUOTE_GENERATED
	rfqRefetched, err := dl.GetRFQByID(ctx, orgA, rfqA.ID)
	require.NoError(t, err)
	assert.Equal(t, spec.StageQuoteGenerated, rfqRefetched.Stage)
}
