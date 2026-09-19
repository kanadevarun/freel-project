package recommendations

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/freel/backend/internal/approvals"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func ensureTestOrg(ctx context.Context, db *sqlx.DB, orgID int64) {
	_, _ = db.ExecContext(ctx, "INSERT IGNORE INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Test Organization', NOW(), NOW())", orgID)
}

func TestRFQSignalMissingInfoDetection(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	ctx := context.Background()
	testOrg := int64(88881)

	ensureTestOrg(ctx, db, testOrg)

	// Clean up test org records
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM rfq_items WHERE rfq_id IN (SELECT id FROM rfqs WHERE org_id = ?)", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM rfqs WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	// 1. Insert Incomplete RFQ: Missing destination and incoterms
	res1, err := db.ExecContext(ctx, `
		INSERT INTO rfqs (org_id, rfq_number, customer_id, origin, destination, incoterms, target_date, status, stage, created_at, updated_at)
		VALUES (?, 'RFQ-TEST-INCOMPLETE-1', 101, 'INNSA', NULL, NULL, NULL, 'SUBMITTED', 'DRAFT', NOW(), NOW())
	`, testOrg)
	if err != nil {
		t.Fatalf("Failed to insert test RFQ 1: %v", err)
	}
	rfq1ID, _ := res1.LastInsertId()

	// 2. Insert Complete RFQ: Has origin, destination, incoterms, target_date, and valid cargo items
	targetDate := time.Now().AddDate(0, 0, 30)
	res2, err := db.ExecContext(ctx, `
		INSERT INTO rfqs (org_id, rfq_number, customer_id, origin, destination, incoterms, target_date, status, stage, created_at, updated_at)
		VALUES (?, 'RFQ-TEST-COMPLETE-2', 101, 'INNSA', 'DEHAM', 'FOB', ?, 'SUBMITTED', 'DRAFT', NOW(), NOW())
	`, testOrg, targetDate)
	if err != nil {
		t.Fatalf("Failed to insert test RFQ 2: %v", err)
	}
	rfq2ID, _ := res2.LastInsertId()

	// Add cargo items for complete RFQ
	_, err = db.ExecContext(ctx, `
		INSERT INTO rfq_items (rfq_id, description, quantity, weight_kg, volume_cbm, created_at, updated_at)
		VALUES (?, 'Industrial Machinery Parts', 10, 2500.00, 12.50, NOW(), NOW())
	`, rfq2ID)
	if err != nil {
		t.Fatalf("Failed to insert test rfq_items: %v", err)
	}

	// 3. Run Generator
	result, err := gen.Generate(ctx, testOrg, "test-corr-rfq-missing", 1)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if result == nil {
		t.Fatalf("Expected non-nil generate result")
	}

	// 4. Verify candidate detection
	recs, total, err := repo.List(ctx, testOrg, RecommendationFilter{
		Category: CategoryRFQ,
	})
	if err != nil {
		t.Fatalf("Failed to list recommendations: %v", err)
	}
	if total == 0 {
		t.Fatalf("Expected at least 1 RFQ recommendation for incomplete RFQ")
	}

	var foundIncomplete bool
	var foundCompleteForMissing bool

	for _, rec := range recs {
		if rec.SourceID == rfq1ID && rec.RuleApplied == "RFQ_MISSING_INFO" {
			foundIncomplete = true
			if rec.ActionType != ActionTypeRequestRFQClarification {
				t.Errorf("Expected action type %s, got %s", ActionTypeRequestRFQClarification, rec.ActionType)
			}
			if rec.FollowupType == nil || *rec.FollowupType != DraftTypeRFQClarification {
				t.Errorf("Expected followup type %s", DraftTypeRFQClarification)
			}

			// Verify evidence items contain missing destination and incoterms
			var foundDest, foundInco bool
			for _, ev := range rec.Evidence {
				if ev.FieldName == "destination" {
					foundDest = true
				}
				if ev.FieldName == "incoterms" {
					foundInco = true
				}
			}
			if !foundDest || !foundInco {
				t.Errorf("Expected evidence to identify destination and incoterms as missing: %+v", rec.Evidence)
			}
		}

		if rec.SourceID == rfq2ID && rec.RuleApplied == "RFQ_MISSING_INFO" {
			foundCompleteForMissing = true
		}
	}

	if !foundIncomplete {
		t.Errorf("Expected RFQ 1 to be flagged for missing information")
	}
	if foundCompleteForMissing {
		t.Errorf("Complete RFQ 2 should NOT have been flagged for missing information")
	}
}

