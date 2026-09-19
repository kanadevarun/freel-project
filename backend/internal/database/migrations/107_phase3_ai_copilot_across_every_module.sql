-- Migration 107: Phase 3 Task 3.10 - AI Copilot Across Every Module
-- Provides durable persistence for context-aware copilot chat sessions, grounded conversation turns, source references, and controlled action requests.

CREATE TABLE IF NOT EXISTS copilot_sessions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    session_id VARCHAR(128) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL DEFAULT 'New Conversation',
    current_module VARCHAR(64) NOT NULL DEFAULT 'DASHBOARD',
    current_route VARCHAR(255) NOT NULL DEFAULT '/dashboard',
    current_record_id VARCHAR(64) NULL,
    is_archived TINYINT(1) NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_copilot_sess_org_user (org_id, user_id, is_archived),
    INDEX idx_copilot_sess_module (org_id, current_module)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS copilot_messages (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(128) NOT NULL,
    org_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    role VARCHAR(20) NOT NULL COMMENT 'user, assistant, system',
    content TEXT NOT NULL,
    confirmed_facts JSON NULL,
    source_references JSON NULL,
    action_proposals JSON NULL,
    draft_content TEXT NULL,
    draft_type VARCHAR(64) NULL,
    confidence_score FLOAT NULL,
    correlation_id VARCHAR(128) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_copilot_msg_session (session_id, created_at),
    INDEX idx_copilot_msg_org (org_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS copilot_action_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    session_id VARCHAR(128) NOT NULL,
    action_type VARCHAR(64) NOT NULL,
    action_title VARCHAR(255) NOT NULL,
    action_payload JSON NOT NULL,
    approval_id BIGINT NULL,
    recommendation_id BIGINT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'PROPOSED' COMMENT 'PROPOSED, PENDING_APPROVAL, EXECUTED, REJECTED',
    result_summary TEXT NULL,
    correlation_id VARCHAR(128) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_copilot_act_org (org_id, created_at),
    INDEX idx_copilot_act_session (session_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
