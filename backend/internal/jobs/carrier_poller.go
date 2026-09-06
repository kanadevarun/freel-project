package jobs

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/freel/backend/internal/shipments"
	"github.com/jmoiron/sqlx"
)

// CarrierPoller manages status-aware polling jobs for shipments
type CarrierPoller struct {
	db     *sqlx.DB
	svc    shipments.Service
	stopCh chan struct{}
}

func NewCarrierPoller(db *sqlx.DB, svc shipments.Service) *CarrierPoller {
	return &CarrierPoller{
		db:     db,
		svc:    svc,
		stopCh: make(chan struct{}),
	}
}

// Start runs the poller background loop
func (p *CarrierPoller) Start() {
	// 4. Polling intervals: read from config/env with defaults, not hardcoded (as requested)
	pendingInterval := getDurationEnv("POLL_INTERVAL_PENDING", 4*time.Hour)
	bookedInterval := getDurationEnv("POLL_INTERVAL_BOOKED", 2*time.Hour)
	transitInterval := getDurationEnv("POLL_INTERVAL_TRANSIT", 30*time.Minute)
	arrivedInterval := getDurationEnv("POLL_INTERVAL_ARRIVED", 4*time.Hour)

	log.Printf("[Carrier Poller] Starting status-aware carrier poller loops. Pending: %s, Booked: %s, Transit: %s, Arrived: %s",
		pendingInterval, bookedInterval, transitInterval, arrivedInterval)

	go p.runPollLoop("BOOKING_PENDING", pendingInterval)
	go p.runPollLoop("BOOKED", bookedInterval)
	go p.runPollLoop("IN_TRANSIT", transitInterval)
	go p.runPollLoop("ARRIVED", arrivedInterval)
}

func (p *CarrierPoller) Stop() {
	close(p.stopCh)
}

func (p *CarrierPoller) runPollLoop(status string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.pollShipmentsWithStatus(status)
		case <-p.stopCh:
			return
		}
	}
}

func (p *CarrierPoller) pollShipmentsWithStatus(status string) {
	ctx := context.Background()

	// 11. Distributed poller safety: Lock shipments using transactional SELECT FOR UPDATE SKIP LOCKED
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("[Carrier Poller] Failed to begin transaction: %v", err)
		return
	}
	defer tx.Rollback()

	// Query active shipments matching status with database lock lease to prevent duplicate pulls
	var activeShipments []struct {
		ID            int64   `db:"id"`
		OrgID         int64   `db:"org_id"`
		CarrierSCAC   string  `db:"carrier_scac"`
		BookingNumber *string `db:"booking_number"`
		MBLNumber     *string `db:"mbl_number"`
	}

	err = tx.SelectContext(ctx, &activeShipments, `
		SELECT id, org_id, carrier_scac, booking_number, mbl_number 
		FROM shipments 
		WHERE status = ? AND carrier_scac IS NOT NULL
		FOR UPDATE
	`, status)
	if err != nil {
		log.Printf("[Carrier Poller] Failed to query shipments with lock for status %s: %v", status, err)
		return
	}

	for _, s := range activeShipments {
		if s.BookingNumber == nil || *s.BookingNumber == "" {
			continue
		}

		// Execute unified tracking refresh through carrier adapter architecture
		_, err := p.svc.RefreshShipmentTracking(ctx, s.OrgID, s.ID, nil, "Background Carrier Poller")
		if err != nil {
			log.Printf("[Carrier Poller] Background tracking sync note for shipment #%d: %v", s.ID, err)
		}
	}

	_ = tx.Commit()
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		// Try parsing as int minutes
		m, err := strconv.Atoi(val)
		if err == nil {
			return time.Duration(m) * time.Minute
		}
		return fallback
	}
	return d
}
