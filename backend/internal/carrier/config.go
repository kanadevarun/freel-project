package carrier

import (
	"context"
	"fmt"
	"os"
	"strings"
)

type Queryer interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type IntegrationConfig struct {
	OrgID        int64
	CarrierSCAC  string
	APIBaseURL   string
	APIKey       string
	AuthType     string
	Capabilities map[string]bool
	IsActive     bool
}

// GetIntegrationConfig retrieves organization-specific carrier settings from the DB and env.
func GetIntegrationConfig(ctx context.Context, db Queryer, orgID int64, scac string) (*IntegrationConfig, error) {
	var isActive bool
	err := db.GetContext(ctx, &isActive, `
		SELECT is_active FROM carrier_integrations 
		WHERE org_id = ? AND carrier_scac = ? LIMIT 1
	`, orgID, scac)
	if err != nil {
		return nil, fmt.Errorf("carrier integration not configured or inactive for org %d and carrier %s: %w", orgID, scac, err)
	}
	if !isActive {
		return nil, fmt.Errorf("carrier integration is disabled for org %d and carrier %s", orgID, scac)
	}

	scacUpper := strings.ToUpper(scac)
	
	// Org-specific env config lookup
	apiKey := os.Getenv(fmt.Sprintf("CARRIER_%s_API_KEY_%d", scacUpper, orgID))
	baseURL := os.Getenv(fmt.Sprintf("CARRIER_%s_BASE_URL_%d", scacUpper, orgID))
	authType := os.Getenv(fmt.Sprintf("CARRIER_%s_AUTH_TYPE_%d", scacUpper, orgID))
	capsStr := os.Getenv(fmt.Sprintf("CARRIER_%s_CAPABILITIES_%d", scacUpper, orgID))

	// Fallbacks for local development mode
	isDev := os.Getenv("APP_ENV") != "production"
	if isDev {
		if apiKey == "" {
			apiKey = os.Getenv(fmt.Sprintf("CARRIER_%s_API_KEY", scacUpper))
		}
		if apiKey == "" {
			apiKey = "dev-mock-key-12345"
		}
		if baseURL == "" {
			baseURL = os.Getenv(fmt.Sprintf("CARRIER_%s_BASE_URL", scacUpper))
		}
		if baseURL == "" {
			baseURL = "http://localhost:8099"
		}
		if authType == "" {
			authType = "API_KEY"
		}
		if capsStr == "" {
			capsStr = "TRACKING,RATES,BOOKING,WEBHOOK"
		}
	} else {
		// Production strictness: fail fast if no credentials
		if apiKey == "" {
			return nil, fmt.Errorf("production error: carrier credential missing for org %d and carrier %s", orgID, scac)
		}
		if baseURL == "" {
			return nil, fmt.Errorf("production error: carrier base URL missing for org %d and carrier %s", orgID, scac)
		}
	}

	capabilities := make(map[string]bool)
	for _, c := range strings.Split(capsStr, ",") {
		c = strings.TrimSpace(strings.ToUpper(c))
		if c != "" {
			capabilities[c] = true
		}
	}

	return &IntegrationConfig{
		OrgID:        orgID,
		CarrierSCAC:  scac,
		APIBaseURL:   baseURL,
		APIKey:       apiKey,
		AuthType:     authType,
		Capabilities: capabilities,
		IsActive:     true,
	}, nil
}
