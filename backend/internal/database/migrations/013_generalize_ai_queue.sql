-- +migrate Up
ALTER TABLE ai_processing_tasks
    ALTER COLUMN document_id DROP NOT NULL;

ALTER TABLE ai_processing_tasks
    ADD COLUMN IF NOT EXISTS entity_type VARCHAR(50) DEFAULT 'CONTRACT',
    ADD COLUMN IF NOT EXISTS entity_id VARCHAR(255);

-- Backfill existing tasks with entity_id of document_id
UPDATE ai_processing_tasks
SET entity_id = document_id::text, entity_type = 'CONTRACT'
WHERE document_id IS NOT NULL AND entity_id IS NULL;

-- +migrate Down
ALTER TABLE ai_processing_tasks
    DROP COLUMN IF EXISTS entity_type,
    DROP COLUMN IF EXISTS entity_id;

ALTER TABLE ai_processing_tasks
    ALTER COLUMN document_id SET NOT NULL;
