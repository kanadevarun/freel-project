-- +migrate Up
-- ─────────────────────────────────────────────────────────────────────────────
-- 042: BOOKINGS TABLE & SHIPMENT HANDOFF RELATIONSHIP
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS bookings (
    id                   BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id               BIGINT NOT NULL,
    rfq_id               BIGINT NOT NULL,
    quote_id             BIGINT NULL,
    booking_number       VARCHAR(100) NOT NULL,
    carrier_id           VARCHAR(50) NULL,
    carrier_name         VARCHAR(255) NOT NULL,
    carrier_scac         VARCHAR(10) NULL,
    status               VARCHAR(50) NOT NULL DEFAULT 'DRAFT',
    origin_port          VARCHAR(100) NOT NULL,
    destination_port     VARCHAR(100) NOT NULL,
    vessel_name          VARCHAR(255) NULL,
    voyage_number        VARCHAR(100) NULL,
    etd                  DATETIME NULL,
    eta                  DATETIME NULL,
    cargo_summary        TEXT NULL,
    special_instructions TEXT NULL,
    created_by           VARCHAR(255) NULL,
    created_at           DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at           DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_bookings_org (org_id),
    INDEX idx_bookings_rfq (rfq_id),
    INDEX idx_bookings_quote (quote_id),
    INDEX idx_bookings_number (booking_number),
    INDEX idx_bookings_status (status),
    CONSTRAINT fk_bookings_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_bookings_rfq FOREIGN KEY (rfq_id) REFERENCES rfqs(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE shipments ADD COLUMN booking_id BIGINT NULL AFTER quote_id;
ALTER TABLE shipments ADD INDEX idx_shipments_booking_id (booking_id);

-- +migrate Down
ALTER TABLE shipments DROP COLUMN booking_id;
DROP TABLE IF EXISTS bookings;
