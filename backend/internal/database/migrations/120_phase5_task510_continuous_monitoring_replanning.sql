-- Migration 120: Phase 5 Task 5.10 - Continuous Monitoring and Replanning
-- Enhances autonomous_plans with health tracking, assumption monitoring, and version lineages.
-- Adds ai_monitoring_events for event deduplication, deterministic filtering, and material change detection.

ALTER TABLE autonomous_plans ADD COLUMN IF NOT EXISTS plan_health VARCHAR(40) NOT NULL DEFAULT 'HEALTHY' COMMENT 'HEALTHY, AT_RISK, STALE, BLOCKED, FAILED, COMPLETED, REPLANNING, ESCALATED, WAITING';
ALTER TABLE autonomous_plans ADD COLUMN IF NOT EXISTS health_reason TEXT NULL COMMENT 'Human-readable explanation of plan health status';
ALTER TABLE autonomous_plans ADD COLUMN IF NOT EXISTS changed_assumptions JSON NULL COMMENT 'List or map of assumptions that were invalidated';
ALTER TABLE autonomous_plans ADD COLUMN IF NOT EXISTS replan_count INT NOT NULL DEFAULT 0 COMMENT 'Count of replans for loop prevention';
ALTER TABLE autonomous_plans ADD COLUMN IF NOT EXISTS last_monitored_at DATETIME NULL COMMENT 'Timestamp when plan was last checked against business events';

CREATE TABLE IF NOT EXISTS ai_monitoring_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    event_id VARCHAR(100) NOT NULL UNIQUE,
    correlation_id VARCHAR(128) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    entity_type VARCHAR(64) NOT NULL,
    entity_id VARCHAR(64) NOT NULL,
    source VARCHAR(64) NOT NULL DEFAULT 'SYSTEM',
    event_payload JSON NOT NULL,
    is_material BOOLEAN NOT NULL DEFAULT FALSE,
    filter_reason VARCHAR(255) NULL,
    ai_evaluated BOOLEAN NOT NULL DEFAULT FALSE,
    replan_triggered BOOLEAN NOT NULL DEFAULT FALSE,
    plan_id VARCHAR(100) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_ame_org_entity (org_id, entity_type, entity_id),
    INDEX idx_ame_org_created (org_id, created_at DESC),
    INDEX idx_ame_org_material (org_id, is_material),
    INDEX idx_ame_plan (org_id, plan_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
