package leads

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// LeadInteraction represents a recorded touchpoint or email exchange with a lead.
type LeadInteraction struct {
	ID                    int64                  `db:"id" json:"id"`
	OrgID                 int64                  `db:"org_id" json:"org_id"`
	LeadID                int64                  `db:"lead_id" json:"lead_id"`
	Channel               string                 `db:"channel" json:"channel"`       // EMAIL | PHONE | LINKEDIN | WHATSAPP
	Direction             string                 `db:"direction" json:"direction"`   // INBOUND | OUTBOUND
	Subject               string                 `db:"subject" json:"subject"`
	Content               string                 `db:"content" json:"content"`
	RawEmailID            string                 `db:"raw_email_id" json:"raw_email_id"`
	ThreadID              string                 `db:"thread_id" json:"thread_id"`
	Sentiment             string                 `db:"sentiment" json:"sentiment"`
	Intent                string                 `db:"intent" json:"intent"`
	LinkedRFQID           *int64                 `db:"linked_rfq_id" json:"linked_rfq_id,omitempty"`
	AIConfidence          int                    `db:"ai_confidence" json:"ai_confidence"`
	AISummary             string                 `db:"ai_summary" json:"ai_summary"`
	DraftedReply          string                 `db:"drafted_reply" json:"drafted_reply"`
	// Thread-awareness fields (added by migration 013e)
	ParentInteractionID   *int64                 `db:"parent_interaction_id" json:"parent_interaction_id,omitempty"`
	PartialRFQContext     map[string]interface{} `db:"-" json:"partial_rfq_context,omitempty"`
	PartialRFQContextRaw  []byte                 `db:"partial_rfq_context" json:"-"` // raw JSONB from DB
	RFCMessageID          string                 `db:"rfc_message_id" json:"rfc_message_id"`
	InReplyTo             string                 `db:"in_reply_to" json:"in_reply_to"`
	ReferencesHeader      string                 `db:"references_header" json:"references_header"`
	Sender                string                 `db:"sender" json:"sender"`
	Recipients            string                 `db:"recipients" json:"recipients"`
	CCRecipients          string                 `db:"cc_recipients" json:"cc_recipients"`
	CreatedAt             time.Time              `db:"created_at" json:"created_at"`
	MailboxID             *int64                 `db:"mailbox_id" json:"mailbox_id,omitempty"`
	Status                string                 `db:"status" json:"status"`
	RetryCount            int                    `db:"retry_count" json:"retry_count"`
	LastRetryAt           *time.Time             `db:"last_retry_at" json:"last_retry_at,omitempty"`
	LastError             *string                `db:"last_error" json:"last_error,omitempty"`
	MailboxEmail          string                 `db:"mailbox_email" json:"mailbox_email,omitempty"`
	IsIdempotent          bool                   `db:"-" json:"idempotent,omitempty"`
}

