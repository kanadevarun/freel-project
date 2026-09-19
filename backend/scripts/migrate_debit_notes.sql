-- Debit Notes Feature Migration
-- Run once against the freel DB to add debit_notes and debit_note_items tables.

CREATE TABLE IF NOT EXISTS debit_notes (
  id                  BIGINT AUTO_INCREMENT PRIMARY KEY,
  org_id              BIGINT NOT NULL,
  debit_note_number   VARCHAR(60) NOT NULL,
  customer_id         BIGINT NOT NULL,
  customer_name       VARCHAR(255) NOT NULL DEFAULT '',
  customer_country    VARCHAR(100) NOT NULL DEFAULT '',
  shipment_id         BIGINT NULL,
  shipment_number     VARCHAR(100) NOT NULL DEFAULT '',
  invoice_id          BIGINT NULL,
  invoice_number      VARCHAR(60) NOT NULL DEFAULT '',
  reason              TEXT NOT NULL DEFAULT '',
  currency            VARCHAR(10) NOT NULL DEFAULT 'USD',
  subtotal            DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  tax_amount          DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  total_amount        DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  status              VARCHAR(30) NOT NULL DEFAULT 'DRAFT',
  issue_date          DATE NULL,
  due_date            DATE NULL,
  notes               TEXT NULL,
  created_by_id       BIGINT NULL,
  created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uq_org_dn_number (org_id, debit_note_number),
  INDEX idx_dn_org_id (org_id),
  INDEX idx_dn_customer_id (customer_id),
  INDEX idx_dn_shipment_id (shipment_id),
  INDEX idx_dn_invoice_id (invoice_id),
  INDEX idx_dn_status (status)
);

CREATE TABLE IF NOT EXISTS debit_note_items (
  id               BIGINT AUTO_INCREMENT PRIMARY KEY,
  org_id           BIGINT NOT NULL,
  debit_note_id    BIGINT NOT NULL,
  description      VARCHAR(500) NOT NULL DEFAULT '',
  service_category VARCHAR(100) NOT NULL DEFAULT '',
  quantity         DECIMAL(12,4) NOT NULL DEFAULT 1.0000,
  unit_price       DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  total_amount     DECIMAL(18,2) NOT NULL DEFAULT 0.00,
  display_order    INT NOT NULL DEFAULT 0,
  created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_dni_debit_note_id (debit_note_id),
  FOREIGN KEY fk_dni_debit_note (debit_note_id) REFERENCES debit_notes(id) ON DELETE CASCADE
);
