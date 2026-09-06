-- Migration: 026_mailbox_settings.sql
-- Description: Add sync_frequency and processing_enabled to org_connected_mailboxes

ALTER TABLE org_connected_mailboxes
ADD COLUMN sync_frequency VARCHAR(50) DEFAULT 'Real-time',
ADD COLUMN processing_enabled BOOLEAN DEFAULT TRUE;
