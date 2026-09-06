-- +migrate Up

CREATE TABLE addon_configs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    pricing_model VARCHAR(50) NOT NULL DEFAULT 'per_unit',
    unit_price DECIMAL(10,2) NOT NULL,
    is_recurring BOOLEAN DEFAULT true,
    available_for_plans JSON DEFAULT (JSON_ARRAY()),
    provider_product_id VARCHAR(255),
    provider_price_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

ALTER TABLE subscription_addons ADD COLUMN addon_config_id BIGINT;
ALTER TABLE subscription_addons ADD FOREIGN KEY (addon_config_id) REFERENCES addon_configs(id) ON DELETE SET NULL;

-- Seed default add-ons
INSERT INTO addon_configs (name, description, pricing_model, unit_price, is_recurring, available_for_plans) VALUES
('Additional AI Emails', 'Purchase additional AI email processing capacity.', 'per_unit', 0.05, true, '[]'),
('Additional Storage', 'Add 100 GB of extra document storage.', 'per_unit', 10.00, true, '[]'),
('Priority Support', 'Get 24/7 priority support with 1-hour response time.', 'flat_fee', 49.00, true, '[]');

-- +migrate Down

ALTER TABLE subscription_addons DROP FOREIGN KEY subscription_addons_ibfk_2;
ALTER TABLE subscription_addons DROP COLUMN addon_config_id;
DROP TABLE IF EXISTS addon_configs CASCADE;
