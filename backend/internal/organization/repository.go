package organization

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// Repository defines the data access methods for Organization.
type Repository interface {
	Create(ctx context.Context, org *Organization) error
	GetByID(ctx context.Context, id int64) (*Organization, error)
	Update(ctx context.Context, org *Organization) error
}

type repository struct {
	db *sqlx.DB
}

// NewRepository creates a new organization repository.
// Simple meaning: It gets the database connection ready so we can run SQL queries for organizations.
// Example: repo := NewRepository(sqlxDB)
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

// Create inserts a new organization into the database.
// Simple meaning: It takes the organization data and saves it permanently into the Postgres database.
// Example: err := repo.Create(ctx, &Organization{Name: "Acme Corp"})
func (r *repository) Create(ctx context.Context, org *Organization) error {
	query := `INSERT INTO organizations (name, created_at, updated_at) VALUES (:name, :created_at, :updated_at) RETURNING id`
	
	// Prepare the statement with named arguments
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	return stmt.QueryRowxContext(ctx, org).Scan(&org.ID)
}

// GetByID retrieves an organization by its ID.
// Simple meaning: It looks up the database to find a specific company workspace using its unique ID number.
// Example: org, err := repo.GetByID(ctx, 123)
func (r *repository) GetByID(ctx context.Context, id int64) (*Organization, error) {
	var org Organization
	query := `SELECT id, name, created_at, updated_at FROM organizations WHERE id = $1`
	err := r.db.GetContext(ctx, &org, query, id)
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// Update updates an existing organization.
// Simple meaning: It changes the saved details of a company workspace in the database.
// Example: err := repo.Update(ctx, &Organization{ID: 123, Name: "Acme Global"})
func (r *repository) Update(ctx context.Context, org *Organization) error {
	query := `UPDATE organizations SET name = :name, updated_at = :updated_at WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, query, org)
	return err
}
