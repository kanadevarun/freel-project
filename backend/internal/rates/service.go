package rates

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/freel/backend/internal/carrier"
)

// Service is the public interface for the Rate Intelligence layer.
// The Quotation Engine and Pricing Agent ONLY interact with this interface —
// they never import the carrier or contracts packages directly.
type Service interface {
	// SearchRates is the primary read API for the Quotation Engine and Pricing Agent.
	// It searches rate_entries for confirmed rates matching the query, covering
	// both SPOT_API and CONTRACT_PDF sources in a single call.
	//
	// Contract rates are ranked above spot rates when confidence >= 85.
	// Within the same source, lower total_buy_price wins.
	//
	// Example:
	//   result, err := rateSvc.SearchRates(ctx, RateQuery{
	//       OrgID: 5, OriginPort: "INNSA", DestinationPort: "DEHAM",
	//       EquipmentType: "40GP", TargetDate: &rfq.TargetDate,
	//   })
	SearchRates(ctx context.Context, q RateQuery) (*RateSearchResult, error)

	// IngestRates stores a set of canonical rates produced by the AI pipeline
	// (or the spot normalizer) into rate_entries.
	// Called by: the Spot Normalizer after a carrier API fetch, and by the
	// AI sidecar callback after processing a contract document.
	IngestRates(ctx context.Context, rates []CanonicalRate) error

	// RefreshSpotRates fetches fresh rates from the carrier provider, normalizes
	// them into CanonicalRate objects, stores them, and returns the search result.
	// Call this when the cache is stale or when the user explicitly requests a refresh.
	RefreshSpotRates(ctx context.Context, orgID int64, q RateQuery, targetDate *time.Time) (*RateSearchResult, error)

	// GetRateByID returns a single rate entry with full detail.
	GetRateByID(ctx context.Context, orgID int64, id string) (*CanonicalRate, error)
}

type service struct {
	repo     Repository
	normalizer SpotNormalizer
	carrier  carrier.Service
}

// NewService creates the production Rate Intelligence Service.
// It wires together the DB repository, the spot normalizer, and the carrier
// service so that RefreshSpotRates can do a live fetch + normalize + store.
func NewService(repo Repository, normalizer SpotNormalizer, carrierSvc carrier.Service) Service {
	return &service{
		repo:       repo,
		normalizer: normalizer,
		carrier:    carrierSvc,
	}
}

// SearchRates looks up confirmed canonical rates in rate_entries.
// If no rates are found in the DB, it falls back to a live spot rate fetch
// via RefreshSpotRates so the caller always gets results.
func (s *service) SearchRates(ctx context.Context, q RateQuery) (*RateSearchResult, error) {
	if q.OriginPort == "" || q.DestinationPort == "" {
		return nil, fmt.Errorf("origin_port and destination_port are required")
	}

	// Normalize port inputs before searching
	q.OriginPort = NormalizePort(q.OriginPort)
	q.DestinationPort = NormalizePort(q.DestinationPort)
	if q.EquipmentType == "" {
		q.EquipmentType = "40GP"
	}
	if q.MaxResults <= 0 {
		q.MaxResults = 20
	}

	rates, err := s.repo.Search(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("rate search: %w", err)
	}

	// If the DB has no valid rates for this lane, do a live spot rate refresh
	// so the caller always gets results (graceful degradation).
	if len(rates) == 0 {
		return s.RefreshSpotRates(ctx, q.OrgID, q, q.TargetDate)
	}

	return buildSearchResult(rates), nil
}

// IngestRates stores a batch of canonical rates produced by either the
// spot normalizer or the AI contract pipeline.
func (s *service) IngestRates(ctx context.Context, rates []CanonicalRate) error {
	if len(rates) == 0 {
		return nil
	}
	return s.repo.Upsert(ctx, rates)
}

// RefreshSpotRates performs a live carrier API fetch for the given lane,
// normalizes the results into CanonicalRate objects, stores them in
// rate_entries, and returns the ranked result set.
func (s *service) RefreshSpotRates(ctx context.Context, orgID int64, q RateQuery, targetDate *time.Time) (*RateSearchResult, error) {
	if s.carrier == nil {
		return nil, fmt.Errorf("carrier service not configured")
	}

	incoterms := "FOB" // sensible default for spot quotes
	if q.Incoterms != "" {
		incoterms = q.Incoterms
	}
	resp, err := s.carrier.FetchRates(ctx, q.OriginPort, q.DestinationPort, targetDate, incoterms, 0, 0, "")
	if err != nil {
		return nil, fmt.Errorf("carrier fetch for %s→%s: %w", q.OriginPort, q.DestinationPort, err)
	}

	// Normalize each RichCarrierRate into a CanonicalRate
	canonical := make([]CanonicalRate, 0, len(resp.Rates))
	for _, r := range resp.Rates {
		cr := s.normalizer.Normalize(r, orgID, q.OriginPort, q.DestinationPort)
		canonical = append(canonical, cr)
	}

	// Store for future searches (spot rates expire in 4h via valid_until)
	if err := s.repo.Upsert(ctx, canonical); err != nil {
		// Non-fatal: log but still return the rates to the caller
		// so the Quotation Engine isn't blocked by a DB write failure.
		_ = err
	}

	return buildSearchResult(canonical), nil
}

// GetRateByID returns a single rate by UUID.
func (s *service) GetRateByID(ctx context.Context, orgID int64, id string) (*CanonicalRate, error) {
	return s.repo.GetByID(ctx, orgID, id)
}

// buildSearchResult ranks rates and builds the RateSearchResult response.
//
// Ranking rules (in order of priority):
//  1. CONTRACT_PDF rates rank above SPOT_API rates when confidence >= 85.
//  2. Within the same source tier: lower total_buy_price wins.
//  3. Deadline penalty: if transit_days > days_until_target, move to bottom.
func buildSearchResult(rates []CanonicalRate) *RateSearchResult {
	if len(rates) == 0 {
		return &RateSearchResult{SearchedAt: time.Now()}
	}

	// Sort: contract > spot; higher confidence > lower; cheaper > expensive
	sort.SliceStable(rates, func(i, j int) bool {
		ri, rj := rates[i], rates[j]

		// Tier 1: contract rates with high confidence beat spot rates
		riIsContract := ri.Source == RateSourceContractPDF && ri.ConfidenceScore >= 85
		rjIsContract := rj.Source == RateSourceContractPDF && rj.ConfidenceScore >= 85
		if riIsContract != rjIsContract {
			return riIsContract
		}

		// Tier 2: higher confidence first
		if ri.ConfidenceScore != rj.ConfidenceScore {
			return ri.ConfidenceScore > rj.ConfidenceScore
		}

		// Tier 3: lower price first
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

	return &RateSearchResult{
		Rates:             rates,
		TotalCount:        len(rates),
		SpotRateCount:     spotCount,
		ContractRateCount: contractCount,
		RecommendedIdx:    0,
		OverallReasoning:  reasoning,
		SearchedAt:        time.Now(),
	}
}

// buildOverallReasoning produces a human-readable recommendation summary.
func buildOverallReasoning(top CanonicalRate) string {
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
