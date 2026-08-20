-- +migrate Up

-- 1. SHIPMENT_DOCUMENTS
CREATE TABLE IF NOT EXISTS shipment_documents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    shipment_id     BIGINT NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    doc_type        VARCHAR(50) NOT NULL,   -- MBL, HBL, COMMERCIAL_INVOICE, PACKING_LIST
    s3_key          VARCHAR(1000) NOT NULL,
    file_name       VARCHAR(500) NOT NULL,
    file_type       VARCHAR(20) DEFAULT 'PDF',
    status          VARCHAR(50) NOT NULL DEFAULT 'PENDING_VERIFICATION',
    extracted_data  JSONB DEFAULT '{}',     -- Extracted key fields
    raw_ocr_text    TEXT,                   -- Raw OCR text preserved for audit
    ai_summary      TEXT,
    verified_by     BIGINT REFERENCES users(id) ON DELETE SET NULL,
    verified_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT uq_shipment_doc_type UNIQUE (shipment_id, doc_type)
);

-- 2. SHIPMENT_DOCUMENT_DISCREPANCIES
CREATE TABLE IF NOT EXISTS shipment_document_discrepancies (
    id              BIGSERIAL PRIMARY KEY,
    org_id          BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    shipment_id     BIGINT NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    field_name      VARCHAR(100) NOT NULL, -- gross_weight, package_count, container_numbers, seal_numbers, etc.
    expected_value  TEXT,
    actual_value    TEXT,
    source_document VARCHAR(50) NOT NULL,  -- e.g. MBL
    target_document VARCHAR(50) NOT NULL,  -- e.g. HBL
    status          VARCHAR(50) NOT NULL DEFAULT 'OPEN', -- OPEN, RESOLVED
    resolved_by     BIGINT REFERENCES users(id) ON DELETE SET NULL,
    resolved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT uq_discrepancy_key UNIQUE (shipment_id, field_name, source_document, target_document)
);

CREATE INDEX idx_ship_docs_shipment ON shipment_documents(shipment_id);
CREATE INDEX idx_ship_docs_status ON shipment_documents(status);
CREATE INDEX idx_ship_discrepancies_shipment ON shipment_document_discrepancies(shipment_id);

-- +migrate Down
DROP TABLE IF EXISTS shipment_document_discrepancies;
DROP TABLE IF EXISTS shipment_documents;
