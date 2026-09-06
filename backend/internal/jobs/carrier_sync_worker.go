package jobs

import (
	"context"
	"log"
	"time"

	"github.com/freel/backend/internal/carrier/domain"
	carrierService "github.com/freel/backend/internal/carrier/service"
	"github.com/jmoiron/sqlx"
)

type CarrierSyncWorker struct {
	db         *sqlx.DB
	carrierSvc carrierService.CarrierService
	stopCh     chan struct{}
}

func NewCarrierSyncWorker(db *sqlx.DB, carrierSvc carrierService.CarrierService) *CarrierSyncWorker {
	return &CarrierSyncWorker{
		db:         db,
		carrierSvc: carrierSvc,
		stopCh:     make(chan struct{}),
	}
}

func (w *CarrierSyncWorker) Start() {
	log.Println("[Carrier Sync Worker] Starting robust background sync loop...")
	go w.runLoop()
}

func (w *CarrierSyncWorker) Stop() {
	close(w.stopCh)
}

func (w *CarrierSyncWorker) runLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.processDueIntegrations()
		case <-w.stopCh:
			return
		}
	}
}

func (w *CarrierSyncWorker) processDueIntegrations() {
	ctx := context.Background()

	// Select active, connected integrations that are due for a retry or regular sync
	var integrations []struct {
		ID             int64  `db:"id"`
		OrgID          int64  `db:"org_id"`
		CarrierSCAC    string `db:"carrier_scac"`
		FailedAttempts int    `db:"failed_attempts"`
	}

	err := w.db.SelectContext(ctx, &integrations, `
		SELECT id, org_id, carrier_scac, failed_attempts 
		FROM carrier_integrations
		WHERE is_active = 1
		  AND (next_retry_time IS NULL OR next_retry_time <= NOW())
		  AND connection_status = 'CONNECTED'
	`)
	if err != nil {
		log.Printf("[Carrier Sync Worker] Failed to fetch due integrations: %v", err)
		return
	}

	for _, ci := range integrations {
		log.Printf("[Carrier Sync Worker] Running scheduled sync for integration %d (Org %d, SCAC %s)", ci.ID, ci.OrgID, ci.CarrierSCAC)
		
		var syncErr error
		if w.carrierSvc != nil {
			_, syncErr = w.carrierSvc.SyncNow(ctx, ci.OrgID, ci.ID, domain.SyncNowRequest{Operation: string(domain.SyncOpFullSync)})
		}

		if syncErr != nil {
			// Calculate exponential backoff: 5m * 2^failed_attempts, max 24 hours (1440m)
			attempts := ci.FailedAttempts + 1
			multiplier := 1 << (attempts - 1)
			backoffMinutes := 5 * multiplier
			if backoffMinutes > 1440 {
				backoffMinutes = 1440
			}

			log.Printf("[Carrier Sync Worker] Scheduled sync failed for integration %d. Attempt %d. Next retry in %dm. Error: %v", ci.ID, attempts, backoffMinutes, syncErr)

			_, dbErr := w.db.ExecContext(ctx, `
				UPDATE carrier_integrations 
				SET failed_attempts = ?, 
					error_details = ?, 
					next_retry_time = DATE_ADD(NOW(), INTERVAL ? MINUTE),
					sync_status = 'Failed',
					last_failure_at = NOW(),
					last_synced_at = NOW()
				WHERE id = ?
			`, attempts, syncErr.Error(), backoffMinutes, ci.ID)
			
			if dbErr != nil {
				log.Printf("[Carrier Sync Worker] Failed to update failure status for integration %d: %v", ci.ID, dbErr)
			}
		} else {
			log.Printf("[Carrier Sync Worker] Scheduled sync completed for integration %d", ci.ID)
			
			_, dbErr := w.db.ExecContext(ctx, `
				UPDATE carrier_integrations 
				SET failed_attempts = 0, 
					error_details = NULL, 
					next_retry_time = DATE_ADD(NOW(), INTERVAL 4 HOUR),
					sync_status = 'Completed',
					last_success_at = NOW(),
					last_synced_at = NOW()
				WHERE id = ?
			`, ci.ID)

			if dbErr != nil {
				log.Printf("[Carrier Sync Worker] Failed to update success status for integration %d: %v", ci.ID, dbErr)
			}
		}
	}
}
