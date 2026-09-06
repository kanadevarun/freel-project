-- +migrate Up
ALTER TABLE leads ADD COLUMN assigned_to BIGINT NULL;
ALTER TABLE leads ADD COLUMN assigned_at DATETIME NULL;
ALTER TABLE leads ADD CONSTRAINT fk_leads_assigned_to FOREIGN KEY (assigned_to) REFERENCES users(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS lead_tags (
    lead_id BIGINT NOT NULL,
    tag VARCHAR(100) NOT NULL,
    PRIMARY KEY (lead_id, tag),
    CONSTRAINT fk_lead_tags_lead_id FOREIGN KEY (lead_id) REFERENCES leads(id) ON DELETE CASCADE
);

ALTER TABLE activities ADD COLUMN user_id BIGINT NULL;
ALTER TABLE activities ADD CONSTRAINT fk_activities_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

-- +migrate Down
ALTER TABLE activities DROP FOREIGN KEY fk_activities_user_id;
ALTER TABLE activities DROP COLUMN user_id;
DROP TABLE IF EXISTS lead_tags;
ALTER TABLE leads DROP FOREIGN KEY fk_leads_assigned_to;
ALTER TABLE leads DROP COLUMN assigned_to;
ALTER TABLE leads DROP COLUMN assigned_at;