func TestQuotationMarginAndPricingWarnings(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	ctx := context.Background()
	testOrg := int64(88882)

	ensureTestOrg(ctx, db, testOrg)

	// Clean up test org records
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM quotations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	// 1. Insert Low Margin Quotation: 4.5% margin (below 10.0% threshold)
	res1, err := db.ExecContext(ctx, `
		INSERT INTO quotations (org_id, quotation_number, status, total_amount, total_cost, gross_profit, gross_margin_pct, valid_until, created_at, updated_at)
		VALUES (?, 'QT-TEST-LOW-MARGIN', 'DRAFT', 2000.00, 1910.00, 90.00, 4.5000, DATE_ADD(NOW(), INTERVAL 20 DAY), NOW(), NOW())
	`, testOrg)
	if err != nil {
		t.Fatalf("Failed to insert test quote 1: %v", err)
	}
	q1ID, _ := res1.LastInsertId()

	// 2. Insert Missing Cost Quotation: total_amount = 3000.00, total_cost = 0.00
	res2, err := db.ExecContext(ctx, `
		INSERT INTO quotations (org_id, quotation_number, status, total_amount, total_cost, gross_profit, gross_margin_pct, valid_until, created_at, updated_at)
		VALUES (?, 'QT-TEST-MISSING-COST', 'PENDING_APPROVAL', 3000.00, 0.00, 3000.00, 100.0000, DATE_ADD(NOW(), INTERVAL 20 DAY), NOW(), NOW())
	`, testOrg)
	if err != nil {
		t.Fatalf("Failed to insert test quote 2: %v", err)
	}
	q2ID, _ := res2.LastInsertId()

	// 3. Insert Healthy Quotation: 25.0% margin
	_, err = db.ExecContext(ctx, `
		INSERT INTO quotations (org_id, quotation_number, status, total_amount, total_cost, gross_profit, gross_margin_pct, valid_until, created_at, updated_at)
		VALUES (?, 'QT-TEST-HEALTHY', 'APPROVED', 4000.00, 3000.00, 1000.00, 25.0000, DATE_ADD(NOW(), INTERVAL 20 DAY), NOW(), NOW())
	`, testOrg)
	if err != nil {
		t.Fatalf("Failed to insert test quote 3: %v", err)
	}

	// 4. Generate recommendations
	_, err = gen.Generate(ctx, testOrg, "test-corr-quote-pricing", 1)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// 5. Verify pricing recommendations
	recs, _, err := repo.List(ctx, testOrg, RecommendationFilter{
		Category: CategoryPricing,
	})
	if err != nil {
		t.Fatalf("Failed to list pricing recommendations: %v", err)
	}

	var foundLowMargin, foundMissingCost bool
	for _, rec := range recs {
		if rec.SourceID == q1ID && rec.RuleApplied == "QUOTATION_MARGIN_ALERT" {
			foundLowMargin = true
			if !rec.RequiresApproval {
				t.Errorf("Expected low margin alert to require approval")
			}
			if rec.RiskLevel != RiskHigh && rec.RiskLevel != RiskCritical {
				t.Errorf("Expected high/critical risk level, got %s", rec.RiskLevel)
			}
		}
		if rec.SourceID == q2ID && rec.RuleApplied == "QUOTATION_MARGIN_ALERT" {
			foundMissingCost = true
			if rec.Priority != PriorityCritical {
				t.Errorf("Expected critical priority for missing cost components")
			}
		}
	}

	if !foundLowMargin {
		t.Errorf("Expected quote 1 to trigger low margin pricing warning")
	}
	if !foundMissingCost {
		t.Errorf("Expected quote 2 to trigger missing cost pricing warning")
	}

	// 6. Verify Read-Only Safety: Quotation values in DB must NOT have changed
	var checkQ struct {
		TotalAmount    float64 `db:"total_amount"`
		TotalCost      float64 `db:"total_cost"`
		GrossMarginPct float64 `db:"gross_margin_pct"`
	}
	_ = db.GetContext(ctx, &checkQ, "SELECT total_amount, total_cost, gross_margin_pct FROM quotations WHERE id = ?", q1ID)
	if checkQ.TotalAmount != 2000.00 || checkQ.TotalCost != 1910.00 || checkQ.GrossMarginPct != 4.5 {
		t.Errorf("Quotation values were mutated! Expected 2000, 1910, 4.5, got: %+v", checkQ)
	}
}

