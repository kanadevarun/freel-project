-- +migrate Up

-- ─────────────────────────────────────────────────────────────────────────────
-- SHIPMENTS
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS shipments (
    id                BIGSERIAL PRIMARY KEY,
    org_id            BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    rfq_id            BIGINT UNIQUE REFERENCES rfqs(id) ON DELETE SET NULL,
    quote_id          BIGINT REFERENCES rfq_quotes(id) ON DELETE SET NULL,
    carrier_scac      VARCHAR(10) NOT NULL REFERENCES carriers(scac),
    booking_number    VARCHAR(100),
    mbl_number        VARCHAR(100),
    hbl_number        VARCHAR(100),
    container_numbers TEXT[] DEFAULT '{}',
    status            VARCHAR(50) NOT NULL DEFAULT 'BOOKING_PENDING',
    origin_port       VARCHAR(10) NOT NULL,
    destination_port  VARCHAR(10) NOT NULL,
    vessel_name       VARCHAR(200),
    voyage_number     VARCHAR(100),
    etd               TIMESTAMPTZ,
    eta               TIMESTAMPTZ,
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);

-- ─────────────────────────────────────────────────────────────────────────────
-- SHIPMENT MILESTONES
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS shipment_milestones (
    id              BIGSERIAL PRIMARY KEY,
    shipment_id     BIGINT NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    milestone_code  VARCHAR(50) NOT NULL, -- BOOKED, DEPARTED, IN_TRANSIT, ARRIVED, DELIVERED
    description     VARCHAR(255),
    planned_date    TIMESTAMPTZ,
    actual_date     TIMESTAMPTZ,
    status          VARCHAR(30) NOT NULL DEFAULT 'PLANNED',
    location        VARCHAR(100),
    notes           TEXT,
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ─────────────────────────────────────────────────────────────────────────────
-- SHIPMENT EXCEPTIONS
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS shipment_exceptions (
    id              BIGSERIAL PRIMARY KEY,
    shipment_id     BIGINT NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    exception_type  VARCHAR(50) NOT NULL, -- ROLLOVER, DELAY, CUSTOMS_HOLD, PORT_CONGESTION, WEATHER
    severity        VARCHAR(20) NOT NULL, -- INFO, WARNING, CRITICAL
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    resolved        BOOLEAN NOT NULL DEFAULT FALSE,
    resolved_at     TIMESTAMPTZ,
    ai_summary      TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ─────────────────────────────────────────────────────────────────────────────
-- SHIPMENT PROCESSED EVENTS (Idempotency / Duplication Protection)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS shipment_processed_events (
    event_id    VARCHAR(255) PRIMARY KEY,
    shipment_id BIGINT REFERENCES shipments(id) ON DELETE CASCADE,
    processed_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_shipments_org ON shipments(org_id);
CREATE INDEX idx_shipments_status ON shipments(status);
CREATE INDEX idx_shipments_booking ON shipments(booking_number);
CREATE INDEX idx_shipments_mbl ON shipments(mbl_number);
CREATE INDEX idx_shipments_hbl ON shipments(hbl_number);
CREATE INDEX idx_milestones_shipment ON shipment_milestones(shipment_id);
CREATE INDEX idx_exceptions_shipment ON shipment_exceptions(shipment_id);
