package rates

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Repository handles all database operations for the rate_entries table.
type Repository interface {
	Upsert(ctx context.Context, rates []CanonicalRate) error
	Search(ctx context.Context, q RateQuery) ([]CanonicalRate, error)
	GetByID(ctx context.Context, orgID int64, id string) (*CanonicalRate, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Upsert(ctx context.Context, rates []CanonicalRate) error {
	if len(rates) == 0 {
		return nil
	}

	const query = `
INSERT INTO rate_entries (
    id, org_id, source, source_ref, contract_doc_id,
    origin_port, destination_port, via_port, service_code,
    carrier_scac, carrier_name, vessel_name, equipment_type,
    ocean_freight, origin_charges, destination_charges, surcharges, total_buy_price,
    currency_original, exchange_rate_used,
    included_charges, excluded_charges,
    free_days_origin, free_days_destination, transit_days,
    incoterms, commodity_restrictions, routing_conditions,
    valid_from, valid_until,
    confidence_score, extraction_status, extracted_by,
    nautical_miles, co2_per_teu,
    created_at, updated_at
) VALUES (
    :id, :org_id, :source, :source_ref, :contract_doc_id,
    :origin_port, :destination_port, :via_port, :service_code,
    :carrier_scac, :carrier_name, :vessel_name, :equipment_type,
    :ocean_freight, :origin_charges, :destination_charges, :surcharges, :total_buy_price,
    :currency_original, :exchange_rate_used,
    :included_charges, :excluded_charges,
    :free_days_origin, :free_days_destination, :transit_days,
    :incoterms, :commodity_restrictions, :routing_conditions,
    :valid_from, :valid_until,
    :confidence_score, :extraction_status, :extracted_by,
    :nautical_miles, :co2_per_teu,
    NOW(), NOW()
)
ON DUPLICATE KEY UPDATE
    ocean_freight        = VALUES(ocean_freight),
    origin_charges       = VALUES(origin_charges),
    destination_charges  = VALUES(destination_charges),
    surcharges           = VALUES(surcharges),
    total_buy_price      = VALUES(total_buy_price),
    transit_days         = VALUES(transit_days),
    free_days_origin     = VALUES(free_days_origin),
    free_days_destination= VALUES(free_days_destination),
    valid_from           = VALUES(valid_from),
    valid_until          = VALUES(valid_until),
    confidence_score     = VALUES(confidence_score),
    extraction_status    = VALUES(extraction_status),
    vessel_name          = VALUES(vessel_name),
    service_code         = VALUES(service_code),
    nautical_miles       = VALUES(nautical_miles),
    co2_per_teu          = VALUES(co2_per_teu),
    contract_doc_id      = VALUES(contract_doc_id),
    updated_at           = NOW()
`
	type dbRow struct {
		ID                    string  `db:"id"`
		OrgID                 int64   `db:"org_id"`
		Source                string  `db:"source"`
		SourceRef             string  `db:"source_ref"`
		ContractDocID         *string `db:"contract_doc_id"`
		OriginPort            string  `db:"origin_port"`
		DestinationPort       string  `db:"destination_port"`
		ViaPort               string  `db:"via_port"`
		ServiceCode           string  `db:"service_code"`
		CarrierSCAC           string  `db:"carrier_scac"`
		CarrierName           string  `db:"carrier_name"`
		VesselName            string  `db:"vessel_name"`
		EquipmentType         string  `db:"equipment_type"`
		OceanFreight          float64 `db:"ocean_freight"`
		OriginCharges         float64 `db:"origin_charges"`
		DestinationCharges    float64 `db:"destination_charges"`
		Surcharges            []byte  `db:"surcharges"`
		TotalBuyPrice         float64 `db:"total_buy_price"`
		CurrencyOriginal      string  `db:"currency_original"`
		ExchangeRateUsed      float64 `db:"exchange_rate_used"`
		IncludedCharges       []byte  `db:"included_charges"`
		ExcludedCharges       []byte  `db:"excluded_charges"`
		FreeDaysOrigin        int     `db:"free_days_origin"`
		FreeDaysDestination   int     `db:"free_days_destination"`
		TransitDays           *int    `db:"transit_days"`
		Incoterms             string  `db:"incoterms"`
		CommodityRestrictions []byte  `db:"commodity_restrictions"`
		RoutingConditions     string  `db:"routing_conditions"`
		ValidFrom             string  `db:"valid_from"`
		ValidUntil            string  `db:"valid_until"`
		ConfidenceScore       int     `db:"confidence_score"`
		ExtractionStatus      string  `db:"extraction_status"`
		ExtractedBy           string  `db:"extracted_by"`
		NauticalMiles         int     `db:"nautical_miles"`
		CO2PerTEU             float64 `db:"co2_per_teu"`
	}

	rows := make([]dbRow, 0, len(rates))
	for _, cr := range rates {
		surchargesJSON, err := json.Marshal(cr.Surcharges)
		if err != nil {
			return fmt.Errorf("marshal surcharges for rate %s: %w", cr.ID, err)
		}

		included := cr.IncludedCharges
		if included == nil {
			included = []string{}
		}
		incJSON, _ := json.Marshal(included)

		excluded := cr.ExcludedCharges
		if excluded == nil {
			excluded = []string{}
		}
		excJSON, _ := json.Marshal(excluded)

		restrictions := cr.CommodityRestrictions
		if restrictions == nil {
			restrictions = []string{}
		}
		resJSON, _ := json.Marshal(restrictions)

		rows = append(rows, dbRow{
			ID:                    cr.ID,
			OrgID:                 cr.OrgID,
			Source:                string(cr.Source),
			SourceRef:             cr.SourceRef,
			ContractDocID:         cr.ContractDocID,
			OriginPort:            cr.OriginPort,
			DestinationPort:       cr.DestinationPort,
			ViaPort:               cr.ViaPort,
			ServiceCode:           cr.ServiceCode,
			CarrierSCAC:           cr.CarrierSCAC,
			CarrierName:           cr.CarrierName,
			VesselName:            cr.VesselName,
			EquipmentType:         cr.EquipmentType,
			OceanFreight:          cr.OceanFreight,
			OriginCharges:         cr.OriginCharges,
			DestinationCharges:    cr.DestinationCharges,
			Surcharges:            surchargesJSON,
			TotalBuyPrice:         cr.TotalBuyPrice,
			CurrencyOriginal:      cr.CurrencyOriginal,
			ExchangeRateUsed:      cr.ExchangeRateUsed,
			IncludedCharges:       incJSON,
			ExcludedCharges:       excJSON,
			FreeDaysOrigin:        cr.FreeDaysOrigin,
			FreeDaysDestination:   cr.FreeDaysDestination,
			TransitDays:           cr.TransitDays,
			Incoterms:             cr.Incoterms,
			CommodityRestrictions: resJSON,
			RoutingConditions:     cr.RoutingConditions,
			ValidFrom:             cr.ValidFrom.Format("2006-01-02"),
			ValidUntil:            cr.ValidUntil.Format("2006-01-02"),
			ConfidenceScore:       cr.ConfidenceScore,
			ExtractionStatus:      string(cr.ExtractionStatus),
			ExtractedBy:           cr.ExtractedBy,
			NauticalMiles:         cr.NauticalMiles,
			CO2PerTEU:             cr.CO2PerTEU,
		})
	}

	_, err := r.db.NamedExecContext(ctx, query, rows)
	return err
}

func (r *repository) Search(ctx context.Context, q RateQuery) ([]CanonicalRate, error) {
	equipmentType := q.EquipmentType
	if equipmentType == "" {
		equipmentType = "40GP"
	}
	maxResults := q.MaxResults
	if maxResults <= 0 {
		maxResults = 20
	}

	const baseQuery = `
SELECT
    id, org_id, source, source_ref, contract_doc_id,
    origin_port, destination_port, COALESCE(via_port,'') AS via_port,
    COALESCE(service_code,'') AS service_code,
    carrier_scac, carrier_name, COALESCE(vessel_name,'') AS vessel_name,
    equipment_type,
    ocean_freight, origin_charges, destination_charges,
    surcharges, total_buy_price,
    currency_original, exchange_rate_used,
    included_charges, excluded_charges,
    free_days_origin, free_days_destination, transit_days,
    COALESCE(incoterms,'') AS incoterms,
    commodity_restrictions,
    COALESCE(routing_conditions,'') AS routing_conditions,
    valid_from, valid_until,
    confidence_score, extraction_status, extracted_by,
    COALESCE(nautical_miles, 0) AS nautical_miles,
    COALESCE(co2_per_teu, 0) AS co2_per_teu,
    created_at, updated_at
FROM rate_entries
WHERE
    org_id           = ?
    AND origin_port      = ?
    AND destination_port = ?
    AND equipment_type   = ?
    AND extraction_status = 'CONFIRMED'
    AND valid_from  <= ?
    AND valid_until >= ?
ORDER BY
    CASE source WHEN 'CONTRACT_PDF' THEN 0 ELSE 1 END,
    confidence_score DESC,
    total_buy_price  ASC
LIMIT ?
`
	var args []interface{}

	if q.TargetDate != nil {
		targetStr := q.TargetDate.Format("2006-01-02")
		args = []interface{}{q.OrgID, q.OriginPort, q.DestinationPort, equipmentType, targetStr, targetStr, maxResults}
	} else {
		args = []interface{}{q.OrgID, q.OriginPort, q.DestinationPort, equipmentType, "2099-12-31", "2000-01-01", maxResults}
	}

	type dbRow struct {
		ID                       string  `db:"id"`
		OrgID                    int64   `db:"org_id"`
		Source                   string  `db:"source"`
		SourceRef                string  `db:"source_ref"`
		ContractDocID            *string `db:"contract_doc_id"`
		OriginPort               string  `db:"origin_port"`
		DestinationPort          string  `db:"destination_port"`
		ViaPort                  string  `db:"via_port"`
		ServiceCode              string  `db:"service_code"`
		CarrierSCAC              string  `db:"carrier_scac"`
		CarrierName              string  `db:"carrier_name"`
		VesselName               string  `db:"vessel_name"`
		EquipmentType            string  `db:"equipment_type"`
		OceanFreight             float64 `db:"ocean_freight"`
		OriginCharges            float64 `db:"origin_charges"`
		DestinationCharges       float64 `db:"destination_charges"`
		SurchargesRaw            []byte  `db:"surcharges"`
		TotalBuyPrice            float64 `db:"total_buy_price"`
		CurrencyOriginal         string  `db:"currency_original"`
		ExchangeRateUsed         float64 `db:"exchange_rate_used"`
		IncludedChargesRaw       []byte  `db:"included_charges"`
		ExcludedChargesRaw       []byte  `db:"excluded_charges"`
		FreeDaysOrigin           int     `db:"free_days_origin"`
		FreeDaysDestination      int     `db:"free_days_destination"`
		TransitDays              *int    `db:"transit_days"`
		Incoterms                string  `db:"incoterms"`
		CommodityRestrictionsRaw []byte  `db:"commodity_restrictions"`
		RoutingConditions        string  `db:"routing_conditions"`
		ValidFrom                string  `db:"valid_from"`
		ValidUntil               string  `db:"valid_until"`
		ConfidenceScore          int     `db:"confidence_score"`
		ExtractionStatus         string  `db:"extraction_status"`
		ExtractedBy              string  `db:"extracted_by"`
		NauticalMiles            int     `db:"nautical_miles"`
		CO2PerTEU                float64 `db:"co2_per_teu"`
		CreatedAt                string  `db:"created_at"`
		UpdatedAt                string  `db:"updated_at"`
	}

	rows := []dbRow{}
	if err := r.db.SelectContext(ctx, &rows, baseQuery, args...); err != nil {
		return nil, fmt.Errorf("rate_entries search: %w", err)
	}

	results := make([]CanonicalRate, 0, len(rows))
	for _, row := range rows {
		var surcharges []Surcharge
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
		cr := CanonicalRate{
			ID:                    row.ID,
			OrgID:                 row.OrgID,
			Source:                RateSource(row.Source),
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
			ExtractionStatus:      ExtractionStatus(row.ExtractionStatus),
			ExtractedBy:           row.ExtractedBy,
			NauticalMiles:         row.NauticalMiles,
			CO2PerTEU:             row.CO2PerTEU,
		}
		results = append(results, cr)
	}
	return results, nil
}

// GetByID returns a single rate entry by its UUID.
func (r *repository) GetByID(ctx context.Context, orgID int64, id string) (*CanonicalRate, error) {
	const query = `
SELECT id, org_id, source, source_ref, contract_doc_id,
    origin_port, destination_port, COALESCE(via_port,'') AS via_port,
    COALESCE(service_code,'') AS service_code,
    carrier_scac, carrier_name, COALESCE(vessel_name,'') AS vessel_name,
    equipment_type,
    ocean_freight, origin_charges, destination_charges,
    surcharges, total_buy_price, currency_original, exchange_rate_used,
    included_charges, excluded_charges,
    free_days_origin, free_days_destination, transit_days,
    COALESCE(incoterms,'') AS incoterms,
    commodity_restrictions, COALESCE(routing_conditions,'') AS routing_conditions,
    valid_from, valid_until,
    confidence_score, extraction_status, extracted_by,
    COALESCE(nautical_miles, 0) AS nautical_miles,
    COALESCE(co2_per_teu, 0) AS co2_per_teu,
    created_at, updated_at
FROM rate_entries
WHERE id = ? AND org_id = ?
`
	row := struct {
		ID                    string  `db:"id"`
		OrgID                 int64   `db:"org_id"`
		Source                string  `db:"source"`
		SourceRef             string  `db:"source_ref"`
		ContractDocID         *string `db:"contract_doc_id"`
		OriginPort            string  `db:"origin_port"`
		DestinationPort       string  `db:"destination_port"`
		ViaPort               string  `db:"via_port"`
		ServiceCode           string  `db:"service_code"`
		CarrierSCAC           string  `db:"carrier_scac"`
		CarrierName           string  `db:"carrier_name"`
		VesselName            string  `db:"vessel_name"`
		EquipmentType         string  `db:"equipment_type"`
		OceanFreight          float64 `db:"ocean_freight"`
		OriginCharges         float64 `db:"origin_charges"`
		DestinationCharges    float64 `db:"destination_charges"`
		SurchargesRaw         []byte  `db:"surcharges"`
		TotalBuyPrice         float64 `db:"total_buy_price"`
		CurrencyOriginal      string  `db:"currency_original"`
		ExchangeRateUsed      float64 `db:"exchange_rate_used"`
		IncludedChargesRaw    []byte  `db:"included_charges"`
		ExcludedChargesRaw    []byte  `db:"excluded_charges"`
		FreeDaysOrigin        int     `db:"free_days_origin"`
		FreeDaysDestination   int     `db:"free_days_destination"`
		TransitDays           *int    `db:"transit_days"`
		Incoterms             string  `db:"incoterms"`
		CommodityRestrictionsRaw []byte `db:"commodity_restrictions"`
		RoutingConditions     string  `db:"routing_conditions"`
		ValidFrom             string  `db:"valid_from"`
		ValidUntil            string  `db:"valid_until"`
		ConfidenceScore       int     `db:"confidence_score"`
		ExtractionStatus      string  `db:"extraction_status"`
		ExtractedBy           string  `db:"extracted_by"`
		NauticalMiles         int     `db:"nautical_miles"`
		CO2PerTEU             float64 `db:"co2_per_teu"`
		CreatedAt             string  `db:"created_at"`
		UpdatedAt             string  `db:"updated_at"`
	}{}
	if err := r.db.GetContext(ctx, &row, query, id, orgID); err != nil {
		return nil, fmt.Errorf("rate_entries get by id: %w", err)
	}
	var surcharges []Surcharge
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
	cr := &CanonicalRate{
		ID:                    row.ID,
		OrgID:                 row.OrgID,
		Source:                RateSource(row.Source),
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
		ExtractionStatus:      ExtractionStatus(row.ExtractionStatus),
		ExtractedBy:           row.ExtractedBy,
		NauticalMiles:         row.NauticalMiles,
		CO2PerTEU:             row.CO2PerTEU,
	}
	return cr, nil
}
