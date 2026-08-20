-- +migrate Up

-- 1. Alter shipment status check constraint to include 'CLOSED'
ALTER TABLE shipments DROP CONSTRAINT IF EXISTS chk_shipment_status;
ALTER TABLE shipments ADD CONSTRAINT chk_shipment_status CHECK (status IN (
    'BOOKING_PENDING',
    'BOOKED',
    'DEPARTED',
    'IN_TRANSIT',
    'ARRIVED',
    'DELIVERED',
    'EXCEPTION',
    'CLOSED'
));

-- 2. Create shipment_customer_invoices table
CREATE TABLE IF NOT EXISTS shipment_customer_invoices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          BIGINT NOT NULL REFERENCES organizations(id),
    shipment_id     BIGINT NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    invoice_number  TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'DRAFT'
                    CHECK (status IN ('DRAFT', 'APPROVED', 'SENT', 'PAID', 'OVERDUE')),
    due_date        TIMESTAMPTZ,
    currency        TEXT NOT NULL DEFAULT 'USD',
    total_amount    NUMERIC(14, 2) NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_customer_invoice_number_org ON shipment_customer_invoices (org_id, invoice_number);
CREATE INDEX IF NOT EXISTS idx_customer_invoices_shipment ON shipment_customer_invoices (shipment_id, org_id);

-- 3. Create shipment_customer_invoice_items table
CREATE TABLE IF NOT EXISTS shipment_customer_invoice_items (
    id              BIGSERIAL PRIMARY KEY,
    org_id          BIGINT NOT NULL REFERENCES organizations(id),
    invoice_id      UUID NOT NULL REFERENCES shipment_customer_invoices(id) ON DELETE CASCADE,
    charge_code     TEXT NOT NULL DEFAULT '',
    description     TEXT NOT NULL DEFAULT '',
    quantity        NUMERIC(10, 3) NOT NULL DEFAULT 1,
    unit_price      NUMERIC(14, 2) NOT NULL DEFAULT 0,
    total_amount    NUMERIC(14, 2) NOT NULL DEFAULT 0,
    currency        TEXT NOT NULL DEFAULT 'USD',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_customer_invoice_items_invoice ON shipment_customer_invoice_items (invoice_id);

-- 4. Create shipment_finance_profitability table
CREATE TABLE IF NOT EXISTS shipment_finance_profitability (
    id                   BIGSERIAL PRIMARY KEY,
    org_id               BIGINT NOT NULL REFERENCES organizations(id),
    shipment_id          BIGINT UNIQUE NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    quoted_sell_price    NUMERIC(14, 2) NOT NULL DEFAULT 0,
    quoted_buy_price     NUMERIC(14, 2) NOT NULL DEFAULT 0,
    actual_carrier_cost  NUMERIC(14, 2) NOT NULL DEFAULT 0,
    additional_charges   NUMERIC(14, 2) NOT NULL DEFAULT 0,
    actual_revenue       NUMERIC(14, 2) NOT NULL DEFAULT 0,
    expected_profit      NUMERIC(14, 2) NOT NULL DEFAULT 0,
    actual_profit        NUMERIC(14, 2) NOT NULL DEFAULT 0,
    expected_margin_pct  NUMERIC(5, 2) NOT NULL DEFAULT 0,
    actual_margin_pct    NUMERIC(5, 2) NOT NULL DEFAULT 0,
    variance             NUMERIC(14, 2) NOT NULL DEFAULT 0,
    profitability_status TEXT NOT NULL DEFAULT 'PENDING'
                         CHECK (profitability_status IN ('PENDING', 'ON_TARGET', 'UNDER_TARGET', 'NEGATIVE_MARGIN')),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_profitability_shipment ON shipment_finance_profitability (shipment_id, org_id);
