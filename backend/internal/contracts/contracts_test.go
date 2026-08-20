package contracts_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/freel/backend/internal/contracts"
	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/files"
	"github.com/freel/backend/internal/rates"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContractLifecycle(t *testing.T) {
	// 1. Connect to local development database
	dbURL := "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC&multiStatements=true"
	db, err := database.Connect(dbURL)
	require.NoError(t, err)
	defer db.Close()

	// Clean up tables to prevent test pollution and paging issues
	_, _ = db.Exec("DELETE FROM rate_entries")
	_, _ = db.Exec("DELETE FROM contract_documents")
	_, _ = db.Exec("INSERT INTO carriers (scac, name) VALUES ('MAEU', 'Maersk Line') ON DUPLICATE KEY UPDATE name = VALUES(name)")
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (1, 'Test Freight Forwarder', NOW(), NOW()) ON DUPLICATE KEY UPDATE name = VALUES(name)")
	_, _ = db.Exec("INSERT INTO users (id, cognito_sub, email, created_at, updated_at) VALUES (1, 'test-sub-1', 'test@example.com', NOW(), NOW()) ON DUPLICATE KEY UPDATE email = VALUES(email)")

	ctx := context.Background()
	orgID := int64(1) // seeded organization
	userID := int64(1)

	// 2. Initialize dependencies
	localFilesSvc := files.NewLocalService("./uploads", "http://localhost:8080/uploads")
	contractsRepo := contracts.NewRepository(db)
	mockAIBridge := &mockAIBridge{}
	ratesRepo := rates.NewRepository(db)
	spotNormalizer := rates.NewSpotNormalizer()
	rateSvc := rates.NewService(ratesRepo, spotNormalizer, nil)
	contractsSvc := contracts.NewService(contractsRepo, localFilesSvc, mockAIBridge, rateSvc, "http://localhost:8080/internal/contracts/callback")

	// 3. Upload a dummy contract PDF
	dummyContent := bytes.NewReader([]byte("%PDF-1.4 mock content"))
	carrierSCAC := "MAEU"
	doc, err := contractsSvc.UploadContract(ctx, orgID, userID, "maersk_contract_2026.pdf", dummyContent, 20, &carrierSCAC)
	require.NoError(t, err)
	assert.NotEmpty(t, doc.ID)
	assert.Equal(t, contracts.StatusQueued, doc.Status)

	// Allow goroutine to execute trigger
	time.Sleep(100 * time.Millisecond)

	// Fetch document from DB to check status advanced to OCR_PROCESSING
	fetchedDoc, err := contractsSvc.GetDocument(ctx, orgID, doc.ID)
	require.NoError(t, err)
	assert.Equal(t, contracts.StatusOCRProcessing, fetchedDoc.Status)

	// 4. Simulate Python AI sidecar Callback payload
	confirmedRatesDraft := []rates.CanonicalRate{
		{
			ID:                 uuid.NewString(),
			OrgID:              orgID,
			OriginPort:         "INNSA",
			DestinationPort:    "DEHAM",
			ViaPort:            "SGSIN",
			ServiceCode:        "AE-1",
			CarrierSCAC:        "MAEU",
			CarrierName:        "Maersk",
			VesselName:         "MAERSK MC-KINNEY MOLLER",
			EquipmentType:      "40GP",
			OceanFreight:       2800.0,
			OriginCharges:      180.0,
			DestinationCharges: 220.0,
			Surcharges: []rates.Surcharge{
				{Code: "BAF", Description: "Bunker Adjustment Factor", Amount: 350.0, Unit: rates.SurchargeUnitPerTEU, Included: true},
			},
			TotalBuyPrice: 3200.0,
			ValidFrom:     time.Now().UTC(),
			ValidUntil:    time.Now().UTC().Add(48 * time.Hour),
		},
	}

	flaggedItemsDraft := []contracts.ReviewItemDraft{
		{
			ExtractedData: map[string]interface{}{
				"origin_port":      "INNSA",
				"destination_port": "AEJEA",
				"carrier_scac":     "MAEU",
				"carrier_name":     "Maersk",
				"equipment_type":   "40GP",
				"ocean_freight":    12400.0,
				"total_buy_price":  12730.0,
				"valid_from":       "2026-09-01T00:00:00Z",
				"valid_until":      "2026-12-31T23:59:59Z",
			},
			ConfidenceScore: 62,
			ReviewFlags:     []string{"PRICE_ANOMALY"},
			AIReasoning:     "Ocean Freight is extremely high compared to market averages.",
			SourcePage:      3,
			SourceText:      "Origin Nhava Sheva to Jebel Ali: USD 12,400 per 40GP.",
		},
	}

	callbackPayload := contracts.AIProcessingCallback{
		DocumentID:     doc.ID,
		OrgID:          orgID,
		Status:         "COMPLETED",
		ConfirmedRates: confirmedRatesDraft,
		FlaggedItems:   flaggedItemsDraft,
		ProcessingLog: []contracts.LogEntry{
			{Step: "AI_PROCESSING", Timestamp: time.Now().UTC(), Message: "Completed layout parsing and normalizations"},
		},
		AISummary: "Simulated Maersk ocean contract.",
	}

	// Invoke callback
	err = contractsSvc.HandleAICallback(ctx, callbackPayload)
	require.NoError(t, err)

	// 5. Assertions

	// a) Contract document should be PENDING_REVIEW (since we had 1 flagged item)
	finalDoc, err := contractsSvc.GetDocument(ctx, orgID, doc.ID)
	require.NoError(t, err)
	assert.Equal(t, contracts.StatusPendingReview, finalDoc.Status)
	assert.Equal(t, 2, finalDoc.ExtractedRateCount)
	assert.Equal(t, 1, finalDoc.ConfirmedRateCount)
	assert.Equal(t, 1, finalDoc.PendingReviewCount)

	// b) Confirmed rate should be searchable via rates.Service
	rateResult, err := rateSvc.SearchRates(ctx, rates.RateQuery{
		OrgID:           orgID,
		OriginPort:      "INNSA",
		DestinationPort: "DEHAM",
		EquipmentType:   "40GP",
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, rateResult.TotalCount, 1)

	foundConfirmed := false
	for _, r := range rateResult.Rates {
		t.Logf("Rate: ID=%s, Source=%s, ContractDocID=%v, TargetDocID=%s", r.ID, r.Source, r.ContractDocID, doc.ID)
		if r.ContractDocID != nil {
			t.Logf("Compare: %s == %s", *r.ContractDocID, doc.ID)
		}
		if r.Source == rates.RateSourceContractPDF && r.ContractDocID != nil && *r.ContractDocID == doc.ID {
			foundConfirmed = true
			assert.Equal(t, 2800.0, r.OceanFreight)
			assert.Equal(t, 3200.0, r.TotalBuyPrice)
			break
		}
	}
	assert.True(t, foundConfirmed, "should find confirmed rate from contract in search results")

	// c) Flagged rate should be in the review queue
	reviewItems, err := contractsRepo.ListReviewItems(ctx, orgID, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(reviewItems), 1)

	foundReviewItem := false
	var pendingItemID string
	for _, ri := range reviewItems {
		if ri.ContractDocID == doc.ID {
			foundReviewItem = true
			pendingItemID = ri.ID
			assert.Equal(t, contracts.ReviewStatusPending, ri.Status)
			assert.Equal(t, 62, ri.Confidence)
			assert.Contains(t, ri.ReviewFlags, "PRICE_ANOMALY")

			// Check extracted data JSON
			var extracted rates.CanonicalRate
			err = json.Unmarshal(ri.ExtractedData, &extracted)
			require.NoError(t, err)
			assert.Equal(t, 12400.0, extracted.OceanFreight)
			break
		}
	}
	assert.True(t, foundReviewItem, "should find flagged review item in review queue")

	// d) Test human approval/correction of the flagged item
	correctedRatePayload := rates.CanonicalRate{
		OriginPort:         "INNSA",
		DestinationPort:    "AEJEA",
		CarrierSCAC:        "MAEU",
		CarrierName:        "Maersk",
		EquipmentType:      "40GP",
		OceanFreight:       3800.0, // Corrected price
		OriginCharges:      150.0,
		DestinationCharges: 180.0,
		TotalBuyPrice:      4130.0,
		ValidFrom:          time.Now().UTC(),
		ValidUntil:         time.Now().UTC().Add(48 * time.Hour),
	}
	correctedBytes, err := json.Marshal(correctedRatePayload)
	require.NoError(t, err)

	err = contractsSvc.ApproveReviewItem(ctx, orgID, pendingItemID, userID, correctedBytes, "Corrected ocean freight rate from Maersk contract sheet.")
	require.NoError(t, err)

	// Verify status in DB updated to CORRECTED
	updatedReviewItem, err := contractsRepo.GetReviewItemByID(ctx, orgID, pendingItemID)
	require.NoError(t, err)
	assert.Equal(t, contracts.ReviewStatusCorrected, updatedReviewItem.Status)

	// Verify corrected rate is now searchable in rate_entries
	rateSearchResult, err := rateSvc.SearchRates(ctx, rates.RateQuery{
		OrgID:           orgID,
		OriginPort:      "INNSA",
		DestinationPort: "AEJEA",
		EquipmentType:   "40GP",
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, rateSearchResult.TotalCount, 1)
	assert.Equal(t, 3800.0, rateSearchResult.Rates[0].OceanFreight)
	assert.Equal(t, 4130.0, rateSearchResult.Rates[0].TotalBuyPrice)
}

// mockAIBridge implements contracts.AIBridge for unit testing
type mockAIBridge struct{}

func (b *mockAIBridge) TriggerProcessing(ctx context.Context, req contracts.ProcessingRequest) error {
	return nil // mock success
}

func (b *mockAIBridge) TriggerResumption(ctx context.Context, req contracts.ResumptionRequest) error {
	return nil // mock success
}

