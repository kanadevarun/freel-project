package organization

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// Repository defines the data access methods for Organization.
type Repository interface {
	Create(ctx context.Context, org *Organization) error
	GetByID(ctx context.Context, id int64) (*Organization, error)
	Update(ctx context.Context, org *Organization) error
	UpdateLogo(ctx context.Context, id int64, logoURL string) error
	GetNotificationPreferences(ctx context.Context, orgID int64) (*NotificationPreferences, error)
	UpdateNotificationPreferences(ctx context.Context, prefs *NotificationPreferences) error
	GetEmailSettings(ctx context.Context, orgID int64) (*EmailSettings, error)
	UpdateEmailSettings(ctx context.Context, prefs *EmailSettings) error
	GetConnectedMailboxes(ctx context.Context, orgID int64) ([]ConnectedMailbox, error)
	GetConnectedMailboxByID(ctx context.Context, id int64, orgID int64) (*ConnectedMailbox, error)
	AddConnectedMailbox(ctx context.Context, mailbox *ConnectedMailbox) error
	UpdateConnectedMailbox(ctx context.Context, mailbox *ConnectedMailbox) error
	DeleteConnectedMailbox(ctx context.Context, id int64, orgID int64) error
	UpdateMailboxStatus(ctx context.Context, id int64, orgID int64, status string, updateSync bool) error
	GetCarrierIntegrations(ctx context.Context, orgID int64) ([]CarrierIntegration, error)
	GetCarrierIntegrationByID(ctx context.Context, id int64, orgID int64) (*CarrierIntegration, error)
	AddCarrierIntegration(ctx context.Context, ci *CarrierIntegration) error
	UpdateCarrierIntegration(ctx context.Context, ci *CarrierIntegration) error
	DeleteCarrierIntegration(ctx context.Context, id int64, orgID int64) error
	UpdateCarrierIntegrationStatus(ctx context.Context, id int64, orgID int64, connStatus string, syncStatus string, updateSyncTime bool) error
	GetDB() *sqlx.DB
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetDB() *sqlx.DB {
	return r.db
}

func (r *repository) Create(ctx context.Context, org *Organization) error {
	query := `INSERT INTO organizations (name, created_at, updated_at) VALUES (?, NOW(), NOW())`
	res, err := r.db.ExecContext(ctx, query, org.Name)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		org.ID = id
	}
	return nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (*Organization, error) {
	var org Organization
	query := `SELECT 
		id, name, 
		legal_name, registration_number, tax_number, website, 
		primary_email, phone_number, support_email, 
		address, city, state, country, postal_code, 
		industry, company_type, default_currency, default_timezone, date_format, logo_url,
		default_language, measurement_system, weight_unit, dimension_unit, volume_unit, time_format,
		created_at, updated_at 
	FROM organizations WHERE id = ?`
	err := r.db.GetContext(ctx, &org, query, id)
	return &org, err
}

func (r *repository) Update(ctx context.Context, org *Organization) error {
	query := `UPDATE organizations SET 
		name = :name, 
		legal_name = :legal_name,
		registration_number = :registration_number,
		tax_number = :tax_number,
		website = :website,
		primary_email = :primary_email,
		phone_number = :phone_number,
		support_email = :support_email,
		address = :address,
		city = :city,
		state = :state,
		country = :country,
		postal_code = :postal_code,
		industry = :industry,
		company_type = :company_type,
		default_currency = :default_currency,
		default_timezone = :default_timezone,
		date_format = :date_format,
		default_language = :default_language,
		measurement_system = :measurement_system,
		weight_unit = :weight_unit,
		dimension_unit = :dimension_unit,
		volume_unit = :volume_unit,
		time_format = :time_format,
		updated_at = NOW()
	WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, query, org)
	return err
}

func (r *repository) UpdateLogo(ctx context.Context, id int64, logoURL string) error {
	query := `UPDATE organizations SET logo_url = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, logoURL, id)
	return err
}

func (r *repository) GetNotificationPreferences(ctx context.Context, orgID int64) (*NotificationPreferences, error) {
	var prefs NotificationPreferences
	query := `SELECT * FROM org_notification_preferences WHERE org_id = ?`
	err := r.db.GetContext(ctx, &prefs, query, orgID)
	return &prefs, err
}

