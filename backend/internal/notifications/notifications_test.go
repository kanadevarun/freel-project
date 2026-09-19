package notifications

import (
	"context"
	"fmt"
	"testing"
	"time"

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

func cleanupTestNotifications(ctx context.Context, db *sqlx.DB, orgID int64) {
	_, _ = db.ExecContext(ctx, "DELETE FROM notification_escalation_events WHERE org_id = ?", orgID)
	_, _ = db.ExecContext(ctx, "DELETE FROM notifications WHERE org_id = ?", orgID)
	_, _ = db.ExecContext(ctx, "DELETE FROM user_notification_preferences WHERE org_id = ?", orgID)
}

func TestDeduplicationHash(t *testing.T) {
	h1 := makeDedupHash(1, ModuleApprovals, "APPROVAL", "100", TypePendingApproval)
	h2 := makeDedupHash(1, ModuleApprovals, "APPROVAL", "100", TypePendingApproval)
	h3 := makeDedupHash(1, ModuleApprovals, "APPROVAL", "101", TypePendingApproval)
	h4 := makeDedupHash(2, ModuleApprovals, "APPROVAL", "100", TypePendingApproval)

	if h1 != h2 {
		t.Errorf("Expected hashes for identical inputs to match: %s != %s", h1, h2)
	}
	if h1 == h3 {
		t.Errorf("Expected different record IDs to have different hashes: %s == %s", h1, h3)
	}
	if h1 == h4 {
		t.Errorf("Expected different org IDs to have different hashes: %s == %s", h1, h4)
	}
}

func TestNotificationCRUDAndIsolation(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	testOrg1 := int64(8881)
	testOrg2 := int64(8882)

	cleanupTestNotifications(ctx, db, testOrg1)
	cleanupTestNotifications(ctx, db, testOrg2)
	defer cleanupTestNotifications(ctx, db, testOrg1)
	defer cleanupTestNotifications(ctx, db, testOrg2)

	// 1. Create notification in Org 1
	n1 := &Notification{
		OrgID:            testOrg1,
		SourceModule:     ModuleShipments,
		SourceRecordType: "SHIPMENT_ALERT",
		SourceRecordID:   "ALERT-999",
		NotificationType: TypeCriticalShipmentException,
		Title:            "Port Delay Detected",
		Message:          "Vessel arrival delayed by 48 hours.",
		Severity:         SeverityHigh,
		Priority:         PriorityHigh,
		Status:           StatusActive,
		ActionRequired:   true,
		DedupHash:        makeDedupHash(testOrg1, ModuleShipments, "SHIPMENT_ALERT", "ALERT-999", TypeCriticalShipmentException),
		CorrelationID:    "corr-test-1",
	}

	err := repo.Create(ctx, n1)
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}
	if n1.ID <= 0 {
		t.Fatalf("Expected valid notification ID, got %d", n1.ID)
	}

	// 2. Tenant isolation: Org 2 must see 0 notifications
	itemsOrg2, totalOrg2, err := repo.List(ctx, NotificationFilter{OrgID: testOrg2})
	if err != nil {
		t.Fatalf("List error for Org 2: %v", err)
	}
	if totalOrg2 != 0 || len(itemsOrg2) != 0 {
		t.Errorf("Expected 0 notifications for Org 2, got total=%d len=%d", totalOrg2, len(itemsOrg2))
	}

	// Org 1 must see 1 notification
	itemsOrg1, totalOrg1, err := repo.List(ctx, NotificationFilter{OrgID: testOrg1})
	if err != nil {
		t.Fatalf("List error for Org 1: %v", err)
	}
	if totalOrg1 != 1 || len(itemsOrg1) != 1 {
		t.Fatalf("Expected 1 notification for Org 1, got total=%d len=%d", totalOrg1, len(itemsOrg1))
	}

	// 3. Mark as Read
	err = repo.MarkRead(ctx, testOrg1, n1.ID, true)
	if err != nil {
		t.Fatalf("Failed to mark as read: %v", err)
	}

	readNotif, err := repo.GetByID(ctx, testOrg1, n1.ID)
	if err != nil || readNotif == nil {
		t.Fatalf("Failed to fetch notification: %v", err)
	}
	if !readNotif.IsRead || readNotif.ReadAt == nil {
		t.Errorf("Expected notification to be read with ReadAt timestamp")
	}

	// 4. Mark as Unread
	err = repo.MarkRead(ctx, testOrg1, n1.ID, false)
	if err != nil {
		t.Fatalf("Failed to mark as unread: %v", err)
	}

	unreadNotif, _ := repo.GetByID(ctx, testOrg1, n1.ID)
	if unreadNotif.IsRead || unreadNotif.ReadAt != nil {
		t.Errorf("Expected notification to be unread with nil ReadAt")
	}

	// 5. Unread Count & Stats
	count, err := repo.GetUnreadCount(ctx, testOrg1, 0, "SUPER_ADMIN")
	if err != nil || count != 1 {
		t.Errorf("Expected unread count 1, got %d (err: %v)", count, err)
	}

	stats, err := repo.GetStats(ctx, testOrg1, 0, "SUPER_ADMIN")
	if err != nil || stats == nil {
		t.Fatalf("Failed to get stats: %v", err)
	}
	if stats.Unread != 1 || stats.Total != 1 || stats.ActionRequired != 1 {
		t.Errorf("Unexpected stats: %+v", stats)
	}

	// 6. Dismissal
	err = repo.Dismiss(ctx, testOrg1, n1.ID)
	if err != nil {
		t.Fatalf("Failed to dismiss: %v", err)
	}

	// Dismissed item should be excluded from standard unread list
	countAfterDismiss, _ := repo.GetUnreadCount(ctx, testOrg1, 0, "SUPER_ADMIN")
	if countAfterDismiss != 0 {
		t.Errorf("Expected unread count 0 after dismiss, got %d", countAfterDismiss)
	}
}

