package rates

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	carrier "github.com/freel/backend/internal/carrier"
	carrierService "github.com/freel/backend/internal/carrier/service"
	audit "github.com/freel/backend/internal/audit"
	auditDomain "github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/rates/spec"
)

// BusinessLogic interface for Rate Management & Intelligence
type BusinessLogic interface {
	// Intelligence & search
	SearchRates(ctx context.Context, q spec.RateQuery) (*spec.RateSearchResult, error)
	IngestRates(ctx context.Context, rates []spec.CanonicalRate) error
	RefreshSpotRates(ctx context.Context, orgID int64, q spec.RateQuery, targetDate *time.Time) (*spec.RateSearchResult, error)
	GetRateByID(ctx context.Context, orgID int64, id string) (*spec.CanonicalRate, error)

	// Live Carrier Integration Rates (Task 5)
	SearchCarrierLiveRates(ctx context.Context, orgID int64, req spec.CarrierRateSearchRequest) (*spec.CarrierRateSearchResponse, error)
	SetCarrierIntegrationService(carrierSvc carrierService.CarrierService)

	// Rate Management (Task 19.1)
	CreateRate(ctx context.Context, req spec.CreateRateRequest) (*spec.Rate, error)
	GetRate(ctx context.Context, orgID, rateID int64) (*spec.RateDetail, error)
	ListRates(ctx context.Context, filters spec.ListRatesRequest) (*spec.ListRatesResponse, error)
	UpdateRate(ctx context.Context, req spec.UpdateRateRequest) (*spec.Rate, error)
	ArchiveRate(ctx context.Context, req spec.ArchiveRateRequest) error
	GetRateSummary(ctx context.Context, orgID int64) (*spec.RateSummaryKPIs, error)

	// Rate Charges & Pricing (Task 19.2)
	GetRatePricing(ctx context.Context, orgID, rateID int64) (*spec.RatePricingSummary, error)
	AddRateCharge(ctx context.Context, req spec.CreateRateChargeRequest) (*spec.RateChargeItem, *spec.RatePricingSummary, error)
	UpdateRateCharge(ctx context.Context, req spec.UpdateRateChargeRequest) (*spec.RateChargeItem, *spec.RatePricingSummary, error)
	DeleteRateCharge(ctx context.Context, req spec.DeleteRateChargeRequest) (*spec.RatePricingSummary, error)
	ReorderRateCharges(ctx context.Context, req spec.ReorderRateChargesRequest) (*spec.RatePricingSummary, error)

	// Carrier Rate Contracts & Versions (Task 19.3)
	CreateRateContract(ctx context.Context, req spec.CreateRateContractRequest) (*spec.RateContract, error)
	GetRateContract(ctx context.Context, orgID, contractID int64) (*spec.RateContract, error)
	ListRateContracts(ctx context.Context, filters spec.ListRateContractsRequest) (*spec.ListRateContractsResponse, error)
	UpdateRateContract(ctx context.Context, req spec.UpdateRateContractRequest) (*spec.RateContract, error)
	ArchiveRateContract(ctx context.Context, orgID, contractID int64, archivedBy string) error
	RenewRateContract(ctx context.Context, req spec.RenewRateContractRequest) (*spec.RateContract, error)
	GetRateContractSummary(ctx context.Context, orgID int64) (*spec.RateContractSummary, error)
	GetRatesByContract(ctx context.Context, orgID, contractID int64) ([]*spec.RateListItem, error)

	// Rate Versioning & Downstream Protection
	CreateNewRateVersion(ctx context.Context, req spec.CreateRateVersionRequest) (*spec.Rate, error)
	GetRateVersionHistory(ctx context.Context, orgID, rateID int64) ([]spec.RateVersionHistory, error)
	GetRateVersionChain(ctx context.Context, orgID, rateID int64) ([]spec.RateVersionChainItem, error)

	// Task 19.4: Spot Rate Requests, Responses & Comparison Intelligence
	CreateSpotRateRequest(ctx context.Context, req spec.CreateSpotRateRequestRequest) (*spec.SpotRateRequest, error)
	GetSpotRateRequest(ctx context.Context, orgID, requestID int64) (*spec.SpotRateRequest, error)
	ListSpotRateRequests(ctx context.Context, filters spec.ListSpotRateRequestsRequest) (*spec.ListSpotRateRequestsResponse, error)
	UpdateSpotRateRequest(ctx context.Context, req spec.UpdateSpotRateRequestRequest) (*spec.SpotRateRequest, error)
	SendSpotRateRequest(ctx context.Context, orgID, requestID int64, user string) (*spec.SpotRateRequest, error)
	CancelSpotRateRequest(ctx context.Context, orgID, requestID int64, user string) error

	CreateSpotRateResponse(ctx context.Context, req spec.CreateSpotRateResponseRequest) (*spec.SpotRateResponse, error)
	GetSpotRateResponse(ctx context.Context, orgID, responseID int64) (*spec.SpotRateResponse, error)
	GetSpotRateResponses(ctx context.Context, orgID, requestID int64) ([]*spec.SpotRateResponse, error)
	UpdateSpotRateResponse(ctx context.Context, req spec.UpdateSpotRateResponseRequest) (*spec.SpotRateResponse, error)
	SelectPreferredSpotRate(ctx context.Context, req spec.SelectPreferredSpotRateRequest) (*spec.SpotRateResponse, error)

	CompareSpotRates(ctx context.Context, orgID, requestID int64) (*spec.SpotRateComparison, error)
	GetSpotRateSummary(ctx context.Context, orgID int64) (*spec.SpotRateRequestSummary, error)

	// Task 19.6: Rate Lifecycle Intelligence & Attention
	GetRateLifecycleDashboard(ctx context.Context, orgID int64) (*RateLifecycleSummary, error)
	GetRateLifecycleEvents(ctx context.Context, orgID int64, limit int) ([]*RateLifecycleEvent, error)
	GetRatesRequiringAttention(ctx context.Context, orgID int64) ([]*RateAttentionItem, error)
	GetContractsRequiringAttention(ctx context.Context, orgID int64) ([]*ContractAttentionItem, error)
	EvaluateRateLifecycleForOrg(ctx context.Context, orgID int64) (*RateLifecycleSummary, error)

	// Task 19.7: Rate Analytics & Procurement Intelligence
	GetRateAnalyticsOverview(ctx context.Context, orgID int64) (*RateAnalyticsOverview, error)
	GetRateAnalyticsTrends(ctx context.Context, orgID int64, days int) ([]RateTrendDataPoint, error)
	GetCarrierRatePerformance(ctx context.Context, orgID int64) ([]CarrierRatePerformance, error)
	GetLaneRatePerformance(ctx context.Context, orgID int64) ([]LaneRatePerformance, error)
	GetRateLifecycleAnalytics(ctx context.Context, orgID int64) (*RateLifecycleAnalytics, error)
	GetSpotSourcingPerformance(ctx context.Context, orgID int64) (*SpotSourcingPerformance, error)
	GetRateCommercialInsights(ctx context.Context, orgID int64) ([]CommercialImpactInsight, error)
}

type businessLogic struct {
	dl                 Datalayer
	normalizer         SpotNormalizer
	carrier            carrier.Service
	carrierRatesEngine *CarrierRatesEngine
}

func NewBusinessLogic(dl Datalayer, normalizer SpotNormalizer, carrierSvc carrier.Service) BusinessLogic {
	return &businessLogic{
		dl:         dl,
		normalizer: normalizer,
		carrier:    carrierSvc,
	}
}

func (b *businessLogic) SetCarrierIntegrationService(carrierSvc carrierService.CarrierService) {
	b.carrierRatesEngine = NewCarrierRatesEngine(carrierSvc)
}

