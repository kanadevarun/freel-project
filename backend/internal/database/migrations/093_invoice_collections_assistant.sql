-- +migrate Up
ALTER TABLE ai_recommendations 
ADD COLUMN IF NOT EXISTS invoice_id BIGINT NULL AFTER quotation_id;

CREATE INDEX IF NOT EXISTS idx_ai_rec_invoice ON ai_recommendations(org_id, invoice_id);
CREATE INDEX IF NOT EXISTS idx_ai_rec_finance_lookup ON ai_recommendations(org_id, category, invoice_id, customer_id, status);

-- +migrate Down
DROP INDEX IF EXISTS idx_ai_rec_finance_lookup ON ai_recommendations;
DROP INDEX IF EXISTS idx_ai_rec_invoice ON ai_recommendations;

ALTER TABLE ai_recommendations
DROP COLUMN IF EXISTS invoice_id;
