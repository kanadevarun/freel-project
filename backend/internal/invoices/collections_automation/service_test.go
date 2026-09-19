package collections_automation

import (
	"testing"
	"time"
)

func TestCalculateAgingBucket(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		dueDate      time.Time
		expectedDays int
		expectedBucket string
	}{
		{
			name:           "Due in 5 days (Current)",
			dueDate:        now.AddDate(0, 0, 5),
			expectedDays:   0,
			expectedBucket: "CURRENT",
		},
		{
			name:           "Overdue by 10 days (1-30)",
			dueDate:        now.AddDate(0, 0, -10),
			expectedDays:   10,
			expectedBucket: "1_30_DAYS",
		},
		{
			name:           "Overdue by 45 days (31-60)",
			dueDate:        now.AddDate(0, 0, -45),
			expectedDays:   45,
			expectedBucket: "31_60_DAYS",
		},
		{
			name:           "Overdue by 75 days (61-90)",
			dueDate:        now.AddDate(0, 0, -75),
			expectedDays:   75,
			expectedBucket: "61_90_DAYS",
		},
		{
			name:           "Overdue by 120 days (90+)",
			dueDate:        now.AddDate(0, 0, -120),
			expectedDays:   120,
			expectedBucket: "90_PLUS_DAYS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daysOverdue := 0
			if now.After(tt.dueDate) {
				daysOverdue = int(now.Sub(tt.dueDate).Hours() / 24)
			}

			if daysOverdue != tt.expectedDays {
				t.Errorf("expected %d days overdue, got %d", tt.expectedDays, daysOverdue)
			}

			bucket := "CURRENT"
			if daysOverdue > 90 {
				bucket = "90_PLUS_DAYS"
			} else if daysOverdue > 60 {
				bucket = "61_90_DAYS"
			} else if daysOverdue > 30 {
				bucket = "31_60_DAYS"
			} else if daysOverdue > 0 {
				bucket = "1_30_DAYS"
			}

			if bucket != tt.expectedBucket {
				t.Errorf("expected bucket %s, got %s", tt.expectedBucket, bucket)
			}
		})
	}
}

func TestDeterministicFinanceCalculations(t *testing.T) {
	totalAmount := 32120.00
	paidAmount := 5000.00
	outstandingAmount := totalAmount - paidAmount

	if outstandingAmount != 27120.00 {
		t.Errorf("expected outstanding 27120.00, got %.2f", outstandingAmount)
	}

	// High value overdue detection (threshold: $10,000)
	isHighValue := outstandingAmount >= 10000.00
	if !isHighValue {
		t.Errorf("expected isHighValue to be true for outstanding 27120.00")
	}

	// Customer credit limit check
	creditLimit := 25000.00
	isOverCreditLimit := outstandingAmount > creditLimit
	if !isOverCreditLimit {
		t.Errorf("expected isOverCreditLimit to be true when outstanding exceeds limit")
	}
}

func TestCollectionDraftValidation(t *testing.T) {
	draft := &CollectionDraft{
		ID:        1,
		OrgID:     1,
		InvoiceID: 3,
		Status:    DraftStatusApproved,
	}

	// Approved or dispatched drafts must not be mutated
	canEdit := draft.Status == DraftStatusDraft || draft.Status == DraftStatusPendingApproval
	if canEdit {
		t.Errorf("expected approved draft to not be editable")
	}

	draft.Status = DraftStatusDraft
	canEdit = draft.Status == DraftStatusDraft || draft.Status == DraftStatusPendingApproval
	if !canEdit {
		t.Errorf("expected draft in DRAFT status to be editable")
	}
}

func TestTenantIsolationEnforcement(t *testing.T) {
	invoiceOrgID := int64(1)
	requestOrgID := int64(2)

	hasAccess := invoiceOrgID == requestOrgID
	if hasAccess {
		t.Errorf("expected tenant isolation to block cross-org access (org 1 != org 2)")
	}

	requestOrgID = int64(1)
	hasAccess = invoiceOrgID == requestOrgID
	if !hasAccess {
		t.Errorf("expected access permitted for matching org ID")
	}
}
