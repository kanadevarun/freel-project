package recommendations

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func TestContractExpiryDetection(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	ctx := context.Background()
	testOrg := int64(88893)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM contracts WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	// 1. Insert expired contract
	pastDate := time.Now().Add(-15 * 24 * time.Hour)
	res1, err := db.ExecContext(ctx, `
		INSERT INTO contracts (org_id, contract_reference, contract_name, contract_type, party_id, party_name, transport_mode, status, currency, contract_value, effective_date, expiry_date, owner, created_at, updated_at)
		VALUES (?, 'CNT-TEST-EXP-001', 'Expired Ocean Agreement', 'Carrier Agreement', 1, 'Maersk Test Line', 'OCEAN', 'ACTIVE', 'USD', 75000.00, NOW() - INTERVAL 365 DAY, ?, 'Sarah Jenkins', NOW(), NOW())
	`, testOrg, pastDate)
	if err != nil {
		t.Fatalf("Failed to insert expired contract: %v", err)
	}
	expiredID, _ := res1.LastInsertId()

	// 2. Insert soon-to-expire contract (12 days)
	soonDate := time.Now().Add(12 * 24 * time.Hour)
	res2, err := db.ExecContext(ctx, `
		INSERT INTO contracts (org_id, contract_reference, contract_name, contract_type, party_id, party_name, transport_mode, status, currency, contract_value, effective_date, expiry_date, owner, created_at, updated_at)
		VALUES (?, 'CNT-TEST-SOON-002', 'Expiring Airfreight Agreement', 'Carrier Agreement', 2, 'Cargolux Test', 'AIR', 'ACTIVE', 'USD', 120000.00, NOW() - INTERVAL 180 DAY, ?, 'Marcus Vance', NOW(), NOW())
	`, testOrg, soonDate)
	if err != nil {
		t.Fatalf("Failed to insert expiring contract: %v", err)
	}
	soonID, _ := res2.LastInsertId()

	// 3. Run generator
	result, err := gen.Generate(ctx, testOrg, "test-corr-contract-expiry", 1)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if result == nil {
		t.Fatalf("Expected non-nil generate result")
	}

	// 4. Verify expired recommendation
	expiredRecs, _, err := repo.List(ctx, testOrg, RecommendationFilter{
		ContractID: &expiredID,
	})
	if err != nil {
		t.Fatalf("Failed to list recommendations for expired contract: %v", err)
	}
	if len(expiredRecs) == 0 {
		t.Fatalf("Expected recommendation for expired contract, got 0")
	}
	expRec := expiredRecs[0]
	if expRec.RuleApplied != "CONTRACT_EXPIRED" {
		t.Errorf("Expected RuleApplied CONTRACT_EXPIRED, got %s", expRec.RuleApplied)
	}
	if expRec.Priority != PriorityCritical {
		t.Errorf("Expected Priority critical, got %s", expRec.Priority)
	}
	if expRec.Category != CategoryContract {
		t.Errorf("Expected Category contract, got %s", expRec.Category)
	}
	if expRec.ContractID == nil || *expRec.ContractID != expiredID {
		t.Errorf("Expected ContractID %d, got %v", expiredID, expRec.ContractID)
	}
	if expRec.ActionType != ActionTypeReviewExpiringContract {
		t.Errorf("Expected ActionType %s, got %s", ActionTypeReviewExpiringContract, expRec.ActionType)
	}

	// 5. Verify expiring soon recommendation
	soonRecs, _, err := repo.List(ctx, testOrg, RecommendationFilter{
		ContractID: &soonID,
	})
	if err != nil {
		t.Fatalf("Failed to list recommendations for soon-to-expire contract: %v", err)
	}
	if len(soonRecs) == 0 {
		t.Fatalf("Expected recommendation for expiring contract, got 0")
	}
	soonRec := soonRecs[0]
	if soonRec.RuleApplied != "CONTRACT_EXPIRING_SOON" {
		t.Errorf("Expected RuleApplied CONTRACT_EXPIRING_SOON, got %s", soonRec.RuleApplied)
	}
	if soonRec.ActionType != ActionTypePrepareRenewalReminder {
		t.Errorf("Expected ActionType %s, got %s", ActionTypePrepareRenewalReminder, soonRec.ActionType)
	}
}

