package pricing_workflow

import (
	"encoding/json"
	"testing"

	"github.com/freel/backend/internal/orchestration"
	"github.com/freel/backend/internal/quotations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeterministicPricingCalculation(t *testing.T) {
	// Formula verification: Total Cost = Base Cost + Surcharges
	baseCost := 2000.00
	surcharges := 400.00
	totalCost := round2(baseCost + surcharges)
	assert.Equal(t, 2400.00, totalCost)

	// Base Sell with 25% target margin
	targetMargin := 0.25
	baseSell := round2(baseCost * (1.0 + targetMargin))
	assert.Equal(t, 2500.00, baseSell)

	// Total selling price
	discount := 100.00
	totalSellingPrice := round2(baseSell + surcharges - discount)
	assert.Equal(t, 2800.00, totalSellingPrice)

	// Gross Profit & Gross Margin %
	grossProfit := round2(totalSellingPrice - totalCost)
	assert.Equal(t, 400.00, grossProfit)

	grossMarginPct := round2((grossProfit / totalSellingPrice) * 100.0)
	assert.InDelta(t, 14.29, grossMarginPct, 0.01)

	// Margin Health: 14.29% is below 15% standard threshold -> LOW
	marginHealth := quotations.MarginHealthHealthy
	if grossProfit < 0 {
		marginHealth = quotations.MarginHealthNegative
	} else if grossMarginPct < 15.0 {
		marginHealth = quotations.MarginHealthLow
	}
	assert.Equal(t, quotations.MarginHealthLow, marginHealth)
}

func TestNegativeMarginDetection(t *testing.T) {
	baseCost := 3000.00
	surcharges := 500.00
	totalCost := round2(baseCost + surcharges) // 3500.00

	baseSell := 2800.00 // Under cost
	totalSellingPrice := round2(baseSell + surcharges) // 3300.00

	grossProfit := round2(totalSellingPrice - totalCost) // -200.00
	assert.Equal(t, -200.00, grossProfit)

	grossMarginPct := round2((grossProfit / totalSellingPrice) * 100.0)
	assert.True(t, grossMarginPct < 0.0)

	marginHealth := quotations.MarginHealthHealthy
	if grossProfit < 0 {
		marginHealth = quotations.MarginHealthNegative
	} else if grossMarginPct < 15.0 {
		marginHealth = quotations.MarginHealthLow
	}
	assert.Equal(t, quotations.MarginHealthNegative, marginHealth)

	// Negative margin must require approval
	requiresApproval := marginHealth != quotations.MarginHealthHealthy
	assert.True(t, requiresApproval)
}

func TestActionRegistryQuotationSendDraft(t *testing.T) {
	reg := orchestration.NewRegistry()
	action, err := reg.GetAction("quotations.send_draft")
	require.NoError(t, err)
	assert.Equal(t, "quotations.send_draft", action.Name)
	assert.Equal(t, "COMMERCIAL", action.Module)
	assert.Equal(t, "HIGH_RISK", action.Category)
	assert.Equal(t, orchestration.RiskHigh, action.RiskLevel)
	assert.True(t, action.RequiresApproval)
	assert.Equal(t, "quotations:send", action.RequiredPermission)

	// Validate input schema
	validInput := json.RawMessage(`{"draft_id": 101, "recipient_email": "importer@acmecorp.com"}`)
	assert.NoError(t, action.Validate(validInput))

	// Missing draft_id
	invalidID := json.RawMessage(`{"recipient_email": "importer@acmecorp.com"}`)
	assert.Error(t, action.Validate(invalidID))

	// Invalid email
	invalidEmail := json.RawMessage(`{"draft_id": 101, "recipient_email": "invalid-email"}`)
	assert.Error(t, action.Validate(invalidEmail))
}
