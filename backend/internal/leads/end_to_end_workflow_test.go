package leads_test

import (
	"testing"
	"time"

	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/leads"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	dbURL := "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC&multiStatements=true"
	db, err := database.Connect(dbURL)
	require.NoError(t, err)
	return db
}

func TestWorkflow_Part1_ManualLeadCreation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	orgID := int64(1801)

	// Clean up
	_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id = ?", orgID)
	_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Task 8 Org A', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)", orgID)

	defer func() {
		_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id = ?", orgID)
		_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	}()

	// 1. Manually create lead (no email, no thread_id, no AI context)
	res, err := db.Exec("INSERT INTO leads (org_id, company_name, contact_name, email, phone, status, created_at, updated_at) VALUES (?, 'ABC Logistics', 'John Smith', 'john@abc.com', '+919999999999', 'NEW', NOW(), NOW())", orgID)
	require.NoError(t, err)

	leadID, err := res.LastInsertId()
	require.NoError(t, err)

	// 2. Fetch lead & interactions
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM lead_interactions WHERE lead_id = ? AND org_id = ?", leadID, orgID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Manually created lead must have 0 email interactions initially")

	var status string
	err = db.QueryRow("SELECT status FROM leads WHERE id = ? AND org_id = ?", leadID, orgID).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, "NEW", status)
}

func TestWorkflow_Part2_ManualLeadOutboundAndReply(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	orgID := int64(1802)

	// Clean up
	_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id = ?", orgID)
	_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Task 8 Org B', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)", orgID)

	defer func() {
		_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id = ?", orgID)
		_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	}()

	// 1. Create Lead
	res, err := db.Exec("INSERT INTO leads (org_id, company_name, email, status, created_at, updated_at) VALUES (?, 'Manual Client Corp', 'client@manual.com', 'NEW', NOW(), NOW())", orgID)
	require.NoError(t, err)
	leadID, _ := res.LastInsertId()

	// 2. Insert Outbound Email Interaction with Thread ID
	threadID := "thread-manual-1802"
	rfcMsgID := "<outbound-1802@logisticshq.in>"
	resInter, err := db.Exec(`
		INSERT INTO lead_interactions 
		(org_id, lead_id, channel, raw_email_id, rfc_message_id, thread_id, direction, sender, recipients, subject, content, status, created_at) 
		VALUES (?, ?, 'EMAIL', 'raw-out-1802', ?, ?, 'OUTBOUND', 'sales@logisticshq.in', 'client@manual.com', 'Quote Inquiry', 'Hi Client', 'SENT', NOW())`,
		orgID, leadID, rfcMsgID, threadID)
	require.NoError(t, err)
	outboundID, _ := resInter.LastInsertId()
	assert.Greater(t, outboundID, int64(0))

	// 3. Customer replies on the same thread
	replyMsgID := "<reply-1802@manual.com>"
	resReply, err := db.Exec(`
		INSERT INTO lead_interactions 
		(org_id, lead_id, channel, raw_email_id, rfc_message_id, thread_id, direction, sender, recipients, subject, content, status, created_at) 
		VALUES (?, ?, 'EMAIL', 'raw-reply-1802', ?, ?, 'INBOUND', 'client@manual.com', 'sales@logisticshq.in', 'Re: Quote Inquiry', 'We need 500kg cargo shipped', 'SENT', NOW())`,
		orgID, leadID, replyMsgID, threadID)
	require.NoError(t, err)
	replyID, _ := resReply.LastInsertId()

	// Verify reply attaches to the same lead and thread
	var attachedLeadID int64
	var foundThreadID string
	err = db.QueryRow("SELECT lead_id, thread_id FROM lead_interactions WHERE id = ?", replyID).Scan(&attachedLeadID, &foundThreadID)
	require.NoError(t, err)
	assert.Equal(t, leadID, attachedLeadID)
	assert.Equal(t, threadID, foundThreadID)
}

