-- Migration: 013c_sales_crm.sql
-- Create lead_interactions table for tracking raw email communication and thread context.

CREATE TABLE IF NOT EXISTS lead_interactions (
    id            BIGSERIAL PRIMARY KEY,
    org_id        BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    lead_id       BIGINT NOT NULL REFERENCES leads(id) ON DELETE CASCADE,
    channel       VARCHAR(50) NOT NULL,   -- EMAIL | PHONE | LINKEDIN | WHATSAPP
    direction     VARCHAR(10) NOT NULL,   -- INBOUND | OUTBOUND
    subject       VARCHAR(500),
    content       TEXT NOT NULL,
    raw_email_id  VARCHAR(255),            -- External email message-id for threading
    thread_id     VARCHAR(255),            -- Conversation thread identifier
    sentiment     VARCHAR(50),             -- POSITIVE | NEUTRAL | NEGATIVE
    intent        VARCHAR(50),             -- RFQ_REQUEST | QUESTION | MEETING | UNSUBSCRIBE | FOLLOW_UP
    linked_rfq_id BIGINT REFERENCES rfqs(id) ON DELETE SET NULL,
    ai_confidence INT DEFAULT 0,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_lead_interactions_lead ON lead_interactions(lead_id);
CREATE INDEX IF NOT EXISTS idx_lead_interactions_thread ON lead_interactions(thread_id);
