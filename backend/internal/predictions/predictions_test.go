package predictions

import (
	"context"
	"strings"
	"testing"
	"time"
)

type mockRepo struct {
	predictions map[string]*Prediction
	audits      []*PredictionAuditHistory
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		predictions: make(map[string]*Prediction),
		audits:      make([]*PredictionAuditHistory, 0),
	}
}

func (m *mockRepo) Create(ctx context.Context, p *Prediction) error {
	m.predictions[p.PredictionID] = p
	return nil
}

func (m *mockRepo) GetByID(ctx context.Context, orgID int64, predictionID string) (*Prediction, error) {
	p, ok := m.predictions[predictionID]
	if !ok || p.OrgID != orgID {
		return nil, nil
	}
	return p, nil
}

func (m *mockRepo) GetByIdempotencyKey(ctx context.Context, orgID int64, key string) (*Prediction, error) {
	for _, p := range m.predictions {
		if p.OrgID == orgID && p.IdempotencyKey == key {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) List(ctx context.Context, orgID int64, params FilterParams) ([]*Prediction, int, error) {
	var res []*Prediction
	for _, p := range m.predictions {
		if p.OrgID == orgID {
			res = append(res, p)
		}
	}
	return res, len(res), nil
}

func (m *mockRepo) UpdateStatus(ctx context.Context, orgID int64, predictionID string, newStatus PredictionStatus, reviewStatus string, userID *int64, notes *string) error {
	p, ok := m.predictions[predictionID]
	if !ok || p.OrgID != orgID {
		return ErrPredictionNotFound
	}
	p.Status = newStatus
	p.ReviewStatus = reviewStatus
	p.ReviewedBy = userID
	now := time.Now()
	p.ReviewedAt = &now
	p.ReviewNotes = notes
	return nil
}

func (m *mockRepo) RecordAudit(ctx context.Context, a *PredictionAuditHistory) error {
	m.audits = append(m.audits, a)
	return nil
}

func (m *mockRepo) GetAuditHistory(ctx context.Context, orgID int64, predictionID string) ([]*PredictionAuditHistory, error) {
	var res []*PredictionAuditHistory
	for _, a := range m.audits {
		if a.OrgID == orgID && a.PredictionID == predictionID {
			res = append(res, a)
		}
	}
	return res, nil
}

func (m *mockRepo) RecordOutcome(ctx context.Context, orgID int64, predictionID string, outcomeStatus OutcomeStatus, outcomeValue *string, feedbackNotes *string) error {
	p, ok := m.predictions[predictionID]
	if !ok || p.OrgID != orgID {
		return ErrPredictionNotFound
	}
	p.ActualOutcomeStatus = outcomeStatus
	p.ActualOutcomeValue = outcomeValue
	p.FeedbackNotes = feedbackNotes
	return nil
}

func (m *mockRepo) GetActivePrediction(ctx context.Context, orgID int64, module string, relatedRecordType string, relatedRecordID string, predictionType string) (*Prediction, error) {
	for _, p := range m.predictions {
		if p.OrgID == orgID && p.Module == module && p.RelatedRecordType == relatedRecordType && p.RelatedRecordID == relatedRecordID && p.PredictionType == predictionType {
			if p.Status == StatusPublished || p.Status == StatusAcknowledged || p.Status == StatusInReview {
				return p, nil
			}
		}
	}
	return nil, nil
}

func (m *mockRepo) SupersedeExisting(ctx context.Context, orgID int64, module string, relatedRecordType string, relatedRecordID string, predictionType string, exceptPredictionID string) error {
	for _, p := range m.predictions {
		if p.OrgID == orgID && p.Module == module && p.RelatedRecordType == relatedRecordType && p.RelatedRecordID == relatedRecordID && p.PredictionType == predictionType {
			if p.PredictionID != exceptPredictionID {
				p.Status = StatusSuperseded
			}
		}
	}
	return nil
}

type mockSidecar struct {
	response *SidecarPredictionResponse
	err      error
}

func (m *mockSidecar) GeneratePrediction(ctx context.Context, req *GenerateRequest) (*SidecarPredictionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	resp := *m.response
	resp.OrgID = req.OrgID
	return &resp, nil
}

func TestValidationRules(t *testing.T) {
	repo := newMockRepo()
	sidecar := &mockSidecar{}
	svc := NewService(repo, sidecar)
	ctx := context.Background()

	t.Run("Rejects invalid org", func(t *testing.T) {
		_, err := svc.GenerateAndPersist(ctx, 0, nil, &GenerateRequest{
			Module:            "shipments",
			PredictionType:    "SHIPMENT_ETA_DELAY",
			RelatedRecordType: "SHIPMENT",
			RelatedRecordID:   "101",
		})
		if err != ErrInvalidOrgID {
			t.Fatalf("expected ErrInvalidOrgID, got %v", err)
		}
	})

	t.Run("Rejects invalid module", func(t *testing.T) {
		_, err := svc.GenerateAndPersist(ctx, 2, nil, &GenerateRequest{
			Module:            "invalid_module",
			PredictionType:    "SHIPMENT_ETA_DELAY",
			RelatedRecordType: "SHIPMENT",
			RelatedRecordID:   "101",
		})
		if err == nil || !strings.Contains(err.Error(), "unsupported prediction module") {
			t.Fatalf("expected unsupported module error, got %v", err)
		}
	})

	t.Run("Rejects invalid prediction type", func(t *testing.T) {
		_, err := svc.GenerateAndPersist(ctx, 2, nil, &GenerateRequest{
			Module:            "shipments",
			PredictionType:    "INVALID_PREDICTION",
			RelatedRecordType: "SHIPMENT",
			RelatedRecordID:   "101",
		})
		if err == nil || !strings.Contains(err.Error(), "unsupported prediction type") {
			t.Fatalf("expected unsupported prediction type error, got %v", err)
		}
	})

	t.Run("Rejects missing related record ID", func(t *testing.T) {
		_, err := svc.GenerateAndPersist(ctx, 2, nil, &GenerateRequest{
			Module:            "shipments",
			PredictionType:    "SHIPMENT_ETA_DELAY",
			RelatedRecordType: "SHIPMENT",
			RelatedRecordID:   "",
		})
		if err != ErrMissingRecordID {
			t.Fatalf("expected ErrMissingRecordID, got %v", err)
		}
	})
}

func TestGroundedPredictionAndLifecycle(t *testing.T) {
	repo := newMockRepo()
	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-test-101",
			Module:              "shipments",
			PredictionType:      "SHIPMENT_ETA_DELAY",
			RelatedRecordType:   "SHIPMENT",
			RelatedRecordID:     "101",
			PredictionStatement: "Predicted 36h ETA delay due to port congestion",
			Severity:            "HIGH",
			ConfidenceScore:     0.88,
			ConfidenceBand:      "HIGH",
			Explanation:         "Congestion multiplier 1.3x observed at Rotterdam",
			SourceReferences: []SourceReference{
				{
					SourceModule:    "shipments",
					SourceRecordID:  "101",
					SourceField:     "milestones.transshipment",
					SourceTimestamp: time.Now().UTC().Format(time.RFC3339),
				},
			},
			SourceTimestamp:   time.Now().UTC().Format(time.RFC3339),
			RecommendedAction: ptr("Send customer ETA notification"),
			ActionType:        ptr("shipments.send_customer_update"),
			RequiresApproval:  true,
		},
	}
	svc := NewService(repo, sidecar)
	ctx := context.Background()

	// 1. Generate & Persist
	pred, err := svc.GenerateAndPersist(ctx, 2, ptr(int64(10)), &GenerateRequest{
		Module:            "shipments",
		PredictionType:    "SHIPMENT_ETA_DELAY",
		RelatedRecordType: "SHIPMENT",
		RelatedRecordID:   "101",
		RecordContext:     map[string]interface{}{"status": "IN_TRANSIT"},
	})
	if err != nil {
		t.Fatalf("failed to generate prediction: %v", err)
	}

	if pred.PredictionID != "pred-test-101" {
		t.Errorf("expected prediction ID pred-test-101, got %s", pred.PredictionID)
	}
	if pred.Status != StatusPublished {
		t.Errorf("expected StatusPublished, got %s", pred.Status)
	}

	// 2. Tenant Isolation check: Org 1 cannot view Org 2's prediction
	_, err = svc.GetPrediction(ctx, 1, pred.PredictionID)
	if err != ErrPredictionNotFound {
		t.Errorf("expected ErrPredictionNotFound across tenants, got %v", err)
	}

	// 3. Org 2 retrieves cleanly
	retrieved, err := svc.GetPrediction(ctx, 2, pred.PredictionID)
	if err != nil || retrieved == nil {
		t.Fatalf("failed to retrieve prediction for org 2: %v", err)
	}

	// 4. Acknowledge
	if err := svc.AcknowledgePrediction(ctx, 2, pred.PredictionID, ptr(int64(10))); err != nil {
		t.Fatalf("failed to acknowledge: %v", err)
	}
	retrieved, _ = svc.GetPrediction(ctx, 2, pred.PredictionID)
	if retrieved.Status != StatusAcknowledged {
		t.Errorf("expected StatusAcknowledged, got %s", retrieved.Status)
	}

	// 5. Request Action
	if err := svc.RequestAction(ctx, 2, pred.PredictionID, ptr(int64(10)), "Proceed with customer alert"); err != nil {
		t.Fatalf("failed to request action: %v", err)
	}
	retrieved, _ = svc.GetPrediction(ctx, 2, pred.PredictionID)
	if retrieved.Status != StatusAwaitingApproval {
		t.Errorf("expected StatusAwaitingApproval, got %s", retrieved.Status)
	}

	// 6. Record Outcome
	outcomeVal := "Observed 34h delay"
	notes := "Prediction was accurate within 2 hours"
	if err := svc.RecordOutcome(ctx, 2, pred.PredictionID, OutcomeCorrect, &outcomeVal, &notes); err != nil {
		t.Fatalf("failed to record outcome: %v", err)
	}
	retrieved, _ = svc.GetPrediction(ctx, 2, pred.PredictionID)
	if retrieved.ActualOutcomeStatus != OutcomeCorrect {
		t.Errorf("expected OutcomeCorrect, got %s", retrieved.ActualOutcomeStatus)
	}

	// 7. Audit History
	history, err := svc.GetAuditHistory(ctx, 2, pred.PredictionID)
	if err != nil || len(history) < 3 {
		t.Fatalf("expected at least 3 audit entries, got %d (err: %v)", len(history), err)
	}
}

