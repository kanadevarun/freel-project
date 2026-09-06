package leads_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/leads"
	"github.com/freel/backend/internal/organization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConversationMatchingScenarios(t *testing.T) {
	dbURL := "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC&multiStatements=true"
	db, err := database.Connect(dbURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	orgA := int64(9801)
	orgB := int64(9802)

	// Seed organizations
	_, err = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Match Org A', NOW(), NOW()) ON DUPLICATE KEY UPDATE name = VALUES(name)", orgA)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Match Org B', NOW(), NOW()) ON DUPLICATE KEY UPDATE name = VALUES(name)", orgB)
	require.NoError(t, err)

	// Clean existing data for clean runs
	cleanup := func() {
		_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id IN (?, ?)", orgA, orgB)
		_, _ = db.Exec("DELETE FROM leads WHERE org_id IN (?, ?)", orgA, orgB)
	}
	cleanup()
	defer cleanup()

	// Instantiate business layers
	dl := leads.NewDataLayer(db)
	orgRepo := organization.NewRepository(db)
	gp := organization.NewGmailProvider()
	eb := events.NewInProcessBus()
	bl := leads.NewBusinessLogic(dl, eb, orgRepo, gp, nil)

	// Seed existing Lead in Org A
	res, err := db.Exec("INSERT INTO leads (org_id, company_name, status, created_at, updated_at, email) VALUES (?, 'Active Org A Client', 'NEW', NOW(), NOW(), 'customer@freel-testing.local')", orgA)
	require.NoError(t, err)
	leadID, err := res.LastInsertId()
	require.NoError(t, err)

	// Seed an existing interaction (Interaction ID: 601) in Org A under Lead leadID
	// Ensure ai_summary, drafted_reply, sentiment, and intent are explicitly set to avoid scan errors (NULL to string)
	_, err = db.Exec(fmt.Sprintf(`
		INSERT INTO lead_interactions (id, org_id, lead_id, channel, direction, subject, content, raw_email_id, thread_id, rfc_message_id, sender, recipients, status, sentiment, intent, ai_summary, drafted_reply, created_at, updated_at)
		VALUES (601, ?, %d, 'EMAIL', 'INBOUND', 'RFQ Steel Inquiry', 'Content payload', 'msg-123', 'thread-123', '<message-rfc-601@freel.local>', 'customer@freel-testing.local', 'sales@logisticshq.in', 'SENT', 'NEUTRAL', 'RFQ_REQUEST', '', '', NOW(), NOW())
	`, leadID), orgA)
	require.NoError(t, err)

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM lead_interactions WHERE org_id = ? AND thread_id = ?)", orgA, "thread-123").Scan(&exists)
	require.NoError(t, err)
	log.Printf("[TEST_DEBUG] Seeded interaction exists in DB: %t", exists)

	t.Run("Test 1 — Gmail Thread Match", func(t *testing.T) {
		// Inbound email with same Thread ID
		inbound := leads.InboundEmail{
			From:         "customer@freel-testing.local",
			Sender:       "customer@freel-testing.local",
			Recipients:   "sales@logisticshq.in",
			Subject:      "Re: RFQ Steel Inquiry",
			Body:         "Follow-up body text",
			RawEmailID:   "new-gmail-msg-threadmatch",
			ThreadID:     "thread-123",
			RFCMessageID: "<reply-msg-threadmatch@freel.local>",
		}

		inter, err := bl.ProcessInboundEmail(ctx, int32(orgA), inbound)
		require.NoError(t, err)
		assert.Equal(t, leadID, inter.LeadID)
		assert.Equal(t, "thread-123", inter.ThreadID)
		require.NotNil(t, inter.ParentInteractionID)
		assert.Equal(t, int64(601), *inter.ParentInteractionID)

		// Verify no new Lead was created
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM leads WHERE org_id = ?", orgA).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Should not create a duplicate Lead on Thread ID match")
	})

	t.Run("Test 2 — In-Reply-To Match", func(t *testing.T) {
		// Inbound email with In-Reply-To matching message-rfc-601
		inbound := leads.InboundEmail{
			From:         "customer@freel-testing.local",
			Sender:       "customer@freel-testing.local",
			Recipients:   "sales@logisticshq.in",
			Subject:      "Re: RFQ Steel Inquiry",
			Body:         "Follow-up using In-Reply-To",
			RawEmailID:   "new-gmail-msg-inreplyto",
			ThreadID:     "different-thread-id", // different thread ID but matching In-Reply-To
			InReplyTo:    "<message-rfc-601@freel.local>",
			RFCMessageID: "<reply-msg-inreplyto@freel.local>",
		}

		inter, err := bl.ProcessInboundEmail(ctx, int32(orgA), inbound)
		require.NoError(t, err)
		assert.Equal(t, leadID, inter.LeadID)
		require.NotNil(t, inter.ParentInteractionID)
		assert.Equal(t, int64(601), *inter.ParentInteractionID)

		// Verify no new Lead was created
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM leads WHERE org_id = ?", orgA).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Should not create a duplicate Lead on In-Reply-To match")
	})

	t.Run("Test 3 — References Match", func(t *testing.T) {
		// Inbound email with ReferencesHeader containing multiple IDs, one matches message-rfc-601
		inbound := leads.InboundEmail{
			From:             "customer@freel-testing.local",
			Sender:           "customer@freel-testing.local",
			Recipients:       "sales@logisticshq.in",
			Subject:          "Re: RFQ Steel Inquiry",
			Body:             "Follow-up using References",
			RawEmailID:       "new-gmail-msg-references",
			ThreadID:         "some-other-thread-id",
			ReferencesHeader: "<first@domain.com> <second@domain.com> <message-rfc-601@freel.local>",
			RFCMessageID:     "<reply-msg-references@freel.local>",
		}

		inter, err := bl.ProcessInboundEmail(ctx, int32(orgA), inbound)
		require.NoError(t, err)
		assert.Equal(t, leadID, inter.LeadID)
		require.NotNil(t, inter.ParentInteractionID)
		assert.Equal(t, int64(601), *inter.ParentInteractionID)

		// Verify no new Lead was created
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM leads WHERE org_id = ?", orgA).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Should not create a duplicate Lead on References match")
	})

	t.Run("Test 4 — Independent New Email", func(t *testing.T) {
		// Inbound email from same customer but with NO matching thread/reply headers
		inbound := leads.InboundEmail{
			From:         "customer@freel-testing.local",
			Sender:       "customer@freel-testing.local",
			Recipients:   "sales@logisticshq.in",
			Subject:      "A Completely New Freight Request",
			Body:         "Hi, I want a rate for a completely new route from Mundra to Felixstowe.",
			RawEmailID:   "new-independent-inquiry-msg",
			ThreadID:     "thread-completely-new-999",
			RFCMessageID: "<independent-new-inquiry@freel.local>",
		}

		inter, err := bl.ProcessInboundEmail(ctx, int32(orgA), inbound)
		require.NoError(t, err)
		// Since there is no conversation match (different Gmail Thread ID, no InReplyTo, no References),
		// this starts a new conversation and MUST create a new Lead!
		assert.NotEqual(t, leadID, inter.LeadID)

		// Verify that a new Lead record was created
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM leads WHERE org_id = ?", orgA).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 2, count, "Should create a new Lead for independent inquiry")
	})

	t.Run("Test 5 — Duplicate Gmail Message", func(t *testing.T) {
		// Resending msg-123 (which was seeded initially)
		inbound := leads.InboundEmail{
			From:         "customer@freel-testing.local",
			Sender:       "customer@freel-testing.local",
			Recipients:   "sales@logisticshq.in",
			Subject:      "RFQ Steel Inquiry",
			Body:         "Content payload",
			RawEmailID:   "msg-123", // Matches raw_email_id seeded initially
			ThreadID:     "thread-123",
			RFCMessageID: "<message-rfc-601@freel.local>",
		}

		inter, err := bl.ProcessInboundEmail(ctx, int32(orgA), inbound)
		require.NoError(t, err)
		// Should return the existing interaction and flag it as idempotent
		assert.True(t, inter.IsIdempotent)
		assert.Equal(t, int64(601), inter.ID)
	})

	t.Run("Test 6 — Organization Isolation", func(t *testing.T) {
		// Org B receives an inbound message with Thread ID thread-123
		inbound := leads.InboundEmail{
			From:         "customer@freel-testing.local",
			Sender:       "customer@freel-testing.local",
			Recipients:   "sales@logisticshq.in",
			Subject:      "Re: RFQ Steel Inquiry",
			Body:         "Org B thread matching check",
			RawEmailID:   "orgb-inbound-msg",
			ThreadID:     "thread-123", // Same thread ID as Org A conversation
			RFCMessageID: "<orgb-msg@freel.local>",
		}

		inter, err := bl.ProcessInboundEmail(ctx, int32(orgB), inbound)
		require.NoError(t, err)
		// Org B has no Lead matching Org A's thread, so thread-123 must NOT resolve to Org A's Lead!
		assert.NotEqual(t, leadID, inter.LeadID)

		// Verify that a new Lead was created under Org B
		var countB int
		err = db.QueryRow("SELECT COUNT(*) FROM leads WHERE org_id = ?", orgB).Scan(&countB)
		require.NoError(t, err)
		assert.Equal(t, 1, countB, "Should create a new Lead in Org B because of Org Isolation")
	})
}
