-- Migration 087: Unified AI Runtime, Prompt Management, and Observability Foundation
-- Defines authoritative prompt templates and runtime execution telemetry tables.

CREATE TABLE IF NOT EXISTS ai_prompt_templates (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id INT NULL COMMENT 'NULL for system-wide defaults, or specific organization ID',
    prompt_key VARCHAR(100) NOT NULL COMMENT 'Canonical identifier, e.g. pricing.analyst',
    version VARCHAR(20) NOT NULL DEFAULT '1.0.0' COMMENT 'Semantic version e.g. 1.0.0',
    workflow_name VARCHAR(100) NOT NULL COMMENT 'Associated agent or workflow',
    purpose TEXT NOT NULL COMMENT 'Business purpose and usage description',
    template_content MEDIUMTEXT NOT NULL COMMENT 'Prompt text template with placeholder variables',
    input_variables JSON NULL COMMENT 'List of expected input variable names',
    output_format VARCHAR(50) NOT NULL DEFAULT 'JSON' COMMENT 'JSON, TEXT, MARKDOWN',
    is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT '1 if active, 0 if disabled',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_prompt_key_version_org (prompt_key, version, org_id),
    INDEX idx_apt_workflow (workflow_name),
    INDEX idx_apt_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ai_execution_traces (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id INT NOT NULL,
    task_id BIGINT NULL COMMENT 'Linked ai_processing_tasks ID if queued',
    thread_id VARCHAR(150) NULL COMMENT 'LangGraph or execution thread ID',
    request_id VARCHAR(100) NOT NULL COMMENT 'Unique request execution ID',
    correlation_id VARCHAR(100) NULL COMMENT 'End-to-end tracing correlation ID',
    workflow_name VARCHAR(100) NOT NULL COMMENT 'Pricing, Sales, Operations, Contracts, etc.',
    prompt_key VARCHAR(100) NULL,
    prompt_version VARCHAR(20) NULL,
    primary_provider VARCHAR(50) NOT NULL DEFAULT 'gemini',
    primary_model VARCHAR(100) NOT NULL,
    final_provider VARCHAR(50) NOT NULL,
    final_model VARCHAR(100) NOT NULL,
    failover_occurred TINYINT(1) NOT NULL DEFAULT 0,
    failover_reason TEXT NULL,
    is_mock TINYINT(1) NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'STARTED' COMMENT 'STARTED, COMPLETED, FAILED, CANCELLED',
    duration_ms INT NULL,
    error_category VARCHAR(100) NULL COMMENT 'Taxonomy category e.g. provider_rate_limit',
    error_message TEXT NULL,
    execution_metadata JSON NULL COMMENT 'Sanitized telemetry metadata (no secrets/tokens)',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME NULL,
    INDEX idx_aet_org_created (org_id, created_at),
    INDEX idx_aet_thread (thread_id),
    INDEX idx_aet_task (task_id),
    INDEX idx_aet_correlation (correlation_id),
    INDEX idx_aet_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
