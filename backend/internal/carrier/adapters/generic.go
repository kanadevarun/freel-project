package adapters

import (
	"strings"

	"github.com/freel/backend/internal/carrier/domain"
)

// GenericAdapter provides fallback execution for carriers that follow standard ocean EDI/API patterns.
type GenericAdapter struct {
	BaseAdapter
}

func NewGenericAdapter(code, scac string) *GenericAdapter {
	if code == "" {
		code = "GENERIC"
	}
	if scac == "" {
		scac = "GENERIC"
	}
	return &GenericAdapter{
		BaseAdapter: BaseAdapter{
			Code:        strings.ToUpper(code),
			CarrierSCAC: strings.ToUpper(scac),
			Capabilities: []domain.Capability{
				domain.CapTracking,
				domain.CapRates,
				domain.CapBooking,
				domain.CapDocuments,
			},
		},
	}
}
