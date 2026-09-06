package domain

import (
	"encoding/json"
	"time"
)

// CredentialField describes an input field required to authenticate with a carrier API.
type CredentialField struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Type        string `json:"type"` // "text", "password"
	Required    bool   `json:"required"`
	Placeholder string `json:"placeholder,omitempty"`
	Description string `json:"description,omitempty"`
}

// CarrierProvider represents a globally registered shipping line or logistics provider.
type CarrierProvider struct {
	ID                    int64             `json:"id" db:"id"`
	Code                  string            `json:"code" db:"code"`                                     // e.g. "MAERSK", "MSC", "HAPAG_LLOYD"
	Name                  string            `json:"name" db:"name"`                                     // e.g. "A.P. Moller – Maersk"
	SCAC                  string            `json:"scac" db:"scac"`                                     // e.g. "MAEU", "MSCU"
	ModesJSON             string            `json:"-" db:"modes"`                                       // JSON array in DB
	Modes                 []string          `json:"modes"`                                              // Parsed modes: ["OCEAN", "INTERMODAL"]
	AdapterKey            string            `json:"adapter_key" db:"adapter_key"`                       // e.g. "MAERSK_ADAPTER"
	IsActive              bool              `json:"is_active" db:"is_active"`                           // Supported globally
	SupportedCapsJSON     string            `json:"-" db:"supported_capabilities"`                      // JSON array in DB
	SupportedCapabilities []Capability      `json:"supported_capabilities"`                             // ["TRACKING", "RATES", "BOOKING"]
	CredentialFields      []CredentialField `json:"credential_fields"`                                  // Dynamic credential schema
	Description           *string           `json:"description,omitempty" db:"description"`             // Provider overview
	DocumentationURL      *string           `json:"documentation_url,omitempty" db:"documentation_url"` // Developer docs URL
	LogoURL               *string           `json:"logo_url,omitempty" db:"logo_url"`                   // SVG / PNG logo path
	CreatedAt             time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at" db:"updated_at"`
}

// UnmarshalDBFields deserializes raw JSON fields from DB columns into Go slices.
func (cp *CarrierProvider) UnmarshalDBFields() {
	if cp.ModesJSON != "" {
		_ = json.Unmarshal([]byte(cp.ModesJSON), &cp.Modes)
	}
	if len(cp.Modes) == 0 {
		cp.Modes = []string{"OCEAN"}
	}

	if cp.SupportedCapsJSON != "" {
		var rawCaps []string
		if err := json.Unmarshal([]byte(cp.SupportedCapsJSON), &rawCaps); err == nil {
			cp.SupportedCapabilities = make([]Capability, 0, len(rawCaps))
			for _, rc := range rawCaps {
				if capEnum, ok := ParseCapability(rc); ok {
					cp.SupportedCapabilities = append(cp.SupportedCapabilities, capEnum)
				}
			}
		}
	}
	if len(cp.SupportedCapabilities) == 0 {
		cp.SupportedCapabilities = []Capability{CapTracking, CapRates, CapBooking, CapDocuments}
	}

	// Dynamic provider-aware credential schemas
	switch cp.Code {
	case "MAERSK":
		cp.CredentialFields = []CredentialField{
			{Key: "api_key", Label: "Consumer Key / API Key", Type: "password", Required: true, Placeholder: "Enter Maersk Developer API Key", Description: "Generated in the Maersk Developer Portal"},
			{Key: "api_secret", Label: "Consumer Secret / API Secret", Type: "password", Required: true, Placeholder: "Enter Maersk Secret", Description: "Secret key for OAuth2 client credentials"},
			{Key: "client_id", Label: "Customer Client ID (Optional)", Type: "text", Required: false, Placeholder: "e.g. LOGISTIQ-MAEU-01", Description: "Internal reference ID"},
		}
	case "MSC":
		cp.CredentialFields = []CredentialField{
			{Key: "api_key", Label: "Subscription Key (Ocp-Apim-Subscription-Key)", Type: "password", Required: true, Placeholder: "Enter MSC API Subscription Key", Description: "Primary subscription key from MSC Developer Portal"},
			{Key: "client_id", Label: "OAuth Client ID", Type: "text", Required: true, Placeholder: "Enter MSC OAuth Client ID", Description: "Azure AD B2C application client ID"},
			{Key: "client_secret", Label: "OAuth Client Secret", Type: "password", Required: true, Placeholder: "Enter MSC Client Secret", Description: "Secret key for token generation"},
			{Key: "account_id", Label: "Customer Account Number", Type: "text", Required: false, Placeholder: "e.g. MSC-904128", Description: "Commercial contract account identifier"},
		}
	case "HAPAG_LLOYD":
		cp.CredentialFields = []CredentialField{
			{Key: "api_key", Label: "API Key", Type: "password", Required: true, Placeholder: "Enter Hapag-Lloyd API Key", Description: "Client API Key from Hapag-Lloyd Developer Portal"},
			{Key: "client_secret", Label: "Client Secret", Type: "password", Required: true, Placeholder: "Enter Client Secret", Description: "Secret paired with API Key"},
			{Key: "account_id", Label: "Web Account / Customer ID", Type: "text", Required: false, Placeholder: "e.g. HL-892110", Description: "Booking and tariff account number"},
		}
	case "CMA_CGM":
		cp.CredentialFields = []CredentialField{
			{Key: "api_key", Label: "API Key", Type: "password", Required: true, Placeholder: "Enter CMA CGM API Key", Description: "Issued in CMA CGM API Gateway"},
			{Key: "client_id", Label: "Partner Client ID", Type: "text", Required: false, Placeholder: "e.g. CMA-PARTNER-01"},
			{Key: "client_secret", Label: "Partner Secret", Type: "password", Required: false, Placeholder: "Enter Partner Secret"},
		}
	case "ONE":
		cp.CredentialFields = []CredentialField{
			{Key: "api_key", Label: "ONE API Key", Type: "password", Required: true, Placeholder: "Enter ONE API Key", Description: "API Gateway token from Ocean Network Express"},
			{Key: "account_id", Label: "Customer Code", Type: "text", Required: false, Placeholder: "e.g. ONE-CUST-4421"},
		}
	case "EVERGREEN":
		cp.CredentialFields = []CredentialField{
			{Key: "api_key", Label: "ShipmentLink User ID / API Key", Type: "text", Required: true, Placeholder: "Enter Evergreen User ID / API Key"},
			{Key: "client_secret", Label: "ShipmentLink Password / Secret", Type: "password", Required: true, Placeholder: "Enter ShipmentLink Password"},
		}
	case "COSCO":
		cp.CredentialFields = []CredentialField{
			{Key: "api_key", Label: "SynCon Hub App Key", Type: "password", Required: true, Placeholder: "Enter COSCO App Key", Description: "Issued in COSCO SynCon Hub Developer Portal"},
			{Key: "client_secret", Label: "SynCon Hub App Secret", Type: "password", Required: true, Placeholder: "Enter COSCO App Secret"},
		}
	default:
		cp.CredentialFields = []CredentialField{
			{Key: "api_key", Label: "API Key / Access Token", Type: "password", Required: true, Placeholder: "Enter carrier API Key or Token"},
			{Key: "api_secret", Label: "API Secret (Optional)", Type: "password", Required: false, Placeholder: "Enter API Secret if required"},
			{Key: "base_url", Label: "Custom Endpoint URL (Optional)", Type: "text", Required: false, Placeholder: "https://api.carrier.com/v1"},
		}
	}
}
