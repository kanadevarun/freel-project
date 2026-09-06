package shipments_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	carrierDomain "github.com/freel/backend/internal/carrier/domain"
	carrierRepoPkg "github.com/freel/backend/internal/carrier/repository"
	carrierSvcPkg "github.com/freel/backend/internal/carrier/service"
	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/config"
	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/shipments"
	"github.com/freel/backend/internal/shipments/spec"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTrackingTest(t *testing.T) (*sqlx.DB, shipments.Service, carrierSvcPkg.CarrierService) {
	cfg := config.LoadConfig()
	db, err := database.Connect(cfg.DatabaseURL)
	require.NoError(t, err)

	eventBus := events.NewInProcessBus()
	shipmentsRepo := shipments.NewRepository(db)
	shipmentsSvc := shipments.NewService(shipmentsRepo, db, eventBus, "http://localhost:8080")

	carrierRepo := carrierRepoPkg.NewCarrierRepository(db)
	carrierSvc := carrierSvcPkg.NewCarrierService(carrierRepo, "TestTrackingSecretKey32BytesLong!")
	shipmentsSvc.SetCarrierService(carrierSvc)

	return db, shipmentsSvc, carrierSvc
}

func TestCarrierTrackingIntegrationFlow(t *testing.T) {
	db, shipmentsSvc, carrierSvc := setupTrackingTest(t)
	defer db.Close()

	ctx := context.Background()
	testOrgID := int64(888101)
	testUserID := int64(777101)

	// Clean up test org
	_, _ = db.Exec("DELETE FROM shipment_milestones WHERE shipment_id IN (SELECT id FROM shipments WHERE org_id = ?)", testOrgID)
	_, _ = db.Exec("DELETE FROM carrier_tracking_events WHERE org_id = ?", testOrgID)
	_, _ = db.Exec("DELETE FROM tracking_positions WHERE org_id = ?", testOrgID)
	_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id = ?", testOrgID)
	_, _ = db.Exec("DELETE FROM shipments WHERE org_id = ?", testOrgID)
	_, _ = db.Exec("DELETE FROM organizations WHERE id = ?", testOrgID)

	_, err := db.Exec("INSERT INTO organizations (id, name) VALUES (?, 'Test Tracking Org')", testOrgID)
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM shipment_milestones WHERE shipment_id IN (SELECT id FROM shipments WHERE org_id = ?)", testOrgID)
		_, _ = db.Exec("DELETE FROM carrier_tracking_events WHERE org_id = ?", testOrgID)
		_, _ = db.Exec("DELETE FROM tracking_positions WHERE org_id = ?", testOrgID)
		_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id = ?", testOrgID)
		_, _ = db.Exec("DELETE FROM shipments WHERE org_id = ?", testOrgID)
		_, _ = db.Exec("DELETE FROM organizations WHERE id = ?", testOrgID)
	}()

	// 1. Mock Carrier API Gateway (serving real DCSA v2 events)
	mockGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/track-and-trace/v2/events") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"eventID": "MAEU-EVT-1001",
					"eventType": "EQUIPMENT",
					"eventDateTime": "2026-08-10T08:00:00Z",
					"eventClassifierCode": "ACT",
					"equipmentEventTypeCode": "LOAD",
					"equipmentReference": "MSKU1092837",
					"carrierBookingReference": "BKG-MAEU-99218",
					"eventLocation": {
						"locationName": "Nhava Sheva Terminal",
						"UNLocationCode": "INNSA"
					}
				},
				{
					"eventID": "MAEU-EVT-1002",
					"eventType": "TRANSPORT",
					"eventDateTime": "2026-08-11T12:00:00Z",
					"eventClassifierCode": "ACT",
					"transportEventTypeCode": "DEPA",
					"equipmentReference": "MSKU1092837",
					"carrierBookingReference": "BKG-MAEU-99218",
					"transportCall": {
						"carrierVoyageNumber": "2608W",
						"UNLocationCode": "INNSA",
						"vessel": {
							"vesselName": "MAERSK MC-KINNEY MOLLER"
						},
						"location": {
							"locationName": "Nhava Sheva"
						}
					}
				},
				{
					"eventID": "MAEU-EVT-1003",
					"eventType": "TRANSPORT",
					"eventDateTime": "2026-08-26T16:00:00Z",
					"eventClassifierCode": "EST",
					"transportEventTypeCode": "ARRI",
					"equipmentReference": "MSKU1092837",
					"carrierBookingReference": "BKG-MAEU-99218",
					"transportCall": {
						"carrierVoyageNumber": "2608W",
						"UNLocationCode": "NLRTM",
						"location": {
							"locationName": "Port of Rotterdam"
						}
					}
				}
			]`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer mockGateway.Close()

	// 2. Connect Maersk integration for tenant
	connectReq := carrierDomain.ConnectCarrierRequest{
		CarrierSCAC:      "MAEU",
		Environment:      "SANDBOX",
		ConnectionMethod: "API",
		Credentials: map[string]interface{}{
			"api_key":   "maersk-test-key",
			"base_url":  mockGateway.URL,
			"client_id": "LOGISTIQ-TEST",
		},
		Capabilities: []string{"TRACKING", "RATES", "BOOKING"},
	}
	integration, err := carrierSvc.ConnectCarrier(ctx, testOrgID, testUserID, connectReq)
	require.NoError(t, err)
	require.NotNil(t, integration)

	// 3. Create a test Shipment with carrier MAEU
	bkgNum := "BKG-MAEU-99218"
	mblNum := "MBL-MAEU-777"
	shipment := &spec.Shipment{
		OrgID:            testOrgID,
		CarrierSCAC:      "MAEU",
		BookingNumber:    &bkgNum,
		MBLNumber:        &mblNum,
		ContainerNumbers: spec.JSONStringSlice{"MSKU1092837"},
		Status:           spec.BOOKED,
		OriginPort:       "INNSA",
		DestinationPort:  "NLRTM",
	}

	repo := shipments.NewRepository(db)
	err = repo.CreateShipment(ctx, shipment)
	require.NoError(t, err)
	require.Greater(t, shipment.ID, int64(0))

	// 4. Trigger Manual "Sync Tracking" Action
	res, err := shipmentsSvc.RefreshShipmentTracking(ctx, testOrgID, shipment.ID, &testUserID, "Test Freight Operator")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, res.Success)
	assert.Equal(t, spec.TrackingFreshnessLive, res.DataFreshness)
	assert.Contains(t, res.Message, "Successfully synchronized")

	// 5. Verify Ingested Events in Database
	var eventCount int
	err = db.Get(&eventCount, "SELECT COUNT(*) FROM carrier_tracking_events WHERE org_id = ? AND shipment_id = ?", testOrgID, shipment.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, eventCount, "All 3 DCSA events should be persisted into carrier_tracking_events")

	// 6. Verify Shipment State Updates (Status Advanced, Vessel, ETA, ETD)
	updatedSh, err := repo.GetShipmentByID(ctx, testOrgID, shipment.ID)
	require.NoError(t, err)
	require.NotNil(t, updatedSh)

	assert.Equal(t, spec.DEPARTED, updatedSh.Status, "Shipment status should advance to DEPARTED")
	require.NotNil(t, updatedSh.ETA)
	assert.Equal(t, 26, updatedSh.ETA.Day(), "ETA should be updated to August 26")

	// 7. Verify Idempotency: Run synchronization 3 more times
	for i := 0; i < 3; i++ {
		repeatRes, err := shipmentsSvc.RefreshShipmentTracking(ctx, testOrgID, shipment.ID, &testUserID, "Test Scheduler")
		require.NoError(t, err)
		assert.True(t, repeatRes.Success)
	}

	var postRepeatCount int
	err = db.Get(&postRepeatCount, "SELECT COUNT(*) FROM carrier_tracking_events WHERE org_id = ? AND shipment_id = ?", testOrgID, shipment.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, postRepeatCount, "Idempotency check: Repeated sync must NOT create duplicate tracking events")

	// 8. Verify Carrier Outage / Failure Preservation
	failingGateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error": "Maersk API gateway temporarily overloaded"}`))
	}))
	defer failingGateway.Close()

	// Update integration credentials to point to failing gateway
	_, err = carrierSvc.UpdateCarrier(ctx, testOrgID, testUserID, integration.ID, carrierDomain.UpdateCarrierRequest{
		Credentials: map[string]interface{}{
			"api_key":  "maersk-test-key",
			"base_url": failingGateway.URL,
		},
	})
	require.NoError(t, err)

	failRes, err := shipmentsSvc.RefreshShipmentTracking(ctx, testOrgID, shipment.ID, &testUserID, "Manual Retry")
	require.NoError(t, err)
	require.NotNil(t, failRes)
	assert.True(t, failRes.UsedFallback, "Failure must fallback safely to persisted operational records")
	assert.Contains(t, failRes.Message, "Existing tracking history is preserved")

	// Verify events are STILL intact after failure
	var postFailCount int
	err = db.Get(&postFailCount, "SELECT COUNT(*) FROM carrier_tracking_events WHERE org_id = ? AND shipment_id = ?", testOrgID, shipment.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, postFailCount, "Carrier failure must NEVER delete historical tracking events")
}

