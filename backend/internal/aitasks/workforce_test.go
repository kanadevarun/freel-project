package aitasks

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestWorkforceStatusContract(t *testing.T) {
	expectedStatuses := []string{
		WorkforceStatusQueued,
		WorkforceStatusClaimed,
		WorkforceStatusProcessing,
		WorkforceStatusWaitingForApproval,
		WorkforceStatusPaused,
		WorkforceStatusRetrying,
		WorkforceStatusCompleted,
		WorkforceStatusCompletedWithFailover,
		WorkforceStatusCompletedInMockMode,
		WorkforceStatusFailed,
		WorkforceStatusCancelled,
		WorkforceStatusStale,
		WorkforceStatusUnknown,
	}

	for _, st := range expectedStatuses {
		meta := GetStatusMetadata(st)
		if meta.Value != st {
			t.Errorf("expected value '%s', got '%s'", st, meta.Value)
		}
		if meta.Label == "" {
			t.Errorf("status '%s' has empty label", st)
		}
		if meta.Description == "" {
			t.Errorf("status '%s' has empty description", st)
		}
	}

	// Verify terminal status flags
	if !GetStatusMetadata(WorkforceStatusCompleted).IsTerminal {
		t.Errorf("expected completed to be terminal")
	}
	if !GetStatusMetadata(WorkforceStatusFailed).IsTerminal {
		t.Errorf("expected failed to be terminal")
	}
	if !GetStatusMetadata(WorkforceStatusCancelled).IsTerminal {
		t.Errorf("expected cancelled to be terminal")
	}
	if GetStatusMetadata(WorkforceStatusProcessing).IsTerminal {
		t.Errorf("processing must not be terminal")
	}

	// Verify actionable and retry flags
	if !GetStatusMetadata(WorkforceStatusFailed).RetryPermitted {
		t.Errorf("expected failed status to permit retry")
	}
	if !GetStatusMetadata(WorkforceStatusStale).RetryPermitted {
		t.Errorf("expected stale status to permit retry")
	}
	if GetStatusMetadata(WorkforceStatusCompleted).RetryPermitted {
		t.Errorf("completed status must not permit retry")
	}
	if !GetStatusMetadata(WorkforceStatusWaitingForApproval).ApprovalRequired {
		t.Errorf("waiting_for_approval must require approval")
	}
}

func TestResolveTaskWorkforceStatus(t *testing.T) {
	now := time.Now()
	expiredLease := now.Add(-5 * time.Minute)
	activeLease := now.Add(5 * time.Minute)
	workerID := "worker-01"

	tests := []struct {
		name       string
		task       *Task
		isFailover bool
		isMock     bool
		wantStatus string
	}{
		{
			name: "Queued Unclaimed",
			task: &Task{
				Status: StatusQueued,
			},
			wantStatus: WorkforceStatusQueued,
		},
		{
			name: "Queued Claimed",
			task: &Task{
				Status:   StatusQueued,
				WorkerID: &workerID,
			},
			wantStatus: WorkforceStatusClaimed,
		},
		{
			name: "Active Processing",
			task: &Task{
				Status:         StatusProcessing,
				LeaseExpiresAt: &activeLease,
			},
			wantStatus: WorkforceStatusProcessing,
		},
		{
			name: "Stale Processing (Expired Lease)",
			task: &Task{
				Status:         StatusProcessing,
				LeaseExpiresAt: &expiredLease,
			},
			wantStatus: WorkforceStatusStale,
		},
		{
			name: "Waiting for Sign-Off",
			task: &Task{
				Status: StatusWaitingForApproval,
			},
			wantStatus: WorkforceStatusWaitingForApproval,
		},
		{
			name: "Retrying Backoff",
			task: &Task{
				Status: StatusRetrying,
			},
			wantStatus: WorkforceStatusRetrying,
		},
		{
			name: "Standard Completed",
			task: &Task{
				Status: StatusCompleted,
			},
			wantStatus: WorkforceStatusCompleted,
		},
		{
			name: "Completed with Failover",
			task: &Task{
				Status: StatusCompleted,
			},
			isFailover: true,
			wantStatus: WorkforceStatusCompletedWithFailover,
		},
		{
			name: "Completed in Mock Mode",
			task: &Task{
				Status: StatusCompleted,
			},
			isMock:     true,
			wantStatus: WorkforceStatusCompletedInMockMode,
		},
		{
			name: "Permanent Failure",
			task: &Task{
				Status: StatusFailed,
			},
			wantStatus: WorkforceStatusFailed,
		},
		{
			name: "Cancelled Task",
			task: &Task{
				Status: StatusCancelled,
			},
			wantStatus: WorkforceStatusCancelled,
		},
		{
			name:       "Nil Task",
			task:       nil,
			wantStatus: WorkforceStatusUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st, meta := ResolveTaskWorkforceStatus(tt.task, tt.isFailover, tt.isMock)
			if st != tt.wantStatus {
				t.Errorf("got status '%s', want '%s'", st, tt.wantStatus)
			}
			if meta.Value != tt.wantStatus {
				t.Errorf("got meta value '%s', want '%s'", meta.Value, tt.wantStatus)
			}
		})
	}
}

