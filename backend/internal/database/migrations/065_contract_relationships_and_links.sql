-- 065_contract_relationships_and_links.sql

CREATE TABLE contract_links (
    id INT AUTO_INCREMENT PRIMARY KEY,
    org_id INT NOT NULL,
    contract_id INT NOT NULL,
    linked_entity_type VARCHAR(50) NOT NULL, -- e.g., RATE, QUOTATION, SPOT_RATE_REQUEST, CUSTOMER
    linked_entity_id INT NOT NULL,
    link_type VARCHAR(50) NOT NULL, -- e.g., PRIMARY, COMMERCIAL_RATE
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    notes TEXT,
    created_by VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Ensuring a contract can't link to the exact same entity multiple times active
    UNIQUE KEY uk_contract_link (org_id, contract_id, linked_entity_type, linked_entity_id),
    
    INDEX idx_org_contract (org_id, contract_id),
    INDEX idx_org_entity (org_id, linked_entity_type, linked_entity_id)
);

CREATE TABLE contract_link_history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    org_id INT NOT NULL,
    contract_id INT NOT NULL,
    contract_link_id INT NOT NULL,
    linked_entity_type VARCHAR(50) NOT NULL,
    linked_entity_id INT NOT NULL,
    link_type VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL, -- LINKED, UNLINKED, PRIMARY_CHANGED, UPDATED
    previous_metadata JSON,
    performed_by VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_org_contract_history (org_id, contract_id)
);
