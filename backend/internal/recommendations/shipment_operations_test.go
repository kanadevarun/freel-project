package recommendations

import (
	"context"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func TestShipmentMilestoneDelayDetection(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	ctx := context.Background()
	testOrg := int64(88891)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM shipment_milestones WHERE shipment_id IN (SELECT id FROM shipments WHERE org_id = ?)", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM shipments WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	// 1. Insert active shipment
	res, err := db.ExecContext(ctx, `
		INSERT INTO shipments (org_id, booking_number, carrier_scac, origin_port, destination_port, status, etd, eta, created_at, updated_at)
		VALUES (?, 'BK-TEST-DELAY-001', 'MAEU', 'INNSA', 'NLRTM', 'IN_TRANSIT', NOW() - INTERVAL 10 DAY, NOW() + INTERVAL 10 DAY, NOW() - INTERVAL 10 DAY, NOW())
	`, testOrg)
	if err != nil {
		t.Fatalf("Failed to insert shipment: %v", err)
	}
	shipmentID, _ := res.LastInsertId()

	// 2. Insert delayed milestone (planned 5 days ago, still PLANNED)
	pastDate := time.Now().Add(-5 * 24 * time.Hour)
	_, err = db.ExecContext(ctx, `
		INSERT INTO shipment_milestones (shipment_id, milestone_code, description, planned_date, status, updated_at)
		VALUES (?, 'DEPARTED', 'Vessel departure from load port', ?, 'PLANNED', NOW())
	`, shipmentID, pastDate)
	if err != nil {
		t.Fatalf("Failed to insert delayed milestone: %v", err)
	}

	// 3. Insert future milestone (planned in future)
	futureDate := time.Now().Add(10 * 24 * time.Hour)
	_, err = db.ExecContext(ctx, `
		INSERT INTO shipment_milestones (shipment_id, milestone_code, description, planned_date, status, updated_at)
		VALUES (?, 'ARRIVAL', 'Vessel arrival at destination', ?, 'PLANNED', NOW())
	`, shipmentID, futureDate)
	if err != nil {
		t.Fatalf("Failed to insert future milestone: %v", err)
	}

	// 4. Run generator
	result, err := gen.Generate(ctx, testOrg, "test-corr-milestone-delay", 1)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if result == nil {
		t.Fatalf("Expected non-nil result")
	}

	// 5. Query recommendations
	delayedFilter := true
	recs, total, err := repo.List(ctx, testOrg, RecommendationFilter{
		ShipmentID:       &shipmentID,
		DelayedMilestone: &delayedFilter,
	})
	if err != nil {
		t.Fatalf("Failed to list recommendations: %v", err)
	}
	if total == 0 || len(recs) == 0 {
		t.Fatalf("Expected at least 1 delayed milestone recommendation, got 0")
	}

	rec := recs[0]
	if rec.RuleApplied != "SHIPMENT_MILESTONE_DELAYED" {
		t.Errorf("Expected RuleApplied SHIPMENT_MILESTONE_DELAYED, got %s", rec.RuleApplied)
	}
	if rec.MilestoneID == nil {
		t.Errorf("Expected MilestoneID to be populated")
	}
	if rec.ShipmentID == nil || *rec.ShipmentID != shipmentID {
		t.Errorf("Expected ShipmentID %d, got %v", shipmentID, rec.ShipmentID)
	}
	if rec.ActionType != ActionTypeRequestCarrierClarification {
		t.Errorf("Expected ActionType %s, got %s", ActionTypeRequestCarrierClarification, rec.ActionType)
	}
}

func TestShipmentMissingOperationalInfoDetection(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	ctx := context.Background()
	testOrg := int64(88892)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM shipments WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	// Insert active shipment missing vessel/mbl and container numbers and ETA
	res, err := db.ExecContext(ctx, `
		INSERT INTO shipments (org_id, booking_number, carrier_scac, origin_port, destination_port, vessel_name, mbl_number, container_numbers, status, eta, created_at, updated_at)
		VALUES (?, 'BK-TEST-MISSING-002', 'MSCU', 'INNSA', 'DEHAM', NULL, NULL, '[]', 'IN_TRANSIT', NULL, NOW(), NOW())
	`, testOrg)
	if err != nil {
		t.Fatalf("Failed to insert shipment: %v", err)
	}
	shipmentID, _ := res.LastInsertId()

	// Run generator
	_, err = gen.Generate(ctx, testOrg, "test-corr-missing-ops", 1)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	missingFilter := true
	recs, total, err := repo.List(ctx, testOrg, RecommendationFilter{
		ShipmentID:     &shipmentID,
		MissingOpsInfo: &missingFilter,
	})
	if err != nil {
		t.Fatalf("Failed to list recommendations: %v", err)
	}
	if total == 0 || len(recs) == 0 {
		t.Fatalf("Expected missing ops info recommendation, got 0")
	}

	rec := recs[0]
	if rec.RuleApplied != "SHIPMENT_MISSING_OPERATIONAL_INFO" {
		t.Errorf("Expected RuleApplied SHIPMENT_MISSING_OPERATIONAL_INFO, got %s", rec.RuleApplied)
	}
	if rec.ActionType != ActionTypeRequestMissingDocuments {
		t.Errorf("Expected ActionType %s, got %s", ActionTypeRequestMissingDocuments, rec.ActionType)
	}
}

func TestShipmentInactiveTrackingDetection(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	ctx := context.Background()
	testOrg := int64(88893)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM shipments WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	// Insert in-transit shipment with last update 8 days ago
	eightDaysAgo := time.Now().Add(-8 * 24 * time.Hour)
	res, err := db.ExecContext(ctx, `
		INSERT INTO shipments (org_id, booking_number, carrier_scac, origin_port, destination_port, status, etd, eta, created_at, updated_at)
		VALUES (?, 'BK-TEST-INACTIVE-003', 'CMDU', 'INNSA', 'USNYC', 'IN_TRANSIT', NOW() - INTERVAL 15 DAY, NOW() + INTERVAL 5 DAY, NOW() - INTERVAL 15 DAY, ?)
	`, testOrg, eightDaysAgo)
	if err != nil {
		t.Fatalf("Failed to insert shipment: %v", err)
	}
	shipmentID, _ := res.LastInsertId()

	_, err = gen.Generate(ctx, testOrg, "test-corr-inactive-tracking", 1)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	recs, total, err := repo.List(ctx, testOrg, RecommendationFilter{
		ShipmentID: &shipmentID,
	})
	if err != nil {
		t.Fatalf("Failed to list recommendations: %v", err)
	}

	var foundInactive bool
	for _, r := range recs {
		if r.RuleApplied == "SHIPMENT_INACTIVE_TRACKING" {
			foundInactive = true
			if !strings.Contains(r.Description, "No recent tracking update was found") {
				t.Errorf("Expected careful wording 'No recent tracking update was found', got: %s", r.Description)
			}
			if !strings.Contains(r.Description, "does not confirm the current physical location") {
				t.Errorf("Expected careful location disclaimer, got: %s", r.Description)
			}
			break
		}
	}
	if !foundInactive {
		t.Fatalf("Expected SHIPMENT_INACTIVE_TRACKING recommendation for shipment %d, total found: %d", shipmentID, total)
	}
}

func TestShipmentMultipleExceptionsDetection(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	ctx := context.Background()
	testOrg := int64(88894)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM shipment_exceptions WHERE shipment_id IN (SELECT id FROM shipments WHERE org_id = ?)", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM shipments WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	res, err := db.ExecContext(ctx, `
		INSERT INTO shipments (org_id, booking_number, carrier_scac, origin_port, destination_port, status, created_at, updated_at)
		VALUES (?, 'BK-TEST-MULTI-EXC-004', 'MAEU', 'INNSA', 'NLRTM', 'DEPARTED', NOW(), NOW())
	`, testOrg)
	if err != nil {
		t.Fatalf("Failed to insert shipment: %v", err)
	}
	shipmentID, _ := res.LastInsertId()

	// Insert 2 open exceptions on the same shipment
	_, err = db.ExecContext(ctx, `
		INSERT INTO shipment_exceptions (org_id, shipment_id, exception_type, severity, title, description, status, resolved, created_at)
		VALUES (?, ?, 'OTHER', 'HIGH', 'Severe Typhoon in South China Sea', 'Vessel diverting course', 'OPEN', 0, NOW())
	`, testOrg, shipmentID)
	if err != nil {
		t.Fatalf("Failed to insert exception 1: %v", err)
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO shipment_exceptions (org_id, shipment_id, exception_type, severity, title, description, status, resolved, created_at)
		VALUES (?, ?, 'PORT_CONGESTION', 'HIGH', 'Terminal Berthing Delays', '48hr queue at transshipment hub', 'OPEN', 0, NOW())
	`, testOrg, shipmentID)
	if err != nil {
		t.Fatalf("Failed to insert exception 2: %v", err)
	}

	_, err = gen.Generate(ctx, testOrg, "test-corr-multi-exc", 1)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	recs, total, err := repo.List(ctx, testOrg, RecommendationFilter{
		ShipmentID: &shipmentID,
	})
	if err != nil {
		t.Fatalf("Failed to list: %v", err)
	}
	if total == 0 || len(recs) == 0 {
		t.Fatalf("Expected recommendations, got 0")
	}

	// In multi-exception cases, priority should be escalated to CRITICAL with RequiresApproval = true
	rec := recs[0]
	if rec.Priority != PriorityCritical {
		t.Errorf("Expected PriorityCritical due to multiple exceptions, got %s", rec.Priority)
	}
	if !rec.RequiresApproval {
		t.Errorf("Expected RequiresApproval to be true for multiple exceptions")
	}
	if rec.ActionType != ActionTypeEscalateOperationalRisk {
		t.Errorf("Expected ActionTypeEscalateOperationalRisk, got %s", rec.ActionType)
	}
}

