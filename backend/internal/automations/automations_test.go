package automations

import (
	"context"
	"fmt"
	"testing"
	"time"

	auditRepoPkg "github.com/freel/backend/internal/audit/repository"
	auditSvcPkg "github.com/freel/backend/internal/audit/service"
	"github.com/freel/backend/internal/recommendations"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("mysql", "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&multiStatements=true")
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	return db
}

func ensureTestOrg(ctx context.Context, db *sqlx.DB, orgID int64) {
	_, _ = db.ExecContext(ctx, `
		INSERT IGNORE INTO organizations (id, name, created_at, updated_at)
		VALUES (?, 'Automations Test Org', NOW(), NOW())
	`, orgID)
}

func cleanupTestOrg(ctx context.Context, db *sqlx.DB, orgID int64) {
	_, _ = db.ExecContext(ctx, "DELETE FROM ai_operational_insights WHERE org_id = ?", orgID)
	_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", orgID)
	_, _ = db.ExecContext(ctx, "DELETE FROM ai_automation_executions WHERE org_id = ?", orgID)
	_, _ = db.ExecContext(ctx, "DELETE FROM ai_automations WHERE org_id = ?", orgID)
	_, _ = db.ExecContext(ctx, "DELETE FROM customer_invoices WHERE org_id = ?", orgID)
	_, _ = db.ExecContext(ctx, "DELETE FROM customers WHERE org_id = ?", orgID)
	_, _ = db.ExecContext(ctx, "DELETE FROM organizations WHERE id = ?", orgID)
}

func TestCalculateNextRun(t *testing.T) {
	refTime := time.Date(2026, 9, 8, 7, 0, 0, 0, time.UTC)

	// 1. Daily schedule: 08:00 UTC (future today)
	nextDaily, err := CalculateNextRun(ScheduleTypeDaily, "08:00", nil, "UTC", refTime)
	if err != nil {
		t.Fatalf("CalculateNextRun failed: %v", err)
	}
	expectedDaily := time.Date(2026, 9, 8, 8, 0, 0, 0, time.UTC)
	if !nextDaily.Equal(expectedDaily) {
		t.Errorf("Expected next daily run %v, got %v", expectedDaily, nextDaily)
	}

	// 2. Daily schedule: 06:00 UTC (past today -> should run tomorrow at 06:00)
	nextTomorrow, err := CalculateNextRun(ScheduleTypeDaily, "06:00", nil, "UTC", refTime)
	if err != nil {
		t.Fatalf("CalculateNextRun failed: %v", err)
	}
	expectedTomorrow := time.Date(2026, 9, 9, 6, 0, 0, 0, time.UTC)
	if !nextTomorrow.Equal(expectedTomorrow) {
		t.Errorf("Expected next run tomorrow %v, got %v", expectedTomorrow, nextTomorrow)
	}

	// 3. Hourly schedule: should be next hour
	nextHourly, err := CalculateNextRun(ScheduleTypeHourly, "08:15", nil, "UTC", refTime)
	if err != nil {
		t.Fatalf("CalculateNextRun hourly failed: %v", err)
	}
	expectedHourly := time.Date(2026, 9, 8, 8, 15, 0, 0, time.UTC)
	if !nextHourly.Equal(expectedHourly) {
		t.Errorf("Expected next hourly run %v, got %v", expectedHourly, nextHourly)
	}

	// 4. Weekly schedule: Tuesday is Sept 8, 2026. Schedule on Wednesday at 10:00
	wed := "WED"
	nextWeekly, err := CalculateNextRun(ScheduleTypeWeekly, "10:00", &wed, "UTC", refTime)
	if err != nil {
		t.Fatalf("CalculateNextRun weekly failed: %v", err)
	}
	expectedWeekly := time.Date(2026, 9, 9, 10, 0, 0, 0, time.UTC)
	if !nextWeekly.Equal(expectedWeekly) {
		t.Errorf("Expected next weekly run %v, got %v", expectedWeekly, nextWeekly)
	}
}

func TestAutomationCRUDAndTenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	orgA := int64(88901)
	orgB := int64(88902)

	ensureTestOrg(ctx, db, orgA)
	ensureTestOrg(ctx, db, orgB)
	defer cleanupTestOrg(ctx, db, orgA)
	defer cleanupTestOrg(ctx, db, orgB)

	repo := NewRepository(db)
	auditRepo := auditRepoPkg.NewMySQLRepository(db)
	auditSvc := auditSvcPkg.NewService(auditRepo, db)
	recRepo := recommendations.NewRepository(db)
	recGen := recommendations.NewGenerator(db, recRepo)
	svc := NewService(repo, recGen, recRepo, auditSvc)

	// 1. Create automation for Org A
	createdA, err := svc.CreateAutomation(ctx, orgA, 1, CreateAutomationInput{
		Name:           "Daily Invoice Review Org A",
		AutomationType: AutomationTypeDailyOverdueInvoiceReview,
		ScheduleType:   ScheduleTypeDaily,
		ScheduleTime:   "08:00",
		Timezone:       "UTC",
	})
	if err != nil {
		t.Fatalf("Failed to create automation for Org A: %v", err)
	}
	if createdA.ID <= 0 {
		t.Fatalf("Expected valid automation ID")
	}

	// 2. Tenant isolation check: Org B must not see Org A's automation
	_, err = svc.GetAutomation(ctx, orgB, createdA.ID)
	if err == nil {
		t.Fatalf("Expected cross-tenant access to fail with not found")
	}

	listB, totalB, err := svc.ListAutomations(ctx, orgB, AutomationFilter{})
	if err != nil {
		t.Fatalf("Failed to list automations for Org B: %v", err)
	}
	if totalB != 0 || len(listB) != 0 {
		t.Errorf("Expected 0 automations for Org B, got %d", totalB)
	}

	// 3. Update automation in Org A
	newDesc := "Updated daily review description"
	updatedA, err := svc.UpdateAutomation(ctx, orgA, createdA.ID, 1, UpdateAutomationInput{
		Name:        "Renamed Daily Review",
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("Failed to update automation: %v", err)
	}
	if updatedA.Name != "Renamed Daily Review" {
		t.Errorf("Expected name to be updated, got %s", updatedA.Name)
	}

	// 4. Enable / Disable toggle
	disabledA, err := svc.SetEnabled(ctx, orgA, createdA.ID, false, 1)
	if err != nil {
		t.Fatalf("Failed to disable automation: %v", err)
	}
	if disabledA.IsEnabled {
		t.Errorf("Expected automation to be disabled")
	}

	enabledA, err := svc.SetEnabled(ctx, orgA, createdA.ID, true, 1)
	if err != nil {
		t.Fatalf("Failed to enable automation: %v", err)
	}
	if !enabledA.IsEnabled {
		t.Errorf("Expected automation to be enabled")
	}

	// 5. Delete automation
	if err := svc.DeleteAutomation(ctx, orgA, createdA.ID, 1); err != nil {
		t.Fatalf("Failed to delete automation: %v", err)
	}
	_, err = svc.GetAutomation(ctx, orgA, createdA.ID)
	if err == nil {
		t.Errorf("Expected deleted automation to not be found")
	}
}

