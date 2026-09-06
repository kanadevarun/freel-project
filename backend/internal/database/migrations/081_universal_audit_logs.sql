-- Migration 081: Universal Audit Logs Foundation & Enterprise Schema
-- Extends the audit_logs table to support universal logging across all LogisticsHQ modules:
-- Authentication, Users, Roles, Leads, RFQs, Quotes, Bookings, Shipments, Tracking, Rates, Contracts, Customers, Outreach, Documents, Approvals, Invoices, Payments, Carrier Integrations, Settings, and future AI Agents.

-- 1. Ensure the audit_logs table exists
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    user_id BIGINT NULL,
    actor_type VARCHAR(50) NOT NULL DEFAULT 'USER',
    actor_name VARCHAR(255) NULL,
    actor_role VARCHAR(100) NULL,
    action VARCHAR(100) NOT NULL,
    module VARCHAR(100) NOT NULL DEFAULT 'GENERAL',
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    resource_name VARCHAR(255) NULL,
    description TEXT NOT NULL,
    result VARCHAR(50) NOT NULL DEFAULT 'SUCCESS',
    error_message TEXT NULL,
    before_data JSON NULL,
    after_data JSON NULL,
    changes JSON NULL,
    metadata JSON NULL,
    ip_address VARCHAR(100) NULL,
    user_agent VARCHAR(500) NULL,
    details JSON NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_audit_logs_org_created (org_id, created_at DESC),
    INDEX idx_audit_logs_org_module (org_id, module),
    INDEX idx_audit_logs_org_action (org_id, action),
    INDEX idx_audit_logs_org_resource (org_id, resource_type, resource_id),
    INDEX idx_audit_logs_org_actor (org_id, user_id),
    INDEX idx_audit_logs_org_result (org_id, result),
    CONSTRAINT fk_audit_logs_org_id FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_audit_logs_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
