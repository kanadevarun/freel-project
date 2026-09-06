ALTER TABLE carrier_integrations
ADD COLUMN connection_method VARCHAR(50) DEFAULT 'API',
ADD COLUMN credentials_json JSON,
ADD COLUMN environment VARCHAR(50) DEFAULT 'Production',
ADD COLUMN connection_status VARCHAR(50) DEFAULT 'Connected',
ADD COLUMN sync_status VARCHAR(255),
ADD COLUMN last_synced_at DATETIME,
ADD COLUMN capabilities JSON;
