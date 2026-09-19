-- 127_sportal_platform_settings.sql
-- Task S17: SPortal Settings, Platform Administration, Organization Controls, Platform Configuration & Operational Administration

CREATE TABLE IF NOT EXISTS sportal_platform_settings (
    setting_key VARCHAR(64) NOT NULL PRIMARY KEY,
    setting_value TEXT NOT NULL,
    category VARCHAR(32) NOT NULL DEFAULT 'PLATFORM',
    data_type VARCHAR(20) NOT NULL DEFAULT 'STRING',
    description VARCHAR(255) NULL,
    is_sensitive BOOLEAN NOT NULL DEFAULT FALSE,
    updated_by VARCHAR(128) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

INSERT INTO sportal_platform_settings (setting_key, setting_value, category, data_type, description, is_sensitive, updated_by)
VALUES
('platform_name', 'LogisticsHQ Enterprise SaaS Control Center', 'PLATFORM', 'STRING', 'Primary brand name of the internal SaaS platform', FALSE, 'system'),
('maintenance_mode', 'false', 'PLATFORM', 'BOOLEAN', 'Global maintenance mode disabling external mutations', FALSE, 'system'),
('maintenance_banner_text', '', 'PLATFORM', 'STRING', 'Broadcast announcement banner displayed across all portals', FALSE, 'system'),
('default_currency', 'USD', 'DEFAULTS', 'STRING', 'Default platform billing currency', FALSE, 'system'),
('default_timezone', 'UTC', 'DEFAULTS', 'STRING', 'Platform default timezone for operational timestamps', FALSE, 'system'),
('default_measurement', 'Metric (kg, cm, km)', 'DEFAULTS', 'STRING', 'Default unit of measurement across freight calculations', FALSE, 'system'),
('session_timeout_minutes', '120', 'SECURITY', 'INTEGER', 'Inactivity duration before requiring re-authentication', FALSE, 'system'),
('max_login_attempts', '5', 'SECURITY', 'INTEGER', 'Failed login attempts before temporary account lock', FALSE, 'system'),
('password_min_length', '10', 'SECURITY', 'INTEGER', 'Minimum length required for internal staff passwords', FALSE, 'system'),
('require_mfa_internal', 'false', 'SECURITY', 'BOOLEAN', 'Mandatory multi-factor authentication for internal staff', FALSE, 'system'),
('ai_global_enabled', 'true', 'AI_AUTONOMY', 'BOOLEAN', 'Master switch for AI reasoning and copilot services', FALSE, 'system'),
('event_mesh_dead_letter_alert_threshold', '25', 'OPERATIONS', 'INTEGER', 'Dead letter queue size triggering critical alerts', FALSE, 'system'),
('external_carrier_sync_interval_sec', '300', 'OPERATIONS', 'INTEGER', 'Frequency of background EDI/API sync jobs in seconds', FALSE, 'system')
ON DUPLICATE KEY UPDATE updated_at = CURRENT_TIMESTAMP;
