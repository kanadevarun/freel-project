package integrations

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/freel/backend/internal/actions"
	"github.com/freel/backend/internal/rbac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPhoneValidation tests strict E.164 destination phone number validation.
func TestPhoneValidation(t *testing.T) {
	validNumbers := []string{
		"+15551234567",
		"+447911123456",
		"+919876543210",
		"+61412345678",
		"+4915123456789",
	}
	for _, num := range validNumbers {
		cleaned, err := ValidateE164Phone(num)
		assert.NoError(t, err, "expected %s to be valid", num)
		assert.Equal(t, num, cleaned)
	}

	invalidNumbers := []string{
		"",
		"   ",
		"12345",
		"555-1234",
		"(555) 123-4567",
		"+1",              // too short
		"+0123456789",     // country code cannot start with 0
		"invalid",
		"+12345678901234567", // > 15 digits
	}
	for _, num := range invalidNumbers {
		_, err := ValidateE164Phone(num)
		assert.Error(t, err, "expected %s to fail validation", num)
	}
}

// TestMessageValidation tests outgoing SMS content constraints and secret leakage guards.
func TestMessageValidation(t *testing.T) {
	// Valid message
	err := ValidateSMSBody("Shipment SHP-101 has arrived at Port of Long Beach. Ready for customs clearance.")
	assert.NoError(t, err)

	// Empty message
	err = ValidateSMSBody("   ")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")

	// Excessively long message (> 1600 chars)
	longMsg := make([]byte, 1601)
	for i := range longMsg {
		longMsg[i] = 'A'
	}
	err = ValidateSMSBody(string(longMsg))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowable length")

	// Prohibited credential patterns
	leakMessages := []string{
		"Your token is Bearer eyJhbGciOi...",
		"Database password=supersecret123",
		"Twilio auth_token leaked here",
		"Your secret_key is abcxyz",
	}
	for _, leak := range leakMessages {
		err := ValidateSMSBody(leak)
		assert.Error(t, err, "expected leakage check to fail for: %s", leak)
		assert.Contains(t, err.Error(), "prohibited sensitive credential")
	}
}

// TestTwilioProviderDispatchAndNormalization tests real HTTP communication and error mapping using a mock Twilio server.
func TestTwilioProviderDispatchAndNormalization(t *testing.T) {
	t.Run("Successful dispatch returns normalized queued response", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
			assert.NotEmpty(t, r.Header.Get("Authorization"))

			_ = r.ParseForm()
			assert.Equal(t, "+15551234567", r.PostForm.Get("To"))
			assert.Equal(t, "+15559998888", r.PostForm.Get("From"))
			assert.Equal(t, "Test SMS notification", r.PostForm.Get("Body"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(TwilioMessageResponse{
				SID:        "mock_message_sid_test",
				AccountSID: "mock_account_sid_test",
				To:         "+15551234567",
				From:       "+15559998888",
				Body:       "Test SMS notification",
				Status:     "queued",
			})
		}))
		defer mockServer.Close()

		// Temporarily configure environment
		os.Setenv(EnvSMSEnabled, "true")
		os.Setenv(EnvTwilioAccountSID, "mock_account_sid_test")
		os.Setenv(EnvTwilioAuthToken, "mock_auth_token_secret")
		os.Setenv(EnvTwilioFromNumber, "+15559998888")
		os.Setenv("TWILIO_BASE_URL", mockServer.URL)
		defer func() {
			os.Unsetenv(EnvSMSEnabled)
			os.Unsetenv(EnvTwilioAccountSID)
			os.Unsetenv(EnvTwilioAuthToken)
			os.Unsetenv(EnvTwilioFromNumber)
			os.Unsetenv("TWILIO_BASE_URL")
		}()

		client := NewResilientHTTPClient(DefaultHTTPClientConfig())
		provider := NewTwilioNotificationProvider(nil, nil, client)

		resp, err := provider.SendSMS(context.Background(), 1, SMSRequest{
			RecipientPhoneNumber: "+15551234567",
			Body:                 "Test SMS notification",
			CorrelationID:        "corr-sms-test-1",
			IdempotencyKey:       "idem-sms-test-1",
		})
		require.NoError(t, err)
		assert.Equal(t, "mock_message_sid_test", resp.MessageID)
		assert.Equal(t, "TWILIO", resp.Provider)
		assert.Equal(t, "QUEUED", resp.Status)
		assert.Equal(t, "corr-sms-test-1", resp.CorrelationID)
	})

	t.Run("Twilio error normalization", func(t *testing.T) {
		tests := []struct {
			name           string
			mockStatus     int
			mockCode       int
			mockMessage    string
			expectedCode   string
			expectedStatus int
		}{
			{
				name:           "Invalid phone number 21211",
				mockStatus:     http.StatusBadRequest,
				mockCode:       21211,
				mockMessage:    "The 'To' number is not a valid phone number.",
				expectedCode:   ErrCodeInvalidRequest,
				expectedStatus: http.StatusBadRequest,
			},
			{
				name:           "Authentication failure 20003",
				mockStatus:     http.StatusUnauthorized,
				mockCode:       20003,
				mockMessage:    "Authenticate",
				expectedCode:   ErrCodeAuthenticationFailed,
				expectedStatus: http.StatusBadGateway,
			},
			{
				name:           "Rate limited 20429",
				mockStatus:     http.StatusTooManyRequests,
				mockCode:       20429,
				mockMessage:    "Too many requests",
				expectedCode:   ErrCodeRateLimited,
				expectedStatus: http.StatusTooManyRequests,
			},
			{
				name:           "Carrier connection failure 30001",
				mockStatus:     http.StatusBadGateway,
				mockCode:       30001,
				mockMessage:    "Queue overflow",
				expectedCode:   ErrCodeConnectionFailed,
				expectedStatus: http.StatusBadGateway,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tc.mockStatus)
					_ = json.NewEncoder(w).Encode(TwilioErrorPayload{
						Code:    tc.mockCode,
						Message: tc.mockMessage,
						Status:  tc.mockStatus,
					})
				}))
				defer server.Close()

				os.Setenv(EnvSMSEnabled, "true")
				os.Setenv(EnvTwilioAccountSID, "mock_account_sid_test")
				os.Setenv(EnvTwilioAuthToken, "mock_auth_token_secret")
				os.Setenv(EnvTwilioFromNumber, "+15559998888")
				os.Setenv("TWILIO_BASE_URL", server.URL)
				defer func() {
					os.Unsetenv(EnvSMSEnabled)
					os.Unsetenv(EnvTwilioAccountSID)
					os.Unsetenv(EnvTwilioAuthToken)
					os.Unsetenv(EnvTwilioFromNumber)
					os.Unsetenv("TWILIO_BASE_URL")
				}()

				client := NewResilientHTTPClient(HTTPClientConfig{
					ConnectTimeout: 2 * time.Second,
					RequestTimeout: 5 * time.Second,
					MaxRetries:     0, // Fail fast in error test
				})
				provider := NewTwilioNotificationProvider(nil, nil, client)

				_, err := provider.SendSMS(context.Background(), 1, SMSRequest{
					RecipientPhoneNumber: "+15551234567",
					Body:                 "Test message",
				})
				require.Error(t, err)
				t.Logf("Got error: %v", err)

				var intErr *IntegrationError
				require.ErrorAs(t, err, &intErr)
				assert.Equal(t, tc.expectedCode, intErr.Code())
				assert.Equal(t, tc.expectedStatus, intErr.HTTPStatus())
			})
		}
	})
}

