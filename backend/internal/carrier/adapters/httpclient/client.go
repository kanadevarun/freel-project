package httpclient

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/freel/backend/internal/carrier/domain"
)

// HTTPClientConfig defines client parameters for a carrier adapter.
type HTTPClientConfig struct {
	BaseURL        string
	Timeout        time.Duration
	MaxRetries     int
	InitialBackoff time.Duration
	ProviderCode   string
}

// CarrierHTTPClient handles resilient, redacted HTTP communication with shipping line APIs.
type CarrierHTTPClient struct {
	client       *http.Client
	baseURL      string
	maxRetries   int
	backoff      time.Duration
	providerCode string
}

// NewCarrierHTTPClient constructs a new CarrierHTTPClient with default safety bounds.
func NewCarrierHTTPClient(cfg HTTPClientConfig) *CarrierHTTPClient {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	retries := cfg.MaxRetries
	if retries <= 0 {
		retries = 2
	}
	backoff := cfg.InitialBackoff
	if backoff <= 0 {
		backoff = 250 * time.Millisecond
	}

	return &CarrierHTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL:      strings.TrimRight(cfg.BaseURL, "/"),
		maxRetries:   retries,
		backoff:      backoff,
		providerCode: cfg.ProviderCode,
	}
}

// GenerateCorrelationID creates a unique trace identifier for technical audit trails.
func GenerateCorrelationID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "cr-" + hex.EncodeToString(b)
}

// Execute performs an HTTP request with exponential backoff for retryable errors.
func (c *CarrierHTTPClient) Execute(ctx context.Context, method, path string, headers map[string]string, body []byte) ([]byte, int, error) {
	url := path
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		url = fmt.Sprintf("%s/%s", c.baseURL, strings.TrimLeft(path, "/"))
	}

	corrID := GenerateCorrelationID()
	var lastErr error
	var respStatusCode int
	var respBody []byte

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Backoff before retry
			select {
			case <-ctx.Done():
				return nil, 0, domain.NewIntegrationError(
					c.providerCode,
					method,
					domain.ErrCodeTimeout,
					"Request context timed out before carrier response",
					http.StatusRequestTimeout,
					false,
					ctx.Err(),
				)
			case <-time.After(c.backoff * time.Duration(1<<uint(attempt-1))):
			}
		}

		var bodyReader io.Reader
		if len(body) > 0 {
			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, 0, domain.NewIntegrationError(
				c.providerCode,
				method,
				domain.ErrCodeInvalidRequest,
				"Failed to construct carrier HTTP request",
				http.StatusBadRequest,
				false,
				err,
			)
		}

		// Inject standard headers
		req.Header.Set("Accept", "application/json")
		if len(body) > 0 {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("X-Correlation-ID", corrID)
		req.Header.Set("User-Agent", "LogisticsHQ-Carrier-Engine/2.0")

		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			// Check if context was cancelled or timed out
			if ctx.Err() != nil {
				return nil, 0, domain.NewIntegrationError(
					c.providerCode,
					method,
					domain.ErrCodeTimeout,
					"Carrier API request timed out",
					http.StatusGatewayTimeout,
					true,
					err,
				)
			}
			// Retryable network failure
			continue
		}

		respStatusCode = resp.StatusCode
		respBody, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		// Evaluate response status
		if respStatusCode >= 200 && respStatusCode < 300 {
			return respBody, respStatusCode, nil
		}

		// Handle specific carrier HTTP status codes
		switch respStatusCode {
		case http.StatusUnauthorized:
			return nil, respStatusCode, domain.NewIntegrationError(
				c.providerCode,
				method,
				domain.ErrCodeAuthFailed,
				"Carrier authentication failed. Please verify your API Key or Client Secret.",
				respStatusCode,
				false,
				fmt.Errorf("unauthorized (401)"),
			)
		case http.StatusForbidden:
			return nil, respStatusCode, domain.NewIntegrationError(
				c.providerCode,
				method,
				domain.ErrCodeForbidden,
				"Access to this carrier API resource is forbidden. Please check your account permissions.",
				respStatusCode,
				false,
				fmt.Errorf("forbidden (403)"),
			)
		case http.StatusNotFound:
			return nil, respStatusCode, domain.NewIntegrationError(
				c.providerCode,
				method,
				domain.ErrCodeNotFound,
				"The requested carrier resource or tracking reference was not found.",
				respStatusCode,
				false,
				fmt.Errorf("not found (404)"),
			)
		case http.StatusTooManyRequests:
			lastErr = domain.NewIntegrationError(
				c.providerCode,
				method,
				domain.ErrCodeRateLimited,
				"Carrier API rate limit reached. Please try again later.",
				respStatusCode,
				true,
				fmt.Errorf("rate limited (429)"),
			)
			// Rate limit: retry after backoff
			continue
		case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			lastErr = domain.NewIntegrationError(
				c.providerCode,
				method,
				domain.ErrCodeUnavailable,
				"The carrier's API server is temporarily unavailable or timed out.",
				respStatusCode,
				true,
				fmt.Errorf("server error (%d)", respStatusCode),
			)
			// Gateway error: retry after backoff
			continue
		default:
			// 4xx client errors (non-retryable) or unhandled 5xx
			userMsg := fmt.Sprintf("Carrier returned error response (%d)", respStatusCode)
			if respStatusCode >= 400 && respStatusCode < 500 {
				return nil, respStatusCode, domain.NewIntegrationError(
					c.providerCode,
					method,
					domain.ErrCodeInvalidRequest,
					userMsg,
					respStatusCode,
					false,
					fmt.Errorf("client error (%d)", respStatusCode),
				)
			}
			lastErr = domain.NewIntegrationError(
				c.providerCode,
				method,
				domain.ErrCodeInternal,
				userMsg,
				respStatusCode,
				false,
				fmt.Errorf("carrier server error (%d)", respStatusCode),
			)
		}
	}

	if lastErr != nil {
		return nil, respStatusCode, lastErr
	}
	return nil, respStatusCode, fmt.Errorf("carrier request failed after %d attempts", c.maxRetries)
}
