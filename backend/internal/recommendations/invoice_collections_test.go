package recommendations

import (
	"context"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func TestOverdueInvoiceAndAgingBandsDetection(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	ctx := context.Background()
	testOrg := int64(88950)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM customer_invoices WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM customers WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	// 1. Insert customer
	cRes, err := db.ExecContext(ctx, `
		INSERT INTO customers (org_id, name, status, created_at, updated_at)
		VALUES (?, 'Pacific Horizon Logistics', 'Active', NOW(), NOW())
	`, testOrg)
	if err != nil {
		t.Fatalf("Failed to insert customer: %v", err)
	}
	custID, _ := cRes.LastInsertId()

	// 2. Insert overdue invoice (due 25 days ago -> 16-30 days aging band)
	due25DaysAgo := time.Now().Add(-25 * 24 * time.Hour)
	invRes, err := db.ExecContext(ctx, `
		INSERT INTO customer_invoices (org_id, customer_id, customer_name, invoice_number, invoice_date, status, currency, total_amount, balance_due, due_date, created_at, updated_at)
		VALUES (?, ?, 'Pacific Horizon Logistics', 'INV-TEST-25D', NOW() - INTERVAL 30 DAY, 'Overdue', 'USD', 12500.00, 12500.00, ?, NOW() - INTERVAL 30 DAY, NOW())
	`, testOrg, custID, due25DaysAgo)
	if err != nil {
		t.Fatalf("Failed to insert overdue invoice: %v", err)
	}
	invID, _ := invRes.LastInsertId()

	// 3. Run generator
	result, err := gen.Generate(ctx, testOrg, "test-corr-overdue-aging", 1)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if result == nil {
		t.Fatalf("Expected non-nil result")
	}

	// 4. Query recommendations filtered by invoice_id
	recs, total, err := repo.List(ctx, testOrg, RecommendationFilter{
		InvoiceID: &invID,
	})
	if err != nil {
		t.Fatalf("Failed to list recommendations: %v", err)
	}
	if total == 0 || len(recs) == 0 {
		t.Fatalf("Expected overdue invoice recommendation, got 0")
	}

	rec := recs[0]
	if rec.InvoiceID == nil || *rec.InvoiceID != invID {
		t.Errorf("Expected InvoiceID %d, got %v", invID, rec.InvoiceID)
	}
	if rec.ActionType != ActionTypeReviewOverdueInvoice {
		t.Errorf("Expected ActionType %s, got %s", ActionTypeReviewOverdueInvoice, rec.ActionType)
	}
	if !strings.Contains(rec.Description, "16-30 days") {
		t.Errorf("Expected description to contain '16-30 days' aging band, got: %s", rec.Description)
	}

	// Verify Evidence
	rec.UnmarshalDetails()
	foundDaysOverdue := false
	for _, ev := range rec.Evidence {
		if ev.FieldName == "days_overdue" && strings.Contains(ev.ObservedValue.(string), "16-30 days") {
			foundDaysOverdue = true
			break
		}
	}
	if !foundDaysOverdue {
		t.Errorf("Expected evidence to contain days_overdue with aging band 16-30 days")
	}
}

func TestUpcomingCollectionPrioritiesDetection(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	ctx := context.Background()
	testOrg := int64(88951)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM customer_invoices WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM customers WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	cRes, _ := db.ExecContext(ctx, `
		INSERT INTO customers (org_id, name, status, created_at, updated_at)
		VALUES (?, 'Apex Global Cargo', 'Active', NOW(), NOW())
	`, testOrg)
	custID, _ := cRes.LastInsertId()

	// In 3 days (within 7-day window)
	dueIn3Days := time.Now().Add(3 * 24 * time.Hour)
	invRes, err := db.ExecContext(ctx, `
		INSERT INTO customer_invoices (org_id, customer_id, customer_name, invoice_number, invoice_date, status, currency, total_amount, balance_due, due_date, created_at, updated_at)
		VALUES (?, ?, 'Apex Global Cargo', 'INV-UPCOMING-001', NOW(), 'Issued', 'USD', 18500.00, 18500.00, ?, NOW(), NOW())
	`, testOrg, custID, dueIn3Days)
	if err != nil {
		t.Fatalf("Failed to insert upcoming invoice: %v", err)
	}
	invID, _ := invRes.LastInsertId()

	// Run generator
	_, err = gen.Generate(ctx, testOrg, "test-corr-upcoming", 1)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	upcomingFilter := true
	recs, total, err := repo.List(ctx, testOrg, RecommendationFilter{
		UpcomingOnly: &upcomingFilter,
		InvoiceID:    &invID,
	})
	if err != nil {
		t.Fatalf("Failed to list recommendations: %v", err)
	}
	if total == 0 || len(recs) == 0 {
		t.Fatalf("Expected upcoming invoice recommendation, got 0")
	}

	rec := recs[0]
	if rec.RuleApplied != "UPCOMING_COLLECTION_PRIORITY" {
		t.Errorf("Expected RuleApplied UPCOMING_COLLECTION_PRIORITY, got %s", rec.RuleApplied)
	}
	if rec.ActionType != ActionTypePreparePaymentReminder {
		t.Errorf("Expected ActionType %s, got %s", ActionTypePreparePaymentReminder, rec.ActionType)
	}
}

func TestPaymentRiskSignalsDetection(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	ctx := context.Background()
	testOrg := int64(88952)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM customer_invoices WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM customers WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	cRes, _ := db.ExecContext(ctx, `
		INSERT INTO customers (org_id, name, status, created_at, updated_at)
		VALUES (?, 'Nippon Express Partner', 'Active', NOW(), NOW())
	`, testOrg)
	custID, _ := cRes.LastInsertId()

	// 1. Stalled Partial Payment: total $10,000, balance $4,000, overdue
	pastDue := time.Now().Add(-10 * 24 * time.Hour)
	res1, _ := db.ExecContext(ctx, `
		INSERT INTO customer_invoices (org_id, customer_id, customer_name, invoice_number, invoice_date, status, currency, total_amount, balance_due, due_date, created_at, updated_at)
		VALUES (?, ?, 'Nippon Express Partner', 'INV-STALLED-01', NOW() - INTERVAL 20 DAY, 'Overdue', 'USD', 10000.00, 4000.00, ?, NOW() - INTERVAL 20 DAY, NOW())
	`, testOrg, custID, pastDue)
	invID1, _ := res1.LastInsertId()

	// 2. Unresolved Dispute
	res2, err := db.ExecContext(ctx, `
		INSERT INTO customer_invoices (org_id, customer_id, customer_name, invoice_number, invoice_date, status, currency, total_amount, balance_due, due_date, created_at, updated_at)
		VALUES (?, ?, 'Nippon Express Partner', 'INV-DISPUTE-01', NOW() - INTERVAL 15 DAY, 'Disputed', 'USD', 7500.00, 7500.00, ?, NOW() - INTERVAL 15 DAY, NOW())
	`, testOrg, custID, pastDue)
	if err != nil {
		t.Fatalf("Failed to insert dispute invoice: %v", err)
	}
	invID2, _ := res2.LastInsertId()

	// Run generator
	_, err = gen.Generate(ctx, testOrg, "test-corr-risk-signals", 1)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify stalled partial payment
	recs1, _, err := repo.List(ctx, testOrg, RecommendationFilter{InvoiceID: &invID1})
	if err != nil || len(recs1) == 0 {
		t.Fatalf("Expected stalled partial payment recommendation, got 0 (err: %v)", err)
	}
	if recs1[0].ActionType != ActionTypeVerifyPaymentStatus && recs1[0].ActionType != ActionTypeReviewOverdueInvoice {
		t.Errorf("Unexpected action type for stalled partial payment: %s", recs1[0].ActionType)
	}

	// Verify dispute
	recs2, _, err := repo.List(ctx, testOrg, RecommendationFilter{InvoiceID: &invID2})
	if err != nil || len(recs2) == 0 {
		t.Fatalf("Expected dispute recommendation, got 0 (err: %v)", err)
	}
	foundDispute := false
	for _, r := range recs2 {
		if r.ActionType == ActionTypeReviewInvoiceDispute {
			foundDispute = true
			break
		}
	}
	if !foundDispute {
		t.Errorf("Expected ActionTypeReviewInvoiceDispute in recommendations for invoice %d", invID2)
	}
}

func TestInvoiceDataQualityChecks(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	ctx := context.Background()
	testOrg := int64(88953)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM customer_invoices WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM customers WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	cRes, _ := db.ExecContext(ctx, `
		INSERT INTO customers (org_id, name, status, created_at, updated_at)
		VALUES (?, 'Kuehne Freight Desk', 'Active', NOW(), NOW())
	`, testOrg)
	custID, _ := cRes.LastInsertId()

	// 1. Non-positive total with positive balance
	res1, err := db.ExecContext(ctx, `
		INSERT INTO customer_invoices (org_id, customer_id, customer_name, invoice_number, invoice_date, status, currency, total_amount, balance_due, due_date, created_at, updated_at)
		VALUES (?, ?, 'Kuehne Freight Desk', 'INV-ZERO-TOTAL', NOW(), 'Issued', 'USD', 0.00, 5000.00, NOW() + INTERVAL 10 DAY, NOW(), NOW())
	`, testOrg, custID)
	if err != nil {
		t.Fatalf("Failed to insert zero total invoice: %v", err)
	}
	invID1, _ := res1.LastInsertId()

	// 2. Impossible totals: balance_due > total_amount
	res2, err := db.ExecContext(ctx, `
		INSERT INTO customer_invoices (org_id, customer_id, customer_name, invoice_number, invoice_date, status, currency, total_amount, balance_due, due_date, created_at, updated_at)
		VALUES (?, ?, 'Kuehne Freight Desk', 'INV-BAD-BALANCE', NOW(), 'Issued', 'USD', 3000.00, 9999.00, NOW() + INTERVAL 5 DAY, NOW(), NOW())
	`, testOrg, custID)
	if err != nil {
		t.Fatalf("Failed to insert bad balance invoice: %v", err)
	}
	invID2, _ := res2.LastInsertId()

	// Run generator
	_, err = gen.Generate(ctx, testOrg, "test-corr-data-quality", 1)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	dqFilter := true
	recs, total, err := repo.List(ctx, testOrg, RecommendationFilter{
		DataQualityOnly: &dqFilter,
	})
	if err != nil {
		t.Fatalf("Failed to list data quality recommendations: %v", err)
	}
	if total < 2 || len(recs) < 2 {
		t.Fatalf("Expected at least 2 data quality recommendations, got %d", total)
	}

	foundZeroTotal := false
	foundBadBalance := false
	for _, r := range recs {
		if r.InvoiceID != nil && *r.InvoiceID == invID1 {
			foundZeroTotal = true
		}
		if r.InvoiceID != nil && *r.InvoiceID == invID2 {
			foundBadBalance = true
		}
	}
	if !foundZeroTotal {
		t.Errorf("Zero total invoice recommendation not found")
	}
	if !foundBadBalance {
		t.Errorf("Impossible balance invoice recommendation not found")
	}
}

func TestCollectionDraftGenerationAndSafety(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	svc := NewService(repo, gen, nil)
	ctx := context.Background()
	testOrg := int64(88954)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM customer_invoices WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM customers WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	cRes, _ := db.ExecContext(ctx, `
		INSERT INTO customers (org_id, name, status, created_at, updated_at)
		VALUES (?, 'Oceanic Transport Corp', 'Active', NOW(), NOW())
	`, testOrg)
	custID, _ := cRes.LastInsertId()

	invRes, _ := db.ExecContext(ctx, `
		INSERT INTO customer_invoices (org_id, customer_id, customer_name, invoice_number, invoice_date, status, currency, total_amount, balance_due, due_date, created_at, updated_at)
		VALUES (?, ?, 'Oceanic Transport Corp', 'INV-DRAFT-001', NOW() - INTERVAL 20 DAY, 'Overdue', 'USD', 8800.00, 8800.00, NOW() - INTERVAL 12 DAY, NOW() - INTERVAL 20 DAY, NOW())
	`, testOrg, custID)
	invID, _ := invRes.LastInsertId()

	_, err := gen.Generate(ctx, testOrg, "test-corr-draft", 1)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	recs, _, _ := repo.List(ctx, testOrg, RecommendationFilter{InvoiceID: &invID})
	if len(recs) == 0 {
		t.Fatalf("Expected recommendation for invoice %d", invID)
	}
	recID := recs[0].ID

	// 1. Generate Draft
	drafted, err := svc.GenerateDraft(ctx, testOrg, recID, 1, "Credit Controller", "corr-draft-001")
	if err != nil {
		t.Fatalf("GenerateDraft failed: %v", err)
	}
	if drafted.DraftSubject == nil || *drafted.DraftSubject == "" {
		t.Errorf("Expected non-empty draft subject")
	}
	if drafted.DraftBody == nil || *drafted.DraftBody == "" {
		t.Errorf("Expected non-empty draft body")
	}
	if drafted.DraftStatus != "DRAFTED" {
		t.Errorf("Expected draft status DRAFTED, got %s", drafted.DraftStatus)
	}

	// Verify draft references real invoice data
	if !strings.Contains(*drafted.DraftSubject, "INV-DRAFT-001") {
		t.Errorf("Expected draft subject to mention INV-DRAFT-001, got: %s", *drafted.DraftSubject)
	}

	// 2. Save Edited Draft
	editedBody := *drafted.DraftBody + "\n\nNote: Please remit via Wire Transfer before Friday."
	saved, err := svc.SaveDraft(ctx, testOrg, recID, SaveDraftInput{
		Subject: *drafted.DraftSubject,
		Body:    editedBody,
	}, 1, "Credit Controller", "corr-draft-002")
	if err != nil {
		t.Fatalf("SaveDraft failed: %v", err)
	}
	if saved.DraftStatus != "SAVED" {
		t.Errorf("Expected draft status SAVED, got %s", saved.DraftStatus)
	}
	if !strings.Contains(*saved.DraftBody, "Wire Transfer before Friday") {
		t.Errorf("Expected saved draft body to contain edited text")
	}

	// 3. Action Preview (Read-Only Guarantee)
	preview, err := svc.GetActionPreview(ctx, testOrg, recID, 1, "Credit Controller", "corr-preview-001")
	if err != nil {
		t.Fatalf("GetActionPreview failed: %v", err)
	}
	if preview.ProposedAction == "" {
		t.Errorf("Expected ProposedAction to be populated")
	}
	if preview.ExpectedEffect == "" {
		t.Errorf("Expected ExpectedEffect to be populated")
	}

	// 4. Verify Read-Only Guarantee: Invoice balance, status, totals MUST REMAIN UNCHANGED
	var currentStatus string
	var currentBalance float64
	err = db.QueryRowContext(ctx, "SELECT status, balance_due FROM customer_invoices WHERE id = ?", invID).Scan(&currentStatus, &currentBalance)
	if err != nil {
		t.Fatalf("Failed to query invoice: %v", err)
	}
	if currentStatus != "Overdue" {
		t.Errorf("Read-only violation! Invoice status mutated to %s", currentStatus)
	}
	if currentBalance != 8800.00 {
		t.Errorf("Read-only violation! Invoice balance mutated to %f", currentBalance)
	}
}

func TestInvoiceEvidenceAndCustomerCollectionSummary(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo, nil, nil)
	ctx := context.Background()
	testOrg := int64(88955)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM customer_invoice_items WHERE invoice_id IN (SELECT id FROM customer_invoices WHERE org_id = ?)", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM customer_invoice_payments WHERE invoice_id IN (SELECT id FROM customer_invoices WHERE org_id = ?)", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM customer_invoices WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM customers WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	cRes, _ := db.ExecContext(ctx, `
		INSERT INTO customers (org_id, name, status, created_at, updated_at)
		VALUES (?, 'Transatlantic Maritime', 'Active', NOW(), NOW())
	`, testOrg)
	custID, _ := cRes.LastInsertId()

	// Insert invoice: total $10,000, paid $4,000, balance $6,000, due 20 days ago
	due20DaysAgo := time.Now().Add(-20 * 24 * time.Hour)
	invRes, _ := db.ExecContext(ctx, `
		INSERT INTO customer_invoices (org_id, customer_id, customer_name, invoice_number, invoice_date, status, currency, total_amount, paid_amount, balance_due, due_date, created_at, updated_at)
		VALUES (?, ?, 'Transatlantic Maritime', 'INV-EVID-001', NOW() - INTERVAL 25 DAY, 'Overdue', 'USD', 10000.00, 4000.00, 6000.00, ?, NOW() - INTERVAL 25 DAY, NOW())
	`, testOrg, custID, due20DaysAgo)
	invID, _ := invRes.LastInsertId()

	// Insert payment record of $4,000
	_, _ = db.ExecContext(ctx, `
		INSERT INTO customer_invoice_payments (org_id, invoice_id, amount, payment_date, payment_method, payment_ref, notes, created_at)
		VALUES (?, ?, 4000.00, NOW() - INTERVAL 5 DAY, 'WIRE', 'WIRE-TX-9901', 'Partial remittance', NOW())
	`, testOrg, invID)

	// 1. Test GetInvoiceEvidence
	evidencePayload, err := svc.GetInvoiceEvidence(ctx, testOrg, invID)
	if err != nil {
		t.Fatalf("GetInvoiceEvidence failed: %v", err)
	}
	if evidencePayload.InvoiceID != invID {
		t.Errorf("Expected InvoiceID %d, got %d", invID, evidencePayload.InvoiceID)
	}
	if evidencePayload.AgingBand != "16-30_DAYS" {
		t.Errorf("Expected AgingBand 16-30_DAYS, got %s", evidencePayload.AgingBand)
	}
	if evidencePayload.PaidAmount != 4000.00 {
		t.Errorf("Expected PaidAmount 4000.00, got %f", evidencePayload.PaidAmount)
	}
	if evidencePayload.BalanceDue != 6000.00 {
		t.Errorf("Expected BalanceDue 6000.00, got %f", evidencePayload.BalanceDue)
	}
	if len(evidencePayload.Payments) != 1 {
		t.Errorf("Expected 1 payment record, got %d", len(evidencePayload.Payments))
	}

	// 2. Test GetCustomerCollectionSummary
	summary, err := svc.GetCustomerCollectionSummary(ctx, testOrg, custID)
	if err != nil {
		t.Fatalf("GetCustomerCollectionSummary failed: %v", err)
	}
	if summary.CustomerID != custID {
		t.Errorf("Expected CustomerID %d, got %d", custID, summary.CustomerID)
	}
	if summary.TotalOutstanding != 6000.00 {
		t.Errorf("Expected TotalOutstanding 6000.00, got %f", summary.TotalOutstanding)
	}
	if summary.OverdueInvoicesCount != 1 {
		t.Errorf("Expected OverdueInvoicesCount 1, got %d", summary.OverdueInvoicesCount)
	}
	if len(summary.Invoices) != 1 {
		t.Errorf("Expected 1 open invoice, got %d", len(summary.Invoices))
	}
}

