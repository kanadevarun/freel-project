package rates

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	carrierDomain "github.com/freel/backend/internal/carrier/domain"
	carrierService "github.com/freel/backend/internal/carrier/service"
	"github.com/freel/backend/internal/rates/spec"
)

// CarrierRatesEngine handles live multi-carrier rate querying and normalization.
type CarrierRatesEngine struct {
	carrierSvc carrierService.CarrierService
}

// NewCarrierRatesEngine creates a new carrier rates engine.
func NewCarrierRatesEngine(carrierSvc carrierService.CarrierService) *CarrierRatesEngine {
	return &CarrierRatesEngine{
		carrierSvc: carrierSvc,
	}
}

// SearchLiveRates queries all eligible connected carrier integrations for an organization.
func (e *CarrierRatesEngine) SearchLiveRates(ctx context.Context, orgID int64, req spec.CarrierRateSearchRequest) (*spec.CarrierRateSearchResponse, error) {
	now := time.Now().UTC()

	resp := &spec.CarrierRateSearchResponse{
		Success:         true,
		OriginPort:      req.OriginPort,
		DestinationPort: req.DestinationPort,
		EquipmentType:   req.EquipmentType,
		Rates:           []spec.CarrierRateComparisonItem{},
		TotalRatesCount: 0,
		CarriersQueried: 0,
		SearchedAt:      now,
	}

	if e.carrierSvc == nil {
		resp.Message = "Carrier integration service is unavailable."
		return resp, nil
	}

	// 1. Validate mandatory route parameters
	if strings.TrimSpace(req.OriginPort) == "" || strings.TrimSpace(req.DestinationPort) == "" {
		resp.Message = "Origin and destination ports are required to fetch carrier rates."
		return resp, nil
	}

	if strings.EqualFold(req.OriginPort, req.DestinationPort) {
		resp.Message = "Origin and destination ports cannot be identical."
		return resp, nil
	}

	eqType := strings.ToUpper(strings.TrimSpace(req.EquipmentType))
	if eqType == "" {
		eqType = "40HC"
	}
	resp.EquipmentType = eqType

	// 2. Fetch all active integrations for this tenant
	integrations, err := e.carrierSvc.GetIntegrations(ctx, orgID)
	if err != nil {
		log.Printf("[CarrierRatesEngine] Error fetching integrations for org %d: %v", orgID, err)
		resp.Message = "Unable to retrieve carrier integrations for your organization."
		return resp, nil
	}

	if len(integrations) == 0 {
		resp.Message = "No carrier integrations configured. Connect a carrier in Settings > Carrier Integrations to fetch live rates."
		return resp, nil
	}

	// 3. Filter integrations by requested carrier SCAC (if specified) and capability
	var eligibleIntegrations []carrierDomain.CarrierIntegrationView
	for _, ci := range integrations {
		if !ci.IsEnabled || ci.ConnectionStatus == carrierDomain.StatusDisabled {
			continue
		}

		if req.CarrierSCAC != "" && !strings.EqualFold(ci.CarrierSCAC, req.CarrierSCAC) && !strings.EqualFold(ci.CarrierName, req.CarrierSCAC) {
			continue
		}

		hasRateCap := false
		for _, cap := range ci.Capabilities {
			if cap == carrierDomain.CapRates || cap == carrierDomain.CapSpotRates || cap == carrierDomain.CapContractRates {
				hasRateCap = true
				break
			}
		}

		if hasRateCap {
			eligibleIntegrations = append(eligibleIntegrations, ci)
		}
	}

	if len(eligibleIntegrations) == 0 {
		if req.CarrierSCAC != "" {
			resp.Message = fmt.Sprintf("Carrier integration (%s) is not configured or does not have Rates capability enabled.", req.CarrierSCAC)
		} else {
			resp.Message = "None of your connected carriers have the Rates capability enabled. Enable Rates in Settings > Carrier Integrations."
		}
		return resp, nil
	}

	// 4. Build standard carrier RateRequest
	validDate := now
	if req.CargoReadyDate != "" {
		if parsedDate, err := time.Parse("2006-01-02", req.CargoReadyDate); err == nil {
			validDate = parsedDate
		}
	}

	carrierReq := carrierDomain.RateRequest{
		OriginPort:      strings.ToUpper(strings.TrimSpace(req.OriginPort)),
		DestinationPort: strings.ToUpper(strings.TrimSpace(req.DestinationPort)),
		EquipmentType:   eqType,
		Commodity:       req.Commodity,
		ValidDate:       validDate,
	}

	// 5. Query each eligible carrier adapter
	allRates := make([]spec.CarrierRateComparisonItem, 0)
	var cheapestRate *spec.CarrierRateComparisonItem
	var fastestRate *spec.CarrierRateComparisonItem

	for _, ci := range eligibleIntegrations {
		resp.CarriersQueried++

		rates, err := e.carrierSvc.GetRates(ctx, orgID, ci.ID, carrierReq)
		if err != nil {
			log.Printf("[CarrierRatesEngine] Failed to fetch rates from %s (%s): %v", ci.CarrierName, ci.CarrierSCAC, err)
			continue
		}

		for _, r := range rates {
			// Respect RateType filter
			if req.RateType == "SPOT" && r.IsContractRate {
				continue
			}
			if req.RateType == "CONTRACT" && !r.IsContractRate {
				continue
			}

			item := spec.CarrierRateComparisonItem{
				RateID:             r.RateID,
				CarrierSCAC:        r.CarrierSCAC,
				CarrierName:        ci.CarrierName,
				OriginPort:         r.OriginPort,
				DestinationPort:    r.DestinationPort,
				EquipmentType:      r.EquipmentType,
				ServiceCode:        r.ServiceCode,
				VesselName:         r.VesselName,
				Currency:           r.Currency,
				OceanFreight:       r.OceanFreight,
				OriginCharges:      r.OriginCharges,
				DestinationCharges: r.DestinationCharges,
				TotalBuyPrice:      r.TotalBuyPrice,
				TransitDays:        r.TransitDays,
				FreeDays:           r.FreeDays,
				ValidFrom:          r.ValidFrom,
				ValidUntil:         r.ValidUntil,
				IsContractRate:     r.IsContractRate,
			}

			allRates = append(allRates, item)
		}
	}

	if len(allRates) == 0 {
		resp.Message = fmt.Sprintf("No carrier rates currently available for route %s -> %s with equipment %s.", req.OriginPort, req.DestinationPort, eqType)
		return resp, nil
	}

	// 6. Compute comparison analytics (Cheapest, Fastest, Best Value)
	for i := range allRates {
		r := &allRates[i]
		if cheapestRate == nil || r.TotalBuyPrice < cheapestRate.TotalBuyPrice {
			cheapestRate = r
		}
		if r.TransitDays > 0 {
			if fastestRate == nil || r.TransitDays < fastestRate.TransitDays {
				fastestRate = r
			}
		}
	}

	for i := range allRates {
		r := &allRates[i]
		if cheapestRate != nil && r.RateID == cheapestRate.RateID {
			r.IsCheapest = true
		}
		if fastestRate != nil && r.RateID == fastestRate.RateID {
			r.IsFastest = true
		}
		if r.IsCheapest && r.IsFastest {
			r.IsBestValue = true
		}
	}

	resp.Rates = allRates
	resp.TotalRatesCount = len(allRates)
	if cheapestRate != nil {
		resp.CheapestCarrier = &cheapestRate.CarrierName
		resp.CheapestAmount = &cheapestRate.TotalBuyPrice
	}
	if fastestRate != nil {
		resp.FastestCarrier = &fastestRate.CarrierName
		resp.FastestTransit = &fastestRate.TransitDays
	}
	resp.Message = fmt.Sprintf("Successfully retrieved %d live rate options from connected carriers.", len(allRates))

	return resp, nil
}
