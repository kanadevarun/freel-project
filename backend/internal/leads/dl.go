package leads

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/freel/backend/internal/leads/spec"
	"github.com/jmoiron/sqlx"
)

type Datalayer interface {
	Create(ctx context.Context, lead *spec.Lead) error
	GetByID(ctx context.Context, orgID int32, id int32) (*spec.Lead, error)
	List(ctx context.Context, orgID int32, limit int, offset int, status *string, search *string, source *string) ([]*spec.Lead, int, error)
	Update(ctx context.Context, lead *spec.Lead) error
	Delete(ctx context.Context, orgID int32, id int32) error
	GetByEmail(ctx context.Context, orgID int32, email string) (*spec.Lead, error)
	LogInteraction(ctx context.Context, inter *LeadInteraction) error
	ListInteractions(ctx context.Context, orgID int32, leadID int32) ([]*LeadInteraction, error)
	FindByThreadID(ctx context.Context, orgID int32, threadID string) ([]*LeadInteraction, error)
	GetInteractionByRawEmailID(ctx context.Context, orgID int32, rawEmailID string) (*LeadInteraction, error)
	GetInteractionByID(ctx context.Context, orgID int32, id int64) (*LeadInteraction, error)
	GetInteractionByRFCMessageID(ctx context.Context, orgID int32, rfcMessageID string) (*LeadInteraction, error)
	UpdateInteractionStatusAndIDs(ctx context.Context, orgID int64, id int64, status string, rawEmailID string, rfcMessageID string, threadID string) error
	LockInteractionForRetry(ctx context.Context, orgID int64, id int64) (bool, error)
	UpdateInteractionRetry(ctx context.Context, orgID int64, id int64, status string, lastError *string, incrementRetry bool, rawEmailID string, rfcMessageID string, threadID string) error
	GetDraft(ctx context.Context, orgID int64, leadID int64, parentInteractionID int64) (*LeadEmailDraft, error)
	SaveDraft(ctx context.Context, draft *LeadEmailDraft) error
	DeleteDraft(ctx context.Context, orgID int64, leadID int64, parentInteractionID int64) error
	CreateAITask(ctx context.Context, orgID int64, entityType string, entityID string, taskType string, payload map[string]interface{}) error
	UpdateInteractionAI(ctx context.Context, orgID int64, id int64, intent string, sentiment string, confidence int, linkedRFQID *int64, aiSummary string, draftedReply string) error
	UpdateInteractionContext(ctx context.Context, orgID int64, id int64, partialCtx map[string]interface{}) error
	EnsureCustomerForLead(ctx context.Context, lead *spec.Lead) error
	GetCustomerIDByCompanyName(ctx context.Context, orgID int32, name string) (int32, error)
	GetTags(ctx context.Context, leadID int32) ([]string, error)
	GetTagsBatch(ctx context.Context, leadIDs []int32) (map[int32][]string, error)
	SetTags(ctx context.Context, leadID int32, tags []string) error
	CreateActivity(ctx context.Context, orgID int32, entityType string, entityID int32, action string, description string, userID *int64) error
	GetActivities(ctx context.Context, orgID int32, leadID int32) ([]spec.TimelineEvent, error)
	UserExistsInOrg(ctx context.Context, orgID int32, userID int64) (bool, error)
	PurgeNonLogisticsLeads(ctx context.Context, orgID int32) error
}

type dataLayer struct {
	db *sqlx.DB
}

func NewDataLayer(db *sqlx.DB) Datalayer {
	return &dataLayer{db: db}
}

func (d *dataLayer) EnsureCustomerForLead(ctx context.Context, lead *spec.Lead) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if a customer record already exists for this org + company name
	var customerID int64
	err = tx.QueryRowxContext(ctx,
		`SELECT id FROM customers WHERE org_id = ? AND name = ? LIMIT 1`,
		lead.OrgID, lead.CompanyName,
	).Scan(&customerID)
	if err != nil {
		// No customer yet — create one directly in the customers table
		contactName := ""
		if lead.ContactName != nil {
			contactName = *lead.ContactName
		}
		emailStr := ""
		if lead.Email != nil {
			emailStr = *lead.Email
		}
		phoneStr := ""
		if lead.Phone != nil {
			phoneStr = *lead.Phone
		}

		res, err := tx.ExecContext(ctx, `
			INSERT INTO customers (org_id, name, contact_name, contact_email, contact_phone, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, 'ACTIVE', NOW(), NOW())
		`, lead.OrgID, lead.CompanyName, contactName, emailStr, phoneStr)
		if err != nil {
			return err
		}
		customerID, err = res.LastInsertId()
		if err != nil {
			return err
		}

		// Also insert a contact row linked to this new customer
		if contactName != "" || emailStr != "" {
			parts := strings.SplitN(contactName, " ", 2)
			firstName := parts[0]
			lastName := ""
			if len(parts) > 1 {
				lastName = parts[1]
			}
			_, _ = tx.ExecContext(ctx, `
				INSERT INTO contacts (org_id, customer_id, first_name, last_name, email, phone, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
			`, lead.OrgID, customerID, firstName, lastName, emailStr, phoneStr)
		}
	}

	return tx.Commit()
}