func TestControlledActionPreview(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	svc := NewService(repo, gen, nil)
	ctx := context.Background()
	testOrg := int64(88883)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM rfqs WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	res, err := db.ExecContext(ctx, `
		INSERT INTO rfqs (org_id, rfq_number, customer_id, origin, destination, incoterms, status, stage, created_at, updated_at)
		VALUES (?, 'RFQ-TEST-PREVIEW', 105, 'INNSA', NULL, NULL, 'SUBMITTED', 'DRAFT', NOW(), NOW())
	`, testOrg)
	if err != nil {
		t.Fatalf("Failed to insert RFQ: %v", err)
	}
	rfqID, _ := res.LastInsertId()

	_, _ = gen.Generate(ctx, testOrg, "test-preview-corr", 1)

	recs, _, err := repo.List(ctx, testOrg, RecommendationFilter{
		RFQID: &rfqID,
	})
	if err != nil || len(recs) == 0 {
		t.Fatalf("Expected recommendation for test RFQ, got: %v", err)
	}
	recID := recs[0].ID

	// Call GetActionPreview
	preview, err := svc.GetActionPreview(ctx, testOrg, recID, 12, "Jane Operator", "prev-corr-001")
	if err != nil {
		t.Fatalf("GetActionPreview failed: %v", err)
	}

	if preview.RecommendationID != recID {
		t.Errorf("Expected RecommendationID %d, got %d", recID, preview.RecommendationID)
	}
	if preview.SourceType != "RFQ" || preview.SourceID != rfqID {
		t.Errorf("Expected SourceType RFQ #%d, got %s #%d", rfqID, preview.SourceType, preview.SourceID)
	}
	if preview.ProposedAction == "" || preview.ExpectedEffect == "" {
		t.Errorf("Expected non-empty ProposedAction and ExpectedEffect")
	}
	if preview.InitiatedByUser != "Jane Operator" {
		t.Errorf("Expected InitiatedByUser Jane Operator, got %s", preview.InitiatedByUser)
	}
	if len(preview.Evidence) == 0 {
		t.Errorf("Expected verified factual evidence in preview")
	}
}

func TestApprovalRequestIntegration(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	approvalsRepo := approvals.NewRepository(db)
	approvalsSvc := approvals.NewService(approvalsRepo)
	svc := NewService(repo, gen, nil)
	svc.SetApprovalsService(approvalsSvc)

	ctx := context.Background()
	testOrg := int64(88884)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM approval_requests WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM quotations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	// Insert high-risk quote pending approval
	res, err := db.ExecContext(ctx, `
		INSERT INTO quotations (org_id, quotation_number, status, total_amount, total_cost, gross_profit, gross_margin_pct, valid_until, created_at, updated_at)
		VALUES (?, 'QT-TEST-HITL-APP', 'PENDING_APPROVAL', 5000.00, 3600.00, 1400.00, 28.0000, DATE_ADD(NOW(), INTERVAL 10 DAY), NOW(), NOW())
	`, testOrg)
	if err != nil {
		t.Fatalf("Failed to insert quote: %v", err)
	}
	qID, _ := res.LastInsertId()

	_, _ = gen.Generate(ctx, testOrg, "test-corr-hitl", 1)

	recs, _, err := repo.List(ctx, testOrg, RecommendationFilter{
		QuotationID: &qID,
	})
	if err != nil || len(recs) == 0 {
		t.Fatalf("Expected recommendation for quote %d", qID)
	}
	recID := recs[0].ID

	// Submit for approval
	updatedRec, err := svc.RequestApproval(ctx, testOrg, recID, RequestApprovalInput{
		Notes: "Margin verified at 28%. Requesting signoff from pricing lead.",
	}, 15, "Pricing Analyst", "app-corr-101")
	if err != nil {
		t.Fatalf("RequestApproval failed: %v", err)
	}

	if updatedRec.ApprovalID == nil || *updatedRec.ApprovalID <= 0 {
		t.Fatalf("Expected recommendation to have non-nil ApprovalID after approval request")
	}

	// Verify approval request was created in approval_requests table
	var appReq struct {
		ID       int64  `db:"id"`
		Status   string `db:"status"`
		Title    string `db:"title"`
		Category string `db:"category"`
	}
	err = db.GetContext(ctx, &appReq, "SELECT id, status, title, category FROM approval_requests WHERE id = ? AND org_id = ?", *updatedRec.ApprovalID, testOrg)
	if err != nil {
		t.Fatalf("Approval request record not found in DB: %v", err)
	}
	if appReq.Status != "Pending" {
		t.Errorf("Expected approval request status Pending, got %s", appReq.Status)
	}
}

