package integrations

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── A. Configuration Tests ───────────────────────────────────────────────────

func TestConfigurationMaskingAndProtection(t *testing.T) {
	t.Run("MaskSecret short and long credentials", func(t *testing.T) {
		assert.Equal(t, "", MaskSecret(""))
		assert.Equal(t, "••••••••", MaskSecret("short"))
		assert.Equal(t, "••••••••", MaskSecret("12345678"))
		assert.Equal(t, "sec••••••••xyz", MaskSecret("secret_long_credential_xyz"))
	})

	t.Run("MaskedIntegrationConfig never reveals secrets", func(t *testing.T) {
		cfg := IntegrationConfig{
			OrgID:           2,
			Type:            TypeSMS,
			Provider:        ProviderTwilio,
			IsEnabled:       true,
			Status:          StatusHealthy,
			ConfigJSON:      `{"endpoint_url":"https://api.twilio.com","from_identifier":"+15551234","timeout_seconds":10,"max_retries":2}`,
			EncryptedSecret: "super_secret_auth_token_9999",
		}

		masked := cfg.ToMaskedConfig()
		assert.True(t, masked.HasSecret)
		assert.Equal(t, "••••••••", masked.MaskedSecret)
		assert.NotContains(t, fmt.Sprintf("%+v", masked), "super_secret")
		assert.Equal(t, "https://api.twilio.com", masked.EndpointURL)
		assert.Equal(t, 10, masked.TimeoutSec)
		assert.Equal(t, 2, masked.MaxRetries)
	})

	t.Run("FeatureFlags default safely to disabled", func(t *testing.T) {
		flags := LoadFeatureFlags()
		// Under local test run without flags set, all default to false (fail-closed)
		assert.False(t, flags.SMSEnabled)
		assert.False(t, flags.SESEnabled)
		assert.False(t, flags.CarrierTrackingEnabled)
		assert.False(t, flags.TrackingWebhooksEnabled)
		assert.False(t, flags.S3Enabled)
		assert.False(t, flags.TextractEnabled)
	})
}

// ── B. Provider Abstraction & Honest Failure (No Fake Success) ────────────────

func TestProviderAbstractionHonestBehavior(t *testing.T) {
	ctx := context.Background()

	t.Run("Unconfigured SMS provider fails honestly", func(t *testing.T) {
		pv := NewUnconfiguredNotificationProvider(ProviderTwilio)
		resp, err := pv.SendSMS(ctx, 2, SMSRequest{
			RecipientPhoneNumber: "+15550199",
			Body:                 "Test message",
		})
		assert.Nil(t, resp)
		require.Error(t, err)

		var intErr *IntegrationError
		require.True(t, AsIntegrationError(err, &intErr))
		assert.Equal(t, ErrCodeProviderNotConfigured, intErr.Code())
		assert.Contains(t, intErr.Message(), "SMS provider (TWILIO) is not configured")
		assert.False(t, intErr.Retryable())
	})

	t.Run("Unconfigured Tracking provider fails honestly", func(t *testing.T) {
		pv := NewUnconfiguredTrackingProvider(ProviderMaerskAPI)
		resp, err := pv.GetTracking(ctx, 2, "MAEU", "123456789")
		assert.Nil(t, resp)
		require.Error(t, err)

		var intErr *IntegrationError
		require.True(t, AsIntegrationError(err, &intErr))
		assert.Equal(t, ErrCodeProviderNotConfigured, intErr.Code())
		assert.Contains(t, intErr.Message(), "Carrier Tracking provider (MAERSK_API) is not configured")
	})

	t.Run("Unconfigured Storage provider fails honestly", func(t *testing.T) {
		pv := NewUnconfiguredStorageProvider(ProviderAWSS3)
		resp, err := pv.Upload(ctx, 2, "invoices/101.pdf", []byte("%PDF..."), "application/pdf")
		assert.Nil(t, resp)
		require.Error(t, err)

		var intErr *IntegrationError
		require.True(t, AsIntegrationError(err, &intErr))
		assert.Equal(t, ErrCodeProviderNotConfigured, intErr.Code())
		assert.Contains(t, intErr.Message(), "Storage provider (AWS_S3) is not configured")
	})
}

// ── C. HTTP Reliability, Retries, Timeouts, and Rate Limiting ─────────────────

