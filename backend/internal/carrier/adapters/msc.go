package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/freel/backend/internal/carrier/adapters/httpclient"
	"github.com/freel/backend/internal/carrier/domain"
)

// MscAdapter integrates MSC Mediterranean Shipping Company API (SCAC: MSCU).
type MscAdapter struct {
	BaseAdapter
}

// NewMscAdapter constructs the official MSC carrier adapter.
func NewMscAdapter() *MscAdapter {
	return &MscAdapter{
		BaseAdapter: BaseAdapter{
			Code:        "MSC",
			CarrierSCAC: "MSCU",
			Capabilities: []domain.Capability{
				domain.CapTracking,
				domain.CapRates,
				domain.CapContractRates,
				domain.CapBooking,
				domain.CapDocuments,
			},
		},
	}
}

func (a *MscAdapter) getBaseURL(creds domain.DecryptedCredentials, env domain.Environment) string {
	if creds.BaseURL != "" {
		return creds.BaseURL
	}
	if env == domain.EnvSandbox {
		return "https://api.msc.com/sandbox"
	}
	return "https://api.msc.com"
}

func (a *MscAdapter) getHTTPClient(creds domain.DecryptedCredentials, env domain.Environment) *httpclient.CarrierHTTPClient {
	return httpclient.NewCarrierHTTPClient(httpclient.HTTPClientConfig{
		BaseURL:        a.getBaseURL(creds, env),
		Timeout:        15 * time.Second,
		MaxRetries:     2,
		InitialBackoff: 200 * time.Millisecond,
		ProviderCode:   a.Code,
	})
}

func (a *MscAdapter) buildHeaders(creds domain.DecryptedCredentials) map[string]string {
	headers := map[string]string{
		"Accept": "application/json",
	}
	if creds.APIKey != "" {
		headers["Ocp-Apim-Subscription-Key"] = creds.APIKey
	}
	return headers
}

// TestConnection verifies connectivity and credentials against MSC API Gateway.
func (a *MscAdapter) TestConnection(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment) (*domain.TestConnectionResult, error) {
	start := time.Now()

	if creds.APIKey == "" {
		return &domain.TestConnectionResult{
			Success:            false,
			Message:            "Missing MSC Subscription Key. Please configure your Ocp-Apim-Subscription-Key.",
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
			Message:            fmt.Sprintf("MSC connection check returned: %v", err),
			LatencyMs:          latency,
			TestedEnvironment:  env,
			TestedCapabilities: a.Capabilities,
			ErrorCode:          domain.ErrCodeAuthFailed,
			HTTPStatus:         statusCode,
		}, nil
	}

	return &domain.TestConnectionResult{
		Success:            true,
		Message:            "Successfully verified connection to MSC API Gateway.",
		LatencyMs:          latency,
		TestedEnvironment:  env,
		TestedCapabilities: a.Capabilities,
		HTTPStatus:         statusCode,
	}, nil
}

// GetTracking queries MSC DCSA Track & Trace API.
func (a *MscAdapter) GetTracking(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.TrackingRequest) (*domain.NormalizedTrackingResult, error) {
	if err := a.CheckCapability(domain.CapTracking); err != nil {
		return nil, err
	}

	client := a.getHTTPClient(creds, env)
	headers := a.buildHeaders(creds)

	queryParams := url.Values{}
	ref := req.ContainerNumber
	if req.ContainerNumber != "" {
		queryParams.Set("equipmentReference", req.ContainerNumber)
	} else if req.BookingNumber != "" {
		queryParams.Set("carrierBookingReference", req.BookingNumber)
		ref = req.BookingNumber
	} else if req.MBLNumber != "" {
		queryParams.Set("transportDocumentReference", req.MBLNumber)
		ref = req.MBLNumber
	} else {
		return nil, domain.NewIntegrationError(
			a.Code,
			"GET_TRACKING",
			domain.ErrCodeInvalidRequest,
			"Tracking identifier required for MSC",
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

	var events []domain.DCSAEvent
	if err := json.Unmarshal(respBody, &events); err != nil {
		return nil, domain.NewIntegrationError(
			a.Code,
			"GET_TRACKING",
			domain.ErrCodeInternal,
			"Failed to parse MSC Track & Trace events",
			http.StatusInternalServerError,
			false,
			err,
		)
	}

	return NormalizeDCSATrackingEvents(a.CarrierSCAC, ref, events), nil
}
