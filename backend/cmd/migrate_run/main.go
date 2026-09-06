package main

import (
	"log"
	"os"
	
	"github.com/freel/backend/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	db, err := database.Connect(os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	_, err = db.Exec(`ALTER TABLE leads ADD COLUMN assigned_to BIGINT NULL`)
	if err != nil {
		log.Printf("Migration info: assigned_to column: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE leads ADD COLUMN assigned_at DATETIME NULL`)
	if err != nil {
		log.Printf("Migration info: assigned_at column: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE leads ADD CONSTRAINT fk_leads_assigned_to FOREIGN KEY (assigned_to) REFERENCES users(id) ON DELETE SET NULL`)
	if err != nil {
		log.Printf("Migration info: constraint fk_leads_assigned_to: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS lead_tags (
		lead_id BIGINT NOT NULL,
		tag VARCHAR(100) NOT NULL,
		PRIMARY KEY (lead_id, tag),
		CONSTRAINT fk_lead_tags_lead_id FOREIGN KEY (lead_id) REFERENCES leads(id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.Printf("Migration info: lead_tags table: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE activities ADD COLUMN user_id BIGINT NULL`)
	if err != nil {
		log.Printf("Migration info: user_id column: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE activities ADD CONSTRAINT fk_activities_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL`)
	if err != nil {
		log.Printf("Migration info: constraint fk_activities_user_id: %v", err)
	}

	// Connected Mailboxes extensions
	_, err = db.Exec(`ALTER TABLE org_connected_mailboxes ADD COLUMN provider VARCHAR(50) NOT NULL DEFAULT 'IMAP'`)
	if err != nil {
		log.Printf("Migration info: org_connected_mailboxes provider: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE org_connected_mailboxes ADD COLUMN access_token_encrypted TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: org_connected_mailboxes access_token_encrypted: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE org_connected_mailboxes ADD COLUMN refresh_token_encrypted TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: org_connected_mailboxes refresh_token_encrypted: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE org_connected_mailboxes ADD COLUMN token_expiry TIMESTAMP NULL`)
	if err != nil {
		log.Printf("Migration info: org_connected_mailboxes token_expiry: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE org_connected_mailboxes ADD COLUMN oauth_scopes TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: org_connected_mailboxes oauth_scopes: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE org_connected_mailboxes ADD COLUMN sync_cursor VARCHAR(255) NULL`)
	if err != nil {
		log.Printf("Migration info: org_connected_mailboxes sync_cursor: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE org_connected_mailboxes ADD COLUMN last_sync_started_at TIMESTAMP NULL`)
	if err != nil {
		log.Printf("Migration info: org_connected_mailboxes last_sync_started_at: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE org_connected_mailboxes ADD COLUMN last_sync_success_at TIMESTAMP NULL`)
	if err != nil {
		log.Printf("Migration info: org_connected_mailboxes last_sync_success_at: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE org_connected_mailboxes ADD COLUMN last_sync_error TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: org_connected_mailboxes last_sync_error: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE org_connected_mailboxes ADD COLUMN last_processed_message_id VARCHAR(255) NULL`)
	if err != nil {
		log.Printf("Migration info: org_connected_mailboxes last_processed_message_id: %v", err)
	}

	// Lead Interactions extensions
	_, err = db.Exec(`ALTER TABLE lead_interactions ADD COLUMN rfc_message_id VARCHAR(255) NULL`)
	if err != nil {
		log.Printf("Migration info: lead_interactions rfc_message_id: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE lead_interactions ADD COLUMN in_reply_to VARCHAR(255) NULL`)
	if err != nil {
		log.Printf("Migration info: lead_interactions in_reply_to: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE lead_interactions ADD COLUMN references_header TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: lead_interactions references_header: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE lead_interactions ADD COLUMN sender VARCHAR(255) NULL`)
	if err != nil {
		log.Printf("Migration info: lead_interactions sender: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE lead_interactions ADD COLUMN recipients TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: lead_interactions recipients: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE lead_interactions ADD COLUMN cc_recipients TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: lead_interactions cc_recipients: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE lead_interactions ADD COLUMN mailbox_id BIGINT NULL`)
	if err != nil {
		log.Printf("Migration info: lead_interactions mailbox_id: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE lead_interactions ADD CONSTRAINT fk_lead_interactions_mailbox FOREIGN KEY (mailbox_id) REFERENCES org_connected_mailboxes(id) ON DELETE SET NULL`)
	if err != nil {
		log.Printf("Migration info: constraint fk_lead_interactions_mailbox: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE lead_interactions ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT 'SENT'`)
	if err != nil {
		log.Printf("Migration info: lead_interactions status: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE lead_interactions ADD COLUMN retry_count INT NOT NULL DEFAULT 0`)
	if err != nil {
		log.Printf("Migration info: lead_interactions retry_count: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE lead_interactions ADD COLUMN last_retry_at DATETIME NULL`)
	if err != nil {
		log.Printf("Migration info: lead_interactions last_retry_at: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE lead_interactions ADD COLUMN last_error TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: lead_interactions last_error: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS lead_email_drafts (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		lead_id BIGINT NOT NULL,
		parent_interaction_id BIGINT NOT NULL,
		mailbox_id BIGINT NULL,
		recipients TEXT NULL,
		cc_recipients TEXT NULL,
		subject VARCHAR(500) NULL,
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		UNIQUE KEY unique_lead_parent_draft (org_id, lead_id, parent_interaction_id)
	)`)
	if err != nil {
		log.Printf("Migration info: lead_email_drafts table: %v", err)
	}

	// Shipment Exceptions Task 16.5 extensions
	_, err = db.Exec(`ALTER TABLE shipment_exceptions ADD COLUMN org_id BIGINT NULL`)
	if err == nil {
		_, _ = db.Exec(`UPDATE shipment_exceptions se JOIN shipments s ON se.shipment_id = s.id SET se.org_id = s.org_id`)
		_, _ = db.Exec(`UPDATE shipment_exceptions SET org_id = 1 WHERE org_id IS NULL`)
		_, _ = db.Exec(`ALTER TABLE shipment_exceptions MODIFY COLUMN org_id BIGINT NOT NULL`)
	} else {
		log.Printf("Migration info: org_id column: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE shipment_exceptions ADD COLUMN status VARCHAR(30) NOT NULL DEFAULT 'OPEN'`)
	if err != nil {
		log.Printf("Migration info: status column: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE shipment_exceptions ADD COLUMN resolved_by BIGINT NULL`)
	if err == nil {
		_, _ = db.Exec(`ALTER TABLE shipment_exceptions ADD CONSTRAINT fk_se_user FOREIGN KEY (resolved_by) REFERENCES users(id) ON DELETE SET NULL`)
	} else {
		log.Printf("Migration info: resolved_by column: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE shipment_exceptions ADD COLUMN resolution_notes TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: resolution_notes column: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE shipment_exceptions ADD COLUMN updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP`)
	if err != nil {
		log.Printf("Migration info: updated_at column: %v", err)
	}

	// Drop and update CHECK constraints
	_, _ = db.Exec(`ALTER TABLE shipment_exceptions DROP CONSTRAINT chk_exception_severity`)
	_, _ = db.Exec(`ALTER TABLE shipment_exceptions DROP CONSTRAINT chk_exception_type`)

	_, err = db.Exec(`ALTER TABLE shipment_exceptions ADD CONSTRAINT chk_exception_severity CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL'))`)
	if err != nil {
		log.Printf("Migration info: constraint chk_exception_severity: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE shipment_exceptions ADD CONSTRAINT chk_exception_status CHECK (status IN ('OPEN', 'ACKNOWLEDGED', 'IN_PROGRESS', 'RESOLVED', 'DISMISSED'))`)
	if err != nil {
		log.Printf("Migration info: constraint chk_exception_status: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE shipment_exceptions ADD CONSTRAINT chk_exception_type CHECK (exception_type IN ('SCHEDULE_DELAY', 'ETD_DELAY', 'ETA_DELAY', 'VESSEL_ROLLOVER', 'PORT_CONGESTION', 'CUSTOMS_HOLD', 'DOCUMENT_ISSUE', 'CARRIER_DELAY', 'ROUTE_DEVIATION', 'CONTAINER_ISSUE', 'OTHER'))`)
	if err != nil {
		log.Printf("Migration info: constraint chk_exception_type: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE shipment_exceptions ADD CONSTRAINT uq_exception_type_shipment UNIQUE (shipment_id, exception_type, source_event_id)`)
	if err != nil {
		log.Printf("Migration info: constraint uq_exception_type_shipment: %v", err)
	}

	// Task 18.6 Quotations Conversion extensions
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN converted_at DATETIME NULL`)
	if err != nil {
		log.Printf("Migration info: quotations converted_at: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN converted_by VARCHAR(255) NULL`)
	if err != nil {
		log.Printf("Migration info: quotations converted_by: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN converted_booking_id BIGINT NULL`)
	if err != nil {
		log.Printf("Migration info: quotations converted_booking_id: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN converted_shipment_id BIGINT NULL`)
	if err != nil {
		log.Printf("Migration info: quotations converted_shipment_id: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN conversion_status VARCHAR(50) NOT NULL DEFAULT 'NOT_CONVERTED'`)
	if err != nil {
		log.Printf("Migration info: quotations conversion_status: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN conversion_notes TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: quotations conversion_notes: %v", err)
	}

	_, _ = db.Exec(`ALTER TABLE quotations ADD INDEX idx_quotation_org_booking (org_id, converted_booking_id)`)
	_, _ = db.Exec(`ALTER TABLE quotations ADD INDEX idx_quotation_org_shipment (org_id, converted_shipment_id)`)
	_, _ = db.Exec(`ALTER TABLE quotations ADD INDEX idx_quotation_org_conv_status (org_id, conversion_status)`)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS quotation_conversion_history (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		quotation_id BIGINT NOT NULL,
		booking_id BIGINT NULL,
		shipment_id BIGINT NULL,
		action VARCHAR(50) NOT NULL,
		status VARCHAR(50) NOT NULL,
		message TEXT NULL,
		performed_by VARCHAR(255) NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_qch_org_quote_created (org_id, quotation_id, created_at DESC),
		INDEX idx_qch_org_booking (org_id, booking_id),
		CONSTRAINT fk_quotation_conversion_history_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
		CONSTRAINT fk_quotation_conversion_history_quote FOREIGN KEY (quotation_id) REFERENCES quotations(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: quotation_conversion_history table: %v", err)
	}

	// Task 18.2 Pricing & Margins
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN total_cost DECIMAL(15,2) NOT NULL DEFAULT 0.00`)
	if err != nil {
		log.Printf("Migration info: quotations total_cost: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN gross_profit DECIMAL(15,2) NOT NULL DEFAULT 0.00`)
	if err != nil {
		log.Printf("Migration info: quotations gross_profit: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN gross_margin_pct DECIMAL(8,4) NOT NULL DEFAULT 0.0000`)
	if err != nil {
		log.Printf("Migration info: quotations gross_margin_pct: %v", err)
	}

	// Task 18.3 Templates & Commercial Terms
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN payment_terms VARCHAR(50) NOT NULL DEFAULT 'PREPAID'`)
	if err != nil {
		log.Printf("Migration info: quotations payment_terms: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN commercial_terms TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: quotations commercial_terms: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN customer_notes TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: quotations customer_notes: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN internal_notes TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: quotations internal_notes: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN template_id BIGINT NULL`)
	if err != nil {
		log.Printf("Migration info: quotations template_id: %v", err)
	}

	// Task 18.4 Lifecycle & Review
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN submitted_for_review_at DATETIME NULL`)
	if err != nil {
		log.Printf("Migration info: quotations submitted_for_review_at: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN submitted_for_review_by VARCHAR(255) NULL`)
	if err != nil {
		log.Printf("Migration info: quotations submitted_for_review_by: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN approved_at DATETIME NULL`)
	if err != nil {
		log.Printf("Migration info: quotations approved_at: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN approved_by VARCHAR(255) NULL`)
	if err != nil {
		log.Printf("Migration info: quotations approved_by: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN approval_notes TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: quotations approval_notes: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN changes_requested_at DATETIME NULL`)
	if err != nil {
		log.Printf("Migration info: quotations changes_requested_at: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN changes_requested_by VARCHAR(255) NULL`)
	if err != nil {
		log.Printf("Migration info: quotations changes_requested_by: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN changes_requested_reason TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: quotations changes_requested_reason: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN sent_by VARCHAR(255) NULL`)
	if err != nil {
		log.Printf("Migration info: quotations sent_by: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN first_viewed_at DATETIME NULL`)
	if err != nil {
		log.Printf("Migration info: quotations first_viewed_at: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN last_viewed_at DATETIME NULL`)
	if err != nil {
		log.Printf("Migration info: quotations last_viewed_at: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN view_count INT NOT NULL DEFAULT 0`)
	if err != nil {
		log.Printf("Migration info: quotations view_count: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN declined_at DATETIME NULL`)
	if err != nil {
		log.Printf("Migration info: quotations declined_at: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN declined_reason TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: quotations declined_reason: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN cancelled_at DATETIME NULL`)
	if err != nil {
		log.Printf("Migration info: quotations cancelled_at: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN cancelled_by VARCHAR(255) NULL`)
	if err != nil {
		log.Printf("Migration info: quotations cancelled_by: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN cancelled_reason TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: quotations cancelled_reason: %v", err)
	}

	// Task 18.6 Quotation-to-Booking Operational Conversion
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN converted_at DATETIME NULL`)
	if err != nil {
		log.Printf("Migration info: quotations converted_at: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN converted_by VARCHAR(255) NULL`)
	if err != nil {
		log.Printf("Migration info: quotations converted_by: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN converted_booking_id BIGINT NULL`)
	if err != nil {
		log.Printf("Migration info: quotations converted_booking_id: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN converted_shipment_id BIGINT NULL`)
	if err != nil {
		log.Printf("Migration info: quotations converted_shipment_id: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN conversion_status VARCHAR(50) NOT NULL DEFAULT 'NOT_CONVERTED'`)
	if err != nil {
		log.Printf("Migration info: quotations conversion_status: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE quotations ADD COLUMN conversion_notes TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: quotations conversion_notes: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS quotation_conversion_history (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		quotation_id BIGINT NOT NULL,
		booking_id BIGINT NULL,
		shipment_id BIGINT NULL,
		action VARCHAR(50) NOT NULL,
		status VARCHAR(50) NOT NULL,
		message TEXT NULL,
		performed_by VARCHAR(255) NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_qch_org_quote_created (org_id, quotation_id, created_at DESC),
		INDEX idx_qch_org_booking (org_id, booking_id),
		CONSTRAINT fk_quotation_conversion_history_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
		CONSTRAINT fk_quotation_conversion_history_quote FOREIGN KEY (quotation_id) REFERENCES quotations(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: quotation_conversion_history table: %v", err)
	}

	// Task 18.7 Booking Confirmation, Commercial Handover & Lineage Traceability
	_, err = db.Exec(`ALTER TABLE bookings ADD COLUMN source_quotation_id BIGINT NULL`)
	if err != nil {
		log.Printf("Migration info: bookings source_quotation_id: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE bookings ADD COLUMN source_quote_number VARCHAR(100) NULL`)
	if err != nil {
		log.Printf("Migration info: bookings source_quote_number: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE bookings ADD COLUMN commercial_snapshot_at DATETIME NULL`)
	if err != nil {
		log.Printf("Migration info: bookings commercial_snapshot_at: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE bookings ADD COLUMN commercial_handover_status VARCHAR(50) NOT NULL DEFAULT 'PENDING'`)
	if err != nil {
		log.Printf("Migration info: bookings commercial_handover_status: %v", err)
	}

	_, _ = db.Exec(`ALTER TABLE bookings ADD INDEX idx_bookings_org_source_quote (org_id, source_quotation_id)`)
	_, _ = db.Exec(`ALTER TABLE bookings ADD INDEX idx_bookings_org_handover_status (org_id, commercial_handover_status)`)

	_, err = db.Exec(`ALTER TABLE shipments ADD COLUMN source_quotation_id BIGINT NULL`)
	if err != nil {
		log.Printf("Migration info: shipments source_quotation_id: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE shipments ADD COLUMN source_booking_id BIGINT NULL`)
	if err != nil {
		log.Printf("Migration info: shipments source_booking_id: %v", err)
	}

	_, _ = db.Exec(`ALTER TABLE shipments ADD INDEX idx_shipments_org_source_quote (org_id, source_quotation_id)`)
	_, _ = db.Exec(`ALTER TABLE shipments ADD INDEX idx_shipments_org_source_booking (org_id, source_booking_id)`)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS quotation_operational_handover_history (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		quotation_id BIGINT NOT NULL,
		booking_id BIGINT NULL,
		shipment_id BIGINT NULL,
		event_type VARCHAR(64) NOT NULL,
		description TEXT NOT NULL,
		metadata JSON NULL,
		performed_by VARCHAR(255) NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_qohh_org_quote (org_id, quotation_id, created_at DESC),
		INDEX idx_qohh_org_booking (org_id, booking_id),
		INDEX idx_qohh_org_shipment (org_id, shipment_id),
		CONSTRAINT fk_qohh_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
		CONSTRAINT fk_qohh_quote FOREIGN KEY (quotation_id) REFERENCES quotations(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: quotation_operational_handover_history table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS rates (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		rate_reference VARCHAR(100) NOT NULL,
		carrier_name VARCHAR(100) NOT NULL,
		carrier_code VARCHAR(20) NULL,
		service_provider VARCHAR(100) NULL,
		rate_type VARCHAR(30) NOT NULL DEFAULT 'SPOT',
		transport_mode VARCHAR(50) NOT NULL DEFAULT 'Ocean FCL',
		service_type VARCHAR(50) NOT NULL DEFAULT 'FCL',
		equipment_type VARCHAR(50) NULL DEFAULT '40GP',
		origin_port VARCHAR(255) NOT NULL,
		origin_code VARCHAR(20) NULL,
		destination_port VARCHAR(255) NOT NULL,
		destination_code VARCHAR(20) NULL,
		currency VARCHAR(10) NOT NULL DEFAULT 'USD',
		base_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
		effective_date DATE NOT NULL,
		expiry_date DATE NOT NULL,
		status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',
		carrier_reference VARCHAR(100) NULL,
		contract_reference VARCHAR(100) NULL,
		notes TEXT NULL,
		created_by VARCHAR(255) NULL,
		updated_by VARCHAR(255) NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_rates_org_status (org_id, status),
		INDEX idx_rates_org_carrier (org_id, carrier_name),
		INDEX idx_rates_org_lane (org_id, origin_port, destination_port),
		INDEX idx_rates_org_dates (org_id, effective_date, expiry_date),
		INDEX idx_rates_org_ref (org_id, rate_reference)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: rates table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS rate_charge_items (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		rate_id BIGINT NOT NULL,
		charge_category VARCHAR(50) NOT NULL DEFAULT 'FREIGHT',
		charge_code VARCHAR(50) NOT NULL DEFAULT '',
		charge_name VARCHAR(150) NOT NULL,
		calculation_basis VARCHAR(50) NOT NULL DEFAULT 'FLAT',
		quantity DECIMAL(12,4) NOT NULL DEFAULT 1.0000,
		unit_price DECIMAL(14,4) NOT NULL DEFAULT 0.0000,
		currency VARCHAR(10) NOT NULL DEFAULT 'USD',
		minimum_amount DECIMAL(14,4) NULL,
		maximum_amount DECIMAL(14,4) NULL,
		included_in_base_rate BOOLEAN NOT NULL DEFAULT FALSE,
		display_order INT NOT NULL DEFAULT 0,
		notes TEXT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_rci_org_rate (org_id, rate_id),
		INDEX idx_rci_org_rate_order (org_id, rate_id, display_order),
		INDEX idx_rci_org_category (org_id, charge_category),
		INDEX idx_rci_created_at (created_at),
		CONSTRAINT fk_rci_rate FOREIGN KEY (rate_id) REFERENCES rates(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: rate_charge_items table: %v", err)
	}

	// Migration 060: Rate contracts and versioning
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS rate_contracts (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		contract_reference VARCHAR(100) NOT NULL,
		carrier_name VARCHAR(100) NOT NULL,
		carrier_code VARCHAR(20) NULL,
		contract_name VARCHAR(255) NOT NULL,
		contract_type VARCHAR(50) NOT NULL DEFAULT 'ANNUAL_SERVICE',
		transport_mode VARCHAR(50) NULL DEFAULT 'Ocean FCL',
		currency VARCHAR(10) NULL DEFAULT 'USD',
		effective_date DATE NOT NULL,
		expiry_date DATE NOT NULL,
		status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',
		renewal_status VARCHAR(30) NOT NULL DEFAULT 'NOT_STARTED',
		renewal_owner VARCHAR(255) NULL,
		notes TEXT NULL,
		created_by VARCHAR(255) NULL,
		updated_by VARCHAR(255) NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_rc_org_carrier (org_id, carrier_name),
		INDEX idx_rc_org_status (org_id, status),
		INDEX idx_rc_org_expiry (org_id, expiry_date),
		INDEX idx_rc_org_renewal (org_id, renewal_status),
		INDEX idx_rc_org_ref (org_id, contract_reference)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: rate_contracts table: %v", err)
	}

	// Columns on rates for versioning & contract lineage
	cols := []string{
		`ALTER TABLE rates ADD COLUMN contract_id BIGINT NULL`,
		`ALTER TABLE rates ADD COLUMN version_number INT NOT NULL DEFAULT 1`,
		`ALTER TABLE rates ADD COLUMN version_status VARCHAR(30) NOT NULL DEFAULT 'CURRENT'`,
		`ALTER TABLE rates ADD COLUMN supersedes_rate_id BIGINT NULL`,
		`ALTER TABLE rates ADD COLUMN superseded_by_rate_id BIGINT NULL`,
		`ALTER TABLE rates ADD COLUMN version_created_at DATETIME NULL`,
	}
	for _, colSQL := range cols {
		_, _ = db.Exec(colSQL)
	}

	indexes := []string{
		`ALTER TABLE rates ADD INDEX idx_rates_org_contract (org_id, contract_id)`,
		`ALTER TABLE rates ADD INDEX idx_rates_org_contract_version (org_id, contract_id, version_number)`,
		`ALTER TABLE rates ADD INDEX idx_rates_org_version_status (org_id, version_status)`,
		`ALTER TABLE rates ADD INDEX idx_rates_org_supersedes (org_id, supersedes_rate_id)`,
	}
	for _, idxSQL := range indexes {
		_, _ = db.Exec(idxSQL)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS rate_version_history (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		rate_id BIGINT NOT NULL,
		version_number INT NOT NULL DEFAULT 1,
		action VARCHAR(50) NOT NULL DEFAULT 'RATE_VERSION_CREATED',
		previous_rate_id BIGINT NULL,
		new_rate_id BIGINT NULL,
		description TEXT NOT NULL,
		performed_by VARCHAR(255) NULL,
		metadata JSON NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_rvh_org_rate (org_id, rate_id),
		INDEX idx_rvh_org_action (org_id, action),
		INDEX idx_rvh_org_created (org_id, created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: rate_version_history table: %v", err)
	}

	// 061_spot_rate_requests_and_responses.sql
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS spot_rate_requests (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		request_reference VARCHAR(64) NOT NULL,
		customer_id BIGINT NULL,
		customer_name VARCHAR(255) NULL,
		origin_port VARCHAR(128) NOT NULL,
		origin_code VARCHAR(32) NULL,
		destination_port VARCHAR(128) NOT NULL,
		destination_code VARCHAR(32) NULL,
		transport_mode VARCHAR(64) NOT NULL DEFAULT 'Ocean FCL',
		service_type VARCHAR(64) NOT NULL DEFAULT 'FCL',
		equipment_type VARCHAR(64) NULL DEFAULT '40GP',
		commodity VARCHAR(255) NULL,
		cargo_weight DECIMAL(12, 2) NULL,
		cargo_volume DECIMAL(12, 2) NULL,
		container_quantity INT NOT NULL DEFAULT 1,
		ready_date DATE NOT NULL,
		target_currency VARCHAR(8) NOT NULL DEFAULT 'USD',
		required_by_date DATE NOT NULL,
		status VARCHAR(32) NOT NULL DEFAULT 'DRAFT',
		notes TEXT NULL,
		created_by VARCHAR(128) NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_srr_org_status (org_id, status),
		INDEX idx_srr_org_ref (org_id, request_reference),
		INDEX idx_srr_org_lane (org_id, origin_port, destination_port),
		INDEX idx_srr_org_created (org_id, created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: spot_rate_requests table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS spot_rate_responses (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		spot_rate_request_id BIGINT NOT NULL,
		carrier_name VARCHAR(255) NOT NULL,
		carrier_code VARCHAR(32) NULL,
		supplier_name VARCHAR(255) NULL,
		rate_id BIGINT NULL,
		currency VARCHAR(8) NOT NULL DEFAULT 'USD',
		base_amount DECIMAL(14, 2) NOT NULL DEFAULT 0.00,
		total_amount DECIMAL(14, 2) NOT NULL DEFAULT 0.00,
		transit_days INT NULL,
		free_days_origin INT NOT NULL DEFAULT 0,
		free_days_destination INT NOT NULL DEFAULT 0,
		valid_from DATE NOT NULL,
		valid_until DATE NOT NULL,
		routing_notes TEXT NULL,
		response_notes TEXT NULL,
		status VARCHAR(32) NOT NULL DEFAULT 'RECEIVED',
		is_preferred BOOLEAN NOT NULL DEFAULT FALSE,
		responded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_by VARCHAR(128) NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_sresp_org_req (org_id, spot_rate_request_id),
		INDEX idx_sresp_org_carrier (org_id, carrier_name),
		INDEX idx_sresp_org_pref (org_id, is_preferred),
		INDEX idx_sresp_org_valid (org_id, valid_until)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: spot_rate_responses table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS spot_rate_response_charges (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		spot_rate_response_id BIGINT NOT NULL,
		charge_category VARCHAR(64) NOT NULL DEFAULT 'FREIGHT',
		charge_name VARCHAR(128) NOT NULL,
		calculation_basis VARCHAR(32) NOT NULL DEFAULT 'FLAT',
		quantity DECIMAL(12, 2) NOT NULL DEFAULT 1.00,
		unit_price DECIMAL(14, 2) NOT NULL DEFAULT 0.00,
		currency VARCHAR(8) NOT NULL DEFAULT 'USD',
		display_order INT NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_srrc_org_resp (org_id, spot_rate_response_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: spot_rate_response_charges table: %v", err)
	}

	// Task 19.5 Migration 062: Rate-to-Quotation Integration
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS quotation_rate_selections (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		quotation_id BIGINT NOT NULL,
		rate_id BIGINT NULL,
		spot_rate_request_id BIGINT NULL,
		spot_rate_response_id BIGINT NULL,
		rate_source_type VARCHAR(32) NOT NULL DEFAULT 'MANAGED_RATE',
		selected_by VARCHAR(128) NULL,
		selected_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		notes TEXT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_qrs_org_quotation (org_id, quotation_id),
		INDEX idx_qrs_org_rate (org_id, rate_id),
		INDEX idx_qrs_org_spot_resp (org_id, spot_rate_response_id),
		INDEX idx_qrs_org_active (org_id, quotation_id, is_active)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: quotation_rate_selections table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS quotation_rate_snapshots (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		quotation_id BIGINT NOT NULL,
		quotation_rate_selection_id BIGINT NOT NULL,
		source_rate_id BIGINT NULL,
		source_rate_version INT NULL,
		source_contract_id BIGINT NULL,
		source_spot_rate_request_id BIGINT NULL,
		source_spot_rate_response_id BIGINT NULL,
		carrier_name VARCHAR(128) NOT NULL,
		carrier_reference VARCHAR(128) NULL,
		transport_mode VARCHAR(64) NOT NULL,
		service_type VARCHAR(64) NULL,
		equipment_type VARCHAR(64) NULL,
		origin VARCHAR(128) NOT NULL,
		destination VARCHAR(128) NOT NULL,
		currency VARCHAR(8) NOT NULL DEFAULT 'USD',
		base_rate DECIMAL(14, 2) NOT NULL DEFAULT 0.00,
		additional_charges DECIMAL(14, 2) NOT NULL DEFAULT 0.00,
		commercial_total DECIMAL(14, 2) NOT NULL DEFAULT 0.00,
		pricing_snapshot JSON NOT NULL,
		valid_from DATE NULL,
		valid_until DATE NULL,
		snapshot_created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_by VARCHAR(128) NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_qrsnap_org_quotation (org_id, quotation_id),
		INDEX idx_qrsnap_org_selection (org_id, quotation_rate_selection_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: quotation_rate_snapshots table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS quotation_rate_selection_history (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		quotation_id BIGINT NOT NULL,
		event_type VARCHAR(64) NOT NULL,
		previous_selection_id BIGINT NULL,
		new_selection_id BIGINT NULL,
		description TEXT NOT NULL,
		metadata JSON NULL,
		performed_by VARCHAR(128) NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_qrsh_org_quotation (org_id, quotation_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: quotation_rate_selection_history table: %v", err)
	}

	// ── Migration 063: Rate Lifecycle Intelligence & Commercial Risk ─────────────
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS rate_lifecycle_events (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		rate_id BIGINT NULL,
		contract_id BIGINT NULL,
		event_type VARCHAR(64) NOT NULL,
		previous_status VARCHAR(32) NULL,
		current_status VARCHAR(32) NOT NULL,
		description TEXT NOT NULL,
		metadata JSON NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_rle_org_rate (org_id, rate_id),
		INDEX idx_rle_org_contract (org_id, contract_id),
		INDEX idx_rle_org_event (org_id, event_type),
		INDEX idx_rle_org_created (org_id, created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: rate_lifecycle_events table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS quotation_rate_risk_events (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		quotation_id BIGINT NOT NULL,
		quotation_rate_snapshot_id BIGINT NULL,
		source_rate_id BIGINT NULL,
		source_contract_id BIGINT NULL,
		source_spot_rate_response_id BIGINT NULL,
		risk_type VARCHAR(64) NOT NULL,
		severity VARCHAR(16) NOT NULL DEFAULT 'WARNING',
		headline VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		recommended_action TEXT NULL,
		is_resolved BOOLEAN DEFAULT FALSE,
		resolved_by VARCHAR(128) NULL,
		resolved_at TIMESTAMP NULL,
		metadata JSON NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_qrre_org_quote (org_id, quotation_id),
		INDEX idx_qrre_org_resolved (org_id, is_resolved),
		INDEX idx_qrre_org_severity (org_id, severity),
		INDEX idx_qrre_org_rate (org_id, source_rate_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: quotation_rate_risk_events table: %v", err)
	}

	// ── Migration 069: Customer Foundation & Directory ──────────────────────────
	custCols := []string{
		`ALTER TABLE customers ADD COLUMN customer_code VARCHAR(50) NULL AFTER org_id`,
		`ALTER TABLE customers ADD COLUMN trading_name VARCHAR(255) NULL AFTER name`,
		`ALTER TABLE customers ADD COLUMN customer_type VARCHAR(50) NOT NULL DEFAULT 'SHIPPER' AFTER trading_name`,
		`ALTER TABLE customers ADD COLUMN tax_id VARCHAR(100) NULL AFTER industry`,
		`ALTER TABLE customers ADD COLUMN pan_number VARCHAR(50) NULL AFTER tax_id`,
		`ALTER TABLE customers ADD COLUMN eori_number VARCHAR(50) NULL AFTER pan_number`,
		`ALTER TABLE customers ADD COLUMN currency VARCHAR(10) NOT NULL DEFAULT 'USD' AFTER eori_number`,
		`ALTER TABLE customers ADD COLUMN payment_terms VARCHAR(50) NOT NULL DEFAULT 'NET30' AFTER currency`,
		`ALTER TABLE customers ADD COLUMN credit_limit DECIMAL(14,2) NOT NULL DEFAULT 0.00 AFTER payment_terms`,
		`ALTER TABLE customers ADD COLUMN health_score INT NOT NULL DEFAULT 80 AFTER credit_limit`,
		`ALTER TABLE customers ADD COLUMN account_owner_id BIGINT NULL AFTER health_score`,
		`ALTER TABLE customers ADD COLUMN website VARCHAR(255) NULL AFTER domain`,
		`ALTER TABLE customers ADD COLUMN country VARCHAR(100) NULL AFTER website`,
		`ALTER TABLE customers ADD COLUMN city VARCHAR(100) NULL AFTER country`,
		`ALTER TABLE customers ADD COLUMN notes TEXT NULL AFTER contact_phone`,
		`ALTER TABLE customers ADD COLUMN archived_at TIMESTAMP NULL AFTER updated_at`,
	}
	for _, sqlStr := range custCols {
		_, _ = db.Exec(sqlStr)
	}

	custIndexes := []string{
		`ALTER TABLE customers ADD UNIQUE INDEX uq_customers_org_code (org_id, customer_code)`,
		`ALTER TABLE customers ADD INDEX idx_customers_org_status (org_id, status)`,
		`ALTER TABLE customers ADD INDEX idx_customers_org_type (org_id, customer_type)`,
		`ALTER TABLE customers ADD INDEX idx_customers_org_owner (org_id, account_owner_id)`,
		`ALTER TABLE customers ADD INDEX idx_customers_org_name (org_id, name)`,
	}
	for _, sqlStr := range custIndexes {
		_, _ = db.Exec(sqlStr)
	}

	contactCols := []string{
		`ALTER TABLE contacts ADD COLUMN department VARCHAR(100) NULL AFTER job_title`,
		`ALTER TABLE contacts ADD COLUMN is_primary BOOLEAN NOT NULL DEFAULT FALSE AFTER department`,
		`ALTER TABLE contacts ADD COLUMN notes TEXT NULL AFTER is_primary`,
	}
	for _, sqlStr := range contactCols {
		_, _ = db.Exec(sqlStr)
	}
	_, _ = db.Exec(`ALTER TABLE contacts ADD INDEX idx_contacts_org_customer (org_id, customer_id)`)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS customer_addresses (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		customer_id BIGINT NOT NULL,
		address_type VARCHAR(50) NOT NULL DEFAULT 'BILLING',
		label VARCHAR(150) NULL,
		address_line_1 VARCHAR(255) NOT NULL,
		address_line_2 VARCHAR(255) NULL,
		city VARCHAR(100) NOT NULL,
		state VARCHAR(100) NULL,
		postal_code VARCHAR(30) NULL,
		country_code VARCHAR(10) NOT NULL DEFAULT 'US',
		country VARCHAR(100) NOT NULL DEFAULT 'United States',
		is_primary_billing BOOLEAN NOT NULL DEFAULT FALSE,
		is_primary_shipping BOOLEAN NOT NULL DEFAULT FALSE,
		contact_name VARCHAR(150) NULL,
		contact_phone VARCHAR(50) NULL,
		contact_email VARCHAR(150) NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		CONSTRAINT fk_cust_addr_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
		CONSTRAINT fk_cust_addr_cust FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: customer_addresses table: %v", err)
	}
	_, _ = db.Exec(`ALTER TABLE customer_addresses ADD INDEX idx_cust_addr_lookup (org_id, customer_id, address_type)`)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS customer_lead_links (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		customer_id BIGINT NOT NULL,
		lead_id BIGINT NOT NULL,
		converted_by_user_id BIGINT NULL,
		conversion_notes TEXT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT fk_cll_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
		CONSTRAINT fk_cll_cust FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
		CONSTRAINT fk_cll_lead FOREIGN KEY (lead_id) REFERENCES leads(id) ON DELETE CASCADE,
		CONSTRAINT uq_cll_lead UNIQUE (org_id, lead_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: customer_lead_links table: %v", err)
	}
	_, _ = db.Exec(`ALTER TABLE customer_lead_links ADD INDEX idx_cll_cust (org_id, customer_id)`)

	_, _ = db.Exec(`UPDATE customers SET customer_code = CONCAT('CUST-', YEAR(NOW()), '-', LPAD(id, 5, '0')) WHERE customer_code IS NULL OR customer_code = ''`)

	// Migration 070: Customer Financial & Relationship Management
	log.Println("Applying Migration 070: Customer Financial & Relationship Management...")
	_, _ = db.Exec(`ALTER TABLE customers ADD COLUMN credit_status VARCHAR(50) NOT NULL DEFAULT 'GOOD_STANDING'`)
	_, _ = db.Exec(`ALTER TABLE customers ADD COLUMN secondary_owner_id BIGINT NULL`)
	_, _ = db.Exec(`ALTER TABLE customers ADD COLUMN commercial_notes TEXT NULL`)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS customer_ownership_history (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		customer_id BIGINT NOT NULL,
		previous_owner_id BIGINT NULL,
		new_owner_id BIGINT NOT NULL,
		ownership_type VARCHAR(50) NOT NULL DEFAULT 'PRIMARY',
		changed_by_user_id BIGINT NULL,
		change_reason VARCHAR(255) NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_cust_owner_hist_org_cust (org_id, customer_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: customer_ownership_history table: %v", err)
	}

	_, _ = db.Exec(`ALTER TABLE contacts ADD COLUMN contact_role VARCHAR(50) NOT NULL DEFAULT 'COMMERCIAL'`)
	_, _ = db.Exec(`ALTER TABLE contacts ADD COLUMN mobile VARCHAR(50) NULL`)

	// Migration 071: Customer Intelligence, Health & Risk Management
	log.Println("Applying Migration 071: Customer Intelligence, Health & Risk Management...")
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS customer_health_evaluations (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		customer_id BIGINT NOT NULL,
		health_status VARCHAR(50) NOT NULL DEFAULT 'INSUFFICIENT_DATA',
		health_score INT NOT NULL DEFAULT 50,
		contributing_factors_json TEXT NULL,
		evaluated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_che_org_cust (org_id, customer_id),
		INDEX idx_che_status (org_id, health_status)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: customer_health_evaluations table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS customer_risk_events (
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
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: customer_risk_events table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS customer_opportunity_events (
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
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: customer_opportunity_events table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS customer_intelligence_events (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		customer_id BIGINT NOT NULL,
		event_type VARCHAR(100) NOT NULL,
		title VARCHAR(255) NOT NULL,
		description TEXT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_cie_org_cust (org_id, customer_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: customer_intelligence_events table: %v", err)
	}

	log.Println("Applying Migration 072: Outreach Audience & Attribution...")
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS outreach_campaign_leads (
		campaign_id BIGINT NOT NULL,
		lead_id BIGINT NOT NULL,
		added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (campaign_id, lead_id),
		CONSTRAINT fk_ocl_campaign FOREIGN KEY (campaign_id) REFERENCES outreach_campaigns(id) ON DELETE CASCADE,
		CONSTRAINT fk_ocl_lead FOREIGN KEY (lead_id) REFERENCES leads(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: outreach_campaign_leads table: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE outreach_sequences ADD COLUMN name VARCHAR(255) NULL`)
	if err != nil {
		log.Printf("Migration info: outreach_sequences name: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE outreach_sequences ADD COLUMN subject VARCHAR(500) NULL`)
	if err != nil {
		log.Printf("Migration info: outreach_sequences subject: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE outreach_sequences ADD COLUMN body TEXT NULL`)
	if err != nil {
		log.Printf("Migration info: outreach_sequences body: %v", err)
	}

	_, err = db.Exec(`ALTER TABLE leads ADD COLUMN campaign_id BIGINT NULL`)
	if err != nil {
		log.Printf("Migration info: leads campaign_id: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE leads ADD COLUMN converted_from_outreach_at TIMESTAMP NULL`)
	if err != nil {
		log.Printf("Migration info: leads converted_from_outreach_at: %v", err)
	}
	_, err = db.Exec(`ALTER TABLE leads ADD CONSTRAINT fk_leads_campaign FOREIGN KEY (campaign_id) REFERENCES outreach_campaigns(id) ON DELETE SET NULL`)
	if err != nil {
		log.Printf("Migration info: constraint fk_leads_campaign: %v", err)
	}

	log.Println("Applying Migration 073: Outreach Activities...")
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS outreach_activities (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		lead_id BIGINT NULL,
		customer_id BIGINT NULL,
		activity_type VARCHAR(50) NOT NULL,
		subject VARCHAR(255) NOT NULL,
		description TEXT NULL,
		status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
		priority VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
		scheduled_at TIMESTAMP NULL,
		completed_at TIMESTAMP NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		created_by BIGINT NULL,
		CONSTRAINT fk_oa_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
		CONSTRAINT fk_oa_lead FOREIGN KEY (lead_id) REFERENCES leads(id) ON DELETE CASCADE,
		CONSTRAINT fk_oa_customer FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE SET NULL,
		CONSTRAINT fk_oa_user FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: outreach_activities table: %v", err)
	}

	// Seed activity records
	_, _ = db.Exec(`INSERT INTO outreach_activities (org_id, lead_id, activity_type, subject, description, status, priority, scheduled_at)
		SELECT l.org_id, l.id, 'CALL', 'Follow-up Call', 'Follow-up Call to John Smith', 'PENDING', 'HIGH', DATE_ADD(NOW(), INTERVAL 2 HOUR)
		FROM leads l WHERE l.company_name LIKE '%Oceanic%' OR l.id = 1 LIMIT 1`)

	_, _ = db.Exec(`INSERT INTO outreach_activities (org_id, lead_id, activity_type, subject, description, status, priority, scheduled_at)
		SELECT l.org_id, l.id, 'EMAIL', 'Email Follow-up', 'Send follow-up details to Priya', 'IN_PROGRESS', 'MEDIUM', DATE_ADD(NOW(), INTERVAL 5 HOUR)
		FROM leads l WHERE l.company_name LIKE '%Transworld%' OR l.id = 2 LIMIT 1`)

	_, _ = db.Exec(`INSERT INTO outreach_activities (org_id, lead_id, activity_type, subject, description, status, priority, scheduled_at)
		SELECT l.org_id, l.id, 'MEETING', 'Proposal Discussion', 'Discuss Q3 contract options', 'PENDING', 'HIGH', DATE_ADD(NOW(), INTERVAL 1 DAY)
		FROM leads l WHERE l.company_name LIKE '%Global%' OR l.id = 3 LIMIT 1`)

	_, _ = db.Exec(`INSERT INTO outreach_activities (org_id, lead_id, activity_type, subject, description, status, priority, scheduled_at)
		SELECT l.org_id, l.id, 'EMAIL', 'Follow-up Email', 'Check in on inactive account', 'PENDING', 'LOW', DATE_ADD(DATE_ADD(NOW(), INTERVAL 1 DAY), INTERVAL 6 HOUR)
		FROM leads l WHERE l.company_name LIKE '%Bluewave%' OR l.id = 4 LIMIT 1`)

	_, _ = db.Exec(`INSERT INTO outreach_activities (org_id, lead_id, activity_type, subject, description, status, priority, scheduled_at)
		SELECT l.org_id, l.id, 'CALL', 'Contract Renewal Call', 'Discuss upcoming renewals', 'OVERDUE', 'HIGH', DATE_SUB(NOW(), INTERVAL 1 DAY)
		FROM leads l WHERE l.company_name LIKE '%Speedex%' OR l.id = 5 LIMIT 1`)

	log.Println("Applying Migration 076: Approvals Module Foundation...")
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS approval_requests (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		request_code VARCHAR(50) NOT NULL,
		title VARCHAR(255) NOT NULL,
		category VARCHAR(50) NOT NULL DEFAULT 'DOCUMENTS',
		type VARCHAR(100) NOT NULL DEFAULT 'Document Approval',
		status VARCHAR(50) NOT NULL DEFAULT 'Pending',
		priority VARCHAR(20) NOT NULL DEFAULT 'MEDIUM',
		related_entity_type VARCHAR(50) NULL,
		related_entity_id BIGINT NULL,
		related_ref VARCHAR(100) NULL,
		customer_name VARCHAR(255) NULL,
		customer_id BIGINT NULL,
		shipment_id BIGINT NULL,
		document_id BIGINT NULL,
		booking_id BIGINT NULL,
		requested_by_id BIGINT NULL,
		requested_by_name VARCHAR(100) NOT NULL,
		department VARCHAR(100) NULL DEFAULT 'Operations',
		avatar VARCHAR(10) NULL DEFAULT 'AS',
		due_date DATETIME NULL,
		due_text VARCHAR(50) NULL DEFAULT '7 days left',
		assigned_to VARCHAR(100) NULL,
		approved_by VARCHAR(100) NULL,
		approved_at DATETIME NULL,
		rejected_by VARCHAR(100) NULL,
		rejected_at DATETIME NULL,
		rejection_reason TEXT NULL,
		comments TEXT NULL,
		description TEXT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		CONSTRAINT fk_ar_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
		CONSTRAINT fk_ar_customer FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE SET NULL,
		CONSTRAINT fk_ar_shipment FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE SET NULL,
		CONSTRAINT fk_ar_document FOREIGN KEY (document_id) REFERENCES shipment_documents(id) ON DELETE SET NULL
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: approval_requests table: %v", err)
	}

	// Seed initial sample approval records if table is empty
	var approvalCount int
	_ = db.Get(&approvalCount, "SELECT COUNT(*) FROM approval_requests WHERE org_id = 1")
	if approvalCount == 0 {
		log.Println("Seeding initial approval records into approval_requests...")
		_, _ = db.Exec(`INSERT INTO approval_requests (
			org_id, request_code, title, category, type, status, priority, related_ref, customer_name, customer_id, requested_by_name, department, avatar, due_date, due_text, description
		) VALUES 
		(1, 'DOC-APP-1247', 'House Bill of Lading - SHP-250812', 'DOCUMENTS', 'Document Approval', 'Overdue', 'HIGH', 'Shipment: SHP-250812', 'ABC Exports Pvt Ltd', 1, 'Arjun Singh', 'Operations', 'AS', '2026-08-12 10:00:00', 'Overdue', 'House Bill of Lading requires manager sign-off for container release at Mumbai port.'),
		(1, 'DOC-APP-1246', 'Commercial Invoice - INV-10293', 'DOCUMENTS', 'Document Approval', 'Pending', 'MEDIUM', 'Shipment: SHP-250799', 'XYZ Logistics Inc.', 1, 'Neha Kapoor', 'Documentation', 'NK', '2026-08-16 14:00:00', '2 days left', 'Commercial invoice discrepancy check required before customs clearance filing.'),
		(1, 'COM-APP-0891', 'Credit Limit Increase Request', 'COMMERCIAL', 'Commercial Approval', 'Pending', 'HIGH', 'Customer: CUST-00734', 'Global Auto Movers', 1, 'Rohit Mehta', 'Sales', 'RM', '2026-08-18 09:00:00', '4 days left', 'Request to increase credit limit from $50,000 to $100,000 for peak season volume.'),
		(1, 'OPS-APP-0562', 'Rate Exception - SHP-250788', 'OPERATIONS', 'Operations Approval', 'Pending', 'MEDIUM', 'Shipment: SHP-250788', 'Oceanic Freight Co.', 1, 'Pooja Shah', 'Operations', 'PS', '2026-08-20 11:00:00', '6 days left', 'Demurrage waiver request due to port congestion delay at Nhava Sheva.'),
		(1, 'FIN-APP-0321', 'Invoice Discount Approval', 'FINANCE', 'Finance Approval', 'Pending', 'LOW', 'Invoice: INV-10277', 'Blue Sky Imports', 1, 'Vikram Kumar', 'Finance', 'VK', '2026-08-21 16:00:00', '7 days left', '5% early payment discount authorization for key customer invoice INV-10277.'),
		(1, 'DOC-APP-1245', 'Packing List - SHP-250812', 'DOCUMENTS', 'Document Approval', 'Pending', 'LOW', 'Shipment: SHP-250812', 'ABC Exports Pvt Ltd', 1, 'Arjun Singh', 'Operations', 'AS', '2026-08-21 17:00:00', '8 days left', 'Cargo weight verification against packing list data.'),
		(1, 'DOC-APP-1240', 'Master Bill of Lading - SHP-250750', 'DOCUMENTS', 'Document Approval', 'Approved', 'HIGH', 'Shipment: SHP-250750', 'Indo-US Traders', 1, 'Arjun Singh', 'Operations', 'AS', '2026-08-05 08:00:00', 'Approved', 'MBL carrier release verified and approved.'),
		(1, 'COM-APP-0880', 'Special Spot Quotation Pricing', 'COMMERCIAL', 'Commercial Approval', 'Rejected', 'HIGH', 'Quotation: QT-9912', 'Pacific Rim Corp', 1, 'Rohit Mehta', 'Sales', 'RM', '2026-08-04 13:00:00', 'Rejected', 'Requested rate below minimum margin threshold.')`)
	}

	log.Println("Applying Migration 077: Invoices Module Foundation...")
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS customer_invoices (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		invoice_number VARCHAR(100) NOT NULL,
		customer_id BIGINT NOT NULL,
		customer_name VARCHAR(255) NOT NULL,
		customer_country VARCHAR(100) NOT NULL DEFAULT '',
		shipment_id BIGINT NULL,
		shipment_number VARCHAR(100) NOT NULL DEFAULT '',
		booking_id BIGINT NULL,
		booking_number VARCHAR(100) NOT NULL DEFAULT '',
		quotation_id BIGINT NULL,
		quote_number VARCHAR(100) NOT NULL DEFAULT '',
		route VARCHAR(255) NOT NULL DEFAULT '',
		origin VARCHAR(100) NOT NULL DEFAULT '',
		destination VARCHAR(100) NOT NULL DEFAULT '',
		invoice_date DATE NOT NULL,
		due_date DATE NOT NULL,
		days_left VARCHAR(50) NOT NULL DEFAULT '',
		currency VARCHAR(10) NOT NULL DEFAULT 'USD',
		subtotal DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
		tax_amount DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
		discount_amount DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
		total_amount DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
		paid_amount DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
		balance_due DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
		status VARCHAR(50) NOT NULL DEFAULT 'Draft',
		type VARCHAR(50) NOT NULL DEFAULT 'CUSTOMER_AR',
		bookmarked BOOLEAN NOT NULL DEFAULT FALSE,
		is_my_invoice BOOLEAN NOT NULL DEFAULT TRUE,
		creator_name VARCHAR(100) NOT NULL DEFAULT '',
		created_by_id BIGINT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		CONSTRAINT uq_cust_inv_num_org UNIQUE (org_id, invoice_number),
		CONSTRAINT fk_cust_inv_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
		CONSTRAINT fk_cust_inv_customer FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE RESTRICT
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: customer_invoices table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS customer_invoice_items (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		invoice_id BIGINT NOT NULL,
		description VARCHAR(255) NOT NULL,
		service_category VARCHAR(100) NOT NULL DEFAULT '',
		quantity DECIMAL(12, 4) NOT NULL DEFAULT 1.0000,
		unit_price DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
		total_amount DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
		display_order INT NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT fk_cust_inv_item_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
		CONSTRAINT fk_cust_inv_item_invoice FOREIGN KEY (invoice_id) REFERENCES customer_invoices(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: customer_invoice_items table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS customer_invoice_payments (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		invoice_id BIGINT NOT NULL,
		payment_ref VARCHAR(100) NOT NULL,
		amount DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
		payment_method VARCHAR(100) NOT NULL DEFAULT 'Wire Transfer',
		status VARCHAR(50) NOT NULL DEFAULT 'Completed',
		payment_date DATE NOT NULL,
		notes TEXT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT fk_cust_inv_pay_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
		CONSTRAINT fk_cust_inv_pay_invoice FOREIGN KEY (invoice_id) REFERENCES customer_invoices(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: customer_invoice_payments table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS customer_invoice_documents (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		invoice_id BIGINT NOT NULL,
		document_name VARCHAR(255) NOT NULL,
		file_size VARCHAR(50) NOT NULL DEFAULT '250 KB',
		file_type VARCHAR(100) NOT NULL DEFAULT 'application/pdf',
		s3_key TEXT NULL,
		uploaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT fk_cust_inv_doc_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
		CONSTRAINT fk_cust_inv_doc_invoice FOREIGN KEY (invoice_id) REFERENCES customer_invoices(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: customer_invoice_documents table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS customer_invoice_history (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		org_id BIGINT NOT NULL,
		invoice_id BIGINT NOT NULL,
		title VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		user_name VARCHAR(100) NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT fk_cust_inv_hist_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
		CONSTRAINT fk_cust_inv_hist_invoice FOREIGN KEY (invoice_id) REFERENCES customer_invoices(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		log.Printf("Migration info: customer_invoice_history table: %v", err)
	}

	// Seed initial customer_invoices if empty for org 1
	var invoiceCount int
	_ = db.Get(&invoiceCount, "SELECT COUNT(*) FROM customer_invoices WHERE org_id = 1")
	if invoiceCount == 0 {
		log.Println("Seeding initial invoice records into customer_invoices table...")
		// Ensure a default customer exists for foreign key constraint
		var defaultCustID int64
		_ = db.Get(&defaultCustID, "SELECT id FROM customers WHERE org_id = 1 LIMIT 1")
		if defaultCustID == 0 {
			res, _ := db.Exec(`INSERT INTO customers (org_id, name, customer_code, email, created_at, updated_at) VALUES (1, 'Global Traders Inc.', 'CUST-001', 'billing@globaltraders.com', NOW(), NOW())`)
			if res != nil {
				defaultCustID, _ = res.LastInsertId()
			}
			if defaultCustID == 0 {
				defaultCustID = 1
			}
		}

		// Insert 8 realistic invoice records matching Task 1 UI
		invoicesSeed := []struct {
			num, cust, country, ship, route, invDate, dueDate, daysLeft, status, creator string
			total, subtotal, tax, disc, paid, bal float64
			bm, my bool
		}{
			{"INV-2026-0456", "Global Traders Inc.", "USA", "SH-2026-00124", "Shanghai ➔ Los Angeles", "2026-08-15", "2026-08-30", "15 days left", "Issued", "By Varun Sharma", 24650.00, 22410.00, 0.00, 0.00, 0.00, 24650.00, true, true},
			{"INV-2026-0455", "Oceanic Imports Pvt. Ltd.", "India", "SH-2026-00123", "Nhava Sheva ➔ New York", "2026-08-14", "2026-08-29", "14 days left", "Partially Paid", "By Priya Nair", 18940.00, 17200.00, 1740.00, 0.00, 10000.00, 8940.00, false, false},
			{"INV-2026-0454", "Bright Star Ltd.", "UK", "SH-2026-00122", "Dubai ➔ Felixstowe", "2026-08-12", "2026-08-27", "12 days left", "Overdue", "By Varun Sharma", 32120.00, 30000.00, 2120.00, 0.00, 0.00, 32120.00, true, true},
			{"INV-2026-0453", "Techtronics GmbH", "Germany", "SH-2026-00121", "Hamburg ➔ Mumbai", "2026-08-11", "2026-08-26", "11 days left", "Paid", "By Priya Nair", 15780.00, 15780.00, 0.00, 0.00, 15780.00, 0.00, false, false},
			{"INV-2026-0452", "Southern Retail LLC", "USA", "SH-2026-00120", "Los Angeles ➔ Chicago", "2026-08-10", "2026-08-25", "10 days left", "Paid", "By Rohan Mehta", 9650.00, 9650.00, 0.00, 0.00, 9650.00, 0.00, false, true},
			{"INV-2026-0451", "Alpha Logistics", "Singapore", "SH-2026-00119", "Singapore ➔ Sydney", "2026-08-09", "2026-08-24", "9 days left", "Pending Approval", "By Varun Sharma", 7850.00, 7850.00, 0.00, 0.00, 0.00, 7850.00, false, true},
			{"INV-2026-0450", "East Coast Traders", "Canada", "SH-2026-00118", "Vancouver ➔ Seattle", "2026-08-08", "2026-08-23", "8 days left", "Draft", "By Priya Nair", 11320.00, 11320.00, 0.00, 0.00, 0.00, 11320.00, false, false},
			{"INV-2026-0449", "Sunrise Exports", "India", "SH-2026-00117", "Chennai ➔ Los Angeles", "2026-08-06", "2026-08-21", "6 days left", "Issued", "By Rohan Mehta", 26450.00, 26450.00, 0.00, 0.00, 0.00, 26450.00, false, false},
		}

		for _, invSeed := range invoicesSeed {
			res, err := db.Exec(`INSERT INTO customer_invoices (
				org_id, invoice_number, customer_id, customer_name, customer_country, shipment_number, route, invoice_date, due_date, days_left, currency, subtotal, tax_amount, discount_amount, total_amount, paid_amount, balance_due, status, type, bookmarked, is_my_invoice, creator_name, created_at, updated_at
			) VALUES (
				1, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'USD', ?, ?, ?, ?, ?, ?, ?, 'CUSTOMER_AR', ?, ?, ?, NOW(), NOW()
			)`, invSeed.num, defaultCustID, invSeed.cust, invSeed.country, invSeed.ship, invSeed.route, invSeed.invDate, invSeed.dueDate, invSeed.daysLeft, invSeed.subtotal, invSeed.tax, invSeed.disc, invSeed.total, invSeed.paid, invSeed.bal, invSeed.status, invSeed.bm, invSeed.my, invSeed.creator)

			if err == nil && res != nil {
				invID, _ := res.LastInsertId()
				if invID > 0 {
					// Seed line items for invID
					_, _ = db.Exec(`INSERT INTO customer_invoice_items (org_id, invoice_id, description, service_category, quantity, unit_price, total_amount, display_order) VALUES
						(1, ?, 'Ocean Freight (40ft FCL High Cube)', 'Freight', 1, ?, ?, 1),
						(1, ?, 'Documentation Fee & B/L Issuance', 'Documentation', 1, 350.00, 350.00, 2)`,
						invID, invSeed.subtotal-350.00, invSeed.subtotal-350.00, invID)

					// Seed documents
					_, _ = db.Exec(`INSERT INTO customer_invoice_documents (org_id, invoice_id, document_name, file_size, file_type) VALUES
						(1, ?, 'Commercial Invoice.pdf', '245 KB', 'application/pdf'),
						(1, ?, 'Bill of Lading.pdf', '320 KB', 'application/pdf'),
						(1, ?, 'Packing List.pdf', '180 KB', 'application/pdf')`, invID, invID, invID)

					// Seed history
					_, _ = db.Exec(`INSERT INTO customer_invoice_history (org_id, invoice_id, title, description, user_name) VALUES
						(1, ?, 'Invoice Created', 'Generated in system', 'System')`, invID)
				}
			}
		}
	}

	// ── Migration 081: Universal Audit Logs Foundation ─────────────────────
	_, _ = db.Exec(`ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS actor_type VARCHAR(50) NOT NULL DEFAULT 'USER'`)
	_, _ = db.Exec(`ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS actor_name VARCHAR(255) NULL`)
	_, _ = db.Exec(`ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS actor_role VARCHAR(100) NULL`)
	_, _ = db.Exec(`ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS module VARCHAR(100) NOT NULL DEFAULT 'GENERAL'`)
	_, _ = db.Exec(`ALTER TABLE audit_logs MODIFY COLUMN resource_id VARCHAR(255) NOT NULL`)
	_, _ = db.Exec(`ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS resource_name VARCHAR(255) NULL`)
	_, _ = db.Exec(`ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS description TEXT NULL`)
	_, _ = db.Exec(`ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS result VARCHAR(50) NOT NULL DEFAULT 'SUCCESS'`)
	_, _ = db.Exec(`ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS error_message TEXT NULL`)
	_, _ = db.Exec(`ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS before_data JSON NULL`)
	_, _ = db.Exec(`ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS after_data JSON NULL`)
	_, _ = db.Exec(`ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS changes JSON NULL`)
	_, _ = db.Exec(`ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS metadata JSON NULL`)
	_, _ = db.Exec(`ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS ip_address VARCHAR(100) NULL`)
	_, _ = db.Exec(`ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS user_agent VARCHAR(500) NULL`)

	log.Println("Migration applied successfully")
}


