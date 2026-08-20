-- +migrate Up

-- ─────────────────────────────────────────────────────────────────────────────
-- CARRIERS — master carrier registry
-- Stores SCAC codes, display names, aliases, and integration status.
-- The carriers table is the single source of truth for carrier identity
-- used by both the Spot Rate and Contract Rate engines.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE carriers (
    scac        VARCHAR(10) PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    aliases     TEXT[]       NOT NULL DEFAULT '{}',
    logo_url    TEXT,
    website     TEXT,
    api_enabled BOOLEAN      NOT NULL DEFAULT FALSE,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO carriers (scac, name, aliases, api_enabled) VALUES
    ('MAEU', 'Maersk',       ARRAY['Maersk Line', 'AP Moller Maersk', 'A.P. Moller'], TRUE),
    ('MSCU', 'MSC',          ARRAY['Mediterranean Shipping Company', 'MSC Line'],     TRUE),
    ('CMDU', 'CMA CGM',      ARRAY['CMA-CGM', 'CMA CGM Group'],                      FALSE),
    ('ONEY', 'ONE',          ARRAY['Ocean Network Express', 'ONE Line'],              TRUE),
    ('HLCU', 'Hapag-Lloyd',  ARRAY['Hapag Lloyd', 'H-L', 'HapagLloyd'],              FALSE),
    ('EGLV', 'Evergreen',    ARRAY['Evergreen Line', 'Evergreen Marine'],            FALSE),
    ('COSU', 'COSCO',        ARRAY['COSCO Shipping', 'China Ocean Shipping'],        FALSE),
    ('ZIMU', 'ZIM',          ARRAY['ZIM Integrated Shipping', 'ZIM Line'],           FALSE),
    ('YMLU', 'Yang Ming',    ARRAY['Yang Ming Marine'],                              FALSE),
    ('HDMU', 'HMM',          ARRAY['Hyundai Merchant Marine', 'Hyundai'],            FALSE);

-- ─────────────────────────────────────────────────────────────────────────────
-- RATE_ENTRIES — the canonical unified rate store
--
-- This is the single table the Quotation Engine reads from, regardless of
-- whether the rate came from a live carrier API (source=SPOT_API) or an
-- uploaded contract PDF (source=CONTRACT_PDF).
--
-- Key design rules:
--   1. All monetary values stored in USD (normalized at ingestion time).
--   2. All port codes stored as UN/LOCODE (normalized at ingestion time).
--   3. Only rates with extraction_status='CONFIRMED' are served to the
--      Quotation Engine. Unvalidated rows are invisible to it.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE rate_entries (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      BIGINT      NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Source tracking
    source          VARCHAR(30)  NOT NULL,
    -- SPOT_API | CONTRACT_PDF | MANUAL | EMAIL
    source_ref      VARCHAR(255),
    -- "mock-provider" | "maersk-api-v2" | <contract_document UUID>

    -- Route — always stored as UN/LOCODE (5-char)
    origin_port         VARCHAR(10)  NOT NULL,
    destination_port    VARCHAR(10)  NOT NULL,
    via_port            VARCHAR(10),
    service_code        VARCHAR(30),

    -- Carrier
    carrier_scac    VARCHAR(10)  NOT NULL REFERENCES carriers(scac),
    carrier_name    VARCHAR(100) NOT NULL,
    vessel_name     VARCHAR(200),

    -- Equipment
    equipment_type  VARCHAR(10)  NOT NULL DEFAULT '40GP',
    -- 20GP | 40GP | 40HC | 45HC | REEFER

    -- Pricing — all in USD
    ocean_freight           NUMERIC(12,2) NOT NULL,
    origin_charges          NUMERIC(12,2) NOT NULL DEFAULT 0,
    destination_charges     NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_buy_price         NUMERIC(12,2) NOT NULL,
    -- total_buy_price = ocean_freight + origin_charges + destination_charges
    --                 + all included surcharges
    currency_original   VARCHAR(5)    NOT NULL DEFAULT 'USD',
    exchange_rate_used  NUMERIC(10,6) NOT NULL DEFAULT 1.0,

    -- Surcharges — normalized JSONB array of Surcharge objects
    -- Schema: [{"code":"BAF","description":"...","amount":350.00,"unit":"PER_TEU","included":true}]
    surcharges  JSONB NOT NULL DEFAULT '[]',

    -- Included / excluded charge codes (for display + RFQ comparison)
    included_charges    TEXT[] NOT NULL DEFAULT '{}',
    excluded_charges    TEXT[] NOT NULL DEFAULT '{}',

    -- Cargo / service conditions
    free_days_origin        INT  NOT NULL DEFAULT 0,
    free_days_destination   INT  NOT NULL DEFAULT 14,
    transit_days            INT,
    incoterms               VARCHAR(10),
    commodity_restrictions  TEXT[] NOT NULL DEFAULT '{}',
    routing_conditions      TEXT,

    -- Validity window
    valid_from      DATE NOT NULL,
    valid_until     DATE NOT NULL,

    -- Data quality
    confidence_score    SMALLINT    NOT NULL DEFAULT 100,
    -- 0-100; API data = 100, AI-extracted = 0-99
    extraction_status   VARCHAR(30) NOT NULL DEFAULT 'CONFIRMED',
    -- CONFIRMED | PENDING_REVIEW | FLAGGED | REJECTED
    extracted_by        VARCHAR(100),
    -- "spot-api" | "agent:contract-reader" | "human:{user_id}"
    review_flags        TEXT[]      NOT NULL DEFAULT '{}',
    -- e.g., ["PRICE_ANOMALY", "PORT_UNKNOWN"]
    reviewed_by         BIGINT      REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at         TIMESTAMPTZ,

    -- Operational metadata
    nautical_miles  INT,
    co2_per_teu     NUMERIC(8,3),

    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Primary search path: org + route + equipment + validity.
-- This index is hit on every call to rates.Service.SearchRates().
CREATE INDEX idx_rate_entries_route
    ON rate_entries (org_id, origin_port, destination_port, equipment_type);

CREATE INDEX idx_rate_entries_validity
    ON rate_entries (valid_from, valid_until);

CREATE INDEX idx_rate_entries_carrier
    ON rate_entries (carrier_scac);

CREATE INDEX idx_rate_entries_status
    ON rate_entries (extraction_status);

CREATE INDEX idx_rate_entries_source
    ON rate_entries (source);

-- GIN indexes for JSONB + array columns used in charge filtering
CREATE INDEX idx_rate_entries_surcharges
    ON rate_entries USING GIN (surcharges);

CREATE INDEX idx_rate_entries_included
    ON rate_entries USING GIN (included_charges);

-- +migrate Down

DROP INDEX IF EXISTS idx_rate_entries_included;
DROP INDEX IF EXISTS idx_rate_entries_surcharges;
DROP INDEX IF EXISTS idx_rate_entries_source;
DROP INDEX IF EXISTS idx_rate_entries_status;
DROP INDEX IF EXISTS idx_rate_entries_carrier;
DROP INDEX IF EXISTS idx_rate_entries_validity;
DROP INDEX IF EXISTS idx_rate_entries_route;
DROP TABLE IF EXISTS rate_entries;
DROP TABLE IF EXISTS carriers;