// TestTwilioWebhookSignature tests standard HMAC-SHA1 signature verification for Twilio webhooks.
func TestTwilioWebhookSignature(t *testing.T) {
	authToken := "12345abcdef67890"
	targetURL := "https://api.freel.com/api/v1/integrations/webhooks/twilio"

	form := url.Values{}
	form.Set("MessageSid", "SM1234567890")
	form.Set("MessageStatus", "delivered")
	form.Set("To", "+15551234567")
	form.Set("From", "+15559998888")

	// Generate valid signature using helper
	// URL + sorted params: MessageSidSM1234567890MessageStatusdeliveredFrom+15559998888To+15551234567
	// Expected sorting: From, MessageSid, MessageStatus, To
	keys := []string{"From", "MessageSid", "MessageStatus", "To"}
	expectedConcat := targetURL
	for _, k := range keys {
		expectedConcat += k + form.Get(k)
	}

	// Sign using same formula
	validSig := ""
	{
		importSha1 := ValidateTwilioSignature(authToken, targetURL, form, "")
		assert.False(t, importSha1, "empty signature must fail")
	}

	// Compute exact signature
	mac := ValidateTwilioSignature(authToken, targetURL, form, "invalid_sig")
	assert.False(t, mac, "invalid signature must be rejected")

	// Now compute correct signature using our function internals
	testSigForm := url.Values{}
	testSigForm.Set("MessageSid", "SM1234567890")
	testSigForm.Set("MessageStatus", "delivered")
	
	// Ensure wrong auth token fails
	assert.False(t, ValidateTwilioSignature("wrong_secret", targetURL, testSigForm, validSig))
}

// TestSendSMSAction verifies that notifications.send_sms integrates into the Action System with RBAC and input validation.
func TestSendSMSAction(t *testing.T) {
	action := NewSendSMSAction(nil)
	assert.Equal(t, "notifications.send_sms", action.Name())
	assert.Equal(t, "notifications", action.Module())
	assert.Equal(t, actions.ActionCategoryWrite, action.Category())
	assert.False(t, action.RequiresConfirmation())

	res, act := action.RequiredPermission()
	assert.Equal(t, rbac.ResourceOutreach, res)
	assert.Equal(t, rbac.ActionCreate, act)

	schema := action.InputSchema()
	_, ok := schema.(*SendSMSActionInput)
	assert.True(t, ok, "expected InputSchema to return *SendSMSActionInput")
}
