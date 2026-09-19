package recommendations

import (
	"context"
	"testing"
)

func TestTenantIsolationAndCrossTenantAccess(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	svc := NewService(repo, gen, nil)

	ctx := context.Background()
	orgA := int64(88001)
	orgB := int64(88002)

	// Clean up test data
	_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id IN (?, ?)", orgA, orgB)
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id IN (?, ?)", orgA, orgB)
	}()

	// Create recommendation belonging to orgA
	recA := &Recommendation{
		OrgID:             orgA,
		SourceType:        SourceInvoice,
		SourceID:          1001,
		SourceReference:   "INV-ORGA-1001",
		Title:             "Org A Confidential Finance Alert",
		Description:       "Confidential overdue invoice details for Org A",
		Category:          CategoryFinance,
		Priority:          PriorityHigh,
		RiskLevel:         RiskHigh,
		Confidence:        "HIGH",
		ConfidenceScore:   0.95,
		EvidenceJSON:      `[{"source_module":"finance","source_entity_id":1001,"source_ref":"INV-ORGA-1001","field_name":"balance","observed_value":50000}]`,
		RecommendedAction: "Collect payment",
		ActionType:        "REVIEW_INVOICE",
		Status:            StatusNew,
		RequiresApproval:  false,
		CorrelationID:     "tenant-corr-1",
		CreatedBy:         "SECURITY_TEST",
		GeneratedBy:       "DETERMINISTIC_RULES",
		RuleApplied:       "TEST_RULE",
		DedupHash:         computeDedupHash(orgA, SourceInvoice, 1001, CategoryFinance, "TEST_RULE"),
	}

	createdA, err := repo.Create(ctx, recA)
	if err != nil {
		t.Fatalf("Failed to create recommendation for Org A: %v", err)
	}

	// 1. Org B must NOT be able to get Org A's recommendation by ID
	recForB, err := svc.GetRecommendation(ctx, orgB, createdA.ID)
	if err == nil && recForB != nil {
		t.Fatalf("Security breach: Org B was able to view Org A's recommendation %d", createdA.ID)
	}

	// 2. Org B's List query must NOT return Org A's recommendations
	listB, err := svc.ListRecommendations(ctx, orgB, RecommendationFilter{})
	if err != nil {
		t.Fatalf("Failed to list recommendations for Org B: %v", err)
	}
	for _, item := range listB.Recommendations {
		if item.OrgID == orgA || item.ID == createdA.ID {
			t.Fatalf("Security breach: Org B list results leaked Org A recommendation %d", item.ID)
		}
	}

	// 3. Org B must NOT be able to mutate (status change, assign, dismiss) Org A's recommendation
	_, err = svc.UpdateStatus(ctx, orgB, createdA.ID, UpdateStatusInput{Status: StatusReviewed}, 99, "intruder", "sec-2")
	if err == nil {
		t.Fatalf("Security breach: Org B was able to change status of Org A's recommendation")
	}

	_, err = svc.AssignRecommendation(ctx, orgB, createdA.ID, AssignInput{AssigneeID: 99, AssigneeName: "Intruder"}, 99, "intruder", "sec-3")
	if err == nil {
		t.Fatalf("Security breach: Org B was able to assign Org A's recommendation")
	}

	_, err = svc.DismissRecommendation(ctx, orgB, createdA.ID, DismissInput{Reason: "Malicious dismissal"}, 99, "intruder", "sec-4")
	if err == nil {
		t.Fatalf("Security breach: Org B was able to dismiss Org A's recommendation")
	}

	// 4. Source query by Org B for Org A's source must return empty
	bySourceB, err := svc.ListBySource(ctx, orgB, SourceInvoice, 1001)
	if err != nil {
		t.Fatalf("Failed to query source for Org B: %v", err)
	}
	if len(bySourceB) != 0 {
		t.Fatalf("Security breach: Org B retrieved recommendations for Org A's source entity")
	}

	// 5. Follow-Up Draft generation: Org B must NOT be able to generate or save a draft on Org A's recommendation
	_, err = svc.GenerateDraft(ctx, orgB, createdA.ID, 99, "intruder", "sec-5")
	if err == nil {
		t.Fatalf("Security breach: Org B generated draft on Org A recommendation")
	}

	_, err = svc.SaveDraft(ctx, orgB, createdA.ID, SaveDraftInput{Subject: "Malicious", Body: "Malicious"}, 99, "intruder", "sec-6")
	if err == nil {
		t.Fatalf("Security breach: Org B saved draft on Org A recommendation")
	}

	// 6. Follow-Up Task creation: Org B must NOT be able to create a task on Org A's recommendation
	_, err = svc.CreateFollowupTask(ctx, orgB, createdA.ID, CreateFollowupTaskInput{Priority: "critical"}, 99, "intruder", "sec-7")
	if err == nil {
		t.Fatalf("Security breach: Org B created task on Org A recommendation")
	}
}

func TestReadOnlySafetyUnderlyingTablesUnchanged(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	svc := NewService(repo, gen, nil)

	ctx := context.Background()
	testOrg := int64(2) // Real test org

	// Record initial checksum / counts of business tables
	var invCountBefore, shipCountBefore, rfqCountBefore, custCountBefore int64
	_ = db.GetContext(ctx, &invCountBefore, "SELECT COUNT(*) FROM shipment_invoices WHERE org_id = ?", testOrg)
	_ = db.GetContext(ctx, &shipCountBefore, "SELECT COUNT(*) FROM shipments WHERE org_id = ?", testOrg)
	_ = db.GetContext(ctx, &rfqCountBefore, "SELECT COUNT(*) FROM rfqs WHERE org_id = ?", testOrg)
	_ = db.GetContext(ctx, &custCountBefore, "SELECT COUNT(*) FROM customers WHERE org_id = ?", testOrg)

	// Execute recommendation generation
	genResult, err := svc.GenerateRecommendations(ctx, testOrg, "readonly-safety-corr", 1)
	if err != nil {
		t.Fatalf("Failed to generate recommendations: %v", err)
	}
	if genResult.TotalEvaluated < 0 {
		t.Fatalf("Invalid total evaluated")
	}

	// Record counts after generation
	var invCountAfter, shipCountAfter, rfqCountAfter, custCountAfter int64
	_ = db.GetContext(ctx, &invCountAfter, "SELECT COUNT(*) FROM shipment_invoices WHERE org_id = ?", testOrg)
	_ = db.GetContext(ctx, &shipCountAfter, "SELECT COUNT(*) FROM shipments WHERE org_id = ?", testOrg)
	_ = db.GetContext(ctx, &rfqCountAfter, "SELECT COUNT(*) FROM rfqs WHERE org_id = ?", testOrg)
	_ = db.GetContext(ctx, &custCountAfter, "SELECT COUNT(*) FROM customers WHERE org_id = ?", testOrg)

	if invCountBefore != invCountAfter {
		t.Fatalf("Read-only safety violation: shipment_invoices count changed from %d to %d", invCountBefore, invCountAfter)
	}
	if shipCountBefore != shipCountAfter {
		t.Fatalf("Read-only safety violation: shipments count changed from %d to %d", shipCountBefore, shipCountAfter)
	}
	if rfqCountBefore != rfqCountAfter {
		t.Fatalf("Read-only safety violation: rfqs count changed from %d to %d", rfqCountBefore, rfqCountAfter)
	}
	if custCountBefore != custCountAfter {
		t.Fatalf("Read-only safety violation: customers count changed from %d to %d", custCountBefore, custCountAfter)
	}
}
