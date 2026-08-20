-- +migrate Up
CREATE TABLE IF NOT EXISTS pricing_rules (
    id              BIGSERIAL PRIMARY KEY,
    org_id          BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    rule_name       VARCHAR(100) NOT NULL,
    rule_type       VARCHAR(50) NOT NULL,   -- LANE | CUSTOMER_TIER | EQUIPMENT | COMMODITY | DEFAULT
    conditions      JSONB NOT NULL DEFAULT '{}'::jsonb,
    markup_pct      DECIMAL(5,2),           -- E.g. 20.00 for 20% markup
    markup_flat     DECIMAL(12,2),          -- E.g. 150.00 for flat charge
    min_margin_pct  DECIMAL(5,2) DEFAULT 5.0,
    priority        INT DEFAULT 0,          -- Higher = applied first
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Seed some default rules for Org ID = 1 (Freel Global Logistics Pvt Ltd)
INSERT INTO pricing_rules (org_id, rule_name, rule_type, conditions, markup_pct, priority)
VALUES 
    (1, 'Default Markup Policy', 'DEFAULT', '{}'::jsonb, 20.00, 0),
    (1, 'INNSA to DEHAM Lane Promo', 'LANE', '{"origin": "INNSA", "destination": "DEHAM"}'::jsonb, 12.00, 10),
    (1, 'Enterprise Tier Discount', 'CUSTOMER_TIER', '{"tier": "ENTERPRISE"}'::jsonb, 10.00, 5)
ON CONFLICT DO NOTHING;

-- +migrate Down
DROP TABLE IF EXISTS pricing_rules;
