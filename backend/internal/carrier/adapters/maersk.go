package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/freel/backend/internal/carrier/adapters/httpclient"
	"github.com/freel/backend/internal/carrier/adapters/token"
	"github.com/freel/backend/internal/carrier/domain"
)

// MaerskAdapter integrates A.P. Moller – Maersk Line API (SCAC: MAEU).
type MaerskAdapter struct {
	BaseAdapter
	tokenManager *token.TokenManager
}

// NewMaerskAdapter constructs the official Maersk carrier adapter.
func NewMaerskAdapter() *MaerskAdapter {
	return &MaerskAdapter{
		BaseAdapter: BaseAdapter{
			Code:        "MAERSK",
			CarrierSCAC: "MAEU",
			Capabilities: []domain.Capability{
				domain.CapTracking,
				domain.CapRates,
				domain.CapContractRates,
				domain.CapSpotRates,
				domain.CapBooking,
				domain.CapDocuments,
			},
		},
		tokenManager: token.GetDefaultTokenManager(),
	}
}

func (a *MaerskAdapter) getBaseURL(creds domain.DecryptedCredentials, env domain.Environment) string {
	if creds.BaseURL != "" {
		return creds.BaseURL
	}
	if env == domain.EnvSandbox {
		return "https://api.maersk.com"
	}
	return "https://api.maersk.com"
}

func (a *MaerskAdapter) getHTTPClient(creds domain.DecryptedCredentials, env domain.Environment) *httpclient.CarrierHTTPClient {
	return httpclient.NewCarrierHTTPClient(httpclient.HTTPClientConfig{
		BaseURL:        a.getBaseURL(creds, env),
		Timeout:        15 * time.Second,
		MaxRetries:     2,
		InitialBackoff: 200 * time.Millisecond,
		ProviderCode:   a.Code,
	})
}

// buildHeaders sets authentication headers conforming to Maersk API specifications.
func (a *MaerskAdapter) buildHeaders(creds domain.DecryptedCredentials) map[string]string {
	headers := map[string]string{
		"Accept": "application/json",
	}
	if creds.APIKey != "" {
		headers["Consumer-Key"] = creds.APIKey
		headers["X-API-Key"] = creds.APIKey
	}
	return headers
}

// TestConnection verifies connectivity and authentication with Maersk API gateway.
func (a *MaerskAdapter) TestConnection(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment) (*domain.TestConnectionResult, error) {
	start := time.Now()

	if creds.APIKey == "" && creds.ClientID == "" {
		return &domain.TestConnectionResult{
			Success:            false,
			Message:            "Missing Maersk API Key or Client ID. Please provide valid developer credentials.",
			LatencyMs:          time.Since(start).Milliseconds(),
			TestedEnvironment:  env,
			TestedCapabilities: a.Capabilities,
			ErrorCode:          domain.ErrCodeInvalidConfig,
			HTTPStatus:         http.StatusBadRequest,
		}, nil
	}

	client := a.getHTTPClient(creds, env)
	headers := a.buildHeaders(creds)

	// Ping Track & Trace endpoint or OAuth token endpoint
	testEndpoint := "/track-and-trace/v2/events?limit=1"
	if creds.ClientID != "" && creds.ClientSecret != "" {
		// Attempt OAuth token acquisition
		_, err := a.tokenManager.GetToken(ctx, token.TokenRequest{
			TokenURL:     a.getBaseURL(creds, env) + "/oauth2/access_token",
			ClientID:     creds.ClientID,
			ClientSecret: creds.ClientSecret,
			ProviderCode: a.Code,
			Environment:  env,
		})
		latency := time.Since(start).Milliseconds()
		if err != nil {
			return &domain.TestConnectionResult{
				Success:            false,
				Message:            fmt.Sprintf("Maersk OAuth2 authentication failed: %v", err),
				LatencyMs:          latency,
				TestedEnvironment:  env,
				TestedCapabilities: a.Capabilities,
				ErrorCode:          domain.ErrCodeAuthFailed,
				HTTPStatus:         http.StatusUnauthorized,
			}, nil
		}
		return &domain.TestConnectionResult{
			Success:            true,
			Message:            "Successfully authenticated with Maersk Developer Portal OAuth2 Gateway.",
			LatencyMs:          latency,
			TestedEnvironment:  env,
			TestedCapabilities: a.Capabilities,
			HTTPStatus:         http.StatusOK,
		}, nil
	}

	// Direct API Key ping
	_, statusCode, err := client.Execute(ctx, "GET", testEndpoint, headers, nil)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		// If 401/403 or network issue
		return &domain.TestConnectionResult{
			Success:            false,
			Message:            fmt.Sprintf("Maersk API connection check returned: %v", err),
			LatencyMs:          latency,
			TestedEnvironment:  env,
			TestedCapabilities: a.Capabilities,
			ErrorCode:          domain.ErrCodeAuthFailed,
			HTTPStatus:         statusCode,
		}, nil
	}

	return &domain.TestConnectionResult{
		Success:            true,
		Message:            "Successfully connected to Maersk API gateway.",
		LatencyMs:          latency,
		TestedEnvironment:  env,
		TestedCapabilities: a.Capabilities,
		HTTPStatus:         statusCode,
	}, nil
}

