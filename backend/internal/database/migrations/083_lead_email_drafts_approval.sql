-- 083_lead_email_drafts_approval.sql
-- Extend lead_email_drafts with review and approval lifecycle states

ALTER TABLE lead_email_drafts
    ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'DRAFT',
    ADD COLUMN IF NOT EXISTS approval_id BIGINT NULL,
    ADD COLUMN IF NOT EXISTS sent_at DATETIME NULL,
    ADD COLUMN IF NOT EXISTS error_message TEXT NULL,
    ADD INDEX IF NOT EXISTS idx_draft_status (org_id, status),
    ADD INDEX IF NOT EXISTS idx_draft_approval (approval_id);
