package rfq_test

import (
	"context"
	"testing"
	"time"

	"github.com/freel/backend/internal/ai"
	"github.com/freel/backend/internal/carrier"
	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/rates"
	"github.com/freel/backend/internal/rfq"
	"github.com/freel/backend/internal/rfq/mocks"
	"github.com/freel/backend/internal/rfq/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ── Test Setup ───────────────────────────────────────────────────────────────

// newServiceWithMocks creates a new RFQ service with a mocked repository and event bus.
type mockEntitlementService struct{}
func (m *mockEntitlementService) CheckEntitlement(ctx context.Context, orgID int64, metricName string) error { return nil }
func (m *mockEntitlementService) IncrementUsage(ctx context.Context, orgID int64, metricName string, amount int) error { return nil }

func newServiceWithMocks(t *testing.T) (rfq.BusinessLogic, *mocks.MockDatalayer, events.Bus) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockDatalayer(ctrl)
	eventBus := events.NewInProcessBus() 
	
	// satisfy the carrier.Service dependency using mock provider for tests
	carrierProvider := carrier.NewMockProvider()
	carrierSvc := carrier.NewService(carrierProvider)

	// satisfy the rates.Service dependency
	ratesRepo := rates.NewRepository(nil)
	spotNormalizer := rates.NewSpotNormalizer()
	rateSvc := rates.NewService(ratesRepo, spotNormalizer, carrierSvc)

	// satisfy the ai.Gateway and ai.PromptManager dependencies
	mockProviders := map[string]ai.Provider{
		"mock": ai.NewMockProvider(),
	}
	aiGateway := ai.NewGateway(mockProviders)
	promptManager := ai.NewPromptManager()

	svc := rfq.NewBusinessLogic(mockRepo, eventBus, rateSvc, aiGateway, promptManager, &mockEntitlementService{})
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

func TestService_ParseShipmentRequest(t *testing.T) {
	t.Run("Successfully parses raw shipment request email via AI", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockDatalayer(ctrl)
		eventBus := events.NewInProcessBus()

		// satisfy dependencies
		carrierProvider := carrier.NewMockProvider()
		carrierSvc := carrier.NewService(carrierProvider)
		ratesRepo := rates.NewRepository(nil)
		spotNormalizer := rates.NewSpotNormalizer()
		rateSvc := rates.NewService(ratesRepo, spotNormalizer, carrierSvc)

		// Create a custom provider returning the exact expected JSON format
		customResponse := `{
			"data": {
				"origin": "Nhava Sheva",
				"destination": "Hamburg",
				"incoterms": "FOB",
				"weight": "15 Tons",
				"volume": "30 CBM"
			},
			"confidence_score": 95,
			"missing_fields": ["Target Date"]
		}`
		mockProviders := map[string]ai.Provider{
			"mock": &customMockProvider{response: customResponse},
		}
		aiGateway := ai.NewGateway(mockProviders)
		promptManager := ai.NewPromptManager()

		svc := rfq.NewBusinessLogic(mockRepo, eventBus, rateSvc, aiGateway, promptManager, &mockEntitlementService{})

		resp, err := svc.ParseShipmentRequest(context.Background(), "Please quote from Nhava Sheva to Hamburg FOB, weight 15 Tons, 30 CBM")
		require.NoError(t, err)
		assert.NotNil(t, resp)

		dataMap := resp.Data.(map[string]interface{})
		innerData := dataMap["data"].(map[string]interface{})

		assert.Equal(t, "Nhava Sheva", *(innerData["origin"].(*string)))
		assert.Equal(t, "Hamburg", *(innerData["destination"].(*string)))
		assert.Equal(t, "FOB", *(innerData["incoterms"].(*string)))
		assert.Equal(t, "15 Tons", *(innerData["weight"].(*string)))
		assert.Equal(t, "30 CBM", *(innerData["volume"].(*string)))
		assert.Equal(t, 95, dataMap["confidence_score"].(int))
		assert.Equal(t, []string{"Target Date"}, dataMap["missing_fields"].([]string))
	})
}