func (d *dataLayer) Create(ctx context.Context, lead *spec.Lead) error {
	query := `
		INSERT INTO leads (
			org_id, company_name, contact_name, email, phone, status, source, notes, location, assigned_to, assigned_at, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
		)
	`
	res, err := d.db.ExecContext(ctx, query, lead.OrgID, lead.CompanyName, lead.ContactName, lead.Email, lead.Phone, lead.Status, lead.Source, lead.Notes, lead.Location, lead.AssignedTo, lead.AssignedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	lead.ID = int32(id)
	return nil
}

func (d *dataLayer) GetByID(ctx context.Context, orgID int32, id int32) (*spec.Lead, error) {
	query := `
		SELECT 
			l.*, 
			COALESCE(CONCAT(u.first_name, ' ', u.last_name), u.email) as assigned_to_name,
			r.id as linked_rfq_id,
			r.rfq_number as linked_rfq_number,
			oc.name as campaign_name
		FROM leads l
		LEFT JOIN users u ON l.assigned_to = u.id
		LEFT JOIN rfqs r ON r.lead_id = l.id AND r.org_id = l.org_id
		LEFT JOIN outreach_campaigns oc ON l.campaign_id = oc.id
		WHERE l.id = ? AND l.org_id = ?
		ORDER BY r.created_at DESC
		LIMIT 1
	`
	var lead spec.Lead
	err := d.db.GetContext(ctx, &lead, query, id, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &lead, nil
}

func (d *dataLayer) GetByEmail(ctx context.Context, orgID int32, email string) (*spec.Lead, error) {
	query := `
		SELECT 
			l.*, 
			COALESCE(CONCAT(u.first_name, ' ', u.last_name), u.email) as assigned_to_name,
			r.id as linked_rfq_id,
			r.rfq_number as linked_rfq_number,
			oc.name as campaign_name
		FROM leads l
		LEFT JOIN users u ON l.assigned_to = u.id
		LEFT JOIN rfqs r ON r.lead_id = l.id AND r.org_id = l.org_id
		LEFT JOIN outreach_campaigns oc ON l.campaign_id = oc.id
		WHERE l.email = ? AND l.org_id = ?
		ORDER BY l.created_at DESC
		LIMIT 1
	`
	var lead spec.Lead
	err := d.db.GetContext(ctx, &lead, query, email, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &lead, nil
}

func (d *dataLayer) List(ctx context.Context, orgID int32, limit int, offset int, status *string, search *string, source *string) ([]*spec.Lead, int, error) {
	var leads []*spec.Lead
	var total int

	baseQuery := `FROM leads l LEFT JOIN users u ON l.assigned_to = u.id WHERE l.org_id = ?`
	args := []interface{}{orgID}

	if status != nil {
		baseQuery += ` AND l.status = ?`
		args = append(args, *status)
	}

	if source != nil && *source != "" {
		baseQuery += ` AND l.source = ?`
		args = append(args, *source)
	}

	if search != nil && *search != "" {
		baseQuery += ` AND (l.company_name LIKE ? OR l.contact_name LIKE ? OR l.email LIKE ?)`
		sArg := "%" + *search + "%"
		args = append(args, sArg, sArg, sArg)
	}

	countQuery := `SELECT count(*) ` + baseQuery
	err := d.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*spec.Lead{}, 0, nil
	}

	selectQuery := `
		SELECT 
			l.*, 
			COALESCE(CONCAT(u.first_name, ' ', u.last_name), u.email) as assigned_to_name,
			r.id as linked_rfq_id,
			r.rfq_number as linked_rfq_number,
			oc.name as campaign_name
		FROM leads l 
		LEFT JOIN users u ON l.assigned_to = u.id 
		LEFT JOIN rfqs r ON r.lead_id = l.id AND r.org_id = l.org_id
		LEFT JOIN outreach_campaigns oc ON l.campaign_id = oc.id
		WHERE l.org_id = ?
	`
	selectArgs := []interface{}{orgID}

	if status != nil {
		selectQuery += ` AND l.status = ?`
		selectArgs = append(selectArgs, *status)
	}

	if source != nil && *source != "" {
		selectQuery += ` AND l.source = ?`
		selectArgs = append(selectArgs, *source)
	}

	if search != nil && *search != "" {
		selectQuery += ` AND (l.company_name LIKE ? OR l.contact_name LIKE ? OR l.email LIKE ?)`
		sArg := "%" + *search + "%"
		selectArgs = append(selectArgs, sArg, sArg, sArg)
	}

	selectQuery += ` ORDER BY l.created_at DESC LIMIT ? OFFSET ?`
	selectArgs = append(selectArgs, limit, offset)

	err = d.db.SelectContext(ctx, &leads, selectQuery, selectArgs...)
	if err != nil {
		return nil, 0, err
	}

	return leads, total, nil
}


func (d *dataLayer) Update(ctx context.Context, lead *spec.Lead) error {
	query := `
		UPDATE leads SET
			company_name = ?,
			contact_name = ?,
			email = ?,
			phone = ?,
			status = ?,
			source = ?,
			ai_score = ?,
			ai_research_report = ?,
			notes = ?,
			location = ?,
			assigned_to = ?,
			assigned_at = ?,
			updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	result, err := d.db.ExecContext(ctx, query,
		lead.CompanyName, lead.ContactName, lead.Email, lead.Phone,
		lead.Status, lead.Source, lead.AIScore, lead.AIResearchReport,
		lead.Notes, lead.Location, lead.AssignedTo, lead.AssignedAt,
		lead.ID, lead.OrgID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (d *dataLayer) Delete(ctx context.Context, orgID int32, id int32) error {
	query := `DELETE FROM leads WHERE id = ? AND org_id = ?`
	result, err := d.db.ExecContext(ctx, query, id, orgID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (d *dataLayer) CreateAITask(ctx context.Context, orgID int64, entityType string, entityID string, taskType string, payload map[string]interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal task payload: %w", err)
	}

	const query = `
INSERT INTO ai_processing_tasks (
    org_id, entity_type, entity_id, task_type, payload, status, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, 'QUEUED', NOW(), NOW()
)
`
	_, err = d.db.ExecContext(ctx, query, orgID, entityType, entityID, taskType, payloadJSON)
	return err
}

// helper for arg index building
func itoa(i int) string {
	return strconv.Itoa(i)
}

func (d *dataLayer) GetTags(ctx context.Context, leadID int32) ([]string, error) {
	var tags []string
	query := `SELECT tag FROM lead_tags WHERE lead_id = ? ORDER BY tag ASC`
	err := d.db.SelectContext(ctx, &tags, query, leadID)
	if err != nil {
		return nil, err
	}
	if tags == nil {
		return []string{}, nil
	}
	return tags, nil
}

func (d *dataLayer) GetTagsBatch(ctx context.Context, leadIDs []int32) (map[int32][]string, error) {
	result := make(map[int32][]string)
	for _, id := range leadIDs {
		result[id] = []string{}
	}
	if len(leadIDs) == 0 {
		return result, nil
	}

	query, args, err := sqlx.In(`SELECT lead_id, tag FROM lead_tags WHERE lead_id IN (?) ORDER BY tag ASC`, leadIDs)
	if err != nil {
		return nil, err
	}
	query = d.db.Rebind(query)

	type leadTagRow struct {
		LeadID int32  `db:"lead_id"`
		Tag    string `db:"tag"`
	}
	var rows []leadTagRow
	err = d.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.LeadID] = append(result[row.LeadID], row.Tag)
	}
	return result, nil
}

func (d *dataLayer) SetTags(ctx context.Context, leadID int32, tags []string) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM lead_tags WHERE lead_id = ?`, leadID)
	if err != nil {
		return err
	}

	if len(tags) > 0 {
		seen := make(map[string]bool)
		var uniqueTags []string
		for _, t := range tags {
			trimmed := strings.TrimSpace(t)
			if trimmed == "" {
				continue
			}
			lower := strings.ToLower(trimmed)
			if !seen[lower] {
				seen[lower] = true
				uniqueTags = append(uniqueTags, trimmed)
			}
		}

		for _, tag := range uniqueTags {
			_, err = tx.ExecContext(ctx, `INSERT INTO lead_tags (lead_id, tag) VALUES (?, ?)`, leadID, tag)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (d *dataLayer) CreateActivity(ctx context.Context, orgID int32, entityType string, entityID int32, action string, description string, userID *int64) error {
	query := `
		INSERT INTO activities (org_id, entity_type, entity_id, action, description, user_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW())
	`
	_, err := d.db.ExecContext(ctx, query, orgID, entityType, entityID, action, description, userID)
	return err
}

func (d *dataLayer) GetActivities(ctx context.Context, orgID int32, leadID int32) ([]spec.TimelineEvent, error) {
	var dbEvents []struct {
		Action      string    `db:"action"`
		Description string    `db:"description"`
		Actor       *string   `db:"actor"`
		CreatedAt   time.Time `db:"created_at"`
	}
	query := `
		SELECT 
			a.action, 
			a.description, 
			COALESCE(CONCAT(u.first_name, ' ', u.last_name), u.email, 'System') as actor,
			a.created_at
		FROM activities a
		LEFT JOIN users u ON a.user_id = u.id
		WHERE a.entity_type = 'LEAD' AND a.entity_id = ? AND a.org_id = ?
	`
	err := d.db.SelectContext(ctx, &dbEvents, query, leadID, orgID)
	if err != nil {
		return nil, err
	}

	var dbInteractions []struct {
		ID         int64     `db:"id"`
		Channel    string    `db:"channel"`
		Direction  string    `db:"direction"`
		Subject    *string   `db:"subject"`
		Content    string    `db:"content"`
		CreatedAt  time.Time `db:"created_at"`
	}
	queryInter := `
		SELECT id, channel, direction, subject, content, created_at
		FROM lead_interactions
		WHERE lead_id = ? AND org_id = ?
	`
	err = d.db.SelectContext(ctx, &dbInteractions, queryInter, leadID, orgID)
	if err != nil {
		return nil, err
	}

	var events []spec.TimelineEvent

	for _, dbEv := range dbEvents {
		actor := "System"
		if dbEv.Actor != nil {
			actor = *dbEv.Actor
		}
		events = append(events, spec.TimelineEvent{
			Action:      dbEv.Action,
			Description: dbEv.Description,
			Actor:       actor,
			Source:      "SYSTEM",
			Timestamp:   dbEv.CreatedAt,
		})
	}

	for _, inter := range dbInteractions {
		subj := ""
		if inter.Subject != nil {
			subj = *inter.Subject
		}
		
		desc := inter.Content
		if len(desc) > 200 {
			desc = desc[:197] + "..."
		}
		
		if inter.Channel == "EMAIL" {
			desc = fmt.Sprintf("Email %s: \"%s\" - %s", strings.ToLower(inter.Direction), desc, subj)
		} else {
			desc = fmt.Sprintf("Interaction (%s) %s: %s", strings.ToLower(inter.Channel), strings.ToLower(inter.Direction), desc)
		}

		actor := "AI Agent"
		source := "AI"
		if inter.Direction == "OUTBOUND" {
			actor = "System"
			source = "SYSTEM"
		} else {
			actor = "Customer"
			source = "EXTERNAL"
		}
		
		events = append(events, spec.TimelineEvent{
			Action:        inter.Channel + "_" + inter.Direction,
			Description:   desc,
			Actor:         actor,
			Source:        source,
			Timestamp:     inter.CreatedAt,
			InteractionID: &inter.ID,
		})
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})

	return events, nil
}

func (d *dataLayer) UserExistsInOrg(ctx context.Context, orgID int32, userID int64) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM org_members 
		WHERE org_id = ? AND user_id = ? AND status = 'ACTIVE'
	`
	err := d.db.GetContext(ctx, &count, query, orgID, userID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *dataLayer) GetCustomerIDByCompanyName(ctx context.Context, orgID int32, name string) (int32, error) {
	var customerID int32
	query := `SELECT id FROM customers WHERE org_id = ? AND name = ? LIMIT 1`
	err := d.db.GetContext(ctx, &customerID, query, orgID, name)
	if err != nil {
		return 0, err
	}
	return customerID, nil
}