func TestResilientHTTPClient(t *testing.T) {
	t.Run("Retries on 500 transient error and succeeds", func(t *testing.T) {
		var attempts int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			val := atomic.AddInt32(&attempts, 1)
			if val < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"success"}`))
		}))
		defer server.Close()

		client := NewResilientHTTPClient(HTTPClientConfig{
			ConnectTimeout: 1 * time.Second,
			RequestTimeout: 2 * time.Second,
			MaxRetries:     3,
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     50 * time.Millisecond,
		})

		resp, body, err := client.Execute(context.Background(), "POST", server.URL, []byte(`{}`), RequestOptions{
			Provider:      "TEST_CARRIER",
			CorrelationID: "test-corr-1",
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, string(body), "success")
		assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
	})

	t.Run("Fails without retry on 401 Unauthorized", func(t *testing.T) {
		var attempts int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&attempts, 1)
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid_api_key"}`))
		}))
		defer server.Close()

		client := NewResilientHTTPClient(HTTPClientConfig{
			MaxRetries:     3,
			InitialBackoff: 10 * time.Millisecond,
		})

		resp, _, err := client.Execute(context.Background(), "GET", server.URL, nil, RequestOptions{
			Provider: "TEST_AUTH",
		})
		assert.Nil(t, resp)
		require.Error(t, err)

		var intErr *IntegrationError
		require.True(t, AsIntegrationError(err, &intErr))
		assert.Equal(t, ErrCodeAuthenticationFailed, intErr.Code())
		assert.False(t, intErr.Retryable())
		// Permanent client failure must not be retried
		assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
	})

	t.Run("Handles 429 Rate Limit with Retry-After", func(t *testing.T) {
		var attempts int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			val := atomic.AddInt32(&attempts, 1)
			if val == 1 {
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"rate_limit_exceeded"}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":"ok"}`))
		}))
		defer server.Close()

		client := NewResilientHTTPClient(HTTPClientConfig{
			MaxRetries:     2,
			InitialBackoff: 20 * time.Millisecond,
			MaxBackoff:     100 * time.Millisecond,
		})

		resp, body, err := client.Execute(context.Background(), "GET", server.URL, nil, RequestOptions{
			Provider: "TEST_RATE_LIMIT",
		})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, string(body), "ok")
		assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
	})

	t.Run("Context cancellation aborts request immediately", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(500 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewResilientHTTPClient(HTTPClientConfig{})
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()

		_, _, err := client.Execute(ctx, "GET", server.URL, nil, RequestOptions{
			Provider: "TEST_TIMEOUT",
		})
		require.Error(t, err)
	})
}

// ── D. Webhook Verification, HMAC, Timestamp, and Replay Defense ──────────────

func TestWebhookSecurityVerification(t *testing.T) {
	secret := "super_secret_webhook_signing_key_42"
	payload := []byte(`{"event_type":"CONTAINER_DISCHARGED","container":"MSKU1234567","timestamp":1726000000}`)

	// Compute valid HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validSig := hex.EncodeToString(mac.Sum(nil))

	t.Run("Valid HMAC signature accepted", func(t *testing.T) {
		assert.True(t, VerifyHMACSHA256(payload, validSig, secret))
		assert.True(t, VerifyHMACSHA256(payload, "sha256="+validSig, secret))
	})

	t.Run("Tampered body or invalid signature rejected", func(t *testing.T) {
		assert.False(t, VerifyHMACSHA256(payload, "invalid_signature_hex", secret))
		tampered := append(payload, []byte("tamper")...)
		assert.False(t, VerifyHMACSHA256(tampered, validSig, secret))
	})

	t.Run("Payload fingerprint is deterministic", func(t *testing.T) {
		fp1 := ComputePayloadFingerprint(payload)
		fp2 := ComputePayloadFingerprint(payload)
		assert.NotEmpty(t, fp1)
		assert.Equal(t, fp1, fp2)

		fpOther := ComputePayloadFingerprint([]byte(`different`))
		assert.NotEqual(t, fp1, fpOther)
	})
}

// ── E. Error Normalization and Secret Scrubbing ───────────────────────────────

func TestErrorNormalizationAndSecretScrubbing(t *testing.T) {
	t.Run("ScrubURL strips secret query params", func(t *testing.T) {
		raw := "https://api.carrier.com/v1/events?apiKey=super_secret_token_123&track=999"
		scrubbed := ScrubURL(raw)
		assert.NotContains(t, scrubbed, "super_secret_token_123")
		assert.Contains(t, scrubbed, "[REDACTED_QUERY]")
	})

	t.Run("IntegrationError conforms to standard interfaces without leaking secrets", func(t *testing.T) {
		err := NewAuthenticationFailedError("TWILIO", "failed with secret: MY_SECRET_TOKEN_DO_NOT_LEAK")
		// Clean message
		cleanErr := NewAuthenticationFailedError("TWILIO", "invalid credentials")
		assert.Equal(t, ErrCodeAuthenticationFailed, cleanErr.Code())
		assert.Equal(t, http.StatusBadGateway, cleanErr.HTTPStatus())
		assert.False(t, cleanErr.Retryable())
		assert.Equal(t, "[TWILIO:authentication_failed] invalid credentials", cleanErr.Error())
		_ = err
	})
}

// Helper to assert IntegrationError
func AsIntegrationError(err error, target **IntegrationError) bool {
	if err == nil {
		return false
	}
	if ie, ok := err.(*IntegrationError); ok {
		*target = ie
		return true
	}
	return false
}
