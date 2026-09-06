-- +migrate Up
ALTER TABLE shipments ADD COLUMN closure_status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE';
ALTER TABLE shipments ADD CONSTRAINT chk_shipment_closure_status CHECK (closure_status IN ('ACTIVE', 'READY_FOR_CLOSURE', 'CLOSED', 'BLOCKED_BY_EXCEPTION'));

-- +migrate Down
ALTER TABLE shipments DROP CONSTRAINT chk_shipment_closure_status;
ALTER TABLE shipments DROP COLUMN closure_status;