func TestContractMissingOwnerDetection(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	ctx := context.Background()
	testOrg := int64(88894)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM contracts WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	// Insert contract with no owner and far future expiry
	futureDate := time.Now().Add(200 * 24 * time.Hour)
	res, err := db.ExecContext(ctx, `
		INSERT INTO contracts (org_id, contract_reference, contract_name, contract_type, party_id, party_name, transport_mode, status, currency, contract_value, effective_date, expiry_date, owner, created_at, updated_at)
		VALUES (?, 'CNT-TEST-UNOWNED-003', 'Unassigned Trucking Agreement', 'Vendor Agreement', 3, 'FastDray Logistics', 'ROAD', 'ACTIVE', 'USD', 45000.00, NOW(), ?, NULL, NOW(), NOW())
	`, testOrg, futureDate)
	if err != nil {
		t.Fatalf("Failed to insert unowned contract: %v", err)
	}
	contractID, _ := res.LastInsertId()

	_, err = gen.Generate(ctx, testOrg, "test-corr-unowned-contract", 1)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	recs, _, err := repo.List(ctx, testOrg, RecommendationFilter{
		ContractID: &contractID,
	})
	if err != nil {
		t.Fatalf("Failed to list recommendations: %v", err)
	}
	if len(recs) == 0 {
		t.Fatalf("Expected recommendation for unowned contract, got 0")
	}

	rec := recs[0]
	if rec.RuleApplied != "CONTRACT_MISSING_OWNER" {
		t.Errorf("Expected RuleApplied CONTRACT_MISSING_OWNER, got %s", rec.RuleApplied)
	}
	if rec.ActionType != ActionTypeAssignContractOwner {
		t.Errorf("Expected ActionType %s, got %s", ActionTypeAssignContractOwner, rec.ActionType)
	}
}

func TestDocumentExpiryAndMissingDetection(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	ctx := context.Background()
	testOrg := int64(88895)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM shipment_documents WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM shipments WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	// 1. Insert shipment
	resShip, err := db.ExecContext(ctx, `
		INSERT INTO shipments (org_id, booking_number, carrier_scac, origin_port, destination_port, status, created_at, updated_at)
		VALUES (?, 'BK-TEST-DOCS-001', 'MSCU', 'CNSHA', 'USLAX', 'BOOKED', NOW(), NOW())
	`, testOrg)
	if err != nil {
		t.Fatalf("Failed to insert shipment: %v", err)
	}
	shipmentID, _ := resShip.LastInsertId()

	// 2. Insert expired document
	pastExp := time.Now().Add(-10 * 24 * time.Hour)
	resDoc1, err := db.ExecContext(ctx, `
		INSERT INTO shipment_documents (org_id, shipment_id, doc_type, file_name, status, document_date, expires_at, created_at, updated_at)
		VALUES (?, ?, 'EXPORT_LICENSE', 'export_license_exp.pdf', 'PENDING', NOW() - INTERVAL 90 DAY, ?, NOW(), NOW())
	`, testOrg, shipmentID, pastExp)
	if err != nil {
		t.Fatalf("Failed to insert expired doc: %v", err)
	}
	doc1ID, _ := resDoc1.LastInsertId()

	// 3. Insert missing required document
	resDoc2, err := db.ExecContext(ctx, `
		INSERT INTO shipment_documents (org_id, shipment_id, doc_type, file_name, status, created_at, updated_at)
		VALUES (?, ?, 'CERTIFICATE_OF_ORIGIN', 'coo_pending.pdf', 'MISSING', NOW(), NOW())
	`, testOrg, shipmentID)
	if err != nil {
		t.Fatalf("Failed to insert missing doc: %v", err)
	}
	doc2ID, _ := resDoc2.LastInsertId()

	// 4. Run generator
	_, err = gen.Generate(ctx, testOrg, "test-corr-docs", 1)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// 5. Verify expired document recommendation
	doc1IDStr := fmt.Sprintf("%d", doc1ID)
	recs1, _, err := repo.List(ctx, testOrg, RecommendationFilter{
		DocumentID: &doc1IDStr,
	})
	if err != nil {
		t.Fatalf("Failed to query expired doc rec: %v", err)
	}
	if len(recs1) == 0 {
		t.Fatalf("Expected recommendation for expired document, got 0")
	}
	if recs1[0].RuleApplied != "DOCUMENT_EXPIRED" {
		t.Errorf("Expected DOCUMENT_EXPIRED, got %s", recs1[0].RuleApplied)
	}
	if recs1[0].Category != CategoryCompliance {
		t.Errorf("Expected Category compliance, got %s", recs1[0].Category)
	}

	// 6. Verify missing document recommendation
	doc2IDStr := fmt.Sprintf("%d", doc2ID)
	recs2, _, err := repo.List(ctx, testOrg, RecommendationFilter{
		DocumentID: &doc2IDStr,
	})
	if err != nil {
		t.Fatalf("Failed to query missing doc rec: %v", err)
	}
	if len(recs2) == 0 {
		t.Fatalf("Expected recommendation for missing document, got 0")
	}
	if recs2[0].RuleApplied != "DOCUMENT_MISSING_REQUIRED" {
		t.Errorf("Expected DOCUMENT_MISSING_REQUIRED, got %s", recs2[0].RuleApplied)
	}
}

