-- +migrate Up

-- Add columns allowing NULL first for backward compatibility
ALTER TABLE shipment_exceptions ADD COLUMN org_id BIGINT NULL;
ALTER TABLE shipment_exceptions ADD COLUMN status VARCHAR(30) NOT NULL DEFAULT 'OPEN';
ALTER TABLE shipment_exceptions ADD COLUMN resolved_by BIGINT NULL;
ALTER TABLE shipment_exceptions ADD COLUMN resolution_notes TEXT NULL;
ALTER TABLE shipment_exceptions ADD COLUMN updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;

-- Populate org_id dynamically from linked shipments if there is existing data
UPDATE shipment_exceptions se JOIN shipments s ON se.shipment_id = s.id SET se.org_id = s.org_id;
UPDATE shipment_exceptions SET org_id = 1 WHERE org_id IS NULL;

-- Enforce NOT NULL on org_id
ALTER TABLE shipment_exceptions MODIFY COLUMN org_id BIGINT NOT NULL;

-- Foreign key check for resolved_by operator
ALTER TABLE shipment_exceptions ADD CONSTRAINT fk_se_user FOREIGN KEY (resolved_by) REFERENCES users(id) ON DELETE SET NULL;

-- Drop old check constraints
ALTER TABLE shipment_exceptions DROP CONSTRAINT chk_exception_severity;
ALTER TABLE shipment_exceptions DROP CONSTRAINT chk_exception_type;

-- Add updated check constraints
ALTER TABLE shipment_exceptions ADD CONSTRAINT chk_exception_severity CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL'));
ALTER TABLE shipment_exceptions ADD CONSTRAINT chk_exception_status CHECK (status IN ('OPEN', 'ACKNOWLEDGED', 'IN_PROGRESS', 'RESOLVED', 'DISMISSED'));
ALTER TABLE shipment_exceptions ADD CONSTRAINT chk_exception_type CHECK (exception_type IN ('SCHEDULE_DELAY', 'ETD_DELAY', 'ETA_DELAY', 'VESSEL_ROLLOVER', 'PORT_CONGESTION', 'CUSTOMS_HOLD', 'DOCUMENT_ISSUE', 'CARRIER_DELAY', 'ROUTE_DEVIATION', 'CONTAINER_ISSUE', 'OTHER'));

-- Add unique constraint to prevent duplicate processing
ALTER TABLE shipment_exceptions ADD CONSTRAINT uq_exception_type_shipment UNIQUE (shipment_id, exception_type, source_event_id);

-- +migrate Down
ALTER TABLE shipment_exceptions DROP FOREIGN KEY fk_se_user;
ALTER TABLE shipment_exceptions DROP INDEX uq_exception_type_shipment;

ALTER TABLE shipment_exceptions DROP CONSTRAINT chk_exception_status;
ALTER TABLE shipment_exceptions DROP CONSTRAINT chk_exception_severity;
ALTER TABLE shipment_exceptions DROP CONSTRAINT chk_exception_type;

ALTER TABLE shipment_exceptions DROP COLUMN org_id;
ALTER TABLE shipment_exceptions DROP COLUMN status;
ALTER TABLE shipment_exceptions DROP COLUMN resolved_by;
ALTER TABLE shipment_exceptions DROP COLUMN resolution_notes;
ALTER TABLE shipment_exceptions DROP COLUMN updated_at;

ALTER TABLE shipment_exceptions ADD CONSTRAINT chk_exception_severity CHECK (severity IN ('INFO', 'WARNING', 'CRITICAL'));
ALTER TABLE shipment_exceptions ADD CONSTRAINT chk_exception_type CHECK (exception_type IN ('ROLLOVER', 'DELAY', 'CUSTOMS_HOLD', 'PORT_CONGESTION', 'WEATHER'));
