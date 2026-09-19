package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env from various possible locations
	for _, envPath := range []string{".env", "../.env", "../../.env", "backend/.env"} {
		if err := godotenv.Load(envPath); err == nil {
			break
		}
	}

	host := getenv("DB_HOST", "127.0.0.1")
	port := getenv("DB_PORT", "3306")
	user := getenv("DB_USER", "root")
	pass := getenv("DB_PASSWORD", "")
	name := getenv("DB_NAME", "freel_mysql")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true", user, pass, host, port, name)
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Cannot ping DB: %v", err)
	}

	var sql []byte
	for _, sqlPath := range []string{"scripts/migrate_debit_notes.sql", "../scripts/migrate_debit_notes.sql", "../../scripts/migrate_debit_notes.sql", "backend/scripts/migrate_debit_notes.sql"} {
		data, err := os.ReadFile(sqlPath)
		if err == nil {
			sql = data
			break
		}
	}
	if len(sql) == 0 {
		log.Fatalf("Cannot read migration file migrate_debit_notes.sql")
	}

	if _, err := db.Exec(string(sql)); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("✅ Debit Note migration applied successfully!")
	fmt.Println("   Tables created: debit_notes, debit_note_items")
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
