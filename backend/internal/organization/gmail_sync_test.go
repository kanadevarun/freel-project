package organization_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/freel/backend/internal/common/crypto"
	"github.com/freel/backend/internal/organization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndValidateOAuthState(t *testing.T) {
	encKey := "my-test-secret-encryption-key-32b"
	t.Setenv("MAILBOX_ENCRYPTION_KEY", encKey)

	orgID := int64(42)

	state, err := organization.GenerateOAuthState(orgID)
	require.NoError(t, err)
	assert.NotEmpty(t, state)

	validatedOrgID, err := organization.ValidateOAuthState(state)
	require.NoError(t, err)
	assert.Equal(t, orgID, validatedOrgID)
}

func TestValidateOAuthStateExpired(t *testing.T) {
	encKey := "my-test-secret-encryption-key-32b"
	t.Setenv("MAILBOX_ENCRYPTION_KEY", encKey)

	// State containing an old timestamp
	// org_id = 42, timestamp = 1000 (Unix epoch)
	orgID := int64(42)
	timestamp := int64(1000)

	msg := fmt.Sprintf("%d:%d", orgID, timestamp)
	// Compute HMAC
	mac := hmac.New(sha256.New, []byte(encKey))
	mac.Write([]byte(msg))
	sig := mac.Sum(nil)
	sigBase64 := base64.RawURLEncoding.EncodeToString(sig)
	
	stateStr := fmt.Sprintf("%s:%s", msg, sigBase64)
	expiredState := base64.RawURLEncoding.EncodeToString([]byte(stateStr))

	_, err := organization.ValidateOAuthState(expiredState)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestGmailProviderSyncMailbox(t *testing.T) {
	encKey := "my-test-secret-encryption-key-32b"
	t.Setenv("MAILBOX_ENCRYPTION_KEY", encKey)
	t.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")

	// Set up local HTTP mock server for Google/Gmail REST APIs
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 1. Mock List Messages
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/messages") && !strings.Contains(r.URL.Path, "/messages/") {
			response := map[string]interface{}{
				"messages": []map[string]string{
					{"id": "msg_003"},
					{"id": "msg_002"},
					{"id": "msg_001"},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// 2. Mock Get Message Detail for msg_001 (HTML message)
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/messages/msg_001") {
			response := map[string]interface{}{
				"id":           "msg_001",
				"threadId":     "thread_001",
				"snippet":      "Hello logistics team!",
				"internalDate": "1724580000000",
				"payload": map[string]interface{}{
					"mimeType": "text/html",
					"body": map[string]string{
						"data": base64.URLEncoding.EncodeToString([]byte("<p>Please find attached our shipping RFQ.</p>")),
					},
					"headers": []map[string]string{
						{"name": "Subject", "value": "Shipping RFQ"},
						{"name": "From", "value": "Customer <customer@example.com>"},
						{"name": "To", "value": "sales@logistics.com"},
						{"name": "Message-ID", "value": "<msg1@rfc.com>"},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// 3. Mock Get Message Detail for msg_002 (Plain text message)
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/messages/msg_002") {
			response := map[string]interface{}{
				"id":           "msg_002",
				"threadId":     "thread_001",
				"snippet":      "Follow up on RFQ",
				"internalDate": "1724581000000",
				"payload": map[string]interface{}{
					"mimeType": "text/plain",
					"body": map[string]string{
						"data": base64.URLEncoding.EncodeToString([]byte("Here are the updated dimensions.")),
					},
					"headers": []map[string]string{
						{"name": "Subject", "value": "Re: Shipping Shipping RFQ"},
						{"name": "From", "value": "Customer <customer@example.com>"},
						{"name": "To", "value": "sales@logistics.com"},
						{"name": "Message-ID", "value": "<msg2@rfc.com>"},
						{"name": "In-Reply-To", "value": "<msg1@rfc.com>"},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// 4. Mock Get Message Detail for msg_003 (Self-sent message)
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/messages/msg_003") {
			response := map[string]interface{}{
				"id":           "msg_003",
				"threadId":     "thread_001",
				"snippet":      "Reply from sales",
				"internalDate": "1724582000000",
				"payload": map[string]interface{}{
					"mimeType": "text/plain",
					"body": map[string]string{
						"data": base64.URLEncoding.EncodeToString([]byte("Thanks, we will check it.")),
					},
					"headers": []map[string]string{
						{"name": "Subject", "value": "Re: Shipping RFQ"},
						{"name": "From", "value": "sales@logistics.com"},
						{"name": "To", "value": "customer@example.com"},
						{"name": "Message-ID", "value": "<msg3@rfc.com>"},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Direct Gmail Provider's HTTP requests to mock server
	gp := organization.NewGmailProvider()
	gp.HTTPClient = server.Client()

	// Redirect raw URL calls to mock server by replacing googleapis URLs in provider.go using test URL overrides
	gp.HTTPClient.Transport = &mockURLTranslatorTransport{
		TargetURL: server.URL,
		Base:      http.DefaultTransport,
	}

	accessTokenEnc, err := crypto.Encrypt("mock-access-token", encKey)
	require.NoError(t, err)
	refreshTokenEnc, err := crypto.Encrypt("mock-refresh-token", encKey)
	require.NoError(t, err)

	expiry := time.Now().Add(1 * time.Hour)
	scopes := "gmail.readonly"
	cursor := "msg_000" // start cursor

	mailbox := &organization.ConnectedMailbox{
		ID:                    101,
		OrgID:                 8888,
		Email:                 "sales@logistics.com",
		OwnerName:             "Sales Inbound",
		MailboxType:           "Shared / Team",
		Status:                "CONNECTED",
		Provider:              "GMAIL",
		AccessTokenEncrypted:  &accessTokenEnc,
		RefreshTokenEncrypted: &refreshTokenEnc,
		TokenExpiry:           &expiry,
		OAuthScopes:           &scopes,
		SyncCursor:            &cursor,
	}

	var syncedEmails []organization.SyncEmail
	emailHandler := func(ctx context.Context, email organization.SyncEmail) error {
		syncedEmails = append(syncedEmails, email)
		return nil
	}

	onTokenUpdate := func(ctx context.Context, accessToken string, expiry time.Time) error {
		return nil
	}

	err = gp.SyncMailbox(context.Background(), mailbox, emailHandler, onTokenUpdate)
	require.NoError(t, err)

	// We expect 2 synced emails: msg_001 and msg_002 (msg_003 is skipped since it is sent from sales@logistics.com, which is the mailbox owner!)
	assert.Len(t, syncedEmails, 2)

	// Check msg_001 HTML extraction
	assert.Equal(t, "<msg1@rfc.com>", syncedEmails[0].MessageID)
	assert.Equal(t, "customer@example.com", syncedEmails[0].From)
	assert.Equal(t, "Shipping RFQ", syncedEmails[0].Subject)
	assert.Equal(t, "Please find attached our shipping RFQ.", syncedEmails[0].Body)

	// Check msg_002 Plain Text and reply threading headers
	assert.Equal(t, "<msg2@rfc.com>", syncedEmails[1].MessageID)
	assert.Equal(t, "<msg1@rfc.com>", syncedEmails[1].InReplyTo)
	assert.Equal(t, "Here are the updated dimensions.", syncedEmails[1].Body)
}

type mockURLTranslatorTransport struct {
	TargetURL string
	Base      http.RoundTripper
}

func (t *mockURLTranslatorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Re-route request to local mock server URL
	mockURL, err := url.Parse(t.TargetURL)
	if err != nil {
		return nil, err
	}
	req.URL.Scheme = mockURL.Scheme
	req.URL.Host = mockURL.Host
	return t.Base.RoundTrip(req)
}

func TestGmailProviderSendEmail(t *testing.T) {
	encKey := "my-test-secret-encryption-key-32b"
	t.Setenv("MAILBOX_ENCRYPTION_KEY", encKey)
	t.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")

	var receivedBody []byte
	var receivedHeaders string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "POST" && strings.Contains(r.URL.Path, "/messages/send") {
			var sendReq struct {
				Raw      string `json:"raw"`
				ThreadID string `json:"threadId"`
			}
			err := json.NewDecoder(r.Body).Decode(&sendReq)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			rawBytes, err := base64.RawURLEncoding.DecodeString(sendReq.Raw)
			if err != nil {
				// Try standard URL encoding in case
				rawBytes, err = base64.URLEncoding.DecodeString(sendReq.Raw)
			}
			require.NoError(t, err)

			receivedBody = rawBytes
			receivedHeaders = string(rawBytes)

			response := map[string]string{
				"id":       "sent_msg_999",
				"threadId": "thread_sent_999",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	gp := organization.NewGmailProvider()
	gp.HTTPClient = server.Client()
	gp.HTTPClient.Transport = &mockURLTranslatorTransport{
		TargetURL: server.URL,
		Base:      http.DefaultTransport,
	}

	accessTokenEnc, err := crypto.Encrypt("mock-access-token", encKey)
	require.NoError(t, err)
	refreshTokenEnc, err := crypto.Encrypt("mock-refresh-token", encKey)
	require.NoError(t, err)

	expiry := time.Now().Add(1 * time.Hour)
	scopes := "gmail.send"

	mailbox := &organization.ConnectedMailbox{
		ID:                    102,
		OrgID:                 9999,
		Email:                 "sales@logistics.com",
		OwnerName:             "Sales Team",
		Status:                "CONNECTED",
		Provider:              "GMAIL",
		AccessTokenEncrypted:  &accessTokenEnc,
		RefreshTokenEncrypted: &refreshTokenEnc,
		TokenExpiry:           &expiry,
		OAuthScopes:           &scopes,
	}

	onTokenUpdate := func(ctx context.Context, accessToken string, expiry time.Time) error {
		return nil
	}

	gmailMsgID, rfcMsgID, threadID, err := gp.SendEmail(
		context.Background(),
		mailbox,
		"customer@example.com",
		"Incomplete RFQ details",
		"Please provide the weight.",
		"thread_abc123",
		"<parent@example.com>",
		"<parent@example.com> <older@example.com>",
		onTokenUpdate,
	)

	require.NoError(t, err)
	assert.NotEmpty(t, gmailMsgID)
	assert.NotEmpty(t, rfcMsgID)
	assert.Equal(t, "thread_sent_999", threadID)
	assert.NotEmpty(t, receivedBody)

	// Verify headers in MIME content
	headers := receivedHeaders
	assert.Contains(t, headers, "From: Sales Team <sales@logistics.com>")
	assert.Contains(t, headers, "To: customer@example.com")
	assert.Contains(t, headers, "Subject: Re: Incomplete RFQ details")
	assert.Contains(t, headers, "In-Reply-To: <parent@example.com>")
	assert.Contains(t, headers, "References: <parent@example.com> <older@example.com>")
	assert.Contains(t, headers, "Please provide the weight.")
}
