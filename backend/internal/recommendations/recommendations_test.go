package recommendations

import (
	"context"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("mysql", "root@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true")
	if err != nil {
		t.Skipf("Skipping integration test; MariaDB not available: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping integration test; MariaDB ping failed: %v", err)
	}
	return db
}

func TestRecommendationStatusTransitions(t *testing.T) {
	// Valid transitions
	validPairs := [][2]string{
		{StatusNew, StatusReviewed},
		{StatusNew, StatusAssigned},
		{StatusNew, StatusDismissed},
		{StatusReviewed, StatusAssigned},
		{StatusReviewed, StatusDismissed},
		{StatusReviewed, StatusCompleted},
		{StatusAssigned, StatusReviewed},
		{StatusAssigned, StatusApproved},
		{StatusAssigned, StatusRejected},
		{StatusAssigned, StatusCompleted},
		{StatusAssigned, StatusDismissed},
		{StatusApproved, StatusCompleted},
		{StatusApproved, StatusFailed},
	}

	for _, p := range validPairs {
		if !IsValidTransition(p[0], p[1]) {
			t.Errorf("Expected transition from %s to %s to be valid", p[0], p[1])
		}
	}

	// Invalid transitions
	invalidPairs := [][2]string{
		{StatusCompleted, StatusNew},
		{StatusCompleted, StatusReviewed},
		{StatusDismissed, StatusNew},
		{StatusDismissed, StatusApproved},
		{StatusFailed, StatusNew},
		{StatusApproved, StatusNew},
	}

	for _, p := range invalidPairs {
		if IsValidTransition(p[0], p[1]) {
			t.Errorf("Expected transition from %s to %s to be invalid", p[0], p[1])
		}
	}
}

func TestDeduplicationHashStability(t *testing.T) {
	hash1 := computeDedupHash(2, SourceInvoice, 101, CategoryFinance, "OVERDUE_INVOICE_COLLECTIONS")
	hash2 := computeDedupHash(2, SourceInvoice, 101, CategoryFinance, "OVERDUE_INVOICE_COLLECTIONS")
	hash3 := computeDedupHash(3, SourceInvoice, 101, CategoryFinance, "OVERDUE_INVOICE_COLLECTIONS")
	hash4 := computeDedupHash(2, SourceInvoice, 102, CategoryFinance, "OVERDUE_INVOICE_COLLECTIONS")

	if hash1 != hash2 {
		t.Errorf("Expected identical inputs to produce identical dedup hashes")
	}
	if hash1 == hash3 {
		t.Errorf("Different org_id must produce different dedup hashes")
	}
	if hash1 == hash4 {
		t.Errorf("Different source_id must produce different dedup hashes")
	}
}

