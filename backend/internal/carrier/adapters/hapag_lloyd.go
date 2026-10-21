package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/freel/backend/internal/carrier/adapters/httpclient"
	"github.com/freel/backend/internal/carrier/domain"
)

// HapagLloydAdapter integrates Hapag-Lloyd API (SCAC: HLCU).
type HapagLloydAdapter struct {
	BaseAdapter
}

// NewHapagLloydAdapter constructs the official Hapag-Lloyd carrier adapter.
func NewHapagLloydAdapter() *HapagLloydAdapter {
	return &HapagLloydAdapter{
		BaseAdapter: BaseAdapter{
			Code:        "HAPAG_LLOYD",
			CarrierSCAC: "HLCU",
			Capabilities: []domain.Capability{
				domain.CapTracking,
			},
		},
	}
}

func (a *HapagLloydAdapter) getBaseURL(creds domain.DecryptedCredentials, env domain.Environment) string {
	if creds.BaseURL != "" {
		return creds.BaseURL
	}
	if env == domain.EnvSandbox {
		return "https://api.hlag.com/sandbox"
	}
	return "https://api.hlag.com"
}

func (a *HapagLloydAdapter) getHTTPClient(creds domain.DecryptedCredentials, env domain.Environment) *httpclient.CarrierHTTPClient {
	return httpclient.NewCarrierHTTPClient(httpclient.HTTPClientConfig{
		BaseURL:        a.getBaseURL(creds, env),
		Timeout:        15 * time.Second,
		MaxRetries:     2,
		InitialBackoff: 200 * time.Millisecond,
		ProviderCode:   a.Code,
	})
}

func (a *HapagLloydAdapter) getCredentials(creds domain.DecryptedCredentials) (clientID, clientSecret string) {
	return creds.GetConsumerKey(), creds.GetAPISecret()
}

func (a *HapagLloydAdapter) buildHeaders(creds domain.DecryptedCredentials) map[string]string {
	headers := map[string]string{
		"Accept": "application/json",
	}
	clientID, clientSecret := a.getCredentials(creds)
	if clientID != "" {
		headers["X-IBM-Client-Id"] = clientID
	}
	if clientSecret != "" {
		headers["X-IBM-Client-Secret"] = clientSecret
	}
	return headers
}

// TestConnection verifies connectivity and credentials against Hapag-Lloyd API Gateway.
func (a *HapagLloydAdapter) TestConnection(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment) (*domain.TestConnectionResult, error) {
	start := time.Now()

	clientID, _ := a.getCredentials(creds)
	if clientID == "" {
		return &domain.TestConnectionResult{
			Success:            false,
			Message:            "Missing Hapag-Lloyd Client ID / API Key. Please provide valid developer credentials.",
			LatencyMs:          time.Since(start).Milliseconds(),
			TestedEnvironment:  env,
			TestedCapabilities: a.Capabilities,
			ErrorCode:          domain.ErrCodeInvalidConfig,
			HTTPStatus:         http.StatusBadRequest,
		}, nil
	}

	client := a.getHTTPClient(creds, env)
	headers := a.buildHeaders(creds)

	_, statusCode, err := client.Execute(ctx, "GET", "/track-and-trace/v2/events?limit=1", headers, nil)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return &domain.TestConnectionResult{
			Success:            false,
			Message:            fmt.Sprintf("Hapag-Lloyd connection check returned: %s", domain.SanitizeError(err)),
			LatencyMs:          latency,
			TestedEnvironment:  env,
			TestedCapabilities: a.Capabilities,
			ErrorCode:          domain.ErrCodeAuthFailed,
			HTTPStatus:         statusCode,
		}, nil
	}

	return &domain.TestConnectionResult{
		Success:            true,
		Message:            "Successfully verified connection to Hapag-Lloyd API Gateway.",
		LatencyMs:          latency,
		TestedEnvironment:  env,
		TestedCapabilities: a.Capabilities,
		HTTPStatus:         statusCode,
	}, nil
}

// GetTracking queries Hapag-Lloyd DCSA Track & Trace API.
func (a *HapagLloydAdapter) GetTracking(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.TrackingRequest) (*domain.NormalizedTrackingResult, error) {
	if err := a.CheckCapability(domain.CapTracking); err != nil {
		return nil, err
	}

	client := a.getHTTPClient(creds, env)
	headers := a.buildHeaders(creds)

	queryParams := url.Values{}
	ref := strings.TrimSpace(req.ContainerNumber)
	bkgRef := strings.TrimSpace(req.BookingNumber)
	mblRef := strings.TrimSpace(req.MBLNumber)

	if ref != "" {
		queryParams.Set("equipmentReference", ref)
	} else if bkgRef != "" {
		queryParams.Set("carrierBookingReference", bkgRef)
		ref = bkgRef
	} else if mblRef != "" {
		queryParams.Set("transportDocumentReference", mblRef)
		ref = mblRef
	} else {
		return nil, domain.NewIntegrationError(
			a.Code,
			"GET_TRACKING",
			domain.ErrCodeInvalidRequest,
			"Tracking identifier required for Hapag-Lloyd",
			http.StatusBadRequest,
			false,
			fmt.Errorf("missing tracking identifier"),
		)
	}

	endpoint := fmt.Sprintf("/track-and-trace/v2/events?%s", queryParams.Encode())
	respBody, _, err := client.Execute(ctx, "GET", endpoint, headers, nil)
	if err != nil {
		return nil, err
	}

	trimmed := strings.TrimSpace(string(respBody))
	if trimmed == "" || trimmed == "[]" {
		return NormalizeDCSATrackingEvents(a.CarrierSCAC, ref, []domain.DCSAEvent{}), nil
	}

	var events []domain.DCSAEvent
	if err := json.Unmarshal(respBody, &events); err != nil {
		return nil, domain.NewIntegrationError(
			a.Code,
			"GET_TRACKING",
			domain.ErrCodeInternal,
			"Failed to parse Hapag-Lloyd Track & Trace events",
			http.StatusInternalServerError,
			false,
			err,
		)
	}

	return NormalizeDCSATrackingEvents(a.CarrierSCAC, ref, events), nil
}
