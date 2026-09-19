package approvals_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/freel/backend/internal/approvals"
)

// mockRepo implements approvals.Repository for unit tests
type mockRepo struct {
	items     map[int64]*approvals.ApprovalRequest
	decisions []*approvals.ApprovalDecision
	nextID    int64
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		items:     make(map[int64]*approvals.ApprovalRequest),
		decisions: make([]*approvals.ApprovalDecision, 0),
		nextID:    1,
	}
}

func (m *mockRepo) GetApprovalsByOrg(ctx context.Context, orgID int64) ([]*approvals.ApprovalRequest, error) {
	var list []*approvals.ApprovalRequest
	for _, req := range m.items {
		if req.OrgID == orgID {
			list = append(list, req)
		}
	}
	return list, nil
}

func (m *mockRepo) GetApprovalByID(ctx context.Context, orgID int64, id int64) (*approvals.ApprovalRequest, error) {
	req, ok := m.items[id]
	if !ok || req.OrgID != orgID {
		return nil, fmt.Errorf("approval request %d not found", id)
	}
	return req, nil
}

func (m *mockRepo) GetApprovalByCode(ctx context.Context, orgID int64, code string) (*approvals.ApprovalRequest, error) {
	for _, req := range m.items {
		if req.OrgID == orgID && req.RequestCode == code {
			return req, nil
		}
	}
	return nil, fmt.Errorf("approval request with code %s not found", code)
}

func (m *mockRepo) GetApprovalByReference(ctx context.Context, orgID int64, ref string) (*approvals.ApprovalRequest, error) {
	for _, req := range m.items {
		if req.OrgID == orgID && req.RelatedRef != nil && *req.RelatedRef == ref {
			return req, nil
		}
	}
	return nil, fmt.Errorf("approval request with ref %s not found", ref)
}

func (m *mockRepo) GetPendingApprovalByActionAndThread(ctx context.Context, orgID int64, actionName string, threadID string) (*approvals.ApprovalRequest, error) {
	return nil, fmt.Errorf("not found")
}

func (m *mockRepo) CreateApproval(ctx context.Context, req *approvals.ApprovalRequest) error {
	req.ID = m.nextID
	m.nextID++
	m.items[req.ID] = req
	return nil
}

func (m *mockRepo) UpdateApprovalStatus(ctx context.Context, orgID int64, id int64, status string, actorName string, notes string, reason string) (*approvals.ApprovalRequest, error) {
	req, err := m.GetApprovalByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	req.Status = status
	now := time.Now()
	if status == approvals.StatusApproved {
		req.ApprovedBy = &actorName
		req.ApprovedAt = &now
		if notes != "" {
			req.Comments = &notes
		}
	} else if status == approvals.StatusRejected {
		req.RejectedBy = &actorName
		req.RejectedAt = &now
		if reason != "" {
			req.RejectionReason = &reason
		}
	}
	return req, nil
}

func (m *mockRepo) CancelApproval(ctx context.Context, orgID int64, id int64, actorName string, notes string) (*approvals.ApprovalRequest, error) {
	req, err := m.GetApprovalByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	req.Status = approvals.StatusCancelled
	return req, nil
}

func (m *mockRepo) ReturnForChanges(ctx context.Context, orgID int64, id int64, actorName string, reason string, notes string) (*approvals.ApprovalRequest, error) {
	req, err := m.GetApprovalByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	req.Status = approvals.StatusReturnedForChanges
	now := time.Now()
	req.ReturnedBy = &actorName
	req.ReturnedAt = &now
	req.ReturnedReason = &reason
	return req, nil
}

func (m *mockRepo) ExpireStaleApprovals(ctx context.Context, orgID int64) (int64, error) {
	return 0, nil
}

func (m *mockRepo) GetApprovalStats(ctx context.Context, orgID int64) (*approvals.ApprovalStats, error) {
	return &approvals.ApprovalStats{}, nil
}

func (m *mockRepo) ResolveUserName(ctx context.Context, userID int64) (string, error) {
	return fmt.Sprintf("User_%d", userID), nil
}

