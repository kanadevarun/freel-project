-- Migration 071: Customer Intelligence, Health & Risk Management

-- 1. Customer Health Evaluations Table
CREATE TABLE IF NOT EXISTS customer_health_evaluations (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  org_id BIGINT NOT NULL,
  customer_id BIGINT NOT NULL,
  health_status VARCHAR(50) NOT NULL DEFAULT 'INSUFFICIENT_DATA',
  health_score INT NOT NULL DEFAULT 50,
  contributing_factors_json TEXT NULL,
  evaluated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_che_org_cust (org_id, customer_id),
  INDEX idx_che_status (org_id, health_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. Customer Risk Events Table
CREATE TABLE IF NOT EXISTS customer_risk_events (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  org_id BIGINT NOT NULL,
  customer_id BIGINT NOT NULL,
  risk_type VARCHAR(100) NOT NULL,
  severity VARCHAR(50) NOT NULL DEFAULT 'ATTENTION',
  title VARCHAR(255) NOT NULL,
  description TEXT NULL,
  detected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  is_resolved BOOLEAN NOT NULL DEFAULT FALSE,
  resolved_at TIMESTAMP NULL,
  resolved_by BIGINT NULL,
  resolution_note TEXT NULL,
  INDEX idx_cre_org_cust (org_id, customer_id),
  INDEX idx_cre_org_resolved (org_id, is_resolved, severity)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Customer Opportunity Events Table
CREATE TABLE IF NOT EXISTS customer_opportunity_events (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  org_id BIGINT NOT NULL,
  customer_id BIGINT NOT NULL,
  opportunity_type VARCHAR(100) NOT NULL,
  priority VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
  title VARCHAR(255) NOT NULL,
  reason TEXT NULL,
  suggested_action TEXT NULL,
  related_record_code VARCHAR(100) NULL,
  detected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_coe_org_cust (org_id, customer_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. Customer Intelligence Audit Events Table
CREATE TABLE IF NOT EXISTS customer_intelligence_events (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  org_id BIGINT NOT NULL,
  customer_id BIGINT NOT NULL,
  event_type VARCHAR(100) NOT NULL,
  title VARCHAR(255) NOT NULL,
  description TEXT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_cie_org_cust (org_id, customer_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
