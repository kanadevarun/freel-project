-- +migrate Up

-- ─────────────────────────────────────────────────────────────────────────────
-- 015_operations_hardening.sql
--
-- Phase 3 Hardening: Enforces DB-level constraints, adds durable carrier event
-- store, extends the AI task queue for production-grade locking, and adds
-- source_event_id audit trail columns.
-- ─────────────────────────────────────────────────────────────────────────────


-- ─────────────────────────────────────────────────────────────────────────────
-- 1. SHIPMENTS — Status CHECK constraint
-- ─────────────────────────────────────────────────────────────────────────────
ALTER TABLE shipments
    ADD CONSTRAINT chk_shipment_status CHECK (status IN (
        'BOOKING_PENDING',
        'BOOKED',
        'DEPARTED',
        'IN_TRANSIT',
        'ARRIVED',
        'DELIVERED',
        'EXCEPTION'
    ));


-- ─────────────────────────────────────────────────────────────────────────────
-- 2. SHIPMENT_MILESTONES
--    - Unique milestone code per shipment (prevents duplicate DEPARTED etc.)
--    - Status CHECK constraint
--    - source_event_id for full audit trail back to the originating event
-- ─────────────────────────────────────────────────────────────────────────────
ALTER TABLE shipment_milestones
    ADD CONSTRAINT uq_milestone_shipment_code UNIQUE (shipment_id, milestone_code),
    ADD CONSTRAINT chk_milestone_status CHECK (status IN ('PLANNED', 'COMPLETED')),
    ADD COLUMN IF NOT EXISTS source_event_id VARCHAR(255);


-- ─────────────────────────────────────────────────────────────────────────────
-- 3. SHIPMENT_EXCEPTIONS
--    - Exception type CHECK constraint
--    - Severity CHECK constraint
--    - source_event_id for deduplication and audit
-- ─────────────────────────────────────────────────────────────────────────────
ALTER TABLE shipment_exceptions
    ADD CONSTRAINT chk_exception_type CHECK (exception_type IN (
        'ROLLOVER', 'DELAY', 'CUSTOMS_HOLD', 'PORT_CONGESTION', 'WEATHER'
    )),
    ADD CONSTRAINT chk_exception_severity CHECK (severity IN ('INFO', 'WARNING', 'CRITICAL')),
    ADD COLUMN IF NOT EXISTS source_event_id VARCHAR(255);


-- ─────────────────────────────────────────────────────────────────────────────
-- 4. CARRIER_TRACKING_EVENTS
--
--    The durable raw event store. All inbound carrier data (API, webhook,
--    email, manual) is written here FIRST before any queue insert. This means
--    raw events are never lost even if the queue worker crashes or the Python
--    sidecar is unavailable.
--
--    Lifecycle: RECEIVED → QUEUED → PROCESSING → PROCESSED | FAILED
--    Matching:  UNMATCHED → MATCHED | AMBIGUOUS
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS carrier_tracking_events (
    id                  BIGSERIAL PRIMARY KEY,
    org_id              BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Canonical event identifier (must be globally unique per carrier+event)
    event_id            VARCHAR(255) NOT NULL,

    -- Where this event came from
    source_type         VARCHAR(20) NOT NULL DEFAULT 'MANUAL', -- API|WEBHOOK|EMAIL|MANUAL|POLLING

    -- Carrier identification
    carrier_scac        VARCHAR(10),

    -- Shipment identifiers extracted from the event (may be partial)
    booking_number      VARCHAR(100),
    container_number    VARCHAR(50),
    mbl_number          VARCHAR(100),
    hbl_number          VARCHAR(100),
    vessel_name         VARCHAR(200),
    voyage_number       VARCHAR(100),

    -- Event content
    milestone_code      VARCHAR(50),
    event_time          TIMESTAMPTZ,
    location            VARCHAR(200),
    raw_description     TEXT,
    raw_payload         JSONB,        -- Full original payload (webhook JSON, email body, API response)

    -- Resolved shipment (NULL if UNMATCHED)
    shipment_id         BIGINT REFERENCES shipments(id) ON DELETE SET NULL,

    -- State machines
    matching_status     VARCHAR(30) NOT NULL DEFAULT 'UNMATCHED'
                            CHECK (matching_status IN ('UNMATCHED', 'MATCHED', 'AMBIGUOUS')),
    processing_status   VARCHAR(30) NOT NULL DEFAULT 'RECEIVED'
                            CHECK (processing_status IN (
                                'RECEIVED', 'QUEUED', 'PROCESSING', 'PROCESSED', 'FAILED'
                            )),

    received_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Deduplication: one event_id per org
    CONSTRAINT uq_carrier_event_org UNIQUE (event_id, org_id)
);