func TestTenantIsolationForInvoiceCollections(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	svc := NewService(repo, gen, nil)
	ctx := context.Background()
	orgA := int64(88956)
	orgB := int64(88957)

	ensureTestOrg(ctx, db, orgA)
	ensureTestOrg(ctx, db, orgB)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id IN (?, ?)", orgA, orgB)
		_, _ = db.ExecContext(ctx, "DELETE FROM customer_invoices WHERE org_id IN (?, ?)", orgA, orgB)
		_, _ = db.ExecContext(ctx, "DELETE FROM customers WHERE org_id IN (?, ?)", orgA, orgB)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id IN (?, ?)", orgA, orgB)
	}()

	// Org A customer & invoice
	cA, _ := db.ExecContext(ctx, "INSERT INTO customers (org_id, name, status, created_at, updated_at) VALUES (?, 'Org A Debtor', 'Active', NOW(), NOW())", orgA)
	custA, _ := cA.LastInsertId()
	iA, _ := db.ExecContext(ctx, "INSERT INTO customer_invoices (org_id, customer_id, customer_name, invoice_number, invoice_date, status, currency, total_amount, balance_due, due_date, created_at, updated_at) VALUES (?, ?, 'Org A Debtor', 'INV-ORGA-01', NOW() - INTERVAL 10 DAY, 'Overdue', 'USD', 5000.00, 5000.00, NOW() - INTERVAL 10 DAY, NOW(), NOW())", orgA, custA)
	invA, _ := iA.LastInsertId()

	// Org B generator
	_, _ = gen.Generate(ctx, orgA, "corr-iso-a", 1)

	// Org B tries to query Org A's recommendations
	recsB, totalB, _ := repo.List(ctx, orgB, RecommendationFilter{InvoiceID: &invA})
	if totalB != 0 || len(recsB) != 0 {
		t.Fatalf("Tenant isolation failure! Org B was able to see Org A recommendations")
	}

	// Org B tries to get Org A's invoice evidence
	_, err := svc.GetInvoiceEvidence(ctx, orgB, invA)
	if err == nil {
		t.Fatalf("Tenant isolation failure! Org B was able to get evidence for Org A invoice")
	}

	// Org B tries to get Org A's customer collection summary
	_, err = svc.GetCustomerCollectionSummary(ctx, orgB, custA)
	if err == nil {
		t.Fatalf("Tenant isolation failure! Org B was able to get collection summary for Org A customer")
	}
}
