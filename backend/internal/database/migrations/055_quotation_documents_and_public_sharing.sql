-- 055_quotation_documents_and_public_sharing.sql
-- Quotation PDF Documents and Secure Public Sharing Links (Task 18.5)

CREATE TABLE IF NOT EXISTS quotation_documents (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,
    document_type VARCHAR(32) NOT NULL DEFAULT 'PDF',
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    version INT NOT NULL DEFAULT 1,
    generated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    generated_by BIGINT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_quotation_docs_org_quote (org_id, quotation_id),
    INDEX idx_quotation_docs_type (quotation_id, document_type),
    CONSTRAINT fk_quotation_documents_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_quotation_documents_quote FOREIGN KEY (quotation_id) REFERENCES quotations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS quotation_public_links (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,
    public_token VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    expires_at DATETIME NULL,
    created_by BIGINT NOT NULL,
    revoked_at DATETIME NULL,
    revoked_by BIGINT NULL,
    revocation_reason VARCHAR(255) NULL,
    last_accessed_at DATETIME NULL,
    access_count INT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX uq_quote_public_token (public_token),
    INDEX idx_quote_public_links_org_quote (org_id, quotation_id),
    INDEX idx_quote_public_links_status (quotation_id, status),
    CONSTRAINT fk_quote_public_links_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_quote_public_links_quote FOREIGN KEY (quotation_id) REFERENCES quotations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
