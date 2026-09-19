-- ==============================================================================
-- LogisticsHQ Migration 119: Phase 5 Task 5.9 Multi-Step AI Planning and Execution
-- ==============================================================================

-- 1. Extend autonomous_plans with goal_type, current_step_id, stop_conditions, priority
ALTER TABLE autonomous_plans
  ADD COLUMN IF NOT EXISTS goal_type VARCHAR(64) NOT NULL DEFAULT 'OPERATIONAL' AFTER goal,
  ADD COLUMN IF NOT EXISTS current_step_id VARCHAR(64) NULL AFTER status,
  ADD COLUMN IF NOT EXISTS stop_conditions JSON NULL AFTER replan_reason,
  ADD COLUMN IF NOT EXISTS fallback_strategy TEXT NULL AFTER stop_conditions,
  ADD COLUMN IF NOT EXISTS priority VARCHAR(32) NOT NULL DEFAULT 'MEDIUM' AFTER risk_level;

-- 2. Extend autonomous_plan_steps with preconditions, retry_policy, compensation_action
ALTER TABLE autonomous_plan_steps
  ADD COLUMN IF NOT EXISTS preconditions JSON NULL AFTER condition_predicate,
  ADD COLUMN IF NOT EXISTS retry_policy JSON NULL AFTER max_attempts,
  ADD COLUMN IF NOT EXISTS compensation_action JSON NULL AFTER fallback_action;

-- 3. Concurrent entity planning conflict index
CREATE INDEX IF NOT EXISTS idx_plans_entity_conflict 
  ON autonomous_plans (org_id, related_entity_type, related_entity_id, status);

-- 4. Multi-step cross-module planning index
CREATE INDEX IF NOT EXISTS idx_plans_module_status
  ON autonomous_plans (org_id, module, status, priority);
