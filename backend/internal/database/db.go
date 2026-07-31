package database

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Postgres driver
)

// Connect connects to the Postgres database using the provided URL and connection pooling settings.
// Simple meaning: It opens up the pipeline to the database so our app can save and read data.
// Example: db, err := database.Connect("postgres://user:pass@localhost:5432/db")
func Connect(dbURL string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	// Set connection pool limits for stability
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// Ping checks if the database is alive and reachable.
// Simple meaning: It sends a quick "hello, are you there?" to the database to ensure it's still connected.
// Example: err := database.Ping(ctx, db)
func Ping(ctx context.Context, db *sqlx.DB) error {
	return db.PingContext(ctx)
}
