-- ==============================================================================
-- Migration 097: AI Memory and Personalization (Phase 2 Task 2.10)
-- Persistent, secure, user-controlled AI memory and preferences for LogisticsHQ.
-- ==============================================================================

-- 1. AI Preferences Table
CREATE TABLE IF NOT EXISTS ai_preferences (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL DEFAULT 0, -- 0 represents organization-scoped preference
    scope VARCHAR(32) NOT NULL DEFAULT 'USER', -- 'USER' or 'ORGANIZATION'
    preference_key VARCHAR(64) NOT NULL,
    preference_value TEXT NOT NULL,
    value_type VARCHAR(32) NOT NULL DEFAULT 'STRING', -- 'STRING', 'BOOLEAN', 'NUMBER', 'JSON'
    description VARCHAR(255) NULL,
    source VARCHAR(64) NOT NULL DEFAULT 'USER_SETTING', -- 'USER_SETTING', 'ASSISTANT_CONFIRMED', 'ORGANIZATION_POLICY'
    explicitly_confirmed TINYINT(1) NOT NULL DEFAULT 1,
    is_disabled TINYINT(1) NOT NULL DEFAULT 0,
    disabled_at DATETIME NULL,
    last_used_at DATETIME NULL,
    created_by VARCHAR(255) NULL,
    updated_by VARCHAR(255) NULL,
    correlation_id VARCHAR(255) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ai_pref_org_user (org_id, user_id),
    INDEX idx_ai_pref_scope (scope),
    INDEX idx_ai_pref_key (preference_key),
    UNIQUE INDEX uq_ai_pref_scope_key (org_id, user_id, preference_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. AI Memory Items Table
CREATE TABLE IF NOT EXISTS ai_memory_items (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL DEFAULT 0, -- 0 represents organization-scoped memory
    scope VARCHAR(32) NOT NULL DEFAULT 'USER', -- 'USER' or 'ORGANIZATION'
    memory_type VARCHAR(64) NOT NULL, -- 'RESPONSE_STYLE', 'SUMMARY_PREFERENCE', 'TERMINOLOGY', 'BUSINESS_FACT', 'MODULE_PREFERENCE', 'EXPLANATION_DEPTH'
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    structured_value JSON NULL,
    source_type VARCHAR(64) NOT NULL DEFAULT 'EXPLICIT_USER', -- 'EXPLICIT_USER', 'ASSISTANT_PROPOSED', 'ORGANIZATION_POLICY'
    source_reference VARCHAR(255) NULL,
    evidence TEXT NULL,
    confidence DECIMAL(5,2) NOT NULL DEFAULT 1.00,
    explicitly_confirmed TINYINT(1) NOT NULL DEFAULT 1,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE', -- 'ACTIVE', 'DISABLED', 'EXPIRED', 'DELETED', 'PENDING_REVIEW'
    review_at DATETIME NULL,
    expires_at DATETIME NULL,
    last_used_at DATETIME NULL,
    created_by VARCHAR(255) NULL,
    updated_by VARCHAR(255) NULL,
    correlation_id VARCHAR(255) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ai_mem_org_user (org_id, user_id),
    INDEX idx_ai_mem_scope (scope),
    INDEX idx_ai_mem_status (status),
    INDEX idx_ai_mem_type (memory_type),
    INDEX idx_ai_mem_expires (expires_at),
    INDEX idx_ai_mem_review (review_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. AI User Personalization Settings Table
CREATE TABLE IF NOT EXISTS ai_user_personalization_settings (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    personalization_enabled TINYINT(1) NOT NULL DEFAULT 1,
    preferred_response_style VARCHAR(64) NOT NULL DEFAULT 'CONCISE', -- 'CONCISE', 'DETAILED', 'EXECUTIVE'
    preferred_summary_depth VARCHAR(64) NOT NULL DEFAULT 'STANDARD', -- 'BRIEF', 'STANDARD', 'COMPREHENSIVE'
    preferred_currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    preferred_timezone VARCHAR(64) NOT NULL DEFAULT 'UTC',
    preferred_date_format VARCHAR(32) NOT NULL DEFAULT 'YYYY-MM-DD',
    preferred_default_module VARCHAR(64) NOT NULL DEFAULT 'DASHBOARD',
    explanation_level VARCHAR(32) NOT NULL DEFAULT 'STANDARD', -- 'DIRECT', 'STANDARD', 'IN_DEPTH'
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX uq_ai_user_pers (org_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. AI Memory Audit Events Table
CREATE TABLE IF NOT EXISTS ai_memory_audit_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL DEFAULT 0,
    memory_item_id BIGINT NULL,
    event_type VARCHAR(64) NOT NULL, -- 'PROPOSED', 'CONFIRMED', 'CREATED', 'UPDATED', 'DISABLED', 'REENABLED', 'DELETED', 'CLEARED', 'REJECTED_SENSITIVE', 'BLOCKED_AUTH', 'USED_IN_RUNTIME'
    scope VARCHAR(32) NOT NULL DEFAULT 'USER',
    actor_name VARCHAR(255) NOT NULL,
    actor_id BIGINT NULL,
    details TEXT NULL,
    correlation_id VARCHAR(255) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ai_mem_audit_org (org_id, created_at),
    INDEX idx_ai_mem_audit_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
