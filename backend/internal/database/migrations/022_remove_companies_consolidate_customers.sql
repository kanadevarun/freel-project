-- Migration: 022_remove_companies_consolidate_customers.sql
-- Description: Remove the redundant `companies` table.
--   The `organizations` table is now the single source of truth for the freight
--   forwarder's own company details.
--   The `customers` table is restructured to hold client company info directly,
--   eliminating the need for the `companies` join table.
--   The `contacts` table company_id now references `customers` directly.

-- ── Step 1: Drop FK constraints that reference companies ──────────────────────

ALTER TABLE contacts
  DROP FOREIGN KEY fk_cont_comp;

ALTER TABLE customers
  DROP FOREIGN KEY fk_cust_comp;

-- ── Step 2: Add company-level fields directly to customers ────────────────────
-- (These were previously stored in `companies` and joined via `customers`)

ALTER TABLE customers
  ADD COLUMN name        VARCHAR(255)  NOT NULL DEFAULT '' AFTER org_id,
  ADD COLUMN domain      VARCHAR(255)  NULL AFTER name,
  ADD COLUMN industry    VARCHAR(255)  NULL AFTER domain,
  ADD COLUMN contact_name  VARCHAR(255) NULL AFTER industry,
  ADD COLUMN contact_email VARCHAR(255) NULL AFTER contact_name,
  ADD COLUMN contact_phone VARCHAR(50)  NULL AFTER contact_email,
  MODIFY COLUMN company_id BIGINT NULL; -- make nullable since it will be phased out

-- ── Step 3: Update contacts to reference customers instead of companies ────────

ALTER TABLE contacts
  ADD COLUMN customer_id BIGINT NULL AFTER org_id;

ALTER TABLE contacts
  ADD CONSTRAINT fk_cont_customer
  FOREIGN KEY (customer_id) REFERENCES customers (id) ON DELETE SET NULL;

-- Keep company_id column but remove FK (will be cleaned up once data migrated)
-- ALTER TABLE contacts DROP COLUMN company_id;  -- skipped for safety

-- ── Step 4: Drop the companies table (no FK references left) ─────────────────
-- Only safe because all tables have been migrated off the FK above.

DROP TABLE IF EXISTS companies;
