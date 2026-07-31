-- +migrate Up

CREATE TABLE opportunities (
    id BIGSERIAL PRIMARY KEY,
    org_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    customer_id BIGINT NOT NULL REFERENCES customers(id),
    title VARCHAR(255) NOT NULL,
    amount DECIMAL(15,2),
    stage VARCHAR(50) DEFAULT 'PROSPECTING',
    expected_close_date DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- +migrate Down

DROP TABLE opportunities;
