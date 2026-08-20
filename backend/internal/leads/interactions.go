package leads

import (
	"context"
	"encoding/json"
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
	CreatedAt             time.Time              `db:"created_at" json:"created_at"`
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

	query := `
		INSERT INTO lead_interactions (
			org_id, lead_id, channel, direction, subject, content, 
			raw_email_id, thread_id, sentiment, intent, linked_rfq_id, ai_confidence,
			ai_summary, drafted_reply, parent_interaction_id, partial_rfq_context, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
		)
	`
	res, err := d.db.ExecContext(ctx, query,
		inter.OrgID, inter.LeadID, inter.Channel, inter.Direction, inter.Subject, inter.Content,
		inter.RawEmailID, inter.ThreadID, inter.Sentiment, inter.Intent, inter.LinkedRFQID, inter.AIConfidence,
		inter.AISummary, inter.DraftedReply, inter.ParentInteractionID, contextJSON,
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
		SELECT id, org_id, lead_id, channel, direction, subject, content, 
		       raw_email_id, thread_id, sentiment, intent, linked_rfq_id, ai_confidence, 
		       ai_summary, drafted_reply, parent_interaction_id, partial_rfq_context, created_at
		FROM lead_interactions
		WHERE org_id = ? AND lead_id = ?
		ORDER BY created_at DESC
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
		SELECT id, org_id, lead_id, channel, direction, subject, content, 
		       raw_email_id, thread_id, sentiment, intent, linked_rfq_id, ai_confidence, 
		       ai_summary, drafted_reply, parent_interaction_id, partial_rfq_context, created_at
		FROM lead_interactions
		WHERE org_id = ? AND thread_id = ?
		ORDER BY created_at ASC
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