func TestEscalationTransitionAndAudit(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	testOrg := int64(8883)

	cleanupTestNotifications(ctx, db, testOrg)
	defer cleanupTestNotifications(ctx, db, testOrg)

	n := &Notification{
		OrgID:            testOrg,
		SourceModule:     ModuleApprovals,
		SourceRecordType: "APPROVAL",
		SourceRecordID:   "REQ-777",
		NotificationType: TypePendingApproval,
		Title:            "Urgent Discount Approval",
		Message:          "Sales rep requested 25% discount.",
		Severity:         SeverityHigh,
		Priority:         PriorityHigh,
		Status:           StatusActive,
		ActionRequired:   true,
		DedupHash:        makeDedupHash(testOrg, ModuleApprovals, "APPROVAL", "REQ-777", TypePendingApproval),
		CorrelationID:    "corr-esc-test",
	}
	if err := repo.Create(ctx, n); err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}

	// Escalate to CRITICAL
	err := repo.Escalate(ctx, testOrg, n.ID, SeverityCritical, "Exceeded 48h SLA window", "TIME_THRESHOLD", 48, "corr-esc-test")
	if err != nil {
		t.Fatalf("Failed to escalate notification: %v", err)
	}

	// Verify notification state
	updated, err := repo.GetByID(ctx, testOrg, n.ID)
	if err != nil || updated == nil {
		t.Fatalf("Failed to fetch notification: %v", err)
	}
	if !updated.IsEscalated {
		t.Errorf("Expected is_escalated = true")
	}
	if updated.Severity != SeverityCritical {
		t.Errorf("Expected severity CRITICAL, got %s", updated.Severity)
	}
	if updated.EscalationLevel != 1 {
		t.Errorf("Expected escalation_level = 1, got %d", updated.EscalationLevel)
	}

	// Verify escalation audit log
	events, err := repo.ListEscalations(ctx, testOrg, 10)
	if err != nil {
		t.Fatalf("Failed to list escalations: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 escalation event, got %d", len(events))
	}
	ev := events[0]
	if ev.NotificationID != n.ID || ev.NewSeverity != SeverityCritical || ev.PreviousSeverity != SeverityHigh {
		t.Errorf("Unexpected escalation event record: %+v", ev)
	}
}

