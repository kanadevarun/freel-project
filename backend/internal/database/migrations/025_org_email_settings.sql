-- Migration: 025_org_email_settings.sql

CREATE TABLE IF NOT EXISTS org_email_settings (
    org_id BIGINT PRIMARY KEY REFERENCES organizations(id) ON DELETE CASCADE,
    process_logistics_inquiries BOOLEAN DEFAULT TRUE,
    track_email_threads BOOLEAN DEFAULT TRUE,
    smart_filtering BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS org_connected_mailboxes (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    owner_name VARCHAR(255) NOT NULL,
    mailbox_type VARCHAR(50) NOT NULL DEFAULT 'Individual',
    is_primary BOOLEAN DEFAULT FALSE,
    status VARCHAR(50) NOT NULL DEFAULT 'Connected',
    last_synced_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_org_connected_mailboxes_org_id ON org_connected_mailboxes(org_id);

-- Seed with data matching the visual mockup
INSERT IGNORE INTO org_connected_mailboxes (org_id, email, owner_name, mailbox_type, is_primary, status, last_synced_at)
VALUES 
(1, 'sales@abcfreight.com', 'Primary business mailbox', 'Shared / Team', TRUE, 'Connected', CURRENT_TIMESTAMP),
(1, 'varun@abcfreight.com', 'Varun Kanade', 'Individual', FALSE, 'Connected', CURRENT_TIMESTAMP),
(1, 'neha@abcfreight.com', 'Neha Sharma', 'Individual', FALSE, 'Connected', CURRENT_TIMESTAMP);
