-- 056_quotation_booking_conversion.sql
-- Quotation-to-Booking Operational Conversion & Commercial Handover (Task 18.6)

-- 1. Extend quotations table with conversion tracking and references
ALTER TABLE quotations
    ADD COLUMN IF NOT EXISTS converted_at DATETIME NULL AFTER cancelled_reason,
    ADD COLUMN IF NOT EXISTS converted_by VARCHAR(255) NULL AFTER converted_at,
    ADD COLUMN IF NOT EXISTS converted_booking_id BIGINT NULL AFTER converted_by,
    ADD COLUMN IF NOT EXISTS converted_shipment_id BIGINT NULL AFTER converted_booking_id,
    ADD COLUMN IF NOT EXISTS conversion_status VARCHAR(50) NOT NULL DEFAULT 'NOT_CONVERTED' AFTER converted_shipment_id,
    ADD COLUMN IF NOT EXISTS conversion_notes TEXT NULL AFTER conversion_status;

-- Add indexes for efficient conversion queries and lookups
ALTER TABLE quotations
    ADD INDEX idx_quotation_org_booking (org_id, converted_booking_id),
    ADD INDEX idx_quotation_org_shipment (org_id, converted_shipment_id),
    ADD INDEX idx_quotation_org_conv_status (org_id, conversion_status);

-- 2. Create immutable quotation conversion history table
CREATE TABLE IF NOT EXISTS quotation_conversion_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,
    booking_id BIGINT NULL,
    shipment_id BIGINT NULL,
    action VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    message TEXT NULL,
    performed_by VARCHAR(255) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_qch_org_quote_created (org_id, quotation_id, created_at DESC),
    INDEX idx_qch_org_booking (org_id, booking_id),
    CONSTRAINT fk_quotation_conversion_history_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_quotation_conversion_history_quote FOREIGN KEY (quotation_id) REFERENCES quotations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