func TestNotificationPreferences(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	testOrg := int64(8884)
	testUser := int64(1234)

	cleanupTestNotifications(ctx, db, testOrg)
	defer cleanupTestNotifications(ctx, db, testOrg)

	// 1. Get default preferences
	prefs, err := repo.GetPreferences(ctx, testOrg, testUser)
	if err != nil || prefs == nil {
		t.Fatalf("Failed to get default preferences: %v", err)
	}
	if prefs.MinSeverity != SeverityInformational || !prefs.InAppEnabled {
		t.Errorf("Unexpected default preferences: %+v", prefs)
	}

	// 2. Update preferences
	prefs.MinSeverity = SeverityHigh
	prefs.AssignedOnly = true
	prefs.FinanceEnabled = false

	err = repo.UpdatePreferences(ctx, prefs)
	if err != nil {
		t.Fatalf("Failed to update preferences: %v", err)
	}

	// 3. Verify persisted preferences
	updated, err := repo.GetPreferences(ctx, testOrg, testUser)
	if err != nil || updated == nil {
		t.Fatalf("Failed to fetch updated preferences: %v", err)
	}
	if updated.MinSeverity != SeverityHigh || !updated.AssignedOnly || updated.FinanceEnabled {
		t.Errorf("Preferences not persisted correctly: %+v", updated)
	}
}

func TestEvaluationEngineDedup(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	engine := NewEngine(db, repo)
	ctx := context.Background()

	// Evaluate against existing real Org 1 data
	c1, err := engine.Evaluate(ctx, 1)
	if err != nil {
		t.Fatalf("First engine evaluation failed: %v", err)
	}

	// Second evaluation should deduplicate existing open items and create 0 new notifications
	c2, err := engine.Evaluate(ctx, 1)
	if err != nil {
		t.Fatalf("Second engine evaluation failed: %v", err)
	}
	if c2 != 0 {
		t.Errorf("Expected 0 new notifications on immediate re-evaluation due to deduplication, got %d", c2)
	}

	fmt.Printf("Engine generated %d notifications on initial run, %d on re-evaluation (deduplication verified)\n", c1, c2)
}

