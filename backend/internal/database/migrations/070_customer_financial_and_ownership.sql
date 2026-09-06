-- Migration 070: Customer Financial & Relationship Management

-- 1. Enhance customers table with credit status, secondary owner, and commercial notes
ALTER TABLE customers
  ADD COLUMN IF NOT EXISTS credit_status VARCHAR(50) NOT NULL DEFAULT 'GOOD_STANDING',
  ADD COLUMN IF NOT EXISTS secondary_owner_id BIGINT NULL,
  ADD COLUMN IF NOT EXISTS commercial_notes TEXT NULL;

-- 2. Customer Ownership Audit History Table
CREATE TABLE IF NOT EXISTS customer_ownership_history (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  org_id BIGINT NOT NULL,
  customer_id BIGINT NOT NULL,
  previous_owner_id BIGINT NULL,
  new_owner_id BIGINT NOT NULL,
  ownership_type VARCHAR(50) NOT NULL DEFAULT 'PRIMARY',
  changed_by_user_id BIGINT NULL,
  change_reason VARCHAR(255) NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_cust_owner_hist_org_cust (org_id, customer_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Enhance contacts table with contact role and mobile number
ALTER TABLE contacts
  ADD COLUMN IF NOT EXISTS contact_role VARCHAR(50) NOT NULL DEFAULT 'COMMERCIAL',
  ADD COLUMN IF NOT EXISTS mobile VARCHAR(50) NULL;
