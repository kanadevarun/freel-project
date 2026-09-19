package event_workflows

import (
	"context"
	"testing"
	"time"
)

type mockRepository struct {
	events    map[string]*EventRecord
	workflows map[string]*WorkflowInstance
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		events:    make(map[string]*EventRecord),
		workflows: make(map[string]*WorkflowInstance),
	}
}

func (m *mockRepository) SaveEvent(ctx context.Context, e *EventRecord) (*EventRecord, error) {
	e.ID = int64(len(m.events) + 1)
	m.events[e.DedupKey] = e
	return e, nil
}

func (m *mockRepository) GetEventByDedupKey(ctx context.Context, orgID int64, dedupKey string) (*EventRecord, error) {
	e, ok := m.events[dedupKey]
	if ok && e.OrgID == orgID {
		return e, nil
	}
	return nil, nil
}

func (m *mockRepository) GetEventByID(ctx context.Context, orgID, eventID int64) (*EventRecord, error) {
	for _, e := range m.events {
		if e.ID == eventID && e.OrgID == orgID {
			return e, nil
		}
	}
	return nil, nil
}

func (m *mockRepository) ListEvents(ctx context.Context, orgID int64, limit, offset int) ([]*EventRecord, int, error) {
	var list []*EventRecord
	for _, e := range m.events {
		if e.OrgID == orgID {
			list = append(list, e)
		}
	}
	return list, len(list), nil
}

func (m *mockRepository) UpdateEventStatus(ctx context.Context, orgID, eventID int64, status string, failureReason *string) error {
	for _, e := range m.events {
		if e.ID == eventID && e.OrgID == orgID {
			e.Status = status
			e.FailureReason = failureReason
			now := time.Now()
			e.ProcessedAt = &now
			return nil
		}
	}
	return nil
}

func (m *mockRepository) SaveWorkflowInstance(ctx context.Context, w *WorkflowInstance) (*WorkflowInstance, error) {
	w.ID = int64(len(m.workflows) + 1)
	m.workflows[w.DedupKey] = w
	return w, nil
}

func (m *mockRepository) GetWorkflowInstanceByDedupKey(ctx context.Context, orgID int64, dedupKey string) (*WorkflowInstance, error) {
	w, ok := m.workflows[dedupKey]
	if ok && w.OrgID == orgID {
		return w, nil
	}
	return nil, nil
}

func (m *mockRepository) GetWorkflowInstanceByID(ctx context.Context, orgID, id int64) (*WorkflowInstance, error) {
	for _, w := range m.workflows {
		if w.ID == id && w.OrgID == orgID {
			return w, nil
		}
	}
	return nil, nil
}

func (m *mockRepository) ListWorkflowInstances(ctx context.Context, orgID int64, limit, offset int) ([]*WorkflowInstance, int, error) {
	var list []*WorkflowInstance
	for _, w := range m.workflows {
		if w.OrgID == orgID {
			list = append(list, w)
		}
	}
	return list, len(list), nil
}

func (m *mockRepository) UpdateWorkflowStatus(ctx context.Context, orgID, id int64, status, actionStatus string, approvalID *int64, errorMsg *string) error {
	for _, w := range m.workflows {
		if w.ID == id && w.OrgID == orgID {
			w.Status = status
			w.ActionStatus = actionStatus
			if approvalID != nil {
				w.ApprovalID = approvalID
			}
			w.ErrorMessage = errorMsg
			return nil
		}
	}
	return nil
}

func (m *mockRepository) GetCustomerContext(ctx context.Context, orgID, customerID int64) (map[string]interface{}, error) {
	return map[string]interface{}{"customer_tier": "ENTERPRISE"}, nil
}

func (m *mockRepository) GetShipmentContext(ctx context.Context, orgID, shipmentID int64) (map[string]interface{}, error) {
	return map[string]interface{}{"status": "IN_TRANSIT", "booking_number": "BKG-101"}, nil
}

func (m *mockRepository) GetInvoiceContext(ctx context.Context, orgID, invoiceID int64) (map[string]interface{}, error) {
	return map[string]interface{}{"outstanding_amount": 15000.0, "days_overdue": 20}, nil
}

func (m *mockRepository) GetContractContext(ctx context.Context, orgID, contractID int64) (map[string]interface{}, error) {
	return map[string]interface{}{"party_name": "Apex Drayage", "status": "ACTIVE"}, nil
}

