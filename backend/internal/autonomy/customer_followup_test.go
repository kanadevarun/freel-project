package autonomy_test

import (
	"context"
	"testing"

	"github.com/freel/backend/internal/autonomy"
	"github.com/stretchr/testify/assert"
)

func TestCustomerFollowup_OptOutEnforcement(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{
		custEvalResp: &autonomy.SidecarCustomerFollowupEvalResponse{
			CustomerID:       101,
			EventType:        "SHIPMENT_DELAY",
			Decision:         "STOP",
			DecisionReason:   "Customer has explicitly opted out of communications.",
			Urgency:          "LOW",
			RequiresApproval: false,
			CorrelationID:    "corr-opt-out",
		},
	}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	// Save customer preferences as Opted Out
	_ = repo.SaveCustomerPreferences(context.Background(), &autonomy.CustomerCommunicationPreferences{
		OrgID:            2,
		CustomerID:       101,
		PreferredChannel: "EMAIL",
		OptOut:           true,
		OptOutReason:     stringPtr("Client opted out of automated alerts"),
	})

	req := autonomy.IngestCustomerFollowupEventRequest{
		CustomerID:       101,
		EventType:        "SHIPMENT_DELAY",
		Payload:          map[string]interface{}{"delay_hours": 36.0},
		DeduplicationKey: "dedup-opt-out-test-1",
	}

	res, err := svc.ProcessCustomerFollowupEvent(context.Background(), 2, nil, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "STOP", res.Decision)
	assert.Contains(t, res.DecisionReason, "opted out")
}

func TestCustomerFollowup_EventDeduplication(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	// Pre-seed an existing record with dedup key
	dedupKey := "fixed-dedup-key-cust-101"
	_, _ = repo.CreateFollowupRecord(context.Background(), &autonomy.CustomerFollowupRecord{
		OrgID:          2,
		CustomerID:     101,
		EventType:      "QUOTATION_EXPIRING",
		Channel:        "EMAIL",
		RecipientEmail: "v.malhotra@apexlogistics.com",
		RecipientName:  "Vikram Malhotra",
		Subject:        "Quotation Expiring",
		FullBody:       "Body text",
		Status:         "DRAFT",
		IdempotencyKey: dedupKey,
	})

	req := autonomy.IngestCustomerFollowupEventRequest{
		CustomerID:       101,
		EventType:        "QUOTATION_EXPIRING",
		Payload:          map[string]interface{}{"quotation_number": "QT-099"},
		DeduplicationKey: dedupKey,
	}

	res, err := svc.ProcessCustomerFollowupEvent(context.Background(), 2, nil, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "MONITOR", res.Decision)
	assert.Contains(t, res.DecisionReason, "Duplicate event detected")
}

func TestCustomerFollowup_ResponseClassification(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{
		custClassResp: &autonomy.SidecarClassifyCustomerResponseResponse{
			CustomerID:          101,
			Classification:      "CONFIRMATION_APPROVAL",
			Sentiment:           "POSITIVE",
			RecommendedNextStep: "RESOLVE_AND_STOP",
			Confidence:          0.97,
		},
	}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)

	rec, _ := repo.CreateFollowupRecord(context.Background(), &autonomy.CustomerFollowupRecord{
		OrgID:          2,
		CustomerID:     101,
		EventType:      "SHIPMENT_DELAY",
		Channel:        "EMAIL",
		RecipientEmail: "v.malhotra@apexlogistics.com",
		RecipientName:  "Vikram Malhotra",
		Subject:        "Shipment Advisory",
		FullBody:       "Body",
		Status:         "SENT",
		IdempotencyKey: "dedup-resp-101",
	})

	res, err := svc.IngestCustomerResponse(context.Background(), 2, nil, rec.ID, "Looks good, proceed.")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "CONFIRMATION_APPROVAL", res.Classification)
	assert.Equal(t, "POSITIVE", res.Sentiment)
	assert.Equal(t, "RESOLVE_AND_STOP", res.RecommendedNextStep)
}

func stringPtr(s string) *string {
	return &s
}

