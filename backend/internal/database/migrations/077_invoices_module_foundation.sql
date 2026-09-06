-- Phase 5: Finance & Invoices
-- Migration 077: Invoices Module Foundation (Customer Invoices, Line Items, History, Documents & Payments)

CREATE TABLE IF NOT EXISTS customer_invoices (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id          BIGINT NOT NULL,
    invoice_number  VARCHAR(100) NOT NULL,
    customer_id     BIGINT NOT NULL,
    customer_name   VARCHAR(255) NOT NULL,
    customer_country VARCHAR(100) NOT NULL DEFAULT '',
    shipment_id     BIGINT NULL,
    shipment_number VARCHAR(100) NOT NULL DEFAULT '',
    booking_id      BIGINT NULL,
    booking_number  VARCHAR(100) NOT NULL DEFAULT '',
    quotation_id    BIGINT NULL,
    quote_number    VARCHAR(100) NOT NULL DEFAULT '',
    route           VARCHAR(255) NOT NULL DEFAULT '',
    origin          VARCHAR(100) NOT NULL DEFAULT '',
    destination     VARCHAR(100) NOT NULL DEFAULT '',
    invoice_date    DATE NOT NULL,
    due_date        DATE NOT NULL,
    days_left       VARCHAR(50) NOT NULL DEFAULT '',
    currency        VARCHAR(10) NOT NULL DEFAULT 'USD',
    subtotal        DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    tax_amount      DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    discount_amount DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    total_amount    DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    paid_amount     DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    balance_due     DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    status          VARCHAR(50) NOT NULL DEFAULT 'Draft',
    type            VARCHAR(50) NOT NULL DEFAULT 'CUSTOMER_AR',
    bookmarked      BOOLEAN NOT NULL DEFAULT FALSE,
    is_my_invoice   BOOLEAN NOT NULL DEFAULT TRUE,
    creator_name    VARCHAR(100) NOT NULL DEFAULT '',
    created_by_id   BIGINT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT uq_cust_inv_num_org UNIQUE (org_id, invoice_number),
    CONSTRAINT fk_cust_inv_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_cust_inv_customer FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_cust_invoices_org_status ON customer_invoices (org_id, status);
CREATE INDEX idx_cust_invoices_org_customer ON customer_invoices (org_id, customer_id);
CREATE INDEX idx_cust_invoices_org_shipment ON customer_invoices (org_id, shipment_id);

CREATE TABLE IF NOT EXISTS customer_invoice_items (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id          BIGINT NOT NULL,
    invoice_id      BIGINT NOT NULL,
    description     VARCHAR(255) NOT NULL,
    service_category VARCHAR(100) NOT NULL DEFAULT '',
    quantity        DECIMAL(12, 4) NOT NULL DEFAULT 1.0000,
    unit_price      DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    total_amount    DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    display_order   INT NOT NULL DEFAULT 0,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_cust_inv_item_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_cust_inv_item_invoice FOREIGN KEY (invoice_id) REFERENCES customer_invoices(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_cust_inv_items_invoice ON customer_invoice_items (org_id, invoice_id);

CREATE TABLE IF NOT EXISTS customer_invoice_payments (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id          BIGINT NOT NULL,
    invoice_id      BIGINT NOT NULL,
    payment_ref     VARCHAR(100) NOT NULL,
    amount          DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    payment_method  VARCHAR(100) NOT NULL DEFAULT 'Wire Transfer',
    status          VARCHAR(50) NOT NULL DEFAULT 'Completed',
    payment_date    DATE NOT NULL,
    notes           TEXT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_cust_inv_pay_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_cust_inv_pay_invoice FOREIGN KEY (invoice_id) REFERENCES customer_invoices(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_cust_inv_pay_invoice ON customer_invoice_payments (org_id, invoice_id);

CREATE TABLE IF NOT EXISTS customer_invoice_documents (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id          BIGINT NOT NULL,
    invoice_id      BIGINT NOT NULL,
    document_name   VARCHAR(255) NOT NULL,
    file_size       VARCHAR(50) NOT NULL DEFAULT '250 KB',
    file_type       VARCHAR(100) NOT NULL DEFAULT 'application/pdf',
    s3_key          TEXT NULL,
    uploaded_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_cust_inv_doc_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_cust_inv_doc_invoice FOREIGN KEY (invoice_id) REFERENCES customer_invoices(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_cust_inv_doc_invoice ON customer_invoice_documents (org_id, invoice_id);

CREATE TABLE IF NOT EXISTS customer_invoice_history (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id          BIGINT NOT NULL,
    invoice_id      BIGINT NOT NULL,
    title           VARCHAR(255) NOT NULL,
    description     TEXT NOT NULL,
    user_name       VARCHAR(100) NOT NULL DEFAULT '',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_cust_inv_hist_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_cust_inv_hist_invoice FOREIGN KEY (invoice_id) REFERENCES customer_invoices(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_cust_inv_hist_invoice ON customer_invoice_history (org_id, invoice_id);
