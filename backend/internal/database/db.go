package database

import (
	"context"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/jmoiron/sqlx"
)

// Connect connects to the MySQL database using the provided DSN and connection pooling settings.
// Example: db, err := database.Connect("root@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true")
func Connect(dbURL string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", dbURL)
	if err != nil {
		return nil, err
	}

	// Set connection pool limits for stability
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Successfully connected to MySQL database!")
	return db, nil
}

// Ping checks if the database is alive and reachable.
// Simple meaning: It sends a quick "hello, are you there?" to the database to ensure it's still connected.
// Example: err := database.Ping(ctx, db)
func Ping(ctx context.Context, db *sqlx.DB) error {
	return db.PingContext(ctx)
}
