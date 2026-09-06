-- +migrate Up

-- Alter shipment_documents to support comprehensive operational document metadata, compliance lifecycle, and lineage
ALTER TABLE shipment_documents 
    ADD COLUMN IF NOT EXISTS document_name VARCHAR(255) NULL,
    ADD COLUMN IF NOT EXISTS category VARCHAR(100) NOT NULL DEFAULT 'TRANSPORT',
    ADD COLUMN IF NOT EXISTS description TEXT NULL,
    ADD COLUMN IF NOT EXISTS file_url VARCHAR(1000) NULL,
    ADD COLUMN IF NOT EXISTS file_size BIGINT NULL,
    ADD COLUMN IF NOT EXISTS mime_type VARCHAR(100) NULL,
    ADD COLUMN IF NOT EXISTS uploaded_by VARCHAR(255) NULL,
    ADD COLUMN IF NOT EXISTS uploaded_at DATETIME NULL,
    ADD COLUMN IF NOT EXISTS reviewed_by VARCHAR(255) NULL,
    ADD COLUMN IF NOT EXISTS reviewed_at DATETIME NULL,
    ADD COLUMN IF NOT EXISTS rejection_reason TEXT NULL,
    ADD COLUMN IF NOT EXISTS expires_at DATETIME NULL,
    ADD COLUMN IF NOT EXISTS document_date DATETIME NULL,
    ADD COLUMN IF NOT EXISTS reference_number VARCHAR(255) NULL,
    ADD COLUMN IF NOT EXISTS source VARCHAR(50) NOT NULL DEFAULT 'SHIPMENT',
    ADD COLUMN IF NOT EXISTS source_id BIGINT NULL;

CREATE INDEX IF NOT EXISTS idx_ship_docs_cat ON shipment_documents(org_id, shipment_id, category);
CREATE INDEX IF NOT EXISTS idx_ship_docs_src ON shipment_documents(org_id, source, source_id);

-- +migrate Down
ALTER TABLE shipment_documents
    DROP COLUMN IF EXISTS document_name,
    DROP COLUMN IF EXISTS category,
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS file_url,
    DROP COLUMN IF EXISTS file_size,
    DROP COLUMN IF EXISTS mime_type,
    DROP COLUMN IF EXISTS uploaded_by,
    DROP COLUMN IF EXISTS uploaded_at,
    DROP COLUMN IF EXISTS reviewed_by,
    DROP COLUMN IF EXISTS reviewed_at,
    DROP COLUMN IF EXISTS rejection_reason,
    DROP COLUMN IF EXISTS expires_at,
    DROP COLUMN IF EXISTS document_date,
    DROP COLUMN IF EXISTS reference_number,
    DROP COLUMN IF EXISTS source,
    DROP COLUMN IF EXISTS source_id;
