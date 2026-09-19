package integrations

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// HTTPClientConfig defines reliability and timeout settings for external communication.
type HTTPClientConfig struct {
	ConnectTimeout time.Duration
	RequestTimeout time.Duration
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

// DefaultHTTPClientConfig supplies battle-tested defaults for external providers.
func DefaultHTTPClientConfig() HTTPClientConfig {
	return HTTPClientConfig{
		ConnectTimeout: 5 * time.Second,
		RequestTimeout: 15 * time.Second,
		MaxRetries:     3,
		InitialBackoff: 200 * time.Millisecond,
		MaxBackoff:     3 * time.Second,
	}
}

// ResilientHTTPClient wraps standard http.Client with bounded exponential retries, rate-limiting, and error normalization.
type ResilientHTTPClient struct {
	client *http.Client
	config HTTPClientConfig
}

// NewResilientHTTPClient creates a resilient HTTP client.
func NewResilientHTTPClient(cfg HTTPClientConfig) *ResilientHTTPClient {
	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = 5 * time.Second
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 15 * time.Second
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}
	if cfg.InitialBackoff <= 0 {
		cfg.InitialBackoff = 200 * time.Millisecond
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = 3 * time.Second
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   cfg.ConnectTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	return &ResilientHTTPClient{
		client: &http.Client{
			Transport: transport,
			Timeout:   cfg.RequestTimeout,
		},
		config: cfg,
	}
}

// RequestOptions configures an individual external HTTP invocation.
type RequestOptions struct {
	Provider       string
	CorrelationID  string
	IdempotencyKey string
	AuthHeader     string // e.g. "Bearer ..." or "Basic ..."
	Headers        map[string]string
}

// Execute performs an external HTTP call with bounded retries, rate-limit backoff, and error normalization.
func (c *ResilientHTTPClient) Execute(ctx context.Context, method, url string, body []byte, opts RequestOptions) (*http.Response, []byte, error) {
	var lastErr error
	var resp *http.Response
	var respBody []byte

	maxAttempts := 1 + c.config.MaxRetries
	backoff := c.config.InitialBackoff

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check for parent context cancellation
		if ctx.Err() != nil {
			return nil, nil, &IntegrationError{
				CodeValue:       ErrCodeTimeout,
				MessageValue:    "request cancelled by context: " + ctx.Err().Error(),
				HTTPStatusValue: http.StatusGatewayTimeout,
				ProviderValue:   opts.Provider,
				CorrelationID:   opts.CorrelationID,
			}
		}

		var bodyReader io.Reader
		if len(body) > 0 {
			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, nil, NewInvalidRequestError(opts.Provider, err.Error())
		}

		// Inject tracing and headers
		if opts.CorrelationID != "" {
			req.Header.Set("X-Correlation-ID", opts.CorrelationID)
		}
		if opts.IdempotencyKey != "" {
			req.Header.Set("Idempotency-Key", opts.IdempotencyKey)
		}
		if opts.AuthHeader != "" {
			req.Header.Set("Authorization", opts.AuthHeader)
		}
		for k, v := range opts.Headers {
			req.Header.Set(k, v)
		}

		resp, err = c.client.Do(req)
		if err != nil {
			lastErr = err
			// Check if network/timeout error is retryable
			if attempt < maxAttempts && isRetryableNetworkError(err) {
				c.sleepWithJitter(ctx, backoff)
				backoff = c.nextBackoff(backoff)
				continue
			}
			return nil, nil, NewConnectionFailedError(opts.Provider, ScrubURL(url)+": "+err.Error())
		}

		// Read response body
		respBody, err = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = err
			if attempt < maxAttempts {
				c.sleepWithJitter(ctx, backoff)
				backoff = c.nextBackoff(backoff)
				continue
			}
			return nil, nil, NewConnectionFailedError(opts.Provider, "failed to read response body: "+err.Error())
		}

		// Handle Rate Limiting (429)
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfterSec := parseRetryAfter(resp.Header.Get("Retry-After"))
			if attempt < maxAttempts {
				sleepDuration := backoff
				if retryAfterSec > 0 {
					sleepDuration = time.Duration(retryAfterSec) * time.Second
					if sleepDuration > c.config.MaxBackoff {
						sleepDuration = c.config.MaxBackoff
					}
				}
				c.sleepWithJitter(ctx, sleepDuration)
				backoff = c.nextBackoff(backoff)
				continue
			}
			return nil, respBody, NewRateLimitedError(opts.Provider, retryAfterSec)
		}

		// Handle Transient Server Errors (500, 502, 503, 504)
		if resp.StatusCode >= 500 && resp.StatusCode <= 504 {
			if attempt < maxAttempts {
				c.sleepWithJitter(ctx, backoff)
				backoff = c.nextBackoff(backoff)
				continue
			}
			return nil, respBody, NewProviderUnavailableError(opts.Provider, fmt.Sprintf("received HTTP %d", resp.StatusCode))
		}

		// Permanent Client Errors (400, 401, 403, 404, 422) - DO NOT RETRY
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, respBody, NewAuthenticationFailedError(opts.Provider, "invalid credentials or rejected authorization")
		}
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return nil, respBody, NewInvalidRequestError(opts.Provider, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)))
		}

		// Success (2xx)
		return resp, respBody, nil
	}

	if lastErr != nil {
		return nil, nil, NewConnectionFailedError(opts.Provider, lastErr.Error())
	}
	return nil, nil, NewProviderUnavailableError(opts.Provider, "maximum retries exceeded")
}

func (c *ResilientHTTPClient) nextBackoff(current time.Duration) time.Duration {
	next := current * 2
	if next > c.config.MaxBackoff {
		return c.config.MaxBackoff
	}
	return next
}

func (c *ResilientHTTPClient) sleepWithJitter(ctx context.Context, duration time.Duration) {
	// Add up to 25% random jitter
	jitterMax := int64(duration / 4)
	if jitterMax > 0 {
		if n, err := rand.Int(rand.Reader, big.NewInt(jitterMax)); err == nil {
			duration += time.Duration(n.Int64())
		}
	}
	select {
	case <-time.After(duration):
	case <-ctx.Done():
	}
}

func isRetryableNetworkError(err error) bool {
	if err == nil {
		return false
	}
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout()
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "connection reset") ||
		strings.Contains(s, "connection refused") ||
		strings.Contains(s, "broken pipe") ||
		strings.Contains(s, "deadline exceeded") ||
		strings.Contains(s, "timeout")
}

func parseRetryAfter(val string) int {
	val = strings.TrimSpace(val)
	if val == "" {
		return 0
	}
	if sec, err := strconv.Atoi(val); err == nil && sec > 0 {
		return sec
	}
	return 0
}

// ScrubURL removes sensitive query parameters (e.g. key, token, secret) from URLs before logging.
func ScrubURL(rawURL string) string {
	parts := strings.SplitN(rawURL, "?", 2)
	if len(parts) == 1 {
		return rawURL
	}
	return parts[0] + "?[REDACTED_QUERY]"
}
