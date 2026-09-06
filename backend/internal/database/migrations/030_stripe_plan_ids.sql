-- +migrate Up

ALTER TABLE subscription_plans
ADD COLUMN provider_price_id_monthly VARCHAR(255),
ADD COLUMN provider_price_id_annual VARCHAR(255);

-- +migrate Down

ALTER TABLE subscription_plans
DROP COLUMN provider_price_id_monthly,
DROP COLUMN provider_price_id_annual;