// GetTracking queries Maersk DCSA Track & Trace API v2.
func (a *MaerskAdapter) GetTracking(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.TrackingRequest) (*domain.NormalizedTrackingResult, error) {
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
			"Container number, booking number, or MBL number must be provided for Maersk tracking",
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
			"Failed to deserialize Maersk Track & Trace events",
			http.StatusInternalServerError,
			false,
			err,
		)
	}

	normalized := NormalizeDCSATrackingEvents(a.CarrierSCAC, ref, events)
	return normalized, nil
}

// GetRates queries Maersk Spot & Tariff pricing engine.
func (a *MaerskAdapter) GetRates(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.RateRequest) ([]domain.NormalizedCarrierRate, error) {
	if err := a.CheckCapability(domain.CapRates); err != nil {
		return nil, err
	}

	client := a.getHTTPClient(creds, env)
	headers := a.buildHeaders(creds)

	bodyBytes, _ := json.Marshal(req)
	respBody, _, err := client.Execute(ctx, "POST", "/rates/v1/spot-rates", headers, bodyBytes)
	if err != nil {
		return nil, err
	}

	return NormalizeMaerskRates(a.CarrierSCAC, respBody)
}

// GetContractRates queries Maersk contracted commercial rates.
func (a *MaerskAdapter) GetContractRates(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.ContractRateRequest) ([]domain.NormalizedCarrierRate, error) {
	if err := a.CheckCapability(domain.CapContractRates); err != nil {
		return nil, err
	}

	client := a.getHTTPClient(creds, env)
	headers := a.buildHeaders(creds)

	bodyBytes, _ := json.Marshal(req)
	respBody, _, err := client.Execute(ctx, "POST", "/rates/v1/contract-rates", headers, bodyBytes)
	if err != nil {
		return nil, err
	}

	return NormalizeMaerskRates(a.CarrierSCAC, respBody)
}

// GetSpotRates queries Maersk instant spot pricing.
func (a *MaerskAdapter) GetSpotRates(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.SpotRateRequest) ([]domain.NormalizedCarrierRate, error) {
	if err := a.CheckCapability(domain.CapSpotRates); err != nil {
		return nil, err
	}

	client := a.getHTTPClient(creds, env)
	headers := a.buildHeaders(creds)

	bodyBytes, _ := json.Marshal(req)
	respBody, _, err := client.Execute(ctx, "POST", "/rates/v1/spot-rates", headers, bodyBytes)
	if err != nil {
		return nil, err
	}

	return NormalizeMaerskRates(a.CarrierSCAC, respBody)
}

// CreateBooking creates a new container space booking with Maersk.
func (a *MaerskAdapter) CreateBooking(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.BookingRequest) (*domain.NormalizedBookingResult, error) {
	if err := a.CheckCapability(domain.CapBooking); err != nil {
		return nil, err
	}

	client := a.getHTTPClient(creds, env)
	headers := a.buildHeaders(creds)

	bodyBytes, _ := json.Marshal(req)
	respBody, _, err := client.Execute(ctx, "POST", "/booking/v1/bookings", headers, bodyBytes)
	if err != nil {
		return nil, err
	}

	var res domain.NormalizedBookingResult
	if err := json.Unmarshal(respBody, &res); err != nil {
		return nil, domain.NewIntegrationError(
			a.Code,
			"CREATE_BOOKING",
			domain.ErrCodeInternal,
			"Failed to parse Maersk booking confirmation",
			http.StatusInternalServerError,
			false,
			err,
		)
	}
	res.CarrierSCAC = a.CarrierSCAC
	return &res, nil
}

// GetBooking fetches status of an existing Maersk booking.
func (a *MaerskAdapter) GetBooking(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, bookingRef string) (*domain.NormalizedBookingResult, error) {
	if err := a.CheckCapability(domain.CapBooking); err != nil {
		return nil, err
	}

	client := a.getHTTPClient(creds, env)
	headers := a.buildHeaders(creds)

	endpoint := fmt.Sprintf("/booking/v1/bookings/%s", bookingRef)
	respBody, _, err := client.Execute(ctx, "GET", endpoint, headers, nil)
	if err != nil {
		return nil, err
	}

	var res domain.NormalizedBookingResult
	if err := json.Unmarshal(respBody, &res); err != nil {
		return nil, domain.NewIntegrationError(
			a.Code,
			"GET_BOOKING",
			domain.ErrCodeInternal,
			"Failed to parse Maersk booking status",
			http.StatusInternalServerError,
			false,
			err,
		)
	}
	res.CarrierSCAC = a.CarrierSCAC
	return &res, nil
}

// GetDocuments retrieves transport documents from Maersk API.
func (a *MaerskAdapter) GetDocuments(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.DocumentRequest) ([]domain.NormalizedDocumentResult, error) {
	if err := a.CheckCapability(domain.CapDocuments); err != nil {
		return nil, err
	}

	client := a.getHTTPClient(creds, env)
	headers := a.buildHeaders(creds)

	endpoint := fmt.Sprintf("/documents/v1/bookings/%s", req.BookingNumber)
	respBody, _, err := client.Execute(ctx, "GET", endpoint, headers, nil)
	if err != nil {
		return nil, err
	}

	var docs []domain.NormalizedDocumentResult
	if err := json.Unmarshal(respBody, &docs); err != nil {
		return nil, domain.NewIntegrationError(
			a.Code,
			"GET_DOCUMENTS",
			domain.ErrCodeInternal,
			"Failed to parse Maersk documents response",
			http.StatusInternalServerError,
			false,
			err,
		)
	}
	return docs, nil
}
