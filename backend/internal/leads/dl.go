package leads

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/freel/backend/internal/leads/spec"
	"github.com/jmoiron/sqlx"
)

type Datalayer interface {
	Create(ctx context.Context, lead *spec.Lead) error
	GetByID(ctx context.Context, orgID int32, id int32) (*spec.Lead, error)
	List(ctx context.Context, orgID int32, limit int, offset int, status *string) ([]*spec.Lead, int, error)
	Update(ctx context.Context, lead *spec.Lead) error
	Delete(ctx context.Context, orgID int32, id int32) error
}

type dataLayer struct {
	db *sqlx.DB
}

func NewDataLayer(db *sqlx.DB) Datalayer {
	return &dataLayer{db: db}
}

func (d *dataLayer) Create(ctx context.Context, lead *spec.Lead) error {
	query := `
		INSERT INTO leads (
			org_id, company_name, contact_name, email, phone, status, source
		) VALUES (
			:org_id, :company_name, :contact_name, :email, :phone, :status, :source
		) RETURNING id, created_at, updated_at
	`
	rows, err := d.db.NamedQueryContext(ctx, query, lead)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.StructScan(lead)
	}
	return errors.New("no rows returned after insert")
}

func (d *dataLayer) GetByID(ctx context.Context, orgID int32, id int32) (*spec.Lead, error) {
	query := `SELECT * FROM leads WHERE id = $1 AND org_id = $2`
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

func (d *dataLayer) List(ctx context.Context, orgID int32, limit int, offset int, status *string) ([]*spec.Lead, int, error) {
	var leads []*spec.Lead
	var total int

	// Build query dynamically based on status filter
	baseQuery := `FROM leads WHERE org_id = $1`
	args := []interface{}{orgID}
	argIdx := 2

	if status != nil {
		baseQuery += ` AND status = $2`
		args = append(args, *status)
		argIdx++
	}

	countQuery := `SELECT count(*) ` + baseQuery
	err := d.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*spec.Lead{}, 0, nil
	}

	selectQuery := `SELECT * ` + baseQuery + ` ORDER BY created_at DESC LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)
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
			company_name = :company_name,
			contact_name = :contact_name,
			email = :email,
			phone = :phone,
			status = :status,
			source = :source,
			ai_score = :ai_score,
			ai_research_report = :ai_research_report,
			updated_at = NOW()
		WHERE id = :id AND org_id = :org_id
		RETURNING updated_at
	`
	rows, err := d.db.NamedQueryContext(ctx, query, lead)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&lead.UpdatedAt)
	}
	return sql.ErrNoRows
}

func (d *dataLayer) Delete(ctx context.Context, orgID int32, id int32) error {
	query := `DELETE FROM leads WHERE id = $1 AND org_id = $2`
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

// helper for arg index building
func itoa(i int) string {
	return strconv.Itoa(i)
}
