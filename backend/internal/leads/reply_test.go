package leads_test

import (
	"bytes"
	"context"
	"encoding/json"
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

type mockTransport struct {
	roundTrip func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTrip(req)
}

func TestReplyToInteraction(t *testing.T) {
	encKey := "my-secure-test-passphrase-32-bytes"
	t.Setenv("MAILBOX_ENCRYPTION_KEY", encKey)
	t.Setenv("GOOGLE_CLIENT_ID", "dummy-client-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "dummy-client-secret")

	dbURL := "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC&multiStatements=true"
	db, err := database.Connect(dbURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	orgID := int64(9991)
	otherOrgID := int64(9992)

	// Seed organizations
	_, err = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Reply Test Org A', NOW(), NOW()) ON DUPLICATE KEY UPDATE name = VALUES(name)", orgID)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Reply Test Org B', NOW(), NOW()) ON DUPLICATE KEY UPDATE name = VALUES(name)", otherOrgID)
	require.NoError(t, err)

	// Clean existing test data
	_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id IN (?, ?)", orgID, otherOrgID)
	_, _ = db.Exec("DELETE FROM org_connected_mailboxes WHERE org_id IN (?, ?)", orgID, otherOrgID)
	_, _ = db.Exec("DELETE FROM leads WHERE org_id IN (?, ?)", orgID, otherOrgID)

	defer func() {
		_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id IN (?, ?)", orgID, otherOrgID)
		_, _ = db.Exec("DELETE FROM org_connected_mailboxes WHERE org_id IN (?, ?)", orgID, otherOrgID)
		_, _ = db.Exec("DELETE FROM leads WHERE org_id IN (?, ?)", orgID, otherOrgID)
	}()

	// Seed Leads
	_, err = db.Exec("INSERT INTO leads (id, org_id, company_name, status, created_at, updated_at) VALUES (801, ?, 'Test Client Inc', 'NEW', NOW(), NOW())", orgID)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO leads (id, org_id, company_name, status, created_at, updated_at) VALUES (802, ?, 'Other Client Inc', 'NEW', NOW(), NOW())", otherOrgID)
	require.NoError(t, err)

	// Encrypt tokens for seeding
	dummyAccessEnc, err := crypto.Encrypt("dummy-access-token", encKey)
	require.NoError(t, err)
	dummyRefreshEnc, err := crypto.Encrypt("dummy-refresh-token", encKey)
	require.NoError(t, err)

	// Seed Mailboxes
	// Org A Mailbox 1 (Primary)
	_, err = db.Exec(`
		INSERT INTO org_connected_mailboxes (id, org_id, email, owner_name, mailbox_type, provider, status, access_token_encrypted, refresh_token_encrypted, is_primary) 
		VALUES (901, ?, 'primary@orga.com', 'Primary Owner', 'Individual', 'GMAIL', 'CONNECTED', ?, ?, 1)
	`, orgID, dummyAccessEnc, dummyRefreshEnc)
	require.NoError(t, err)

	// Org A Mailbox 2 (Secondary)
	_, err = db.Exec(`
		INSERT INTO org_connected_mailboxes (id, org_id, email, owner_name, mailbox_type, provider, status, access_token_encrypted, refresh_token_encrypted, is_primary) 
		VALUES (902, ?, 'secondary@orga.com', 'Secondary Owner', 'Individual', 'GMAIL', 'CONNECTED', ?, ?, 0)
	`, orgID, dummyAccessEnc, dummyRefreshEnc)
	require.NoError(t, err)

	// Org B Mailbox
	_, err = db.Exec(`
		INSERT INTO org_connected_mailboxes (id, org_id, email, owner_name, mailbox_type, provider, status, access_token_encrypted, refresh_token_encrypted, is_primary) 
		VALUES (903, ?, 'primary@orgb.com', 'Org B Owner', 'Individual', 'GMAIL', 'CONNECTED', ?, ?, 1)
	`, otherOrgID, dummyAccessEnc, dummyRefreshEnc)
	require.NoError(t, err)

	// Seed Parent Inbound Interaction
	_, err = db.Exec(`
		INSERT INTO lead_interactions (id, org_id, lead_id, channel, direction, subject, content, raw_email_id, thread_id, rfc_message_id, sender, recipients, mailbox_id, status, sentiment, intent, ai_summary, drafted_reply)
		VALUES (701, ?, 801, 'EMAIL', 'INBOUND', 'Freight quote request', 'Hi, please quote Nhava Sheva to Hamburg.', 'gmail-msg-123', 'gmail-thread-456', '<rfc-msg-123@mail.com>', 'shipper@client.com', 'secondary@orga.com', 902, 'SENT', 'NEUTRAL', 'RFQ_REQUEST', '', '')
	`, orgID)
	require.NoError(t, err)

	// Instantiate layers
	dl := leads.NewDataLayer(db)
	orgRepo := organization.NewRepository(db)

	gmailProvider := organization.NewGmailProvider()
	mockTransportInst := &mockTransport{}
	gmailProvider.HTTPClient.Transport = mockTransportInst

	bl := leads.NewBusinessLogic(dl, nil, orgRepo, gmailProvider, nil)

	t.Run("Successful manual reply updates status from PENDING to SENT and maps headers", func(t *testing.T) {
		mockTransportInst.roundTrip = func(req *http.Request) (*http.Response, error) {
			respJSON := map[string]string{
				"id":       "gmail-outbound-id-888",
				"threadId": "gmail-thread-456",
			}
			respBytes, _ := json.Marshal(respJSON)
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(respBytes)),
			}, nil
		}

		outbound, err := bl.ReplyToInteraction(
			ctx,
			orgID,
			801,
			701,
			"",
			"shipper@client.com",
			"",
			"Re: Freight quote request",
			"We are working on your quote.",
		)
		require.NoError(t, err)
		require.NotNil(t, outbound)

		// Assert mapping of identifiers
		assert.Equal(t, "SENT", outbound.Status)
		assert.Equal(t, "gmail-outbound-id-888", outbound.RawEmailID)
		assert.Equal(t, "gmail-thread-456", outbound.ThreadID)
		assert.Equal(t, "<rfc-msg-123@mail.com>", outbound.InReplyTo)
		assert.Contains(t, outbound.ReferencesHeader, "<rfc-msg-123@mail.com>")
		// Assert correct mailbox selection (rule 1: preferentially match parent mailbox ID = 902)
		assert.Equal(t, int64(902), *outbound.MailboxID)
		assert.Equal(t, "secondary@orga.com", outbound.Sender)

		// Verify stored in DB
		dbFetched, err := dl.GetInteractionByID(ctx, int32(orgID), outbound.ID)
		require.NoError(t, err)
		assert.Equal(t, "SENT", dbFetched.Status)
		assert.Equal(t, "gmail-outbound-id-888", dbFetched.RawEmailID)
	})

	t.Run("Failed manual reply updates status to FAILED and preserves fields", func(t *testing.T) {
		mockTransportInst.roundTrip = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 500,
				Body:       io.NopCloser(bytes.NewReader([]byte("Internal Gmail Error"))),
			}, nil
		}

		outbound, err := bl.ReplyToInteraction(
			ctx,
			orgID,
			801,
			701,
			"",
			"shipper@client.com",
			"",
			"Re: Freight quote request",
			"Failed attempt content.",
		)
		assert.Error(t, err)
		require.NotNil(t, outbound)

		// Status should be FAILED
		assert.Equal(t, "FAILED", outbound.Status)
		// Body must be preserved
		assert.Equal(t, "Failed attempt content.", outbound.Content)

		// Verify status in DB is FAILED
		dbFetched, err := dl.GetInteractionByID(ctx, int32(orgID), outbound.ID)
		require.NoError(t, err)
		assert.Equal(t, "FAILED", dbFetched.Status)
		assert.Equal(t, "Failed attempt content.", dbFetched.Content)
	})

	t.Run("Strict organization isolation mismatch returns error", func(t *testing.T) {
		// Attempting to reply to Org A interaction using Org B credentials
		_, err := bl.ReplyToInteraction(
			ctx,
			otherOrgID,
			801,
			701,
			"",
			"shipper@client.com",
			"",
			"Re: Freight quote request",
			"Unauthorized",
		)
		assert.Error(t, err)
	})
}
