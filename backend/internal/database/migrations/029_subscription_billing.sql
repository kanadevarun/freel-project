-- +migrate Up

CREATE TABLE subscription_plans (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price_monthly DECIMAL(10,2) NOT NULL,
    price_annual DECIMAL(10,2) NOT NULL,
    features JSON DEFAULT (JSON_ARRAY()),
    limits JSON DEFAULT (JSON_OBJECT()),
    provider_product_id VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE billing_customers (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    provider_customer_id VARCHAR(255) NOT NULL,
    billing_email VARCHAR(255),
    tax_id VARCHAR(255),
    billing_address JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE(org_id),
    FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE TABLE billing_payment_methods (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    provider_payment_method_id VARCHAR(255) NOT NULL,
    card_brand VARCHAR(50),
    card_last4 VARCHAR(4),
    exp_month INT,
    exp_year INT,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE(provider_payment_method_id),
    FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE TABLE organization_subscriptions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    plan_id BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'inactive',
    billing_cycle VARCHAR(50) NOT NULL DEFAULT 'monthly',
    current_period_start TIMESTAMP NULL,
    current_period_end TIMESTAMP NULL,
    cancel_at_period_end BOOLEAN DEFAULT false,
    provider_subscription_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE(org_id),
    FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (plan_id) REFERENCES subscription_plans(id) ON DELETE RESTRICT
);

CREATE TABLE subscription_usage (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    current_usage INT DEFAULT 0,
    limit_amount INT,
    period_start TIMESTAMP NULL,
    period_end TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE(org_id, metric_name, period_start),
    FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE TABLE subscription_addons (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    addon_name VARCHAR(255) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    price_per_unit DECIMAL(10,2) NOT NULL,
    provider_subscription_item_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE TABLE invoices (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    provider_invoice_id VARCHAR(255),
    number VARCHAR(100),
    amount_due DECIMAL(10,2) NOT NULL,
    amount_paid DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    status VARCHAR(50) NOT NULL,
    invoice_pdf_url TEXT,
    issued_at TIMESTAMP NULL,
    paid_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE(provider_invoice_id),
    FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- Seed initial plans
INSERT INTO subscription_plans (name, description, price_monthly, price_annual, features, limits, is_active)
VALUES 
(
    'Starter',
    'Perfect for small teams getting started.',
    99.00,
    950.40,
    '["Up to 5 team members", "AI email processing (500/mo)", "Up to 100 RFQs", "Up to 100 Shipments", "Email support"]',
    '{"team_members": 5, "ai_email_processing": 500, "rfqs": 100, "shipments": 100, "carrier_connections": 1, "storage_gb": 10}',
    true
),
(
    'Growth',
    'Designed for growing freight forwarders.',
    299.00,
    2870.40,
    '["Unlimited team members", "AI email processing (2,000/mo)", "Up to 300 RFQs", "Up to 300 Shipments", "Carrier integrations", "Advanced reports"]',
    '{"team_members": -1, "ai_email_processing": 2000, "rfqs": 300, "shipments": 300, "carrier_connections": 10, "storage_gb": 100}',
    true
),
(
    'Professional',
    'Advanced features for large teams.',
    599.00,
    5750.40,
    '["Unlimited team members", "AI email processing (5,000/mo)", "Up to 1,000 RFQs", "Up to 1,000 Shipments", "Multi-branch support", "Priority support"]',
    '{"team_members": -1, "ai_email_processing": 5000, "rfqs": 1000, "shipments": 1000, "carrier_connections": 50, "storage_gb": 500}',
    true
);

-- +migrate Down

DROP TABLE IF EXISTS invoices CASCADE;
DROP TABLE IF EXISTS subscription_addons CASCADE;
DROP TABLE IF EXISTS subscription_usage CASCADE;
DROP TABLE IF EXISTS organization_subscriptions CASCADE;
DROP TABLE IF EXISTS billing_payment_methods CASCADE;
DROP TABLE IF EXISTS billing_customers CASCADE;
DROP TABLE IF EXISTS subscription_plans CASCADE;
