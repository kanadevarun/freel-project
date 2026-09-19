-- 084_action_idempotency_keys.sql
-- Idempotency store for the centralized business action system

CREATE TABLE IF NOT EXISTS action_idempotency_keys (
    org_id BIGINT NOT NULL,
    idempotency_key VARCHAR(191) NOT NULL,
    action_name VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'IN_PROGRESS',
    result_json LONGTEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL,
    PRIMARY KEY (org_id, idempotency_key),
    INDEX idx_action_idempotency_expiry (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