func TestComplianceRequirementAndObligationDetection(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	ctx := context.Background()
	testOrg := int64(88896)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM contract_obligations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM contract_compliance_requirements WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM contracts WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	// 1. Insert contract
	resContract, err := db.ExecContext(ctx, `
		INSERT INTO contracts (org_id, contract_reference, contract_name, contract_type, party_id, party_name, transport_mode, status, currency, contract_value, effective_date, expiry_date, owner, created_at, updated_at)
		VALUES (?, 'CNT-TEST-COMP-004', 'Carrier Master SLA', 'Carrier Agreement', 4, 'Hapag-Lloyd Test', 'OCEAN', 'ACTIVE', 'USD', 200000.00, NOW(), NOW() + INTERVAL 180 DAY, 'Elena Rostova', NOW(), NOW())
	`, testOrg)
	if err != nil {
		t.Fatalf("Failed to insert contract: %v", err)
	}
	contractID, _ := resContract.LastInsertId()

	// 2. Insert pending compliance requirement
	resComp, err := db.ExecContext(ctx, `
		INSERT INTO contract_compliance_requirements (org_id, contract_id, requirement_type, title, description, responsible_party, valid_until, status, risk_severity, created_at, updated_at)
		VALUES (?, ?, 'INSURANCE', 'Annual Marine Cargo Liability Certificate', 'Certificate of insurance with minimum $5M cover', 'CARRIER', NOW() + INTERVAL 45 DAY, 'PENDING', 'HIGH', NOW(), NOW())
	`, testOrg, contractID)
	if err != nil {
		t.Fatalf("Failed to insert compliance requirement: %v", err)
	}
	compID, _ := resComp.LastInsertId()

	// 3. Insert overdue contract obligation
	pastDue := time.Now().Add(-5 * 24 * time.Hour)
	_, err = db.ExecContext(ctx, `
		INSERT INTO contract_obligations (org_id, contract_id, obligation_reference, title, description, obligation_type, category, responsible_party, priority, status, due_date, created_at, updated_at)
		VALUES (?, ?, 'OBL-TEST-001', 'Quarterly Carbon Emission Audit Report', 'Submit verified IMO emission data', 'REPORTING', 'GENERAL', 'CARRIER', 'HIGH', 'ACTIVE', ?, NOW(), NOW())
	`, testOrg, contractID, pastDue)
	if err != nil {
		t.Fatalf("Failed to insert obligation: %v", err)
	}

	// 4. Run generator
	_, err = gen.Generate(ctx, testOrg, "test-corr-compliance", 1)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// 5. Check compliance recommendation
	compRecs, _, err := repo.List(ctx, testOrg, RecommendationFilter{
		ComplianceID: &compID,
	})
	if err != nil {
		t.Fatalf("Failed to query compliance recs: %v", err)
	}
	if len(compRecs) == 0 {
		t.Fatalf("Expected recommendation for pending compliance, got 0")
	}
	if compRecs[0].RuleApplied != "COMPLIANCE_REQUIREMENT_PENDING" {
		t.Errorf("Expected COMPLIANCE_REQUIREMENT_PENDING, got %s", compRecs[0].RuleApplied)
	}
	if compRecs[0].ActionType != ActionTypeReviewComplianceChecklist {
		t.Errorf("Expected ActionType %s, got %s", ActionTypeReviewComplianceChecklist, compRecs[0].ActionType)
	}
}

