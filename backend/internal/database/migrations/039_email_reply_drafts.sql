CREATE TABLE lead_email_drafts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    lead_id BIGINT NOT NULL,
    parent_interaction_id BIGINT NOT NULL,
    mailbox_id BIGINT NULL,
    recipients TEXT NULL,
    cc_recipients TEXT NULL,
    subject VARCHAR(500) NULL,
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY unique_lead_parent_draft (org_id, lead_id, parent_interaction_id)
);