func TestManualRunAndExecutionLinkage(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	testOrg := int64(88903)

	ensureTestOrg(ctx, db, testOrg)
	defer cleanupTestOrg(ctx, db, testOrg)

	repo := NewRepository(db)
	auditRepo := auditRepoPkg.NewMySQLRepository(db)
	auditSvc := auditSvcPkg.NewService(auditRepo, db)
	recRepo := recommendations.NewRepository(db)
	recGen := recommendations.NewGenerator(db, recRepo)
	svc := NewService(repo, recGen, recRepo, auditSvc)

	// Insert overdue customer invoice
	pastDue := time.Now().Add(-25 * 24 * time.Hour)
	_, _ = db.ExecContext(ctx, `
		INSERT INTO customers (id, org_id, name, account_status, created_at, updated_at)
		VALUES (101, ?, 'Acme Importers', 'ACTIVE', NOW(), NOW())
	`, testOrg)
	_, err := db.ExecContext(ctx, `
		INSERT INTO customer_invoices (org_id, customer_id, customer_name, invoice_number, currency, total_amount, balance_due, status, invoice_date, due_date, created_at, updated_at)
		VALUES (?, 101, 'Acme Importers', 'INV-TEST-AUTO-01', 'USD', 14500.00, 14500.00, 'Overdue', NOW() - INTERVAL 30 DAY, ?, NOW() - INTERVAL 30 DAY, NOW())
	`, testOrg, pastDue)
	if err != nil {
		t.Fatalf("Failed to insert invoice: %v", err)
	}

	// Create automation
	auto, err := svc.CreateAutomation(ctx, testOrg, 1, CreateAutomationInput{
		Name:           "Daily Receivables Review Test",
		AutomationType: AutomationTypeDailyOverdueInvoiceReview,
		ScheduleType:   ScheduleTypeDaily,
		ScheduleTime:   "08:00",
		Timezone:       "UTC",
	})
	if err != nil {
		t.Fatalf("Failed to create automation: %v", err)
	}

	// Trigger manual run
	exec, err := svc.TriggerManualRun(ctx, testOrg, auto.ID, 1)
	if err != nil {
		t.Fatalf("TriggerManualRun failed: %v", err)
	}
	if exec.ID <= 0 {
		t.Fatalf("Expected valid execution ID")
	}
	if exec.TriggerType != TriggerTypeManual {
		t.Errorf("Expected trigger type MANUAL, got %s", exec.TriggerType)
	}

	// Wait for synchronous / background completion
	time.Sleep(1500 * time.Millisecond)

	completedExec, err := svc.GetExecution(ctx, testOrg, exec.ID)
	if err != nil {
		t.Fatalf("Failed to fetch execution: %v", err)
	}
	if completedExec.Status != ExecutionStatusCompleted {
		t.Errorf("Expected status COMPLETED, got %s (err: %v)", completedExec.Status, completedExec.ErrorMessage)
	}
	if completedExec.RecordsReviewed == 0 {
		t.Errorf("Expected records reviewed > 0")
	}

	// Verify recommendation linkage
	recs, err := svc.GetExecutionRecommendations(ctx, testOrg, exec.ID)
	if err != nil {
		t.Fatalf("Failed to get execution recommendations: %v", err)
	}
	if len(recs) == 0 {
		t.Fatalf("Expected at least 1 recommendation linked to execution #%d", exec.ID)
	}

	rec := recs[0]
	if rec.AutomationID == nil || *rec.AutomationID != auto.ID {
		t.Errorf("Expected rec.AutomationID to be %d, got %v", auto.ID, rec.AutomationID)
	}
	if rec.ExecutionID == nil || *rec.ExecutionID != exec.ID {
		t.Errorf("Expected rec.ExecutionID to be %d, got %v", exec.ID, rec.ExecutionID)
	}

	// Verify re-running deduplicates and doesn't explode recommendation count
	exec2, err := svc.TriggerManualRun(ctx, testOrg, auto.ID, 1)
	if err != nil {
		t.Fatalf("Second manual run failed: %v", err)
	}
	time.Sleep(1500 * time.Millisecond)

	completedExec2, err := svc.GetExecution(ctx, testOrg, exec2.ID)
	if err != nil {
		t.Fatalf("Failed to fetch second execution: %v", err)
	}
	if completedExec2.RecommendationsUpdated == 0 && completedExec2.RecommendationsCreated == 0 {
		t.Errorf("Expected recommendations to be updated or created in run 2")
	}
}

func TestSchedulerPollingAndExecution(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	testOrg := int64(88904)

	ensureTestOrg(ctx, db, testOrg)
	defer cleanupTestOrg(ctx, db, testOrg)

	repo := NewRepository(db)
	auditRepo := auditRepoPkg.NewMySQLRepository(db)
	auditSvc := auditSvcPkg.NewService(auditRepo, db)
	recRepo := recommendations.NewRepository(db)
	recGen := recommendations.NewGenerator(db, recRepo)
	svc := NewService(repo, recGen, recRepo, auditSvc)

	// Create automation due in the past
	auto, err := svc.CreateAutomation(ctx, testOrg, 1, CreateAutomationInput{
		Name:           "Due Scheduled Automation Test",
		AutomationType: AutomationTypeDailyContractDocumentExpiryReview,
		ScheduleType:   ScheduleTypeDaily,
		ScheduleTime:   "08:00",
		Timezone:       "UTC",
	})
	if err != nil {
		t.Fatalf("Failed to create automation: %v", err)
	}

	// Set next_execution_at to past
	pastTime := time.Now().Add(-10 * time.Minute)
	err = repo.UpdateExecutionTimestamps(ctx, auto.ID, time.Time{}, &pastTime, "", "")
	if err != nil {
		t.Fatalf("Failed to set past execution time: %v", err)
	}

	// Query due automations
	due, err := repo.GetDueAutomations(ctx, 10)
	if err != nil {
		t.Fatalf("GetDueAutomations failed: %v", err)
	}

	var found bool
	for _, a := range due {
		if a.ID == auto.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Expected automation #%d to be due", auto.ID)
	}

	// Instantiate and run single dispatch
	sched := NewScheduler(repo, svc, 1*time.Second)
	sched.(*scheduler).checkAndExecute()

	// Wait and verify execution was recorded and next run advanced
	time.Sleep(1 * time.Second)

	refreshedAuto, err := svc.GetAutomation(ctx, testOrg, auto.ID)
	if err != nil {
		t.Fatalf("Failed to fetch refreshed automation: %v", err)
	}
	if refreshedAuto.LastExecutionAt == nil {
		t.Errorf("Expected LastExecutionAt to be set")
	}
	if refreshedAuto.NextExecutionAt == nil || !refreshedAuto.NextExecutionAt.After(time.Now()) {
		t.Errorf("Expected NextExecutionAt to be in the future, got %v", refreshedAuto.NextExecutionAt)
	}
}

