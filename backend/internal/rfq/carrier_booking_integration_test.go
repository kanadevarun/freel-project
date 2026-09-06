package rfq_test

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
	"github.com/freel/backend/internal/rfq"
	"github.com/freel/backend/internal/rfq/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCarrierRepoForRFQ implements repository.CarrierRepository for testing
type mockCarrierRepoForRFQ struct {
	integrations []carrierDomain.CarrierIntegration
}

func (m *mockCarrierRepoForRFQ) ListProviders(ctx context.Context) ([]carrierDomain.CarrierProvider, error) {
	return nil, nil
}
func (m *mockCarrierRepoForRFQ) GetProviderByCode(ctx context.Context, code string) (*carrierDomain.CarrierProvider, error) {
	return nil, nil
}
func (m *mockCarrierRepoForRFQ) GetProviderBySCAC(ctx context.Context, scac string) (*carrierDomain.CarrierProvider, error) {
	return nil, nil
}
func (m *mockCarrierRepoForRFQ) ListIntegrations(ctx context.Context, orgID int64) ([]carrierDomain.CarrierIntegration, error) {
	var list []carrierDomain.CarrierIntegration
	for _, ci := range m.integrations {
		if ci.OrgID == orgID {
			ci.UnmarshalRuntimeFields()
			list = append(list, ci)
		}
	}
	return list, nil
}
func (m *mockCarrierRepoForRFQ) GetIntegrationByID(ctx context.Context, orgID int64, id int64) (*carrierDomain.CarrierIntegration, error) {
	for _, ci := range m.integrations {
		if ci.OrgID == orgID && ci.ID == id {
			ci.UnmarshalRuntimeFields()
			return &ci, nil
		}
	}
	return nil, nil
}
func (m *mockCarrierRepoForRFQ) GetIntegrationBySCAC(ctx context.Context, orgID int64, scac string, env carrierDomain.Environment) (*carrierDomain.CarrierIntegration, error) {
	for _, ci := range m.integrations {
		if ci.OrgID == orgID && ci.CarrierSCAC == scac && ci.Environment == env {
			ci.UnmarshalRuntimeFields()
			return &ci, nil
		}
	}
	return nil, nil
}
func (m *mockCarrierRepoForRFQ) CreateIntegration(ctx context.Context, ci *carrierDomain.CarrierIntegration) error {
	return nil
}
func (m *mockCarrierRepoForRFQ) UpdateIntegration(ctx context.Context, ci *carrierDomain.CarrierIntegration) error {
	return nil
}
func (m *mockCarrierRepoForRFQ) DeleteIntegration(ctx context.Context, orgID int64, id int64) error {
	return nil
}
func (m *mockCarrierRepoForRFQ) UpdateStatus(ctx context.Context, orgID int64, id int64, status carrierDomain.ConnectionStatus, lastError *string) error {
	return nil
}
func (m *mockCarrierRepoForRFQ) UpdateSyncMetadata(ctx context.Context, orgID int64, id int64, syncStatus string, success bool, lastErr *string) error {
	return nil
}
func (m *mockCarrierRepoForRFQ) RecordAuditLog(ctx context.Context, orgID int64, userID *int64, action string, resourceID int64, details map[string]interface{}) error {
	return nil
}
func (m *mockCarrierRepoForRFQ) CreateSyncJob(ctx context.Context, job *carrierDomain.IntegrationSyncJob) error {
	return nil
}
func (m *mockCarrierRepoForRFQ) UpdateSyncJob(ctx context.Context, job *carrierDomain.IntegrationSyncJob) error {
	return nil
}
func (m *mockCarrierRepoForRFQ) GetSyncJobByID(ctx context.Context, orgID int64, jobID int64) (*carrierDomain.IntegrationSyncJob, error) {
	return nil, nil
}
func (m *mockCarrierRepoForRFQ) GetRunningSyncJob(ctx context.Context, orgID int64, integrationID int64) (*carrierDomain.IntegrationSyncJob, error) {
	return nil, nil
}
func (m *mockCarrierRepoForRFQ) ListSyncJobs(ctx context.Context, orgID int64, filter carrierDomain.SyncHistoryFilter) ([]carrierDomain.IntegrationSyncJob, int, error) {
	return nil, 0, nil
}
func (m *mockCarrierRepoForRFQ) CreateWebhookEvent(ctx context.Context, evt *carrierDomain.CarrierWebhookEvent) error {
	return nil
}
func (m *mockCarrierRepoForRFQ) GetWebhookEventByFingerprint(ctx context.Context, orgID int64, fingerprint string) (*carrierDomain.CarrierWebhookEvent, error) {
	return nil, nil
}
func (m *mockCarrierRepoForRFQ) UpdateWebhookEventStatus(ctx context.Context, orgID int64, eventID int64, status carrierDomain.WebhookEventStatus, errMsg *string) error {
	return nil
}
func (m *mockCarrierRepoForRFQ) ListWebhookEvents(ctx context.Context, orgID int64, integrationID *int64, limit int) ([]carrierDomain.CarrierWebhookEvent, error) {
	return nil, nil
}

