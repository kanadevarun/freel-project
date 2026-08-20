package leads

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/leads/spec"
	"github.com/jmoiron/sqlx"
)

type Datalayer interface {
	Create(ctx context.Context, lead *spec.Lead) error
	GetByID(ctx context.Context, orgID int32, id int32) (*spec.Lead, error)
	List(ctx context.Context, orgID int32, limit int, offset int, status *string) ([]*spec.Lead, int, error)
	Update(ctx context.Context, lead *spec.Lead) error
	Delete(ctx context.Context, orgID int32, id int32) error
	GetByEmail(ctx context.Context, orgID int32, email string) (*spec.Lead, error)
	LogInteraction(ctx context.Context, inter *LeadInteraction) error
	ListInteractions(ctx context.Context, orgID int32, leadID int32) ([]*LeadInteraction, error)
	FindByThreadID(ctx context.Context, orgID int32, threadID string) ([]*LeadInteraction, error)
	CreateAITask(ctx context.Context, orgID int64, entityType string, entityID string, taskType string, payload map[string]interface{}) error
	UpdateInteractionAI(ctx context.Context, orgID int64, id int64, intent string, sentiment string, confidence int, linkedRFQID *int64, aiSummary string, draftedReply string) error
	UpdateInteractionContext(ctx context.Context, orgID int64, id int64, partialCtx map[string]interface{}) error
	EnsureCustomerForLead(ctx context.Context, lead *spec.Lead) error
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

	// Check if company already exists
	var companyID int64
	err = tx.QueryRowxContext(ctx, `SELECT id FROM companies WHERE org_id = ? AND name = ? LIMIT 1`, lead.OrgID, lead.CompanyName).Scan(&companyID)
	if err != nil {
		res, err := tx.ExecContext(ctx, `
			INSERT INTO companies (org_id, name, created_at, updated_at)
			VALUES (?, ?, NOW(), NOW())
		`, lead.OrgID, lead.CompanyName)
		if err != nil {
			return err
		}
		companyID, err = res.LastInsertId()
		if err != nil {
			return err
		}
	}

	// Check if customer link exists
	var custID int64
	err = tx.QueryRowxContext(ctx, `SELECT id FROM customers WHERE org_id = ? AND company_id = ? LIMIT 1`, lead.OrgID, companyID).Scan(&custID)
	if err != nil {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO customers (org_id, company_id, status, created_at, updated_at)
			VALUES (?, ?, 'ACTIVE', NOW(), NOW())
		`, lead.OrgID, companyID)
		if err != nil {
			return err
		}
	}

	// Create contact if details provided
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

	if contactName != "" || emailStr != "" {
		parts := strings.SplitN(contactName, " ", 2)
		firstName := parts[0]
		lastName := ""
		if len(parts) > 1 {
			lastName = parts[1]
		}
		_, _ = tx.ExecContext(ctx, `
			INSERT INTO contacts (org_id, company_id, first_name, last_name, email, phone, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
		`, lead.OrgID, companyID, firstName, lastName, emailStr, phoneStr)
	}

	return tx.Commit()
}

func (d *dataLayer) Create(ctx context.Context, lead *spec.Lead) error {
	query := `
		INSERT INTO leads (
			org_id, company_name, contact_name, email, phone, status, source, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, NOW(), NOW()
		)
	`
	res, err := d.db.ExecContext(ctx, query, lead.OrgID, lead.CompanyName, lead.ContactName, lead.Email, lead.Phone, lead.Status, lead.Source)
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
	query := `SELECT * FROM leads WHERE id = ? AND org_id = ?`
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
	query := `SELECT * FROM leads WHERE email = ? AND org_id = ?`
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

func (d *dataLayer) List(ctx context.Context, orgID int32, limit int, offset int, status *string) ([]*spec.Lead, int, error) {
	var leads []*spec.Lead
	var total int

	baseQuery := `FROM leads WHERE org_id = ?`
	args := []interface{}{orgID}

	if status != nil {
		baseQuery += ` AND status = ?`
		args = append(args, *status)
	}

	countQuery := `SELECT count(*) ` + baseQuery
	err := d.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*spec.Lead{}, 0, nil
	}

	selectQuery := `SELECT * ` + baseQuery + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	err = d.db.SelectContext(ctx, &leads, selectQuery, args...)
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
			updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	result, err := d.db.ExecContext(ctx, query,
		lead.CompanyName, lead.ContactName, lead.Email, lead.Phone,
		lead.Status, lead.Source, lead.AIScore, lead.AIResearchReport,
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
