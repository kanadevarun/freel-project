-- 085_unified_ai_approvals.sql
-- Unified AI Human Approval, Confirmation, and Resume Handling

ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS actor_type VARCHAR(50) NOT NULL DEFAULT 'USER';
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS source VARCHAR(100) NULL;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS action_name VARCHAR(100) NULL;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS risk_level VARCHAR(50) NOT NULL DEFAULT 'MEDIUM';
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS required_permission VARCHAR(100) NULL;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS ai_task_id BIGINT NULL;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS thread_id VARCHAR(100) NULL;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS checkpoint_id VARCHAR(100) NULL;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS proposed_payload LONGTEXT NULL;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS approval_reference VARCHAR(100) NULL;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS correlation_id VARCHAR(100) NULL;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(255) NULL;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS expires_at DATETIME NULL;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS cancelled_by VARCHAR(100) NULL;
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS cancelled_at DATETIME NULL;

CREATE INDEX IF NOT EXISTS idx_ar_approval_ref ON approval_requests (approval_reference);
CREATE INDEX IF NOT EXISTS idx_ar_thread ON approval_requests (thread_id);
CREATE INDEX IF NOT EXISTS idx_ar_action ON approval_requests (action_name);
CREATE INDEX IF NOT EXISTS idx_ar_expires ON approval_requests (expires_at);
