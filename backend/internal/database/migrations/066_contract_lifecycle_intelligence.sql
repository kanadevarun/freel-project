-- 066_contract_lifecycle_intelligence.sql

-- 1. Immutable lifecycle intelligence events
CREATE TABLE contract_lifecycle_intelligence_events (
    id INT AUTO_INCREMENT PRIMARY KEY,
    org_id INT NOT NULL,
    contract_id INT NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    previous_state VARCHAR(50),
    new_state VARCHAR(50),
    severity VARCHAR(20) NOT NULL DEFAULT 'INFO', -- INFO, ATTENTION, WARNING, CRITICAL
    description TEXT,
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_org_contract_intel (org_id, contract_id),
    INDEX idx_org_intel_event (org_id, event_type),
    INDEX idx_org_intel_severity (org_id, severity)
);

-- 2. Renewal workflow tracking
CREATE TABLE contract_renewal_tracking (
    id INT AUTO_INCREMENT PRIMARY KEY,
    org_id INT NOT NULL,
    contract_id INT NOT NULL,
    renewal_status VARCHAR(50) NOT NULL DEFAULT 'NOT_STARTED', -- NOT_STARTED, IN_PROGRESS, RENEWED, ABANDONED
    renewal_start_date DATE,
    target_completion_date DATE,
    successor_contract_id INT,
    owner VARCHAR(100),
    notes TEXT,
    created_by VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_org_contract_renewal (org_id, contract_id),
    INDEX idx_org_renewal_status (org_id, renewal_status),
    INDEX idx_org_successor_contract (org_id, successor_contract_id)
);

-- 3. Contract lifecycle & commercial risk events
CREATE TABLE contract_risk_events (
    id INT AUTO_INCREMENT PRIMARY KEY,
    org_id INT NOT NULL,
    contract_id INT NOT NULL,
    risk_type VARCHAR(50) NOT NULL, -- e.g. EXPIRING_WITH_ACTIVE_RATES, EXPIRED_UNRESOLVED, RENEWAL_OVERDUE
    severity VARCHAR(20) NOT NULL DEFAULT 'WARNING', -- CRITICAL, WARNING, ATTENTION, INFO
    description TEXT NOT NULL,
    is_resolved BOOLEAN NOT NULL DEFAULT FALSE,
    resolved_by VARCHAR(50),
    resolved_at TIMESTAMP NULL,
    resolution_notes TEXT,
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_org_contract_risk (org_id, contract_id),
    INDEX idx_org_risk_active (org_id, is_resolved, severity)
);
