CREATE TABLE IF NOT EXISTS org_notification_preferences (
    org_id BIGINT PRIMARY KEY,
    new_rfq_received BOOLEAN DEFAULT TRUE,
    new_quote_received BOOLEAN DEFAULT TRUE,
    shipment_status_updates BOOLEAN DEFAULT TRUE,
    shipment_exceptions BOOLEAN DEFAULT TRUE,
    invitation_accepted BOOLEAN DEFAULT TRUE,
    invoice_payment_events BOOLEAN DEFAULT TRUE,
    system_security_alerts BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE
);
