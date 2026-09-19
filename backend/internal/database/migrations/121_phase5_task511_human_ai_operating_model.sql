-- Phase 5 Task 5.11: Human + AI Operating Model Migration
-- Establishes authoritative schema for human + AI collaboration, decision provenance,
-- human overrides, human edits, stop controls, and approval invalidation tracking.

CREATE TABLE IF NOT EXISTS human_ai_decisions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    decision_id VARCHAR(100) NOT NULL UNIQUE,
    correlation_id VARCHAR(100) NOT NULL,
    plan_id VARCHAR(100) NULL,
    step_id VARCHAR(100) NULL,
    approval_id BIGINT NULL,
    module VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(100) NOT NULL,
    operating_mode VARCHAR(50) NOT NULL DEFAULT 'HUMAN_REVIEW',
    autonomy_level VARCHAR(50) NOT NULL DEFAULT 'LEVEL_2',
    title VARCHAR(255) NOT NULL,
    context_summary TEXT NOT NULL,
    facts JSON NOT NULL,
    predictions JSON NULL,
    ai_recommendation TEXT NOT NULL,
    original_ai_payload LONGTEXT NULL,
    human_edited_payload LONGTEXT NULL,
    alternatives JSON NULL,
    confidence VARCHAR(20) NOT NULL DEFAULT 'HIGH',
    data_sufficiency VARCHAR(30) NOT NULL DEFAULT 'SUFFICIENT',
    risk_level VARCHAR(20) NOT NULL DEFAULT 'MEDIUM',
    is_reversible TINYINT(1) NOT NULL DEFAULT 1,
    decision_status VARCHAR(40) NOT NULL DEFAULT 'PENDING',
    human_decision VARCHAR(50) NULL,
    decision_reason TEXT NULL,
    decided_by_id BIGINT NULL,
    decided_by_name VARCHAR(100) NULL,
    decided_at DATETIME NULL,
    feedback_type VARCHAR(50) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_human_ai_org_status (org_id, decision_status),
    INDEX idx_human_ai_entity (org_id, module, entity_type, entity_id),
    INDEX idx_human_ai_plan (plan_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Extend autonomous_plans with human intervention and stop control columns if not exist
SET @col_exists = (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'autonomous_plans' AND column_name = 'operating_mode');
SET @sql = IF(@col_exists = 0, 'ALTER TABLE autonomous_plans ADD COLUMN operating_mode VARCHAR(50) NOT NULL DEFAULT \'AI_EXECUTE\' AFTER status', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'autonomous_plans' AND column_name = 'stopped_by_user_id');
SET @sql = IF(@col_exists = 0, 'ALTER TABLE autonomous_plans ADD COLUMN stopped_by_user_id BIGINT NULL AFTER replan_count, ADD COLUMN stopped_at DATETIME NULL AFTER stopped_by_user_id, ADD COLUMN stop_reason TEXT NULL AFTER stopped_at', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists = (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'autonomous_plans' AND column_name = 'human_modified');
SET @sql = IF(@col_exists = 0, 'ALTER TABLE autonomous_plans ADD COLUMN human_modified TINYINT(1) NOT NULL DEFAULT 0 AFTER stop_reason', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Extend autonomous_plan_steps with human edit tracking
SET @col_exists = (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'autonomous_plan_steps' AND column_name = 'human_edited_content');
SET @sql = IF(@col_exists = 0, 'ALTER TABLE autonomous_plan_steps ADD COLUMN human_edited_content LONGTEXT NULL AFTER parameters', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Extend approval_requests with invalidation reason if not exists
SET @col_exists = (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'approval_requests' AND column_name = 'invalidation_reason');
SET @sql = IF(@col_exists = 0, 'ALTER TABLE approval_requests ADD COLUMN invalidation_reason TEXT NULL AFTER returned_reason', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
