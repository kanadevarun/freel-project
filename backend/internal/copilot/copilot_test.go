package copilot

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock repository for service testing
type mockRepository struct {
	sessions       map[string]*CopilotSession
	messages       []*CopilotMessage
	actions        []*CopilotActionHistory
	approvalReqIDs []int64
	recIDs         []int64
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		sessions: make(map[string]*CopilotSession),
		messages: make([]*CopilotMessage, 0),
		actions:  make([]*CopilotActionHistory, 0),
	}
}

func (m *mockRepository) CreateSession(ctx context.Context, s *CopilotSession) error {
	m.sessions[s.SessionID] = s
	return nil
}

func (m *mockRepository) GetSession(ctx context.Context, orgID int64, sessionID string) (*CopilotSession, error) {
	if s, ok := m.sessions[sessionID]; ok && s.OrgID == orgID {
		return s, nil
	}
	return nil, nil
}

func (m *mockRepository) ListSessions(ctx context.Context, orgID int64, userID int64, limit int) ([]CopilotSession, error) {
	res := make([]CopilotSession, 0)
	for _, s := range m.sessions {
		if s.OrgID == orgID && s.UserID == userID {
			res = append(res, *s)
		}
	}
	return res, nil
}

func (m *mockRepository) ArchiveSession(ctx context.Context, orgID int64, sessionID string) error {
	if s, ok := m.sessions[sessionID]; ok && s.OrgID == orgID {
		s.IsArchived = true
	}
	return nil
}

func (m *mockRepository) UpdateSessionContext(ctx context.Context, orgID int64, sessionID, module, route string, recordID *string) error {
	if s, ok := m.sessions[sessionID]; ok && s.OrgID == orgID {
		s.CurrentModule = module
		s.CurrentRoute = route
		s.CurrentRecordID = recordID
	}
	return nil
}

func (m *mockRepository) SaveMessage(ctx context.Context, msg *CopilotMessage) error {
	m.messages = append(m.messages, msg)
	return nil
}

func (m *mockRepository) ListMessages(ctx context.Context, orgID int64, sessionID string, limit int) ([]CopilotMessage, error) {
	res := make([]CopilotMessage, 0)
	for _, msg := range m.messages {
		if msg.OrgID == orgID && msg.SessionID == sessionID {
			res = append(res, *msg)
		}
	}
	return res, nil
}

func (m *mockRepository) LogAction(ctx context.Context, a *CopilotActionHistory) error {
	a.ID = int64(len(m.actions) + 1)
	m.actions = append(m.actions, a)
	return nil
}

func (m *mockRepository) UpdateActionStatus(ctx context.Context, orgID int64, id int64, status string, resultSummary *string) error {
	for _, a := range m.actions {
		if a.OrgID == orgID && a.ID == id {
			a.Status = status
			a.ResultSummary = resultSummary
		}
	}
	return nil
}

