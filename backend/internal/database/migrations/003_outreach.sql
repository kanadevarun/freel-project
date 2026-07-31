-- +migrate Up

CREATE TABLE outreach_campaigns (
    id BIGSERIAL PRIMARY KEY,
    org_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'DRAFT',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE outreach_sequences (
    id BIGSERIAL PRIMARY KEY,
    campaign_id BIGINT NOT NULL REFERENCES outreach_campaigns(id) ON DELETE CASCADE,
    step_number INT NOT NULL,
    channel VARCHAR(50) NOT NULL, -- EMAIL, LINKEDIN
    template TEXT NOT NULL,
    delay_days INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- +migrate Down

DROP TABLE outreach_sequences;
DROP TABLE outreach_campaigns;