func TestContractComplianceEvidenceRetrieval(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	testOrg := int64(88897)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM contracts WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	res, err := db.ExecContext(ctx, `
		INSERT INTO contracts (org_id, contract_reference, contract_name, contract_type, party_id, party_name, transport_mode, status, currency, contract_value, effective_date, expiry_date, owner, created_at, updated_at)
		VALUES (?, 'CNT-EVID-005', 'Master Service Agreement', 'Customer SLA', 5, 'Acme Global Corp', 'MULTIMODAL', 'ACTIVE', 'USD', 500000.00, NOW() - INTERVAL 100 DAY, NOW() + INTERVAL 25 DAY, 'Alex Reed', NOW(), NOW())
	`, testOrg)
	if err != nil {
		t.Fatalf("Failed to insert contract: %v", err)
	}
	contractID, _ := res.LastInsertId()

	evidence, err := repo.GetContractEvidence(ctx, testOrg, contractID)
	if err != nil {
		t.Fatalf("GetContractEvidence failed: %v", err)
	}
	if evidence == nil {
		t.Fatalf("Expected non-nil ContractEvidencePayload")
	}

	if evidence.ContractReference != "CNT-EVID-005" {
		t.Errorf("Expected reference CNT-EVID-005, got %s", evidence.ContractReference)
	}
	if !evidence.IsExpiringSoon {
		t.Errorf("Expected IsExpiringSoon to be true for contract expiring in 25 days")
	}
	if evidence.DaysUntilExpiry <= 0 || evidence.DaysUntilExpiry > 30 {
		t.Errorf("Expected DaysUntilExpiry between 1 and 30, got %d", evidence.DaysUntilExpiry)
	}
	if evidence.HasMissingOwner {
		t.Errorf("Expected HasMissingOwner to be false, owner is 'Alex Reed'")
	}

	summary, err := repo.GetContractComplianceSummary(ctx, testOrg)
	if err != nil {
		t.Fatalf("GetContractComplianceSummary failed: %v", err)
	}
	if summary.TotalContracts == 0 {
		t.Errorf("Expected at least 1 total contract, got %d", summary.TotalContracts)
	}
	if summary.ExpiringSoonContracts == 0 {
		t.Errorf("Expected at least 1 expiring soon contract, got %d", summary.ExpiringSoonContracts)
	}
}

