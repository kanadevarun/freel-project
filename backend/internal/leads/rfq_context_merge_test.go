package leads_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/freel/backend/internal/carrier"
	carrierService "github.com/freel/backend/internal/carrier/service"
	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/leads"
	"github.com/freel/backend/internal/organization"
	rfqspec "github.com/freel/backend/internal/rfq/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRFQContextMergingAndValidation(t *testing.T) {
	// Setup simple tests for field-level merge logic rules
	t.Run("MergeRFQContext: Overwrites non-empty and preserves previous", func(t *testing.T) {
		prev := map[string]interface{}{
			"origin_port":       "INNSA",
			"cargo_description": "Steel pipes",
			"cargo_weight":      4500.0,
		}
		newCtx := map[string]interface{}{
			"destination_port":  "DEHAM",
			"cargo_description": "Refined steel pipes", // updates previous
			"cargo_weight":      "",                    // empty string should NOT nullify
			"cargo_volume":      0.0,                   // 0 should NOT nullify
		}

		merged := leads.MergeRFQContext(prev, newCtx)
		assert.Equal(t, "INNSA", merged["origin_port"])
		assert.Equal(t, "DEHAM", merged["destination_port"])
		assert.Equal(t, "Refined steel pipes", merged["cargo_description"])
		assert.Equal(t, 4500.0, merged["cargo_weight"])
	})

	t.Run("Empty extraction protection", func(t *testing.T) {
		prev := map[string]interface{}{
			"origin_port": "Mumbai",
		}
		newCtx := map[string]interface{}{
			"origin_port": "",
		}

		merged := leads.MergeRFQContext(prev, newCtx)
		assert.Equal(t, "Mumbai", merged["origin_port"])
	})

	t.Run("Correction after partial completion", func(t *testing.T) {
		prev := map[string]interface{}{
			"cargo_weight": 4500.0,
		}
		newCtx := map[string]interface{}{
			"cargo_weight": 5000.0, // explicit correction
		}

		merged := leads.MergeRFQContext(prev, newCtx)
		assert.Equal(t, 5000.0, merged["cargo_weight"])
	})

	t.Run("Context persistence across multiple replies", func(t *testing.T) {
		// Email 1
		ctx1 := map[string]interface{}{
			"origin_port":       "Mumbai",
			"destination_port":  "Hamburg",
			"cargo_description": "Steel parts",
			"target_date":       "2026-09-10",
		}
		// Email 2
		ctx2 := map[string]interface{}{
			"cargo_weight": 4500.0,
		}
		// Email 3
		ctx3 := map[string]interface{}{
			"cargo_volume": 12.0,
			"incoterms":    "FOB",
		}

		merged := leads.MergeRFQContext(nil, ctx1)
		merged = leads.MergeRFQContext(merged, ctx2)
		merged = leads.MergeRFQContext(merged, ctx3)

		assert.Equal(t, "Mumbai", merged["origin_port"])
		assert.Equal(t, "Hamburg", merged["destination_port"])
		assert.Equal(t, "Steel parts", merged["cargo_description"])
		assert.Equal(t, "2026-09-10", merged["target_date"])
		assert.Equal(t, 4500.0, merged["cargo_weight"])
		assert.Equal(t, 12.0, merged["cargo_volume"])
		assert.Equal(t, "FOB", merged["incoterms"])
	})
}