func TestDraftGenerationTemplatesSafety(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	svc := NewService(repo, gen, nil)
	ctx := context.Background()
	testOrg := int64(88885)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
	}()

	// 1. Manually create candidate for RFQ Clarification
	ev := []EvidenceItem{
		{SourceModule: "rfq", FieldName: "destination", ObservedValue: "Missing", Description: "Delivery port undefined"},
		{SourceModule: "rfq", FieldName: "incoterms", ObservedValue: "Missing", Description: "Trade incoterm not provided"},
	}
	evBytes, _ := json.Marshal(ev)
	cName := "Nordic Timber AB"
	ftRFQ := DraftTypeRFQClarification

	rec1, err := repo.Create(ctx, &Recommendation{
		OrgID:            testOrg,
		SourceType:       SourceRFQ,
		SourceID:         701,
		SourceReference:  "RFQ-TEST-DRAFT-1",
		Title:            "RFQ Missing Info",
		Description:      "Inquiry lacks destination and incoterms.",
		Category:         CategoryRFQ,
		Priority:         PriorityHigh,
		RiskLevel:        RiskMedium,
		Confidence:       "HIGH",
		ConfidenceScore:  0.95,
		EvidenceJSON:     string(evBytes),
		RecommendedAction: "Request clarification from Nordic Timber AB.",
		ActionType:       ActionTypeRequestRFQClarification,
		Status:           StatusNew,
		CustomerName:     &cName,
		FollowupType:     &ftRFQ,
		DedupHash:        computeDedupHash(testOrg, SourceRFQ, 701, CategoryRFQ, "TEST_DRAFT_RFQ"),
		DraftStatus:      DraftStatusNotGenerated,
	})
	if err != nil {
		t.Fatalf("Failed to create rec1: %v", err)
	}

	// Generate draft for RFQ
	drafted1, err := svc.GenerateDraft(ctx, testOrg, rec1.ID, 10, "Sales Rep", "draft-corr-1")
	if err != nil {
		t.Fatalf("GenerateDraft failed: %v", err)
	}
	if drafted1.DraftSubject == nil || drafted1.DraftBody == nil {
		t.Fatalf("Expected draft subject and body to be populated")
	}

	// Verify customer draft includes missing fields
	if !strings.Contains(*drafted1.DraftBody, "destination") || !strings.Contains(*drafted1.DraftBody, "incoterms") {
		t.Errorf("Expected draft body to list missing destination and incoterms: %s", *drafted1.DraftBody)
	}

	// 2. Test Internal Pricing Review Note
	ftPricing := DraftTypeInternalPricingReview
	rec2, err := repo.Create(ctx, &Recommendation{
		OrgID:            testOrg,
		SourceType:       SourceQuotation,
		SourceID:         801,
		SourceReference:  "QT-TEST-DRAFT-2",
		Title:            "Quotation Low Margin Alert",
		Description:      "Quotation has low margin of 4.2% with total cost $19,100.00.",
		Category:         CategoryPricing,
		Priority:         PriorityCritical,
		RiskLevel:        RiskCritical,
		Confidence:       "HIGH",
		ConfidenceScore:  0.99,
		EvidenceJSON:     string(evBytes),
		RecommendedAction: "Verify carrier buy rates and required markup.",
		ActionType:       ActionTypeRequestPricingApproval,
		Status:           StatusNew,
		CustomerName:     &cName,
		FollowupType:     &ftPricing,
		DedupHash:        computeDedupHash(testOrg, SourceQuotation, 801, CategoryPricing, "TEST_DRAFT_PRICING"),
		DraftStatus:      DraftStatusNotGenerated,
	})
	if err != nil {
		t.Fatalf("Failed to create rec2: %v", err)
	}

	drafted2, err := svc.GenerateDraft(ctx, testOrg, rec2.ID, 10, "Pricing Analyst", "draft-corr-2")
	if err != nil {
		t.Fatalf("GenerateDraft failed for pricing: %v", err)
	}
	if !strings.Contains(*drafted2.DraftSubject, "Internal Note") {
		t.Errorf("Expected internal subject line for internal pricing review note: %s", *drafted2.DraftSubject)
	}

	// 3. Test Quotation Follow-Up does not leak internal costs to customer
	ftQuote := DraftTypeQuotationFollowup
	rec3, err := repo.Create(ctx, &Recommendation{
		OrgID:            testOrg,
		SourceType:       SourceQuotation,
		SourceID:         901,
		SourceReference:  "QT-TEST-DRAFT-3",
		Title:            "Quotation Approaching Expiry",
		Description:      "Quotation valid until 2026-10-31.",
		Category:         CategoryQuotation,
		Priority:         PriorityMedium,
		RiskLevel:        RiskMedium,
		Confidence:       "HIGH",
		ConfidenceScore:  0.90,
		EvidenceJSON:     string(evBytes),
		RecommendedAction: "Follow up with customer buyer.",
		ActionType:       ActionTypeFollowupQuote,
		Status:           StatusNew,
		CustomerName:     &cName,
		FollowupType:     &ftQuote,
		DedupHash:        computeDedupHash(testOrg, SourceQuotation, 901, CategoryQuotation, "TEST_DRAFT_QUOTE"),
		DraftStatus:      DraftStatusNotGenerated,
	})
	if err != nil {
		t.Fatalf("Failed to create rec3: %v", err)
	}

	drafted3, err := svc.GenerateDraft(ctx, testOrg, rec3.ID, 10, "Sales Rep", "draft-corr-3")
	if err != nil {
		t.Fatalf("GenerateDraft failed: %v", err)
	}
	// Verify customer draft avoids leaking confidential margin or internal cost
	if strings.Contains(strings.ToLower(*drafted3.DraftBody), "margin") || strings.Contains(strings.ToLower(*drafted3.DraftBody), "total cost") {
		t.Errorf("Customer draft must NOT leak internal margin or total cost: %s", *drafted3.DraftBody)
	}
}

