package organization_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/freel/backend/internal/common/crypto"
	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/organization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectedMailboxOperations(t *testing.T) {
	// Setup test environment encryption key
	encKey := "my-secure-test-passphrase-32-bytes"
	t.Setenv("MAILBOX_ENCRYPTION_KEY", encKey)

	// 1. Connect to local development database
	dbURL := "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC&multiStatements=true"
	db, err := database.Connect(dbURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	orgID := int64(8888)
	otherOrgID := int64(8889)

	// Seed organizations for foreign key constraints
	_, err = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Mailbox Test Org A', NOW(), NOW()) ON DUPLICATE KEY UPDATE name = VALUES(name)", orgID)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Mailbox Test Org B', NOW(), NOW()) ON DUPLICATE KEY UPDATE name = VALUES(name)", otherOrgID)
	require.NoError(t, err)

	// Clean up existing test data
	_, _ = db.Exec("DELETE FROM org_connected_mailboxes WHERE org_id IN (?, ?)", orgID, otherOrgID)

	defer func() {
		_, _ = db.Exec("DELETE FROM org_connected_mailboxes WHERE org_id IN (?, ?)", orgID, otherOrgID)
	}()

	repo := organization.NewRepository(db)
	svc := organization.NewService(repo, nil, nil)

	// --- Test Scenario 1: Key Check Failure ---
	t.Run("Fails when encryption key is not set", func(t *testing.T) {
		t.Setenv("MAILBOX_ENCRYPTION_KEY", "")
		req := organization.ConnectMailboxRequest{
			Email:        "gmail-test@example.com",
			OwnerName:    "Gmail Owner",
			MailboxType:  "Individual",
			Provider:     "GMAIL",
			AccessToken:  "secret-oauth-access-token",
			RefreshToken: "secret-oauth-refresh-token",
			OAuthScopes:  []string{"https://mail.google.com/"},
		}
		_, err := svc.ConnectMailbox(ctx, orgID, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MAILBOX_ENCRYPTION_KEY is not configured")
	})

	// Restore key for subsequent tests
	t.Setenv("MAILBOX_ENCRYPTION_KEY", encKey)

	// --- Test Scenario 2: Mailbox Creation & Encryption ---
	var mailboxID int64
	t.Run("Create connected mailbox with encrypted OAuth credentials", func(t *testing.T) {
		rawAccessToken := "secret-oauth-access-token-123"
		rawRefreshToken := "secret-oauth-refresh-token-456"
		expiry := time.Now().Add(1 * time.Hour)

		req := organization.ConnectMailboxRequest{
			Email:        "gmail-test@example.com",
			OwnerName:    "Gmail Owner",
			MailboxType:  "Individual",
			Provider:     "GMAIL",
			AccessToken:  rawAccessToken,
			RefreshToken: rawRefreshToken,
			TokenExpiry:  &expiry,
			OAuthScopes:  []string{"https://mail.google.com/", "offline"},
		}

		mailbox, err := svc.ConnectMailbox(ctx, orgID, req)
		require.NoError(t, err)
		require.NotNil(t, mailbox)
		mailboxID = mailbox.ID

		assert.Equal(t, "gmail-test@example.com", mailbox.Email)
		assert.Equal(t, "GMAIL", mailbox.Provider)
		assert.Equal(t, "CONNECTED", mailbox.Status)

		// Verify encrypted tokens are stored in DB
		fetched, err := svc.GetConnectedMailboxByID(ctx, mailboxID, orgID)
		require.NoError(t, err)
		require.NotNil(t, fetched)

		assert.NotNil(t, fetched.AccessTokenEncrypted)
		assert.NotNil(t, fetched.RefreshTokenEncrypted)

		// Plain text values must never match encrypted text
		assert.NotEqual(t, rawAccessToken, *fetched.AccessTokenEncrypted)
		assert.NotEqual(t, rawRefreshToken, *fetched.RefreshTokenEncrypted)

		// Verify decryptability only inside backend service
		decryptedAccess, err := crypto.Decrypt(*fetched.AccessTokenEncrypted, encKey)
		require.NoError(t, err)
		assert.Equal(t, rawAccessToken, decryptedAccess)

		decryptedRefresh, err := crypto.Decrypt(*fetched.RefreshTokenEncrypted, encKey)
		require.NoError(t, err)
		assert.Equal(t, rawRefreshToken, decryptedRefresh)
	})

	// --- Test Scenario 3: Response Serialization Security ---
	t.Run("Response serialization does not expose credentials", func(t *testing.T) {
		fetched, err := svc.GetConnectedMailboxByID(ctx, mailboxID, orgID)
		require.NoError(t, err)

		data, err := json.Marshal(fetched)
		require.NoError(t, err)

		jsonStr := string(data)
		assert.NotContains(t, jsonStr, "access_token_encrypted")
		assert.NotContains(t, jsonStr, "refresh_token_encrypted")
		assert.NotContains(t, jsonStr, "secret-oauth-access-token-123")
		assert.NotContains(t, jsonStr, "secret-oauth-refresh-token-456")
	})

	// --- Test Scenario 4: Disconnect Behavior (Safe and Idempotent) ---
	t.Run("Disconnect mailbox safely clears credentials and sets state", func(t *testing.T) {
		err := svc.DisconnectMailbox(ctx, mailboxID, orgID)
		require.NoError(t, err)

		// Verify database state
		fetched, err := svc.GetConnectedMailboxByID(ctx, mailboxID, orgID)
		require.NoError(t, err)
		assert.Equal(t, "DISCONNECTED", fetched.Status)
		assert.Nil(t, fetched.AccessTokenEncrypted)
		assert.Nil(t, fetched.RefreshTokenEncrypted)
		assert.Nil(t, fetched.SyncCursor)

		// Test idempotency (repeated calls do not fail or corrupt data)
		err = svc.DisconnectMailbox(ctx, mailboxID, orgID)
		assert.NoError(t, err)

		fetched2, err := svc.GetConnectedMailboxByID(ctx, mailboxID, orgID)
		require.NoError(t, err)
		assert.Equal(t, "DISCONNECTED", fetched2.Status)
		assert.Nil(t, fetched2.AccessTokenEncrypted)
		assert.Nil(t, fetched2.RefreshTokenEncrypted)
	})

	// --- Test Scenario 5: Organization Isolation ---
	t.Run("Strict organization isolation on retrieve, update, and disconnect", func(t *testing.T) {
		// Attempt to fetch mailbox of Org A (8888) using Org B (8889)
		fetched, err := svc.GetConnectedMailboxByID(ctx, mailboxID, otherOrgID)
		// Should return error or nil since SQL has WHERE id = ? AND org_id = ?
		assert.True(t, err != nil || fetched == nil)

		// Attempt to update mailbox of Org A using Org B
		updateReq := organization.UpdateMailboxRequest{
			OwnerName:         "Unauthorized Update",
			MailboxType:       "Shared / Team",
			SyncFrequency:     "Real-time",
			ProcessingEnabled: false,
		}
		err = svc.UpdateMailbox(ctx, mailboxID, otherOrgID, updateReq)
		assert.Error(t, err)

		// Attempt to disconnect mailbox of Org A using Org B
		err = svc.DisconnectMailbox(ctx, mailboxID, otherOrgID)
		assert.Error(t, err)

		// Attempt to remove mailbox of Org A using Org B
		err = svc.RemoveMailbox(ctx, mailboxID, otherOrgID)
		assert.Error(t, err)
	})
}
