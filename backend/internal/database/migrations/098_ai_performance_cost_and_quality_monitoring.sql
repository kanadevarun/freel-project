-- Migration: 098_ai_performance_cost_and_quality_monitoring.sql
-- Phase 2 Task 2.11: AI Performance, Cost, and Quality Monitoring

-- 1. Extend ai_execution_traces with structured tokens, cost, latency, assistant & feature tracking
ALTER TABLE ai_execution_traces
    ADD COLUMN IF NOT EXISTS user_id BIGINT NULL AFTER org_id,
    ADD COLUMN IF NOT EXISTS feature VARCHAR(100) NOT NULL DEFAULT 'generic' AFTER workflow_name,
    ADD COLUMN IF NOT EXISTS assistant VARCHAR(100) NOT NULL DEFAULT 'general_assistant' AFTER feature,
    ADD COLUMN IF NOT EXISTS module VARCHAR(100) NOT NULL DEFAULT 'SYSTEM' AFTER assistant,
    ADD COLUMN IF NOT EXISTS request_type VARCHAR(50) NOT NULL DEFAULT 'COMPLETION' AFTER module,
    ADD COLUMN IF NOT EXISTS runtime_route VARCHAR(255) NULL AFTER request_type,
    ADD COLUMN IF NOT EXISTS input_tokens INT NULL DEFAULT 0 AFTER duration_ms,
    ADD COLUMN IF NOT EXISTS output_tokens INT NULL DEFAULT 0 AFTER input_tokens,
    ADD COLUMN IF NOT EXISTS total_tokens INT NULL DEFAULT 0 AFTER output_tokens,
    ADD COLUMN IF NOT EXISTS estimated_cost DECIMAL(10, 6) NULL DEFAULT 0.000000 AFTER total_tokens,
    ADD COLUMN IF NOT EXISTS cost_currency VARCHAR(10) NOT NULL DEFAULT 'USD' AFTER estimated_cost,
    ADD COLUMN IF NOT EXISTS retry_count INT NOT NULL DEFAULT 0 AFTER cost_currency,
    ADD COLUMN IF NOT EXISTS safety_status VARCHAR(50) NOT NULL DEFAULT 'PASSED' AFTER retry_count,
    ADD COLUMN IF NOT EXISTS grounding_status VARCHAR(50) NOT NULL DEFAULT 'GROUNDED' AFTER safety_status,
    ADD COLUMN IF NOT EXISTS is_memory_assisted TINYINT(1) NOT NULL DEFAULT 0 AFTER grounding_status;

-- Add performance indexes for observability querying
ALTER TABLE ai_execution_traces
    ADD INDEX IF NOT EXISTS idx_traces_org_created (org_id, created_at),
    ADD INDEX IF NOT EXISTS idx_traces_feature (org_id, feature),
    ADD INDEX IF NOT EXISTS idx_traces_status (org_id, status),
    ADD INDEX IF NOT EXISTS idx_traces_assistant (org_id, assistant);

-- 2. Configurable model and provider pricing catalog
CREATE TABLE IF NOT EXISTS ai_model_pricing (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    model_name VARCHAR(100) NOT NULL,
    input_cost_per_1k_tokens DECIMAL(10, 6) NOT NULL DEFAULT 0.000150,
    output_cost_per_1k_tokens DECIMAL(10, 6) NOT NULL DEFAULT 0.000600,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    effective_from DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_provider_model (provider, model_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. AI Quality Evaluation records
CREATE TABLE IF NOT EXISTS ai_quality_evaluations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    execution_id BIGINT NULL,
    feature VARCHAR(100) NOT NULL,
    evaluation_type VARCHAR(64) NOT NULL,
    score DECIMAL(5, 2) NOT NULL DEFAULT 1.00,
    pass_status VARCHAR(32) NOT NULL DEFAULT 'PASSED',
    evidence_status VARCHAR(32) NOT NULL DEFAULT 'SUFFICIENT',
    safety_status VARCHAR(32) NOT NULL DEFAULT 'SAFE',
    grounding_status VARCHAR(32) NOT NULL DEFAULT 'VALID',
    reviewer_type VARCHAR(32) NOT NULL DEFAULT 'AUTOMATED_CHECK',
    reviewer_id BIGINT NULL,
    evaluation_notes TEXT NULL,
    correlation_id VARCHAR(128) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_quality_org_created (org_id, created_at),
    INDEX idx_quality_feature (org_id, feature),
    INDEX idx_quality_pass_status (org_id, pass_status),
    INDEX idx_quality_correlation (correlation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. Configurable operational health thresholds
CREATE TABLE IF NOT EXISTS ai_health_thresholds (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL DEFAULT 0,
    threshold_key VARCHAR(64) NOT NULL,
    threshold_value DECIMAL(10, 2) NOT NULL,
    unit VARCHAR(32) NOT NULL,
    severity VARCHAR(32) NOT NULL DEFAULT 'WARNING',
    description VARCHAR(255) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_org_threshold (org_id, threshold_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 5. Persistent safety and security event tracking
CREATE TABLE IF NOT EXISTS ai_security_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    severity VARCHAR(32) NOT NULL DEFAULT 'HIGH',
    actor_type VARCHAR(32) NOT NULL DEFAULT 'USER',
    actor_id BIGINT NULL,
    resource_type VARCHAR(64) NULL,
    resource_id VARCHAR(128) NULL,
    sanitized_details TEXT NULL,
    correlation_id VARCHAR(128) NULL,
    ip_address VARCHAR(64) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_sec_org_created (org_id, created_at),
    INDEX idx_sec_event_type (event_type),
    INDEX idx_sec_severity (severity)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
