package rates

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/freel/backend/internal/rates/spec"
	"github.com/jmoiron/sqlx"
)

// Datalayer handles all database interactions for rate_entries and rates tables.
type Datalayer interface {
	// Canonical rate entries
	Upsert(ctx context.Context, rates []spec.CanonicalRate) error
	Search(ctx context.Context, q spec.RateQuery) ([]spec.CanonicalRate, error)
	GetByID(ctx context.Context, orgID int64, id string) (*spec.CanonicalRate, error)

	// Rate Management (Task 19.1)
	CreateRate(ctx context.Context, rate *spec.Rate) error
	GetRateByID(ctx context.Context, orgID, rateID int64) (*spec.Rate, error)
	GetRateByReference(ctx context.Context, orgID int64, ref string) (*spec.Rate, error)
	ListRates(ctx context.Context, filters *spec.ListRatesRequest) ([]*spec.RateListItem, int, error)
	UpdateRate(ctx context.Context, rate *spec.Rate) error
	ArchiveRate(ctx context.Context, orgID, rateID int64, archivedBy string) error
	GetRateSummaryKPIs(ctx context.Context, orgID int64) (*spec.RateSummaryKPIs, error)

	// Rate Charges & Pricing (Task 19.2)
	CreateRateCharge(ctx context.Context, charge *spec.RateChargeItem) error
	GetRateCharges(ctx context.Context, orgID, rateID int64) ([]spec.RateChargeItem, error)
	GetRateChargeByID(ctx context.Context, orgID, rateID, chargeID int64) (*spec.RateChargeItem, error)
	UpdateRateCharge(ctx context.Context, charge *spec.RateChargeItem) error
	DeleteRateCharge(ctx context.Context, orgID, rateID, chargeID int64) error
	ReorderRateCharges(ctx context.Context, orgID, rateID int64, chargeIDs []int64) error

	// Carrier Rate Contracts & Versions (Task 19.3)
	CreateRateContract(ctx context.Context, contract *spec.RateContract) error
	GetRateContractByID(ctx context.Context, orgID, contractID int64) (*spec.RateContract, error)
	ListRateContracts(ctx context.Context, filters *spec.ListRateContractsRequest) ([]*spec.RateContractListItem, int, error)
	UpdateRateContract(ctx context.Context, contract *spec.RateContract) error
	ArchiveRateContract(ctx context.Context, orgID, contractID int64, archivedBy string) error
	GetRateContractSummary(ctx context.Context, orgID int64) (*spec.RateContractSummary, error)
	GetRatesByContract(ctx context.Context, orgID, contractID int64) ([]*spec.RateListItem, error)

	// Version Lineage & Downstream Protection
	IsRateReferencedDownstream(ctx context.Context, orgID, rateID int64) (bool, error)
	CreateRateVersionRecord(ctx context.Context, newRate *spec.Rate, history *spec.RateVersionHistory) error
	GetRateVersionHistory(ctx context.Context, orgID, rateID int64) ([]spec.RateVersionHistory, error)
	GetRateVersionChain(ctx context.Context, orgID, rateID int64) ([]spec.RateVersionChainItem, error)
	MarkRateSuperseded(ctx context.Context, orgID, previousRateID, supersedingRateID int64) error

	// Task 19.4: Spot Rate Requests, Responses & Comparison
	CreateSpotRateRequest(ctx context.Context, req *spec.SpotRateRequest) error
	GetSpotRateRequestByID(ctx context.Context, orgID, requestID int64) (*spec.SpotRateRequest, error)
	ListSpotRateRequests(ctx context.Context, filters *spec.ListSpotRateRequestsRequest) ([]*spec.SpotRateRequestListItem, int, error)
	UpdateSpotRateRequest(ctx context.Context, req *spec.SpotRateRequest) error
	CancelSpotRateRequest(ctx context.Context, orgID, requestID int64) error

	CreateSpotRateResponse(ctx context.Context, resp *spec.SpotRateResponse, charges []spec.SpotRateResponseCharge) error
	GetSpotRateResponseByID(ctx context.Context, orgID, responseID int64) (*spec.SpotRateResponse, error)
	GetSpotRateResponsesByRequestID(ctx context.Context, orgID, requestID int64) ([]*spec.SpotRateResponse, error)
	UpdateSpotRateResponse(ctx context.Context, resp *spec.SpotRateResponse, charges []spec.SpotRateResponseCharge) error
	SelectPreferredSpotRateResponse(ctx context.Context, orgID, requestID, responseID int64) error

	GetSpotRateRequestSummary(ctx context.Context, orgID int64) (*spec.SpotRateRequestSummary, error)

	// Task 19.6: Rate Lifecycle Intelligence & Attention
	CreateRateLifecycleEvent(ctx context.Context, event *RateLifecycleEvent) error
	GetRateLifecycleEvents(ctx context.Context, orgID int64, limit int) ([]*RateLifecycleEvent, error)
	GetRateLifecycleSummary(ctx context.Context, orgID int64) (*RateLifecycleSummary, error)
	GetRatesRequiringAttention(ctx context.Context, orgID int64) ([]*RateAttentionItem, error)
	GetContractsRequiringAttention(ctx context.Context, orgID int64) ([]*ContractAttentionItem, error)
	GetAllActiveAndExpiringRatesForEvaluation(ctx context.Context, orgID int64) ([]*spec.Rate, error)
	GetAllActiveAndExpiringContractsForEvaluation(ctx context.Context, orgID int64) ([]*spec.RateContract, error)
	UpdateRateStatusDirect(ctx context.Context, orgID, rateID int64, newStatus string) error
	UpdateContractStatusDirect(ctx context.Context, orgID, contractID int64, newStatus, newRenewalStatus string) error

	// Task 19.7: Rate Analytics & Procurement Intelligence
	GetRateAnalyticsOverview(ctx context.Context, orgID int64) (*RateAnalyticsOverview, error)
	GetRateAnalyticsTrends(ctx context.Context, orgID int64, days int) ([]RateTrendDataPoint, error)
	GetCarrierRatePerformance(ctx context.Context, orgID int64) ([]CarrierRatePerformance, error)
	GetLaneRatePerformance(ctx context.Context, orgID int64) ([]LaneRatePerformance, error)
	GetRateLifecycleAnalytics(ctx context.Context, orgID int64) (*RateLifecycleAnalytics, error)
	GetSpotSourcingPerformance(ctx context.Context, orgID int64) (*SpotSourcingPerformance, error)
	GetCommercialRiskExposure(ctx context.Context, orgID int64) (int, error)
}

type dataLayer struct {
	db *sqlx.DB
}

func NewDataLayer(db *sqlx.DB) Datalayer {
	return &dataLayer{db: db}
}

func (d *dataLayer) Upsert(ctx context.Context, rates []spec.CanonicalRate) error {
	if len(rates) == 0 {
		return nil
	}

	const query = `
INSERT INTO rate_entries (
	id, org_id, source, source_ref, contract_doc_id,
	origin_port, destination_port, via_port, service_code,
	carrier_scac, carrier_name, vessel_name, equipment_type,
	ocean_freight, origin_charges, destination_charges, surcharges,
	total_buy_price, currency_original, exchange_rate_used,
	included_charges, excluded_charges,
	free_days_origin, free_days_destination, transit_days, incoterms,
	commodity_restrictions, routing_conditions,
	valid_from, valid_until,
	confidence_score, extraction_status, extracted_by,
	nautical_miles, co2_per_teu, created_at, updated_at
) VALUES (
	:id, :org_id, :source, :source_ref, :contract_doc_id,
	:origin_port, :destination_port, :via_port, :service_code,
	:carrier_scac, :carrier_name, :vessel_name, :equipment_type,
	:ocean_freight, :origin_charges, :destination_charges, :surcharges,
	:total_buy_price, :currency_original, :exchange_rate_used,
	:included_charges, :excluded_charges,
	:free_days_origin, :free_days_destination, :transit_days, :incoterms,
	:commodity_restrictions, :routing_conditions,
	:valid_from, :valid_until,
	:confidence_score, :extraction_status, :extracted_by,
	:nautical_miles, :co2_per_teu, NOW(), NOW()
)
ON DUPLICATE KEY UPDATE
	ocean_freight          = VALUES(ocean_freight),
	origin_charges         = VALUES(origin_charges),
	destination_charges    = VALUES(destination_charges),
	surcharges             = VALUES(surcharges),
	total_buy_price        = VALUES(total_buy_price),
	free_days_origin       = VALUES(free_days_origin),
	free_days_destination  = VALUES(free_days_destination),
	transit_days           = VALUES(transit_days),
	valid_from             = VALUES(valid_from),
	valid_until            = VALUES(valid_until),
	confidence_score       = VALUES(confidence_score),
	extraction_status      = VALUES(extraction_status),
	extracted_by           = VALUES(extracted_by),
	updated_at             = NOW()`

	type dbRow struct {
		ID                     string   `db:"id"`
		OrgID                  int64    `db:"org_id"`
		Source                 string   `db:"source"`
		SourceRef              string   `db:"source_ref"`
		ContractDocID          *string  `db:"contract_doc_id"`
		OriginPort             string   `db:"origin_port"`
		DestinationPort        string   `db:"destination_port"`
		ViaPort                string   `db:"via_port"`
		ServiceCode            string   `db:"service_code"`
		CarrierSCAC            string   `db:"carrier_scac"`
		CarrierName            string   `db:"carrier_name"`
		VesselName             string   `db:"vessel_name"`
		EquipmentType          string   `db:"equipment_type"`
		OceanFreight           float64  `db:"ocean_freight"`
		OriginCharges          float64  `db:"origin_charges"`
		DestinationCharges     float64  `db:"destination_charges"`
		Surcharges             []byte   `db:"surcharges"`
		TotalBuyPrice          float64  `db:"total_buy_price"`
		CurrencyOriginal       string   `db:"currency_original"`
		ExchangeRateUsed       float64  `db:"exchange_rate_used"`
		IncludedCharges        []byte   `db:"included_charges"`
		ExcludedCharges        []byte   `db:"excluded_charges"`
		FreeDaysOrigin         int      `db:"free_days_origin"`
		FreeDaysDestination    int      `db:"free_days_destination"`
		TransitDays            *int     `db:"transit_days"`
		Incoterms              string   `db:"incoterms"`
		CommodityRestrictions  []byte   `db:"commodity_restrictions"`
		RoutingConditions      string   `db:"routing_conditions"`
		ValidFrom              time.Time `db:"valid_from"`
		ValidUntil             time.Time `db:"valid_until"`
		ConfidenceScore        int      `db:"confidence_score"`
		ExtractionStatus       string   `db:"extraction_status"`
		ExtractedBy            string   `db:"extracted_by"`
		NauticalMiles          int      `db:"nautical_miles"`
		CO2PerTEU              float64  `db:"co2_per_teu"`
	}

	rows := make([]dbRow, 0, len(rates))
	for _, r := range rates {
		surchargesJSON, err := json.Marshal(r.Surcharges)
		if err != nil {
			surchargesJSON = []byte("[]")
		}
		incJSON, _ := json.Marshal(r.IncludedCharges)
		excJSON, _ := json.Marshal(r.ExcludedCharges)
		commJSON, _ := json.Marshal(r.CommodityRestrictions)

		rows = append(rows, dbRow{
			ID:                    r.ID,
			OrgID:                 r.OrgID,
			Source:                r.Source,
			SourceRef:             r.SourceRef,
			ContractDocID:         r.ContractDocID,
			OriginPort:            r.OriginPort,
			DestinationPort:       r.DestinationPort,
			ViaPort:               r.ViaPort,
			ServiceCode:           r.ServiceCode,
			CarrierSCAC:           r.CarrierSCAC,
			CarrierName:           r.CarrierName,
			VesselName:            r.VesselName,
			EquipmentType:         r.EquipmentType,
			OceanFreight:          r.OceanFreight,
			OriginCharges:         r.OriginCharges,
			DestinationCharges:    r.DestinationCharges,
			Surcharges:            surchargesJSON,
			TotalBuyPrice:         r.TotalBuyPrice,
			CurrencyOriginal:      r.CurrencyOriginal,
			ExchangeRateUsed:      r.ExchangeRateUsed,
			IncludedCharges:       incJSON,
			ExcludedCharges:       excJSON,
			FreeDaysOrigin:        r.FreeDaysOrigin,
			FreeDaysDestination:   r.FreeDaysDestination,
			TransitDays:           r.TransitDays,
			Incoterms:             r.Incoterms,
			CommodityRestrictions: commJSON,
			RoutingConditions:     r.RoutingConditions,
			ValidFrom:             r.ValidFrom,
			ValidUntil:            r.ValidUntil,
			ConfidenceScore:       r.ConfidenceScore,
			ExtractionStatus:      r.ExtractionStatus,
			ExtractedBy:           r.ExtractedBy,
			NauticalMiles:         r.NauticalMiles,
			CO2PerTEU:             r.CO2PerTEU,
		})
	}

	for _, row := range rows {
		if _, err := d.db.NamedExecContext(ctx, query, row); err != nil {
			return fmt.Errorf("rate_entries upsert (id=%s): %w", row.ID, err)
		}
	}
	return nil
}

