package rates_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	carrierDomain "github.com/freel/backend/internal/carrier/domain"
	carrierService "github.com/freel/backend/internal/carrier/service"
	"github.com/freel/backend/internal/common/crypto"
	"github.com/freel/backend/internal/rates"
	"github.com/freel/backend/internal/rates/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCarrierRepo implements repository.CarrierRepository for testing
type mockCarrierRepo struct {
	integrations []carrierDomain.CarrierIntegration
}

func (m *mockCarrierRepo) ListProviders(ctx context.Context) ([]carrierDomain.CarrierProvider, error) {
	return nil, nil
}
func (m *mockCarrierRepo) GetProviderByCode(ctx context.Context, code string) (*carrierDomain.CarrierProvider, error) {
	return nil, nil
}
func (m *mockCarrierRepo) GetProviderBySCAC(ctx context.Context, scac string) (*carrierDomain.CarrierProvider, error) {
	return nil, nil
}
func (m *mockCarrierRepo) ListIntegrations(ctx context.Context, orgID int64) ([]carrierDomain.CarrierIntegration, error) {
	var list []carrierDomain.CarrierIntegration
	for _, ci := range m.integrations {
		if ci.OrgID == orgID {
			ci.UnmarshalRuntimeFields()
			list = append(list, ci)
		}
	}
	return list, nil
}
func (m *mockCarrierRepo) GetIntegrationByID(ctx context.Context, orgID int64, id int64) (*carrierDomain.CarrierIntegration, error) {
	for _, ci := range m.integrations {
		if ci.OrgID == orgID && ci.ID == id {
			ci.UnmarshalRuntimeFields()
			return &ci, nil
		}
	}
	return nil, nil
}
func (m *mockCarrierRepo) GetIntegrationBySCAC(ctx context.Context, orgID int64, scac string, env carrierDomain.Environment) (*carrierDomain.CarrierIntegration, error) {
	for _, ci := range m.integrations {
		if ci.OrgID == orgID && ci.CarrierSCAC == scac && ci.Environment == env {
			ci.UnmarshalRuntimeFields()
			return &ci, nil
		}
	}
	return nil, nil
}
func (m *mockCarrierRepo) CreateIntegration(ctx context.Context, ci *carrierDomain.CarrierIntegration) error {
	return nil
}
func (m *mockCarrierRepo) UpdateIntegration(ctx context.Context, ci *carrierDomain.CarrierIntegration) error {
	return nil
}
func (m *mockCarrierRepo) DeleteIntegration(ctx context.Context, orgID int64, id int64) error {
	return nil
}
func (m *mockCarrierRepo) UpdateStatus(ctx context.Context, orgID int64, id int64, status carrierDomain.ConnectionStatus, lastError *string) error {
	return nil
}
func (m *mockCarrierRepo) UpdateSyncMetadata(ctx context.Context, orgID int64, id int64, syncStatus string, success bool, lastErr *string) error {
	return nil
}
func (m *mockCarrierRepo) RecordAuditLog(ctx context.Context, orgID int64, userID *int64, action string, resourceID int64, details map[string]interface{}) error {
	return nil
}
func (m *mockCarrierRepo) CreateSyncJob(ctx context.Context, job *carrierDomain.IntegrationSyncJob) error {
	return nil
}
func (m *mockCarrierRepo) UpdateSyncJob(ctx context.Context, job *carrierDomain.IntegrationSyncJob) error {
	return nil
}
func (m *mockCarrierRepo) GetSyncJobByID(ctx context.Context, orgID int64, jobID int64) (*carrierDomain.IntegrationSyncJob, error) {
	return nil, nil
}
func (m *mockCarrierRepo) GetRunningSyncJob(ctx context.Context, orgID int64, integrationID int64) (*carrierDomain.IntegrationSyncJob, error) {
	return nil, nil
}
func (m *mockCarrierRepo) ListSyncJobs(ctx context.Context, orgID int64, filter carrierDomain.SyncHistoryFilter) ([]carrierDomain.IntegrationSyncJob, int, error) {
	return nil, 0, nil
}
func (m *mockCarrierRepo) CreateWebhookEvent(ctx context.Context, evt *carrierDomain.CarrierWebhookEvent) error {
	return nil
}
func (m *mockCarrierRepo) GetWebhookEventByFingerprint(ctx context.Context, orgID int64, fingerprint string) (*carrierDomain.CarrierWebhookEvent, error) {
	return nil, nil
}
func (m *mockCarrierRepo) UpdateWebhookEventStatus(ctx context.Context, orgID int64, eventID int64, status carrierDomain.WebhookEventStatus, errMsg *string) error {
	return nil
}
func (m *mockCarrierRepo) ListWebhookEvents(ctx context.Context, orgID int64, integrationID *int64, limit int) ([]carrierDomain.CarrierWebhookEvent, error) {
	return nil, nil
}

