-- Migration 088: AI Evaluation Results and Safety Gate History
-- Stores deterministic test evaluation results, safety gate outcomes, and release verification runs.

CREATE TABLE IF NOT EXISTS ai_evaluation_results (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id INT NOT NULL,
    test_run_id VARCHAR(100) NOT NULL COMMENT 'Unique identifier for the evaluation run',
    scenario_id VARCHAR(100) NOT NULL COMMENT 'Unique scenario identifier, e.g. PRICING_001_NORMAL',
    scenario_name VARCHAR(150) NOT NULL,
    scenario_version VARCHAR(20) NOT NULL DEFAULT '1.0.0',
    agent_key VARCHAR(50) NOT NULL COMMENT 'pricing, sales, operations, contracts, compliance, finance, leads, outreach, system',
    category VARCHAR(50) NOT NULL COMMENT 'safety_gate, agent_eval, business_invariant',
    status VARCHAR(20) NOT NULL COMMENT 'PASS, FAIL, SKIPPED',
    expected_result TEXT NOT NULL,
    actual_result_summary TEXT NOT NULL,
    error_category VARCHAR(100) NULL,
    provider_mode VARCHAR(50) NOT NULL DEFAULT 'deterministic_test' COMMENT 'deterministic_test, mock, live',
    prompt_version VARCHAR(20) NULL,
    action_names_invoked JSON NULL COMMENT 'Array of action names executed during scenario',
    approval_state VARCHAR(50) NULL COMMENT 'NONE, PENDING, APPROVED, REJECTED, EXPIRED, CANCELLED',
    task_state VARCHAR(50) NULL COMMENT 'Canonical workforce status: completed, waiting_for_approval, failed, etc.',
    checkpoint_state VARCHAR(50) NULL COMMENT 'NONE, SAVED, INTERRUPTED, RESUMED',
    correlation_id VARCHAR(100) NULL,
    duration_ms INT NOT NULL DEFAULT 0,
    started_at DATETIME(3) NOT NULL,
    completed_at DATETIME(3) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_aer_org_run (org_id, test_run_id),
    INDEX idx_aer_scenario (scenario_id, scenario_version),
    INDEX idx_aer_status (status),
    INDEX idx_aer_agent (agent_key),
    INDEX idx_aer_category (category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
