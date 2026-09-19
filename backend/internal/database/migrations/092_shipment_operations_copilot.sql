-- +migrate Up
ALTER TABLE ai_recommendations 
ADD COLUMN IF NOT EXISTS shipment_id BIGINT NULL AFTER quotation_id,
ADD COLUMN IF NOT EXISTS milestone_id BIGINT NULL AFTER shipment_id,
ADD COLUMN IF NOT EXISTS exception_id BIGINT NULL AFTER milestone_id,
ADD COLUMN IF NOT EXISTS booking_id BIGINT NULL AFTER exception_id;

CREATE INDEX IF NOT EXISTS idx_ai_rec_shipment ON ai_recommendations(org_id, shipment_id);
CREATE INDEX IF NOT EXISTS idx_ai_rec_exception ON ai_recommendations(org_id, exception_id);
CREATE INDEX IF NOT EXISTS idx_ai_rec_booking ON ai_recommendations(org_id, booking_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_ai_rec_booking ON ai_recommendations;
DROP INDEX IF EXISTS idx_ai_rec_exception ON ai_recommendations;
DROP INDEX IF EXISTS idx_ai_rec_shipment ON ai_recommendations;

ALTER TABLE ai_recommendations
DROP COLUMN IF EXISTS booking_id,
DROP COLUMN IF EXISTS exception_id,
DROP COLUMN IF EXISTS milestone_id,
DROP COLUMN IF EXISTS shipment_id;
