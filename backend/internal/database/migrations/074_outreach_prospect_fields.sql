-- +migrate Up
ALTER TABLE outreach_campaign_leads ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE';
ALTER TABLE outreach_campaign_leads ADD COLUMN current_step INT NOT NULL DEFAULT 1;
ALTER TABLE outreach_campaign_leads ADD COLUMN last_activity_at TIMESTAMP NULL;
ALTER TABLE outreach_campaign_leads ADD COLUMN next_scheduled_at TIMESTAMP NULL;

-- +migrate Down
ALTER TABLE outreach_campaign_leads DROP COLUMN status;
ALTER TABLE outreach_campaign_leads DROP COLUMN current_step;
ALTER TABLE outreach_campaign_leads DROP COLUMN last_activity_at;
ALTER TABLE outreach_campaign_leads DROP COLUMN next_scheduled_at;