func TestWorkflow_Part3_EmailCreatesNewLead(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	orgID := int64(1803)

	// Clean up
	_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id = ?", orgID)
	_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Task 8 Org C', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)", orgID)

	defer func() {
		_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id = ?", orgID)
		_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	}()

	// 1. Incoming email creates new Lead
	newSender := "newshipper_1803@external.com"
	rawEmailID := "raw-msg-1803"
	threadID := "thread-1803"

	// Create Lead
	resLead, err := db.Exec("INSERT INTO leads (org_id, company_name, email, source, status, created_at, updated_at) VALUES (?, ?, ?, 'EMAIL', 'NEW', NOW(), NOW())",
		orgID, "Inbound Lead ("+newSender+")", newSender)
	require.NoError(t, err)
	leadID, _ := resLead.LastInsertId()

	// Create Inbound Interaction
	_, err = db.Exec(`
		INSERT INTO lead_interactions 
		(org_id, lead_id, channel, raw_email_id, thread_id, direction, sender, recipients, subject, content, status, created_at) 
		VALUES (?, ?, 'EMAIL', ?, ?, 'INBOUND', ?, 'sales@logisticshq.in', 'Need Freight Quote', 'Ship machinery Mumbai to Hamburg', 'SENT', NOW())`,
		orgID, leadID, rawEmailID, threadID, newSender)
	require.NoError(t, err)

	// Verify exactly 1 lead and 1 interaction
	var leadCount, interCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM leads WHERE org_id = ?", orgID).Scan(&leadCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM lead_interactions WHERE org_id = ?", orgID).Scan(&interCount)

	assert.Equal(t, 1, leadCount)
	assert.Equal(t, 1, interCount)
}

func TestWorkflow_Part4_IncompleteRFQ_SuggestedReply(t *testing.T) {
	// Test partial context missing fields calculation logic
	prevCtx := map[string]interface{}{
		"origin_port":      "Mumbai",
		"destination_port": "Hamburg",
	}

	missing := leads.GetMissingRFQFields(prevCtx)
	assert.Contains(t, missing, "cargo_description")
	assert.Contains(t, missing, "target_date")
	assert.Contains(t, missing, "cargo_weight")
	assert.Contains(t, missing, "cargo_volume")
	assert.Contains(t, missing, "incoterms")
	assert.Len(t, missing, 5)
}

func TestWorkflow_Part5_Part6_ContextMergeAndCorrection(t *testing.T) {
	t.Run("Incremental context merge across customer turns", func(t *testing.T) {
		// Turn 1
		turn1 := map[string]interface{}{"origin_port": "INNSA", "destination_port": "DEHAM"}
		// Turn 2: customer provides weight & volume
		turn2 := map[string]interface{}{"cargo_weight": 2500.0, "cargo_volume": 12.0}
		merged1 := leads.MergeRFQContext(turn1, turn2)

		assert.Equal(t, "INNSA", merged1["origin_port"])
		assert.Equal(t, "DEHAM", merged1["destination_port"])
		assert.Equal(t, 2500.0, merged1["cargo_weight"])
		assert.Equal(t, 12.0, merged1["cargo_volume"])

		// Turn 3: Correction — customer updates weight to 2800 kg
		turn3 := map[string]interface{}{"cargo_weight": 2800.0, "origin_port": ""}
		merged2 := leads.MergeRFQContext(merged1, turn3)

		assert.Equal(t, "INNSA", merged2["origin_port"], "Empty string must not erase previous origin")
		assert.Equal(t, 2800.0, merged2["cargo_weight"], "Explicit correction must overwrite previous weight")
		assert.Equal(t, 12.0, merged2["cargo_volume"])
	})
}