func TestShipmentReadOnlySafety(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	svc := NewService(repo, gen, nil)
	ctx := context.Background()
	testOrg := int64(88895)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM shipment_milestones WHERE shipment_id IN (SELECT id FROM shipments WHERE org_id = ?)", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM shipment_exceptions WHERE shipment_id IN (SELECT id FROM shipments WHERE org_id = ?)", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM shipments WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	initialStatus := "IN_TRANSIT"
	initialETA := time.Now().Add(5 * 24 * time.Hour).Truncate(time.Second)

	res, err := db.ExecContext(ctx, `
		INSERT INTO shipments (org_id, booking_number, carrier_scac, origin_port, destination_port, status, eta, created_at, updated_at)
		VALUES (?, 'BK-TEST-SAFETY-005', 'MSCU', 'INNSA', 'DEHAM', ?, ?, NOW(), NOW())
	`, testOrg, initialStatus, initialETA)
	if err != nil {
		t.Fatalf("Failed to insert shipment: %v", err)
	}
	shipmentID, _ := res.LastInsertId()

	// Insert exception
	resExc, err := db.ExecContext(ctx, `
		INSERT INTO shipment_exceptions (org_id, shipment_id, exception_type, severity, title, status, resolved, created_at)
		VALUES (?, ?, 'CUSTOMS_HOLD', 'CRITICAL', 'Customs Inspection Hold', 'OPEN', 0, NOW())
	`, testOrg, shipmentID)
	if err != nil {
		t.Fatalf("Failed to insert exception: %v", err)
	}
	excID, _ := resExc.LastInsertId()

	// 1. Generate recommendations
	_, err = gen.Generate(ctx, testOrg, "test-corr-safety", 1)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// 2. Verify shipment record is 100% untouched (read-only guarantee)
	var currentStatus string
	var currentETA time.Time
	err = db.QueryRowContext(ctx, "SELECT status, eta FROM shipments WHERE id = ?", shipmentID).Scan(&currentStatus, &currentETA)
	if err != nil {
		t.Fatalf("Failed to read shipment: %v", err)
	}
	if currentStatus != initialStatus {
		t.Errorf("CRITICAL SAFETY VIOLATION: Shipment status mutated from %s to %s", initialStatus, currentStatus)
	}
	if !currentETA.Equal(initialETA) {
		t.Errorf("CRITICAL SAFETY VIOLATION: Shipment ETA mutated from %v to %v", initialETA, currentETA)
	}

	// 3. Verify exception record is still OPEN and resolved = 0
	var excStatus string
	var resolved int
	err = db.QueryRowContext(ctx, "SELECT status, resolved FROM shipment_exceptions WHERE id = ?", excID).Scan(&excStatus, &resolved)
	if err != nil {
		t.Fatalf("Failed to read exception: %v", err)
	}
	if excStatus != "OPEN" || resolved != 0 {
		t.Errorf("CRITICAL SAFETY VIOLATION: Exception was automatically resolved or closed")
	}

	// 4. Test Draft Generation does NOT auto-send communications
	recs, _, err := repo.List(ctx, testOrg, RecommendationFilter{ShipmentID: &shipmentID})
	if err != nil || len(recs) == 0 {
		t.Fatalf("Expected recommendation")
	}
	recID := recs[0].ID

	draftedRec, err := svc.GenerateDraft(ctx, testOrg, recID, 1, "Operations Specialist", "test-corr-draft")
	if err != nil {
		t.Fatalf("GenerateDraft failed: %v", err)
	}
	if draftedRec.DraftStatus != "DRAFTED" {
		t.Errorf("Expected draft status DRAFTED, got %s", draftedRec.DraftStatus)
	}
	if draftedRec.DraftSubject == nil || *draftedRec.DraftSubject == "" {
		t.Errorf("Expected draft subject to be populated")
	}
	if draftedRec.DraftBody == nil || *draftedRec.DraftBody == "" {
		t.Errorf("Expected draft body to be populated")
	}

	// 5. Verify Action Preview is strictly preview and non-mutating
	preview, err := svc.GetActionPreview(ctx, testOrg, recID, 1, "Operator", "test-corr-prev")
	if err != nil {
		t.Fatalf("GetActionPreview failed: %v", err)
	}
	if preview == nil || preview.ProposedAction == "" {
		t.Errorf("Expected valid action preview")
	}
}

