package adapters

import (
	"github.com/freel/backend/internal/carrier/domain"
)

// CoscoAdapter implements the CarrierAdapter interface for COSCO Shipping Lines (SCAC: COSU).
type CoscoAdapter struct {
	BaseAdapter
}

func NewCoscoAdapter() *CoscoAdapter {
	return &CoscoAdapter{
		BaseAdapter: BaseAdapter{
			Code:        "COSCO",
			CarrierSCAC: "COSU",
			Capabilities: []domain.Capability{
				domain.CapTracking,
				domain.CapRates,
				domain.CapDocuments,
			},
		},
	}
}