func TestCancelExecution(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	testOrg := int64(88905)

	ensureTestOrg(ctx, db, testOrg)
	defer cleanupTestOrg(ctx, db, testOrg)

	repo := NewRepository(db)
	auditRepo := auditRepoPkg.NewMySQLRepository(db)
	auditSvc := auditSvcPkg.NewService(auditRepo, db)
	recRepo := recommendations.NewRepository(db)
	recGen := recommendations.NewGenerator(db, recRepo)
	svc := NewService(repo, recGen, recRepo, auditSvc)

	auto, err := svc.CreateAutomation(ctx, testOrg, 1, CreateAutomationInput{
		Name:           "Cancel Test Automation",
		AutomationType: AutomationTypeShipmentExceptionReview,
		ScheduleType:   ScheduleTypeDaily,
	})
	if err != nil {
		t.Fatalf("Failed to create automation: %v", err)
	}

	// Create a queued execution
	exec := &AutomationExecution{
		AutomationID: auto.ID,
		OrgID:        testOrg,
		CorrelationID: "test-cancel-corr",
		TriggerType:  TriggerTypeScheduled,
		Status:       ExecutionStatusQueued,
	}
	createdExec, err := repo.CreateExecution(ctx, exec)
	if err != nil {
		t.Fatalf("Failed to create execution: %v", err)
	}

	// Cancel execution
	cancelled, err := svc.CancelExecution(ctx, testOrg, createdExec.ID, 1, "Testing cancellation")
	if err != nil {
		t.Fatalf("Failed to cancel execution: %v", err)
	}
	if cancelled.Status != ExecutionStatusCancelled {
		t.Errorf("Expected status CANCELLED, got %s", cancelled.Status)
	}
	if cancelled.ErrorMessage == nil || *cancelled.ErrorMessage != "Testing cancellation" {
		t.Errorf("Expected error message 'Testing cancellation', got %v", cancelled.ErrorMessage)
	}
}

