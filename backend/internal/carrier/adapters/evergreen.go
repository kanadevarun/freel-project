package adapters

import (
	"github.com/freel/backend/internal/carrier/domain"
)

// EvergreenAdapter implements the CarrierAdapter interface for Evergreen Marine Corporation (SCAC: EGLV).
type EvergreenAdapter struct {
	BaseAdapter
}

func NewEvergreenAdapter() *EvergreenAdapter {
	return &EvergreenAdapter{
		BaseAdapter: BaseAdapter{
			Code:        "EVERGREEN",
			CarrierSCAC: "EGLV",
			Capabilities: []domain.Capability{
				domain.CapTracking,
				domain.CapRates,
				domain.CapBooking,
			},
		},
	}
}