func TestCarrierRatesEngine_SearchLiveRates(t *testing.T) {
	ctx := context.Background()
	testOrgID := int64(18133)

	// Mock Maersk API server returning real rate quotes
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/rates/v1/spot-rates") || strings.Contains(r.URL.Path, "/rates/v1/contract-rates") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"rates": [
					{
						"priceId": "MSK-RT-1001",
						"origin": "INNSA",
						"destination": "NLRTM",
						"commodity": "General Cargo",
						"equipmentType": "40HC",
						"currency": "USD",
						"baseOceanPrice": 2100.00,
						"originSurcharge": 250.00,
						"destSurcharge": 350.00,
						"totalPrice": 2700.00,
						"transitTimeDays": 22,
						"freeDays": 14,
						"validityStart": "2026-08-01T00:00:00Z",
						"validityEnd": "2026-09-30T23:59:59Z",
						"isContract": false
					},
					{
						"priceId": "MSK-RT-1002",
						"origin": "INNSA",
						"destination": "NLRTM",
						"commodity": "General Cargo",
						"equipmentType": "40HC",
						"currency": "USD",
						"baseOceanPrice": 2400.00,
						"originSurcharge": 250.00,
						"destSurcharge": 350.00,
						"totalPrice": 3000.00,
						"transitTimeDays": 18,
						"freeDays": 21,
						"validityStart": "2026-08-01T00:00:00Z",
						"validityEnd": "2026-09-30T23:59:59Z",
						"isContract": false
					}
				]
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	encryptionKey := "12345678901234567890123456789012"
	carrierName := "A.P. Moller – Maersk"
	rawCreds := fmt.Sprintf(`{"api_key":"test-maersk-key","base_url":"%s"}`, mockServer.URL)
	encCreds, err := crypto.Encrypt(rawCreds, encryptionKey)
	require.NoError(t, err)
	capsJSON := `["TRACKING", "RATES", "SPOT_RATES", "BOOKING"]`

	// Create test integrations for org
	repo := &mockCarrierRepo{
		integrations: []carrierDomain.CarrierIntegration{
			{
				ID:                   101,
				OrgID:                testOrgID,
				CarrierSCAC:          "MAEU",
				CarrierName:          &carrierName,
				ConnectionStatus:     carrierDomain.StatusConnected,
				Environment:          carrierDomain.EnvSandbox,
				EncryptedCredentials: &encCreds,
				CapabilitiesJSON:     &capsJSON,
				IsEnabled:            true,
				CreatedAt:            time.Now(),
				UpdatedAt:            time.Now(),
			},
		},
	}

	carrierSvc := carrierService.NewCarrierService(repo, encryptionKey)
	engine := rates.NewCarrierRatesEngine(carrierSvc)

	t.Run("Query live rates across connected carriers", func(t *testing.T) {
		req := spec.CarrierRateSearchRequest{
			OriginPort:      "INNSA",
			DestinationPort: "NLRTM",
			EquipmentType:   "40HC",
		}

		resp, err := engine.SearchLiveRates(ctx, testOrgID, req)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "INNSA", resp.OriginPort)
		assert.Equal(t, "NLRTM", resp.DestinationPort)
		assert.Equal(t, "40HC", resp.EquipmentType)
		assert.Equal(t, 2, resp.TotalRatesCount)
		assert.NotNil(t, resp.CheapestCarrier)
		assert.NotNil(t, resp.CheapestAmount)
		assert.Equal(t, 2700.0, *resp.CheapestAmount)
		assert.NotNil(t, resp.FastestCarrier)
		assert.Equal(t, 18, *resp.FastestTransit)

		// Verify rates have preserved currencies
		for _, r := range resp.Rates {
			assert.Equal(t, "USD", r.Currency)
			assert.Greater(t, r.TotalBuyPrice, float64(0))
			assert.False(t, r.ValidFrom.IsZero())
			assert.False(t, r.ValidUntil.IsZero())
		}
	})

	t.Run("Filter by RateType SPOT vs CONTRACT", func(t *testing.T) {
		reqSpot := spec.CarrierRateSearchRequest{
			OriginPort:      "INNSA",
			DestinationPort: "NLRTM",
			EquipmentType:   "40HC",
			RateType:        "SPOT",
		}
		respSpot, err := engine.SearchLiveRates(ctx, testOrgID, reqSpot)
		require.NoError(t, err)
		assert.True(t, respSpot.Success)
		for _, r := range respSpot.Rates {
			assert.False(t, r.IsContractRate)
		}
	})

	t.Run("Unconnected carrier returns diagnostic message", func(t *testing.T) {
		req := spec.CarrierRateSearchRequest{
			OriginPort:      "INNSA",
			DestinationPort: "NLRTM",
			EquipmentType:   "40HC",
			CarrierSCAC:     "HLCU", // Hapag-Lloyd not connected for this org
		}
		resp, err := engine.SearchLiveRates(ctx, testOrgID, req)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, 0, resp.TotalRatesCount)
		assert.Contains(t, resp.Message, "not configured")
	})

	t.Run("Tenant isolation: zero-data org sees 0 carrier rates", func(t *testing.T) {
		zeroOrgID := int64(99999)
		req := spec.CarrierRateSearchRequest{
			OriginPort:      "INNSA",
			DestinationPort: "NLRTM",
			EquipmentType:   "40HC",
		}
		resp, err := engine.SearchLiveRates(ctx, zeroOrgID, req)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, 0, resp.TotalRatesCount)
		assert.Contains(t, resp.Message, "No carrier integrations configured")
	})
}
