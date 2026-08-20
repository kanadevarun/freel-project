-- +migrate Up

-- ─────────────────────────────────────────────────────────────────────────────
-- CONTRACT_DOCUMENTS
-- Tracks uploaded carrier contract documents and their AI extraction status.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE contract_documents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    carrier_scac    VARCHAR(10) REFERENCES carriers(scac),
    carrier_name    VARCHAR(100),
    file_name       VARCHAR(500) NOT NULL,
    s3_key          VARCHAR(1000) NOT NULL, -- or local path
    file_type       VARCHAR(20) NOT NULL DEFAULT 'PDF', -- PDF | XLSX | EMAIL | IMAGE
    file_size_bytes BIGINT,
    page_count      INT,

    -- Processing State
    status          VARCHAR(50) NOT NULL DEFAULT 'QUEUED',
    -- QUEUED | OCR_PROCESSING | AI_EXTRACTING | PENDING_REVIEW | CONFIRMED | FAILED

    -- Extraction Summary
    extracted_rate_count    INT NOT NULL DEFAULT 0,
    confirmed_rate_count    INT NOT NULL DEFAULT 0,
    pending_review_count    INT NOT NULL DEFAULT 0,
    failed_rate_count       INT NOT NULL DEFAULT 0,

    -- AI Processing Metadata
    processing_started_at   TIMESTAMPTZ,
    processing_completed_at TIMESTAMPTZ,
    processing_log          JSONB NOT NULL DEFAULT '[]', -- JSON array of {step, timestamp, message}
    ai_document_summary     TEXT,

    -- Human Review
    reviewed_by             BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at             TIMESTAMPTZ,
    review_notes            TEXT,

    created_by              BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_contract_documents_org_id ON contract_documents(org_id);
CREATE INDEX idx_contract_documents_status ON contract_documents(status);
CREATE INDEX idx_contract_documents_carrier ON contract_documents(carrier_scac);

-- ─────────────────────────────────────────────────────────────────────────────
-- RATE_REVIEW_QUEUE
-- Flagged rate extractions that need human correction / validation.
-- ─────────────────────────────────────────────────────────────────────────────
CREATE TABLE rate_review_queue (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    contract_doc_id     UUID NOT NULL REFERENCES contract_documents(id) ON DELETE CASCADE,

    -- AI extraction result draft
    extracted_data      JSONB NOT NULL, -- proposed CanonicalRate fields
    confidence_score    SMALLINT NOT NULL DEFAULT 0,
    review_flags        TEXT[] NOT NULL DEFAULT '{}',
    ai_reasoning        TEXT,

    -- Verification context
    source_page         INT,
    source_text         TEXT,
    source_image_url    TEXT,

    -- Status
    status              VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    -- PENDING | APPROVED | REJECTED | CORRECTED

    reviewed_by         BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at         TIMESTAMPTZ,
    corrected_data      JSONB,
    review_notes        TEXT,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rate_review_queue_org_status ON rate_review_queue(org_id, status);
CREATE INDEX idx_rate_review_queue_doc ON rate_review_queue(contract_doc_id);

-- Alter rate_entries to support referencing contract_documents
ALTER TABLE rate_entries
ADD COLUMN IF NOT EXISTS contract_doc_id UUID REFERENCES contract_documents(id) ON DELETE SET NULL;

-- +migrate Down
ALTER TABLE rate_entries DROP COLUMN IF EXISTS contract_doc_id;
DROP TABLE IF EXISTS rate_review_queue;
DROP TABLE IF EXISTS contract_documents;
