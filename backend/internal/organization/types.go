package organization

import "time"

// Organization represents a tenant in the system.
type Organization struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"` // Used as internal workspace name or sync with legal_name
	
	// Profile fields
	LegalName          *string `json:"legal_name" db:"legal_name"`
	RegistrationNumber *string `json:"registration_number" db:"registration_number"`
	TaxNumber          *string `json:"tax_number" db:"tax_number"`
	Website            *string `json:"website" db:"website"`
	
	PrimaryEmail       *string `json:"primary_email" db:"primary_email"`
	PhoneNumber        *string `json:"phone_number" db:"phone_number"`
	SupportEmail       *string `json:"support_email" db:"support_email"`
	
	Address            *string `json:"address" db:"address"`
	City               *string `json:"city" db:"city"`
	State              *string `json:"state" db:"state"`
	Country            *string `json:"country" db:"country"`
	PostalCode         *string `json:"postal_code" db:"postal_code"`
	
	Industry           *string `json:"industry" db:"industry"`
	CompanyType        *string `json:"company_type" db:"company_type"`
	
	DefaultCurrency    *string `json:"default_currency" db:"default_currency"`
	DefaultTimezone    *string `json:"default_timezone" db:"default_timezone"`
	DateFormat         *string `json:"date_format" db:"date_format"`
	LogoURL            *string `json:"logo_url" db:"logo_url"`

	// Workspace Preferences (Phase 1)
	DefaultLanguage    *string `json:"default_language" db:"default_language"`
	MeasurementSystem  *string `json:"measurement_system" db:"measurement_system"`
	WeightUnit         *string `json:"weight_unit" db:"weight_unit"`
	DimensionUnit      *string `json:"dimension_unit" db:"dimension_unit"`
	VolumeUnit         *string `json:"volume_unit" db:"volume_unit"`
	TimeFormat         *string `json:"time_format" db:"time_format"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateOrgRequest represents the payload for creating an organization.
type CreateOrgRequest struct {
	Name string `json:"name" validate:"required"`
}

// UpdateOrgRequest represents the payload for updating an organization (legacy)
type UpdateOrgRequest struct {
	Name string `json:"name" validate:"required"`
}

// CompanyProfileUpdateRequest represents the payload for updating the full company profile.
type CompanyProfileUpdateRequest struct {
	Name               string  `json:"name" validate:"required"`
	LegalName          *string `json:"legal_name"`
	RegistrationNumber *string `json:"registration_number"`
	TaxNumber          *string `json:"tax_number"`
	Website            *string `json:"website"`
	PrimaryEmail       *string `json:"primary_email"`
	PhoneNumber        *string `json:"phone_number"`
	SupportEmail       *string `json:"support_email"`
	Address            *string `json:"address"`
	City               *string `json:"city"`
	State              *string `json:"state"`
	Country            *string `json:"country"`
	PostalCode         *string `json:"postal_code"`
	Industry           *string `json:"industry"`
	CompanyType        *string `json:"company_type"`
	DefaultCurrency    *string `json:"default_currency"`
	DefaultTimezone    *string `json:"default_timezone"`
	DateFormat         *string `json:"date_format"`
	
	// Workspace Preferences
	DefaultLanguage    *string `json:"default_language"`
	MeasurementSystem  *string `json:"measurement_system"`
	WeightUnit         *string `json:"weight_unit"`
	DimensionUnit      *string `json:"dimension_unit"`
	VolumeUnit         *string `json:"volume_unit"`
	TimeFormat         *string `json:"time_format"`
}

// InviteRequest represents the payload for inviting a user.
type InviteRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required"`
}

// NotificationPreferences represents the notification preferences for an organization.
type NotificationPreferences struct {
	OrgID                 int64     `json:"org_id" db:"org_id"`
	NewRFQReceived        bool      `json:"new_rfq_received" db:"new_rfq_received"`
	NewQuoteReceived      bool      `json:"new_quote_received" db:"new_quote_received"`
	ShipmentStatusUpdates bool      `json:"shipment_status_updates" db:"shipment_status_updates"`
	ShipmentExceptions    bool      `json:"shipment_exceptions" db:"shipment_exceptions"`
	InvitationAccepted    bool      `json:"invitation_accepted" db:"invitation_accepted"`
	InvoicePaymentEvents  bool      `json:"invoice_payment_events" db:"invoice_payment_events"`
	SystemSecurityAlerts  bool      `json:"system_security_alerts" db:"system_security_alerts"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// UpdateNotificationPreferencesRequest represents the payload to update preferences.
type UpdateNotificationPreferencesRequest struct {
	NewRFQReceived        bool `json:"new_rfq_received"`
	NewQuoteReceived      bool `json:"new_quote_received"`
	ShipmentStatusUpdates bool `json:"shipment_status_updates"`
	ShipmentExceptions    bool `json:"shipment_exceptions"`
	InvitationAccepted    bool `json:"invitation_accepted"`
	InvoicePaymentEvents  bool `json:"invoice_payment_events"`
	SystemSecurityAlerts  bool `json:"system_security_alerts"`
}

// EmailSettings represents the AI processing settings for an organization.
type EmailSettings struct {
	OrgID                     int64     `json:"org_id" db:"org_id"`
	ProcessLogisticsInquiries bool      `json:"process_logistics_inquiries" db:"process_logistics_inquiries"`
	TrackEmailThreads         bool      `json:"track_email_threads" db:"track_email_threads"`
	SmartFiltering            bool      `json:"smart_filtering" db:"smart_filtering"`
	CreatedAt                 time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at" db:"updated_at"`
}

