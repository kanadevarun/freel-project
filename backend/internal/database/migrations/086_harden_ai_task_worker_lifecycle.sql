-- 086_harden_ai_task_worker_lifecycle.sql
-- Harden Unified AI Task Queue, Worker Lifecycle, Retries, and Status Synchronization

ALTER TABLE ai_processing_tasks ADD COLUMN IF NOT EXISTS worker_id VARCHAR(100) NULL;
ALTER TABLE ai_processing_tasks ADD COLUMN IF NOT EXISTS lease_expires_at DATETIME NULL;
ALTER TABLE ai_processing_tasks ADD COLUMN IF NOT EXISTS heartbeat_at DATETIME NULL;
ALTER TABLE ai_processing_tasks ADD COLUMN IF NOT EXISTS started_at DATETIME NULL;
ALTER TABLE ai_processing_tasks ADD COLUMN IF NOT EXISTS completed_at DATETIME NULL;
ALTER TABLE ai_processing_tasks ADD COLUMN IF NOT EXISTS available_at DATETIME NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE ai_processing_tasks ADD COLUMN IF NOT EXISTS max_retries INT NOT NULL DEFAULT 3;
ALTER TABLE ai_processing_tasks ADD COLUMN IF NOT EXISTS thread_id VARCHAR(150) NULL;
ALTER TABLE ai_processing_tasks ADD COLUMN IF NOT EXISTS correlation_id VARCHAR(100) NULL;
ALTER TABLE ai_processing_tasks ADD COLUMN IF NOT EXISTS approval_id BIGINT NULL;
ALTER TABLE ai_processing_tasks ADD COLUMN IF NOT EXISTS acting_user_id BIGINT NULL;
ALTER TABLE ai_processing_tasks ADD COLUMN IF NOT EXISTS actor_type VARCHAR(50) NOT NULL DEFAULT 'AI_AGENT';
ALTER TABLE ai_processing_tasks ADD COLUMN IF NOT EXISTS last_error_code VARCHAR(100) NULL;

CREATE INDEX IF NOT EXISTS idx_apt_status_avail ON ai_processing_tasks (status, available_at);
CREATE INDEX IF NOT EXISTS idx_apt_status_lease ON ai_processing_tasks (status, lease_expires_at);
CREATE INDEX IF NOT EXISTS idx_apt_org_status ON ai_processing_tasks (org_id, status);
CREATE INDEX IF NOT EXISTS idx_apt_org_thread ON ai_processing_tasks (org_id, thread_id);
CREATE INDEX IF NOT EXISTS idx_apt_approval ON ai_processing_tasks (approval_id);
