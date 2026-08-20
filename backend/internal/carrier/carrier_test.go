package carrier_test

import (
	"context"
	"testing"

	"github.com/freel/backend/internal/carrier"
	"github.com/freel/backend/internal/carrier/adapters"
	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/rates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCarrierIntegrationFoundation(t *testing.T) {
	// 1. Connect to local development database
	dbURL := "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC&multiStatements=true"
	db, err := database.Connect(dbURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	orgID := int64(9999) // test specific tenant ID

	// Seed org and carriers for FK constraints
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Test Carrier Org', NOW(), NOW()) ON DUPLICATE KEY UPDATE name = VALUES(name)", orgID)
	_, _ = db.Exec("INSERT INTO carriers (scac, name) VALUES ('MAEU', 'Maersk'), ('MSC', 'MSC'), ('CMA', 'CMA CGM') ON DUPLICATE KEY UPDATE name = VALUES(name)")

	// Clean up any test pollution
	_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id = ?", orgID)

	// SCENARIO 5: Carrier not configured
	_, err = carrier.GetIntegrationConfig(ctx, db, orgID, "MAEU")
	assert.Error(t, err, "Should fail when no integration is configured in DB")

	// Insert test carrier integrations
	_, err = db.Exec(`
		INSERT INTO carrier_integrations (org_id, carrier_scac, is_active, created_at, updated_at)
		VALUES (?, 'MAEU', 1, NOW(), NOW()), (?, 'MSC', 1, NOW(), NOW()), (?, 'CMA', 0, NOW(), NOW())
	`, orgID, orgID, orgID)
	require.NoError(t, err)
	defer func() {
		_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id = ?", orgID)
	}()

	// SCENARIO 6: Unsupported capability & disabled integration
	disabledCfg, err := carrier.GetIntegrationConfig(ctx, db, orgID, "CMA")
	assert.Error(t, err, "Should error when integration is disabled")
	assert.Nil(t, disabledCfg)

	// Setup environment variables for Maersk and MSC configurations
	t.Setenv("CARRIER_MAEU_API_KEY_9999", "maersk-key-secure")
	t.Setenv("CARRIER_MAEU_BASE_URL_9999", "https://api.maersk.com/v2")
	t.Setenv("CARRIER_MAEU_CAPABILITIES_9999", "TRACKING,WEBHOOK")

	t.Setenv("CARRIER_MSC_API_KEY_9999", "msc-key-secure")
	t.Setenv("CARRIER_MSC_BASE_URL_9999", "https://api.msc.com/v1")
	t.Setenv("CARRIER_MSC_CAPABILITIES_9999", "TRACKING,RATES,WEBHOOK")

	// Verify configuration parsing
	maerskCfg, err := carrier.GetIntegrationConfig(ctx, db, orgID, "MAEU")
	require.NoError(t, err)
	assert.Equal(t, "maersk-key-secure", maerskCfg.APIKey)
	assert.Equal(t, "https://api.maersk.com/v2", maerskCfg.APIBaseURL)
	assert.True(t, maerskCfg.Capabilities["TRACKING"])
	assert.False(t, maerskCfg.Capabilities["RATES"], "RATES capability should be unsupported for Maersk")

	// SCENARIO 4: Invalid/missing credentials in production mode
	t.Setenv("APP_ENV", "production")
	t.Setenv("CARRIER_MAEU_API_KEY_9999", "") // clear credential
	_, err = carrier.GetIntegrationConfig(ctx, db, orgID, "MAEU")
	assert.Error(t, err, "Should error in production when API key is missing")

	// Restore dev/test environment mode
	t.Setenv("APP_ENV", "development")
	t.Setenv("CARRIER_MAEU_API_KEY_9999", "maersk-key-secure")

	// SCENARIO 8: Unknown carrier
	_, err = adapters.GetTrackingProvider(db, orgID, "XYZU")
	assert.Error(t, err, "Should error for unrecognized SCAC code")

	// Get adapters
	maerskAdapter, err := adapters.GetTrackingProvider(db, orgID, "MAEU")
	require.NoError(t, err)
	
	mscAdapter, err := adapters.GetTrackingProvider(db, orgID, "MSC")
	require.NoError(t, err)

	// SCENARIO 6: Unsupported capability test
	// We didn't enable BOOKING capability for MSC
	bookingProvider, err := adapters.GetBookingProvider(db, orgID, "MSC")
	require.NoError(t, err)
	_, err = bookingProvider.GetBooking(ctx, "B12345")
	assert.ErrorContains(t, err, "unsupported capability: BOOKING")

	// SCENARIO 4: Invalid credentials error triggers authentication failure
	t.Setenv("CARRIER_MAEU_API_KEY_9999", "invalid")
	badMaerskAdapter, err := adapters.GetTrackingProvider(db, orgID, "MAEU")
	require.NoError(t, err)
	_, err = badMaerskAdapter.GetTrackingEvents(ctx, carrier.TrackingRequest{BookingNumber: "B1234"})
	assert.ErrorContains(t, err, "API authentication failure")
	t.Setenv("CARRIER_MAEU_API_KEY_9999", "maersk-key-secure") // restore

	// SCENARIO 1: Maersk tracking response -> NormalizedTrackingEvent conversion
	events, err := maerskAdapter.GetTrackingEvents(ctx, carrier.TrackingRequest{BookingNumber: "B12345"})
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Contains(t, events[0].Description, "Vessel departed origin port")

	// Parse raw Maersk webhook payload
	maerskWebhookRaw := []byte(`{
		"container": "MSKU1234567",
		"status": "VESSEL DEPARTED",
		"event_date": "2026-08-14T17:42:38Z",
		"location": "Nhava Sheva"
	}`)
	parsedMaersk, err := maerskAdapter.(carrier.WebhookProvider).ParseWebhookPayload(maerskWebhookRaw)
	require.NoError(t, err)
	assert.Equal(t, "DEPARTED", parsedMaersk.MilestoneCode)
	assert.Equal(t, "Nhava Sheva", parsedMaersk.Location)

	// SCENARIO 2: MSC tracking response -> NormalizedTrackingEvent conversion
	mscWebhookRaw := []byte(`{
		"equipmentNo": "MSKU1234567",
		"milestone": "DEP",
		"timestamp": "2026-08-14T17:42:38Z",
		"port": "INNSA"
	}`)
	parsedMSC, err := mscAdapter.(carrier.WebhookProvider).ParseWebhookPayload(mscWebhookRaw)
	require.NoError(t, err)
	assert.Equal(t, "DEPARTED", parsedMSC.MilestoneCode)
	assert.Equal(t, "INNSA", parsedMSC.Location)

	// SCENARIO 7: Malformed carrier response
	badWebhookRaw := []byte(`{"bad": "format"}`)
	_, err = mscAdapter.(carrier.WebhookProvider).ParseWebhookPayload(badWebhookRaw)
	assert.Error(t, err, "Should fail when parsing invalid payload structure")

	// SCENARIO 3: Carrier rate response -> CanonicalRate conversion
	// Using SpotNormalizer to map rich carrier rates to CanonicalRate structures
	normalizer := rates.NewSpotNormalizer()
	richRate := carrier.RichCarrierRate{
		CarrierName:        "Maersk",
		BuyPrice:           1250.00,
		TransitDays:        22,
		FreeDays:           14,
		VesselName:         "MAERSK NEPTUNE",
		ServiceCode:        "AS1",
		OceanFreight:       1000.00,
		OriginCharges:      150.00,
		DestinationCharges: 100.00,
		CO2Emissions:       2.45,
		NauticalMiles:      2600,
	}
	canonicalRate := normalizer.Normalize(richRate, orgID, "INNSA", "DEHAM")
	assert.Equal(t, "MAEU", canonicalRate.CarrierSCAC)
	assert.Equal(t, 1250.00, canonicalRate.TotalBuyPrice)
	assert.Equal(t, 1000.00, canonicalRate.OceanFreight)
	assert.Equal(t, 22, *canonicalRate.TransitDays)
	assert.Equal(t, 14, canonicalRate.FreeDaysDestination)
}