func (r *repository) UpdateNotificationPreferences(ctx context.Context, prefs *NotificationPreferences) error {
	query := `UPDATE org_notification_preferences SET 
		new_rfq_received = :new_rfq_received,
		new_quote_received = :new_quote_received,
		shipment_status_updates = :shipment_status_updates,
		shipment_exceptions = :shipment_exceptions,
		invitation_accepted = :invitation_accepted,
		invoice_payment_events = :invoice_payment_events,
		system_security_alerts = :system_security_alerts,
		updated_at = NOW()
	WHERE org_id = :org_id`
	_, err := r.db.NamedExecContext(ctx, query, prefs)
	return err
}

func (r *repository) GetEmailSettings(ctx context.Context, orgID int64) (*EmailSettings, error) {
	var settings EmailSettings
	query := `SELECT * FROM org_email_settings WHERE org_id = ?`
	err := r.db.GetContext(ctx, &settings, query, orgID)
	return &settings, err
}

func (r *repository) UpdateEmailSettings(ctx context.Context, settings *EmailSettings) error {
	query := `UPDATE org_email_settings SET 
		process_logistics_inquiries = :process_logistics_inquiries,
		track_email_threads = :track_email_threads,
		smart_filtering = :smart_filtering,
		updated_at = NOW()
	WHERE org_id = :org_id`
	_, err := r.db.NamedExecContext(ctx, query, settings)
	return err
}