func TestMapTaskTypeToAgent(t *testing.T) {
	tests := []struct {
		taskType  string
		wantAgent string
		wantMod   string
	}{
		{"PRICING_ANALYZE", "pricing", "PRICING"},
		{"PRICING_RESUME", "pricing", "PRICING"},
		{"EMAIL_PARSE", "sales", "SALES"},
		{"SALES_INBOUND", "sales", "SALES"},
		{"CARRIER_UPDATE_PARSE", "operations", "OPERATIONS"},
		{"OPERATIONS_TRACKING", "operations", "OPERATIONS"},
		{"DOC_PROCESS", "contracts", "CONTRACTS"},
		{"CONTRACT_PARSE", "contracts", "CONTRACTS"},
		{"DOC_VERIFY", "compliance", "COMPLIANCE"},
		{"COMPLIANCE_AUDIT", "compliance", "COMPLIANCE"},
		{"BILL_RECONCILE", "finance", "FINANCE"},
		{"INVOICE_AUDIT", "finance", "FINANCE"},
		{"LEAD_SCORING", "leads", "LEADS"},
		{"OUTREACH_GEN", "outreach", "OUTREACH"},
	}

	for _, tt := range tests {
		agentKey, _, mod, _ := MapTaskTypeToAgent(tt.taskType)
		if agentKey != tt.wantAgent {
			t.Errorf("taskType '%s': got agent '%s', want '%s'", tt.taskType, agentKey, tt.wantAgent)
		}
		if mod != tt.wantMod {
			t.Errorf("taskType '%s': got module '%s', want '%s'", tt.taskType, mod, tt.wantMod)
		}
	}
}

func TestResolveRelatedRef(t *testing.T) {
	rfqType := "rfq"
	rfqID := "102"
	task := &Task{
		ID:         55,
		EntityType: &rfqType,
		EntityID:   &rfqID,
		TaskType:   "PRICING_ANALYZE",
	}

	label, id, nav := ResolveRelatedRef(task)
	if label != "RFQ #102" {
		t.Errorf("got label '%s', want 'RFQ #102'", label)
	}
	if id != "102" {
		t.Errorf("got id '%s', want '102'", id)
	}
	if nav != "/dashboard/rfqs/102" {
		t.Errorf("got nav '%s', want '/dashboard/rfqs/102'", nav)
	}
}

func TestGetWorkforceSummary_MockDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)
	svc := NewService(repo)

	orgID := int64(2)

	// 1. Mock task counts query
	countRows := sqlmock.NewRows([]string{
		"total", "queued", "processing", "stale", "waiting_for_approval",
		"retrying", "failed", "completed_24h", "last_heartbeat",
	}).AddRow(10, 2, 3, 1, 1, 1, 1, 2, time.Now())
	mock.ExpectQuery("SELECT(.*)FROM ai_processing_tasks(.*)WHERE org_id = ?").
		WithArgs(orgID).
		WillReturnRows(countRows)

	// 2. Mock execution traces query
	traceRows := sqlmock.NewRows([]string{
		"failover_count", "mock_count", "avg_duration",
	}).AddRow(1, 0, 1850.5)
	mock.ExpectQuery("SELECT(.*)FROM ai_execution_traces(.*)WHERE org_id = ?").
		WithArgs(orgID).
		WillReturnRows(traceRows)

	// 3. Mock agent breakdown query
	agentRows := sqlmock.NewRows([]string{
		"task_type", "active_count", "completed_24h", "failed_24h", "waiting_approvals",
	}).
		AddRow("PRICING_ANALYZE", 2, 1, 0, 1).
		AddRow("EMAIL_PARSE", 1, 1, 0, 0)
	mock.ExpectQuery("SELECT(.*)FROM ai_processing_tasks(.*)GROUP BY task_type").
		WithArgs(orgID).
		WillReturnRows(agentRows)

	// 4. Mock failures query
	failRows := sqlmock.NewRows([]string{
		"category", "cnt",
	}).AddRow("provider_rate_limit", 1)
	mock.ExpectQuery("SELECT(.*)FROM ai_execution_traces(.*)GROUP BY category").
		WithArgs(orgID).
		WillReturnRows(failRows)

	// 5. Mock check table
	chkRows := sqlmock.NewRows([]string{"1"}).AddRow(1)
	mock.ExpectQuery("SELECT 1 FROM ai_checkpoints LIMIT 1").
		WillReturnRows(chkRows)

	// 6. Mock stale check
	staleRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT(.*) FROM ai_processing_tasks WHERE status = 'PROCESSING'").
		WillReturnRows(staleRows)

	summary, err := svc.GetWorkforceSummary(context.Background(), orgID)
	if err != nil {
		t.Fatalf("GetWorkforceSummary failed: %v", err)
	}

	if summary.OrgID != orgID {
		t.Errorf("expected org_id %d, got %d", orgID, summary.OrgID)
	}
	if summary.QueuedTasks != 2 {
		t.Errorf("expected 2 queued tasks, got %d", summary.QueuedTasks)
	}
	if summary.ProcessingTasks != 3 {
		t.Errorf("expected 3 processing tasks, got %d", summary.ProcessingTasks)
	}
	if summary.CompletedWithFailover != 1 {
		t.Errorf("expected 1 failover completion, got %d", summary.CompletedWithFailover)
	}

	// Verify all 8 agents exist in map
	expectedAgents := []string{"pricing", "sales", "operations", "contracts", "compliance", "finance", "leads", "outreach"}
	for _, k := range expectedAgents {
		if _, exists := summary.ByAgent[k]; !exists {
			t.Errorf("agent '%s' missing from workforce summary", k)
		}
	}

	// Verify pricing agent metrics
	pricing := summary.ByAgent["pricing"]
	if pricing.ActiveTasks != 2 {
		t.Errorf("expected 2 active pricing tasks, got %d", pricing.ActiveTasks)
	}
	if pricing.Status != "ACTIVE" {
		t.Errorf("expected pricing status ACTIVE, got %s", pricing.Status)
	}
}