func TestTenantIsolationForRFQAndQuotations(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	svc := NewService(repo, gen, nil)
	ctx := context.Background()

	orgA := int64(88886)
	orgB := int64(88887)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id IN (?, ?)", orgA, orgB)
	}()

	// Create recommendation for Org A
	recA, err := repo.Create(ctx, &Recommendation{
		OrgID:            orgA,
		SourceType:       SourceRFQ,
		SourceID:         991,
		SourceReference:  "RFQ-ORGA-991",
		Title:            "Org A RFQ Recommendation",
		Description:      "Confidential pricing for Org A.",
		Category:         CategoryRFQ,
		Priority:         PriorityHigh,
		RiskLevel:        RiskMedium,
		Confidence:       "HIGH",
		ConfidenceScore:  0.90,
		EvidenceJSON:     "[]",
		RecommendedAction: "Review Org A inquiry.",
		ActionType:       ActionTypeRequestRFQClarification,
		Status:           StatusNew,
		DedupHash:        computeDedupHash(orgA, SourceRFQ, 991, CategoryRFQ, "TENANT_TEST"),
		DraftStatus:      DraftStatusNotGenerated,
	})
	if err != nil {
		t.Fatalf("Failed to create Org A recommendation: %v", err)
	}

	// 1. Org B cannot get Org A recommendation
	rec, err := svc.GetRecommendation(ctx, orgB, recA.ID)
	if err == nil && rec != nil {
		t.Errorf("Cross-tenant leakage! Org B was able to fetch Org A recommendation %d", recA.ID)
	}

	// 2. Org B cannot get Action Preview for Org A recommendation
	preview, err := svc.GetActionPreview(ctx, orgB, recA.ID, 99, "Attacker", "corr-hack")
	if err == nil && preview != nil {
		t.Errorf("Cross-tenant leakage! Org B was able to preview Org A recommendation %d", recA.ID)
	}

	// 3. Org B cannot generate draft for Org A recommendation
	draft, err := svc.GenerateDraft(ctx, orgB, recA.ID, 99, "Attacker", "corr-hack")
	if err == nil && draft != nil {
		t.Errorf("Cross-tenant leakage! Org B was able to generate draft for Org A recommendation %d", recA.ID)
	}
}