func TestRecommendationLifecycleAndIdempotency(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	svc := NewService(repo, gen, nil)

	ctx := context.Background()
	testOrg := int64(99999) // Isolated test org

	// Cleanup test org recommendations before and after
	_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
	}()

	// 1. Manually create a test recommendation
	hash := computeDedupHash(testOrg, SourceShipment, 501, CategoryOperations, "TEST_RULE")
	cand := &Recommendation{
		OrgID:             testOrg,
		SourceType:        SourceShipment,
		SourceID:          501,
		SourceReference:   "SH-TEST-501",
		Title:             "Test Operational Recommendation",
		Description:       "Test description for shipment verification",
		Category:          CategoryOperations,
		Priority:          PriorityHigh,
		RiskLevel:         RiskHigh,
		Confidence:        "HIGH",
		ConfidenceScore:   0.92,
		EvidenceJSON:      `[{"source_module":"shipments","source_entity_id":501,"source_ref":"SH-TEST-501","field_name":"status","observed_value":"IN_TRANSIT","description":"Milestone delayed"}]`,
		RecommendedAction: "Verify shipment location with carrier",
		ActionType:        "INVESTIGATE_EXCEPTION",
		Status:            StatusNew,
		RequiresApproval:  false,
		CorrelationID:     "corr-test-1",
		CreatedBy:         "UNIT_TEST",
		GeneratedBy:       "DETERMINISTIC_RULES",
		RuleApplied:       "TEST_RULE",
		DedupHash:         hash,
	}

	created, err := repo.Create(ctx, cand)
	if err != nil {
		t.Fatalf("Failed to create test recommendation: %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("Expected valid inserted ID, got %d", created.ID)
	}
	if len(created.Evidence) != 1 {
		t.Fatalf("Expected 1 unpacked evidence item, got %d", len(created.Evidence))
	}

	// 2. Test GetByID
	fetched, err := svc.GetRecommendation(ctx, testOrg, created.ID)
	if err != nil {
		t.Fatalf("Failed to get recommendation: %v", err)
	}
	if fetched.Title != created.Title {
		t.Errorf("Expected title %s, got %s", created.Title, fetched.Title)
	}

	// 3. Test Mark Reviewed
	reviewed, err := svc.MarkReviewed(ctx, testOrg, created.ID, 12, "tester", "corr-2")
	if err != nil {
		t.Fatalf("Failed to mark reviewed: %v", err)
	}
	if reviewed.Status != StatusReviewed {
		t.Errorf("Expected status %s, got %s", StatusReviewed, reviewed.Status)
	}
	if reviewed.ReviewedAt == nil {
		t.Errorf("Expected reviewed_at timestamp to be set")
	}

	// 4. Test Assign
	assigned, err := svc.AssignRecommendation(ctx, testOrg, created.ID, AssignInput{
		AssigneeID:   42,
		AssigneeName: "Jane Dispatcher",
	}, 12, "tester", "corr-3")
	if err != nil {
		t.Fatalf("Failed to assign recommendation: %v", err)
	}
	if assigned.Status != StatusAssigned {
		t.Errorf("Expected status %s, got %s", StatusAssigned, assigned.Status)
	}
	if assigned.AssigneeName == nil || *assigned.AssigneeName != "Jane Dispatcher" {
		t.Errorf("Expected assignee name 'Jane Dispatcher', got %v", assigned.AssigneeName)
	}

	// 5. Test Invalid Transition (Assigned -> New is not allowed)
	_, err = svc.UpdateStatus(ctx, testOrg, created.ID, UpdateStatusInput{Status: StatusNew}, 12, "tester", "corr-4")
	if err == nil {
		t.Errorf("Expected error on invalid transition assigned -> new, got nil")
	}

	// 6. Test Dismiss requires reason
	_, err = svc.DismissRecommendation(ctx, testOrg, created.ID, DismissInput{Reason: ""}, 12, "tester", "corr-5")
	if err == nil {
		t.Errorf("Expected error when dismissing without reason, got nil")
	}

	// Dismiss with valid reason
	dismissed, err := svc.DismissRecommendation(ctx, testOrg, created.ID, DismissInput{Reason: "False alarm; carrier updated GPS"}, 12, "tester", "corr-6")
	if err != nil {
		t.Fatalf("Failed to dismiss recommendation: %v", err)
	}
	if dismissed.Status != StatusDismissed {
		t.Errorf("Expected status %s, got %s", StatusDismissed, dismissed.Status)
	}
	if dismissed.DismissedReason == nil || *dismissed.DismissedReason != "False alarm; carrier updated GPS" {
		t.Errorf("Expected dismissal reason to be stored")
	}

	// 7. Test Idempotency: DedupHash lookup
	existing, err := repo.GetByDedupHash(ctx, testOrg, hash)
	if err != nil {
		t.Fatalf("Failed to get by dedup hash: %v", err)
	}
	if existing == nil || existing.ID != created.ID {
		t.Fatalf("Expected to find existing record by dedup hash")
	}
}

