-- Migration: 069_customer_foundation_and_directory.sql
-- Description: Customer Foundation & Directory schema enhancements, operational locations/addresses, and lead conversion links.

-- 1. Enhance customers table with enterprise freight-forwarder attributes
ALTER TABLE customers
  ADD COLUMN customer_code     VARCHAR(50)   NULL AFTER org_id,
  ADD COLUMN trading_name      VARCHAR(255)  NULL AFTER name,
  ADD COLUMN customer_type     VARCHAR(50)   NOT NULL DEFAULT 'SHIPPER' AFTER trading_name,
  ADD COLUMN tax_id            VARCHAR(100)  NULL AFTER industry,
  ADD COLUMN pan_number        VARCHAR(50)   NULL AFTER tax_id,
  ADD COLUMN eori_number       VARCHAR(50)   NULL AFTER pan_number,
  ADD COLUMN currency          VARCHAR(10)   NOT NULL DEFAULT 'USD' AFTER eori_number,
  ADD COLUMN payment_terms     VARCHAR(50)   NOT NULL DEFAULT 'NET30' AFTER currency,
  ADD COLUMN credit_limit      DECIMAL(14,2) NOT NULL DEFAULT 0.00 AFTER payment_terms,
  ADD COLUMN health_score      INT           NOT NULL DEFAULT 80 AFTER credit_limit,
  ADD COLUMN account_owner_id  BIGINT        NULL AFTER health_score,
  ADD COLUMN website           VARCHAR(255)  NULL AFTER domain,
  ADD COLUMN country           VARCHAR(100)  NULL AFTER website,
  ADD COLUMN city              VARCHAR(100)  NULL AFTER country,
  ADD COLUMN notes             TEXT          NULL AFTER contact_phone,
  ADD COLUMN archived_at       TIMESTAMP     NULL AFTER updated_at;

-- Indexes for performance & tenant isolation
CREATE UNIQUE INDEX uq_customers_org_code ON customers(org_id, customer_code);
CREATE INDEX idx_customers_org_status ON customers(org_id, status);
CREATE INDEX idx_customers_org_type ON customers(org_id, customer_type);
CREATE INDEX idx_customers_org_owner ON customers(org_id, account_owner_id);
CREATE INDEX idx_customers_org_name ON customers(org_id, name);

-- 2. Enhance contacts table for customer contacts
ALTER TABLE contacts
  ADD COLUMN department        VARCHAR(100)  NULL AFTER job_title,
  ADD COLUMN is_primary        BOOLEAN       NOT NULL DEFAULT FALSE AFTER department,
  ADD COLUMN notes             TEXT          NULL AFTER is_primary;

CREATE INDEX idx_contacts_org_customer ON contacts(org_id, customer_id);

-- 3. Dedicated Customer Operational Addresses
CREATE TABLE IF NOT EXISTS customer_addresses (
    id                   BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id               BIGINT NOT NULL,
    customer_id          BIGINT NOT NULL,
    address_type         VARCHAR(50) NOT NULL DEFAULT 'BILLING', -- BILLING, SHIPPING, WAREHOUSE, REGISTERED_OFFICE, PORT_FACILITY
    label                VARCHAR(150) NULL, -- e.g. "Main Factory", "Rotterdam Warehouse"
    address_line_1       VARCHAR(255) NOT NULL,
    address_line_2       VARCHAR(255) NULL,
    city                 VARCHAR(100) NOT NULL,
    state                VARCHAR(100) NULL,
    postal_code          VARCHAR(30)  NULL,
    country_code         VARCHAR(10)  NOT NULL DEFAULT 'US',
    country              VARCHAR(100) NOT NULL DEFAULT 'United States',
    is_primary_billing   BOOLEAN NOT NULL DEFAULT FALSE,
    is_primary_shipping  BOOLEAN NOT NULL DEFAULT FALSE,
    contact_name         VARCHAR(150) NULL,
    contact_phone        VARCHAR(50) NULL,
    contact_email        VARCHAR(150) NULL,
    created_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_cust_addr_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_cust_addr_cust FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE
);

CREATE INDEX idx_cust_addr_lookup ON customer_addresses(org_id, customer_id, address_type);

-- 4. Customer Lead Conversion Audit & History Links
CREATE TABLE IF NOT EXISTS customer_lead_links (
    id                   BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id               BIGINT NOT NULL,
    customer_id          BIGINT NOT NULL,
    lead_id              BIGINT NOT NULL,
    converted_by_user_id BIGINT NULL,
    conversion_notes     TEXT NULL,
    created_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_cll_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_cll_cust FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
    CONSTRAINT fk_cll_lead FOREIGN KEY (lead_id) REFERENCES leads(id) ON DELETE CASCADE,
    CONSTRAINT uq_cll_lead UNIQUE (org_id, lead_id)
);

CREATE INDEX idx_cll_cust ON customer_lead_links(org_id, customer_id);

-- 5. Backfill existing customers with customer codes if null
UPDATE customers 
SET customer_code = CONCAT('CUST-', YEAR(NOW()), '-', LPAD(id, 5, '0'))
WHERE customer_code IS NULL OR customer_code = '';
