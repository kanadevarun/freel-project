package domain

import (
	"encoding/json"
	"time"
)

// CarrierIntegration represents the tenant-specific carrier connection record.
type CarrierIntegration struct {
	ID                   int64            `json:"id" db:"id"`
	OrgID                int64            `json:"org_id" db:"org_id"`
	CarrierProviderID    *int64           `json:"carrier_provider_id,omitempty" db:"carrier_provider_id"`
	CarrierSCAC          string           `json:"carrier_scac" db:"carrier_scac"`
	CarrierName          *string          `json:"carrier_name,omitempty" db:"carrier_name"`
	ConnectionMethod     ConnectionMethod `json:"connection_method" db:"connection_method"`
	Environment          Environment      `json:"environment" db:"environment"`
	ConnectionStatus     ConnectionStatus `json:"connection_status" db:"connection_status"`
	IsEnabled            bool             `json:"is_active" db:"is_active"`
	CredentialsJSON      *string          `json:"-" db:"credentials_json"`         // Legacy JSON
	EncryptedCredentials *string          `json:"-" db:"encrypted_credentials"`    // AES-256-GCM encrypted
	CredentialMaskJSON   *string          `json:"-" db:"credential_mask"`          // JSON string of masked keys
	CapabilitiesJSON     *string          `json:"-" db:"capabilities"`             // JSON string of enabled capabilities
	ConfigOptionsJSON    *string          `json:"-" db:"config_options"`            // JSON string of custom config
	SyncStatus           *string          `json:"sync_status,omitempty" db:"sync_status"`
	LastSyncedAt         *time.Time       `json:"last_synced_at,omitempty" db:"last_synced_at"`
	LastSuccessAt        *time.Time       `json:"last_success_at,omitempty" db:"last_success_at"`
	LastFailureAt        *time.Time       `json:"last_failure_at,omitempty" db:"last_failure_at"`
	LastError            *string          `json:"last_error,omitempty" db:"last_error"`
	FailedAttempts       int              `json:"failed_attempts" db:"failed_attempts"`
	ErrorDetails         *string          `json:"-" db:"error_details"`
	NextRetryTime        *time.Time       `json:"next_retry_time,omitempty" db:"next_retry_time"`
	CreatedAt            time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at" db:"updated_at"`

	// Parsed runtime fields
	Capabilities []Capability           `json:"capabilities"`
	Config       map[string]interface{} `json:"config,omitempty"`
}

// UnmarshalRuntimeFields parses DB JSON strings into typed Go structures.
func (ci *CarrierIntegration) UnmarshalRuntimeFields() {
	if ci.CapabilitiesJSON != nil && *ci.CapabilitiesJSON != "" {
		// Can be array of strings or map of string->bool
		var strSlice []string
		if err := json.Unmarshal([]byte(*ci.CapabilitiesJSON), &strSlice); err == nil {
			ci.Capabilities = make([]Capability, 0, len(strSlice))
			for _, s := range strSlice {
				if c, ok := ParseCapability(s); ok {
					ci.Capabilities = append(ci.Capabilities, c)
				}
			}
		} else {
			var boolMap map[string]bool
			if err := json.Unmarshal([]byte(*ci.CapabilitiesJSON), &boolMap); err == nil {
				ci.Capabilities = make([]Capability, 0, len(boolMap))
				for k, v := range boolMap {
					if v {
						if c, ok := ParseCapability(k); ok {
							ci.Capabilities = append(ci.Capabilities, c)
						}
					}
				}
			}
		}
	}
	if len(ci.Capabilities) == 0 {
		ci.Capabilities = []Capability{CapTracking}
	}

	if ci.ConfigOptionsJSON != nil && *ci.ConfigOptionsJSON != "" {
		_ = json.Unmarshal([]byte(*ci.ConfigOptionsJSON), &ci.Config)
	}
}

// CarrierIntegrationView is the safe representation sent to frontend clients.
// Under no circumstances does this struct expose decrypted secrets or raw keys.
type CarrierIntegrationView struct {
	ID                  int64                  `json:"id"`
	OrgID               int64                  `json:"org_id"`
	CarrierSCAC         string                 `json:"carrier_scac"`
	CarrierName         string                 `json:"carrier_name"`
	CarrierCode         string                 `json:"carrier_code,omitempty"`
	Environment         Environment            `json:"environment"`
	ConnectionMethod    ConnectionMethod       `json:"connection_method"`
	ConnectionStatus    ConnectionStatus       `json:"connection_status"`
	HealthState         IntegrationHealthState `json:"health_state"`
	HealthReason        string                 `json:"health_reason,omitempty"`
	IsEnabled           bool                   `json:"is_active"`
	Capabilities        []Capability           `json:"capabilities"`
	SupportedCaps       []Capability           `json:"supported_capabilities"`
	CredentialMask      map[string]string      `json:"credentials_mask"`
	HasCredentials      bool                   `json:"has_credentials"`
	SyncStatus          string                 `json:"sync_status"`
	IsSyncing           bool                   `json:"is_syncing"`
	LastSyncedAt        *time.Time             `json:"last_synced_at,omitempty"`
	LastSuccessAt       *time.Time             `json:"last_success_at,omitempty"`
	LastFailureAt       *time.Time             `json:"last_failure_at,omitempty"`
	LastError           *string                `json:"last_error,omitempty"`
	FailedAttempts      int                    `json:"failed_attempts"`
	ConsecutiveFailures int                    `json:"consecutive_failures"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

// ConnectCarrierRequest defines the payload to register a new carrier integration.
type ConnectCarrierRequest struct {
	CarrierSCAC      string                 `json:"carrier_scac"`
	Environment      string                 `json:"environment"`
	ConnectionMethod string                 `json:"connection_method"`
	Credentials      map[string]interface{} `json:"credentials,omitempty"`
	CredentialsJSON  string                 `json:"credentials_json,omitempty"` // For UI string compatibility
	Capabilities     interface{}            `json:"capabilities,omitempty"`     // map[string]bool or []string
	ConfigOptions    map[string]interface{} `json:"config_options,omitempty"`
}

// UpdateCarrierRequest defines the payload to update an existing carrier integration.
type UpdateCarrierRequest struct {
	Environment      *string                `json:"environment,omitempty"`
	ConnectionMethod *string                `json:"connection_method,omitempty"`
	Credentials      map[string]interface{} `json:"credentials,omitempty"`
	CredentialsJSON  *string                `json:"credentials_json,omitempty"`
	Capabilities     interface{}            `json:"capabilities,omitempty"`
	ConfigOptions    map[string]interface{} `json:"config_options,omitempty"`
	IsEnabled        *bool                  `json:"is_active,omitempty"`
}

// TestDirectRequest allows pre-flight testing of credentials before persisting.
type TestDirectRequest struct {
	CarrierSCAC      string                 `json:"carrier_scac"`
	Environment      string                 `json:"environment"`
	ConnectionMethod string                 `json:"connection_method"`
	Credentials      map[string]interface{} `json:"credentials,omitempty"`
	CredentialsJSON  string                 `json:"credentials_json,omitempty"`
}