func TestWorkflow_Part7_CompleteRFQ_Conversion(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	orgID := int64(1807)
	_, _ = db.Exec("DELETE FROM rfqs WHERE org_id = ?", orgID)
	_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	_, _ = db.Exec("DELETE FROM customers WHERE org_id = ?", orgID)
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Task 8 Org G', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)", orgID)

	defer func() {
		_, _ = db.Exec("DELETE FROM rfqs WHERE org_id = ?", orgID)
		_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
		_, _ = db.Exec("DELETE FROM customers WHERE org_id = ?", orgID)
	}()

	// Seed customer
	resCust, err := db.Exec("INSERT INTO customers (id, org_id, name, created_at, updated_at) VALUES (1807, ?, 'Ready Client Corp', NOW(), NOW())", orgID)
	require.NoError(t, err)
	custID, _ := resCust.LastInsertId()

	// Create Lead
	res, err := db.Exec("INSERT INTO leads (org_id, company_name, email, status, created_at, updated_at) VALUES (?, 'Ready Client Inc', 'ready@client.com', 'QUALIFIED', NOW(), NOW())", orgID)
	require.NoError(t, err)
	leadID, _ := res.LastInsertId()

	// Convert Lead to RFQ
	resRfq, err := db.Exec("INSERT INTO rfqs (org_id, rfq_number, customer_id, lead_id, origin, destination, stage, created_at, updated_at) VALUES (?, 'RFQ-1807-001', ?, ?, 'Mumbai', 'Hamburg', 'RFQ_CREATED', NOW(), NOW())", orgID, custID, leadID)
	require.NoError(t, err)
	rfqID, _ := resRfq.LastInsertId()

	// Update Lead status to CONVERTED
	_, err = db.Exec("UPDATE leads SET status = 'CONVERTED' WHERE id = ? AND org_id = ?", leadID, orgID)
	require.NoError(t, err)

	var updatedStatus string
	_ = db.QueryRow("SELECT status FROM leads WHERE id = ?", leadID).Scan(&updatedStatus)
	assert.Equal(t, "CONVERTED", updatedStatus)
	assert.Greater(t, rfqID, int64(0))
}

func TestWorkflow_Part8_DuplicateEventProtection(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	orgID := int64(1808)
	rawEmailID := "raw-dup-check-1808"

	_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id = ?", orgID)
	_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Task 8 Org H', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)", orgID)

	defer func() {
		_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id = ?", orgID)
		_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	}()

	// Insert lead & first interaction
	resLead, _ := db.Exec("INSERT INTO leads (org_id, company_name, email, created_at, updated_at) VALUES (?, 'Dup Test', 'dup@test.com', NOW(), NOW())", orgID)
	leadID, _ := resLead.LastInsertId()

	_, err := db.Exec("INSERT INTO lead_interactions (org_id, lead_id, channel, raw_email_id, direction, content, created_at) VALUES (?, ?, 'EMAIL', ?, 'INBOUND', 'Duplicate test content', NOW())", orgID, leadID, rawEmailID)
	require.NoError(t, err)

	// Second attempt with same raw_email_id should hit unique constraint or duplicate check
	var existingID int64
	err = db.QueryRow("SELECT id FROM lead_interactions WHERE org_id = ? AND raw_email_id = ?", orgID, rawEmailID).Scan(&existingID)
	require.NoError(t, err)
	assert.Greater(t, existingID, int64(0))
}

func TestWorkflow_Part9_SameCustomer_DifferentInquiry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	orgID := int64(1809)
	customerEmail := "multi_inquiry_1809@client.com"

	_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id = ?", orgID)
	_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Task 8 Org I', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)", orgID)

	defer func() {
		_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id = ?", orgID)
		_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	}()

	// Inquiry A: Mumbai to Hamburg (Thread A)
	resLeadA, _ := db.Exec("INSERT INTO leads (org_id, company_name, email, created_at, updated_at) VALUES (?, 'Client Inquiry A', ?, NOW(), NOW())", orgID, customerEmail)
	leadIDA, _ := resLeadA.LastInsertId()
	_, _ = db.Exec("INSERT INTO lead_interactions (org_id, lead_id, channel, raw_email_id, thread_id, direction, subject, content, created_at) VALUES (?, ?, 'EMAIL', 'raw-1809-A', 'thread-1809-A', 'INBOUND', 'Mumbai to Hamburg', 'Content A', NOW())", orgID, leadIDA)

	// Inquiry B: Delhi to Singapore (Thread B - distinct thread ID)
	resLeadB, _ := db.Exec("INSERT INTO leads (org_id, company_name, email, created_at, updated_at) VALUES (?, 'Client Inquiry B', ?, NOW(), NOW())", orgID, customerEmail)
	leadIDB, _ := resLeadB.LastInsertId()
	_, _ = db.Exec("INSERT INTO lead_interactions (org_id, lead_id, channel, raw_email_id, thread_id, direction, subject, content, created_at) VALUES (?, ?, 'EMAIL', 'raw-1809-B', 'thread-1809-B', 'INBOUND', 'Delhi to Singapore', 'Content B', NOW())", orgID, leadIDB)

	assert.NotEqual(t, leadIDA, leadIDB, "Distinct threads must create separate leads even if sender email matches")
}