// memoryBookingDL provides in-memory Datalayer for booking unit tests
type memoryBookingDL struct {
	rfq.Datalayer
	bookings map[int64]*spec.RFQBooking
}

func newMemoryBookingDL() *memoryBookingDL {
	return &memoryBookingDL{
		bookings: make(map[int64]*spec.RFQBooking),
	}
}

func (m *memoryBookingDL) GetBookingByIDOnly(ctx context.Context, orgID int32, bookingID int64) (*spec.RFQBooking, error) {
	b, ok := m.bookings[bookingID]
	if !ok || b.OrgID != int64(orgID) {
		return nil, nil
	}
	cpy := *b
	return &cpy, nil
}

func (m *memoryBookingDL) UpdateCarrierBookingResult(ctx context.Context, orgID int32, bookingID int64, carrierRef, carrierStatus string, confirmationRef, carrierError *string, bookedAt *time.Time, vesselName, voyageNum *string, etd, eta *time.Time, newStatus string) error {
	b, ok := m.bookings[bookingID]
	if !ok || b.OrgID != int64(orgID) {
		return nil
	}
	if carrierRef != "" {
		b.CarrierBookingReference = &carrierRef
	}
	b.CarrierBookingStatus = &carrierStatus
	b.CarrierConfirmationReference = confirmationRef
	b.CarrierBookingError = carrierError
	b.CarrierBookedAt = bookedAt
	if vesselName != nil {
		b.VesselName = vesselName
	}
	if voyageNum != nil {
		b.VoyageNumber = voyageNum
	}
	if etd != nil {
		b.ETD = etd
	}
	if eta != nil {
		b.ETA = eta
	}
	if newStatus != "" {
		b.Status = newStatus
	}
	return nil
}

func (m *memoryBookingDL) CreateActivity(ctx context.Context, orgID int32, entityType string, entityID int64, action, description, actor string) error {
	return nil
}

func (m *memoryBookingDL) GetBookingWorkspaceDetail(ctx context.Context, orgID int32, bookingID int64) (*spec.BookingDetailResponse, error) {
	b, _ := m.GetBookingByIDOnly(ctx, orgID, bookingID)
	if b == nil {
		return nil, nil
	}
	return &spec.BookingDetailResponse{
		Booking: *b,
	}, nil
}

