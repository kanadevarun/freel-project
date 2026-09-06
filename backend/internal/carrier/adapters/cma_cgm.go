package adapters

import (
	"github.com/freel/backend/internal/carrier/domain"
)

// CmaCgmAdapter implements the CarrierAdapter interface for CMA CGM Group (SCAC: CMDU / CMA).
type CmaCgmAdapter struct {
	BaseAdapter
}

func NewCmaCgmAdapter() *CmaCgmAdapter {
	return &CmaCgmAdapter{
		BaseAdapter: BaseAdapter{
			Code:        "CMA_CGM",
			CarrierSCAC: "CMDU",
			Capabilities: []domain.Capability{
				domain.CapTracking,
				domain.CapRates,
				domain.CapContractRates,
				domain.CapBooking,
				domain.CapDocuments,
			},
		},
	}
}
