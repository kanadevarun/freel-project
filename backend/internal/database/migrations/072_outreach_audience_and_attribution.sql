-- +migrate Up

CREATE TABLE IF NOT EXISTS outreach_campaign_leads (
    campaign_id BIGINT NOT NULL,
    lead_id BIGINT NOT NULL,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (campaign_id, lead_id),
    CONSTRAINT fk_ocl_campaign FOREIGN KEY (campaign_id) REFERENCES outreach_campaigns(id) ON DELETE CASCADE,
    CONSTRAINT fk_ocl_lead FOREIGN KEY (lead_id) REFERENCES leads(id) ON DELETE CASCADE
);

ALTER TABLE leads ADD COLUMN campaign_id BIGINT NULL;
ALTER TABLE leads ADD COLUMN converted_from_outreach_at TIMESTAMP NULL;
ALTER TABLE leads ADD CONSTRAINT fk_leads_campaign FOREIGN KEY (campaign_id) REFERENCES outreach_campaigns(id) ON DELETE SET NULL;

-- +migrate Down
ALTER TABLE leads DROP FOREIGN KEY fk_leads_campaign;
ALTER TABLE leads DROP COLUMN campaign_id;
ALTER TABLE leads DROP COLUMN converted_from_outreach_at;
DROP TABLE IF EXISTS outreach_campaign_leads;