func TestCarrierBookingEngine_SubmitAndSync(t *testing.T) {
	ctx := context.Background()
	testOrgID := int32(18133)
	bookingID := int64(7001)

	// Mock Maersk API Gateway for Booking endpoints
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/booking/v1/bookings") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"booking_number": "MSK-BKG-889912",
				"status": "CONFIRMED",
				"confirmation_ref": "CONF-99210",
				"vessel_name": "MAERSK MC-KINNEY MOLLER",
				"voyage_number": "2608W",
				"etd": "2026-08-20T08:00:00Z",
				"eta": "2026-09-10T18:00:00Z"
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
	capsJSON := `["TRACKING", "RATES", "BOOKING"]`

	// Create test integrations for org
	repo := &mockCarrierRepoForRFQ{
		integrations: []carrierDomain.CarrierIntegration{
			{
				ID:                   101,
				OrgID:                int64(testOrgID),
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
	dl := newMemoryBookingDL()

	scac := "MAEU"
	dl.bookings[bookingID] = &spec.RFQBooking{
		ID:              bookingID,
		OrgID:           int64(testOrgID),
		RFQID:           12001,
		BookingNumber:   "BKG-18133-7001",
		CarrierName:     "Maersk",
		CarrierSCAC:     &scac,
		Status:          spec.BookingStatusRequested,
		OriginPort:      "INNSA",
		DestinationPort: "NLRTM",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	engine := rfq.NewCarrierBookingEngine(dl, carrierSvc)

	t.Run("Submit carrier booking successfully", func(t *testing.T) {
		req := spec.BookWithCarrierRequest{
			BookingID: bookingID,
		}

		resp, err := engine.SubmitCarrierBooking(ctx, testOrgID, bookingID, req, "Test User")
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.Equal(t, spec.BookingStatusConfirmed, resp.Booking.Status)
		assert.NotNil(t, resp.Booking.CarrierBookingReference)
		assert.Equal(t, "MSK-BKG-889912", *resp.Booking.CarrierBookingReference)
		assert.Equal(t, "CONFIRMED", *resp.Booking.CarrierBookingStatus)
		assert.NotNil(t, resp.Booking.CarrierBookedAt)
		assert.NotNil(t, resp.Booking.VesselName)
		assert.Equal(t, "MAERSK MC-KINNEY MOLLER", *resp.Booking.VesselName)
	})

	t.Run("Idempotency: duplicate submit does not re-book", func(t *testing.T) {
		req := spec.BookWithCarrierRequest{
			BookingID: bookingID,
		}

		firstRef := *dl.bookings[bookingID].CarrierBookingReference

		resp, err := engine.SubmitCarrierBooking(ctx, testOrgID, bookingID, req, "Test User")
		require.NoError(t, err)
		assert.Equal(t, firstRef, *resp.Booking.CarrierBookingReference)
	})

	t.Run("Sync carrier booking status", func(t *testing.T) {
		resp, err := engine.SyncCarrierBooking(ctx, testOrgID, bookingID, "Test User")
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "CONFIRMED", *resp.Booking.CarrierBookingStatus)
		assert.Equal(t, "MSK-BKG-889912", *resp.Booking.CarrierBookingReference)
	})

	t.Run("Unconnected carrier returns configuration error", func(t *testing.T) {
		unconnectedBookingID := int64(7002)
		hlcuSCAC := "HLCU"
		dl.bookings[unconnectedBookingID] = &spec.RFQBooking{
			ID:              unconnectedBookingID,
			OrgID:           int64(testOrgID),
			RFQID:           12002,
			BookingNumber:   "BKG-18133-7002",
			CarrierName:     "Hapag-Lloyd",
			CarrierSCAC:     &hlcuSCAC,
			Status:          spec.BookingStatusDraft,
			OriginPort:      "INNSA",
			DestinationPort: "DEHAM",
		}

		req := spec.BookWithCarrierRequest{
			BookingID: unconnectedBookingID,
		}

		resp, err := engine.SubmitCarrierBooking(ctx, testOrgID, unconnectedBookingID, req, "Test User")
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "not configured")
	})

	t.Run("ResolveCarrierSCAC maps common carrier names correctly", func(t *testing.T) {
		assert.Equal(t, "MAEU", rfq.ResolveCarrierSCAC("", "Maersk Line"))
		assert.Equal(t, "MSCU", rfq.ResolveCarrierSCAC("", "MSC Mediterranean Shipping"))
		assert.Equal(t, "HLCU", rfq.ResolveCarrierSCAC("", "Hapag-Lloyd AG"))
		assert.Equal(t, "CMDU", rfq.ResolveCarrierSCAC("", "CMA CGM Group"))
		assert.Equal(t, "ONEY", rfq.ResolveCarrierSCAC("", "Ocean Network Express (ONE)"))
		assert.Equal(t, "EGLV", rfq.ResolveCarrierSCAC("", "Evergreen Marine"))
	})
}
