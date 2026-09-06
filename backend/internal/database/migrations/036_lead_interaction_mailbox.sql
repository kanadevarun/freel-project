ALTER TABLE lead_interactions
ADD COLUMN mailbox_id BIGINT NULL;

ALTER TABLE lead_interactions
ADD CONSTRAINT fk_lead_interactions_mailbox
FOREIGN KEY (mailbox_id)
REFERENCES org_connected_mailboxes(id)
ON DELETE SET NULL;