func TestPredictionKillSwitchGovernance(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepo()
	sidecar := &mockSidecar{
		response: &SidecarPredictionResponse{
			PredictionID:        "pred-test-ks-101",
			Module:              "shipments",
			PredictionType:      "SHIPMENT_ETA_DELAY",
			RelatedRecordType:   "SHIPMENT",
			RelatedRecordID:     "101",
			PredictionStatement: "Predicted delay",
			Severity:            "LOW",
			ConfidenceScore:     0.90,
			ConfidenceBand:      "HIGH",
			Explanation:         "Nominal schedule telemetry",
			SourceReferences: []SourceReference{
				{
					SourceModule:    "shipments",
					SourceRecordID:  "101",
					SourceField:     "eta",
					SourceTimestamp: time.Now().UTC().Format(time.RFC3339),
				},
			},
			SourceTimestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}

	svc := NewService(repo, sidecar)
	// Verify normal execution works without kill switch
	req := &GenerateRequest{
		Module:            "shipments",
		PredictionType:    "SHIPMENT_ETA_DELAY",
		RelatedRecordType: "SHIPMENT",
		RelatedRecordID:   "101",
	}
	pred, err := svc.GenerateAndPersist(ctx, 1, ptr(int64(1)), req)
	if err != nil || pred == nil {
		t.Fatalf("expected successful prediction generation, got error: %v", err)
	}

	// Verify kill switch helper logic
	impl, ok := svc.(*service)
	if !ok {
		t.Fatalf("expected *service implementation")
	}

	// When db is nil, kill switch returns false, ""
	killed, reason := impl.isKillSwitchActive(ctx, 1, "shipments", "SHIPMENT_ETA_DELAY")
	if killed {
		t.Errorf("expected killed=false with nil db, got true (reason: %s)", reason)
	}
}

func ptr[T any](v T) *T {
	return &v
}
