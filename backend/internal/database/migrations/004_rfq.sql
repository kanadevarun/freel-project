-- +migrate Up

CREATE TABLE rfqs (
    id BIGSERIAL PRIMARY KEY,
    org_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    rfq_number VARCHAR(50) NOT NULL,
    customer_id BIGINT NOT NULL REFERENCES customers(id),
    stage VARCHAR(50) DEFAULT 'STAGE_RFQ_CREATED',
    
    -- Routing
    origin VARCHAR(255),
    destination VARCHAR(255),
    incoterms VARCHAR(50),
    target_date DATE,
    
    -- Assignments
    sales_assignee_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    pricing_assignee_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (org_id, rfq_number)
);

CREATE TABLE rfq_items (
    id BIGSERIAL PRIMARY KEY,
    rfq_id BIGINT NOT NULL REFERENCES rfqs(id) ON DELETE CASCADE,
    description TEXT,
    quantity INT NOT NULL,
    weight_kg DECIMAL(10,2),
    volume_cbm DECIMAL(10,2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE rfq_quotes (
    id BIGSERIAL PRIMARY KEY,
    rfq_id BIGINT NOT NULL REFERENCES rfqs(id) ON DELETE CASCADE,
    carrier_name VARCHAR(100) NOT NULL,
    transit_time_days INT,
    buy_price DECIMAL(15,2) NOT NULL,
    sell_price DECIMAL(15,2) NOT NULL,
    is_recommended BOOLEAN DEFAULT false,
    status VARCHAR(50) DEFAULT 'DRAFT', -- DRAFT, SUBMITTED, ACCEPTED, REJECTED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- +migrate Down

DROP TABLE rfq_quotes;
DROP TABLE rfq_items;
DROP TABLE rfqs;