func TestHighRiskApprovalSafetyGate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	svc := NewService(repo, gen, nil)

	ctx := context.Background()
	testOrg := int64(99998)

	_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
	}()

	// Create high-risk recommendation requiring approval
	hash := computeDedupHash(testOrg, SourceContract, 301, CategoryContract, "APPROVAL_RULE")
	rec := &Recommendation{
		OrgID:             testOrg,
		SourceType:        SourceContract,
		SourceID:          301,
		SourceReference:   "CTR-HIGH-301",
		Title:             "High Risk Contract Amendment",
		Description:       "Requires explicit executive approval before completion",
		Category:          CategoryContract,
		Priority:          PriorityCritical,
		RiskLevel:         RiskCritical,
		Confidence:        "HIGH",
		ConfidenceScore:   0.95,
		EvidenceJSON:      `[]`,
		RecommendedAction: "Review and approve renewal terms",
		ActionType:        "REVIEW_CONTRACT",
		Status:            StatusAssigned,
		RequiresApproval:  true,
		CorrelationID:     "gate-corr-1",
		CreatedBy:         "UNIT_TEST",
		GeneratedBy:       "DETERMINISTIC_RULES",
		RuleApplied:       "APPROVAL_RULE",
		DedupHash:         hash,
	}

	created, err := repo.Create(ctx, rec)
	if err != nil {
		t.Fatalf("Failed to create high-risk recommendation: %v", err)
	}

	// Attempt to complete directly without approval -> MUST FAIL
	_, err = svc.UpdateStatus(ctx, testOrg, created.ID, UpdateStatusInput{Status: StatusCompleted}, 1, "tester", "gate-corr-2")
	if err == nil {
		t.Errorf("Safety violation: High-risk recommendation was allowed to complete without approval")
	}

	// Transition to approved first
	approved, err := svc.UpdateStatus(ctx, testOrg, created.ID, UpdateStatusInput{Status: StatusApproved}, 1, "tester", "gate-corr-3")
	if err != nil {
		t.Fatalf("Failed to approve recommendation: %v", err)
	}
	if approved.Status != StatusApproved {
		t.Errorf("Expected status %s, got %s", StatusApproved, approved.Status)
	}

	// Now completing after approval -> MUST SUCCEED
	completed, err := svc.UpdateStatus(ctx, testOrg, created.ID, UpdateStatusInput{Status: StatusCompleted}, 1, "tester", "gate-corr-4")
	if err != nil {
		t.Fatalf("Failed to complete approved recommendation: %v", err)
	}
	if completed.Status != StatusCompleted {
		t.Errorf("Expected status %s, got %s", StatusCompleted, completed.Status)
	}
}