CREATE INDEX IF NOT EXISTS idx_cte_org           ON carrier_tracking_events(org_id);
CREATE INDEX IF NOT EXISTS idx_cte_carrier       ON carrier_tracking_events(carrier_scac);
CREATE INDEX IF NOT EXISTS idx_cte_booking       ON carrier_tracking_events(booking_number);
CREATE INDEX IF NOT EXISTS idx_cte_container     ON carrier_tracking_events(container_number);
CREATE INDEX IF NOT EXISTS idx_cte_mbl           ON carrier_tracking_events(mbl_number);
CREATE INDEX IF NOT EXISTS idx_cte_shipment      ON carrier_tracking_events(shipment_id);
CREATE INDEX IF NOT EXISTS idx_cte_matching      ON carrier_tracking_events(matching_status);
CREATE INDEX IF NOT EXISTS idx_cte_processing    ON carrier_tracking_events(processing_status);
CREATE INDEX IF NOT EXISTS idx_cte_received      ON carrier_tracking_events(received_at DESC);


-- ─────────────────────────────────────────────────────────────────────────────
-- 5. AI_PROCESSING_TASKS — Queue control columns
--
--    locked_by:    Worker identity (e.g., "python-sidecar-1") for distributed
--                  claiming. Prevents two workers from processing same task.
--    locked_at:    When the lock was acquired.
--    completed_at: When the task finished (success or permanent failure).
--    available_at: Enables retry with delay (task invisible until this time).
-- ─────────────────────────────────────────────────────────────────────────────
ALTER TABLE ai_processing_tasks
    ADD COLUMN IF NOT EXISTS locked_by   VARCHAR(255),
    ADD COLUMN IF NOT EXISTS locked_at   TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS available_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Composite index for efficient queue polling:
-- WHERE status = 'QUEUED' AND available_at <= NOW()
CREATE INDEX IF NOT EXISTS idx_ai_tasks_available
    ON ai_processing_tasks(available_at, status)
    WHERE status IN ('QUEUED', 'RETRYING');


-- ─────────────────────────────────────────────────────────────────────────────
-- 6. SHIPMENT_PROCESSED_EVENTS — Ensure index exists for fast lookups
--    (PRIMARY KEY on event_id already provides unique constraint)
-- ─────────────────────────────────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_spe_shipment ON shipment_processed_events(shipment_id);


-- +migrate Down

ALTER TABLE shipments DROP CONSTRAINT IF EXISTS chk_shipment_status;

ALTER TABLE shipment_milestones
    DROP CONSTRAINT IF EXISTS uq_milestone_shipment_code,
    DROP CONSTRAINT IF EXISTS chk_milestone_status,
    DROP COLUMN IF EXISTS source_event_id;

ALTER TABLE shipment_exceptions
    DROP CONSTRAINT IF EXISTS chk_exception_type,
    DROP CONSTRAINT IF EXISTS chk_exception_severity,
    DROP COLUMN IF EXISTS source_event_id;

DROP TABLE IF EXISTS carrier_tracking_events;

DROP INDEX IF EXISTS idx_ai_tasks_available;
ALTER TABLE ai_processing_tasks
    DROP COLUMN IF EXISTS locked_by,
    DROP COLUMN IF EXISTS locked_at,
    DROP COLUMN IF EXISTS completed_at,
    DROP COLUMN IF EXISTS available_at;

DROP INDEX IF EXISTS idx_spe_shipment;