func (d *dataLayer) Search(ctx context.Context, q spec.RateQuery) ([]spec.CanonicalRate, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "org_id = ?")
	args = append(args, q.OrgID)

	conditions = append(conditions, "extraction_status = ?")
	args = append(args, ExtractionStatusConfirmed)

	if q.OriginPort != "" {
		conditions = append(conditions, "origin_port = ?")
		args = append(args, q.OriginPort)
	}
	if q.DestinationPort != "" {
		conditions = append(conditions, "destination_port = ?")
		args = append(args, q.DestinationPort)
	}
	if q.EquipmentType != "" {
		conditions = append(conditions, "equipment_type = ?")
		args = append(args, q.EquipmentType)
	}
	if q.TargetDate != nil {
		conditions = append(conditions, "valid_from <= ? AND valid_until >= ?")
		args = append(args, *q.TargetDate, *q.TargetDate)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + joinWithAnd(conditions)
	}

	limit := q.MaxResults
	if limit <= 0 {
		limit = 20
	}

	query := fmt.Sprintf(`
SELECT
	id, org_id, source, source_ref, contract_doc_id,
	origin_port, destination_port, via_port, service_code,
	carrier_scac, carrier_name, vessel_name, equipment_type,
	ocean_freight, origin_charges, destination_charges, surcharges,
	total_buy_price, currency_original, exchange_rate_used,
	included_charges, excluded_charges,
	free_days_origin, free_days_destination, transit_days, incoterms,
	commodity_restrictions, routing_conditions,
	valid_from, valid_until,
	confidence_score, extraction_status, extracted_by,
	nautical_miles, co2_per_teu, created_at, updated_at
FROM rate_entries
%s
ORDER BY total_buy_price ASC
LIMIT ?`, whereClause)

	args = append(args, limit)

	type rawRow struct {
		ID                     string    `db:"id"`
		OrgID                  int64     `db:"org_id"`
		Source                 string    `db:"source"`
		SourceRef              string    `db:"source_ref"`
		ContractDocID          *string   `db:"contract_doc_id"`
		OriginPort             string    `db:"origin_port"`
		DestinationPort        string    `db:"destination_port"`
		ViaPort                string    `db:"via_port"`
		ServiceCode            string    `db:"service_code"`
		CarrierSCAC            string    `db:"carrier_scac"`
		CarrierName            string    `db:"carrier_name"`
		VesselName             string    `db:"vessel_name"`
		EquipmentType          string    `db:"equipment_type"`
		OceanFreight           float64   `db:"ocean_freight"`
		OriginCharges          float64   `db:"origin_charges"`
		DestinationCharges     float64   `db:"destination_charges"`
		SurchargesRaw          []byte    `db:"surcharges"`
		TotalBuyPrice          float64   `db:"total_buy_price"`
		CurrencyOriginal       string    `db:"currency_original"`
		ExchangeRateUsed       float64   `db:"exchange_rate_used"`
		IncludedChargesRaw     []byte    `db:"included_charges"`
		ExcludedChargesRaw     []byte    `db:"excluded_charges"`
		FreeDaysOrigin         int       `db:"free_days_origin"`
		FreeDaysDestination    int       `db:"free_days_destination"`
		TransitDays            *int      `db:"transit_days"`
		Incoterms              string    `db:"incoterms"`
		CommodityRestrictionsRaw []byte  `db:"commodity_restrictions"`
		RoutingConditions      string    `db:"routing_conditions"`
		ValidFrom              time.Time `db:"valid_from"`
		ValidUntil             time.Time `db:"valid_until"`
		ConfidenceScore        int       `db:"confidence_score"`
		ExtractionStatus       string    `db:"extraction_status"`
		ExtractedBy            string    `db:"extracted_by"`
		NauticalMiles          int       `db:"nautical_miles"`
		CO2PerTEU              float64   `db:"co2_per_teu"`
		CreatedAt              time.Time `db:"created_at"`
		UpdatedAt              time.Time `db:"updated_at"`
	}

	var rows []rawRow
	if err := d.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("rate_entries search: %w", err)
	}

	results := make([]spec.CanonicalRate, 0, len(rows))
	for _, row := range rows {
		var surcharges []spec.Surcharge
		if len(row.SurchargesRaw) > 0 {
			_ = json.Unmarshal(row.SurchargesRaw, &surcharges)
		}
		var inc, exc, comm []string
		if len(row.IncludedChargesRaw) > 0 {
			_ = json.Unmarshal(row.IncludedChargesRaw, &inc)
		}
		if len(row.ExcludedChargesRaw) > 0 {
			_ = json.Unmarshal(row.ExcludedChargesRaw, &exc)
		}
		if len(row.CommodityRestrictionsRaw) > 0 {
			_ = json.Unmarshal(row.CommodityRestrictionsRaw, &comm)
		}

		results = append(results, spec.CanonicalRate{
			ID:                    row.ID,
			OrgID:                 row.OrgID,
			Source:                row.Source,
			SourceRef:             row.SourceRef,
			ContractDocID:         row.ContractDocID,
			OriginPort:            row.OriginPort,
			DestinationPort:       row.DestinationPort,
			ViaPort:               row.ViaPort,
			ServiceCode:           row.ServiceCode,
			CarrierSCAC:           row.CarrierSCAC,
			CarrierName:           row.CarrierName,
			VesselName:            row.VesselName,
			EquipmentType:         row.EquipmentType,
			OceanFreight:          row.OceanFreight,
			OriginCharges:         row.OriginCharges,
			DestinationCharges:    row.DestinationCharges,
			Surcharges:            surcharges,
			TotalBuyPrice:         row.TotalBuyPrice,
			CurrencyOriginal:      row.CurrencyOriginal,
			ExchangeRateUsed:      row.ExchangeRateUsed,
			IncludedCharges:       inc,
			ExcludedCharges:       exc,
			FreeDaysOrigin:        row.FreeDaysOrigin,
			FreeDaysDestination:   row.FreeDaysDestination,
			TransitDays:           row.TransitDays,
			Incoterms:             row.Incoterms,
			CommodityRestrictions: comm,
			RoutingConditions:     row.RoutingConditions,
			ValidFrom:             row.ValidFrom,
			ValidUntil:            row.ValidUntil,
			ConfidenceScore:       row.ConfidenceScore,
			ExtractionStatus:      row.ExtractionStatus,
			ExtractedBy:           row.ExtractedBy,
			NauticalMiles:         row.NauticalMiles,
			CO2PerTEU:             row.CO2PerTEU,
			CreatedAt:             row.CreatedAt,
			UpdatedAt:             row.UpdatedAt,
		})
	}
	return results, nil
}

func (d *dataLayer) GetByID(ctx context.Context, orgID int64, id string) (*spec.CanonicalRate, error) {
	const query = `
SELECT
	id, org_id, source, source_ref, contract_doc_id,
	origin_port, destination_port, via_port, service_code,
	carrier_scac, carrier_name, vessel_name, equipment_type,
	ocean_freight, origin_charges, destination_charges, surcharges,
	total_buy_price, currency_original, exchange_rate_used,
	included_charges, excluded_charges,
	free_days_origin, free_days_destination, transit_days, incoterms,
	commodity_restrictions, routing_conditions,
	valid_from, valid_until,
	confidence_score, extraction_status, extracted_by,
	nautical_miles, co2_per_teu, created_at, updated_at
FROM rate_entries
WHERE id = ? AND org_id = ?`

	var row struct {
		ID                     string    `db:"id"`
		OrgID                  int64     `db:"org_id"`
		Source                 string    `db:"source"`
		SourceRef              string    `db:"source_ref"`
		ContractDocID          *string   `db:"contract_doc_id"`
		OriginPort             string    `db:"origin_port"`
		DestinationPort        string    `db:"destination_port"`
		ViaPort                string    `db:"via_port"`
		ServiceCode            string    `db:"service_code"`
		CarrierSCAC            string    `db:"carrier_scac"`
		CarrierName            string    `db:"carrier_name"`
		VesselName             string    `db:"vessel_name"`
		EquipmentType          string    `db:"equipment_type"`
		OceanFreight           float64   `db:"ocean_freight"`
		OriginCharges          float64   `db:"origin_charges"`
		DestinationCharges     float64   `db:"destination_charges"`
		SurchargesRaw          []byte    `db:"surcharges"`
		TotalBuyPrice          float64   `db:"total_buy_price"`
		CurrencyOriginal       string    `db:"currency_original"`
		ExchangeRateUsed       float64   `db:"exchange_rate_used"`
		IncludedChargesRaw     []byte    `db:"included_charges"`
		ExcludedChargesRaw     []byte    `db:"excluded_charges"`
		FreeDaysOrigin         int       `db:"free_days_origin"`
		FreeDaysDestination    int       `db:"free_days_destination"`
		TransitDays            *int      `db:"transit_days"`
		Incoterms              string    `db:"incoterms"`
		CommodityRestrictionsRaw []byte  `db:"commodity_restrictions"`
		RoutingConditions      string    `db:"routing_conditions"`
		ValidFrom              time.Time `db:"valid_from"`
		ValidUntil             time.Time `db:"valid_until"`
		ConfidenceScore        int       `db:"confidence_score"`
		ExtractionStatus       string    `db:"extraction_status"`
		ExtractedBy            string    `db:"extracted_by"`
		NauticalMiles          int       `db:"nautical_miles"`
		CO2PerTEU              float64   `db:"co2_per_teu"`
	}

	if err := d.db.GetContext(ctx, &row, query, id, orgID); err != nil {
		return nil, fmt.Errorf("rate_entries get by id: %w", err)
	}

	var surcharges []spec.Surcharge
	if len(row.SurchargesRaw) > 0 {
		_ = json.Unmarshal(row.SurchargesRaw, &surcharges)
	}
	var inc, exc, comm []string
	if len(row.IncludedChargesRaw) > 0 {
		_ = json.Unmarshal(row.IncludedChargesRaw, &inc)
	}
	if len(row.ExcludedChargesRaw) > 0 {
		_ = json.Unmarshal(row.ExcludedChargesRaw, &exc)
	}
	if len(row.CommodityRestrictionsRaw) > 0 {
		_ = json.Unmarshal(row.CommodityRestrictionsRaw, &comm)
	}

	return &spec.CanonicalRate{
		ID:                    row.ID,
		OrgID:                 row.OrgID,
		Source:                row.Source,
		SourceRef:             row.SourceRef,
		ContractDocID:         row.ContractDocID,
		OriginPort:            row.OriginPort,
		DestinationPort:       row.DestinationPort,
		ViaPort:               row.ViaPort,
		ServiceCode:           row.ServiceCode,
		CarrierSCAC:           row.CarrierSCAC,
		CarrierName:           row.CarrierName,
		VesselName:            row.VesselName,
		EquipmentType:         row.EquipmentType,
		OceanFreight:          row.OceanFreight,
		OriginCharges:         row.OriginCharges,
		DestinationCharges:    row.DestinationCharges,
		Surcharges:            surcharges,
		TotalBuyPrice:         row.TotalBuyPrice,
		CurrencyOriginal:      row.CurrencyOriginal,
		ExchangeRateUsed:      row.ExchangeRateUsed,
		IncludedCharges:       inc,
		ExcludedCharges:       exc,
		FreeDaysOrigin:        row.FreeDaysOrigin,
		FreeDaysDestination:   row.FreeDaysDestination,
		TransitDays:           row.TransitDays,
		Incoterms:             row.Incoterms,
		CommodityRestrictions: comm,
		RoutingConditions:     row.RoutingConditions,
		ConfidenceScore:       row.ConfidenceScore,
		ExtractionStatus:      row.ExtractionStatus,
		ExtractedBy:           row.ExtractedBy,
		NauticalMiles:         row.NauticalMiles,
		CO2PerTEU:             row.CO2PerTEU,
	}, nil
}

// ═══════════════════════════════════════════════════════════════════════════════
// Rate Management Foundation Implementations (Task 19.1)
// ═══════════════════════════════════════════════════════════════════════════════

