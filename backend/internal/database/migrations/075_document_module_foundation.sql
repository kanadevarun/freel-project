-- +migrate Up
-- 075: DOCUMENT MODULE FOUNDATION (Extending shipment_documents for General & Relational Documents)

ALTER TABLE shipment_documents MODIFY shipment_id BIGINT NULL;

ALTER TABLE shipment_documents ADD COLUMN customer_id BIGINT NULL AFTER shipment_id;
ALTER TABLE shipment_documents ADD COLUMN lead_id BIGINT NULL AFTER customer_id;
ALTER TABLE shipment_documents ADD COLUMN booking_id BIGINT NULL AFTER lead_id;

ALTER TABLE shipment_documents ADD COLUMN original_file_name VARCHAR(500) NULL AFTER file_name;
ALTER TABLE shipment_documents ADD COLUMN file_path VARCHAR(1000) NULL AFTER s3_key;
ALTER TABLE shipment_documents ADD COLUMN mime_type VARCHAR(100) NULL AFTER file_type;
ALTER TABLE shipment_documents ADD COLUMN file_size BIGINT DEFAULT 0 AFTER mime_type;

ALTER TABLE shipment_documents ADD INDEX idx_ship_docs_customer (customer_id);
ALTER TABLE shipment_documents ADD INDEX idx_ship_docs_lead (lead_id);
ALTER TABLE shipment_documents ADD INDEX idx_ship_docs_booking (booking_id);

-- +migrate Down
ALTER TABLE shipment_documents DROP INDEX idx_ship_docs_booking;
ALTER TABLE shipment_documents DROP INDEX idx_ship_docs_lead;
ALTER TABLE shipment_documents DROP INDEX idx_ship_docs_customer;

ALTER TABLE shipment_documents DROP COLUMN file_size;
ALTER TABLE shipment_documents DROP COLUMN mime_type;
ALTER TABLE shipment_documents DROP COLUMN file_path;
ALTER TABLE shipment_documents DROP COLUMN original_file_name;
ALTER TABLE shipment_documents DROP COLUMN booking_id;
ALTER TABLE shipment_documents DROP COLUMN lead_id;
ALTER TABLE shipment_documents DROP COLUMN customer_id;