// ── Tests ───────────────────────────────────────────────────────────────────

func TestEligibilityEvaluation(t *testing.T) {
	svc := &service{}

	tests := []struct {
		eventType  string
		expectedOk bool
		expectedWf string
	}{
		{"shipment.milestone_missed", true, "SHIPMENT_EXCEPTION_RESPONSE"},
		{"shipment.exception_created", true, "SHIPMENT_EXCEPTION_RESPONSE"},
		{"invoice.overdue", true, "INVOICE_COLLECTION_ESCALATION"},
		{"contract.expiring", true, "CONTRACT_COMPLIANCE_RENEWAL"},
		{"rfq.submitted", true, "RFQ_QUOTATION_DISPATCH"},
		{"lead.created", true, "LEAD_FOLLOWUP"},
		{"unsupported.event.type", false, ""},
	}

	for _, tt := range tests {
		ok, wf := svc.evaluateEligibility(tt.eventType, "RECORD")
		if ok != tt.expectedOk {
			t.Errorf("evaluateEligibility(%q) ok = %v, expected %v", tt.eventType, ok, tt.expectedOk)
		}
		if wf != tt.expectedWf {
			t.Errorf("evaluateEligibility(%q) wf = %q, expected %q", tt.eventType, wf, tt.expectedWf)
		}
	}
}

func TestEventDeduplicationAndIdempotency(t *testing.T) {
	repo := newMockRepository()
	ctx := context.Background()

	event1 := &EventRecord{
		OrgID:            1,
		EventType:        "shipment.delayed",
		SourceModule:     "SHIPMENTS",
		SourceRecordType: "SHIPMENT",
		SourceRecordID:   "101",
		DedupKey:         "evt:1:shipment.delayed:SHIPMENT:101:corr-1",
		Status:           "RECEIVED",
	}
	_, err := repo.SaveEvent(ctx, event1)
	if err != nil {
		t.Fatalf("Failed to save event: %v", err)
	}

	found, err := repo.GetEventByDedupKey(ctx, 1, "evt:1:shipment.delayed:SHIPMENT:101:corr-1")
	if err != nil || found == nil {
		t.Fatalf("Expected to find event by dedup key")
	}

	// Different org with same dedup key should return nil (tenant isolation)
	crossFound, _ := repo.GetEventByDedupKey(ctx, 2, "evt:1:shipment.delayed:SHIPMENT:101:corr-1")
	if crossFound != nil {
		t.Errorf("Cross-tenant event leak detected")
	}
}

func TestWorkflowInstanceLifecycleState(t *testing.T) {
	repo := newMockRepository()
	ctx := context.Background()

	wf := &WorkflowInstance{
		OrgID:            1,
		EventID:          1,
		WorkflowType:     "SHIPMENT_EXCEPTION_RESPONSE",
		Status:           "RUNNING",
		TriggerEventType: "shipment.milestone_missed",
		SourceModule:     "SHIPMENTS",
		SourceRecordType: "SHIPMENT",
		SourceRecordID:   "101",
		ActionStatus:     "NOT_STARTED",
		DedupKey:         "wf:1:shipment.milestone_missed:101",
	}
	saved, err := repo.SaveWorkflowInstance(ctx, wf)
	if err != nil || saved.ID <= 0 {
		t.Fatalf("Failed to save workflow instance")
	}

	// Update to AWAITING_APPROVAL with approval ID
	var approvalID int64 = 42
	err = repo.UpdateWorkflowStatus(ctx, 1, saved.ID, "AWAITING_APPROVAL", "AWAITING_APPROVAL", &approvalID, nil)
	if err != nil {
		t.Fatalf("Failed to update workflow status: %v", err)
	}

	updated, err := repo.GetWorkflowInstanceByID(ctx, 1, saved.ID)
	if err != nil || updated == nil {
		t.Fatalf("Failed to retrieve updated workflow")
	}
	if updated.Status != "AWAITING_APPROVAL" || updated.ActionStatus != "AWAITING_APPROVAL" {
		t.Errorf("Status mismatch: got %s, %s", updated.Status, updated.ActionStatus)
	}
	if updated.ApprovalID == nil || *updated.ApprovalID != 42 {
		t.Errorf("ApprovalID mismatch: expected 42, got %v", updated.ApprovalID)
	}
}
