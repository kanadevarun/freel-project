package rfq_test

import (
	"context"
	"testing"
	"time"

	"github.com/freel/backend/internal/carrier"
	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/rfq"
	"github.com/freel/backend/internal/rfq/mocks"
	"github.com/freel/backend/internal/rfq/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ── Test Setup ───────────────────────────────────────────────────────────────

// newServiceWithMocks creates a new RFQ service with a mocked repository and event bus.
func newServiceWithMocks(t *testing.T) (rfq.BusinessLogic, *mocks.MockDatalayer, events.Bus) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockDatalayer(ctrl)
	eventBus := events.NewInProcessBus() 
	
	// satisfy the carrier.Service dependency using mock provider for tests
	carrierProvider := carrier.NewMockProvider()
	carrierSvc := carrier.NewService(carrierProvider)

	svc := rfq.NewBusinessLogic(mockRepo, eventBus, carrierSvc)
	return svc, mockRepo, eventBus
}

// ── Tests ────────────────────────────────────────────────────────────────────

func TestService_CreateRFQ(t *testing.T) {
	t.Run("Successfully creates RFQ and publishes event", func(t *testing.T) {
		svc, mockRepo, eventBus := newServiceWithMocks(t)
		
		mockRepo.EXPECT().CreateRFQ(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, r *spec.RFQ) error {
			r.ID = 1
			return nil
		}).Times(1)

		mockRepo.EXPECT().CreateRFQItem(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, item *spec.RFQItem) error {
			item.ID = 1
			return nil
		}).Times(1)

		eventChan := make(chan events.Event, 1)
		eventBus.Subscribe(events.EventRFQCreated, func(e events.Event) {
			eventChan <- e
		})

		req := spec.CreateRFQRequest{
			OrgID:      5,
			CustomerID: 99,
			Items: []spec.RFQItem{
				{Description: "Electronics", Quantity: 100},
			},
		}
		
		createdRfq, err := svc.CreateRFQ(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, int32(1), createdRfq.ID)
		assert.Equal(t, int32(5), createdRfq.OrgID)
		assert.Equal(t, int32(99), createdRfq.CustomerID)
		assert.Equal(t, spec.StageRFQCreated, createdRfq.Stage)
		assert.Contains(t, createdRfq.RFQNumber, "RFQ-")

		select {
		case e := <-eventChan:
			assert.Equal(t, events.EventRFQCreated, e.Type)
			payload := e.Payload.(map[string]interface{})
			assert.Equal(t, int32(1), payload["rfq_id"])
		case <-time.After(1 * time.Second):
			t.Fatal("Expected RFQCreated event to be published, but it timed out")
		}
	})
}

func TestService_AdvanceStage(t *testing.T) {
	t.Run("Successfully advances stage and publishes event", func(t *testing.T) {
		svc, mockRepo, eventBus := newServiceWithMocks(t)
		
		existingRFQ := &spec.RFQ{
			ID:    1,
			OrgID: 5,
			Stage: spec.StageRFQCreated,
		}

		mockRepo.EXPECT().GetRFQByID(gomock.Any(), int32(5), int32(1)).Return(existingRFQ, nil).Times(1)
		mockRepo.EXPECT().UpdateStage(gomock.Any(), int32(5), int32(1), spec.StagePricingAssigned).Return(nil).Times(1)

		eventChan := make(chan events.Event, 1)
		eventBus.Subscribe(events.EventRFQAssigned, func(e events.Event) {
			eventChan <- e
		})

		updatedRfq, err := svc.AdvanceStage(context.Background(), 5, 1, spec.StagePricingAssigned)

		require.NoError(t, err)
		assert.Equal(t, spec.StagePricingAssigned, updatedRfq.Stage)

		select {
		case e := <-eventChan:
			assert.Equal(t, events.EventRFQAssigned, e.Type)
			payload := e.Payload.(map[string]interface{})
			assert.Equal(t, spec.StagePricingAssigned, payload["new_stage"])
		case <-time.After(1 * time.Second):
			t.Fatal("Expected rfq.assigned event to be published")
		}
	})

	t.Run("Fails if stage is invalid", func(t *testing.T) {
		svc, mockRepo, _ := newServiceWithMocks(t)
		
		existingRFQ := &spec.RFQ{ID: 1, OrgID: 5, Stage: spec.StageRFQCreated}

		mockRepo.EXPECT().GetRFQByID(gomock.Any(), int32(5), int32(1)).Return(existingRFQ, nil).Times(1)
		mockRepo.EXPECT().UpdateStage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		updatedRfq, err := svc.AdvanceStage(context.Background(), 5, 1, "INVALID_STAGE")

		require.Error(t, err)
		assert.Nil(t, updatedRfq)
	})
}
