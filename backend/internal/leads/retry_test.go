package leads_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/freel/backend/internal/common/crypto"
	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/leads"
	"github.com/freel/backend/internal/organization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryAndDrafts(t *testing.T) {
	encKey := "my-secure-test-passphrase-32-bytes"
	t.Setenv("MAILBOX_ENCRYPTION_KEY", encKey)
	t.Setenv("GOOGLE_CLIENT_ID", "dummy-client-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "dummy-client-secret")

	dbURL := "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC&multiStatements=true"
	db, err := database.Connect(dbURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	orgID := int64(9995)
	otherOrgID := int64(9996)

	// Seed organizations
	_, err = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Retry Test Org A', NOW(), NOW()) ON DUPLICATE KEY UPDATE name = VALUES(name)", orgID)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Retry Test Org B', NOW(), NOW()) ON DUPLICATE KEY UPDATE name = VALUES(name)", otherOrgID)
	require.NoError(t, err)

	// Clean existing test data
	_, _ = db.Exec("DELETE FROM lead_email_drafts WHERE org_id IN (?, ?)", orgID, otherOrgID)
	_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id IN (?, ?)", orgID, otherOrgID)
	_, _ = db.Exec("DELETE FROM org_connected_mailboxes WHERE org_id IN (?, ?)", orgID, otherOrgID)
	_, _ = db.Exec("DELETE FROM leads WHERE org_id IN (?, ?)", orgID, otherOrgID)

	defer func() {
		_, _ = db.Exec("DELETE FROM lead_email_drafts WHERE org_id IN (?, ?)", orgID, otherOrgID)
		_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id IN (?, ?)", orgID, otherOrgID)
		_, _ = db.Exec("DELETE FROM org_connected_mailboxes WHERE org_id IN (?, ?)", orgID, otherOrgID)
		_, _ = db.Exec("DELETE FROM leads WHERE org_id IN (?, ?)", orgID, otherOrgID)
	}()

	// Seed Leads
	_, err = db.Exec("INSERT INTO leads (id, org_id, company_name, status, created_at, updated_at) VALUES (851, ?, 'Test Client A', 'NEW', NOW(), NOW())", orgID)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO leads (id, org_id, company_name, status, created_at, updated_at) VALUES (852, ?, 'Test Client B', 'NEW', NOW(), NOW())", otherOrgID)
	require.NoError(t, err)

	// Encrypt tokens
	accessTokenEnc, err := crypto.Encrypt("dummy_access", encKey)
	require.NoError(t, err)
	refreshTokenEnc, err := crypto.Encrypt("dummy_refresh", encKey)
	require.NoError(t, err)

	// Seed mailboxes
	_, err = db.Exec(`INSERT INTO org_connected_mailboxes (id, org_id, email, status, owner_name, provider, access_token_encrypted, refresh_token_encrypted, token_expiry, created_at, updated_at)
		VALUES (105, ?, 'sales@testorg.com', 'CONNECTED', 'Primary Mailbox', 'GMAIL', ?, ?, DATE_ADD(NOW(), INTERVAL 1 HOUR), NOW(), NOW())`, orgID, accessTokenEnc, refreshTokenEnc)
	require.NoError(t, err)

	// Seed failed outbound interaction (eligible for retry)
	// Ensure ai_summary and drafted_reply are set to '' to avoid sqlx scan errors (NULL to string)
	_, err = db.Exec(`INSERT INTO lead_interactions (id, org_id, lead_id, mailbox_id, channel, direction, subject, content, raw_email_id, thread_id, sentiment, intent, parent_interaction_id, rfc_message_id, sender, recipients, status, ai_summary, drafted_reply, created_at, updated_at)
		VALUES (751, ?, 851, 105, 'EMAIL', 'OUTBOUND', 'Outbound Reply Test', 'My outbound body', '', 'thread_id_abc', 'NEUTRAL', 'RFQ_REQUEST_INCOMPLETE', NULL, '<parent_msg_123@freel.local>', 'sales@testorg.com', 'customer@example.com', 'FAILED', '', '', NOW(), NOW())`, orgID)
	require.NoError(t, err)

	// Setup Business Logic and Provider Mocks
	dl := leads.NewDataLayer(db)
	orgRepo := organization.NewRepository(db)
	gp := organization.NewGmailProvider()

	bl := leads.NewBusinessLogic(dl, nil, orgRepo, gp, nil)

	// Inject a custom mocked transport to gp to mock the actual Gmail API send request
	var mockSendSuccess = true
	var mockSendErr error
	gp.HTTPClient.Transport = &mockTransport{
		roundTrip: func(req *http.Request) (*http.Response, error) {
			if !mockSendSuccess {
				if mockSendErr != nil {
					return nil, mockSendErr
				}
				return &http.Response{
					StatusCode: http.StatusForbidden,
					Body:       io.NopCloser(bytes.NewBufferString(`{"error": "insufficient scopes"}`)),
				}, nil
			}
			// Success mock response
			respBody := `{"id": "msg_success_999", "threadId": "thread_sent_999"}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(respBody)),
			}, nil
		},
	}

	t.Run("Successful retry transitions FAILED -> PENDING -> SENT and resets errors", func(t *testing.T) {
		mockSendSuccess = true
		mockSendErr = nil

		inter, err := bl.RetryEmailInteraction(ctx, orgID, 851, 751)
		require.NoError(t, err)
		assert.Equal(t, "SENT", inter.Status)
		assert.Equal(t, 1, inter.RetryCount)
		assert.Nil(t, inter.LastError)
		assert.Equal(t, "msg_success_999", inter.RawEmailID)
		assert.Equal(t, "thread_sent_999", inter.ThreadID)
	})

	t.Run("Concurrency lock blocks multiple retry requests", func(t *testing.T) {
		// Reset state back to FAILED
		_, err = db.Exec("UPDATE lead_interactions SET status = 'FAILED', retry_count = 0, last_error = NULL WHERE id = 751")
		require.NoError(t, err)

		// Call LockInteractionForRetry atomically
		locked1, err := dl.LockInteractionForRetry(ctx, orgID, 751)
		require.NoError(t, err)
		assert.True(t, locked1)

		// Second concurrent call should fail to acquire lock
		locked2, err := dl.LockInteractionForRetry(ctx, orgID, 751)
		require.NoError(t, err)
		assert.False(t, locked2)
	})

	t.Run("Failed retry transitions FAILED -> PENDING -> FAILED and sanitizes errors", func(t *testing.T) {
		// Reset state back to FAILED
		_, err = db.Exec("UPDATE lead_interactions SET status = 'FAILED', retry_count = 0, last_error = NULL WHERE id = 751")
		require.NoError(t, err)

		mockSendSuccess = false
		mockSendErr = errors.New("network timeout connection refused")

		// Execute retry
		inter, err := bl.RetryEmailInteraction(ctx, orgID, 851, 751)
		require.Error(t, err)
		assert.Equal(t, "FAILED", inter.Status)
		assert.Equal(t, 1, inter.RetryCount)
		require.NotNil(t, inter.LastError)
		// Error must be sanitized before exposure to client!
		assert.Equal(t, "Could not send this email. Please try again.", *inter.LastError)
	})

	t.Run("Email Draft Persistence CRUD and Org Isolation Checks", func(t *testing.T) {
		parentInteractionID := int64(751)

		// 1. Get draft initially (should return nil)
		draft, err := bl.GetDraft(ctx, orgID, 851, parentInteractionID)
		require.NoError(t, err)
		assert.Nil(t, draft)

		// 2. Save draft
		d1 := &leads.LeadEmailDraft{
			OrgID:               orgID,
			LeadID:              851,
			ParentInteractionID: parentInteractionID,
			MailboxID:           nil,
			Recipients:          "customer@test.com",
			CCRecipients:        "manager@test.com",
			Subject:             "Re: RFQ Query",
			Content:             "This is my draft body content",
		}
		err = bl.SaveDraft(ctx, d1)
		require.NoError(t, err)

		// 3. Retrieve draft
		draft, err = bl.GetDraft(ctx, orgID, 851, parentInteractionID)
		require.NoError(t, err)
		require.NotNil(t, draft)
		assert.Equal(t, "customer@test.com", draft.Recipients)
		assert.Equal(t, "This is my draft body content", draft.Content)

		// 4. Update draft content
		draft.Content = "Updated draft body content"
		err = bl.SaveDraft(ctx, draft)
		require.NoError(t, err)

		draft, err = bl.GetDraft(ctx, orgID, 851, parentInteractionID)
		require.NoError(t, err)
		assert.Equal(t, "Updated draft body content", draft.Content)

		// 5. Org Isolation Checks (retrieve via otherOrgID should return nil)
		otherDraft, err := bl.GetDraft(ctx, otherOrgID, 851, parentInteractionID)
		require.NoError(t, err)
		assert.Nil(t, otherDraft)

		// 6. Delete draft
		err = bl.DeleteDraft(ctx, orgID, 851, parentInteractionID)
		require.NoError(t, err)

		draft, err = bl.GetDraft(ctx, orgID, 851, parentInteractionID)
		require.NoError(t, err)
		assert.Nil(t, draft)
	})
}