func TestContractComplianceDraftGeneration(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo, nil, nil)
	ctx := context.Background()
	testOrg := int64(88898)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM contracts WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	contractID := int64(9991)
	cand := &Recommendation{
		OrgID:              testOrg,
		SourceType:         SourceContract,
		SourceID:           contractID,
		SourceReference:    "CNT-DRAFT-006",
		Title:              "Contract Expiry Notice: CNT-DRAFT-006",
		Description:        "Agreement with Pacific Ocean Freight ($80000.00) expires on 2026-09-25. Renewal preparation required.",
		Category:           CategoryContract,
		Priority:           PriorityHigh,
		RiskLevel:          RiskHigh,
		Confidence:         "HIGH",
		ConfidenceScore:    0.95,
		EvidenceJSON:       "[]",
		RecommendedAction:  "Prepare contract renewal reminder and verify counterparty renewal notice obligations.",
		ActionType:         ActionTypePrepareRenewalReminder,
		Status:             StatusNew,
		RequiresApproval:   true,
		CorrelationID:      "test-corr-draft",
		CreatedBy:          "SYSTEM",
		GeneratedBy:        "DETERMINISTIC_RULES",
		RuleApplied:        "CONTRACT_EXPIRING_SOON",
		DedupHash:          computeDedupHash(testOrg, SourceContract, contractID, CategoryContract, "CONTRACT_EXPIRING_SOON"),
		ContractID:         &contractID,
		DraftStatus:        DraftStatusNotGenerated,
	}

	created, err := repo.Create(ctx, cand)
	if err != nil {
		t.Fatalf("Failed to create recommendation: %v", err)
	}

	updated, err := svc.GenerateDraft(ctx, testOrg, created.ID, 1, "Legal Officer", "corr-draft-gen")
	if err != nil {
		t.Fatalf("GenerateDraft failed: %v", err)
	}
	if updated.DraftStatus != DraftStatusDrafted {
		t.Errorf("Expected DraftStatus %s, got %s", DraftStatusDrafted, updated.DraftStatus)
	}
	if updated.DraftSubject == nil || !strings.Contains(*updated.DraftSubject, "CNT-DRAFT-006") {
		t.Errorf("DraftSubject should contain contract reference, got: %v", updated.DraftSubject)
	}
	if updated.DraftBody == nil || !strings.Contains(*updated.DraftBody, "renewal") {
		t.Errorf("DraftBody should reference renewal, got: %v", updated.DraftBody)
	}

	// Verify draft is editable and saved
	saveInput := SaveDraftInput{
		Subject: "Edited Renewal Notice: Agreement CNT-DRAFT-006",
		Body:    "Custom legal text verified by officer.",
	}
	saved, err := svc.SaveDraft(ctx, testOrg, created.ID, saveInput, 1, "Legal Officer", "corr-draft-save")
	if err != nil {
		t.Fatalf("SaveDraft failed: %v", err)
	}
	if saved.DraftStatus != "SAVED" && saved.DraftStatus != DraftStatusEdited {
		t.Errorf("Expected DraftStatus SAVED or EDITED, got %s", saved.DraftStatus)
	}
	if *saved.DraftSubject != saveInput.Subject {
		t.Errorf("Expected subject %s, got %s", saveInput.Subject, *saved.DraftSubject)
	}
}

func TestTenantIsolationForContractCompliance(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	orgA := int64(88899)
	orgB := int64(88900)

	ensureTestOrg(ctx, db, orgA)
	ensureTestOrg(ctx, db, orgB)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM contracts WHERE org_id IN (?, ?)", orgA, orgB)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id IN (?, ?)", orgA, orgB)
	}()

	res, err := db.ExecContext(ctx, `
		INSERT INTO contracts (org_id, contract_reference, contract_name, contract_type, party_id, party_name, transport_mode, status, currency, contract_value, effective_date, expiry_date, owner, created_at, updated_at)
		VALUES (?, 'CNT-ISOLATION-A', 'Confidential OrgA Contract', 'Vendor Agreement', 1, 'OrgA Carrier', 'OCEAN', 'ACTIVE', 'USD', 100000.00, NOW(), NOW() + INTERVAL 30 DAY, 'Alice', NOW(), NOW())
	`, orgA)
	if err != nil {
		t.Fatalf("Failed to insert contract for orgA: %v", err)
	}
	contractIDA, _ := res.LastInsertId()

	// Query from Org B must fail to find Org A's contract
	_, err = repo.GetContractEvidence(ctx, orgB, contractIDA)
	if err == nil {
		t.Errorf("Expected error accessing Org A contract from Org B, but succeeded")
	}
}
