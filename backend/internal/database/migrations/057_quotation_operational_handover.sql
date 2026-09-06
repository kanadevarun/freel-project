-- 057_quotation_operational_handover.sql
-- Booking Confirmation, Commercial Handover & Quote-to-Operations Traceability (Task 18.7)

-- 1. Extend bookings table with commercial handover lineage fields
ALTER TABLE bookings
    ADD COLUMN IF NOT EXISTS source_quotation_id BIGINT NULL AFTER quote_id,
    ADD COLUMN IF NOT EXISTS source_quote_number VARCHAR(100) NULL AFTER source_quotation_id,
    ADD COLUMN IF NOT EXISTS commercial_snapshot_at DATETIME NULL AFTER source_quote_number,
    ADD COLUMN IF NOT EXISTS commercial_handover_status VARCHAR(50) NOT NULL DEFAULT 'PENDING' AFTER commercial_snapshot_at;

-- Add indexes for booking quotation lineage
ALTER TABLE bookings
    ADD INDEX idx_bookings_org_source_quote (org_id, source_quotation_id),
    ADD INDEX idx_bookings_org_handover_status (org_id, commercial_handover_status);

-- 2. Extend shipments table with quote & booking lineage references
ALTER TABLE shipments
    ADD COLUMN IF NOT EXISTS source_quotation_id BIGINT NULL AFTER booking_id,
    ADD COLUMN IF NOT EXISTS source_booking_id BIGINT NULL AFTER source_quotation_id;

ALTER TABLE shipments
    ADD INDEX idx_shipments_org_source_quote (org_id, source_quotation_id),
    ADD INDEX idx_shipments_org_source_booking (org_id, source_booking_id);

-- 3. Create quotation_operational_handover_history table
CREATE TABLE IF NOT EXISTS quotation_operational_handover_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,
    booking_id BIGINT NULL,
    shipment_id BIGINT NULL,
    event_type VARCHAR(64) NOT NULL,
    description TEXT NOT NULL,
    metadata JSON NULL,
    performed_by VARCHAR(255) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_qohh_org_quote (org_id, quotation_id, created_at DESC),
    INDEX idx_qohh_org_booking (org_id, booking_id),
    INDEX idx_qohh_org_shipment (org_id, shipment_id),
    CONSTRAINT fk_qohh_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_qohh_quote FOREIGN KEY (quotation_id) REFERENCES quotations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