func TestAdvancedLifecycleAcknowledgeSnoozeAndEscalate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	testOrg1 := int64(8885)
	testOrg2 := int64(8886)
	testUser := int64(42)

	cleanupTestNotifications(ctx, db, testOrg1)
	cleanupTestNotifications(ctx, db, testOrg2)
	defer cleanupTestNotifications(ctx, db, testOrg1)
	defer cleanupTestNotifications(ctx, db, testOrg2)

	// 1. Create notification
	n := &Notification{
		OrgID:            testOrg1,
		SourceModule:     ModuleShipments,
		SourceRecordType: "SHIPMENT",
		SourceRecordID:   "SHP-8885",
		NotificationType: TypeCriticalShipmentException,
		Title:            "Port Congestion Delay",
		Message:          "Vessel berthed with 72-hour delay at transshipment hub.",
		Severity:         SeverityHigh,
		Priority:         PriorityHigh,
		Status:           StatusActive,
		ActionRequired:   true,
		DedupHash:        makeDedupHash(testOrg1, ModuleShipments, "SHIPMENT", "SHP-8885", TypeCriticalShipmentException),
		CorrelationID:    "corr-test-lifecycle-1",
	}
	if err := repo.Create(ctx, n); err != nil {
		t.Fatalf("Failed to create test notification: %v", err)
	}

	// 2. Test AI Details update
	summary := "Vessel delay at Singapore transshipment hub affecting onward distribution"
	reason := "Transshipment dwell exceeded 48h SLA threshold"
	groupKey := "SHP-DELAY-SIN"
	score := 85.5
	if err := repo.UpdateAIDetails(ctx, testOrg1, n.ID, summary, reason, groupKey, score); err != nil {
		t.Fatalf("Failed to update AI details: %v", err)
	}

	// Verify AI details stored
	loaded, err := repo.GetByID(ctx, testOrg1, n.ID)
	if err != nil || loaded == nil {
		t.Fatalf("Failed to fetch notification: %v", err)
	}
	if loaded.AISummary == nil || *loaded.AISummary != summary {
		t.Errorf("AI summary mismatch: got %v", loaded.AISummary)
	}
	if loaded.GroupKey == nil || *loaded.GroupKey != groupKey {
		t.Errorf("Group key mismatch: got %v", loaded.GroupKey)
	}
	if loaded.AIPriorityScore == nil || *loaded.AIPriorityScore != score {
		t.Errorf("AI priority score mismatch: got %v", loaded.AIPriorityScore)
	}

	// 3. Test Snooze
	snoozeUntil := time.Now().Add(4 * time.Hour)
	if err := repo.Snooze(ctx, testOrg1, n.ID, snoozeUntil); err != nil {
		t.Fatalf("Failed to snooze notification: %v", err)
	}

	loadedSnoozed, _ := repo.GetByID(ctx, testOrg1, n.ID)
	if !loadedSnoozed.IsSnoozed || loadedSnoozed.DeliveryStatus != "SNOOZED" {
		t.Errorf("Expected IsSnoozed=true and DeliveryStatus=SNOOZED, got %v, %s", loadedSnoozed.IsSnoozed, loadedSnoozed.DeliveryStatus)
	}

	// 4. Tenant isolation: Org 2 cannot acknowledge Org 1's notification
	err = repo.Acknowledge(ctx, testOrg2, n.ID, testUser)
	if err == nil {
		t.Errorf("Expected error when Org 2 attempts to acknowledge Org 1 notification")
	}

	// 5. Test Acknowledge in Org 1
	if err := repo.Acknowledge(ctx, testOrg1, n.ID, testUser); err != nil {
		t.Fatalf("Failed to acknowledge notification: %v", err)
	}

	loadedAck, _ := repo.GetByID(ctx, testOrg1, n.ID)
	if !loadedAck.IsAcknowledged || loadedAck.DeliveryStatus != "ACKNOWLEDGED" {
		t.Errorf("Expected IsAcknowledged=true and DeliveryStatus=ACKNOWLEDGED, got %v, %s", loadedAck.IsAcknowledged, loadedAck.DeliveryStatus)
	}
	if loadedAck.AcknowledgedBy == nil || *loadedAck.AcknowledgedBy != testUser {
		t.Errorf("Expected AcknowledgedBy=%d, got %v", testUser, loadedAck.AcknowledgedBy)
	}

	// 6. Test Escalation Event recording with AI fields
	aiEscSummary := "Escalated to Operations Lead due to critical supply chain SLA breach"
	recAction := "Re-route freight via express feeder carrier"
	approvalID := int64(9901)
	escEvent := &EscalationEvent{
		NotificationID:      n.ID,
		OrgID:               testOrg1,
		EscalationLevel:     2,
		EscalationReason:    "Exceeded 48h unacknowledged threshold",
		PreviousSeverity:    SeverityHigh,
		NewSeverity:         SeverityCritical,
		TriggerType:         "MANUAL",
		ThresholdHours:      48,
		AIEscalationSummary: &aiEscSummary,
		RecommendedAction:   &recAction,
		ApprovalID:          &approvalID,
		CorrelationID:       "corr-test-esc-1",
	}
	if _, err := repo.CreateEscalationEvent(ctx, escEvent); err != nil {
		t.Fatalf("Failed to record escalation event: %v", err)
	}

	events, err := repo.ListEscalations(ctx, testOrg1, 10)
	if err != nil || len(events) == 0 {
		t.Fatalf("Expected escalation events for Org 1, got %v, err=%v", len(events), err)
	}
	if events[0].ApprovalID == nil || *events[0].ApprovalID != approvalID {
		t.Errorf("Expected ApprovalID=%d, got %v", approvalID, events[0].ApprovalID)
	}
}