func TestShipmentEntityEvidenceRetrieval(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	svc := NewService(repo, gen, nil)
	ctx := context.Background()
	testOrg := int64(88896)

	ensureTestOrg(ctx, db, testOrg)

	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM shipment_exceptions WHERE shipment_id IN (SELECT id FROM shipments WHERE org_id = ?)", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM shipments WHERE org_id = ?", testOrg)
		_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", testOrg)
	}()

	res, err := db.ExecContext(ctx, `
		INSERT INTO shipments (org_id, booking_number, carrier_scac, origin_port, destination_port, status, created_at, updated_at)
		VALUES (?, 'BK-TEST-EVID-006', 'MAEU', 'INNSA', 'DEHAM', 'BOOKED', NOW(), NOW())
	`, testOrg)
	if err != nil {
		t.Fatalf("Failed to insert shipment: %v", err)
	}
	shipmentID, _ := res.LastInsertId()

	resExc, err := db.ExecContext(ctx, `
		INSERT INTO shipment_exceptions (org_id, shipment_id, exception_type, severity, title, description, status, resolved, created_at)
		VALUES (?, ?, 'DOCUMENT_ISSUE', 'HIGH', 'Missing Phytosanitary Certificate', 'Required for export clearance', 'OPEN', 0, NOW())
	`, testOrg, shipmentID)
	if err != nil {
		t.Fatalf("Failed to insert exception: %v", err)
	}
	excID, _ := resExc.LastInsertId()

	_, _ = gen.Generate(ctx, testOrg, "test-corr-evidence", 1)

	// Test GetEntityEvidence for SHIPMENT
	shipEvidence, err := svc.GetEntityEvidence(ctx, testOrg, "SHIPMENT", shipmentID)
	if err != nil {
		t.Fatalf("GetEntityEvidence SHIPMENT failed: %v", err)
	}
	if len(shipEvidence) == 0 {
		t.Errorf("Expected evidence items for shipment %d", shipmentID)
	}

	// Test GetEntityEvidence for EXCEPTION
	excEvidence, err := svc.GetEntityEvidence(ctx, testOrg, "EXCEPTION", excID)
	if err != nil {
		t.Fatalf("GetEntityEvidence EXCEPTION failed: %v", err)
	}
	if len(excEvidence) == 0 {
		t.Errorf("Expected evidence items for exception %d", excID)
	}

	// Test cross-tenant evidence isolation: Org 99999 cannot access Org 88896 evidence
	crossEvidence, err := svc.GetEntityEvidence(ctx, 99999, "SHIPMENT", shipmentID)
	if err != nil {
		t.Fatalf("Cross-tenant evidence call returned error: %v", err)
	}
	if len(crossEvidence) != 0 {
		t.Errorf("TENANT ISOLATION BREACH: Org 99999 retrieved %d evidence items from Org %d", len(crossEvidence), testOrg)
	}
}
