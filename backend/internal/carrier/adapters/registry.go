package adapters

import (
	"context"
	"fmt"
	"strings"

	"github.com/freel/backend/internal/carrier"
)

// resolveConfiguredAdapter resolves and configures the tracking/webhook/booking adapter for org+scac
func resolveConfiguredAdapter(db carrier.Queryer, orgID int64, scac string) (*MockTrackingAdapter, error) {
	scacUpper := strings.ToUpper(scac)
	
	// Valid carriers check
	validCarriers := map[string]bool{
		"MAEU":  true,
		"MSK":   true,
		"MSC":   true,
		"CMA":   true,
		"CMDU":  true,
		"HAPAG": true,
		"HLCU":  true,
	}
	if !validCarriers[scacUpper] {
		return nil, fmt.Errorf("unknown carrier: SCAC code %s not recognized", scac)
	}

	// Dynamic lookup in carrier_integrations DB table
	cfg, err := carrier.GetIntegrationConfig(context.Background(), db, orgID, scac)
	if err != nil {
		return nil, err
	}

	return NewMockTrackingAdapter(scacUpper, cfg), nil
}

// GetTrackingProvider resolves a TrackingProvider by org ID and carrier SCAC
func GetTrackingProvider(db carrier.Queryer, orgID int64, scac string) (carrier.TrackingProvider, error) {
	return resolveConfiguredAdapter(db, orgID, scac)
}

// GetBookingProvider resolves a BookingProvider by org ID and carrier SCAC
func GetBookingProvider(db carrier.Queryer, orgID int64, scac string) (carrier.BookingProvider, error) {
	return resolveConfiguredAdapter(db, orgID, scac)
}

// GetWebhookProvider resolves a WebhookProvider by org ID and carrier SCAC
func GetWebhookProvider(db carrier.Queryer, orgID int64, scac string) (carrier.WebhookProvider, error) {
	return resolveConfiguredAdapter(db, orgID, scac)
}