func (r *repository) GetConnectedMailboxes(ctx context.Context, orgID int64) ([]ConnectedMailbox, error) {
	var mailboxes []ConnectedMailbox
	query := `SELECT * FROM org_connected_mailboxes WHERE org_id = ? ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &mailboxes, query, orgID)
	return mailboxes, err
}

func (r *repository) GetConnectedMailboxByID(ctx context.Context, id int64, orgID int64) (*ConnectedMailbox, error) {
	var mailbox ConnectedMailbox
	query := `SELECT * FROM org_connected_mailboxes WHERE id = ? AND org_id = ?`
	err := r.db.GetContext(ctx, &mailbox, query, id, orgID)
	if err != nil {
		return nil, err
	}
	return &mailbox, nil
}

func (r *repository) AddConnectedMailbox(ctx context.Context, mailbox *ConnectedMailbox) error {
	query := `INSERT INTO org_connected_mailboxes (org_id, email, owner_name, mailbox_type, is_primary, status, sync_frequency, processing_enabled, provider, access_token_encrypted, refresh_token_encrypted, token_expiry, oauth_scopes, sync_cursor, last_sync_started_at, last_sync_success_at, last_sync_error, last_processed_message_id, created_at, updated_at) 
	VALUES (:org_id, :email, :owner_name, :mailbox_type, :is_primary, :status, :sync_frequency, :processing_enabled, :provider, :access_token_encrypted, :refresh_token_encrypted, :token_expiry, :oauth_scopes, :sync_cursor, :last_sync_started_at, :last_sync_success_at, :last_sync_error, :last_processed_message_id, NOW(), NOW())`
	res, err := r.db.NamedExecContext(ctx, query, mailbox)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		mailbox.ID = id
	}
	return nil
}

func (r *repository) UpdateConnectedMailbox(ctx context.Context, mailbox *ConnectedMailbox) error {
	query := `UPDATE org_connected_mailboxes SET 
		owner_name = :owner_name,
		mailbox_type = :mailbox_type,
		sync_frequency = :sync_frequency,
		processing_enabled = :processing_enabled,
		status = :status,
		provider = :provider,
		access_token_encrypted = :access_token_encrypted,
		refresh_token_encrypted = :refresh_token_encrypted,
		token_expiry = :token_expiry,
		oauth_scopes = :oauth_scopes,
		sync_cursor = :sync_cursor,
		last_sync_started_at = :last_sync_started_at,
		last_sync_success_at = :last_sync_success_at,
		last_sync_error = :last_sync_error,
		last_processed_message_id = :last_processed_message_id,
		updated_at = NOW()
	WHERE id = :id AND org_id = :org_id`
	_, err := r.db.NamedExecContext(ctx, query, mailbox)
	return err
}

func (r *repository) DeleteConnectedMailbox(ctx context.Context, id int64, orgID int64) error {
	query := `DELETE FROM org_connected_mailboxes WHERE id = ? AND org_id = ?`
	_, err := r.db.ExecContext(ctx, query, id, orgID)
	return err
}

func (r *repository) UpdateMailboxStatus(ctx context.Context, id int64, orgID int64, status string, updateSync bool) error {
	query := `UPDATE org_connected_mailboxes SET status = ?, updated_at = NOW()`
	args := []interface{}{status}
	if updateSync {
		query += `, last_synced_at = NOW()`
	}
	query += ` WHERE id = ? AND org_id = ?`
	args = append(args, id, orgID)
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *repository) GetCarrierIntegrations(ctx context.Context, orgID int64) ([]CarrierIntegration, error) {
	var integrations []CarrierIntegration
	query := `
		SELECT 
			ci.id, ci.org_id, ci.carrier_scac, ci.connection_method, ci.environment, 
			ci.connection_status, ci.sync_status, ci.last_synced_at, ci.capabilities, 
			ci.failed_attempts, ci.error_details, ci.next_retry_time,
			ci.is_active, ci.created_at, ci.updated_at,
			COALESCE(c.name, ci.carrier_scac) as carrier_name
		FROM carrier_integrations ci
		LEFT JOIN carriers c ON ci.carrier_scac = c.scac
		WHERE ci.org_id = ?
		ORDER BY ci.created_at DESC`
	err := r.db.SelectContext(ctx, &integrations, query, orgID)
	return integrations, err
}

func (r *repository) GetCarrierIntegrationByID(ctx context.Context, id int64, orgID int64) (*CarrierIntegration, error) {
	var ci CarrierIntegration
	query := `
		SELECT 
			ci.id, ci.org_id, ci.carrier_scac, ci.connection_method, ci.environment, 
			ci.connection_status, ci.sync_status, ci.last_synced_at, ci.capabilities, 
			ci.failed_attempts, ci.error_details, ci.next_retry_time,
			ci.is_active, ci.created_at, ci.updated_at,
			COALESCE(c.name, ci.carrier_scac) as carrier_name
		FROM carrier_integrations ci
		LEFT JOIN carriers c ON ci.carrier_scac = c.scac
		WHERE ci.id = ? AND ci.org_id = ?`
	err := r.db.GetContext(ctx, &ci, query, id, orgID)
	return &ci, err
}

func (r *repository) AddCarrierIntegration(ctx context.Context, ci *CarrierIntegration) error {
	query := `
		INSERT INTO carrier_integrations (
			org_id, carrier_scac, connection_method, credentials_json, environment, 
			connection_status, capabilities, is_active, created_at, updated_at
		) VALUES (
			:org_id, :carrier_scac, :connection_method, :credentials_json, :environment, 
			:connection_status, :capabilities, :is_active, NOW(), NOW()
		)`
	res, err := r.db.NamedExecContext(ctx, query, ci)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		ci.ID = id
	}
	return nil
}

func (r *repository) UpdateCarrierIntegration(ctx context.Context, ci *CarrierIntegration) error {
	query := `
		UPDATE carrier_integrations SET 
			connection_method = :connection_method,
			credentials_json = COALESCE(:credentials_json, credentials_json),
			environment = :environment,
			capabilities = :capabilities,
			updated_at = NOW()
		WHERE id = :id AND org_id = :org_id`
	_, err := r.db.NamedExecContext(ctx, query, ci)
	return err
}

func (r *repository) DeleteCarrierIntegration(ctx context.Context, id int64, orgID int64) error {
	query := `DELETE FROM carrier_integrations WHERE id = ? AND org_id = ?`
	_, err := r.db.ExecContext(ctx, query, id, orgID)
	return err
}

func (r *repository) UpdateCarrierIntegrationStatus(ctx context.Context, id int64, orgID int64, connStatus string, syncStatus string, updateSyncTime bool) error {
	query := `UPDATE carrier_integrations SET connection_status = ?, sync_status = ?, updated_at = NOW()`
	args := []interface{}{connStatus, syncStatus}
	if updateSyncTime {
		query += `, last_synced_at = NOW()`
	}
	query += ` WHERE id = ? AND org_id = ?`
	args = append(args, id, orgID)
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}
