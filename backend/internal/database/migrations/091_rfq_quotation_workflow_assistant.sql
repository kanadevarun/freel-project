-- Migration 091: RFQ and Quotation Workflow Assistant
-- Extends ai_recommendations with explicit rfq_id and quotation_id fields for direct indexed linkage,
-- plus indexes for RFQ and quotation recommendation queries.

ALTER TABLE ai_recommendations 
    ADD COLUMN IF NOT EXISTS rfq_id BIGINT NULL,
    ADD COLUMN IF NOT EXISTS quotation_id BIGINT NULL;

CREATE INDEX IF NOT EXISTS idx_ai_rec_rfq ON ai_recommendations (org_id, rfq_id);
CREATE INDEX IF NOT EXISTS idx_ai_rec_quote ON ai_recommendations (org_id, quotation_id);