func (m *mockRepository) GetActionByID(ctx context.Context, orgID int64, id int64) (*CopilotActionHistory, error) {
	for _, a := range m.actions {
		if a.OrgID == orgID && a.ID == id {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockRepository) ListActions(ctx context.Context, orgID int64, sessionID string, limit int) ([]CopilotActionHistory, error) {
	res := make([]CopilotActionHistory, 0)
	for _, a := range m.actions {
		if a.OrgID == orgID {
			if sessionID == "" || a.SessionID == sessionID {
				res = append(res, *a)
			}
		}
	}
	return res, nil
}

func (m *mockRepository) RetrieveModuleContext(ctx context.Context, orgID int64, module string, recordID *string) ([]map[string]interface{}, map[string]interface{}, error) {
	records := []map[string]interface{}{
		{
			"id":             101,
			"org_id":         orgID,
			"booking_number": "BK-2026-001",
			"carrier_scac":   "MSCU",
			"status":         "IN_TRANSIT",
		},
	}
	metrics := map[string]interface{}{
		"total_shipments": 1,
		"active_rfqs":     0,
	}
	return records, metrics, nil
}

func (m *mockRepository) CreateApprovalRequest(ctx context.Context, orgID, userID int64, userName, title, category, actionName string, payload map[string]interface{}, corrID string) (int64, error) {
	newID := int64(1001 + len(m.approvalReqIDs))
	m.approvalReqIDs = append(m.approvalReqIDs, newID)
	return newID, nil
}

func (m *mockRepository) CreateRecommendation(ctx context.Context, orgID int64, title, category, description, sourceModule string, payload map[string]interface{}, corrID string) (int64, error) {
	newID := int64(501 + len(m.recIDs))
	m.recIDs = append(m.recIDs, newID)
	return newID, nil
}

// TestChatWorkflowWithSidecar verifies end-to-end Chat integration with mock sidecar
func TestChatWorkflowWithSidecar(t *testing.T) {
	// Setup mock HTTP sidecar server
	mockSidecar := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/copilot/chat" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("X-LogisticsHQ-Service-Key") != DefaultServiceKey {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req SidecarChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Check tenant isolation: req.Context.OrgID must equal 1
		if req.Context.OrgID != 1 {
			http.Error(w, "forbidden org", http.StatusForbidden)
			return
		}

		resp := SidecarChatResponse{
			Answer:             "Found shipment BK-2026-001 in transit from MSCU.",
			ConfirmedFacts:     []string{"Shipment BK-2026-001 is IN_TRANSIT with MSCU."},
			SourceReferences:   []SidecarSourceRef{{RecordType: "SHIPMENT", RecordID: "101", Title: "Shipment BK-2026-001"}},
			Signals:            []string{"Carrier on schedule"},
			AIInterpretation:   "Operations are progressing normally.",
			Recommendations:    []string{"Monitor ETA for milestone updates."},
			SuggestedFollowups: []string{"Show container tracking"},
			Confidence:         0.96,
			RequiresApproval:   false,
			CorrelationID:      req.CorrelationID,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockSidecar.Close()

	repo := newMockRepository()
	svc := NewService(repo, WithSidecar(mockSidecar.URL, DefaultServiceKey))

	ctx := context.Background()
	chatResp, err := svc.Chat(ctx, 1, 42, "tester", "ADMIN", ChatInput{
		CurrentModule: "SHIPMENTS",
		CurrentRoute:  "/shipments/101",
		Query:         "What is the status of shipment BK-2026-001?",
	})

	if err != nil {
		t.Fatalf("Chat returned unexpected error: %v", err)
	}

	if chatResp == nil {
		t.Fatal("Expected chat response, got nil")
	}

	if len(chatResp.ConfirmedFacts) != 1 {
		t.Errorf("Expected 1 confirmed fact, got %d", len(chatResp.ConfirmedFacts))
	}

	if chatResp.Confidence < 0.9 {
		t.Errorf("Expected confidence >= 0.9, got %f", chatResp.Confidence)
	}

	// Verify messages saved to repo
	if len(repo.messages) != 2 {
		t.Fatalf("Expected 2 messages (user + assistant) in repo, got %d", len(repo.messages))
	}
	if repo.messages[0].Role != "user" || repo.messages[1].Role != "assistant" {
		t.Errorf("Unexpected message roles: %s, %s", repo.messages[0].Role, repo.messages[1].Role)
	}
}

// TestConsequentialActionRequiresApproval tests that consequential action proposals are gated by approval_requests
func TestConsequentialActionRequiresApproval(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	ctx := context.Background()

	// 1. Consequential escalation action -> MUST result in PENDING_APPROVAL
	actionReq := ExecuteActionInput{
		ActionType:    "REQUEST_ESCALATION",
		ActionTitle:   "Escalate container delay to Operations Director",
		ActionPayload: map[string]interface{}{"shipment_id": "101", "level": 2},
		SessionID:     "ses-test-123",
		Reason:        "Delayed 48 hours",
	}

	res, err := svc.ExecuteAction(ctx, 1, 42, "tester", actionReq)
	if err != nil {
		t.Fatalf("ExecuteAction failed: %v", err)
	}

	if res.Status != "PENDING_APPROVAL" {
		t.Errorf("Expected status PENDING_APPROVAL for consequential escalation, got %s", res.Status)
	}

	if res.ApprovalID == nil || *res.ApprovalID <= 0 {
		t.Errorf("Expected non-nil ApprovalID for approval gate, got %v", res.ApprovalID)
	}

	// 2. Safe recommendation creation -> EXECUTED into Recommendation Center
	recReq := ExecuteActionInput{
		ActionType:    "CREATE_RECOMMENDATION",
		ActionTitle:   "Request updated certificate of origin",
		ActionPayload: map[string]interface{}{"category": "operations", "description": "Cert of origin expires next week"},
		SessionID:     "ses-test-123",
	}

	recRes, err := svc.ExecuteAction(ctx, 1, 42, "tester", recReq)
	if err != nil {
		t.Fatalf("ExecuteAction failed for recommendation: %v", err)
	}

	if recRes.Status != "EXECUTED" {
		t.Errorf("Expected status EXECUTED for recommendation, got %s", recRes.Status)
	}

	if recRes.RecommendationID == nil || *recRes.RecommendationID <= 0 {
		t.Errorf("Expected non-nil RecommendationID, got %v", recRes.RecommendationID)
	}
}

// TestTenantIsolationEnforced verifies that cross-tenant access returns no records or errors
func TestTenantIsolationEnforced(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	ctx := context.Background()

	// Create session for Org 1
	sessionID := "ses-org-1"
	repo.sessions[sessionID] = &CopilotSession{
		ID:        1,
		OrgID:     1,
		UserID:    10,
		SessionID: sessionID,
		Title:     "Org 1 Session",
	}

	// Org 2 user attempts to fetch Org 1 session -> Must return nil
	s, err := svc.GetSession(ctx, 2, sessionID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if s != nil {
		t.Errorf("Security violation: Org 2 was able to retrieve session belonging to Org 1: %+v", s)
	}
}
