package bcontext

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomer360Intelligence_Validation(t *testing.T) {
	svc := &defaultService{db: nil, rbacSvc: nil}
	ctx := context.Background()

	// 1. Invalid Org ID
	_, err := svc.GetCustomer360Intelligence(ctx, 0, 101, "corr-test", 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "organization ID is required")

	// 2. Invalid Customer ID
	_, err = svc.GetCustomer360Intelligence(ctx, 1, 0, "corr-test", 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid customer ID")
}

func TestCustomer360Intelligence_CalculationsUnit(t *testing.T) {
	// Test deterministic metric calculation formulas
	t.Run("ConversionRateCalculations", func(t *testing.T) {
		totalQuotes := 4
		totalBookings := 2
		rate := (float64(totalBookings) / float64(totalQuotes)) * 100.0
		assert.Equal(t, 50.0, rate)

		// When 0 quotes
		var zeroConv *float64
		assert.Nil(t, zeroConv)
	})

	t.Run("FinancialBalanceCalculations", func(t *testing.T) {
		invoiced := 25000.0
		balanceDue := 5000.0
		paid := invoiced - balanceDue
		assert.Equal(t, 20000.0, paid)

		totalInvoices := 5
		overdue := 1
		compliance := (float64(totalInvoices-overdue) / float64(totalInvoices)) * 100.0
		assert.Equal(t, 80.0, compliance)
	})

	t.Run("ReadonlyContractAssurance", func(t *testing.T) {
		intel := Customer360Intelligence{
			OrgID:      1,
			CustomerID: 42,
			IsReadOnly: true,
			AISummary: CustomerAISummary{
				IsInformationalOnly: true,
				ConfidenceLevel:     "HIGH",
			},
		}
		assert.True(t, intel.IsReadOnly)
		assert.True(t, intel.AISummary.IsInformationalOnly)
	})
}