func TestUnconfiguredCarrierAndCapabilityGuards(t *testing.T) {
	db, shipmentsSvc, carrierSvc := setupTrackingTest(t)
	defer db.Close()

	ctx := context.Background()
	testOrgID := int64(888102)
	testUserID := int64(777102)

	// Clean up
	_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id = ?", testOrgID)
	_, _ = db.Exec("DELETE FROM shipments WHERE org_id = ?", testOrgID)
	_, _ = db.Exec("DELETE FROM organizations WHERE id = ?", testOrgID)

	_, err := db.Exec("INSERT INTO organizations (id, name) VALUES (?, 'Test Unconfigured Org')", testOrgID)
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id = ?", testOrgID)
		_, _ = db.Exec("DELETE FROM shipments WHERE org_id = ?", testOrgID)
		_, _ = db.Exec("DELETE FROM organizations WHERE id = ?", testOrgID)
	}()

	// 1. Create a shipment with an unconnected carrier (e.g. HLCU)
	bkg := "BKG-HLCU-001"
	sh := &spec.Shipment{
		OrgID:            testOrgID,
		CarrierSCAC:      "HLCU",
		BookingNumber:    &bkg,
		ContainerNumbers: spec.JSONStringSlice{"HLXU1234567"},
		Status:           spec.BOOKED,
		OriginPort:       "INNSA",
		DestinationPort:  "DEHAM",
	}
	repo := shipments.NewRepository(db)
	err = repo.CreateShipment(ctx, sh)
	require.NoError(t, err)

	// 2. Sync should return informative message that HLCU is not connected
	res, err := shipmentsSvc.RefreshShipmentTracking(ctx, testOrgID, sh.ID, &testUserID, "User")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Contains(t, res.Message, "Carrier integration (HLCU) is not configured")
	assert.True(t, res.UsedFallback)

	// 3. Connect HLCU but with TRACKING capability omitted (Only RATES enabled)
	connectReq := carrierDomain.ConnectCarrierRequest{
		CarrierSCAC:      "HLCU",
		Environment:      "SANDBOX",
		ConnectionMethod: "API",
		Credentials: map[string]interface{}{
			"api_key": "hlcu-key",
		},
		Capabilities: []string{"RATES"}, // No TRACKING
	}
	_, err = carrierSvc.ConnectCarrier(ctx, testOrgID, testUserID, connectReq)
	require.NoError(t, err)

	// 4. Sync should return informative message that TRACKING is not enabled
	resCap, err := shipmentsSvc.RefreshShipmentTracking(ctx, testOrgID, sh.ID, &testUserID, "User")
	require.NoError(t, err)
	assert.Contains(t, resCap.Message, "Tracking capability is not enabled for carrier connection (HLCU)")
}