func (b *businessLogic) SearchCarrierLiveRates(ctx context.Context, orgID int64, req spec.CarrierRateSearchRequest) (*spec.CarrierRateSearchResponse, error) {
	if b.carrierRatesEngine == nil {
		return &spec.CarrierRateSearchResponse{
			Success:         true,
			OriginPort:      req.OriginPort,
			DestinationPort: req.DestinationPort,
			EquipmentType:   req.EquipmentType,
			Rates:           []spec.CarrierRateComparisonItem{},
			Message:         "Carrier integration engine is not initialized. Connect a carrier in Settings > Carrier Integrations to fetch live rates.",
			SearchedAt:      time.Now().UTC(),
		}, nil
	}
	return b.carrierRatesEngine.SearchLiveRates(ctx, orgID, req)
}

func (b *businessLogic) IngestRates(ctx context.Context, rates []spec.CanonicalRate) error {
	return b.dl.Upsert(ctx, rates)
}

func (b *businessLogic) GetRateByID(ctx context.Context, orgID int64, id string) (*spec.CanonicalRate, error) {
	return b.dl.GetByID(ctx, orgID, id)
}

func (b *businessLogic) SearchRates(ctx context.Context, q spec.RateQuery) (*spec.RateSearchResult, error) {
	rates, err := b.dl.Search(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("rates.Search: %w", err)
	}

	if len(rates) == 0 {
		return &spec.RateSearchResult{
			Rates:             []spec.CanonicalRate{},
			TotalCount:        0,
			SpotRateCount:     0,
			ContractRateCount: 0,
			RecommendedIdx:    -1,
			OverallReasoning:  "No rates found matching the query criteria.",
			SearchedAt:        time.Now(),
		}, nil
	}

	return rankAndFormatRates(rates), nil
}

func (b *businessLogic) RefreshSpotRates(ctx context.Context, orgID int64, q spec.RateQuery, targetDate *time.Time) (*spec.RateSearchResult, error) {
	if b.carrier == nil {
		return nil, fmt.Errorf("rates.RefreshSpotRates: carrier service is nil")
	}

	fetchResp, err := b.carrier.FetchRates(ctx, q.OriginPort, q.DestinationPort, targetDate, q.Incoterms, 0, 0, "")
	if err != nil {
		return nil, fmt.Errorf("rates.RefreshSpotRates carrier.FetchRates: %w", err)
	}

	canonicalList := make([]spec.CanonicalRate, 0, len(fetchResp.Rates))
	for _, raw := range fetchResp.Rates {
		c := b.normalizer.Normalize(raw, orgID, q.OriginPort, q.DestinationPort)
		canonicalList = append(canonicalList, c)
	}

	if err := b.dl.Upsert(ctx, canonicalList); err != nil {
		return nil, fmt.Errorf("rates.RefreshSpotRates dl.Upsert: %w", err)
	}

	return b.SearchRates(ctx, q)
}

func (b *businessLogic) CreateRate(ctx context.Context, req spec.CreateRateRequest) (*spec.Rate, error) {
	if req.CarrierName == "" {
		return nil, fmt.Errorf("carrier_name is required")
	}
	if req.OriginPort == "" {
		return nil, fmt.Errorf("origin_port is required")
	}
	if req.DestinationPort == "" {
		return nil, fmt.Errorf("destination_port is required")
	}
	if req.OriginPort == req.DestinationPort {
		return nil, fmt.Errorf("origin_port and destination_port cannot be identical")
	}
	if req.BaseAmount < 0 {
		return nil, fmt.Errorf("base_amount cannot be negative")
	}

	effDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effective_date format (YYYY-MM-DD expected)")
	}
	expDate, err := time.Parse("2006-01-02", req.ExpiryDate)
	if err != nil {
		return nil, fmt.Errorf("invalid expiry_date format (YYYY-MM-DD expected)")
	}
	if expDate.Before(effDate) {
		return nil, fmt.Errorf("expiry_date cannot be earlier than effective_date")
	}

	rateRef := req.RateReference
	if rateRef == "" {
		rateRef = fmt.Sprintf("RAT-%d-%04d", time.Now().Year(), time.Now().UnixNano()%10000)
	}

	now := time.Now().UTC()
	status, _ := CalculateRateValidity(effDate, expDate, RateStatusActive, now)

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}
	rateType := req.RateType
	if rateType == "" {
		rateType = "SPOT"
	}
	mode := req.TransportMode
	if mode == "" {
		mode = "Ocean FCL"
	}
	serviceType := req.ServiceType
	if serviceType == "" {
		serviceType = "FCL"
	}

	author := req.Author
	rate := &spec.Rate{
		OrgID:             req.OrgID,
		RateReference:     rateRef,
		CarrierName:       req.CarrierName,
		CarrierCode:       req.CarrierCode,
		ServiceProvider:   req.ServiceProvider,
		RateType:          rateType,
		TransportMode:     mode,
		ServiceType:       serviceType,
		EquipmentType:     req.EquipmentType,
		OriginPort:        req.OriginPort,
		OriginCode:        req.OriginCode,
		DestinationPort:   req.DestinationPort,
		DestinationCode:   req.DestinationCode,
		Currency:          currency,
		BaseAmount:        req.BaseAmount,
		EffectiveDate:     effDate,
		ExpiryDate:        expDate,
		Status:            status,
		CarrierReference:  req.CarrierReference,
		ContractReference: req.ContractReference,
		Notes:             req.Notes,
		CreatedBy:         &author,
		UpdatedBy:         &author,
	}

	if err := b.dl.CreateRate(ctx, rate); err != nil {
		return nil, err
	}

	return rate, nil
}

