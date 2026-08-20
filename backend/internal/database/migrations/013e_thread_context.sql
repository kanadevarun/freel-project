-- Migration: 013e_thread_context.sql
-- Purpose: Add thread-awareness to lead_interactions so the Sales Agent can resume
--          incomplete RFQ conversations when a customer replies on the same thread.
--
-- parent_interaction_id: FK back to the interaction this is a reply to
-- partial_rfq_context:   Structured JSON with cumulative extracted RFQ fields from the
--                        whole conversation so far — passed back to the AI agent on reply.
--                        Example: {"origin_port": "INNSA", "destination_port": "DEHAM", ...}

ALTER TABLE lead_interactions
    ADD COLUMN IF NOT EXISTS parent_interaction_id BIGINT REFERENCES lead_interactions(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS partial_rfq_context   JSONB DEFAULT '{}';

CREATE INDEX IF NOT EXISTS idx_lead_interactions_parent ON lead_interactions(parent_interaction_id);

COMMENT ON COLUMN lead_interactions.parent_interaction_id IS
    'Links a reply email to the original incomplete interaction it is responding to.';

COMMENT ON COLUMN lead_interactions.partial_rfq_context IS
    'Structured JSON snapshot of all RFQ fields extracted so far in this conversation thread.
     Used to restore AI context when the customer sends a follow-up reply.
     Keys: origin_port, destination_port, incoterms, cargo_description, cargo_weight,
           cargo_volume, target_date, lead_name.';
