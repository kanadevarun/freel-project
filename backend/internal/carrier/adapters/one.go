package adapters

import (
	"github.com/freel/backend/internal/carrier/domain"
)

// OneAdapter implements the CarrierAdapter interface for Ocean Network Express (SCAC: ONEY).
type OneAdapter struct {
	BaseAdapter
}

func NewOneAdapter() *OneAdapter {
	return &OneAdapter{
		BaseAdapter: BaseAdapter{
			Code:        "ONE",
			CarrierSCAC: "ONEY",
			Capabilities: []domain.Capability{
				domain.CapTracking,
				domain.CapRates,
				domain.CapBooking,
			},
		},
	}
}