// LeadEmailDraft represents an auto-saved in-progress email reply draft.
type LeadEmailDraft struct {
	ID                  int64     `db:"id" json:"id"`
	OrgID               int64     `db:"org_id" json:"org_id"`
	LeadID              int64     `db:"lead_id" json:"lead_id"`
	ParentInteractionID int64     `db:"parent_interaction_id" json:"parent_interaction_id"`
	MailboxID           *int64    `db:"mailbox_id" json:"mailbox_id"`
	From                string    `db:"-" json:"from"`
	Recipients          string    `db:"recipients" json:"to"`
	CCRecipients        string    `db:"cc_recipients" json:"cc"`
	Subject             string    `db:"subject" json:"subject"`
	Content             string    `db:"content" json:"body"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
}

// UnmarshalPartialRFQContext deserializes PartialRFQContextRaw into PartialRFQContext.
func (li *LeadInteraction) UnmarshalPartialRFQContext() {
	if len(li.PartialRFQContextRaw) > 0 && string(li.PartialRFQContextRaw) != "{}" {
		_ = json.Unmarshal(li.PartialRFQContextRaw, &li.PartialRFQContext)
	}
}

// LogInteraction inserts a new lead interaction row into the database.
func (d *dataLayer) LogInteraction(ctx context.Context, inter *LeadInteraction) error {
	// Serialize PartialRFQContext to JSON if provided
	var contextJSON []byte
	if inter.PartialRFQContext != nil {
		var marshalErr error
		contextJSON, marshalErr = json.Marshal(inter.PartialRFQContext)
		if marshalErr != nil {
			contextJSON = []byte("{}")
		}
	} else {
		contextJSON = []byte("{}")
	}

	createdAtVal := time.Now()
	if !inter.CreatedAt.IsZero() {
		createdAtVal = inter.CreatedAt
	}

	statusVal := inter.Status
	if statusVal == "" {
		statusVal = "SENT"
	}

	query := `
		INSERT INTO lead_interactions (
			org_id, lead_id, channel, direction, subject, content, 
			raw_email_id, thread_id, sentiment, intent, linked_rfq_id, ai_confidence,
			ai_summary, drafted_reply, parent_interaction_id, partial_rfq_context,
			rfc_message_id, in_reply_to, references_header, sender, recipients, cc_recipients,
			created_at, updated_at, mailbox_id, status
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?, ?
		)
	`
	res, err := d.db.ExecContext(ctx, query,
		inter.OrgID, inter.LeadID, inter.Channel, inter.Direction, inter.Subject, inter.Content,
		inter.RawEmailID, inter.ThreadID, inter.Sentiment, inter.Intent, inter.LinkedRFQID, inter.AIConfidence,
		inter.AISummary, inter.DraftedReply, inter.ParentInteractionID, contextJSON,
		inter.RFCMessageID, inter.InReplyTo, inter.ReferencesHeader, inter.Sender, inter.Recipients, inter.CCRecipients,
		createdAtVal, inter.MailboxID, statusVal,
	)
	if err != nil {
		return fmt.Errorf("insert lead interaction: %w", err)
	}
	inter.ID, err = res.LastInsertId()
	if err != nil {
		return fmt.Errorf("get lead interaction id: %w", err)
	}
	return nil
}

// ListInteractions retrieves all interactions for a specific lead.
func (d *dataLayer) ListInteractions(ctx context.Context, orgID int32, leadID int32) ([]*LeadInteraction, error) {
	var list []*LeadInteraction
	query := `
		SELECT li.id, li.org_id, li.lead_id, li.channel, li.direction, li.subject, li.content, 
		       li.raw_email_id, li.thread_id, li.sentiment, li.intent, li.linked_rfq_id, li.ai_confidence, 
		       li.ai_summary, li.drafted_reply, li.parent_interaction_id, li.partial_rfq_context,
		       COALESCE(li.rfc_message_id, '') AS rfc_message_id,
		       COALESCE(li.in_reply_to, '') AS in_reply_to,
		       COALESCE(li.references_header, '') AS references_header,
		       COALESCE(li.sender, '') AS sender,
		       COALESCE(li.recipients, '') AS recipients,
		       COALESCE(li.cc_recipients, '') AS cc_recipients,
		       li.created_at, li.mailbox_id, li.status,
		       li.retry_count, li.last_retry_at, li.last_error,
		       COALESCE(m.email, '') AS mailbox_email
		FROM lead_interactions li
		LEFT JOIN org_connected_mailboxes m ON li.mailbox_id = m.id
		WHERE li.org_id = ? AND li.lead_id = ?
		ORDER BY li.created_at DESC
	`
	err := d.db.SelectContext(ctx, &list, query, orgID, leadID)
	if err != nil {
		return nil, fmt.Errorf("select lead interactions: %w", err)
	}
	for _, li := range list {
		li.UnmarshalPartialRFQContext()
	}
	return list, nil
}

// FindByThreadID finds interactions by thread_id to support email conversation tracking.
// Returns interactions ordered oldest-first so the caller can find the most recent incomplete one.
func (d *dataLayer) FindByThreadID(ctx context.Context, orgID int32, threadID string) ([]*LeadInteraction, error) {
	var list []*LeadInteraction
	query := `
		SELECT li.id, li.org_id, li.lead_id, li.channel, li.direction, li.subject, li.content, 
		       li.raw_email_id, li.thread_id, li.sentiment, li.intent, li.linked_rfq_id, li.ai_confidence, 
		       li.ai_summary, li.drafted_reply, li.parent_interaction_id, li.partial_rfq_context,
		       COALESCE(li.rfc_message_id, '') AS rfc_message_id,
		       COALESCE(li.in_reply_to, '') AS in_reply_to,
		       COALESCE(li.references_header, '') AS references_header,
		       COALESCE(li.sender, '') AS sender,
		       COALESCE(li.recipients, '') AS recipients,
		       COALESCE(li.cc_recipients, '') AS cc_recipients,
		       li.created_at, li.mailbox_id, li.status,
		       li.retry_count, li.last_retry_at, li.last_error,
		       COALESCE(m.email, '') AS mailbox_email
		FROM lead_interactions li
		LEFT JOIN org_connected_mailboxes m ON li.mailbox_id = m.id
		WHERE li.org_id = ? AND li.thread_id = ?
		ORDER BY li.created_at ASC
	`
	err := d.db.SelectContext(ctx, &list, query, orgID, threadID)
	if err != nil {
		return nil, fmt.Errorf("select lead interactions by thread: %w", err)
	}
	for _, li := range list {
		li.UnmarshalPartialRFQContext()
	}
	return list, nil
}

// UpdateInteractionAI updates the intent, sentiment, linked rfq, ai confidence, summary and drafted reply of a lead interaction.
func (d *dataLayer) UpdateInteractionAI(ctx context.Context, orgID int64, id int64, intent string, sentiment string, confidence int, linkedRFQID *int64, aiSummary string, draftedReply string) error {
	query := `
		UPDATE lead_interactions
		SET intent = ?, sentiment = ?, ai_confidence = ?, linked_rfq_id = ?, ai_summary = ?, drafted_reply = ?
		WHERE org_id = ? AND id = ?
	`
	_, err := d.db.ExecContext(ctx, query, intent, sentiment, confidence, linkedRFQID, aiSummary, draftedReply, orgID, id)
	if err != nil {
		return fmt.Errorf("update lead interaction AI: %w", err)
	}
	return nil
}

// UpdateInteractionContext persists the cumulative partial RFQ context extracted so far in a conversation thread.
// Called from SalesCallback when intent is RFQ_REQUEST_INCOMPLETE so the next reply can restore state.
func (d *dataLayer) UpdateInteractionContext(ctx context.Context, orgID int64, id int64, partialCtx map[string]interface{}) error {
	contextJSON, err := json.Marshal(partialCtx)
	if err != nil {
		return fmt.Errorf("marshal partial rfq context: %w", err)
	}
	query := `
		UPDATE lead_interactions
		SET partial_rfq_context = ?
		WHERE org_id = ? AND id = ?
	`
	_, err = d.db.ExecContext(ctx, query, contextJSON, orgID, id)
	if err != nil {
		return fmt.Errorf("update interaction context: %w", err)
	}
	return nil
}

// GetInteractionByRawEmailID retrieves an interaction by raw_email_id for deduplication.
func (d *dataLayer) GetInteractionByRawEmailID(ctx context.Context, orgID int32, rawEmailID string) (*LeadInteraction, error) {
	var inter LeadInteraction
	query := `
		SELECT li.id, li.org_id, li.lead_id, li.channel, li.direction, li.subject, li.content, 
		       li.raw_email_id, li.thread_id, li.sentiment, li.intent, li.linked_rfq_id, li.ai_confidence, 
		       li.ai_summary, li.drafted_reply, li.parent_interaction_id, li.partial_rfq_context,
		       COALESCE(li.rfc_message_id, '') AS rfc_message_id,
		       COALESCE(li.in_reply_to, '') AS in_reply_to,
		       COALESCE(li.references_header, '') AS references_header,
		       COALESCE(li.sender, '') AS sender,
		       COALESCE(li.recipients, '') AS recipients,
		       COALESCE(li.cc_recipients, '') AS cc_recipients,
		       li.created_at, li.mailbox_id, li.status,
		       li.retry_count, li.last_retry_at, li.last_error,
		       COALESCE(m.email, '') AS mailbox_email
		FROM lead_interactions li
		LEFT JOIN org_connected_mailboxes m ON li.mailbox_id = m.id
		WHERE li.org_id = ? AND li.raw_email_id = ?
		LIMIT 1
	`
	err := d.db.GetContext(ctx, &inter, query, orgID, rawEmailID)
	if err != nil {
		return nil, err
	}
	inter.UnmarshalPartialRFQContext()
	return &inter, nil
}

func (d *dataLayer) GetInteractionByID(ctx context.Context, orgID int32, id int64) (*LeadInteraction, error) {
	var inter LeadInteraction
	query := `
		SELECT li.id, li.org_id, li.lead_id, li.channel, li.direction, li.subject, li.content, 
		       li.raw_email_id, li.thread_id, li.sentiment, li.intent, li.linked_rfq_id, li.ai_confidence, 
		       li.ai_summary, li.drafted_reply, li.parent_interaction_id, li.partial_rfq_context,
		       COALESCE(li.rfc_message_id, '') AS rfc_message_id,
		       COALESCE(li.in_reply_to, '') AS in_reply_to,
		       COALESCE(li.references_header, '') AS references_header,
		       COALESCE(li.sender, '') AS sender,
		       COALESCE(li.recipients, '') AS recipients,
		       COALESCE(li.cc_recipients, '') AS cc_recipients,
		       li.created_at, li.mailbox_id, li.status,
		       li.retry_count, li.last_retry_at, li.last_error,
		       COALESCE(m.email, '') AS mailbox_email
		FROM lead_interactions li
		LEFT JOIN org_connected_mailboxes m ON li.mailbox_id = m.id
		WHERE li.org_id = ? AND li.id = ?
		LIMIT 1
	`
	err := d.db.GetContext(ctx, &inter, query, orgID, id)
	if err != nil {
		return nil, err
	}
	inter.UnmarshalPartialRFQContext()
	return &inter, nil
}

func (d *dataLayer) GetInteractionByRFCMessageID(ctx context.Context, orgID int32, rfcMessageID string) (*LeadInteraction, error) {
	var inter LeadInteraction
	query := `
		SELECT li.id, li.org_id, li.lead_id, li.channel, li.direction, li.subject, li.content, 
		       li.raw_email_id, li.thread_id, li.sentiment, li.intent, li.linked_rfq_id, li.ai_confidence, 
		       li.ai_summary, li.drafted_reply, li.parent_interaction_id, li.partial_rfq_context,
		       COALESCE(li.rfc_message_id, '') AS rfc_message_id,
		       COALESCE(li.in_reply_to, '') AS in_reply_to,
		       COALESCE(li.references_header, '') AS references_header,
		       COALESCE(li.sender, '') AS sender,
		       COALESCE(li.recipients, '') AS recipients,
		       COALESCE(li.cc_recipients, '') AS cc_recipients,
		       li.created_at, li.mailbox_id, li.status,
		       li.retry_count, li.last_retry_at, li.last_error,
		       COALESCE(m.email, '') AS mailbox_email
		FROM lead_interactions li
		LEFT JOIN org_connected_mailboxes m ON li.mailbox_id = m.id
		WHERE li.org_id = ? AND li.rfc_message_id = ?
		LIMIT 1
	`
	err := d.db.GetContext(ctx, &inter, query, orgID, rfcMessageID)
	if err != nil {
		return nil, err
	}
	inter.UnmarshalPartialRFQContext()
	return &inter, nil
}

func (d *dataLayer) UpdateInteractionStatusAndIDs(ctx context.Context, orgID int64, id int64, status string, rawEmailID string, rfcMessageID string, threadID string) error {
	query := `
		UPDATE lead_interactions
		SET status = ?, raw_email_id = ?, rfc_message_id = ?, thread_id = ?, updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	_, err := d.db.ExecContext(ctx, query, status, rawEmailID, rfcMessageID, threadID, orgID, id)
	return err
}

