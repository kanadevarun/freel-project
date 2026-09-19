-- Migration 094: Contract, Document, and Compliance Workflow Assistant
-- Adds contract_id, document_id, compliance_id, carrier_id to ai_recommendations with strategic indexes

ALTER TABLE ai_recommendations
    ADD COLUMN IF NOT EXISTS contract_id BIGINT NULL,
    ADD COLUMN IF NOT EXISTS document_id VARCHAR(64) NULL,
    ADD COLUMN IF NOT EXISTS compliance_id BIGINT NULL,
    ADD COLUMN IF NOT EXISTS carrier_id BIGINT NULL;

-- Strategic query indexes for contract, document, and compliance assistant lookups
CREATE INDEX IF NOT EXISTS idx_ai_rec_contract ON ai_recommendations (contract_id);
CREATE INDEX IF NOT EXISTS idx_ai_rec_document ON ai_recommendations (document_id);
CREATE INDEX IF NOT EXISTS idx_ai_rec_compliance ON ai_recommendations (compliance_id);
CREATE INDEX IF NOT EXISTS idx_ai_rec_carrier ON ai_recommendations (carrier_id);
CREATE INDEX IF NOT EXISTS idx_ai_rec_contract_lookup ON ai_recommendations (org_id, category, contract_id, status);