func TestTenantIsolationInTrackingSync(t *testing.T) {
	db, shipmentsSvc, _ := setupTrackingTest(t)
	defer db.Close()

	ctx := context.Background()
	orgA := int64(888103)
	orgB := int64(888104)

	_, _ = db.Exec("DELETE FROM shipments WHERE org_id IN (?, ?)", orgA, orgB)
	_, _ = db.Exec("DELETE FROM organizations WHERE id IN (?, ?)", orgA, orgB)

	_, _ = db.Exec("INSERT INTO organizations (id, name) VALUES (?, 'Org A'), (?, 'Org B')", orgA, orgB)
	defer func() {
		_, _ = db.Exec("DELETE FROM shipments WHERE org_id IN (?, ?)", orgA, orgB)
		_, _ = db.Exec("DELETE FROM organizations WHERE id IN (?, ?)", orgA, orgB)
	}()

	bkgA := "BKG-A-100"
	shA := &spec.Shipment{
		OrgID:            orgA,
		CarrierSCAC:      "MAEU",
		BookingNumber:    &bkgA,
		ContainerNumbers: spec.JSONStringSlice{"MSKU9990001"},
		Status:           spec.BOOKED,
		OriginPort:       "INNSA",
		DestinationPort:  "NLRTM",
	}
	repo := shipments.NewRepository(db)
	err := repo.CreateShipment(ctx, shA)
	require.NoError(t, err)

	// Org B attempting to sync Org A's shipment must fail with Resource Not Found (Tenant Isolation)
	userID := int64(999)
	_, err = shipmentsSvc.RefreshShipmentTracking(ctx, orgB, shA.ID, &userID, "Attacker")
	assert.Error(t, err, "Tenant B must not be able to refresh or access Tenant A's shipment")
}