func (b *businessLogic) GetRate(ctx context.Context, orgID, rateID int64) (*spec.RateDetail, error) {
	rate, err := b.dl.GetRateByID(ctx, orgID, rateID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	status, days := CalculateRateValidity(rate.EffectiveDate, rate.ExpiryDate, rate.Status, now)
	rate.Status = status

	tags := []string{rate.RateType, rate.TransportMode}
	if rate.EquipmentType != nil && *rate.EquipmentType != "" {
		tags = append(tags, *rate.EquipmentType)
	}
	if status == RateStatusExpiringSoon {
		tags = append(tags, "Expiring Soon")
	}

	detail := &spec.RateDetail{
		Rate:            *rate,
		DaysUntilExpiry: days,
		IsExpired:       status == RateStatusExpired,
		IsExpiringSoon:  status == RateStatusExpiringSoon,
		LaneDisplay:     fmt.Sprintf("%s → %s", rate.OriginPort, rate.DestinationPort),
		ValidityText:    FormatValidityText(status, days, rate.ExpiryDate),
		Tags:            tags,
	}

	return detail, nil
}

func (b *businessLogic) ListRates(ctx context.Context, filters spec.ListRatesRequest) (*spec.ListRatesResponse, error) {
	rows, total, err := b.dl.ListRates(ctx, &filters)
	if err != nil {
		return nil, err
	}

	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	page := filters.Page
	if page <= 0 {
		page = 1
	}
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	return &spec.ListRatesResponse{
		Rates:      rows,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (b *businessLogic) UpdateRate(ctx context.Context, req spec.UpdateRateRequest) (*spec.Rate, error) {
	rate, err := b.dl.GetRateByID(ctx, req.OrgID, req.ID)
	if err != nil {
		return nil, err
	}

	// Task 19.3 Commercial Protection: If rate is referenced downstream, create a new rate version instead of mutating
	isReferenced, _ := b.dl.IsRateReferencedDownstream(ctx, req.OrgID, req.ID)
	if isReferenced && (req.BaseAmount != nil || req.Currency != nil || req.EffectiveDate != nil || req.ExpiryDate != nil) {
		// Create new revision version
		reason := "Revision created from rate edit (rate referenced in downstream records)"
		versionReq := spec.CreateRateVersionRequest{
			OrgID:             req.OrgID,
			RateID:            req.ID,
			User:              req.Updater,
			BaseAmount:        req.BaseAmount,
			Currency:          req.Currency,
			EffectiveDate:     req.EffectiveDate,
			ExpiryDate:        req.ExpiryDate,
			CarrierReference:  req.CarrierReference,
			ContractReference: req.ContractReference,
			Notes:             req.Notes,
			RevisionReason:    reason,
		}
		return b.CreateNewRateVersion(ctx, versionReq)
	}

	if req.CarrierName != nil && *req.CarrierName != "" {
		rate.CarrierName = *req.CarrierName
	}
	if req.CarrierCode != nil {
		rate.CarrierCode = req.CarrierCode
	}
	if req.ServiceProvider != nil {
		rate.ServiceProvider = req.ServiceProvider
	}
	if req.RateType != nil && *req.RateType != "" {
		rate.RateType = *req.RateType
	}
	if req.TransportMode != nil && *req.TransportMode != "" {
		rate.TransportMode = *req.TransportMode
	}
	if req.ServiceType != nil && *req.ServiceType != "" {
		rate.ServiceType = *req.ServiceType
	}
	if req.EquipmentType != nil {
		rate.EquipmentType = req.EquipmentType
	}
	if req.OriginPort != nil && *req.OriginPort != "" {
		rate.OriginPort = *req.OriginPort
	}
	if req.OriginCode != nil {
		rate.OriginCode = req.OriginCode
	}
	if req.DestinationPort != nil && *req.DestinationPort != "" {
		rate.DestinationPort = *req.DestinationPort
	}
	if req.DestinationCode != nil {
		rate.DestinationCode = req.DestinationCode
	}
	if req.Currency != nil && *req.Currency != "" {
		rate.Currency = *req.Currency
	}
	if req.BaseAmount != nil {
		if *req.BaseAmount < 0 {
			return nil, fmt.Errorf("base_amount cannot be negative")
		}
		rate.BaseAmount = *req.BaseAmount
	}
	if req.EffectiveDate != nil {
		eff, err := time.Parse("2006-01-02", *req.EffectiveDate)
		if err != nil {
			return nil, fmt.Errorf("invalid effective_date format (YYYY-MM-DD expected)")
		}
		rate.EffectiveDate = eff
	}
	if req.ExpiryDate != nil {
		exp, err := time.Parse("2006-01-02", *req.ExpiryDate)
		if err != nil {
			return nil, fmt.Errorf("invalid expiry_date format (YYYY-MM-DD expected)")
		}
		rate.ExpiryDate = exp
	}
	if rate.ExpiryDate.Before(rate.EffectiveDate) {
		return nil, fmt.Errorf("expiry_date cannot be earlier than effective_date")
	}

	if req.Status != nil && *req.Status != "" {
		rate.Status = *req.Status
	} else {
		now := time.Now().UTC()
		st, _ := CalculateRateValidity(rate.EffectiveDate, rate.ExpiryDate, rate.Status, now)
		rate.Status = st
	}

	if req.CarrierReference != nil {
		rate.CarrierReference = req.CarrierReference
	}
	if req.ContractReference != nil {
		rate.ContractReference = req.ContractReference
	}
	if req.Notes != nil {
		rate.Notes = req.Notes
	}

	rate.UpdatedBy = &req.Updater

	if err := b.dl.UpdateRate(ctx, rate); err != nil {
		return nil, err
	}

	return rate, nil
}

func (b *businessLogic) ArchiveRate(ctx context.Context, req spec.ArchiveRateRequest) error {
	return b.dl.ArchiveRate(ctx, req.OrgID, req.ID, req.User)
}

func (b *businessLogic) GetRateSummary(ctx context.Context, orgID int64) (*spec.RateSummaryKPIs, error) {
	return b.dl.GetRateSummaryKPIs(ctx, orgID)
}

func rankAndFormatRates(rates []spec.CanonicalRate) *spec.RateSearchResult {
	sort.SliceStable(rates, func(i, j int) bool {
		ri, rj := rates[i], rates[j]
		if ri.Source == RateSourceContractPDF && rj.Source == RateSourceSpotAPI {
			if ri.ConfidenceScore >= 85 {
				return true
			}
		}
		if ri.Source == RateSourceSpotAPI && rj.Source == RateSourceContractPDF {
			if rj.ConfidenceScore >= 85 {
				return false
			}
		}
		return ri.TotalBuyPrice < rj.TotalBuyPrice
	})

	var spotCount, contractCount int
	for _, r := range rates {
		if r.Source == RateSourceSpotAPI {
			spotCount++
		} else if r.Source == RateSourceContractPDF {
			contractCount++
		}
	}

	top := rates[0]
	reasoning := buildOverallReasoning(top)

	return &spec.RateSearchResult{
		Rates:             rates,
		TotalCount:        len(rates),
		SpotRateCount:     spotCount,
		ContractRateCount: contractCount,
		RecommendedIdx:    0,
		OverallReasoning:  reasoning,
		SearchedAt:        time.Now(),
	}
}

func buildOverallReasoning(top spec.CanonicalRate) string {
	sourceLabel := "spot rate"
	if top.Source == RateSourceContractPDF {
		sourceLabel = "contract rate"
	}
	transitStr := ""
	if top.TransitDays != nil {
		transitStr = fmt.Sprintf(", %d days transit", *top.TransitDays)
	}
	return fmt.Sprintf(
		"%s is recommended (%s, confidence %d/100%s). "+
			"Buy price: USD %.2f/container. Free days at destination: %d.",
		top.CarrierName, sourceLabel, top.ConfidenceScore, transitStr,
		top.TotalBuyPrice, top.FreeDaysDestination,
	)
}

// ── Task 19.2: Rate Charges Business Logic Implementation ─────────────────────

func (b *businessLogic) GetRatePricing(ctx context.Context, orgID, rateID int64) (*spec.RatePricingSummary, error) {
	rate, err := b.dl.GetRateByID(ctx, orgID, rateID)
	if err != nil {
		return nil, fmt.Errorf("rates.GetRatePricing get rate: %w", err)
	}

	charges, err := b.dl.GetRateCharges(ctx, orgID, rateID)
	if err != nil {
		return nil, fmt.Errorf("rates.GetRatePricing get charges: %w", err)
	}

	summary := CalculateRatePricing(rate.ID, rate.RateReference, rate.BaseAmount, rate.Currency, charges)
	return &summary, nil
}

func (b *businessLogic) AddRateCharge(ctx context.Context, req spec.CreateRateChargeRequest) (*spec.RateChargeItem, *spec.RatePricingSummary, error) {
	rate, err := b.dl.GetRateByID(ctx, req.OrgID, req.RateID)
	if err != nil {
		return nil, nil, fmt.Errorf("rates.AddRateCharge verify rate: %w", err)
	}

	// Rate editability rule: Archived rates reject mutations
	if rate.Status == RateStatusArchived {
		return nil, nil, fmt.Errorf("cannot add charge to an archived rate record")
	}

	if req.ChargeName == "" {
		return nil, nil, fmt.Errorf("charge name is required")
	}
	if req.CalculationBasis == "" {
		req.CalculationBasis = CalculationBasisFlat
	}
	if req.ChargeCategory == "" {
		req.ChargeCategory = ChargeCategoryFreight
	}
	if req.Currency == "" {
		req.Currency = rate.Currency
	}
	if req.Quantity <= 0 {
		req.Quantity = 1.0
	}
	if req.UnitPrice < 0 {
		return nil, nil, fmt.Errorf("unit price cannot be negative")
	}

	charge := &spec.RateChargeItem{
		OrgID:              req.OrgID,
		RateID:             req.RateID,
		ChargeCategory:     req.ChargeCategory,
		ChargeCode:         req.ChargeCode,
		ChargeName:         req.ChargeName,
		CalculationBasis:   req.CalculationBasis,
		Quantity:           req.Quantity,
		UnitPrice:          req.UnitPrice,
		Currency:           req.Currency,
		MinimumAmount:      req.MinimumAmount,
		MaximumAmount:      req.MaximumAmount,
		IncludedInBaseRate: req.IncludedInBaseRate,
		DisplayOrder:       req.DisplayOrder,
		Notes:              req.Notes,
	}

	if err := b.dl.CreateRateCharge(ctx, charge); err != nil {
		return nil, nil, fmt.Errorf("rates.AddRateCharge create: %w", err)
	}

	pricing, err := b.GetRatePricing(ctx, req.OrgID, req.RateID)
	if err != nil {
		return nil, nil, err
	}

	return charge, pricing, nil
}

func (b *businessLogic) UpdateRateCharge(ctx context.Context, req spec.UpdateRateChargeRequest) (*spec.RateChargeItem, *spec.RatePricingSummary, error) {
	rate, err := b.dl.GetRateByID(ctx, req.OrgID, req.RateID)
	if err != nil {
		return nil, nil, fmt.Errorf("rates.UpdateRateCharge verify rate: %w", err)
	}

	if rate.Status == RateStatusArchived {
		return nil, nil, fmt.Errorf("cannot modify charges on an archived rate record")
	}

	existing, err := b.dl.GetRateChargeByID(ctx, req.OrgID, req.RateID, req.ChargeID)
	if err != nil {
		return nil, nil, fmt.Errorf("rates.UpdateRateCharge get charge: %w", err)
	}

	if req.ChargeCategory != nil {
		existing.ChargeCategory = *req.ChargeCategory
	}
	if req.ChargeCode != nil {
		existing.ChargeCode = *req.ChargeCode
	}
	if req.ChargeName != nil {
		existing.ChargeName = *req.ChargeName
	}
	if req.CalculationBasis != nil {
		existing.CalculationBasis = *req.CalculationBasis
	}
	if req.Quantity != nil {
		if *req.Quantity < 0 {
			return nil, nil, fmt.Errorf("quantity cannot be negative")
		}
		existing.Quantity = *req.Quantity
	}
	if req.UnitPrice != nil {
		if *req.UnitPrice < 0 {
			return nil, nil, fmt.Errorf("unit price cannot be negative")
		}
		existing.UnitPrice = *req.UnitPrice
	}
	if req.Currency != nil {
		existing.Currency = *req.Currency
	}
	if req.MinimumAmount != nil {
		existing.MinimumAmount = req.MinimumAmount
	}
	if req.MaximumAmount != nil {
		existing.MaximumAmount = req.MaximumAmount
	}
	if req.IncludedInBaseRate != nil {
		existing.IncludedInBaseRate = *req.IncludedInBaseRate
	}
	if req.DisplayOrder != nil {
		existing.DisplayOrder = *req.DisplayOrder
	}
	if req.Notes != nil {
		existing.Notes = req.Notes
	}

	if err := b.dl.UpdateRateCharge(ctx, existing); err != nil {
		return nil, nil, fmt.Errorf("rates.UpdateRateCharge update: %w", err)
	}

	pricing, err := b.GetRatePricing(ctx, req.OrgID, req.RateID)
	if err != nil {
		return nil, nil, err
	}

	return existing, pricing, nil
}

func (b *businessLogic) DeleteRateCharge(ctx context.Context, req spec.DeleteRateChargeRequest) (*spec.RatePricingSummary, error) {
	rate, err := b.dl.GetRateByID(ctx, req.OrgID, req.RateID)
	if err != nil {
		return nil, fmt.Errorf("rates.DeleteRateCharge verify rate: %w", err)
	}

	if rate.Status == RateStatusArchived {
		return nil, fmt.Errorf("cannot delete charges from an archived rate record")
	}

	if err := b.dl.DeleteRateCharge(ctx, req.OrgID, req.RateID, req.ChargeID); err != nil {
		return nil, fmt.Errorf("rates.DeleteRateCharge delete: %w", err)
	}

	return b.GetRatePricing(ctx, req.OrgID, req.RateID)
}

func (b *businessLogic) ReorderRateCharges(ctx context.Context, req spec.ReorderRateChargesRequest) (*spec.RatePricingSummary, error) {
	rate, err := b.dl.GetRateByID(ctx, req.OrgID, req.RateID)
	if err != nil {
		return nil, fmt.Errorf("rates.ReorderRateCharges verify rate: %w", err)
	}

	if rate.Status == RateStatusArchived {
		return nil, fmt.Errorf("cannot reorder charges on an archived rate record")
	}

	if err := b.dl.ReorderRateCharges(ctx, req.OrgID, req.RateID, req.ChargeIDs); err != nil {
		return nil, fmt.Errorf("rates.ReorderRateCharges reorder: %w", err)
	}

	return b.GetRatePricing(ctx, req.OrgID, req.RateID)
}

// ── Task 19.3: Rate Contract & Versioning Business Logic ──────────────────────

func (b *businessLogic) CreateRateContract(ctx context.Context, req spec.CreateRateContractRequest) (*spec.RateContract, error) {
	if req.CarrierName == "" {
		return nil, fmt.Errorf("carrier_name is required")
	}
	if req.ContractName == "" {
		return nil, fmt.Errorf("contract_name is required")
	}
	if req.EffectiveDate == "" || req.ExpiryDate == "" {
		return nil, fmt.Errorf("effective_date and expiry_date are required (YYYY-MM-DD)")
	}

	effDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effective_date format (YYYY-MM-DD expected)")
	}
	expDate, err := time.Parse("2006-01-02", req.ExpiryDate)
	if err != nil {
		return nil, fmt.Errorf("invalid expiry_date format (YYYY-MM-DD expected)")
	}
	if expDate.Before(effDate) {
		return nil, fmt.Errorf("expiry_date cannot be earlier than effective_date")
	}

	ref := req.ContractReference
	if ref == "" {
		ref = fmt.Sprintf("CTR-%s-%d", strings.ToUpper(strings.ReplaceAll(req.CarrierName, " ", "")), time.Now().Unix()%100000)
	}

	status := CalculateRateContractStatus(effDate, expDate, ContractStatusActive)
	renewalStatus := CalculateContractRenewalStatus(expDate, RenewalStatusNotStarted)

	author := req.Author
	contract := &spec.RateContract{
		OrgID:             req.OrgID,
		ContractReference: ref,
		CarrierName:       req.CarrierName,
		CarrierCode:       req.CarrierCode,
		ContractName:      req.ContractName,
		ContractType:      req.ContractType,
		TransportMode:     req.TransportMode,
		Currency:          req.Currency,
		EffectiveDate:     effDate,
		ExpiryDate:        expDate,
		Status:            status,
		RenewalStatus:     renewalStatus,
		RenewalOwner:      req.RenewalOwner,
		Notes:             req.Notes,
		CreatedBy:         &author,
		UpdatedBy:         &author,
	}

	if err := b.dl.CreateRateContract(ctx, contract); err != nil {
		return nil, err
	}

	return contract, nil
}

func (b *businessLogic) GetRateContract(ctx context.Context, orgID, contractID int64) (*spec.RateContract, error) {
	contract, err := b.dl.GetRateContractByID(ctx, orgID, contractID)
	if err != nil {
		return nil, err
	}
	// Compute dynamic status
	contract.Status = CalculateRateContractStatus(contract.EffectiveDate, contract.ExpiryDate, contract.Status)
	contract.RenewalStatus = CalculateContractRenewalStatus(contract.ExpiryDate, contract.RenewalStatus)
	return contract, nil
}

func (b *businessLogic) ListRateContracts(ctx context.Context, filters spec.ListRateContractsRequest) (*spec.ListRateContractsResponse, error) {
	items, total, err := b.dl.ListRateContracts(ctx, &filters)
	if err != nil {
		return nil, err
	}

	limit := filters.Limit
	if limit <= 0 {
		limit = 15
	}
	page := filters.Page
	if page <= 0 {
		page = 1
	}
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	return &spec.ListRateContractsResponse{
		Contracts:  items,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (b *businessLogic) UpdateRateContract(ctx context.Context, req spec.UpdateRateContractRequest) (*spec.RateContract, error) {
	contract, err := b.dl.GetRateContractByID(ctx, req.OrgID, req.ID)
	if err != nil {
		return nil, err
	}

	if contract.Status == ContractStatusArchived {
		return nil, fmt.Errorf("cannot modify an archived contract")
	}

	if req.ContractReference != nil && *req.ContractReference != "" {
		contract.ContractReference = *req.ContractReference
	}
	if req.CarrierName != nil && *req.CarrierName != "" {
		contract.CarrierName = *req.CarrierName
	}
	if req.CarrierCode != nil {
		contract.CarrierCode = req.CarrierCode
	}
	if req.ContractName != nil && *req.ContractName != "" {
		contract.ContractName = *req.ContractName
	}
	if req.ContractType != nil && *req.ContractType != "" {
		contract.ContractType = *req.ContractType
	}
	if req.TransportMode != nil {
		contract.TransportMode = req.TransportMode
	}
	if req.Currency != nil {
		contract.Currency = req.Currency
	}
	if req.EffectiveDate != nil && *req.EffectiveDate != "" {
		eff, err := time.Parse("2006-01-02", *req.EffectiveDate)
		if err != nil {
			return nil, fmt.Errorf("invalid effective_date format (YYYY-MM-DD expected)")
		}
		contract.EffectiveDate = eff
	}
	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		exp, err := time.Parse("2006-01-02", *req.ExpiryDate)
		if err != nil {
			return nil, fmt.Errorf("invalid expiry_date format (YYYY-MM-DD expected)")
		}
		contract.ExpiryDate = exp
	}
	if contract.ExpiryDate.Before(contract.EffectiveDate) {
		return nil, fmt.Errorf("expiry_date cannot be earlier than effective_date")
	}

	if req.Status != nil && *req.Status != "" {
		contract.Status = *req.Status
	} else {
		contract.Status = CalculateRateContractStatus(contract.EffectiveDate, contract.ExpiryDate, contract.Status)
	}

	if req.RenewalStatus != nil && *req.RenewalStatus != "" {
		contract.RenewalStatus = *req.RenewalStatus
	} else {
		contract.RenewalStatus = CalculateContractRenewalStatus(contract.ExpiryDate, contract.RenewalStatus)
	}

	if req.RenewalOwner != nil {
		contract.RenewalOwner = req.RenewalOwner
	}
	if req.Notes != nil {
		contract.Notes = req.Notes
	}

	contract.UpdatedBy = &req.Updater

	if err := b.dl.UpdateRateContract(ctx, contract); err != nil {
		return nil, err
	}

	return contract, nil
}

func (b *businessLogic) ArchiveRateContract(ctx context.Context, orgID, contractID int64, archivedBy string) error {
	return b.dl.ArchiveRateContract(ctx, orgID, contractID, archivedBy)
}

func (b *businessLogic) RenewRateContract(ctx context.Context, req spec.RenewRateContractRequest) (*spec.RateContract, error) {
	contract, err := b.dl.GetRateContractByID(ctx, req.OrgID, req.ID)
	if err != nil {
		return nil, err
	}

	if req.NewExpiryDate != "" {
		newExp, err := time.Parse("2006-01-02", req.NewExpiryDate)
		if err != nil {
			return nil, fmt.Errorf("invalid new_expiry_date format (YYYY-MM-DD expected)")
		}
		if newExp.Before(contract.EffectiveDate) {
			return nil, fmt.Errorf("new_expiry_date cannot be earlier than effective_date")
		}
		contract.ExpiryDate = newExp
	}

	renewalStatus := req.RenewalStatus
	if renewalStatus == "" {
		renewalStatus = RenewalStatusRenewed
	}
	contract.RenewalStatus = renewalStatus
	contract.Status = CalculateRateContractStatus(contract.EffectiveDate, contract.ExpiryDate, ContractStatusActive)
	contract.UpdatedBy = &req.User
	if req.Notes != nil {
		contract.Notes = req.Notes
	}

	if err := b.dl.UpdateRateContract(ctx, contract); err != nil {
		return nil, err
	}

	return contract, nil
}

func (b *businessLogic) GetRateContractSummary(ctx context.Context, orgID int64) (*spec.RateContractSummary, error) {
	return b.dl.GetRateContractSummary(ctx, orgID)
}

func (b *businessLogic) GetRatesByContract(ctx context.Context, orgID, contractID int64) ([]*spec.RateListItem, error) {
	return b.dl.GetRatesByContract(ctx, orgID, contractID)
}

// ── Rate Versioning Core Workflow ─────────────────────────────────────────────

func (b *businessLogic) CreateNewRateVersion(ctx context.Context, req spec.CreateRateVersionRequest) (*spec.Rate, error) {
	// 1. Fetch current rate
	parentRate, err := b.dl.GetRateByID(ctx, req.OrgID, req.RateID)
	if err != nil {
		return nil, fmt.Errorf("rates.CreateNewRateVersion get parent rate: %w", err)
	}

	if parentRate.Status == RateStatusArchived {
		return nil, fmt.Errorf("cannot create new version of an archived rate")
	}

	// 2. Fetch charge items snapshot from parent rate
	parentCharges, err := b.dl.GetRateCharges(ctx, req.OrgID, req.RateID)
	if err != nil {
		return nil, fmt.Errorf("rates.CreateNewRateVersion get parent charges: %w", err)
	}

	// 3. Prepare new rate entity cloning parent fields
	newVersionNumber := parentRate.VersionNumber + 1
	author := req.User
	now := time.Now().UTC()

	baseAmount := parentRate.BaseAmount
	if req.BaseAmount != nil {
		baseAmount = *req.BaseAmount
	}
	currency := parentRate.Currency
	if req.Currency != nil && *req.Currency != "" {
		currency = *req.Currency
	}
	effDate := parentRate.EffectiveDate
	if req.EffectiveDate != nil && *req.EffectiveDate != "" {
		if d, err := time.Parse("2006-01-02", *req.EffectiveDate); err == nil {
			effDate = d
		}
	}
	expDate := parentRate.ExpiryDate
	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		if d, err := time.Parse("2006-01-02", *req.ExpiryDate); err == nil {
			expDate = d
		}
	}

	status, _ := CalculateRateValidity(effDate, expDate, RateStatusActive, now)

	carrierRef := parentRate.CarrierReference
	if req.CarrierReference != nil {
		carrierRef = req.CarrierReference
	}
	contractRef := parentRate.ContractReference
	if req.ContractReference != nil {
		contractRef = req.ContractReference
	}
	notes := parentRate.Notes
	if req.Notes != nil {
		notes = req.Notes
	}

	parentRateID := parentRate.ID
	newRate := &spec.Rate{
		OrgID:              req.OrgID,
		RateReference:      parentRate.RateReference,
		CarrierName:        parentRate.CarrierName,
		CarrierCode:        parentRate.CarrierCode,
		ServiceProvider:    parentRate.ServiceProvider,
		RateType:           parentRate.RateType,
		TransportMode:      parentRate.TransportMode,
		ServiceType:        parentRate.ServiceType,
		EquipmentType:      parentRate.EquipmentType,
		OriginPort:         parentRate.OriginPort,
		OriginCode:         parentRate.OriginCode,
		DestinationPort:    parentRate.DestinationPort,
		DestinationCode:    parentRate.DestinationCode,
		Currency:           currency,
		BaseAmount:         baseAmount,
		EffectiveDate:      effDate,
		ExpiryDate:         expDate,
		Status:             status,
		ContractID:         parentRate.ContractID,
		VersionNumber:      newVersionNumber,
		VersionStatus:      VersionStatusCurrent,
		SupersedesRateID:   &parentRateID,
		SupersededByRateID: nil,
		VersionCreatedAt:   &now,
		CarrierReference:   carrierRef,
		ContractReference:  contractRef,
		Notes:              notes,
		CreatedBy:          &author,
		UpdatedBy:          &author,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	// 4. Construct audit description
	desc := fmt.Sprintf("Version %d created by %s. Revision: %s", newVersionNumber, author, req.RevisionReason)
	if baseAmount != parentRate.BaseAmount {
		desc += fmt.Sprintf(" (Base: %s %.2f → %.2f)", currency, parentRate.BaseAmount, baseAmount)
	}

	history := &spec.RateVersionHistory{
		OrgID:          req.OrgID,
		VersionNumber:  newVersionNumber,
		Action:         ActionRateVersionCreated,
		PreviousRateID: &parentRateID,
		Description:    desc,
		PerformedBy:    &author,
		CreatedAt:      now,
	}

	// 5. Persist rate version record & supersede parent
	if err := b.dl.CreateRateVersionRecord(ctx, newRate, history); err != nil {
		return nil, fmt.Errorf("rates.CreateNewRateVersion record: %w", err)
	}

	// 6. Snapshot and clone all charge items to the new version
	if len(req.ChargeUpdates) > 0 {
		for i, cu := range req.ChargeUpdates {
			charge := &spec.RateChargeItem{
				OrgID:              req.OrgID,
				RateID:             newRate.ID,
				ChargeCategory:     cu.ChargeCategory,
				ChargeCode:         cu.ChargeCode,
				ChargeName:         cu.ChargeName,
				CalculationBasis:   cu.CalculationBasis,
				Quantity:           cu.Quantity,
				UnitPrice:          cu.UnitPrice,
				Currency:           cu.Currency,
				MinimumAmount:      cu.MinimumAmount,
				MaximumAmount:      cu.MaximumAmount,
				IncludedInBaseRate: cu.IncludedInBaseRate,
				DisplayOrder:       i,
				Notes:              cu.Notes,
			}
			_ = b.dl.CreateRateCharge(ctx, charge)
		}
	} else {
		for _, pc := range parentCharges {
			clonedCharge := &spec.RateChargeItem{
				OrgID:              req.OrgID,
				RateID:             newRate.ID,
				ChargeCategory:     pc.ChargeCategory,
				ChargeCode:         pc.ChargeCode,
				ChargeName:         pc.ChargeName,
				CalculationBasis:   pc.CalculationBasis,
				Quantity:           pc.Quantity,
				UnitPrice:          pc.UnitPrice,
				Currency:           pc.Currency,
				MinimumAmount:      pc.MinimumAmount,
				MaximumAmount:      pc.MaximumAmount,
				IncludedInBaseRate: pc.IncludedInBaseRate,
				DisplayOrder:       pc.DisplayOrder,
				Notes:              pc.Notes,
			}
			_ = b.dl.CreateRateCharge(ctx, clonedCharge)
		}
	}

	_, _ = audit.Record(ctx, auditDomain.CreateAuditLogParams{
		OrgID:        req.OrgID,
		Action:       auditDomain.ActionCreate,
		Module:       auditDomain.ModuleRateManagement,
		ResourceType: "RATE",
		ResourceID:   fmt.Sprintf("%d", newRate.ID),
		ResourceName: fmt.Sprintf("%s → %s (%s)", newRate.OriginPort, newRate.DestinationPort, newRate.CarrierName),
		Description:  desc,
		Result:       auditDomain.ResultSuccess,
		Metadata: map[string]interface{}{
			"version_number": newVersionNumber,
			"carrier_name":   newRate.CarrierName,
			"origin":         newRate.OriginPort,
			"destination":    newRate.DestinationPort,
			"base_amount":    newRate.BaseAmount,
			"currency":       newRate.Currency,
		},
	})

	return newRate, nil
}

func (b *businessLogic) GetRateVersionHistory(ctx context.Context, orgID, rateID int64) ([]spec.RateVersionHistory, error) {
	return b.dl.GetRateVersionHistory(ctx, orgID, rateID)
}

func (b *businessLogic) GetRateVersionChain(ctx context.Context, orgID, rateID int64) ([]spec.RateVersionChainItem, error) {
	chain, err := b.dl.GetRateVersionChain(ctx, orgID, rateID)
	if err != nil {
		return nil, err
	}

	// Enrich commercial totals from pricing engine for each item in lineage
	for i := range chain {
		if pricing, err := b.GetRatePricing(ctx, orgID, chain[i].RateID); err == nil {
			chain[i].CommercialTotal = pricing.CommercialTotal
			chain[i].ChargeCount = pricing.ChargeCount
		}
	}

	return chain, nil
}

// ── Task 19.4: Spot Rate Requests, Responses & Comparison BL Implementation ──

func (b *businessLogic) CreateSpotRateRequest(ctx context.Context, req spec.CreateSpotRateRequestRequest) (*spec.SpotRateRequest, error) {
	if req.OriginPort == "" || req.DestinationPort == "" {
		return nil, fmt.Errorf("origin and destination ports are required")
	}
	if req.RequiredByDate == "" {
		return nil, fmt.Errorf("required_by_date is required")
	}
	if req.ReadyDate == "" {
		return nil, fmt.Errorf("ready_date is required")
	}
	if req.TransportMode == "" {
		req.TransportMode = "Ocean FCL"
	}
	if req.ServiceType == "" {
		req.ServiceType = "FCL"
	}
	if req.TargetCurrency == "" {
		req.TargetCurrency = "USD"
	}
	if req.ContainerQuantity <= 0 {
		req.ContainerQuantity = 1
	}

	ref := req.RequestReference
	if ref == "" {
		ref = fmt.Sprintf("SPOT-%s-%d", time.Now().Format("20060102"), time.Now().Unix()%10000)
	}

	record := &spec.SpotRateRequest{
		OrgID:             req.OrgID,
		RequestReference:  ref,
		CustomerID:        req.CustomerID,
		CustomerName:      req.CustomerName,
		OriginPort:        req.OriginPort,
		OriginCode:        req.OriginCode,
		DestinationPort:   req.DestinationPort,
		DestinationCode:   req.DestinationCode,
		TransportMode:     req.TransportMode,
		ServiceType:       req.ServiceType,
		EquipmentType:     req.EquipmentType,
		Commodity:         req.Commodity,
		CargoWeight:       req.CargoWeight,
		CargoVolume:       req.CargoVolume,
		ContainerQuantity: req.ContainerQuantity,
		ReadyDate:         req.ReadyDate,
		TargetCurrency:    req.TargetCurrency,
		RequiredByDate:    req.RequiredByDate,
		Status:            SpotRequestSent,
		Notes:             req.Notes,
		CreatedBy:         &req.User,
	}

	if err := b.dl.CreateSpotRateRequest(ctx, record); err != nil {
		return nil, fmt.Errorf("rates.CreateSpotRateRequest: %w", err)
	}
	return record, nil
}

func (b *businessLogic) GetSpotRateRequest(ctx context.Context, orgID, requestID int64) (*spec.SpotRateRequest, error) {
	req, err := b.dl.GetSpotRateRequestByID(ctx, orgID, requestID)
	if err != nil {
		return nil, err
	}

	// Update status dynamically based on current responses & time
	responses, err := b.dl.GetSpotRateResponsesByRequestID(ctx, orgID, requestID)
	if err == nil {
		calculatedStatus := CalculateSpotRateRequestStatus(req, responses, time.Now().UTC())
		if calculatedStatus != req.Status && req.Status != SpotRequestCancelled {
			req.Status = calculatedStatus
			_ = b.dl.UpdateSpotRateRequest(ctx, req)
		}
	}

	return req, nil
}

func (b *businessLogic) ListSpotRateRequests(ctx context.Context, filters spec.ListSpotRateRequestsRequest) (*spec.ListSpotRateRequestsResponse, error) {
	items, totalCount, err := b.dl.ListSpotRateRequests(ctx, &filters)
	if err != nil {
		return nil, err
	}

	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	page := filters.Page
	if page <= 0 {
		page = 1
	}
	totalPages := (totalCount + limit - 1) / limit

	return &spec.ListSpotRateRequestsResponse{
		Requests:   items,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (b *businessLogic) UpdateSpotRateRequest(ctx context.Context, req spec.UpdateSpotRateRequestRequest) (*spec.SpotRateRequest, error) {
	existing, err := b.dl.GetSpotRateRequestByID(ctx, req.OrgID, req.ID)
	if err != nil {
		return nil, fmt.Errorf("rates.UpdateSpotRateRequest find: %w", err)
	}

	if req.CustomerName != nil {
		existing.CustomerName = req.CustomerName
	}
	if req.TransportMode != nil {
		existing.TransportMode = *req.TransportMode
	}
	if req.ServiceType != nil {
		existing.ServiceType = *req.ServiceType
	}
	if req.EquipmentType != nil {
		existing.EquipmentType = req.EquipmentType
	}
	if req.Commodity != nil {
		existing.Commodity = req.Commodity
	}
	if req.CargoWeight != nil {
		existing.CargoWeight = req.CargoWeight
	}
	if req.CargoVolume != nil {
		existing.CargoVolume = req.CargoVolume
	}
	if req.ContainerQuantity != nil {
		existing.ContainerQuantity = *req.ContainerQuantity
	}
	if req.ReadyDate != nil {
		existing.ReadyDate = *req.ReadyDate
	}
	if req.TargetCurrency != nil {
		existing.TargetCurrency = *req.TargetCurrency
	}
	if req.RequiredByDate != nil {
		existing.RequiredByDate = *req.RequiredByDate
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	if req.Notes != nil {
		existing.Notes = req.Notes
	}

	if err := b.dl.UpdateSpotRateRequest(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (b *businessLogic) SendSpotRateRequest(ctx context.Context, orgID, requestID int64, user string) (*spec.SpotRateRequest, error) {
	req, err := b.dl.GetSpotRateRequestByID(ctx, orgID, requestID)
	if err != nil {
		return nil, err
	}
	req.Status = SpotRequestSent
	if err := b.dl.UpdateSpotRateRequest(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (b *businessLogic) CancelSpotRateRequest(ctx context.Context, orgID, requestID int64, user string) error {
	return b.dl.CancelSpotRateRequest(ctx, orgID, requestID)
}

func (b *businessLogic) CreateSpotRateResponse(ctx context.Context, req spec.CreateSpotRateResponseRequest) (*spec.SpotRateResponse, error) {
	request, err := b.dl.GetSpotRateRequestByID(ctx, req.OrgID, req.SpotRateRequestID)
	if err != nil {
		return nil, fmt.Errorf("rates.CreateSpotRateResponse verify request: %w", err)
	}

	if req.CarrierName == "" {
		return nil, fmt.Errorf("carrier name is required")
	}
	if req.Currency == "" {
		req.Currency = request.TargetCurrency
	}
	if req.ValidFrom == "" {
		req.ValidFrom = time.Now().Format("2006-01-02")
	}
	if req.ValidUntil == "" {
		req.ValidUntil = request.RequiredByDate
	}

	response := &spec.SpotRateResponse{
		OrgID:                req.OrgID,
		SpotRateRequestID:    req.SpotRateRequestID,
		CarrierName:          req.CarrierName,
		CarrierCode:          req.CarrierCode,
		SupplierName:         req.SupplierName,
		RateID:               req.RateID,
		Currency:             req.Currency,
		BaseAmount:           req.BaseAmount,
		TransitDays:          req.TransitDays,
		FreeDaysOrigin:       req.FreeDaysOrigin,
		FreeDaysDestination:  req.FreeDaysDestination,
		ValidFrom:            req.ValidFrom,
		ValidUntil:           req.ValidUntil,
		RoutingNotes:         req.RoutingNotes,
		ResponseNotes:        req.ResponseNotes,
		Status:               SpotResponseReceived,
		IsPreferred:          false,
		CreatedBy:            &req.User,
	}

	var charges []spec.SpotRateResponseCharge
	for i, c := range req.Charges {
		charges = append(charges, spec.SpotRateResponseCharge{
			OrgID:            req.OrgID,
			ChargeCategory:   c.ChargeCategory,
			ChargeName:       c.ChargeName,
			CalculationBasis: c.CalculationBasis,
			Quantity:         c.Quantity,
			UnitPrice:        c.UnitPrice,
			Currency:         c.Currency,
			DisplayOrder:     i,
		})
	}

	if err := b.dl.CreateSpotRateResponse(ctx, response, charges); err != nil {
		return nil, fmt.Errorf("rates.CreateSpotRateResponse: %w", err)
	}

	// Re-evaluate request status
	responses, _ := b.dl.GetSpotRateResponsesByRequestID(ctx, req.OrgID, req.SpotRateRequestID)
	newStatus := CalculateSpotRateRequestStatus(request, responses, time.Now().UTC())
	if newStatus != request.Status {
		request.Status = newStatus
		_ = b.dl.UpdateSpotRateRequest(ctx, request)
	}

	return b.dl.GetSpotRateResponseByID(ctx, req.OrgID, response.ID)
}

func (b *businessLogic) GetSpotRateResponse(ctx context.Context, orgID, responseID int64) (*spec.SpotRateResponse, error) {
	return b.dl.GetSpotRateResponseByID(ctx, orgID, responseID)
}

func (b *businessLogic) GetSpotRateResponses(ctx context.Context, orgID, requestID int64) ([]*spec.SpotRateResponse, error) {
	return b.dl.GetSpotRateResponsesByRequestID(ctx, orgID, requestID)
}

func (b *businessLogic) UpdateSpotRateResponse(ctx context.Context, req spec.UpdateSpotRateResponseRequest) (*spec.SpotRateResponse, error) {
	existing, err := b.dl.GetSpotRateResponseByID(ctx, req.OrgID, req.ResponseID)
	if err != nil {
		return nil, fmt.Errorf("rates.UpdateSpotRateResponse find: %w", err)
	}

	if req.CarrierName != nil {
		existing.CarrierName = *req.CarrierName
	}
	if req.CarrierCode != nil {
		existing.CarrierCode = req.CarrierCode
	}
	if req.SupplierName != nil {
		existing.SupplierName = req.SupplierName
	}
	if req.Currency != nil {
		existing.Currency = *req.Currency
	}
	if req.BaseAmount != nil {
		existing.BaseAmount = *req.BaseAmount
	}
	if req.TransitDays != nil {
		existing.TransitDays = req.TransitDays
	}
	if req.FreeDaysOrigin != nil {
		existing.FreeDaysOrigin = *req.FreeDaysOrigin
	}
	if req.FreeDaysDestination != nil {
		existing.FreeDaysDestination = *req.FreeDaysDestination
	}
	if req.ValidFrom != nil {
		existing.ValidFrom = *req.ValidFrom
	}
	if req.ValidUntil != nil {
		existing.ValidUntil = *req.ValidUntil
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	if req.RoutingNotes != nil {
		existing.RoutingNotes = req.RoutingNotes
	}
	if req.ResponseNotes != nil {
		existing.ResponseNotes = req.ResponseNotes
	}

	var charges []spec.SpotRateResponseCharge
	for i, c := range req.Charges {
		charges = append(charges, spec.SpotRateResponseCharge{
			OrgID:            req.OrgID,
			SpotRateResponseID: existing.ID,
			ChargeCategory:   c.ChargeCategory,
			ChargeName:       c.ChargeName,
			CalculationBasis: c.CalculationBasis,
			Quantity:         c.Quantity,
			UnitPrice:        c.UnitPrice,
			Currency:         c.Currency,
			DisplayOrder:     i,
		})
	}

	if err := b.dl.UpdateSpotRateResponse(ctx, existing, charges); err != nil {
		return nil, err
	}

	return b.dl.GetSpotRateResponseByID(ctx, req.OrgID, existing.ID)
}

func (b *businessLogic) SelectPreferredSpotRate(ctx context.Context, req spec.SelectPreferredSpotRateRequest) (*spec.SpotRateResponse, error) {
	// 1. Verify existence & ownership
	response, err := b.dl.GetSpotRateResponseByID(ctx, req.OrgID, req.ResponseID)
	if err != nil {
		return nil, fmt.Errorf("rates.SelectPreferredSpotRate find response: %w", err)
	}

	if response.SpotRateRequestID != req.SpotRateRequestID {
		return nil, fmt.Errorf("response does not belong to specified spot request")
	}

	// 2. Select preferred response safely (preserves competing responses)
	if err := b.dl.SelectPreferredSpotRateResponse(ctx, req.OrgID, req.SpotRateRequestID, req.ResponseID); err != nil {
		return nil, fmt.Errorf("rates.SelectPreferredSpotRate: %w", err)
	}

	return b.dl.GetSpotRateResponseByID(ctx, req.OrgID, req.ResponseID)
}

func (b *businessLogic) CompareSpotRates(ctx context.Context, orgID, requestID int64) (*spec.SpotRateComparison, error) {
	request, err := b.dl.GetSpotRateRequestByID(ctx, orgID, requestID)
	if err != nil {
		return nil, fmt.Errorf("rates.CompareSpotRates get request: %w", err)
	}

	responses, err := b.dl.GetSpotRateResponsesByRequestID(ctx, orgID, requestID)
	if err != nil {
		return nil, fmt.Errorf("rates.CompareSpotRates get responses: %w", err)
	}

	comp := CalculateSpotRateComparison(request, responses)
	return &comp, nil
}

func (b *businessLogic) GetSpotRateSummary(ctx context.Context, orgID int64) (*spec.SpotRateRequestSummary, error) {
	return b.dl.GetSpotRateRequestSummary(ctx, orgID)
}

// ── Task 19.6: Rate Lifecycle Intelligence Implementations ────────────────────

func (b *businessLogic) GetRateLifecycleDashboard(ctx context.Context, orgID int64) (*RateLifecycleSummary, error) {
	return b.dl.GetRateLifecycleSummary(ctx, orgID)
}

func (b *businessLogic) GetRateLifecycleEvents(ctx context.Context, orgID int64, limit int) ([]*RateLifecycleEvent, error) {
	return b.dl.GetRateLifecycleEvents(ctx, orgID, limit)
}

func (b *businessLogic) GetRatesRequiringAttention(ctx context.Context, orgID int64) ([]*RateAttentionItem, error) {
	return b.dl.GetRatesRequiringAttention(ctx, orgID)
}

func (b *businessLogic) GetContractsRequiringAttention(ctx context.Context, orgID int64) ([]*ContractAttentionItem, error) {
	return b.dl.GetContractsRequiringAttention(ctx, orgID)
}

func (b *businessLogic) EvaluateRateLifecycleForOrg(ctx context.Context, orgID int64) (*RateLifecycleSummary, error) {
	now := time.Now()

	// 1. Evaluate Rates
	ratesList, err := b.dl.GetAllActiveAndExpiringRatesForEvaluation(ctx, orgID)
	if err == nil {
		for _, r := range ratesList {
			if r.Status == RateStatusArchived || r.Status == "SUPERSEDED" {
				continue
			}
			newStatus, _ := EvaluateRateLifecycleStatus(r.Status, &r.EffectiveDate, &r.ExpiryDate, now)
			if newStatus != r.Status {
				_ = b.dl.UpdateRateStatusDirect(ctx, orgID, r.ID, newStatus)

				var eventType string
				if newStatus == RateStatusExpired {
					eventType = EventRateExpired
				} else if newStatus == RateStatusExpiringSoon {
					eventType = EventRateExpiringSoon
				}

				if eventType != "" {
					desc := fmt.Sprintf("Rate %s (%s) transitioned from %s to %s", r.CarrierName, r.TransportMode, r.Status, newStatus)
					_ = b.dl.CreateRateLifecycleEvent(ctx, &RateLifecycleEvent{
						OrgID:          orgID,
						RateID:         &r.ID,
						EventType:      eventType,
						PreviousStatus: r.Status,
						CurrentStatus:  newStatus,
						Description:    desc,
					})
				}
			}
		}
	}

	// 2. Evaluate Contracts
	contractsList, err := b.dl.GetAllActiveAndExpiringContractsForEvaluation(ctx, orgID)
	if err == nil {
		for _, c := range contractsList {
			if c.Status == ContractStatusArchived {
				continue
			}
			newStatus, newRenewal, _ := EvaluateContractLifecycleStatus(c.Status, c.RenewalStatus, &c.ExpiryDate, now)
			if newStatus != c.Status || newRenewal != c.RenewalStatus {
				_ = b.dl.UpdateContractStatusDirect(ctx, orgID, c.ID, newStatus, newRenewal)

				var eventType string
				if newStatus == ContractStatusExpired {
					eventType = EventContractExpired
				} else if newStatus == ContractStatusExpiringSoon {
					eventType = EventContractExpiringSoon
				}

				if eventType != "" {
					desc := fmt.Sprintf("Contract %s (%s) transitioned from %s to %s", c.ContractReference, c.CarrierName, c.Status, newStatus)
					_ = b.dl.CreateRateLifecycleEvent(ctx, &RateLifecycleEvent{
						OrgID:          orgID,
						ContractID:     &c.ID,
						EventType:      eventType,
						PreviousStatus: c.Status,
						CurrentStatus:  newStatus,
						Description:    desc,
					})
				}
			}
		}
	}


	return b.dl.GetRateLifecycleSummary(ctx, orgID)
}

// ── Task 19.7: Rate Analytics & Procurement Intelligence ─────────────────────

func (b *businessLogic) GetRateAnalyticsOverview(ctx context.Context, orgID int64) (*RateAnalyticsOverview, error) {
	result, err := b.dl.GetRateAnalyticsOverview(ctx, orgID)
	if err != nil {
		return &RateAnalyticsOverview{}, nil // safe zero-value for empty orgs
	}
	return result, nil
}

func (b *businessLogic) GetRateAnalyticsTrends(ctx context.Context, orgID int64, days int) ([]RateTrendDataPoint, error) {
	// Validate days: accept only 7, 30, 90; default to 30
	switch days {
	case 7, 30, 90:
		// valid
	default:
		days = 30
	}
	result, err := b.dl.GetRateAnalyticsTrends(ctx, orgID, days)
	if err != nil {
		return []RateTrendDataPoint{}, nil
	}
	return result, nil
}

func (b *businessLogic) GetCarrierRatePerformance(ctx context.Context, orgID int64) ([]CarrierRatePerformance, error) {
	result, err := b.dl.GetCarrierRatePerformance(ctx, orgID)
	if err != nil {
		return []CarrierRatePerformance{}, nil
	}
	return result, nil
}

func (b *businessLogic) GetLaneRatePerformance(ctx context.Context, orgID int64) ([]LaneRatePerformance, error) {
	result, err := b.dl.GetLaneRatePerformance(ctx, orgID)
	if err != nil {
		return []LaneRatePerformance{}, nil
	}
	return result, nil
}

func (b *businessLogic) GetRateLifecycleAnalytics(ctx context.Context, orgID int64) (*RateLifecycleAnalytics, error) {
	result, err := b.dl.GetRateLifecycleAnalytics(ctx, orgID)
	if err != nil {
		return &RateLifecycleAnalytics{}, nil
	}
	return result, nil
}

func (b *businessLogic) GetSpotSourcingPerformance(ctx context.Context, orgID int64) (*SpotSourcingPerformance, error) {
	result, err := b.dl.GetSpotSourcingPerformance(ctx, orgID)
	if err != nil {
		return &SpotSourcingPerformance{CarrierParticipation: []SpotCarrierParticipation{}}, nil
	}
	return result, nil
}

func (b *businessLogic) GetRateCommercialInsights(ctx context.Context, orgID int64) ([]CommercialImpactInsight, error) {
	// Aggregate all analytics data, then run deterministic engine
	overview, _ := b.dl.GetRateAnalyticsOverview(ctx, orgID)
	lifecycle, _ := b.dl.GetRateLifecycleAnalytics(ctx, orgID)
	spot, _ := b.dl.GetSpotSourcingPerformance(ctx, orgID)
	carriers, _ := b.dl.GetCarrierRatePerformance(ctx, orgID)
	lanes, _ := b.dl.GetLaneRatePerformance(ctx, orgID)
	riskExposure, _ := b.dl.GetCommercialRiskExposure(ctx, orgID)

	// Ensure nil safety
	if overview == nil {
		overview = &RateAnalyticsOverview{}
	}
	if lifecycle == nil {
		lifecycle = &RateLifecycleAnalytics{}
	}
	if spot == nil {
		spot = &SpotSourcingPerformance{}
	}

	insights := GenerateRateCommercialInsights(overview, lifecycle, spot, carriers, lanes, riskExposure)
	if insights == nil {
		insights = []CommercialImpactInsight{}
	}
	return insights, nil
}