func TestListWorkforceTasks_IsolationAndFiltering(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)
	svc := NewService(repo)

	orgID := int64(42)
	filter := WorkforceTaskFilter{
		Limit:  10,
		Offset: 0,
	}

	entityType := "rfq"
	entityID := "999"
	errReason := "Model context window exceeded [REDACTED]"
	now := time.Now()

	columns := []string{
		"id", "org_id", "document_id", "entity_type", "entity_id", "task_type", "resource_id",
		"payload", "status", "error_message", "last_error_code", "retry_count", "max_retries",
		"worker_id", "lease_expires_at", "heartbeat_at", "started_at", "completed_at", "available_at",
		"thread_id", "correlation_id", "approval_id", "acting_user_id", "actor_type",
		"created_at", "updated_at",
	}

	rows := sqlmock.NewRows(columns).AddRow(
		101, orgID, nil, &entityType, &entityID, "PRICING_ANALYZE", nil,
		"{}", "FAILED", &errReason, "PROVIDER_ERR", 3, 3,
		"worker-1", nil, nil, nil, nil, nil,
		"thread-1", "corr-1", nil, nil, "SYSTEM",
		now, now,
	)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT(.*)FROM ai_processing_tasks(.*)WHERE org_id = ?").
		WithArgs(orgID).
		WillReturnRows(countRows)

	mock.ExpectQuery("SELECT(.*)FROM ai_processing_tasks(.*)WHERE org_id = ?").
		WithArgs(orgID, 10, 0).
		WillReturnRows(rows)

	tasks, total, err := svc.ListWorkforceTasks(context.Background(), orgID, filter)
	if err != nil {
		t.Fatalf("ListWorkforceTasks failed: %v", err)
	}

	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]
	if task.ID != 101 {
		t.Errorf("expected task id 101, got %d", task.ID)
	}
	if task.Status != WorkforceStatusFailed {
		t.Errorf("expected status %s, got %s", WorkforceStatusFailed, task.Status)
	}
	if !task.CanRetry {
		t.Errorf("expected failed task to be retriable")
	}
	if task.CanCancel {
		t.Errorf("failed task must not be cancellable")
	}
	if task.ErrorMessage == nil || *task.ErrorMessage != errReason {
		t.Errorf("expected safe error message '%s', got '%v'", errReason, task.ErrorMessage)
	}
}

func TestOperationalActionEligibility(t *testing.T) {
	// 1. Retry eligibility
	failedMeta := GetStatusMetadata(WorkforceStatusFailed)
	if !failedMeta.RetryPermitted {
		t.Errorf("failed status must permit retry")
	}
	staleMeta := GetStatusMetadata(WorkforceStatusStale)
	if !staleMeta.RetryPermitted {
		t.Errorf("stale status must permit retry")
	}
	completedMeta := GetStatusMetadata(WorkforceStatusCompleted)
	if completedMeta.RetryPermitted {
		t.Errorf("completed status must not permit retry")
	}

	// 2. Cancellation eligibility
	queuedMeta := GetStatusMetadata(WorkforceStatusQueued)
	if queuedMeta.IsTerminal || !queuedMeta.IsActionable {
		t.Errorf("queued status must be cancellable (non-terminal, actionable)")
	}
	processingMeta := GetStatusMetadata(WorkforceStatusProcessing)
	if processingMeta.IsTerminal || !processingMeta.IsActionable {
		t.Errorf("processing status must be cancellable")
	}
	cancelledMeta := GetStatusMetadata(WorkforceStatusCancelled)
	if !cancelledMeta.IsTerminal {
		t.Errorf("cancelled status must be terminal")
	}
}