func (d *dataLayer) LockInteractionForRetry(ctx context.Context, orgID int64, id int64) (bool, error) {
	query := `
		UPDATE lead_interactions
		SET status = 'PENDING', retry_count = retry_count + 1, last_retry_at = NOW(), last_error = NULL, updated_at = NOW()
		WHERE id = ? AND org_id = ? AND direction = 'OUTBOUND' AND status = 'FAILED'
	`
	res, err := d.db.ExecContext(ctx, query, id, orgID)
	if err != nil {
		return false, fmt.Errorf("lock interaction for retry: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("get rows affected for retry lock: %w", err)
	}
	return rows > 0, nil
}

func (d *dataLayer) UpdateInteractionRetry(ctx context.Context, orgID int64, id int64, status string, lastError *string, incrementRetry bool, rawEmailID string, rfcMessageID string, threadID string) error {
	var err error
	if incrementRetry {
		query := `
			UPDATE lead_interactions
			SET status = ?, retry_count = retry_count + 1, last_retry_at = NOW(), last_error = ?, raw_email_id = ?, rfc_message_id = ?, thread_id = ?, updated_at = NOW()
			WHERE org_id = ? AND id = ?
		`
		_, err = d.db.ExecContext(ctx, query, status, lastError, rawEmailID, rfcMessageID, threadID, orgID, id)
	} else {
		query := `
			UPDATE lead_interactions
			SET status = ?, last_error = ?, raw_email_id = ?, rfc_message_id = ?, thread_id = ?, updated_at = NOW()
			WHERE org_id = ? AND id = ?
		`
		_, err = d.db.ExecContext(ctx, query, status, lastError, rawEmailID, rfcMessageID, threadID, orgID, id)
	}
	if err != nil {
		return fmt.Errorf("update interaction retry state: %w", err)
	}
	return nil
}

func (d *dataLayer) GetDraft(ctx context.Context, orgID int64, leadID int64, parentInteractionID int64) (*LeadEmailDraft, error) {
	var draft LeadEmailDraft
	query := `
		SELECT id, org_id, lead_id, parent_interaction_id, mailbox_id, 
		       COALESCE(recipients, '') AS recipients, 
		       COALESCE(cc_recipients, '') AS cc_recipients, 
		       COALESCE(subject, '') AS subject, 
		       content, created_at, updated_at
		FROM lead_email_drafts
		WHERE org_id = ? AND lead_id = ? AND parent_interaction_id = ?
		LIMIT 1
	`
	err := d.db.GetContext(ctx, &draft, query, orgID, leadID, parentInteractionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get lead email draft: %w", err)
	}
	return &draft, nil
}

func (d *dataLayer) SaveDraft(ctx context.Context, draft *LeadEmailDraft) error {
	query := `
		INSERT INTO lead_email_drafts (
			org_id, lead_id, parent_interaction_id, mailbox_id, recipients, cc_recipients, subject, content, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
		) ON DUPLICATE KEY UPDATE
			mailbox_id = VALUES(mailbox_id),
			recipients = VALUES(recipients),
			cc_recipients = VALUES(cc_recipients),
			subject = VALUES(subject),
			content = VALUES(content),
			updated_at = NOW()
	`
	_, err := d.db.ExecContext(ctx, query,
		draft.OrgID, draft.LeadID, draft.ParentInteractionID, draft.MailboxID,
		draft.Recipients, draft.CCRecipients, draft.Subject, draft.Content,
	)
	if err != nil {
		return fmt.Errorf("save lead email draft: %w", err)
	}
	return nil
}

func (d *dataLayer) DeleteDraft(ctx context.Context, orgID int64, leadID int64, parentInteractionID int64) error {
	query := `
		DELETE FROM lead_email_drafts
		WHERE org_id = ? AND lead_id = ? AND parent_interaction_id = ?
	`
	_, err := d.db.ExecContext(ctx, query, orgID, leadID, parentInteractionID)
	if err != nil {
		return fmt.Errorf("delete lead email draft: %w", err)
	}
	return nil
}

func (d *dataLayer) PurgeNonLogisticsLeads(ctx context.Context, orgID int32) error {
	// 1. Delete interactions belonging to bank/non-logistics leads
	queryInter := `
		DELETE FROM lead_interactions 
		WHERE lead_id IN (
			SELECT id FROM leads 
			WHERE (? = 0 OR org_id = ?) AND (
				email LIKE '%sbi%' OR email LIKE '%alerts%' OR email LIKE '%noreply%' 
				OR email LIKE '%no-reply%' OR email LIKE '%donotreply%' OR email LIKE '%bank%' 
				OR email LIKE '%subscribe%' OR email LIKE '%newsletter%' OR email LIKE '%indianexpress%'
				OR email LIKE '%promo%' OR email LIKE '%marketing%' OR email LIKE '%notification%'
				OR email LIKE '%linkedin%' OR email LIKE '%naukri%'
				OR company_name LIKE '%cbsalerts%' OR company_name LIKE '%Inbound Lead (cbsalerts%'
				OR company_name LIKE '%Inbound Lead (subscribe%' OR company_name LIKE '%SBI%'
				OR notes LIKE '%CBSSBI ALERT%' OR notes LIKE '%hold for INR%' OR notes LIKE '%Indian Express%'
			)
		)
	`
	_, _ = d.db.ExecContext(ctx, queryInter, orgID, orgID)

	// 2. Delete non-logistics leads
	queryLeads := `
		DELETE FROM leads 
		WHERE (? = 0 OR org_id = ?) AND (
			email LIKE '%sbi%' OR email LIKE '%alerts%' OR email LIKE '%noreply%' 
			OR email LIKE '%no-reply%' OR email LIKE '%donotreply%' OR email LIKE '%bank%' 
			OR email LIKE '%subscribe%' OR email LIKE '%newsletter%' OR email LIKE '%indianexpress%'
			OR email LIKE '%promo%' OR email LIKE '%marketing%' OR email LIKE '%notification%'
			OR email LIKE '%linkedin%' OR email LIKE '%naukri%'
			OR company_name LIKE '%cbsalerts%' OR company_name LIKE '%Inbound Lead (cbsalerts%'
			OR company_name LIKE '%Inbound Lead (subscribe%' OR company_name LIKE '%SBI%'
			OR notes LIKE '%CBSSBI ALERT%' OR notes LIKE '%hold for INR%' OR notes LIKE '%Indian Express%'
		)
	`
	_, err := d.db.ExecContext(ctx, queryLeads, orgID, orgID)
	return err
}


