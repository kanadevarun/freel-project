-- Migration: 052_quotation_pricing.sql
-- Description: Create quotation_charge_items table for line-item pricing and charges, and add margin summary fields to quotations table.

CREATE TABLE IF NOT EXISTS quotation_charge_items (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,

    charge_code VARCHAR(50) NOT NULL,
    charge_name VARCHAR(255) NOT NULL,
    charge_category VARCHAR(50) NOT NULL DEFAULT 'OTHER',
    charge_type VARCHAR(20) NOT NULL DEFAULT 'SELL',
    calculation_basis VARCHAR(50) NOT NULL DEFAULT 'FLAT',

    quantity DECIMAL(15,4) NOT NULL DEFAULT 1.0000,
    unit_price DECIMAL(15,2) NOT NULL DEFAULT 0.00,

    cost_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    sell_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,

    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    exchange_rate DECIMAL(15,6) NOT NULL DEFAULT 1.000000,

    tax_rate DECIMAL(8,4) NOT NULL DEFAULT 0.0000,
    tax_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,

    discount_type VARCHAR(20) NOT NULL DEFAULT 'NONE',
    discount_value DECIMAL(15,4) NOT NULL DEFAULT 0.0000,
    discount_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,

    total_cost DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    total_sell DECIMAL(15,2) NOT NULL DEFAULT 0.00,

    display_order INT NOT NULL DEFAULT 0,
    is_optional BOOLEAN NOT NULL DEFAULT FALSE,
    notes TEXT NULL,

    created_by VARCHAR(255) NULL,
    updated_by VARCHAR(255) NULL,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_qcharges_org_quote (org_id, quotation_id),
    INDEX idx_qcharges_org_quote_order (org_id, quotation_id, display_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Extend quotations table with financial margin summary fields
ALTER TABLE quotations
    ADD COLUMN total_cost DECIMAL(15,2) NOT NULL DEFAULT 0.00 AFTER total_amount,
    ADD COLUMN gross_profit DECIMAL(15,2) NOT NULL DEFAULT 0.00 AFTER total_cost,
    ADD COLUMN gross_margin_pct DECIMAL(8,4) NOT NULL DEFAULT 0.0000 AFTER gross_profit;

