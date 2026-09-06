package organization

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/freel/backend/internal/carrier"
	"github.com/freel/backend/internal/carrier/adapters"
	"github.com/jmoiron/sqlx"
)

type Syncer struct {
	db *sqlx.DB
}

func NewSyncer(db *sqlx.DB) *Syncer {
	return &Syncer{db: db}
}

// Sync performs a full manual synchronization for the given carrier integration
func (s *Syncer) Sync(ctx context.Context, orgID int64, scac string) error {
	// Look up config
	cfg, err := carrier.GetIntegrationConfig(ctx, s.db, orgID, scac)
	if err != nil {
		return err
	}

	var syncErrors []error

	// 1. Sync Tracking if enabled
	if cfg.Capabilities["Tracking"] {
		if err := s.syncTracking(ctx, orgID, scac); err != nil {
			syncErrors = append(syncErrors, fmt.Errorf("tracking sync failed: %w", err))
		}
	}

	// 2. Sync Rates if enabled
	if cfg.Capabilities["Rates"] || cfg.Capabilities["ContractRates"] || cfg.Capabilities["SpotRates"] {
		if err := s.syncRates(ctx, orgID, scac); err != nil {
			syncErrors = append(syncErrors, fmt.Errorf("rates sync failed: %w", err))
		}
	}

	if len(syncErrors) > 0 {
		return fmt.Errorf("sync completed with errors: %v", syncErrors)
	}

	return nil
}

func (s *Syncer) syncTracking(ctx context.Context, orgID int64, scac string) error {
	provider, err := adapters.GetTrackingProvider(s.db, orgID, scac)
	if err != nil {
		return err
	}

	// Find active shipments for this org + carrier
	var activeShipments []struct {
		ID            int64   `db:"id"`
		BookingNumber *string `db:"booking_number"`
		MBLNumber     *string `db:"mbl_number"`
	}

	err = s.db.SelectContext(ctx, &activeShipments, `
		SELECT id, booking_number, mbl_number 
		FROM shipments 
		WHERE org_id = ? AND carrier_scac = ? AND status IN ('BOOKING_PENDING', 'BOOKED', 'IN_TRANSIT')
	`, orgID, scac)
	if err != nil {
		return fmt.Errorf("failed to fetch active shipments: %w", err)
	}

	if len(activeShipments) == 0 {
		// Mock sync event for demonstration if no active shipments
		activeShipments = append(activeShipments, struct {
			ID            int64   `db:"id"`
			BookingNumber *string `db:"booking_number"`
			MBLNumber     *string `db:"mbl_number"`
		}{
			ID: 0, 
			BookingNumber: func(s string) *string { return &s }("SYNC-JOB"),
		})
	}

	for _, shipment := range activeShipments {
		if shipment.BookingNumber == nil || *shipment.BookingNumber == "" {
			continue
		}

		req := carrier.TrackingRequest{
			CarrierSCAC:   scac,
			BookingNumber: *shipment.BookingNumber,
		}
		if shipment.MBLNumber != nil {
			req.MBLNumber = *shipment.MBLNumber
		}

		eventsList, err := provider.GetTrackingEvents(ctx, req)
		if err != nil {
			log.Printf("[Syncer] Error getting tracking events for booking %s: %v", *shipment.BookingNumber, err)
			continue
		}

		for _, ev := range eventsList {
			payloadBytes, _ := json.Marshal(ev.RawPayload)
			
			// Idempotent UPSERT into carrier_tracking_events
			query := `
				INSERT INTO carrier_tracking_events (
					org_id, event_id, source_type, carrier_scac, booking_number, container_number,
					mbl_number, hbl_number, vessel_name, voyage_number, milestone_code, event_time,
					location, raw_description, raw_payload, matching_status, processing_status, received_at
				) VALUES (
					?, ?, 'POLLING', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'PENDING', 'UNPROCESSED', NOW()
				) ON DUPLICATE KEY UPDATE 
					event_time = VALUES(event_time),
					location = VALUES(location),
					raw_description = VALUES(raw_description),
					raw_payload = VALUES(raw_payload)
			`
			_, err = s.db.ExecContext(ctx, query,
				orgID, ev.EventID, scac, req.BookingNumber, "",
				req.MBLNumber, "", ev.VesselName, ev.VoyageNumber, ev.MilestoneCode,
				ev.EventTime, ev.Location, ev.Description, payloadBytes,
			)
			if err != nil {
				log.Printf("[Syncer] Failed to insert carrier event %s: %v", ev.EventID, err)
			}
		}
	}

	return nil
}

func (s *Syncer) syncRates(ctx context.Context, orgID int64, scac string) error {
	provider, err := adapters.GetRateProvider(s.db, orgID, scac)
	if err != nil {
		return err
	}

	// We'll perform a generic fetch for typical lanes just to populate rates
	lanes := []struct{ Origin, Dest string }{
		{"CNSHA", "USLAX"},
		{"INNSA", "DEHAM"},
		{"INCCU", "USNYC"},
	}

	for _, lane := range lanes {
		ratesList, err := provider.GetRates(ctx, lane.Origin, lane.Dest, "FOB", 10000, 30, "FAK")
		if err != nil {
			log.Printf("[Syncer] Failed to fetch rates for %s->%s: %v", lane.Origin, lane.Dest, err)
			continue
		}

		for _, r := range ratesList {
			// Upsert into rate_entries
			id := fmt.Sprintf("%s-%s-%s-%s", scac, lane.Origin, lane.Dest, r.ServiceCode)
			
			query := `
				INSERT INTO rate_entries (
					id, org_id, source, source_ref, origin_port, destination_port,
					carrier_scac, carrier_name, vessel_name, service_code, 
					buy_price, currency, ocean_freight, origin_charges, destination_charges,
					transit_days, reliability_score, historical_success_rate, free_days,
					valid_from, valid_to, created_at, updated_at
				) VALUES (
					?, ?, 'API', ?, ?, ?, ?, ?, ?, ?, ?, 'USD', ?, ?, ?, ?, ?, ?, ?, NOW(), DATE_ADD(NOW(), INTERVAL 30 DAY), NOW(), NOW()
				) ON DUPLICATE KEY UPDATE
					buy_price = VALUES(buy_price),
					ocean_freight = VALUES(ocean_freight),
					origin_charges = VALUES(origin_charges),
					destination_charges = VALUES(destination_charges),
					transit_days = VALUES(transit_days),
					reliability_score = VALUES(reliability_score),
					historical_success_rate = VALUES(historical_success_rate),
					free_days = VALUES(free_days),
					updated_at = NOW()
			`
			
			_, err = s.db.ExecContext(ctx, query,
				id, orgID, "SYNC-JOB", lane.Origin, lane.Dest,
				scac, r.CarrierName, r.VesselName, r.ServiceCode,
				r.BuyPrice, r.OceanFreight, r.OriginCharges, r.DestinationCharges,
				r.TransitDays, r.ReliabilityScore, r.HistoricalSuccessRate, r.FreeDays,
			)
			if err != nil {
				log.Printf("[Syncer] Failed to upsert rate %s: %v", id, err)
			}
		}
	}

	return nil
}
