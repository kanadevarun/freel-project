-- Migration 053: Quotation Templates, Commercial Terms & Lineage Reference

-- 1. Create quotation_templates table
CREATE TABLE IF NOT EXISTS quotation_templates (
    id VARCHAR(36) PRIMARY KEY,
    org_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100) NULL,
    description TEXT NULL,
    transport_mode VARCHAR(50) NOT NULL DEFAULT 'OCEAN',
    service_type VARCHAR(50) NOT NULL DEFAULT 'FCL',
    direction VARCHAR(50) NOT NULL DEFAULT 'EXPORT',
    incoterm VARCHAR(20) NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    validity_days INT NOT NULL DEFAULT 30,
    payment_terms VARCHAR(100) NOT NULL DEFAULT 'PREPAID_NET_30',
    terms_and_conditions TEXT NULL,
    customer_notes_template TEXT NULL,
    internal_notes_template TEXT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by VARCHAR(36) NULL,
    updated_by VARCHAR(36) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_quote_tpl_org (org_id),
    INDEX idx_quote_tpl_org_active (org_id, is_active),
    INDEX idx_quote_tpl_mode (org_id, transport_mode, service_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. Create quotation_template_charges table
CREATE TABLE IF NOT EXISTS quotation_template_charges (
    id VARCHAR(36) PRIMARY KEY,
    template_id VARCHAR(36) NOT NULL,
    org_id VARCHAR(36) NOT NULL,
    charge_code VARCHAR(100) NOT NULL,
    charge_name VARCHAR(255) NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'FREIGHT',
    basis VARCHAR(50) NOT NULL DEFAULT 'PER_CONTAINER',
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    unit_price DECIMAL(15, 4) NOT NULL DEFAULT 0.0000,
    quantity DECIMAL(15, 4) NOT NULL DEFAULT 1.0000,
    unit_cost DECIMAL(15, 4) NOT NULL DEFAULT 0.0000,
    discount_type VARCHAR(50) NOT NULL DEFAULT 'NONE',
    discount_value DECIMAL(15, 4) NOT NULL DEFAULT 0.0000,
    tax_rate_pct DECIMAL(7, 4) NOT NULL DEFAULT 0.0000,
    is_mandatory BOOLEAN NOT NULL DEFAULT FALSE,
    display_order INT NOT NULL DEFAULT 0,
    notes TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_tpl_charge_template (template_id),
    INDEX idx_tpl_charge_org (org_id),
    CONSTRAINT fk_tpl_charge_template FOREIGN KEY (template_id) REFERENCES quotation_templates(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Add template_id lineage reference to quotations table if not exists
SET @dbname = DATABASE();
SET @tablename = "quotations";
SET @columnname = "template_id";
SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
    WHERE
      (table_name = @tablename)
      AND (table_schema = @dbname)
      AND (column_name = @columnname)
  ) > 0,
  "SELECT 1",
  CONCAT("ALTER TABLE ", @tablename, " ADD COLUMN ", @columnname, " VARCHAR(36) NULL AFTER lead_id, ADD INDEX idx_quote_template (org_id, template_id);")
));
PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;
