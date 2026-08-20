-- +migrate Up

-- ─────────────────────────────────────────────────────────────────────────────
-- AI_PROCESSING_TASKS
--
-- Simple meaning:
--   This table stores background tasks for the multi-agent AI.
--   Instead of the Go backend sending direct web requests to the Python sidecar
--   (which can fail or timeout if the server restarts), Go writes a task row here.
--   The Python worker continuously polls this table, processes the tasks,
--   and updates their status.
--
-- Example data:
--   task_type = 'PROCESS'
--   payload = {"document_id": "3ae5c3ab...", "s3_key": "raw/contracts/contract.pdf", "callback_url": "..."}
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS ai_processing_tasks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    document_id     UUID NOT NULL REFERENCES contract_documents(id) ON DELETE CASCADE,
    task_type       VARCHAR(50) NOT NULL, -- e.g., 'PROCESS' (new upload) or 'RESUME' (human approval)
    payload         JSONB NOT NULL,       -- Extra parameters (s3_key, notes, callback_url, corrected_rates)
    status          VARCHAR(20) NOT NULL DEFAULT 'QUEUED', -- QUEUED | PROCESSING | COMPLETED | FAILED
    attempts        INT NOT NULL DEFAULT 0, -- Track retries
    max_attempts    INT NOT NULL DEFAULT 3, -- Max retries before flagging as failed
    last_error      TEXT,                  -- Stores panic traceback or HTTP failure logs
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index the status field to keep the queue polling query fast.
CREATE INDEX IF NOT EXISTS idx_ai_tasks_status ON ai_processing_tasks (status);

-- +migrate Down
DROP TABLE IF EXISTS ai_processing_tasks;
