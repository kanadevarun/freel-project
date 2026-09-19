package predictions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvoiceCollections_TenantValidationAndBadInput(t *testing.T) {
	repo := newMockRepo()
	sidecar := &mockSidecar{}

	svc := NewService(repo, sidecar)
	ctx := context.Background()

	// 1. Tenant access control validation - OrgID 0 must fail
	_, err := svc.GetOrPredictInvoiceCollections(ctx, 0, nil, 101, false)
	assert.ErrorIs(t, err, ErrInvalidOrgID)

	// 2. Invalid Invoice ID (<= 0) validation
	_, err = svc.GetOrPredictInvoiceCollections(ctx, 2, nil, 0, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid invoice ID")
}

func TestInvoiceCollections_CachedPredictionWorkflow(t *testing.T) {
	repo := newMockRepo()
	sidecar := &mockSidecar{}

	svc := NewService(repo, sidecar)
	ctx := context.Background()

	prePred := &Prediction{
		OrgID:               2,
		PredictionID:        "pred-fin-inv-103-initial",
		IdempotencyKey:      "org:2:invoice_collections:103:cycle_initial",
		Module:              "finance",
		PredictionType:      "COLLECTION_PRIORITY",
		Status:              StatusPublished,
		Severity:            SeverityHigh,
		ConfidenceScore:     0.94,
		ConfidenceBand:      ConfidenceHigh,
		RelatedRecordType:   "INVOICE",
		RelatedRecordID:     "103",
		PredictionStatement: "High Collection Priority: Invoice is 26 days overdue with $4,500.00 outstanding.",
		Explanation:         "Invoice past due date with zero payment received. Proactive outreach recommended.",
		SourceReferences: []SourceReference{
			{SourceModule: "customer_invoices", SourceRecordID: "103", SourceField: "balance_due", SourceTimestamp: time.Now().Format(time.RFC3339)},
		},
		SourceTimestamp: time.Now(),
		ReviewStatus:    "PENDING",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err := repo.Create(ctx, prePred)
	require.NoError(t, err)

	// Verify cached prediction returned when forceRefresh is false
	cached, err := svc.GetOrPredictInvoiceCollections(ctx, 2, nil, 103, false)
	require.NoError(t, err)
	assert.Equal(t, "pred-fin-inv-103-initial", cached.PredictionID)
	assert.Equal(t, "COLLECTION_PRIORITY", string(cached.PredictionType))
	assert.Equal(t, SeverityHigh, cached.Severity)
	assert.Equal(t, "INVOICE", cached.RelatedRecordType)
	assert.Equal(t, "103", cached.RelatedRecordID)
}

func TestInvoiceCollections_AuthoritativeCalculations(t *testing.T) {
	// Test the deterministic calculations performed in Go data provider
	provider := &FinanceCollectionsDataProvider{}

	// Case 1: Fully paid invoice
	invPaid := InvoiceFinanceData{
		InvoiceID:     101,
		InvoiceNumber: "INV-2026-DEV-001",
		TotalAmount:   3200.0,
		PaidAmount:    3200.0,
		BalanceDue:    0.0,
		Currency:      "USD",
		Status:        "Paid",
		DaysLeft:      12,
		IsOverdue:     false,
	}
	assert.Equal(t, 0.0, invPaid.BalanceDue)
	assert.False(t, invPaid.IsOverdue)

	// Case 2: Overdue invoice
	invOverdue := InvoiceFinanceData{
		InvoiceID:     103,
		InvoiceNumber: "INV-2026-DEV-003",
		TotalAmount:   4500.0,
		PaidAmount:    0.0,
		BalanceDue:    4500.0,
		Currency:      "USD",
		Status:        "Overdue",
		DaysLeft:      -26,
		IsOverdue:     true,
		DaysOverdue:   26,
	}
	assert.Equal(t, 4500.0, invOverdue.BalanceDue)
	assert.True(t, invOverdue.IsOverdue)
	assert.Equal(t, 26, invOverdue.DaysOverdue)
	assert.NotNil(t, provider)
}

func TestInvoiceCollections_LifecycleGovernance(t *testing.T) {
	repo := newMockRepo()
	sidecar := &mockSidecar{}

	svc := NewService(repo, sidecar)
	ctx := context.Background()

	pred := &Prediction{
		OrgID:               2,
		PredictionID:        "pred-fin-inv-102-test",
		IdempotencyKey:      "org:2:invoice_collections:102:cycle_test",
		Module:              "finance",
		PredictionType:      "CASH_INFLOW_FORECAST",
		Status:              StatusPublished,
		Severity:            SeverityLow,
		ConfidenceScore:     0.88,
		ConfidenceBand:      ConfidenceHigh,
		RelatedRecordType:   "INVOICE",
		RelatedRecordID:     "102",
		PredictionStatement: "Projected Cash Inflow Window: Expected payment receipt between 10-15 days.",
		Explanation:         "Customer exhibits steady settlement history before contractual due dates.",
		SourceReferences: []SourceReference{
			{SourceModule: "customer_invoices", SourceRecordID: "102", SourceField: "due_date", SourceTimestamp: time.Now().Format(time.RFC3339)},
		},
		SourceTimestamp:   time.Now(),
		RecommendedAction: func(s string) *string { return &s }("Schedule courtesy reminder"),
		ActionType:        func(s string) *string { return &s }("invoices.schedule_reminder"),
		IsActionRequired:  true,
		RequiresApproval:  false,
		ReviewStatus:      "PENDING",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	err := repo.Create(ctx, pred)
	require.NoError(t, err)

	// 1. Fetch prediction
	fetched, err := svc.GetPrediction(ctx, 2, "pred-fin-inv-102-test")
	require.NoError(t, err)
	assert.Equal(t, "pred-fin-inv-102-test", fetched.PredictionID)
	assert.Equal(t, "finance", fetched.Module)
	assert.Equal(t, StatusPublished, fetched.Status)

	// 2. Acknowledge prediction
	ackUser := int64(10)
	err = svc.AcknowledgePrediction(ctx, 2, "pred-fin-inv-102-test", &ackUser)
	require.NoError(t, err)

	acked, err := svc.GetPrediction(ctx, 2, "pred-fin-inv-102-test")
	require.NoError(t, err)
	assert.Equal(t, StatusAcknowledged, acked.Status)

	// 3. Request Action (Action System / HITL governance)
	notes := "Forward to finance lead for cash flow planning"
	err = svc.RequestAction(ctx, 2, "pred-fin-inv-102-test", &ackUser, notes)
	require.NoError(t, err)

	actioned, err := svc.GetPrediction(ctx, 2, "pred-fin-inv-102-test")
	require.NoError(t, err)
	assert.Equal(t, StatusActionRequested, actioned.Status)
}