func TestCustomerFollowupDraftAndTaskIdempotency(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	svc := NewService(repo, gen, nil)

	ctx := context.Background()
	testOrg := int64(99997)
	testCustID := int64(777)
	testCustName := "Acme Global Freight"
	testFollowupType := FollowupTypeQuotationFollowup
	ownerID := int64(10)
	ownerName := "Varun Kanade"

	_, _ = db.ExecContext(ctx, "DELETE FROM customer_followup_tasks WHERE org_id = ?", testOrg)
	_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM customer_followup_tasks WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
	}()

	// 1. Create a customer follow-up recommendation
	hash := computeDedupHash(testOrg, SourceQuotation, 888, CategorySales, "QUOTE_APPROACHING_EXPIRY")
	rec := &Recommendation{
		OrgID:              testOrg,
		SourceType:         SourceQuotation,
		SourceID:           888,
		SourceReference:    "QT-2026-888",
		Title:              "Quotation QT-2026-888 Approaching Expiry",
		Description:        "Customer Acme Global Freight quotation requires follow-up prior to expiration",
		Category:           CategorySales,
		Priority:           PriorityHigh,
		RiskLevel:          RiskMedium,
		Confidence:         "HIGH",
		ConfidenceScore:    0.92,
		EvidenceJSON:       `[{"type":"QUOTE_EXPIRY","description":"Quote valid until 2026-09-15"}]`,
		RecommendedAction:  "Contact customer contact to confirm whether they plan to book.",
		ActionType:         "CUSTOMER_FOLLOWUP",
		Status:             StatusNew,
		RequiresApproval:   false,
		CorrelationID:      "followup-corr-1",
		CreatedBy:          "UNIT_TEST",
		GeneratedBy:        "DETERMINISTIC_RULES",
		RuleApplied:        "QUOTE_APPROACHING_EXPIRY",
		DedupHash:          hash,
		CustomerID:         &testCustID,
		CustomerName:       &testCustName,
		FollowupType:       &testFollowupType,
		SuggestedOwnerID:   &ownerID,
		SuggestedOwnerName: &ownerName,
	}

	created, err := repo.Create(ctx, rec)
	if err != nil {
		t.Fatalf("Failed to create recommendation: %v", err)
	}

	// 2. Draft is initially empty / NOT_GENERATED
	if created.DraftStatus != DraftStatusNotGenerated {
		t.Errorf("Expected draft status %s, got %s", DraftStatusNotGenerated, created.DraftStatus)
	}

	// 3. Explicitly generate draft
	withDraft, err := svc.GenerateDraft(ctx, testOrg, created.ID, 1, "tester", "followup-corr-2")
	if err != nil {
		t.Fatalf("Failed to generate draft: %v", err)
	}
	if withDraft.DraftStatus != DraftStatusDrafted {
		t.Errorf("Expected draft status %s, got %s", DraftStatusDrafted, withDraft.DraftStatus)
	}
	if withDraft.DraftSubject == nil || *withDraft.DraftSubject == "" {
		t.Errorf("Draft subject should be populated")
	}
	if withDraft.DraftBody == nil || *withDraft.DraftBody == "" {
		t.Errorf("Draft body should be populated")
	}

	// 4. Edit and save draft
	editedSubject := "Updated Subject: Acme Freight Follow-up"
	editedBody := "Hello Acme Team, checking in on quotation QT-2026-888."
	savedDraft, err := svc.SaveDraft(ctx, testOrg, created.ID, SaveDraftInput{
		Subject: editedSubject,
		Body:    editedBody,
	}, 1, "tester", "followup-corr-3")
	if err != nil {
		t.Fatalf("Failed to save draft: %v", err)
	}
	if *savedDraft.DraftSubject != editedSubject || *savedDraft.DraftBody != editedBody {
		t.Errorf("Saved draft does not match edited text")
	}

	// 5. Create Follow-up Task
	taskInput := CreateFollowupTaskInput{
		AssigneeID:   &ownerID,
		AssigneeName: &ownerName,
		Priority:     PriorityHigh,
		Notes:        "Please call customer representative by Thursday",
	}
	task1, err := svc.CreateFollowupTask(ctx, testOrg, created.ID, taskInput, 1, "tester", "followup-corr-4")
	if err != nil {
		t.Fatalf("Failed to create followup task: %v", err)
	}
	if task1.ID <= 0 {
		t.Errorf("Expected valid task ID, got %d", task1.ID)
	}
	if task1.CustomerName != testCustName {
		t.Errorf("Expected customer name %s, got %s", testCustName, task1.CustomerName)
	}

	// 6. Test IDEMPOTENCY: Repeated task creation for the same recommendation returns existing task without duplicate
	task2, err := svc.CreateFollowupTask(ctx, testOrg, created.ID, taskInput, 1, "tester", "followup-corr-5")
	if err != nil {
		t.Fatalf("Repeated task creation failed: %v", err)
	}
	if task2.ID != task1.ID {
		t.Errorf("Idempotency violation: Expected same task ID %d, got %d", task1.ID, task2.ID)
	}

	// Verify total count of tasks in db is exactly 1
	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM customer_followup_tasks WHERE org_id = ? AND recommendation_id = ?", testOrg, created.ID)
	if err != nil {
		t.Fatalf("Failed to count tasks: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected exactly 1 task in database, found %d", count)
	}

	// 7. Verify Followup Stats
	stats, err := svc.GetFollowupStats(ctx, testOrg)
	if err != nil {
		t.Fatalf("Failed to get followup stats: %v", err)
	}
	if stats.TotalFollowups < 1 {
		t.Errorf("Expected at least 1 total followup in stats, got %d", stats.TotalFollowups)
	}
	if stats.TasksCreatedCount < 1 {
		t.Errorf("Expected at least 1 task created in stats, got %d", stats.TasksCreatedCount)
	}
}

