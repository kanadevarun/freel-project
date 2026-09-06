-- 076_approvals_module_foundation.sql
-- Creates approval_requests table for enterprise approval workflows across LogisticsHQ

CREATE TABLE IF NOT EXISTS approval_requests (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    request_code VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'DOCUMENTS', -- DOCUMENTS, COMMERCIAL, OPERATIONS, FINANCE
    type VARCHAR(100) NOT NULL DEFAULT 'Document Approval', -- Document Approval, Commercial Approval, Operations Approval, Finance Approval
    status VARCHAR(50) NOT NULL DEFAULT 'Pending', -- Pending, In Review, Approved, Rejected, Overdue, Cancelled
    priority VARCHAR(20) NOT NULL DEFAULT 'MEDIUM', -- LOW, MEDIUM, HIGH, URGENT

    -- Related context links
    related_entity_type VARCHAR(50) NULL, -- SHIPMENT, CUSTOMER, DOCUMENT, INVOICE, QUOTATION, BOOKING, LEAD, RFQ
    related_entity_id BIGINT NULL,
    related_ref VARCHAR(100) NULL, -- e.g. Shipment: SHP-250812 or Customer: CUST-00734
    customer_name VARCHAR(255) NULL,
    customer_id BIGINT NULL,
    shipment_id BIGINT NULL,
    document_id BIGINT NULL,
    booking_id BIGINT NULL,

    -- Requester and Department
    requested_by_id BIGINT NULL,
    requested_by_name VARCHAR(100) NOT NULL,
    department VARCHAR(100) NULL DEFAULT 'Operations',
    avatar VARCHAR(10) NULL DEFAULT 'AS',

    -- Due dates
    due_date DATETIME NULL,
    due_text VARCHAR(50) NULL DEFAULT '7 days left',

    -- Decision and audit fields
    assigned_to VARCHAR(100) NULL,
    approved_by VARCHAR(100) NULL,
    approved_at DATETIME NULL,
    rejected_by VARCHAR(100) NULL,
    rejected_at DATETIME NULL,
    rejection_reason TEXT NULL,
    comments TEXT NULL,
    description TEXT NULL,

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_ar_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_ar_customer FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE SET NULL,
    CONSTRAINT fk_ar_shipment FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE SET NULL,
    CONSTRAINT fk_ar_document FOREIGN KEY (document_id) REFERENCES shipment_documents(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_ar_org_status ON approval_requests(org_id, status);
CREATE INDEX idx_ar_category ON approval_requests(org_id, category);
CREATE INDEX idx_ar_priority ON approval_requests(org_id, priority);
CREATE INDEX idx_ar_created_at ON approval_requests(org_id, created_at);