func TestPhase3OperationalIntelligenceAndLifecycle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	testOrg := int64(88905)
	otherOrg := int64(88906)

	ensureTestOrg(ctx, db, testOrg)
	ensureTestOrg(ctx, db, otherOrg)
	defer cleanupTestOrg(ctx, db, testOrg)
	defer cleanupTestOrg(ctx, db, otherOrg)

	repo := NewRepository(db)
	auditRepo := auditRepoPkg.NewMySQLRepository(db)
	auditSvc := auditSvcPkg.NewService(auditRepo, db)
	recRepo := recommendations.NewRepository(db)
	recGen := recommendations.NewGenerator(db, recRepo)
	svc := NewService(repo, recGen, recRepo, auditSvc)

	// 1. Create Phase 3 event-driven automation with approval gating
	auto, err := svc.CreateAutomation(ctx, testOrg, 1, CreateAutomationInput{
		Name:           "Shipment Exception Immediate Detector",
		AutomationType: AutomationTypeShipmentExceptionReview,
		TriggerType:    TriggerTypeShipmentExceptionDetected,
		Scope:          "ORGANIZATION",
		ApprovalPolicy: ApprovalPolicyAlwaysRequire,
		OwnerTeam:      "OPERATIONS",
		Priority:       "HIGH",
		ScheduleType:   ScheduleTypeDaily,
		ScheduleTime:   "08:00",
		TargetModules:  []string{"shipments", "exceptions"},
		AllowedActions: []string{"INVESTIGATE_EXCEPTION", "ESCALATE_OPERATIONAL_RISK"},
	})
	if err != nil {
		t.Fatalf("Failed to create Phase 3 automation: %v", err)
	}
	if auto.TriggerType != TriggerTypeShipmentExceptionDetected {
		t.Errorf("Expected trigger type %s, got %s", TriggerTypeShipmentExceptionDetected, auto.TriggerType)
	}
	if auto.ApprovalPolicy != ApprovalPolicyAlwaysRequire {
		t.Errorf("Expected approval policy %s, got %s", ApprovalPolicyAlwaysRequire, auto.ApprovalPolicy)
	}

	// 2. Test EvaluateBusinessEvent with idempotency
	idempotencyKey := fmt.Sprintf("test-evt-idemp-%d", time.Now().UnixNano())
	evalRes, err := svc.EvaluateBusinessEvent(ctx, testOrg, 1, EvaluateEventInput{
		EventType:       TriggerTypeShipmentExceptionDetected,
		SourceModule:    "shipments",
		SourceRecordID:  101,
		SourceRecordRef: "SHP-101",
		Payload:         []byte(`{"severity":"CRITICAL","delay_hours":48,"exception_type":"CUSTOMS_HOLD"}`),
		IdempotencyKey:  idempotencyKey,
	})
	if err != nil {
		t.Fatalf("Failed to evaluate business event: %v", err)
	}
	if !evalRes.EventProcessed {
		t.Fatalf("Expected event to be processed")
	}
	if evalRes.SuppressedDuplicate {
		t.Errorf("First event should not be marked as duplicate")
	}
	if len(evalRes.InsightsCreated) == 0 {
		t.Fatalf("Expected operational insight to be created")
	}
	insight := evalRes.InsightsCreated[0]
	if insight.SourceModule != "shipments" || insight.SourceRecordID != 101 {
		t.Errorf("Unexpected insight source: module %s, id %d", insight.SourceModule, insight.SourceRecordID)
	}
	if !insight.IsApprovalRequired {
		t.Errorf("Expected approval to be required for high risk signal")
	}

	// 3. Test Duplicate Event Suppression via Idempotency Key
	dupRes, err := svc.EvaluateBusinessEvent(ctx, testOrg, 1, EvaluateEventInput{
		EventType:       TriggerTypeShipmentExceptionDetected,
		SourceModule:    "shipments",
		SourceRecordID:  101,
		SourceRecordRef: "SHP-101",
		IdempotencyKey:  idempotencyKey,
	})
	if err != nil {
		t.Fatalf("Failed to evaluate duplicate event: %v", err)
	}
	if !dupRes.SuppressedDuplicate {
		t.Errorf("Expected duplicate event to be suppressed by idempotency key")
	}

	// 4. Test Listing and Filtering Operational Insights
	insights, total, err := svc.ListInsights(ctx, testOrg, OperationalInsightFilter{
		SourceModule: "shipments",
		Status:       InsightStatusActive,
	})
	if err != nil {
		t.Fatalf("Failed to list insights: %v", err)
	}
	if total < 1 || len(insights) < 1 {
		t.Fatalf("Expected at least 1 insight, got total=%d, len=%d", total, len(insights))
	}

	// 5. Test Acknowledge and Dismiss Insight Lifecycle
	ackInsight, err := svc.AcknowledgeInsight(ctx, testOrg, insight.ID, 1)
	if err != nil {
		t.Fatalf("Failed to acknowledge insight: %v", err)
	}
	if ackInsight.Status != InsightStatusAcknowledged {
		t.Errorf("Expected status %s, got %s", InsightStatusAcknowledged, ackInsight.Status)
	}

	dismInsight, err := svc.DismissInsight(ctx, testOrg, insight.ID, 1, "Resolved by operations coordinator")
	if err != nil {
		t.Fatalf("Failed to dismiss insight: %v", err)
	}
	if dismInsight.Status != InsightStatusDismissed {
		t.Errorf("Expected status %s, got %s", InsightStatusDismissed, dismInsight.Status)
	}

	// 6. Test Tenant Isolation for Insights
	_, err = svc.GetInsight(ctx, otherOrg, insight.ID)
	if err == nil {
		t.Errorf("Expected cross-tenant access to fail with ErrInsightNotFound")
	}

	// 7. Test Execution Retry Lifecycle
	failedExec := &AutomationExecution{
		AutomationID: auto.ID,
		OrgID:        testOrg,
		CorrelationID: "test-failed-retry",
		TriggerType:  TriggerTypeManual,
		Status:       ExecutionStatusFailed,
		RetryCount:   0,
	}
	createdFailed, err := repo.CreateExecution(ctx, failedExec)
	if err != nil {
		t.Fatalf("Failed to create failed execution: %v", err)
	}

	retriedExec, err := svc.RetryExecution(ctx, testOrg, createdFailed.ID, 1)
	if err != nil {
		t.Fatalf("Failed to retry execution: %v", err)
	}
	if retriedExec.RetryCount != 1 {
		t.Errorf("Expected retry count 1, got %d", retriedExec.RetryCount)
	}

	// 8. Test Stats include active insights and waiting approval
	stats, err := svc.GetStats(ctx, testOrg)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}
	if stats.TotalAutomations < 1 {
		t.Errorf("Expected at least 1 automation in stats")
	}
}