func (d *dataLayer) CreateRate(ctx context.Context, rate *spec.Rate) error {
	const query = `
INSERT INTO rates (
	org_id, rate_reference, carrier_name, carrier_code, service_provider,
	rate_type, transport_mode, service_type, equipment_type,
	origin_port, origin_code, destination_port, destination_code,
	currency, base_amount, effective_date, expiry_date,
	status, carrier_reference, contract_reference, notes,
	created_by, updated_by, created_at, updated_at
) VALUES (
	:org_id, :rate_reference, :carrier_name, :carrier_code, :service_provider,
	:rate_type, :transport_mode, :service_type, :equipment_type,
	:origin_port, :origin_code, :destination_port, :destination_code,
	:currency, :base_amount, :effective_date, :expiry_date,
	:status, :carrier_reference, :contract_reference, :notes,
	:created_by, :updated_by, NOW(), NOW()
)`
	res, err := d.db.NamedExecContext(ctx, query, rate)
	if err != nil {
		return fmt.Errorf("rates create: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		rate.ID = id
	}
	return nil
}

func (d *dataLayer) GetRateByID(ctx context.Context, orgID, rateID int64) (*spec.Rate, error) {
	const query = `
SELECT 
	id, org_id, rate_reference, carrier_name, carrier_code, service_provider,
	rate_type, transport_mode, service_type, equipment_type,
	origin_port, origin_code, destination_port, destination_code,
	currency, base_amount, effective_date, expiry_date,
	status, carrier_reference, contract_reference, notes,
	created_by, updated_by, created_at, updated_at
FROM rates
WHERE id = ? AND org_id = ?`

	var rate spec.Rate
	if err := d.db.GetContext(ctx, &rate, query, rateID, orgID); err != nil {
		return nil, fmt.Errorf("rates get by id: %w", err)
	}
	return &rate, nil
}

func (d *dataLayer) GetRateByReference(ctx context.Context, orgID int64, ref string) (*spec.Rate, error) {
	const query = `
SELECT 
	id, org_id, rate_reference, carrier_name, carrier_code, service_provider,
	rate_type, transport_mode, service_type, equipment_type,
	origin_port, origin_code, destination_port, destination_code,
	currency, base_amount, effective_date, expiry_date,
	status, carrier_reference, contract_reference, notes,
	created_by, updated_by, created_at, updated_at
FROM rates
WHERE rate_reference = ? AND org_id = ?`

	var rate spec.Rate
	if err := d.db.GetContext(ctx, &rate, query, ref, orgID); err != nil {
		return nil, fmt.Errorf("rates get by ref: %w", err)
	}
	return &rate, nil
}

func (d *dataLayer) ListRates(ctx context.Context, filters *spec.ListRatesRequest) ([]*spec.RateListItem, int, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "org_id = ?")
	args = append(args, filters.OrgID)

	if filters.Status != "" && filters.Status != "ALL" {
		conditions = append(conditions, "status = ?")
		args = append(args, filters.Status)
	}
	if filters.RateType != "" && filters.RateType != "ALL" {
		conditions = append(conditions, "rate_type = ?")
		args = append(args, filters.RateType)
	}
	if filters.TransportMode != "" && filters.TransportMode != "ALL" {
		conditions = append(conditions, "transport_mode = ?")
		args = append(args, filters.TransportMode)
	}
	if filters.ServiceType != "" && filters.ServiceType != "ALL" {
		conditions = append(conditions, "service_type = ?")
		args = append(args, filters.ServiceType)
	}
	if filters.EquipmentType != "" && filters.EquipmentType != "ALL" {
		conditions = append(conditions, "equipment_type = ?")
		args = append(args, filters.EquipmentType)
	}
	if filters.CarrierName != "" && filters.CarrierName != "ALL" {
		conditions = append(conditions, "carrier_name = ?")
		args = append(args, filters.CarrierName)
	}
	if filters.Origin != "" {
		conditions = append(conditions, "(origin_port LIKE ? OR origin_code LIKE ?)")
		args = append(args, "%"+filters.Origin+"%", "%"+filters.Origin+"%")
	}
	if filters.Destination != "" {
		conditions = append(conditions, "(destination_port LIKE ? OR destination_code LIKE ?)")
		args = append(args, "%"+filters.Destination+"%", "%"+filters.Destination+"%")
	}
	if filters.ValidDate != nil {
		conditions = append(conditions, "effective_date <= ? AND expiry_date >= ?")
		args = append(args, *filters.ValidDate, *filters.ValidDate)
	}
	if filters.Search != "" {
		conditions = append(conditions, "(rate_reference LIKE ? OR carrier_name LIKE ? OR origin_port LIKE ? OR destination_port LIKE ? OR carrier_reference LIKE ? OR contract_reference LIKE ?)")
		pat := "%" + filters.Search + "%"
		args = append(args, pat, pat, pat, pat, pat, pat)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + joinWithAnd(conditions)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM rates %s", whereClause)
	var totalCount int
	if err := d.db.GetContext(ctx, &totalCount, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("rates list count: %w", err)
	}

	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	page := filters.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	orderCol := "created_at"
	if filters.SortBy != "" {
		switch filters.SortBy {
		case "rate_reference", "carrier_name", "base_amount", "effective_date", "expiry_date", "status":
			orderCol = filters.SortBy
		}
	}
	orderDir := "DESC"
	if filters.SortOrder == "ASC" || filters.SortOrder == "asc" {
		orderDir = "ASC"
	}

	dataQuery := fmt.Sprintf(`
SELECT 
	id, rate_reference, carrier_name, carrier_code,
	rate_type, transport_mode, service_type, equipment_type,
	origin_port, origin_code, destination_port, destination_code,
	currency, base_amount, 
	DATE_FORMAT(effective_date, '%%Y-%%m-%%d') AS effective_date,
	DATE_FORMAT(expiry_date, '%%Y-%%m-%%d') AS expiry_date,
	status, carrier_reference, contract_reference,
	COALESCE(version_number, 1) AS version_number,
	COALESCE(version_status, 'CURRENT') AS version_status,
	supersedes_rate_id, superseded_by_rate_id,
	updated_at
FROM rates
%s
ORDER BY %s %s
LIMIT ? OFFSET ?`, whereClause, orderCol, orderDir)

	dataArgs := append(args, limit, offset)
	var rows []*spec.RateListItem
	if err := d.db.SelectContext(ctx, &rows, dataQuery, dataArgs...); err != nil {
		return nil, 0, fmt.Errorf("rates list select: %w", err)
	}

	now := time.Now().UTC()
	for _, row := range rows {
		eff, _ := time.Parse("2006-01-02", row.EffectiveDate)
		exp, _ := time.Parse("2006-01-02", row.ExpiryDate)
		newStatus, days := CalculateRateValidity(eff, exp, row.Status, now)
		row.Status = newStatus
		row.DaysUntilExpiry = days
	}

	return rows, totalCount, nil
}

func (d *dataLayer) UpdateRate(ctx context.Context, rate *spec.Rate) error {
	const query = `
UPDATE rates SET
	carrier_name = :carrier_name,
	carrier_code = :carrier_code,
	service_provider = :service_provider,
	rate_type = :rate_type,
	transport_mode = :transport_mode,
	service_type = :service_type,
	equipment_type = :equipment_type,
	origin_port = :origin_port,
	origin_code = :origin_code,
	destination_port = :destination_port,
	destination_code = :destination_code,
	currency = :currency,
	base_amount = :base_amount,
	effective_date = :effective_date,
	expiry_date = :expiry_date,
	status = :status,
	carrier_reference = :carrier_reference,
	contract_reference = :contract_reference,
	notes = :notes,
	updated_by = :updated_by,
	updated_at = NOW()
WHERE id = :id AND org_id = :org_id`

	_, err := d.db.NamedExecContext(ctx, query, rate)
	if err != nil {
		return fmt.Errorf("rates update: %w", err)
	}
	return nil
}

func (d *dataLayer) ArchiveRate(ctx context.Context, orgID, rateID int64, archivedBy string) error {
	const query = `
UPDATE rates SET
	status = 'ARCHIVED',
	updated_by = ?,
	updated_at = NOW()
WHERE id = ? AND org_id = ?`

	_, err := d.db.ExecContext(ctx, query, archivedBy, rateID, orgID)
	if err != nil {
		return fmt.Errorf("rates archive: %w", err)
	}
	return nil
}

func (d *dataLayer) GetRateSummaryKPIs(ctx context.Context, orgID int64) (*spec.RateSummaryKPIs, error) {
	kpis := &spec.RateSummaryKPIs{
		TopCarriers:   []spec.CarrierCoverageSummary{},
		RecentUpdates: []spec.RecentRateUpdate{},
	}

	type statusCountRow struct {
		EffectiveDate time.Time `db:"effective_date"`
		ExpiryDate    time.Time `db:"expiry_date"`
		Status        string    `db:"status"`
	}

	var allRates []statusCountRow
	ratesQuery := `SELECT effective_date, expiry_date, status FROM rates WHERE org_id = ? AND status != 'ARCHIVED'`
	if err := d.db.SelectContext(ctx, &allRates, ratesQuery, orgID); err != nil {
		return nil, fmt.Errorf("rates summary status counts: %w", err)
	}

	now := time.Now().UTC()
	kpis.TotalRates = len(allRates)

	for _, rRow := range allRates {
		st, _ := CalculateRateValidity(rRow.EffectiveDate, rRow.ExpiryDate, rRow.Status, now)
		switch st {
		case RateStatusActive:
			kpis.ActiveRates++
		case RateStatusExpiringSoon:
			kpis.ExpiringSoonRates++
		case RateStatusExpired:
			kpis.ExpiredRates++
		}
	}

	if kpis.TotalRates > 0 {
		kpis.ActivePct = float64(kpis.ActiveRates) / float64(kpis.TotalRates) * 100
		kpis.ExpiringSoonPct = float64(kpis.ExpiringSoonRates) / float64(kpis.TotalRates) * 100
		kpis.ExpiredPct = float64(kpis.ExpiredRates) / float64(kpis.TotalRates) * 100
	}

	laneQuery := `SELECT COUNT(DISTINCT CONCAT(origin_port, '->', destination_port)) FROM rates WHERE org_id = ? AND status != 'ARCHIVED'`
	_ = d.db.GetContext(ctx, &kpis.LanesCovered, laneQuery, orgID)

	topCarrierQuery := `
SELECT 
	carrier_name,
	carrier_code,
	COUNT(DISTINCT CONCAT(origin_port, '->', destination_port)) AS lane_count,
	COUNT(*) AS rate_count
FROM rates
WHERE org_id = ? AND status != 'ARCHIVED'
GROUP BY carrier_name, carrier_code
ORDER BY lane_count DESC, rate_count DESC
LIMIT 5`

	var topCarriers []spec.CarrierCoverageSummary
	if err := d.db.SelectContext(ctx, &topCarriers, topCarrierQuery, orgID); err == nil {
		totalLanes := kpis.LanesCovered
		if totalLanes == 0 {
			totalLanes = 1
		}
		for i := range topCarriers {
			topCarriers[i].SharePct = (float64(topCarriers[i].LaneCount) / float64(totalLanes)) * 100
			if topCarriers[i].SharePct > 100 {
				topCarriers[i].SharePct = 100
			}
		}
		kpis.TopCarriers = topCarriers
	}

	recentQuery := `
SELECT 
	id, rate_reference, carrier_name, origin_port, destination_port,
	base_amount, currency, status, updated_at
FROM rates
WHERE org_id = ?
ORDER BY updated_at DESC
LIMIT 4`

	var recentUpdates []spec.RecentRateUpdate
	if err := d.db.SelectContext(ctx, &recentUpdates, recentQuery, orgID); err == nil {
		kpis.RecentUpdates = recentUpdates
	}

	return kpis, nil
}

// ── Task 19.2: Rate Charges Data Layer Implementation ─────────────────────────

func (d *dataLayer) CreateRateCharge(ctx context.Context, charge *spec.RateChargeItem) error {
	query := `
INSERT INTO rate_charge_items (
	org_id, rate_id, charge_category, charge_code, charge_name,
	calculation_basis, quantity, unit_price, currency,
	minimum_amount, maximum_amount, included_in_base_rate,
	display_order, notes, created_at, updated_at
) VALUES (
	:org_id, :rate_id, :charge_category, :charge_code, :charge_name,
	:calculation_basis, :quantity, :unit_price, :currency,
	:minimum_amount, :maximum_amount, :included_in_base_rate,
	:display_order, :notes, NOW(), NOW()
)`

	res, err := d.db.NamedExecContext(ctx, query, charge)
	if err != nil {
		return fmt.Errorf("rates.CreateRateCharge: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		charge.ID = id
	}
	return nil
}

func (d *dataLayer) GetRateCharges(ctx context.Context, orgID, rateID int64) ([]spec.RateChargeItem, error) {
	query := `
SELECT 
	id, org_id, rate_id, charge_category, charge_code, charge_name,
	calculation_basis, quantity, unit_price, currency,
	minimum_amount, maximum_amount, included_in_base_rate,
	display_order, notes, created_at, updated_at
FROM rate_charge_items
WHERE org_id = ? AND rate_id = ?
ORDER BY display_order ASC, id ASC`

	var charges []spec.RateChargeItem
	if err := d.db.SelectContext(ctx, &charges, query, orgID, rateID); err != nil {
		return nil, fmt.Errorf("rates.GetRateCharges: %w", err)
	}
	return charges, nil
}

func (d *dataLayer) GetRateChargeByID(ctx context.Context, orgID, rateID, chargeID int64) (*spec.RateChargeItem, error) {
	query := `
SELECT 
	id, org_id, rate_id, charge_category, charge_code, charge_name,
	calculation_basis, quantity, unit_price, currency,
	minimum_amount, maximum_amount, included_in_base_rate,
	display_order, notes, created_at, updated_at
FROM rate_charge_items
WHERE org_id = ? AND rate_id = ? AND id = ?
LIMIT 1`

	var charge spec.RateChargeItem
	if err := d.db.GetContext(ctx, &charge, query, orgID, rateID, chargeID); err != nil {
		return nil, fmt.Errorf("rates.GetRateChargeByID: %w", err)
	}
	return &charge, nil
}

func (d *dataLayer) UpdateRateCharge(ctx context.Context, charge *spec.RateChargeItem) error {
	query := `
UPDATE rate_charge_items
SET 
	charge_category = :charge_category,
	charge_code = :charge_code,
	charge_name = :charge_name,
	calculation_basis = :calculation_basis,
	quantity = :quantity,
	unit_price = :unit_price,
	currency = :currency,
	minimum_amount = :minimum_amount,
	maximum_amount = :maximum_amount,
	included_in_base_rate = :included_in_base_rate,
	display_order = :display_order,
	notes = :notes,
	updated_at = NOW()
WHERE org_id = :org_id AND rate_id = :rate_id AND id = :id`

	res, err := d.db.NamedExecContext(ctx, query, charge)
	if err != nil {
		return fmt.Errorf("rates.UpdateRateCharge: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("rate charge not found or no changes made")
	}
	return nil
}

func (d *dataLayer) DeleteRateCharge(ctx context.Context, orgID, rateID, chargeID int64) error {
	query := `DELETE FROM rate_charge_items WHERE org_id = ? AND rate_id = ? AND id = ?`
	res, err := d.db.ExecContext(ctx, query, orgID, rateID, chargeID)
	if err != nil {
		return fmt.Errorf("rates.DeleteRateCharge: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("rate charge not found")
	}
	return nil
}

func (d *dataLayer) ReorderRateCharges(ctx context.Context, orgID, rateID int64, chargeIDs []int64) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("rates.ReorderRateCharges begin tx: %w", err)
	}
	defer tx.Rollback()

	query := `UPDATE rate_charge_items SET display_order = ?, updated_at = NOW() WHERE org_id = ? AND rate_id = ? AND id = ?`
	for order, id := range chargeIDs {
		if _, err := tx.ExecContext(ctx, query, order, orgID, rateID, id); err != nil {
			return fmt.Errorf("rates.ReorderRateCharges update order %d for id %d: %w", order, id, err)
		}
	}

	return tx.Commit()
}

// ── Task 19.3: Carrier Rate Contracts & Versions Implementation ───────────────

func (d *dataLayer) CreateRateContract(ctx context.Context, contract *spec.RateContract) error {
	query := `
INSERT INTO rate_contracts (
	org_id, contract_reference, carrier_name, carrier_code, contract_name,
	contract_type, transport_mode, currency, effective_date, expiry_date,
	status, renewal_status, renewal_owner, notes, created_by, updated_by,
	created_at, updated_at
) VALUES (
	:org_id, :contract_reference, :carrier_name, :carrier_code, :contract_name,
	:contract_type, :transport_mode, :currency, :effective_date, :expiry_date,
	:status, :renewal_status, :renewal_owner, :notes, :created_by, :updated_by,
	NOW(), NOW()
)`

	res, err := d.db.NamedExecContext(ctx, query, contract)
	if err != nil {
		return fmt.Errorf("rates.CreateRateContract: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		contract.ID = id
	}
	return nil
}

func (d *dataLayer) GetRateContractByID(ctx context.Context, orgID, contractID int64) (*spec.RateContract, error) {
	query := `
SELECT 
	id, org_id, contract_reference, carrier_name, carrier_code, contract_name,
	contract_type, transport_mode, currency, effective_date, expiry_date,
	status, renewal_status, renewal_owner, notes, created_by, updated_by,
	created_at, updated_at
FROM rate_contracts
WHERE org_id = ? AND id = ?
LIMIT 1`

	var contract spec.RateContract
	if err := d.db.GetContext(ctx, &contract, query, orgID, contractID); err != nil {
		return nil, fmt.Errorf("rates.GetRateContractByID: %w", err)
	}
	return &contract, nil
}

func (d *dataLayer) ListRateContracts(ctx context.Context, filters *spec.ListRateContractsRequest) ([]*spec.RateContractListItem, int, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "rc.org_id = ?")
	args = append(args, filters.OrgID)

	if filters.Search != "" {
		conditions = append(conditions, "(rc.contract_reference LIKE ? OR rc.carrier_name LIKE ? OR rc.contract_name LIKE ?)")
		p := "%" + filters.Search + "%"
		args = append(args, p, p, p)
	}
	if filters.CarrierName != "" && filters.CarrierName != "ALL" {
		conditions = append(conditions, "rc.carrier_name = ?")
		args = append(args, filters.CarrierName)
	}
	if filters.ContractType != "" && filters.ContractType != "ALL" {
		conditions = append(conditions, "rc.contract_type = ?")
		args = append(args, filters.ContractType)
	}
	if filters.Status != "" && filters.Status != "ALL" {
		conditions = append(conditions, "rc.status = ?")
		args = append(args, filters.Status)
	}
	if filters.RenewalStatus != "" && filters.RenewalStatus != "ALL" {
		conditions = append(conditions, "rc.renewal_status = ?")
		args = append(args, filters.RenewalStatus)
	}
	if filters.TransportMode != "" && filters.TransportMode != "ALL" {
		conditions = append(conditions, "rc.transport_mode = ?")
		args = append(args, filters.TransportMode)
	}

	whereClause := joinWithAnd(conditions)
	if whereClause != "" {
		whereClause = "WHERE " + whereClause
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM rate_contracts rc %s`, whereClause)
	var totalCount int
	if err := d.db.GetContext(ctx, &totalCount, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("rates.ListRateContracts count: %w", err)
	}

	limit := filters.Limit
	if limit <= 0 {
		limit = 15
	}
	page := filters.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	orderCol := "rc.created_at"
	orderDir := "DESC"
	if filters.SortBy == "expiry_date" {
		orderCol = "rc.expiry_date"
	} else if filters.SortBy == "carrier_name" {
		orderCol = "rc.carrier_name"
	}
	if filters.SortOrder == "ASC" || filters.SortOrder == "asc" {
		orderDir = "ASC"
	}

	selectQuery := fmt.Sprintf(`
SELECT 
	rc.id, rc.contract_reference, rc.carrier_name, rc.carrier_code, rc.contract_name,
	rc.contract_type, COALESCE(rc.transport_mode, 'Ocean FCL') as transport_mode,
	COALESCE(rc.currency, 'USD') as currency,
	DATE_FORMAT(rc.effective_date, '%%Y-%%m-%%d') as effective_date,
	DATE_FORMAT(rc.expiry_date, '%%Y-%%m-%%d') as expiry_date,
	rc.status, rc.renewal_status, rc.renewal_owner, rc.updated_at,
	COALESCE((SELECT COUNT(*) FROM rates r WHERE r.org_id = rc.org_id AND r.contract_id = rc.id), 0) as linked_rate_count
FROM rate_contracts rc
%s
ORDER BY %s %s
LIMIT %d OFFSET %d`, whereClause, orderCol, orderDir, limit, offset)

	var items []*spec.RateContractListItem
	if err := d.db.SelectContext(ctx, &items, selectQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("rates.ListRateContracts select: %w", err)
	}

	now := time.Now().UTC()
	for _, item := range items {
		if exp, err := time.Parse("2006-01-02", item.ExpiryDate); err == nil {
			item.DaysUntilExpiry = int(exp.Sub(now).Hours() / 24)
		}
	}

	return items, totalCount, nil
}

func (d *dataLayer) UpdateRateContract(ctx context.Context, contract *spec.RateContract) error {
	query := `
UPDATE rate_contracts
SET 
	contract_reference = :contract_reference,
	carrier_name = :carrier_name,
	carrier_code = :carrier_code,
	contract_name = :contract_name,
	contract_type = :contract_type,
	transport_mode = :transport_mode,
	currency = :currency,
	effective_date = :effective_date,
	expiry_date = :expiry_date,
	status = :status,
	renewal_status = :renewal_status,
	renewal_owner = :renewal_owner,
	notes = :notes,
	updated_by = :updated_by,
	updated_at = NOW()
WHERE org_id = :org_id AND id = :id`

	res, err := d.db.NamedExecContext(ctx, query, contract)
	if err != nil {
		return fmt.Errorf("rates.UpdateRateContract: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("rate contract not found or unchanged")
	}
	return nil
}

func (d *dataLayer) ArchiveRateContract(ctx context.Context, orgID, contractID int64, archivedBy string) error {
	query := `
UPDATE rate_contracts
SET status = 'ARCHIVED', updated_by = ?, updated_at = NOW()
WHERE org_id = ? AND id = ?`

	res, err := d.db.ExecContext(ctx, query, archivedBy, orgID, contractID)
	if err != nil {
		return fmt.Errorf("rates.ArchiveRateContract: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("rate contract not found")
	}
	return nil
}

func (d *dataLayer) GetRateContractSummary(ctx context.Context, orgID int64) (*spec.RateContractSummary, error) {
	summary := &spec.RateContractSummary{}

	query := `
SELECT 
	COUNT(*) as total_contracts,
	COUNT(CASE WHEN status = 'ACTIVE' THEN 1 END) as active_contracts,
	COUNT(CASE WHEN status = 'EXPIRING_SOON' THEN 1 END) as expiring_soon_contracts,
	COUNT(CASE WHEN status = 'EXPIRED' THEN 1 END) as expired_contracts,
	COUNT(CASE WHEN renewal_status IN ('NOT_STARTED', 'IN_PROGRESS') AND (status = 'EXPIRING_SOON' OR status = 'EXPIRED') THEN 1 END) as renewal_required
FROM rate_contracts
WHERE org_id = ?`

	type counts struct {
		Total        int `db:"total_contracts"`
		Active       int `db:"active_contracts"`
		ExpiringSoon int `db:"expiring_soon_contracts"`
		Expired      int `db:"expired_contracts"`
		RenewalReq   int `db:"renewal_required"`
	}
	var c counts
	if err := d.db.GetContext(ctx, &c, query, orgID); err == nil {
		summary.TotalContracts = c.Total
		summary.ActiveContracts = c.Active
		summary.ExpiringSoonContracts = c.ExpiringSoon
		summary.ExpiredContracts = c.Expired
		summary.RenewalRequired = c.RenewalReq
	}

	linkedRateQuery := `SELECT COUNT(*) FROM rates WHERE org_id = ? AND contract_id IS NOT NULL`
	_ = d.db.GetContext(ctx, &summary.TotalLinkedRates, linkedRateQuery, orgID)

	expiringListQuery := `
SELECT 
	rc.id, rc.contract_reference, rc.carrier_name, rc.carrier_code, rc.contract_name,
	rc.contract_type, COALESCE(rc.transport_mode, 'Ocean FCL') as transport_mode,
	COALESCE(rc.currency, 'USD') as currency,
	DATE_FORMAT(rc.effective_date, '%%Y-%%m-%%d') as effective_date,
	DATE_FORMAT(rc.expiry_date, '%%Y-%%m-%%d') as expiry_date,
	rc.status, rc.renewal_status, rc.renewal_owner, rc.updated_at,
	COALESCE((SELECT COUNT(*) FROM rates r WHERE r.org_id = rc.org_id AND r.contract_id = rc.id), 0) as linked_rate_count
FROM rate_contracts rc
WHERE rc.org_id = ? AND (rc.status = 'EXPIRING_SOON' OR rc.status = 'EXPIRED')
ORDER BY rc.expiry_date ASC
LIMIT 5`

	var expiringList []spec.RateContractListItem
	if err := d.db.SelectContext(ctx, &expiringList, expiringListQuery, orgID); err == nil {
		now := time.Now().UTC()
		for i := range expiringList {
			if exp, err := time.Parse("2006-01-02", expiringList[i].ExpiryDate); err == nil {
				expiringList[i].DaysUntilExpiry = int(exp.Sub(now).Hours() / 24)
			}
		}
		summary.ExpiringSoonList = expiringList
	}

	return summary, nil
}

func (d *dataLayer) GetRatesByContract(ctx context.Context, orgID, contractID int64) ([]*spec.RateListItem, error) {
	query := `
SELECT 
	id, rate_reference, carrier_name, carrier_code, rate_type,
	transport_mode, service_type, equipment_type,
	origin_port, origin_code, destination_port, destination_code,
	currency, base_amount,
	DATE_FORMAT(effective_date, '%%Y-%%m-%%d') as effective_date,
	DATE_FORMAT(expiry_date, '%%Y-%%m-%%d') as expiry_date,
	status, contract_id, version_number, version_status,
	carrier_reference, contract_reference, updated_at
FROM rates
WHERE org_id = ? AND contract_id = ?
ORDER BY version_number DESC, id DESC`

	var items []*spec.RateListItem
	if err := d.db.SelectContext(ctx, &items, query, orgID, contractID); err != nil {
		return nil, fmt.Errorf("rates.GetRatesByContract: %w", err)
	}
	return items, nil
}

func (d *dataLayer) IsRateReferencedDownstream(ctx context.Context, orgID, rateID int64) (bool, error) {
	// Check if this rate has been referenced in quotations table
	var count int
	query := `SELECT COUNT(*) FROM quotations WHERE org_id = ? AND rate_id = ?`
	if err := d.db.GetContext(ctx, &count, query, orgID, rateID); err != nil {
		// If quotations table has no rate_id column, fallback to safe check
		return false, nil
	}
	return count > 0, nil
}

func (d *dataLayer) CreateRateVersionRecord(ctx context.Context, newRate *spec.Rate, history *spec.RateVersionHistory) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("rates.CreateRateVersionRecord begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. Insert new rate row
	rateQuery := `
INSERT INTO rates (
	org_id, rate_reference, carrier_name, carrier_code, service_provider,
	rate_type, transport_mode, service_type, equipment_type,
	origin_port, origin_code, destination_port, destination_code,
	currency, base_amount, effective_date, expiry_date, status,
	contract_id, version_number, version_status, supersedes_rate_id,
	superseded_by_rate_id, version_created_at,
	carrier_reference, contract_reference, notes, created_by, updated_by,
	created_at, updated_at
) VALUES (
	:org_id, :rate_reference, :carrier_name, :carrier_code, :service_provider,
	:rate_type, :transport_mode, :service_type, :equipment_type,
	:origin_port, :origin_code, :destination_port, :destination_code,
	:currency, :base_amount, :effective_date, :expiry_date, :status,
	:contract_id, :version_number, :version_status, :supersedes_rate_id,
	:superseded_by_rate_id, NOW(),
	:carrier_reference, :contract_reference, :notes, :created_by, :updated_by,
	NOW(), NOW()
)`

	res, err := tx.NamedExecContext(ctx, rateQuery, newRate)
	if err != nil {
		return fmt.Errorf("rates.CreateRateVersionRecord insert rate: %w", err)
	}
	newID, err := res.LastInsertId()
	if err == nil {
		newRate.ID = newID
	}

	// 2. Mark previous rate as SUPERSEDED
	if newRate.SupersedesRateID != nil && *newRate.SupersedesRateID > 0 {
		markQuery := `
UPDATE rates
SET version_status = 'SUPERSEDED', superseded_by_rate_id = ?, updated_at = NOW()
WHERE org_id = ? AND id = ?`
		if _, err := tx.ExecContext(ctx, markQuery, newRate.ID, newRate.OrgID, *newRate.SupersedesRateID); err != nil {
			return fmt.Errorf("rates.CreateRateVersionRecord supersede previous rate: %w", err)
		}
	}

	// 3. Record immutable audit log in rate_version_history
	if history != nil {
		history.RateID = newRate.ID
		history.NewRateID = &newRate.ID
		historyQuery := `
INSERT INTO rate_version_history (
	org_id, rate_id, version_number, action, previous_rate_id, new_rate_id,
	description, performed_by, metadata, created_at
) VALUES (
	:org_id, :rate_id, :version_number, :action, :previous_rate_id, :new_rate_id,
	:description, :performed_by, :metadata, NOW()
)`
		if _, err := tx.NamedExecContext(ctx, historyQuery, history); err != nil {
			return fmt.Errorf("rates.CreateRateVersionRecord audit history: %w", err)
		}
	}

	return tx.Commit()
}

func (d *dataLayer) GetRateVersionHistory(ctx context.Context, orgID, rateID int64) ([]spec.RateVersionHistory, error) {
	query := `
SELECT 
	id, org_id, rate_id, version_number, action, previous_rate_id, new_rate_id,
	description, performed_by, metadata, created_at
FROM rate_version_history
WHERE org_id = ? AND (rate_id = ? OR previous_rate_id = ? OR new_rate_id = ?)
ORDER BY created_at DESC, id DESC`

	var history []spec.RateVersionHistory
	if err := d.db.SelectContext(ctx, &history, query, orgID, rateID, rateID, rateID); err != nil {
		return nil, fmt.Errorf("rates.GetRateVersionHistory: %w", err)
	}
	return history, nil
}

func (d *dataLayer) GetRateVersionChain(ctx context.Context, orgID, rateID int64) ([]spec.RateVersionChainItem, error) {
	var currentRate spec.Rate
	if err := d.db.GetContext(ctx, &currentRate, `SELECT id, rate_reference, contract_id, version_number, supersedes_rate_id, superseded_by_rate_id FROM rates WHERE org_id = ? AND id = ?`, orgID, rateID); err != nil {
		return nil, fmt.Errorf("rates.GetRateVersionChain get current rate: %w", err)
	}

	// Trace root rate ID up the supersedes hierarchy
	rootID := currentRate.ID
	curr := currentRate
	for curr.SupersedesRateID != nil && *curr.SupersedesRateID > 0 {
		var parent spec.Rate
		if err := d.db.GetContext(ctx, &parent, `SELECT id, rate_reference, contract_id, version_number, supersedes_rate_id, superseded_by_rate_id FROM rates WHERE org_id = ? AND id = ?`, orgID, *curr.SupersedesRateID); err == nil {
			rootID = parent.ID
			curr = parent
		} else {
			break
		}
	}

	// Query all rates in the version tree
	query := `
SELECT 
	id, rate_reference, version_number, version_status, base_amount, currency,
	effective_date, expiry_date, status, supersedes_rate_id, superseded_by_rate_id, created_at
FROM rates
WHERE org_id = ? AND (
	id = ? 
	OR supersedes_rate_id = ? 
	OR superseded_by_rate_id = ? 
	OR rate_reference = ?
	OR (contract_id IS NOT NULL AND contract_id = ?)
)
ORDER BY version_number ASC, id ASC`

	contractIDVal := int64(0)
	if currentRate.ContractID != nil {
		contractIDVal = *currentRate.ContractID
	}

	var ratesInLineage []spec.Rate
	_ = d.db.SelectContext(ctx, &ratesInLineage, query, orgID, rootID, rootID, rootID, currentRate.RateReference, contractIDVal)

	chain := make([]spec.RateVersionChainItem, len(ratesInLineage))
	for i, r := range ratesInLineage {
		chain[i] = spec.RateVersionChainItem{
			RateID:           r.ID,
			RateReference:    r.RateReference,
			VersionNumber:    r.VersionNumber,
			VersionStatus:    r.VersionStatus,
			BaseAmount:       r.BaseAmount,
			Currency:         r.Currency,
			EffectiveDate:    r.EffectiveDate.Format("2006-01-02"),
			ExpiryDate:       r.ExpiryDate.Format("2006-01-02"),
			Status:           r.Status,
			CommercialTotal:  r.BaseAmount,
			SupersededByRate: r.SupersededByRateID,
			CreatedAt:        r.CreatedAt,
		}
	}

	return chain, nil
}

func (d *dataLayer) MarkRateSuperseded(ctx context.Context, orgID, previousRateID, supersedingRateID int64) error {
	query := `
UPDATE rates
SET version_status = 'SUPERSEDED', superseded_by_rate_id = ?, updated_at = NOW()
WHERE org_id = ? AND id = ?`

	_, err := d.db.ExecContext(ctx, query, supersedingRateID, orgID, previousRateID)
	if err != nil {
		return fmt.Errorf("rates.MarkRateSuperseded: %w", err)
	}
	return nil
}

// ── Task 19.4: Spot Rate Requests, Responses & Comparison DL Implementation ──

func (d *dataLayer) CreateSpotRateRequest(ctx context.Context, req *spec.SpotRateRequest) error {
	query := `
INSERT INTO spot_rate_requests (
	org_id, request_reference, customer_id, customer_name,
	origin_port, origin_code, destination_port, destination_code,
	transport_mode, service_type, equipment_type, commodity,
	cargo_weight, cargo_volume, container_quantity, ready_date,
	target_currency, required_by_date, status, notes, created_by,
	created_at, updated_at
) VALUES (
	:org_id, :request_reference, :customer_id, :customer_name,
	:origin_port, :origin_code, :destination_port, :destination_code,
	:transport_mode, :service_type, :equipment_type, :commodity,
	:cargo_weight, :cargo_volume, :container_quantity, :ready_date,
	:target_currency, :required_by_date, :status, :notes, :created_by,
	NOW(), NOW()
)`

	res, err := d.db.NamedExecContext(ctx, query, req)
	if err != nil {
		return fmt.Errorf("rates.CreateSpotRateRequest: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		req.ID = id
	}
	return nil
}

func (d *dataLayer) GetSpotRateRequestByID(ctx context.Context, orgID, requestID int64) (*spec.SpotRateRequest, error) {
	query := `
SELECT 
	id, org_id, request_reference, customer_id, customer_name,
	origin_port, origin_code, destination_port, destination_code,
	transport_mode, service_type, equipment_type, commodity,
	cargo_weight, cargo_volume, container_quantity,
	DATE_FORMAT(ready_date, '%Y-%m-%d') AS ready_date,
	target_currency,
	DATE_FORMAT(required_by_date, '%Y-%m-%d') AS required_by_date,
	status, notes, created_by, created_at, updated_at
FROM spot_rate_requests
WHERE org_id = ? AND id = ?`

	var req spec.SpotRateRequest
	if err := d.db.GetContext(ctx, &req, query, orgID, requestID); err != nil {
		return nil, fmt.Errorf("rates.GetSpotRateRequestByID: %w", err)
	}
	return &req, nil
}

func (d *dataLayer) ListSpotRateRequests(ctx context.Context, filters *spec.ListSpotRateRequestsRequest) ([]*spec.SpotRateRequestListItem, int, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "s.org_id = ?")
	args = append(args, filters.OrgID)

	if filters.Status != "" && filters.Status != "ALL" {
		conditions = append(conditions, "s.status = ?")
		args = append(args, filters.Status)
	}
	if filters.TransportMode != "" && filters.TransportMode != "ALL" {
		conditions = append(conditions, "s.transport_mode = ?")
		args = append(args, filters.TransportMode)
	}
	if filters.Search != "" {
		conditions = append(conditions, "(s.request_reference LIKE ? OR s.customer_name LIKE ? OR s.origin_port LIKE ? OR s.destination_port LIKE ?)")
		pat := "%" + filters.Search + "%"
		args = append(args, pat, pat, pat, pat)
	}

	whereClause := "WHERE " + joinWithAnd(conditions)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM spot_rate_requests s %s", whereClause)
	var totalCount int
	if err := d.db.GetContext(ctx, &totalCount, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("rates.ListSpotRateRequests count: %w", err)
	}

	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	page := filters.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	orderCol := "s.created_at"
	orderDir := "DESC"

	dataQuery := fmt.Sprintf(`
SELECT 
	s.id, s.request_reference, s.customer_name,
	s.origin_port, s.origin_code, s.destination_port, s.destination_code,
	s.transport_mode, s.service_type, s.equipment_type, s.container_quantity,
	DATE_FORMAT(s.ready_date, '%%Y-%%m-%%d') AS ready_date,
	DATE_FORMAT(s.required_by_date, '%%Y-%%m-%%d') AS required_by_date,
	s.status, s.target_currency,
	DATE_FORMAT(s.created_at, '%%Y-%%m-%%d %%H:%%i') AS created_at,
	COALESCE(COUNT(r.id), 0) AS response_count,
	COALESCE(MAX(CASE WHEN r.is_preferred = 1 THEN 1 ELSE 0 END), 0) AS has_preferred
FROM spot_rate_requests s
LEFT JOIN spot_rate_responses r ON r.spot_rate_request_id = s.id AND r.org_id = s.org_id
%s
GROUP BY s.id
ORDER BY %s %s
LIMIT ? OFFSET ?`, whereClause, orderCol, orderDir)

	dataArgs := append(args, limit, offset)
	var rows []*spec.SpotRateRequestListItem
	if err := d.db.SelectContext(ctx, &rows, dataQuery, dataArgs...); err != nil {
		return nil, 0, fmt.Errorf("rates.ListSpotRateRequests select: %w", err)
	}

	now := time.Now().UTC()
	for _, row := range rows {
		reqDate, err := time.Parse("2006-01-02", row.RequiredByDate)
		if err == nil {
			days := int(reqDate.Sub(now).Hours() / 24)
			row.DaysUntilRequired = days
		}
	}

	return rows, totalCount, nil
}

func (d *dataLayer) UpdateSpotRateRequest(ctx context.Context, req *spec.SpotRateRequest) error {
	query := `
UPDATE spot_rate_requests SET
	customer_name = :customer_name,
	transport_mode = :transport_mode,
	service_type = :service_type,
	equipment_type = :equipment_type,
	commodity = :commodity,
	cargo_weight = :cargo_weight,
	cargo_volume = :cargo_volume,
	container_quantity = :container_quantity,
	ready_date = :ready_date,
	target_currency = :target_currency,
	required_by_date = :required_by_date,
	status = :status,
	notes = :notes,
	updated_at = NOW()
WHERE org_id = :org_id AND id = :id`

	_, err := d.db.NamedExecContext(ctx, query, req)
	if err != nil {
		return fmt.Errorf("rates.UpdateSpotRateRequest: %w", err)
	}
	return nil
}

func (d *dataLayer) CancelSpotRateRequest(ctx context.Context, orgID, requestID int64) error {
	query := `UPDATE spot_rate_requests SET status = 'CANCELLED', updated_at = NOW() WHERE org_id = ? AND id = ?`
	_, err := d.db.ExecContext(ctx, query, orgID, requestID)
	if err != nil {
		return fmt.Errorf("rates.CancelSpotRateRequest: %w", err)
	}
	return nil
}

func (d *dataLayer) CreateSpotRateResponse(ctx context.Context, resp *spec.SpotRateResponse, charges []spec.SpotRateResponseCharge) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("rates.CreateSpotRateResponse begin tx: %w", err)
	}
	defer tx.Rollback()

	// Calculate total amount from base amount + charges
	totalAmount := resp.BaseAmount
	for _, c := range charges {
		totalAmount += c.Quantity * c.UnitPrice
	}
	resp.TotalAmount = totalAmount

	respQuery := `
INSERT INTO spot_rate_responses (
	org_id, spot_rate_request_id, carrier_name, carrier_code,
	supplier_name, rate_id, currency, base_amount, total_amount,
	transit_days, free_days_origin, free_days_destination,
	valid_from, valid_until, routing_notes, response_notes,
	status, is_preferred, responded_at, created_by, created_at, updated_at
) VALUES (
	:org_id, :spot_rate_request_id, :carrier_name, :carrier_code,
	:supplier_name, :rate_id, :currency, :base_amount, :total_amount,
	:transit_days, :free_days_origin, :free_days_destination,
	:valid_from, :valid_until, :routing_notes, :response_notes,
	:status, :is_preferred, NOW(), :created_by, NOW(), NOW()
)`

	res, err := tx.NamedExecContext(ctx, respQuery, resp)
	if err != nil {
		return fmt.Errorf("rates.CreateSpotRateResponse insert: %w", err)
	}
	respID, err := res.LastInsertId()
	if err == nil {
		resp.ID = respID
	}

	for i, c := range charges {
		c.OrgID = resp.OrgID
		c.SpotRateResponseID = resp.ID
		c.DisplayOrder = i
		chargeQuery := `
INSERT INTO spot_rate_response_charges (
	org_id, spot_rate_response_id, charge_category, charge_name,
	calculation_basis, quantity, unit_price, currency, display_order,
	created_at, updated_at
) VALUES (
	:org_id, :spot_rate_response_id, :charge_category, :charge_name,
	:calculation_basis, :quantity, :unit_price, :currency, :display_order,
	NOW(), NOW()
)`
		if _, err := tx.NamedExecContext(ctx, chargeQuery, c); err != nil {
			return fmt.Errorf("rates.CreateSpotRateResponse insert charge: %w", err)
		}
	}

	return tx.Commit()
}

func (d *dataLayer) GetSpotRateResponseByID(ctx context.Context, orgID, responseID int64) (*spec.SpotRateResponse, error) {
	query := `
SELECT 
	id, org_id, spot_rate_request_id, carrier_name, carrier_code,
	supplier_name, rate_id, currency, base_amount, total_amount,
	transit_days, free_days_origin, free_days_destination,
	DATE_FORMAT(valid_from, '%Y-%m-%d') AS valid_from,
	DATE_FORMAT(valid_until, '%Y-%m-%d') AS valid_until,
	routing_notes, response_notes, status, is_preferred,
	responded_at, created_by, created_at, updated_at
FROM spot_rate_responses
WHERE org_id = ? AND id = ?`

	var resp spec.SpotRateResponse
	if err := d.db.GetContext(ctx, &resp, query, orgID, responseID); err != nil {
		return nil, fmt.Errorf("rates.GetSpotRateResponseByID: %w", err)
	}

	// Fetch charges
	chargesQuery := `
SELECT 
	id, org_id, spot_rate_response_id, charge_category, charge_name,
	calculation_basis, quantity, unit_price, currency, (quantity * unit_price) AS total_charge_amount,
	display_order, created_at, updated_at
FROM spot_rate_response_charges
WHERE org_id = ? AND spot_rate_response_id = ?
ORDER BY display_order ASC, id ASC`

	var charges []spec.SpotRateResponseCharge
	_ = d.db.SelectContext(ctx, &charges, chargesQuery, orgID, responseID)
	resp.Charges = charges

	return &resp, nil
}

func (d *dataLayer) GetSpotRateResponsesByRequestID(ctx context.Context, orgID, requestID int64) ([]*spec.SpotRateResponse, error) {
	query := `
SELECT 
	id, org_id, spot_rate_request_id, carrier_name, carrier_code,
	supplier_name, rate_id, currency, base_amount, total_amount,
	transit_days, free_days_origin, free_days_destination,
	DATE_FORMAT(valid_from, '%Y-%m-%d') AS valid_from,
	DATE_FORMAT(valid_until, '%Y-%m-%d') AS valid_until,
	routing_notes, response_notes, status, is_preferred,
	responded_at, created_by, created_at, updated_at
FROM spot_rate_responses
WHERE org_id = ? AND spot_rate_request_id = ?
ORDER BY is_preferred DESC, total_amount ASC, id ASC`

	var rows []*spec.SpotRateResponse
	if err := d.db.SelectContext(ctx, &rows, query, orgID, requestID); err != nil {
		return nil, fmt.Errorf("rates.GetSpotRateResponsesByRequestID: %w", err)
	}

	for _, r := range rows {
		chargesQuery := `
SELECT 
	id, org_id, spot_rate_response_id, charge_category, charge_name,
	calculation_basis, quantity, unit_price, currency, (quantity * unit_price) AS total_charge_amount,
	display_order, created_at, updated_at
FROM spot_rate_response_charges
WHERE org_id = ? AND spot_rate_response_id = ?
ORDER BY display_order ASC, id ASC`
		var charges []spec.SpotRateResponseCharge
		_ = d.db.SelectContext(ctx, &charges, chargesQuery, orgID, r.ID)
		r.Charges = charges
	}

	return rows, nil
}

func (d *dataLayer) UpdateSpotRateResponse(ctx context.Context, resp *spec.SpotRateResponse, charges []spec.SpotRateResponseCharge) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("rates.UpdateSpotRateResponse begin tx: %w", err)
	}
	defer tx.Rollback()

	totalAmount := resp.BaseAmount
	for _, c := range charges {
		totalAmount += c.Quantity * c.UnitPrice
	}
	resp.TotalAmount = totalAmount

	query := `
UPDATE spot_rate_responses SET
	carrier_name = :carrier_name,
	carrier_code = :carrier_code,
	supplier_name = :supplier_name,
	currency = :currency,
	base_amount = :base_amount,
	total_amount = :total_amount,
	transit_days = :transit_days,
	free_days_origin = :free_days_origin,
	free_days_destination = :free_days_destination,
	valid_from = :valid_from,
	valid_until = :valid_until,
	routing_notes = :routing_notes,
	response_notes = :response_notes,
	status = :status,
	updated_at = NOW()
WHERE org_id = :org_id AND id = :id`

	if _, err := tx.NamedExecContext(ctx, query, resp); err != nil {
		return fmt.Errorf("rates.UpdateSpotRateResponse exec: %w", err)
	}

	// Replace charges if provided
	if len(charges) > 0 {
		_, _ = tx.ExecContext(ctx, `DELETE FROM spot_rate_response_charges WHERE org_id = ? AND spot_rate_response_id = ?`, resp.OrgID, resp.ID)
		for i, c := range charges {
			c.OrgID = resp.OrgID
			c.SpotRateResponseID = resp.ID
			c.DisplayOrder = i
			chargeQuery := `
INSERT INTO spot_rate_response_charges (
	org_id, spot_rate_response_id, charge_category, charge_name,
	calculation_basis, quantity, unit_price, currency, display_order,
	created_at, updated_at
) VALUES (
	:org_id, :spot_rate_response_id, :charge_category, :charge_name,
	:calculation_basis, :quantity, :unit_price, :currency, :display_order,
	NOW(), NOW()
)`
			if _, err := tx.NamedExecContext(ctx, chargeQuery, c); err != nil {
				return fmt.Errorf("rates.UpdateSpotRateResponse insert charge: %w", err)
			}
		}
	}

	return tx.Commit()
}

func (d *dataLayer) SelectPreferredSpotRateResponse(ctx context.Context, orgID, requestID, responseID int64) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("rates.SelectPreferredSpotRateResponse begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. Reset any previous preferred flag for this request
	resetQuery := `UPDATE spot_rate_responses SET is_preferred = FALSE, status = 'RECEIVED', updated_at = NOW() WHERE org_id = ? AND spot_rate_request_id = ?`
	if _, err := tx.ExecContext(ctx, resetQuery, orgID, requestID); err != nil {
		return fmt.Errorf("rates.SelectPreferredSpotRateResponse reset: %w", err)
	}

	// 2. Set preferred response
	setQuery := `UPDATE spot_rate_responses SET is_preferred = TRUE, status = 'SELECTED', updated_at = NOW() WHERE org_id = ? AND id = ? AND spot_rate_request_id = ?`
	if _, err := tx.ExecContext(ctx, setQuery, orgID, responseID, requestID); err != nil {
		return fmt.Errorf("rates.SelectPreferredSpotRateResponse set preferred: %w", err)
	}

	// 3. Mark request as SELECTED
	reqQuery := `UPDATE spot_rate_requests SET status = 'SELECTED', updated_at = NOW() WHERE org_id = ? AND id = ?`
	if _, err := tx.ExecContext(ctx, reqQuery, orgID, requestID); err != nil {
		return fmt.Errorf("rates.SelectPreferredSpotRateResponse update request status: %w", err)
	}

	return tx.Commit()
}

func (d *dataLayer) GetSpotRateRequestSummary(ctx context.Context, orgID int64) (*spec.SpotRateRequestSummary, error) {
	summary := &spec.SpotRateRequestSummary{}

	query := `
SELECT
	COUNT(*) AS total_requests,
	COUNT(CASE WHEN status IN ('DRAFT', 'SENT', 'PARTIALLY_RESPONDED') THEN 1 END) AS open_requests,
	COUNT(CASE WHEN status IN ('SENT', 'PARTIALLY_RESPONDED') THEN 1 END) AS awaiting_responses,
	COUNT(CASE WHEN status = 'RESPONDED' THEN 1 END) AS fully_responded,
	COUNT(CASE WHEN status = 'SELECTED' THEN 1 END) AS selected_requests,
	COUNT(CASE WHEN status NOT IN ('SELECTED', 'CANCELLED', 'EXPIRED') AND required_by_date BETWEEN CURDATE() AND DATE_ADD(CURDATE(), INTERVAL 3 DAY) THEN 1 END) AS expiring_soon
FROM spot_rate_requests
WHERE org_id = ?`

	row := d.db.QueryRowContext(ctx, query, orgID)
	if err := row.Scan(
		&summary.TotalRequests,
		&summary.OpenRequests,
		&summary.AwaitingResponses,
		&summary.FullyResponded,
		&summary.SelectedRequests,
		&summary.ExpiringSoon,
	); err != nil {
		return nil, fmt.Errorf("rates.GetSpotRateRequestSummary: %w", err)
	}

	// Recent requests
	recentFilter := &spec.ListSpotRateRequestsRequest{
		OrgID: orgID,
		Limit: 5,
		Page:  1,
	}
	recent, _, _ := d.ListSpotRateRequests(ctx, recentFilter)
	summary.RecentSpotRequests = recent

	return summary, nil
}

func joinWithAnd(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	res := parts[0]
	for i := 1; i < len(parts); i++ {
		res += " AND " + parts[i]
	}
	return res
}

// ── Task 19.6: Rate Lifecycle Intelligence Data Layer ─────────────────────────

func (d *dataLayer) CreateRateLifecycleEvent(ctx context.Context, event *RateLifecycleEvent) error {
	query := `
		INSERT INTO rate_lifecycle_events (
			org_id, rate_id, contract_id, event_type, previous_status, current_status, description, metadata, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`
	_, err := d.db.ExecContext(ctx, query, event.OrgID, event.RateID, event.ContractID, event.EventType, event.PreviousStatus, event.CurrentStatus, event.Description, event.MetadataJSON)
	return err
}

func (d *dataLayer) GetRateLifecycleEvents(ctx context.Context, orgID int64, limit int) ([]*RateLifecycleEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, org_id, rate_id, contract_id, event_type, previous_status, current_status, description,
		       COALESCE(metadata, '') AS metadata, created_at
		FROM rate_lifecycle_events
		WHERE org_id = ?
		ORDER BY id DESC
		LIMIT ?
	`
	var events []*RateLifecycleEvent
	if err := d.db.SelectContext(ctx, &events, query, orgID, limit); err != nil {
		return nil, fmt.Errorf("rates.GetRateLifecycleEvents: %w", err)
	}
	for _, e := range events {
		if e.MetadataJSON != "" {
			var m interface{}
			if err := json.Unmarshal([]byte(e.MetadataJSON), &m); err == nil {
				e.Metadata = m
			}
		}
	}
	if events == nil {
		events = []*RateLifecycleEvent{}
	}
	return events, nil
}

func (d *dataLayer) GetRateLifecycleSummary(ctx context.Context, orgID int64) (*RateLifecycleSummary, error) {
	summary := &RateLifecycleSummary{}

	// Rate counts
	rateQuery := `
		SELECT 
			COUNT(*) AS total,
			COUNT(CASE WHEN status = 'ACTIVE' THEN 1 END) AS active,
			COUNT(CASE WHEN status = 'EXPIRING_SOON' THEN 1 END) AS expiring_soon,
			COUNT(CASE WHEN status = 'EXPIRED' THEN 1 END) AS expired,
			COUNT(CASE WHEN status = 'SUPERSEDED' THEN 1 END) AS superseded
		FROM rates
		WHERE org_id = ? AND status != 'ARCHIVED'
	`
	row := d.db.QueryRowContext(ctx, rateQuery, orgID)
	_ = row.Scan(&summary.TotalRates, &summary.ActiveRates, &summary.ExpiringSoonRates, &summary.ExpiredRates, &summary.SupersededRates)

	// Contract counts
	contractQuery := `
		SELECT 
			COUNT(*) AS total,
			COUNT(CASE WHEN status = 'ACTIVE' THEN 1 END) AS active,
			COUNT(CASE WHEN status = 'EXPIRING_SOON' THEN 1 END) AS expiring,
			COUNT(CASE WHEN status = 'EXPIRED' THEN 1 END) AS expired,
			COUNT(CASE WHEN renewal_status IN ('IN_PROGRESS', 'NOT_STARTED') AND status IN ('EXPIRING_SOON', 'EXPIRED') THEN 1 END) AS renewal_req
		FROM rate_contracts
		WHERE org_id = ? AND status != 'ARCHIVED'
	`
	crow := d.db.QueryRowContext(ctx, contractQuery, orgID)
	_ = crow.Scan(&summary.TotalContracts, &summary.ActiveContracts, &summary.ExpiringContracts, &summary.ExpiredContracts, &summary.ContractsRequiringRenewal)

	// Quotations at risk count
	riskQuery := `
		SELECT COUNT(DISTINCT quotation_id)
		FROM quotation_rate_risk_events
		WHERE org_id = ? AND is_resolved = FALSE
	`
	_ = d.db.QueryRowContext(ctx, riskQuery, orgID).Scan(&summary.QuotationsAtRisk)

	return summary, nil
}

func (d *dataLayer) GetRatesRequiringAttention(ctx context.Context, orgID int64) ([]*RateAttentionItem, error) {
	query := `
		SELECT 
			r.id AS rate_id,
			r.carrier_name,
			COALESCE(r.carrier_code, '') AS carrier_code,
			r.rate_type,
			r.version_number,
			r.origin_port,
			r.destination_port,
			r.transport_mode,
			COALESCE(r.equipment_type, '') AS equipment_type,
			r.currency,
			r.base_amount,
			DATE_FORMAT(r.effective_date, '%Y-%m-%d') AS valid_from,
			DATE_FORMAT(r.expiry_date, '%Y-%m-%d') AS valid_until,
			r.status,
			COALESCE(DATEDIFF(r.expiry_date, CURDATE()), 999) AS days_remaining,
			CASE 
				WHEN r.status = 'EXPIRED' OR (r.expiry_date IS NOT NULL AND DATEDIFF(r.expiry_date, CURDATE()) < 0) THEN 'EXPIRED'
				WHEN r.status = 'EXPIRING_SOON' AND DATEDIFF(r.expiry_date, CURDATE()) <= 7 THEN 'EXPIRING_7D'
				WHEN r.status = 'EXPIRING_SOON' OR (r.expiry_date IS NOT NULL AND DATEDIFF(r.expiry_date, CURDATE()) <= 30) THEN 'EXPIRING_30D'
				WHEN r.status = 'SUPERSEDED' THEN 'SUPERSEDED'
				ELSE 'ACTIVE'
			END AS attention_bucket,
			(SELECT COUNT(DISTINCT qrs.quotation_id) FROM quotation_rate_selections qrs WHERE qrs.org_id = r.org_id AND qrs.rate_id = r.id AND qrs.is_active = TRUE) AS affected_quotes,
			COALESCE(c.contract_reference, '') AS contract_code
		FROM rates r
		LEFT JOIN rate_contracts c ON r.org_id = c.org_id AND r.contract_id = c.id
		WHERE r.org_id = ?
		  AND (
		      r.status IN ('EXPIRING_SOON', 'EXPIRED', 'SUPERSEDED')
		      OR (r.expiry_date IS NOT NULL AND DATEDIFF(r.expiry_date, CURDATE()) <= 30)
		  )
		ORDER BY 
			CASE 
				WHEN r.status = 'EXPIRED' THEN 1
				WHEN r.status = 'EXPIRING_SOON' THEN 2
				WHEN r.status = 'SUPERSEDED' THEN 3
				ELSE 4
			END,
			days_remaining ASC
		LIMIT 100
	`
	var items []*RateAttentionItem
	if err := d.db.SelectContext(ctx, &items, query, orgID); err != nil {
		return nil, fmt.Errorf("rates.GetRatesRequiringAttention: %w", err)
	}
	if items == nil {
		items = []*RateAttentionItem{}
	}
	return items, nil
}

func (d *dataLayer) GetContractsRequiringAttention(ctx context.Context, orgID int64) ([]*ContractAttentionItem, error) {
	query := `
		SELECT 
			c.id AS contract_id,
			c.carrier_name,
			c.contract_reference AS contract_code,
			c.contract_name AS contract_title,
			DATE_FORMAT(c.effective_date, '%Y-%m-%d') AS start_date,
			DATE_FORMAT(c.expiry_date, '%Y-%m-%d') AS end_date,
			c.status,
			c.renewal_status,
			COALESCE(DATEDIFF(c.expiry_date, CURDATE()), 999) AS days_remaining,
			(SELECT COUNT(*) FROM rates r WHERE r.org_id = c.org_id AND r.contract_id = c.id AND r.status != 'ARCHIVED') AS linked_rates_count,
			(
				SELECT COUNT(DISTINCT qrs.quotation_id) 
				FROM quotation_rate_selections qrs 
				JOIN rates r2 ON qrs.rate_id = r2.id
				WHERE qrs.org_id = c.org_id AND r2.contract_id = c.id AND qrs.is_active = TRUE
			) AS affected_quotes
		FROM rate_contracts c
		WHERE c.org_id = ?
		  AND c.status IN ('EXPIRING_SOON', 'EXPIRED')
		ORDER BY days_remaining ASC
		LIMIT 50
	`
	var items []*ContractAttentionItem
	if err := d.db.SelectContext(ctx, &items, query, orgID); err != nil {
		return nil, fmt.Errorf("rates.GetContractsRequiringAttention: %w", err)
	}
	if items == nil {
		items = []*ContractAttentionItem{}
	}
	return items, nil
}

func (d *dataLayer) GetAllActiveAndExpiringRatesForEvaluation(ctx context.Context, orgID int64) ([]*spec.Rate, error) {
	query := `
		SELECT id, org_id, rate_reference, carrier_name, COALESCE(carrier_code, '') AS carrier_code, rate_type,
		       version_number, contract_id, origin_port, destination_port, transport_mode,
		       COALESCE(equipment_type, '') AS equipment_type, currency, base_amount,
		       effective_date, expiry_date, status, created_at, updated_at
		FROM rates
		WHERE org_id = ? AND status != 'ARCHIVED'
	`
	var ratesList []*spec.Rate
	if err := d.db.SelectContext(ctx, &ratesList, query, orgID); err != nil {
		return nil, err
	}
	return ratesList, nil
}

func (d *dataLayer) GetAllActiveAndExpiringContractsForEvaluation(ctx context.Context, orgID int64) ([]*spec.RateContract, error) {
	query := `
		SELECT id, org_id, contract_reference, carrier_name, COALESCE(carrier_code, '') AS carrier_code,
		       contract_name, contract_type, effective_date, expiry_date, status, renewal_status,
		       created_at, updated_at
		FROM rate_contracts
		WHERE org_id = ? AND status != 'ARCHIVED'
	`
	var contractsList []*spec.RateContract
	if err := d.db.SelectContext(ctx, &contractsList, query, orgID); err != nil {
		return nil, err
	}
	return contractsList, nil
}

func (d *dataLayer) UpdateRateStatusDirect(ctx context.Context, orgID, rateID int64, newStatus string) error {
	query := `UPDATE rates SET status = ?, updated_at = NOW() WHERE org_id = ? AND id = ?`
	_, err := d.db.ExecContext(ctx, query, newStatus, orgID, rateID)
	return err
}

func (d *dataLayer) UpdateContractStatusDirect(ctx context.Context, orgID, contractID int64, newStatus, newRenewalStatus string) error {
	query := `UPDATE rate_contracts SET status = ?, renewal_status = ?, updated_at = NOW() WHERE org_id = ? AND id = ?`
	_, err := d.db.ExecContext(ctx, query, newStatus, newRenewalStatus, orgID, contractID)
	return err
}

// ── Task 19.7: Rate Analytics & Procurement Intelligence ────────────────────────

func (d *dataLayer) GetRateAnalyticsOverview(ctx context.Context, orgID int64) (*RateAnalyticsOverview, error) {
	const q = `
		SELECT
		  COALESCE(SUM(CASE WHEN 1=1 THEN 1 ELSE 0 END), 0)                                          AS total_rates,
		  COALESCE(SUM(CASE WHEN status = 'ACTIVE' THEN 1 ELSE 0 END), 0)                            AS active_rates,
		  COALESCE(SUM(CASE WHEN status = 'EXPIRING_SOON' THEN 1 ELSE 0 END), 0)                     AS expiring_soon_rates,
		  COALESCE(SUM(CASE WHEN status = 'EXPIRED' THEN 1 ELSE 0 END), 0)                           AS expired_rates,
		  COALESCE(SUM(CASE WHEN status = 'SUPERSEDED' THEN 1 ELSE 0 END), 0)                        AS superseded_rates,
		  COALESCE(SUM(CASE WHEN status = 'DRAFT' THEN 1 ELSE 0 END), 0)                             AS draft_rates
		FROM rates
		WHERE org_id = ? AND status != 'ARCHIVED'
	`
	type rateRow struct {
		TotalRates        int `db:"total_rates"`
		ActiveRates       int `db:"active_rates"`
		ExpiringSoonRates int `db:"expiring_soon_rates"`
		ExpiredRates      int `db:"expired_rates"`
		SupersededRates   int `db:"superseded_rates"`
		DraftRates        int `db:"draft_rates"`
	}
	var rr rateRow
	if err := d.db.GetContext(ctx, &rr, q, orgID); err != nil {
		return nil, fmt.Errorf("GetRateAnalyticsOverview rates: %w", err)
	}

	const qContracts = `
		SELECT
		  COALESCE(COUNT(*), 0)                                                                       AS total_contracts,
		  COALESCE(SUM(CASE WHEN status = 'ACTIVE' THEN 1 ELSE 0 END), 0)                            AS active_contracts,
		  COALESCE(SUM(CASE WHEN renewal_status IN ('NOT_STARTED','PENDING') AND status = 'ACTIVE'
		               AND end_date IS NOT NULL AND end_date <= DATE_ADD(NOW(), INTERVAL 30 DAY) THEN 1 ELSE 0 END), 0) AS contracts_requiring_renewal
		FROM rate_contracts
		WHERE org_id = ? AND status != 'ARCHIVED'
	`
	type contractRow struct {
		TotalContracts            int `db:"total_contracts"`
		ActiveContracts           int `db:"active_contracts"`
		ContractsRequiringRenewal int `db:"contracts_requiring_renewal"`
	}
	var cr contractRow
	if err := d.db.GetContext(ctx, &cr, qContracts, orgID); err != nil {
		return nil, fmt.Errorf("GetRateAnalyticsOverview contracts: %w", err)
	}

	const qCoverage = `
		SELECT
		  COALESCE(COUNT(DISTINCT CONCAT(origin_port,'|',destination_port,'|',transport_mode)), 0) AS total_lanes_covered,
		  COALESCE(COUNT(DISTINCT carrier_name), 0)                                                AS total_carriers
		FROM rates
		WHERE org_id = ? AND status IN ('ACTIVE','EXPIRING_SOON')
	`
	type coverageRow struct {
		TotalLanesCovered int `db:"total_lanes_covered"`
		TotalCarriers     int `db:"total_carriers"`
	}
	var cov coverageRow
	if err := d.db.GetContext(ctx, &cov, qCoverage, orgID); err != nil {
		return nil, fmt.Errorf("GetRateAnalyticsOverview coverage: %w", err)
	}

	const qSpot = `
		SELECT
		  COALESCE(COUNT(*), 0)                                                                 AS total_spot_requests,
		  COALESCE(SUM(CASE WHEN status = 'RESPONDED' OR status = 'SELECTED' THEN 1 ELSE 0 END), 0) AS spot_requests_responded,
		  COALESCE(SUM(CASE WHEN status = 'SELECTED' THEN 1 ELSE 0 END), 0)                   AS spot_requests_selected,
		  COALESCE(SUM(CASE WHEN status = 'EXPIRED' THEN 1 ELSE 0 END), 0)                    AS spot_requests_expired
		FROM spot_rate_requests
		WHERE org_id = ?
	`
	type spotRow struct {
		TotalSpotRequests     int `db:"total_spot_requests"`
		SpotRequestsResponded int `db:"spot_requests_responded"`
		SpotRequestsSelected  int `db:"spot_requests_selected"`
		SpotRequestsExpired   int `db:"spot_requests_expired"`
	}
	var sr spotRow
	if err := d.db.GetContext(ctx, &sr, qSpot, orgID); err != nil {
		return nil, fmt.Errorf("GetRateAnalyticsOverview spot: %w", err)
	}

	const qQ = `
		SELECT
		  COALESCE((SELECT COUNT(*) FROM quotation_rate_selections qrs
		            JOIN quotations q ON q.id = qrs.quotation_id
		            WHERE q.org_id = ?), 0)                                               AS quote_rate_selection_count,
		  COALESCE((SELECT COUNT(*) FROM quotation_rate_risk_events qrre
		            JOIN quotations q ON q.id = qrre.quotation_id
		            WHERE q.org_id = ? AND qrre.is_resolved = FALSE), 0)                 AS quotation_risk_count
	`
	type qRow struct {
		QuoteToRateSelectionCount    int `db:"quote_rate_selection_count"`
		QuotationCommercialRiskCount int `db:"quotation_risk_count"`
	}
	var qr qRow
	if err := d.db.GetContext(ctx, &qr, qQ, orgID, orgID); err != nil {
		return nil, fmt.Errorf("GetRateAnalyticsOverview quotations: %w", err)
	}

	return &RateAnalyticsOverview{
		TotalRates:                   rr.TotalRates,
		ActiveRates:                  rr.ActiveRates,
		ExpiringSoonRates:            rr.ExpiringSoonRates,
		ExpiredRates:                 rr.ExpiredRates,
		SupersededRates:              rr.SupersededRates,
		DraftRates:                   rr.DraftRates,
		TotalContracts:               cr.TotalContracts,
		ActiveContracts:              cr.ActiveContracts,
		ContractsRequiringRenewal:    cr.ContractsRequiringRenewal,
		TotalLanesCovered:            cov.TotalLanesCovered,
		TotalCarriers:                cov.TotalCarriers,
		TotalSpotRequests:            sr.TotalSpotRequests,
		SpotRequestsResponded:        sr.SpotRequestsResponded,
		SpotRequestsSelected:         sr.SpotRequestsSelected,
		SpotRequestsExpired:          sr.SpotRequestsExpired,
		QuoteToRateSelectionCount:    qr.QuoteToRateSelectionCount,
		QuotationCommercialRiskCount: qr.QuotationCommercialRiskCount,
	}, nil
}

func (d *dataLayer) GetRateAnalyticsTrends(ctx context.Context, orgID int64, days int) ([]RateTrendDataPoint, error) {
	if days <= 0 {
		days = 30
	}
	// Build a date series and left-join activity counts per day.
	// Uses per-status created_at from the rates table and lifecycle events for expired/superseded.
	const q = `
		SELECT
		  DATE_FORMAT(d.dt, '%Y-%m-%d')   AS date,
		  COALESCE(r.rates_created, 0)    AS rates_created,
		  COALESCE(r.rates_activated, 0)  AS rates_activated,
		  COALESCE(le.rates_expired, 0)   AS rates_expired,
		  COALESCE(ls.rates_superseded, 0) AS rates_superseded,
		  COALESCE(rc.contracts_created, 0) AS contracts_created,
		  COALESCE(sr.spot_requests_created, 0) AS spot_requests_created,
		  COALESCE(sr.spot_requests_selected, 0) AS spot_requests_selected
		FROM (
		  SELECT DATE_SUB(CURDATE(), INTERVAL n DAY) AS dt
		  FROM (
		    SELECT a.N + b.N*10 AS n
		    FROM
		      (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4
		       UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) a,
		      (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4
		       UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) b
		  ) nums
		  WHERE a.N + b.N*10 < ?
		) d
		LEFT JOIN (
		  SELECT DATE(created_at) AS day,
		         COUNT(*)         AS rates_created,
		         SUM(CASE WHEN status = 'ACTIVE' THEN 1 ELSE 0 END) AS rates_activated
		  FROM rates
		  WHERE org_id = ? AND created_at >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
		  GROUP BY DATE(created_at)
		) r ON r.day = d.dt
		LEFT JOIN (
		  SELECT DATE(created_at) AS day,
		         COUNT(*)         AS rates_expired
		  FROM rate_lifecycle_events
		  WHERE org_id = ? AND event_type = 'RATE_EXPIRED'
		    AND created_at >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
		  GROUP BY DATE(created_at)
		) le ON le.day = d.dt
		LEFT JOIN (
		  SELECT DATE(created_at) AS day,
		         COUNT(*)         AS rates_superseded
		  FROM rate_lifecycle_events
		  WHERE org_id = ? AND event_type = 'RATE_SUPERSEDED'
		    AND created_at >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
		  GROUP BY DATE(created_at)
		) ls ON ls.day = d.dt
		LEFT JOIN (
		  SELECT DATE(created_at) AS day, COUNT(*) AS contracts_created
		  FROM rate_contracts
		  WHERE org_id = ? AND created_at >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
		  GROUP BY DATE(created_at)
		) rc ON rc.day = d.dt
		LEFT JOIN (
		  SELECT DATE(srr.created_at) AS day,
		         COUNT(*)             AS spot_requests_created,
		         SUM(CASE WHEN srr.status = 'SELECTED' THEN 1 ELSE 0 END) AS spot_requests_selected
		  FROM spot_rate_requests srr
		  WHERE srr.org_id = ? AND srr.created_at >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
		  GROUP BY DATE(srr.created_at)
		) sr ON sr.day = d.dt
		ORDER BY d.dt ASC
	`
	var rows []RateTrendDataPoint
	err := d.db.SelectContext(ctx, &rows, q,
		days,
		orgID, days,
		orgID, days,
		orgID, days,
		orgID, days,
		orgID, days,
	)
	if err != nil {
		return nil, fmt.Errorf("GetRateAnalyticsTrends: %w", err)
	}
	if rows == nil {
		rows = []RateTrendDataPoint{}
	}
	return rows, nil
}

func (d *dataLayer) GetCarrierRatePerformance(ctx context.Context, orgID int64) ([]CarrierRatePerformance, error) {
	const q = `
		SELECT
		  r.carrier_name,
		  COALESCE(r.carrier_code, '')                                              AS carrier_code,
		  COUNT(*)                                                                  AS total_rates,
		  SUM(CASE WHEN r.status = 'ACTIVE' THEN 1 ELSE 0 END)                     AS active_rates,
		  SUM(CASE WHEN r.status = 'EXPIRING_SOON' THEN 1 ELSE 0 END)              AS expiring_rates,
		  SUM(CASE WHEN r.status = 'EXPIRED' THEN 1 ELSE 0 END)                    AS expired_rates,
		  COUNT(DISTINCT CONCAT(r.origin_port,'|',r.destination_port,'|',r.transport_mode)) AS lanes_covered,
		  COALESCE((SELECT COUNT(DISTINCT rc.id) FROM rate_contracts rc
		            WHERE rc.org_id = r.org_id AND rc.carrier_name = r.carrier_name
		            AND rc.status != 'ARCHIVED'), 0)                                AS contracts_count,
		  COALESCE(resp.response_count, 0)                                         AS spot_responses_count,
		  COALESCE(resp.selection_count, 0)                                        AS spot_selections,
		  COALESCE(AVG(r.transit_days), 0)                                         AS avg_transit_days
		FROM rates r
		LEFT JOIN (
		  SELECT srr.carrier_name,
		         COUNT(*)                                            AS response_count,
		         SUM(CASE WHEN srr.status = 'SELECTED' THEN 1 ELSE 0 END) AS selection_count
		  FROM spot_rate_responses srr
		  JOIN spot_rate_requests req ON req.id = srr.spot_request_id AND req.org_id = ?
		  GROUP BY srr.carrier_name
		) resp ON resp.carrier_name = r.carrier_name
		WHERE r.org_id = ? AND r.status != 'ARCHIVED'
		GROUP BY r.carrier_name, r.carrier_code, r.org_id, resp.response_count, resp.selection_count
		ORDER BY active_rates DESC, total_rates DESC
	`
	var rows []CarrierRatePerformance
	if err := d.db.SelectContext(ctx, &rows, q, orgID, orgID); err != nil {
		return nil, fmt.Errorf("GetCarrierRatePerformance: %w", err)
	}
	if rows == nil {
		rows = []CarrierRatePerformance{}
	}
	// Compute derived fields
	for i := range rows {
		if rows[i].SpotResponsesCount > 0 {
			rows[i].SelectionRate = float64(rows[i].SpotSelections) / float64(rows[i].SpotResponsesCount) * 100
		}
		// Rate health: HEALTHY if <10% expiring+expired, ATTENTION if <30%, CRITICAL otherwise
		total := rows[i].TotalRates
		if total == 0 {
			rows[i].RateHealthStatus = "HEALTHY"
			continue
		}
		atRisk := rows[i].ExpiringRates + rows[i].ExpiredRates
		ratio := float64(atRisk) / float64(total)
		switch {
		case ratio >= 0.3:
			rows[i].RateHealthStatus = "CRITICAL"
		case ratio >= 0.1:
			rows[i].RateHealthStatus = "ATTENTION"
		default:
			rows[i].RateHealthStatus = "HEALTHY"
		}
	}
	return rows, nil
}

func (d *dataLayer) GetLaneRatePerformance(ctx context.Context, orgID int64) ([]LaneRatePerformance, error) {
	// Step 1: Get lane-level counts
	const qLanes = `
		SELECT
		  origin_port,
		  destination_port,
		  transport_mode,
		  COALESCE(MAX(service_type), '')    AS service_type,
		  COALESCE(MAX(equipment_type), '')  AS equipment_type,
		  COUNT(*)                           AS available_rates,
		  SUM(CASE WHEN status IN ('ACTIVE','EXPIRING_SOON') THEN 1 ELSE 0 END) AS active_rates,
		  COUNT(DISTINCT carrier_name)       AS carrier_count
		FROM rates
		WHERE org_id = ? AND status != 'ARCHIVED'
		GROUP BY origin_port, destination_port, transport_mode
		ORDER BY active_rates DESC, available_rates DESC
	`
	type laneRow struct {
		Origin        string `db:"origin_port"`
		Destination   string `db:"destination_port"`
		TransportMode string `db:"transport_mode"`
		ServiceType   string `db:"service_type"`
		EquipmentType string `db:"equipment_type"`
		AvailableRates int   `db:"available_rates"`
		ActiveRates    int   `db:"active_rates"`
		CarrierCount   int   `db:"carrier_count"`
	}
	var laneRows []laneRow
	if err := d.db.SelectContext(ctx, &laneRows, qLanes, orgID); err != nil {
		return nil, fmt.Errorf("GetLaneRatePerformance lanes: %w", err)
	}

	// Step 2: Get per-lane, per-currency price stats
	const qPrices = `
		SELECT
		  origin_port, destination_port, transport_mode, currency,
		  MIN(base_amount) AS cheapest_rate,
		  AVG(base_amount) AS average_rate,
		  MAX(base_amount) AS highest_rate
		FROM rates
		WHERE org_id = ? AND status IN ('ACTIVE','EXPIRING_SOON') AND base_amount > 0
		GROUP BY origin_port, destination_port, transport_mode, currency
	`
	type priceRow struct {
		Origin        string  `db:"origin_port"`
		Destination   string  `db:"destination_port"`
		TransportMode string  `db:"transport_mode"`
		Currency      string  `db:"currency"`
		CheapestRate  float64 `db:"cheapest_rate"`
		AverageRate   float64 `db:"average_rate"`
		HighestRate   float64 `db:"highest_rate"`
	}
	var priceRows []priceRow
	if err := d.db.SelectContext(ctx, &priceRows, qPrices, orgID); err != nil {
		return nil, fmt.Errorf("GetLaneRatePerformance prices: %w", err)
	}

	// Step 3: Get spot request counts per lane
	const qSpot = `
		SELECT origin_port, destination_port, transport_mode,
		       COUNT(*) AS spot_request_count,
		       SUM(CASE WHEN status='SELECTED' THEN 1 ELSE 0 END) AS selected_rate_count
		FROM spot_rate_requests
		WHERE org_id = ?
		GROUP BY origin_port, destination_port, transport_mode
	`
	type spotLaneRow struct {
		Origin           string `db:"origin_port"`
		Destination      string `db:"destination_port"`
		TransportMode    string `db:"transport_mode"`
		SpotRequestCount int    `db:"spot_request_count"`
		SelectedCount    int    `db:"selected_rate_count"`
	}
	var spotLaneRows []spotLaneRow
	if err := d.db.SelectContext(ctx, &spotLaneRows, qSpot, orgID); err != nil {
		return nil, fmt.Errorf("GetLaneRatePerformance spot: %w", err)
	}

	// Build price map: key -> []LaneCurrencyBreakdown
	type laneKey struct{ Origin, Destination, Mode string }
	priceMap := make(map[laneKey][]LaneCurrencyBreakdown)
	for _, pr := range priceRows {
		k := laneKey{pr.Origin, pr.Destination, pr.TransportMode}
		priceMap[k] = append(priceMap[k], LaneCurrencyBreakdown{
			Currency:     pr.Currency,
			CheapestRate: pr.CheapestRate,
			AverageRate:  pr.AverageRate,
			HighestRate:  pr.HighestRate,
		})
	}
	spotMap := make(map[laneKey]spotLaneRow)
	for _, sr := range spotLaneRows {
		spotMap[laneKey{sr.Origin, sr.Destination, sr.TransportMode}] = sr
	}

	result := make([]LaneRatePerformance, 0, len(laneRows))
	for _, lr := range laneRows {
		k := laneKey{lr.Origin, lr.Destination, lr.TransportMode}
		spotInfo := spotMap[k]
		cov := "COVERED"
		if lr.ActiveRates == 0 {
			cov = "UNCOVERED"
		} else if lr.CarrierCount <= 1 {
			cov = "LIMITED"
		}
		result = append(result, LaneRatePerformance{
			Origin:            lr.Origin,
			Destination:       lr.Destination,
			TransportMode:     lr.TransportMode,
			ServiceType:       lr.ServiceType,
			EquipmentType:     lr.EquipmentType,
			AvailableRates:    lr.AvailableRates,
			ActiveRates:       lr.ActiveRates,
			CarrierCount:      lr.CarrierCount,
			SpotRequestCount:  spotInfo.SpotRequestCount,
			SelectedRateCount: spotInfo.SelectedCount,
			CurrencyBreakdown: priceMap[k],
			CoverageStatus:    cov,
		})
	}
	return result, nil
}

func (d *dataLayer) GetRateLifecycleAnalytics(ctx context.Context, orgID int64) (*RateLifecycleAnalytics, error) {
	const q = `
		SELECT
		  COALESCE(SUM(CASE WHEN status = 'ACTIVE' THEN 1 ELSE 0 END), 0)          AS active_count,
		  COALESCE(SUM(CASE WHEN status = 'EXPIRING_SOON' THEN 1 ELSE 0 END), 0)   AS expiring_soon_count,
		  COALESCE(SUM(CASE WHEN status = 'EXPIRED' THEN 1 ELSE 0 END), 0)         AS expired_count,
		  COALESCE(SUM(CASE WHEN status = 'SUPERSEDED' THEN 1 ELSE 0 END), 0)      AS superseded_count,
		  COALESCE(SUM(CASE WHEN status = 'ARCHIVED' THEN 1 ELSE 0 END), 0)        AS archived_count,
		  COALESCE(SUM(CASE WHEN status = 'DRAFT' THEN 1 ELSE 0 END), 0)           AS draft_count
		FROM rates
		WHERE org_id = ?
	`
	var la RateLifecycleAnalytics
	if err := d.db.GetContext(ctx, &la, q, orgID); err != nil {
		return nil, fmt.Errorf("GetRateLifecycleAnalytics: %w", err)
	}
	la.TotalRates = la.Active + la.ExpiringSoon + la.Expired + la.Superseded + la.Archived + la.Draft

	// Contract renewal count
	const qCR = `
		SELECT COALESCE(COUNT(*), 0)
		FROM rate_contracts
		WHERE org_id = ? AND status = 'ACTIVE'
		  AND renewal_status IN ('NOT_STARTED','PENDING')
		  AND end_date IS NOT NULL AND end_date <= DATE_ADD(NOW(), INTERVAL 30 DAY)
	`
	var crCount int
	if err := d.db.GetContext(ctx, &crCount, qCR, orgID); err != nil {
		return nil, fmt.Errorf("GetRateLifecycleAnalytics contracts: %w", err)
	}
	la.ContractRenewalRequired = crCount

	// Unresolved commercial risk events
	const qRisk = `
		SELECT COALESCE(COUNT(*), 0)
		FROM quotation_rate_risk_events qrre
		JOIN quotations q ON q.id = qrre.quotation_id
		WHERE q.org_id = ? AND qrre.is_resolved = FALSE
	`
	var riskCount int
	if err := d.db.GetContext(ctx, &riskCount, qRisk, orgID); err != nil {
		return nil, fmt.Errorf("GetRateLifecycleAnalytics risks: %w", err)
	}
	la.CommercialRiskEvents = riskCount

	return &la, nil
}

func (d *dataLayer) GetSpotSourcingPerformance(ctx context.Context, orgID int64) (*SpotSourcingPerformance, error) {
	const q = `
		SELECT
		  COALESCE(COUNT(*), 0)                                                     AS total_requests,
		  COALESCE(SUM(CASE WHEN status = 'SENT' THEN 1 ELSE 0 END), 0)             AS awaiting_responses,
		  COALESCE(SUM(CASE WHEN status IN ('RESPONDED','SELECTED') THEN 1 ELSE 0 END), 0) AS fully_responded,
		  COALESCE(SUM(CASE WHEN status = 'SELECTED' THEN 1 ELSE 0 END), 0)         AS selected,
		  COALESCE(SUM(CASE WHEN status = 'EXPIRED' THEN 1 ELSE 0 END), 0)          AS expired,
		  COALESCE(SUM(CASE WHEN status = 'CANCELLED' THEN 1 ELSE 0 END), 0)        AS cancelled
		FROM spot_rate_requests
		WHERE org_id = ?
	`
	var sp SpotSourcingPerformance
	if err := d.db.GetContext(ctx, &sp, q, orgID); err != nil {
		return nil, fmt.Errorf("GetSpotSourcingPerformance requests: %w", err)
	}

	// Avg responses per request
	const qAvg = `
		SELECT COALESCE(AVG(resp_count), 0)
		FROM (
		  SELECT srr.spot_request_id, COUNT(*) AS resp_count
		  FROM spot_rate_responses srr
		  JOIN spot_rate_requests req ON req.id = srr.spot_request_id AND req.org_id = ?
		  GROUP BY srr.spot_request_id
		) sub
	`
	if err := d.db.GetContext(ctx, &sp.AverageResponsesPerRequest, qAvg, orgID); err != nil {
		sp.AverageResponsesPerRequest = 0
	}

	// Computed rates
	if sp.TotalRequests > 0 {
		sp.SelectionRate = float64(sp.Selected) / float64(sp.TotalRequests) * 100
		sp.ResponseRate = float64(sp.FullyResponded) / float64(sp.TotalRequests) * 100
	}

	// Top 5 carrier participants
	const qPart = `
		SELECT srr.carrier_name,
		       COUNT(*)                                             AS responses_count,
		       SUM(CASE WHEN srr.status = 'SELECTED' THEN 1 ELSE 0 END) AS selections_count
		FROM spot_rate_responses srr
		JOIN spot_rate_requests req ON req.id = srr.spot_request_id AND req.org_id = ?
		GROUP BY srr.carrier_name
		ORDER BY responses_count DESC
		LIMIT 5
	`
	var parts []SpotCarrierParticipation
	if err := d.db.SelectContext(ctx, &parts, qPart, orgID); err == nil && parts != nil {
		sp.CarrierParticipation = parts
	}
	if sp.CarrierParticipation == nil {
		sp.CarrierParticipation = []SpotCarrierParticipation{}
	}
	return &sp, nil
}

func (d *dataLayer) GetCommercialRiskExposure(ctx context.Context, orgID int64) (int, error) {
	const q = `
		SELECT COALESCE(COUNT(*), 0)
		FROM quotation_rate_risk_events qrre
		JOIN quotations q ON q.id = qrre.quotation_id
		WHERE q.org_id = ? AND qrre.is_resolved = FALSE
	`
	var count int
	if err := d.db.GetContext(ctx, &count, q, orgID); err != nil {
		return 0, fmt.Errorf("GetCommercialRiskExposure: %w", err)
	}
	return count, nil
}
