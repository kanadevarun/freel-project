package jobs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/freel/backend/internal/common/crypto"
	"github.com/freel/backend/internal/config"
	"github.com/freel/backend/internal/leads"
	"github.com/freel/backend/internal/organization"
	"github.com/jmoiron/sqlx"
)

type MailboxSyncWorker struct {
	db          *sqlx.DB
	leadsBL     leads.BusinessLogic
	orgRepo     organization.Repository
	stopCh      chan struct{}
	activeSyncs map[int64]bool
	syncMutex   sync.Mutex
}

func NewMailboxSyncWorker(db *sqlx.DB, leadsBL leads.BusinessLogic, orgRepo organization.Repository) *MailboxSyncWorker {
	return &MailboxSyncWorker{
		db:          db,
		leadsBL:     leadsBL,
		orgRepo:     orgRepo,
		stopCh:      make(chan struct{}),
		activeSyncs: make(map[int64]bool),
	}
}

func (w *MailboxSyncWorker) Start() {
	log.Println("[Mailbox Sync Worker] Starting robust background sync loop...")
	go w.runLoop()
}

func (w *MailboxSyncWorker) Stop() {
	close(w.stopCh)
}

func (w *MailboxSyncWorker) runLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.syncAllActiveMailboxes()
		case <-w.stopCh:
			return
		}
	}
}

func (w *MailboxSyncWorker) syncAllActiveMailboxes() {
	ctx := context.Background()

	// 1. Fetch active Gmail mailboxes from DB
	var mailboxes []organization.ConnectedMailbox
	query := `
		SELECT * FROM org_connected_mailboxes 
		WHERE status != 'DISCONNECTED' 
		  AND provider = 'GMAIL' 
		  AND processing_enabled = 1
	`
	err := w.db.SelectContext(ctx, &mailboxes, query)
	if err != nil {
		log.Printf("[Mailbox Sync Worker] Failed to fetch active mailboxes: %v", err)
		return
	}

	for _, mb := range mailboxes {
		m := mb // copy
		// 2. Prevent concurrent sync of the same mailbox
		w.syncMutex.Lock()
		if w.activeSyncs[m.ID] {
			w.syncMutex.Unlock()
			log.Printf("[Mailbox Sync Worker] Mailbox ID %d (%s) is already syncing, skipping.", m.ID, m.Email)
			continue
		}
		w.activeSyncs[m.ID] = true
		w.syncMutex.Unlock()

		// Run sync in background goroutine so one mailbox error doesn't block another
		go func(mailbox organization.ConnectedMailbox) {
			defer func() {
				w.syncMutex.Lock()
				delete(w.activeSyncs, mailbox.ID)
				w.syncMutex.Unlock()
			}()

			log.Printf("[Mailbox Sync Worker] Starting background sync for %s", mailbox.Email)
			if err := w.SyncMailbox(context.Background(), &mailbox); err != nil {
				log.Printf("[Mailbox Sync Worker] Sync failed for %s: %v", mailbox.Email, err)
			} else {
				log.Printf("[Mailbox Sync Worker] Sync completed successfully for %s", mailbox.Email)
			}
		}(m)
	}
}

// SyncMailboxNow triggers synchronous sync for one mailbox with lock protection and isolation.
func (w *MailboxSyncWorker) SyncMailboxNow(ctx context.Context, id int64, orgID int64) error {
	// Fetch mailbox
	mb, err := w.orgRepo.GetConnectedMailboxByID(ctx, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to fetch mailbox: %w", err)
	}
	if mb.Status == "DISCONNECTED" {
		return errors.New("cannot sync a disconnected mailbox")
	}

	// Prevent concurrent sync
	w.syncMutex.Lock()
	if w.activeSyncs[mb.ID] {
		w.syncMutex.Unlock()
		return errors.New("mailbox is already undergoing synchronization")
	}
	w.activeSyncs[mb.ID] = true
	w.syncMutex.Unlock()

	defer func() {
		w.syncMutex.Lock()
		delete(w.activeSyncs, mb.ID)
		w.syncMutex.Unlock()
	}()

	return w.SyncMailbox(ctx, mb)
}

func (w *MailboxSyncWorker) SyncMailbox(ctx context.Context, mb *organization.ConnectedMailbox) error {
	// Update status in DB to SYNCING
	err := w.orgRepo.UpdateMailboxStatus(ctx, mb.ID, mb.OrgID, "SYNCING", false)
	if err != nil {
		return fmt.Errorf("failed to update status to SYNCING: %w", err)
	}

	// Update last_sync_started_at
	_, _ = w.db.ExecContext(ctx, "UPDATE org_connected_mailboxes SET last_sync_started_at = NOW() WHERE id = ?", mb.ID)

	provider := organization.NewGmailProvider()

	// Ingestion Handler callback
	ingestHandler := func(c context.Context, email organization.SyncEmail) error {
		inbound := leads.InboundEmail{
			MailboxID:        mb.ID,
			RawEmailID:       email.RawEmailID,
			RFCMessageID:     email.RFCMessageID,
			ThreadID:         email.ThreadID,
			From:             email.From,
			To:               email.To,
			Subject:          email.Subject,
			Body:             email.Body,
			MessageID:        email.MessageID,
			InReplyTo:        email.InReplyTo,
			ReferencesHeader: email.ReferencesHeader,
			Sender:           email.Sender,
			Recipients:       email.Recipients,
			CCRecipients:     email.CCRecipients,
			ReceivedAt:       email.ReceivedAt,
		}
		// Ingest using the leads module's shared ingestion method
		_, err := w.leadsBL.ProcessInboundEmail(c, int32(mb.OrgID), inbound)
		return err
	}

	// Token refresh persistence callback
	tokenUpdate := func(c context.Context, accessToken string, expiry time.Time) error {
		encKey := os.Getenv(config.EnvMailboxEncryptionKey)
		encrypted, err := crypto.Encrypt(accessToken, encKey)
		if err != nil {
			return err
		}
		_, err = w.db.ExecContext(c, `
			UPDATE org_connected_mailboxes 
			SET access_token_encrypted = ?, token_expiry = ?, updated_at = NOW() 
			WHERE id = ?
		`, encrypted, expiry, mb.ID)
		return err
	}

	// Perform actual Gmail API synchronization
	syncErr := provider.SyncMailbox(ctx, mb, ingestHandler, tokenUpdate)

	// Save sync cursor, errors, timestamps, and set status back to CONNECTED
	status := "CONNECTED"
	var errStr *string
	if syncErr != nil {
		status = "ERROR"
		msg := syncErr.Error()
		errStr = &msg
	}

	var cursorStr *string
	if mb.SyncCursor != nil {
		cursorStr = mb.SyncCursor
	}

	// Update DB record
	var query string
	if status == "CONNECTED" {
		query = `
			UPDATE org_connected_mailboxes 
			SET status = ?, 
			    sync_cursor = ?, 
			    last_sync_error = NULL, 
			    last_sync_success_at = NOW(), 
			    last_synced_at = NOW(), 
			    updated_at = NOW() 
			WHERE id = ?
		`
		_, err = w.db.ExecContext(ctx, query, status, cursorStr, mb.ID)
	} else {
		query = `
			UPDATE org_connected_mailboxes 
			SET status = ?, 
			    sync_cursor = ?, 
			    last_sync_error = ?, 
			    updated_at = NOW() 
			WHERE id = ?
		`
		_, err = w.db.ExecContext(ctx, query, status, cursorStr, errStr, mb.ID)
	}

	if err != nil {
		return fmt.Errorf("failed to save sync finalization: %w (original error: %v)", err, syncErr)
	}

	return syncErr
}
