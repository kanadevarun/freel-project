-- ═══════════════════════════════════════════════════════════════════════════
-- Migration 059: Rate Charges and Commercial Pricing Breakdown
-- Task 19.2 — Itemized rate charges, calculation bases, and pricing breakdown
-- ═══════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS rate_charge_items (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    rate_id BIGINT NOT NULL,
    charge_category VARCHAR(50) NOT NULL DEFAULT 'FREIGHT',
    charge_code VARCHAR(50) NOT NULL DEFAULT '',
    charge_name VARCHAR(150) NOT NULL,
    calculation_basis VARCHAR(50) NOT NULL DEFAULT 'FLAT',
    quantity DECIMAL(12,4) NOT NULL DEFAULT 1.0000,
    unit_price DECIMAL(14,4) NOT NULL DEFAULT 0.0000,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    minimum_amount DECIMAL(14,4) NULL,
    maximum_amount DECIMAL(14,4) NULL,
    included_in_base_rate BOOLEAN NOT NULL DEFAULT FALSE,
    display_order INT NOT NULL DEFAULT 0,
    notes TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_rci_org_rate (org_id, rate_id),
    INDEX idx_rci_org_rate_order (org_id, rate_id, display_order),
    INDEX idx_rci_org_category (org_id, charge_category),
    INDEX idx_rci_created_at (created_at),

    CONSTRAINT fk_rci_rate FOREIGN KEY (rate_id) REFERENCES rates(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