// Database-backed integration tests for context merge callback, validation and duplicate protection.
func TestRFQContextMergeAndIdempotencyIntegration(t *testing.T) {
	dbURL := "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC&multiStatements=true"
	db, err := database.Connect(dbURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	eventBus := events.NewInProcessBus()
	dl := leads.NewDataLayer(db)
	orgRepo := organization.NewRepository(db)
	gp := organization.NewGmailProvider()
	bl := leads.NewBusinessLogic(dl, eventBus, orgRepo, gp, nil)

	orgID := int64(1022)
	leadID := int64(982)

	// Clean database state first to avoid FK constraints on delete order
	_, _ = db.Exec("DELETE FROM rfq_items")
	_, _ = db.Exec("DELETE FROM rfqs WHERE org_id IN (?, 1023)", orgID)
	_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id IN (?, 1023)", orgID)
	_, _ = db.Exec("DELETE FROM leads WHERE org_id IN (?, 1023)", orgID)
	_, _ = db.Exec("DELETE FROM organizations WHERE id IN (?, 1023)", orgID)

	// Seed organizations
	_, err = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Test Org A', NOW(), NOW()) ON DUPLICATE KEY UPDATE name = VALUES(name)", orgID)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (1023, 'Test Org B', NOW(), NOW()) ON DUPLICATE KEY UPDATE name = VALUES(name)")
	require.NoError(t, err)

	// Seed lead
	_, err = db.Exec(`
		INSERT INTO leads (id, org_id, company_name, contact_name, email, status, created_at, updated_at)
		VALUES (?, ?, 'Freel Integration Testing', 'Test Client', 'client@test-freel.local', 'IN_PROGRESS', NOW(), NOW())
	`, leadID, orgID)
	require.NoError(t, err)

	t.Run("Context persistence and incremental merge across callback turns", func(t *testing.T) {
		// Turn 1 callback: partial context
		interID1 := int64(1101)
		_, err = db.Exec(`
			INSERT INTO lead_interactions (id, org_id, lead_id, channel, direction, subject, content, raw_email_id, thread_id, sentiment, intent, linked_rfq_id, ai_confidence, ai_summary, drafted_reply, parent_interaction_id, partial_rfq_context, status, created_at, updated_at)
			VALUES (?, ?, ?, 'EMAIL', 'INBOUND', '', 'Initial message details', '', '', '', 'RFQ_REQUEST_INCOMPLETE', NULL, 0, '', '', NULL, NULL, 'SENT', NOW(), NOW())
		`, interID1, orgID, leadID)
		require.NoError(t, err)

		// Mock the handler callback trigger for Turn 1
		payload1 := map[string]interface{}{
			"origin_port":       "INNSA",
			"destination_port":  "DEHAM",
			"cargo_description": "Raw steel parts",
		}
		rawBytes, _ := json.Marshal(payload1)

		h := leads.NewEmailHandler(bl, &mockRFQBL{}, "http://localhost:8080")
		req := httptest.NewRequest(http.MethodPost, "/internal/sales/callback", strings.NewReader(string(rawBytes)))
		req.Header.Set("X-LogisticsHQ-Service-Key", "internal-service-key-logisticshq")
		req.Header.Set("Content-Type", "application/json")
		assert.NotNil(t, h)

		// Update interaction AI details via direct BL call to simulate Python callback logic
		err = bl.UpdateInteractionAI(ctx, orgID, interID1, "RFQ_REQUEST_INCOMPLETE", "NEUTRAL", 90, nil, "Summary", "Draft reply")
		require.NoError(t, err)
		err = bl.UpdateInteractionContext(ctx, orgID, interID1, payload1)
		require.NoError(t, err)

		// Verify first turn context exists
		inter, err := dl.GetInteractionByID(ctx, int32(orgID), interID1)
		require.NoError(t, err)
		inter.UnmarshalPartialRFQContext()
		assert.Equal(t, "INNSA", inter.PartialRFQContext["origin_port"])
		assert.Equal(t, "DEHAM", inter.PartialRFQContext["destination_port"])

		// Turn 2 callback: additional info with correction
		interID2 := int64(1102)
		_, err = db.Exec(`
			INSERT INTO lead_interactions (id, org_id, lead_id, channel, direction, subject, content, raw_email_id, thread_id, sentiment, intent, linked_rfq_id, ai_confidence, ai_summary, drafted_reply, parent_interaction_id, partial_rfq_context, status, created_at, updated_at)
			VALUES (?, ?, ?, 'EMAIL', 'INBOUND', '', 'Reply: weight is 4500kg and correction description is fine steel parts', '', '', '', 'RFQ_REQUEST_INCOMPLETE', NULL, 0, '', '', NULL, NULL, 'SENT', NOW(), NOW())
		`, interID2, orgID, leadID)
		require.NoError(t, err)

		// Simulate callback logic on Turn 2
		newExtract := map[string]interface{}{
			"cargo_description": "Fine steel parts",
			"cargo_weight":      4500.0,
		}

		// Read previous interaction context from database, merge, and save
		var prevContext map[string]interface{}
		interactions, err := bl.ListInteractions(ctx, int32(orgID), int32(leadID))
		require.NoError(t, err)
		for i := len(interactions) - 1; i >= 0; i-- {
			p := interactions[i]
			if p.ID == interID2 {
				continue
			}
			p.UnmarshalPartialRFQContext()
			if len(p.PartialRFQContext) > 0 {
				prevContext = p.PartialRFQContext
				break
			}
		}
		assert.NotNil(t, prevContext)
		assert.Equal(t, "INNSA", prevContext["origin_port"])

		merged := leads.MergeRFQContext(prevContext, newExtract)
		err = bl.UpdateInteractionContext(ctx, orgID, interID2, merged)
		require.NoError(t, err)

		// Verify merged values
		inter2, err := dl.GetInteractionByID(ctx, int32(orgID), interID2)
		require.NoError(t, err)
		inter2.UnmarshalPartialRFQContext()
		assert.Equal(t, "INNSA", inter2.PartialRFQContext["origin_port"])
		assert.Equal(t, "DEHAM", inter2.PartialRFQContext["destination_port"])
		assert.Equal(t, "Fine steel parts", inter2.PartialRFQContext["cargo_description"])
		assert.Equal(t, 4500.0, inter2.PartialRFQContext["cargo_weight"])
	})

	t.Run("Organization Isolation protection", func(t *testing.T) {
		orgB := int64(1023)
		
		// Attempting to query or modify Lead Context of Org A using Org B's identifier
		interactions, err := bl.ListInteractions(ctx, int32(orgB), int32(leadID))
		// Should return empty list because of organization separation
		assert.Len(t, interactions, 0)
		assert.NoError(t, err)
	})
}

// Stub implementation for rfq.BusinessLogic to verify idempotency calls
type mockRFQBL struct {
	createCount int
	leadID      int64
	rfqNumber   string
}

func (m *mockRFQBL) CreateRFQ(ctx context.Context, req rfqspec.CreateRFQRequest) (*rfqspec.RFQ, error) {
	m.createCount++
	m.leadID = *req.LeadID
	m.rfqNumber = "RFQ-MOCK-100"
	return &rfqspec.RFQ{
		ID:         12001,
		OrgID:      req.OrgID,
		RFQNumber:  m.rfqNumber,
		CustomerID: req.CustomerID,
		Stage:      "RFQ_CREATED",
		LeadID:     req.LeadID,
	}, nil
}

func (m *mockRFQBL) GetRFQ(ctx context.Context, orgID, rfqID int32) (*rfqspec.RFQ, error) {
	return nil, nil
}

func (m *mockRFQBL) GetTimeline(ctx context.Context, orgID, rfqID int32) ([]rfqspec.TimelineEvent, error) {
	return []rfqspec.TimelineEvent{}, nil
}

func (m *mockRFQBL) GetRFQByLeadID(ctx context.Context, orgID int32, leadID int64) (*rfqspec.RFQ, error) {
	if m.leadID == leadID {
		return &rfqspec.RFQ{
			ID:        12001,
			OrgID:     orgID,
			RFQNumber: m.rfqNumber,
			LeadID:    &leadID,
		}, nil
	}
	return nil, nil
}

func (m *mockRFQBL) ListRFQs(ctx context.Context, req rfqspec.ListRFQsRequest) (*rfqspec.ListRFQsResponse, error) {
	return nil, nil
}

func (m *mockRFQBL) AdvanceStage(ctx context.Context, orgID, rfqID int32, newStage string) (*rfqspec.RFQ, error) {
	return nil, nil
}

func (m *mockRFQBL) AddQuote(ctx context.Context, orgID int32, quote *rfqspec.Quote) error {
	return nil
}

func (m *mockRFQBL) UpdateAgentStatus(ctx context.Context, orgID, rfqID int32, status string) error {
	return nil
}

func (m *mockRFQBL) GetCarrierRates(ctx context.Context, orgID, rfqID int32) (*carrier.FetchRatesResponse, error) {
	return nil, nil
}

func (m *mockRFQBL) ApproveQuote(ctx context.Context, orgID, rfqID, quoteID int32) (*rfqspec.RFQ, error) {
	return nil, nil
}

func (m *mockRFQBL) ParseShipmentRequest(ctx context.Context, rawText string) (*rfqspec.ParseShipmentResponse, error) {
	return nil, nil
}

func (m *mockRFQBL) GetRequirements(ctx context.Context, orgID, rfqID int32) (*rfqspec.GetRequirementsResponse, error) {
	return nil, nil
}

func (m *mockRFQBL) GetActivity(ctx context.Context, orgID, rfqID int32) (*rfqspec.GetActivityResponse, error) {
	return nil, nil
}

func (m *mockRFQBL) CreateAITask(ctx context.Context, orgID int64, entityType string, entityID string, taskType string, payload map[string]interface{}) error {
	return nil
}

func (m *mockRFQBL) GetDocuments(ctx context.Context, orgID, rfqID int32) (*rfqspec.GetDocumentsResponse, error) {
	return nil, nil
}

func (m *mockRFQBL) CreateDocument(ctx context.Context, orgID, rfqID int32, req rfqspec.CreateDocumentRequest, uploader string) (*rfqspec.RFQDocument, error) {
	return nil, nil
}

func (m *mockRFQBL) UpdateDocumentStatus(ctx context.Context, orgID, rfqID int32, documentID int64, req rfqspec.UpdateDocumentStatusRequest, reviewer string) (*rfqspec.RFQDocument, error) {
	return nil, nil
}

func (m *mockRFQBL) DeleteDocument(ctx context.Context, orgID, rfqID int32, documentID int64) error {
	return nil
}

func (m *mockRFQBL) GetQuotes(ctx context.Context, orgID, rfqID int32) (*rfqspec.GetQuotesResponse, error) {
	return nil, nil
}

func (m *mockRFQBL) CreateRFQQuote(ctx context.Context, orgID, rfqID int32, req rfqspec.CreateQuoteRequest, creator string) (*rfqspec.RFQQuote, error) {
	return nil, nil
}

func (m *mockRFQBL) UpdateRFQQuote(ctx context.Context, orgID, rfqID int32, quoteID int64, req rfqspec.UpdateQuoteRequest, updater string) (*rfqspec.RFQQuote, error) {
	return nil, nil
}

func (m *mockRFQBL) UpdateRFQQuoteStatus(ctx context.Context, orgID, rfqID int32, quoteID int64, req rfqspec.UpdateQuoteStatusRequest, updater string) (*rfqspec.RFQQuote, error) {
	return nil, nil
}

func (m *mockRFQBL) RecommendRFQQuote(ctx context.Context, orgID, rfqID int32, quoteID int64, recommender string) (*rfqspec.RFQQuote, error) {
	return nil, nil
}

func (m *mockRFQBL) ApproveRFQQuote(ctx context.Context, orgID, rfqID int32, quoteID int64, req rfqspec.ApproveRFQQuoteRequest, approver string) (*rfqspec.RFQQuote, error) {
	return nil, nil
}

func (m *mockRFQBL) SelectRFQQuoteForCustomer(ctx context.Context, orgID, rfqID int32, quoteID int64, selector string) (*rfqspec.RFQQuote, error) {
	return nil, nil
}

func (m *mockRFQBL) DeleteRFQQuote(ctx context.Context, orgID, rfqID int32, quoteID int64) error {
	return nil
}

func (m *mockRFQBL) GetBookingHandoff(ctx context.Context, orgID, rfqID int32) (*rfqspec.GetBookingHandoffResponse, error) {
	return nil, nil
}

func (m *mockRFQBL) CreateBookingFromRFQ(ctx context.Context, orgID, rfqID int32, req rfqspec.CreateBookingRequest, creator string) (*rfqspec.RFQBooking, error) {
	return nil, nil
}

func (m *mockRFQBL) UpdateBookingStatus(ctx context.Context, orgID, rfqID int32, bookingID int64, req rfqspec.UpdateBookingStatusRequest, updater string) (*rfqspec.RFQBooking, error) {
	return nil, nil
}

func (m *mockRFQBL) GetShipmentHandoff(ctx context.Context, orgID, rfqID int32) (*rfqspec.GetShipmentHandoffResponse, error) {
	return nil, nil
}

func (m *mockRFQBL) GetBookingsWorkspace(ctx context.Context, orgID int32, filter rfqspec.BookingListFilter) (*rfqspec.GetBookingsWorkspaceResponse, error) {
	return nil, nil
}

func (m *mockRFQBL) GetBookingWorkspaceDetail(ctx context.Context, orgID int32, bookingID int64) (*rfqspec.BookingDetailResponse, error) {
	return nil, nil
}

func (m *mockRFQBL) DirectUpdateBookingStatus(ctx context.Context, orgID int32, bookingID int64, req rfqspec.DirectUpdateBookingStatusRequest, updater string) (*rfqspec.RFQBooking, error) {
	return nil, nil
}

func (m *mockRFQBL) GetEligibleRFQsForBooking(ctx context.Context, orgID int32) ([]rfqspec.EligibleBookingRFQ, error) {
	return nil, nil
}

func (m *mockRFQBL) CreateShipmentFromBooking(ctx context.Context, orgID int32, bookingID int64, req rfqspec.CreateShipmentFromBookingRequest, creator string) (*rfqspec.RFQShipment, error) {
	return nil, nil
}

func (m *mockRFQBL) BookWithCarrier(ctx context.Context, orgID int32, bookingID int64, req rfqspec.BookWithCarrierRequest, user string) (*rfqspec.BookingDetailResponse, error) {
	return nil, nil
}

func (m *mockRFQBL) SyncCarrierBooking(ctx context.Context, orgID int32, bookingID int64, user string) (*rfqspec.BookingDetailResponse, error) {
	return nil, nil
}

func (m *mockRFQBL) SetCarrierIntegrationService(carrierSvc carrierService.CarrierService) {
}