func (m *mockRepo) RecordDecision(ctx context.Context, orgID int64, approvalID int64, actionName *string, decision string, actorID *int64, actorName string, reason *string, notes *string, correlationID *string) error {
	m.decisions = append(m.decisions, &approvals.ApprovalDecision{
		ID:         int64(len(m.decisions) + 1),
		OrgID:      orgID,
		ApprovalID: approvalID,
		Decision:   decision,
		ActorName:  actorName,
		CreatedAt:  time.Now(),
	})
	return nil
}

func (m *mockRepo) GetDecisionHistory(ctx context.Context, orgID int64, approvalID int64) ([]*approvals.ApprovalDecision, error) {
	var list []*approvals.ApprovalDecision
	for _, d := range m.decisions {
		if d.OrgID == orgID && d.ApprovalID == approvalID {
			list = append(list, d)
		}
	}
	return list, nil
}

func (m *mockRepo) UpdateExecutionState(ctx context.Context, orgID int64, id int64, status string, result *string, errStr *string) error {
	req, ok := m.items[id]
	if ok && req.OrgID == orgID {
		req.ExecutionStatus = status
	}
	return nil
}

// Tests
func TestApprovalCreation(t *testing.T) {
	repo := newMockRepo()
	svc := approvals.NewService(repo)

	input := &approvals.CreateApprovalInput{
		Category:    "OPERATIONS",
		Type:        "SHIPMENT_EXCEPTION",
		Title:       "Port Congestion Delay Approval",
		Description: "Request to update delivery milestone due to severe port congestion",
		RiskLevel:   "HIGH_RISK",
	}

	app, err := svc.CreateApproval(context.Background(), 1, input, "Dispatcher John")
	if err != nil {
		t.Fatalf("unexpected error creating approval: %v", err)
	}

	if app.ID == 0 {
		t.Errorf("expected generated ID, got 0")
	}
	if app.Status != approvals.StatusPendingApproval {
		t.Errorf("expected status %s, got %s", approvals.StatusPendingApproval, app.Status)
	}
	if !strings.Contains(app.RequestCode, "-APP-") {
		t.Errorf("expected request code to contain -APP-, got %s", app.RequestCode)
	}
}

func TestApproveRequest_SeparationOfDuties(t *testing.T) {
	repo := newMockRepo()
	svc := approvals.NewService(repo)

	requester := "Alice Manager"
	input := &approvals.CreateApprovalInput{
		Category:    "FINANCE",
		Type:        "PAYMENT_RECONCILIATION",
		Title:       "High-Value Invoice Adjustment",
		Description: "Reconcile payment of $50,000",
		RiskLevel:   "HIGH_RISK",
	}

	app, err := svc.CreateApproval(context.Background(), 1, input, requester)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Requester attempts to self-approve high-risk action
	_, err = svc.ApproveRequest(context.Background(), 1, app.ID, requester, 101, "Self approving")
	if err == nil {
		t.Fatalf("expected separation of duties error, got nil")
	}

	if !strings.Contains(err.Error(), "separation of duties") {
		t.Errorf("expected separation of duties message, got %v", err)
	}

	// Another user approves successfully
	approver := "Bob Director"
	approved, err := svc.ApproveRequest(context.Background(), 1, app.ID, approver, 102, "Looks good and verified")
	if err != nil {
		t.Fatalf("unexpected error when approver is different: %v", err)
	}

	if approved.Status != approvals.StatusApproved {
		t.Errorf("expected status APPROVED, got %s", approved.Status)
	}
	if approved.ApprovedBy == nil || *approved.ApprovedBy != approver {
		t.Errorf("expected ApprovedBy to be %s", approver)
	}
}

func TestRejectRequest_ReasonRequired(t *testing.T) {
	repo := newMockRepo()
	svc := approvals.NewService(repo)

	app, err := svc.CreateApproval(context.Background(), 1, &approvals.CreateApprovalInput{
		Category: "SALES",
		Type:     "QUOTE_APPROVAL",
		Title:    "Special discount quote",
	}, "Sales Rep")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Attempt reject without reason
	_, err = svc.RejectRequest(context.Background(), 1, app.ID, "Manager", 201, "", "")
	if err == nil {
		t.Fatalf("expected rejection reason error, got nil")
	}

	// Reject with valid reason
	rejected, err := svc.RejectRequest(context.Background(), 1, app.ID, "Manager", 201, "Discount exceeds allowable margin", "Contact customer")
	if err != nil {
		t.Fatalf("unexpected error rejecting: %v", err)
	}

	if rejected.Status != approvals.StatusRejected {
		t.Errorf("expected status REJECTED, got %s", rejected.Status)
	}
	if rejected.RejectionReason == nil || *rejected.RejectionReason != "Discount exceeds allowable margin" {
		t.Errorf("expected rejection reason to match")
	}
}