// UpdateEmailSettingsRequest represents the payload to update email settings.
type UpdateEmailSettingsRequest struct {
	ProcessLogisticsInquiries bool `json:"process_logistics_inquiries"`
	TrackEmailThreads         bool `json:"track_email_threads"`
	SmartFiltering            bool `json:"smart_filtering"`
}

// ConnectedMailbox represents an email mailbox connected to the workspace.
type ConnectedMailbox struct {
	ID                     int64      `json:"id" db:"id"`
	OrgID                  int64      `json:"org_id" db:"org_id"`
	Email                  string     `json:"email" db:"email"`
	OwnerName              string     `json:"owner_name" db:"owner_name"`
	MailboxType            string     `json:"mailbox_type" db:"mailbox_type"` // e.g., 'Shared / Team', 'Individual'
	IsPrimary              bool       `json:"is_primary" db:"is_primary"`
	Status                 string     `json:"status" db:"status"` // e.g., 'PENDING', 'CONNECTED', 'SYNCING', 'ERROR', 'DISCONNECTED'
	LastSyncedAt           time.Time  `json:"last_synced_at" db:"last_synced_at"`
	SyncFrequency          string     `json:"sync_frequency" db:"sync_frequency"`
	ProcessingEnabled      bool       `json:"processing_enabled" db:"processing_enabled"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
	Provider               string     `json:"provider" db:"provider"` // GMAIL, MICROSOFT, IMAP
	AccessTokenEncrypted   *string    `json:"-" db:"access_token_encrypted"` // Never serialize to JSON
	RefreshTokenEncrypted  *string    `json:"-" db:"refresh_token_encrypted"` // Never serialize to JSON
	TokenExpiry            *time.Time `json:"token_expiry" db:"token_expiry"`
	OAuthScopes            *string    `json:"oauth_scopes" db:"oauth_scopes"`
	SyncCursor             *string    `json:"sync_cursor" db:"sync_cursor"`
	LastSyncStartedAt      *time.Time `json:"last_sync_started_at" db:"last_sync_started_at"`
	LastSyncSuccessAt      *time.Time `json:"last_sync_success_at" db:"last_sync_success_at"`
	LastSyncError          *string    `json:"last_sync_error" db:"last_sync_error"`
	LastProcessedMessageID *string    `json:"last_processed_message_id" db:"last_processed_message_id"`
}

// ConnectMailboxRequest represents the payload to add a new connected mailbox.
type ConnectMailboxRequest struct {
	Email        string     `json:"email" validate:"required,email"`
	OwnerName    string     `json:"owner_name" validate:"required"`
	MailboxType  string     `json:"mailbox_type" validate:"required"`
	Provider     string     `json:"provider" validate:"required,oneof=GMAIL MICROSOFT IMAP"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	TokenExpiry  *time.Time `json:"token_expiry"`
	OAuthScopes  []string   `json:"oauth_scopes"`
}

// UpdateMailboxRequest represents the payload to update an existing connected mailbox.
type UpdateMailboxRequest struct {
	OwnerName         string `json:"owner_name" validate:"required"`
	MailboxType       string `json:"mailbox_type" validate:"required"`
	SyncFrequency     string `json:"sync_frequency" validate:"required"`
	ProcessingEnabled bool   `json:"processing_enabled"`
}

// CarrierIntegration represents a carrier connection for the workspace.
type CarrierIntegration struct {
	ID               int64      `json:"id" db:"id"`
	OrgID            int64      `json:"org_id" db:"org_id"`
	CarrierSCAC      string     `json:"carrier_scac" db:"carrier_scac"`
	CarrierName      string     `json:"carrier_name" db:"carrier_name"`
	ConnectionMethod string     `json:"connection_method" db:"connection_method"`
	CredentialsJSON  *string    `json:"-" db:"credentials_json"` // Never return credentials to frontend
	Environment      string     `json:"environment" db:"environment"`
	ConnectionStatus string     `json:"connection_status" db:"connection_status"`
	SyncStatus       *string    `json:"sync_status" db:"sync_status"`
	LastSyncedAt     *time.Time `json:"last_synced_at" db:"last_synced_at"`
	Capabilities     *string    `json:"capabilities" db:"capabilities"`
	FailedAttempts   int        `json:"failed_attempts" db:"failed_attempts"`
	ErrorDetails     *string    `json:"error_details" db:"error_details"`
	NextRetryTime    *time.Time `json:"next_retry_time" db:"next_retry_time"`
	IsActive         bool       `json:"is_active" db:"is_active"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// ConnectCarrierRequest represents the payload to add a new carrier integration.
type ConnectCarrierRequest struct {
	CarrierSCAC      string  `json:"carrier_scac" validate:"required"`
	ConnectionMethod string  `json:"connection_method" validate:"required"`
	Environment      string  `json:"environment" validate:"required"`
	CredentialsJSON  string  `json:"credentials_json" validate:"required"`
	Capabilities     *string `json:"capabilities"`
}

// UpdateCarrierRequest represents the payload to update an existing carrier integration.
type UpdateCarrierRequest struct {
	ConnectionMethod string  `json:"connection_method"`
	Environment      string  `json:"environment"`
	CredentialsJSON  string  `json:"credentials_json"`
	Capabilities     *string `json:"capabilities"`
}