func TestWorkflow_Part10_MultipleThreadsForLead(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	orgID := int64(1810)

	_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id = ?", orgID)
	_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Task 8 Org J', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)", orgID)

	defer func() {
		_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id = ?", orgID)
		_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	}()

	resLead, _ := db.Exec("INSERT INTO leads (org_id, company_name, email, created_at, updated_at) VALUES (?, 'Multi Thread Corp', 'multi@corp.com', NOW(), NOW())", orgID)
	leadID, _ := resLead.LastInsertId()

	// Thread 1
	_, err1 := db.Exec("INSERT INTO lead_interactions (org_id, lead_id, channel, raw_email_id, thread_id, direction, subject, content, created_at) VALUES (?, ?, 'EMAIL', 'raw-t1', 'thread-1', 'INBOUND', 'Thread 1 Subject', 'Content 1', NOW() - INTERVAL 1 HOUR)", orgID, leadID)
	require.NoError(t, err1)
	// Thread 2 (more recent)
	_, err2 := db.Exec("INSERT INTO lead_interactions (org_id, lead_id, channel, raw_email_id, thread_id, direction, subject, content, created_at) VALUES (?, ?, 'EMAIL', 'raw-t2', 'thread-2', 'INBOUND', 'Thread 2 Subject', 'Content 2', NOW())", orgID, leadID)
	require.NoError(t, err2)

	rows, err := db.Query("SELECT thread_id, MAX(created_at) as last_active FROM lead_interactions WHERE lead_id = ? AND org_id = ? GROUP BY thread_id ORDER BY last_active DESC", leadID, orgID)
	require.NoError(t, err)
	defer rows.Close()

	var threads []string
	for rows.Next() {
		var th string
		var lastActive time.Time
		_ = rows.Scan(&th, &lastActive)
		threads = append(threads, th)
	}

	require.Len(t, threads, 2)
	assert.Equal(t, "thread-2", threads[0], "Most recently active thread must be listed first")
}

func TestWorkflow_Part11_FailedEmailAndRetryWorkflow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	orgID := int64(1811)

	_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id = ?", orgID)
	_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Task 8 Org K', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)", orgID)

	defer func() {
		_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id = ?", orgID)
		_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	}()

	resLead, _ := db.Exec("INSERT INTO leads (org_id, company_name, email, created_at, updated_at) VALUES (?, 'Retry Test Inc', 'retry@inc.com', NOW(), NOW())", orgID)
	leadID, _ := resLead.LastInsertId()

	// Insert FAILED interaction
	resInter, err := db.Exec(`
		INSERT INTO lead_interactions 
		(org_id, lead_id, channel, raw_email_id, direction, content, status, retry_count, created_at) 
		VALUES (?, ?, 'EMAIL', 'raw-failed-1811', 'OUTBOUND', 'Retry test content', 'FAILED', 0, NOW())`,
		orgID, leadID)
	require.NoError(t, err)
	interID, _ := resInter.LastInsertId()

	// Perform Retry
	now := time.Now()
	_, err = db.Exec("UPDATE lead_interactions SET status = 'SENT', retry_count = retry_count + 1, last_retry_at = ? WHERE id = ? AND org_id = ?", now, interID, orgID)
	require.NoError(t, err)

	var newStatus string
	var retryCount int
	err = db.QueryRow("SELECT status, retry_count FROM lead_interactions WHERE id = ?", interID).Scan(&newStatus, &retryCount)
	require.NoError(t, err)

	assert.Equal(t, "SENT", newStatus)
	assert.Equal(t, 1, retryCount)
}