func TestReturnRequest_Lifecycle(t *testing.T) {
	repo := newMockRepo()
	svc := approvals.NewService(repo)

	app, err := svc.CreateApproval(context.Background(), 1, &approvals.CreateApprovalInput{
		Category: "CONTRACTS",
		Type:     "CONTRACT_REVISION",
		Title:    "Annual Carrier Agreement",
	}, "Legal Paralegal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Return without reason
	_, err = svc.ReturnRequest(context.Background(), 1, app.ID, "General Counsel", 301, "", "")
	if err == nil {
		t.Fatalf("expected return reason error, got nil")
	}

	// Return with reason
	returned, err := svc.ReturnRequest(context.Background(), 1, app.ID, "General Counsel", 301, "Missing indemnification clause", "Please request revised terms")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if returned.Status != approvals.StatusReturnedForChanges {
		t.Errorf("expected status RETURNED_FOR_CHANGES, got %s", returned.Status)
	}
	if returned.ReturnedReason == nil || *returned.ReturnedReason != "Missing indemnification clause" {
		t.Errorf("expected returned reason to match")
	}
}

func TestApprovalRequirements(t *testing.T) {
	repo := newMockRepo()
	svc := approvals.NewService(repo)

	reqs, err := svc.GetApprovalRequirements(context.Background(), 1, "shipments.update_milestone")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reqs.RequiresConfirmation {
		t.Errorf("expected RequiresConfirmation to be true")
	}
	if reqs.RequiredPermission != "operations:update" {
		t.Errorf("expected required permission operations:update, got %s", reqs.RequiredPermission)
	}
	if !reqs.SeparationOfDuties {
		t.Errorf("expected SeparationOfDuties to be true")
	}
}

func TestGetActionPreview_EmailCommunication(t *testing.T) {
	repo := newMockRepo()
	svc := approvals.NewService(repo)

	actionName := "sales.send_clarification_email"
	payload := `{"recipient": "customer@acme.com", "subject": "Missing Hazmat Docs", "body": "Please provide MSDS sheet for shipment #104."}`
	evidence := "Shipment #104 contains Class 3 Flammable liquids without attached MSDS documentation"
	impact := "Will dispatch clarification email to consignee and hold container clearance"

	input := &approvals.CreateApprovalInput{
		Category:              "DOCUMENTS",
		Type:                  "AI_ACTION",
		Title:                 "Send Clarification Email for Hazmat",
		ActionName:            actionName,
		ProposedPayload:       payload,
		Evidence:              evidence,
		ImpactSummary:         impact,
		ExternalCommunication: true,
		IsReversible:          false,
		RiskLevel:             "HIGH_RISK",
	}

	app, err := svc.CreateApproval(context.Background(), 1, input, "AI Agent")
	if err != nil {
		t.Fatalf("unexpected error creating approval: %v", err)
	}

	preview, err := svc.GetActionPreview(context.Background(), 1, app.ID)
	if err != nil {
		t.Fatalf("unexpected error getting action preview: %v", err)
	}

	if preview.ActionName != actionName {
		t.Errorf("expected action name %s, got %s", actionName, preview.ActionName)
	}
	if !preview.ExternalCommunication {
		t.Errorf("expected ExternalCommunication to be true")
	}
	if preview.IsReversible {
		t.Errorf("expected IsReversible to be false")
	}
	if preview.MessagePreview == nil {
		t.Fatalf("expected message preview details, got nil")
	}
	if preview.MessagePreview.Recipient != "customer@acme.com" {
		t.Errorf("expected recipient customer@acme.com, got %s", preview.MessagePreview.Recipient)
	}
	if preview.MessagePreview.Subject != "Missing Hazmat Docs" {
		t.Errorf("expected subject 'Missing Hazmat Docs', got %s", preview.MessagePreview.Subject)
	}
}

func TestActionExecution_ViaActionExecutor(t *testing.T) {
	repo := newMockRepo()
	svc := approvals.NewService(repo)

	executedAction := ""
	var executedPayload []byte
	svc.SetActionExecutor(func(ctx context.Context, orgID int64, actionName string, input []byte, actorName string, userID int64, approvalRef string) (map[string]interface{}, error) {
		executedAction = actionName
		executedPayload = input
		return map[string]interface{}{"status": "applied", "rate_id": 402}, nil
	})

	actionName := "pricing.apply_selected_rate"
	payload := `{"rate_id": 402, "rfq_id": 88}`
	input := &approvals.CreateApprovalInput{
		Category:        "COMMERCIAL",
		Type:            "COMMERCIAL_APPROVAL",
		Title:           "Apply Spot Rate for RFQ-88",
		ActionName:      actionName,
		ProposedPayload: payload,
		RiskLevel:       "MEDIUM",
	}

	app, err := svc.CreateApproval(context.Background(), 1, input, "Pricing Specialist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Approver approves and triggers execution
	approved, err := svc.ApproveRequest(context.Background(), 1, app.ID, "Pricing Director", 501, "Approved best rate")
	if err != nil {
		t.Fatalf("unexpected error approving request: %v", err)
	}

	if approved.Status != approvals.StatusApproved {
		t.Errorf("expected status APPROVED, got %s", approved.Status)
	}
	if executedAction != actionName {
		t.Errorf("expected ActionExecutor to be invoked with %s, got %s", actionName, executedAction)
	}
	if string(executedPayload) != payload {
		t.Errorf("expected payload %s, got %s", payload, string(executedPayload))
	}
	if approved.ExecutionStatus != "COMPLETED" {
		t.Errorf("expected execution status COMPLETED, got %s", approved.ExecutionStatus)
	}
}

func TestCancelRequest_Restrictions(t *testing.T) {
	repo := newMockRepo()
	svc := approvals.NewService(repo)

	app, err := svc.CreateApproval(context.Background(), 1, &approvals.CreateApprovalInput{
		Category: "OPERATIONS",
		Type:     "OPERATIONS_APPROVAL",
		Title:    "Reroute Container",
	}, "Dispatcher")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First approval
	_, err = svc.ApproveRequest(context.Background(), 1, app.ID, "Lead Operator", 701, "Looks fine")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Cannot cancel an already approved approval
	_, err = svc.CancelRequest(context.Background(), 1, app.ID, "Lead Operator", 701, "Try cancel")
	if err == nil {
		t.Fatalf("expected error cancelling approved request, got nil")
	}
	if !strings.Contains(err.Error(), "already approved") {
		t.Errorf("expected 'already approved' error, got %v", err)
	}
}

func TestDecisionHistory_AuditLogTracking(t *testing.T) {
	repo := newMockRepo()
	svc := approvals.NewService(repo)

	app, err := svc.CreateApproval(context.Background(), 1, &approvals.CreateApprovalInput{
		Category: "FINANCE",
		Type:     "INVOICE_ADJUSTMENT",
		Title:    "Waive Demurrage Fee",
	}, "Account Manager")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First decision: Return for changes
	_, err = svc.ReturnRequest(context.Background(), 1, app.ID, "Finance Manager", 801, "Need formal carrier receipt", "Attach document")
	if err != nil {
		t.Fatalf("unexpected error on return: %v", err)
	}

	// Second decision: Approve
	_, err = svc.ApproveRequest(context.Background(), 1, app.ID, "Finance Director", 802, "Receipt verified")
	if err != nil {
		t.Fatalf("unexpected error on approve: %v", err)
	}

	// Verify complete history
	history, err := svc.GetDecisionHistory(context.Background(), 1, app.ID)
	if err != nil {
		t.Fatalf("unexpected error getting history: %v", err)
	}

	if len(history) != 2 {
		t.Fatalf("expected 2 decisions in history, got %d", len(history))
	}
	if history[0].Decision != "RETURN_FOR_CHANGES" {
		t.Errorf("expected first decision to be RETURN_FOR_CHANGES, got %s", history[0].Decision)
	}
	if history[1].Decision != "APPROVE" {
		t.Errorf("expected second decision to be APPROVE, got %s", history[1].Decision)
	}
}

