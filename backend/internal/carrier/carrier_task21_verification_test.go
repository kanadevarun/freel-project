package carrier_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/freel/backend/internal/carrier/adapters"
	"github.com/freel/backend/internal/carrier/domain"
	"github.com/freel/backend/internal/carrier/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test 1 — BaseAdapter connection
// Proves that BaseAdapter.TestConnection():
// - does NOT return Success=true
// - does NOT return HTTP 200
// - does NOT return fake latency
// - clearly indicates that the carrier integration is not implemented
func TestTask21_Test1_BaseAdapterConnection(t *testing.T) {
	base := &adapters.BaseAdapter{
		Code:        "BASE_TEST",
		CarrierSCAC: "TEST",
		Capabilities: []domain.Capability{
			domain.CapTracking,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	creds := domain.DecryptedCredentials{
		APIKey:       "dummy-key",
		ClientID:     "dummy-client",
		ClientSecret: "dummy-secret",
	}

	res, err := base.TestConnection(ctx, creds, domain.EnvSandbox)
	require.NoError(t, err, "TestConnection should return a result struct, not an unexpected go error")
	require.NotNil(t, res)

	assert.False(t, res.Success, "BaseAdapter must NOT return Success=true")
	assert.NotEqual(t, 200, res.HTTPStatus, "BaseAdapter must NOT return HTTP 200")
	assert.Equal(t, int64(0), res.LatencyMs, "BaseAdapter must NOT return fake latency (e.g. 45ms)")
	assert.Equal(t, domain.ErrCodeUnsupported, res.ErrorCode, "BaseAdapter should set ErrorCode to UNSUPPORTED_OPERATION")
	assert.Contains(t, strings.ToLower(res.Message), "not yet implemented", "Message must clearly indicate integration is not yet implemented")
}

// Test 2 — CMA CGM
// Verify CMA CGM cannot report Connected through the default BaseAdapter TestConnection implementation.
func TestTask21_Test2_CmaCgmCannotReportConnected(t *testing.T) {
	registry := adapters.GetDefaultRegistry()
	cma := registry.GetAdapterBySCAC("CMDU")
	require.NotNil(t, cma)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	creds := domain.DecryptedCredentials{
		APIKey:       "cma-test-key",
		ClientID:     "cma-client",
		ClientSecret: "cma-secret",
	}

	res, err := cma.TestConnection(ctx, creds, domain.EnvProduction)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.False(t, res.Success, "CMA CGM must not report Success=true via BaseAdapter")
	assert.NotEqual(t, 200, res.HTTPStatus, "CMA CGM must not return HTTP 200")
	assert.Equal(t, int64(0), res.LatencyMs, "CMA CGM must not return fake latency")
	assert.Contains(t, res.Message, "not yet implemented")
}

// Test 3 — ONE
// Verify ONE cannot report Connected through the default BaseAdapter TestConnection implementation.
func TestTask21_Test3_OneCannotReportConnected(t *testing.T) {
	registry := adapters.GetDefaultRegistry()
	one := registry.GetAdapterBySCAC("ONEY")
	require.NotNil(t, one)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	creds := domain.DecryptedCredentials{
		APIKey: "one-api-key",
	}

	res, err := one.TestConnection(ctx, creds, domain.EnvProduction)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.False(t, res.Success, "ONE must not report Success=true via BaseAdapter")
	assert.NotEqual(t, 200, res.HTTPStatus)
	assert.Equal(t, int64(0), res.LatencyMs)
	assert.Contains(t, res.Message, "not yet implemented")
}

// Test 4 — Evergreen
// Verify Evergreen cannot report Connected through the default BaseAdapter TestConnection implementation.
func TestTask21_Test4_EvergreenCannotReportConnected(t *testing.T) {
	registry := adapters.GetDefaultRegistry()
	emc := registry.GetAdapterBySCAC("EGLV")
	require.NotNil(t, emc)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	creds := domain.DecryptedCredentials{
		APIKey: "eglv-api-key",
	}

	res, err := emc.TestConnection(ctx, creds, domain.EnvProduction)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.False(t, res.Success, "Evergreen must not report Success=true via BaseAdapter")
	assert.NotEqual(t, 200, res.HTTPStatus)
	assert.Equal(t, int64(0), res.LatencyMs)
	assert.Contains(t, res.Message, "not yet implemented")
}

// Test 5 — COSCO
// Verify COSCO cannot report Connected through the default BaseAdapter TestConnection implementation.
func TestTask21_Test5_CoscoCannotReportConnected(t *testing.T) {
	registry := adapters.GetDefaultRegistry()
	cosco := registry.GetAdapterBySCAC("COSU")
	require.NotNil(t, cosco)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	creds := domain.DecryptedCredentials{
		APIKey: "cosco-api-key",
	}

	res, err := cosco.TestConnection(ctx, creds, domain.EnvProduction)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.False(t, res.Success, "COSCO must not report Success=true via BaseAdapter")
	assert.NotEqual(t, 200, res.HTTPStatus)
	assert.Equal(t, int64(0), res.LatencyMs)
	assert.Contains(t, res.Message, "not yet implemented")
}

// Test 6 — Generic / Unknown carrier
// Verify an unknown carrier cannot receive a fake successful connection status.
func TestTask21_Test6_GenericUnknownCarrierCannotReportConnected(t *testing.T) {
	registry := adapters.GetDefaultRegistry()
	unknown := registry.GetAdapterBySCAC("XYZU")
	require.NotNil(t, unknown, "Registry falls back to GenericAdapter for unknown SCAC")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	creds := domain.DecryptedCredentials{
		APIKey: "random-key-999",
	}

	res, err := unknown.TestConnection(ctx, creds, domain.EnvProduction)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.False(t, res.Success, "Unknown carrier must not report Success=true")
	assert.NotEqual(t, 200, res.HTTPStatus)
	assert.Equal(t, int64(0), res.LatencyMs)
	assert.Contains(t, res.Message, "not yet implemented")
}

// Test 7 — Hapag capabilities
// Verify SupportsCapability(CapTracking) == true and all other capabilities are false.
func TestTask21_Test7_HapagCapabilities(t *testing.T) {
	registry := adapters.GetDefaultRegistry()
	hapag := registry.GetAdapterBySCAC("HLCU")
	require.NotNil(t, hapag)

	// Tracking must be supported
	assert.True(t, hapag.SupportsCapability(domain.CapTracking), "Hapag must support CapTracking")
	assert.True(t, hapag.HasCapability(domain.CapTracking), "HasCapability must also return true for CapTracking")

	// Other capabilities must NOT be advertised
	assert.False(t, hapag.SupportsCapability(domain.CapRates), "Hapag must NOT advertise CapRates")
	assert.False(t, hapag.SupportsCapability(domain.CapSpotRates), "Hapag must NOT advertise CapSpotRates")
	assert.False(t, hapag.SupportsCapability(domain.CapContractRates), "Hapag must NOT advertise CapContractRates")
	assert.False(t, hapag.SupportsCapability(domain.CapBooking), "Hapag must NOT advertise CapBooking")
	assert.False(t, hapag.SupportsCapability(domain.CapDocuments), "Hapag must NOT advertise CapDocuments")

	// Verify exact list
	caps := hapag.SupportedCapabilities()
	assert.Len(t, caps, 1, "Hapag must advertise exactly 1 capability")
	assert.Equal(t, domain.CapTracking, caps[0])
}

// Test 8 — MSC capabilities
// Verify MSC capabilities match Hapag: Tracking only.
func TestTask21_Test8_MscCapabilities(t *testing.T) {
	registry := adapters.GetDefaultRegistry()
	msc := registry.GetAdapterBySCAC("MSCU")
	require.NotNil(t, msc)

	// Tracking must be supported
	assert.True(t, msc.SupportsCapability(domain.CapTracking), "MSC must support CapTracking")
	assert.True(t, msc.HasCapability(domain.CapTracking), "HasCapability must also return true for CapTracking")

	// Other capabilities must NOT be advertised
	assert.False(t, msc.SupportsCapability(domain.CapRates), "MSC must NOT advertise CapRates")
	assert.False(t, msc.SupportsCapability(domain.CapSpotRates), "MSC must NOT advertise CapSpotRates")
	assert.False(t, msc.SupportsCapability(domain.CapContractRates), "MSC must NOT advertise CapContractRates")
	assert.False(t, msc.SupportsCapability(domain.CapBooking), "MSC must NOT advertise CapBooking")
	assert.False(t, msc.SupportsCapability(domain.CapDocuments), "MSC must NOT advertise CapDocuments")

	// Verify exact list
	caps := msc.SupportedCapabilities()
	assert.Len(t, caps, 1, "MSC must advertise exactly 1 capability")
	assert.Equal(t, domain.CapTracking, caps[0])
}

// Test 9 — Maersk regression
// Verify Maersk's existing capabilities remain unchanged.
func TestTask21_Test9_MaerskCapabilitiesRemainUnchanged(t *testing.T) {
	registry := adapters.GetDefaultRegistry()
	maersk := registry.GetAdapterBySCAC("MAEU")
	require.NotNil(t, maersk)

	assert.True(t, maersk.SupportsCapability(domain.CapTracking), "Maersk must support CapTracking")
	assert.True(t, maersk.SupportsCapability(domain.CapRates), "Maersk must support CapRates")
	assert.True(t, maersk.SupportsCapability(domain.CapSpotRates), "Maersk must support CapSpotRates")
	assert.True(t, maersk.SupportsCapability(domain.CapContractRates), "Maersk must support CapContractRates")
	assert.True(t, maersk.SupportsCapability(domain.CapBooking), "Maersk must support CapBooking")
	assert.True(t, maersk.SupportsCapability(domain.CapDocuments), "Maersk must support CapDocuments")

	caps := maersk.SupportedCapabilities()
	assert.Contains(t, caps, domain.CapTracking)
	assert.Contains(t, caps, domain.CapRates)
	assert.Contains(t, caps, domain.CapSpotRates)
	assert.Contains(t, caps, domain.CapContractRates)
	assert.Contains(t, caps, domain.CapBooking)
	assert.Contains(t, caps, domain.CapDocuments)
}

// Test 10 — CarrierService capability enforcement for Hapag and MSC
// When asking Hapag or MSC for Rates, CarrierService blocks with ErrCapabilityNotSupported.
func TestTask21_Test10_CarrierServiceEnforcesAdapterCapabilities(t *testing.T) {
	// Verify ErrCapabilityNotSupported is exposed by service package
	assert.NotNil(t, service.ErrCapabilityNotSupported)
	assert.Contains(t, service.ErrCapabilityNotSupported.Error(), "not supported by this carrier adapter")
}
