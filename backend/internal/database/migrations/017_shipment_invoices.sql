-- Phase 5: Finance & Reconciliation
-- Migration 017: shipment_invoices, shipment_invoice_items, shipment_finance_discrepancies

CREATE TABLE IF NOT EXISTS shipment_invoices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          BIGINT NOT NULL REFERENCES organizations(id),
    shipment_id     BIGINT NOT NULL REFERENCES shipments(id),
    invoice_number  TEXT NOT NULL,
    vendor_name     TEXT NOT NULL DEFAULT '',
    vendor_ref      TEXT NOT NULL DEFAULT '',
    s3_key          TEXT NOT NULL DEFAULT '',
    file_name       TEXT NOT NULL DEFAULT '',
    currency        TEXT NOT NULL DEFAULT 'USD',
    total_amount    NUMERIC(14, 2) NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'PENDING_RECONCILIATION'
                    CHECK (status IN ('DRAFT', 'PENDING_RECONCILIATION', 'APPROVED', 'DISCREPANCY', 'PAID')),
    ai_summary      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoices_shipment ON shipment_invoices (shipment_id, org_id);
-- Enforce one invoice per invoice number per org
CREATE UNIQUE INDEX IF NOT EXISTS uq_invoice_number_org ON shipment_invoices (org_id, invoice_number);

CREATE TABLE IF NOT EXISTS shipment_invoice_items (
    id              BIGSERIAL PRIMARY KEY,
    org_id          BIGINT NOT NULL REFERENCES organizations(id),
    invoice_id      UUID NOT NULL REFERENCES shipment_invoices(id) ON DELETE CASCADE,
    charge_code     TEXT NOT NULL DEFAULT '',
    description     TEXT NOT NULL DEFAULT '',
    quantity        NUMERIC(10, 3) NOT NULL DEFAULT 1,
    unit_price      NUMERIC(14, 2) NOT NULL DEFAULT 0,
    total_amount    NUMERIC(14, 2) NOT NULL DEFAULT 0,
    currency        TEXT NOT NULL DEFAULT 'USD',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoice_items_invoice ON shipment_invoice_items (invoice_id);

CREATE TABLE IF NOT EXISTS shipment_finance_discrepancies (
    id              BIGSERIAL PRIMARY KEY,
    org_id          BIGINT NOT NULL REFERENCES organizations(id),
    shipment_id     BIGINT NOT NULL REFERENCES shipments(id),
    invoice_id      UUID NOT NULL REFERENCES shipment_invoices(id) ON DELETE CASCADE,
    charge_code     TEXT NOT NULL DEFAULT '',
    field_name      TEXT NOT NULL DEFAULT '',
    expected_value  TEXT NOT NULL DEFAULT '',
    actual_value    TEXT NOT NULL DEFAULT '',
    source          TEXT NOT NULL DEFAULT '',  -- 'CONTRACT', 'QUOTE', 'MANIFEST'
    status          TEXT NOT NULL DEFAULT 'OPEN'
                    CHECK (status IN ('OPEN', 'REVIEWED', 'RESOLVED')),
    resolved_by     BIGINT REFERENCES users(id),
    resolved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_finance_discrepancy ON shipment_finance_discrepancies
    (invoice_id, charge_code, field_name, source);
CREATE INDEX IF NOT EXISTS idx_finance_discrepancies_invoice ON shipment_finance_discrepancies (invoice_id, org_id);
