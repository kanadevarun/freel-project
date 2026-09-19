-- Migration 106: Phase 3 Task 3.9 - Advanced Notifications and Escalations
-- Extends notifications with lifecycle delivery states, acknowledgment tracking, snooze capability, semantic grouping, and AI prioritization.
-- Extends notification_escalation_events with AI reasoning, action proposals, and approval links.

-- 1. Add advanced notification lifecycle and AI fields to notifications table if they don't exist
ALTER TABLE notifications
    ADD COLUMN IF NOT EXISTS delivery_status VARCHAR(32) NOT NULL DEFAULT 'DELIVERED' COMMENT 'PENDING, QUEUED, AWAITING_APPROVAL, APPROVED, SENDING, DELIVERED, READ, ACKNOWLEDGED, SNOOZED, ESCALATED, FAILED, CANCELLED, EXPIRED',
    ADD COLUMN IF NOT EXISTS is_acknowledged TINYINT(1) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS acknowledged_at DATETIME NULL,
    ADD COLUMN IF NOT EXISTS acknowledged_by BIGINT NULL,
    ADD COLUMN IF NOT EXISTS is_snoozed TINYINT(1) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS snoozed_until DATETIME NULL,
    ADD COLUMN IF NOT EXISTS group_key VARCHAR(128) NULL COMMENT 'Semantic cluster or correlation grouping key',
    ADD COLUMN IF NOT EXISTS ai_summary TEXT NULL COMMENT 'AI generated executive operational summary',
    ADD COLUMN IF NOT EXISTS ai_escalation_reason TEXT NULL COMMENT 'AI explanation of why issue is critical/escalating',
    ADD COLUMN IF NOT EXISTS ai_priority_score FLOAT NULL COMMENT 'AI evaluated priority score from 0.00 to 100.00';

-- Indexes for performance and filtering
CREATE INDEX IF NOT EXISTS idx_notif_delivery_status ON notifications (org_id, delivery_status);
CREATE INDEX IF NOT EXISTS idx_notif_ack ON notifications (org_id, is_acknowledged);
CREATE INDEX IF NOT EXISTS idx_notif_snooze ON notifications (org_id, is_snoozed, snoozed_until);
CREATE INDEX IF NOT EXISTS idx_notif_group ON notifications (org_id, group_key);

-- 2. Add advanced escalation tracking to notification_escalation_events
ALTER TABLE notification_escalation_events
    ADD COLUMN IF NOT EXISTS ai_escalation_summary TEXT NULL COMMENT 'AI generated operational impact explanation',
    ADD COLUMN IF NOT EXISTS recommended_action VARCHAR(128) NULL COMMENT 'Recommended next-step action intent',
    ADD COLUMN IF NOT EXISTS action_proposal_id VARCHAR(128) NULL COMMENT 'Reference to Centralized Action Proposal if created',
    ADD COLUMN IF NOT EXISTS approval_id BIGINT NULL COMMENT 'Reference to approval_requests if high-risk action requires HITL approval';

CREATE INDEX IF NOT EXISTS idx_esc_approval ON notification_escalation_events (org_id, approval_id);