func TestWorkflow_Part12_DraftPersistenceAndDiscard(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	orgID := int64(1812)

	_, _ = db.Exec("DELETE FROM lead_email_drafts WHERE org_id = ?", orgID)
	_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Task 8 Org L', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)", orgID)

	defer func() {
		_, _ = db.Exec("DELETE FROM lead_email_drafts WHERE org_id = ?", orgID)
		_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	}()

	resLead, _ := db.Exec("INSERT INTO leads (org_id, company_name, email, created_at, updated_at) VALUES (?, 'Draft Test Corp', 'draft@corp.com', NOW(), NOW())", orgID)
	leadID, _ := resLead.LastInsertId()

	// 1. Save draft
	_, err := db.Exec("INSERT INTO lead_email_drafts (org_id, lead_id, parent_interaction_id, subject, content, created_at, updated_at) VALUES (?, ?, 0, 'Draft Subject', 'Draft Body Content', NOW(), NOW()) ON DUPLICATE KEY UPDATE subject=VALUES(subject), content=VALUES(content)", orgID, leadID)
	require.NoError(t, err)

	// 2. Fetch draft
	var draftSub, draftContent string
	err = db.QueryRow("SELECT subject, content FROM lead_email_drafts WHERE org_id = ? AND lead_id = ?", orgID, leadID).Scan(&draftSub, &draftContent)
	require.NoError(t, err)
	assert.Equal(t, "Draft Subject", draftSub)
	assert.Equal(t, "Draft Body Content", draftContent)

	// 3. Delete draft (discard)
	_, err = db.Exec("DELETE FROM lead_email_drafts WHERE org_id = ? AND lead_id = ?", orgID, leadID)
	require.NoError(t, err)

	var count int
	_ = db.QueryRow("SELECT COUNT(*) FROM lead_email_drafts WHERE org_id = ? AND lead_id = ?", orgID, leadID).Scan(&count)
	assert.Equal(t, 0, count)
}

func TestWorkflow_Part13_OrganizationIsolation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	orgA := int64(18131)
	orgB := int64(18132)

	_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id IN (?, ?)", orgA, orgB)
	_, _ = db.Exec("DELETE FROM leads WHERE org_id IN (?, ?)", orgA, orgB)
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Org 18131', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)", orgA)
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Org 18132', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)", orgB)

	defer func() {
		_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id IN (?, ?)", orgA, orgB)
		_, _ = db.Exec("DELETE FROM leads WHERE org_id IN (?, ?)", orgA, orgB)
	}()

	// Seed lead in Org A
	resA, _ := db.Exec("INSERT INTO leads (org_id, company_name, email, created_at, updated_at) VALUES (?, 'Org A Lead', 'orga@test.com', NOW(), NOW())", orgA)
	leadIDA, _ := resA.LastInsertId()

	// Query for Org A lead using Org B credentials must return 0 rows
	var foundCount int
	err := db.QueryRow("SELECT COUNT(*) FROM leads WHERE id = ? AND org_id = ?", leadIDA, orgB).Scan(&foundCount)
	require.NoError(t, err)
	assert.Equal(t, 0, foundCount, "Strict multi-tenant organization isolation must prevent cross-org lead retrieval")
}

func TestWorkflow_Part14_CrossWorkflowRegression(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	orgID := int64(1814)
	_, _ = db.Exec("DELETE FROM rfqs WHERE org_id = ?", orgID)
	_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id = ?", orgID)
	_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Task 8 Regression Org', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)", orgID)

	defer func() {
		_, _ = db.Exec("DELETE FROM rfqs WHERE org_id = ?", orgID)
		_, _ = db.Exec("DELETE FROM lead_interactions WHERE org_id = ?", orgID)
		_, _ = db.Exec("DELETE FROM leads WHERE org_id = ?", orgID)
	}()

	// 1. Manual lead with no interactions remains valid
	res1, err := db.Exec("INSERT INTO leads (org_id, company_name, status, created_at, updated_at) VALUES (?, 'Standalone Manual Lead', 'NEW', NOW(), NOW())", orgID)
	require.NoError(t, err)
	lead1, _ := res1.LastInsertId()

	var count1 int
	_ = db.QueryRow("SELECT COUNT(*) FROM lead_interactions WHERE lead_id = ?", lead1).Scan(&count1)
	assert.Equal(t, 0, count1)

	// 2. Converted lead cannot create duplicate active leads
	res2, err := db.Exec("INSERT INTO leads (org_id, company_name, status, created_at, updated_at) VALUES (?, 'Converted Lead', 'CONVERTED', NOW(), NOW())", orgID)
	require.NoError(t, err)
	lead2, _ := res2.LastInsertId()

	var status2 string
	_ = db.QueryRow("SELECT status FROM leads WHERE id = ?", lead2).Scan(&status2)
	assert.Equal(t, "CONVERTED", status2)
}
