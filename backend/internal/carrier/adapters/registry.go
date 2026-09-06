package adapters

import (
	"context"
	"strings"
	"sync"

	"github.com/freel/backend/internal/carrier"
)

// AdapterRegistry maintains the mapping between carrier provider codes/SCACs and their concrete adapters.
type AdapterRegistry struct {
	mu           sync.RWMutex
	byCode       map[string]CarrierAdapter
	bySCAC       map[string]CarrierAdapter
}

var (
	defaultRegistry *AdapterRegistry
	once            sync.Once
)

// GetDefaultRegistry returns the singleton registry with pre-registered shipping line adapters.
func GetDefaultRegistry() *AdapterRegistry {
	once.Do(func() {
		defaultRegistry = NewAdapterRegistry()
		defaultRegistry.Register(NewMaerskAdapter())
		defaultRegistry.Register(NewMscAdapter())
		defaultRegistry.Register(NewHapagLloydAdapter())
		defaultRegistry.Register(NewCmaCgmAdapter())
		defaultRegistry.Register(NewOneAdapter())
		defaultRegistry.Register(NewEvergreenAdapter())
		defaultRegistry.Register(NewCoscoAdapter())
	})
	return defaultRegistry
}

// NewAdapterRegistry constructs an empty AdapterRegistry.
func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{
		byCode: make(map[string]CarrierAdapter),
		bySCAC: make(map[string]CarrierAdapter),
	}
}

// Register adds an adapter to the registry under its provider code and SCAC.
func (r *AdapterRegistry) Register(adapter CarrierAdapter) {
	r.mu.Lock()
	defer r.mu.Unlock()

	code := strings.ToUpper(adapter.ProviderCode())
	scac := strings.ToUpper(adapter.SCAC())

	r.byCode[code] = adapter
	r.bySCAC[scac] = adapter

	// Map known alias SCACs
	switch code {
	case "MAERSK":
		r.bySCAC["MSK"] = adapter
		r.bySCAC["MAEU"] = adapter
	case "MSC":
		r.bySCAC["MSC"] = adapter
		r.bySCAC["MSCU"] = adapter
	case "HAPAG_LLOYD":
		r.bySCAC["HAPAG"] = adapter
		r.bySCAC["HLCU"] = adapter
	case "CMA_CGM":
		r.bySCAC["CMA"] = adapter
		r.bySCAC["CMDU"] = adapter
	case "ONE":
		r.bySCAC["ONE"] = adapter
		r.bySCAC["ONEY"] = adapter
	case "EVERGREEN":
		r.bySCAC["EGLV"] = adapter
		r.bySCAC["EVERGREEN"] = adapter
	case "COSCO":
		r.bySCAC["COSU"] = adapter
		r.bySCAC["COSCO"] = adapter
	}
}

// GetAdapter resolves an adapter by provider code (e.g. "MAERSK", "MSC").
func (r *AdapterRegistry) GetAdapter(code string) (CarrierAdapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapter, ok := r.byCode[strings.ToUpper(strings.TrimSpace(code))]
	return adapter, ok
}

// GetAdapterBySCAC resolves an adapter by carrier SCAC code (e.g. "MAEU", "MSCU", "HLCU").
// If no direct adapter is registered, it returns a configured GenericAdapter.
func (r *AdapterRegistry) GetAdapterBySCAC(scac string) CarrierAdapter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	scacUpper := strings.ToUpper(strings.TrimSpace(scac))
	if adapter, ok := r.bySCAC[scacUpper]; ok {
		return adapter
	}

	// Fallback to generic adapter for custom/unregistered SCACs
	return NewGenericAdapter(scacUpper, scacUpper)
}

// ─────────────────────────────────────────────────────────────────────────────
// Legacy Adapter Compatibility Resolvers
// ─────────────────────────────────────────────────────────────────────────────

func resolveConfiguredAdapter(db carrier.Queryer, orgID int64, scac string) (*MockTrackingAdapter, error) {
	scacUpper := strings.ToUpper(scac)
	cfg, err := carrier.GetIntegrationConfig(context.Background(), db, orgID, scac)
	if err != nil {
		return nil, err
	}
	return NewMockTrackingAdapter(scacUpper, cfg), nil
}

// GetTrackingProvider resolves a TrackingProvider by org ID and carrier SCAC.
func GetTrackingProvider(db carrier.Queryer, orgID int64, scac string) (carrier.TrackingProvider, error) {
	return resolveConfiguredAdapter(db, orgID, scac)
}

// GetBookingProvider resolves a BookingProvider by org ID and carrier SCAC.
func GetBookingProvider(db carrier.Queryer, orgID int64, scac string) (carrier.BookingProvider, error) {
	return resolveConfiguredAdapter(db, orgID, scac)
}

// GetWebhookProvider resolves a WebhookProvider by org ID and carrier SCAC.
func GetWebhookProvider(db carrier.Queryer, orgID int64, scac string) (carrier.WebhookProvider, error) {
	return resolveConfiguredAdapter(db, orgID, scac)
}

// GetRateProvider resolves a RateProvider by org ID and carrier SCAC.
func GetRateProvider(db carrier.Queryer, orgID int64, scac string) (carrier.RateProvider, error) {
	return resolveConfiguredAdapter(db, orgID, scac)
}