type customMockProvider struct {
	response string
}

func (p *customMockProvider) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	return p.response, nil
}

func TestService_GetRFQ_And_Timeline(t *testing.T) {
	t.Run("GetRFQ retrieves RFQ with items, customer details, and lead linkage", func(t *testing.T) {
		svc, mockRepo, _ := newServiceWithMocks(t)
		leadID := int64(171)
		custEmail := "client@globalexports.com"
		weight := 12500.0
		vol := 28.0

		expectedRFQ := &spec.RFQ{
			ID:            101,
			OrgID:         1,
			RFQNumber:     "RFQ-20260827-001",
			CustomerID:    42,
			CustomerName:  "Global Exports Ltd",
			CustomerEmail: &custEmail,
			Stage:         spec.StageRFQCreated,
			LeadID:        &leadID,
			Items: []spec.RFQItem{
				{
					ID:          1,
					RFQID:       101,
					Description: "Steel Components",
					Quantity:    1,
					WeightKG:    &weight,
					VolumeCBM:   &vol,
				},
			},
		}

		mockRepo.EXPECT().GetRFQByID(gomock.Any(), int32(1), int32(101)).Return(expectedRFQ, nil).Times(1)

		res, err := svc.GetRFQ(context.Background(), 1, 101)
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, "RFQ-20260827-001", res.RFQNumber)
		assert.Equal(t, "Global Exports Ltd", res.CustomerName)
		assert.Equal(t, int64(171), *res.LeadID)
		assert.Len(t, res.Items, 1)
	})

	t.Run("GetTimeline returns chronological events across Lead and RFQ", func(t *testing.T) {
		svc, mockRepo, _ := newServiceWithMocks(t)
		leadID := int64(171)

		expectedRFQ := &spec.RFQ{
			ID:        101,
			OrgID:     1,
			RFQNumber: "RFQ-20260827-001",
			LeadID:    &leadID,
		}

		timelineEvents := []spec.TimelineEvent{
			{
				ID:          "act-1",
				EntityType:  "LEAD",
				EntityID:    171,
				Category:    "LEAD",
				Action:      "CREATED",
				Description: "Lead created from customer inquiry",
				Actor:       "System",
				Timestamp:   time.Now().Add(-2 * time.Hour),
			},
			{
				ID:          "inter-1",
				EntityType:  "LEAD",
				EntityID:    171,
				Category:    "EMAIL",
				Action:      "EMAIL_INBOUND",
				Description: "INBOUND: Quote for 40ft container",
				Actor:       "Customer",
				Timestamp:   time.Now().Add(-1 * time.Hour),
			},
			{
				ID:          "act-2",
				EntityType:  "LEAD",
				EntityID:    171,
				Category:    "AI",
				Action:      "AI_PARSED",
				Description: "AI extracted shipment details (Origin: Mumbai, Destination: Hamburg)",
				Actor:       "Email Parser AI",
				Timestamp:   time.Now().Add(-45 * time.Minute),
			},
			{
				ID:          "act-3",
				EntityType:  "RFQ",
				EntityID:    101,
				Category:    "RFQ",
				Action:      "RFQ_CREATED",
				Description: "RFQ RFQ-20260827-001 created from Lead #171",
				Actor:       "System",
				Timestamp:   time.Now().Add(-30 * time.Minute),
			},
		}

		mockRepo.EXPECT().GetRFQByID(gomock.Any(), int32(1), int32(101)).Return(expectedRFQ, nil).Times(1)
		mockRepo.EXPECT().GetRFQTimeline(gomock.Any(), int32(1), int32(101), &leadID).Return(timelineEvents, nil).Times(1)

		events, err := svc.GetTimeline(context.Background(), 1, 101)
		require.NoError(t, err)
		assert.Len(t, events, 4)
		assert.Equal(t, "CREATED", events[0].Action)
		assert.Equal(t, "RFQ_CREATED", events[3].Action)
	})
}

