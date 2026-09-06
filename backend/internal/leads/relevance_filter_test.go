package leads_test

import (
	"context"
	"testing"
	"time"

	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/leads"
	"github.com/freel/backend/internal/organization"
	"github.com/stretchr/testify/require"
)

func TestIsLogisticsEmail(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		body     string
		expected bool
	}{
		{
			name:     "Logistics RFQ Inquiry",
			subject:  "Quote for 40ft container from Mumbai to Hamburg",
			body:     "Hi, please send us freight rates for 15,000 kg cargo FOB Incoterms.",
			expected: true,
		},
		{
			name:     "Personal Email Message",
			subject:  "Hey Varun, catch up tomorrow?",
			body:     "Hi Varun! Are you free for lunch tomorrow at 1 PM? Let me know!",
			expected: false,
		},
		{
			name:     "Irrelevant General Email",
			subject:  "Weekend plans and recipes",
			body:     "Here is the recipe for the pasta we talked about last night.",
			expected: false,
		},
		{
			name:     "Logistics Tracking Inquiry",
			subject:  "Status of shipment #9082",
			body:     "Can you give an update on the bill of lading and vessel arrival date in Hamburg?",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := leads.IsLogisticsEmail(tt.subject, tt.body)
			if got != tt.expected {
				t.Errorf("IsLogisticsEmail(%q, %q) = %v, expected %v", tt.subject, tt.body, got, tt.expected)
			}
		})
	}
}

func TestProcessInboundEmailRelevanceAndNewLead(t *testing.T) {
	dbURL := "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC&multiStatements=true"
	db, err := database.Connect(dbURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	orgID := int64(9960)

	// Seed organization
	_, err = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Relevance Org', NOW(), NOW()) ON DUPLICATE KEY UPDATE name = VALUES(name)", orgID)
	require.NoError(t, err)

	cleanup := func() {
		_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id = ?", orgID)
		_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	}
	cleanup()
	defer cleanup()

	dl := leads.NewDataLayer(db)
	orgRepo := organization.NewRepository(db)
	gp := organization.NewGmailProvider()
	eb := events.NewInProcessBus()
	bl := leads.NewBusinessLogic(dl, eb, orgRepo, gp, nil)

	// 1. Process negative non-logistics emails (Bank alerts, job alerts, newsletters, OTPs, personal emails)
	negativeEmails := []leads.InboundEmail{
		{
			RawEmailID: "neg-1-sbi-bank",
			From:       "cbsalerts.sbi@alerts.sbi.bank.in",
			Subject:    "CBSSBI ALERT",
			Body:       "Hold for INR 13,800.00 created in deposit Ac XXXXX586357 on 27/08/26.",
		},
		{
			RawEmailID: "neg-2-linkedin-job",
			From:       "jobalerts-noreply@linkedin.com",
			Subject:    "Staff Technical Program Manager - Actively recruiting",
			Body:       "Top jobs recommended for you this week.",
		},
		{
			RawEmailID: "neg-3-indian-express",
			From:       "subscribe@indianexpressonline.in",
			Subject:    "Indian Express Daily Briefing",
			Body:       "Top headlines and breaking news updates for today.",
		},
		{
			RawEmailID: "neg-4-hdfc-otp",
			From:       "no-reply@hdfcbank.com",
			Subject:    "Your One Time Password (OTP) for transaction",
			Body:       "Your OTP code is 984012. Do not share it with anyone.",
		},
		{
			RawEmailID: "neg-5-personal-lunch",
			From:       "friend@gmail.com",
			Subject:    "Hey, catch up tomorrow for lunch?",
			Body:       "Are you free for lunch tomorrow at 1 PM?",
		},
	}

	for i, negEmail := range negativeEmails {
		inter, err := bl.ProcessInboundEmail(ctx, int32(orgID), negEmail)
		require.NoError(t, err)
		require.NotNil(t, inter)
		require.Equal(t, "IGNORED", inter.Status, "Negative case %d (%s) should have status IGNORED", i+1, negEmail.From)
	}

	// Verify NO leads were created in DB for negative emails
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM leads WHERE org_id = ?", orgID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count, "Expected 0 leads in DB for negative cases")

	// 2. Process a legitimate logistics inquiry from shipper@company.com -> Creates Lead 1
	logisticsEmail1 := leads.InboundEmail{
		MailboxID:  0,
		RawEmailID: "msg-logistics-rel-1",
		ThreadID:   "thread-freight-101",
		From:       "shipper@company.com",
		Subject:    "RFQ: Freight rate for 20ft container",
		Body:       "Please quote sea freight from Nhava Sheva to Rotterdam.",
		ReceivedAt: time.Now(),
	}

	inter1, err := bl.ProcessInboundEmail(ctx, int32(orgID), logisticsEmail1)
	require.NoError(t, err)
	require.NotNil(t, inter1)
	lead1ID := inter1.LeadID
	require.Greater(t, lead1ID, int64(0))

	// Verify Lead was created in DB
	err = db.QueryRow("SELECT COUNT(*) FROM leads WHERE org_id = ?", orgID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count, "Expected exactly 1 lead in DB for authentic logistics inquiry")

	// 3. Process a SECOND, NEW independent logistics inquiry from the SAME sender (different email thread)
	logisticsEmail2 := leads.InboundEmail{
		MailboxID:  0,
		RawEmailID: "msg-logistics-rel-2",
		From:       "shipper@company.com",
		Subject:    "RFQ: Separate shipment from Shanghai to Dubai",
		Body:       "We have a new project shipping 5000 kg air freight from Shanghai to Dubai.",
		ReceivedAt: time.Now(),
	}

	inter2, err := bl.ProcessInboundEmail(ctx, int32(orgID), logisticsEmail2)
	require.NoError(t, err)
	require.NotNil(t, inter2)
	lead2ID := inter2.LeadID
	require.Greater(t, lead2ID, int64(0))

	// Verify that the second new email created a NEW separate Lead ID (not merged into Lead 1)
	require.NotEqual(t, lead1ID, lead2ID, "Expected separate Lead IDs for independent new emails from same sender")
}
